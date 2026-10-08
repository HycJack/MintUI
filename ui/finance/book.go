package finance

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/data"
	"github.com/HycJack/MintUI/ui/theme"
)

// ── the book ───────────────────────────────────────────────────────────────

// OrderBookOptions configure an OrderBook.
type OrderBookOptions struct {
	// Bids and Asks are the two sides, the caller's. They are sorted here, so
	// the caller may hand them over in whatever order its feed produced.
	Bids, Asks []Price
	// Rows is how many levels of each side to show, zero for ten. A book with
	// a thousand levels is a data feed, not a screen, and a screen that
	// claims to be showing it while showing ten of them is lying about scale.
	Rows int
	// Selected is the level the keys move from, as an index into the levels
	// this computes; -1 for none.
	Selected *int
	// Height is the book's height. It is required, and the rows are sized from
	// it: a book with no height grows to fit every level, which is every
	// level drawn rather than a book.
	Height float32
	// PriceSize is how many shares are at each level as a share of the largest,
	// which is what the depth bar is scaled against.
	ShowTotals bool
	// Label names the book; it is required, because two columns of numbers
	// with no names on them are a puzzle.
	Label string
}

// OrderBook is the two sides of a market side by side: the bids descending,
// the asks ascending, and how much rests at each price.
//
// The best price is at the *centre*, next to the spread, not at the top. That
// is the whole difference between an order book and two lists: a reader
// looking at a book is looking at what the market will do next, and the price
// it will do it at is the one in the middle. Putting the bids at the top means
// the best bid is the topmost of a long column and is not where the eye goes.
func OrderBook(c *ui.Context, opts OrderBookOptions) *ui.Element {
	if opts.Label == "" {
		panic("finance: OrderBook needs a Label; two columns of numbers with no names on " +
			"them are a puzzle")
	}
	if opts.Height <= 0 {
		panic("finance: OrderBook needs a Height; a book with no height grows to fit every " +
			"level rather than scrolling")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	rows := opts.Rows
	if rows <= 0 {
		rows = 10
	}

	levels := Levels(opts.Bids, opts.Asks)
	bids := takeSide(levels, SideBid, rows)
	asks := takeSide(levels, SideAsk, rows)
	depth := Depth{Bids: bids, Asks: asks}
	maxSize := depth.MaxSize()
	gap, _, spread := Spread(opts.Bids, opts.Asks)

	rowH := (opts.Height - u*8) / float32(rows*2)
	if rowH < u*3 {
		rowH = u * 3
	}

	return ui.Box(c).FillWidth().Label(opts.Label).Role(ui.RoleTable).Height(opts.Height).
		Children(func() {
			ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 2).Children(func() {
				bookHead(c, "Bid", "Shares")
				ui.Box(c).Grow(1)
				bookHead(c, "Ask", "Shares")
			})
			for i := 0; i < rows; i++ {
				ui.Row(c).FillWidth().Height(rowH).Shrink(0).AlignItems(ui.Center).
					Gap(u * 2).Children(func() {
					if i < len(bids) {
						bookRow(c, bids[i], maxSize, ui.End, opts.ShowTotals, u, k)
					} else {
						ui.Box(c).Grow(1)
					}
					ui.Box(c).Grow(1)
					if i < len(asks) {
						bookRow(c, asks[i], maxSize, ui.Start, opts.ShowTotals, u, k)
					} else {
						ui.Box(c).Grow(1)
					}
				})
			}
			spreadRow(c, u, func() {
				if spread {
					// The spread sits between the two best prices, which is the
					// one place in a book that answers "what would it cost me".
					ui.Box(c).Grow(1)
					ui.Text(c, "Spread "+FormatPrice(gap)).
						TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
					ui.Box(c).Grow(1)
				}
			})
		})
}

// bookHead is one side's column names. It is a function rather than two strings
// in the row above so that the two halves cannot disagree about which is which.
func bookHead(c *ui.Context, side, size string) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	return ui.Row(c).Grow(1).AlignItems(ui.Center).Gap(u * 1.5).Children(func() {
		ui.Text(c, side).TextColor(k.TextMuted).
			FontSize(core.FontSize(c, theme.CaptionSize)).Grow(1)
		ui.Text(c, size).TextColor(k.TextMuted).
			FontSize(core.FontSize(c, theme.CaptionSize)).Grow(1)
	})
}

