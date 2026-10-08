package input

import (
	"strconv"
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// The standalone pieces a bigger control is built out of — a stepper button,
// a slider that says its number, a code editor, a pressed button and a titled
// panel — and what each one owes its caller.

// ── NumberInputButton ───────────────────────────────────────────────────────

func TestNumberInputButtonReportsItsPress(t *testing.T) {
	pressed := 0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		if NumberInputButton(c, NumberInputButtonOptions{Label: "Increase"}).Clicked() {
			pressed++
		}
	}, 200, 80)
	if err := tt.Click("Increase"); err != nil {
		t.Fatal(err)
	}
	if pressed != 1 {
		t.Fatalf("the first press was reported %d times, want 1", pressed)
	}
	if err := tt.Click("Increase"); err != nil {
		t.Fatal(err)
	}
	if pressed != 2 {
		t.Fatalf("the second press was reported %d times, want 2", pressed)
	}
}

func TestNumberInputButtonTakesNoPressWhileDisabled(t *testing.T) {
	pressed := 0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		if NumberInputButton(c, NumberInputButtonOptions{Label: "Decrease", Disabled: true}).Clicked() {
			pressed++
		}
	}, 200, 80)
	if err := tt.Click("Decrease"); err != nil {
		t.Fatal(err)
	}
	if pressed != 0 {
		t.Errorf("a disabled stepper took %d presses, want none", pressed)
	}
}

func TestNumberInputButtonInsistsOnAName(t *testing.T) {
	wantsPanic(t, "a stepper with no name", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			NumberInputButton(c, NumberInputButtonOptions{})
		}, 200, 80)
	})
}

func TestNumberInputButtonInDarkMode(t *testing.T) {
	tt := dark(t, core.Dark, func(c *ui.Context) {
		NumberInputButton(c, NumberInputButtonOptions{Label: "Increase"})
	})
	if _, ok := tt.Find("Increase"); !ok {
		t.Errorf("the stepper must be there in the dark window: %q", tt.Texts())
	}
}

// ── SliderControl ───────────────────────────────────────────────────────────

func TestSliderControlShowsItsValueBesideTheRail(t *testing.T) {
	v := 30.0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		SliderControl(c, &v, SliderControlOptions{Min: 0, Max: 100, Label: "Volume", Caption: "Vol"})
	}, 440, 120)
	rail, ok := tt.Find("Volume")
	if !ok {
		t.Fatalf("the rail is named for its reader: %q", tt.Texts())
	}
	caption, ok := tt.Find("Vol")
	if !ok {
		t.Fatalf("the caption sits beside the rail: %q", tt.Texts())
	}
	number, ok := tt.Find("30")
	if !ok {
		t.Fatalf("the value is said out loud beside the rail: %q", tt.Texts())
	}
	if number.X < rail.X+rail.W {
		t.Errorf("the reading is at x %.0f, which is not to the right of the rail ending at x %.0f", number.X, rail.X+rail.W)
	}
	if caption.X > number.X {
		t.Errorf("the caption comes before the number, and it is at x %.0f with the number at x %.0f", caption.X, number.X)
	}
}

func TestSliderControlMovesTheValueAndItsReading(t *testing.T) {
	v := 30.0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		SliderControl(c, &v, SliderControlOptions{Min: 0, Max: 100, Label: "Volume"})
	}, 440, 120)
	rail, ok := tt.Find("Volume")
	if !ok {
		t.Fatalf("the rail is named for its reader: %q", tt.Texts())
	}
	// A drag across the rail sets the value to the pointer, and the reading
	// beside it is the same number, so it moves with the thumb.
	x := rail.X + rail.W*0.85
	tt.Press(x, rail.Y+rail.H/2)
	tt.Release(x, rail.Y+rail.H/2)
	if v <= 60 || v > 100 {
		t.Fatalf("dragging the rail to 85%% of it landed on %.2f, want somewhere past 60", v)
	}
	if !tt.HasText(strconv.FormatFloat(v, 'f', -1, 64)) {
		t.Errorf("the reading is the value, and the value is %.2f: %q", v, tt.Texts())
	}
}

