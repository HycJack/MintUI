package navigation_test

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/navigation"
)

// Example shows the chrome that moves a person around the window: the rail on
// the far left, and the filter lists that narrow what the main area shows.
func Example() {
	open := 27
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		u := core.Density(c).Unit()

		ui.Row(c).Fill().Children(func() {
			navigation.Rail(c, "Rosa Delgado",
				[]navigation.RailItem{{Name: "Callbacks", Selected: true}, {Name: "Team"}},
				[]navigation.RailItem{{Name: "Alerts", Badge: true}, {Name: "Settings"}})

			ui.Column(c).Width(280).Padding(u * 4).Gap(u * 2).Children(func() {
				navigation.FilterRow(c, "All open", navigation.FilterRowOptions{
					Count: &open, Selected: true,
				})
				navigation.FilterRow(c, "High impact", navigation.FilterRowOptions{})

				viewsOpen := true
				if navigation.Group(c, navigation.GroupOptions{
					Title: "Views", Open: viewsOpen,
				}, func() {
					navigation.FilterRow(c, "Repeat failures", navigation.FilterRowOptions{})
				}).Toggled() {
					viewsOpen = !viewsOpen
				}
			})
		})
	}, 800, 600)

	// Output:
}
