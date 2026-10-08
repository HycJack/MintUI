package finance

import (
	"strconv"
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/chart"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/data"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// Point is one sample of a price's history: its time and its value. It is
// [chart.Point] aliased rather than declared, so a sparkline fed the same
// points as a full chart cannot be given a subtly different shape.
type Point = chart.Point

// spreadRow is a row of figures with the room left over between them pushed
// apart. It is here because four of this package's cards want exactly this
// and a row with a Grow(1) between two pairs of figures is three lines each
// time.
func spreadRow(c *ui.Context, u float32, children func()) {
	ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.25).Children(children)
}

// lowerFold and containsFold are the case-insensitive match every search in
// this package uses. They are spelled out once because a search that is
// case-sensitive in one screen and not in the next is a search that finds
// things only when they typed them correctly.
func lowerFold(s string) string { return strings.ToLower(s) }

func containsFold(haystack, needle string) bool {
	return needle == "" || strings.Contains(strings.ToLower(haystack), needle)
}

// Symbol is an instrument's own words: what it is called, what it trades in
// and where.
//
// It is a struct rather than a string because every screen here has to say
// more than the ticker. A quote card with just "RIVR" on it is a card that
// cannot be told from four hundred others in a grid, and the exchange is the
// thing that makes a ticker unique — the same three letters trade in two
// places.
type Symbol struct {
	// Ticker is what it is called, and is required: a symbol with no ticker
	// cannot be looked up, and every screen here is a way of looking one up.
	Ticker string
	// Exchange is where it trades; empty leaves it out rather than drawing a
	// separator before nothing.
	Exchange string
	// Name is the full name, for a card and for a reader who does not know
	// the ticker.
	Name string
}

// Full is the ticker and the exchange, the two together, and is what a
// searchable field matches on: "rivr/l" finds the instrument and "rivr" alone
// finds two of them.
func (s Symbol) Full() string {
	if s.Exchange == "" {
		return s.Ticker
	}
	return s.Ticker + "/" + s.Exchange
}

// Label is what the symbol is called out loud, and is never empty for the
// same reason the ticker is required.
func (s Symbol) Label() string {
	if s.Exchange == "" {
		return s.Ticker
	}
	return s.Ticker + " on " + s.Exchange
}

// ── the badge ──────────────────────────────────────────────────────────────

// SymbolBadgeOptions configure a SymbolBadge.
type SymbolBadgeOptions struct {
	// Symbol is the instrument. Its ticker is required.
	Symbol Symbol
	// Size scales the badge; zero takes the library's own.
	Size float32
	// ShowExchange puts the exchange beside the ticker, which is what a
	// sidebar needs and what a dense grid does not have room for.
	ShowExchange bool
}

