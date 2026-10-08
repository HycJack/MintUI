package input

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// Transfer is a pair of lists with the arrows between them: what is on offer
// on one side, what has been taken to the other.
//
// Both sides are the caller's []string and the choices are the caller's
// []Choice, so a form submitting a Transfer submits the two slices it already
// had. Nothing here decides what may be taken where — a caller with a rule,
// only accounts the user may touch or only branches with no unmerged work,
// answers it by not offering the choice at all, which is a rule about the
// data rather than about the widget.
//
// The ticks are one list shared by both sides rather than one each. Two lists
// would be a tick meaning "wanted" on one side and "chosen" on the other, and
// a value moved across would arrive with its old tick still set — a transfer
// that quietly puts things back on the next press.

// TransferOptions configure a Transfer.
type TransferOptions struct {
	// Left and Right name the two sides for assistive technology. Both are
	// required: two columns of ticks and two arrows say which way each arrow
	// goes and nothing about what is on either side of it, and a reader who
	// cannot tell which column is the offer from which is the taken has no
	// way to use the control at all.
	Left, Right string
	// Choices is everything on offer, in the order the columns are read. A
	// value already on the right is not offered on the left.
	Choices []Choice
	// Height is how tall each column is. It is required: a column that grows
	// to fit two hundred choices and takes the page with it is not a
	// transfer, it is the whole window.
	Height float32
	// Column is how wide each side is; zero shares what the row gives it.
	Column float32
	// Selectable is which choices are ticked, one bool per choice. It is nil
	// for a transfer with nothing to tick, where a press on a row is itself
	// the move.
	//
	// The ticks are shared by both columns so a choice keeps its tick when it
	// crosses, and "put that back" is a second press of the same tick rather
	// than a second list to keep in step with the first.
	Selectable *[]bool
	// InOrder puts each column in the choices' own order rather than in the
	// order things were moved. It is on by default: a column that reorders
	// itself as it is used is a column nobody can find anything in, and
	// "moved across last" is not a useful way to remember anything.
	InOrder bool
	// Disabled takes the whole control out of play.
	Disabled bool
}

// TransferResult carries a Transfer and what it moved.
type TransferResult struct {
	// Element is the whole control: the two columns and the arrows between.
	Element *ui.Element
	// right and left report a move this frame and which way.
	right, left bool
	// moved is how many values crossed this frame, which is what a caller
	// asking "did anything change" needs and what a bool cannot say.
	moved int
}

// MovedRight reports that something was taken from the left to the right this
// frame.
func (r TransferResult) MovedRight() bool { return r.right }

// MovedLeft reports that something was put back this frame.
func (r TransferResult) MovedLeft() bool { return r.left }

// Moved is how many values crossed this frame, whichever way. It is a count
// rather than a bool because the press of "take all" moves a hundred, and a
// caller holding only a bool cannot tell one of them from a hundred.
func (r TransferResult) Moved() int { return r.moved }

// Transfer is a pair of lists and the arrows that move values between them.
//
// It writes into the caller's slices rather than into copies it made first,
// and it enforces the two rules that make a transfer correct and no others: a
// value is on one side or the other and never both, and pressing an arrow
// twice does not move it twice.
func Transfer(c *ui.Context, left, right *[]string, opts TransferOptions) TransferResult {
	if left == nil || right == nil {
		panic("input: Transfer needs both sides to point at; it keeps neither of its own")
	}
	if opts.Left == "" || opts.Right == "" {
		panic("input: Transfer needs names for both sides; two columns of ticks and two arrows " +
			"say which way each arrow goes and nothing about what is on either side of it")
	}
	if len(opts.Choices) == 0 {
		panic("input: Transfer needs at least one choice; two empty columns have nothing to move")
	}
	if opts.Height <= 0 {
		panic("input: Transfer needs a Height; a column that grows to fit its choices takes " +
			"the page with it, which is not a transfer but the whole window")
	}
	checkChoicesHaveLabels(opts.Choices, "Transfer")

	if opts.Selectable != nil && len(*opts.Selectable) < len(opts.Choices) {
		// Grown rather than refused. The tick list is the one piece of state
		// here a caller adds a choice to and forgets, and a column indexing
		// past its own ticks would panic on a control that is only drawn.
		grown := make([]bool, len(opts.Choices))
		copy(grown, *opts.Selectable)
		*opts.Selectable = grown
	}
	// A value on both sides is a caller who appended one and did not take the
	// other off, and the columns would then show it twice with two different
	// meanings. The taken side wins: a value that has been taken is taken.
	*left = withoutAll(*left, *right)

	u := core.Density(c).Unit()
	var r TransferResult

	host := ui.Row(c).FillWidth().AlignItems(ui.Stretch).Gap(u * 1.5)
	// Everything is built inside the host's own Children call, because an
	// element made out here lands in the caller's container rather than in
	// this control's — the rule the note in ui/input's shared.go is about.
	host.Children(func() {
		transferColumn(c, opts, left, right, false, func(moved int) {
			r.right, r.moved = moved > 0, moved
		})
		transferArrows(c, opts, left, right, &r)
		transferColumn(c, opts, left, right, true, func(moved int) {
			r.left, r.moved = moved > 0, moved
		})
	})
	r.Element = host
	return r
}

