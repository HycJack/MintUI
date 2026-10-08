package navigation

import (
	"fmt"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// TabsOptions configure Tabs.
type TabsOptions struct {
	// Selected is the caller's index into Tabs' items.
	Selected *int
	// Vertical stacks the tabs down the side, for a panel that is taller
	// than it is wide.
	Vertical bool
	// Filled draws a surface behind the whole strip instead of leaving it on
	// the window.
	Filled bool
	// Grow makes every tab take an equal share, which is what a top-level
	// navigation wants and a two-tab toggle does not.
	Grow bool
}

// Tab is one entry of a Tabs strip.
type Tab struct {
	// Label is what the tab shows.
	Label string
	// Icon draws before the label; a tab with an icon and a label needs both
	// to be worth the width.
	Icon *ui.SVG
	// Badge is a short trailing mark — a count, a dot.
	Badge string
	// Disabled greys the tab and blocks the press.
	Disabled bool
}

// Tabs switches between views. Selection is the caller's index, so the tab
// strip never disagrees with what the view is showing.
func Tabs(c *ui.Context, items []Tab, opts TabsOptions) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if opts.Selected == nil {
		panic("navigation: Tabs needs a *int to report its selection into")
	}
	if len(items) == 0 {
		panic("navigation: Tabs needs at least one item")
	}

	bar := ui.Row(c).AlignItems(ui.Center).Gap(u)
	if opts.Vertical {
		bar = ui.Column(c).AlignItems(ui.Start).Gap(u * 0.5)
	}
	if opts.Filled {
		bar.Background(k.Surface)
	}
	bar.Children(func() {
		for i, it := range items {
			tab := it
			idx := i
			sel := *opts.Selected == i
			fg := k.TextMuted
			if sel {
				fg = k.Text
			}
			if tab.Disabled {
				fg = k.TextFaint
			}
			// Row, not Box: an icon, a label and a badge are one line. ui.Box is
			// ui.Column, and a tab whose icon sits over its title is not a tab.
			box := ui.Row(c).Padding(u*1.5, u*2.5).Radius(theme.ControlRadius).
				AlignItems(ui.Center).Gap(u).Label(tab.Label).
				Disabled(tab.Disabled).Children(func() {
				if tab.Icon != nil {
					sz := core.FontSize(c, theme.RowSize)
					ui.Icon(c, tab.Icon).Size(sz, sz).TextColor(fg)
				}
				ui.Text(c, tab.Label).TextColor(fg).SingleLine().
					FontSize(core.FontSize(c, theme.RowSize))
				if tab.Badge != "" {
					ui.Box(c).Padding(u*0.75, u*1.5).Radius(theme.PillRadius).
						Background(k.Surface).Children(func() {
						ui.Text(c, tab.Badge).TextColor(k.TextMuted).
							FontSize(core.FontSize(c, theme.CaptionSize))
					})
				}
				// The selection underline follows the strip's own direction,
				// which is why it is drawn here rather than by the caller. It
				// is absolute: a child in the row's own layout would take a
				// slot beside the label, whose text then wraps out of its way
				// while the bar runs from the label's right edge into
				// nothing.
				if sel {
					if opts.Vertical {
						ui.Box(c).Absolute().Left(0).Top(0).Bottom(0).
							Width(theme.BorderWidth * 2).Background(k.Text)
					} else {
						ui.Box(c).Absolute().Left(0).Right(0).Bottom(0).
							Height(theme.BorderWidth * 2).Background(k.Text)
					}
				}
			})
			if opts.Grow {
				box.Grow(1)
			}
			if on := box.Clicked(); on && !tab.Disabled {
				*opts.Selected = idx
			}
		}
	})
	return bar
}

// BreadcrumbItem is one hop of a Breadcrumb.
type BreadcrumbItem struct {
	// Label is what the hop shows.
	Label string
	// Current marks the page you are on; it is not a link, and saying so
	// stops a person clicking their way back to where they already are.
	Current bool
}

// BreadcrumbResult carries a Breadcrumb and the hop that was pressed.
type BreadcrumbResult struct {
	// Element is the trail.
	Element *ui.Element
	// Index is the hop pressed this frame, or -1.
	Index int
}

// Pressed reports a press of the hop at i.
func (r BreadcrumbResult) Pressed(i int) bool { return r.Index == i }

