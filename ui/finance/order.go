package finance

import (
	"strconv"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/data"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/overlay"
	"github.com/HycJack/MintUI/ui/theme"
)

// ── entering an order ──────────────────────────────────────────────────────

// OrderKind is what an order does when it fills.
type OrderKind int

const (
	// OrderMarket fills at whatever the market is offering now.
	OrderMarket OrderKind = iota
	// OrderLimit fills at the caller's price or better.
	OrderLimit
	// OrderStop becomes a market order when the price reaches the trigger.
	OrderStop
)

func (k OrderKind) String() string {
	switch k {
	case OrderLimit:
		return "Limit"
	case OrderStop:
		return "Stop"
	}
	return "Market"
}

// OrderEntryOptions configure an OrderEntry.
type OrderEntryOptions struct {
	// Instrument is what is being traded; its ticker is required.
	Instrument Symbol
	// Side is which way. It is the caller's and not this component's, because
	// a screen that defaults to "buy" and forgets to show which way it
	// defaulted to is how an order for the wrong side gets sent.
	Side Side
	// Kind is the order type, and Price the limit or trigger price. A market
	// order leaves Price alone.
	Kind  OrderKind
	Price *float64
	// Size is how much, in shares. It is a pointer because the size field and
	// the notional figure under it are two views of one number and a caller
	// that typed into either must see the other change.
	Size *float64
	// Last is the last traded price, for the estimate under the ticket and for
	// what a market order is worth.
	Last float64
	// Maximum is what the account may spend, and the size field will not go
	// past it. Zero is no limit, which is most retail tickets and none of a
	// professional one.
	Maximum float64
	// Placing is the caller's flag, and the ticket's submit button reports and
	// the caller decides. Nothing here sends an order: a component that could
	// send one could not be shown a confirmation first, and could not be
	// drawn in a test at all.
	Placing bool
	// Submitting reports the submit button being pressed this frame.
	Submitting bool
	// Label names the ticket; empty builds one from the instrument.
	Label string
	// Currency is the unit on the estimates.
	Currency string
}

// OrderEntryResult carries an OrderEntry.
type OrderEntryResult struct {
	// Element is the ticket.
	Element *ui.Element
	// submitting reports the submit button being pressed this frame.
	submitting bool
}

// Submitting reports the submit button being pressed, and false on every other
// frame. It is separate from the Placing flag — which this component reads and
// never writes — so that "the reader asked" and "the order is going" are two
// different answers a caller can act on.
func (r OrderEntryResult) Submitting() bool { return r.submitting }

