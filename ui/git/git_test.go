package git

import (
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// ── FileStatus: the letter, the word, the severity and the ink ──────────────
//
// These are asserted one status at a time and value at a time rather than
// through a loop over a list of the package's own names. A loop over
// FileStatus's own constants would pass if the whole vocabulary shifted by
// one, which is exactly the change that would go unnoticed: a "Renamed" that
// started saying "R" for something else, or a Deletion that stopped being
// the danger ink and became the warning one.

func TestStatusLetterIsGitsOwnLetter(t *testing.T) {
	want := map[FileStatus]string{
		Clean:      "",
		Modified:   "M",
		Added:      "A",
		Deleted:    "D",
		Renamed:    "R",
		Untracked:  "??",
		Conflicted: "U",
		Typechange: "T",
	}
	for status, letter := range want {
		if got := StatusLetter(status); got != letter {
			t.Errorf("StatusLetter(%v) = %q, want %q", status, got, letter)
		}
	}
}

func TestStatusWordIsSpelledOut(t *testing.T) {
	want := map[FileStatus]string{
		Clean:      "Clean",
		Modified:   "Modified",
		Added:      "Added",
		Deleted:    "Deleted",
		Renamed:    "Renamed",
		Untracked:  "Untracked",
		Conflicted: "Conflicted",
		Typechange: "Typechange",
	}
	for status, word := range want {
		if got := StatusWord(status); got != word {
			t.Errorf("StatusWord(%v) = %q, want %q", status, got, word)
		}
		if got := status.String(); got != word {
			t.Errorf("FileStatus(%d).String() = %q, want %q", int(status), got, word)
		}
	}
}

func TestStatusSeverityIsTheRampAndNotAChoice(t *testing.T) {
	want := map[FileStatus]core.Severity{
		Clean:      core.Neutral,
		Modified:   core.Warning,
		Added:      core.Success,
		Deleted:    core.Danger,
		Renamed:    core.Accent,
		Untracked:  core.Accent,
		Conflicted: core.Danger,
		Typechange: core.Warning,
	}
	for status, sev := range want {
		if got := StatusSeverity(status); got != sev {
			t.Errorf("StatusSeverity(%v) = %v, want %v", status, got, sev)
		}
	}
}

func TestStatusInkIsTheSeverityInkInBothAppearances(t *testing.T) {
	light, dark := theme.Light(), theme.Dark()
	cases := []struct {
		status   FileStatus
		lightHex string
		darkHex  string
	}{
		{Clean, "#18181b", "#f4f4f5"},      // Neutral's foreground is body ink
		{Modified, "#9a5410", "#e5a04a"},   // Warning
		{Added, "#1f7a3d", "#5fc47c"},      // Success
		{Deleted, "#c62b30", "#ff6b6f"},    // Danger
		{Renamed, "#1d4ed8", "#9dc0ff"},    // Accent's foreground is AccentText
		{Untracked, "#1d4ed8", "#9dc0ff"},  // Accent
		{Conflicted, "#c62b30", "#ff6b6f"}, // Danger: the loudest there is
		{Typechange, "#9a5410", "#e5a04a"}, // Warning
	}
	for _, tc := range cases {
		if got := StatusInk(tc.status, light); hex(got) != tc.lightHex {
			t.Errorf("StatusInk(%v, light) = %s, want %s", tc.status, hex(got), tc.lightHex)
		}
		if got := StatusInk(tc.status, dark); hex(got) != tc.darkHex {
			t.Errorf("StatusInk(%v, dark) = %s, want %s", tc.status, hex(got), tc.darkHex)
		}
	}
}

func TestStatusPairGivesCleanAPlainSurface(t *testing.T) {
	k := theme.Light()
	bg, fg := StatusPair(Clean, k)
	if hex(bg) != "#f4f4f5" {
		t.Errorf("Clean's background = %s, want Surface #f4f4f5", hex(bg))
	}
	if fg != k.TextMuted {
		t.Errorf("Clean's ink = %s, want the palette's TextMuted %s", hex(fg), hex(k.TextMuted))
	}
	// Every other status takes the pair straight out of Severity, so a pill
	// and a letter cannot disagree about what a Deletion looks like.
	bg, fg = StatusPair(Deleted, k)
	if hex(bg) != "#fdecec" || hex(fg) != "#c62b30" {
		t.Errorf("Deleted's pair = %s/%s, want #fdecec/#c62b30", hex(bg), hex(fg))
	}
}

func TestAnUnknownStatusPanicsRatherThanDrawingNothing(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("an undefined FileStatus should panic, not draw an empty letter")
		}
		if msg, _ := r.(string); !strings.Contains(msg, "git: unknown FileStatus") {
			t.Errorf("panic = %v, want it to say which package and what", r)
		}
	}()
	StatusLetter(FileStatus(99))
}

// ── SplitDiff: the numbers, asserted one line at a time ─────────────────────

// realDiff is a unified diff of the shape git writes: two hunks, a file
// header, an added line, a replaced pair of lines and the note about a
// missing trailing newline. It is pasted in rather than generated so that
// the test is checking the parser against git's output and not against
// whatever this package happens to produce.
const realDiff = "diff --git a/ui/core/core.go b/ui/core/core.go\n" +
	"index 83db48f..bf269f4 100644\n" +
	"--- a/ui/core/core.go\n" +
	"+++ b/ui/core/core.go\n" +
	"@@ -1,6 +1,7 @@\n" +
	" package core\n" +
	" \n" +
	" import \"github.com/egoist/mygo/ui\"\n" +
	" \n" +
	"+// Use applies s to the window.\n" +
	" func Use(c *ui.Context, s Settings) {\n" +
	" }\n" +
	"@@ -20,4 +21,4 @@ func Use(c *ui.Context, s Settings) {\n" +
	" }\n" +
	" \n" +
	"-// old comment\n" +
	"+// new comment\n" +
	" // trailing\n" +
	"\\ No newline at end of file\n"

func TestSplitDiffCutsEveryHunkLine(t *testing.T) {
	lines := SplitDiff(realDiff)
	if len(lines) != 12 {
		t.Fatalf("SplitDiff returned %d lines, want 12:\n%s", len(lines), dump(lines))
	}

	want := []DiffLine{
		{Op: OpContext, OldNo: 1, NewNo: 1, Text: "package core"},
		{Op: OpContext, OldNo: 2, NewNo: 2, Text: ""},
		{Op: OpContext, OldNo: 3, NewNo: 3, Text: `import "github.com/egoist/mygo/ui"`},
		{Op: OpContext, OldNo: 4, NewNo: 4, Text: ""},
		{Op: OpAdd, OldNo: 0, NewNo: 5, Text: "// Use applies s to the window."},
		{Op: OpContext, OldNo: 5, NewNo: 6, Text: "func Use(c *ui.Context, s Settings) {"},
		{Op: OpContext, OldNo: 6, NewNo: 7, Text: "}"},
		{Op: OpContext, OldNo: 20, NewNo: 21, Text: "}"},
		{Op: OpContext, OldNo: 21, NewNo: 22, Text: ""},
		{Op: OpDel, OldNo: 22, NewNo: 0, Text: "// old comment"},
		{Op: OpAdd, OldNo: 0, NewNo: 23, Text: "// new comment"},
		{Op: OpContext, OldNo: 23, NewNo: 24, Text: "// trailing"},
	}
	for i, w := range want {
		got := lines[i]
		if got != w {
			t.Errorf("line %d = %+v\nwant %+v", i, got, w)
		}
	}
}

