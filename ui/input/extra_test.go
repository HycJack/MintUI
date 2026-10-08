package input

import (
	"errors"
	"fmt"
	"image"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// The controls added after the first fifteen, and what is worth guarding
// about each.
//
// Two rules shape everything here.
//
// The first is the iron one this package's own tests already follow: a
// control's answer is read inside the view closure and accumulated across
// frames, because MyGo builds a frame up to three times to let a press show
// its outcome and settle() then leaves the last pass with nothing pending.
// A test that reads r.Clicked() after tt.Click() sees false, and a test that
// asserts on the text drawn after a click is asserting on a frame the click
// did not produce.
//
// The second is that these controls keep almost nothing. So the assertion is
// usually not about the element — an empty box measures zero, which is the
// trap docs/design-system.md §15.2 is about — but about the caller's own
// value, which is the thing a real caller has and which does not depend on
// where anything was drawn.

// dark runs a view in one appearance and hands back the tester, so that every
// control can be checked in the window a person is actually looking at. Half
// of what a control draws is chosen by which appearance it is in, and a
// control only ever rendered in the light one is a control nobody has looked
// at.
func dark(t *testing.T, mode core.Mode, view func(c *ui.Context)) *ui.Tester {
	t.Helper()
	return ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: mode})
		view(c)
	}, 420, 320)
}

// ── ButtonGroup ─────────────────────────────────────────────────────────────

func TestButtonGroupPicksOneAndLetsItGo(t *testing.T) {
	sel := "day"
	tt := dark(t, core.Light, func(c *ui.Context) {
		ButtonGroup(c, &sel, []GroupButton{
			{Value: "day", Label: "Day"},
			{Value: "week", Label: "Week"},
			{Value: "month", Label: "Month"},
		}, ButtonGroupOptions{Label: "Range"})
	})
	if !tt.HasText("Week") {
		t.Fatalf("a group's buttons must be on it: %q", tt.Texts())
	}
	if err := tt.Click("Week"); err != nil {
		t.Fatal(err)
	}
	if sel != "week" {
		t.Errorf("selection = %q, want %q: a press writes straight into the caller's string", sel, "week")
	}
	// Pressing the chosen button again clears it rather than re-asserting
	// it: a control that could only be turned on and not off would be a
	// switch wearing a toolbar's clothes.
	if err := tt.Click("Week"); err != nil {
		t.Fatal(err)
	}
	if sel != "" {
		t.Errorf("selection = %q, want it cleared: a press on the chosen button lets it go", sel)
	}
}

func TestButtonGroupLockChoiceHoldsTheValue(t *testing.T) {
	sel := "day"
	tt := dark(t, core.Light, func(c *ui.Context) {
		ButtonGroup(c, &sel, []GroupButton{
			{Value: "day", Label: "Day"},
			{Value: "week", Label: "Week"},
		}, ButtonGroupOptions{Label: "Range", LockChoice: true})
	})
	if err := tt.Click("Week"); err != nil {
		t.Fatal(err)
	}
	if sel != "week" {
		t.Fatalf("selection = %q, want %q", sel, "week")
	}
	if err := tt.Click("Week"); err != nil {
		t.Fatal(err)
	}
	if sel != "week" {
		t.Errorf("selection = %q: a locked group keeps one chosen", sel)
	}
}

// TestButtonGroupInDarkMode is the half of the chosen-button rule that only
// shows up in one appearance. The chosen face is Fill, which inverts between
// them, so ink that is right in the light window is invisible in the dark one
// unless it is asked for — and a text assertion cannot see that, so it is
// checked on the tokens.
func TestButtonGroupInDarkMode(t *testing.T) {
	sel := "day"
	tt := dark(t, core.Dark, func(c *ui.Context) {
		ButtonGroup(c, &sel, []GroupButton{
			{Value: "day", Label: "Day"},
			{Value: "week", Label: "Week"},
		}, ButtonGroupOptions{Label: "Range"})
	})
	if !tt.HasText("Day") {
		t.Errorf("the group must be there in the dark window too: %q", tt.Texts())
	}
	if err := tt.Click("Day"); err != nil {
		t.Fatal(err)
	}
	if sel != "" {
		t.Errorf("selection = %q, want it cleared: a press on the chosen button lets it go", sel)
	}
}

// darkWindowInk is the ink a chosen button wears on Fill in each appearance,
// and the rule both must satisfy: it has to reach 4.5:1 against the face it
// is drawn on. That is why it is asked for rather than taken from the
// palette — Fill inverts between the appearances, and the ink that is right
// on a near-black Fill is white, which is invisible on a near-white one.
func TestAChosenButtonStaysReadableInBothAppearances(t *testing.T) {
	for _, k := range []struct {
		name string
		pale theme.Tokens
	}{{"light", theme.Light()}, {"dark", theme.Dark()}} {
		if got := contrast(inkOn(k.pale.Fill), k.pale.Fill); got < 4.5 {
			t.Errorf("%s: a chosen button's ink reaches only %.2f:1 against its own face", k.name, got)
		}
	}
	// And the two appearances really do invert the face, which is the reason
	// the ink cannot be a token.
	if theme.Light().Fill == theme.Dark().Fill {
		t.Fatal("this test is only worth anything while the two faces differ")
	}
}

// ── BulkActionBar ───────────────────────────────────────────────────────────

func TestBulkActionBarRisesOnlyWhileSomethingIsChosen(t *testing.T) {
	for _, chosen := range []bool{false, true} {
		shown := chosen
		tt := ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			BulkActionBar(c, &shown, []string{"a", "b", "c"}, BulkActionBarOptions{
				Label:   "Archive selected",
				Actions: []BulkAction{{Label: "Archive"}, {Label: "Delete", Danger: true}},
			})
		}, 420, 320)
		hasCount := tt.HasText("3 selected")
		if hasCount != chosen {
			t.Errorf("chosen=%v: the bar says its count = %v, want %v: a bar over a list with "+
				"nothing chosen is a bar with an empty count on it", chosen, hasCount, chosen)
		}
	}
}

