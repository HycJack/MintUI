package code

import (
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// The last three panels: what is in the file, what is being looked for, and
// what could be typed next. Plus the flame graph, which is the one thing here
// that draws a chart rather than a list.

// ── SymbolOutline ───────────────────────────────────────────────────────────

// SymbolKind is what a symbol in the outline is.
type SymbolKind int

const (
	// SymbolFunc is a function or a method.
	SymbolFunc SymbolKind = iota
	// SymbolType is a type, a class or an interface.
	SymbolType
	// SymbolValue is a variable, a constant or a field.
	SymbolValue
	// SymbolSection is anything else worth a heading — a region comment, a
	// `#region`, a label.
	SymbolSection
)

func (k SymbolKind) String() string {
	switch k {
	case SymbolType:
		return "Type"
	case SymbolValue:
		return "Value"
	case SymbolSection:
		return "Section"
	}
	return "Func"
}

// Symbol is one entry of an outline.
type Symbol struct {
	// Name is what it is called.
	Name string
	// Kind is what sort of thing it is.
	Kind SymbolKind
	// Line is the line it is on, from one.
	Line int
	// Detail is the signature or the type, as the caller formatted it. Empty
	// draws none, and it is the caller's because a signature is a language's
	// spelling and only the caller has parsed the language.
	Detail string
	// Children are what is inside it, for a type or a section.
	Children []Symbol
}

// SymbolOutlineOptions configure a SymbolOutline.
type SymbolOutlineOptions struct {
	// Name is what the panel is called — usually the file's name — and is
	// required.
	Name string
	// Symbols are what is in the file, and they are the caller's: which lines
	// are declarations is a question about a language's grammar, and the
	// highlighter in this package is a scanner and not a parser. A caller that
	// has a tree walk puts it here.
	Symbols []Symbol
	// Height is the viewport's own height, and is required.
	Height float32
	// Width is the panel's own width.
	Width float32
	// Selected is the symbol chosen, counted from zero.
	Selected int
	// State is where the panel is among its rows.
	State *ui.ListState
}

// SymbolOutlineResult carries a SymbolOutline and what was pressed in it.
type SymbolOutlineResult struct {
	// Element is the panel.
	Element *ui.Element
	// pressed is the symbol pressed this frame, as a path of indices from the
	// root, and nil for none — a path rather than an index, for the reason
	// ui/code's VariablesPanel says.
	pressed []int
	// shown is how many rows there are after the open blocks.
	shown int
}

// Pressed is the path to the symbol pressed this frame, and nil for none.
func (r SymbolOutlineResult) Pressed() []int { return r.pressed }

// Shown is how many rows there are.
func (r SymbolOutlineResult) Shown() int { return r.shown }

// SymbolOutline is what a file contains, as a list of its declarations.
//
// It is the one panel here whose content this package cannot produce, and
// that is deliberate: a tree walk is a parser's job, and this package has
// chosen a scanner everywhere else. A caller that wants an outline and does
// not have a parser of its own is better served by a scanner than by a
// highlighter that pretends to be one.
func SymbolOutline(c *ui.Context, opts SymbolOutlineOptions) SymbolOutlineResult {
	if opts.Name == "" {
		panic("code: SymbolOutline needs a Name; a list of declarations that says nothing about " +
			"which file they are in cannot be jumped to")
	}
	if opts.Height <= 0 {
		panic("code: SymbolOutline needs a Height; without one it draws every symbol")
	}

	flat := flattenSymbols(opts.Symbols)
	var r SymbolOutlineResult
	// The body is a function, not an element, so that CodePanel can
	// draw the header above it.
	view := func() {
		scroll := ui.Scroll(c).Height(opts.Height).FillWidth()
		scroll.Children(func() {
			ui.List(c, opts.State, len(flat), func(i int) {
				s := flat[i]
				row := symbolRow(c, s)
				row.Children(func() {
					if row.Clicked() {
						r.pressed = s.path
					}
				})
			}).FillWidth().Role(ui.RoleNone)
		})
		r.shown = len(flat)

	}
	host := CodePanel(c, view, layoutContainer(opts.Width), func() {
		CodeHeader(c, opts.Name, nil, func() {
			CodeBadge(c, itoa(len(flat))+" symbols", core.Neutral)
		})
	})

	r.Element = host
	return r
}

type flatSymbol struct {
	Symbol
	depth int
	path  []int
}

// flattenSymbols is a tree of symbols as the rows to draw, with every level
// shown: an outline is short — a file has tens of declarations, not thousands
// — so hiding the inside of a type would only make it harder to find the
// thing that is inside it.
func flattenSymbols(all []Symbol) []flatSymbol {
	var out []flatSymbol
	var walk func(syms []Symbol, depth int, path []int)
	walk = func(syms []Symbol, depth int, path []int) {
		for i := range syms {
			at := append(append([]int{}, path...), i)
			out = append(out, flatSymbol{Symbol: syms[i], depth: depth, path: at})
			walk(syms[i].Children, depth+1, at)
		}
	}
	walk(all, 0, nil)
	return out
}

func symbolRow(c *ui.Context, s flatSymbol) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	row := ui.Row(c).FillWidth().Shrink(0).AlignItems(ui.Center).
		Role(ui.RoleTreeItem).Label(s.Name)
	row.Children(func() {
		ui.Box(c).Width(float32(s.depth) * u * 2).Shrink(0)
		// The kind is a mark and not a word: an outline read for the shape of
		// it is read at a glance, and three words in a column of two hundred
		// names is a column of three hundred things.
		kind, ink := symbolFace(c, s.Kind, k)
		ui.Text(c, kind).Font(MonoStack).SingleLine().Shrink(0).
			FontSize(core.FontSize(c, theme.CaptionSize)).TextColor(ink).Width(u * 2)
		ui.Text(c, s.Name).Font(MonoStack).SingleLine().Shrink(0).
			FontSize(core.FontSize(c, theme.RowSize)).TextColor(k.Text)
		if s.Detail != "" {
			ui.Text(c, s.Detail).Font(MonoStack).SingleLine().Shrink(1).Grow(1).
				FontSize(core.FontSize(c, theme.CaptionSize)).TextColor(k.TextMuted)
		}
		ui.Box(c).Grow(1)
		CodeCaption(c, itoa(s.Line))
	})
	return row
}

