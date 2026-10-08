package overlay

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/layout"
)

// PanelOptions configure a Panel. They are layout's, not overlay's: the panel
// is drawn below both layers, so that a Select's menu and a Dialog's body are
// the same surface.
type PanelOptions = layout.PanelOptions

// Panel gives host the look of a floating layer and fills it.
//
// It is a re-export of [layout.Panel]. An overlay contains controls, so this
// package imports input; if the panel lived here, input could not reach it to
// draw a drop-down menu. One layer down is the only place both can stand, so
// the panel is defined there and named here, where every overlay's own caller
// already expects to find it.
func Panel(c *ui.Context, host *ui.Element, opts PanelOptions, body func()) *ui.Element {
	return layout.Panel(c, host, opts, body)
}
