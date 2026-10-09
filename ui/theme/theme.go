// Package theme holds the colours, spacing, radii and type sizes every
// component draws with.
//
// A Tokens set is one appearance. Light and Dark are the two the library
// ships; a caller that needs its own builds one and passes it to
// core.Settings.
//
// Tokens do not change over an app's life: a window that follows the system
// resolves a fresh set each frame, but the values themselves are constants.
package theme

import (
	"math"

	"github.com/egoist/mygo/ui"
)

// Tokens is one appearance's palette.
//
// The role names follow MyGo's own semantics, which read backwards from
// intuition and are the easiest thing to get wrong:
//
//	Background is the window behind everything, white in a light theme
//	Surface    is a panel raised on it, light grey in a light theme
//
// So a card is Surface-on-Background. Inverting the two leaves cards with
// nothing to lift them off their column and the whole board reads flat.
type Tokens struct {
	// Background is the window: the rail, the main area, cards and sheets.
	Background ui.Color
	// Surface is what sits on it: sidebar, board columns, icon tiles, chips.
	Surface ui.Color
	// SurfaceHover is Surface under the pointer.
	SurfaceHover ui.Color
	// SurfacePressed is Surface while a press is held.
	SurfacePressed ui.Color
	// Border separates and outlines.
	Border ui.Color

	// Text is the primary foreground, TextMuted the secondary one.
	Text      ui.Color
	TextMuted ui.Color
	TextFaint ui.Color

	// Fill is a saturated-by-ink surface: the active rail item, primary
	// buttons, toasts. OnFill is the text that goes on it.
	Fill   ui.Color
	OnFill ui.Color

	// Accent is the system's highlight, used only where a control needs it.
	Accent ui.Color

	// Severity ramps, each a background/foreground pair.
	DangerBg   ui.Color
	Danger     ui.Color
	WarningBg  ui.Color
	Warning    ui.Color
	SuccessBg  ui.Color
	Success    ui.Color
	AccentBg   ui.Color
	AccentText ui.Color

	// Lively is the one bright colour in the interface: the presence dot and
	// the empty state's action. It is the same in both appearances.
	Lively ui.Color
}

// IsDark reports whether k is one of the dark palettes, by how much light its
// background throws back. Tokens carry no mode of their own so that a caller
// can mix a dark background with a custom fill.
func (k Tokens) IsDark() bool {
	return relativeLuminance(k.Background) < 0.5
}

// relativeLuminance is how bright a colour looks, 0 for black and 1 for white.
// The weights and the curve are WCAG's for sRGB rather than a plain average,
// so IsDark and the contrast assertions in the tests are asking about the same
// light.
//
// Half is the midpoint of that light, not of the byte range. WCAG's curve puts
// the weighting where the eye is, and the eye resolves shadows far more finely
// than highlights, so sRGB #808080 — half way between black and white by value
// — is 0.216 here and reads dark; the crossing is nearer #bcbcbc. That follows
// from the scale rather than being a slip in it, and it is harmless where it
// bites: at #808080 either glyph colour is legible (3.59:1 for the dark theme's
// #f4f4f5, 4.49:1 for the light theme's #18181b), so calling such a window dark
// cannot make it unreadable. Both shipped palettes are far outside the band
// where the question arises — #ffffff is 1.000, #111113 is 0.006 — and so is
// any background a caller is likely to invent.
func relativeLuminance(c ui.Color) float64 {
	f := func(v uint8) float64 {
		x := float64(v) / 255
		if x <= 0.03928 {
			return x / 12.92
		}
		return math.Pow((x+0.055)/1.055, 2.4)
	}
	return 0.2126*f(c.R) + 0.7152*f(c.G) + 0.0722*f(c.B)
}

// Light is the library's light palette.
func Light() Tokens {
	return Tokens{
		Background:     ui.Hex("#ffffff"),
		Surface:        ui.Hex("#f4f4f5"),
		SurfaceHover:   ui.Hex("#ececee"),
		SurfacePressed: ui.Hex("#e4e4e7"),
		Border:         ui.Hex("#d9d9de"),
		Text:           ui.Hex("#18181b"),
		TextMuted:      ui.Hex("#64646d"),
		TextFaint:      ui.Hex("#7f7f89"),
		Fill:           ui.Hex("#1a1a1f"),
		OnFill:         ui.Hex("#ffffff"),
		Accent:         ui.Hex("#2563eb"),
		DangerBg:       ui.Hex("#fdecec"),
		Danger:         ui.Hex("#c62b30"),
		WarningBg:      ui.Hex("#fdf1e6"),
		Warning:        ui.Hex("#9a5410"),
		SuccessBg:      ui.Hex("#e8f5e9"),
		Success:        ui.Hex("#1f7a3d"),
		AccentBg:       ui.Hex("#e8effd"),
		AccentText:     ui.Hex("#1d4ed8"),
		Lively:         ui.Hex("#d9f24b"),
	}
}

