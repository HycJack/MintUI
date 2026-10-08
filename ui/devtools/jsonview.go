package devtools

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/theme"
)

// JsonViewerOptions configure a JsonViewer.
type JsonViewerOptions struct {
	// Source is the JSON text as it came off the wire. It is parsed here
	// rather than handed over parsed, because a response viewer whose whole
	// job is to look at what came back should not need a caller to have
	// already decided whether it is JSON.
	Source string
	// Expanded is the caller's open/closed state for each path, so that a
	// branch somebody opened stays open across frames and a branch they shut
	// stays shut. It is the caller's because "which branches are open" is a
	// place in a response somebody can come back to.
	Expanded map[string]bool
	// Toggled writes the path of the branch that was opened or shut. Empty
	// when none was.
	Toggled *string
	// Copied writes the path of the row whose copy button was pressed, empty
	// when none was. The text itself is the caller's: this component has no
	// clipboard, and drawing a value into one is the caller's business.
	Copied *string
	// Search narrows the tree to the branches whose path or value contains
	// it. A tree is not re-parsed for a search — Flatten's output is walked —
	// so typing in a search field cannot make the same response parse
	// differently twice.
	Search string
	// MaxRows caps the rows drawn, for a response with forty thousand of
	// them. Zero is no cap.
	MaxRows int
	// Height is the viewport's height, which a scroll area needs before it
	// will scroll.
	Height float32
}

// JsonViewerResult carries a JsonViewer and what was done in it.
type JsonViewerResult struct {
	// Element is the whole viewer.
	Element *ui.Element
	// parsed reports whether the source was JSON at all, so a caller can
	// show a parse error beside it rather than inside it.
	parsed bool
}

// Valid reports whether the source parsed. The error text is shown in the
// viewer itself; this is for a caller that wants to mark the tab it is in.
func (r JsonViewerResult) Valid() bool { return r.parsed }

// JsonViewer is a response body as a tree: one row per value, open and shut
// by the caller's own map.
//
// The rows are Flatten's, and the tree is drawn from them rather than by
// walking the value as it draws. That is what makes a branch open and shut
// without re-parsing: the set of rows is the same on every frame, so a
// branch's open state is a lookup rather than a decision about what to
// descend into.
//
// A body that is not JSON says so, with what the parser said, rather than
// showing an empty tree. An empty tree and a parse failure look identical
// otherwise, and one of them means the response was fine.
func JsonViewer(c *ui.Context, opts JsonViewerOptions) JsonViewerResult {
	if opts.Toggled == nil {
		panic("devtools: JsonViewer needs the *string Toggled writes to")
	}
	if opts.Copied == nil {
		panic("devtools: JsonViewer needs the *string Copied writes to")
	}
	value, err := ParseJSON(opts.Source)
	u := core.Density(c).Unit()
	var r JsonViewerResult
	r.Element = ui.Column(c).FillWidth().Gap(u * 2).
		Label(core.Msg(c, "devtools.json", "JSON"))
	r.Element.Children(func() {
		if err != nil {
			parseError(c, err.Error())
			return
		}
		r.parsed = true
		jsonRows(c, opts, value)
	})
	return r
}

// parseError is what a body that is not JSON says.
func parseError(c *ui.Context, msg string) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	ui.Column(c).FillWidth().Padding(u*2.5, u*3).Radius(theme.ControlRadius).
		Background(k.DangerBg).Gap(u * 0.5).Label(core.Msg(c, "devtools.badJSON", "Not JSON")).
		Children(func() {
			ui.Text(c, core.Msg(c, "devtools.badJSON", "Not JSON")).TextColor(k.Danger).
				Bold().FontSize(core.FontSize(c, theme.RowSize))
			// The parser's own words, because "Not JSON" on its own sends somebody
			// hunting for what they did wrong to the JSON rather than to the body.
			ui.Text(c, msg).TextColor(k.Danger).Font(monoFamily).
				FontSize(core.FontSize(c, theme.MetaSize))
		})
}

