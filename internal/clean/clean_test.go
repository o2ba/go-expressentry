package clean

import "testing"

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

func TestCompareRoundNumbers(t *testing.T) {
	order := []string{"90", "91a", "91b", "92", "453"}
	for i := 1; i < len(order); i++ {
		if CompareRoundNumbers(order[i-1], order[i]) >= 0 {
			t.Errorf("%s not before %s", order[i-1], order[i])
		}
	}
}
