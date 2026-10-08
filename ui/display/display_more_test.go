package display

import (
	"image"
	"image/color"
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// The tests here assert what a person would get: the words and names on the
// screen, the boxes they occupy, and the value a component reported back into
// the caller's own variables. Nothing reaches into an element's fields, and a
// press is asserted on the number it was reported as rather than on the
// picture it settled into — a click settles on the next frame, and a test
// that waits for the picture is testing the picture.

// findNamed fails the test when the window shows nothing by that name.
func findNamed(t *testing.T, tt *ui.Tester, name string) ui.Rect {
	t.Helper()
	r, ok := tt.Find(name)
	if !ok {
		t.Fatalf("nothing shows %q; shown: %v", name, tt.Texts())
	}
	return r
}

func bitmapOf(w, h int) *ui.Bitmap {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	blue := color.RGBA{R: 90, G: 140, B: 220, A: 255}
	for y := range h {
		for x := range w {
			img.Set(x, y, blue)
		}
	}
	return ui.NewBitmap(img)
}

// ── Icon ────────────────────────────────────────────────────────────────────

// A glyph has no word of its own, so its name is the only thing a screen
// reader has. A test looks it up by the same name.
func TestIconIsNamedAfterWhatItDraws(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Icon(c, IconCallbacks, IconOptions{Name: "Callbacks"})
	}, 200, 120)
	findNamed(t, tt, "Callbacks")
}

func TestIconTakesItsSizeFromTheOption(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Icon(c, IconBell, IconOptions{Name: "Alerts", Size: 28})
	}, 200, 120)
	got := findNamed(t, tt, "Alerts")
	if got.W != 28 || got.H != 28 {
		t.Errorf("icon is %v, want a 28×28 box", got)
	}
}

func TestIconDefaultsToTheNavigationSize(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Icon(c, IconBell, IconOptions{Name: "Alerts"})
	}, 200, 120)
	got := findNamed(t, tt, "Alerts")
	if got.W != theme.IconSize || got.H != theme.IconSize {
		t.Errorf("icon is %v, want the shared %v", got, theme.IconSize)
	}
}

func TestIconNeedsAName(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("an unnamed Icon should panic: a glyph is the one element with no word of its own")
		}
	}()
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Icon(c, IconBell, IconOptions{})
	}, 200, 120)
}

func TestIconRejectsANameItDoesNotHave(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("an unknown icon name should panic rather than draw a blank square")
		}
	}()
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Icon(c, IconName("trolley"), IconOptions{Name: "Trolley"})
	}, 200, 120)
}

// Every built-in name has to parse: one that does not is a blank square that
// reads as a design choice rather than as a bug.
func TestEveryBuiltInIconDraws(t *testing.T) {
	names := make([]IconName, 0, len(iconShapes))
	for n := range iconShapes {
		names = append(names, n)
	}
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ui.Row(c).Wrap().Children(func() {
			for _, n := range names {
				Icon(c, n, IconOptions{Name: string(n), Size: 16})
			}
		})
	}, 600, 400)
}

// ── Badge ───────────────────────────────────────────────────────────────────

func TestBadgeCountsAndIsNamed(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ui.Box(c).Children(func() {
			Avatar(c, "Rosa Delgado")
			Badge(c, 3, BadgeOptions{Name: "3 unread"}).Attach(ui.AnchorTopRight, ui.AnchorCenter)
		})
	}, 240, 160)
	if !tt.HasText("3") {
		t.Errorf("badge did not draw its count: %v", tt.Texts())
	}
	findNamed(t, tt, "3 unread")
}

func TestBadgeSaysNothingWhenThereIsNothingToCount(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Badge(c, 0, BadgeOptions{Name: "unread"})
	}, 240, 160)
	if tt.HasText("0") {
		t.Errorf("a count of zero should draw no badge at all, drew %v", tt.Texts())
	}
}

func TestBadgeCapsItsNumber(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Badge(c, 1284, BadgeOptions{Name: "unread", Max: 99})
	}, 240, 160)
	if !tt.HasText("99+") {
		t.Errorf("a four-digit badge should read as 99+, drew %v", tt.Texts())
	}
	if tt.HasText("1,284") {
		t.Error("the badge printed the whole count; it is wider than the avatar it is stuck to")
	}
}

