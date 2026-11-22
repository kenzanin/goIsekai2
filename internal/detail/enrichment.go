package detail

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"goisekai2/internal/plugin"
	"goisekai2/internal/plugin/tools"
)

// EnrichmentSource is one loaded enrichment-only plugin.
type EnrichmentSource struct {
	ID   string
	Name string
	rt   plugin.Runtime
}

// LoadEnrichment reads data/enrichment/<id>/main.json, keeping only manifests
// that declare the enrichment capability. Scraper folders are never loaded.
func LoadEnrichment(dir string, ts *tools.ToolSet) ([]*EnrichmentSource, error) {
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return nil, nil
	}
	entries, err := plugin.Discover(dir)
	if err != nil {
		return nil, err
	}
	out := make([]*EnrichmentSource, 0, len(entries))
	for _, entry := range entries {
		m := entry.Manifest
		if !hasCapability(m.Capabilities, "enrichment") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, entry.ID, m.Entry))
		if err != nil {
			return nil, err
		}
		var rt plugin.Runtime
		switch m.Runtime {
		case "lua":
			v, err := plugin.NewLuaRuntimeWithTools(string(raw), ts)
			if err != nil {
				return nil, err
			}
			if ts != nil {
				v.SetTools(ts)
			}
			rt = v
		case "js":
			v, err := plugin.NewJSRuntime(string(raw))
			if err != nil {
				return nil, err
			}
			if ts != nil {
				v.SetTools(ts)
			}
			rt = v
		default:
			continue
		}
		out = append(out, &EnrichmentSource{ID: entry.ID, Name: m.Name, rt: rt})
	}
	return out, nil
}

func hasCapability(caps []string, want string) bool {
	for _, c := range caps {
		if strings.EqualFold(c, want) {
			return true
		}
	}
	return false
}

// Enrich calls the source's enrich method and returns its data object.
// Envelope keys outside the enrichment contract are ignored by the caller.
func (s *EnrichmentSource) Enrich(ctx context.Context, params any) (map[string]any, error) {
	raw, _ := json.Marshal(params)
	env, _ := json.Marshal(map[string]any{"api": "host", "method": "enrich", "params": json.RawMessage(raw)})
	out, err := s.rt.Call(ctx, env)
	if err != nil {
		return nil, err
	}
	var parsed map[string]any
	if err := json.Unmarshal(out, &parsed); err != nil {
		return nil, err
	}
	if ok, _ := parsed["ok"].(bool); !ok {
		return nil, fmt.Errorf("enrichment %s: ok=false", s.ID)
	}
	data, _ := parsed["data"].(map[string]any)
	if data == nil {
		data = map[string]any{}
	}
	return data, nil
}
