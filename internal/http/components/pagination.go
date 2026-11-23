package components

import "strconv"

// PageWindow returns a compact page list: the first four and last five pages
// are always visible, with an ellipsis standing in for any skipped range. The
// page around the current selection is included when it falls in the middle.
func PageWindow(current, total int) []string {
	if total <= 9 {
		out := make([]string, total)
		for i := range out {
			out[i] = strconv.Itoa(i + 1)
		}
		return out
	}
	seen := map[int]bool{}
	out := make([]string, 0, 13)
	add := func(n int) {
		if n > 0 && n <= total && !seen[n] {
			seen[n] = true
			out = append(out, strconv.Itoa(n))
		}
	}
	gap := func() {
		if len(out) == 0 || out[len(out)-1] != "…" {
			out = append(out, "…")
		}
	}
	for n := 1; n <= 4; n++ {
		add(n)
	}
	if current <= 6 {
		gap()
		for n := total - 4; n <= total; n++ {
			add(n)
		}
	} else if current >= total-5 {
		gap()
		for n := total - 4; n <= total; n++ {
			add(n)
		}
	} else {
		gap()
		add(current - 1)
		add(current)
		add(current + 1)
		gap()
		for n := total - 4; n <= total; n++ {
			add(n)
		}
	}
	return out
}
