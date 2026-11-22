package cache

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
)

// PagePath returns the v3 on-disk path for one chapter page.
func PagePath(root, plugin, slug string, number float64, chapterID string, page int, ext string) string {
	id8 := chapterID
	if len(id8) > 8 {
		id8 = id8[:8]
	}
	chapter := "ch" + strconv.FormatFloat(number, 'f', -1, 64) + "-" + id8
	return filepath.Join(root, "img", plugin, "manga", NormalizeSlug(slug), chapter,
		fmt.Sprintf("p%03d.%s", page, ext))
}

// CoverPath returns the v3 on-disk path for one cover image.
func CoverPath(root, plugin, hash, ext string) string {
	return filepath.Join(root, "img", plugin, "cover", hash+"."+ext)
}

// APIPath returns the v3 on-disk path for one cached API payload.
func APIPath(root, plugin, method, hash string) string {
	return filepath.Join(root, "api", plugin, method, hash+".json")
}

// NormalizeSlug derives a filesystem-safe slug from a display name.
func NormalizeSlug(name string) string {
	var out strings.Builder
	dash := false
	for _, r := range []rune(strings.ToLower(strings.TrimSpace(name))) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			out.WriteRune(r)
			dash = false
		case !dash:
			out.WriteRune('-')
			dash = true
		}
	}
	return strings.Trim(out.String(), "-")
}
