package pages

import (
	"fmt"
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/git"
	"github.com/HycJack/MintUI/ui/showcase"
	"github.com/HycJack/MintUI/ui/theme"
)

// Demo state outlives the frame: every demo below hands its component
// a pointer, and a pointer into a frame-local is a click the next
// frame undoes — a tab that will not switch, a dropdown that snaps
// shut, a slider that springs back.
var (
	message    = "Widen the commit table's gutter"
	amend      = false
	noneBranch = 0
	loud       = core.Danger
	changesSel = 0
	histSel    = 1
	stashSel   = 0
	tagSel     = "v0.2"
)

func init() {
	showcase.Register(showcase.Page{
		Package: "git",
		Title:   "ui/git — 版本控制里不随窗口变化的那一半",
		Note:    "提交、分支、改动、差异、冲突、评审、引用、命令；字母、行号、计数全部由纯函数先算出来",
		Width:   1000,
		Height:  6380,
		Want: []string{
			// commits
			"Fix the crash on an empty body", "Merge branch 'fix/empty-body'",
			"Ship the gallery page", "Andre Thomson", "Mia Chen", "Ravi Patel",
			"Commit message", "Staged: 4", "Amend last commit", "Commit",
			// branches
			"main", "fix/empty-body", "release/0.4", "origin/main", "Switch branch",
			// changes
			"ui/git/diffview.go", "ui/git/shared.go", "ui/showcase/pages/git.go",
			"Conflicted", "Modified", "Added", "Deleted", "Renamed", "Untracked",
			"Typechange", "no changes",
			// the diff
			"@@ -38,7 +38,10 @@ func DiffViewer(c *ui.Context, unified string, opts DiffViewerOptions) *ui.E",
			"\t\tpanic(\"git: DiffViewer needs a Height; a diff with no height is every line of it\")",
			// merge
			"ours", "theirs", "both", "neither",
			"ours · main", "theirs · fix/empty-body",
			// history and blame
			"BlameView — 作者在左边；WithGutter 才有行号，WithDate 才有绝对日期",
			// review
			"Use PackR instead of dispatch", "Lena Ford", "Resolved",
			// refs
			"stash@{0}", "WIP on the gallery", "tag: v0.2", "annotated",
			// commands
			"$ git diff --cached", "$ git log --oneline", "exit 128",
			"On branch main", "fatal:",
			" not a git repository (or any of the parent directories): .git",
			// the empty half
			"No commits", "No branches", "No changes",
		},
		Render: func(c *ui.Context) {
			gitPage(c)
		},
	})
}

// gitPage is a repository at rest: what it is, what changed, what the two
// sides of a pull request say, and the pure functions underneath all of it.
//
// Everything on it is a fixed little dataset. A gallery page whose commits
// moved would be a page nobody could compare against the last one, and a
// screenshot that changes between runs is a screenshot nobody reviews. So the
// history is six commits, the diff is twelve lines, and the archive of
// conflicts is one — and the components are handed all of it, because none of
// them may read a disk or run a command.
func gitPage(c *ui.Context) {
	commitSection(c)
	branchSection(c)
	changeSection(c)
	diffSection(c)
	wordSection(c)
	mergeSection(c)
	historySection(c)
	reviewSection(c)
	refSection(c)
	commandSection(c)
	gitPureSection(c)
}

// ── the data ────────────────────────────────────────────────────────────────

// gitHistory is six commits, oldest first, which is the order CommitsFor
// wants. It has a branch, a tag and a merge so that the graph has something
// to compute: a linear history draws one lane and proves nothing about the
// algorithm.
var gitHistory = []git.Commit{
	{
		Hash:    "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a90",
		Subject: "First commit", Author: "Andre Thomson", When: "2026-03-02",
		Refs: []string{"tag: v0.1"},
	},
	{
		Hash:    "1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c",
		Subject: "Add the composer", Author: "Andre Thomson", When: "2026-03-14",
		Parents: []string{"0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a90"},
	},
	{
		Hash:    "2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d",
		Subject: "Ship the gallery page", Author: "Mia Chen", When: "2026-10-05",
		Parents: []string{"1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c"},
		Refs:    []string{"main"},
	},
	{
		Hash:    "3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e",
		Subject: "Fix the crash on an empty body", Author: "Ravi Patel", When: "2026-10-06",
		Parents: []string{"1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c"},
		Refs:    []string{"fix/empty-body"},
	},
	{
		Hash:    "4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f",
		Subject: "Widen the commit table's gutter", Author: "Ravi Patel", When: "2026-10-06",
		Parents: []string{"3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e"},
	},
	{
		Hash:    "5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a",
		Subject: "Merge branch 'fix/empty-body'", Author: "Mia Chen", When: "2026-10-06",
		Parents: []string{"2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d", "4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f"},
		Refs:    []string{"main", "tag: v0.2"},
	},
}

var gitBranches = []git.Branch{
	{Name: "main", Current: true, Subject: "Merge branch 'fix/empty-body'"},
	{Name: "fix/empty-body", Ahead: 2, Subject: "Widen the commit table's gutter"},
	{Name: "release/0.4", Subject: "Tag v0.2"},
	{Name: "origin/main", Remote: true, Subject: "Ship the gallery page"},
	{Name: "origin/release/0.4", Remote: true, Behind: 3, Subject: "First commit"},
}

var gitChanges = []git.Change{
	{Path: "ui/git/diffview.go", Status: git.Modified, Added: 6, Deleted: 2},
	{Path: "ui/git/shared.go", Status: git.Conflicted, Added: 3, Deleted: 3},
	{Path: "ui/showcase/pages/git.go", Status: git.Added, Staged: true, Added: 412},
	{Path: "ui/git/ansi.go", Status: git.Renamed, Staged: true, Added: 0, Deleted: 0},
	{Path: "ui/git/draft.go", Status: git.Deleted, Added: 0, Deleted: 96},
	{Path: "ui/git/notes.txt", Status: git.Untracked, Added: 0, Deleted: 0},
}

