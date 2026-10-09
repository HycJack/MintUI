package pages

import (
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/agent"
	"github.com/HycJack/MintUI/ui/code"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/showcase"
)

// Demo state outlives the frame: every demo below hands its component
// a pointer, and a pointer into a frame-local is a click the next
// frame undoes — a tab that will not switch, a dropdown that snaps
// shut, a slider that springs back.
var (
	selectedCard       = false
	stepSel            = 4
	callOpen           = true
	groupOpen          = true
	groupClosed        = false
	cmdOpen            = true
	failedOpen         = true
	changedCard        = true
	artifactSel        = 0
	currentVersion     = 1
	shotSel            = 0
	checkpointSel      = 1
	memorySel          = 1
	subSel             = 2
	serverSel          = -1
	draft              = "postgres — the sqlite fixtures are stale"
	picked             = ""
	permissionRemember = false
	approvalChosen     = 1
	approvalRemember   = false
	wellSel            = 3
)

// The page is one run of an agent, from its header to the review of what it
// changed, drawn with every component this package has. The order is the
// order a person reads a run in rather than the order the components are
// declared in, because a gallery is read as a thing and not as an index.
//
// Everything on the page is fixed: the same steps, the same three versions of
// the same file, the same counts. A page that drew differently each time it
// was opened could not be compared against the last one, and a screenshot
// that changed between runs is a screenshot nobody reviews.

func init() {
	showcase.Register(showcase.Page{
		Package: "agent",
		Title:   "ui/agent — Agent 运行过程",
		Note:    "状态、步骤、工具调用、权限请求",
		Width:   1000,
		Height:  5500,
		Want: []string{
			// the run itself
			"main", "orchestrator", "opus", "~/work/callbacks",
			// status
			"Running", "Waiting", "Failed", "Cancelled", "Queued",
			// steps
			"read the brief", "search the repo", "draw the components",
			"Ship the agent gallery page",
			// tool calls
			"Read", "ui/core/state.go", "Grep", "1 failed", "no such file",
			// commands and files
			"go test ./ui/agent -count=1", "exit 0", "exit 2", "ui/agent/well.go",
			// artifacts
			"Artifacts", "report.html", "v3", "v2 after review", "coverage.csv",
			// screenshots
			"gallery page", "failing test", "merge review",
			// checkpoints and memory
			"Checkpoints", "before the merge", "current", "Memory", "go.version",
			// servers and sandbox
			"MCP servers", "filesystem", "1 of 2 connected", "seatbelt", "May not",
			// sub-agents
			"explorer", "tester", "go-test",
			// requests and modals
			"Which database should the migration target?", "Write wants permission",
			"Allow Bash?", "Always",
			// the diffs
			"1 conflict", "1 to resolve",
			// the transcript
			"Keep ours or take theirs?",
		},
		Render: func(c *ui.Context) {
			runPage(c)
		},
	})
}

