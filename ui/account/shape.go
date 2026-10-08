package account

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/layout"
)

// panelOptions is the little that every floating panel in this package needs,
// which is the least of layout.PanelOptions: these panels hang under a
// trigger and carry a list, so they are always compact and never titled.
type panelOptions struct {
	// label names the panel. It is required: a menu that is not called
	// anything announces as a list of buttons with nothing saying what it is
	// a list of.
	label string
	// body fills it.
	body func()
	// compact wears the smaller radius and the row-size heading, which is
	// what a menu wants as against a sheet.
	compact bool
}

// panel dresses a popover's panel element with the library's one floating
// face: the radius, the hairline, the shadow and the padding.
//
// It is a thin call on layout.Panel rather than a second implementation,
// because a switcher menu that was half a DIP off the one a Select shows
// would read as a bug rather than as a menu. See ui/layout/panel.go on why
// Panel lives in layout and is reachable from everywhere.
func panelFace(c *ui.Context, host *ui.Element, opts panelOptions) {
	if opts.label == "" {
		panic("account: a floating panel needs a name; an unnamed menu is a " +
			"list of buttons with nothing saying what it is a list of")
	}
	layout.Panel(c, host, layout.PanelOptions{
		Label:   opts.label,
		Compact: opts.compact,
	}, func() {
		if opts.body != nil {
			opts.body()
		}
	})
}

// glyph is one mark of this package's own, parsed. The built-in set in
// display is reached through display.Icon, which hands back an element rather
// than the glyph itself, so a mark that has to go into an input.IconButton —
// which takes a *ui.SVG — has to be spelled out here.
//
// They are the same 24×24 stroke style as the rest of the interface, so a
// glyph drawn here next to one from display is one mark of the same hand.
func glyph(shapes string) *ui.SVG {
	return ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" ` +
		`fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" ` +
		`stroke-linejoin="round">` + shapes + `</svg>`))
}

// The marks this package needs and display cannot hand over as a glyph.
var (
	glyphKey     = glyph(`<circle cx="8" cy="12" r="4"/><path d="M12 12h8M17 12v3M20 12v2"/>`)
	glyphClock   = glyph(`<circle cx="12" cy="12" r="8"/><path d="M12 7.5V12l3 2"/>`)
	glyphChevron = glyph(`<path d="m9.5 6.5 5.5 5.5-5.5 5.5"/>`)
	glyphGear    = glyph(`<circle cx="12" cy="12" r="3"/>` +
		`<path d="M12 3.5v2.2M12 18.3v2.2M20.5 12h-2.2M5.7 12H3.5M18 6l-1.6 1.6M7.6 16.4 6 18M18 18l-1.6-1.6M7.6 7.6 6 6"/>`)
	glyphRefresh = glyph(`<path d="M19 12a7 7 0 1 1-2.2-5.1"/><path d="M19.5 4v4h-4"/>`)
	glyphTeam    = glyph(`<circle cx="9" cy="8.5" r="3"/>` +
		`<path d="M3.5 19.5c.6-3.2 2.8-4.8 5.5-4.8s4.9 1.6 5.5 4.8"/>` +
		`<path d="M16 6.2a3 3 0 0 1 0 5.6M17.5 15.2c2 .7 3.2 2.2 3.6 4.3"/>`)
	glyphStar    = glyph(`<path d="m12 4 2.5 5.2 5.5.8-4 3.9 1 5.6-5-2.7-5 2.7 1-5.6-4-3.9 5.5-.8Z"/>`)
	glyphLaptop  = glyph(`<rect x="4" y="5.5" width="16" height="10" rx="2"/><path d="M2.5 19h19"/>`)
	glyphPhone   = glyph(`<path d="M6 4h3l1.5 4-2 1.4a12 12 0 0 0 6.1 6.1L16 13.5l4 1.5v3a2 2 0 0 1-2.2 2A15.5 15.5 0 0 1 4 6.2 2 2 0 0 1 6 4Z"/>`)
	glyphSliders = glyph(`<path d="M4 6h9M17 6h3M4 12h3M11 12h9M4 18h7M15 18h5"/>` +
		`<circle cx="15" cy="6" r="2"/><circle cx="9" cy="12" r="2"/><circle cx="13" cy="18" r="2"/>`)
	glyphQuestion = glyph(`<circle cx="12" cy="12" r="8"/><path d="M9.8 9.6a2.3 2.3 0 1 1 3 2.2v1.4M12 16.4v.1"/>`)
	glyphDismiss  = glyph(`<path d="m6.5 6.5 11 11M17.5 6.5l-11 11"/>`)
	glyphSend     = glyph(`<path d="M4 12 20 5l-4 15-4-6Z"/><path d="m12 14 8-9"/>`)
)
