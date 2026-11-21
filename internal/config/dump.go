package config

import (
	"fmt"
	"io"
	"strings"
)

// Dump writes the registry as an ini document: each section once, each key
// preceded by its doc as a `; ` comment and carrying its default value. The
// output is generated from the registry only, so no default is ever typed
// twice. Section and key order follow the registry declaration order.
func (r *Registry) Dump(w io.Writer) error {
	section := ""
	first := true
	for _, e := range r.Entries() {
		if e.Section != section {
			if !first {
				if _, err := fmt.Fprintln(w); err != nil {
					return err
				}
			}
			if _, err := fmt.Fprintf(w, "[%s]\n", e.Section); err != nil {
				return err
			}
			section = e.Section
			first = false
		}
		line := e.Key + " ="
		if e.Default != "" {
			line += " " + e.Default
		}
		if _, err := fmt.Fprintf(w, "; %s\n%s\n", e.Doc, line); err != nil {
			return err
		}
	}
	return nil
}

// DumpString returns Dump's output as a string.
func (r *Registry) DumpString() string {
	var b strings.Builder
	_ = r.Dump(&b)
	return b.String()
}
