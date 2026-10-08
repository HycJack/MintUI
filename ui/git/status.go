// Package git holds the parts of a version-control interface that are not the
// window: a status has a letter and a word, a diff is a list of lines with
// numbers on both sides, a branch list is a slice.
//
// The reason it is written this way is that none of it can be checked by
// looking at a screenshot. A wrong line number in a diff reads exactly like a
// right one; a status that says Modified where it should say Renamed is a
// colour that looks plausible. So the parts that decide are plain functions
// over plain data, tested by their return values, and the components below
// only draw what those functions said.
package git

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// FileStatus is where one path stands against the commit a working tree is
// checked out at.
//
// It is the zero value Clean: a path with nothing wrong with it is the common
// case and the one a caller should get without saying anything.
type FileStatus int

const (
	// Clean is tracked and identical to the commit.
	Clean FileStatus = iota
	// Modified is tracked and edited.
	Modified
	// Added is staged as a new path.
	Added
	// Deleted is tracked and removed.
	Deleted
	// Renamed is tracked under a new name.
	Renamed
	// Untracked is not in the repository at all — git's own "??" column.
	Untracked
	// Conflicted is a path a merge left with two versions in it. It is the
	// one status that has to be louder than the rest.
	Conflicted
	// Typechange is the same path with a different kind in it, such as a
	// file that became a symlink.
	Typechange
)

// statusFace is everything a status says at once: the letter in a two-column
// listing, the word beside it, and how loudly it should be said.
//
// It is a value rather than three functions because the three never disagree
// with each other in a correct drawing, and three functions let them drift.
type statusFace struct {
	// Letter is the one or two characters git puts in its XY column. It is
	// what a person who reads git output every day is looking for, so it is
	// the same letter git uses rather than an icon of our own.
	Letter string
	// Word is the same thing for somebody who does not.
	Word string
	// Severity is how loudly to say it, which is the only thing that decides
	// a colour. Nothing in this package picks a colour.
	Severity core.Severity
}

// faces is the whole status vocabulary, in one place, so a new status is one
// line rather than three edits in three files that can disagree.
//
// The index is the FileStatus, which is why Clean is written out rather than
// left to the zero value of a slice entry.
var faces = [...]statusFace{
	Clean:      {"", "Clean", core.Neutral},
	Modified:   {"M", "Modified", core.Warning},
	Added:      {"A", "Added", core.Success},
	Deleted:    {"D", "Deleted", core.Danger},
	Renamed:    {"R", "Renamed", core.Accent},
	Untracked:  {"??", "Untracked", core.Accent},
	Conflicted: {"U", "Conflicted", core.Danger},
	Typechange: {"T", "Typechange", core.Warning},
}

// valid reports whether s is one of the statuses above. A FileStatus that
// came from a byte of a status file and not from these names is a bug in the
// caller, and it is caught here rather than drawing an empty square.
func (s FileStatus) valid() bool { return s >= Clean && int(s) < len(faces) }

// face is the mapping itself, panicking on a status nobody defined.
func (s FileStatus) face() statusFace {
	if !s.valid() {
		panic("git: unknown FileStatus " + itoa(int(s)))
	}
	return faces[s]
}

// String is the status's word, so a %v in a test or a log says "Modified"
// rather than "1".
func (s FileStatus) String() string { return s.face().Word }

// StatusLetter is the one or two characters git puts in a two-column
// listing: M, A, D, R, ?? and so on. Clean is the empty string rather than a
// space, because a clean path has nothing to say and a column of blanks is
// what a clean working tree looks like.
//
// A status nobody defined panics: this is the mapping the whole package
// draws from, and a silently blank letter is a file nobody can find.
func StatusLetter(s FileStatus) string { return s.face().Letter }

// StatusWord is the status in full, for a tooltip, a label for assistive
// technology and a callout that has room for it.
func StatusWord(s FileStatus) string { return s.face().Word }

// StatusSeverity is how loudly the status should be said, and the only thing
// in this package that decides a colour. The colour itself is StatusInk.
func StatusSeverity(s FileStatus) core.Severity { return s.face().Severity }

// StatusInk is the ink the letter is drawn in: the foreground of its
// severity, or the window's ordinary text when it has none.
//
// Clean has no severity colour on purpose. A working tree with nothing in it
// is the resting state of every repository, and painting its rows in the
// window's ink is what makes it read as resting rather than as an alert.
func StatusInk(s FileStatus, k theme.Tokens) ui.Color {
	_, fg := s.face().Severity.Pair(k)
	return fg
}

// StatusPair is the background and ink a status is drawn with, for a pill or
// a row that wants to be as loud as the status rather than as loud as its
// letter. Clean gets a plain surface rather than the neutral pair, because
// neutral's foreground is body ink and a body-ink pill on a body-ink page is
// a pill nobody notices.
func StatusPair(s FileStatus, k theme.Tokens) (bg, fg ui.Color) {
	if s == Clean {
		return k.Surface, k.TextMuted
	}
	return s.face().Severity.Pair(k)
}

// itoa is a small int-to-string for the two panics in this package that have
// to name a number. It is here rather than in internal because the messages
// are part of what this file is for.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
