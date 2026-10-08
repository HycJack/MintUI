package agent

import (
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/code"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// ToolCall is one thing a run asked a tool to do.
type ToolCall struct {
	// Tool is the tool's name — "Read", "Grep", "Bash". It is required: a
	// call with no tool is a call to nothing.
	Tool string
	// Summary is one line saying what it was asked for, in the caller's
	// words: the path it was given, the pattern it was given.
	Summary string
	// Status is where the call is.
	Status Status
	// Duration is how long it took, already formatted.
	Duration string
	// Result is a short figure for what came back: "18 matches", "2.1 kB".
	Result string
}

// ToolCallCardResult carries a ToolCallCard and what was pressed in it.
type ToolCallCardResult struct {
	// Element is the card.
	Element *ui.Element
	toggled bool
}

// Toggled reports that the detail was opened or closed this frame. The
// *bool itself is the caller's; this only says that it changed.
func (r ToolCallCardResult) Toggled() bool { return r.toggled }

// ToolCallCardOptions configure a ToolCallCard.
type ToolCallCardOptions struct {
	// Call is the call. Required.
	Call ToolCall
	// Server names the MCP server a tool came from, drawn as a chip. A tool
	// that is built into the agent has none and none is drawn.
	Server string
	// Detail is what the call was given and what it said back — the
	// arguments as JSON, the output, the error. Empty draws a card with
	// nothing to open, which is the right shape for a call with a one-line
	// summary and nothing behind it.
	Detail string
	// Open is whether the detail is showing, and the caller's.
	Open *bool
	// Height is the detail's viewport height, and is required when there is
	// a detail: without one the detail grows to fit and a long tool result
	// takes the whole window with it.
	Height float32
	// Empty draws instead of the card when there is no call.
	Empty string
}

// ToolCallCard is one tool call: which tool, asked for what, and what it
// said back.
//
// The detail is the caller's to open rather than the card's to decide,
// because a run that is being read wants them closed and a run that is being
// debugged wants them open, and only the window knows which of those it is.
func ToolCallCard(c *ui.Context, opts ToolCallCardOptions) ToolCallCardResult {
	u := core.Density(c).Unit()
	if opts.Call.Tool == "" && opts.Empty == "" {
		panic("agent: ToolCallCard needs a call; a card with no tool in it is a call to nothing")
	}
	if opts.Detail != "" && opts.Height <= 0 {
		panic("agent: ToolCallCard with a Detail needs a Height; without one a long tool " +
			"result takes the window with it")
	}
	if opts.Call.Tool == "" {
		return ToolCallCardResult{Element: emptyCard(c, opts.Empty)}
	}
	open := opts.Open != nil && *opts.Open
	var r ToolCallCardResult

	call := opts.Call
	card := panel(c, layout.ContainerOptions{
		Surface: true, Radius: theme.CardRadius, Pad: u * 2, Gap: u * 1.25,
	}, func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.5).Children(func() {
			ui.Text(c, call.Tool).TextColor(accent(c)).
				Font(code.MonoStack).FontSize(core.FontSize(c, theme.RowSize)).
				Bold().Shrink(0).SingleLine()
			if call.Summary != "" {
				ui.Text(c, call.Summary).TextColor(core.Tokens(c).Text).
					FontSize(core.FontSize(c, theme.RowSize)).MaxLines(1)
			}
			ui.Box(c).Grow(1)
			if opts.Server != "" {
				display.Tag(c, opts.Server, display.TagOptions{Tone: core.Neutral})
			}
			if call.Result != "" {
				display.Text(c, call.Result, display.TextOptions{Muted: true, MaxLines: 1})
			}
			AgentStatus(c, AgentStatusOptions{Status: call.Status, Busy: StatusBusy(call.Status), BusySet: true})
			if call.Duration != "" {
				display.Text(c, call.Duration, display.TextOptions{Faint: true, MaxLines: 1})
			}
			if opts.Detail != "" {
				name := "Expand " + call.Tool
				if open {
					name = "Collapse " + call.Tool
				}
				if display.Icon(c, chevron(open), display.IconOptions{
					Name: name, Size: u * 4, Muted: true,
				}).Clicked() {
					if opts.Open != nil {
						*opts.Open = !*opts.Open
					}
					r.toggled = true
				}
			}
		})
		if opts.Detail != "" && open {
			layout.Divider(c, layout.DividerOptions{})
			// A box rather than a panel: the detail is the tool's own output
			// and should read as output, in the monospaced face, without the
			// level column and the filter a build's output needs.
			layout.ScrollArea(c, layout.ScrollAreaOptions{
				Vertical: true, Height: detailHeight(c, opts.Height), Pad: u,
			}, func() {
				ui.Column(c).FillWidth().Gap(0).Children(func() {
					for _, line := range strings.Split(opts.Detail, "\n") {
						code.CodeMono(c, line, theme.RowSize)
					}
				})
			}).Element.FillWidth()
		}
	})
	card.Label(call.Tool + " call")
	r.Element = card
	return r
}

