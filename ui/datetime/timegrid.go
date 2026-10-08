package datetime

import (
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// HourHeight returns how tall one hour is in a time column at this density.
//
// It is a number rather than a scale because the events in a column are laid
// out in pixels: a chip's height is its length in minutes multiplied by this,
// and the two have to agree to the pixel, or a 30-minute booking is 29 minutes
// tall and nothing lines up with anything.
func HourHeight(u float32) float32 { return max(56, u*14) }

// HourY returns how far down a time column a moment sits, given the column's
// first hour and its scale.
//
// Every view in this package places things with it, and it is exported
// because a caller drawing its own thing in a time column — a drag handle, a
// technician's marker — has to put it at the height the column does, or it
// will not line up with the event it belongs to.
func HourY(t time.Time, startHour int, hourHeight float32) float32 {
	minutes := float32(t.Hour()*60+t.Minute()) - float32(startHour*60)
	// An event from the day before or the day after is clipped to the window
	// rather than placed off its top or bottom, where it would draw over the
	// neighbouring column.
	return minutes / 60 * hourHeight
}

// TimeGridOptions configure a CalendarTimeGrid.
type TimeGridOptions struct {
	// Day is any day in the column; only its date matters.
	Day time.Time
	// StartHour and EndHour are the hours on screen; zero means 8 to 18.
	StartHour, EndHour int
	// HourHeight is the scale; zero uses HourHeight for the density.
	HourHeight float32
	// Pickable makes each hour a button and reports the time a press landed
	// on, snapped to the hour. It picks hours and nothing finer: a quarter of
	// an hour is four bands inside 56 DIP, which are not targets anybody can
	// hit, and the minutes belong to TimePicker, which has a control for them.
	Pickable bool
	// Now draws the current-time line. The zero time draws none — nothing in
	// this package reads the clock to find out what time it is.
	Now time.Time
	// Hours draws the hour numbers down the side, which is what a standalone
	// day column wants and what a week view draws once for all seven columns.
	Hours bool
	// Disabled takes the clicks away from the hours it names.
	Disabled func(time.Time) bool
}

// TimeGridResult carries a CalendarTimeGrid and the time the user pressed.
type TimeGridResult struct {
	// Element is the column of hour rules, with the hour numbers beside it.
	Element *ui.Element
	// picked is the time pressed this frame, and got whether there was one.
	picked time.Time
	got    bool
}

// Picked returns the time the user pressed, and whether they pressed one.
func (r TimeGridResult) Picked() (time.Time, bool) { return r.picked, r.got }

// CalendarTimeGrid is the ruled backdrop of a time column: a line at every
// hour, the hours down the side, and the current-time line across it.
//
// It draws rather than fills, and it is deliberately not what holds the
// events: an hour column's height must be the height of its hours, so that a
// chip's position means the same thing in every column. A column that grew to
// fit its longest event would push the hours below it out of line with the
// hours of the columns beside it.
func CalendarTimeGrid(c *ui.Context, opts TimeGridOptions) TimeGridResult {
	k, u := core.Tokens(c), core.Density(c).Unit()
	start, end := hourWindow(opts.StartHour, opts.EndHour)
	hourHeight := opts.HourHeight
	if hourHeight <= 0 {
		hourHeight = HourHeight(u)
	}
	day := StartOfDay(opts.Day)
	labelWidth := float32(0)
	if opts.Hours {
		labelWidth = u * 9
	}

	var r TimeGridResult
	col := ui.Row(c).FillWidth().FillHeight().Children(func() {
		ui.Column(c).FillWidth().FillHeight().Children(func() {
			for hour := start; hour < end; hour++ {
				at := day.Add(time.Duration(hour) * time.Hour)
				// The rule is drawn rather than bordered so that it is one
				// pixel at the top of the hour and moves nothing sitting on
				// it, which is every event in the column.
				rule := func(p *ui.Painter, rect ui.Rect) {
					internal.Rule(p, rect, k.Border)
				}
				if !opts.Pickable {
					ui.Box(c).FillWidth().Height(hourHeight).Shrink(0).Draw(rule)
					continue
				}
				disabled := opts.Disabled != nil && opts.Disabled(at)
				btn := ui.Button(c, "").FillWidth().Height(hourHeight).Shrink(0).
					Radius(theme.SmallRadius).Background(ui.Transparent).
					TextColor(k.TextMuted).Disabled(disabled).
					Label(HourLabel(hour) + " " + FormatDate(day))
				btn.Draw(func(p *ui.Painter, rect ui.Rect) {
					// Hover is drawn rather than themed, so that the band can
					// have no background of its own — the column behind it is
					// the window's, not the grid's — and still answer a
					// pointer.
					if btn.Hovered() {
						p.Fill(rect, k.SurfaceHover, 0)
					}
					rule(p, rect)
				})
				if btn.Clicked() {
					r.picked, r.got = at, true
				}
			}
		})
		if opts.Hours {
			ui.Column(c).Width(labelWidth).Shrink(0).FillHeight().Children(func() {
				for hour := start; hour < end; hour++ {
					ui.Text(c, HourLabel(hour)).TextColor(k.TextFaint).
						FontSize(theme.CaptionSize).Height(hourHeight).
						TextAlign(ui.End).SingleLine()
				}
			})
		}
		// The line is over the rules and not under them: a reader looking for
		// the time looks for where now is first, and what is on second.
		if !opts.Now.IsZero() {
			CurrentTimeIndicator(c, CurrentTimeIndicatorOptions{
				Now:        opts.Now,
				StartHour:  start,
				HourHeight: hourHeight,
			})
		}
	})
	r.Element = col
	return r
}

// HourLabel writes an hour as 09 — two digits, always.
//
// Two digits because a column reading 9, 10, 11 is ragged and one reading 09,
// 10, 11 is not, and a ruler is a thing that has to line up.
func HourLabel(hour int) string {
	if hour < 10 {
		return "0" + itoa(hour)
	}
	return itoa(hour)
}

// CurrentTimeIndicatorOptions configure a CurrentTimeIndicator.
type CurrentTimeIndicatorOptions struct {
	// Now is the caller's now. The zero time draws nothing at all.
	Now time.Time
	// StartHour is the column's first hour and HourHeight its scale; they
	// have to be the column's own, or the line lands at the wrong minute.
	StartHour  int
	HourHeight float32
	// Inset is how far the line stops short of the column's right edge.
	Inset float32
}

// CurrentTimeIndicator is the line across a time column that says what time it
// is now.
//
// It is absolutely placed inside the column rather than being a row in one,
// because it is not one of the column's hours: it is a reading, and putting it
// in the flow would move every hour below it down by its own thickness. The
// dot on the leading end is what makes it findable — a hairline across forty
// events is easy to miss, and the presence dot is the one mark in this library
// that says "here".
func CurrentTimeIndicator(c *ui.Context, opts CurrentTimeIndicatorOptions) *ui.Element {
	if opts.Now.IsZero() {
		return nil
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	hourHeight := opts.HourHeight
	if hourHeight <= 0 {
		hourHeight = HourHeight(u)
	}
	top := HourY(opts.Now, opts.StartHour, hourHeight)
	inset := opts.Inset
	if inset == 0 {
		inset = u
	}
	return ui.Box(c).Absolute().Top(top).Left(0).FillWidth().Height(u).
		Margin(0, inset, 0, 0).Shrink(0).
		Label(core.Msg(c, "datetime.now", "Now")).
		Draw(func(p *ui.Painter, rect ui.Rect) {
			p.Fill(ui.Rect{X: rect.X + u, Y: rect.Y, W: max(rect.W-u*2.5, 0), H: rect.H},
				k.Danger, rect.H/2)
			internal.Dot(p, rect.X, rect.Y+rect.H/2, rect.H/2, k.Danger)
		})
}
