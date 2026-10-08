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

// Draw lays the page out headlessly and reports what it drew. It calls
// core.Use first, because that is the library's one hard rule and a page that
// broke it would be a page that proves nothing.
func (p Page) Draw() []string {
	tt := ui.NewTester(func(c *ui.Context) {
		useForShowcase(c)
		p.Paint(c)
	}, p.Width, p.Height)
	return tt.Texts()
}

// Missing returns the strings in Want that the page did not draw.
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