// detailHeight rounds the viewport down to a whole number of code lines.
// A viewport that stops mid-line shows the top sliver of the next glyph
// above its own edge, which reads as a drawing mistake rather than as
// somewhere to scroll. The pad is part of the budget on both ends.
func detailHeight(c *ui.Context, h float32) float32 {
	line := code.LineHeight(c)
	pad := core.Density(c).Unit()
	if line <= 0 {
		return h
	}
	inner := h - pad*2
	lines := int(inner / line)
	if lines < 1 {
		return h
	}
	return float32(lines)*line + pad*2
}

// emptyCard is what a component draws when it has nothing to show and the
// caller has said what that nothing should say.
func emptyCard(c *ui.Context, text string) *ui.Element {
	u := core.Density(c).Unit()
	return panel(c, layout.ContainerOptions{
		Surface: true, Radius: theme.CardRadius, Pad: u * 2.5,
	}, func() {
		display.Text(c, text, display.TextOptions{Muted: true, MaxLines: 2})
	})
}

// ToolCallGroupResult carries a ToolCallGroup and what was pressed in it.
type ToolCallGroupResult struct {
	// Element is the group.
	Element *ui.Element
	toggled bool
	// selected is the call pressed this frame, or -1.
	selected int
	answered bool
}

// Toggled reports that the group was opened or closed this frame.
func (r ToolCallGroupResult) Toggled() bool { return r.toggled }

// Selected is the call pressed this frame, counted from zero, and -1 for
// none.
func (r ToolCallGroupResult) Selected() int {
	if !r.answered {
		return -1
	}
	return r.selected
}

// ToolCallGroupOptions configure a ToolCallGroup.
type ToolCallGroupOptions struct {
	// Tool is the tool the calls are to. It is required.
	Tool string
	// Calls are the calls, in the order they were made.
	Calls []ToolCall
	// Open is whether the calls are listed. It is the caller's, because a
	// run's own opinion about which calls are worth seeing changes faster
	// than a component's could.
	Open *bool
	// Height is how many calls the open group shows at once, and is
	// required when open: a run that made two hundred calls cannot draw
	// them all into the middle of a transcript.
	Height float32
	// Empty draws instead of the group when there are no calls.
	Empty string
}

