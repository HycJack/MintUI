package layout

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// Panel is the library's one answer to "what does a floating layer look
// like": the radius, the hairline, the shadow, the padding and the title bar.
//
// It lives in layout rather than in overlay so that the two sides of the
// floating relationship can both reach it. An overlay contains buttons, so
// overlay imports input; if Panel lived there, input could not use it to draw
// the menu of a Select without importing the whole overlay package back. One
// layer down is the only place both can stand.

// PanelOptions configure a Panel.
type PanelOptions struct {
	// Title heads the panel. Empty draws no header at all, which is what a
	// tip and a bare popover want.
	Title string
	// Subtitle is the second line under the title: what the panel is for,
	// where the title only names it.
	Subtitle string
	// TitleSize is the title's size. Zero gives a dialog the sheet size and
	// everything smaller the row size, which is the difference between a
	// heading and a label.
	TitleSize float32
	// Compact is the smaller panel: tighter padding and the control radius.
	// A popover, a tip and a confirm wear it; a dialog or a sheet does not.
	Compact bool
	// Round maps the panel's radius onto its four corners, in the order
	// top-left, top-right, bottom-right, bottom-left.
	//
	// It exists for the shapes that are square where they meet the window:
	// a drawer hung from an edge, a split button's two halves. Everywhere
	// else it is nil and the radius family applies as it is, which is the
	// point of having one family.
	Round func(radius float32) (topLeft, topRight, bottomRight, bottomLeft float32)
	// Pad overrides the inner padding.
	Pad float32
	// MaxWidth caps the panel, which is what keeps a one-line tip from
	// stretching across a wide window.
	MaxWidth float32
	// Rule draws a hairline under the header, so a title and the form under
	// it do not run into one line of text.
	Rule bool
	// Role names the panel for assistive technology. RoleAuto — the zero
	// value — leaves whatever the host already says, which is how the
	// DialogBase path keeps a panel a dialog and an AlertDialog an alert.
	Role ui.Role
	// Label names the panel; empty uses the title, and an untitled panel is
	// left unnamed, as a tip is: it describes the control it hangs on
	// rather than standing on its own.
	Label string
}

