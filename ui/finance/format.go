package finance

import (
	"math"
	"strconv"
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// ── prices ─────────────────────────────────────────────────────────────────

// PriceDecimals is how many places a price is written to, and the reason
// [FormatPrice] has no parameter for it.
//
// Two, everywhere, on purpose. One place is how every currency in the world
// quotes most instruments, and three is what a reader has to skip past before
// they reach the figure they are looking at. More importantly, it is the same
// number for every row: a column whose figures have different widths is a
// column the eye has to re-measure, and a quote table is read by scanning down
// a column rather than by reading across it.
const PriceDecimals = 2

// FormatPrice writes a price as the digits, two places and a thousands
// separator: 1234.5 becomes "1,234.50".
//
// No currency symbol. A price symbol is the *column's* business, not the
// figure's: a table whose every cell said "$" would be a table of $1,234.50
// and $1,234.50 next to a column of sizes reading 1,234.50 and 1,234.50, and
// the reader would have to work out which was which. The unit goes in the
// column head, where it is stated once.
//
// Negative prices get a leading minus and get their grouping before it: -1234.5
// is "-1,234.50", not "-1,234.50" with the separator on the wrong side of the
// sign. There is no such thing as a negative price in most markets, but there
// is in futures and there is in a currency pair's cross rate, and a formatter
// that cannot write one is a formatter that will print something wrong the
// first time somebody asks for a spread.
func FormatPrice(v float64) string {
	neg := v < 0
	if neg {
		v = -v
	}
	s := strconv.FormatFloat(v, 'f', PriceDecimals, 64)
	if neg {
		s = "-" + s
	}
	return group(s)
}

// group puts a thousands separator into the whole part of a written number,
// leaving the decimal point and its places alone. It works on the written
// string rather than on the float because it has to: a float64 has no idea
// where its decimal point is until it has been written out with a precision,
// and doing the grouping before the formatting would put separators inside
// the digits the rounding is about to change.
func group(s string) string {
	dot := strings.IndexByte(s, '.')
	whole, frac := s, ""
	if dot >= 0 {
		whole, frac = s[:dot], s[dot:]
	}
	neg := strings.HasPrefix(whole, "-")
	if neg {
		whole = whole[1:]
	}
	// Three digits or fewer has nothing to separate. The sign does not count
	// for that: "-99" is no longer than "999" is wide, and a separator in
	// "-1,234" but not in "-999" would be the right answer, not the one that
	// makes the pair line up.
	if len(whole) <= 3 {
		return s
	}
	var b strings.Builder
	if neg {
		b.WriteByte('-')
	}
	head := len(whole) % 3
	if head == 0 {
		head = 3
	}
	b.WriteString(whole[:head])
	for i := head; i < len(whole); i += 3 {
		b.WriteByte(',')
		b.WriteString(whole[i : i+3])
	}
	b.WriteString(frac)
	return b.String()
}

// FormatCompact is a big figure written short, for a card's headline where the
// exact number is in a tooltip: 1,234,500 becomes "1.23M".
//
// It is *not* [FormatVolume]: that one is a share count, which has its own
// carry rules and its own tests, and a share count is a count of things rather
// than an amount of money. Two functions because the boundaries differ.
func FormatCompact(v float64) string {
	abs := math.Abs(v)
	switch {
	case abs >= 1e9:
		return trim(v/1e9) + "B"
	case abs >= 1e6:
		return trim(v/1e6) + "M"
	case abs >= 1e3:
		return trim(v/1e3) + "K"
	}
	return strconv.FormatFloat(v, 'f', -1, 64)
}

// trim is one place and one significant digit after it, written out without
// trailing zeroes: 1.20 becomes "1.2" and 1.00 becomes "1".
func trim(v float64) string {
	s := strconv.FormatFloat(v, 'f', 1, 64)
	if strings.HasSuffix(s, ".0") {
		return s[:len(s)-2]
	}
	return s
}

// ── change ─────────────────────────────────────────────────────────────────

// FormatChange writes a change and the severity it should be drawn in.
//
// Three rules, each of which is a rule about a lie:
//
//   - the sign is always there. A rise is "+1.25%". Writing "1.25%" for a rise
//     makes a green number and a red number the same string, so a reader
//     scanning the column has to see the colour to know the direction — and a
//     reader who cannot see colour cannot read the column at all.
//   - the previous value is what the share is of. "v == prev" is exactly zero
//     change, so it writes "0.00%" and *not* "+0.00%". The two look identical
//     on screen and only one of them is true; a flat instrument is the single
//     most common thing in a portfolio and it must not be dressed up as a
//     move. This is also why the comparison is on the values and not on the
//     rounded percentage: 1.001 against 1.0 rounds to 0.10% and is a real
//     change, and 1.0 against 1.0 is not one however it is written.
//   - the severity comes from the number, not from the caller's colour. A
//     component that took a colour would let a row be drawn as a fall while its
//     own text says it rose, and the two would be found at different times by
//     different readers.
func FormatChange(v, prev float64) (string, core.Severity) {
	if prev == 0 || v == prev {
		return "0.00%", core.Neutral
	}
	pct := (v - prev) / prev * 100
	// Two places, unless the change is real and smaller than two places can
	// say. A rise of 0.0001% rounds to "+0.00%", which is a sign in front of
	// a zero and says the instrument moved when the figure says it did not —
	// the same lie the flat rule above exists to prevent, reached from the
	// other side. Three places is enough for any figure this library can be
	// handed and is still a fixed width, so the column rule survives.
	places := 2
	if math.Abs(pct) < 0.005 {
		places = 4
	}
	s := strconv.FormatFloat(math.Abs(pct), 'f', places, 64)
	if pct > 0 {
		return "+" + s + "%", core.Success
	}
	return "-" + s + "%", core.Danger
}

// FormatSigned writes a signed money figure for a change column: "+$1,240.00"
// against "-$1,240.00", with a zero that has no sign.
//
// The currency symbol is back here, unlike in [FormatPrice], because this
// figure is a *difference* and a column of differences with no unit is a column
// of numbers the reader has to look elsewhere to interpret. A price column has
// a head; a change column is three columns wide and its unit is carried by the
// figure.
func FormatSigned(v float64) string {
	if v == 0 {
		return FormatPrice(0)
	}
	sign := "+"
	if v < 0 {
		sign = ""
	}
	return sign + FormatPrice(v)
}

// ── volume ─────────────────────────────────────────────────────────────────

// FormatVolume writes a count of shares, carrying at each of the three
// boundaries a reader is looking for:
//
//	999     → "999"
//	1,000   → "1.00K"
//	1,500   → "1.50K"
//	999,999 → "1.00M"
//	1,000,000 → "1.00B"
//
// The boundaries are exact on purpose. A volume that reads "1K" at 1,000 and
// "0.99K" at 990 is correct and useless: a reader watching for turnover wants
// to know the moment it crosses a thousand, and "0.99K" hides it two columns
// early. So the carry happens at exactly the thousand, and *below* it the
// number is written out in full, because a count under a thousand has room to
// be written out and nobody counts that far by eye.
//
// Two places above a thousand, so that "1.50K" and "1.52K" are the same width
// and a column of them lines up. The width matters for the same reason the
// price's width does — this is a right-aligned column read by last digit.
func FormatVolume(n int) string {
	neg := n < 0
	if neg {
		n = -n
	}
	if n < 1000 {
		// Below a thousand the count is written out in full. It is short
		// enough to have room, and nobody counts that far by eye — so the
		// figures that a reader is actually comparing are the ones that get
		// room to be exact.
		return sign(neg) + strconv.Itoa(n)
	}
	// A cascade, and the order matters. Each unit is carried up when the
	// figure *written in the unit below it* has reached 1000: 999,999
	// divides to 999.999 thousand, which would print as "1000.00K" — a
	// column reading that has just passed a million and says it has not,
	// which is precisely the boundary this function exists to get right.
	k := float64(n) / 1e3
	m := k / 1e3
	b := m / 1e3
	switch {
	case round2(m) >= 1000:
		return sign(neg) + twoPlaces(b) + "B"
	case round2(k) >= 1000:
		return sign(neg) + twoPlaces(m) + "M"
	}
	return sign(neg) + twoPlaces(k) + "K"
}

func sign(neg bool) string {
	if neg {
		return "-"
	}
	return ""
}

// twoPlaces is a carried figure with exactly two decimal places — one for the
// digit before the point and two after it, which is what makes every value in
// a column the same width: "1.00K" and "9.99K" are both five characters.
func twoPlaces(v float64) string {
	return strconv.FormatFloat(v, 'f', 2, 64)
}

// round2 is a figure to two places, which is the precision a carried count is
// written at.
func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

// ── sizes ──────────────────────────────────────────────────────────────────

// FormatSize writes a position's size: whole numbers as thousands-separated
// and fractions as four places, because a quarter of a share is not a thing
// that happens but three hundredths of one is.
//
// Zero shares writes "0", and not "0.0000": a row with no position in it is
// closed, and four zeroes is the shape of a very small position.
func FormatSize(shares float64) string {
	if shares == math.Trunc(shares) {
		// The sign is written by hand rather than folded into the integer,
		// because a minus after the digits is not a number anybody has read
		// before: "-120" is a short position, "120-" is a typo.
		n := int64(shares)
		if n < 0 {
			return "-" + group(strconv.FormatInt(-n, 10))
		}
		return group(strconv.FormatInt(n, 10))
	}
	return strconv.FormatFloat(shares, 'f', 4, 64)
}

// FormatMoney writes an amount of a currency with its symbol, for a place
// where the unit is not carried by a column head: a portfolio total, a fee, a
// margin requirement.
func FormatMoney(v float64, symbol string) string {
	return symbol + FormatPrice(v)
}

// ── the order book ─────────────────────────────────────────────────────────

// Side is which half of a book a level is on.
type Side int

const (
	// SideBid is what buyers will pay.
	SideBid Side = iota
	// SideAsk is what sellers will take.
	SideAsk
)

// Price is one order on one side: a price and a size.
//
// A struct rather than a bare float because the two numbers are one fact. An
// order book passed as two parallel slices is two slices that can be lined up
// wrongly by one assignment, and a bid's size shown against an ask's price is
// the single most misleading thing a trading screen can draw.
type Price struct {
	// Price is where the order sits.
	Price float64
	// Size is how much of it, in shares. It is negative on the ask side by
	// convention on some venues and is always written positive here.
	Size float64
}

// Level is one row of a book: the price, the size resting at it, and the
// running total from the top of that side down to this row.
type Level struct {
	// Side is which half this row is on.
	Side Side
	// Price is where the orders sit.
	Price float64
	// Size is what rests at this price alone.
	Size float64
	// Total is everything from the best price on this side down to this one.
	//
	// It is the number a depth bar is drawn from, and it is computed rather
	// than left to the reader: a bar that showed each row's own size would be
	// a chart of how the book is distributed, and the question a depth chart
	// answers is "how much would I have to take to get through this", which
	// is a running sum.
	Total float64
	// Orders is how many orders are at this price, when the caller knows it.
	// Zero means "not told", not "none", because a book that does not carry
	// the count is most books.
	Orders int
}

// Depth is a whole book: the bids and the asks, best first on each side.
type Depth struct {
	Bids []Level
	Asks []Level
}

// MaxSize is the largest single level on either side of a book, which is what
// the depth bars are scaled against. Zero for an empty book.
func (d Depth) MaxSize() float64 {
	max := 0.0
	for _, l := range append(append([]Level{}, d.Bids...), d.Asks...) {
		max = math.Max(max, l.Size)
	}
	return max
}

// Levels is one side of a book, with the running total computed.
//
// The prices are sorted descending on the bids and ascending on the asks, so
// that the first row of each is the best price — the one a market order would
// hit first — whatever order the caller handed them in. Sorting here rather
// than asking the caller to is deliberate: an order book that is not in price
// order is not an order book, and a component that trusted its input would
// show a book whose top row is not the top of the market.
//
// The total runs from the best price down, which is the direction a reader
// takes it in: to buy, they walk the asks from the best upward and accumulate
// what each level would cost them.
func Levels(bids, asks []Price) []Level {
	out := make([]Level, 0, len(bids)+len(asks))

	b := append([]Price(nil), bids...)
	sortPrices(b, SideBid)
	total := 0.0
	for _, p := range b {
		total += p.Size
		out = append(out, Level{
			Side: SideBid, Price: p.Price, Size: p.Size, Total: total,
		})
	}

	a := append([]Price(nil), asks...)
	sortPrices(a, SideAsk)
	total = 0.0
	for _, p := range a {
		total += p.Size
		out = append(out, Level{
			Side: SideAsk, Price: p.Price, Size: p.Size, Total: total,
		})
	}
	return out
}

// sortPrices puts one side of a book in the order a reader takes it in: best
// price first, bids descending and asks ascending.
func sortPrices(ps []Price, side Side) {
	// An insertion sort, written out rather than reached for: a book side is
	// tens of levels, the code is four lines, and sort.Slice's closure would
	// be longer and would allocate a reflect-backed swapper for it.
	for i := 1; i < len(ps); i++ {
		for j := i; j > 0 && before(ps[j], ps[j-1], side); j-- {
			ps[j], ps[j-1] = ps[j-1], ps[j]
		}
	}
}

// before reports whether a comes before b in the order of one side.
func before(a, b Price, side Side) bool {
	if side == SideBid {
		return a.Price > b.Price
	}
	return a.Price < b.Price
}

// BestBid and BestAsk are the top of each side, and false for an empty one.
// They are functions rather than indexing Levels because the best price is a
// question a spread indicator asks every frame and an index into a slice it
// has to build first is the wrong way round.
func BestBid(bids []Price) (float64, bool) {
	best, seen := 0.0, false
	for _, p := range bids {
		if !seen || p.Price > best {
			best, seen = p.Price, true
		}
	}
	return best, seen
}

func BestAsk(asks []Price) (float64, bool) {
	best, seen := 0.0, false
	for _, p := range asks {
		if !seen || p.Price < best {
			best, seen = p.Price, true
		}
	}
	return best, seen
}

// Spread is the gap between the best bid and the best ask, and the share of it
// as a fraction of the midpoint — the two numbers a spread indicator shows.
//
// A crossed book, where the bid is above the ask, is a negative spread and is
// reported as one rather than being clamped to zero: a crossed book is a real
// thing that happens and it means something specific, and a spread indicator
// that showed "0.00" for it would be hiding the only interesting thing on the
// screen.
func Spread(bids, asks []Price) (gap, fraction float64, ok bool) {
	bid, haveBid := BestBid(bids)
	ask, haveAsk := BestAsk(asks)
	if !haveBid || !haveAsk {
		return 0, 0, false
	}
	gap = ask - bid
	mid := (ask + bid) / 2
	if mid == 0 {
		return gap, 0, true
	}
	return gap, gap / mid, true
}

// ── small shared pieces ────────────────────────────────────────────────────

// stat is one figure in a row of figures: a value, a label and a tone.
type stat struct {
	Value string
	Label string
	Unit  string
	Tone  core.Severity
	// Mono draws the value in the monospaced face. True for every number that
	// sits in a column, false for the ones that stand alone.
	Mono bool
}

// quoteBand is the band a headline figure sits on: a surface a step below the
// window, for a number that would otherwise be a line of text among lines of
// text. It is a function rather than a Container call at each site because
// three components in this package want exactly this and a container options
// struct spelled three ways is three chances to spell it differently.
func quoteBand(c *ui.Context, children func()) *ui.Element {
	u := core.Density(c).Unit()
	return layout.Container(c, layout.ContainerOptions{
		Surface: true, Radius: theme.SmallRadius, Pad: u * 1.5,
	}, children)
}

// itoa is a small count as a string, written out rather than pulled in through
// strconv at each of the nine places in this package that prints one.
func itoa(n int) string { return strconv.Itoa(n) }

// figure is one of those drawn. It is a function rather than a component
// because nothing outside this package has a use for it with a name, and a
// component that is not part of the package's surface is one more thing to
// keep working.
func figure(c *ui.Context, s stat) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	ink := k.Text
	if s.Tone != core.Neutral {
		_, ink = s.Tone.Pair(k)
	}
	col := ui.Column(c).Label(s.Label).Children(func() {
		ui.Row(c).AlignItems(ui.Center).Gap(u * 0.5).Children(func() {
			if s.Mono {
				mono(c, s.Value, theme.StatSize, ink)
			} else {
				ui.Text(c, s.Value).FontSize(core.FontSize(c, theme.StatSize)).
					TextColor(ink).Bold()
			}
			if s.Unit != "" {
				ui.Text(c, s.Unit).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
			}
		})
		ui.Text(c, s.Label).TextColor(k.TextMuted).
			FontSize(core.FontSize(c, theme.CaptionSize)).MaxLines(1)
	})
	return col
}
