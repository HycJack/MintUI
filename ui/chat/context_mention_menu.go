package chat

import (
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// ContextItem is one thing the menu can bring into the conversation.
type ContextItem struct {
	// Name is what the row says and what the query is matched against.
	// Required: a row with no name is a tag with nothing beside it.
	Name string
	// Kind is what the thing is — "file", "project", "doc" — drawn as the
	// small tag before the name. Empty draws no tag, which is what a kind
	// that says nothing is.
	Kind string
	// Meta is the line under the name — a path, a location, a size.
	// Empty draws no second line.
	Meta string
}

// ContextMentionMenuOptions configure a ContextMentionMenu.
type ContextMentionMenuOptions struct {
	// Items are the things on offer, in the order they are shown.
	Items []ContextItem
	// Query is what has been typed after the "@", when the menu filters as
	// the reader types. nil shows everything: the menu is also used whole,
	// where the caller opens it and the reader chooses from all of it.
	// Filtering is by Name, ignoring case, for the reason the command menu
	// filters the same way — the name is the one line every kind of item
	// has.
	Query *string
	// MaxHeight caps the panel; zero takes six rows.
	MaxHeight float32
	// Label names the menu; empty takes the library's "Context".
	Label string
}

// ContextMentionMenuResult carries a ContextMentionMenu and the item chosen.
type ContextMentionMenuResult struct {
	// Element is the menu, nil when nothing matches.
	Element *ui.Element
	// picked is -1 until a row is pressed.
	picked int
}

// Picked is the index of the row pressed this frame, counted in the rows as
// drawn — the filtered set when a query is in — -1 for none. The caller
// holds the same slice and applies the same filter to recover the item,
// which is the same arrangement the command palette makes: the menu reports
// where the reader pointed, and the caller decides what that row refers to.
func (r ContextMentionMenuResult) Picked() int { return r.picked }

// ContextMentionMenu is the list of things an "@" can pull into a
// conversation: files, projects, documents.
//
// It differs from MentionMenu in what a row is. The mention menu's rows are
// people and their locations — one line each, typed as "name: where" —
// because a mention has to say who. A context item is a thing, and a thing
// carries a kind, which is why each row here wears a small tag: the kind is
// what lets a reader tell "README.md the file" from "README.md the project"
// at the size the menu is, and a second line would have to do the same work
// in a place with no room for it.
//
// Like the two composer menus, it draws nothing when nothing matches rather
// than an empty panel: a panel that appears with no rows says the feature is
// broken, where a panel that does not appear says the typing matched
// nothing.
func ContextMentionMenu(c *ui.Context, opts ContextMentionMenuOptions) ContextMentionMenuResult {
	for i, item := range opts.Items {
		if item.Name == "" {
			panic("chat: ContextMentionMenu item " + internal.Commas(i) +
				" needs a Name; a row with no name is a tag with nothing beside it")
		}
	}

	q := ""
	if opts.Query != nil {
		q = strings.ToLower(strings.TrimSpace(*opts.Query))
	}
	shown := opts.Items[:0:0]
	for _, item := range opts.Items {
		if q == "" || strings.Contains(strings.ToLower(item.Name), q) {
			shown = append(shown, item)
		}
	}

	var res ContextMentionMenuResult
	res.picked = -1
	if len(shown) == 0 {
		return res
	}

	height := opts.MaxHeight
	if height <= 0 {
		height = 6 * commandRowHeight(c)
	}
	name := opts.Label
	if name == "" {
		name = core.Msg(c, "chat.context", core.Def("Context"))
	}

	k, u := core.Tokens(c), core.Density(c).Unit()
	panel := func() {
		ui.Column(c).FillWidth().Gap(u * 0.25).Label(name).Children(func() {
			for i, item := range shown {
				index := i
				row := ui.ButtonBase(c).FillWidth().Justify(ui.Start).AlignItems(ui.Center).
					Gap(u*1.5).Padding(u*0.75, u*1.5).
					Radius(theme.SmallRadius).Label(item.Name)
				if row.Hovered() {
					row.Background(k.Surface)
				}
				if row.Clicked() {
					res.picked = index
				}
				row.Children(func() {
					// The kind is a tag and not a word because it is a
					// category: drawn with its own background it reads as a
					// label on the row, and a bare word would read as part of
					// the name.
					if item.Kind != "" {
						ui.Box(c).Padding(u*0.25, u*1.25).Radius(theme.PillRadius).
							Background(k.Surface).Shrink(0).Children(func() {
							ui.Text(c, item.Kind).TextColor(k.TextMuted).
								FontSize(core.FontSize(c, theme.CaptionSize))
						})
					}
					ui.Column(c).Grow(1).AlignItems(ui.Start).Children(func() {
						ui.Text(c, item.Name).TextColor(k.Text).
							FontSize(core.FontSize(c, theme.RowSize)).SingleLine()
						if item.Meta != "" {
							ui.Text(c, item.Meta).TextColor(k.TextMuted).
								FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
						}
					})
				})
			}
		})
	}

	// The scroll is built here rather than by the caller, for the reason the
	// composer's menus build theirs: a child is whatever was being built when
	// it was made.
	scrolled := ui.Scroll(c).MaxHeight(height).FillWidth()
	scrolled.Children(panel)
	res.Element = scrolled
	return res
}
