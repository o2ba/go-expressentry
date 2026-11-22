package raw

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// DrawNumber is a round number: usually "453", sometimes a bare 453, once
// "91b". null reads as "".
type DrawNumber string

func (t *DrawNumber) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	switch {
	case bytes.Equal(b, []byte("null")):
		*t = ""
		return nil
	case bytes.HasPrefix(b, []byte(`"`)):
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		*t = DrawNumber(strings.TrimSpace(s))
		return nil
	}
	var n json.Number
	if err := json.Unmarshal(b, &n); err != nil {
		return fmt.Errorf("not text or a number: %s", b)
	}
	*t = DrawNumber(n)
	return nil
}

// Int is a whole number IRCC may write as a number, a string, or a string
// with thousands separators. An empty string or null reads as 0.
type Int int

func (n *Int) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if bytes.Equal(b, []byte("null")) {
		*n = 0
		return nil
	}
	s := string(b)
	if strings.HasPrefix(s, `"`) {
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		s = strings.NewReplacer(",", "", " ", "", " ", "", " ", "").Replace(s)
		if s == "" {
			*n = 0
			return nil
		}
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("not a whole number: %s", b)
	}
	*n = Int(v)
	return nil
}