func runPage(c *ui.Context) {
	// ── status ─────────────────────────────────────────────────────────────
	showcase.Section(c, "状态 · AgentStatus / AgentParts / AgentCaption / AgentCard")

	// The six statuses in a row is the first thing worth looking at: it is
	// the mapping from a word to a colour, and a page that shows the six of
	// them together is the only place that mapping can be seen all at once.
	showcase.Field(c, "StatusTone: 每个状态一个颜色")
	// A plain row rather than showcase.Stack: six pills of the same size in
	// six wrapping rows is a column, and the point of this one is that they
	// are comparable side by side.
	ui.Row(c).FillWidth().Gap(12).Children(func() {
		for _, s := range []agent.Status{
			agent.StatusQueued, agent.StatusRunning, agent.StatusWaiting,
			agent.StatusDone, agent.StatusFailed, agent.StatusCancelled,
		} {
			agent.AgentStatus(c, agent.AgentStatusOptions{Status: s, Pill: true})
		}
	})

	showcase.Field(c, "AgentStatus with a detail line and a glyph")
	ui.Column(c).FillWidth().Gap(core.Density(c).Unit()).Children(func() {
		agent.AgentStatus(c, agent.AgentStatusOptions{
			Status: agent.StatusRunning, Label: "running the suite",
			Detail:   "go test ./ui/agent — 41 packages, 3 failures so far",
			Duration: "2m 14s", Icon: true,
		})
		agent.AgentStatus(c, agent.AgentStatusOptions{
			Status: agent.StatusWaiting, Label: "waiting for your answer",
			Detail: "the run has stopped; nothing will move until it is answered",
		})
	})

	showcase.Field(c, "AgentParts — a run's header")
	agent.AgentParts(c, agent.AgentPartsOptions{
		Name: "main", Role: "orchestrator · claude opus",
		Status: agent.StatusRunning, Footer: "~/work/callbacks · branch main · started 14:02",
		Badges: []agent.AgentBadge{
			{Label: "opus", Tone: core.Accent},
			{Label: "12 steps"},
			{Label: "31 tool calls"},
			{Label: "sandboxed"},
		},
	})

	showcase.Field(c, "AgentCard — one agent in a list")
	ui.Row(c).FillWidth().Gap(6).Children(func() {
		agent.AgentCard(c, agent.AgentCardOptions{
			Name: "explorer", Role: "read-only search", Status: agent.StatusDone,
			Steps: 6, Done: 6, Tools: 24, Meter: true, Footer: "finished 14:09",
			Selected: &selectedCard,
		})
		agent.AgentCard(c, agent.AgentCardOptions{
			Name: "tester", Role: "runs the suite", Status: agent.StatusRunning,
			Steps: 2, Done: 1, Tools: 3, Meter: true, Footer: "14:11",
		})
	})

	showcase.Field(c, "AgentCaption — who said it and when")
	ui.Column(c).FillWidth().Gap(4).Children(func() {
		agent.AgentCaption(c, agent.AgentCaptionOptions{
			Name: "main", Role: "orchestrator", At: "14:02:11", Note: "read the brief",
		})
		agent.AgentCaption(c, agent.AgentCaptionOptions{
			Name: "explorer", Role: "read-only", At: "14:09:40",
			Note: "12 files, no edits", Tone: core.Success,
		})
	})

	// ── steps ──────────────────────────────────────────────────────────────
	showcase.Section(c, "步骤 · AgentStepList / AgentProgress / AgentPlan")

	steps := []agent.Step{
		{Title: "read the brief", Detail: "README.md, docs/design-system.md",
			Status: agent.StatusDone, Duration: "0.4s"},
		{Title: "search the repo", Detail: "rg -n StatusTone ui/ — 14 hits in 6 files",
			Status: agent.StatusDone, Depth: 1, Tool: "Grep", Duration: "1.2s"},
		{Title: "read ui/core/state.go", Detail: "81 lines",
			Status: agent.StatusDone, Depth: 1, Tool: "Read", Duration: "0.3s"},
		{Title: "write ui/agent/status.go", Detail: "146 lines added, 0 removed",
			Status: agent.StatusDone, Depth: 1, Tool: "Write", Duration: "3.8s"},
		{Title: "draw the components", Detail: "26 of 26 drawn",
			Status: agent.StatusRunning, Duration: "1m 02s"},
		{Title: "check both appearances", Status: agent.StatusQueued, Depth: 1},
		{Title: "write the gallery page", Status: agent.StatusQueued},
	}

	showcase.Field(c, "AgentStepList — numbers and levels are computed, not stored")
	stepOpen := &agent.Open[int]{}
	stepOpen.Open(1)
	agent.AgentStepList(c, agent.AgentStepListOptions{
		Steps: steps, Open: stepOpen, Selected: &stepSel, Numbers: true,
	})

	showcase.Field(c, "AgentProgress — a length that is known, and one that is not")
	ui.Row(c).FillWidth().Gap(8).AlignItems(ui.Start).Children(func() {
		ui.Box(c).Grow(1).Children(func() {
			agent.AgentProgress(c, agent.AgentProgressOptions{
				Status: agent.StatusRunning, Done: 4, Total: 7,
				Detail: "three steps to go", ShowPercent: true,
			})
		})
		ui.Box(c).Grow(1).Children(func() {
			agent.AgentProgress(c, agent.AgentProgressOptions{
				Status: agent.StatusRunning, Done: 4,
				Detail: "the plan is still being written",
			})
		})
		ui.Box(c).Grow(1).Children(func() {
			agent.AgentStepList(c, agent.AgentStepListOptions{
				Empty: "Nothing has happened yet — the run has not reported a first step.",
			})
		})
	})

	showcase.Field(c, "AgentPlan — the plan somebody approves before the work starts")
	agent.AgentPlan(c, agent.AgentPlanOptions{
		Title: "Ship the agent gallery page",
		Goal:  "one page per ui/ package, both appearances, a screenshot nobody has to guess about",
		Steps: steps, ShowProgress: true,
	})

	// ── tool calls ─────────────────────────────────────────────────────────
	showcase.Section(c, "工具调用 · ToolCallCard / ToolCallGroup / CommandExecutionCard / FileChangeCard")

	showcase.Field(c, "ToolCallCard — one call, and what it said back")
	agent.ToolCallCard(c, agent.ToolCallCardOptions{
		Call: agent.ToolCall{
			Tool: "Read", Summary: "ui/core/state.go", Status: agent.StatusDone,
			Result: "4.1 kB", Duration: "0.2s",
		},
		Server: "filesystem",
		Detail: "{\n  \"path\": \"ui/core/state.go\",\n  \"start\": 1,\n  \"limit\": 81\n}",
		Open:   &callOpen, Height: 92,
	})

	showcase.Field(c, "ToolCallGroup — the same tool, gathered up")
	agent.ToolCallGroup(c, agent.ToolCallGroupOptions{
		Tool: "Read", Open: &groupOpen, Height: 130,
		Calls: []agent.ToolCall{
			{Tool: "Read", Summary: "ui/core/core.go", Status: agent.StatusDone,
				Result: "6.2 kB", Duration: "0.2s"},
			{Tool: "Read", Summary: "ui/core/state.go", Status: agent.StatusDone,
				Result: "4.1 kB", Duration: "0.2s"},
			{Tool: "Read", Summary: "ui/theme/theme.go", Status: agent.StatusFailed,
				Result: "no such file", Duration: "0.1s"},
		},
	})
	agent.ToolCallGroup(c, agent.ToolCallGroupOptions{
		Tool: "Grep", Open: &groupClosed,
		Calls: []agent.ToolCall{
			{Tool: "Grep", Summary: "StatusTone in ui/", Status: agent.StatusDone,
				Result: "14 matches", Duration: "1.2s"},
			{Tool: "Grep", Summary: "Merge3 in ui/", Status: agent.StatusDone,
				Result: "3 matches", Duration: "0.6s"},
			{Tool: "Grep", Summary: "open.Split", Status: agent.StatusDone,
				Result: "1 match", Duration: "0.3s"},
		},
	})

	showcase.Field(c, "CommandExecutionCard — the command, the code, the output")
	agent.CommandExecutionCard(c, agent.CommandExecutionCardOptions{
		Open: &cmdOpen, Height: 108,
		Exec: agent.CommandExecution{
			Command: "go test ./ui/agent -count=1",
			Dir:     "~/work/callbacks", ExitCode: 0, Status: agent.StatusDone,
			Duration: "4.1s", Truncated: 12,
			Output: []string{
				"--- FAIL: TestMerge3OnARealConflict (0.00s)",
				"    agent_test.go:81: line 3 is Removed, want Added",
				"FAIL\tcallbacks/ui/agent\t0.31s",
				"ok  \tcallbacks/ui/agent\t0.31s",
			},
		},
	})
	agent.CommandExecutionCard(c, agent.CommandExecutionCardOptions{
		Open: &failedOpen, Height: 84,
		Exec: agent.CommandExecution{
			Command: "go build ./ui/agent", Dir: "~/work/callbacks",
			ExitCode: 2, Status: agent.StatusFailed, Duration: "0.9s",
			Output: []string{
				"./ui/agent/steps.go:118:4: declared and not used: k",
				"./ui/agent/steps.go:141:2: no new variables on left side of :=",
			},
		},
	})

	showcase.Field(c, "FileChangeCard — what a run did to each file")
	ui.Row(c).FillWidth().Gap(8).Children(func() {
		for _, f := range []agent.FileChangeCardOptions{
			{Path: "ui/agent/well.go", Change: agent.FileAdded, Added: 180, Bar: true,
				Selected: &changedCard, Footer: "new file"},
			{Path: "ui/agent/shared.go", Change: agent.FileModified,
				Added: 14, Removed: 3, Bar: true},
			{Path: "ui/agent/draft.go", Change: agent.FileRemoved, Removed: 96,
				Footer: "deleted"},
		} {
			f := f
			ui.Box(c).Grow(1).Children(func() {
				agent.FileChangeCard(c, f)
			})
		}
	})
	agent.FileChangeCard(c, agent.FileChangeCardOptions{
		Path: "ui/agent/agent.go", Change: agent.FileRenamed,
		From: "ui/agent/agent.go", To: "ui/agent/doc.go",
	})

	// ── artifacts ──────────────────────────────────────────────────────────
	showcase.Section(c, "产物 · ArtifactPanel / ArtifactVersionSwitcher / ScreenshotStream")

	showcase.Field(c, "ArtifactPanel, ArtifactVersionSwitcher, ScreenshotStream")
	ui.Row(c).FillWidth().Gap(8).AlignItems(ui.Start).Children(func() {
		ui.Column(c).Grow(1).Shrink(0).Gap(6).Children(func() {
			agent.ArtifactPanel(c, agent.ArtifactPanelOptions{
				Title: "Artifacts", ShowVersions: true, Selected: &artifactSel,
				Artifacts: []agent.Artifact{
					{Name: "report.html", Kind: "HTML", Size: "12 kB", Version: 3,
						Status: agent.StatusDone, Note: "the gallery page, rendered"},
					{Name: "shot.png", Kind: "PNG", Size: "84 kB", Version: 2,
						Status: agent.StatusDone, Note: "the page in the dark palette"},
					{Name: "coverage.csv", Kind: "CSV", Size: "3 kB", Version: 1,
						Status: agent.StatusRunning, Note: "being written"},
				},
			})
			agent.ArtifactVersionSwitcher(c, agent.ArtifactVersionSwitcherOptions{
				Name: "report.html", Current: &currentVersion,
				Versions: []agent.ArtifactVersion{
					{Number: 1, Note: "the first pass"},
					{Number: 2, Label: "after review", Status: agent.StatusDone},
					{Number: 3, Label: "before dark", Status: agent.StatusRunning},
				},
			})
			agent.ArtifactPanel(c, agent.ArtifactPanelOptions{
				Title: "Artifacts",
				Empty: "This run produced nothing — it answered a question and stopped.",
			})
		})
		ui.Column(c).Grow(1).Shrink(0).Gap(6).Children(func() {
			agent.ScreenshotStream(c, agent.ScreenshotStreamOptions{
				Columns: 2, Selected: &shotSel,
				Shots: []agent.Screenshot{
					{Name: "gallery page", At: "14:07", Pixels: "1200×800", Note: "light"},
					{Name: "failing test", At: "14:07", Pixels: "900×600", Note: "the red one"},
					{Name: "merge review", At: "14:08", Pixels: "1440×900", Note: "one conflict"},
					{Name: "dark page", At: "14:08", Pixels: "1200×800", Note: "dark"},
				},
			})
			agent.ScreenshotStream(c, agent.ScreenshotStreamOptions{
				Empty: "No pictures yet — this run never looked at a window.",
			})
		})
	})
	// ── what the run remembers ─────────────────────────────────────────────
	showcase.Section(c, "会话 · CheckpointList / MemoryPanel / MCPServerList / SandboxStatus / SubAgentTree")

	showcase.Field(c, "CheckpointList and MemoryPanel — what can be undone, and what is kept")
	ui.Row(c).FillWidth().Gap(8).Children(func() {
		ui.Box(c).Grow(1).Children(func() {
			agent.CheckpointList(c, agent.CheckpointListOptions{
				Checkpoints: []agent.Checkpoint{
					{ID: "cp-1", Label: "before the schema change", At: "13:40", Files: 3,
						Status: agent.StatusDone},
					{ID: "cp-2", Label: "before the merge", At: "13:58", Files: 7,
						Status: agent.StatusDone, Current: true},
				},
				Selected: &checkpointSel,
			})
		})
		ui.Box(c).Grow(1).Children(func() {
			agent.MemoryPanel(c, agent.MemoryPanelOptions{
				Selected: &memorySel,
				Items: []agent.Memory{
					{Key: "go.version", Value: "1.27.1 — GOTOOLCHAIN=local", Kind: "fact",
						Used: 3, Age: "2h"},
					{Key: "fixtures", Value: "internal/store/testdata", Kind: "trap",
						Pinned: true, Age: "2h"},
					{Key: "gallery", Value: "go run ./cmd/gallery -shots <dir>", Kind: "howto",
						Used: 2, Age: "1d"},
				},
			})
		})
	})
	ui.Row(c).FillWidth().Gap(8).Children(func() {
		ui.Box(c).Grow(1).Children(func() {
			agent.CheckpointList(c, agent.CheckpointListOptions{
				Empty: "No checkpoints — nothing in this run was worth being able to undo.",
			})
		})
		ui.Box(c).Grow(1).Children(func() {
			agent.MemoryPanel(c, agent.MemoryPanelOptions{
				Empty: "Nothing remembered — this is the first run of its kind.",
			})
		})
	})

	subOpen := &agent.Open[string]{}
	subOpen.Open("explorer")
	subOpen.Open("tester")
	showcase.Field(c, "MCPServerList and SubAgentTree — where the tools come from, and what they started")
	ui.Row(c).FillWidth().Gap(8).AlignItems(ui.Start).Children(func() {
		ui.Column(c).Grow(1).Shrink(0).Gap(6).Children(func() {
			agent.MCPServerList(c, agent.MCPServerListOptions{
				Servers: []agent.MCPServer{
					{Name: "filesystem", Transport: "stdio", Tools: 12, Status: agent.StatusDone},
					{Name: "github", Transport: "http", Endpoint: "https://mcp.example/github",
						Tools: 30, Status: agent.StatusFailed, Note: "token expired at 13:51"},
				},
				Selected: &serverSel,
			})
			agent.MCPServerList(c, agent.MCPServerListOptions{
				Empty: "No servers configured — every tool here is built into the agent.",
			})
		})
		ui.Column(c).Grow(1).Shrink(0).Gap(6).Children(func() {
			agent.SubAgentTree(c, agent.SubAgentTreeOptions{
				Open: subOpen, Selected: &subSel,
				Root: agent.SubAgent{
					Name: "main", Role: "orchestrator", Status: agent.StatusRunning,
					Steps: 12, Tools: 31,
					Children: []agent.SubAgent{
						{Name: "explorer", Role: "read-only search", Status: agent.StatusDone,
							Steps: 6, Tools: 24, Note: "no edits"},
						{Name: "tester", Role: "runs the suite", Status: agent.StatusRunning,
							Steps: 2, Tools: 3, Children: []agent.SubAgent{
								{Name: "go-test", Status: agent.StatusRunning, Note: "./ui/agent"},
							}},
					},
				},
			})
			agent.SubAgentTree(c, agent.SubAgentTreeOptions{
				Empty: "No sub-agents — this run did everything itself.",
			})
		})
	})
	showcase.Field(c, "SandboxStatus — what the run may and may not touch")
	ui.Row(c).FillWidth().Gap(8).Children(func() {
		ui.Box(c).Grow(2).Children(func() {
			agent.SandboxStatus(c, agent.SandboxOptions{
				On: true, Mode: "seatbelt", WorkingDir: "~/work",
				Allow: []string{"read under ~/work", "run go test", "run gofmt"},
				Deny:  []string{"write outside ~/work", "any network call"},
				Note:  "Anything outside these rules stops and asks.",
			})
		})
		ui.Box(c).Grow(1).Children(func() {
			agent.SandboxStatus(c, agent.SandboxOptions{
				On: false, Note: "This run has no sandbox: nothing is between it and the machine.",
			})
		})
	})

	// ── asking ─────────────────────────────────────────────────────────────
	showcase.Section(c, "询问 · HumanInputRequest / PermissionPrompt / ToolApprovalDialog")

	showcase.Field(c, "HumanInputRequest — a question in the middle of the run")
	agent.HumanInputRequest(c, agent.HumanInputRequestOptions{
		From: "main", Question: "Which database should the migration target?",
		Why:   "Two are configured and only one has the fixtures this run was tested against.",
		Draft: &draft, Urgency: agent.StatusWaiting, Deadline: "2m",
		FieldLabel: "Target database",
	})
	agent.HumanInputRequest(c, agent.HumanInputRequestOptions{
		From: "main", Question: "Which branch should the merge resolve to?",
		Why:    "Both changed the same line in shared.go.",
		Chosen: &picked, Urgency: agent.StatusQueued,
		Choices: []string{"keep ours", "take theirs", "ask me each time"},
	})

	// The two modals below are layers over the window when they are used as
	// they are meant to be, which on a gallery page would put a scrim across
	// everything else on it. Preview draws the same panel in the flow, so
	// both can be seen at once and the page stays readable.
	showcase.Field(c, "PermissionPrompt and ToolApprovalDialog — Preview：同一个面板，画在浮层之外")
	// Straight into the row, with no wrapper: each preview is already exactly
	// as wide as its panel, and a box around one only gives the row something
	// whose width has to be worked out twice.
	ui.Row(c).FillWidth().Gap(12).AlignItems(ui.Start).Children(func() {
		agent.PermissionPrompt(c, agent.PermissionPromptOptions{
			Preview: true, Remember: &permissionRemember, Width: 440,
			Request: agent.PermissionRequest{
				Tool: "Write", What: "write ui/agent/status.go",
				Why:    "the plan has a step for it and it is the only file left",
				Detail: "{\n  \"path\": \"ui/agent/status.go\",\n  \"lines\": \"+146\"\n}",
				Server: "filesystem", Risk: core.Warning, Scope: "for this run",
			},
		})
		agent.ToolApprovalDialog(c, agent.ToolApprovalDialogOptions{
			Preview: true, Chosen: &approvalChosen, Remember: &approvalRemember,
			Width: 440,
			Approval: agent.ToolApproval{
				Tool: "Bash", Server: "shell",
				Args: "git push origin main",
				Why:  "the run has finished and the work is committed",
				Risk: core.Danger, SameTool: 6,
				Choices: []string{"Deny", "Allow once", "Always"},
			},
		})
	})

	// ── the transcript ─────────────────────────────────────────────────────
	showcase.Section(c, "过程 · AgentWellRows")

	showcase.Field(c, "AgentWellRows — one column, one shape of row per kind")
	agent.AgentWellRows(c, agent.AgentWellRowsOptions{
		Selected: &wellSel,
		Rows: []agent.WellRow{
			{Kind: agent.RowStep, At: "14:02", Status: agent.StatusDone,
				Agent: "main", Step: agent.Step{Title: "read the brief", Status: agent.StatusDone}},
			{Kind: agent.RowThinking, At: "14:02", Status: agent.StatusDone,
				Agent: "main",
				Text:  "The gallery page for ui/agent has to draw every component, and two of them are layers over the window."},
			{Kind: agent.RowToolCall, At: "14:03", Status: agent.StatusDone,
				Call: agent.ToolCall{Tool: "Read", Summary: "ui/showcase/page.go",
					Result: "4.1 kB", Duration: "0.2s", Status: agent.StatusDone}},
			{Kind: agent.RowCommand, At: "14:04", Status: agent.StatusRunning,
				Exec: agent.CommandExecution{Command: "go test ./ui/showcase/...",
					Dir: "~/work/callbacks", Status: agent.StatusRunning, Duration: "11s"}},
			{Kind: agent.RowFileChange, At: "14:05", Status: agent.StatusDone,
				Agent: "main", Added: 180,
				File: agent.ReviewFile{Path: "ui/agent/well.go", Change: agent.FileAdded}},
			{Kind: agent.RowNote, At: "14:06", Status: agent.StatusWaiting,
				Agent: "main",
				Text:  "Two branches changed the same line in shared.go — the merge cannot decide it."},
			{Kind: agent.RowRequest, At: "14:06", Status: agent.StatusWaiting,
				Agent: "main", Text: "Keep ours or take theirs?"},
		},
	})
	agent.AgentWellRows(c, agent.AgentWellRowsOptions{
		Empty: "Nothing has happened yet.",
	})

	// ── the diffs ──────────────────────────────────────────────────────────
	showcase.Section(c, "改动 · AgentDiff / MultiFileDiffReview / Merge3")

	showcase.Field(c, "AgentDiff — one file the run changed")
	agent.AgentDiff(c, agent.AgentDiffOptions{
		Path: "ui/agent/shared.go", Change: agent.FileModified,
		Old:  "func itoa(n int) string { return strconv.Itoa(n) }\n",
		New:  "func itoa(n int) string {\n\treturn strconv.Itoa(n)\n}\n",
		Lang: code.Go, Height: 96, Context: 2,
		Caption: "gofmt found three lines worth folding",
	})

	showcase.Field(c, "MultiFileDiffReview — both sides merged against the base")
	diffBase := "package agent\n\nfunc itoa(n int) string { return strconv.Itoa(n) }\n\n// signed renders a diff count.\n"
	diffOurs := "package agent\n\nfunc itoa(n int) string {\n\treturn strconv.Itoa(n)\n}\n\n// signed renders a diff count.\n"
	diffTheirs := "package agent\n\nfunc itoa(n int) string { return strconv.FormatInt(int64(n), 10) }\n\n// signed renders a diff count.\n"
	agent.MultiFileDiffReview(c, agent.MultiFileDiffReviewOptions{
		Ours: "main", Theirs: "review", Height: 190,
		// Held across frames, not a fresh map each one: a set made inside the
		// page would lose every signature the moment the frame was over.
		Approved: *showcase.State(c, "agent/review/approved",
			map[string]bool{"ui/agent/status.go": true}),
		Files: []agent.ReviewFile{
			{Path: "ui/agent/shared.go", Change: agent.FileModified,
				Base: diffBase, Ours: diffOurs, Theirs: diffTheirs},
			{Path: "ui/agent/status.go", Change: agent.FileAdded,
				Ours: "package agent\n"},
			{Path: "ui/agent/session.go", Change: agent.FileModified,
				Base: "package agent\n", Ours: "package agent\n\n// Checkpoints.\n"},
			{Path: "ui/agent/draft.go", Change: agent.FileRemoved},
		},
	})

	showcase.Field(c, "Merge3 — 冲突是三行：一个删除，两个新增")
	// Each heading takes the same share as the column under it, which is the
	// only thing that makes a three-column comparison readable: a heading
	// that does not line up with its column is a caption on the wrong text.
	ui.Row(c).FillWidth().Gap(6).Children(func() {
		for _, name := range []string{"base", "ours", "theirs"} {
			ui.Column(c).Grow(1).Shrink(0).Gap(2).Children(func() {
				display.Text(c, name, display.TextOptions{Muted: true, Mono: true, MaxLines: 1})
			})
		}
	})
	ui.Row(c).FillWidth().Gap(6).Children(func() {
		for _, side := range []string{diffBase, diffOurs, diffTheirs} {
			ui.Column(c).Grow(1).Gap(2).Children(func() {
				for _, line := range mergeLines(side) {
					display.Text(c, line, display.TextOptions{Mono: true, MaxLines: 1})
				}
			})
		}
	})

	// A closing note, so the page ends on something that says what the page
	// is rather than on the last component drawn.
	showcase.Section(c, "约定")
	ui.Column(c).FillWidth().Gap(4).Children(func() {
		for _, line := range []string{
			"组件不持有状态：展开、选中、审批、记住，全部是调用方的指针。",
			"状态到颜色是一个纯函数 StatusTone(s Status) core.Severity。",
			"三方合并是一个纯函数 Merge3(base, ours, theirs) []code.DiffLine。",
			"步骤的编号和层级是算出来的：StepNumber(depth) 加调用方的顺序。",
			"给不了的东西 panic(\"agent: …\")，不是画一个空盒子。",
		} {
			display.Text(c, line, display.TextOptions{Muted: true})
		}
	})
}

// mergeLines is a version of a file as its lines, for the page's own
// side-by-side of the three a merge is made from. A blank line is shown as a
// middot because a row of nothing in the middle of three columns is a row
// that reads as a mistake.
func mergeLines(s string) []string {
	var out []string
	for _, line := range strings.Split(strings.TrimSuffix(s, "\n"), "\n") {
		if line == "" {
			line = "·"
		}
		out = append(out, line)
	}
	return out
}
