package agent

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/code"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/overlay"
	"github.com/HycJack/MintUI/ui/theme"
)

// ── a run asking a person something ─────────────────────────────────────────

// HumanInputRequestResult carries a HumanInputRequest and what was answered.
type HumanInputRequestResult struct {
	// Element is the request.
	Element *ui.Element
	// asked is the answer that was given this frame, empty for none.
	asked string
	sent  bool
	// chosen is the choice that was picked this frame, empty for none.
	chosen string
}

// Asked is the answer that was given this frame, and empty in a frame in which
// nothing was. It is the Draft's contents when the request had a free-text
// field and the choice's value when it had buttons — the answer is one thing
// whichever control it was typed or clicked into, and the caller should not
// have to know which.
func (r HumanInputRequestResult) Asked() string { return r.asked }

// Chosen is the value of the button that was pressed this frame, empty for
// none. It is empty for a request answered in the free-text field, which is
// what distinguishes the two answers.
func (r HumanInputRequestResult) Chosen() string { return r.chosen }

// HumanInputRequestOptions configure a HumanInputRequest.
type HumanInputRequestOptions struct {
	// From is who is asking — the agent's name. It is required: a question
	// with nobody behind it is a wall of text with a question mark in it.
	From string
	// Question is what they want. It is required.
	Question string
	// Why is the reason the run is asking, shown under the question so that
	// the answer can be given without scrolling back through the transcript
	// to find out what prompted it.
	Why string
	// Choices are the answers on offer. With none, a free-text field is drawn
	// instead: a question with no obvious answers is a question, and giving
	// it three buttons would be inventing answers.
	Choices []string
	// Draft is what has been typed into the free-text field, and the caller's.
	Draft *string
	// Chosen is the choice that is picked, and the caller's. It is not drawn
	// as anything other than the pressed button, because the point of asking
	// is that nothing has been picked yet.
	Chosen *string
	// Urgency is the status the request wears. A run that cannot continue
	// until this is answered is Waiting; one that can is Queued, and saying
	// Waiting about the second would train the reader to ignore the first.
	Urgency Status
	// Deadline is when it has to be answered by, already formatted.
	Deadline string
	// FieldLabel names the free-text field, and takes the question's own
	// first line when it is empty — a field with no name is a mystery box.
	FieldLabel string
}

// HumanInputRequest is a run stopping to ask a person something.
//
// It is a card in the transcript rather than a layer on top of the window,
// and that is the whole design decision. A run's questions arrive in the
// middle of work and are part of the story: the answer belongs beside the
// question that caused it, and a dialog would put the question over
// everything and hide the run it came from. The two things in this package
// that are layers — a permission and a tool approval — are layers because
// they stop the run outright, and a run that has stopped is not showing the
// reader anything else.
func HumanInputRequest(c *ui.Context, opts HumanInputRequestOptions) HumanInputRequestResult {
	u := core.Density(c).Unit()
	k := core.Tokens(c)
	if opts.From == "" {
		panic("agent: HumanInputRequest needs a From; a question with nobody behind it is not " +
			"a question anybody can answer")
	}
	if opts.Question == "" {
		panic("agent: HumanInputRequest needs a Question")
	}
	if len(opts.Choices) > 0 && opts.Chosen == nil && opts.Draft == nil {
		panic("agent: HumanInputRequest with choices needs somewhere to put the answer; a run " +
			"that asked a question nobody can answer has stopped for nothing")
	}
	var r HumanInputRequestResult

	card := panel(c, layout.ContainerOptions{
		Surface: true, Radius: theme.CardRadius, Pad: u * 2.5, Gap: u * 1.5,
	}, func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 2).Children(func() {
			display.Avatar(c, opts.From)
			AgentStatus(c, AgentStatusOptions{Status: opts.Urgency, Pill: true})
			ui.Box(c).Grow(1)
			if opts.Deadline != "" {
				display.Text(c, "by "+opts.Deadline, display.TextOptions{Faint: true, MaxLines: 1})
			}
		})
		display.Text(c, opts.Question, display.TextOptions{Bold: true, MaxLines: 3})
		if opts.Why != "" {
			display.Text(c, opts.Why, display.TextOptions{Muted: true, MaxLines: 3})
		}
		if len(opts.Choices) > 0 {
			ui.Row(c).FillWidth().Wrap().GapX(u).GapY(u * 0.75).Children(func() {
				for _, choice := range opts.Choices {
					choice := choice
					picked := opts.Chosen != nil && *opts.Chosen == choice
					btn := input.Button(c, choice, input.ButtonOptions{
						Primary: picked,
						Label:   opts.From + ": " + choice,
					})
					if btn.Clicked() {
						if opts.Chosen != nil {
							*opts.Chosen = choice
						}
						r.sent = true
						r.chosen = choice
						r.asked = choice
					}
				}
			})
		} else if opts.Draft != nil {
			label := opts.FieldLabel
			if label == "" {
				label = opts.Question
			}
			ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.5).Children(func() {
				input.TextInput(c, opts.Draft, input.TextInputOptions{
					Label: label, Placeholder: "Type an answer",
				})
				if input.Button(c, "Send", input.ButtonOptions{Primary: true}).Clicked() {
					r.sent = true
					r.asked = *opts.Draft
				}
			})
		}
	})
	// The card is outlined in its urgency's own ink rather than filled with
	// it: a filled card in the warning colour is a card that shouts over the
	// transcript it sits in, and the one thing that must stand out in a
	// transcript of a hundred lines is a run that has stopped.
	if _, tint := StatusTone(opts.Urgency).Pair(k); tint != k.Text {
		card.Border(theme.BorderWidth*2, tint).Radius(theme.CardRadius)
	}
	card.Label(opts.From + " asks: " + opts.Question)
	r.Element = card
	return r
}