// bookRow is one level: the price, the shares, and a bar behind them showing
// what share of the book's largest level this one is.
//
// The bar is drawn in the side's own tone — the bid's green, the ask's red —
// and it fills from the outside in, so the two sides' bars meet at the spread.
// A reader scanning the middle column sees the two halves of the market
// filling away from each other, which is the shape of a liquid market.
func bookRow(c *ui.Context, l Level, maxSize float64, align ui.Align, totals bool,
	u float32, k theme.Tokens) *ui.Element {
	share := 0.0
	if maxSize > 0 {
		share = l.Size / maxSize
	}
	ink := k.Success
	if l.Side == SideAsk {
		ink = k.Danger
	}
	size := FormatVolume(int(l.Size + 0.5))
	if totals {
		size = FormatVolume(int(l.Total+0.5)) + " tot"
	}

	return ui.Row(c).Grow(1).AlignItems(ui.Center).Gap(u * 1.5).
		Label(levelLabel(l)).Role(ui.RoleNone).
		Draw(func(p *ui.Painter, r ui.Rect) {
			if share <= 0 {
				return
			}
			// Clipped to the row's own box rather than to a column's, so a
			// bar never escapes into the row above when the book has more
			// levels than the height can hold.
			p.Clip(r, 0, func() {
				w := r.W * float32(share)
				if align == ui.Start {
					p.Fill(ui.Rect{X: r.X, Y: r.Y, W: w, H: r.H}, ink.Alpha(0.12), 0)
				} else {
					p.Fill(ui.Rect{X: r.X + r.W - w, Y: r.Y, W: w, H: r.H}, ink.Alpha(0.12), 0)
				}
			})
		}).Children(func() {
		if align == ui.Start {
			mono(c, FormatPrice(l.Price), theme.RowSize, k.Text)
			ui.Text(c, size).TextColor(k.TextMuted).Grow(1).TextAlign(ui.End).
				FontSize(core.FontSize(c, theme.RowSize))
		} else {
			ui.Text(c, size).TextColor(k.TextMuted).Grow(1).
				FontSize(core.FontSize(c, theme.RowSize))
			mono(c, FormatPrice(l.Price), theme.RowSize, k.Text)
		}
	})
}

// levelLabel is a level's whole sentence, for a screen reader. A row of two
// numbers with no words says nothing at all when read out.
func levelLabel(l Level) string {
	side := "bid"
	if l.Side == SideAsk {
		side = "ask"
	}
	return side + " " + FormatPrice(l.Price) + ", " + internal.Commas(int(l.Size+0.5)) + " shares"
}

// takeSide is one side's levels, best first, at most rows of them.
func takeSide(levels []Level, side Side, rows int) []Level {
	var out []Level
	for _, l := range levels {
		if l.Side == side {
			out = append(out, l)
			if len(out) == rows {
				break
			}
		}
	}
	return out
}

// ── the ladder ─────────────────────────────────────────────────────────────

// DepthLadderOptions configure a DepthLadder.
type DepthLadderOptions struct {
	// Bids and Asks are the two sides, the caller's.
	Bids, Asks []Price
	// Levels is how many of each side to show, zero for fifteen.
	Levels int
	// Height is the ladder's height, and is required: a ladder with no height
	// grows to fit every level.
	Height float32
	// Bins is how many steps the price ladder is cut into across its whole
	// range, zero for twenty-four. Coarse bins aggregate: a ladder with one
	// row per price in a hundred-thousand-share book is a list, not a picture
	// of depth.
	Bins int
	// Label names the ladder.
	Label string
}

