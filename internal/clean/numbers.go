package clean

import (
	"cmp"
	"strconv"
	"strings"
)

// CompareRoundNumbers orders round numbers by their digits, then by what
// follows them: 90 < 91a < 91b < 92 < 453.
func CompareRoundNumbers(a, b string) int {
	an, as := splitNumber(a)
	bn, bs := splitNumber(b)
	return cmp.Or(cmp.Compare(an, bn), cmp.Compare(as, bs))
}

func splitNumber(s string) (int, string) {
	i := strings.IndexFunc(s, func(r rune) bool { return r < '0' || r > '9' })
	if i < 0 {
		i = len(s)
	}
	n, _ := strconv.Atoi(s[:i])
	return n, s[i:]
}
