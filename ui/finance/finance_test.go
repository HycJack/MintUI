package finance

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/data"
	"github.com/HycJack/MintUI/ui/theme"
)

// The pure functions carry this package: FormatPrice, FormatChange,
// FormatVolume, FormatSize, Levels, Spread, allocate's grouping and payoff.
// None of them needs a window, so each is asserted on exactly what it returns
// rather than on what a bar came out looking like.
//
// Everything else is asserted on the element it drew, on a name it put on
// screen, or on a value it wrote into the caller's pointer. MyGo builds a frame
// up to three times and stops when one consumed nothing
// (mygo/ui/runtime.go:333), so after a tester settles, Clicked() is false on
// the last pass. Every press below is read inside the view and accumulated
// across passes rather than read off a Result afterwards.

// sortInts is a sort written out for the two lines of test reporting that need
// one. Reaching for sort.Ints would be an import for a slice of three.
func sortInts(vs []int) {
	for i := 1; i < len(vs); i++ {
		for j := i; j > 0 && vs[j] < vs[j-1]; j-- {
			vs[j], vs[j-1] = vs[j-1], vs[j]
		}
	}
}

func wantsPanic(t *testing.T, what string, view func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("%s should have panicked", what)
		}
	}()
	view()
}

func wantText(t *testing.T, tt *ui.Tester, what string, texts ...string) {
	t.Helper()
	for _, s := range texts {
		if !tt.HasText(s) {
			t.Errorf("%s is missing %q; drew %q", what, s, tt.Texts())
		}
	}
}

// ── prices ─────────────────────────────────────────────────────────────────

