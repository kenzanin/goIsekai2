// Package updates runs the background maintenance and auto-update loop.
package updates

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"goisekai2/internal/config"
	"goisekai2/internal/plugin"
	"goisekai2/internal/store"

	"gorm.io/gorm"
)

// Scheduler owns the single maintenance goroutine.
type Scheduler struct {
	store    *store.Store
	host     *plugin.Host
	snapshot func() (*config.Snapshot, error)
	jitter   time.Duration
	log      *slog.Logger
	DBPath   string
}

// New creates a scheduler. snapshot is re-read on every tick.
func New(st *store.Store, host *plugin.Host, snapshot func() (*config.Snapshot, error)) *Scheduler {
	return &Scheduler{store: st, host: host, snapshot: snapshot, jitter: 15 * time.Minute, log: slog.Default()}
}

// Run owns the tick loop until ctx is cancelled.
func (s *Scheduler) Run(ctx context.Context) error {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			func() {
				defer func() {
					if r := recover(); r != nil {
						s.log.Error("auto-update panic recovered", "panic", r)
					}
				}()
				s.Tick(ctx, time.Now())
			}()
		}
	}
}

// Tick runs one maintenance pass, including its due check.
func (s *Scheduler) Tick(ctx context.Context, now time.Time) {
	snap, err := s.snapshot()
	if err != nil {
		s.log.Error("auto-update config", "error", err)
		return
	}
	interval := time.Duration(snap.Int("updates.update_interval_days")) * 24 * time.Hour
	if !s.due(ctx, interval, now) {
		return
	}
	if s.jitter > 0 {
		time.Sleep(time.Duration(rand.Int63n(int64(s.jitter))))
	}
	capPer := int(snap.Int("updates.update_per_plugin_limit"))
	candidates, err := s.candidates(ctx, now.Unix(), capPer)
	if err != nil {
		s.log.Error("auto-update candidates", "error", err)
		return
	}
	checked, okCount, errCount := 0, 0, 0
	delay := time.Duration(snap.Int("updates.update_request_delay_ms")) * time.Millisecond
	for i, c := range candidates {
		if ctx.Err() != nil {
			return
		}
		checked++
		if err := s.refreshOne(ctx, c.Plugin, c.Slug, now.Unix()); err != nil {
			errCount++
			s.log.Error("auto-update failed", "plugin", c.Plugin, "slug", c.Slug, "error", err)
		} else {
			okCount++
		}
		if d := paceDelay(i, len(candidates), delay); d > 0 {
			time.Sleep(d)
		}
	}
	_ = s.store.PutPluginState(ctx, "update", "last_run", now.Format(time.RFC3339))
	s.log.Info("auto-update run", "checked", checked, "ok", okCount, "error", errCount)
	s.maintenance(ctx, snap, now)
}

// paceDelay returns the pause before the next title, never after the last.
func paceDelay(index, count int, delay time.Duration) time.Duration {
	if count < 2 || index >= count-1 {
		return 0
	}
	return delay
}

func (s *Scheduler) due(ctx context.Context, interval time.Duration, now time.Time) bool {
	raw, err := s.store.GetPluginState(ctx, "update", "last_run")
	if err != nil || raw == "" {
		return true
	}
	at, err := time.Parse(time.RFC3339, raw)
	return err != nil || !now.Before(at.Add(interval))
}

func (s *Scheduler) candidates(ctx context.Context, now int64, capPer int) ([]store.UpdateCandidate, error) {
	rows, err := s.store.UpdateCandidates(ctx, now, 1000)
	if err != nil {
		return nil, err
	}
	counts := make(map[string]int)
	out := make([]store.UpdateCandidate, 0, len(rows))
	for _, row := range rows {
		if counts[row.Plugin] >= capPer {
			continue
		}
		counts[row.Plugin]++
		out = append(out, row)
	}
	return out, nil
}

