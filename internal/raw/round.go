package raw

import (
	"encoding/json"
	"fmt"
)

// Bands is how many pool distribution fields (dd1..dd18) a round carries.
const Bands = 18

// Round is one invitation round. Field names follow IRCC's keys.
type Round struct {
	DrawNumber   DrawNumber `json:"drawNumber"`   // usually "453", once "91b"
	DrawDate     string     `json:"drawDate"`     // 2026-11-12
	DrawDateFull string     `json:"drawDateFull"` // November 12, 2026 (localized)
	DrawDateTime string     `json:"drawDateTime"` // as IRCC writes it, e.g. "November 12, 2026 at 14:23:41 UTC"
	DrawName     string     `json:"drawName"`     // category, e.g. "Canadian Experience Class"
	DrawSize     Int        `json:"drawSize"`     // invitations
	DrawCRS      Int        `json:"drawCRS"`      // lowest CRS score invited
	DrawText2    string     `json:"drawText2"`    // programs covered
	DrawCutOff   string     `json:"drawCutOff"`   // tie-break timestamp

	// Pool CRS distribution on DistributionAsOn: Distribution[0] is dd1.
	DistributionAsOn string     `json:"drawDistributionAsOn"`
	Distribution     [Bands]Int `json:"-"`

	// BadDistribution lists the ddN fields that aren't counts. The round
	// still reads; its pool shouldn't be trusted.
	BadDistribution []BadField `json:"-"`
}

// BadField is a field whose value couldn't be read.
type BadField struct {
	Key   string
	Value string
}

// UnmarshalJSON reads the named fields as usual and dd1..dd18 into
// Distribution.
func (r *Round) UnmarshalJSON(b []byte) error {
	type plain Round
	if err := json.Unmarshal(b, (*plain)(r)); err != nil {
		return err
	}
	var all map[string]json.RawMessage
	if err := json.Unmarshal(b, &all); err != nil {
		return err
	}
	for i := range Bands {
		key := fmt.Sprintf("dd%d", i+1)
		raw, ok := all[key]
		if !ok {
			continue
		}
		if err := json.Unmarshal(raw, &r.Distribution[i]); err != nil {
			r.BadDistribution = append(r.BadDistribution, BadField{Key: key, Value: string(raw)})
		}
	}
	return nil
}
