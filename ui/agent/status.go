package agent

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/feedback"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// Status is where a run, a step, a tool call or a server is. It is one type
// for all of them on purpose: a transcript is a column of rows whose status
// words all mean the same six things, and a run of differently-worded statuses
// down one list reads as a list of different kinds of thing.
type Status int

const (
	// StatusQueued is work that is going to happen and has not started. It is
	// the zero value so that a step nobody set a status on is pending rather
	// than done.
	StatusQueued Status = iota
	// StatusRunning is work in progress with no outcome yet.
	StatusRunning
	// StatusWaiting is work stopped on a person: a permission, a question, an
	// approval. It is the one status that is not neutral to look away from,
	// because it is the only one that is waiting on somebody.
	StatusWaiting
	// StatusDone is work that finished as intended.
	StatusDone
	// StatusFailed is work that finished and did not.
	StatusFailed
	// StatusCancelled is work stopped on purpose and will not resume.
	StatusCancelled
)

// String is the word a status is written as. It is the one place the words
// live, so a column of statuses and a status bar say the same thing.
func (s Status) String() string {
	switch s {
	case StatusRunning:
		return "Running"
	case StatusWaiting:
		return "Waiting"
	case StatusDone:
		return "Done"
	case StatusFailed:
		return "Failed"
	case StatusCancelled:
		return "Cancelled"
	case StatusQueued:
		return "Queued"
	}
	panic("agent: unknown Status " + itoa(int(s)))
}

// StatusTone is the severity a status is drawn with.
//
// It is a function and not a field because the mapping is the whole point of
// this file: a status has to mean the same colour wherever it appears — in a
// pill, on a dot, in the bar of a plan, in the margin of a diff — and a
// caller who wrote `core.Success` at four of those sites and `core.Warning`
// at the fifth has built a screen where "done" is sometimes green and
// sometimes amber. One mapping, checked item by item, is the whole defence.
//
// The mapping is:
//
//	Queued    Neutral    nothing is being claimed yet
//	Running   Accent     the system's own highlight, because it is moving
//	Waiting   Warning    the only status that needs a person
//	Done      Success
//	Failed    Danger
//	Cancelled Neutral    stopped on purpose; not good, not bad
//
// Cancelled and Queued are both Neutral on purpose. Cancelled in Danger would
// make every abandoned run look like a failure, and a person who stops a run
// then scans the page for what they just broke would find red.
func StatusTone(s Status) core.Severity {
	switch s {
	case StatusQueued, StatusCancelled:
		return core.Neutral
	case StatusRunning:
		return core.Accent
	case StatusWaiting:
		return core.Warning
	case StatusDone:
		return core.Success
	case StatusFailed:
		return core.Danger
	}
	panic("agent: unknown Status " + itoa(int(s)))
}

// StatusBusy reports whether a status is work happening right now, which is
// the one question any component drawing a mark has to ask: a running step
// moves, a finished one does not.
//
// Waiting is not busy. Nothing is moving while a run waits for a person — it
// is stopped, and a stopped thing drawn as moving says the system is working
// on it when in fact it is waiting to be answered.
func StatusBusy(s Status) bool { return s == StatusRunning }

// StatusIcon is the glyph a status wears where there is no room for a word:
// the mark on a step, the sign on a server row, the chip on a sub-agent.
//
// The names come from the library's own set rather than a private one, so a
// run's marks and the board's marks are one drawing hand.
func StatusIcon(s Status) display.IconName {
	switch s {
	case StatusQueued:
		return display.IconClock
	case StatusRunning:
		return display.IconRefresh
	case StatusWaiting:
		return display.IconBell
	case StatusDone:
		return display.IconCheck
	case StatusFailed:
		return display.IconWarning
	case StatusCancelled:
		return display.IconDismiss
	}
	panic("agent: unknown Status " + itoa(int(s)))
}

