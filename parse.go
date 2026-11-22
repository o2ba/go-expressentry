package expressentry

import (
	"cmp"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/o2ba/go-expressentry/internal/clean"
	"github.com/o2ba/go-expressentry/internal/corrections"
	"github.com/o2ba/go-expressentry/internal/ircctime"
	"github.com/o2ba/go-expressentry/internal/raw"
)

// Parse reads the English and French rounds files. The English file is the
// source of truth: it says which rounds exist and gives every value. The
// French one only adds French text, matched by round number. fr may be nil.
// If fr can't be read, the rounds still come back with English text only
// and a [FrenchUnavailable] issue. French rounds that English doesn't have
// are reported as [FrenchOnly], and values the two files disagree on as
// [FrenchMismatch].
//
// The error wraps [ErrUnexpectedFormat] when en isn't a rounds file or none
// of its rounds can be read.
func Parse(en, fr []byte) (*Rounds, error) {
	r, _, err := parse(en, fr)
	return r, err
}

// parse is Parse, and also reports whether French text from fr was used.
func parse(en, fr []byte) (r *Rounds, usedFrench bool, err error) {
	enFeed, err := raw.Parse(en)
	if err != nil {
		return nil, false, fmt.Errorf("%w: English file: %w", ErrUnexpectedFormat, err)
	}
	var p parser
	p.correct(corrections.English, enFeed)
	french := map[string]raw.Round{}
	if fr != nil {
		frFeed, err := raw.Parse(fr)
		if err != nil {
			p.issue(Issue{Kind: FrenchUnavailable, Err: err})
		} else {
			usedFrench = true
			p.correct(corrections.French, frFeed)
			for _, u := range frFeed.Unreadable {
				p.issue(Issue{Kind: FrenchUnavailable, Round: RoundNumber(u.Number), Err: u.Err})
			}
			for _, r := range frFeed.Rounds {
				if _, ok := french[string(r.DrawNumber)]; !ok {
					french[string(r.DrawNumber)] = r
				}
			}
		}
	}
	out := p.rounds(enFeed, french)
	if len(out.Draws) == 0 {
		return nil, false, fmt.Errorf("%w: English file: no round has a number", ErrUnexpectedFormat)
	}
	return out, usedFrench, nil
}

type parser struct {
	issues []Issue
}

// correct applies the known corrections to feed's rounds, reporting each
// as Repaired with IRCC's original value.
func (p *parser) correct(lang corrections.Lang, feed *raw.Feed) {
	for i := range feed.Rounds {
		r := &feed.Rounds[i]
		for _, c := range corrections.Apply(corrections.All, lang, r) {
			p.issue(Issue{Kind: Repaired, Round: RoundNumber(r.DrawNumber), Field: c.Field, Value: c.Wrong, Err: fmt.Errorf("corrected by hand: %s", c.Why)})
		}
	}
}

func (p *parser) issue(i Issue) {
	p.issues = append(p.issues, i)
}

func (p *parser) rounds(en *raw.Feed, french map[string]raw.Round) *Rounds {
	for _, u := range en.Unreadable {
		p.issue(Issue{Kind: Unreadable, Round: RoundNumber(u.Number), Err: u.Err})
	}

	rounds := slices.Clone(en.Rounds)
	slices.SortStableFunc(rounds, func(a, b raw.Round) int {
		return clean.CompareRoundNumbers(string(a.DrawNumber), string(b.DrawNumber))
	})

	out := &Rounds{}
	seen := map[string]bool{}
	pools := map[Date]pooled{}
	for _, r := range rounds {
		n := string(r.DrawNumber)
		switch {
		case n == "":
			p.issue(Issue{Kind: Unreadable, Field: "drawNumber", Err: errors.New("missing round number")})
			continue
		case seen[n]:
			p.issue(Issue{Kind: Unreadable, Round: RoundNumber(n), Field: "drawNumber", Value: n, Err: errors.New("repeated round number")})
			continue
		}
		seen[n] = true

		d := p.draw(r, p.french(r, french))
		if pool, ok := p.pool(r); ok {
			d.PoolAsOf = pool.AsOf
			p.keepPool(pools, n, r.DistributionAsOn, pool)
		}
		out.Draws = append(out.Draws, d)
	}

	p.frenchOnly(french, seen, en.Unreadable)
	slices.Reverse(out.Draws)
	for _, k := range pools {
		out.Pools = append(out.Pools, k.pool)
	}
	slices.SortFunc(out.Pools, func(a, b Pool) int {
		switch {
		case b.AsOf.Before(a.AsOf):
			return -1
		case a.AsOf.Before(b.AsOf):
			return 1
		}
		return 0
	})
	out.Issues = p.issues
	return out
}

