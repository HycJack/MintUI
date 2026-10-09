package pages

import (
	"fmt"
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/code"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/showcase"
	"github.com/HycJack/MintUI/ui/theme"
)

// The page is one development window, in the order a person uses it: open a
// file, read it, see what changed, run the thing and read its output, stop in
// it and read the variables, and profile it.
//
// The order is a session rather than an index of files, because that is the
// only order in which the package reads as one thing: a viewer, a diff and a
// hex view that share one set of rows is the point of ui/code, and a gallery
// that listed the panels alphabetically would hide it.
//
// Everything on the page is a fixed little dataset — the same thirty-six lines
// of Go, the same seven log lines, the same four breakpoints — and none of it
// is read off a disk or run as a command. A gallery page whose content moved
// could not be compared against the last one, and a screenshot that changes
// between two runs is a screenshot nobody reviews.

func init() {
	showcase.Register(CodePage())
}

// CodePage is the gallery's page for ui/code.
func CodePage() showcase.Page {
	return showcase.Page{
		Package: "code",
		Title:   "ui/code — 一个开发窗口读代码的那一半",
		Note:    "源码、差异、终端、构建输出、断点、火焰图；代码和日志全部由调用方传入，纯函数先算出答案再画",
		Width:   1000,
		Height:  5380,
		Want: []string{
			"文件 · CodeViewer / Minimap",
			"CodeViewer — 行号、当前行、查找标记、折叠箭头都出自同一套 codeRow",
			"Minimap — 同一份 Source 的侧面",
			"面板的头 · CodePanel / CodeHeader / CodeBadge / CodeCaption",
			"CodeHeader — 一行：标题让位，计数不让位",
			"CodeBadge — 严重度是 core.Severity 给的，组件自己不挑颜色",
			"CodeCaption / CodeMono / CodeSelectMark / CodeMetrics — 面板周围的字",
			"六门语言，加一种不是语言的 · 同一个组件，Lang 换一下",
			"差异 · Diff / DiffLine",
			"Unified — 一列，两道行号槽",
			"两种形状之外没有第三种答案",
			"Split — 两列，一行的高度放得下两边",
			"终端 · Terminal / TerminalTabs / TerminalSearch / CommandHistory",
			"TerminalTabs — 选中的那个是抬起来的，不是一个药丸",
			"Terminal — 尾部是提示符，Level 是调用方给的一行一个",
			"TerminalSearch — 终端没有列表可给，只有两个数",
			"CommandHistory — 最新的在下面",
			"构建 · OutputPanel / ProblemsPanel / LogViewer",
			"OutputPanel — 头部说还在跑，坏行染 DangerBg",
			"ProblemsPanel — 按说的顺序，不按严重度排序",
			"LogViewer — 左边的时间是调用方的格式，右边的来源只在要的时候画",
			"不过滤",
			"Filter: Warn — 只剩三行，头上的计数也跟着变",
			"查找 · FindWidget / SearchPanel / CodeMatchTokens",
			"FindWidget — 计数在条子里，因为「4/212」才是答案，框里的字只是一半",
			"SearchPanel — 每一条都带着它命中的那一行",
			"CodeMatchTokens — 行外的命中长一个样子",
			"调试 · BreakpointList / CallStack / VariablesPanel / SymbolOutline",
			"CallStack — 最内层在上面，就是人读程序的反过来",
			"VariablesPanel — 值是调用方自己的格式",
			"BreakpointList — 关掉的画删除线，因为还能再打开",
			"SymbolOutline — 内容是调用方的树，本包不解析语法",
			"运行 · DebugToolbar / ProcessList / CompletionMenu",
			"DebugToolbar — 状态贴在行尾，不夹在两个按钮中间",
			"ProcessList — CPU 与内存是调用方给的字符串",
			"CompletionMenu — 菜单归它要补的那个字段，不归窗口",
			"字节与剖析 · HexViewer / Flamegraph",
			"HexViewer — 偏移、两组字节、可打印的字符，三处对齐",
			"Flamegraph — 宽度是调用方的比例，所以宽度是必须的",
			"纯函数 — 先算出答案，再画",
			"Highlight — 一行切成若干 run，每个 run 一个 TokenKind",
			"HighlightSpans — 只有种类没有文字；HighlightAll — 一次给一整个文件",
			"FindMatches / CodeMatchTokens — 偏移是字节，标记按偏移切开",
			"DiffLines — 一段文本变成若干行，每行三件事",
			"ParseAnsi — 转义序列拿掉，剩下的按程序要的画",
			"TokenInk / TokenKind — 种类到墨色的一处映射，图例和它读的是同一个函数",
			"SymbolKind / LogLevel / DiffLineKind — 三组枚举，都带一个 String 和一个严重度",
			"CodeRowsOptions — 每个画行的组件共用这一份参数",
			"ui/code/viewer.go",
			"ui/code/frame.png",
			// what the components actually drew
			"main · 1 fold shut", "strings.Split", "package",
			"sample.go", "sample.ts", "app.py", "lib.rs", "queries.sql", "build.sh",
			"notes.txt", "Go", "JavaScript", "Python", "Rust", "SQL", "Shell", "LangPlain",
			"ui/code/board.go", "ui/code/shared.go · 4e5f6a7",
			"Terminals", "zsh", "server", "Find in terminal", "1/12", "History",
			"go test ./ui/code/... -count=1", "build", "go vet ./ui/code/...",
			"declared and not used: rows", "shots: 21 pages, scale 2",
			"Find in file", "Replace", "Search results", "3 to replace",
			"Breakpoints", "ui/code/viewer.go:104", "len(runes) == 0",
			"Call stack", "Variables", "Outline",
			"Continue", "Step over", "Stop", "Stopped at line 104",
			"Processes", "cmd/gallery", "Completions", "Flamegraph — callbacks",
			"Keyword", "Type", "String", "Number", "Comment", "Function", "Punct",
			"AnsiDefault", "AnsiRed", "Trace", "Debug", "Info", "Warn", "Error",
			"Added", "Removed", "Same", "Func", "Value", "Section",
		},
		Render: func(c *ui.Context) {
			codePage(c)
		},
	}
}

func codePage(c *ui.Context) {
	viewerSection(c)
	chromeSection(c)
	langSection(c)
	codeDiffSection(c)
	terminalSection(c)
	buildSection(c)
	findSection(c)
	debugSection(c)
	runSection(c)
	profileSection(c)
	codePureSection(c)
}

// ── the data ────────────────────────────────────────────────────────────────

