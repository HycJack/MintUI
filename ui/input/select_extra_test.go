package input

import (
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// The pieces a dropdown is built out of — the closed trigger, the open panel,
// the rows and the marks in them — and what each one owes its caller.

// accentPixels counts the pixels of a frame close enough to accent to be the
// tick drawn in it, which is how a mark is checked that has no words to be
// found by.
func accentPixels(t *testing.T, tt *ui.Tester, accent ui.Color) int {
	t.Helper()
	frame := tt.Image()
	n := 0
	for y := 0; y < frame.Bounds().Dy(); y++ {
		for x := 0; x < frame.Bounds().Dx(); x++ {
			p := frame.RGBAAt(x, y)
			if p.A > 200 && abs32(p.R, accent.R) < 24 && abs32(p.G, accent.G) < 24 && abs32(p.B, accent.B) < 24 {
				n++
			}
		}
	}
	return n
}

// ── SelectField ─────────────────────────────────────────────────────────────

func TestSelectFieldSaysItsPlaceholderUntilThereIsAChoice(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		SelectField(c, SelectFieldOptions{Label: "Assign to", Placeholder: "Nobody yet"})
	}, 360, 120)
	if !tt.HasText("Nobody yet") {
		t.Fatalf("with no choice the trigger says its placeholder: %q", tt.Texts())
	}
	if tt.HasText("Dana Reyes") {
		t.Error("the placeholder must not show the words of a choice it does not have")
	}

	chosen := "Dana Reyes"
	tt = ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		SelectField(c, SelectFieldOptions{Label: "Assign to", Placeholder: "Nobody yet", Value: chosen})
	}, 360, 120)
	if !tt.HasText("Dana Reyes") {
		t.Fatalf("with a choice the trigger says its words: %q", tt.Texts())
	}
	if tt.HasText("Nobody yet") {
		t.Error("a trigger with a choice does not say its placeholder")
	}
}

func TestSelectFieldInsistsOnAName(t *testing.T) {
	wantsPanic(t, "a trigger with no name of its own", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			SelectField(c, SelectFieldOptions{Placeholder: "Nobody yet"})
		}, 300, 100)
	})
}

func TestSelectFieldNeedsSomethingToSay(t *testing.T) {
	wantsPanic(t, "a trigger with no choice and no words", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			SelectField(c, SelectFieldOptions{Label: "Assign to"})
		}, 300, 100)
	})
}

func TestSelectFieldInDarkMode(t *testing.T) {
	chosen := "Hillside"
	tt := dark(t, core.Dark, func(c *ui.Context) {
		SelectField(c, SelectFieldOptions{Label: "Branch", Placeholder: "Every branch", Value: chosen})
	})
	if !tt.HasText("Hillside") {
		t.Errorf("the dark trigger must say its choice: %q", tt.Texts())
	}
}

// ── SelectPanel ─────────────────────────────────────────────────────────────

func TestSelectPanelHoldsItsRowsAndChooses(t *testing.T) {
	sel := ""
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		SelectPanel(c, SelectPanelOptions{Label: "Assign to"}, func() {
			if SelectRow(c, SelectRowOptions{Label: "Dana Reyes"}).Clicked() {
				sel = "dana"
			}
			if SelectRow(c, SelectRowOptions{Label: "Sam Okafor"}).Clicked() {
				sel = "sam"
			}
			SelectNote(c, "↑↓ select · ↵ confirm")
		})
	}, 360, 220)
	for _, want := range []string{"Dana Reyes", "Sam Okafor", "↑↓ select · ↵ confirm"} {
		if !tt.HasText(want) {
			t.Fatalf("the panel must hold its rows and its note: %q", tt.Texts())
		}
	}
	if err := tt.Click("Sam Okafor"); err != nil {
		t.Fatal(err)
	}
	if sel != "sam" {
		t.Fatalf("pressing a row reports itself, which the caller writes: %q", sel)
	}
}

func TestSelectPanelNeedsABody(t *testing.T) {
	wantsPanic(t, "a panel with nothing to hold", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			SelectPanel(c, SelectPanelOptions{Label: "Assign to"}, nil)
		}, 300, 160)
	})
}

func TestSelectPanelInDarkMode(t *testing.T) {
	tt := dark(t, core.Dark, func(c *ui.Context) {
		SelectPanel(c, SelectPanelOptions{Label: "Branches"}, func() {
			SelectRow(c, SelectRowOptions{Label: "Hillside", Selected: true})
			SelectRow(c, SelectRowOptions{Label: "Meadowbrook"})
		})
	})
	if !tt.HasText("Hillside") || !tt.HasText("Meadowbrook") {
		t.Errorf("the panel and its rows must show in the dark: %q", tt.Texts())
	}
}

// ── SelectRow ───────────────────────────────────────────────────────────────

func TestSelectRowReportsItsPressToTheCaller(t *testing.T) {
	sel := ""
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		if SelectRow(c, SelectRowOptions{Label: "Hillside"}).Clicked() {
			sel = "h"
		}
		if SelectRow(c, SelectRowOptions{Label: "Meadowbrook"}).Clicked() {
			sel = "m"
		}
	}, 320, 100)
	if err := tt.Click("Hillside"); err != nil {
		t.Fatal(err)
	}
	if sel != "h" {
		t.Fatalf("the first press chose %q, want h", sel)
	}
	if err := tt.Click("Meadowbrook"); err != nil {
		t.Fatal(err)
	}
	if sel != "m" {
		t.Fatalf("the second press chose %q, want m", sel)
	}
}

