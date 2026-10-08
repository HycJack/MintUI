package input

import (
	"image"
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// inkOf returns the colour the pad draws strokes in, so a test does not have
// to hard-code a token that the palette may change.
func inkOf(c *ui.Context) ui.Color { return core.Tokens(c).Text }

// inkBoundsIn scans a frame for the ink and returns the union of the pixels
// that lie inside keep. A zero result means there was no ink there.
func inkBoundsIn(img image.Image, ink ui.Color, keep func(ui.Rect) bool) ui.Rect {
	b := img.Bounds()
	var out ui.Rect
	seen := false
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			pix := colorAt(img, x, y)
			// A stroke is painted with round caps and a soft edge, so an exact
			// match is too strict; one 8-bit step of slack is not.
			if abs32(pix.R, ink.R) > 8 || abs32(pix.G, ink.G) > 8 || abs32(pix.B, ink.B) > 8 {
				continue
			}
			p := ui.Rect{X: float32(x), Y: float32(y), W: 1, H: 1}
			if !keep(p) {
				continue
			}
			if !seen {
				out, seen = p, true
				continue
			}
			lo, hi := out, p
			if p.X < lo.X {
				lo.X = p.X
			}
			if p.Y < lo.Y {
				lo.Y = p.Y
			}
			if p.X+p.W > hi.X+hi.W {
				hi.W = p.X + p.W - hi.X
			}
			if p.Y+p.H > hi.Y+hi.H {
				hi.H = p.Y + p.H - hi.Y
			}
			out = lo
			out.W = hi.W
			out.H = hi.H
		}
	}
	return out
}

func abs32(a, b uint8) uint8 {
	if a > b {
		return a - b
	}
	return b - a
}

// colorAt reads one pixel back as a MyGo colour so the comparison below is in
// the same 8-bit space the palette is written in.
func colorAt(img image.Image, x, y int) ui.Color {
	r, g, b, _ := img.At(x, y).RGBA()
	return ui.Color{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: 255}
}

// TestASignatureLandsInsideItsPad exists because this was wrong once and every
// existing test still passed: the strokes were painted in window coordinates
// while the points were the pad's own, so a signature — drawn live or restored
// from storage — appeared at the window's top-left corner. A test that only
// asks "was a stroke recorded" cannot see that, so this one asks where the ink
// actually is.
func TestASignatureLandsInsideItsPad(t *testing.T) {
	// The pad is pushed away from the origin so that a bug cannot hide by
	// coincidence: ink at the corner is nowhere near it.
	strokes := []Stroke{
		{Points: []SignaturePoint{{X: 20, Y: 14}, {X: 44, Y: 30}, {X: 70, Y: 16}}},
	}

	var ink ui.Color
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ink = inkOf(c)
		ui.Column(c).Padding(60).AlignItems(ui.Start).Children(func() {
			SignaturePad(c, &strokes, SignaturePadOptions{Label: "Signature"})
		})
	}, 800, 500)

	pad, ok := tt.Find("Signature")
	if !ok {
		t.Fatal("the pad must be named so a test can find it")
	}
	if pad.X < 20 || pad.Y < 20 {
		t.Fatalf("the pad is at %v; this test needs it away from the origin so a "+
			"signature painted at the corner cannot look right", pad)
	}

	inside := inkBoundsIn(tt.Image(), ink, func(p ui.Rect) bool {
		return p.X >= pad.X && p.Y >= pad.Y &&
			p.X < pad.X+pad.W && p.Y < pad.Y+pad.H
	})
	if inside == (ui.Rect{}) {
		t.Error("no ink inside the pad — the signature painted somewhere else. " +
			"SignaturePoint is the pad's own coordinate, ui.Painter is the window's")
	}
}

// TestASignatureIsNotPaintedAtTheWindowCorner is the same bug from the other
// side: assert the failure mode itself is gone, so a regression says "it moved
// to the corner" rather than "the ink moved, somewhere".
func TestASignatureIsNotPaintedAtTheWindowCorner(t *testing.T) {
	strokes := []Stroke{
		{Points: []SignaturePoint{{X: 8, Y: 8}, {X: 30, Y: 20}}},
	}
	var ink ui.Color
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ink = inkOf(c)
		ui.Column(c).Padding(60).AlignItems(ui.Start).Children(func() {
			SignaturePad(c, &strokes, SignaturePadOptions{Label: "Signature"})
		})
	}, 800, 500)

	corner := inkBoundsIn(tt.Image(), ink, func(p ui.Rect) bool {
		return p.X < 30 && p.Y < 30
	})
	if corner != (ui.Rect{}) {
		t.Errorf("ink at the window's top-left corner %v: the strokes are being "+
			"painted in window coordinates instead of the pad's", corner)
	}
	_ = theme.BodySize
}
