package data

import (
	"strconv"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/theme"
)

// Slide is one page of a Carousel: what it shows, and what it is called —
// which is what its dot is labelled with and what a screen reader says.
type Slide struct {
	Label string
	Body  func()
}

// CarouselOptions configure a Carousel.
type CarouselOptions struct {
	// Slides are the pages, in order.
	Slides []Slide
	// At is the slide on show, in the caller's state: the arrows and the
	// dots move it and nothing else does. That is the whole of a carousel's
	// state, which is why a carousel can be sent to a slide from a menu, a
	// keyboard or a test as easily as from its own controls.
	At *int
	// Height is the height of the window the slides show in. It is
	// required: a carousel with no height grows to fit its slides, and then
	// there is no window for a slide to be paged across.
	Height float32
	// Width is the width a slide is laid out in; zero takes the carousel's
	// own, measured from the frame before.
	Width float32
	// Dots draws a dot per slide under the window, for a carousel with more
	// than two slides: the arrows say there is a next and the dots say how
	// many there are.
	Dots bool
	// NoArrows draws no arrows at the sides, for a caller with its own
	// controls — a caption that pages, a row of chips.
	NoArrows bool
	// Loop pages from the last slide back to the first. Off, the arrows stop
	// at the ends, which is what says how many slides there are.
	Loop bool
}

// Carousel shows one slide at a time, with the arrows and the dots that page
// between them, sliding across as it goes.
//
// A carousel that plays itself is one whose caller plays it: keep At in your
// own state, ask for a frame with c.After, and move At when the time comes.
// A carousel that kept the time itself would move on at whatever rate the
// window happened to paint, which is a clock nobody can set.
func Carousel(c *ui.Context, opts CarouselOptions) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if opts.At == nil {
		panic("data: Carousel needs the At index to point at")
	}
	if len(opts.Slides) == 0 {
		panic("data: Carousel needs at least one slide")
	}
	if *opts.At < 0 || *opts.At >= len(opts.Slides) {
		panic("data: Carousel is showing slide " + strconv.Itoa(*opts.At) +
			", which is not one of its slides")
	}
	if opts.Height <= 0 {
		panic("data: Carousel needs a Height; without one it cannot page")
	}

	at := *opts.At
	n := len(opts.Slides)
	step := func(d int) {
		to := at + d
		switch {
		case to < 0:
			if opts.Loop {
				to = n - 1
			} else {
				return
			}
		case to > n-1:
			if opts.Loop {
				to = 0
			} else {
				return
			}
		}
		*opts.At = to
	}

	view := ui.Column(c).FillWidth().Shrink(0).Radius(theme.CardRadius)
	view.Children(func() {
		window := ui.Box(c).FillWidth().Height(opts.Height).Shrink(0).
			Background(k.Surface).Radius(theme.CardRadius).Clip()

		// The width a slide is laid out in is said outright, because a scroll
		// window hands its content what fits rather than what it asked for.
		// The window's own width is the one it had in the frame before, and
		// there is none on the frame the carousel appears on — that frame
		// lays the slides out at the width of what is in them and asks for
		// the one that measures the window.
		w := opts.Width
		if measured := window.Bounds().W; measured > 0 {
			w = measured
		} else {
			c.Invalidate()
		}
		// The slides move across rather than being swapped. A slide that
		// disappears while another appears reads as two things; a slide that
		// moves reads as one carousel moving. Where the desktop asks for
		// less motion the movement is not animated at all, which leaves the
		// same slide in the middle of the same window.
		off := window.Animate("slide", float32(at),
			time.Duration(core.Motion(c, 220))*time.Millisecond) * w
		window.Children(func() {
			// The slides are laid over the window rather than in it, so that
			// the window's own width does not come from them: a window sized
			// by its slides, whose slides are sized by the window, would be
			// two numbers chasing each other, and a carousel beside anything
			// else would grow a little every frame.
			row := ui.Row(c).Absolute().Top(0).Left(0).AlignItems(ui.Stretch).Shrink(0)
			row.Children(func() {
				for _, s := range opts.Slides {
					slide := ui.Box(c).FillHeight().Shrink(0).Role(ui.RoleNone)
					if w > 0 {
						slide.Width(w)
					}
					slide.Children(func() {
						if s.Body != nil {
							s.Body()
						}
					})
				}
			})
			row.Left(-off)
		})

		if opts.NoArrows && !opts.Dots {
			return
		}
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u*1.5).
			Padding(u*2, u*3).Children(func() {
			if !opts.NoArrows {
				if input.IconButton(c, iconBack, "Previous slide", input.ButtonOptions{
					Disabled: at == 0 && !opts.Loop,
				}).Clicked() {
					step(-1)
				}
				ui.Box(c).Grow(1)
			}
			if opts.Dots {
				for i, s := range opts.Slides {
					dot := ui.Box(c).Height(u * 2).Shrink(0).Radius(theme.PillRadius).
						Role(ui.RoleNone).Label(dotLabel(c, s, i, n))
					if i == at {
						// The slide on show is the long one: a row of equal
						// dots says "there is one of these" without saying
						// which, which is the one thing the row is for.
						dot.Width(u * 5).Background(k.Text)
					} else {
						dot.Width(u * 2).Background(k.Border)
					}
					if dot.Clicked() {
						*opts.At = i
					}
				}
			}
			if !opts.NoArrows {
				ui.Box(c).Grow(1)
				if input.IconButton(c, iconNext, "Next slide", input.ButtonOptions{
					Disabled: at == n-1 && !opts.Loop,
				}).Clicked() {
					step(1)
				}
			}
		})
	})
	return view
}

// dotLabel names a slide's dot: what it shows and how far along the carousel
// it is, so that a dot read out says "Hillside Dental, 2 of 5" rather than
// "button". The words are the library's, so they go through the window's
// overrides like every other word here.
func dotLabel(c *ui.Context, s Slide, i, n int) string {
	at := strconv.Itoa(i+1) + " of " + strconv.Itoa(n)
	if s.Label == "" {
		return core.Msg(c, "data.slideAt", "Slide ") + at
	}
	return s.Label + ", " + at
}
