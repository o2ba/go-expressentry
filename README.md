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

## Where the data comes from

The client reads IRCC's two rounds files, which back the table on the
[Express Entry rounds page](https://www.canada.ca/en/immigration-refugees-citizenship/corporate/mandate/policies-operational-instructions-agreements/ministerial-instructions/express-entry-rounds.html):

| File | URL |
| --- | --- |
| English (`DefaultEnglishURL`) | `https://www.canada.ca/content/dam/ircc/documents/json/ee_rounds_123_en.json` |
| French (`DefaultFrenchURL`) | `https://www.canada.ca/content/dam/ircc/documents/json/ee_rounds_123_fr.json` |

`WithURLs` points the client somewhere else, such as a mirror or a test
server.

**The English file is the source of truth.** It decides which rounds exist
and gives every number, date and category. The French file only adds the
French text of `Name` and `Programs`, matched by round number. A round the
French file doesn't have gets empty French text.

## How IRCC shapes the files

Each file is a JSON object with a `classes` key (a CSS class for the web
table) and a `rounds` array, newest round first. The English file is about
740 KB, and the French one a little larger. Every value is a string, even
numbers. This is round 453 in the English file, cut down:

```json
{
  "classes": "wb-tables",
  "rounds": [
    {
      "drawNumber": "453",
      "drawNumberURL": "<a href='...invitations.html?q=453'>453</a>",
      "drawDate": "2026-10-28",
      "drawDateFull": "October 28, 2026",
      "drawDateTime": "October 28, 2026 at 10:34:14 UTC",
      "drawName": "Transport Occupations, 2026-Version 2",
      "drawSize": "150",
      "drawCRS": "477",
      "mitext": "<a href='...'>Invitations to apply for permanent residence under the Express Entry system #453</a>",
      "DrawText1": "<a href='...'>See full text of Ministerial Instruction</a>",
      "drawText2": "Federal Skilled Worker Program, Canadian Experience Class and Federal Skilled Trades Program",
      "drawCutOff": "June 4, 2026 at 21:48:22 UTC",
      "drawDistributionAsOn": "October 25, 2026",
      "dd1": "787",
      "dd2": "20,535",
      "dd3": "71,112",
      "dd18": "228,724"
    }
  ]
}
```

| Key | What it holds |
| --- | --- |
| `drawNumber` | The round number. Usually digits, but once `"91b"`. |
| `drawDate` | The round's date as `YYYY-MM-DD`, the same in both files. |
| `drawDateFull` | The same date written out, in the file's language. |
| `drawDateTime` | When invitations went out, typed by hand: `"October 28, 2026 at 10:34:14 UTC"` in English, `"28 octobre 2026 à 10:34:14 UTC"` in French. |
| `drawCutOff` | The tie-break time, typed the same way. Empty in the earliest rounds. |
| `drawName` | IRCC's name for the round, such as `"Canadian Experience Class"`. It differs between the files. |
| `drawText2` | The programs the round covered, as text. |
| `drawSize`, `drawCRS` | Invitations and the lowest score invited, with thousands separators: `"5,000"` in English, `"5 000"` in French. |
| `drawDistributionAsOn` | The date of the pool snapshot published with the round, written out. |
| `dd1` to `dd18` | Profiles in the pool by score: `dd1` 601-1200, `dd2` 501-600, `dd3` 451-500 split into `dd4` to `dd8` (10 points each, highest first), `dd9` 401-450 split into `dd10` to `dd14`, `dd15` 351-400, `dd16` 301-350, `dd17` 0-300 and `dd18` the total. Early rounds have `"0"` everywhere. |
| `drawNumberURL` | An HTML link to the round's ministerial instructions: `invitations.html?q=453` for newer rounds, a page of its own such as `invitations-1.html` for old ones. Newer rounds use the site's internal `/content/canadasite` path, which canada.ca redirects. Read into `Instructions` as an absolute public URL. |
| `mitext`, `DrawText1` | The same link with other text. Not read. IRCC's own links have mistakes: in round 65 these two point to round 64's page, and the French link for round 149 is broken (`invitations-14.html9`). Links are passed on as published. |

The times are typed by hand and vary: ordinal suffixes
(`"1st"`), doubled spaces, a missing `UTC`, and in old rounds typos such as
`"at12:48:30"`, `"20:19:59 PM"` or a time with no year. Rounds 208 to 333
hold all of the typos this package repairs.

## What you get

`Rounds` and `RoundsSince` return a `*Rounds`:

```go
type Rounds struct {
	Draws   []Draw  // every readable round, newest first
	Pools   []Pool  // each CRS pool snapshot once, newest first
	Issues  []Issue // problems with single rounds, in the order found
	Version Version // Last-Modified of each file, for RoundsSince
}

type Draw struct {
	Number       RoundNumber // "453"; text because of round "91b"
	Date         Date        // 2026-10-28, a calendar day with no time zone
	InvitedAt    time.Time   // 2026-10-28T10:34:14Z, UTC
	Category     Category    // Transport
	Name         Name        // {EN: "Transport Occupations, 2026-Version 2", FR: "Métiers des transports 2026-Version 2"}
	Invitations  int         // 150
	MinimumCRS   int         // 477
	TieBreak     time.Time   // 2026-06-04T21:48:22Z, UTC
	Programs     Name        // programs the round covered, in English and French
	Instructions Links       // {EN: "https://www.canada.ca/en/.../invitations.html?q=453", FR: "https://www.canada.ca/fr/..."}
	PoolAsOf     Date        // 2026-10-25, the matching Pool, or zero
}

type Pool struct {
	AsOf  Date   // 2026-10-25
	Total int    // 228724
	Bands []Band // 601-1200, 501-600, 451-500, 401-450, 351-400, 301-350, 0-300
}

type Links struct {
	EN, FR string // absolute canada.ca URLs
}

type Band struct {
	Min, Max, Count int
	Bands           []Band // 10-point parts of 451-500 and 401-450
}
```

Fields map to IRCC's keys like this:

| Field | From IRCC |
| --- | --- |
| `Number` | `drawNumber` |
| `Date` | `drawDate` |
| `InvitedAt` | `drawDateTime` |
| `TieBreak` | `drawCutOff` |
| `Category`, `Name` | `drawName` |
| `Programs` | `drawText2` |
| `Instructions` | the link in `drawNumberURL` |
| `Invitations`, `MinimumCRS` | `drawSize`, `drawCRS` |
| `PoolAsOf`, `Pool` | `drawDistributionAsOn`, `dd1` to `dd18` |

A time field is the zero `time.Time` when IRCC gives none. Several rounds
share one pool snapshot, so each date appears once in `Pools`.

Categories: `CEC`, `French`, `PNP`, `Healthcare`, `Trades`, `Transport`,
`Managers`, `Physicians`, `Military`, `STEM`, `Education`, `General` and
`Other`. `Category.Name()` gives each one in English and French. A round
whose name matches none is `Other`, and IRCC's wording stays in `Name`.

## Errors

The call returns an error only when there is nothing useful to return:

| Error | When |
| --- | --- |
| `*HTTPError` | The English file answered with a status other than 200 OK. `StatusCode` and `URL` say which. |
| `ErrUnexpectedFormat` (use `errors.Is`) | The English file isn't JSON, has no rounds, has no round that can be read, or is larger than 16 MB. |
| `context.Canceled`, `context.DeadlineExceeded` (use `errors.Is`) | The context was cancelled or its deadline passed. |
| Other errors from `net/http` | The English file couldn't be downloaded at all: a DNS or TLS failure, a dropped connection, or the client's 30-second timeout. |

`Parse` only returns `ErrUnexpectedFormat`. A French file that fails in any
of these ways is not an error: see `FrenchUnavailable` below.

## Issues

A problem with one round never fails the call. It is listed in
`Rounds.Issues`, with the `Round`, IRCC's `Field` and `Value`, and an `Err`
that explains it:

| Kind | Meaning |
| --- | --- |
| `Unreadable` | The round couldn't be read, or its number was missing or repeated. It is left out. |
| `Invalid` | A field couldn't be read and is left empty. The round is kept. |
| `Repaired` | A field had a known typo, such as `"May 31, 2024 at12:48:30 UTC"` or a time with no year, and was read anyway. |
| `PoolConflict` | Two rounds gave different counts for the same pool date. The earlier round's counts are kept. |
| `FrenchUnavailable` | The French file couldn't be downloaded or read, so names have English text only. With `Round` set, only that French round couldn't be read. |
| `FrenchOnly` | A round is in the French file but not the English one. It is left out. |
| `FrenchMismatch` | The two files disagree on a round's `drawDate`, `drawSize`, `drawCRS` or pool counts. The English value is kept, and all differing pool counts of a round come in one issue. |

## When English and French disagree

The English values always win, and every disagreement is reported:

- A French round that English doesn't have gives a `FrenchOnly` issue and
  no draw.
- A shared round with a different number of invitations, minimum score or
  pool count gives a `FrenchMismatch` issue for that field. The French text
  is still used.
- A shared round with a different `drawDate` is taken to be another round
  that IRCC numbered wrongly. It gives a `FrenchMismatch` issue, and the
  round keeps English text only.
- An English round that French doesn't have yet is not an issue. Its
  French text is empty.

## Development

```sh
go test -race ./...
go test -run='^$' -fuzz=FuzzParse -fuzztime=30s .
EXPRESSENTRY_LIVE=1 go test -run '^TestLive$' -v .
```

`TestLive` reads IRCC's live files and fails on any issue other than a
known repair. The Live workflow runs it every Monday and opens an issue when it
fails (or comments on the open one), so a new format shows up before it
reaches callers.

## License

MIT
