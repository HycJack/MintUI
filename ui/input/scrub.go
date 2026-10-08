package input

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// ScrubInput is a value you change by dragging sideways under the pointer,
// with no rail, no thumb and no track.
//
// It is the control for the numbers that sit in the middle of a canvas — a
// layer's opacity, a photo's warmth, a handle's angle — where the number is
// already on screen and the job is nudging it, not setting it. A slider there
// would put a rail between the person and the thing they are changing, and a
// number field would ask them to type a value they can only see.
//
// The number is the caller's float and is clamped and stepped into range on
// the way in, on the same terms as ui/input's Slider: the control is the only
// thing that knows the bounds.

// ScrubOptions configure a Scrub.
type ScrubOptions struct {
	// Min is the lowest value a drag can reach and Max the highest, and they
	// are pointers because zero is a bound people want: a temperature runs
	// below zero, an opacity does not, and neither can say which it means
	// with a plain zero.
	Min, Max *float64
	// Step snaps the value. Zero is every value in between, which for a
	// control nobody is typing into is right: the whole point is nudging.
	Step float64
	// Sensitivity is how much a pixel of drag is worth, as a fraction of the
	// range. Zero is a fiftieth: fifty pixels end to end, which is about the
	// width of a track of thumb, so a drag of the hand's length covers the
	// whole range without a value ever flying past it.
	Sensitivity float32
	// Format is how the value is written on the control, as "%.0f%%". Empty
	// writes the number on its own.
	Format string
	// Prefix sits before the number and Suffix after it, for a unit that is
	// decoration rather than part of the value.
	Prefix, Suffix string
	// Width is the control's own width; zero is what its words need.
	Width float32
	// Label names the control for assistive technology, and is required: a
	// drag has no visible affordance at all — not even a rail — so without a
	// name there is nothing to read out and nothing saying what it measures.
	Label string
	// Disabled greys the control out and takes it out of the tab order.
	Disabled bool
}

// ScrubResult carries a Scrub and whether it moved.
type ScrubResult struct {
	// Element is the control.
	Element *ui.Element
	changed bool
}

// Changed reports that the value moved under the drag this frame.
func (r ScrubResult) Changed() bool { return r.changed }

// ScrubInput is a number changed by dragging across it, which writes into the
// caller's float.
//
// It has no visual track on purpose. The number is the display and the drag is
// the input: a control on a canvas that put a rail over the thing being
// adjusted would be covering the very pixels the person is looking at to judge
// the result.
//
// Right and left move it by the step, because a scrub with no keyboard is a
// control that can only be used by a hand with a pointer — and a drag with the
// pointer is the one adjustment on a screen that is hard to make exactly.
func ScrubInput(c *ui.Context, value *float64, opts ScrubOptions) ScrubResult {
	if value == nil {
		panic("input: ScrubInput needs a value to point at; it keeps no value of its own")
	}
	if opts.Label == "" {
		panic("input: ScrubInput needs options.Label; a drag has no visible affordance at " +
			"all, so without a name there is nothing to read out")
	}
	lo, hi := 0.0, 1.0
	if opts.Min != nil {
		lo = *opts.Min
	}
	if opts.Max != nil {
		hi = *opts.Max
	}
	if hi <= lo {
		panic("input: ScrubInput has a Min at or above its Max, which scrubs nothing")
	}
	*value = clampStep(*value, lo, hi, opts.Step)
	k, u := core.Tokens(c), core.Density(c).Unit()

	perPixel := opts.Sensitivity
	if perPixel <= 0 {
		perPixel = 0.02
	}

	var r ScrubResult
	e := ui.Box(c).FillWidth().Height(core.ControlHeight(c)).Shrink(0).Focusable().
		Radius(theme.SmallRadius).Background(k.Surface).
		BorderWidth(theme.BorderWidth).BorderColor(k.Border).
		Role(ui.RoleSlider).Label(opts.Label).Tooltip(opts.Label).
		Disabled(opts.Disabled).Range(lo, hi, *value)
	if opts.Width > 0 {
		e.Width(opts.Width)
	}
	e.Children(func() {
		// The cursor is the whole of the affordance, and it changes to the one
		// that means "this moves sideways" — which is a thing this control
		// has to say out loud, since it has no track to say it with.
		e.Cursor(ui.CursorResizeEW)
		if dx, _, held := e.Dragged(); held && !opts.Disabled {
			r.changed = moveScrub(value, float64(dx)*float64(perPixel)*(hi-lo), lo, hi, opts.Step) || r.changed
		}
		if !opts.Disabled {
			switch {
			case e.Shortcut(0, ui.KeyRight), e.Shortcut(0, ui.KeyUp):
				r.changed = moveScrub(value, scrubStep(opts.Step, lo, hi), lo, hi, opts.Step) || r.changed
			case e.Shortcut(0, ui.KeyLeft), e.Shortcut(0, ui.KeyDown):
				r.changed = moveScrub(value, -scrubStep(opts.Step, lo, hi), lo, hi, opts.Step) || r.changed
			}
		}
		// The face while the drag is going, which is the one piece of feedback
		// the control gives: a scrub that looks identical before and during a
		// drag gives nothing to say "yes, that is the thing you are holding".
		switch {
		case e.Pressed() && !opts.Disabled:
			e.Background(k.AccentBg).BorderColor(k.Accent)
		case e.Hovered() && !opts.Disabled:
			e.Background(k.SurfaceHover)
		case opts.Disabled:
			e.Background(k.Surface)
		}
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 0.75).Children(func() {
			ink := k.Text
			if opts.Disabled {
				ink = k.TextFaint
			}
			if opts.Prefix != "" {
				fieldText(c, opts.Prefix, k.TextMuted)
			}
			ui.Text(c, writeValue(opts.Format, *value)).TextColor(ink).
				FontSize(core.FontSize(c, theme.StatSize)).SingleLine()
			if opts.Suffix != "" {
				fieldText(c, opts.Suffix, k.TextMuted)
			}
		})
	})
	r.Element = e
	return r
}

