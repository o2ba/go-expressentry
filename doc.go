// Package expressentry reads the Express Entry rounds of invitations that
// Immigration, Refugees and Citizenship Canada (IRCC) publishes, and returns
// them as typed Go values.
//
// IRCC publishes every round in two undocumented JSON files, one in English
// and one in French, behind its "Rounds of invitations" page. The files are
// edited by hand: numbers arrive as strings with separators, dates are
// written several ways, and an old round can be malformed. This package
// absorbs that, so callers get one [Rounds] value with real dates, numbers
// and categories.
//
// Use a [Client] to download and read the files:
//
//	c := expressentry.NewClient()
//	r, err := c.Rounds(ctx)
//
// or [Parse] when you already have the bytes.
//
// A problem with one round never fails the whole call. The round is left
// out or repaired, and the problem is listed in [Rounds.Issues]. An error
// means nothing usable came back.
//
// The package keeps no state and polls nothing. Calling [Client.Rounds] on
// a schedule, storing results and detecting new rounds are the caller's job.
//
// This package is not affiliated with or endorsed by IRCC.
package expressentry