func TestSliderControlClampsWhatItIsHanded(t *testing.T) {
	v := 130.0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		SliderControl(c, &v, SliderControlOptions{Min: 0, Max: 100, Label: "Volume"})
	}, 440, 120)
	if v != 100 {
		t.Fatalf("a value past the top of the rail is clamped onto it: %.2f, want 100", v)
	}
	if !tt.HasText("100") {
		t.Errorf("the reading is the clamped value: %q", tt.Texts())
	}
}

func TestSliderControlRefusesARangeWithNoRoomInIt(t *testing.T) {
	v := 10.0
	wantsPanic(t, "a range with no room in it", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			SliderControl(c, &v, SliderControlOptions{Min: 10, Max: 10, Label: "Volume"})
		}, 440, 120)
	})
}

func TestSliderControlInsistsOnAName(t *testing.T) {
	v := 10.0
	wantsPanic(t, "a rail with no name", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			SliderControl(c, &v, SliderControlOptions{Min: 0, Max: 100})
		}, 440, 120)
	})
}

func TestSliderControlInDarkMode(t *testing.T) {
	v := 42.0
	tt := darkTester(440, 120, func(c *ui.Context) {
		SliderControl(c, &v, SliderControlOptions{Min: 0, Max: 100, Label: "Volume", Caption: "Vol"})
	})
	if !tt.HasText("Vol") || !tt.HasText("42") {
		t.Errorf("the caption and the reading must show in the dark: %q", tt.Texts())
	}
}

// ── TextInputEditor ─────────────────────────────────────────────────────────

func TestTextInputEditorCountsItsLines(t *testing.T) {
	code := "fn main() {}"
	tt := fieldTester(520, 260, func(c *ui.Context) {
		TextInputEditor(c, &code, TextInputEditorOptions{Label: "Snippet", Lines: 4})
	})
	if !tt.HasText("1 line") {
		t.Fatalf("a single line of code is one line: %q", tt.Texts())
	}
	typingInto(t, tt, "Snippet", "\n\n")
	if code != "fn main() {}\n\n" {
		t.Fatalf("the editor edits the caller's string: %q", code)
	}
	if !tt.HasText("3 lines") {
		t.Errorf("the count is the value's, so it moved with the text: %q", tt.Texts())
	}
}

func TestTextInputEditorHoldsToItsLineCap(t *testing.T) {
	code := "a\nb"
	tt := fieldTester(520, 260, func(c *ui.Context) {
		TextInputEditor(c, &code, TextInputEditorOptions{Label: "Snippet", Lines: 4, MaxLines: 3})
	})
	typingInto(t, tt, "Snippet", "\nc\nd\ne")
	if countLines(code) > 3 {
		t.Fatalf("a capped editor keeps its value to three lines; it holds %d: %q", countLines(code), code)
	}
	if code != "a\nb\nc" {
		t.Fatalf("the cap keeps the lines from the top and drops the rest: %q", code)
	}
	if !tt.HasText("3 lines") {
		t.Errorf("the count says what the cap kept: %q", tt.Texts())
	}
}

func TestTextInputEditorInsistsOnAName(t *testing.T) {
	code := ""
	wantsPanic(t, "a code field with no name", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			TextInputEditor(c, &code, TextInputEditorOptions{})
		}, 300, 200)
	})
}

func TestTextInputEditorRefusesANegativeCap(t *testing.T) {
	code := ""
	wantsPanic(t, "a cap that is a negative number of lines", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			TextInputEditor(c, &code, TextInputEditorOptions{Label: "Snippet", MaxLines: -1})
		}, 300, 200)
	})
}

func TestTextInputEditorInDarkMode(t *testing.T) {
	code := "x := 1\ny := 2"
	tt := darkTester(520, 260, func(c *ui.Context) {
		// More lines than the value holds, so the well's centre — where the
		// fill is sampled — is the well itself and not the text in it.
		TextInputEditor(c, &code, TextInputEditorOptions{Label: "Snippet", Lines: 6})
	})
	if !tt.HasText("2 lines") {
		t.Errorf("the count must show in the dark: %q", tt.Texts())
	}
	// The well is the dark surface, not the light one it would be if the
	// editor painted its own background out of a palette of its own.
	wantFill(t, tt, "Snippet", theme.Dark(), theme.Dark().Surface)
}

