package input

import (
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// Segmented is a row of mutually exclusive choices, one of them selected.
//
// The selection is a value the caller owns: point at it and the control
// follows, which keeps it a plain part of app state rather than something the
// widget hides.
func Segmented(c *ui.Context, selected *int, labels ...string) *ui.Element {
	if len(labels) == 0 {
		panic("input: Segmented needs at least one label")
	}
	if selected == nil {
		panic("input: Segmented needs a selection to point at")
	}
	k := core.Tokens(c)
	if *selected < 0 || *selected >= len(labels) {
		panic("input: Segmented selection is out of range")
	}

	row := ui.Segmented(c, selected, labels...)
	return row.Background(k.Surface).TextColor(k.Text)
}

// SearchField is a text input shaped for a query, with a leading glyph.
//
// The text is the caller's string, so a filtered list can be recomputed as it
// is typed without a round trip through the widget.
func SearchField(c *ui.Context, query *string, placeholder string) *ui.Element {
	if query == nil {
		panic("input: SearchField needs a query to point at")
	}
	k := core.Tokens(c)
	field := ui.SearchField(c, query).Placeholder(placeholder).
		Label(placeholder).FillWidth()
	if strings.TrimSpace(*query) == "" {
		field.TextColor(k.TextMuted)
	}
	return field
}

// SwitchOptions configure a Switch.
type SwitchOptions struct {
	// Label is the text beside the switch; empty hides it.
	Label string
	// LabelFirst puts the text before the switch, for a control that reads
	// top-down rather than left-to-right.
	LabelFirst bool
}

// Switch is an on/off control. It changes the bool it points at the moment the
// user presses it, so no handler is needed for the value itself.
func Switch(c *ui.Context, on *bool, opts SwitchOptions) *ui.Element {
	if on == nil {
		panic("input: Switch needs a value to point at")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	sw := func() *ui.Element {
		s := ui.Switch(c, on)
		if opts.Label != "" {
			s = s.Label(opts.Label)
		}
		return s
	}
	if opts.Label == "" || !opts.LabelFirst {
		return sw()
	}
	return ui.Row(c).AlignItems(ui.Center).Gap(u * 2).Label(opts.Label).Children(func() {
		ui.Text(c, opts.Label).TextColor(k.Text).FontSize(theme.RowSize)
		sw()
	})
}
