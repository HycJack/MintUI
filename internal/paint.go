// Package internal holds the drawing helpers components share. Nothing here
// is part of the library's surface: it exists so that components compose the
// same few marks the same way.
package internal

import "github.com/egoist/mygo/ui"

// Rule draws a hairline across the top of r.
func Rule(p *ui.Painter, r ui.Rect, col ui.Color) {
	p.Fill(ui.Rect{X: r.X, Y: r.Y, W: r.W, H: 1}, col, 0)
}

// Dot draws a filled circle of radius rad centred at cx, cy.
func Dot(p *ui.Painter, cx, cy, rad float32, col ui.Color) {
	var path ui.Path
	path.Circle(cx, cy, rad)
	p.FillPath(&path, col)
}

// Ring draws a circle's outline.
func Ring(p *ui.Painter, cx, cy, rad, width float32, col ui.Color) {
	var path ui.Path
	path.Circle(cx, cy, rad)
	p.StrokePath(&path, width, col)
}

// Meter draws a run of bars, filled to the left and empty to the right, the
// way a priority or a signal strength reads at a glance.
func Meter(p *ui.Painter, r ui.Rect, filled, total int, col, empty ui.Color) {
	if total <= 0 || r.W <= 0 {
		return
	}
	gap := r.W / float32(total*2-1)
	bar := (r.W - gap*float32(total-1)) / float32(total)
	if bar <= 0 {
		return
	}
	for i := range total {
		x := r.X + float32(i)*(bar+gap)
		c := empty
		if i < filled {
			c = col
		}
		p.Fill(ui.Rect{X: x, Y: r.Y, W: bar, H: r.H}, c, bar/2)
	}
}

// Tiles draws rounded tiles side by side across r, each dropped by step more
// than the one before it. It backs the illustrations that stand in for absent
// content, where the point is the shape rather than the detail.
func Tiles(p *ui.Painter, r ui.Rect, weights []float32, step, radius float32, col ui.Color) {
	total := float32(0)
	for _, w := range weights {
		total += w
	}
	gaps := r.W * 0.06 * float32(max(0, len(weights)-1))
	if total <= 0 || r.W <= gaps {
		return
	}
	usable := (r.W - gaps) / total
	x := r.X
	for i, w := range weights {
		width := w * usable
		y := r.Y + step*float32(len(weights)-1-i)
		p.Fill(ui.Rect{X: x, Y: y, W: width, H: r.H}, col, radius)
		x += width + r.W*0.06
	}
}
