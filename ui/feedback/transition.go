package feedback

import "github.com/egoist/mygo/ui"

// The two components here hand work to MyGo's own Element.Transition, which
// is what makes a row slide to its new place in a board column instead of
// jumping. Neither of them holds a clock or a start time: the transition
// keeps those on the element, keyed by the element's own identity, which is
// the one piece of state this package can use without owning it.

const (
	// defaultTransitionDuration is how long a change takes when the caller
	// says nothing, in milliseconds: long enough to follow, short enough not
	// to wait for.
	defaultTransitionDuration float32 = 220
	// defaultStaggerStep is how much longer each item of a staggered list
	// takes than the one before it, in milliseconds.
	defaultStaggerStep float32 = 40
)

// LayoutTransitionOptions configure a LayoutTransition.
type LayoutTransitionOptions struct {
	// Key identifies the element from frame to frame, and is required.
	//
	// MyGo will fall back on an element's place among its siblings, which
	// is fine until the list reorders — which is the only reason to want a
	// transition in the first place. So this component insists: a
	// transition without a key silently animates the wrong element once the
	// thing it was watching moves, which is worse than not animating at all.
	Key any
	// Duration is how long the change takes, in milliseconds. Zero takes
	// 220, which is long enough to follow and short enough not to wait for.
	Duration float32
	// Enter is where the element comes from as it appears; Exit is where it
	// goes. Both may be nil, which leaves that half alone. Enter and Exit
	// default to a fade when only one of the two is asked for, because an
	// element that appears from nowhere is the commonest way to get a
	// transition wrong.
	Enter, Exit *ui.Motion
	// Move follows the element as the layout puts it somewhere else — the
	// reordering case. It is off by default because a transition that moves
	// and resizes everything makes a window that changes by one row feel
	// like it changed wholesale.
	Move bool
	// Colors moves the element's background and border as they change, for a
	// row that changes severity as well as place.
	Colors bool
}

// LayoutTransition draws its child and animates it from wherever the last
// frame showed it to wherever this frame's layout puts it.
//
// Under reduced motion the child is drawn as a plain box with no transition
// at all. Zeroing the Duration would not do: ui.ElementTransition reads a
// zero Duration as "use the default", so a dead spec handed to MyGo starts a
// 200ms animation in exactly the windows that asked for none.
func LayoutTransition(c *ui.Context, opts LayoutTransitionOptions, child func()) *ui.Element {
	if opts.Key == nil {
		panic("feedback: LayoutTransition needs a Key; without one the element " +
			"has no identity from frame to frame and the transition follows " +
			"the wrong thing as soon as the list reorders")
	}
	spec := transitionSpec(c, opts.Duration, opts.Enter, opts.Exit, opts.Move, opts.Colors)

	e := ui.Box(c).FillWidth().Key(opts.Key)
	if child != nil {
		e.Children(child)
	}
	if spec.Animates() {
		e.Transition(spec.transition())
	}
	return e
}

// transitionSpec resolves a set of transition options against the window's
// motion preference, filling in the half of the entrance or exit the caller
// left out.
func transitionSpec(c *ui.Context, dur float32, enter, exit *ui.Motion, move, colors bool) MotionSpec {
	if dur <= 0 {
		dur = defaultTransitionDuration
	}
	// Asking for one half and getting both is friendlier than asking for an
	// Enter, getting a fade in and a jump out, and wondering which was
	// which.
	if enter != nil && exit == nil {
		exit = enter
	}
	if exit != nil && enter == nil {
		enter = exit
	}
	if enter == nil && exit == nil && move {
		// A move with no entrance still needs one: a row that has just been
		// added would otherwise be in its place from the first frame while
		// its neighbours slide past it, which reads as a row that was there
		// all along and is now not.
		enter, exit = fade, fade
	}
	return resolveMotion(c, dur, enter, exit).with(move, colors)
}

// with returns the spec with its layout-following properties set, so the
// caller does not repeat the pattern of building a spec and then filling in
// the half of it that has nothing to do with motion.
func (m MotionSpec) with(move, colors bool) MotionSpec {
	m.Move, m.Colors = move, colors
	return m
}

// StaggerOptions configure a Stagger.
type StaggerOptions struct {
	// Step is how much longer each item after the first takes, in
	// milliseconds. Zero takes 40, about the gap between two items
	// arriving one after another rather than together.
	Step float32
	// Duration is how long the first item's entrance takes. Zero takes the
	// standard transition duration.
	Duration float32
	// Enter is where each item comes from. Nil takes the rise, which reads
	// as a list arriving from below rather than blinking into being.
	Enter *ui.Motion
	// Exit is where each item goes; nil takes the same as Enter.
	Exit *ui.Motion
	// Move follows each item as the layout puts it somewhere else, so a
	// reordered list slides as well as arriving.
	Move bool
}

// Stagger hands out each child's entrance in turn, so a list arrives as a
// list rather than as one block.
//
// It is a builder rather than a single call because the caller owns the list:
// it has the items, it knows how to draw each one, and only it knows which
// child is which. A component that took a slice would have to be told how to
// draw one item to be worth anything.
//
//	s := feedback.NewStagger(c, feedback.StaggerOptions{})
//	ui.Column(c).Children(func() {
//	    for _, item := range items {
//	        s.Wrap(ui.Row(c).Label(item.Title).Children(...))
//	    }
//	})
//
// The items are identified by their place among their siblings, which is
// enough here: a staggered list is one that is being built, and a list that
// is being reordered wants LayoutTransition and real keys.
type Stagger struct {
	c    *ui.Context
	opts StaggerOptions
	n    int
}

// NewStagger returns a Stagger for one list. Give every item of that list to
// the same Stagger: the item number it counts is what puts the delay between
// them, so a second Stagger starts the count over.
func NewStagger(c *ui.Context, opts StaggerOptions) *Stagger {
	return &Stagger{c: c, opts: opts}
}

// Wrap returns e with the next item's entrance attached, and hands back e
// itself so it can be used inline.
func (s *Stagger) Wrap(e *ui.Element) *ui.Element {
	// Read the spec before advancing the count, so the first item is item
	// zero and takes the base duration rather than one step past it.
	spec := s.spec(s.n)
	s.n++
	if !spec.Animates() {
		return e
	}
	return e.Transition(spec.transition())
}

// spec is what item i is given: the base entrance, stretched by i steps so
// the items come in one after another.
func (s *Stagger) spec(i int) MotionSpec {
	step := s.opts.Step
	if step <= 0 {
		step = defaultStaggerStep
	}
	dur := s.opts.Duration
	if dur <= 0 {
		dur = defaultTransitionDuration
	}
	// The window's motion preference comes out of the duration at the end,
	// so a window that asked for no motion gets no delay either: a list
	// that appears one item at a time is the same motion, only slower, and
	// slowing it down is not honouring the request.
	dur += step * float32(i)

	enter := s.opts.Enter
	if enter == nil {
		enter = rise
	}
	exit := s.opts.Exit
	if exit == nil {
		exit = enter
	}
	spec := resolveMotion(s.c, dur, enter, exit)
	return spec.with(s.opts.Move, false)
}
