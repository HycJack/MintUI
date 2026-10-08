package agent

import (
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/code"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/feedback"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// StepMaxDepth is how far a step list nests. Five levels is already a plan
// nobody can hold in their head, and a hard ceiling is what keeps the number
// in the margin from growing until it costs more width than the step's title.
const StepMaxDepth = 4

// Open is the set of things a caller has opened: a step's detail, a
// sub-agent's children, a checkpoint's file list.
//
// It is the caller's, and every component here writes to it rather than
// keeping a copy. That is the whole reason it exists: a component that
// remembered what was open would lose it the moment the view was rebuilt from
// a transcript, and a transcript is rebuilt every time it arrives.
type Open[K comparable] struct {
	items map[K]bool
}

// Has reports whether something is showing what it holds.
func (o *Open[K]) Has(item K) bool { return o.items[item] }

// Open puts something's contents on show.
func (o *Open[K]) Open(item K) {
	if o.items == nil {
		o.items = map[K]bool{}
	}
	o.items[item] = true
}

// Close puts them away again.
func (o *Open[K]) Close(item K) { delete(o.items, item) }

// Set opens or closes something.
func (o *Open[K]) Set(item K, open bool) {
	if open {
		o.Open(item)
		return
	}
	o.Close(item)
}

// Len is how many things are showing their contents.
func (o *Open[K]) Len() int { return len(o.items) }

// Step is one step of a run, or one line of a plan.
type Step struct {
	// Title is what the step is. It is required.
	Title string
	// Detail is the second line: the tool's arguments, the command's output,
	// the reason the step took as long as it took.
	Detail string
	// Status is where the step is.
	Status Status
	// Depth is how far down the plan the step sits: zero for a step of the
	// run, one for a step of the step above it, and so on.
	//
	// It is clamped rather than refused, because a transcript that jumps from
	// a step to a sub-step two levels down has said something about its own
	// shape, and drawing that shape faithfully is more useful than refusing
	// the whole run over it.
	Depth int
	// Tool names the tool a step is a call to, drawn as a chip beside it.
	Tool string
	// Duration is how long the step took, already formatted.
	Duration string
	// Tool is nil for a step that is not a tool call, so nothing is drawn
	// where there is no tool.
}

// StepNumber is the number the first step of a level wears — the shape every
// step at that level shares, with each of its own place still to be filled
// in.
//
// Steps are numbered 1, 2, 3 down the plan and 1.1, 1.2 down the first step's
// own, so how many levels a number has is decided by the depth and what each
// level says is decided by the position. Splitting it that way is what keeps
// the two from disagreeing: the depth is the only thing every step at a level
// shares, and the position is the only thing two steps at a level never
// share. The list fills the shape in through the same stepNumber this is
// written in, which is the only reason the two cannot drift apart.
//
// A step deeper than StepMaxDepth is numbered as if it were at the ceiling,
// and a negative depth as if it were at the floor. Both are a caller whose
// hierarchy has come apart, and clamping draws the shape that is there rather
// than refusing to draw the run.
func StepNumber(depth int) string {
	if depth < 0 {
		depth = 0
	}
	if depth > StepMaxDepth {
		depth = StepMaxDepth
	}
	first := make([]int, depth+1)
	for i := range first {
		first[i] = 1
	}
	return stepNumber(first...)
}

// stepNumber joins a step's ordinal and its parents' into "3.2.1". It is the
// only thing that builds a step number, and StepNumber is written in terms of
// it rather than beside it.
func stepNumber(ordinals ...int) string {
	parts := make([]string, len(ordinals))
	for i, n := range ordinals {
		if n < 0 {
			n = 0
		}
		parts[i] = itoa(n)
	}
	return strings.Join(parts, ".")
}

// clampDepth is a step's depth inside the ceiling.
func clampDepth(d int) int {
	if d < 0 {
		return 0
	}
	if d > StepMaxDepth {
		return StepMaxDepth
	}
	return d
}

// numberSteps is the number each step wears, in the order they were given.
//
// The counters are per level and are reset whenever a step comes back up, so
// 1, 1.1, 1.2, 2, 2.1 rather than 1, 1.1, 1.2, 2, 2.2 — the second step's
// first sub-step is 2.1 because it is the first sub-step of the second step,
// and numbering it 2.2 would make a reader go looking for a missing 2.1.
func numberSteps(steps []Step) []string {
	counters := make([]int, StepMaxDepth+2)
	out := make([]string, len(steps))
	prev := -1
	for i, s := range steps {
		d := clampDepth(s.Depth)
		if d > prev+1 {
			// The step skipped a level, which means no step has ever been at
			// any level between: this one is the first at each of them, and a
			// "0" in a number is not a thing a reader can read.
			for j := prev + 1; j <= d; j++ {
				counters[j] = 1
			}
		} else {
			counters[d]++
		}
		for j := d + 1; j < len(counters); j++ {
			counters[j] = 0
		}
		// The whole path, not just this level's ordinal: a sub-step is
		// numbered inside its parent, and a number that left the parent out
		// would be "1.2" twice over in the same list.
		out[i] = stepNumber(counters[:d+1]...)
		prev = d
	}
	return out
}

// AgentStepListResult carries an AgentStepList and what was pressed in it.
type AgentStepListResult struct {
	// Element is the list.
	Element  *ui.Element
	selected int
	toggled  int
	// answered and unturned report that the two answers landed, so that -1
	// can mean "nothing happened" rather than "the first row" and "the first
	// step". They are separate because a press on a row and a press on a
	// chevron are separate things and either can happen alone.
	answered bool
	unturned bool
}

// Selected is the row pressed this frame, counted from zero, and -1 when none
// was.
func (r AgentStepListResult) Selected() int {
	if !r.answered {
		return -1
	}
	return r.selected
}

// Toggled is the step whose detail was opened or closed this frame, or -1 in
// a frame in which no chevron was pressed. The set itself is the caller's;
// this only says which key changed.
func (r AgentStepListResult) Toggled() int {
	if !r.unturned {
		return -1
	}
	return r.toggled
}

// AgentStepListOptions configure an AgentStepList.
type AgentStepListOptions struct {
	// Steps are the steps, in the order they happened. Their Depth says how
	// the list nests and their order says how it numbers them.
	Steps []Step
	// Open is which steps are showing their detail. Nil closes them all,
	// which is what a transcript arriving in one piece wants.
	Open *Open[int]
	// Selected is the row marked as the one the keys move from, -1 for none.
	Selected *int
	// Numbers puts the step number in the margin. Off is opt-in rather than
	// opt-out because a plan with sub-steps and no numbers is unreadable,
	// while a list of four flat steps reads fine without them.
	Numbers bool
	// Empty draws instead of the steps when there are none. It is required
	// in that case: a list component cannot know what a run with no steps
	// should say, and drawing nothing would read as a run that failed to
	// render.
	Empty string
}

// AgentStepList is a run's steps: numbered, nested by depth, and each one
// openable onto its own detail.
//
// The numbers are computed rather than taken from the caller. A caller
// building a transcript knows the order and the depth of a step and almost
// never knows its number, because the number is a fact about everything
// above it — which means a stored number is a number that is wrong as soon
// as a step is inserted above another.
func AgentStepList(c *ui.Context, opts AgentStepListOptions) AgentStepListResult {
	u := core.Density(c).Unit()
	for i, s := range opts.Steps {
		if s.Title == "" {
			panic("agent: AgentStepList step " + itoa(i) + " has no Title; an unnamed step is a " +
				"gap in the middle of a run")
		}
	}
	if len(opts.Steps) == 0 && opts.Empty == "" {
		panic("agent: AgentStepList with no steps needs opts.Empty; an empty list that says " +
			"nothing is indistinguishable from one that failed to draw")
	}

	numbers := numberSteps(opts.Steps)
	var r AgentStepListResult
	r.selected = -1

	list := ui.Column(c).FillWidth().Gap(u * 0.5).Children(func() {
		for i, s := range opts.Steps {
			i, s := i, s
			stepRow(c, stepRowOptions{
				step:     s,
				number:   numbers[i],
				showNo:   opts.Numbers,
				open:     opts.Open != nil && opts.Open.Has(i),
				selected: opts.Selected != nil && *opts.Selected == i,
			}, func() {
				// A row press only reports: choosing a step is the caller's,
				// because what choosing one means depends on what the window
				// is for.
				r.answered = true
				r.selected = i
			}, func() {
				if opts.Open != nil {
					opts.Open.Set(i, !opts.Open.Has(i))
				}
				r.unturned = true
				r.toggled = i
			})
		}
	})
	if len(opts.Steps) == 0 {
		empty := panel(c, layout.ContainerOptions{
			Surface: true, Radius: theme.CardRadius, Pad: u * 3,
		}, func() {
			display.Text(c, opts.Empty, display.TextOptions{Muted: true})
		})
		r.Element = empty
		return r
	}
	r.Element = list
	return r
}

// stepRowOptions is one row's own settings, kept separate so that AgentPlan
// and AgentStepList agree on how a step is drawn.
type stepRowOptions struct {
	step     Step
	number   string
	showNo   bool
	open     bool
	selected bool
}

// stepRow is one step: its number, its mark, its title, and its detail when it
// is open.
//
// onPress is the row and onToggle is the chevron, and they are separate
// because MyGo hands a press to the innermost control under the pointer: a
// chevron's press never reaches the row, so a row that reported both from one
// Clicked() would open and select in a single gesture.
func stepRow(c *ui.Context, opts stepRowOptions, onPress, onToggle func()) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	s := opts.step
	tone := StatusTone(s.Status)
	_, ink := tone.Pair(k)
	if tone == core.Neutral {
		ink = k.TextFaint
	}

	// The number is drawn in the monospaced face and its own fixed column, so
	// that a list whose first step is "10" and whose second is "1" still has
	// every title starting in the same place.
	indent := u * 4 * float32(clampDepth(s.Depth))
	noCol := k.TextFaint
	if opts.selected {
		noCol = k.Accent
	}

	row := ui.Box(c).FillWidth().Radius(theme.SmallRadius).Children(func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u*1.5).
			Padding(u*0.75, u).Children(func() {
			if opts.showNo {
				no := ui.Text(c, opts.number).TextColor(noCol).Font(code.MonoStack).
					FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0).SingleLine()
				if opts.selected {
					no.Bold()
				}
			}
			// The mark is a dot rather than a glyph: a column of ten steps with
			// ten glyphs in it is a column of icons, and the one difference
			// between them is too small to read down the margin.
			ui.Box(c).Size(u*2.25, u*2.25).Radius(u * 1.25).Shrink(0).
				Background(ink).Label(s.Title + ": " + s.Status.String())
			ui.Column(c).Grow(1).Shrink(0).Gap(u * 0.25).Children(func() {
				ui.Text(c, s.Title).TextColor(k.Text).
					FontSize(core.FontSize(c, theme.RowSize)).MaxLines(2)
				if s.Detail != "" && opts.open {
					ui.Text(c, s.Detail).TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.MetaSize)).MaxLines(4)
				}
			})
			if s.Tool != "" {
				display.Tag(c, s.Tool, display.TagOptions{Tone: core.Neutral})
			}
			if s.Duration != "" {
				ui.Text(c, s.Duration).TextColor(k.TextFaint).
					FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0).SingleLine()
			}
			if s.Detail != "" {
				// The chevron is the one control in the row that changes what
				// is on screen, so it is drawn as a control and named as the
				// step it belongs to rather than as "expand".
				name := toggleName(s.Title, opts.open)
				if display.Icon(c, chevron(opts.open), display.IconOptions{
					Name: name, Size: u * 4, Muted: true,
				}).Clicked() && onToggle != nil {
					onToggle()
				}
			}
		})
	})
	row.Margin(0, 0, 0, indent)
	// The row is named by the step rather than by the mark beside it, so that
	// a test and a screen reader are both naming the same thing a reader
	// would call it.
	row.Label(s.Title)
	if opts.selected {
		// SurfacePressed is the palette's own selection colour — it is what
		// core.Use hands MyGo for Selection — so the chosen row matches a
		// selected row in a system list instead of inventing a second one.
		row.Background(k.SurfacePressed)
	}
	if onPress != nil && row.Clicked() {
		onPress()
	}
}

