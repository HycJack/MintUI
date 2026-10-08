package finance

import (
	"math"
	"strconv"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/data"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// Position is one holding: what it is, how much, and what it is worth now.
type Position struct {
	// ID identifies the holding, and is required: a portfolio is sorted by
	// every column a reader might click, and a row that follows its own
	// number instead of its instrument scrolls to the wrong place when the
	// order changes.
	ID string
	// Instrument is the ticker.
	Instrument string
	// Shares is how much is held. Negative is short, and every figure derived
	// from it keeps the sign rather than taking an absolute value: a short
	// position that shows its loss as a gain is the worst bug this package
	// could have.
	Shares float64
	// Average is what each share cost.
	Average float64
	// Last is what each share is worth now.
	Last float64
	// DayHigh and DayLow are the session's extremes, for the day's move.
	DayHigh, DayLow float64
	// Unrealised is the profit or loss on the position, in money. It is the
	// caller's because whether a position is marked to market, to last trade
	// or to a mid is the caller's exchange's business.
	Unrealised float64
	// Realised is what has been banked on it so far.
	Realised float64
	// Weight is the position's share of the portfolio, 0 to 1, for the
	// allocation chart.
	Weight float64
	// Sector is what it is filed under, for a grouped view.
	Sector string
}

// MarketValue is what the position is worth now, and is computed rather than
// taken from a field: shares times last price is the definition, and a
// component that trusted a stored total would show a portfolio that disagrees
// with its own rows.
func (p Position) MarketValue() float64 { return p.Shares * p.Last }

// CostBasis is what was paid for it.
func (p Position) CostBasis() float64 { return p.Shares * p.Average }

// DayChange is how far the last price has moved today, and is computed from
// the day's low and high rather than from a stored previous close: a position
// table that showed a day's move needs a previous close, and a caller who has
// one has [PriceChangeBadge] for it.
func (p Position) DayChange() float64 { return p.Last - p.DayLow }

// Return is the position's percentage return on cost, and 0 for one with no
// cost basis rather than a division by zero: a position granted rather than
// bought has no return, and Infinity is not a percentage.
func (p Position) Return() float64 {
	cost := p.CostBasis()
	if cost == 0 {
		return 0
	}
	return p.Unrealised / cost * 100
}

// ── the position table ─────────────────────────────────────────────────────

// positionColumns are the position table's columns. Every width is fixed
// because every column of this table holds a number, and a number that wraps
// is a number nobody can compare.
var positionColumns = []data.Column{
	{Title: "Instrument", ID: "instrument", Share: 1},
	{Title: "Shares", ID: "shares", Width: 104, Align: ui.End, Sortable: true},
	{Title: "Avg", ID: "average", Width: 96, Align: ui.End, Sortable: true},
	{Title: "Last", ID: "last", Width: 96, Align: ui.End, Sortable: true},
	{Title: "Value", ID: "value", Width: 116, Align: ui.End, Sortable: true},
	{Title: "P/L", ID: "pnl", Width: 116, Align: ui.End, Sortable: true},
	{Title: "Return", ID: "return", Width: 88, Align: ui.End, Sortable: true},
}

// PositionTableOptions configure a PositionTable.
type PositionTableOptions struct {
	// Positions are the caller's, in the order they should be shown.
	Positions []Position
	// Selected is the row the keys move from, as an index; -1 for none.
	Selected *int
	// Sort is the column the rows are ordered by, in the caller's state.
	Sort *data.Sort
	// Height is the table's height, and is required: a position table with no
	// height grows to fit every holding, which is every holding drawn.
	Height float32
	// State and Scroll keep the table's place between frames.
	State  *ui.ListState
	Scroll *ui.ScrollState
	// ShowReturn leaves the percentage column out, for a narrow panel where
	// the money column is the one that has to stay.
	ShowReturn bool
	// Empty draws instead of the rows when there are none.
	Empty func()
}

// PositionTable is the portfolio's holdings and what each is worth.
//
// It is [data.DataTable] for the reason every table here is: seven things to
// line up, all of them numbers, and the columns have to keep their widths so
// a price does not wrap at a window size nobody was designing for.
func PositionTable(c *ui.Context, opts PositionTableOptions) *ui.Element {
	if opts.Height <= 0 {
		panic("finance: PositionTable needs a Height; a table with no height grows to fit " +
			"every holding rather than scrolling")
	}
	k := core.Tokens(c)
	empty := opts.Empty
	if empty == nil {
		empty = func() {
			ui.Text(c, "No positions").TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.BodySize))
		}
	}

	cols := make([]data.Column, 0, len(positionColumns))
	for _, col := range positionColumns {
		if col.ID == "return" && !opts.ShowReturn {
			continue
		}
		cols = append(cols, col)
	}

	return data.DataTable(c, data.DataTableOptions{
		Columns: cols,
		Rows:    len(opts.Positions),
		Sort:    opts.Sort,
		Cell: func(row, col int) {
			p := opts.Positions[row]
			switch cols[col].ID {
			case "instrument":
				SymbolBadge(c, SymbolBadgeOptions{Symbol: Symbol{Ticker: p.Instrument}})
			case "shares":
				// The sign is kept: a short position shown as a positive size
				// is a position the reader cannot tell the direction of.
				mono(c, FormatSize(p.Shares), theme.RowSize, k.Text)
			case "average":
				mono(c, FormatPrice(p.Average), theme.RowSize, k.Text)
			case "last":
				mono(c, FormatPrice(p.Last), theme.RowSize, k.Text)
			case "value":
				mono(c, FormatPrice(p.MarketValue()), theme.RowSize, k.Text)
			case "pnl":
				ink := k.Text
				if p.Unrealised < 0 {
					ink = k.Danger
				} else if p.Unrealised > 0 {
					ink = k.Success
				}
				mono(c, FormatSigned(p.Unrealised), theme.RowSize, ink)
			case "return":
				r := p.Return()
				ink := k.TextMuted
				if r > 0 {
					ink = k.Success
				} else if r < 0 {
					ink = k.Danger
				}
				mono(c, formatPercent(r), theme.RowSize, ink)
			}
		},
		CellLabel: func(row, col int) string { return opts.Positions[row].Instrument },
		Key:       func(row int) any { return opts.Positions[row].ID },
		Label:     func(row int) string { return opts.Positions[row].Instrument },
		Selected:  opts.Selected,
		State:     opts.State,
		Scroll:    opts.Scroll,
		Height:    opts.Height,
		Empty:     empty,
	}).Element.FillWidth()
}

