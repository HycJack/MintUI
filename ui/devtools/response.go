package devtools

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/theme"
)

// ResponseViewerOptions configure a ResponseViewer.
type ResponseViewerOptions struct {
	// Status is the response's status code and Reason its status line, both
	// already the caller's strings. They are drawn in the tone StatusTone
	// gives rather than in a colour the caller picked, so that "did this
	// work" reads the same on every response in a log.
	Status, Reason string
	// Method and URL are what was asked for.
	Method, URL string
	// Headers are the response's headers, already a map. They are sorted by
	// name for the same reason a JSON object's keys are: a header map has no
	// order of its own, and one that reshuffled itself on every redraw could
	// not be compared with another.
	Headers map[string]string
	// Body is the response body as it came off the wire, before parsing. It
	// is a string and not a parsed value because a body that is not JSON is
	// the case this component most needs to show properly.
	Body string
	// Duration is how long it took, already formatted.
	Duration string
	// Size is how big the body was, already formatted.
	Size string
	// Secrets are the header and body keys whose values are hidden. Nil
	// hides DefaultSecretKeys; an empty non-nil slice hides nothing, which is
	// the only way to say "this response has no secrets in it" on purpose.
	Secrets []string
	// Tab is which part is showing — "headers" or "body" — and Tabbed what
	// was pressed. Two pointers for one fact is awkward, so the choice is
	// made here rather than by the caller deciding what to draw.
	Tab    *string
	Tabbed *string
	// Height is the panel's height; zero lets the layout give it one.
	Height float32
}

// ResponseViewerResult carries a ResponseViewer and what was pressed in it.
type ResponseViewerResult struct {
	// Element is the whole viewer.
	Element *ui.Element
	// redacted is how many values were hidden, which is worth saying: a
	// viewer that silently redacted nothing and one that silently redacted
	// forty look identical from the outside.
	redacted int
}

// Redacted reports how many values were hidden from this response.
func (r ResponseViewerResult) Redacted() int { return r.redacted }

// ResponseViewer is a response: its line, its headers and its body, with the
// body as a tree when it is JSON and as text when it is not.
//
// Redaction happens before anything is drawn and cannot be turned off by
// accident: Secrets nil means the default set, and an empty slice — the only
// way to ask for none — is a deliberate thing to pass. A token in a
// screenshot of a bug report is the failure that costs the most and is the
// easiest to prevent, so the safe reading is the default one.
//
// The status line is coloured by StatusTone, which is core's answer for
// "did this work" rather than this package's, so that a response in a log
// and a status on a service's own page agree.
func ResponseViewer(c *ui.Context, opts ResponseViewerOptions) ResponseViewerResult {
	if opts.Tab == nil {
		panic("devtools: ResponseViewer needs the *string Tab it shows")
	}
	if opts.Tabbed == nil {
		panic("devtools: ResponseViewer needs the *string Tabbed writes to")
	}
	u := core.Density(c).Unit()

	secrets := opts.Secrets
	if secrets == nil {
		secrets = DefaultSecretKeys()
	}

	var r ResponseViewerResult
	viewer := ui.Column(c).FillWidth().Gap(u * 2).FillHeight().
		Label(core.Msg(c, "devtools.response", "Response"))
	viewer.Children(func() {
		responseLine(c, opts)

		// Two tabs, and the body is one of them: a response with four
		// hundred headers is not readable if they are stacked above the body.
		ui.Row(c).FillWidth().Gap(u * 2).Children(func() {
			for _, name := range []string{"headers", "body"} {
				label := core.Msg(c, "devtools."+name, strings.Title(name))
				btn := input.Button(c, label, input.ButtonOptions{})
				if btn.Clicked() {
					*opts.Tabbed = name
					*opts.Tab = name
				}
			}
		})

		if *opts.Tab == "headers" {
			responseHeaders(c, opts, secrets, &r)
			return
		}
		responseBody(c, opts, secrets, &r)
	})
	r.Element = viewer
	return r
}

