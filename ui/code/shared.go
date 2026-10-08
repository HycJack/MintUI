package code

import (
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// The parts every line-showing component in this package is built from, and
// the one rule they all obey.
//
// The rule is that a line number is drawn in exactly one place and the current
// line's band in exactly one place, whichever component is showing the line.
// It is the same rule ui/input's ringField is written for, and for the same
// reason: a field marked by one component and outlined by another comes out
// with two borders. A viewer whose gutter is one pixel narrower than its
// diff's makes every line in the file look misaligned when the two are shown
// side by side, and the only way that never happens is if there is nothing to
// keep in step — one function, called by both.

// MonoStack is the font stack this package draws code in.
//
// It is a stack rather than a name because a name is a guess on every
// platform: SF Mono on a Mac, Menlo on one that has it, Cascadia on a Windows
// that has it, and the system's own monospace everywhere else. A viewer whose
// code fell back to the proportional face would be unreadable in a way that is
// not obvious — the words would all be there, and the columns would not line
// up.
const MonoStack = "SF Mono, Menlo, Consolas, Cascadia Mono, monospace"

// metrics is what a line of code is measured in, at this window's density.
type metrics struct {
	// line is how tall one line of text is, and the row height: the rows are
	// spaced by this and nothing else, so a viewer, a diff and a hex view all
	// agree on where line 40 is without being told.
	line float32
	// gutter is how wide the line-number column is: enough for the widest
	// number the file has, and the same in every component whatever the file.
	gutter float32
	// pad is the padding beside a line's text.
	pad float32
	// size is the type size, which is the row height's other half.
	size float32
}

// metricsOf is the measure of code at this window's density, given how many
// lines there are.
//
// The gutter width comes from the number of lines and not from a constant,
// because a hundred-line file and a ten-thousand-line one have different
// widths and a constant gutter is either a wasted margin on the short one or
// a line number that has nowhere to go on the long one.
func metricsOf(c *ui.Context, lines int) metrics {
	u := core.Density(c).Unit()
	m := metrics{
		line: core.FontSize(c, theme.RowSize) * 1.5,
		pad:  u * 2,
		size: theme.RowSize,
	}
	digits := len(itoa(max(lines, 1)))
	m.gutter = float32(digits)*m.size*0.62 + u*2
	return m
}

// CodeMetrics is what a caller needs to size a window around code: the height
// of a line, so that a caller showing forty lines can say so, and the width of
// the gutter, so that one placing something beside the numbers can leave room.
func CodeMetrics(c *ui.Context, lines int) (lineHeight, gutterWidth float32) {
	m := metricsOf(c, lines)
	return m.line, m.gutter
}

// LineHeight is how tall one line of code is at this window's density. It is
// exported because a caller showing a fixed number of lines needs the number
// and cannot measure an element — an empty box measures zero, which is the
// trap docs/design-system.md §15.2 is about.
func LineHeight(c *ui.Context) float32 { return metricsOf(c, 1).line }

// codeMono is a run of code text in this package's font, which every
// component here reaches rather than writing out: a viewer and a hex view
// whose code is a tenth of a point apart is a pair of panes that do not line
// up, and that is the whole job of a hex view next to a text view.
func codeMono(c *ui.Context, text string, size float32, col ui.Color) *ui.Element {
	return ui.Text(c, text).Font(MonoStack).
		FontSize(core.FontSize(c, size)).TextColor(col).SingleLine().Shrink(0)
}

// CodeMono is one run of code text in this package's font and the window's
// own text colour. It is exported because the panels around a viewer — a
// caption, a tooltip, a path in a tab — are code too, and a caption in the
// proportional face beside a viewer in the monospaced one is the same mistake
// one size smaller.
func CodeMono(c *ui.Context, text string, size float32) *ui.Element {
	k := core.Tokens(c)
	return codeMono(c, text, size, k.Text)
}

// CodeRows is what a caller says about the lines it is showing, so that every
// line-showing component in this package draws them the same way.
type CodeRowsOptions struct {
	// Lines is how many lines the content has, which is what the gutter is
	// measured against.
	Lines int
	// First is the number of the first line shown, counted from one the way a
	// file's own line numbers are. Zero shows no numbers at all, which is what
	// a diff's right-hand side wants and a viewer's left does not.
	First int
	// Current is the number of the line the caret is on, and zero for none.
	// The band is drawn behind it, once, here.
	Current int
	// Selection is the lines the caller has chosen, and is empty for none.
	Selection []int
	// Gutter draws the numbers. It is a caller choice rather than the default
	// because a log view has no line numbers worth drawing and a hex view has
	// offsets that are not line numbers at all.
	Gutter bool
	// CurrentBand tints the current line's band. Off for a diff, where both
	// sides are changes and neither is current.
	CurrentBand bool
	// SelectionBand tints the chosen lines, which is how a multi-select shows
	// itself without a border round every row.
	SelectionBand bool
	// BandColour overrides the current line's tint, for a caller whose current
	// line is a search hit or a breakpoint rather than a caret. Zero alpha
	// takes the window's own.
	BandColour ui.Color
}

// codeRow is one line: its background, its number, and its content.
//
// It is the one place in this package that paints a line's ground or a line's
// number, and both of those are behind a content callback rather than beside
// it. The order matters and is not negotiable: the band has to be the first
// thing drawn or it lands on top of the text, and the number has to be a
// child of the row or it ends up a sibling and sits in the wrong place.
func codeRow(c *ui.Context, o CodeRowsOptions, number int, content func()) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	m := metricsOf(c, o.Lines)

	row := ui.Row(c).FillWidth().Height(m.line).AlignItems(ui.Center).
		Role(ui.RoleNone).Label(codeRowLabel(o, number))
	if o.CurrentBand && o.Current != 0 && number == o.Current {
		band := o.BandColour
		if band.A == 0 {
			band = k.SurfaceHover
		}
		row.Background(band)
	}
	if o.SelectionBand && hasInt(o.Selection, number) {
		row.Background(k.Surface)
	}
	row.Children(func() {
		if o.Gutter && o.First > 0 {
			// Right-aligned so that a line number ends where the line ends
			// rather than where its own first digit starts: a gutter of
			// left-aligned numbers has a ragged edge that the eye reads as a
			// column of something.
			text := ""
			if number >= o.First {
				text = itoa(number)
			}
			num := ui.Row(c).Width(m.gutter).Shrink(0).Justify(ui.End).
				Padding(0, u, 0, u).Role(ui.RoleNone)
			num.Children(func() {
				ui.Text(c, text).Font(MonoStack).SingleLine().Shrink(0).
					FontSize(core.FontSize(c, theme.CaptionSize)).
					TextColor(gutterInk(c, o, number))
			})
		}
		// Row, not Box: a Box here is a Column, so everything a caller
		// puts in a line lands under each other instead of across. A viewer
		// draws one run of tokens and never notices. LogView draws a time, a
		// level, a source and a message, and with a Column under it every
		// row is four rows tall and overflows into the next one — four
		// columns of text, each stepping down one line, which reads as a
		// font that will not sit still rather than as a layout that is
		// wrong.
		//
		// AlignItems(Center) for the same reason the row has it: a caption
		// and a badge have different heights and both have to end up on the
		// line's middle.
		ui.Row(c).Grow(1).Shrink(0).AlignItems(ui.Center).Children(func() {
			if content != nil {
				content()
			}
		})
	})
	return row
}

