package chart

import (
	"image"
	"math"
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// The data every render test draws: a week of callbacks, fixed so a test can
// say exactly where its points belong.
var (
	testDays   = []float64{0, 1, 2, 3, 4, 5, 6}
	testOpen   = []float64{4, 7, 5, 9, 8, 12, 11}
	testClosed = []float64{2, 4, 3, 6, 5, 7, 6}
)

func testSeries() []Series {
	return []Series{
		SeriesFrom("Open", FormLine, testDays, testOpen),
		SeriesFrom("Closed", FormArea, testDays, testClosed),
	}
}

// near reports whether two colours are within a step or two of each other. It
// is for the pixels of a translucent colour, where the scene does the
// compositing and the last unit of a channel is arithmetic; an opaque colour
// is compared exactly.
func near(got, want ui.Color) bool {
	d := func(a, b uint8) int {
		if a > b {
			return int(a) - int(b)
		}
		return int(b) - int(a)
	}
	return d(got.R, want.R) <= 2 && d(got.G, want.G) <= 2 && d(got.B, want.B) <= 2
}

// pixel is the colour at a point of a rendered frame.
func pixel(img *image.RGBA, x, y float32) ui.Color {
	b := img.Bounds()
	px, py := int(x), int(y)
	if px < b.Min.X || py < b.Min.Y || px >= b.Max.X || py >= b.Max.Y {
		return ui.Color{}
	}
	c := img.RGBAAt(px, py)
	return ui.RGB(c.R, c.G, c.B)
}

// finds reports whether any pixel in r is that colour — the way a drawn mark
// is looked for when its edge is antialiased and one pixel of its middle is
// not.
func finds(img *image.RGBA, r ui.Rect, want ui.Color) bool {
	for y := int(r.Y); y < int(r.Y+r.H); y++ {
		for x := int(r.X); x < int(r.X+r.W); x++ {
			got := pixel(img, float32(x), float32(y))
			if want.A == 255 && got == want {
				return true
			}
			if want.A != 255 && near(got, want) {
				return true
			}
		}
	}
	return false
}

// ── the maths ─────────────────────────────────────────────────────────────

// Nice is what makes a chart's ends look chosen: 0 to 97 becomes 0 to 100, and
// its ticks are round numbers rather than 19.4 of them.
func TestNiceRoundsTheEndsToPeopleNumbers(t *testing.T) {
	for _, tc := range []struct {
		lo, hi float64
		count  int
		want   Domain
	}{
		{0, 100, 5, Domain{0, 100}},
		{0, 97, 5, Domain{0, 100}},
		{3, 97, 5, Domain{0, 100}},
		{0, 1, 5, Domain{0, 1}},
		{12, 88, 4, Domain{0, 100}},
		{-7, 7, 7, Domain{-8, 8}},
		{0, 2300, 4, Domain{0, 3000}},
	} {
		if got := Nice(tc.lo, tc.hi, tc.count); got != tc.want {
			t.Errorf("Nice(%v, %v, %d) = %+v, want %+v", tc.lo, tc.hi, tc.count, got, tc.want)
		}
	}
}

// A range of nothing still has to have an axis to sit on.
func TestNiceWidensAFlatRange(t *testing.T) {
	if got, want := Nice(5, 5, 5), (Domain{4, 6}); got != want {
		t.Errorf("Nice(5, 5, 5) = %+v, want %+v", got, want)
	}
}

func TestNiceRejectsAnUpsideDownRange(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("Nice(10, 0) should panic rather than draw a chart backwards")
		}
	}()
	Nice(10, 0, 5)
}

// Linear divides what it is given and rounds nothing: a range that is already
// round stays exactly as it is.
func TestLinearDividesExactlyWhatItIsGiven(t *testing.T) {
	if got, want := Linear(Domain{0, 100}, 2), []float64{0, 50, 100}; !equalFloats(got, want) {
		t.Errorf("Linear(0..100, 2) = %v, want %v", got, want)
	}
	if got, want := Linear(Domain{0, 97}, 4), []float64{0, 24.25, 48.5, 72.75, 97}; !equalFloats(got, want) {
		t.Errorf("Linear(0..97, 4) = %v, want %v", got, want)
	}
	// A count of nothing is one step, not none: an axis always has a start
	// and an end.
	if got, want := Linear(Domain{0, 10}, 0), []float64{0, 10}; !equalFloats(got, want) {
		t.Errorf("Linear(0..10, 0) = %v, want %v", got, want)
	}
	if got, want := Linear(Domain{4, 4}, 3), []float64{4}; !equalFloats(got, want) {
		t.Errorf("Linear(4..4, 3) = %v, want %v", got, want)
	}
}

func equalFloats(got, want []float64) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

// A log axis needs something between its decades to hang a grid line on.
func TestLogMarksEveryDecadeAndTwoOfThem(t *testing.T) {
	if got, want := Log(1, 1000, 10), []float64{1, 2, 5, 10, 20, 50, 100, 200, 500, 1000}; !equalFloats(got, want) {
		t.Errorf("Log(1..1000) = %v, want %v", got, want)
	}
	if got, want := Log(3, 100, 10), []float64{5, 10, 20, 50, 100}; !equalFloats(got, want) {
		t.Errorf("Log(3..100) = %v, want %v", got, want)
	}
	if got, want := Log(1, 8, 2), []float64{1, 2, 4, 8}; !equalFloats(got, want) {
		t.Errorf("Log(1..8, base 2) = %v, want %v", got, want)
	}
}

func TestLogRefusesZero(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("a log axis over zero should panic: zero has no logarithm")
		}
	}()
	Log(0, 100, 10)
}

// A categorical axis shares itself out evenly, with each category's tick in
// the middle of its share rather than on its edge.
func TestBandSharesTheAxisEvenly(t *testing.T) {
	bands := Band([]string{"New", "Root cause", "Fixed"}, 0, 300)
	if len(bands) != 3 {
		t.Fatalf("got %d bands, want 3", len(bands))
	}
	for i, want := range []struct{ start, mid, end float32 }{
		{0, 50, 100}, {100, 150, 200}, {200, 250, 300},
	} {
		b := bands[i]
		if b.Start != want.start || b.Mid != want.mid || b.End != want.end {
			t.Errorf("band %d = %v..%v (mid %v), want %v..%v (mid %v)",
				i, b.Start, b.End, b.Mid, want.start, want.end, want.mid)
		}
		if b.Index != i || b.Label != []string{"New", "Root cause", "Fixed"}[i] {
			t.Errorf("band %d is %q at %d", i, b.Label, b.Index)
		}
	}
}

func TestBandNeedsCategories(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("a band scale with no categories should panic")
		}
	}()
	NewBand(nil, 0, 100)
}