// DepthLadder is depth aggregated into price bands.
//
// It is a chart and it is not ui/chart's, and the reason is specific rather
// than general: a depth ladder's x axis is a price *band* whose width is the
// width of the band, and its bars are drawn from the middle outwards on two
// different scales that share an axis. That is a bar chart with a bilateral
// axis, which is one shape and one piece of arithmetic — and ui/chart already
// has a bar chart, so the axis and the gutters and the label measuring all
// come from it and only the shape is drawn here.
func DepthLadder(c *ui.Context, opts DepthLadderOptions) *ui.Element {
	if opts.Label == "" {
		panic("finance: DepthLadder needs a Label; a picture of depth with no name is a " +
			"gradient")
	}
	if opts.Height <= 0 {
		panic("finance: DepthLadder needs a Height; a ladder with no height grows to fit every " +
			"band rather than scrolling")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	rows := opts.Levels
	if rows <= 0 {
		rows = 15
	}
	bins := opts.Bins
	if bins <= 0 {
		bins = 24
	}

	levels := Levels(opts.Bids, opts.Asks)
	bands := aggregate(levels, bins)
	widest := 0.0
	for _, b := range bands {
		widest = maxf(widest, maxf(b.BidSize, b.AskSize))
	}

	return ui.Box(c).FillWidth().Height(opts.Height).Label(opts.Label).
		Role(ui.RoleTable).Draw(func(p *ui.Painter, r ui.Rect) {
		if r.W <= 0 || r.H <= 0 || len(bands) == 0 {
			return
		}
		rowH := r.H / float32(len(bands))
		mid := r.X + r.W/2
		for i, b := range bands {
			y := r.Y + rowH*float32(i)
			draw := func(size float64, left bool, col ui.Color) {
				half := r.W/2 - u
				span := float32(0.0)
				if widest > 0 {
					span = half * float32(size/widest)
				}
				if span <= 0 {
					return
				}
				x := mid
				if left {
					x -= span
				}
				p.Fill(ui.Rect{X: x, Y: y + rowH*0.1, W: span, H: rowH * 0.8},
					col, u*0.5)
			}
			draw(b.BidSize, true, k.Success.Alpha(0.75))
			draw(b.AskSize, false, k.Danger.Alpha(0.75))
		}
		// The mid line is the last price: everything to its left is what
		// buyers will pay, everything to its right is what sellers will take,
		// and the two sides of it are the two halves of the market.
		p.Line(mid, r.Y, mid, r.Y+r.H, theme.BorderWidth*2, k.TextMuted)
	}).Children(func() {
		ui.Box(c).FillWidth().Height(u * 3).Shrink(0).Label(opts.Label + " axis").
			Children(func() {
				spreadRow(c, u, func() {
					ui.Text(c, "Bids").TextColor(k.Success).
						FontSize(core.FontSize(c, theme.CaptionSize))
					ui.Box(c).Grow(1)
					ui.Text(c, "Asks").TextColor(k.Danger).
						FontSize(core.FontSize(c, theme.CaptionSize))
				})
			})
	})
}

// w is a side's bar width, as a share of the half it may fill.
func w(size, widest float64) float32 {
	if widest <= 0 {
		return 0
	}
	return float32(size / widest)
}

// Band is one price band of a depth ladder: the range it covers and what
// rests in it on either side.
type Band struct {
	// Low and High are the band's price ends.
	Low, High float64
	// BidSize and AskSize are what rests inside the band on each side.
	BidSize, AskSize float64
}

// Label is the band as a range, "101.20 to 101.25", which is what a reader
// needs: a band's own name is its two ends, not its middle.
func (b Band) Label() string {
	return FormatPrice(b.Low) + " to " + FormatPrice(b.High)
}

// aggregate cuts a whole book into price bands.
//
// It bins across the *whole* book rather than each side separately, because a
// ladder's two halves share one price axis and are read against each other:
// the same vertical strip is "what buyers will pay" on the left and "what
// sellers will take" on the right, and bins that were cut independently would
// put the two halves at different prices in the same row.
func aggregate(levels []Level, bins int) []Band {
	if len(levels) == 0 {
		return nil
	}
	lo, hi := levels[0].Price, levels[0].Price
	for _, l := range levels[1:] {
		lo = minf(lo, l.Price)
		hi = maxf(hi, l.Price)
	}
	if bins <= 0 || hi <= lo {
		bins = 1
	}
	out := make([]Band, bins)
	step := (hi - lo) / float64(bins)
	if step <= 0 {
		step = 1
	}
	for i := range out {
		out[i] = Band{Low: lo + step*float64(i), High: lo + step*float64(i+1)}
	}
	for _, l := range levels {
		i := int((l.Price - lo) / step)
		if i >= bins {
			i = bins - 1
		}
		if i < 0 {
			i = 0
		}
		if l.Side == SideBid {
			out[i].BidSize += l.Size
		} else {
			out[i].AskSize += l.Size
		}
	}
	return out
}

// ── the tape ───────────────────────────────────────────────────────────────

// Trade is one print: what traded, how much, at what price, and which way it
// pushed the price.
type Trade struct {
	// Price and Size are the print.
	Price float64
	Size  int
	// Side is which way it was lifted or hit. A blank trade prints in a
	// neutral tone rather than being assigned one, because a blank really
	// does mean "nobody said".
	Side Side
	// When is when, as the caller's string.
	When string
	// Venue is where, for a consolidated tape.
	Venue string
}

// SideAction is what a trade did to the price: lifted the offer, hit the bid,
// or neither known.
type SideAction int

const (
	// ActionUnknown is a trade with no side. It prints neutral rather than
	// being drawn as one of the other two: a consolidated tape regularly
	// cannot say, and guessing would put a number on screen that the reader
	// would act on.
	ActionUnknown SideAction = iota
	// ActionBuy lifted the offer.
	ActionBuy
	// ActionSell hit the bid.
	ActionSell
)

// Trade is exported with a Side of its own rather than reusing the book's
// [Side], because "bid" and "ask" in a book are where an order sits and "buy"
// and "sell" on a tape are what a trade did. They are not the same fact, and a
// tape that drew a buy as the bid column would be one whole meaning off.

// TimeAndSalesOptions configure a TimeAndSales.
type TimeAndSalesOptions struct {
	// Trades are the caller's, newest first. It is not reversed here: a tape
	// is read downwards and the caller has an order it knows about.
	Trades []Trade
	// Height is the tape's height, and is required: a tape with no height
	// grows to fit every trade, which is every trade drawn.
	Height float32
	// State and Scroll keep the tape's place between frames.
	State  *ui.ListState
	Scroll *ui.ScrollState
	// Highlight marks trades in the caller's own accent — a block, an
	// auction — which is a fact about the tape and not about the price.
	Highlight []int
	// MaxSize is the largest print to draw a size bar against, zero for the
	// largest in the tape.
	MaxSize int
	// Label names the tape.
	Label string
}

// TimeAndSales is the tape: every print in the session, newest at the top.
//
// It is [data.DataTable] rather than a list, because a tape has three things
// to line up and all three of them are numbers — the time, the price and the
// size — and a list of one-line rows would be three lists stitched together
// with their relationship held by being in the same row.
func TimeAndSales(c *ui.Context, opts TimeAndSalesOptions) *ui.Element {
	if opts.Label == "" {
		panic("finance: TimeAndSales needs a Label; a tape is only a tape if somebody says " +
			"which one it is")
	}
	if opts.Height <= 0 {
		panic("finance: TimeAndSales needs a Height; a tape with no height grows to fit every " +
			"trade rather than scrolling")
	}
	k := core.Tokens(c)
	maxSize := opts.MaxSize
	if maxSize <= 0 {
		for _, tr := range opts.Trades {
			if tr.Size > maxSize {
				maxSize = tr.Size
			}
		}
	}

	cols := []data.Column{
		{Title: "Time", ID: "time", Width: 88},
		{Title: "Price", ID: "price", Width: 104, Align: ui.End},
		{Title: "Size", ID: "size", Width: 88, Align: ui.End},
	}

	return tableOf(c, tableOptions{
		columns: cols,
		rows:    len(opts.Trades),
		name:    opts.Label,
		cell: func(row, col int) {
			tr := opts.Trades[row]
			switch cols[col].ID {
			case "time":
				ui.Text(c, tr.When).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.RowSize))
			case "price":
				ink := k.Text
				switch tr.Side {
				case SideBid:
					ink = k.Success
				case SideAsk:
					ink = k.Danger
				}
				mono(c, FormatPrice(tr.Price), theme.RowSize, ink)
			case "size":
				ui.Box(c).FillWidth().Label(
					FormatVolume(tr.Size) + " shares").Children(func() {
					size := 0.0
					if maxSize > 0 {
						size = float64(tr.Size) / float64(maxSize)
					}
					// The bar claims the cell's width and draws its share of
					// it: sized by its own content it is a box that only
					// draws, measures zero wide, and the tape's sizes all
					// read as blank.
					ui.Box(c).FillWidth().Height(theme.RowSize).Radius(theme.RowSize / 2).
						Draw(func(p *ui.Painter, r ui.Rect) {
							p.Fill(ui.Rect{X: r.X, Y: r.Y, W: r.W * float32(size), H: r.H},
								k.Border, r.H/2)
						})
				})
			}
		},
		key:   func(row int) any { return row },
		row:   func(row int) string { return opts.Label },
		state: opts.State, scroll: opts.Scroll, height: opts.Height,
		empty: func() {
			ui.Text(c, "No trades yet").TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.BodySize))
		},
	})
}