// OrderEntry is the ticket: what, which way, how much and at what price.
//
// The notional figure — what the order will cost — is written *under* the
// size, and it is the number that decides whether somebody sends the order.
// A ticket that only takes a share count leaves the reader to multiply it in
// their head, and the multiplication is the step where a zero decimal place
// becomes a factor of ten.
func OrderEntry(c *ui.Context, opts OrderEntryOptions) OrderEntryResult {
	if opts.Instrument.Ticker == "" {
		panic("finance: OrderEntry needs an Instrument; a ticket with no ticker on it cannot " +
			"be sent")
	}
	if opts.Price == nil {
		panic("finance: OrderEntry needs a Price to point at; it keeps no price of its own")
	}
	if opts.Size == nil {
		panic("finance: OrderEntry needs a Size to point at; it keeps no size of its own")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	label := opts.Label
	if label == "" {
		label = "Order " + opts.Instrument.Label()
	}

	// The size is capped against the account's limit before anything is drawn,
	// because a size field is the only thing here that can go wrong on its own:
	// a stale value from a previous ticket would otherwise be submitted as
	// one.
	if opts.Maximum > 0 && opts.Last > 0 {
		cap := opts.Maximum / opts.Last
		if *opts.Size > cap {
			*opts.Size = cap
		}
	}
	notional := *opts.Size * *opts.Price

	var res OrderEntryResult
	res.Element = layout.Container(c, layout.ContainerOptions{
		Surface: true, Border: true, Radius: theme.CardRadius,
		Pad: u * 2, Gap: u * 1.5,
	}, func() {
		spreadRow(c, u, func() {
			SymbolBadge(c, SymbolBadgeOptions{Symbol: opts.Instrument, ShowExchange: true})
			ui.Box(c).Grow(1)
			sideWord := "Buy"
			sideTone := k.Success
			if opts.Side == SideAsk {
				sideWord, sideTone = "Sell", k.Danger
			}
			ui.Text(c, sideWord).TextColor(sideTone).
				FontSize(core.FontSize(c, theme.BodySize)).Bold()
		})

		input.Segmented(c, kindIndex(&opts.Kind), "Market", "Limit", "Stop")

		if opts.Kind != OrderMarket {
			input.NumberInput(c, opts.Price, input.NumberInputOptions{
				Label: "Price", Step: 0.01, Suffix: opts.Currency,
				Max: &opts.Last, Width: 0,
			})
		} else {
			// A market order has no price, and showing the last price in the
			// field greyed out would be a price it does not have. It is
			// written as a figure instead, labelled as an estimate.
			ui.Text(c, "Market — last "+FormatPrice(opts.Last)).
				TextColor(k.TextMuted).FontSize(core.FontSize(c, theme.CaptionSize))
		}

		input.NumberInput(c, opts.Size, input.NumberInputOptions{
			Label: "Size", Step: 1, Suffix: "shares",
			Min: zeroRef(), Max: maxRef(opts.Maximum, opts.Last), Width: 0,
		})

		quoteBand(c, func() {
			spreadRow(c, u, func() {
				figure(c, stat{
					Value: FormatPrice(notional), Unit: opts.Currency,
					Label: "Order value", Mono: true,
				})
				if opts.Maximum > 0 {
					figure(c, stat{
						Value: FormatPrice(opts.Maximum - notional),
						Unit:  opts.Currency, Label: "Available after", Mono: true,
						Tone: core.Neutral,
					})
				}
			})
		})

		if input.Button(c, submitLabel(opts), input.ButtonOptions{
			Primary:  true,
			Disabled: opts.Placing || *opts.Size <= 0,
		}).Clicked() {
			res.submitting = true
			opts.Submitting = true
		}
	})
	return res
}

// submitLabel says what the button will do, and says it in the order of the
// trade: "Buy 120 RIVR" rather than "Submit". A reader checking a ticket
// before sending is checking the instrument and the size, and a button
// labelled "Submit" makes them check both again on the confirmation.
func submitLabel(opts OrderEntryOptions) string {
	verb := "Buy"
	if opts.Side == SideAsk {
		verb = "Sell"
	}
	if opts.Placing {
		return "Sending…"
	}
	return verb + " " + FormatSize(*opts.Size) + " " + opts.Instrument.Ticker
}

// kindIndex is an order kind as the index a Segmented matches on, written out
// once so the two can never drift apart — a segmented that highlighted "Limit"
// while the ticket took a stop order's price would be a ticket that sends the
// wrong order without saying so.
func kindIndex(kind *OrderKind) *int {
	i := int(*kind)
	return &i
}

func zeroRef() *float64 { z := 0.0; return &z }

func maxRef(maximum, last float64) *float64 {
	if maximum <= 0 || last <= 0 {
		return nil
	}
	v := maximum / last
	return &v
}

// ── confirming ─────────────────────────────────────────────────────────────

// OrderConfirmOptions configure an OrderConfirm.
type OrderConfirmOptions struct {
	// Open is the *bool the confirmation opens and closes with, and is
	// required: a dialog that closed itself would have to remember whether it
	// was open.
	Open *bool
	// Instrument, Side, Kind, Price and Size describe the order exactly as
	// the ticket had it. They are passed again rather than read back from
	// anywhere, because the thing being confirmed must be a *snapshot*: a
	// confirmation that read live values would change under the reader while
	// they were reading it.
	Instrument Symbol
	Side       Side
	Kind       OrderKind
	Price      float64
	Size       float64
	// Notional is what the order will cost, and Fee what it will charge.
	Notional, Fee float64
	// Last is the price now, so the confirmation can say how far the limit is
	// from it.
	Last float64
	// Currency is the unit on every figure.
	Currency string
	// Placing is the caller's flag; the button is disabled while it is on, so
	// that a slow network does not produce two orders from two presses.
	Placing bool
	// Confirmed reports the confirm button being pressed this frame.
	Confirmed bool
	// Label heads the dialog; empty builds one from the instrument.
	Title string
}

// OrderConfirmResult carries an OrderConfirm.
type OrderConfirmResult struct {
	// Element is the panel, nil while it is closed.
	Element *ui.Element
	// confirmed reports the confirm button being pressed this frame.
	confirmed bool
}

// Confirmed reports the confirm button being pressed. It is reported rather
// than acted on, and the dialog is left open for the caller to close: a
// confirmation that dismissed itself the instant it was accepted would be a
// confirmation that could not show what happened next.
func (r OrderConfirmResult) Confirmed() bool { return r.confirmed }

// OrderConfirm is the last screen before an order goes.
//
// It is [overlay.AlertDialog]'s shape — a panel over the window with two
// buttons — because that is exactly the relationship it has with the window,
// and a second implementation of it would be a second set of rules about what
// Escape does.
func OrderConfirm(c *ui.Context, opts OrderConfirmOptions) OrderConfirmResult {
	if opts.Open == nil {
		panic("finance: OrderConfirm needs the *bool it opens and closes with")
	}
	if opts.Instrument.Ticker == "" {
		panic("finance: OrderConfirm needs an Instrument; a confirmation about nothing in " +
			"particular confirms nothing")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	title := opts.Title
	if title == "" {
		title = "Confirm " + opts.Kind.String() + " order"
	}

	var res OrderConfirmResult
	res.Element = overlay.Dialog(c, opts.Open, overlay.DialogOptions{
		Title: title,
		Rule:  true,
		// Wide enough for its two actions on one row: at the old width the
		// Send order half painted past the sheet's own right edge.
		Width: u * 72,
		Body: func() {
			spreadRow(c, u, func() {
				SymbolBadge(c, SymbolBadgeOptions{Symbol: opts.Instrument})
				sideWord := "Buy"
				if opts.Side == SideAsk {
					sideWord = "Sell"
				}
				ui.Text(c, sideWord+" "+FormatSize(opts.Size)).
					TextColor(k.Text).FontSize(core.FontSize(c, theme.BodySize)).Bold()
			})
			ui.Column(c).FillWidth().Gap(u * 0.75).Children(func() {
				confirmLine(c, "Type", opts.Kind.String())
				if opts.Kind != OrderMarket {
					confirmLine(c, "Price", FormatPrice(opts.Price))
					if opts.Last > 0 {
						confirmLine(c, "Last", FormatPrice(opts.Last))
					}
				}
				confirmLine(c, "Order value", FormatPrice(opts.Notional))
				if opts.Fee != 0 {
					confirmLine(c, "Fee", FormatSigned(opts.Fee))
				}
				if opts.Kind != OrderMarket && opts.Last > 0 {
					confirmLine(c, "Away from market", FormatPrice(opts.Price-opts.Last))
				}
			})
		},
		Actions: func() {
			if input.Button(c, "Cancel", input.ButtonOptions{}).Clicked() {
				*opts.Open = false
			}
			if input.Button(c, "Send order", input.ButtonOptions{
				Primary:  true,
				Disabled: opts.Placing,
			}).Clicked() {
				res.confirmed = true
				opts.Confirmed = true
			}
		},
	})
	return res
}

// confirmLine is one row of the confirmation: a name on the left, a figure on
// the right. Both in the monospaced face for the figure, because a
// confirmation is read one line at a time and the figures have to be findable.
func confirmLine(c *ui.Context, name, value string) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	return ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 2).
		Label(name + " " + value).Children(func() {
		ui.Text(c, name).TextColor(k.TextMuted).
			FontSize(core.FontSize(c, theme.RowSize)).Grow(1)
		mono(c, value, theme.RowSize, k.Text)
	})
}

