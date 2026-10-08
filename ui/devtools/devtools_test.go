package devtools

import (
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
)

// ── JsonViewer ─────────────────────────────────────────────────────────────

func TestJsonViewerShowsTheTree(t *testing.T) {
	var toggled, copied string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		JsonViewer(c, JsonViewerOptions{
			Source: sampleJSON, Toggled: &toggled, Copied: &copied,
		})
	}, 760, 900)
	for _, want := range []string{
		"JSON", "id", "cb_2871", "count", "1,204", "tags", "customer",
		"Riverside Clinic", "boolean", "string",
	} {
		if !tt.HasText(want) {
			t.Errorf("the viewer is missing %q; %q", want, tt.Texts())
		}
	}
}

func TestJsonViewerSaysWhenTheBodyIsNotJSON(t *testing.T) {
	var toggled, copied string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := JsonViewer(c, JsonViewerOptions{
			Source: "<html>not json</html>", Toggled: &toggled, Copied: &copied,
		})
		if r.Valid() {
			t.Error("a body that is not JSON must not report as valid")
		}
	}, 620, 200)
	// An empty tree and a parse failure look identical otherwise, and one of
	// them means the response was fine.
	if !tt.HasText("Not JSON") {
		t.Errorf("the failure is missing; %q", tt.Texts())
	}
}

func TestJsonViewerReportsAValidParse(t *testing.T) {
	var toggled, copied string
	valid := false
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := JsonViewer(c, JsonViewerOptions{Source: sampleJSON, Toggled: &toggled, Copied: &copied})
		valid = r.Valid()
	}, 620, 300)
	if !valid {
		t.Error("the sample should have parsed")
	}
}

func TestJsonViewerOpensABranch(t *testing.T) {
	var toggled, copied string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		JsonViewer(c, JsonViewerOptions{
			Source: `{"a": {"b": 1}}`, Expanded: map[string]bool{"a": false},
			Toggled: &toggled, Copied: &copied,
		})
	}, 620, 300)
	if tt.HasText("\"b\"") && tt.HasText("1") {
		t.Error("a shut branch should not draw its rows")
	}
	if err := tt.Click("Expand a"); err != nil {
		t.Fatal(err)
	}
	if toggled != "a" {
		t.Errorf("opening a branch should report its path, got %q", toggled)
	}
}

func TestJsonViewerNeedsItsPointers(t *testing.T) {
	defer panics(t, "devtools: JsonViewer needs the *string Toggled writes to", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			JsonViewer(c, JsonViewerOptions{Source: `{}`, Copied: new(string)})
		}, 400, 300)
	})
}

// ── JsonTree ───────────────────────────────────────────────────────────────

func TestJsonTreeShowsTheRows(t *testing.T) {
	var toggled string
	value, _ := ParseJSON(`{"a": 1, "b": [2, 3]}`)
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		JsonTree(c, JsonTreeOptions{Nodes: Flatten(value), Toggled: &toggled})
	}, 560, 320)
	for _, want := range []string{"a", "1", "b", "[2]", "b.0", "b.1"} {
		if !tt.HasText(want) {
			t.Errorf("the tree is missing %q; %q", want, tt.Texts())
		}
	}
}

func TestJsonTreeWithNothing(t *testing.T) {
	var toggled string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		JsonTree(c, JsonTreeOptions{Toggled: &toggled})
	}, 460, 200)
	if !tt.HasText("Nothing to show") {
		t.Errorf("an empty tree must say so; %q", tt.Texts())
	}
}

// ── SchemaTree ─────────────────────────────────────────────────────────────

func schema() SchemaNode {
	return SchemaNode{
		Type:        "object",
		Description: "A callback",
		Properties: map[string]SchemaNode{
			"id":    {Type: "string", Description: "The reference"},
			"count": {Type: "number", Description: "How many"},
		},
		Required: []string{"id"},
	}
}

func TestSchemaTreeDrawsTheShape(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		SchemaTree(c, SchemaTreeOptions{Schema: schema()})
	}, 620, 300)
	for _, want := range []string{"Schema", "id", "count", "The reference", "How many"} {
		if !tt.HasText(want) {
			t.Errorf("the schema tree is missing %q; %q", want, tt.Texts())
		}
	}
}

func TestSchemaTreeMarksWhatIsMissing(t *testing.T) {
	value, _ := ParseJSON(`{"id": "cb_1"}`)
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		SchemaTree(c, SchemaTreeOptions{Schema: schema(), Values: Flatten(value)})
	}, 620, 300)
	// "id" is required and present; "count" is neither required nor present.
	if tt.HasText("missing") {
		t.Errorf("nothing required is missing here; %q", tt.Texts())
	}
	if !tt.HasText("absent") {
		t.Errorf("an optional key with no value should say absent; %q", tt.Texts())
	}
}

// ── JsonEditor ─────────────────────────────────────────────────────────────

func TestJsonEditorShowsTheValidityMark(t *testing.T) {
	src := `{"a": 1}`
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		JsonEditor(c, JsonEditorOptions{
			Source: &src, Valid: true,
			Format: func() {}, Minify: func() {},
		})
	}, 620, 360)
	for _, want := range []string{"Editor", "Valid JSON", "Format", "Minify"} {
		if !tt.HasText(want) {
			t.Errorf("the editor is missing %q; %q", want, tt.Texts())
		}
	}
}

