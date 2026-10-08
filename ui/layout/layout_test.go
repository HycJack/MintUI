package layout

import (
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

func TestPageHeaderOrder(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		PageHeader(c, HeaderOptions{
			Crumbs:    []string{"Callbacks", "All open"},
			Title:     "Open callbacks",
			Meta:      "27 open",
			SlotCount: 14,
		})
	}, 900, 240)
	for _, want := range []string{"Callbacks", "All open", "Open callbacks", "27 open", "14 technicians"} {
		if !tt.HasText(want) {
			t.Errorf("header is missing %q; %q", want, tt.Texts())
		}
	}
}

// drag presses at one point, moves in steps to another, and releases — a
// single jump is not a drag.
func drag(tt *ui.Tester, fromX, fromY, toX, toY float32) {
	tt.Press(fromX, fromY)
	for k := 1; k <= 4; k++ {
		f := float32(k) / 4
		tt.Move(fromX+(toX-fromX)*f, fromY+(toY-fromY)*f)
	}
	tt.Release(toX, toY)
}

func TestColumnRevealsTheRest(t *testing.T) {
	revealed := false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := Column(c, ColumnOptions[string]{Title: "New", More: "3 more"}, func() {
			ui.Text(c, "Maple Street Bakery")
		})
		if r.Revealed() {
			revealed = true
		}
	}, 400, 500)
	if !tt.HasText("3 more") {
		t.Fatalf("column is missing its More button: %q", tt.Texts())
	}
	if err := tt.Click("3 more"); err != nil {
		t.Fatal(err)
	}
	if !revealed {
		t.Error("pressing More should report a reveal")
	}
}

func TestColumnEmptyReplacesTheScrollView(t *testing.T) {
	hits := 0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := Column(c, ColumnOptions[string]{
			Title: "Resolved today",
			Empty: func() {
				btn := ui.Button(c, "Resolve a callback")
				if btn.Clicked() {
					hits++
				}
			},
		}, func() { ui.Text(c, "should not be drawn") })
		_ = r
	}, 400, 500)
	if tt.HasText("should not be drawn") {
		t.Error("Empty replaces the body")
	}
	if err := tt.Click("Resolve a callback"); err != nil {
		t.Fatal(err)
	}
	if hits != 1 {
		t.Errorf("a button inside Empty must stay hittable, hits = %d", hits)
	}
}

func TestColumnDropsValues(t *testing.T) {
	var got string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		col := Column(c, ColumnOptions[string]{Title: "New"}, func() {
			ui.Box(c).Size(200, 80).Drag("cb-2871").Label("Maple Street Bakery").Children(func() {})
		})
		if id, ok := col.Dropped(); ok {
			got = id
		}
	}, 400, 500)
	src, _ := tt.Find("Maple Street Bakery")
	drag(tt, src.X+20, src.Y+20, src.X+60, src.Y+70)
	if got != "cb-2871" {
		t.Errorf("dropped = %q, want cb-2871", got)
	}
}

func TestColumnsKeepTheirWidth(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Board(c, func() {
			Column(c, ColumnOptions[string]{Title: "New"}, func() {
				ui.Box(c).FillWidth().Height(40).Label("Card one").Children(func() {})
			})
			Column(c, ColumnOptions[string]{Title: "Root cause review"}, func() {})
		})
	}, 900, 700)
	// Four columns would not fit here: the board scrolls sideways instead of
	// squeezing each lane into an unreadable ribbon.
	card, ok := tt.Find("Card one")
	if !ok {
		t.Fatalf("card missing: %q", tt.Texts())
	}
	if card.W < theme.ColumnWidth-48 {
		t.Errorf("card width = %v, column = %v — columns must scroll rather than shrink",
			card.W, theme.ColumnWidth)
	}
}
