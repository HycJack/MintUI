package agent

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/code"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// ── checkpoints ─────────────────────────────────────────────────────────────

// Checkpoint is one point a run can be put back to.
type Checkpoint struct {
	// ID is what identifies it to the caller that restores it. It is
	// required and it is not drawn: a restore takes an id, and a checkpoint
	// the reader has to map back to one is a checkpoint they will restore
	// the wrong of.
	ID string
	// Label is what happened at it — "before the schema change".
	Label string
	// At is when, already formatted.
	At string
	// Files is how many files it covers, and Bytes how big they were.
	Files, Bytes int
	// Status is where that checkpoint is: taken, being taken, or stale.
	Status Status
	// Current marks the point the run is at now.
	Current bool
}

// CheckpointListResult carries a CheckpointList and what was pressed in it.
type CheckpointListResult struct {
	// Element is the list.
	Element *ui.Element
	// selected is the checkpoint pressed this frame, or -1.
	selected int
	// restored is the id of the one a restore was asked of, empty for none.
	restored string
	answered bool
}

// Selected is the checkpoint pressed this frame, counted from zero, and -1 for
// none.
func (r CheckpointListResult) Selected() int {
	if !r.answered {
		return -1
	}
	return r.selected
}

// Restored is the id of the checkpoint a restore was asked of this frame,
// empty for none. The restore itself is the caller's — putting files back on
// disk is not something a list of rows may do because it was clicked.
func (r CheckpointListResult) Restored() string { return r.restored }

// CheckpointListOptions configure a CheckpointList.
type CheckpointListOptions struct {
	// Checkpoints are the checkpoints, newest first — the order a reader
	// wants them in, because the most recent one is the one they will have
	// just seen go past.
	Checkpoints []Checkpoint
	// Title heads the list. Empty takes the library's own word.
	Title string
	// Selected is the checkpoint marked as the chosen one, -1 for none.
	Selected *int
	// Empty draws instead of the list when there are none, and is required in
	// that case.
	Empty string
}

// CheckpointList is the run of points a run can be put back to.
//
// The one a reader is looking at is the current one, so it is marked rather
// than numbered: a checkpoint list is not a plan, it is a set of doors with
// one of them standing open.
func CheckpointList(c *ui.Context, opts CheckpointListOptions) CheckpointListResult {
	u := core.Density(c).Unit()
	if len(opts.Checkpoints) == 0 && opts.Empty == "" {
		panic("agent: CheckpointList with no checkpoints needs opts.Empty")
	}
	for i, cp := range opts.Checkpoints {
		if cp.ID == "" {
			panic("agent: CheckpointList checkpoint " + itoa(i) + " has no ID; a checkpoint " +
				"that cannot be named cannot be restored")
		}
	}
	if len(opts.Checkpoints) == 0 {
		return CheckpointListResult{selected: -1, Element: emptyCard(c, opts.Empty)}
	}
	title := opts.Title
	if title == "" {
		title = "Checkpoints"
	}
	var r CheckpointListResult
	r.selected = -1

	list := panel(c, layout.ContainerOptions{
		Surface: true, Radius: theme.CardRadius, Pad: u * 2.5, Gap: u * 1.25,
	}, func() {
		display.Text(c, title, display.TextOptions{Bold: true, MaxLines: 1})
		layout.Divider(c, layout.DividerOptions{})
		ui.Column(c).FillWidth().Gap(u * 0.5).Children(func() {
			for i, cp := range opts.Checkpoints {
				i, cp := i, cp
				checkpointRow(c, cp, opts.Selected != nil && *opts.Selected == i, func() {
					r.answered = true
					r.selected = i
				}, func() {
					r.restored = cp.ID
				})
			}
		})
	})
	r.Element = list
	return r
}

