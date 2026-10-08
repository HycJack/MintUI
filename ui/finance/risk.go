package finance

import (
	"math"
	"strconv"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/chart"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/data"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// ── risk ───────────────────────────────────────────────────────────────────

// RiskMeterOptions configure a RiskMeter.
type RiskMeterOptions struct {
	// Value is the figure being measured, between Min and Max.
	Value float64
	// Min and Max bound it; zero to one hundred unless the caller says
	// otherwise, so a plain percentage is the zero-config case.
	Min, Max float64
	// Label is what the meter is for, and is required: a bar of coloured
	// segments says nothing about what it is measuring.
	Label string
	// Caption is the caller's sentence under it — "2.1× your equity",
	// "within the firm's limit".
	Caption string
	// Bands are the ranges and what each is called, smallest first. Empty
	// takes one band over the whole range.
	Bands []RiskBand
	// Unit is drawn after the value; empty draws none.
	Unit string
	// Height is the meter's own height; zero takes the dial's own.
	Height float32
}

// RiskBand is one range of a risk meter and what it is called.
type RiskBand struct {
	// To is where the band ends, in the meter's own values.
	To float64
	// Name is what it is called.
	Name string
}

// RiskMeter is one figure inside a range, with the range named.
//
// It is [chart.Gauge] rather than a bar, because a risk figure is read against
// a *limit* and a limit is a point on a scale with something either side of it.
// A horizontal bar has no sense of "over" — it fills and stops — and a risk
// meter that stops at full scale is a risk meter that cannot show a breach.
//
// The bands are the caller's because what counts as dangerous is a house rule
// and not a library's; the dial's arc, its needle and the measuring of its
// own label are the chart's.
func RiskMeter(c *ui.Context, opts RiskMeterOptions) *ui.Element {
	if opts.Label == "" {
		panic("finance: RiskMeter needs a Label; a dial with no name says nothing about what " +
			"it is measuring")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	lo, hi := opts.Min, opts.Max
	if hi <= lo {
		lo, hi = 0, 100
	}
	// A value past either end is pinned rather than drawn off the dial. A
	// needle at 140% of the limit and a needle at 40% are different facts and
	// a dial with room for only one of them is not showing a breach.
	value := opts.Value
	if value < lo {
		value = lo
	}
	if value > hi {
		value = hi
	}

	zones := make([]chart.GaugeZone, 0, len(opts.Bands))
	for _, b := range opts.Bands {
		zones = append(zones, chart.GaugeZone{
			To: b.To, Name: b.Name,
		})
	}

	caption := opts.Caption
	if caption == "" && len(opts.Bands) > 0 {
		caption = bandName(opts.Bands, opts.Value)
	}

	return ui.Column(c).FillWidth().AlignItems(ui.Center).Gap(u).
		Label(opts.Label).Children(func() {
		chart.Gauge(c, chart.GaugeOptions{
			Value: value, Min: lo, Max: hi,
			Label: caption, Zones: zones,
		})
		ui.Row(c).AlignItems(ui.Center).Gap(u * 0.5).Children(func() {
			mono(c, FormatPrice(value), theme.CaptionSize, k.TextMuted)
			if opts.Unit != "" {
				ui.Text(c, opts.Unit).TextColor(k.TextFaint).
					FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
			}
		})
	})
}

// bandName is which band a value falls in, which is the sentence a reader
// actually wants from a risk meter: not "62.4" but "inside the limit".
func bandName(bands []RiskBand, v float64) string {
	name := bands[len(bands)-1].Name
	for _, b := range bands {
		if v <= b.To {
			name = b.Name
			break
		}
	}
	return name
}

// ── sentiment ──────────────────────────────────────────────────────────────

// SentimentOptions configure a SentimentGauge.
type SentimentOptions struct {
	// Score is the reading, from -1 (everything negative) to 1 (everything
	// positive). Zero is neutral and is written as zero, with no sign.
	Score float64
	// Label names the gauge, and is required.
	Label string
	// Caption is the caller's sentence — "call volume 2.1× the average".
	Caption string
	// Bands are the ranges of the score and what each is called, from the
	// bottom up.
	Bands []SentimentBand
	// Size is the gauge's box; zero lets the layout give it one.
	Size float32
}

// SentimentBand is one range of a sentiment gauge.
type SentimentBand struct {
	// To is where the band ends on the score.
	To float64
	// Name is what it is called.
	Name string
}

// SentimentGauge is one reading on a scale that runs both ways.
//
// It is [chart.BulletChart] rather than a gauge, for a specific reason: a
// gauge's arc starts at its own minimum, so it cannot show a scale whose zero
// is in the middle. A reading of "slightly positive" on a scale from −1 to 1
// is a needle barely off the end of a gauge, which looks like an extreme. A
// bullet's bar starts at zero and is drawn in both directions from it, which is
// the only way a centred scale reads as a centred scale.
func SentimentGauge(c *ui.Context, opts SentimentOptions) *ui.Element {
	if opts.Label == "" {
		panic("finance: SentimentGauge needs a Label; a bar with no name says nothing about " +
			"what it is measuring")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	score := opts.Score
	if score < -1 {
		score = -1
	}
	if score > 1 {
		score = 1
	}
	caption := opts.Caption
	if caption == "" {
		caption = sentimentWord(score)
	}

	return ui.Column(c).FillWidth().AlignItems(ui.Center).Gap(u).
		Label(opts.Label).Children(func() {
		chart.BulletChart(c, chart.BulletOptions{
			Value: score, Min: -1, Max: 1,
			Label: caption,
			Color: sentimentInk(score, k),
			Bands: sentimentBands(opts.Bands),
		})
		// The two ends of the scale are written out because a bullet whose
		// bar runs left and right is a bar a reader will take for one that
		// runs only right; the two words fix it in the space of three
		// characters.
		ui.Row(c).FillWidth().Justify(ui.SpaceBetween).Children(func() {
			ui.Text(c, "Bearish").TextColor(k.Danger).
				FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
			ui.Text(c, "Neutral").TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
			ui.Text(c, "Bullish").TextColor(k.Success).
				FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
		})
	})
}

// sentimentWord is a score in the word a reader would use for it. The
// boundaries are round numbers because a score of 0.31 is "slightly positive"
// and not "31% positive" — the number is a composite of several readings and
// its precision is not real.
func sentimentWord(score float64) string {
	switch {
	case score >= 0.6:
		return "Strongly bullish"
	case score >= 0.15:
		return "Slightly bullish"
	case score <= -0.6:
		return "Strongly bearish"
	case score <= -0.15:
		return "Slightly bearish"
	}
	return "Neutral"
}

func sentimentInk(score float64, k theme.Tokens) ui.Color {
	switch {
	case score >= 0.15:
		return k.Success
	case score <= -0.15:
		return k.Danger
	}
	return k.TextMuted
}

// sentimentBands is the caller's bands, or a default three-band split of a
// score from −1 to 1 at the same boundaries the words use — so the band
// boundaries and the words cannot say two different things.
func sentimentBands(bands []SentimentBand) []chart.BulletBand {
	if len(bands) > 0 {
		out := make([]chart.BulletBand, 0, len(bands))
		from := -1.0
		for _, b := range bands {
			out = append(out, chart.BulletBand{Min: from, Max: b.To, Name: b.Name})
			from = b.To
		}
		return out
	}
	return []chart.BulletBand{
		{Min: -1, Max: -0.15, Name: "Strongly bearish"},
		{Min: -0.15, Max: 0.15, Name: "Neutral"},
		{Min: 0.15, Max: 1, Name: "Strongly bullish"},
	}
}

// ── spread ─────────────────────────────────────────────────────────────────

// SpreadIndicatorOptions configure a SpreadIndicator.
type SpreadIndicatorOptions struct {
	// Bids and Asks are the two sides, the caller's.
	Bids, Asks []Price
	// Label names the indicator; empty builds one.
	Label string
	// Wide is the spread above which the market is called wide, as a
	// fraction of the midpoint. Zero takes 0.5%, which is where a liquid
	// market's spread stops being noise.
	Wide float64
	// ShowBars draws the two sides' depth behind the figure, which is what
	// makes a wide spread legible: the reader can see whether it is wide
	// because the market is thin or because one side has pulled.
	ShowBars bool
	// Height is the indicator's box; zero takes one row.
	Height float32
}

// SpreadIndicator is what the two sides cost between them.
//
// The figure is written as a fraction *and* as money, because the two answer
// different questions: a fraction says how expensive it is relative to the
// instrument, and the money says what a round turn actually costs. A reader
// deciding whether to cross the spread needs both, and needs them in that
// order — the fraction is the judgement, the money is the consequence.
func SpreadIndicator(c *ui.Context, opts SpreadIndicatorOptions) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	gap, fraction, ok := Spread(opts.Bids, opts.Asks)
	if !ok {
		panic("finance: SpreadIndicator needs both sides of a book; a spread with one end is " +
			"not a spread")
	}
	wide := opts.Wide
	if wide <= 0 {
		wide = 0.005
	}
	// A negative fraction is a crossed book and is reported as negative and
	// in the danger tone. A book with the bid above the ask is a real thing
	// that happens and it means something specific; clamping it to zero would
	// hide the one interesting thing on the screen.
	over := fraction > wide
	name := opts.Label
	if name == "" {
		name = "Spread"
	}
	// The magnitude, with the sign carried by the tone rather than by the
	// figure: a spread is a gap and does not "rise", and a crossed book's
	// minus is already said by its being below the line.
	pct := formatShare(math.Abs(fraction) * 100)

	return ui.Column(c).FillWidth().Gap(u * 0.75).Label(name).Children(func() {
		spreadRow(c, u, func() {
			tone := spreadTone(over || fraction < 0)
			figure(c, stat{Value: pct, Label: "of the midpoint", Mono: true, Tone: tone})
			figure(c, stat{Value: FormatPrice(gap), Label: "per share", Mono: true, Tone: tone})
		})
		if over {
			ui.Text(c, "Wide for this market").TextColor(k.Warning).
				FontSize(core.FontSize(c, theme.CaptionSize))
		}
	})
}

func spreadTone(over bool) core.Severity {
	if over {
		return core.Warning
	}
	return core.Success
}

// ── margin ─────────────────────────────────────────────────────────────────

// MarginOptions configure a MarginIndicator.
type MarginOptions struct {
	// Used is how much of the account's margin is in use and Available what is
	// left, both as shares of the whole. They are the caller's rather than a
	// ratio because "maintenance margin" and "initial margin" are two
	// different numbers in two different places and this component does not
	// know which one a venue is calling margin.
	Used, Available float64
	// Label names the indicator; empty builds one.
	Label string
	// Maintenance is the level at which the venue starts asking for money, as
	// a share of the whole. Zero takes 0.25, which is the conventional
	// maintenance level for a margin account.
	Maintenance float64
	// Excess is what would be left if every position moved against the
	// account by the given amount, as a share. Zero is not shown.
	Excess float64
	// ExcessLabel is what the Excess figure is about — "after a 5% move".
	ExcessLabel string
}

// MarginIndicator is how much of the account's buying power is committed and
// how much is still free.
//
// The maintenance level is drawn as a marked point on the same bar as the used
// amount, and that is the whole component. An account's margin is not "how
// much have I used" — it is "how far am I from being told to post more" — and
// a bar with no marked threshold cannot answer the second question at all.
func MarginIndicator(c *ui.Context, opts MarginOptions) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	maintenance := opts.Maintenance
	if maintenance <= 0 {
		maintenance = 0.25
	}
	used := float64(clamp01(opts.Used))
	free := float64(clamp01(opts.Available))
	if opts.Available == 0 {
		free = 1 - used
	}
	name := opts.Label
	if name == "" {
		name = "Margin"
	}
	// Below the maintenance level is the state a reader must be told about
	// before anything else on the panel, so the tone comes first and the
	// figures are read against it.
	tone := core.Success
	below := used >= maintenance
	if below {
		tone = core.Danger
	} else if used >= maintenance*0.75 {
		tone = core.Warning
	}

	return ui.Column(c).FillWidth().Gap(u).Label(name).Children(func() {
		ui.Box(c).FillWidth().Height(u * 3).Shrink(0).Role(ui.RoleNone).
			Label(name + " " + formatShare(used*100) + " used").
			Draw(func(p *ui.Painter, r ui.Rect) {
				p.Fill(ui.Rect{X: r.X, Y: r.Y, W: r.W, H: r.H}, k.SurfacePressed, r.H/2)
				p.Fill(ui.Rect{X: r.X, Y: r.Y, W: r.W * float32(used), H: r.H},
					toneInk(tone, k), r.H/2)
				// The threshold sits on top of the fill rather than in a gutter:
				// a marker beside the bar would have to be read as "somewhere
				// near here", and on a bar this is the difference between
				// comfortable and a call.
				x := r.X + r.W*float32(maintenance)
				p.Line(x, r.Y-u, x, r.Y+r.H+u, theme.BorderWidth*2, k.Text)
				internal.Dot(p, x, r.Y+r.H/2, u*1.25, k.Text)
			})
		spreadRow(c, u, func() {
			figure(c, stat{
				Value: formatShare(used * 100), Label: "In use",
				Mono: true, Tone: tone,
			})
			figure(c, stat{
				Value: formatShare(free * 100), Label: "Free", Mono: true,
			})
			if opts.Excess > 0 {
				label := opts.ExcessLabel
				if label == "" {
					label = "Excess"
				}
				figure(c, stat{
					Value: formatShare(opts.Excess * 100), Label: label, Mono: true,
					Tone: core.Warning,
				})
			}
		})
		ui.Text(c, "Maintenance at "+formatShare(maintenance*100)).
			TextColor(k.TextFaint).FontSize(core.FontSize(c, theme.CaptionSize))
		if below {
			ui.Text(c, "At or above the maintenance level").TextColor(k.Danger).
				FontSize(core.FontSize(c, theme.CaptionSize)).Bold()
		}
	})
}

// ── leverage ───────────────────────────────────────────────────────────────

// LeverageSliderOptions configure a LeverageSlider.
type LeverageSliderOptions struct {
	// Leverage is the account's own leverage, the caller's float, from 1 to
	// Max. It is a pointer because a slider that kept its own number would be
	// a second source of truth about an account's risk.
	Leverage *float64
	// Max is the most the control will set; zero takes 10, past which the
	// steps stop being a decision a person makes deliberately.
	Max float64
	// BuyingPower is what the account would have at this leverage, drawn
	// under the slider. It is the caller's because it is the exchange's
	// arithmetic and not this library's.
	BuyingPower float64
	// Equity is what it is a multiple of, for the same reason.
	Equity float64
	// Label names the slider for assistive technology, and is required: a
	// rail and a thumb carry no words of their own.
	Label string
	// Warning is the leverage above which the number turns to the warning
	// tone, zero for Max itself.
	Warning float64
}

// LeverageSlider is how many times the account's equity can be committed.
//
// It is [input.Slider] with the two figures it implies drawn underneath,
// because a leverage slider that only shows a number is asking a reader to
// remember what "4×" means. "4× — $160,000 buying power on $40,000 equity" is
// the same fact stated in the units a decision is made in.
func LeverageSlider(c *ui.Context, opts LeverageSliderOptions) *ui.Element {
	if opts.Leverage == nil {
		panic("finance: LeverageSlider needs a Leverage to point at; it keeps no ratio of its " +
			"own")
	}
	if opts.Label == "" {
		panic("finance: LeverageSlider needs a Label; a rail and a thumb say nothing about " +
			"what they are measuring")
	}
	u := core.Density(c).Unit()
	max := opts.Max
	if max <= 1 {
		max = 10
	}
	warn := opts.Warning
	if warn <= 0 {
		warn = max
	}
	*opts.Leverage = minf(maxf(*opts.Leverage, 1), max)

	tone := core.Success
	if *opts.Leverage >= warn {
		tone = core.Warning
	}

	return ui.Column(c).FillWidth().Gap(u * 1.5).Label(opts.Label).Children(func() {
		input.Slider(c, opts.Leverage, input.SliderOptions{
			Min: 1, Max: max, Step: 1, ShowValue: true,
			Format: "%.0f×", Label: opts.Label,
		})
		spreadRow(c, u, func() {
			figure(c, stat{
				Value: strconv.FormatFloat(*opts.Leverage, 'f', 0, 64) + "×",
				Label: "Leverage", Mono: true, Tone: tone,
			})
			if opts.BuyingPower > 0 {
				// The unit is the figure's own caption rather than text
				// beside the number: three wide money figures and their
				// units in one row overflow it, and the ones after print
				// straight through the one before.
				figure(c, stat{
					Value: FormatPrice(opts.BuyingPower),
					Label: "buying power", Mono: true,
				})
			}
			if opts.Equity > 0 {
				figure(c, stat{
					Value: FormatPrice(opts.Equity),
					Label: "equity", Mono: true,
				})
			}
		})
	})
}

// ── the market ─────────────────────────────────────────────────────────────

// Session is which part of the trading day it is.
type Session int

const (
	// SessionClosed is the weekend, or after the close.
	SessionClosed Session = iota
	// SessionPreMarket is before the open.
	SessionPreMarket
	// SessionOpen is the continuous session.
	SessionOpen
	// SessionPostMarket is after the close and before the next day's open.
	SessionPostMarket
)

func (s Session) String() string {
	switch s {
	case SessionPreMarket:
		return "Pre-market"
	case SessionOpen:
		return "Open"
	case SessionPostMarket:
		return "After hours"
	}
	return "Closed"
}

// sessionTone is the session's own tone. Open is the success tone because that
// is the state a reader is waiting for; everything else is quiet, and the
// after-hours state is quiet rather than warning because an asset that trades
// after hours is not a problem.
func sessionTone(s Session) core.Severity {
	if s == SessionOpen {
		return core.Success
	}
	return core.Neutral
}

// MarketStatusOptions configure a MarketStatus.
type MarketStatusOptions struct {
	// Session is which part of the day it is, and the caller's: only it knows
	// the exchange's calendar and its holidays, and this package has no clock
	// it could read one from.
	Session Session
	// Exchange is the market's name, drawn beside the state.
	Exchange string
	// Opens and Closes are when the session changes, as the caller's strings.
	Opens, Closes string
	// Since is how long the current session has been going, or how long until
	// it changes, as the caller's string.
	Since string
	// Label names the indicator; empty builds one.
	Label string
}

// MarketStatus is whether the market is open, and when it changes.
//
// It is one dot and one word, not a panel. A reader glancing at a status line
// wants one bit of information; anything more is something to read, and a
// screen that makes the most-looked-at bit on it the most expensive to read is
// a screen where people stop looking.
func MarketStatus(c *ui.Context, opts MarketStatusOptions) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	name := opts.Label
	if name == "" {
		name = "Market status"
	}
	tone := sessionTone(opts.Session)
	_, fg := tone.Pair(k)

	word := opts.Session.String()
	if opts.Exchange != "" {
		word = opts.Exchange + " · " + word
	}

	return ui.Column(c).FillWidth().Gap(u * 0.5).Label(name).Children(func() {
		ui.Row(c).AlignItems(ui.Center).Gap(u).Children(func() {
			dot := ui.Box(c).Size(u*2, u*2).Radius(u).Shrink(0).Role(ui.RoleNone)
			if opts.Session == SessionClosed {
				// A closed market draws an outline rather than a filled dot:
				// a grey dot beside a green one is a distinction in a shade,
				// and this is the one bit of the screen that must survive
				// being printed in black and white.
				dot.BorderWidth(theme.BorderWidth).BorderColor(k.TextFaint)
			} else {
				dot.Background(fg)
			}
			dot.Label(name)
			ui.Text(c, word).TextColor(fg).
				FontSize(core.FontSize(c, theme.RowSize)).Bold().Grow(1)
			if opts.Since != "" {
				ui.Text(c, opts.Since).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
			}
		})
		when := opts.Opens
		if opts.Session == SessionOpen || opts.Session == SessionPostMarket {
			when = opts.Closes
		}
		if when != "" {
			ui.Text(c, when).TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize))
		}
	})
}

