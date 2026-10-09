package overlay

import (
	"github.com/egoist/mygo/ui"
)

// layer is the one place in this package that puts something above a window.
//
// Everything that floats — a dialog, an alert, a sheet — goes through it, so
// the three cannot drift apart on the questions that are easy to get subtly
// wrong: whether the scrim swallows clicks, whether Esc reaches the top layer,
// and whether an untaken layer still takes up room.
//
// It is MyGo's DialogBase with two differences this package needs. A
// non-modal sheet hangs at an edge of the window rather than centred, so it
// gets the backdrop's alignment; and the scrim only appears when the layer is
// modal, so a sheet can be read alongside the page it belongs to.
//
// body is handed the backdrop and the panel, both already made, so no
// component has to assemble the frame itself. It returns the panel: what a
// caller wants back is the thing the user is looking at, not the frame.
func layer(c *ui.Context, open *bool, modal bool, anchor ui.Align,
	body func(back, panel *ui.Element)) *ui.Element {

	if open == nil {
		panic("overlay: a layer needs a *bool to open and close")
	}
	if !*open {
		// An untaken layer draws nothing at all. Building an overlay and
		// hiding it still leaves it in the tree, and a closed dialog that
		// quietly narrows the page behind it shows up as "why does the board
		// look wrong when no dialog is open".
		return ui.Box(c)
	}

	var panel *ui.Element
	ui.Overlay(c, func() {
		back := barrier(ui.Box(c).Absolute().Left(0).Top(0).Right(0).Bottom(0).Center(), modal)
		if anchor != ui.Center {
			// An edge-anchored sheet is pushed to that edge rather than
			// centred: a drawer hung from the right must touch the right, or
			// it reads as a card that happened to float there.
			//
			// AlignItems, not Justify: the backdrop is a column, so its main
			// axis runs down the screen and Justify would move the sheet
			// vertically. AlignItems is the cross axis, which is the one that
			// decides which edge it hangs from.
			back = barrier(ui.Box(c).Absolute().Left(0).Top(0).Right(0).Bottom(0).
				AlignItems(anchor), modal)
		}
		// A press on the scrim closes the layer, and it is read here rather
		// than inside body, so that no component can forget it and end up
		// with a dialog you can only leave with the keyboard.
		//
		// Only a modal layer takes presses. A sheet that does not dim the
		// page is showing the page on purpose, so a press on it belongs to
		// the page, not to the layer.
		if modal && back.Clicked() {
			*open = false
		}
		back.Children(func() {
			if modal {
				// The scrim is what makes a modal layer modal: it takes the
				// clicks the page underneath would otherwise take, in the
				// wash every dimming layer wears, from [dim].
				back.Background(dim)
			}
			panel = ui.Box(c).Role(ui.RoleDialog)
			// Asking whether the panel was clicked is also what marks it
			// clickable, and marking it is the point: without this a press
			// inside the dialog reaches the backdrop and closes it, so the
			// only button you can reach in a form is the one that dismisses it.
			_ = panel.Clicked()
			panel.Children(func() { body(back, panel) })
		})
		// Escape closes the layer too, read here for the same reason as the
		// press above, but read after the body rather than before it, which
		// looks backwards and is not: MyGo hands an overlay key to one
		// registration and takes it out of the delivered set when that one
		// reads it, so of the two reads naming this same backdrop the first
		// to ask is the one that gets it. An alert reads Escape in its body
		// to learn which action Escape picked, and it has to be the one that
		// wins — a fallback reading first closed the alert and left Chosen
		// reporting nothing at all, which is the one thing an alert exists to
		// avoid.
		//
		// Only a modal layer asks. A sheet that does not dim the page is
		// showing the page on purpose, so a stray Escape must not throw away
		// what the user was reading beside it; and a layer that never asks
		// cannot take the Escape a modal layer below it is waiting for.
		if modal && back.OverlayShortcut(0, ui.KeyEscape) {
			*open = false
		}
	})
	return panel
}

// barrier is what a layer puts between itself and the window underneath: a
// wall when the layer is modal, and nothing at all when it is not.
//
// Not taking the window is not the same as doing nothing. The backdrop
// covers the whole window, and MyGo hands a press to the topmost element
// under the pointer before asking it anything, so a full-window backdrop that
// merely declines to be clickable still ends every press behind it: the page
// would be there to look at and dead to use. PassThrough is what lets the
// pointer through, and [Overlay] is where this package learned it.
func barrier(back *ui.Element, modal bool) *ui.Element {
	if modal {
		return back.Modal()
	}
	return back.PassThrough()
}

// anchored is the floating layer that belongs to a control rather than to the
// window: a popover, a tip, a hover card, a confirm.
//
// MyGo exports the non-modal path — the one that leaves the page clickable
// behind the card — as PopoverBase, and keeps the modal path to itself. This
// package needs both, because a menu that leaves the rest of the board live
// behind it and a confirm that does not are different controls, so the modal
// half is written here on the same two primitives.
func anchored(c *ui.Context, anchor *ui.Element, open *bool, modal bool,
	body func(panel *ui.Element)) *ui.Element {

	if open == nil {
		panic("overlay: an anchored layer needs a *bool to open and close")
	}
	if anchor == nil {
		panic("overlay: an anchored layer needs the element it hangs from")
	}
	if !modal {
		// PopoverBase already clamps the panel against the window's edges and
		// takes Esc. Reusing it is the reason a popover near the edge of the
		// screen is readable. It returns nil when shut, which the callers here
		// turn into an inert box so a closed layer never reaches the tree.
		if e := ui.PopoverBase(c, anchor, open, body); e != nil {
			return e
		}
		return ui.Box(c)
	}

	if !*open {
		return ui.Box(c)
	}

	var panel *ui.Element
	ui.Overlay(c, func() {
		back := ui.Box(c).Absolute().Left(0).Top(0).Right(0).Bottom(0).
			Background(ui.RGBA(0, 0, 0, 0.24))
		if back.Clicked() || back.OverlayShortcut(0, ui.KeyEscape) {
			*open = false
		}
		back.Children(func() {
			panel = ui.Box(c).Role(ui.RolePopup).AttachTo(anchor,
				ui.AnchorBottomLeft, ui.AnchorTopLeft)
			panel.Children(func() { body(panel) })
		})
	})
	return panel
}
