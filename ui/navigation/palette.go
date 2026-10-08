package navigation

import (
	"sort"
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/overlay"
	"github.com/HycJack/MintUI/ui/theme"
)

// Command is one thing the palette can do.
type Command struct {
	// ID is what the palette hands back when the command is chosen. It is
	// the caller's own key — a route, an action name, a callback id — and
	// not the title: a caller that switches on the title breaks the moment
	// somebody rewords it for another language.
	ID string
	// Title is what the row shows, and the first thing the query is matched
	// against.
	Title string
	// Subtitle is the second line: where the command goes, or what it does
	// to the thing in front of you.
	Subtitle string
	// Keywords are extra words that should find the command when its title
	// does not: "roster" for the team page, "escalate" for the button that
	// raises a callback. They are never drawn.
	Keywords []string
	// Shortcut is drawn at the right of the row, one key per string, in the
	// same boxes a shortcut wears everywhere else — so a command reads the
	// same here and on the toolbar that also offers it.
	Shortcut []string
	// Disabled greys the row and stops it being chosen. It is not hidden:
	// a palette that dropped its disabled commands would renumber
	// everything below them as you typed, and the highlight would land on a
	// different command each time a letter came out.
	Disabled bool
}

// matchGrade says how well a command answers a query. Lower sorts first.
type matchGrade int

// The grades run from the strongest answer to the weakest, and gradeNone is
// the worst of all — which is what lets a keyword take a command's grade away
// from a title that did not match at all, and lets the comparator below be a
// plain less-than.
const (
	// gradePrefix is a hit at the very start of the title. It is the answer
	// nearly every keystroke is looking for, so it sits above everything.
	gradePrefix matchGrade = iota
	// gradeWord is a hit at the start of a word inside the text, which is
	// what a person means when they type the first part of a word rather
	// than the whole of it: "call" finding "Open callbacks".
	gradeWord
	// gradeSubstring is a hit somewhere inside a word: "back" finds
	// "Callbacks", which is a guess — and a palette that lists its guesses
	// before its answers is one people stop reading.
	gradeSubstring
	// gradeNone is what a command that does not match gets. It never reaches
	// the sort: Match drops those before comparing anything.
	gradeNone
)

// gradeWhere says which text matched. It is a weaker signal than how the text
// matched and a stronger one than where the command happened to sit in the
// list the caller gave.
type gradeWhere int

const (
	// whereTitle outranks whereKeyword: the word a person typed at the front
	// of the thing they are looking for is worth more than a word somebody
	// filed in the back of it, so the title's hit is the one that sorts
	// first when both are the same kind of hit.
	whereTitle gradeWhere = iota
	whereKeyword
)

// Match filters and orders commands for a query.
//
// It is a pure function of its arguments — no context, no window, no frame —
// because the order is the part worth being sure about. A palette that sorts
// "Callbacks" below "Team" when you type "cal" is not a palette with the
// wrong order, it is a palette people stop trusting, and the only way to hold
// that is to assert the order directly rather than to count what came back.
//
// The rules, strongest first:
//
//   - a prefix of the title beats a hit at the start of any word, which beats
//     a hit somewhere inside a word;
//   - a hit in the title beats the same kind of hit in a keyword, so a
//     keyword never outranks the title it was attached to;
//   - a disabled command matches and sorts like any other — see Command.
//     Disabled;
//   - an empty query matches everything, in the order given.
//
// Ties keep the order they came in, so a caller expresses a preference by
// listing its commands in one.
func Match(query string, items []Command) []Command {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return append([]Command(nil), items...)
	}

	type scored struct {
		cmd     Command
		grade   matchGrade
		where   gradeWhere
		order   int
		matched bool
	}
	out := make([]scored, len(items))
	for i, cmd := range items {
		g, w := gradeCommand(cmd, q)
		out[i] = scored{cmd: cmd, grade: g, where: w, order: i, matched: g != gradeNone}
	}
	// Stable, so two commands of one grade keep the caller's order between
	// them rather than being shuffled by the sort's own idea of ties.
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].matched != out[j].matched {
			return out[i].matched
		}
		if out[i].grade != out[j].grade {
			return out[i].grade < out[j].grade
		}
		if out[i].where != out[j].where {
			return out[i].where < out[j].where
		}
		return out[i].order < out[j].order
	})

	res := make([]Command, 0, len(out))
	for _, s := range out {
		if s.matched {
			res = append(res, s.cmd)
		}
	}
	return res
}