// ── a run asking permission ─────────────────────────────────────────────────

// PermissionRequest is a run asking to do something it may not do.
type PermissionRequest struct {
	// Tool is the tool it wants to use. It is required: permission is granted
	// to somebody doing something, not to a run in general.
	Tool string
	// What is the thing it wants done, as a person would say it — "write to
	// ui/agent/status.go".
	What string
	// Why is the reason it gave.
	Why string
	// Detail is the arguments in full, for a request whose What is a summary.
	Detail string
	// Server names the MCP server the tool comes from.
	Server string
	// Risk is how loud the request is, and decides the panel's own tint.
	Risk core.Severity
	// Scope is how long the answer lasts — "this once", "for this run",
	// "always". It is drawn beside the buttons rather than chosen by them,
	// because the scope of a permission is what the buttons are for.
	Scope string
}

// PermissionPromptResult carries a PermissionPrompt and what was answered.
type PermissionPromptResult struct {
	// Element is the dialog, or nil while it is closed.
	Element         *ui.Element
	allowed, denied bool
}

// Allowed reports that permission was given this frame.
func (r PermissionPromptResult) Allowed() bool { return r.allowed }

// Denied reports that it was refused this frame.
func (r PermissionPromptResult) Denied() bool { return r.denied }

// PermissionPromptOptions configure a PermissionPrompt.
type PermissionPromptOptions struct {
	// Request is what permission is being asked for. Required.
	Request PermissionRequest
	// Open is whether the prompt is showing, and the caller's. It is a layer
	// and so it is modal: the run has stopped, and a window that still took
	// presses while a permission was on screen would let a person start
	// something they cannot finish.
	Open *bool
	// Allow and Deny are where the answers go, and the caller's. Both are
	// written in the frame after the press, not from inside the handler, so
	// that a run told to stop is not also told it was allowed.
	Allow, Deny *bool
	// Remember is whether the answer should be kept, and the caller's.
	Remember *bool
	// Preview draws the same panel in the flow of the page rather than as a
	// layer over it, for the library's own gallery and for a window that
	// documents what a run will ask for. It draws the same panel and answers
	// the same pointers; the only thing it does not do is dim the page.
	Preview bool
	// Width is the dialog's width; zero lets it fit its content.
	Width float32
}