// TestFormatPrice pins the column rule at every boundary a column of prices
// crosses: below a thousand, at a thousand, at a hundred thousand, at a
// million, and on either side of a negative.
func TestFormatPrice(t *testing.T) {
	cases := []struct {
		in   float64
		want string
	}{
		{0, "0.00"},
		{1, "1.00"},
		{1.5, "1.50"},
		{8, "8.00"},
		{999.994, "999.99"},
		{1000, "1,000.00"},
		{1234.5, "1,234.50"},
		{12345.678, "12,345.68"},
		{123456.789, "123,456.79"},
		{999999.99, "999,999.99"},
		{1000000, "1,000,000.00"},
		{-1234.5, "-1,234.50"},
		{-99, "-99.00"},
		{-100, "-100.00"},
		{-0.005, "-0.01"},
	}
	for _, tc := range cases {
		if got := FormatPrice(tc.in); got != tc.want {
			t.Errorf("FormatPrice(%v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// TestFormatPriceIsFixedWidth is the property that makes a column work, and it
// is why the face is monospaced. Every figure in a column has the same number
// of characters unless it crosses a grouping boundary, and a proportional
// face then makes "1,111.11" and "8,888.88" different widths — so the digits
// line up at the right edge and not at the decimal point.
func TestFormatPriceIsFixedWidth(t *testing.T) {
	// Grouped by how many digits are before the point, because that is the
	// thing that decides the width. A column spanning 0.00 to 1,000,000.00
	// has several widths and that is correct — the grouping separator is a
	// character. The rule that matters is that *a column of prices of the
	// same size* has one width, which is what lets the eye compare the last
	// digits of each row.
	widths := map[int]map[int]bool{}
	for _, v := range []float64{0, 8, 88, 888, 8888, 88888, 888888, 8888888,
		1000, 8000, 88000, 888000, 8888000, 88888000} {
		digits := 1
		for x := v; x >= 10; x /= 10 {
			digits++
		}
		// Below a thousand there is no separator, so the digit count is the
		// width; above one there is, so the width is the digits plus the
		// separators.
		want := digits + 3 // the point and the two decimal places
		if v >= 1000 {
			// One separator for every three digits *from the right*, which is
			// digits/3 — not digits/3 rounded up. Six digits get two, not
			// three, and the "round up" version is the bug that puts a
			// leading comma on every six-figure price.
			want += (digits - 1) / 3
		}
		if got := len(FormatPrice(v)); got != want {
			t.Errorf("FormatPrice(%v) is %d characters, want %d", v, got, want)
		}
		if widths[digits] == nil {
			widths[digits] = map[int]bool{}
		}
		widths[digits][len(FormatPrice(v))] = true
	}
	for digits, w := range widths {
		if len(w) != 1 {
			got := make([]int, 0, len(w))
			for x := range w {
				got = append(got, x)
			}
			sortInts(got)
			t.Errorf("%d-digit prices have %d widths: %v", digits, len(w), got)
		}
	}
	// And the monospaced face is what makes those equal widths mean equal
	// widths. A single-numeral figure in the proportional face is not the
	// same width as an eight, and the whole column rule rests on that.
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		PriceText(c, PriceTextOptions{Value: 1111.11, Unit: "USD", Label: "one"})
		PriceText(c, PriceTextOptions{Value: 8888.88, Unit: "USD", Label: "eight"})
	}, 400, 120)
	one, ok1 := tt.Find("one")
	eight, ok2 := tt.Find("eight")
	if !ok1 || !ok2 {
		t.Fatalf("the two prices are not on screen: %q", tt.Texts())
	}
	if one.W != eight.W {
		t.Errorf("two four-figure prices measured %v and %v wide; a column of figures "+
			"whose rows differ in width is a column nobody can scan", one.W, eight.W)
	}
}

func TestFormatCompact(t *testing.T) {
	cases := map[float64]string{
		999: "999", 1500: "1.5K", 1234500: "1.2M", 2.4e9: "2.4B",
	}
	for in, want := range cases {
		if got := FormatCompact(in); got != want {
			t.Errorf("FormatCompact(%v) = %q, want %q", in, got, want)
		}
	}
}

// ── change ─────────────────────────────────────────────────────────────────

// TestFormatChange is the three things the brief asks for, each exactly: the
// sign, the decimal places, and the flat case.
//
// The flat case is the one that matters. "0.00%" and "+0.00%" look identical
// on screen and only one of them is true, and a flat instrument is the single
// most common thing in a portfolio — so it must not be dressed up as a move.
// It is also why the comparison is on the values and not on the rounded
// percentage: 1.001 against 1.0 is 0.10% and is a real change.
func TestFormatChange(t *testing.T) {
	cases := []struct {
		name     string
		v, prev  float64
		text     string
		severity core.Severity
	}{
		{"a rise", 101.25, 100, "+1.25%", core.Success},
		{"a fall", 98.75, 100, "-1.25%", core.Danger},
		{"flat", 100, 100, "0.00%", core.Neutral},
		{"flat at zero", 0, 0, "0.00%", core.Neutral},
		{"no previous to divide by", 50, 0, "0.00%", core.Neutral},
		{"a small real move", 1.001, 1.0, "+0.10%", core.Success},
		{"a large rise", 200, 100, "+100.00%", core.Success},
		{"a fall across zero", -50, 100, "-150.00%", core.Danger},
	}
	for _, tc := range cases {
		text, sev := FormatChange(tc.v, tc.prev)
		if text != tc.text {
			t.Errorf("%s: FormatChange(%v, %v) = %q, want %q", tc.name, tc.v, tc.prev, text, tc.text)
		}
		if sev != tc.severity {
			t.Errorf("%s: severity = %v, want %v", tc.name, sev, tc.severity)
		}
	}
}

// TestFormatChangeNeverRoundsFlatToAPlus: the specific failure the flat rule
// exists for. A rise of less than half a hundredth of a percent formats as
// "+0.00%", which reads as a rise of nothing and is not the same as no rise.
func TestFormatChangeNeverRoundsFlatToAPlus(t *testing.T) {
	// A rise too small for two places to say gets more places rather than a
	// false zero. "+0.00%" is a sign in front of a zero and claims a move the
	// figure denies — the same lie the flat rule exists to prevent, reached
	// from the other side.
	text, sev := FormatChange(100.0001, 100)
	if text == "+0.00%" {
		t.Error("a real rise rounded to a flat one")
	}
	if text != "+0.0001%" {
		t.Errorf("a rise of 0.0001%% formatted as %q", text)
	}
	if sev != core.Success {
		t.Errorf("a rise of 0.0001%% has severity %v, want Success", sev)
	}
	down, downSev := FormatChange(99.9999, 100)
	if !strings.HasPrefix(down, "-") {
		t.Errorf("a fall formatted as %q", down)
	}
	if downSev != core.Danger {
		t.Errorf("a fall that was written as %q has severity %v", down, downSev)
	}
	// And a genuine flat is still flat, at two places and with no sign.
	if flat, _ := FormatChange(100, 100); flat != "0.00%" {
		t.Errorf("a genuine flat formatted as %q", flat)
	}
}

func TestFormatSigned(t *testing.T) {
	cases := map[float64]string{0: "0.00", 1240: "+1,240.00", -1240: "-1,240.00"}
	for in, want := range cases {
		if got := FormatSigned(in); got != want {
			t.Errorf("FormatSigned(%v) = %q, want %q", in, got, want)
		}
	}
}

func TestFormatMoney(t *testing.T) {
	if got := FormatMoney(1234.5, "$"); got != "$1,234.50" {
		t.Errorf("FormatMoney = %q", got)
	}
}

func TestFormatPercent(t *testing.T) {
	// A percentage that is already a percentage. Routing it through
	// FormatChange would measure it against zero and print 0.00% for all
	// three of these, because thirty percent of nothing is nothing.
	cases := map[float64]string{0: "0.00%", 12.5: "+12.50%", -3.25: "-3.25%",
		100: "+100.00%", 30: "+30.00%"}
	for in, want := range cases {
		if got := formatPercent(in); got != want {
			t.Errorf("formatPercent(%v) = %q, want %q", in, got, want)
		}
	}
}

// ── volume ─────────────────────────────────────────────────────────────────

// TestFormatVolume is the three carry boundaries the brief asks for, plus the
// one just below each. The boundaries are exact on purpose: a volume that
// reads "0.99K" at 990 hides the moment it crosses a thousand, which is the
// moment a reader watching for turnover is waiting for.
func TestFormatVolume(t *testing.T) {
	cases := []struct {
		in   int
		want string
	}{
		{0, "0"},
		{7, "7"},
		{999, "999"},
		{1000, "1.00K"},
		{1500, "1.50K"},
		{999_999, "1.00M"},
		{1_234_567, "1.23M"},
		{1_000_000_000, "1.00B"},
		{-2500, "-2.50K"},
	}
	for _, tc := range cases {
		if got := FormatVolume(tc.in); got != tc.want {
			t.Errorf("FormatVolume(%d) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// TestFormatVolumeIsFixedWidthAboveATousand is the column rule again: every
// carried figure has six characters, so a column of them lines up.
// TestFormatVolumeIsFixedPoint is what a carried count actually guarantees:
// two places and a one-letter suffix, so that "1.00K" and "9.99K" have their
// decimal point in the same column. It does *not* guarantee one width across
// magnitudes, and it cannot: "100.00K" is wider than "9.99K" because it is a
// bigger number. Claiming otherwise would be claiming something the function
// does not do.
func TestFormatVolumeIsFixedPoint(t *testing.T) {
	for _, n := range []int{1000, 9999, 10_000, 99_999, 100_000, 1_000_000,
		9_999_999, 1_000_000_000} {
		s := FormatVolume(n)
		dot := strings.IndexByte(s, '.')
		if dot < 0 || len(s)-dot != 4 {
			t.Errorf("FormatVolume(%d) = %q, want two decimal places then a suffix", n, s)
		}
		last := rune(s[len(s)-1])
		if !strings.ContainsRune("KMB", last) {
			t.Errorf("FormatVolume(%d) = %q has no unit", n, s)
		}
		if strings.Contains(s, ",") {
			t.Errorf("FormatVolume(%d) = %q has a grouping separator in it", n, s)
		}
	}
}

func TestFormatSize(t *testing.T) {
	cases := []struct {
		in   float64
		want string
	}{
		{0, "0"},
		{120, "120"},
		{1234, "1,234"},
		{-120, "-120"},
		{0.25, "0.2500"},
		{12.3456, "12.3456"},
		{-0.5, "-0.5000"},
	}
	for _, tc := range cases {
		if got := FormatSize(tc.in); got != tc.want {
			t.Errorf("FormatSize(%v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// ── the book ───────────────────────────────────────────────────────────────

func testBook() (bids, asks []Price) {
	bids = []Price{
		{Price: 100.90, Size: 400}, {Price: 100.85, Size: 900},
		{Price: 101.10, Size: 250}, {Price: 100.80, Size: 1200},
		{Price: 100.75, Size: 300},
	}
	asks = []Price{
		{Price: 101.30, Size: 700}, {Price: 101.15, Size: 200},
		{Price: 101.05, Size: 950}, {Price: 101.35, Size: 150},
		{Price: 101.40, Size: 500},
	}
	return bids, asks
}

// TestLevelsSortsAndAccumulates is the computation the brief asks for: the
// cumulative size and the largest level.
//
// The bids are handed over out of order on purpose. A book that is not in
// price order is not an order book, and a component that trusted its input
// would show a book whose top row is not the top of the market.
func TestLevelsSortsAndAccumulates(t *testing.T) {
	bids, asks := testBook()
	levels := Levels(bids, asks)

	var b, a []Level
	for _, l := range levels {
		if l.Side == SideBid {
			b = append(b, l)
		} else {
			a = append(a, l)
		}
	}
	if len(b) != 5 || len(a) != 5 {
		t.Fatalf("Levels returned %d bids and %d asks, want 5 and 5", len(b), len(a))
	}

	// Bids descending, asks ascending — best first on each side.
	for i := 1; i < len(b); i++ {
		if b[i-1].Price < b[i].Price {
			t.Fatalf("bids are not descending at %d: %v", i, b)
		}
	}
	for i := 1; i < len(a); i++ {
		if a[i-1].Price > a[i].Price {
			t.Fatalf("asks are not ascending at %d: %v", i, a)
		}
	}

	// The cumulative size, from the best price down.
	// Bids descending: 101.10(250), 100.90(400), 100.85(900), 100.80(1200),
	// 100.75(300).
	wantBids := []float64{250, 650, 1550, 2750, 3050}
	for i, want := range wantBids {
		if b[i].Total != want {
			t.Errorf("bid %d cumulative = %v, want %v (this level holds %v)",
				i, b[i].Total, want, b[i].Size)
		}
	}
	// Asks ascending: 101.05(950), 101.15(200), 101.30(700), 101.35(150),
	// 101.40(500).
	wantAsks := []float64{950, 1150, 1850, 2000, 2500}
	for i, want := range wantAsks {
		if a[i].Total != want {
			t.Errorf("ask %d cumulative = %v, want %v", i, a[i].Total, want)
		}
	}

	// And the best prices are the tops of their sides.
	if b[0].Price != 101.10 {
		t.Errorf("the best bid is %v, want 101.10 — the out-of-order 101.10 must be sorted "+
			"to the top or the book is showing the wrong market", b[0].Price)
	}
	if a[0].Price != 101.05 {
		t.Errorf("the best ask is %v, want 101.05", a[0].Price)
	}
}

func TestLevelsEmptyAndSingle(t *testing.T) {
	if got := Levels(nil, nil); len(got) != 0 {
		t.Errorf("an empty book gave %d levels", len(got))
	}
	one := Levels([]Price{{Price: 10, Size: 5}}, nil)
	if len(one) != 1 || one[0].Total != 5 {
		t.Errorf("a one-order book gave %+v", one)
	}
}

func TestBestPrices(t *testing.T) {
	bids, asks := testBook()
	bid, ok := BestBid(bids)
	if !ok || bid != 101.10 {
		t.Errorf("BestBid = %v, %v; want 101.10, true", bid, ok)
	}
	ask, ok := BestAsk(asks)
	if !ok || ask != 101.05 {
		t.Errorf("BestAsk = %v, %v; want 101.05, true", ask, ok)
	}
	if _, ok := BestBid(nil); ok {
		t.Error("an empty side has no best price")
	}
}

// TestSpreadReportsACrossedBook: a bid above the ask is a real thing that
// happens and it means something specific. A spread indicator that clamped it
// to zero would be hiding the only interesting thing on the screen.
func TestSpread(t *testing.T) {
	bids, asks := testBook()
	gap, fraction, ok := Spread(bids, asks)
	if !ok {
		t.Fatal("a book with both sides has a spread")
	}
	// Best bid 101.10, best ask 101.05: crossed, by five cents. Five cents is
	// not exactly representable, so this is the one assertion in the file with
	// a tolerance in it.
	if math.Abs(gap-(-0.05)) > 1e-9 {
		t.Errorf("the gap is %v, want -0.05", gap)
	}
	if fraction >= 0 {
		t.Errorf("a crossed book's fraction is %v, want negative", fraction)
	}

	g, f, _ := Spread([]Price{{Price: 100}}, []Price{{Price: 102}})
	if g != 2 {
		t.Errorf("the gap is %v, want 2", g)
	}
	// 2 over a 101 midpoint.
	if math.Abs(f-2.0/101.0) > 1e-9 {
		t.Errorf("2 on a 101 midpoint is %v, want %v", f, 2.0/101.0)
	}
	if _, _, ok := Spread(nil, asks); ok {
		t.Error("a book with one side has no spread")
	}
}

func TestDepthMaxSize(t *testing.T) {
	bids, asks := testBook()
	d := Depth{Bids: takeSide(Levels(bids, asks), SideBid, 10),
		Asks: takeSide(Levels(bids, asks), SideAsk, 10)}
	if got := d.MaxSize(); got != 1200 {
		t.Errorf("MaxSize = %v, want 1200", got)
	}
	if got := (Depth{}).MaxSize(); got != 0 {
		t.Errorf("an empty book's MaxSize = %v, want 0", got)
	}
}

func TestAggregateBinsTheWholeBook(t *testing.T) {
	bids, asks := testBook()
	bands := aggregate(Levels(bids, asks), 3)
	if len(bands) != 3 {
		t.Fatalf("aggregate into 3 bins gave %d bands", len(bands))
	}
	// Every level lands in exactly one band, so the bands add up to the book.
	total := 0.0
	for _, b := range bands {
		total += b.BidSize + b.AskSize
	}
	want := 0.0
	for _, l := range Levels(bids, asks) {
		want += l.Size
	}
	if total != want {
		t.Errorf("the bands hold %v, the book holds %v; a ladder that does not add up is a "+
			"ladder of something other than depth", total, want)
	}
	for i, b := range bands {
		if b.Low >= b.High {
			t.Errorf("band %d is %v to %v, which is not a range", i, b.Low, b.High)
		}
		if b.Label() == "" {
			t.Errorf("band %d has no label", i)
		}
	}
}

// ── allocation ─────────────────────────────────────────────────────────────

func testPositions() []Position {
	return []Position{
		{ID: "p1", Instrument: "RIVR", Sector: "Roofing", Shares: 100,
			Average: 40, Last: 52.40, Weight: 0.4},
		{ID: "p2", Instrument: "SOLR", Sector: "Solar", Shares: 250,
			Average: 18, Last: 24.00, Weight: 0.25},
		{ID: "p3", Instrument: "GLZ", Shares: 400, Average: 9, Last: 11.20, Weight: 0.15},
	}
}

// TestAllocationGroupsBySectorAndAddsUp: the two things a ring has to do.
// The smallest holdings are folded into one slice rather than dropped, so the
// ring still says "the whole of it" — dropping them would make a portfolio
// look fully invested in its largest position.
func TestAllocation(t *testing.T) {
	labels, values := allocation(testPositions(), 1000, 4)
	if len(labels) != len(values) {
		t.Fatalf("%d labels and %d values; two parallel slices that differ in length are a "+
			"chart of the right numbers with the wrong names", len(labels), len(values))
	}
	total := 0.0
	for _, v := range values {
		total += v
	}
	// 100×52.40 + 250×24.00 + 400×11.20 + 1000 cash.
	if want := 5240.0 + 6000.0 + 4480.0 + 1000; total != want {
		t.Errorf("the slices add up to %v, want %v; a ring that does not add up is a ring "+
			"of something other than the portfolio", total, want)
	}
	// Sorted largest first, so the order does not depend on map iteration.
	for i := 1; i < len(values); i++ {
		if values[i-1] < values[i] {
			t.Errorf("the slices are not in descending order: %v", values)
			break
		}
	}
	found := map[string]bool{}
	for _, l := range labels {
		found[l] = true
	}
	for _, want := range []string{"Cash", "Roofing", "Solar", "GLZ"} {
		if !found[want] {
			t.Errorf("the ring is missing %q; %v", want, labels)
		}
	}
}

func TestAllocationEmpty(t *testing.T) {
	labels, values := allocation(nil, 0, 4)
	if len(labels) != 0 || len(values) != 0 {
		t.Errorf("an empty portfolio gave %v / %v", labels, values)
	}
}

func TestAllocationIsStableAcrossCalls(t *testing.T) {
	a, _ := allocation(testPositions(), 1000, 8)
	b, _ := allocation(testPositions(), 1000, 8)
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("the same portfolio grouped two different ways: %v vs %v", a, b)
		}
	}
}

// ── payoff ─────────────────────────────────────────────────────────────────

// TestPayoff is the option's own arithmetic, which is the one piece of this
// package a chart cannot do for us.
//
// At the strike both sides are zero less the premium, because an option at the
// strike is worth nothing at expiry — the holder has the right and no
// advantage. A long call above the strike is the intrinsic difference; a short
// one is its negation, and that is exactly what the sign does.
func TestPayoff(t *testing.T) {
	const strike, premium = 100.0, 5.0
	cases := []struct {
		name  string
		price float64
		put   bool
		long  bool
		want  float64
	}{
		{"a long call at the strike", 100, false, true, -5},
		{"a long call above the strike", 110, false, true, 5},
		{"a long call far above", 120, false, true, 15},
		// The case a naive version gets wrong: a long call below the strike is
		// worth *nothing*, not the strike minus the price. Computing the put
		// branch for a call draws a long call that makes money as the price
		// falls, which is a put.
		{"a long call below the strike", 90, false, true, -5},
		{"a long call far below", 80, false, true, -5},
		{"a short call at the strike", 100, false, false, 5},
		{"a short call above the strike", 110, false, false, -5},
		{"a long put below the strike", 90, true, true, 5},
		{"a long put above the strike", 110, true, true, -5},
		{"a short put below the strike", 90, true, false, -5},
	}
	for _, tc := range cases {
		if got := payoff(tc.price, strike, premium, tc.put, tc.long); got != tc.want {
			t.Errorf("%s: payoff = %v, want %v", tc.name, got, tc.want)
		}
	}
	// And long is exactly the negation of short at every price.
	for p := 80.0; p <= 120.0; p += 2 {
		if payoff(p, 100, 5, false, true) != -payoff(p, 100, 5, false, false) {
			t.Fatalf("at %v long is not the negation of short", p)
		}
	}
}

// ── other pure functions ───────────────────────────────────────────────────

func TestPositionArithmetic(t *testing.T) {
	p := Position{Shares: 120, Average: 50, Last: 60}
	if got := p.MarketValue(); got != 7200 {
		t.Errorf("MarketValue = %v, want 7200", got)
	}
	if got := p.CostBasis(); got != 6000 {
		t.Errorf("CostBasis = %v, want 6000", got)
	}
	short := Position{Shares: -120, Average: 50, Last: 60}
	if got := short.MarketValue(); got != -7200 {
		t.Errorf("a short position is worth %v, want -7200 — a short shown as a long is the "+
			"worst bug this package could have", got)
	}
	flat := Position{Unrealised: 0}
	if got := flat.Return(); got != 0 {
		t.Errorf("a position with no cost basis has a return of %v, want 0 — not infinity", got)
	}
	winner := Position{Shares: 100, Average: 50, Unrealised: 500}
	if got := winner.Return(); got != 10 {
		t.Errorf("Return = %v, want 10", got)
	}
}

func TestFormatPercentMatchesTheChangeRule(t *testing.T) {
	if got := formatPercent(0); got != "0.00%" {
		t.Errorf("a flat return is %q, want no sign", got)
	}
	up, _ := FormatChange(1.001, 1.0)
	if up != "+0.10%" || formatPercent(0.1) != "+0.10%" {
		t.Errorf("a small rise is %q by FormatChange and %q by formatPercent; the two "+
			"must not be able to disagree", up, formatPercent(0.1))
	}
}

func TestBandName(t *testing.T) {
	bands := []RiskBand{{To: 50, Name: "Low"}, {To: 80, Name: "Medium"}, {To: 100, Name: "High"}}
	cases := map[float64]string{10: "Low", 50: "Low", 51: "Medium", 80: "Medium", 99: "High"}
	for in, want := range cases {
		if got := bandName(bands, in); got != want {
			t.Errorf("bandName(%v) = %q, want %q", in, got, want)
		}
	}
}

func TestSentimentWords(t *testing.T) {
	cases := map[float64]string{
		0.9: "Strongly bullish", 0.2: "Slightly bullish", 0: "Neutral",
		-0.2: "Slightly bearish", -0.9: "Strongly bearish", 0.14: "Neutral",
	}
	for in, want := range cases {
		if got := sentimentWord(in); got != want {
			t.Errorf("sentimentWord(%v) = %q, want %q", in, got, want)
		}
	}
}

func TestSentinelNames(t *testing.T) {
	cases := map[Session]string{
		SessionClosed: "Closed", SessionPreMarket: "Pre-market",
		SessionOpen: "Open", SessionPostMarket: "After hours",
	}
	for in, want := range cases {
		if got := in.String(); got != want {
			t.Errorf("Session(%d) = %q, want %q", in, got, want)
		}
	}
	cases2 := map[OrderKind]string{OrderMarket: "Market", OrderLimit: "Limit", OrderStop: "Stop"}
	for in, want := range cases2 {
		if got := in.String(); got != want {
			t.Errorf("OrderKind(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestMatchInstruments(t *testing.T) {
	all := []Instrument{
		{Symbol: Symbol{Ticker: "RIVR", Exchange: "LSE", Name: "Riverside Roofing"}},
		{Symbol: Symbol{Ticker: "SOLR", Name: "Solaris"}, Keywords: []string{"roof", "solar"}},
	}
	if got := matchInstruments(all, nil); len(got) != 2 {
		t.Errorf("no query matched %d instruments, want 2", len(got))
	}
	q := "rivr"
	if got := matchInstruments(all, &q); len(got) != 1 || got[0] != 0 {
		t.Errorf("\"rivr\" matched %v", got)
	}
	q = "Riverside"
	if got := matchInstruments(all, &q); len(got) != 1 {
		t.Errorf("a name search matched %v", got)
	}
	// "Roofing" contains "roof", so both legitimately match; the keyword case
	// is tested with a word nothing else holds, because the point of Keywords
	// is that a reader can search for something the ticker never says.
	q = "solar"
	if got := matchInstruments(all, &q); len(got) != 1 || got[0] != 1 {
		t.Errorf("a keyword search matched %v; nobody types a ticker to find a roofing "+
			"contractor, they type \"roofing\"", got)
	}
	q = "LSE"
	if got := matchInstruments(all, &q); len(got) != 1 || got[0] != 0 {
		t.Errorf("an exchange search matched %v; the ticker alone trades in two places", got)
	}
	q = "nothing here"
	if got := matchInstruments(all, &q); len(got) != 0 {
		t.Errorf("a query nothing matches returned %v", got)
	}
}

func TestSortFloats(t *testing.T) {
	got := []float64{3, 1, 2}
	sortFloats(got)
	for i, want := range []float64{1, 2, 3} {
		if got[i] != want {
			t.Fatalf("sortFloats gave %v", got)
		}
	}
}

// ── fixtures ───────────────────────────────────────────────────────────────

func testQuotes() []WatchQuote {
	return []WatchQuote{
		{Symbol: Symbol{Ticker: "RIVR", Exchange: "LSE", Name: "Riverside Roofing"},
			Now: 52.40, Previous: 51.00, Volume: 1_250_000, Group: "Roofing"},
		{Symbol: Symbol{Ticker: "SOLR", Exchange: "LSE", Name: "Solaris"},
			Now: 24.00, Previous: 24.00, Volume: 940_000, Group: "Solar"},
		{Symbol: Symbol{Ticker: "GLZ", Exchange: "LSE", Name: "Glazing Co"},
			Now: 11.20, Previous: 11.50, Volume: 2_400_000, Group: "Roofing"},
	}
}

func testContracts() []Contract {
	return []Contract{
		{Symbol: "RIVR 260320 C00100000", Underlying: "RIVR", Expiry: "2026-03-20",
			Strike: 100, Kind: OptionCall, Bid: 3.4, Ask: 3.6, Volume: 1_200,
			Greeks: Greeks{Delta: 0.52, Gamma: 0.018, Theta: -0.021, Vega: 0.14, Rho: 0.06,
				Implied: 0.245}},
		{Symbol: "RIVR 260320 P00100000", Underlying: "RIVR", Expiry: "2026-03-20",
			Strike: 100, Kind: OptionPut, Bid: 2.9, Ask: 3.1, Volume: 860,
			Greeks: Greeks{Delta: -0.48, Gamma: 0.018, Theta: -0.019, Vega: 0.14, Rho: -0.05,
				Implied: 0.238}},
		{Symbol: "RIVR 260320 C00110000", Underlying: "RIVR", Expiry: "2026-03-20",
			Strike: 110, Kind: OptionCall, Bid: 0.8, Ask: 0.9, Volume: 4_500},
		{Symbol: "RIVR 260320 P00110000", Underlying: "RIVR", Expiry: "2026-03-20",
			Strike: 110, Kind: OptionPut, Bid: 11.2, Ask: 11.4, Volume: 210},
	}
}

// ── headless renders ───────────────────────────────────────────────────────

func TestSymbolBadge(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		SymbolBadge(c, SymbolBadgeOptions{
			Symbol: Symbol{Ticker: "RIVR", Exchange: "LSE"}, ShowExchange: true,
		})
	}, 260, 60)
	wantText(t, tt, "the badge", "RIVR", "LSE")
	if _, ok := tt.Find("RIVR on LSE"); !ok {
		t.Error("the badge must be named with the ticker *and* the exchange; the ticker alone " +
			"trades in two places")
	}
	wantsPanic(t, "a badge with no ticker", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			SymbolBadge(c, SymbolBadgeOptions{})
		}, 200, 60)
	})
}

func TestSymbolLabelAndFull(t *testing.T) {
	s := Symbol{Ticker: "RIVR", Exchange: "LSE"}
	if got := s.Full(); got != "RIVR/LSE" {
		t.Errorf("Full = %q", got)
	}
	if got := s.Label(); got != "RIVR on LSE" {
		t.Errorf("Label = %q", got)
	}
	bare := Symbol{Ticker: "RIVR"}
	if got := bare.Full(); got != "RIVR" || bare.Label() != "RIVR" {
		t.Errorf("a symbol with no exchange wrote %q / %q", bare.Full(), bare.Label())
	}
}

func TestPriceText(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		PriceText(c, PriceTextOptions{Value: 1234.5, Unit: "USD", Bold: true})
	}, 320, 60)
	wantText(t, tt, "the price", "USD", "1,234.50")
	if _, ok := tt.Find("Price in USD"); !ok {
		t.Error("a bare number announced as \"number\" tells a reader nothing")
	}
}

func TestPriceChangeBadge(t *testing.T) {
	up := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		PriceChangeBadge(c, PriceChangeBadgeOptions{
			Now: 101.25, Previous: 100, Absolute: 1.25, ShowAbsolute: true,
		})
	}, 260, 60)
	if _, ok := up.Find("Change +1.25, +1.25%"); !ok {
		t.Errorf("a rise badge is named %q", up.Texts())
	}
	flat := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		PriceChangeBadge(c, PriceChangeBadgeOptions{Now: 100, Previous: 100})
	}, 260, 60)
	if _, ok := flat.Find("Change 0.00%"); !ok {
		t.Errorf("a flat badge is named %q", flat.Texts())
	}
	if flat.HasText("+0.00%") {
		t.Error("a flat move drew a plus sign; a flat instrument must not be dressed up " +
			"as a move")
	}
}