// jsonRows draws the flattened tree, skipping the branches that are shut and
// the rows that do not match a search.
func jsonRows(c *ui.Context, opts JsonViewerOptions, value any) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	nodes := FilterNodes(Flatten(value), opts.Search)

	drawn := 0
	ui.Column(c).FillWidth().Gap(u * 0.25).Role(ui.RoleTree).Children(func() {
		for _, node := range nodes {
			if opts.MaxRows > 0 && drawn >= opts.MaxRows {
				break
			}
			// A row under a shut branch is not drawn, and the walk below is
			// what stops there rather than a filter here: a filter would have
			// to know each row's parent, which the path gives it, and would
			// then be a second implementation of the same rule.
			if hiddenByAncestor(node, nodes, opts.Expanded) {
				continue
			}
			drawn++
			jsonRow(c, opts, node)
		}
		if len(nodes) == 0 {
			ui.Text(c, core.Msg(c, "devtools.noMatches", "Nothing matches")).
				TextColor(k.TextFaint).FontSize(core.FontSize(c, theme.MetaSize)).SingleLine()
		}
	})
}

// hiddenByAncestor reports whether a row sits inside a branch that is shut.
//
// The path is the whole of the rule: a row is hidden when any prefix of its
// path is a key in the map that is false. That is a few string operations
// per row and no tree, which is what keeps the tree a list of rows rather
// than a structure this package has to keep in step with the parser's.
func hiddenByAncestor(node JSONNode, nodes []JSONNode, expanded map[string]bool) bool {
	if len(expanded) == 0 {
		return false
	}
	path := node.Path
	for {
		cut := lastDot(path)
		if cut < 0 {
			return false
		}
		parent := path[:cut]
		if open, tracked := expanded[parent]; tracked && !open {
			return true
		}
		path = parent
	}
}

// lastDot is where the last dot of a path is, or -1.
func lastDot(p string) int {
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '.' {
			return i
		}
	}
	return -1
}

// FilterNodes is the rows whose path or preview contains a search.
//
// An empty search returns everything, and the slice is the caller's own when
// nothing is filtered out — so the common case allocates nothing on a
// stream that redraws on every keystroke.
func FilterNodes(nodes []JSONNode, search string) []JSONNode {
	needle := lowerASCII(search)
	if needle == "" {
		return nodes
	}
	out := make([]JSONNode, 0, len(nodes))
	for _, n := range nodes {
		if containsASCII(lowerASCII(n.Path), needle) ||
			containsASCII(lowerASCII(JSONPreview(n.Value, 0)), needle) {
			out = append(out, n)
		}
	}
	return out
}