// ── the clock ──────────────────────────────────────────────────────────────

// TradingSessionClockOptions configure a TradingSessionClock.
type TradingSessionClockOptions struct {
	// Now is the caller's clock, and is required rather than read from the
	// system: a trading clock that disagreed with the venue's is worse than
	// no clock, and only the caller knows what "now" means for a feed.
	Now time.Time
	// Opens and Closes are the session's times on Now's date, as the caller's
	// clock says them.
	Opens, Closes time.Time
	// Breaks are the halts inside the session — a lunch break, a halt after a
	// circuit breaker — which the clock draws as gaps.
	Breaks [][2]time.Time
	// Session says which part of the day it is.
	Session Session
	// Zone names the clock's zone, drawn beside the time.
	Zone string
	// Label names the clock for assistive technology, and is required.
	Label string
	// Height is the clock's height; zero takes one density unit.
	Height float32
}

// TradingSessionClock is the day as a bar: how much of it has been traded, how
// much is left, and where the breaks in it are.
//
// It is drawn rather than assembled from elements because it is a bar of time
// against a bar of time — one rectangle and a few marks — and a layout of
// forty boxes would be forty elements laid out on every frame of a clock that
// is, by construction, always running.
func TradingSessionClock(c *ui.Context, opts TradingSessionClockOptions) *ui.Element {
	if opts.Now.IsZero() {
		panic("finance: TradingSessionClock needs Now; a clock reading the system's own time " +
			"would disagree with the venue's")
	}
	if opts.Label == "" {
		panic("finance: TradingSessionClock needs a Label; a bar of time says nothing about " +
			"what it is timing")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	span := opts.Closes.Sub(opts.Opens)
	elapsed := opts.Now.Sub(opts.Opens)
	share := 0.0
	if span > 0 {
		share = elapsed.Seconds() / span.Seconds()
	}
	share = float64(clamp01(share))
	height := opts.Height
	if height <= 0 {
		height = u * 2.5
	}
	time := opts.Now.Format("15:04:05")

	return ui.Column(c).FillWidth().Gap(u * 0.75).Label(opts.Label).Children(func() {
		ui.Row(c).AlignItems(ui.Center).Gap(u).Children(func() {
			mono(c, time, theme.StatSize, k.Text)
			if opts.Zone != "" {
				ui.Text(c, opts.Zone).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
			}
			ui.Box(c).Grow(1)
			ui.Text(c, opts.Session.String()).TextColor(toneInk(sessionTone(opts.Session), k)).
				FontSize(core.FontSize(c, theme.CaptionSize)).Bold()
		})
		ui.Box(c).FillWidth().Height(height).Shrink(0).Role(ui.RoleNone).
			Label(opts.Label + ", " + formatShare(share*100) + " through the session").
			Draw(func(p *ui.Painter, r ui.Rect) {
				p.Fill(ui.Rect{X: r.X, Y: r.Y, W: r.W, H: r.H}, k.SurfacePressed, r.H/2)
				p.Fill(ui.Rect{X: r.X, Y: r.Y, W: r.W * float32(share), H: r.H},
					toneInk(sessionTone(opts.Session), k), r.H/2)
				if span <= 0 {
					return
				}
				for _, b := range opts.Breaks {
					if b[1].Before(b[0]) {
						continue
					}
					lo := float32(b[0].Sub(opts.Opens).Seconds() / span.Seconds())
					hi := float32(b[1].Sub(opts.Opens).Seconds() / span.Seconds())
					if hi <= lo {
						continue
					}
					p.Fill(ui.Rect{X: r.X + r.W*lo, Y: r.Y, W: r.W * (hi - lo), H: r.H},
						k.Background, 0)
				}
				// The open and the close are the two ends of the bar, and they
				// are drawn as marks because a reader asking "how long is left"
				// is measuring from the right-hand one.
				p.Line(r.X+r.W, r.Y-u*0.5, r.X+r.W, r.Y+r.H+u*0.5, theme.BorderWidth*2, k.Text)
				p.Line(r.X, r.Y-u*0.5, r.X, r.Y+r.H+u*0.5, theme.BorderWidth*2, k.Text)
				x := r.X + r.W*float32(share)
				p.Line(x, r.Y-u, x, r.Y+r.H+u, theme.BorderWidth*2, k.Text)
			})
		spreadRow(c, u, func() {
			ui.Text(c, opts.Opens.Format("15:04")).TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
			ui.Box(c).Grow(1)
			ui.Text(c, opts.Closes.Format("15:04")).TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
		})
	})
}

// ── the tape ticker ────────────────────────────────────────────────────────

// TickerTapeOptions configure a TickerTape.
type TickerTapeOptions struct {
	// Quotes are the caller's instruments and prices, in the order they
	// should run.
	Quotes []WatchQuote
	// Speed is how far across the tape a second goes, in DIPs; zero takes the
	// library's own. It is a speed rather than a duration because the tape's
	// width is the caller's and a duration would make the speed depend on it.
	Speed float32
	// Height is the tape's height; zero takes nine density units.
	Height float32
	// Label names the tape, and is required: a row of figures scrolling past
	// is unreadable to a screen reader and unidentifiable in a test.
	Label string
	// Repeat runs the tape twice, which is what a window narrower than the
	// tape needs.
	Repeat bool
}

// TickerTape is a row of instruments and prices, scrolling sideways.
//
// It is drawn rather than assembled out of text elements because it moves on
// every frame of its life: a row of thirty text elements laid out thirty times
// a second is thirty layouts a second to produce a picture that is one painter
// call. The text is measured with [chart.LabelWidth] — the same measurement the
// chart package measures its axis labels with — so the tape and the charts on
// the same screen agree about how wide a run of figures is.
//
// Under reduced motion it draws still. A ticker that will not stop scrolling is
// exactly the thing a reader who has asked for less motion cannot use, and
// there is no gentler version of it that is still a ticker.
func TickerTape(c *ui.Context, opts TickerTapeOptions) *ui.Element {
	if opts.Label == "" {
		panic("finance: TickerTape needs a Label; a row of figures scrolling past is " +
			"unidentifiable without one")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	speed := opts.Speed
	if speed <= 0 {
		speed = u * 24
	}
	height := opts.Height
	if height <= 0 {
		height = u * 9
	}

	gap := u * 3
	items := make([]tapeItem, len(opts.Quotes))
	total := 0.0
	for i, q := range opts.Quotes {
		text, sev := FormatChange(q.Now, q.Previous)
		it := tapeItem{
			name:  q.Symbol.Ticker,
			price: FormatPrice(q.Now),
			move:  text,
			ink:   toneInk(sev, k),
		}
		it.nameW = chart.LabelWidth(c, it.name, theme.CaptionSize)
		it.priceW = it.nameW + u + chart.LabelWidth(c, it.price, theme.CaptionSize)
		it.width = it.priceW + u*2 + chart.LabelWidth(c, it.move, theme.CaptionSize)
		items[i] = it
		total += float64(it.width) + float64(gap)
	}
	// The tape's length in DIPs, as the float the painter's arithmetic uses —
	// every rectangle it is about to build is float32, and a wrap computed in
	// float64 and converted would leave the tape a fraction of a pixel out of
	// its own loop every time it wrapped.
	span := float32(total)
	if span <= 0 {
		span = 1
	}

	e := ui.Box(c).FillWidth().Height(height).Shrink(0).Role(ui.RoleNone).
		Label(opts.Label)
	if opts.Repeat {
		e.Clip()
	}
	return e.Draw(func(p *ui.Painter, r ui.Rect) {
		if r.W <= 0 || len(items) == 0 {
			return
		}
		// The offset comes from the painter's own clock, which is the only
		// clock a painter has that moves between frames. Reading it at build
		// time instead would mean the tape only moved when the window was
		// rebuilt, which for a ticker is never.
		offset := float32(0)
		if !core.Reduced(c) {
			seconds := float32(p.Now().UnixNano()) / 1e9
			shift := seconds * speed
			// Wrapped, not clamped: the tape runs off the left and comes back
			// on the right, which is what makes the loop invisible.
			offset = -(shift - float32(int(shift/span))*span)
		}
		x := r.X + offset
		// Painted twice when the tape repeats, so that a window narrower than
		// the tape is never half empty.
		passes := 1
		if opts.Repeat {
			passes = 2
		}
		for pass := 0; pass < passes; pass++ {
			for _, it := range items {
				if x > r.X+r.W {
					break
				}
				if x+it.width >= r.X {
					y := r.Y + r.H/2
					// Measured once when the item was built, so the three
					// runs land where the widths said they would. Painter.Text
					// draws and returns nothing: the widths are the
					// measurement, and they are not re-measured here.
					p.Text(x, y, it.name, theme.CaptionSize, k.Text)
					p.Text(x+it.nameW+u, y, it.price, theme.CaptionSize, k.Text)
					p.Text(x+it.priceW+u*2, y, it.move, theme.CaptionSize, it.ink)
				}
				x += it.width + gap
			}
		}
	})
}

// tapeItem is one instrument on the tape, measured once when it is built
// rather than measured on every paint — thirty text measurements a frame for a
// strip whose contents did not change is thirty wasted layouts a second.
type tapeItem struct {
	name, price, move string
	ink               ui.Color
	// nameW is where the price starts and priceW where the move does, so the
	// three runs are placed from measurements taken once rather than from
	// anything a painter can report back.
	nameW, priceW, width float32
}

// ── the overview ───────────────────────────────────────────────────────────

// MarketOverviewOptions configure a MarketOverview.
type MarketOverviewOptions struct {
	// Instruments are the caller's, and one row each. They are the same
	// records a watchlist row is, so that the overview and the watchlist can
	// be fed the same list and cannot drift apart.
	Instruments []Instrument
	// Quotes are the prices for those instruments, matched by ticker. A
	// ticker with no quote draws its empty state rather than a zero, which
	// would say the instrument is worth nothing.
	Quotes []WatchQuote
	// Rows and Cols are the grid's shape; zero for three columns and as many
	// rows as are needed.
	Rows, Cols int
	// Selected is the instrument chosen, as an index into Instruments; -1 for
	// none.
	Selected *int
	// Label names the grid, and is required.
	Label string
	// Query hides the instruments that do not match, and is the caller's
	// because the search field beside it is.
	Query *string
}

// MarketOverview is a grid of instruments: the whole market at a glance, one
// card each.
//
// It is a grid rather than a table because each cell carries four things —
// the ticker, the price, the change and a sparkline — and a table's whole
// contract is that its cells carry one thing each in a column. A table of
// those columns would be four columns where a reader wants one card.
func MarketOverview(c *ui.Context, opts MarketOverviewOptions) *ui.Element {
	if opts.Label == "" {
		panic("finance: MarketOverview needs a Label; a grid of instruments with no name is a " +
			"wall of numbers")
	}
	if opts.Selected == nil {
		panic("finance: MarketOverview needs a Selected instrument to point at; it owns no " +
			"list of its own")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	cols := opts.Cols
	if cols <= 0 {
		cols = 3
	}

	shown := make([]int, 0, len(opts.Instruments))
	for i, in := range opts.Instruments {
		if opts.Query == nil || *opts.Query == "" ||
			containsFold(in.Symbol.Ticker, lowerFold(*opts.Query)) ||
			containsFold(in.Symbol.Name, lowerFold(*opts.Query)) {
			shown = append(shown, i)
		}
	}
	rows := (len(shown) + cols - 1) / cols

	return ui.Column(c).FillWidth().Gap(u * 1.5).Label(opts.Label).
		Role(ui.RoleList).Children(func() {
		if len(shown) == 0 {
			ui.Text(c, "No instruments here").TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.BodySize))
			return
		}
		for r := 0; r < rows; r++ {
			ui.Row(c).FillWidth().Gap(u * 1.5).Children(func() {
				for col := 0; col < cols; col++ {
					i := r*cols + col
					if i >= len(shown) {
						ui.Box(c).Grow(1).Shrink(0).Role(ui.RoleNone)
						continue
					}
					in := opts.Instruments[shown[i]]
					quote, has := quoteFor(opts.Quotes, in.Symbol.Ticker)
					marketCard(c, in, quote, has, *opts.Selected == shown[i])
				}
			})
		}
	})
}

// marketCard is one instrument's cell.
func marketCard(c *ui.Context, in Instrument, q WatchQuote, has, selected bool) *ui.Element {
	u := core.Density(c).Unit()
	if !has {
		return layout.Container(c, layout.ContainerOptions{
			Surface: true, Border: true, Radius: theme.CardRadius, Pad: u * 1.5,
		}, func() {
			SymbolBadge(c, SymbolBadgeOptions{Symbol: in.Symbol})
			ui.Text(c, "No quote").TextColor(core.Tokens(c).TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize))
		})
	}
	return QuoteCard(c, QuoteOptions{
		Symbol: in.Symbol, Now: q.Now, Previous: q.Previous,
		DayVolume: q.Volume, Selected: selected,
	})
}

