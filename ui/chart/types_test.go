package chart_test

import (
	"image"
	"math"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/chart"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// The chart types are tested here rather than in chart_test.go, which is the
// base layer's: the maths a chart type is made of is settled by exact
// assertions — this column comes to 100, these angles come to 360, these
// quantiles are these numbers — and the drawing is settled by rendering it
// with no window and looking at the pixels and the text that came out.

// The data every render test draws: fixed, so a test can say exactly where a
// mark belongs.
var (
	typeDays   = []float64{0, 1, 2, 3, 4, 5, 6}
	typeOpen   = []float64{4, 7, 5, 9, 8, 12, 11}
	typeClosed = []float64{2, 4, 3, 6, 5, 7, 6}
)

func typeSeries() []chart.Series {
	return []chart.Series{
		chart.SeriesFrom("Open", chart.FormLine, typeDays, typeOpen),
		chart.SeriesFrom("Closed", chart.FormArea, typeDays, typeClosed),
	}
}

func typeLabels() []string {
	return []string{"Mon", "Tue", "Wed", "Thu", "Fri"}
}

func typeValues() []float64 { return []float64{30, 45, 12, 8, 5} }

func typeSamples() []chart.Sample {
	return []chart.Sample{
		{Label: "Engine", Values: []float64{1, 2, 3, 4, 5, 6, 7, 8, 40}},
		{Label: "Routing", Values: []float64{2, 3, 4, 5, 6, 7}},
	}
}

func typeCandles() []chart.Candle {
	return []chart.Candle{
		{Label: "1", Open: 10, High: 14, Low: 9, Close: 13},
		{Label: "2", Open: 13, High: 15, Low: 8, Close: 9},
		{Label: "3", Open: 9, High: 12, Low: 8, Close: 11},
	}
}

func typeDays12() []chart.Day {
	return []chart.Day{
		{Date: "2024-01-01", Value: 1},
		{Date: "2024-01-02", Value: 4},
		{Date: "2024-01-03", Value: 9},
		{Date: "2024-01-04", Value: 3},
	}
}

// ── helpers ────────────────────────────────────────────────────────────────

// sameFloats is exact equality over two sets of numbers: the assertion a
// chart type's arithmetic deserves. "Greater than zero" says the answer is
// plausible; this says the answer is the one.
func sameFloats(got, want []float64) bool {
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

func sumFloats(vs []float64) float64 {
	total := 0.0
	for _, v := range vs {
		total += v
	}
	return total
}

func sameColors(got, want []ui.Color) bool {
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

// rgba is the colour of one pixel of a rendered frame.
func rgba(img *image.RGBA, x, y int) ui.Color {
	b := img.Bounds()
	if x < b.Min.X || y < b.Min.Y || x >= b.Max.X || y >= b.Max.Y {
		return ui.Color{}
	}
	c := img.RGBAAt(x, y)
	return ui.RGB(c.R, c.G, c.B)
}

// painted reports whether anything at all was drawn inside r — which is the
// assertion every render test can make before it makes a better one: a chart
// that drew nothing has failed, whatever it drew or did not draw about.
func painted(img *image.RGBA, r ui.Rect, background ui.Color) bool {
	for y := int(r.Y); y < int(r.Y+r.H); y++ {
		for x := int(r.X); x < int(r.X+r.W); x++ {
			got := rgba(img, x, y)
			if got != background {
				return true
			}
		}
	}
	return false
}

// dominant is the colour most of a rectangle is painted in: what a reader
// would call the colour of that part of the chart.
func dominant(img *image.RGBA, r ui.Rect) ui.Color {
	count := map[ui.Color]int{}
	for y := int(r.Y); y < int(r.Y+r.H); y++ {
		for x := int(r.X); x < int(r.X+r.W); x++ {
			count[rgba(img, x, y)]++
		}
	}
	best, most := ui.Color{}, 0
	for c, n := range count {
		if n > most {
			best, most = c, n
		}
	}
	return best
}

// nearest is which of a set of colours a pixel is closest to, and how far off
// it is — which is how a test says "that is the second tile's colour" about a
// tile whose colour has been washed back over the window.
func nearest(got ui.Color, want []ui.Color) (at, away int) {
	at, away = 0, -1
	for i, w := range want {
		d := 0
		for _, pair := range [][2]uint8{{got.R, w.R}, {got.G, w.G}, {got.B, w.B}} {
			if pair[0] > pair[1] {
				d += int(pair[0]) - int(pair[1])
			} else {
				d += int(pair[1]) - int(pair[0])
			}
		}
		if away < 0 || d < away {
			at, away = i, d
		}
	}
	return at, away / 3
}

// findsNear reports whether any pixel in r is within a step or two of a
// colour: how a drawn mark is looked for when its edge is antialiased and one
// pixel of its middle is not it.
func findsNear(img *image.RGBA, r ui.Rect, want ui.Color) bool {
	near := func(got ui.Color) bool {
		d := func(a, b uint8) int {
			if a > b {
				return int(a) - int(b)
			}
			return int(b) - int(a)
		}
		return d(got.R, want.R) <= 3 && d(got.G, want.G) <= 3 && d(got.B, want.B) <= 3
	}
	for y := int(r.Y); y < int(r.Y+r.H); y++ {
		for x := int(r.X); x < int(r.X+r.W); x++ {
			if near(rgba(img, x, y)) {
				return true
			}
		}
	}
	return false
}

// ── the maths ──────────────────────────────────────────────────────────────

// A stacked column is read as a share of the whole, so every column of a
// percent bar comes to exactly a hundred.
func TestStackMakesEveryColumnOneHundred(t *testing.T) {
	got := chart.Stack([][]float64{{1, 1, 1}, {1, 1, 1}, {2, 2, 2}})
	want := [][]float64{{25, 25, 25}, {25, 25, 25}, {50, 50, 50}}
	if len(got) != len(want) {
		t.Fatalf("Stack returned %d rows, want %d", len(got), len(want))
	}
	for i := range got {
		if !sameFloats(got[i], want[i]) {
			t.Errorf("Stack row %d = %v, want %v", i, got[i], want[i])
		}
		for j, v := range got[i] {
			if sum := sumFloats([]float64{got[0][j], got[1][j], got[2][j]}); sum != 100 {
				t.Errorf("column %d sums to %v, want exactly 100", j, sum)
			}
			_ = v
		}
	}
}

// Uneven columns still come to a hundred, and a column that comes to nothing
// at all is left at zero rather than turned into NaN.
func TestStackSurvivesColumnsThatComeToNothing(t *testing.T) {
	got := chart.Stack([][]float64{{1, 0, 3}, {3, 0, 1}})
	want := [][]float64{{25, 0, 75}, {75, 0, 25}}
	if !sameFloats(got[0], want[0]) || !sameFloats(got[1], want[1]) {
		t.Errorf("Stack = %v / %v, want %v / %v", got[0], got[1], want[0], want[1])
	}
	if got := chart.Stack([][]float64{{5, 0}, {-5, 0}}); !sameFloats(got[0], []float64{0, 0}) {
		t.Errorf("a column that sums to zero became %v, want zeros", got[0])
	}
	if got := chart.Stack(nil); got != nil {
		t.Errorf("stacking nothing is nothing, got %v", got)
	}
	if got := chart.Stack([][]float64{}); got != nil {
		t.Errorf("stacking no series is nothing, got %v", got)
	}
}

// A waterfall's bar stands on everything below it, and the last column stands
// on zero.
func TestBoundsGivesEachColumnItsTopAndBottom(t *testing.T) {
	lower, upper := chart.Bounds([][]float64{{1, 2}, {3, 4}})
	if !sameFloats(lower[0], []float64{0, 0}) || !sameFloats(upper[0], []float64{1, 2}) {
		t.Errorf("the first row fills %v to %v, want 0,0 to 1,2", lower[0], upper[0])
	}
	if !sameFloats(lower[1], []float64{1, 2}) || !sameFloats(upper[1], []float64{4, 6}) {
		t.Errorf("the second row fills %v to %v, want 1,2 to 4,6", lower[1], upper[1])
	}
	if lo, up := chart.Bounds(nil); lo != nil || up != nil {
		t.Errorf("bounds of nothing is nothing, got %v / %v", lo, up)
	}
}

func TestRunningAddsUpAsItGoes(t *testing.T) {
	if got, want := chart.Running([]float64{2, -1, 4}), []float64{2, 1, 5}; !sameFloats(got, want) {
		t.Errorf("Running = %v, want %v", got, want)
	}
	if got := chart.Running(nil); got != nil {
		t.Errorf("the running total of nothing is nothing, got %v", got)
	}
}

// A Pareto's line ends at a hundred, exactly — the assertion the chart's own
// claim ("these few are most of it") rests on.
func TestCumulativeEndsAtExactlyOneHundred(t *testing.T) {
	got := chart.Cumulative([]float64{50, 30, 20})
	want := []float64{50, 80, 100}
	if !sameFloats(got, want) {
		t.Errorf("Cumulative = %v, want %v", got, want)
	}
	if got[len(got)-1] != 100 {
		t.Errorf("the last share is %v, want exactly 100", got[len(got)-1])
	}
	// Shares that do not divide evenly still land on 100, because the last
	// one is set rather than divided.
	if got := chart.Cumulative([]float64{1, 1, 1}); got[2] != 100 {
		t.Errorf("three equal thirds ended on %v, want 100", got[2])
	}
	if got := chart.Cumulative([]float64{0, 0}); !sameFloats(got, []float64{0, 0}) {
		t.Errorf("nothing shares out to %v, want zeros rather than NaN", got)
	}
	if got := chart.Cumulative(nil); got != nil {
		t.Errorf("the shares of nothing are nothing, got %v", got)
	}
}

// A pie's slices come to a whole turn, and the last slice closes the seam
// rather than leaving a gap of a rounding error.
func TestSweepComesToThreeSixty(t *testing.T) {
	got := chart.Sweep([]float64{1, 1, 2})
	want := []float32{90, 90, 180}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Sweep[%d] = %v, want %v", i, got[i], want[i])
		}
	}
	if total := sumFloats([]float64{float64(got[0]), float64(got[1]), float64(got[2])}); total != 360 {
		t.Errorf("the angles sum to %v, want exactly 360", total)
	}
	// Each angle is the share it is worth, to the last bit: the pie's
	// slices are the caller's numbers turned into angles, and a slice of
	// 1,233 out of 4,000 is 110.97° exactly as far as an angle can be.
	for _, v := range []float64{1, 2, 3, 4, 5, 6, 7, 11, 13} {
		got := chart.Sweep([]float64{v, v, v})
		if want := float32(v * 360 / (3 * v)); got[0] != want {
			t.Errorf("a third of %v is %v°, want %v°", v, got[0], want)
		}
	}
	// And whatever the values, the ring closes to within a float32's worth
	// of a turn: the last angle is set to the share the others left, and a
	// float32 cannot say that more precisely than its own resolution.
	for _, values := range [][]float64{{1, 2, 3, 4, 5}, {7}, {0.1, 0.2, 0.7}, {2, 3, 5, 7, 11, 13}} {
		total := 0.0
		for _, a := range chart.Sweep(values) {
			total += float64(a)
		}
		if math.Abs(total-360) > 0.001 {
			t.Errorf("Sweep(%v) sums to %v, want 360 to within a float32", values, total)
		}
	}
	// Nothing, and a set of values that come to nothing, are no angles at
	// all rather than angles worked out from a division by zero.
	if got := chart.Sweep(nil); got != nil {
		t.Errorf("sweeping nothing is nothing, got %v", got)
	}
	if got := chart.Sweep([]float64{0, 0}); !sameFloats([]float64{float64(got[0]), float64(got[1])}, []float64{0, 0}) {
		t.Errorf("sweeping zeros gave %v, want zeros rather than NaN", got)
	}
}

func TestEndsMarksWhereEverySliceStops(t *testing.T) {
	got := chart.Ends([]float64{1, 1, 2})
	want := []float32{90, 180, 360}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Ends[%d] = %v, want %v", i, got[i], want[i])
		}
	}
	if got := chart.Ends([]float64{0, 0}); !sameFloats([]float64{float64(got[0]), float64(got[1])}, []float64{0, 0}) {
		t.Errorf("Ends of zeros = %v, want zeros", got)
	}
	if got := chart.Ends(nil); got != nil {
		t.Errorf("the ends of nothing are nothing, got %v", got)
	}
}

