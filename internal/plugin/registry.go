package plugin

// Registry indexes discovered manifests by plugin id.
type Registry struct {
	byID map[string]*Manifest
}

// NewRegistry indexes entries by id; on duplicate ids the first in the slice
// wins. Scan order is name order within a root, so merging roots keeps the
// earlier root's plugin, matching the cppIsekai cross-root rule.
func NewRegistry(entries []Entry) *Registry {
	r := &Registry{byID: make(map[string]*Manifest, len(entries))}
	for _, e := range entries {
		if _, ok := r.byID[e.ID]; !ok {
			r.byID[e.ID] = e.Manifest
		}
	}
	return r
}

// Lookup returns the manifest registered under id.
func (r *Registry) Lookup(id string) (*Manifest, bool) {
	m, ok := r.byID[id]
	return m, ok
}