// Dark is the library's dark palette: the same roles, inverted.
func Dark() Tokens {
	return Tokens{
		Background:     ui.Hex("#111113"),
		Surface:        ui.Hex("#1c1c1f"),
		SurfaceHover:   ui.Hex("#26262a"),
		SurfacePressed: ui.Hex("#2f2f34"),
		Border:         ui.Hex("#33333a"),
		Text:           ui.Hex("#f4f4f5"),
		TextMuted:      ui.Hex("#a1a1aa"),
		TextFaint:      ui.Hex("#7a7a86"),
		Fill:           ui.Hex("#f4f4f5"),
		OnFill:         ui.Hex("#18181b"),
		Accent:         ui.Hex("#5b8dff"),
		DangerBg:       ui.Hex("#3a1f21"),
		Danger:         ui.Hex("#ff6b6f"),
		WarningBg:      ui.Hex("#3a2a17"),
		Warning:        ui.Hex("#e5a04a"),
		SuccessBg:      ui.Hex("#16301d"),
		Success:        ui.Hex("#5fc47c"),
		AccentBg:       ui.Hex("#16233d"),
		AccentText:     ui.Hex("#9dc0ff"),
		Lively:         ui.Hex("#d9f24b"),
	}
}

// Density is how much room a component takes, for people who want more or
// less of it. It scales spacing, never type: text stays legible.
type Density int

const (
	// Compact packs controls tightly; it is the zero value, so a Settings
	// that says nothing about density gets it.
	Compact Density = iota
	// Comfortable adds a step to every gap.
	Comfortable
)

// Unit returns one step of the density's spacing scale, in DIPs.
func (d Density) Unit() float32 {
	switch d {
	case Comfortable:
		return 6
	default:
		return 4
	}
}

func (d Density) String() string {
	switch d {
	case Comfortable:
		return "Comfortable"
	}
	return "Compact"
}

// ── radii ──────────────────────────────────────────────────────────────────
// One family, so curves read as related. Pill is a shape, not a size.

const (
	SmallRadius   float32 = 8
	ControlRadius float32 = 14
	CardRadius    float32 = 18
	PanelRadius   float32 = 24
	PillRadius    float32 = 999
)

// ── type ───────────────────────────────────────────────────────────────────
// Sizes are in points and are deliberately half-stepped: a secondary label
// holds together at 13.5 where 13 turns spindly, and a card title wraps later
// at 15.5 than it would at 15.

const (
	DisplaySize float32 = 40   // a page's own heading
	TitleSize   float32 = 24   // a panel's title
	SheetSize   float32 = 21   // a modal's title
	LeadSize    float32 = 17   // an empty state's title
	BodySize    float32 = 15.5 // a card or row's subject
	StatSize    float32 = 14.5 // a figure that matters
	RowSize     float32 = 13.5 // a list row, a filter row
	MetaSize    float32 = 13   // a row's secondary line
	CaptionSize float32 = 12   // a chip, a count, a badge
	MonoSize    float32 = 11.5 // an avatar's initials
	GlyphSize   float32 = 11   // a drawn meter
	IconSize    float32 = 21   // a navigation or button icon
)

// BorderWidth is the outline weight. One value, so borders match everywhere.
const BorderWidth float32 = 1

// Shell metrics: the two widths a window's frame is built from. They are
// fixed numbers rather than fractions for the same reason ColumnWidth is —
// a sidebar that narrows with the window squeezes its own labels until they
// truncate, and the labels are what the sidebar is for.
const (
	// SidebarWidth is a navigation sidebar's width when expanded.
	SidebarWidth float32 = 322
	// RailWidth is the same sidebar collapsed to icons.
	RailWidth float32 = 76
)

// ColumnWidth is the narrowest a KanbanColumn may be before its card titles
// start wrapping. Columns scroll rather than shrink below it.
const ColumnWidth float32 = 272