// gitDiff is a real unified diff: the file headers git writes, and then two
// hunks with line numbers a reader could follow into an editor. It is the
// only input DiffViewer takes, and it is worth noticing that everything it
// draws — both gutters, both counts, the @@ separators — comes out of
// parsing this string.
const gitDiff = `diff --git a/ui/git/diffview.go b/ui/git/diffview.go
index 3a1b2c3..4d5e6f7 100644
--- a/ui/git/diffview.go
+++ b/ui/git/diffview.go
@@ -38,7 +38,10 @@ func DiffViewer(c *ui.Context, unified string, opts DiffViewerOptions) *ui.E
 	k, u := core.Tokens(c), core.Density(c).Unit()
-	if opts.Height <= 0 {
-		panic("git: DiffViewer needs a Height; a diff with no height is every line of it")
+	if opts.Height <= 0 {
+		// A diff has no natural height: for a generated file, a viewer
+		// without one is a window that cannot be closed.
+		panic("git: DiffViewer needs a Height; a diff with no height is every line of it")
 	}
 	lines := SplitDiff(unified)
 	width := gutter(c, lines)
@@ -57,4 +60,5 @@ func DiffViewer(c *ui.Context, unified string, opts DiffViewerOptions) *ui
 	})
 	return view.Element
 }
+
`

var gitConflicts = []git.Conflict{{
	Label: "ui/git/shared.go · hashInk",
	Base: []string{
		"// hashInk is what a hash is drawn in.",
		"func hashInk(c *ui.Context) ui.Color { return k.TextMuted }",
	},
	Ours: []string{
		"// A hash is a reference, not a heading.",
		"func hashInk(c *ui.Context) ui.Color { return k.TextFaint }",
	},
	Theirs: []string{
		"// A hash is a reference, not a heading.",
		"func hashInk(c *ui.Context) ui.Color { return k.TextMuted }",
		"",
		"func hashText(c *ui.Context, hash string) *ui.Element {",
		"\treturn ui.Text(c, hash).TextColor(hashInk(c)).SingleLine()",
		"}",
	},
}}

// The three sides of a merge. The lines are short on purpose: the view gives
// three columns and then four decision buttons, and a line longer than about
// thirty characters is cut at the page's edge rather than wrapped.
var gitMergeBase = []string{
	"func itoa(n int) string {",
	"\treturn itoa(n)",
	"}",
	"",
	"func signed(a, d int) string {",
	"\tif d == 0 {",
	"\t\treturn \"+\" + itoa(a)",
	"\t}",
	"\treturn \"+\" + itoa(a)",
	"}",
}

var gitMergeOurs = []string{
	"func itoa(n int) string {",
	"\treturn itoa(n)",
	"}",
	"",
	"// signed renders a diff count.",
	"func signed(a, d int) string {",
	"\tif d == 0 {",
	"\t\treturn \"+\" + itoa(a)",
	"\t}",
	"\treturn \"+\" + itoa(a)",
	"}",
	"",
}

var gitMergeTheirs = []string{
	"func itoa(n int) string {",
	"\treturn FormatInt(n)",
	"}",
	"",
	"// signed renders a diff count.",
	"func signed(a, d int) string {",
	"\tif d == 0 {",
	"\t\treturn \"+\" + itoa(a)",
	"\t}",
	"\treturn \"+\" + itoa(a)",
	"}",
	"",
}

var gitBlame = []git.BlameLine{
	{Text: "func Hash(c *ui.Context, hash string) {", Author: "Andre", Commit: "0a1b2c3", When: "3 years ago", Date: "2023-04-11"},
	{Text: "\treturn ui.Text(c, hash)", Author: "Andre", Commit: "0a1b2c3", When: "3 years ago", Date: "2023-04-11"},
	{Text: "\t\t.FontSize(MonoSize)", Author: "Mia", Commit: "2c3d4e5", When: "2 days ago", Date: "2026-10-05"},
	{Text: "}", Author: "Mia", Commit: "2c3d4e5", When: "2 days ago", Date: "2026-10-05"},
	{Text: "", Author: "Ravi", Commit: "4e5f6a7", When: "5 hours ago", Date: "2026-10-06"},
	{Text: "// short is what a graph prints.", Author: "Ravi", Commit: "4e5f6a7", When: "5 hours ago", Date: "2026-10-06"},
	{Text: "func (c Commit) short() string {", Author: "Ravi", Commit: "4e5f6a7", When: "5 hours ago", Date: "2026-10-06"},
}

var gitFileHistory = []git.Commit{
	gitHistory[5], gitHistory[4], gitHistory[2],
}

var gitStashes = []git.Stash{
	{Name: "stash@{0}", Message: "WIP on the gallery", Branch: "fix/empty-body", When: "14:12"},
	{Name: "stash@{1}", Message: "before the conflict", Branch: "main", When: "13:58"},
	{Name: "stash@{2}", Message: "experiment with a wider gutter", Branch: "main", When: "Tue"},
}

var gitTags = []git.Tag{
	{Name: "v0.2", Subject: "Merge branch 'fix/empty-body'", Author: "Mia Chen", When: "2026-10-06", Annotated: true},
	{Name: "v0.1", Subject: "First commit", Author: "Andre Thomson", When: "2026-03-02", Annotated: true},
	{Name: "nightly", Subject: "Ship the gallery page", Author: "CI", When: "2026-10-05"},
}

var gitNotes = []git.ReviewNote{
	{
		Author: "Lena Ford", When: "yesterday 16:40",
		Path: "ui/git/diffview.go", Line: 41,
		Body: "This reads well, but the message is the second sentence of a " +
			"comment that is already three lines long. Keep it.",
		Reactions: []git.Reaction{
			{Emoji: "👍", Count: 2, Mine: true},
			{Emoji: "🎉", Count: 1},
		},
	},
	{
		Author: "Sam Ortiz", When: "today 09:12",
		Path: "ui/git/shared.go", Line: 0,
		Body:     "Nothing to do with this hunk — the rename should have moved this function too.",
		Side:     "old",
		Resolved: true,
	},
}

// ── commits ────────────────────────────────────────────────────────────────

func commitSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "提交 · CommitGraph / CommitList / CommitInput")

	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 3)).AlignItems(ui.Start).
		Children(func() {
			ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 1.5)).
				Children(func() {
					showcase.Field(c, "CommitGraph — 泳道是 CommitsFor 从 parents 算出来的")
					// A fixed width for the graph's own rows: the lanes are drawn
					// into it and the subjects share what is left, so a row that
					// filled the whole column would put the last subject against
					// the page's edge.
					ui.Box(c).Width(unit(c, 112)).Shrink(0).Radius(theme.ControlRadius).
						Background(k.Surface).Padding(unit(c, 1)).Children(func() {
						git.CommitGraph(c, gitHistory, git.CommitGraphOptions{
							ShowRefs: true, Selected: gitHistory[5].Hash,
							Width: 3,
						})
					})
				})
			ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 1.5)).
				Children(func() {
					showcase.Field(c, "CommitInput — 消息是调用方的字符串，按钮只报告按过")
					git.CommitInput(c, &message, git.CommitInputOptions{
						Label: "Commit message", Staged: 4, Amend: &amend,
						Placeholder: "What changed, in one line",
						Height:      unit(c, 9),
					})
					ui.Text(c, "空消息时按钮是禁用的：git 本来就拒绝一个空提交，"+
						"按钮提供它等于教错一件事。").TextColor(k.TextFaint).
						FontSize(core.FontSize(c, theme.CaptionSize))
				})
		})

	showcase.Field(c, "CommitList — data.DataTable，四列：主题 / 作者 / 时间")
	selected := 2
	git.CommitList(c, gitHistory, git.CommitListOptions{
		Height: 280, Width: unit(c, 236), Selected: &selected, ShowRefs: true,
	})

	showcase.Field(c, "空列表说什么")
	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 3)).Children(func() {
		ui.Box(c).WidthPercent(share(2)).Shrink(0).Children(func() {
			git.CommitList(c, nil, git.CommitListOptions{
				Height: unit(c, 10), Width: unit(c, 116),
			})
		})
		ui.Box(c).WidthPercent(share(2)).Shrink(0).Children(func() {
			git.CommitGraph(c, nil, git.CommitGraphOptions{})
		})
	})
}

// ── branches ───────────────────────────────────────────────────────────────

func branchSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "分支 · BranchList / BranchSelector")

	showcase.Field(c, "BranchList — 圆点是当前分支，箭头是 ahead / behind")
	selected := 0
	// A row of a branch list gives half its spare room to the tip's subject,
	// so the box has to be wide enough for two of those rather than one: at
	// 116u the subjects are cut at the edge and it reads as a broken panel.
	ui.Box(c).Width(unit(c, 148)).Radius(theme.ControlRadius).
		Background(k.Surface).Padding(unit(c, 1)).Children(func() {
		git.BranchList(c, &selected, gitBranches, git.BranchListOptions{
			Height: unit(c, 34), ShowRemote: true,
		})
	})
	showcase.Field(c, "没有分支时说什么")
	git.BranchList(c, &noneBranch, nil, git.BranchListOptions{Height: unit(c, 8)})

	showcase.Field(c, "BranchSelector — 触发器与面板都是非模态的")
	// The panel is a popover anchored to the trigger's own left edge and
	// drawn over the page rather than inside it, so the stage is the full
	// width of the page and the trigger sits at its left: put the stage in
	// a column and the panel runs off the right edge, because a panel is as
	// wide as the branch rows it holds and those are wide.
	//
	// The stage is tall rather than clipped because the panel lands below the
	// trigger and outside the stage: an open selector has to have somewhere
	// to put the thing it opens, and a stage just tall enough for the
	// trigger is a stage the panel covers the next section with.
	ui.Box(c).Width(unit(c, 236)).Height(unit(c, 66)).Shrink(0).
		Radius(theme.ControlRadius).Background(k.Surface).
		Border(theme.BorderWidth, k.Border).Padding(unit(c, 1.5)).Clip().
		Children(func() {
			ui.Column(c).FillWidth().Gap(unit(c, 1.5)).Children(func() {
				ui.Text(c, "面板挂在触发器下面，宽度是它自己的内容，不是这个盒子").
					TextColor(k.TextFaint).
					FontSize(core.FontSize(c, theme.CaptionSize))
				openBranch := showcase.State(c, "git.389.openBranch", true)
				selBranch := showcase.State(c, "git.389.selBranch", 0)
				git.BranchSelector(c, selBranch, openBranch, gitBranches,
					git.BranchSelectorOptions{Label: "Switch branch"})
			})
		})
}

// ── changes ────────────────────────────────────────────────────────────────

func changeSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "改动 · ChangesList / GitStatusBadge / DiffStat")

	showcase.Field(c, "GitStatusBadge — 字母是 git 的那个字母，底和字由 StatusSeverity 决定")
	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 1.5)).Children(func() {
		for _, s := range []git.FileStatus{
			git.Clean, git.Modified, git.Added, git.Deleted,
			git.Renamed, git.Untracked, git.Conflicted, git.Typechange,
		} {
			git.GitStatusBadge(c, s, git.GitStatusBadgeOptions{WithWord: true})
		}
	})
	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 2)).Children(func() {
		ui.Column(c).Grow(1).Gap(unit(c, 0.75)).Children(func() {
			ui.Text(c, "只有字母").TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize))
			ui.Row(c).Gap(unit(c, 1.5)).Children(func() {
				git.GitStatusBadge(c, git.Modified, git.GitStatusBadgeOptions{})
				git.GitStatusBadge(c, git.Conflicted, git.GitStatusBadgeOptions{})
				git.GitStatusBadge(c, git.Untracked, git.GitStatusBadgeOptions{})
			})
		})
		ui.Column(c).Grow(1).Gap(unit(c, 0.75)).Children(func() {
			ui.Text(c, "Solid · 那一枚该是第一眼落上去的").TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize))
			ui.Row(c).Gap(unit(c, 1.5)).Children(func() {
				git.GitStatusBadge(c, git.Conflicted, git.GitStatusBadgeOptions{
					Severity: &loud, Solid: true,
				})
				git.GitStatusBadge(c, git.Conflicted, git.GitStatusBadgeOptions{
					Severity: &loud,
				})
			})
		})
	})

	showcase.Field(c, "ChangesList — 状态 / 路径 / 行数；暂存的行左边有一道强调色的竖条")
	// Six rows and a head, at the table's own 48-point row height: a table
	// cut through the middle of its last row reads as a clipping bug rather
	// than as "there is more below".
	git.ChangesList(c, gitChanges, git.ChangesListOptions{
		Height: unit(c, 80), Width: unit(c, 236), Selected: &changesSel, WithWord: true,
	})

	showcase.Field(c, "没有改动时说什么")
	git.ChangesList(c, nil, git.ChangesListOptions{
		Height: unit(c, 10), Width: unit(c, 236), WithWord: true,
	})

	showcase.Field(c, "DiffStat — 一串绿点一串红点，宽度由 StatSlices 切，数字由 StatWord 写")
	ui.Column(c).Width(unit(c, 236)).Gap(unit(c, 1)).Children(func() {
		for _, s := range []struct{ added, deleted int }{
			{412, 96}, {6, 2}, {1, 0}, {0, 1}, {3, 3}, {0, 0},
		} {
			ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 3)).AlignItems(ui.Center).
				Children(func() {
					ui.Text(c, git.StatWord(s.added, s.deleted)).TextColor(k.TextMuted).
						Width(unit(c, 18)).
						FontSize(core.FontSize(c, theme.CaptionSize))
					// A box with a width of its own: DiffStat's row is FillWidth
					// and its bar is Grow, so in a wrapper with no width of its
					// own the two resolve to nothing and all eight slots of the
					// bar land on one dot.
					ui.Box(c).Width(unit(c, 214)).Children(func() {
						git.DiffStat(c, s.added, s.deleted, git.DiffStatOptions{})
					})
				})
		}
	})
	ui.Text(c, "StatSlices：两侧都至少占一格，所以一百行新增一行删除也画得出那一点红。").
		TextColor(k.TextFaint).FontSize(core.FontSize(c, theme.CaptionSize))
}

