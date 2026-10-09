package agent

import (
	"strconv"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/code"
	"github.com/HycJack/MintUI/ui/core"
)

// What is worth guarding in this package, and how.
//
// Most of it is arithmetic. Three functions in here decide what a run's
// interface says rather than what it looks like, and all three are checked
// item by item: StatusTone, StepNumber and Merge3. A status that comes out
// the wrong colour is a bug a screenshot will not show, because a green word
// and a red word are both legible; a merge that quietly drops one side's edit
// is a bug nobody sees until the wrong file is written.
//
// The rest of the file is that every component draws what it promises in both
// appearances, and that the few that take a press report it. Those tests
// accumulate inside the view closure rather than reading the result after
// settle, because MyGo builds a frame up to three times and the last pass
// arrives with nothing pending.

// dark runs a view in one appearance, so that every component is checked in a
// window a person is actually looking at. A run's transcript is drawn almost
// entirely out of tinted pairs, and a tinted pair only fails in the appearance
// where it is too close to its own background.
func dark(t *testing.T, mode core.Mode, view func(c *ui.Context)) *ui.Tester {
	t.Helper()
	return ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: mode})
		view(c)
	}, 900, 900)
}

// ── the status mapping ──────────────────────────────────────────────────────

// TestStatusTone is the mapping every mark in this package is coloured from,
// checked one status at a time rather than as a table of the whole function —
// because the failure it guards against is one status changed and the rest
// left alone, and a table diff shows that as one line out of six.
func TestStatusTone(t *testing.T) {
	cases := []struct {
		status Status
		want   core.Severity
		why    string
	}{
		{StatusQueued, core.Neutral,
			"queued work is claiming nothing, so it wears ordinary ink"},
		{StatusRunning, core.Accent,
			"running is the one status that is moving, and the accent is the system's own highlight"},
		{StatusWaiting, core.Warning,
			"waiting is the only status that needs a person, so it must be the one that stands out"},
		{StatusDone, core.Success,
			"done is good news, and it must be the same green the board's resolved pill is"},
		{StatusFailed, core.Danger,
			"failed is the status that must never be mistaken for anything else"},
		{StatusCancelled, core.Neutral,
			"cancelled is neutral on purpose: a person who stops a run and scans the page for what " +
				"they just broke must not find red"},
	}
	for _, c := range cases {
		if got := StatusTone(c.status); got != c.want {
			t.Errorf("StatusTone(%s) = %s, want %s — %s",
				c.status, got, c.want, c.why)
		}
	}
}

// TestStatusToneRefusesAStatusThatIsNotThere: an unknown status is a caller
// who forgot a case in their own switch, and falling back to a colour would
// draw it as something it is not rather than stopping.
func TestStatusToneRefusesAStatusThatIsNotThere(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("StatusTone(99) returned; it should stop instead")
		}
	}()
	_ = StatusTone(Status(99))
}

func TestStatusLabel(t *testing.T) {
	want := map[Status]string{
		StatusQueued:    "Queued",
		StatusRunning:   "Running",
		StatusWaiting:   "Waiting",
		StatusDone:      "Done",
		StatusFailed:    "Failed",
		StatusCancelled: "Cancelled",
	}
	for status, text := range want {
		if got := status.String(); got != text {
			t.Errorf("Status(%d).String() = %q, want %q", int(status), got, text)
		}
	}
}

func TestStatusBusyOnlyRuns(t *testing.T) {
	if !StatusBusy(StatusRunning) {
		t.Error("a running status must say it is busy")
	}
	// Waiting is the case worth stopping for: nothing is moving while a run
	// waits for a person, and a mark that moves there says the system is
	// working on it when in fact it is waiting to be answered.
	for _, s := range []Status{StatusQueued, StatusWaiting, StatusDone, StatusFailed, StatusCancelled} {
		if StatusBusy(s) {
			t.Errorf("%s must not say it is busy", s)
		}
	}
}

// TestWorstStatusIsNeverQuieterThanWhatItWraps: a group or a summary drawn
// with a quieter status than the worst thing inside it is a lie the reader
// cannot check.
func TestWorstStatusIsNeverQuieterThanWhatItWraps(t *testing.T) {
	cases := []struct {
		in   []Status
		want Status
	}{
		{[]Status{StatusDone, StatusDone}, StatusDone},
		{[]Status{StatusDone, StatusRunning}, StatusRunning},
		{[]Status{StatusDone, StatusRunning, StatusFailed}, StatusFailed},
		{[]Status{StatusRunning, StatusWaiting}, StatusWaiting},
		{[]Status{StatusCancelled, StatusQueued}, StatusCancelled},
		{[]Status{StatusDone, StatusCancelled}, StatusCancelled},
		{nil, StatusDone},
	}
	for _, c := range cases {
		if got := WorstStatus(c.in); got != c.want {
			t.Errorf("WorstStatus(%v) = %s, want %s", c.in, got, c.want)
		}
	}
}

// ── step numbering ──────────────────────────────────────────────────────────

