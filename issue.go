package expressentry

import "fmt"

// IssueKind says what went wrong with a round.
type IssueKind int

const (
	// Unreadable: the round couldn't be read and is left out of
	// [Rounds.Draws].
	Unreadable IssueKind = iota + 1
	// Invalid: a field couldn't be read and is left empty; the round is
	// kept.
	Invalid
	// Repaired: a field had a known typo and was read anyway.
	Repaired
	// PoolConflict: two rounds gave different numbers for the same pool
	// snapshot. The earlier round's numbers are kept.
	PoolConflict
	// FrenchUnavailable: the French file couldn't be read, so every [Name]
	// has only its English text, and Round is empty. With Round set, only
	// that French round couldn't be read.
	FrenchUnavailable
	// FrenchOnly: a round is in the French file but not the English one.
	// It is left out of [Rounds.Draws], since the English file decides
	// which rounds exist.
	FrenchOnly
	// FrenchMismatch: a round's French entry gives Field a different Value
	// from the English one. The English value is kept. When the dates
	// differ, the French entry is taken to be another round and its text
	// isn't used.
	FrenchMismatch
)

func (k IssueKind) String() string {
	switch k {
	case Unreadable:
		return "unreadable"
	case Invalid:
		return "invalid"
	case Repaired:
		return "repaired"
	case PoolConflict:
		return "pool conflict"
	case FrenchUnavailable:
		return "French unavailable"
	case FrenchOnly:
		return "French only"
	case FrenchMismatch:
		return "French mismatch"
	default:
		return fmt.Sprintf("IssueKind(%d)", int(k))
	}
}

// Issue is one problem found in the rounds files.
type Issue struct {
	Kind  IssueKind
	Round RoundNumber
	// Field is IRCC's JSON key for the value, such as "drawCutOff", or ""
	// when the issue is about the whole round.
	Field string
	// Value is the text as IRCC published it.
	Value string
	Err   error
}

func (i Issue) String() string {
	s := i.Kind.String()
	if i.Round != "" {
		s = fmt.Sprintf("round %s: %s", i.Round, i.Kind)
	}
	if i.Field != "" {
		s += fmt.Sprintf(" %s %q", i.Field, i.Value)
	}
	if i.Err != nil {
		s += ": " + i.Err.Error()
	}
	return s
}