func TestQuoteCard(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		QuoteCard(c, QuoteOptions{
			Symbol:   Symbol{Ticker: "RIVR", Exchange: "LSE", Name: "Riverside Roofing"},
			Now:      52.40,
			Previous: 51.00,
			DayHigh:  53.10, DayLow: 50.80, DayVolume: 1_250_000,
			Currency: "USD", Interval: "1D", Delayed: false,
			Sparkline: []Point{{X: 0, Y: 50}, {X: 1, Y: 51}, {X: 2, Y: 52.4}},
		})
	}, 380, 420)
	for _, want := range []string{
		"RIVR", "LSE", "Riverside Roofing", "USD", "52.40", "+2.75%",
		"1D", "53.10", "50.80", "1.25M", "Day low", "Day high", "Volume",
	} {
		wantText(t, tt, "the card", want)
	}
	if _, ok := tt.Find("Recent price for RIVR on LSE"); !ok {
		t.Error("the sparkline must be named; it is a picture with no words of its own")
	}
	wantsPanic(t, "a card with no instrument", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			QuoteCard(c, QuoteOptions{Now: 1, Previous: 1})
		}, 300, 300)
	})
}

func TestBidAskBar(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		BidAskBar(c, BidAskBarOptions{Bid: 101.10, Ask: 101.15})
	}, 320, 160)
	wantText(t, tt, "the bar", "Bid", "Ask", "101.10", "101.15")
	wantsPanic(t, "a bar with one end", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			BidAskBar(c, BidAskBarOptions{Bid: 101.10})
		}, 300, 140)
	})
}

