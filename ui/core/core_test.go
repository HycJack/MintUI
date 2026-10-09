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

func TestReducedAndFontSizeFollowTheDesktop(t *testing.T) {
	var reduced bool
	var scaled float32
	tt := ui.NewTester(func(c *ui.Context) {
		Use(c, Settings{})
		reduced = Reduced(c)
		scaled = FontSize(c, 20)
	}, 200, 100)
	tt.SetPreferences(ui.Preferences{ReduceMotion: true, TextScale: 1.5})
	tt.Frame()
	if !reduced {
		t.Error("ReduceMotion was set but Reduced() reported false")
	}
	if scaled != 30 {
		t.Errorf("TextScale 1.5 on a 20pt size gave %v, want 30", scaled)
	}
}

func TestTheOverridesBeatTheDesktop(t *testing.T) {
	var reduced bool
	var scaled float32
	tt := ui.NewTester(func(c *ui.Context) {
		Use(c, Settings{})
		// Read first: a reader that has already run must not lock the window
		// out of the answer it is about to give itself.
		Reduced(c)
		FontSize(c, 20)
		c = WithReducedMotion(c, false)
		c = WithTextScale(c, 2)
		reduced = Reduced(c)
		scaled = FontSize(c, 20)
	}, 200, 100)
	tt.SetPreferences(ui.Preferences{ReduceMotion: true, TextScale: 1.5})
	tt.Frame()
	// A test and a window with no desktop behind it both settle these two
	// themselves, so the explicit setting has to win even when it is the one
	// the desktop would not have chosen.
	if reduced {
		t.Error("WithReducedMotion(false) did not beat the desktop's ReduceMotion")
	}
	if scaled != 40 {
		t.Errorf("WithTextScale(2) on a 20pt size gave %v, want 40", scaled)
	}
}

func TestWithMessagesReplacesTheWindowsCopy(t *testing.T) {
	// A window that swaps language passes a different map every frame, and the
	// second one has to take over rather than lose to the first for arriving
	// second.
	translations := map[string]string{"greeting": "Hello"}
	var got string
	tt := ui.NewTester(func(c *ui.Context) {
		Use(c, Settings{})
		c = WithMessages(c, translations)
		got = Messages(c)["greeting"]
	}, 200, 100)
	translations = map[string]string{"greeting": "你好"}
	tt.Frame()
	if got != "你好" {
		t.Errorf("Messages = %q after the window changed language, want %q", got, "你好")
	}
}

type counterKey struct{}

func TestSetLocalReplacesAStoredValue(t *testing.T) {
	var got int
	ui.NewTester(func(c *ui.Context) {
		Use(c, Settings{})
		// Claim the slot first: a reader that has already run is the ordinary
		// case, and the setter still has to answer.
		Local(c, counterKey{}, func() int { return 0 })
		SetLocal(c, counterKey{}, 7)
		got = Local(c, counterKey{}, func() int { return 0 })
	}, 200, 100)
	if got != 7 {
		t.Errorf("Local after SetLocal = %d, want 7", got)
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
