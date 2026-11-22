package components

import (
	"context"
	"io"

	"github.com/a-h/templ"
)

// rawFragment writes already-sanitized HTML from the embedded README converter.
func rawFragment(html string) templ.Component {
	return templ.ComponentFunc(func(_ context.Context, w io.Writer) error {
		_, err := io.WriteString(w, html)
		return err
	})
}