func TestJsonEditorSaysWhenTheTextIsNotJSON(t *testing.T) {
	src := `{oh no`
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		JsonEditor(c, JsonEditorOptions{
			Source: &src, Valid: false, Error: "unexpected end of JSON input",
		})
	}, 620, 360)
	if !tt.HasText("Not valid JSON") {
		t.Errorf("the mark is missing; %q", tt.Texts())
	}
}

func TestJsonEditorNeedsItsString(t *testing.T) {
	defer panics(t, "devtools: JsonEditor needs the *string Source writes to", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			JsonEditor(c, JsonEditorOptions{Valid: true})
		}, 460, 300)
	})
}

// ── ResponseViewer ─────────────────────────────────────────────────────────

func TestResponseViewerShowsTheHeaders(t *testing.T) {
	tab := "headers"
	var tabbed string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ResponseViewer(c, ResponseViewerOptions{
			Status: "200 OK", Reason: "OK", Method: "GET",
			URL: "https://api.example.com/v1/callbacks",
			Headers: map[string]string{
				"Content-Type": "application/json",
				"X-Request-Id": "req_1",
			},
			Body: sampleJSON, Duration: "42ms", Size: "1.2 kB",
			Tab: &tab, Tabbed: &tabbed,
		})
	}, 720, 460)
	for _, want := range []string{
		"200 OK", "GET", "https://api.example.com/v1/callbacks",
		"Content-Type", "application/json", "X-Request-Id", "42ms", "1.2 kB",
	} {
		if !tt.HasText(want) {
			t.Errorf("the response is missing %q; %q", want, tt.Texts())
		}
	}
}

func TestResponseViewerRedactsTheSecrets(t *testing.T) {
	tab := "headers"
	var tabbed string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ResponseViewer(c, ResponseViewerOptions{
			Status: "200 OK", Tab: &tab, Tabbed: &tabbed,
			Headers: map[string]string{
				"authorization": "Bearer sk_live_very_secret",
				"accept":        "application/json",
			},
		})
	}, 720, 320)
	if tt.HasText("Bearer sk_live_very_secret") {
		t.Errorf("the token is on the screen; %q", tt.Texts())
	}
	if !tt.HasText(redacted) {
		t.Errorf("the mask is missing; %q", tt.Texts())
	}
	// A header that is not a secret is still shown, or the viewer is hiding
	// things nobody asked it to hide.
	if !tt.HasText("application/json") {
		t.Errorf("an ordinary header should still be there; %q", tt.Texts())
	}
}

func TestResponseViewerCountsWhatItHid(t *testing.T) {
	tab := "headers"
	var tabbed string
	hidden := 0
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := ResponseViewer(c, ResponseViewerOptions{
			Status: "200 OK", Tab: &tab, Tabbed: &tabbed,
			Headers: map[string]string{
				"authorization": "Bearer x",
				"cookie":        "session=abc",
				"accept":        "*/*",
			},
		})
		if r.Redacted() > 0 {
			hidden = r.Redacted()
		}
	}, 720, 320)
	if hidden != 2 {
		t.Errorf("two secret headers, got %d", hidden)
	}
}

func TestResponseViewerShowsTheBodyAsATree(t *testing.T) {
	tab := "body"
	var tabbed string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ResponseViewer(c, ResponseViewerOptions{
			Status: "200 OK", Tab: &tab, Tabbed: &tabbed, Body: sampleJSON,
		})
	}, 720, 700)
	if !tt.HasText("cb_2871") {
		t.Errorf("the body tree is missing the value; %q", tt.Texts())
	}
}

func TestResponseViewerShowsANonJSONBodyAsText(t *testing.T) {
	tab := "body"
	var tabbed string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ResponseViewer(c, ResponseViewerOptions{
			Status: "200 OK", Tab: &tab, Tabbed: &tabbed,
			Body: "<html>a redirect page</html>",
		})
	}, 720, 320)
	// Not JSON is the ordinary case for a file or an HTML page; it is drawn
	// rather than reported as a failure.
	if !tt.HasText("<html>a redirect page</html>") {
		t.Errorf("a plain body should be shown; %q", tt.Texts())
	}
}

func TestResponseViewerNeedsItsTabs(t *testing.T) {
	defer panics(t, "devtools: ResponseViewer needs the *string Tab it shows", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			ResponseViewer(c, ResponseViewerOptions{Status: "200", Tabbed: new(string)})
		}, 500, 300)
	})
}

// ── LogStream ──────────────────────────────────────────────────────────────

func TestLogStreamShowsEveryLine(t *testing.T) {
	level := Debug
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		LogStream(c, LogStreamOptions{Lines: logs(), Level: level})
	}, 720, 420)
	for _, want := range []string{
		"Log", "Lines", "6", "DEBUG", "INFO", "WARN", "ERROR", "FATAL",
		"listening on :8080", "out of memory", "api",
	} {
		if !tt.HasText(want) {
			t.Errorf("the stream is missing %q; %q", want, tt.Texts())
		}
	}
}