// toggleName is what a chevron says out loud and what a test clicks: the step
// it belongs to, and what pressing it will do.
func toggleName(title string, open bool) string {
	if open {
		return "Collapse " + title
	}
	return "Expand " + title
}

// chevron is the mark a chevron row wears.
func chevron(open bool) display.IconName {
	if open {
		return display.IconChevronUp
	}
	return display.IconChevron
}

// AgentProgressOptions configure an AgentProgress.
type AgentProgressOptions struct {
	// Status colours the bar and the mark beside it.
	Status Status
	// Done is how much of the work is finished.
	Done int
	// Total is how much there is. Zero or less means the length is not known
	// yet — a plan being written as the run discovers it — and the bar says
	// so by moving rather than by sitting at a percentage that is a guess.
	Total int
	// What names the unit, for "3/7 steps", "2/9 tool calls". It takes part in
	// the bar's name because that is what a screen reader announces.
	What string
	// Detail is a second line under the figure.
	Detail string
	// ShowPercent puts the percentage beside the bar.
	ShowPercent bool
}

// AgentProgress is how far a run has got: the bar, the fraction, and what is
// being counted.
//
// A run whose length is not known draws a bar that moves rather than a bar at
// some percentage, because a bar parked at 30% with nothing behind it reads as
// a stalled job — where a moving one reads as a job. That distinction is
// feedback's, not this component's; what this adds is the fraction and the
// unit, which is the part a reader actually quotes back.
func AgentProgress(c *ui.Context, opts AgentProgressOptions) *ui.Element {
	u := core.Density(c).Unit()
	if opts.Done < 0 {
		panic("agent: AgentProgress cannot have negative work done")
	}
	if opts.Total < 0 {
		panic("agent: AgentProgress cannot have a negative Total; use zero for work of " +
			"unknown length")
	}
	what := opts.What
	if what == "" {
		what = "steps"
	}

	// The bar's own name has to survive being read out loud with nothing
	// visible next to it, so it is a sentence rather than a figure.
	name := opts.What
	if name == "" {
		name = "Progress"
	}
	if opts.Total > 0 {
		name += ": " + itoa(opts.Done) + " of " + itoa(opts.Total)
	} else {
		name += ": " + itoa(opts.Done) + " so far"
	}

	value := float32(-1)
	if opts.Total > 0 {
		value = float32(opts.Done) / float32(opts.Total)
	}

	return panel(c, layout.ContainerOptions{
		Surface: true, Radius: theme.CardRadius, Pad: u * 2.5, Gap: u * 1.5,
	}, func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 2).Children(func() {
			AgentStatus(c, AgentStatusOptions{Status: opts.Status})
			ui.Box(c).Grow(1)
			figure := itoa(opts.Done)
			if opts.Total > 0 {
				figure += " / " + itoa(opts.Total)
			}
			display.Text(c, figure+" "+what, display.TextOptions{Muted: true, MaxLines: 1})
		})
		feedback.Progress(c, feedback.ProgressOptions{
			Label: name, Value: value, Severity: StatusTone(opts.Status),
			ShowPercent: opts.ShowPercent,
		})
		if opts.Detail != "" {
			display.Text(c, opts.Detail, display.TextOptions{Muted: true, MaxLines: 2})
		}
	})
}

