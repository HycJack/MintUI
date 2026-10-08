package chart

import (
	"math"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
)

// goldenAngle is how far round the hue circle each colour of a palette turns:
// 137.5°, the angle that divides a pentagon's corners. Four of them fill the
// circle without landing near each other, so two series next to each other in
// a legend are never two shades of the same hue.
const goldenAngle = 137.508

// The lightness and chroma a palette walks through. Both are mid-scale on
// purpose: a colour light enough to vanish on a dark window and dark enough
// to vanish on a light one is useless to a chart, which has to be readable in
// whichever appearance the desktop is in. Each of the three lightnesses is
// clear of both, and walking through them as well as round the hue means two
// neighbouring colours differ in two directions rather than one.
var (
	paletteLightness = [3]float32{0.62, 0.72, 0.54}
	paletteChroma    = [2]float32{0.13, 0.10}
)

// Palette returns n colours for a chart's series, stepped out from the
// window's accent.
//
// It is the same ladder in both appearances and does not vary by mode: every
// colour on it clears a white window and a near-black one, so a chart does
// not need a second set of colours for the dark theme, and the two do not
// have to be kept in step when one of them is wrong.
//
// n of zero or less asks for no colours and gets none, which is what a chart
// with no series says about itself.
func Palette(c *ui.Context, n int) []ui.Color {
	return PaletteFrom(core.Tokens(c).Accent, n)
}

// PaletteFrom is [Palette] from a colour of the caller's choosing rather than
// the window's accent: a chart that means to be green starts its ladder at
// the green, and the rest of the ladder follows it.
func PaletteFrom(base ui.Color, n int) []ui.Color {
	if n <= 0 {
		return nil
	}
	hue := float64(hueOf(base))
	out := make([]ui.Color, n)
	for i := range n {
		h := math.Mod(hue+float64(i)*goldenAngle, 360)
		if h < 0 {
			h += 360
		}
		out[i] = ui.Oklch(paletteLightness[i%len(paletteLightness)],
			paletteChroma[i%len(paletteChroma)], float32(h))
	}
	return out
}

// hueOf is a colour's place on the hue circle in degrees, from its sRGB
// components. A colour with no chroma — a grey — has no hue, and starts the
// ladder at the top of the circle rather than dividing by its own zero.
func hueOf(c ui.Color) float32 {
	r := float64(c.R) / 255
	g := float64(c.G) / 255
	b := float64(c.B) / 255
	hi, lo := max3(r, g, b), min3(r, g, b)
	d := hi - lo
	if d == 0 {
		return 0
	}
	var h float64
	switch hi {
	case r:
		h = math.Mod((g-b)/d, 6)
	case g:
		h = (b-r)/d + 2
	default:
		h = (r-g)/d + 4
	}
	h *= 60
	if h < 0 {
		h += 360
	}
	return float32(h)
}

func max3(a, b, c float64) float64 { return math.Max(a, math.Max(b, c)) }
func min3(a, b, c float64) float64 { return math.Min(a, math.Min(b, c)) }