// codeSource is the excerpt the viewer and the minimap both read, so the two
// halves of that section cannot drift apart: they are handed the same string.
//
// It is real code from this repository rather than filler, because a page of
// coloured placeholder words cannot show whether the highlighter is right —
// there is nothing in it to be right about. Thirty-six lines is enough for a
// fold, a caret and a search to all have somewhere to be at once.
const codeSource = `package code

import (
	"strings"

	"github.com/egoist/mygo/ui"
)

// splitLines is the text as lines, without inventing
// a last empty one.
func splitLines(s string) []string {
	if s == "" {
		return []string{""}
	}
	lines := strings.Split(s, "\n")
	if n := len(lines); n > 0 && lines[n-1] == "" {
		lines = lines[:n-1]
	}
	for i, l := range lines {
		lines[i] = strings.TrimSuffix(l, "\r")
	}
	return lines
}

// foldAt reports whether a line is inside a
// block somebody has shut.
func foldAt(folds []Fold, line int) bool {
	for _, f := range folds {
		if f.Line >= line {
			break
		}
		if !f.Hidden {
			continue
		}
		return true
	}
	return false
}

// metricsOf is the measure of code at this window's
// density, given how many lines there are.
func metricsOf(c *ui.Context, lines int) metrics {
	m := metrics{pad: 8, size: 13.5}
	digits := len(itoa(max(lines, 1)))
	m.gutter = float32(digits) * m.size * 0.62
	return m
}

// itoa is a small one. Every line number in this
// package goes through it.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}`

// codeLines is how many lines the viewer above is showing, read from the
// source itself rather than typed in beside it: the viewer's own head counts
// them too, and two numbers for one file is one more place to be wrong.
var codeLines = len(strings.Split(codeSource, "\n"))

// plainSource is what a viewer draws for a file it has never heard of: not
// code, not prose, a config file — and LangPlain still finds the string and
// the number in it, which is the two things a reader of a config looks for.
const plainSource = `theme = "dark"
editor.font_size = 13
# the window follows the system
accent = "#d9f24b"`

// codeFolds is one shut block. Which lines can be folded is a question about a
// language's braces and the folds are the caller's slice, so the viewer is
// told rather than asked: line 13 opens foldAt's loop and the nine lines after
// it are gone until it is opened again.
var codeFolds = []code.Fold{{Line: 27, Hidden: true}}

// The two versions the diff compares. The change is one argument and one
// whole new line, which is a diff with both a small hunk and a big one in it.
const diffOld = `func Board(c *ui.Context, cards []Card) {
	ui.Column(c).Children(func() {
		for _, card := range cards {
			drawCard(c, card)
		}
	})
}`

const diffNew = `func Board(c *ui.Context, cards []Card) {
	ui.Column(c).Gap(8).Children(func() {
		for _, card := range cards {
			drawCard(c, card)
		}
		ui.Text(c, plural(len(cards)))
	})
}`

// codeTerminalLines is a shell that ran the suite and then the gallery, and
// found something wrong with itself. The colours are in the string: the
// component draws what it is handed, which is the only reason a build tool's
// own red reaches the screen in this library's palette.
var codeTerminalLines = []string{
	"$ cd ~/work/callbacks",
	"$ \x1b[1mgo test ./ui/code/... -count=1\x1b[0m",
	"ok  \tcallbacks/ui/code\t1.284s",
	"$ go run ./cmd/gallery -fit",
	"code 4600",
	"$ go build ./...",
	"# callbacks/cmd/gallery",
	"\x1b[31mmain.go:112:12: undefined: renderAt\x1b[0m",
	"\x1b[31mFAIL\tcallbacks/cmd/gallery\t0.412s\x1b[0m",
}

// codeTerminalLevels is what TerminalOptions.Level is for: the level is the
// caller's, so the line and the level cannot disagree. It is off by default,
// because a shell's own transcript has no severities.
func codeTerminalLevels(i int) code.LogLevel {
	switch {
	case strings.Contains(codeTerminalLines[i], "FAIL") ||
		strings.Contains(codeTerminalLines[i], "undefined"):
		return code.LogError
	case strings.HasPrefix(codeTerminalLines[i], "#"):
		return code.LogWarn
	}
	return code.LogTrace
}

// codeHistory is the commands, oldest first, with the one that failed still
// in the list: a history that dropped its failures is a history that cannot
// be searched for the reason the last build broke.
var codeHistory = []string{
	"go mod tidy",
	"gofmt -w ui/code",
	"go test ./ui/code/... -count=1",
	"git commit -am 'widen the gutter'",
	"go run ./cmd/gallery -shots /tmp/cx",
	"go build ./cmd/gallery",
}

// codeBuildLines is a build's output: seven lines with the five levels in them
// and one error, so that the head's badge and the error band's colour are both
// visible in the same frame.
var codeBuildLines = []code.LogLine{
	{Level: code.LogDebug, Time: "14:02:11.031", Source: "ui", Text: "settle: 3 builds in 41ms"},
	{Level: code.LogDebug, Time: "14:02:11.074", Source: "input", Text: "pad: box {248 2305 167 72}"},
	{Level: code.LogInfo, Time: "14:02:11.140", Source: "gallery", Text: "shots: 21 pages, scale 2"},
	{Level: code.LogWarn, Time: "14:02:12.006", Source: "gallery", Text: `page "input" taller than 3080`},
	{Level: code.LogWarn, Time: "14:02:12.311", Source: "gallery", Text: "check: 4 pages cut at the bottom"},
	{Level: code.LogError, Time: "14:02:13.882", Source: "chart", Text: "flow link 3 points at a missing node"},
	{Level: code.LogInfo, Time: "14:02:14.010", Source: "gallery", Text: "wrote code.png 1000x5320"},
}

// codeProblems is what the compiler and the linter said about the viewer, in
// the order they said it and not in the order of their severities: the order
// a compiler reports is the order it found them in.
var codeProblems = []code.Problem{
	{Message: "declared and not used: rows", File: "shared.go", Line: 91, Severity: core.Danger, Source: "compiler"},
	{Message: "unknown field Height in opts", File: "viewer.go", Line: 104, Column: 12, Severity: core.Danger, Source: "compiler"},
	{Message: "no Height: draws every line", File: "page.go", Line: 76, Severity: core.Warning, Source: "staticcheck"},
	{Message: `unused import "strings"`, File: "viewer.go", Line: 8, Severity: core.Warning, Source: "compiler"},
	{Message: "folds never read again", File: "shared.go", Line: 274, Severity: core.Neutral, Source: "staticcheck"},
	{Message: "carries a stale first", File: "viewer.go", Line: 120, Severity: core.Warning, Source: "code", Fixed: true},
}

// codeSearchResults is one search across the repository, with the line each
// hit was on, because a list of file names and counts is what a caller has
// and not what a person wants.
var codeSearchResults = []code.SearchResult{
	{Name: "opts.Height", Line: 104, Column: 16,
		Text: "if opts.Height <= 0 { // without it, every line"},
	{Name: "opts.Height", Line: 109, Column: 8,
		Text: `panic("code: CodeViewer needs a Height")`},
	{Name: "opts.Height", Line: 141, Column: 2,
		Text: "Height: opts.Height, // the viewport's own"},
	{Name: "shared.go", Line: 88, Column: 5,
		Text: "func codeRow(c *ui.Context, o CodeRowsOptions, n int) {"},
	{Name: "o.Lines", Line: 140, Column: 3,
		Text: "m := metricsOf(c, o.Lines) // how wide the gutter is"},
	{Name: "metricsOf", Line: 55, Column: 7,
		Text: "gutter := float32(digits) * m.size * 0.62"},
}

