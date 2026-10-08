package ui_test

import (
	"fmt"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/data"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/feedback"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/navigation"
)

// Example assembles the pieces a Callbacks window is made of, in the order a
// view would read them: the theme, then the chrome, then the record, then the
// report of what happened.
func Example() {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		k, u := core.Tokens(c), core.Density(c).Unit()

		ui.Row(c).Fill().Children(func() {
			navigation.Rail(c, "Rosa Delgado",
				[]navigation.RailItem{{Name: "Callbacks", Selected: true}, {Name: "Team"}},
				[]navigation.RailItem{{Name: "Alerts", Badge: true}, {Name: "Settings"}})

			ui.Column(c).Grow(1).Padding(u * 7).Gap(u * 4).Children(func() {
				layout.PageHeader(c, layout.HeaderOptions{
					Crumbs: []string{"Callbacks", "All open"},
					Title:  "Open callbacks",
					Meta:   "27 open",
				})

				data.StatLine(c,
					data.Stat{Value: "27", Label: "open"},
					data.Stat{Value: "$7,340", Label: "open cost", Muted: true})

				layout.Board(c, func() {
					col := layout.Column(c,
						layout.ColumnOptions[string]{Title: "New"},
						func() {
							data.Card(c, data.CardOptions{
								Title: "Maple Street Bakery",
								Meta:  "CB-2871 · AC repair",
								Footer: func() {
									display.Avatar(c, "Nate Coleman")
								},
							}, nil)
						})
					if id, ok := col.Dropped(); ok {
						fmt.Println("moved", id, "into New")
					}
				})
			})

			feedback.Toast(c, "Saved", u)
		})
		_ = k
	}, 1600, 900)

	fmt.Println("Open callbacks", tt.HasText("Open callbacks"))
	fmt.Println("New", tt.HasText("New"))
	fmt.Println("Maple Street Bakery", tt.HasText("Maple Street Bakery"))
	fmt.Println("All open", tt.HasText("All open"))

	// Output:
	// Open callbacks true
	// New true
	// Maple Street Bakery true
	// All open true
}

// Example_input shows the rule the whole library follows: a component owns no
// state. A selection is the caller's variable, and a press comes back as a
// method rather than a callback the caller has to keep in step by hand.
func Example_input() {
	scope := "All open"
	grouped := false

	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		u := core.Density(c).Unit()

		ui.Row(c).Gap(u).Padding(u * 4).Children(func() {
			// The filter rows write to the caller's slice: no setter, no
			// OnChange, nothing to forget.
			for _, name := range []string{"All open", "High impact"} {
				row := navigation.FilterRow(c, name, navigation.FilterRowOptions{
					Selected: scope == name,
				})
				if row.Clicked() {
					scope = name
				}
			}
			input.Switch(c, &grouped, input.SwitchOptions{Label: "Group by branch"})
		})
	}, 900, 120)

	fmt.Println("scope:", scope, "grouped:", grouped)
	// Output: scope: All open grouped: false
}
