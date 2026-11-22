// Package corrections fixes known mistakes in past rounds by hand.
//
// A correction applies only while IRCC's value is still exactly the wrong
// one, so it stops on its own once IRCC fixes or edits the field. Mistakes
// in future rounds aren't covered; ircctime repairs the kinds of typo
// that recur.
package corrections

import "github.com/o2ba/go-expressentry/internal/raw"

// Lang names a rounds file.
type Lang string

const (
	English Lang = "en"
	French  Lang = "fr"
)

// Correction replaces one field of one round in one file.
type Correction struct {
	Lang  Lang
	Round string
	Field string // IRCC's JSON key
	Wrong string // the value as IRCC published it
	Right string
	Why   string
}

const frInstructions = "/fr/immigration-refugies-citoyennete/organisation/mandat/politiques-directives-operationnelles-ententes-accords/instructions-ministerielles/entree-express-rondes/"

// All lists every correction.
var All = []Correction{
	{
		Lang:  French,
		Round: "149",
		Field: "drawNumberURL",
		Wrong: "<a href='" + frInstructions + "invitations-14.html9'>149</a>",
		Right: "<a href='" + frInstructions + "invitations-149.html'>149</a>",
		Why:   "the link's round number was split around .html",
	},
}

// Apply makes the corrections in list that match r, a round from the lang
// file, and returns the ones it made.
func Apply(list []Correction, lang Lang, r *raw.Round) []Correction {
	var made []Correction
	for _, c := range list {
		if c.Lang != lang || c.Round != string(r.DrawNumber) {
			continue
		}
		f := field(r, c.Field)
		if f == nil || *f != c.Wrong {
			continue
		}
		*f = c.Right
		made = append(made, c)
	}
	return made
}

func field(r *raw.Round, key string) *string {
	switch key {
	case "drawDate":
		return &r.DrawDate
	case "drawDateTime":
		return &r.DrawDateTime
	case "drawName":
		return &r.DrawName
	case "drawText2":
		return &r.DrawText2
	case "drawCutOff":
		return &r.DrawCutOff
	case "drawDistributionAsOn":
		return &r.DistributionAsOn
	case "drawNumberURL":
		return &r.DrawNumberURL
	}
	return nil
}