// codeRowLabel names a line for assistive technology and for a test to find
// it by. It is on the row rather than on the number or the text because both
// of those are optional and this is the one thing a line always has.
func codeRowLabel(o CodeRowsOptions, number int) string {
	if number <= 0 {
		return ""
	}
	return "Line " + itoa(number)
}

// gutterInk is the colour a line number is drawn in.
//
// The current line's number is in the window's text and the rest in the faint
// ink: the number beside the line the caret is on is the one being read, and
// making it the same as the other forty-nine would be leaving the reader to
// work out which line they are on by looking at the text rather than at the
// edge.
func gutterInk(c *ui.Context, o CodeRowsOptions, number int) ui.Color {
	k := core.Tokens(c)
	switch {
	case o.Current != 0 && number == o.Current:
		return k.Text
	case hasInt(o.Selection, number):
		return k.Text
	}
	return k.TextFaint
}

// codeTokens draws one line's worth of highlighted text, with any search
// matches marked behind them.
//
// It takes the tokens rather than the line so that every component colours the
// same kinds the same way: a diff's added line and a viewer's line of the same
// text are one drawing function called twice, and a second set of colour
// decisions is the thing that would drift.
//
// The matches are marked by splitting the runs rather than by painting over
// the row afterwards. Marking over the row would need the width of a
// character in pixels, which is not knowable while the frame is being built —
// and a mark one character out of place is worse than no mark, because it
// makes the wrong word look like the one that was found.
// spaceBox draws the room a run of whitespace would have taken. Layout trims
// the whitespace at the end of the text it shapes, and a run sliced off a
// line — a code token, an ANSI span — is nothing but that, so the gap between
// two runs is drawn as a bare box of the measured advance instead. The text
// the reader copies and the screen reader hears stays what was scanned.
//
// The advance is measured with a letter on either end, because a measurement
// of bare whitespace is shaped into nothing at all — the very trimming this
// exists to undo.
func spaceBox(c *ui.Context, ws string, size float32) func() {
	with, _ := c.MeasureText(0, ui.Span{Text: "n" + ws + "n", Font: MonoStack, Size: size})
	bare, _ := c.MeasureText(0, ui.Span{Text: "nn", Font: MonoStack, Size: size})
	w := with - bare
	return func() {
		ui.Box(c).Width(w).Height(1).Shrink(0)
	}
}

