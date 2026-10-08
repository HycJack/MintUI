package feedback

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// Notification is one thing the application wants said once.
type Notification struct {
	// Title is what happened, in the present tense.
	Title string
	// Body is the line under it. Empty is fine for a notification with
	// nothing to add.
	Body string
	// At is when, already formatted by the caller, for the same reason an
	// ActivityItem's is.
	At string
	// Severity marks the row. Zero is core.Neutral, which is right for most
	// of what a window has to say.
	Severity core.Severity
}

// NotificationCenterOptions configure a NotificationCenter.
type NotificationCenterOptions struct {
	// Items are the notifications, newest first. An empty list is not an
	// error here — a window with nothing to say is a window at rest, and
	// drawing an empty panel saying so is better than drawing nothing,
	// which looks like the panel failed.
	Items []Notification
	// Title heads the list; empty draws no heading, for a centre that is
	// already inside a panel with one.
	Title string
	// Empty is what the panel says when there is nothing in it. Empty takes
	// the library's copy, which a window can reword through core.Msg.
	Empty string
	// Rule draws a hairline between rows.
	Rule bool
	// Clearable adds a "Clear all" button over the list and reports its
	// press. It is off by default: a centre that can be emptied in one press
	// should say so, and the caller has to say it.
	Clearable bool
	// Width bounds the centre; zero lets it fill its parent.
	Width float32
}

// NotificationCenterResult carries a NotificationCenter and what was pressed.
type NotificationCenterResult struct {
	// Element is the whole centre.
	Element *ui.Element
	// dismissed is the index of the row whose close was pressed this frame,
	// or -1. It is an index rather than a pointer because the list is the
	// caller's, and a caller acting on it will look it up by the same index.
	dismissed int
	cleared   bool
}

// Dismissed reports the index of the notification whose close button was
// pressed this frame, or -1 when none was. Only the first press of a frame is
// reported: two rows cannot be dismissed in one frame by one pointer, and
// returning the last would make the earlier one's press silently do
// nothing.
func (r NotificationCenterResult) Dismissed() int { return r.dismissed }

// Cleared reports the press of the clear-all button.
func (r NotificationCenterResult) Cleared() bool { return r.cleared }

// NotificationCenter is a window's standing list of things it wants said,
// newest first, each dismissible on its own.
//
// It is a panel and not a stack of toasts because the two answer different
// questions: a toast says what just happened and goes, this says what is
// still outstanding. A window that only ever used toasts has no way to show
// a user the six things that arrived while they were looking away.
func NotificationCenter(c *ui.Context, opts NotificationCenterOptions) NotificationCenterResult {
	for _, n := range opts.Items {
		if n.Title == "" {
			panic("feedback: a Notification needs a Title; a row with no title " +
				"is a dot and a close button named after nothing")
		}
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	heading := opts.Title
	if heading == "" && len(opts.Items) > 0 {
		heading = core.Msg(c, "feedback.notificationCenter.title", "Notifications")
	}
	empty := opts.Empty
	if empty == "" {
		empty = core.Msg(c, "feedback.notificationCenter.empty", "Nothing to report")
	}
	var r NotificationCenterResult
	r.dismissed = -1
	clear := clearAll(c)

	panel := ui.Column(c).FillWidth().Gap(u * 2).Role(ui.RoleList).Label(heading)
	if opts.Width > 0 {
		panel.Width(opts.Width)
	}
	panel.Children(func() {
		if heading != "" {
			ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 2).Children(func() {
				ui.Text(c, heading).TextColor(k.Text).
					FontSize(core.FontSize(c, theme.TitleSize)).Bold()
				ui.Box(c).Grow(1)
				if opts.Clearable && len(opts.Items) > 0 {
					btn := ui.Button(c, "").Shrink(0).Radius(theme.PillRadius).
						Padding(u, u*2.5, u, u*2.5).Background(k.Surface).
						TextColor(k.TextMuted).Label(clear).Tooltip(clear)
					btn.Children(func() {
						ui.Text(c, clear).TextColor(k.TextMuted).
							FontSize(core.FontSize(c, theme.CaptionSize))
					})
					r.cleared = btn.Clicked()
				}
			})
		}

		if len(opts.Items) == 0 {
			// The same panel, the same padding and a sentence instead of
			// rows, so the centre does not jump in size when the first thing
			// arrives — a panel that resizes as content arrives moves every
			// control below it.
			ui.Box(c).FillWidth().Radius(theme.ControlRadius).
				Background(k.Surface).Padding(u*4, u*3.5, u*4, u*3.5).
				Children(func() {
					ui.Text(c, empty).TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.MetaSize))
				})
			return
		}

		ui.Column(c).FillWidth().Role(ui.RoleList).Children(func() {
			for i, n := range opts.Items {
				if i > 0 && opts.Rule {
					layout.Divider(c, layout.DividerOptions{})
				}
				notificationRow(c, i, n, &r)
			}
		})
	})

	r.Element = panel
	return r
}