// ── the orders table ───────────────────────────────────────────────────────

// orderColumns are the orders table's columns. All fixed widths, for the same
// reason as everywhere else here.
var orderColumns = []tableColumn{
	{Title: "Instrument", ID: "instrument", Width: 128},
	{Title: "Type", ID: "kind", Width: 84},
	{Title: "Side", ID: "side", Width: 72},
	{Title: "Price", ID: "price", Width: 108, Align: ui.End},
	{Title: "Size", ID: "size", Width: 96, Align: ui.End},
	{Title: "Filled", ID: "filled", Width: 104, Align: ui.End},
	{Title: "State", ID: "state", Width: 116},
}

// Order is one working order.
type Order struct {
	// ID identifies it, and is required for the same reason every other row
	// here is: the table is sorted by a reader's clicking and the rows have to
	// follow their records.
	ID string
	// Instrument is the ticker.
	Instrument string
	// Kind and Side are what it is and which way.
	Kind OrderKind
	Side Side
	// Price and Size are what was asked for, and Filled how much of it has
	// gone. A partially filled order is a different thing from an unfilled
	// one and the column is what says so.
	Price  float64
	Size   float64
	Filled float64
	// State is the venue's word for it — "Working", "Partially filled",
	// "Cancelled" — and is drawn in the caller's tone because only the caller
	// knows the exchange's vocabulary.
	State string
	// Tone is the state row's severity.
	Tone core.Severity
	// When is when it was placed, as the caller's string.
	When string
}

