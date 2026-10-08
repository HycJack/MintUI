package input

import (
	"math"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// The controls that set a number by moving something rather than by typing
// it: a dial turned with a pointer, and a rail running down the window.
//
// Both are one control underneath, and both borrow their shape from
// ui/input's own Slider — the same capsule thumb, the same rail, the same
// rounding — so that a vertical meter beside a horizontal slider is two of
// one control rather than a new one.
//
// Neither is MyGo's SliderBase turned on its side. That control is laid out
// along the x axis and reads the pointer out of the element's own laid-out
// box; a vertical one built on it would be a horizontal slider with its parts
// swapped, and its own accessibility tree would still say "horizontal".

// The sweep of a dial, in degrees clockwise from straight up: the low end at
// the lower left, the high end at the lower right, and straight up in the
// middle of the two. The 90° gap at the bottom is deliberate — a dial that
// closed all the way round has no way to say which way is the middle, because
// every position is next to the last one, and a control whose low end is next
// to its high end is one a person turns past by accident.
const (
	sweepStart = 225.0
	sweepSpan  = 270.0
	sweepGap   = 90.0
	// radians turns the degrees above into what Painter's circle paths take,
	// so the arithmetic below carries a named factor rather than a bare
	// 0.01745 that nothing can be read against.
	radians = math.Pi / 180
)

// KnobOptions configure a Knob.
type KnobOptions struct {
	// Min is the lowest value the dial can be set to and Max the highest, and
	// they are pointers because zero is a bound people want: a gain starts at
	// one, a frequency at zero, and neither can say which it means with a
	// plain zero. Both nil is a dial of 0 to 1.
	Min, Max *float64
	// Default is what a value below Min comes up to. It is there because a
	// dial cannot have a value outside its own sweep, and a caller whose
	// stored setting is -3 and whose dial runs 0 to 10 would otherwise be
	// pinned at 0 with no way to tell that from a deliberate 0.
	Default float64
	// Step snaps the value, so a dial reads "8" rather than "8.372". Zero is
	// every value in between.
	Step float64
	// Format is how the value is written under the dial, as "%.0f dB".
	Format string
	// ShowValue puts the number under the dial. It is off by default: a dial
	// is for setting a thing roughly, and the exact number belongs in a field
	// beside it when it matters enough to be typed.
	ShowValue bool
	// Size is the dial's diameter; zero is four standard controls, which is
	// the smallest a dial can be and still be turned rather than nudged.
	Size float32
	// Label names the dial for assistive technology, and is required: a dial
	// draws a circle and a mark and no words, so without a name there is
	// nothing to read out and nothing saying what is being measured.
	Label string
	// Disabled greys it out.
	Disabled bool
}

// KnobResult carries a Knob and whether it was moved.
type KnobResult struct {
	// Element is the dial, or the dial and the number under it.
	Element *ui.Element
	changed bool
}

// Changed reports that the value moved under the pointer this frame, as
// opposed to the app having set it.
func (r KnobResult) Changed() bool { return r.changed }

// Knob is a dial: a circle with a mark on it, turned by dragging around it,
// writing a number into the caller's float.
//
// The value is read from where the pointer is rather than from how far it
// moved, which is what makes a dial feel like a dial. A delta-based dial
// cannot be turned to a position — there is no position to turn to — so it
// would be a slider wearing a round face, and every time somebody reached for
// it to set a value they already knew, they would have had to count the way
// round to get it.
func Knob(c *ui.Context, value *float64, opts KnobOptions) KnobResult {
	if value == nil {
		panic("input: Knob needs a value to point at; it keeps no value of its own")
	}
	if opts.Label == "" {
		panic("input: Knob needs options.Label; a circle and a mark say nothing about what " +
			"they measure")
	}
	if opts.Min != nil && opts.Max != nil && *opts.Min >= *opts.Max {
		panic("input: Knob has a Min at or above its Max, which sweeps nothing")
	}
	lo, hi := knobBounds(opts.Min, opts.Max)
	*value = clampStep(*value, lo, hi, opts.Step)
	if *value < lo || *value > hi {
		// Out of the sweep: the caller's stored value came from somewhere
		// else, and the dial is the only thing that knows its own bounds.
		fall := lo
		if opts.Default > lo && opts.Default < hi {
			fall = opts.Default
		}
		*value = fall
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	side := opts.Size
	if side <= 0 {
		side = core.ControlHeight(c) * 4
	}

	var r KnobResult
	// The face is a box of the dial's own size and nothing else, so the box a
	// test measures is the dial itself. An element whose size is whatever its
	// content happens to need measures zero, which is the trap
	// docs/design-system.md §15.2 is about, and a dial is exactly the case
	// that trap catches: it draws a circle and has no content at all.
	face := func() *ui.Element {
		e := ui.Box(c).Size(side, side).Shrink(0).Radius(side/2).Focusable().
			Role(ui.RoleSlider).Label(opts.Label).Tooltip(opts.Label).
			Disabled(opts.Disabled).Range(lo, hi, *value)
		e.Children(func() {
			// Dragged is what says the pointer is down on this element even
			// after it has left the box, which is how a dial is turned past
			// its own edge rather than sticking when the hand runs out of
			// room. PointerPosition still gives the position, outside or in.
			if _, _, held := e.Dragged(); held && !opts.Disabled {
				x, y, _ := e.PointerPosition()
				frac := knobTurn(x, y, side)
				r.changed = setDial(value, clampStep(lo+float64(frac)*(hi-lo), lo, hi, opts.Step)) || r.changed
			}
			if !opts.Disabled {
				switch {
				case e.Shortcut(0, ui.KeyUp), e.Shortcut(0, ui.KeyRight):
					if setDial(value, clampStep(*value+dialStep(opts.Step, lo, hi), lo, hi, opts.Step)) {
						r.changed = true
					}
				case e.Shortcut(0, ui.KeyDown), e.Shortcut(0, ui.KeyLeft):
					if setDial(value, clampStep(*value-dialStep(opts.Step, lo, hi), lo, hi, opts.Step)) {
						r.changed = true
					}
				case e.Shortcut(0, ui.KeyHome):
					r.changed = setDial(value, clampStep(lo, lo, hi, opts.Step)) || r.changed
				case e.Shortcut(0, ui.KeyEnd):
					r.changed = setDial(value, clampStep(hi, lo, hi, opts.Step)) || r.changed
				}
			}
			knobFace(c, e, side, fraction(*value, lo, hi), opts.Disabled)
		})
		return e
	}

	if !opts.ShowValue {
		r.Element = face()
		return r
	}
	col := ui.Column(c).FillWidth().AlignItems(ui.Center).Gap(u).Label(opts.Label)
	col.Children(func() {
		face()
		ui.Text(c, writeValue(opts.Format, *value)).TextColor(k.TextMuted).
			FontSize(core.FontSize(c, theme.StatSize)).SingleLine()
	})
	r.Element = col
	return r
}

// knobBounds resolves a dial's two ends. Both may be left out, which is a
// dial of 0 to 1 — the shape a percentage or a unit fraction wants, and the
// case where writing two bounds buys nothing.
func knobBounds(min, max *float64) (lo, hi float64) {
	if min != nil {
		lo = *min
	}
	if max != nil {
		hi = *max
	}
	if hi <= lo {
		hi = lo + 1
	}
	return lo, hi
}

// setDial writes a dial's value and says whether it moved, so that a caller
// reading Changed does not have to remember which half of the control wrote.
func setDial(value *float64, next float64) bool {
	if next == *value {
		return false
	}
	*value = next
	return true
}

// dialStep is how far the arrows move a dial: its own step, or a fiftieth of
// its sweep when it has none. A fiftieth is small enough that holding the key
// walks to a number rather than running to the end of the dial, and large
// enough that it moves at all — a dial whose arrow key does nothing is a dial
// only a hand with a pointer can use.
func dialStep(step, lo, hi float64) float64 {
	if step > 0 {
		return step
	}
	return (hi - lo) / 50
}

// knobTurn is how far round the dial a pointer at (x, y) in its own box is, as
// a fraction of the sweep from 0 to 1.
func knobTurn(x, y, side float32) float32 {
	// The angle is about the middle of the box rather than about its corner,
	// and the whole thing is turned by ninety degrees first because the sweep
	// is measured from straight up while an angle from atan2 is measured
	// from straight right.
	deg := math.Atan2(float64(y-side/2), float64(x-side/2))/radians + 90
	turn := deg - sweepStart
	for turn < 0 {
		turn += 360
	}
	for turn > 360 {
		turn -= 360
	}
	// In the gap the pointer is snapped to whichever end it is nearer, which
	// is what a rotary control does and is the reason the gap can be there at
	// all: every value in it belongs to both ends, so a dial that did
	// something else with it would either leap to an end or swing the whole
	// way round.
	switch {
	case turn <= sweepSpan:
	case turn-sweepSpan >= sweepGap/2:
		turn = 0
	default:
		turn = sweepSpan
	}
	return float32(turn / sweepSpan)
}

// knobFace draws the dial: the track, the part of it that is filled, and the
// mark at the value.
//
// It is painted rather than assembled out of boxes because the filled part is
// one arc. Three boxes with one of them rotated between them is an arc made
// of three straight pieces and two corners between them, and a corner on a
// dial is the one thing that makes it read as a drawing of a dial.
func knobFace(c *ui.Context, face *ui.Element, side, frac float32, disabled bool) {
	k := core.Tokens(c)
	track, ink := k.Border, k.Accent
	if disabled {
		track, ink = k.Border.Alpha(0.55), k.TextFaint
	}
	face.Draw(func(p *ui.Painter, r ui.Rect) {
		cx, cy := r.X+r.W/2, r.Y+r.H/2
		width := max(1, r.W*0.045)
		ring := r.W/2 - width/2 - 1

		sweepPath(p, cx, cy, ring, 0, 1, width, track)
		if frac > 0 {
			sweepPath(p, cx, cy, ring, 0, frac, width, ink)
		}

		// The mark runs from the middle out to just inside the ring rather
		// than sitting on it: a dot exactly on the ring is half hidden by the
		// ring it is on, and half of the value's own mark is not a mark.
		angle := (sweepStart + frac*sweepSpan) * radians
		sin, cos := math.Sincos(float64(angle))
		p.Line(cx, cy, cx+ring*float32(sin), cy-ring*float32(cos), width, ink)
		p.Fill(ui.Rect{X: cx - width*0.8, Y: cy - width*0.8, W: width * 1.6, H: width * 1.6},
			ink, width*0.8)
	})
	if face.FocusVisible() {
		face.DrawOver(func(p *ui.Painter, r ui.Rect) {
			p.FocusRing(r, [4]float32{r.W / 2, r.W / 2, r.W / 2, r.W / 2})
		})
	}
}

// sweepPath paints the part of a dial's ring between two fractions of the
// sweep.
//
// Ninety-six steps over 270° is one every under three degrees, which at any
// size this library draws a dial is under a pixel. A coarser path shows its
// own corners on the curve, and a finer one costs frames for a difference
// nobody can see.
func sweepPath(p *ui.Painter, cx, cy, rad, from, to, width float32, col ui.Color) {
	if to <= from {
		return
	}
	const steps = 96
	var path ui.Path
	for i := range steps + 1 {
		a := (sweepStart + sweepSpan*(from+(to-from)*float32(i)/float32(steps))) * radians
		sin, cos := math.Sincos(float64(a))
		x, y := cx+rad*float32(sin), cy-rad*float32(cos)
		if i == 0 {
			path.MoveTo(x, y)
			continue
		}
		path.LineTo(x, y)
	}
	p.StrokePath(&path, width, col)
}

// VerticalSliderOptions configure a VerticalSlider.
type VerticalSliderOptions struct {
	// Min is the lowest value the rail can be set to and Max the highest.
	// They are pointers for Knob's reason: zero is a bound people want.
	Min, Max *float64
	// Step snaps the value and is what the arrows, Home and End move by.
	// Zero is every value in between.
	Step float64
	// Format is how the value is written beside the rail, as "%.0f%%".
	Format string
	// ShowValue puts the value beside the rail, which is off by default for
	// the reason Slider's is: a form with its own field for the number would
	// be saying it twice.
	ShowValue bool
	// Height is how tall the rail is; zero is six standard controls, the
	// shortest a rail can be and still read as a level rather than as a bar.
	Height float32
	// Width is the rail's own width; zero is six density units, which is the
	// same proportion as a horizontal rail's height.
	Width float32
	// Label names the rail for assistive technology, and is required, as
	// Slider's is: a rail and a thumb say nothing about what they measure.
	Label string
	// Disabled greys the control out.
	Disabled bool
}

// VerticalSliderResult carries a VerticalSlider and whether it was moved.
type VerticalSliderResult struct {
	// Element is the rail, or the rail and the number beside it.
	Element *ui.Element
	changed bool
}

// Changed reports that the value moved under the pointer or the keys this
// frame.
func (r VerticalSliderResult) Changed() bool { return r.changed }

// VerticalSlider is a rail running down the window, with the largest value at
// the top.
func VerticalSlider(c *ui.Context, value *float64, opts VerticalSliderOptions) VerticalSliderResult {
	if value == nil {
		panic("input: VerticalSlider needs a value to point at; it keeps no value of its own")
	}
	if opts.Label == "" {
		panic("input: VerticalSlider needs options.Label; a rail and a thumb say nothing " +
			"about what they measure")
	}
	if opts.Min != nil && opts.Max != nil && *opts.Min >= *opts.Max {
		panic("input: VerticalSlider has a Min at or above its Max, which slides nothing")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	lo, hi := 0.0, 1.0
	if opts.Min != nil {
		lo = *opts.Min
	}
	if opts.Max != nil {
		hi = *opts.Max
	}
	*value = clampStep(*value, lo, hi, opts.Step)

	h := opts.Height
	if h <= 0 {
		h = core.ControlHeight(c) * 6
	}
	w := opts.Width
	if w <= 0 {
		w = u * 6
	}
	// The thumb overhangs the rail at both ends, so the track inside the rail's
	// own box is shorter by half a thumb at each end: otherwise the value at
	// the very top would draw its thumb half outside the control, and the top
	// of a meter is the reading people look for first.
	inset := knobHeight / 2

	var r VerticalSliderResult
	rail := func() *ui.Element {
		e := ui.Box(c).Width(w).Height(h).Shrink(0).Focusable().
			Role(ui.RoleSlider).Label(opts.Label).Tooltip(opts.Label).
			Disabled(opts.Disabled).Range(lo, hi, *value)
		e.Children(func() {
			if _, _, held := e.Dragged(); held && !opts.Disabled {
				_, y, _ := e.PointerPosition()
				frac := 1 - clamp01f(y/h)
				r.changed = setDial(value, clampStep(lo+float64(frac)*(hi-lo), lo, hi, opts.Step)) || r.changed
			}
			// The arrows, Home and End, read from the rail's own box: a rail
			// with no keys on it is a control that can only be used by a hand
			// with a pointer, and a number is the one thing on a screen that
			// somebody may well be typing rather than aiming at.
			if !opts.Disabled {
				switch {
				case e.Shortcut(0, ui.KeyUp), e.Shortcut(0, ui.KeyRight):
					r.changed = setDial(value, clampStep(*value+opts.Step, lo, hi, opts.Step)) || r.changed
				case e.Shortcut(0, ui.KeyDown), e.Shortcut(0, ui.KeyLeft):
					r.changed = setDial(value, clampStep(*value-opts.Step, lo, hi, opts.Step)) || r.changed
				case e.Shortcut(0, ui.KeyHome):
					r.changed = setDial(value, clampStep(lo, lo, hi, opts.Step)) || r.changed
				case e.Shortcut(0, ui.KeyEnd):
					r.changed = setDial(value, clampStep(hi, lo, hi, opts.Step)) || r.changed
				}
			}
			verticalRail(c, e, inset, fraction(*value, lo, hi), opts.Disabled)
		})
		return e
	}

	if !opts.ShowValue {
		r.Element = rail()
		return r
	}
	col := ui.Column(c).FillWidth().AlignItems(ui.Center).Gap(u).Label(opts.Label)
	col.Children(func() {
		rail()
		ui.Text(c, writeValue(opts.Format, *value)).TextColor(k.TextMuted).
			FontSize(core.FontSize(c, theme.StatSize)).SingleLine()
	})
	r.Element = col
	return r
}

func clamp01f(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// verticalRail paints the rail of a VerticalSlider: the track, the filled part
// above the thumb, and the thumb.
func verticalRail(c *ui.Context, rail *ui.Element, inset, frac float32, disabled bool) {
	k := core.Tokens(c)
	track := k.Border
	thumb := thumbFace(c)
	if disabled {
		track, thumb = k.Border.Alpha(0.55), k.SurfaceHover
	}
	rail.Draw(func(p *ui.Painter, r ui.Rect) {
		cx := r.X + r.W/2
		top, bottom := r.Y+inset, r.Y+r.H-inset
		if bottom <= top {
			return
		}
		p.Fill(ui.Rect{X: cx - trackHeight/2, Y: top, W: trackHeight, H: bottom - top},
			track, trackHeight/2)

		// Filled from the thumb up, not from the top down: a meter filled
		// downwards from its maximum reads as "this much is left", and the
		// thing a meter shows is how much is reached.
		filled := (bottom - top) * frac
		ink := k.Accent
		if disabled {
			ink = ink.Alpha(0.45)
		}
		p.Fill(ui.Rect{X: cx - trackHeight/2, Y: bottom - filled, W: trackHeight, H: filled},
			ink, trackHeight/2)

		box := ui.Rect{X: cx - knobWidth/2, Y: bottom - filled - knobHeight/2, W: knobWidth, H: knobHeight}
		// Two shadows, as ui/input's own slider draws its thumb: a tight one
		// edges it and a soft one lifts it, so a pale thumb still reads
		// against a pale rail.
		p.Shadow(box, knobHeight/2, 0, 0.5, 1, 0, ui.RGBA(0, 0, 0, 0.08))
		p.Shadow(box, knobHeight/2, 0, 1.5, 7, 0, ui.RGBA(0, 0, 0, 0.1))
		p.Fill(box, thumb, knobHeight/2)
		if rail.FocusVisible() {
			p.FocusRing(box, [4]float32{knobHeight / 2, knobHeight / 2, knobHeight / 2, knobHeight / 2})
		}
	})
}