// ToolCallGroup is the same tool called more than once, gathered up.
//
// A run that reads twelve files produces twelve Read cards, and twelve cards
// of equal weight bury the one line that was not a file read. The group is
// the fix: the runs collapse into one row carrying their counts and their
// worst status, and open into the individual calls only when somebody asks.
//
// The collapsed row reports the worst status of the calls inside it rather
// than the most recent one, so a group whose last call failed does not read
// as a group that finished quietly.
func ToolCallGroup(c *ui.Context, opts ToolCallGroupOptions) ToolCallGroupResult {
	u := core.Density(c).Unit()
	if opts.Tool == "" && opts.Empty == "" {
		panic("agent: ToolCallGroup needs a Tool; a group of nothing cannot be counted")
	}
	if opts.Tool == "" {
		return ToolCallGroupResult{selected: -1, Element: emptyCard(c, opts.Empty)}
	}
	if len(opts.Calls) == 0 && opts.Empty == "" {
		panic("agent: ToolCallGroup with no calls needs opts.Empty")
	}
	open := opts.Open != nil && *opts.Open
	var r ToolCallGroupResult
	r.selected = -1

	if len(opts.Calls) == 0 {
		r.Element = emptyCard(c, opts.Empty)
		return r
	}

	statuses := make([]Status, 0, len(opts.Calls))
	failed, running := 0, 0
	for _, call := range opts.Calls {
		statuses = append(statuses, call.Status)
		if call.Status == StatusFailed {
			failed++
		}
		if call.Status == StatusRunning {
			running++
		}
	}
	worst := WorstStatus(statuses)

	group := panel(c, layout.ContainerOptions{
		Surface: true, Radius: theme.CardRadius, Pad: u * 2, Gap: u * 1.25,
	}, func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.5).Children(func() {
			if display.Icon(c, chevron(open), display.IconOptions{
				Name: groupToggleName(opts.Tool, open), Size: u * 4, Muted: true,
			}).Clicked() {
				if opts.Open != nil {
					*opts.Open = !*opts.Open
				}
				r.toggled = true
			}
			ui.Text(c, opts.Tool).TextColor(accent(c)).Font(code.MonoStack).
				FontSize(core.FontSize(c, theme.RowSize)).Bold().Shrink(0).SingleLine()
			display.Text(c, itoa(len(opts.Calls))+" calls", display.TextOptions{Muted: true, MaxLines: 1})
			if failed > 0 {
				display.Tag(c, itoa(failed)+" failed", display.TagOptions{Tone: core.Danger})
			}
			ui.Box(c).Grow(1)
			AgentStatus(c, AgentStatusOptions{Status: worst, Busy: running > 0, BusySet: true})
		})
	})
	if open {
		group.Children(func() {
			layout.Divider(c, layout.DividerOptions{})
			layout.ScrollArea(c, layout.ScrollAreaOptions{
				Vertical: true, Height: opts.Height,
			}, func() {
				ui.Column(c).FillWidth().Gap(u * 0.5).Children(func() {
					for i, call := range opts.Calls {
						i, call := i, call
						callGroupRow(c, call, i == len(opts.Calls)-1, func() {
							r.answered = true
							r.selected = i
						})
					}
				})
			}).Element.FillWidth()
		})
	}
	group.Label(opts.Tool + " calls, " + itoa(len(opts.Calls)))
	r.Element = group
	return r
}

// statusRank is how loud a status is when a run of several has to be
// summarised as one. It is a ranking rather than a severity so that "needs a
// person" and "in flight" can sit in the order a reader cares about: a
// failure first, then the thing that is waiting, then the thing that is still
// moving, and the two that have stopped — one on purpose, one finished — at
// the bottom, where a group of them can share one quiet mark.
func statusRank(s Status) int {
	switch s {
	case StatusFailed:
		return 5
	case StatusWaiting:
		return 4
	case StatusRunning:
		return 3
	case StatusCancelled:
		return 2
	case StatusQueued:
		return 1
	}
	return 0
}

// WorstStatus is the status a run of several is summarised as: the loudest
// of them.
//
// It exists because the alternative is a caller picking one, and the two
// places that would have picked — a group's collapsed row and a run's header
// — would have picked differently. The rule it encodes is that a summary must
// never be quieter than the worst thing inside it: a group whose last call
// failed shown as "done" is a lie a reader cannot check.
func WorstStatus(statuses []Status) Status {
	worst := StatusDone
	for _, s := range statuses {
		if statusRank(s) > statusRank(worst) {
			worst = s
		}
	}
	return worst
}

// groupToggleName is what a group's chevron says out loud.
func groupToggleName(tool string, open bool) string {
	if open {
		return "Collapse " + tool + " calls"
	}
	return "Expand " + tool + " calls"
}