// TestStepNumber is checked at every depth there is, because the whole of the
// function is that one depth differs from the next: a numbering whose levels
// all looked the same would pass a test that only checked the ends.
func TestStepNumber(t *testing.T) {
	cases := []struct {
		depth int
		want  string
	}{
		{0, "1"},
		{1, "1.1"},
		{2, "1.1.1"},
		{3, "1.1.1.1"},
		{StepMaxDepth, "1.1.1.1.1"},
		{StepMaxDepth + 1, "1.1.1.1.1"},
		{StepMaxDepth + 40, "1.1.1.1.1"},
	}
	for _, c := range cases {
		if got := StepNumber(c.depth); got != c.want {
			t.Errorf("StepNumber(%d) = %q, want %q", c.depth, got, c.want)
		}
	}
	// A negative depth is a transcript whose hierarchy has come apart. It is
	// clamped rather than refused, because drawing the shape that is there is
	// more useful than refusing the run over it — but it must still be a
	// number, and the same one as the floor.
	for _, depth := range []int{-1, -7} {
		if got := StepNumber(depth); got != "1" {
			t.Errorf("StepNumber(%d) = %q, want %q", depth, got, "1")
		}
	}
}

// TestStepNumberAndTheListShareOneNumberer: StepNumber is the shape every
// step at a depth shares, and the list fills that shape in with the ordinals
// it counted. If those were two different joiners, a top-level step would be
// numbered "1" and a sub-step would be numbered "1.1" by the shape and "2.1"
// by the list — which is not a bug anybody would see until the numbering
// stopped agreeing with itself down the page.
func TestStepNumberAndTheListShareOneNumberer(t *testing.T) {
	for depth := 0; depth <= StepMaxDepth; depth++ {
		first := make([]int, depth+1)
		for i := range first {
			first[i] = 1
		}
		if got, want := stepNumber(first...), StepNumber(depth); got != want {
			t.Errorf("at depth %d the list would number %q and StepNumber says %q", depth, got, want)
		}
	}
}

// TestStepNumberJoinsOrdinals: the joiner itself, on the numbers a list
// actually counts rather than on ones.
func TestStepNumberJoinsOrdinals(t *testing.T) {
	cases := []struct {
		ordinals []int
		want     string
	}{
		{[]int{1}, "1"},
		{[]int{12}, "12"},
		{[]int{1, 2}, "1.2"},
		{[]int{2, 11}, "2.11"},
		{[]int{2, 1, 3}, "2.1.3"},
	}
	for _, c := range cases {
		if got := stepNumber(c.ordinals...); got != c.want {
			t.Errorf("stepNumber(%v) = %q, want %q", c.ordinals, got, c.want)
		}
	}
}

// TestStepsAreNumberedInTheOrderTheyHappened: the numbering is computed from
// the caller's order and depth and nothing else, which is the property that
// lets a transcript be stored without numbers in it.
func TestStepsAreNumberedInTheOrderTheyHappened(t *testing.T) {
	got := numberSteps([]Step{
		{Title: "a", Depth: 0},
		{Title: "a1", Depth: 1},
		{Title: "a2", Depth: 1},
		{Title: "b", Depth: 0},
		{Title: "b1", Depth: 1},
		{Title: "b1a", Depth: 2},
		{Title: "c", Depth: 0},
	})
	want := []string{"1", "1.1", "1.2", "2", "2.1", "2.1.1", "3"}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Errorf("numbered %v, want %v", got, want)
	}
	// A step that jumps two levels down is padded rather than refused: the
	// depth says what the shape is, and a flat ladder of ones is what that
	// shape means.
	jumped := numberSteps([]Step{
		{Title: "a", Depth: 0},
		{Title: "deep", Depth: 3},
	})
	if jumped[1] != "1.1.1.1" {
		t.Errorf("a step that jumped levels numbered %q, want %q", jumped[1], "1.1.1.1")
	}
}

// ── the three-way merge ─────────────────────────────────────────────────────

// The scenario, in full. Three lines, and both branches rewrote the second one
// to something different — the ordinary bad afternoon, where two people
// touched the same line and neither is going to give way.
const (
	mergeBase   = "package agent\n\nfunc itoa(n int) string { return strconv.Itoa(n) }\n"
	mergeOurs   = "package agent\n\nfunc itoa(n int) string {\n\treturn strconv.Itoa(n)\n}\n"
	mergeTheirs = "package agent\n\nfunc itoa(n int) string { return strconv.FormatInt(int64(n), 10) }\n"
)

// TestMerge3OnARealConflict is the test this package exists for. It asserts
// every line of the output — what happened to it, and what it says — because
// a merge that gets the shape right and the text wrong is still wrong, and a
// count of "three lines" would pass.
func TestMerge3OnARealConflict(t *testing.T) {
	got := Merge3(mergeBase, mergeOurs, mergeTheirs)

	want := []code.DiffLine{
		{Kind: code.DiffSame, Old: 1, New: 1, Text: "package agent"},
		{Kind: code.DiffSame, Old: 2, New: 2, Text: ""},
		// The conflict: the base line goes, and both candidates stay, ours
		// first so that a reader who has to lean leans the right way.
		{Kind: code.DiffRemoved, Old: 3, New: 0,
			Text: "func itoa(n int) string { return strconv.Itoa(n) }"},
		{Kind: code.DiffAdded, Old: 0, New: 3,
			Text: "func itoa(n int) string {"},
		{Kind: code.DiffAdded, Old: 0, New: 4,
			Text: "\treturn strconv.Itoa(n)"},
		{Kind: code.DiffAdded, Old: 0, New: 5, Text: "}"},
		{Kind: code.DiffAdded, Old: 0, New: 6,
			Text: "func itoa(n int) string { return strconv.FormatInt(int64(n), 10) }"},
	}
	assertLines(t, got, want)

	conflicts := Conflicts(mergeBase, mergeOurs, mergeTheirs)
	if len(conflicts) != 1 {
		t.Fatalf("found %d conflicts, want 1: %+v", len(conflicts), conflicts)
	}
	if conflicts[0].Base != "func itoa(n int) string { return strconv.Itoa(n) }" {
		t.Errorf("the conflict's base line is %q", conflicts[0].Base)
	}
	wantOurs := []string{"func itoa(n int) string {", "\treturn strconv.Itoa(n)", "}"}
	if strings.Join(conflicts[0].Ours, "|") != strings.Join(wantOurs, "|") {
		t.Errorf("the conflict's ours is %q, want %q", conflicts[0].Ours, wantOurs)
	}
	wantTheirs := []string{"func itoa(n int) string { return strconv.FormatInt(int64(n), 10) }"}
	if strings.Join(conflicts[0].Theirs, "|") != strings.Join(wantTheirs, "|") {
		t.Errorf("the conflict's theirs is %q, want %q", conflicts[0].Theirs, wantTheirs)
	}
	if conflicts[0].At != 2 {
		t.Errorf("the conflict starts at line %d of the merge, want 2", conflicts[0].At)
	}
	// The band a reviewer sees has to cover the whole conflict, both sides of
	// it, or the hunk is half-marked and reads as agreed-on where it is not.
	if conflicts[0].Lines() != 5 {
		t.Errorf("the conflict covers %d lines of the merge, want 5", conflicts[0].Lines())
	}
}

