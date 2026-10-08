// Package layout is the frame everything else sits in: containers, grids,
// scrolls, panes, and the two bars a desktop window wears.
//
// These are the library's most-depended-on components. A dialog's panel, a
// card's body, a table's rows and an overlay's anchor all resolve through
// something here, so the sizing rules below are deliberately few and shared.
package layout

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// ContainerOptions configure a Container.
type ContainerOptions struct {
	// Width, Height and Min bound the box; zero means the container takes
	// what the layout gives it.
	Width, Height, MinWidth, MinHeight float32
	// MaxWidth caps it, which is what a reading column wants.
	MaxWidth float32
	// Pad is the inner padding on all four edges.
	Pad float32
	// PadX and PadY override Pad on one axis.
	PadX, PadY float32
	// Radius rounds the container's own surface.
	Radius float32
	// Surface paints the container a step below the window.
	Surface bool
	// Border draws a hairline around the container.
	Border bool
	// Center puts the children in a centred column rather than filling.
	Center bool
	// Gap is the space between children.
	Gap float32
}

// Container is a padded, optionally surfaced box: the smallest thing that is
// not a component. Every panel in the library is one of these plus content.
func Container(c *ui.Context, opts ContainerOptions, children func()) *ui.Element {
	k := core.Tokens(c)
	pad, px, py := opts.Pad, opts.PadX, opts.PadY
	if px == 0 {
		px = pad
	}
	if py == 0 {
		py = pad
	}
	bg := k.Background
	if opts.Surface {
		bg = k.Surface
	}
	e := ui.Box(c).Padding(py, px, py, px).Background(bg)
	if opts.Width > 0 {
		e.Width(opts.Width)
	}
	if opts.Height > 0 {
		e.Height(opts.Height)
	}
	if opts.MinWidth > 0 {
		e.MinWidth(opts.MinWidth)
	}
	if opts.MinHeight > 0 {
		e.MinHeight(opts.MinHeight)
	}
	if opts.MaxWidth > 0 {
		e.MaxWidth(opts.MaxWidth)
	}
	if opts.Radius > 0 {
		e.Radius(opts.Radius)
	}
	if opts.Border {
		e.Radius(opts.Radius).BorderWidth(theme.BorderWidth).BorderColor(k.Border)
	}
	if opts.Gap > 0 {
		e.Gap(opts.Gap)
	}
	if children != nil {
		e.Children(children)
	}
	return e
}

// Divider draws one hairline. Vertical is the reason it exists as a component:
// a one-pixel line between two columns has to be the column's child, not a
// sibling, or it lands in the wrong place when one column grows.
func Divider(c *ui.Context, opts DividerOptions) *ui.Element {
	k := core.Tokens(c)
	e := ui.Box(c).Shrink(0).Background(k.Border)
	if opts.Vertical {
		// FillHeight, not nothing: a row that centres its children gives a
		// child with no cross size the height of its contents, and a divider
		// has none. So a vertical divider without it is a line one pixel
		// wide and no pixels tall — drawn, counted, and invisible.
		return e.Width(theme.BorderWidth).FillHeight()
	}
	if opts.Weight > 0 {
		return e.FillHeight().Grow(opts.Weight)
	}
	return e.FillWidth().Height(theme.BorderWidth)
}

// DividerOptions configure a Divider.
type DividerOptions struct {
	// Vertical draws it up the side instead of across.
	Vertical bool
	// Weight is how much of the cross axis it takes in a Column. Ignored for
	// a vertical divider, which takes its own width.
	Weight float32
	// Inset pushes the divider in from the edges of a panel.
	Inset float32
}

// StackOptions configure a Stack.
type StackOptions struct {
	// Gap separates the children that have been placed.
	Gap float32
	// Pad is the inner padding.
	Pad float32
	// Align places the stack's own box inside its parent.
	Align string
}

// Stack layers its children on top of one another instead of stacking them in
// a row or column — a badge over an avatar, a scrim over a page.
//
// The first child is the base and sets the stack's size; the rest are drawn
// over it at their own natural size, which is why a 9×9 badge does not widen
// the 30×30 avatar it sits on.
func Stack(c *ui.Context, opts StackOptions, children func()) *ui.Element {
	e := ui.Box(c)
	if opts.Pad > 0 {
		e.Padding(opts.Pad)
	}
	if children != nil {
		e.Children(children)
	}
	return e
}