// PermissionPrompt is the modal a run puts up when it wants to do something
// it may not do.
//
// It is modal, and Escape does not answer it. A dialog that Escape dismisses
// is a dialog that can be dismissed without answering, and for a permission
// that would mean a run that asked for the right to do something could simply
// be left alone and carry on. The two ways out are the two buttons, and
// nothing else.
func PermissionPrompt(c *ui.Context, opts PermissionPromptOptions) PermissionPromptResult {
	u := core.Density(c).Unit()
	k := core.Tokens(c)
	req := opts.Request
	if req.Tool == "" && !opts.Preview {
		panic("agent: PermissionPrompt needs a Request with a Tool; permission is granted to " +
			"somebody doing something")
	}
	if req.Tool == "" {
		panic("agent: PermissionPrompt needs a Request with a Tool even in Preview; a preview " +
			"of nothing is nothing")
	}
	if !opts.Preview && opts.Open == nil {
		panic("agent: PermissionPrompt needs the *bool it opens and closes")
	}

	var r PermissionPromptResult
	title := req.Tool + " wants permission"

	// The body and the buttons are built once and used by both paths, so the
	// preview in the gallery is the same panel the window shows and not a
	// second drawing that has to be kept in step with it.
	body := func() {
		ui.Column(c).FillWidth().Gap(u * 1.5).Children(func() {
			ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.5).Children(func() {
				display.Icon(c, display.IconWarning, display.IconOptions{
					Name: req.Tool + " permission", Size: u * 5, Tone: req.Risk,
					Muted: req.Risk == core.Neutral,
				})
				ui.Text(c, req.What).TextColor(k.Text).Bold().
					FontSize(core.FontSize(c, theme.BodySize)).MaxLines(3)
			})
			if req.Server != "" {
				display.Text(c, "via "+req.Server, display.TextOptions{Muted: true, MaxLines: 1})
			}
			if req.Why != "" {
				display.Text(c, req.Why, display.TextOptions{Muted: true, MaxLines: 3})
			}
			if req.Detail != "" {
				layout.ScrollArea(c, layout.ScrollAreaOptions{
					Vertical: true, Height: u * 18, Pad: u,
				}, func() {
					ui.Column(c).FillWidth().Gap(0).Children(func() {
						for _, line := range splitLines(req.Detail) {
							code.CodeMono(c, line, theme.RowSize)
						}
					})
				}).Element.FillWidth()
			}
			if req.Scope != "" {
				ui.Text(c, "Answer applies to: "+req.Scope).TextColor(k.TextFaint).
					FontSize(core.FontSize(c, theme.CaptionSize))
			}
			if opts.Remember != nil {
				// A switch, not a checkbox, and the *bool means "ask again":
				// on by default, so the press is a decision to stop asking.
				// A permission that stops being asked for is the one thing
				// about it a reviewer has to be deliberate about.
				ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.5).Children(func() {
					ui.Text(c, "Ask again next time").TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.RowSize)).Grow(1)
					input.Switch(c, opts.Remember, input.SwitchOptions{})
				})
			}
		})
	}
	actions := func() {
		if input.Button(c, "Deny", input.ButtonOptions{
			Danger: true, Label: "Deny " + req.Tool,
		}).Clicked() {
			if opts.Deny != nil {
				*opts.Deny = true
			}
			if opts.Allow != nil {
				*opts.Allow = false
			}
			if opts.Open != nil {
				*opts.Open = false
			}
			r.denied = true
		}
		if input.Button(c, "Allow", input.ButtonOptions{
			Primary: true, Label: "Allow " + req.Tool,
		}).Clicked() {
			if opts.Allow != nil {
				*opts.Allow = true
			}
			if opts.Deny != nil {
				*opts.Deny = false
			}
			if opts.Open != nil {
				*opts.Open = false
			}
			r.allowed = true
		}
	}

	if opts.Preview {
		// The panel's own element is made inside the frame rather than
		// beside it. An element belongs to whatever was being built when it
		// was created, so a panel made out here would land in the caller's
		// column and leave the frame meant to hold it empty.
		// The frame is given the panel's width rather than the width it is
		// handed: a preview is a picture of a dialog, and a full-width frame
		// with a half-width panel in it looks like a dialog that is mostly
		// empty.
		frame := layout.ContainerOptions{
			Pad: u * 3, Surface: true, Radius: theme.PanelRadius,
		}
		if opts.Width > 0 {
			frame.Width = opts.Width + u*6
		}
		r.Element = layout.Container(c, frame, func() {
			display.Text(c, "shown in the flow, not as a layer",
				display.TextOptions{Faint: true, MaxLines: 1})
			host := ui.Box(c).Shrink(0)
			if opts.Width > 0 {
				host.Width(opts.Width)
			}
			overlay.Panel(c, host, overlay.PanelOptions{
				Title: title, TitleSize: theme.SheetSize,
			}, func() {
				body()
				ui.Row(c).FillWidth().Gap(u*2).Justify(ui.End).Margin(u*2.5, 0, 0, 0).
					Children(actions)
			})
		})
		return r
	}

	r.Element = overlay.Dialog(c, opts.Open, overlay.DialogOptions{
		Title: title, Width: opts.Width, Body: body, Actions: actions,
	})
	return r
}

