package code

import (
	"strconv"
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// The panels a development window wears around a viewer: the debugger's four
// lists, the build's output, the problems a compiler found, the outline of a
// file, the find bar, the completion menu and the flame graph.
//
// All of them are lists of rows, all of the rows are the same height, and all
// of them are built in a ui.List so that a list of ten thousand breakpoints
// draws the ten it has room for. They are here together for that reason and
// for no other.

// ── BreakpointList ──────────────────────────────────────────────────────────

// Breakpoint is one line somebody has asked the debugger to stop at.
type Breakpoint struct {
	// File is the path; empty means the file the window is showing, which is
	// what a caller with one file open means by it.
	File string
	// Line is the line's number from one.
	Line int
	// Condition stops there only when it is true, and is empty for always.
	Condition string
	// Hit is how many times it has been reached, which is how a reader tells
	// a breakpoint that has not fired from one that is being hit every time.
	Hit int
	// Enabled is that it is live. A disabled breakpoint is kept and does not
	// stop, which is different from having none: the condition is the most
	// likely reason somebody turned it off and not wanted to retype it.
	Enabled bool
}

// BreakpointOptions configure a BreakpointList.
type BreakpointOptions struct {
	// Name is what the panel is called, and is required.
	Name string
	// Points are the breakpoints, and they are the caller's slice: which lines
	// a debugger is stopping at is state the debugger has, and a copy of it
	// here would be a second answer to a question with one right answer.
	Points []Breakpoint
	// Height is the viewport's own height, and is required.
	Height float32
	// Width is the panel's own width.
	Width float32
	// Selected is which breakpoint the caller has chosen, counted from zero,
	// and -1 for none.
	Selected int
	// State is where the list is among its rows.
	State *ui.ListState
}

// BreakpointResult carries a BreakpointList and what was done in it.
type BreakpointResult struct {
	// Element is the list.
	Element *ui.Element
	// toggled, enabled and removed are what was done this frame, each counted
	// from zero, and -1 for none. They are separate because they are different
	// acts: a reader who clicks a disabled breakpoint expects it back on, and
	// one number saying "something happened" would leave the caller guessing.
	toggled, enabled, removed int
	// selected is the row pressed this frame.
	selected int
}

// Toggled is the breakpoint pressed this frame, and -1 for none.
func (r BreakpointResult) Toggled() int { return r.toggled }

// Enabled is the breakpoint switched on or off this frame, and -1 for none.
func (r BreakpointResult) Enabled() int { return r.enabled }

// Removed is the breakpoint asked to be removed this frame, and -1 for none.
func (r BreakpointResult) Removed() int { return r.removed }

// Selected is the row pressed this frame, and -1 for none.
func (r BreakpointResult) Selected() int { return r.selected }

// BreakpointList is the debugger's list of the lines it will stop at.
//
// A disabled breakpoint is shown struck through rather than greyed: the point
// of keeping one is that it can be turned back on, and a row that looks like
// everything else in the panel cannot be told apart from a live one at a
// glance.
func BreakpointList(c *ui.Context, opts BreakpointOptions) BreakpointResult {
	if opts.Name == "" {
		panic("code: BreakpointList needs a Name; a list of line numbers that says nothing " +
			"about which file they are in cannot be acted on")
	}
	if opts.Height <= 0 {
		panic("code: BreakpointList needs a Height; without one it draws every breakpoint")
	}
	var r BreakpointResult
	r.toggled, r.enabled, r.removed, r.selected = -1, -1, -1, -1
	// The body is a function, not an element, so that CodePanel can draw the
	// header above it. A panel whose name is printed under its contents reads
	// as a footer, and every panel in this package would have had one.
	view := func() {
		scroll := ui.Scroll(c).Height(opts.Height).FillWidth()
		scroll.Children(func() {
			ui.List(c, opts.State, len(opts.Points), func(i int) {
				bp := opts.Points[i]
				at := i
				row := breakpointRow(c, bp, at == opts.Selected)
				row.Children(func() {
					if row.Clicked() {
						r.selected = at
						r.toggled = at
					}
				})
				row.Children(func() {
					off := panelToggle(c, "Enable this breakpoint")
					if off.Clicked() {
						opts.Points[at].Enabled = !opts.Points[at].Enabled
						r.enabled = at
					}
				})
				row.Children(func() {
					del := panelToggle(c, core.Msg(c, "code.removeBreakpoint", core.Def("Remove this breakpoint")))
					if del.Clicked() {
						r.removed = at
					}
				})
			}).FillWidth().Role(ui.RoleNone)
		})
	}

	host := CodePanel(c, view, layoutContainer(opts.Width), func() {
		CodeHeader(c, opts.Name, nil, func() {
			CodeBadge(c, itoa(len(opts.Points))+" breakpoints", core.Neutral)
		})
	})

	r.Element = host
	return r
}

func breakpointRow(c *ui.Context, bp Breakpoint, chosen bool) *ui.Element {
	k := core.Tokens(c)
	row := ui.Row(c).FillWidth().Shrink(0).Gap(0).AlignItems(ui.Center).
		Role(ui.RoleListItem).Label("Breakpoint at line " + itoa(bp.Line))
	if chosen {
		row.Background(k.Surface)
	}
	row.Children(func() {
		// The mark in front is the breakpoint's own state, and it is a mark
		// and not a word: a filled circle for a live one and a hollow one for
		// a disabled one is read in a list without being read at all.
		side := theme.CaptionSize
		dot := ui.Box(c).Size(side, side).Shrink(0).Radius(side / 2)
		if bp.Enabled {
			dot.Background(k.Danger)
		} else {
			dot.BorderWidth(theme.BorderWidth).BorderColor(k.TextFaint)
		}
		if bp.File != "" {
			dot.Label("Breakpoint at " + bp.File + " line " + itoa(bp.Line))
		}
		ui.Box(c).Grow(1)
		where := "line " + itoa(bp.Line)
		if bp.File != "" {
			where = bp.File + ":" + itoa(bp.Line)
		}
		ui.Box(c).Grow(1).Children(func() {
			// A disabled breakpoint is struck through rather than greyed: the
			// reason to keep one is that it can be turned back on, and a row
			// that looks like every other row in the panel cannot be told
			// apart from a live one at a glance. The text claims no height:
			// the box is a column, and grown into height it does not have
			// the line lays out at zero tall and paints over whatever the
			// row puts beside it.
			text := ui.Text(c, where).Font(MonoStack).SingleLine().Shrink(1).
				FontSize(core.FontSize(c, theme.RowSize)).TextColor(k.Text)
			if !bp.Enabled {
				text.Strikethrough().TextColor(k.TextFaint)
			}
		})
		if bp.Condition != "" {
			CodeBadge(c, bp.Condition, core.Warning)
		}
		if bp.Hit > 0 {
			CodeBadge(c, itoa(bp.Hit)+"×", core.Neutral)
		}
	})
	return row
}

// panelToggle is one of the small square buttons at the end of a panel's row.
// It is one function because four panels have them and three of them had
// started to draw theirs differently.
func panelToggle(c *ui.Context, name string) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	e := ui.Box(c).Size(u*6, u*6).Shrink(0).Radius(theme.SmallRadius).
		Role(ui.RoleButton).Label(name).Tooltip(name).Cursor(ui.CursorPointer)
	e.Children(func() {
		ui.Box(c).Size(u*2, u*2).Shrink(0).Radius(u).Background(k.TextFaint)
	})
	return e
}

