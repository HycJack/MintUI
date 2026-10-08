package feedback

import (
	"strconv"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// ProgressOptions configure a Progress.
type ProgressOptions struct {
	// Value is how far along the work is, 0 to 1. A negative value is work
	// of unknown length, which the bar says by moving rather than by
	// filling — a bar parked at 30% with nothing behind it reads as a failed
	// job, where a moving one reads as a job.
	Value float32
	// Width and Height bound the bar. A zero Width takes the parent's, a
	// zero Height one standard bar.
	Width, Height float32
	// Severity colours the filled part. The zero value is core.Neutral,
	// which takes the accent rather than ordinary text: a progress bar in
	// text grey reads as a disabled control.
	Severity core.Severity
	// Label names the work. It is required: a bar with no name reaches a
	// screen reader as "progress bar", which is what every progress bar
	// says.
	Label string
	// ShowPercent puts the value beside the bar as a caption. It is off by
	// default because the bar already shows the value, and two numbers side
	// by side is two chances to disagree.
	ShowPercent bool
	// Period is how long an unknown-length bar takes to cross, in
	// milliseconds. Zero takes the loaders' period, so the two kinds of
	// spinner in one panel do not beat against each other.
	Period float32
}

// Progress is a bar showing how far a piece of work has got.
//
// It holds nothing: Value is the caller's number, read again each frame, so
// several bars can move at their own rates side by side without any of them
// owning a clock.
func Progress(c *ui.Context, opts ProgressOptions) *ui.Element {
	if opts.Label == "" {
		panic("feedback: Progress needs a Label; an unnamed bar says nothing")
	}
	if opts.Value > 1 {
		// The likeliest mistake by far is passing a percentage, which draws
		// as a full bar and reads as a finished job. Worth stopping for.
		panic("feedback: Progress Value is 0 to 1; a value above 1 looks like " +
			"a percentage, which this bar is not")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	h := opts.Height
	if h <= 0 {
		h = u * 2.25
	}
	ink := progressInk(k, opts.Severity)
	still := core.Reduced(c)
	period := opts.Period
	if period <= 0 {
		period = loaderPeriod
	}
	value := clamp01(opts.Value)

	// The bar is built by a closure rather than once up front, because an
	// element belongs to whatever was being built when it was made: a bar
	// made here would land in the caller's container, not in the row the
	// percentage caption puts it in.
	bar := func() *ui.Element {
		e := ui.Box(c).FillWidth().Height(h).Radius(h/2).Clip().Shrink(0).
			Background(k.Surface).Role(ui.RoleProgress).Label(opts.Label).
			Range(0, 1, float64(max(opts.Value, 0)))
		if opts.Width > 0 {
			e.Width(opts.Width)
		}
		return e.Draw(func(p *ui.Painter, r ui.Rect) {
			if opts.Value >= 0 {
				// Determinant: the filled part is the value and no more, in
				// every frame and under either motion setting.
				p.Fill(ui.Rect{X: r.X, Y: r.Y, W: r.W * value, H: r.H}, ink, h/2)
				return
			}
			// Unknown length: a band crossing the track and off the end. Still,
			// the phase is the middle, which puts the band in the centre of the
			// track — the one place a stopped sweep does not read as a claim
			// about how far the work has got.
			w := r.W * 0.3
			x := r.X - w + (r.W+w)*loopPhase(p, ms(period), still)
			p.Fill(ui.Rect{X: x, Y: r.Y, W: w, H: r.H}, ink, h/2)
		})
	}

	if !opts.ShowPercent {
		return bar()
	}
	return ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 2).
		Label(opts.Label).Children(func() {
		// The spacer rather than a wrapper: a Grow on the bar itself would
		// let the caption push it out of the row's share.
		ui.Box(c).Grow(1)
		bar()
		ui.Text(c, percentText(opts.Value)).TextColor(k.TextMuted).
			FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
	})
}

// percentText renders a bar's value as the percentage a person would say, and
// leaves an unknown length as a dash rather than as 0%, which would be a
// claim about the work rather than about the bar.
func percentText(v float32) string {
	if v < 0 {
		return "—"
	}
	return strconv.Itoa(int(v*100+0.5)) + "%"
}

// MeterOptions configure a Meter.
type MeterOptions struct {
	// Value is how much of the capacity is in use, on the same scale as
	// Max: 4.2 of 10 gigabytes, 12 of 16 turns of context. A value above
	// Max is not an error — an over-full store is the one thing this bar
	// exists to shout about — and it draws full, in the danger colour. A
	// negative value is an error: a used capacity below zero is an input
	// mistake, not a reading.
	Value float32
	// Max is the capacity, the number the value is against. It must be
	// above zero: a value against nothing is not a fraction, and this bar
	// will not invent one.
	Max float32
	// Label names the capacity. It is required, for the same reason a
	// Progress's is: "Storage" and "Context" are different meters even when
	// both sit at the same fraction.
	Label string
	// Threshold is a value on the Max scale at which the fill escalates
	// from the accent to the warning, and a tick is drawn at its position.
	// Zero draws no threshold. The escalation to danger needs no setting:
	// over full is over full.
	Threshold float32
	// ShowValue puts "value / max" beside the bar as a caption. Off by
	// default, for the same reason a Progress keeps its percentage off.
	ShowValue bool
	// Width and Height bound the bar, the same way a Progress's do: zero
	// Width takes the parent's, zero Height a standard bar.
	Width, Height float32
}

// Meter is a bar showing how much of a fixed capacity a thing has taken:
// storage against its quota, a context window against its limit, a counter
// against its ceiling.
//
// It is a Meter and not a Progress because the two answer different
// questions. Progress is "how far along is this piece of work": the number
// moves one way and finishes at a hundred percent, and a hundred percent is
// the good outcome. A Meter is "how full is this container": the number
// goes both ways, may never reach the top, and reaching the top is the bad
// outcome. A progress bar at a hundred percent is a finished job; a meter
// at a hundred percent is a full disk.
//
// It holds nothing, the way a Progress does: Value is the caller's number,
// read again each frame, so several meters can move at their own rates
// without any of them owning a clock.
func Meter(c *ui.Context, opts MeterOptions) *ui.Element {
	if opts.Label == "" {
		panic("feedback: Meter needs a Label; an unnamed bar says nothing")
	}
	if opts.Max <= 0 {
		panic("feedback: Meter needs a Max above 0; a value against a zero " +
			"capacity is not a fraction")
	}
	if opts.Value < 0 {
		panic("feedback: Meter Value cannot be negative; a used capacity " +
			"below zero is an input mistake, not a reading")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	h := opts.Height
	if h <= 0 {
		h = u * 2.25
	}
	frac := clamp01(opts.Value / opts.Max)

	// The fill escalates as the capacity runs out: past the threshold the
	// accent becomes a warning, and at the top — or past it — a danger.
	// An over-full value is the only input that makes a meter urgent on its
	// own, and it is the one a caller would least expect to have to mark by
	// hand.
	sev := core.Neutral
	if frac >= 1 {
		sev = core.Danger
	} else if opts.Threshold > 0 && opts.Value >= opts.Threshold {
		sev = core.Warning
	}
	ink := progressInk(k, sev)

	bar := func() *ui.Element {
		e := ui.Box(c).FillWidth().Height(h).Radius(h/2).Clip().Shrink(0).
			Background(k.Surface).Role(ui.RoleProgress).Label(opts.Label).
			Range(0, float64(opts.Max), float64(opts.Value))
		if opts.Width > 0 {
			e.Width(opts.Width)
		}
		return e.Draw(func(p *ui.Painter, r ui.Rect) {
			p.Fill(ui.Rect{X: r.X, Y: r.Y, W: r.W * frac, H: r.H}, ink, h/2)
			// The ticks are the meter's scale: the value's own position, in
			// the window's text, and — while the fill has not reached it —
			// the threshold, in the faint. A tick drawn in the fill's colour
			// would be invisible the moment the fill covers it.
			tw := max(1, h*0.12)
			p.Fill(ui.Rect{X: r.X + r.W*frac - tw/2, Y: r.Y, W: tw, H: r.H},
				k.Text, 0)
			if opts.Threshold > 0 && opts.Value < opts.Threshold {
				tx := r.X + r.W*clamp01(opts.Threshold/opts.Max)
				p.Fill(ui.Rect{X: tx - tw/2, Y: r.Y, W: tw, H: r.H},
					k.TextFaint, 0)
			}
		})
	}

	if !opts.ShowValue {
		return bar()
	}
	return ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 2).
		Label(opts.Label).Children(func() {
		ui.Box(c).Grow(1)
		bar()
		ui.Text(c, meterValue(opts.Value)+" / "+meterValue(opts.Max)).
			TextColor(k.TextMuted).
			FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
	})
}

// meterValue renders a meter's number the way a person reads it off the dial:
// 4.2, not 4.200000476837158, and 10, not 1e+01.
func meterValue(v float32) string {
	return strconv.FormatFloat(float64(v), 'g', 4, 64)
}

// progressInk is the colour a bar's filled part takes. core.Neutral takes the
// accent rather than the ordinary text a neutral mark would take: a progress
// bar is never passive content, and a bar in text grey reads as a disabled
// control rather than as work going on.
func progressInk(k theme.Tokens, sev core.Severity) ui.Color {
	if sev == core.Neutral {
		return k.Accent
	}
	_, fg := sev.Pair(k)
	return fg
}

// severityInk is the colour a severity is marked with when it is drawn as
// something solid — a filled bar, a dot, a side rule — rather than as a pill
// with words on it. core.Neutral takes ordinary text, because a mark in the
// accent would be claiming the system has an opinion where it has none.
func severityInk(k theme.Tokens, sev core.Severity) ui.Color {
	if sev == core.Neutral {
		return k.TextMuted
	}
	_, fg := sev.Pair(k)
	return fg
}

// StatusIndicatorOptions configure a StatusIndicator.
type StatusIndicatorOptions struct {
	// Label is what the status says, and is required: a coloured dot with
	// no word beside it is a decoration, not a status.
	Label string
	// Severity picks the mark's colour. Zero is core.Neutral, which is what
	// an ordinary in-progress callback wants, and takes the same pairing a
	// priority pill takes so a column of statuses and a row of priorities
	// read as one system.
	Severity core.Severity
	// Busy replaces the dot with three moving ones, for work that is
	// happening but has no outcome yet. The pill stops making a claim and
	// starts reporting.
	Busy bool
	// Pill puts the label on the severity's background. Off by default,
	// because most statuses are one pill among several and an unbacked one
	// reads lighter, which is usually what is wanted.
	Pill bool
}

// StatusIndicator is a small mark and a word: "Scheduled", "On hold",
// "Resolved".
func StatusIndicator(c *ui.Context, opts StatusIndicatorOptions) *ui.Element {
	if opts.Label == "" {
		panic("feedback: StatusIndicator needs a Label; a dot with no word is a " +
			"decoration")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	dot := u * 2.25
	// A background is the severity's own pairing rather than any colour of
	// our own, so this and the board's priority pills cannot drift apart.
	// Without one the severity tints the mark alone and the word stays
	// ordinary text, which is the only way a caption stays readable.
	bg, fg := ui.Transparent, severityInk(k, opts.Severity)
	if opts.Pill {
		bg, fg = opts.Severity.Pair(k)
	}

	pill := ui.Row(c).AlignItems(ui.Center).Gap(u * 1.25).Shrink(0).
		Label(opts.Label).Children(func() {
		if opts.Busy {
			DotsLoader(c, LoaderOptions{Size: dot * 2.5, Color: fg, Label: opts.Label})
		} else {
			ui.Box(c).Size(dot, dot).Radius(dot / 2).Background(fg).Shrink(0)
		}
		ui.Text(c, opts.Label).TextColor(fg).
			FontSize(core.FontSize(c, theme.RowSize)).SingleLine()
	})
	if bg.A != 0 {
		pill.Background(bg).Padding(u, u*2, u, u*2).Radius(theme.PillRadius)
	}
	return pill
}

// PresenceState is whether someone is there.
type PresenceState int

const (
	// PresenceOffline is the zero value: an empty circle.
	PresenceOffline PresenceState = iota
	// PresenceOnline is the filled bright dot.
	PresenceOnline
	// PresenceAway is a hollow ring — here, but not now.
	PresenceAway
	// PresenceBusy is moving dots — here, and unable to answer.
	PresenceBusy
)

func (s PresenceState) String() string {
	switch s {
	case PresenceOnline:
		return "Online"
	case PresenceAway:
		return "Away"
	case PresenceBusy:
		return "Busy"
	}
	return "Offline"
}

// PresenceOptions configure a Presence.
type PresenceOptions struct {
	// Name is who is present. It is required: an unattributed dot among
	// several is noise.
	Name string
	// State is which of the four marks to draw.
	State PresenceState
	// ShowName puts the name beside the mark; off for a column that has
	// already said who each row is about.
	ShowName bool
	// Size is the mark's diameter in DIPs. Zero takes the density's.
	Size float32
}

// Presence is one person's mark: online, away, busy or offline.
//
// Only Busy moves, and only because the other three already say their whole
// story — a breathing "away" would be decoration. The online dot is Lively,
// the one colour in the interface that does not follow the appearance: a
// presence dot that went grey on a dark desktop would stop being the thing
// you look for.
func Presence(c *ui.Context, opts PresenceOptions) *ui.Element {
	if opts.Name == "" {
		panic("feedback: Presence needs a Name; an unattributed dot says nothing")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	side := opts.Size
	if side <= 0 {
		side = u * 3.5
	}
	rad := side / 2
	name := opts.Name + " " + opts.State.String()

	// The mark is made inside the row rather than beside it: an element
	// belongs to whatever was being built when it was made, so a mark made
	// out here would land in the caller's container instead of this row's.
	mark := func() *ui.Element {
		switch opts.State {
		case PresenceOnline:
			return ui.Box(c).Size(side, side).Radius(rad).Shrink(0).
				Background(k.Lively)
		case PresenceAway:
			// Hollow, so it cannot be mistaken for online at a glance.
			return ui.Box(c).Size(side, side).Radius(rad).Shrink(0).
				Border(max(1, side*0.14), k.Warning)
		case PresenceBusy:
			// The loader's own mark, not dots inside a circle: that would be
			// five marks, and none of them is the person.
			return DotsLoader(c, LoaderOptions{
				Size: side, Color: k.Accent, Label: name,
			})
		}
		return ui.Box(c).Size(side, side).Radius(rad).Shrink(0).
			Border(theme.BorderWidth, k.TextFaint)
	}

	return ui.Row(c).AlignItems(ui.Center).Gap(u * 1.5).Shrink(0).
		Label(name).Children(func() {
		mark()
		if opts.ShowName {
			ui.Text(c, opts.Name).TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).SingleLine()
		}
	})
}