// A tick is where the number says it is, to the pixel.
func TestTicksFallWhereTheyShould(t *testing.T) {
	s := NewLinear(Domain{0, 100}, 0, 200)
	ticks := TicksFor(s, 5, nil)
	if len(ticks) != 6 {
		t.Fatalf("five steps over 0..100 is six ticks, got %d", len(ticks))
	}
	for i, want := range []struct {
		value float64
		pos   float32
		label string
	}{
		{0, 0, "0"}, {20, 40, "20"}, {40, 80, "40"},
		{60, 120, "60"}, {80, 160, "80"}, {100, 200, "100"},
	} {
		got := ticks[i]
		// The value and the label are exact; the position is a float32 a
		// thousandth of a pixel out, which is the whole of what a pixel is.
		if got.Value != want.value || got.Label != want.label ||
			math.Abs(float64(got.Pos-want.pos)) > 0.001 {
			t.Errorf("tick %d = {%v %v %q}, want {%v %v %q}",
				i, got.Value, got.Pos, got.Label, want.value, want.pos, want.label)
		}
		if !got.Drawn {
			t.Errorf("tick %d should be drawn before anything is measured", i)
		}
	}

	// The same scale placed across a plot that starts 10 DIPs in and is 400
	// wide: 0..100 over that is 10..410.
	placed := s.Span(10, 410)
	if got, want := placed.At(0), float32(10); got != want {
		t.Errorf("At(0) over [10, 410] = %v, want %v", got, want)
	}
	if got, want := placed.At(50), float32(210); got != want {
		t.Errorf("At(50) over [10, 410] = %v, want %v", got, want)
	}
	if got, want := placed.At(100), float32(410); got != want {
		t.Errorf("At(100) over [10, 410] = %v, want %v", got, want)
	}
}

// A y axis reads upwards: its first value is at the bottom of the plot.
func TestTicksOnAYAxisRunTheOtherWayUp(t *testing.T) {
	plot := ui.Rect{X: 20, Y: 40, W: 300, H: 200}
	y := NewLinear(Domain{0, 10}, 0, 1).Span(plot.Y+plot.H, plot.Y)
	if got, want := y.At(0), float32(240); got != want {
		t.Errorf("the bottom of the range is at %v, want the plot's bottom %v", got, want)
	}
	if got, want := y.At(10), float32(40); got != want {
		t.Errorf("the top of the range is at %v, want the plot's top %v", got, want)
	}
	if got, want := y.At(2.5), float32(190); got != want {
		t.Errorf("a quarter of the way up is at %v, want %v", got, want)
	}
}

// A log scale's decades are exactly evenly spaced; the twos and fives sit
// between them, which a division cannot say exactly, so those two are checked
// to a fraction of a pixel.
func TestTicksOnALogAxis(t *testing.T) {
	ticks := TicksFor(NewLog(Domain{1, 1000}, 10, 0, 300), 5, nil)
	for _, want := range []struct {
		value float64
		pos   float32
	}{
		{1, 0}, {10, 100}, {100, 200}, {1000, 300},
	} {
		found := false
		for _, got := range ticks {
			if got.Value == want.value {
				found = true
				if math.Abs(float64(got.Pos-want.pos)) > 0.01 {
					t.Errorf("%v falls at %v, want %v", want.value, got.Pos, want.pos)
				}
			}
		}
		if !found {
			t.Errorf("no tick for %v in %v", want.value, ticks)
		}
	}
}

func TestTicksOfABandAxisAreItsCategories(t *testing.T) {
	ticks := TicksFor(NewBand([]string{"Mon", "Tue", "Wed"}, 0, 300), 5, nil)
	want := []struct {
		value float64
		pos   float32
		label string
	}{{0, 50, "Mon"}, {1, 150, "Tue"}, {2, 250, "Wed"}}
	if len(ticks) != len(want) {
		t.Fatalf("a band axis has one tick per category, got %d", len(ticks))
	}
	for i, w := range want {
		if ticks[i].Value != w.value || ticks[i].Pos != w.pos || ticks[i].Label != w.label {
			t.Errorf("tick %d = %+v, want %+v", i, ticks[i], w)
		}
	}
}

// The pointer and the value are two ways of saying the same thing, and a hover
// that cannot go back to a number is a hover nobody can act on.
func TestScaleReadsBackTheOtherWay(t *testing.T) {
	xs := NewLinear(Domain{0, 100}, 0, 250).Span(20, 270)
	for _, v := range []float64{0, 13, 50, 99.5, 100} {
		back := xs.Value(xs.At(v))
		if math.Abs(back-v) > 0.001 {
			t.Errorf("Value(At(%v)) = %v", v, back)
		}
	}
	ys := NewLog(Domain{1, 1000}, 10, 0, 1).Span(200, 0)
	for _, v := range []float64{1, 7, 100, 1000} {
		back := ys.Value(ys.At(v))
		if math.Abs(back-v)/v > 0.001 {
			t.Errorf("a log scale lost %v on the way back: %v", v, back)
		}
	}
	band := NewBand([]string{"a", "b", "c"}, 0, 300)
	if got := band.Index(10); got != 0 {
		t.Errorf("10 DIPs along a three-band axis is band %d, want 0", got)
	}
	if got := band.Index(250); got != 2 {
		t.Errorf("250 DIPs along a three-band axis is band %d, want 2", got)
	}
	if got := band.Index(-5); got != -1 {
		t.Errorf("a position off the axis is band %d, want -1", got)
	}
}

// Labels are dropped from the middle, never from the ends: the two a reader
// looks for are the ends of the range.
func TestLabelSkipDropsTheMiddleNotTheEnds(t *testing.T) {
	drawn := LabelSkip([]float32{0, 10, 20, 30, 40}, []float32{20, 20, 20, 20, 20}, 4)
	if got, want := drawn, []bool{true, false, false, false, true}; !equalBools(got, want) {
		t.Errorf("labels 20 wide every 10 DIPs = %v, want %v", got, want)
	}
	// Room for all of them: nothing is dropped.
	all := LabelSkip([]float32{0, 40, 80}, []float32{20, 20, 20}, 4)
	if !equalBools(all, []bool{true, true, true}) {
		t.Errorf("labels with room between them = %v, want all drawn", all)
	}
	// Nothing overlaps, so nothing is dropped, however wide the labels are.
	wide := LabelSkip([]float32{0, 60, 120}, []float32{50, 50, 50}, 4)
	if !equalBools(wide, []bool{true, true, true}) {
		t.Errorf("three wide labels with room = %v, want all drawn", wide)
	}
	// The first label is never traded away for the last: an axis that starts
	// unlabelled looks as broken as one that ends so.
	first := LabelSkip([]float32{0, 10, 20}, []float32{30, 30, 30}, 4)
	if !first[0] {
		t.Errorf("the first label was dropped to make room for the last: %v", first)
	}
}

// A y axis' ticks come back from the bottom of the plot, and thinning them
// left to right would keep only the first.
func TestLabelSkipReadsTheRunTheWayItRuns(t *testing.T) {
	// A vertical axis with room for all of its labels: every one is drawn.
	up := LabelSkip([]float32{179, 134, 89, 44, 0}, []float32{8, 20, 8, 20, 16}, 4)
	if !equalBools(up, []bool{true, true, true, true, true}) {
		t.Errorf("a y axis with room for its labels thinned them to %v", up)
	}
	// A crowded one drops its middle and keeps both ends.
	crowded := LabelSkip([]float32{180, 140, 100, 60, 20, 0}, []float32{60, 60, 60, 60, 60, 60}, 4)
	if !crowded[0] || !crowded[len(crowded)-1] {
		t.Errorf("a crowded y axis lost an end: %v", crowded)
	}
	if equalBools(crowded, []bool{true, true, true, true, true, true}) {
		t.Errorf("a crowded y axis drew all six labels: %v", crowded)
	}
}