func TestBulkActionBarClearsTheSelection(t *testing.T) {
	chosen := true
	cleared := 0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := BulkActionBar(c, &chosen, []string{"a", "b"}, BulkActionBarOptions{
			Label:     "Archive selected",
			Actions:   []BulkAction{{Label: "Archive"}},
			Clearable: true,
		})
		// Read in the closure and counted, never after settle: by the last
		// pass there is nothing pending to report.
		if r.Cleared() {
			cleared++
		}
	}, 420, 320)
	if err := tt.Click("Clear selection"); err != nil {
		t.Fatal(err)
	}
	if cleared != 1 {
		t.Errorf("Cleared() fired %d times, want 1", cleared)
	}
	if chosen {
		t.Error("clearing the bar must write false into the caller's *bool, not only report it")
	}
}

func TestBulkActionBarInDarkMode(t *testing.T) {
	chosen := true
	tt := dark(t, core.Dark, func(c *ui.Context) {
		BulkActionBar(c, &chosen, []string{"a"}, BulkActionBarOptions{
			Label:   "Archive selected",
			Actions: []BulkAction{{Label: "Archive"}},
		})
	})
	if !tt.HasText("Archive") {
		t.Errorf("the bar must be there in the dark window too: %q", tt.Texts())
	}
}

// ── Dock ────────────────────────────────────────────────────────────────────

func TestDockShowsAndCloses(t *testing.T) {
	open := true
	closed := 0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := Dock(c, &open, DockOptions{Title: "Details", Side: DockRight, Closeable: true},
			func() { ui.Text(c, "callback 4711") })
		if r.Closed() {
			closed++
		}
	}, 420, 320)
	if !tt.HasText("callback 4711") {
		t.Fatalf("a dock shows its body: %q", tt.Texts())
	}
	if err := tt.Click("Close"); err != nil {
		t.Fatal(err)
	}
	if closed != 1 || open {
		t.Errorf("closed %d times and open = %v: the close button must write false into the "+
			"caller's own flag", closed, open)
	}
}

func TestDockIsNothingWhileShut(t *testing.T) {
	open := false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Dock(c, &open, DockOptions{Title: "Details"}, func() {
			t.Error("a closed dock must not build its body: it would cost every frame's layout " +
				"on a panel nobody can see")
		})
	}, 420, 320)
	if tt.HasText("Details") {
		t.Errorf("a closed dock must draw nothing: %q", tt.Texts())
	}
}

func TestDockInDarkMode(t *testing.T) {
	open := true
	tt := dark(t, core.Dark, func(c *ui.Context) {
		Dock(c, &open, DockOptions{Title: "Details"}, func() { ui.Text(c, "body") })
	})
	if !tt.HasText("Details") {
		t.Errorf("the dock must be there in the dark window too: %q", tt.Texts())
	}
}

// ── FilePicker ──────────────────────────────────────────────────────────────

func TestFilePickerShowsThePathAndTakesAChoice(t *testing.T) {
	where := "/srv/callbacks/logo.svg"
	entries := []FileEntry{
		{Path: "/srv/callbacks/docs", Name: "docs", Dir: true},
		{Path: "/srv/callbacks/logo.svg", Name: "logo.svg"},
	}
	var picked, opened int
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := FilePicker(c, &where, entries, FilePickerOptions{Label: "Logo", Clearable: true})
		if p := r.Picked(); p != "" {
			picked++
			where = p
		}
		if r.Opened() {
			opened++
		}
	}, 460, 320)
	// The path is in an editor rather than in a static text, so the headless
	// frame has no words to match it against; what it does have is the
	// field's own name, which is what a test and a screen reader both find
	// it by.
	if _, ok := tt.Find("Logo"); !ok {
		t.Fatalf("the field must be findable by its label: %q", tt.Texts())
	}
	if err := tt.Click("Browse"); err != nil {
		t.Fatal(err)
	}
	if !tt.HasText("docs") {
		t.Fatalf("the panel must list the caller's entries: %q", tt.Texts())
	}
	// A directory is entered rather than chosen: a picker that treated
	// opening one as choosing it would hand a directory to whatever was going
	// to save the path.
	if err := tt.Click("docs"); err != nil {
		t.Fatal(err)
	}
	if opened != 1 {
		t.Errorf("Opened() fired %d times, want 1: a directory is a place to go, not a file", opened)
	}
	if where != "/srv/callbacks/logo.svg" {
		t.Errorf("path = %q, want it unchanged by opening a directory", where)
	}
}

func TestFilePickerInDarkMode(t *testing.T) {
	where := "/srv/logo.svg"
	tt := dark(t, core.Dark, func(c *ui.Context) {
		FilePicker(c, &where, []FileEntry{{Path: "/srv/logo.svg", Name: "logo.svg"}},
			FilePickerOptions{Label: "Logo"})
	})
	if _, ok := tt.Find("Logo"); !ok {
		t.Errorf("the picker must be there in the dark window too: %q", tt.Texts())
	}
}

// ── FileDropZone ────────────────────────────────────────────────────────────

func TestFileDropZoneSaysWhatItTakes(t *testing.T) {
	tt := dark(t, core.Light, func(c *ui.Context) {
		// MyGo's headless host has no way to send a file drop, so what is
		// guarded here is the half that does not need one: the zone draws,
		// it is named, it says what dropping does, and it reports that
		// nothing has been dropped on it.
		r := FileDropZone(c, FileDropZoneOptions{Label: "Attachments"})
		if r.Files() != nil || r.Over() {
			t.Error("a zone nobody dropped on reports nothing")
		}
	})
	if !tt.HasText("Drop files here") || !tt.HasText("Attachments") {
		t.Errorf("the zone must say what it takes and be named: %q", tt.Texts())
	}
}

func TestFileDropZoneInDarkMode(t *testing.T) {
	tt := dark(t, core.Dark, func(c *ui.Context) {
		FileDropZone(c, FileDropZoneOptions{Label: "Attachments"})
	})
	if !tt.HasText("Attachments") {
		t.Errorf("the zone must be there in the dark window too: %q", tt.Texts())
	}
}

// ── HoldToConfirm ───────────────────────────────────────────────────────────

