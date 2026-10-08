package navigation

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/theme"
)

// SidebarItem is one destination in a Sidebar.
type SidebarItem struct {
	// Label is what the row shows, and what a screen reader announces.
	Label string
	// Icon is the glyph beside the label. A collapsed sidebar shows nothing
	// else, so an item without one cannot be told apart from its
	// neighbours — Sidebar does not check that for you, because a caller
	// folding its own sidebar knows which rows have glyphs and which do not.
	Icon display.IconName
	// Count is the number the row reports; nil draws none.
	Count *int
	// Tone tints the count, for the one count in a list that means
	// something is wrong rather than that there are some.
	Tone core.Severity
	// Selected shows the row as where you are. It is a value, not a
	// pointer: which row is current is the caller's app state, and the
	// sidebar reads it rather than tracking it.
	Selected bool
	// Disabled greys the row and stops it taking the press.
	Disabled bool
}

// SidebarSection is a titled run of rows.
type SidebarSection struct {
	// Title heads the section; empty runs the rows together with no gap,
	// which is what a single list of views wants.
	Title string
	// Items are the rows, top to bottom.
	Items []SidebarItem
}

// SidebarOptions configure a Sidebar.
type SidebarOptions struct {
	// Title sits at the top of the sidebar: the workspace or the product.
	Title string
	// Subtitle is the second line under it.
	Subtitle string
	// Label names the sidebar itself for assistive technology; empty uses
	// the title.
	Label string
	// Width overrides the expanded width; zero gives theme.SidebarWidth.
	//
	// It is a number rather than a fraction for the reason the theme says
	// once and only once: a sidebar that narrows with the window squeezes
	// the labels, and the labels are what the sidebar is for.
	Width float32
	// Collapsed is the caller's fold state. The sidebar draws to the rail's
	// width while it is true, and reports a press of its collapse control
	// through SidebarResult.Collapsed — so the state lives wherever the
	// rest of the app's state lives, and survives a rebuild.
	//
	// Nil is an always-expanded sidebar with no control at all, which is
	// what a window with no room to collapse wants.
	Collapsed *bool
	// Footer is drawn at the bottom, above the collapse control: the
	// "signed in as" row, the connection state.
	Footer func()
}

// SidebarResult carries a Sidebar and what was pressed on it.
type SidebarResult struct {
	// Element is the whole sidebar, expanded or collapsed.
	Element *ui.Element
	// collapsed reports a press of the collapse control this frame.
	collapsed bool
	// pressed is the index of the row pressed, or -1. It is the row's index
	// in SidebarSections.Items, flattened across the sections, because a
	// caller switching views keys on what was chosen and not on where it
	// was drawn — the same row is in a different place once collapsed.
	pressed int
}

// Collapsed reports a press of the collapse control.
func (r SidebarResult) Collapsed() bool { return r.collapsed }

// Pressed returns the flat index of the row pressed this frame, or -1. A
// disabled row reports -1 however hard it is pressed, which is the whole of
// what Disabled means.
func (r SidebarResult) Pressed() int { return r.pressed }

// Sidebar is the navigation column down a window's left: sections of
// destinations, one of them current, foldable to its icons.
//
// It is not Rail. Rail is the icon strip a window wears when there is no room
// for words at all — identity at the top, destinations, tools at the bottom,
// and nothing else. A Sidebar has titles, counts, sections and a footer, and
// can become the rail on demand; Rail stays because a window that will never
// collapse should not pay for the parts it will never draw.
//
// Its own state is the caller's: the fold is a *bool the caller owns, the
// selection is a value the caller reads, and a press comes back as an index.
// Nothing here is worth reopening the window to change.
func Sidebar(c *ui.Context, sections []SidebarSection, opts SidebarOptions) SidebarResult {
	k, u := core.Tokens(c), core.Density(c).Unit()
	name := opts.Label
	if name == "" {
		name = opts.Title
	}

	folded := opts.Collapsed != nil && *opts.Collapsed
	width := theme.SidebarWidth
	if opts.Width > 0 {
		width = opts.Width
	}
	if folded {
		width = theme.RailWidth
	}

	var r SidebarResult
	r.pressed = -1

	col := ui.Column(c).Width(width).Shrink(0).FillHeight().
		Background(k.Surface).Padding(u*3, u*2).Gap(u * 1.5)
	if name != "" {
		// The sidebar is a landmark, so it is named as one: the rows inside
		// it say where they go, and nothing would otherwise say what
		// contains them all.
		col = col.Label(name)
	}

	col.Children(func() {
		if opts.Title != "" {
			// The head is not labelled of its own: the title is its text, and
			// a second node with the sidebar's name would be one more thing
			// for a screen reader to say before the first row.
			head := ui.Column(c).FillWidth().Gap(0).Children(func() {
				// The title is set at the body's size, not the page's: a
				// sidebar is a column, and a 40pt word in one wraps into
				// three lines for no reason a reader can name.
				if folded {
					// Folded there is no room for the words, only for the
					// mark, so the mark is the identity — the same circle
					// and the same initials the rail wears at the top.
					ui.Box(c).Size(u*11, u*11).Radius(u * 5.5).Background(k.Fill).
						Center().Label(opts.Title).Children(func() {
						ui.Text(c, internal.Initials(opts.Title)).TextColor(k.OnFill).
							FontSize(core.FontSize(c, theme.StatSize)).Bold()
					})
					return
				}
				ui.Text(c, opts.Title).TextColor(k.Text).
					FontSize(core.FontSize(c, theme.TitleSize)).Bold().MaxLines(1)
			})
			if opts.Subtitle != "" && !folded {
				head.Children(func() {
					ui.Text(c, opts.Subtitle).TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.MetaSize)).MaxLines(1)
				})
			}
		}

		// The flat index is counted across the sections rather than read
		// off any one of them: it is the key the caller switches views on,
		// and the caller is the only thing that knows what its rows mean.
		idx := 0
		for _, sec := range sections {
			if folded {
				for _, it := range sec.Items {
					sidebarRow(c, it, u, true, idx, &r)
					idx++
				}
				continue
			}
			if sec.Title != "" {
				// A section heading is a caption, not a button: there is
				// nothing behind it to go to, and a row that looks
				// pressable and is not is worse than a line of small text.
				ui.Text(c, sec.Title).TextColor(k.TextFaint).
					FontSize(core.FontSize(c, theme.CaptionSize)).Bold().
					LetterSpacing(0.4).Padding(u*2, u*2.5, u*0.5, u*2.5)
			}
			for _, it := range sec.Items {
				sidebarRow(c, it, u, false, idx, &r)
				idx++
			}
		}

		ui.Spacer(c)
		if opts.Footer != nil {
			opts.Footer()
		}
		if opts.Collapsed != nil {
			r.collapsed = sidebarCollapse(c, *opts.Collapsed, u)
		}
	})

	r.Element = col
	return r
}

