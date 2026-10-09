package pages

import (
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/devtools"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/showcase"
	"github.com/HycJack/MintUI/ui/theme"
)

// Demo state outlives the frame: every demo below hands its component
// a pointer, and a pointer into a frame-local is a click the next
// frame undoes — a tab that will not switch, a dropdown that snaps
// shut, a slider that springs back.
var (
	tab     = "body"
	headTab = "headers"
	sorted  = "db.query"
	method  = "POST"
	url     = ""
	headers = devtools.KeyValues{
		{Key: "Accept", Value: "application/json"},
		{Key: "Authorization", Value: "Bearer cb_live_4f9a2b7c1d8e0a5b"},
	}
	devtools_query = "state=received&limit=20"
	body           = `{ "callback_id": "cb_2871" }`
	sent           = false
	ok             = "https://api.northgate.health/v1/callbacks/2871"
	values         = devtools.KeyValues{
		{Key: "Accept", Value: "application/json"},
		{Key: "Authorization", Value: "Bearer cb_live_4f9a2b7c1d8e0a5b"},
		{Key: "X-Request-Id", Value: "01HV8QZ3M4"},
	}
	seconds = 30
	off     = 0
)

// The page is one debugging session: a response arrives, it is read, a
// request is built to replace it, a trace says where the time went, and a
// dashboard says whether any of that is unusual today.
//
// The order is the order the work happens in rather than the order the files
// are named in, because a gallery of a debugging package is only useful if it
// can be read as a session — "here is the body, here is the tree, here is the
// schema beside it" is a story, and "json.go, metrics.go, response.go" is an
// index.
//
// Everything on the page is fixed: the same response, the same eight
// headers, the same ninety days of uptime, the same ten spans. A page that
// drew differently each time it was opened could not be compared against the
// last one.

func init() {
	showcase.Register(DevtoolsPage())
}

// DevtoolsPage is the gallery's page for ui/devtools.
func DevtoolsPage() showcase.Page {
	return showcase.Page{
		Package: "devtools",
		Title:   "ui/devtools — 一次调试会话",
		Note:    "响应、JSON、树、schema、日志、trace、指标、告警、请求构造器",
		Width:   1000,
		Height:  5620,
		Want: []string{
			// the response and its body
			"Response", "200 OK", "https://api.northgate.health/v1/callbacks",
			"Headers", "Body", "content-type", "x-request-id", "••••••••",
			// the JSON of it
			"JSON", "object", "array", "string", "number", "boolean", "null",
			"authorization", "callback_id", "callback_url",
			"Not JSON", "Nothing to show", "Nothing matches",
			// the tree, the schema and the editor
			"Schema", "missing", "absent", "Editor", "Valid JSON",
			"Not valid JSON", "Format", "Minify",
			// logs
			"Log", "Lines", "DEBUG", "INFO", "WARN", "ERROR", "FATAL",
			// trace and uptime
			"Trace", "Total", "8 spans", "No spans",
			"over 90 days", "operational", "degraded",
			// metrics
			"Requests", "p95 latency", "Queue depth", "Cache hit rate",
			"8% left", "3.4GB / 8GB",
			// dashboard
			"Alerts", "Nothing is wrong", "Acknowledge", "Dismiss",
			"Queue depth above 80 for 10 minutes", "Callback queue slowly draining",
			"ongoing",
			// the request builder
			"Request", "Method", "URL", "Send", "Fill in a URL to send",
			"Header name", "Header value", "Add", "Remove Accept",
			"Refresh", "Off", "30s", "5m",
			// pure functions
			`ParseJSON(body)`, `FormatJSON(src)`, `MinifyJSON(src)`,
			`Redact(value, DefaultSecretKeys())`, `Flatten(value)`,
			`FilterLogs(lines, Warn, "db", 2)`, `RoundDuration(340*time.Millisecond)`,
			`DepthOf(spans, 3)`, `KindOf(value)`, `MarshalIndent(value)`,
			`SlowestSpans(spans, 3)`, `TotalOf(spans)`,
			`StatusTone("404 Not Found")`,
		},
		Render: func(c *ui.Context) {
			devtoolsPage(c)
		},
	}
}

func devtoolsPage(c *ui.Context) {
	responseSection(c)
	jsonSection(c)
	loglineSection(c)
	spanSection(c)
	numbersSection(c)
	alertSection(c)
	requestSection(c)
	devtoolsPureSection(c)
}

// ── the response ─────────────────────────────────────────────────────────

