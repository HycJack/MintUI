package account

import (
	"math"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/theme"
)

// ThemeSelectorOptions configure a ThemeSelector.
type ThemeSelectorOptions struct {
	// Mode is the caller's choice, written as one of the three values
	// below. It is a string rather than a core.Mode because it goes to a
	// settings file and to a deep link, and "system" is what a settings file
	// should say rather than the zero value of an int.
	Mode *string
	// Modes are the choices offered, in order. Empty offers all three, which
	// is the only set that makes sense: a selector that offered Light and
	// Dark with no System would strand somebody whose desktop is the answer.
	Modes []string
	// Accent shows the accent picker under the modes, for a window that
	// offers one. It takes the pointer because a theme panel with a colour
	// control is a different question from whether a window wants one, and
	// the accent is not always about the theme: some windows leave the
	// accent alone entirely.
	Accent *ThemeAccentOptions
}

// ThemeAccentOptions configure the accent row of a ThemeSelector.
type ThemeAccentOptions struct {
	// Value is the caller's accent colour, written as it changes.
	Value *ui.Color
	// Presets are the swatches offered, which must be named: a swatch shows
	// a colour and nothing else, so an unnamed one is read out as nothing.
	Presets []input.Swatch
	// Label names the row; empty is the library's own word for it.
	Label string
}

// ThemeSelectorResult carries a ThemeSelector and the accent control in it.
type ThemeSelectorResult struct {
	// Element is the whole selector.
	Element *ui.Element
}

// ThemeSelector is which appearance the window uses, and optionally which
// accent it is drawn in.
//
// The three modes are a radio group rather than a segmented control because
// they are words, not marks: "System" and "Light" are not two shapes a person
// can tell apart at a glance, and a three-segment strip whose third segment
// is a word is a strip of buttons pretending to be a switch. input.RadioGroup
// already draws the label above each radio and puts them in a column when
// there is room, which is what three sentences want.
//
// The accent is a palette rather than a free picker. A full colour well is
// the right control for a drawing program and the wrong one for a settings
// pane: almost every window wants one of about eight accents, and the eight
// are the same eight everywhere.
func ThemeSelector(c *ui.Context, opts ThemeSelectorOptions) ThemeSelectorResult {
	if opts.Mode == nil {
		panic("account: ThemeSelector needs the *string Mode writes to; the " +
			"selector keeps no appearance of its own")
	}
	modes := opts.Modes
	if len(modes) == 0 {
		modes = defaultModes
	}
	choices := make([]input.Choice, 0, len(modes))
	for _, m := range modes {
		choices = append(choices, input.Choice{Value: m, Label: modeLabel(c, m)})
	}
	// A mode that is not one of the three cannot be drawn as a control that
	// is true, so it is refused rather than drawn unselected — which would
	// look like a setting nobody had touched.
	if *opts.Mode != "" && !hasMode(modes, *opts.Mode) {
		panic("account: ThemeSelector mode " + *opts.Mode + " is not one of its choices")
	}

	var r ThemeSelectorResult
	r.Element = ui.Column(c).FillWidth().Gap(unitOf(c) * 5).Children(func() {
		input.RadioGroup(c, opts.Mode, choices, input.RadioGroupOptions{
			Label: core.Msg(c, "account.appearance", "Appearance"),
		})
		if opts.Accent != nil {
			accentRow(c, opts.Accent)
		}
	})
	return r
}

// defaultModes is System, Light, Dark — in that order, because the one most
// people want is the one that follows the desktop and it is the one whose
// absence has to be noticed.
var defaultModes = []string{"system", "light", "dark"}

// modeLabel is what a mode is called as a choice, which is the word a person
// would use rather than the word stored in a file.
func modeLabel(c *ui.Context, mode string) string {
	switch mode {
	case "light":
		return core.Msg(c, "account.themeLight", "Light")
	case "dark":
		return core.Msg(c, "account.themeDark", "Dark")
	case "system":
		return core.Msg(c, "account.themeSystem", "Match system")
	}
	return mode
}

// hasMode reports whether the choices include a mode.
func hasMode(modes []string, want string) bool {
	for _, m := range modes {
		if m == want {
			return true
		}
	}
	return false
}

// accentRow is the swatches a window's accent can be chosen from.
func accentRow(c *ui.Context, opts *ThemeAccentOptions) {
	if opts.Value == nil {
		panic("account: a theme accent needs the *ui.Color it writes to")
	}
	label := opts.Label
	if label == "" {
		label = core.Msg(c, "account.accent", "Accent colour")
	}
	input.ColorPalette(c, opts.Value, opts.Presets, input.ColorPaletteOptions{
		Label: label,
	})
}

// AccentColorPickerOptions configure an AccentColorPicker.
type AccentColorPickerOptions struct {
	// Value is the caller's colour, written as the well moves or a swatch
	// is pressed.
	Value *ui.Color
	// Presets are the named swatches above the well.
	Presets []input.Swatch
	// Label names the control; it is required, because the well shows a
	// colour and no words at all.
	Label string
	// Contrast is the tone the picker names its choices on. It is an argument
	// rather than a decision so that a window which has overridden its
	// palette is measured against the palette it actually has — the swatches
	// in a custom dark theme are not legible against the library's.
	Contrast *theme.Tokens
}

