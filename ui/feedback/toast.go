package feedback

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// ToastItem is one message sitting in a ToastStore, waiting to be drawn by a
// Toaster.
//
// It is named ToastItem and not Toast because the package already has a
// Toast: the foot-of-window bar that a caller shows and lets expire. The two
// are different promises — that bar says one thing and withdraws itself, an
// item is a row in a list the caller owns — and two components named the
// same would make every doc comment in this file lie.
type ToastItem struct {
	// ID is assigned by ShowToast, in the order toasts were shown. The
	// caller keeps the returned ID rather than reading this back: a toast
	// is the caller's to dismiss, by the number ShowToast gave it.
	ID int
	// Title is what happened, in the present tense. It is required: a toast
	// with nothing to say dismisses itself before it is drawn, which is
	// what a missing title would amount to.
	Title string
	// Body is the line under it. Empty is fine for a toast with nothing to
	// add, the way a Notification's is.
	Body string
	// Severity marks the item with a dot at its leading edge. Zero is
	// core.Neutral, which is right for most of what an application has to
	// say; a toast that is only ever success and danger is a report, and
	// the dot is what says so at a glance.
	Severity core.Severity
}

// ToastStore is the caller's list of toasts.
//
// The library's components hold nothing, and a toast's whole life is state:
// when it was shown, whether it is still up, when it was dismissed. So the
// state lives here, with the caller, rather than in a Toaster: the caller
// holds a store (a package variable, a field on the window's state), shows
// into it from wherever the events happen, and draws a Toaster each frame
// that reads the store and reports what was pressed in it. The store is not
// safe for concurrent use, for the same reason nothing in this library is:
// a frame is the unit, and everything in it happens in one goroutine.
type ToastStore struct {
	toasts []ToastItem
	next   int
}

// ShowToast appends t to the store and returns the ID it was given, which
// the caller keeps for DismissToast.
//
// It assigns the ID rather than taking the caller's, because a caller that
// numbers its toasts by hand is one press away from dismissing the wrong
// one: two events in one frame, two numbers written down, and a dismiss
// aimed at a toast that is no longer the latest.
func ShowToast(store *ToastStore, t ToastItem) int {
	if store == nil {
		panic("feedback: ShowToast needs a store; a toast with nowhere to " +
			"land goes nowhere")
	}
	if t.Title == "" {
		panic("feedback: a toast needs a Title; a toast with nothing to say " +
			"is a button that closes nothing")
	}
	store.next++
	t.ID = store.next
	store.toasts = append(store.toasts, t)
	return t.ID
}

// DismissToast takes the toast with ID off the store, reporting whether
// there was one: true when it took a toast, false when the ID was not in
// the store. A false report is the caller's own news — a dismiss that
// arrives after the toast is already gone — and it is better delivered than
// turned into a panic, which would be an app crashing because it was a
// frame too late.
//
// A nil store reports false rather than panicking: DismissToast is called
// in reaction to a press, and a press must never take the window down.
func DismissToast(store *ToastStore, id int) bool {
	if store == nil {
		return false
	}
	for i := range store.toasts {
		if store.toasts[i].ID == id {
			store.toasts = append(store.toasts[:i], store.toasts[i+1:]...)
			return true
		}
	}
	return false
}

// ToasterOptions configure a Toaster.
type ToasterOptions struct {
	// Width bounds each toast, in DIPs. Zero takes the standard width,
	// which is wide enough for a title and a short body without crowding
	// the window.
	Width float32
	// Center places the stack along the top edge, in the middle of the
	// window, rather than in its top-right corner. Off by default: the
	// corner is where a toast goes so that it does not sit on top of what
	// the window is about, and a caller that wants the middle is making a
	// deliberate choice worth a deliberate flag.
	Center bool
}

// ToasterResult carries a Toaster and what was pressed in it.
type ToasterResult struct {
	// Element is the whole toaster, which fills whatever it is placed in.
	Element *ui.Element
	// dismissed is the ID of the toast whose close was pressed this frame,
	// or -1 when none was.
	dismissed int
}