func equalBools(got, want []bool) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func TestLabelSkipNeedsOneWidthEach(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("a width per position is what the skipping reads; a mismatch should panic")
		}
	}()
	LabelSkip([]float32{0, 1}, []float32{10}, 4)
}

// A scale left out of an axis is not a scale over zero: it is no axis at all.
func TestAnAxisWithNoScaleIsNoAxis(t *testing.T) {
	if (Scale{}).OK() {
		t.Error("the zero scale should not claim to be one")
	}
	if !NewLinear(Domain{0, 1}, 0, 1).OK() {
		t.Error("a built scale should say so")
	}
	if got := TicksFor(Scale{}, 5, nil); len(got) != 0 {
		t.Errorf("an axis with no scale has no ticks, got %d", len(got))
	}
}

// ── the palette ───────────────────────────────────────────────────────────

// A palette has to be n different colours, and every one of them has to read
// on a white window and on a near-black one — which is why it is stepped in
// lightness as well as in hue rather than in hue alone.
func TestPaletteIsDistinctAndClearsBothAppearances(t *testing.T) {
	for _, tc := range []struct {
		name string
		acc  ui.Color
	}{
		{"light", theme.Light().Accent},
		{"dark", theme.Dark().Accent},
	} {
		palette := PaletteFrom(tc.acc, 5)
		if len(palette) != 5 {
			t.Fatalf("%s: PaletteFrom(_, 5) gave %d colours", tc.name, len(palette))
		}
		seen := map[ui.Color]bool{}
		for i, c := range palette {
			if seen[c] {
				t.Errorf("%s: colour %d repeats an earlier one", tc.name, i)
			}
			seen[c] = true
			if c.A != 255 {
				t.Errorf("%s: colour %d is not opaque", tc.name, i)
			}
			// Against both appearances' backgrounds, because a chart draws in
			// whichever one the desktop is in.
			for _, bg := range []struct {
				name string
				col  ui.Color
			}{{"a light window", theme.Light().Background}, {"a dark window", theme.Dark().Background}} {
				if ratio := contrast(c, bg.col); ratio < 2.5 {
					t.Errorf("%s: colour %d is only %.2f:1 on %s, want at least 2.5:1",
						tc.name, i, ratio, bg.name)
				}
			}
			if s := saturation(c); s < 0.3 {
				t.Errorf("%s: colour %d is only %.0f%% saturated, want at least 30%%", tc.name, i, s*100)
			}
		}
		// Neighbours are far apart in hue, so a legend reads without reading
		// the numbers off it.
		for i := 1; i < len(palette); i++ {
			if gap := math.Abs(float64(hueOf(palette[i]) - hueOf(palette[i-1]))); gap < 90 && gap > 270 {
				t.Errorf("%s: colours %d and %d are %.0f° apart", tc.name, i-1, i, gap)
			}
		}
	}
}

// The palette is the accent's own hue walked by the golden angle, with the
// lightness and chroma ladders running alongside: colour i is exactly the one
// the formula says, which is what pins the ladder rather than a screenshot of
// it.
func TestPaletteIsTheFormulaItSaysItIs(t *testing.T) {
	for _, acc := range []ui.Color{theme.Light().Accent, theme.Dark().Accent, ui.Hex("#1f7a3d")} {
		palette := PaletteFrom(acc, 6)
		hue := float64(hueOf(acc))
		for i, got := range palette {
			want := ui.Oklch(paletteLightness[i%len(paletteLightness)],
				paletteChroma[i%len(paletteChroma)],
				float32(math.Mod(hue+float64(i)*goldenAngle, 360)))
			if got.SRGB() != want.SRGB() {
				t.Errorf("colour %d is %v, want %v", i, got.SRGB(), want.SRGB())
			}
		}
	}
	// The first colour off the light accent is this one, to the digit: it is
	// what the whole ladder hangs from.
	if got, want := PaletteFrom(theme.Light().Accent, 1)[0].SRGB(), ui.RGB(0, 150, 189); got != want {
		t.Errorf("the ladder's first colour is %v, want %v", got, want)
	}
}

// A chart that means to be green starts its ladder at the green.
func TestPaletteFollowsItsBase(t *testing.T) {
	blue := PaletteFrom(theme.Light().Accent, 1)
	green := PaletteFrom(ui.Hex("#1f7a3d"), 1)
	if math.Abs(float64(hueOf(blue[0])-hueOf(green[0]))) < 60 {
		t.Errorf("a green base gave a blue palette: %.0f° apart", hueOf(blue[0])-hueOf(green[0]))
	}
	if got := PaletteFrom(theme.Light().Accent, 0); got != nil {
		t.Errorf("a palette of nothing is nothing, got %v", got)
	}
	if got := PaletteFrom(theme.Light().Accent, -3); got != nil {
		t.Errorf("a palette of fewer than nothing is nothing, got %v", got)
	}
}

// The window's palette is its accent's ladder, in both appearances.
func TestPaletteFollowsTheWindow(t *testing.T) {
	for _, mode := range []core.Mode{core.Light, core.Dark} {
		var got []ui.Color
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{Mode: mode})
			got = Palette(c, 4)
		}, 80, 60)
		k := theme.Light()
		if mode == core.Dark {
			k = theme.Dark()
		}
		if expected := PaletteFrom(k.Accent, 4); !equalColors(got, expected) {
			t.Errorf("%v: the window's palette is not its accent's ladder: %v vs %v", mode, got, expected)
		}
	}
}

// equalSRGB compares what a window that is not a wide gamut shows, which is
// all a test can see: the sRGB half of a colour is the same wherever it was
// built.
func equalSRGB(got, want []ui.Color) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i].SRGB() != want[i].SRGB() {
			return false
		}
	}
	return true
}

func equalColors(got, want []ui.Color) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

// ── numbers as labels ─────────────────────────────────────────────────────

func TestNumbersAreWrittenTheWayTheInterfaceWritesThem(t *testing.T) {
	for _, tc := range []struct {
		in       float64
		decimals int
		want     string
	}{
		{12400.5, 1, "12,400.5"},
		{1234567, 0, "1,234,567"},
		{-1234, 0, "-1,234"},
		{-0.004, 2, "0.00"},
		{0, 0, "0"},
		{99, 0, "99"},
		{100, 0, "100"},
		{0.156, 2, "0.16"},
	} {
		if got := Group(tc.in, tc.decimals); got != tc.want {
			t.Errorf("Group(%v, %d) = %q, want %q", tc.in, tc.decimals, got, tc.want)
		}
	}
	if got, want := Auto()(12400, 100), "12,400"; got != want {
		t.Errorf("Auto()(12400, 100) = %q, want %q", got, want)
	}
	if got, want := Auto()(0.5, 0.5), "0.5"; got != want {
		t.Errorf("Auto()(0.5, 0.5) = %q, want %q", got, want)
	}
	// An axis stepped by 2.5 reads like a person wrote it, and not every
	// label on it grows a decimal it does not need.
	for _, tc := range []struct {
		v, step float64
		want    string
	}{
		{0, 2.5, "0"}, {2.5, 2.5, "2.5"}, {5, 2.5, "5"},
		{7.5, 2.5, "7.5"}, {10, 2.5, "10"},
		{0.125, 0.05, "0.12"}, {0.05, 0.05, "0.05"},
		{1200, 200, "1,200"},
	} {
		if got := Auto()(tc.v, tc.step); got != tc.want {
			t.Errorf("Auto()(%v, %v) = %q, want %q", tc.v, tc.step, got, tc.want)
		}
	}
	if got, want := Fixed(2)(3.14159, 0), "3.14"; got != want {
		t.Errorf("Fixed(2)(3.14159) = %q, want %q", got, want)
	}
	if got, want := Compact()(12400, 1000), "12.4k"; got != want {
		t.Errorf("Compact()(12400) = %q, want %q", got, want)
	}
	if got, want := Compact()(1200, 100), "1.2k"; got != want {
		t.Errorf("Compact()(1200) = %q, want %q", got, want)
	}
	if got, want := Compact()(1500000, 1), "1.5M"; got != want {
		t.Errorf("Compact()(1500000) = %q, want %q", got, want)
	}
	if got, want := Percent()(0.156, 0.05), "16%"; got != want {
		t.Errorf("Percent()(0.156) = %q, want %q", got, want)
	}
}