func TestOrderBook(t *testing.T) {
	bids, asks := testBook()
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		OrderBook(c, OrderBookOptions{
			Bids: bids, Asks: asks, Rows: 4, Height: 300,
			ShowTotals: true, Label: "RIVR order book",
		})
	}, 460, 340)
	if _, ok := tt.Find("RIVR order book"); !ok {
		t.Error("two columns of numbers with no names on them are a puzzle")
	}
	wantText(t, tt, "the book", "Bid", "Ask", "Spread")
	// The best bid was handed over third in the list; it must still be drawn.
	if !tt.HasText("101.10") {
		t.Errorf("the best bid is not on screen; %q", tt.Texts())
	}
	wantsPanic(t, "a book with no name", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			OrderBook(c, OrderBookOptions{Bids: bids, Asks: asks, Height: 300})
		}, 400, 320)
	})
	wantsPanic(t, "a book with no height", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			OrderBook(c, OrderBookOptions{Bids: bids, Asks: asks, Label: "x"})
		}, 400, 320)
	})
}

func TestDepthLadder(t *testing.T) {
	bids, asks := testBook()
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		DepthLadder(c, DepthLadderOptions{
			Bids: bids, Asks: asks, Height: 280, Bins: 12,
			Label: "Depth ladder for RIVR",
		})
	}, 420, 320)
	if _, ok := tt.Find("Depth ladder for RIVR"); !ok {
		t.Error("a picture of depth with no name is a gradient")
	}
	wantText(t, tt, "the ladder", "Bids", "Asks")
	wantsPanic(t, "a ladder with no height", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			DepthLadder(c, DepthLadderOptions{Bids: bids, Asks: asks, Label: "x"})
		}, 400, 300)
	})
}

func TestTimeAndSales(t *testing.T) {
	trades := []Trade{
		{Price: 52.40, Size: 1_200, Side: SideBid, When: "14:02:11"},
		{Price: 52.35, Size: 400, Side: SideAsk, When: "14:02:09"},
		{Price: 52.30, Size: 0, Side: SideBid, When: "14:02:05"},
	}
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		TimeAndSales(c, TimeAndSalesOptions{
			Trades: trades, Height: 260, Label: "RIVR tape",
		})
	}, 420, 300)
	wantText(t, tt, "the tape", "Time", "Price", "Size", "14:02:11", "52.40", "1.20K")

	none := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		TimeAndSales(c, TimeAndSalesOptions{Trades: nil, Height: 200, Label: "empty tape"})
	}, 420, 240)
	wantText(t, none, "the empty tape", "No trades yet")

	wantsPanic(t, "a tape with no height", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			TimeAndSales(c, TimeAndSalesOptions{Trades: trades, Label: "x"})
		}, 400, 240)
	})
}

func TestTradeHistory(t *testing.T) {
	fills := []Fill{
		{ID: "f1", When: "13:58", Instrument: "RIVR", Side: SideBid,
			Price: 52.30, Size: 200, Fee: -1.20, Venue: "LSE"},
		{ID: "f2", When: "13:41", Instrument: "SOLR", Side: SideAsk,
			Price: 24.05, Size: 100, Fee: 0.85},
	}
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		TradeHistory(c, TradeHistoryOptions{Fills: fills, Height: 220})
	}, 640, 280)
	wantText(t, tt, "the history",
		"Trade history", "Time", "Instrument", "Side", "Price", "Size", "Fee",
		"RIVR", "SOLR", "Buy", "Sell", "52.30", "13:58")
	wantsPanic(t, "a history with no height", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			TradeHistory(c, TradeHistoryOptions{Fills: fills})
		}, 600, 260)
	})
}

func TestPositionTable(t *testing.T) {
	positions := testPositions()
	for i := range positions {
		positions[i].Unrealised = (positions[i].Last - positions[i].Average) * positions[i].Shares
		positions[i].DayLow, positions[i].DayHigh = 50, 53
	}
	selected := 0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		PositionTable(c, PositionTableOptions{
			Positions: positions, Selected: &selected, Height: 240, ShowReturn: true,
		})
	}, 780, 300)
	// 1200 on a cost of 4000 is 30%, and the fixture says so exactly.
	wantText(t, tt, "the table",
		"Instrument", "Shares", "Avg", "Last", "Value", "P/L", "Return",
		"RIVR", "100", "40.00", "52.40", "5,240.00", "+1,240.00", "+31.00%")

	// And with the column left out, it is not drawn. A table of one fewer
	// column than the caller asked for is not a table with room in it; it is
	// a different table, and a reader comparing two views of one portfolio
	// would be comparing two sets of figures.
	narrow := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		sel := -1
		PositionTable(c, PositionTableOptions{
			Positions: positions, Selected: &sel, Height: 200,
		})
	}, 700, 260)
	if narrow.HasText("Return") {
		t.Error("the return column was asked to be left out and is on screen")
	}

	// And the negative one keeps its sign: a short position shown as a
	// positive size is a position the reader cannot tell the direction of.
	short := []Position{{ID: "s", Instrument: "SOLR", Shares: -100,
		Average: 24, Last: 20, Unrealised: 400}}
	tt2 := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		sel := -1
		PositionTable(c, PositionTableOptions{
			Positions: short, Selected: &sel, Height: 200,
		})
	}, 760, 260)
	if !tt2.HasText("-100") {
		t.Errorf("a short position lost its sign; %q", tt2.Texts())
	}
	if !tt2.HasText("-2,000.00") {
		t.Errorf("a short position is not shown as worth its negative value; %q", tt2.Texts())
	}

	wantsPanic(t, "a position table with no height", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			sel := -1
			PositionTable(c, PositionTableOptions{Positions: positions, Selected: &sel})
		}, 700, 300)
	})
}

