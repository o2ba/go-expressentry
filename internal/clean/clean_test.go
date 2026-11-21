package clean

import (
	"testing"
	"time"
)

func TestCategory(t *testing.T) {
	for name, want := range map[string]string{
		"Canadian Experience Class":                                     "cec",
		"French-Language proficiency 2026-Version 2":                    "french",
		"Provincial Nominee Program":                                    "pnp",
		"Healthcare and Social Services Occupations, 2026-Version 3":    "healthcare",
		"Trades Occupations, 2026-Version 3":                            "trades",
		"Federal Skilled Trades":                                        "trades",
		"Transport Occupations, 2026-Version 2":                         "transport",
		"Senior managers with Canadian Work Experience, 2026-Version 1": "managers",
		"Physicians with Canadian Work Experience, 2026-Version 1":      "physicians",
		"Skilled military recruits, 2026-Version 1":                     "military",
		"STEM occupations (Version 1)":                                  "stem",
		"Education occupations (Version 1)":                             "education",
		"No Program Specified":                                          "general",
		"Federal Skilled Worker":                                        "other",
		"Express Entry system":                                          "other",
	} {
		if got := Category(name); got != want {
			t.Errorf("Category(%q) = %s, want %s", name, got, want)
		}
	}
}

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

func TestCompareRoundNumbers(t *testing.T) {
	order := []string{"90", "91a", "91b", "92", "453"}
	for i := 1; i < len(order); i++ {
		if CompareRoundNumbers(order[i-1], order[i]) >= 0 {
			t.Errorf("%s not before %s", order[i-1], order[i])
		}
	}
}
