# go-expressentry

Typed Go access to the Express Entry rounds of invitations that
Immigration, Refugees and Citizenship Canada (IRCC) publishes.

IRCC lists every round in two undocumented JSON files, one in English and
one in French. They're edited by hand: numbers arrive as `"5,000"` or
`"245 123"`, dates are written several ways, and some old rounds are
malformed. This package reads both files and returns rounds with real dates,
numbers and categories, plus the CRS pool snapshots published with them.

Not affiliated with or endorsed by IRCC. The data is IRCC's.

```sh
go get github.com/o2ba/go-expressentry
```

Requires Go 1.24 or later.

> Work in progress: the API is defined, the parser isn't in yet. `Rounds`
> and `Parse` return `ErrNotImplemented`.

## Usage

```go
c := expressentry.NewClient()
r, err := c.Rounds(ctx)
if err != nil {
	return err
}
for _, d := range r.Draws {
	fmt.Println(d.Number, d.Date, d.Category, d.MinimumCRS, d.Invitations)
}
```

`Parse(en, fr)` does the same with bytes you already have. Options:

| Option | Default |
| --- | --- |
| `WithHTTPClient(hc)` | `http.Client` with a 30-second timeout |
| `WithURLs(en, fr)` | IRCC's published files; an empty `fr` skips French |
| `WithUserAgent(ua)` | `go-expressentry` |

To poll without downloading unchanged files, keep `r.Version` and pass it
back:

```go
r, changed, err := c.RoundsSince(ctx, last.Version)
if err != nil || !changed {
	return err
}
```

A problem with one round never fails the call. The round is left out or
repaired, and the problem is listed in `Rounds.Issues`.

The package keeps no state and polls nothing. Fetching on a schedule,
keeping the last `Version`, storing results and spotting new rounds are up to you.

## License

MIT