// notificationRow draws one notification: its mark, its words, and the
// button that takes it off the list.
func notificationRow(c *ui.Context, index int, n Notification, r *NotificationCenterResult) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	dot := u * 2.25

	ui.Row(c).FillWidth().AlignItems(ui.Start).Gap(u*2.5).
		Padding(u*2.5, 0, u*2.5, 0).Label(n.Title).Children(func() {
		ui.Box(c).Size(dot, dot).Radius(dot/2).Shrink(0).
			Background(severityInk(k, n.Severity)).Margin(u*0.75, 0, 0, 0)
		ui.Column(c).Grow(1).Gap(u * 0.5).Children(func() {
			ui.Text(c, n.Title).TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).Bold()
			if n.Body != "" {
				ui.Text(c, n.Body).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.MetaSize))
			}
			if n.At != "" {
				ui.Text(c, n.At).TextColor(k.TextFaint).
					FontSize(core.FontSize(c, theme.CaptionSize))
			}
		})
		// The close is a plain × rather than an icon button with a glyph
		// package behind it: at this size the character is the icon, and a
		// drawn one would only be a worse ×.
		closeBtn := ui.Button(c, "").Size(u*6.5, u*6.5).Radius(theme.PillRadius).
			Shrink(0).Background(k.Surface).Label("Dismiss " + n.Title).
			Tooltip("Dismiss " + n.Title)
		closeBtn.Children(func() {
			ui.Text(c, "✕").TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize))
		})
		if closeBtn.Clicked() && r.dismissed < 0 {
			r.dismissed = index
		}
	})
}

// clearAll is the copy for the button that empties the centre. It goes
// through core.Msg because it is library copy: a window that needs it to say
// something else should not have to fork the component.
func clearAll(c *ui.Context) string {
	return core.Msg(c, "feedback.notificationCenter.clearAll", "Clear all")
}

// UndoToastOptions configure an UndoToast.
//
// There is no Severity here, deliberately. A toast is drawn on the fill
// surface, which is the one place in the interface with no background to
// pair a severity with; a caller who needs the colour means something has
// gone wrong, and that is an Alert, which has a surface to tint.
type UndoToastOptions struct {
	// Message is what happened. It is required, unlike Toast's: a toast with
	// no message withdraws because there is nothing to say, while an undo
	// with no message is a control that undoes nothing.
	Message string
	// Action labels the undo button. Empty draws no button, which makes an
	// UndoToast a Toast that happens to know it has nothing to offer — and
	// is right for a caller that will follow up with a real one.
	Action string
}

// UndoToastResult carries an UndoToast and what was pressed in it.
type UndoToastResult struct {
	// Element is the whole toast.
	Element *ui.Element
	undo    bool
}

// Undo reports the undo button. It is a single method rather than one for
// the button and one for a dismiss, because there is nothing to dismiss: the
// toast is the caller's to take away, and all it does is offer the undo.
func (r UndoToastResult) Undo() bool { return r.undo }

// UndoToast is a toast that offers to take back what it just reported: an
// archived callback put back on the board, a sent message recalled.
//
// The undo is the reason this is a component and not a toast with a button.
// The button has to sit inside the toast's own surface, at its own height,
// with the same padding — and it has to be the last thing in it, so that a
// second one can never appear beside it and make the toast's width depend on
// which buttons happen to be showing.
func UndoToast(c *ui.Context, opts UndoToastOptions) UndoToastResult {
	if opts.Message == "" {
		panic("feedback: UndoToast needs a Message; a toast with nothing to " +
			"say and something to undo offers only the undo")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	var r UndoToastResult

	action := opts.Action
	if action == "" {
		action = core.Msg(c, "feedback.undoToast.action", "Undo")
	}

	toast := ui.Row(c).AlignItems(ui.Center).Gap(u*2.5).
		Padding(u*2.5, u*4, u*2.5, u*4).Radius(theme.PillRadius).
		Background(k.Fill).Label(opts.Message).
		Shadow(0, u*1.5, u*4.5, 0, ui.RGBA(0, 0, 0, 0.18)).Children(func() {
		ui.Text(c, opts.Message).TextColor(k.OnFill).
			FontSize(core.FontSize(c, theme.RowSize))
		if opts.Action == "" {
			return
		}
		btn := ui.Button(c, "").Height(core.ControlHeight(c)).Shrink(0).
			Radius(theme.PillRadius).Padding(0, u*3.5, 0, u*3.5).
			Background(k.OnFill.Alpha(0.16)).TextColor(k.OnFill).
			Label(action).Tooltip(action)
		btn.Children(func() {
			ui.Text(c, action).TextColor(k.OnFill).
				FontSize(core.FontSize(c, theme.RowSize)).Bold()
		})
		r.undo = btn.Clicked()
	})

	r.Element = toast
	return r
}
