package overlay

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/input"
)

// PopoverOptions configure a Popover.
type PopoverOptions struct {
	// Title heads the panel; empty draws one of body alone.
	Title string
	// Subtitle is the second line under the title.
	Subtitle string
	// Body is the content.
	Body func()
	// Width is the panel's width. Zero takes the anchor's width, which is
	// what keeps a popover under a row reading as belonging to that row.
	Width float32
	// MaxWidth caps it; zero leaves it to the window.
	MaxWidth float32
	// Modal puts a scrim over the window and sends outside presses to it.
	// It is off by default, and that default is the point: these layers hang
	// off something on a page the person is still working in, so a scrim
	// across that page would say the work stopped — when what actually
	// happened is that one of its rows is being explained. Turn it on for the
	// rare popover that genuinely should interrupt.
	Modal bool
	// Label names the panel; empty uses the title.
	Label string
}

// Popover is a panel hung under an element: the extra an element owes its
// reader, such as a row's full address or a card's history.
//
// It does not dim the window unless Modal is set. That default is deliberate:
// a popover hangs off something on a page the person is working in, and a
// scrim across that page says the work stopped — when what actually happened
// is that one of its rows is being explained. A layer that quietly lets clicks
// through is the lesser risk here, because it is anchored: you can see what it
// is attached to.
//
// A press outside the panel and Escape both write false into the caller's
// *bool.
//
//	core.Use(c, core.Settings{})
//	row := ui.Box(c).Label("CB-2871").Children(func() { ui.Text(c, "CB-2871") })
//	Popover(c, row, &app.explaining, PopoverOptions{
//	    Modal:   true,
//	    Title:    "Opened by",
//	    Body:     func() { ui.Text(c, "Dana Reyes") },
//	})
func Popover(c *ui.Context, anchor *ui.Element, open *bool, opts PopoverOptions) *ui.Element {
	panel := anchored(c, anchor, open, opts.Modal, func(p *ui.Element) {
		popoverSurface(c, p, opts.Title, opts.Subtitle, opts.Label, true, opts.MaxWidth, func() {
			if opts.Body != nil {
				opts.Body()
			}
		})
	})
	if panel == nil || opts.Width <= 0 {
		return panel
	}
	panel.Width(opts.Width)
	return panel
}

// popoverSurface dresses host as a floating layer and fills it on the Panel's
// terms: the three popovers' one surface, which is where their hairline,
// radius and shadow all come from. compact is the difference between a note
// and a page — the compact surface a Popover wears, the room surface its wider
// relatives wear — and one place the three cannot drift apart.
func popoverSurface(c *ui.Context, host *ui.Element, title, subtitle, label string,
	compact bool, maxWidth float32, body func()) *ui.Element {

	return Panel(c, host, PanelOptions{
		Title:    title,
		Subtitle: subtitle,
		Compact:  compact,
		MaxWidth: maxWidth,
		Label:    label,
	}, body)
}

// PopoverRoomOptions configure a PopoverRoom. They are a Popover's, on the
// roomier surface.
type PopoverRoomOptions struct {
	// Title heads the panel, at the sheet size the room surface wears; empty
	// draws one of body alone.
	Title string
	// Subtitle is the second line under the title.
	Subtitle string
	// Body is the content.
	Body func()
	// Width is the panel's width. Zero takes the anchor's width, which is
	// what keeps a popover under a row reading as belonging to that row.
	Width float32
	// MaxWidth caps it; zero leaves it to the window.
	MaxWidth float32
	// Modal puts a scrim over the window and sends outside presses to it, on
	// the terms PopoverOptions.Modal spells out. It is off by default, and
	// that default is the point.
	Modal bool
	// Label names the panel; empty uses the title.
	Label string
}

// PopoverRoom is a Popover with room: the full panel's padding and radius
// rather than the compact one, for content that is more than a line — a
// preview with its caption, a description over an action row, a card's
// history that needs the lines.
//
// A Popover wears the compact surface because what it usually owes its
// reader is one or two lines, and the smaller padding is what keeps a
// two-line explanation from reading as a dialog. PopoverRoom is for the
// popover that is a page rather than a note: the room is the point of it,
// and drawing it compact would spend its whole budget on padding. It hangs
// off its anchor, clamps against the window's edges and closes on a press
// outside and on Escape, on exactly the terms a Popover does.
//
// The difference from a Popover is the surface and nothing else: the anchor,
// the modal and the label are its. Where a caller wants the whole width
// rather than the content's, [PopoverWide] is the third of the three.
//
//	core.Use(c, core.Settings{})
//	PopoverRoom(c, row, &app.expanding, PopoverRoomOptions{
//	    Title: "CB-2871",
//	    Body:  func() { preview(); ui.Text(c, "Opened by Dana Reyes") },
//	})
func PopoverRoom(c *ui.Context, anchor *ui.Element, open *bool, opts PopoverRoomOptions) *ui.Element {
	panel := anchored(c, anchor, open, opts.Modal, func(p *ui.Element) {
		popoverSurface(c, p, opts.Title, opts.Subtitle, opts.Label, false, opts.MaxWidth, func() {
			if opts.Body != nil {
				opts.Body()
			}
		})
	})
	if panel == nil || opts.Width <= 0 {
		return panel
	}
	panel.Width(opts.Width)
	return panel
}

