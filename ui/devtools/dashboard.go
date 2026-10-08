package devtools

import (
	"strconv"
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/theme"
)

// StatCardOptions configure a StatCard.
type StatCardOptions struct {
	// Label is what the figure is of, and Title is the figure itself. They
	// are the caller's own strings and not formatted here: a figure this
	// package formatted would have opinions about units and separators that
	// belong to whoever owns the number.
	Label, Value string
	// Delta is how the figure has moved — "+12%", "−340ms" — and Direction
	// says which way is good. It is a string and a bool rather than a signed
	// number because "up" and "down" are not the same for every metric: more
	// requests is good and more errors is not, and a card that coloured its
	// arrow by the sign alone would be wrong about half of them.
	Delta     string
	Direction int
	// Severity is the tone of the card itself, for a figure that is a
	// headline — an error rate above a threshold, an outage count. The
	// zero value is Neutral, which is right for almost every card.
	Severity core.Severity
	// Spark is the recent history as a row of fractions 0 to 1, oldest
	// first. It is a slice of fractions rather than a rendered chart because
	// the card is three lines tall and a chart wants four hundred.
	Spark []float64
	// Width bounds the card; zero shares what the row leaves.
	Width float32
}

// Direction of a change, which is the metric's own and not the sign's.
const (
	// DeltaNone is a figure with no change to report.
	DeltaNone = iota
	// DeltaUp is a change that went up.
	DeltaUp
	// DeltaDown is a change that went down.
	DeltaDown
)

// StatCardResult carries a StatCard.
type StatCardResult struct {
	// Element is the card.
	Element *ui.Element
}

// StatCard is one figure: what it is, how big it is, and how it has moved.
//
// The change is coloured by Direction and not by whether the number in Delta
// has a plus in it. "−340ms" on a latency card is good news and "+340ms" is
// bad, and the same two strings mean the opposite on a card counting
// requests — which is why the direction is asked for rather than read.
//
// The history, when there is one, is a row of vertical bars drawn inside the
// card rather than ui/chart's Sparkline: a sparkline in a card that is three
// lines tall would want its own gutter and its own labels, and what a card
// wants is the shape of the last hour with no numbers on it at all.
func StatCard(c *ui.Context, opts StatCardOptions) StatCardResult {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if opts.Label == "" {
		panic("devtools: StatCard needs a Label; a figure with no name on it " +
			"says nothing about what it is a figure of")
	}

	card := ui.Column(c).FillWidth().Gap(u*0.5).Padding(u*2.5, u*3).
		Radius(theme.CardRadius).Background(k.Surface).Label(opts.Label)
	if opts.Width > 0 {
		card.Width(opts.Width).Shrink(0)
	}
	card.Children(func() {
		ui.Text(c, opts.Label).TextColor(k.TextMuted).
			FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
		ui.Text(c, opts.Value).TextColor(k.Text).Bold().
			FontSize(core.FontSize(c, theme.TitleSize)).SingleLine()
		if opts.Delta != "" || opts.Severity != core.Neutral {
			ui.Row(c).AlignItems(ui.Center).Gap(u).Children(func() {
				if opts.Delta != "" {
					ui.Text(c, opts.Delta).TextColor(deltaInk(k, opts.Direction)).
						FontSize(core.FontSize(c, theme.MetaSize)).SingleLine()
				}
				if opts.Severity != core.Neutral {
					bg, fg := opts.Severity.Pair(k)
					ui.Box(c).Padding(u*0.75, u*1.25).Radius(theme.PillRadius).
						Background(bg).Label(opts.Severity.String()).Children(func() {
						ui.Text(c, opts.Severity.String()).TextColor(fg).Bold().Shrink(0).
							FontSize(core.FontSize(c, theme.CaptionSize))
					})
				}
			})
		}
		if len(opts.Spark) > 0 {
			sparkRow(c, opts.Spark)
		}
	})
	var r StatCardResult
	r.Element = card
	return r
}

