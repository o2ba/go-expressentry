package expressentry

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/o2ba/go-expressentry/internal/clean"
)

// Two real rounds from IRCC's files (2026-11-15). The French file was a
// round behind: it has 452 but not 453.
const enFile = `{"classes":"wb-tables","rounds":[
{"drawNumber":"453","drawDate":"2026-10-28","drawDateFull":"October 28, 2026","drawName":"Transport Occupations, 2026-Version 2","drawSize":"150","drawCRS":"477","drawText2":"Federal Skilled Worker Program, Canadian Experience Class and Federal Skilled Trades Program","drawDateTime":"October 28, 2026 at 10:34:14 UTC","drawCutOff":"June 4, 2026 at 21:48:22 UTC","drawDistributionAsOn":"October 25, 2026","dd1":"787","dd2":"20,535","dd3":"71,112","dd4":"12,162","dd5":"12,342","dd6":"16,330","dd7":"15,803","dd8":"14,475","dd9":"63,505","dd10":"13,159","dd11":"13,155","dd12":"12,768","dd13":"12,543","dd14":"11,880","dd15":"47,911","dd16":"17,247","dd17":"7,627","dd18":"228,724"},
{"drawNumber":"452","drawDate":"2026-10-27","drawDateFull":"October 27, 2026","drawName":"French-Language proficiency 2026-Version 2","drawSize":"5,000","drawCRS":"407","drawText2":"Federal Skilled Worker Program, Canadian Experience Class and Federal Skilled Trades Program","drawDateTime":"October 27, 2026 at 10:17:26 UTC","drawCutOff":"July 13, 2026 at 19:22:00 UTC","drawDistributionAsOn":"October 25, 2026","dd1":"787","dd2":"20,535","dd3":"71,112","dd4":"12,162","dd5":"12,342","dd6":"16,330","dd7":"15,803","dd8":"14,475","dd9":"63,505","dd10":"13,159","dd11":"13,155","dd12":"12,768","dd13":"12,543","dd14":"11,880","dd15":"47,911","dd16":"17,247","dd17":"7,627","dd18":"228,724"}
]}`

const frFile = `{"classes":"wb-tables","rounds":[
{"drawNumber":"452","drawDate":"2026-10-27","drawDateFull":"27 octobre 2026","drawName":"Compétence linguistique en français 2026-Version 2","drawSize":"5 000","drawCRS":"407","drawText2":"Programme des travailleurs qualifiés (fédéral)","drawDateTime":"27 octobre 2026 à 10:17:26 UTC","drawCutOff":"13 juillet 2026 à 19:22:00 UTC","drawDistributionAsOn":"25 octobre 2026","dd1":"787","dd18":"228 724"}
]}`

func mustParse(t *testing.T, en, fr string) *Rounds {
	t.Helper()
	var frBytes []byte
	if fr != "" {
		frBytes = []byte(fr)
	}
	r, err := Parse([]byte(en), frBytes)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func kinds(issues []Issue) []IssueKind {
	var out []IssueKind
	for _, i := range issues {
		out = append(out, i.Kind)
	}
	return out
}

func TestParse(t *testing.T) {
	r := mustParse(t, enFile, frFile)
	if len(r.Issues) != 0 {
		t.Errorf("issues: %v", r.Issues)
	}
	if len(r.Draws) != 2 || len(r.Pools) != 1 {
		t.Fatalf("rounds = %+v", r)
	}

	d := r.Draws[0]
	if d.Number != "453" || d.Category != Transport || d.Invitations != 150 || d.MinimumCRS != 477 || d.Date != (Date{2026, time.October, 28}) {
		t.Errorf("453 = %+v", d)
	}
	if want := time.Date(2026, 10, 28, 10, 34, 14, 0, time.UTC); !d.InvitedAt.Equal(want) {
		t.Errorf("InvitedAt = %v", d.InvitedAt)
	}
	if want := time.Date(2026, 6, 4, 21, 48, 22, 0, time.UTC); !d.TieBreak.Equal(want) || d.TieBreak.Location() != time.UTC {
		t.Errorf("TieBreak = %v", d.TieBreak)
	}
	if d.Name.FR != "" {
		t.Errorf("453 is missing from the French file, FR = %q", d.Name.FR)
	}
	if d.PoolAsOf != (Date{2026, time.October, 25}) {
		t.Errorf("PoolAsOf = %v", d.PoolAsOf)
	}

	f := r.Draws[1]
	if f.Category != French || f.Name.FR != "Compétence linguistique en français 2026-Version 2" || f.Invitations != 5000 {
		t.Errorf("452 = %+v", f)
	}
}

func TestParseIsStable(t *testing.T) {
	a := mustParse(t, enFile, frFile)
	for range 20 {
		if b := mustParse(t, enFile, frFile); !reflect.DeepEqual(a, b) {
			t.Fatal("same input gave a different result")
		}
	}
}

func TestParseWithoutFrench(t *testing.T) {
	r := mustParse(t, enFile, "")
	if r.Draws[1].Name != (Name{EN: "French-Language proficiency 2026-Version 2"}) || len(r.Issues) != 0 {
		t.Errorf("name = %+v, issues = %v", r.Draws[1].Name, r.Issues)
	}
}

func TestParseWithBrokenFrench(t *testing.T) {
	r, err := Parse([]byte(enFile), []byte("<html>blocked</html>"))
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Draws) != 2 || !slices.Equal(kinds(r.Issues), []IssueKind{FrenchUnavailable}) {
		t.Errorf("draws = %d, issues = %v", len(r.Draws), r.Issues)
	}
}

