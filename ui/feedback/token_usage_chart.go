package feedback

import (
	"fmt"
	"strconv"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// TokenUsage is one turn of a conversation, in tokens.
type TokenUsage struct {
	// Turn is which turn this is, as the caller numbers it. One-based is
	// what a person counts, but the number is drawn as given: the library
	// does not renumber a log.
	Turn int
	// Prompt is how many tokens went in: the messages and the context.
	Prompt float64
	// Completion is how many came back: the reply.
	Completion float64
}

// TokenUsageChartOptions configure a TokenUsageChart.
type TokenUsageChartOptions struct {
	// Usage are the turns, in the order they happened, which is the order
	// they are drawn. A chart that reorders its bars is a chart that
	// disagrees with the transcript next to it.
	Usage []TokenUsage
	// Width bounds the chart; zero lets it fill its parent.
	Width float32
	// Height is the bar area, in DIPs: the room the tallest bar takes. Zero
	// takes the standard one. The bars are scaled to it, not to each other.
	Height float32
}

// TokenUsageChart is a small chart of how many tokens each turn of a
// conversation took, one stacked bar per turn: the prompt below, the
// completion above it, and the total of the whole conversation over the
// top.
//
// It is pure drawing and has no axes, for the reason a meter has ticks
// rather than a scale: the number that matters is in the caption, and the
// bars are there to show which turn was the big one, which a height
// against its neighbours says without a gridline. The two segments take
// two of the window's own colours — the accent and the warning — rather
// than a data palette of this file's own: a chart that invents colours is a
// chart that has to be taught the window's appearance twice.
func TokenUsageChart(c *ui.Context, opts TokenUsageChartOptions) *ui.Element {
	if len(opts.Usage) == 0 {
		panic("feedback: TokenUsageChart needs at least one sample; a chart " +
			"of zero turns claims a conversation that never happened")
	}
	for _, s := range opts.Usage {
		if s.Prompt < 0 || s.Completion < 0 {
			panic("feedback: a token count cannot be negative; a count below " +
				"zero is an input mistake, not a usage")
		}
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	barH := opts.Height
	if barH <= 0 {
		barH = u * 14
	}
	max, total := 0.0, 0.0
	for _, s := range opts.Usage {
		if sum := s.Prompt + s.Completion; sum > max {
			max = sum
		}
		total += s.Prompt + s.Completion
	}

	promptLabel := core.Msg(c, "feedback.tokenUsageChart.prompt", "Prompt")
	completionLabel := core.Msg(c, "feedback.tokenUsageChart.completion",
		"Completion")

	col := ui.Column(c).FillWidth().Gap(u * 2)
	if opts.Width > 0 {
		col.Width(opts.Width)
	}
	col.Children(func() {
		// The caption says the total of the whole conversation, with the
		// legend on the far side of it: the total is the number, and the
		// legend exists to be found when a bar's two colours are in
		// question, not to be read first.
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 2).Children(func() {
			ui.Text(c, fmt.Sprintf(
				core.Msg(c, "feedback.tokenUsageChart.tokens", "%s tokens"),
				groupDigits(int64(total)))).
				TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
			ui.Box(c).Grow(1)
			legendSwatch(c, k.Accent, promptLabel)
			legendSwatch(c, k.Warning, completionLabel)
		})

		ui.Row(c).FillWidth().Justify(ui.Center).Gap(u * 3).Children(func() {
			for _, s := range opts.Usage {
				usageBar(c, s, barH, max)
			}
		})
	})
	return col
}

// usageBar draws one turn: a track the bar area is tall, the prompt filling
// its share from the bottom in the accent, the completion stacked on top in
// the warning, and the turn's number under it.
func usageBar(c *ui.Context, s TokenUsage, barH float32, max float64) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	bw := u * 6

	ph, ch := 0.0, 0.0
	if max > 0 {
		ph = float64(barH) * s.Prompt / max
		ch = float64(barH) * s.Completion / max
	}

	ui.Column(c).Width(bw).Gap(u).Children(func() {
		ui.Column(c).Width(bw).Height(barH).Radius(2).Clip().
			Background(k.Surface).Children(func() {
			// The empty top of the track is what the turn did not use: a
			// bar drawn against nothing would make every turn read as its
			// own full height.
			ui.Box(c).Grow(1)
			if ch > 0 {
				ui.Box(c).FillWidth().Height(float32(ch)).
					Background(k.Warning)
			}
			if ph > 0 {
				ui.Box(c).FillWidth().Height(float32(ph)).
					Background(k.Accent)
			}
		})
		ui.Text(c, strconv.Itoa(s.Turn)).TextColor(k.TextFaint).
			FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
	})
}

// legendSwatch is one entry of the chart's legend: a square of the colour
// and the word it names.
func legendSwatch(c *ui.Context, color ui.Color, label string) {
	u := core.Density(c).Unit()
	side := u * 2.5
	ui.Row(c).AlignItems(ui.Center).Gap(u).Shrink(0).Children(func() {
		ui.Box(c).Size(side, side).Radius(1).Background(color).Shrink(0)
		ui.Text(c, label).TextColor(core.Tokens(c).TextMuted).
			FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
	})
}

// groupDigits renders a whole number with its thousands grouped, the way a
// token count is said: 12,480, not 12480 and not 12.5k.
func groupDigits(n int64) string {
	neg := n < 0
	if neg {
		n = -n
	}
	s := strconv.FormatInt(n, 10)
	var b []byte
	for i, ch := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b = append(b, ',')
		}
		b = append(b, byte(ch))
	}
	if neg {
		return "-" + string(b)
	}
	return string(b)
}