func TestSplitDiffGivesAnAddedLineNoOldNumber(t *testing.T) {
	// The number a line does not have is 0 and not a copy of the other
	// side's: "line 5" on the right with nothing on the left is an
	// addition, and giving it 5 on both sides is how a viewer ends up with
	// two columns of numbers that mean different things.
	for i, l := range SplitDiff(realDiff) {
		switch l.Op {
		case OpAdd:
			if l.OldNo != 0 {
				t.Errorf("line %d is an addition but has old number %d", i, l.OldNo)
			}
			if l.NewNo <= 0 {
				t.Errorf("line %d is an addition with no new number", i)
			}
		case OpDel:
			if l.NewNo != 0 {
				t.Errorf("line %d is a deletion but has new number %d", i, l.NewNo)
			}
			if l.OldNo <= 0 {
				t.Errorf("line %d is a deletion with no old number", i)
			}
		}
	}
}

func TestHunksKeepsTheHeadersAndTheCounts(t *testing.T) {
	hunks := Hunks(realDiff)
	if len(hunks) != 2 {
		t.Fatalf("Hunks returned %d hunks, want 2", len(hunks))
	}
	if hunks[0].OldFrom != 1 || hunks[0].NewFrom != 1 {
		t.Errorf("hunk 1 starts at %d/%d, want 1/1", hunks[0].OldFrom, hunks[0].NewFrom)
	}
	if hunks[0].OldCount != 6 || hunks[0].NewCount != 7 {
		t.Errorf("hunk 1 counts %d/%d, want 6/7", hunks[0].OldCount, hunks[0].NewCount)
	}
	if hunks[0].Header != "@@ -1,6 +1,7 @@" {
		t.Errorf("hunk 1 header = %q", hunks[0].Header)
	}
	if hunks[1].OldFrom != 20 || hunks[1].NewFrom != 21 {
		t.Errorf("hunk 2 starts at %d/%d, want 20/21", hunks[1].OldFrom, hunks[1].NewFrom)
	}
	if !strings.Contains(hunks[1].Header, "func Use") {
		t.Errorf("hunk 2 header lost the function name git put after the @@: %q", hunks[1].Header)
	}
	// The two hunks' lines are exactly SplitDiff's, in the same order: the
	// one is the other flattened, so the two cannot disagree.
	total := 0
	for _, h := range hunks {
		total += len(h.Lines)
	}
	if total != len(SplitDiff(realDiff)) {
		t.Errorf("the hunks hold %d lines and SplitDiff %d", total, len(SplitDiff(realDiff)))
	}
}

func TestSplitDiffSkipsTheNoNewlineNote(t *testing.T) {
	// The note is about the line above it and is not a line of either
	// file. Drawing it would put a line on screen that the editor cannot
	// find, and consuming a number for it would push every number under it
	// out by one.
	for i, l := range SplitDiff(realDiff) {
		if strings.Contains(l.Text, "No newline") {
			t.Errorf("line %d is the no-newline note: %+v", i, l)
		}
	}
}

func TestSplitDiffReadsAnOmittedCountAsOne(t *testing.T) {
	// "@@ -1 +1 @@" is what git writes for a one-line hunk, and a parser
	// that insists on the comma silently drops it.
	lines := SplitDiff("@@ -8 +8 @@\n-old\n+new\n context\n")
	want := []DiffLine{
		{Op: OpDel, OldNo: 8, NewNo: 0, Text: "old"},
		{Op: OpAdd, OldNo: 0, NewNo: 8, Text: "new"},
		{Op: OpContext, OldNo: 9, NewNo: 9, Text: "context"},
	}
	if len(lines) != len(want) {
		t.Fatalf("got %d lines, want %d:\n%s", len(lines), len(want), dump(lines))
	}
	for i := range want {
		if lines[i] != want[i] {
			t.Errorf("line %d = %+v, want %+v", i, lines[i], want[i])
		}
	}
}

func TestSplitDiffSkipsAHeaderItCannotRead(t *testing.T) {
	// The input is git's output, not an argument a caller got wrong, so an
	// unexpected header is skipped rather than being a crash. The lines
	// before the next header it can read are simply not shown.
	lines := SplitDiff("@@ nonsense @@\n lost\n@@ -2,1 +2,1 @@\n-kept\n+also kept\n")
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2:\n%s", len(lines), dump(lines))
	}
	if lines[0].Text != "kept" || lines[1].Text != "also kept" {
		t.Errorf("lines = %+v", lines)
	}
}

func TestSplitDiffOnSomethingWithNoHunksIsNil(t *testing.T) {
	if got := SplitDiff("just some text\nwith no hunks in it"); got != nil {
		t.Errorf("SplitDiff on a patch with no hunks = %+v, want nil", got)
	}
	if got := SplitDiff(""); got != nil {
		t.Errorf("SplitDiff(\"\") = %+v, want nil", got)
	}
}

func TestCountsAndStatWord(t *testing.T) {
	added, deleted := Counts(SplitDiff(realDiff))
	if added != 2 || deleted != 1 {
		t.Errorf("Counts = +%d -%d, want +2 -1", added, deleted)
	}
	if got := StatWord(2, 1); got != "+2 -1" {
		t.Errorf("StatWord(2,1) = %q", got)
	}
	// A zero on either side is left out: "+0 -3" says "changed" about a
	// file where only lines went.
	if got := StatWord(0, 3); got != "-3" {
		t.Errorf("StatWord(0,3) = %q, want \"-3\"", got)
	}
	if got := StatWord(3, 0); got != "+3" {
		t.Errorf("StatWord(3,0) = %q, want \"+3\"", got)
	}
	if got := StatWord(0, 0); got != "no changes" {
		t.Errorf("StatWord(0,0) = %q, want \"no changes\"", got)
	}
	if got := StatWordOf(realDiff); got != "+2 -1" {
		t.Errorf("StatWordOf = %q, want \"+2 -1\"", got)
	}
}

// ── ANSI ────────────────────────────────────────────────────────────────────

