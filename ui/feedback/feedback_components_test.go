package feedback

import (
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// ── Meter ──────────────────────────────────────────────────────────────────

func TestMeterFillsToItsValue(t *testing.T) {
	const label = "Storage"
	tt := render(420, 120, func(c *ui.Context) {
		ui.Box(c).Padding(10).Children(func() {
			Meter(c, MeterOptions{Label: label, Value: 4.2, Max: 10, Width: 200})
		})
	})

	r := box(t, tt, label)
	if r.W != 200 {
		t.Errorf("bar is %v wide, want the 200 it was given", r.W)
	}
	// 42% full: a quarter of the way is fill, three quarters is still the
	// track, and neither sample sits on the value's own tick.
	light := theme.Light()
	wantColor(t, tt, r.X+r.W*0.25, r.Y+r.H/2, light.Accent, "the filled part")
	wantColor(t, tt, r.X+r.W*0.75, r.Y+r.H/2, light.Surface,
		"the part past the value is still the track")
}

func TestMeterShowsItsValueWhenAsked(t *testing.T) {
	tt := render(420, 120, func(c *ui.Context) {
		ui.Box(c).Padding(10).Children(func() {
			Meter(c, MeterOptions{
				Label: "Context", Value: 12, Max: 16, Width: 200, ShowValue: true,
			})
		})
	})
	wantTexts(t, tt, "Context", "12 / 16")
}

func TestMeterMarksItsThreshold(t *testing.T) {
	tt := render(420, 120, func(c *ui.Context) {
		ui.Box(c).Padding(10).Children(func() {
			Meter(c, MeterOptions{
				Label: "Storage", Value: 4.2, Max: 10, Width: 200, Threshold: 7.5,
			})
		})
	})
	r := box(t, tt, "Storage")
	light := theme.Light()
	// The fill has not reached the threshold, so it is still the accent and
	// the threshold shows as a faint tick at its own position on the scale.
	wantColor(t, tt, r.X+r.W*0.25, r.Y+r.H/2, light.Accent,
		"the fill below the threshold is still the accent")
	wantColor(t, tt, r.X+r.W*0.75, r.Y+r.H/2, light.TextFaint,
		"the tick at the threshold's position")
}

func TestMeterEscalatesPastItsThreshold(t *testing.T) {
	tt := render(420, 120, func(c *ui.Context) {
		ui.Box(c).Padding(10).Children(func() {
			Meter(c, MeterOptions{
				Label: "Storage", Value: 9, Max: 10, Width: 200, Threshold: 7.5,
			})
		})
	})
	r := box(t, tt, "Storage")
	wantColor(t, tt, r.X+r.W*0.25, r.Y+r.H/2, theme.Light().Warning,
		"past the threshold the fill is the warning")
}

func TestMeterIsDangerWhenOverFull(t *testing.T) {
	tt := render(420, 120, func(c *ui.Context) {
		ui.Box(c).Padding(10).Children(func() {
			Meter(c, MeterOptions{Label: "Storage", Value: 12, Max: 10, Width: 200})
		})
	})
	r := box(t, tt, "Storage")
	// Over full draws full, in the danger: a store past its capacity is the
	// one reading a meter exists to shout about, and it needs no threshold
	// to be marked.
	wantColor(t, tt, r.X+r.W*0.25, r.Y+r.H/2, theme.Light().Danger,
		"an over-full value is the danger")
}

func TestMeterRejectsWhatItCannotDraw(t *testing.T) {
	wantsPanic(t, "Meter needs a Label", func(c *ui.Context) {
		Meter(c, MeterOptions{Value: 1, Max: 2})
	})
	wantsPanic(t, "Max above 0", func(c *ui.Context) {
		Meter(c, MeterOptions{Label: "Storage", Value: 1, Max: 0})
	})
	wantsPanic(t, "cannot be negative", func(c *ui.Context) {
		Meter(c, MeterOptions{Label: "Storage", Value: -1, Max: 10})
	})
}

func TestMeterInDarkMode(t *testing.T) {
	tt := renderDark(420, 120, func(c *ui.Context) {
		ui.Box(c).Padding(10).Children(func() {
			Meter(c, MeterOptions{Label: "Context", Value: 0.9, Max: 3, Width: 200})
		})
	})
	r := box(t, tt, "Context")
	// The value sits at three tenths, so the samples at either side are
	// clear of its tick: the dark palette's accent on the dark palette's
	// track, not the light pair a component that hard-codes an appearance
	// would draw.
	wantColor(t, tt, r.X+r.W*0.15, r.Y+r.H/2, theme.Dark().Accent,
		"the fill in the dark palette")
	wantColor(t, tt, r.X+r.W*0.6, r.Y+r.H/2, theme.Dark().Surface,
		"the track in the dark palette")
}

// ── Toast / Toaster ────────────────────────────────────────────────────────

func TestToasterDrawsWhatWasShown(t *testing.T) {
	var store ToastStore
	ShowToast(&store, ToastItem{
		Title: "Saved", Body: "Three callbacks archived", Severity: core.Success,
	})
	ShowToast(&store, ToastItem{Title: "Sync failed", Severity: core.Danger})

	tt := render(520, 300, func(c *ui.Context) {
		Toaster(c, &store, ToasterOptions{})
	})
	wantTexts(t, tt, "Saved", "Three callbacks archived", "Sync failed")
}

func TestDismissToastTakesItOffTheList(t *testing.T) {
	var store ToastStore
	id := ShowToast(&store, ToastItem{Title: "Saved", Body: "Three archived"})

	tt := render(320, 200, func(c *ui.Context) {
		Toaster(c, &store, ToasterOptions{})
	})
	if !tt.HasText("Saved") {
		t.Fatalf("the toast shown is not on screen: %q", tt.Texts())
	}

	if !DismissToast(&store, id) {
		t.Error("dismissing the toast's own ID should report it took it")
	}
	if DismissToast(&store, id) {
		t.Error("a second dismiss should report there was nothing to take")
	}

	tt2 := render(320, 200, func(c *ui.Context) {
		Toaster(c, &store, ToasterOptions{})
	})
	if tt2.HasText("Saved") {
		t.Error("the toast is still on screen after its dismiss")
	}
}

func TestToasterReportsTheCloseItWasPressed(t *testing.T) {
	var store ToastStore
	id := ShowToast(&store, ToastItem{Title: "Saved"})

	reported := -1
	tt := render(320, 240, func(c *ui.Context) {
		if d := Toaster(c, &store, ToasterOptions{}).Dismissed(); d >= 0 {
			reported = d
			DismissToast(&store, d)
		}
	})
	// The close is named after the toast it closes, the way an alert's is.
	if err := tt.Click("Dismiss Saved"); err != nil {
		t.Fatal(err)
	}
	if reported != id {
		t.Errorf("the toaster reported %d, want the pressed toast's id %d",
			reported, id)
	}
}

func TestShowToastNeedsATitle(t *testing.T) {
	wantsPanic(t, "a toast needs a Title", func(c *ui.Context) {
		var store ToastStore
		ShowToast(&store, ToastItem{Body: "no title"})
	})
}

func TestToasterInDarkMode(t *testing.T) {
	var store ToastStore
	ShowToast(&store, ToastItem{Title: "Saved", Severity: core.Success})

	tt := renderDark(320, 200, func(c *ui.Context) {
		Toaster(c, &store, ToasterOptions{})
	})
	wantTexts(t, tt, "Saved")

	// The toast is drawn on the fill surface, which is the one place in the
	// interface with no background to pair a severity with: the dark
	// palette's fill, not the light one a component that hard-codes an
	// appearance would draw.
	r := box(t, tt, "Saved")
	wantColor(t, tt, r.X+5, r.Y+r.H/2, theme.Dark().Fill,
		"the toast's surface in the dark palette")
	wantNotColor(t, tt, r.X+5, r.Y+r.H/2, theme.Light().Fill,
		"not the light palette's fill")
}

// ── CostBreakdown ──────────────────────────────────────────────────────────

func TestCostBreakdownListsItsRowsAndTotal(t *testing.T) {
	tt := render(420, 240, func(c *ui.Context) {
		ui.Box(c).Padding(10).Children(func() {
			CostBreakdown(c, CostBreakdownOptions{
				Rows: []CostRow{
					{Name: "Model", Amount: 1.2},
					{Name: "Storage", Amount: 0.3, Muted: true},
				},
				Unit: "$", ShowTotal: true,
			})
		})
	})
	wantTexts(t, tt, "Model", "$1.20", "Storage", "$0.30", "Total", "$1.50")
}

func TestCostBreakdownNeedsRows(t *testing.T) {
	wantsPanic(t, "CostBreakdown needs at least one Row", func(c *ui.Context) {
		CostBreakdown(c, CostBreakdownOptions{Unit: "$"})
	})
	wantsPanic(t, "a cost row needs a Name", func(c *ui.Context) {
		CostBreakdown(c, CostBreakdownOptions{
			Rows: []CostRow{{Amount: 1}}, Unit: "$",
		})
	})
}

func TestCostBreakdownInDarkMode(t *testing.T) {
	tt := renderDark(420, 240, func(c *ui.Context) {
		ui.Box(c).Padding(10).Children(func() {
			CostBreakdown(c, CostBreakdownOptions{
				Rows: []CostRow{{Name: "Model", Amount: 1.2}},
				Unit: "$",
			})
		})
	})
	wantTexts(t, tt, "Model", "$1.20")

	// The room between the name and the figure is the window itself: a row
	// drawn from the light palette would show the light background there.
	r := box(t, tt, "Model")
	wantColor(t, tt, r.X+r.W/2, r.Y+r.H/2, theme.Dark().Background,
		"the window the rows sit in, in the dark palette")
	wantNotColor(t, tt, r.X+r.W/2, r.Y+r.H/2, theme.Light().Background,
		"not the light background")
}

// ── EvalResultTable ────────────────────────────────────────────────────────

func TestEvalResultTableCountsWhatPassed(t *testing.T) {
	rows := []EvalRow{
		{Name: "Smoke", Score: "100%", Pass: true},
		{Name: "Regression", Score: "98%", Pass: true},
		{Name: "Latency", Score: "1.2s", Pass: true},
		{Name: "Memory", Score: "over budget", Pass: false},
		{Name: "Startup", Score: "0.8s", Pass: false},
	}
	tt := render(420, 320, func(c *ui.Context) {
		ui.Box(c).Padding(10).Children(func() {
			EvalResultTable(c, EvalResultTableOptions{Rows: rows})
		})
	})
	wantTexts(t, tt, "3 of 5 passed", "Smoke", "100%",
		"Memory", "over budget", "Startup", "0.8s")
}

func TestEvalResultTableMarksItsRows(t *testing.T) {
	tt := render(420, 200, func(c *ui.Context) {
		ui.Box(c).Padding(10).Children(func() {
			EvalResultTable(c, EvalResultTableOptions{Rows: []EvalRow{
				{Name: "Smoke", Score: "100%", Pass: true},
				{Name: "Memory", Score: "over budget", Pass: false},
			}})
		})
	})
	// The marks are the severities' own inks: a row that passed is the
	// success, a row that failed is the danger, and a test that does not
	// say so would not notice either being replaced by the other.
	passed := box(t, tt, "Smoke passed")
	wantColor(t, tt, passed.X+4, passed.Y+passed.H/2, theme.Light().Success,
		"the mark on the row that passed")
	failed := box(t, tt, "Memory failed")
	wantColor(t, tt, failed.X+4, failed.Y+failed.H/2, theme.Light().Danger,
		"the mark on the row that failed")
}

func TestEvalResultTableNeedsRows(t *testing.T) {
	wantsPanic(t, "EvalResultTable needs at least one EvalRow",
		func(c *ui.Context) {
			EvalResultTable(c, EvalResultTableOptions{})
		})
}

func TestEvalResultTableInDarkMode(t *testing.T) {
	tt := renderDark(420, 200, func(c *ui.Context) {
		ui.Box(c).Padding(10).Children(func() {
			EvalResultTable(c, EvalResultTableOptions{Rows: []EvalRow{
				{Name: "Smoke", Score: "100%", Pass: true},
				{Name: "Memory", Score: "over budget", Pass: false},
			}})
		})
	})
	wantTexts(t, tt, "1 of 2 passed")

	passed := box(t, tt, "Smoke passed")
	wantColor(t, tt, passed.X+4, passed.Y+passed.H/2, theme.Dark().Success,
		"the dark palette's success, not the light one")
}

// ── TokenUsageChart ────────────────────────────────────────────────────────

func TestTokenUsageChartAddsItsTurnsUp(t *testing.T) {
	// The turn numbers are eleven and twelve rather than one and two, so a
	// test that looks them up cannot be satisfied by a digit of the total.
	usage := []TokenUsage{
		{Turn: 11, Prompt: 800, Completion: 400},
		{Turn: 12, Prompt: 500, Completion: 340},
	}
	tt := render(420, 200, func(c *ui.Context) {
		ui.Box(c).Padding(10).Children(func() {
			TokenUsageChart(c, TokenUsageChartOptions{Usage: usage})
		})
	})
	wantTexts(t, tt, "2,040 tokens", "Prompt", "Completion", "11", "12")
}

func TestTokenUsageChartDrawsItsBars(t *testing.T) {
	usage := []TokenUsage{
		{Turn: 11, Prompt: 800, Completion: 400},
		{Turn: 12, Prompt: 500, Completion: 340},
	}
	tt := render(420, 200, func(c *ui.Context) {
		ui.Box(c).Padding(10).Children(func() {
			TokenUsageChart(c, TokenUsageChartOptions{Usage: usage})
		})
	})
	light := theme.Light()

	// Turn 11 is the tall one: its prompt is two thirds of the bar area and
	// its completion the rest, and the turn's number sits under the bar,
	// centred on it. Sampled a little way up from the track's bottom, the
	// column is the accent; higher still, the warning.
	label := box(t, tt, "11")
	x := label.X + label.W/2
	wantColor(t, tt, x, label.Y-10, light.Accent, "the prompt's segment")
	wantColor(t, tt, x, label.Y-45, light.Warning, "the completion's segment")
	wantColor(t, tt, x, label.Y-65, light.Background,
		"above the bar is the window, not ink")
}

func TestTokenUsageChartNeedsSamples(t *testing.T) {
	wantsPanic(t, "TokenUsageChart needs at least one sample",
		func(c *ui.Context) {
			TokenUsageChart(c, TokenUsageChartOptions{})
		})
}

func TestTokenUsageChartInDarkMode(t *testing.T) {
	tt := renderDark(420, 200, func(c *ui.Context) {
		ui.Box(c).Padding(10).Children(func() {
			TokenUsageChart(c, TokenUsageChartOptions{Usage: []TokenUsage{
				{Turn: 11, Prompt: 800, Completion: 400},
				{Turn: 12, Prompt: 500, Completion: 340},
			}})
		})
	})
	wantTexts(t, tt, "2,040 tokens")

	label := box(t, tt, "11")
	wantColor(t, tt, label.X+label.W/2, label.Y-10, theme.Dark().Accent,
		"the dark palette's accent in the prompt's segment")
}

// ── ToolRegistryPanel ──────────────────────────────────────────────────────

func TestToolRegistryPanelCountsItsTools(t *testing.T) {
	tools := []ToolEntry{
		{Name: "search", Description: "Find files in the workspace",
			Enabled: true},
		{Name: "deploy", Description: "Ship a build", Enabled: false},
	}
	tt := render(420, 240, func(c *ui.Context) {
		ui.Box(c).Padding(10).Children(func() {
			ToolRegistryPanel(c, ToolRegistryPanelOptions{Tools: tools})
		})
	})
	wantTexts(t, tt, "Tools", "2 tools", "search",
		"Find files in the workspace", "deploy", "Ship a build")
}

func TestToolRegistryPanelMarksWhoIsIn(t *testing.T) {
	tools := []ToolEntry{
		{Name: "search", Description: "Find files in the workspace",
			Enabled: true},
		{Name: "deploy", Description: "Ship a build", Enabled: false},
	}
	tt := render(420, 240, func(c *ui.Context) {
		ui.Box(c).Padding(10).Children(func() {
			ToolRegistryPanel(c, ToolRegistryPanelOptions{Tools: tools})
		})
	})
	light := theme.Light()

	// The marks are a presence's: in is the one bright dot that does not
	// follow the appearance, out is the hollow ring.
	in := box(t, tt, "search enabled")
	wantColor(t, tt, in.X+6, in.Y+in.H/2, light.Lively,
		"the filled dot on the tool that is in")

	out := box(t, tt, "deploy disabled")
	// A pixel a quarter of the dot in from the left edge is inside the
	// ring and not on it: one-pixel borders antialias, and the test wants
	// the hollow, not the edge.
	wantColor(t, tt, out.X+3, out.Y+out.H/2, light.Background,
		"the hollow ring shows the window through its middle")
}

func TestToolRegistryPanelNeedsTools(t *testing.T) {
	wantsPanic(t, "ToolRegistryPanel needs at least one ToolEntry",
		func(c *ui.Context) {
			ToolRegistryPanel(c, ToolRegistryPanelOptions{})
		})
}

func TestToolRegistryPanelInDarkMode(t *testing.T) {
	tt := renderDark(420, 240, func(c *ui.Context) {
		ui.Box(c).Padding(10).Children(func() {
			ToolRegistryPanel(c, ToolRegistryPanelOptions{Tools: []ToolEntry{
				{Name: "search", Description: "Find files in the workspace",
					Enabled: true},
			}})
		})
	})
	wantTexts(t, tt, "Tools", "1 tools", "search")

	// Lively is the one colour that does not follow the appearance: the
	// dot is the same in both, and the window the rows sit in is the dark
	// one.
	in := box(t, tt, "search enabled")
	wantColor(t, tt, in.X+6, in.Y+in.H/2, theme.Dark().Lively,
		"the presence dot keeps its colour in the dark")
	wantColor(t, tt, in.X+in.W/2, in.Y+in.H/2, theme.Dark().Background,
		"the window the row sits in, in the dark palette")
}

// ── TraceViewer ────────────────────────────────────────────────────────────

func TestTraceViewerListsItsEvents(t *testing.T) {
	events := []TraceEvent{
		{Name: "Fetch tools", At: "14:02:01", Duration: "220ms",
			Severity: core.Neutral},
		{Name: "Model call", At: "14:02:01", Duration: "3.4s",
			Severity: core.Accent},
		{Name: "Tool failed", At: "14:02:05", Duration: "18ms",
			Severity: core.Danger},
	}
	tt := render(520, 280, func(c *ui.Context) {
		ui.Box(c).Padding(10).Children(func() {
			TraceViewer(c, TraceViewerOptions{Events: events, Rule: true})
		})
	})
	wantTexts(t, tt, "Trace", "Fetch tools", "14:02:01", "220ms",
		"Model call", "3.4s", "Tool failed", "14:02:05", "18ms")
}

func TestTraceViewerMarksTheSpansThatMatter(t *testing.T) {
	tt := render(520, 280, func(c *ui.Context) {
		ui.Box(c).Padding(10).Children(func() {
			TraceViewer(c, TraceViewerOptions{Events: []TraceEvent{
				{Name: "Fetch tools", At: "14:02:01",
					Duration: "220ms", Severity: core.Neutral},
				{Name: "Tool failed", At: "14:02:05", Duration: "18ms",
					Severity: core.Danger},
			}})
		})
	})
	light := theme.Light()

	// The ordinary span is the muted ink and the failed one the danger: a
	// trace where every dot is the same colour has nothing to point at.
	neutral := box(t, tt, "Fetch tools")
	wantColor(t, tt, neutral.X+4, neutral.Y+neutral.H/2, light.TextMuted,
		"the ordinary span's muted mark")
	failed := box(t, tt, "Tool failed")
	wantColor(t, tt, failed.X+4, failed.Y+failed.H/2, light.Danger,
		"the failed span's danger mark")
}

func TestTraceViewerNeedsEvents(t *testing.T) {
	wantsPanic(t, "TraceViewer needs at least one TraceEvent",
		func(c *ui.Context) {
			TraceViewer(c, TraceViewerOptions{})
		})
}

func TestTraceViewerInDarkMode(t *testing.T) {
	tt := renderDark(520, 280, func(c *ui.Context) {
		ui.Box(c).Padding(10).Children(func() {
			TraceViewer(c, TraceViewerOptions{Events: []TraceEvent{
				{Name: "Tool failed", At: "14:02:05", Duration: "18ms",
					Severity: core.Danger},
			}})
		})
	})
	wantTexts(t, tt, "Trace", "Tool failed", "14:02:05", "18ms")

	failed := box(t, tt, "Tool failed")
	wantColor(t, tt, failed.X+4, failed.Y+failed.H/2, theme.Dark().Danger,
		"the dark palette's danger, not the light one")
}