// gradeCommand scores one command against an already-lowercased query: the
// best kind of hit anywhere the command can be found by, with the title
// breaking a tie against its own keywords.
func gradeCommand(cmd Command, q string) (matchGrade, gradeWhere) {
	best, where := gradeCommandText(cmd.Title, q, whereTitle)
	for _, kw := range cmd.Keywords {
		if g, _ := gradeCommandText(kw, q, whereKeyword); g < best {
			best, where = g, whereKeyword
		}
	}
	return best, where
}

// gradeCommandText scores one string, where saying which text it was.
func gradeCommandText(s, q string, where gradeWhere) (matchGrade, gradeWhere) {
	i := strings.Index(strings.ToLower(s), q)
	switch {
	case i < 0:
		return gradeNone, where
	case i == 0:
		return gradePrefix, where
	case isWordStart(strings.ToLower(s), i):
		return gradeWord, where
	default:
		return gradeSubstring, where
	}
}

// isWordStart reports whether the match at i begins a word. Everything that
// is not a letter or a digit is a boundary, so "CB-2871" is found by "2871"
// and "CB-2871" both findable, and a camel-cased ID is found by its parts.
func isWordStart(s string, i int) bool {
	if i == 0 {
		return true
	}
	return !isWordByte(s[i-1])
}

func isWordByte(b byte) bool {
	switch {
	case b >= 'a' && b <= 'z', b >= 'A' && b <= 'Z', b >= '0' && b <= '9':
		return true
	}
	return false
}

// CommandPaletteOptions configure a CommandPalette.
type CommandPaletteOptions struct {
	// NonModal leaves the window behind undimmed. It is off by default,
	// because a palette is something you asked for on purpose and the scrim
	// says everything else has stopped mattering — which is true while you
	// are typing and false the moment you glance at the board to check what
	// you meant to type. Escape closes it either way.
	//
	// A gallery page wants this: it has to show the palette and what is
	// around it in one frame.
	NonModal bool
	// Query is the caller's typed text. The palette writes into it as the
	// person types and reads it to decide what matches, so the same string
	// can be shown elsewhere — in the search field that opened it, in a "no
	// results for …" line — without a round trip through a widget.
	Query *string
	// Highlight is the caller's index into the drawn rows: where the
	// highlight sits, and which row Enter runs.
	//
	// It is the caller's because the highlight is app state like any other:
	// a window that reopens the palette on the query it had wants the same
	// row lit, and a palette that remembered it itself would need telling
	// to forget.
	Highlight *int
	// Width is the panel's width; zero lets it fit its content.
	Width float32
	// Rows caps how many commands are drawn; zero gives eight.
	Rows int
	// Placeholder is the field's prompt; empty uses the library's.
	Placeholder string
	// Label names the palette for assistive technology; empty uses the
	// library's. It is the name and nothing else — a palette whose only
	// heading is drawn is a panel nobody can find with a screen reader.
	Label string
	// Empty is what the list says when nothing matched; empty uses the
	// library's, because a panel of nothing at all looks like a bug.
	Empty string
}

// CommandPaletteResult carries a CommandPalette and what it ran.
type CommandPaletteResult struct {
	// Element is the panel, or nil while the palette is closed.
	Element *ui.Element
	// chosen is the ID of the command that ran this frame.
	chosen string
	// ran separates "the command whose ID is the empty string ran" from
	// "nothing ran" — the same pair the library uses everywhere else,
	// because a zero value that is also a real answer needs a flag.
	ran bool
	// rows are the commands drawn this frame, in the order Match returned
	// them, capped at the option. Highlight indexes this slice.
	rows []Command
	// total is how many commands matched before the cap, which is what a
	// caller says "3 of 28" from.
	total int
	// highlight is where the highlight sits, for the frame it was settled.
	highlight int
}