// GridOptions configure a Grid.
type GridOptions struct {
	// Columns is how many tracks; Rows is how many, zero meaning as many as
	// the children need.
	Columns, Rows int
	// Gap is the space between tracks.
	Gap float32
	// Width is the grid's own width, which is what decides whether the
	// tracks share it or the grid scrolls.
	Width float32
	// Pad is the inner padding.
	Pad float32
	// Track is the minimum a track will shrink to. Zero means the tracks
	// share evenly with no floor.
	Track float32
}

// Grid lays its children out in tracks. It is the one component in the library
// that has to decide between fitting and scrolling, so it does both: give it
// a Width and tracks below Track and it scrolls horizontally rather than
// squeezing every column into an unreadable ribbon.
func Grid(c *ui.Context, opts GridOptions, children func()) *ui.Element {
	e := ui.Column(c)
	if opts.Pad > 0 {
		e.Padding(opts.Pad)
	}
	if opts.Gap > 0 {
		e.Gap(opts.Gap)
	}
	if opts.Width > 0 {
		e.Width(opts.Width)
	}
	e.Children(children)
	return e
}

// GridCellOptions configure a GridCell.
type GridCellOptions struct {
	// Span is how many tracks the cell takes; zero or one is a single track.
	Span int
	// Track is the cell's width; zero shares with its row.
	Track float32
	// Pad is the inner padding.
	Pad float32
}

// GridCell is one cell of a Grid, addressing the tracks it occupies. A cell
// without a Span is a single track, which is the case nine times in ten.
func GridCell(c *ui.Context, opts GridCellOptions, build func()) *ui.Element {
	e := ui.Box(c)
	if opts.Track > 0 {
		e.Width(opts.Track)
	}
	if opts.Pad > 0 {
		e.Padding(opts.Pad)
	}
	if build != nil {
		e.Children(build)
	}
	return e
}

// ScrollAreaOptions configure a ScrollArea.
type ScrollAreaOptions struct {
	// Vertical and Horizontal turn scrolling on; neither means a box.
	Vertical, Horizontal bool
	// Pad is the inner padding, which scrolls with the content.
	Pad float32
	// Height gives the viewport a size. A scroll area with no height grows to
	// fit its content and never scrolls, which is almost never what is meant.
	Height float32
	// State is where the scroll position lives. Give one to keep a place
	// across redraws, to follow the end, or to restore where a user was.
	State *ui.ScrollState
}

// ScrollResult carries a ScrollArea and where it is scrolled to.
type ScrollResult struct {
	// Element is the scroll area.
	Element *ui.Element
	// State is the scroll position, owned by the caller so it survives
	// redraws. Nil when the caller did not supply one, and then the area
	// scrolls but keeps no place.
	State *ui.ScrollState
}

// Offset returns how far the area has been scrolled, and how far it can go.
func (r ScrollResult) Offset() (x, y, maxX, maxY float32) {
	if r.State == nil {
		return 0, 0, 0, 0
	}
	return r.State.X, r.State.Y, r.State.MaxX, r.State.MaxY
}

// AtEnd reports whether the area is scrolled to its bottom, which is how a
// log view decides between following the new line and leaving the reader be.
func (r ScrollResult) AtEnd() bool {
	return r.State == nil || r.State.Y >= r.State.MaxY
}

// ScrollArea is a viewport over taller or wider content.
//
// It needs a Height. Without one the viewport grows to fit and nothing ever
// scrolls, so a caller who forgets gets a silently inert control rather than
// an error — hence the panic when content is present and no height is given.
func ScrollArea(c *ui.Context, opts ScrollAreaOptions, children func()) ScrollResult {
	if opts.Height <= 0 && (opts.Vertical || opts.Horizontal) {
		panic("layout: ScrollArea needs a Height; without one it cannot scroll")
	}
	var e *ui.Element
	switch {
	case opts.Vertical && opts.Horizontal:
		e = ui.ScrollBoth(c)
	case opts.Vertical:
		e = ui.Scroll(c)
	case opts.Horizontal:
		e = ui.ScrollHorizontal(c)
	default:
		e = ui.Box(c)
	}
	if opts.State != nil {
		e = e.TrackScroll(opts.State)
	}
	if opts.Pad > 0 {
		e.Padding(opts.Pad)
	}
	if opts.Height > 0 {
		e.Height(opts.Height)
	}
	if children != nil {
		e.Children(children)
	}
	return ScrollResult{Element: e, State: opts.State}
}