// formatShare writes a share of something — a margin in use, a position's
// weight, a maintenance level — with two places and *no* sign.
//
// It is a separate function from formatPercent because the sign is the whole
// difference and it is the thing that makes one right and the other wrong.
// "+30.00% used" reads as a rise of thirty percent in a usage figure, and a
// usage figure does not rise.
func formatShare(v float64) string {
	return strconv.FormatFloat(v, 'f', 2, 64) + "%"
}

// formatPercent writes a percentage that is already a percentage, with two
// places and a sign.
//
// It cannot go through [FormatChange], which measures a change against a
// previous value: a share of 30 is "30.00%" here and would be "0.00%" there,
// because thirty percent of nothing is nothing. The one rule it does share is
// the important one — a flat figure carries no sign — and that is written out
// rather than borrowed.
func formatPercent(v float64) string {
	if v == 0 {
		return "0.00%"
	}
	s := strconv.FormatFloat(math.Abs(v), 'f', 2, 64)
	if v > 0 {
		return "+" + s + "%"
	}
	return "-" + s + "%"
}

// ── profit and loss ────────────────────────────────────────────────────────

// PnLOptions configure a PnLDisplay.
type PnLOptions struct {
	// Unrealised is what the open positions are worth more or less than they
	// cost, and Realised what has been banked. Both are the caller's: which
	// of them a desk calls "P&L" is a house convention, not a library's.
	Unrealised float64
	Realised   float64
	// Currency is the unit, written on the total only. The two parts are
	// labelled and the total carries the unit, because the unit on both is
	// the same number three times.
	Currency string
	// Label names the display for assistive technology; empty builds one.
	Label string
	// Size is the total's type size; zero takes the card's figure size.
	Size float32
}