// ── CallStack ───────────────────────────────────────────────────────────────

// Frame is one entry of a call stack.
type Frame struct {
	// Name is the function's name.
	Name string
	// File and Line are where it is, and Line is from one.
	File string
	Line int
	// Where is the relation to the one below — "here", "caller", "3 frames
	// up" — which is what a debugger shows and what a plain list of names
	// does not.
	Where string
}

// CallStackOptions configure a CallStack.
type CallStackOptions struct {
	// Name is what the panel is called, and is required.
	Name string
	// Frames are the stack, innermost first.
	Frames []Frame
	// Selected is which frame the debugger is stopped in, counted from zero.
	Selected int
	// Height and Width are the panel's own size; the height is required.
	Height float32
	Width  float32
	State  *ui.ListState
}

// CallStackResult carries a CallStack and what was chosen in it.
type CallStackResult struct {
	// Element is the list.
	Element *ui.Element
	// selected is the frame pressed this frame, counted from zero, or -1.
	selected int
}

// Selected is the frame pressed this frame, and -1 for none. It is the
// caller's to act on — which frame a debugger shows is the debugger's own
// business — but it is reported so that a caller does not have to put a press
// handler on every row to find out.
func (r CallStackResult) Selected() int { return r.selected }

// CallStack is the debugger's stack: the function that stopped, and what it
// was called from.
//
// The frames are shown innermost first, which is the order the stack is in
// and the reverse of the order a reader reads a program: the one thing they
// came to look at is at the top and not at the bottom.
func CallStack(c *ui.Context, opts CallStackOptions) CallStackResult {
	if opts.Name == "" {
		panic("code: CallStack needs a Name; a list of function names that says nothing about " +
			"which ones they are in cannot be read out")
	}
	if opts.Height <= 0 {
		panic("code: CallStack needs a Height; without one it draws every frame")
	}
	var r CallStackResult
	r.selected = -1
	// The body is a function, not an element, so that CodePanel can draw the
	// header above it.
	view := func() {
		scroll := ui.Scroll(c).Height(opts.Height).FillWidth()
		scroll.Children(func() {
			ui.List(c, opts.State, len(opts.Frames), func(i int) {
				f := opts.Frames[i]
				at := i
				row := callStackRow(c, f, at == opts.Selected)
				row.Children(func() {
					if row.Clicked() {
						r.selected = at
					}
				})
			}).FillWidth().Role(ui.RoleNone)
		})
	}

	host := CodePanel(c, view, layoutContainer(opts.Width), func() {
		CodeHeader(c, opts.Name, nil, func() {
			CodeBadge(c, itoa(len(opts.Frames))+" frames", core.Neutral)
		})
	})

	r.Element = host
	return r
}

