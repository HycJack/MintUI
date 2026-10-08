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
		Panel(c, p, PanelOptions{
			Title:    opts.Title,
			Subtitle: opts.Subtitle,
			Compact:  true,
			MaxWidth: opts.MaxWidth,
			Label:    opts.Label,
		}, func() {
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
