package core

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/theme"
)

// msgKey is where a window's copy overrides live on the root element.
type msgKey struct{}

// Msg returns the string a component shows for key: the window's override if
// one is installed, otherwise def.
//
// Library copy goes through here so a window can be localised or reworded
// without touching the library. Business copy never does — "Riverside Clinic"
// is the caller's string, not the library's.
func Msg(c *ui.Context, key, def string) string {
	m := ui.Local(c.Root(), msgKey{}, func() map[string]string { return map[string]string{} })
	if s, ok := (*m)[key]; ok {
		return s
	}
	return def
}

// WithMessages installs the window's copy overrides and returns s, so it can
// wrap Use in a single expression.
//
//	core.Use(c, core.Settings{})
//	c = core.WithMessages(c, map[string]string{"input.required": "必填"})
//
// Call it in the same frame as Use: Use replaces the window's state, and the
// messages hang off it.
func WithMessages(c *ui.Context, m map[string]string) *ui.Context {
	// Through the stored pointer, not through init: ui.Local skips init once
	// the key exists, so a window whose copy changes between frames would
	// keep the map it was handed first, for as long as it lived.
	*ui.Local(c.Root(), msgKey{}, func() map[string]string { return nil }) = m
	return c
}

// msgMap returns the window's copy overrides as a value, creating the map the
// first time. MyGo's Local hands back a pointer so the store can hold one
// address per key; everything here works on the map behind it.

// Messages returns the window's copy overrides, never nil.
func Messages(c *ui.Context) map[string]string {
	m := ui.Local(c.Root(), msgKey{}, func() map[string]string { return map[string]string{} })
	return *m
}

// Local and SetLocal wrap MyGo's own per-element state, which is keyed on the
// root element and so survives a frame. Storing a pointer per key is what makes
// a second lookup find the first one's value.

// Local returns this frame's value for k, creating it with init the first
// time. It is how a component keeps something across frames without owning a
// struct: the value lives on the window root and is looked up by key.
//
// A component that needs this takes a typed key rather than a string, so two
// components cannot collide by naming.
func Local[T any](c *ui.Context, k any, init func() T) T {
	v := ui.Local(c.Root(), k, init)
	return *v
}

// SetLocal stores v for k. A component that must report a change upward does
// it in the frame after the user acted: mutate, SetLocal, then invalidate.
func SetLocal[T any](c *ui.Context, k any, v T) {
	// Through the stored pointer, not through init: ui.Local skips init once
	// the key exists, so a setter would quietly do nothing to any key that
	// Local has already been asked about.
	*ui.Local(c.Root(), k, func() T { return v }) = v
}

// Motion returns a duration in milliseconds, scaled by the window's motion
// preference. Durations in the library are written as plain constants and pass
// through here, so honouring "reduce motion" is one call rather than a check
// at every animated site.
func Motion(c *ui.Context, d float32) float32 {
	if Reduced(c) {
		return 0
	}
	return d
}

// override is the window's own answer for k, or else the one the desktop
// gave. It is a lookup rather than a first-writer-wins slot because a reader
// must not settle the question merely by asking first: the setters below write
// through the same slot, and a test or a window with no desktop behind it has
// to be able to answer after anything has already read.
func override[T any](c *ui.Context, k any, desktop T) T {
	if v := *ui.Local(c.Root(), k, func() *T { return nil }); v != nil {
		return *v
	}
	return desktop
}

// setOverride answers k for the rest of the window's life, which is what makes
// an explicit setting beat whatever the desktop says.
func setOverride[T any](c *ui.Context, k any, v T) {
	*ui.Local(c.Root(), k, func() *T { return nil }) = &v
}

// Reduced reports whether the desktop asked for reduced motion. Components that
// animate check it through Motion; a component that would otherwise animate
// something that only exists in motion — a spinner, a caret — should read it
// directly and draw its resting state instead.
func Reduced(c *ui.Context) bool {
	return override(c, reducedKey{}, c.Preferences().ReduceMotion)
}

// WithReducedMotion sets the window's motion preference for tests and for a
// window that is not a real desktop window. It wins over the desktop's setting.
func WithReducedMotion(c *ui.Context, on bool) *ui.Context {
	setOverride(c, reducedKey{}, on)
	return c
}

type reducedKey struct{}

// FontSize returns the desktop's text scale applied to a library size, so a
// component that sets FontSize(theme.RowSize) still follows the system setting
// instead of ignoring it.
func FontSize(c *ui.Context, size float32) float32 {
	scale := override(c, scaleKey{}, c.Preferences().TextScale)
	if scale <= 0 {
		return size
	}
	return size * scale
}

// WithTextScale sets the desktop text scale for tests. Like WithReducedMotion it
// wins over the desktop's setting.
func WithTextScale(c *ui.Context, scale float32) *ui.Context {
	setOverride(c, scaleKey{}, scale)
	return c
}

type scaleKey struct{}

// ControlHeight is the height of a standard control — a button, a text field,
// a select — at this window's density. Components size themselves from it
// rather than each keeping its own number, so Comfortable really does make
// every control taller.
func ControlHeight(c *ui.Context) float32 {
	switch Density(c) {
	case theme.Comfortable:
		return 36
	default:
		return 30
	}
}

// Def is the default copy for a component that shows one. It exists so a
// component can be read top to bottom: Def tells you what it says when nobody
// overrides it.
func Def(s string) string { return s }