// ── the diff ───────────────────────────────────────────────────────────────

func diffSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "差异 · DiffViewer / Hunks")

	showcase.Field(c, "DiffViewer — 两道行号槽，行号是数出来的不是画上去的")
	// The full content width, because the viewport scrolls sideways and a
	// half-page one turns most of a diff into a horizontal scrollbar: the
	// long lines then have to be scrolled to be read at all, which is the
	// opposite of what a viewer is for.
	git.DiffViewer(c, gitDiff, git.DiffViewerOptions{
		Height: unit(c, 62), File: "ui/git/diffview.go", WithHunks: true,
	})

	showcase.Field(c, "Hunks — @@ 头与两个起点，全是切出来的")
	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 3)).AlignItems(ui.Start).
		Children(func() {
			ui.Box(c).Width(unit(c, 116)).Radius(theme.ControlRadius).
				Background(k.Surface).Padding(unit(c, 1.5)).Children(func() {
				ui.Column(c).FillWidth().Gap(unit(c, 1)).Children(func() {
					for i, h := range git.Hunks(gitDiff) {
						ui.Text(c, fmt.Sprintf("hunk %d", i+1)).TextColor(k.TextFaint).
							FontSize(core.FontSize(c, theme.CaptionSize))
						ui.Text(c, fmt.Sprintf("old %d+%d  new %d+%d",
							h.OldFrom, h.OldCount, h.NewFrom, h.NewCount)).
							TextColor(k.Text).FontSize(core.FontSize(c, theme.CaptionSize))
					}
					added, deleted := git.Counts(git.SplitDiff(gitDiff))
					ui.Text(c, fmt.Sprintf("Counts → +%d −%d", added, deleted)).
						TextColor(k.Text).FontSize(core.FontSize(c, theme.CaptionSize))
					ui.Text(c, "StatWordOf → "+git.StatWordOf(gitDiff)).
						TextColor(k.Text).FontSize(core.FontSize(c, theme.CaptionSize))
				})
			})
			ui.Column(c).Width(unit(c, 116)).Gap(unit(c, 1)).Children(func() {
				ui.Text(c, "Collapsed — 改动两侧各留几行没变的，中间跳过").TextColor(k.Text).
					FontSize(core.FontSize(c, theme.CaptionSize)).Bold()
				ui.Text(c, "WithHunks — 把 @@ 头画成分隔行").TextColor(k.Text).
					FontSize(core.FontSize(c, theme.CaptionSize)).Bold()
				ui.Text(c, "Width / State — 视口自己的尺寸与位置").
					TextColor(k.Text).FontSize(core.FontSize(c, theme.CaptionSize)).Bold()
				ui.Text(c, "DiffViewer 只画 SplitDiff 切出来的行；@@ 头归它开头的那一行，"+
					"不是另走一趟 diff。整块的底是 Background，两侧的行号是 TextFaint，"+
					"改动的整行染 SuccessBg / DangerBg。").TextColor(k.TextFaint).
					FontSize(core.FontSize(c, theme.CaptionSize))
			})
		})
}

// ── word level ─────────────────────────────────────────────────────────────

func wordSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "词级 · InlineDiff / InlineSplit")

	before := "the body was empty, so the parse returned 0 and the view drew a blank card"
	after := "the body was empty, so the parse returned nil and the view drew the empty card"
	showcase.Field(c, "InlineDiff — 改掉的词加下划线，不是整行涂色")
	git.InlineDiff(c, before, after, git.InlineDiffOptions{Width: unit(c, 236)})

	showcase.Field(c, "InlineSplit 切出来的词：左边 removed，右边 added，Changed 才是画下划线的那几个")
	removed, added := git.InlineSplit(before, after)
	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 3)).AlignItems(ui.Start).
		Children(func() {
			inlineWords(c, "removed", removed, k.Danger)
			inlineWords(c, "added", added, k.Success)
		})

	showcase.Field(c, "WithLineNumbers · 每个词旁边是它在那一侧里的序号")
	git.InlineDiff(c, "port 8080", "port 8443", git.InlineDiffOptions{
		Width: unit(c, 236), WithLineNumbers: true,
	})
}