// AccentColorPickerResult carries an AccentColorPicker.
type AccentColorPickerResult struct {
	// Element is the swatches and the well.
	Element *ui.Element
}

// AccentColorPicker is the colour a window is accented with: a set of named
// swatches and a well for anything else.
//
// The presets come first. A picker with only a well makes somebody who wants
// the ordinary blue hunt through a colour space for it, and the reason most
// windows pick an accent at all is that the default was not it. The well is
// underneath for the window whose brand colour is not in anybody's palette.
//
// Contrast is a pointer so that a window which has replaced its tokens passes
// them in; every swatch's label and border are then computed against what is
// actually behind them rather than against the library's light palette.
func AccentColorPicker(c *ui.Context, opts AccentColorPickerOptions) AccentColorPickerResult {
	if opts.Value == nil {
		panic("account: AccentColorPicker needs the *ui.Color it writes to")
	}
	if opts.Label == "" {
		panic("account: AccentColorPicker needs a Label; the well shows a " +
			"colour and no words, so without one there is nothing to read out")
	}
	u := unitOf(c)

	// The preview is drawn against the palette the caller will actually be
	// accented with, which is not always the library's own: a window that has
	// replaced its tokens would otherwise be shown a swatch set that has
	// already been picked for a different background.
	k := core.Tokens(c)
	if opts.Contrast != nil {
		k = *opts.Contrast
	}

	var r AccentColorPickerResult
	r.Element = ui.Column(c).FillWidth().Gap(u * 2).Children(func() {
		if len(opts.Presets) > 0 {
			input.ColorPalette(c, opts.Value, opts.Presets, input.ColorPaletteOptions{
				Label: opts.Label,
			})
		}
		input.ColorPicker(c, opts.Value, input.ColorPickerOptions{Label: opts.Label})

		// A filled pill, an outlined one and a tinted tag: the three places an
		// accent ends up in the rest of the app. A colour well tells somebody
		// what they picked; this tells them whether they picked a usable one,
		// which is the question they actually have.
		ui.Row(c).FillWidth().Gap(u).AlignItems(ui.Center).Children(func() {
			ui.Box(c).Padding(u*0.75, u*2).Radius(theme.PillRadius).
				Background(*opts.Value).Children(func() {
				ui.Text(c, core.Msg(c, "account.accentFilled", "Filled")).
					TextColor(onFillOf(k, *opts.Value)).
					FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
			})
			ui.Box(c).Padding(u*0.75, u*2).Radius(theme.PillRadius).
				Background(k.Surface).BorderWidth(theme.BorderWidth * 2).
				BorderColor(*opts.Value).Children(func() {
				ui.Text(c, core.Msg(c, "account.accentOutline", "Outline")).
					TextColor(k.Text).
					FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
			})
			display.Tag(c, core.Msg(c, "account.accentTag", "Tag"),
				display.TagOptions{Tone: core.Accent})
		})
	})
	return r
}

// onFillOf is the ink that goes on top of an accent.
//
// It is the palette's own OnFill whenever the accent is one the palette
// already pairs a foreground with — a caller's custom accent usually is not —
// and white or near-black chosen by the accent's own lightness otherwise. A
// hard-coded foreground on an arbitrary colour is how an accent picker ends
// up offering a swatch nobody can read the label on.
func onFillOf(k theme.Tokens, accent ui.Color) ui.Color {
	if accent == k.Accent {
		return k.AccentText
	}
	return contrastInk(accent)
}

// contrastInk is white or near-black, whichever reads better on a colour. The
// threshold is the usual one: a colour is "light" when its luminance is above
// half, which is where the two inks cross over.
func contrastInk(col ui.Color) ui.Color {
	if relativeLuminance(col) > 0.5 {
		return ui.Hex("#18181b")
	}
	return ui.Hex("#ffffff")
}

// relativeLuminance is the WCAG definition: the channels are linearised
// before they are combined, which is what makes the midpoint a midpoint. A
// plain weighted average of the bytes puts the crossover for a mid grey at
// the wrong place, and the error is worst exactly where it matters — an
// accent of medium lightness is the common case.
func relativeLuminance(col ui.Color) float64 {
	lin := func(v uint8) float64 {
		x := float64(v) / 255
		if x <= 0.03928 {
			return x / 12.92
		}
		return math.Pow((x+0.055)/1.055, 2.4)
	}
	return 0.2126*lin(col.R) + 0.7152*lin(col.G) + 0.0722*lin(col.B)
}

// unitOf is the density's spacing step. It is one function so that the
// handful of sites in this package that only need a gap do not each say the
// whole expression.
func unitOf(c *ui.Context) float32 { return core.Density(c).Unit() }