// checkpointRow is one checkpoint: what happened at it, what it covers, and
// the button that asks to be put back there.
func checkpointRow(c *ui.Context, cp Checkpoint, selected bool, onPress, onRestore func()) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	row := ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u*1.5).
		Padding(u*0.75, u).Radius(theme.SmallRadius).Children(func() {
		display.Icon(c, StatusIcon(cp.Status), display.IconOptions{
			Name: cp.Label + ": " + cp.Status.String(), Size: u * 4,
			Tone: StatusTone(cp.Status), Muted: StatusTone(cp.Status) == core.Neutral,
		})
		ui.Column(c).Grow(1).Shrink(0).Gap(0).Children(func() {
			ui.Text(c, cp.Label).TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).MaxLines(1)
			caption := join(" · ", cp.At, countFiles(cp.Files), countBytes(cp.Bytes))
			if caption != "" {
				ui.Text(c, caption).TextColor(k.TextFaint).
					FontSize(core.FontSize(c, theme.CaptionSize)).MaxLines(1)
			}
		})
		if cp.Current {
			display.Tag(c, "current", display.TagOptions{Tone: core.Accent})
		}
		if cp.Status == StatusDone && !cp.Current {
			if display.Icon(c, display.IconRefresh, display.IconOptions{
				Name: "Restore " + cp.Label, Size: u * 4, Muted: true,
			}).Clicked() && onRestore != nil {
				onRestore()
			}
		}
	})
	row.Label(cp.Label)
	if selected {
		row.Background(k.SurfacePressed)
	}
	if onPress != nil && row.Clicked() {
		onPress()
	}
}

// countFiles is a checkpoint's file count as a phrase, or nothing.
func countFiles(n int) string {
	if n <= 0 {
		return ""
	}
	return itoa(n) + " files"
}

// countBytes is a checkpoint's size as a phrase, or nothing.
func countBytes(n int) string {
	if n <= 0 {
		return ""
	}
	if n >= 1<<20 {
		return itoa(n>>20) + " MB"
	}
	return itoa(n>>10) + " kB"
}

// ── what the run remembers ─────────────────────────────────────────────────

// Memory is one thing a run learned and will carry to the next one.
type Memory struct {
	// Key is what the thing is about — "the go version", "where the fixtures
	// live". It is required: a memory with no key cannot be looked up.
	Key string
	// Value is what it learned.
	Value string
	// Kind is what sort of thing it is — "fact", "preference", "trap".
	Kind string
	// Pinned keeps it in the list whatever the caller would drop, because a
	// pinned memory is a decision somebody made on purpose.
	Pinned bool
	// Age is how old it is, already formatted.
	Age string
	// Used is how many runs have read it this session.
	Used int
}

// MemoryPanelResult carries a MemoryPanel and what was pressed in it.
type MemoryPanelResult struct {
	// Element is the panel.
	Element *ui.Element
	// selected is the memory pressed this frame, or -1.
	selected int
	// dropped is the key of the one a drop was asked of, empty for none.
	dropped  string
	answered bool
}

// Selected is the memory pressed this frame, counted from zero, and -1 for
// none.
func (r MemoryPanelResult) Selected() int {
	if !r.answered {
		return -1
	}
	return r.selected
}

// Dropped is the key of the memory a drop was asked of this frame, empty for
// none.
func (r MemoryPanelResult) Dropped() string { return r.dropped }

// MemoryPanelOptions configure a MemoryPanel.
type MemoryPanelOptions struct {
	// Title heads the panel. Empty takes the library's own word.
	Title string
	// Items are the things remembered, in the order they were learned.
	Items []Memory
	// Selected is the memory marked as the chosen one, -1 for none.
	Selected *int
	// Empty draws instead of the list when there is nothing remembered, and
	// is required in that case.
	Empty string
}