// TestMerge3OnAChangeOnlyOneSideMade: the common case, and the one where a
// merge that over-reaches is most expensive. Nothing is reported as a
// conflict and nothing is lost.
func TestMerge3OnAChangeOnlyOneSideMade(t *testing.T) {
	got := Merge3(mergeBase, mergeOurs, mergeBase)
	if n := len(Conflicts(mergeBase, mergeOurs, mergeBase)); n != 0 {
		t.Errorf("our change alone reported %d conflicts: %+v", n, Conflicts(mergeBase, mergeOurs, mergeBase))
	}
	assertLines(t, got, []code.DiffLine{
		{Kind: code.DiffSame, Old: 1, New: 1, Text: "package agent"},
		{Kind: code.DiffSame, Old: 2, New: 2, Text: ""},
		{Kind: code.DiffRemoved, Old: 3, Text: "func itoa(n int) string { return strconv.Itoa(n) }"},
		{Kind: code.DiffAdded, New: 3, Text: "func itoa(n int) string {"},
		{Kind: code.DiffAdded, New: 4, Text: "\treturn strconv.Itoa(n)"},
		{Kind: code.DiffAdded, New: 5, Text: "}"},
	})

	// And the other way round, because "ours" is only ours by which argument
	// it was.
	got = Merge3(mergeBase, mergeBase, mergeOurs)
	if n := len(Conflicts(mergeBase, mergeBase, mergeOurs)); n != 0 {
		t.Errorf("their change alone reported %d conflicts: %+v", n, Conflicts(mergeBase, mergeBase, mergeOurs))
	}
}

// TestMerge3WhenBothBranchesMadeTheSameEdit: two branches that arrived at the
// same text have not disagreed about anything, and reporting it as a conflict
// would train a reviewer to dismiss the conflicts that are real.
func TestMerge3WhenBothBranchesMadeTheSameEdit(t *testing.T) {
	got := Merge3(mergeBase, mergeOurs, mergeOurs)
	if n := len(Conflicts(mergeBase, mergeOurs, mergeOurs)); n != 0 {
		t.Errorf("two identical edits reported %d conflicts: %+v", n,
			Conflicts(mergeBase, mergeOurs, mergeOurs))
	}
	assertLines(t, got, []code.DiffLine{
		{Kind: code.DiffSame, Old: 1, New: 1, Text: "package agent"},
		{Kind: code.DiffSame, Old: 2, New: 2, Text: ""},
		{Kind: code.DiffRemoved, Old: 3, Text: "func itoa(n int) string { return strconv.Itoa(n) }"},
		{Kind: code.DiffAdded, New: 3, Text: "func itoa(n int) string {"},
		{Kind: code.DiffAdded, New: 4, Text: "\treturn strconv.Itoa(n)"},
		{Kind: code.DiffAdded, New: 5, Text: "}"},
	})
}

// TestMerge3PrefersTheDeletion: one branch deleted a line and the other
// rewrote it. The delete stands, because the alternative puts back a line
// somebody deliberately took out.
func TestMerge3PrefersTheDeletion(t *testing.T) {
	base := "a\nb\nc"
	theirs := "a\nb edited\nc"
	got := Merge3(base, "a\nc", theirs)
	assertLines(t, got, []code.DiffLine{
		{Kind: code.DiffSame, Old: 1, New: 1, Text: "a"},
		{Kind: code.DiffRemoved, Old: 2, Text: "b"},
		{Kind: code.DiffSame, Old: 3, New: 2, Text: "c"},
	})

	// Both deleted it: the merged file does not have the line, so the diff
	// says it went. Dropping it would make a merge of a deleted file look
	// like a merge that changed nothing.
	got = Merge3(base, "a\nc", "a\nc")
	assertLines(t, got, []code.DiffLine{
		{Kind: code.DiffSame, Old: 1, New: 1, Text: "a"},
		{Kind: code.DiffRemoved, Old: 2, Text: "b"},
		{Kind: code.DiffSame, Old: 3, New: 2, Text: "c"},
	})
}

