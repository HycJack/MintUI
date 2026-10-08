package project

import (
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/chart"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/theme"
)

// FormatElapsed is a duration as a clock, which is the form a person reads it
// in: "1:05:00" rather than "3900s", because the first is what every timer
// in the world shows and the second has to be converted in the reader's head.
//
// It goes to hours without a limit, because a sprint can be longer than a
// day and a clock that wrapped at 24 hours would lose the very hours somebody
// is checking. Under a minute it says "0:00" rather than "0:00:00", because
// the seconds are noise at that size and the two shorter forms are the two
// anybody actually starts a timer with.
func FormatElapsed(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	total := int(d / time.Second)
	switch {
	case total < 3600:
		// Minutes unpadded and seconds padded, which is what a stopwatch
		// does: "0:05" and "1:05", never "00:05". Padding the minutes as well
		// makes a fresh timer read 00:00, which looks like it has been
		// running for a while and counts the wrong number of digits.
		return itoa(total/60) + ":" + pad2(total%60)
	default:
		return itoa(total/3600) + ":" + pad2((total/60)%60) + ":" + pad2(total%60)
	}
}

// pad2 writes a number in two digits, which is what a clock does.
func pad2(n int) string {
	if n < 10 {
		return "0" + itoa(n)
	}
	return itoa(n)
}

// TimeEntry is one stretch of work somebody logged.
type TimeEntry struct {
	// Task is what it was logged against.
	Task string
	// Who logged it.
	Who string
	// Spent is how long it took.
	Spent time.Duration
	// Billable marks the stretch somebody is charging for, which is the only
	// thing that makes it a different kind of row rather than a detail.
	Billable bool
}

// TimeTrackerOptions configure a TimeTracker.
type TimeTrackerOptions struct {
	// Running is whether the clock is going, and Elapsed what it has
	// accumulated. Both are the caller's: a stopwatch that owned its own
	// start time would have to be told about the window closing, and the
	// time worked is a fact about work rather than about a widget.
	Running *bool
	Elapsed time.Duration
	// Entries are the stretches already logged, newest first as the caller
	// has them.
	Entries []TimeEntry
	// Started and Stopped ask for the clock to be started or stopped.
	Started func()
	Stopped func()
	// Logged asks for the running time to be written to an entry.
	Logged func()
}

// TimeTrackerResult carries a TimeTracker and what was done in it.
type TimeTrackerResult struct {
	// Element is the whole tracker.
	Element *ui.Element
	// running is this frame's clock, for a caller that wants the number
	// rather than the drawing.
	running bool
}

// Running reports the clock's state as this frame saw it.
func (r TimeTrackerResult) Running() bool { return r.running }

// TimeTracker is a clock and the stretches already logged: the running time
// big, a start and a stop, and the list underneath.
//
// The clock is the caller's elapsed time formatted by FormatElapsed, not a
// second hand drawn here. That means the number on the screen and the number
// the caller would log are the same formatting of the same value, which is the
// only way a person can trust that what they read is what will be recorded.
//
// Billable entries are marked in the accent rather than being pulled out into
// a separate list, because "which of these can I invoice for" is a question
// about the whole list rather than about one of its halves.
func TimeTracker(c *ui.Context, opts TimeTrackerOptions) TimeTrackerResult {
	if opts.Running == nil {
		panic("project: TimeTracker needs the *bool Running writes to")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	var r TimeTrackerResult
	r.running = *opts.Running
	r.Element = ui.Column(c).FillWidth().Gap(u * 2).
		Label(core.Msg(c, "project.timeTracker", "Time")).Children(func() {
		// The clock: a figure big enough to read across a room, with the
		// running mark beside it as a word rather than as a colour, because
		// a running clock that is only coloured is a running clock that is
		// invisible to somebody who cannot see the colour.
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 2).Children(func() {
			ui.Text(c, FormatElapsed(opts.Elapsed)).TextColor(k.Text).Grow(1).
				FontSize(core.FontSize(c, theme.DisplaySize)).Font(monoFamily).SingleLine()
			if *opts.Running {
				sev := core.Success
				_, fg := sev.Pair(k)
				ui.Text(c, core.Msg(c, "project.running", "Running")).
					TextColor(fg).Shrink(0).
					FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
			}
		})

		ui.Row(c).FillWidth().Gap(u * 2).Children(func() {
			if *opts.Running {
				if opts.Stopped != nil && input.Button(c, core.Msg(c, "project.stop", "Stop"),
					input.ButtonOptions{Primary: true}).Clicked() {
					opts.Stopped()
				}
			} else {
				if opts.Started != nil && input.Button(c, core.Msg(c, "project.start", "Start"),
					input.ButtonOptions{Primary: true}).Clicked() {
					opts.Started()
				}
			}
			if opts.Logged != nil {
				if input.Button(c, core.Msg(c, "project.logTime", "Log time"),
					input.ButtonOptions{Disabled: opts.Elapsed <= 0}).Clicked() {
					opts.Logged()
				}
			}
		})

		for _, e := range opts.Entries {
			entryRow(c, e)
		}
	})
	return r
}

