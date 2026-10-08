package feedback

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// LoadingOverlayOptions configure a LoadingOverlay.
type LoadingOverlayOptions struct {
	// Message says what is being waited for. It is required: a spinner on a
	// scrim with no words leaves the reader choosing between "working" and
	// "hung", which is not a distinction they should have to make.
	Message string
	// Loader picks the shape and the colour. The zero value is a
	// BarsLoader at the standard size in the accent, which is what most
	// callers want.
	Loader LoaderOptions
	// Dismissable adds a Cancel button under the message and reports its
	// press. It is off by default: an overlay that cancels is a dialog, and
	// a dialog has a scrim the user can see they are now inside one.
	Dismissable bool
	// CancelLabel names the cancel button; empty takes the library's
	// "Cancel".
	CancelLabel string
	// Scrim is how far the overlay darkens what is under it, 0 to 1. Zero
	// takes a standard scrim, which is dark enough to say "not this" and
	// light enough that the content under it is still recognisable.
	Scrim float32
	// Radius rounds the panel; zero takes the card radius.
	Radius float32
}

// LoadingOverlayResult carries a LoadingOverlay and the press of its cancel.
type LoadingOverlayResult struct {
	// Element is the overlay, which wraps the content it covers.
	Element   *ui.Element
	dismissed bool
}

// Dismissed reports the cancel button.
func (r LoadingOverlayResult) Dismissed() bool { return r.dismissed }

// LoadingOverlay puts a scrim and a spinner over content that is still
// arriving, so the content stays where it is going to be instead of being
// replaced by a spinner and then by the content.
//
// The child is the base of the stack and sets the overlay's size. A caller
// with nothing to show underneath may pass nil, and then the overlay covers
// whatever its parent gives it — which is right for a panel that is empty
// because it is loading, and wrong for a page that has content behind it.
func LoadingOverlay(c *ui.Context, opts LoadingOverlayOptions, child func()) LoadingOverlayResult {
	if opts.Message == "" {
		panic("feedback: LoadingOverlay needs a Message; a spinner on a scrim " +
			"leaves the reader guessing between working and hung")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	scrim := clamp01(opts.Scrim)
	if scrim == 0 {
		scrim = 0.55
	}
	rad := opts.Radius
	if rad <= 0 {
		rad = theme.CardRadius
	}
	cancel := opts.CancelLabel
	if cancel == "" {
		cancel = core.Msg(c, "feedback.loadingOverlay.cancel", "Cancel")
	}
	var r LoadingOverlayResult

	e := layout.Stack(c, layout.StackOptions{}, func() {
		if child != nil {
			child()
		} else {
			ui.Box(c).Fill()
		}
		// Absolute over the base rather than a second child, so the overlay
		// covers the content without taking any room from it: an overlay
		// that reflowed what it was covering would change the layout of the
		// thing it is only meant to be waiting for.
		ui.Box(c).Absolute().Fill().Background(k.Fill.Alpha(scrim)).
			Children(func() {
				ui.Box(c).Grow(1).Center().Children(func() {
					ui.Column(c).AlignItems(ui.Center).Gap(u*2.5).
						Background(k.Background).Radius(rad).
						Padding(u*5, u*6.5, u*5, u*6.5).Label(opts.Message).
						Children(func() {
							BarsLoader(c, opts.Loader)
							ui.Text(c, opts.Message).TextColor(k.Text).
								FontSize(core.FontSize(c, theme.RowSize))
							if opts.Dismissable {
								btn := ui.Button(c, "").Height(core.ControlHeight(c)).
									Radius(theme.PillRadius).Padding(0, u*4, 0, u*4).
									Background(k.Surface).TextColor(k.Text).
									Border(theme.BorderWidth, k.Border).
									Label(cancel).Tooltip(cancel)
								btn.Children(func() {
									ui.Text(c, cancel).TextColor(k.Text).
										FontSize(core.FontSize(c, theme.RowSize))
								})
								r.dismissed = btn.Clicked()
							}
						})
				})
			})
	})
	r.Element = e
	return r
}