// TestMerge3KeepsBothBranchesInsertions: two branches that each added a line
// in the same place are not in conflict — there is no base line they
// disagreed about — and both are kept, ours first.
func TestMerge3KeepsBothBranchesInsertions(t *testing.T) {
	got := Merge3("a\nb", "a\nmine\nb", "a\ntheirs\nb")
	assertLines(t, got, []code.DiffLine{
		{Kind: code.DiffSame, Old: 1, New: 1, Text: "a"},
		{Kind: code.DiffAdded, New: 2, Text: "mine"},
		{Kind: code.DiffAdded, New: 3, Text: "theirs"},
		{Kind: code.DiffSame, Old: 2, New: 4, Text: "b"},
	})
	if n := len(Conflicts("a\nb", "a\nmine\nb", "a\ntheirs\nb")); n != 0 {
		t.Errorf("two insertions reported %d conflicts", n)
	}
}

// TestMerge3OnTheDegenerateTexts: empty bases, empty sides and a file that was
// never written are all things that arrive, and none of them may panic or
// invent lines.
func TestMerge3OnTheDegenerateTexts(t *testing.T) {
	assertLines(t, Merge3("", "", ""), nil)
	assertLines(t, Merge3("", "new\nlines\n", ""), []code.DiffLine{
		{Kind: code.DiffAdded, New: 1, Text: "new"},
		{Kind: code.DiffAdded, New: 2, Text: "lines"},
	})
	assertLines(t, Merge3("gone\n", "", ""), []code.DiffLine{
		{Kind: code.DiffRemoved, Old: 1, Text: "gone"},
	})
}

// assertLines is the one way this file says what a merge decided: kind, both
// numbers and the text, for every line, in order.
func assertLines(t *testing.T, got, want []code.DiffLine) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("the merge produced %d lines, want %d:\ngot  %s\nwant %s",
			len(got), len(want), showLines(got), showLines(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("line %d is %s, want %s", i, showLines(got[i:i+1]), showLines(want[i:i+1]))
		}
	}
}

func showLines(ls []code.DiffLine) string {
	var b strings.Builder
	for _, l := range ls {
		b.WriteString("\n\t" + l.Kind.String() + " old=" + itoa(l.Old) + " new=" + itoa(l.New) +
			" " + strconv.Quote(l.Text))
	}
	if b.Len() == 0 {
		return "(nothing)"
	}
	return b.String()
}

// ── every component draws ───────────────────────────────────────────────────

// demo is a run's worth of data, drawn by every component below. It is a
// function rather than a set of package variables so that a test can change
// one thing about it without the rest moving underneath.
func demo() []Step {
	return []Step{
		{Title: "read the brief", Status: StatusDone, Duration: "0.4s"},
		{Title: "search the repo", Detail: "rg -n StatusTone ui/", Status: StatusDone,
			Depth: 1, Tool: "Grep", Duration: "1.2s"},
		{Title: "read ui/core/state.go", Status: StatusDone, Depth: 1, Duration: "0.3s"},
		{Title: "draw the components", Detail: "26 components", Status: StatusRunning},
	}
}