// responseSection is a response as a person reads one: the line, the headers
// with the secret ones hidden, and the body as a tree because the body is
// JSON.
func responseSection(c *ui.Context) {
	showcase.Section(c, "响应 · ResponseViewer / StatusTone")
	showcase.Field(c, "ResponseViewer — 默认藏起 token；StatusTone 决定状态行的颜色")
	// A frame of the viewer's own height: it fills the height it is given, so
	// drawn straight into the page it would take the page's height and leave
	// everything below it at the bottom of a very long empty column.
	ui.Box(c).FillWidth().Height(unit(c, 112)).Children(func() {
		viewer := devtools.ResponseViewer(c, devtools.ResponseViewerOptions{
			Status: "200", Reason: "OK", Method: "POST",
			URL:      "https://api.northgate.health/v1/callbacks",
			Duration: "184ms", Size: "2.4 kB",
			Body: responseBody(),
			Tab:  &tab, Tabbed: new(string),
			Height: unit(c, 40),
			Headers: map[string]string{
				"content-type":          "application/json; charset=utf-8",
				"x-request-id":          "01HV8QZ3M4",
				"date":                  "Tue, 10 Mar 2026 09:12:44 GMT",
				"authorization":         "Bearer cb_live_4f9a2b7c1d8e0a5b",
				"cache-control":         "no-store",
				"x-ratelimit-remaining": "4,996",
			},
		})
		resultLine(c, "ResponseViewer.Redacted()", itoaOf(viewer.Redacted()))
	})

	showcase.Field(c, "StatusTone — 状态码到颜色只有一条路")
	ui.Row(c).FillWidth().Gap(unit(c, 2)).Children(func() {
		for _, s := range []struct{ code, reason string }{
			{"204", "No Content"}, {"302", "Found"}, {"404", "Not Found"},
			{"429", "Too Many Requests"}, {"500", "Internal Server Error"},
		} {
			statusPill(c, s.code+" "+s.reason)
		}
	})
	showcase.Field(c, "同一行，Headers 那一页")
	ui.Box(c).FillWidth().Height(unit(c, 48)).Children(func() {
		devtools.ResponseViewer(c, devtools.ResponseViewerOptions{
			Status: "401", Reason: "Unauthorized", Method: "GET",
			URL:  "https://api.northgate.health/v1/callbacks/2871",
			Body: "callback 2871 is not in this account",
			Tab:  &headTab, Tabbed: new(string), Height: unit(c, 18),
			Headers: map[string]string{
				"content-type":     "text/plain; charset=utf-8",
				"x-request-id":     "01HV8QZ3M5",
				"set-cookie":       "session=abc123; Path=/; HttpOnly",
				"www-authenticate": "Bearer realm=\"northgate\"",
			},
		})
	})
}

// responseBody is the JSON a callback endpoint answers with: a token in it, a
// list, a null and a number, because a viewer that cannot draw one of those
// cannot draw a real response.
func responseBody() string {
	return `{
  "callback_id": "cb_2871",
  "callback_url": "https://northgate.health/hooks/cb_2871",
  "state": "received",
  "attempts": 2,
  "duration_ms": 184.5,
  "authorization": "Bearer cb_live_4f9a2b7c1d8e0a5b",
  "customer": {
    "name": "Riverside Clinic",
    "callback": null,
    "tags": ["hvac", "recurring"]
  },
  "delivered": true
}`
}

// ── the body ─────────────────────────────────────────────────────────────