func callStackRow(c *ui.Context, f Frame, chosen bool) *ui.Element {
	k := core.Tokens(c)
	row := ui.Row(c).FillWidth().Shrink(0).Gap(0).AlignItems(ui.Center).
		Role(ui.RoleListItem).Label(f.Name)
	if chosen {
		row.Background(k.Surface)
	}
	row.Children(func() {
		CodeBadge(c, itoa(f.Line), core.Neutral)
		ui.Text(c, f.Name).Font(MonoStack).SingleLine().Shrink(1).Grow(1).
			FontSize(core.FontSize(c, theme.RowSize)).TextColor(k.Text)
		if f.Where != "" {
			CodeCaption(c, f.Where)
		}
	})
	return row
}

// ── VariablesPanel ──────────────────────────────────────────────────────────

// Variable is one thing in scope.
type Variable struct {
	// Name is what it is called and Value is what it holds, as the caller's
	// own formatting: a debugger that printed a number as "0x1f" where the
	// caller wanted "31" would be a debugger with an opinion about types.
	Name, Value string
	// Expanded is that a tree below it is open, which is what a reader wants
	// to see rather than a chevron on everything.
	Expanded bool
	// Children are the fields of a struct or the entries of a map.
	Children []Variable
}

// VariablesOptions configure a VariablesPanel.
type VariablesOptions struct {
	// Name is what the panel is called, and is required.
	Name string
	// Variables are the things in scope, in the order they should be read.
	Variables []Variable
	// Height and Width are the panel's own size; the height is required.
	Height float32
	Width  float32
	State  *ui.ListState
}

// VariablesResult carries a VariablesPanel and what was pressed in it.
type VariablesResult struct {
	// Element is the panel.
	Element *ui.Element
	// pressed is the variable pressed this frame, as a path of indices from
	// the root, and nil for none. A path rather than an index because the
	// thing a caller wants is the variable, and a variable inside a struct is
	// not at the top level however long the panel is.
	pressed []int
}

// Pressed is the path to the variable pressed this frame — the index into the
// root's variables, then the index into that one's children, and so on — and
// nil for none.
func (r VariablesResult) Pressed() []int { return r.pressed }

