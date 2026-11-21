package expressentry

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// feeds serves an English and a French rounds file with Last-Modified
// times, honouring If-Modified-Since the way canada.ca does.
type feeds struct {
	mu       sync.Mutex
	files    map[string]served
	requests map[string]int
	agents   []string
}

type served struct {
	body     string
	modified time.Time
	status   int
}

func newFeeds(t *testing.T) (*feeds, *Client) {
	t.Helper()
	at := time.Date(2026, 10, 28, 13, 37, 19, 0, time.UTC)
	f := &feeds{
		files: map[string]served{
			"/en.json": {body: enFile, modified: at},
			"/fr.json": {body: frFile, modified: at.Add(-3 * time.Second)},
		},
		requests: map[string]int{},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		file, ok := f.files[r.URL.Path]
		f.requests[r.URL.Path]++
		f.agents = append(f.agents, r.UserAgent())
		f.mu.Unlock()
		switch {
		case !ok:
			http.NotFound(w, r)
		case file.status != 0:
			w.WriteHeader(file.status)
		default:
			http.ServeContent(w, r, r.URL.Path, file.modified, bytes.NewReader([]byte(file.body)))
		}
	}))
	t.Cleanup(srv.Close)
	return f, NewClient(WithURLs(srv.URL+"/en.json", srv.URL+"/fr.json"), WithUserAgent("test/1"))
}

func (f *feeds) set(path string, s served) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.files[path] = s
}

func (f *feeds) count(path string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.requests[path]
}

func TestRounds(t *testing.T) {
	f, c := newFeeds(t)
	r, err := c.Rounds(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Draws) != 2 || r.Draws[1].Name.FR == "" || len(r.Issues) != 0 {
		t.Fatalf("rounds = %+v", r)
	}
	if want := (Version{EN: f.files["/en.json"].modified, FR: f.files["/fr.json"].modified}); r.Version != want {
		t.Errorf("Version = %+v, want %+v", r.Version, want)
	}
	for _, ua := range f.agents {
		if ua != "test/1" {
			t.Errorf("User-Agent = %q", ua)
		}
	}
}

func TestRoundsSinceUnchanged(t *testing.T) {
	f, c := newFeeds(t)
	first, err := c.Rounds(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	r, changed, err := c.RoundsSince(context.Background(), first.Version)
	if err != nil || changed || r != nil {
		t.Fatalf("RoundsSince = %v, %v, %v; want nil, false, nil", r, changed, err)
	}
	if f.count("/en.json") != 2 || f.count("/fr.json") != 2 {
		t.Errorf("requests = %v", f.requests)
	}
}

func TestRoundsSinceOneFileChanged(t *testing.T) {
	f, c := newFeeds(t)
	first, err := c.Rounds(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	en := f.files["/en.json"]
	en.modified = en.modified.Add(time.Hour)
	f.set("/en.json", en)

	r, changed, err := c.RoundsSince(context.Background(), first.Version)
	if err != nil || !changed {
		t.Fatalf("RoundsSince = %v, %v", changed, err)
	}
	if r.Draws[1].Name.FR == "" {
		t.Error("French text lost when only English changed")
	}
	if r.Version.EN != en.modified || r.Version.FR != first.Version.FR {
		t.Errorf("Version = %+v", r.Version)
	}
	if f.count("/fr.json") != 3 {
		t.Errorf("French requests = %d, want a refetch after its 304", f.count("/fr.json"))
	}
}

func TestRoundsEnglishHTTPError(t *testing.T) {
	f, c := newFeeds(t)
	f.set("/en.json", served{status: http.StatusServiceUnavailable})
	_, err := c.Rounds(context.Background())
	var he *HTTPError
	if !errors.As(err, &he) || he.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("err = %v, want *HTTPError 503", err)
	}
}

func TestRoundsFrenchFailureKeepsEnglish(t *testing.T) {
	f, c := newFeeds(t)
	f.set("/fr.json", served{status: http.StatusNotFound})
	r, err := c.Rounds(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Draws) != 2 || len(r.Issues) != 1 || r.Issues[0].Kind != FrenchUnavailable || !r.Version.FR.IsZero() {
		t.Fatalf("rounds = %+v", r)
	}
	var he *HTTPError
	if !errors.As(r.Issues[0].Err, &he) || he.StatusCode != http.StatusNotFound {
		t.Errorf("issue err = %v", r.Issues[0].Err)
	}
}

func TestRoundsWithoutFrench(t *testing.T) {
	f, c := newFeeds(t)
	WithURLs(c.englishURL, "")(c)
	r, err := c.Rounds(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if f.count("/fr.json") != 0 || len(r.Issues) != 0 || r.Draws[1].Name.FR != "" {
		t.Errorf("requests = %v, issues = %v", f.requests, r.Issues)
	}
}

func TestRoundsUnexpectedBody(t *testing.T) {
	f, c := newFeeds(t)
	f.set("/en.json", served{body: "<html>maintenance</html>", modified: time.Now()})
	if _, err := c.Rounds(context.Background()); !errors.Is(err, ErrUnexpectedFormat) {
		t.Fatalf("err = %v, want ErrUnexpectedFormat", err)
	}
}

func TestRoundsCancelled(t *testing.T) {
	_, c := newFeeds(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := c.Rounds(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}