func TestLogStreamFiltersByLevel(t *testing.T) {
	level := Error
	shown := 0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := LogStream(c, LogStreamOptions{Lines: logs(), Level: level})
		if r.Shown() > 0 {
			shown = r.Shown()
		}
	}, 720, 420)
	// Error and louder is two lines of the six.
	if shown != 2 {
		t.Errorf("the filter should leave two lines, got %d", shown)
	}
	if tt.HasText("listening on :8080") {
		t.Errorf("a debug line was let through; %q", tt.Texts())
	}
}

func TestLogStreamWithNothingMatching(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		LogStream(c, LogStreamOptions{Lines: logs(), Level: Debug, Search: "zzzz"})
	}, 720, 240)
	if !tt.HasText("No lines match") {
		t.Errorf("the empty state is missing; %q", tt.Texts())
	}
}

func TestLogStreamShowsTheCallersEmptyState(t *testing.T) {
	drew := false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		LogStream(c, LogStreamOptions{
			Lines: nil, Level: Debug,
			Empty: func() {
				ui.Text(c, "Nothing logged yet")
				drew = true
			},
		})
	}, 560, 240)
	if !tt.HasText("Nothing logged yet") || !drew {
		t.Errorf("the caller's empty state should be drawn; %q", tt.Texts())
	}
}

// ── AlertList ──────────────────────────────────────────────────────────────

func alerts() []Alert {
	return []Alert{
		{ID: "a1", Title: "Edge returning 502s", Detail: "About one request in two hundred.",
			Since: "4 min ago", Severity: core.Danger},
		{ID: "a2", Title: "Cache hit rate is low", Detail: "Down to 61%.",
			Since: "20 min ago", Severity: core.Warning, Acknowledged: true},
	}
}

func TestAlertListShowsEveryAlert(t *testing.T) {
	var ack, dismissed string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		AlertList(c, AlertListOptions{Alerts: alerts(), Acknowledged: &ack, Dismissed: &dismissed})
	}, 640, 420)
	for _, want := range []string{
		"Alerts", "Edge returning 502s", "Cache hit rate is low",
		"4 min ago", "Acknowledge", "Dismiss",
	} {
		if !tt.HasText(want) {
			t.Errorf("the list is missing %q; %q", want, tt.Texts())
		}
	}
}

func TestAlertListReportsAnAcknowledge(t *testing.T) {
	var ack, dismissed string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		AlertList(c, AlertListOptions{
			Alerts:       []Alert{{ID: "a1", Title: "Edge returning 502s", Severity: core.Danger}},
			Acknowledged: &ack, Dismissed: &dismissed,
		})
	}, 640, 300)
	if err := tt.Click("Acknowledge"); err != nil {
		t.Fatal(err)
	}
	if ack != "a1" {
		t.Errorf("acknowledging should report the alert, got %q", ack)
	}
}

func TestAlertListWithNothingWrong(t *testing.T) {
	var ack, dismissed string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		AlertList(c, AlertListOptions{Acknowledged: &ack, Dismissed: &dismissed})
	}, 460, 200)
	if !tt.HasText("Nothing is wrong") {
		t.Errorf("an empty list should say so; %q", tt.Texts())
	}
}

// ── IncidentCard ───────────────────────────────────────────────────────────

func TestIncidentCard(t *testing.T) {
	opened := ""
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		IncidentCard(c, IncidentCardOptions{
			Incident: Incident{
				ID: "i1", Title: "Elevated error rate", Impact: "8% of callbacks",
				Started: "09:02", Ended: "09:41", Severity: core.Danger,
				Services: []string{"api", "edge"},
			},
			Open: func(id string) { opened = id },
		})
	}, 640, 300)
	for _, want := range []string{"Elevated error rate", "8% of callbacks",
		"09:02 – 09:41", "api", "edge", "Danger"} {
		if !tt.HasText(want) {
			t.Errorf("the card is missing %q; %q", want, tt.Texts())
		}
	}
	if err := tt.Click("Elevated error rate"); err != nil {
		t.Fatal(err)
	}
	if opened != "i1" {
		t.Errorf("pressing the card should ask for its detail, got %q", opened)
	}
}

func TestIncidentCardSaysOngoing(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		IncidentCard(c, IncidentCardOptions{
			Incident: Incident{ID: "i2", Title: "Disk is filling", Started: "09:02",
				Severity: core.Warning, Ongoing: true},
		})
	}, 640, 260)
	if !tt.HasText("ongoing") {
		t.Errorf("an unfinished incident says ongoing; %q", tt.Texts())
	}
}

// ── TraceWaterfall ─────────────────────────────────────────────────────────

func spans() []Span {
	ms := time.Millisecond
	return []Span{
		{Name: "GET /callbacks", Start: 0, Duration: 100 * ms},
		{Name: "handler", Start: 2 * ms, Duration: 95 * ms},
		{Name: "db.query", Start: 10 * ms, Duration: 40 * ms},
		{Name: "cache.get", Start: 60 * ms, Duration: 5 * ms, Error: true},
	}
}