func TestParseRejectsEnglish(t *testing.T) {
	for name, body := range map[string]string{
		"html":      `<html>blocked</html>`,
		"no rounds": `{"rounds":[]}`,
		"all bad":   `{"rounds":[{"drawSize":"about 5k"}]}`,
		"no number": `{"rounds":[{"drawSize":"5"}]}`,
	} {
		if _, err := Parse([]byte(body), nil); !errors.Is(err, ErrUnexpectedFormat) {
			t.Errorf("%s: err = %v, want ErrUnexpectedFormat", name, err)
		}
	}
}

func TestBadFieldsAreReportedNotFatal(t *testing.T) {
	r := mustParse(t, `{"rounds":[{"drawNumber":"9","drawDate":"2026-01-01","drawName":"Canadian Experience Class","drawDateTime":"sometime","drawDistributionAsOn":"soon","dd18":"5"}]}`, "")
	d := r.Draws[0]
	if len(r.Draws) != 1 || !d.InvitedAt.IsZero() || !d.PoolAsOf.IsZero() {
		t.Errorf("draw = %+v", d)
	}
	var fields []string
	for _, i := range r.Issues {
		if i.Kind != Invalid || i.Round != "9" {
			t.Errorf("issue = %v", i)
		}
		fields = append(fields, i.Field)
	}
	if !slices.Equal(fields, []string{"drawDateTime", "drawDistributionAsOn"}) {
		t.Errorf("fields = %q", fields)
	}
}

func TestDateWithoutCountsIsNoPool(t *testing.T) {
	r := mustParse(t, `{"rounds":[{"drawNumber":"1","drawDate":"2015-01-31","drawName":"No Program Specified","drawDistributionAsOn":"January 31, 2015"}]}`, "")
	if !r.Draws[0].PoolAsOf.IsZero() || len(r.Pools) != 0 || len(r.Issues) != 0 {
		t.Errorf("PoolAsOf = %v, pools = %v, issues = %v", r.Draws[0].PoolAsOf, r.Pools, r.Issues)
	}
}

func TestLetteredRoundNumbers(t *testing.T) {
	r := mustParse(t, `{"rounds":[
	  {"drawNumber":"91a","drawDate":"2018-05-30","drawName":"Provincial Nominee Program","drawSize":"200","drawCRS":"902"},
	  {"drawNumber":"92","drawDate":"2018-06-13","drawName":"No Program Specified","drawSize":"3,750","drawCRS":"451"},
	  {"drawNumber":"90","drawDate":"2018-05-23","drawName":"No Program Specified","drawSize":"3,500","drawCRS":"440"},
	  {"drawNumber":"91b","drawDate":"2018-05-30","drawName":"Federal Skilled Trades","drawSize":"500","drawCRS":"288"},
	  {"drawNumber":"89","drawSize":"lots"},
	  {"drawNumber":"90","drawDate":"2018-05-23"}
	]}`, "")
	var got []RoundNumber
	for _, d := range r.Draws {
		got = append(got, d.Number)
	}
	if want := []RoundNumber{"92", "91b", "91a", "90"}; !slices.Equal(got, want) {
		t.Errorf("order = %q, want %q", got, want)
	}
	if !slices.Equal(kinds(r.Issues), []IssueKind{Unreadable, Unreadable}) || r.Issues[0].Round != "89" || r.Issues[1].Round != "90" {
		t.Errorf("issues = %v, want unreadable 89 and repeated 90", r.Issues)
	}
}

func TestRepairedFieldsAreKeptAndReported(t *testing.T) {
	r := mustParse(t, `{"rounds":[{"drawNumber":"208","drawDate":"2021-10-13","drawName":"Provincial Nominee Program","drawDateTime":"October 13, 2021 at 14:10:40 UTC","drawCutOff":"October 7 at 16:38:21 UTC"}]}`, "")
	if want := time.Date(2021, 10, 7, 16, 38, 21, 0, time.UTC); !r.Draws[0].TieBreak.Equal(want) {
		t.Errorf("TieBreak = %v, want %v", r.Draws[0].TieBreak, want)
	}
	if len(r.Issues) != 1 || r.Issues[0].Kind != Repaired || r.Issues[0].Round != "208" || r.Issues[0].Field != "drawCutOff" {
		t.Errorf("issues = %v, want round 208 drawCutOff repaired", r.Issues)
	}
}

