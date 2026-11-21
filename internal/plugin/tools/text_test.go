package tools

import "testing"

func TestTextHelpers(t *testing.T) {
	t.Run("URLEncode", func(t *testing.T) {
		for _, tc := range []struct{ in, want string }{
			{"one two&three=4", "one+two%26three%3D4"},
			{"héllo/wörld", "h%C3%A9llo%2Fw%C3%B6rld"},
		} {
			if got := URLEncode(tc.in); got != tc.want {
				t.Errorf("URLEncode(%q)=%q want %q", tc.in, got, tc.want)
			}
		}
	})

	t.Run("URLDecode", func(t *testing.T) {
		for _, tc := range []struct {
			in   string
			want string
		}{
			{"one+two%26three%3D4", "one two&three=4"},
			{"%E4%BD%A0%E5%A5%BD", "你好"},
		} {
			got, err := URLDecode(tc.in)
			if err != nil {
				t.Errorf("URLDecode(%q): %v", tc.in, err)
			}
			if got != tc.want {
				t.Errorf("URLDecode(%q)=%q want %q", tc.in, got, tc.want)
			}
		}
		for _, bad := range []string{"%zz", "trailing%"} {
			if _, err := URLDecode(bad); err == nil {
				t.Errorf("URLDecode(%q): want error", bad)
			}
		}
	})

	t.Run("JSONBlob", func(t *testing.T) {
		for _, tc := range []struct {
			in   any
			want string
		}{{nil, "null"}, {map[string]int{"b": 2, "a": 1}, `{"a":1,"b":2}`}} {
			if got := JSONBlob(tc.in); got != tc.want {
				t.Errorf("JSONBlob(%v)=%q want %q", tc.in, got, tc.want)
			}
		}
		m := map[string]any{"z": 1, "a": map[string]int{"y": 2, "x": 3}}
		if JSONBlob(m) != JSONBlob(m) || JSONBlob(m) != `{"a":{"x":3,"y":2},"z":1}` {
			t.Errorf("JSONBlob not stable/sorted: %q", JSONBlob(m))
		}
		if got := JSONBlob(make(chan int)); got != "" {
			t.Errorf("JSONBlob(channel)=%q want \"\"", got)
		}
	})

	t.Run("DateEpoch", func(t *testing.T) {
		for _, tc := range []struct {
			in   string
			want int64
		}{
			{"2026-11-22T00:00:00Z", 1795305600},
			{"2026-11-22", 1795305600},
			{"2026-11-22T12:34:56+07:00", 1795325696},
		} {
			got, err := DateEpoch(tc.in)
			if err != nil {
				t.Errorf("DateEpoch(%q): %v", tc.in, err)
			}
			if got != tc.want {
				t.Errorf("DateEpoch(%q)=%d want %d", tc.in, got, tc.want)
			}
		}
		if _, err := DateEpoch("22/11/2026"); err == nil {
			t.Error("DateEpoch(bad): want error")
		}
	})

	t.Run("Clean", func(t *testing.T) {
		for _, tc := range []struct{ in, want string }{
			{"  a \t\t b \n c  ", "a b c"},
			{"a\x00\x01b", "ab"},
			{"　full-width　", "full-width"},
			{"héllo wörld 你好", "héllo wörld 你好"},
			{"", ""},
		} {
			if got := Clean(tc.in); got != tc.want {
				t.Errorf("Clean(%q)=%q want %q", tc.in, got, tc.want)
			}
		}
	})
}
