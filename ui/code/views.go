package code

import (
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// layoutContainer is the box options of a view that has a width and nothing
// else to say, so that the three of them do not each write out the same two
// fields.
func layoutContainer(width float32) layout.ContainerOptions {
	return layout.ContainerOptions{Width: width}
}

// The three views that show something which is not source: a file's shape from
// the side, its bytes, and a program's output. All three go through codeRow
// for the ground and the numbers, which is what keeps their rows the same
// height as a viewer's and lets a window put them side by side.

// ── Minimap ─────────────────────────────────────────────────────────────────

// MinimapOptions configure a Minimap.
type MinimapOptions struct {
	// Name is what the map belongs to, and is required for the reason a
	// viewer's is: a strip of coloured blocks that says nothing about what it
	// is a map of.
	Name string
	// Source is the text it is a map of.
	Source string
	// Lang is which language's words it knows, so that a map of code has the
	// code's colours rather than one flat tone.
	Lang Lang
	// Width is the map's own width; zero is a narrow column, which is the
	// shape a minimap has everywhere: it is a position in a window, not a
	// second window.
	Width float32
	// LineHeight is how tall one block is, and defaults to something smaller
	// than a row: a map that is the same height as the text beside it shows
	// the same lines and tells a reader nothing.
	LineHeight float32
	// Viewport is how many lines are in view, and At is the first of them. It
	// is what the map's window is for.
	Viewport, At int
	// Scroll is where the map is scrolled when the person drags it, and the
	// caller's so that the viewer's place follows.
	Scroll *ui.ScrollState
	// Clicked is the line a press on the map landed on, as a number from one,
	// and zero for none.
	Clicked int
}

// MinimapResult carries a Minimap and what was pressed in it.
type MinimapResult struct {
	// Element is the map.
	Element *ui.Element
	// clicked is the line pressed this frame, counted from one.
	clicked int
}

// Clicked is the line a press on the map landed on, counted from one, and 0
// for none. It is reported rather than applied: moving a viewport to a line
// is the caller's business, because only the caller knows what scrolling its
// viewer means when it already knows where it is.
func (r MinimapResult) Clicked() int { return r.clicked }

// Minimap is a file shown from the side: one narrow block per line, coloured
// by what is in it, with the part being read marked.
//
// It is drawn by sampling rather than by drawing the text. A map of a file of
// three thousand lines that really drew the lines would cost as much as the
// viewer it is meant to help navigate, which is the opposite of the point.
func Minimap(c *ui.Context, opts MinimapOptions) MinimapResult {
	if opts.Name == "" {
		panic("code: Minimap needs a Name; a strip of coloured blocks that says nothing about " +
			"what it is a map of cannot be found by a screen reader")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	lines := splitLines(opts.Source)
	w := opts.Width
	if w <= 0 {
		w = u * 9
	}
	h := opts.LineHeight
	if h <= 0 {
		// A fifth of a row, so that a hundred lines fit where forty do: the
		// whole of a map is that it covers more of the file than the window
		// shows.
		h = LineHeight(c) / 5
	}
	if h < 1 {
		h = 1
	}

	viewFrom, viewTo := opts.At, opts.At+max(opts.Viewport, 1)
	var r MinimapResult
	col := ui.Column(c).Width(w).Shrink(0).Gap(0).Role(ui.RoleNone).Label(opts.Name)
	col.Children(func() {
		for i, line := range lines {
			number := i + 1
			shown := number >= viewFrom && number < viewTo
			band := ui.Color{}
			switch {
			case shown:
				band = k.SurfaceHover
			case len(line) == 0:
				band = k.Border
			}
			block := ui.Box(c).Width(w).Height(h).Shrink(0).Background(band).
				Label(opts.Name + " line " + itoa(number))
			if shown {
				block.Background(k.SurfaceHover)
			}
			// One bar per kind of token, each a fifth of the map's width, so
			// that a line of code reads as a pattern rather than as a colour:
			// a map whose lines are all one tone says only that a line has
			// something on it, and a map whose lines have bars in them says
			// roughly where the words are.
			block.Draw(func(p *ui.Painter, box ui.Rect) {
				// The window's own band wins: a line in view is drawn as the
				// window rather than as the line, because the reader is
				// looking for where the window is.
				if band.A == 0 {
					blockInk(p, box, minimapShape(line, opts.Lang), k)
				}
			})
			col.Children(func() {
				if block.Clicked() {
					r.clicked = number
					// The viewer's place is the caller's to move, but the map
					// has to be redrawn with its window somewhere else.
					if opts.Scroll != nil {
						opts.Scroll.Y = float32(i) * h
					}
				}
			})
		}
	})
	r.Element = col
	return r
}

// minimapShape is the kinds of token a line has, in order and without their
// text: a map of a line is a picture of its kinds, and the text behind it
// would be the whole file again.
func minimapShape(line string, lang Lang) []TokenKind {
	return HighlightSpans(line, lang, false)
}

// blockInk paints a line's shape: a bar per token kind, sized by how much of
// the line that kind is.
func blockInk(p *ui.Painter, r ui.Rect, kinds []TokenKind, k theme.Tokens) {
	if len(kinds) == 0 {
		return
	}
	// The first twenty kinds are enough to draw: past that the bars are a
	// pixel wide each and the shape stops being a picture of the line.
	shown := kinds
	if len(shown) > 20 {
		shown = shown[:20]
	}
	n := float32(len(shown))
	step := r.W / n
	for i, kind := range shown {
		if kind == TokenPlain {
			// Code itself is the map's own ground: a bar for every word of
			// ordinary code would be a solid block, and a solid block is the
			// one shape that says nothing.
			continue
		}
		p.Fill(ui.Rect{X: r.X + step*float32(i), Y: r.Y, W: step, H: r.H},
			tokenInk(k, kind, ui.Color{}).Alpha(0.8), 0)
	}
}

// ── HexViewer ─────────────────────────────────────────────────────────────────

// HexOptions configure a HexViewer.
type HexOptions struct {
	// Name is what the view is called — a path, a device. It is required.
	Name string
	// Data is the bytes. It is a []byte and not a hex string because every
	// caller that has the thing on disk has the bytes, and converting to hex
	// first would be a step nothing needs.
	Data []byte
	// BytesPerRow is how many bytes a row holds. Zero is sixteen, which is
	// what every hex viewer on the desktop uses and what fits a window this
	// library draws.
	BytesPerRow int
	// Height is the viewport's own height, and is required.
	Height float32
	// Address is where the data starts, so that a view of the middle of a
	// file says what offset it is at rather than starting again at zero.
	Address int
	// Width is the view's own width.
	Width float32
	// State is where the view is among its rows.
	State *ui.ListState
	// Selected is the byte the caller has chosen, as an offset from Address.
	Selected int
	// ShowAscii asks for the printable characters beside the bytes, and nil
	// asks for them. It is a pointer and not a bool because a row of bytes
	// with nothing to recognise them by is a row of bytes, so the default has
	// to be the other way round from every bool in this library.
	ShowAscii *bool
}

// HexResult carries a HexViewer and what was pressed in it.
type HexResult struct {
	// Element is the view.
	Element *ui.Element
	// selected is the offset pressed this frame, or -1.
	selected int
}

// Selected is the offset a press on the view landed on, or -1 for none. It is
// an offset into the caller's own slice rather than a row, because a caller
// that wants the byte wants the byte.
func (r HexResult) Selected() int { return r.selected }

// HexViewer is a file as its bytes: the offset, sixteen bytes, and the
// characters they spell.
//
// It is drawn rather than assembled out of boxes because a row is four
// separate alignments — the offset, the bytes, the characters — and three
// boxes per row is three times the elements for a thing that is one paint.
func HexViewer(c *ui.Context, opts HexOptions) HexResult {
	if opts.Name == "" {
		panic("code: HexViewer needs a Name; a column of bytes with nothing saying what they are " +
			"cannot be read out")
	}
	if opts.Height <= 0 {
		panic("code: HexViewer needs a Height; without one it draws every byte of the file")
	}
	perRow := opts.BytesPerRow
	if perRow <= 0 {
		perRow = 16
	}
	ascii := true
	if opts.ShowAscii != nil {
		ascii = *opts.ShowAscii
	}
	rows := (len(opts.Data) + perRow - 1) / perRow
	k, u := core.Tokens(c), core.Density(c).Unit()
	m := metricsOf(c, max(rows*perRow, 1))
	r := HexResult{selected: -1}

	// The body is a function, not an element, so that CodePanel can
	// draw the header above it.
	view := func() {
		scroll := ui.Scroll(c).Height(opts.Height).FillWidth()
		scroll.Children(func() {
			list := ui.List(c, opts.State, rows, func(i int) {
				from := i * perRow
				to := minInt(from+perRow, len(opts.Data))
				band := ui.Color{}
				if opts.Address+from == opts.Selected {
					band = k.AccentBg
				}
				row := ui.Row(c).FillWidth().Height(m.line).AlignItems(ui.Center).
					Grow(0).Shrink(0).Background(band).Font(MonoStack).
					FontSize(core.FontSize(c, theme.CaptionSize)).
					Label(opts.Name + " offset " + itoa(opts.Address+from))
				if band.A != 0 {
					row.Background(band)
				}
				row.Children(func() {
					// The offset, then the bytes in two runs of eight, then the
					// characters. The gap between the runs is what lets a row of
					// sixteen be counted by eye, which is the one thing a hex view
					// is for.
					hexRun(c, toa(opts.Address+from), m.gutter, k.TextFaint)
					hexRun(c, hexBytes(opts.Data[from:minInt(from+8, to)]), 0, k.Text)
					ui.Box(c).Width(u).Shrink(0)
					hexRun(c, hexBytes(opts.Data[minInt(from+8, to):to]), 0, k.Text)
					if ascii {
						ui.Box(c).Width(u).Shrink(0)
						hexRun(c, hexAscii(opts.Data[from:to]), 0, k.TextMuted)
					}
				})
				if row.Clicked() {
					r.selected = opts.Address + from
				}
			}).FillWidth()
			list.Role(ui.RoleNone)
		})

	}
	host := CodePanel(c, view, layoutContainer(opts.Width), func() {
		CodeHeader(c, opts.Name, nil, func() {
			CodeBadge(c, itoa(len(opts.Data))+" bytes", core.Neutral)
		})
	})

	r.Element = host
	return r
}

// hexRun is one monospaced run in a hex row.
func hexRun(c *ui.Context, text string, width float32, col ui.Color) *ui.Element {
	k := core.Tokens(c)
	e := ui.Text(c, text).Font(MonoStack).SingleLine().Shrink(0).
		FontSize(core.FontSize(c, theme.CaptionSize)).TextColor(col)
	if width > 0 {
		e = e.Width(width)
	}
	_ = k
	return e
}

// hexBytes is eight bytes as two-digit pairs with a space between them, which
// is the shape every hex viewer uses and the shape a person reads fastest.
func hexBytes(b []byte) string {
	var sb strings.Builder
	for i, v := range b {
		if i > 0 {
			sb.WriteByte(' ')
		}
		const digits = "0123456789abcdef"
		sb.WriteByte(digits[v>>4])
		sb.WriteByte(digits[v&0xf])
	}
	return sb.String()
}

// hexAscii is the printable characters a run of bytes spells, and a dot for
// each that is not.
//
// The dot is a dot and not a space on purpose: a space in a row of
// characters is indistinguishable from a space in the data, and the row has
// to line up with the bytes above it for that to be worth having.
func hexAscii(b []byte) string {
	var sb strings.Builder
	for _, v := range b {
		if v >= 0x20 && v < 0x7f {
			sb.WriteByte(v)
			continue
		}
		sb.WriteByte('.')
	}
	return sb.String()
}

func toa(n int) string { return itoa(n) }

// ── LogViewer ─────────────────────────────────────────────────────────────────

// LogLevel is how bad a line of output is.
type LogLevel int

const (
	// LogTrace is everything.
	LogTrace LogLevel = iota
	// LogDebug is the ordinary machinery.
	LogDebug
	// LogInfo is progress.
	LogInfo
	// LogWarn is something that did not go as planned.
	LogWarn
	// LogError is something that did.
	LogError
)

func (l LogLevel) String() string {
	switch l {
	case LogDebug:
		return "Debug"
	case LogInfo:
		return "Info"
	case LogWarn:
		return "Warn"
	case LogError:
		return "Error"
	}
	return "Trace"
}

// Severity is the core severity a level is drawn as, so that a log's own words
// and the pills elsewhere in the library cannot disagree about which is an
// error.
func (l LogLevel) Severity() core.Severity {
	switch l {
	case LogWarn:
		return core.Warning
	case LogError:
		return core.Danger
	case LogInfo:
		return core.Accent
	}
	return core.Neutral
}

// LogLine is one line of output.
type LogLine struct {
	// Level is how bad it is.
	Level LogLevel
	// Time is when it was written, as the caller's own format, and empty for
	// none. It is a string because a timestamp is a format and the format is
	// the caller's: a log at 14:02:03.123 and one at 3:04pm are both right
	// and the library is not going to choose between them.
	Time string
	// Text is the line, and it may hold the escape sequences a program
	// printed, which are drawn as the colours the program asked for.
	Text string
	// Source is which thing wrote it — a process, a subsystem — and is empty
	// for a log with only one.
	Source string
}

// ExtraBadge is a word in a panel's head and the colour it is drawn in.
type ExtraBadge struct {
	Text     string
	Severity core.Severity
}

// LogOptions configure a LogViewer.
type LogOptions struct {
	// Name is what the log is called, and is required.
	Name string
	// Lines are the lines of output.
	Lines []LogLine
	// Height is the viewport's own height, and is required.
	Height float32
	// Width is the log's own width.
	Width float32
	// Filter is the lowest level to show, and ShowTime, ShowLevel and
	// ShowSource say which of the row's parts are drawn. Time and level are
	// on by default because a log without them is a wall of words.
	Filter       LogLevel
	ShowTime     *bool
	ShowLevel    *bool
	ShowSource   *bool
	Follow       *bool
	Scroll       *ui.ScrollState
	State        *ui.ListState
	SelectedLine int
	// Extra is one more badge for the head, for a caller that has something
	// of its own to say. OutputPanel is the reason this exists: it is a
	// LogViewer with a "Running" on it, and wrapping a LogViewer in a second
	// panel to say so puts a border inside a border.
	Extra *ExtraBadge
}

// LogResult carries a LogViewer and what was pressed in it.
type LogResult struct {
	// Element is the log.
	Element *ui.Element
	// visible is how many lines passed the filter, which is the number a
	// status bar wants and the only one a caller cannot work out for itself
	// without repeating the filter.
	visible int
	// pressed is the row pressed this frame, or -1.
	pressed int
}

// Visible is how many lines passed the filter.
func (r LogResult) Visible() int { return r.visible }

// Pressed is the row pressed this frame, counted from zero among the visible
// lines, and -1 for none. It counts the visible rows rather than the lines,
// because a caller selecting from a log selects from what it can see.
func (r LogResult) Pressed() int { return r.pressed }

// LogViewer is a program's output, newest at the bottom, in the order it came.
//
// It filters and it scrolls, and it does not re-order: a log read top to
// bottom is a log read in the order things happened, and anything that puts
// the newest first has made a decision about what somebody is looking for
// that the caller did not ask for.
//
// The rows are built in a ui.List, so a log of ten thousand lines draws the
// hundred it has room for.
func LogViewer(c *ui.Context, opts LogOptions) LogResult {
	if opts.Name == "" {
		panic("code: LogViewer needs a Name; a column of output that says nothing about where it " +
			"came from is not a log")
	}
	if opts.Height <= 0 {
		panic("code: LogViewer needs a Height; without one it draws every line of output")
	}
	showTime, showLevel, showSource := true, true, false
	if opts.ShowTime != nil {
		showTime = *opts.ShowTime
	}
	if opts.ShowLevel != nil {
		showLevel = *opts.ShowLevel
	}
	if opts.ShowSource != nil {
		showSource = *opts.ShowSource
	}

	shown := make([]LogLine, 0, len(opts.Lines))
	for _, l := range opts.Lines {
		if l.Level < opts.Filter {
			continue
		}
		shown = append(shown, l)
	}
	rows := CodeRowsOptions{Lines: len(shown), First: 1, Gutter: false}
	k, u := core.Tokens(c), core.Density(c).Unit()

	var r LogResult
	r.visible, r.pressed = len(shown), -1
	// The body is a function, not an element, so that CodePanel can
	// draw the header above it.
	view := func() {
		scroll := ui.Scroll(c).Height(opts.Height).FillWidth()
		if opts.Scroll != nil {
			scroll.TrackScroll(opts.Scroll)
		}
		scroll.Children(func() {
			ui.List(c, opts.State, len(shown), func(i int) {
				line := shown[i]
				_, fg := line.Level.Severity().Pair(k)
				band := ui.Color{}
				if line.Level >= LogError {
					band = k.DangerBg
				}
				row := codeRow(c, rows, i+1, func() {
					if showTime && line.Time != "" {
						CodeCaption(c, line.Time)
						ui.Box(c).Width(u).Shrink(0)
					}
					if showLevel && line.Level > LogTrace {
						CodeBadge(c, line.Level.String(), line.Level.Severity())
						ui.Box(c).Width(u).Shrink(0)
					}
					if showSource && line.Source != "" {
						LogViewerName(c, line.Source)
						ui.Box(c).Width(u).Shrink(0)
					}
					// The text is drawn as the ANSI runs the program printed, so
					// that a build tool's own colours reach the screen: a red
					// "error" in the output should be red here, or the reader has
					// to read the word to know it.
					ansiRow(c, line.Text, fg, band)
				})
				if band.A != 0 {
					row.Background(band)
				}
				if row.Clicked() {
					r.pressed = i
				}
			}).FillWidth().Role(ui.RoleNone)
		})

	}
	host := CodePanel(c, view, layoutContainer(opts.Width), func() {
		CodeHeader(c, opts.Name, nil, func() {
			CodeBadge(c, itoa(r.visible)+" lines", levelSeverity(shown))
			if x := opts.Extra; x != nil && x.Text != "" {
				CodeBadge(c, x.Text, x.Severity)
			}
		})
	})

	r.Element = host
	return r
}

// levelSeverity is the worst level in a log, which is what its badge is drawn
// as: a log with an error in it is not a quiet log, and a badge that counts
// lines and says nothing about them misses the one fact worth showing.
func levelSeverity(lines []LogLine) core.Severity {
	worst := core.Neutral
	for _, l := range lines {
		if l.Level == LogError {
			return core.Danger
		}
		if l.Level == LogWarn {
			worst = core.Warning
		}
	}
	return worst
}

// LogViewerName is the name of whatever wrote a line, drawn the way a name is
// drawn everywhere else in this library: in the muted ink, at the caption
// size, and not as a heading.
func LogViewerName(c *ui.Context, name string) *ui.Element {
	k := core.Tokens(c)
	if name == "" {
		return ui.Box(c).Shrink(0)
	}
	return ui.Text(c, name).SingleLine().Shrink(0).
		FontSize(core.FontSize(c, theme.CaptionSize)).TextColor(k.TextMuted)
}

// ansiRow is one log line, drawn as the ANSI runs the program printed.
func ansiRow(c *ui.Context, text string, fg, band ui.Color) {
	// The shared ANSI part, grown to take the row's room: a line of output
	// fills its row, and a row that shrank would let the next column in.
	CodeAnsiSpans(c, text, theme.RowSize, fg, band).Grow(1).Shrink(1)
}

// ansiInk is the colour one ANSI run is drawn in, from the sixteen a terminal
// has.
//
// They are written out rather than derived, because they are the terminal's
// palette and not the window's: a program that asked for red asked for the
// red a terminal has, and giving it the window's danger colour would be
// wrong in both directions — a red that means something else in this
// library, and a darker red than the terminal's red on a dark window.
//
// The bright forms are lightened towards white rather than being a second set
// of values, which is what a bright colour is on a terminal that only has
// eight.
func ansiInk(k theme.Tokens, s AnsiSpan, base ui.Color, band ui.Color) ui.Color {
	col := ansiPalette(s.Colour, false, k)
	if col.A == 0 {
		if s.Dim {
			return k.TextFaint
		}
		if band.A != 0 {
			return base
		}
		return k.Text
	}
	_ = s.Bright
	return col
}

// ansiBackground is the ground behind one ANSI run, or nothing when the run
// asked for none. A run's background is drawn only where it asked for one:
// filling the whole log with the terminal's own background colour would fight
// the window's.
func ansiBackground(k theme.Tokens, s AnsiSpan) ui.Color {
	return ansiPalette(s.Background, s.Bright, k)
}

// ansiPalette is the sixteen terminal colours as the window's palette, with
// the bright forms lightened. Zero means the terminal's own colour, which is
// why the caller gets alpha zero back and decides what to do about it.
func ansiPalette(col AnsiColour, bright bool, k theme.Tokens) ui.Color {
	switch col {
	case AnsiBlack:
		return k.Text
	case AnsiRed:
		return ui.Hex("#cc3333")
	case AnsiGreen:
		return ui.Hex("#3f9b57")
	case AnsiYellow:
		return ui.Hex("#b08800")
	case AnsiBlue:
		return ui.Hex("#3b74d1")
	case AnsiMagenta:
		return ui.Hex("#a347ba")
	case AnsiCyan:
		return ui.Hex("#2188a0")
	case AnsiWhite:
		return k.TextMuted
	case AnsiBrightBlack:
		return k.TextMuted
	case AnsiBrightRed:
		return ui.Hex("#ff5555")
	case AnsiBrightGreen:
		return ui.Hex("#5cc76a")
	case AnsiBrightYellow:
		return ui.Hex("#e0b53f")
	case AnsiBrightBlue:
		return ui.Hex("#5b8dff")
	case AnsiBrightMagenta:
		return ui.Hex("#c268d6")
	case AnsiBrightCyan:
		return ui.Hex("#38b0c8")
	case AnsiBrightWhite:
		return k.Text
	}
	_ = bright
	return ui.Color{}
}