// SymbolBadge is the instrument's ticker in a pill.
//
// It is a pill rather than plain text because a ticker is looked up rather
// than read: a reader scanning a grid is matching shapes, and a pill around
// each one gives them something to match against. Plain tickers in a column of
// names have to be read one at a time.
func SymbolBadge(c *ui.Context, opts SymbolBadgeOptions) *ui.Element {
	if opts.Symbol.Ticker == "" {
		panic("finance: SymbolBadge needs a Ticker; an instrument with no ticker cannot be " +
			"looked up")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	size := opts.Size
	if size <= 0 {
		size = theme.CaptionSize
	}
	ticker := opts.Symbol.Ticker

	return ui.Box(c).Padding(u*0.75, u*1.5).Radius(theme.PillRadius).
		Background(k.Surface).AlignItems(ui.Center).Gap(u * 0.75).
		Label(opts.Symbol.Label()).Children(func() {
		mono(c, ticker, size, k.Text).Bold()
		if opts.ShowExchange && opts.Symbol.Exchange != "" {
			ui.Text(c, opts.Symbol.Exchange).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
		}
	})
}

// ── the price ──────────────────────────────────────────────────────────────

// PriceTextOptions configure a PriceText.
type PriceTextOptions struct {
	// Value is the price.
	Value float64
	// Unit is what it is priced in; drawn as a prefix. Empty draws none, which
	// is what a column whose head already says "Price (USD)" wants.
	Unit string
	// Size is the type size; zero takes the card's figure size.
	Size float32
	// Decimals overrides [PriceDecimals] for an instrument that quotes in
	// fractions of a cent — a bond, a spread, a yield. Zero takes the
	// library's two.
	Decimals int
	// Bold marks it as the figure on a card.
	Bold bool
	// Label names it for assistive technology. Empty builds one from the unit,
	// and a bare number announced as "number" tells a reader nothing.
	Label string
}

// PriceText is one price, in the monospaced face.
//
// It is its own component because the face is the whole of it. Every other
// string in this library is drawn in the proportional face, where a "1" and an
// "8" are different widths and a right-aligned column of prices lines up at
// the right edge instead of at the decimal point. Monospaced digits put both
// edges on one grid, which is what makes a column scannable.
func PriceText(c *ui.Context, opts PriceTextOptions) *ui.Element {
	k := core.Tokens(c)
	text := FormatPrice(opts.Value)
	if opts.Decimals > 0 {
		text = formatFixed(opts.Value, opts.Decimals)
	}
	label := opts.Label
	if label == "" {
		label = "Price"
		if opts.Unit != "" {
			label = "Price in " + opts.Unit
		}
	}
	size := opts.Size
	if size <= 0 {
		size = theme.StatSize
	}

	row := ui.Row(c).AlignItems(ui.End).Label(label).Children(func() {
		if opts.Unit != "" {
			ui.Text(c, opts.Unit).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
		}
		mono(c, text, size, k.Text)
	})
	if opts.Bold {
		return row.FontWeight(700)
	}
	return row
}

// formatFixed is a price written at a precision other than the library's two.
// [FormatPrice] stays the two-place function because that is the column rule;
// this is for the one figure on a card that quotes in something else — a bond
// in thirty-seconds, a yield, a spread.
func formatFixed(v float64, decimals int) string {
	return group(strconv.FormatFloat(v, 'f', decimals, 64))
}

// ── the change ─────────────────────────────────────────────────────────────

// PriceChangeBadgeOptions configure a PriceChangeBadge.
type PriceChangeBadgeOptions struct {
	// Now is the current price and Previous the one it is being compared
	// with. Both are the caller's: a badge that fetched its own history would
	// be showing a change against a close the caller has already adjusted
	// for a split.
	Now, Previous float64
	// Absolute is the change as an amount as well as a percentage, which a
	// badge wide enough for both shows. It is zero whenever the percentage is
	// zero, because "0" and "+0.00%" beside each other is the lie the
	// percentage rule exists to prevent.
	Absolute float64
	// ShowAbsolute puts the amount in front of the percentage.
	ShowAbsolute bool
	// Tone overrides the severity [FormatChange] chose, for a screen with its
	// own colour convention. The default — zero, Neutral — is not an
	// override: it means "use what the number says".
	Tone core.Severity
	// Size is the type size; zero takes the badge's own.
	Size float32
}

// PriceChangeBadge is a price's move: the percentage, in the tone that
// percentage implies.
//
// The tone is not decoration and is not the caller's: it comes from
// [FormatChange], which decides it from the numbers. A component that took a
// colour would let a row be painted green while its own text read "-1.20%",
// and the two would be found at different times by different readers — a
// screen reader user hears the minus, everyone else sees the green.
func PriceChangeBadge(c *ui.Context, opts PriceChangeBadgeOptions) *ui.Element {
	u := core.Density(c).Unit()
	k := core.Tokens(c)
	text, sev := FormatChange(opts.Now, opts.Previous)
	if opts.Tone != core.Neutral {
		sev = opts.Tone
	}
	bg, fg := sev.Pair(k)
	label := "Change " + text
	if opts.ShowAbsolute {
		label = "Change " + FormatSigned(opts.Absolute) + ", " + text
	}

	size := opts.Size
	if size <= 0 {
		size = theme.CaptionSize
	}
	return ui.Box(c).Padding(u*0.75, u*1.5).Radius(theme.PillRadius).
		Background(bg).Label(label).Shrink(0).Children(func() {
		if opts.ShowAbsolute && opts.Absolute != 0 {
			mono(c, FormatSigned(opts.Absolute), size, fg)
			ui.Text(c, " ").TextColor(fg)
		}
		mono(c, text, size, fg).Bold()
	})
}

// ── bid and ask ────────────────────────────────────────────────────────────

// BidAskBarOptions configure a BidAskBar.
type BidAskBarOptions struct {
	// Bid is what buyers will pay and Ask what sellers will take. Both are the
	// caller's; a book the component fetched would be a book that disagrees
	// with the table beside it.
	Bid, Ask float64
	// Size scales the bar; zero takes one density unit.
	Size float32
	// Label names the bar; empty builds one from the two prices.
	Label string
}

// BidAskBar is the two prices and the gap between them, as a bar.
//
// The bar is drawn with the bid filling it from the left and the ask from the
// right, and the middle is the spread — so the *gap is the thing being
// measured*. A bar that showed the two prices as one scale would make a wide
// spread and a narrow one look the same width, which is the one thing a reader
// looking at a bid/ask bar is asking about.
func BidAskBar(c *ui.Context, opts BidAskBarOptions) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if opts.Bid <= 0 || opts.Ask <= 0 {
		// A zero end is not a price. It is a book that has not arrived yet, and
		// drawing a bar from 0 to 52.40 would say the spread is the whole
		// price — which is exactly the number a reader must not be shown
		// while the other side is still loading.
		panic("finance: BidAskBar needs a positive bid and a positive ask; a bar with one " +
			"end is a line and says nothing about the gap")
	}
	gap, _, ok := Spread(
		[]Price{{Price: opts.Bid, Size: 1}},
		[]Price{{Price: opts.Ask, Size: 1}},
	)
	if !ok {
		panic("finance: BidAskBar needs a bid and an ask; a bar with one end is a line " +
			"and says nothing about the gap")
	}
	size := opts.Size
	if size <= 0 {
		size = u
	}
	label := opts.Label
	if label == "" {
		label = "Bid " + FormatPrice(opts.Bid) + ", ask " + FormatPrice(opts.Ask)
	}

	// The spread is drawn as a share of the midpoint, clamped: a wide or
	// crossed spread must not fill the whole bar and hide the two prices it is
	// drawn between.
	share := float32(0.5)
	if mid := (opts.Bid + opts.Ask) / 2; mid > 0 && gap > 0 {
		share = float32(gap / mid / 2)
		if share > 0.5 {
			share = 0.5
		}
	}
	if share < 0 {
		share = 0
	}

	return ui.Column(c).FillWidth().Gap(u).Label(label).Children(func() {
		ui.Row(c).FillWidth().Gap(u * 1.25).Children(func() {
			ui.Column(c).Grow(1).Label("Bid " + FormatPrice(opts.Bid)).Children(func() {
				ui.Text(c, "Bid").TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize))
				mono(c, FormatPrice(opts.Bid), theme.StatSize, k.Success)
			})
			ui.Column(c).Grow(1).AlignItems(ui.End).Label("Ask " + FormatPrice(opts.Ask)).
				Children(func() {
					ui.Text(c, "Ask").TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.CaptionSize))
					mono(c, FormatPrice(opts.Ask), theme.StatSize, k.Danger)
				})
		})
		ui.Box(c).FillWidth().Height(size).Shrink(0).Radius(size / 2).
			Role(ui.RoleNone).Draw(func(p *ui.Painter, r ui.Rect) {
			p.Fill(ui.Rect{X: r.X, Y: r.Y, W: r.W * (0.5 - share), H: r.H}, k.Success, size/2)
			p.Fill(ui.Rect{X: r.X + r.W*(0.5+share), Y: r.Y, W: r.W * (0.5 - share), H: r.H},
				k.Danger, size/2)
		})
	})
}