// symbolFace is the two-letter mark and the colour each kind is drawn as, so
// that a function and a type can be told apart without reading their names.
func symbolFace(c *ui.Context, kind SymbolKind, k theme.Tokens) (string, ui.Color) {
	switch kind {
	case SymbolType:
		return "T ", TokenInk(c, Type)
	case SymbolValue:
		return "V ", TokenInk(c, Number)
	case SymbolSection:
		return "# ", k.TextMuted
	}
	return "F ", TokenInk(c, Function)
}

// ── Find ────────────────────────────────────────────────────────────────────

// FindOptions configure a FindWidget.
type FindOptions struct {
	// Label names the widget; it is required.
	Label string
	// Query is what to look for, and Found is how many places it was found.
	Query string
	// Found is how many matches, and At which one the widget is on, counted
	// from zero.
	Found, At int
	// CaseSensitive is a caller's switch rather than a modifier: a find bar
	// is not a text field with a Shift.
	CaseSensitive bool
	// Regex is that the query is a pattern rather than words. It is a bool
	// and not a compiled pattern because a caller matching in a different way
	// — a whole word, a glob — has the same needs and a compiled pattern here
	// would be a second engine.
	Regex bool
	// Replace is the other field, and is empty when the caller is only
	// looking. Empty does not draw the field, rather than drawing it empty:
	// a find bar with a replace box in it says the feature is there whether
	// or not it is wanted.
	Replace string
	// CanReplace says a replace field is wanted at all.
	CanReplace bool
	// Closeable puts a button that empties the query.
	Closeable bool
}

// FindResult carries a FindWidget and what was done in it.
type FindResult struct {
	// Element is the widget.
	Element *ui.Element
	// stepped is how far the highlight went this frame, signed, and zero for
	// anything else.
	stepped int
	// replaced is that the replace button was pressed this frame.
	replaced bool
	// closed is that the query was emptied this frame.
	closed bool
}

// Stepped is how far the highlight went this frame, and zero for none.
func (r FindResult) Stepped() int { return r.stepped }

// Replaced reports that the replace button was pressed this frame. The
// substitution is the caller's: this package has the query and the
// replacement and has no business deciding where the result goes.
func (r FindResult) Replaced() bool { return r.replaced }

// Closed reports that the query was emptied this frame.
func (r FindResult) Closed() bool { return r.closed }