func TestEveryComponentDraws(t *testing.T) {
	for _, mode := range []core.Mode{core.Light, core.Dark} {
		tt := dark(t, mode, func(c *ui.Context) {
			steps := demo()
			open := &Open[int]{}
			open.Open(3)
			sel := 3
			callOpen := true
			cmdOpen := true
			groupOpen := true
			chosen := 0
			picked := ""
			subOpen := &Open[string]{}
			subOpen.Open("explorer")
			draft := ""

			AgentStatus(c, AgentStatusOptions{Status: StatusRunning, Label: "drawing", Duration: "2m"})
			AgentStatus(c, AgentStatusOptions{Status: StatusWaiting, Detail: "needs an answer", Pill: true})
			AgentParts(c, AgentPartsOptions{Name: "main", Role: "orchestrator",
				Status: StatusRunning, Footer: "~/work/callbacks",
				Badges: []AgentBadge{{Label: "opus", Tone: core.Accent}, {Label: "main"}}})
			AgentCaption(c, AgentCaptionOptions{Name: "explorer", Role: "read-only",
				At: "14:02", Note: "read 12 files"})
			AgentCard(c, AgentCardOptions{Name: "tester", Role: "go test", Status: StatusDone,
				Steps: 4, Done: 4, Tools: 9, Meter: true, Footer: "finished"})
			AgentStepList(c, AgentStepListOptions{Steps: steps, Open: open, Selected: &sel, Numbers: true})
			AgentStepList(c, AgentStepListOptions{Empty: "No steps yet"})
			AgentProgress(c, AgentProgressOptions{Status: StatusRunning, Done: 3, Total: 7,
				Detail: "running the suite", ShowPercent: true})
			AgentProgress(c, AgentProgressOptions{Status: StatusRunning, Done: 3})
			AgentPlan(c, AgentPlanOptions{Title: "Ship the gallery", Goal: "one page per package",
				Steps: steps, ShowProgress: true})
			ToolCallCard(c, ToolCallCardOptions{
				Call:   ToolCall{Tool: "Read", Summary: "ui/core/state.go", Result: "4.1 kB", Duration: "0.2s", Status: StatusDone},
				Server: "filesystem", Detail: "{\n  \"path\": \"ui/core/state.go\"\n}",
				Open: &callOpen, Height: 60,
			})
			ToolCallGroup(c, ToolCallGroupOptions{Tool: "Read", Open: &groupOpen, Height: 120,
				Calls: []ToolCall{
					{Tool: "Read", Summary: "ui/core/core.go", Status: StatusDone, Duration: "0.2s"},
					{Tool: "Read", Summary: "ui/core/state.go", Status: StatusFailed, Result: "no such file"},
				}})
			CommandExecutionCard(c, CommandExecutionCardOptions{Open: &cmdOpen, Height: 80,
				Exec: CommandExecution{Command: "go test ./ui/agent", Dir: "~/work/callbacks",
					ExitCode: 0, Status: StatusDone, Duration: "4.1s", Truncated: 12,
					Output: []string{"ok  \tcallbacks/ui/agent\t0.17s"}}})
			FileChangeCard(c, FileChangeCardOptions{Path: "ui/agent/well.go", Change: FileAdded,
				Added: 180, Bar: true, Footer: "new file"})
			FileChangeCard(c, FileChangeCardOptions{Path: "ui/agent/doc.go", Change: FileRenamed,
				From: "ui/agent/agent.go", To: "ui/agent/doc.go"})
			ArtifactPanel(c, ArtifactPanelOptions{Title: "Artifacts", ShowVersions: true,
				Artifacts: []Artifact{
					{Name: "report.html", Kind: "HTML", Size: "12 kB", Status: StatusDone,
						Version: 3, Note: "the gallery page"},
					{Name: "shot.png", Kind: "PNG", Size: "84 kB", Status: StatusRunning, Version: 1},
				}})
			ArtifactPanel(c, ArtifactPanelOptions{Title: "Artifacts", Empty: "Nothing produced yet"})
			ArtifactVersionSwitcher(c, ArtifactVersionSwitcherOptions{Name: "report.html", Current: &chosen,
				Versions: []ArtifactVersion{
					{Number: 1},
					{Number: 2, Label: "after review", Status: StatusRunning},
					{Number: 3},
				}})
			ScreenshotStream(c, ScreenshotStreamOptions{Columns: 2, Shots: []Screenshot{
				{Name: "gallery page", At: "14:02", Pixels: "1200×800"},
				{Name: "failing test", At: "14:03", Pixels: "900×600", Note: "the red one"},
			}})
			ScreenshotStream(c, ScreenshotStreamOptions{Empty: "No pictures yet"})
			CheckpointList(c, CheckpointListOptions{Checkpoints: []Checkpoint{
				{ID: "c1", Label: "before the schema change", At: "13:40", Files: 3, Status: StatusDone},
				{ID: "c2", Label: "before the merge", At: "13:58", Files: 7, Status: StatusDone, Current: true},
			}})
			CheckpointList(c, CheckpointListOptions{Empty: "No checkpoints yet"})
			MemoryPanel(c, MemoryPanelOptions{Items: []Memory{
				{Key: "go.version", Value: "1.27.1", Kind: "fact", Used: 3, Age: "2h"},
				{Key: "fixtures", Value: "internal/store/testdata", Kind: "trap", Pinned: true},
			}})
			MemoryPanel(c, MemoryPanelOptions{Empty: "Nothing remembered yet"})
			MCPServerList(c, MCPServerListOptions{Servers: []MCPServer{
				{Name: "filesystem", Transport: "stdio", Tools: 12, Status: StatusDone},
				{Name: "github", Transport: "http", Endpoint: "https://mcp.example", Tools: 30,
					Status: StatusFailed, Note: "token expired"},
			}})
			MCPServerList(c, MCPServerListOptions{Empty: "No servers configured"})
			SandboxStatus(c, SandboxOptions{On: true, Mode: "seatbelt", WorkingDir: "~/work",
				Allow: []string{"read under ~/work", "run go test"},
				Deny:  []string{"write outside ~/work", "any network call"},
				Note:  "Anything else stops and asks."})
			SandboxStatus(c, SandboxOptions{On: false, Note: "This run has no sandbox."})
			SubAgentTree(c, SubAgentTreeOptions{Open: subOpen, Root: SubAgent{
				Name: "main", Role: "orchestrator", Status: StatusRunning, Steps: 4,
				Children: []SubAgent{
					{Name: "explorer", Role: "read-only", Status: StatusDone, Steps: 3, Tools: 9},
					{Name: "tester", Role: "runs the suite", Status: StatusRunning,
						Children: []SubAgent{{Name: "go-test", Status: StatusRunning, Note: "./ui/agent"}}},
				},
			}})
			SubAgentTree(c, SubAgentTreeOptions{Empty: "No sub-agents"})
			HumanInputRequest(c, HumanInputRequestOptions{From: "main", Draft: &draft,
				Question: "Which database should the migration target?",
				Why:      "Two are configured and only one has the fixtures.",
				Urgency:  StatusWaiting, Deadline: "2m", FieldLabel: "Target database"})
			HumanInputRequest(c, HumanInputRequestOptions{From: "main", Urgency: StatusQueued,
				Question: "Which one?", Chosen: &picked,
				Choices: []string{"postgres", "sqlite"}})
			PermissionPrompt(c, PermissionPromptOptions{Preview: true, Remember: &callOpen,
				Request: PermissionRequest{Tool: "Write", What: "write ui/agent/status.go",
					Why: "the run has a change to make", Scope: "for this run", Risk: core.Warning,
					Detail: "{\n  \"path\": \"ui/agent/status.go\"\n}"}})
			ToolApprovalDialog(c, ToolApprovalDialogOptions{Preview: true, Chosen: &chosen,
				Remember: &cmdOpen, Approval: ToolApproval{Tool: "Bash", Server: "shell",
					Args: "git push origin main", Risk: core.Danger, SameTool: 6,
					Choices: []string{"Deny", "Allow once", "Always"}}})
			AgentWellRows(c, AgentWellRowsOptions{Rows: []WellRow{
				{Kind: RowStep, At: "14:00", Status: StatusDone,
					Step: Step{Title: "read the brief", Status: StatusDone}},
				{Kind: RowThinking, At: "14:01", Agent: "main", Status: StatusRunning,
					Text: "The gallery page has to draw every component."},
				{Kind: RowToolCall, At: "14:01", Status: StatusDone,
					Call: ToolCall{Tool: "Read", Summary: "ui/showcase/page.go", Result: "4.1 kB", Status: StatusDone}},
				{Kind: RowCommand, At: "14:02", Status: StatusDone,
					Exec: CommandExecution{Command: "go test ./ui/agent", ExitCode: 0, Status: StatusDone}},
				{Kind: RowFileChange, At: "14:02", Status: StatusDone, Added: 180,
					File: ReviewFile{Path: "ui/agent/well.go", Change: FileAdded}},
				{Kind: RowNote, At: "14:03", Status: StatusWaiting,
					Text: "Two branches changed the same line in shared.go."},
				{Kind: RowRequest, At: "14:03", Agent: "main", Status: StatusWaiting,
					Text: "Keep ours or take theirs?"},
			}})
			AgentWellRows(c, AgentWellRowsOptions{Empty: "Nothing has happened yet"})
			AgentDiff(c, AgentDiffOptions{Path: "ui/agent/shared.go", Change: FileModified,
				Old:  "func itoa(n int) string { return strconv.Itoa(n) }\n",
				New:  "func itoa(n int) string {\n\treturn strconv.Itoa(n)\n}\n",
				Lang: code.Go, Height: 90, Caption: "gofmt found three lines worth folding"})
			MultiFileDiffReview(c, MultiFileDiffReviewOptions{Ours: "main", Theirs: "review",
				Approved: map[string]bool{"ui/agent/shared.go": true}, Height: 200,
				Files: []ReviewFile{
					{Path: "ui/agent/shared.go", Change: FileModified,
						Base: mergeBase, Ours: mergeOurs, Theirs: mergeTheirs},
					{Path: "ui/agent/status.go", Change: FileAdded, Ours: "package agent\n"},
				}})
		})
		for _, want := range []string{
			"read the brief", "read ui/core/state.go", "drawing", "main", "explorer",
			"tester", "search the repo", "go test ./ui/agent", "report.html", "seatbelt",
			"filesystem", "go.version", "before the merge", "gallery page",
			"Which database should the migration target?", "ui/agent/status.go",
			"ui/agent/shared.go", "Allow Bash?", "1 conflict", "Keep ours or take theirs?",
		} {
			if !tt.HasText(want) {
				t.Errorf("%s: the page does not show %q", mode, want)
			}
		}
	}
}