// ── trade history ──────────────────────────────────────────────────────────

// Fill is one completed trade.
type Fill struct {
	// ID identifies it, and is required for the same reason every other row
	// here is: the table is sorted and the rows have to follow their records.
	ID string
	// When and Instrument say what and where.
	When       string
	Instrument string
	// Side is which way the order went.
	Side Side
	// Price and Size are the fill.
	Price float64
	Size  float64
	// Fee is what it cost, and what it paid. A negative fee is a rebate.
	Fee float64
	// Venue is where it happened, for a multi-venue account.
	Venue string
}

// TradeHistoryOptions configure a TradeHistory.
type TradeHistoryOptions struct {
	// Fills are the caller's, in the order they should be shown.
	Fills []Fill
	// Height is the table's height, and is required.
	Height float32
	// Sort is the column the rows are ordered by, in the caller's state.
	Sort *data.Sort
	// State and Scroll keep the table's place between frames.
	State  *ui.ListState
	Scroll *ui.ScrollState
	// Label names the table; empty takes the library's "Trade history".
	Label string
	// Empty draws instead of the rows when there are none.
	Empty func()
}

// TradeHistory is every fill on the account, with what each one cost.
//
// The fee is a column rather than a note because it is the number that turns a
// strategy into a result. A history with prices and no fees says a trade made
// money when it made less than its costs, and the only place that is visible
// is a column.
func TradeHistory(c *ui.Context, opts TradeHistoryOptions) *ui.Element {
	if opts.Height <= 0 {
		panic("finance: TradeHistory needs a Height; a table with no height grows to fit " +
			"every row rather than scrolling")
	}
	k := core.Tokens(c)
	name := opts.Label
	if name == "" {
		name = "Trade history"
	}
	if opts.Empty == nil {
		opts.Empty = func() {
			ui.Text(c, "Nothing has traded on this account yet").TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.BodySize))
		}
	}

	cols := []data.Column{
		{Title: "Time", ID: "when", Width: 120, Sortable: true},
		{Title: "Instrument", ID: "instrument", Share: 1},
		{Title: "Side", ID: "side", Width: 72},
		{Title: "Price", ID: "price", Width: 108, Align: ui.End, Sortable: true},
		{Title: "Size", ID: "size", Width: 96, Align: ui.End, Sortable: true},
		{Title: "Fee", ID: "fee", Width: 96, Align: ui.End, Sortable: true},
		{Title: "Venue", ID: "venue", Width: 96},
	}

	return tableOf(c, tableOptions{
		columns: cols,
		rows:    len(opts.Fills),
		name:    name,
		cell: func(row, col int) {
			f := opts.Fills[row]
			switch cols[col].ID {
			case "when":
				ui.Text(c, f.When).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.RowSize))
			case "instrument":
				SymbolBadge(c, SymbolBadgeOptions{
					Symbol: Symbol{Ticker: f.Instrument},
				})
			case "side":
				side := "—"
				ink := k.TextMuted
				switch f.Side {
				case SideBid:
					side, ink = "Buy", k.Success
				case SideAsk:
					side, ink = "Sell", k.Danger
				}
				ui.Text(c, side).TextColor(ink).
					FontSize(core.FontSize(c, theme.RowSize))
			case "price":
				mono(c, FormatPrice(f.Price), theme.RowSize, k.Text)
			case "size":
				mono(c, FormatSize(f.Size), theme.RowSize, k.Text)
			case "fee":
				// A negative fee is a rebate and is drawn in the success tone:
				// it is money coming back, and a reader scanning for costs
				// needs it to stand out rather than hide among the negatives
				// they expect.
				ink := k.Text
				if f.Fee < 0 {
					ink = k.Success
				}
				mono(c, FormatSigned(f.Fee), theme.RowSize, ink)
			case "venue":
				ui.Text(c, f.Venue).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.RowSize))
			}
		},
		key:   func(row int) any { return opts.Fills[row].ID },
		row:   func(row int) string { return name },
		state: opts.State, scroll: opts.Scroll, height: opts.Height, empty: opts.Empty,
	})
}

