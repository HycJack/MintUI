package code

import (
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// CodeViewerOptions configure a CodeViewer.
type CodeViewerOptions struct {
	// Name is what the view is called — a file's path, a name, a revision. It
	// is required, and it goes in the viewer's own label rather than in its
	// head: a viewer with no name is a rectangle of coloured text that
	// nothing can say anything about.
	Name string
	// Source is the text. It is a string and not a []string because a caller
	// reading a file has a string, and splitting it here would be a second
	// answer to how a file's lines end.
	Source string
	// Lang is which language's words the highlighter knows. LangPlain still
	// highlights strings and numbers, which is what a config file and a shell
	// transcript both want.
	Lang Lang
	// First is the number the first line shown is given, which is 1 for a
	// whole file and something else for a fragment of one. It is why the
	// viewer takes lines as a string: a caller showing the last forty lines
	// of a log file does not have the rest of it in memory.
	First int
	// Current is the number of the line the caret is on, or zero for none. It
	// is a number and not an index so that a caller holding a position in a
	// file can hand it over without converting it.
	Current int
	// Query is what to mark and Find says whether to mark it at all. Empty is
	// the ordinary case and costs nothing: a viewer with no search in it does
	// not compute a match for every line of a file of three thousand.
	Query string
	Find  bool
	// Gutter draws the line numbers. It is on by default and off for a view
	// of a fragment whose numbers would be a lie.
	Gutter *bool
	// Folds are the blocks a person has shut, where they are and whether
	// they are shut. The caller owns them: which lines are foldable is a
	// question about a language's braces, and a scanner is not the thing that
	// should be answering it.
	Folds []Fold
	// Height is the viewport's own height. It is required: a viewer with no
	// height grows to fit a file of three thousand lines and draws all of
	// them, which is the whole reason this component is not a column of text.
	Height float32
	// Width is the viewer's own width, for a pane in a window that knows it.
	Width float32
	// State is the viewer's place among the lines. It is the caller's, so a
	// window showing one file in two panes scrolls them together.
	State *ui.ListState
	// Scroll is where the view is scrolled sideways, for a caller that wants
	// to restore where a person was. Nil scrolls and keeps no place.
	Scroll *ui.ScrollState
	// Caption is what the head says under the name — a branch, a revision, a
	// count of what changed. Empty draws none.
	Caption string
}

// Fold is one block a person has shut or opened, and where it starts.
type Fold struct {
	// Line is the number of the line the fold marker sits on.
	Line int
	// Hidden is that the block after it is not shown.
	Hidden bool
}

// CodeViewerResult carries a CodeViewer and what was done in it.
type CodeViewerResult struct {
	// Element is the viewer: the gutter, the rows, and the head.
	Element *ui.Element
	// changed is that the caret moved this frame.
	changed bool
	// toggled is the line whose fold was pressed this frame, and zero for
	// none. It is a number rather than a bool because the caller owns the
	// folds and the only thing it can act on is which one changed.
	toggled int
}

// Changed reports that the caret moved this frame — a press on a row rather
// than the caller setting it.
func (r CodeViewerResult) Changed() bool { return r.changed }

// Toggled is the line whose fold marker was pressed this frame, or zero. The
// caller keeps the folds and decides what to do with it; the viewer only says
// that one changed, because only the caller knows what folding a block means
// for its document.
func (r CodeViewerResult) Toggled() int { return r.toggled }

// CodeViewer is a file with line numbers and highlighting.
//
// The rows are built in a ui.List rather than as a column of boxes, so a file
// of three thousand lines draws the forty it has room for. That is the whole
// reason this component exists rather than a column of ui.Text: a column of
// three thousand costs the frame everything below the crop, on every resize,
// and on every keystroke that scrolls it.
func CodeViewer(c *ui.Context, opts CodeViewerOptions) CodeViewerResult {
	if opts.Name == "" {
		panic("code: CodeViewer needs a Name; a viewer with no name is a rectangle of coloured " +
			"text that nothing can say anything about")
	}
	if opts.Height <= 0 {
		panic("code: CodeViewer needs a Height; without one it draws every line of the file")
	}
	lines := splitLines(opts.Source)
	if opts.First <= 0 {
		opts.First = 1
	}
	gutter := true
	if opts.Gutter != nil {
		gutter = *opts.Gutter
	}
	rows := CodeRowsOptions{
		Lines: len(lines), First: opts.First, Current: opts.Current,
		Gutter: gutter, CurrentBand: true,
	}
	cache := viewerTokens(c, opts.Name, opts.Lang, opts.Source)

	// Whether the first shown line is inside a block comment is a question
	// about the lines before it, and for a viewer of the tail of a file it is
	// a question about lines the caller never handed over. The comment state
	// is therefore worked out from the top of what there is, and a caller
	// showing a fragment that begins mid-comment passes LangPlain or accepts
	// that the first line is shown as code.
	inComment := false
	for i := range opts.First - 1 {
		if i < len(lines) {
			inComment = blockCommentOpenAfter(lines[i], opts.Lang, inComment)
		}
	}

	var r CodeViewerResult
	// The body is a function, not an element, so that CodePanel can
	// draw the header above it.
	view := func() {
		host := layout.ScrollArea(c, layout.ScrollAreaOptions{
			Vertical: true, Horizontal: true, Height: opts.Height, State: opts.Scroll,
		}, nil).Element.FillWidth()
		if opts.Width > 0 {
			host.Width(opts.Width).Shrink(0)
		}
		host.Children(func() {
			// The list carries the viewport's height itself. A list inside a
			// scroll area with none is a scroll area of no height, which grows
			// to fit its content and draws every row — and a viewer that draws a
			// thousand lines to show forty is the thing this component exists
			// not to be.
			list := ui.List(c, opts.State, len(lines), func(i int) {
				number := opts.First + i
				// The row's own index and its number are two different things and
				// both are needed: the fold is keyed by the number a person sees,
				// and the cache by the index the caller counted by.
				if foldAt(opts.Folds, number) {
					return
				}
				if cache[i] == nil {
					before := inComment
					if i > 0 && i-1 < len(lines) {
						before = blockCommentOpenAfter(lines[i-1], opts.Lang, inComment)
					}
					cache[i] = Highlight(lines[i], opts.Lang, before)
				}
				var marks []MatchSpans
				if opts.Find && opts.Query != "" {
					marks = FindMatches(lines[i], opts.Query)
				}
				row := codeRow(c, rows, number, func() {
					codeTokens(c, cache[i], theme.RowSize, ui.Color{}, marks)
				})
				if isFoldable(lines[i]) {
					row.Children(func() {
						marker := ui.Box(c).Size(theme.RowSize*1.5, theme.RowSize*1.5).
							Shrink(0).Cursor(ui.CursorPointer).Role(ui.RoleDisclosure)
						marker.Label(core.Msg(c, "code.foldBlock", core.Def("Fold this block")))
						marker.Draw(func(p *ui.Painter, box ui.Rect) {
							foldChevron(p, box, !foldHidden(opts.Folds, number))
						})
						if marker.Clicked() {
							setFold(opts.Folds, number)
							r.toggled = number
							// The rows shown are not the rows of the frame
							// before, and nothing else in the window would ask
							// for the frame that shows the change.
							c.Invalidate()
						}
					})
				}
				// The row takes the press, which is how the caret moves. It is the
				// row and not the text inside it because the text is several
				// elements, and a press on any of them should mean the same thing.
				hit := ui.Box(c).Grow(1).FillHeight().Shrink(0).Role(ui.RoleNone)
				hit.Label("Line " + itoa(number))
				if hit.Clicked() {
					rows.Current, opts.Current = number, number
					r.changed = true
				}
			}).FillWidth().Height(opts.Height)
			if opts.Width > 0 {
				list.Width(opts.Width).Shrink(0)
			}
			if opts.Scroll != nil {
				list.TrackScroll(opts.Scroll)
			}
			// The list is the machinery, not a level of the thing being shown:
			// what a reader is in is a document.
			list.Role(ui.RoleNone)
		})

	}
	host := CodePanel(c, view, layout.ContainerOptions{Width: opts.Width}, func() {
		CodeHeader(c, opts.Name, nil, func() {
			if opts.Caption != "" {
				CodeCaption(c, opts.Caption)
			}
			CodeBadge(c, itoa(len(lines))+" lines", core.Neutral)
		})
	})

	r.Element = host
	return r
}

// viewerTokens is the cache of one viewer's tokens, on the window root.
//
// It is keyed by the source's own text rather than by a pointer to it, because
// a caller reading a file has a new string every time and a cache keyed by
// identity would miss every time. Putting the source in the key means an edit
// gives a different key, which is the one case where a stale answer would be
// wrong rather than merely slow.
func viewerTokens(c *ui.Context, name string, lang Lang, source string) map[int][]Token {
	type cacheKey struct {
		name   string
		lang   Lang
		source string
	}
	k := cacheKey{name: name, lang: lang, source: source}
	all := core.Local(c, viewerCacheKey{}, func() map[cacheKey]map[int][]Token {
		return map[cacheKey]map[int][]Token{}
	})
	if m, ok := all[k]; ok {
		return m
	}
	m := map[int][]Token{}
	all[k] = m
	return m
}

// viewerCacheKey is where the token caches hang, on the window root.
type viewerCacheKey struct{}

// foldAt reports whether a line is inside a block somebody has shut.
//
// The folds are in the caller's hands and are read as they are given, so this
// is a walk of the fold list rather than a parse: the nearest fold above the
// line owns it, and a fold that is open owns nothing.
func foldAt(folds []Fold, line int) bool {
	for _, f := range folds {
		if f.Line >= line {
			break
		}
		if !f.Hidden {
			continue
		}
		// A fold owns the lines after it up to the next fold, whichever way
		// the next fold is set: the next fold is a sibling of this one, so
		// whatever state it is in it is its own block's business.
		return true
	}
	return false
}

// foldHidden is whether the block at a line is shut.
func foldHidden(folds []Fold, line int) bool {
	for _, f := range folds {
		if f.Line == line {
			return f.Hidden
		}
	}
	return false
}

// setFold shuts or opens the block at a line, adding it if it is not there.
func setFold(folds []Fold, line int) {
	for i := range folds {
		if folds[i].Line == line {
			folds[i].Hidden = !folds[i].Hidden
			return
		}
	}
	folds = append(folds, Fold{Line: line, Hidden: true})
}

// isFoldable is whether a line is one a caller would put a fold on: one that
// opens something. A closing brace is not, because folding at it would hide
// the block it closes and leave its opening line on its own.
func isFoldable(line string) bool {
	t := strings.TrimSpace(line)
	if t == "" || strings.HasPrefix(t, "}") || strings.HasPrefix(t, "]") || strings.HasPrefix(t, ")") {
		return false
	}
	return strings.HasSuffix(t, "{") || strings.HasSuffix(t, "[") ||
		strings.HasSuffix(t, "(") || strings.HasSuffix(t, ":")
}

// splitLines is the text as lines, without inventing a last empty one.
//
// A file ending in a newline has a line after it that nobody wrote, and a
// viewer that shows an empty last line for every file that ends properly is a
// viewer that is always one line wrong at the end.
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

// foldChevron is the arrow in front of a folded line, drawn rather than
// borrowed because it is a triangle and MyGo's own is private.
func foldChevron(p *ui.Painter, r ui.Rect, open bool) {
	col := p.Theme().TextMuted
	cx, cy := r.X+r.W/2, r.Y+r.H/2
	s := r.W * 0.14
	var path ui.Path
	if open {
		// Pointing down: the block is shown.
		path.MoveTo(cx-s, cy-s/1.6).LineTo(cx+s, cy-s/1.6).LineTo(cx, cy+s)
	} else {
		// Pointing right: the block is folded away.
		path.MoveTo(cx-s/1.6, cy-s).LineTo(cx-s/1.6, cy+s).LineTo(cx+s, cy)
	}
	p.FillPath(&path, col)
}