// Breadcrumb shows where you are in a hierarchy, and lets you go back up it.
func Breadcrumb(c *ui.Context, items []BreadcrumbItem) BreadcrumbResult {
	k, u := core.Tokens(c), core.Density(c).Unit()
	var r BreadcrumbResult
	r.Index = -1

	bar := ui.Row(c).AlignItems(ui.Center).Gap(u * 1.5).Children(func() {
		for i, it := range items {
			if i > 0 {
				ui.Text(c, "/").TextColor(k.TextFaint).
					FontSize(core.FontSize(c, theme.RowSize))
			}
			idx := i
			label := it.Label
			fg := k.TextMuted
			if it.Current {
				fg = k.Text
			}
			el := ui.Box(c).Padding(u*0.5, 0).Children(func() {
				t := ui.Text(c, label).TextColor(fg).
					FontSize(core.FontSize(c, theme.MetaSize))
				if !it.Current {
					t = t.Cursor(ui.CursorPointer).Label(label)
				}
			})
			if !it.Current && el.Clicked() {
				r.Index = idx
			}
		}
	})
	r.Element = bar
	return r
}

// Step is one stage of a Steps strip.
type Step struct {
	Label    string
	Done     bool
	Current  bool
	Disabled bool
}

// StepsOptions configure Steps.
type StepsOptions struct {
	// Vertical runs the steps down instead of across.
	Vertical bool
	// Numbers show 1, 2, 3 rather than a check on what is done.
	Numbers bool
}

// Steps shows progress through an ordered set of stages. It is not navigation:
// a step that is done is not a link back, so nothing here reports a press.
func Steps(c *ui.Context, steps []Step, opts StepsOptions) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	line := func() *ui.Element {
		return ui.Box(c).Background(k.Border).Grow(1)
	}
	marker := func(s Step, i int) *ui.Element {
		// faceInk is the circle's own ink on its fill; labelInk is the
		// label's ink on the page. One colour for both made the current
		// step's label OnFill — near-white on the white page — and the
		// stage vanished from its own marker.
		faceInk, bg, labelInk := k.TextFaint, k.Surface, k.TextMuted
		switch {
		case s.Current:
			faceInk, bg = k.OnFill, k.Fill
			labelInk = k.Text
		case s.Done:
			faceInk, labelInk = k.Text, k.Text
		}
		face := func() *ui.Element {
			if s.Done && !opts.Numbers {
				return ui.Box(c).Size(u*4, u*4).Radius(u * 2).Background(bg).
					Children(func() {
						ui.Text(c, "✓").TextColor(faceInk).
							FontSize(core.FontSize(c, theme.CaptionSize))
					})
			}
			return ui.Box(c).Size(u*4, u*4).Radius(u * 2).Background(bg).
				Children(func() {
					ui.Text(c, fmt.Sprint(i+1)).TextColor(faceInk).
						FontSize(core.FontSize(c, theme.CaptionSize))
				})
		}
		return ui.Row(c).AlignItems(ui.Center).Gap(u * 1.5).Children(func() {
			face()
			ui.Text(c, s.Label).TextColor(labelInk).
				FontSize(core.FontSize(c, theme.RowSize))
		})
	}

	if opts.Vertical {
		return ui.Column(c).AlignItems(ui.Start).Children(func() {
			for i, s := range steps {
				if i > 0 {
					ui.Box(c).Width(theme.BorderWidth).Height(u * 2).
						MarginX(u * 2).Background(k.Border)
				}
				marker(s, i)
			}
		})
	}
	return ui.Row(c).AlignItems(ui.Center).Gap(u * 2).Children(func() {
		for i, s := range steps {
			if i > 0 {
				line()
			}
			marker(s, i)
		}
	})
}

// PaginationOptions configure Pagination.
type PaginationOptions struct {
	// Page is the caller's zero-based page index.
	Page *int
	// Pages is the total.
	Pages int
	// Window is how many numbers show either side of the current page.
	Window int
	// Prev and Next label the two arrow buttons.
	Prev, Next string
}

// PaginationResult carries a Pagination and the page it moved to.
type PaginationResult struct {
	Element *ui.Element
	// Page is the new index, or -1 when nothing was pressed.
	Page int
}