// FindWidget is the bar above a viewer for finding something in the file.
//
// The counts are in the bar rather than in the caller because "match 4 of 212"
// is the answer and "the field's text" is only half of it: a find bar without
// the counts tells a person that something matched and leaves them to work out
// whether they have been through all of it.
func FindWidget(c *ui.Context, opts FindOptions) FindResult {
	if opts.Label == "" {
		panic("code: FindWidget needs a Label; a bar of numbers with no name cannot be found by " +
			"a screen reader or by a test")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	var r FindResult
	host := ui.Row(c).FillWidth().Shrink(0).Gap(u).AlignItems(ui.Center).
		Role(ui.RoleGroup).Label(opts.Label)
	host.Children(func() {
		in := ui.TextInputBase(c, &opts.Query).Label(opts.Label).
			Height(core.ControlHeight(c) - u).Radius(theme.SmallRadius).
			Background(k.Background).TextColor(k.Text)
		host.Children(func() {
			in.Width(u * 22).Shrink(0)
		})

		stepper := func(name string, by int) {
			btn := ui.ButtonBase(c).Size(u*6, u*6).Shrink(0).Radius(theme.SmallRadius).
				Background(k.Surface).TextColor(k.Text).Role(ui.RoleButton).
				Label(name).Tooltip(name)
			btn.Children(func() {
				ui.Text(c, name).SingleLine().FontSize(core.FontSize(c, theme.BodySize))
			})
			if btn.Clicked() {
				r.stepped = by
			}
		}
		host.Children(func() {
			stepper("↑", -1)
			stepper("↓", 1)
		})
		host.Children(func() {
			// Zero of zero is "no matches" and not "match 1 of 0": the bar is
			// empty of matches the moment a query has not been typed at all,
			// and saying "1/0" for that is a small lie a reader has to notice.
			say := itoa(opts.Found) + " matches"
			if opts.Query == "" {
				say = ""
			}
			CodeCaption(c, say)
		})
		if opts.CaseSensitive {
			host.Children(func() { CodeBadge(c, "Aa", core.Accent) })
		}
		if opts.Regex {
			host.Children(func() { CodeBadge(c, ".*", core.Accent) })
		}
		if opts.CanReplace {
			host.Children(func() {
				rep := ui.TextInputBase(c, &opts.Replace).
					Label(core.Msg(c, "code.replaceWith", core.Def("Replace with"))).
					Height(core.ControlHeight(c) - u).Radius(theme.SmallRadius).
					Background(k.Background).TextColor(k.Text)
				rep.Width(u * 14).Shrink(0)
			})
			host.Children(func() {
				btn := ui.ButtonBase(c).Height(core.ControlHeight(c)-u*2).Shrink(0).
					Radius(theme.PillRadius).Padding(0, u*2.5, 0, u*2.5).
					Background(k.Surface).TextColor(k.Text).Role(ui.RoleButton).
					Label(core.Msg(c, "code.replaceOne", core.Def("Replace"))).
					Tooltip(core.Msg(c, "code.replaceOne", core.Def("Replace")))
				if btn.Clicked() {
					r.replaced = true
				}
				btn.Children(func() {
					ui.Text(c, core.Msg(c, "code.replaceOne", core.Def("Replace"))).
						SingleLine().FontSize(core.FontSize(c, theme.RowSize))
				})
			})
		}
		if opts.Closeable {
			host.Children(func() {
				off := codeTabClose(c, core.Msg(c, "code.closeFind", core.Def("Close find")))
				if off.Clicked() {
					opts.Query = ""
					r.closed = true
				}
			})
		}
	})
	r.Element = host
	return r
}

// ── SearchPanel ─────────────────────────────────────────────────────────────

// SearchResult is one thing a search found.
type SearchResult struct {
	// Name is the file it is in and Line is the line, from one, and Column the
	// character on it, from one. All three are the caller's: a search over a
	// repository has its own file naming and its own line endings, and the
	// panel draws what it is given.
	Name     string
	Line     int
	Column   int
	Text     string
	Children []SearchResult
}

// SearchPanelOptions configure a SearchPanel.
type SearchPanelOptions struct {
	// Name is what the panel is called — the query — and is required.
	Name string
	// Query is what was searched for, and Results are what was found.
	Query   string
	Results []SearchResult
	// Height is the viewport's own height, and is required.
	Height float32
	// Width is the panel's own width.
	Width float32
	// State is where the panel is among its rows.
	State *ui.ListState
	// Selected is the result chosen, counted from zero among the shown rows.
	Selected int
	// Replaces is how many results would be changed by a replace, and zero
	// draws nothing: a count of zero is not a thing anybody needs to be told.
	Replaces int
}

// SearchPanelResult carries a SearchPanel and what was pressed in it.
type SearchPanelResult struct {
	// Element is the panel.
	Element *ui.Element
	// pressed is the result pressed this frame, and -1 for none.
	pressed int
	// shown is how many rows there are.
	shown int
}

// Pressed is the result pressed this frame, and -1 for none.
func (r SearchPanelResult) Pressed() int { return r.pressed }

// Shown is how many rows there are.
func (r SearchPanelResult) Shown() int { return r.shown }

// SearchPanel is what a search across files found, as a list of results with
// the lines they were found on.
//
// Every result is shown with its line, because a list of file names and counts
// is what a caller has and not what a person wants: the question a search
// asks is almost always "show me the line", and a panel that only offers the
// name makes the reader go and find it a second time.
func SearchPanel(c *ui.Context, opts SearchPanelOptions) SearchPanelResult {
	if opts.Name == "" {
		panic("code: SearchPanel needs a Name; results with nothing saying what was searched for " +
			"cannot be read out")
	}
	if opts.Height <= 0 {
		panic("code: SearchPanel needs a Height; without one it draws every result")
	}
	flat := flattenResults(opts.Results)
	var r SearchPanelResult
	r.pressed = -1

	// The body is a function, not an element, so that CodePanel can
	// draw the header above it.
	view := func() {
		scroll := ui.Scroll(c).Height(opts.Height).FillWidth()
		scroll.Children(func() {
			ui.List(c, opts.State, len(flat), func(i int) {
				res := flat[i]
				at := i
				row := searchResultRow(c, res, at == opts.Selected)
				row.Children(func() {
					if row.Clicked() {
						r.pressed = at
					}
				})
			}).FillWidth().Role(ui.RoleNone)
		})
		r.shown = len(flat)

	}
	host := CodePanel(c, view, layoutContainer(opts.Width), func() {
		CodeHeader(c, opts.Name, nil, func() {
			CodeBadge(c, itoa(len(flat))+" results", core.Neutral)
			if opts.Replaces > 0 {
				CodeBadge(c, itoa(opts.Replaces)+" to replace", core.Warning)
			}
		})
	})

	r.Element = host
	return r
}

type flatResult struct {
	SearchResult
	depth int
}

func flattenResults(all []SearchResult) []flatResult {
	var out []flatResult
	var walk func(rs []SearchResult, depth int)
	walk = func(rs []SearchResult, depth int) {
		for _, r := range rs {
			out = append(out, flatResult{SearchResult: r, depth: depth})
			walk(r.Children, depth+1)
		}
	}
	walk(all, 0)
	return out
}

func searchResultRow(c *ui.Context, res flatResult, chosen bool) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	// Two lines tall when there is a line to show and one when there is not,
	// stated rather than measured: a list that measures its rows lays the next
	// result's name on top of this one's preview.
	rows := float32(1)
	if res.Text != "" {
		rows = 2
	}
	row := ui.Column(c).FillWidth().Shrink(0).Gap(u * 0.25).AlignItems(ui.Start).
		Height(metricsOf(c, 1).line * rows).
		Role(ui.RoleTreeItem).Label(res.Name)
	if chosen {
		row.Background(k.Surface)
	}
	row.Children(func() {
		ui.Row(c).FillWidth().Shrink(0).AlignItems(ui.Center).Gap(u).Children(func() {
			ui.Box(c).Width(float32(res.depth) * u * 2).Shrink(0)
			ui.Text(c, res.Name).Font(MonoStack).SingleLine().Shrink(1).Grow(1).
				FontSize(core.FontSize(c, theme.RowSize)).TextColor(k.Text)
			CodeCaption(c, itoa(res.Line))
		})
		if res.Text != "" {
			// The line the match was found on, drawn with the query's own words
			// marked in it — through the same splitting the viewer's rows use,
			// so that a search hit looks the same here as it does there.
			marks := FindMatches(res.Text, strings.TrimSpace(res.Name))
			if len(marks) == 0 {
				marks = FindMatches(res.Text, res.QueryText())
			}
			ui.Box(c).Padding(0, u*2, 0, u*2).Radius(theme.SmallRadius).
				Background(k.Surface).Grow(1).FillWidth().Children(func() {
				codeTokens(c, []Token{{Text: res.Text, Kind: TokenPlain}},
					theme.CaptionSize, ui.Color{}, marks)
			})
		}
	})
	return row
}