// ── a run asking to have a tool approved ────────────────────────────────────

// ToolApproval is one tool call waiting to be let through.
type ToolApproval struct {
	// Tool is the tool's name. It is required.
	Tool string
	// Server names the MCP server it comes from, which is often the thing a
	// reviewer is actually being asked about.
	Server string
	// Args are the arguments, as the tool was given them.
	Args string
	// Why is the reason the run gave for making the call.
	Why string
	// Risk is how loud the approval is, and tints the panel.
	Risk core.Severity
	// Choices are the answers from left to right. They are required and the
	// last one is the default — the one Enter takes. A default is not
	// optional here because the alternative is a keyboard that does nothing,
	// and somebody approving forty calls in a row is somebody not reading the
	// fortieth.
	Choices []string
	// SameTool is how many calls to this tool this run has already had. It
	// is what turns "always allow" from a button into a decision: the number
	// is the size of the thing being agreed to.
	SameTool int
}

// ToolApprovalDialogResult carries a ToolApprovalDialog and what was chosen.
type ToolApprovalDialogResult struct {
	// Element is the dialog, or nil while it is closed.
	Element  *ui.Element
	chosen   int
	answered bool
}

// Chosen is the index in ToolApproval.Choices of the button that was pressed
// this frame, and -1 in a frame in which none was.
func (r ToolApprovalDialogResult) Chosen() int {
	if !r.answered {
		return -1
	}
	return r.chosen
}

// ToolApprovalDialogOptions configure a ToolApprovalDialog.
type ToolApprovalDialogOptions struct {
	// Approval is the call waiting. Required.
	Approval ToolApproval
	// Open is whether the dialog is showing, and the caller's.
	Open *bool
	// Chosen is which choice was picked, counted from zero, and the caller's.
	// The dialog writes it and reports it; the run is what acts on it.
	Chosen *int
	// Remember is whether the answer should be kept for the rest of the run,
	// and the caller's.
	Remember *bool
	// Preview draws the same panel in the flow of the page rather than as a
	// layer over it, for the same reason PermissionPrompt has one.
	Preview bool
	// Width is the dialog's width; zero lets it fit its content.
	Width float32
}