// callGroupRow is one call inside an open group: its status, its summary, its
// figure. The last row is marked so that a group of eight does not read as
// eight separate cards.
func callGroupRow(c *ui.Context, call ToolCall, last bool, onPress func()) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	_, ink := StatusTone(call.Status).Pair(k)
	if StatusTone(call.Status) == core.Neutral {
		ink = k.TextFaint
	}
	row := ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u*1.5).
		Padding(u*0.5, u).Children(func() {
		ui.Box(c).Size(u*2, u*2).Radius(u).Shrink(0).Background(ink).
			Label(call.Tool + ": " + call.Status.String())
		display.Text(c, call.Summary, display.TextOptions{Muted: true, MaxLines: 1})
		ui.Box(c).Grow(1)
		if call.Result != "" {
			display.Text(c, call.Result, display.TextOptions{Faint: true, MaxLines: 1})
		}
		if call.Duration != "" {
			display.Text(c, call.Duration, display.TextOptions{Faint: true, MaxLines: 1})
		}
	})
	row.Label(call.Tool + ": " + call.Summary)
	if !last {
		layout.Divider(c, layout.DividerOptions{})
	}
	if onPress != nil && row.Clicked() {
		onPress()
	}
}

// CommandExecution is one command a run ran.
type CommandExecution struct {
	// Command is the line as it was typed, shell prompt and all. It is
	// required: a command card without the command is a card about a
	// mystery.
	Command string
	// Dir is the working directory it ran in.
	Dir string
	// Status is where it is.
	Status Status
	// ExitCode is what it exited with, and is only read once it has.
	ExitCode int
	// Duration is how long it ran for.
	Duration string
	// Output is what it printed.
	Output []string
	// Truncated is how many lines of output were dropped from the end. It is
	// shown rather than the lines: a caller with a thousand lines of output
	// passes the last few hundred and says how many it left out, because a
	// silent truncation reads as a command that printed nothing more.
	Truncated int
}

// CommandExecutionCardResult carries a CommandExecutionCard and what was
// pressed in it.
type CommandExecutionCardResult struct {
	// Element is the card.
	Element *ui.Element
	toggled bool
	// selected is the output row pressed this frame, or -1.
	selected int
	answered bool
}

// Toggled reports that the output was opened or closed this frame.
func (r CommandExecutionCardResult) Toggled() bool { return r.toggled }

// Selected is the output row pressed this frame, counted from zero, and -1 for
// none.
func (r CommandExecutionCardResult) Selected() int {
	if !r.answered {
		return -1
	}
	return r.selected
}

// CommandExecutionCardOptions configure a CommandExecutionCard.
type CommandExecutionCardOptions struct {
	// Exec is the command. Required.
	Exec CommandExecution
	// Open is whether the output is showing, and the caller's.
	Open *bool
	// Height is the output's viewport height, and is required when there is
	// output: without one a test run's output takes the window with it.
	Height float32
	// Scroll is where the output is scrolled to, and the caller's.
	Scroll *ui.ScrollState
	// Empty draws instead of the card when there is no command.
	Empty string
}

