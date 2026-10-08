package devtools_test

import (
	"fmt"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/devtools"
)

// Example shows a response the way somebody reads one when something has gone
// wrong: the status line, the headers with the token already hidden, and the
// body as a tree.
//
// The redaction is the point of the page. ResponseViewer hides the default
// secret keys unless it is explicitly told otherwise, and an empty slice is
// the only way to ask for none — so the safe reading is the default one and
// the cost of getting it wrong is paid on purpose rather than by accident.
func Example() {
	tab := "headers"
	var tabbed, acknowledged string

	body := `{
	  "id": "cb_2871",
	  "count": 1204,
	  "ok": true,
	  "error": null,
	  "tags": ["backend", "flaky"],
	  "customer": {"name": "Riverside Clinic", "seats": 12}
	}`

	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		devtools.ResponseViewer(c, devtools.ResponseViewerOptions{
			Status: "200 OK", Reason: "OK", Method: "GET",
			URL: "https://api.example.com/v1/callbacks",
			Headers: map[string]string{
				"content-type":  "application/json",
				"authorization": "Bearer sk_live_very_secret",
				"x-request-id":  "req_1",
			},
			Body:     body,
			Duration: "42ms",
			Size:     "1.2 kB",
			Tab:      &tab, Tabbed: &tabbed,
		})

		devtools.AlertList(c, devtools.AlertListOptions{
			Alerts: []devtools.Alert{{
				ID: "a1", Title: "Edge returning 502s",
				Detail: "About one request in two hundred.",
				Since:  "4 min ago", Severity: core.Danger,
			}},
			Acknowledged: &acknowledged,
			Dismissed:    new(string),
		})
	}, 760, 640)

	// Output:
}

// ExampleTraceWaterfall shows a trace, and the arithmetic behind it.
//
// SpanBar returns a span's left edge and its width as percentages of the
// whole, so the bars in a waterfall are where the spans really were: four
// spans at 0, 30, 50 and 90 per cent of a hundred milliseconds add up to a
// hundred.
func ExampleTraceWaterfall() {
	ms := time.Millisecond
	spans := []devtools.Span{
		{Name: "GET /callbacks", Start: 0, Duration: 100 * ms},
		{Name: "handler", Start: 2 * ms, Duration: 95 * ms},
		{Name: "db.query", Start: 10 * ms, Duration: 40 * ms},
		{Name: "cache.get", Start: 60 * ms, Duration: 5 * ms, Error: true},
	}

	var selected, picked string
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		devtools.TraceWaterfall(c, devtools.TraceWaterfallOptions{
			Spans: spans, Selected: &selected, SelectedID: &picked, ShowDepth: true,
		})
	}, 820, 400)

	total := devtools.TotalOf(spans)
	fmt.Println("total:", devtools.RoundDuration(total))
	for _, s := range spans {
		left, width := devtools.SpanBar(s, total)
		fmt.Printf("%-16s left %.0f%% width %.0f%%\n", s.Name, left, width)
	}

	// Output:
	// total: 100ms
	// GET /callbacks   left 0% width 100%
	// handler          left 2% width 95%
	// db.query         left 10% width 40%
	// cache.get        left 60% width 5%
}

// ExampleUptimeBar shows ninety days of a service, with the days nobody
// measured left as the surface rather than painted as a success.
func ExampleUptimeBar() {
	now := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)
	var probes []devtools.Slot
	for d := 1; d <= 14; d++ {
		probes = append(probes, devtools.Slot{
			At: now.AddDate(0, 0, -(14 - d)), Up: true,
		})
	}
	// Two bad days, one of them degraded rather than down.
	probes[4].Up, probes[4].Error = false, "timeout"
	probes[9].Error = "slow"

	slots := devtools.UptimeSlots(probes)
	fmt.Println("slots:", len(slots))
	fmt.Println("availability:", devtools.UptimePercent(slots))
	fmt.Println("tone:", devtools.UptimeTone(devtools.UptimePercent(slots)))

	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		devtools.UptimeBar(c, devtools.UptimeBarOptions{Slots: slots, Label: "API"})
	}, 760, 160)

	// Output:
	// slots: 90
	// availability: 0.85714287
	// tone: Danger
}

// ExampleParseJSON shows the parsing on its own, which is how a caller would
// check a response against a fixture before deciding whether to draw it.
//
// The tree is map[string]any all the way down — the same shape every other
// JSON library in every other language produces — so a value parsed here can
// be compared against one parsed anywhere else.
func ExampleParseJSON() {
	value, err := devtools.ParseJSON(`{"id": "cb_2871", "count": 1204, "tags": ["a"]}`)
	fmt.Println("err:", err)

	obj := value.(map[string]any)
	fmt.Println("id:", obj["id"])
	fmt.Println("count:", devtools.FormatNumber(obj["count"].(float64)))
	fmt.Println("kind of tags:", devtools.KindOf(obj["tags"]))
	for _, node := range devtools.Flatten(value) {
		if node.Key != "" {
			fmt.Printf("%-8s %s %s\n", node.Kind, node.Path,
				devtools.JSONPreview(node.Value, 20))
		}
	}

	// Output:
	// err: <nil>
	// id: cb_2871
	// count: 1,204
	// kind of tags: array
	// number   count 1,204
	// string   id cb_2871
	// array    tags [1]
	// string   tags.0 a
}
