package pages

import (
	"fmt"
	"math"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/finance"
	"github.com/HycJack/MintUI/ui/showcase"
	"github.com/HycJack/MintUI/ui/theme"
)

// Demo state outlives the frame: every demo below hands its component
// a pointer, and a pointer into a frame-local is a click the next
// frame undoes — a tab that will not switch, a dropdown that snaps
// shut, a slider that springs back.
var (
	listSel     = 1
	watchSel    = 2
	search      = "riv"
	searchSel   = -1
	overviewSel = 0
	bookSel     = -1
	limit       = 101.20
	size        = 100.0
	marketSize  = 50.0
	marketPrice = 101.25
	orderSel    = 0
	open        = true
	posSel      = 0
	baseline    = 40000.0
	chainSel    = 1
	leverage    = 4.0
)

// The page is one trading day: what an instrument is doing now, what the book
// looks like underneath it, an order going in, what it left behind, and the
// risk figures a desk watches while all of that is happening.
//
// Every figure on it is a number the caller computed. This package formats and
// draws; it does not fetch a quote, size an order or hold a position, so the
// page's "live" prices are written out in the source where a reader can see
// exactly what they are made of — and a screenshot of this page is the same
// twice.

func init() {
	showcase.Register(showcase.Page{
		Package: "finance",
		Title:   "ui/finance — 报价、盘口、订单、持仓、风险",
		Note:    "价格、盘口、成交、订单、持仓、期权、风险、日历",
		Width:   1000,
		Height:  6300,
		Want: []string{
			// quotes
			"Riverside Clinic", "QuoteCard — 价格最大，涨跌在下面，不在旁边",
			"+1.25%", "-1.80%", "0.00%", "QuoteList — 表格形式：一行读到底", "Watchlist — 列表形式：价格贴在右边", "Market overview",
			// A tape is a painter: its Label is read out by a screen reader and
			// is never words on the page, and what it does paint is figures
			// scrolling past, not a name. The line above it is what a person
			// reads there, so that is what it promises.
			"TickerTape — 横向滚动的一行数字",
			"RIVR on L", "Euro / dollar",
			// the book
			"Order book", "Bid", "Ask", "Spread 0.03", "bid 101.22, 4,200 shares",
			"ask 101.25, 3,100 shares", "Depth ladder", "Time and sales",
			"Trade history",
			// orders
			"OrderEntry — 名义金额写在数量下面", "Buy 100 RIVR", "Sell 50 RIVR", "Order value",
			"Available after", "Orders", "Partially filled", "Cancelled",
			"Order history", "Buy 100 at 101.22", "Sell 100 at 101.25",
			"OrderConfirm over a desk",
			// positions
			"Shares", "Avg", "Last", "Value", "Return", "Portfolio", "Total value",
			"68,960.00", "Portfolio allocation", "Cash", "Allocation",
			"Profit and loss", "Realised", "Unrealised", "Performance",
			"Benchmark", "USD axis",
			// options
			"RIVR 2025-03-21", "Call bid", "Put ask", "Greeks of RIVR",
			"how much that delta itself moves", "Payoff: break-even at 104.40, break-even 104.40",
			"Short put: break-even at 97.90, break-even 97.90",
			// risk
			"Gauge: 2.1× your equity", "Gauge: over the firm's limit",
			"Bullet: call volume 2.1× the average", "Bearish", "Neutral", "Bullish",
			"4×", "buying power", "equity", "Margin 62.00% used", "Maintenance at 25.00%",
			"LSE · Open", "LSE · Pre-market", "LSE · After hours", "LSE · Closed",
			"14:06:00", "London", "Trading session, 71.76% through the session",
			// calendars
			"Earnings", "Economic calendar", "Riverside Clinic Q3", "Bank rate decision",
			"CPI, year on year",
			// the pure functions
			"FormatPrice(1234.5)", "1,234.50", "FormatMoney(1234.5, \"$\")",
			"$1,234.50", "FormatChange(101.25, 100)", "+1.25%  · Success",
			"FormatSigned(1240)", "+1,240.00", "FormatCompact(1234500)", "1.2M",
			"FormatVolume(1500)", "1.50K", "FormatSize(0.5)", "0.5000",
			"BestBid(bids)", "101.22  (true)  ← 空的一侧是 false", "BestAsk(asks)",
			"Spread(bids, asks)", "0.03  0.0003  (true)  ← 交叉的盘口是负价差",
			"Levels(bids, asks)", "16 行，Total 从最好的一边累加",
		},
		Render: func(c *ui.Context) {
			financePage(c)
		},
		// ClockDriven: the TickerTape in the quotes section takes its scroll
		// offset from the painter's own clock, so two renders of this page are
		// two frames of a tape in motion and a tape held still would be showing
		// something the component never does. The gate draws with reduced
		// motion, where the tape rests, so this says where the page would go
		// if that were ever not so.
		ClockDriven: true,
	})
}

