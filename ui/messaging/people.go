package messaging

import (
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/data"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/overlay"
	"github.com/HycJack/MintUI/ui/theme"
)

// ── members ────────────────────────────────────────────────────────────────

// Member is somebody in a channel or a group.
type Member struct {
	// ID identifies them and is required, for the same reason a channel's is:
	// the list is reordered by the caller and the rows follow their records.
	ID string
	// Name is what the reader sees, and is required: a member row with no
	// name is a presence dot with nothing to attach it to.
	Name string
	// Handle is their @name.
	Handle string
	// Role is what they do in the channel.
	Role string
	// Presence is how reachable they are.
	Presence Presence
	// Since is when they last were seen, as the caller's string.
	Since string
	// LocalTime is what the clock says where they are.
	LocalTime string
	// Title is their job, for a card over an avatar.
	Title string
}

// MemberRole is one of the columns a channel's member list sorts by. It is a
// string rather than an index because the caller's sort state is a string and
// the two cannot name the same column two ways.
type MemberRole string

const (
	// RoleByName sorts alphabetically; it is the default order, because a
	// list of two hundred people is a list people are looked up in.
	RoleByName MemberRole = "name"
	// RoleByPresence puts the people who are here first, which is what a
	// channel is checked for.
	RoleByPresence MemberRole = "presence"
	// RoleByRole puts the channel's own roles first.
	RoleByRole MemberRole = "role"
)

// memberColumns are the list's columns. The widths are fixed because a member
// row's columns carry text that must not wrap: a handle truncated to "@sa" is
// a handle nobody can reply to.
var memberColumns = []data.Column{
	{Title: "Member", ID: "name", Share: 1},
	{Title: "Role", ID: "role", Width: 128},
	{Title: "Local time", ID: "time", Width: 108, Align: ui.End},
}

// MemberListOptions configure a MemberList.
type MemberListOptions struct {
	// Members are the caller's, in the order they should be shown.
	Members []Member
	// Selected is the member the keys move from, as an index; -1 for none.
	Selected *int
	// Sort is the column the rows are ordered by, in the caller's state. The
	// table asks for a new one and the caller reorders Members with it.
	Sort *data.Sort
	// State and Scroll keep the table's place between frames.
	State  *ui.ListState
	Scroll *ui.ScrollState
	// Height is the table's height. It is required: a member list with no
	// height grows to fit every member, which for a channel of two hundred is
	// a page nobody can scroll past the first screen of.
	Height float32
	// Query filters by name, handle or role, and is the caller's because the
	// search field beside it is.
	Query *string
	// ShowRoles and ShowTimes are the two columns worth leaving out in a
	// narrow panel: a popover has room for the name and the dot, and a table
	// of one column is a list.
	ShowRoles, ShowTimes bool
	// Title heads the panel; empty takes the library's "Members".
	Title string
}

// MemberListResult carries a MemberList.
type MemberListResult struct {
	// Element is the list.
	Element *ui.Element
	// shown is how many rows survived the query.
	shown int
	// opened is the member whose card was opened this frame, or -1.
	opened int
}

// Shown is how many members passed the query, which is what an empty state has
// to be able to distinguish from a channel with nobody in it.
func (r MemberListResult) Shown() int { return r.shown }

// Opened is the member whose card was opened this frame, as an index into the
// members it was given, or -1. The card itself is the caller's layer — opened
// from a Popover over the row — and only the caller knows which one is
// showing.
func (r MemberListResult) Opened() int { return r.opened }

