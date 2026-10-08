package input

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// The sliders. MyGo has a slider and no range slider, so the first is MyGo's
// with this library's face on the number beside it, and the second is one
// ui.SliderBase standing in for both thumbs: the input — the drag, the
// arrows, Home and End, the focus, the range assistive technology is told
// about — is entirely MyGo's, and only the painting is ours.

// knobWidth and knobHeight are the thumb's size, the capsule MyGo's own
// slider draws, so that a single-value slider and a range slider have thumbs
// that match each other.
const (
	knobWidth  float32 = 20
	knobHeight float32 = 16
)

// trackHeight is the rail the thumb runs on, as thick as a Progress bar.
const trackHeight float32 = 6

// SliderOptions configure a Slider.
type SliderOptions struct {
	// Min is the lowest value the slider can be set to and Max the highest.
	// Max must be above Min: a slider with no room to move is a bar that
	// cannot be set.
	Min, Max float64
	// Step makes the values Min and the multiples of Step from it, with a
	// tick mark at each, as a slider with detents has. Zero is every value
	// in between, which is right for a percentage and wrong for a price.
	Step float64
	// Format is how the value is written beside the slider, as "%.0f%%". It
	// is a format rather than a prefix and a suffix because minutes and
	// dollars are both "%.0f" to the caller and neither is to the library.
	// Empty writes the number on its own.
	Format string
	// ShowValue puts the value beside the slider. It is off by default: a
	// form with its own field for the number would be saying it twice, and a
	// form without one should ask for it.
	ShowValue bool
	// Label names the slider for assistive technology, and is required: a
	// slider draws a rail and a thumb and no words, so without a name there
	// is nothing to read out and nothing saying what is being measured.
	Label string
	// Disabled greys the control out: it takes neither presses nor focus.
	Disabled bool
}

// Slider is a rail and a thumb setting a number between two bounds, which it
// writes into the caller's float.
//
// The value is clamped into [Min, Max] and onto the step before it is drawn
// and before anything else reads it, so a value that arrived from a stored
// record, a query or a test is put in range rather than drawn off the end of
// the rail. That clamp is the control's job, because the control is the only
// thing that knows the bounds.
func Slider(c *ui.Context, value *float64, opts SliderOptions) *ui.Element {
	if value == nil {
		panic("input: Slider needs a value to point at")
	}
	checkBounds(opts.Min, opts.Max, opts.Step, "Slider")
	if opts.Label == "" {
		panic("input: Slider needs options.Label; a rail and a thumb say nothing about what they measure")
	}
	*value = clampStep(*value, opts.Min, opts.Max, opts.Step)
	k, u := core.Tokens(c), core.Density(c).Unit()

	if !opts.ShowValue {
		return singleRail(c, value, opts)
	}
	// The rail is built inside the row rather than beside it and put in after:
	// MyGo gives a child to the parent whose Children call is running, so an
	// element made before the row exists is its sibling, and a value printed
	// under the rail instead of beside it.
	row := ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 2)
	row.Children(func() {
		// Grow is asked for here and not in singleRail, because a slider is a
		// child of a row in this branch and of whatever the caller built in
		// the other: grown in a column it would take the free space down the
		// page rather than across it, and a slider three hundred tall is not a
		// slider.
		singleRail(c, value, opts).Grow(1)
		ui.Text(c, writeValue(opts.Format, *value)).TextColor(k.TextMuted).
			FontSize(core.FontSize(c, theme.StatSize)).SingleLine()
	})
	return row
}

