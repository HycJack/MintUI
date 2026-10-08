package input

import (
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
)

func TestButtonIsNamed(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Button(c, "Log callback", ButtonOptions{})
	}, 300, 120)
	if !tt.HasText("Log callback") {
		t.Errorf("button label missing: %q", tt.Texts())
	}
}

func TestButtonNeedsAName(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("an unnamed button should panic")
		}
	}()
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Button(c, "", ButtonOptions{})
	}, 200, 100)
}

func TestButtonClick(t *testing.T) {
	hits := 0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		if Button(c, "Save", ButtonOptions{Primary: true}).Clicked() {
			hits++
		}
	}, 300, 120)
	if err := tt.Click("Save"); err != nil {
		t.Fatal(err)
	}
	if hits != 1 {
		t.Errorf("hits = %d, want 1", hits)
	}
}

func TestIconButtonInsistsOnAName(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		IconButton(c, plusIcon, "New callback", ButtonOptions{})
	}, 200, 120)
	if !tt.HasText("New callback") {
		t.Errorf("icon button must be named: %q", tt.Texts())
	}
}

func TestSegmentedRejectsAnOutOfRangeSelection(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("a selection past the last label should panic")
		}
	}()
	sel := 5
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Segmented(c, &sel, "Team", "Mine")
	}, 300, 100)
}

func TestSegmentedFollowsItsPointer(t *testing.T) {
	sel := 0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Segmented(c, &sel, "Team", "Mine")
	}, 300, 100)
	_ = tt
	if sel != 0 {
		t.Fatalf("the caller owns the selection; it started at %d", sel)
	}
}

func TestSwitchFlipsItsBool(t *testing.T) {
	on := false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Switch(c, &on, SwitchOptions{Label: "Group by branch"})
	}, 320, 120)
	if !tt.HasText("Group by branch") {
		t.Fatalf("switch must be named: %q", tt.Texts())
	}
	if err := tt.Click("Group by branch"); err != nil {
		t.Fatal(err)
	}
	if !on {
		t.Error("a switch changes the bool it points at, with no handler")
	}
}

func TestSearchFieldTypes(t *testing.T) {
	q := ""
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		SearchField(c, &q, "Search callbacks")
	}, 400, 120)
	if err := tt.Click("Search callbacks"); err != nil {
		t.Fatal(err)
	}
	tt.Type("hillside")
	if q != "hillside" {
		t.Errorf("query = %q, want %q", q, "hillside")
	}
}

var plusIcon = ui.MustParseSVG([]byte(
	`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" ` +
		`stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M12 5.5v13M5.5 12h13"/></svg>`))