// quoteFor is a ticker in the prices list, and false when there is none.
func quoteFor(quotes []WatchQuote, ticker string) (WatchQuote, bool) {
	for _, q := range quotes {
		if q.Symbol.Ticker == ticker {
			return q, true
		}
	}
	return WatchQuote{}, false
}

// ── level two ──────────────────────────────────────────────────────────────

// Level2QuotesOptions configure a Level2Quotes.
type Level2QuotesOptions struct {
	// Bids and Asks are the two sides of the book, the caller's.
	Bids, Asks []Price
	// Rows is how many levels of each side to show, zero for eight.
	Rows int
	// Height is the panel's height, and is required: a book with no height
	// grows to fit every level.
	Height float32
	// State and Scroll keep the list's place between frames.
	State  *ui.ListState
	Scroll *ui.ScrollState
	// Highlight marks levels the caller has drawn against — its own resting
	// orders — which is the whole point of level two over level one.
	Highlight []float64
	// Selected is the level the keys move from; -1 for none.
	Selected *int
	// Label names the panel; empty builds one from the sides.
	Label string
}

// Level2Quotes is the whole book, both sides, in one list.
//
// It is one list rather than the two of an [OrderBook] because level two is
// read by *ownership*: a reader is looking for their own price in the book and
// for what is on either side of it, and two columns puts those two things as
// far apart as the window is wide. One list puts them next to each other.
func Level2Quotes(c *ui.Context, opts Level2QuotesOptions) *ui.Element {
	if opts.Height <= 0 {
		panic("finance: Level2Quotes needs a Height; a book with no height grows to fit every " +
			"level rather than scrolling")
	}
	if opts.Selected == nil {
		panic("finance: Level2Quotes needs a Selected level to point at; it owns no book of " +
			"its own")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	rows := opts.Rows
	if rows <= 0 {
		rows = 8
	}
	name := opts.Label
	if name == "" {
		name = "Level two"
	}

	levels := Levels(opts.Bids, opts.Asks)
	// Best first overall: every bid above the best ask, then the asks below
	// them, so the top of the list is the price the market will trade at.
	ordered := orderBookForLevel2(levels)
	depth := Depth{}
	for _, l := range ordered {
		if l.Side == SideBid {
			depth.Bids = append(depth.Bids, l)
		} else {
			depth.Asks = append(depth.Asks, l)
		}
	}
	maxSize := depth.MaxSize()
	if len(ordered) > rows*2 {
		ordered = ordered[:rows*2]
	}

	return ui.Box(c).FillWidth().Height(opts.Height).Label(name).Role(ui.RoleTable).
		Children(func() {
			data.List(c, data.ListOptions{
				Rows: len(ordered), Height: opts.Height, Selected: opts.Selected,
				State: opts.State, Scroll: opts.Scroll,
				Key:   func(i int) any { return ordered[i].Price },
				Label: func(i int) string { return levelLabel(ordered[i]) },
			}, func(i int) {
				l := ordered[i]
				ink := k.Success
				if l.Side == SideAsk {
					ink = k.Danger
				}
				share := 0.0
				if maxSize > 0 {
					share = l.Size / maxSize
				}
				line := ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u*1.25).
					Padding(u*0.5, u*1.5).Radius(theme.SmallRadius).
					Label(levelLabel(l)).Role(ui.RoleNone).
					Draw(func(p *ui.Painter, r ui.Rect) {
						if share > 0 {
							p.Fill(ui.Rect{X: r.X, Y: r.Y, W: r.W * float32(share), H: r.H},
								ink.Alpha(0.12), 0)
						}
					})
				if isHighlighted(opts.Highlight, l.Price) {
					line.Background(k.AccentBg)
				}
				if *opts.Selected == i {
					line.Background(k.SurfaceHover)
				}
				line.Children(func() {
					mono(c, FormatPrice(l.Price), theme.RowSize, ink)
					ui.Box(c).Grow(1)
					mono(c, FormatVolume(int(l.Size+0.5)), theme.RowSize, k.TextMuted)
					mono(c, FormatVolume(int(l.Total+0.5)), theme.RowSize, k.TextFaint)
				})
			})
		})
}