// ── Toggle ──────────────────────────────────────────────────────────────────

func TestToggleFlipsItsFlagAndKeepsItsFace(t *testing.T) {
	on := false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Toggle(c, &on, "Only favourites", ToggleOptions{})
	}, 320, 80)
	if err := tt.Click("Only favourites"); err != nil {
		t.Fatal(err)
	}
	if !on {
		t.Fatal("a press flips the caller's flag to true")
	}
	if err := tt.Click("Only favourites"); err != nil {
		t.Fatal(err)
	}
	if on {
		t.Fatal("the next press takes it back")
	}
}

func TestAPressedToggleFillsItselfInDarkMode(t *testing.T) {
	on := true
	tt := darkTester(320, 80, func(c *ui.Context) {
		Toggle(c, &on, "Only favourites", ToggleOptions{})
	})
	r := boxOf(t, tt, "Only favourites")
	frame := tt.Image()
	fill := theme.Dark().Fill
	found := false
	for y := int(r.Y); y < int(r.Y+r.H) && !found; y++ {
		for x := int(r.X); x < int(r.X+r.W) && !found; x++ {
			if samePixel(fieldPixel(frame, float32(x), float32(y)), fill) {
				found = true
			}
		}
	}
	if !found {
		t.Errorf("a pressed toggle is filled with the dark window's ink; none of its box is %v", fill)
	}
}

func TestToggleInsistsOnAName(t *testing.T) {
	on := false
	wantsPanic(t, "a filled box with no words", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			Toggle(c, &on, "", ToggleOptions{})
		}, 300, 80)
	})
}

func TestToggleNeedsAFlagToPointAt(t *testing.T) {
	wantsPanic(t, "a toggle with no flag of its own", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			Toggle(c, nil, "Only favourites", ToggleOptions{})
		}, 300, 80)
	})
}

// ── DockPanel ───────────────────────────────────────────────────────────────

func TestDockPanelShowsItsBodyAndCloses(t *testing.T) {
	open := true
	closed := 0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := DockPanel(c, &open, DockPanelOptions{Title: "Inspector", Closeable: true}, func() {
			ui.Text(c, "callback 4711")
		})
		if r.Closed() {
			closed++
		}
	}, 420, 320)
	if !tt.HasText("callback 4711") {
		t.Fatalf("an open panel shows its body: %q", tt.Texts())
	}
	if err := tt.Click("Close"); err != nil {
		t.Fatal(err)
	}
	if closed != 1 || open {
		t.Errorf("closed %d times and open = %v: the close button must write false into the caller's own flag", closed, open)
	}
}

func TestDockPanelIsNothingWhileShut(t *testing.T) {
	open := false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		DockPanel(c, &open, DockPanelOptions{Title: "Inspector"}, func() {
			t.Error("a closed panel must not build its body: it would cost every frame's layout on a panel nobody can see")
		})
	}, 420, 320)
	if tt.HasText("Inspector") {
		t.Errorf("a closed panel draws nothing: %q", tt.Texts())
	}
}

func TestDockPanelNeedsATitle(t *testing.T) {
	open := true
	wantsPanic(t, "a panel a reader cannot name", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			DockPanel(c, &open, DockPanelOptions{}, func() {
				ui.Text(c, "body")
			})
		}, 300, 200)
	})
}

func TestDockPanelInDarkMode(t *testing.T) {
	open := true
	tt := dark(t, core.Dark, func(c *ui.Context) {
		DockPanel(c, &open, DockPanelOptions{Title: "Inspector"}, func() { ui.Text(c, "body") })
	})
	if !tt.HasText("Inspector") || !tt.HasText("body") {
		t.Errorf("the panel and its body must show in the dark: %q", tt.Texts())
	}
}