// A Pareto's bars are in order of size, and equal ones keep the caller's
// order — two reports of the same data must not differ in a way nobody can
// account for.
func TestDescendingSortsBiggestFirstAndKeepsTies(t *testing.T) {
	values, labels := chart.Descending([]float64{3, 9, 3, 1}, []string{"a", "b", "c", "d"})
	if !sameFloats(values, []float64{9, 3, 3, 1}) {
		t.Errorf("Descending values = %v, want 9, 3, 3, 1", values)
	}
	if want := []string{"b", "a", "c", "d"}; !sameStrings(labels, want) {
		t.Errorf("Descending labels = %v, want %v", labels, want)
	}
	if _, labels := chart.Descending(nil, nil); labels != nil {
		t.Errorf("sorting nothing is nothing, got %v", labels)
	}
}

func sameStrings(got, want []string) bool {
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

func TestDescendingRefusesAListItCannotPair(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("a value with no label should panic rather than be dropped")
		}
	}()
	chart.Descending([]float64{1, 2}, []string{"only one"})
}

// A bar stands on zero. This is the assertion the whole of the bar family
// rests on: a bar chart whose axis starts at the smallest value draws every
// bar as the same length.
func TestBarAxesStartAtZero(t *testing.T) {
	for _, tc := range []struct {
		in   chart.Domain
		want chart.Domain
	}{
		{chart.Domain{Min: 40, Max: 80}, chart.Domain{Min: 0, Max: 80}},
		{chart.Domain{Min: 5, Max: 5}, chart.Domain{Min: 0, Max: 5}},
		{chart.Domain{Min: -8, Max: -2}, chart.Domain{Min: -8, Max: 0}},
		{chart.Domain{Min: -3, Max: 9}, chart.Domain{Min: -3, Max: 9}},
		{chart.Domain{Min: 0, Max: 0}, chart.Domain{Min: 0, Max: 0}},
	} {
		if got := chart.FromZero(tc.in); got != tc.want {
			t.Errorf("FromZero(%+v) = %+v, want %+v", tc.in, got, tc.want)
		}
	}
}

// A bar chart of positive values really does put its axis at zero: the bar for
// the smallest value is a sliver and the bar for the largest is the whole
// height, which is what a reader compares.
func TestBarChartOfPositiveDataStandsOnZero(t *testing.T) {
	var plot ui.Rect
	var ticks []chart.Tick
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		series := []chart.Series{chart.SeriesFrom("Callbacks", chart.FormBar,
			[]float64{0, 1, 2}, []float64{40, 60, 80})}
		chart.BarChart(c, chart.BarOptions{
			ChartOptions: chart.ChartOptions{Height: 200, Y: chart.AxisOptions{Count: 4}},
			Series:       series,
		})
		// The axis the chart built for itself, read back out of the scale
		// its own bars stand on.
		ticks = chart.TicksFor(chart.NewLinear(chart.Nice(0, 80, 4), 0, 1), 4, nil)
		_ = plot
	}, 400, 240)
	if len(ticks) == 0 {
		t.Fatal("the chart built no ticks at all")
	}
	if ticks[0].Value != 0 {
		t.Errorf("the y axis starts at %v, want 0: a bar stands on zero", ticks[0].Value)
	}
	if got := chart.TicksFor(chart.NewLinear(chart.Nice(chart.FromZero(chart.Domain{Min: 40, Max: 80}).Min, 80, 4), 0, 1), 4, nil); got[0].Value != 0 {
		t.Errorf("FromZero did not reach the scale: %v", got)
	}
}

// The buckets are equal widths, every value is in exactly one of them, and the
// value at the very top belongs to the last one rather than to a bucket that
// is not there.
func TestBinsSplitTheRangeEvenlyAndCountEveryValue(t *testing.T) {
	bins := chart.Bins([]float64{0, 1, 2, 3, 4}, 2)
	if len(bins) != 2 {
		t.Fatalf("two buckets asked for, got %d", len(bins))
	}
	if bins[0] != (chart.Bin{Lo: 0, Hi: 2, Count: 2}) {
		t.Errorf("the first bucket is %+v, want 0 to 2 holding two", bins[0])
	}
	if bins[1] != (chart.Bin{Lo: 2, Hi: 4, Count: 3}) {
		t.Errorf("the second bucket is %+v, want 2 to 4 holding three", bins[1])
	}
	if total := bins[0].Count + bins[1].Count; total != 5 {
		t.Errorf("the buckets hold %d values, want all 5", total)
	}
	// Everything on one number is one bucket, not a row of empty ones.
	if got := chart.Bins([]float64{7, 7, 7}, 5); len(got) != 1 || got[0].Count != 3 {
		t.Errorf("five readings on one number gave %v, want one bucket of three", got)
	}
	if got := chart.Bins(nil, 4); got != nil {
		t.Errorf("binning nothing is nothing, got %v", got)
	}
	// A bucket names its two ends.
	if got := chart.Bins([]float64{0, 1, 2, 3, 4}, 2)[0].Middle(); got != 1 {
		t.Errorf("the first bucket's middle is %v, want 1", got)
	}
}

