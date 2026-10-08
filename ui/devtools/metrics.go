package devtools

import (
	"sort"
	"strconv"
	"time"

	"github.com/HycJack/MintUI/ui/core"
)

// Slot is one probe: whether a thing was up at a moment.
type Slot struct {
	// At is when it was probed.
	At time.Time
	// Up is whether it answered.
	Up bool
	// Error is why it did not, for a slot that was down. It is the caller's
	// string because the reason belongs to whatever was being probed — a
	// certificate about to expire reads nothing like a disk that is full.
	Error string
}

// SlotState is what a slot draws as: the same three answers plus a "none".
type SlotState int

const (
	// SlotUnknown is a slot with no record. It is not the same as down: a
	// year of history that starts today has no record of yesterday either.
	SlotUnknown SlotState = iota
	// SlotUp answered.
	SlotUp
	// SlotDown did not.
	SlotDown
	// SlotPartial answered, but with something wrong — the distinction
	// between "down" and "degraded" is the one an uptime bar exists to make,
	// because a service answering 500s is not serving.
	SlotPartial
)

// UptimeSlotsDays is how many days an uptime bar draws. Ninety is three
// months, which is the shortest window in which a weekly pattern is visible
// — and a bar shorter than that shows one bad afternoon and calls it a trend.
const UptimeSlotsDays = 90

// UptimeSlots turns a set of probes into one square per day, oldest first.
//
// The rules are the ones that make an uptime bar honest:
//
//   - No record is unknown, not up. A service nobody has monitored has not
//     been up all along, and drawing ninety green squares for a service that
//     was installed on Tuesday is a lie told in the friendliest colour
//     available.
//   - A day with at least one failed probe is down for the day, however many
//     others succeeded. An uptime bar answers "was it ever not working", not
//     "what fraction of the time", because a person reading it is asking
//     whether there was a day they should not have relied on it.
//   - A day whose failures all carry an error is partial rather than down.
//     Degraded is a different conversation from broken, and a bar that paints
//     both the same colour cannot be read.
//
// The bars are folded into days here rather than at each drawing site, so
// that a bar in a row of statistics and a bar on a service's own page are the
// same ninety numbers.
func UptimeSlots(records []Slot) []SlotState {
	out := make([]SlotState, UptimeSlotsDays)
	if len(records) == 0 {
		return out
	}

	// The window ends on the newest record and runs back ninety days, rather
	// than covering the first ninety days of the year. A bar anchored to
	// January would be empty for most of the year and would start filling
	// again on the first of January, which is not what anybody means by
	// "the last ninety days".
	last := records[0].At
	for _, r := range records {
		if r.At.After(last) {
			last = r.At
		}
	}
	first := dayNumber(last) - (UptimeSlotsDays - 1)

	// Fold each day to one answer: any failure downgrades the day, and an
	// unknown can only be a day with no probes at all.
	for _, r := range records {
		i := dayNumber(r.At) - first
		if i < 0 || i >= UptimeSlotsDays {
			continue
		}
		switch {
		case !r.Up:
			out[i] = worse(out[i], SlotDown)
		case r.Error != "":
			out[i] = worse(out[i], SlotPartial)
		case out[i] == SlotUnknown:
			out[i] = SlotUp
		}
	}
	return out
}

// dayNumber is a date as a whole number of days since 1970-01-01.
//
// It is the days-from-civil algorithm rather than a subtraction of two
// time.Time values because a subtraction across a daylight-saving boundary
// is an hour short, and an uptime bar whose last day is one hour off is a
// bar that puts a probe in the wrong square on two days a year and nobody
// can say why.
func dayNumber(t time.Time) int {
	y, mon, d := t.Date()
	// March is the pivot: a leap day falls at the end of the year we shift
	// the whole thing to, which is why the second and third months are
	// counted from March and not from January.
	m := int(mon)
	if m <= 2 {
		y--
		m += 12
	}
	era := y / 400
	if y < 0 {
		era = (y - 399) / 400
	}
	yoe := y - era*400
	doy := (153*(m-3)+2)/5 + d - 1
	doe := yoe*365 + yoe/4 - yoe/100 + doy
	return era*146097 + doe - 719468
}

// worse is the worse of two answers, where down is worse than partial and
// partial worse than up. It is a function because "which of these two
// answers is the one to draw" is a rule and a rule written as an if at each
// place is three ifs.
func worse(a, b SlotState) SlotState {
	if rank(b) > rank(a) {
		return b
	}
	return a
}

// rank is how serious an answer is.
func rank(s SlotState) int {
	switch s {
	case SlotDown:
		return 3
	case SlotPartial:
		return 2
	case SlotUp:
		return 1
	}
	return 0
}