// ── the quote card ─────────────────────────────────────────────────────────

// QuoteOptions configure a QuoteCard.
type QuoteOptions struct {
	// Symbol is the instrument, Now its price and Previous the one it is
	// compared with.
	Symbol   Symbol
	Now      float64
	Previous float64
	// DayHigh and DayLow are the extremes of the session, and DayVolume how
	// much traded. Any of them zero is left out rather than written as a zero,
	// which would say the instrument has not traded at all.
	DayHigh, DayLow float64
	DayVolume       int
	// Currency is the unit, drawn beside the price and in the figures' heads.
	Currency string
	// Delayed says the price is not live, which is written beside it. It is
	// the caller's flag because only the caller knows what the feed is.
	Delayed bool
	// Interval is the period the change is over — "1D", "1W", "YTD" — and is
	// drawn above the change.
	Interval string
	// Selected marks the card in a grid, and is the caller's flag: two views
	// of one watchlist must agree about which instrument is current.
	Selected bool
	// Sparkline is the recent price as series points, drawn under the figures.
	// Empty draws no sparkline rather than a flat one, which would be a
	// claim about the day.
	Sparkline []Point
}

// QuoteCard is one instrument's headline: its symbol, its price, its move and
// the shape of the last while.
//
// The price is the largest thing on the card and the change sits under it,
// never beside it. A quote card is glanced at, and the price is what it is
// glanced at for; the move is what it is *read* for, and it is read second.
func QuoteCard(c *ui.Context, opts QuoteOptions) *ui.Element {
	if opts.Symbol.Ticker == "" {
		panic("finance: QuoteCard needs a Symbol; a card about nothing in particular is not a " +
			"quote")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	_, sev := FormatChange(opts.Now, opts.Previous)

	return layout.Container(c, layout.ContainerOptions{
		Surface: true, Border: true, Radius: theme.CardRadius,
		Pad: u * 2, Gap: u * 1.5,
	}, func() {
		spreadRow(c, u, func() {
			SymbolBadge(c, SymbolBadgeOptions{Symbol: opts.Symbol, ShowExchange: true})
			ui.Box(c).Grow(1)
			if opts.Interval != "" {
				ui.Text(c, opts.Interval).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
			}
		})
		if opts.Symbol.Name != "" {
			ui.Text(c, opts.Symbol.Name).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize)).MaxLines(1)
		}

		PriceText(c, PriceTextOptions{
			Value: opts.Now, Unit: opts.Currency, Size: theme.TitleSize, Bold: true,
			Label: "Price of " + opts.Symbol.Label(),
		})

		spreadRow(c, u, func() {
			PriceChangeBadge(c, PriceChangeBadgeOptions{
				Now: opts.Now, Previous: opts.Previous,
				Absolute: opts.Now - opts.Previous, ShowAbsolute: true,
			})
			if opts.Delayed {
				ui.Text(c, "Delayed").TextColor(k.Warning).
					FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
			}
		})

		if len(opts.Sparkline) > 1 {
			ui.Box(c).FillWidth().Height(u * 5).Shrink(0).
				Label("Recent price for " + opts.Symbol.Label()).Draw(func(p *ui.Painter, r ui.Rect) {
				sparkline(c, p, r, opts.Sparkline, toneInk(sev, k))
			})
		}

		if opts.DayHigh > 0 || opts.DayVolume > 0 {
			spreadRow(c, u, func() {
				if opts.DayLow > 0 {
					figure(c, stat{Value: FormatPrice(opts.DayLow), Label: "Day low", Mono: true})
				}
				if opts.DayHigh > 0 {
					figure(c, stat{Value: FormatPrice(opts.DayHigh), Label: "Day high", Mono: true})
				}
				if opts.DayVolume > 0 {
					figure(c, stat{Value: FormatVolume(opts.DayVolume), Label: "Volume", Mono: true})
				}
			})
		}
	})
}

