package feedback

import (
	"math"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
)

// This file holds how every animated component in the package decides to
// move. Keeping the decision in one place is what makes "reduce motion"
// mean the same thing in a loader and in a list that reorders.

// twoPi is one turn of a circle, in radians. Loaders that go round or up and
// down measure their phase in turns, because a period is then always 1 and
// the arithmetic below never has to think about how long a second is.
const twoPi = float32(2 * math.Pi)

// ms turns a duration written as a plain number of milliseconds — the way
// every duration in this library is written — into a time.Duration.
func ms(v float32) time.Duration {
	return time.Duration(v * float32(time.Millisecond))
}

// wave returns 0 to 1, going up and coming back down once per turn. It is
// the one shape every loader here moves along: height, alpha and radius all
// read from it, so the five loaders move at the same speed and in step
// rather than each inventing an easing of its own.
func wave(phase float32) float32 {
	return 0.5 - 0.5*float32(math.Cos(float64(phase*twoPi)))
}

// loopPhase returns where a repeating animation is at p.Now(), as a fraction
// of one period, and asks for the frames that keep it going.
//
// still is core.Reduced, read where a Context is to hand rather than inside
// the paint, because a Painter has no Context. It decides whether there is a
// phase at all: an animation that must not move is not given a frozen frame
// of its loop, because a wave caught mid-rise reads as a stuck interface
// rather than as a quiet one. The shapes below read still beside their
// moving value and draw the pose that goes with it.
func loopPhase(p *ui.Painter, period time.Duration, still bool) float32 {
	if still {
		return 0
	}
	if period < time.Millisecond {
		period = time.Millisecond
	}
	// AnimationFrame repaints the last frame rather than rebuilding it, so a
	// loader that only moves costs its drawing and not the whole view.
	p.AnimationFrame()
	return float32(p.Now().UnixNano()%int64(period)) / float32(period)
}

// MotionSpec is an animation resolved against the window's motion
// preference. It is what the components that hand work to MyGo's own
// Transition build, and it is dead when the window asked for reduced motion:
// Duration is zero and there is nothing to run.
//
// It is a value rather than a call to Element.Transition inside each
// component because "does this window animate at all" is a question worth
// asking once and checking in a test, rather than a condition copied into
// five places.
type MotionSpec struct {
	// Duration is how long the change takes.
	Duration time.Duration
	// Enter is where an element comes from as it appears, Exit where one
	// goes. Either may be nil, which leaves that half alone.
	Enter, Exit *ui.Motion
	// Move follows the element as the layout puts it somewhere else, and
	// Colors moves its background and border. Both are how a reordering
	// list slides rather than jumping.
	Move, Colors bool
}

// Animates reports whether the spec runs at all.
//
// A zero Duration is the point: ui.ElementTransition reads a zero Duration as
// "use the default", so handing MyGo a dead spec would quietly start a 200ms
// animation for a window that asked for none. Callers check this and skip the
// Transition call entirely.
func (m MotionSpec) Animates() bool { return m.Duration > 0 }

// transition returns the MyGo transition the spec describes, for the callers
// that have already checked Animates.
func (m MotionSpec) transition() ui.ElementTransition {
	return ui.ElementTransition{
		Duration: m.Duration,
		Enter:    m.Enter,
		Exit:     m.Exit,
		Position: m.Move,
		Size:     m.Move,
		Colors:   m.Colors,
	}
}

// resolveMotion turns a duration in milliseconds into a spec, dead when the
// window asked for reduced motion. d is core.Motion's argument, so a caller
// writes a plain constant and the window's preference still decides.
func resolveMotion(c *ui.Context, d float32, enter, exit *ui.Motion) MotionSpec {
	return MotionSpec{
		Duration: ms(core.Motion(c, d)),
		Enter:    enter,
		Exit:     exit,
	}
}

// fade is the entrance a component falls back to: the element is there, and
// comes in without the movement. It is the default because a fade reads
// correctly on any shape, where a slide reads as a mistake when the shape is
// not where the slide came from.
var fade = &ui.Motion{Opacity: 0}

// rise is the entrance a list uses: it comes up a little as it fades in, so
// the list reads as arriving from below rather than blinking into being.
var rise = &ui.Motion{Opacity: 0, Y: 8}
