package data

import (
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
)

func cardView(decorated bool) func(c *ui.Context) {
	return func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: core.Light})
		Card(c, CardOptions{
			Title:      "Maple Street Bakery",
			Meta:       "CB-2871 · AC repair",
			FooterRule: decorated,
			Footer:     func() { ui.Text(c, "Nate Coleman $310") },
		}, func() { ui.Text(c, "third callback") }).Width(280)
	}
}

func TestCardSlots(t *testing.T) {
	tt := ui.NewTester(cardView(false), 340, 260)
	for _, want := range []string{"Maple Street Bakery", "CB-2871 · AC repair", "Nate Coleman $310", "third callback"} {
		if !tt.HasText(want) {
			t.Errorf("card is missing %q; %q", want, tt.Texts())
		}
	}
}

func TestCardWithoutSlots(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Card(c, CardOptions{}, func() { ui.Text(c, "Only body") })
	}, 300, 200)
	if got := tt.Texts(); len(got) != 1 || got[0] != "Only body" {
		t.Errorf("an empty card should draw only its body, drew %q", got)
	}
}

func TestCardIsDraggableWhenGivenAValue(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Card(c, CardOptions{Title: "Draggable", Draggable: "cb-2871"}, nil)
	}, 300, 200)
	if _, ok := tt.Find("Draggable"); !ok {
		t.Error("a draggable card should still lay out")
	}
}

func TestStatLineReadsValueThenUnit(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		StatLine(c,
			Stat{Value: "27", Label: "open"},
			Stat{Value: "$7,340", Label: "open cost", Muted: true},
		)
	}, 500, 120)
	for _, want := range []string{"27 open", "$7,340 open cost"} {
		if !tt.HasText(want) {
			t.Errorf("stat line is missing %q; %q", want, tt.Texts())
		}
	}
	// The unit must not come first: "open 27" is a figure nobody reads.
	for _, wrong := range []string{"open 27"} {
		if tt.HasText(wrong) {
			t.Errorf("stat line reads %q", wrong)
		}
	}
}