// entryRow is one logged stretch.
func entryRow(c *ui.Context, e TimeEntry) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	ui.Row(c).FillWidth().Gap(u*2).AlignItems(ui.Center).
		Padding(u, u*2).Radius(theme.ControlRadius).Background(k.Surface).
		Label(e.Task).Role(ui.RoleListItem).Children(func() {
		ui.Text(c, e.Task).TextColor(k.Text).Grow(1).Shrink(0).
			FontSize(core.FontSize(c, theme.RowSize)).SingleLine()
		ui.Text(c, FormatElapsed(e.Spent)).TextColor(k.TextMuted).Shrink(0).
			FontSize(core.FontSize(c, theme.MetaSize)).Font(monoFamily).SingleLine()
		if e.Billable {
			display.Tag(c, core.Msg(c, "project.billable", "Billable"),
				display.TagOptions{Tone: core.Accent})
		}
	})
}

// WorkloadEntry is one person's share of the work.
type WorkloadEntry struct {
	// Who is the person.
	Who string
	// Assigned is how many tasks are on their plate and Done how many of
	// them are closed. Both are counts the caller keeps.
	Assigned, Done int
	// Capacity is how many they can take on at once, zero for no opinion —
	// which draws the row without a full marker rather than claiming they
	// are at their limit.
	Capacity int
}

// WorkloadViewOptions configure a WorkloadView.
type WorkloadViewOptions struct {
	// Entries are the people, in the caller's order. The view does not sort
	// them: a workload somebody has arranged by themselves is an answer they
	// already worked out.
	Entries []WorkloadEntry
	// Height is the viewport's height, which a scroll area needs before it
	// will scroll.
	Height float32
	// Width is the width the rows are laid out in. Zero measures the window.
	Width float32
	// Empty draws instead of the rows when there are none.
	Empty func()
}

// WorkloadViewResult carries a WorkloadView.
type WorkloadViewResult struct {
	// Element is the whole view.
	Element *ui.Element
}

// WorkloadView is who has how much on: one row per person, a bar of what is
// assigned against what they can hold, and the numbers beside it.
//
// It is a chart, not a list of hand-drawn bars, because every rule about
// drawing a bar properly already lives in ui/chart — the scale that starts at
// zero, the palette that gives each row its own colour in both appearances,
// the frame that measures the gutter. A workload view is the one place where
// somebody is comparing magnitudes across rows, which is exactly what a chart
// is for and exactly what a column of divs gets wrong.
//
// The rows go in as one series per person so the legend can name them, which
// is what makes a colour readable back to a name. With no capacity anywhere
// the chart is skipped and the rows are drawn plainly, because a bar against
// nothing is a proportion of an unknown.
func WorkloadView(c *ui.Context, opts WorkloadViewOptions) WorkloadViewResult {
	k, u := core.Tokens(c), core.Density(c).Unit()

	var r WorkloadViewResult
	r.Element = ui.Column(c).FillWidth().Gap(u * 2).
		Label(core.Msg(c, "project.workload", "Workload"))
	r.Element.Children(func() {
		if len(opts.Entries) == 0 {
			if opts.Empty != nil {
				opts.Empty()
				return
			}
			ui.Text(c, core.Msg(c, "project.noWorkload", "Nobody has anything assigned")).
				TextColor(k.TextFaint).FontSize(core.FontSize(c, theme.RowSize)).SingleLine()
			return
		}
		if !anyCapacity(opts.Entries) {
			workloadRows(c, opts)
			return
		}
		workloadChart(c, opts)
	})
	return r
}