// isHighlighted is whether a price is one the caller marked, compared on the
// price and not on an index: a level's place in the list moves as the book
// moves and its price does not.
func isHighlighted(marks []float64, price float64) bool {
	for _, m := range marks {
		if m == price {
			return true
		}
	}
	return false
}

// orderBookForLevel2 is the whole book in the order a reader wants it: the
// best bids descending, then the best asks ascending. A level-two view sorted
// any other way puts the two halves of the market apart.
func orderBookForLevel2(levels []Level) []Level {
	out := make([]Level, 0, len(levels))
	for _, l := range levels {
		if l.Side == SideBid {
			out = append(out, l)
		}
	}
	for _, l := range levels {
		if l.Side == SideAsk {
			out = append(out, l)
		}
	}
	return out
}

// ── the quote list ─────────────────────────────────────────────────────────

// QuoteListOptions configure a QuoteList.
type QuoteListOptions struct {
	// Quotes are the caller's, in the order they should be shown.
	Quotes []WatchQuote
	// Selected is the row the keys move from; -1 for none.
	Selected *int
	// Sort is the column the rows are ordered by, in the caller's state.
	Sort *data.Sort
	// Height is the table's height, and is required.
	Height float32
	// State and Scroll keep the table's place between frames.
	State  *ui.ListState
	Scroll *ui.ScrollState
	// Label names the table; empty builds one.
	Label string
	// Empty draws instead of the rows when there are none.
	Empty func()
}

