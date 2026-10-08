package layout

import (
	"image"
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
)

// A vertical divider is a line one pixel wide, so the only thing that can go
// wrong with it is that it has no length. Counting painted pixels says so
// where measuring its box does not: an element with no height still measures
// whatever width it was given, so a zero-length divider passes a width test
// and shows nothing at all on screen.
func TestAVerticalDividerHasLength(t *testing.T) {
	var ink ui.Color
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ink = core.Tokens(c).Border
		ui.Column(c).Padding(40 * core.Density(c).Unit()).AlignItems(ui.Start).Children(func() {
			ui.Row(c).Height(20 * core.Density(c).Unit()).AlignItems(ui.Center).Children(func() {
				Divider(c, DividerOptions{Vertical: true})
			})
		})
	}, 400, 200)

	img := tt.Image()
	lit, widest, tallest := 0, 0, 0
	run := 0
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		run = 0
		for x := b.Min.X; x < b.Max.X; x++ {
			if !nearInk(img, x, y, ink) {
				continue
			}
			run++
			lit++
		}
		if run > widest {
			widest = run
		}
		if run > 0 {
			tallest++
		}
	}
	// 20 units tall comes out around 70 pixels; a line that short is a line
	// nobody can see.
	if lit < 40 {
		t.Fatalf("a vertical divider 20 units tall painted %d pixels; it is a line, "+
			"so a short one is a line that has been drawn and cannot be seen", lit)
	}
	if widest > 4 {
		t.Fatalf("a vertical divider painted %d pixels across in one row, so it is "+
			"taking width it was never given: a band, not a line", widest)
	}
	if tallest < 20 {
		t.Fatalf("the divider is only %d rows tall; it is standing in a row 20 units "+
			"high and should run the whole of it", tallest)
	}
}

// nearInk is loose by one 8-bit step, because a border is painted through
// whatever the platform's compositor does with it.
func nearInk(img image.Image, x, y int, ink ui.Color) bool {
	r, g, b, _ := img.At(x, y).RGBA()
	dr, dg, db := int(r>>8)-int(ink.R), int(g>>8)-int(ink.G), int(b>>8)-int(ink.B)
	return dr*dr+dg*dg+db*db < 2500
}