// anyCapacity reports whether any entry says what it can hold. One entry with
// a capacity is enough for the whole view to be a chart: the bars are read
// against the people who have said so, and the ones who have not are simply
// the ones with no line drawn over them.
func anyCapacity(entries []WorkloadEntry) bool {
	for _, e := range entries {
		if e.Capacity > 0 {
			return true
		}
	}
	return false
}

// workloadRows is the plain list, for a team where nobody has said what they
// can hold.
func workloadRows(c *ui.Context, opts WorkloadViewOptions) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	ui.Column(c).FillWidth().Gap(u).Children(func() {
		for _, e := range opts.Entries {
			ui.Row(c).FillWidth().Gap(u*2).AlignItems(ui.Center).
				Padding(u, u*2).Radius(theme.ControlRadius).
				Background(k.Surface).Label(e.Who).
				Role(ui.RoleListItem).Children(func() {
				display.Avatar(c, e.Who)
				ui.Text(c, e.Who).TextColor(k.Text).Grow(1).
					FontSize(core.FontSize(c, theme.RowSize)).SingleLine()
				ui.Text(c, itoa(e.Done)+" / "+itoa(e.Assigned)).TextColor(k.TextMuted).
					Shrink(0).FontSize(core.FontSize(c, theme.MetaSize)).SingleLine()
			})
		}
	})
}

// workloadChart is the chart, for a team that has said what it can hold.
//
// Two bars per person — what is on them and what they can take — rather than
// a stack. A stack would put the two numbers on top of one another and make
// the comparison somebody actually wants, "how close am I to my limit",
// into a subtraction. Side by side is a comparison, which is why it is the
// shape this chart is drawn in.
func workloadChart(c *ui.Context, opts WorkloadViewOptions) {
	names := make([]string, 0, len(opts.Entries))
	assigned := make([]float64, 0, len(opts.Entries))
	capacity := make([]float64, 0, len(opts.Entries))
	for _, e := range opts.Entries {
		names = append(names, e.Who)
		assigned = append(assigned, float64(e.Assigned))
		capacity = append(capacity, float64(e.Capacity))
	}

	series := []chart.Series{
		chart.SeriesFrom(core.Msg(c, "project.assigned", "Assigned"), chart.FormBar,
			sequence(len(assigned)), assigned),
		chart.SeriesFrom(core.Msg(c, "project.capacity", "Capacity"), chart.FormBar,
			sequence(len(capacity)), capacity),
	}

	chart.BarChart(c, chart.BarOptions{
		ChartOptions: chart.ChartOptions{
			Label:   core.Msg(c, "project.workload", "Workload"),
			Height:  opts.Height,
			Surface: true,
			Border:  true,
			Legend:  chart.LegendOptions{Entries: chart.EntriesOf(c, series)},
			X:       chart.AxisOptions{Label: "Person"},
			Y:       chart.AxisOptions{Side: chart.Left, Label: core.Msg(c, "project.tasks", "Tasks"), Count: 5},
		},
		// Horizontal, because people's names are long and writing them
		// underneath a column turns the chart into a column of sideways
		// words. Down the left they are read top to bottom.
		Horizontal: true,
		Series:     series,
		Ratio:      workloadRatio,
	})
}

// workloadRatio is how much of each person's band the two bars between them
// take. Four fifths, because the gap between the two bars is what says "these
// are two numbers about one person" rather than "these are two people".
const workloadRatio float32 = 0.8