// french returns the French entry for en, or an empty one when there is
// none or it looks like another round. Values that differ are reported;
// the English ones are always kept. A field is compared only when both
// files have it and both values read.
func (p *parser) french(en raw.Round, french map[string]raw.Round) raw.Round {
	fr, ok := french[string(en.DrawNumber)]
	if !ok {
		return raw.Round{}
	}
	n := RoundNumber(en.DrawNumber)
	both := func(field string) bool { return en.Keys[field] && fr.Keys[field] }
	mismatch := func(field, enValue, frValue string) {
		p.issue(Issue{Kind: FrenchMismatch, Round: n, Field: field, Value: frValue, Err: fmt.Errorf("the English file has %q", enValue)})
	}
	if both("drawDate") && !sameDay(en.DrawDate, fr.DrawDate) {
		mismatch("drawDate", en.DrawDate, fr.DrawDate)
		return raw.Round{}
	}
	if both("drawSize") && fr.DrawSize != en.DrawSize {
		mismatch("drawSize", itoa(en.DrawSize), itoa(fr.DrawSize))
	}
	if both("drawCRS") && fr.DrawCRS != en.DrawCRS {
		mismatch("drawCRS", itoa(en.DrawCRS), itoa(fr.DrawCRS))
	}

	// A stale French pool would differ in most bands, so all the bands
	// that differ go in one issue.
	var keys, enValues, frValues []string
	for i := range en.Distribution {
		key := fmt.Sprintf("dd%d", i+1)
		if !both(key) || badCount(en, key) || badCount(fr, key) || fr.Distribution[i] == en.Distribution[i] {
			continue
		}
		keys = append(keys, key)
		enValues = append(enValues, itoa(en.Distribution[i]))
		frValues = append(frValues, itoa(fr.Distribution[i]))
	}
	if len(keys) > 0 {
		mismatch(strings.Join(keys, ","), strings.Join(enValues, ","), strings.Join(frValues, ","))
	}
	return fr
}

func sameDay(a, b string) bool {
	a, b = strings.TrimSpace(a), strings.TrimSpace(b)
	da, errA := ParseDate(a)
	db, errB := ParseDate(b)
	if errA != nil || errB != nil {
		return a == b
	}
	return da == db
}

func badCount(r raw.Round, key string) bool {
	return slices.ContainsFunc(r.BadDistribution, func(f raw.BadField) bool { return f.Key == key })
}

func itoa(i raw.Int) string { return strconv.Itoa(int(i)) }

// frenchOnly reports French rounds the English file doesn't have, in round
// order. Rounds English has but couldn't read aren't French-only.
func (p *parser) frenchOnly(french map[string]raw.Round, english map[string]bool, unreadable []raw.Unreadable) {
	var only []string
	for n := range french {
		if n == "" || english[n] || slices.ContainsFunc(unreadable, func(u raw.Unreadable) bool { return string(u.Number) == n }) {
			continue
		}
		only = append(only, n)
	}
	slices.SortFunc(only, func(a, b string) int {
		return cmp.Or(clean.CompareRoundNumbers(a, b), strings.Compare(a, b))
	})
	for _, n := range only {
		p.issue(Issue{Kind: FrenchOnly, Round: RoundNumber(n)})
	}
}

func (p *parser) draw(en, fr raw.Round) Draw {
	n := RoundNumber(en.DrawNumber)
	d := Draw{
		Number:       n,
		Category:     Category(clean.Category(en.DrawName)),
		Name:         Name{EN: en.DrawName, FR: fr.DrawName},
		Invitations:  int(en.DrawSize),
		MinimumCRS:   int(en.DrawCRS),
		Programs:     Name{EN: en.DrawText2, FR: fr.DrawText2},
		Instructions: Links{EN: link(en.DrawNumberURL), FR: link(fr.DrawNumberURL)},
	}
	if date, err := ParseDate(en.DrawDate); err == nil {
		d.Date = date
	} else {
		p.issue(Issue{Kind: Invalid, Round: n, Field: "drawDate", Value: en.DrawDate, Err: err})
	}
	d.InvitedAt = p.timestamp(n, "drawDateTime", en.DrawDateTime, en.DrawDate)
	d.TieBreak = p.timestamp(n, "drawCutOff", en.DrawCutOff, en.DrawDate)
	return d
}

