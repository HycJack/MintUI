package input

import (
	"fmt"
	"math"
	"strconv"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// The controls that choose a value nobody can type: a number of stars, a
// color, a set of words. They share one problem the rest of the package does
// not — what they draw is a color the caller chose, and no theme token can
// be relied on to read on top of it.

// RatingOptions configure a Rating.
type RatingOptions struct {
	// Max is how many stars. Five is the number a person expects, so it is
	// what an unset Max gets.
	Max int
	// NoClear stops a press on the value already given from taking it away.
	// It is off by default, because a rating that can only be lowered by
	// first reaching the bottom is a rating that cannot be left unset: a
	// form has to submit something, and "no opinion" is one of the answers.
	NoClear bool
	// ReadOnly shows the value and takes no presses, for a card showing
	// someone else's rating.
	ReadOnly bool
	// Label names the control for assistive technology, and is required: the
	// stars are marks and there are no words beside them, so without a name
	// a rating reads out as nothing at all.
	Label string
	// Disabled greys the control out.
	Disabled bool
}

// Rating is a row of stars choosing a number, which it writes into the
// caller's int.
//
// A press on the value already given takes it back to nothing, unless
// NoClear is set: without that the lowest a rating could go would be one star,
// and a form would have to invent a value for the person who had not rated
// yet.
//
// The stars take ink rather than a colour. A star is not a priority and not a
// state, and the one bright colour in this library is the presence dot, which
// is the only thing that is allowed to be that bright.
func Rating(c *ui.Context, value *int, opts RatingOptions) *ui.Element {
	if value == nil {
		panic("input: Rating needs a value to point at")
	}
	if opts.Label == "" {
		panic("input: Rating needs options.Label; its stars are marks with no words beside them, so there is nothing to read out")
	}
	max := opts.Max
	if max == 0 {
		max = 5
	}
	if max < 0 {
		panic("input: Rating Max cannot be negative")
	}
	// Clamped rather than refused: a rating out of range is a number from
	// somewhere else — an average, a stored value — and the control is the
	// only thing that knows how many stars there are.
	*value = int(math.Max(0, math.Min(float64(max), float64(*value))))
	k, u := core.Tokens(c), core.Density(c).Unit()

	side := u * 6
	row := ui.Row(c).Gap(u * 0.5).AlignItems(ui.Center).Label(opts.Label)
	row.Children(func() {
		for i := 1; i <= max; i++ {
			// The focus ring is drawn below, in the star's own shape, so
			// MyGo's box-shaped one is turned off: two rings round a star is
			// one ring and one stray rectangle.
			star := ui.ButtonBase(c).Size(side, side).Shrink(0).FocusRing(false).
				Label(strconv.Itoa(i) + " of " + strconv.Itoa(max)).
				Tooltip(strconv.Itoa(i)).Disabled(opts.Disabled || opts.ReadOnly)
			if star.Clicked() {
				switch {
				case opts.NoClear:
					*value = i
				case *value == i:
					*value = 0
				default:
					*value = i
				}
			}
			full := i <= *value
			fg := k.Border.Mix(k.Text, 0.45)
			if full {
				fg = k.Text
			}
			if opts.Disabled {
				fg = k.TextFaint
			}
			star.Draw(func(p *ui.Painter, r ui.Rect) {
				drawStar(p, r, fg, full)
				if star.FocusVisible() {
					p.FocusRing(r, [4]float32{theme.SmallRadius, theme.SmallRadius, theme.SmallRadius, theme.SmallRadius})
				}
			})
		}
	})
	return row
}

// drawStar draws a five-pointed star in r, filled or as an outline.
//
// The inner radius is the pentagram's 0.382, which is what makes ten points
// on two circles read as one star rather than as a cog.
func drawStar(p *ui.Painter, r ui.Rect, col ui.Color, filled bool) {
	cx, cy := r.X+r.W/2, r.Y+r.H/2
	outer, inner := r.W/2, r.W*0.191
	var path ui.Path
	for i := 0; i < 10; i++ {
		rad := outer
		if i%2 == 1 {
			rad = inner
		}
		angle := math.Pi/2 + float64(i)*math.Pi/5
		x, y := cx+float32(math.Cos(angle))*rad, cy-float32(math.Sin(angle))*rad
		if i == 0 {
			path.MoveTo(x, y)
			continue
		}
		path.LineTo(x, y)
	}
	path.Close()
	if filled {
		p.FillPath(&path, col)
		return
	}
	p.StrokePath(&path, 1.5, col)
}

// Swatch is one color of a palette.
type Swatch struct {
	// Color is the color itself.
	Color ui.Color
	// Name is what the swatch is called, to assistive technology and in its
	// tooltip. It is required: a swatch shows a color and nothing else, so a
	// swatch with no name is read out as nothing — worse than a wrong name,
	// which at least is one a person can correct.
	Name string
}

// ColorPaletteOptions configure a ColorPalette.
type ColorPaletteOptions struct {
	// Label names the group for assistive technology: "Colour", which the
	// swatches cannot say between them. It is the only text this control has,
	// so an empty one leaves a run of colors that announces as nothing.
	Label string
	// Disabled greys out every swatch.
	Disabled bool
}

// ColorPalette is a grid of colors choosing one of them, which it writes into
// the caller's ui.Color.
//
// The chosen swatch is marked twice over, and the two marks ask different
// things of the color. The ring is the ink that reads on the surface the
// palette sits on, and the tick is the ink that reads on the swatch. One ink
// cannot do both jobs: in a dark window a dark swatch and the dark surface
// under it are close enough that a mark in the window's text colour sits at
// the same value as the swatch and disappears into it, and the very same
// swatch on a light window wants the opposite ink.
func ColorPalette(c *ui.Context, selected *ui.Color, swatches []Swatch, opts ColorPaletteOptions) *ui.Element {
	if selected == nil {
		panic("input: ColorPalette needs a color to point at")
	}
	if len(swatches) == 0 {
		panic("input: ColorPalette needs at least one swatch")
	}
	for i, sw := range swatches {
		if sw.Name == "" {
			panic("input: ColorPalette swatch " + strconv.Itoa(i) + " has no Name; a color with no name is read out as nothing")
		}
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	side := u * 8

	// A run of fixed squares that wraps onto further lines, rather than a
	// grid of tracks: a swatch is chosen by its color, so it has to be the
	// same size as every other one whatever the width of the field it sits
	// in, and a swatch that stretched with its track would be measured
	// rather than looked at.
	run := ui.Row(c).Gap(u * 1.5).Wrap().AlignContent(ui.Start)
	if opts.Label != "" {
		run.Label(opts.Label)
	}
	run.Children(func() {
		for _, sw := range swatches {
			chosen := *selected == sw.Color
			cell := ui.ButtonBase(c).Size(side, side).Radius(theme.SmallRadius).
				Background(sw.Color).Label(sw.Name).Tooltip(sw.Name).
				Disabled(opts.Disabled)
			if cell.Clicked() {
				*selected = sw.Color
				chosen = true
			}
			// A hairline of the window's own edge under every swatch, so one
			// of very nearly the surface's color is still a swatch and not a
			// patch of nothing.
			cell.BorderWidth(theme.BorderWidth).BorderColor(k.Border.Mix(k.Text, 0.2))
			cell.DrawOver(func(p *ui.Painter, r ui.Rect) {
				if !chosen {
					return
				}
				rad := theme.SmallRadius
				p.Stroke(r, inkOn(k.Background), rad, 2)
				var tick ui.Path
				tick.MoveTo(r.X+r.W*0.28, r.Y+r.H*0.52).
					LineTo(r.X+r.W*0.44, r.Y+r.H*0.68).
					LineTo(r.X+r.W*0.74, r.Y+r.H*0.32)
				p.StrokePath(&tick, 2, inkOn(sw.Color))
			})
		}
	})
	return run
}

// ColorPickerOptions configure a ColorPicker.
type ColorPickerOptions struct {
	// Label names the control for assistive technology, and is required: the
	// well shows one color and no words, so without a name there is nothing
	// to read out and nothing saying what is being picked.
	Label string
	// Disabled greys the well out.
	Disabled bool
}

// ColorPicker is a well showing a color, which opens a panel to pick a new
// one and writes it into the caller's ui.Color.
//
// The panel is MyGo's own picker — the square of saturation and brightness,
// the hue and opacity rails, the hex field and its swatches — re-themed by
// core.Use along with everything else, so it is the desktop's control with
// this window's colours rather than a second one drawn to look like it.
//
// The well's own border asks the surface under it which ink reads, for the
// same reason the palette's ring does: a swatch of very nearly the
// background's own color has to be findable, and in a dark window that is
// the dark swatch that needs it.
func ColorPicker(c *ui.Context, color *ui.Color, opts ColorPickerOptions) *ui.Element {
	if color == nil {
		panic("input: ColorPicker needs a color to point at")
	}
	if opts.Label == "" {
		panic("input: ColorPicker needs options.Label; its well shows a color and no words, so there is nothing to read out")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	well := ui.ButtonBase(c).Padding(u * 0.75).Radius(theme.SmallRadius).
		Background(k.Surface).Shrink(0).Label(opts.Label).Tooltip(hexOf(*color)).
		Disabled(opts.Disabled)
	well.BorderWidth(theme.BorderWidth).BorderColor(inkOn(k.Background).Alpha(0.35))
	open := ui.Local(well, openKey{}, func() bool { return false })
	if well.Clicked() {
		*open = !*open
	}
	now := *color
	well.Children(func() {
		ui.Box(c).Size(u*7, u*5).Radius(theme.SmallRadius).Shrink(0).Role(ui.RoleNone).
			Background(now).Clip()
	})
	ui.PopoverBase(c, well, open, func(panel *ui.Element) {
		panelFace(c, panel)
		ui.ColorPicker(c, color)
	})
	return well
}

// hexOf is a color as "#rrggbb", which is what the well's tooltip says and
// what a colour is called when it is not called anything else.
func hexOf(col ui.Color) string {
	if col.A == 255 {
		return fmt.Sprintf("#%02x%02x%02x", col.R, col.G, col.B)
	}
	return fmt.Sprintf("#%02x%02x%02x%02x", col.R, col.G, col.B, col.A)
}

// TagsInputOptions configure a TagsInput.
type TagsInputOptions struct {
	// Label names the control for assistive technology, and is required: a
	// field of chips and a magnifier has no words of its own to be read out
	// by.
	Label string
	// Placeholder is what the field shows while it is empty, and is
	// required: a field for tags that does not say what a tag is leaves the
	// first chip unexplained.
	Placeholder string
	// Suggestions are offered below the field as they are typed, and taken
	// from there with the arrows or Enter.
	Suggestions []string
	// Max caps how many tags there may be; zero is no cap. It is applied by
	// dropping the tags past it, which a field that adds its own chips cannot
	// be talked out of: the tag appears for the one frame the press was read
	// in and is gone by the next. Stopping the field adding it at all is
	// TokenField's own rule to make, and the cap here is the backstop.
	Max int
	// Disabled greys the control out.
	Disabled bool
}

// TagsInput is a field that collects words: a tag is typed and added with
// Enter or a comma, taken away with its own button or with Backspace in the
// empty field, and the suggestions narrow as it is typed.
//
// The tags are the caller's []string, in the order they were added, so
// whatever a tag means downstream is decided by whoever wrote it rather than
// by a control that happens to hold them first.
//
// It is built on ui.TokenField, which is where the whole behaviour comes from
// — the chips, the removal buttons, Enter and comma, Backspace taking the
// last one, the suggestions, Escape closing them. The cap is applied after,
// because that is the only place a control can reach a slice it was handed.
func TagsInput(c *ui.Context, tags *[]string, opts TagsInputOptions) *ui.Element {
	if tags == nil {
		panic("input: TagsInput needs a slice of tags to point at")
	}
	if opts.Label == "" {
		panic("input: TagsInput needs options.Label; a field of chips has no words of its own to be read out by")
	}
	if opts.Placeholder == "" {
		panic("input: TagsInput needs options.Placeholder; a field for tags that does not say what a tag is leaves the first one unexplained")
	}
	if opts.Max < 0 {
		panic("input: TagsInput Max cannot be negative")
	}
	field := ui.TokenField(c, tags, opts.Suggestions).
		Label(opts.Label).Tooltip(opts.Placeholder).Disabled(opts.Disabled)
	if opts.Max > 0 && len(*tags) > opts.Max {
		*tags = (*tags)[:opts.Max]
	}
	return field
}
