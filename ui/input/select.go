package input

import (
	"strconv"

	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// The dropdowns. Every one of them hangs its panel off a trigger with
// ui.PopoverBase — the same anchored layer ui/overlay.Popover is built on,
// which already moves the panel back inside the window when it would
// overflow it and closes it on Escape or on a press outside — and every one
// of them writes the caller's pointer rather than keeping a copy of the
// choice.

// dropdownMaxHeight is how tall a panel of options is allowed to be before it
// scrolls: a dozen rows, which is enough to read as a list rather than as a
// window, and short enough that a dropdown near the bottom of the window does
// not cover the page it was opened from.
const dropdownMaxHeight float32 = 260

// SelectOptions configure a Select.
type SelectOptions struct {
	// Label names the control for assistive technology, and is required: the
	// trigger's own text is the choice, so it changes as the choice changes
	// and leaves no stable name to read out.
	Label string
	// Placeholder is what the trigger says while nothing is chosen, such as
	// "Assign to". It is required for the reason a date field needs one: a
	// box that opens and says nothing looks like a text field that happens
	// to have an arrow in it.
	Placeholder string
	// Disabled greys the control out: it takes neither clicks nor focus.
	Disabled bool
}

// Select is a drop-down choosing one of a set, out of the caller's string.
//
// The value is that string and not an index, so a form submits it as it
// stands and a caller matching on it does not have to hold its choices in
// step with its state.
//
// It is built on ui.SelectBase, which is where the whole keyboard comes from:
// Down opens the panel, the arrows move the highlight, Enter chooses, Escape
// closes. Only the faces are this library's.
func Select(c *ui.Context, selected *string, choices []Choice, opts SelectOptions) *ui.Element {
	if selected == nil {
		panic("input: Select needs a selection to point at")
	}
	if len(choices) == 0 {
		panic("input: Select needs at least one choice")
	}
	if opts.Label == "" {
		panic("input: Select needs options.Label; its trigger shows the choice, so it has no name of its own to read out")
	}
	if opts.Placeholder == "" {
		panic("input: Select needs options.Placeholder; a trigger with no choice and no words is a box that opens")
	}
	checkChoicesHaveLabels(choices, "Select")
	if *selected != "" && !hasValue(checkChoices(choices), *selected) {
		panic("input: Select selected " + *selected + ", which is not one of its choices")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	sel := ui.SelectBase[string](c, selected)
	sel.Trigger.Height(core.ControlHeight(c)).Radius(theme.ControlRadius).
		Padding(0, u*2.5).Background(k.Background).TextColor(k.Text).
		BorderWidth(theme.BorderWidth).BorderColor(k.Border).
		FillWidth().Justify(ui.SpaceBetween).Label(opts.Label).Tooltip(opts.Label).
		Disabled(opts.Disabled)
	sel.Trigger.Children(func() {
		label, fg := opts.Placeholder, k.TextFaint
		for _, ch := range choices {
			if ch.Value == *selected {
				label, fg = ch.Label, k.Text
				break
			}
		}
		if opts.Disabled {
			fg = k.TextFaint
		}
		ui.Text(c, label).SingleLine().TextColor(fg).FontSize(core.FontSize(c, theme.BodySize))
		chevron(c)
	})

	sel.Popup(func(panel *ui.Element) {
		panelFace(c, panel)
		popupList(c, dropdownMaxHeight).Children(func() {
			for _, ch := range choices {
				item := sel.Item(ch.Value).FillWidth().Justify(ui.Start).
					Padding(u*0.75, u*1.5).Radius(theme.SmallRadius).TextColor(k.Text)
				switch {
				case item.Highlighted():
					// The muted highlight rather than the accent one a system
					// select uses: a whole row of accent is a shout, and this
					// row is walked one press at a time.
					item.Background(k.SurfacePressed)
				case ch.Value == *selected:
					item.Background(k.Surface)
				}
				item.Children(func() {
					ui.Text(c, ch.Label).SingleLine().FontSize(core.FontSize(c, theme.BodySize))
				})
			}
		})
	})
	return sel.Trigger
}

// SelectFieldOptions configure a SelectField.
type SelectFieldOptions struct {
	// Label names the control for assistive technology, and is required: the
	// trigger's own text is the choice, so it changes as the choice changes
	// and leaves no stable name to read out.
	Label string
	// Value is the words of the current choice, shown on the trigger. Empty
	// shows the Placeholder instead, in faint ink.
	Value string
	// Placeholder is what the trigger says while there is no choice, and is
	// required whenever Value is empty: a box with an arrow and nothing to
	// say is a box that opens.
	Placeholder string
	// Disabled greys the control out: it takes neither clicks nor focus.
	Disabled bool
}

// SelectField is the closed-state face of a dropdown: the box with the
// current choice's words and the arrow, before any panel is showing.
//
// It is the face Select's trigger wears, drawn through ui.SelectBase with
// the panel already wired to it, as a piece a caller can build a dropdown
// of its own from: SelectField for the closed state, SelectPanel for the
// open one, SelectRow for the rows in it. The caller is what keeps the
// panel's open flag and moves it between the two, because whether a list is
// showing is the moment of its own interface, the same reason dropdown
// triggers in this package keep theirs.
//
// It is a face rather than a control: it shows the choice it is told and
// takes no press of its own, so a press lands where the caller put the
// control rather than on a flag nobody handed it.
func SelectField(c *ui.Context, opts SelectFieldOptions) *ui.Element {
	if opts.Label == "" {
		panic("input: SelectField needs options.Label; its trigger shows the choice, so it has no name of its own to read out")
	}
	if opts.Value == "" && opts.Placeholder == "" {
		panic("input: SelectField needs a Value or a Placeholder; a trigger with no choice and no words is a box that opens")
	}
	trigger := ui.Box(c).Role(ui.RolePopUpButton).Label(opts.Label).Tooltip(opts.Label)
	selectTriggerFace(c, trigger, opts.Disabled)
	trigger.Children(func() { selectTriggerContent(c, opts.Value, opts.Placeholder, opts.Disabled) })
	return trigger
}

// SelectPanelOptions configure a SelectPanel.
type SelectPanelOptions struct {
	// Label names the panel for assistive technology and for a test to find
	// it by.
	Label string
	// MaxHeight is how tall the panel's list may be before it scrolls. Zero
	// is the dropdowns' own cap, a dozen rows' worth.
	MaxHeight float32
}

// SelectPanel is the open-state half of a dropdown: the floating surface a
// panel of options is built on, with the library's one floating look — the
// radius, the hairline, the shadow, the padding — and the list's own
// scrolling.
//
// body builds the rows inside it, which is why the rows are the caller's
// SelectRow and not this panel's business: a panel knows how it looks, and
// the caller knows what a row of its list is.
//
// It is the surface Select and MultiSelect build their panels on, taken out
// to a piece of its own. It is a list rather than a popup: it stays where
// its caller puts it, and the caller's PopoverBase is what hangs it off a
// trigger if it is to float.
func SelectPanel(c *ui.Context, opts SelectPanelOptions, body func()) *ui.Element {
	if body == nil {
		panic("input: SelectPanel needs a body to hold its rows; a panel with nothing in it is a shadow")
	}
	maxHeight := opts.MaxHeight
	if maxHeight <= 0 {
		maxHeight = dropdownMaxHeight
	}
	host := ui.Box(c).Role(ui.RoleList)
	panelFace(c, host)
	if opts.Label != "" {
		host.Label(opts.Label)
	}
	host.Children(func() {
		popupList(c, maxHeight).Children(body)
	})
	return host
}

// SelectRowOptions configure a SelectRow.
type SelectRowOptions struct {
	// Label is the row's words, and is required: a row with a tick and
	// nothing beside it is a mark with no name.
	Label string
	// Selected is that the row's option is the one chosen. A selected row is
	// filled and carries a SelectCheck in its mark slot, so it still reads
	// as the one with the colours off.
	Selected bool
	// Note is a second, faint word at the row's trailing edge, such as a
	// shortcut. Empty leaves it out.
	Note string
}

// SelectRow is one row of a single-choice panel: the mark slot, the words,
// and the selected state. It takes a press and reports it as Clicked, on the
// row's own element, which is the same convention every row in this package
// keeps — the caller that built the panel is the one that writes the choice
// the press meant.
//
// Its mark slot is a check and not a box, which is the difference from a
// MultiSelectCheck: a single choice has no group to check into, and a box in
// front of the chosen row would read as a second control the row is sitting
// next to.
func SelectRow(c *ui.Context, opts SelectRowOptions) *ui.Element {
	if opts.Label == "" {
		panic("input: SelectRow needs a Label; a tick with no words beside it has nothing to read out")
	}
	f := optionFace{Chosen: opts.Selected, Note: opts.Note}
	if opts.Selected {
		f.Mark = checkGlyph
	}
	return optionRow(c, opts.Label, "", f)
}

// SelectCheck is the tick on a selected row, drawn alone: the two strokes a
// check is, in the window's accent, at the size a row's mark slot is.
//
// It is the same path markFace draws inside its box, taken out so a row, a
// table and a result line can all mark what is chosen with one tick rather
// than each inventing its own.
func SelectCheck(c *ui.Context) *ui.Element {
	return markFace(c, checkGlyph, core.Density(c).Unit()*3.75)
}

// SelectNote is the quiet line at the bottom of a panel: a keyboard hint
// such as "↑↓ select · ↵ confirm", in the faintest ink the window has, at
// the size a caption is.
//
// It is the words the caller gives it, not copy the library owns, because
// the hint is about the caller's panel — its keys, its confirm — and a
// library word in the place of a panel's own instruction would be the
// library talking where the panel should be.
func SelectNote(c *ui.Context, text string) *ui.Element {
	if text == "" {
		panic("input: SelectNote needs a text; a hairline with no words is a divider")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	return ui.Box(c).Padding(u*0.5, u*1.5).Children(func() {
		ui.Text(c, text).SingleLine().TextColor(k.TextFaint).
			FontSize(core.FontSize(c, theme.CaptionSize))
	})
}

// MultiSelectCheckOptions configure a MultiSelectCheck.
type MultiSelectCheckOptions struct {
	// Label is the row's words, and is required, as SelectRow's is.
	Label string
	// Checked is that the row's option is in the caller's selection. A
	// checked row is filled and its box carries the tick; an unchecked one
	// shows the empty box, so the row says what pressing it will do.
	Checked bool
	// Note is a second, faint word at the row's trailing edge. Empty leaves
	// it out.
	Note string
}

// MultiSelectCheck is one checkbox-style row of a multi-choice: the box, its
// checked state, and the words beside it, taking a press and reporting it as
// Clicked for the caller to add to or take out of its selection.
//
// It is the row MultiSelect draws its options with, taken out to a piece of
// its own. The difference from SelectRow is the mark: a multi-choice's rows
// carry a box, because the choice is a set and a box is what a member of a
// set is marked with — SelectRow's lone check would say "the one" where
// this row may be the seventh of twenty.
func MultiSelectCheck(c *ui.Context, opts MultiSelectCheckOptions) *ui.Element {
	if opts.Label == "" {
		panic("input: MultiSelectCheck needs a Label; a box and a tick with no words beside them have nothing to read out")
	}
	m := tickMark
	if opts.Checked {
		m = tickOn
	}
	return optionRow(c, opts.Label, "", optionFace{Mark: m, Chosen: opts.Checked, Note: opts.Note})
}

// SelectSearchOptions configure a SelectSearch.
type SelectSearchOptions struct {
	// Label names the control, as Select's does, and is required.
	Label string
	// Placeholder is the trigger's text while nothing is chosen.
	Placeholder string
	// Search is the panel's field placeholder, and is required for the same
	// reason: a list that opens onto an empty box cannot be narrowed.
	Search string
	// Disabled greys the control out.
	Disabled bool
}

// SelectSearch is a drop-down of one of a set with a field above it that
// narrows the list as the query is typed.
//
// The query is the caller's, next to the selection, because the two are read
// together: a form that submits a search has to submit what was searched for
// as well as what was found. The field is this package's own SearchField, so
// the query inside the panel looks like the query above it, with the same
// magnifier, the same clear button and the same Escape behaviour.
//
// The panel stays open as the query changes — closing on every keystroke
// would make it impossible to keep typing — and Enter takes the first match,
// which is the one the prefix ordering put at the top and so the one most
// likely to be the one meant.
func SelectSearch(c *ui.Context, selected, query *string, choices []Choice, opts SelectSearchOptions) *ui.Element {
	if selected == nil {
		panic("input: SelectSearch needs a selection to point at")
	}
	if query == nil {
		panic("input: SelectSearch needs a query to point at")
	}
	if len(choices) == 0 {
		panic("input: SelectSearch needs at least one choice")
	}
	if opts.Label == "" {
		panic("input: SelectSearch needs options.Label; its trigger shows the choice, so it has no name of its own to read out")
	}
	if opts.Placeholder == "" {
		panic("input: SelectSearch needs options.Placeholder; a trigger with no choice and no words is a box that opens")
	}
	if opts.Search == "" {
		panic("input: SelectSearch needs options.Search for the field in its panel")
	}
	checkChoicesHaveLabels(choices, "SelectSearch")
	if *selected != "" && !hasValue(checkChoices(choices), *selected) {
		panic("input: SelectSearch selected " + *selected + ", which is not one of its choices")
	}

	trigger, open := dropdownTrigger(c, opts.Label, labelOfChoice(choices, *selected), opts.Placeholder)
	ui.PopoverBase(c, trigger, open, func(panel *ui.Element) {
		panelFace(c, panel)
		shown := matchingChoices(choices, *query)
		field := SearchField(c, query, opts.Search)
		switch {
		case field.Submitted() && len(shown) > 0:
			*selected = shown[0].Value
			*open = false
		case len(shown) == 0:
			empty(c, noMatches(c))
		default:
			popupList(c, dropdownMaxHeight-core.ControlHeight(c)).Children(func() {
				for _, ch := range shown {
					item := optionRow(c, ch.Label, ch.Value, optionFace{Chosen: ch.Value == *selected})
					if item.Clicked() {
						*selected = ch.Value
						*open = false
					}
				}
			})
		}
	})
	return trigger
}

// MultiSelectOptions configure a MultiSelect.
type MultiSelectOptions struct {
	// Label names the control, as Select's does, and is required.
	Label string
	// Placeholder is the trigger's text while nothing is chosen.
	Placeholder string
	// Named caps how many of the chosen labels the trigger says before it
	// counts the rest. Zero is one, which is what fits on a line.
	Named int
	// Disabled greys the control out.
	Disabled bool
}

// MultiSelectResult is a MultiSelect and the one thing its pointers cannot
// say.
type MultiSelectResult struct {
	// Element is the trigger, with the panel built under it while it shows.
	Element *ui.Element
	// dismissed is that the panel was showing and this frame it is not.
	dismissed bool
}

// Dismissed reports that the panel was open and this frame it is not, because
// Escape was pressed or the press landed elsewhere on the page. It is the
// moment to re-apply whatever the chosen values filter, because until then
// they may still change.
func (r MultiSelectResult) Dismissed() bool { return r.dismissed }

// MultiSelect is a drop-down choosing any number of a set, out of the
// caller's []string.
//
// It keeps the panel open as options are pressed, which is the whole point of
// it and the reason it cannot be a Select with a wider slice behind it: a set
// of twenty is chosen by pressing twenty, so closing on the first would make
// the control unusable for anything over about three.
//
// It also carries Select all and Clear. A multi-select without them can only
// be emptied one press at a time, and a count of twenty is exactly the case
// where that hurts.
func MultiSelect(c *ui.Context, selected *[]string, choices []Choice, opts MultiSelectOptions) MultiSelectResult {
	if selected == nil {
		panic("input: MultiSelect needs a selection to point at")
	}
	if len(choices) == 0 {
		panic("input: MultiSelect needs at least one choice")
	}
	if opts.Label == "" {
		panic("input: MultiSelect needs options.Label; its trigger shows the choice, so it has no name of its own to read out")
	}
	if opts.Placeholder == "" {
		panic("input: MultiSelect needs options.Placeholder; a trigger with no choice and no words is a box that opens")
	}
	checkChoicesHaveLabels(choices, "MultiSelect")
	all := checkChoices(choices)
	u := core.Density(c).Unit()

	named := opts.Named
	if named <= 0 {
		named = 1
	}
	trigger, open := dropdownTrigger(c, opts.Label, summaryOfChosen(choices, *selected, named), opts.Placeholder)
	was := *open
	ui.PopoverBase(c, trigger, open, func(panel *ui.Element) {
		panelFace(c, panel)
		ui.Row(c).FillWidth().Gap(u).Justify(ui.End).Children(func() {
			take := Button(c, selectAllLabel(c), ButtonOptions{Label: selectAllLabel(c)})
			if take.Clicked() {
				// The choices' own order rather than the slice's, so a set
				// filled from the top always reads back the same way.
				*selected = all
			}
			leave := Button(c, clearAllLabel(c), ButtonOptions{Label: clearAllLabel(c)})
			if leave.Clicked() {
				*selected = nil
			}
		})
		popupList(c, dropdownMaxHeight-core.ControlHeight(c)).Children(func() {
			for _, ch := range choices {
				chosen := hasValue(*selected, ch.Value)
				m := tickMark
				if chosen {
					m = tickOn
				}
				item := optionRow(c, ch.Label, ch.Value, optionFace{Mark: m, Chosen: chosen})
				if !item.Clicked() {
					continue
				}
				if chosen {
					*selected = without(*selected, ch.Value)
					continue
				}
				*selected = append(*selected, ch.Value)
			}
		})
	})
	return MultiSelectResult{Element: trigger, dismissed: was && !*open}
}

// CascaderOptions configure a Cascader.
type CascaderOptions struct {
	// Label names the control, as Select's does, and is required.
	Label string
	// Placeholder is the trigger's text while nothing is chosen.
	Placeholder string
	// Separator joins the chosen path in the closed trigger, since the value
	// of a cascader is a path and a reader wants it on one line. It is " / "
	// unless the caller says otherwise.
	Separator string
	// Disabled greys the control out.
	Disabled bool
}

// Cascader is a choice in levels: a city, then a district of it. Each level is
// a column in the panel, and the columns are the ones the path has already
// chosen, so a path two deep is two columns rather than a tree to dig through.
//
// The path is the caller's []string, from the top down, and it is checked
// against the nodes before anything is drawn: a path that is not in the tree
// would otherwise draw a panel of nothing and read as an empty level rather
// than as a stale value.
func Cascader(c *ui.Context, path *[]string, nodes []Node, opts CascaderOptions) *ui.Element {
	if path == nil {
		panic("input: Cascader needs a path to point at")
	}
	if len(nodes) == 0 {
		panic("input: Cascader needs at least one node")
	}
	if opts.Label == "" {
		panic("input: Cascader needs options.Label; its trigger shows the choice, so it has no name of its own to read out")
	}
	if opts.Placeholder == "" {
		panic("input: Cascader needs options.Placeholder; a trigger with no choice and no words is a box that opens")
	}
	sep := opts.Separator
	if sep == "" {
		sep = " / "
	}
	checkNodes(nodes, "Cascader")
	if !pathIn(nodes, *path) {
		panic("input: Cascader path [" + strings.Join(*path, sep) + "] is not a path through the tree it was given")
	}
	u := core.Density(c).Unit()

	trigger, open := dropdownTrigger(c, opts.Label, summaryOfPath(nodes, *path, sep), opts.Placeholder)
	ui.PopoverBase(c, trigger, open, func(panel *ui.Element) {
		panelFace(c, panel)
		columns := ui.Row(c).AlignItems(ui.Start).Gap(u)
		columns.Children(func() {
			level := nodes
			// One column per level the path has reached, plus the one it
			// could reach next: an empty path shows the roots alone, which is
			// what makes a cascader usable without opening anything first.
			for depth := 0; depth <= len(*path) && len(level) > 0; depth++ {
				here, at := level, depth
				// The column is made here, inside the row's own Children
				// call, because that is what makes it the row's child rather
				// than its neighbour.
				col := ui.Column(c).Shrink(0).Gap(u * 0.25)
				col.Children(func() {
					for _, n := range here {
						// A node already on the path is filled, so the chosen
						// line can be read off the column without counting
						// columns across.
						chosen := at < len(*path) && (*path)[at] == n.Value
						item := optionRow(c, n.Label, n.Value, optionFace{Chosen: chosen})
						if !item.Clicked() {
							continue
						}
						// Choosing one step deeper drops the rest: the
						// district of a different city is not still chosen.
						keep := (*path)[:min(at, len(*path))]
						next := make([]string, 0, len(keep)+1)
						next = append(next, keep...)
						*path = append(next, n.Value)
					}
				})
				if at >= len(*path) {
					break
				}
				step := findNode(level, (*path)[at])
				if step == nil {
					break
				}
				level = step.Children
			}
		})
	})
	return trigger
}

// ComboboxOptions configure a Combobox.
type ComboboxOptions struct {
	// Label names the control for assistive technology. The field's text is
	// whatever was typed, so it is required for the reason Select's is.
	Label string
	// Disabled greys the control out.
	Disabled bool
}

// Combobox is a field that suggests: it takes what is typed, narrows the
// suggestions to those containing it, and hands back the one that was taken.
//
// The value is the caller's string and not an index into the suggestions, so
// the set of suggestions may be whatever the caller has this frame — forty
// thousand branches narrowed by a backend, say — and the control does not need
// it to be a fixed set it holds against an index.
//
// It is built on MyGo's own combobox, which is where the keyboard comes from:
// Up and Down move among the suggestions and Enter takes one, as the text
// stops being a filter and becomes a choice.
func Combobox(c *ui.Context, selected *string, suggestions []string, opts ComboboxOptions) *ui.Element {
	if selected == nil {
		panic("input: Combobox needs a value to point at")
	}
	if opts.Label == "" {
		panic("input: Combobox needs options.Label; its field shows what was typed, so it has no name of its own to read out")
	}
	for i, s := range suggestions {
		if s == "" {
			panic("input: Combobox suggestion " + strconv.Itoa(i) + " is empty")
		}
	}
	return ui.Combobox(c, selected, suggestions).Label(opts.Label).Tooltip(opts.Label)
}

// TreeSelectOptions configure a TreeSelect.
type TreeSelectOptions struct {
	// Label names the control, as Select's does, and is required.
	Label string
	// Placeholder is the trigger's text while nothing is chosen.
	Placeholder string
	// Disabled greys the control out.
	Disabled bool
}

// TreeSelect is a choice from a tree rather than from a list: a branch can be
// chosen as well as a leaf, and the branches open where they were left rather
// than starting shut every time.
//
// The branches' own open state is the control's and not the caller's. It is
// not a fact about the data — nothing downstream of it changes — and a caller
// holding a bool per node would be modelling the interface in its own state.
func TreeSelect(c *ui.Context, selected *string, nodes []Node, opts TreeSelectOptions) *ui.Element {
	if selected == nil {
		panic("input: TreeSelect needs a selection to point at")
	}
	if len(nodes) == 0 {
		panic("input: TreeSelect needs at least one node")
	}
	if opts.Label == "" {
		panic("input: TreeSelect needs options.Label; its trigger shows the choice, so it has no name of its own to read out")
	}
	if opts.Placeholder == "" {
		panic("input: TreeSelect needs options.Placeholder; a trigger with no choice and no words is a box that opens")
	}
	checkNodes(nodes, "TreeSelect")
	if *selected != "" && findNodeIn(nodes, *selected) == nil {
		panic("input: TreeSelect selected " + *selected + ", which is not in the tree it was given")
	}

	trigger, open := dropdownTrigger(c, opts.Label, labelOfNode(nodes, *selected), opts.Placeholder)
	state := *ui.Local(trigger, treeKey{}, func() *treeState { return newTreeState() })
	ui.PopoverBase(c, trigger, open, func(panel *ui.Element) {
		panelFace(c, panel)
		popupList(c, dropdownMaxHeight).Children(func() {
			ui.Tree(c, func() { treeRows(c, nodes, state, selected) })
		})
	})
	return trigger
}

// treeRows draws a tree into a TreeSelect's panel, choosing into selected.
func treeRows(c *ui.Context, nodes []Node, state *treeState, selected *string) {
	for _, n := range nodes {
		if len(n.Children) == 0 {
			if ui.TreeItem(c, n.Label, nil, nil).Clicked() {
				*selected = n.Value
			}
			continue
		}
		// A branch is a choice too, which is the reason this is a tree
		// select and not a tree: a rule that covers every branch has to be
		// nameable.
		if ui.TreeItem(c, n.Label, state.openFor(n.Value), func() {
			treeRows(c, n.Children, state, selected)
		}).Clicked() {
			*selected = n.Value
		}
	}
}

// Node is one branch of a tree a Cascader or a TreeSelect chooses from.
type Node struct {
	// Value is what the caller's state holds for it. It is separate from
	// Label for the same reason a Choice's is: the value is matched on and
	// submitted, the label is read.
	Value string
	// Label is the words shown for it.
	Label string
	// Children are the branch's own, empty for a leaf.
	Children []Node
}

// treeKey is where a TreeSelect's open branches live.
type treeKey struct{}

// treeState is which branches of a tree are open, by value. It holds pointers
// rather than bools because ui.TreeItem keeps the pointer it is given for as
// long as it draws, and a map entry cannot be addressed.
type treeState struct{ open map[string]*bool }

func newTreeState() *treeState { return &treeState{open: map[string]*bool{}} }

func (s *treeState) openFor(v string) *bool {
	if p, ok := s.open[v]; ok {
		return p
	}
	p := new(bool)
	s.open[v] = p
	return p
}

// checkNodes rejects a tree with a valueless or labelless node, which the two
// tree controls have no way to draw.
func checkNodes(nodes []Node, who string) {
	for _, n := range nodes {
		if n.Value == "" {
			panic("input: " + who + " has a node with no Value; a branch nothing can be chosen by is not a choice")
		}
		if n.Label == "" {
			panic("input: " + who + " node " + n.Value + " has no Label")
		}
		checkNodes(n.Children, who)
	}
}

// findNode is the node of value among nodes, or nil.
func findNode(nodes []Node, value string) *Node {
	for _, n := range nodes {
		if n.Value == value {
			found := n
			return &found
		}
	}
	return nil
}

// findNodeIn is the node of value anywhere in the tree, or nil.
func findNodeIn(nodes []Node, value string) *Node {
	if n := findNode(nodes, value); n != nil {
		return n
	}
	for _, n := range nodes {
		if deep := findNodeIn(n.Children, value); deep != nil {
			return deep
		}
	}
	return nil
}

// pathIn reports whether path is a real path down the tree: each of its steps
// a node of the level before it.
func pathIn(nodes []Node, path []string) bool {
	level := nodes
	for i, v := range path {
		step := findNode(level, v)
		if step == nil {
			return false
		}
		if i == len(path)-1 {
			return true
		}
		level = step.Children
	}
	return true
}

// labelOfChoice is what a trigger showing one choice says: the choice's own
// words, or "" while there is no choice, which the trigger shows faintly.
func labelOfChoice(choices []Choice, selected string) string {
	for _, ch := range choices {
		if ch.Value == selected {
			return ch.Label
		}
	}
	return ""
}

// summaryOfChosen is what a trigger showing any number of choices says.
func summaryOfChosen(choices []Choice, selected []string, named int) string {
	if len(selected) == 0 {
		return ""
	}
	labels := make([]string, 0, len(selected))
	for _, ch := range choices {
		if hasValue(selected, ch.Value) {
			labels = append(labels, ch.Label)
		}
	}
	return summary(labels, named)
}

// labelOfNode is what a tree's trigger says: the chosen node's words, or "".
func labelOfNode(nodes []Node, selected string) string {
	if n := findNodeIn(nodes, selected); n != nil {
		return n.Label
	}
	return ""
}

// summaryOfPath is what a cascader's trigger says: the whole path, which is
// the one case where the value is more than one word.
func summaryOfPath(nodes []Node, path []string, sep string) string {
	names := make([]string, 0, len(path))
	level := nodes
	for _, v := range path {
		step := findNode(level, v)
		if step == nil {
			break
		}
		names = append(names, step.Label)
		level = step.Children
	}
	if len(names) == len(path) {
		return strings.Join(names, sep)
	}
	return summary(names, len(names))
}

// empty is what a panel says when it has nothing to list.
func empty(c *ui.Context, what string) *ui.Element {
	k := core.Tokens(c)
	return ui.Text(c, what).TextColor(k.TextFaint).
		FontSize(core.FontSize(c, theme.RowSize)).SingleLine()
}