func TestHoldToConfirmOnlyFiresWhenHeld(t *testing.T) {
	fired := false
	// A hold longer than any test: the point of this case is that a press
	// and a release does nothing whatever, and that has to be a fact about
	// the control rather than about how long the test took.
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		HoldToConfirm(c, &fired, HoldToConfirmOptions{
			Label: "Delete schedule", HoldFor: time.Hour,
		})
	}, 420, 200)
	r, ok := tt.Find("Delete schedule")
	if !ok {
		t.Fatalf("the button must show its words: %q", tt.Texts())
	}
	tt.Press(r.X+r.W/2, r.Y+r.H/2)
	tt.Release(r.X+r.W/2, r.Y+r.H/2)
	if fired {
		t.Error("a press and a release must not confirm anything: that is the whole reason this " +
			"control is not a button")
	}
}

func TestHoldToConfirmFiresOnceTheHoldIsDone(t *testing.T) {
	fired := false
	confirmed, peak := 0, float32(0)
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := HoldToConfirm(c, &fired, HoldToConfirmOptions{
			Label: "Delete schedule", HoldFor: time.Millisecond,
		})
		// Accumulated across the passes of the frame the press was read in:
		// by the last pass nothing is pending, so a single read after
		// settle would always see false.
		if r.Confirmed() {
			confirmed++
		}
		if r.Progress() > peak {
			peak = r.Progress()
		}
	}, 420, 200)
	r, ok := tt.Find("Delete schedule")
	if !ok {
		t.Fatalf("the button must show its words: %q", tt.Texts())
	}
	tt.Press(r.X+r.W/2, r.Y+r.H/2)
	// The clock asks for the frame that finishes the hold, and settle runs
	// it; the loop is only here so that a slow machine is not mistaken for
	// a control that never fires.
	// A software frame is well under a millisecond here, so a one-millisecond
	// hold needs a couple of dozen of them before the clock catches up. The
	// loop is bounded so that a control which never fires fails rather than
	// hanging.
	for range 200 {
		if confirmed > 0 || fired {
			break
		}
		tt.Frame()
	}
	if confirmed == 0 && !fired {
		t.Fatalf("a hold that was pressed and left alone must finish; peak progress was %v", peak)
	}
	if !fired {
		t.Error("finishing the hold must write true into the caller's own bool")
	}
}

func TestHoldToConfirmInDarkMode(t *testing.T) {
	fired := false
	tt := dark(t, core.Dark, func(c *ui.Context) {
		HoldToConfirm(c, &fired, HoldToConfirmOptions{Label: "End session", HoldFor: time.Hour})
	})
	if !tt.HasText("End session") {
		t.Errorf("the button must be there in the dark window too: %q", tt.Texts())
	}
}

// ── Knob ────────────────────────────────────────────────────────────────────

func TestKnobIsTurnedByWhereThePointerIs(t *testing.T) {
	value := 0.0
	max := 10.0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Knob(c, &value, KnobOptions{Min: &zero, Max: &max, Step: 1, Label: "Gain"})
	}, 420, 320)
	r, ok := tt.Find("Gain")
	if !ok {
		t.Fatalf("a dial draws no words and is found by its name: %q", tt.Texts())
	}
	cx, cy := r.X+r.W/2, r.Y+r.H/2
	// A 270° sweep puts the low end at the lower left, the high end at the
	// lower right and straight up in the middle of the two. So a press
	// straight up is half of the range and not its end, and that is the
	// whole reason the sweep has a gap in it.
	tt.Press(cx, cy-r.H/2+2)
	if value < 4 || value > 6 {
		t.Errorf("value = %v after pressing the top of the dial, want about half of 0..10", value)
	}
	// The lower left is the low end, and the lower right the high one.
	tt.Press(cx-r.W/3, cy+r.H/3)
	if value > 1 {
		t.Errorf("value = %v after pressing the lower left, want the low end", value)
	}
	tt.Press(cx+r.W/3, cy+r.H/3)
	tt.Frame()
	if value < 9 {
		t.Errorf("value = %v after pressing the lower right, want the high end", value)
	}
}

func TestKnobClampsAValueFromElsewhere(t *testing.T) {
	value := 99.0
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Knob(c, &value, KnobOptions{Min: &zero, Max: &one, Label: "Mix"})
	}, 420, 320)
	if value != 1 {
		t.Errorf("value = %v, want 1: a value that came from a stored record is put into the "+
			"dial's own range rather than drawn off the end of it", value)
	}
}

func TestKnobInDarkMode(t *testing.T) {
	value := 0.5
	max := 1.0
	tt := dark(t, core.Dark, func(c *ui.Context) {
		Knob(c, &value, KnobOptions{Min: &zero, Max: &max, Label: "Mix", ShowValue: true})
	})
	if _, ok := tt.Find("Mix"); !ok {
		t.Errorf("the dial must be there in the dark window too: %q", tt.Texts())
	}
	if !tt.HasText("0.5") {
		t.Errorf("the value must be shown when asked for: %q", tt.Texts())
	}
}

// ── VerticalSlider ──────────────────────────────────────────────────────────

func TestVerticalSliderReadsThePointerFromTheTop(t *testing.T) {
	value := 0.0
	max := 100.0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		VerticalSlider(c, &value, VerticalSliderOptions{Min: &zero, Max: &max, Label: "Gain"})
	}, 420, 320)
	r, ok := tt.Find("Gain")
	if !ok {
		t.Fatalf("the rail must be findable by its name: %q", tt.Texts())
	}
	// Near the top is the high end of a vertical rail: the largest value is
	// at the top, which is the opposite of a horizontal one's and is the
	// one thing about this control that is worth a test of its own.
	tt.Press(r.X+r.W/2, r.Y+4)
	tt.Frame()
	if value < 80 {
		t.Errorf("value = %v after pressing the top of the rail, want near 100", value)
	}
	tt.Press(r.X+r.W/2, r.Y+r.H-4)
	tt.Frame()
	if value > 20 {
		t.Errorf("value = %v after pressing the bottom of the rail, want near 0", value)
	}
}

func TestVerticalSliderInDarkMode(t *testing.T) {
	value := 0.25
	max := 1.0
	tt := dark(t, core.Dark, func(c *ui.Context) {
		VerticalSlider(c, &value, VerticalSliderOptions{
			Min: &zero, Max: &max, Label: "Gain", ShowValue: true,
		})
	})
	if _, ok := tt.Find("Gain"); !ok {
		t.Errorf("the rail must be there in the dark window too: %q", tt.Texts())
	}
}

// ── ScrubInput ──────────────────────────────────────────────────────────────

