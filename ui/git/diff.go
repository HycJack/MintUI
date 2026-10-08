package git

import (
	"strconv"
	"strings"
)

// This file is the whole reason the package is shaped the way it is: a diff
// is data, and the component that draws it draws these structs and nothing
// else.
//
// Line numbers are the part that cannot be checked by eye. A viewer that is
// one line out looks exactly like a viewer that is right until somebody
// notices the "42" beside a line that is 41 in their editor. So the numbers
// are computed once, here, from the hunk headers, and tested against a real
// diff rather than looked at.

// DiffOp is what happened to one line.
type DiffOp int

const (
	// OpContext is a line that is in both files.
	OpContext DiffOp = iota
	// OpAdd is a line only the new file has.
	OpAdd
	// OpDel is a line only the old file has.
	OpDel
)

func (o DiffOp) String() string {
	switch o {
	case OpAdd:
		return "+"
	case OpDel:
		return "-"
	}
	return " "
}

// DiffLine is one line of a diff, with where it sat on each side.
//
// A line is in one file or the other or both, and the number it has in the
// file it is not in is 0 rather than a copy of the other one: "line 42" on
// the right of a viewer, with nothing on the left, is an addition, and giving
// it 42 on both sides is how a viewer ends up with two columns of numbers
// that mean different things.
type DiffLine struct {
	// Op is what happened to the line.
	Op DiffOp
	// OldNo is the line's 1-based number in the old file, or 0 when it is
	// not in the old file.
	OldNo int
	// NewNo is the line's 1-based number in the new file, or 0 when it is
	// not in the new file.
	NewNo int
	// Text is the line itself, with the +/-/space marker taken off. What is
	// left is what the editor shows, so a viewer can print it unchanged.
	Text string
}

// Hunk is one @@ section of a diff: the run of changes between two contexts.
type Hunk struct {
	// Header is the @@ line verbatim, which carries the function name git
	// puts after the second @@ and which is worth showing — it is how a
	// reader knows which function they are looking at without scrolling.
	Header string
	// OldFrom and NewFrom are where each side's run starts.
	OldFrom, NewFrom int
	// OldCount and NewCount are how many lines each side's run has. They come
	// from the header and are kept as they were found, because they are what
	// a caller checks a truncated diff against; the lines' own numbers are
	// counted from the lines, not from here.
	OldCount, NewCount int
	// Lines are the hunk's lines, in order.
	Lines []DiffLine
}

// Hunks cuts a unified diff into its hunks.
//
// The file headers git writes before the first @@ — "diff --git", "index",
// "---", "+++", "rename from" and the rest — are not hunks and are not
// returned. They describe the file rather than its contents, and a caller
// showing them has to want them.
//
// Anything unrecognised is skipped rather than being an error. The input here
// is git's output and a patch out of a build system, not an argument a caller
// got wrong, and a viewer that panics on a header it did not expect is a
// viewer that crashes on a real repository. A @@ line whose two ranges do not
// parse is skipped the same way, and the lines before the next one that does
// parse are simply not shown — which is the one thing this function can get
// wrong, and it is the kind of wrong a person reads past.
func Hunks(unified string) []Hunk {
	var out []Hunk
	// cur is the hunk being filled, or -1 when the line is not in one. It is
	// an index rather than a pointer into out because append can move the
	// backing array out from under a pointer.
	cur := -1
	// old and next are the running line numbers, seeded from each header.
	old, next := 0, 0

	for _, line := range splitLines(unified) {
		if strings.HasPrefix(line, "@@") {
			from, count, nf, nc, ok := parseHunkHeader(line)
			if !ok {
				cur = -1
				continue
			}
			out = append(out, Hunk{
				Header: line, OldFrom: from, NewFrom: nf,
				OldCount: count, NewCount: nc,
			})
			cur = len(out) - 1
			old, next = from, nf
			continue
		}
		if cur < 0 {
			continue
		}
		if d, ok := hunkLine(line, &old, &next); ok {
			out[cur].Lines = append(out[cur].Lines, d)
		}
	}
	return out
}

// SplitDiff is every hunk's lines, in order, as one list — which is what a
// viewer that draws one continuous column wants.
//
// It is Hunks flattened, so the two can never disagree about what a diff
// says, and it returns nil for a diff with no hunks rather than a list of
// blanks.
func SplitDiff(unified string) []DiffLine {
	hunks := Hunks(unified)
	var out []DiffLine
	for _, h := range hunks {
		out = append(out, h.Lines...)
	}
	return out
}

