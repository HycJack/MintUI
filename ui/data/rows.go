package data

import (
	"fmt"
	"slices"
	"strings"

	"github.com/egoist/mygo/ui"
)

// This file holds the two halves of a table that are not drawing: the order
// its rows are in, and which of them are chosen.
//
// They are here as plain functions because they are the two things a table
// gets wrong most easily, and neither can be checked by looking at a
// screenshot. An unstable sort shuffles equal rows every time the user
// clicks a second column; a Shift-click measured from the top of the table
// instead of from the row last chosen takes rows the user never touched.
// Both look right in a picture and are wrong in the data.

// Sort says which column the rows are ordered by, and in which direction.
//
// It is the library's own sort order: the header of a DataTable reads the
// one in the caller's state, and Rows orders the caller's rows with it — the
// arrow over a column and the order of the rows cannot disagree about which
// column is sorted, because they are the same two words.
type Sort = ui.SortOrder

// Rows returns rows ordered by sort, comparing two of them with the
// comparator the caller registered for that column — Rows has no opinion
// what a column's values mean, and guessing is how "$9" sorts above "$10".
//
// The sort is stable: rows that compare equal keep the order they came in,
// whichever way the column is sorted. A user who clicks Priority and then
// Technician has said something about priority and technician, not about the
// order two customers happened to be loaded in; without stability the third
// click would quietly reshuffle them again. Ascending and descending are
// both stable, so a column sorted twice reads the same as it did the first
// time.
//
// Rows copies. A table that reorders the slice it was handed would reorder
// the caller's data under it, and a view that is a function of its state
// would show a different table each time it ran.
//
// A column with no comparator cannot be sorted, and asking for one panics:
// quietly returning the rows in the order they came in would look like a
// sort that did nothing.
func Rows[R any](rows []R, sort Sort, by map[string]func(a, b R) int) []R {
	out := slices.Clone(rows)
	if sort.Column == "" || len(out) < 2 {
		return out
	}
	less, ok := by[sort.Column]
	if !ok {
		panic(fmt.Sprintf("data: Rows cannot sort by %q; give it a comparator, or take the column out of the table", sort.Column))
	}
	if sort.Descending {
		slices.SortStableFunc(out, func(a, b R) int { return less(b, a) })
		return out
	}
	slices.SortStableFunc(out, less)
	return out
}

// Toggle returns the sort a click on a column's header asks for: that column
// ascending from another, and the other way round from itself. A header
// clicked twice therefore reads the same order as a column sorted once, in
// the opposite direction — never a third state.
func Toggle(sort Sort, column string) Sort {
	if sort.Column == column {
		sort.Descending = !sort.Descending
		return sort
	}
	return Sort{Column: column}
}

// ByText compares two rows by a string of theirs, for a column of names,
// references and words.
func ByText[R any](key func(R) string) func(a, b R) int {
	return func(a, b R) int { return strings.Compare(key(a), key(b)) }
}

// ByNumber compares two rows by a number of theirs, for a column of figures.
// It takes a float64 because money arrives as cents, hours as fractions, and
// a comparator that cannot be handed either is one nobody uses.
func ByNumber[R any](key func(R) float64) func(a, b R) int {
	return func(a, b R) int {
		x, y := key(a), key(b)
		switch {
		case x < y:
			return -1
		case x > y:
			return 1
		}
		return 0
	}
}

// Selectable is a set of chosen rows, and the row a Shift-click measures its
// range from.
//
// It is the caller's, not the table's: the table adds to it as rows are
// clicked and the caller reads it to say "3 of 27 selected". Its zero value
// is an empty choice with no anchor, so a Shift-click before any other click
// chooses the row clicked rather than a range from somewhere unknown.
type Selectable struct {
	chosen   map[int]bool
	anchor   int
	anchored bool
}

// Click chooses row as the click on it asks for: a plain click chooses that
// row alone, add puts it to the choice or takes it out, and extend chooses
// every row from the anchor to row, both ways.
//
// The anchor is the row last chosen by a plain or an added click — not the
// first row, and not where the table begins. It is what makes Shift-click do
// what a user means by it: click row 0, click row 9 with Shift, and rows 0
// through 9 are chosen — not the rows 0 through 8 that a range measured
// from the top of the table picks once the table is scrolled down.
//
// The anchor stays where it was after an extend, so a second Shift-click
// reaches from the same row and widens the range in the other direction, as
// it does everywhere else.
func (s *Selectable) Click(row int, add, extend bool) {
	switch {
	case extend && s.anchored:
		lo, hi := min(s.anchor, row), max(s.anchor, row)
		if !add {
			s.None()
		}
		for r := lo; r <= hi; r++ {
			s.set(r)
		}
	case add:
		if s.Has(row) {
			s.unset(row)
		} else {
			s.set(row)
		}
		s.anchor, s.anchored = row, true
	default:
		s.None()
		s.set(row)
		s.anchor, s.anchored = row, true
	}
}

// All chooses every one of the count rows, for a Select all in a toolbar.
func (s *Selectable) All(count int) {
	for r := range max(count, 0) {
		s.set(r)
	}
}

// None chooses no row.
func (s *Selectable) None() {
	clear(s.chosen)
}

// Invert chooses the rows of the count that are not chosen and takes the
// others out: the complement, over the rows there are. A toolbar's Invert
// means exactly this, and the count is the table's to say — an inverted
// choice over the rows on screen is not the same set as one over the table.
func (s *Selectable) Invert(count int) {
	for r := range max(count, 0) {
		if s.Has(r) {
			s.unset(r)
		} else {
			s.set(r)
		}
	}
}

// Has reports whether a row is chosen.
func (s *Selectable) Has(row int) bool { return s.chosen[row] }

// Len returns how many rows are chosen.
func (s *Selectable) Len() int { return len(s.chosen) }

// Anchor returns the row an extend measures from, and -1 before any click
// has chosen one.
func (s *Selectable) Anchor() int {
	if !s.anchored {
		return -1
	}
	return s.anchor
}

// Rows returns the chosen rows, ascending, so a caller reading it says the
// same thing twice in the same order.
func (s *Selectable) Rows() []int {
	out := make([]int, 0, len(s.chosen))
	for r := range s.chosen {
		out = append(out, r)
	}
	slices.Sort(out)
	return out
}

func (s *Selectable) set(row int) {
	if s.chosen == nil {
		s.chosen = map[int]bool{}
	}
	s.chosen[row] = true
}

func (s *Selectable) unset(row int) { delete(s.chosen, row) }