func financePage(c *ui.Context) {
	financeQuoteSection(c)
	financeBookSection(c)
	financeOrderSection(c)
	financePositionSection(c)
	financeOptionSection(c)
	financeRiskSection(c)
	financeCalendarSection(c)
	financePureSection(c)
}

// ── the day's data ─────────────────────────────────────────────────────────

func financeBids() []finance.Price {
	return []finance.Price{
		{Price: 101.22, Size: 4200}, {Price: 101.21, Size: 1800},
		{Price: 101.20, Size: 6300}, {Price: 101.19, Size: 2500},
		{Price: 101.18, Size: 9100}, {Price: 101.15, Size: 1200},
		{Price: 101.10, Size: 5400}, {Price: 101.05, Size: 300},
	}
}

func financeAsks() []finance.Price {
	return []finance.Price{
		{Price: 101.25, Size: 3100}, {Price: 101.26, Size: 1500},
		{Price: 101.28, Size: 4700}, {Price: 101.30, Size: 2200},
		{Price: 101.32, Size: 6800}, {Price: 101.35, Size: 900},
		{Price: 101.40, Size: 3900}, {Price: 101.45, Size: 1600},
	}
}

func financeQuotes() []finance.WatchQuote {
	return []finance.WatchQuote{
		{Symbol: finance.Symbol{Ticker: "RIVR", Exchange: "L", Name: "Riverside Clinic"},
			Now: 101.25, Previous: 100.00, Volume: 1284000, Group: "Healthcare"},
		{Symbol: finance.Symbol{Ticker: "DNTL", Exchange: "L", Name: "Northgate Dental"},
			Now: 44.80, Previous: 45.62, Volume: 210300},
		{Symbol: finance.Symbol{Ticker: "HRBR", Exchange: "L", Name: "Harbour Cafe"},
			Now: 12.40, Previous: 12.40, Volume: 84200},
		{Symbol: finance.Symbol{Ticker: "MAPL", Exchange: "N", Name: "Maple Bakery"},
			Now: 7.15, Previous: 6.98, Volume: 1200000, Group: "Hospitality"},
		{Symbol: finance.Symbol{Ticker: "EURS", Exchange: "N", Name: "Euro / dollar"},
			Now: 1.0842, Previous: 1.0910, Volume: 998000},
	}
}

func financePositions() []finance.Position {
	return []finance.Position{
		{ID: "p1", Instrument: "RIVR", Shares: 400, Average: 96.40, Last: 101.25,
			DayHigh: 101.60, DayLow: 100.80, Unrealised: 1940, Realised: 320,
			Weight: 0.52, Sector: "Healthcare"},
		{ID: "p2", Instrument: "DNTL", Shares: 250, Average: 41.10, Last: 44.80,
			DayHigh: 45.90, DayLow: 44.20, Unrealised: 925, Weight: 0.14,
			Sector: "Healthcare"},
		{ID: "p3", Instrument: "MAPL", Shares: 1200, Average: 6.20, Last: 7.15,
			DayHigh: 7.20, DayLow: 6.95, Unrealised: 1140, Weight: 0.21,
			Sector: "Hospitality"},
		{ID: "p4", Instrument: "HRBR", Shares: -300, Average: 12.90, Last: 12.40,
			DayHigh: 12.85, DayLow: 12.30, Unrealised: 150, Weight: -0.04,
			Sector: "Hospitality"},
	}
}

// fig is one figure in the monospaced stack, right-aligned — the rule every
// price column in this library is drawn by, kept here so the page's own
// numbers obey it too.
func fig(c *ui.Context, s string, col ui.Color, size float32) *ui.Element {
	return ui.Text(c, s).TextColor(col).Font(finance.MonoFont).
		FontSize(core.FontSize(c, size)).TextAlign(ui.End).Shrink(0)
}

func figMuted(c *ui.Context, s string) *ui.Element {
	return fig(c, s, core.Tokens(c).TextMuted, theme.RowSize)
}

func figHead(c *ui.Context, s string) *ui.Element {
	return fig(c, s, core.Tokens(c).TextFaint, theme.CaptionSize)
}

// ── quotes ─────────────────────────────────────────────────────────────────

func financeQuoteSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "报价 · PriceText / PriceChangeBadge / SymbolBadge / QuoteCard / QuoteList / Watchlist")

	quotes := financeQuotes()

	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).WidthPercent(39).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "QuoteCard — 价格最大，涨跌在下面，不在旁边")
			finance.QuoteCard(c, finance.QuoteOptions{
				Symbol: quotes[0].Symbol, Now: quotes[0].Now, Previous: quotes[0].Previous,
				DayHigh: 101.60, DayLow: 100.80, DayVolume: 1284000,
				Currency: "USD", Interval: "1D",
				Sparkline: financeSpark(101.25, 14),
			})
			showcase.Field(c, "QuoteCard — 延迟报价、选中、第二个标的")
			ui.Row(c).Gap(unit(c, 2)).Children(func() {
				finance.QuoteCard(c, finance.QuoteOptions{
					Symbol: quotes[2].Symbol, Now: quotes[2].Now,
					Previous: quotes[2].Previous, Currency: "USD",
					Delayed: true, Interval: "1W", Selected: true,
				})
				finance.QuoteCard(c, finance.QuoteOptions{
					Symbol: quotes[4].Symbol, Now: quotes[4].Now,
					Previous: quotes[4].Previous, Currency: "USD", Interval: "YTD",
				})
			})
		})

		ui.Column(c).WidthPercent(59).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "PriceText、PriceChangeBadge、SymbolBadge — 三个最基本的一件")
			ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.End).Children(func() {
				fig(c, "101.25", k.Text, theme.StatSize).Label("Last price, 101.25")
				fig(c, "$101.25", k.Text, theme.StatSize).Label("Last price, in dollars")
				ui.Column(c).Gap(unit(c, 0.5)).Shrink(0).Children(func() {
					finance.PriceChangeBadge(c, finance.PriceChangeBadgeOptions{
						Now: 101.25, Previous: 100,
					})
					finance.PriceChangeBadge(c, finance.PriceChangeBadgeOptions{
						Now: 44.80, Previous: 45.62, ShowAbsolute: true, Absolute: -0.82,
					})
					finance.PriceChangeBadge(c, finance.PriceChangeBadgeOptions{
						Now: 12.40, Previous: 12.40,
					})
				})
				ui.Column(c).Gap(unit(c, 1)).Children(func() {
					finance.SymbolBadge(c, finance.SymbolBadgeOptions{
						Symbol: quotes[0].Symbol,
					})
					finance.SymbolBadge(c, finance.SymbolBadgeOptions{
						Symbol: quotes[0].Symbol, ShowExchange: true,
					})
				})
			})

			showcase.Field(c, "QuoteList — 表格形式：一行读到底")
			finance.QuoteList(c, finance.QuoteListOptions{
				Quotes: quotes, Selected: &listSel, Height: 200,
			})

			showcase.Field(c, "Watchlist — 列表形式：价格贴在右边")
			finance.Watchlist(c, finance.WatchlistOptions{
				Quotes: quotes, Selected: &watchSel, Height: 200, Grouped: true,
			})
		})
	})

	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).WidthPercent(34).Shrink(0).Gap(unit(c, 1.5)).Children(func() {
			showcase.Field(c, "SymbolSearch — 关键词也参与匹配")
			finance.SymbolSearch(c, finance.SymbolSearchOptions{
				Query:       &search,
				Instruments: financeInstruments(),
				Selected:    &searchSel, Results: 6,
			})
			ui.Text(c, "query = \"riv\" · 屋顶、水管这类词也找得到").TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize))
		})
		ui.Column(c).WidthPercent(64).Shrink(0).Gap(unit(c, 1.5)).Children(func() {
			showcase.Field(c, "MarketOverview — 每个标的一张卡，不是四列表")
			finance.MarketOverview(c, finance.MarketOverviewOptions{
				Instruments: financeInstruments(), Quotes: quotes,
				Cols: 4, Selected: &overviewSel, Label: "Market overview",
			})
			showcase.Field(c, "TickerTape — 横向滚动的一行数字")
			finance.TickerTape(c, finance.TickerTapeOptions{
				Quotes: quotes[:3], Label: "Ticker tape", Repeat: true, Height: 40,
			})
		})
	})
}

func financeInstruments() []finance.Instrument {
	return []finance.Instrument{
		{Symbol: finance.Symbol{Ticker: "RIVR", Exchange: "L", Name: "Riverside Clinic"},
			Keywords: []string{"roofing", "clinic", "heating"}},
		{Symbol: finance.Symbol{Ticker: "DNTL", Exchange: "L", Name: "Northgate Dental"},
			Keywords: []string{"dental", "clinic"}},
		{Symbol: finance.Symbol{Ticker: "HRBR", Exchange: "L", Name: "Harbour Cafe"},
			Keywords: []string{"cafe", "hospitality"}},
		{Symbol: finance.Symbol{Ticker: "MAPL", Exchange: "N", Name: "Maple Bakery"},
			Keywords: []string{"bakery", "food"}},
		{Symbol: finance.Symbol{Ticker: "EURS", Exchange: "N", Name: "Euro / dollar"},
			Keywords: []string{"fx", "currency"}},
		{Symbol: finance.Symbol{Ticker: "RIVX", Exchange: "N", Name: "Riverside warrants"},
			Keywords: []string{"warrant", "roofing"}},
	}
}

func financeSpark(base float64, n int) []finance.Point {
	out := make([]finance.Point, n)
	v := base * 0.97
	for i := range out {
		// A deterministic walk: the same sparkline on the next run, so two
		// screenshots of this page can be compared.
		v += math.Sin(float64(i)*1.7) * base * 0.004
		out[i] = finance.Point{
			X: float64(i), Y: v,
		}
	}
	return out
}

// ── the book ───────────────────────────────────────────────────────────────

