package code

import (
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// The four things a build window shows around a viewer: its output, what the
// compiler complained about, what is in the file, and what was searched for.

// ── OutputPanel ─────────────────────────────────────────────────────────────

// OutputOptions configure an OutputPanel.
type OutputOptions struct {
	// Name is what the panel is called — "build", "test", "deploy" — and is
	// required.
	Name string
	// Lines are the lines of output, as LogView takes them, so that a caller
	// with a build's output and a caller with a test run's are drawn by the
	// same function and cannot come out looking like two different things.
	Lines []LogLine
	// Height is the viewport's own height, and is required.
	Height float32
	// Width is the panel's own width.
	Width float32
	// Filter is the lowest level shown.
	Filter LogLevel
	// State and Scroll are where the panel is.
	State  *ui.ListState
	Scroll *ui.ScrollState
	// Follow keeps the end in view as lines arrive.
	Follow *bool
	// Running is that the work is still going, which is what the head's mark
	// says and what a stopped spinner would otherwise hide.
	Running bool
	// Detail draws each line's time, level and source.
	Detail *bool
}

// OutputResult carries an OutputPanel and what was pressed in it.
type OutputResult struct {
	// Element is the panel.
	Element *ui.Element
	// pressed is the row pressed this frame, counted from zero, or -1.
	pressed int
	// running reports that the work is still going, which is what a caller
	// needs to stop showing a spinner and what it cannot work out for itself:
	// a spinner that outlives its work is worse than no spinner.
	running bool
	// lines is how many passed the filter.
	lines int
}

// Pressed is the row pressed this frame, and -1 for none.
func (r OutputResult) Pressed() int { return r.pressed }

// Running reports that the work is still going.
func (r OutputResult) Running() bool { return r.running }

// Lines is how many lines passed the filter.
func (r OutputResult) Lines() int { return r.lines }

// OutputPanel is the output of a piece of work — a build, a test run, a
// deploy — with its level column and its filter.
//
// It is LogView with a head, and it is written as its own component rather
// than as a LogView with options because the two answer different questions:
// a log is something to read and an output is something to watch, and the head
// that says whether the work is still going belongs to the second and not to
// the first.
func OutputPanel(c *ui.Context, opts OutputOptions) OutputResult {
	if opts.Name == "" {
		panic("code: OutputPanel needs a Name; output that says nothing about what produced it " +
			"cannot be told from any other output")
	}
	if opts.Height <= 0 {
		panic("code: OutputPanel needs a Height; without one it draws every line of output")
	}
	detail := true
	if opts.Detail != nil {
		detail = *opts.Detail
	}
	follow := false
	if opts.Follow != nil {
		follow = *opts.Follow
	}

	var r OutputResult

	// OutputPanel is a LogView with a word on its head, not a panel around
	// one: LogView already draws a box with a border and a header, and
	// wrapping that in a second box puts a border inside a border and two
	// headers where one belongs. So the "Running" badge is handed down to
	// the log's own header instead.
	var extra *ExtraBadge
	if opts.Running {
		extra = &ExtraBadge{
			Text:     core.Msg(c, "code.running", core.Def("Running")),
			Severity: core.Accent,
		}
	}
	log := LogView(c, LogOptions{
		Name: opts.Name, Lines: opts.Lines, Height: opts.Height, Width: opts.Width,
		Filter: opts.Filter, State: opts.State, Scroll: opts.Scroll,
		ShowLevel: boolOf(detail), Follow: boolOf(follow),
		Extra: extra,
	})
	r.pressed, r.lines, r.running = log.Pressed(), log.Visible(), opts.Running
	r.Element = log.Element
	return r
}

func boolOf(b bool) *bool { return &b }

// levelFilterText is what the panel says it is showing, for a filter that is
// not showing everything. It goes through core.Msg like every other word in
// this library, so a window can reword it.
func levelFilterText(c *ui.Context, f LogLevel) string {
	return core.Msg(c, "code.filterLevel",
		core.Def("showing ")+strings.ToLower(f.String())+" and above")
}

// ── ProblemsPanel ───────────────────────────────────────────────────────────

// Problem is one thing a compiler, a linter or a type checker complained about.
type Problem struct {
	// Message is what it said, and is required: a problem with no message is
	// a mark in a list.
	Message string
	// File and Line are where it is, Line from one.
	File string
	Line int
	// Column is which character on the line, counted from one, and zero for
	// the whole line.
	Column int
	// Severity is how bad it is, and it goes through core.Severity for the
	// same reason everything else in this library does: the words and the
	// colour must not be able to disagree.
	Severity core.Severity
	// Source is what said it — a compiler, a linter, a rule name — and is
	// empty for one that has only one.
	Source string
	// Fixed is that it has been dealt with. A fixed problem is kept and shown
	// struck through, because a linter that removes its findings from the list
	// the moment they are fixed makes the list jump under the reader.
	Fixed bool
}

// ProblemsOptions configure a ProblemsPanel.
type ProblemsOptions struct {
	// Name is what the panel is called, and is required.
	Name string
	// Problems are what was found, and they are the caller's slice.
	Problems []Problem
	// Height is the viewport's own height, and is required.
	Height float32
	// Width is the panel's own width.
	Width float32
	// ErrorsOnly keeps only the problems that are errors, which is what a
	// build's summary line counts.
	ErrorsOnly bool
	// State is where the panel is among its rows.
	State *ui.ListState
	// Selected is the row chosen, counted from zero.
	Selected int
}

// ProblemsResult carries a ProblemsPanel and what was pressed in it.
type ProblemsResult struct {
	// Element is the panel.
	Element *ui.Element
	// pressed is the problem pressed this frame, counted from zero, or -1.
	pressed int
	// shown is how many were shown after the filter.
	shown int
	// errors is how many of those are errors.
	errors int
}

// Pressed is the problem pressed this frame, and -1 for none.
func (r ProblemsResult) Pressed() int { return r.pressed }

// Shown is how many problems passed the filter.
func (r ProblemsResult) Shown() int { return r.shown }

// Errors is how many of the shown problems are errors.
func (r ProblemsResult) Errors() int { return r.errors }

// ProblemsPanel is what a compiler found, in the order it said it.
//
// It is in the order the tool reported rather than sorted by severity,
// because the order a compiler reports is the order it found them in and
// sorting it puts the last thing it found — the one that stopped it — at the
// top of a list a reader is part-way down.
func ProblemsPanel(c *ui.Context, opts ProblemsOptions) ProblemsPanelResult {
	return problemsPanel(c, opts)
}

// ProblemsPanelResult is what a ProblemsPanel carries. It is a separate name
// from the options because the catalogue's name and the reader's name should
// not have to differ, and a caller writing both on one line should not have to
// spell them differently.
type ProblemsPanelResult = ProblemsResult

// problemsPanel is ProblemsPanel's body, so that the exported name and the
// result name do not have to be two spellings of one thing.
func problemsPanel(c *ui.Context, opts ProblemsOptions) ProblemsPanelResult {
	if opts.Name == "" {
		panic("code: ProblemsPanel needs a Name; a list of messages that says nothing about which " +
			"tool said them cannot be acted on")
	}
	if opts.Height <= 0 {
		panic("code: ProblemsPanel needs a Height; without one it draws every problem")
	}
	shown := make([]Problem, 0, len(opts.Problems))
	errors := 0
	for _, p := range opts.Problems {
		if opts.ErrorsOnly && p.Severity != core.Danger {
			continue
		}
		shown = append(shown, p)
		if p.Severity == core.Danger && !p.Fixed {
			errors++
		}
	}

	var r ProblemsPanelResult
	r.pressed, r.shown, r.errors = -1, len(shown), errors

	// The body is a function, not an element, so that CodePanel can draw the
	// header above it rather than below it.
	view := func() {
		scroll := ui.Scroll(c).Height(opts.Height).FillWidth()
		scroll.Children(func() {
			ui.List(c, opts.State, len(shown), func(i int) {
				p := shown[i]
				at := i
				row := problemRow(c, p, at == opts.Selected)
				row.Children(func() {
					if row.Clicked() {
						r.pressed = at
					}
				})
			}).FillWidth().Role(ui.RoleNone)
		})
	}
	host := CodePanel(c, view, layoutContainer(opts.Width), func() {
		CodeHeader(c, opts.Name, nil, func() {
			// The errors are a badge of their own and not folded into the
			// count: a build says "3 problems" and a person needs to know
			// how many of them stop it without counting anything.
			if errors > 0 {
				CodeBadge(c, itoa(errors)+" errors", core.Danger)
			}
			CodeBadge(c, itoa(len(shown))+" problems", problemSeverity(shown))
		})
	})

	r.Element = host
	return r
}

// problemSeverity is the worst severity in a list, for the head's badge.
func problemSeverity(all []Problem) core.Severity {
	worst := core.Neutral
	for _, p := range all {
		if p.Fixed {
			continue
		}
		switch p.Severity {
		case core.Danger:
			return core.Danger
		case core.Warning:
			worst = core.Warning
		}
	}
	return worst
}

func problemRow(c *ui.Context, p Problem, chosen bool) *ui.Element {
	k := core.Tokens(c)
	_ = k
	// The height is the package's own line height, stated rather than left to
	// the tallest child: the message is drawn at the row size and the source
	// beside it is a pill, and a pill is shorter than a line of text, so a row
	// left to measure itself comes out a couple of points short and the next
	// row's words land on top of this one's.
	row := ui.Row(c).FillWidth().Shrink(0).AlignItems(ui.Center).
		Height(metricsOf(c, 1).line).
		Role(ui.RoleListItem).Label(p.Message)
	if chosen {
		row.Background(k.Surface)
	}
	row.Children(func() {
		// The mark is the severity and the words are the message, and the
		// mark leads because it is the thing a reader is scanning the list
		// for. It repeats the severity in shape as well as in colour, which is
		// the one thing that survives somebody who cannot see the colour.
		glyph, ink := severityGlyph(p.Severity, k)
		if glyph != nil {
			ui.Icon(c, glyph).TextColor(ink).Size(theme.CaptionSize*1.3, theme.CaptionSize*1.3).Shrink(0)
			ui.Box(c).Width(theme.CaptionSize * 0.5).Shrink(0)
		}
		// The message takes the room the row has left and nothing else: the
		// Grow belongs on the box, which is in a row and so grows along the
		// width. On the text — inside a box, which is a column — it asks for
		// height the box does not have, and the message ends up drawn half a
		// row below the place and the source beside it.
		ui.Box(c).Grow(1).FillHeight().Shrink(0).Children(func() {
			text := ui.Text(c, p.Message).SingleLine().Shrink(1).
				FontSize(core.FontSize(c, theme.RowSize)).TextColor(ink)
			if p.Fixed {
				text.Strikethrough().TextColor(k.TextFaint)
			}
		})
		where := ""
		if p.File != "" {
			where = p.File + ":" + itoa(p.Line)
			if p.Column > 0 {
				where += ":" + itoa(p.Column)
			}
		}
		if where != "" {
			CodeCaption(c, where)
		}
		if p.Source != "" {
			CodeBadge(c, p.Source, core.Neutral)
		}
	})
	return row
}

// severityGlyph is the mark a severity is drawn as, and it is drawn rather than
// picked from ui/input's because those are private to that package and a second
// copy would be two sets of marks that drift.
func severityGlyph(sev core.Severity, k theme.Tokens) (*ui.SVG, ui.Color) {
	_, fg := sev.Pair(k)
	switch sev {
	case core.Danger:
		return glyphCrossCircle, k.Danger
	case core.Warning:
		return glyphWarnTriangle, k.Warning
	case core.Success:
		return glyphTickCircle, k.Success
	case core.Accent:
		return glyphInfoCircle, k.Accent
	}
	return nil, fg
}

var (
	glyphCrossCircle  = mustGlyph(`<circle cx="12" cy="12" r="8.5"/><path d="M12 7.5v5.5"/><path d="M12 16.4v.6"/>`)
	glyphWarnTriangle = mustGlyph(`<path d="M12 4.75 21 19.5H3Z"/><path d="M12 10.5V14.5"/><path d="M12 17v.6"/>`)
	glyphTickCircle   = mustGlyph(`<circle cx="12" cy="12" r="8.5"/><path d="M8 12.2l2.8 2.8L16.2 9.4"/>`)
	glyphInfoCircle   = mustGlyph(`<circle cx="12" cy="12" r="8.5"/><path d="M12 11.25V17"/><path d="M12 7.6v.6"/>`)
)
