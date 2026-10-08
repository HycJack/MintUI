package theme

import (
	"math"
	"testing"

	"github.com/egoist/mygo/ui"
)

// contrast is the WCAG ratio between two opaque colours, so a test can say
// "this text is unreadable" instead of eyeballing a screenshot.
func contrast(a, b ui.Color) float64 {
	la, lb := lum(a), lum(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

func lum(c ui.Color) float64 {
	f := func(v uint8) float64 {
		x := float64(v) / 255
		if x <= 0.03928 {
			return x / 12.92
		}
		return math.Pow((x+0.055)/1.055, 2.4)
	}
	return 0.2126*f(c.R) + 0.7152*f(c.G) + 0.0722*f(c.B)
}

func TestPalettesDiffer(t *testing.T) {
	l, d := Light(), Dark()
	for name, pair := range map[string][2]ui.Color{
		"Surface":      {l.Surface, d.Surface},
		"SurfaceHover": {l.SurfaceHover, d.SurfaceHover},
		"Text":         {l.Text, d.Text},
		"Fill":         {l.Fill, d.Fill},
		"Accent":       {l.Accent, d.Accent},
	} {
		if pair[0] == pair[1] {
			t.Errorf("%s is the same in light and dark: %+v", name, pair[0])
		}
	}
	if l.IsDark() || !d.IsDark() {
		t.Error("IsDark does not report the palette it was asked about")
	}
}

func TestTextIsReadableOnEverySurface(t *testing.T) {
	for name, k := range map[string]Tokens{"light": Light(), "dark": Dark()} {
		for _, s := range []struct {
			what   string
			fg, bg ui.Color
		}{
			{"text on background", k.Text, k.Background},
			{"text on surface", k.Text, k.Surface},
			{"text on surface hover", k.Text, k.SurfaceHover},
			{"text on fill", k.OnFill, k.Fill},
			{"muted on surface", k.TextMuted, k.Surface},
		} {
			if got := contrast(s.fg, s.bg); got < 4.5 {
				t.Errorf("%s: %s is %.2f:1, below the 4.5 body-text floor", name, s.what, got)
			}
		}
		// Faint carries marks, not reading: the separator between a card's
		// reference and its subject. It still has to be visible.
		if got := contrast(k.TextFaint, k.Surface); got < 3 {
			t.Errorf("%s: faint marks are %.2f:1, below 3", name, got)
		}
	}
}

func TestSeverityTextIsReadableOnItsPill(t *testing.T) {
	for name, k := range map[string]Tokens{"light": Light(), "dark": Dark()} {
		for _, s := range []struct {
			what   string
			fg, bg ui.Color
		}{
			{"danger", k.Danger, k.DangerBg},
			{"warning", k.Warning, k.WarningBg},
			{"success", k.Success, k.SuccessBg},
			{"accent", k.AccentText, k.AccentBg},
		} {
			if got := contrast(s.fg, s.bg); got < 4.5 {
				t.Errorf("%s: %s pill text is %.2f:1", name, s.what, got)
			}
		}
	}
}

func TestBordersAreQuiet(t *testing.T) {
	for name, k := range map[string]Tokens{"light": Light(), "dark": Dark()} {
		if got := contrast(k.Border, k.Surface); got > 2 {
			t.Errorf("%s: a border at %.2f:1 draws a line, not an edge", name, got)
		}
	}
}

func TestHoverAndPressedAreDistinguishable(t *testing.T) {
	for name, k := range map[string]Tokens{"light": Light(), "dark": Dark()} {
		if contrast(k.Surface, k.SurfaceHover) < 1.03 {
			t.Errorf("%s: hover at %.3f:1 is invisible", name, contrast(k.Surface, k.SurfaceHover))
		}
		if contrast(k.Surface, k.SurfacePressed) < 1.06 {
			t.Errorf("%s: pressed at %.3f:1 is too close to rest", name, contrast(k.Surface, k.SurfacePressed))
		}
		if contrast(k.SurfaceHover, k.SurfacePressed) < 1.02 {
			t.Errorf("%s: pressed does not differ from hover", name)
		}
	}
}

func TestColumnWidthKeepsTitlesOnOneLine(t *testing.T) {
	if ColumnWidth < 240 || ColumnWidth > 320 {
		t.Errorf("ColumnWidth = %v: below 240 a card title wraps twice, "+
			"above 320 too few lanes fit a laptop screen", ColumnWidth)
	}
}

func TestDensityScalesSpaceNotType(t *testing.T) {
	if Compact.Unit() >= Comfortable.Unit() {
		t.Errorf("compact %v should be tighter than comfortable %v",
			Compact.Unit(), Comfortable.Unit())
	}
	if TitleSize <= RowSize || RowSize <= CaptionSize || CaptionSize <= 0 {
		t.Errorf("type scale is not ascending: %v > %v > %v",
			TitleSize, RowSize, CaptionSize)
	}
}