var codeBreakpoints = []code.Breakpoint{
	{File: "ui/code/viewer.go", Line: 104, Hit: 12, Enabled: true},
	{File: "ui/code/viewer.go", Line: 141, Condition: "len(runes) == 0", Hit: 3, Enabled: true},
	{File: "cmd/gallery/main.go", Line: 41, Condition: "i > 3", Enabled: false},
	{File: "ui/code/shared.go", Line: 88, Hit: 41, Enabled: true},
}

var codeFrames = []code.Frame{
	{Name: "CodeViewer", File: "ui/code/viewer.go", Line: 104, Where: "here"},
	{Name: "viewerTokens", File: "ui/code/viewer.go", Line: 124, Where: "1 frame up"},
	{Name: "Page.Draw", File: "ui/showcase/page.go", Line: 106, Where: "3 frames up"},
	{Name: "main", File: "cmd/gallery/main.go", Line: 96, Where: "7 frames up"},
}

var codeVariables = []code.Variable{
	{Name: "r", Value: "CodeViewerResult{Element: 0xc000…}"},
	{Name: "cache", Value: "map[int][]code.Token  len 3"},
	{Name: "rows", Value: "code.CodeRowsOptions{Lines: 36}"},
	{
		Name: "opts", Value: "CodeViewerOptions{…}", Expanded: true,
		Children: []code.Variable{
			{Name: "Name", Value: `string  "viewer.go"`},
			{Name: "Lang", Value: "code.Go"},
			{Name: "Height", Value: "float32  240"},
		},
	},
}

var codeSymbols = []code.Symbol{
	{Name: "CodeViewer", Kind: code.SymbolFunc, Line: 104,
		Detail: "func(c, opts) CodeViewerResult"},
	{
		Name: "CodeViewerOptions", Kind: code.SymbolType, Line: 13,
		Detail: "struct { … }",
		Children: []code.Symbol{
			{Name: "Source", Kind: code.SymbolValue, Line: 23, Detail: "string"},
			{Name: "Folds", Kind: code.SymbolValue, Line: 49, Detail: "[]Fold"},
			{Name: "Height", Kind: code.SymbolValue, Line: 53, Detail: "float32"},
		},
	},
	{Name: "viewerTokens", Kind: code.SymbolFunc, Line: 233,
		Detail: "func(c, name, lang, src) map[int][]Token"},
	{Name: "the token cache", Kind: code.SymbolSection, Line: 227,
		Detail: "// keyed on the text, not a pointer"},
}

var codeProcesses = []code.Process{
	{Name: "cmd/gallery", Args: []string{"-shots", "/tmp/cx"}, Pid: "48213",
		CPU: "94%", Memory: "312 MB", Running: true},
	{Name: "go", Args: []string{"test", "./ui/code/..."}, Pid: "48241",
		CPU: "61%", Memory: "1.1 GB", Running: true},
	{Name: "gopls", Args: []string{"check", "viewer.go"}, Pid: "11907",
		CPU: "3%", Memory: "284 MB", Running: false},
	{Name: "callbacks-native", Pid: "40118", CPU: "0%", Memory: "96 MB", Running: true},
}

var codeCompletions = []code.Completion{
	{Label: "CodeMetrics", Detail: "func(c, 36) (f32, f32)", Kind: code.SymbolFunc, Selected: true},
	{Label: "CodeMono", Detail: "func(c, s, size)", Kind: code.SymbolFunc},
	{Label: "CodePanel", Detail: "func(c, body, opts, head)", Kind: code.SymbolFunc},
	{Label: "CodeRowsOptions", Detail: "type · 6 fields", Kind: code.SymbolType},
	{Label: "CodeTokens", Detail: "private to the package", Kind: code.SymbolValue},
	{Label: "LineHeight", Detail: "func(c) f32", Kind: code.SymbolFunc},
}

// codeFlame is a profile of the gallery run: the whole call at the top, and
// underneath it the work. The widths are the caller's own numbers, because a
// flame graph's whole meaning is in its proportions and a panel that guessed
// the unit would draw a chart of the wrong thing beautifully.
var codeFlame = code.FlameFrame{
	Name: "main", Self: 12, Total: 840,
	File: "cmd/gallery/main.go", Line: 96,
	Children: []code.FlameFrame{
		{
			Name: "fit", Self: 640, Total: 828, File: "cmd/gallery/main.go", Line: 162,
			Children: []code.FlameFrame{
				{Name: "render", Self: 601, Total: 601, File: "cmd/gallery/main.go", Line: 176},
				{Name: "BottomFlush", Self: 27, Total: 27, File: "cmd/gallery/main.go", Line: 201},
			},
		},
	},
}

// codeFrameBytes is the first twenty-four bytes of a 640 × 480 RGBA PNG: the
// signature, the chunk's length, the chunk's name, and the two dimensions. A
// reader can check every byte of it against the format, which is the point of
// putting a hex view next to a viewer of a text file.
var codeFrameBytes = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
	0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
	0x00, 0x00, 0x02, 0x80, 0x00, 0x00, 0x01, 0xe0,
}

// ── the file ────────────────────────────────────────────────────────────────

// viewerSection is a file open in a window: the text, and the map of it
// down the side. The two are handed the same string, which is the only way
// they cannot disagree about which line the caret is on.
func viewerSection(c *ui.Context) {
	showcase.Section(c, "文件 · CodeViewer / Minimap")

	showcase.Field(c, "CodeViewer — 行号、当前行、查找标记、折叠箭头都出自同一套 codeRow")
	// Two thirds and one third: the viewer's own width is decided by the
	// lines it has to show, and a map that is half the page is a second
	// window rather than a position in one.
	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 4)).AlignItems(ui.Start).
		Children(func() {
			ui.Column(c).WidthPercent(share(2)).Shrink(0).Children(func() {
				code.CodeViewer(c, code.CodeViewerOptions{
					Name: "ui/code/viewer.go", Source: codeSource, Lang: code.Go,
					Height: codeRows(c, 12), Width: unit(c, 116),
					Current: 22, Query: "lines", Find: true,
					Folds: codeFolds, Caption: "main · 1 fold shut",
				})
			})
			ui.Column(c).Width(unit(c, 112)).Shrink(0).Gap(unit(c, 1.5)).
				Children(func() {
					showcase.Field(c, "Minimap — 同一份 Source 的侧面")
					code.Minimap(c, code.MinimapOptions{
						Name: "ui/code/viewer.go", Source: codeSource, Lang: code.Go,
						Width: unit(c, 6), LineHeight: unit(c, 1),
						At: 1, Viewport: 12,
					})
				})
		})
	codeNote(c, []string{
		"Current 是行号不是索引：调用方手里那个位置直接就能交过来。",
		"Find 为 false 时一个匹配也不算——三千行的文件不该为一个空查询算匹配。",
		"Folds 是调用方的切片：哪些行能折是语言括号的问题，扫描器不该替它回答。",
		"地图不知道折叠——它只读 Source，所以折起来的那一段在地图上还是十一道色。",
	})
}

