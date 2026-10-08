package chat

import (
	"strconv"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/feedback"
	"github.com/HycJack/MintUI/ui/theme"
)

// What a conversation costs and how full its window is.
//
// The arithmetic here is deliberately plain functions over two numbers, with no
// element and no Context: a bill is worth exactly as much as the figures it
// was computed from, and a cost estimate that could only be checked by looking
// at a screenshot would be one nobody would check.

// PerMillion is what a rate is per. A model's price is published per million
// tokens and in dollars, and every rate this package is handed is one of those,
// so the conversion happens in one place instead of at each call site — which
// is where a factor of a thousand goes to be lost.
const PerMillion = 1_000_000.0

// Estimate is what a conversation's tokens cost, in dollars.
//
// inRate and outRate are dollars per million tokens, as published. The two
// sides are counted separately because they are priced separately and the gap
// between them is the whole reason a long conversation is cheap: a million
// tokens of context at one rate and a thousand tokens of answer at another
// are two line items, and collapsing them into one number hides the answer.
func Estimate(inTokens, outTokens int, inRate, outRate float64) float64 {
	return float64(inTokens)*inRate/PerMillion + float64(outTokens)*outRate/PerMillion
}

// EstimateTotal is what one token count costs when the caller cannot say which
// side of the conversation it was on — which is the live counter's problem,
// because a token is counted before the reply that will spend it has arrived.
//
// It charges the higher of the two rates rather than guessing a split. A
// caller that knows should call Estimate; a caller that does not is owed a
// figure that is never too small, because a counter that under-reports is a
// bill that arrives as a surprise.
func EstimateTotal(tokens int, inRate, outRate float64) float64 {
	if tokens <= 0 {
		return 0
	}
	return float64(tokens) * max(inRate, outRate) / PerMillion
}

// FormatCost is a cost as the dollars and cents a reader recognises, with four
// decimal places.
//
// Four and not two because the numbers that matter in a chat are small: a long
// conversation costs a few cents, and at two places every one of them reads
// "$0.00". It drops to two once the figure is large enough for them to say
// something, so a caller's own bill does not come out as "$1234.5600".
func FormatCost(cost float64) string {
	if cost <= 0 {
		return "$0.00"
	}
	if cost < 0.01 {
		return "$" + strconv.FormatFloat(cost, 'f', 4, 64)
	}
	if cost < 1000 {
		return "$" + strconv.FormatFloat(cost, 'f', 2, 64)
	}
	return "$" + internal.Commas(int(cost))
}

// FormatTokens is a token count as something short: the number itself up to
// ten thousand, then a rounded figure in thousands.
//
// The switch is at ten thousand rather than a thousand because "12.3k" is
// harder to read than "12,345" and a conversation reaches four figures early;
// the count is a running total, not a headline, and only becomes a headline
// once it is too long for the strip it sits in.
func FormatTokens(n int) string {
	switch {
	case n < 0:
		return "0"
	case n < 10_000:
		return internal.Commas(n)
	case n < 1_000_000:
		return strconv.FormatFloat(float64(n)/1000, 'f', 1, 64) + "k"
	}
	return strconv.FormatFloat(float64(n)/1_000_000, 'f', 1, 64) + "M"
}

// TokenCounterOptions configure a TokenCounter.
type TokenCounterOptions struct {
	// Tokens is the count so far. It is the caller's number, read again each
	// frame, so a counter in a header and one in a composer cannot disagree.
	Tokens int
	// Cost is what those tokens came to. Zero draws no cost: a counter that
	// has no price to show would otherwise print "$0.00" and read as free.
	Cost float64
	// Label names the counter; empty takes the library's "Tokens".
	Label string
	// ShowCost is implied by a non-zero Cost; it is here so a caller can say
	// the price is genuinely zero rather than unknown.
	ShowCost bool
}

// TokenCounter is "1,284 tokens · $0.0031", the line a conversation reports
// itself with.
//
// It is two pieces of text in a row rather than a bar, because the reader who
// wants this number wants the number. A bar would answer a question nobody
// asked — how much of something is left — and would have to be read against
// something to mean anything.
func TokenCounter(c *ui.Context, opts TokenCounterOptions) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()

	name := opts.Label
	if name == "" {
		name = core.Msg(c, "chat.tokens", core.Def("Tokens"))
	}

	e := ui.Row(c).AlignItems(ui.Center).Gap(u * 0.75).Shrink(0).Label(name)
	e.Children(func() {
		ui.Text(c, FormatTokens(opts.Tokens)).TextColor(k.TextMuted).
			FontSize(core.FontSize(c, theme.CaptionSize))
		ui.Text(c, name).TextColor(k.TextFaint).
			FontSize(core.FontSize(c, theme.CaptionSize))
		if opts.Cost > 0 || opts.ShowCost {
			ui.Text(c, "·").TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize))
			ui.Text(c, FormatCost(opts.Cost)).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize))
		}
	})
	return e
}