// MemberList is the people in a channel: their names, their presence, what
// they do and what time it is where they are.
//
// It is [data.DataTable] because a channel's members are records in columns —
// there is one row per person and four things to say about each — and a list
// of one-line rows would be four lists stitched together with the relationship
// between a name and its dot held by their being in the same row.
func MemberList(c *ui.Context, opts MemberListOptions) MemberListResult {
	if opts.Selected == nil {
		panic("messaging: MemberList needs a Selected member to point at; it owns no list " +
			"of its own")
	}
	if opts.Height <= 0 {
		panic("messaging: MemberList needs a Height; a list with no height grows to fit every " +
			"member rather than scrolling")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	title := opts.Title
	if title == "" {
		title = core.Msg(c, "messaging.members", core.Def("Members"))
	}

	shown := make([]int, 0, len(opts.Members))
	for i, m := range opts.Members {
		if memberMatches(m, opts.Query) {
			shown = append(shown, i)
		}
	}

	cols := make([]data.Column, 0, len(memberColumns))
	for _, col := range memberColumns {
		switch col.ID {
		case "role":
			if !opts.ShowRoles {
				continue
			}
		case "time":
			if !opts.ShowTimes {
				continue
			}
		}
		cols = append(cols, col)
	}

	res := MemberListResult{shown: len(shown), opened: -1}

	res.Element = layout.Container(c, layout.ContainerOptions{
		Surface: true, Border: true, Radius: theme.CardRadius, Gap: u, Pad: u * 1.5,
	}, func() {
		ui.Text(c, title).TextColor(k.Text).
			FontSize(core.FontSize(c, theme.RowSize)).Bold()
		table := data.DataTable(c, data.DataTableOptions{
			Columns: cols,
			Rows:    len(shown),
			Sort:    opts.Sort,
			Cell: func(row, col int) {
				m := opts.Members[shown[row]]
				switch cols[col].ID {
				case "name":
					// A Row, not a Box: a Box is a column, and the three
					// parts of a member row belong beside each other — the
					// name is looked up by the person next to it, not above
					// it.
					ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.25).
						Role(ui.RoleNone).Children(func() {
						display.Avatar(c, m.Name).Grow(0)
						// The box is a column, so the name claims no height:
						// grown into height it does not have it lays out at
						// zero tall and the handle prints straight through it.
						ui.Box(c).FillWidth().Children(func() {
							ui.Text(c, m.Name).TextColor(k.Text).SingleLine().
								FontSize(core.FontSize(c, theme.RowSize))
							if m.Handle != "" {
								ui.Text(c, m.Handle).TextColor(k.TextFaint).
									FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
							}
						})
						OnlineStatus(c, OnlineStatusOptions{
							Presence: m.Presence, Name: m.Name, Size: u * 1.75,
						}).Grow(0)
					})
				case "role":
					if m.Role == "" {
						ui.Text(c, "—").TextColor(k.TextFaint).
							FontSize(core.FontSize(c, theme.RowSize))
						return
					}
					ui.Text(c, m.Role).TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.RowSize)).MaxLines(1)
				case "time":
					if m.LocalTime == "" {
						ui.Text(c, "—").TextColor(k.TextFaint).
							FontSize(core.FontSize(c, theme.RowSize))
						return
					}
					ui.Text(c, m.LocalTime).TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.RowSize))
				}
			},
			CellLabel: func(row, col int) string { return opts.Members[shown[row]].Name },
			Key:       func(row int) any { return opts.Members[shown[row]].ID },
			Label:     func(row int) string { return opts.Members[shown[row]].Name },
			Selected:  opts.Selected,
			State:     opts.State,
			Scroll:    opts.Scroll,
			Height:    opts.Height,
			Empty: func() {
				ui.Text(c, emptyMembers(opts.Query)).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.BodySize))
			},
		})
		// Built in here rather than beside the panel above: an element
		// belongs to whatever container is current where it is made, so
		// one made out there lands above the panel and the list comes out
		// with its own folder heading sitting underneath it.
		table.Element.FillWidth()
	})
	return res
}

// memberMatches is the query rule: name, handle or role. The role is included
// because "who is in here" is very often asked as "who is on the rota", and a
// search that only knew names would answer the second question with nothing.
func memberMatches(m Member, query *string) bool {
	if query == nil || *query == "" {
		return true
	}
	needle := lower(*query)
	return contains(lower(m.Name), needle) ||
		contains(lower(m.Handle), needle) ||
		contains(lower(m.Role), needle)
}

func emptyMembers(query *string) string {
	if query != nil && *query != "" {
		return "Nobody here matches that"
	}
	return "Nobody else in this channel"
}

// ── emoji ──────────────────────────────────────────────────────────────────

// EmojiGroup is one page of the picker: a category's worth.
type EmojiGroup struct {
	// Name is what the category is called, and is required: a page of faces
	// with no name is a page a screen reader reads as forty identical glyphs.
	Name string
	// Emoji are the entries, each the glyph itself and the name a reader
	// would call it. The name is not decoration — it is what gets inserted
	// into the message when the glyph cannot be, and what is announced.
	Emoji []Emoji
	// Search are the words that should find this group, so a picker with no
	// search of its own can be handed the caller's.
	Search []string
}

