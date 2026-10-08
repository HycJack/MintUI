package git

// A commit graph is drawn, not written: the dot you see beside a commit is
// where the algorithm below decided it goes, and it is decided from the
// parents rather than from the order the commits happened to arrive in.
//
// The whole computation is a pure function of the commits and the width the
// caller has room for, so the awkward cases — a merge that opens a lane, a
// branch that closes one, a commit that is already in a lane — are tested as
// numbers rather than judged from a picture.

// Commit is one commit, as a graph needs to know it.
//
// The times are strings on purpose. Formatting a date needs a timezone and a
// reference instant, and a library that picked one would be wrong for half the
// world and untestable for all of it. Whoever has the commit's clock formats
// the string.
type Commit struct {
	// Hash is the full commit hash. It is what matches one commit to a lane.
	Hash string
	// Short is the abbreviated hash a graph draws — seven characters, as git
	// does. Empty falls back to the first seven of Hash.
	Short string
	// Subject is the commit's one-line message.
	Subject string
	// Author is who wrote it, already formatted for display.
	Author string
	// When is when it was, already formatted for display.
	When string
	// Parents are the full hashes of the commits it came from, in order. The
	// first parent continues the lane the commit is in; the rest are the
	// lanes a merge opens. Empty is the first commit in a repository.
	Parents []string
	// Refs are the branch and tag names pointing at this commit, drawn
	// beside it. Empty draws nothing, which is most commits.
	Refs []string
}

// short is the seven characters a graph prints.
func (c Commit) short() string {
	if c.Short != "" {
		return c.Short
	}
	if len(c.Hash) <= 7 {
		return c.Hash
	}
	return c.Hash[:7]
}

// GraphRow is one commit laid out in the graph: where its node sits, which
// lanes are still open under it, and how wide the drawing has to be.
//
// Lanes and Column are numbers rather than marks so that a caller can draw
// the graph with its own glyphs — a lane is a line at x = lane * spacing,
// and whether that line is a merge or a branch tip is the caller's own
// painting.
type GraphRow struct {
	// Commit is the commit the row is for.
	Commit Commit
	// Column is the lane this commit's node sits in. On a linear history
	// every row is 0, which is why a graph with no merges is a single line.
	Column int
	// Merge reports that the commit has more than one parent, and so is
	// where two lines become one.
	Merge bool
	// Lanes are the lanes still open on the row below this one, ascending.
	// A row with no lanes below it is the tip of a branch.
	Lanes []int
	// Width is how many lane columns the view has to draw across for this
	// row. It is the widest row's number rather than one number for the
	// whole graph, so a graph that closes its lanes gets a gutter that
	// closes with it.
	Width int
	// Over reports that a lane was left out because it did not fit in the
	// width given. A caller that shows it should say something: a graph
	// that quietly drops a branch looks like a graph with no branch.
	Over bool
}