// OrderHistory is every order the account has had today, newest first.
//
// It is [OrderTable] with a time column and without the working-order
// bookkeeping, because the two are not the same screen: an order history is
// looked *up* — "what happened to the order I sent at ten?" — and it is
// scanned down a column of times, while a working orders table is watched and
// is a thing to act on. Sharing the columns and not the purpose is why they
// are two components with one column set between them.
func OrderHistory(c *ui.Context, opts OrderHistoryOptions) *ui.Element {
	if opts.Height <= 0 {
		panic("finance: OrderHistory needs a Height; a table with no height grows to fit every " +
			"order rather than scrolling")
	}
	k := core.Tokens(c)
	empty := opts.Empty
	if empty == nil {
		empty = func() {
			ui.Text(c, "Nothing has traded today").TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.BodySize))
		}
	}

	cols := []tableColumn{
		{Title: "Time", ID: "when", Width: 120},
		{Title: "Instrument", ID: "instrument", Width: 128},
		{Title: "Side", ID: "side", Width: 72},
		{Title: "Type", ID: "kind", Width: 84},
		{Title: "Price", ID: "price", Width: 108, Align: ui.End},
		{Title: "Size", ID: "size", Width: 96, Align: ui.End},
		{Title: "State", ID: "state", Share: 1},
	}

	return tableOf(c, tableOptions{
		columns: cols,
		rows:    len(opts.Orders),
		name:    "Order history",
		cell: func(row, col int) {
			o := opts.Orders[row]
			switch cols[col].ID {
			case "when":
				ui.Text(c, o.When).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.RowSize))
			case "instrument":
				SymbolBadge(c, SymbolBadgeOptions{Symbol: Symbol{Ticker: o.Instrument}})
			case "side":
				sideWord, ink := "Buy", k.Success
				if o.Side == SideAsk {
					sideWord, ink = "Sell", k.Danger
				}
				ui.Text(c, sideWord).TextColor(ink).
					FontSize(core.FontSize(c, theme.RowSize))
			case "kind":
				ui.Text(c, o.Kind.String()).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.RowSize))
			case "price":
				if o.Kind == OrderMarket {
					ui.Text(c, "Market").TextColor(k.TextFaint).
						FontSize(core.FontSize(c, theme.RowSize))
					return
				}
				mono(c, FormatPrice(o.Price), theme.RowSize, k.Text)
			case "size":
				mono(c, FormatSize(o.Size), theme.RowSize, k.Text)
			case "state":
				ink := k.Text
				if o.Tone != core.Neutral {
					_, ink = o.Tone.Pair(k)
				}
				ui.Text(c, o.State).TextColor(ink).MaxLines(1).
					FontSize(core.FontSize(c, theme.RowSize))
			}
		},
		key:   func(row int) any { return opts.Orders[row].ID },
		row:   func(row int) string { return opts.Orders[row].Instrument },
		state: opts.State, scroll: opts.Scroll, height: opts.Height, empty: empty,
	})
}

// OrderHistoryOptions configure an OrderHistory.
type OrderHistoryOptions struct {
	// Orders are the caller's, in the order they should be shown — which is
	// not sorted here: a history is a log, and the order of a log is the
	// thing it is for.
	Orders []Order
	// Height is the table's height, and is required.
	Height float32
	// State and Scroll keep the table's place between frames.
	State  *ui.ListState
	Scroll *ui.ScrollState
	// Empty draws instead of the rows when there are none.
	Empty func()
}
