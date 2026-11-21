package raw

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Feed is the whole rounds file.
type Feed struct {
	Rounds     []Round
	Unreadable []Unreadable // rounds left out of Rounds
}

// Unreadable is a round that didn't decode.
type Unreadable struct {
	Index  int        // position in IRCC's rounds array
	Number DrawNumber // drawNumber, when it could be read
	Err    error
}

func (u *Unreadable) Error() string {
	return fmt.Sprintf("round %d (%q): %v", u.Index, u.Number, u.Err)
}

func (u *Unreadable) Unwrap() error {
	return u.Err
}

// UnmarshalJSON decodes each round on its own, so one bad round doesn't
// lose the rest.
func (f *Feed) UnmarshalJSON(b []byte) error {
	var raw struct {
		Rounds []json.RawMessage `json:"rounds"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	f.Rounds, f.Unreadable = nil, nil
	for i, msg := range raw.Rounds {
		var r Round
		if err := json.Unmarshal(msg, &r); err != nil {
			var id struct {
				DrawNumber DrawNumber `json:"drawNumber"`
			}
			_ = json.Unmarshal(msg, &id)
			f.Unreadable = append(f.Unreadable, Unreadable{Index: i, Number: id.DrawNumber, Err: err})
			continue
		}
		f.Rounds = append(f.Rounds, r)
	}
	return nil
}

// Parse decodes a rounds file and checks it looks like one: it must hold at
// least one round that reads.
func Parse(body []byte) (*Feed, error) {
	var f Feed
	if err := json.Unmarshal(body, &f); err != nil {
		return nil, err
	}
	if len(f.Rounds) == 0 {
		if len(f.Unreadable) > 0 {
			return nil, fmt.Errorf("no readable rounds in feed: %w", &f.Unreadable[0])
		}
		return nil, errors.New("no rounds in feed")
	}
	return &f, nil
}