// MemoryPanel is what a run remembers: one row per thing, the key on the left
// and the value on the right.
//
// A memory is the one thing in a run whose value is the point, so the value
// takes the room and the key is fixed-width beside it — a column of keys is
// what lets a reader scan for one without reading all of them.
func MemoryPanel(c *ui.Context, opts MemoryPanelOptions) MemoryPanelResult {
	u := core.Density(c).Unit()
	if len(opts.Items) == 0 && opts.Empty == "" {
		panic("agent: MemoryPanel with nothing remembered needs opts.Empty")
	}
	for i, m := range opts.Items {
		if m.Key == "" {
			panic("agent: MemoryPanel item " + itoa(i) + " has no Key; a memory that cannot " +
				"be looked up is a sentence")
		}
	}
	if len(opts.Items) == 0 {
		return MemoryPanelResult{selected: -1, Element: emptyCard(c, opts.Empty)}
	}
	title := opts.Title
	if title == "" {
		title = "Memory"
	}
	var r MemoryPanelResult
	r.selected = -1

	panel := panel(c, layout.ContainerOptions{
		Surface: true, Radius: theme.CardRadius, Pad: u * 2.5, Gap: u * 1.25,
	}, func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 2).Children(func() {
			display.Text(c, title, display.TextOptions{Bold: true, MaxLines: 1})
			ui.Box(c).Grow(1)
			display.Text(c, itoa(len(opts.Items))+" remembered",
				display.TextOptions{Muted: true, MaxLines: 1})
		})
		layout.Divider(c, layout.DividerOptions{})
		ui.Column(c).FillWidth().Gap(u * 0.5).Children(func() {
			for i, m := range opts.Items {
				i, m := i, m
				memoryRow(c, m, opts.Selected != nil && *opts.Selected == i, func() {
					r.answered = true
					r.selected = i
				}, func() {
					if !m.Pinned {
						r.dropped = m.Key
					}
				})
			}
		})
	})
	r.Element = panel
	return r
}

// memoryRow is one remembered thing.
func memoryRow(c *ui.Context, m Memory, selected bool, onPress, onDrop func()) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	row := ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u*1.5).
		Padding(u*0.75, u).Radius(theme.SmallRadius).Children(func() {
		// A fixed width for the key, so that the values start in one column.
		// A key that needs more than this wraps to two lines rather than
		// pushing the value out, which is the lesser of two wrongs: a
		// wrapped key is still readable and a value off the right edge is
		// not.
		ui.Text(c, m.Key).TextColor(k.TextMuted).Font(code.MonoStack).
			FontSize(core.FontSize(c, theme.CaptionSize)).Width(u * 22).Shrink(0).MaxLines(2)
		ui.Text(c, m.Value).TextColor(k.Text).
			FontSize(core.FontSize(c, theme.RowSize)).MaxLines(2).Grow(1)
		if m.Pinned {
			display.Tag(c, "pinned", display.TagOptions{Tone: core.Accent})
		}
		if m.Kind != "" {
			display.Tag(c, m.Kind, display.TagOptions{Tone: core.Neutral})
		}
		if m.Used > 1 {
			display.Text(c, "used "+itoa(m.Used), display.TextOptions{Faint: true, MaxLines: 1})
		}
		if m.Age != "" {
			display.Text(c, m.Age, display.TextOptions{Faint: true, MaxLines: 1})
		}
		if !m.Pinned {
			if display.IconClose(c, u).Clicked() && onDrop != nil {
				onDrop()
			}
		}
	})
	row.Label(m.Key + ": " + m.Value)
	if selected {
		row.Background(k.SurfacePressed)
	}
	if onPress != nil && row.Clicked() {
		onPress()
	}
}

// ── the MCP servers ─────────────────────────────────────────────────────────

// MCPServer is one server an agent can reach its tools through.
type MCPServer struct {
	// Name is the server's name, and is required: the tools a server offers
	// are named for it, and a list of tool sets with no server above them is
	// a list nobody can reason about.
	Name string
	// Transport is how it is reached — "stdio", "http".
	Transport string
	// Endpoint is where, for a transport that has one.
	Endpoint string
	// Tools is how many tools it offers.
	Tools int
	// Status is where the connection is.
	Status Status
	// Note is one line about it.
	Note string
}

