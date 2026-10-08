package chat

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// KnowledgeEntry is one thing the project knows, as the panel shows it.
type KnowledgeEntry struct {
	// Title is what the entry is called. Required: the row is read by its
	// title, and a row with no title is a kind tag floating.
	Title string
	// Kind is what the entry is — "doc", "note", "decision" — drawn as the
	// small tag before the title, for the reason ContextMentionMenu draws
	// kinds: the category is what tells two entries apart at the size the
	// panel is.
	Kind string
	// Meta is the line under the title — a date, a path, a source.
	// Empty draws no second line.
	Meta string
}

// ProjectKnowledgePanelOptions configure a ProjectKnowledgePanel.
type ProjectKnowledgePanelOptions struct {
	// Project is the project the knowledge is about. Required: a panel of
	// entries without the project they are about is a shelf with no name on
	// it, and the entries belong to somebody else.
	Project string
	// Entries are the rows, in the order they are shown. Empty draws the
	// panel's empty line rather than an empty column: the difference is the
	// difference between "nothing here yet" and a panel that failed.
	Entries []KnowledgeEntry
}

// ProjectKnowledgePanel is the side panel of what a project knows: the
// documents, notes and decisions a conversation can draw on.
//
// It is a header over a list and nothing more — the header carries the two
// facts the panel is about, the project and how many entries it holds, and
// the list is the panel. It does not open the entries: what an entry means
// when chosen — opened, cited, attached — belongs to the caller, for the
// reason the card's Open button reports and does not open.
func ProjectKnowledgePanel(c *ui.Context, opts ProjectKnowledgePanelOptions) *ui.Element {
	if opts.Project == "" {
		panic("chat: ProjectKnowledgePanel needs a Project; a panel of entries with no " +
			"project is a shelf with no name on it")
	}
	for i, e := range opts.Entries {
		if e.Title == "" {
			panic("chat: ProjectKnowledgePanel entry " + internal.Commas(i) +
				" needs a Title; a row with no title is a kind tag floating")
		}
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	n := internal.Commas(len(opts.Entries))
	unit := core.Msg(c, "chat.knowledge.entries", core.Def("entries"))
	if len(opts.Entries) == 1 {
		unit = core.Msg(c, "chat.knowledge.entry", core.Def("entry"))
	}

	panel := ui.Column(c).FillWidth().Gap(u*1.5).
		Radius(theme.CardRadius).Background(k.Surface).
		Border(theme.BorderWidth, k.Border).Padding(u * 2).Label(opts.Project)
	return panel.Children(func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.5).Children(func() {
			ui.Text(c, opts.Project).TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).Bold().SingleLine().Grow(1)
			ui.Box(c).Padding(u*0.25, u*1.25).Radius(theme.PillRadius).
				Background(k.Background).Shrink(0).Children(func() {
				ui.Text(c, n+" "+unit).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize))
			})
		})
		if len(opts.Entries) == 0 {
			ui.Text(c, core.Msg(c, "chat.knowledge.none",
				core.Def("No entries yet"))).
				TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.MetaSize))
			return
		}
		ui.Column(c).FillWidth().Gap(u * 0.25).Children(func() {
			for _, e := range opts.Entries {
				row := ui.Column(c).FillWidth().Gap(u*0.25).
					Padding(u*0.75, u*1.5).Radius(theme.SmallRadius).Label(e.Title)
				if row.Hovered() {
					row.Background(k.SurfaceHover)
				}
				row.Children(func() {
					ui.Row(c).AlignItems(ui.Center).Gap(u * 1.5).Children(func() {
						if e.Kind != "" {
							ui.Box(c).Padding(u*0.25, u*1.25).
								Radius(theme.PillRadius).Background(k.Background).
								Shrink(0).Children(func() {
								ui.Text(c, e.Kind).TextColor(k.TextMuted).
									FontSize(core.FontSize(c, theme.CaptionSize))
							})
						}
						ui.Text(c, e.Title).TextColor(k.Text).
							FontSize(core.FontSize(c, theme.RowSize)).SingleLine()
					})
					if e.Meta != "" {
						ui.Text(c, e.Meta).TextColor(k.TextMuted).
							FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
					}
				})
			}
		})
	})
}