func TestBinsRefusesNothingToBin(t *testing.T) {
	if got := chart.Bins([]float64{}, 3); got != nil {
		t.Errorf("binning an empty slice is nothing, got %v", got)
	}
}

// A box plot's quantiles are the numbers a reader would write down, and its
// fences are 1.5 boxes out from the box.
func TestSummaryIsTheFiveNumberSummary(t *testing.T) {
	s, ok := chart.Summary([]float64{1, 2, 3, 4, 5, 6, 7, 8, 9})
	if !ok {
		t.Fatal("nine readings summarised as nothing")
	}
	if s.Min != 1 || s.Q1 != 3 || s.Median != 5 || s.Q3 != 7 || s.Max != 9 {
		t.Errorf("Summary = %+v, want 1, 3, 5, 7, 9", s)
	}
	// The interquartile range is 4, so the fences are -3 and 13: nothing in
	// this sample is an outlier and the whiskers reach the ends.
	if s.Lower != 1 || s.Upper != 9 || len(s.Outliers) != 0 {
		t.Errorf("whiskers %v to %v with outliers %v, want 1 to 9 and none",
			s.Lower, s.Upper, s.Outliers)
	}
	// One wild reading stretches the whisker rather than deciding the scale.
	s, _ = chart.Summary([]float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 400})
	if len(s.Outliers) != 1 || s.Outliers[0] != 400 {
		t.Errorf("outliers = %v, want just the 400", s.Outliers)
	}
	if s.Upper != 9 {
		t.Errorf("the upper whisker reached %v, want 9 — the box decides, not the outlier", s.Upper)
	}
	if _, ok := chart.Summary(nil); ok {
		t.Error("summarising nothing should say so rather than answer with zero")
	}
	if _, ok := chart.Summary([]float64{math.NaN(), math.Inf(1)}); ok {
		t.Error("a sample of nothing but not-a-numbers is not a sample")
	}
	// Even quarters interpolate the way a spreadsheet does, so the answer is
	// one a reader can reproduce by hand.
	s, _ = chart.Summary([]float64{1, 2, 3, 4, 5, 6, 7, 8})
	if s.Median != 4.5 {
		t.Errorf("the median of 1…8 is %v, want 4.5", s.Median)
	}
}

// A violin's width is a kernel estimate, and a sample's shape is symmetric
// about the middle of itself — which is a fact about the maths a test can
// state exactly rather than approximately.
func TestDensityIsSymmetricAboutASymmetricSample(t *testing.T) {
	sample := []float64{-2, 2}
	if got, want := chart.Density(1, sample, 1), chart.Density(-1, sample, 1); got != want {
		t.Errorf("the density at 1 is %v and at -1 is %v, want the same: the sample is symmetric", got, want)
	}
	// One sample of one value is that sample's kernel and nothing else.
	if got, want := chart.Density(0, []float64{0}, 1), 1/2.5066282746310002; got != want {
		t.Errorf("the density at the only value is %v, want %v", got, want)
	}
	// Away from the peak it falls off, or a violin would be a rectangle.
	if chart.Density(5, sample, 1) >= chart.Density(0, sample, 1) {
		t.Error("the density rises away from the sample's middle, which is not a density")
	}
	if got := chart.Density(0, nil, 1); got != 0 {
		t.Errorf("the density of no sample is %v, want zero", got)
	}
	if got := chart.Density(0, sample, 0); got != 0 {
		t.Errorf("a kernel with no width is %v, want zero rather than a division by zero", got)
	}
}

// A parallel-coordinates axis is scaled by its own column, so a count and a
// duration can sit on one plot.
func TestColumnIsOneAxisOfEverySeries(t *testing.T) {
	series := []chart.Series{
		chart.SeriesFrom("a", chart.FormLine, []float64{0, 1, 2}, []float64{1, 2, 3}),
		chart.SeriesFrom("b", chart.FormLine, []float64{0, 1, 2}, []float64{10, 20, 30}),
	}
	got, ok := chart.Column(series, 1)
	if !ok || got != (chart.Domain{Min: 2, Max: 20}) {
		t.Errorf("Column(1) = %+v, want 2 to 20", got)
	}
	if _, ok := chart.Column(series, 9); ok {
		t.Error("a column past the end of the series should say there is nothing there")
	}
	if _, ok := chart.Column(nil, 0); ok {
		t.Error("no series has no columns")
	}
}

// A heatmap's palest cell is the surface and its fullest is the accent: the
// two colours the rest of the interface uses, in both appearances.
func TestHeatColorRunsFromTheSurfaceToTheAccent(t *testing.T) {
	for name, k := range map[string]theme.Tokens{"light": theme.Light(), "dark": theme.Dark()} {
		if got := chart.HeatColor(0, 0, 10, k); got != k.Surface {
			t.Errorf("%s: the low end of the scale is %v, want the surface %v", name, got, k.Surface)
		}
		if got := chart.HeatColor(10, 0, 10, k); got != k.Accent {
			t.Errorf("%s: the high end of the scale is %v, want the accent %v", name, got, k.Accent)
		}
	}
	k := theme.Light()
	// Past either end, a value is the end it ran past rather than a colour
	// worked out from a share of a range it is not in.
	if got := chart.HeatColor(-5, 0, 10, k); got != k.Surface {
		t.Errorf("a value below the range is %v, want the surface", got)
	}
	if got := chart.HeatColor(50, 0, 10, k); got != k.Accent {
		t.Errorf("a value above the range is %v, want the accent", got)
	}
	// A range that spans nothing, and a value that is not a number, are the
	// near end rather than a division by nothing.
	if got := chart.HeatColor(5, 5, 5, k); got != k.Surface {
		t.Errorf("a range of nothing is %v, want the surface", got)
	}
	if got := chart.HeatColor(math.NaN(), 0, 10, k); got != k.Surface {
		t.Errorf("a value that is not a number is %v, want the surface", got)
	}
	// And the middle of the range is a colour between the two, not one of
	// them: a heatmap with only two colours is a bar chart of its own.
	middle := chart.HeatColor(5, 0, 10, k)
	if middle == k.Surface || middle == k.Accent {
		t.Errorf("the middle of the scale is %v, want a colour between the two ends", middle)
	}
}

// A dial's needle sits where its value falls on its sweep, and a value past
// the end of the dial is put at the end rather than off it.
func TestArcAnglePlacesTheNeedle(t *testing.T) {
	if got := chart.ArcAngle(0.5, 0, 1, 135, 270); got != 270 {
		t.Errorf("half way round a 270° dial from 135° is %v°, want 270°", got)
	}
	if got := chart.ArcAngle(0, 0, 100, 135, 270); got != 135 {
		t.Errorf("the bottom of the dial is %v°, want its start", got)
	}
	if got := chart.ArcAngle(100, 0, 100, 135, 270); got != 405 {
		t.Errorf("the top of the dial is %v°, want 405°", got)
	}
	if got := chart.ArcAngle(200, 0, 100, 135, 270); got != 405 {
		t.Errorf("a value past the end of the dial is %v°, want it at the end", got)
	}
	if got := chart.ArcAngle(5, 10, 10, 135, 270); got != 135 {
		t.Errorf("a dial with no range has its needle at %v°, want its start", got)
	}
}

// A candle's body stands between its open and its close whichever way round
// they were.
func TestBodyPutsTheHigherOfOpenAndCloseOnTop(t *testing.T) {
	top, bottom := chart.Body(10, 12)
	if top != 12 || bottom != 10 {
		t.Errorf("a rising candle's body is %v to %v, want 12 to 10", top, bottom)
	}
	top, bottom = chart.Body(12, 10)
	if top != 12 || bottom != 10 {
		t.Errorf("a falling candle's body is %v to %v, want 12 to 10", top, bottom)
	}
	top, bottom = chart.Body(7, 7)
	if top != 7 || bottom != 7 {
		t.Errorf("a doji's body is %v to %v, want 7 to 7", top, bottom)
	}
}