// OrderTableOptions configure an OrderTable.
type OrderTableOptions struct {
	// Orders are the caller's, in the order they should be shown.
	Orders []Order
	// Height is the table's height, and is required.
	Height float32
	// Sort is the column the rows are ordered by, in the caller's state.
	Sort *data.Sort
	// Selected is the row the keys move from; -1 for none.
	Selected *int
	// State and Scroll keep the table's place between frames.
	State  *ui.ListState
	Scroll *ui.ScrollState
	// Empty draws instead of the rows when there are none.
	Empty func()
}

// OrderTable is the working orders and what has happened to each.
//
// The filled column is there because "working" and "partly working" are
// different things and an orders table with only a state column makes a
// reader click through to find out how much of an order is left.
func OrderTable(c *ui.Context, opts OrderTableOptions) *ui.Element {
	if opts.Height <= 0 {
		panic("finance: OrderTable needs a Height; a table with no height grows to fit every " +
			"row rather than scrolling")
	}
	k := core.Tokens(c)
	empty := opts.Empty
	if empty == nil {
		empty = func() {
			ui.Text(c, "No working orders").TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.BodySize))
		}
	}

	return tableOf(c, tableOptions{
		columns: orderColumns,
		rows:    len(opts.Orders),
		name:    "Orders",
		cell: func(row, col int) {
			o := opts.Orders[row]
			switch orderColumns[col].ID {
			case "instrument":
				SymbolBadge(c, SymbolBadgeOptions{Symbol: Symbol{Ticker: o.Instrument}})
			case "kind":
				ui.Text(c, o.Kind.String()).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.RowSize))
			case "side":
				sideWord, ink := "Buy", k.Success
				if o.Side == SideAsk {
					sideWord, ink = "Sell", k.Danger
				}
				ui.Text(c, sideWord).TextColor(ink).
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
			case "filled":
				if o.Filled <= 0 {
					ui.Text(c, "—").TextColor(k.TextFaint).
						FontSize(core.FontSize(c, theme.RowSize))
					return
				}
				mono(c, FormatSize(o.Filled)+" of "+FormatSize(o.Size), theme.RowSize, k.Text)
			case "state":
				ink := k.Text
				if o.Tone != core.Neutral {
					_, ink = o.Tone.Pair(k)
				}
				ui.Text(c, o.State).TextColor(ink).
					FontSize(core.FontSize(c, theme.RowSize)).MaxLines(1)
			}
		},
		key:   func(row int) any { return opts.Orders[row].ID },
		row:   func(row int) string { return opts.Orders[row].Instrument },
		state: opts.State, scroll: opts.Scroll, height: opts.Height, empty: empty,
	})
}