// sequence is 0, 1, 2 … for n, which is the x of a band axis: on a band axis
// a point's x is which category it belongs to rather than a value.
func sequence(n int) []float64 {
	out := make([]float64, n)
	for i := range out {
		out[i] = float64(i)
	}
	return out
}

// PomodoroPhase is which part of a pomodoro is running.
type PomodoroPhase int

const (
	// Focus is the twenty-five minutes of work.
	Focus PomodoroPhase = iota
	// Break is the short rest between two of them.
	Break
)

// PomodoroDefaults are the two lengths a timer starts with. Twenty-five and
// five are the numbers the technique was published with and the ones every
// timer in the world ships, so a caller that wants different ones passes them
// rather than expecting them to have been guessed.
const (
	DefaultFocusMinutes = 25
	DefaultBreakMinutes = 5
)

// PomodoroState is where a timer is, from the two lengths and how long it has
// been going.
type PomodoroState struct {
	// Phase is which part is running.
	Phase PomodoroPhase
	// Left is how long is left in it, never negative.
	Left time.Duration
	// Fraction is how much of the phase is gone, 0 to 1. It is the number
	// that is drawn as an arc, and it is computed rather than left to the
	// caller's arithmetic because the two must agree at the boundaries or
	// the ring and the clock will disagree at exactly the moment somebody
	// is watching both.
	Fraction float32
	// Done reports that the phase has finished.
	Done bool
}

// PomodoroAt is where a timer is after elapsed, with the two lengths in
// minutes. Zero lengths fall back to the defaults, so a caller that has not
// configured the timer still gets a working one rather than an arc that never
// moves.
//
// Elapsed wraps: a timer left running overnight comes back round to the
// start of the cycle rather than sitting at zero, because a stopped-looking
// timer is indistinguishable from a broken one.
func PomodoroAt(elapsed time.Duration, focusMin, breakMin int) PomodoroState {
	if focusMin <= 0 {
		focusMin = DefaultFocusMinutes
	}
	if breakMin <= 0 {
		breakMin = DefaultBreakMinutes
	}
	focus := time.Duration(focusMin) * time.Minute
	brk := time.Duration(breakMin) * time.Minute
	cycle := focus + brk
	if cycle <= 0 || elapsed < 0 {
		elapsed = 0
	}

	// Whole cycles are what the mod strips off, so the phase below is
	// within one cycle and there is exactly one place that knows which half
	// of it we are in.
	//
	// Each branch passes *its own* length as the full one. Passing the focus
	// length for the break as well would put the break's dial at five
	// minutes' worth of a twenty-five minute arc — the ring would be a fifth
	// full at the exact moment the break started.
	at := elapsed % cycle
	if at < focus {
		return pomodoroState(Focus, focus-at, elapsed, focus)
	}
	return pomodoroState(Break, cycle-at, elapsed, brk)
}

// pomodoroState fills in one phase's numbers from a known remaining time.
func pomodoroState(phase PomodoroPhase, left, elapsed, full time.Duration) PomodoroState {
	gone := full - left
	f := float32(gone) / float32(full)
	if f < 0 {
		f = 0
	}
	if f > 1 {
		f = 1
	}
	return PomodoroState{Phase: phase, Left: left, Fraction: f, Done: left <= 0}
}

// PomodoroOptions configure a PomodoroTimer.
type PomodoroOptions struct {
	// Elapsed is how long the timer has been going. It is the caller's, as
	// it is for TimeTracker: the clock belongs to the app, and a widget
	// that owned one would have to be told when the window closed.
	Elapsed time.Duration
	// Running is whether it is going.
	Running *bool
	// Focus and Break are the two lengths in minutes; zero takes the
	// defaults, which are twenty-five and five.
	Focus, Break int
	// Started and Stopped ask for it to be started or stopped.
	Started func()
	Stopped func()
	// Reset asks for the cycle to begin again. It is a separate ask from
	// Stop because "stop" and "start over" are different things to somebody
	// halfway through a rest.
	Reset func()
}