// moveScrub gives the value a drag of delta and puts it back in range and on
// the step, saying whether it moved at all.
func moveScrub(value *float64, delta, lo, hi, step float64) bool {
	next := clampStep(*value+delta, lo, hi, step)
	if next == *value {
		return false
	}
	*value = next
	return true
}

// scrubStep is how far the arrow keys move a scrub that was given no step: a
// fiftieth of its range, the same share of the range a pixel of drag is worth
// by default. The keys and the pointer therefore move the value by the same
// amount per gesture, which is what makes the two ways of using it feel like
// one control.
func scrubStep(step, lo, hi float64) float64 {
	if step > 0 {
		return step
	}
	return (hi - lo) / 50
}

// SignaturePoint is one sample of a stroke: where it was, in the pad's own
// coordinates, not in the window's.
type SignaturePoint struct {
	X, Y float32
}

// Stroke is one mark the person made: from the point the pointer went down to
// the point it came up.
//
// It is a list rather than a path, because a signature is sampled and not
// drawn: the person moves faster than the window can paint, and a stroke
// recorded as a path would lose every point between two frames. Keeping the
// samples means a signature can be replayed, redrawn at another size, or sent
// to a server that wants the same points the person made.
type Stroke struct {
	// Points are the samples, in the order they were taken. The first is
	// where the pointer went down and the last is where it came up.
	Points []SignaturePoint
}

// SignaturePadOptions configure a SignaturePad.
type SignaturePadOptions struct {
	// Label names the pad for assistive technology, and is required: a pad is
	// a blank rectangle, and a blank rectangle is the one control in this
	// library that says nothing at all about what it is for.
	Label string
	// Width and Height are the pad's own size. A pad with no height would be
	// a line, and a signature needs room to sign in, so the height defaults
	// and the width fills the row.
	Width, Height float32
	// Ink is what the strokes are drawn with. Zero alpha takes the window's
	// text colour, which is the only colour that reads as ink in both
	// appearances — a signature is drawn on whatever the pad is sitting on
	// and cannot ask what that is.
	Ink ui.Color
	// Clearable puts a button under the pad that empties it, because a
	// signature is the one field in this library whose value is not a
	// pointer to a string and so has no other way back to nothing.
	Clearable bool
	// Disabled greys the pad out and takes its strokes.
	Disabled bool
}

// SignaturePadResult carries a SignaturePad and what was done on it.
type SignaturePadResult struct {
	// Element is the pad, with the clear button under it when there is one.
	Element *ui.Element
	// inked is that the pad holds at least one stroke as of this frame.
	inked bool
	// cleared reports that the clear button was pressed this frame.
	cleared bool
}

// Inked reports that the pad holds at least one stroke, which is what a form
// asks before it submits a signature. It is not the same as len(strokes) > 0
// read after the frame: a stroke taken in this frame is in the caller's slice
// immediately, and the answer a form needs is the one that was true when it
// looked.
func (r SignaturePadResult) Inked() bool { return r.inked }

// Cleared reports that the clear button was pressed this frame.
func (r SignaturePadResult) Cleared() bool { return r.cleared }