func financeBookSection(c *ui.Context) {
	showcase.Section(c, "盘口 · OrderBook / DepthLadder / BidAskBar / SpreadIndicator / TimeAndSales / TradeHistory")

	bids, asks := financeBids(), financeAsks()

	// Two rows of two rather than four across: a book, a ladder and a tape all
	// put three numbers side by side per row, and a third of this page is not
	// enough for three numbers to sit in without crossing.
	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "OrderBook — 最好的价格在中间")
			finance.OrderBook(c, finance.OrderBookOptions{
				Bids: bids, Asks: asks, Rows: 6, Height: 300,
				Selected: &bookSel, ShowTotals: true, Label: "Order book",
			})
		})
		ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "DepthLadder — 价格分箱，两边各一把尺")
			finance.DepthLadder(c, finance.DepthLadderOptions{
				Bids: bids, Asks: asks, Levels: 6, Bins: 16,
				Height: 300, Label: "Depth ladder",
			})
		})
	})
	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "TimeAndSales — 每一笔：时间、价格、数量、方向")
			finance.TimeAndSales(c, finance.TimeAndSalesOptions{
				Trades: financeTrades(), Height: 300,
				Highlight: []int{2}, Label: "Time and sales",
			})
		})
		ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "Levels — 一侧的书加上累计量，全是算出来的")
			financeLevelsTable(c, bids, asks)
		})
	})
	// The one table on this page that needs the whole width: seven columns of
	// figures, and a fee column squeezed to half a word is not a fee column.
	showcase.Field(c, "TradeHistory — 成交与费用，七列要整页才排得开")
	finance.TradeHistory(c, finance.TradeHistoryOptions{
		Fills: financeFills(), Height: 220,
	})

	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "BidAskBar — 中间那一段就是价差")
			finance.BidAskBar(c, finance.BidAskBarOptions{Bid: 101.22, Ask: 101.25})
			finance.BidAskBar(c, finance.BidAskBarOptions{Bid: 12.38, Ask: 12.42})
			showcase.Field(c, "SpreadIndicator — 分数在前，金额在后")
			finance.SpreadIndicator(c, finance.SpreadIndicatorOptions{
				Bids: bids, Asks: asks, ShowBars: true, Height: 90,
			})
			finance.SpreadIndicator(c, finance.SpreadIndicatorOptions{
				Bids: asks, Asks: bids, Height: 60,
			})
		})
		ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "SpreadIndicator — 第二个是交叉的盘口，负价差照实写")
			finance.SpreadIndicator(c, finance.SpreadIndicatorOptions{
				Bids: asks, Asks: bids, Height: 60,
			})
		})
	})
}

func financeTrades() []finance.Trade {
	return []finance.Trade{
		{Price: 101.25, Size: 1200, Side: finance.SideAsk, When: "14:06:31", Venue: "LSE"},
		{Price: 101.24, Size: 300, Side: finance.SideBid, When: "14:06:28", Venue: "LSE"},
		{Price: 101.24, Size: 8400, Side: finance.SideAsk, When: "14:06:24", Venue: "CHIX"},
		{Price: 101.22, Size: 600, Side: finance.SideBid, When: "14:06:20", Venue: "LSE"},
		{Price: 101.20, Size: 250, Side: finance.SideAsk, When: "14:06:19"},
		{Price: 101.18, Size: 5100, Side: finance.SideBid, When: "14:06:12", Venue: "LSE"},
		{Price: 101.16, Size: 900, Side: finance.SideAsk, When: "14:06:08", Venue: "LSE"},
	}
}

func financeFills() []finance.Fill {
	return []finance.Fill{
		{ID: "f1", When: "14:02:11", Instrument: "RIVR", Side: finance.SideBid,
			Price: 101.20, Size: 200, Fee: -1.20, Venue: "LSE"},
		{ID: "f2", When: "13:41:02", Instrument: "MAPL", Side: finance.SideAsk,
			Price: 7.10, Size: 1200, Fee: 2.14, Venue: "LSE"},
		{ID: "f3", When: "11:22:48", Instrument: "RIVR", Side: finance.SideAsk,
			Price: 100.95, Size: 100, Fee: 0.61, Venue: "CHIX"},
		{ID: "f4", When: "09:35:00", Instrument: "DNTL", Side: finance.SideBid,
			Price: 44.55, Size: 250, Fee: -0.84, Venue: "LSE"},
	}
}

func financeLevelsTable(c *ui.Context, bids, asks []finance.Price) {
	k := core.Tokens(c)
	levels := finance.Levels(bids, asks)
	show := levels
	if len(show) > 8 {
		show = show[:8]
	}
	ui.Column(c).FillWidth().Gap(1).Children(func() {
		ui.Row(c).FillWidth().Gap(unit(c, 2)).Children(func() {
			ui.Text(c, "Side").TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize)).Grow(1)
			figHead(c, "PRICE").Width(78).Shrink(0)
			figHead(c, "SIZE").Width(70).Shrink(0)
			figHead(c, "TOTAL").Width(76).Shrink(0)
			figHead(c, "ORDERS").Grow(1)
		})
		for _, l := range show {
			side := "bid"
			if l.Side == finance.SideAsk {
				side = "ask"
			}
			ui.Row(c).FillWidth().Gap(unit(c, 2)).Children(func() {
				ui.Text(c, side).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.RowSize)).Grow(1)
				fig(c, finance.FormatPrice(l.Price), k.Text, theme.RowSize).Width(78).Shrink(0)
				fig(c, finance.FormatVolume(int(l.Size)), k.Text, theme.RowSize).Width(70).Shrink(0)
				fig(c, finance.FormatVolume(int(l.Total)), k.TextMuted, theme.RowSize).
					Width(76).Shrink(0)
				fig(c, fmt.Sprintf("%d", l.Orders), k.TextFaint, theme.RowSize).Grow(1)
			})
		}
	})
}