// splitTrailingWS cuts a run's trailing spaces and tabs off, which are the
// part layout would not draw anyway. The rest is the run's visible text.
func splitTrailingWS(text string) (body, ws string) {
	cut := strings.TrimRight(text, " \t")
	return text[:len(cut)], text[len(cut):]
}

func codeTokens(c *ui.Context, tokens []Token, size float32, base ui.Color, marks []MatchSpans) *ui.Element {
	k := core.Tokens(c)
	row := ui.Row(c).AlignItems(ui.Center).Gap(0).Shrink(0)
	fontSize := core.FontSize(c, size)
	at := 0
	for _, t := range tokens {
		if t.Text == "" {
			continue
		}
		// A whitespace token is all trailing whitespace by definition, so it
		// is a box of measured width and nothing else.
		if strings.TrimSpace(t.Text) == "" {
			sb := spaceBox(c, t.Text, fontSize)
			row.Children(sb)
			at += len(t.Text)
			continue
		}
		body, ws := splitTrailingWS(t.Text)
		for _, piece := range splitAtMarks(body, at, marks) {
			piece := piece
			row.Children(func() {
				e := ui.Text(c, piece.text).Font(MonoStack).SingleLine().Shrink(0).
					FontSize(core.FontSize(c, size)).
					TextColor(tokenInk(k, t.Kind, base))
				if piece.marked {
					// A mark behind the run, at a quarter of the type's own
					// alpha: enough to find a hit in a line of a colour it is
					// already mixed into, quiet enough to read through.
					e.TextBackground(k.Warning.Alpha(0.3))
				}
			})
		}
		if ws != "" {
			sb := spaceBox(c, ws, fontSize)
			row.Children(sb)
		}
		at += len(t.Text)
	}
	return row
}

// markedPiece is a run of a token that is either inside a search match or not.
type markedPiece struct {
	text   string
	marked bool
}

// splitAtMarks cuts one token's text at the boundaries of the matches that
// fall inside it, so that a match spanning two tokens — a keyword and a
// string, say — is marked across both rather than half of each.
//
// The offsets in a span are into the whole line, so the token's own start is
// passed in and taken off; getting that wrong is what makes a viewer mark the
// right words on the wrong line.
func splitAtMarks(text string, lineAt int, marks []MatchSpans) []markedPiece {
	if len(marks) == 0 || text == "" {
		return []markedPiece{{text: text}}
	}
	from, to := lineAt, lineAt+len(text)
	cuts := []markBound{{at: from}}
	for _, m := range marks {
		if m.From <= from || m.To <= from || m.From >= to || m.To > to {
			continue
		}
		cuts = append(cuts, markBound{at: m.From}, markBound{at: m.To})
	}
	cuts = append(cuts, markBound{at: to})
	sortSlice(cuts, func(a, b markBound) bool { return a.at < b.at })

	var out []markedPiece
	for i := 0; i+1 < len(cuts); i++ {
		if cuts[i].at >= cuts[i+1].at {
			continue
		}
		out = append(out, markedPiece{
			text:   text[cuts[i].at-from : cuts[i+1].at-from],
			marked: isMarked(cuts, i),
		})
	}
	if len(out) == 0 {
		return []markedPiece{{text: text}}
	}
	return out
}

// markBound is one end of a cut in a token, and whether that end closes a
// match rather than opening one.
type markBound struct {
	at     int
	marked bool
}

// isMarked whether the piece starting after cuts[i] is inside a match: the
// run between two boundaries is a match when the one that ends it is a match's
// end.
func isMarked(cuts []markBound, i int) bool { return cuts[i+1].marked }