// ── one-tap trading ────────────────────────────────────────────────────────

// QuickTradeButtonsOptions configure a QuickTradeButtons.
type QuickTradeButtonsOptions struct {
	// Size is the order size each button places, and is the caller's: a quick
	// trade button whose size the component decided would be a button placing
	// an order nobody chose.
	Size float64
	// Instrument is what is being traded; its ticker is required.
	Instrument Symbol
	// Bid and Ask are the prices the two buttons use, so that the buttons
	// cannot disagree with the book beside them.
	Bid, Ask float64
	// Side is which way the row trades in; a row with two buttons is a bid
	// button and an ask button, and this is which of them is the default.
	Side Side
	// Placing is the caller's flag, and Disabling the two buttons while it is
	// on: two quick-trade presses in a row are two orders, which is a very
	// expensive way to learn what a double-click is.
	Placing bool
	// Ordered is the instrument ordered this frame.
	Ordered bool
	// Label names the row for assistive technology; empty builds one.
	Label string
}

// QuickTradeButtonsResult carries a QuickTradeButtons.
type QuickTradeButtonsResult struct {
	// Element is the row.
	Element *ui.Element
	// ordered reports a press this frame, and which side.
	side Side
	// pressed reports that a button was pressed at all.
	pressed bool
}

// Ordered reports a press this frame and which side it was. Both halves
// because a caller that only learns "something was pressed" has to guess which
// button, and the two buttons are opposite trades.
func (r QuickTradeButtonsResult) Ordered() (Side, bool) { return r.side, r.pressed }

// QuickTradeButtons is the pair of buttons that put on a size at the bid or
// the ask, in one press.
//
// The prices are *on* the buttons. A quick-trade row whose prices are somewhere
// else on the screen is a row where the reader has to look away from the thing
// they are pressing to check what they are pressing, and a misread in that
// direction is an order in the wrong direction at the wrong price.
func QuickTradeButtons(c *ui.Context, opts QuickTradeButtonsOptions) QuickTradeButtonsResult {
	if opts.Instrument.Ticker == "" {
		panic("finance: QuickTradeButtons needs an Instrument; a button with no ticker on " +
			"it places an order nobody can check")
	}
	if opts.Size <= 0 {
		panic("finance: QuickTradeButtons needs a positive Size; a quick trade of nothing is " +
			"not a quick trade")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	label := opts.Label
	if label == "" {
		label = "Quick trade " + opts.Instrument.Label()
	}

	var res QuickTradeButtonsResult
	res.Element = ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.5).
		Label(label).Role(ui.RoleGroup).Children(func() {
		for _, side := range []Side{SideBid, SideAsk} {
			price := opts.Bid
			verb := "Buy"
			if side == SideAsk {
				price, verb = opts.Ask, "Sell"
			}
			buttonLabel := verb + " " + FormatSize(opts.Size) + " at " + FormatPrice(price)
			btn := input.Button(c, buttonLabel, input.ButtonOptions{
				Primary:  side == opts.Side,
				Disabled: opts.Placing,
				Label:    buttonLabel,
			})
			if btn.Clicked() {
				res.side, res.pressed = side, true
				opts.Ordered = true
			}
		}
	})
	_ = k
	return res
}

// ── options ────────────────────────────────────────────────────────────────