// jsonSection is the same body three ways: parsed and walked into rows, drawn
// as a tree of the caller's own rows, and written back out as text. The
// schema stands beside the response because a body on its own says nothing
// about what it should have been.
func jsonSection(c *ui.Context) {
	showcase.Section(c, "JSON · JsonViewer / JsonTree / JsonEditor / JSONPreview / SchemaTree")

	toggled := showcase.State(c, "devtools.208.toggled", "")
	copied := showcase.State(c, "devtools.208.copied", "")
	ui.Row(c).FillWidth().Gap(unit(c, 4)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).WidthPercent(50).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "JsonViewer — 自己解析，每一行都带自己的类型")
			open := map[string]bool{
				"": true, "customer": true, "customer.tags": true,
			}
			devtools.JsonViewer(c, devtools.JsonViewerOptions{
				Source: responseBody(), Expanded: open,
				Toggled: toggled, Copied: copied, Height: unit(c, 36),
			})
			resultLine(c, "JsonViewer.Toggled() · Copied()",
				quoteOrEmpty(*toggled)+" · "+quoteOrEmpty(*copied))

			showcase.Field(c, "JsonViewer — 搜的是行，不是重新解析")
			devtools.JsonViewer(c, devtools.JsonViewerOptions{
				Source: responseBody(), Expanded: open,
				Toggled: toggled, Copied: copied, Search: "cb_2871",
			})
			showcase.Field(c, "JsonTree — 一行都没有的时候说没有")
			devtools.JsonTree(c, devtools.JsonTreeOptions{
				Nodes: nil, Toggled: toggled,
			})
			showcase.Field(c, "JsonViewer — 搜不到的时候说搜不到")
			devtools.JsonViewer(c, devtools.JsonViewerOptions{
				Source: responseBody(), Toggled: toggled, Copied: copied,
				Search: "zzz",
			})
			showcase.Field(c, "JsonViewer — 不是 JSON 的时候说解析器说了什么")
			devtools.JsonViewer(c, devtools.JsonViewerOptions{
				Source:  "<!doctype html><title>502</title>",
				Toggled: toggled, Copied: copied,
			})
		})

		ui.Column(c).WidthPercent(48).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "JsonTree — 同样的行，不要工具栏")
			secret := devtools.DefaultSecretKeys()
			value, _ := devtools.ParseJSON(responseBody())
			clean := devtools.Redact(value, secret)
			nodes := devtools.Flatten(clean)
			for _, path := range devtools.RedactedPaths(value, secret) {
				for i := range nodes {
					if nodes[i].Path == path {
						nodes[i].Redacted = true
					}
				}
			}
			shut := map[string]bool{"": true, "customer": true}
			devtools.JsonTree(c, devtools.JsonTreeOptions{
				Nodes: nodes, Expanded: shut, Toggled: toggled, Height: unit(c, 22),
			})
			resultLine(c, "JSONPreview(value, 40)", devtools.JSONPreview(nodes[1].Value, 40))

			showcase.Field(c, "SchemaTree — 该有的没有，用 missing 说")
			devtools.SchemaTree(c, devtools.SchemaTreeOptions{
				Schema: callbackSchema(), Values: nodes,
			})

			showcase.Field(c, "JsonEditor — 有效无效都说出来，而且可改")
			editor := showcase.State(c, "devtools.268.editor", responseBody())
			bad := showcase.State(c, "devtools.268.bad", "  { oops")
			devtools.JsonEditor(c, devtools.JsonEditorOptions{
				Source: editor, Valid: devtools.ValidJSON(*editor), Height: 8,
				Format: func() {}, Minify: func() {},
			})
			devtools.JsonEditor(c, devtools.JsonEditorOptions{
				Source: bad, Valid: false, Error: "unexpected end of JSON input",
				Height: 4, Format: func() {}, Minify: func() {},
			})
		})
	})
}

// callbackSchema is the shape the endpoint promises. It asks for two keys the
// response does not have, so the tree has something to say about a body that
// is wrong rather than only about one that is right.
func callbackSchema() devtools.SchemaNode {
	return devtools.SchemaNode{
		Type: "object", Description: "a callback as the queue stored it",
		Required: []string{"callback_id", "customer"},
		Properties: map[string]devtools.SchemaNode{
			"callback_id":   {Type: "string", Description: "cb_ followed by digits"},
			"callback_url":  {Type: "string", Description: "where the machine posts"},
			"state":         {Type: "string", Description: "received, queued or sent"},
			"attempts":      {Type: "integer", Description: "how many times it was tried"},
			"duration_ms":   {Type: "number", Description: "how long the machine took"},
			"authorization": {Type: "string", Description: "the machine's own key"},
			"delivered":     {Type: "boolean", Description: "whether it got through"},
			"customer": {
				Type: "object", Description: "who the callback is about",
				Required: []string{"name", "phone"},
				Properties: map[string]devtools.SchemaNode{
					"name":     {Type: "string", Description: "as it was typed in"},
					"phone":    {Type: "string", Description: "for the on-call rota"},
					"callback": {Type: "string", Description: "the machine's own name for it"},
					"tags": {Type: "array", Description: "what it is about",
						Items: &devtools.SchemaNode{Type: "string"}},
				},
			},
		},
	}
}

// ── the log ──────────────────────────────────────────────────────────────