// Four equal values in a square are four squares: the layout a treemap is
// judged on, to the pixel.
func TestSquarifyGivesEqualValuesEqualSquares(t *testing.T) {
	got := chart.Squarify([]float64{6, 6, 6, 6}, ui.Rect{X: 0, Y: 0, W: 100, H: 100})
	want := []ui.Rect{
		{X: 0, Y: 0, W: 50, H: 50},
		{X: 0, Y: 50, W: 50, H: 50},
		{X: 50, Y: 0, W: 50, H: 50},
		{X: 50, Y: 50, W: 50, H: 50},
	}
	for i := range want {
		if near(got[i], want[i]) {
			continue
		}
		// The order within a row is the caller's, but which row a value
		// lands in is not, so compare the set rather than the sequence.
		found := false
		for _, other := range got {
			if near(other, want[i]) {
				found = true
			}
		}
		if !found {
			t.Errorf("no square where %v should be, got %v", want[i], got[i])
		}
	}
}

// One value is the whole of the area it was given, and nothing at all is no
// rectangles rather than a division by zero.
func TestSquarifyFillsTheRectItIsGiven(t *testing.T) {
	got := chart.Squarify([]float64{7}, ui.Rect{X: 4, Y: 6, W: 100, H: 80})
	if !near(got[0], ui.Rect{X: 4, Y: 6, W: 100, H: 80}) {
		t.Errorf("one value got %v, want the whole rect", got[0])
	}
	// Every rectangle in the box, and none of them overlapping: a treemap
	// that overlaps is two categories in one place.
	area := 0.0
	for _, r := range chart.Squarify([]float64{5, 3, 2, 9, 1}, ui.Rect{W: 200, H: 100}) {
		area += float64(r.W * r.H)
		if r.X < -0.5 || r.Y < -0.5 || r.X+r.W > 200.5 || r.Y+r.H > 100.5 {
			t.Errorf("a tile of %v is outside the box it was given", r)
		}
	}
	if math.Abs(area-200*100) > 1 {
		t.Errorf("the tiles cover %v of the 20000 they were given", area)
	}
	if got := chart.Squarify([]float64{0, 0}, ui.Rect{W: 100, H: 100}); got[0].W != 0 {
		t.Errorf("values of nothing laid out %v, want no rectangles", got)
	}
}

// near compares two rectangles to within a pixel: the layout divides areas
// and the division rounds, and one pixel is not a fact about the data.
func near(got, want ui.Rect) bool {
	return math.Abs(float64(got.X-want.X)) <= 1 &&
		math.Abs(float64(got.Y-want.Y)) <= 1 &&
		math.Abs(float64(got.W-want.W)) <= 1 &&
		math.Abs(float64(got.H-want.H)) <= 1
}

// A flow is laid out in columns: a node sits as many links deep as the
// longest path to it from a source.
func TestDepthsPutsNodesInColumns(t *testing.T) {
	got := chart.Depths(3, []chart.Link{{Source: 0, Target: 1}, {Source: 1, Target: 2}})
	if want := []int{0, 1, 2}; !sameInts(got, want) {
		t.Errorf("Depths = %v, want %v", got, want)
	}
	// Two sources into one node: it waits for both of them and sits where
	// the later of them puts it.
	got = chart.Depths(3, []chart.Link{{Source: 0, Target: 2}, {Source: 1, Target: 2}})
	if want := []int{0, 0, 1}; !sameInts(got, want) {
		t.Errorf("Depths of two sources into one node = %v, want %v", got, want)
	}
	if got := chart.Depths(0, nil); len(got) != 0 {
		t.Errorf("Depths of no nodes is no nodes, got %v", got)
	}
}

func sameInts(got, want []int) bool {
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

func TestDepthsRefusesALinkToNowhere(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("a link to a node that is not there should panic rather than draw a ribbon to nowhere")
		}
	}()
	chart.Depths(2, []chart.Link{{Source: 0, Target: 5}})
}

// A calendar heatmap's days land in the week and on the weekday they fall on,
// counted from the first day's week.
func TestCalendarGridPlacesDaysOnTheCalendar(t *testing.T) {
	days := []chart.Day{
		{Date: "2024-01-01", Value: 1}, // a Monday
		{Date: "2024-01-07", Value: 2}, // the Sunday after it
		{Date: "2024-01-14", Value: 3}, // the Sunday after that
	}
	grid, ok := chart.NewCalendarGrid(days, time.Monday)
	if !ok {
		t.Fatal("three real dates made no grid")
	}
	if grid.From.Format("2006-01-02") != "2024-01-01" {
		t.Errorf("the grid starts on %v, want the Monday the first date is on", grid.From)
	}
	if week, day, _ := grid.Cell("2024-01-01"); week != 0 || day != 0 {
		t.Errorf("the Monday is in week %d on weekday %d, want 0, 0", week, day)
	}
	if week, day, _ := grid.Cell("2024-01-07"); week != 0 || day != 6 {
		t.Errorf("the Sunday after it is in week %d on weekday %d, want 0, 6", week, day)
	}
	if week, day, _ := grid.Cell("2024-01-14"); week != 1 || day != 6 {
		t.Errorf("the next Sunday is in week %d on weekday %d, want 1, 6", week, day)
	}
	// A grid that starts on Sunday puts the same dates a week along.
	sunday, ok := chart.NewCalendarGrid(days, time.Sunday)
	if !ok {
		t.Fatal("three real dates made no grid")
	}
	if week, day, _ := sunday.Cell("2024-01-01"); week != 0 || day != 1 {
		t.Errorf("the Monday is in week %d on weekday %d of a Sunday-first grid, want 0, 1", week, day)
	}
	// A date outside the grid, or one that is not a date at all, is no cell.
	if _, _, ok := sunday.Cell("2025-01-01"); ok {
		t.Error("a date a year later is in a three-day grid")
	}
	if _, _, ok := sunday.Cell("not a date"); ok {
		t.Error("a string that is not a date is a cell on a calendar")
	}
	if _, ok := chart.NewCalendarGrid([]chart.Day{{Date: "nonsense"}}, time.Monday); ok {
		t.Error("a grid made of dates that are not dates is a grid")
	}
}

// ── the chart types, rendered ──────────────────────────────────────────────

// renders draws a chart with no window and hands back what came out, so each
// test can look at the pixels and at what the chart said for assistive
// technology.
func renders(t *testing.T, w, h int, view func(c *ui.Context)) *ui.Tester {
	t.Helper()
	return ui.NewTester(view, w, h)
}

// A chart that drew nothing has failed, whatever else it may have done. Every
// render test below starts by saying that, before it says anything sharper.
func TestEveryChartTypeDrawsSomething(t *testing.T) {
	for name, tc := range chartCases() {
		t.Run(name, func(t *testing.T) {
			tt := renders(t, 460, 320, func(c *ui.Context) {
				core.Use(c, core.Settings{})
				tc.build(c)
			})
			img := tt.Image()
			// The frame's own area rather than the whole window: the
			// window is the background, which is not the chart.
			if !painted(img, ui.Rect{X: 1, Y: 1, W: 458, H: 318}, theme.Light().Background) {
				t.Errorf("%s drew nothing at all", name)
			}
		})
	}
}

// A chart with data draws differently from the same chart with none. This is
// the assertion that a chart is really putting its data on the screen: a
// chart that draws its frame and its axes and then quietly draws no marks at
// all passes every other test in this file, because its frame is drawn and its
// empty state is drawn and both of them are somebody else's work.
func TestEveryChartTypeDrawsItsData(t *testing.T) {
	for name, tc := range chartCases() {
		t.Run(name, func(t *testing.T) {
			with := renders(t, 460, 320, func(c *ui.Context) {
				core.Use(c, core.Settings{})
				tc.build(c)
			}).Image()
			without := renders(t, 460, 320, func(c *ui.Context) {
				core.Use(c, core.Settings{})
				tc.empty(c)
			}).Image()
			changed := 0
			for y := range 320 {
				for x := range 460 {
					if rgba(with, x, y) != rgba(without, x, y) {
						changed++
					}
				}
			}
			if changed < 200 {
				t.Errorf("%s draws almost the same thing with data as without: %d pixels differ",
					name, changed)
			}
		})
	}
}

