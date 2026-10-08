package data_test

import (
	"github.com/HycJack/MintUI/ui/theme"
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/data"
)

// Example shows a card: the three slots a record needs, and nothing else. The
// card draws its own padding, surface and radius, so a caller lays out what
// the record says and the library decides how a record looks.
func Example() {
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		u := core.Density(c).Unit()

		data.StatLine(c,
			data.Stat{Value: "27", Label: "open"},
			data.Stat{Value: "$7,340", Label: "open cost", Muted: true})

		data.Card(c, data.CardOptions{
			Title: "Hillside Dental",
			Meta:  "CB-2873 · third callback in 9 days",
			// Draggable turns the card into a drag source carrying this value,
			// which is what a column drops onto it.
			Draggable: "cb-2873",
			Footer: func() {
				ui.Text(c, "$1,280").FontSize(theme.RowSize)
			},
		}, func() {
			// The body sits between the meta line and the footer.
			ui.Text(c, "Andre Thomas")
		}).Width(u * 40)
	}, 400, 400)

	// Output:
}