// ── what the components report ──────────────────────────────────────────────

// TestAgentStepListReportsTheRowPressedAndTheDetailToggled reads both
// answers out of the view closure and counts them, because MyGo builds a
// frame up to three times and the last pass arrives with nothing pending —
// so a result read after settle is a result read too late.
func TestAgentStepListReportsTheRowPressedAndTheDetailToggled(t *testing.T) {
	var pressed, toggled int
	open := &Open[int]{}
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := AgentStepList(c, AgentStepListOptions{
			Steps: []Step{
				{Title: "first", Detail: "the detail"},
				{Title: "second", Detail: "the other detail"},
			},
			Open: open, Numbers: true,
		})
		if r.Selected() >= 0 {
			pressed = r.Selected()
		}
		if r.Toggled() >= 0 {
			toggled = r.Toggled()
		}
	}, 500, 300)

	if pressed != 0 {
		t.Errorf("before any press Selected() reported %d, want -1", pressed)
	}
	if err := tt.Click("second"); err != nil {
		t.Fatal(err)
	}
	if pressed != 1 {
		t.Errorf("clicking the second row reported row %d, want 1", pressed)
	}
	if err := tt.Click("Collapse Expand second"); err == nil {
		t.Fatal("the chevron cannot be named before the detail is open")
	}
	if err := tt.Click("Expand second"); err != nil {
		t.Fatal(err)
	}
	if toggled != 1 {
		t.Errorf("the chevron reported step %d, want 1", toggled)
	}
	// The open set is the caller's, and it is what the next frame reads.
	if !open.Has(1) {
		t.Error("toggling the chevron did not put step 1 into the caller's open set")
	}
}

// TestToolCallCardTogglesThroughTheCallersBool: the component keeps no state,
// so the pointer the caller passed is the only thing that changes.
func TestToolCallCardTogglesThroughTheCallersBool(t *testing.T) {
	open := false
	var toggles int
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := ToolCallCard(c, ToolCallCardOptions{
			Call:   ToolCall{Tool: "Read", Summary: "ui/core/state.go", Status: StatusDone},
			Detail: "{\n  \"path\": \"ui/core/state.go\"\n}",
			Open:   &open, Height: 60,
		})
		if r.Toggled() {
			toggles++
		}
	}, 520, 300)

	if !tt.HasText("Read") {
		t.Fatalf("the card must name its tool: %q", tt.Texts())
	}
	if open {
		t.Fatal("the card started open; the caller said closed")
	}
	if err := tt.Click("Expand Read"); err != nil {
		t.Fatal(err)
	}
	if !open {
		t.Error("the chevron did not write true into the caller's bool")
	}
	if toggles != 1 {
		t.Errorf("Toggled() was true %d times, want 1", toggles)
	}
	tt.Frame()
	if err := tt.Click("Collapse Read"); err != nil {
		t.Fatal(err)
	}
	if open {
		t.Error("the second press did not write false back into the caller's bool")
	}
}