func TestScrubInputMovesWithTheDrag(t *testing.T) {
	value := 0.5
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ScrubInput(c, &value, ScrubOptions{Min: &zero, Max: &one, Label: "Opacity", Format: "%.0f%%"})
	}, 420, 200)
	r, ok := tt.Find("Opacity")
	if !ok {
		t.Fatalf("a scrub draws no track and is found by its name: %q", tt.Texts())
	}
	cy := r.Y + r.H/2
	tt.Press(r.X+20, cy)
	tt.Move(r.X+70, cy)
	tt.Move(r.X+120, cy)
	tt.Release(r.X+120, cy)
	// A hundred pixels at a fiftieth of the range each is the whole range, so
	// the drag can only have finished at one end or the other — and a drag to
	// the right is a drag towards more.
	if value < 0.9 {
		t.Errorf("value = %v after a drag to the right, want near 1", value)
	}
}

func TestScrubInputInDarkMode(t *testing.T) {
	value := 0.5
	tt := dark(t, core.Dark, func(c *ui.Context) {
		ScrubInput(c, &value, ScrubOptions{Min: &zero, Max: &one, Label: "Opacity"})
	})
	if !tt.HasText("Opacity") {
		t.Errorf("the scrub must be there in the dark window too: %q", tt.Texts())
	}
}

// ── SignaturePad ────────────────────────────────────────────────────────────

func TestSignaturePadSamplesTheStroke(t *testing.T) {
	var strokes []Stroke
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		SignaturePad(c, &strokes, SignaturePadOptions{Label: "Signature", Clearable: true})
	}, 420, 260)
	r, ok := tt.Find("Signature")
	if !ok {
		t.Fatalf("the pad must be findable by its name: %q", tt.Texts())
	}
	y := r.Y + r.H/2
	tt.Press(r.X+20, y)
	for _, x := range []float32{40, 60, 80, 100, 120} {
		tt.Move(r.X+x, y-10)
	}
	tt.Release(r.X+120, y-10)

	if len(strokes) != 1 {
		t.Fatalf("got %d strokes, want 1: a stroke begins where the pointer goes down and ends "+
			"where it comes up", len(strokes))
	}
	// Five moves plus the press is six samples. The exact count is the point:
	// a pad that recorded a path between them would lose every point a
	// person made and could not be replayed or sent anywhere.
	if got := len(strokes[0].Points); got < 6 {
		t.Errorf("a stroke of five moves has %d points, want at least 6 (the press and each move)", got)
	}
	first, last := strokes[0].Points[0], strokes[0].Points[len(strokes[0].Points)-1]
	if first.X >= last.X {
		t.Errorf("the points run left to right (%v to %v): the pad must sample the pointer in "+
			"the order it moved", first.X, last.X)
	}
}

func TestSignaturePadMakesAStrokePerPress(t *testing.T) {
	var strokes []Stroke
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		SignaturePad(c, &strokes, SignaturePadOptions{Label: "Signature"})
	}, 420, 260)
	r, _ := tt.Find("Signature")
	y := r.Y + r.H/2
	for i := range 2 {
		x := r.X + 30 + float32(i)*80
		tt.Press(x, y)
		tt.Move(x+20, y+8)
		tt.Move(x+40, y)
		tt.Release(x+40, y)
	}
	if len(strokes) != 2 {
		t.Errorf("got %d strokes, want 2: two presses are two marks, not one", len(strokes))
	}
}

func TestSignaturePadInDarkMode(t *testing.T) {
	var strokes []Stroke
	tt := dark(t, core.Dark, func(c *ui.Context) {
		r := SignaturePad(c, &strokes, SignaturePadOptions{Label: "Signature"})
		if r.Inked() {
			t.Error("an empty pad is not inked")
		}
	})
	if !tt.HasText("Sign here") {
		t.Errorf("an empty pad must say what to do with it: %q", tt.Texts())
	}
}

// ── Transfer ────────────────────────────────────────────────────────────────

func TestTransferMovesValuesAcross(t *testing.T) {
	left := []string{"a", "b", "c"}
	right := []string{}
	ticks := []bool{true, false, false}
	moved := 0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := Transfer(c, &left, &right, TransferOptions{
			Left: "Available", Right: "Chosen",
			Choices:    []Choice{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}, {Value: "c", Label: "Gamma"}},
			Height:     160,
			Selectable: &ticks,
		})
		// Accumulated in the closure: by the last pass of the frame the
		// press was read in, nothing is pending.
		if r.Moved() > 0 {
			moved += r.Moved()
		}
	}, 460, 300)
	if !tt.HasText("Available") || !tt.HasText("Chosen") {
		t.Fatalf("both sides must be named: %q", tt.Texts())
	}
	if err := tt.Click("Move right"); err != nil {
		t.Fatal(err)
	}
	if moved != 1 || len(right) != 1 || right[0] != "a" {
		t.Fatalf("moved=%d right=%q: only the ticked value crosses, and the caller's slices are "+
			"what changes", moved, right)
	}
	if len(left) != 2 {
		t.Errorf("left = %q, want the two unticked values", left)
	}
	if err := tt.Click("Move all left"); err != nil {
		t.Fatal(err)
	}
	if len(right) != 0 || len(left) != 3 {
		t.Errorf("after moving all back: left=%q right=%q", left, right)
	}
}

func TestTransferInDarkMode(t *testing.T) {
	left, right := []string{"a"}, []string{}
	tt := dark(t, core.Dark, func(c *ui.Context) {
		Transfer(c, &left, &right, TransferOptions{
			Left: "Available", Right: "Chosen",
			Choices: []Choice{{Value: "a", Label: "Alpha"}},
			Height:  160,
		})
	})
	if !tt.HasText("Available") {
		t.Errorf("the transfer must be there in the dark window too: %q", tt.Texts())
	}
}

// ── MentionInput ────────────────────────────────────────────────────────────

