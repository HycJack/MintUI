package core_test

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
)

// Example shows the one call every frame of every view starts with, and what
// it makes available afterwards.
func Example() {
	ui.NewTester(func(c *ui.Context) {
		// Once at the top, every frame. Nothing in the library reads the
		// system appearance behind this call's back.
		core.Use(c, core.Settings{Mode: core.System})

		k := core.Tokens(c)         // the palette for this frame
		u := core.Density(c).Unit() // the one spacing step everything is a multiple of

		ui.Box(c).Background(k.Surface).Padding(u * 4).Children(func() {
			ui.Text(c, "Callbacks").TextColor(k.Text)
		})
	}, 400, 200)

	// Output:
}