func TestStripANSI(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"a colour", "\x1b[31mred\x1b[0m", "red"},
		{"several parameters", "\x1b[1;32;40mbold green\x1b[m", "bold green"},
		{"nothing to strip", "plain text", "plain text"},
		{"the empty string", "", ""},
		{"only an escape", "\x1b[0m", ""},
		{"a cursor move", "a\x1b[2Kb", "ab"},
		{"an erase", "before\x1b[Kafter", "beforeafter"},
		// The title inside an OSC is part of the sequence, not text: the
		// sequence is telling a terminal what to call the window.
		{"a window title", "\x1b]0;a title\x07the text", "the text"},
		{"a title ended by ST", "\x1b]0;a title\x1b\\text", "text"},
		{"a character set", "x\x1b(B y", "x y"},
		{"a two-character escape", "x\x1b(B y", "x y"},
		{"a device control string", "\x1bP1$r0m\x1b\\after", "after"},
		{"a lone ESC at the end", "text\x1b", "text"},
		{"a sequence that never ends", "text\x1b[31", "text"},
		{"a multi-byte character after a colour", "\x1b[36m蓝色\x1b[0m", "蓝色"},
	}
	for _, tc := range cases {
		if got := StripANSI(tc.in); got != tc.want {
			t.Errorf("%s: StripANSI(%q) = %q, want %q", tc.name, tc.in, got, tc.want)
		}
	}
}

func TestStripANSILeavesPlainTextByteForByte(t *testing.T) {
	// The common case has to come back the same string it went in as, with
	// no copy made: a diff with no colour in it is the common diff.
	in := "diff --git a/main.go b/main.go\n@@ -1 +1 @@\n"
	if got := StripANSI(in); got != in {
		t.Errorf("StripANSI changed text with no escapes in it")
	}
}

func TestParseANSICutsRunsAndKeepsTheCodes(t *testing.T) {
	spans := ParseANSI("\x1b[31mred\x1b[0m plain \x1b[1;32mbold green\x1b[0m")
	want := []struct {
		text  string
		codes []int
	}{
		{"red", []int{31}},
		{" plain ", []int{0}},
		{"bold green", []int{1, 32}},
	}
	if len(spans) != len(want) {
		t.Fatalf("got %d spans, want %d: %+v", len(spans), len(want), spans)
	}
	for i, w := range want {
		if spans[i].Text != w.text {
			t.Errorf("span %d text = %q, want %q", i, spans[i].Text, w.text)
		}
		if !sameInts(spans[i].Codes, w.codes) {
			t.Errorf("span %d codes = %v, want %v", i, spans[i].Codes, w.codes)
		}
	}
}

func TestParseANSIOnTextWithNoEscapesIsOneSpan(t *testing.T) {
	spans := ParseANSI("nothing coloured here")
	if len(spans) != 1 || spans[0].Text != "nothing coloured here" || spans[0].Codes != nil {
		t.Errorf("spans = %+v, want one uncoloured run", spans)
	}
	// Even for nothing at all there is a run, because a component drawing
	// spans needs something to draw before the first escape too.
	if spans := ParseANSI(""); len(spans) != 1 {
		t.Errorf("ParseANSI(\"\") = %+v, want one empty run", spans)
	}
}

func TestParseANSIIgnoresSequencesThatAreNotColours(t *testing.T) {
	// A cursor move changes the screen rather than a character's colour,
	// and a component that drew one would have to be a terminal emulator.
	spans := ParseANSI("a\x1b[2Jb\x1b[?25lc")
	var joined strings.Builder
	for _, sp := range spans {
		joined.WriteString(sp.Text)
	}
	if joined.String() != "abc" {
		t.Errorf("the runs join back to %q, want \"abc\": %+v", joined.String(), spans)
	}
	for i, sp := range spans {
		if strings.ContainsRune(sp.Text, rune(esc)) {
			t.Errorf("run %d still holds an escape byte: %q", i, sp.Text)
		}
	}
}

func TestAnsiInkMapsTheSixteenColoursOntoTheRamp(t *testing.T) {
	light, dark := theme.Light(), theme.Dark()
	cases := []struct {
		code     int
		lightHex string
		darkHex  string
		bold     bool
	}{
		{31, "#c62b30", "#ff6b6f", false}, // red: something is wrong
		{91, "#c62b30", "#ff6b6f", false}, // bright red, the same meaning
		{32, "#1f7a3d", "#5fc47c", false}, // green
		{33, "#9a5410", "#e5a04a", false}, // yellow: needs attention
		{34, "#2563eb", "#5b8dff", false}, // blue
		{37, "#18181b", "#f4f4f5", false}, // white: body ink
		{30, "#18181b", "#f4f4f5", false}, // black: body ink
		{1, "#18181b", "#f4f4f5", true},   // bold, no colour of its own
	}
	for _, tc := range cases {
		ink, bold, ok := AnsiInk(tc.code, light)
		if !ok {
			t.Errorf("AnsiInk(%d) reported nothing about colour", tc.code)
			continue
		}
		if hex(ink) != tc.lightHex || bold != tc.bold {
			t.Errorf("AnsiInk(%d, light) = %s bold=%v, want %s bold=%v",
				tc.code, hex(ink), bold, tc.lightHex, tc.bold)
		}
		if ink, _, ok := AnsiInk(tc.code, dark); !ok || hex(ink) != tc.darkHex {
			t.Errorf("AnsiInk(%d, dark) = %s ok=%v, want %s", tc.code, hex(ink), ok, tc.darkHex)
		}
	}
}

func TestAnsiInkSaysNoAboutCodesThatAreNotColour(t *testing.T) {
	if _, _, ok := AnsiInk(5, theme.Light()); ok {
		t.Error("a blink code should report nothing about ink")
	}
}

func TestValidUTF8IsTheQuestionBothPreviewsAsk(t *testing.T) {
	if !ValidUTF8("héllo 世界") {
		t.Error("valid text was rejected")
	}
	if ValidUTF8(string([]byte{0xff, 0xfe, 0x00})) {
		t.Error("bytes that are not text were accepted")
	}
}

// ── the commit graph ────────────────────────────────────────────────────────

// The history below is the awkward one: a root, a line of two commits off it,
// a side branch that also comes off the root, and a merge that takes both.
// Drawn it is

//	c1 ── c2 ── c4
//	 \          /
//	 └──── s1 ───
//
// which is exactly what a merge looks like and exactly what is wrong if the
// lanes are worked out forwards.
var mergeHistory = []Commit{
	{Hash: "c1", Subject: "root"},
	{Hash: "c2", Parents: []string{"c1"}},
	{Hash: "s1", Parents: []string{"c1"}},
	{Hash: "c3", Parents: []string{"c2"}},
	{Hash: "c4", Parents: []string{"c3", "s1"}},
}

