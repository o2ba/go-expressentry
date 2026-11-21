package clean

import "strings"

// rules map words in IRCC's English round name to a category ID. The first
// match wins, so narrower names come before ones they contain ("Senior
// managers with Canadian Work Experience" isn't the Canadian Experience
// Class).
var rules = []struct {
	words    []string
	category string
}{
	{[]string{"physician"}, "physicians"},
	{[]string{"senior manager"}, "managers"},
	{[]string{"military"}, "military"},
	{[]string{"french"}, "french"},
	{[]string{"healthcare", "health care"}, "healthcare"},
	{[]string{"transport"}, "transport"},
	{[]string{"trade"}, "trades"},
	{[]string{"stem", "science, technology"}, "stem"},
	{[]string{"education"}, "education"},
	{[]string{"provincial nominee"}, "pnp"},
	{[]string{"canadian experience class"}, "cec"},
	{[]string{"no program specified", "general", "all programs"}, "general"},
}

// Category maps IRCC's English round name ("Trades Occupations,
// 2026-Version 3") to a category ID, or "other" when nothing matches.
func Category(name string) string {
	lower := strings.ToLower(name)
	for _, r := range rules {
		for _, w := range r.words {
			if containsWord(lower, w) {
				return r.category
			}
		}
	}
	return "other"
}

// containsWord reports whether w appears in s starting at a word boundary,
// so "stem" doesn't match "system".
func containsWord(s, w string) bool {
	for i := 0; ; {
		j := strings.Index(s[i:], w)
		if j < 0 {
			return false
		}
		at := i + j
		if at == 0 || !isLetter(s[at-1]) {
			return true
		}
		i = at + 1
	}
}

func isLetter(b byte) bool { return b >= 'a' && b <= 'z' }