// Chosen returns the ID of the command that ran this frame, and whether one
// ran at all.
func (r CommandPaletteResult) Chosen() (string, bool) { return r.chosen, r.ran }

// Highlight returns the index of the highlighted row among the rows drawn,
// and how many were drawn.
func (r CommandPaletteResult) Highlight() (index, rows int) { return r.highlight, len(r.rows) }

// Results returns the commands drawn this frame, in order.
func (r CommandPaletteResult) Results() []Command { return r.rows }

// Total returns how many commands matched, before the rows cap.
func (r CommandPaletteResult) Total() int { return r.total }

// CommandPalette is a field over the whole application's commands: type a few
// letters, move through what is left with the arrows, press Enter.
//
// The layer is ui/overlay's Dialog and the panel is its Panel, so a palette
// wears the same radius, hairline and shadow as every other layer and is
// dismissed by the same press on the scrim and the same Escape. What the
// palette adds is the list, the keys and Match.
//
// Keys:
//
//	Down   the next row, wrapping from the last back to the first
//	Up     the previous row, wrapping the other way
//	Enter  runs the highlighted row
//	Esc    closes, as the layer's own Escape does
//
// The arrows are read with OverlayShortcut on the panel rather than Shortcut
// on the field, so they arrive wherever the focus is — a person who clicked a
// row has moved the focus off the field, and the arrows have to keep working.
// That has one price, and it is paid on purpose: MyGo gives a scrolling
// container the arrows before it gives them to an overlay, so the list is
// capped at Rows rows rather than scrolled. A palette whose list scrolled
// would stop moving the highlight the moment the list overflowed, which is
// exactly when moving it matters most.
func CommandPalette(c *ui.Context, open *bool, items []Command, opts CommandPaletteOptions) CommandPaletteResult {
	if open == nil {
		panic("navigation: CommandPalette needs the *bool it opens and closes")
	}
	if opts.Query == nil {
		panic("navigation: CommandPalette needs a *string for what has been typed")
	}
	if opts.Highlight == nil {
		panic("navigation: CommandPalette needs a *int for the highlighted row")
	}
	if !*open {
		return CommandPaletteResult{highlight: -1}
	}

	rows := opts.Rows
	if rows <= 0 {
		rows = 8
	}
	label := opts.Label
	if label == "" {
		label = core.Msg(c, "palette.title", "Command palette")
	}
	placeholder := opts.Placeholder
	if placeholder == "" {
		placeholder = core.Msg(c, "palette.placeholder", "Type a command")
	}
	empty := opts.Empty
	if empty == "" {
		empty = core.Msg(c, "palette.empty", "No matching commands")
	}

	res := CommandPaletteResult{highlight: -1}
	matched := Match(*opts.Query, items)
	res.total = len(matched)
	if len(matched) > rows {
		matched = matched[:rows]
	}
	res.rows = matched

	// Only a row that is drawn can be highlighted. Left alone, a highlight
	// on a row the cap has moved past would run a command nobody can see.
	if *opts.Highlight < 0 || *opts.Highlight >= len(matched) {
		*opts.Highlight = 0
	}

	d := overlay.Dialog(c, open, overlay.DialogOptions{
		NonModal: opts.NonModal,
		Body:     func() { paletteBody(c, open, matched, opts, label, placeholder, empty, &res) },
		Width:    opts.Width,
	})
	if d != nil {
		// The dialog carries no heading, so it is named here rather than
		// through overlay.DialogOptions: a panel nobody can name is a
		// window to a screen reader.
		d.Label(label)
	}
	res.Element = d
	return res
}