// CostEstimatorOptions configure a CostEstimator.
type CostEstimatorOptions struct {
	// InTokens and OutTokens are what has been sent and what has come back.
	InTokens, OutTokens int
	// InRate and OutRate are dollars per million, as published.
	InRate, OutRate float64
	// Model is the name of the model being priced. It is required when the
	// rates are not both given: a cost with no model behind it is a number
	// nobody can check against a bill.
	Model string
	// Breakdown asks for the two sides separately rather than one figure.
	Breakdown bool
	// Label names the estimate; empty takes the model's name.
	Label string
}

// CostEstimator is what a conversation has cost, and what it will cost to keep
// going at this rate.
//
// The two figures are kept apart on purpose. A single total answers "has this
// got expensive" and nothing else, and the question a reader actually has
// halfway through a long session is "is the answer I am reading costing me,
// or is it everything I have already sent" — which is the difference between
// the two rates, and can be an order of magnitude.
func CostEstimator(c *ui.Context, opts CostEstimatorOptions) *ui.Element {
	if opts.InRate < 0 || opts.OutRate < 0 {
		panic("chat: CostEstimator was given a negative rate; a price is not a debt")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	total := Estimate(opts.InTokens, opts.OutTokens, opts.InRate, opts.OutRate)
	name := opts.Label
	if name == "" {
		name = opts.Model
	}
	if name == "" {
		name = core.Msg(c, "chat.cost", core.Def("Cost"))
	}

	e := ui.Column(c).AlignItems(ui.Start).Gap(u * 0.25).Label(name)
	e.Children(func() {
		row := func(label, value string, strong bool) {
			ui.Row(c).AlignItems(ui.Center).Gap(u * 1.5).Children(func() {
				ui.Text(c, label).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize))
				t := ui.Text(c, value).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize))
				if strong {
					t.TextColor(k.Text)
				}
			})
		}
		if opts.Breakdown {
			row("In", FormatCost(Estimate(opts.InTokens, 0, opts.InRate, opts.OutRate)), false)
			row("Out", FormatCost(Estimate(0, opts.OutTokens, opts.InRate, opts.OutRate)), false)
		}
		row("Total", FormatCost(total), true)
		if opts.Model != "" {
			ui.Text(c, opts.Model).TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize))
		}
	})
	return e
}

// ContextWindowMeterOptions configure a ContextWindowMeter.
type ContextWindowMeterOptions struct {
	// Used is how much of the window is taken.
	Used int
	// Window is how much there is. It must be positive: a window of nothing
	// cannot be filled, and dividing by it is the kind of mistake that shows
	// up as a bar at 300%.
	Window int
	// WarnAt and DangerAt are the fractions at which the meter changes tone.
	// They are absolute rather than relative, because "close" means close to
	// the model's limit and the model's limit is a number, not a proportion
	// of whatever this conversation has used so far.
	WarnAt, DangerAt float32
	// Label names the meter; empty takes the library's "Context".
	Label string
}

// ContextWindowMeter is how full the model's window is, and how close that is
// to being a problem.
//
// It is a filled bar and a number rather than a dial, for the reason the
// board's columns are fixed widths: a reader glancing at a transcript wants
// "how much is left", and a filled bar read against its track answers that in
// one pass where a number does not.
func ContextWindowMeter(c *ui.Context, opts ContextWindowMeterOptions) *ui.Element {
	if opts.Window <= 0 {
		panic("chat: ContextWindowMeter needs a positive Window; a window of nothing " +
			"cannot be filled")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	warn, danger := opts.WarnAt, opts.DangerAt
	if warn <= 0 {
		warn = 0.7
	}
	if danger <= 0 {
		danger = 0.9
	}
	fraction := float32(opts.Used) / float32(opts.Window)
	sev := core.Neutral
	switch {
	case fraction >= danger:
		sev = core.Danger
	case fraction >= warn:
		sev = core.Warning
	}

	name := opts.Label
	if name == "" {
		name = core.Msg(c, "chat.context", core.Def("Context"))
	}

	e := ui.Column(c).FillWidth().Gap(u).Label(name)
	e.Children(func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.5).Children(func() {
			ui.Text(c, name).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
			ui.Box(c).Grow(1)
			ui.Text(c, FormatTokens(opts.Used)+" / "+FormatTokens(opts.Window)).
				TextColor(k.TextMuted).FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
		})
		feedback.Progress(c, feedback.ProgressOptions{
			Label:    name,
			Value:    fraction,
			Height:   u * 1.5,
			Severity: sev,
		})
		// The mark is a signal bar rather than a colour: the bar's own tone is
		// the warning, and colour alone would say nothing to a reader who
		// cannot see it.
		if sev != core.Neutral {
			ui.Row(c).AlignItems(ui.Center).Gap(u).Children(func() {
				display.Meter(c, 4, 4, sev)
				ui.Text(c, severityWord(c, sev)).TextColor(mustFg(sev, k)).
					FontSize(core.FontSize(c, theme.CaptionSize))
			})
		}
	})
	return e
}