// singleRail is the rail of a Slider: MyGo's own, with a step on it when the
// caller asked for detents, named and either enabled or not. It is separate
// from Slider because the two layouts need it in different places — inside a
// row beside the number, or straight into whatever the caller built.
func singleRail(c *ui.Context, value *float64, opts SliderOptions) *ui.Element {
	// One of the two, chosen before either is made, and never both.
	// ui.Slider attaches what it builds to the parent it is given, so making
	// one and then making a StepSlider leaves the first rail in the tree
	// behind the second: a stepped slider would draw two of them, and the
	// whole suite would pass, because the labels, the value and the clicks
	// all come from the second.
	var rail *ui.Element
	if opts.Step > 0 {
		// StepSlider is the same control with detents, and it snaps while it
		// is dragged, so a thumb can never come to rest between two ticks.
		rail = ui.StepSlider(c, value, opts.Min, opts.Max, opts.Step)
	} else {
		rail = ui.Slider(c, value, opts.Min, opts.Max)
	}
	return rail.Label(opts.Label).Disabled(opts.Disabled)
}

// RangeSliderOptions configure a RangeSlider.
type RangeSliderOptions struct {
	// Min is the lowest value the range can reach and Max the highest.
	Min, Max float64
	// Step makes the values Min and the multiples of Step from it, with a
	// tick mark at each. Zero is every value in between.
	Step float64
	// Format is how each end of the range is written beside the slider.
	Format string
	// ShowValue puts the two numbers beside the slider.
	ShowValue bool
	// Label names the slider for assistive technology, and is required, as
	// Slider's is.
	Label string
	// Disabled greys the control out.
	Disabled bool
}

// RangeSlider is one rail with two thumbs, setting a low and a high, both of
// which it writes into the caller's floats.
//
// The two are kept in range and in order whatever they were handed: a filter
// asking for "between 90 and 10" is asking for nothing, and drawing it as it
// was asked would show two thumbs that had swapped places.
//
// Both thumbs are one ui.SliderBase. It is a single-value control, so the
// value it keeps is the thumb under the pointer rather than either end of the
// range, and each change goes to whichever end is nearer. That is what makes
// a pair of thumbs out of one control instead of two controls fighting over
// one press.
func RangeSlider(c *ui.Context, low, high *float64, opts RangeSliderOptions) *ui.Element {
	if low == nil || high == nil {
		panic("input: RangeSlider needs both ends of the range to point at")
	}
	checkBounds(opts.Min, opts.Max, opts.Step, "RangeSlider")
	if opts.Label == "" {
		panic("input: RangeSlider needs options.Label; a rail and two thumbs say nothing about what they measure")
	}
	orderRange(low, high, opts.Min, opts.Max, opts.Step)
	k, u := core.Tokens(c), core.Density(c).Unit()

	if !opts.ShowValue {
		host := ui.Box(c).FillWidth()
		rangeRail(c, host, low, high, opts)
		return host
	}
	// As in Slider, the rail is built where it is to sit rather than made
	// first and put in afterwards: an element made before its container is
	// that container's sibling, and a pair of numbers under the rail rather
	// than under it would be a second control wearing its numbers.
	col := ui.Column(c).FillWidth().Gap(u)
	col.Children(func() {
		rangeRail(c, ui.Box(c).FillWidth(), low, high, opts)
		ui.Text(c, writeValue(opts.Format, *low)+" – "+writeValue(opts.Format, *high)).
			TextColor(k.TextMuted).FontSize(core.FontSize(c, theme.StatSize)).SingleLine()
	})
	return col
}