// hunkLine turns one line of a hunk's body into a DiffLine, advancing the two
// running numbers. It reports false for the lines that are not content: git's
// "\ No newline at end of file" note, and a trailing empty line left by a file
// that ended in a newline.
//
// The note is skipped rather than drawn because it is a remark about the line
// above it, not a line of either file — and because it consumes no line
// number, drawing it would push every number under it out by one.
func hunkLine(line string, old, next *int) (DiffLine, bool) {
	if strings.HasPrefix(line, `\`) {
		return DiffLine{}, false
	}
	// The counters are seeded from the @@ header and hold the number the
	// NEXT line will have, so a line takes the counter's value and the
	// counter then moves on. Incrementing first and printing the counter is
	// off by one on every line of every diff, which is the kind of wrong
	// nobody notices until somebody follows a line number into their editor.
	switch {
	case strings.HasPrefix(line, "+"):
		d := DiffLine{Op: OpAdd, NewNo: *next, Text: line[1:]}
		*next++
		return d, true
	case strings.HasPrefix(line, "-"):
		d := DiffLine{Op: OpDel, OldNo: *old, Text: line[1:]}
		*old++
		return d, true
	}
	// Everything else in a hunk body is context, with or without the leading
	// space. git writes one; tools that have trimmed trailing whitespace
	// write an empty line for a context line that is itself empty, and
	// dropping those would silently shorten a diff of a file full of blanks.
	d := DiffLine{Op: OpContext, OldNo: *old, NewNo: *next, Text: strings.TrimPrefix(line, " ")}
	*old++
	*next++
	return d, true
}

// splitLines is a diff as its lines, without the empty piece a trailing
// newline leaves behind.
//
// It matters more than it looks: a file that ends in a newline — which is
// every text file anybody writes — hands back a final empty string, and a
// parser that reads that as a context line draws a blank row at the bottom
// of every diff and numbers every line above it one too high.
func splitLines(unified string) []string {
	lines := strings.Split(unified, "\n")
	if n := len(lines); n > 0 && lines[n-1] == "" {
		lines = lines[:n-1]
	}
	return lines
}

// parseHunkHeader reads "@@ -old,count +new,count @@ optional" and gives the
// two starts and the two counts. A count git leaves out is 1, which is the
// unified format's own rule and the reason "@@ -1 +1 @@" is a real thing to
// have to parse.
func parseHunkHeader(line string) (oldFrom, oldCount, newFrom, newCount int, ok bool) {
	rest := strings.TrimPrefix(line, "@@")
	// The header's second "@@" closes the ranges; anything after it is the
	// function name and is kept in Hunk.Header rather than parsed.
	if end := strings.Index(rest, "@@"); end >= 0 {
		rest = rest[:end]
	}
	fields := strings.Fields(rest)
	if len(fields) != 2 {
		return 0, 0, 0, 0, false
	}
	oldFrom, oldCount, ok = parseRange(fields[0], '-')
	if !ok {
		return 0, 0, 0, 0, false
	}
	newFrom, newCount, ok = parseRange(fields[1], '+')
	if !ok {
		return 0, 0, 0, 0, false
	}
	return oldFrom, oldCount, newFrom, newCount, true
}

// parseRange reads one side of a hunk header: the sign, the first line, and
// the count after a comma.
func parseRange(field string, sign byte) (from, count int, ok bool) {
	if len(field) == 0 || field[0] != sign {
		return 0, 0, false
	}
	rest := field[1:]
	before, after, hasCount := strings.Cut(rest, ",")
	from, err := strconv.Atoi(before)
	if err != nil || from < 0 {
		return 0, 0, false
	}
	count = 1
	if hasCount {
		count, err = strconv.Atoi(after)
		if err != nil || count < 0 {
			return 0, 0, false
		}
	}
	return from, count, true
}

// Counts is how many lines a diff adds and deletes, and nothing else. It is
// the number every summary in a git interface is made of, and it is here as a
// function so that a stat line, a row's two figures and a badge cannot each
// count a diff their own way.
func Counts(lines []DiffLine) (added, deleted int) {
	for _, l := range lines {
		switch l.Op {
		case OpAdd:
			added++
		case OpDel:
			deleted++
		}
	}
	return added, deleted
}

// StatWord is the "+4 -1" every git interface puts beside a file, built from
// Counts so that it cannot disagree with the numbers it sits next to.
//
// A count of zero is left out, because "+0 -0" says "changed" about a file
// nothing happened to.
func StatWord(added, deleted int) string {
	var parts []string
	if added > 0 {
		parts = append(parts, "+"+itoa(added))
	}
	if deleted > 0 {
		parts = append(parts, "-"+itoa(deleted))
	}
	if len(parts) == 0 {
		return "no changes"
	}
	return strings.Join(parts, " ")
}

// StatWordOf is StatWord for a whole diff, in one call.
func StatWordOf(unified string) string {
	added, deleted := Counts(SplitDiff(unified))
	return StatWord(added, deleted)
}
