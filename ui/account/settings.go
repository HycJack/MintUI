package account

import (
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/overlay"
	"github.com/HycJack/MintUI/ui/theme"
)

// SettingsLayoutOptions configure a SettingsLayout.
type SettingsLayoutOptions struct {
	// Sections are the rows down the left, in order. The one whose ID is in
	// Selected is the one drawn on the right.
	Sections []Setting
	// Selected is the open section's ID. It is the caller's because which
	// section is open is part of where the window is: a deep link names it,
	// and a window restored from last session restores it.
	Selected *string
	// Chosen writes the ID of the section that was pressed, empty when none
	// was. It is separate from Selected so that a caller can decide whether
	// to follow the press — a menu bar's settings opens a window and wants
	// to know which section was asked for, while the settings window itself
	// wants to go there.
	Chosen *string
	// Query is the search text, kept here so the search box and the section
	// list agree about what is being looked for.
	Query *string
	// Searchable draws the search field above the sections. Nil draws none,
	// which is right for a settings pane of three rows where a search box is
	// a control that cannot do anything useful.
	Searchable bool
	// Section is the right-hand pane's content, built for whichever section
	// is selected.
	Section func(id string)
	// SidebarWidth is the left column's width; zero is the library's own
	// rail-adjacent width, which is deliberately fixed — a sidebar that
	// narrows with the window squeezes its own labels until they truncate,
	// and the labels are what the sidebar is for.
	SidebarWidth float32
}

// SettingsLayoutResult carries a SettingsLayout and what was chosen in it.
type SettingsLayoutResult struct {
	// Element is the whole layout.
	Element *ui.Element
	// side and main are the two panes, named so that a test can measure the
	// content each holds rather than the box around it: an empty box
	// measures zero however much room it was given. See
	// docs/design-system.md §15.2.
	side, main *ui.Element
}

// Sidebar returns the left column and Main the right, for a caller that has
// to put something of its own in one of them.
func (r SettingsLayoutResult) Sidebar() *ui.Element { return r.side }
func (r SettingsLayoutResult) Main() *ui.Element    { return r.main }

// SettingsLayout is a settings window: the sections down the left, the open
// one on the right, and a search box over both.
//
// It is a SplitPane rather than a hand-built row because the pane can be
// dragged, which is worth having in a window wide enough for two columns and
// is harmless in one that is not. The fraction is given rather than left to
// the caller because there is a right answer — a section list and a form —
// and every window having to rediscover it is how a settings pane ends up
// with a three-word sidebar.
//
// The search filters the list rather than jumping to a panel. A section that
// is not in the results is a section that does not exist as far as somebody
// looking for it is concerned, and trying to open one that is filtered out
// leaves the right-hand pane showing something that is not on the left.
func SettingsLayout(c *ui.Context, opts SettingsLayoutOptions) SettingsLayoutResult {
	if opts.Selected == nil {
		panic("account: SettingsLayout needs the *string Selected writes to")
	}
	if opts.Chosen == nil {
		panic("account: SettingsLayout needs the *string Chosen writes to")
	}
	if opts.Searchable && opts.Query == nil {
		panic("account: a searchable settings layout needs the *string Query " +
			"the search field writes to")
	}
	u := core.Density(c).Unit()

	// A selection the search has filtered out is deliberately left alone. The
	// right-hand pane keeps showing it while the left no longer lists it,
	// which looks odd for a frame — and the alternative is a search that
	// silently moves somebody to a different panel while they are typing,
	// which is worse than odd every time.
	sections := opts.Sections
	if opts.Searchable {
		sections = MatchSettings(*opts.Query, opts.Sections)
	}

	var r SettingsLayoutResult
	split := layout.SplitPane(c, layout.SplitPaneOptions{
		SideBySide: true,
		First:      settingsSidebarFraction,
		Min:        settingsSidebarMin,
		Max:        settingsSidebarMax,
		Gutter:     u * 1.5,
		FirstName:  core.Msg(c, "account.settingsSections", "Settings sections"),
		SecondName: core.Msg(c, "account.settingsPanel", "Settings panel"),
	}, func() {
		r.side = settingsSidebar(c, opts, sections)
	}, func() {
		r.main = settingsPanel(c, opts, *opts.Selected)
	})
	r.Element = split.Element
	return r
}

