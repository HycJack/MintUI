package display_test

import (
	"github.com/HycJack/MintUI/ui/theme"
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/display"
)

// Example shows the marks that carry meaning without words: a face built from
// a name, and a meter that states a priority in bars rather than a sentence.
func Example() {
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		k, u := core.Tokens(c), core.Density(c).Unit()

		ui.Row(c).Gap(u * 4).Padding(u * 4).Children(func() {
			display.Avatar(c, "Nate Coleman")

			ui.Box(c).Padding(u*1.25, u*2.5, u*1.25, u*2.5).
				Radius(theme.PillRadius).Background(k.DangerBg).
				Children(func() {
					ui.Row(c).Gap(u * 1.5).Children(func() {
						display.Meter(c, 4, 4, core.Danger)
						ui.Text(c, "Critical").TextColor(k.Danger)
					})
				})
		})
	}, 600, 200)

	// Output:
}