// transferColumn is one side: its name, how many are on it, and a scrolling
// list of rows.
//
// left and right are the caller's two slices and isRight says which of them
// this column shows. Both are passed rather than derived, because the two
// columns are built by one call and the only way for them to disagree about
// which slice is which is to be built from a helper that guesses — and a
// transfer whose two columns move out of each other's slices loses one value
// per press.
func transferColumn(c *ui.Context, opts TransferOptions, left, right *[]string, isRight bool, moved func(int)) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	name, from, to := opts.Left, left, right
	if isRight {
		name, from, to = opts.Right, right, left
	}

	col := ui.Column(c).Shrink(0).Gap(u).FillHeight()
	if opts.Column > 0 {
		col.Width(opts.Column)
	} else {
		col.Grow(1)
	}
	col.Children(func() {
		// The count is in the header rather than in a row, so it says the
		// same thing at every height the column is scrolled to.
		ui.Row(c).FillWidth().AlignItems(ui.Center).Justify(ui.SpaceBetween).Children(func() {
			ui.Text(c, name).TextColor(k.TextMuted).SingleLine().Shrink(1).
				FontSize(core.FontSize(c, theme.CaptionSize)).FontWeight(600)
			ui.Text(c, itoa(len(*from))).TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
		})

		popupList(c, opts.Height).Children(func() {
			if len(*from) == 0 {
				empty(c, core.Msg(c, "input.nothingHere", core.Def("Nothing here")))
				return
			}
			for _, ch := range transferOrder(opts, *from) {
				choice := ch
				at := indexOfChoice(opts.Choices, choice.Value)
				chosen := at >= 0 && opts.Selectable != nil && (*opts.Selectable)[at]

				m := noMark
				if opts.Selectable != nil {
					m = tickMark
					if chosen {
						m = tickOn
					}
				}
				row := optionRow(c, choice.Label, choice.Value, optionFace{Mark: m, Chosen: chosen}).
					Disabled(opts.Disabled).
					// The tooltip is the whole name: a column narrower than
					// its longest choice truncates, and the words cut off are
					// the words that say which item this row is.
					Tooltip(choice.Label)
				if !row.Clicked() {
					continue
				}
				switch {
				case opts.Selectable != nil && at >= 0:
					// A tick rather than a move: the arrows are what read
					// the ticks, and a transfer whose rows moved as well
					// would have two answers to what a row is for.
					(*opts.Selectable)[at] = !(*opts.Selectable)[at]
					moved(0)
				case at >= 0:
					moved(transferCross(from, choice.Value, to))
				}
			}
		})
	})
	return col
}

// transferOrder is the values one side shows, in the choices' own order.
func transferOrder(opts TransferOptions, from []string) []Choice {
	out := make([]Choice, 0, len(from))
	if opts.InOrder {
		for _, ch := range opts.Choices {
			if hasValue(from, ch.Value) {
				out = append(out, ch)
			}
		}
		return out
	}
	// The moved order, as the caller's slice has it. A caller that wants the
	// choices' order leaves InOrder on.
	for _, v := range from {
		if ch, ok := findChoice(opts.Choices, v); ok {
			out = append(out, ch)
		}
	}
	return out
}