// TestASelectedRowShowsItsTick: the tick is the only thing that says the row
// is the one, so it is checked in pixels — a row whose tick is missing reads
// as an unchosen row with a filled face, and a value assertion cannot tell
// the two apart.
func TestASelectedRowShowsItsTick(t *testing.T) {
	ticks := func(selected bool) int {
		tt := ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			SelectRow(c, SelectRowOptions{Label: "Branch", Selected: selected})
		}, 320, 80)
		return accentPixels(t, tt, theme.Light().Accent)
	}
	if n := ticks(true); n < 8 {
		t.Errorf("a selected row draws its tick; the frame holds %d accent pixels, want at least 8", n)
	}
	if n := ticks(false); n > 0 {
		t.Errorf("an unselected row is not marked; the frame holds %d accent pixels, want none", n)
	}
}

func TestSelectRowInsistsOnAName(t *testing.T) {
	wantsPanic(t, "a row with no words beside its mark", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			SelectRow(c, SelectRowOptions{Selected: true})
		}, 300, 100)
	})
}

func TestSelectRowInDarkMode(t *testing.T) {
	tt := dark(t, core.Dark, func(c *ui.Context) {
		SelectRow(c, SelectRowOptions{Label: "Hillside", Selected: true})
	})
	if !tt.HasText("Hillside") {
		t.Fatalf("the row must show in the dark: %q", tt.Texts())
	}
	if n := accentPixels(t, tt, theme.Dark().Accent); n < 8 {
		t.Errorf("the dark row keeps its tick in the dark accent; found %d accent pixels, want at least 8", n)
	}
}

// ── SelectCheck ─────────────────────────────────────────────────────────────

func TestSelectCheckDrawsItsTick(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		SelectCheck(c)
	}, 120, 60)
	if n := accentPixels(t, tt, theme.Light().Accent); n < 8 {
		t.Errorf("a check is its tick; the frame holds %d accent pixels, want at least 8", n)
	}
}

func TestSelectCheckInDarkMode(t *testing.T) {
	tt := dark(t, core.Dark, func(c *ui.Context) {
		SelectCheck(c)
	})
	if n := accentPixels(t, tt, theme.Dark().Accent); n < 8 {
		t.Errorf("the dark check keeps its tick; found %d accent pixels, want at least 8", n)
	}
}

// ── SelectNote ──────────────────────────────────────────────────────────────

func TestSelectNoteSaysItsWords(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		SelectNote(c, "↑↓ select · ↵ confirm")
	}, 320, 60)
	if !tt.HasText("↑↓ select · ↵ confirm") {
		t.Errorf("the note is its words: %q", tt.Texts())
	}
}

func TestSelectNoteNeedsItsWords(t *testing.T) {
	wantsPanic(t, "a note with nothing to say", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			SelectNote(c, "")
		}, 300, 60)
	})
}

func TestSelectNoteInDarkMode(t *testing.T) {
	tt := dark(t, core.Dark, func(c *ui.Context) {
		SelectNote(c, "↑↓ select · ↵ confirm")
	})
	if !tt.HasText("↑↓ select · ↵ confirm") {
		t.Errorf("the note must show in the dark: %q", tt.Texts())
	}
}

// ── MultiSelectCheck ────────────────────────────────────────────────────────

func TestMultiSelectCheckTogglesItsSelection(t *testing.T) {
	sel := []string{"h"}
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r1 := MultiSelectCheck(c, MultiSelectCheckOptions{Label: "Hillside", Checked: hasValue(sel, "h")})
		if r1.Clicked() {
			if hasValue(sel, "h") {
				sel = without(sel, "h")
			} else {
				sel = append(sel, "h")
			}
		}
		r2 := MultiSelectCheck(c, MultiSelectCheckOptions{Label: "Meadowbrook", Checked: hasValue(sel, "m")})
		if r2.Clicked() {
			if hasValue(sel, "m") {
				sel = without(sel, "m")
			} else {
				sel = append(sel, "m")
			}
		}
	}, 320, 100)
	if err := tt.Click("Meadowbrook"); err != nil {
		t.Fatal(err)
	}
	if strings.Join(sel, ",") != "h,m" {
		t.Fatalf("pressing an unchecked row adds it: %v, want [h m]", sel)
	}
	if err := tt.Click("Hillside"); err != nil {
		t.Fatal(err)
	}
	if strings.Join(sel, ",") != "m" {
		t.Fatalf("pressing a checked row takes it out: %v, want [m]", sel)
	}
}

func TestMultiSelectCheckInsistsOnAName(t *testing.T) {
	wantsPanic(t, "a row with a box and no words", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			MultiSelectCheck(c, MultiSelectCheckOptions{Checked: true})
		}, 300, 100)
	})
}

func TestMultiSelectCheckInDarkMode(t *testing.T) {
	tt := dark(t, core.Dark, func(c *ui.Context) {
		MultiSelectCheck(c, MultiSelectCheckOptions{Label: "Hillside", Checked: true})
		MultiSelectCheck(c, MultiSelectCheckOptions{Label: "Meadowbrook"})
	})
	if !tt.HasText("Hillside") || !tt.HasText("Meadowbrook") {
		t.Errorf("both rows must show in the dark: %q", tt.Texts())
	}
	// The checked row carries its box's tick in the dark accent; the
	// unchecked one does not, so the two still tell apart what they hold.
	if n := accentPixels(t, tt, theme.Dark().Accent); n < 8 {
		t.Errorf("the checked row keeps its tick in the dark; found %d accent pixels, want at least 8", n)
	}
}