// chromeSection is the furniture every panel in this package wears: the
// panel's own face, its head, the pill in the head, and the two numbers a
// caller needs before it can lay anything out.
func chromeSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "面板的头 · CodePanel / CodeHeader / CodeBadge / CodeCaption")

	showcase.Field(c, "CodeHeader — 一行：标题让位，计数不让位")
	ui.Box(c).Width(unit(c, 236)).Shrink(0).Radius(theme.SmallRadius).
		Background(k.Background).Border(theme.BorderWidth, k.Border).Clip().
		Children(func() {
			code.CodeHeader(c, "ui/code/shared.go · 4e5f6a7",
				[]string{fmt.Sprintf("%d lines", codeLines)}, func() {
					code.CodeBadge(c, "modified", core.Warning)
					code.CodeSelectMark(c, true)
				})
		})

	showcase.Field(c, "CodeBadge — 严重度是 core.Severity 给的，组件自己不挑颜色")
	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 1.5)).Children(func() {
		for _, sev := range []core.Severity{
			core.Neutral, core.Accent, core.Success, core.Warning, core.Danger,
		} {
			code.CodeBadge(c, sev.String(), sev)
		}
		code.CodeBadge(c, "2 to replace", core.Warning)
		code.CodeBadge(c, "running", core.Accent)
	})

	showcase.Field(c, "CodeCaption / CodeMono / CodeSelectMark / CodeMetrics — 面板周围的字")
	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 4)).AlignItems(ui.Start).
		Children(func() {
			ui.Column(c).Width(unit(c, 116)).Shrink(0).Gap(unit(c, 1)).Children(func() {
				ui.Row(c).Gap(unit(c, 1.5)).AlignItems(ui.Center).Children(func() {
					code.CodeCaption(c, "ui/code/viewer.go:104")
					code.CodeSelectMark(c, false)
					code.CodeSelectMark(c, true)
				})
				ui.Row(c).Gap(unit(c, 1.5)).AlignItems(ui.Center).Children(func() {
					code.CodeMono(c, "github.com/HycJack/MintUI/ui/code", theme.RowSize)
					code.CodeCaption(c, "路径也是代码")
				})
				codeCaption2(c, "空字符串画一条高度为 0 的线，而不是一个空盒子")
			})
			ui.Column(c).Width(unit(c, 116)).Shrink(0).Gap(unit(c, 1)).Children(func() {
				line, gutter := code.CodeMetrics(c, codeLines)
				codeCaption2(c, fmt.Sprintf("CodeMetrics(c, %d) = %g, %g", codeLines, line, gutter))
				codeCaption2(c, fmt.Sprintf("LineHeight(c) = %g", code.LineHeight(c)))
				codeCaption2(c, "行高决定能画几行；行号槽的宽度按文件的行数算")
				codeCaption2(c, "所以一个一千行的文件和一个十行的不会共用一个左边缘")
			})
		})
}

// langSection is the highlighter's whole range: six word lists and the one
// that is not a language at all. Every specimen is the same component with a
// different Lang, which is the claim this section exists to make visible.
func langSection(c *ui.Context) {
	showcase.Section(c, "六门语言，加一种不是语言的 · 同一个组件，Lang 换一下")

	langs := []struct {
		lang code.Lang
		file string
		src  string
	}{
		// Lines are kept to about thirty-four characters because that is what
		// fits in seventy-seven units of a monospaced face at the row size:
		// a specimen whose longest line is clipped shows a scrollbar instead
		// of showing the language.
		{code.Go, "sample.go", `func Toggle(w *os.File, on bool) {
	if on {
		fmt.Fprintln(w, "on")
	} else {
		fmt.Fprintln(w, "off")
	}
}`},
		{code.JavaScript, "sample.ts", `const r = await fetch("/v1");
if (!r.ok) throw Error(r.status);
export async function list() {
  return r.json();
}`},
		{code.Python, "app.py", `def resolve(cb: dict) -> str:
    # An AC repair costs 184.50.
    if cb["priority"] == "high":
        return "page someone"
    return "due " + cb["due"]`},
		{code.Rust, "lib.rs", `fn median(v: &mut [f64]) -> f64 {
    if v.is_empty() { return 0.0; }
    v.sort_by(f64::total_cmp);
    v[v.len() / 2]
}`},
		{code.SQL, "queries.sql", `SELECT customer, count(*)
  FROM callbacks
 WHERE priority = 'high'
 GROUP BY customer
 ORDER BY 2 DESC;`},
		{code.Shell, "build.sh", `go test ./ui/code/... -count=1
grep -rn "Height" ui/code
# the gutter is the file's`},
	}
	// Three to a row at seventy-seven units each: 3 × 77 + 2 × 2 = 235 of the
	// 236 the page has between its margins, so the third specimen ends inside
	// the page rather than past it.
	ui.Column(c).Width(unit(c, 236)).Gap(unit(c, 3)).Children(func() {
		for start := 0; start < len(langs); start += 3 {
			row := langs[start : start+3]
			ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 2)).AlignItems(ui.Start).
				Children(func() {
					for _, s := range row {
						one := s
						ui.Column(c).Width(unit(c, 77)).Shrink(0).Gap(unit(c, 0.75)).
							Children(func() {
								ui.Row(c).Gap(unit(c, 1)).AlignItems(ui.Center).
									Children(func() {
										ui.Text(c, one.lang.String()).TextColor(core.Tokens(c).Text).
											FontSize(core.FontSize(c, theme.RowSize)).Bold()
										code.CodeCaption(c, one.file)
									})
								// Gutter off in the specimens: six files of
								// four lines each are about the words, and a
								// line-number column in every one of them
								// would spend a third of each on digits.
								code.CodeViewer(c, code.CodeViewerOptions{
									Name: one.file, Source: one.src, Lang: one.lang,
									Height: codeRows(c, 7), Gutter: codeNo(),
								})
							})
					}
				})
		}
	})
	// The seventh one is not a language, and it is here because a page that
	// only showed six would leave it as a word in a table: LangPlain is what
	// a viewer draws for a file it has never heard of, and it still knows a
	// string from a number.
	ui.Column(c).Width(unit(c, 236)).Gap(unit(c, 0.75)).Children(func() {
		ui.Row(c).Gap(unit(c, 1)).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "LangPlain").TextColor(core.Tokens(c).Text).
				FontSize(core.FontSize(c, theme.RowSize)).Bold()
			code.CodeCaption(c, "notes.txt · 一个没有语言的文件")
		})
		code.CodeViewer(c, code.CodeViewerOptions{
			Name: "notes.txt", Source: plainSource, Lang: code.LangPlain,
			Height: codeRows(c, 4), Gutter: codeNo(),
		})
	})
	codeNote(c, []string{
		"LangPlain 不是一种语言，是没有语言：仍然认字符串和数字——配置文件和 shell 的记录都要这个。",
		"高亮是扫描器不是解析器：它不建树、不认识函数，半写的一行也照画不误。",
	})
}