// sidebarRow draws one row and records the press against index.
//
// The index is handed in rather than counted here because the flat index is
// the key the caller switches views on, and counting rows inside the drawing
// helper would tie the reported index to the order this file happens to walk
// them in — which is not the same thing once a section is skipped folded.
func sidebarRow(c *ui.Context, it SidebarItem, u float32, folded bool, index int, r *SidebarResult) {
	k := core.Tokens(c)

	bg, fg := k.Surface, k.TextMuted
	if it.Selected {
		// The selected row is the only filled one. A sidebar of filled rows
		// is a stack of pills, and nothing on it is current any more.
		bg, fg = k.Background, k.Text
	}
	if it.Disabled {
		bg, fg = k.Surface, k.TextFaint
	}

	pad, side := u*2.5, u*3.5
	if folded {
		pad, side = u*1.75, u*1.5
	}
	// Row, not Box: a sidebar line is an icon, a label, a count and a
	// chevron in one line. ui.Box is ui.Column and would put them in a stack,
	// which is how a sidebar turns into a column of squares.
	row := ui.Row(c).FillWidth().Padding(pad, side).Radius(theme.ControlRadius).
		Background(bg).AlignItems(ui.Center).Gap(u * 2).
		Label(it.Label).Disabled(it.Disabled)

	row.Children(func() {
		if it.Icon != "" {
			// The glyph takes the muted tone when the row is not current, so
			// it reads as part of the label rather than as a separate mark
			// competing with it.
			ui.Box(c).Shrink(0).Children(func() {
				display.Icon(c, it.Icon, display.IconOptions{
					Name:  it.Label,
					Size:  theme.IconSize,
					Tone:  it.Tone,
					Muted: !it.Selected,
				})
			})
		}
		if !folded {
			ui.Text(c, it.Label).TextColor(fg).
				FontSize(core.FontSize(c, theme.BodySize)).Grow(1).MaxLines(1)
		}
		if !folded && it.Count != nil {
			_, countFg := it.Tone.Pair(k)
			chipBg := k.Border
			if it.Tone != core.Neutral {
				chipBg, _ = it.Tone.Pair(k)
			} else if it.Selected {
				chipBg = k.Surface
			}
			if countFg == k.Text {
				countFg = k.TextMuted
			}
			ui.Box(c).Shrink(0).Padding(u*0.75, u*2).Radius(theme.PillRadius).
				Background(chipBg).Children(func() {
				ui.Text(c, internal.Commas(*it.Count)).TextColor(countFg).
					FontSize(core.FontSize(c, theme.CaptionSize))
			})
		}
	})

	// The press is read after the children, so that a press on the count
	// chip is the row's press rather than nothing at all — the chip is a
	// label for the row, not a control of its own.
	if row.Clicked() && !it.Disabled {
		r.pressed = index
	}
}

// sidebarCollapse is the control that folds the sidebar to its icons.
func sidebarCollapse(c *ui.Context, collapsed bool, u float32) bool {
	k := core.Tokens(c)
	// The control names what it will do, not what it has done: a sidebar
	// showing its labels offers to take them away.
	label := core.Msg(c, "sidebar.collapse", "Collapse sidebar")
	name := "collapse"
	if collapsed {
		label = core.Msg(c, "sidebar.expand", "Expand sidebar")
		name = "expand"
	}
	btn := ui.Box(c).FillWidth().Padding(u*2, u*3).Radius(theme.ControlRadius).
		Background(k.Surface).AlignItems(ui.Center).Gap(u * 2).
		Label(label).Cursor(ui.CursorPointer).Children(func() {
		ui.Box(c).Shrink(0).Children(func() {
			display.Icon(c, display.IconName(name),
				display.IconOptions{Name: label, Size: theme.IconSize, Muted: true})
		})
		if !collapsed {
			ui.Text(c, label).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.MetaSize)).Grow(1)
		}
	})
	return btn.Clicked()
}
