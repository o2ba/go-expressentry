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

The package keeps no state and polls nothing. Fetching on a schedule,
keeping the last `Version`, storing results and spotting new rounds are up
to you.

## What you get

| Field | From IRCC | Notes |
| --- | --- | --- |
| `Number` | `drawNumber` | Text, because of round `91b` |
| `Date` | `drawDate` | A calendar day with no time zone |
| `InvitedAt` | `drawDateTime` | UTC |
| `TieBreak` | `drawCutOff` | UTC |
| `Category` | `drawName` | One of the constants below |
| `Name`, `Programs` | `drawName`, `drawText2` | English and French |
| `Invitations`, `MinimumCRS` | `drawSize`, `drawCRS` | |
| `PoolAsOf` | `drawDistributionAsOn` | Date of the matching `Pool` |

`Draws` and `Pools` are newest first. Each `Pool` is the CRS score
distribution on one date, with the 451-500 and 401-450 bands split into
10-point parts. Several rounds share one snapshot, so each date appears once.

Categories: `CEC`, `French`, `PNP`, `Healthcare`, `Trades`, `Transport`,
`Managers`, `Physicians`, `Military`, `STEM`, `Education`, `General` and
`Other`. `Category.Name()` gives each one in English and French. A round
whose name matches none is `Other`, and IRCC's wording stays in `Name`.

## Issues

A problem with one round never fails the call. It is listed in
`Rounds.Issues`:

| Kind | Meaning |
| --- | --- |
| `Unreadable` | The round couldn't be read, or its number was missing or repeated. It is left out. |
| `Invalid` | A field couldn't be read and is left empty. The round is kept. |
| `Repaired` | A field had a known typo, such as `"May 31, 2024 at12:48:30 UTC"` or a time with no year, and was read anyway. |
| `PoolConflict` | Two rounds gave different counts for the same pool date. The earlier round's counts are kept. |
| `FrenchUnavailable` | The French file couldn't be downloaded or read. Names have English text only. |

`err` is only returned when the English file can't be downloaded
(`*HTTPError`), isn't a rounds file (`ErrUnexpectedFormat`), or the
context ends.

## Development

```sh
go test -race ./...
go test -run='^$' -fuzz=FuzzParse -fuzztime=30s .
```

## License

MIT