// MCPServerListResult carries an MCPServerList and what was pressed in it.
type MCPServerListResult struct {
	// Element is the list.
	Element  *ui.Element
	selected int
	answered bool
}

// Selected is the server pressed this frame, counted from zero, and -1 for
// none.
func (r MCPServerListResult) Selected() int {
	if !r.answered {
		return -1
	}
	return r.selected
}

// MCPServerListOptions configure an MCPServerList.
type MCPServerListOptions struct {
	// Servers are the servers, in the order they were configured.
	Servers []MCPServer
	// Selected is the server marked as the chosen one, -1 for none.
	Selected *int
	// Empty draws instead of the list when there are none, and is required in
	// that case: "no servers" is a configuration somebody chose.
	Empty string
}

// MCPServerList is the servers a run can reach, and whether each is up.
//
// The connection is the whole point of the list and it is the first thing in
// each row, because a tool call that fails on a server that was down is a
// failure with an obvious cause and one that fails on a server that was up is
// not.
func MCPServerList(c *ui.Context, opts MCPServerListOptions) MCPServerListResult {
	u := core.Density(c).Unit()
	if len(opts.Servers) == 0 && opts.Empty == "" {
		panic("agent: MCPServerList with no servers needs opts.Empty")
	}
	for i, s := range opts.Servers {
		if s.Name == "" {
			panic("agent: MCPServerList server " + itoa(i) + " has no Name")
		}
	}
	if len(opts.Servers) == 0 {
		return MCPServerListResult{selected: -1, Element: emptyCard(c, opts.Empty)}
	}
	var r MCPServerListResult
	r.selected = -1

	list := panel(c, layout.ContainerOptions{
		Surface: true, Radius: theme.CardRadius, Pad: u * 2.5, Gap: u * 1.25,
	}, func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 2).Children(func() {
			display.Text(c, "MCP servers", display.TextOptions{Bold: true, MaxLines: 1})
			ui.Box(c).Grow(1)
			statuses := make([]Status, 0, len(opts.Servers))
			for _, s := range opts.Servers {
				statuses = append(statuses, s.Status)
			}
			AgentStatus(c, AgentStatusOptions{
				Status: WorstStatus(statuses),
				Label:  serverSummary(statuses),
			})
		})
		layout.Divider(c, layout.DividerOptions{})
		ui.Column(c).FillWidth().Gap(u * 0.5).Children(func() {
			for i, s := range opts.Servers {
				i, s := i, s
				mcpRow(c, s, opts.Selected != nil && *opts.Selected == i, func() {
					r.answered = true
					r.selected = i
				})
			}
		})
	})
	r.Element = list
	return r
}

// serverSummary is the one word for several servers' connections: "2 of 3
// connected". It is one figure rather than three marks because the question a
// reader has is about the set.
func serverSummary(statuses []Status) string {
	up, total := 0, len(statuses)
	for _, s := range statuses {
		if s == StatusDone || s == StatusRunning {
			up++
		}
	}
	return itoa(up) + " of " + itoa(total) + " connected"
}

