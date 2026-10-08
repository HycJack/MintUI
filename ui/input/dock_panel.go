package input

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// DockPanelOptions configure a DockPanel.
type DockPanelOptions struct {
	// Title heads the panel, and is required for the same reason a dialog's
	// is: a panel with no name is read out as a group, and a panel a reader
	// cannot name is a panel they cannot close.
	Title string
	// Subtitle is the second line under the title, as a Dock's is.
	Subtitle string
	// Width is the panel's width; zero takes what its content needs, which
	// is the answer for a panel that holds a form of fixed shape.
	Width float32
	// Closeable puts a button at the trailing edge of the header that writes
	// false into the caller's *bool. A panel with no way out is a panel that
	// has taken the row it is in.
	Closeable bool
}

// DockPanelResult carries a DockPanel and whether it was closed.
type DockPanelResult struct {
	// Element is the panel, or nil while it is shut.
	Element *ui.Element
	// closed is that the close button was pressed this frame.
	closed bool
}

// Closed reports that the close button was pressed this frame, so a caller
// can put whatever else it was holding — a width, a selection — back.
func (r DockPanelResult) Closed() bool { return r.closed }

// DockPanel is a titled panel that stays where it is put: a title bar, a
// close button, and the body the caller builds under it.
//
// It is a panel in the ordinary layout rather than a layer over it. The
// difference from a Dock, which is its floating sibling, is where the two
// sit: a Dock hangs from an edge of the window, over the content, square
// where it meets the edge — a bar that rises over a list, or a drawer hung
// from a side — and it moves out of the way when it is shut. A DockPanel is
// a resident of its row: it is in the layout, it takes the space it is given
// and holds it, and the window is built around it. An inspector next to a
// list, a details pane in a three-column view, a console under a canvas, are
// DockPanels; the bar that lifts over the list when a row is chosen is a
// Dock.
//
// It returns nil while shut, for the reason a Dock does: a panel that stays
// in the layout with its body built costs every frame's layout on a panel
// nobody can see.
func DockPanel(c *ui.Context, open *bool, opts DockPanelOptions, body func()) DockPanelResult {
	if open == nil {
		panic("input: DockPanel needs the *bool it opens and closes")
	}
	if !*open {
		return DockPanelResult{}
	}
	if opts.Title == "" {
		panic("input: DockPanel needs a Title; a panel a reader cannot name is a panel they cannot close")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	var r DockPanelResult
	host := ui.Column(c).FillWidth()
	if opts.Width > 0 {
		host.Width(opts.Width).Shrink(0)
	}
	// The panel's own face, as a Dock has it: one radius, one hairline, one
	// shadow, decided once in layout. The rule under the header is on,
	// because a titled panel with body straight under it needs the line
	// that keeps the two apart; a Dock does not always, because its header
	// is a bar of its own.
	panel := layout.Panel(c, host, layout.PanelOptions{
		Title: opts.Title, Subtitle: opts.Subtitle, Rule: true, Label: opts.Title,
	}, nil)
	panel.Children(func() {
		if opts.Closeable {
			closeName := core.Msg(c, "input.close", core.Def("Close"))
			btn := ui.ButtonBase(c).Size(u*6.5, u*6.5).Shrink(0).Radius(theme.PillRadius).
				Background(k.Surface).Role(ui.RoleButton).
				Label(closeName).Tooltip(closeName).Children(func() {
				ui.Icon(c, glyphCross).Size(u*3.5, u*3.5).TextColor(k.TextMuted)
			})
			if btn.Clicked() {
				r.closed = true
				*open = false
			}
		}
		if body != nil {
			body()
		}
	})
	r.Element = panel
	return r
}