// toneInk is the ink a severity is written in, out of the ramp rather than off
// it: a card's change is coloured by what the number said and not by what the
// caller thought.
//
// Neutral resolves to the muted tone rather than to the body tone: a figure
// with nothing to say about it is furniture, and painting it in the body tone
// would make it compete with the ones that do.
func toneInk(sev core.Severity, k theme.Tokens) ui.Color {
	if sev == core.Neutral {
		return k.TextMuted
	}
	_, fg := sev.Pair(k)
	return fg
}

// ── the sparkline under a figure ───────────────────────────────────────────

// sparkline is a series drawn as a shape with no axes, no labels and no
// frame — the one chart there is room for inside a card.
//
// It is painted here rather than built from [chart] because ui/chart's
// sparkline belongs to a frame that measures gutters, and a card's sparkline
// has a box and nothing else: there is no axis on it to give room to. The
// shape is what the reader takes from it, and the exact last value is in the
// price above.
func sparkline(c *ui.Context, p *ui.Painter, r ui.Rect, points []Point, col ui.Color) {
	if len(points) < 2 || r.W <= 0 || r.H <= 0 {
		return
	}
	lo, hi := points[0].Y, points[0].Y
	for _, pt := range points[1:] {
		lo = minf(lo, pt.Y)
		hi = maxf(hi, pt.Y)
	}
	span := hi - lo
	if span <= 0 {
		// A flat series is a line down the middle, not an empty box: "the
		// price did not move" is a fact and an empty rectangle is not one.
		span = 1
		lo = points[0].Y - 0.5
	}
	step := r.W / float32(len(points)-1)
	at := func(i int) (float32, float32) {
		y := r.Y + r.H - float32((points[i].Y-lo)/span)*r.H
		return r.X + step*float32(i), y
	}
	var path ui.Path
	x, y := at(0)
	path.MoveTo(x, y)
	for i := 1; i < len(points); i++ {
		x, y = at(i)
		path.LineTo(x, y)
	}
	p.StrokePath(&path, 2, col)

	// The last point gets a dot, because that is the one the price above it
	// is and the reader wants to see them agree.
	x, y = at(len(points) - 1)
	internal.Dot(p, x, y, r.H*0.12, col)
}