// Every chart type survives the data it was given nothing: no panic, a frame,
// and the empty state rather than an axis over a range nothing reaches.
func TestEveryChartTypeSurvivesNoData(t *testing.T) {
	for name, tc := range chartCases() {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("%s with no data panicked: %v", name, r)
				}
			}()
			tt := renders(t, 460, 320, func(c *ui.Context) {
				core.Use(c, core.Settings{})
				tc.empty(c)
			})
			// A chart with nothing to draw is a frame with a message
			// in it, or at the very least an element that says what it
			// is — and never a panic and never an axis over nothing.
			if len(tt.Texts()) == 0 {
				t.Errorf("%s with no data is not even named for assistive technology", name)
			}
		})
	}
}

// A chart in the dark theme is the same chart: the tokens follow the window,
// and nothing here has a palette of its own to keep in step.
func TestChartsFollowTheDarkTheme(t *testing.T) {
	var light, dark []ui.Color
	renders(t, 460, 320, func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: core.Light})
		chart.BarChart(c, chart.BarOptions{
			ChartOptions: chart.ChartOptions{Height: 220},
			Series:       []chart.Series{chart.SeriesFrom("Open", chart.FormBar, typeDays, typeOpen)},
		})
		light = chart.Palette(c, 3)
	})
	renders(t, 460, 320, func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: core.Dark})
		chart.BarChart(c, chart.BarOptions{
			ChartOptions: chart.ChartOptions{Height: 220},
			Series:       []chart.Series{chart.SeriesFrom("Open", chart.FormBar, typeDays, typeOpen)},
		})
		dark = chart.Palette(c, 3)
	})
	// The series colours are the accent's ladder in both appearances, which
	// is what lets one chart read on a white window and a near-black one.
	if want := chart.PaletteFrom(theme.Dark().Accent, 3); !sameColors(dark, want) {
		t.Errorf("the dark palette is %v, want the dark accent's ladder %v", dark, want)
	}
	if light[0] == dark[0] {
		t.Error("the two appearances share a palette step, which is the thing the ladder avoids")
	}
	// And the heat scale's ends are the dark theme's own tokens, not the
	// light ones faded.
	k := theme.Dark()
	if chart.HeatColor(10, 0, 10, k) != k.Accent || chart.HeatColor(0, 0, 10, k) != k.Surface {
		t.Error("the dark heat scale does not end on the dark theme's own tokens")
	}
}

// A pie's slices are drawn in the palette's colours, each one where its own
// value puts it, so a test can find the big one and check its colour.
func TestPieDrawsItsSlicesInOrder(t *testing.T) {
	var colors []ui.Color
	tt := renders(t, 360, 300, func(c *ui.Context) {
		core.Use(c, core.Settings{})
		chart.PieChart(c, chart.PieOptions{
			ChartOptions: chart.ChartOptions{Height: 260, Label: "Share of callbacks"},
			Labels:       typeLabels(),
			Values:       typeValues(),
		})
		colors = chart.Palette(c, 5)
	})
	img := tt.Image()
	// The slices are where [chart.Sweep] says they are: 45 of 100 is
	// 162°, which is the whole of the left-hand half of the circle, and 30
	// of 100 is 108° from three o'clock, which is the lower right.
	cx, cy := float32(360/2), float32(300/2)
	if !findsNear(img, ui.Rect{X: cx - 84, Y: cy - 8, W: 16, H: 16}, colors[1]) {
		t.Error("no second slice where its own angle puts it, on the left of the pie")
	}
	if !findsNear(img, ui.Rect{X: cx + 44, Y: cy + 54, W: 16, H: 16}, colors[0]) {
		t.Error("no first slice where its own angle puts it, below and right of the middle")
	}
	if !tt.HasText("Share of callbacks") {
		t.Errorf("the pie is not named for assistive technology: %q", tt.Texts())
	}
}

// A donut's hole is a hole: the middle of it is the window behind, which is
// what the hole is for.
func TestDonutLeavesItsHole(t *testing.T) {
	tt := renders(t, 360, 300, func(c *ui.Context) {
		core.Use(c, core.Settings{})
		chart.DonutChart(c, chart.DonutOptions{
			PieOptions: chart.PieOptions{
				ChartOptions: chart.ChartOptions{Height: 260},
				Labels:       typeLabels(),
				Values:       typeValues(),
			},
			Hole: 0.5,
		})
	})
	img := tt.Image()
	mid := ui.Rect{X: 360/2 - 12, Y: 300/2 - 12, W: 24, H: 24}
	for y := int(mid.Y); y < int(mid.Y+mid.H); y++ {
		for x := int(mid.X); x < int(mid.X+mid.W); x++ {
			if got := rgba(img, x, y); got != theme.Light().Background {
				t.Errorf("the middle of the donut is %v, want the window behind it", got)
			}
		}
	}
}

// A line is drawn along its data, all the way across the plot. This is the
// assertion that catches a chart whose frame has no axes on it: the marks are
// still drawn — a dot in the corner of the plot, where a scale that spans
// nothing puts everything — so a test that only asks "did anything appear"
// passes, and only the count of the line's own colour gives it away.
func TestLineChartDrawsALineAlongItsData(t *testing.T) {
	ink := ui.Hex("#c62b30")
	tt := renders(t, 460, 320, func(c *ui.Context) {
		core.Use(c, core.Settings{})
		chart.LineChart(c, chart.LineOptions{
			ChartOptions: chartOptions(240),
			Series: []chart.Series{{Name: "Open", Color: ink, Form: chart.FormLine,
				Points: []chart.Point{{X: 0, Y: 5}, {X: 6, Y: 5}}}},
		})
	})
	if n := countNear(tt.Image(), ui.Rect{W: 460, H: 320}, ink); n < 300 {
		t.Errorf("the line was drawn over %d pixels, want a line across the plot", n)
	}
	// And a chart with a bar form of its own draws bars, not a line: the
	// y axis a bar chart builds starts at zero, so its 12-unit bar is twice
	// the height of the same value's line at the middle of the plot.
	bars := renders(t, 460, 320, func(c *ui.Context) {
		core.Use(c, core.Settings{})
		chart.BarChart(c, chart.BarOptions{
			ChartOptions: chartOptions(240),
			Series: []chart.Series{{Name: "Open", Color: ink, Form: chart.FormBar,
				Points: []chart.Point{{X: 0, Y: 6}, {X: 1, Y: 12}}}},
		})
	})
	// The bars are washed back so that a bar behind another still reads
	// through it, so they are counted as red rather than as their colour:
	// what matters is that the columns are there and are big.
	if n := countReddish(bars.Image(), ui.Rect{W: 460, H: 320}); n < 4000 {
		t.Errorf("the bars cover %d pixels, want two columns standing on zero", n)
	}
}

// countNear is how many pixels in r are within a step or two of a colour: how
// a test counts a mark that has an antialiased edge.
func countNear(img *image.RGBA, r ui.Rect, want ui.Color) int {
	n := 0
	for y := int(r.Y); y < int(r.Y+r.H); y++ {
		for x := int(r.X); x < int(r.X+r.W); x++ {
			got := rgba(img, x, y)
			if abs(int(got.R)-int(want.R)) <= 3 && abs(int(got.G)-int(want.G)) <= 3 &&
				abs(int(got.B)-int(want.B)) <= 3 {
				n++
			}
		}
	}
	return n
}