func TestCommitsForPutsTheMergeWhereItBelongs(t *testing.T) {
	rows := CommitsFor(mergeHistory, 0)
	if len(rows) != 5 {
		t.Fatalf("got %d rows, want 5", len(rows))
	}

	// Same order in, same order out.
	for i, row := range rows {
		if row.Commit.Hash != mergeHistory[i].Hash {
			t.Fatalf("row %d is %s, want %s", i, row.Commit.Hash, mergeHistory[i].Hash)
		}
	}

	// The root is in lane 0 with nothing below it: it is the bottom of the
	// drawing and there is nothing under the bottom.
	if rows[0].Column != 0 {
		t.Errorf("the root is in column %d, want 0", rows[0].Column)
	}
	if len(rows[0].Lanes) != 0 {
		t.Errorf("the root has lanes below it: %v", rows[0].Lanes)
	}

	// The side branch opens the second lane and the merge closes it again:
	// five commits, two lanes at the widest, and no third.
	widest := 0
	for _, r := range rows {
		widest = max(widest, r.Width)
	}
	if widest != 2 {
		t.Errorf("the graph is %d lanes wide, want 2:\n%s", widest, dumpGraph(rows))
	}

	// The merge is in lane 0 — the line it merged into — and is the only
	// row that is a merge.
	if !rows[4].Merge {
		t.Error("the last row is a merge and did not say so")
	}
	if rows[4].Column != 0 {
		t.Errorf("the merge is in column %d, want 0 (the line it merged into)", rows[4].Column)
	}
	for i, r := range rows[:4] {
		if r.Merge {
			t.Errorf("row %d (%s) claims to be a merge", i, r.Commit.Hash)
		}
	}
}

func TestCommitsForClosesTheLaneAMergeEnds(t *testing.T) {
	// c4's second parent is s1, which already has a lane of its own. A
	// merge that added a second one would draw the same line twice and push
	// every lane after it a column out, so the widest row stays at two.
	rows := CommitsFor(mergeHistory, 0)
	last := rows[len(rows)-1]
	if len(last.Lanes) != 2 {
		t.Errorf("the merge's row has %d lanes below it (%v), want 2",
			len(last.Lanes), last.Lanes)
	}
}

func TestCommitsForOnALineIsAllOneLane(t *testing.T) {
	line := []Commit{
		{Hash: "a"}, {Hash: "b", Parents: []string{"a"}},
		{Hash: "c", Parents: []string{"b"}}, {Hash: "d", Parents: []string{"c"}},
	}
	rows := CommitsFor(line, 0)
	if len(rows) != 4 {
		t.Fatalf("got %d rows, want 4", len(rows))
	}
	for i, r := range rows {
		if r.Column != 0 {
			t.Errorf("row %d is in column %d; a linear history is one line", i, r.Column)
		}
		if r.Width != 1 {
			t.Errorf("row %d is %d lanes wide, want 1", i, r.Width)
		}
		if r.Merge {
			t.Errorf("row %d claims to be a merge", i)
		}
	}
	// The graph is drawn oldest at the bottom, so every row except the
	// oldest has the one lane running on below it, and the oldest has
	// nothing under it at all.
	for i, r := range rows[1:] {
		if len(r.Lanes) != 1 || r.Lanes[0] != 0 {
			t.Errorf("row %d lanes = %v, want [0]", i+1, r.Lanes)
		}
	}
	if len(rows[0].Lanes) != 0 {
		t.Errorf("the oldest row has lanes below it: %v", rows[0].Lanes)
	}
}

func TestCommitsForATruncatedHistoryOpensItsOwnLane(t *testing.T) {
	// A shallow clone names parents it was not given. Panicking on one would
	// make a graph that cannot be drawn at all out of a list that is merely
	// short, so the commit opens a lane of its own.
	rows := CommitsFor([]Commit{
		{Hash: "c2", Parents: []string{"c1"}},        // c1 was not given
		{Hash: "top", Parents: []string{"c2", "s1"}}, // neither was s1
	}, 0)
	if len(rows) != 2 {
		t.Fatalf("got %d rows, want 2", len(rows))
	}
	if rows[0].Merge {
		t.Error("c2 has one parent and should not claim to be a merge")
	}
	if !rows[1].Merge {
		t.Error("the top commit has two parents and should say so")
	}
	if rows[1].Column != 0 {
		t.Errorf("the merge is in column %d, want 0", rows[1].Column)
	}
}

func TestCommitsForSaysSoWhenALaneDoesNotFit(t *testing.T) {
	rows := CommitsFor(mergeHistory, 1)
	for i, r := range rows {
		if !r.Over {
			t.Errorf("row %d did not report that a lane did not fit in one column", i)
		}
		if r.Width > 1 {
			t.Errorf("row %d is %d lanes wide in a one-column graph", i, r.Width)
		}
	}
}

func TestCommitsForOnNothingIsNil(t *testing.T) {
	if got := CommitsFor(nil, 4); got != nil {
		t.Errorf("CommitsFor(nil) = %+v, want nil", got)
	}
}

func TestCommitShortFallsBackToSevenCharacters(t *testing.T) {
	full := Commit{Hash: "83db48fbf4d1a2c3d4e5f60718293a4b5c6d7e8f"}
	if got := full.short(); got != "83db48f" {
		t.Errorf("short() = %q, want 83db48f", got)
	}
	if got := (Commit{Hash: "abc", Short: "xyz"}).short(); got != "xyz" {
		t.Errorf("an explicit Short should win, got %q", got)
	}
	if got := (Commit{Hash: "abc"}).short(); got != "abc" {
		t.Errorf("a hash shorter than seven should be left alone, got %q", got)
	}
}

// ── the stat bar ────────────────────────────────────────────────────────────

func TestStatSlices(t *testing.T) {
	cases := []struct {
		name              string
		added, deleted    int
		slots             int
		wantAdds, wantDls int
	}{
		{"nothing changed", 0, 0, 8, 0, 0},
		{"only additions fill the bar", 50, 0, 8, 8, 0},
		{"only deletions fill the bar", 0, 50, 8, 0, 8},
		{"evenly split", 4, 4, 8, 4, 4},
		{"mostly deletions", 1, 99, 8, 1, 7},
		{"mostly additions", 99, 1, 8, 7, 1},
		{"one of each in one slot", 1, 1, 1, 1, 0},
		{"no slots at all", 5, 5, 0, 0, 0},
	}
	for _, tc := range cases {
		adds, dels := StatSlices(tc.added, tc.deleted, tc.slots)
		if adds != tc.wantAdds || dels != tc.wantDls {
			t.Errorf("%s: StatSlices(%d,%d,%d) = +%d -%d, want +%d -%d",
				tc.name, tc.added, tc.deleted, tc.slots, adds, dels, tc.wantAdds, tc.wantDls)
		}
		if tc.slots > 0 && adds+dels > tc.slots {
			t.Errorf("%s: %d slots used out of %d", tc.name, adds+dels, tc.slots)
		}
	}
}

func TestStatSlicesNeverHidesADeletion(t *testing.T) {
	// A hundred additions and one deletion drawing a full green bar and no
	// red at all is the one shape a stat bar must never be: the deletion is
	// what the reader is looking for.
	adds, dels := StatSlices(1000, 1, 8)
	if dels < 1 {
		t.Errorf("a 1000:1 diff drew %d deletions, want at least 1", dels)
	}
	if adds+dels != 8 {
		t.Errorf("StatSlices(1000,1,8) used %d of 8 slots", adds+dels)
	}
}

// ── the inline diff ─────────────────────────────────────────────────────────