// CommandExecutionCard is a command, its exit code, and what it printed.
//
// The exit code is a chip rather than a line of text because it is the one
// thing about a command a reader looks for first and it is also the one
// thing a transcript of a successful run would otherwise never mention. Its
// severity comes from the status, not from the number: a command that exited
// 1 inside a script that means to is not a failure, and only the caller knows
// which it was.
func CommandExecutionCard(c *ui.Context, opts CommandExecutionCardOptions) CommandExecutionCardResult {
	u := core.Density(c).Unit()
	if opts.Exec.Command == "" && opts.Empty == "" {
		panic("agent: CommandExecutionCard needs a Command")
	}
	if len(opts.Exec.Output) > 0 && opts.Height <= 0 {
		panic("agent: CommandExecutionCard with output needs a Height for it")
	}
	if opts.Exec.Command == "" {
		return CommandExecutionCardResult{selected: -1, Element: emptyCard(c, opts.Empty)}
	}
	open := opts.Open != nil && *opts.Open
	exec := opts.Exec
	var r CommandExecutionCardResult
	r.selected = -1

	card := panel(c, layout.ContainerOptions{
		Surface: true, Radius: theme.CardRadius, Pad: u * 2, Gap: u * 1.25,
	}, func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.5).Children(func() {
			display.Icon(c, display.IconPanel, display.IconOptions{
				Name: "Terminal", Size: u * 4, Muted: true,
			})
			ui.Text(c, exec.Command).TextColor(core.Tokens(c).Text).Font(code.MonoStack).
				FontSize(core.FontSize(c, theme.RowSize)).MaxLines(2)
			ui.Box(c).Grow(1)
			if len(exec.Output) > 0 {
				name := "Expand output"
				if open {
					name = "Collapse output"
				}
				if display.Icon(c, chevron(open), display.IconOptions{
					Name: name, Size: u * 4, Muted: true,
				}).Clicked() {
					if opts.Open != nil {
						*opts.Open = !*opts.Open
					}
					r.toggled = true
				}
			}
		})
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.5).Children(func() {
			if exec.Dir != "" {
				display.Text(c, exec.Dir, display.TextOptions{Faint: true, Mono: true, MaxLines: 1})
			}
			ui.Box(c).Grow(1)
			if exec.Truncated > 0 {
				display.Text(c, itoa(exec.Truncated)+" lines not shown",
					display.TextOptions{Faint: true, MaxLines: 1})
			}
			if exec.Duration != "" {
				display.Text(c, exec.Duration, display.TextOptions{Faint: true, MaxLines: 1})
			}
			if exec.Status == StatusDone || exec.Status == StatusFailed {
				display.Tag(c, "exit "+itoa(exec.ExitCode),
					display.TagOptions{Tone: StatusTone(exec.Status)})
			}
			AgentStatus(c, AgentStatusOptions{
				Status: exec.Status, Busy: StatusBusy(exec.Status), BusySet: true,
			})
		})
	})
	if open && len(exec.Output) > 0 {
		card.Children(func() {
			layout.Divider(c, layout.DividerOptions{})
			// The output is drawn here rather than through a log panel: a log
			// panel carries a level column and a filter, and a command's
			// output has neither — everything it printed came out at the same
			// level, and a header of its own under a card that has already
			// said "exit 0" is a second answer to a question already
			// answered.
			layout.ScrollArea(c, layout.ScrollAreaOptions{
				Vertical: true, Height: opts.Height,
			}, func() {
				ui.Column(c).FillWidth().Gap(0).Children(func() {
					for i, line := range exec.Output {
						i, line := i, line
						row := ui.Box(c).FillWidth().Label(line)
						row.Children(func() {
							code.CodeMono(c, line, theme.RowSize)
						})
						if row.Clicked() {
							r.answered = true
							r.selected = i
						}
					}
				})
			}).Element.FillWidth()
		})
	}
	card.Label("Command: " + exec.Command)
	r.Element = card
	return r
}

// FileChangeCardResult carries a FileChangeCard and what was pressed in it.
type FileChangeCardResult struct {
	// Element is the card.
	Element *ui.Element
	pressed bool
}

// Pressed reports a press on the card this frame.
func (r FileChangeCardResult) Pressed() bool { return r.pressed }

// FileChangeCardOptions configure a FileChangeCard.
type FileChangeCardOptions struct {
	// Path is the file that changed. It is required.
	Path string
	// Change is what happened to it.
	Change FileStatus
	// Added and Removed are the line counts. Negative ones are ignored: a
	// file cannot have lost a negative number of lines, and a caller that
	// computed one has a bug that showing "-3 added" would hide.
	Added, Removed int
	// Bar draws the two counts as a run of bars, which is what makes a list
	// of files scannable — the sizes are comparable down a column, where
	// the digits are not.
	Bar bool
	// Selected marks the card as the chosen one, and the caller's.
	Selected *bool
	// From and To are the two revisions, for a card that says what it was
	// moved between.
	From, To string
	// Footer is one line under the card.
	Footer string
}

