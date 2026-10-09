package core

import (
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/theme"
)

func TestUseGivesTokensAndDensity(t *testing.T) {
	ui.NewTester(func(c *ui.Context) {
		Use(c, Settings{})
		if Tokens(c) != theme.Light() {
			t.Error("Tokens did not come from the light palette")
		}
		if Density(c) != theme.Compact {
			t.Errorf("density = %v, want Compact", Density(c))
		}
		if Density(c).Unit() <= 0 {
			t.Error("a density must give a usable unit")
		}
		if IsDark(c) {
			t.Error("a light window should not report dark")
		}
	}, 400, 200)
}

func TestDarkModeSwapsThePalette(t *testing.T) {
	ui.NewTester(func(c *ui.Context) {
		Use(c, Settings{Mode: Dark})
		if Tokens(c) != theme.Dark() {
			t.Error("Tokens did not come from the dark palette")
		}
		if !IsDark(c) {
			t.Error("IsDark should agree with the mode")
		}
	}, 400, 200)
}

func TestCallerCanOverrideAPalette(t *testing.T) {
	brand := theme.Light()
	brand.Accent = ui.Hex("#ff0055")
	ui.NewTester(func(c *ui.Context) {
		Use(c, Settings{Light: &brand})
		if Tokens(c).Accent != ui.Hex("#ff0055") {
			t.Error("an overridden palette was dropped")
		}
	}, 400, 200)
}

func TestComfortableDensifies(t *testing.T) {
	ui.NewTester(func(c *ui.Context) {
		Use(c, Settings{Density: theme.Comfortable})
		if Density(c).Unit() <= theme.Compact.Unit() {
			t.Errorf("comfortable unit %v is not larger than compact %v",
				Density(c).Unit(), theme.Compact.Unit())
		}
	}, 400, 200)
}

func TestTokensBeforeUsePanic(t *testing.T) {
	for name, get := range map[string]func(c *ui.Context){
		"Tokens":  func(c *ui.Context) { _ = Tokens(c) },
		"Density": func(c *ui.Context) { _ = Density(c) },
		"IsDark":  func(c *ui.Context) { _ = IsDark(c) },
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s before core.Use should panic: a window that "+
						"forgot Use would otherwise draw in the zero palette", name)
				}
			}()
			ui.NewTester(get, 200, 100)
		}()
	}
}

func TestUseGivesMyGoAUsableFontSize(t *testing.T) {
	var got float32
	ui.NewTester(func(c *ui.Context) {
		Use(c, Settings{})
		got = c.Theme().FontSize
	}, 200, 100)
	// MyGo lays plain ui.Text out from this, so a zero here is not a
	// cosmetic slip: every widget and every run of library text falls back to
	// MyGo's own default size and the two stop agreeing.
	if got != theme.BodySize {
		t.Errorf("c.Theme().FontSize = %v, want the library's body size %v", got, theme.BodySize)
	}
}

func TestUseKeepsTheFontFamilyAlreadyInEffect(t *testing.T) {
	const family = "Inter"
	var got string
	ui.NewTester(func(c *ui.Context) {
		c.SetTheme(&ui.Theme{Font: family})
		Use(c, Settings{})
		got = c.Theme().Font
	}, 200, 100)
	// Use re-themes the window rather than building one from nothing, so the
	// family the window already had has to survive it.
	if got != family {
		t.Errorf("c.Theme().Font = %q, want the window's own %q", got, family)
	}
}

func TestUseIsIdempotentWithinAFrame(t *testing.T) {
	ui.NewTester(func(c *ui.Context) {
		Use(c, Settings{})
		k := Tokens(c)
		Use(c, Settings{})
		if Tokens(c) != k {
			t.Error("a second Use in the same frame changed the palette")
		}
	}, 200, 100)
}
