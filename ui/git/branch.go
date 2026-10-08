package git

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// Branch is one branch, as a branch list has to show it.
type Branch struct {
	// Name is the branch's name, without the refs/heads/ that a full
	// reference carries and nobody reads.
	Name string
	// Current marks the branch the working tree is on. It is drawn as a dot
	// on the row and named for assistive technology, because it is the one
	// fact about a branch a person cannot read off the list.
	Current bool
	// Remote marks a branch that lives on a server. Its name is written the
	// same way — the same name, in the same ink — because it is the same
	// branch as far as switching goes, and only fetching is different.
	Remote bool
	// Ahead and Behind are how far the local branch has moved from what it is
	// tracking: commits it has that the remote does not, and the other way
	// round. Both zero is the common case and says nothing.
	Ahead, Behind int
	// Subject is the branch tip's one-line commit message.
	Subject string
}

// BranchListOptions configure a BranchList.
type BranchListOptions struct {
	// Height is the height of the list. It is required: a branch list with no
	// height grows to fit every branch there is, which is the one thing a
	// repository never has few of.
	Height float32
	// ShowRemote draws the remote branches as well as the local ones. A list
	// that shows both is twice as long and is what a fetch-and-look window
	// wants; a list that shows one is what a checkout menu wants.
	ShowRemote bool
	// Empty draws instead of the list when there is nothing in it.
	Empty func()
}

// BranchListResult carries a BranchList and the branch chosen in it.
type BranchListResult struct {
	// Element is the whole list.
	Element *ui.Element
	// picked is the branch chosen this frame, -1 for none.
	picked int
}

// Picked returns the index of the branch chosen this frame, or -1.
//
// It is a frame rather than a field: the selection is the caller's pointer
// and stays where they left it, so reading this twice in one frame means one
// press was seen twice rather than two presses having happened.
func (r BranchListResult) Picked() int { return r.picked }

// BranchList is a scrollable list of branches, one chosen at a time.
//
// The chosen branch is the caller's *int, not this component's: a branch list
// that remembered its own selection would disagree with the toolbar that
// switched the branch, and the two would be right about different things. A
// click writes the index and reports it; nothing here holds a choice.
//
// The row is what git prints — a dot for the branch the tree is on, the name,
// the ahead/behind counts, then the tip's subject — because anybody who came
// to a branch list already reads that shape.
func BranchList(c *ui.Context, selected *int, branches []Branch, opts BranchListOptions) BranchListResult {
	if selected == nil {
		panic("git: BranchList needs a selection to point at; it keeps no branch of its own")
	}
	if opts.Height <= 0 {
		panic("git: BranchList needs a Height; a branch list with no height is every branch there is")
	}
	u := core.Density(c).Unit()

	shown := make([]int, 0, len(branches))
	for i, b := range branches {
		if !opts.ShowRemote && b.Remote {
			continue
		}
		shown = append(shown, i)
	}

	var res BranchListResult
	res.picked = -1

	list := ui.Scroll(c).Height(opts.Height).FillWidth().Gap(u * 0.25).Role(ui.RoleList)
	list.Children(func() {
		if len(shown) == 0 {
			emptyBranches(c, opts.Empty)
			return
		}
		for _, i := range shown {
			chosen := *selected == i
			branchRow(c, branches[i], chosen, func() {
				*selected = i
				res.picked = i
			})
		}
	})
	res.Element = list
	return res
}

// branchRow is one branch: the dot, the name, how far out it is, and the
// tip's subject. pick asks the row what to do, which is how the list keeps
// the selection and the selector keeps the closing.
func branchRow(c *ui.Context, b Branch, chosen bool, pick func()) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	row := pickRow(c, branchLabel(b), chosen, func() {
		dot := k.Border
		if b.Current {
			dot = k.Text
		}
		ui.Box(c).Size(u*1.75, u*1.75).Shrink(0).Radius(u).
			Background(dot).Label(markName(b.Current))
		ui.Text(c, b.Name).FontSize(core.FontSize(c, theme.RowSize)).SingleLine()
		branchCounts(c, b)
		ui.Box(c).Grow(1)
		if b.Subject != "" {
			ui.Text(c, b.Subject).TextColor(k.TextMuted).Grow(1).
				Ellipsis(b.Subject).
				FontSize(core.FontSize(c, theme.MetaSize)).SingleLine()
		}
	})
	if row.Clicked() {
		pick()
	}
	return row
}

// markName is what the dot on a branch row is called out loud. A dot that
// means "this is the branch you are on" has to say so; a dot that is only a
// dot reads as a bullet.
func markName(current bool) string {
	if current {
		return "current branch"
	}
	return "branch"
}

// branchCounts is the ahead/behind a branch is out by, and nothing at all
// when it is not out by anything: a pair of arrows beside every branch in a
// list that is mostly even is noise that teaches the reader nothing.
func branchCounts(c *ui.Context, b Branch) {
	if b.Ahead <= 0 && b.Behind <= 0 {
		return
	}
	u := core.Density(c).Unit()
	ui.Box(c).Gap(u * 0.25).Children(func() {
		if b.Ahead > 0 {
			arrowCount(c, "↑"+itoa(b.Ahead), core.Success, "ahead by")
		}
		if b.Behind > 0 {
			arrowCount(c, "↓"+itoa(b.Behind), core.Warning, "behind by")
		}
	})
}