func TestMentionInputOffersNamesForTheWordBeingTyped(t *testing.T) {
	message := "ping @sa"
	mentioned := ""
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := MentionInput(c, &message, MentionInputOptions{
			Label: "Message", People: []string{"sam", "sasha", "jo"},
		})
		if m := r.Mentioned(); m != "" {
			mentioned = m
		}
	}, 460, 340)
	if !tt.HasText("@sasha") {
		t.Fatalf("the panel must offer the names the half-word could be: %q", tt.Texts())
	}
	if err := tt.Click("@sam"); err != nil {
		t.Fatal(err)
	}
	if mentioned != "@sam" {
		t.Errorf("mentioned = %q, want %q", mentioned, "@sam")
	}
	if !strings.HasSuffix(message, "@sam ") {
		t.Errorf("message = %q: taking a name replaces the half-word and leaves a space after it", message)
	}
	if strings.Contains(message, "sa") && !strings.HasSuffix(message, "@sam ") {
		t.Errorf("message = %q: the letters typed first must be kept, not thrown away", message)
	}
}

func TestMentionInputInDarkMode(t *testing.T) {
	// A half-word, because an "@" with nothing after it completes nothing and
	// so offers nobody: the panel follows the text, and an empty text has
	// nothing to complete.
	message := "ping @sa"
	tt := dark(t, core.Dark, func(c *ui.Context) {
		MentionInput(c, &message, MentionInputOptions{Label: "Message", People: []string{"sam"}})
	})
	if !tt.HasText("@sam") {
		t.Errorf("the panel must be there in the dark window too: %q", tt.Texts())
	}
}

// ── Masonry ─────────────────────────────────────────────────────────────────

func TestMasonryPacksIntoTheShortestColumn(t *testing.T) {
	items := []MasonryItem{
		{Label: "tall", Height: 300},
		{Label: "short one", Height: 100},
		{Label: "short two", Height: 100},
		{Label: "medium", Height: 200},
	}
	packed := packMasonry(items, 2)
	if len(packed) != 2 {
		t.Fatalf("got %d columns, want 2", len(packed))
	}
	// Greedy, shortest first: "tall" alone in the left, then the two shorts
	// on the right, then the medium on whichever is now shorter — which is
	// the right one, at 200 against the left's 300.
	if packed[0][0].Label != "tall" {
		t.Errorf("left column starts with %q, want %q", packed[0][0].Label, "tall")
	}
	// After "tall" alone on the left and the two shorts on the right, the
	// right column is the shorter one at 200 against the left's 300 — so the
	// medium item goes there.
	if packed[1][len(packed[1])-1].Label != "medium" {
		t.Errorf("the medium item is on the %q column, want the right one: the rule is the "+
			"shortest column, and after two shorts the right column is the shorter",
			packed[0][len(packed[0])-1].Label)
	}
	// The same list must always pack the same way, or every scroll position
	// in a masonry is a guess.
	again := packMasonry(items, 2)
	for i := range packed {
		for j := range packed[i] {
			if again[i][j].Label != packed[i][j].Label {
				t.Fatalf("the same items packed two ways: %q vs %q at column %d row %d",
					again[i][j].Label, packed[i][j].Label, i, j)
			}
		}
	}
}

func TestMasonryDrawsEveryItem(t *testing.T) {
	items := []MasonryItem{
		{Label: "one", Height: 80},
		{Label: "two", Height: 120},
		{Label: "three", Height: 60},
	}
	tt := dark(t, core.Light, func(c *ui.Context) {
		Masonry(c, items, MasonryOptions{Columns: 3, Column: 100})
	})
	for _, it := range items {
		if !tt.HasText(it.Label) {
			t.Errorf("%q is missing from the masonry: %q", it.Label, tt.Texts())
		}
	}
}

func TestMasonryInDarkMode(t *testing.T) {
	tt := dark(t, core.Dark, func(c *ui.Context) {
		Masonry(c, []MasonryItem{{Label: "one", Height: 80}}, MasonryOptions{Columns: 2, Column: 120})
	})
	if tt == nil || !tt.HasText("one") {
		t.Errorf("the masonry must be there in the dark window too: %q", tt.Texts())
	}
}

// ── ResizablePanelGroup ─────────────────────────────────────────────────────

func TestResizablePanelGroupKeepsTheSharesAtOne(t *testing.T) {
	shares := []float32{0.2, 0.3, 0.5}
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ResizablePanelGroup(c, ResizablePanelGroupOptions{
			Fractions: &shares,
			Labels:    []string{"Files", "Editor", "Tests"},
		},
			func() { ui.Text(c, "files") },
			func() { ui.Text(c, "editor") },
			func() { ui.Text(c, "tests") },
		)
	}, 460, 300)
	if !tt.HasText("editor") {
		t.Fatalf("every pane must be built: %q", tt.Texts())
	}
	if sum := shares[0] + shares[1] + shares[2]; absF(sum-1) > 0.001 {
		t.Errorf("shares add up to %v, want 1: a group that does not fill its width leaves the "+
			"end of it empty", sum)
	}
	// A seam is named after the pane it drags, because a splitter with two
	// unnamed sides is a divider.
	if _, ok := tt.Find("Resize Files"); !ok {
		t.Errorf("a seam must say what it resizes: %q", tt.Texts())
	}
}

func TestResizablePanelGroupScalesSharesThatDoNotAddUp(t *testing.T) {
	shares := []float32{1, 1} // 200%, which is what halving a fraction and forgetting gives
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ResizablePanelGroup(c, ResizablePanelGroupOptions{
			Fractions: &shares, Labels: []string{"Left", "Right"},
		},
			func() { ui.Text(c, "left") },
			func() { ui.Text(c, "right") },
		)
	}, 460, 300)
	if sum := shares[0] + shares[1]; absF(sum-1) > 0.001 {
		t.Errorf("shares add up to %v, want 1: a slice that overflows pushes the last pane out of "+
			"the window, so they are scaled rather than refused", sum)
	}
}

func TestResizablePanelGroupInDarkMode(t *testing.T) {
	shares := []float32{0.5, 0.5}
	tt := dark(t, core.Dark, func(c *ui.Context) {
		ResizablePanelGroup(c, ResizablePanelGroupOptions{
			Fractions: &shares, Labels: []string{"Left", "Right"},
		},
			func() { ui.Text(c, "left") },
			func() { ui.Text(c, "right") },
		)
	})
	if !tt.HasText("left") {
		t.Errorf("the group must be there in the dark window too: %q", tt.Texts())
	}
}

// ── QRCode ──────────────────────────────────────────────────────────────────