// codeDiffSection is the difference between two of the files above, in the two
// shapes a window can show it. The unified one gets half the page and the
// split one gets all of it, because two columns of the same lines need twice
// the width to say the same thing.
func codeDiffSection(c *ui.Context) {
	showcase.Section(c, "差异 · Diff / DiffLine")

	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 4)).AlignItems(ui.Start).
		Children(func() {
			ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 1.5)).
				Children(func() {
					showcase.Field(c, "Unified — 一列，两道行号槽")
					code.Diff(c, code.DiffOptions{
						Name: "ui/code/board.go", Old: diffOld, New: diffNew,
						Lang: code.Go, Context: 1, Height: unit(c, 46),
						Caption: "3d4e5f6 · 4e5f6a7",
					})
				})
			ui.Column(c).Width(unit(c, 112)).Shrink(0).Gap(unit(c, 1.5)).
				Children(func() {
					showcase.Field(c, "两种形状之外没有第三种答案")
					codeCaption2(c, "差异是 DiffLines(old, new, context) 算出来的，"+
						"画的时候不再问一遍。")
					codeCaption2(c, "两侧的行号都是数出来的；缺的一侧画空档，"+
						"不是画一个 0——没有文件有第 0 行。")
					codeCaption2(c, "改动的整行染 SuccessBg / DangerBg，"+
						"行首的 + 与 − 才是那句话本身。")
				})
		})

	showcase.Field(c, "Split — 两列，一行的高度放得下两边")
	code.Diff(c, code.DiffOptions{
		Name: "ui/code/board.go", Old: diffOld, New: diffNew,
		Lang: code.Go, Context: 1, Height: unit(c, 40), Split: true,
	})
}

// terminalSection is the terminal a person runs things in, and the two things
// that hang off it: a search across what it printed and the list of what was
// typed into it.
func terminalSection(c *ui.Context) {
	showcase.Section(c, "终端 · Terminal / TerminalTabs / TerminalSearch / CommandHistory")

	showcase.Field(c, "TerminalTabs — 选中的那个是抬起来的，不是一个药丸")
	code.TerminalTabs(c, code.TerminalTabsOptions{
		Label: "Terminals", Tabs: []string{"zsh", "build", "server"},
		Current: 1, Closeable: true,
	})

	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 4)).AlignItems(ui.Start).
		Children(func() {
			ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 1.5)).
				Children(func() {
					showcase.Field(c, "Terminal — 尾部是提示符，Level 是调用方给的一行一个")
					code.Terminal(c, code.TerminalOptions{
						Name: "zsh", Lines: codeTerminalLines,
						Height: codeRows(c, 10), Width: unit(c, 116),
						Prompt: "go run ./cmd/gallery -check", PromptName: "zsh",
						Follow: true, Level: codeTerminalLevels,
					})
				})
			ui.Column(c).Width(unit(c, 112)).Shrink(0).Gap(unit(c, 1.5)).
				Children(func() {
					showcase.Field(c, "TerminalSearch — 终端没有列表可给，只有两个数")
					code.TerminalSearch(c, code.TerminalSearchOptions{
						Label: "Find in terminal", Query: "gallery",
						Found: 12, At: 0,
					})
					showcase.Field(c, "CommandHistory — 最新的在下面")
					code.CommandHistory(c, code.CommandHistoryOptions{
						Name: "History", Commands: codeHistory,
						Height: unit(c, 34), Width: unit(c, 112), At: 4,
					})
				})
		})
	codeNote(c, []string{
		"Follow 是终端的，LogViewer 故意没有：一个终端要的是别滚走，一个日志要的是别重排。",
		"命令历史最新的在最后，和 shell 自己相反——面板是从上往下读的。",
	})
}

// buildSection is the work: what a build said, what the compiler made of it,
// and the log the two of them are read out of.
func buildSection(c *ui.Context) {
	showcase.Section(c, "构建 · OutputPanel / ProblemsPanel / LogViewer")

	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 4)).AlignItems(ui.Start).
		Children(func() {
			ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 1.5)).
				Children(func() {
					showcase.Field(c, "OutputPanel — 头部说还在跑，坏行染 DangerBg")
					// Seven whole rows, and no level column: the level is
					// already the row's colour and its band, and a column of
					// the word beside it says it twice.
					code.OutputPanel(c, code.OutputOptions{
						Name: "build", Lines: codeBuildLines,
						Height: codeRows(c, 7), Width: unit(c, 116),
						Running: true, Detail: codeNo(),
					})
				})
			ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 1.5)).
				Children(func() {
					showcase.Field(c, "ProblemsPanel — 按说的顺序，不按严重度排序")
					code.ProblemsPanel(c, code.ProblemsOptions{
						Name: "go vet ./ui/code/...", Problems: codeProblems,
						Height: unit(c, 40), Width: unit(c, 116), Selected: 1,
					})
				})
		})

	showcase.Field(c, "LogViewer — 左边的时间是调用方的格式，右边的来源只在要的时候画")
	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 4)).AlignItems(ui.Start).
		Children(func() {
			ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 1.5)).
				Children(func() {
					showcase.Field(c, "不过滤")
					code.LogViewer(c, code.LogOptions{
						Name: "build", Lines: codeBuildLines,
						Height: codeRows(c, 7), Width: unit(c, 116),
						ShowSource: codeYes(),
					})
				})
			ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 1.5)).
				Children(func() {
					showcase.Field(c, "Filter: Warn — 只剩三行，头上的计数也跟着变")
					code.LogViewer(c, code.LogOptions{
						Name: "build · warnings and worse", Lines: codeBuildLines,
						Height: codeRows(c, 7), Width: unit(c, 116),
						Filter: code.LogWarn,
					})
				})
		})
	codeNote(c, []string{
		"OutputPanel 就是 LogViewer 加一个头：日志是读的东西，输出是看的东西。",
		"LogViewer 不重排。从上往下读就是事情发生的顺序，谁把最新的放到第一行是替调用方做了决定。",
	})
}

// findSection is finding something: in this file, and in the repository.
func findSection(c *ui.Context) {
	showcase.Section(c, "查找 · FindWidget / SearchPanel / CodeMatchTokens")

	showcase.Field(c, "FindWidget — 计数在条子里，因为「4/212」才是答案，框里的字只是一半")
	// Held across frames: the bar types into both of them and the close button
	// empties the query, so a pair of this frame's strings would take back
	// every edit on the frame after it was made.
	findQuery := showcase.State(c, "code.find.query", "Height")
	findReplace := showcase.State(c, "code.find.replace", "opts.Height")
	code.FindWidget(c, code.FindOptions{
		Label: "Find in file", Query: findQuery, Found: 12, At: 3,
		Replace: findReplace, CanReplace: true, Closeable: true,
		Regex: true,
	})

	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 4)).AlignItems(ui.Start).
		Children(func() {
			ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 1.5)).
				Children(func() {
					showcase.Field(c, "SearchPanel — 每一条都带着它命中的那一行")
					code.SearchPanel(c, code.SearchPanelOptions{
						Name: "Search results", Query: "opts.Height",
						Results: codeSearchResults, Replaces: 3,
						Height: codeRows(c, 11), Width: unit(c, 116), Selected: 1,
					})
				})
			ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 1.5)).
				Children(func() {
					showcase.Field(c, "CodeMatchTokens — 行外的命中长一个样子")
					codeMatchStrip(c)
					codeCaption2(c, "标记是按字节偏移把 token 切开的，所以一个命中跨过 "+
						"关键字和字符串时两边都亮；按像素画在行上面做不到这件事。")
				})
		})
}

