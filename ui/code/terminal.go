package code

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// The terminal, and the two things that hang off it: a search across it and a
// row of tabs over it.

// TerminalOptions configure a Terminal.
type TerminalOptions struct {
	// Name is what the terminal is called — a shell, a program, a machine. It
	// is required, for the same reason every other name here is.
	Name string
	// Lines are the lines shown, newest last.
	Lines []string
	// Height is the viewport's own height, and is required.
	Height float32
	// Width is the terminal's own width.
	Width float32
	// Scroll is where the terminal is scrolled to, and it is the caller's so
	// that a viewer's place survives a redraw. A terminal that keeps its own
	// scroll position is a terminal that forgets where it was every time the
	// window repainted.
	Scroll *ui.ScrollState
	// Prompt is the line shown after the output, with the caret at the end.
	// Empty draws none, for a terminal showing output rather than waiting for
	// something.
	Prompt string
	// PromptName is what the prompt belongs to — a shell, a process — and is
	// drawn before it.
	PromptName string
	// Follow keeps the end in view as lines arrive, unless the person has
	// scrolled away from it. It is the behaviour a chat has and a build log
	// wants and a log view deliberately does not have, which is why it is
	// asked for here rather than assumed.
	Follow bool
	// Level is the severity of a line, when the caller is drawing a build's
	// output rather than a shell's. The zero value means nothing is coloured.
	Level func(i int) LogLevel
}

// TerminalResult carries a Terminal and what was done in it.
type TerminalResult struct {
	// Element is the terminal.
	Element *ui.Element
	// atEnd is whether the viewport is at its end as of this frame.
	atEnd bool
	// submitted is that Enter was pressed at the prompt.
	submitted bool
}

// AtEnd reports whether the terminal is scrolled to its end, which is what a
// caller checking a follow toggle needs.
func (r TerminalResult) AtEnd() bool { return r.atEnd }

// Submitted reports that Enter was pressed at the prompt. It is read inside
// the view, because by the last pass of the frame the press was read in there
// is nothing pending to report — which is the rule this package's tests
// already follow and the reason every result here is read in the closure.
func (r TerminalResult) Submitted() bool { return r.submitted }

