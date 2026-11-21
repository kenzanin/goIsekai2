package tools

import (
	"reflect"
	"testing"
)

func TestRegexHelpers(t *testing.T) {
	t.Run("Match", func(t *testing.T) {
		for _, tc := range []struct {
			pattern, text string
			want          bool
			wantErr       bool
		}{
			{"a+", "xaax", true, false},
			{"^b", "ab", false, false},
			{"(", "x", false, true},
		} {
			got, err := Match(tc.pattern, tc.text)
			if (err != nil) != tc.wantErr {
				t.Fatalf("Match(%q,%q) err=%v wantErr=%v", tc.pattern, tc.text, err, tc.wantErr)
			}
			if err == nil && got != tc.want {
				t.Errorf("Match(%q,%q)=%v want %v", tc.pattern, tc.text, got, tc.want)
			}
		}
	})

	t.Run("Find", func(t *testing.T) {
		for _, tc := range []struct {
			pattern, text string
			want          string
			wantErr       bool
		}{
			{"a.c", "xxabcxx", "abc", false},
			{"z", "abc", "", false},
			{"(", "x", "", true},
		} {
			got, err := Find(tc.pattern, tc.text)
			if (err != nil) != tc.wantErr {
				t.Fatalf("Find(%q,%q) err=%v wantErr=%v", tc.pattern, tc.text, err, tc.wantErr)
			}
			if err == nil && got != tc.want {
				t.Errorf("Find(%q,%q)=%q want %q", tc.pattern, tc.text, got, tc.want)
			}
		}
	})

	t.Run("Capture", func(t *testing.T) {
		for _, tc := range []struct {
			pattern, text string
			want          string
			wantErr       bool
		}{
			{`name="(\w+)"`, `<a name="go">`, "go", false},
			{"ab", "cab", "ab", false},  // no groups: whole match
			{"(x)(y)", "ab", "", false}, // no match
			{"(", "x", "", true},
		} {
			got, err := Capture(tc.pattern, tc.text)
			if (err != nil) != tc.wantErr {
				t.Fatalf("Capture(%q,%q) err=%v wantErr=%v", tc.pattern, tc.text, err, tc.wantErr)
			}
			if err == nil && got != tc.want {
				t.Errorf("Capture(%q,%q)=%q want %q", tc.pattern, tc.text, got, tc.want)
			}
		}
	})

	t.Run("FindAll", func(t *testing.T) {
		for _, tc := range []struct {
			pattern, text string
			limit         int
			want          []string
			wantErr       bool
		}{
			{`\d+`, "a1b22c333", -1, []string{"1", "22", "333"}, false},
			{`\d+`, "a1b22c333", 2, []string{"1", "22"}, false},
			{`\d+`, "a1b22c333", 0, nil, false},
			{`\d+`, "abc", -1, nil, false},
			{"(", "x", -1, nil, true},
		} {
			got, err := FindAll(tc.pattern, tc.text, tc.limit)
			if (err != nil) != tc.wantErr {
				t.Fatalf("FindAll(%q,%q,%d) err=%v wantErr=%v", tc.pattern, tc.text, tc.limit, err, tc.wantErr)
			}
			if err == nil && !reflect.DeepEqual(got, tc.want) {
				t.Errorf("FindAll(%q,%q,%d)=%v want %v", tc.pattern, tc.text, tc.limit, got, tc.want)
			}
		}
	})

	t.Run("Replace", func(t *testing.T) {
		for _, tc := range []struct{ pattern, text, repl, want string }{
			{`(a+)b`, "aab ab", "<$1>", "<aa> <a>"},
			{"o", "foo", "0", "f00"},
		} {
			got, err := Replace(tc.pattern, tc.text, tc.repl)
			if err != nil {
				t.Fatalf("Replace(%q,%q,%q): %v", tc.pattern, tc.text, tc.repl, err)
			}
			if got != tc.want {
				t.Errorf("Replace(%q,%q,%q)=%q want %q", tc.pattern, tc.text, tc.repl, got, tc.want)
			}
		}
		if _, err := Replace("(", "x", "y"); err == nil {
			t.Error("Replace(invalid): want error")
		}
	})

	t.Run("Quote", func(t *testing.T) {
		if got, want := Quote("1.5+1"), `1\.5\+1`; got != want {
			t.Errorf("Quote=%q want %q", got, want)
		}
		if ok, err := Match(Quote("1.5+1"), "a1.5+1b"); err != nil || !ok {
			t.Errorf("quoted literal not literal: ok=%v err=%v", ok, err)
		}
	})

	t.Run("FindIndex", func(t *testing.T) {
		for _, tc := range []struct {
			pattern, text string
			want          []int
			wantErr       bool
		}{
			{"b+", "abbc", []int{1, 3}, false},
			{"z", "abc", nil, false},
			{"(", "x", nil, true},
		} {
			got, err := FindIndex(tc.pattern, tc.text)
			if (err != nil) != tc.wantErr {
				t.Fatalf("FindIndex(%q,%q) err=%v wantErr=%v", tc.pattern, tc.text, err, tc.wantErr)
			}
			if err == nil && !reflect.DeepEqual(got, tc.want) {
				t.Errorf("FindIndex(%q,%q)=%v want %v", tc.pattern, tc.text, got, tc.want)
			}
		}
	})
}

func TestRegexUnicodeIsRuneAware(t *testing.T) {
	ok, err := Match(`\p{Han}+`, "日本語 manga")
	if err != nil || !ok {
		t.Fatalf("Han match = %v, %v", ok, err)
	}
	got, err := Find(`.`, "日本")
	if err != nil || got != "日" {
		t.Fatalf("dot match = %q, %v", got, err)
	}
}