// codeMatchStrip is three rows of highlighted code with one query's matches
// marked behind them, drawn by CodeMatchTokens rather than by the viewer's own
// rows — so a match outside a row looks the same as one inside it.
func codeMatchStrip(c *ui.Context) {
	k := core.Tokens(c)
	ui.Box(c).Width(unit(c, 116)).Radius(theme.SmallRadius).
		Background(k.Background).Border(theme.BorderWidth, k.Border).Clip().
		Padding(unit(c, 1)).Children(func() {
		ui.Column(c).FillWidth().Gap(unit(c, 0.5)).Children(func() {
			code.CodeCaption(c, "FindMatches(\"…\", \"Height\")")
			for _, line := range []string{
				"if opts.Height <= 0 {",
				"\tpanic(\"needs a Height\")",
				"\tHeight: opts.Height,",
			} {
				text := line
				ui.Row(c).FillWidth().Height(20).AlignItems(ui.Center).
					Children(func() {
						code.CodeMatchTokens(c,
							code.Highlight(text, code.Go, false),
							code.FindMatches(text, "Height"),
							theme.RowSize)
					})
			}
		})
	})
}

// debugSection is the debugger's four lists: where it is stopped, what
// called it, what is in scope, and what it will stop at next.
func debugSection(c *ui.Context) {
	showcase.Section(c, "调试 · BreakpointList / CallStack / VariablesPanel / SymbolOutline")

	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 4)).AlignItems(ui.Start).
		Children(func() {
			ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 1.5)).
				Children(func() {
					showcase.Field(c, "CallStack — 最内层在上面，就是人读程序的反过来")
					code.CallStack(c, code.CallStackOptions{
						Name: "Call stack", Frames: codeFrames,
						Height: unit(c, 30), Width: unit(c, 116), Selected: 0,
					})
				})
			ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 1.5)).
				Children(func() {
					showcase.Field(c, "VariablesPanel — 值是调用方自己的格式")
					code.VariablesPanel(c, code.VariablesOptions{
						Name: "Variables", Variables: codeVariables,
						Height: unit(c, 30), Width: unit(c, 116),
					})
				})
		})

	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 4)).AlignItems(ui.Start).
		Children(func() {
			ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 1.5)).
				Children(func() {
					showcase.Field(c, "BreakpointList — 关掉的画删除线，因为还能再打开")
					code.BreakpointList(c, code.BreakpointOptions{
						Name: "Breakpoints", Points: codeBreakpoints,
						Height: unit(c, 30), Width: unit(c, 116), Selected: 0,
					})
				})
			ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 1.5)).
				Children(func() {
					showcase.Field(c, "SymbolOutline — 内容是调用方的树，本包不解析语法")
					code.SymbolOutline(c, code.SymbolOutlineOptions{
						Name: "Outline", Symbols: codeSymbols,
						Height: unit(c, 30), Width: unit(c, 116), Selected: 1,
					})
				})
		})
	codeNote(c, []string{
		"四个列表的行高一样，因为它们都走 codeRow：一个调试面板里四种行高是一种噪音。",
		"哪一个被选中是调用方的一个指针，组件自己不存状态。",
	})
}

// runSection is the toolbar that drives it, the processes that are running
// under it, and the menu that appears when something is typed.
func runSection(c *ui.Context) {
	showcase.Section(c, "运行 · DebugToolbar / ProcessList / CompletionMenu")

	showcase.Field(c, "DebugToolbar — 状态贴在行尾，不夹在两个按钮中间")
	code.DebugToolbar(c, code.DebugToolbarOptions{
		Name: "Debug controls",
		Actions: []code.DebuggerAction{
			{Name: "Continue", Primary: true},
			{Name: "Step over"},
			{Name: "Step into"},
			{Name: "Step out"},
			{Name: "Stop", Disabled: true},
		},
		Status: "Stopped at line 104",
	})

	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 4)).AlignItems(ui.Start).
		Children(func() {
			ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 1.5)).
				Children(func() {
					showcase.Field(c, "ProcessList — CPU 与内存是调用方给的字符串")
					code.ProcessList(c, code.ProcessOptions{
						Name: "Processes", Processes: codeProcesses,
						Height: unit(c, 34), Width: unit(c, 116), Selected: 0,
					})
				})
			ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 1.5)).
				Children(func() {
					showcase.Field(c, "CompletionMenu — 菜单归它要补的那个字段，不归窗口")
					code.CompletionMenu(c, code.CompletionMenuOptions{
						Label: "Completions", Completions: codeCompletions,
						Width: unit(c, 116), Prefix: "code.", Highlight: 0,
					})
					codeCaption2(c, "Prefix 画在菜单头上，因为调用方自己的框可能已经"+
						"把它画在光标左边了——两处都画是同一句话说两遍。")
				})
		})
}

// profileSection is the two views that are not lines of source: the bytes of
// a file, and where a program's time went.
func profileSection(c *ui.Context) {
	showcase.Section(c, "字节与剖析 · HexViewer / Flamegraph")

	showcase.Field(c, "HexViewer — 偏移、两组字节、可打印的字符，三处对齐")
	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 4)).AlignItems(ui.Start).
		Children(func() {
			ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 1.5)).
				Children(func() {
					code.HexViewer(c, code.HexOptions{
						Name: "ui/code/frame.png", Data: codeFrameBytes,
						Height: codeRows(c, 2), Address: 0, Selected: 0,
					})
				})
			ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 1.5)).
				Children(func() {
					showcase.Field(c, "Flamegraph — 宽度是调用方的比例，所以宽度是必须的")
					code.Flamegraph(c, code.FlamegraphOptions{
						Name: "Flamegraph — callbacks", Frame: codeFlame,
						Width: unit(c, 116), Height: unit(c, 5),
						Highlight: "BottomFlush",
					})
				})
		})
	codeNote(c, []string{
		"十六进制的行和源码的行是同一套 codeRow：偏移也是右对齐的，不然一列偏移没有列的样子。",
		"火焰图是一张画，不是每个 span 一个盒子——一个服务进程的 profile 有几万个 span。",
	})
}

// ── the pure half ───────────────────────────────────────────────────────────

