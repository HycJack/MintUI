// Package finance is the market half of the interface: quotes, order books,
// positions, orders, risk and the charts that go with all of them.
//
// # Charts come from ui/chart and tables from ui/data
//
// There is not one chart drawn in this package. Every plot here — the
// performance curve, the allocation ring, the payoff diagram, the earnings
// calendar, the market heatmap — is a [chart] component with this package's
// data in it, and every table is [data.DataTable]. That is not a shortcut
// around building them; it is the reason they are correct. A hand-drawn chart
// has no measured gutters, so its axis labels collide at one window size and
// not another, and a hand-drawn table has no rule about what a column does when
// the window is too narrow. Fifty-four chart types already answered both
// questions once, and answering them a second time here would be fifty-four
// ways to be wrong by a pixel.
//
// What this package owns is everything around a chart: the formatting, the
// shapes, the calculations and the controls.
//
// # The pure functions are the contract
//
// [FormatPrice], [FormatChange], [FormatVolume] and [Levels] need no window
// and carry the rules a trading screen is most easily got wrong. A price that
// is 1 instead of 1.00 is a price nobody can scan; a change that reads
// +0.00% when nothing changed is a lie the reader acts on; a volume that
// reads 1000 instead of 1K in a column of the same width is a column the eye
// has to re-measure. Each is tested on exactly what it returns rather than on
// what a bar came out looking like.
//
// # Money is the monospaced face
//
// Every figure in this package is drawn in the monospaced stack rather than in
// the proportional one, and that is not a styling preference. A column of
// right-aligned prices is read by the position of each figure's last digit, and
// in a proportional face "1" is a different width from "8" — so the digits line
// up at the right edge and not at the decimal point, and the column looks
// ragged however carefully it is laid out. Monospaced digits put both edges on
// the same grid.
//
// [MonoFont] is that stack. ui/theme does not carry one — it carries sizes,
// not faces — so it is declared here rather than reaching into ui/code, whose
// stack is for source code and whose package a quote card has no business
// importing.
//
// # Nothing here trades
//
// There is no order router, no exchange client and no position store. An order
// entry reports what the reader filled in and the caller decides whether to
// send it. A component that could place a trade could not be shown a
// confirmation, and could not be drawn in a test at all.
package finance

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
)

// MonoFont is the stack every figure in this package is drawn in.
//
// A stack rather than a name, because a name is a guess on every platform: SF
// Mono on a Mac, Menlo on one that has it, Cascadia on a Windows that has it,
// and the system's own monospace everywhere else.
//
// It is here rather than in ui/theme because theme holds colours, spacing,
// radii and *sizes*, and a face is none of those — and it is here rather than
// borrowed from ui/code because that package's stack is for source code, whose
// line numbers and columns are its own business. A quote card importing a code
// viewer to get a font is a dependency nobody could explain.
const MonoFont = "SF Mono, Menlo, Consolas, Cascadia Mono, monospace"

// mono is one run of figures in this package's face, which is what every
// number here goes through. It is a helper rather than a literal at each site
// because a face set on one run of text and not the next is how a column ends
// up half monospaced.
func mono(c *ui.Context, text string, size float32, col ui.Color) *ui.Element {
	return ui.Text(c, text).Font(MonoFont).FontSize(core.FontSize(c, size)).TextColor(col)
}