// VariablesPanel is what is in scope where the debugger has stopped, with the
// things that have fields of their own able to be opened.
//
// It is a tree drawn as a list rather than as ui.Tree because every row in it
// has to be the same height as a row of a call stack beside it, and a tree
// that sizes its rows to their content cannot promise that.
func VariablesPanel(c *ui.Context, opts VariablesOptions) VariablesResult {
	if opts.Name == "" {
		panic("code: VariablesPanel needs a Name; a list of names and values that says nothing " +
			"about where they are from cannot be read out")
	}
	if opts.Height <= 0 {
		panic("code: VariablesPanel needs a Height; without one it draws every variable")
	}
	var r VariablesResult
	flat := flattenVariables(opts.Variables)
	// The body is a function, not an element, so that CodePanel can draw the
	// header above it.
	view := func() {
		scroll := ui.Scroll(c).Height(opts.Height).FillWidth()
		scroll.Children(func() {
			ui.List(c, opts.State, len(flat), func(i int) {
				v := flat[i]
				row := variableRow(c, v)
				row.Children(func() {
					if row.Clicked() {
						r.pressed = v.path
					}
				})
			}).FillWidth().Role(ui.RoleNone)
		})
	}

	host := CodePanel(c, view, layoutContainer(opts.Width), func() {
		CodeHeader(c, opts.Name, nil, func() {
			CodeBadge(c, itoa(len(flat))+" in scope", core.Neutral)
		})
	})

	r.Element = host
	return r
}

// flatVariable is one row of a VariablesPanel: the variable, how deep it is,
// and where it came from.
type flatVariable struct {
	Variable
	depth int
	path  []int
}

// flattenVariables is the open parts of a tree as the rows to draw.
//
// Only the open ones, because a struct with a thousand fields has a thousand
// rows and a reader looking at two of them: building the closed ones would
// cost the frame a thousand rows for a tree that is two deep.
func flattenVariables(all []Variable) []flatVariable {
	var out []flatVariable
	var walk func(vars []Variable, depth int, path []int)
	walk = func(vars []Variable, depth int, path []int) {
		for i := range vars {
			at := append(append([]int{}, path...), i)
			out = append(out, flatVariable{Variable: vars[i], depth: depth, path: at})
			if vars[i].Expanded {
				walk(vars[i].Children, depth+1, at)
			}
		}
	}
	walk(all, 0, nil)
	return out
}

func variableRow(c *ui.Context, v flatVariable) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	row := ui.Row(c).FillWidth().Shrink(0).AlignItems(ui.Center).Role(ui.RoleListItem).
		Label(v.Name)
	row.Children(func() {
		// The indent is the depth, drawn as padding rather than as a margin
		// on the name: a margin moves the name and a padding moves both, so a
		// deep value stays in the column its shallow neighbours are in.
		ui.Box(c).Width(float32(v.depth) * u * 2).Shrink(0)
		CodeSelectMark(c, false)
		ui.Text(c, v.Name).Font(MonoStack).SingleLine().Shrink(0).
			FontSize(core.FontSize(c, theme.RowSize)).TextColor(k.Text)
		CodeCaption(c, ":")
		ui.Text(c, v.Value).Font(MonoStack).SingleLine().Shrink(1).Grow(1).
			FontSize(core.FontSize(c, theme.RowSize)).TextColor(k.AccentText)
		if len(v.Children) > 0 {
			CodeBadge(c, itoa(len(v.Children)), core.Neutral)
		}
	})
	return row
}

// ── ProcessList ─────────────────────────────────────────────────────────────

// Process is one thing running.
type Process struct {
	// Name is what it is called.
	Name string
	// Args are its arguments, which is what tells two of the same name apart.
	Args []string
	// Pid is its process id, as the caller's own number, and it is a string
	// because a pid is not always a number: a container's is a name.
	Pid string
	// CPU and Memory are how much of the machine it is using, as the caller
	// formatted them, so that the panel does not decide between "12%" and
	// "0.12" for a reader of either.
	CPU, Memory string
	// Running is that it is going, and is false for one that has stopped.
	Running bool
}

// ProcessOptions configure a ProcessList.
type ProcessOptions struct {
	// Name is what the panel is called, and is required.
	Name string
	// Processes are what is running, and they are the caller's slice: whether
	// a process is still there is the caller's business, not the list's.
	Processes []Process
	// Height and Width are the panel's own size; the height is required.
	Height float32
	Width  float32
	// Running is the filter: zero for all of them, and a bit set of the
	// states to keep. It is a mask because "running but not ours" is a thing
	// people ask for and two bools is the start of a third option.
	Running uint8
	// State is where the list is among its rows.
	State *ui.ListState
	// Selected is the row chosen, counted from zero.
	Selected int
}