// logSection is a stream and what a reader does to it: the level is a filter
// of "this and louder", the search covers the message and the source, and the
// limit keeps the newest.
func loglineSection(c *ui.Context) {
	showcase.Section(c, "日志 · LogStream / FilterLogs")
	lines := logLines()
	// Three streams, three levels: the buttons write whichever one the reader
	// pressed, so a level read from a constant would be this frame's.
	allLevel := showcase.State(c, "devtools.logs.all", devtools.Info)
	warnLevel := showcase.State(c, "devtools.logs.warn", devtools.Warn)
	noneLevel := showcase.State(c, "devtools.logs.none", devtools.Debug)
	showcase.Field(c, "LogStream — 五级都在，最新的在下面；筛选之后是几行也写出来")
	ui.Row(c).FillWidth().Gap(unit(c, 4)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).WidthPercent(62).Shrink(0).Gap(unit(c, 2)).Children(func() {
			stream := devtools.LogStream(c, devtools.LogStreamOptions{
				Lines: lines, Level: allLevel, Follow: true, Height: unit(c, 40),
			})
			resultLine(c, "LogStream.Shown()", itoaOf(stream.Shown()))

			showcase.Field(c, "只要 WARN 以上的 · 搜 db · 最多两行")
			ui.Row(c).FillWidth().Gap(unit(c, 2)).Children(func() {
				ui.Column(c).WidthPercent(60).Shrink(0).Children(func() {
					devtools.LogStream(c, devtools.LogStreamOptions{
						Lines: lines, Level: warnLevel, Search: "db",
						Limit: 2, Height: unit(c, 14),
					})
				})
				ui.Column(c).WidthPercent(38).Shrink(0).Gap(unit(c, 1)).Children(func() {
					for _, lvl := range []devtools.LogLevel{
						devtools.Debug, devtools.Warn, devtools.Error, devtools.Fatal,
					} {
						resultLine(c, lvl.String(), itoaOf(len(
							devtools.FilterLogs(lines, lvl, "", 0))))
					}
				})
			})
		})

		ui.Column(c).WidthPercent(36).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "LogStream — 筛完什么都没有")
			devtools.LogStream(c, devtools.LogStreamOptions{
				Lines: lines, Level: noneLevel, Search: "nothing says this",
				Empty: func() {
					display.Text(c, "Nothing in the last hour says that.",
						display.TextOptions{Muted: true})
				},
			})
			showcase.Field(c, "每一级自己是什么颜色")
			ui.Row(c).Gap(unit(c, 1.5)).Children(func() {
				for _, l := range []devtools.LogLevel{
					devtools.Debug, devtools.Info, devtools.Warn,
					devtools.Error, devtools.Fatal,
				} {
					logPill(c, l)
				}
			})
		})
	})
}

// logLines is one run of a service's log, with the three levels a night
// actually contains: a warning that was handled, a failure, and the line
// after it that says what it did about it.
func logLines() []devtools.LogLine {
	base := time.Date(2026, 3, 10, 9, 12, 0, 0, time.UTC)
	at := func(sec int) time.Time { return base.Add(time.Duration(sec) * time.Second) }
	return []devtools.LogLine{
		{At: at(0), Level: devtools.Info, Source: "server", Message: "listening on :8080"},
		{At: at(1), Level: devtools.Debug, Source: "server", Message: "config reloaded from disk"},
		{At: at(4), Level: devtools.Info, Source: "http", Message: "POST /v1/callbacks 200 in 184ms"},
		{At: at(5), Level: devtools.Warn, Source: "db", Message: "pool at 8 of 10 connections"},
		{At: at(9), Level: devtools.Error, Source: "db", Message: "query took 2.4s: SELECT callbacks"},
		{At: at(9), Level: devtools.Info, Source: "db", Message: "connection returned to the pool"},
		{At: at(12), Level: devtools.Debug, Source: "http", Message: "cache hit for callback 2871"},
		{At: at(15), Level: devtools.Warn, Source: "queue", Message: "retry 2 of 5 for cb_2871"},
		{At: at(18), Level: devtools.Fatal, Source: "worker", Message: "gave up on cb_2871"},
		{At: at(19), Level: devtools.Info, Source: "worker", Message: "parked cb_2871 for a human"},
	}
}

// ── the trace ────────────────────────────────────────────────────────────

// traceSection is where the time went, and the ninety days a service has
// been up. The two are beside each other because they are the two questions
// asked first: what was slow, and has it been slow all week.
func spanSection(c *ui.Context) {
	showcase.Section(c, "Trace 与在线率 · TraceWaterfall / SpanBar / UptimeBar / UptimeSlots / ServiceStatus")

	selected := showcase.State(c, "devtools.396.selected", "db.query")
	selectedID := showcase.State(c, "devtools.396.selectedID", "")
	spans := traceSpans()
	ui.Row(c).FillWidth().Gap(unit(c, 4)).AlignItems(ui.Start).Children(func() {
		// Grow rather than WidthPercent: the waterfall's rows are the widest
		// thing on the page, and a percentage that its own content is wider
		// than pushes the column beside it off the edge.
		ui.Column(c).Grow(1).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "TraceWaterfall — 每根条的长度就是它真的占了多久")
			devtools.TraceWaterfall(c, devtools.TraceWaterfallOptions{
				Spans: spans, Selected: selected, SelectedID: selectedID,
				ShowDepth: true,
			})
			resultLine(c, "SelectedID()", quoteOrEmpty(*selectedID))
			showcase.Field(c, "Sort — 慢的排前面；排序会丢掉嵌套")
			devtools.TraceWaterfall(c, devtools.TraceWaterfallOptions{
				Spans: spans, Selected: &sorted, SelectedID: selectedID, Sort: true,
			})
			showcase.Field(c, "什么都没有的 trace")
			devtools.TraceWaterfall(c, devtools.TraceWaterfallOptions{
				Spans: nil, Selected: selected, SelectedID: selectedID,
			})
		})

		ui.Column(c).Grow(1).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "ServiceStatus 和 UptimeBar — 方块是探测，不是天数")
			for _, svc := range []devtools.ServiceStatusOptions{
				{Name: "API", State: devtools.StateOperational,
					Detail: "eu-west · v2.4.0", Uptime: 0.9998,
					Slots: devtools.UptimeSlots(probes("api", 0.9998))},
				{Name: "Webhook delivery", State: devtools.StateDegraded,
					Detail: "eu-west · v2.4.0", Uptime: 0.9942,
					Slots: devtools.UptimeSlots(probes("hook", 0.9942))},
				{Name: "Callback queue", State: devtools.StateDown,
					Detail: "eu-west · v2.3.9 since 08:40", Uptime: 0.91,
					Slots: devtools.UptimeSlots(probes("queue", 0.91))},
			} {
				devtools.ServiceStatus(c, svc)
			}
			showcase.Field(c, "UptimeBar — 没有记录的每一天是灰的，不是绿的")
			devtools.UptimeBar(c, devtools.UptimeBarOptions{
				Label: "API", Slots: devtools.UptimeSlots(probes("api", 0.9)),
			})
		})
	})
}