// severityWord is what the meter says when the window is nearly gone. It takes
// the Context so that the words go through core.Msg like every other word the
// library owns, and can be reworded by a window that wants different ones.
func severityWord(c *ui.Context, sev core.Severity) string {
	if sev == core.Danger {
		return core.Msg(c, "chat.context.full", core.Def("Nearly full"))
	}
	return core.Msg(c, "chat.context.runningOut", core.Def("Running out"))
}

// mustFg is a severity's foreground. The meter's own colours come from
// Severity.Pair, so asking it here rather than picking one is what keeps a
// warning and its word from being two different warnings.
func mustFg(sev core.Severity, k theme.Tokens) ui.Color {
	_, fg := sev.Pair(k)
	return fg
}

// VoiceWaveformOptions configure a VoiceWaveform.
type VoiceWaveformOptions struct {
	// Levels are the bar heights, 0 to 1, one per bar. It is the caller's
	// array: a waveform is a picture of a moment, and the moment is the
	// caller's to hold.
	Levels []float32
	// Bars is how many bars to draw when Levels is empty or short. Zero takes
	// a row wide enough to read as a waveform.
	Bars int
	// Width and Height are the mark's size in DIPs. Zero takes the rest of
	// its row and one density unit tall, which is the height a level meter
	// drawn with strokes has to be to read at all.
	Width, Height float32
	// Color draws the bars; zero alpha takes the accent, the one colour that
	// says "the system is working" in both appearances.
	Color ui.Color
	// Label names the mark for assistive technology; it is required, since a
	// waveform is a picture of a sound and carries no words of its own.
	Label string
}

// VoiceWaveform is the level meter beside a voice input: a row of bars whose
// heights are the levels.
//
// It is drawn rather than assembled out of elements, because a level meter
// redraws every frame during speech and a row of forty boxes would be forty
// elements laid out forty times a second to produce a picture that is one
// painter call. Under reduced motion it draws the levels it was given, at
// half height, rather than a flat line: the mark still shows that sound is
// arriving, and a silent row of equal bars would say the microphone has
// failed.
func VoiceWaveform(c *ui.Context, opts VoiceWaveformOptions) *ui.Element {
	if opts.Label == "" {
		panic("chat: VoiceWaveform needs a Label; a waveform is a picture of a sound " +
			"and carries no words of its own")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	bars := opts.Bars
	if bars <= 0 {
		bars = 24
	}
	levels := opts.Levels
	if len(levels) < bars {
		// A short array is padded rather than stretched: stretching would
		// claim there is more signal than there is, and a padded end reads as
		// the quiet it is.
		padded := make([]float32, bars)
		copy(padded, levels)
		levels = padded
	}
	levels = levels[:bars]

	col := opts.Color
	if col.A == 0 {
		col = k.Accent
	}
	height := opts.Height
	if height <= 0 {
		height = u * 4
	}
	still := core.Reduced(c)
	levels = append([]float32(nil), levels...)

	e := ui.Box(c).FillWidth().Height(height).Shrink(0).Role(ui.RoleStatus).Label(opts.Label)
	if opts.Width > 0 {
		e = e.Width(opts.Width).Shrink(0)
	}
	e.Draw(func(p *ui.Painter, r ui.Rect) {
		// The gap is a share of the pitch, not of the whole meter. As a share
		// of r.W it is 0.18*r.W between every pair of bars, so from seven bars
		// up the gaps alone are wider than the meter and the bar comes out
		// negative — which is how this drew nothing at all at any width, and
		// the default is twenty-four bars.
		pitch := r.W / float32(bars)
		gap := pitch * 0.42
		bar := pitch - gap
		if bar <= 0 {
			return
		}
		for i := range bars {
			// Half height and from the floor up when the window asked for no
			// motion: the mark still has a shape, and a level meter that had
			// flattened to nothing would read as a microphone that has failed
			// rather than as a quieter interface.
			level := levels[i]
			alpha := float32(1)
			if still {
				level = 0.5 * (level + 0.25)
				alpha = 0.55
			}
			h := r.H * clamp01(level)
			if h <= 0 {
				continue
			}
			// The colour is resolved again inside the paint rather than
			// captured, so a meter repainting across a palette change takes
			// the window's colour now rather than the one it was built with.
			p.Fill(ui.Rect{
				X: r.X + float32(i)*(bar+gap), Y: r.Y + r.H - h,
				W: bar, H: h,
			}, col.Alpha(alpha), bar/2)
		}
	})
	return e
}

func clamp01(v float32) float32 {
	switch {
	case v < 0:
		return 0
	case v > 1:
		return 1
	}
	return v
}