// rangeRail is the rail of a RangeSlider: one ui.SliderBase with both thumbs
// painted on it, the value it keeps being the thumb under the pointer rather
// than either end of the range.
//
// host is the element the thumb is remembered on and the rail is built inside,
// which is how the two ends survive a frame: MyGo keys a local on its element,
// and the local has to outlive the element it is read from, so it cannot hang
// on the rail itself.
func rangeRail(c *ui.Context, host *ui.Element, low, high *float64, opts RangeSliderOptions) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	grabbed := *ui.Local(host, grabbedKey{}, func() *float64 { v := *low; return &v })

	// The detents need room under the rail, so a stepped rail is taller than
	// a plain one and the track sits above its middle.
	detents := float32(0)
	if opts.Step > 0 {
		detents = u*1.5 + 1
	}
	height := knobHeight + detents
	host.Children(func() {
		rail := ui.SliderBase(c, grabbed, opts.Min, opts.Max).
			Height(height).PaddingX(knobWidth / 2).FillWidth().
			FocusRing(false).Label(opts.Label).Disabled(opts.Disabled)
		if opts.Step > 0 {
			rail.Step(opts.Step)
		}
		if rail.Changed() {
			moveRangeEnd(grabbed, low, high, opts.Min, opts.Max, opts.Step)
		}
		lo, hi := fraction(*low, opts.Min, opts.Max), fraction(*high, opts.Min, opts.Max)
		rail.Draw(func(p *ui.Painter, r ui.Rect) {
			inset := knobWidth / 2
			cy := r.Y + height/2 - detents/2
			at := func(frac float32) float32 { return r.X + inset + (r.W-2*inset)*frac }
			from, to := at(lo), at(hi)
			p.Fill(ui.Rect{X: r.X, Y: cy - trackHeight/2, W: r.W, H: trackHeight}, k.Border, trackHeight/2)
			p.Fill(ui.Rect{X: from, Y: cy - trackHeight/2, W: to - from, H: trackHeight}, k.Accent, trackHeight/2)
			if opts.Step > 0 {
				paintTicks(p, k.TextMuted.Alpha(0.6), r, inset, cy+knobHeight/2+1, u*1.5, opts.Min, opts.Max, opts.Step)
			}
			for _, frac := range []float32{lo, hi} {
				thumb := ui.Rect{X: at(frac) - knobWidth/2, Y: cy - knobHeight/2, W: knobWidth, H: knobHeight}
				// Two shadows, as MyGo's own knob has: a tight one edges the
				// thumb and a soft one lifts it, so a white thumb still reads
				// against a white page.
				p.Shadow(thumb, knobHeight/2, 0, 0.5, 1, 0, ui.RGBA(0, 0, 0, 0.08))
				p.Shadow(thumb, knobHeight/2, 0, 1.5, 7, 0, ui.RGBA(0, 0, 0, 0.1))
				p.Fill(thumb, thumbFace(c), knobHeight/2)
				if rail.FocusVisible() {
					p.FocusRing(thumb, [4]float32{knobHeight / 2, knobHeight / 2, knobHeight / 2, knobHeight / 2})
				}
			}
		})
	})
	return host
}

// grabbedKey is where the thumb a RangeSlider is dragging lives. It is a type
// of its own so nothing else on the same element is found by its name.
type grabbedKey struct{}

// moveRangeEnd gives a drag to whichever end of the range is nearer, and
// re-anchors the grabbed value onto the end that took it, so that the next
// frame's distance is measured from where the thumb now is. A drag past an
// end stops at that end rather than pushing it along, which is what keeps the
// two from trading places.
func moveRangeEnd(grabbed *float64, low, high *float64, min, max, step float64) {
	v := clampStep(*grabbed, min, max, step)
	switch {
	case v <= *low:
		*low = v
	case v >= *high:
		*high = v
	case v-*low <= *high-v:
		*low = v
	default:
		*high = v
	}
	*grabbed = v
}

// orderRange puts both ends in range and low below high, which is what a
// range means and the only order a caller can be handed back in.
func orderRange(low, high *float64, min, max, step float64) {
	*low = clampStep(*low, min, max, step)
	*high = clampStep(*high, min, max, step)
	if *low > *high {
		*low, *high = *high, *low
	}
}

// checkBounds rejects the three sets of numbers a slider cannot do anything
// sensible with: a range with no room in it, a step that runs backwards, and
// a step wider than the whole range.
//
// That last one is worth stopping for. A 0..3 rail stepping by 10 has exactly
// one value it can take — its low end — so it draws as a slider, takes a
// pointer, answers the arrows, and never moves. A control that cannot move is
// a segmented control of one segment, or a Switch, and saying so here is
// cheaper than leaving someone to discover it by dragging.
func checkBounds(min, max, step float64, who string) {
	if min != min || max != max {
		panic("input: " + who + " was given a bound that is not a number")
	}
	if max <= min {
		panic("input: " + who + " needs Max above Min; a range with no room in it has nothing to slide")
	}
	if step < 0 {
		panic("input: " + who + " was given a negative Step")
	}
	if step > 0 && step > max-min {
		panic("input: " + who + " was given a Step wider than its range, which leaves the rail one value wide and nothing to slide between")
	}
}

