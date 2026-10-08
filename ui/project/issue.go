package project

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/data"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/feedback"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/theme"
)

// feedbackProgress is a named bar, wrapped so the two call sites in this
// package that want one do not each spell out the library's option struct.
func feedbackProgress(c *ui.Context, label string, value float32, sev core.Severity, height float32) {
	feedback.Progress(c, feedback.ProgressOptions{
		Label:    label,
		Value:    value,
		Severity: sev,
		Height:   height,
	})
}

// monoFamily is the monospaced face, asked for by name because MyGo takes the
// family as a string. It is a package constant rather than a literal at each
// site so that "which face is monospaced here" has one answer.
const monoFamily = "monospace"

// IssueIdBadgeOptions configure an IssueIdBadge.
type IssueIdBadgeOptions struct {
	// Prefix is the tracker — "CB". Empty draws no badge, which is right for
	// a task with no reference rather than a badge reading "-".
	Prefix string
	// Number is the reference's number, zero for none.
	Number int
	// Severity picks the badge's tone. The zero value is core.Neutral,
	// because an ordinary ticket is not a thing that needs attention and a
	// board where every reference is a warning is a board nobody looks at.
	Severity core.Severity
	// Name is what the badge is called out loud, for the case where the
	// tracker prefix alone would not say enough. Empty is the reference.
	Name string
}

// IssueIdBadge is a task's reference as a small mark: "CB-1042".
//
// It is monospaced on purpose. A reference is read by comparing it with
// another — "is this the same one?" — and a proportional face gives CB-1042
// and CB-1742 the same width at a glance, which is exactly the comparison the
// badge exists to make harder to get wrong.
//
// It is also the only element in this package with no word of its own, so it
// must be named: a badge a screen reader cannot call is a shape.
func IssueIdBadge(c *ui.Context, opts IssueIdBadgeOptions) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	ref := FormatIssueID(opts.Prefix, opts.Number)
	if ref == "" {
		// Nothing to show rather than a badge with no words in it. A badge
		// reading "-" is worse than no badge: it looks like a reference
		// whose number somebody forgot.
		return ui.Box(c)
	}
	name := opts.Name
	if name == "" {
		name = ref
	}
	bg, fg := opts.Severity.Pair(k)
	return ui.Box(c).Padding(u*0.5, u*1.25).Radius(theme.PillRadius).
		Background(bg).Label(name).Children(func() {
		ui.Text(c, ref).TextColor(fg).Font(monoFamily).Shrink(0).
			FontSize(core.FontSize(c, theme.CaptionSize))
	})
}

// IssueCardOptions configure an IssueCard.
type IssueCardOptions struct {
	// Task is the card's record.
	Task Task
	// Prefix is the tracker, for the reference.
	Prefix string
	// Body draws under the title, for the comment or the description the
	// caller wants on the card. Nil draws none.
	Body func()
	// Footer draws along the bottom, for the actions. Nil draws none.
	Footer func()
	// Selected marks the card as the open one.
	Selected bool
}

// IssueCard is a task as a panel rather than as a lane's card: the reference,
// the title, whatever the caller puts in the middle, and whatever actions they
// put at the foot.
//
// It is data.Card with the reference put in the header slot, so that the
// panel's radius, its hairline and its header/body/footer arrangement are the
// library's and not a second version of them. A card that is a panel with
// slightly different corners is the first thing that reads as two products.
func IssueCard(c *ui.Context, opts IssueCardOptions) *ui.Element {
	task := opts.Task
	sev := StatusSeverity(task.Status)

	card := data.Card(c, data.CardOptions{
		Title: task.Title,
		Meta:  task.Due,
		Meta2: task.Assignee,
		Header: func() {
			ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(unit(c) * 1.5).Children(func() {
				IssueIdBadge(c, IssueIdBadgeOptions{
					Prefix:   opts.Prefix,
					Number:   task.IDNumber(),
					Severity: sev,
				})
				display.Tag(c, string(task.Status), display.TagOptions{Tone: sev})
				ui.Box(c).Grow(1)
				if task.Priority > 0 {
					PriorityPill(c, task.Priority)
				}
			})
		},
		// The rule only when there is something below it: a card whose
		// footer is one line of metadata does not need one, and a rule with
		// nothing after it is a line drawn for no reason.
		Footer:     opts.Footer,
		FooterRule: opts.Footer != nil,
	}, opts.Body)
	if opts.Selected {
		// A second hairline in the accent, drawn over the card's own. The
		// card is still the library's card — same radius, same border, same
		// padding — and only the outline says this is the open one, which is
		// what keeps a selected card from looking like a different kind of
		// object than the one above it.
		card.BorderWidth(theme.BorderWidth * 2).BorderColor(core.Tokens(c).Accent)
	}
	return card
}

// Subtask is one line of a task's checklist.
type Subtask struct {
	// ID is the subtask's own identity, which is what a toggle is addressed
	// by — a checkbox's value is a bool, so the id is the only thing that
	// says *which* one was ticked.
	ID string
	// Title is what it says.
	Title string
	// Done marks it closed.
	Done bool
	// Assignee is who is on it, empty for nobody.
	Assignee string
}

