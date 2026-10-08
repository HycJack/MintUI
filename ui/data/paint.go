package data

import (
	"math"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// The marks this package draws, in the same 24×24 stroke style as the rest
// of the library, so a tree's arrow and a carousel's chevron are the same
// shape as the ones a nav item carries.
func glyph(shapes string) *ui.SVG {
	return ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" ` +
		`fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" ` +
		`stroke-linejoin="round">` + shapes + `</svg>`))
}

var (
	iconBack  = glyph(`<path d="m14.5 5.5-6.5 6.5 6.5 6.5"/>`)
	iconNext  = glyph(`<path d="m9.5 5.5 6.5 6.5-6.5 6.5"/>`)
	iconPlus  = glyph(`<path d="M12 5.5v13M5.5 12h13"/>`)
	iconMinus = glyph(`<path d="M5.5 12h13"/>`)
)

// bar is the width of the mark down the side of the row the keyboard moves
// from. Every row reserves it, so choosing a row does not shift the text
// beside it.
const bar float32 = theme.BorderWidth * 2

// paintRow gives a row its surface: AccentBg while it is chosen, and the
// bar down its side while it is the row the keys move from.
//
// It is called before the row's content, because the bar is what it adds
// and a bar added after the content would sit at the row's other end.
func paintRow(c *ui.Context, r *ui.Element, chosen, lead bool) *ui.Element {
	k := core.Tokens(c)
	if chosen {
		r.Background(k.AccentBg)
	}
	return r.Children(func() {
		// The mark is there whether or not it is painted, because a row that
		// grew a mark when it was chosen would shift its own text.
		mark := ui.Box(c).Width(bar).Shrink(0).FillHeight()
		if lead {
			mark.Background(k.Accent)
		}
	})
}

// sortMark draws the arrow over the column the rows are ordered by, pointing
// the way they are ordered: up for ascending, down for descending.
func sortMark(c *ui.Context, size float32, descending bool, col ui.Color) *ui.Element {
	return ui.Box(c).Size(size, size).Shrink(0).Role(ui.RoleNone).Draw(func(p *ui.Painter, r ui.Rect) {
		var path ui.Path
		y := r.Y + r.H*0.35
		if descending {
			y = r.Y + r.H*0.65
		}
		path.MoveTo(r.X+r.W*0.1, y).LineTo(r.X+r.W*0.5, r.Y+r.H*0.5).
			LineTo(r.X+r.W*0.9, y)
		p.StrokePath(&path, 1.4, col)
	})
}

// chevron draws the mark of a disclosure: pointing right, turned by deg
// degrees clockwise, so the thing it opens turns down as it opens.
//
// It is drawn rather than taken from the glyphs because it turns: a mark
// that swapped between two shapes reads as a flicker where a mark that
// turns reads as one thing moving.
func chevron(c *ui.Context, size, deg float32, col ui.Color) *ui.Element {
	return ui.Box(c).Size(size, size).Shrink(0).Role(ui.RoleNone).Draw(func(p *ui.Painter, r ui.Rect) {
		sin, cos := math.Sincos(float64(deg) * math.Pi / 180)
		s32, c32 := float32(sin), float32(cos)
		cx, cy := r.X+r.W/2, r.Y+r.H/2
		at := func(x, y float32) (float32, float32) {
			return cx + x*c32 - y*s32, cy + x*s32 + y*c32
		}
		d := r.W / 6
		x0, y0 := at(-d, -2*d)
		x1, y1 := at(d, 0)
		x2, y2 := at(-d, 2*d)
		var path ui.Path
		path.MoveTo(x0, y0).LineTo(x1, y1).LineTo(x2, y2)
		p.StrokePath(&path, 1.4, col)
	})
}

// twisty is the mark down the side of a tree row with children, which opens
// and closes them, and reports the press.
//
// It is not drawn at all on a row without children: a mark that does nothing
// is a control that lies about what it will do.
func twisty(c *ui.Context, size float32, open bool) (mark *ui.Element, toggled bool) {
	k := core.Tokens(c)
	turn := float32(0)
	if open {
		turn = 1
	}
	// What the mark is called is library copy, so it goes through the
	// window's overrides: a tree in another language is read out in that
	// language, and "Expand" is not a word in it.
	name := core.Msg(c, "data.expand", "Expand")
	if open {
		name = core.Msg(c, "data.collapse", "Collapse")
	}
	mark = ui.Box(c).Size(size, size).Shrink(0).Role(ui.RoleNone).Label(name)
	toggled = mark.Clicked()
	// It turns as it opens, and not at all where the desktop asks for less
	// motion: a mark that swings is a mark that draws the eye.
	ms := time.Duration(core.Motion(c, 150)) * time.Millisecond
	return mark.Children(func() {
		chevron(c, size*0.62, mark.Animate("open", turn, ms)*90, k.TextMuted)
	}), toggled
}