// AgentStatusOptions configure an AgentStatus.
type AgentStatusOptions struct {
	// Status is what the mark says. It is required in the sense that the zero
	// value is a real status: a caller who meant to set one and did not gets
	// "Queued", which is what a step nobody has looked at yet should say.
	Status Status
	// Label says what is happening, in the run's own words: "Running tests",
	// "Waiting for your answer". Empty takes the status's own word, so a
	// caller that only wants a dot writes nothing.
	Label string
	// Detail is the second line under the label — how far along, what it is
	// waiting for. Empty draws no second line.
	Detail string
	// Duration is a short figure for how long this has been in its status,
	// set beside the label rather than inside it so that "3m" is not read as
	// part of the sentence.
	Duration string
	// Pill puts the label on the status's own background. Off by default: a
	// status is usually one mark among several, and an unbacked one reads
	// lighter, which is what a column of them wants.
	Pill bool
	// Busy forces the moving mark on or off. Zero takes what the status says:
	// running moves, everything else rests.
	Busy bool
	// BusySet says Busy was meant, so that Busy=false can be asked for.
	BusySet bool
	// Icon puts the status's glyph before the label.
	Icon bool
}

// AgentStatus is a status and what it is about: the mark, the word, and the
// detail under it.
//
// The mark is feedback's status indicator rather than a dot drawn here, so
// that a run's status and a board callback's status are one mark and not two
// that happen to look alike.
func AgentStatus(c *ui.Context, opts AgentStatusOptions) *ui.Element {
	u := core.Density(c).Unit()
	label := opts.Label
	if label == "" {
		label = opts.Status.String()
	}
	busy := StatusBusy(opts.Status)
	if opts.BusySet {
		busy = opts.Busy
	}

	// The row is built inside the column rather than beside it, because an
	// element belongs to whatever was being built when it was made: a row made
	// out here would land in the caller's container and leave the column empty.
	//
	// Neither the row nor the column asks for a width, and that is the whole
	// reason a status can sit at the end of a card's header: a status is
	// almost always the last thing in a row, and a status that filled the row
	// would take all of it and squeeze every word before it to nothing.
	row := func() {
		if opts.Icon {
			tone := StatusTone(opts.Status)
			display.Icon(c, StatusIcon(opts.Status), display.IconOptions{
				Name: label, Size: u * 4,
				// A neutral status has nothing to shout about, and the accent
				// mark beside ordinary ink is the one icon that would.
				Tone: tone, Muted: tone == core.Neutral,
			})
		}
		feedback.StatusIndicator(c, feedback.StatusIndicatorOptions{
			Label:    label,
			Severity: StatusTone(opts.Status),
			Busy:     busy,
			Pill:     opts.Pill,
		})
		if opts.Duration != "" {
			display.Text(c, opts.Duration, display.TextOptions{Muted: true, MaxLines: 1})
		}
	}
	if opts.Detail == "" {
		head := ui.Row(c).AlignItems(ui.Center).Gap(u * 1.5)
		head.Children(row)
		return head
	}
	return ui.Column(c).Gap(u * 0.5).Children(func() {
		ui.Row(c).AlignItems(ui.Center).Gap(u * 1.5).Children(row)
		display.Text(c, opts.Detail, display.TextOptions{Muted: true, MaxLines: 2})
	})
}

// AgentBadge is one small chip on a run's header: a model, a branch, a count.
// It is its own type rather than a display.Tag so that a caller writing a
// header reads as a list of facts rather than as a list of styling calls.
type AgentBadge struct {
	// Label is the words on the chip. It is required.
	Label string
	// Tone is the chip's severity; Neutral is a chip that is only saying a
	// thing, which is most of them.
	Tone core.Severity
}

// AgentPartsOptions configure an AgentParts.
type AgentPartsOptions struct {
	// Name is the run's own name — the agent, or the task. It is required:
	// a header with no subject is a row of chips floating.
	Name string
	// Role is the second line under the name: what this agent is for.
	Role string
	// Status is where the run is.
	Status Status
	// StatusLabel overrides the status's own word, for a run that wants to
	// say what it is doing rather than that it is running.
	StatusLabel string
	// Badges are the chips along the right of the header.
	Badges []AgentBadge
	// Footer sits under the header: elapsed time, a working directory, a
	// hint at what happens next.
	Footer string
}