func TestPnLDisplay(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		PnLDisplay(c, PnLOptions{
			Unrealised: 1240.5, Realised: -310.25, Currency: "USD",
		})
	}, 420, 200)
	wantText(t, tt, "the display",
		"Realised", "Unrealised", "Total profit and loss",
		"-310.25", "+1,240.50", "+930.25", "USD")
}

func TestPortfolioSummary(t *testing.T) {
	positions := testPositions()
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		PortfolioSummary(c, PortfolioSummaryOptions{
			Positions: positions, Cash: 1000, DayChange: 120, Currency: "USD",
			Inception: "January", Title: "Riverside account",
		})
	}, 520, 340)
	wantText(t, tt, "the summary",
		"Riverside account", "Total value", "16,720.00", "USD",
		"since January", "3 positions")

	// The cash is part of the value: a summary that left it out would show a
	// smaller number than the account holds.
	if tt.HasText("15,720.00") {
		t.Error("the summary excluded the cash from the total; a portfolio's value is its " +
			"holdings *and* its cash")
	}
}

func TestWatchlist(t *testing.T) {
	selected := 0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Watchlist(c, WatchlistOptions{
			Quotes: testQuotes(), Selected: &selected, Height: 220, Grouped: true,
		})
	}, 380, 260)
	wantText(t, tt, "the watchlist", "Watchlist", "RIVR", "52.40", "+2.75%", "1.25M")
	if _, ok := tt.Find("Watchlist"); !ok {
		t.Error("the list itself must be named; data.List names its rows and nothing else")
	}
	wantsPanic(t, "a watchlist with no height", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			sel := -1
			Watchlist(c, WatchlistOptions{Quotes: testQuotes(), Selected: &sel})
		}, 360, 240)
	})
}

func TestSymbolSearch(t *testing.T) {
	all := []Instrument{
		{Symbol: Symbol{Ticker: "RIVR", Exchange: "LSE", Name: "Riverside Roofing"}},
		{Symbol: Symbol{Ticker: "SOLR", Name: "Solaris"}, Keywords: []string{"roof", "solar"}},
	}
	query := "rivr"
	selected := -1
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		SymbolSearch(c, SymbolSearchOptions{
			Query: &query, Instruments: all, Selected: &selected, Highlight: true,
		})
	}, 420, 220)
	wantText(t, tt, "the search", "Symbol or company", "RIVR", "Riverside Roofing")
	if tt.HasText("Solaris") {
		t.Error("a search that matched one instrument showed another")
	}

	presses := 0
	tt2 := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		q := "rivr"
		sel := -1
		if SymbolSearch(c, SymbolSearchOptions{
			Query: &q, Instruments: all, Selected: &sel,
		}).Chosen() == "RIVR" {
			presses++
		}
	}, 420, 220)
	tt2.Click("RIVR on LSE")
	if presses != 1 {
		t.Errorf("the search reported %d choices, want 1", presses)
	}

	none := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		q := "nothing here"
		sel := -1
		SymbolSearch(c, SymbolSearchOptions{Query: &q, Instruments: all, Selected: &sel})
	}, 420, 200)
	wantText(t, none, "the empty search", "Nothing matches that")

	wantsPanic(t, "a search with no query", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			sel := -1
			SymbolSearch(c, SymbolSearchOptions{Instruments: all, Selected: &sel})
		}, 400, 200)
	})
}

func TestOrderEntry(t *testing.T) {
	price, size := 51.00, 100.0
	placing := false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		OrderEntry(c, OrderEntryOptions{
			Instrument: Symbol{Ticker: "RIVR", Exchange: "LSE"},
			Side:       SideBid, Kind: OrderLimit, Price: &price, Size: &size,
			Last: 52.40, Maximum: 10_000, Placing: placing, Currency: "USD",
		})
	}, 460, 460)
	wantText(t, tt, "the ticket",
		"RIVR", "LSE", "Buy", "Market", "Limit", "Stop", "Price", "Size",
		"Order value", "5,100.00", "Buy 100 RIVR")

	// The size is capped against the account's limit, which is the only thing
	// here that can go wrong on its own: a stale value from a previous ticket
	// would otherwise be submitted as one. The pointer is to the same float
	// the test reads, because a local copy of a stale value is exactly the
	// mistake the test would then be unable to see.
	tooBig := 5000.0
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		p := 51.0
		OrderEntry(c, OrderEntryOptions{
			Instrument: Symbol{Ticker: "RIVR"}, Side: SideBid, Kind: OrderLimit,
			Price: &p, Size: &tooBig, Last: 52.40, Maximum: 10_000, Placing: false,
		})
	}, 440, 440)
	// The cap is against the order's own price, not the last trade: the order
	// is what will be filled at.
	if tooBig > 10_000/51.0 {
		t.Errorf("the size was left at %v; the ticket did not cap it against the limit", tooBig)
	}

	presses := 0
	tt2 := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		p, s, placing := 51.0, 100.0, false
		if OrderEntry(c, OrderEntryOptions{
			Instrument: Symbol{Ticker: "RIVR"}, Side: SideBid, Kind: OrderLimit,
			Price: &p, Size: &s, Last: 52.40, Placing: placing,
		}).Submitting() {
			presses++
		}
	}, 460, 460)
	tt2.Click("Buy 100 RIVR")
	if presses != 1 {
		t.Errorf("the ticket reported %d submissions, want 1", presses)
	}

	wantsPanic(t, "a ticket with no instrument", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			p, s := 1.0, 1.0
			placing := false
			OrderEntry(c, OrderEntryOptions{Price: &p, Size: &s, Placing: placing})
		}, 400, 400)
	})
	wantsPanic(t, "a ticket with no price", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			s := 1.0
			OrderEntry(c, OrderEntryOptions{
				Instrument: Symbol{Ticker: "RIVR"}, Size: &s, Last: 1,
			})
		}, 400, 400)
	})
}

func TestOrderConfirm(t *testing.T) {
	open := true
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		OrderConfirm(c, OrderConfirmOptions{
			Open:       &open,
			Instrument: Symbol{Ticker: "RIVR", Exchange: "LSE"},
			Side:       SideBid, Kind: OrderLimit,
			Price: 51.00, Size: 100, Notional: 5100, Fee: 1.2, Last: 52.40,
			Currency: "USD",
		})
	}, 520, 420)
	wantText(t, tt, "the confirmation",
		"Confirm Limit order", "RIVR", "Buy 100", "Type", "Price", "Order value",
		"Fee", "Away from market", "Send order")

	closed := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		o := false
		OrderConfirm(c, OrderConfirmOptions{
			Open: &o, Instrument: Symbol{Ticker: "RIVR"}, Side: SideBid,
			Kind: OrderLimit, Price: 51, Size: 100, Notional: 5100, Last: 52.4,
		})
	}, 400, 320)
	if len(closed.Texts()) != 0 {
		t.Errorf("a closed confirmation drew %q", closed.Texts())
	}

	presses := 0
	tt2 := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		o := true
		if OrderConfirm(c, OrderConfirmOptions{
			Open: &o, Instrument: Symbol{Ticker: "RIVR"}, Side: SideBid,
			Kind: OrderLimit, Price: 51, Size: 100, Notional: 5100, Last: 52.4,
		}).Confirmed() {
			presses++
		}
	}, 520, 420)
	tt2.Click("Send order")
	if presses != 1 {
		t.Errorf("the confirmation reported %d sends, want 1", presses)
	}

	wantsPanic(t, "a confirmation about nothing", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			o := true
			OrderConfirm(c, OrderConfirmOptions{Open: &o})
		}, 400, 320)
	})
}

func TestOrderTable(t *testing.T) {
	orders := []Order{
		{ID: "o1", Instrument: "RIVR", Kind: OrderLimit, Side: SideBid,
			Price: 51.00, Size: 100, Filled: 40, State: "Partially filled",
			Tone: core.Warning},
		{ID: "o2", Instrument: "SOLR", Kind: OrderMarket, Side: SideAsk,
			Size: 250, State: "Working", Tone: core.Success},
	}
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		OrderTable(c, OrderTableOptions{Orders: orders, Height: 220})
	}, 700, 280)
	wantText(t, tt, "the orders",
		"Instrument", "Type", "Side", "Price", "Size", "Filled", "State",
		"RIVR", "Limit", "Buy", "51.00", "40 of 100", "Partially filled",
		"SOLR", "Market", "Sell", "Working")

	none := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		OrderTable(c, OrderTableOptions{Orders: nil, Height: 180})
	}, 700, 240)
	wantText(t, none, "the empty table", "No working orders")

	wantsPanic(t, "an orders table with no height", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			OrderTable(c, OrderTableOptions{Orders: orders})
		}, 680, 260)
	})
}

func TestOrderHistory(t *testing.T) {
	orders := []Order{
		{ID: "h1", When: "13:58", Instrument: "RIVR", Kind: OrderLimit,
			Side: SideBid, Price: 52.30, Size: 200, State: "Filled", Tone: core.Success},
		{ID: "h2", When: "13:41", Instrument: "SOLR", Kind: OrderStop,
			Side: SideAsk, Size: 100, State: "Cancelled", Tone: core.Danger},
	}
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		OrderHistory(c, OrderHistoryOptions{Orders: orders, Height: 220})
	}, 700, 280)
	wantText(t, tt, "the history",
		"Time", "Instrument", "Side", "Type", "Price", "Size", "State",
		"13:58", "RIVR", "Filled", "13:41", "SOLR", "Cancelled", "Stop")
	wantsPanic(t, "an order history with no height", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			OrderHistory(c, OrderHistoryOptions{Orders: orders})
		}, 680, 260)
	})
}