// Contract is one option: its strike, its expiry, its price and its Greeks.
type Contract struct {
	// Symbol is the option's own contract identifier, and is required: an
	// options chain's rows are identified by their strike *and* their expiry,
	// and one without the other is half an instrument.
	Symbol string
	// Underlying is the instrument it is on.
	Underlying string
	// Expiry is the date, as the caller's string.
	Expiry string
	// Strike is the price the option is on.
	Strike float64
	// Kind is Call or Put.
	Kind OptionKind
	// Bid, Ask and Last are the prices.
	Bid, Ask, Last float64
	// Volume is the day's.
	Volume int
	// Greeks are the sensitivities, and every one of them is the caller's:
	// there are four conventions for the sign of a position's delta and this
	// library has no business picking one.
	Greeks Greeks
}

// OptionKind is which kind of option a contract is.
type OptionKind string

const (
	// OptionCall is the right to buy the underlying at the strike.
	OptionCall OptionKind = "Call"
	// OptionPut is the right to sell it at the strike.
	OptionPut OptionKind = "Put"
)

// Greeks are an option's five sensitivities. They are values rather than
// functions because every one of them is a number the caller's pricing model
// produced, and a component that recomputed a delta would be a component
// quietly disagreeing with the model the reader priced the trade with.
type Greeks struct {
	Delta   float64
	Gamma   float64
	Theta   float64
	Vega    float64
	Rho     float64
	Implied float64
}

// greekColumns are the Greeks table's columns. Every one is fixed width and
// every one is numeric, which is the whole case for a table here rather than
// for five more columns in the chain.
var greekColumns = []tableColumn{
	{Title: "Greeks", ID: "name", Width: 116},
	{Title: "Value", ID: "value", Width: 116, Align: ui.End},
	{Title: "For", ID: "what", Share: 1},
}

// greekRow is one Greek and what it is sensitive to.
type greekRow struct {
	name  string
	value float64
	what  string
	// decimals is how many places the figure is written to. Delta and gamma
	// are fractions and theta is a fraction of the premium per day, and
	// writing them all to two places would write a delta of 0.005 as 0.01.
	decimals int
}

// greekRows is what a Greeks table shows, in the order a trader reads them:
// delta first because it is what tells them whether the position is directional
// at all, vega next because it says how much of that is the market's, and
// theta last because it is the one that is certain.
var greekRows = []greekRow{
	{"Delta", 0, "how much of a move in the underlying the option takes with it", 4},
	{"Gamma", 0, "how much that delta itself moves", 4},
	{"Theta", 0, "what a day costs the position, as a share of the premium", 4},
	{"Vega", 0, "how much a point of implied volatility is worth", 4},
	{"Rho", 0, "how much a point of interest rates is worth", 4},
}

// GreeksTableOptions configure a GreeksTable.
type GreeksTableOptions struct {
	// Greeks are the caller's sensitivities, and Contract what they are the
	// sensitivities of — which is what a row's third column says.
	Greeks   Greeks
	Contract Contract
	// Height is the table's height; zero takes one row a Greek.
	Height float32
	// State and Scroll keep the table's place between frames.
	State  *ui.ListState
	Scroll *ui.ScrollState
}

