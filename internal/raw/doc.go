// Package raw models IRCC's Express Entry rounds file
// (ee_rounds_123_en.json and _fr.json) as published, so the rest of the
// package never touches its JSON.
//
// The file is undocumented. Numbers arrive as strings, often with thousands
// separators ("5,000"), and sometimes as plain numbers; Int accepts both.
// Round numbers aren't always numbers ("91b"), so they are DrawNumber.
package raw