// ── orders ─────────────────────────────────────────────────────────────────

func financeOrderSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "订单 · OrderEntry / OrderConfirm / OrderTable / OrderHistory / QuickTradeButtons")

	bids := financeBids()

	// The two tickets beside each other, the two tables on the full width
	// below: an orders table is seven columns of figures, and a third of this
	// page is not enough for seven of them to sit in.
	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "OrderEntry — 名义金额写在数量下面")
			finance.OrderEntry(c, finance.OrderEntryOptions{
				Instrument: finance.Symbol{Ticker: "RIVR", Exchange: "L", Name: "Riverside Clinic"},
				Side:       finance.SideBid, Kind: finance.OrderLimit,
				Price: &limit, Size: &size, Last: 101.25, Maximum: 50000,
				Currency: "USD", Label: "Order ticket",
			})
			showcase.Field(c, "OrderEntry — 市价单，没有限价")
			finance.OrderEntry(c, finance.OrderEntryOptions{
				Instrument: finance.Symbol{Ticker: "RIVR", Exchange: "L", Name: "Riverside Clinic"},
				Side:       finance.SideAsk, Kind: finance.OrderMarket,
				Price: &marketPrice, Size: &marketSize, Last: 101.25, Currency: "USD",
				Label: "Order ticket",
			})
		})

		ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "QuickTradeButtons — 价格就写在按钮上")
			finance.QuickTradeButtons(c, finance.QuickTradeButtonsOptions{
				Size: 100,
				Instrument: finance.Symbol{Ticker: "RIVR", Exchange: "L",
					Name: "Riverside Clinic"},
				Bid: 101.22, Ask: 101.25, Side: finance.SideBid,
			})
			ui.Text(c, "两个按钮上的价格就是这一行的买价和卖价").TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize))
			showcase.Field(c, "OrderConfirm — 最后一道屏幕，浮在整窗之上")
			financeOrderConfirmShot(c, bids)
		})
	})

	showcase.Field(c, "OrderTable — 工作单，以及成交了多少")
	finance.OrderTable(c, finance.OrderTableOptions{
		Orders: financeOrders(), Height: 232, Selected: &orderSel,
	})
	showcase.Field(c, "OrderHistory — 今天每一张单子，按时间；查的那一屏和盯的那一屏不是同一屏")
	finance.OrderHistory(c, finance.OrderHistoryOptions{
		Orders: financeOrders(), Height: 232,
	})
}

func financeOrders() []finance.Order {
	return []finance.Order{
		{ID: "o1", Instrument: "RIVR", Kind: finance.OrderLimit, Side: finance.SideBid,
			Price: 101.20, Size: 200, Filled: 200, State: "Filled",
			Tone: core.Success, When: "14:02:11"},
		{ID: "o2", Instrument: "MAPL", Kind: finance.OrderLimit, Side: finance.SideAsk,
			Price: 7.18, Size: 1200, Filled: 400, State: "Partially filled",
			Tone: core.Warning, When: "13:41:02"},
		{ID: "o3", Instrument: "RIVR", Kind: finance.OrderStop, Side: finance.SideAsk,
			Price: 99.80, Size: 100, Filled: 0, State: "Working", When: "11:22:48"},
		{ID: "o4", Instrument: "DNTL", Kind: finance.OrderLimit, Side: finance.SideBid,
			Price: 44.55, Size: 250, Filled: 0, State: "Cancelled",
			Tone: core.Danger, When: "09:35:00"},
	}
}

// financeOrderConfirmShot draws an open OrderConfirm into a window of its own
// and puts it on the page as a picture: a confirmation is an AlertDialog, and
// an alert's scrim over this page would dim every other component on it.
func financeOrderConfirmShot(c *ui.Context, bids []finance.Price) {
	mode := core.Light
	if core.IsDark(c) {
		mode = core.Dark
	}
	bid, _ := finance.BestBid(bids)
	shot := ui.NewTester(func(inner *ui.Context) {
		core.Use(inner, core.Settings{Mode: mode})
		// Stand-in for the desk the confirmation is raised over.
		ui.Box(inner).Fill().Background(core.Tokens(inner).Background).Children(func() {
			ui.Column(inner).Fill().Padding(core.Density(inner).Unit() * 3).
				Gap(core.Density(inner).Unit()).Children(func() {
				ui.Text(inner, "Order ticket").TextColor(core.Tokens(inner).Text).
					FontSize(core.FontSize(inner, theme.RowSize)).Bold()
				fig(inner, finance.FormatPrice(101.20), core.Tokens(inner).TextMuted,
					theme.RowSize)
			})
		})
		finance.OrderConfirm(inner, finance.OrderConfirmOptions{
			Open: &open, Instrument: finance.Symbol{Ticker: "RIVR", Exchange: "L",
				Name: "Riverside Clinic"},
			Side: finance.SideBid, Kind: finance.OrderLimit,
			Price: 101.20, Size: 100, Notional: 10120, Fee: 1.20,
			Last: bid, Currency: "USD", Title: "Order confirm",
		})
	}, 460, 380).Image()

	display.Image(c, ui.NewBitmap(shot), display.ImageOptions{
		Width: 460, Height: 380, Radius: theme.SmallRadius,
		Name: "OrderConfirm over a desk",
	})
}