func TestInlineSplitMarksOnlyTheWordThatChanged(t *testing.T) {
	removed, added := InlineSplit("the quick brown fox", "the slow brown fox")

	if len(removed) != 4 || len(added) != 4 {
		t.Fatalf("got %d removed and %d added, want 4 and 4", len(removed), len(added))
	}
	// One word changed, and the other three are marked as unchanged — the
	// whole of what an inline diff is for.
	var changedOld, changedNew []int
	for _, w := range removed {
		if w.Changed {
			changedOld = append(changedOld, w.No)
		}
	}
	for _, w := range added {
		if w.Changed {
			changedNew = append(changedNew, w.No)
		}
	}
	if !sameInts(changedOld, []int{2}) {
		t.Errorf("changed words on the left = %v, want [2]", changedOld)
	}
	if !sameInts(changedNew, []int{2}) {
		t.Errorf("changed words on the right = %v, want [2]", changedNew)
	}
	if got := strings.TrimSpace(removed[1].Text); got != "quick" {
		t.Errorf("the changed word on the left is %q, want \"quick\"", got)
	}
	if got := strings.TrimSpace(added[1].Text); got != "slow" {
		t.Errorf("the changed word on the right is %q, want \"slow\"", got)
	}
}

func TestInlineSplitPutsBothSidesBackTogether(t *testing.T) {
	// A diff that silently drops a space is a diff of a different string,
	// so joining every word of a side must give that side back exactly.
	for _, before := range []string{"", "one", "one two three", "  leading spaces"} {
		removed, _ := InlineSplit(before, "something else entirely")
		var back strings.Builder
		for _, w := range removed {
			back.WriteString(w.Text)
		}
		if back.String() != before {
			t.Errorf("InlineSplit(%q) puts back %q", before, back.String())
		}
	}
}

func TestInlineSplitOfTheSameTextHasNothingChanged(t *testing.T) {
	removed, added := InlineSplit("exactly the same words here", "exactly the same words here")
	for _, w := range append(removed, added...) {
		if w.Changed {
			t.Errorf("word %d (%q) of identical text was marked changed", w.No, w.Text)
		}
	}
}

func TestInlineSplitOfTwoEmptiesIsEmpty(t *testing.T) {
	removed, added := InlineSplit("", "")
	if len(removed) != 0 || len(added) != 0 {
		t.Errorf("InlineSplit(\"\",\"\") = %d/%d words, want 0/0", len(removed), len(added))
	}
}

// ── helpers ─────────────────────────────────────────────────────────────────

// hex is a colour as the web version of this interface writes it, which is
// how a test here reads a palette value back.
func hex(c ui.Color) string {
	const digits = "0123456789abcdef"
	buf := []byte{'#', 0, 0, 0, 0, 0, 0}
	n := 0
	for _, v := range []uint8{c.R, c.G, c.B} {
		buf[1+n*2] = digits[v>>4]
		buf[2+n*2] = digits[v&0xf]
		n++
	}
	return string(buf)
}

func sameInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// dump is a diff printed line by line, for a failure message.
func dump(lines []DiffLine) string {
	var b strings.Builder
	for i, l := range lines {
		b.WriteString("  ")
		b.WriteString(itoa(i))
		b.WriteString(": ")
		b.WriteString(l.Op.String())
		b.WriteString(" old=")
		b.WriteString(itoa(l.OldNo))
		b.WriteString(" new=")
		b.WriteString(itoa(l.NewNo))
		b.WriteString(" ")
		b.WriteString(l.Text)
		b.WriteString("\n")
	}
	return b.String()
}

func dumpGraph(rows []GraphRow) string {
	var b strings.Builder
	for i, r := range rows {
		b.WriteString("  ")
		b.WriteString(itoa(i))
		b.WriteString(": ")
		b.WriteString(r.Commit.Hash)
		b.WriteString(" col=")
		b.WriteString(itoa(r.Column))
		b.WriteString(" width=")
		b.WriteString(itoa(r.Width))
		b.WriteString(" lanes=")
		b.WriteString(itoa(len(r.Lanes)))
		if r.Merge {
			b.WriteString(" merge")
		}
		b.WriteString("\n")
	}
	return b.String()
}

// ── the components, rendered ────────────────────────────────────────────────
//
// Every component here gets at least one headless render, because a component
// that has never been drawn is a component whose first exercise is somebody
// else's window. The assertions are on what a person would read off the
// screen — the visible words and whether a thing is findable — and never on
// internal fields.

var demoBranches = []Branch{
	{Name: "main", Current: true, Subject: "Ship the composer"},
	{Name: "fix/crash-on-empty", Subject: "Don't panic on an empty body", Ahead: 2, Behind: 3},
	{Name: "origin/main", Remote: true, Behind: 3},
}

var demoCommits = []Commit{
	{
		Hash: "c4f2a91", Subject: "Fix the crash on an empty body", Author: "Ada Lovelace",
		When: "2 hours ago", Parents: []string{"c3", "s1"}, Refs: []string{"fix/crash"},
	},
	{Hash: "c3b8e07", Subject: "Add a test for the empty case", Author: "Ada Lovelace", When: "3 hours ago", Parents: []string{"c2"}},
	{Hash: "s1a2c44", Subject: "Bump the timeout", Author: "Grace Hopper", When: "yesterday", Parents: []string{"c1"}},
	{Hash: "c2d7715", Subject: "Rename the field", Author: "Ada Lovelace", When: "2 days ago", Parents: []string{"c1"}},
	{Hash: "c100000", Subject: "First commit", Author: "Grace Hopper", When: "3 weeks ago"},
}

var demoChanges = []Change{
	{Path: "ui/git/diff.go", Status: Modified, Added: 42, Deleted: 7},
	{Path: "ui/git/status.go", Status: Added, Added: 88, Staged: true},
	{Path: "ui/core/core.go", Status: Deleted, Deleted: 120},
	{Path: "notes.txt", Status: Untracked, Added: 3},
}

func TestBranchListDrawsTheBranchesItWasGiven(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		sel := 0
		BranchList(c, &sel, demoBranches, BranchListOptions{Height: 300})
	}, 520, 260)
	for _, want := range []string{"main", "fix/crash-on-empty", "Ship the composer", "↑2", "↓3"} {
		if !tt.HasText(want) {
			t.Errorf("branch list is missing %q; it drew %q", want, tt.Texts())
		}
	}
	// The remote branch is left out by default: a list that shows both is
	// twice as long, and a checkout menu is not a fetch log.
	if tt.HasText("origin/main") {
		t.Error("the remote branch should be hidden unless ShowRemote is set")
	}
}

func TestBranchListShowsTheRemotesWhenAsked(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		sel := 0
		BranchList(c, &sel, demoBranches, BranchListOptions{Height: 300, ShowRemote: true})
	}, 520, 300)
	if !tt.HasText("origin/main") {
		t.Errorf("ShowRemote should have drawn the remote branch; it drew %q", tt.Texts())
	}
}

func TestBranchListSaysWhenThereIsNothing(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		sel := 0
		BranchList(c, &sel, nil, BranchListOptions{Height: 200})
	}, 400, 160)
	if !tt.HasText("No branches") {
		t.Errorf("an empty branch list should say so; it drew %q", tt.Texts())
	}
}