func TestTraceWaterfallShowsEverySpan(t *testing.T) {
	var selected, picked string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		TraceWaterfall(c, TraceWaterfallOptions{
			Spans: spans(), Selected: &selected, SelectedID: &picked,
			ShowDepth: true,
		})
	}, 820, 400)
	for _, want := range []string{
		"Trace", "Total", "100ms", "4 spans",
		"GET /callbacks", "handler", "db.query", "cache.get", "40ms",
	} {
		if !tt.HasText(want) {
			t.Errorf("the waterfall is missing %q; %q", want, tt.Texts())
		}
	}
}

func TestTraceWaterfallMarksAFailedSpan(t *testing.T) {
	var selected, picked string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		TraceWaterfall(c, TraceWaterfallOptions{
			Spans: spans(), Selected: &selected, SelectedID: &picked,
		})
	}, 820, 400)
	if !tt.HasText("✕ cache.get") {
		t.Errorf("a failed span should be marked; %q", tt.Texts())
	}
}

func TestTraceWaterfallSortsOnRequest(t *testing.T) {
	var selected, picked string
	slowest := ""
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		TraceWaterfall(c, TraceWaterfallOptions{
			Spans: spans(), Selected: &selected, SelectedID: &picked, Sort: true,
		})
		if picked != "" {
			slowest = picked
		}
	}, 820, 400)
	if err := tt.Click("db.query"); err != nil {
		t.Fatal(err)
	}
	// A press must reach the caller, in a sorted waterfall as in an unsorted
	// one — the ordering is a drawing decision and must not change what the
	// rows report.
	if slowest != "db.query" {
		t.Errorf("pressing a row should report it, got %q", slowest)
	}
	if !tt.HasText("GET /callbacks") {
		t.Errorf("sorting should not lose any span; %q", tt.Texts())
	}
}

func TestTraceWaterfallWithNoSpans(t *testing.T) {
	var selected, picked string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		TraceWaterfall(c, TraceWaterfallOptions{
			Selected: &selected, SelectedID: &picked,
		})
	}, 560, 240)
	if !tt.HasText("No spans") {
		t.Errorf("an empty trace should say so; %q", tt.Texts())
	}
}

func TestTraceWaterfallNeedsItsPointers(t *testing.T) {
	defer panics(t, "devtools: TraceWaterfall needs the *string Selected writes to", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			TraceWaterfall(c, TraceWaterfallOptions{SelectedID: new(string)})
		}, 500, 300)
	})
}

// ── ResourceGauge / SystemMonitor ──────────────────────────────────────────

func metrics() []Metric {
	return []Metric{
		{Name: "Heap", Value: 620, Limit: 1024, Unit: "MB", HigherIsWorse: true},
		{Name: "Error rate", Value: 8.1, Limit: 10, Unit: "%", HigherIsWorse: true},
		{Name: "Queue depth", Value: 3},
	}
}

func TestResourceGaugeShowsTheFigureAndTheFraction(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ResourceGauge(c, ResourceGaugeOptions{Metric: metrics()[0]})
	}, 420, 120)
	for _, want := range []string{"Heap", "620MB / 1024MB"} {
		if !tt.HasText(want) {
			t.Errorf("the gauge is missing %q; %q", want, tt.Texts())
		}
	}
}

func TestResourceGaugeNeedsAName(t *testing.T) {
	defer panics(t, "devtools: ResourceGauge needs a metric with a Name", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			ResourceGauge(c, ResourceGaugeOptions{Metric: Metric{Value: 1, Limit: 2}})
		}, 400, 200)
	})
}

func TestSystemMonitorShowsEveryMetric(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		SystemMonitor(c, SystemMonitorOptions{Metrics: metrics(), Refreshed: "2 s ago"})
	}, 560, 420)
	for _, want := range []string{
		"System", "Refreshed", "2 s ago",
		"Heap", "620MB / 1024MB", "Error rate", "8.1% / 10%", "Queue depth",
	} {
		if !tt.HasText(want) {
			t.Errorf("the monitor is missing %q; %q", want, tt.Texts())
		}
	}
}

func TestSystemMonitorWithNothingToWatch(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		SystemMonitor(c, SystemMonitorOptions{})
	}, 460, 200)
	if !tt.HasText("Nothing to watch") {
		t.Errorf("an empty monitor should say so; %q", tt.Texts())
	}
}

// ── ServiceStatus / UptimeBar ──────────────────────────────────────────────

func TestServiceStatusShowsTheStateInWords(t *testing.T) {
	slots := UptimeSlots(append(probes(), Slot{At: day(2026, 14), Up: false}))
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ServiceStatus(c, ServiceStatusOptions{
			Name: "API", State: StateOperational, Detail: "v2.14.3, us-east-1",
			Slots: slots,
		})
	}, 720, 240)
	for _, want := range []string{"API", "operational", "v2.14.3, us-east-1", "over 90 days"} {
		if !tt.HasText(want) {
			t.Errorf("the status is missing %q; %q", want, tt.Texts())
		}
	}
}

