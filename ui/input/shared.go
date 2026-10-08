package input

import (
	"math"
	"strconv"
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// The parts every choice control in this package is built from live here, so
// a checkbox in a settings list, a tick in a dropdown and a box in front of a
// multi-select option are literally the same drawing.

// Choice is one option of a group control: the value the caller's state
// holds, and the words shown for it.
//
// The two are separate because the value is what a form submits and what a
// caller matches on, while the label is what a person reads and what a
// window may reword. A choice whose Label is empty is not a choice: the row
// would be an unmarked box, so the controls that draw one panic without it.
type Choice struct {
	Value string
	Label string
}

// mark is what sits in front of an option: nothing, an empty box, a box with
// a tick in it, or a dash for a set that is part chosen.
type mark int

const (
	noMark mark = iota
	tickMark
	tickOn
	dashMark
	// checkGlyph is the tick alone, without the box around it, for a row
	// that marks its chosen state with a check rather than with a filled
	// box — a single-choice list's selected row, which a box in front of it
	// would read as a second control.
	checkGlyph
)

// checkChoices is every Choice's value in the order they were given, which is
// what a control checks a selection against and what Select all writes.
func checkChoices(choices []Choice) []string {
	out := make([]string, len(choices))
	for i, ch := range choices {
		out[i] = ch.Value
	}
	return out
}

// hasValue reports whether v is in list.
func hasValue(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}

// without is list with every copy of v taken out, keeping the order of what
// is left. A group rewrites the caller's slice rather than mutating it, so
// the caller's own view of what is chosen cannot hold a value the control has
// already taken back.
func without(list []string, v string) []string {
	out := make([]string, 0, len(list))
	for _, s := range list {
		if s != v {
			out = append(out, s)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// inkOn is the ink that reads on top of bg: white or black, whichever
// contrasts more with it.
//
// A mark on a color the caller chose cannot use a theme token, because tokens
// are picked for the window rather than for every color a person can choose.
// The same mark that is legible on a pale swatch vanishes on a dark one, and
// in a dark window it is the dark swatch that is easy to lose. Asking the
// color what ink it takes is what keeps the mark visible either way.
func inkOn(bg ui.Color) ui.Color {
	if bg.A < 255 {
		// A mark on a translucent swatch has to decide against whatever
		// shows through it; a mid gray stands in for an unknown backdrop, and
		// neither ink then wins, so the choice is at least a stable one.
		bg = bg.Over(ui.RGB(128, 128, 128))
	}
	white, black := ui.RGB(255, 255, 255), ui.RGB(0, 0, 0)
	if contrast(white, bg) >= contrast(black, bg) {
		return white
	}
	return black
}

// contrast is the WCAG ratio between a and b, from 1 to 21. It is the same
// measure ui/theme's test holds the library to, so a mark judged by it here
// cannot be less visible than the text the test already guards.
func contrast(a, b ui.Color) float32 {
	l1, l2 := relLum(a), relLum(b)
	if l1 < l2 {
		l1, l2 = l2, l1
	}
	return (l1 + 0.05) / (l2 + 0.05)
}

// relLum is the relative luminance of c, as WCAG defines it. The colors here
// are straight sRGB, so the transfer function is the sRGB one rather than a
// plain average: a dark blue and a dark green are not the same darkness to
// the eye, and an average of their bytes would not know.
func relLum(c ui.Color) float32 {
	lin := func(v uint8) float32 {
		f := float32(v) / 255
		if f <= 0.03928 {
			return f / 12.92
		}
		return float32(math.Pow(float64(f+0.055)/1.055, 2.4))
	}
	return 0.2126*lin(c.R) + 0.7152*lin(c.G) + 0.0722*lin(c.B)
}

// markFace draws the box in front of an option or a choice: empty, a tick, or
// a dash for a set that is part chosen.
//
// A chosen mark is a tick in OnFill on Accent, the one pair of tokens
// ui/theme's test holds to 4.5:1, rather than a tick in the window's text
// colour. That is not a detail: a tick in Text on an Accent box disappears the
// moment the accent is dark, which is exactly the case a dark window hits.
func markFace(c *ui.Context, m mark, side float32) *ui.Element {
	k := core.Tokens(c)
	box := ui.Box(c).Size(side, side).Shrink(0).Role(ui.RoleNone).Radius(side * 0.28)
	switch m {
	case tickOn, dashMark:
		box.Background(k.Accent)
	case tickMark:
		box.Background(k.Background).
			BorderWidth(theme.BorderWidth).BorderColor(k.Border.Mix(k.Text, 0.25))
	case checkGlyph:
		// The tick stands alone, in the window's accent: the one colour that
		// says "this is the one" without a surface to stand on.
		box.DrawOver(func(p *ui.Painter, r ui.Rect) { paintTick(p, r, k.Accent) })
		return box
	default:
		return box
	}
	// The accent is a token, but a window may set it to anything at all, so
	// even here the mark asks its own background which ink it takes.
	ink := inkOn(k.Accent)
	box.DrawOver(func(p *ui.Painter, r ui.Rect) {
		switch m {
		case tickOn:
			paintTick(p, r, ink)
		case dashMark:
			p.Line(r.X+r.W*0.22, r.Y+r.H/2, r.X+r.W*0.78, r.Y+r.H/2, 1.5, ink)
		}
	})
	return box
}

// paintTick is the tick's own path, which a mark box and a stand-alone
// SelectCheck both need: the two strokes a check is, drawn at whatever size
// its box gives it.
func paintTick(p *ui.Painter, r ui.Rect, ink ui.Color) {
	var path ui.Path
	path.MoveTo(r.X+r.W*0.24, r.Y+r.H*0.52).
		LineTo(r.X+r.W*0.44, r.Y+r.H*0.72).
		LineTo(r.X+r.W*0.78, r.Y+r.H*0.3)
	p.StrokePath(&path, 1.5, ink)
}

// chevron draws the arrow of a control that opens something below it. It is
// drawn here rather than taken from MyGo, which keeps its own as a private
// helper; the path is the same one, in the window's muted text.
func chevron(c *ui.Context) {
	k := core.Tokens(c)
	side := core.Density(c).Unit() * 2.5
	ui.Box(c).Size(side, side).Shrink(0).Role(ui.RoleNone).
		Draw(func(p *ui.Painter, r ui.Rect) {
			var path ui.Path
			path.MoveTo(r.X+r.W*0.1, r.Y+r.H*0.3).
				LineTo(r.X+r.W*0.5, r.Y+r.H*0.7).
				LineTo(r.X+r.W*0.9, r.Y+r.H*0.3)
			p.StrokePath(&path, 1.5, k.TextMuted)
		})
}

// panelFace gives the panel of a dropdown the look of the library's floating
// layers.
//
// It is a call, not a copy. It used to be a copy: ui/overlay builds its own
// buttons with input.Button, so input could not import overlay, and the panel
// was written out again here with a comment saying to keep the two in step.
// The panel now lives in ui/layout — below both — so there is one of it and
// this can say so.
func panelFace(c *ui.Context, panel *ui.Element) *ui.Element {
	return layout.Panel(c, panel, layout.PanelOptions{Compact: true}, nil)
}

// openKey is where a dropdown's open flag lives. It is a type of its own so
// that no other control on the same element can be found by the same name.
type openKey struct{}

// dropdownTrigger builds the button a panel of options hangs under, and the
// *bool saying whether that panel is showing.
//
// open is the control's own rather than the caller's, and that is the one
// piece of state a choice control keeps. Whether a list is showing is a
// moment of the interface, not a fact about the data: a caller holding it
// would have to remember to close it after every choice, and a form with two
// dropdowns in it would have to keep two flags in step. It lives on the
// trigger element, which MyGo keeps across frames, so the two dropdowns of a
// window open and close on their own.
//
// show is the text for the closed state and placeholder what it says while
// there is nothing chosen. Both are passed in rather than computed, because
// what a dropdown's trigger says — a count, a node's name, a whole path — is
// the caller's to say and not the library's to guess.
func dropdownTrigger(c *ui.Context, name, show, placeholder string) (*ui.Element, *bool) {
	trigger := ui.ButtonBase(c)
	selectTriggerFace(c, trigger, false)
	trigger.Label(name).Tooltip(name)
	open := ui.Local(trigger, openKey{}, func() bool { return false })
	if trigger.Clicked() {
		*open = !*open
	}
	trigger.Children(func() { selectTriggerContent(c, show, placeholder, false) })
	return trigger, open
}

// selectTriggerFace gives an element the look of a dropdown's closed-state
// trigger: the control's height and radius, the border, the room the arrow
// takes. It is the same face Select's trigger wears, drawn through
// ui.SelectBase, so the two cannot drift apart.
func selectTriggerFace(c *ui.Context, e *ui.Element, disabled bool) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	e.Height(core.ControlHeight(c)).Radius(theme.ControlRadius).
		Padding(0, u*2.5).Background(k.Background).TextColor(k.Text).
		BorderWidth(theme.BorderWidth).BorderColor(k.Border).
		FillWidth().Justify(ui.SpaceBetween)
	if disabled {
		e.Disabled(true)
	}
}

// selectTriggerContent is what a dropdown's trigger shows: the current
// choice's words, or the placeholder in faint ink while there is no choice,
// and the arrow that says what pressing it does.
func selectTriggerContent(c *ui.Context, show, placeholder string, disabled bool) {
	k := core.Tokens(c)
	label, fg := show, k.Text
	if label == "" {
		// Nothing chosen yet: the trigger says so in the faint ink, in
		// the caller's own words, so it is not mistaken for a value.
		label, fg = placeholder, k.TextFaint
	}
	if disabled {
		fg = k.TextFaint
	}
	ui.Text(c, label).SingleLine().TextColor(fg).FontSize(core.FontSize(c, theme.BodySize))
	chevron(c)
}

// optionFace is what an option row in a panel is showing.
type optionFace struct {
	// Mark is the box in front of the label, or noMark for a row that has
	// none — a single-choice list, where the chosen row is marked by its
	// whole face instead.
	Mark mark
	// Chosen is that the row's option is in the caller's selection. A chosen
	// row is filled, so it still reads as chosen with the colours off.
	//
	// There is no "highlighted" here. A single-choice panel gets that from
	// ui.SelectBase, which tracks the row the pointer or the arrows are on;
	// a panel built out of plain rows has nothing to track them with, and
	// inventing a second highlight for it would be one that follows the
	// pointer and not the keyboard.
	Chosen bool
	// Note is a second, faint word at the row's trailing edge: a shortcut
	// hint, a count, a date. It is drawn only when present, so a row with
	// none is exactly what it was.
	Note string
}

// optionRow is one row of a panel of options, taking a press and reporting
// it as Clicked. Every dropdown in the package draws its rows with it, so a
// single choice, a search result and a checkbox in a multi-select all read
// the same.
//
// key is the option's own identifier among its siblings, and empty where the
// caller has none to give. It is what a filtered panel needs: its rows are
// built in match order, which changes on every keystroke, and a row MyGo
// identifies by where it sits loses the keyboard to a different option the
// moment the list moves under it.
//
// It is a ButtonBase rather than a plain row, which is the only exported way
// to get a row that takes presses. The cost is that every option is a stop of
// Tab, which inside a panel is the point: the panel is a layer the keyboard
// walks, not something only a pointer can reach.
func optionRow(c *ui.Context, label, key string, f optionFace) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	row := ui.ButtonBase(c).FillWidth().Justify(ui.Start).AlignItems(ui.Center).Gap(u).
		Padding(u*0.75, u*1.5).Radius(theme.SmallRadius).TextColor(k.Text)
	if key != "" {
		row = row.Key(key)
	}
	// Named on the row and not only on the words inside it: what a reader is
	// handed, and what a test clicking by name lands on, is the row.
	row = row.Label(label)
	// The pointer's own row, read at build time as MyGo's own Link reads its
	// hover. A panel with no hover reads as a picture rather than as
	// something to point at, and pointing is how most of its rows get chosen.
	switch {
	case f.Chosen:
		row.Background(k.Surface)
	case row.Hovered():
		row.Background(k.SurfaceHover)
	}
	row.Children(func() {
		if f.Mark != noMark {
			markFace(c, f.Mark, u*3.75)
		}
		ui.Text(c, label).SingleLine().FontSize(core.FontSize(c, theme.BodySize))
		if f.Note != "" {
			ui.Text(c, f.Note).SingleLine().TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.MetaSize))
		}
	})
	return row
}

