package overlay

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// Tooltip hangs a small panel off anchor: what the control does, said where
// the control is, for the case where there is no room to say it in the
// control's own label.
//
// It shows once the pointer has rested on anchor, or at once as the keyboard
// focus comes to it, and goes as the pointer leaves, as the control is
// pressed, or with Escape. That timing is MyGo's, not this package's, so a
// tip cannot be made to appear early — which is what keeps a page of buttons
// from flickering through a tip each.
//
//	core.Use(c, core.Settings{})
//	save := input.Button(c, "Save", input.ButtonOptions{})
//	Tooltip(c, save, "Saves the callback and clears the form")
func Tooltip(c *ui.Context, anchor *ui.Element, text string) *ui.Element {
	if anchor == nil {
		panic("overlay: Tooltip needs the element it hangs on")
	}
	if text == "" {
		panic("overlay: Tooltip needs text; an empty tip describes nothing and covers the control it hides")
	}
	u := core.Density(c).Unit()

	ui.TooltipBase(c, anchor, func(tip *ui.Element) {
		Panel(c, tip, PanelOptions{
			Compact:  true,
			MaxWidth: u * 70,
			Label:    text,
		}, func() {
			ui.Text(c, text).FontSize(core.FontSize(c, theme.CaptionSize)).
				TextColor(core.Tokens(c).Text)
		})
	})
	// The anchor comes back, not the tip: the tip only exists on the frames
	// it shows, and what the caller named is what it will want again.
	return anchor
}
