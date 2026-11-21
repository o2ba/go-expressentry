// Package ircctime reads the dates and times IRCC types by hand in its
// rounds files. Timestamp and Day accept the usual variations (ordinal
// suffixes, stray commas and spaces, a missing "UTC"); Repair reads the
// known typos in old rounds that they reject. Every time is UTC.
package ircctime

import (
	"regexp"
	"strings"
	"time"
)

var (
	ordinal = regexp.MustCompile(`(\d)(st|nd|rd|th)\b`)
	spaces  = regexp.MustCompile(`\s+`)
)

// normalize undoes the ways IRCC's hand-typed English dates vary: ordinal
// suffixes ("August 31st"), missing or extra commas and spaces, and a
// trailing "UTC" that is sometimes left out (the times are always UTC).
func normalize(s string) string {
	s = strings.ReplaceAll(s, "\u00a0", " ")
	s = ordinal.ReplaceAllString(s, "$1")
	s = strings.ReplaceAll(s, ",", " ")
	s = strings.ReplaceAll(s, " at ", " ")
	s = strings.TrimSuffix(strings.TrimSpace(spaces.ReplaceAllString(s, " ")), " UTC")
	return strings.TrimSpace(s)
}

// Timestamp parses an English time like "October 28, 2026 at 10:34:14 UTC",
// in UTC.
func Timestamp(s string) (time.Time, bool) {
	t, err := time.ParseInLocation("January 2 2006 15:04:05", normalize(s), time.UTC)
	return t, err == nil
}

// Day parses an English date like "October 25, 2026". The result is
// midnight UTC; only its calendar day means anything.
func Day(s string) (time.Time, bool) {
	t, err := time.Parse("January 2 2006", normalize(s))
	return t, err == nil
}
