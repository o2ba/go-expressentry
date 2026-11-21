package expressentry_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/o2ba/go-expressentry"
)

func ExampleClient_Rounds() {
	c := expressentry.NewClient(
		expressentry.WithHTTPClient(&http.Client{Timeout: 10 * time.Second}),
		expressentry.WithUserAgent("example/1.0"),
	)
	r, err := c.Rounds(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	for _, issue := range r.Issues {
		log.Print(issue)
	}
	if len(r.Draws) > 0 {
		d := r.Draws[0]
		fmt.Printf("Round %s on %s: %d invited, minimum CRS %d (%s)\n",
			d.Number, d.Date, d.Invitations, d.MinimumCRS, d.Category.Name().EN)
	}
}