// ── the frame ─────────────────────────────────────────────────────────────

// A frame's left gutter is its y axis' widest label, measured; the right is
// nothing at all.
func TestFrameMeasuresItsGutters(t *testing.T) {
	var frame FrameResult
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		u := core.Density(c).Unit()
		frame = Frame(c, FrameOptions{
			Height: 200,
			X:      AxisOptions{Scale: NewLinear(Domain{0, 10}, 0, 1), Count: 5},
			Y:      AxisOptions{Side: Left, Scale: NewLinear(Domain{0, 10}, 0, 1), Count: 5},
		}, nil)
		// The same two numbers Measure works from, measured the same way.
		want := LabelWidth(c, "10", theme.CaptionSize) + u
		if got := frame.Inset().Left; got != want {
			t.Errorf("left gutter is %v, want the widest y label %v plus a step %v", got,
				LabelWidth(c, "10", theme.CaptionSize), u)
		}
		if got := frame.Inset().Right; got != 0 {
			t.Errorf("right gutter is %v, want nothing: a chart's data ends at its right edge", got)
		}
		wantBottom := LabelHeight(c, theme.CaptionSize) + u
		if got := frame.Inset().Bottom; got != wantBottom {
			t.Errorf("bottom gutter is %v, want a line of labels %v plus a step %v", got,
				LabelHeight(c, theme.CaptionSize), u)
		}
		// No legend and no axis name, but the topmost y tick's label is
		// centred on the plot's top edge, so half of it reaches above: the
		// frame keeps that half clear.
		wantTop := LabelHeight(c, theme.CaptionSize) / 2
		if got := frame.Inset().Top; got != wantTop {
			t.Errorf("top gutter is %v, want half a y label %v", got, wantTop)
		}
	}, 400, 200)
}

// A wider number makes a wider gutter: the frame measures, it does not guess a
// constant and hope.
func TestFrameGrowsWithItsWidestLabel(t *testing.T) {
	var narrow, wide float32
	build := func(hi float64) *float32 {
		var got float32
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			f := Frame(c, FrameOptions{Height: 200,
				Y: AxisOptions{Side: Left, Scale: NewLinear(Domain{0, hi}, 0, 1), Count: 5}}, nil)
			got = f.Inset().Left
		}, 400, 200)
		return &got
	}
	narrow, wide = *build(10), *build(1000000)
	if wide <= narrow {
		t.Errorf("a chart of 1,200,000 has the same left gutter (%v) as one of 10 (%v)", wide, narrow)
	}
	// And it is exactly that label's width plus a step.
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		f := Frame(c, FrameOptions{Height: 200,
			Y: AxisOptions{Side: Left, Scale: NewLinear(Domain{0, 1000000}, 0, 1), Count: 5}}, nil)
		want := LabelWidth(c, "1,000,000", theme.CaptionSize) + core.Density(c).Unit()
		if got := f.Inset().Left; got != want {
			t.Errorf("left gutter %v, want %v (%q measured)", got, want, "1,000,000")
		}
	}, 400, 200)
}

// The legend and an axis' name each take their own strip of the top.
func TestFrameKeepsRoomForTheLegend(t *testing.T) {
	var frame FrameResult
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		series := testSeries()
		frame = Frame(c, FrameOptions{
			Height: 200,
			Y:      AxisOptions{Side: Left, Scale: NewLinear(Domain{0, 10}, 0, 1), Count: 5},
			Legend: LegendOptions{Entries: EntriesOf(c, series)},
		}, nil)
		u := core.Density(c).Unit()
		// The legend's strip, a step clear of the plot, and half of the
		// topmost y tick's label — which is centred on the plot's top edge
		// and would otherwise read straight through the legend.
		want := LegendHeight(c, LegendOptions{Entries: EntriesOf(c, series)}) + u +
			LabelHeight(c, theme.CaptionSize)/2
		if got := frame.Inset().Top; got != want {
			t.Errorf("top gutter is %v, want the legend %v plus a step %v plus half a label", got,
				LegendHeight(c, LegendOptions{Entries: EntriesOf(c, series)}), u)
		}
	}, 400, 200)
}

// A chart with no y axis has nothing to leave room for on the left.
func TestFrameWithoutAnAxisHasNoGutterForIt(t *testing.T) {
	var frame FrameResult
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		frame = Frame(c, FrameOptions{Height: 200, Pad: 8}, nil)
	}, 400, 200)
	if got, want := frame.Inset().Left, float32(8); got != want {
		t.Errorf("left gutter is %v, want just the padding %v", got, want)
	}
	if got := frame.Plot(); got.X != 8 {
		t.Errorf("the plot starts at %v, want inside the %v padding", got.X, 8)
	}
}

// The plot is inside the frame, and never larger than it: a frame too small
// for its own gutters draws no plot rather than an upside-down one.
func TestFrameHandsOutAPlotInsideItself(t *testing.T) {
	var frame FrameResult
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		frame = Frame(c, FrameOptions{Height: 200, Pad: 10,
			X: AxisOptions{Scale: NewLinear(Domain{0, 10}, 0, 1)},
			Y: AxisOptions{Side: Left, Scale: NewLinear(Domain{0, 10}, 0, 1)}}, nil)
	}, 400, 200)
	bounds, plot := frame.Bounds(), frame.Plot()
	if plot.W <= 0 || plot.H <= 0 {
		t.Fatalf("the plot is %+v, want some room", plot)
	}
	if plot.X < bounds.X || plot.Y < bounds.Y ||
		plot.X+plot.W > bounds.X+bounds.W || plot.Y+plot.H > bounds.Y+bounds.H {
		t.Errorf("the plot %+v is not inside the frame %+v", plot, bounds)
	}
	if got, want := plot, PlotRect(bounds, frame.Inset()); got != want {
		t.Errorf("the plot %+v is not outer less the inset (%v)", got, want)
	}
	if got := PlotRect(ui.Rect{W: 10, H: 10}, Inset{Left: 40, Top: 40}); got.W != 0 || got.H != 0 {
		t.Errorf("a frame smaller than its gutters got plot %+v, want an empty one", got)
	}
}