// tokenInk is what each kind of token is drawn in.
//
// Every one of them is derived from the window's own palette rather than
// written out, because the palette is what knows about the two appearances.
// The one that is not a token is the plain code colour: a file with every
// token the same colour would be unreadable, and a file where the *code* is
// the odd one out would be unreadable in the other direction — the words are
// the text, and the text is what a reader's eye follows.
func tokenInk(k theme.Tokens, kind TokenKind, base ui.Color) ui.Color {
	if base.A != 0 {
		// A caller that has given the line a colour of its own — a diff's
		// added line, a log line's level — has the text, and the tokens are
		// drawn as steps away from it rather than as the palette's own hues,
		// which would put a diff's red code on a green line.
		switch kind {
		case Keyword, Type:
			return base.Mix(k.Accent, 0.45)
		case String:
			return base.Mix(k.Success, 0.5)
		case Number:
			return base.Mix(k.Warning, 0.35)
		case Comment:
			return base.Mix(k.TextFaint, 0.6)
		case Function:
			return base.Mix(k.AccentText, 0.4)
		}
		return base
	}
	switch kind {
	case Keyword:
		return k.Accent
	case Type:
		return k.AccentText
	case String:
		return k.Success
	case Number:
		return k.Warning
	case Comment:
		return k.TextFaint
	case Function:
		return k.Lively.Mix(k.Text, 0.35)
	case Punct:
		return k.TextMuted
	}
	return k.Text
}

// TokenInk is what a kind of token is drawn in at this window's palette, for
// the panels around a viewer that have to match it: a legend, a tooltip, a
// caption saying which colours mean what.
func TokenInk(c *ui.Context, kind TokenKind) ui.Color {
	return tokenInk(core.Tokens(c), kind, ui.Color{})
}

// CodePanel is the box a line-showing component's rows go in: the ground, the
// hairline and the rounding, and nothing else.
//
// It is exported because a panel of code in a window this library draws — a
// diff in a sheet, a hex view beside a viewer — should be the same shape as
// the ones the components make themselves, and there is no way to ask a
// component for just its face.
// CodePanel is the box every panel in this package is drawn in: a header, a
// body, and a rule around the two.
//
// The header comes in as a function rather than as an element because a
// container's children are appended, and a caller that built the panel first
// and the header second would put the header at the *bottom* of the panel —
// which is where a footer goes, not a header. Taking it here means it is
// built first and drawn first, and there is no way to get the order wrong.
func CodePanel(c *ui.Context, body func(), opts layout.ContainerOptions, header func()) *ui.Element {
	k := core.Tokens(c)
	if opts.Radius == 0 {
		opts.Radius = theme.SmallRadius
	}
	// The body is built first — the caller had to build it to pass it — and
	// the header is drawn before it, because a container appends children and
	// a header added after the body lands underneath it. That is not a
	// detail: a panel whose name is printed below its contents reads as a
	// footer, and every panel here would have had one.
	panel := layout.Container(c, opts, func() {
		if header != nil {
			header()
		}
		if body != nil {
			body()
		}
	})
	panel.Background(k.Background)
	panel.BorderWidth(theme.BorderWidth).BorderColor(k.Border)
	panel.Clip()
	return panel
}

// CodeCaption is the small line of words above or below a view: a file's path,
// a count, a language. It is code's own chrome and every panel here wants
// one, so it is written once.
func CodeCaption(c *ui.Context, text string) *ui.Element {
	k := core.Tokens(c)
	if text == "" {
		return ui.Box(c).Shrink(0).Height(0)
	}
	return ui.Text(c, text).SingleLine().Shrink(0).
		FontSize(core.FontSize(c, theme.CaptionSize)).TextColor(k.TextMuted)
}

// CodeBadge is a small pill for a count or a state in a panel's head: how many
// problems, how many matches, whether a thing is running.
//
// The severity is the caller's and it goes through core.Severity rather than
// being picked here, because core.Severity is the one thing in this library
// that decides which background goes with which word, and a second decision
// about that is how two pills end up disagreeing.
func CodeBadge(c *ui.Context, text string, sev core.Severity) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	bg, fg := sev.Pair(k)
	return ui.Row(c).Shrink(0).Padding(u*0.25, u*1.5).Radius(theme.PillRadius).
		Background(bg).Label(text).Children(func() {
		ui.Text(c, text).SingleLine().Shrink(0).
			FontSize(core.FontSize(c, theme.CaptionSize)).TextColor(fg)
	})
}