// Pagination moves through pages. It shows a **window** of numbers rather than
// all of them: a hundred pages of "1 2 3 … 100" is a wall, not a control.
//
// A page outside [0, Pages) is drawn as the nearest one inside it rather than
// refused: two bars can share one page pointer — a narrow one drawn beside a
// wide one — and then the narrower bar is asked for a page it does not have.
// Drawn as it stood, it showed two arrows and no numbers between them; drawn
// at the nearest page it has, it is a bar on the last page and its arrows walk
// back from there. The caller's own page is written only by a press, so
// nothing here moves their state behind their back.
func Pagination(c *ui.Context, opts PaginationOptions) PaginationResult {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if opts.Page == nil {
		panic("navigation: Pagination needs a *int for its page")
	}
	if opts.Pages <= 0 {
		panic("navigation: Pagination needs at least one page")
	}
	cur := min(max(*opts.Page, 0), opts.Pages-1)
	window := opts.Window
	if window <= 0 {
		window = 2
	}
	prev, next := opts.Prev, opts.Next
	if prev == "" {
		prev = core.Msg(c, "pagination.prev", "Previous")
	}
	if next == "" {
		next = core.Msg(c, "pagination.next", "Next")
	}

	var r PaginationResult
	r.Page = -1
	arrow := func(label string, target int) *ui.Element {
		e := ui.Box(c).Padding(u*1, u*2).Radius(theme.ControlRadius).
			Background(k.Surface).Label(label).
			Disabled(target < 0 || target >= opts.Pages).Children(func() {
			ui.Text(c, label).TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize))
		})
		// The range check is repeated here rather than left to Disabled: a
		// disabled box still reports a press to whatever wrapped it, and an
		// arrow that walks off the end of the list is a real bug, not a look.
		if e.Clicked() && target >= 0 && target < opts.Pages {
			*opts.Page = target
			r.Page = target
		}
		return e
	}

	bar := ui.Row(c).AlignItems(ui.Center).Gap(u).Children(func() {
		arrow(prev, cur-1)
		pages := windowAround(cur, opts.Pages, window)
		for _, p := range pages {
			sel := p == cur
			col, bg := k.TextMuted, k.Surface
			if sel {
				col, bg = k.OnFill, k.Fill
			}
			e := ui.Box(c).Padding(u, u*2).Radius(theme.ControlRadius).
				Background(bg).Label(fmt.Sprint(p + 1)).Children(func() {
				ui.Text(c, fmt.Sprint(p+1)).TextColor(col).
					FontSize(core.FontSize(c, theme.RowSize))
			})
			if e.Clicked() {
				*opts.Page = p
				r.Page = p
			}
		}
		arrow(next, cur+1)
	})
	r.Element = bar
	return r
}

// windowAround returns the page numbers to show for a window of w either side
// of cur, always keeping cur visible and never going out of range.
func windowAround(cur, total, w int) []int {
	if total <= 0 {
		return nil
	}
	// Both ends are clamped: hi because a window runs off the end of the list,
	// lo because cur itself can — a filter that shrank the result set leaves
	// the caller's page past the last one — and an unclamped lo under a
	// clamped hi gave the slice a negative capacity, which panics.
	lo, hi := cur-w, cur+w
	if lo < 0 {
		lo = 0
	}
	if hi >= total {
		hi = total - 1
	}
	if lo > hi {
		lo = hi
	}
	out := make([]int, 0, hi-lo+1)
	for i := lo; i <= hi; i++ {
		out = append(out, i)
	}
	return out
}

// ToolbarOptions configure a Toolbar.
type ToolbarOptions struct {
	// Label names the toolbar for assistive technology.
	Label string
}

// Toolbar is a row of actions. It draws a surface and a hairline so a set of
// controls reads as one thing rather than as buttons floating on a page.
func Toolbar(c *ui.Context, opts ToolbarOptions, children func()) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	e := ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u).
		Padding(u*1.25, u*2).Background(k.Surface)
	if opts.Label != "" {
		e = e.Label(opts.Label)
	}
	if children != nil {
		e.Children(children)
	}
	return e
}

// ToolbarSeparator draws the gap between two groups of actions.
func ToolbarSeparator(c *ui.Context) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	return ui.Box(c).Width(theme.BorderWidth).Height(u * 5).Background(k.Border)
}
