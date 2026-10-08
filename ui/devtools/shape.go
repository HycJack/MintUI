package devtools

import (
	"strconv"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
)

// monoFamily is the monospaced face, asked for by name because MyGo takes the
// family as a string.
//
// It is used for every value, key, path, header and duration in this package
// and for nothing else. The reason is that all of those are things people
// compare character by character — two ids, two status codes, two durations —
// and a proportional face makes "I" and "1" and "l" the same width, which is
// the one comparison a dashboard of numbers is constantly asking for.
const monoFamily = "monospace"

// unit is the density's spacing step. The helpers here that only need a gap
// say it once rather than spelling out core.Density(c).Unit() at each site.
func unit(c *ui.Context) float32 { return core.Density(c).Unit() }

// itoa is a non-negative int as text. The numbers this package prints are
// counts — lines, headers, keys, days — and none of them is large enough for
// strconv's formatting to matter, so one small wrapper keeps every call site
// from importing it for a single call.
func itoa(n int) string { return strconv.Itoa(n) }

// The two marks this package needs as glyphs. They are the same 24×24 stroke
// style as the rest of the interface, so a mark here sits beside one from
// display without either looking borrowed.
//
// They are package values rather than drawn inline because a copy button
// appears in several trees and three copies of one cross is three chances for
// them to be three different crosses.
var (
	glyphCopy  = glyph(`<rect x="8.5" y="8.5" width="11" height="11" rx="2.5"/><path d="M15.5 5.5H6.5a2 2 0 0 0-2 2v9"/>`)
	glyphClose = glyph(`<path d="m6.5 6.5 11 11M17.5 6.5l-11 11"/>`)
	glyphTick  = glyph(`<path d="m5 12.5 4.5 4.5L19 7"/>`)
	glyphWarn  = glyph(`<path d="M12 4.5 21 20H3Z"/><path d="M12 10v4.5M12 17.2v.1"/>`)
)

// glyph parses one mark.
func glyph(shapes string) *ui.SVG {
	return ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" ` +
		`fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" ` +
		`stroke-linejoin="round">` + shapes + `</svg>`))
}