// SignaturePad is a surface a person signs on, and the strokes they make.
//
// The strokes are the caller's slice and nothing is kept here. That is the
// whole reason the type is a slice of strokes rather than a drawing: a
// signature is personal data, it has to be sent somewhere, and the library
// holding the only copy of it would mean a caller who wanted to keep it had
// to ask the pad to give it back.
//
// A stroke ends when the pointer goes up, and a new one begins when it goes
// down again — including when it goes down somewhere else on the pad, because
// that is two letters and not one.
func SignaturePad(c *ui.Context, strokes *[]Stroke, opts SignaturePadOptions) SignaturePadResult {
	if strokes == nil {
		panic("input: SignaturePad needs a slice of strokes to point at; it holds no " +
			"signature of its own, and a signature it alone held would be one nobody could save")
	}
	if opts.Label == "" {
		panic("input: SignaturePad needs a Label; a blank rectangle is the one control here " +
			"that says nothing about what it is for")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	ink := opts.Ink
	if ink.A == 0 {
		ink = k.Text
	}
	h := opts.Height
	if h <= 0 {
		h = core.ControlHeight(c) * 5
	}

	var r SignaturePadResult
	// The pad is a closure because it both holds the stroke being drawn and
	// paints it: an element has to exist before its own Draw can name it, and
	// the store it hangs that stroke on has to hang off the pad itself, which
	// MyGo only keeps across frames while the pad is in the same place.
	pad := func() *ui.Element {
		e := ui.Box(c).FillWidth().Height(h).Shrink(0).Radius(theme.ControlRadius).
			Background(k.Background).BorderWidth(theme.BorderWidth).BorderColor(k.Border).
			Cursor(ui.CursorPointer).Role(ui.RoleNone).Label(opts.Label).
			Disabled(opts.Disabled)
		if opts.Width > 0 {
			e.Width(opts.Width)
		}
		return e.Children(func() {
			r.inked = len(*strokes) > 0

			// The stroke being drawn is held on the pad rather than in the
			// caller's slice: it is not a stroke yet — it has no end — and
			// appending a partial one would mean a signature saved while
			// somebody is still writing it.
			live := *ui.Local(e, signatureKey{}, func() *Stroke { return &Stroke{} })
			if _, _, held := e.Dragged(); held && !opts.Disabled {
				x, y, _ := e.PointerPosition()
				live.Points = append(live.Points, SignaturePoint{X: x, Y: y})
			} else if len(live.Points) > 0 {
				// The pointer came up, so the stroke has an end now and
				// belongs to the caller. A press that never moved is still
				// kept, because a signature is sometimes one deliberate dot
				// and one dot is a signature.
				*strokes = append(*strokes, *live)
				*live = Stroke{}
				// The slice grew, so the caller's view of it has to be
				// redrawn: nothing the user did would ask for that frame.
				c.Invalidate()
			}

			e.Draw(func(p *ui.Painter, r ui.Rect) {
				drawSignature(p, r, *strokes, *live, ink)
			})
			// The hint is inside the pad and under the strokes: it is what
			// the pad says while it is empty, and an empty pad is the only
			// state it is ever said in.
			if !r.inked && !opts.Disabled {
				ui.Text(c, core.Msg(c, "input.signHere", core.Def("Sign here"))).
					TextColor(k.TextFaint).FontSize(core.FontSize(c, theme.RowSize))
			}
		})
	}

	col := ui.Column(c).FillWidth().Gap(u).AlignItems(ui.Center)
	col.Children(func() {
		pad()
		if opts.Clearable {
			clear := core.Msg(c, "input.clearSignature", core.Def("Clear signature"))
			btn := ui.ButtonBase(c).Height(core.ControlHeight(c)-u*2).Shrink(0).
				Radius(theme.PillRadius).Padding(0, u*3, 0, u*3).
				Background(k.Surface).TextColor(k.TextMuted).
				Role(ui.RoleButton).Label(clear).Tooltip(clear).Disabled(opts.Disabled)
			if btn.Clicked() {
				*strokes = nil
				r.cleared = true
				c.Invalidate()
			}
			btn.Children(func() {
				ui.Text(c, clear).SingleLine().
					FontSize(core.FontSize(c, theme.RowSize)).TextColor(k.TextMuted)
			})
		}
	})
	r.Element = col
	return r
}

// signatureKey is where the stroke being drawn lives. A type of its own, so
// that nothing else on the pad is found by the same name.
type signatureKey struct{}

// drawSignature paints the strokes and the one being drawn, as a run of round
// segments rather than as a path through the points.
//
// The segments matter: a polyline through a signature's samples shows its
// corners, and a signature is the one drawing in an interface where a corner
// is read as a wrong letter. A round cap at each sample is what a pen leaves.
func drawSignature(p *ui.Painter, box ui.Rect, strokes []Stroke, live Stroke, ink ui.Color) {
	const width = 2.5
	for _, s := range strokes {
		signatureRun(p, box, s.Points, ink, width)
	}
	signatureRun(p, box, live.Points, ink, width)
}

// signatureRun paints one stroke's points as a run of segments with a dot at
// each end, so a stroke of one point is a dot rather than nothing at all.
//
// box is the pad's own rectangle and it is not optional. SignaturePoint holds
// the pointer's position **relative to the pad** — that is what
// PointerPosition returns — while ui.Painter works in window coordinates. So
// the run has to be offset by the pad's origin, or every signature, drawn or
// restored from storage, lands at the window's top-left corner instead of in
// the box the person is signing in.
func signatureRun(p *ui.Painter, box ui.Rect, points []SignaturePoint, ink ui.Color, width float32) {
	if len(points) == 0 {
		return
	}
	at := func(i int) ui.Rect {
		return ui.Rect{
			X: box.X + points[i].X - width/2,
			Y: box.Y + points[i].Y - width/2,
			W: width, H: width,
		}
	}
	for i := range points {
		cur := at(i)
		p.Fill(cur, ink, width/2)
		if i == 0 {
			continue
		}
		prev := at(i - 1)
		p.Line(prev.X, prev.Y, cur.X, cur.Y, width, ink)
	}
}