// AgentPlanOptions configure an AgentPlan.
type AgentPlanOptions struct {
	// Title heads the plan. It is required: a plan is the one thing in a run
	// that is worth naming, because it is what a person approves before the
	// work starts.
	Title string
	// Goal is the one sentence the plan is for.
	Goal string
	// Steps are the steps, with their Depth saying how the plan nests.
	Steps []Step
	// Open is which steps show their detail.
	Open *Open[int]
	// Selected is the step marked as chosen.
	Selected *int
	// ShowProgress puts a bar over the plan's head. Off is opt-in because a
	// plan that has not started has nothing for a bar to say, and an empty
	// bar over three steps reads as a stalled one.
	ShowProgress bool
	// Empty draws instead of the steps when there are none.
	Empty string
}

// AgentPlan is a plan: its goal, how much of it is done, and the steps.
//
// It is a header plus AgentStepList rather than a variant of it, because the
// two are the same list wearing different amounts: a plan is a list somebody
// is meant to read before agreeing to it, and a step list is a list somebody
// is watching happen. Sharing the row is what stops them from drifting.
func AgentPlan(c *ui.Context, opts AgentPlanOptions) *ui.Element {
	u := core.Density(c).Unit()
	if opts.Title == "" {
		panic("agent: AgentPlan needs a Title; the plan is the one thing in a run that is " +
			"worth naming")
	}
	if len(opts.Steps) == 0 && opts.Empty == "" {
		panic("agent: AgentPlan with no steps needs opts.Empty")
	}

	done := 0
	for _, s := range opts.Steps {
		if s.Status == StatusDone {
			done++
		}
	}

	return panel(c, layout.ContainerOptions{
		Surface: true, Radius: theme.CardRadius, Pad: u * 3, Gap: u * 2,
	}, func() {
		ui.Column(c).FillWidth().Gap(u * 0.5).Children(func() {
			display.Text(c, opts.Title, display.TextOptions{Bold: true})
			if opts.Goal != "" {
				display.Text(c, opts.Goal, display.TextOptions{Muted: true, MaxLines: 2})
			}
		})
		if opts.ShowProgress && len(opts.Steps) > 0 {
			feedback.Progress(c, feedback.ProgressOptions{
				Label:    opts.Title + ": " + itoa(done) + " of " + itoa(len(opts.Steps)) + " steps",
				Value:    float32(done) / float32(len(opts.Steps)),
				Severity: core.Accent,
			})
		}
		layout.Divider(c, layout.DividerOptions{})
		AgentStepList(c, AgentStepListOptions{
			Steps: opts.Steps, Open: opts.Open, Selected: opts.Selected,
			Numbers: true, Empty: opts.Empty,
		})
	})
}