// jsonRow is one line of the tree: its indent, its key, its type, its value
// and a copy button.
func jsonRow(c *ui.Context, opts JsonViewerOptions, node JSONNode) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	branch := node.Kind == KindObject || node.Kind == KindArray

	row := ui.Row(c).FillWidth().Gap(u).AlignItems(ui.Center).
		Padding(u*0.5, u).Radius(theme.SmallRadius).Role(ui.RoleTreeItem).
		Label(node.Path)
	row.Children(func() {
		// The indent is a spacer rather than a padding: padding on the row
		// would put the tree's left edge inside every row's own background,
		// which makes a hovered row a different colour from where it sits.
		ui.Box(c).Width(unit(c) * float32(node.Depth)).Shrink(0)

		if branch {
			toggle := ui.Box(c).Size(u*4, u*4).Shrink(0).Cursor(ui.CursorPointer).
				Role(ui.RoleButton).Label(branchName(c, node, opts.Expanded)).Children(func() {
				ui.Text(c, chevronFor(isOpen(opts.Expanded, node.Path))).
					TextColor(k.TextMuted).FontSize(core.FontSize(c, theme.CaptionSize))
			})
			if toggle.Clicked() {
				*opts.Toggled = node.Path
			}
		} else {
			ui.Box(c).Size(u*4, u*4).Shrink(0)
		}

		if node.Key != "" {
			ui.Text(c, node.Key).TextColor(k.Text).Shrink(0).
				Font(monoFamily).FontSize(core.FontSize(c, theme.RowSize)).SingleLine()
			ui.Text(c, ":").TextColor(k.TextFaint).Shrink(0).
				Font(monoFamily).FontSize(core.FontSize(c, theme.RowSize))
		}
		ui.Text(c, node.Kind.String()).TextColor(k.TextFaint).Shrink(0).
			FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()

		if !branch {
			ui.Text(c, JSONPreview(node.Value, jsonPreviewWidth)).TextColor(valueInk(k, node)).
				Font(monoFamily).FontSize(core.FontSize(c, theme.RowSize)).Grow(1).SingleLine()
			if input.IconButton(c, glyphCopy, core.Msg(c, "devtools.copy", "Copy"), input.ButtonOptions{}).
				Clicked() {
				*opts.Copied = node.Path
			}
		} else {
			ui.Text(c, countOf(node.Value)).TextColor(k.TextFaint).Shrink(0).
				FontSize(core.FontSize(c, theme.CaptionSize))
			ui.Box(c).Grow(1)
		}
	})
}

// jsonPreviewWidth is how much of a value a row shows before the ellipsis.
const jsonPreviewWidth = 60

// valueInk is a row's value in the tone its kind calls for: strings in the
// body tone, everything else in the muted one. Numbers are not coloured like
// strings on purpose — a response with fifty numbers would come out looking
// like fifty warnings.
func valueInk(k theme.Tokens, node JSONNode) ui.Color {
	if node.Redacted {
		return k.Danger
	}
	if node.Kind == KindString {
		return k.Success
	}
	return k.TextMuted
}

// countOf is what a branch's row says about its size: "3 keys" or "12 items".
// The two words differ because "3 items" for an object's keys is a category
// error a person notices.
func countOf(v any) string {
	switch t := v.(type) {
	case map[string]any:
		if len(t) == 1 {
			return "1 key"
		}
		return itoa(len(t)) + " keys"
	case []any:
		if len(t) == 1 {
			return "1 item"
		}
		return itoa(len(t)) + " items"
	}
	return ""
}

// branchName is what a branch's toggle is called: the action and what it acts
// on. "Expand customer" says what pressing it does; a bare "customer" says
// only where it is, which is already what the row is saying.
func branchName(c *ui.Context, node JSONNode, expanded map[string]bool) string {
	what := node.Key
	if what == "" {
		what = core.Msg(c, "devtools.root", "the response")
	}
	if isOpen(expanded, node.Path) {
		return core.Msg(c, "devtools.collapse", "Collapse ") + what
	}
	return core.Msg(c, "devtools.expand", "Expand ") + what
}

// isOpen reports whether a branch is open, and whether it is tracked at all.
// A branch the caller has never mentioned is closed: the alternative is a
// fully expanded tree by default, which for a real response is a page of
// scroll and tells nobody anything.
func isOpen(expanded map[string]bool, path string) bool {
	return expanded[path]
}

// chevronFor is the mark beside a branch: pointing down when it is open and
// right when it is shut. It is the same mark navigation uses for a closed
// group, so the gesture is the same everywhere.
func chevronFor(open bool) string {
	if open {
		return "⌄"
	}
	return "›"
}

// JsonTreeOptions configure a JsonTree.
type JsonTreeOptions struct {
	// Nodes are the rows, as Flatten produced them. A tree takes rows rather
	// than a value because a caller may have filtered them, redacted them or
	// got them from somewhere other than this package, and a component that
	// insisted on parsing its own would refuse all three.
	Nodes []JSONNode
	// Expanded and Toggled are as JsonViewer's: a map and the path of the
	// branch that was pressed.
	Expanded map[string]bool
	Toggled  *string
	// MaxRows caps the rows drawn; zero is no cap.
	MaxRows int
	// Height is the viewport's height.
	Height float32
}