// CodeHeader is a panel's head: a title, whatever the panel is counting, and
// anything the caller wants on the other side.
//
// It is one row and not three so that a panel whose title is long shrinks its
// title rather than its badge — the badge is the thing being looked at, and a
// count that has been truncated is a count nobody can read.
func CodeHeader(c *ui.Context, title string, badges []string, trailing func()) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	head := ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.5).
		Height(core.ControlHeight(c) - u).Shrink(0).Role(ui.RoleNone).
		Label(title)
	head.Children(func() {
		ui.Text(c, title).SingleLine().Shrink(1).Grow(1).
			FontSize(core.FontSize(c, theme.RowSize)).TextColor(k.Text).FontWeight(600)
		for _, b := range badges {
			CodeBadge(c, b, core.Neutral)
		}
		if trailing != nil {
			trailing()
		}
	})
	return head
}

// CodeSelectMark is the mark in front of a row a caller has chosen, which is
// the one mark this package needs: a list of breakpoints, a list of processes,
// a list of matches.
//
// It is a filled box with a tick rather than a tick on its own because a row
// in one of those panels is chosen or not and there is nothing in between,
// and a bare tick in a list of rows is a character rather than a control.
func CodeSelectMark(c *ui.Context, on bool) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	side := u * 3.75
	if !on {
		return ui.Box(c).Size(side, side).Shrink(0).Radius(side * 0.28).
			Background(k.Background).BorderWidth(theme.BorderWidth).
			BorderColor(k.Border.Mix(k.Text, 0.25)).Role(ui.RoleNone)
	}
	ink := inkOnMark(k.Accent)
	box := ui.Box(c).Size(side, side).Shrink(0).Radius(side * 0.28).
		Background(k.Accent).Role(ui.RoleNone)
	box.DrawOver(func(p *ui.Painter, r ui.Rect) {
		var path ui.Path
		path.MoveTo(r.X+r.W*0.24, r.Y+r.H*0.52).
			LineTo(r.X+r.W*0.44, r.Y+r.H*0.72).
			LineTo(r.X+r.W*0.78, r.Y+r.H*0.3)
		p.StrokePath(&path, 1.5, ink)
	})
	return box
}

// inkOnMark is the ink a tick is drawn in on the accent, asked for from the
// colour itself for the reason ui/input's markFace is: the accent is a token
// and a window may set it to anything.
func inkOnMark(bg ui.Color) ui.Color {
	white, black := ui.RGB(255, 255, 255), ui.RGB(0, 0, 0)
	if relLum(white)-relLum(bg) >= relLum(black)-relLum(bg) {
		return white
	}
	return black
}

// relLum is the relative luminance of a colour, which is what decides which
// ink reads on it.
func relLum(c ui.Color) float32 {
	lin := func(v uint8) float32 {
		f := float32(v) / 255
		if f <= 0.03928 {
			return f / 12.92
		}
		return pow32(f+0.055, 2.4)
	}
	return 0.2126*lin(c.R) + 0.7152*lin(c.G) + 0.0722*lin(c.B)
}

// CodeMatchTokens draws one line's text with a search's matches marked behind
// them. It is exported for the panels that show a match outside a row — the
// find bar's preview, a search result in a tree — so that a match looks the
// same wherever it is shown.
func CodeMatchTokens(c *ui.Context, tokens []Token, marks []MatchSpans, size float32) *ui.Element {
	return codeTokens(c, tokens, size, ui.Color{}, marks)
}

// sortSlice is a sort by a comparison on a local type, because this file has
// one place that needs it and sort.Slice would be a closure over a local type
// for the sake of four lines.
func sortSlice[T any](list []T, less func(a, b T) bool) {
	for i := 1; i < len(list); i++ {
		for j := i; j > 0 && less(list[j], list[j-1]); j-- {
			list[j], list[j-1] = list[j-1], list[j]
		}
	}
}

// pow32 is the sRGB transfer function's exponentiation, written out because
// the only use in this file is the 2.4 of the standard's formula and
// math.Pow would be an import for one call.
func pow32(f, n float32) float32 {
	out := float32(1)
	for range int(n) {
		out *= f
	}
	return out
}

// hasInt reports whether v is in list, for the few places a caller hands over
// a set of line numbers rather than a range.
func hasInt(list []int, v int) bool {
	for _, n := range list {
		if n == v {
			return true
		}
	}
	return false
}

// itoa is a small one. Every line number in this package goes through it and
// most of them are under a hundred, so it is written out rather than reaching
// for strconv and its error paths.
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
}
