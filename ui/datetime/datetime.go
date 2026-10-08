// Package datetime holds everything the interface shows about dates and
// times: the grids a month is drawn from, the pickers that choose a day or an
// hour, the views that lay events out, and the small marks that say "now" or
// "in three days".
//
// # Time is the caller's
//
// Nothing here calls [time.Now]. A component that needs to know what "now" is
// is given it, as a field, on every frame:
//
//	Countdown(c, CountdownOptions{Now: clock.Now(), Target: job.DueAt})
//
// The reason is testability and honesty at once. A component that reads the
// clock itself draws a different frame every time it is rendered, so a
// snapshot test of it is a test of the wall clock, and a caller cannot freeze
// time to look at a view at 3am the next morning. Passing "now" in also
// makes the dependency visible: a form full of pickers that each quietly read
// a different clock is a form whose fields disagree at midnight on a
// day-saving boundary.
//
// # State is the caller's too
//
// The selected day is the caller's [time.Time] or a pointer to one, the open
// month is a pointer to a [time.Time] the arrows move, and a press comes back
// as a method on the result:
//
//	var from, to time.Time
//	r := DateRangePicker(c, DateRangePickerOptions{Start: &from, End: &to})
//	if r.Changed() {
//		save(from, to)
//	}
//
// # Weeks
//
// A week starts on Monday and Monday is 0, in the names as much as in the
// arithmetic — [time.Time.Weekday] has Sunday as 0 and the grid has Monday as
// 0, and the conversion is written once, in [Weekday]. The names come from
// the caller as a [WeekdayNames], so a window that shows them in another
// language passes them in; nothing here translates.
package datetime

import (
	"strconv"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// The marks this package draws, in the same 24×24 stroke style as the rest of
// the library, so a calendar's arrows and a card's chevron are the same shape.
func icon(shapes string) *ui.SVG {
	return ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" ` +
		`fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" ` +
		`stroke-linejoin="round">` + shapes + `</svg>`))
}

var (
	iconPrev = icon(`<path d="m14.5 5.5-6.5 6.5 6.5 6.5"/>`)
	iconNext = icon(`<path d="m9.5 5.5 6.5 6.5-6.5 6.5"/>`)
	iconUp   = icon(`<path d="m5.5 14.5 6.5-6.5 6.5 6.5"/>`)
	iconDown = icon(`<path d="m5.5 9.5 6.5 6.5 6.5-6.5"/>`)
)

// MinCellSize is the smallest a day cell may be, in DIPs.
//
// It is a floor rather than a preference: below it a cell stops being a target
// a fingertip or a screen cursor can hit, and a calendar that cannot be aimed
// at is a calendar that needs a mouse. It is also the width a seven-column
// month needs before the day numbers have to shrink, so raising it is the
// first thing to try when a picker is too cramped.
const MinCellSize float32 = 28

// cellSide is the side of one day cell at this window's density. Comfortable
// makes it larger, and MinCellSize keeps Compact from dropping under the floor.
func cellSide(u float32) float32 { return max(MinCellSize, u*7) }

// stepButton is one of the small square buttons a picker and a calendar use to
// move: an hour, a minute, a month. It is a real ui.Button so it takes focus,
// announces itself and stops working when disabled — none of which a drawn
// triangle would do.
func stepButton(c *ui.Context, glyph *ui.SVG, name string, size float32) *ui.Element {
	k := core.Tokens(c)
	return ui.Button(c, "").Size(size, size).Shrink(0).
		Radius(theme.ControlRadius).Background(k.Surface).TextColor(k.Text).
		Label(name).Tooltip(name).Children(func() {
		ui.Icon(c, glyph).TextColor(k.Text).Size(size*0.5, size*0.5)
	})
}

func itoa(n int) string { return strconv.Itoa(n) }