func (p *parser) timestamp(n RoundNumber, field, value, drawDate string) time.Time {
	if value == "" {
		return time.Time{}
	}
	if t, ok := ircctime.Timestamp(value); ok {
		return t
	}
	if t, ok := ircctime.Repair(value, drawDate); ok {
		p.issue(Issue{Kind: Repaired, Round: n, Field: field, Value: value})
		return t
	}
	p.issue(Issue{Kind: Invalid, Round: n, Field: field, Value: value, Err: errors.New("not a time")})
	return time.Time{}
}

// pool reads dd1..dd18. IRCC's order, highest scores first: dd1 601-1200,
// dd2 501-600, dd3 451-500 split into dd4..dd8 (10 points each, 491-500
// down to 451-460), dd9 401-450 split into dd10..dd14, dd15 351-400,
// dd16 301-350, dd17 0-300, dd18 the total. Early rounds have a date but
// no counts; they get no pool.
func (p *parser) pool(r raw.Round) (Pool, bool) {
	if r.DistributionAsOn == "" {
		return Pool{}, false
	}
	if len(r.BadDistribution) > 0 {
		for _, b := range r.BadDistribution {
			p.issue(Issue{Kind: Invalid, Round: RoundNumber(r.DrawNumber), Field: b.Key, Value: b.Value, Err: errors.New("not a count")})
		}
		return Pool{}, false
	}
	day, ok := ircctime.Day(r.DistributionAsOn)
	if !ok {
		p.issue(Issue{Kind: Invalid, Round: RoundNumber(r.DrawNumber), Field: "drawDistributionAsOn", Value: r.DistributionAsOn, Err: errors.New("not a date")})
		return Pool{}, false
	}
	dd := func(i int) int { return int(r.Distribution[i-1]) }
	tens := func(first, top int) []Band {
		out := make([]Band, 5)
		for i := range out {
			hi := top - 10*i
			out[i] = Band{Min: hi - 9, Max: hi, Count: dd(first + i)}
		}
		return out
	}
	pool := Pool{
		AsOf:  DateOf(day),
		Total: dd(18),
		Bands: []Band{
			{Min: 601, Max: 1200, Count: dd(1)},
			{Min: 501, Max: 600, Count: dd(2)},
			{Min: 451, Max: 500, Count: dd(3), Bands: tens(4, 500)},
			{Min: 401, Max: 450, Count: dd(9), Bands: tens(10, 450)},
			{Min: 351, Max: 400, Count: dd(15)},
			{Min: 301, Max: 350, Count: dd(16)},
			{Min: 0, Max: 300, Count: dd(17)},
		},
	}
	return pool, pool.Total != 0
}

type pooled struct {
	round string
	pool  Pool
}

// keepPool keeps the first round's snapshot for each date. Rounds arrive
// oldest first, so a later round that disagrees is reported, not kept.
func (p *parser) keepPool(pools map[Date]pooled, round, asOf string, pool Pool) {
	first, ok := pools[pool.AsOf]
	if !ok {
		pools[pool.AsOf] = pooled{round: round, pool: pool}
		return
	}
	if !samePool(first.pool, pool) {
		p.issue(Issue{
			Kind:  PoolConflict,
			Round: RoundNumber(round),
			Field: "drawDistributionAsOn",
			Value: asOf,
			Err:   fmt.Errorf("pool differs from round %s's, which was kept", first.round),
		})
	}
}

func samePool(a, b Pool) bool {
	return a.AsOf == b.AsOf && a.Total == b.Total && sameBands(a.Bands, b.Bands)
}

func sameBands(a, b []Band) bool {
	return slices.EqualFunc(a, b, func(x, y Band) bool {
		return x.Min == y.Min && x.Max == y.Max && x.Count == y.Count && sameBands(x.Bands, y.Bands)
	})
}

var hrefPattern = regexp.MustCompile(`href=['"]([^'"]+)['"]`)

// link takes the URL out of one of IRCC's HTML anchors and makes it
// absolute. Newer rounds link through the site's internal /content/canadasite
// path, which canada.ca redirects to the public one.
func link(anchor string) string {
	m := hrefPattern.FindStringSubmatch(anchor)
	if m == nil {
		return ""
	}
	u, err := url.Parse(m[1])
	if err != nil {
		return ""
	}
	u = canadaCA.ResolveReference(u)
	if u.Host == canadaCA.Host {
		u.Path = strings.TrimPrefix(u.Path, "/content/canadasite")
	}
	return u.String()
}

var canadaCA = &url.URL{Scheme: "https", Host: "www.canada.ca", Path: "/"}