// ── the pieces, rendered ───────────────────────────────────────────────────

// The grid line under the middle of the plot is the tick at the middle of the
// range, to the pixel.
func TestGridLinesLandOnTheirTicks(t *testing.T) {
	var frame FrameResult
	// Four steps over 0 to 10 put a tick at 5, dead centre — which is the one
	// a pixel assertion can be exact about.
	x := NewLinear(Domain{0, 10}, 0, 1)
	y := NewLinear(Domain{0, 10}, 0, 1)
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		frame = Frame(c, FrameOptions{Height: 200,
			X: AxisOptions{Scale: x, Count: 4},
			Y: AxisOptions{Side: Left, Scale: y, Count: 4}}, func(f FrameResult) {
			Grid(c, GridOptions{Frame: f,
				X: AxisOptions{Scale: x, Count: 4},
				Y: AxisOptions{Side: Left, Scale: y, Count: 4}})
		})
	}, 400, 200)

	plot := frame.Plot()
	img := tt.Image()
	border := theme.Light().Border
	// The vertical line at the middle value, well above the horizontal one at
	// the bottom edge.
	at := ui.Rect{X: plot.X + plot.W/2, Y: plot.Y + 8, W: 1, H: 4}
	if !finds(img, at, border) {
		t.Errorf("no grid line at x=%v, where the middle tick belongs", plot.X+plot.W/2)
	}
	// And one across the plot at the middle value.
	if !finds(img, ui.Rect{X: plot.X + 8, Y: plot.Y + plot.H/2, W: 4, H: 1}, border) {
		t.Errorf("no grid line at y=%v, where the middle tick belongs", plot.Y+plot.H/2)
	}
	// Halfway between two ticks there is no line: a quarter of the way along
	// is one, an eighth is between two, and a line there would be a grid line
	// this axis never asked for.
	if finds(img, ui.Rect{X: plot.X + plot.W/8 - 1, Y: plot.Y + 8, W: 3, H: 4}, border) {
		t.Errorf("a grid line was drawn at %v, between two ticks", plot.X+plot.W/8)
	}
}

// An axis is drawn, and named for assistive technology, and its labels are
// measured: two charts whose y labels differ have gutters that differ.
func TestAxisIsNamedAndMeasured(t *testing.T) {
	var frame FrameResult
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		x := NewLinear(Domain{0, 10}, 0, 1)
		y := NewLinear(Domain{0, 100}, 0, 1)
		frame = Frame(c, FrameOptions{Height: 200, Pad: 4,
			X: AxisOptions{Scale: x, Count: 5, Label: "Day", Ticks: true},
			Y: AxisOptions{Side: Left, Scale: y, Count: 4, Label: "Callbacks", Ticks: true}}, func(f FrameResult) {
			Axis(c, f, AxisOptions{Scale: x, Count: 5, Label: "Day", Ticks: true})
			Axis(c, f, AxisOptions{Side: Left, Scale: y, Count: 4, Label: "Callbacks"})
		})
	}, 400, 200)

	if !tt.HasText("Day axis") || !tt.HasText("Callbacks axis") {
		t.Errorf("the axes are not named for assistive technology: %q", tt.Texts())
	}
	// An x axis' name takes its own strip under the labels: the bottom gutter
	// is the padding, a line of tick labels with their marks, a step, and a
	// line of name.
	var labels, name, pad float32
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		pad = 4
		labels = LabelHeight(c, theme.CaptionSize) + core.Density(c).Unit()*2.5
		name = LabelHeight(c, theme.RowSize) + core.Density(c).Unit()
	}, 100, 100)
	if got, want := frame.Inset().Bottom, pad+labels+name; got != want {
		t.Errorf("the bottom gutter is %v, want %v: padding, a line of labels and a line of name", got, want)
	}
}

// A legend names every series, and takes the palette's colours where the
// series did not choose.
func TestLegendNamesEverySeries(t *testing.T) {
	var entries []LegendEntry
	var palette []ui.Color
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		series := testSeries()
		entries = EntriesOf(c, series)
		palette = Palette(c, len(series))
		Legend(c, LegendOptions{Entries: entries})
	}, 300, 60)
	for _, series := range testSeries() {
		if !tt.HasText(series.Name) {
			t.Errorf("the legend is missing %q: %q", series.Name, tt.Texts())
		}
	}
	if entries[0].Color != palette[0] || entries[1].Color != palette[1] {
		t.Errorf("the legend's colours are not the palette's: %v vs %v", entries, palette)
	}
	if entries[0].Mark != MarkLine || entries[1].Mark != MarkLine {
		t.Errorf("a line and an area are both lines in a legend, got %v and %v", entries[0].Mark, entries[1].Mark)
	}
	if entries[0].Color == entries[1].Color {
		t.Error("two series took the same colour off a palette of two")
	}
}

// The line is drawn where the data says, at the colour it was given.
func TestPlotDrawsWhereTheDataIs(t *testing.T) {
	ink := ui.Hex("#c62b30")
	var frame FrameResult
	x := NewLinear(Domain{0, 6}, 0, 1)
	y := NewLinear(Domain{0, 12}, 0, 1)
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		frame = Frame(c, FrameOptions{Height: 200,
			X: AxisOptions{Scale: x, Count: 7},
			Y: AxisOptions{Side: Left, Scale: y, Count: 6}}, func(f FrameResult) {
			Plot(c, PlotOptions{Frame: f, Width: 3,
				Series: Series{Name: "Open", Color: ink, Form: FormLine,
					Points: []Point{{X: 0, Y: 12}, {X: 6, Y: 0}}},
				X: x, Y: y})
		})
	}, 400, 200)

	plot := frame.Plot()
	// The series runs from the top left of the plot to the bottom right, so
	// each quarter of the way along is a quarter of the way down.
	for _, f := range []float32{0.25, 0.5, 0.75} {
		want := ui.Rect{X: plot.X + plot.W*f, Y: plot.Y + plot.H*f, W: 2, H: 2}
		if !finds(tt.Image(), want, ink) {
			t.Errorf("no line at %v, which is where the data should be", want)
		}
	}
	// And nowhere it is not: the top right corner is empty plot.
	if finds(tt.Image(), ui.Rect{X: plot.X + plot.W - 8, Y: plot.Y, W: 6, H: 6}, ink) {
		t.Error("the line was drawn where there is no data")
	}
}