// mcpRow is one server: its name, how it is reached, how many tools it has,
// and whether it is up.
func mcpRow(c *ui.Context, s MCPServer, selected bool, onPress func()) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	// The row and its note live in a column of their own: the note is a
	// second line under the row, where it reads as a remark about the
	// server, rather than appended to the row itself — appended there it sat
	// after the status mark, and the two collided whenever the endpoint had
	// not handed the width back.
	col := ui.Column(c).FillWidth().Gap(u * 0.5)
	col.Children(func() {
		row := ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u*1.5).
			Padding(u*0.75, u).Radius(theme.SmallRadius).Children(func() {
			ui.Text(c, s.Name).TextColor(k.Text).Bold().
				FontSize(core.FontSize(c, theme.RowSize)).MaxLines(1).Grow(1)
			if s.Transport != "" {
				display.Tag(c, s.Transport, display.TagOptions{Tone: core.Neutral})
			}
			if s.Tools > 0 {
				display.Text(c, itoa(s.Tools)+" tools", display.TextOptions{Muted: true, MaxLines: 1})
			}
			if s.Endpoint != "" {
				// The endpoint yields: an unshrinking command line this long
				// pushed the status out of the row entirely.
				display.Text(c, s.Endpoint, display.TextOptions{Faint: true, Mono: true, MaxLines: 1}).Shrink(1)
			}
			AgentStatus(c, AgentStatusOptions{Status: s.Status})
		})
		if s.Note != "" {
			display.Text(c, s.Note, display.TextOptions{Faint: true, MaxLines: 1}).
				Padding(0, 0, 0, u*0.75)
		}
		col.Label(s.Name + ", " + s.Status.String())
		if selected {
			row.Background(k.SurfacePressed)
		}
		if onPress != nil && row.Clicked() {
			onPress()
		}
	})
}

// ── the sandbox ─────────────────────────────────────────────────────────────

// SandboxOptions configure a SandboxStatus.
type SandboxOptions struct {
	// On says the run is inside a sandbox at all. Off is not the same as on
	// with everything allowed: a run with no sandbox is a run with nothing
	// between it and the machine, and that is worth saying in the same words
	// rather than as a list of permissions that happens to be complete.
	On bool
	// Mode names the sandbox's policy — "seatbelt", "danger-full-access".
	Mode string
	// WorkingDir is where the run is allowed to write.
	WorkingDir string
	// Allow and Deny are the rules, grouped so that a reader can see at a
	// glance that the deny list is not empty. A sandbox whose rules are one
	// undifferentiated run of text is a sandbox nobody has checked.
	Allow, Deny []string
	// Note is one line about the policy.
	Note string
}

// SandboxStatus is what a run is allowed to do, and where.
//
// It is a panel rather than a row because the answer is never one word: a
// reviewer is checking two lists — what the run may touch and what it may not
// — and both of them have to be on the same surface to be compared.
func SandboxStatus(c *ui.Context, opts SandboxOptions) *ui.Element {
	u := core.Density(c).Unit()
	k := core.Tokens(c)
	if !opts.On && opts.Mode == "" && opts.Note == "" &&
		len(opts.Allow) == 0 && len(opts.Deny) == 0 {
		panic("agent: SandboxStatus with nothing said about the sandbox; a run that is not " +
			"sandboxed and a run nobody has configured yet are very different things, and " +
			"drawing the same panel for both would hide which one this is")
	}

	tone := core.Success
	label := "Sandboxed"
	if !opts.On {
		// No sandbox is not a warning — it is the absence of the thing that
		// would have been a warning, and saying it in the warning colour
		// would put every unsandboxed run on the page in amber for a state
		// the caller may have chosen on purpose.
		tone = core.Neutral
		label = "Not sandboxed"
	}

	return panel(c, layout.ContainerOptions{
		Surface: true, Radius: theme.CardRadius, Pad: u * 2.5, Gap: u * 1.5,
	}, func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 2).Children(func() {
			display.Icon(c, display.IconSettings, display.IconOptions{
				Name: "Sandbox", Size: u * 5,
				Tone: tone, Muted: tone == core.Neutral,
			})
			AgentStatus(c, AgentStatusOptions{
				Status: sandboxStatusFor(opts.On), Label: label, Pill: true,
				Icon: false,
			})
			if opts.Mode != "" {
				display.Tag(c, opts.Mode, display.TagOptions{Tone: tone})
			}
			ui.Box(c).Grow(1)
			if opts.WorkingDir != "" {
				display.Text(c, opts.WorkingDir, display.TextOptions{Faint: true, Mono: true, MaxLines: 1})
			}
		})
		if len(opts.Allow) > 0 || len(opts.Deny) > 0 {
			ui.Row(c).FillWidth().AlignItems(ui.Start).Gap(u * 2).Children(func() {
				ruleList(c, "May", opts.Allow, k.Success)
				ruleList(c, "May not", opts.Deny, k.Danger)
			})
		}
		if opts.Note != "" {
			display.Text(c, opts.Note, display.TextOptions{Muted: true, MaxLines: 2})
		}
	})
}