// inlineWords is one side of an inline diff, with the changed words marked, so
// a reader can see the split rather than trust it.
func inlineWords(c *ui.Context, title string, words []git.InlineWord, ink ui.Color) {
	k := core.Tokens(c)
	ui.Column(c).Width(unit(c, 116)).Shrink(0).Gap(unit(c, 0.5)).Children(func() {
		ui.Text(c, title).TextColor(k.Text).
			FontSize(core.FontSize(c, theme.CaptionSize)).Bold()
		ui.Row(c).FillWidth().Gap(unit(c, 0.5)).Wrap().Children(func() {
			for _, w := range words {
				bg := k.Surface
				mark := ink
				if w.Changed {
					bg = k.SurfaceHover
					mark = ink
				}
				ui.Box(c).Radius(theme.SmallRadius).Background(bg).
					Padding(unit(c, 0.25), unit(c, 0.5)).Children(func() {
					ui.Text(c, strings.TrimSpace(w.Text)).TextColor(mark).
						FontSize(core.FontSize(c, theme.CaptionSize))
				})
			}
		})
	})
}

// ── merge ──────────────────────────────────────────────────────────────────

func mergeSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "合并 · ConflictResolver / ThreeWayMerge")

	showcase.Field(c, "ConflictResolver — 四个按钮每个 hunk 都在，选中的那个戴强调色")
	choices := []git.Resolution{git.KeepOurs, git.KeepBoth, git.Unresolved, git.KeepNeither}
	git.ConflictResolver(c, gitConflicts, git.ConflictResolverOptions{
		Choices: choices, Width: unit(c, 236),
		Ours: "ours · main", Theirs: "theirs · fix/empty-body",
	})

	showcase.Field(c, "Resolution.Apply — 解决的结果是算出来的，不是画出来的")
	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 2)).AlignItems(ui.Start).
		Children(func() {
			for _, r := range []git.Resolution{
				git.KeepOurs, git.KeepTheirs, git.KeepBoth, git.KeepNeither,
			} {
				lines := r.Apply(gitConflicts[0])
				ui.Column(c).Grow(1).Gap(unit(c, 0.75)).Children(func() {
					ui.Box(c).Padding(unit(c, 0.25), unit(c, 1)).Radius(theme.PillRadius).
						Background(k.Surface).Children(func() {
						ui.Text(c, r.String()).TextColor(k.Text).
							FontSize(core.FontSize(c, theme.CaptionSize)).Bold()
					})
					ui.Column(c).FillWidth().Gap(0).Children(func() {
						for _, line := range lines {
							ui.Text(c, line).TextColor(k.TextMuted).
								FontSize(core.FontSize(c, theme.CaptionSize)).MaxLines(1)
						}
						if len(lines) == 0 {
							ui.Text(c, "（这个 hunk 从文件里消失）").TextColor(k.TextFaint).
								FontSize(core.FontSize(c, theme.CaptionSize))
						}
					})
				})
			}
		})

	showcase.Field(c, "ThreeWayMerge — base / ours / theirs 三列对齐，右侧一列按钮")
	// One resolution per row of ours, which is what the component asks for:
	// it walks ours and writes choices[i], so a shorter slice is a panic
	// rather than a row with no button.
	mergeChoices := make([]git.Resolution, len(gitMergeOurs))
	for i := range mergeChoices {
		mergeChoices[i] = []git.Resolution{
			git.KeepOurs, git.KeepTheirs, git.KeepBoth, git.KeepNeither,
		}[i%4]
	}
	git.ThreeWayMerge(c, git.ThreeWayMergeOptions{
		Height: unit(c, 30), Base: gitMergeBase, Ours: gitMergeOurs,
		Theirs: gitMergeTheirs, OursLabel: "main", TheirsLabel: "review",
		Choices: mergeChoices,
	})
}

// ── history and blame ──────────────────────────────────────────────────────

// historySection puts the two tables one under the other at the full content
// width rather than side by side. Both of them have a column of words that
// only reads when it is not clipped — a commit subject and a line of blame —
// and halving the width halves what fits in them, so a side-by-side pair
// shows three words and a scrollbar where the whole story would have fit.
func historySection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "历史 · FileHistory / BlameView")

	showcase.Field(c, "FileHistory — 一个文件的日志，就是列更少的 CommitList")
	git.FileHistory(c, "ui/git/diffview.go", gitFileHistory, git.FileHistoryOptions{
		Height: unit(c, 44), Width: unit(c, 236), Selected: &histSel,
		Path: "ui/git/diffview.go",
	})
	ui.Text(c, "这个文件没有历史时说什么").TextColor(k.TextMuted).
		FontSize(core.FontSize(c, theme.CaptionSize))
	git.FileHistory(c, "ui/git/never-existed.go", nil, git.FileHistoryOptions{
		Height: unit(c, 8), Width: unit(c, 116),
		Path: "ui/git/never-existed.go",
	})

	showcase.Field(c, "BlameView — 作者在左边；WithGutter 才有行号，WithDate 才有绝对日期")
	ui.Box(c).Width(unit(c, 236)).Radius(theme.ControlRadius).
		Background(k.Surface).Padding(unit(c, 1)).Children(func() {
		git.BlameView(c, gitBlame, git.BlameViewOptions{
			Height: unit(c, 44), WithGutter: true, WithDate: true,
		})
	})
}

// ── review ─────────────────────────────────────────────────────────────────

func reviewSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "评审 · PullRequestCard / ReviewComment / ReviewNote")

	showcase.Field(c, "PullRequestCard — 整张卡是那一下点击；Merge / Close 是它的子控件")
	// One card per row, at the full content width, and no Height: a card
	// given a height smaller than its content is squeezed until the title is
	// five lines and the action row lands on top of it. Three across a
	// 944-point page leaves 300 a piece, which is narrower than a title.
	action := func(labels ...string) func() {
		return func() {
			ui.Row(c).Gap(unit(c, 1.5)).Children(func() {
				for i, label := range labels {
					bg, fg := k.Surface, k.Text
					if i == 0 {
						bg, fg = k.Fill, k.OnFill
					}
					ui.Box(c).Height(unit(c, 6)).Padding(0, unit(c, 2.5)).
						Radius(theme.PillRadius).Background(bg).Center().
						Children(func() {
							ui.Text(c, label).TextColor(fg).
								FontSize(core.FontSize(c, theme.CaptionSize))
						})
				}
			})
		}
	}
	ui.Column(c).Width(unit(c, 236)).Gap(unit(c, 2)).Children(func() {
		git.PullRequestCard(c, git.PullRequest{
			Number: 412, Title: "Use PackR instead of dispatch",
			Author: "Ravi Patel", From: "fix/empty-body", To: "main",
			State: "open", Comments: 6, Reviews: 2,
			ChecksPassed: 20, ChecksTotal: 21, ChecksFailed: 1,
		}, git.PullRequestCardOptions{
			Actions: action("Merge", "Close", "Convert to draft"),
		})
		git.PullRequestCard(c, git.PullRequest{
			Number: 409, Title: "Bump the monospaced size to 11.5",
			Author: "Mia Chen", From: "release/0.4", To: "main",
			State: "merged", Comments: 2, Reviews: 1,
			ChecksPassed: 21, ChecksTotal: 21,
		}, git.PullRequestCardOptions{})
		git.PullRequestCard(c, git.PullRequest{
			Number: 405, Title: "Nightly: rebuild the gallery",
			Author: "CI", From: "nightly", To: "main",
			State: "closed", Draft: true, ChecksPassed: 18, ChecksTotal: 21,
			ChecksFailed: 3,
		}, git.PullRequestCardOptions{})
	})

	showcase.Field(c, "ReviewComment — 挂在它说的那一行上；已解决的变灰但不消失")
	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 3)).AlignItems(ui.Start).
		Children(func() {
			for i := range gitNotes {
				note := gitNotes[i]
				git.ReviewComment(c, note, git.ReviewCommentOptions{
					Width: unit(c, 116), Reactions: true,
					Action: func() {
						ui.Text(c, "Resolve / Reply / ...").TextColor(k.TextFaint).
							FontSize(core.FontSize(c, theme.CaptionSize))
					},
				})
			}
		})

	showcase.Field(c, "ReviewNote.Reactions — 自己点过的那个是实心的")
	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 2)).Children(func() {
		for _, r := range gitNotes[0].Reactions {
			mine := "别人的"
			if r.Mine {
				mine = "我点过的"
			}
			ui.Column(c).Gap(unit(c, 0.5)).AlignItems(ui.Center).Children(func() {
				ui.Text(c, r.Emoji).TextColor(k.Text).
					FontSize(core.FontSize(c, theme.RowSize))
				ui.Text(c, fmt.Sprintf("%d · %s", r.Count, mine)).
					TextColor(k.TextMuted).FontSize(core.FontSize(c, theme.CaptionSize))
			})
		}
		ui.Text(c, "Reaction 是数据，ReviewComment 才是画它的那一个；"+
			"emoji 来自大家认得的那一套，不另画。").TextColor(k.TextFaint).
			FontSize(core.FontSize(c, theme.CaptionSize))
	})
}

// ── refs ───────────────────────────────────────────────────────────────────

func refSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "引用 · StashList / TagList")

	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 3)).AlignItems(ui.Start).
		Children(func() {
			ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 1.5)).Children(func() {
				showcase.Field(c, "StashList — stash@{0} 写在表里，因为它就是错误信息里那个名字")
				git.StashList(c, gitStashes, git.StashListOptions{
					Height: unit(c, 44), Width: unit(c, 116), Selected: &stashSel,
				})
				git.StashList(c, nil, git.StashListOptions{
					Height: unit(c, 8), Width: unit(c, 116),
				})
			})
			ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 1.5)).Children(func() {
				showcase.Field(c, "TagList — 有注释的标签戴告警色，因为那是一条 somebody 写过的消息")
				git.TagList(c, gitTags, git.TagListOptions{
					Height: unit(c, 44), Width: unit(c, 116),
					Selected: &tagSel, ShowPrefix: true,
				})
				git.TagList(c, nil, git.TagListOptions{
					Height: unit(c, 8), Width: unit(c, 116),
				})
				ui.Row(c).Gap(unit(c, 1.5)).AlignItems(ui.Center).Children(func() {
					ui.Box(c).Padding(unit(c, 0.25), unit(c, 1.25)).
						Radius(theme.PillRadius).Background(k.WarningBg).Children(func() {
						ui.Text(c, "annotated").TextColor(k.Warning).
							FontSize(core.FontSize(c, theme.CaptionSize))
					})
					ui.Box(c).Padding(unit(c, 0.25), unit(c, 1.25)).
						Radius(theme.PillRadius).Background(k.AccentBg).Children(func() {
						ui.Text(c, "lightweight").TextColor(k.AccentText).
							FontSize(core.FontSize(c, theme.CaptionSize))
					})
					ui.Text(c, "带注释的那 somebody 写过一段话").
						TextColor(k.TextFaint).FontSize(core.FontSize(c, theme.CaptionSize))
				})
			})
		})
}

// ── commands ───────────────────────────────────────────────────────────────

func commandSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "命令 · CommandBlock / AnsiText")

	showcase.Field(c, "CommandBlock — 命令在上面，输出在下面，失败的挂一个 exit 码")
	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 3)).AlignItems(ui.Start).
		Children(func() {
			ui.Box(c).WidthPercent(share(2)).Shrink(0).Children(func() {
				git.CommandBlock(c, git.CommandBlockOptions{
					Command: "git log --oneline",
					Height:  unit(c, 26),
					Result: git.CommandResult{Code: 0, Out: strings.Join([]string{
						"5f6a7b8 Merge branch 'fix/empty-body'",
						"4e5f6a7 Widen the commit table's gutter",
						"3d4e5f6 Fix the crash on an empty body",
						"2c3d4e5 Ship the gallery page",
					}, "\n")},
				})
			})
			ui.Box(c).WidthPercent(share(2)).Shrink(0).Children(func() {
				git.CommandBlock(c, git.CommandBlockOptions{
					Command: "git diff --cached",
					Height:  unit(c, 26), Collapsed: 3,
					Result: git.CommandResult{
						Code: 128,
						Out:  "diff --git a/ui/git/draft.go b/ui/git/draft.go\ndeleted file mode 100644\nindex 3a1b2c3..0000000",
						Err:  "\x1b[1;31mfatal:\x1b[0m not a git repository (or any of the parent directories): .git",
					},
				})
			})
		})

	showcase.Field(c, "AnsiText — git 自己的颜色被保留下来，用这一套调色板重画")
	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 3)).AlignItems(ui.Start).
		Children(func() {
			ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 0.5)).Children(func() {
				ui.Text(c, "git status --short").TextColor(k.Text).
					FontSize(core.FontSize(c, theme.CaptionSize)).Bold()
				git.AnsiText(c, strings.Join([]string{
					"\x1b[32mOn branch main\x1b[0m",
					"\x1b[32mYour branch is ahead of 'origin/main' by 1 commit.\x1b[0m",
					"",
					"\x1b[31mmodified:   ui/git/diffview.go\x1b[0m",
					"\x1b[31mmodified:   ui/git/shared.go\x1b[0m",
					"\x1b[31mconflicted: ui/git/notes.txt\x1b[0m",
					"\x1b[1;33muntracked: ui/git/draft.txt\x1b[0m",
				}, "\n"), git.AnsiTextOptions{Mono: true})
			})
			ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 0.5)).Children(func() {
				ui.Text(c, "StripANSI 之后的样子（测量、搜索都用它）").TextColor(k.Text).
					FontSize(core.FontSize(c, theme.CaptionSize)).Bold()
				git.AnsiText(c, git.StripANSI(
					"\x1b[31mmodified:\x1b[0m   ui/git/diffview.go\n\x1b[1;33muntracked:\x1b[0m ui/git/draft.txt"),
					git.AnsiTextOptions{Mono: true})
			})
		})
}