// Terminal is a program's output with a prompt at the end of it.
//
// It scrolls in ui/layout's scroll area rather than a scroll of its own, so
// that a terminal next to a viewer scrolls with the same thumb, the same
// overscroll and the same overshoot as everything else long in this library.
// Two scrollers that feel different side by side are worse than one scroller
// used twice.
func Terminal(c *ui.Context, opts TerminalOptions) TerminalResult {
	if opts.Name == "" {
		panic("code: Terminal needs a Name; output with nothing saying what produced it cannot " +
			"be read out or found")
	}
	if opts.Height <= 0 {
		panic("code: Terminal needs a Height; without one it draws every line of output")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	m := metricsOf(c, len(opts.Lines)+1)

	var r TerminalResult
	// Following means going to the end, and MaxFloat32 is how a caller says
	// "the end" to a scroll container: it clamps it to whatever the end turns
	// out to be this frame. Going to a remembered offset instead would put
	// the end one line short every time a line arrived, because the end was
	// the end of the frame before.
	if opts.Scroll != nil && opts.Follow {
		opts.Scroll.Y = maxFloat(opts.Scroll.Y, opts.Scroll.MaxY)
	}

	// The body is a function, not an element, so that CodePanel can
	// draw the header above it.
	scroll := func() {
		host := layout.ScrollArea(c, layout.ScrollAreaOptions{
			Vertical: true, Horizontal: true, Height: opts.Height, State: opts.Scroll,
		}, nil).Element.FillWidth()
		host.Children(func() {
			col := ui.Column(c).FillWidth().Gap(0)
			col.Children(func() {
				for i, line := range opts.Lines {
					text := line
					level := LogTrace
					if opts.Level != nil {
						level = opts.Level(i)
					}
					_, fg := level.Severity().Pair(k)
					col.Children(func() {
						row := ui.Row(c).FillWidth().Height(m.line).AlignItems(ui.Center).
							Grow(0).Shrink(0).Label(terminalLineLabel(opts.Name, i))
						row.Children(func() {
							ansiRow(c, text, fg, ui.Color{})
						})
					})
				}
				if opts.Prompt != "" {
					col.Children(func() {
						prompt := ui.Row(c).FillWidth().Height(m.line).AlignItems(ui.Center).
							Grow(0).Shrink(0).Focusable().
							Label(core.Msg(c, "code.prompt", core.Def("Prompt")))
						prompt.Children(func() {
							if opts.PromptName != "" {
								CodeBadge(c, opts.PromptName, core.Accent)
								ui.Box(c).Width(u).Shrink(0)
							}
							codeMono(c, opts.Prompt, theme.RowSize, k.Text)
						})
						// Enter at the prompt is the one thing a terminal has to
						// answer for itself, because everything else about it is
						// the caller's program doing the answering.
						if prompt.Shortcut(0, ui.KeyEnter) {
							r.submitted = true
						}
					})
				}
			})
		})
		if opts.Scroll != nil {
			r.atEnd = opts.Scroll.Y >= opts.Scroll.MaxY
		}
		host.Role(ui.RoleNone)

	}
	host := CodePanel(c, scroll, layoutContainer(opts.Width), func() {
		CodeHeader(c, opts.Name, nil, nil)
	})

	r.Element = host
	return r
}

func terminalLineLabel(name string, i int) string {
	return name + " line " + itoa(i+1)
}

func maxFloat(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

// TerminalSearchOptions configure a TerminalSearch.
type TerminalSearchOptions struct {
	// Label names the field; it is required, as for every field in this
	// library.
	Label string
	// Query is what to look for, and Found is how many places it was found.
	Query string
	// Found is how many matches there are, and At which one the field is on,
	// counted from zero.
	Found, At int
	// Next and Prev go one match forwards or back, and are reported through
	// TerminalSearchResult.
	Next, Prev bool
	// CaseSensitive is a caller's switch rather than a modifier: a terminal
	// is not a text field with a Shift.
	CaseSensitive bool
}

// TerminalSearchResult carries a TerminalSearch and where it moved.
type TerminalSearchResult struct {
	// Element is the field.
	Element *ui.Element
	// moved is how far the highlight went this frame.
	moved int
}

// Moved is how many matches the highlight went this frame: one for a press of
// either arrow, zero for anything else, and it is signed — back is negative —
// because a caller wrapping its own list needs to know which way it went, not
// only that it went.
func (r TerminalSearchResult) Moved() int { return r.moved }

// TerminalSearch is a field for finding something in a terminal's output.
//
// It is one field and one pair of arrows rather than a strip of options,
// because a terminal is unbounded and has no list to offer: the answer to a
// search in one is always "match 4 of 212", and the two numbers are what the
// field is for.
func TerminalSearch(c *ui.Context, opts TerminalSearchOptions) TerminalSearchResult {
	if opts.Label == "" {
		panic("code: TerminalSearch needs a Label; a box of numbers with no name cannot be " +
			"found by a screen reader or by a test")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	var r TerminalSearchResult
	host := ui.Row(c).Shrink(0).Gap(u).AlignItems(ui.Center).Label(opts.Label)
	host.Children(func() {
		// The field is a bare editor rather than a decorated search field: a
		// terminal's search sits in a strip of its own with the arrows beside
		// it, and a magnifier inside a strip that already says what it is is
		// a mark that says it twice.
		in := ui.TextInputBase(c, &opts.Query).Label(opts.Label).
			Height(core.ControlHeight(c) - u).Radius(theme.SmallRadius).
			Font(MonoStack).FontSize(core.FontSize(c, theme.RowSize)).
			Background(k.Background).TextColor(k.Text)
		host.Children(func() {
			in.Width(u * 18).Shrink(0)
			stepper := func(name string, by int) {
				btn := ui.ButtonBase(c).Size(u*6, u*6).Shrink(0).Radius(theme.SmallRadius).
					Background(k.Surface).TextColor(k.Text).Role(ui.RoleButton).
					Label(name).Tooltip(name)
				btn.Children(func() {
					ui.Text(c, name).SingleLine().FontSize(core.FontSize(c, theme.BodySize))
				})
				if btn.Clicked() {
					r.moved = by
				}
			}
			stepper("↑", -1)
			stepper("↓", 1)
			ui.Text(c, itoa(opts.At+1)+"/"+itoa(opts.Found)).SingleLine().Shrink(0).
				FontSize(core.FontSize(c, theme.CaptionSize)).TextColor(k.TextMuted)
		})
	})
	r.Element = host
	return r
}

// TerminalTabsOptions configure a TerminalTabs.
type TerminalTabsOptions struct {
	// Label names the row; it is required.
	Label string
	// Tabs are the terminals, by name.
	Tabs []string
	// Current is which one is showing, counted from zero.
	Current int
	// Closeable puts a button on each tab, reported through the result.
	Closeable bool
}

// TerminalTabsResult carries a TerminalTabs and what was pressed in it.
type TerminalTabsResult struct {
	// Element is the row.
	Element *ui.Element
	// chosen is the tab pressed this frame, counted from zero, and -1 for
	// none.
	chosen int
	// closed is the tab closed this frame, counted from zero, and -1 for
	// none. It is separate from chosen because closing the tab a person is
	// looking at is a different thing from switching to it, and a caller
	// holding only one number would have to guess which was meant.
	closed int
}

// Chosen is the tab pressed this frame, and -1 for none.
func (r TerminalTabsResult) Chosen() int { return r.chosen }

// Closed is the tab closed this frame, and -1 for none.
func (r TerminalTabsResult) Closed() int { return r.closed }

// TerminalTabs is a row of terminals' names, with the one showing marked.
//
// The chosen tab is raised off the row rather than given a colour of its own,
// which is the same rule ui/input's ToggleGroup follows and for the same
// reason: a row of tabs and a row of pills are two different things, and
// making them look alike is how a reader ends up looking for a pill.
func TerminalTabs(c *ui.Context, opts TerminalTabsOptions) TerminalTabsResult {
	if opts.Label == "" {
		panic("code: TerminalTabs needs a Label; a row of names with nothing saying what they " +
			"are cannot be read out")
	}
	if len(opts.Tabs) == 0 {
		panic("code: TerminalTabs needs at least one tab")
	}
	if opts.Current < 0 || opts.Current >= len(opts.Tabs) {
		panic("code: TerminalTabs Current is out of range")
	}
	u := core.Density(c).Unit()

	var r TerminalTabsResult
	r.chosen, r.closed = -1, -1
	row := ui.Row(c).FillWidth().Shrink(0).Gap(u * 0.5).Role(ui.RoleTabList).
		Label(opts.Label)
	row.Children(func() {
		for i := range opts.Tabs {
			at, name := i, opts.Tabs[i]
			// One closure per tab rather than a loop of elements: an element
			// made outside its parent's Children call is that parent's
			// sibling, so a row built by looping over pre-made tabs would show
			// them stacked above the row rather than inside it.
			row.Children(func() {
				tab := codeTab(c, name, at == opts.Current)
				if tab.Clicked() {
					r.chosen = at
				}
				if !opts.Closeable {
					return
				}
				row.Children(func() {
					off := codeTabClose(c, name)
					if off.Clicked() {
						r.closed = at
					}
				})
			})
		}
	})
	r.Element = row
	return r
}

// codeTab is one terminal's name, raised off the row when it is the one
// showing.
//
// Raised rather than coloured, and for the same reason ui/input's chosen
// segment is: a row of tabs and a row of pills are two different things, and
// making them look alike is how a reader ends up looking for a pill in a
// place there are none.
func codeTab(c *ui.Context, name string, chosen bool) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	tab := ui.ButtonBase(c).Shrink(0).Radius(theme.SmallRadius).
		Padding(u*0.75, u*2).Role(ui.RoleTab).Checked(chosen).
		Label(name).Tooltip(name)
	ink := k.TextMuted
	if chosen {
		ink = k.Text
		tab.Background(k.Surface).TextColor(k.Text)
		tab.Shadow(0, 1, 2, 0, ui.RGBA(0, 0, 0, 0.12))
	} else if tab.Hovered() {
		tab.Background(k.SurfaceHover)
	}
	tab.Children(func() {
		ui.Text(c, name).SingleLine().FontSize(core.FontSize(c, theme.RowSize)).
			TextColor(ink)
	})
	return tab
}

// codeTabClose is the cross on a tab's close button. It is a button and not a
// glyph inside the tab: a press inside the tab that opens the tab and a press
// on the cross that closes it are two different acts, and a control where the
// same three pixels do both is a control nobody trusts.
func codeTabClose(c *ui.Context, name string) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	close := core.Msg(c, "code.closeTab", core.Def("Close this tab"))
	name = name + " " + close
	off := ui.Box(c).Size(u*5.5, u*5.5).Shrink(0).Radius(u * 2.75).
		Role(ui.RoleButton).Label(name).Tooltip(name).Cursor(ui.CursorPointer)
	off.Children(func() {
		ui.Icon(c, closeGlyph).TextColor(k.TextFaint).Size(u*3, u*3)
	})
	return off
}

// closeGlyph is the cross on a tab's close button, drawn on the same grid as
// ui/input's glyphs.
var closeGlyph = mustGlyph(`<path d="M6.5 6.5l11 11M17.5 6.5l-11 11"/>`)

// mustGlyph is ui/input's glyph parser, reached rather than copied: a second
// parser of the same XML in the same window is a second set of bugs, and the
// only thing this package needs of it is the shape of a cross.
func mustGlyph(shapes string) *ui.SVG {
	return ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" ` +
		`fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" ` +
		`stroke-linejoin="round">` + shapes + `</svg>`))
}