// FileChangeCard is one file a run touched: its path, what happened to it,
// and how much of it changed.
//
// The bar is drawn from the two counts rather than from a total, so a file
// that lost forty lines and gained two is a short red bar and a sliver of
// green — which is the true shape of that change, and a total would have
// hidden it behind forty-two.
func FileChangeCard(c *ui.Context, opts FileChangeCardOptions) FileChangeCardResult {
	u := core.Density(c).Unit()
	if opts.Path == "" {
		panic("agent: FileChangeCard needs a Path")
	}
	if opts.Added < 0 || opts.Removed < 0 {
		panic("agent: FileChangeCard cannot have negative line counts")
	}
	added, removed := opts.Added, opts.Removed
	k := core.Tokens(c)
	var r FileChangeCardResult

	card := panel(c, layout.ContainerOptions{
		Surface: true, Radius: theme.CardRadius, Pad: u * 1.75, Gap: u * 1,
	}, func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.5).Children(func() {
			display.Icon(c, FileIcon(opts.Change), display.IconOptions{
				Name: opts.Path + ": " + opts.Change.String(), Size: u * 4,
				Tone: FileTone(opts.Change), Muted: FileTone(opts.Change) == core.Neutral,
			})
			ui.Text(c, opts.Path).TextColor(k.Text).Font(code.MonoStack).
				FontSize(core.FontSize(c, theme.RowSize)).MaxLines(1).Grow(1)
			display.Tag(c, opts.Change.String(), display.TagOptions{Tone: FileTone(opts.Change)})
			if added > 0 {
				ui.Text(c, signed(added)).TextColor(k.Success).
					Font(code.MonoStack).FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
			}
			if removed > 0 {
				ui.Text(c, signed(-removed)).TextColor(k.Danger).
					Font(code.MonoStack).FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
			}
		})
		if opts.Bar && added+removed > 0 {
			// The two halves share one bar's width in proportion to their
			// sizes, which is what makes a list of files comparable: the
			// reader is looking at shapes, not at digits. The track is drawn
			// behind them and the row is clipped, so the two fills come out
			// as one bar with rounded ends rather than as two rectangles.
			total := float32(added + removed)
			ui.Row(c).FillWidth().Height(u * 2.5).Radius(u * 1.25).Clip().Shrink(0).
				Background(k.SurfaceHover).Children(func() {
				if added > 0 {
					ui.Box(c).WidthPercent(float32(added) / total * 100).FillHeight().
						Background(k.Success).Shrink(0)
				}
				if removed > 0 {
					ui.Box(c).WidthPercent(float32(removed) / total * 100).FillHeight().
						Background(k.Danger).Shrink(0)
				}
			})
		}
		if opts.From != "" || opts.To != "" || opts.Footer != "" {
			ui.Text(c, join(" · ", revisionLine(opts.From, opts.To), opts.Footer)).
				TextColor(k.TextFaint).FontSize(core.FontSize(c, theme.CaptionSize)).MaxLines(1)
		}
	})
	card.Label(opts.Path + ", " + opts.Change.String() + ", " + signed(added) + " " + signed(-removed))
	if opts.Selected != nil && *opts.Selected {
		card.Background(k.SurfacePressed)
	}
	if card.Clicked() {
		r.pressed = true
	}
	r.Element = card
	return r
}

// FileIcon is the glyph a file's change wears where a path is all there is
// room for.
//
// Renamed gets the external-arrow rather than borrowing a status: a move is
// neither an arrival nor a departure, and a glyph that says either would be
// wrong in the one case where the reader most needs to be told apart from a
// delete.
func FileIcon(f FileStatus) display.IconName {
	switch f {
	case FileAdded:
		return display.IconPlus
	case FileRemoved:
		return display.IconTrash
	case FileRenamed:
		return display.IconExternal
	case FileModified:
		return display.IconEdit
	}
	panic("agent: unknown FileStatus " + itoa(int(f)))
}

// revisionLine is the two revisions a card was moved between, or nothing.
func revisionLine(from, to string) string {
	switch {
	case from != "" && to != "":
		return from + " → " + to
	case to != "":
		return "→ " + to
	case from != "":
		return from + " →"
	}
	return ""
}