// traceSpans is one request's trace, parent before child and with the two
// spans that failed marked, because a waterfall is scanned for the red one.
func traceSpans() []devtools.Span {
	return []devtools.Span{
		{Name: "POST /v1/callbacks", Start: 0, Duration: 184 * time.Millisecond},
		{Name: "auth", Start: 2 * time.Millisecond, Duration: 6 * time.Millisecond},
		{Name: "read body", Start: 9 * time.Millisecond, Duration: 4 * time.Millisecond},
		{Name: "db.query", Start: 14 * time.Millisecond, Duration: 142 * time.Millisecond},
		{Name: "db.pool wait", Start: 20 * time.Millisecond, Duration: 11 * time.Millisecond},
		{Name: "db.row", Start: 31 * time.Millisecond, Duration: 3 * time.Millisecond, Error: true},
		{Name: "render", Start: 158 * time.Millisecond, Duration: 5 * time.Millisecond},
		{Name: "deliver webhook", Start: 163 * time.Millisecond, Duration: 21 * time.Millisecond},
	}
}

// probes is ninety days of a service being probed, with the given share of
// them bad. The days that are not probed at all are unknown rather than up,
// which is the rule that makes the bar honest.
func probes(name string, uptime float32) []devtools.Slot {
	day := time.Date(2026, 3, 10, 9, 0, 0, 0, time.UTC)
	var out []devtools.Slot
	for i := devtools.UptimeSlotsDays; i >= 0; i-- {
		at := day.AddDate(0, 0, -i)
		switch {
		case i > 60:
			// The service only went up three months ago: before that there
			// is no record, which the bar draws as grey rather than green.
		case int(float32(i)*10)%9 < int((1-uptime)*10):
			out = append(out, devtools.Slot{At: at, Up: false, Error: "connect: refused"})
		case int(float32(i)*7)%5 == 0:
			out = append(out, devtools.Slot{At: at, Up: true, Error: "slow: 1.2s"})
		default:
			out = append(out, devtools.Slot{At: at, Up: true})
		}
		_ = name
	}
	return out
}

// ── the numbers ──────────────────────────────────────────────────────────

// metricSection is what the machine is doing and what a dashboard says about
// it. The gauges are beside the cards because a card is a number and a gauge
// is that number against a limit.
func numbersSection(c *ui.Context) {
	showcase.Section(c, "指标 · SystemMonitor / ResourceGauge / MetricFigure / StatCard / MetricCard")
	ui.Row(c).FillWidth().Gap(unit(c, 4)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).WidthPercent(50).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "SystemMonitor — 数字是调用方的，本包不采样也不存历史")
			devtools.SystemMonitor(c, devtools.SystemMonitorOptions{
				Refreshed: "10 March 09:14:02",
				Metrics:   metrics(),
			})
			showcase.Field(c, "ResourceGauge — 没有上限的指标画成空的，不是满的")
			for _, m := range metrics() {
				devtools.ResourceGauge(c, devtools.ResourceGaugeOptions{Metric: m})
			}
			showcase.Field(c, "SystemMonitor — 一个指标也没有")
			devtools.SystemMonitor(c, devtools.SystemMonitorOptions{
				Empty: func() {
					display.Text(c, "This machine is not being watched.",
						display.TextOptions{Muted: true})
				},
			})
		})

		ui.Column(c).WidthPercent(48).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "StatCard — 涨跌由方向决定，不由正负号决定")
			ui.Row(c).FillWidth().Gap(unit(c, 2)).Children(func() {
				for _, card := range []devtools.StatCardOptions{
					{Label: "Requests", Value: "48,204", Delta: "+12%",
						Direction: devtools.DeltaUp, Spark: spark(0.3, 1, 12)},
					{Label: "p95 latency", Value: "184ms", Delta: "−340ms",
						Direction: devtools.DeltaUp, Spark: spark(1, 0.6, 9)},
					{Label: "Error rate", Value: "2.1%", Delta: "+1.4%",
						Direction: devtools.DeltaDown, Severity: core.Danger,
						Spark: spark(0.2, 0.9, 4)},
				} {
					card := card
					ui.Box(c).Grow(1).Children(func() {
						devtools.StatCard(c, card)
					})
				}
			})
			showcase.Field(c, "MetricFigure — 越小越好的指标画的是剩下的，不是用掉的")
			for _, m := range metrics() {
				ui.Row(c).FillWidth().Gap(unit(c, 2)).AlignItems(ui.Center).Children(func() {
					ui.Text(c, m.Name).TextColor(core.Tokens(c).TextMuted).
						FontSize(core.FontSize(c, theme.CaptionSize)).Width(unit(c, 18)).Shrink(0)
					ui.Text(c, devtools.MetricFigure(m)).TextColor(core.Tokens(c).Text).
						Font(monoFamily).FontSize(core.FontSize(c, theme.MetaSize)).SingleLine()
				})
			}
		})
	})
}

