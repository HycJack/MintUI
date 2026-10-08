package feedback

import (
	"fmt"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// ToolEntry is one tool in a ToolRegistryPanel: what it is called, what it
// does, and whether it is in.
type ToolEntry struct {
	// Name is what the tool is called. It is required: a description with
	// no name beside it cannot be enabled, named in an error, or found
	// again.
	Name string
	// Description is the one line that says what it does. Empty draws the
	// name alone, for a panel of names the caller already knows.
	Description string
	// Enabled is whether the tool is registered and answering. It decides
	// the row's mark: a filled dot when it is, a hollow one when it is
	// not.
	Enabled bool
}

// ToolRegistryPanelOptions configure a ToolRegistryPanel.
type ToolRegistryPanelOptions struct {
	// Tools are the registered tools, in the order to show them, which is
	// the caller's: the library does not sort enabled before disabled,
	// because a registry read in the order the tools were added is a
	// registry the reader can trace back to where they came from.
	Tools []ToolEntry
	// Title heads the panel; empty takes the library's "Tools".
	Title string
	// Width bounds the panel; zero lets it fill its parent.
	Width float32
}

// ToolRegistryPanel is the list of the tools an application has registered:
// one row each, a count over the list, and a mark on each row that says
// which of them will answer when called.
//
// The mark is a Presence's: a filled dot in the one colour that does not
// follow the appearance when the tool is in, a hollow ring when it is
// out, because a registry is a roll call — who is there — and the panel
// borrows the roll call's marks rather than inventing a second set that a
// presence column and a tool column could drift apart in. The count is in
// the header for the reason an evaluation table's is: the number a caller
// is usually after is the one that is stated, not the one that is summed.
func ToolRegistryPanel(c *ui.Context, opts ToolRegistryPanelOptions) *ui.Element {
	if len(opts.Tools) == 0 {
		panic("feedback: ToolRegistryPanel needs at least one ToolEntry; a " +
			"registry of nothing is a heading with a count of zero under it")
	}
	for _, tool := range opts.Tools {
		if tool.Name == "" {
			panic("feedback: a tool entry needs a Name; a description with " +
				"no name beside it cannot be enabled or found")
		}
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	title := opts.Title
	if title == "" {
		title = core.Msg(c, "feedback.toolRegistryPanel.title", "Tools")
	}
	count := fmt.Sprintf(
		core.Msg(c, "feedback.toolRegistryPanel.count", "%d tools"),
		len(opts.Tools))

	e := ui.Column(c).FillWidth().Gap(u * 2).Role(ui.RoleList)
	if opts.Width > 0 {
		e.Width(opts.Width)
	}
	e.Children(func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 2).Children(func() {
			ui.Text(c, title).TextColor(k.Text).
				FontSize(core.FontSize(c, theme.TitleSize)).Bold()
			ui.Box(c).Grow(1)
			ui.Text(c, count).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
		})
		for _, tool := range opts.Tools {
			toolEntryRow(c, tool)
		}
	})
	return e
}

// toolEntryRow draws one tool: its mark, its name, and the line under the
// name that says what it does.
func toolEntryRow(c *ui.Context, tool ToolEntry) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	dot := u * 3

	word := core.Msg(c, "feedback.toolRegistryPanel.enabled", "enabled")
	if !tool.Enabled {
		word = core.Msg(c, "feedback.toolRegistryPanel.disabled", "disabled")
	}

	// The mark is made inside the row rather than beside it, for the reason
	// a Presence's is: an element belongs to whatever was being built when
	// it was made.
	mark := func() *ui.Element {
		if tool.Enabled {
			// Lively, the presence dot's colour, which is the same in both
			// appearances: a tool that is in is the thing the reader looks
			// for, and it would stop being that on a dark desktop.
			return ui.Box(c).Size(dot, dot).Radius(dot / 2).Shrink(0).
				Background(k.Lively)
		}
		return ui.Box(c).Size(dot, dot).Radius(dot/2).Shrink(0).
			Border(theme.BorderWidth, k.TextFaint)
	}

	ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 2).
		Label(tool.Name + " " + word).Children(func() {
		mark()
		ui.Column(c).Grow(1).Gap(u * 0.5).Children(func() {
			ui.Text(c, tool.Name).TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).SingleLine()
			if tool.Description != "" {
				ui.Text(c, tool.Description).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.MetaSize)).SingleLine()
			}
		})
	})
}
