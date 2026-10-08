package overlay

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// dim is the wash a modal layer lays over the window: one flat darkening,
// rather than a token, because there is only one right answer for "how much
// light does a page lose when something asks to be dealt with first".
//
// Every layer that dims — a dialog, an alert, and an [Overlay] drawn on its
// own — takes its colour from this one place, so the three cannot drift
// apart on a question that has one answer.
var dim = ui.RGBA(0, 0, 0, 0.4)

// OverlayOptions configure an Overlay.
type OverlayOptions struct {
	// Clickable makes the backdrop take the presses that land on it and
	// report them through Clicked, as a dialog's scrim does. A backdrop that
	// only dims the page under a layer the caller dismisses by other means
	// leaves it off, and the presses go on to the window behind.
	Clickable bool
}

// OverlayResult carries an Overlay and the press on its backdrop.
type OverlayResult struct {
	// Element is the backdrop.
	Element *ui.Element
	pressed bool
}

// Clicked reports the backdrop itself being pressed — the press that missed
// everything on top of it — the way a dialog reads its scrim, for the
// "pressed outside, close" case. It is false on a frame in which nothing
// landed on it, and always false while Clickable is off, because a backdrop
// that does not take presses cannot be pressed.
//
// Read it inside the view, as every report in this library is: the view runs
// up to three times per frame and the settled pass has no press.
func (r OverlayResult) Clicked() bool { return r.pressed }

// Overlay is the darkening layer under a floating one: the same flat wash a
// modal dialog lays over the window, drawn on its own.
//
// A Dialog or a Drawer already carries its backdrop — the scrim is what makes
// them modal — and building one of those on top of an Overlay would draw the
// wash twice. Overlay is for the layer the library does not own: a panel the
// caller lays over the window itself, that wants the window dimmed behind it
// and wants to know the press that lands on the wash rather than on the
// panel, because that press is the one that says "close".
//
//	core.Use(c, core.Settings{})
//	inspecting := Overlay(c, OverlayOptions{Clickable: true})
//	if inspecting.Clicked() {
//	    app.inspecting = false
//	}
func Overlay(c *ui.Context, opts OverlayOptions) OverlayResult {
	var res OverlayResult
	ui.Overlay(c, func() {
		back := ui.Box(c).Absolute().Left(0).Top(0).Right(0).Bottom(0).
			Background(dim)
		switch {
		case opts.Clickable:
			res.pressed = back.Clicked()
		default:
			// A backdrop that only dims lets the pointer reach the window
			// under it; PassThrough is the letting through for an element
			// that covers the whole window, and without it the page behind
			// a non-clickable wash would be dark but inert.
			back.PassThrough()
		}
		res.Element = back
	})
	return res
}

// OverlaySide is which side of its anchor point an OverlayAnchored hangs
// from: the edge of the surface that faces the point.
type OverlaySide int

const (
	// OverlayBelow puts the surface's top edge a step under the point, with
	// the caret, when there is one, on that top edge.
	OverlayBelow OverlaySide = iota
	// OverlayAbove puts the surface's bottom edge a step over the point.
	OverlayAbove
	// OverlayStart puts the surface's right edge a step to the left of the
	// point.
	OverlayStart
	// OverlayEnd puts the surface's left edge a step to the right of the
	// point.
	OverlayEnd
)

// OverlayAnchoredOptions configure an OverlayAnchored.
type OverlayAnchoredOptions struct {
	// X and Y are the anchor's place in window coordinates: the point the
	// surface is put beside, and the one the caret points at when there is
	// one.
	X, Y float32
	// Side is which side of that point the surface hangs from, and is
	// required: a surface has to know which of its edges faces the point, or
	// it hangs from nothing.
	Side OverlaySide
	// Title heads the surface, at the row size a floating layer wears; empty
	// draws one of body alone.
	Title string
	// Subtitle is the second line under the title.
	Subtitle string
	// Body is the content, and is required: a surface with nothing in it is a
	// shadow with nothing under it.
	Body func()
	// Caret draws the small arrow between the surface and its point, on the
	// edge that faces it, so the surface reads as pointing at the thing it is
	// about rather than floating near it.
	Caret bool
	// Width is the surface's width; zero lets it fit its content.
	Width float32
	// MaxWidth caps it; zero leaves it to the window.
	MaxWidth float32
	// Label names the surface; empty uses the title.
	Label string
}

