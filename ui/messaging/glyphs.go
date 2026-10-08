package messaging

import "github.com/egoist/mygo/ui"

// The marks a message channel needs on top of ui/display's set: the receipt
// ticks, the emoji panel's categories, the call's controls, the presence ring.
//
// Same reasoning as ui/media's, and the same reason it is repeated rather than
// shared: these are two packages that both had to answer "the vocabulary ran
// out", and the honest answer for each is a small addition of its own rather
// than a fourth package underneath both of them that exists only to hold ten
// glyphs.
var glyphs = map[string]string{
	// sent is one tick: the server has it.
	"sent": `<path d="m5 12.5 4 4L19 7"/>`,
	// delivered is two: it has reached the other end's device.
	"delivered": `<path d="m2.5 12.5 3.5 3.5L16 6.5"/><path d="m9.5 15.5 1.5 1.5L21 7"/>`,
	// read is two filled, so the difference from delivered survives being
	// drawn in one colour: a reader who cannot see the ticks' colour has to be
	// able to tell "they got it" from "they read it".
	"read": `<path d="m2.5 12.5 3.5 3.5L16 6.5"/><path d="m9.5 15.5 1.5 1.5L21 7"/>`,
	// failed is the one state with a word, because it is the one a reader has
	// to act on.
	"failed":    `<path d="M12 4.5 21 20H3Z"/><path d="M12 10v4.5M12 17.2v.1"/>`,
	"clock":     `<circle cx="12" cy="12" r="8"/><path d="M12 7.5V12l3 2"/>`,
	"reply":     `<path d="M9 10 4 14.5 9 19"/><path d="M4 14.5h9a6 6 0 0 0 6-6V4"/>`,
	"pin":       `<path d="M9.5 3.5h5l-.5 5 3 3.5H7l3-3.5Z"/><path d="M12 12v8.5"/>`,
	"mic":       `<rect x="9" y="3.5" width="6" height="10" rx="3"/><path d="M5.5 11a6.5 6.5 0 0 0 13 0M12 17.5V20"/>`,
	"micOff":    `<rect x="9" y="3.5" width="6" height="10" rx="3"/><path d="M5.5 11a6.5 6.5 0 0 0 13 0M12 17.5V20"/><path d="m4 4 16 16"/>`,
	"camera":    `<rect x="3.5" y="7" width="17" height="12" rx="3"/><path d="m8.5 7 1.5-2.5h4L15.5 7"/>`,
	"cameraOff": `<rect x="3.5" y="7" width="17" height="12" rx="3"/><path d="m8.5 7 1.5-2.5h4L15.5 7"/><path d="m4 4 16 16"/>`,
	"screen":    `<rect x="3.5" y="4.5" width="17" height="12" rx="2.5"/><path d="M9 20h6M12 16.5V20"/>`,
	"handup":    `<path d="M6.5 4h-2A2.5 2.5 0 0 0 2 6.5C2 14 8 20 15.5 20c2.2 0 4-1.8 4-4v-2a2 2 0 0 0-2-2h-2v-5a1.5 1.5 0 0 0-3 0V10h-1V5.5a1.5 1.5 0 0 0-3 0V10h-1V4Z"/>`,
	"smile":     `<circle cx="12" cy="12" r="8.5"/><path d="M8.5 14a4.5 4.5 0 0 0 7 0"/><path d="M9 9.5v.1M15 9.5v.1"/>`,
	"search":    `<circle cx="11" cy="11" r="6"/><path d="m15.5 15.5 4 4"/>`,
	"send":      `<path d="M4 12 20.5 4 15 20.5l-2.5-6Z"/><path d="m12.5 14.5 8-10.5"/>`,
	"mail":      `<rect x="3.5" y="5.5" width="17" height="13" rx="3"/><path d="m4.5 7.5 7.5 5.5 7.5-5.5"/>`,
	"at":        `<circle cx="12" cy="12" r="3.5"/><path d="M15.5 8.5v4.5a2.5 2.5 0 0 0 5 0V12a8.5 8.5 0 1 0-3.3 6.7"/>`,
	"thread":    `<path d="M4 6h16M4 12h10M4 18h7"/><path d="m17 10 4 4-4 4"/>`,
}

// glyph is one of the marks above, panicking on an unknown name: a blank
// receipt says "sent" by saying nothing at all, which is the worst of the
// three answers.
func glyph(name string) *ui.SVG {
	shapes, ok := glyphs[name]
	if !ok {
		panic("messaging: unknown glyph " + name)
	}
	return ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" ` +
		`fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" ` +
		`stroke-linejoin="round">` + shapes + `</svg>`))
}