// codePureSection is everything in this package that decides rather than
// draws. It is last because it is the reason the rest of the page is the way
// it is: a token's colour, a line's numbers, a diff's hunks and a log's
// filter are all answers to a question about data, asked before an element is
// made.
func codePureSection(c *ui.Context) {
	showcase.Section(c, "纯函数 — 先算出答案，再画")

	showcase.Field(c, "Highlight — 一行切成若干 run，每个 run 一个 TokenKind")
	cols := []float32{96, 20, 20, 22, 30}
	codeHeadRow(c, cols, "text", "kind", "kind.String()", "ink", "")
	for _, t := range code.Highlight(
		`panic("code: CodeViewer needs a Height")`, code.Go, false) {
		codePureRow(c, cols, t.Text, codeItoa(int(t.Kind)), t.Kind.String(),
			inkName(c, code.TokenInk(c, t.Kind)), "")
	}

	showcase.Field(c, "HighlightSpans — 只有种类没有文字；HighlightAll — 一次给一整个文件")
	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 4)).AlignItems(ui.Start).
		Children(func() {
			ui.Column(c).Width(unit(c, 116)).Shrink(0).Gap(unit(c, 1)).Children(func() {
				codeCaption2(c, `HighlightSpans("x = 1 // n", Go, false)`)
				for i, kind := range code.HighlightSpans("x = 1 // n", code.Go, false) {
					chipRow(c, kind.String(), code.TokenInk(c, kind), codeItoa(i))
				}
			})
			ui.Column(c).Width(unit(c, 116)).Shrink(0).Gap(unit(c, 1)).Children(func() {
				codeCaption2(c, `HighlightAll([]string{"package code", "// note", "var n = 36"}, Go)`)
				for i, line := range code.HighlightAll([]string{
					"package code", "// note", "var n = 36",
				}, code.Go) {
					codeCaption2(c, fmt.Sprintf("line %d → %s", i+1, kindsOf(line)))
				}
			})
		})

	showcase.Field(c, "FindMatches / CodeMatchTokens — 偏移是字节，标记按偏移切开")
	// Four columns, because the values are four: the needle, how many places
	// it was found, where the first one starts and ends, and the words that
	// live there. Five column widths for four values is how the last one ends
	// up in a ten-unit column and every needle prints as "He…".
	matchCols := []float32{24, 16, 18, 92}
	matchLine := "if opts.Height <= 0 { runes }"
	codeHeadRow(c, matchCols, "needle", "matches", "from..to", "text")
	for _, q := range []string{"Height", "n", "rune"} {
		spans := code.FindMatches(matchLine, q)
		if len(spans) == 0 {
			codePureRow(c, matchCols, q, "0", "—", "（这一行没有）")
			continue
		}
		for i, s := range spans {
			codePureRow(c, matchCols, q, codeItoa(i),
				fmt.Sprintf("%d..%d", s.From, s.To),
				fmt.Sprintf("%q", matchLine[s.From:s.To]))
		}
	}

	showcase.Field(c, "DiffLines — 一段文本变成若干行，每行三件事")
	diffCols := []float32{18, 12, 12, 140}
	codeHeadRow(c, diffCols, "kind", "old", "new", "text")
	for _, l := range code.DiffLines(diffOld, diffNew, 1) {
		codePureRow(c, diffCols, l.Kind.String(),
			codeOrDash(l.Old), codeOrDash(l.New), l.Text)
	}

	showcase.Field(c, "ParseAnsi — 转义序列拿掉，剩下的按程序要的画")
	ansi := "\x1b[1;31mfatal:\x1b[0m not a repository \x1b[2m(quiet)\x1b[0m"
	ansiCols := []float32{78, 30, 16, 16, 26}
	codeHeadRow(c, ansiCols, "text", "colour", "bold", "dim", "background")
	for _, s := range code.ParseAnsi(ansi) {
		codePureRow(c, ansiCols, fmt.Sprintf("%q", s.Text),
			ansiName(s.Colour), codeYesNo(s.Bold), codeYesNo(s.Dim),
			ansiName(s.Background))
	}

	showcase.Field(c, "TokenInk / TokenKind — 种类到墨色的一处映射，图例和它读的是同一个函数")
	kinds := []code.TokenKind{
		code.TokenPlain, code.Keyword, code.Type, code.String,
		code.Number, code.Comment, code.Function, code.Punct,
	}
	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 1)).Children(func() {
		for _, kind := range kinds {
			one := kind
			ui.Column(c).Width(unit(c, 28)).Shrink(0).Gap(unit(c, 0.5)).
				AlignItems(ui.Center).Children(func() {
				ui.Box(c).Width(unit(c, 28)).Height(unit(c, 4)).
					Radius(theme.PillRadius).Background(code.TokenInk(c, one))
				ui.Text(c, one.String()).TextColor(core.Tokens(c).TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize)).MaxLines(1)
			})
		}
	})

	showcase.Field(c, "SymbolKind / LogLevel / DiffLineKind — 三组枚举，都带一个 String 和一个严重度")
	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 4)).AlignItems(ui.Start).
		Children(func() {
			ui.Column(c).Width(unit(c, 76)).Shrink(0).Gap(unit(c, 1)).Children(func() {
				codeCaption2(c, "SymbolKind — 大纲里每个符号的缩写")
				for _, s := range []code.SymbolKind{
					code.SymbolFunc, code.SymbolType,
					code.SymbolValue, code.SymbolSection,
				} {
					codePureRow(c, []float32{20, 30}, s.String(), codeItoa(int(s)))
				}
			})
			ui.Column(c).Width(unit(c, 76)).Shrink(0).Gap(unit(c, 1)).Children(func() {
				codeCaption2(c, "LogLevel — 整词与 core.Severity")
				for _, l := range []code.LogLevel{
					code.LogTrace, code.LogDebug, code.LogInfo,
					code.LogWarn, code.LogError,
				} {
					levelPair(c, l)
				}
			})
			ui.Column(c).Width(unit(c, 76)).Shrink(0).Gap(unit(c, 1)).Children(func() {
				codeCaption2(c, "DiffLineKind — 差异里三种行")
				for _, d := range []code.DiffLineKind{
					code.DiffSame, code.DiffAdded, code.DiffRemoved,
				} {
					codePureRow(c, []float32{30, 24}, d.String(), codeItoa(int(d)))
				}
			})
		})

	showcase.Field(c, "CodeRowsOptions — 每个画行的组件共用这一份参数")
	rowsCols := []float32{34, 34, 122}
	codeHeadRow(c, rowsCols, "field", "value", "what it changes")
	for _, row := range []struct{ field, value, what string }{
		{"Lines", codeItoa(codeLines), "行号槽的宽度就是按它算的"},
		{"First", "12", "0 表示一行行号都不画：右侧的一列要的就是这个"},
		{"Current", "104", "当前行；CurrentBand 关时它什么也不画"},
		{"Selection", "[12 13]", "选中行；SelectionBand 关时它什么也不画"},
		{"Gutter", "true", "行号本体"},
		{"CurrentBand", "true", "当前行底下的一道"},
		{"SelectionBand", "true", "选中行底下的一道"},
		{"BandColour", "{0 0 0 0}", "覆盖当前行那道：零透明就用自己的"},
	} {
		codePureRow(c, rowsCols, row.field, row.value, row.what)
	}
	codeCaption2(c, "viewer、diff、hex、日志画行时读的是同一个函数，所以四者不可能"+
		"差一个像素。")

	ui.Column(c).Width(unit(c, 236)).Gap(unit(c, 0.75)).Children(func() {
		for _, line := range []string{
			"本包的组件不读磁盘、不跑命令：源码、日志、差异、字节、profile 全是参数。",
			"颜色不是本包挑的：token 的墨色由 TokenInk 给，严重度由 core.Severity 给。",
		} {
			ui.Text(c, line).TextColor(core.Tokens(c).TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize))
		}
	})
}