// SubtaskListOptions configure a SubtaskList.
type SubtaskListOptions struct {
	// Subtasks are the lines, in the caller's order. The list does not sort
	// them: an unchecked item is usually the one somebody is working on and
	// moving it about helps nobody.
	Subtasks []Subtask
	// Toggled writes the id of the line that was ticked or unticked, empty
	// when none was. It is the id rather than the index because a list
	// sorted by somebody else would make an index mean a different line
	// each frame.
	Toggled *string
	// Done, if given, is the count of closed lines and Total the number
	// there are; nil draws no count.
	Done, Total *int
}

// SubtaskListResult carries a SubtaskList and what was done in it.
type SubtaskListResult struct {
	// Element is the whole list.
	Element *ui.Element
	// toggled is the line that was pressed this frame.
	toggled string
}

// Toggled returns the id of the line that was ticked or unticked this frame,
// empty when none was.
func (r SubtaskListResult) Toggled() string { return r.toggled }

// SubtaskList is a task's checklist: a checkbox and a line of text per item,
// with the count above it when there is one.
//
// The checkbox is input.Checkbox, which is the component that gets the three
// states right — unchecked, checked, and the half-done one a subtask list
// never uses but a caller may — and which takes the keyboard. A hand-drawn
// square here would be a second control that a keyboard cannot reach.
//
// Nothing here marks a subtask done. The press is written to Toggled and the
// caller flips the bool and asks again, which is the only way the rest of the
// app can agree about what is closed.
func SubtaskList(c *ui.Context, opts SubtaskListOptions) SubtaskListResult {
	if opts.Toggled == nil {
		panic("project: SubtaskList needs the *string Toggled writes to")
	}
	if (opts.Done == nil) != (opts.Total == nil) {
		panic("project: SubtaskList takes Done and Total together; " +
			"a count of how many without a count of how many is not a count")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	var r SubtaskListResult
	list := ui.Column(c).FillWidth().Gap(u).Role(ui.RoleList).
		Label(core.Msg(c, "project.subtasks", "Subtasks"))
	list.Children(func() {
		if opts.Done != nil {
			ui.Text(c, itoa(*opts.Done)+"/"+itoa(*opts.Total)+" "+core.Msg(c, "project.done", "done")).
				TextColor(k.TextMuted).FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
		}
		for _, s := range opts.Subtasks {
			subtaskRow(c, opts, s, &r.toggled)
		}
	})
	r.Element = list
	return r
}

// subtaskRow is one line of the checklist.
func subtaskRow(c *ui.Context, opts SubtaskListOptions, s Subtask, toggled *string) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	// The row itself, opened here rather than left to the list: the tick, the
	// line of text and the avatar are three parts of one line, and a column
	// that never opens a row draws a checklist as a stack of stacks.
	ui.Row(c).FillWidth().Gap(u).AlignItems(ui.Center).Children(func() {
		// The checkbox takes a CheckState and not a bool, so the done flag is
		// translated into one and back. That is a two-line cost for using the
		// control that already has the three states and the keyboard.
		state := input.Unchecked
		if s.Done {
			state = input.Checked
		}
		// The checkbox is given the line's name as its label and not as text
		// to draw: it draws a label of its own, so passing s.Title as both
		// would put every subtask on screen twice, and the two copies would
		// disagree about whether the line is closed. It is also wrapped in a
		// shrink-to-fit box, because a checkbox fills the row it is in and a
		// row is not what a checklist line is.
		var box *ui.Element
		ui.Box(c).Shrink(0).Children(func() {
			box = input.Checkbox(c, &state, "", input.CheckboxOptions{Label: s.Title})
		})
		if box.Clicked() {
			*toggled = s.ID
		}
		// A checked line is drawn in the faint tone rather than struck
		// through: a rule through the words fights with the tick beside them,
		// and the tone already says the line is closed.
		ink := k.Text
		if s.Done {
			ink = k.TextFaint
		}
		ui.Text(c, s.Title).TextColor(ink).Grow(1).Shrink(0).
			FontSize(core.FontSize(c, theme.RowSize)).SingleLine()
		if s.Assignee != "" {
			ui.Box(c).Shrink(0).Children(func() {
				display.Avatar(c, s.Assignee)
			})
		}
	})
}

// MilestoneOptions configure a MilestoneProgress.
type MilestoneProgressOptions struct {
	// Title is what the milestone is called.
	Title string
	// Done and Total are its progress.
	Done, Total int
	// Due is when it is wanted, already formatted. Empty draws no line under
	// the bar rather than an empty one.
	Due string
	// Height is the bar's height; zero is the library's own.
	Height float32
}

// MilestoneProgressResult carries a MilestoneProgress.
type MilestoneProgressResult struct {
	// Element is the whole thing.
	Element *ui.Element
}