func TestServiceStatusWithoutHistoryShowsAFigure(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ServiceStatus(c, ServiceStatusOptions{
			Name: "API", State: StateDegraded, Uptime: 0.994,
		})
	}, 560, 220)
	if !tt.HasText("degraded") {
		t.Errorf("the state should be said; %q", tt.Texts())
	}
	if !tt.HasText("uptime 99.40%") {
		t.Errorf("the figure is missing; %q", tt.Texts())
	}
}

func TestServiceStatusNeedsAName(t *testing.T) {
	defer panics(t, "devtools: ServiceStatus needs a Name", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			ServiceStatus(c, ServiceStatusOptions{State: StateOperational, Uptime: 1})
		}, 400, 200)
	})
}

func probes() []Slot {
	var out []Slot
	for d := 1; d <= 14; d++ {
		out = append(out, Slot{At: day(2026, d), Up: true})
	}
	return out
}

func TestUptimeBarShowsNinetySquaresWorthOfNumber(t *testing.T) {
	slots := UptimeSlots(probes())
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		UptimeBar(c, UptimeBarOptions{Slots: slots, Label: "API"})
	}, 720, 140)
	if !tt.HasText("over 90 days") {
		t.Errorf("the bar should say what window it covers; %q", tt.Texts())
	}
	// Fourteen good days out of the fourteen that were measured.
	if !tt.HasText("100.00%") {
		t.Errorf("the figure is missing; %q", tt.Texts())
	}
}

func TestUptimeBarNeedsALabel(t *testing.T) {
	defer panics(t, "devtools: UptimeBar needs a Label", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			UptimeBar(c, UptimeBarOptions{Slots: UptimeSlots(nil)})
		}, 500, 200)
	})
}

// ── StatCard ───────────────────────────────────────────────────────────────

func TestStatCardShowsTheFigureAndTheChange(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		StatCard(c, StatCardOptions{
			Label: "Callbacks resolved", Value: "1,204",
			Delta: "−340ms", Direction: DeltaDown,
			Severity: core.Warning,
			Spark:    []float64{0.2, 0.6, 0.4, 0.9, 0.7},
		})
	}, 320, 220)
	for _, want := range []string{"Callbacks resolved", "1,204", "−340ms", "Warning"} {
		if !tt.HasText(want) {
			t.Errorf("the card is missing %q; %q", want, tt.Texts())
		}
	}
}

func TestStatCardWithoutAChangeOrHistory(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		StatCard(c, StatCardOptions{Label: "Open callbacks", Value: "27"})
	}, 280, 140)
	if !tt.HasText("Open callbacks") || !tt.HasText("27") {
		t.Errorf("the card did not draw; %q", tt.Texts())
	}
	if tt.HasText("Recent history") {
		t.Errorf("there is no history to draw; %q", tt.Texts())
	}
}

func TestStatCardNeedsALabel(t *testing.T) {
	defer panics(t, "devtools: StatCard needs a Label", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			StatCard(c, StatCardOptions{Value: "27"})
		}, 300, 140)
	})
}

// ── KeyValueInput ──────────────────────────────────────────────────────────

func TestKeyValueInputShowsThePairs(t *testing.T) {
	kvs := KeyValues{{"Accept", "application/json"}, {"X-Request-Id", "req_1"}}
	var newKey, newValue string
	removed := ""
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := KeyValueInput(c, KeyValueInputOptions{
			Values: &kvs, NewKey: &newKey, NewValue: &newValue,
			KeyLabel: "Header", ValueLabel: "Value",
		})
		if r.Removed() != "" {
			removed = r.Removed()
		}
	}, 720, 320)
	for _, want := range []string{"Accept", "X-Request-Id", "Add", "New header", "New value"} {
		if !tt.HasText(want) {
			t.Errorf("the control is missing %q; %q", want, tt.Texts())
		}
	}
	// The values are in editor content, which a text query cannot see: what
	// is on the screen here is the names, and the values belong to whatever
	// is going to send the request.
	if got, _ := kvs.Get("Accept"); got != "application/json" {
		t.Errorf("the caller's set lost a value: %q", got)
	}
	// Nothing was removed by being drawn.
	if removed != "" {
		t.Errorf("a row was removed without being pressed: %q", removed)
	}
}

func TestKeyValueInputRemovesARow(t *testing.T) {
	// Three rows, and the one removed is the first of them. Removing the
	// first is the case a loop gets wrong: the rows below move up into the
	// gap, so anything that steps forward past it has skipped a row without
	// drawing anything to say so.
	kvs := KeyValues{
		{"Accept", "*/*"},
		{"X-Request-Id", "req_1"},
		{"X-Trace", "trace_9"},
	}
	var newKey, newValue string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		KeyValueInput(c, KeyValueInputOptions{
			Values: &kvs, NewKey: &newKey, NewValue: &newValue,
			KeyLabel: "Header", ValueLabel: "Value",
		})
	}, 720, 320)
	// The remove button is named for the row it removes rather than "Remove",
	// because a row of identically-named buttons is a row of buttons that
	// cannot be told apart by anything reading them out.
	if _, ok := tt.Find("Remove Accept"); !ok {
		t.Fatalf("the remove button is not named after its row; %q", tt.Texts())
	}
	if err := tt.Click("Remove Accept"); err != nil {
		t.Fatal(err)
	}
	if len(kvs) != 2 {
		t.Fatalf("one row should have gone, got %v", kvs)
	}
	// And the two that remain keep their order and their values.
	if kvs[0].Key != "X-Request-Id" || kvs[0].Value != "req_1" {
		t.Errorf("the row after the removed one took its place wrongly: %v", kvs[0])
	}
	if kvs[1].Key != "X-Trace" || kvs[1].Value != "trace_9" {
		t.Errorf("the last row moved: %v", kvs[1])
	}
	// Both survivors are still on the screen. This is the half of the
	// assertion that catches the row after the removed one being stepped
	// over: the set is right but one row of it is not drawn, which is a row
	// that vanishes for a frame and reappears on the next.
	for _, want := range []string{"X-Request-Id", "X-Trace"} {
		if !tt.HasText(want) {
			t.Errorf("%q is still in the set but is no longer drawn; %q", want, tt.Texts())
		}
	}
	if tt.HasText("Accept") {
		t.Errorf("the removed row is still drawn; %q", tt.Texts())
	}
}

