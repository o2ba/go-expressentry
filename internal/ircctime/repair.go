package ircctime

import (
	"regexp"
	"strconv"
	"time"
)

// typoFix rewrites one mistake IRCC made by hand in old rounds' times.
type typoFix struct {
	pattern *regexp.Regexp
	replace func(match []string) string
}

var typoFixes = []typoFix{
	{
		// "January 23, 2025 at 2025-01-23 15:30:04 UTC"
		pattern: regexp.MustCompile(`\bat\s+\d{4}-\d{2}-\d{2}\s+`),
		replace: func([]string) string { return "at " },
	},
	{
		// "March 02. 2024 at 01:58:34 UTC"
		pattern: regexp.MustCompile(`(\d)\.\s`),
		replace: func(m []string) string { return m[1] + " " },
	},
	{
		// "May 31, 2024 at12:48:30 UTC"
		pattern: regexp.MustCompile(`\bat(\d)`),
		replace: func(m []string) string { return "at " + m[1] },
	},
	{
		// "October 25, 2023 at 15:48:39 AM UTC", "September 30, 2023 at 04:22:17 AM UTC"
		pattern: regexp.MustCompile(`\b(\d{1,2})(:\d{2}:\d{2})\s*([AaPp][Mm])\b`),
		replace: meridiem,
	},
}

// meridiem turns a 12-hour time into a 24-hour one. IRCC sometimes put AM or
// PM after a time that was already on a 24-hour clock ("20:19:59 PM"); the
// hour wins then.
func meridiem(m []string) string {
	hour, _ := strconv.Atoi(m[1])
	if hour <= 12 {
		switch {
		case (m[3] == "PM" || m[3] == "pm") && hour < 12:
			hour += 12
		case (m[3] == "AM" || m[3] == "am") && hour == 12:
			hour = 0
		}
	}
	return strconv.Itoa(hour) + m[2]
}

// Repair reads times IRCC typed badly in old rounds, which
// Timestamp rejects. When IRCC left the year out ("October 7 at 16:38:21
// UTC"), it comes from the round's drawDate, a year earlier if that would
// put the time after the draw.
func Repair(s, drawDate string) (time.Time, bool) {
	for _, f := range typoFixes {
		s = f.pattern.ReplaceAllStringFunc(s, func(match string) string {
			return f.replace(f.pattern.FindStringSubmatch(match))
		})
	}
	if t, ok := Timestamp(s); ok {
		return t, true
	}

	draw, err := time.Parse(time.DateOnly, drawDate)
	if err != nil {
		return time.Time{}, false
	}
	t, err := time.ParseInLocation("January 2 15:04:05", normalize(s), time.UTC)
	if err != nil {
		return time.Time{}, false
	}
	t = t.AddDate(draw.Year(), 0, 0)
	if t.After(draw.AddDate(0, 0, 1)) {
		t = t.AddDate(-1, 0, 0)
	}
	return t, true
}