// paletteBody is the palette's own content: the field, the rows, and the keys.
// It is a separate function so that everything it closes over — the rows it
// was handed, the caller's query and highlight — reads the same as the code
// that decided them.
func paletteBody(c *ui.Context, open *bool, rows []Command, opts CommandPaletteOptions,
	label, placeholder, empty string, res *CommandPaletteResult) {
	k, u := core.Tokens(c), core.Density(c).Unit()

	// The keyboard is registered on this column, which is the panel's own
	// element: the keys belong to the palette, not to its field. It is
	// built empty and filled afterwards, because an element has to exist
	// before a closure inside its own children can ask it about a key.
	keys := ui.Box(c).FillWidth()
	keys.Children(func() {
		// Escape, read here and only for a non-modal palette. An overlay key
		// is delivered to one registration — the last one made — and
		// ui/overlay's layer has already asked for Escape by the time this
		// body is built when the layer is modal, so asking for it again here
		// would take it from the layer that owns it. A non-modal layer asks
		// for nothing, which is why the promise that Escape closes the palette
		// either way is kept in this body rather than in ui/overlay.
		if opts.NonModal && keys.OverlayShortcut(0, ui.KeyEscape) {
			*open = false
		}
		paletteStep(keys, len(rows), opts.Highlight)

		field := ui.TextInputBase(c, opts.Query).FillWidth().
			Padding(u*0.5, u*3).Radius(theme.ControlRadius).
			FontSize(core.FontSize(c, theme.BodySize)).TextColor(k.Text).
			Background(k.Surface).BorderWidth(theme.BorderWidth).
			BorderColor(k.Border).Placeholder(placeholder).
			Label(placeholder).AutoFocus()

		entered := field.Submitted()
		if len(rows) == 0 {
			ui.Text(c, empty).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.RowSize)).Padding(u*2, u*3)
		}
		for i, cmd := range rows {
			row := paletteRow(c, cmd, i == *opts.Highlight, u)
			// A press runs its row, and Enter runs the highlighted one. A
			// disabled row is left where it is and does neither: Enter on a
			// greyed command runs the command below it would be worse than
			// nothing happening.
			if row.Clicked() || (entered && i == *opts.Highlight) {
				if !cmd.Disabled {
					res.chosen, res.ran = cmd.ID, true
					*open = false
				}
			}
		}
	})
	keys.Label(label).Description(empty)
	// Read after the children, because a key moved it in there.
	res.highlight = *opts.Highlight
}

// paletteStep moves the caller's highlight by one row for Down and Up, and
// reports whether it moved.
//
// It wraps in both directions: a list of nine commands should take nine
// presses of Down to get back to the top rather than eleven, and somebody
// holding the key down should not have to know where the end is.
func paletteStep(panel *ui.Element, n int, highlight *int) bool {
	step := 0
	switch {
	case panel.OverlayShortcut(0, ui.KeyDown):
		step = 1
	case panel.OverlayShortcut(0, ui.KeyUp):
		step = -1
	}
	if step == 0 || n == 0 {
		return false
	}
	*highlight = ((*highlight+step)%n + n) % n
	return true
}

// paletteRow draws one command: its title, the line under it, and its
// shortcut at the right.
func paletteRow(c *ui.Context, cmd Command, on bool, u float32) *ui.Element {
	k := core.Tokens(c)

	bg, fg, sub := k.Surface, k.Text, k.TextMuted
	if on {
		bg = k.SurfaceHover
	}
	if cmd.Disabled {
		fg, sub = k.TextFaint, k.TextFaint
	}

	// Row, not Box: a command line is title, hint and shortcut across.
	// Stacked, the shortcut lands on the subtitle.
	row := ui.Row(c).FillWidth().Padding(u*2, u*3).Radius(theme.ControlRadius).
		Background(bg).AlignItems(ui.Center).Gap(u * 3).
		Label(cmd.Title).Disabled(cmd.Disabled)
	row.Children(func() {
		ui.Column(c).Grow(1).Gap(0).Children(func() {
			ui.Text(c, cmd.Title).TextColor(fg).
				FontSize(core.FontSize(c, theme.BodySize)).MaxLines(1)
			if cmd.Subtitle != "" {
				ui.Text(c, cmd.Subtitle).TextColor(sub).
					FontSize(core.FontSize(c, theme.MetaSize)).MaxLines(1)
			}
		})
		if len(cmd.Shortcut) > 0 && !cmd.Disabled {
			display.Kbd(c, cmd.Shortcut...)
		}
	})
	return row
}
