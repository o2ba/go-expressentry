package clean

import (
	"testing"
	"time"
)

func TestRecoverTimestamp(t *testing.T) {
	for _, tc := range []struct {
		round    string
		value    string
		drawDate string
		want     time.Time
	}{
		{"333", "January 23, 2025 at 2025-01-23 15:30:04 UTC", "2025-01-23", time.Date(2025, 1, 23, 15, 30, 4, 0, time.UTC)},
		{"322", "March 02. 2024 at 01:58:34 UTC", "2024-12-16", time.Date(2024, 3, 2, 1, 58, 34, 0, time.UTC)},
		{"310", "August 15, 2024 at 2024-08-15 15:11:54 UTC", "2024-08-15", time.Date(2024, 8, 15, 15, 11, 54, 0, time.UTC)},
		{"297", "May 31, 2024 at12:48:30 UTC", "2024-05-31", time.Date(2024, 5, 31, 12, 48, 30, 0, time.UTC)},
		{"277", "August 30. 2023 at 17:30:21 UTC", "2023-12-06", time.Date(2023, 8, 30, 17, 30, 21, 0, time.UTC)},
		{"270", "October 25, 2023 at 15:48:39 AM UTC", "2023-10-25", time.Date(2023, 10, 25, 15, 48, 39, 0, time.UTC)},
		{"270", "September 30, 2023 at 04:22:17 AM UTC", "2023-10-25", time.Date(2023, 9, 30, 4, 22, 17, 0, time.UTC)},
		{"210", "June 22, 2021 at 20:19:59 PM UTC", "2021-11-10", time.Date(2021, 6, 22, 20, 19, 59, 0, time.UTC)},
		{"208", "October 7 at 16:38:21 UTC", "2021-10-13", time.Date(2021, 10, 7, 16, 38, 21, 0, time.UTC)},
		{"12-hour PM", "June 22, 2021 at 08:19:59 PM UTC", "2021-11-10", time.Date(2021, 6, 22, 20, 19, 59, 0, time.UTC)},
		{"12-hour midnight", "June 22, 2021 at 12:19:59 AM UTC", "2021-11-10", time.Date(2021, 6, 22, 0, 19, 59, 0, time.UTC)},
		{"no year, previous year", "December 30 at 09:00:00 UTC", "2022-01-05", time.Date(2021, 12, 30, 9, 0, 0, 0, time.UTC)},
	} {
		got, ok := RecoverTimestamp(tc.value, tc.drawDate)
		if !ok || !got.Equal(tc.want) {
			t.Errorf("round %s: RecoverTimestamp(%q) = %v, %v; want %v", tc.round, tc.value, got, ok, tc.want)
		}
	}
}

func TestRecoverTimestampGivesUp(t *testing.T) {
	for _, tc := range []struct{ value, drawDate string }{
		{"sometime", "2026-01-01"},
		{"October 7 at 16:38:21 UTC", ""},
		{"October 7 at 16:38:21 UTC", "soon"},
	} {
		if got, ok := RecoverTimestamp(tc.value, tc.drawDate); ok {
			t.Errorf("RecoverTimestamp(%q, %q) = %v, want no time", tc.value, tc.drawDate, got)
		}
	}
}

func TestStrictTimestampStillRejectsTypos(t *testing.T) {
	for _, s := range []string{
		"January 23, 2025 at 2025-01-23 15:30:04 UTC",
		"March 02. 2024 at 01:58:34 UTC",
		"May 31, 2024 at12:48:30 UTC",
		"October 25, 2023 at 15:48:39 AM UTC",
		"October 7 at 16:38:21 UTC",
	} {
		if got, ok := Timestamp(s); ok {
			t.Errorf("Timestamp(%q) = %v, want it left to RecoverTimestamp", s, got)
		}
	}
}