// QueryText is the line's own text when there is nothing better to match
// against, which is the case where a caller has set the panel's Name to the
// query and a result's Name to a file. It exists so that a caller does not
// have to pass the query down to every result.
func (r SearchResult) QueryText() string { return r.Text }

// ── CompletionMenu ──────────────────────────────────────────────────────────

// Completion is one thing that could be typed next.
type Completion struct {
	// Label is the words shown, and is required.
	Label string
	// Detail is the kind and the signature — "func", "const string" — drawn at
	// the row's trailing edge. Empty draws none.
	Detail string
	// Kind is what sort of thing it is, which is what the mark in front is
	// drawn as.
	Kind SymbolKind
	// Insert is the text put in when it is taken. It is separate from Label
	// because a completion can be shown as what it will look like and
	// inserted as what it is: a snippet is labelled by its first line and
	// inserts a whole declaration, and a panel that inserted the label would
	// insert half a thing.
	Insert string
	// Selected is that it is the one the arrows are on.
	Selected bool
}

// CompletionMenuOptions configure a CompletionMenu.
type CompletionMenuOptions struct {
	// Label names the menu; it is required, for the reason every other name
	// here is.
	Label string
	// Completions are what could be typed, in the order they should be read:
	// best first, which is the order the caller's ranking has them in.
	Completions []Completion
	// Width is the menu's own width; zero is as wide as its widest row.
	Width float32
	// MaxHeight is how tall the menu may be before it scrolls. Zero is ten
	// rows, which is more than fits under a caret and few enough that the menu
	// does not cover the line being typed.
	MaxHeight float32
	// Prefix is what has been typed, and is drawn at the menu's head so that
	// a reader knows what the completions complete. It is optional because a
	// caller whose field shows it already has nothing to add by showing it
	// twice.
	Prefix string
	// Highlight is the completion the arrows are on, counted from zero.
	Highlight int
}