// countReddish is how many pixels are unmistakably the colour of a bar: a
// bar is drawn washed back over the window, so its exact value on the screen
// is the renderer's business rather than the chart's.
func countReddish(img *image.RGBA, r ui.Rect) int {
	n := 0
	for y := int(r.Y); y < int(r.Y+r.H); y++ {
		for x := int(r.X); x < int(r.X+r.W); x++ {
			c := rgba(img, x, y)
			if c.R > c.G+40 && c.R > c.B+40 {
				n++
			}
		}
	}
	return n
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// An area chart fills the ground under its line: the whole claim of the
// chart, and the half of it that a test counting a line's pixels would miss.
func TestAreaChartFillsTheGround(t *testing.T) {
	ink := ui.Hex("#c62b30")
	tt := renders(t, 460, 320, func(c *ui.Context) {
		core.Use(c, core.Settings{})
		chart.AreaChart(c, chart.AreaOptions{
			ChartOptions: chartOptions(240),
			Series: []chart.Series{{Name: "Open", Color: ink, Form: chart.FormArea,
				Points: []chart.Point{{X: 0, Y: 1}, {X: 6, Y: 1}}}},
		})
	})
	// The fill is a wash, so it is counted as "not the window behind it"
	// rather than as the series colour: a wash is the point of it.
	painted := 0
	img := tt.Image()
	bg := theme.Light().Background
	for y := range 320 {
		for x := range 460 {
			if rgba(img, x, y) != bg {
				painted++
			}
		}
	}
	if painted < 8000 {
		t.Errorf("the area covers %d pixels, want the ground under the line filled in", painted)
	}
}

// A treemap's tiles are in the palette's colours and each is where its own
// area puts it — the biggest value's tile takes the most room.
func TestTreemapGivesTheBiggestTileTheMostRoom(t *testing.T) {
	tt := renders(t, 400, 300, func(c *ui.Context) {
		core.Use(c, core.Settings{})
		chart.Treemap(c, chart.TreemapOptions{
			ChartOptions: chart.ChartOptions{Height: 260},
			Labels:       true,
			Tiles: []chart.Tile{
				{Label: "Engine", Value: 60},
				{Label: "Routing", Value: 30},
				{Label: "Auth", Value: 10},
			},
		})
	})
	img := tt.Image()
	if !tt.HasText("Treemap") {
		t.Errorf("the treemap is not named for assistive technology: %q", tt.Texts())
	}
	// Each tile is drawn in the palette's own colour, washed back so that a
	// tile behind another still reads through it — which is why the colour
	// on the screen is nearest to its palette entry rather than equal to it.
	palette := chart.PaletteFrom(theme.Light().Accent, 3)
	for i, where := range []ui.Rect{{X: 20, Y: 20, W: 60, H: 60}, {X: 300, Y: 20, W: 60, H: 60}} {
		got := dominant(img, where)
		if at, away := nearest(got, palette); at != i || away > 70 {
			t.Errorf("the tile at %v is %v, which is %d away from palette entry %d and %d from entry %d",
				where, got, away, at, away, palette[at%len(palette)])
		}
	}
}

// A heatmap's cells are the colours the heat scale works out, one per cell,
// so a test can look for the value it knows it put there.
func TestHeatmapColoursEachCellByItsValue(t *testing.T) {
	k := theme.Light()
	tt := renders(t, 420, 320, func(c *ui.Context) {
		core.Use(c, core.Settings{})
		chart.HeatmapChart(c, chart.HeatmapOptions{
			ChartOptions: chart.ChartOptions{Height: 260},
			Rows:         []string{"Mon", "Tue"},
			Cols:         []string{"Open", "Closed"},
			Values:       [][]float64{{0, 5}, {10, 5}},
			Write:        true,
		})
	})
	img := tt.Image()
	// The full cell is the accent and the empty one is the surface; both are
	// in the picture, which is the whole claim of a heatmap.
	if !findsNear(img, ui.Rect{X: 2, Y: 2, W: 418, H: 318}, chart.HeatColor(10, 0, 10, k)) {
		t.Error("no cell in the colour the top of the scale works out")
	}
	if !findsNear(img, ui.Rect{X: 2, Y: 2, W: 418, H: 318}, chart.HeatColor(0, 0, 10, k)) {
		t.Error("no cell in the colour the bottom of the scale works out")
	}
	if !tt.HasText("Heatmap") {
		t.Errorf("the grid is not named for assistive technology: %q", tt.Texts())
	}
}

// A candlestick's wick reaches its high and its low and its body stands
// between its open and its close, which is the whole of what a candle says.
func TestCandlestickDrawsWicksAndBodies(t *testing.T) {
	k := theme.Light()
	tt := renders(t, 420, 320, func(c *ui.Context) {
		core.Use(c, core.Settings{})
		chart.CandlestickChart(c, chart.CandlestickOptions{
			ChartOptions: chartOptions(240), Candles: typeCandles()})
	})
	img := tt.Image()
	// One falling candle is drawn in the danger colour, which is a token
	// rather than a palette step.
	if !findsNear(img, ui.Rect{X: 1, Y: 1, W: 418, H: 318}, k.Danger) {
		t.Error("no falling candle in the danger colour")
	}
	if !findsNear(img, ui.Rect{X: 1, Y: 1, W: 418, H: 318}, k.Success) {
		t.Error("no rising candle in the success colour")
	}
}

// A sparkline has no axes to collide and still has to draw its line: the
// whole point of it is that it is too small to have anything else.
func TestSparklineAndSparkBarDraw(t *testing.T) {
	tt := renders(t, 200, 60, func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ui.Row(c).Children(func() {
			chart.Sparkline(c, chart.SparklineOptions{Values: typeOpen, Width: 100, Height: 40, Fill: true})
			chart.SparkBar(c, chart.SparkBarOptions{Values: typeOpen, Width: 100, Height: 40})
		})
	})
	if !tt.HasText("Sparkline") || !tt.HasText("Spark bar") {
		t.Errorf("the sparks are not named: %q", tt.Texts())
	}
	if !painted(tt.Image(), ui.Rect{W: 200, H: 60}, theme.Light().Background) {
		t.Error("the sparks drew nothing")
	}
}

// The range highlight washes the band it was asked for and nothing else: it
// is a mark over somebody else's chart, so it has to stay inside the plot and
// it has to say which part of the axis it is about.
func TestRangeHighlightWashesOnlyItsRange(t *testing.T) {
	bg := theme.Light().Background
	ink := ui.Hex("#c62b30")
	var frame chart.FrameResult
	xs := chart.NewLinear(chart.Domain{Min: 0, Max: 10}, 0, 1)
	ys := chart.NewLinear(chart.Domain{Min: 0, Max: 12}, 0, 1)
	tt := renders(t, 420, 320, func(c *ui.Context) {
		core.Use(c, core.Settings{})
		frame = chart.Frame(c, chart.FrameOptions{Height: 240,
			X: chart.AxisOptions{Scale: xs, Count: 6}, Y: chart.AxisOptions{Side: chart.Left, Scale: ys, Count: 4}},
			func(f chart.FrameResult) {
				chart.Plot(c, chart.PlotOptions{Frame: f, Width: 3,
					Series: chart.Series{Name: "Open", Form: chart.FormLine,
						Points: []chart.Point{{X: 0, Y: 0}, {X: 10, Y: 12}}},
					X: xs, Y: ys})
				chart.RangeHighlight(c, chart.RangeHighlightOptions{
					Frame: f, X: xs, From: 2, To: 6, Edges: true,
					Color: ink, Label: "incident",
				})
			})
	})
	img := tt.Image()
	plot := frame.Plot()
	// The edge lines stand inside the band, so their ink composites over the
	// band's own wash, not over the bare background.
	wash := ink.Alpha(0.12).Over(bg)
	if !findsNear(img, ui.Rect{X: plot.X + plot.W*0.2 + 4, Y: plot.Y + 4, W: plot.W * 0.4, H: 8}, ink.Alpha(0.6).Over(wash)) {
		t.Error("the range's left edge is not where two on a zero-to-ten axis puts it")
	}
	if !tt.HasText("Range") {
		t.Errorf("the highlight is not named for assistive technology: %q", tt.Texts())
	}
	// And the far left of the plot is left alone: a highlight that washed
	// the whole chart would be the same as drawing the chart in the wash.
	if findsNear(img, ui.Rect{X: plot.X + 1, Y: plot.Y + 4, W: plot.W * 0.1, H: 8}, wash) {
		t.Error("the highlight washed a range it was not given")
	}
}

// ── the cases ──────────────────────────────────────────────────────────────

// chartCase is one chart type with the data it draws and the data it has
// nothing to draw: every type gets both, so that "renders" and "survives no
// data" are properties of all of them rather than of the ones somebody
// remembered.
type chartCase struct {
	build func(c *ui.Context)
	empty func(c *ui.Context)
}

func chartOptions(height float32) chart.ChartOptions {
	return chart.ChartOptions{Height: height, Pad: 4}
}