// UptimePercent is how much of a set of slots was up, 0 to 1.
//
// Only days with a record count. An unknown day is excluded from both sides
// of the division rather than counted as an up one: the alternative is a
// service installed last week showing 97% uptime, which is arithmetically
// defensible and practically a lie. A window with no records at all is 0,
// because "no data" and "a perfect record" must not both be 1.
func UptimePercent(slots []SlotState) float32 {
	known, up := 0, 0
	for _, s := range slots {
		if s == SlotUnknown {
			continue
		}
		known++
		if s == SlotUp {
			up++
		}
	}
	if known == 0 {
		return 0
	}
	return float32(up) / float32(known)
}

// UptimeTone is the severity a service's availability is drawn at.
//
// The steps are where somebody would open an incident. A tenth of a percent
// of downtime in a quarter is three hours and is nothing; five per cent is a
// day somebody spent telling customers about.
func UptimeTone(uptime float32) core.Severity {
	switch {
	case uptime >= 0.999:
		return core.Success
	case uptime >= 0.99:
		return core.Neutral
	case uptime >= 0.95:
		return core.Warning
	default:
		return core.Danger
	}
}

// Span is one span of a trace: something that took a length of time inside a
// longer thing.
type Span struct {
	// Name is what it is called, as the tracing system named it.
	Name string
	// Start is how far into the whole trace it began.
	Start time.Duration
	// Duration is how long it took.
	Duration time.Duration
	// Error marks a span that failed, which is the one thing anybody
	// scanning a waterfall is looking for.
	Error bool
}

// SpanBar is where a span sits in a trace, as percentages of the whole.
//
// The three are the left edge, the width and the right edge, and they are
// returned separately because they are three different numbers: a caller
// drawing a row needs the left edge to place the bar and the width to size
// it, and a bar drawn as one rectangle from the left would need the caller
// to work the arithmetic out for it. They sum to a hundred across a trace,
// which is the property a waterfall's whole value rests on — a span shown at
// 40% of a trace really did take 40% of its time.
//
// A span is clamped to the trace rather than refused: a trace whose spans do
// not add up is a broken trace, and refusing to draw any of it would leave
// somebody with no way to see the one span that is obviously wrong. A span
// that runs past the end is drawn ending at the end, and a negative one as no
// width at all.
func SpanBar(s Span, total time.Duration) (left, width float32) {
	if total <= 0 || s.Duration <= 0 {
		return 0, 0
	}
	pct := func(d time.Duration) float32 {
		return float32(d) / float32(total) * 100
	}
	left = pct(s.Start)
	width = pct(s.Duration)
	if left < 0 {
		left = 0
	}
	if left+width > 100 {
		width = 100 - left
	}
	if width < 0 {
		width = 0
	}
	return left, width
}

// TotalOf is how long a whole trace took: the end of its last span.
//
// The end of the furthest span, not the sum of them — spans nest, so a trace
// of a request that spent 40ms in a database call and 100ms waiting for it
// took 140ms, not 40 and not 100. Getting this wrong scales the whole
// waterfall, so it is one function rather than something each caller sums.
func TotalOf(spans []Span) time.Duration {
	var end time.Duration
	for _, s := range spans {
		if e := s.Start + s.Duration; e > end {
			end = e
		}
	}
	return end
}

// DepthOf is how deep a span sits inside the others, which is the row it is
// indented to.
//
// A span is a child of whichever span contains it and starts after it. The
// outermost spans are at depth zero, which is what makes the waterfall read
// as nesting rather than as a list that happens to be in order.
func DepthOf(spans []Span, at int) int {
	if at < 0 || at >= len(spans) {
		return 0
	}
	s := spans[at]
	depth := 0
	for i := range spans {
		if i == at {
			continue
		}
		p := spans[i]
		if p.Start <= s.Start && p.Start+p.Duration >= s.Start+s.Duration &&
			p.Duration > s.Duration {
			depth++
		}
	}
	return depth
}

// LogLevel is how loud a log line is.
type LogLevel int

const (
	// Debug is something written only to be read while looking for a problem.
	Debug LogLevel = iota
	// Info is the ordinary running commentary.
	Info
	// Warn is something that was handled but should not have been.
	Warn
	// Error is something that failed.
	Error
	// Fatal is something that stopped the process.
	Fatal
)

// String is the level's own word, in the capitalisation log tooling uses.
func (l LogLevel) String() string {
	switch l {
	case Info:
		return "INFO"
	case Warn:
		return "WARN"
	case Error:
		return "ERROR"
	case Fatal:
		return "FATAL"
	}
	return "DEBUG"
}

// ParseLogLevel reads a level as a log file writes it. The names are matched
// without regard to case and an unknown name is Debug rather than an error:
// a viewer that refused to draw a line because somebody wrote "warn" in lower
// case is not a viewer anybody would keep open.
func ParseLogLevel(name string) LogLevel {
	switch upperASCII(name) {
	case "INFO":
		return Info
	case "WARN", "WARNING":
		return Warn
	case "ERROR", "ERR":
		return Error
	case "FATAL", "CRITICAL", "PANIC":
		return Fatal
	}
	return Debug
}