func TestBadgeNeedsAName(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("an unnamed Badge should panic: a count with no word is a smudge")
		}
	}()
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Badge(c, 3, BadgeOptions{})
	}, 240, 160)
}

func TestPresenceDotIsNamed(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ui.Box(c).Children(func() {
			Avatar(c, "Dana Reyes")
			PresenceDot(c, BadgeOptions{Name: "2 alerts"}).Attach(ui.AnchorTopRight, ui.AnchorCenter)
		})
	}, 240, 160)
	// A dot has no text of its own, so the name is the whole content: the
	// test finds it by the only thing a reader would hear.
	findNamed(t, tt, "2 alerts")
}

func TestPresenceDotNeedsAName(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("an unnamed PresenceDot should panic")
		}
	}()
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		PresenceDot(c, BadgeOptions{})
	}, 240, 160)
}

// ── AvatarGroup ─────────────────────────────────────────────────────────────

var crew = []string{
	"Rosa Delgado", "Dana Reyes", "Omar Haddad", "Ingrid Sol",
	"Priya Raman", "Tomas Novak", "Bea Lindqvist",
}

func TestAvatarGroupDrawsFacesAndTheRest(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		AvatarGroup(c, crew, AvatarGroupOptions{Max: 4, More: "See the other 3"})
	}, 400, 160)
	// Four faces and then the number that stands for the other three.
	for _, n := range crew[:4] {
		if !tt.HasText(n) {
			t.Errorf("the group is missing %q", n)
		}
	}
	if tt.HasText(crew[4]) {
		t.Error("the group drew a face it was told to cap")
	}
	if !tt.HasText("+3") {
		t.Errorf("the group did not count the rest: %v", tt.Texts())
	}
}

func TestAvatarGroupReportsHowManyAreBehind(t *testing.T) {
	// Asserted on the number the component carries, not on the "+3" drawn
	// beside it: the count is a claim, and the glyph is a consequence of it.
	var overflow int
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		overflow = AvatarGroup(c, crew, AvatarGroupOptions{Max: 4, More: "See the other 3"}).Overflow()
	}, 400, 160)
	if overflow != 3 {
		t.Errorf("Overflow() = %d, want 3 — four faces of seven leave three behind", overflow)
	}
}

func TestAvatarGroupFitsWithoutAnOverflow(t *testing.T) {
	var overflow int
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		overflow = AvatarGroup(c, crew[:3], AvatarGroupOptions{Max: 4}).Overflow()
	}, 400, 160)
	if overflow != 0 {
		t.Errorf("a group that fits reported %d behind a count it never drew", overflow)
	}
	if tt.HasText("+") {
		t.Errorf("a group that fits should draw no count: %v", tt.Texts())
	}
}

func TestAvatarGroupReportsThePressOnTheCount(t *testing.T) {
	expanded := false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		// The state the caller keeps. Asserting it and not the picture is
		// the honest check: a press settles on the next frame.
		if AvatarGroup(c, crew, AvatarGroupOptions{Max: 4, More: "See the other 3"}).Expanded() {
			expanded = true
		}
	}, 400, 160)
	if expanded {
		t.Fatal("the group was open before anything was pressed")
	}
	if err := tt.Click("+3"); err != nil {
		t.Fatal(err)
	}
	if !expanded {
		t.Error("pressing the count should report that the group was expanded")
	}
}

func TestAvatarGroupNeedsANameForTheFacesBehindTheCount(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("a group with a hidden rest and no More should panic: a bare +3 is a puzzle")
		}
	}()
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		AvatarGroup(c, crew, AvatarGroupOptions{Max: 4})
	}, 400, 160)
}

// A group of two that fits has no overflow control, so it is named after what
// it is rather than after a control it does not have.
func TestAvatarGroupThatFitsIsNamedForItsFaces(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		AvatarGroup(c, crew[:2], AvatarGroupOptions{Max: 4})
	}, 400, 160)
	if _, ok := tt.Find("2 people on this callback"); !ok {
		t.Errorf("a group that fits is not named for the people in it: %v", tt.Texts())
	}
}

// ── Image ───────────────────────────────────────────────────────────────────

