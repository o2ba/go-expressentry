package expressentry

import (
	"fmt"
	"strings"
	"testing"
)

func TestUnreadableEnglishRoundIsNotFrenchOnly(t *testing.T) {
	r := mustParse(t, `{"rounds":[{"drawNumber":"9","drawDate":"2026-01-01","drawName":"Canadian Experience Class"},{"drawNumber":"8","drawSize":"lots"}]}`,
		`{"rounds":[{"drawNumber":"9","drawName":"neuf"},{"drawNumber":"8","drawName":"huit"}]}`)
	for _, i := range r.Issues {
		if i.Kind == FrenchOnly {
			t.Errorf("round 8 is in English (unreadable) but reported %v", i)
		}
	}
}

func TestUnreadableFrenchCountIsNotAMismatch(t *testing.T) {
	r := mustParse(t, `{"rounds":[{"drawNumber":"9","drawDate":"2026-01-01","drawName":"Canadian Experience Class","dd5":"12342","dd18":"1"}]}`,
		`{"rounds":[{"drawNumber":"9","drawName":"neuf","dd5":"n/d","dd18":"1"}]}`)
	for _, i := range r.Issues {
		if i.Kind == FrenchMismatch {
			t.Errorf("unreadable French count reported as a value: %v", i)
		}
	}
}

func TestUnreadableEnglishCountIsNotAMismatch(t *testing.T) {
	r := mustParse(t, `{"rounds":[{"drawNumber":"9","drawDate":"2026-01-01","drawName":"Canadian Experience Class","dd5":"n/a","dd18":"1"}]}`,
		`{"rounds":[{"drawNumber":"9","drawName":"neuf","dd5":"12342","dd18":"1"}]}`)
	for _, i := range r.Issues {
		if i.Kind == FrenchMismatch {
			t.Errorf("unreadable English count reported as a value: %v", i)
		}
	}
}

func TestStaleFrenchPoolIsOneIssue(t *testing.T) {
	en, fr := `"dd18":"100"`, `"dd18":"90"`
	for i := 1; i <= 17; i++ {
		en += fmt.Sprintf(`,"dd%d":"%d"`, i, i*10)
		fr += fmt.Sprintf(`,"dd%d":"%d"`, i, i*10+1)
	}
	r := mustParse(t, `{"rounds":[{"drawNumber":"9","drawDate":"2026-01-01","drawName":"Canadian Experience Class",`+en+`}]}`,
		`{"rounds":[{"drawNumber":"9","drawName":"neuf",`+fr+`}]}`)
	if len(r.Issues) != 1 || r.Issues[0].Kind != FrenchMismatch || !strings.HasPrefix(r.Issues[0].Field, "dd1,dd2,") {
		t.Errorf("issues = %v, want one FrenchMismatch for all bands", r.Issues)
	}
}

func TestFrenchDateWithStraySpaceMatches(t *testing.T) {
	r := mustParse(t, `{"rounds":[{"drawNumber":"9","drawDate":"2026-01-01","drawName":"Canadian Experience Class"}]}`,
		`{"rounds":[{"drawNumber":"9","drawDate":"2026-01-01 ","drawName":"neuf"}]}`)
	if r.Draws[0].Name.FR == "" {
		t.Error("French text dropped over a stray space")
	}
}

func TestFieldMissingInEnglishIsNotCompared(t *testing.T) {
	r := mustParse(t, `{"rounds":[{"drawNumber":"9","drawDate":"2026-01-01","drawName":"Canadian Experience Class"}]}`,
		`{"rounds":[{"drawNumber":"9","drawName":"neuf","drawCRS":"500"}]}`)
	if len(r.Issues) != 0 {
		t.Error("mismatch against a field English doesn't have")
	}
}

func TestFrenchOnlyOrderIsStable(t *testing.T) {
	seen := map[string]bool{}
	for range 200 {
		r := mustParse(t, `{"rounds":[{"drawNumber":"9","drawDate":"2026-01-01","drawName":"Canadian Experience Class"}]}`,
			`{"rounds":[{"drawNumber":"91","drawName":"a"},{"drawNumber":"091","drawName":"b"},{"drawNumber":"0091","drawName":"c"}]}`)
		seen[fmt.Sprint(r.Issues)] = true
	}
	if len(seen) > 1 {
		t.Errorf("order varies: %v", seen)
	}
}