// QuoteList is a table of instruments with their prices and their moves.
//
// It is the table form of a [Watchlist]: the same records, with the columns
// lined up rather than pinned to the edges. A watchlist is scanned and a quote
// list is read across, and the two want different shapes out of the same data.
func QuoteList(c *ui.Context, opts QuoteListOptions) *ui.Element {
	if opts.Height <= 0 {
		panic("finance: QuoteList needs a Height; a table with no height grows to fit every " +
			"row rather than scrolling")
	}
	k := core.Tokens(c)
	name := opts.Label
	if name == "" {
		name = "Quotes"
	}
	empty := opts.Empty
	if empty == nil {
		empty = func() {
			ui.Text(c, "No quotes").TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.BodySize))
		}
	}

	cols := []tableColumn{
		{Title: "Instrument", ID: "instrument", Share: 1},
		{Title: "Price", ID: "price", Width: 112, Align: ui.End, Sortable: true},
		{Title: "Change", ID: "change", Width: 112, Align: ui.End, Sortable: true},
		{Title: "Volume", ID: "volume", Width: 100, Align: ui.End, Sortable: true},
		{Title: "When", ID: "when", Width: 104, Align: ui.End},
	}

	return tableOf(c, tableOptions{
		columns: cols,
		rows:    len(opts.Quotes),
		name:    name,
		cell: func(row, col int) {
			q := opts.Quotes[row]
			switch cols[col].ID {
			case "instrument":
				spreadRow(c, core.Density(c).Unit(), func() {
					SymbolBadge(c, SymbolBadgeOptions{
						Symbol: q.Symbol, ShowExchange: true,
					})
				})
			case "price":
				mono(c, FormatPrice(q.Now), theme.RowSize, k.Text)
			case "change":
				text, sev := FormatChange(q.Now, q.Previous)
				mono(c, text, theme.RowSize, toneInk(sev, k))
			case "volume":
				mono(c, FormatVolume(q.Volume), theme.RowSize, k.TextMuted)
			case "when":
				ui.Text(c, q.Symbol.Exchange).TextColor(k.TextFaint).
					FontSize(core.FontSize(c, theme.RowSize))
			}
		},
		key:   func(row int) any { return opts.Quotes[row].Symbol.Ticker },
		row:   func(row int) string { return opts.Quotes[row].Symbol.Label() },
		state: opts.State, scroll: opts.Scroll, height: opts.Height, empty: empty,
	})
}