// Dismissed reports the ID of the toast whose close button was pressed this
// frame, or -1 when none was. Only the first press of a frame is reported,
// for the reason the NotificationCenter's is. The caller acts on it with
// DismissToast, in the frame after the press, the way the library's other
// press results are acted on.
func (r ToasterResult) Dismissed() int { return r.dismissed }

// Toaster draws the toasts in a store, stacked in the top-right corner of
// whatever it is placed in, newest at the bottom — the order a person
// dismisses them in.
//
// It is a toaster and not a NotificationCenter because a toast is a
// transient report the caller also owns: the centre is a standing list the
// panel owns, and the toasts here are the caller's list, read fresh each
// frame. The component does not take toasts away on its own; it reports the
// press and the caller dismisses, which is the only way a toast can go away
// when the caller knows something the component does not — that the work
// the toast announced has been undone, for instance, and the toast should
// not outlive it.
//
// The toaster fills its parent and lets the pointer through everything it
// draws, so it can sit over a whole window without turning it into a
// surface that eats clicks. Place it at the root of a window, as a sibling
// of the content, and it draws on top.
func Toaster(c *ui.Context, store *ToastStore, opts ToasterOptions) ToasterResult {
	if store == nil {
		panic("feedback: Toaster needs a store; a toaster with no list " +
			"announces nothing and cannot report a press")
	}
	u := core.Density(c).Unit()

	width := opts.Width
	if width <= 0 {
		width = u * 34
	}
	var r ToasterResult
	r.dismissed = -1

	// The column is built inside the wrapper rather than beside it, for the
	// reason a meter's bar is: an element belongs to whatever was being
	// built when it was made, so one made out here would land in the
	// caller's container instead of the toaster.
	e := ui.Box(c).Fill().PassThrough().Children(func() {
		col := ui.Column(c).Width(width).Gap(u * 2).Absolute().Top(u * 3)
		if opts.Center {
			// Full-width and centre-aligned rather than centred by position:
			// a column pinned to the middle of the window would re-pin itself
			// when a toast arrived and moved its own width.
			col.Left(0).Right(0).AlignItems(ui.Center)
		} else {
			col.Right(u * 3)
		}
		col.Children(func() {
			for _, t := range store.toasts {
				toastItem(c, t, &r)
			}
		})
	})
	r.Element = e
	return r
}

// toastItem draws one row of the toaster: the severity's dot, the words,
// and the button that takes the row off the list.
func toastItem(c *ui.Context, t ToastItem, r *ToasterResult) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	dot := u * 2.25

	ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u*2).
		Padding(u*2.5, u*3.5, u*2.5, u*3.5).Radius(theme.ControlRadius).
		Background(k.Fill).Label(t.Title).
		Shadow(0, u*1.5, u*4.5, 0, ui.RGBA(0, 0, 0, 0.18)).Children(func() {
		ui.Box(c).Size(dot, dot).Radius(dot / 2).Shrink(0).
			Background(severityInk(k, t.Severity))
		ui.Column(c).Grow(1).Gap(u * 0.5).Children(func() {
			ui.Text(c, t.Title).TextColor(k.OnFill).
				FontSize(core.FontSize(c, theme.RowSize)).Bold()
			if t.Body != "" {
				ui.Text(c, t.Body).TextColor(k.OnFill.Alpha(0.75)).
					FontSize(core.FontSize(c, theme.MetaSize))
			}
		})
		// The close is on the toast's own fill, the way an UndoToast's
		// undo is: the fill has no background to pair a severity with, and
		// a control that sits on it takes the fill's own inverse, faded.
		closeBtn := ui.Button(c, "").Size(u*6, u*6).Radius(theme.PillRadius).
			Shrink(0).Background(k.OnFill.Alpha(0.16)).
			Label("Dismiss " + t.Title).Tooltip("Dismiss " + t.Title)
		closeBtn.Children(func() {
			ui.Text(c, "✕").TextColor(k.OnFill).
				FontSize(core.FontSize(c, theme.CaptionSize))
		})
		if closeBtn.Clicked() && r.dismissed < 0 {
			r.dismissed = t.ID
		}
	})
}