// ProcessResult carries a ProcessList and what was done in it.
type ProcessResult struct {
	// Element is the list.
	Element *ui.Element
	// selected is the row pressed this frame, and -1 for none.
	selected int
	// stopped is the process asked to be stopped this frame, and -1 for
	// none.
	stopped int
}

// Selected is the row pressed this frame, and -1 for none.
func (r ProcessResult) Selected() int { return r.selected }

// Stopped is the process asked to be stopped this frame, and -1 for none. It
// is asked for rather than done: what happens to a process is the caller's,
// and the only thing the list knows is that somebody pressed the button.
func (r ProcessResult) Stopped() int { return r.stopped }

// ProcessList is what is running, with what it is using and a way to stop it.
//
// The CPU and memory are the caller's own strings, drawn as given, because
// the panel is a list and not a sampler: it has no idea what a process's CPU
// time means numerically and inventing a bar from a percentage it was handed
// would be a bar of the wrong thing.
func ProcessList(c *ui.Context, opts ProcessOptions) ProcessResult {
	if opts.Name == "" {
		panic("code: ProcessList needs a Name; a list of process names that says nothing about " +
			"where they run cannot be acted on")
	}
	if opts.Height <= 0 {
		panic("code: ProcessList needs a Height; without one it draws every process")
	}

	shown := make([]Process, 0, len(opts.Processes))
	for _, p := range opts.Processes {
		if opts.Running != 0 && processMask(p) != opts.Running {
			continue
		}
		shown = append(shown, p)
	}

	var r ProcessResult
	r.selected, r.stopped = -1, -1
	// The body is a function, not an element, so that CodePanel can
	// draw the header above it.
	view := func() {
		view := ui.Scroll(c).Height(opts.Height).FillWidth()
		view.Children(func() {
			ui.List(c, opts.State, len(shown), func(i int) {
				p := shown[i]
				at := i
				row := processRow(c, p, at == opts.Selected)
				row.Children(func() {
					if row.Clicked() {
						r.selected = at
					}
				})
				row.Children(func() {
					stop := panelToggle(c, core.Msg(c, "code.stopProcess", core.Def("Stop this process")))
					if stop.Clicked() {
						r.stopped = at
					}
				})
			}).FillWidth().Role(ui.RoleNone)
		})

	}

	host := CodePanel(c, view, layoutContainer(opts.Width), func() {
		CodeHeader(c, opts.Name, nil, func() {
			CodeBadge(c, itoa(len(shown))+" processes", core.Neutral)
		})
	})

	r.Element = host
	return r
}

// processMask is the state a process is in as a bit, so that the filter is a
// mask and not a pair of booleans.
func processMask(p Process) uint8 {
	if p.Running {
		return 1
	}
	return 2
}

func processRow(c *ui.Context, p Process, chosen bool) *ui.Element {
	k := core.Tokens(c)
	row := ui.Row(c).FillWidth().Shrink(0).Gap(0).AlignItems(ui.Center).
		Role(ui.RoleListItem).Label(p.Name)
	if chosen {
		row.Background(k.Surface)
	}
	row.Children(func() {
		CodeSelectMark(c, p.Running)
		ui.Text(c, p.Name).Font(MonoStack).SingleLine().Shrink(0).
			FontSize(core.FontSize(c, theme.RowSize)).TextColor(k.Text)
		if len(p.Args) > 0 {
			ui.Text(c, strings.Join(p.Args, " ")).Font(MonoStack).SingleLine().
				Shrink(1).Grow(1).FontSize(core.FontSize(c, theme.CaptionSize)).
				TextColor(k.TextMuted)
		}
		if p.CPU != "" {
			CodeBadge(c, p.CPU, cpuSeverity(p.CPU))
		}
		if p.Memory != "" {
			CodeBadge(c, p.Memory, core.Neutral)
		}
		if p.Pid != "" {
			CodeCaption(c, p.Pid)
		}
	})
	return row
}

