package ircctime

import (
	"testing"
	"time"
)

func TestTimestampFormats(t *testing.T) {
	want := time.Date(2026, 8, 31, 10, 42, 13, 0, time.UTC)
	for _, s := range []string{
		"August 31, 2026 at 10:42:13 UTC",
		"August 31st, 2026 at 10:42:13 UTC",
		"August 31st 2026 at 10:42:13 UTC",
		"August 31, 2026 at 10:42:13",
		"August 31, 2026 at 10:42:13  UTC",
		"August 31, 2026 at  10:42:13 UTC",
	} {
		if got, ok := Timestamp(s); !ok || !got.Equal(want) {
			t.Errorf("Timestamp(%q) = %v, %v", s, got, ok)
		}
	}
	if got, ok := Timestamp("October 01, 2026 at 12:47:55 UTC"); !ok || got.Day() != 1 {
		t.Errorf("zero-padded day = %v, %v", got, ok)
	}
}

func TestDay(t *testing.T) {
	got, ok := Day("October 25th, 2026")
	if !ok || got.Year() != 2026 || got.Month() != time.October || got.Day() != 25 {
		t.Errorf("Day = %v, %v", got, ok)
	}
	if _, ok := Day("soon"); ok {
		t.Error("read soon as a day")
	}
}