// ToolApprovalDialog is the modal that stands between a run and a tool it has
// not been allowed yet.
//
// It differs from a permission in one way that matters: a permission is asked
// once for a thing and a tool approval is asked again and again for the same
// kind of thing. So the dialog shows how many times it has already been
// approved, and its last button is the one that stops asking — because the
// decision somebody is really making on the fortieth call is whether to keep
// being asked at all.
func ToolApprovalDialog(c *ui.Context, opts ToolApprovalDialogOptions) ToolApprovalDialogResult {
	u := core.Density(c).Unit()
	k := core.Tokens(c)
	a := opts.Approval
	if a.Tool == "" {
		panic("agent: ToolApprovalDialog needs a Tool; approval of nothing is not a decision")
	}
	if len(a.Choices) == 0 {
		panic("agent: ToolApprovalDialog needs at least one choice; a dialog with no buttons " +
			"is a wall")
	}
	if !opts.Preview && opts.Open == nil {
		panic("agent: ToolApprovalDialog needs the *bool it opens and closes")
	}
	for i, choice := range a.Choices {
		if choice == "" {
			panic("agent: ToolApprovalDialog choice " + itoa(i) + " is empty")
		}
	}
	var r ToolApprovalDialogResult
	title := "Allow " + a.Tool + "?"

	body := func() {
		ui.Column(c).FillWidth().Gap(u * 1.5).Children(func() {
			ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.5).Children(func() {
				display.Icon(c, display.IconWarning, display.IconOptions{
					Name: a.Tool + " approval", Size: u * 5, Tone: a.Risk,
					Muted: a.Risk == core.Neutral,
				})
				ui.Column(c).Grow(1).Shrink(0).Gap(0).Children(func() {
					ui.Text(c, a.Tool).TextColor(k.Text).Bold().
						FontSize(core.FontSize(c, theme.BodySize))
					if a.Server != "" {
						display.Text(c, "on "+a.Server, display.TextOptions{Muted: true, MaxLines: 1})
					}
				})
				if a.SameTool > 1 {
					display.Tag(c, "call "+itoa(a.SameTool)+" to "+a.Tool,
						display.TagOptions{Tone: core.Warning})
				}
			})
			if a.Why != "" {
				display.Text(c, a.Why, display.TextOptions{Muted: true, MaxLines: 2})
			}
			if a.Args != "" {
				layout.ScrollArea(c, layout.ScrollAreaOptions{
					Vertical: true, Height: u * 16, Pad: u,
				}, func() {
					ui.Column(c).FillWidth().Gap(0).Children(func() {
						for _, line := range splitLines(a.Args) {
							code.CodeMono(c, line, theme.RowSize)
						}
					})
				}).Element.FillWidth()
			}
			if opts.Remember != nil {
				ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.5).Children(func() {
					ui.Text(c, "Don't ask again for this tool").TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.RowSize)).Grow(1)
					input.Switch(c, opts.Remember, input.SwitchOptions{})
				})
			}
		})
	}
	actions := func() {
		for i, choice := range a.Choices {
			i, choice := i, choice
			// The last choice is the affirmative and the only one filled:
			// a row of three filled buttons is three things shouting.
			btn := input.Button(c, choice, input.ButtonOptions{
				Primary: i == len(a.Choices)-1,
				Label:   choice + " " + a.Tool,
			})
			if btn.Clicked() {
				if opts.Chosen != nil {
					*opts.Chosen = i
				}
				if opts.Open != nil {
					*opts.Open = false
				}
				r.answered = true
				r.chosen = i
			}
		}
	}

	if opts.Preview {
		// The panel's own element is made inside the frame rather than
		// beside it. An element belongs to whatever was being built when it
		// was created, so a panel made out here would land in the caller's
		// column and leave the frame meant to hold it empty.
		// The frame is given the panel's width rather than the width it is
		// handed: a preview is a picture of a dialog, and a full-width frame
		// with a half-width panel in it looks like a dialog that is mostly
		// empty.
		frame := layout.ContainerOptions{
			Pad: u * 3, Surface: true, Radius: theme.PanelRadius,
		}
		if opts.Width > 0 {
			frame.Width = opts.Width + u*6
		}
		r.Element = layout.Container(c, frame, func() {
			display.Text(c, "shown in the flow, not as a layer",
				display.TextOptions{Faint: true, MaxLines: 1})
			host := ui.Box(c).Shrink(0)
			if opts.Width > 0 {
				host.Width(opts.Width)
			}
			overlay.Panel(c, host, overlay.PanelOptions{
				Title: title, TitleSize: theme.SheetSize,
			}, func() {
				body()
				ui.Row(c).FillWidth().Gap(u*2).Justify(ui.End).Margin(u*2.5, 0, 0, 0).
					Children(actions)
			})
		})
		return r
	}

	r.Element = overlay.Dialog(c, opts.Open, overlay.DialogOptions{
		Title: title, Width: opts.Width, Body: body, Actions: actions,
	})
	return r
}
