package layout_test

import (
	"github.com/HycJack/MintUI/ui/theme"
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/layout"
)

// Example shows a board: lanes side by side, each one scrolling on its own,
// and the whole thing scrolling sideways when the window is too narrow. A
// lane never narrows — its width is fixed so a card title stays on one line.
func Example() {
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		k, u := core.Tokens(c), core.Density(c).Unit()

		layout.PageHeader(c, layout.HeaderOptions{
			Crumbs:    []string{"Callbacks", "All open"},
			Title:     "Open callbacks",
			Meta:      "27 open",
			SlotCount: 14,
		})

		layout.Board(c, func() {
			for _, lane := range []string{"New", "Root cause review", "Fix scheduled"} {
				col := layout.Column(c,
					layout.ColumnOptions[string]{Title: lane},
					func() {
						ui.Box(c).FillWidth().Height(u * 20).
							Radius(theme.PanelRadius).Background(k.Surface)
					})
				// Dropped is typed: the lane says what it takes.
				if id, ok := col.Dropped(); ok {
					_ = id
				}
				if col.Hovered() {
					_ = lane
				}
			}
		})
	}, 1600, 700)

	// Output:
}