// ── the pure half ──────────────────────────────────────────────────────────

// gitPureSection is everything in this package that decides rather than draws.
//
// It is at the end of the page because it is the reason the rest of the page
// is the way it is: a letter, a line number, a count, a lane and a word-level
// split are all answers to a question about data, and every one of them is
// asked before a single element is made.
func gitPureSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "纯函数 — 先算出答案，再画")

	showcase.Field(c, "CommitsFor — 泳道从 parents 反着走一遍算出来")
	ui.Box(c).Width(unit(c, 236)).Radius(theme.ControlRadius).
		Background(k.Surface).Padding(unit(c, 1.5)).Children(func() {
		lanes := gitColWidths(40, 20, 20, 26, 30)
		gitPureRow(c, lanes, "hash", "col", "merge", "lanes below", "width")
		for _, r := range git.CommitsFor(gitHistory, 0) {
			gitPureRow(c, lanes, r.Commit.Hash[:7], gitItoa(r.Column),
				gitYesNo(r.Merge), fmt.Sprintf("%v", r.Lanes), gitItoa(r.Width))
		}
		ui.Text(c, "commits 是最老在前：CommitsFor 自己反着读回去，"+
			"调用方反一次就够，错一次画出来却看着像对的。").TextColor(k.TextFaint).
			FontSize(core.FontSize(c, theme.CaptionSize))
	})

	showcase.Field(c, "SplitDiff — 行号从 @@ 头数出来，加进来的行旧侧是 0")
	// Full width, because the text column is the one worth reading and a
	// half-page table spends most of its width on a two-digit number.
	split := gitColWidths(16, 14, 14, 178)
	gitPureRow(c, split, "op", "old", "new", "text")
	for _, l := range git.SplitDiff(gitDiff) {
		gitPureRow(c, split, gitOpWord(l.Op), gitItoa(l.OldNo),
			gitItoa(l.NewNo), l.Text)
	}

	showcase.Field(c, "Hunks / Counts / StatWord / StatSlices — 同一段 diff 的另外三个读法")
	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 4)).AlignItems(ui.Start).
		Children(func() {
			ui.Column(c).WidthPercent(share(2)).Shrink(0).Children(func() {
				hunkCols := gitColWidths(44, 18, 14, 18)
				gitPureRow(c, hunkCols, "hunk header", "old from", "old n", "new from")
				for _, h := range git.Hunks(gitDiff) {
					gitPureRow(c, hunkCols, h.Header, gitItoa(h.OldFrom),
						gitItoa(h.OldCount), gitItoa(h.NewFrom))
				}
			})
			ui.Column(c).WidthPercent(share(2)).Shrink(0).Children(func() {
				sum := gitColWidths(44, 30, 28)
				a, d := git.Counts(git.SplitDiff(gitDiff))
				gitPureRow(c, sum, "Counts(SplitDiff)", gitItoa(a), gitItoa(d))
				gitPureRow(c, sum, "StatWord(4, 1)", git.StatWord(4, 1), "")
				gitPureRow(c, sum, "StatWord(0, 0)", git.StatWord(0, 0), "")
				gitPureRow(c, sum, "StatWordOf(gitDiff)", git.StatWordOf(gitDiff), "")
				adds, dels := git.StatSlices(4, 1, 8)
				gitPureRow(c, sum, "StatSlices(4, 1, 8)", fmt.Sprintf("%d / %d", adds, dels), "槽")
				adds, dels = git.StatSlices(100, 1, 8)
				gitPureRow(c, sum, "StatSlices(100, 1, 8)", fmt.Sprintf("%d / %d", adds, dels), "槽")
				gitPureRow(c, sum, "StatSlices(0, 0, 8)", "0 / 0", "槽")
			})
		})

	showcase.Field(c, "状态映射 — 字母、整词、音量、墨色，一处算完四处画")
	// The pair is drawn in the same row as the names it belongs to: a pill
	// in a box of its own lands at the left edge of the column above it,
	// which reads as a second table rather than as a fifth column.
	cols := gitColWidths(24, 28, 26, 28)
	ui.Box(c).Width(unit(c, 236)).Radius(theme.ControlRadius).
		Background(k.Surface).Padding(unit(c, 1.5)).Children(func() {
		ui.Column(c).FillWidth().Gap(unit(c, 0.5)).Children(func() {
			ui.Row(c).Gap(unit(c, 1.5)).Children(func() {
				for i, name := range []string{
					"FileStatus", "StatusLetter", "StatusWord", "StatusSeverity",
				} {
					ui.Text(c, name).TextColor(k.TextMuted).Width(unit(c, cols[i])).
						FontSize(core.FontSize(c, theme.CaptionSize)).Bold()
				}
				ui.Text(c, "StatusPair · StatusInk").TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize)).Bold()
			})
			for _, st := range []git.FileStatus{
				git.Clean, git.Modified, git.Added, git.Deleted,
				git.Renamed, git.Untracked, git.Conflicted, git.Typechange,
			} {
				ui.Row(c).Gap(unit(c, 1.5)).AlignItems(ui.Center).Children(func() {
					for i, cell := range []string{
						gitItoa(int(st)), gitOrDash(git.StatusLetter(st)),
						git.StatusWord(st), fmt.Sprintf("%v", git.StatusSeverity(st)),
					} {
						ui.Text(c, cell).TextColor(k.TextMuted).Width(unit(c, cols[i])).
							FontSize(core.FontSize(c, theme.CaptionSize)).MaxLines(1)
					}
					bg, fg := git.StatusPair(st, k)
					ui.Box(c).Padding(unit(c, 0.25), unit(c, 1)).
						Radius(theme.PillRadius).Background(bg).Children(func() {
						ui.Text(c, st.String()+" · ink").TextColor(fg).
							FontSize(core.FontSize(c, theme.CaptionSize))
					})
				})
			}
		})
	})

	showcase.Field(c, "ANSI — ParseANSI 留代码，StripANSI 留字，AnsiInk 把代码说成这套调色板里的一个颜色")
	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 4)).AlignItems(ui.Start).
		Children(func() {
			ui.Column(c).WidthPercent(share(2)).Shrink(0).Children(func() {
				ui.Text(c, "ParseANSI(\"\\x1b[1;31mfatal:\\x1b[0m not a repository\")").
					TextColor(k.Text).FontSize(core.FontSize(c, theme.CaptionSize)).Bold()
				spans := gitColWidths(46, 40)
				gitPureRow(c, spans, "text", "codes")
				for _, s := range git.ParseANSI("\x1b[1;31mfatal:\x1b[0m not a repository") {
					gitPureRow(c, spans, fmt.Sprintf("%q", s.Text), fmt.Sprintf("%v", s.Codes))
				}
				strip := gitColWidths(46, 78)
				gitPureRow(c, strip, "StripANSI",
					fmt.Sprintf("%q", git.StripANSI("\x1b[1;31mfatal:\x1b[0m not a repository")))
				gitPureRow(c, strip, "ValidUTF8(\"ok\")", gitYesNoBool(git.ValidUTF8("ok")))
				gitPureRow(c, strip, "ValidUTF8(0xff)", gitYesNoBool(git.ValidUTF8("\xff")))
			})
			ui.Column(c).WidthPercent(share(2)).Shrink(0).Children(func() {
				ui.Text(c, "AnsiInk(code, k) — 十六个 ANSI 色落到四个 token 上").
					TextColor(k.Text).FontSize(core.FontSize(c, theme.CaptionSize)).Bold()
				ink := gitColWidths(24, 28, 18, 18, 36)
				gitPureRow(c, ink, "code", "ink", "bold", "ok", "")
				for _, code := range []int{1, 31, 32, 33, 34, 35, 36, 39, 7} {
					col, bold, ok := git.AnsiInk(code, k)
					gitPureRow(c, ink, gitItoa(code), gitTokenName(c, col, ok),
						gitYesNoBool(bold), gitYesNoBool(ok), "")
				}
			})
		})

	ui.Column(c).FillWidth().Gap(unit(c, 1)).Children(func() {
		for _, line := range []string{
			"组件不跑 git、不读磁盘：提交列表、文件树、diff、归档条目都是参数。",
			"状态到颜色是 StatusSeverity，颜色本身是 StatusInk / StatusPair。",
			"解决冲突是 Resolution.Apply 的算术，不是画出来的。",
			"给不了的东西 panic(\"git: …\")，不是画一个空盒子。",
		} {
			ui.Text(c, line).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize))
		}
	})
}