func metrics() []devtools.Metric {
	return []devtools.Metric{
		{Name: "Memory", Value: 3.4, Limit: 8, Unit: "GB", HigherIsWorse: true},
		{Name: "Queue depth", Value: 42, Limit: 100, Unit: "", HigherIsWorse: true},
		{Name: "p95 latency", Value: 184, Limit: 250, Unit: "ms", HigherIsWorse: true},
		{Name: "Cache hit rate", Value: 92, Limit: 100, Unit: "%", HigherIsWorse: false},
		{Name: "Goroutines", Value: 61, Unit: "", HigherIsWorse: true},
	}
}

func spark(from, to float64, n int) []float64 {
	out := make([]float64, n)
	for i := range out {
		out[i] = from + (to-from)*float64(i)/float64(n-1)
	}
	return out
}

// ── the dashboard ────────────────────────────────────────────────────────

// dashboardSection is what a person looks at first: what is wrong, and what
// they can do about it. An acknowledged alert stays, greyed, because "we know
// about it" and "it is gone" are different facts.
func alertSection(c *ui.Context) {
	showcase.Section(c, "告警与状态 · IncidentCard / AlertList / ServiceStatus")
	ui.Row(c).FillWidth().Gap(unit(c, 4)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).WidthPercent(58).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "AlertList — 确认过的还在，只是灰了")
			ack := showcase.State(c, "devtools.567.ack", "")
			dismissed := showcase.State(c, "devtools.567.dismissed", "")
			devtools.AlertList(c, devtools.AlertListOptions{
				Acknowledged: ack, Dismissed: dismissed,
				Alerts: []devtools.Alert{
					{ID: "a-1", Title: "Queue depth above 80 for 10 minutes",
						Detail: "Two machines are retrying the same callback.", Since: "08:52",
						Severity: core.Warning},
					{ID: "a-2", Title: "Webhook delivery failing for eu-west",
						Detail: "4,201 callbacks are parked and waiting for a human.",
						Since:  "08:40", Severity: core.Danger},
					{ID: "a-3", Title: "Disk at 91% on db-3",
						Detail: "Acknowledged at 07:15; the log rotation is being looked at.",
						Since:  "06:58", Severity: core.Warning, Acknowledged: true},
				},
			})
			resultLine(c, "Acknowledged() · Dismissed()",
				quoteOrEmpty(*ack)+" · "+quoteOrEmpty(*dismissed))
			showcase.Field(c, "AlertList — 什么都没有的时候是绿的")
			devtools.AlertList(c, devtools.AlertListOptions{
				Acknowledged: ack, Dismissed: dismissed,
				Alerts: []devtools.Alert{},
			})
		})

		ui.Column(c).WidthPercent(40).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "IncidentCard — 结束时间是调用方给的，所以不自己算")
			devtools.IncidentCard(c, devtools.IncidentCardOptions{
				Open: func(id string) {},
				Incident: devtools.Incident{
					ID: "i-91", Title: "Webhook delivery down in eu-west",
					Impact:  "Every callback from 6 European clinics was queued, not sent.",
					Started: "08:40", Ended: "09:12", Severity: core.Danger,
					Services: []string{"webhooks", "eu-west", "callbacks"},
				},
			})
			devtools.IncidentCard(c, devtools.IncidentCardOptions{
				Incident: devtools.Incident{
					ID: "i-92", Title: "Callback queue slowly draining",
					Impact:  "The backlog is falling again; nothing is being dropped.",
					Started: "09:20", Ongoing: true, Severity: core.Warning,
					Services: []string{"queue"},
				},
			})
		})
	})
}