// The three numbers of a settings pane's split. They are constants rather
// than options because there is one right answer and every window that is
// allowed to have its own will end up with a sidebar too narrow for its own
// labels.
const (
	settingsSidebarFraction = 0.26
	settingsSidebarMin      = 190
	settingsSidebarMax      = 320
)

// settingsSidebar is the column of sections, with the search box over it.
func settingsSidebar(c *ui.Context, opts SettingsLayoutOptions, sections []Setting) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	side := ui.Column(c).FillWidth().FillHeight().Gap(u*3).Padding(0, 0, u, 0)
	side.Children(func() {
		if opts.Searchable {
			input.SearchField(c, opts.Query,
				core.Msg(c, "account.searchSettings", "Search settings"))
		}
		if len(sections) == 0 {
			ui.Text(c, core.Msg(c, "account.noMatches", "Nothing matches")).
				TextColor(k.TextFaint).FontSize(core.FontSize(c, theme.MetaSize)).SingleLine()
			return
		}
		for _, s := range sections {
			sectionRow(c, opts, s)
		}
	})
	return side
}

// sectionRow is one row of the section list.
func sectionRow(c *ui.Context, opts SettingsLayoutOptions, s Setting) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	open := s.ID == *opts.Selected

	row := ui.Row(c).FillWidth().Padding(u*1.25, u*2).Radius(theme.ControlRadius).
		AlignItems(ui.Center).Gap(u * 2).Cursor(ui.CursorPointer).
		Role(ui.RoleTab).Label(s.Name)
	if open {
		row.Background(k.SurfaceHover)
	}
	if row.Clicked() {
		*opts.Chosen = s.ID
		*opts.Selected = s.ID
	}
	row.Children(func() {
		ui.Text(c, s.Name).TextColor(k.Text).
			FontSize(core.FontSize(c, theme.RowSize)).Grow(1).SingleLine()
	})
}

// settingsPanel is the right-hand pane, and it says so when the selection is
// not a section it has.
func settingsPanel(c *ui.Context, opts SettingsLayoutOptions, id string) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	main := ui.Column(c).Grow(1).Shrink(0).FillHeight().Gap(u*3).Padding(u, 0, u, 0)

	if !hasSetting(opts.Sections, id) {
		main.Children(func() {
			ui.Text(c, core.Msg(c, "account.pickSection", "Pick a section")).
				TextColor(k.TextFaint).FontSize(core.FontSize(c, theme.RowSize))
		})
		return main
	}
	main.Children(func() {
		if opts.Section != nil {
			opts.Section(id)
		}
	})
	return main
}

// hasSetting reports whether a list of sections holds an ID.
func hasSetting(sections []Setting, id string) bool {
	for _, s := range sections {
		if s.ID == id {
			return true
		}
	}
	return false
}

// SettingsSearchOptions configure a SettingsSearch.
type SettingsSearchOptions struct {
	// Settings are the rows to search, in the order they should appear when
	// nothing has been typed.
	Settings []Setting
	// Query is the caller's search string.
	Query *string
	// Chosen writes the ID of the row that was pressed, empty when none was.
	Chosen *string
	// Limit caps how many rows are shown, because a search over two hundred
	// settings that matches everything is a list, not a search. Zero shows
	// every match.
	Limit int
	// Placeholder is the field's hint; empty takes the library's own.
	Placeholder string
}

// SettingsSearchResult carries a SettingsSearch and what was chosen in it.
type SettingsSearchResult struct {
	// Element is the field and the rows under it.
	Element *ui.Element
	// shown is how many rows MatchSettings produced and the limit allowed.
	shown int
}

// Shown reports how many rows the search is offering, which is what a caller
// says as "12 settings match" under the list.
func (r SettingsSearchResult) Shown() int { return r.shown }