// MilestoneProgress is how much of a milestone is done, as a bar with its
// count beside it.
//
// The tone is MilestoneTone's, so the last tenth is the accent rather than a
// warning: a milestone at ninety per cent is not in trouble, it is nearly
// finished, and colouring it as though it were would make every late
// milestone in the plan look identical.
func MilestoneProgress(c *ui.Context, opts MilestoneProgressOptions) MilestoneProgressResult {
	if opts.Title == "" {
		panic("project: MilestoneProgress needs a Title; a bar with no name " +
			"reaches a screen reader as \"progress bar\"")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	var r MilestoneProgressResult
	r.Element = ui.Column(c).FillWidth().Gap(u * 1.5).Label(opts.Title).Children(func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 2).Children(func() {
			ui.Text(c, opts.Title).TextColor(k.Text).Grow(1).
				FontSize(core.FontSize(c, theme.RowSize)).SingleLine()
			ui.Text(c, itoa(opts.Done)+" / "+itoa(opts.Total)).TextColor(k.TextMuted).Shrink(0).
				FontSize(core.FontSize(c, theme.MetaSize)).SingleLine()
		})
		feedbackProgress(c, opts.Title, MilestoneFraction(opts.Done, opts.Total),
			MilestoneTone(opts.Done, opts.Total), opts.Height)
		if opts.Due != "" {
			ui.Text(c, opts.Due).TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
		}
	})
	return r
}

// LabelManagerOptions configure a LabelManager.
type LabelManagerOptions struct {
	// Labels are the caller's label names, in the caller's order.
	Labels []string
	// Toggled writes the label that was pressed, empty when none was. It is
	// the label and not a bool because the set of chosen ones is the
	// caller's: this component draws the row and asks.
	Toggled *string
	// Name is the caller's string for a new label's name, and empty shows
	// the mint row. Adding one is asked for separately so a read-only list
	// needs no field.
	Name *string
	// Added asks for a label to be created; nil draws no mint row.
	Added func(name string)
}

// LabelManagerResult carries a LabelManager and what was done in it.
type LabelManagerResult struct {
	// Element is the whole manager.
	Element *ui.Element
	// toggled and added are this frame's answers.
	toggled, added string
}

// Toggled returns the label pressed this frame, empty when none was.
func (r LabelManagerResult) Toggled() string { return r.toggled }

// Added returns the name typed into the mint row at the moment it was pressed.
func (r LabelManagerResult) Added() string { return r.added }

// LabelManager is the set of labels a board knows, as chips, with a way to add
// one.
//
// The chips are input.ChoiceChips rather than a row of boxes drawn here:
// they wrap, they refuse to go past their cap, and a pressed chip at the cap
// leaves its face unchanged so that a person can see it did not take. Those
// are three behaviours, and reimplementing them is three chances to be
// subtly worse than the control that already has them.
//
// It is not a MultiSelect. A label is a short name with no description, and
// MultiSelect carries Select all and Clear — which is right for a form
// choosing from two hundred customers and is clutter on a row of eight
// labels.
func LabelManager(c *ui.Context, opts LabelManagerOptions) LabelManagerResult {
	if opts.Toggled == nil {
		panic("project: LabelManager needs the *string Toggled writes to")
	}
	if opts.Added != nil && opts.Name == nil {
		panic("project: a mint row needs the *string Name writes to; a label " +
			"held only here has no name once the frame is over")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	var r LabelManagerResult
	mgr := ui.Column(c).FillWidth().Gap(u * 2).
		Label(core.Msg(c, "project.labels", "Labels"))
	mgr.Children(func() {
		if len(opts.Labels) == 0 {
			ui.Text(c, core.Msg(c, "project.noLabels", "No labels yet")).
				TextColor(k.TextFaint).FontSize(core.FontSize(c, theme.RowSize)).SingleLine()
		} else {
			chosen := []string{}
			input.ChoiceChips(c, &chosen, choicesOf(opts.Labels),
				input.ChoiceChipsOptions{
					Label: core.Msg(c, "project.labels", "Labels"),
					Wrap:  true,
				})
			if len(chosen) > 0 {
				r.toggled = chosen[0]
			}
		}
		if opts.Added != nil {
			ui.Row(c).FillWidth().Gap(u * 2).AlignItems(ui.End).Children(func() {
				ui.Box(c).Grow(1).Shrink(0).Children(func() {
					input.TextInput(c, opts.Name, input.TextInputOptions{
						Label:       core.Msg(c, "project.newLabel", "New label"),
						Placeholder: core.Msg(c, "project.newLabelHint", "bug, backend…"),
					})
				})
				mint := input.Button(c, core.Msg(c, "project.addLabel", "Add"),
					input.ButtonOptions{Disabled: *opts.Name == ""})
				if mint.Clicked() && *opts.Name != "" {
					r.added = *opts.Name
					opts.Added(*opts.Name)
				}
			})
		}
	})
	r.Element = mgr
	return r
}

// choicesOf turns a list of names into the choices a chip group takes.
func choicesOf(names []string) []input.Choice {
	out := make([]input.Choice, 0, len(names))
	for _, n := range names {
		out = append(out, input.Choice{Value: n, Label: n})
	}
	return out
}

// unit is the density's spacing step. The handful of helpers in this package
// that only need a gap say it once here rather than spelling the whole
// expression at each site.
func unit(c *ui.Context) float32 { return core.Density(c).Unit() }