// ── positions ──────────────────────────────────────────────────────────────

func financePositionSection(c *ui.Context) {
	showcase.Section(c, "持仓 · PositionTable / PortfolioSummary / PnLDisplay / AssetAllocationChart / PerformanceChart")

	positions := financePositions()

	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "PositionTable — 七样数字对齐，空头保留负号")
			finance.PositionTable(c, finance.PositionTableOptions{
				Positions: positions, Selected: &posSel, Height: 210,
				ShowReturn: true,
			})
			showcase.Field(c, "PortfolioSummary 与 PnLDisplay")
			finance.PortfolioSummary(c, finance.PortfolioSummaryOptions{
				Positions: positions, Cash: 12400, DayChange: 412,
				Currency: "USD", Inception: "March 2023", ShowAllocation: true,
				Width: 300, Height: 200,
			})
			finance.PnLDisplay(c, finance.PnLOptions{
				Unrealised: 3790, Realised: 1240, Currency: "USD",
			})
		})

		ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "AssetAllocationChart — 现金也在环里")
			finance.AssetAllocationChart(c, finance.AssetAllocationOptions{
				Positions: positions, Cash: 12400, Width: 300, Height: 220,
				Label: "Allocation",
			})
			showcase.Field(c, "PerformanceChart — 与基准比，不是价格图")
			finance.PerformanceChart(c, finance.PerformanceChartOptions{
				Baseline: &baseline, Height: 220, Grid: true,
				Label: "Performance", YLabel: "USD",
				Labels: []string{"Mar", "Jun", "Sep", "Dec", "Mar", "Jun"},
				Series: []finance.PerformanceSeries{
					{Name: "Account", Points: financeSeries(6, 52400)},
					{Name: "Benchmark", Points: financeSeries(6, 47100)},
				},
			})
		})
	})
}

func financeSeries(n int, end float64) []finance.Point {
	out := make([]finance.Point, n)
	for i := range out {
		t := float64(i) / float64(n-1)
		out[i] = finance.Point{X: float64(i), Y: end*0.82 + t*(end-end*0.82)}
	}
	return out
}

// ── options ────────────────────────────────────────────────────────────────

func financeOptionSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "期权 · OptionChain / GreeksTable / PayoffDiagram")

	contracts := financeContracts()

	// The chain spans the page on its own: nine columns of prices want the
	// whole width, and in a half column its middle — the strike, the one
	// column every reader starts from — sat past the viewport's edge.
	showcase.Field(c, "OptionChain — 一行一个行权价，左看涨右看跌")
	finance.OptionChain(c, finance.OptionChainOptions{
		Contracts: contracts, Underlying: "RIVR", Expiry: "2025-03-21",
		Height: 280, Selected: &chainSel,
	})

	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "GreeksTable — 五个敏感度，各自是什么的敏感度")
			finance.GreeksTable(c, finance.GreeksTableOptions{
				Greeks:   finance.Greeks{Delta: 0.42, Gamma: 0.018, Theta: -0.06, Vega: 0.11, Rho: 0.09, Implied: 0.24},
				Contract: contracts[1], Height: 210,
			})
			showcase.Field(c, "PayoffDiagram — 盈亏平衡点从现价和行权价算出来")
			finance.PayoffDiagram(c, finance.PayoffOptions{
				Spot: 101.25, Strike: 102, Expiry: "2025-03-21",
				Premium: 2.40, Long: true, Steps: 41, Grid: true,
				Height: 220, Label: "Payoff: break-even at 104.40",
			})
		})

		ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "PayoffDiagram — 卖出的那一张，曲线上下翻过来")
			finance.PayoffDiagram(c, finance.PayoffOptions{
				Spot: 101.25, Strike: 100, Expiry: "2025-03-21",
				Premium: 2.10, Put: true, Long: false, Steps: 41, Grid: true,
				Height: 200, Label: "Short put: break-even at 97.90",
			})
			ui.Text(c, "Break-even = strike ± premium，方向由 long/short 决定").TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize))
		})
	})
}

