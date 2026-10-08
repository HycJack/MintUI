package navigation

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

func internalInitials(name string) string { return internal.Initials(name) }

// FilterRowOptions configure a FilterRow.
type FilterRowOptions struct {
	// Count is the number the row is reporting; nil draws none.
	Count *int
	// Severity tints the count, for a count that means something is wrong
	// rather than merely that there are some.
	Severity core.Severity
	// Indented marks the row as one level in, under a group heading.
	Indented bool
	// Selected fills the row and darkens its text.
	Selected bool
	// Label names the row for assistive technology; empty uses the text.
	Label string
}

// FilterRow is one selectable entry in a filter list.
//
// Only the selected row is filled. Rows left empty sit straight on the panel
// behind them, which is what keeps a long list from reading as a stack of
// pills — the selection is the only thing that should catch the eye.
func FilterRow(c *ui.Context, text string, opts FilterRowOptions) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()

	var bg ui.Color
	fg := k.TextMuted
	chipBg, chipFg := k.Border, k.TextMuted
	if opts.Selected {
		bg, fg = k.Background, k.Text
		chipBg = k.Surface
	}
	if opts.Severity != core.Neutral {
		chipBg, chipFg = opts.Severity.Pair(k)
	}

	label := opts.Label
	if label == "" {
		label = text
	}
	row := ui.Row(c).FillWidth().Padding(u*2.5, u*3.5, u*2.5, u*3.5).
		Radius(theme.ControlRadius).Background(bg).AlignItems(ui.Center).Label(label)

	row.Children(func() {
		if opts.Indented {
			// A short rule standing in for the tree line a group heading has.
			ui.Box(c).Width(u * 3).Height(1.5).Background(k.Border).Shrink(0)
		}
		ui.Text(c, text).TextColor(fg).FontSize(theme.RowSize).Grow(1)
		if opts.Count != nil {
			ui.Box(c).Padding(u, u*2.25, u, u*2.25).Radius(theme.PillRadius).
				Background(chipBg).Children(func() {
				ui.Text(c, internal.Commas(*opts.Count)).
					TextColor(chipFg).FontSize(theme.CaptionSize)
			})
		}
	})
	return row
}

// GroupOptions configure a Group.
type GroupOptions struct {
	// Title heads the group.
	Title string
	// Icon marks the group; nil leaves it out.
	Icon *ui.SVG
	// Open is the caller's fold state. A Group does not hold it: two lists
	// showing the same group should agree about whether it is open.
	Open bool
	// Indent puts the group's rows one level in, under a tree line.
	Indent bool
}

// GroupResult carries a Group's heading and what the user did with it.
type GroupResult struct {
	// Element is the whole group, heading included.
	Element *ui.Element
	// Toggled reports a press of the heading this frame.
	toggled bool
}

// Toggled reports a press of the heading.
func (r GroupResult) Toggled() bool { return r.toggled }

// Group is a heading over a set of rows that folds away.
//
// It reports the toggle rather than holding it, so the fold state lives with
// the rest of the app's state and survives a rebuild.
func Group(c *ui.Context, opts GroupOptions, body func()) GroupResult {
	k, u := core.Tokens(c), core.Density(c).Unit()
	var r GroupResult

	r.Element = ui.Column(c).FillWidth().Children(func() {
		head := ui.Row(c).FillWidth().Padding(u*2.5, u*3, u*2.5, u*3).
			Radius(theme.ControlRadius).AlignItems(ui.Center).Label(opts.Title)
		head.Children(func() {
			if opts.Icon != nil {
				ui.Icon(c, opts.Icon).TextColor(k.TextMuted).Size(u*4.5, u*4.5)
			}
			ui.Text(c, opts.Title).TextColor(k.TextMuted).FontSize(theme.RowSize).Grow(1)
			ui.Icon(c, chevron(opts.Open)).TextColor(k.TextMuted).Size(u*4, u*4)
		})
		r.toggled = head.Clicked()

		if opts.Open && body != nil {
			ui.Column(c).FillWidth().Gap(u * 0.5).Children(body)
		}
	})
	return r
}

// chevron points the way a group will move when pressed.
func chevron(open bool) *ui.SVG {
	if open {
		return chevronDown
	}
	return chevronRight
}

var (
	chevronDown  = mustIcon(`<path d="m7 10 5 5 5-5"/>`)
	chevronRight = mustIcon(`<path d="m10 7 5 5-5 5"/>`)
)

func mustIcon(shapes string) *ui.SVG {
	return ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" ` +
		`fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" ` +
		`stroke-linejoin="round">` + shapes + `</svg>`))
}
