package agent

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// The small parts a run's window is assembled from: the list of agents, the
// code one of them wrote, and the switch that turns one of them off.

// AgentRow is one agent in a list of agents.
type AgentRow struct {
	// Name is the agent's name. Required: an avatar and a status with no
	// name are a mark and a word about nobody.
	Name string
	// Status is where the agent is, in the words the caller wants — usually
	// a Status's own word, but free to be "2 of 5 steps", for a caller that
	// knows more than the six words. It is a string and not a Status on
	// purpose: a roster is read in one pass, and a row that has to decide
	// which mark to draw is a row that is not done.
	Status string
	// Meta is the figure at the right of the row — an elapsed time, a step
	// count. Empty draws no figure.
	Meta string
}

// AgentRows is the list of agents: one compact row per agent, stacked.
//
// It is the roster, and AgentWellRows is the transcript, and that is the
// whole difference between the two. The well's rows are one run's events in
// the order they happened — steps, calls, commands, questions — each with a
// gutter for its time and its mark, because the well is read as a story. A
// roster's rows are the agents themselves, one line each, read as a list:
// who is there, where each of them is, one figure. A story and a list that
// shared a component would have to choose between a gutter nobody wants on a
// list and a list a story is not.
func AgentRows(c *ui.Context, rows []AgentRow) *ui.Element {
	if len(rows) == 0 {
		panic("agent: AgentRows needs at least one row; an empty roster is a gap")
	}
	for i, r := range rows {
		if r.Name == "" {
			panic("agent: AgentRows row " + itoa(i) +
				" needs a Name; an avatar and a status with no name are about nobody")
		}
	}
	u := core.Density(c).Unit()

	return ui.Column(c).FillWidth().Gap(u * 0.5).Children(func() {
		for _, r := range rows {
			row := ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u*1.5).
				Padding(u*0.75, u).Radius(theme.SmallRadius).Label(r.Name)
			if row.Hovered() {
				row.Background(core.Tokens(c).Surface)
			}
			row.Children(func() {
				display.Avatar(c, r.Name)
				display.Text(c, r.Name, display.TextOptions{Bold: true, MaxLines: 1})
				ui.Box(c).Grow(1)
				if r.Status != "" {
					display.Text(c, r.Status, display.TextOptions{Muted: true, MaxLines: 1})
				}
				if r.Meta != "" {
					display.Text(c, r.Meta, display.TextOptions{Faint: true, MaxLines: 1})
				}
			})
		}
	})
}

// AgentCodeOptions configure an AgentCode.
type AgentCodeOptions struct {
	// Code is what the agent wrote. Required: it is the whole content of the
	// component, and an empty block is a plate.
	Code string
	// Lang is the language, drawn as the tag at the right of the header.
	// Empty draws no tag, which is what a snippet whose language is not
	// known is.
	Lang string
	// Copy draws the copy button beside the tag.
	Copy bool
}

// AgentCode is a small block of code one of the agents wrote: the text,
// monospaced, under a header that names the language and can copy it.
//
// It is display only. A code block in a run's transcript is a thing that
// happened, not a document being edited — so there are no line numbers, no
// folding and no gutter, and the one affordance is the copy button, because
// the reader's next move with a snippet is to take it somewhere else. The
// copy button is input's own, so the tick, the announcement and the
// "copied" rename come from the one place every other copy button in the app
// gets them from.
func AgentCode(c *ui.Context, opts AgentCodeOptions) *ui.Element {
	if opts.Code == "" {
		panic("agent: AgentCode needs Code; an empty code block is a plate")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	code := opts.Code
	return panel(c, layout.ContainerOptions{
		Surface: true, Radius: theme.ControlRadius, Pad: u * 1.5, Gap: u,
	}, func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u).Children(func() {
			if opts.Lang != "" {
				ui.Box(c).Padding(u*0.25, u*1.25).Radius(theme.PillRadius).
					Background(k.Background).Shrink(0).Children(func() {
					ui.Text(c, opts.Lang).TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.CaptionSize))
				})
			}
			ui.Box(c).Grow(1)
			if opts.Copy {
				input.CopyButton(c, &code, input.CopyButtonOptions{
					Plain: true,
					Label: core.Msg(c, "agent.code.copy", core.Def("Copy code")),
				})
			}
		})
		display.Text(c, opts.Code, display.TextOptions{Mono: true})
	})
}

// AgentToggleOptions configure an AgentToggle.
type AgentToggleOptions struct {
	// Name is the agent the switch is about. Required: a switch with no
	// agent beside it is a switch about nothing, and the reader has to
	// guess what it turns off.
	Name string
	// Description is the line under the name — what the agent is for.
	// Empty draws a row with no second line.
	Description string
}

// AgentToggle is the row that turns an agent on or off: the agent's name and
// its description on the left, the switch on the right.
//
// It is a configuration row and not a Switch, and the relationship is the
// same one a checkbox in a settings form has with the checkbox itself: the
// switch is input's, whole — the press, the keyboard, the focus, the
// animation all come from input.Switch — and this component adds only the
// name and the description on the left, and the one fact the bool means
// here, which is "this agent is enabled". A caller that wants the same
// switch about something else writes its own row; what this row owns is the
// reading, not the control.
func AgentToggle(c *ui.Context, enabled *bool, opts AgentToggleOptions) *ui.Element {
	if enabled == nil {
		panic("agent: AgentToggle needs the enabled flag to point at; it keeps none " +
			"of its own")
	}
	if opts.Name == "" {
		panic("agent: AgentToggle needs a Name; a switch with no agent beside it is " +
			"about nothing")
	}
	u := core.Density(c).Unit()

	row := ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u*2).
		Padding(u, u*1.5).Radius(theme.SmallRadius).Label(opts.Name)
	return row.Children(func() {
		display.Avatar(c, opts.Name)
		ui.Column(c).Grow(1).Shrink(1).Gap(u * 0.25).Children(func() {
			display.Text(c, opts.Name, display.TextOptions{Bold: true, MaxLines: 1})
			if opts.Description != "" {
				display.Text(c, opts.Description, display.TextOptions{Muted: true, MaxLines: 1})
			}
		})
		input.Switch(c, enabled, input.SwitchOptions{Label: opts.Name})
	})
}
