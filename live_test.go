package expressentry

import (
	"context"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/o2ba/go-expressentry/internal/corrections"
)

// TestLive reads IRCC's published files and fails on any issue other than a
// known repair, so a format change shows up before callers hit it. It also
// fails when a correction no longer matches, so ones IRCC fixed get removed.
// It only runs with EXPRESSENTRY_LIVE=1.
func TestLive(t *testing.T) {
	if os.Getenv("EXPRESSENTRY_LIVE") != "1" {
		t.Skip("set EXPRESSENTRY_LIVE=1 to read IRCC's live files")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	r, err := NewClient(WithUserAgent("go-expressentry-ci")).Rounds(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Draws) == 0 || len(r.Pools) == 0 {
		t.Fatalf("%d draws, %d pools", len(r.Draws), len(r.Pools))
	}
	for _, i := range r.Issues {
		if i.Kind != Repaired {
			t.Errorf("%v", i)
		}
	}
	for _, d := range r.Draws {
		if d.Date.IsZero() || d.InvitedAt.IsZero() {
			t.Errorf("round %s: date %v, invited at %v", d.Number, d.Date, d.InvitedAt)
		}
	}
	for _, c := range corrections.All {
		if !slices.ContainsFunc(r.Issues, func(i Issue) bool {
			return i.Kind == Repaired && string(i.Round) == c.Round && i.Field == c.Field && i.Value == c.Wrong
		}) {
			t.Errorf("correction for round %s %s (%s) no longer matches IRCC's file; if IRCC fixed it, remove it from internal/corrections/list.go", c.Round, c.Field, c.Lang)
		}
	}
	newest := r.Draws[0]
	t.Logf("%d rounds, newest %s on %s; %d repaired fields", len(r.Draws), newest.Number, newest.Date, len(r.Issues))
}
