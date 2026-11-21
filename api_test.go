package expressentry

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestNewClientDefaults(t *testing.T) {
	c := NewClient()
	if c.englishURL != DefaultEnglishURL || c.frenchURL != DefaultFrenchURL {
		t.Errorf("urls = %q %q", c.englishURL, c.frenchURL)
	}
	if c.httpClient == nil || c.httpClient.Timeout != defaultTimeout {
		t.Errorf("http client = %+v", c.httpClient)
	}
}

func TestOptions(t *testing.T) {
	hc := &http.Client{}
	c := NewClient(WithHTTPClient(hc), WithURLs("http://en", ""), WithUserAgent("test"))
	if c.httpClient != hc || c.englishURL != "http://en" || c.frenchURL != "" || c.userAgent != "test" {
		t.Errorf("client = %+v", c)
	}
	if NewClient(WithHTTPClient(nil)).httpClient == nil {
		t.Error("nil http client replaced the default")
	}
}

func TestCategoriesAreNamed(t *testing.T) {
	cats := Categories()
	if len(cats) != len(categoryNames) {
		t.Fatalf("%d categories, %d names", len(cats), len(categoryNames))
	}
	for _, c := range cats {
		if !c.Valid() || c.Name().EN == "" || c.Name().FR == "" {
			t.Errorf("%q: valid %v, name %+v", c, c.Valid(), c.Name())
		}
	}
	if Category("nope").Valid() {
		t.Error("unknown category is valid")
	}
}

func TestDate(t *testing.T) {
	d, err := ParseDate("2026-11-12")
	if err != nil {
		t.Fatal(err)
	}
	if d != (Date{2026, time.November, 12}) || d.String() != "2026-11-12" {
		t.Errorf("date = %+v %q", d, d)
	}
	if !d.Before(Date{2026, time.November, 13}) || d.Before(d) {
		t.Error("Before")
	}
	toronto, err := time.LoadLocation("America/Toronto")
	if err == nil && DateOf(d.In(toronto)) != d {
		t.Errorf("In(Toronto) = %v", d.In(toronto))
	}
	if _, err := ParseDate("November 12, 2026"); err == nil {
		t.Error("parsed a written-out date")
	}
}

func TestDateJSON(t *testing.T) {
	var v struct{ D, Z Date }
	if err := json.Unmarshal([]byte(`{"D":"2026-11-12","Z":""}`), &v); err != nil {
		t.Fatal(err)
	}
	if v.D != (Date{2026, time.November, 12}) || !v.Z.IsZero() {
		t.Errorf("decoded %+v", v)
	}
	b, err := json.Marshal(v)
	if err != nil || string(b) != `{"D":"2026-11-12","Z":""}` {
		t.Errorf("encoded %s, %v", b, err)
	}
}

func TestIssueString(t *testing.T) {
	i := Issue{Kind: Repaired, Round: "453", Field: "drawCutOff", Value: "x"}
	if got := i.String(); got != `round 453: repaired drawCutOff "x"` {
		t.Errorf("String() = %s", got)
	}
}