// A bar stands on the baseline and stops at its value.
func TestPlotBarsStandOnTheBaseline(t *testing.T) {
	ink := ui.Hex("#1f7a3d")
	var frame FrameResult
	x := NewBand([]string{"New", "Root", "Fixed"}, 0, 1)
	y := NewLinear(Domain{0, 10}, 0, 1)
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		frame = Frame(c, FrameOptions{Height: 200,
			X: AxisOptions{Scale: x},
			Y: AxisOptions{Side: Left, Scale: y, Count: 5}}, func(f FrameResult) {
			Plot(c, PlotOptions{Frame: f, X: x, Y: y, Series: Series{
				Name: "n", Form: FormBar, Color: ink,
				Points: []Point{{X: 0, Y: 10}, {X: 1, Y: 5}, {X: 2, Y: 0}}}})
		})
	}, 400, 200)

	plot := frame.Plot()
	img := tt.Image()
	filled := ink.Alpha(0.85).Over(theme.Light().Background)
	xs := x.Span(plot.X, plot.X+plot.W)
	ys := y.Span(plot.Y+plot.H, plot.Y)

	// The tall bar: its top is at the top of the range, and its body reaches
	// down to the baseline.
	if !finds(img, ui.Rect{X: xs.At(0) - 4, Y: plot.Y + 2, W: 8, H: 4}, filled) {
		t.Errorf("the bar of 10 does not reach the top of the range at y=%v", plot.Y)
	}
	if !finds(img, ui.Rect{X: xs.At(0) - 4, Y: plot.Y + plot.H - 6, W: 8, H: 4}, filled) {
		t.Errorf("the bar of 10 does not stand on the baseline at y=%v", plot.Y+plot.H)
	}
	// The middle one stops halfway.
	if finds(img, ui.Rect{X: xs.At(1) - 4, Y: plot.Y + 2, W: 8, H: 4}, filled) {
		t.Error("the bar of 5 reaches the top of the range")
	}
	if !finds(img, ui.Rect{X: xs.At(1) - 4, Y: ys.At(5) - 2, W: 8, H: 4}, filled) {
		t.Errorf("the bar of 5 does not stop at its own value, y=%v", ys.At(5))
	}
	// A bar of nothing is nothing but a hair at the baseline, not a full bar.
	if finds(img, ui.Rect{X: xs.At(2) - 4, Y: plot.Y + 4, W: 8, H: plot.H/2 - 8}, filled) {
		t.Error("a bar of zero is drawn as a full-height bar")
	}
	// Bars stay inside their bands: the first one does not reach the plot's
	// left edge.
	if finds(img, ui.Rect{X: plot.X, Y: plot.Y + 8, W: 2, H: plot.H - 16}, filled) {
		t.Error("a bar is drawn over the axis")
	}
}

// A tooltip appears beside the point it is about, inside the plot, on the fill
// colour the interface draws every raised card in.
func TestTooltipAppearsBesideThePoint(t *testing.T) {
	var frame FrameResult
	var hover Hover
	x := NewLinear(Domain{0, 10}, 0, 1)
	y := NewLinear(Domain{0, 10}, 0, 1)
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		frame = Frame(c, FrameOptions{Height: 200, Hover: &hover,
			X: AxisOptions{Scale: x, Count: 5},
			Y: AxisOptions{Side: Left, Scale: y, Count: 5}}, func(f FrameResult) {
			Tooltip(c, TooltipOptions{Frame: f, Hover: &hover, Title: "Tuesday",
				Rows: []TooltipRow{{Name: "Open", Color: theme.Light().Accent, Value: "12"}}})
		})
	}, 400, 200)

	tt.Move(120, 90)
	tt.Frame()
	if !hover.Over {
		t.Fatal("the pointer at (120, 90) should be over the plot")
	}
	img := tt.Image()
	plot := frame.Plot()
	fill := theme.Light().Fill
	// The card sits to the right of the pointer, centred on the plot.
	card := ui.Rect{
		X: hover.X + 4,
		Y: plot.Y + (plot.H-40)/2,
		W: plot.W,
		H: 40,
	}
	if !finds(img, card, fill) {
		t.Errorf("no tooltip card to the right of x=%v", hover.X)
	}
	// Nothing to the left of the pointer: the card is on one side or the
	// other, not both.
	if finds(img, ui.Rect{X: plot.X, Y: card.Y, W: hover.X - plot.X - 6, H: card.H}, fill) {
		t.Error("the tooltip was drawn on both sides of its point")
	}
	// A chart with no pointer draws no tooltip.
	tt.Move(2, 2)
	tt.Frame()
	if finds(tt.Image(), card, fill) {
		t.Error("the tooltip stayed drawn with the pointer off the chart")
	}
}

// An empty chart says what is missing instead of drawing an axis over nothing.
func TestEmptySaysWhatIsMissing(t *testing.T) {
	var frame FrameResult
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		frame = Frame(c, FrameOptions{Height: 200}, func(f FrameResult) {
			Empty(c, EmptyOptions{Frame: f, Title: "Nothing resolved yet",
				Body: "No callbacks were resolved today."})
		})
	}, 400, 200)

	if !tt.HasText("Nothing resolved yet. No callbacks were resolved today.") {
		t.Errorf("the empty state is not named for assistive technology: %q", tt.Texts())
	}
	plot := frame.Plot()
	// Its mark is drawn: the one bright colour in the interface, inside the
	// plot area and nowhere else.
	if !finds(tt.Image(), plot, theme.Light().Lively) {
		t.Error("the empty state drew no mark in the plot area")
	}
	// And it drew no scale: nothing in the plot but the mark. (The frame's
	// own two edges are outside the area this looks at.)
	inside := ui.Rect{X: plot.X + 2, Y: plot.Y + 2, W: plot.W - 4, H: plot.H - 4}
	if finds(tt.Image(), inside, theme.Light().Border) {
		t.Error("an empty chart drew a grid line or an axis over nothing")
	}
}

// A reference line is drawn at the value it stands for, not at some other
// height.
func TestAnnotationDrawsItsLineAtTheValue(t *testing.T) {
	accent := theme.Light().Accent
	var frame FrameResult
	x := NewLinear(Domain{0, 10}, 0, 1)
	y := NewLinear(Domain{0, 10}, 0, 1)
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		frame = Frame(c, FrameOptions{Height: 200,
			X: AxisOptions{Scale: x, Count: 5},
			Y: AxisOptions{Side: Left, Scale: y, Count: 5}}, func(f FrameResult) {
			Annotation(c, AnnotationOptions{Frame: f, X: x, Y: y,
				Line: &Line{Value: 7.5, Text: "target", Dashed: true}})
		})
	}, 400, 200)

	plot := frame.Plot()
	img := tt.Image()
	ys := y.Span(plot.Y+plot.H, plot.Y)
	if !finds(img, ui.Rect{X: plot.X + 20, Y: ys.At(7.5) - 2, W: 20, H: 5}, accent) {
		t.Errorf("no reference line at y=%v, where the value 7.5 belongs", ys.At(7.5))
	}
	if finds(img, ui.Rect{X: plot.X + 20, Y: ys.At(2.5) - 2, W: 20, H: 5}, accent) {
		t.Error("the reference line was drawn at the wrong height")
	}
}

func TestAnnotationNeedsSomethingToSay(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("an annotation with nothing to annotate should panic")
		}
	}()
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Frame(c, FrameOptions{Height: 200}, func(fr FrameResult) {
			Annotation(c, AnnotationOptions{Frame: fr})
		})
	}, 300, 200)
}