func minf(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func maxf(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// ── search ─────────────────────────────────────────────────────────────────

// Instrument is something the caller can find: a symbol and enough words to
// match a query against.
type Instrument struct {
	Symbol Symbol
	// Keywords are the other words that should find it — the sector, the
	// exchange's own description, the name a reader would type. They are not
	// shown; they are only matched, which is why they are separate from the
	// symbol's own name.
	Keywords []string
}

// SymbolSearchOptions configure a SymbolSearch.
type SymbolSearchOptions struct {
	// Query is the caller's string; the field writes it.
	Query *string
	// Instruments are the caller's, and they are all of them: this component
	// does not fetch and does not know where a list of instruments comes from.
	Instruments []Instrument
	// Results is how many matches to show, zero for eight. A search that
	// showed every match is a list, and a list under a search field is a
	// table the reader has to scroll instead of a shortlist.
	Results int
	// Selected is the instrument chosen, as an index into Instruments; -1 for
	// none.
	Selected *int
	// Highlight puts the chosen instrument's row in the surface's hover step,
	// so the keyboard's selection and the pointer's agree.
	Highlight bool
}

// SymbolSearchResult carries a SymbolSearch.
type SymbolSearchResult struct {
	// Element is the field and the results under it.
	Element *ui.Element
	// chosen is the ticker picked this frame, or "".
	chosen string
}

// Chosen is the instrument the reader picked, by ticker. It is reported rather
// than written because the instrument behind a ticker is the caller's — this
// component has a list of what it matched and nothing else, and a ticker is
// not enough to trade.
func (r SymbolSearchResult) Chosen() string { return r.chosen }

// SymbolSearch is a field and the instruments it finds.
//
// It matches on the ticker, the full symbol, the name and the keywords, all
// case-insensitively. The keywords are the part that matters: a reader looking
// for a roofing contractor types "roof" and not "RIVR", and a search that only
// knew tickers would find nothing for them.
func SymbolSearch(c *ui.Context, opts SymbolSearchOptions) SymbolSearchResult {
	if opts.Query == nil {
		panic("finance: SymbolSearch needs a Query to point at; it keeps no text of its own")
	}
	if opts.Selected == nil {
		panic("finance: SymbolSearch needs a Selected instrument to point at; it owns no " +
			"list of its own")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	limit := opts.Results
	if limit <= 0 {
		limit = 8
	}

	matches := matchInstruments(opts.Instruments, opts.Query)
	shown := matches
	if len(shown) > limit {
		shown = shown[:limit]
	}

	var res SymbolSearchResult
	res.Element = ui.Column(c).FillWidth().Gap(u).Label("Find an instrument").Children(func() {
		input.SearchInput(c, opts.Query, input.SearchInputOptions{
			Label: "Symbol or company", Placeholder: "Symbol or company",
		})
		if len(matches) == 0 {
			// Said plainly and briefly: a result list with nothing in it is
			// the whole of the answer, and a long sentence here would be a
			// panel opening on every keystroke that matched nothing.
			ui.Text(c, "Nothing matches that").TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.BodySize))
			return
		}
		for _, i := range shown {
			in := opts.Instruments[i]
			index := i
			row := ui.Box(c).FillWidth().Padding(u*0.75, u*1.5).Radius(theme.SmallRadius).
				Label(in.Symbol.Label()).Role(ui.RoleNone).Cursor(ui.CursorPointer)
			if opts.Highlight && *opts.Selected == index {
				row.Background(k.SurfaceHover)
			}
			if row.Hovered() {
				row.Background(k.SurfaceHover)
			}
			if row.Clicked() {
				*opts.Selected = index
				res.chosen = in.Symbol.Ticker
			}
			row.Children(func() {
				SymbolBadge(c, SymbolBadgeOptions{Symbol: in.Symbol})
				if in.Symbol.Name != "" {
					ui.Text(c, in.Symbol.Name).TextColor(k.TextMuted).Grow(1).MaxLines(1).
						FontSize(core.FontSize(c, theme.RowSize))
				}
				ui.Box(c).Grow(1)
				if in.Symbol.Exchange != "" {
					ui.Text(c, in.Symbol.Exchange).TextColor(k.TextFaint).
						FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
				}
			})
		}
		if len(matches) > len(shown) {
			// The count of what is not shown, rather than nothing: a reader
			// who knows there are more will type more, and one who is told
			// nothing will assume the list is complete.
			ui.Text(c, internal.Plural(len(matches)-len(shown), "more match", "more matches")).
				TextColor(k.TextFaint).FontSize(core.FontSize(c, theme.CaptionSize))
		}
	})
	return res
}

// matchInstruments is the search, in the order the words are worth matching:
// a ticker first, then the full symbol, then the name, then the keywords.
func matchInstruments(all []Instrument, query *string) []int {
	if query == nil || *query == "" {
		return rangeIndexes(len(all))
	}
	needle := lowerFold(*query)
	var out []int
	for i, in := range all {
		if containsFold(in.Symbol.Ticker, needle) ||
			containsFold(in.Symbol.Full(), needle) ||
			containsFold(in.Symbol.Name, needle) {
			out = append(out, i)
			continue
		}
		for _, word := range in.Keywords {
			if containsFold(word, needle) {
				out = append(out, i)
				break
			}
		}
	}
	return out
}

func rangeIndexes(n int) []int {
	out := make([]int, n)
	for i := range out {
		out[i] = i
	}
	return out
}

// ── the watchlist ──────────────────────────────────────────────────────────

// WatchQuote is one row of a watchlist: an instrument and its price and move.
type WatchQuote struct {
	Symbol   Symbol
	Now      float64
	Previous float64
	// Volume is the day's, left out of the row when zero.
	Volume int
	// Group is what the reader filed it under; a watchlist grouped by sector
	// is a different screen from a flat one and the same rows.
	Group string
}

// WatchlistOptions configure a Watchlist.
type WatchlistOptions struct {
	// Quotes are the caller's, in the order they should be shown.
	Quotes []WatchQuote
	// Selected is the row the keys move from, as an index into Quotes; -1 for
	// none.
	Selected *int
	// Height is the list's height, and is required: a watchlist with no height
	// grows to fit every row, which for a hundred instruments is a page
	// nobody can scroll past the first screen of.
	Height float32
	// State and Scroll keep the list's place between frames.
	State  *ui.ListState
	Scroll *ui.ScrollState
	// Grouped draws a rule and a heading between the groups, which is what a
	// watchlist sorted by sector wants.
	Grouped bool
	// Empty draws instead of the rows when there are none.
	Empty func()
}

// Watchlist is a column of instruments and their prices.
//
// It is [data.List] rather than a table because there are two things per row
// and both of them are fixed: a badge on the left and a price on the right. A
// table earns its rules and its head when there are four or more things to
// line up, and a watchlist has none of that — it is a list of rows whose price
// is on the right edge.
func Watchlist(c *ui.Context, opts WatchlistOptions) *ui.Element {
	if opts.Height <= 0 {
		panic("finance: Watchlist needs a Height; a list with no height grows to fit every " +
			"row rather than scrolling")
	}
	if opts.Selected == nil {
		panic("finance: Watchlist needs a Selected row to point at; it owns no list of its own")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	return dataList(c, len(opts.Quotes), opts.Height, opts.Selected, opts.State, opts.Scroll,
		"Watchlist", func(row int) {
			q := opts.Quotes[row]
			_, sev := FormatChange(q.Now, q.Previous)
			line := ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u*1.25).
				Padding(u*0.75, u*1.5).Radius(theme.SmallRadius).
				Label(q.Symbol.Label()).Role(ui.RoleNone)
			if *opts.Selected == row {
				line.Background(k.SurfaceHover)
			}
			line.Children(func() {
				SymbolBadge(c, SymbolBadgeOptions{Symbol: q.Symbol})
				if opts.Grouped && row > 0 && opts.Quotes[row-1].Group != q.Group {
					ui.Text(c, q.Group).TextColor(k.TextFaint).
						FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
				}
				ui.Box(c).Grow(1)
				if q.Volume > 0 {
					ui.Text(c, FormatVolume(q.Volume)).TextColor(k.TextFaint).
						FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
				}
				ui.Column(c).AlignItems(ui.End).Shrink(0).
					Label(q.Symbol.Label() + ", " + FormatPrice(q.Now)).Children(func() {
					mono(c, FormatPrice(q.Now), theme.RowSize, k.Text)
					changeText(c, q.Now, q.Previous)
				})
			})
			_ = sev
		}, opts.Empty)
}

// changeText is the percentage under a price in a row, in the tone the
// percentage implies.
func changeText(c *ui.Context, now, prev float64) *ui.Element {
	k := core.Tokens(c)
	text, sev := FormatChange(now, prev)
	ink := k.TextFaint
	if sev != core.Neutral {
		ink = toneInk(sev, k)
	}
	return mono(c, text, theme.CaptionSize, ink)
}

// dataList is the shared row list, so that a watchlist and a quote list cannot
// each have their own idea of what a row is. It is data.List, wrapped in a
// named box because data.List names its rows and nothing else — and a column
// of rows whose every row announces its own name and which has no name of its
// own reads as a set of unrelated rows.
func dataList(c *ui.Context, n int, height float32, selected *int,
	state *ui.ListState, scroll *ui.ScrollState, name string,
	row func(i int), empty func()) *ui.Element {
	return ui.Box(c).FillWidth().Label(name).Role(ui.RoleList).Children(func() {
		data.List(c, data.ListOptions{
			Rows: n, Height: height, Selected: selected,
			State: state, Scroll: scroll, Empty: empty,
			Key:   func(i int) any { return i },
			Label: func(i int) string { return name },
		}, row)
	})
}