func TestImageShowsTheBitmap(t *testing.T) {
	shot := bitmapOf(8, 8)
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		// The column is only here to give the box a top-left of its own: an
		// element dropped straight into the root of a window has no edge to
		// measure, and a test that measured nothing would pass for free.
		ui.Column(c).Padding(16).Children(func() {
			Image(c, shot, ImageOptions{
				Width: 64, Height: 48, Name: "Riverside Clinic", Cover: true,
				Radius: theme.SmallRadius,
			})
		})
	}, 300, 200)
	got := findNamed(t, tt, "Riverside Clinic")
	if got.W < 60 || got.H < 44 {
		t.Errorf("the picture is %v, want about a 64×48 box", got)
	}
	if tt.HasText("Photo") {
		t.Error("a loaded picture should not still be showing its placeholder")
	}
}

func TestImagePlaceholderIsNotAFailure(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ui.Row(c).Gap(12).Children(func() {
			Image(c, nil, ImageOptions{Width: 80, Height: 80, Name: "Rosa Delgado", Placeholder: "Photo"})
			Image(c, nil, ImageOptions{Width: 80, Height: 80, Name: "Dana Reyes", Failed: "Unavailable"})
		})
	}, 400, 200)
	if !tt.HasText("Photo") {
		t.Errorf("the placeholder is missing: %v", tt.Texts())
	}
	if !tt.HasText("Unavailable") {
		t.Errorf("the failure state is missing: %v", tt.Texts())
	}
	// Both boxes are still found by name: a slot that has failed is still
	// the place a person's name is, and a test that could not find it could
	// not check the failure either.
	findNamed(t, tt, "Rosa Delgado")
	findNamed(t, tt, "Dana Reyes")
}

func TestImageKeepsItsShapeFromTheRatio(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Image(c, nil, ImageOptions{Width: 120, Ratio: 1.5, Name: "Job photo"})
	}, 400, 200)
	got := findNamed(t, tt, "Job photo")
	if got.W < 118 || got.W > 122 {
		t.Errorf("the box is %v wide, want the 120 it was given", got)
	}
	if got.H < 78 || got.H > 82 {
		t.Errorf("the box is %v high, want 120/1.5 = 80", got)
	}
}

func TestImageNeedsAName(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("an unnamed Image should panic: a picture has no words of its own to be read")
		}
	}()
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Image(c, nil, ImageOptions{Width: 40, Height: 40})
	}, 200, 160)
}

// ── the dark palette ────────────────────────────────────────────────────────

// Both appearances draw every one of these, and draw them the same way: a
// component that only works against a white window is a component the other
// window does not have.
func TestTheNewMarksSurviveTheDarkPalette(t *testing.T) {
	for _, mode := range []core.Mode{core.Light, core.Dark} {
		names := []string{"Callbacks", "3 unread", "2 alerts"}
		tt := ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{Mode: mode})
			ui.Row(c).Gap(u4()).Padding(16).Children(func() {
				ui.Column(c).Gap(8).Children(func() {
					Icon(c, IconCallbacks, IconOptions{Name: "Callbacks"})
					Badge(c, 3, BadgeOptions{Name: "3 unread", Tone: core.Danger})
					PresenceDot(c, BadgeOptions{Name: "2 alerts"})
					AvatarGroup(c, crew, AvatarGroupOptions{Max: 4, More: "See the other 3"})
				})
				Image(c, nil, ImageOptions{Width: 64, Height: 64, Name: "Riverside Clinic", Placeholder: "Photo"})
				Image(c, nil, ImageOptions{Width: 64, Height: 64, Name: "Dana Reyes", Failed: "Unavailable"})
			})
		}, 700, 320)
		for _, want := range append(names, "Photo", "Unavailable", "+3") {
			if !tt.HasText(want) {
				t.Errorf("%v mode lost %q; shown: %v", mode, want, tt.Texts())
			}
		}
		for _, want := range names {
			if _, ok := tt.Find(want); !ok {
				t.Errorf("%v mode shows nothing named %q", mode, want)
			}
		}
	}
}

// u4 is the density unit, read from a setting the tests do not build: the
// numbers in the layout above are only there to keep the marks apart, and
// going through core for them would mean calling Use twice in one view.
func u4() float32 { return 4 }