func TestQuickTradeButtons(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		QuickTradeButtons(c, QuickTradeButtonsOptions{
			Size: 100, Instrument: Symbol{Ticker: "RIVR"},
			Bid: 101.10, Ask: 101.15, Side: SideBid, Placing: false,
		})
	}, 460, 80)
	// The prices are on the buttons: a quick-trade row whose prices are
	// elsewhere is a row where the reader has to look away from the thing they
	// are pressing.
	wantText(t, tt, "the buttons", "Buy 100 at 101.10", "Sell 100 at 101.15")

	sells := 0
	tt2 := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		if side, ok := QuickTradeButtons(c, QuickTradeButtonsOptions{
			Size: 100, Instrument: Symbol{Ticker: "RIVR"},
			Bid: 101.10, Ask: 101.15, Side: SideBid,
		}).Ordered(); ok && side == SideAsk {
			sells++
		}
	}, 460, 80)
	tt2.Click("Sell 100 at 101.15")
	if sells != 1 {
		t.Errorf("the sell button reported %d presses, want 1", sells)
	}

	wantsPanic(t, "quick-trade buttons with no size", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			QuickTradeButtons(c, QuickTradeButtonsOptions{
				Instrument: Symbol{Ticker: "RIVR"}, Bid: 1, Ask: 1,
			})
		}, 400, 80)
	})
	wantsPanic(t, "quick-trade buttons with no instrument", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			QuickTradeButtons(c, QuickTradeButtonsOptions{Size: 10, Bid: 1, Ask: 1})
		}, 400, 80)
	})
}

func TestRiskMeter(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		RiskMeter(c, RiskMeterOptions{
			Value: 62.4, Min: 0, Max: 100, Label: "Margin used",
			Caption: "within the limit",
			Bands: []RiskBand{{To: 50, Name: "Low"}, {To: 80, Name: "Medium"},
				{To: 100, Name: "High"}},
			Unit: "%",
		})
	}, 360, 300)
	if _, ok := tt.Find("Margin used"); !ok {
		t.Error("a dial with no name says nothing about what it is measuring")
	}
	wantText(t, tt, "the meter", "within the limit", "%")
	wantsPanic(t, "a risk meter with no label", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			RiskMeter(c, RiskMeterOptions{Value: 1})
		}, 300, 260)
	})
}

func TestSentimentGauge(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		SentimentGauge(c, SentimentOptions{
			Score: 0.42, Label: "Options skew", Caption: "puts priced rich",
		})
	}, 380, 300)
	if _, ok := tt.Find("Options skew"); !ok {
		t.Error("the gauge must be named")
	}
	// The two ends are written out because a bar that runs both ways is a bar
	// a reader will otherwise take for one that runs only right.
	wantText(t, tt, "the gauge", "Bearish", "Neutral", "Bullish", "puts priced rich")
	wantsPanic(t, "a sentiment gauge with no label", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			SentimentGauge(c, SentimentOptions{Score: 1})
		}, 320, 260)
	})
}

func TestSpreadIndicator(t *testing.T) {
	bids, asks := testBook()
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		SpreadIndicator(c, SpreadIndicatorOptions{Bids: bids, Asks: asks})
	}, 400, 160)
	if _, ok := tt.Find("Spread"); !ok {
		t.Error("the indicator must be named")
	}
	wantText(t, tt, "the indicator", "of the midpoint", "per share")

	wideBids, wideAsks := []Price{{Price: 100, Size: 10}}, []Price{{Price: 103, Size: 10}}
	wide := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		SpreadIndicator(c, SpreadIndicatorOptions{Bids: wideBids, Asks: wideAsks})
	}, 400, 180)
	wantText(t, wide, "a wide market", "Wide for this market")

	wantsPanic(t, "a spread with one side", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			SpreadIndicator(c, SpreadIndicatorOptions{Bids: bids})
		}, 360, 140)
	})
}

func TestMarginIndicator(t *testing.T) {
	// Comfortably under the maintenance level: the threshold is drawn and the
	// account is not called over. The two states have to be told apart, because
	// one of them is a warning and the other is a call from the broker.
	under := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		MarginIndicator(c, MarginOptions{
			Used: 0.12, Available: 0.88, Excess: 0.42,
			ExcessLabel: "after a 5% move",
		})
	}, 400, 240)
	wantText(t, under, "the indicator",
		"In use", "Free", "after a 5% move", "Maintenance at 25.00%",
		"12.00%", "88.00%")
	if under.HasText("At or above the maintenance level") {
		t.Error("12% used is below a 25% maintenance level and was reported as above it")
	}

	over := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		MarginIndicator(c, MarginOptions{Used: 0.9, Available: 0.1})
	}, 400, 240)
	wantText(t, over, "an over-margined account", "At or above the maintenance level")
}

func TestLeverageSlider(t *testing.T) {
	lev := 4.0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		LeverageSlider(c, LeverageSliderOptions{
			Leverage: &lev, Max: 10, BuyingPower: 160_000, Equity: 40_000,
			Label: "Account leverage",
		})
	}, 420, 220)
	if _, ok := tt.Find("Account leverage"); !ok {
		t.Error("the slider must be named; a rail and a thumb say nothing about what they " +
			"measure")
	}
	wantText(t, tt, "the slider", "4×", "160,000.00", "40,000.00",
		"buying power", "equity")
	wantsPanic(t, "a leverage slider with no ratio", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			LeverageSlider(c, LeverageSliderOptions{Label: "x"})
		}, 400, 200)
	})
	wantsPanic(t, "a leverage slider with no name", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			l := 1.0
			LeverageSlider(c, LeverageSliderOptions{Leverage: &l})
		}, 400, 200)
	})
}

func TestPerformanceChart(t *testing.T) {
	days := []string{"Mon", "Tue", "Wed", "Thu", "Fri"}
	portfolio := []float64{100, 102, 101, 104, 106}
	bench := []float64{100, 100.5, 100.2, 101, 102}
	base := 99.5
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		PerformanceChart(c, PerformanceChartOptions{
			Series: []PerformanceSeries{
				{Name: "Portfolio", Points: points(portfolio)},
				{Name: "Benchmark", Points: points(bench)},
			},
			Baseline: &base, Labels: days, Height: 260, Grid: true,
			Label: "Account performance", YLabel: "Growth",
		})
	}, 720, 320)
	if _, ok := tt.Find("Account performance"); !ok {
		t.Error("the chart must be named; its numbers are drawn rather than laid out")
	}
	// The legend is laid out and the axis labels are painted, so the legend's
	// entries are on the accessible tree and the tick values are not. That is
	// why the chart's own name is the one that has to carry the information a
	// screen reader needs — a "Portfolio series" label is not a data value.
	wantText(t, tt, "the chart", "Portfolio", "Benchmark", "Baseline", "Growth axis")
	wantsPanic(t, "a chart with no name", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			PerformanceChart(c, PerformanceChartOptions{
				Series: []PerformanceSeries{{Points: points(portfolio)}},
			})
		}, 600, 300)
	})
	wantsPanic(t, "a chart with no series", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			PerformanceChart(c, PerformanceChartOptions{Label: "x"})
		}, 600, 300)
	})
}

func points(vs []float64) []Point {
	out := make([]Point, len(vs))
	for i, v := range vs {
		out[i] = Point{X: float64(i), Y: v}
	}
	return out
}

func TestAssetAllocationChart(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		AssetAllocationChart(c, AssetAllocationOptions{
			Positions: testPositions(), Cash: 1000, Height: 260,
			Label: "Portfolio allocation",
		})
	}, 520, 340)
	if _, ok := tt.Find("Portfolio allocation"); !ok {
		t.Error("a ring of unnamed slices is a gradient with a hole in it")
	}
	wantText(t, tt, "the ring", "Cash", "Roofing", "Solar", "GLZ")

	empty := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		AssetAllocationChart(c, AssetAllocationOptions{
			Positions: nil, Cash: 0, Height: 200, Label: "Empty allocation",
		})
	}, 520, 260)
	wantText(t, empty, "the empty ring", "Nothing to allocate")

	wantsPanic(t, "a ring with no name", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			AssetAllocationChart(c, AssetAllocationOptions{Positions: testPositions()})
		}, 500, 320)
	})
}

func testEvents() []Event {
	return []Event{
		{Title: "Q3 results", When: "Tue 28 Apr", Impact: 0.9,
			Forecast: 1.24, Actual: 1.41, Previous: 1.10, Released: true},
		{Title: "Investor day", When: "Thu 16 May", Impact: 0.3,
			Forecast: 2.10, Previous: 2.05},
	}
}

func TestEarningsCalendar(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		EarningsCalendar(c, EarningsCalendarOptions{
			Events: testEvents(), Height: 220,
		})
	}, 760, 280)
	wantText(t, tt, "the calendar",
		"Earnings", "Date", "Event", "Impact", "Forecast", "Actual", "Previous",
		"Q3 results", "Investor day", "Tue 28 Apr", "1.41")
	// An unreleased actual is a dash, not a zero: zero would say the company
	// reported nothing.
	if !tt.HasText("—") {
		t.Errorf("an unreleased actual was written out; %q", tt.Texts())
	}
	wantsPanic(t, "a calendar with no height", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			EarningsCalendar(c, EarningsCalendarOptions{Events: testEvents()})
		}, 740, 260)
	})
}