// responseLine is the status, the method, the URL, the size and the duration,
// in one row above everything else.
func responseLine(c *ui.Context, opts ResponseViewerOptions) {
	k, u := core.Tokens(c), core.Density(c).Unit()

	ui.Row(c).FillWidth().Gap(u * 2).AlignItems(ui.Center).Children(func() {
		sev := StatusTone(opts.Status)
		bg, fg := sev.Pair(k)
		line := strings.TrimSpace(opts.Status + " " + opts.Reason)
		ui.Box(c).Padding(u*0.75, u*2).Radius(theme.PillRadius).Background(bg).
			Label(line).Children(func() {
			ui.Text(c, line).TextColor(fg).Font(monoFamily).
				FontSize(core.FontSize(c, theme.CaptionSize)).Bold().Shrink(0)
		})
		if opts.Method != "" {
			ui.Text(c, opts.Method).TextColor(k.TextMuted).Shrink(0).Font(monoFamily).
				FontSize(core.FontSize(c, theme.MetaSize)).SingleLine()
		}
		ui.Text(c, opts.URL).TextColor(k.Text).Grow(1).Font(monoFamily).
			FontSize(core.FontSize(c, theme.MetaSize)).SingleLine()
		if opts.Size != "" {
			ui.Text(c, opts.Size).TextColor(k.TextFaint).Shrink(0).
				FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
		}
		if opts.Duration != "" {
			ui.Text(c, opts.Duration).TextColor(k.TextFaint).Shrink(0).
				FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
		}
	})
}

// responseHeaders is the header list, sorted, with the secret ones hidden.
func responseHeaders(c *ui.Context, opts ResponseViewerOptions, secrets []string, r *ResponseViewerResult) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	names := make([]string, 0, len(opts.Headers))
	for name := range opts.Headers {
		names = append(names, name)
	}
	sort.Strings(names)

	if len(names) == 0 {
		ui.Text(c, core.Msg(c, "devtools.noHeaders", "No headers")).
			TextColor(k.TextFaint).FontSize(core.FontSize(c, theme.MetaSize)).SingleLine()
		return
	}
	hidden := secretSet(secrets)
	ui.Column(c).FillWidth().Gap(u * 0.5).Children(func() {
		for _, name := range names {
			value := opts.Headers[name]
			if hidden[strings.ToLower(name)] {
				value = redacted
				r.redacted++
			}
			ui.Row(c).FillWidth().Gap(u * 2).Label(name).Children(func() {
				ui.Text(c, name).TextColor(k.TextMuted).Shrink(0).Font(monoFamily).
					Width(headerNameWidth).FontSize(core.FontSize(c, theme.MetaSize)).SingleLine()
				ui.Text(c, value).TextColor(k.Text).Grow(1).Font(monoFamily).
					FontSize(core.FontSize(c, theme.MetaSize)).SingleLine()
			})
		}
	})
}

// headerNameWidth is how wide the name column is. HTTP header names top out
// around twenty-five characters, so this fits the longest one anybody uses
// without a hyphen being wrapped away from its value.
const headerNameWidth float32 = 150

// responseBody is the body as a tree when it is JSON and as text when it is
// not — with the secret values already replaced in both.
func responseBody(c *ui.Context, opts ResponseViewerOptions, secrets []string, r *ResponseViewerResult) {
	value, err := ParseJSON(opts.Body)
	if err != nil {
		// Not JSON is the ordinary case for a file, an HTML page and a
		// redirect's body. It is drawn as text rather than as an error.
		ui.Text(c, opts.Body).TextColor(core.Tokens(c).TextMuted).Font(monoFamily).
			FontSize(core.FontSize(c, theme.MetaSize)).Grow(1)
		return
	}
	clean := Redact(value, secrets)
	r.redacted += len(RedactedPaths(value, secrets))

	marks := map[string]bool{}
	for _, path := range RedactedPaths(value, secrets) {
		marks[path] = true
	}
	nodes := Flatten(clean)
	for i := range nodes {
		nodes[i].Redacted = marks[nodes[i].Path]
	}

	JsonTree(c, JsonTreeOptions{
		Nodes:   nodes,
		Toggled: new(string),
		Height:  opts.Height,
	})
}

// secretSet is a set of lower-cased names to hide.
func secretSet(keys []string) map[string]bool {
	out := make(map[string]bool, len(keys))
	for _, k := range keys {
		out[strings.ToLower(k)] = true
	}
	return out
}

// StatusTone is the severity an HTTP status is drawn at.
//
// The bands are the ones people act on: under 300 worked, under 400 is a
// question about the request, under 500 is the service's problem. Everything
// else — 201 against 200, 204 against 200 — is a distinction a developer
// makes in text and nobody needs from a colour.
func StatusTone(status string) core.Severity {
	code := statusCode(status)
	switch {
	case code == 0:
		return core.Neutral
	case code < 300:
		return core.Success
	case code < 400:
		return core.Accent
	case code < 500:
		return core.Warning
	default:
		return core.Danger
	}
}

