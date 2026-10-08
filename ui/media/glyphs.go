package media

import "github.com/egoist/mygo/ui"

// The marks a transport needs that ui/display's set does not carry: play,
// pause, the skip pair, the record dot, the crop frame, the pen.
//
// They are in this package's own 24×24 stroke style — one weight, round caps —
// so that a row of transport buttons is four marks of one hand rather than
// three of this library's and one of somewhere else. They are *ui.SVG rather
// than display.IconName because display.Icon is a closed set and a closed set
// is the right answer for a board's navigation and the wrong one for a
// transport: a play button that renders as a square because the vocabulary ran
// out is worse than a four-line addition here.
var glyphs = map[string]string{
	"play":       `<path d="m8.5 5.5 10 6.5-10 6.5Z"/>`,
	"pause":      `<path d="M9 5.5v13M15 5.5v13"/>`,
	"back":       `<path d="M18 6v12L9 12Z"/><path d="M6 6v12"/>`,
	"forward":    `<path d="M6 6v12l9-6Z"/><path d="M18 6v12"/>`,
	"record":     `<circle cx="12" cy="12" r="5.5" fill="currentColor" stroke="none"/>`,
	"stop":       `<rect x="7" y="7" width="10" height="10" rx="2"/>`,
	"mute":       `<path d="M4 9.5h3.5L12 6v12L7.5 14.5H4Z"/><path d="m16 9.5 4 5M20 9.5l-4 5"/>`,
	"volume":     `<path d="M4 9.5h3.5L12 6v12L7.5 14.5H4Z"/><path d="M15.5 9a4 4 0 0 1 0 6M18 6.5a7.5 7.5 0 0 1 0 11"/>`,
	"volumeDown": `<path d="M4 9.5h3.5L12 6v12L7.5 14.5H4Z"/><path d="M15.5 9a4 4 0 0 1 0 6"/>`,
	"camera":     `<rect x="3.5" y="7" width="17" height="12" rx="3"/><path d="m8.5 7 1.5-2.5h4L15.5 7"/>`,
	"screen":     `<rect x="3.5" y="4.5" width="17" height="12" rx="2.5"/><path d="M9 20h6M12 16.5V20"/>`,
	"mic":        `<rect x="9" y="3.5" width="6" height="10" rx="3"/><path d="M5.5 11a6.5 6.5 0 0 0 13 0M12 17.5V20"/>`,
	"crop":       `<path d="M6.5 2.5v15h15M2.5 6.5h15v15"/>`,
	"pen":        `<path d="M4.5 19.5h4L19 9a2.1 2.1 0 0 0-3-3L5.5 16.5Z"/>`,
	"split":      `<path d="M12 4v16"/><path d="m6.5 10.5 2-3 2 3M6.5 13.5l2 3 2-3"/>`,
	"closed":     `<path d="m6.5 6.5 11 11M17.5 6.5l-11 11"/>`,
	"captions":   `<rect x="3.5" y="5" width="17" height="14" rx="3"/><path d="M10 10.5a2.5 2.5 0 1 0 0 3M17 10.5a2.5 2.5 0 1 0 0 3"/>`,
}

// glyph is one of the marks above. An unknown name panics rather than drawing
// nothing: a blank transport button is a control that looks broken, and a
// broken control in a test is indistinguishable from a component that was
// never drawn at all.
func glyph(name string) *ui.SVG {
	shapes, ok := glyphs[name]
	if !ok {
		panic("media: unknown glyph " + name)
	}
	return ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" ` +
		`fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" ` +
		`stroke-linejoin="round">` + shapes + `</svg>`))
}

// transportGlyph is the mark for a play button, which is two different shapes
// rather than one shape that changes colour. The triangle and the square are
// the two states of the same control and every player in the world draws them
// that way; a button whose colour changed would need a legend.
func transportGlyph(playing bool) *ui.SVG {
	if playing {
		return glyph("pause")
	}
	return glyph("play")
}

// transportName is what the play button says out loud. It says what the press
// will do, not what the button currently is — "Pause" beside a triangle is
// read as a button that pauses, which is not what it does.
func transportName(playing bool) string {
	if playing {
		return "Pause"
	}
	return "Play"
}