// sparkRow is the little history under a figure: one vertical bar per
// reading, all sharing one height so that the tallest fills it.
func sparkRow(c *ui.Context, values []float64) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	peak := 0.0
	for _, v := range values {
		if v > peak {
			peak = v
		}
	}
	// An all-zero history has no scale, and drawing bars at a third of the
	// height because they are all zero would be a claim about a trend.
	filled := peak > 0

	ui.Row(c).FillWidth().AlignItems(ui.End).Gap(u * 0.35).Height(sparkHeight).
		Label(core.Msg(c, "devtools.history", "Recent history")).Children(func() {
		for _, v := range values {
			h := sparkHeight * 0.15
			if filled && v > 0 {
				h = sparkHeight * float32(v/peak)
				if h < 1 {
					h = 1
				}
			}
			ink := k.Accent
			if filled && v == 0 {
				ink = k.Surface
			}
			ui.Box(c).Grow(1).FillHeight().Shrink(0).Radius(h / 2).Background(ink)
		}
	})
}

// sparkHeight is how tall a card's history is: three lines of a card, no
// more. It has to be enough for the shape to be readable and small enough
// that the figure above it stays the figure.
const sparkHeight float32 = 22

// deltaInk is a change's colour, from the direction and not the sign.
func deltaInk(k theme.Tokens, dir int) ui.Color {
	switch dir {
	case DeltaUp:
		return k.Success
	case DeltaDown:
		return k.Danger
	}
	return k.TextMuted
}

// APIRequestBuilderOptions configure an APIRequestBuilder.
type APIRequestBuilderOptions struct {
	// Method is the caller's HTTP verb, written as it changes, and URL the
	// address. Both are plain strings rather than a parsed request, because
	// a builder that refused a half-typed URL would be no use to somebody
	// building one.
	Method, URL *string
	// Query, Headers and Body are the caller's strings for the three parts of
	// a request. Keys and values for headers and query are the caller's own
	// KeyValues rather than a string, because splitting a string back into
	// pairs is how a value containing an equals sign gets broken.
	Query, Body *string
	Headers     *KeyValues
	// Sent writes true when the request is sent.
	Sent *bool
	// Methods are the verbs offered; empty offers the five that cover
	// essentially every request anybody makes by hand.
	Methods []string
	// Sendable says whether the request is complete enough to send. It is
	// the caller's because what counts as complete is theirs — this package
	// has no opinion about whether a body is required for a DELETE.
	Sendable bool
}

// defaultMethods is the verb set a builder offers when the caller does not
// name its own. POST, GET, PUT, PATCH and DELETE are the five; anything else
// is rare enough that a caller who needs it can pass the list.
func defaultMethods() []string {
	return []string{"GET", "POST", "PUT", "PATCH", "DELETE"}
}

// APIRequestBuilderResult carries an APIRequestBuilder and what was sent.
type APIRequestBuilderResult struct {
	// Element is the whole builder.
	Element *ui.Element
	// sent reports a press of the send button.
	sent bool
}

// Sent reports that the request was sent this frame.
func (r APIRequestBuilderResult) Sent() bool { return r.sent }