// gitColWidths is the column widths of one of the tables above, in spacing units.
//
// They are written out rather than divided up because the first column is a
// name and the rest are values: splitting the row evenly gives the name the
// same room as a line number and squeezes the text a reader came for.
func gitColWidths(w ...float32) []float32 { return w }

// gitPureRow is one line of a pure function's output: the values it returned,
// each in a column of stated width. Nothing here grows, because a row that
// fills its column and then grows its last one pushes that one off the edge of
// the page — and the value it clips is the value somebody was looking for.
func gitPureRow(c *ui.Context, cols []float32, cells ...string) {
	k := core.Tokens(c)
	const gap = 1.5
	ui.Row(c).Gap(unit(c, gap)).Children(func() {
		for i, cell := range cells {
			if i >= len(cols) {
				break
			}
			ui.Text(c, cell).TextColor(k.TextMuted).Width(unit(c, cols[i])).
				FontSize(core.FontSize(c, theme.CaptionSize)).MaxLines(1)
		}
	})
}

// gitOpWord is the DiffOp as a word rather than as the one character it prints.
// A column of spaces and plus signs is not readable, and the character is
// already on the line in DiffViewer.
func gitOpWord(op git.DiffOp) string {
	switch op {
	case git.OpAdd:
		return "+ 新增"
	case git.OpDel:
		return "− 删除"
	}
	return "· 上下文"
}

// gitTokenName is what AnsiInk returned, as the token's name rather than its
// hex: the table is about the mapping, and the mapping is a name.
func gitTokenName(c *ui.Context, col ui.Color, ok bool) string {
	if !ok {
		return "—（不管颜色）"
	}
	k := core.Tokens(c)
	for _, pair := range []struct {
		name string
		col  ui.Color
	}{
		{"Text", k.Text}, {"Danger", k.Danger}, {"Success", k.Success},
		{"Warning", k.Warning}, {"Accent", k.Accent},
	} {
		if pair.col == col {
			return pair.name
		}
	}
	return "别的"
}

// gitOrDash keeps a column of values aligned: an empty string in a table reads
// as a rendering bug, a dash reads as "there is nothing here".
func gitOrDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

// gitItoa and gitYesNoBool are the two conversions the tables above want,
// in one place, because every value in them arrived as an int or a bool.
func gitItoa(n int) string { return fmt.Sprintf("%d", n) }

func gitYesNoBool(b bool) string {
	if b {
		return "是"
	}
	return "否"
}

// gitYesNo is gitYesNoBool under the name these tables use.
func gitYesNo(b bool) string { return gitYesNoBool(b) }