// sandboxStatusFor is the run status a sandbox is drawn with, so that the
// panel's own mark and everything else's agree.
func sandboxStatusFor(on bool) Status {
	if on {
		return StatusDone
	}
	return StatusQueued
}

// ruleList is one column of the sandbox's rules.
func ruleList(c *ui.Context, title string, rules []string, ink ui.Color) {
	u := core.Density(c).Unit()
	k := core.Tokens(c)
	if len(rules) == 0 {
		return
	}
	ui.Column(c).Grow(1).Shrink(0).Gap(u * 0.5).Children(func() {
		ui.Text(c, title).TextColor(ink).FontSize(core.FontSize(c, theme.CaptionSize)).Bold()
		for _, rule := range rules {
			ui.Text(c, rule).TextColor(k.Text).Font(code.MonoStack).
				FontSize(core.FontSize(c, theme.CaptionSize)).MaxLines(1)
		}
	})
}

// ── the sub-agents ──────────────────────────────────────────────────────────

// SubAgent is one agent another agent started.
type SubAgent struct {
	// Name is what it is called. It is required.
	Name string
	// Role is what it is for.
	Role string
	// Status is where it is.
	Status Status
	// Steps and Tools are how much it has done.
	Steps, Tools int
	// Note is one line about it.
	Note string
	// Children are the agents it started in turn.
	Children []SubAgent
}

// SubAgentTreeResult carries a SubAgentTree and what was pressed in it.
type SubAgentTreeResult struct {
	// Element is the tree.
	Element *ui.Element
	// selected is the row pressed this frame, counted in the order the rows
	// are drawn, or -1.
	selected int
	// toggled is the agent whose children opened or closed, or empty.
	toggled  string
	answered bool
}

// Selected is the row pressed this frame, counted in the order the rows are
// drawn, and -1 for none. Counting by drawn row rather than by a path means
// a caller can use it straight as a row number into whatever list it built
// the tree from.
func (r SubAgentTreeResult) Selected() int {
	if !r.answered {
		return -1
	}
	return r.selected
}

// Toggled is the name of the agent whose children opened or closed this
// frame, empty for none.
func (r SubAgentTreeResult) Toggled() string { return r.toggled }

// SubAgentTreeOptions configure a SubAgentTree.
type SubAgentTreeOptions struct {
	// Root is the agent at the top. Required.
	Root SubAgent
	// Open is which agents are showing their children, keyed by name, and
	// the caller's. It is keyed by name rather than by a path because a tree
	// is rebuilt from a transcript on every frame and paths do not survive
	// that; names are the transcript's own words and do.
	Open *Open[string]
	// Selected is the row marked as the chosen one, counted in the order the
	// rows are drawn, -1 for none.
	Selected *int
	// Empty draws instead of the tree when the root has no name, and is
	// required in that case.
	Empty string
}