// transferArrows is the column between the two: one arrow for the ticked
// values and one for all of them, both ways.
//
// Four buttons rather than one button that grows, because "take the ticked" and
// "take everything" are different acts with different consequences, and a
// control that made the difference by how many boxes happened to be ticked
// would be one press away from taking ninety of a hundred by mistake.
func transferArrows(c *ui.Context, opts TransferOptions, left, right *[]string, r *TransferResult) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	arrows := []struct{ name, tip, mark string }{
		{core.Msg(c, "input.moveRight", core.Def("Move right")),
			core.Msg(c, "input.moveRightTip", core.Def("Move the ticked values right")), "→"},
		{core.Msg(c, "input.moveAllRight", core.Def("Move all right")),
			core.Msg(c, "input.moveAllRightTip", core.Def("Move every value right")), "⇉"},
		{core.Msg(c, "input.moveAllLeft", core.Def("Move all left")),
			core.Msg(c, "input.moveAllLeftTip", core.Def("Move every value back")), "⇇"},
		{core.Msg(c, "input.moveLeft", core.Def("Move left")),
			core.Msg(c, "input.moveLeftTip", core.Def("Move the ticked values back")), "←"},
	}
	col := ui.Column(c).Justify(ui.Center).Gap(u * 0.5).Shrink(0)
	col.Children(func() {
		for i, a := range arrows {
			side := left
			if i >= 2 {
				side = right
			}
			// An arrow with nothing to move is not drawn at all rather than
			// drawn dead: a greyed arrow still looks like an arrow, and a
			// press on it does nothing whatever, which is worse than there
			// being none.
			if len(*side) == 0 {
				continue
			}
			if opts.Selectable == nil && (i == 0 || i == 3) {
				// With nothing to tick, the ticked arrow would move nothing
				// ever, so the two all-arrows stand in for the four.
				continue
			}
			btn := ui.ButtonBase(c).Size(u*7.5, u*6).Shrink(0).Radius(theme.SmallRadius).
				Background(k.Surface).TextColor(k.Text).Role(ui.RoleButton).
				Label(a.name).Tooltip(a.tip).Disabled(opts.Disabled).
				Children(func() {
					ui.Text(c, a.mark).TextColor(k.TextMuted).SingleLine().
						FontSize(core.FontSize(c, theme.BodySize))
				})
			if !btn.Clicked() {
				continue
			}
			all := i == 1 || i == 2
			moved := transferCrossAll(opts, left, right, i < 2, all)
			r.right, r.left = r.right || (moved > 0 && i < 2), r.left || (moved > 0 && i >= 2)
			r.moved = moved
		}
	})
}

// transferCrossAll moves values across between the two slices and returns how
// many went. towards says which way, and all says whether it is the ticked
// ones or every one of them.
func transferCrossAll(opts TransferOptions, left, right *[]string, towards, all bool) int {
	// Rewritten rather than appended into: appending would write into the
	// caller's own spare capacity, which is another value's space as far as
	// the caller is concerned.
	l, r := cloneOf(*left), cloneOf(*right)
	*left, *right = l, r

	from, to := left, right
	if !towards {
		from, to = right, left
	}
	if len(*from) == 0 {
		return 0
	}
	if all {
		moved := len(*from)
		*to = append(*to, *from...)
		*from = nil
		return moved
	}
	kept := make([]string, 0, len(*from))
	for _, v := range *from {
		at := indexOfChoice(opts.Choices, v)
		if at >= 0 && opts.Selectable != nil && (*opts.Selectable)[at] {
			*to = append(*to, v)
			continue
		}
		kept = append(kept, v)
	}
	return len(*from) - len(kept)
}

// transferCross moves one named value from one side to the other and reports
// whether it went. It is the no-ticks case, where a press on a row is itself
// the move rather than a tick for an arrow to read.
func transferCross(from *[]string, value string, to *[]string) int {
	if !hasValue(*from, value) {
		return 0
	}
	*from = without(*from, value)
	*to = append(cloneOf(*to), value)
	return 1
}

// withoutAll is list with every value of gone taken out. It is without for a
// whole slice at once, which the transfer needs to reconcile a value that has
// somehow ended up on both sides.
func withoutAll(list, gone []string) []string {
	if len(gone) == 0 {
		return list
	}
	out := make([]string, 0, len(list))
	for _, v := range list {
		if hasValue(gone, v) {
			continue
		}
		out = append(out, v)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func findChoice(choices []Choice, value string) (Choice, bool) {
	for _, ch := range choices {
		if ch.Value == value {
			return ch, true
		}
	}
	return Choice{}, false
}

func indexOfChoice(choices []Choice, value string) int {
	for i, ch := range choices {
		if ch.Value == value {
			return i
		}
	}
	return -1
}

// cloneOf is a copy of a slice, so that a transfer rewrites the caller's
// slices rather than appending into them in place.
func cloneOf(list []string) []string {
	out := make([]string, len(list))
	copy(out, list)
	return out
}