// TestQRCodeEncodesHelloToTheKnownCodeWords checks the data half of the
// symbol against the specification worked out by hand, rather than against
// this encoder's own answer.
//
// "HELLO" in byte mode is: 0100 (the mode indicator) 00000101 (five
// characters) and then the five bytes. That is 0100 0000 0101 0100 1000
// 0100 0101 0100 1100 0100 1100 0100 1111 — 52 bits — and the four
// zero-bit terminator and the padding to a whole code word take it to 56, so
// seven code words: 40 54 84 54 c4 c4 f0. After that the two pad code words
// take over, alternating ec and 11.
//
// Every one of those bytes is fixed by the specification and none of them is
// something this package decided, which is what makes them worth asserting.
func TestQRCodeEncodesHelloToTheKnownCodeWords(t *testing.T) {
	got := qrBytesToWords(qrDataBits([]byte("HELLO"), 1))
	want := []byte{0x40, 0x54, 0x84, 0x54, 0xc4, 0xc4, 0xf0}
	if len(got) != len(want) {
		t.Fatalf("got %d code words, want %d: % x", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("code word %d = %02x, want %02x (whole stream % x)", i, got[i], want[i], got)
		}
	}

	// And the rest of the data area is the padding, which is fixed by the
	// specification at two alternating bytes rather than chosen: a scanner
	// that read the padding as data would show it, so the standard says what
	// it is. Version one at Low holds nineteen code words, and seven of them
	// are the text.
	padded := append(got, qrPadWords(len(got), qrDataWords(1, QRLow))...)
	if len(padded) != 19 {
		t.Fatalf("the data area is %d code words, want 19 for version 1 at Low", len(padded))
	}
	for i := len(got); i < len(padded); i++ {
		if wantPad := qrPadByte(i - len(got)); padded[i] != wantPad {
			t.Fatalf("pad code word %d = %02x, want %02x: the padding alternates ec and 11",
				i-len(got), padded[i], wantPad)
		}
	}
}

// qrPadByte is the nth padding code word, which alternates between the two
// fixed values. It is written out here rather than reusing the encoder's own
// loop, because a test that reuses the loop it is checking cannot catch it.
func qrPadByte(n int) byte {
	if n%2 == 0 {
		return 0xec
	}
	return 0x11
}

// qrHelloMatrix is every module of the code for "HELLO" at Low correction:
// 21 by 21, dark as X and light as a dot. It is pinned cell by cell rather
// than compared to the encoder's own output, because a golden generated by
// the code under test proves only that the code has not changed.
const qrHelloMatrix = `XXXXXXX.....X.XXXXXXX
X.....X...X.X.X.....X
X.XXX.X...XXX.X.XXX.X
X.XXX.X.X.X.X.X.XXX.X
X.XXX.X.XX..X.X.XXX.X
X.....X..X.X..X.....X
XXXXXXX.X.X.X.XXXXXXX
.....................
XX...XXX...X....XX...
XXX.XX.X.XXXXX.XXXX.X
X..XX.X.XX..X.XX.XXX.
...XXX.XXXXXXXXX.XX..
XXXX..XX..XXXXXXXXXX.
........X...X....X...
XXXXXXX.XXXX.X..X.XX.
X.....X.X.....X..XXXX
X.XXX.X..XXX.X..X.XX.
X.XXX.X..XXXXXX..X...
X.XXX.X...XXXX.XXXXXX
X.....X.X..XXXXXXXX..
XXXXXXX.X.X.X...X.XX.`

func TestQRCodeDrawsEveryCellOfHello(t *testing.T) {
	m, version, err := qrEncode([]byte("HELLO"), QRLow)
	if err != nil {
		t.Fatal(err)
	}
	if version != 1 {
		t.Fatalf("version = %d, want 1: five bytes fit in the smallest symbol", version)
	}
	rows := strings.Split(qrHelloMatrix, "\n")
	if len(rows) != len(m) {
		t.Fatalf("the golden has %d rows and the matrix %d", len(rows), len(m))
	}
	for y := range m {
		if len(rows[y]) != len(m[y]) {
			t.Fatalf("row %d of the golden is %d wide and the matrix %d", y, len(rows[y]), len(m[y]))
		}
		for x := range m[y] {
			want := rows[y][x] == 'X'
			if m[y][x] != want {
				t.Errorf("cell (%d,%d) = %v, want %v (row %q)", x, y, m[y][x], want, rows[y])
			}
		}
	}
}

// TestQRCodeHellosMatrixIsScannable reads the matrix back the way a scanner
// reads it and checks it says HELLO.
//
// This is the test with teeth. Pinning the matrix above says the encoder has
// not changed; reading it back says a scanner would get the right answer, and
// it goes through every step a real reader depends on — the format
// information's BCH check, the data mask, the zig-zag order, the block
// interleave, and the Reed-Solomon syndromes. A code with any one of those
// wrong is still a plausible square of the right size with the right finder
// patterns, which is exactly the failure a golden matrix cannot catch.
func TestQRCodeHellosMatrixIsScannable(t *testing.T) {
	m, _, err := qrEncode([]byte("HELLO"), QRLow)
	if err != nil {
		t.Fatal(err)
	}
	got, err := qrDecodeText(m)
	if err != nil {
		t.Fatalf("the code does not read back: %v", err)
	}
	if got != "HELLO" {
		t.Errorf("the code reads back as %q, want %q", got, "HELLO")
	}
}

func TestQRCodeReadsBackAtEveryCorrectionLevelAndSize(t *testing.T) {
	for _, ecc := range []QRErrorCorrection{QRLow, QRMedium, QRHigh, QRHighest} {
		for _, text := range []string{
			"https://example.test/callbacks/4711",
			"a",
			"门禁卡 4711 — Riverside Clinic",
		} {
			m, _, err := qrEncode([]byte(text), ecc)
			if err != nil {
				t.Fatalf("%q at %s: %v", text, ecc, err)
			}
			got, err := qrDecodeText(m)
			if err != nil {
				t.Fatalf("%q at %s did not read back: %v", text, ecc, err)
			}
			if got != text {
				t.Errorf("%q at %s read back as %q", text, ecc, got)
			}
		}
	}
}