func TestKeyValueInputHidesASecretButKeepsIt(t *testing.T) {
	kvs := KeyValues{{"Authorization", "Bearer sk_live_secret"}, {"Accept", "*/*"}}
	var newKey, newValue string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		KeyValueInput(c, KeyValueInputOptions{
			Values: &kvs, NewKey: &newKey, NewValue: &newValue,
			KeyLabel: "Header", ValueLabel: "Value",
			Secrets: DefaultSecretKeys(),
		})
	}, 720, 320)
	if tt.HasText("Bearer sk_live_secret") {
		t.Errorf("the secret is on the screen; %q", tt.Texts())
	}
	if !tt.HasText(redacted) {
		t.Errorf("the mask is missing; %q", tt.Texts())
	}
	// Hiding a value from the screen is not hiding it from the request: a
	// control that dropped it would quietly send one without it.
	if got, _ := kvs.Get("Authorization"); got != "Bearer sk_live_secret" {
		t.Errorf("the caller's set was changed: %q", got)
	}
}

func TestKeyValueInputAddsAPair(t *testing.T) {
	kvs := KeyValues{}
	var newKey, newValue string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		KeyValueInput(c, KeyValueInputOptions{
			Values: &kvs, NewKey: &newKey, NewValue: &newValue,
			KeyLabel: "Header", ValueLabel: "Value",
		})
	}, 720, 320)
	if err := tt.Click("New header"); err != nil {
		t.Fatal(err)
	}
	tt.Type("X-Trace")
	if err := tt.Click("Add"); err != nil {
		t.Fatal(err)
	}
	if len(kvs) != 1 || kvs[0].Key != "X-Trace" {
		t.Errorf("the pair was not added: %v", kvs)
	}
}

func TestKeyValueInputNeedsItsLabels(t *testing.T) {
	defer panics(t, "devtools: KeyValueInput needs a KeyLabel and a ValueLabel", func() {
		kvs := KeyValues{}
		var newKey, newValue string
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			KeyValueInput(c, KeyValueInputOptions{
				Values: &kvs, NewKey: &newKey, NewValue: &newValue, KeyLabel: "Header",
			})
		}, 500, 300)
	})
}

// ── APIRequestBuilder ──────────────────────────────────────────────────────

func TestAPIRequestBuilderShowsItsParts(t *testing.T) {
	method, url, query, body := "GET", "https://api.example.com/v1/callbacks", "", ""
	headers := KeyValues{{"Accept", "application/json"}}
	sent := false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := APIRequestBuilder(c, APIRequestBuilderOptions{
			Method: &method, URL: &url, Query: &query, Body: &body,
			Headers: &headers, Sent: &sent, Sendable: true,
		})
		if r.Sent() {
			t.Error("nothing was pressed on this page")
		}
	}, 760, 700)
	for _, want := range []string{
		"Request", "Method", "URL", "Header name", "Header value",
		"Accept", "Query", "Body", "Send",
	} {
		if !tt.HasText(want) {
			t.Errorf("the builder is missing %q; %q", want, tt.Texts())
		}
	}
}

func TestAPIRequestBuilderSaysWhenItCannotSend(t *testing.T) {
	method, url, query, body := "GET", "", "", ""
	headers := KeyValues{}
	sent := false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		APIRequestBuilder(c, APIRequestBuilderOptions{
			Method: &method, URL: &url, Query: &query, Body: &body,
			Headers: &headers, Sent: &sent, Sendable: false,
		})
	}, 760, 600)
	if !tt.HasText("Fill in a URL to send") {
		t.Errorf("the reason should be said; %q", tt.Texts())
	}
}

func TestAPIRequestBuilderNeedsItsPointers(t *testing.T) {
	defer panics(t, "devtools: APIRequestBuilder needs the *string Method and URL", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			APIRequestBuilder(c, APIRequestBuilderOptions{Sent: new(bool)})
		}, 500, 300)
	})
}

// ── DashboardFilterBar ─────────────────────────────────────────────────────

