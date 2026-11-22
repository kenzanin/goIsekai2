package store

import (
	"context"
	"sort"
	"strings"
	"unicode"

	"gorm.io/gorm"

	sqlb "goisekai2/internal/store/sql"
)

// SearchTitles returns BM25-ranked local results (or LIKE fallback).
func (s *Store) SearchTitles(ctx context.Context, term string) ([]sqlb.SearchHit, error) {
	return sqlb.Search(s.db.WithContext(ctx), term, s.ftsOK)
}

// IndexSavedTitle keeps the FTS row in sync with a saved-title write.
func (s *Store) IndexSavedTitle(ctx context.Context, m *SavedManga) error {
	if !s.ftsOK {
		return nil
	}
	return sqlb.IndexTitle(s.db.WithContext(ctx), m.Plugin, m.Slug, m.Name, m.AltTitles, m.Synopsis)
}

// DeleteSavedTitle removes the FTS row when a title leaves the library.
func (s *Store) DeleteSavedTitle(ctx context.Context, plugin, slug string) error {
	if !s.ftsOK {
		return nil
	}
	return sqlb.DeleteTitle(s.db.WithContext(ctx), plugin, slug)
}

func tokens(text string) map[string]int {
	out := make(map[string]int)
	var word strings.Builder
	flush := func() {
		if word.Len() == 0 {
			return
		}
		out[strings.ToLower(word.String())]++
		word.Reset()
	}
	for _, r := range []rune(text) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			word.WriteRune(r)
		} else {
			flush()
		}
	}
	flush()
	return out
}

func dice(a, b map[string]int) float64 {
	if len(a)+len(b) == 0 {
		return 0
	}
	shared := 0
	for k, count := range a {
		if other, ok := b[k]; ok {
			if count < other {
				shared += count
			} else {
				shared += other
			}
		}
	}
	return 2 * float64(shared) / float64(len(a)+len(b))
}

type dsU struct{ parent map[int]int }

func (u *dsU) find(i int) int {
	for u.parent[i] != i {
		u.parent[i] = u.parent[u.parent[i]]
		i = u.parent[i]
	}
	return i
}

func (u *dsU) union(a, b int) {
	ra, rb := u.find(a), u.find(b)
	if ra != rb {
		u.parent[rb] = ra
	}
}

// RecomputeDupGroups rebuilds dup_group from current data. It never trusts
// values stored on disk and writes every group in one transaction.
func (s *Store) RecomputeDupGroups(ctx context.Context) error {
	var all []SavedManga
	if err := s.db.WithContext(ctx).Where("added_at > 0").Find(&all).Error; err != nil {
		return err
	}
	u := &dsU{parent: make(map[int]int, len(all))}
	for i := range all {
		u.parent[i] = i
	}
	for i := range all {
		hits, err := s.SearchTitles(ctx, strings.Fields(all[i].Name + " " + all[i].Synopsis)[0])
		if err != nil || len(hits) == 0 {
			continue
		}
		j := make(map[string]int, len(all))
		for k := range all {
			j[all[k].Plugin+"\x00"+all[k].Slug] = k
		}
		nameA := tokens(all[i].Name)
		synA := tokens(all[i].Synopsis + " " + all[i].Authors + " " + all[i].Artists)
		for _, hit := range hits {
			idx, ok := j[hit.Plugin+"\x00"+hit.Slug]
			if !ok || idx == i {
				continue
			}
			nameScore := dice(nameA, tokens(all[idx].Name))
			synScore := dice(synA, tokens(all[idx].Synopsis+" "+all[idx].Authors+" "+all[idx].Artists))
			if nameScore >= 0.75 || synScore >= 0.75 || 0.5*nameScore+0.5*synScore >= 0.6 {
				u.union(i, idx)
			}
		}
	}
	groups := make(map[int][]int)
	for i := range all {
		root := u.find(i)
		groups[root] = append(groups[root], i)
	}
	roots := make([]int, 0, len(groups))
	for root := range groups {
		roots = append(roots, root)
	}
	sort.Ints(roots)
	assign := make(map[int]int)
	next := 1
	for _, root := range roots {
		members := groups[root]
		if len(members) < 2 {
			assign[members[0]] = 0
			continue
		}
		for _, idx := range members {
			assign[idx] = next
		}
		next++
	}
	return s.Tx(ctx, func(txCtx context.Context, tx *gorm.DB) error {
		for i := range all {
			err := tx.WithContext(txCtx).Model(&SavedManga{}).
				Where("plugin = ? AND slug = ?", all[i].Plugin, all[i].Slug).
				Update("dup_group", assign[i]).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
}
