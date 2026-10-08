package navigation

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/theme"
)

// ToolbarGroupOptions configure a ToolbarGroup.
type ToolbarGroupOptions struct {
	// Label names the group for assistive technology. It is required: a run
	// of buttons is announced as a run of buttons otherwise, and "Save, Undo,
	// Redo, Print, Preview" is a list, not a grouping — a reader cannot hear
	// which of those edits the document and which of them export it.
	Label string
}

// ToolbarGroup is a run of related actions inside a Toolbar.
//
// It exists so the bar can say what its parts are for rather than only where
// they are. ToolbarSeparator draws the gap, which is a visual answer to
// "where does one end and the next begin"; ToolbarGroup answers "what are
// these", which a gap cannot. A toolbar of one group needs no separators, and
// a toolbar of six needs five gaps and three groups.
func ToolbarGroup(c *ui.Context, opts ToolbarGroupOptions, children func()) *ui.Element {
	if opts.Label == "" {
		panic("navigation: ToolbarGroup needs a Label; an unnamed run of buttons is a list, not a group")
	}
	u := core.Density(c).Unit()
	return ui.Row(c).AlignItems(ui.Center).Gap(u * 0.5).
		Label(opts.Label).Role(ui.RoleGroup).
		Children(children)
}

// ToolbarButtonOptions configure a ToolbarButton.
type ToolbarButtonOptions struct {
	// Icon is the glyph. A button with no label needs one — a bare box with
	// nothing in it is not a button, it is a gap.
	Icon display.IconName
	// Primary fills the button with ink, for the one action a surface is
	// about.
	Primary bool
	// Danger tints the label and leaves it unfilled, for the action that
	// destroys. It is deliberately not a second Primary: a red filled button
	// on every toolbar is a toolbar where nothing can be pressed safely.
	Danger bool
	// Disabled greys the button and blocks the press.
	Disabled bool
	// Active presses the button in — a toggle that is on, such as a filter
	// chip in a toolbar.
	Active bool
	// Label names the button for assistive technology; empty uses the text.
	Label string
}

// ToolbarButton is one action in a Toolbar: an icon, a word, or both.
//
// It reports the press and holds nothing. There is no Selected to set,
// because the only thing a toolbar button has that a plain button does not is
// the tool it is grouped with — and that is ToolbarGroup's business, not the
// button's. A button that is on is Active, which is the caller reading its
// own state back out.
//
// Icon and label together is the normal shape; an icon alone is a button whose
// label is on the tooltip and nowhere else, which is why the name is
// required either way.
func ToolbarButton(c *ui.Context, label string, opts ToolbarButtonOptions) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	name := opts.Label
	if name == "" {
		name = label
	}
	if name == "" {
		// Icon-only is allowed, but then the icon is the button's whole
		// content and its Name is the only word it will ever be given — so
		// the name is required, not optional, exactly as it is in
		// display.Icon. A control with neither a word nor a glyph is a gap.
		panic("navigation: ToolbarButton needs a label or options.Label")
	}

	tone := core.Neutral
	if opts.Danger {
		tone = core.Danger
	}

	bg, fg := k.Surface, k.Text
	switch {
	case opts.Primary:
		bg, fg = k.Fill, k.OnFill
	case opts.Danger:
		fg = k.Danger
	case opts.Disabled:
		bg, fg = k.Surface, k.TextFaint
	case opts.Active:
		bg, fg = k.SurfacePressed, k.Text
	}

	// Row, not Box: the icon and the label are side by side, and a Box
	// would stack them — then Height would cut the label in half.
	btn := ui.Row(c).Height(core.ControlHeight(c)).Padding(0, u*3).
		Radius(theme.ControlRadius).Background(bg).
		AlignItems(ui.Center).Gap(u * 1.5).Label(name).
		Disabled(opts.Disabled).Children(func() {
		if opts.Icon != "" {
			// The glyph takes the same ink as the word beside it, so a
			// danger button is one red thing rather than a red word next to
			// a grey mark. Muted would win over Tone in display.Icon, which
			// is why it is off for the danger case.
			ui.Box(c).Shrink(0).Children(func() {
				display.Icon(c, opts.Icon, display.IconOptions{
					Name:  name,
					Size:  theme.IconSize,
					Tone:  tone,
					Muted: !opts.Primary && !opts.Active && !opts.Danger,
				})
			})
		}
		if label != "" {
			ui.Text(c, label).TextColor(fg).
				FontSize(core.FontSize(c, theme.RowSize)).MaxLines(1)
		}
	})
	return btn
}