// ── the request ──────────────────────────────────────────────────────────

// builderSection is the other half of a debugging session: putting a request
// together by hand, and the two controls that sit above every dashboard.
func requestSection(c *ui.Context) {
	showcase.Section(c, "构造请求 · APIRequestBuilder / KeyValueInput / DashboardFilterBar / RefreshIntervalSelector")
	ui.Row(c).FillWidth().Gap(unit(c, 4)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).WidthPercent(54).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "APIRequestBuilder — 头部、查询、正文；发不出去时说为什么")
			builder := devtools.APIRequestBuilder(c, devtools.APIRequestBuilderOptions{
				Method: &method, URL: &url, Headers: &headers,
				Query: &devtools_query, Body: &body, Sent: &sent,
			})
			resultLine(c, "APIRequestBuilder.Sent()", fmtBool(builder.Sent()))
			showcase.Field(c, "同一个构造器，地址填了就能发")
			devtools.APIRequestBuilder(c, devtools.APIRequestBuilderOptions{
				Method: &method, URL: &ok, Sent: new(bool), Sendable: true,
			})
		})

		ui.Column(c).WidthPercent(44).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "KeyValueInput — 藏起来的值还在请求里")
			newKey := showcase.State(c, "devtools.636.newKey", "")
			newValue := showcase.State(c, "devtools.636.newValue", "")
			kv := devtools.KeyValueInput(c, devtools.KeyValueInputOptions{
				NewKey: newKey, NewValue: newValue,
				Values: &values, KeyLabel: "Header name", ValueLabel: "Header value",
				Secrets:     devtools.DefaultSecretKeys(),
				Placeholder: "X-Request-Id",
			})
			resultLine(c, "KeyValueInput.Removed()", quoteOrEmpty(kv.Removed()))

			showcase.Field(c, "DashboardFilterBar — 每个筛选直接写回调用方的变量")
			rangeSel := showcase.State(c, "devtools.646.rangeSel", "received")
			querySel := showcase.State(c, "devtools.646.querySel", "eu-west")
			search := showcase.State(c, "devtools.646.search", "2871")
			placed := showcase.State(c, "devtools.646.placed", 20)
			devtools.DashboardFilterBar(c, devtools.DashboardFilterBarOptions{
				Searched: func() {},
				Filters: []devtools.DashboardFilterOptions{
					{Group: "Environment", Selected: querySel,
						Options: []string{"eu-west", "eu-central", "staging"}},
					{Group: "State", Selected: rangeSel,
						Options: []string{"received", "queued", "delivered"}},
					{Group: "Limit", Placed: placed, Min: 5, Max: 100},
					{Search: search, SearchLabel: "Callback id"},
				},
			})

			showcase.Field(c, "RefreshIntervalSelector — 关掉刷新也是一个答案")
			devtools.RefreshIntervalSelector(c, devtools.RefreshIntervalOptions{
				Seconds: &seconds,
			})
			devtools.RefreshIntervalSelector(c, devtools.RefreshIntervalOptions{
				Seconds: &off,
			})
		})
	})
}

// ── the functions ────────────────────────────────────────────────────────