// GreeksTable is an option's five sensitivities and what each one is the
// sensitivity *of*.
//
// The third column is the reason this is a table rather than five numbers. A
// reader who does not trade options cannot tell delta from theta, and five
// figures with no names beside them are not a help; a figure with "how much a
// move in the underlying the option takes with it" beside it is.
//
// The implied volatility is not a Greek and is not in this table. It is the
// one figure here that can be *calculated* from the chain, and it belongs
// beside the chain rather than in a table of sensitivities — a reader looking
// for the IV of a strike is looking at that strike's row.
func GreeksTable(c *ui.Context, opts GreeksTableOptions) *ui.Element {
	if opts.Contract.Symbol == "" {
		panic("finance: GreeksTable needs a Contract; sensitivities belong to an instrument " +
			"and not to a column of numbers")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	height := opts.Height
	if height <= 0 {
		height = u * 13 * float32(len(greekRows))
	}

	rows := make([]greekRow, len(greekRows))
	values := []float64{
		opts.Greeks.Delta, opts.Greeks.Gamma, opts.Greeks.Theta,
		opts.Greeks.Vega, opts.Greeks.Rho,
	}
	for i, row := range greekRows {
		rows[i] = greekRow{
			name: row.name, value: values[i], what: row.what, decimals: row.decimals,
		}
	}

	return tableOf(c, tableOptions{
		columns: greekColumns,
		rows:    len(rows),
		name:    "Greeks of " + opts.Contract.Symbol,
		cell: func(row, col int) {
			switch greekColumns[col].ID {
			case "name":
				ui.Text(c, rows[row].name).TextColor(k.Text).
					FontSize(core.FontSize(c, theme.RowSize))
			case "value":
				mono(c, strconv.FormatFloat(rows[row].value, 'f', rows[row].decimals, 64),
					theme.RowSize, greekInk(rows[row].value, k))
			case "what":
				ui.Text(c, rows[row].what).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize)).MaxLines(1)
			}
		},
		key: func(row int) any { return rows[row].name },
		row: func(row int) string {
			return rows[row].name + " " +
				strconv.FormatFloat(rows[row].value, 'f', rows[row].decimals, 64)
		},
		state: opts.State, scroll: opts.Scroll, height: height,
		empty: func() {},
	})
}

// greekInk is a Greek's tone. Positive values in the success tone and negative
// in the danger one, with theta inverted — theta is the cost of holding, so a
// positive theta is a cost and is drawn as one. A reader who has just learned
// that theta is always negative does not need to be told, and one who has not
// would be misled by the sign.
func greekInk(v float64, k theme.Tokens) ui.Color {
	switch {
	case v > 0:
		return k.Success
	case v < 0:
		return k.Danger
	}
	return k.TextMuted
}

// OptionChainOptions configure an OptionChain.
type OptionChainOptions struct {
	// Contracts are the caller's, in the order they should be shown.
	Contracts []Contract
	// Underlying is the instrument the chain is on, and its ticker is
	// required: a chain with no underlying is a list of strikes.
	Underlying string
	// Expiry is the date the chain is for, as the caller's string.
	Expiry string
	// Height is the table's height, and is required.
	Height float32
	// State and Scroll keep the table's place between frames.
	State  *ui.ListState
	Scroll *ui.ScrollState
	// Selected is the contract the keys move from; -1 for none.
	Selected *int
	// Sort is the column the rows are ordered by, in the caller's state.
	Sort *data.Sort
}

// chainColumns are the chain's columns. The strike takes the middle so that
// the calls and the puts either side of it are compared at the same distance
// from the centre — which is what a chain is read for.
var chainColumns = []tableColumn{
	{Title: "Call bid", ID: "callBid", Width: 96, Align: ui.End},
	{Title: "Call ask", ID: "callAsk", Width: 96, Align: ui.End},
	{Title: "Call vol", ID: "callVol", Width: 92, Align: ui.End},
	{Title: "IV", ID: "callIV", Width: 88, Align: ui.End},
	{Title: "Strike", ID: "strike", Width: 112, Align: ui.End},
	{Title: "IV", ID: "putIV", Width: 88, Align: ui.End},
	{Title: "Put vol", ID: "putVol", Width: 92, Align: ui.End},
	{Title: "Put ask", ID: "putAsk", Width: 96, Align: ui.End},
	{Title: "Put bid", ID: "putBid", Width: 96, Align: ui.End},
}

