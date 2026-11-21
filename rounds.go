package expressentry

import "time"

// Rounds is everything read from the rounds files.
type Rounds struct {
	// Draws lists every readable round, newest first.
	Draws []Draw
	// Pools lists each CRS pool snapshot once, newest first. Several rounds
	// can share a snapshot.
	Pools []Pool
	// Issues lists rounds left out, values repaired and snapshots that
	// disagree, in the order they were found.
	Issues []Issue
}

// Draw is one round of invitations.
type Draw struct {
	Number RoundNumber
	Date   Date
	// InvitedAt is when IRCC sent the invitations, in UTC, or the zero Time
	// when IRCC gives no time.
	InvitedAt time.Time
	Category  Category
	// Name is IRCC's own name for the round, such as "Canadian Experience
	// Class".
	Name        Name
	Invitations int
	MinimumCRS  int
	// TieBreak is the tie-breaking rule: among candidates with MinimumCRS,
	// only those who entered the pool before it were invited. It is in UTC,
	// as IRCC publishes it, or the zero Time when IRCC gives none.
	TieBreak time.Time
	// Programs names the programs the round covered.
	Programs Name
	// PoolAsOf is the date of the [Pool] snapshot published with this round,
	// or the zero Date when there is none.
	PoolAsOf Date
}

// RoundNumber is IRCC's number for a round. It is text because IRCC once
// numbered a round "91b".
type RoundNumber string

// Pool is the CRS score distribution of every profile in the pool on AsOf.
type Pool struct {
	AsOf  Date
	Total int
	// Bands lists score ranges, highest scores first.
	Bands []Band
}

// Band counts profiles with a CRS score from Min to Max, both included.
// IRCC splits the 401-450 and 451-500 bands again; those parts are in
// Bands.
type Band struct {
	Min   int
	Max   int
	Count int
	Bands []Band
}

// Name is text in English and French. FR is empty when no French file was
// read.
type Name struct {
	EN string
	FR string
}