// Panel gives host the look of a floating layer and fills it: the radius, the
// hairline, the shadow, the padding, and the title bar when there is one.
//
// It is the only place in the library that decides any of those, so a change
// here moves every dialog, sheet, popover, tip and confirm together. host is
// an element the component making the layer already has — a panel handed out
// by ui.DialogBase or ui.PopoverBase, or a fresh box — and it comes back
// styled, for the caller to place.
func Panel(c *ui.Context, host *ui.Element, opts PanelOptions, body func()) *ui.Element {
	if host == nil {
		panic("overlay: Panel needs the element it dresses; a layer with no element has no shape to give")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	pad := opts.Pad
	if pad == 0 {
		pad = u * 2.5
		if opts.Compact {
			pad = u * 1.5
		}
	}
	radius := theme.PanelRadius
	if opts.Compact {
		radius = theme.ControlRadius
	}
	titleSize := opts.TitleSize
	if titleSize == 0 {
		titleSize = theme.SheetSize
		if opts.Compact {
			titleSize = theme.RowSize
		}
	}

	host.Column().Padding(pad).
		Background(k.Background).
		TextColor(k.Text).
		Shadow(0, u*2.5, u*7.5, 0, ui.RGBA(0, 0, 0, shadowAlpha(c)))
	if opts.Round != nil {
		tl, tr, br, bl := opts.Round(radius)
		host.Radius(tl, tr, br, bl)
	} else {
		host.Radius(radius)
	}
	host.BorderWidth(theme.BorderWidth).BorderColor(k.Border)
	if opts.MaxWidth > 0 {
		host.MaxWidth(opts.MaxWidth)
	}
	if opts.Role != ui.RoleAuto {
		host.Role(opts.Role)
	}
	if name := panelName(opts); name != "" {
		host.Label(name)
	}

	host.Children(func() {
		if opts.Title != "" || opts.Subtitle != "" {
			header := ui.Column(c).FillWidth().Label(panelName(opts))
			if opts.Title != "" {
				header.Children(func() {
					ui.Text(c, opts.Title).FontSize(core.FontSize(c, titleSize)).
						Bold().TextColor(k.Text).Role(ui.RoleHeading)
				})
			}
			if opts.Subtitle != "" {
				header.Children(func() {
					ui.Text(c, opts.Subtitle).FontSize(core.FontSize(c, theme.MetaSize)).TextColor(k.TextMuted)
				})
			}
			if opts.Rule {
				header.Children(func() {
					// The rule is the header's own child, so it stops at the
					// panel's padding: one run to the panel's corners would
					// sit under its radius and poke out of the curve.
					Divider(c, DividerOptions{})
				})
			}
		}
		if body != nil {
			body()
		}
	})
	return host
}

// panelName is what the panel is called to assistive technology and to a test
// looking for it.
func panelName(opts PanelOptions) string {
	if opts.Label != "" {
		return opts.Label
	}
	return opts.Title
}

// shadowAlpha is how dark a panel's shadow is. A dark window already sits low
// against its own page, so the same shadow reads as a heavier box there than
// in the light one; a deeper alpha in dark keeps the panel equally lifted off
// the page in both appearances instead of only in one.
func shadowAlpha(c *ui.Context) float32 {
	if core.IsDark(c) {
		return 0.34
	}
	return 0.18
}

// layer lays a panel over the window: a backdrop covering it that dismisses
// the layer, and the panel, hung from align — centred, or against one edge.
//
// modal says the window behind waits. With it the backdrop is MyGo's modal
// backdrop, which is what makes the page behind inert and the Esc key the
// layer's. Without it the backdrop is drawn here, which is what a drawer hung
// from an edge needs: it centres what is inside it and then moves the whole
// thing to the side, and MyGo's centred backdrop cannot be un-centred.
//
// open is the caller's, and every way out writes to it.
func layer(c *ui.Context, open *bool, modal bool, align ui.Align, build func(back, panel *ui.Element)) *ui.Element {
	if open == nil {
		panic("overlay: a layer needs the *bool it opens and closes")
	}
	if !*open {
		// MyGo lays a modal layer over the window whether or not it is
		// showing, so asking for a closed one still costs a frame's layout.
		// Callers guard the call themselves; Dialog says the same thing.
		return nil
	}

	if modal {
		return ui.DialogBase(c, open, func(back, panel *ui.Element) {
			// DialogBase centres the panel, and a Box lays its one child out
			// down the page — so Justify on it would move the panel along the
			// vertical, which is not what hanging a sheet from an edge means.
			// Row first, then Justify along it.
			back.Row().Justify(align)
			build(back, panel)
		})
	}

	var panel *ui.Element
	ui.Overlay(c, func() {
		back := ui.Row(c).Absolute().Left(0).Top(0).Right(0).Bottom(0).
			Justify(align).AlignItems(ui.Center)
		// Clicked is what marks the backdrop as taking presses, so a press
		// that missed the panel lands here rather than on the page under it.
		if back.Clicked() {
			*open = false
		}
		back.Children(func() {
			panel = ui.Box(c)
			swallow(panel)
			panel.Children(func() { build(back, panel) })
		})
	})
	return panel
}

// anchored hangs a panel off anchor, under it, moving it above when there is
// no room below.
//
// modal says a scrim covers the window while the panel shows, as layer does.
// Without it the panel is MyGo's popover: it goes on a press outside itself
// and lets that press through to what is under the pointer, which is what a
// menu on a page has to do.
func anchored(c *ui.Context, anchor *ui.Element, open *bool, modal bool, build func(panel *ui.Element)) *ui.Element {
	if open == nil {
		panic("overlay: an anchored layer needs the *bool it opens and closes")
	}
	if anchor == nil {
		panic("overlay: an anchored layer needs the element it hangs on")
	}
	if !modal {
		return ui.PopoverBase(c, anchor, open, build)
	}
	if !*open {
		return nil
	}

	var panel *ui.Element
	ui.Overlay(c, func() {
		back := ui.Box(c).Absolute().Left(0).Top(0).Right(0).Bottom(0).Modal()
		if back.Clicked() {
			*open = false
		}
		panel = ui.Box(c).Role(ui.RolePopup).
			AttachTo(anchor, ui.AnchorBottomLeft, ui.AnchorTopLeft).
			Margin(core.Density(c).Unit(), 0, 0, 0)
		swallow(panel)
		panel.Children(func() { build(panel) })
		// Escape is read from the panel rather than from the scrim, as
		// MyGo's own popover does: an overlay hands its keys to what was
		// built last, and the panel is the layer a person is looking at.
		if panel.OverlayShortcut(0, ui.KeyEscape) {
			*open = false
		}
	})
	return panel
}

// swallow makes a panel take the presses that land on it, so a press there
// does not fall through to the window behind.
//
// Clicked is the exported way to mark an element as taking presses. Calling
// it and dropping the answer is deliberate: a press that reached the panel
// found nothing interactive inside it, because the press is given to the
// innermost control under the pointer — so the panel only ever reports the
// presses meant for its own surface, and swallowing them is all it has to do.
func swallow(e *ui.Element) { e.Clicked() }