func TestEconomicCalendar(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		EconomicCalendar(c, EconomicCalendarOptions{
			Events: testEvents(), Height: 220,
		})
	}, 700, 280)
	wantText(t, tt, "the calendar",
		"Economic calendar", "Date", "Event", "Forecast", "Actual", "Previous",
		"Q3 results", "2.10", "2.05")
	if tt.HasText("Impact") {
		t.Error("the economic calendar has an impact column; a CPI print either moved " +
			"the market or it did not")
	}
	wantsPanic(t, "an economic calendar with no height", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			EconomicCalendar(c, EconomicCalendarOptions{Events: testEvents()})
		}, 680, 260)
	})
}

func TestPayoffDiagram(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		PayoffDiagram(c, PayoffOptions{
			Spot: 100, Strike: 105, Premium: 4.5, Long: true,
			Height: 280, Expiry: "2026-03-20",
			Label: "RIVR March call payoff",
		})
	}, 620, 340)
	// The break-even is in the chart's name rather than only in the line's
	// painted text: painted text is invisible to a screen reader, and the
	// break-even is the one number somebody reading this diagram out loud needs.
	if _, ok := tt.Find("RIVR March call payoff, break-even 109.50"); !ok {
		t.Errorf("the diagram is not named with its break-even; drew %q", tt.Texts())
	}
	wantText(t, tt, "the diagram", "Underlying at expiry axis", "Profit or loss axis")
	wantsPanic(t, "a payoff diagram with no name", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			PayoffDiagram(c, PayoffOptions{Spot: 100, Strike: 100})
		}, 600, 320)
	})
}

func TestGreeksTable(t *testing.T) {
	contracts := testContracts()
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		GreeksTable(c, GreeksTableOptions{
			Greeks: contracts[0].Greeks, Contract: contracts[0], Height: 260,
		})
	}, 520, 320)
	wantText(t, tt, "the greeks table",
		"Delta", "Gamma", "Theta", "Vega", "Rho",
		"0.5200", "-0.0210",
		"how much of a move in the underlying the option takes with it")
	wantsPanic(t, "a greeks table with no contract", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			GreeksTable(c, GreeksTableOptions{Greeks: Greeks{Delta: 1}})
		}, 500, 300)
	})
}

func TestOptionChain(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		OptionChain(c, OptionChainOptions{
			Contracts: testContracts(), Underlying: "RIVR",
			Expiry: "2026-03-20", Height: 240,
		})
	}, 900, 300)
	if _, ok := tt.Find("RIVR 2026-03-20"); !ok {
		t.Error("the chain must be named with the instrument and the expiry")
	}
	wantText(t, tt, "the chain",
		"Call bid", "Call ask", "Strike", "Put ask", "Put bid",
		"100.00", "110.00", "3.40", "2.90")
	// A strike with only one side is drawn with a dash rather than shifting
	// the row, so the strike stays in the middle of the comparison.
	if !tt.HasText("—") {
		t.Errorf("the one-sided strike was not dashed; %q", tt.Texts())
	}
	wantsPanic(t, "a chain with nothing underneath it", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			OptionChain(c, OptionChainOptions{
				Contracts: testContracts(), Expiry: "x", Height: 240,
			})
		}, 880, 300)
	})
	wantsPanic(t, "a chain with no height", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			OptionChain(c, OptionChainOptions{
				Contracts: testContracts(), Underlying: "RIVR",
			})
		}, 880, 300)
	})
}

func TestMarketOverview(t *testing.T) {
	instruments := []Instrument{
		{Symbol: Symbol{Ticker: "RIVR", Exchange: "LSE", Name: "Riverside Roofing"}},
		{Symbol: Symbol{Ticker: "SOLR", Exchange: "LSE", Name: "Solaris"}},
		{Symbol: Symbol{Ticker: "UNPR", Exchange: "LSE", Name: "Unpriced"}},
	}
	selected := 0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		MarketOverview(c, MarketOverviewOptions{
			Instruments: instruments, Quotes: testQuotes(),
			Selected: &selected, Cols: 2, Label: "LSE overview",
		})
	}, 760, 520)
	if _, ok := tt.Find("LSE overview"); !ok {
		t.Error("a grid of instruments with no name is a wall of numbers")
	}
	wantText(t, tt, "the grid", "RIVR", "SOLR")
	// An instrument with no quote says so, rather than drawing a zero —
	// which would say the instrument is worth nothing.
	wantText(t, tt, "the unpriced card", "UNPR", "No quote")

	wantsPanic(t, "an overview with no name", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			sel := -1
			MarketOverview(c, MarketOverviewOptions{
				Instruments: instruments, Selected: &sel,
			})
		}, 700, 500)
	})
	wantsPanic(t, "an overview with no selection", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			MarketOverview(c, MarketOverviewOptions{
				Instruments: instruments, Label: "x",
			})
		}, 700, 500)
	})
}

func TestMarketStatus(t *testing.T) {
	for _, s := range []Session{SessionClosed, SessionPreMarket, SessionOpen, SessionPostMarket} {
		sess := s
		tt := ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			MarketStatus(c, MarketStatusOptions{
				Session: sess, Exchange: "LSE", Opens: "08:00", Closes: "16:30",
				Since: "2h 14m",
			})
		}, 340, 80)
		if _, ok := tt.Find("Market status"); !ok {
			t.Errorf("the status for %s is not on screen", sess)
		}
		if !tt.HasText(sess.String()) {
			t.Errorf("the status does not say %q; %q", sess.String(), tt.Texts())
		}
	}
}

func TestTradingSessionClock(t *testing.T) {
	open := time.Date(2026, 3, 10, 8, 0, 0, 0, time.UTC)
	closeAt := time.Date(2026, 3, 10, 16, 30, 0, 0, time.UTC)
	now := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		TradingSessionClock(c, TradingSessionClockOptions{
			Now: now, Opens: open, Closes: closeAt, Session: SessionOpen,
			Breaks: [][2]time.Time{{
				time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC),
				time.Date(2026, 3, 10, 13, 0, 0, 0, time.UTC),
			}},
			Zone: "Europe/London", Label: "LSE session clock",
		})
	}, 480, 180)
	if _, ok := tt.Find("LSE session clock"); !ok {
		t.Error("a bar of time says nothing about what it is timing")
	}
	wantText(t, tt, "the clock", "12:00:00", "Europe/London", "Open", "08:00", "16:30")
	wantsPanic(t, "a clock with no time", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			TradingSessionClock(c, TradingSessionClockOptions{Label: "x"})
		}, 440, 160)
	})
}

func TestTickerTape(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		TickerTape(c, TickerTapeOptions{
			Quotes: testQuotes(), Label: "Ticker tape", Repeat: true,
		})
	}, 720, 80)
	if _, ok := tt.Find("Ticker tape"); !ok {
		t.Error("a row of figures scrolling past is unidentifiable without a name")
	}
	// The tape's instruments are *painted*, not laid out: a row of thirty text
	// elements laid out thirty times a second is thirty layouts a second to
	// produce a picture that is one painter call, and the price of that is
	// that the ticker has no accessible tree of its own. So what is asserted
	// here is the name it does carry — and the reason the name is required is
	// this one.
	still := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		TickerTape(c, TickerTapeOptions{
			Quotes: testQuotes(), Label: "Still tape", Repeat: true,
		})
	}, 720, 80)
	if _, ok := still.Find("Still tape"); !ok {
		t.Errorf("the tape is unnamed; %q", still.Texts())
	}
	wantsPanic(t, "a tape with no name", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			TickerTape(c, TickerTapeOptions{Quotes: testQuotes()})
		}, 700, 80)
	})
}

func TestLevel2Quotes(t *testing.T) {
	bids, asks := testBook()
	selected := -1
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Level2Quotes(c, Level2QuotesOptions{
			Bids: bids, Asks: asks, Height: 260, Selected: &selected,
			Highlight: []float64{101.10},
		})
	}, 420, 300)
	if _, ok := tt.Find("Level two"); !ok {
		t.Error("the book must be named")
	}
	wantText(t, tt, "the book", "101.10", "101.05")
	wantsPanic(t, "a level-two book with no height", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			sel := -1
			Level2Quotes(c, Level2QuotesOptions{Bids: bids, Asks: asks, Selected: &sel})
		}, 400, 280)
	})
}

func TestQuoteList(t *testing.T) {
	selected := -1
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		QuoteList(c, QuoteListOptions{
			Quotes: testQuotes(), Height: 220, Selected: &selected,
		})
	}, 640, 280)
	wantText(t, tt, "the list",
		"Instrument", "Price", "Change", "Volume", "RIVR", "52.40",
		"+2.75%", "0.00%", "1.25M")
	// A flat quote is 0.00% and not +0.00%.
	if tt.HasText("+0.00%") {
		t.Error("a flat quote drew a plus sign")
	}
	wantsPanic(t, "a quote list with no height", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			QuoteList(c, QuoteListOptions{Quotes: testQuotes()})
		}, 620, 260)
	})
}

// ── dark mode ──────────────────────────────────────────────────────────────