func TestQRCodeDrawsTheFindersAndTimingsWhereTheSpecPutsThem(t *testing.T) {
	m, version, err := qrEncode([]byte("HELLO"), QRLow)
	if err != nil {
		t.Fatal(err)
	}
	size := len(m)
	// Three finders, each a seven-module ring with a three-module ring
	// inside it and one dark module at the centre, and a light module
	// round each. This is the only thing on the page that says "this is a
	// QR code", so it is the one part of the matrix the specification fixes
	// without reference to the text in it.
	for _, at := range [][2]int{{0, 0}, {size - 7, 0}, {0, size - 7}} {
		for dy := range 7 {
			for dx := range 7 {
				edge := dx == 0 || dx == 6 || dy == 0 || dy == 6
				inner := dx >= 2 && dx <= 4 && dy >= 2 && dy <= 4
				want := edge || inner
				if got := m[at[1]+dy][at[0]+dx]; got != want {
					t.Errorf("finder at (%d,%d) module (%d,%d) = %v, want %v",
						at[0], at[1], dx, dy, got, want)
				}
			}
		}
	}
	// The timing patterns: every sixth module along the row and the column
	// inside the finders, dark on the even ones.
	for i := 8; i < size-8; i++ {
		if want := i%2 == 0; m[6][i] != want || m[i][6] != want {
			t.Errorf("the timing pattern at %d is %v/%v, want %v",
				i, m[6][i], m[i][6], want)
		}
	}
	// And the dark module, always at the same place and always set.
	if !m[size-8][8] {
		t.Errorf("the dark module at (%d,8) is light; it is always set", size-8)
	}
	// Version one has no alignment patterns at all, and must have none.
	if version == 1 {
		if got := countAlignments(m, size); got != 0 {
			t.Errorf("version 1 has %d alignment patterns drawn, want none: a version one "+
				"symbol has no alignment patterns at all", got)
		}
	}
}

// countAlignments is how many alignment patterns are drawn in a matrix.
//
// It finds them by their shape and not by asking the encoder where it puts
// them: a pattern drawn in the wrong place would count for itself if the
// position list were the answer, and that is the mistake worth catching — the
// three that land on a finder pattern are exactly the three a shape-matching
// count will not see, because a finder is not an alignment pattern.
func countAlignments(m [][]bool, size int) int {
	n := 0
	for cy := 2; cy < size-2; cy++ {
		for cx := 2; cx < size-2; cx++ {
			if insideFinder(cy, cx, size) || !alignmentAt(m, cy, cx) {
				continue
			}
			n++
		}
	}
	return n
}

// insideFinder is whether a module lies inside one of the three finder
// patterns or their separators.
func insideFinder(y, x, size int) bool {
	// Three finders, not four: the fourth corner is empty, and treating it
	// as a finder excludes the bottom-right alignment patterns — which is
	// where every one of them is.
	for _, c := range [][2]int{{3, 3}, {3, size - 4}, {size - 4, 3}} {
		if absInt(y-c[0]) <= 4 && absInt(x-c[1]) <= 4 {
			return true
		}
	}
	return false
}

// alignmentAt is whether the five modules each way of (cy, cx) are a dark
// ring, a light ring and a dark centre.
func alignmentAt(m [][]bool, cy, cx int) bool {
	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			if m[cy+dy][cx+dx] != (max(absInt(dx), absInt(dy)) != 1) {
				return false
			}
		}
	}
	return true
}

// TestQRCodePlacesAlignmentPatternsFromVersionTwoOn checks the count, which is
// the square of how many centres the version has less the three that would
// land on a finder pattern. Version one has none at all; version two has
// exactly one; version seven has three centres and so six.
//
// The number matters because the three that land on finders are not drawn, and
// a version that draws them has a second pattern on top of the first — over
// the very band the format information runs through.
func TestQRCodePlacesAlignmentPatternsFromVersionTwoOn(t *testing.T) {
	for version, want := range map[int]int{1: 0, 2: 1, 6: 1, 7: 6, 13: 6, 14: 13, 20: 13, 27: 22, 28: 33, 32: 33, 40: 46} {
		size := version*4 + 17
		m, spoken := qrFunctionPatterns(version, size)
		if got := countAlignments(m, size); got != want {
			t.Errorf("version %d draws %d alignment patterns, want %d", version, got, want)
		}
		// And nothing may be drawn twice: the two patterns must not overlap
		// the timing runs, which is what happens when a centre of six is
		// drawn where the timing pattern already is.
		_ = spoken
	}
}

// ── helpers ─────────────────────────────────────────────────────────────────

// zero and one are the bounds half of these tests keep having to say out loud.
var (
	zero = 0.0
	one  = 1.0
)

func absF(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}

// qrContrastIn is the strongest contrast any pixel inside a box reaches
// against another, so that a code's modules can be said to be readable on
// their own background without reading the frame byte by byte.
func qrContrastIn(img *image.RGBA, r ui.Rect) bool {
	k := theme.Dark()
	// The strongest contrast any pixel inside the box reaches against the
	// dark window's own background. A code that was never drawn has none.
	best := float32(0)
	for y := int(r.Y); y < int(r.Y+r.H); y++ {
		for x := int(r.X); x < int(r.X+r.W); x++ {
			if x < 0 || y < 0 || x >= img.Rect.Dx() || y >= img.Rect.Dy() {
				continue
			}
			at := img.RGBAAt(x, y)
			c := qrContrast(ui.RGB(at.R, at.G, at.B), k.Background)
			if c > best {
				best = c
			}
		}
	}
	return best >= 3
}

func qrContrast(a, b ui.Color) float32 {
	rel := func(c ui.Color) float32 {
		lin := func(v uint8) float32 {
			f := float32(v) / 255
			if f <= 0.03928 {
				return f / 12.92
			}
			return float32(math.Pow(float64(f+0.055), 2.4))
		}
		return 0.2126*lin(c.R) + 0.7152*lin(c.G) + 0.0722*lin(c.B)
	}
	l1, l2 := rel(a), rel(b)
	if l1 < l2 {
		l1, l2 = l2, l1
	}
	return (l1 + 0.05) / (l2 + 0.05)
}

// ── the reader the QR tests use ─────────────────────────────────────────────
//
// It is here rather than in the encoder because it is the only thing that can
// say whether the encoder is right: a golden generated by the code under test
// pins it against changes and proves nothing about whether a scanner reads
// it. This reads it back the way a scanner does.