// CompletionMenuResult carries a CompletionMenu and what was taken from it.
type CompletionMenuResult struct {
	// Element is the menu.
	Element *ui.Element
	// taken is the index chosen this frame, or -1.
	taken int
}

// Taken is the completion chosen this frame, counted from zero, and -1 for
// none. It is an index rather than the text because the caller has the list
// and the list may be reordered between the press and the frame it is read in.
func (r CompletionMenuResult) Taken() int { return r.taken }

// CompletionMenu is the list of what could be typed next, under the caret.
//
// It is a panel and not a floating layer, because a completion menu belongs to
// the field it is completing: hung off the field it moves with the field, and
// hung off the window it does not. The caller places it; this package draws
// it.
func CompletionMenu(c *ui.Context, opts CompletionMenuOptions) CompletionMenuResult {
	if opts.Label == "" {
		panic("code: CompletionMenu needs a Label; a list of words that says nothing about what " +
			"they complete cannot be read out")
	}
	if len(opts.Completions) == 0 {
		panic("code: CompletionMenu needs at least one completion; an empty menu is a gap over " +
			"the thing being typed")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	tall := opts.MaxHeight
	if tall <= 0 {
		tall = LineHeight(c) * 10
	}

	var r CompletionMenuResult
	r.taken = -1
	host := ui.Column(c).FillWidth().Shrink(0).Gap(0).Role(ui.RoleList).
		Label(opts.Label)
	if opts.Width > 0 {
		host.Width(opts.Width).Shrink(0)
	}
	host.Children(func() {
		if opts.Prefix != "" {
			ui.Row(c).FillWidth().Shrink(0).Padding(u*0.5, u*1.5).
				Children(func() {
					codeMono(c, opts.Prefix, theme.CaptionSize, k.TextMuted)
				})
		}
		list := ui.Scroll(c).MaxHeight(tall).Gap(0)
		list.Children(func() {
			for i, comp := range opts.Completions {
				at, comp := i, comp
				if comp.Label == "" {
					panic("code: CompletionMenu item " + itoa(at) + " has no Label; a row of " +
						"nothing cannot be chosen from")
				}
				list.Children(func() {
					row := completionRow(c, comp, at == opts.Highlight)
					if row.Clicked() {
						r.taken = at
					}
				})
			}
		})
	})
	// The panel is the same shell every other layer in this library is, so
	// that a menu over a field and a panel in a sheet are one shape.
	r.Element = panelFace(c, host)
	return r
}

func completionRow(c *ui.Context, comp Completion, highlighted bool) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	row := ui.Row(c).FillWidth().Shrink(0).AlignItems(ui.Center).Gap(u).
		Role(ui.RoleListItem).Label(comp.Label)
	switch {
	case highlighted:
		// The highlight is the pressed grey rather than the accent: a menu
		// that opens over a field which may itself be accented would have
		// two of the same colour within a hand's width.
		row.Background(k.SurfacePressed)
	case row.Hovered():
		row.Background(k.SurfaceHover)
	}
	row.Children(func() {
		kind, ink := symbolFace(c, comp.Kind, k)
		ui.Text(c, kind).Font(MonoStack).SingleLine().Shrink(0).
			FontSize(core.FontSize(c, theme.CaptionSize)).TextColor(ink).Width(u * 2)
		ui.Text(c, comp.Label).Font(MonoStack).SingleLine().Shrink(1).Grow(1).
			FontSize(core.FontSize(c, theme.RowSize)).TextColor(k.Text)
		if comp.Detail != "" {
			ui.Text(c, comp.Detail).Font(MonoStack).SingleLine().Shrink(0).
				FontSize(core.FontSize(c, theme.CaptionSize)).TextColor(k.TextFaint)
		}
	})
	return row
}