// Emoji is one entry: the glyph and its name.
type Emoji struct {
	// Glyph is the character itself.
	Glyph string
	// Name is what it is called and what is written into the message when the
	// application cannot insert the glyph. Required for the same reason.
	Name string
}

// EmojiPickerOptions configure an EmojiPicker.
type EmojiPickerOptions struct {
	// Anchor is the button the panel hangs off, and is required.
	Anchor *ui.Element
	// Open is the *bool the panel opens and closes with, and is required.
	Open *bool
	// Groups are the caller's categories, in the order the tabs should run.
	Groups []EmojiGroup
	// Picked is the name chosen this frame, or "". It is reported rather than
	// written, because inserting an emoji is the composer's business and
	// only the composer knows where the caret is.
	Picked string
	// Query is the search field's value, the caller's. A picker with a search
	// that cannot search is a grid of faces and nothing else.
	Query *string
	// Recent is what the picker shows first when there is room for a second
	// grid — the emoji this reader uses are the ones they need again.
	Recent []Emoji
	// Label names the panel; empty takes the library's "Emoji".
	Label string
	// Width is the panel's width; zero lets it fit the grid.
	Width float32
}

// EmojiPickerResult carries an EmojiPicker.
type EmojiPickerResult struct {
	// Element is the anchor, with the panel hung off it.
	Element *ui.Element
	// picked is the emoji chosen this frame, or "".
	picked string
	// searches are the groups a query matched, which is the whole answer to
	// "what does this search find".
	searches int
}

// Picked is the emoji chosen this frame, by its name. A name rather than the
// glyph because the caller's message may be going somewhere that cannot hold
// the character, and a name that can be turned into a glyph later is a name
// that still works.
func (r EmojiPickerResult) Picked() string { return r.picked }

// Searches is how many groups the current query matched. Zero with a non-empty
// query is the state that says "nothing matches", and it is what the panel's
// own empty state is built on.
func (r EmojiPickerResult) Searches() int { return r.searches }

// EmojiPicker is the panel of emoji beside a composer: a search, a page of
// categories, and the entry chosen.
//
// It writes nothing into a message. The composer owns the caret and the draft,
// and inserting a character into the middle of somebody's half-typed word is
// exactly the thing a component with no idea where the caret is gets wrong —
// so the picker reports the entry and the composer puts it in.
func EmojiPicker(c *ui.Context, opts EmojiPickerOptions) EmojiPickerResult {
	if opts.Anchor == nil {
		panic("messaging: EmojiPicker needs the Anchor it hangs off")
	}
	if opts.Open == nil {
		panic("messaging: EmojiPicker needs the *bool it opens and closes with")
	}
	if len(opts.Groups) == 0 {
		panic("messaging: EmojiPicker needs at least one Group; a picker with no emoji is an " +
			"empty panel")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	label := opts.Label
	if label == "" {
		label = core.Msg(c, "messaging.emoji", core.Def("Emoji"))
	}

	matches := searchGroups(opts.Groups, opts.Query)
	res := EmojiPickerResult{searches: len(matches)}

	overlay.Popover(c, opts.Anchor, opts.Open, overlay.PopoverOptions{
		Modal: false,
		Width: opts.Width,
		Label: label,
		Body: func() {
			if opts.Query != nil {
				input.SearchInput(c, opts.Query, input.SearchInputOptions{
					Label:       "Search emoji",
					Placeholder: "Search emoji",
					AutoFocus:   true,
				})
			}
			if len(matches) == 0 {
				ui.Text(c, "No emoji matches that").TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.BodySize))
				return
			}
			for _, g := range matches {
				if g.Name != "" {
					ui.Text(c, g.Name).TextColor(k.TextFaint).
						FontSize(core.FontSize(c, theme.CaptionSize))
				}
				emojiGrid(c, g.Emoji, &res.picked, u)
			}
		},
	})

	// Picked accumulates rather than being assigned: the view runs up to three
	// times per frame and the settled pass has no press, so an assignment
	// would replace the answer with the empty one.
	if opts.Picked != "" {
		res.picked = opts.Picked
	}
	return res
}