// APIRequestBuilder is a request put together by hand: the verb, the address,
// the query, the headers and the body, with the pieces that assemble them
// reusable.
//
// It does not send anything itself. A component that could fire a request
// would be a component with a network stack behind it, and the caller — who
// knows what to do with a failure — is better placed to ask. Sent reports the
// press and the assembled request is the caller's to read.
//
// The pieces are shared with KeyValueInput rather than drawn here, because a
// header row and a query row are the same control and a builder that drew
// its own would be the second place where "adding a pair" was worked out.
func APIRequestBuilder(c *ui.Context, opts APIRequestBuilderOptions) APIRequestBuilderResult {
	if opts.Method == nil || opts.URL == nil {
		panic("devtools: APIRequestBuilder needs the *string Method and URL to " +
			"point at; it keeps no request of its own")
	}
	if opts.Sent == nil {
		panic("devtools: APIRequestBuilder needs the *bool Sent writes to")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	methods := opts.Methods
	if len(methods) == 0 {
		methods = defaultMethods()
	}

	var newHeaderKey, newHeaderValue string

	var r APIRequestBuilderResult
	r.Element = ui.Column(c).FillWidth().Gap(u * 2).
		Label(core.Msg(c, "devtools.builder", "Request")).Children(func() {
		// The verb and the address on one line: they are one thing to read,
		// and two stacked fields would put the address a row below the word
		// that says what kind of address it is.
		ui.Row(c).FillWidth().Gap(u * 2).AlignItems(ui.End).Children(func() {
			// The verb in a box of its own, because a select fills the row it
			// is in: left as it was, the address beside it was squeezed to
			// nothing, and the address is the one field a builder cannot do
			// without.
			chosen := *opts.Method
			ui.Box(c).Shrink(0).Children(func() {
				input.Select(c, &chosen, choicesOf(methods), input.SelectOptions{
					Label:       core.Msg(c, "devtools.method", "Method"),
					Placeholder: core.Msg(c, "devtools.method", "Method"),
				})
			})
			*opts.Method = chosen

			ui.Box(c).Grow(1).Shrink(0).Children(func() {
				input.TextInput(c, opts.URL, input.TextInputOptions{
					Label:       core.Msg(c, "devtools.url", "URL"),
					Placeholder: "https://api.example.com/v1/callbacks",
				})
			})
		})

		if opts.Headers != nil {
			KeyValueInput(c, KeyValueInputOptions{
				Values:     opts.Headers,
				NewKey:     &newHeaderKey,
				NewValue:   &newHeaderValue,
				KeyLabel:   core.Msg(c, "devtools.headers", "Header name"),
				ValueLabel: core.Msg(c, "devtools.headerValue", "Header value"),
				Secrets:    DefaultSecretKeys(),
			})
		}
		if opts.Query != nil {
			input.TextArea(c, opts.Query, input.TextAreaOptions{
				Label: core.Msg(c, "devtools.query", "Query"),
				Lines: 3,
			})
		}
		if opts.Body != nil {
			input.TextArea(c, opts.Body, input.TextAreaOptions{
				Label:       core.Msg(c, "devtools.body", "Body"),
				Placeholder: `{ "callback_id": "cb_2871" }`,
				Lines:       6,
			})
		}

		ui.Row(c).FillWidth().Gap(u * 2).AlignItems(ui.Center).Children(func() {
			// A request that cannot be sent says why rather than offering a
			// button that does nothing: a disabled button with no words
			// beside it is the least helpful control there is.
			if !opts.Sendable {
				ui.Text(c, core.Msg(c, "devtools.incomplete", "Fill in a URL to send")).
					TextColor(k.TextFaint).FontSize(core.FontSize(c, theme.MetaSize)).SingleLine()
			}
			ui.Box(c).Grow(1)
			send := input.Button(c, core.Msg(c, "devtools.send", "Send"),
				input.ButtonOptions{Primary: true, Disabled: !opts.Sendable})
			if send.Clicked() && opts.Sendable {
				r.sent = true
				*opts.Sent = true
			}
		})
	})
	return r
}

// choicesOf turns a list of names into the choices a select takes.
func choicesOf(names []string) []input.Choice {
	out := make([]input.Choice, 0, len(names))
	for _, n := range names {
		out = append(out, input.Choice{Value: n, Label: n})
	}
	return out
}

// DashboardFilterOptions configure a DashboardFilterBar.
type DashboardFilterOptions struct {
	// Group is the filter's name, drawn at the head of the group so that a
	// row of them is a row of named filters rather than a row of controls.
	Group string
	// Selected is the chosen value, and Selected2.. are for a second and
	// third control in the same group. They are separate pointers rather
	// than a slice because each one is a different kind of value — a range,
	// a moment, a set — and a []string would make the caller convert at
	// every site.
	Selected *string
	// Options are the choices offered; empty offers none and the group is
	// not drawn at all, which is what a filter that has no options yet
	// should look like.
	Options []string
	// Placed is the caller's choice among the positions, and Placed2 is for a
	// second control in the same group.
	Placed  *int
	Placed2 *int
	// Min and Max bound a range control, which is drawn only when both are
	// given.
	Min, Max int
	// Search is the caller's query, drawn beside the filters when non-nil.
	Search *string
	// SearchLabel names that field; empty takes the library's own.
	SearchLabel string
}

// DashboardFilterBarOptions configure a DashboardFilterBar.
type DashboardFilterBarOptions struct {
	// Filters are the groups, left to right. A group with no options draws
	// nothing at all — not an empty box — because a filter nobody can set
	// is furniture.
	Filters []DashboardFilterOptions
	// Searched asks for the search to be submitted, and is the only way the
	// bar takes an action: pressing a filter changes the caller's state, and
	// pressing enter on the search asks the caller to go and apply it.
	Searched func()
}

// DashboardFilterBarResult carries a DashboardFilterBar and what was pressed.
type DashboardFilterBarResult struct {
	// Element is the whole bar.
	Element *ui.Element
}

// DashboardFilterBar is the row of controls above a dashboard: what it is
// showing, for how long, and a search.
//
// Every control writes straight into the caller's own pointer. A dashboard
// that kept a copy of "which range is chosen" would be a second answer to a
// question the URL is already answering, and the two would disagree the
// moment somebody shared a link.
func DashboardFilterBar(c *ui.Context, opts DashboardFilterBarOptions) DashboardFilterBarResult {
	k, u := core.Tokens(c), core.Density(c).Unit()

	var r DashboardFilterBarResult
	r.Element = ui.Column(c).FillWidth().Gap(u * 2).
		Label(core.Msg(c, "devtools.filters", "Filters")).Children(func() {
		if len(opts.Filters) > 0 {
			ui.Row(c).FillWidth().Gap(u * 3).AlignItems(ui.End).Wrap().Children(func() {
				for _, f := range opts.Filters {
					if len(f.Options) > 0 {
						filterSelect(c, f)
					}
				}
			})
		}
		for _, f := range opts.Filters {
			if f.Min > 0 || f.Max > 0 {
				filterRange(c, f)
			}
		}
		for _, f := range opts.Filters {
			if f.Search != nil {
				filterSearch(c, f, opts.Searched)
			}
		}
		if len(opts.Filters) == 0 {
			ui.Text(c, core.Msg(c, "devtools.noFilters", "No filters")).
				TextColor(k.TextFaint).FontSize(core.FontSize(c, theme.MetaSize)).SingleLine()
		}
	})
	return r
}

// filterSelect is one group of choices, as a select.
func filterSelect(c *ui.Context, f DashboardFilterOptions) {
	group := f.Group
	if group == "" {
		group = core.Msg(c, "devtools.filter", "Filter")
	}
	ui.Box(c).Shrink(0).Children(func() {
		input.Select(c, f.Selected, choicesOf(f.Options), input.SelectOptions{
			Label:       group,
			Placeholder: group,
		})
	})
}

// filterRange is one group's range, as a slider between its two ends.
func filterRange(c *ui.Context, f DashboardFilterOptions) {
	if f.Placed == nil {
		return
	}
	u := core.Density(c).Unit()
	group := f.Group
	if group == "" {
		group = core.Msg(c, "devtools.range", "Range")
	}
	// The slider's own value is a float64 and the range is two ints, so the
	// caller's int is the same variable the slider writes through. That is
	// the price of reusing the control, and it is worth paying: a filter's
	// range would otherwise be a second control with its own detents and its
	// own keyboard.
	value := float64(*f.Placed)
	ui.Row(c).FillWidth().Gap(u * 2).AlignItems(ui.Center).Children(func() {
		input.Slider(c, &value, input.SliderOptions{
			Label: group, ShowValue: true,
			Min: float64(f.Min), Max: float64(f.Max), Step: 1,
		})
	})
	*f.Placed = int(value)
}

// filterSearch is the search field, which submits on Enter.
func filterSearch(c *ui.Context, f DashboardFilterOptions, searched func()) {
	label := f.SearchLabel
	if label == "" {
		label = core.Msg(c, "devtools.search", "Search")
	}
	field := input.SearchField(c, f.Search, label)
	if field.Submitted() && searched != nil {
		searched()
	}
}

// RefreshIntervalOptions configure a RefreshIntervalSelector.
type RefreshIntervalOptions struct {
	// Seconds is the chosen interval, written as it changes. It is a number
	// rather than an index into the list so that a caller holding "30" in a
	// URL keeps working when the list is reordered.
	Seconds *int
	// Options are the intervals offered in seconds; empty offers off, five,
	// thirty, sixty and five minutes, which is the set every dashboard and
	// every monitoring tool ships.
	Options []int
}

// RefreshIntervalSelectorResult carries a RefreshIntervalSelector.
type RefreshIntervalSelectorResult struct {
	// Element is the control.
	Element *ui.Element
}

// defaultIntervals is the interval set, in seconds. Zero is first and means
// not refreshing at all, which is a real answer for somebody looking at a
// frozen snapshot and is worth having before the fastest one.
func defaultIntervals() []int { return []int{0, 5, 30, 60, 300} }

// RefreshIntervalSelector is how often a dashboard refreshes itself.
//
// It is a segmented control rather than a select because the choices are a
// short ordered list of similar sizes, which is the case a segmented control
// is for: every answer is visible at once and one press sets it. A select
// would hide five numbers behind a closed trigger for no benefit.
//
// Off is in the list and is first. Somebody who has deliberately stopped a
// dashboard from refreshing — to read a number that is about to change, or to
// save a battery — has done something this control should be able to say so.
func RefreshIntervalSelector(c *ui.Context, opts RefreshIntervalOptions) RefreshIntervalSelectorResult {
	if opts.Seconds == nil {
		panic("devtools: RefreshIntervalSelector needs the *int Seconds writes to")
	}
	list := opts.Options
	if len(list) == 0 {
		list = defaultIntervals()
	}
	if *opts.Seconds < 0 {
		panic("devtools: RefreshIntervalSelector is given " + strconv.Itoa(*opts.Seconds) +
			" seconds; a negative interval is not a slow one")
	}
	at := -1
	for i, v := range list {
		if v == *opts.Seconds {
			at = i
			break
		}
	}
	if at < 0 {
		// A value the list does not offer is refused rather than snapped to
		// the nearest: a URL carrying "17 seconds" would silently become
		// thirty, and the number on the screen would not be the number in the
		// address.
		panic("devtools: RefreshIntervalSelector is set to " + strconv.Itoa(*opts.Seconds) +
			", which is not one of its intervals")
	}

	labels := make([]string, len(list))
	for i, v := range list {
		labels[i] = intervalLabel(v)
	}
	selected := at
	input.Segmented(c, &selected, labels...)
	*opts.Seconds = list[selected]

	var r RefreshIntervalSelectorResult
	r.Element = ui.Box(c).Label(core.Msg(c, "devtools.refresh", "Refresh"))
	return r
}

// intervalLabel is one interval as the button reads: "Off", "30s", "5m".
func intervalLabel(seconds int) string {
	switch {
	case seconds <= 0:
		return "Off"
	case seconds < 60:
		return strconv.Itoa(seconds) + "s"
	case seconds%60 == 0 && seconds < 3600:
		return strconv.Itoa(seconds/60) + "m"
	default:
		return trimFloat(float64(seconds)/60) + "m"
	}
}

// TrimToken is a word taken out of a string, for the search a caller runs
// over a value before handing it to a control. It exists so that "the search
// field" has one meaning of "trimmed" rather than one per call site.
func TrimToken(s string) string { return strings.TrimSpace(s) }