// The crosshair follows the pointer, and only follows it while it is over the
// chart: a rule drawn through the middle of somebody's data with no pointer
// near it is a chart lying about where the reader is looking.
func TestCrosshairFollowsTheHover(t *testing.T) {
	var frame FrameResult
	var hover Hover
	x := NewLinear(Domain{0, 10}, 0, 1)
	y := NewLinear(Domain{0, 10}, 0, 1)
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		frame = Frame(c, FrameOptions{Height: 200, Hover: &hover,
			X: AxisOptions{Scale: x, Count: 5},
			Y: AxisOptions{Side: Left, Scale: y, Count: 5}}, func(f FrameResult) {
			Crosshair(c, CrosshairOptions{Frame: f, Hover: &hover,
				ValueX: "5.0", ValueY: "4.2"})
		})
	}, 400, 200)

	tt.Move(150, 80)
	tt.Frame()
	plot := frame.Plot()
	if !hover.Over {
		t.Fatalf("(150, 80) should be over the plot %+v", plot)
	}
	if hover.X != 150 || hover.Y != 80 {
		t.Errorf("the hover is at (%v, %v), want the pointer's (150, 80)", hover.X, hover.Y)
	}
	// The values under the pointer, read through the scales.
	xs := x.Span(plot.X, plot.X+plot.W)
	ys := y.Span(plot.Y+plot.H, plot.Y)
	if math.Abs(hover.ValueX-xs.Value(150)) > 0.001 || math.Abs(hover.ValueY-ys.Value(80)) > 0.001 {
		t.Errorf("the hover reads (%v, %v) where the scales say (%v, %v)",
			hover.ValueX, hover.ValueY, xs.Value(150), ys.Value(80))
	}
	img := tt.Image()
	rule := theme.Light().TextMuted.Alpha(0.55).Over(theme.Light().Background)
	if !finds(img, ui.Rect{X: 149, Y: plot.Y + 10, W: 3, H: 4}, rule) {
		t.Errorf("no vertical rule at the pointer's x=%v", hover.X)
	}
	if !finds(img, ui.Rect{X: plot.X + 10, Y: 79, W: 4, H: 3}, rule) {
		t.Errorf("no horizontal rule at the pointer's y=%v", hover.Y)
	}
	// Off the chart, the rules are gone.
	tt.Move(150, 260)
	tt.Frame()
	if hover.Over {
		t.Error("a pointer below the chart should not be over the plot")
	}
	if finds(tt.Image(), ui.Rect{X: 149, Y: plot.Y + 10, W: 3, H: 4}, rule) {
		t.Error("the crosshair stayed drawn with the pointer off the chart")
	}
}

// A brush is a drag that selects a range of the axis, and the values it
// catches are the caller's.
func TestBrushSelectsARangeOfTheAxis(t *testing.T) {
	var frame FrameResult
	var sel Range
	var changed bool
	x := NewLinear(Domain{0, 10}, 0, 1)
	y := NewLinear(Domain{0, 10}, 0, 1)
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		frame = Frame(c, FrameOptions{Height: 200,
			X: AxisOptions{Scale: x, Count: 5},
			Y: AxisOptions{Side: Left, Scale: y, Count: 5}}, func(f FrameResult) {
			brush := Brush(c, BrushOptions{Frame: f, X: x, Range: &sel, Handles: true})
			if brush.Changed() {
				changed = true
			}
		})
	}, 400, 200)

	if sel.Active {
		t.Fatal("a brush that has not been dragged holds no selection")
	}
	plot := frame.Plot()
	xs := x.Span(plot.X, plot.X+plot.W)
	from, to := plot.X+plot.W*0.25, plot.X+plot.W*0.75

	// A drag is pressed, moved in steps and released: one jump is not a drag.
	tt.Press(from, plot.Y+plot.H/2)
	for i := 1; i <= 4; i++ {
		tt.Move(from+(to-from)*float32(i)/4, plot.Y+plot.H/2)
	}
	tt.Release(to, plot.Y+plot.H/2)
	tt.Frame()

	if !sel.Active {
		t.Fatalf("the drag selected nothing: %+v", sel)
	}
	if !changed {
		t.Error("the drag did not report that the selection moved")
	}
	if sel.Dragging {
		t.Error("the drag is over; the range should have settled")
	}
	if wantMin, wantMax := xs.Value(from), xs.Value(to); math.Abs(sel.Min-wantMin) > 0.01 ||
		math.Abs(sel.Max-wantMax) > 0.01 {
		t.Errorf("selected %v..%v, want %v..%v", sel.Min, sel.Max, wantMin, wantMax)
	}
	if sel.Span() < 4 {
		t.Errorf("a quarter of the axis selected a span of %v", sel.Span())
	}
	// The selection is drawn as a wash over the chart.
	if !finds(tt.Image(), plot, theme.Light().Accent.Alpha(0.12).Over(theme.Light().Background)) {
		t.Error("the selection was not drawn over the chart")
	}
	// A click on a brush clears it: a press and a release in the same place
	// is a click, not a range.
	tt.Press(plot.X+plot.W/2, plot.Y+plot.H/2)
	tt.Release(plot.X+plot.W/2, plot.Y+plot.H/2)
	tt.Frame()
	if sel.Active || sel.Min != 0 || sel.Max != 0 {
		t.Errorf("a click left a selection behind: %+v", sel)
	}
}

func TestBrushNeedsARangeToPutWhatItCatches(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("a brush with nowhere to put its selection should panic")
		}
	}()
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		x := NewLinear(Domain{0, 10}, 0, 1)
		Frame(c, FrameOptions{Height: 200}, func(f FrameResult) {
			Brush(c, BrushOptions{Frame: f, X: x})
		})
	}, 300, 200)
}

// A chart's axes thin their labels until they fit: the ones that are dropped
// are the ones in the middle, and no two of the ones left touch.
func TestAxisThinsLabelsUntilTheyFit(t *testing.T) {
	var frame FrameResult
	x := NewLinear(Domain{0, 100}, 0, 1)
	opts := AxisOptions{Scale: x, Count: 12, Size: theme.RowSize}
	// A 240-wide window for thirteen labels: they cannot all fit, so some of
	// them have to go.
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		frame = Frame(c, FrameOptions{Height: 60, X: opts}, func(fr FrameResult) {
			Axis(c, fr, opts)
		})
	}, 240, 80)

	// The ticks the axis settled on, read from the plot it was given — the
	// set a reader actually sees, measured the way the axis measures them.
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		gap := core.Density(c).Unit()
		plot := frame.Plot()
		ticks := Ticks(c, x.Span(plot.X, plot.X+plot.W),
			TickOptions{Count: opts.Count, Size: opts.Size})

		drawn := 0
		var prevLabel string
		var prevEnd float32
		for _, tk := range ticks {
			if !tk.Drawn {
				continue
			}
			drawn++
			start := tk.Pos - LabelWidth(c, tk.Label, opts.Size)/2
			if prevLabel != "" && start < prevEnd+gap {
				t.Errorf("labels %q and %q overlap: the second starts at %v, the first ended at %v",
					prevLabel, tk.Label, start, prevEnd)
			}
			prevLabel, prevEnd = tk.Label, start+LabelWidth(c, tk.Label, opts.Size)
		}
		if drawn == len(ticks) {
			t.Errorf("all %d labels were drawn in %v DIPs; a narrow chart has to thin them",
				drawn, plot.W)
		}
		if drawn < 2 {
			t.Errorf("thinning left %d label; an axis has to keep both ends", drawn)
		}
		if !ticks[0].Drawn || !ticks[len(ticks)-1].Drawn {
			t.Errorf("the ends of the range were dropped: %v", ticks)
		}
	}, 240, 80)
}

// ── a whole chart ─────────────────────────────────────────────────────────