// TestMultiFileDiffReviewApprovesByPath: the approved set is a set rather
// than a count precisely so that the seventh file approved is still approved
// after the eighth is looked at.
func TestMultiFileDiffReviewApprovesByPath(t *testing.T) {
	approved := map[string]bool{}
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		MultiFileDiffReview(c, MultiFileDiffReviewOptions{
			Approved: approved, Ours: "main", Theirs: "review", Height: 180,
			Files: []ReviewFile{
				{Path: "ui/agent/shared.go", Change: FileModified,
					Base: mergeBase, Ours: mergeOurs, Theirs: mergeTheirs},
				{Path: "ui/agent/status.go", Change: FileAdded, Ours: "package agent\n"},
			},
		})
	}, 900, 420)

	if err := tt.Click("Approve ui/agent/shared.go"); err != nil {
		t.Fatal(err)
	}
	if !approved["ui/agent/shared.go"] {
		t.Error("the approve mark did not put the path into the caller's set")
	}
	if approved["ui/agent/status.go"] {
		t.Error("approving one file approved another")
	}
	if err := tt.Click("Withdraw approval for ui/agent/shared.go"); err != nil {
		t.Fatal(err)
	}
	if approved["ui/agent/shared.go"] {
		t.Error("withdrawing did not take the path back out")
	}
}

// TestReviewApproveWritesIntoTheCallersSet and
// TestReviewNeedsACallersApprovedMap: Options arrives by value, so a map
// allocated inside the approve handler was written into this frame's copy and
// gone by the next one. A non-nil map shares its header with the caller's and
// survives; a nil one has nothing to write through, and asking for one is
// better than inventing one the caller cannot read.
func TestReviewApproveWritesIntoTheCallersSet(t *testing.T) {
	approved := map[string]bool{}
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		MultiFileDiffReview(c, MultiFileDiffReviewOptions{
			Files:    []ReviewFile{{Path: "a.go", Change: FileAdded, Ours: "package a\n"}},
			Approved: approved, Height: 180,
		})
	}, 600, 400)
	if err := tt.Click("Approve a.go"); err != nil {
		t.Fatal(err)
	}
	if !approved["a.go"] {
		t.Error("the approval did not reach the caller's map")
	}
}

func TestReviewNeedsACallersApprovedMap(t *testing.T) {
	defer func() {
		got := recover()
		if got == nil {
			t.Fatal("a review drew itself with no Approved map, so an approval would be " +
				"written into the frame's copy and lost at its end")
		}
		if msg, _ := got.(string); !strings.HasPrefix(msg, "agent: ") {
			t.Errorf("panicked with %v, want an agent: message", got)
		}
	}()
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		MultiFileDiffReview(c, MultiFileDiffReviewOptions{
			Files:  []ReviewFile{{Path: "a.go", Change: FileAdded, Ours: "package a\n"}},
			Height: 180,
		})
	}, 600, 400)
}

// TestAgentVersionSwitcherChoosesIntoTheCallersInt, and
// TestCheckpointListAsksToBeRestoredById: two more of the same rule — the
// answer is the caller's state, and the result only names what changed.
func TestAgentVersionSwitcherChoosesIntoTheCallersInt(t *testing.T) {
	current := 0
	var chosen int
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := ArtifactVersionSwitcher(c, ArtifactVersionSwitcherOptions{
			Name: "report.html", Current: &current,
			Versions: []ArtifactVersion{{Number: 1}, {Number: 2, Label: "after review"}, {Number: 3}},
		})
		if r.Chosen() >= 0 {
			chosen = r.Chosen()
		}
	}, 600, 200)

	if err := tt.Click("report.html v3"); err != nil {
		t.Fatal(err)
	}
	if current != 2 {
		t.Errorf("choosing v3 wrote index %d into the caller, want 2", current)
	}
	if chosen != 2 {
		t.Errorf("the result reported choice %d, want 2", chosen)
	}
}

func TestCheckpointListAsksToBeRestoredById(t *testing.T) {
	var restored string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := CheckpointList(c, CheckpointListOptions{Checkpoints: []Checkpoint{
			{ID: "c1", Label: "before the schema change", At: "13:40", Status: StatusDone},
			{ID: "c2", Label: "before the merge", At: "13:58", Status: StatusDone, Current: true},
		}})
		if id := r.Restored(); id != "" {
			restored = id
		}
	}, 600, 260)

	if err := tt.Click("Restore before the schema change"); err != nil {
		t.Fatal(err)
	}
	if restored != "c1" {
		t.Errorf("the restore asked for %q, want %q", restored, "c1")
	}
	// The checkpoint the run is at is not restorable, and there is no button
	// to press on it: putting files back from where you already are is not a
	// thing anybody means to do.
	if _, ok := tt.Find("Restore before the merge"); ok {
		t.Error("the current checkpoint offered a restore")
	}
}

// ── what the components refuse ──────────────────────────────────────────────