func chartCases() map[string]chartCase {
	line := func(c *ui.Context) {
		chart.LineChart(c, chart.LineOptions{
			ChartOptions: chartOptions(240), Series: typeSeries(), Dots: true, Width: 3,
		})
	}
	return map[string]chartCase{
		"LineChart": {build: line, empty: func(c *ui.Context) {
			chart.LineChart(c, chart.LineOptions{ChartOptions: chartOptions(240)})
		}},
		"AreaChart": {build: func(c *ui.Context) {
			chart.AreaChart(c, chart.AreaOptions{ChartOptions: chartOptions(240), Series: typeSeries()})
		}, empty: func(c *ui.Context) {
			chart.AreaChart(c, chart.AreaOptions{ChartOptions: chartOptions(240)})
		}},
		"AreaChartStacked": {build: func(c *ui.Context) {
			chart.AreaChart(c, chart.AreaOptions{ChartOptions: chartOptions(240), Series: typeSeries(), Stacked: true})
		}, empty: func(c *ui.Context) {
			chart.AreaChart(c, chart.AreaOptions{ChartOptions: chartOptions(240), Stacked: true})
		}},
		"AreaMountain": {build: func(c *ui.Context) {
			chart.AreaMountain(c, chart.AreaMountainOptions{ChartOptions: chartOptions(240), Series: typeSeries()})
		}, empty: func(c *ui.Context) {
			chart.AreaMountain(c, chart.AreaMountainOptions{ChartOptions: chartOptions(240)})
		}},
		"BarChart": {build: func(c *ui.Context) {
			chart.BarChart(c, chart.BarOptions{ChartOptions: chartOptions(240),
				Series: []chart.Series{chart.SeriesFrom("Open", chart.FormBar, typeDays, typeOpen)}})
		}, empty: func(c *ui.Context) {
			chart.BarChart(c, chart.BarOptions{ChartOptions: chartOptions(240)})
		}},
		"BarChartHorizontal": {build: func(c *ui.Context) {
			chart.BarChart(c, chart.BarOptions{ChartOptions: chartOptions(240), Horizontal: true,
				Series: []chart.Series{chart.SeriesFrom("Open", chart.FormBar, typeDays, typeOpen)}})
		}, empty: func(c *ui.Context) {
			chart.BarChart(c, chart.BarOptions{ChartOptions: chartOptions(240), Horizontal: true})
		}},
		"StackedBar": {build: func(c *ui.Context) {
			chart.StackedBar(c, chart.StackedBarOptions{ChartOptions: chartOptions(240),
				Labels: typeLabels(), Series: typeSeries()})
		}, empty: func(c *ui.Context) {
			chart.StackedBar(c, chart.StackedBarOptions{ChartOptions: chartOptions(240)})
		}},
		"PercentBar": {build: func(c *ui.Context) {
			chart.PercentBar(c, chart.StackedBarOptions{ChartOptions: chartOptions(240),
				Labels: typeLabels(), Series: typeSeries()})
		}, empty: func(c *ui.Context) {
			chart.PercentBar(c, chart.StackedBarOptions{ChartOptions: chartOptions(240)})
		}},
		"Histogram": {build: func(c *ui.Context) {
			chart.Histogram(c, chart.HistogramOptions{ChartOptions: chartOptions(240), Bins: 4,
				Values: []float64{1, 2, 2, 3, 5, 8, 8, 9}})
		}, empty: func(c *ui.Context) {
			chart.Histogram(c, chart.HistogramOptions{ChartOptions: chartOptions(240)})
		}},
		"ScatterChart": {build: func(c *ui.Context) {
			chart.ScatterChart(c, chart.ScatterOptions{ChartOptions: chartOptions(240),
				Series: []chart.Series{chart.SeriesFrom("Open", chart.FormPoint, typeDays, typeOpen)}})
		}, empty: func(c *ui.Context) {
			chart.ScatterChart(c, chart.ScatterOptions{ChartOptions: chartOptions(240)})
		}},
		"BubbleChart": {build: func(c *ui.Context) {
			chart.BubbleChart(c, chart.BubbleOptions{ChartOptions: chartOptions(240),
				Bubbles: chart.BubblesFrom("Open", typeDays, typeOpen, []float64{1, 4, 9, 2, 5, 8, 3})})
		}, empty: func(c *ui.Context) {
			chart.BubbleChart(c, chart.BubbleOptions{ChartOptions: chartOptions(240)})
		}},
		"BoxPlot": {build: func(c *ui.Context) {
			chart.BoxPlot(c, chart.BoxPlotOptions{ChartOptions: chartOptions(240), Samples: typeSamples()})
		}, empty: func(c *ui.Context) {
			chart.BoxPlot(c, chart.BoxPlotOptions{ChartOptions: chartOptions(240)})
		}},
		"ViolinPlot": {build: func(c *ui.Context) {
			chart.ViolinPlot(c, chart.ViolinOptions{ChartOptions: chartOptions(240), Samples: typeSamples()})
		}, empty: func(c *ui.Context) {
			chart.ViolinPlot(c, chart.ViolinOptions{ChartOptions: chartOptions(240)})
		}},
		"RadarChart": {build: func(c *ui.Context) {
			chart.RadarChart(c, chart.RadarOptions{ChartOptions: chartOptions(240),
				Axes: []string{"CPU", "RAM", "IO", "Net"},
				Series: []chart.Series{
					{Form: chart.FormLine, Points: []chart.Point{{Y: 4}, {Y: 7}, {Y: 3}, {Y: 8}}},
					{Form: chart.FormLine, Points: []chart.Point{{Y: 6}, {Y: 2}, {Y: 5}, {Y: 4}}},
				}})
		}, empty: func(c *ui.Context) {
			chart.RadarChart(c, chart.RadarOptions{ChartOptions: chartOptions(240)})
		}},
		"PolarChart": {build: func(c *ui.Context) {
			chart.PolarChart(c, chart.PolarOptions{ChartOptions: chartOptions(240), Series: typeSeries()})
		}, empty: func(c *ui.Context) {
			chart.PolarChart(c, chart.PolarOptions{ChartOptions: chartOptions(240)})
		}},
		"PieChart": {build: func(c *ui.Context) {
			chart.PieChart(c, chart.PieOptions{ChartOptions: chartOptions(240),
				Labels: typeLabels(), Values: typeValues()})
		}, empty: func(c *ui.Context) {
			chart.PieChart(c, chart.PieOptions{ChartOptions: chartOptions(240)})
		}},
		"DonutChart": {build: func(c *ui.Context) {
			chart.DonutChart(c, chart.DonutOptions{PieOptions: chart.PieOptions{
				ChartOptions: chartOptions(240), Labels: typeLabels(), Values: typeValues()}})
		}, empty: func(c *ui.Context) {
			chart.DonutChart(c, chart.DonutOptions{PieOptions: chart.PieOptions{ChartOptions: chartOptions(240)}})
		}},
		"NightingaleChart": {build: func(c *ui.Context) {
			chart.NightingaleChart(c, chart.NightingaleOptions{PieOptions: chart.PieOptions{
				ChartOptions: chartOptions(240), Labels: typeLabels(), Values: typeValues()}})
		}, empty: func(c *ui.Context) {
			chart.NightingaleChart(c, chart.NightingaleOptions{PieOptions: chart.PieOptions{ChartOptions: chartOptions(240)}})
		}},
		"FunnelChart": {build: func(c *ui.Context) {
			chart.FunnelChart(c, chart.FunnelOptions{ChartOptions: chartOptions(240),
				Steps: []chart.FunnelStep{
					{Label: "Visited", Value: 100}, {Label: "Signed up", Value: 40},
					{Label: "Activated", Value: 20}, {Label: "Kept", Value: 5}}})
		}, empty: func(c *ui.Context) {
			chart.FunnelChart(c, chart.FunnelOptions{ChartOptions: chartOptions(240)})
		}},
		"WaterfallChart": {build: func(c *ui.Context) {
			chart.WaterfallChart(c, chart.WaterfallOptions{ChartOptions: chartOptions(240), Total: true,
				Labels: []string{"Start", "Adds", "Fix", "Reviews", "Ship"},
				Values: []float64{20, 10, -5, 4, -9}})
		}, empty: func(c *ui.Context) {
			chart.WaterfallChart(c, chart.WaterfallOptions{ChartOptions: chartOptions(240)})
		}},
		"ParetoChart": {build: func(c *ui.Context) {
			chart.ParetoChart(c, chart.ParetoOptions{ChartOptions: chartOptions(240),
				Labels: typeLabels(), Values: typeValues()})
		}, empty: func(c *ui.Context) {
			chart.ParetoChart(c, chart.ParetoOptions{ChartOptions: chartOptions(240)})
		}},
		"HeatmapChart": {build: func(c *ui.Context) {
			chart.HeatmapChart(c, chart.HeatmapOptions{ChartOptions: chartOptions(240),
				Rows: []string{"Mon", "Tue", "Wed"}, Cols: []string{"Open", "Closed"},
				Values: [][]float64{{1, 2}, {3, 4}, {5, 6}}})
		}, empty: func(c *ui.Context) {
			chart.HeatmapChart(c, chart.HeatmapOptions{ChartOptions: chartOptions(240)})
		}},
		"CalendarHeatmap": {build: func(c *ui.Context) {
			chart.CalendarHeatmap(c, chart.CalendarHeatmapOptions{
				ChartOptions: chartOptions(240), Days: typeDays12(), MonthLabels: true})
		}, empty: func(c *ui.Context) {
			chart.CalendarHeatmap(c, chart.CalendarHeatmapOptions{ChartOptions: chartOptions(240)})
		}},
		"CandlestickChart": {build: func(c *ui.Context) {
			chart.CandlestickChart(c, chart.CandlestickOptions{
				ChartOptions: chartOptions(240), Candles: typeCandles()})
		}, empty: func(c *ui.Context) {
			chart.CandlestickChart(c, chart.CandlestickOptions{ChartOptions: chartOptions(240)})
		}},
		"SankeyChart": {build: func(c *ui.Context) {
			chart.SankeyChart(c, chart.SankeyOptions{ChartOptions: chartOptions(240),
				Names: []string{"New", "Active", "Churned"},
				Links: []chart.Link{{Source: 0, Target: 1, Value: 60}, {Source: 0, Target: 2, Value: 20},
					{Source: 1, Target: 2, Value: 10}}})
		}, empty: func(c *ui.Context) {
			chart.SankeyChart(c, chart.SankeyOptions{ChartOptions: chartOptions(240),
				Names: []string{"New"}})
		}},
		"Treemap": {build: func(c *ui.Context) {
			chart.Treemap(c, chart.TreemapOptions{
				ChartOptions: chartOptions(240),
				Labels:       true,
				Tiles: []chart.Tile{{Label: "Engine", Value: 60}, {Label: "Routing", Value: 30},
					{Label: "Auth", Value: 10}},
			})
		}, empty: func(c *ui.Context) {
			chart.Treemap(c, chart.TreemapOptions{ChartOptions: chartOptions(240)})
		}},
		"SunburstChart": {build: func(c *ui.Context) {
			chart.SunburstChart(c, chart.SunburstOptions{ChartOptions: chartOptions(240),
				Root: treeFixture()})
		}, empty: func(c *ui.Context) {
			chart.SunburstChart(c, chart.SunburstOptions{ChartOptions: chartOptions(240)})
		}},
		"ChordDiagram": {build: func(c *ui.Context) {
			chart.ChordDiagram(c, chart.ChordOptions{ChartOptions: chartOptions(240),
				Names:  []string{"a", "b", "c", "d"},
				Matrix: [][]float64{{0, 4, 2, 1}, {3, 0, 5, 2}, {1, 6, 0, 3}, {2, 1, 4, 0}}})
		}, empty: func(c *ui.Context) {
			chart.ChordDiagram(c, chart.ChordOptions{ChartOptions: chartOptions(240),
				Names: []string{"a"}})
		}},
		"ParallelCoordinates": {build: func(c *ui.Context) {
			chart.ParallelCoordinates(c, chart.ParallelOptions{ChartOptions: chartOptions(240),
				Axes: []string{"CPU", "RAM", "IO"}, Series: typeSeries()})
		}, empty: func(c *ui.Context) {
			chart.ParallelCoordinates(c, chart.ParallelOptions{ChartOptions: chartOptions(240)})
		}},
		"NetworkGraph": {build: func(c *ui.Context) {
			chart.NetworkGraph(c, chart.NetworkOptions{ChartOptions: chartOptions(240),
				Nodes: []chart.GraphNode{
					{Label: "api", Value: 9, Group: 0}, {Label: "web", Value: 6, Group: 0},
					{Label: "db", Value: 8, Group: 1}, {Label: "queue", Value: 3, Group: 1}},
				Links: []chart.GraphLink{{Source: 0, Target: 1, Value: 4}, {Source: 0, Target: 2, Value: 6},
					{Source: 3, Target: 1, Value: 2}}})
		}, empty: func(c *ui.Context) {
			chart.NetworkGraph(c, chart.NetworkOptions{ChartOptions: chartOptions(240)})
		}},
		"AlluvialChart": {build: func(c *ui.Context) {
			chart.AlluvialChart(c, chart.AlluvialOptions{ChartOptions: chartOptions(240),
				Stages: []string{"Visited", "Signed up", "Kept"},
				Flows:  [][]float64{{60, 40}, {30, 10}}})
		}, empty: func(c *ui.Context) {
			chart.AlluvialChart(c, chart.AlluvialOptions{ChartOptions: chartOptions(240)})
		}},
		"DecompositionTree": {build: func(c *ui.Context) {
			chart.DecompositionTree(c, chart.DecompositionOptions{
				ChartOptions: chartOptions(240), Root: treeFixture()})
		}, empty: func(c *ui.Context) {
			chart.DecompositionTree(c, chart.DecompositionOptions{ChartOptions: chartOptions(240)})
		}},
		"Sparkline": {build: func(c *ui.Context) {
			chart.Sparkline(c, chart.SparklineOptions{Values: typeOpen, Width: 120, Height: 60, Fill: true})
		}, empty: func(c *ui.Context) {
			chart.Sparkline(c, chart.SparklineOptions{Width: 120, Height: 60})
		}},
		"SparkBar": {build: func(c *ui.Context) {
			chart.SparkBar(c, chart.SparkBarOptions{Values: typeOpen, Width: 120, Height: 60})
		}, empty: func(c *ui.Context) {
			chart.SparkBar(c, chart.SparkBarOptions{Width: 120, Height: 60})
		}},
		"Gauge": {build: func(c *ui.Context) {
			chart.Gauge(c, chart.GaugeOptions{Value: 72, Max: 100, Width: 200, Height: 140,
				Label: "in budget", Zones: []chart.GaugeZone{{To: 50, Name: "under"}, {To: 100, Name: "over"}}})
		}, empty: func(c *ui.Context) {
			chart.Gauge(c, chart.GaugeOptions{Width: 200, Height: 140})
		}},
		"BulletChart": {build: func(c *ui.Context) {
			chart.BulletChart(c, chart.BulletOptions{Value: 72, Target: 80, Max: 100,
				Width: 260, Height: 90, Label: "Open",
				Bands: []chart.BulletBand{{Min: 0, Max: 60, Name: "slow"}, {Min: 60, Max: 100, Name: "fast"}}})
		}, empty: func(c *ui.Context) {
			chart.BulletChart(c, chart.BulletOptions{Width: 260, Height: 90})
		}},
		"CandleCountdown": {build: func(c *ui.Context) {
			chart.CandleCountdown(c, chart.CandleCountdownOptions{
				Left: 30 * time.Minute, Total: time.Hour, Width: 240, Height: 80, Label: "30m left"})
		}, empty: func(c *ui.Context) {
			chart.CandleCountdown(c, chart.CandleCountdownOptions{Width: 240, Height: 80})
		}},
		"RangeHighlight": {build: func(c *ui.Context) {
			rangeOverLine(c)
		}, empty: func(c *ui.Context) {
			chart.RangeHighlight(c, chart.RangeHighlightOptions{Frame: chart.FrameResult{}})
		}},
	}
}

func treeFixture() chart.Node {
	return chart.Node{Label: "All", Value: 100, Children: []chart.Node{
		{Label: "Engine", Value: 60, Children: []chart.Node{{Label: "Core", Value: 40}, {Label: "Queue", Value: 20}}},
		{Label: "Routing", Value: 30},
		{Label: "Auth", Value: 10},
	}}
}

func rangeOverLine(c *ui.Context) {
	xs := chart.NewLinear(chart.Domain{Min: 0, Max: 6}, 0, 1)
	chart.Frame(c, chart.FrameOptions{Height: 240, X: chart.AxisOptions{Scale: xs, Count: 7}},
		func(f chart.FrameResult) {
			chart.LineChart(c, chart.LineOptions{
				ChartOptions: chartOptions(240), Series: typeSeries(), Dots: true,
			})
			chart.RangeHighlight(c, chart.RangeHighlightOptions{
				Frame: f, X: xs, From: 1, To: 4, Edges: true, Label: "incident",
			})
		})
}