// AgentParts is a run's header: who it is, where it is, and the few facts
// about it that are worth having without opening anything.
//
// It is the component a window puts at the top of a run, so it is the one
// that has to survive being narrow: the name and the status are never
// dropped, and the chips are what give way — which is why the header wraps
// rather than truncating.
func AgentParts(c *ui.Context, opts AgentPartsOptions) *ui.Element {
	u := core.Density(c).Unit()
	if opts.Name == "" {
		panic("agent: AgentParts needs a Name; a run's header with no subject is a row of chips")
	}
	for i, b := range opts.Badges {
		if b.Label == "" {
			panic("agent: AgentParts badge " + itoa(i) + " has no Label")
		}
	}

	return panel(c, layout.ContainerOptions{
		Surface: true, Radius: theme.CardRadius, Pad: u * 3, Gap: u * 2,
	}, func() {
		ui.Row(c).FillWidth().Wrap().AlignItems(ui.Center).GapX(u * 2).GapY(u).Children(func() {
			display.Avatar(c, opts.Name)
			ui.Column(c).Grow(1).Shrink(0).Gap(u * 0.25).Children(func() {
				display.Text(c, opts.Name, display.TextOptions{Bold: true, MaxLines: 1})
				if opts.Role != "" {
					display.Text(c, opts.Role, display.TextOptions{Muted: true, MaxLines: 1})
				}
			})
			AgentStatus(c, AgentStatusOptions{Status: opts.Status, Label: opts.StatusLabel})
		})
		if len(opts.Badges) > 0 {
			ui.Row(c).FillWidth().Wrap().GapX(u * 2).GapY(u).Children(func() {
				for _, b := range opts.Badges {
					display.Tag(c, b.Label, display.TagOptions{Tone: b.Tone})
				}
			})
		}
		if opts.Footer != "" {
			layout.Divider(c, layout.DividerOptions{})
			display.Text(c, opts.Footer, display.TextOptions{Muted: true, MaxLines: 1})
		}
	})
}

// AgentCaptionOptions configure an AgentCaption.
type AgentCaptionOptions struct {
	// Name is who the row under it is. It is required: an unattributed line
	// in a transcript is a line nobody can attribute.
	Name string
	// Role qualifies the name: "explorer", "reviewer".
	Role string
	// At is when it was said, already formatted.
	At string
	// Tone colours the name; Neutral leaves it in body ink.
	Tone core.Severity
	// Note is a short aside after the name, in the secondary tone.
	Note string
}

// AgentCaption is the quiet line under a piece of an agent's output: who
// produced it and when.
//
// It is small on purpose. A caption in a transcript is the one row nobody
// reads deliberately, and it is the row that decides whether the ones above
// it can be skimmed — so it is one line, it never wraps, and it carries no
// more than the four facts that are needed to place what is above it.
func AgentCaption(c *ui.Context, opts AgentCaptionOptions) *ui.Element {
	u := core.Density(c).Unit()
	if opts.Name == "" {
		panic("agent: AgentCaption needs a Name; a caption with nobody under it is a stray mark")
	}
	k := core.Tokens(c)
	_, fg := opts.Tone.Pair(k)
	if opts.Tone == core.Neutral {
		fg = k.TextMuted
	}

	name := opts.Name
	if opts.Role != "" {
		name += " · " + opts.Role
	}
	return ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.5).Children(func() {
		ui.Text(c, name).TextColor(fg).FontSize(core.FontSize(c, theme.CaptionSize)).
			FontWeight(600).SingleLine()
		if opts.Note != "" {
			ui.Text(c, opts.Note).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine().MaxLines(1)
		}
		ui.Box(c).Grow(1)
		if opts.At != "" {
			ui.Text(c, opts.At).TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
		}
	})
}