func TestBranchListPicksAcrossBuilds(t *testing.T) {
	// MyGo builds a frame up to three times so that the result of an event
	// shows up in the frame it happened in, and settle() runs the passes
	// out. The consequence for a test is that Clicked() is false in the
	// last pass, so a result read only there is always -1. The result is
	// therefore accumulated across passes in the view closure, which is
	// what this does — and which is why the assertion is on the pointer the
	// caller owns rather than on text drawn afterwards.
	var picked int = -1
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		sel := 0
		if v := BranchList(c, &sel, demoBranches, BranchListOptions{Height: 300}); v.Picked() >= 0 {
			picked = v.Picked()
		}
	}, 520, 260)
	if err := tt.Click("fix/crash-on-empty"); err != nil {
		t.Fatal(err)
	}
	if picked != 1 {
		t.Errorf("Picked() = %d, want 1 (the second branch)", picked)
	}
}

func TestBranchListNeedsASelectionAndAHeight(t *testing.T) {
	mustPanic(t, "git: BranchList needs a selection", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			BranchList(c, nil, demoBranches, BranchListOptions{Height: 200})
		}, 400, 200)
	})
	mustPanic(t, "git: BranchList needs a Height", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			sel := 0
			BranchList(c, &sel, demoBranches, BranchListOptions{})
		}, 400, 200)
	})
}

func TestBranchSelectorDrawsTheCurrentBranchAndItsPanel(t *testing.T) {
	open := false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		sel := 0
		BranchSelector(c, &sel, &open, demoBranches, BranchSelectorOptions{
			Label: "Branch", Height: 200,
		})
	}, 560, 360)
	if !tt.HasText("main") {
		t.Errorf("the trigger should show the current branch; it drew %q", tt.Texts())
	}
	if err := tt.Click("Branch"); err != nil {
		t.Fatal(err)
	}
	if !open {
		t.Fatal("pressing the trigger did not write true into the caller's flag")
	}
	for _, want := range []string{"fix/crash-on-empty", "origin/main"} {
		if !tt.HasText(want) {
			t.Errorf("the panel is missing %q; it drew %q", want, tt.Texts())
		}
	}
}

func TestBranchSelectorInsistsOnALabel(t *testing.T) {
	mustPanic(t, "git: BranchSelector needs a Label", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			sel, open := 0, false
			BranchSelector(c, &sel, &open, demoBranches, BranchSelectorOptions{})
		}, 400, 200)
	})
	mustPanic(t, "git: BranchSelector selection", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			sel, open := 7, false
			BranchSelector(c, &sel, &open, demoBranches, BranchSelectorOptions{Label: "Branch"})
		}, 400, 200)
	})
}

func TestCommitListDrawsTheLog(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		CommitList(c, demoCommits, CommitListOptions{Height: 320, ShowRefs: true})
	}, 760, 360)
	for _, want := range []string{
		"Fix the crash on an empty body", "c4f2a91", "Ada Lovelace", "2 hours ago",
		"fix/crash", "First commit",
	} {
		if !tt.HasText(want) {
			t.Errorf("the log is missing %q; it drew %q", want, tt.Texts())
		}
	}
}

func TestCommitListSaysWhenThereAreNoCommits(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		CommitList(c, nil, CommitListOptions{Height: 200})
	}, 500, 200)
	if !tt.HasText("No commits") {
		t.Errorf("an empty log should say so; it drew %q", tt.Texts())
	}
}

func TestCommitGraphDrawsTheSubjectsAndKeepsTheLayout(t *testing.T) {
	var rows []GraphRow
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		res := CommitGraph(c, demoCommits, CommitGraphOptions{ShowRefs: true})
		rows = res.Rows()
	}, 700, 320)
	if !tt.HasText("Fix the crash on an empty body") {
		t.Errorf("the graph is missing its newest commit; it drew %q", tt.Texts())
	}
	if !tt.HasText("First commit") {
		t.Errorf("the graph is missing its oldest commit; it drew %q", tt.Texts())
	}
	if len(rows) != len(demoCommits) {
		t.Fatalf("Rows() gave %d rows for %d commits", len(rows), len(demoCommits))
	}
	if !rows[0].Merge {
		t.Error("the newest commit has two parents and should be drawn as a merge")
	}
}

func TestCommitInputDrawsAndRefusesAnEmptyMessage(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		msg := "  "
		var amend bool
		CommitInput(c, &msg, CommitInputOptions{
			Label: "Commit message", Staged: 3, Amend: &amend, Height: 90,
		})
	}, 620, 320)
	if !tt.HasText("Nothing staged") && !tt.HasText("Staged: 3") {
		t.Errorf("the staged count is missing; it drew %q", tt.Texts())
	}
	if !tt.HasText("A commit needs a message") {
		t.Errorf("a blank message should say why; it drew %q", tt.Texts())
	}
	if !tt.HasText("Amend last commit") {
		t.Errorf("the amend box is missing; it drew %q", tt.Texts())
	}
}

func TestCommitInputNeedsAMessageToPointAt(t *testing.T) {
	mustPanic(t, "git: CommitInput needs a message", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			CommitInput(c, nil, CommitInputOptions{Label: "Commit message"})
		}, 400, 200)
	})
}

func TestChangesListDrawsEveryPathAndStatus(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		sel := 0
		ChangesList(c, demoChanges, ChangesListOptions{Height: 320, WithWord: true})
		_ = sel
	}, 720, 340)
	for _, want := range []string{
		"ui/git/diff.go", "Modified", "ui/git/status.go", "Added",
		"ui/core/core.go", "Deleted", "notes.txt", "Untracked", "+42 -7",
	} {
		if !tt.HasText(want) {
			t.Errorf("the change list is missing %q; it drew %q", want, tt.Texts())
		}
	}
}

func TestGitStatusBadgeWearsTheSevennity(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ui.Column(c).Gap(4).Children(func() {
			GitStatusBadge(c, Modified, GitStatusBadgeOptions{})
			GitStatusBadge(c, Conflicted, GitStatusBadgeOptions{WithWord: true})
		})
	}, 400, 160)
	// The badge is named by the word whatever it shows, because two glyphs
	// are not a description of anything.
	if _, ok := tt.Find("Conflicted"); !ok {
		t.Errorf("the conflicted badge should be named for assistive technology; it drew %q", tt.Texts())
	}
}

func TestDiffViewerDrawsTheLinesWithTheirNumbers(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		DiffViewer(c, realDiff, DiffViewerOptions{
			Height: 340, File: "ui/core/core.go", WithHunks: true,
		})
	}, 760, 400)
	for _, want := range []string{
		"ui/core/core.go", "package core", "// Use applies s to the window.",
		"// new comment", "@@ -1,6 +1,7 @@",
	} {
		if !tt.HasText(want) {
			t.Errorf("the diff is missing %q; it drew %q", want, tt.Texts())
		}
	}
	// A deleted line is on screen too — a viewer that only drew the new side
	// would be a reviewer with no way to see what was taken away.
	if !tt.HasText("// old comment") {
		t.Errorf("the removed line is missing; the viewer drew %q", tt.Texts())
	}
}