func arrowCount(c *ui.Context, s string, tone core.Severity, name string) {
	k := core.Tokens(c)
	_, fg := tone.Pair(k)
	ui.Text(c, s).TextColor(fg).Label(name + " " + s).
		FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
}

// branchLabel is what a branch row is called out loud: the name, and which
// branch it is when the row is the one the tree is on.
func branchLabel(b Branch) string {
	switch {
	case b.Current && b.Remote:
		return b.Name + ", current branch, remote"
	case b.Current:
		return b.Name + ", current branch"
	case b.Remote:
		return b.Name + ", remote"
	}
	return b.Name
}

// emptyBranches is what a branch list says when there is nothing in it, or
// what the caller says instead.
func emptyBranches(c *ui.Context, draw func()) {
	if draw != nil {
		draw()
		return
	}
	k := core.Tokens(c)
	ui.Text(c, noRows(c, "git.noBranches", core.Def("No branches"))).
		TextColor(k.TextMuted).FontSize(core.FontSize(c, theme.RowSize))
}

// BranchSelectorOptions configure a BranchSelector.
type BranchSelectorOptions struct {
	// Label names the control for assistive technology and heads the panel it
	// opens. It is required: a trigger whose text is a branch name says
	// nothing about what pressing it does.
	Label string
	// Height caps the panel's list; zero lets it take what the branches need.
	Height float32
}

// BranchSelectorResult carries a BranchSelector and what was done with it.
type BranchSelectorResult struct {
	// Element is the trigger. The panel is an overlay and is not part of it.
	Element *ui.Element
	// picked is the branch chosen from the panel this frame, -1 for none.
	picked int
	// opened reports the panel being opened this frame, so a caller can
	// fetch the list it is about to show.
	opened bool
}

// Picked returns the index of the branch chosen from the panel this frame, or
// -1.
func (r BranchSelectorResult) Picked() int { return r.picked }

// Opened reports the panel being opened this frame.
func (r BranchSelectorResult) Opened() bool { return r.opened }

// BranchSelector is the trigger a branch is switched from: the current
// branch's name, and the list of the others under it.
//
// Both halves of its state are the caller's — the selection, and whether the
// panel is showing. An open flag kept on the trigger would be a second source
// of truth beside the caller's, and the two would drift the first time a
// caller closed the panel from somewhere else, which is what a command-K
// does.
//
// Choosing from the panel writes the selection and closes it, because a
// dropdown that stays open after a choice has not been made is a menu, not a
// selector.
func BranchSelector(c *ui.Context, selected *int, open *bool, branches []Branch, opts BranchSelectorOptions) BranchSelectorResult {
	if selected == nil || open == nil {
		panic("git: BranchSelector needs the selection and the open flag to point at")
	}
	if opts.Label == "" {
		panic("git: BranchSelector needs a Label; a trigger showing a branch name says nothing about what it does")
	}
	if len(branches) == 0 {
		panic("git: BranchSelector needs at least one branch to show the current one")
	}
	if *selected < 0 || *selected >= len(branches) {
		panic("git: BranchSelector selection " + itoa(*selected) + " is not one of its " + itoa(len(branches)) + " branches")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	var res BranchSelectorResult
	res.picked = -1

	trigger := ui.ButtonBase(c).Height(core.ControlHeight(c)).Radius(theme.ControlRadius).
		Padding(0, u*2.5).FillWidth().Background(k.Background).TextColor(k.Text).
		BorderWidth(theme.BorderWidth).BorderColor(k.Border).
		Label(opts.Label).Tooltip(opts.Label)
	if trigger.Clicked() {
		*open = !*open
		res.opened = *open
	}
	trigger.Children(func() {
		ui.Text(c, branches[*selected].Name).SingleLine().Grow(1).
			TextAlign(ui.Start).FontSize(core.FontSize(c, theme.RowSize))
		chevron(c)
	})
	res.Element = trigger

	if !*open {
		return res
	}

	ui.PopoverBase(c, trigger, open, func(host *ui.Element) {
		layout.Panel(c, host, layout.PanelOptions{
			Title: opts.Label, Rule: true, Label: opts.Label,
		}, func() {
			body := ui.Column(c).FillWidth().Gap(u * 0.25).MaxHeight(280)
			body.Children(func() {
				for i := range branches {
					branchRow(c, branches[i], *selected == i, func() {
						*selected = i
						*open = false
						res.picked = i
					})
				}
			})
		})
	})
	return res
}

// chevron is the arrow of a trigger that opens something under it, drawn here
// rather than taken from MyGo so that it is in this window's own muted ink
// whatever the host's theme happens to be.
func chevron(c *ui.Context) {
	k := core.Tokens(c)
	side := core.Density(c).Unit() * 2.5
	ui.Box(c).Size(side, side).Shrink(0).Role(ui.RoleNone).
		Draw(func(p *ui.Painter, r ui.Rect) {
			var path ui.Path
			path.MoveTo(r.X+r.W*0.1, r.Y+r.H*0.3).
				LineTo(r.X+r.W*0.5, r.Y+r.H*0.7).
				LineTo(r.X+r.W*0.9, r.Y+r.H*0.3)
			p.StrokePath(&path, 1.5, k.TextMuted)
		})
}