func (s *Scheduler) refreshOne(ctx context.Context, pluginID, slug string, now int64) error {
	params, _ := json.Marshal(map[string]string{"slug": slug, "id": slug})
	raw, err := s.host.Call(ctx, pluginID, "chapters", params)
	if err != nil {
		StampAttempt(ctx, s.store, pluginID, slug, false, err.Error(), now)
		return err
	}
	var env struct {
		OK   bool `json:"ok"`
		Data struct {
			Chapters []store.Chapter `json:"chapters"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &env); err != nil || !env.OK {
		StampAttempt(ctx, s.store, pluginID, slug, false, "invalid envelope", now)
		return fmt.Errorf("invalid chapters envelope")
	}
	chapters := env.Data.Chapters
	for i := range chapters {
		chapters[i].Plugin, chapters[i].Slug = pluginID, slug
	}
	if err := s.store.UpsertChapters(ctx, pluginID, slug, chapters); err != nil {
		StampAttempt(ctx, s.store, pluginID, slug, false, err.Error(), now)
		return err
	}
	growth := false
	var saved store.SavedManga
	if err := s.store.DB().WithContext(ctx).Where("plugin = ? AND slug = ?", pluginID, slug).First(&saved).Error; err == nil {
		growth = len(chapters) > saved.TotalCh
		updates := map[string]any{"total_ch": len(chapters)}
		if growth {
			updates["status_flags"] = gorm.Expr("status_flags | ?", store.StatusNew)
		}
		_ = s.store.DB().WithContext(ctx).Model(&store.SavedManga{}).
			Where("plugin = ? AND slug = ?", pluginID, slug).Updates(updates).Error
	}
	StampAttempt(ctx, s.store, pluginID, slug, true, "", now)
	return nil
}

func (s *Scheduler) maintenance(ctx context.Context, snap *config.Snapshot, now time.Time) {
	if err := s.store.PruneExpiredCache(ctx, now.Unix()); err != nil {
		s.log.Error("auto-update cache prune", "error", err)
	}
	dir := filepath.Dir(s.DBPath)
	mark, err := s.store.BackupWatermark(ctx)
	if err == nil && now.Sub(mark) >= time.Duration(snap.Int("updates.backup_interval_h"))*time.Hour {
		if err := s.backup(ctx, dir, int(snap.Int("updates.backup_keep")), now); err != nil {
			s.log.Error("auto-update backup", "error", err)
		}
	}
	if err := s.store.RecomputeDupGroups(ctx); err != nil {
		s.log.Error("auto-update recluster", "error", err)
	}
}

func (s *Scheduler) backup(ctx context.Context, dir string, keep int, now time.Time) error {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	name := now.Format("20060102-150405") + ".db"
	if err := s.store.BackupInto(ctx, filepath.Join(dir, name)); err != nil {
		return err
	}
	_ = s.store.PutPluginState(ctx, "backup", "last_run", now.Format(time.RFC3339))
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".db") {
			files = append(files, entry.Name())
		}
	}
	for i := 0; i < len(files)-keep; i++ {
		_ = os.Remove(filepath.Join(dir, files[i]))
	}
	return nil
}

// StampAttempt records one scheduled or manual attempt with backoff.
func StampAttempt(ctx context.Context, st *store.Store, pluginID, slug string, ok bool, errText string, now int64) {
	prev, err := st.GetUpdateAttempt(ctx, pluginID, slug)
	if err != nil {
		prev = &store.UpdateAttempt{Plugin: pluginID, Slug: slug}
	}
	count, until := int64(0), int64(0)
	if !ok {
		count = int64(prev.ConsecutiveErrors) + 1
		hours := math.Pow(2, float64(count))
		if hours > 72 {
			hours = 72
		}
		until = now + int64(hours*3600)
	}
	_ = st.PutUpdateAttempt(ctx, &store.UpdateAttempt{
		Plugin: pluginID, Slug: slug, LastAttemptAt: now,
		Outcome: outcome(ok), Error: errText, ConsecutiveErrors: int(count), BackoffUntil: until,
	})
}

func outcome(ok bool) string {
	if ok {
		return "ok"
	}
	return "error"
}