func TestDiffViewerNeedsAHeight(t *testing.T) {
	mustPanic(t, "git: DiffViewer needs a Height", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			DiffViewer(c, realDiff, DiffViewerOptions{})
		}, 600, 300)
	})
}

func TestDiffViewerInDarkMode(t *testing.T) {
	// The text draws the same in both appearances, so what has to be checked
	// is that the window resolved the dark palette and drew in it.
	var resolved theme.Tokens
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: core.Dark})
		resolved = core.Tokens(c)
		DiffViewer(c, realDiff, DiffViewerOptions{Height: 300, File: "core.go"})
	}, 720, 360)
	tt.SetDark(true)
	tt.Frame()
	for _, want := range []string{"core.go", "package core", "// new comment"} {
		if !tt.HasText(want) {
			t.Errorf("the dark diff is missing %q; it drew %q", want, tt.Texts())
		}
	}
	if resolved.Background != theme.Dark().Background || !resolved.IsDark() {
		t.Errorf("the window resolved %s, want the dark palette's %s",
			hex(resolved.Background), hex(theme.Dark().Background))
	}
	// An added line wears SuccessBg, which is a different colour in the two
	// appearances — so the same diff is not the same picture twice.
	if resolved.SuccessBg == theme.Light().SuccessBg {
		t.Error("the dark diff is tinting its additions with the light palette")
	}
}

func TestThreeWayMergeDrawsThreeColumns(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ThreeWayMerge(c, ThreeWayMergeOptions{
			Height: 300,
			Base:   []string{"func main() {", "\tprintln(1)", "}"},
			Ours:   []string{"func main() {", "\tprintln(2)", "}"},
			Theirs: []string{"func main() {", "\tprintln(3)", "\tprintln(4)", "}"},
			Choices: []Resolution{
				KeepOurs, Unresolved, Unresolved,
			},
		})
	}, 820, 360)
	for _, want := range []string{"Base", "Ours", "Theirs", "println(2)", "println(4)"} {
		if !tt.HasText(want) {
			t.Errorf("the merge is missing %q; it drew %q", want, tt.Texts())
		}
	}
}

func TestConflictResolverDrawsBothSidesAndTheFourWaysOut(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		choices := []Resolution{Unresolved}
		ConflictResolver(c, []Conflict{{
			Label:  "ui/git/diff.go",
			Ours:   []string{"keep the old call"},
			Theirs: []string{"keep the new call"},
		}}, ConflictResolverOptions{Choices: choices})
	}, 760, 360)
	for _, want := range []string{"ui/git/diff.go", "keep the old call", "keep the new call", "Both", "Neither"} {
		if !tt.HasText(want) {
			t.Errorf("the resolver is missing %q; it drew %q", want, tt.Texts())
		}
	}
}

func TestConflictResolverRefusesToGuessAtTheChoices(t *testing.T) {
	// The choices are the caller's slice, and a resolver that invented one
	// would be a second piece of state beside the file somebody is about to
	// write.
	mustPanic(t, "git: ConflictResolver needs the caller's choices", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			ConflictResolver(c, []Conflict{{Label: "a"}}, ConflictResolverOptions{})
		}, 500, 300)
	})
	mustPanic(t, "git: ConflictResolver has 2 conflicts", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			ConflictResolver(c,
				[]Conflict{{Label: "a"}, {Label: "b"}},
				ConflictResolverOptions{Choices: []Resolution{KeepOurs}})
		}, 500, 300)
	})
}

func TestResolutionApplyIsArithmeticNotPainting(t *testing.T) {
	conf := Conflict{
		Ours:   []string{"a", "b"},
		Theirs: []string{"c", "d"},
	}
	cases := []struct {
		res  Resolution
		want []string
	}{
		{KeepOurs, []string{"a", "b"}},
		{KeepTheirs, []string{"c", "d"}},
		{KeepBoth, []string{"a", "b", "c", "d"}},
		{KeepNeither, nil},
		{Unresolved, []string{"a", "b"}},
	}
	for _, tc := range cases {
		got := tc.res.Apply(conf)
		if !sameStrings(got, tc.want) {
			t.Errorf("%v.Apply = %v, want %v", tc.res, got, tc.want)
		}
	}
	// KeepBoth must not alias the caller's slice: appending to the result
	// must not be able to reach back into the conflict.
	both := KeepBoth.Apply(conf)
	both[0] = "changed"
	if conf.Ours[0] != "a" {
		t.Error("Apply handed back the caller's own slice")
	}
}

func TestBlameViewDrawsAuthorsAndLines(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		BlameView(c, []BlameLine{
			{Text: "package core", Author: "Ada Lovelace", Commit: "c1a2b3c", When: "3 months ago", Date: "2025-01-14"},
			{Text: "", Author: "Grace Hopper", Commit: "d4e5f6a", When: "2 weeks ago", Date: "2025-06-02"},
			{Text: "// Use applies s.", Author: "Ada Lovelace", Commit: "c1a2b3c", When: "3 months ago", Date: "2025-01-14"},
		}, BlameViewOptions{Height: 260, WithDate: true, WithGutter: true})
	}, 760, 320)
	for _, want := range []string{"Ada Lovelace", "Grace Hopper", "c1a2b3c", "package core", "2025-06-02"} {
		if !tt.HasText(want) {
			t.Errorf("blame is missing %q; it drew %q", want, tt.Texts())
		}
	}
}

func TestFileHistoryDrawsTheCommitsOfOneFile(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		sel := 0
		FileHistory(c, "ui/git/diff.go", demoCommits, FileHistoryOptions{
			Height: 300, Selected: &sel,
		})
	}, 700, 340)
	for _, want := range []string{"Fix the crash on an empty body", "Ada Lovelace", "First commit"} {
		if !tt.HasText(want) {
			t.Errorf("the file history is missing %q; it drew %q", want, tt.Texts())
		}
	}
}

func TestFileHistoryNeedsItsPath(t *testing.T) {
	mustPanic(t, "git: FileHistory needs the path", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			FileHistory(c, "", demoCommits, FileHistoryOptions{Height: 200})
		}, 500, 250)
	})
}

func TestPullRequestCardDrawsTheWholeRequest(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		PullRequestCard(c, PullRequest{
			Number: 412, Title: "Fix the composer crash", Author: "Ada Lovelace",
			From: "fix/crash", To: "main", State: "Open",
			ChecksPassed: 20, ChecksTotal: 21, ChecksFailed: 1,
		}, PullRequestCardOptions{})
	}, 760, 280)
	for _, want := range []string{"#412", "Fix the composer crash", "fix/crash → main", "Open", "1 failing", "20/21"} {
		if !tt.HasText(want) {
			t.Errorf("the card is missing %q; it drew %q", want, tt.Texts())
		}
	}
}