// PomodoroResult carries a PomodoroTimer and where it is.
type PomodoroResult struct {
	// Element is the whole timer.
	Element *ui.Element
	// state is this frame's position in the cycle.
	state PomodoroState
}

// State returns the timer as it is this frame, for a caller that wants the
// numbers rather than the drawing — a menu bar badge, a title, a log line.
func (r PomodoroResult) State() PomodoroState { return r.state }

// PomodoroTimer is a twenty-five minutes of work and five of rest, as a ring
// and a clock.
//
// The ring is chart.Gauge rather than a drawn arc, because a gauge is already
// the component for "a value out of a known range with an arc around it", and
// drawing a second one here would be a second answer to where the needle
// sits when the value is out of range and what happens at the two ends.
//
// The phase is a word as well as a colour — "Focus" and "Break" beside the
// ring — for the same reason the running clock in TimeTracker says "Running":
// a state carried by colour alone is invisible to somebody who cannot see it.
func PomodoroTimer(c *ui.Context, opts PomodoroOptions) PomodoroResult {
	if opts.Running == nil {
		panic("project: PomodoroTimer needs the *bool Running writes to")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	state := PomodoroAt(opts.Elapsed, opts.Focus, opts.Break)
	var r PomodoroResult
	r.state = state

	full := opts.Focus
	if state.Phase == Break {
		full = opts.Break
	}
	if full <= 0 {
		full = DefaultFocusMinutes
		if state.Phase == Break {
			full = DefaultBreakMinutes
		}
	}

	phaseName := core.Msg(c, "project.focus", "Focus")
	sev := core.Accent
	if state.Phase == Break {
		phaseName = core.Msg(c, "project.break", "Break")
		sev = core.Success
	}
	_, fg := sev.Pair(k)

	r.Element = ui.Column(c).FillWidth().Gap(u * 2).AlignItems(ui.Center).
		Label(core.Msg(c, "project.pomodoro", "Timer")).Children(func() {
		// The needle is what is *left*, so the ring drains as the phase runs
		// out rather than filling up. A ring that fills is the shape of a
		// progress bar, and the whole reason this is a dial is that the
		// number beside it counts down.
		//
		// The dial is given the minutes still to run rather than a fraction,
		// because a dial draws its own value and a fraction of a phase draws
		// as "0". Its own label is the phase and nothing more, for the same
		// reason the running clock in TimeTracker says "Running": a state
		// carried by colour is invisible to somebody who cannot see it, and a
		// dial that also wrote the time would say the same thing twice within
		// forty pixels of itself.
		_, ink := sev.Pair(k)
		chart.Gauge(c, chart.GaugeOptions{
			Label:  phaseName,
			Value:  state.Left.Minutes(),
			Min:    0,
			Max:    float64(full),
			Color:  ink,
			Height: u * 40,
		})
		// The clock under the ring rather than inside it: a dial has room
		// for one number, and the one a person reads while working is the
		// one in the form a clock shows. The phase is not written again
		// under it, because the dial above already says it.
		ui.Text(c, FormatElapsed(state.Left)).TextColor(fg).Bold().
			FontSize(core.FontSize(c, theme.StatSize)).Font(monoFamily).SingleLine()

		ui.Row(c).Gap(u * 2).Children(func() {
			if *opts.Running {
				if opts.Stopped != nil && input.Button(c, core.Msg(c, "project.stop", "Stop"),
					input.ButtonOptions{Primary: true}).Clicked() {
					opts.Stopped()
				}
			} else if opts.Started != nil && input.Button(c, core.Msg(c, "project.start", "Start"),
				input.ButtonOptions{Primary: true}).Clicked() {
				opts.Started()
			}
			if opts.Reset != nil && input.Button(c, core.Msg(c, "project.reset", "Reset"),
				input.ButtonOptions{}).Clicked() {
				opts.Reset()
			}
		})
	})
	return r
}