// statusCode is the number at the front of a status line, or zero when there
// is not one. The line is "200 OK" as often as it is "200", and a viewer
// should not care which it was given.
func statusCode(status string) int {
	s := strings.TrimSpace(status)
	n := 0
	seen := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		n = n*10 + int(c-'0')
		seen = true
		if n > 999 {
			return 0
		}
	}
	if !seen {
		return 0
	}
	return n
}

// LogStreamOptions configure a LogStream.
type LogStreamOptions struct {
	// Lines are the log, oldest first as the caller has them. The stream
	// filters and takes the newest of what is left; it does not reorder the
	// caller's slice, because a caller with a ring buffer would find its
	// buffer shuffled every frame.
	Lines []LogLine
	// Level is the quietest level shown, and everything louder comes with
	// it. It is the caller's so that the filter survives the pane being
	// closed and reopened.
	Level LogLevel
	// Search is the text matched against the message and the source.
	Search string
	// Follow keeps the view at the end of the stream, for a log that is
	// still being written to. Off for a log somebody is reading back.
	Follow bool
	// Limit caps the lines drawn; zero is no cap.
	Limit int
	// Height is the viewport's height.
	Height float32
	// Empty draws instead of the lines when the filter leaves nothing.
	Empty func()
}

// LogStreamResult carries a LogStream and what was pressed in it.
type LogStreamResult struct {
	// Element is the whole stream.
	Element *ui.Element
	// shown is how many lines the filter produced.
	shown int
}

// Shown reports how many lines the filter let through, which is what a header
// says when it says "142 lines".
func (r LogStreamResult) Shown() int { return r.shown }

// LogStream is a log: lines in a level's colour, filtered by level and text,
// newest at the bottom.
//
// It is a list of rows rather than a ui.List because a log is not a
// navigable set: there is no chosen row, no keyboard order and nothing to
// reorder. Wrapping it in a list would give it a selection model it has no
// use for and a cursor somebody could move into a log they were reading.
//
// The level filter is "this level and louder" — see FilterLogs — and the
// limit keeps the newest, because during an incident the line that just
// arrived is the one being looked for.
func LogStream(c *ui.Context, opts LogStreamOptions) LogStreamResult {
	k, u := core.Tokens(c), core.Density(c).Unit()
	shown := FilterLogs(opts.Lines, opts.Level, opts.Search, opts.Limit)

	var r LogStreamResult
	r.shown = len(shown)
	r.Element = ui.Column(c).FillWidth().Gap(u).
		Label(core.Msg(c, "devtools.logs", "Log")).Children(func() {
		ui.Row(c).FillWidth().Gap(u * 2).AlignItems(ui.Center).Children(func() {
			ui.Text(c, core.Msg(c, "devtools.lineCount", "Lines")).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
			ui.Text(c, itoa(len(shown))).TextColor(k.Text).Grow(1).
				FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
			for _, lvl := range []LogLevel{Debug, Info, Warn, Error} {
				lvl := lvl
				if input.Button(c, lvl.String(), input.ButtonOptions{
					Primary: lvl == opts.Level,
				}).Clicked() {
					opts.Level = lvl
				}
			}
		})
		if len(shown) == 0 {
			if opts.Empty != nil {
				opts.Empty()
			} else {
				ui.Text(c, core.Msg(c, "devtools.noLogs", "No lines match")).
					TextColor(k.TextFaint).FontSize(core.FontSize(c, theme.MetaSize)).SingleLine()
			}
			return
		}
		for _, l := range shown {
			logRow(c, l)
		}
	})
	return r
}

// logRow is one line: its level, its time, where it came from and what it
// said.
func logRow(c *ui.Context, l LogLine) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	_, ink := l.Level.Severity().Pair(k)

	ui.Row(c).FillWidth().Gap(u*2).AlignItems(ui.Start).
		Padding(u*0.5, u*1.5).Radius(theme.SmallRadius).Background(k.Surface).
		Label(l.Message).Role(ui.RoleListItem).Children(func() {
		// The level as a word, not as a colour: the colour is on the level's
		// own foreground and a log line is text.
		ui.Text(c, l.Level.String()).TextColor(ink).Shrink(0).Font(monoFamily).
			FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
		ui.Text(c, l.At.Format("15:04:05")).TextColor(k.TextFaint).Shrink(0).
			Font(monoFamily).FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
		if l.Source != "" {
			ui.Text(c, l.Source).TextColor(k.TextMuted).Shrink(0).
				FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
		}
		ui.Text(c, l.Message).TextColor(k.Text).Grow(1).
			FontSize(core.FontSize(c, theme.RowSize))
	})
}

