package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
)

// PagePath returns the concept-v3 on-disk path for one chapter page.
// Chapter numbers are zero-padded to six digits, an id suffix breaks
// duplicate-number collisions, and the manga directory is bounded and
// hash-suffixed so long or colliding slugs stay distinct.
func PagePath(root, namespace, slug string, number float64, chapterID string, page int, ext string) string {
	ch := chapterDir(number, chapterID)
	return filepath.Join(root, "img", namespace, "manga", SlugDir(slug), ch,
		fmt.Sprintf("p%03d.%s", page, ext))
}

// chapterDir formats a chapter folder as chNNN with an optional last-eight
// chapter-id suffix. Fractional chapter numbers keep their decimal part.
func chapterDir(number float64, chapterID string) string {
	s := strconv.FormatFloat(number, 'f', -1, 64)
	parts := strings.SplitN(s, ".", 2)
	ch := "ch" + fmt.Sprintf("%06s", parts[0])
	if len(parts) == 2 {
		ch += "." + parts[1]
	}
	if chapterID != "" {
		id8 := chapterID
		if len(id8) > 8 {
			id8 = id8[len(id8)-8:]
		}
		ch += "-" + id8
	}
	return ch
}

// CoverPath returns the concept-v3 on-disk path for one cover image. The
// fixed-point ratio rides in the filename so no sidecar is needed.
func CoverPath(root, namespace, hash string, ratio float64, ext string) string {
	r4 := 0
	if ratio > 0 {
		r4 = int(math.Round(ratio * 10000))
	}
	return filepath.Join(root, "img", namespace, "cover",
		fmt.Sprintf("%s-%d.%s", hash, r4, ext))
}

// SlugDir bounds a manga slug to 40 characters and appends an eight-character
// hash of the full slug so distinct long titles never share a directory.
func SlugDir(slug string) string {
	slug = NormalizeSlug(slug)
	runes := []rune(slug)
	if len(runes) > 40 {
		runes = runes[:40]
	}
	sum := sha256.Sum256([]byte(slug))
	return string(runes) + "-" + hex.EncodeToString(sum[:])[:8]
}

// Namespace derives a stable cache namespace from a plugin's display name,
// falling back to the plugin id when the name is blank.
func Namespace(pluginID, displayName string) string {
	if name := NormalizeSlug(displayName); name != "" {
		return name
	}
	return pluginID
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
