package corrections

import (
	"testing"

	"github.com/o2ba/go-expressentry/internal/raw"
)

func TestApply(t *testing.T) {
	list := []Correction{{Lang: French, Round: "9", Field: "drawName", Wrong: "nuef", Right: "neuf"}}
	for _, tc := range []struct {
		name  string
		lang  Lang
		round raw.Round
		want  string
		made  int
	}{
		{"wrong value", French, raw.Round{DrawNumber: "9", DrawName: "nuef"}, "neuf", 1},
		{"already fixed", French, raw.Round{DrawNumber: "9", DrawName: "neuf (révisé)"}, "neuf (révisé)", 0},
		{"other file", English, raw.Round{DrawNumber: "9", DrawName: "nuef"}, "nuef", 0},
		{"other round", French, raw.Round{DrawNumber: "8", DrawName: "nuef"}, "nuef", 0},
	} {
		r := tc.round
		made := Apply(list, tc.lang, &r)
		if r.DrawName != tc.want || len(made) != tc.made {
			t.Errorf("%s: DrawName = %q, made %d; want %q, %d", tc.name, r.DrawName, len(made), tc.want, tc.made)
		}
	}
}

func TestAllNameKnownFields(t *testing.T) {
	for _, c := range All {
		if field(&raw.Round{}, c.Field) == nil || c.Wrong == c.Right || c.Why == "" || (c.Lang != English && c.Lang != French) {
			t.Errorf("bad correction %+v", c)
		}
	}
}