// TestTheThingsThisPackageCannotGiveArePanics: every component here either
// draws what it was asked for or stops. A component that quietly draws an
// empty box when it was handed nothing is a bug that reaches a user as a
// blank panel, and the only place to catch it is at the call.
func TestTheThingsThisPackageCannotGiveArePanics(t *testing.T) {
	cases := []struct {
		what string
		run  func()
	}{
		{"AgentParts without a name", func() {
			draw(func(c *ui.Context) { AgentParts(c, AgentPartsOptions{Status: StatusRunning}) })
		}},
		{"AgentCaption without a name", func() {
			draw(func(c *ui.Context) { AgentCaption(c, AgentCaptionOptions{}) })
		}},
		{"AgentCard without a name", func() {
			draw(func(c *ui.Context) { AgentCard(c, AgentCardOptions{}) })
		}},
		{"a step with no title", func() {
			draw(func(c *ui.Context) { AgentStepList(c, AgentStepListOptions{Steps: []Step{{}}}) })
		}},
		{"a step list with nothing and no empty state", func() {
			draw(func(c *ui.Context) { AgentStepList(c, AgentStepListOptions{}) })
		}},
		{"a plan with no title", func() {
			draw(func(c *ui.Context) { AgentPlan(c, AgentPlanOptions{}) })
		}},
		{"progress counting backwards", func() {
			draw(func(c *ui.Context) {
				AgentProgress(c, AgentProgressOptions{Done: 3, Total: -1})
			})
		}},
		{"a diff with no path", func() {
			draw(func(c *ui.Context) { AgentDiff(c, AgentDiffOptions{Height: 40}) })
		}},
		{"a diff with no height", func() {
			draw(func(c *ui.Context) {
				AgentDiff(c, AgentDiffOptions{Path: "a.go", Old: "a\n", New: "b\n"})
			})
		}},
		{"a review of no files", func() {
			draw(func(c *ui.Context) {
				MultiFileDiffReview(c, MultiFileDiffReviewOptions{Height: 40})
			})
		}},
		{"a tool call with no tool", func() {
			draw(func(c *ui.Context) { ToolCallCard(c, ToolCallCardOptions{}) })
		}},
		{"a tool detail with no height", func() {
			draw(func(c *ui.Context) {
				ToolCallCard(c, ToolCallCardOptions{
					Call: ToolCall{Tool: "Read"}, Detail: "{}", Open: new(bool),
				})
			})
		}},
		{"a command with no command", func() {
			draw(func(c *ui.Context) { CommandExecutionCard(c, CommandExecutionCardOptions{}) })
		}},
		{"command output with no viewport", func() {
			draw(func(c *ui.Context) {
				CommandExecutionCard(c, CommandExecutionCardOptions{
					Exec: CommandExecution{Command: "ls", Output: []string{"a"}},
				})
			})
		}},
		{"a file with no path", func() {
			draw(func(c *ui.Context) { FileChangeCard(c, FileChangeCardOptions{}) })
		}},
		{"a file with negative counts", func() {
			draw(func(c *ui.Context) {
				FileChangeCard(c, FileChangeCardOptions{Path: "a.go", Added: -1})
			})
		}},
		{"an artifact panel with no title", func() {
			draw(func(c *ui.Context) { ArtifactPanel(c, ArtifactPanelOptions{}) })
		}},
		{"a version switcher with no versions", func() {
			draw(func(c *ui.Context) {
				ArtifactVersionSwitcher(c, ArtifactVersionSwitcherOptions{Name: "a.html"})
			})
		}},
		{"a version switcher with no name", func() {
			draw(func(c *ui.Context) {
				ArtifactVersionSwitcher(c, ArtifactVersionSwitcherOptions{
					Versions: []ArtifactVersion{{Number: 1}},
				})
			})
		}},
		{"a checkpoint with no id", func() {
			draw(func(c *ui.Context) {
				CheckpointList(c, CheckpointListOptions{Checkpoints: []Checkpoint{{Label: "x"}}})
			})
		}},
		{"a memory with no key", func() {
			draw(func(c *ui.Context) {
				MemoryPanel(c, MemoryPanelOptions{Items: []Memory{{Value: "x"}}})
			})
		}},
		{"an MCP server with no name", func() {
			draw(func(c *ui.Context) {
				MCPServerList(c, MCPServerListOptions{Servers: []MCPServer{{Transport: "stdio"}}})
			})
		}},
		{"a sub-agent tree with no root", func() {
			draw(func(c *ui.Context) { SubAgentTree(c, SubAgentTreeOptions{}) })
		}},
		{"a question with nobody behind it", func() {
			draw(func(c *ui.Context) {
				HumanInputRequest(c, HumanInputRequestOptions{Question: "which one?"})
			})
		}},
		{"a permission with no tool", func() {
			draw(func(c *ui.Context) { PermissionPrompt(c, PermissionPromptOptions{}) })
		}},
		{"a permission with no *bool to open with", func() {
			draw(func(c *ui.Context) {
				PermissionPrompt(c, PermissionPromptOptions{Request: PermissionRequest{Tool: "Write"}})
			})
		}},
		{"a tool approval with no choices", func() {
			draw(func(c *ui.Context) {
				ToolApprovalDialog(c, ToolApprovalDialogOptions{
					Approval: ToolApproval{Tool: "Bash"}, Open: new(bool),
				})
			})
		}},
		{"a well with nothing and no empty state", func() {
			draw(func(c *ui.Context) { AgentWellRows(c, AgentWellRowsOptions{}) })
		}},
	}
	for _, c := range cases {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s was allowed through; it should stop", c.what)
				}
			}()
			c.run()
		}()
	}
}

// draw runs a view in a throwaway window, which is what a component needs
// before it can be asked to stop.
func draw(view func(c *ui.Context)) {
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		view(c)
	}, 400, 300)
}
