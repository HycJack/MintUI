package data

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// Term is one pair of a DescriptionList: what is being described, and what
// it is.
type Term struct {
	// Term says what is being described.
	Term string
	// Value is the description.
	Value string
	// Body draws the description instead of Value, for the pair whose value
	// is not text: an avatar, a control, a figure among figures.
	Body func()
}

// DescriptionListOptions configure a DescriptionList.
type DescriptionListOptions struct {
	// TermWidth is the width of the column of terms, so that every value
	// starts in the same place. Zero takes ShareTerm.
	TermWidth float32
	// ShareTerm is how much of the width the terms take when TermWidth is
	// zero: 0.3 is three parts in ten. It is a share rather than a number of
	// DIPs because a description list is read in panels of every width, and
	// the column of terms belongs at a different width in each of them.
	ShareTerm float32
	// Columns is how many pairs there are across. One is a column of pairs,
	// which is what a narrow panel wants; two puts a long list in two.
	Columns int
	// Rules draws a hairline under each pair, for a list of values where the
	// eye needs the pairs to stay apart.
	Rules bool
	// MutedValues draws the values in the muted tone, for a list whose terms
	// are the point rather than its values.
	MutedValues bool
}

// DescriptionList is pairs of a name and what it names: the rows of a
// settings page, the details under a summary, the facts beside a figure.
//
// The terms are in the muted tone at the meta size and the values in the
// ordinary one at the body size, because a description list is read by
// looking for the value and glancing at what it belongs to — not the other
// way round.
func DescriptionList(c *ui.Context, opts DescriptionListOptions, terms ...Term) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if opts.Columns < 1 {
		opts.Columns = 1
	}
	share := opts.ShareTerm
	if share <= 0 {
		share = 0.3
	}
	valueCol := k.Text
	if opts.MutedValues {
		valueCol = k.TextMuted
	}

	// One pair: the term in a column of its own, the value taking the rest.
	// Both are stretched to the height of the taller of the two, so that a
	// value of two lines keeps its term beside the first rather than
	// floating at the top.
	pair := func(t Term) {
		ui.Column(c).FillWidth().Shrink(0).Children(func() {
			ui.Row(c).FillWidth().Shrink(0).AlignItems(ui.Stretch).
				Padding(u, u*1.5).Children(func() {
				name := ui.Text(c, t.Term).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.MetaSize)).Grow(1)
				if opts.TermWidth > 0 {
					name.Width(opts.TermWidth).Shrink(0)
				} else {
					name.WidthPercent(share * 100).Shrink(0)
				}
				ui.Box(c).Grow(1).Children(func() {
					switch {
					case t.Body != nil:
						t.Body()
					case t.Value != "":
						// No Grow on the value: the box is a column, and a
						// value grown into height it does not have lays out
						// at zero tall and paints over the row's other cells.
						ui.Text(c, t.Value).TextColor(valueCol)
					}
				})
			})
			if opts.Rules {
				layout.Divider(c, layout.DividerOptions{})
			}
		})
	}

	if opts.Columns == 1 {
		return ui.Column(c).FillWidth().Shrink(0).Children(func() {
			for _, t := range terms {
				pair(t)
			}
		})
	}
	// Across several columns: each pair takes its own share of the row, so a
	// pair is never squeezed by the pair beside it, and the row wraps when
	// the window is too narrow for them all.
	return ui.Row(c).FillWidth().Wrap().Children(func() {
		for _, t := range terms {
			ui.Box(c).WidthPercent(100 / float32(opts.Columns)).Shrink(0).
				Children(func() { pair(t) })
		}
	})
}
