package agent

import (
	"strconv"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/layout"
)

// The small shared helpers. They are here rather than in each file because
// three of them are needed by every file, and a package with six copies of
// `strconv.Itoa` behind six slightly different names is a package where a
// caller can no longer find where a number was rendered.

// itoa renders a count. The library's own internal.Commas is for the figures
// that appear in a table cell; the counts in a run are steps, tools and lines,
// none of which reaches four digits before the scroll does.
func itoa(n int) string { return strconv.Itoa(n) }

// join concatenates the parts with a separator, skipping the empty ones.
//
// Empty parts are skipped rather than rendered as gaps because most of the
// optional fields in this package are exactly that: a line of metadata with a
// figure in it, or with a word in it, and never both empty.
func join(sep string, parts ...string) string {
	out := ""
	for _, p := range parts {
		if p == "" {
			continue
		}
		if out != "" {
			out += sep
		}
		out += p
	}
	return out
}

// signed renders a diff count with the sign that side of a diff carries:
// "+12", "−3", and nothing at all for zero. A zero is left off rather than
// written "+0", which is a figure about nothing wearing the clothes of one.
func signed(n int) string {
	switch {
	case n > 0:
		return "+" + itoa(n)
	case n < 0:
		return "−" + itoa(-n)
	}
	return ""
}

// panel is layout.Container with the one thing this package's containers all
// need added to it: a width.
//
// A Box's own width is its content, and every container here is full of rows
// that fill their parent — so the container has no content to measure itself
// from, collapses to whatever the percentages happened to resolve against,
// and truncates its own headers. The library's own cards are FillWidth for
// this reason and not for any other. Anything narrower than its parent puts
// itself in a Grow(1) box, where FillWidth means "the share I was given".
func panel(c *ui.Context, opts layout.ContainerOptions, children func()) *ui.Element {
	return layout.Container(c, opts, children).FillWidth()
}

// accent is the window's highlight, read through core so that a component
// drawn in a dark window marks its selection with the dark window's accent.
// It is a function rather than a package variable because the palette is
// resolved per frame, and a component that cached one would be drawing last
// frame's colours on this frame's window.
func accent(c *ui.Context) ui.Color { return core.Tokens(c).Accent }