// emojiGrid is one page of the picker: the entries in a wrapping grid, each
// pressable and each named.
//
// picked is written to rather than returned, because the grid is drawn inside
// the panel's body closure and a value returned from there would be one frame
// behind the panel it belongs to.
func emojiGrid(c *ui.Context, entries []Emoji, picked *string, u float32) {
	k := core.Tokens(c)
	const columns = 8
	side := u * 7
	rows := (len(entries) + columns - 1) / columns
	if rows <= 0 {
		return
	}
	ui.Column(c).FillWidth().Gap(u * 0.5).Children(func() {
		for r := 0; r < rows; r++ {
			ui.Row(c).FillWidth().Gap(u * 0.5).Children(func() {
				for col := 0; col < columns; col++ {
					i := r*columns + col
					if i >= len(entries) {
						// The row is padded with empty slots so the grid stays
						// rectangular. A ragged last row reads as a different
						// number of emoji per row, which makes the panel feel
						// like it has been cut short.
						ui.Box(c).Grow(1).Height(side).Shrink(0).Role(ui.RoleNone)
						continue
					}
					e := entries[i]
					cell := ui.Box(c).Grow(1).Height(side).Shrink(0).
						Radius(theme.SmallRadius).Center().
						Label(e.Name).Role(ui.RoleNone).Cursor(ui.CursorPointer)
					if cell.Hovered() {
						cell.Background(k.SurfaceHover)
					}
					if cell.Clicked() {
						*picked = e.Name
					}
					cell.Children(func() {
						ui.Text(c, e.Glyph).FontSize(core.FontSize(c, theme.BodySize))
					})
				}
			})
		}
	})
}

// searchGroups is the groups a query matches: everything when there is no
// query, and otherwise every group holding an entry or a word the query names.
//
// It searches the words rather than the glyphs, because nobody types a
// character to search for it — they type "fire" — and a picker that searched
// the glyphs would find fire when somebody pasted one and nothing otherwise.
func searchGroups(groups []EmojiGroup, query *string) []EmojiGroup {
	if query == nil || strings.TrimSpace(*query) == "" {
		return groups
	}
	needle := lower(strings.TrimSpace(*query))
	var out []EmojiGroup
	for _, g := range groups {
		hit := EmojiGroup{Name: g.Name}
		for _, word := range g.Search {
			if contains(lower(word), needle) {
				// A category word matching does not make every entry in it
				// match; it makes the category worth showing, and the entries
				// are then filtered by their own names.
				hit.Search = nil
				break
			}
		}
		for _, e := range g.Emoji {
			if contains(lower(e.Name), needle) || contains(lower(e.Glyph), needle) {
				hit.Emoji = append(hit.Emoji, e)
			}
		}
		if len(hit.Emoji) > 0 || contains(lower(g.Name), needle) {
			out = append(out, hit)
		}
	}
	return out
}

// Emoji is the common set a channel gets when its caller has none: enough that
// a reader can say the six things a message is mostly made of. It is exported
// because a caller with no emoji budget of its own needs something to draw,
// and an empty picker is worse than a small one.
func DefaultGroups() []EmojiGroup {
	return []EmojiGroup{
		{Name: "Reactions", Search: []string{"like", "heart", "yes", "no", "ok"}, Emoji: []Emoji{
			{Glyph: "👍", Name: "thumbs up"},
			{Glyph: "👎", Name: "thumbs down"},
			{Glyph: "❤️", Name: "red heart"},
			{Glyph: "🎉", Name: "tada"},
			{Glyph: "👀", Name: "eyes"},
			{Glyph: "🙏", Name: "thank you"},
		}},
		{Name: "Faces", Search: []string{"smile", "laugh", "cry"}, Emoji: []Emoji{
			{Glyph: "😀", Name: "grinning"},
			{Glyph: "😅", Name: "grinning with sweat"},
			{Glyph: "🤔", Name: "thinking"},
			{Glyph: "😅", Name: "sweat smile"},
			{Glyph: "😭", Name: "loudly crying"},
			{Glyph: "🙃", Name: "upside down"},
		}},
		{Name: "Status", Search: []string{"check", "cross", "warning"}, Emoji: []Emoji{
			{Glyph: "✅", Name: "check"},
			{Glyph: "❌", Name: "cross"},
			{Glyph: "⚠️", Name: "warning"},
			{Glyph: "🔥", Name: "on fire"},
			{Glyph: "🚀", Name: "launch"},
			{Glyph: "📌", Name: "pin"},
		}},
	}
}