// AgentCardResult carries an AgentCard and what was pressed in it.
type AgentCardResult struct {
	// Element is the card.
	Element *ui.Element
	opened  bool
}

// Opened reports a press on the card's own surface this frame.
func (r AgentCardResult) Opened() bool { return r.opened }

// AgentCardOptions configure an AgentCard.
type AgentCardOptions struct {
	// Name is the agent's name. Required.
	Name string
	// Role is what it does.
	Role string
	// Status is where it is.
	Status Status
	// Steps and Tools are the two counts a card of an agent always has, in
	// that order: how much of the plan is done, and how much the agent has
	// touched. Zero of either is left off rather than shown as a zero.
	Steps, Tools int
	// Done is how many of Steps are finished.
	Done int
	// Meter draws Steps as a run of bars under the card's header.
	Meter bool
	// Selected marks the card as the chosen one. The caller's, so that the
	// choice survives the view being rebuilt.
	Selected *bool
	// Footer is one line under the card.
	Footer string
}

// AgentCard is one agent in a list of agents: its name, its role, where it is,
// and how much it has done.
//
// The card is pressable across its whole surface rather than only on a
// "Open" button, because the thing a reader is choosing between two of is the
// agent, not a verb — and a card whose only affordance is a small button in
// the corner is a card nobody presses. The press is reported, never acted
// on: what opening an agent means belongs to the caller.
func AgentCard(c *ui.Context, opts AgentCardOptions) AgentCardResult {
	u := core.Density(c).Unit()
	if opts.Name == "" {
		panic("agent: AgentCard needs a Name")
	}

	var r AgentCardResult
	chosen := opts.Selected != nil && *opts.Selected
	card := panel(c, layout.ContainerOptions{
		Surface: true, Radius: theme.CardRadius, Pad: u * 2.5, Gap: u * 1.5,
	}, func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 2).Children(func() {
			display.Avatar(c, opts.Name)
			ui.Column(c).Grow(1).Shrink(0).Gap(u * 0.25).Children(func() {
				display.Text(c, opts.Name, display.TextOptions{Bold: true, MaxLines: 1})
				if opts.Role != "" {
					display.Text(c, opts.Role, display.TextOptions{Muted: true, MaxLines: 1})
				}
			})
			AgentStatus(c, AgentStatusOptions{Status: opts.Status})
		})
		if opts.Meter && opts.Steps > 0 {
			// The bar is filled to what is finished, not to how many steps
			// there are: a plan with four of ten done reads as four tenths,
			// and the remaining six are the row's own label.
			feedback.Progress(c, feedback.ProgressOptions{
				Label: opts.Name + " progress", Value: float32(opts.Done) / float32(opts.Steps),
				Severity: StatusTone(opts.Status),
			})
		}
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 2).Children(func() {
			if opts.Steps > 0 {
				display.Text(c, itoa(opts.Done)+"/"+itoa(opts.Steps)+" steps",
					display.TextOptions{Muted: true, MaxLines: 1})
			}
			if opts.Tools > 0 {
				display.Text(c, itoa(opts.Tools)+" tool calls",
					display.TextOptions{Muted: true, MaxLines: 1})
			}
			ui.Box(c).Grow(1)
			if opts.Footer != "" {
				display.Text(c, opts.Footer, display.TextOptions{Faint: true, MaxLines: 1})
			}
		})
	})
	if chosen {
		// Selection is a hairline in the accent and nothing else: a column of
		// selected cards filled in accent reads as a column of links, and the
		// reader loses the cards underneath.
		card.Border(theme.BorderWidth*2, core.Tokens(c).Accent).Radius(theme.CardRadius)
	}
	card.Label(opts.Name)
	// Asking whether the card was clicked is also what marks it as taking
	// presses, so the whole surface is the target rather than only the words.
	if card.Clicked() {
		r.opened = true
	}
	r.Element = card
	return r
}