// OverlayAnchored is a floating surface placed at the caller's own
// coordinates: the look of a popover — the background, the radius, the
// hairline, the shadow — hung not off an element but at a point.
//
// A Popover hangs off its anchor, and MyGo clamps it against the window's
// edges from there. OverlayAnchored is for the place a caller already knows
// and has no element to hang from — over a chart mark, beside a map pin, at
// the end of a drag — where the coordinates are the information, and the
// side the surface faces is decided by the caller rather than by what room
// there happens to be. The Side decides which edge faces the point, and the
// Caret, when there is one, sits on that edge in the gap between the surface
// and the point.
//
// It takes no *bool: the caller owns when it is drawn, for the reason the
// caller owns the coordinates — both are the information, and a component
// that kept either would be a component the caller could not move.
//
//	core.Use(c, core.Settings{})
//	OverlayAnchored(c, OverlayAnchoredOptions{
//	    X: 340, Y: 210, Side: OverlayBelow, Caret: true,
//	    Title: "North seam",
//	    Body:  func() { ui.Text(c, "Split 40mm, photographed") },
//	})
func OverlayAnchored(c *ui.Context, opts OverlayAnchoredOptions) *ui.Element {
	if opts.Side < OverlayBelow || opts.Side > OverlayEnd {
		panic("overlay: OverlayAnchored needs a Side of below, above, start or end; a surface " +
			"that does not know which edge faces its point hangs from nothing")
	}
	if opts.Body == nil {
		panic("overlay: OverlayAnchored needs a Body; a surface with nothing in it is a shadow " +
			"with nothing under it")
	}
	k := core.Tokens(c)
	u := core.Density(c).Unit()
	winW, winH := c.Size()
	gap := u

	var panel *ui.Element
	ui.Overlay(c, func() {
		// The caret is built before the panel so the panel, which is opaque,
		// covers the caret's base: the arrow meets the surface at the
		// hairline rather than sitting on top of it.
		if opts.Caret {
			caret := ui.Box(c).Absolute().Size(caretWidth, caretHeight).
				Background(k.Background)
			caret.Draw(func(p *ui.Painter, r ui.Rect) {
				path := caretPath(r, opts.Side)
				p.FillPath(&path, k.Background)
				p.StrokePath(&path, theme.BorderWidth, k.Border)
			})
			switch opts.Side {
			case OverlayBelow:
				caret.Left(opts.X).Top(opts.Y + gap - caretHeight)
			case OverlayAbove:
				caret.Left(opts.X).Top(opts.Y - gap)
			case OverlayStart:
				caret.Left(opts.X - gap).Top(opts.Y)
			default: // OverlayEnd
				caret.Left(opts.X + gap - caretWidth).Top(opts.Y)
			}
		}

		host := ui.Box(c).Absolute()
		switch opts.Side {
		case OverlayBelow:
			host.Left(opts.X).Top(opts.Y + gap)
		case OverlayAbove:
			host.Left(opts.X).Bottom(winH - (opts.Y - gap))
		case OverlayStart:
			host.Top(opts.Y).Right(winW - (opts.X - gap))
		default: // OverlayEnd
			host.Left(opts.X + gap).Top(opts.Y)
		}
		panel = Panel(c, host, PanelOptions{
			Title:    opts.Title,
			Subtitle: opts.Subtitle,
			Compact:  true,
			MaxWidth: opts.MaxWidth,
			Label:    opts.Label,
		}, opts.Body)
		if opts.Width > 0 {
			panel.Width(opts.Width)
		}
	})
	return panel
}

// caretWidth and caretHeight are the arrow's size: wide enough to read as an
// arrow rather than a notch, short enough to sit in the gap the surface keeps
// from its point without crowding either.
const (
	caretWidth  float32 = 12
	caretHeight float32 = 8
)

// caretPath is the triangle the caret wears, with the apex on the edge that
// faces the point: the arrow says which way the surface is pointing, and the
// side it was hung from is the direction it points.
func caretPath(r ui.Rect, side OverlaySide) ui.Path {
	var path ui.Path
	switch side {
	case OverlayBelow: // apex up, towards the point above
		path.MoveTo(r.X+r.W/2, r.Y).
			LineTo(r.X+r.W, r.Y+r.H).
			LineTo(r.X, r.Y+r.H)
	case OverlayAbove: // apex down, towards the point below
		path.MoveTo(r.X+r.W/2, r.Y+r.H).
			LineTo(r.X+r.W, r.Y).
			LineTo(r.X, r.Y)
	case OverlayStart: // apex right, towards the point to the right
		path.MoveTo(r.X+r.W, r.Y+r.H/2).
			LineTo(r.X, r.Y+r.H).
			LineTo(r.X, r.Y)
	default: // OverlayEnd: apex left, towards the point to the left
		path.MoveTo(r.X, r.Y+r.H/2).
			LineTo(r.X+r.W, r.Y+r.H).
			LineTo(r.X+r.W, r.Y)
	}
	path.Close()
	return path
}
