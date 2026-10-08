package data

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// Event is one entry of a Timeline: when it happened, what happened, and
// whatever else is said about it.
type Event struct {
	// When is the stamp of the event, already formatted. A Timeline draws
	// what it is given and formats nothing: "3 days ago" and "on the 14th"
	// are the caller's to choose between, and a log of callbacks and a log
	// of payments are not read on the same clock.
	When string
	// Title says what happened. It is the line the eye finds.
	Title string
	// Detail is the line under it, in the muted tone.
	Detail string
	// Tone ranks the entry, for the dot down the line: the one entry that
	// has to be findable by looking is the one that wears a colour.
	Tone core.Severity
	// Marker draws the mark instead of the dot, for an entry that carries a
	// mark of its own.
	Marker func()
}

// TimelineOptions configure a Timeline.
type TimelineOptions struct {
	// Height is the height of the timeline. It is required: a timeline with
	// no height grows to fit its events, which is every event drawn rather
	// than a timeline.
	Height float32
	// Gap is the space between events; zero for events that touch.
	Gap float32
	// Stamps puts each event's stamp in a column of its own down the left,
	// where a run of times can be read as a run. Off, the stamp sits above
	// what it stamps.
	Stamps bool
	// StampWidth is the width of that column; zero is the library's own.
	StampWidth float32
	// NoLine draws no rule down the left for the dots to hang on, for a
	// timeline of two events where a rule says nothing.
	NoLine bool
	// RowHeight is the least an event is high; zero is the library's own.
	RowHeight float32
	// Width is the width the events are laid out in; zero takes the window
	// they are in.
	Width float32
	// State is the list's place between frames; nil keeps it in the
	// timeline's own.
	State *ui.ListState
	// Scroll is where the timeline is scrolled, when the caller keeps one.
	Scroll *ui.ScrollState
	// Empty draws instead of the events when there are none.
	Empty func()
}

// Timeline is a run of events down a rule: the time of each, a mark for
// what it was, and what happened. It builds only the events in view, as
// the other row views do.
//
// Nothing here is pressed. A timeline says what happened; what to do about
// it belongs to the caller, and a mark that took a click would be a control
// wearing a fact's clothes.
func Timeline(c *ui.Context, opts TimelineOptions, events ...Event) *ui.Element {
	u := core.Density(c).Unit()
	h := opts.RowHeight
	if h <= 0 {
		h = rowHeight(u)
	}
	stampW := opts.StampWidth
	if stampW <= 0 {
		stampW = u * 20
	}
	if opts.State != nil {
		opts.State.Label = func(row int) string { return events[row].Title }
	}
	return rowsWindow(c, windowOptions{
		rows: len(events), height: opts.Height, width: opts.Width,
		gap:  opts.Gap,
		role: ui.RoleList, state: opts.State, scroll: opts.Scroll, empty: opts.Empty,
		row: func(i int, _ []float32) {
			event(c, events[i], opts, u, h, stampW, i == 0, i == len(events)-1)
		},
	})
}

// event draws one entry: its stamp, the rule and the mark down the side, and
// what it says.
func event(c *ui.Context, e Event, opts TimelineOptions, u, h, stampW float32, first, last bool) {
	k := core.Tokens(c)
	// An event is at least a row high, so that a run of one-line events has
	// the air between them that makes it a run rather than a paragraph.
	r := ui.Row(c).FillWidth().MinHeight(h).Shrink(0).AlignItems(ui.Stretch).
		Role(ui.RoleListItem).Padding(0, u*0.5)

	r.Children(func() {
		if e.When != "" && opts.Stamps {
			// The stamps in a column of their own, so that a run of times
			// reads as a run rather than as one more line under each of them.
			ui.Row(c).Width(stampW).Shrink(0).AlignItems(ui.Start).
				Padding(0, u*0.5).Children(func() {
				ui.Text(c, e.When).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.MetaSize)).SingleLine()
			})
		}
	})

	// The rule and the mark on it, drawn rather than laid out: a rule that
	// stopped above the mark and started again below it would be one mark
	// in three, and a run of dots with a gap between each is a column of
	// nothing.
	r.Children(func() {
		if e.Marker != nil {
			ui.Box(c).Width(u * 5).Shrink(0).Center().Children(e.Marker)
			return
		}
		ui.Box(c).Width(u * 5).Shrink(0).Draw(func(p *ui.Painter, rect ui.Rect) {
			cx, cy := rect.X+rect.W/2, rect.Y+u*2.5
			if !opts.NoLine {
				from, to := rect.Y, rect.Y+rect.H
				if first {
					from = cy
				}
				if last {
					to = cy
				}
				if to > from {
					p.Fill(ui.Rect{X: cx - theme.BorderWidth/2, Y: from,
						W: theme.BorderWidth, H: to - from}, k.Border, 0)
				}
			}
			col := k.TextFaint
			if e.Tone != core.Neutral {
				_, col = e.Tone.Pair(k)
			}
			internal.Dot(p, cx, cy, u*1.5, col)
		})
	})

	r.Children(func() {
		ui.Column(c).Grow(1).Gap(u * 0.5).Children(func() {
			if e.When != "" && !opts.Stamps {
				ui.Text(c, e.When).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
			}
			ui.Text(c, e.Title).TextColor(k.Text).
				FontSize(core.FontSize(c, theme.BodySize)).SingleLine()
			if e.Detail != "" {
				ui.Text(c, e.Detail).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.MetaSize))
			}
		})
	})
}