// A chart of the library's own pieces draws without a window, a pointer or a
// clock, over several frames, and names everything it shows.
func TestAWholeChartDraws(t *testing.T) {
	var frame FrameResult
	var hover Hover
	var brush Range
	series := testSeries()
	domain, ok := SeriesDomain(series)
	if !ok {
		t.Fatal("the test data has no y values")
	}
	if got, want := domain, (Domain{2, 12}); got != want {
		t.Errorf("the data spans %+v, want %+v", got, want)
	}
	yd := Nice(domain.Min, domain.Max, 4)
	xs := NewLinear(Domain{0, 6}, 0, 1)
	ys := NewLinear(yd, 0, 1)

	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		frame = Frame(c, FrameOptions{
			Height: 220,
			Label:  "Open callbacks by day",
			X:      AxisOptions{Scale: xs, Count: 7, Label: "Day"},
			Y:      AxisOptions{Side: Left, Scale: ys, Count: 5, Label: "Callbacks", Ticks: true},
			Legend: LegendOptions{Entries: EntriesOf(c, series)},
			Hover:  &hover,
			Border: true,
		}, func(f FrameResult) {
			Grid(c, GridOptions{Frame: f,
				X: AxisOptions{Scale: xs, Count: 7},
				Y: AxisOptions{Side: Left, Scale: ys, Count: 5}})
			for _, s := range series {
				Plot(c, PlotOptions{Frame: f, Series: s, X: xs, Y: ys, Dots: true})
			}
			Annotation(c, AnnotationOptions{Frame: f, X: xs, Y: ys,
				Line: &Line{Value: 8, Text: "target", Dashed: true}})
			Brush(c, BrushOptions{Frame: f, X: xs, Range: &brush})
			Crosshair(c, CrosshairOptions{Frame: f, Hover: &hover})
			Axis(c, f, AxisOptions{Scale: xs, Count: 7, Label: "Day"})
			Axis(c, f, AxisOptions{Side: Left, Scale: ys, Count: 5, Label: "Callbacks", Ticks: true})
			Tooltip(c, TooltipOptions{Frame: f, Hover: &hover, Title: "Tuesday",
				Rows: []TooltipRow{
					{Name: "Open", Color: EntriesOf(c, series)[0].Color, Value: "9"},
					{Name: "Closed", Color: EntriesOf(c, series)[1].Color, Value: "6"},
				}})
		})
	}, 520, 260)

	for range 3 {
		tt.Move(260, 120)
		tt.Frame()
	}
	for _, want := range []string{
		"Open callbacks by day", "Open", "Closed", "Day axis", "Callbacks axis",
		"Open series", "Closed series", "Series",
	} {
		if !tt.HasText(want) {
			t.Errorf("the chart is missing %q: %q", want, tt.Texts())
		}
	}
	plot := frame.Plot()
	if plot.W < 300 || plot.H < 100 {
		t.Errorf("the plot is %+v, want a chart's worth of room", plot)
	}
	if frame.Inset().Right != 0 {
		t.Errorf("right gutter %v, want nothing", frame.Inset().Right)
	}
	// The nice domain puts its own ceiling at the plot's top and its floor at
	// the bottom, so a point's place on the chart is its value's.
	placed := ys.Span(plot.Y+plot.H, plot.Y)
	if got, want := placed.At(yd.Max), float32(plot.Y); math.Abs(float64(got-want)) > 0.01 {
		t.Errorf("the top of the range %v is at y=%v, want the plot's top %v", yd.Max, got, want)
	}
	if got, want := placed.At(yd.Min), plot.Y+plot.H; math.Abs(float64(got-want)) > 0.01 {
		t.Errorf("the bottom of the range %v is at y=%v, want the plot's bottom %v", yd.Min, got, want)
	}
	if got, want := placed.At(12), plot.Y+plot.H*0.2; math.Abs(float64(got-want)) > 0.5 {
		t.Errorf("the value 12 is at y=%v, want %v on a range of 0 to %v", got, want, yd.Max)
	}
}

// The dark appearance is the desktop's, so a chart must not hardcode a
// palette — and a palette that reads in both is one set, not two.
func TestAWholeChartInTheDark(t *testing.T) {
	var light, dark []ui.Color
	render := func(mode core.Mode, got *[]ui.Color) *ui.Tester {
		series := testSeries()
		domain, _ := SeriesDomain(series)
		x := NewLinear(Domain{0, 6}, 0, 1)
		y := NewLinear(Nice(domain.Min, domain.Max, 4), 0, 1)
		return ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{Mode: mode})
			*got = Palette(c, len(series))
			Frame(c, FrameOptions{Height: 200,
				X:      AxisOptions{Scale: x, Count: 7},
				Y:      AxisOptions{Side: Left, Scale: y, Count: 5},
				Legend: LegendOptions{Entries: EntriesOf(c, series)}}, func(fr FrameResult) {
				Grid(c, GridOptions{Frame: fr,
					X: AxisOptions{Scale: x, Count: 7},
					Y: AxisOptions{Side: Left, Scale: y, Count: 5}})
				Plot(c, PlotOptions{Frame: fr, Series: series[0], X: x, Y: y, Dots: true})
			})
		}, 400, 200)
	}
	tt := render(core.Light, &light)
	render(core.Dark, &dark)

	if !tt.HasText("Open") || !tt.HasText("Closed") {
		t.Errorf("the dark chart lost its legend: %q", tt.Texts())
	}
	// The wide-gamut part of a colour is the hue the accent happened to have,
	// so it is compared on what every screen shows: the sRGB.
	if !equalSRGB(light, dark) {
		t.Errorf("the palette changed with the appearance: %v vs %v", light, dark)
	}
	// The same palette has to clear the dark window too, which is what makes
	// it one palette rather than two.
	for i, c := range dark {
		if ratio := contrast(c, theme.Dark().Background); ratio < 2.5 {
			t.Errorf("colour %d is only %.2f:1 on a dark window", i, ratio)
		}
	}
}

// nil2ctx is a context with the library's light palette installed, for the
// few assertions about what a window resolves that do not need a window.
func nil2ctx() *ui.Context {
	var c *ui.Context
	ui.NewTester(func(got *ui.Context) {
		core.Use(got, core.Settings{})
		c = got
	}, 40, 40)
	return c
}

// contrast is the WCAG ratio of two opaque colours.
func contrast(a, b ui.Color) float64 {
	la, lb := luminance(a)+0.05, luminance(b)+0.05
	if la < lb {
		la, lb = lb, la
	}
	return la / lb
}

func luminance(c ui.Color) float64 {
	channel := func(v uint8) float64 {
		s := float64(v) / 255
		if s <= 0.03928 {
			return s / 12.92
		}
		return math.Pow((s+0.055)/1.055, 2.4)
	}
	return 0.2126*channel(c.R) + 0.7152*channel(c.G) + 0.0722*channel(c.B)
}

// saturation is how far a colour is from the grey of its own brightness.
func saturation(c ui.Color) float64 {
	hi, lo := math.Max(float64(c.R), math.Max(float64(c.G), float64(c.B))),
		math.Min(float64(c.R), math.Min(float64(c.G), float64(c.B)))
	if hi == 0 {
		return 0
	}
	return (hi - lo) / hi
}
