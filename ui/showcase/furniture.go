package showcase

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// useForShowcase is core.Use with the library's own defaults, in one place, so
// a page never has to remember it and never has to guess at a mode.
func useForShowcase(c *ui.Context) {
	core.Use(c, core.Settings{})
}

// Section is the one piece of page furniture every page shares: a quiet rule
// with a name above it, so a scrolled page can be read without its source.
//
// It reads this library's palette and no other. A page that mixed in tokens
// from a second theme would put two backgrounds in one frame, and the result
// is subtly wrong in a way that is hard to name.
func Section(c *ui.Context, title string) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	ui.Column(c).FillWidth().MarginY(u * 3).Gap(u * 1.5).Children(func() {
		ui.Text(c, title).TextColor(k.TextMuted).
			FontSize(core.FontSize(c, theme.CaptionSize)).FontWeight(600)
		ui.Box(c).FillWidth().Height(theme.BorderWidth).Background(k.Border)
	})
}

// Field is Section for a name that wants less weight — a caption, a group of
// similar things rather than a new section.
func Field(c *ui.Context, title string) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	ui.Text(c, title).TextColor(k.TextMuted).
		FontSize(core.FontSize(c, theme.CaptionSize)).MarginY(u * 1.5)
}

// Stack lays a row of components out with a gap, wrapping when it runs out of
// width. Every page uses it, so a row of unrelated widgets looks the same
// everywhere in the gallery.
func Stack(c *ui.Context, gap float32, children func()) {
	u := core.Density(c).Unit()
	ui.Row(c).FillWidth().Gap(gap * u).Wrap().AlignItems(ui.Center).Children(children)
}