// cpuSeverity is how a CPU figure is drawn, from the number the caller put in
// the string.
//
// It is read rather than asked for because the figure is a caller's string: a
// caller that formats "12.5" and a caller that formats "0.125" have both been
// right, and a panel that assumed one of them would paint a process that is
// idle as one that is pegged.
func cpuSeverity(s string) core.Severity {
	v, err := strconv.ParseFloat(strings.TrimSuffix(strings.TrimSpace(s), "%"), 64)
	if err != nil {
		return core.Neutral
	}
	switch {
	case v >= 90:
		return core.Danger
	case v >= 60:
		return core.Warning
	}
	return core.Neutral
}

// ── CommandHistory ──────────────────────────────────────────────────────────

// CommandOptions configure a CommandHistory.
type CommandHistoryOptions struct {
	// Name is what the panel is called, and is required.
	Name string
	// Commands are the ones that have been run, oldest first.
	Commands []string
	// At is which one the field is showing, counted from zero, and -1 for
	// none — which is the state before anything has been typed, and is
	// different from showing the last one.
	At int
	// Height and Width are the panel's own size; the height is required.
	Height float32
	Width  float32
	State  *ui.ListState
	// Chosen is the command pressed this frame, counted from zero, or -1.
	Chosen int
	// Repeatable says the arrows walk the history, which is the point of
	// having it and is off for a list of what somebody ran rather than
	// something they can run again.
	Repeatable bool
}

// CommandHistoryResult carries a CommandHistory and what was asked for.
type CommandHistoryResult struct {
	// Element is the panel.
	Element *ui.Element
	// chosen is the command pressed this frame, counted from zero, or -1.
	chosen int
	// stepped is how far the history moved this frame — one for the up arrow,
	// minus one for the down — and zero for anything else.
	stepped int
}

// Chosen is the command pressed this frame, and -1 for none.
func (r CommandHistoryResult) Chosen() int { return r.chosen }

// Stepped is how far the history moved this frame, and zero for none. It is
// signed, because walking back and walking forwards are different places to
// end up and a caller that only had a count would have to work out which from
// its own idea of where it was.
func (r CommandHistoryResult) Stepped() int { return r.stepped }

// CommandHistory is the commands somebody has run, with the ones that failed
// marked.
//
// The newest is at the bottom, so that the last thing they did is the last
// thing they see — which is the opposite of a shell's own history, and
// deliberate: a panel a person reads is read top to bottom, and a shell's
// order exists so that the up arrow walks backwards through it.
func CommandHistory(c *ui.Context, opts CommandHistoryOptions) CommandHistoryResult {
	if opts.Name == "" {
		panic("code: CommandHistory needs a Name; a list of commands that says nothing about " +
			"where they were run cannot be read out")
	}
	if opts.Height <= 0 {
		panic("code: CommandHistory needs a Height; without one it draws every command")
	}

	var r CommandHistoryResult
	r.chosen = -1
	k, u := core.Tokens(c), core.Density(c).Unit()

	// The body is a function, not an element, so that CodePanel can
	// draw the header above it.
	view := func() {
		view := ui.Scroll(c).Height(opts.Height).FillWidth()
		view.Children(func() {
			ui.List(c, opts.State, len(opts.Commands), func(i int) {
				cmd := opts.Commands[i]
				at := i
				row := ui.Row(c).FillWidth().Shrink(0).AlignItems(ui.Center).
					Role(ui.RoleListItem).Label(cmd)
				if at == opts.At {
					row.Background(k.Surface)
				}
				row.Children(func() {
					// The prompt mark is in the window's faint ink and is not a
					// character: a shell prompt is a thing the reader recognises
					// by its shape, and a literal "$" in a panel of commands is a
					// dollar sign.
					ui.Box(c).Width(u).Shrink(0).Radius(u * 0.5).Background(k.TextFaint)
					ui.Text(c, cmd).Font(MonoStack).SingleLine().Shrink(1).Grow(1).
						FontSize(core.FontSize(c, theme.RowSize)).TextColor(k.Text)
					if row.Clicked() {
						r.chosen = at
						if opts.Repeatable {
							r.stepped = 1
						}
					}
				})
			}).FillWidth().Role(ui.RoleNone)
		})

	}

	host := CodePanel(c, view, layoutContainer(opts.Width), func() {
		CodeHeader(c, opts.Name, nil, func() {
			CodeBadge(c, itoa(len(opts.Commands))+" commands", core.Neutral)
		})
	})

	r.Element = host
	return r
}