func TestDashboardFilterBarShowsItsControls(t *testing.T) {
	var placed = 30
	var search string
	searched := false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		DashboardFilterBar(c, DashboardFilterBarOptions{
			Filters: []DashboardFilterOptions{
				{Group: "Service", Selected: new(string),
					Options: []string{"api", "edge"}},
				{Group: "Window", Placed: &placed, Min: 0, Max: 100},
				{Search: &search},
			},
			Searched: func() { searched = true },
		})
	}, 860, 340)
	for _, want := range []string{"Filters", "Service", "Window", "Search"} {
		if !tt.HasText(want) {
			t.Errorf("the bar is missing %q; %q", want, tt.Texts())
		}
	}
	// A filter bar whose pointers moved without anything being pressed is a
	// filter bar that has taken over the page's own state.
	if placed != 30 {
		t.Errorf("the range moved on its own: %d", placed)
	}
	if search != "" {
		t.Errorf("the query moved on its own: %q", search)
	}
	if searched {
		t.Error("the search submitted without Enter being pressed")
	}
}

func TestDashboardFilterBarWithNoFilters(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		DashboardFilterBar(c, DashboardFilterBarOptions{})
	}, 560, 200)
	if !tt.HasText("No filters") {
		t.Errorf("an empty bar should say so; %q", tt.Texts())
	}
}

// ── RefreshIntervalSelector ────────────────────────────────────────────────

func TestRefreshIntervalSelector(t *testing.T) {
	seconds := 30
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		RefreshIntervalSelector(c, RefreshIntervalOptions{Seconds: &seconds})
	}, 560, 160)
	for _, want := range []string{"Off", "5s", "30s", "1m", "5m"} {
		if !tt.HasText(want) {
			t.Errorf("the selector is missing %q; %q", want, tt.Texts())
		}
	}
}

func TestRefreshIntervalSelectorChangesTheValue(t *testing.T) {
	seconds := 30
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		RefreshIntervalSelector(c, RefreshIntervalOptions{Seconds: &seconds})
	}, 560, 160)
	if err := tt.Click("5m"); err != nil {
		t.Fatal(err)
	}
	if seconds != 300 {
		t.Errorf("choosing 5m should write 300 seconds, got %d", seconds)
	}
}

func TestRefreshIntervalSelectorRefusesAValueItCannotShow(t *testing.T) {
	defer panics(t, "which is not one of its intervals", func() {
		seconds := 17
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			RefreshIntervalSelector(c, RefreshIntervalOptions{Seconds: &seconds})
		}, 500, 200)
	})
}

func TestRefreshIntervalSelectorRefusesANegativeInterval(t *testing.T) {
	defer panics(t, "a negative interval is not a slow one", func() {
		seconds := -5
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			RefreshIntervalSelector(c, RefreshIntervalOptions{Seconds: &seconds})
		}, 500, 200)
	})
}

// ── the dark palette ───────────────────────────────────────────────────────