// OptionChain is every strike of one expiry with its calls and puts.
//
// A chain is one row per *strike*, not one row per contract, with the call on
// the left and the put on the right. That is the only arrangement in which a
// reader can see that one strike is expensive on both sides — which is
// exactly the thing a chain exists to show and the thing two separate tables
// cannot show at all.
func OptionChain(c *ui.Context, opts OptionChainOptions) *ui.Element {
	if opts.Underlying == "" {
		panic("finance: OptionChain needs an Underlying; a chain with nothing underneath it is " +
			"a list of strikes")
	}
	if opts.Height <= 0 {
		panic("finance: OptionChain needs a Height; a table with no height grows to fit every " +
			"strike rather than scrolling")
	}
	k := core.Tokens(c)

	// Pair the contracts up by strike: one row per strike, with whatever of
	// each side exists beside it. A row whose other half is missing draws an
	// em-dash rather than shifting, so the strike stays in the middle where
	// the comparison is made.
	strikes := map[float64][2]*Contract{}
	order := []float64{}
	for i := range opts.Contracts {
		c := &opts.Contracts[i]
		pair, seen := strikes[c.Strike]
		if !seen {
			order = append(order, c.Strike)
		}
		if c.Kind == OptionCall {
			pair[0] = c
		} else {
			pair[1] = c
		}
		strikes[c.Strike] = pair
	}
	sortFloats(order)

	return tableOf(c, tableOptions{
		columns: chainColumns,
		rows:    len(order),
		name:    opts.Underlying + " " + opts.Expiry,
		cell: func(row, col int) {
			call, put := strikes[order[row]][0], strikes[order[row]][1]
			switch chainColumns[col].ID {
			case "callBid":
				chainPrice(c, call, true)
			case "callAsk":
				chainPrice(c, call, false)
			case "callVol":
				chainVolume(c, call)
			case "callIV":
				chainIV(c, call)
			case "strike":
				quoteBand(c, func() {
					spreadRow(c, 0, func() {
						mono(c, FormatPrice(order[row]), theme.RowSize, k.Text).Bold()
					})
				})
			case "putIV":
				chainIV(c, put)
			case "putVol":
				chainVolume(c, put)
			case "putAsk":
				chainPrice(c, put, false)
			case "putBid":
				chainPrice(c, put, true)
			}
		},
		key: func(row int) any { return order[row] },
		row: func(row int) string {
			return FormatPrice(order[row]) + " strike"
		},
		state: opts.State, scroll: opts.Scroll, height: opts.Height,
		empty: func() {
			ui.Text(c, "No contracts listed for this expiry").TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.BodySize))
		},
	})
}

func chainPrice(c *ui.Context, o *Contract, bid bool) {
	k := core.Tokens(c)
	if o == nil {
		ui.Text(c, "—").TextColor(k.TextFaint).
			FontSize(core.FontSize(c, theme.RowSize))
		return
	}
	price := o.Bid
	if !bid {
		price = o.Ask
	}
	mono(c, FormatPrice(price), theme.RowSize, k.Text)
}

func chainVolume(c *ui.Context, o *Contract) {
	k := core.Tokens(c)
	if o == nil {
		ui.Text(c, "—").TextColor(k.TextFaint).
			FontSize(core.FontSize(c, theme.RowSize))
		return
	}
	ui.Text(c, FormatVolume(o.Volume)).TextColor(k.TextMuted).
		FontSize(core.FontSize(c, theme.RowSize))
}

func chainIV(c *ui.Context, o *Contract) {
	k := core.Tokens(c)
	if o == nil {
		ui.Text(c, "—").TextColor(k.TextFaint).
			FontSize(core.FontSize(c, theme.RowSize))
		return
	}
	if o.Greeks.Implied <= 0 {
		ui.Text(c, "—").TextColor(k.TextFaint).
			FontSize(core.FontSize(c, theme.RowSize))
		return
	}
	mono(c, strconv.FormatFloat(o.Greeks.Implied*100, 'f', 1, 64)+"%", theme.RowSize, k.Text)
}

// sortFloats is a sort by hand, for the handful of strikes on one chain. An
// insertion sort is four lines and sorts.Slice's closure is longer and
// allocates a reflect-backed swapper to do it.
func sortFloats(vs []float64) {
	for i := 1; i < len(vs); i++ {
		for j := i; j > 0 && vs[j] < vs[j-1]; j-- {
			vs[j], vs[j-1] = vs[j-1], vs[j]
		}
	}
}