// JsonTreeResult carries a JsonTree and what was pressed in it.
type JsonTreeResult struct {
	// Element is the whole tree.
	Element *ui.Element
}

// JsonTree is the rows on their own, with no parsing and no toolbar.
//
// It is the same tree JsonViewer draws, split out because the two are used in
// different places: a response viewer wants a field and a search over it,
// while a diff, a log line and a schema view all want the rows and nothing
// else. One drawing of a tree and two ways of reaching it, rather than two
// drawings that drift.
func JsonTree(c *ui.Context, opts JsonTreeOptions) JsonTreeResult {
	if opts.Toggled == nil {
		panic("devtools: JsonTree needs the *string Toggled writes to")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	var r JsonTreeResult

	tree := ui.Column(c).FillWidth().Gap(u * 0.25).Role(ui.RoleTree).
		Label(core.Msg(c, "devtools.json", "JSON"))
	tree.Children(func() {
		shown := 0
		for _, node := range opts.Nodes {
			if opts.MaxRows > 0 && shown >= opts.MaxRows {
				break
			}
			if hiddenByAncestor(node, opts.Nodes, opts.Expanded) {
				continue
			}
			shown++
			treeRow(c, opts, node)
		}
		if len(opts.Nodes) == 0 {
			ui.Text(c, core.Msg(c, "devtools.empty", "Nothing to show")).
				TextColor(k.TextFaint).FontSize(core.FontSize(c, theme.MetaSize)).SingleLine()
		}
	})
	r.Element = tree
	return r
}

// treeRow is one line of a bare tree: no copy button, no path label.
func treeRow(c *ui.Context, opts JsonTreeOptions, node JSONNode) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	branch := node.Kind == KindObject || node.Kind == KindArray

	ui.Row(c).FillWidth().Gap(u).AlignItems(ui.Center).
		Padding(u*0.5, u).Radius(theme.SmallRadius).
		Label(node.Path).Children(func() {
		ui.Box(c).Width(unit(c) * float32(node.Depth)).Shrink(0)
		if branch {
			toggle := ui.Box(c).Size(u*4, u*4).Shrink(0).Cursor(ui.CursorPointer).
				Role(ui.RoleButton).Label(branchName(c, node, opts.Expanded)).Children(func() {
				ui.Text(c, chevronFor(isOpen(opts.Expanded, node.Path))).
					TextColor(k.TextMuted).FontSize(core.FontSize(c, theme.CaptionSize))
			})
			if toggle.Clicked() {
				*opts.Toggled = node.Path
			}
		} else {
			ui.Box(c).Size(u*4, u*4).Shrink(0)
		}
		if node.Key != "" {
			ui.Text(c, node.Key).TextColor(k.Text).Shrink(0).
				Font(monoFamily).FontSize(core.FontSize(c, theme.RowSize)).SingleLine()
		}
		ui.Text(c, JSONPreview(node.Value, jsonPreviewWidth)).TextColor(k.TextMuted).
			Font(monoFamily).FontSize(core.FontSize(c, theme.RowSize)).Grow(1).SingleLine()
	})
}

// SchemaTreeOptions configure a SchemaTree.
type SchemaTreeOptions struct {
	// Schema is the shape being described.
	Schema SchemaNode
	// Values are the rows of the actual response, which the tree shows beside
	// the schema so that a key with no value and a value with no key are both
	// visible. Nil draws the schema alone.
	Values []JSONNode
	// Height is the viewport's height.
	Height float32
}

// SchemaTreeResult carries a SchemaTree.
type SchemaTreeResult struct {
	// Element is the whole thing.
	Element *ui.Element
}

