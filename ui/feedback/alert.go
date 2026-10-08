package feedback

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// AlertOptions configure an Alert.
type AlertOptions struct {
	// Title is the one line that says what happened.
	Title string
	// Body is the line under it. Either a title or a body is required: an
	// alert with neither is a coloured box, which is worse than nothing
	// because it looks like it is about to say something.
	Body string
	// Severity picks the colours, and it is the only thing that does. The
	// same core.Severity.Pair the board's priority pills use, so a critical
	// pill and a critical alert are recognisably the same red.
	Severity core.Severity
	// Action labels a button that does something about it — "Retry",
	// "Reassign". Empty draws none: an alert usually has no action, and a
	// disabled-looking button on every alert trains people to ignore them.
	Action string
	// Dismissable adds a close button and reports its press.
	Dismissable bool
}

// AlertResult carries an Alert and what was pressed in it.
type AlertResult struct {
	// Element is the whole alert.
	Element           *ui.Element
	pressed, actioned bool
}

// Pressed reports the close button, or the action button when the alert has
// no separate close.
func (r AlertResult) Pressed() bool { return r.pressed }

// Actioned reports the action button.
func (r AlertResult) Actioned() bool { return r.actioned }

// Alert is a message that interrupts what the user was doing: a sync failed, a
// callback came back with an error, a quota ran out.
//
// It is inlined rather than floating, because an alert is about the content
// it sits in — a card that failed to save says so on the card. The severity
// tints a rule down its leading edge rather than the whole surface, so the
// words stay at the ordinary contrast they were designed for while the edge
// still says which of the five this is.
func Alert(c *ui.Context, opts AlertOptions) AlertResult {
	if opts.Title == "" && opts.Body == "" {
		panic("feedback: Alert needs a Title or a Body; an empty one is a " +
			"coloured box that looks like it is about to say something")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	title := opts.Title
	var r AlertResult
	rule := u * 1.5

	alert := ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u*3).
		Background(k.Surface).Radius(theme.ControlRadius).
		Padding(u*2.5, u*3.5, u*2.5, u*3.5).Role(ui.RoleGroup)
	if title == "" {
		// With no title the alert is named by its body, so a screen reader
		// reaches something to say.
		title = opts.Body
		alert.Label(opts.Body)
	} else {
		alert.Label(title)
	}

	alert.Children(func() {
		if opts.Dismissable {
			// Before the text, because an alert is read by its rule first
			// and dismissing it is the fastest way out of one.
			close := ui.Button(c, "").Size(u*7, u*7).Radius(theme.PillRadius).
				Shrink(0).Background(k.SurfaceHover).Label("Dismiss " + title)
			close.Children(func() {
				ui.Text(c, "✕").TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize))
			})
			r.pressed = close.Clicked()
		}

		// The rule: a full-height block at the leading edge, which is the
		// one place a severity can be shown without tinting the words.
		ui.Box(c).Shrink(0).Width(rule).FillHeight().AlignSelf(ui.Stretch).Radius(rule / 2).
			Background(severityInk(k, opts.Severity))

		ui.Column(c).Grow(1).Gap(u * 0.75).Children(func() {
			if opts.Title != "" {
				ui.Text(c, title).TextColor(k.Text).
					FontSize(core.FontSize(c, theme.BodySize)).Bold()
			}
			if opts.Body != "" {
				ui.Text(c, opts.Body).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.MetaSize))
			}
		})

		if opts.Action != "" {
			btn := ui.Button(c, "").Shrink(0).Height(core.ControlHeight(c)).
				Radius(theme.PillRadius).Padding(0, u*3.5, 0, u*3.5).
				Background(k.Background).TextColor(k.Text).
				Border(theme.BorderWidth, k.Border).Label(opts.Action).
				Tooltip(opts.Action)
			btn.Children(func() {
				ui.Text(c, opts.Action).TextColor(k.Text).
					FontSize(core.FontSize(c, theme.RowSize))
			})
			r.actioned = btn.Clicked()
			if !opts.Dismissable {
				// One button means one meaning, so an alert with an action
				// and no close reports that button's press as the press.
				r.pressed = r.pressed || r.actioned
			}
		}
	})

	r.Element = alert
	return r
}