// clampStep puts v in [min, max] and on a step from min.
//
// The order is the one ui/sliderBase uses — into the range first, onto the
// step second — and the arithmetic is MyGo's own, digit for digit, because a
// value this clamps has to be one the rail can already rest on. Snapping
// first and clamping after would leave the value at the top of the range
// whenever the step does not divide it: 11 on a 0..10 rail stepping by 4 would
// clamp to 10, which is not a step at all, and the first press would then move
// it to 8 and the value would jump the moment it was touched.
//
// That is also why a stepped rail stops short of its own top when the step
// does not divide the range. A rail that offers 0, 4 and 8 for a 0..10 range
// is telling the truth about the values it can take; one that also offered 10
// would be offering a value the first press would take away again.
func clampStep(v, min, max, step float64) float64 {
	v = math.Max(min, math.Min(max, v))
	if step <= 0 {
		return v
	}
	v = min + math.Round((v-min)/step)*step
	if v > max {
		// The rounding went past the top, so the last step that fits is the
		// one before it. A range narrower than a step has no step that fits
		// at all, and this leaves the value at the low end; checkBounds
		// refuses that range outright, so the rail is never asked to be one.
		v -= step
		if v < min {
			v = min
		}
	}
	// Away from the drift of floating point, as 0.1 steps add up: to the
	// decimals of the step and of the bound, as a browser rounds.
	if d := decimalsOf(step, min); d < 15 {
		v, _ = strconv.ParseFloat(strconv.FormatFloat(v, 'f', d, 64), 64)
	}
	return v
}

// decimalsOf is how many digits a number has after the point, the more of the
// values given.
func decimalsOf(values ...float64) int {
	d := 0
	for _, v := range values {
		mant, exp, _ := strings.Cut(strconv.FormatFloat(math.Abs(v), 'e', -1, 64), "e")
		e, _ := strconv.Atoi(exp)
		digits := 0
		if _, frac, ok := strings.Cut(mant, "."); ok {
			digits = len(frac)
		}
		if digits-e > d {
			d = digits - e
		}
	}
	return d
}

// fraction is where v is from min to max, from 0 to 1.
func fraction(v, min, max float64) float32 {
	if max <= min {
		return 0
	}
	return float32(math.Max(0, math.Min(1, (v-min)/(max-min))))
}

// paintTicks draws the detents of a stepped rail at y, tall, as MyGo's own
// slider draws them: hairlines under the track, one per step, and none at all
// where there would be too many of them to tell apart.
func paintTicks(p *ui.Painter, col ui.Color, r ui.Rect, inset, y, tall float32, min, max, step float64) {
	n := int(math.Floor((max-min)/step+1e-9)) + 1
	if n < 2 || n > 61 {
		return
	}
	for i := 0; i < n; i++ {
		x := r.X + inset + (r.W-2*inset)*float32(i)/float32(n-1)
		p.Fill(ui.Rect{X: x - 0.5, Y: y, W: 1, H: tall}, col, 0)
	}
}

// thumbFace is the face of a thumb: white where the window is light, and a
// pale grey where it is dark, because a white thumb on a dark window is a
// hole in it.
func thumbFace(c *ui.Context) ui.Color {
	if core.IsDark(c) {
		return ui.RGB(224, 225, 225)
	}
	return ui.RGB(255, 255, 255)
}

// writeValue is the number as the caller asked for it, and as itself when
// they did not ask.
func writeValue(format string, v float64) string {
	if format == "" {
		return strconv.FormatFloat(v, 'f', -1, 64)
	}
	return fmt.Sprintf(format, v)
}