// SettingsSearch is a field and the settings that match it.
//
// It is input.SearchField rather than a text box with an icon beside it, so
// that clearing it, submitting it and the way it looks in a narrow column are
// all the library's. The rows are matched by MatchSettings, which is the
// ranking rule; this component only decides how many of the answer to draw.
//
// The list is capped. A search that matches everything has not narrowed
// anything, and drawing two hundred rows of it in place of the settings pane
// is a worse answer than drawing the first eight and saying how many there
// were.
func SettingsSearch(c *ui.Context, opts SettingsSearchOptions) SettingsSearchResult {
	if opts.Query == nil {
		panic("account: SettingsSearch needs the *string Query writes to")
	}
	if opts.Chosen == nil {
		panic("account: SettingsSearch needs the *string Chosen writes to")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	found := MatchSettings(*opts.Query, opts.Settings)
	shown := len(found)
	if opts.Limit > 0 && shown > opts.Limit {
		found = found[:opts.Limit]
	}

	var r SettingsSearchResult
	r.shown = shown
	// Named for what it is — the results — rather than for the field above
	// it. Two elements called "Search settings" means a reader hears the same
	// name twice, and a test looking for the field finds the whole column.
	r.Element = ui.Column(c).FillWidth().Gap(u * 2).
		Label(core.Msg(c, "account.searchResults", "Settings results")).Children(func() {
		placeholder := opts.Placeholder
		if placeholder == "" {
			placeholder = core.Msg(c, "account.searchSettings", "Search settings")
		}
		input.SearchField(c, opts.Query, placeholder)
		for _, s := range found {
			resultRow(c, opts, s)
		}
		if shown > len(found) {
			ui.Text(c, core.Msg(c, "account.moreMatches", "More below")).
				TextColor(k.TextFaint).FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
		}
	})
	return r
}

// resultRow is one match.
func resultRow(c *ui.Context, opts SettingsSearchOptions, s Setting) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	row := ui.Row(c).FillWidth().Gap(u*3).Padding(u*1.5, u*2).
		Radius(theme.ControlRadius).AlignItems(ui.Center).Cursor(ui.CursorPointer).
		Label(s.Name).Role(ui.RoleListItem)
	if row.Clicked() {
		*opts.Chosen = s.ID
	}
	row.Children(func() {
		ui.Column(c).Grow(1).Shrink(0).Gap(u * 0.25).Children(func() {
			ui.Text(c, s.Name).TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).SingleLine()
			if s.Hint != "" {
				ui.Text(c, s.Hint).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
			}
		})
	})
}

// Release is one line of a what's-new list.
type Release struct {
	// Version is what the release is called, and is what a deep link into a
	// particular version carries.
	Version string
	// Date is when it shipped, already formatted.
	Date string
	// What is what changed, one sentence per line. They are the caller's
	// sentences because what changed is a fact about their app.
	What []string
	// Current marks the release the window is running. It is the one shown by
	// default and the one a "what's new" dialog opens on.
	Current bool
}

// WhatsNewOptions configure a WhatsNewDialog.
type WhatsNewOptions struct {
	// Open is the caller's bool: the dialog is modal and every way out of it
	// — Escape, the scrim, the button — writes false into this.
	Open *bool
	// Releases are the versions, newest first as the caller has them.
	Releases []Release
	// Selected is the release being read, as an index into Releases. It is
	// the caller's because which release somebody is reading is a place in
	// the app they can come back to, and a dialog that opened on a different
	// one each time it was shown would be useless for that.
	Selected *int
	// Body draws under the heading for a release that needs more than a list
	// of sentences — a screenshot, a diagram. Nil draws the sentences alone.
	Body func(version string)
}

// WhatsNewResult carries a WhatsNewDialog and the release it is showing.
type WhatsNewResult struct {
	// Element is the dialog, or nil while it is shut.
	Element *ui.Element
}