// SubAgentTree is the agents one agent started, and the ones they started in
// turn.
//
// The rows are walked rather than listed because the set of rows a tree shows
// is decided entirely by what is open — which is a small set held by the
// caller — so walking it every frame costs what the open set costs and not
// what the whole tree costs. A run that started two hundred agents and opened
// three of them draws three rows and their children.
func SubAgentTree(c *ui.Context, opts SubAgentTreeOptions) SubAgentTreeResult {
	u := core.Density(c).Unit()
	if opts.Root.Name == "" && opts.Empty == "" {
		panic("agent: SubAgentTree needs a Root with a Name; an agent with no name is not " +
			"something a tree can be about")
	}
	if opts.Root.Name == "" {
		return SubAgentTreeResult{selected: -1, Element: emptyCard(c, opts.Empty)}
	}
	var r SubAgentTreeResult
	r.selected = -1

	row := 0
	tree := ui.Column(c).FillWidth().Gap(u * 0.5).Children(func() {
		// The root's children are always on show: the root is the run itself,
		// which whatever drew the tree has already shown somewhere else, and a
		// tree whose whole job is the agents a run started should not open
		// with one row and a chevron.
		subAgentRows(c, []SubAgent{opts.Root}, 0, opts, &row, &r, true)
	})
	tree.Label(opts.Root.Name + " and its sub-agents")
	r.Element = tree
	return r
}

// subAgentRows draws one level of agents and, under each open one, the level
// below it. first is the root level, whose children are on show whatever the
// caller's open set says.
func subAgentRows(c *ui.Context, agents []SubAgent, depth int, opts SubAgentTreeOptions, row *int, r *SubAgentTreeResult, first bool) {
	for _, a := range agents {
		here := *row
		*row++
		subAgentRow(c, a, subAgentRowOptions{
			depth:    depth,
			kids:     len(a.Children),
			open:     opts.Open != nil && opts.Open.Has(a.Name),
			selected: opts.Selected != nil && *opts.Selected == here,
		}, func() {
			r.answered = true
			r.selected = here
		}, func() {
			if opts.Open != nil {
				opts.Open.Set(a.Name, !opts.Open.Has(a.Name))
			}
			r.toggled = a.Name
		})
		if len(a.Children) > 0 && (first || (opts.Open != nil && opts.Open.Has(a.Name))) {
			subAgentRows(c, a.Children, depth+1, opts, row, r, false)
		}
	}
}

// subAgentRowOptions is one agent row's own settings.
type subAgentRowOptions struct {
	depth    int
	kids     int
	open     bool
	selected bool
}

// subAgentRow is one agent: its name, its role, its status, and how far down
// the tree it sits.
func subAgentRow(c *ui.Context, a SubAgent, opts subAgentRowOptions, onPress, onToggle func()) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	row := ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u*1.5).
		Padding(u*0.75, u).Radius(theme.SmallRadius).Children(func() {
		if opts.kids > 0 {
			name := "Expand " + a.Name
			if opts.open {
				name = "Collapse " + a.Name
			}
			if display.Icon(c, chevron(opts.open), display.IconOptions{
				Name: name, Size: u * 4, Muted: true,
			}).Clicked() && onToggle != nil {
				onToggle()
			}
		} else {
			// The gap where a chevron would be, so that every row's name
			// starts in the same place whether or not it has children.
			ui.Box(c).Width(u * 4).Shrink(0)
		}
		display.Avatar(c, a.Name)
		ui.Column(c).Grow(1).Shrink(0).Gap(0).Children(func() {
			ui.Text(c, a.Name).TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).MaxLines(1)
			if a.Role != "" || a.Note != "" {
				ui.Text(c, join(" · ", a.Role, a.Note)).TextColor(k.TextFaint).
					FontSize(core.FontSize(c, theme.CaptionSize)).MaxLines(1)
			}
		})
		if a.Steps > 0 {
			display.Text(c, itoa(a.Steps)+" steps", display.TextOptions{Muted: true, MaxLines: 1})
		}
		if a.Tools > 0 {
			display.Text(c, itoa(a.Tools)+" tools", display.TextOptions{Muted: true, MaxLines: 1})
		}
		AgentStatus(c, AgentStatusOptions{Status: a.Status})
	})
	row.Margin(0, 0, 0, u*4*float32(clampDepth(opts.depth)))
	row.Label(a.Name + ", " + a.Status.String())
	if opts.selected {
		row.Background(k.SurfacePressed)
	}
	if onPress != nil && row.Clicked() {
		onPress()
	}
}