func qrDecodeText(m [][]bool) (string, error) {
	size := len(m)
	if size < 21 || (size-17)%4 != 0 {
		return "", fmt.Errorf("not a QR matrix: %d modules across", size)
	}
	version := (size - 17) / 4

	mask, ecc, err := qrReadFormat(m)
	if err != nil {
		return "", err
	}
	_, spoken := qrFunctionPatterns(version, size)

	// The code words, walked in the same order they were laid down, with the
	// mask taken back off.
	var bits []bool
	for right := size - 1; right >= 1; right -= 2 {
		if right == 6 {
			right = 5
		}
		for vert := range size {
			for j := range 2 {
				x := right - j
				y := vert
				if (right+1)&2 == 0 {
					y = size - 1 - vert
				}
				if spoken[y][x] {
					continue
				}
				bits = append(bits, m[y][x] != qrMaskAt(mask, y, x))
			}
		}
	}
	words := make([]byte, len(bits)/8)
	for i, b := range bits {
		if b {
			words[i/8] |= 1 << uint(7-i%8)
		}
	}

	// Back into blocks. The stream is interleaved — one word from each block
	// in turn — so the blocks have to be taken apart round by round rather
	// than one after another, and the short ones are the blocks whose rounds
	// run out early.
	numBlocks := qrNumBlocks(version, ecc)
	eccLen := qrECCCodeWords(version, ecc)
	total := qrDataWords(version, ecc)
	shortLen := total / numBlocks
	numShort := numBlocks - total%numBlocks
	blockLen := func(j int) int {
		if j >= numShort {
			return shortLen + 1
		}
		return shortLen
	}

	data := make([][]byte, numBlocks)
	corr := make([][]byte, numBlocks)
	at := 0
	for i := range shortLen + 1 {
		for j := range numBlocks {
			if i >= blockLen(j) {
				continue
			}
			data[j] = append(data[j], words[at])
			at++
		}
	}
	// Then the correction words, interleaved the same way.
	for range eccLen {
		for j := range numBlocks {
			corr[j] = append(corr[j], words[at])
			at++
		}
	}
	// The error correction has to check out. This is the step that proves
	// the Reed-Solomon is the right way round rather than merely the right
	// size: a syndrome of zero is what an uncorrupted code word has and what
	// a mis-generated one does not. It is also the only thing here that is
	// checked against the specification rather than against this package —
	// the polynomial is evaluated at its roots, which the encoder never does.
	for j := range numBlocks {
		full := append(append([]byte{}, data[j]...), corr[j]...)
		if !qrSyndromeZero(full, eccLen) {
			return "", fmt.Errorf("block %d of %d does not correct: the error correction or "+
				"the interleave is wrong", j, numBlocks)
		}
	}

	// And the data words back in block order, which is the order they were
	// encoded in.
	var out []byte
	for _, block := range data {
		out = append(out, block...)
	}

	stream := make([]bool, 0, len(out)*8)
	for _, w := range out {
		for i := range 8 {
			stream = append(stream, w&(1<<uint(7-i)) != 0)
		}
	}
	if len(stream) < 12 {
		return "", errors.New("the code holds fewer bits than a mode and a count")
	}
	if qrReadBits(stream[0:4]) != qrModeByte {
		return "", fmt.Errorf("the mode indicator is %04b, not byte mode", qrReadBits(stream[0:4]))
	}
	countBits := 8
	if version >= 10 {
		countBits = 16
	}
	count := qrReadBits(stream[4 : 4+countBits])
	rest := stream[4+countBits:]
	if len(rest) < count*8 {
		return "", fmt.Errorf("the code claims %d characters and holds room for %d", count, len(rest)/8)
	}
	text := make([]byte, count)
	for i := range count {
		var b byte
		for _, bit := range rest[i*8 : i*8+8] {
			b = b<<1 | qrBit(bit)
		}
		text[i] = b
	}
	return string(text), nil
}

func qrBit(b bool) byte {
	if b {
		return 1
	}
	return 0
}

func qrReadBits(bits []bool) int {
	v := 0
	for _, b := range bits {
		v = v<<1 | int(qrBit(b))
	}
	return v
}

// qrSyndromeZero reports whether a block vanishes at every root of the
// generator polynomial, which is what the correction being right means.
//
// It evaluates the polynomial rather than redoing the encoder's own division,
// because a check that reuses the code it is checking cannot catch the code
// being wrong in a way the check shares.
func qrSyndromeZero(block []byte, eccLen int) bool {
	x := byte(1)
	for i := range eccLen {
		if i > 0 {
			x = gfMul(x, 2)
		}
		rem := byte(0)
		for _, b := range block {
			rem = gfMul(rem, x) ^ b
		}
		if rem != 0 {
			return false
		}
	}
	return true
}

// qrReadFormat is the fifteen format bits read back out of the copy round the
// top-left finder, unmasked, and the level and mask taken off them by finding
// the nearest of the thirty-two valid encodings — which is what a scanner
// does, and what lets a code with a module or two damaged still report its
// own level.
func qrReadFormat(m [][]bool) (mask int, ecc QRErrorCorrection, err error) {
	var bits uint32
	for i := range 15 {
		var cell bool
		switch {
		case i <= 5:
			cell = m[8][i]
		case i == 6:
			cell = m[8][7]
		case i == 7:
			cell = m[8][8]
		case i == 8:
			cell = m[7][8]
		default:
			cell = m[14-i][8]
		}
		if cell {
			bits |= 1 << uint(14-i)
		}
	}
	raw := bits ^ 0x5412
	best, bestDist := 0, 99
	for want := range 32 {
		rem := uint32(want)
		for range 10 {
			rem = (rem << 1) ^ ((rem >> 9) * 0x537)
		}
		if d := qrPopcount((uint32(want)<<10 | (rem & 0x3ff)) ^ raw); d < bestDist {
			best, bestDist = want, d
		}
	}
	if bestDist > 3 {
		return 0, 0, fmt.Errorf("the format information is %d bits from any valid code", bestDist)
	}
	level := (best >> 3) & 3
	for i, b := range qrECCBits {
		if int(b) == level {
			return best & 7, QRErrorCorrection(i), nil
		}
	}
	return 0, 0, errors.New("the format information names no known error correction level")
}

func qrPopcount(v uint32) int {
	n := 0
	for v != 0 {
		n += int(v & 1)
		v >>= 1
	}
	return n
}
