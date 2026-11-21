package raw

import (
	"errors"
	"testing"
)

func TestParse(t *testing.T) {
	body := []byte(`{"classes":"x","rounds":[
	  {"drawNumber":"453","drawDate":"2026-11-12","drawName":"Canadian Experience Class",
	   "drawSize":"5,000","drawCRS":"515","drawCutOff":"November 3, 2026 at 10:02:11 UTC",
	   "drawDistributionAsOn":"November 10, 2026","dd1":"612","dd18":"245 123","unknownField":true},
	  {"drawNumber":452,"drawSize":1500,"drawCRS":null}
	]}`)
	f, err := Parse(body)
	if err != nil {
		t.Fatal(err)
	}
	r := f.Rounds[0]
	if r.DrawNumber != "453" || r.DrawSize != 5000 || r.DrawCRS != 515 {
		t.Errorf("numbers = %q %d %d", r.DrawNumber, r.DrawSize, r.DrawCRS)
	}
	if r.Distribution[0] != 612 || r.Distribution[17] != 245123 || r.Distribution[5] != 0 {
		t.Errorf("distribution = %v", r.Distribution)
	}
	if r.DrawName != "Canadian Experience Class" || r.DistributionAsOn != "November 10, 2026" {
		t.Errorf("strings = %q %q", r.DrawName, r.DistributionAsOn)
	}
	if f.Rounds[1].DrawNumber != "452" || f.Rounds[1].DrawSize != 1500 || f.Rounds[1].DrawCRS != 0 {
		t.Errorf("plain numbers = %+v", f.Rounds[1])
	}
}

func TestParseRejects(t *testing.T) {
	for name, body := range map[string]string{
		"html":              `<html>blocked</html>`,
		"no rounds":         `{"rounds":[]}`,
		"bad size":          `{"rounds":[{"drawSize":"about 5k"}]}`,
		"rounds not a list": `{"rounds":{}}`,
	} {
		if _, err := Parse([]byte(body)); err == nil {
			t.Errorf("%s: parsed, want an error", name)
		}
	}
}

func TestParseSkipsUnreadableRounds(t *testing.T) {
	body := []byte(`{"rounds":[
	  {"drawNumber":"92","drawSize":"3,500"},
	  {"drawNumber":"91b","drawSize":"3,750"},
	  {"drawNumber":"91a","drawSize":"lots"},
	  {"drawNumber":90,"drawSize":"3,500"}
	]}`)
	f, err := Parse(body)
	if err != nil {
		t.Fatal(err)
	}
	var got []DrawNumber
	for _, r := range f.Rounds {
		got = append(got, r.DrawNumber)
	}
	if len(got) != 3 || got[0] != "92" || got[1] != "91b" || got[2] != "90" {
		t.Errorf("rounds = %q, want 92 91b 90", got)
	}
	if len(f.Unreadable) != 1 || f.Unreadable[0].Index != 2 || f.Unreadable[0].Number != "91a" {
		t.Errorf("unreadable = %+v, want round 91a at index 2", f.Unreadable)
	}
}

func TestParseWithNoReadableRoundsWrapsTheFirst(t *testing.T) {
	_, err := Parse([]byte(`{"rounds":[{"drawNumber":"91a","drawSize":"lots"}]}`))
	var u *Unreadable
	if !errors.As(err, &u) || u.Number != "91a" || u.Index != 0 {
		t.Fatalf("err = %v, want *Unreadable for round 91a", err)
	}
	if errors.Unwrap(u) == nil {
		t.Error("Unreadable doesn't unwrap to its decode error")
	}
}