func TestPullRequestStateIsTheToneOfTheNews(t *testing.T) {
	cases := map[string]core.Severity{
		"Open":   core.Accent,
		"open":   core.Accent,
		"merged": core.Success,
		"Merged": core.Success,
		"Closed": core.Neutral,
		"draft":  core.Neutral,
		"":       core.Accent,
	}
	for state, want := range cases {
		if got := PullRequestState(state); got != want {
			t.Errorf("PullRequestState(%q) = %v, want %v", state, got, want)
		}
	}
}

func TestReviewCommentDrawsWhereItHangs(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ReviewComment(c, ReviewNote{
			Author: "Grace Hopper", When: "an hour ago",
			Body: "This drops the error case.", Path: "ui/git/diff.go", Line: 42,
			Reactions: []Reaction{{Emoji: "👍", Count: 3, Mine: true}},
		}, ReviewCommentOptions{Reactions: true})
	}, 700, 320)
	for _, want := range []string{"Grace Hopper", "This drops the error case.", "ui/git/diff.go:42", "3 reactions"} {
		if !tt.HasText(want) {
			t.Errorf("the comment is missing %q; it drew %q", want, tt.Texts())
		}
	}
}

func TestStashListDrawsTheEntries(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		StashList(c, []Stash{
			{Name: "stash@{0}", Message: "WIP on the composer", Branch: "fix/crash", When: "2 hours ago"},
			{Name: "stash@{1}", Message: "Debugging the crash", Branch: "main", When: "yesterday"},
		}, StashListOptions{Height: 240})
	}, 700, 280)
	for _, want := range []string{"stash@{0}", "WIP on the composer", "fix/crash", "Debugging the crash"} {
		if !tt.HasText(want) {
			t.Errorf("the stash list is missing %q; it drew %q", want, tt.Texts())
		}
	}
}

func TestTagListTellsAnAnnotatedTagFromALightweightOne(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		TagList(c, []Tag{
			{Name: "v2.1.0", Subject: "The composer rewrite", When: "a week ago", Annotated: true},
			{Name: "latest", Subject: "Whatever is on main", When: "2 hours ago"},
		}, TagListOptions{Height: 220})
	}, 700, 260)
	for _, want := range []string{"v2.1.0", "The composer rewrite", "latest", "Whatever is on main"} {
		if !tt.HasText(want) {
			t.Errorf("the tag list is missing %q; it drew %q", want, tt.Texts())
		}
	}
}

func TestInlineDiffDrawsBothSides(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		InlineDiff(c, "the quick brown fox", "the slow brown fox", InlineDiffOptions{})
	}, 720, 220)
	if !tt.HasText("quick") || !tt.HasText("slow") {
		t.Errorf("the inline diff is missing a side; it drew %q", tt.Texts())
	}
	// The words that did not change are on both sides, because a diff that
	// showed only the change would be two words rather than a diff.
	for _, want := range []string{"the", "brown", "fox"} {
		if !tt.HasText(want) {
			t.Errorf("the inline diff dropped the unchanged word %q; it drew %q", want, tt.Texts())
		}
	}
}

func TestDiffStatDrawsTheBarAndTheFigures(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		DiffStat(c, 4, 1, DiffStatStat())
	}, 420, 120)
	if !tt.HasText("+4 -1") {
		t.Errorf("the stat should print its figures; it drew %q", tt.Texts())
	}
	// The bar is named by the same figures, so a screen reader gets the
	// numbers rather than "group".
	if _, ok := tt.Find("+4 -1, 7 of 8 added"); !ok {
		t.Errorf("the bar should be named with its figures")
	}
}

func TestCommandBlockShowsWhatCameBack(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		CommandBlock(c, CommandBlockOptions{
			Command: "git log --oneline -5",
			Result:  CommandResult{Code: 128, Err: "fatal: not a git repository"},
		})
	}, 700, 280)
	for _, want := range []string{"$ git log --oneline -5", "fatal: not a git repository", "exit 128"} {
		if !tt.HasText(want) {
			t.Errorf("the command block is missing %q; it drew %q", want, tt.Texts())
		}
	}
}

func TestCommandBlockNeedsTheCommandItRan(t *testing.T) {
	mustPanic(t, "git: CommandBlock needs the command", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			CommandBlock(c, CommandBlockOptions{})
		}, 500, 200)
	})
}

func TestAnsiTextDrawsTheRunsWithoutTheEscapes(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		AnsiText(c, "\x1b[31mfatal:\x1b[0m not a git repository", AnsiTextOptions{})
	}, 640, 160)
	if !tt.HasText("fatal:") || !tt.HasText("not a git repository") {
		t.Errorf("the ANSI text lost a run; it drew %q", tt.Texts())
	}
	for _, s := range tt.Texts() {
		if strings.ContainsRune(s, rune(esc)) {
			t.Errorf("an escape byte reached the screen: %q", s)
		}
	}
}

func TestAnsiTextInDarkMode(t *testing.T) {
	// Git's palette is the one thing in this package that arrives coloured
	// from outside, so dark mode is where it has to be re-mapped rather
	// than passed through. The red in this string becomes the window's
	// danger ink, which is a different colour in each appearance.
	var red, green ui.Color
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: core.Dark})
		k := core.Tokens(c)
		red, _, _ = AnsiInk(31, k)
		green, _, _ = AnsiInk(32, k)
		AnsiText(c, "\x1b[31mred\x1b[32mgreen\x1b[0m plain", AnsiTextOptions{})
	}, 520, 160)
	tt.SetDark(true)
	tt.Frame()
	for _, want := range []string{"red", "green", "plain"} {
		if !tt.HasText(want) {
			t.Errorf("the dark ANSI text is missing %q; it drew %q", want, tt.Texts())
		}
	}
	if hex(red) != hex(theme.Dark().Danger) {
		t.Errorf("red mapped to %s in dark, want the dark danger ink %s",
			hex(red), hex(theme.Dark().Danger))
	}
	if hex(green) != hex(theme.Dark().Success) {
		t.Errorf("green mapped to %s in dark, want the dark success ink %s",
			hex(green), hex(theme.Dark().Success))
	}
	if red == theme.Light().Danger {
		t.Error("the dark window is drawing ANSI red in the light palette's ink")
	}
}

// mustPanic runs f and checks that it stopped with a message mentioning want.
// The message is matched rather than the panic type because the message is
// what somebody reads when a component is handed nothing.
func mustPanic(t *testing.T, want string, f func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("expected a panic mentioning %q", want)
			return
		}
		msg, ok := r.(string)
		if !ok {
			t.Errorf("panic value is %T, want a string: %v", r, r)
			return
		}
		if !strings.Contains(msg, want) {
			t.Errorf("panic = %q, want it to mention %q", msg, want)
		}
	}()
	f()
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// DiffStatStat is the option set the stat test draws with, named so the call
// above reads as a drawing rather than as a struct literal in the middle of
// one.
func DiffStatStat() DiffStatOptions { return DiffStatOptions{WithCounts: true} }
