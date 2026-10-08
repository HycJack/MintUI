// Package showcase is the library's own gallery: one page per package, each
// drawing that package's components in one column.
//
// It exists so the components can be looked at. A component library with
// 368 pieces and no way to see them is a component library nobody can check,
// and screenshots of one application prove only the fifteen components that
// application happens to use.
//
// A page is not a test. Tests assert that a component draws what it promised;
// a page draws everything the package has, laid out so a person can scroll
// through it and say "that one looks wrong".
package showcase

import (
	"image"
	"sort"

	"github.com/egoist/mygo/ui"
)

// Page is one package's worth of components.
type Page struct {
	// Package is the ui/ package this page covers, e.g. "overlay".
	Package string
	// Title heads the page.
	Title string
	// Note is one line saying what the package is for, shown under the title.
	Note string
	// Width and Height are the page's own size when it is rendered alone.
	// A page that is a gallery of everything it has is tall; these are the
	// values a standalone screenshot uses.
	Width, Height int
	// Render draws the page. The harness has already called core.Use.
	Render func(c *ui.Context)
	// Anchored says the page has content pinned to the window's own edges —
	// a selection bar along the bottom, a drawer down the right. Such content
	// sits at the bottom of the page by definition, so the page cannot be
	// checked for being cut off, and its Height stays whatever its author
	// said it is.
	//
	// It is a way of saying "I know, and I have looked", not a way of getting
	// out of a warning. A page that sets it and is still cut in half is a
	// page with a bug in it.
	Anchored bool
	// Wants are strings the page must actually draw. They are what stops a
	// page from quietly going blank, and they are checked by the gate test
	// and by `gallery -check`.
	Want []string
}

// Pages is every page in the gallery, keyed by package.
var Pages = map[string]Page{}

// Register adds a page. It panics on a duplicate, which would otherwise mean
// two files quietly fighting over one package.
func Register(p Page) {
	if p.Package == "" || p.Render == nil {
		panic("showcase: a page needs a package and a Render")
	}
	if _, dup := Pages[p.Package]; dup {
		panic("showcase: two pages claim package " + p.Package)
	}
	if p.Width == 0 {
		p.Width = 1000
	}
	if p.Height == 0 {
		p.Height = 1400
	}
	Pages[p.Package] = p
}

// Sorted returns the pages in a stable order, so a gallery run and the PNGs
// it writes come out the same twice.
func Sorted() []Page {
	out := make([]Page, 0, len(Pages))
	for _, p := range Pages {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Package < out[j].Package })
	return out
}

// Gutter is the page's own margin. A section that fills the page has to stop
// short of its edge, or a card's border and a field's text are cut in half and
// the picture cannot tell a deliberate flush edge from a clipped one.
//
// It lives here rather than in each page because a page is a thing that shows
// other things: the one thing it must never do is let its own contents touch
// the window.
const Gutter = 28

// Paint draws the page inside the page's own gutter. Every way of showing a
// page — the window, the headless check, the PNGs — goes through here, so a
// page cannot be flush against its window in one of them and inset in
// another.
//
// The column is FillWidth and nothing more: Fill would fix the column's
// height to the window's, and a page whose content grew past its declared
// height would have every section compressed to a fraction of itself —
// painted overlapping whatever came after — rather than simply running past
// the bottom, which is the one failure a reader (and the bottom-row check)
// can see.
func (p Page) Paint(c *ui.Context) {
	ui.Column(c).FillWidth().Padding(0, Gutter, 0, Gutter).Children(func() {
		p.Render(c)
	})
}

// Draw lays the page out headlessly and reports the text it painted. It calls
// core.Use first, because that is the library's one hard rule and a page that
// broke it would be a page that proves nothing.
//
// What it reports is painted text, not every string the engine can point a
// reader at. Tester.Texts() is the engine's own list of the strings on the
// page, and it includes the Label an element carries for assistive
// technology, so it answers "can this be read out" where a Want entry asks
// "can this be seen". A page of empty labelled boxes is in that list in full
// and shows a blank window.
func (p Page) Draw() []string {
	tt := ui.NewTester(func(c *ui.Context) {
		useForShowcase(c)
		p.Paint(c)
	}, p.Width, p.Height)
	return paintedTexts(tt)
}

// paintedTexts drops the strings the engine reported that nothing was drawn
// for.
//
// MyGo does not say which of its strings were painted and which were only
// named: an element's text and its label are both unexported and neither has
// a getter, so the frame is the only place the difference shows. The question
// asked of each string is therefore whether its box holds a picture — glyphs
// put pixels of their own into the box they were laid out in — which is a
// discriminator and not a proof, and its two limits are worth writing down.
//
// The first is the string's other occurrences. Tester.Find returns the box of
// the first one the engine reported and nothing reaches the rest, so a string
// reported more than once is taken on the engine's word: a page that captions
// an image slot "空槽位" beneath the slot of that name is keeping its promise,
// and the slot's own empty box is the only one this can look at.
//
// The second is the box itself. A box this window shows none of — clipped to
// no width or no height, as content below a section or inside a collapsed one
// is — cannot be judged from pixels, so its strings are taken on the engine's
// word too. And a box that varies for some other reason, a border or a
// gradient or an icon or a sibling's pixels, passes whether or not the string
// is in it: the label's own box is all there is to look at.
func paintedTexts(tt *ui.Tester) []string {
	reported := tt.Texts()
	times := map[string]int{}
	for _, s := range reported {
		times[s]++
	}

	img := tt.Image()
	out := make([]string, 0, len(reported))
	for _, s := range reported {
		if times[s] > 1 {
			out = append(out, s)
			continue
		}
		if r, ok := tt.Find(s); ok {
			box := image.Rect(int(r.X), int(r.Y), int(r.X+r.W), int(r.Y+r.H)).
				Intersect(img.Bounds())
			if !box.Empty() && oneColour(img, box) {
				continue
			}
		}
		out = append(out, s)
	}
	return out
}

// oneColour reports whether a box is a single flat colour, which is what an
// element that was laid out and painted nothing leaves behind.
//
// A box holding glyphs cannot come back as one colour: the glyphs are drawn
// over whatever was already there, and the edge between them is blended
// through every colour in between. The scan stops at the first pixel that
// differs, so the whole box is only ever walked for a box that is the answer.
func oneColour(img *image.RGBA, box image.Rectangle) bool {
	flat := img.RGBAAt(box.Min.X, box.Min.Y)
	for y := box.Min.Y; y < box.Max.Y; y++ {
		for x := box.Min.X; x < box.Max.X; x++ {
			if img.RGBAAt(x, y) != flat {
				return false
			}
		}
	}
	return true
}

// Missing returns the strings in Want that the page did not paint. drawn is
// what Draw returned, and a Want entry is a promise that a person can see the
// text, so a string nothing was drawn for does not keep it.
func (p Page) Missing(drawn []string) []string {
	have := map[string]bool{}
	for _, d := range drawn {
		have[d] = true
	}
	var out []string
	for _, w := range p.Want {
		if !have[w] {
			out = append(out, w)
		}
	}
	return out
}