// PnLDisplay is the money made and the money not yet made.
//
// The parts are in a row above the total rather than in a column with it. A
// column would make the total look like a third line of the same kind of
// thing; the total is a different kind of thing — it is the answer — and it is
// bigger and on its own line so that it is found first.
func PnLDisplay(c *ui.Context, opts PnLOptions) *ui.Element {
	u := core.Density(c).Unit()
	label := opts.Label
	if label == "" {
		label = "Profit and loss"
	}
	return ui.Column(c).FillWidth().Gap(u).Label(label).Children(func() {
		spreadRow(c, u, func() {
			figure(c, stat{
				Value: FormatSigned(opts.Realised), Label: "Realised", Mono: true,
				Tone: pnlTone(opts.Realised),
			})
			figure(c, stat{
				Value: FormatSigned(opts.Unrealised), Label: "Unrealised", Mono: true,
				Tone: pnlTone(opts.Unrealised),
			})
		})
		// The total is always drawn. A P&L with two figures and no total is
		// two questions where a reader wanted one, and a boolean to hide it
		// would have a false zero: a caller who did not know about the field
		// would get the parts and wonder where the answer was.
		{
			total := opts.Realised + opts.Unrealised
			spreadRow(c, u, func() {
				_, tone := FormatChange(total, 0)
				quoteBand(c, func() {
					spreadRow(c, u, func() {
						figure(c, stat{
							Value: FormatSigned(total), Unit: opts.Currency,
							Label: "Total profit and loss", Mono: true, Tone: tone,
						})
					})
				})
			})
		}
	})
}

// pnlTone is what tone a profit or loss figure is drawn in. Zero is neutral
// rather than either extreme: a figure that has neither made nor lost money
// has no story in it, and painting it green says it did.
func pnlTone(v float64) core.Severity {
	switch {
	case v > 0:
		return core.Success
	case v < 0:
		return core.Danger
	}
	return core.Neutral
}

// ── the portfolio ──────────────────────────────────────────────────────────

// PortfolioSummaryOptions configure a PortfolioSummary.
type PortfolioSummaryOptions struct {
	// Positions are the caller's holdings; the figures below are computed from
	// them rather than taken from the caller, so the summary can never
	// disagree with the table underneath it.
	Positions []Position
	// Cash is what is not invested, and is part of the total: a portfolio's
	// value is its holdings *and* its cash, and a summary that left the cash
	// out would show a smaller number than the account holds.
	Cash float64
	// DayChange is what the whole account is worth today, and is the caller's
	// because only it knows which of its positions moved.
	DayChange float64
	// Currency is the unit for every figure.
	Currency string
	// Inception is when the account started, for the return since then.
	Inception string
	// Title heads the summary; empty takes the library's "Portfolio".
	Title string
	// ShowAllocation draws the ring on the right, which is what a wide
	// summary has room for.
	ShowAllocation bool
	// Width and Height are the ring's box; zero lets it take its share.
	Width, Height float32
}

// PortfolioSummary is the whole account in a card: what it is worth, what it
// did today, and what it is made of.
//
// The total is the largest figure on it. That is a decision about what a
// reader opens this card for: they open it to find out what they have, and
// everything else on the card is the detail they look at afterwards.
func PortfolioSummary(c *ui.Context, opts PortfolioSummaryOptions) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	title := opts.Title
	if title == "" {
		title = "Portfolio"
	}

	total := opts.Cash
	for _, p := range opts.Positions {
		total += p.MarketValue()
	}
	prev := total - opts.DayChange
	_, sev := FormatChange(total, prev)

	return layout.Container(c, layout.ContainerOptions{
		Surface: true, Border: true, Radius: theme.CardRadius,
		Pad: u * 2.5, Gap: u * 2,
	}, func() {
		ui.Text(c, title).TextColor(k.TextMuted).
			FontSize(core.FontSize(c, theme.CaptionSize))
		spreadRow(c, u, func() {
			figure(c, stat{
				Value: FormatPrice(total), Unit: opts.Currency,
				Label: "Total value", Mono: true, Tone: sev,
			})
			if opts.Inception != "" {
				ui.Text(c, "since "+opts.Inception).TextColor(k.TextFaint).
					FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
			}
		})
		spreadRow(c, u, func() {
			PriceChangeBadge(c, PriceChangeBadgeOptions{
				Now: total, Previous: prev, Absolute: opts.DayChange,
				ShowAbsolute: true,
			})
			if n := len(opts.Positions); n > 0 {
				ui.Text(c, pluralPositions(n)).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
			}
		})
		if opts.ShowAllocation {
			AssetAllocationChart(c, AssetAllocationOptions{
				Positions: opts.Positions, Cash: opts.Cash,
				Width: opts.Width, Height: opts.Height,
				Label: title + " allocation",
			})
		}
	})
}

func pluralPositions(n int) string {
	if n == 1 {
		return "1 position"
	}
	return itoa(n) + " positions"
}