// ── DebugToolbar ────────────────────────────────────────────────────────────

// DebuggerAction is one button in the debugger's toolbar.
type DebuggerAction struct {
	// Name is the button's label, and is required: a toolbar of glyphs is a
	// toolbar nobody can use.
	Name string
	// Glyph is drawn before the name. Optional.
	Glyph *ui.SVG
	// Primary fills the button with ink, for the one action a window wants
	// found rather than avoided.
	Primary bool
	// Disabled takes it out of play.
	Disabled bool
}

// DebugToolbarOptions configure a DebugToolbar.
type DebugToolbarOptions struct {
	// Name is what the toolbar is called, and is required.
	Name string
	// Actions are the buttons, in the order they are pressed by their
	// keyboard shortcuts — left to right, which is the order a person reads
	// them and the order a debugger has used for thirty years.
	Actions []DebuggerAction
	// Running is that the program is going, and changes which of the actions
	// make sense: a toolbar offering "continue" to a program that is not
	// stopped is a toolbar with a button that does nothing.
	Running bool
	// Status is what the toolbar says about the program, at its trailing
	// edge.
	Status string
}

// DebugToolbarResult carries a DebugToolbar and what was pressed.
type DebugToolbarResult struct {
	// Element is the toolbar.
	Element *ui.Element
	// pressed is the action's name this frame, and "" for none.
	pressed string
}

// Pressed is the action pressed this frame, by its name, and "" for none.
func (r DebugToolbarResult) Pressed() string { return r.pressed }

// DebugToolbar is the row of buttons above a viewer that stop, step and
// continue a program.
//
// The status is at the trailing edge and not in the middle, because the
// buttons are read left to right and the status is not a button: putting it
// between two of them makes it look like the sixth action.
func DebugToolbar(c *ui.Context, opts DebugToolbarOptions) DebugToolbarResult {
	if opts.Name == "" {
		panic("code: DebugToolbar needs a Name; a row of buttons that says nothing about what " +
			"they act on cannot be read out")
	}
	if len(opts.Actions) == 0 {
		panic("code: DebugToolbar needs at least one action; an empty toolbar is a gap")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	var r DebugToolbarResult
	row := ui.Row(c).FillWidth().Shrink(0).Gap(u).AlignItems(ui.Center).
		Role(ui.RoleToolbar).Label(opts.Name)
	row.Children(func() {
		for _, a := range opts.Actions {
			action := a
			if action.Name == "" {
				panic("code: DebugToolbar has an action with no Name; a toolbar of glyphs is a " +
					"toolbar nobody can use")
			}
			row.Children(func() {
				bg, fg := k.Surface, k.Text
				if action.Primary {
					bg, fg = k.Fill, inkOnMark(k.Fill)
				}
				btn := ui.ButtonBase(c).Height(core.ControlHeight(c)-u*2).Shrink(0).
					Radius(theme.PillRadius).Padding(0, u*2.5, 0, u*2.5).
					Background(bg).TextColor(fg).Role(ui.RoleButton).
					Label(action.Name).Tooltip(action.Name).Disabled(action.Disabled)
				if btn.Clicked() {
					r.pressed = action.Name
				}
				btn.Children(func() {
					if action.Glyph != nil {
						ui.Icon(c, action.Glyph).Size(u*4, u*4).TextColor(fg).Shrink(0)
					}
					ui.Text(c, action.Name).SingleLine().
						FontSize(core.FontSize(c, theme.RowSize)).TextColor(fg)
				})
			})
		}
		ui.Box(c).Grow(1)
		if opts.Status != "" {
			CodeBadge(c, opts.Status, debugSeverity(opts.Status))
		}
	})
	r.Element = row
	return r
}

// debugSeverity is how the toolbar's status word is drawn, read from the word
// itself so that a caller passing "Running" and one passing "Error" get the
// two answers they meant without a severity option to keep in step.
func debugSeverity(status string) core.Severity {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "error", "failed", "crashed", "exited":
		return core.Danger
	case "stopped", "paused", "breakpoint":
		return core.Warning
	case "running":
		return core.Success
	}
	return core.Neutral
}