// Alert is one thing a viewer is showing about a system's state.
type Alert struct {
	// ID is the alert's own identity, for dismissing it.
	ID string
	// Title is the one line that says what is happening.
	Title string
	// Detail is the paragraph under it: what it affects and what to do.
	Detail string
	// Since is when it started, already formatted.
	Since string
	// Severity is the tone. Warning and Danger are the two that matter; a
	// whole list of informational alerts is a list nobody reads.
	Severity core.Severity
	// Acknowledged marks an alert somebody has seen. Acknowledged alerts stay
	// in the list, greyed, because "we know about it" and "it is gone" are
	// different states and only the second is a reason to hide a row.
	Acknowledged bool
}

// AlertListOptions configure an AlertList.
type AlertListOptions struct {
	// Alerts are the rows, in the caller's order — which should be newest
	// first, and which the list does not sort, because a caller's order is
	// usually one of severity and the list would be re-deciding it with less
	// information.
	Alerts []Alert
	// Acknowledged writes the id of the alert that was acknowledged.
	Acknowledged *string
	// Dismissed writes the id of the alert that was dismissed.
	Dismissed *string
	// Empty draws instead of the rows when there are none.
	Empty func()
}

// AlertListResult carries an AlertList and what was done in it.
type AlertListResult struct {
	// Element is the whole list.
	Element *ui.Element
}

// AlertList is what is currently wrong, each row with the two things that can
// be done about it.
//
// Acknowledged and dismissed are separate, and both keep the row. An alert
// somebody has acknowledged is still true; hiding it would make the list say
// the system is fine when it is merely known about, which is the difference
// between "nothing is wrong" and "nothing is wrong that we have not looked
// at".
func AlertList(c *ui.Context, opts AlertListOptions) AlertListResult {
	if opts.Acknowledged == nil {
		panic("devtools: AlertList needs the *string Acknowledged writes to")
	}
	if opts.Dismissed == nil {
		panic("devtools: AlertList needs the *string Dismissed writes to")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	var r AlertListResult
	r.Element = ui.Column(c).FillWidth().Gap(u * 2).
		Label(core.Msg(c, "devtools.alerts", "Alerts")).Children(func() {
		if len(opts.Alerts) == 0 {
			if opts.Empty != nil {
				opts.Empty()
				return
			}
			ui.Text(c, core.Msg(c, "devtools.noAlerts", "Nothing is wrong")).
				TextColor(k.Success).FontSize(core.FontSize(c, theme.RowSize)).SingleLine()
			return
		}
		for _, a := range opts.Alerts {
			alertRow(c, opts, a)
		}
	})
	return r
}

// alertRow is one alert.
func alertRow(c *ui.Context, opts AlertListOptions, a Alert) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	bg, _ := a.Severity.Pair(k)
	ink := k.Text
	if a.Acknowledged {
		ink = k.TextFaint
	}

	ui.Column(c).FillWidth().Gap(u).Padding(u*2, u*2.5).Radius(theme.ControlRadius).
		Background(bg).Label(a.Title).Role(ui.RoleGroup).Children(func() {
		ui.Row(c).FillWidth().Gap(u * 2).AlignItems(ui.Center).Children(func() {
			ui.Text(c, a.Title).TextColor(ink).Grow(1).Bold().
				FontSize(core.FontSize(c, theme.RowSize)).SingleLine()
			if a.Since != "" {
				ui.Text(c, a.Since).TextColor(k.TextFaint).Shrink(0).
					FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
			}
		})
		if a.Detail != "" {
			ui.Text(c, a.Detail).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.MetaSize))
		}
		ui.Row(c).FillWidth().Gap(u * 2).Justify(ui.End).Children(func() {
			if input.Button(c, core.Msg(c, "devtools.acknowledge", "Acknowledge"),
				input.ButtonOptions{}).Clicked() {
				*opts.Acknowledged = a.ID
			}
			if input.Button(c, core.Msg(c, "devtools.dismiss", "Dismiss"),
				input.ButtonOptions{Danger: true}).Clicked() {
				*opts.Dismissed = a.ID
			}
		})
	})
}

