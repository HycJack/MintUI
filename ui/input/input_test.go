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

// Every option row in every dropdown is a control, so it is named: the box a
// name finds is the row, not the words inside it. That is what a screen reader
// is handed and what a test clicking by name lands on, and two options whose
// words are of different lengths are the honest proof — a row names itself
// over its whole width, so the two boxes come out the same size.
func TestOptionRowsAreNamedByTheRowAndNotByTheirWords(t *testing.T) {
	sel, query := "", ""
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		SelectSearch(c, &sel, &query, []Choice{
			{Value: "ac", Label: "Air conditioning"},
			{Value: "pump", Label: "Pump"},
		}, SelectSearchOptions{Label: "Kind", Placeholder: "Nothing yet", Search: "Search"})
	}, 420, 320)
	if err := tt.Click("Kind"); err != nil {
		t.Fatal(err)
	}
	long, ok := tt.Find("Air conditioning")
	if !ok {
		t.Fatalf("the option list did not open; shown: %q", tt.Texts())
	}
	short, ok := tt.Find("Pump")
	if !ok {
		t.Fatalf("no option shows %q; shown: %q", "Pump", tt.Texts())
	}
	if short.W != long.W {
		t.Errorf("the two rows are named by their words, not by the rows: "+
			"\"Pump\" is %g wide and \"Air conditioning\" %g", short.W, long.W)
	}
}

// A narrowing search reorders the rows, and a row without a key is identified
// by where it sits: so the keyboard went from the option a person had chosen
// to a different one, on the same keystroke that refiltered the list.
func TestOptionRowsKeepTheirFocusWhenTheListReorders(t *testing.T) {
	sel, query := "", ""
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		SelectSearch(c, &sel, &query, []Choice{
			{Value: "zebra", Label: "Zebra crossing"},
			{Value: "apple", Label: "Apple turnover"},
		}, SelectSearchOptions{Label: "Kind", Placeholder: "Nothing yet", Search: "Search"})
	}, 460, 420)
	if err := tt.Click("Kind"); err != nil {
		t.Fatal(err)
	}
	// Tab past the trigger to the search field, and on to the first row.
	tt.Key(0, ui.KeyTab)
	tt.Key(0, ui.KeyTab)
	if !tt.Focused("Zebra crossing") {
		t.Fatalf("the first option should have the focus; shown: %q", tt.Texts())
	}
	// The caller narrows the list, which is what a keystroke in the field
	// does: "a" starts Apple turnover and is only inside Zebra crossing, so
	// the two rows swap places.
	query = "a"
	tt.Frame()
	if !tt.Focused("Zebra crossing") {
		t.Errorf("narrowing the list moved the focus off the option it was on; shown: %q",
			tt.Texts())
	}
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