// WhatsNewDialog is what a release changed, as a dialog with a release list
// down one side.
//
// It is overlay.Dialog rather than a panel built here, so that it is modal by
// the library's own rules: the scrim makes the window behind inert, Escape
// closes it, and a press outside the panel goes to the scrim rather than
// through to the page underneath. All three of those are the kind of thing
// that is right in a dialog and subtly wrong in a hand-rolled one.
//
// The list is a list of releases rather than a single one because somebody
// who has skipped four versions wants to know what changed since the one they
// last saw, not only what changed last Tuesday.
func WhatsNewDialog(c *ui.Context, opts WhatsNewOptions) WhatsNewResult {
	if opts.Open == nil {
		panic("account: WhatsNewDialog needs the *bool it opens and closes")
	}
	if len(opts.Releases) == 0 {
		panic("account: WhatsNewDialog needs at least one release; a dialog " +
			"with nothing in it is a panel with a close button")
	}
	if opts.Selected == nil {
		panic("account: WhatsNewDialog needs the *int of the release being read")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	sel := *opts.Selected
	if sel < 0 || sel >= len(opts.Releases) {
		sel = indexOfCurrent(opts.Releases)
		*opts.Selected = sel
	}
	shown := opts.Releases[sel]

	return WhatsNewResult{Element: overlay.Dialog(c, opts.Open, overlay.DialogOptions{
		Title:    shown.Version,
		Subtitle: shown.Date,
		Width:    whatsNewWidth,
		MaxWidth: whatsNewMaxWidth,
		Rule:     true,
		Body: func() {
			ui.Row(c).FillWidth().Gap(u * 5).AlignItems(ui.Start).Children(func() {
				// The list of releases is what makes the dialog worth being a
				// dialog rather than a banner: somebody arriving after three
				// weeks wants to page back.
				ui.Column(c).Width(releaseListWidth).Shrink(0).Gap(u * 0.5).
					Role(ui.RoleList).Label(core.Msg(c, "account.releases", "Releases")).
					Children(func() {
						for i, rel := range opts.Releases {
							releaseRow(c, opts, i, rel)
						}
					})
				ui.Column(c).Grow(1).Shrink(0).Gap(u * 2).Children(func() {
					ui.Text(c, shown.Version).TextColor(k.Text).Bold().
						FontSize(core.FontSize(c, theme.SheetSize)).SingleLine()
					for _, line := range shown.What {
						releaseLine(c, line)
					}
					if opts.Body != nil {
						opts.Body(shown.Version)
					}
				})
			})
		},
		Actions: func() {
			ui.Box(c).Grow(1)
			if input.Button(c, core.Msg(c, "account.gotIt", "Got it"),
				input.ButtonOptions{Primary: true}).Clicked() {
				*opts.Open = false
			}
		},
	})}
}

// The two numbers of the dialog's split: wide enough for a list and a page of
// sentences side by side, narrow enough to leave the window visible behind it.
const (
	whatsNewWidth    float32 = 680
	releaseListWidth float32 = 150
	whatsNewMaxWidth float32 = 760
)

// releaseRow is one version in the list, marked when it is the one being read.
func releaseRow(c *ui.Context, opts WhatsNewOptions, i int, rel Release) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	on := i == *opts.Selected

	row := ui.Row(c).FillWidth().Padding(u*1.25, u*1.5).Radius(theme.ControlRadius).
		AlignItems(ui.Center).Gap(u).Cursor(ui.CursorPointer).
		Role(ui.RoleListItem).Label(rel.Version)
	if on {
		row.Background(k.SurfaceHover)
	}
	if row.Clicked() {
		*opts.Selected = i
	}
	row.Children(func() {
		ui.Text(c, rel.Version).TextColor(k.Text).Grow(1).
			FontSize(core.FontSize(c, theme.MetaSize)).SingleLine()
		if rel.Current {
			display.Tag(c, core.Msg(c, "account.current", "New"),
				display.TagOptions{Tone: core.Accent})
		}
	})
}

// releaseLine is one sentence of what changed, with a mark in front of it. The
// mark repeats the sentence's importance in shape as well as in colour, which
// is the one thing that survives somebody who cannot see the colour.
func releaseLine(c *ui.Context, line string) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	ui.Row(c).FillWidth().Gap(u * 2).AlignItems(ui.Start).Children(func() {
		ui.Box(c).Width(u).MinHeight(core.FontSize(c, theme.MetaSize)).Shrink(0).
			Margin(u, u*1.6, 0, 0).Radius(u / 2).Background(k.Accent)
		ui.Text(c, strings.TrimSpace(line)).TextColor(k.Text).
			FontSize(core.FontSize(c, theme.RowSize)).Grow(1)
	})
}

// indexOfCurrent is where the dialog opens when the caller has not said: on
// the release the window is running, or on the newest one if none of them is
// marked as current, which is what a list of old releases deserves.
func indexOfCurrent(list []Release) int {
	for i, r := range list {
		if r.Current {
			return i
		}
	}
	return 0
}