// levelPair is one LogLevel drawn as the word it prints and the pair it is
// painted with, side by side: the two are one decision and the table is about
// the decision.
func levelPair(c *ui.Context, l code.LogLevel) {
	sev := l.Severity()
	ui.Row(c).Gap(unit(c, 1)).AlignItems(ui.Center).Children(func() {
		ui.Text(c, l.String()).TextColor(core.Tokens(c).TextMuted).Width(unit(c, 18)).
			FontSize(core.FontSize(c, theme.CaptionSize))
		code.CodeBadge(c, sev.String(), sev)
	})
}

// chipRow is one kind of token as a chip of its own colour, so that the table
// above it and the colour below it are visibly the same answer.
func chipRow(c *ui.Context, name string, ink ui.Color, at string) {
	ui.Row(c).Gap(unit(c, 1)).AlignItems(ui.Center).Children(func() {
		ui.Box(c).Width(unit(c, 4)).Height(unit(c, 3)).Radius(theme.PillRadius).
			Background(ink).Shrink(0)
		ui.Text(c, name).TextColor(core.Tokens(c).TextMuted).
			FontSize(core.FontSize(c, theme.CaptionSize)).Width(unit(c, 16))
		ui.Text(c, at).TextColor(core.Tokens(c).TextFaint).
			FontSize(core.FontSize(c, theme.CaptionSize))
	})
}

// codeNote is the two or three lines of prose a section ends with. It is a
// function because six sections want one and a gallery that spelled it out
// six times would spell it out seven the next time.
func codeNote(c *ui.Context, lines []string) {
	ui.Column(c).Width(unit(c, 236)).Gap(unit(c, 0.5)).MarginY(unit(c, 1)).
		Children(func() {
			for _, line := range lines {
				ui.Text(c, line).TextColor(core.Tokens(c).TextFaint).
					FontSize(core.FontSize(c, theme.CaptionSize))
			}
		})
}

// ── the little helpers ──────────────────────────────────────────────────────

// codeHeadRow is a pure function table's column names. It is its own function
// so that every table on this page starts the same way, which is what makes
// two tables comparable at a glance.
func codeHeadRow(c *ui.Context, cols []float32, names ...string) {
	ui.Row(c).Gap(unit(c, 1.5)).MarginY(unit(c, 0.5)).Children(func() {
		for i, name := range names {
			if i >= len(cols) {
				break
			}
			ui.Text(c, name).TextColor(core.Tokens(c).TextMuted).Width(unit(c, cols[i])).
				FontSize(core.FontSize(c, theme.CaptionSize)).Bold().MaxLines(1)
		}
	})
}

// codePureRow is one line of a pure function's output: each value in a column
// of stated width. Nothing grows, because a row that fills its column and then
// grows its last one pushes that one off the page — and the value it clips is
// the value somebody came for.
func codePureRow(c *ui.Context, cols []float32, cells ...string) {
	ui.Row(c).Gap(unit(c, 1.5)).Children(func() {
		for i, cell := range cells {
			if i >= len(cols) {
				break
			}
			ui.Text(c, cell).TextColor(core.Tokens(c).TextMuted).Width(unit(c, cols[i])).
				FontSize(core.FontSize(c, theme.CaptionSize)).MaxLines(1)
		}
	})
}

// codeCaption2 is a line of prose at caption size. It is here rather than
// spelled out because ui/code's own CodeCaption is a single line with no
// wrapping, which is right for a panel's head and wrong for a sentence.
func codeCaption2(c *ui.Context, text string) *ui.Element {
	return ui.Text(c, text).TextColor(core.Tokens(c).TextMuted).
		FontSize(core.FontSize(c, theme.CaptionSize))
}

// codeRows is n whole rows of code, which is what every viewport here is given
// rather than a round number of units: a viewport whose height is not a whole
// number of rows ends on half a line, and half a line of code reads as a
// rendering bug rather than as "there is more below".
func codeRows(c *ui.Context, n float32) float32 { return code.LineHeight(c) * n }

// codeYes and codeNo are the *bool three options in this package take, so that
// a caller can say "off" as plainly as it says "on". They are helpers rather
// than a bool field because a bool field would make a false and an unset the
// same thing, which is exactly the distinction the pointers exist to keep.
func codeYes() *bool { t := true; return &t }

func codeNo() *bool { f := false; return &f }

func codeItoa(n int) string { return fmt.Sprintf("%d", n) }

func codeYesNo(b bool) string {
	if b {
		return "是"
	}
	return "否"
}

func codeOrDash(n int) string {
	if n == 0 {
		return "—"
	}
	return codeItoa(n)
}

// kindsOf is a row of tokens as their kinds, which is what HighlightAll hands
// back per line and what a table of it wants to say.
func kindsOf(tokens []code.Token) string {
	parts := make([]string, 0, len(tokens))
	for _, t := range tokens {
		if t.Kind != code.TokenPlain {
			parts = append(parts, t.Kind.String())
		}
	}
	if len(parts) == 0 {
		return "（全是代码）"
	}
	return strings.Join(parts, " · ")
}

// inkName is a colour as the token it came from, because a table of a mapping
// is about the mapping and the mapping has names in it.
func inkName(c *ui.Context, col ui.Color) string {
	tk := core.Tokens(c)
	for _, pair := range []struct {
		name string
		col  ui.Color
	}{
		{"Text", tk.Text}, {"TextMuted", tk.TextMuted}, {"TextFaint", tk.TextFaint},
		{"Accent", tk.Accent}, {"AccentText", tk.AccentText}, {"Success", tk.Success},
		{"Warning", tk.Warning}, {"Danger", tk.Danger}, {"Lively", tk.Lively},
		// A function's ink is Lively mixed towards the window's own text, so
		// the table can say which two tokens it is a mix of — which is the
		// honest answer, and more useful than "some colour".
		{"Lively 混 Text", tk.Lively.Mix(tk.Text, 0.35)},
	} {
		if pair.col == col {
			return pair.name
		}
	}
	return "混出来的颜色"
}

// ansiName is an AnsiColour as the name the constant has in ui/code, because
// the type has no String() of its own and a table of the mapping wants words
// rather than the numbers the sixteen colours really are.
func ansiName(col code.AnsiColour) string {
	switch col {
	case code.AnsiBlack:
		return "AnsiBlack"
	case code.AnsiRed:
		return "AnsiRed"
	case code.AnsiGreen:
		return "AnsiGreen"
	case code.AnsiYellow:
		return "AnsiYellow"
	case code.AnsiBlue:
		return "AnsiBlue"
	case code.AnsiMagenta:
		return "AnsiMagenta"
	case code.AnsiCyan:
		return "AnsiCyan"
	case code.AnsiWhite:
		return "AnsiWhite"
	case code.AnsiBrightBlack:
		return "AnsiBrightBlack"
	case code.AnsiBrightRed:
		return "AnsiBrightRed"
	case code.AnsiBrightGreen:
		return "AnsiBrightGreen"
	case code.AnsiBrightYellow:
		return "AnsiBrightYellow"
	case code.AnsiBrightBlue:
		return "AnsiBrightBlue"
	case code.AnsiBrightMagenta:
		return "AnsiBrightMagenta"
	case code.AnsiBrightCyan:
		return "AnsiBrightCyan"
	case code.AnsiBrightWhite:
		return "AnsiBrightWhite"
	}
	return "AnsiDefault"
}
