package tools

import "regexp"

// ponytail: every call compiles afresh; add a compiled-pattern cache keyed by
// pattern when scrape profiles show the recompile cost.

// Match reports whether pattern matches anywhere in text.
func Match(pattern, text string) (bool, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false, err
	}
	return re.MatchString(text), nil
}

// Find returns the leftmost match of pattern in text, or "" when nothing
// matches; a pattern that does not compile is an error.
func Find(pattern, text string) (string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	}
	return re.FindString(text), nil
}

// Capture returns the first capture group of the leftmost match, or the whole
// match when the pattern has no groups; it returns "" when nothing matches.
func Capture(pattern, text string) (string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	}
	m := re.FindStringSubmatch(text)
	if m == nil {
		return "", nil
	}
	if re.NumSubexp() == 0 {
		return m[0], nil
	}
	return m[1], nil
}

// FindAll returns every non-overlapping match in text, at most limit of them
// (limit < 0: all, 0: none), or nil when nothing matches.
func FindAll(pattern, text string, limit int) ([]string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return re.FindAllString(text, limit), nil
}

// Replace replaces every match of pattern in text with repl, where $1 and
// ${name} references expand to captures; repl is not taken literally.
func Replace(pattern, text, repl string) (string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	}
	return re.ReplaceAllString(text, repl), nil
}

// Quote escapes every regex metacharacter in s so the result matches s
// literally.
func Quote(s string) string { return regexp.QuoteMeta(s) }

// FindIndex returns the half-open byte offsets [start,end) of the leftmost
// match of pattern in text, or nil when nothing matches.
func FindIndex(pattern, text string) ([]int, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return re.FindStringIndex(text), nil
}