func financeContracts() []finance.Contract {
	mk := func(strike float64) finance.Contract {
		return finance.Contract{
			Symbol: "RIVR", Underlying: "RIVR", Expiry: "2025-03-21", Strike: strike,
			Greeks: finance.Greeks{Delta: 0.42, Gamma: 0.018, Theta: -0.06,
				Vega: 0.11, Rho: 0.09, Implied: 0.24},
		}
	}
	out := make([]finance.Contract, 0, 6)
	for _, s := range []float64{100, 101, 102, 103, 104} {
		call := mk(s)
		call.Kind = finance.OptionCall
		call.Bid, call.Ask, call.Last, call.Volume = s-101.0+2.4, s-101.0+2.6, s-101.0+2.5, 120
		put := mk(s)
		put.Kind = finance.OptionPut
		put.Bid, put.Ask, put.Last, put.Volume = 2.6-(s-101.0), 2.8-(s-101.0), 2.7-(s-101.0), 240
		out = append(out, call, put)
	}
	return out
}

// ── risk ───────────────────────────────────────────────────────────────────

func financeRiskSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "风险 · RiskMeter / SentimentGauge / LeverageSlider / MarginIndicator / MarketStatus / TradingSessionClock")

	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).WidthPercent(33).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "RiskMeter — 一个数字和一个上限比")
			finance.RiskMeter(c, finance.RiskMeterOptions{
				Value: 62, Max: 100, Unit: "%", Caption: "2.1× your equity",
				Label: "Risk",
				Bands: []finance.RiskBand{
					{To: 50, Name: "Low"}, {To: 75, Name: "Elevated"},
					{To: 100, Name: "High"},
				},
			})
			showcase.Field(c, "RiskMeter — 越过了 120% 的线")
			finance.RiskMeter(c, finance.RiskMeterOptions{
				Value: 120, Max: 100, Unit: "%", Label: "Risk",
				Caption: "over the firm's limit", Bands: []finance.RiskBand{
					{To: 50, Name: "Low"}, {To: 100, Name: "At the limit"},
					{To: 150, Name: "Over"},
				},
			})
			showcase.Field(c, "SentimentGauge — 以零为中心的双向刻度")
			finance.SentimentGauge(c, finance.SentimentOptions{
				Score: 0.35, Caption: "call volume 2.1× the average", Label: "Sentiment",
				Bands: []finance.SentimentBand{
					{To: -0.5, Name: "Bearish"}, {To: 0, Name: "Neutral"},
					{To: 0.5, Name: "Mild"}, {To: 1, Name: "Bullish"},
				},
				Size: 320,
			})
		})

		ui.Column(c).WidthPercent(32).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "LeverageSlider — 倍数下面写清楚是多少钱")
			finance.LeverageSlider(c, finance.LeverageSliderOptions{
				Leverage: &leverage, Max: 10, BuyingPower: 160000, Equity: 40000,
				Warning: 6, Label: "Leverage",
			})
			showcase.Field(c, "MarginIndicator — 维持保证金是同一条杠上的一个记号")
			finance.MarginIndicator(c, finance.MarginOptions{
				Used: 0.62, Available: 0.38, Maintenance: 0.25,
				Excess: 0.18, ExcessLabel: "after a 5% move",
			})
			finance.MarginIndicator(c, finance.MarginOptions{
				Used: 0.31, Available: 0.69,
			})
		})

		ui.Column(c).WidthPercent(31).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "MarketStatus — 一个点加一个词")
			for _, s := range []finance.Session{
				finance.SessionOpen, finance.SessionPreMarket,
				finance.SessionPostMarket, finance.SessionClosed,
			} {
				finance.MarketStatus(c, finance.MarketStatusOptions{
					Session: s, Exchange: "LSE", Since: "since 08:00",
				})
			}
			showcase.Field(c, "TradingSessionClock — 一天里已经走完的那一段")
			now := time.Date(2024, 11, 8, 14, 6, 0, 0, time.UTC)
			finance.TradingSessionClock(c, finance.TradingSessionClockOptions{
				Now:    now,
				Opens:  time.Date(2024, 11, 8, 8, 0, 0, 0, time.UTC),
				Closes: time.Date(2024, 11, 8, 16, 30, 0, 0, time.UTC),
				Breaks: [][2]time.Time{{
					time.Date(2024, 11, 8, 12, 0, 0, 0, time.UTC),
					time.Date(2024, 11, 8, 13, 0, 0, 0, time.UTC),
				}},
				Session: finance.SessionOpen, Zone: "London",
				Label: "Trading session", Height: 28,
			})
			ui.Text(c, "14:06 · 开盘已过 6 小时 06 分，午休那一小时是空的").TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize))
		})
	})
}

// ── calendars ──────────────────────────────────────────────────────────────

func financeCalendarSection(c *ui.Context) {
	showcase.Section(c, "日历 · EarningsCalendar / EconomicCalendar")

	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "EarningsCalendar — 预期、实际、去年")
			finance.EarningsCalendar(c, finance.EarningsCalendarOptions{
				Events: financeEvents(), Height: 280,
			})
		})
		ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "EconomicCalendar — 同一张表，换一套列")
			finance.EconomicCalendar(c, finance.EconomicCalendarOptions{
				Events: financeEvents(), Height: 280,
			})
		})
	})
}