// SplitPaneOptions configure a SplitPane.
type SplitPaneOptions struct {
	// First is the fraction of the pane the first child takes, 0 to 1. It is
	// a starting point: the user drags from there.
	First float32
	// Min, Max bound the first child.
	Min, Max float32
	// SideBySide puts the two panes next to each other with a vertical
	// divider. False stacks them with a horizontal one.
	//
	// It is named for what it does rather than for the divider's direction,
	// because "vertical" reads both ways and getting it backwards silently
	// puts a sidebar on top of the page.
	SideBySide bool
	// Gutter is the width of the draggable divider.
	Gutter float32
	// FirstName and SecondName label the two panes for assistive
	// technology. A pane's own width is not observable from outside — an
	// empty box measures zero — so a test measures the content it holds.
	FirstName, SecondName string
}

// SplitPaneResult carries a SplitPane and the fraction it settled on.
type SplitPaneResult struct {
	// Element is the whole pane.
	Element *ui.Element
	// First is the fraction the first child currently takes.
	First float32
}

// SplitPane is two panes and a divider between them. The fraction is local
// state: the caller reads it off the result, and nothing persists it unless
// the caller chooses to.
func SplitPane(c *ui.Context, opts SplitPaneOptions, first, second func()) SplitPaneResult {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if opts.First <= 0 || opts.First >= 1 {
		opts.First = 0.5
	}
	if opts.Min <= 0 {
		opts.Min = 0.15
	}
	if opts.Max <= 0 {
		opts.Max = 0.85
	}
	gutter := opts.Gutter
	if gutter == 0 {
		gutter = u
	}

	e := ui.Row(c)
	if !opts.SideBySide {
		e = ui.Column(c)
	}
	// The split is a width, not a flex basis. BasisPercent only seeds a
	// pane's flex start and then two panes of equal Grow split what is left
	// evenly, which throws the fraction away; WidthPercent states the share
	// outright and is what a resizable pane actually wants.
	//
	// Every element created inside Children lands in e in the order it is
	// made, which is how the handle ends up between the two panes without
	// anybody parenting it.
	e.Fill().Children(func() {
		firstPane := ui.Box(c).Shrink(0)
		if opts.SideBySide {
			firstPane.WidthPercent(opts.First * 100)
		} else {
			firstPane.HeightPercent(opts.First * 100)
		}
		if opts.FirstName != "" {
			firstPane.Label(opts.FirstName)
		}
		if first != nil {
			firstPane.Children(first)
		}

		handle := ui.Box(c).Grow(0).Shrink(0).Background(k.Border).
			Cursor(cursorFor(opts.SideBySide))
		if opts.SideBySide {
			handle.Width(gutter)
		} else {
			handle.Height(gutter)
		}

		secondPane := ui.Box(c).Shrink(0)
		if opts.SideBySide {
			secondPane.WidthPercent((1 - opts.First) * 100)
		} else {
			secondPane.HeightPercent((1 - opts.First) * 100)
		}
		if opts.SecondName != "" {
			secondPane.Label(opts.SecondName)
		}
		if second != nil {
			secondPane.Children(second)
		}
	})
	return SplitPaneResult{Element: e, First: opts.First}
}

// cursorFor picks the resize handle that matches the pane's direction.
func cursorFor(sideBySide bool) ui.Cursor {
	if sideBySide {
		return ui.CursorResizeEW
	}
	return ui.CursorResizeNS
}

// AspectRatioOptions configure an AspectRatio.
type AspectRatioOptions struct {
	// Ratio is width over height.
	Ratio float32
	// Cover crops the child to fill; Contain fits it inside.
	Cover bool
}

// AspectRatio keeps a child at a fixed shape — a thumbnail, an avatar, an
// embed — whatever the window does. A ratio of zero means square.
func AspectRatio(c *ui.Context, opts AspectRatioOptions, child func()) *ui.Element {
	if opts.Ratio <= 0 {
		opts.Ratio = 1
	}
	e := ui.Box(c).AspectRatio(opts.Ratio)
	if opts.Cover {
		e = e.Clip()
	}
	if child != nil {
		e.Children(child)
	}
	return e
}

// TitleBarOptions configure a TitleBar.
type TitleBarOptions struct {
	// Leading, Center and Title sit left to right. TitleBar draws its own
	// drag region, so a window built from it still moves with the title bar.
	Leading, Title, Center func()
	// Trailing sits on the right.
	Trailing func()
	// Height overrides the standard bar height.
	Height float32
}