// PopoverWideOptions configure a PopoverWide.
type PopoverWideOptions struct {
	// Title heads the panel, at the sheet size the room surface wears; empty
	// draws one of body alone.
	Title string
	// Subtitle is the second line under the title.
	Subtitle string
	// Body is the content, laid out across the full width the panel takes.
	Body func()
	// MaxWidth caps the width, for a window wider than the content deserves;
	// zero leaves it to the parent.
	MaxWidth float32
	// Modal puts a scrim over the window and sends outside presses to it, on
	// the terms PopoverOptions.Modal spells out. It is off by default, and
	// that default is the point.
	Modal bool
	// Label names the panel; empty uses the title.
	Label string
}

// PopoverWide is a popover that takes the whole width of its parent: the room
// surface of a PopoverRoom with a FillWidth on top, for the content that is a
// form or a long list rather than a note — a settings block hung off a header,
// a list of rows that wraps.
//
// Where a Popover fits itself to its anchor and a PopoverRoom fits itself to
// its content, a PopoverWide asks for the full width and lets the content run
// across it: a form field that is the width of its column reads as a field,
// and the same form crammed under a button reads as a card. It shares the
// room surface with PopoverRoom — the two are one panel at two widths, and
// the width is the whole of their difference, which is why they share the
// surface helper rather than each keeping their own.
//
//	core.Use(c, core.Settings{})
//	PopoverWide(c, header, &app.filtering, PopoverWideOptions{
//	    Title: "Filter",
//	    Body:  func() { filterFields() },
//	})
func PopoverWide(c *ui.Context, anchor *ui.Element, open *bool, opts PopoverWideOptions) *ui.Element {
	panel := anchored(c, anchor, open, opts.Modal, func(p *ui.Element) {
		popoverSurface(c, p, opts.Title, opts.Subtitle, opts.Label, false, opts.MaxWidth, func() {
			if opts.Body != nil {
				opts.Body()
			}
		})
	})
	if panel == nil {
		return panel
	}
	panel.FillWidth()
	return panel
}

// Choice is how a Popconfirm ended.
type Choice int

const (
	// NoChoice is the frame in which nobody answered.
	NoChoice Choice = iota
	// Confirmed is the affirmative button being pressed.
	Confirmed
	// Cancelled is the dismissive button — or Escape, or a press outside,
	// which mean the same thing here.
	Cancelled
)

// PopconfirmOptions configure a Popconfirm.
type PopconfirmOptions struct {
	// Title is the question. It is required: a confirm with nothing to
	// confirm would be a button the person is being asked to trust.
	Title string
	// Body is one sentence saying what follows from the answer.
	Body string
	// Confirm names the affirmative button; required.
	Confirm string
	// Cancel names the dismissive button; required.
	Cancel string
	// Destructive draws Confirm in the danger colour, for an answer that
	// destroys rather than starts something.
	Destructive bool
	// Modal puts a scrim over the window and sends outside presses to it.
	// It is off by default, and that default is the point: this layer hangs
	// off something on a page the person is still working in, so a scrim
	// across that page would say the work stopped — when what actually
	// happened is that one of its rows is being explained.
	Modal bool
}

// PopconfirmResult carries a Popconfirm and what the user answered.
type PopconfirmResult struct {
	// Element is the panel, or nil while it is closed.
	Element *ui.Element
	choice  Choice
}

// Chosen returns which answer the person gave.
func (r PopconfirmResult) Chosen() Choice { return r.choice }

// Popconfirm asks a small question where it was asked: is that the row you
// meant to delete? It hangs off the element it is about rather than taking
// the window, so the thing being confirmed stays on screen while it is.
//
// It does not dim the window unless Modal is set, and as everywhere here a press outside
// the panel and Escape write false into the caller's *bool and report
// Cancelled.
//
//	core.Use(c, core.Settings{})
//	Popconfirm(c, del, &app.asking, PopconfirmOptions{
//	    Title:      "Delete CB-2871?",
//	    Body:       "The callback and its history go with it.",
//	    Confirm:    "Delete",
//	    Cancel:     "Keep",
//	    Destructive: true,
//	    Modal:       true,
//	})
func Popconfirm(c *ui.Context, anchor *ui.Element, open *bool, opts PopconfirmOptions) PopconfirmResult {
	if opts.Title == "" {
		panic("overlay: Popconfirm needs a Title; a question with no question in it is not a confirmation")
	}
	if opts.Confirm == "" || opts.Cancel == "" {
		panic("overlay: Popconfirm needs both a Confirm and a Cancel label")
	}

	res := PopconfirmResult{}
	// Whether the layer was open as it was asked for: a confirm that closed
	// without a button was dismissed, and being dismissed is an answer.
	was := *open
	res.Element = anchored(c, anchor, open, opts.Modal, func(p *ui.Element) {
		Panel(c, p, PanelOptions{
			Title:    opts.Title,
			Subtitle: opts.Body,
			Compact:  true,
		}, func() {
			u := core.Density(c).Unit()
			ui.Row(c).FillWidth().Gap(u).Justify(ui.End).Margin(u, 0, 0, 0).
				Children(func() {
					cancel := input.Button(c, opts.Cancel, input.ButtonOptions{Label: opts.Cancel})
					if cancel.Clicked() {
						res.choice = Cancelled
					}
					confirm := input.Button(c, opts.Confirm, input.ButtonOptions{
						Label:   opts.Confirm,
						Primary: !opts.Destructive,
						Danger:  opts.Destructive,
					})
					if confirm.Clicked() {
						res.choice = Confirmed
					}
				})
		})
	})
	if was && !*open && res.choice == NoChoice {
		// Escape, or a press outside: both mean "not now", which is what
		// Cancel means.
		res.choice = Cancelled
	}
	return res
}