// CommitsFor lays commits out as graph rows.
//
// commits is oldest first — the order the history happened in, which is the
// reverse of what `git log` prints. A caller with a log reverses it once,
// here, rather than every call site having to remember which way round this
// is; getting it the wrong way round produces a graph that looks plausible
// and is not, which is the worst possible failure for this function.
//
// The lanes themselves are worked out from the other end. A lane is "the
// commit that comes next", and going forwards from the root there is no way
// to know which commit that is until the list has been read to the end — so
// the walk runs from the newest backwards, where the first commit's parents
// say exactly which lanes the rows below it need, and the rows are then
// turned back round for drawing. Doing it forwards is what makes a merge
// open a lane of its own at the wrong end of the graph.
//
// width is how many lane columns the caller has room for. Zero or less means
// as many as the history needs, which is what a caller sizing its gutter
// from the result wants. Past the width, lanes are not tracked: a graph in
// three columns with a fourth lane would draw a line into the text, so the
// lane is dropped, the commit that needed it is drawn in the last lane there
// is, and the rows from there on say Over.
//
// A commit with a parent nothing in the list points at — a history the
// caller truncated, or a shallow clone — opens its own lane rather than
// panicking. A graph that cannot be drawn at all because one commit is
// missing is the wrong thing to do with a truncated list.
func CommitsFor(commits []Commit, width int) []GraphRow {
	if len(commits) == 0 {
		return nil
	}
	// Backwards: newest first, which is the order lanes can be read in.
	newest := make([]Commit, 0, len(commits))
	for i := len(commits) - 1; i >= 0; i-- {
		newest = append(newest, commits[i])
	}

	var lanes []string
	over := false
	// walked[i] is the row for newest[i], with the lanes that cross the row
	// and go on to the commits above it.
	walked := make([]GraphRow, 0, len(newest))

	for _, cm := range newest {
		col := laneOf(lanes, cm.Hash)
		if col < 0 {
			if width > 0 && len(lanes) >= width {
				// No room for this commit's own lane. It is still shown; only
				// the line leading to it is missing.
				over = true
				col = width - 1
				if col < 0 {
					col = 0
				}
			} else {
				lanes = append(lanes, cm.Hash)
				col = len(lanes) - 1
			}
		}

		lanes, over = carry(lanes, col, cm.Parents, width, over)
		walked = append(walked, GraphRow{
			Commit: cm,
			Column: col,
			Merge:  len(cm.Parents) > 1,
			Lanes:  openLanes(lanes),
			Width:  max(len(lanes), col+1),
		})
	}

	// Round the rows back for drawing. Nothing else changes: a row's Lanes
	// are the lanes below it, and the walk that computed them ran from the
	// top down, so they are already the lanes that carry on from this row to
	// the ones under it. Only the order of the rows is reversed.
	out := make([]GraphRow, len(walked))
	for i := range walked {
		src := walked[len(walked)-1-i]
		src.Width = max(src.Width, len(src.Lanes))
		src.Over = over
		out[i] = src
	}
	return out
}

// openLanes is which of the lanes are waiting for a commit, ascending.
func openLanes(lanes []string) []int {
	if len(lanes) == 0 {
		return nil
	}
	out := make([]int, 0, len(lanes))
	for i, h := range lanes {
		if h != "" {
			out = append(out, i)
		}
	}
	return out
}

// laneOf is the lane waiting for hash, or -1.
func laneOf(lanes []string, hash string) int {
	for i, h := range lanes {
		if h == hash {
			return i
		}
	}
	return -1
}

// carry moves a lane past one commit: the first parent continues it, any
// other parent opens a lane beside it, and a commit with no parents closes
// its lane.
//
// The side lanes go in just after col rather than at the end, so a merge's
// second parent is drawn next to the line it merged into rather than out at
// the far edge with a long horizontal run to get back.
func carry(lanes []string, col int, parents []string, width int, over bool) ([]string, bool) {
	side := parents[min(1, len(parents)):]

	next := make([]string, 0, len(lanes)+len(side))
	next = append(next, lanes[:col]...)
	if len(parents) > 0 {
		next = append(next, parents[0])
	} else {
		next = append(next, "")
	}
	tail := lanes[min(col+1, len(lanes)):]
	for _, p := range side {
		if width > 0 && len(next)+len(tail) >= width {
			over = true
			break
		}
		next = append(next, p)
	}
	next = append(next, tail...)

	// A merge closes a branch: its second parent is already waiting in a
	// lane of its own, and adding a second lane for it would draw the same
	// line twice and push every lane after it one column out.
	return unique(squeeze(next)), over
}

// squeeze drops the closed lanes out of the list, keeping the rest in order.
// A closed lane is not a hole in the graph, it is a gap between two lines.
func squeeze(lanes []string) []string {
	out := make([]string, 0, len(lanes))
	for _, h := range lanes {
		if h != "" {
			out = append(out, h)
		}
	}
	return out
}

// unique drops every lane after the first one waiting for the same commit.
func unique(lanes []string) []string {
	seen := make(map[string]bool, len(lanes))
	out := lanes[:0]
	for _, h := range lanes {
		if seen[h] {
			continue
		}
		seen[h] = true
		out = append(out, h)
	}
	return out
}