// SchemaTree is what a response was supposed to look like, beside what it
// actually is.
//
// The missing marks are the point of it. A key the schema requires and the
// response does not have is drawn with a warning mark, and a key the response
// has and the schema never mentioned is drawn with a muted one — both beside
// the name rather than in a list at the bottom, because the question a person
// is asking is about each key as they read it.
func SchemaTree(c *ui.Context, opts SchemaTreeOptions) SchemaTreeResult {
	u := core.Density(c).Unit()
	have := make(map[string]bool, len(opts.Values))
	for _, v := range opts.Values {
		have[v.Path] = true
	}
	required := requiredPaths(opts.Schema)

	var r SchemaTreeResult
	tree := ui.Column(c).FillWidth().Gap(u * 0.25).Role(ui.RoleTree).
		Label(core.Msg(c, "devtools.schema", "Schema"))
	tree.Children(func() {
		for _, node := range SchemaRows(opts.Schema) {
			schemaRow(c, node, have[leafOf(node.Path)], required[leafOf(node.Path)])
		}
	})
	r.Element = tree
	return r
}

// schemaRow is one line of the schema tree.
func schemaRow(c *ui.Context, node JSONNode, present, isRequired bool) {
	k, u := core.Tokens(c), core.Density(c).Unit()

	ui.Row(c).FillWidth().Gap(u).AlignItems(ui.Start).
		Padding(u*0.5, u).Role(ui.RoleTreeItem).Label(node.Path).Children(func() {
		ui.Box(c).Width(unit(c) * float32(node.Depth)).Shrink(0)
		ink := k.Text
		switch {
		case isRequired && !present:
			// Required and missing: the one combination that is a bug rather
			// than a surprise, and the only one that gets the danger colour.
			ink = k.Danger
		case !present:
			ink = k.TextFaint
		}
		if node.Key != "" {
			ui.Text(c, node.Key).TextColor(ink).Shrink(0).
				Font(monoFamily).FontSize(core.FontSize(c, theme.RowSize)).SingleLine()
		}
		if s, _ := node.Value.(string); s != "" {
			ui.Text(c, s).TextColor(k.TextMuted).Grow(1).
				FontSize(core.FontSize(c, theme.MetaSize)).SingleLine()
		}
		switch {
		case isRequired && !present:
			ui.Text(c, core.Msg(c, "devtools.missing", "missing")).TextColor(k.Danger).
				Shrink(0).FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
		case !present:
			ui.Text(c, core.Msg(c, "devtools.absent", "absent")).TextColor(k.TextFaint).
				Shrink(0).FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
		}
	})
}

// requiredPaths is every leaf the schema says must be there, keyed by its own
// last step so that the comparison against a response's keys is one of names.
func requiredPaths(s SchemaNode) map[string]bool {
	out := map[string]bool{}
	collectRequired(s, out)
	return out
}

// collectRequired walks a schema recording what it requires.
func collectRequired(s SchemaNode, out map[string]bool) {
	for _, name := range s.Required {
		out[name] = true
	}
	for _, sub := range s.Properties {
		collectRequired(sub, out)
	}
	if s.Items != nil {
		collectRequired(*s.Items, out)
	}
}

// leafOf is the last step of a path, which is the key a schema names.
func leafOf(p string) string {
	if i := lastDot(p); i >= 0 {
		return p[i+1:]
	}
	return p
}

// JsonEditorOptions configure a JsonEditor.
type JsonEditorOptions struct {
	// Source is the text being edited, written as it is typed.
	Source *string
	// Valid reports whether what is in the field is JSON, so the border can
	// change without this component parsing twice.
	Valid bool
	// Format asks for the text to be rewritten with one key per line and two
	// spaces of indent. It is an ask rather than something done on the way
	// past, because reformatting somebody's JSON while they are typing in it
	// moves the caret.
	Format func()
	// Minify asks for the same text with no whitespace at all, for the moment
	// somebody needs to paste it somewhere that does not like newlines.
	Minify func()
	// Error is what the parser said, when Valid is false. Empty draws no
	// message, for a field that is empty rather than wrong.
	Error string
	// Height is the field's height in lines; zero is eight.
	Height int
}