// matchingChoices is the choices containing q, ignoring case, those starting
// with it first. It is what a search field filters a panel with: the prefix
// matches are what a person typing the first letters of a name means, and
// they are the ones to show first.
func matchingChoices(choices []Choice, q string) []Choice {
	q = strings.ToLower(strings.TrimSpace(q))
	if q == "" {
		return choices
	}
	var starts, contains []Choice
	for _, ch := range choices {
		switch l := strings.ToLower(ch.Label); {
		case strings.HasPrefix(l, q):
			starts = append(starts, ch)
		case strings.Contains(l, q):
			contains = append(contains, ch)
		}
	}
	return append(starts, contains...)
}

// summary is what a trigger showing several choices says: up to limit of them
// by name, then how many are left. Naming three of twenty says more than
// "20 selected", and naming twenty says nothing at all.
func summary(labels []string, limit int) string {
	switch {
	case len(labels) == 0:
		return ""
	case len(labels) <= limit:
		return strings.Join(labels, ", ")
	}
	return strings.Join(labels[:limit], ", ") + " +" + strconv.Itoa(len(labels)-limit)
}

// popupList is the scrolling column of rows a panel holds. It scrolls because
// a dropdown of two hundred branches must not push the page it hangs over
// off the window.
func popupList(c *ui.Context, maxHeight float32) *ui.Element {
	u := core.Density(c).Unit()
	return ui.Scroll(c).MaxHeight(maxHeight).Gap(u * 0.25)
}

