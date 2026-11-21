package expressentry

import "testing"

func FuzzParse(f *testing.F) {
	f.Add([]byte(enFile), []byte(frFile))
	f.Add([]byte(`{"rounds":[{"drawNumber":"91b","drawSize":"5 000","drawCutOff":"October 7 at 16:38:21 UTC"}]}`), []byte(nil))
	f.Fuzz(func(t *testing.T, en, fr []byte) {
		r, err := Parse(en, fr)
		if err != nil {
			return
		}
		for _, d := range r.Draws {
			if d.Number == "" || !d.Category.Valid() {
				t.Fatalf("draw = %+v", d)
			}
		}
	})
}