// devtoolsPureSection is the part of the package with nothing a person
// presses: the rules the components above are made of.
func devtoolsPureSection(c *ui.Context) {
	showcase.Section(c, "纯函数 · 只调不测的那一半")
	value, _ := devtools.ParseJSON(responseBody())
	lines := logLines()
	spans := traceSpans()

	ui.Row(c).FillWidth().Gap(unit(c, 6)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).WidthPercent(50).Shrink(0).Gap(unit(c, 1.5)).Children(func() {
			specimen(c, `ParseJSON(body)`, devtools.KindOf(value).String())
			specimen(c, `KindOf(value)`, devtools.KindOf(value).String())
			specimen(c, `FormatJSON(src)`, firstLine(devtools.FormatJSON(`{"a":1,"b":[2,3]}`)))
			specimen(c, `MinifyJSON(src)`, devtools.MinifyJSON("{\n  \"a\": 1\n}"))
			specimen(c, `MarshalIndent(value)`, firstLine(mustIndent(value)))
			specimen(c, `Flatten(value)`, itoaOf(len(devtools.Flatten(value)))+" rows")
			specimen(c, `JSONPreview(value, 40)`,
				devtools.JSONPreview(value, 40))
			specimen(c, `Redact(value, DefaultSecretKeys())`,
				devtools.JSONPreview(
					redactAt(value, "authorization"), 40))
			specimen(c, `RedactedPaths(value, DefaultSecretKeys())`,
				joinAll(devtools.RedactedPaths(value, devtools.DefaultSecretKeys())))
			specimen(c, `DefaultSecretKeys()`,
				itoaOf(len(devtools.DefaultSecretKeys()))+" names")
			specimen(c, `FilterNodes(nodes, "cb_2871")`,
				itoaOf(len(devtools.FilterNodes(devtools.Flatten(value), "cb_2871")))+
					" rows")
		})

		ui.Column(c).WidthPercent(50).Shrink(0).Gap(unit(c, 1.5)).Children(func() {
			specimen(c, `FilterLogs(lines, Warn, "db", 2)`,
				itoaOf(len(devtools.FilterLogs(lines, devtools.Warn, "db", 2)))+
					" lines, newest kept")
			specimen(c, `ParseLogLevel("warning")`,
				devtools.ParseLogLevel("warning").String())
			specimen(c, `RoundDuration(340*time.Millisecond)`,
				devtools.RoundDuration(340*time.Millisecond))
			specimen(c, `RoundDuration(90*time.Second)`,
				devtools.RoundDuration(90*time.Second))
			specimen(c, `TotalOf(spans)`, devtools.RoundDuration(devtools.TotalOf(spans)))
			specimen(c, `SlowestSpans(spans, 3)`,
				joinSpanNames(devtools.SlowestSpans(spans, 3)))
			specimen(c, `DepthOf(spans, 3)`, itoaOf(devtools.DepthOf(spans, 3)))
			specimen(c, `SpanBar(spans[3], TotalOf(spans))`,
				spanBarText(spans[3], devtools.TotalOf(spans)))
			specimen(c, `UptimeSlots(probes)`, itoaOf(len(
				devtools.UptimeSlots(probes("api", 0.98))))+" days")
			specimen(c, `UptimePercent(slots)`, fmtFloat(
				devtools.UptimePercent(devtools.UptimeSlots(probes("api", 0.98)))))
			specimen(c, `UptimeTone(0.994)`, devtools.UptimeTone(0.994).String())
			// StatusTone is the exported face of the package's own status
			// parser, so what is on the page is the tone it returns for a
			// status line with a reason on the end of it.
			specimen(c, `StatusTone("404 Not Found")`,
				devtools.StatusTone("404 Not Found").String())
			specimen(c, `ServiceStateTone("degraded")`,
				devtools.ServiceStateTone(devtools.StateDegraded).String())
			specimen(c, `MetricTone(m)`, devtools.MetricTone(metrics()[2]).String())
		})
	})
}

// ── the data ─────────────────────────────────────────────────────────────

// logPill is one level as a row of log lines draws it: the level's own word
// in the tone its severity pair gives.
func logPill(c *ui.Context, l devtools.LogLevel) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	_, fg := l.Severity().Pair(k)
	ui.Box(c).Padding(u*0.5, u*1.5).Radius(theme.PillRadius).
		Background(k.Surface).Label(l.String()).Children(func() {
		ui.Text(c, l.String()).TextColor(fg).
			Font(monoFamily).FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
	})
}

// statusPill is a status line in the tone StatusTone gives it, which is the
// same answer a response viewer and a service page both give.
func statusPill(c *ui.Context, line string) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	sev := devtools.StatusTone(line)
	bg, fg := sev.Pair(k)
	ui.Box(c).Padding(u*0.75, u*2).Radius(theme.PillRadius).Background(bg).
		Label(line).Children(func() {
		ui.Text(c, line).TextColor(fg).Font(monoFamily).Bold().Shrink(0).
			FontSize(core.FontSize(c, theme.CaptionSize))
	})
}

// ── small pieces ─────────────────────────────────────────────────────────

func mustIndent(v any) string {
	out, err := devtools.MarshalIndent(v)
	if err != nil {
		return "error: " + err.Error()
	}
	return out
}

func firstLine(s string) string {
	for i, r := range s {
		if r == '\n' {
			return s[:i] + " …"
		}
	}
	return s
}

// redactAt is one branch of a redacted value, for the table: the value under
// a secret key after Redact has been over it.
func redactAt(v any, key string) any {
	for _, n := range devtools.Flatten(v) {
		if n.Path == key {
			return n.Value
		}
	}
	return nil
}

func joinAll(xs []string) string {
	out := "["
	for i, x := range xs {
		if i > 0 {
			out += " "
		}
		out += x
	}
	return out + "]"
}

func joinSpanNames(spans []devtools.Span) string {
	out := "["
	for i, s := range spans {
		if i > 0 {
			out += " "
		}
		out += s.Name
	}
	return out + "]"
}

// spanBarText is a span's place in a trace as the two numbers it is: the
// left edge and the width, both as percentages of the whole.
func spanBarText(s devtools.Span, total time.Duration) string {
	left, width := devtools.SpanBar(s, total)
	return "left " + fmtFloat(left) + "% · width " + fmtFloat(width) + "%"
}

var _ = input.Button
