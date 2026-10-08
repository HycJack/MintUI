package feedback

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
)

// BlinkHighlightOptions configure a BlinkHighlight.
type BlinkHighlightOptions struct {
	// Level is how strongly the highlight tints, 0 to 1, and it is the
	// caller's number.
	//
	// That is the whole design of this component: the caller owns when the
	// blink happens and how often, because only the caller knows what
	// changed. A component with a clock inside it would have to be told
	// "blink now" and then keep the time, and this package keeps nothing.
	// A caller easing Level down over a few frames gets exactly what a timed
	// blink would have given, and can stop it half way.
	Level float32
	// Color is the tint. Zero alpha takes the accent, which is the colour
	// that says "look here" without also saying how bad it is.
	Color ui.Color
	// Radius rounds the tint to match the child's own corners. Zero leaves it
	// square, which is right for a child that fills its panel and wrong for
	// one sitting inside a card: a tint that does not follow the card's
	// corners draws a rectangle of attention around a rounded object.
	Radius float32
}

// BlinkHighlight draws its child with a highlight behind it, at the strength
// the caller says.
//
// The default is off, and reduced motion keeps it off whatever the caller
// says: a highlight that fades in and out is motion whether or not anything
// moved, and the person who asked for less motion asked for less of it
// everywhere. Under reduced motion the child is drawn exactly as it would
// have been without the component, so a caller can leave the call in place
// rather than taking it out around a preference.
func BlinkHighlight(c *ui.Context, opts BlinkHighlightOptions, child func()) *ui.Element {
	ink := opts.Color
	if ink.A == 0 {
		ink = core.Tokens(c).Accent
	}
	var level float32
	if !core.Reduced(c) {
		level = clamp01(opts.Level)
	}

	e := ui.Box(c).FillWidth()
	if level > 0 {
		// The tint goes behind the child rather than over it, so the words
		// keep the contrast they were designed for instead of being washed
		// out by the thing meant to draw attention to them.
		e.Background(ink.Alpha(level * 0.2))
	}
	if opts.Radius > 0 {
		e.Radius(opts.Radius)
	}
	if child != nil {
		e.Children(child)
	}
	return e
}