// Severity is the tone a log level is drawn at, which is core's rather than
// this package's: the levels in a log and the pills on a board are the same
// four answers to the same question.
func (l LogLevel) Severity() core.Severity {
	switch l {
	case Info, Debug:
		return core.Neutral
	case Warn:
		return core.Warning
	case Error, Fatal:
		return core.Danger
	}
	return core.Neutral
}

// LogLine is one line of a stream.
type LogLine struct {
	// At is when it was written.
	At time.Time
	// Level is how loud it was.
	Level LogLevel
	// Source is what wrote it — a file, a service, a test name.
	Source string
	// Message is the line itself.
	Message string
}

// FilterLogs is the lines a view shows for a level, a search and a limit.
//
// The level filter is "this level and louder" rather than "exactly this
// level", because a reader who has asked for errors wants to see them
// alongside what they came after. The search is case-insensitive over the
// message and the source, because a person looking for a request id looks
// for it in both.
//
// The limit is applied last and keeps the *newest* lines, not the first: a
// stream that stops drawing at ten thousand lines and shows the ones from
// last Tuesday is worse than useless during an incident.
func FilterLogs(lines []LogLine, level LogLevel, search string, limit int) []LogLine {
	needle := lowerASCII(search)
	out := make([]LogLine, 0, len(lines))
	for _, l := range lines {
		if l.Level < level {
			continue
		}
		if needle != "" &&
			!containsASCII(lowerASCII(l.Message), needle) &&
			!containsASCII(lowerASCII(l.Source), needle) {
			continue
		}
		out = append(out, l)
	}
	if limit > 0 && len(out) > limit {
		out = out[len(out)-limit:]
	}
	return out
}

// upperASCII is strings.ToUpper for ASCII. It is here rather than imported
// because the log filter runs over every line of a stream on every frame, and
// a stream can be large: Go's own strings.ToUpper walks runes and allocates
// for the ones that change, and log text is overwhelmingly ASCII that does
// not.
func upperASCII(s string) string {
	hasLower := false
	for i := 0; i < len(s); i++ {
		if c := s[i]; c >= 'a' && c <= 'z' {
			hasLower = true
			break
		}
	}
	if !hasLower {
		return s
	}
	b := []byte(s)
	for i, c := range b {
		if c >= 'a' && c <= 'z' {
			b[i] = c - ('a' - 'A')
		}
	}
	return string(b)
}

// lowerASCII is strings.ToLower for the same reason as upperASCII.
func lowerASCII(s string) string {
	hasUpper := false
	for i := 0; i < len(s); i++ {
		if c := s[i]; c >= 'A' && c <= 'Z' {
			hasUpper = true
			break
		}
	}
	if !hasUpper {
		return s
	}
	b := []byte(s)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + ('a' - 'A')
		}
	}
	return string(b)
}

// containsASCII is strings.Contains, named apart so that the three helpers
// above read as a set rather than as three unrelated functions.
func containsASCII(s, sub string) bool {
	if len(sub) == 0 {
		return true
	}
	if len(sub) > len(s) {
		return false
	}
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// SlowestSpans are the spans worth looking at first: the longest few, longest
// first.
//
// It is a copy of a sort rather than a sort of the caller's slice, because a
// log or a trace is usually being filtered at the same time as it is being
// ranked, and a component that reordered the caller's list would make the two
// answers depend on the order they were asked in.
func SlowestSpans(spans []Span, n int) []Span {
	out := make([]Span, len(spans))
	copy(out, spans)
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Duration > out[j].Duration
	})
	if n > 0 && len(out) > n {
		out = out[:n]
	}
	return out
}

// RoundDuration is a duration in the fewest units that say it exactly enough
// for a dashboard — "1.2s", "340ms", "2m".
//
// The threshold at which a span is reported in milliseconds rather than
// microseconds is one millisecond, because anything under it is far below what
// anybody reading a waterfall is looking at and the extra digits are noise.
func RoundDuration(d time.Duration) string {
	switch {
	case d <= 0:
		return "0"
	case d < time.Millisecond:
		return "µs"
	case d < time.Second:
		return trimFloat(float64(d)/float64(time.Millisecond)) + "ms"
	case d < time.Minute:
		return trimFloat(float64(d)/float64(time.Second)) + "s"
	case d < time.Hour:
		return trimFloat(float64(d)/float64(time.Minute)) + "m"
	default:
		return trimFloat(float64(d)/float64(time.Hour)) + "h"
	}
}

// trimFloat is one decimal place with the trailing zero taken off, so 1.0 is
// "1" and 1.25 is "1.3".
//
// The rounding is manual rather than strconv's because strconv has no format
// that means "at most one decimal place": %.1f always writes the zero, and
// every duration on a dashboard would read "1.0s" beside "340ms".
func trimFloat(f float64) string {
	tenths := int64(f*10 + 0.5)
	if tenths < 0 {
		tenths = -tenths
	}
	whole, frac := tenths/10, tenths%10
	if frac == 0 {
		return strconv.FormatInt(whole, 10)
	}
	return strconv.FormatInt(whole, 10) + "." + strconv.FormatInt(frac, 10)
}