// ResultOptions configure a Result.
type ResultOptions struct {
	// Title is the outcome in a few words: "Callback resolved",
	// "Import failed". It is required: a result is the one component here
	// whose whole job is to name what happened.
	Title string
	// Body says what to do about it, in one sentence.
	Body string
	// Severity picks the mark's colour and the surface behind it.
	Severity core.Severity
	// Action labels the button that continues — "Try again", "Open the
	// callback". Empty draws none, which is right for a result the user
	// only has to read.
	Action string
}

// ResultResult carries a Result and the press of its action.
type ResultResult struct {
	// Element is the whole result.
	Element *ui.Element
	pressed bool
}

// Pressed reports the action button.
func (r ResultResult) Pressed() bool { return r.pressed }

// Result is the end of a piece of work, said once: success, failure, or
// something in between.
//
// It is a panel rather than an inline alert because it is the end of the
// story rather than an interruption inside it, and it is drawn on its own
// surface with a large mark because after a wait the first question is
// whether it worked, and that should be answerable from across the room.
func Result(c *ui.Context, opts ResultOptions) ResultResult {
	if opts.Title == "" {
		panic("feedback: Result needs a Title; naming the outcome is its whole job")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	ink := severityInk(k, opts.Severity)
	mark := u * 13
	var r ResultResult

	panel := ui.Column(c).FillWidth().AlignItems(ui.Center).Gap(u*2.5).
		Background(k.Surface).Radius(theme.CardRadius).
		Padding(u*6.5, u*6, u*6.5, u*6).Role(ui.RoleGroup).Label(opts.Title).
		Children(func() {
			// The mark is drawn rather than typed: it has to be the same weight
			// as the panel it heads and hold up at any size, which an icon
			// font does not promise.
			ui.Box(c).Size(mark, mark).Shrink(0).Draw(func(p *ui.Painter, r ui.Rect) {
				drawOutcomeMark(p, r, opts.Severity, ink)
			})
			ui.Text(c, opts.Title).TextColor(k.Text).
				FontSize(core.FontSize(c, theme.LeadSize)).Bold().TextAlign(ui.Center)
			if opts.Body != "" {
				ui.Text(c, opts.Body).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.MetaSize)).TextAlign(ui.Center)
			}
			if opts.Action != "" {
				btn := ui.Button(c, "").Height(core.ControlHeight(c)+u).Shrink(0).
					Radius(theme.PillRadius).Padding(0, u*5, 0, u*5).
					Background(k.Fill).TextColor(k.OnFill).
					Label(opts.Action).Tooltip(opts.Action)
				btn.Children(func() {
					ui.Text(c, opts.Action).TextColor(k.OnFill).
						FontSize(core.FontSize(c, theme.RowSize))
				})
				r.pressed = btn.Clicked()
			}
		})

	r.Element = panel
	return r
}

// drawOutcomeMark draws the symbol a Result is named by: a ring, and inside
// it the mark of the severity. The glyphs are strokes rather than glyphs
// from a font, so the tick and the cross are the same weight as the ring
// around them and land on the same baseline at any panel size.
func drawOutcomeMark(p *ui.Painter, r ui.Rect, sev core.Severity, ink ui.Color) {
	rad := min(r.W, r.H) / 2
	cx, cy := r.X+rad, r.Y+rad
	inner := rad * 0.82
	width := max(1, rad*0.12)

	switch sev {
	case core.Success:
		// A tick: up and across, drawn as two strokes.
		s := inner * 0.42
		p.Line(cx-s, cy, cx-s*0.25, cy+s*0.7, width, ink)
		p.Line(cx-s*0.25, cy+s*0.7, cx+s, cy-s*0.55, width, ink)
	case core.Danger:
		// A cross, the same two strokes turned into a stop.
		s := inner * 0.42
		p.Line(cx-s, cy-s, cx+s, cy+s, width, ink)
		p.Line(cx-s, cy+s, cx+s, cy-s, width, ink)
	case core.Warning:
		// A bar and a dot: the exclamation mark, at the ring's own weight.
		s := inner * 0.5
		p.Line(cx, cy-s*0.55, cx, cy+s*0.15, width, ink)
		internal.Dot(p, cx, cy+s*0.6, width/2, ink)
	default:
		// Neutral and Accent have nothing to assert, so the ring is closed
		// with a dot rather than a tick that would claim more than it knows.
		internal.Dot(p, cx, cy, inner*0.26, ink)
	}
	internal.Ring(p, cx, cy, rad, width, ink.Alpha(0.35))
}