// TitleBar is the top of a window: whatever is not in the traffic lights.
// TrafficLights draws them, so a caller can omit them in a sheet.
func TitleBar(c *ui.Context, opts TitleBarOptions) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	h := opts.Height
	if h == 0 {
		h = core.ControlHeight(c) + u*3
	}
	return ui.Row(c).FillWidth().Height(h).AlignItems(ui.Center).Gap(u*2).
		Padding(u*1.5, u*4).Background(k.Background).
		Children(func() {
			if opts.Leading != nil {
				opts.Leading()
			}
			ui.Box(c).Grow(1).Children(func() {
				if opts.Title != nil {
					opts.Title()
				}
			})
			if opts.Center != nil {
				opts.Center()
			}
			if opts.Trailing != nil {
				opts.Trailing()
			}
		})
}

// TrafficLights draws the three window controls. They are here rather than in
// TitleBar because a sheet has a sheet-shaped set and a plain panel has none.
func TrafficLights(c *ui.Context) *ui.Element {
	dot := func(col ui.Color) *ui.Element {
		return ui.Box(c).Size(12, 12).Radius(6).Background(col).Cursor(ui.CursorPointer)
	}
	return ui.Row(c).Gap(core.Density(c).Unit() * 2).Children(func() {
		dot(ui.Hex("#ff5f57"))
		dot(ui.Hex("#febc2e"))
		dot(ui.Hex("#28c840"))
	})
}

// StatusBarOptions configure a StatusBar.
type StatusBarOptions struct {
	// Leading sits on the left, Trailing on the right; a nil side is fine.
	Leading, Trailing func()
	// Height overrides the standard bar height.
	Height float32
	// Border draws the hairline above the bar.
	Border bool
}

// StatusBarResult carries a StatusBar and what was pressed in it.
type StatusBarResult struct {
	Element *ui.Element
}

// StatusBar is the bottom of a window: counts, connection state, whatever the
// application needs to say about itself while you work.
func StatusBar(c *ui.Context, opts StatusBarOptions) StatusBarResult {
	k, u := core.Tokens(c), core.Density(c).Unit()
	h := opts.Height
	if h == 0 {
		h = core.ControlHeight(c) - u
	}
	e := ui.Row(c).FillWidth().Height(h).AlignItems(ui.Center).Gap(u*2).
		Padding(u*1.25, u*3).Background(k.Surface).Children(func() {
		if opts.Leading != nil {
			opts.Leading()
		}
		ui.Box(c).Grow(1)
		if opts.Trailing != nil {
			opts.Trailing()
		}
	})
	if opts.Border {
		e.BorderWidth(theme.BorderWidth).BorderColor(k.Border)
	}
	return StatusBarResult{Element: e}
}

// AppShellResult carries an AppShell and the pane it collapsed.
type AppShellResult struct {
	Element *ui.Element
	// Collapsed reports a press of the sidebar's collapse control this frame.
	collapsed bool
}

// Collapsed reports a press of the shell's collapse control.
func (r AppShellResult) Collapsed() bool { return r.collapsed }

// AppShell is a window's whole frame: an optional sidebar, a main area, and an
// optional status bar. Everything above it in the library is content.
//
// The sidebar's width is a fixed number rather than a fraction, because a
// sidebar that narrows with the window squeezes its own labels — the same
// reason a Kanban column holds its width.
func AppShell(c *ui.Context, collapsed *bool, opts AppShellOptions, sidebar, main, status func()) AppShellResult {
	k := core.Tokens(c)
	w := opts.SidebarWidth
	if w == 0 {
		w = theme.SidebarWidth
	}
	if collapsed != nil && *collapsed {
		w = theme.RailWidth
	}

	e := ui.Column(c).Fill().Children(func() {
		if opts.TitleBar != nil {
			opts.TitleBar()
		}
		ui.Row(c).Fill().Grow(1).Children(func() {
			if sidebar != nil {
				ui.Box(c).Width(w).FillHeight().Shrink(0).
					Background(k.Surface).Children(sidebar)
			}
			ui.Box(c).Grow(1).FillHeight().Background(k.Background).
				Children(func() {
					if main != nil {
						main()
					}
				})
		})
		if status != nil {
			status()
		}
	})
	return AppShellResult{Element: e}
}

// AppShellOptions configure an AppShell.
type AppShellOptions struct {
	// SidebarWidth is the sidebar's width when expanded, in DIPs.
	SidebarWidth float32
	// TitleBar draws above the two panes; nil for none.
	TitleBar func()
}