// noMatches is what a searchable panel says when the query left it empty. It
// goes through core.Msg so a window can reword it, as it can every other word
// the library owns.
func noMatches(c *ui.Context) string {
	return core.Msg(c, "input.noMatches", core.Def("No matches"))
}

// selectAllLabel and clearAllLabel are the words on the two actions every
// multi-select needs. Without them a list of twenty is only emptied by
// twenty presses, which is the reason these two exist at all.
func selectAllLabel(c *ui.Context) string {
	return core.Msg(c, "input.selectAll", core.Def("Select all"))
}

func clearAllLabel(c *ui.Context) string {
	return core.Msg(c, "input.clearAll", core.Def("Clear all"))
}

// put makes the element the parent will hold.
//
// It is a note rather than a helper because the rule it is about is the one
// MyGo's Children hides: a child is whatever was CREATED while the Children
// call was running, so an element made before its container is that
// container's sibling, and no amount of putting it in afterwards will make it
// a child. A container is therefore always built first and its parts built
// inside it, which is why Slider and RangeSlider both put the rail where it is
// to sit rather than making it first and moving it in.

// checkChoicesHaveLabels rejects a set with a choice nothing is drawn for, so
// a caller that forgot a label hears about it at the control rather than
// finding an empty row in a screenshot.
func checkChoicesHaveLabels(choices []Choice, who string) {
	for i, ch := range choices {
		if ch.Label == "" {
			panic("input: " + who + " choice " + strconv.Itoa(i) + " has no Label; the row would be a mark with no words beside it")
		}
	}
}