func TestDevtoolsSurvivesTheDarkPalette(t *testing.T) {
	// Every sink the page writes to is declared here and read at the end. A
	// dark pass that collects answers and does not look at them cannot catch
	// a control firing on its own, which is the failure a dark palette exists
	// to surface.
	var (
		toggled, copied, tabbed, ack, dismissed, picked string
		query, url, body, search                        string
		tab                                             = "headers"
	)
	var sent, saved, started, stopped, reset bool
	var placed, seconds = 30, 30
	// Seeded with a verb the selector offers, so the trigger is showing a real
	// choice rather than its placeholder — and so "the method moved on its
	// own" is an assertion about a change rather than about being empty.
	method := "GET"
	var service string
	headers := KeyValues{{"Authorization", "Bearer sk_live_secret"}, {"Accept", "*/*"}}
	kv := KeyValues{{"X-Trace", "req_1"}}
	kvs := KeyValues{}
	var kvNewKey, kvNewValue, kvsNewKey, kvsNewValue string
	schema := SchemaNode{
		Type: "object", Description: "A callback",
		Properties: map[string]SchemaNode{
			"id":    {Type: "string", Description: "The reference"},
			"count": {Type: "number", Description: "How many"},
		},
		Required: []string{"id"},
	}

	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: core.Dark})

		JsonViewer(c, JsonViewerOptions{
			Source: sampleJSON, Toggled: &toggled, Copied: &copied,
			Expanded: map[string]bool{"customer": true},
		})
		JsonTree(c, JsonTreeOptions{
			Nodes: Flatten(MustParseJSON(`{"a":1,"b":[2]}`)), Toggled: &toggled,
		})
		JsonEditor(c, JsonEditorOptions{
			Source: &body, Valid: true, Format: func() {}, Minify: func() {},
		})
		SchemaTree(c, SchemaTreeOptions{
			Schema: schema, Values: Flatten(MustParseJSON(`{"id":"cb_1"}`)),
		})
		ResponseViewer(c, ResponseViewerOptions{
			Status: "500 Internal Server Error", Reason: "Internal Server Error",
			Method: "POST", URL: "https://api.example.com/v1/callbacks",
			Headers: map[string]string{
				"content-type": "application/json", "authorization": "Bearer x",
			},
			Body: sampleJSON, Duration: "1.2 s", Size: "4.1 kB",
			Tab: &tab, Tabbed: &tabbed,
		})
		LogStream(c, LogStreamOptions{Lines: logs(), Level: Debug})
		AlertList(c, AlertListOptions{
			Alerts: alerts(), Acknowledged: &ack, Dismissed: &dismissed,
		})
		IncidentCard(c, IncidentCardOptions{
			Incident: Incident{ID: "i1", Title: "Elevated error rate",
				Started: "09:02", Ended: "09:41", Severity: core.Danger,
				Services: []string{"api"}},
			Open: func(string) { started = true },
		})
		TraceWaterfall(c, TraceWaterfallOptions{
			Spans: spans(), Selected: &picked, SelectedID: &picked, ShowDepth: true,
		})
		SystemMonitor(c, SystemMonitorOptions{
			Metrics: metrics(), Refreshed: "2 s ago",
		})
		ResourceGauge(c, ResourceGaugeOptions{Metric: metrics()[1]})
		ServiceStatus(c, ServiceStatusOptions{
			Name: "API", State: StateDegraded, Detail: "v2.14.3",
			Slots: UptimeSlots(probes()),
		})
		StatCard(c, StatCardOptions{
			Label: "Callbacks resolved", Value: "1,204",
			Delta: "+12%", Direction: DeltaUp, Spark: []float64{0.2, 0.6, 0.9},
		})
		UptimeBar(c, UptimeBarOptions{Slots: UptimeSlots(probes()), Label: "API"})
		KeyValueInput(c, KeyValueInputOptions{
			Values: &kv, NewKey: &kvNewKey, NewValue: &kvNewValue,
			KeyLabel: "Header", ValueLabel: "Value",
			Secrets: DefaultSecretKeys(),
		})
		KeyValueInput(c, KeyValueInputOptions{
			Values: &kvs, NewKey: &kvsNewKey, NewValue: &kvsNewValue,
			KeyLabel: "Field", ValueLabel: "Value",
		})
		APIRequestBuilder(c, APIRequestBuilderOptions{
			Method: &method, URL: &url, Query: &query, Body: &body,
			Headers: &headers, Sent: &sent, Sendable: true,
		})
		DashboardFilterBar(c, DashboardFilterBarOptions{
			Filters: []DashboardFilterOptions{
				{Group: "Service", Selected: &service,
					Options: []string{"api", "edge"}},
				{Group: "Window", Placed: &placed, Min: 0, Max: 100},
				{Search: &search},
			},
			Searched: func() { stopped = true },
		})
		RefreshIntervalSelector(c, RefreshIntervalOptions{Seconds: &seconds})
	}, 1400, 2400)

	for _, want := range []string{
		"JSON", "Editor", "Schema", "Response", "Log", "Alerts",
		"Elevated error rate", "Trace", "GET /callbacks", "System",
		"Heap", "API", "Callbacks resolved", "1,204", "Request",
		"Filters", "30s",
	} {
		if !tt.HasText(want) {
			t.Errorf("the dark pass is missing %q; %q", want, tt.Texts())
		}
	}

	// The secret must still be a secret in the dark: a token on a screenshot
	// taken in a dim room is the same leak.
	if tt.HasText("Bearer sk_live_secret") {
		t.Error("a secret is on the screen in the dark palette")
	}
	// Seventeen components on one frame is seventeen chances to write to a
	// caller's pointer. Nothing here was pressed, so nothing may have moved.
	if started || stopped || sent {
		t.Errorf("something fired without being pressed: started=%v stopped=%v sent=%v",
			started, stopped, sent)
	}
	for _, got := range []struct{ what, v string }{
		{"toggled", toggled}, {"copied", copied}, {"tabbed", tabbed},
		{"acknowledged", ack}, {"dismissed", dismissed}, {"picked", picked},
		{"query", query}, {"url", url}, {"body", body}, {"search", search},
		{"a new key", kvNewKey}, {"a new value", kvNewValue},
	} {
		if got.v != "" {
			t.Errorf("%s was written to without being pressed: %q", got.what, got.v)
		}
	}
	if placed != 30 || seconds != 30 || tab != "headers" ||
		method != "GET" || service != "" {
		t.Errorf("a control moved on its own: placed=%d seconds=%d tab=%q "+
			"method=%q service=%q", placed, seconds, tab, method, service)
	}
	if len(kv) != 1 || len(kvs) != 0 || len(headers) != 2 {
		t.Errorf("a set changed without a key being pressed: kv=%v kvs=%v headers=%v",
			kv, kvs, headers)
	}
	_ = saved
	_ = reset
}

// ── helpers ────────────────────────────────────────────────────────────────

// panics asserts that building the thing panics with a message containing
// want. Anything a component cannot be given is a panic rather than a guess,
// so these are as much a part of the contract as what it draws.
func panics(t *testing.T, want string, fn func()) {
	t.Helper()
	defer func() {
		t.Helper()
		r := recover()
		if r == nil {
			t.Fatalf("expected a panic saying %q", want)
		}
		got, ok := r.(string)
		if !ok {
			t.Fatalf("the panic value is %T, not a string: %v", r, r)
		}
		if !strings.Contains(got, want) {
			t.Errorf("panic = %q, want it to contain %q", got, want)
		}
	}()
	fn()
}
