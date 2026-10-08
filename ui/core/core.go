// Package core holds what every other package needs to share: the window's
// settings, its palette, and the small types components speak in.
//
// A view calls Use once, at the top of the frame; everything drawn below
// resolves against the settings that call installed.
package core

import (
	"fmt"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/theme"
)

// Mode chooses a window's appearance.
type Mode int

const (
	// System follows the desktop, re-reading it as it changes.
	System Mode = iota
	Light
	Dark
)

func (m Mode) String() string {
	switch m {
	case Light:
		return "Light"
	case Dark:
		return "Dark"
	}
	return "System"
}

// Settings is a window's theme. A nil token set uses the library default for
// that appearance, so a window only overrides what it means to change.
type Settings struct {
	// Mode is the appearance; System follows the desktop.
	Mode Mode
	// Density scales spacing, not type.
	Density theme.Density
	// Light and Dark override the library palettes when set.
	Light *theme.Tokens
	Dark  *theme.Tokens
}

// window is what Use resolves for one frame.
type window struct {
	tokens  theme.Tokens
	density theme.Density
	set     bool
}

type key struct{}

// Use applies s to the window and must be called first in every frame of the
// view. Calling it once at the top of the view is the convention.
//
// It re-themes MyGo itself as well as the library, so system widgets — text
// inputs, selects, menus — come out in the same palette as everything drawn
// here. Without that a form would be half library-coloured and half desktop.
//
// An invalid Mode or Density panics: a window rendering in an appearance
// nobody chose is worse than a crash.
func Use(c *ui.Context, s Settings) {
	if s.Mode < System || s.Mode > Dark {
		panic(fmt.Sprintf("core: invalid Mode %d", s.Mode))
	}
	if s.Density < theme.Compact || s.Density > theme.Comfortable {
		panic(fmt.Sprintf("core: invalid Density %d", s.Density))
	}

	dark := c.Theme().Dark
	switch s.Mode {
	case Light:
		dark = false
	case Dark:
		dark = true
	}

	k := theme.Light()
	if s.Light != nil {
		k = *s.Light
	}
	if dark {
		k = theme.Dark()
		if s.Dark != nil {
			k = *s.Dark
		}
	}

	density := s.Density

	c.SetTheme(&ui.Theme{
		Dark:           k.IsDark(),
		Background:     k.Background,
		Surface:        k.Surface,
		SurfaceHover:   k.SurfaceHover,
		SurfacePressed: k.SurfacePressed,
		Border:         k.Border,
		Text:           k.Text,
		TextMuted:      k.TextMuted,
		Accent:         k.Accent,
		AccentHover:    k.Accent,
		AccentPressed:  k.Accent,
		AccentText:     k.OnFill,
		Danger:         k.Danger,
		Warning:        k.Warning,
		Success:        k.Success,
		Selection:      k.SurfacePressed,
		Focus:          k.Accent,
		Scrollbar:      k.TextMuted.Alpha(0.5),
		ScrollbarWidth: 10,
		Radius:         theme.ControlRadius,
		Spacing:        density.Unit(),
	})

	*ui.Local(c.Root(), key{}, func() window { return window{} }) =
		window{tokens: k, density: density, set: true}
}

// Tokens returns the window's palette. It falls back to the light one when a
// view never called Use, so a component can be rendered on its own in a test.
func Tokens(c *ui.Context) theme.Tokens {
	return useWindow(c).tokens
}

// useWindow returns the window's resolved settings, panicking when the view
// never called Use. Falling back to a default palette would look like a
// rendering bug on a dark desktop instead of the missing call it is.
func useWindow(c *ui.Context) *window {
	w := ui.Local(c.Root(), key{}, func() window { return window{} })
	if !w.set {
		panic("core: a view read the theme before core.Use(c, Settings{})")
	}
	return w
}

// Density returns the window's density. Callers usually want its Unit rather
// than the value itself: it is the one number every gap in the library is
// built from.
func Density(c *ui.Context) theme.Density { return useWindow(c).density }

// IsDark reports whether the window is resolving the dark palette. Components
// use it where a colour is not simply a token, such as a mark that has to keep
// its contrast against whatever it sits on.
func IsDark(c *ui.Context) bool { return Tokens(c).IsDark() }

// Severity ranks how loud a status is, so pills and badges agree on which
// background goes with which word.
type Severity int

const (
	// Neutral is ordinary content: counts, labels, a medium priority.
	Neutral Severity = iota
	// Accent is the system's highlight.
	Accent
	// Success is good news.
	Success
	// Warning needs attention.
	Warning
	// Danger is a failure or an escalation.
	Danger
)

func (s Severity) String() string {
	switch s {
	case Accent:
		return "Accent"
	case Success:
		return "Success"
	case Warning:
		return "Warning"
	case Danger:
		return "Danger"
	}
	return "Neutral"
}

// Pair returns the background and foreground a severity is drawn with.
func (s Severity) Pair(k theme.Tokens) (bg, fg ui.Color) {
	switch s {
	case Danger:
		return k.DangerBg, k.Danger
	case Warning:
		return k.WarningBg, k.Warning
	case Success:
		return k.SuccessBg, k.Success
	case Accent:
		return k.AccentBg, k.AccentText
	}
	return k.Surface, k.Text
}
