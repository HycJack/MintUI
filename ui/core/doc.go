// Package core resolves one window's appearance. It is the only place in the
// library that touches MyGo's system theme.
//
// A view calls [Use] once per frame, at the top, before anything else:
//
//	core.Use(c, core.Settings{})
//
// From then on [Tokens] gives the palette for this frame and [Density] gives
// the spacing step every component multiplies. Reading either before Use
// panics: a view that forgets the call would otherwise draw in a fallback
// palette, which on a dark desktop looks like a rendering fault rather than a
// missing line.
//
// Use also themes MyGo itself, so a text field or a native scroll bar matches
// the palette the components are drawing.
package core