// ── Flamegraph ──────────────────────────────────────────────────────────────

// FlameFrame is one span of a call stack's cost: what it was, and what it took.
//
// Widths and heights are the caller's own numbers. A flame graph is a picture
// of a profile, and the profile is the caller's: it knows whether a number is
// microseconds, samples or bytes, and a panel that guessed would draw a chart
// of the wrong thing beautifully.
type FlameFrame struct {
	// Name is what the span was called.
	Name string
	// Self is how much of the time was in this span and not in anything below
	// it, and Total is how much including what is below.
	Self, Total float64
	// File and Line are where it is, and are optional.
	File string
	Line int
	// Children are the spans it made, which are drawn inside it.
	Children []FlameFrame
}

// FlamegraphOptions configure a Flamegraph.
type FlamegraphOptions struct {
	// Name is what the graph is called, and is required.
	Name string
	// Frame is the root span.
	Frame FlameFrame
	// Width is the graph's own width, and is required: a flame graph's whole
	// meaning is in its proportions, and without a width the proportions are
	// nothing.
	Width float32
	// Height is how tall one level is; zero is a standard control at this
	// window's density.
	Height float32
	// Value is the total the widths are fractions of, which is the root's own
	// total when it is set to zero.
	Value float64
	// Highlight is the span the reader is looking at, by name, and empty for
	// none. It is a name and not a path because a flame graph is read by
	// what a frame was called and not by where in the tree it sits.
	Highlight string
}

// FlamegraphResult carries a Flamegraph and what was pressed in it.
type FlamegraphResult struct {
	// Element is the graph.
	Element *ui.Element
	// pressed is the name of the span pressed this frame, and "" for none.
	pressed string
}

// Pressed is the name of the span pressed this frame, and "" for none. It is a
// name rather than a number because a span's place in the tree changes as the
// profile is re-read, and a caller that wanted to act on one wants the frame.
func (r FlamegraphResult) Pressed() string { return r.pressed }

