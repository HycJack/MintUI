package feedback

import (
	"strconv"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// CostRow is one line of a CostBreakdown: what was charged, for how much.
type CostRow struct {
	// Name is what the charge is for: "Model", "Storage", "Tax". It is
	// required: an amount with no charge beside it cannot be checked
	// against the invoice it came from.
	Name string
	// Amount is how much, in the breakdown's Unit, already in the unit's
	// own scale — dollars in dollars, not cents.
	Amount float64
	// Muted draws the row in the secondary ink, for lines that are part of
	// the arithmetic rather than a charge of their own: an offset, a
	// rounding line, a figure the reader is being shown for completeness.
	Muted bool
}

// CostBreakdownOptions configure a CostBreakdown.
type CostBreakdownOptions struct {
	// Rows are the charges, in the order to show them. The order is the
	// caller's, the way a Notification's newest-first order is its caller's:
	// the library does not guess whether the biggest charge or the most
	// recent one goes first.
	Rows []CostRow
	// ShowTotal appends a total line under a rule, in the window's text and
	// bold, which is what marks it as the figure to argue about. The total
	// is the sum of the rows as given — including the muted ones, because
	// a total that skips a line the reader can see is a total that cannot
	// be checked.
	ShowTotal bool
	// Unit is set before each amount, like "$" or "¥" (or "CNY " with its
	// own space). Empty prints the bare number, for a breakdown of a
	// quantity that is not money: credits, minutes, seats.
	Unit string
	// Width bounds the breakdown; zero lets it fill its parent.
	Width float32
}

// CostBreakdown is a list of what a thing cost, one line per charge, with
// an optional total set off under a rule.
//
// It is a breakdown and not a table because a table implies a column of
// figures you might add up yourself, and a breakdown says the addition has
// already been done: the total line is the answer, and the rule above it is
// the line the argument crosses. A caller with more to say — per-row detail
// on a press, a column of dates — has a real table to build, and this is
// not it.
//
// It draws on the surface rather than a card, the way a NotificationCenter
// does, because it is content inside a panel that already has a surface.
func CostBreakdown(c *ui.Context, opts CostBreakdownOptions) *ui.Element {
	if len(opts.Rows) == 0 {
		panic("feedback: CostBreakdown needs at least one Row; a breakdown " +
			"of nothing is a panel with no number in it")
	}
	for _, row := range opts.Rows {
		if row.Name == "" {
			panic("feedback: a cost row needs a Name; an amount with no " +
				"charge beside it cannot be checked")
		}
	}
	u := core.Density(c).Unit()

	total := 0.0
	for _, row := range opts.Rows {
		total += row.Amount
	}

	col := ui.Column(c).FillWidth().Gap(u * 1.5).Role(ui.RoleTable)
	if opts.Width > 0 {
		col.Width(opts.Width)
	}
	col.Children(func() {
		for _, row := range opts.Rows {
			costRow(c, row.Name, costAmount(row.Amount, opts.Unit), row.Muted,
				false)
		}
		if opts.ShowTotal {
			layout.Divider(c, layout.DividerOptions{})
			costRow(c,
				core.Msg(c, "feedback.costBreakdown.total", "Total"),
				costAmount(total, opts.Unit), false, true)
		}
	})
	return col
}

// costRow draws one line of a breakdown: the charge on the left, the figure
// on the right, the room between them doing the work of lining the two up.
func costRow(c *ui.Context, name, amount string, muted, bold bool) {
	k, u := core.Tokens(c), core.Density(c).Unit()

	fg, amtFg := k.Text, k.Text
	if muted {
		// A muted row is secondary content, not disabled content: the name
		// takes the muted ink, the figure the faint one, so a line of
		// offsets reads as quieter than a line of charges without either
		// reading as gone.
		fg, amtFg = k.TextMuted, k.TextFaint
	}
	if bold {
		amtFg = k.Text
	}

	ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 2).
		Label(name).Children(func() {
		t := ui.Text(c, name).TextColor(fg).
			FontSize(core.FontSize(c, theme.RowSize))
		if bold {
			t.Bold()
		}
		ui.Box(c).Grow(1)
		a := ui.Text(c, amount).TextColor(amtFg).
			FontSize(core.FontSize(c, theme.RowSize)).SingleLine()
		if bold {
			a.Bold()
		}
	})
}

// costAmount renders a figure the way an invoice does it: the unit set
// before two decimal places, "$1.20". A caller whose unit is not money and
// does not want decimals passes an empty Unit and reads the bare number.
func costAmount(v float64, unit string) string {
	return unit + strconv.FormatFloat(v, 'f', 2, 64)
}
