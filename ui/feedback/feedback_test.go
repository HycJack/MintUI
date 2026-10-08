package feedback

import (
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

func TestEmptyExplainsAndOffers(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		if Empty(c, EmptyOptions{
			Title:  "Nothing resolved yet today",
			Body:   "Callbacks land here once the customer confirms the fix.",
			Action: "Resolve a callback",
		}).Pressed() {
			tt2 := "pressed"
			_ = tt2
		}
	}, 600, 420)
	for _, want := range []string{"Nothing resolved yet today", "Callbacks land here", "Resolve a callback"} {
		if !tt.HasText(want) {
			t.Errorf("empty state is missing %q; %q", want, tt.Texts())
		}
	}
}

func TestEmptyWithoutAnAction(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Empty(c, EmptyOptions{Title: "No callbacks"})
	}, 600, 320)
	if !tt.HasText("No callbacks") {
		t.Errorf("title missing: %q", tt.Texts())
	}
	if err := tt.Click("Resolve a callback"); err == nil {
		t.Error("an empty state with no action should not draw one")
	}
}

func TestEmptyReportsItsAction(t *testing.T) {
	pressed := false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		if Empty(c, EmptyOptions{Title: "No callbacks", Action: "Resolve a callback"}).Pressed() {
			pressed = true
		}
	}, 600, 420)
	if err := tt.Click("Resolve a callback"); err != nil {
		t.Fatal(err)
	}
	if !pressed {
		t.Error("the empty state's action did not report a press")
	}
}

func TestToastCarriesItsMessage(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Toast(c, "Saved", theme.Compact.Unit())
	}, 800, 120)
	if !tt.HasText("Saved") {
		t.Errorf("toast missing: %q", tt.Texts())
	}
}

func TestToastWithdrawsWhenEmpty(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Toast(c, "", theme.Compact.Unit())
	}, 800, 120)
	if got := tt.Texts(); len(got) != 0 {
		t.Errorf("an empty message should draw nothing, drew %q", got)
	}
}