// poolRound returns one IRCC round whose pool is as of asOf, with every
// band set to n (the two split bands to 5*n), so the total is 15*n.
func poolRound(number, asOf string, n int) string {
	var dd []string
	for i := 1; i <= 17; i++ {
		count := n
		if i == 3 || i == 9 {
			count = 5 * n
		}
		dd = append(dd, fmt.Sprintf(`"dd%d":"%d"`, i, count))
	}
	return fmt.Sprintf(`{"drawNumber":"%s","drawDate":"2026-01-01","drawName":"Canadian Experience Class","drawDistributionAsOn":"%s",%s,"dd18":"%d"}`,
		number, asOf, strings.Join(dd, ","), 15*n)
}

func file(rounds ...string) string {
	return `{"rounds":[` + strings.Join(rounds, ",") + `]}`
}

func TestPool(t *testing.T) {
	p := mustParse(t, enFile, "").Pools[0]
	if p.AsOf != (Date{2026, time.October, 25}) || p.Total != 228724 || len(p.Bands) != 7 {
		t.Fatalf("pool = %+v", p)
	}
	sum := 0
	for _, b := range p.Bands {
		sum += b.Count
		sub := 0
		for _, s := range b.Bands {
			sub += s.Count
		}
		if len(b.Bands) > 0 && sub != b.Count {
			t.Errorf("band %d-%d: parts add to %d, want %d", b.Min, b.Max, sub, b.Count)
		}
	}
	if sum != p.Total {
		t.Errorf("bands add to %d, want %d", sum, p.Total)
	}
	if top := p.Bands[2].Bands[0]; top.Min != 491 || top.Max != 500 || top.Count != 12162 {
		t.Errorf("491-500 = %+v", top)
	}
}

func TestPoolsDedupeByDateNewestFirst(t *testing.T) {
	r := mustParse(t, file(
		poolRound("12", "March 3, 2026", 30),
		poolRound("10", "February 1, 2026", 10),
		poolRound("11", "February 1, 2026", 10),
		poolRound("13", "March 3, 2026", 30),
		poolRound("14", "April 7, 2026", 40),
		`{"drawNumber":"9","drawDate":"2015-01-31","drawName":"No Program Specified","drawDistributionAsOn":"January 31, 2015"}`,
	), "")
	if len(r.Issues) != 0 {
		t.Errorf("issues = %v, want none for identical snapshots", r.Issues)
	}
	var got []string
	for _, p := range r.Pools {
		got = append(got, p.AsOf.String())
	}
	if want := []string{"2026-04-07", "2026-03-03", "2026-02-01"}; !slices.Equal(got, want) {
		t.Errorf("pools = %q, want %q", got, want)
	}
}

func TestPoolConflictKeepsEarliestRound(t *testing.T) {
	r := mustParse(t, file(
		poolRound("22", "March 3, 2026", 31),
		poolRound("20", "March 3, 2026", 30),
		poolRound("21", "March 3, 2026", 30),
		poolRound("23", "March 3, 2026", 32),
	), "")
	if len(r.Pools) != 1 || r.Pools[0].Total != 450 {
		t.Fatalf("pools = %+v, want round 20's snapshot only", r.Pools)
	}
	if !slices.Equal(kinds(r.Issues), []IssueKind{PoolConflict, PoolConflict}) || r.Issues[0].Round != "22" || r.Issues[1].Round != "23" {
		t.Errorf("issues = %v, want conflicts for rounds 22 and 23", r.Issues)
	}
	for _, d := range r.Draws {
		if d.PoolAsOf != (Date{2026, time.March, 3}) {
			t.Errorf("round %s PoolAsOf = %v", d.Number, d.PoolAsOf)
		}
	}
}

func TestBadCountKeepsRoundWithoutPool(t *testing.T) {
	r := mustParse(t, `{"rounds":[{"drawNumber":"9","drawDate":"2026-01-01","drawName":"Canadian Experience Class","drawDistributionAsOn":"January 1, 2026","dd3":"lots","dd18":"5"}]}`, "")
	if len(r.Draws) != 1 || !r.Draws[0].PoolAsOf.IsZero() || len(r.Pools) != 0 {
		t.Fatalf("draws = %+v, pools = %v", r.Draws, r.Pools)
	}
	if len(r.Issues) != 1 || r.Issues[0].Kind != Invalid || r.Issues[0].Field != "dd3" {
		t.Errorf("issues = %v", r.Issues)
	}
}

func TestRepeatedFrenchRoundKeepsFirst(t *testing.T) {
	r := mustParse(t, `{"rounds":[{"drawNumber":"9","drawDate":"2026-01-01","drawName":"Canadian Experience Class"}]}`,
		`{"rounds":[{"drawNumber":"9","drawName":"premier"},{"drawNumber":"9","drawName":"second"}]}`)
	if got := r.Draws[0].Name.FR; got != "premier" {
		t.Errorf("Name.FR = %q, want premier", got)
	}
}

func TestCleanCategoriesAreValid(t *testing.T) {
	for _, id := range clean.Categories() {
		if !Category(id).Valid() {
			t.Errorf("clean returns %q, which isn't a valid Category", id)
		}
	}
}