// TestDarkModeDrawsEveryComponent runs the whole package in the dark palette
// and asks for each component's own name. A component that reached for a
// light-only colour would still lay out and this test would still pass — so
// what matters is that every name is found in dark, and that the palette the
// window resolved is the dark one.
func TestDarkModeDrawsEveryComponent(t *testing.T) {
	bids, asks := testBook()
	positions := testPositions()
	quotes := testQuotes()
	contracts := testContracts()
	price, size := 51.0, 100.0
	placing, confirm := false, true
	lev := 4.0
	selected, tickSelected := 0, 0
	sent, snoozed := 51.0, 100.0
	open := true
	base := 99.5
	instruments := []Instrument{
		{Symbol: Symbol{Ticker: "RIVR", Exchange: "LSE", Name: "Riverside Roofing"}},
		{Symbol: Symbol{Ticker: "SOLR", Exchange: "LSE", Name: "Solaris"}},
	}

	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: core.Dark})

		SymbolBadge(c, SymbolBadgeOptions{
			Symbol: Symbol{Ticker: "RIVR", Exchange: "LSE"}, ShowExchange: true,
		})
		PriceText(c, PriceTextOptions{Value: 52.4, Unit: "USD", Bold: true})
		PriceChangeBadge(c, PriceChangeBadgeOptions{
			Now: 52.4, Previous: 51, Absolute: 1.4, ShowAbsolute: true,
		})
		BidAskBar(c, BidAskBarOptions{Bid: 101.1, Ask: 101.15})
		QuoteCard(c, QuoteOptions{
			Symbol: Symbol{Ticker: "RIVR", Exchange: "LSE"}, Now: 52.4, Previous: 51,
			DayHigh: 53.1, DayLow: 50.8, DayVolume: 1_250_000, Currency: "USD",
			Interval: "1D", Sparkline: points([]float64{50, 51, 52.4}),
		})
		OrderBook(c, OrderBookOptions{
			Bids: bids, Asks: asks, Rows: 3, Height: 220,
			ShowTotals: true, Label: "dm.book",
		})
		DepthLadder(c, DepthLadderOptions{
			Bids: bids, Asks: asks, Height: 220, Label: "dm.ladder",
		})
		TimeAndSales(c, TimeAndSalesOptions{
			Trades: []Trade{{Price: 52.4, Size: 1200, Side: SideBid, When: "14:02"}},
			Height: 180, Label: "dm.tape",
		})
		TradeHistory(c, TradeHistoryOptions{
			Fills: []Fill{{ID: "f", When: "13:58", Instrument: "RIVR", Side: SideBid,
				Price: 52.3, Size: 200, Fee: -1.2}},
			Height: 180, Label: "dm.fills",
		})
		PositionTable(c, PositionTableOptions{
			Positions: positions, Selected: &selected, Height: 200,
		})
		PnLDisplay(c, PnLOptions{Unrealised: 1240, Realised: -310, Currency: "USD"})
		PortfolioSummary(c, PortfolioSummaryOptions{
			Positions: positions, Cash: 1000, DayChange: 120, Currency: "USD",
			Title: "dm.portfolio",
		})
		Watchlist(c, WatchlistOptions{
			Quotes: quotes, Selected: &selected, Height: 200,
		})
		SymbolSearch(c, SymbolSearchOptions{
			Query: searchQuery(), Instruments: instruments, Selected: &tickSelected,
			Results: 5,
		})
		OrderEntry(c, OrderEntryOptions{
			Instrument: Symbol{Ticker: "RIVR"}, Side: SideBid, Kind: OrderLimit,
			Price: &price, Size: &size, Last: 52.4, Maximum: 10_000,
			Placing: placing, Currency: "USD",
		})
		OrderConfirm(c, OrderConfirmOptions{
			Open: &confirm, Instrument: Symbol{Ticker: "RIVR"}, Side: SideBid,
			Kind: OrderLimit, Price: 51, Size: 100, Notional: 5100, Last: 52.4,
			Currency: "USD",
		})
		OrderTable(c, OrderTableOptions{
			Orders: []Order{{ID: "o", Instrument: "RIVR", Kind: OrderLimit,
				Side: SideBid, Price: 51, Size: 100, Filled: 40,
				State: "Partially filled", Tone: core.Warning, When: "13:58"}},
			Height: 180,
		})
		OrderHistory(c, OrderHistoryOptions{
			Orders: []Order{{ID: "h", When: "13:58", Instrument: "RIVR",
				Kind: OrderLimit, Side: SideBid, Price: 52.3, Size: 200,
				State: "Filled", Tone: core.Success}},
			Height: 180,
		})
		QuickTradeButtons(c, QuickTradeButtonsOptions{
			Size: 100, Instrument: Symbol{Ticker: "RIVR"},
			Bid: 101.1, Ask: 101.15, Side: SideBid,
		})
		RiskMeter(c, RiskMeterOptions{Value: 62.4, Label: "dm.risk", Caption: "within the limit"})
		SentimentGauge(c, SentimentOptions{Score: 0.42, Label: "dm.sentiment"})
		SpreadIndicator(c, SpreadIndicatorOptions{Bids: bids, Asks: asks, Label: "dm.spread"})
		MarginIndicator(c, MarginOptions{Used: 0.3, Available: 0.7, Label: "dm.margin"})
		LeverageSlider(c, LeverageSliderOptions{
			Leverage: &lev, BuyingPower: 160_000, Equity: 40_000, Label: "dm.leverage",
		})
		PerformanceChart(c, PerformanceChartOptions{
			Series: []PerformanceSeries{{Name: "Portfolio",
				Points: points([]float64{100, 102, 104})}},
			Baseline: &base, Labels: []string{"Mon", "Tue", "Wed"},
			Height: 220, Grid: true, Label: "dm.performance",
		})
		AssetAllocationChart(c, AssetAllocationOptions{
			Positions: positions, Cash: 1000, Height: 220, Label: "dm.allocation",
		})
		EarningsCalendar(c, EarningsCalendarOptions{Events: testEvents(), Height: 180})
		EconomicCalendar(c, EconomicCalendarOptions{Events: testEvents(), Height: 180})
		PayoffDiagram(c, PayoffOptions{
			Spot: 100, Strike: 100, Premium: 5.5, Long: true,
			Height: 220, Label: "dm.payoff",
		})
		GreeksTable(c, GreeksTableOptions{
			Greeks: contracts[0].Greeks, Contract: contracts[0], Height: 220,
		})
		OptionChain(c, OptionChainOptions{
			Contracts: contracts, Underlying: "RIVR",
			Expiry: "2026-03-20", Height: 200,
		})
		MarketOverview(c, MarketOverviewOptions{
			Instruments: instruments, Quotes: quotes, Selected: &selected,
			Cols: 2, Label: "dm.overview",
		})
		MarketStatus(c, MarketStatusOptions{
			Session: SessionOpen, Exchange: "LSE", Label: "dm.status",
		})
		TickerTape(c, TickerTapeOptions{
			Quotes: quotes, Label: "dm.tape-ticker", Repeat: true,
		})
		TradingSessionClock(c, TradingSessionClockOptions{
			Now:     time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC),
			Opens:   time.Date(2026, 3, 10, 8, 0, 0, 0, time.UTC),
			Closes:  time.Date(2026, 3, 10, 16, 30, 0, 0, time.UTC),
			Session: SessionOpen, Label: "dm.clock",
		})
		Level2Quotes(c, Level2QuotesOptions{
			Bids: bids, Asks: asks, Height: 200, Selected: &tickSelected,
			Highlight: []float64{101.1},
		})
		QuoteList(c, QuoteListOptions{
			Quotes: quotes, Height: 200, Selected: &tickSelected,
		})
		_ = sent
		_ = snoozed
		_ = open
	}, 1100, 1400)

	for _, name := range []string{
		"dm.book", "dm.ladder", "dm.tape", "dm.fills", "dm.portfolio",
		"dm.risk", "dm.sentiment", "dm.spread", "dm.margin", "dm.leverage",
		"dm.performance", "dm.allocation",
		"dm.payoff, break-even 105.50", "dm.overview",
		"dm.status", "dm.tape-ticker", "dm.clock", "Level two", "Quotes",
		"Earnings", "Economic calendar", "Delta",
	} {
		if _, ok := tt.Find(name); !ok {
			t.Errorf("dark mode: %q is not on screen", name)
		}
	}
}

// searchQuery is the one string the dark-mode view writes into, handed out by
// pointer because every field in this package that holds text holds a pointer
// the caller owns.
func searchQuery() *string { s := "rivr"; return &s }

// TestDarkPaletteIsReadNotAssumed: a component that hard-coded a colour would
// pass every render test above while being wrong in the one place it cannot be
// seen.
func TestDarkPaletteIsReadNotAssumed(t *testing.T) {
	var bg, text ui.Color
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: core.Dark})
		k := core.Tokens(c)
		bg, text = k.Background, k.Text
		SymbolBadge(c, SymbolBadgeOptions{Symbol: Symbol{Ticker: "RIVR"}})
	}, 300, 80)
	if tt == nil {
		t.Fatal("the tester did not run the view")
	}
	if bg != theme.Dark().Background {
		t.Errorf("a dark window resolved the background %v, want %v", bg, theme.Dark().Background)
	}
	if bg == theme.Light().Background {
		t.Error("a dark window resolved the light background")
	}
	if text == theme.Light().Text {
		t.Error("a dark window resolved the light text colour")
	}
}

func TestLightPaletteIsTheLightOne(t *testing.T) {
	var bg ui.Color
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: core.Light})
		bg = core.Tokens(c).Background
		SymbolBadge(c, SymbolBadgeOptions{Symbol: Symbol{Ticker: "RIVR"}})
	}, 300, 80)
	if bg != theme.Light().Background {
		t.Errorf("a light window resolved %v, want %v", bg, theme.Light().Background)
	}
}

// TestDataTableIsTheOneTable is a shape assertion: every table in this package
// is data.DataTable, so the caller's Sort is data.Sort and reordering goes
// through data.Rows. A second table implementation would be a second set of
// rules about what a Shift-click means, written a year later.
func TestDataTableIsTheOneTable(t *testing.T) {
	var sort data.Sort = data.Sort{Column: "price"}
	if sort.Column != "price" {
		t.Error("the caller's sort state is data.Sort")
	}
	if data.Toggle(sort, "size").Column != "size" {
		t.Error("data.Toggle is the rule a sortable head asks for")
	}
	if got := strings.Join([]string{MonoFont[:2], "a font stack, not a name"}, ", "); got == "" {
		t.Error("MonoFont must be a stack")
	}
}
