package chart

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// Mark is the little glyph a legend entry carries. It should match what the
// series draws — a dot for a line chart's points, a square for bars — so the
// legend is a key to the chart rather than a list of names beside it.
type Mark uint8

const (
	// MarkSquare is a filled square, for bars and areas.
	MarkSquare Mark = iota
	// MarkDot is a filled circle, for a series drawn as points.
	MarkDot
	// MarkLine is a short thick rule, for a series drawn as a line.
	MarkLine
)

// LegendEntry is one series as the legend names it.
type LegendEntry struct {
	// Name is what the series is called.
	Name string
	// Color is the series' colour: the same one its marks are drawn with, so
	// the legend is a key rather than a caption.
	Color ui.Color
	// Mark is the glyph beside the name; the square unless it is set.
	Mark Mark
}

// EntriesOf is the legend of a set of series: each one named and coloured,
// with the colour from the window's palette where the series did not choose
// one. A chart's series and its legend then cannot disagree about which
// colour is which, because there is only one place that decides.
func EntriesOf(c *ui.Context, series []Series) []LegendEntry {
	if len(series) == 0 {
		return nil
	}
	palette := Palette(c, len(series))
	out := make([]LegendEntry, len(series))
	for i, s := range series {
		color := s.Color
		if color == (ui.Color{}) {
			color = palette[i]
		}
		out[i] = LegendEntry{Name: s.Name, Color: color, Mark: markOf(s.Form)}
	}
	return out
}

func markOf(form Form) Mark {
	switch form {
	case FormPoint:
		return MarkDot
	case FormBar:
		return MarkSquare
	case FormArea, FormLine:
		return MarkLine
	}
	return MarkSquare
}

// LegendOptions configure a [Legend].
type LegendOptions struct {
	// Entries are the series being named, left to right or top to bottom.
	Entries []LegendEntry
	// Vertical stacks the entries instead of laying them in a row, for a
	// legend beside a chart rather than above it.
	Vertical bool
	// Size is the name's font size; zero uses [theme.MetaSize].
	Size float32
	// Swatch is the glyph's size; zero uses one density step times three.
	Swatch float32
}

// Legend is the chart's key: a mark and a name per series, in a row or a
// column.
//
// It is the one piece of a chart that is real elements rather than drawing —
// its names are text a screen reader can read, which is what a legend is for
// when the marks themselves are only colours.
func Legend(c *ui.Context, opts LegendOptions) *ui.Element {
	if len(opts.Entries) == 0 {
		// A legend with nothing in it is not a mistake: a chart whose
		// series are all named on the axes has no legend to draw.
		return ui.Box(c)
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	size := opts.Size
	if size <= 0 {
		size = theme.MetaSize
	}
	swatch := opts.Swatch
	if swatch <= 0 {
		swatch = u * 3
	}

	row := ui.Row(c)
	if opts.Vertical {
		row = ui.Column(c)
	}
	return row.Gap(u * 3).AlignItems(ui.Center).Label("Series").Children(func() {
		for _, entry := range opts.Entries {
			ui.Row(c).Gap(u * 1.5).AlignItems(ui.Center).Label(entry.Name).Children(func() {
				mark := ui.Box(c).Size(swatch, swatch).Shrink(0)
				switch entry.Mark {
				case MarkDot:
					mark.Radius(swatch / 2)
				case MarkLine:
					mark.Height(max(swatch/3, 1)).Radius(swatch / 6)
				case MarkSquare:
					mark.Radius(u * 0.75)
				}
				mark.Background(entry.Color)
				ui.Text(c, entry.Name).TextColor(k.TextMuted).FontSize(size)
			})
		}
	})
}

// LegendHeight is how much room a legend takes across: the height of a row of
// entries, or of the whole column when it is a vertical one, which a frame
// needs to keep out of its plot area.
func LegendHeight(c *ui.Context, opts LegendOptions) float32 {
	u := core.Density(c).Unit()
	size := opts.Size
	if size <= 0 {
		size = theme.MetaSize
	}
	one := max(LabelHeight(c, size), opts.Swatch)
	if one <= 0 {
		one = u * 4
	}
	if opts.Vertical {
		n := len(opts.Entries)
		return one*float32(n) + u*2*float32(max(n-1, 0))
	}
	return one
}