// JsonEditorResult carries a JsonEditor.
type JsonEditorResult struct {
	// Element is the whole editor.
	Element *ui.Element
}

// JsonEditor is a body of JSON as text that can be edited.
//
// It is input.TextArea and not a text editor of this package's own, because
// undo, selection, the caret and what the arrow keys do are all things the
// control already gets right and a re-implementation gets wrong.
//
// The validity mark is shown rather than enforced: a viewer that refuses to
// show text because it does not parse is useless, because half of what people
// paste into one is half-finished.
func JsonEditor(c *ui.Context, opts JsonEditorOptions) JsonEditorResult {
	if opts.Source == nil {
		panic("devtools: JsonEditor needs the *string Source writes to; it keeps " +
			"no text of its own")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	lines := opts.Height
	if lines <= 0 {
		lines = 8
	}

	var r JsonEditorResult
	r.Element = ui.Column(c).FillWidth().Gap(u * 2).
		Label(core.Msg(c, "devtools.editor", "Editor")).Children(func() {
		input.TextArea(c, opts.Source, input.TextAreaOptions{
			Label:       core.Msg(c, "devtools.editor", "Editor"),
			Placeholder: core.Msg(c, "devtools.editorHint", `{ "paste": "a response" }`),
			Error:       opts.Error,
			Lines:       lines,
		})

		// The mark, in words rather than in colour: a border that turns red
		// is invisible to somebody who cannot see red, and what is valid JSON
		// is exactly the thing they most need to be told.
		mark := core.Msg(c, "devtools.valid", "Valid JSON")
		ink := k.Success
		if !opts.Valid {
			mark = core.Msg(c, "devtools.invalid", "Not valid JSON")
			ink = k.Danger
		}
		ui.Text(c, mark).TextColor(ink).FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()

		ui.Row(c).FillWidth().Gap(u * 2).Children(func() {
			ui.Box(c).Grow(1)
			if opts.Format != nil && input.Button(c, core.Msg(c, "devtools.format", "Format"),
				input.ButtonOptions{}).Clicked() {
				opts.Format()
			}
			if opts.Minify != nil && input.Button(c, core.Msg(c, "devtools.minify", "Minify"),
				input.ButtonOptions{}).Clicked() {
				opts.Minify()
			}
		})
	})
	return r
}

// FormatJSON rewrites JSON text with one key per line and two spaces of
// indent. It is a function because "Format" is a button anybody can press and
// the answer has to be the same everywhere it is offered.
//
// Text that does not parse is returned unchanged. Returning an error here
// would force every caller to handle it before showing the user a button
// whose only failure mode is a parse they have already been shown.
func FormatJSON(src string) string {
	value, err := ParseJSON(src)
	if err != nil {
		return src
	}
	out, err := MarshalIndent(value)
	if err != nil {
		return src
	}
	return out
}

// MinifyJSON rewrites JSON text with no whitespace at all, for pasting
// somewhere that does not like newlines. As with FormatJSON, text that does
// not parse comes back unchanged.
func MinifyJSON(src string) string {
	value, err := ParseJSON(src)
	if err != nil {
		return src
	}
	out, err := marshalCompact(value)
	if err != nil {
		return src
	}
	return out
}

// ValidJSON reports whether the text parses, for a component that needs the
// answer as a bool and nothing else.
func ValidJSON(src string) bool {
	_, err := ParseJSON(src)
	return err == nil
}

// titleCase is a lower-case word as a heading. It is here because the two tab
// labels are words this package chooses rather than messages it ships, and
// strings.Title is deprecated for exactly the reason this exists — it
// upper-cases letters inside words.
func titleCase(s string) string {
	if s == "" {
		return s
	}
	if s[0] >= 'a' && s[0] <= 'z' {
		return string(s[0]-('a'-'A')) + s[1:]
	}
	return s
}