// Flamegraph is where a program's time went: a tree of spans, each drawn as
// wide as the share of the work it was.
//
// It is drawn as one paint rather than as a box per span, because a profile of
// a server process has tens of thousands of spans and a tree of boxes would
// be a tree of elements — where the answer is a few thousand rectangles on one
// surface, which is the same picture for a fraction of the cost.
func Flamegraph(c *ui.Context, opts FlamegraphOptions) FlamegraphResult {
	if opts.Name == "" {
		panic("code: Flamegraph needs a Name; a chart of spans that says nothing about which " +
			"program they were cannot be read out")
	}
	if opts.Width <= 0 {
		panic("code: Flamegraph needs a Width; a flame graph's meaning is in its proportions, and " +
			"without one they are nothing")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	h := opts.Height
	if h <= 0 {
		h = core.ControlHeight(c) - u
	}
	total := opts.Value
	if total <= 0 {
		total = opts.Frame.Total
	}
	if total <= 0 {
		total = 1
	}

	depth := flameDepth(opts.Frame)
	var r FlamegraphResult
	// The body is a function, not an element, so that CodePanel can
	// draw the header above it.
	graph := func() {
		// Its own height rather than FillHeight: the panel it is drawn in is
		// as tall as what is in it, so a fill resolves to nothing and the
		// whole graph — tens of thousands of rectangles — is painted into a
		// box one pixel tall.
		graph := ui.Box(c).FillWidth().Height(float32(depth)*h + u).Shrink(0).
			Background(k.Background).Role(ui.RoleImage).Label(opts.Name)
		graph.Children(func() {
			graph.Draw(func(p *ui.Painter, box ui.Rect) {
				// The top of the graph is the bottom of the stack, because that is
				// where a flame graph starts: the whole call at the top, and the
				// spans it made underneath it.
				var walk func(f FlameFrame, x, y, w float32)
				walk = func(f FlameFrame, x, y, w float32) {
					if w <= 0 || y < 0 {
						return
					}
					hot := f.Self > f.Total/2
					fill := k.Surface
					ink := k.Text
					switch {
					case opts.Highlight != "" && f.Name == opts.Highlight:
						fill, ink = k.Accent, k.AccentText
					case hot:
						// The frames that spent most of their own time here are
						// the ones a profile is read for, so they are the ones
						// drawn in the window's own text colour and everything
						// else is drawn quieter. It is the same reading as a heat
						// map and it costs no colour scale.
						fill, ink = k.SurfaceHover, k.Text
					}
					p.Fill(ui.Rect{X: x, Y: y, W: w, H: h - 1}, fill, 0)
					p.Fill(ui.Rect{X: x, Y: y + h - 1, W: w, H: 1}, k.Border, 0)
					if w > u*4 {
						// The name is centred in its own frame rather than
						// placed at seventy-two percent of its height:
						// Painter.Text takes the top-left corner, so a fixed
						// fraction puts the name below the bottom of every
						// frame but the last, and each frame paints over the
						// name of the one above it.
						p.Text(x+u*0.75, y+(h-1-theme.CaptionSize*1.2)/2,
							clipTo(f.Name, int(w/u)-1), theme.CaptionSize, ink)
					}
					// The children share the parent's width in the proportion of
					// their own totals, left to right, which is what makes a wide
					// frame with a wide child under it obviously so.
					at := x
					for _, child := range f.Children {
						cw := w * float32(child.Total/maxFloat64(f.Total, 1e-9))
						if cw <= 0 {
							continue
						}
						walk(child, at, y+h, cw)
						at += cw
					}
				}
				walk(opts.Frame, box.X, box.Y, box.W)
			})
			if graph.Clicked() {
				r.pressed = opts.Frame.Name
			}
		})

	}
	host := CodePanel(c, graph, layoutContainer(opts.Width), func() {
		CodeHeader(c, opts.Name, []string{itoa(depth) + " levels"}, nil)
	})

	r.Element = host
	return r
}

// flameDepth is how many levels deep the widest stack is, which is how tall
// the graph has to be.
func flameDepth(f FlameFrame) int {
	deep := 1
	for _, c := range f.Children {
		if d := flameDepth(c); d >= deep {
			deep = d + 1
		}
	}
	return deep
}

func maxFloat64(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// clipTo is a name as it fits in a given number of characters, which is what
// a flame graph draws: a truncated name with a tilde is what every flame graph
// on the desktop shows, and a full name that runs past its own frame is a name
// drawn over its neighbour.
func clipTo(s string, n int) string {
	r := []rune(s)
	if n <= 1 || len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "~"
}

// panelFace is the shell a floating layer in this package wears, and it is
// ui/layout's Panel rather than a second one for the same reason ui/input has
// panelFace and not its own: a package that answered "what does a layer look
// like" twice would have two layers.
func panelFace(c *ui.Context, host *ui.Element) *ui.Element {
	return layout.Panel(c, host, layout.PanelOptions{Compact: true, Pad: u1(c)}, nil)
}

func u1(c *ui.Context) float32 { return core.Density(c).Unit() }
