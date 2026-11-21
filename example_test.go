package expressentry_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/o2ba/go-expressentry"
)

func ExampleParse() {
	en := []byte(`{"rounds":[{"drawNumber":"453","drawDate":"2026-10-28","drawName":"Transport Occupations, 2026-Version 2",
		"drawSize":"150","drawCRS":"477","drawCutOff":"June 4, 2026 at 21:48:22 UTC"}]}`)
	fr := []byte(`{"rounds":[{"drawNumber":"453","drawName":"Métiers des transports 2026-Version 2"}]}`)

	r, err := expressentry.Parse(en, fr)
	if err != nil {
		log.Fatal(err)
	}
	d := r.Draws[0]
	fmt.Printf("Round %s on %s: %d invited, minimum CRS %d\n", d.Number, d.Date, d.Invitations, d.MinimumCRS)
	fmt.Println(d.Category.Name().EN, "/", d.Name.FR)
	fmt.Println("Tie-break:", d.TieBreak.Format(time.RFC3339))
	// Output:
	// Round 453 on 2026-10-28: 150 invited, minimum CRS 477
	// Transport / Métiers des transports 2026-Version 2
	// Tie-break: 2026-06-04T21:48:22Z
}

func ExampleClient_RoundsSince() {
	c := expressentry.NewClient(expressentry.WithUserAgent("example/1.0"))
	ctx := context.Background()

	r, err := c.Rounds(ctx)
	if err != nil {
		log.Fatal(err)
	}
	last := r.Version

	for range time.Tick(15 * time.Minute) {
		r, changed, err := c.RoundsSince(ctx, last)
		if err != nil {
			log.Print(err)
			continue
		}
		if !changed {
			continue
		}
		last = r.Version
		for _, issue := range r.Issues {
			log.Print(issue)
		}
		fmt.Println("Newest round:", r.Draws[0].Number)
	}
}