func financeEvents() []finance.Event {
	return []finance.Event{
		{Title: "Riverside Clinic Q3", When: "12 Nov", Impact: 0.8,
			Forecast: 1.24, Actual: 1.31, Previous: 1.18, Released: true},
		{Title: "Northgate Dental Q3", When: "14 Nov", Impact: 0.5,
			Forecast: 0.62, Previous: 0.58, Released: false},
		{Title: "Harbour Cafe Q3", When: "19 Nov", Impact: 0.3,
			Forecast: 0.11, Previous: 0.12, Released: false},
		{Title: "Bank rate decision", When: "21 Nov", Impact: 1,
			Forecast: 4.75, Previous: 4.75, Released: false},
		{Title: "CPI, year on year", When: "22 Nov", Impact: 0.9,
			Forecast: 2.6, Previous: 2.4, Released: false},
	}
}

// ── the pure functions ─────────────────────────────────────────────────────

func financePureSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "纯函数 · 一张交易屏最容易搞错的那几个")

	bids, asks := financeBids(), financeAsks()
	bid, hasBid := finance.BestBid(bids)
	ask, hasAsk := finance.BestAsk(asks)
	gap, fraction, ok := finance.Spread(bids, asks)
	change, tone := finance.FormatChange(101.25, 100)
	flat, flatTone := finance.FormatChange(12.40, 12.40)
	down, downTone := finance.FormatChange(44.80, 45.62)

	line := func(call string, value string, mono bool, col ui.Color) {
		ui.Row(c).FillWidth().Gap(unit(c, 2)).AlignItems(ui.Center).Children(func() {
			ui.Text(c, call).TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).Width(220).Shrink(0)
			if mono {
				fig(c, value, col, theme.RowSize).Grow(1)
			} else {
				ui.Text(c, value).TextColor(col).
					FontSize(core.FontSize(c, theme.RowSize)).Grow(1)
			}
		})
	}

	ui.Row(c).FillWidth().Gap(unit(c, 4)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 1)).Children(func() {
			line("FormatPrice(1234.5)", finance.FormatPrice(1234.5), true, k.Text)
			line("FormatPrice(-1234.5)", finance.FormatPrice(-1234.5), true, k.Text)
			line("FormatMoney(1234.5, \"$\")", finance.FormatMoney(1234.5, "$"), true, k.Text)
			line("FormatSigned(1240)", finance.FormatSigned(1240), true, k.Text)
			line("FormatSigned(-1240)", finance.FormatSigned(-1240), true, k.Text)
			line("FormatSigned(0)", finance.FormatSigned(0)+"  ← 零没有正负号", true, k.Text)
			line("FormatCompact(1234500)", finance.FormatCompact(1234500), true, k.Text)
			line("FormatCompact(999)", finance.FormatCompact(999), true, k.Text)
			line("FormatVolume(999)", finance.FormatVolume(999), true, k.Text)
			line("FormatVolume(1500)", finance.FormatVolume(1500), true, k.Text)
			line("FormatVolume(999999)", finance.FormatVolume(999999), true, k.Text)
			line("FormatSize(0.5)", finance.FormatSize(0.5), true, k.Text)
			line("FormatSize(1200)", finance.FormatSize(1200), true, k.Text)
			line("FormatSize(0)", finance.FormatSize(0)+"  ← 没有持仓就是 0", true, k.Text)
		})
		ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 1)).Children(func() {
			line("FormatChange(101.25, 100)", change+"  · "+tone.String(), false, k.Text)
			line("FormatChange(12.40, 12.40)", flat+"  · "+flatTone.String(), false, k.Text)
			line("FormatChange(44.80, 45.62)", down+"  · "+downTone.String(), false, k.Text)
			line("FormatChange(1.001, 1.0)", mustChange(1.001, 1.0), false, k.Text)
			line("BestBid(bids)",
				fmt.Sprintf("%.2f  (%v)  ← 空的一侧是 false", bid, hasBid), true, k.Text)
			line("BestAsk(asks)",
				fmt.Sprintf("%.2f  (%v)", ask, hasAsk), true, k.Text)
			line("BestBid(nil)",
				fmt.Sprintf("%.2f  (%v)", 0.0, hasBestBid(nil)), true, k.Text)
			line("Spread(bids, asks)",
				fmt.Sprintf("%.2f  %.4f  (%v)  ← 交叉的盘口是负价差", gap, fraction, ok),
				true, k.Text)
			line("Levels(bids, asks)",
				fmt.Sprintf("%d 行，Total 从最好的一边累加", len(finance.Levels(bids, asks))),
				true, k.Text)
		})
	})
}

func mustChange(now, prev float64) string {
	s, _ := finance.FormatChange(now, prev)
	return s + "  · 1.0 对 1.0 是没有变"
}

func hasBestBid(bids []finance.Price) bool {
	_, ok := finance.BestBid(bids)
	return ok
}