// IncidentCardOptions configure an IncidentCard.
type IncidentCardOptions struct {
	// Incident is what happened.
	Incident Incident
	// Open is pressed when the card is; it is handed the incident's id. It is
	// nil for a card that is only being read, and the press is then refused
	// in this package's own code rather than left to a greyed button.
	Open func(id string)
}

// IncidentCardResult carries an IncidentCard and what was pressed in it.
type IncidentCardResult struct {
	// Element is the whole card.
	Element *ui.Element
}

// Incident is one outage or degradation.
type Incident struct {
	// ID is what it is addressed by.
	ID string
	// Title is what happened, in the words the caller's status page uses.
	Title string
	// Impact is who it affected and how badly.
	Impact string
	// Started is when it began and Ended when it finished, as text. An empty
	// Ended means it is still going, which is drawn differently.
	Started, Ended string
	// Severity is the tone. It is the caller's, because an incident page
	// usually has its own ladder — SEV-1 through SEV-4, or degraded versus
	// down — and mapping it here would put this package in the middle of a
	// vocabulary that is not its.
	Severity core.Severity
	// Ongoing marks an incident that has not finished, which is redundant
	// with an empty Ended and kept because "no end yet" is a fact about the
	// row and "the end string happens to be blank" is a fact about a string.
	Ongoing bool
	// Services are the names of what was affected.
	Services []string
}

// IncidentCard is one incident: its title, its impact, how long it lasted and
// what it touched.
//
// The duration is drawn from the two formatted strings it was given rather
// than worked out, because "Started" and "Ended" are the caller's own formats
// and subtracting them would mean knowing those formats. An incident with no
// end says "ongoing" instead, which is the same fact and reads better than a
// dash.
func IncidentCard(c *ui.Context, opts IncidentCardOptions) IncidentCardResult {
	k, u := core.Tokens(c), core.Density(c).Unit()
	in := opts.Incident

	card := ui.Column(c).FillWidth().Gap(u).Padding(u*2.5, u*3).
		Radius(theme.CardRadius).Background(k.Background).
		BorderWidth(theme.BorderWidth).BorderColor(k.Border).
		Cursor(ui.CursorPointer).Label(in.Title).Role(ui.RoleButton)
	if card.Clicked() && opts.Open != nil {
		opts.Open(in.ID)
	}
	card.Children(func() {
		ui.Row(c).FillWidth().Gap(u * 2).AlignItems(ui.Center).Children(func() {
			sev := in.Severity
			bg, fg := sev.Pair(k)
			ui.Box(c).Padding(u*0.75, u*2).Radius(theme.PillRadius).Background(bg).
				Label(sev.String()).Children(func() {
				ui.Text(c, sev.String()).TextColor(fg).Bold().Shrink(0).
					FontSize(core.FontSize(c, theme.CaptionSize))
			})
			ui.Text(c, in.Title).TextColor(k.Text).Grow(1).Bold().
				FontSize(core.FontSize(c, theme.BodySize)).SingleLine()
			ui.Text(c, durationOf(c, in)).TextColor(k.TextMuted).Shrink(0).
				FontSize(core.FontSize(c, theme.MetaSize)).SingleLine()
		})
		if in.Impact != "" {
			ui.Text(c, in.Impact).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.RowSize))
		}
		if len(in.Services) > 0 {
			ui.Row(c).FillWidth().Gap(u).Wrap().Children(func() {
				for _, s := range in.Services {
					ui.Text(c, s).TextColor(k.TextFaint).Shrink(0).
						FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
				}
			})
		}
	})
	var r IncidentCardResult
	r.Element = card
	return r
}

// durationOf is how long an incident lasted, from the two formatted strings
// it was given. There is no arithmetic here on purpose: Started and Ended
// are the caller's own formats, and taking them apart to subtract would mean
// knowing their format. An incident with no end says "ongoing" instead.
func durationOf(c *ui.Context, in Incident) string {
	if in.Ongoing || in.Ended == "" {
		return core.Msg(c, "devtools.ongoing", "ongoing")
	}
	return in.Started + " – " + in.Ended
}

// MarshalIndent writes a parsed value as text with one key per line and two
// spaces of indent. It is the writer behind FormatJSON, exported so a caller
// that already has a tree can write it without going back through text.
func MarshalIndent(v any) (string, error) {
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// marshalCompact writes a parsed value with no whitespace at all.
func marshalCompact(v any) (string, error) {
	out, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(out), nil
}
