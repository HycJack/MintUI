package feedback

import (
	"image"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// The tests here render each component without a window and say what was
// drawn: the texts a person would read, the boxes they occupy, and the pixels
// they paint. Nothing here asserts that something animated, because a clock
// cannot be set from a test — ui's headless frames run on the wall clock — and
// a test that waits for a spinner to reach a phase is a test that fails on a
// loaded machine. What is asserted instead is what is on screen: a resting
// pose, a filled fraction, a colour from the palette.

// ── helpers ────────────────────────────────────────────────────────────────

// render runs a view in a window of the given size with the library's light
// palette, and returns the tester. Every test starts this way so that no test
// depends on another having set the preference first.
func render(w, h int, view func(c *ui.Context)) *ui.Tester {
	return ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		// A painted background, so a pixel that is not ink is a known
		// colour rather than whatever the surface happened to be.
		ui.Box(c).Fill().Background(core.Tokens(c).Background).
			Children(func() { view(c) })
	}, w, h)
}

// renderStill is render for a window that asked for reduced motion. It is a
// separate helper because getting it wrong — reading core.Reduced before
// core.Use has run, or installing it a frame late — silently produces a test
// that thinks it covered reduced motion and did not.
func renderStill(w, h int, view func(c *ui.Context)) *ui.Tester {
	return ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		core.WithReducedMotion(c, true)
		ui.Box(c).Fill().Background(core.Tokens(c).Background).
			Children(func() { view(c) })
	}, w, h)
}

// renderDark is render for a window following the desktop into dark. The mode
// is core.System so that the tester's own SetDark is what decides, which is
// how a real window behaves and which is the only way to catch a component
// that hard-codes one appearance.
func renderDark(w, h int, view func(c *ui.Context)) *ui.Tester {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: core.System})
		ui.Box(c).Fill().Background(core.Tokens(c).Background).
			Children(func() { view(c) })
	}, w, h)
	tt.SetDark(true)
	return tt
}

// box returns the box of the element showing or labelled s, failing the test
// when there is none. An empty ui.Box measures zero wide, so anything that has
// to be measured has to be found by a name.
func box(t *testing.T, tt *ui.Tester, s string) ui.Rect {
	t.Helper()
	r, ok := tt.Find(s)
	if !ok {
		t.Fatalf("nothing showing or labelled %q; the frame has %q", s, tt.Texts())
	}
	return r
}

// at returns the colour of one pixel, in DIPs.
func at(tt *ui.Tester, x, y float32) ui.Color { return atPx(tt.Image(), x, y) }

// atPx is at for a frame already in hand. Tester.Image copies the whole
// pixel buffer on every call, so anything that scans takes it once.
func atPx(img *image.RGBA, x, y float32) ui.Color {
	px, py := int(x), int(y)
	if !(image.Point{X: px, Y: py}).In(img.Bounds()) {
		return ui.Color{}
	}
	i := img.PixOffset(px, py)
	return ui.Color{R: img.Pix[i], G: img.Pix[i+1], B: img.Pix[i+2], A: img.Pix[i+3]}
}

// sameColor compares two colours as they are painted, which is straight sRGB
// and straight alpha: a solid fill comes back exactly as it was set, so an
// exact comparison is the strongest assertion available and anything looser
// would hide a component drawing the wrong token.
func sameColor(got, want ui.Color) bool {
	return got.R == want.R && got.G == want.G && got.B == want.B && got.A == want.A
}

// wantColor fails unless the pixel at (x, y) is exactly want.
func wantColor(t *testing.T, tt *ui.Tester, x, y float32, want ui.Color, what string) {
	t.Helper()
	if got := at(tt, x, y); !sameColor(got, want) {
		t.Errorf("%s: pixel (%v, %v) is %v, want %v", what, x, y, got, want)
	}
}

// wantNotColor fails when the pixel at (x, y) is exactly c, which is how a
// test says "this is not ink here" rather than "this is not nothing here".
func wantNotColor(t *testing.T, tt *ui.Tester, x, y float32, c ui.Color, what string) {
	t.Helper()
	if got := at(tt, x, y); sameColor(got, c) {
		t.Errorf("%s: pixel (%v, %v) should not be %v", what, x, y, c)
	}
}

// wantTexts fails unless every one of want is on screen.
func wantTexts(t *testing.T, tt *ui.Tester, want ...string) {
	t.Helper()
	for _, w := range want {
		if !tt.HasText(w) {
			t.Errorf("missing text %q; the frame has %q", w, tt.Texts())
		}
	}
}

// wantOnly fails unless the frame's accessible names are exactly want, in
// order. Texts reports what elements are *named*, not the glyphs they draw,
// which is why a component that says one thing and draws another needs this
// rather than wantTexts.
func wantOnly(t *testing.T, tt *ui.Tester, want ...string) {
	t.Helper()
	got := tt.Texts()
	if len(got) != len(want) {
		t.Errorf("frame names %q, want exactly %q", got, want)
		return
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("frame names %q, want exactly %q", got, want)
			return
		}
	}
}

// inkOf returns the colour a translucent ink makes when it is laid over the
// window background, which is how a component that paints with an alpha is
// asserted on rather than merely checked for existing.
func inkOf(c ui.Color) ui.Color {
	return c.Over(theme.Light().Background)
}

// wantInk is wantColor for a translucent ink. A composite of two colours is
// rounded twice — once by ui.Color.Over and once by the renderer — so the two
// can differ by one in a channel, and an exact comparison would report that
// as a component drawing the wrong colour.
func wantInk(t *testing.T, tt *ui.Tester, x, y float32, want ui.Color, what string) {
	t.Helper()
	got := at(tt, x, y)
	d := func(a, b uint8) int {
		if a > b {
			return int(a - b)
		}
		return int(b - a)
	}
	if d(got.R, want.R) > 1 || d(got.G, want.G) > 1 || d(got.B, want.B) > 1 ||
		got.A != want.A {
		t.Errorf("%s: pixel (%v, %v) is %v, want %v (±1 per channel)", what, x, y, got, want)
	}
}

// wantsPanic fails unless building the view panics with a message mentioning
// want. Every component here panics rather than drawing nothing for an input
// it cannot honour, so each of those panics is part of the contract.
func wantsPanic(t *testing.T, want string, view func(c *ui.Context)) {
	t.Helper()
	defer func() {
		got := recover()
		if got == nil {
			t.Errorf("building this should have panicked: %s", want)
			return
		}
		if msg, _ := got.(string); !strings.Contains(msg, want) {
			t.Errorf("panicked with %q, want it to mention %q", msg, want)
		}
	}()
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		view(c)
	}, 400, 300)
}

// inkColumn counts the inked pixels down column x between y0 and y1. Dark is
// ink: every colour this package draws with is either near-black, or white,
// or one of a handful of mid tones that are well above this line.
func inkColumn(img *image.RGBA, x, y0, y1 int) int {
	n := 0
	for y := y0; y < y1; y++ {
		if atPx(img, float32(x), float32(y)).R < 128 {
			n++
		}
	}
	return n
}

// inkAfter returns the first column at or after x that is a bar from y0 to
// y1 — two thirds of the height at least, rather than exactly all of it,
// because the caret is a little shorter than the row it sits in and its ends
// are antialiased. Two thirds is what tells a caret from a letter: a glyph is
// at most half of a column here, a caret is most of it.
func inkAfter(img *image.RGBA, x, y0, y1 int) (int, bool) {
	bar := barInk(y0, y1)
	for col := x; col < x+80; col++ {
		if inkColumn(img, col, y0, y1) >= bar {
			return col, true
		}
	}
	return 0, false
}

// barInk is how much of a column counts as a bar rather than a letter.
func barInk(y0, y1 int) int { return 2 * (y1 - y0) / 3 }

// inkBarWidth returns how many columns from col onward are all bar.
func inkBarWidth(img *image.RGBA, col, y0, y1 int) int {
	n := 0
	for inkColumn(img, col+n, y0, y1) >= barInk(y0, y1) {
		n++
		if n > 16 {
			break
		}
	}
	return n
}

// inkRuns counts the runs of is-true pixels down column x between y0 and y1.
// It is how a test counts rows of a placeholder without giving each row a
// label, which a skeleton has no business having.
func inkRuns(img *image.RGBA, x int, y0, y1 int, is func(ui.Color) bool) int {
	runs, inside := 0, false
	for y := y0; y < y1; y++ {
		on := is(atPx(img, float32(x), float32(y)))
		if on && !inside {
			runs, inside = runs+1, true
		} else if !on {
			inside = false
		}
	}
	return runs
}

// ── Progress ───────────────────────────────────────────────────────────────

func TestProgressFillsToItsValue(t *testing.T) {
	const label = "Resolving callbacks"
	tt := render(400, 120, func(c *ui.Context) {
		ui.Row(c).Padding(10).Children(func() {
			Progress(c, ProgressOptions{Label: label, Value: 0.4, Width: 200})
		})
	})

	r := box(t, tt, label)
	if r.W != 200 {
		t.Errorf("bar is %v wide, want the 200 it was given", r.W)
	}
	// The filled part is the value and no more: at 40% a pixel a quarter of
	// the way along is ink and one at 60% is still the track.
	light := theme.Light()
	wantColor(t, tt, r.X+r.W*0.25, r.Y+r.H/2, light.Accent, "the filled part")
	wantColor(t, tt, r.X+r.W*0.6, r.Y+r.H/2, light.Surface,
		"the part past 40% is still the track")
}

func TestProgressShowsItsPercentageWhenAsked(t *testing.T) {
	tt := render(420, 120, func(c *ui.Context) {
		ui.Row(c).Padding(10).Children(func() {
			Progress(c, ProgressOptions{
				Label: "Replaying the queue", Value: 0.62, Width: 220, ShowPercent: true,
			})
		})
	})
	wantTexts(t, tt, "Replaying the queue", "62%")
}

func TestProgressWithAnUnknownValueStillDraws(t *testing.T) {
	tt := render(400, 120, func(c *ui.Context) {
		ui.Row(c).Padding(10).Children(func() {
			Progress(c, ProgressOptions{Label: "Backfilling", Value: -1, Width: 200})
		})
	})
	r := box(t, tt, "Backfilling")
	if r.W != 200 {
		t.Errorf("bar is %v wide, want 200", r.W)
	}
	// No percentage is claimed: an unknown length has no number to print,
	// and printing 0% would be a claim about the work rather than the bar.
	wantOnly(t, tt, "Backfilling")
}

func TestProgressRejectsWhatItCannotDraw(t *testing.T) {
	wantsPanic(t, "Progress needs a Label", func(c *ui.Context) {
		Progress(c, ProgressOptions{Value: 0.5})
	})
	wantsPanic(t, "0 to 1", func(c *ui.Context) {
		Progress(c, ProgressOptions{Label: "Sync", Value: 42})
	})
}

// ── StatusIndicator ────────────────────────────────────────────────────────

func TestStatusIndicatorTakesItsColoursFromTheSeverity(t *testing.T) {
	tt := render(320, 120, func(c *ui.Context) {
		ui.Row(c).Padding(10).Children(func() {
			StatusIndicator(c, StatusIndicatorOptions{
				Label: "On hold", Severity: core.Warning, Pill: true,
			})
		})
	})
	wantTexts(t, tt, "On hold")

	r := box(t, tt, "On hold")
	bg, _ := core.Warning.Pair(theme.Light())
	// The pill's own background, taken from the same Pair the board's
	// priority pills use: at its widest point on the left edge, clear of the
	// mark and the word.
	wantColor(t, tt, r.X+1, r.Y+r.H/2, bg,
		"a pill backed by the severity's own background")
}

func TestStatusIndicatorInDarkMode(t *testing.T) {
	tt := renderDark(320, 120, func(c *ui.Context) {
		ui.Row(c).Padding(10).Children(func() {
			StatusIndicator(c, StatusIndicatorOptions{
				Label: "Resolved", Severity: core.Success, Pill: true,
			})
		})
	})
	wantTexts(t, tt, "Resolved")

	r := box(t, tt, "Resolved")
	bg, _ := core.Success.Pair(theme.Dark())
	wantColor(t, tt, r.X+1, r.Y+r.H/2, bg,
		"the dark palette's success background, not the light one")
}

func TestStatusIndicatorCanBeBusy(t *testing.T) {
	tt := render(320, 120, func(c *ui.Context) {
		ui.Row(c).Padding(10).Children(func() {
			StatusIndicator(c, StatusIndicatorOptions{Label: "Syncing", Busy: true})
		})
	})
	// A busy status still says what it is: the loader inside it is labelled
	// with the status, so a screen reader reads "Syncing" rather than
	// "Loading".
	wantTexts(t, tt, "Syncing")
	if !tt.HasText("Syncing") {
		t.Error("a busy status lost its own name")
	}
}

func TestStatusIndicatorNeedsSomethingToSay(t *testing.T) {
	wantsPanic(t, "StatusIndicator needs a Label", func(c *ui.Context) {
		StatusIndicator(c, StatusIndicatorOptions{Severity: core.Danger})
	})
}

// ── Presence ───────────────────────────────────────────────────────────────

func TestPresenceMarksWhoIsThere(t *testing.T) {
	tt := render(320, 120, func(c *ui.Context) {
		ui.Row(c).Padding(10).Children(func() {
			Presence(c, PresenceOptions{Name: "Ada Lovelace", State: PresenceOnline})
			Presence(c, PresenceOptions{Name: "Grace Hopper", State: PresenceAway,
				ShowName: true})
		})
	})
	wantTexts(t, tt, "Ada Lovelace Online", "Grace Hopper Away")

	// The online dot is Lively — the one colour that does not follow the
	// appearance — and a test that does not say so would not notice it being
	// replaced by TextMuted.
	online := box(t, tt, "Ada Lovelace Online")
	wantColor(t, tt, online.X+online.W/2, online.Y+online.H/2,
		theme.Light().Lively, "the online dot")
}

func TestPresenceNeedsAName(t *testing.T) {
	wantsPanic(t, "Presence needs a Name", func(c *ui.Context) {
		Presence(c, PresenceOptions{State: PresenceOnline})
	})
}

// ── Result ─────────────────────────────────────────────────────────────────

func TestResultNamesTheOutcome(t *testing.T) {
	// A press is reported for the frame it happened in, so the caller acts
	// there — which is why the flag is latched here rather than read off a
	// result rebuilt every frame. This is the pattern every press assertion
	// in the library's tests uses.
	pressed := false
	tt := render(520, 360, func(c *ui.Context) {
		ui.Box(c).Padding(20).Children(func() {
			if Result(c, ResultOptions{
				Title:    "Callback resolved",
				Body:     "Grace Hopper confirmed the fix at 14:20.",
				Severity: core.Success,
				Action:   "Open the callback",
			}).Pressed() {
				pressed = true
			}
		})
	})
	wantTexts(t, tt, "Callback resolved", "Grace Hopper confirmed", "Open the callback")

	if err := tt.Click("Open the callback"); err != nil {
		t.Fatal(err)
	}
	if !pressed {
		t.Error("the result's action did not report its press")
	}
}

func TestResultInDarkModeDrawsItsMark(t *testing.T) {
	tt := renderDark(520, 360, func(c *ui.Context) {
		ui.Box(c).Padding(20).Children(func() {
			Result(c, ResultOptions{
				Title: "Import failed", Body: "Two rows had no customer.",
				Severity: core.Danger,
			})
		})
	})
	wantTexts(t, tt, "Import failed")
	// The panel is on the dark surface and the words are the dark text
	// colour: a component that read the light palette would put white words
	// on a near-black panel and the test would catch it here.
	r := box(t, tt, "Import failed")
	k := theme.Dark()
	wantColor(t, tt, r.X+r.W/2, r.Y+2, k.Surface, "the panel's surface in the dark palette")
}

func TestResultNeedsAnOutcome(t *testing.T) {
	wantsPanic(t, "Result needs a Title", func(c *ui.Context) {
		Result(c, ResultOptions{Body: "Something happened."})
	})
}

// ── Alert ──────────────────────────────────────────────────────────────────

func TestAlertCarriesItsSeverityAndItsAction(t *testing.T) {
	var actioned, pressed bool
	tt := render(520, 200, func(c *ui.Context) {
		ui.Box(c).Padding(10).Children(func() {
			res := Alert(c, AlertOptions{
				Title:    "Sync failed",
				Body:     "Riverside Clinic did not answer.",
				Severity: core.Danger,
				Action:   "Retry",
			})
			actioned = actioned || res.Actioned()
			pressed = pressed || res.Pressed()
		})
	})
	wantTexts(t, tt, "Sync failed", "Riverside Clinic did not answer.", "Retry")

	r := box(t, tt, "Sync failed")
	// The leading rule is the severity's foreground, not a tint of it: it is
	// the one place a severity can be shown without touching the words.
	wantColor(t, tt, r.X+14+3, r.Y+r.H/2, theme.Light().Danger, "the leading rule")

	if err := tt.Click("Retry"); err != nil {
		t.Fatal(err)
	}
	if !actioned {
		t.Error("the alert's action did not report its press")
	}
	if !pressed {
		t.Error("an alert with one button should report that button as the press")
	}
}

func TestAlertInDarkMode(t *testing.T) {
	tt := renderDark(520, 200, func(c *ui.Context) {
		ui.Box(c).Padding(10).Children(func() {
			Alert(c, AlertOptions{
				Title:    "Quota reached",
				Body:     "Two hundred callbacks a day is the plan's limit.",
				Severity: core.Warning,
			})
		})
	})
	wantTexts(t, tt, "Quota reached")
	r := box(t, tt, "Quota reached")
	wantColor(t, tt, r.X+14+3, r.Y+r.H/2, theme.Dark().Warning,
		"the dark palette's warning rule")
	// And the alert is on the dark surface, not the light one: a component
	// that read the wrong palette would show the warning rule in the right
	// colour on a panel the wrong colour.
	wantColor(t, tt, r.X+r.W-6, r.Y+r.H/2, theme.Dark().Surface,
		"the alert's own surface in the dark palette")
	wantNotColor(t, tt, r.X+r.W-6, r.Y+r.H/2, theme.Light().Surface,
		"the light palette's surface")
}

func TestAlertCanBeDismissed(t *testing.T) {
	pressed := false
	tt := render(520, 200, func(c *ui.Context) {
		ui.Box(c).Padding(10).Children(func() {
			if Alert(c, AlertOptions{
				Title: "Sync failed", Severity: core.Danger, Dismissable: true,
			}).Pressed() {
				pressed = true
			}
		})
	})
	// A close named after what it closes, so a list of alerts does not offer
	// the same unlabelled × a dozen times to a screen reader.
	if err := tt.Click("Dismiss Sync failed"); err != nil {
		t.Fatal(err)
	}
	if !pressed {
		t.Error("the close button did not report its press")
	}
}

func TestAlertNeedsSomethingToSay(t *testing.T) {
	wantsPanic(t, "Alert needs a Title or a Body", func(c *ui.Context) {
		Alert(c, AlertOptions{Severity: core.Danger})
	})
}

// ── ActivityFeed ───────────────────────────────────────────────────────────

func TestActivityFeedListsWhatHappened(t *testing.T) {
	tt := render(520, 320, func(c *ui.Context) {
		ui.Box(c).Padding(10).Children(func() {
			ActivityFeed(c, ActivityFeedOptions{Rule: true, Items: []ActivityItem{
				{Title: "Callback assigned", Detail: "To Ada Lovelace",
					Time: "14:02", Severity: core.Accent},
				{Title: "Customer called back", Detail: "Fix confirmed",
					Time: "14:11", Severity: core.Success},
				{Title: "Sync failed", Detail: "Riverside Clinic",
					Time: "14:20", Severity: core.Danger},
			}})
		})
	})
	wantTexts(t, tt, "Callback assigned", "To Ada Lovelace", "14:02",
		"Customer called back", "Fix confirmed", "14:11",
		"Sync failed", "Riverside Clinic", "14:20")

	// Every row starts at the same x, which is the only thing a column of
	// marks lines up on: a feed whose rows are left edges of text would be
	// ragged the moment one title wrapped.
	for _, title := range []string{"Callback assigned", "Customer called back", "Sync failed"} {
		r := box(t, tt, title)
		if r.X != 10 {
			t.Errorf("row %q starts at x=%v, want the panel's own left edge 10", title, r.X)
		}
		if r.Y <= 0 {
			t.Errorf("row %q is at y=%v, which is not below anything", title, r.Y)
		}
	}
	// And they run downwards, in the order given.
	first, second := box(t, tt, "Callback assigned"), box(t, tt, "Customer called back")
	if second.Y <= first.Y {
		t.Errorf("the second row is at y=%v, not below the first at %v", second.Y, first.Y)
	}
}

func TestActivityFeedWithoutRules(t *testing.T) {
	tt := render(520, 240, func(c *ui.Context) {
		ui.Box(c).Padding(10).Children(func() {
			ActivityFeed(c, ActivityFeedOptions{NoRule: true, Items: []ActivityItem{
				{Title: "Callback assigned", Time: "14:02"},
				{Title: "Sync failed", Time: "14:20", Severity: core.Danger},
			}})
		})
	})
	wantTexts(t, tt, "Callback assigned", "Sync failed")
}

func TestActivityFeedNeedsSomethingToList(t *testing.T) {
	wantsPanic(t, "at least one item", func(c *ui.Context) {
		ActivityFeed(c, ActivityFeedOptions{})
	})
}

// ── Skeleton, Shimmer, ShimmerText ─────────────────────────────────────────

func TestSkeletonFillsTheBoxItIsGiven(t *testing.T) {
	tt := render(400, 200, func(c *ui.Context) {
		ui.Box(c).Padding(10).Children(func() {
			Skeleton(c, SkeletonOptions{Width: 180, Height: 24, Label: "name"})
			Skeleton(c, SkeletonOptions{FillWidth: true, Height: 24, Label: "meta"})
		})
	})
	r := box(t, tt, "name")
	if r.W != 180 || r.H != 24 {
		t.Errorf("skeleton is %v×%v, want 180×24", r.W, r.H)
	}
	wantColor(t, tt, r.X+r.W/2, r.Y+r.H/2, theme.Light().SurfaceHover,
		"the placeholder's fill")

	// FillWidth takes the rest of the panel, which is what makes a column of
	// placeholders line up with the column it is standing in for.
	full := box(t, tt, "meta")
	if full.W <= r.W {
		t.Errorf("a fill-width skeleton is %v wide, want more than the fixed %v",
			full.W, r.W)
	}
}

func TestSkeletonNeedsASize(t *testing.T) {
	wantsPanic(t, "Skeleton needs a Width", func(c *ui.Context) {
		Skeleton(c, SkeletonOptions{Height: 20})
	})
}

func TestShimmerCrossesASkeletonAndLeavesItAloneWhenStill(t *testing.T) {
	// Reduced motion: the highlight is not drawn at all, so every pixel of the
	// block is the placeholder's own fill. This is the assertion the
	// preference can actually be caught by — "the highlight is at the far
	// left" would only ever happen at one phase of the clock.
	still := renderStill(400, 200, func(c *ui.Context) {
		ui.Box(c).Padding(10).Children(func() {
			Shimmer(c, ShimmerOptions{Label: "customer"}, func() {
				Skeleton(c, SkeletonOptions{FillWidth: true, Height: 24, Label: "customer"})
			})
		})
	})
	r := box(t, still, "customer")
	fill := theme.Light().SurfaceHover
	for _, x := range []float32{0.1, 0.35, 0.6, 0.9} {
		wantColor(t, still, r.X+r.W*x, r.Y+r.H/2, fill,
			"a still shimmer leaves the whole block as the placeholder")
	}

	// Not still: the block is still drawn, in the same place and the same
	// colour, and it is the highlight moving over it that changes.
	moving := render(400, 200, func(c *ui.Context) {
		ui.Box(c).Padding(10).Children(func() {
			Shimmer(c, ShimmerOptions{Label: "customer"}, func() {
				Skeleton(c, SkeletonOptions{FillWidth: true, Height: 24, Label: "customer"})
			})
		})
	})
	mr := box(t, moving, "customer")
	if mr.W != r.W || mr.H != r.H {
		t.Errorf("the shimmer moved the block: %v×%v against %v×%v",
			mr.W, mr.H, r.W, r.H)
	}
}

func TestShimmerTextDrawsOneLinePerLine(t *testing.T) {
	const optsWidth = 240
	tt := renderStill(400, 260, func(c *ui.Context) {
		ui.Box(c).Padding(10).Children(func() {
			ShimmerText(c, ShimmerTextOptions{
				Lines: 3, Width: optsWidth, Label: "summary",
			})
		})
	})
	wantTexts(t, tt, "summary")
	r := box(t, tt, "summary")

	// Three runs of placeholder down a column inside the block, which is how
	// a paragraph of three lines is drawn. The last line is short, so the
	// column at its left edge crosses three runs and a column near the right
	// crosses only two.
	img := tt.Image()
	x := int(r.X) + 4
	if got := inkRuns(img, x, int(r.Y), int(r.Y+r.H), func(c ui.Color) bool {
		return sameColor(c, theme.Light().SurfaceHover)
	}); got != 3 {
		t.Errorf("column x=%d crosses %d lines of placeholder, want 3", x, got)
	}
	// The last line is 0.55 of the width, so a column past that crosses only
	// the two full lines: a paragraph whose last line were full width would
	// read as a table rather than as prose.
	short := int(r.X) + int(optsWidth*0.8)
	if got := inkRuns(img, short, int(r.Y), int(r.Y+r.H), func(c ui.Color) bool {
		return sameColor(c, theme.Light().SurfaceHover)
	}); got != 2 {
		t.Errorf("column x=%d crosses %d lines, want 2: the last line should "+
			"be the short one", short, got)
	}
}

func TestShimmerTextRejectsWhatItCannotDraw(t *testing.T) {
	wantsPanic(t, "at least one line", func(c *ui.Context) {
		ShimmerText(c, ShimmerTextOptions{Width: 200})
	})
	wantsPanic(t, "ShimmerText needs a Width", func(c *ui.Context) {
		ShimmerText(c, ShimmerTextOptions{Lines: 2})
	})
}

// ── CountUp, Typewriter ────────────────────────────────────────────────────

func TestCountUpShowsTheFigureItIsGiven(t *testing.T) {
	tt := render(400, 200, func(c *ui.Context) {
		ui.Box(c).Padding(20).Children(func() {
			CountUp(c, CountUpOptions{Value: 1234, Suffix: " callbacks", Bold: true})
		})
	})
	// Two names, and on the first frame they are the same: what it draws is
	// the value it has, and what it says is the value it is heading for.
	// They only come apart once a change is counting, which is the case the
	// next test pins down.
	wantOnly(t, tt, "1,234 callbacks", "1,234 callbacks")
	if r := box(t, tt, "1,234 callbacks"); r.W == 0 || r.H == 0 {
		t.Errorf("the figure is %v×%v, want something drawn", r.W, r.H)
	}
}

func TestCountUpNamesTheValueItIsHeadingFor(t *testing.T) {
	tt := render(400, 200, func(c *ui.Context) {
		ui.Box(c).Padding(20).Children(func() {
			CountUp(c, CountUpOptions{Value: 4820, Prefix: "$"})
		})
	})
	r := box(t, tt, "$4,820")
	if r.W == 0 || r.H == 0 {
		t.Errorf("the figure is %v×%v, want something drawn", r.W, r.H)
	}
	wantTexts(t, tt, "$4,820")
}

func TestCountUpIsTheNumberNowWhenMotionIsReduced(t *testing.T) {
	// Reduced motion is not a shorter count-up: a figure that ticks part way
	// to its value leaves the reader holding a number that is not the number.
	still := renderStill(400, 200, func(c *ui.Context) {
		ui.Box(c).Padding(20).Children(func() {
			CountUp(c, CountUpOptions{Value: 4820, Prefix: "$"})
		})
	})
	wantOnly(t, still, "$4,820", "$4,820")
}

func TestTypewriterShowsOnlyWhatItIsTold(t *testing.T) {
	tt := render(400, 200, func(c *ui.Context) {
		ui.Box(c).Padding(20).Children(func() {
			Typewriter(c, TypewriterOptions{Text: "Callback resolved", Shown: 8})
		})
	})
	// The row is named after the whole line and holds the prefix, so the two
	// are distinguishable: a test that only looked at Texts would not be able
	// to tell "only the prefix is drawn" from "the whole line is drawn".
	wantOnly(t, tt, "Callback resolved", "Callback")
}

func TestTypewriterIsNamedAfterTheWholeLine(t *testing.T) {
	tt := render(400, 200, func(c *ui.Context) {
		ui.Box(c).Padding(20).Children(func() {
			Typewriter(c, TypewriterOptions{Text: "Callback resolved", Shown: 8})
		})
	})
	// Findable by the whole text, though only a prefix is on screen: a
	// screen reader should hear the line rather than its first eight
	// characters.
	r := box(t, tt, "Callback resolved")
	prefix := box(t, tt, "Callback")
	if r.W <= prefix.W {
		t.Errorf("the line is %v wide and its prefix alone is %v: the row is "+
			"not wrapping the drawn text", r.W, prefix.W)
	}
}

func TestTypewriterCaretIsStillWhenMotionIsReduced(t *testing.T) {
	still := renderStill(400, 200, func(c *ui.Context) {
		ui.Box(c).Padding(20).Children(func() {
			Typewriter(c, TypewriterOptions{Text: "Resolving", Shown: 4})
		})
	})
	r := box(t, still, "Resolving")
	text := box(t, still, "Reso")
	y0, y1 := int(r.Y)+1, int(r.Y+r.H)-1

	// The caret is a solid bar as tall as the line, which is what tells it
	// apart from the letters beside it: a glyph is a handful of pixels in a
	// column, a caret is the whole height of it.
	img := still.Image()
	caret, ok := inkAfter(img, int(text.X+text.W)-1, y0, y1)
	if !ok {
		t.Fatalf("no full-height caret after the drawn prefix %q", "Reso")
	}
	wantColor(t, still, float32(caret)+0.5, r.Y+r.H/2, theme.Light().Text,
		"a caret drawn in the text colour")
	// It is a thin bar standing on its own: as wide as the component says a
	// caret is, and no taller-looking than the text either side of it. A bar
	// found inside a glyph would be a letter, not a caret.
	width := inkBarWidth(img, caret, y0, y1)
	w, _ := caretSize(theme.BodySize)
	if int(w) < width || width > int(w)+2 {
		t.Errorf("the bar is %d columns wide, want the %v a caret is", width, w)
	}
	tall := inkColumn(img, caret, y0, y1)
	for _, side := range []int{caret - 1, caret + width} {
		if inkColumn(img, side, y0, y1) >= tall {
			t.Errorf("column %d is %d inked, no less than the bar's %d: what "+
				"was found is part of the text and not a caret",
				side, inkColumn(img, side, y0, y1), tall)
		}
	}

	// And the policy itself, which is the part a rendered frame cannot show:
	// the caret fades only when the window lets it move.
	if got := caretInk(true, 0); got != 1 {
		t.Errorf("a still caret is drawn at %v ink, want it fully drawn", got)
	}
	if got := caretInk(false, 0); got >= 1 {
		t.Errorf("a moving caret at the back of its blink is %v ink, want it "+
			"faded — otherwise it never blinks", got)
	}
}

func TestTypewriterRejectsAPrefixBeforeTheStart(t *testing.T) {
	wantsPanic(t, "Typewriter needs Text", func(c *ui.Context) {
		Typewriter(c, TypewriterOptions{Shown: 3})
	})
	wantsPanic(t, "cannot be before the start", func(c *ui.Context) {
		Typewriter(c, TypewriterOptions{Text: "Saved", Shown: -1})
	})
}

// ── BlinkHighlight ─────────────────────────────────────────────────────────

func TestBlinkHighlightTintsToTheLevelItIsGiven(t *testing.T) {
	// The level is the caller's number rather than a clock inside the
	// component, which is what makes this determinate: level 1 tints now,
	// and level 0 does not, in whichever frame it is asked.
	lit := render(400, 200, func(c *ui.Context) {
		ui.Box(c).Padding(10).Children(func() {
			BlinkHighlight(c, BlinkHighlightOptions{Level: 1, Radius: 8}, func() {
				ui.Box(c).FillWidth().Height(40).Label("row").Children(func() {
					ui.Text(c, "Riverside Clinic")
				})
			})
		})
	})
	r := box(t, lit, "row")
	wantInk(t, lit, r.X+r.W/2, r.Y+r.H-2, inkOf(theme.Light().Accent.Alpha(0.2)),
		"a highlight at level 1")

	unlit := render(400, 200, func(c *ui.Context) {
		ui.Box(c).Padding(10).Children(func() {
			BlinkHighlight(c, BlinkHighlightOptions{Level: 0, Radius: 8}, func() {
				ui.Box(c).FillWidth().Height(40).Label("row").Children(func() {
					ui.Text(c, "Riverside Clinic")
				})
			})
		})
	})
	ur := box(t, unlit, "row")
	wantColor(t, unlit, ur.X+ur.W/2, ur.Y+ur.H-2,
		theme.Light().Background, "level 0 draws no tint")
}

func TestBlinkHighlightIsInertWhenMotionIsReduced(t *testing.T) {
	// Reduced motion keeps the highlight off whatever level it is given: a
	// fade is motion whether or not anything moved, and the person who asked
	// for less motion asked for less of it everywhere.
	row := func(level float32) func(c *ui.Context) {
		return func(c *ui.Context) {
			ui.Box(c).Padding(10).Children(func() {
				BlinkHighlight(c, BlinkHighlightOptions{Level: level, Radius: 8},
					func() {
						ui.Box(c).FillWidth().Height(40).Label("row").Children(func() {
							ui.Text(c, "Riverside Clinic")
						})
					})
			})
		}
	}
	still := renderStill(400, 200, row(1))
	lit := render(400, 200, row(1))
	r, lr := box(t, still, "row"), box(t, lit, "row")

	wantColor(t, still, r.X+r.W/2, r.Y+r.H-2, theme.Light().Background,
		"a still highlight draws the row as it would be without the component")
	wantInk(t, lit, lr.X+lr.W/2, lr.Y+lr.H-2, inkOf(theme.Light().Accent.Alpha(0.2)),
		"the same row with motion on")
}

// ── LoadingOverlay ─────────────────────────────────────────────────────────

func TestLoadingOverlayCoversWhatItWaitsFor(t *testing.T) {
	dismissed := false
	tt := render(520, 320, func(c *ui.Context) {
		ui.Box(c).Padding(20).Children(func() {
			res := LoadingOverlay(c, LoadingOverlayOptions{
				Message: "Importing 4,812 callbacks", Dismissable: true,
			}, func() {
				ui.Box(c).FillWidth().Height(120).Background(theme.Light().Surface).
					Label("table")
			})
			dismissed = dismissed || res.Dismissed()
		})
	})
	wantTexts(t, tt, "Importing 4,812 callbacks", "Cancel")

	// The content underneath is still in the tree, still where it was, and
	// still its own size: the overlay covers it rather than replacing it,
	// which is the whole reason to have one.
	r := box(t, tt, "table")
	if r.W == 0 || r.H != 120 {
		t.Errorf("the covered content is %v×%v, want the 120-tall row it is "+
			"standing in for", r.W, r.H)
	}

	// And the scrim is over it: a pixel near its left edge, clear of the
	// centred panel, is the fill colour at the scrim's alpha over the
	// surface rather than the surface itself.
	wantInk(t, tt, r.X+8, r.Y+4,
		theme.Light().Fill.Alpha(0.55).Over(theme.Light().Surface),
		"the scrim over the content it covers")
	wantColor(t, tt, r.X+r.W/2, r.Y+r.H/2, theme.Light().Background,
		"the panel, which the scrim lets through")

	if err := tt.Click("Cancel"); err != nil {
		t.Fatal(err)
	}
	if !dismissed {
		t.Error("the overlay's cancel did not report its press")
	}
}

func TestLoadingOverlayNeedsSomethingToWaitFor(t *testing.T) {
	wantsPanic(t, "LoadingOverlay needs a Message", func(c *ui.Context) {
		LoadingOverlay(c, LoadingOverlayOptions{}, nil)
	})
}

// ── NotificationCenter ─────────────────────────────────────────────────────

func TestNotificationCenterListsWhatIsOutstanding(t *testing.T) {
	dismissed, cleared := -1, false
	tt := render(520, 420, func(c *ui.Context) {
		ui.Box(c).Padding(10).Children(func() {
			res := NotificationCenter(c, NotificationCenterOptions{
				Title: "Notifications", Clearable: true, Rule: true,
				Items: []Notification{
					{Title: "Sync failed", Body: "Riverside Clinic", At: "14:20",
						Severity: core.Danger},
					{Title: "Two callbacks need a callback", At: "14:04",
						Severity: core.Warning},
				},
			})
			if res.Dismissed() >= 0 {
				dismissed = res.Dismissed()
			}
			cleared = cleared || res.Cleared()
		})
	})
	wantTexts(t, tt, "Notifications", "Clear all", "Sync failed",
		"Riverside Clinic", "14:20", "Two callbacks need a callback", "14:04")

	// The close is named after what it closes, so a list of six does not
	// offer the same unlabelled × six times.
	if err := tt.Click("Dismiss Two callbacks need a callback"); err != nil {
		t.Fatal(err)
	}
	if dismissed != 1 {
		t.Errorf("Dismissed() = %d, want 1 — the index of the row closed", dismissed)
	}

	if err := tt.Click("Clear all"); err != nil {
		t.Fatal(err)
	}
	if !cleared {
		t.Error("clear all did not report its press")
	}
}

func TestNotificationCenterSaysWhenThereIsNothing(t *testing.T) {
	tt := render(520, 240, func(c *ui.Context) {
		ui.Box(c).Padding(10).Children(func() {
			NotificationCenter(c, NotificationCenterOptions{Title: "Notifications"})
		})
	})
	// An empty centre is not an error: a window with nothing to say is a
	// window at rest, and saying so beats drawing nothing, which looks like
	// the panel failed.
	wantTexts(t, tt, "Notifications", "Nothing to report")
}

func TestNotificationCenterNeedsSomethingToSay(t *testing.T) {
	wantsPanic(t, "Notification needs a Title", func(c *ui.Context) {
		NotificationCenter(c, NotificationCenterOptions{
			Items: []Notification{{Body: "Riverside Clinic did not answer."}},
		})
	})
}

func TestNotificationCenterCopyCanBeReworded(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		c = core.WithMessages(c, map[string]string{
			"feedback.notificationCenter.empty": "没有待处理的通知",
		})
		ui.Box(c).Fill().Children(func() {
			NotificationCenter(c, NotificationCenterOptions{Title: "通知"})
		})
	}, 520, 240)
	wantTexts(t, tt, "没有待处理的通知")
}

// ── UndoToast ──────────────────────────────────────────────────────────────

func TestUndoToastOffersTheUndo(t *testing.T) {
	undo := false
	tt := render(520, 160, func(c *ui.Context) {
		ui.Box(c).Padding(20).Children(func() {
			undo = undo || UndoToast(c, UndoToastOptions{
				Message: "Callback archived", Action: "Undo",
			}).Undo()
		})
	})
	wantTexts(t, tt, "Callback archived", "Undo")

	// The toast is drawn on the fill surface, the way Toast is, and takes no
	// severity: there is no background here to pair one with.
	r := box(t, tt, "Callback archived")
	wantColor(t, tt, r.X+r.W/2, r.Y+1, theme.Light().Fill, "the toast's surface")

	if err := tt.Click("Undo"); err != nil {
		t.Fatal(err)
	}
	if !undo {
		t.Error("the undo did not report its press")
	}
}

func TestUndoToastWithoutAnAction(t *testing.T) {
	tt := render(520, 160, func(c *ui.Context) {
		ui.Box(c).Padding(20).Children(func() {
			UndoToast(c, UndoToastOptions{Message: "Callback archived"})
		})
	})
	wantTexts(t, tt, "Callback archived")
	if err := tt.Click("Undo"); err == nil {
		t.Error("an undo toast with no action should not draw an undo")
	}
}

func TestUndoToastNeedsSomethingToUndo(t *testing.T) {
	wantsPanic(t, "UndoToast needs a Message", func(c *ui.Context) {
		UndoToast(c, UndoToastOptions{Action: "Undo"})
	})
}

// ── LayoutTransition, Stagger ──────────────────────────────────────────────

func TestLayoutTransitionNeedsAKey(t *testing.T) {
	// MyGo will fall back on a sibling's place, which breaks the moment the
	// list reorders — the only reason to want a transition at all. So this
	// component refuses rather than animating the wrong element later.
	wantsPanic(t, "LayoutTransition needs a Key", func(c *ui.Context) {
		LayoutTransition(c, LayoutTransitionOptions{}, func() {
			ui.Text(c, "row")
		})
	})
}

func TestLayoutTransitionDrawsItsChildEitherWay(t *testing.T) {
	view := func(c *ui.Context) {
		ui.Box(c).Padding(10).Children(func() {
			LayoutTransition(c, LayoutTransitionOptions{Key: "row-1"}, func() {
				// Named after its own text, which is what Label does when it
				// is set to the same string: the element is named *and* draws
				// its text, rather than one replacing the other.
				ui.Text(c, "Riverside Clinic").Label("Riverside Clinic")
			})
		})
	}
	for name, tt := range map[string]*ui.Tester{
		"moving":     render(400, 200, view),
		"not moving": renderStill(400, 200, view),
	} {
		r := box(t, tt, "Riverside Clinic")
		if r.W == 0 || r.H == 0 {
			t.Errorf("%s: the child is %v×%v, want a drawn row", name, r.W, r.H)
		}
	}
}

func TestLayoutTransitionIsInertWhenMotionIsReduced(t *testing.T) {
	// A dead spec is not one with a zero duration: ui.ElementTransition reads
	// zero as "use the default", so a dead spec handed over would start a
	// 200ms animation in exactly the windows that asked for none.
	var spec MotionSpec
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		core.WithReducedMotion(c, true)
		spec = transitionSpec(c, 220, fade, fade, true, false)
	}, 200, 100)
	if spec.Animates() {
		t.Errorf("a window with reduced motion still got a %v transition", spec.Duration)
	}

	var moving MotionSpec
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		moving = transitionSpec(c, 220, fade, fade, true, false)
	}, 200, 100)
	if !moving.Animates() {
		t.Error("a window that did not reduce motion got no transition at all")
	}
	if moving.Duration.Milliseconds() != 220 {
		t.Errorf("transition duration = %v, want the 220 that was asked for",
			moving.Duration.Milliseconds())
	}
}

func TestStaggerGivesEachItemItsOwnEntrance(t *testing.T) {
	var s *Stagger
	var first, third MotionSpec
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		s = NewStagger(c, StaggerOptions{})
		first, third = s.spec(0), s.spec(2)
	}, 200, 100)

	if !first.Animates() || !third.Animates() {
		t.Fatal("a window that did not reduce motion should stagger")
	}
	wantGap := 2 * time.Duration(defaultStaggerStep) * time.Millisecond
	if got := third.Duration - first.Duration; got != wantGap {
		t.Errorf("item 2 starts %v after item 0, want %v: the step is what "+
			"puts the items in turn", got, wantGap)
	}
}

func TestStaggerIsInertWhenMotionIsReduced(t *testing.T) {
	// A list that appears one item at a time is the same motion as any other,
	// only slower — so a window that asked for none gets the whole list at
	// once rather than the same list stretched over half a second.
	var s *Stagger
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		core.WithReducedMotion(c, true)
		s = NewStagger(c, StaggerOptions{})
		if s.spec(0).Animates() || s.spec(9).Animates() {
			t.Error("a window with reduced motion still staggered its items")
		}
	}, 200, 100)
}

func TestStaggerDrawsEveryItemItWraps(t *testing.T) {
	names := []string{"Ada Lovelace", "Grace Hopper", "Alan Turing"}
	var wrapped int
	tt := render(520, 320, func(c *ui.Context) {
		st := NewStagger(c, StaggerOptions{})
		ui.Column(c).Padding(10).Children(func() {
			for _, name := range names {
				row := ui.Row(c).FillWidth().Label(name).Children(func() {
					ui.Text(c, name)
				})
				st.Wrap(row)
				wrapped++
			}
		})
	})
	wantTexts(t, tt, names...)
	if wrapped != len(names) {
		t.Errorf("the loop wrapped %d items, want %d", wrapped, len(names))
	}
	// And they are in the order they were given, each below the last: the
	// stagger is what makes a list arrive as a list, so it has not taken the
	// rows out of their order to do it.
	prev := float32(-1)
	for _, name := range names {
		r := box(t, tt, name)
		if r.Y <= prev {
			t.Errorf("row %q is at y=%v, not below %v", name, r.Y, prev)
		}
		prev = r.Y
	}
}

// ── InterruptButton ────────────────────────────────────────────────────────

func TestInterruptButtonStopsWhatIsRunning(t *testing.T) {
	pressed := false
	tt := render(400, 160, func(c *ui.Context) {
		ui.Box(c).Padding(20).Children(func() {
			pressed = pressed || InterruptButton(c, InterruptButtonOptions{}).Pressed()
		})
	})
	wantTexts(t, tt, "Stop")

	if err := tt.Click("Stop"); err != nil {
		t.Fatal(err)
	}
	if !pressed {
		t.Error("the interrupt did not report its press")
	}
}

func TestInterruptButtonSaysThatItHasBeenAsked(t *testing.T) {
	pressed := false
	tt := render(400, 160, func(c *ui.Context) {
		ui.Box(c).Padding(20).Children(func() {
			pressed = pressed || InterruptButton(c,
				InterruptButtonOptions{Busy: true}).Pressed()
		})
	})
	// It says what is happening rather than what it would do: "Stopping…"
	// tells the reader their press landed, where a greyed-out "Stop" tells
	// them they may have already pressed it.
	for _, name := range tt.Texts() {
		if name != "Stopping…" {
			t.Errorf("a busy interrupt offers %q; it should only say it is "+
				"stopping, not that it is still offering to stop", name)
		}
	}
	if len(tt.Texts()) == 0 {
		t.Error("a busy interrupt says nothing at all")
	}

	if err := tt.Click("Stopping…"); err != nil {
		t.Fatal(err)
	}
	if pressed {
		t.Error("a busy interrupt took a second press, which is the one it exists " +
			"to prevent")
	}
}

func TestInterruptButtonCanBeRenamed(t *testing.T) {
	tt := render(400, 160, func(c *ui.Context) {
		ui.Box(c).Padding(20).Children(func() {
			InterruptButton(c, InterruptButtonOptions{Label: "Cancel upload"})
		})
	})
	wantTexts(t, tt, "Cancel upload")
}

func TestInterruptButtonIsDangerousUnlessToldOtherwise(t *testing.T) {
	// The default is the danger colour, because the button exists to stop
	// something; Caution is the way out of it, and not a Severity whose zero
	// value is core.Neutral and would quietly make the safe default the
	// undangerous one.
	for _, tc := range []struct {
		name string
		opts InterruptButtonOptions
		want ui.Color
	}{
		{"danger by default", InterruptButtonOptions{}, theme.Light().Danger},
		{"caution when told", InterruptButtonOptions{Caution: true}, theme.Light().Text},
	} {
		tt := render(400, 160, func(c *ui.Context) {
			ui.Box(c).Padding(20).Children(func() {
				InterruptButton(c, tc.opts)
			})
		})
		// The stop square is the button's own foreground, drawn solid, so it
		// is both the mark and the colour under test. It is found by looking
		// for a solid block rather than at a fixed offset, so the assertion
		// is about what was drawn and not about where the padding puts it.
		// The stop square is drawn in the button's own foreground, so it is
		// both the mark and the colour under test. It is found by looking for
		// the first run of pixels that are not the button's own background —
		// the mark is the first thing in the button whatever the label and
		// however the button is centred — and then sampled three pixels in, so
		// the assertion is about the colour rather than about the
		// antialiasing on its edge.
		r := box(t, tt, "Stop")
		img := tt.Image()
		bg := atPx(img, r.X+3, r.Y+r.H/2)
		mark := -1
		for x := int(r.X) + 4; x < int(r.X+r.W); x++ {
			px := atPx(img, float32(x), r.Y+r.H/2)
			// Neither the button's own background nor the window's: the mark
			// is the first thing inside the button that is neither.
			if sameColor(px, bg) || sameColor(px, theme.Light().Background) {
				continue
			}
			mark = x
			break
		}
		if mark < 0 {
			t.Fatalf("%s: the button drew nothing but its own background", tc.name)
		}
		wantColor(t, tt, float32(mark)+3.5, r.Y+r.H/2, tc.want, tc.name)
	}
}

// ── the loaders ────────────────────────────────────────────────────────────

func TestLoadersDrawTheirMark(t *testing.T) {
	const side = 32
	// All five take the same options and the same size, which is the whole
	// reason there is one LoaderOptions: swapping a loader is one word.
	loaders := map[string]func(c *ui.Context, opts LoaderOptions) *ui.Element{
		"Bars":  BarsLoader,
		"Dots":  DotsLoader,
		"Orbit": OrbitLoader,
		"Pulse": PulseLoader,
		"Wave":  WaveLoader,
	}
	for name, fn := range loaders {
		tt := renderStill(200, 120, func(c *ui.Context) {
			ui.Box(c).Padding(10).Children(func() {
				fn(c, LoaderOptions{Size: side, Label: "Loading the queue"})
			})
		})
		r := box(t, tt, "Loading the queue")
		if r.W != side || r.H != side {
			t.Errorf("%s loader is %v×%v, want %v×%v", name, r.W, r.H, side, side)
		}
	}
}

func TestLoadersNameWhatTheyAreLoadingByDefault(t *testing.T) {
	tt := renderStill(200, 120, func(c *ui.Context) {
		ui.Box(c).Padding(10).Children(func() { BarsLoader(c, LoaderOptions{}) })
	})
	wantTexts(t, tt, "Loading")
}

func TestLoadersRestWhenMotionIsReduced(t *testing.T) {
	// A still loader is given a pose it is legible in, not a frame of its
	// loop: a wave caught mid-rise reads as a stuck interface. Each of these
	// asserts that pose, and each one is a different pose — which is the
	// point, and also what makes a loader that forgot the preference fail.
	const side = 40

	t.Run("bars are level", func(t *testing.T) {
		tt := renderStill(120, 120, func(c *ui.Context) {
			ui.Box(c).Padding(10).Children(func() {
				BarsLoader(c, LoaderOptions{Size: side})
			})
		})
		levelBars(t, tt, 10, 10, side, 3)
	})

	t.Run("dots are equally inked", func(t *testing.T) {
		tt := renderStill(120, 120, func(c *ui.Context) {
			ui.Box(c).Padding(10).Children(func() {
				DotsLoader(c, LoaderOptions{Size: side})
			})
		})
		// Every dot at full ink: a dimmed dot at rest reads as a mark that
		// failed rather than as three dots that are not moving.
		equalDots(t, tt, 10, 10, side, 3)
	})

	t.Run("the orbit dot is at the top", func(t *testing.T) {
		tt := renderStill(120, 120, func(c *ui.Context) {
			ui.Box(c).Padding(10).Children(func() {
				OrbitLoader(c, LoaderOptions{Size: side})
			})
		})
		accent := theme.Light().Accent
		rad := float32(side) / 2
		// Straight up: ink. Straight down: the ring's track at a fifth of the
		// ink, which is not the same colour and so not a dot that fell off.
		wantColor(t, tt, float32(10)+rad, 10+rad-(rad-1), accent,
			"the orbit dot, straight up at rest")
		wantNotColor(t, tt, float32(10)+rad, 10+rad+(rad-1), accent,
			"the bottom of the ring, which is a track and not a dot")
		wantNotColor(t, tt, 10+1, 10+1, accent, "the corner, outside the ring")
	})

	t.Run("the pulse dot is at middle size", func(t *testing.T) {
		tt := renderStill(120, 120, func(c *ui.Context) {
			ui.Box(c).Padding(10).Children(func() {
				PulseLoader(c, LoaderOptions{Size: side})
			})
		})
		rad := float32(side) / 2
		// The resting dot is at 0.62 of the box, so its centre is ink and the
		// corner is not — and the centre is the ink at the alpha that size
		// implies, which pins the pose rather than just the presence.
		wantInk(t, tt, float32(10)+rad, 10+rad,
			inkOf(theme.Light().Accent.Alpha(0.45+0.55*restRadius)),
			"the middle of the resting dot")
		wantColor(t, tt, 11, 11, theme.Light().Background,
			"the corner, which the resting dot does not reach")
		wantColor(t, tt, float32(10)+rad, 10+rad+rad, theme.Light().Background,
			"just past the resting dot's edge, below and not on it")
	})

	t.Run("the wave run is level", func(t *testing.T) {
		tt := renderStill(120, 120, func(c *ui.Context) {
			ui.Box(c).Padding(10).Children(func() {
				WaveLoader(c, LoaderOptions{Size: side})
			})
		})
		levelBars(t, tt, 10, 10, side, 4)
	})
}

// levelBars asserts that n bars stand at the same height, filling the lower
// half of a side-by-side box at (x, y): ink just above the middle of every
// bar, and nothing just below the top of the box. It is the shape a level
// meter has when it is not moving, which is what a still BarsLoader and a
// still WaveLoader both draw.
func levelBars(t *testing.T, tt *ui.Tester, x, y, side float32, n int) {
	t.Helper()
	accent := theme.Light().Accent
	gap := side * 0.16
	bar := (side - gap*float32(n-1)) / float32(n)
	top, mid := y+side/2-1, y+side/2+1
	for i := range n {
		bx := x + float32(i)*(bar+gap) + bar/2
		wantColor(t, tt, bx, mid, accent, "the filled half of a resting bar")
		wantNotColor(t, tt, bx, top, accent, "above a resting bar")
	}
}

// equalDots asserts that n dots across a side-by-side box are all the same
// ink, which is what a still DotsLoader draws.
func equalDots(t *testing.T, tt *ui.Tester, x, y, side float32, n int) {
	t.Helper()
	accent := theme.Light().Accent
	for i := range n {
		cx := x + side*(float32(i)+0.5)/float32(n)
		wantColor(t, tt, cx, y+side/2, accent, "each resting dot at full ink")
	}
}

// TestLoaderShapesAnswerBothWays is the loaders' policy in one place: the same
// shape functions give the resting pose when the tick is still and a moving
// value when it is not, and the two are never the same. Asserting the moving
// half against a phase rather than against a clock is what makes it a test.
func TestLoaderShapesAnswerBothWays(t *testing.T) {
	still, moving := tick{still: true}, tick{phase: 0.25}

	if got := barHeight(still, 0); got != restBar {
		t.Errorf("a still bar is %v tall, want the %v every bar stands at", got, restBar)
	}
	for i := range 3 {
		if barHeight(still, i) != barHeight(still, 0) {
			t.Errorf("a still bar run is not level: bar %d differs from bar 0", i)
		}
		if dotAlpha(still, i) != restDot {
			t.Errorf("a still dot is %v inked, want the %v they are all drawn at",
				dotAlpha(still, i), restDot)
		}
	}
	if orbitAngle(still) != restAngle {
		t.Errorf("a still orbit dot is at %v rad, want %v — straight up",
			orbitAngle(still), restAngle)
	}
	if pulseSize(still) != restRadius {
		t.Errorf("a still pulse dot is at %v of its box, want %v",
			pulseSize(still), restRadius)
	}

	// Moving: the run is no longer level, the dots no longer match, and the
	// orbit dot is somewhere other than the top.
	level := true
	for i := range 3 {
		if barHeight(moving, i) != barHeight(moving, 0) {
			level = false
		}
	}
	if level {
		t.Error("a moving bar run is level at phase 0.25; the bars are a third of " +
			"a cycle apart and must not be")
	}
	if dotAlpha(moving, 0) == dotAlpha(moving, 1) {
		t.Error("a moving dot run is evenly inked; the dots are a third of a cycle " +
			"apart and must not be")
	}
	if orbitAngle(moving) == restAngle {
		t.Error("a moving orbit dot is still at the top")
	}
}

func TestWaveRunsThroughZeroRatherThanTakingTurns(t *testing.T) {
	// The two bar runs are different shapes, which is the reason there are
	// two loaders and not one with a count: one is three shapes taking turns
	// to be tall, the other is a waveform with crests and troughs.
	moving := tick{phase: 0.25}
	var lo, hi float32 = 2, 0
	for i := range 4 {
		h := waveHeight(moving, i)
		lo, hi = min(lo, h), max(hi, h)
	}
	if lo < restFloor {
		t.Errorf("the wave run dips to %v, below the %v floor: a bar at zero is "+
			"a gap and the run stops reading as a run", lo, restFloor)
	}
	if hi-lo < 0.2 {
		t.Errorf("the wave run spans %v to %v, which is not a waveform", lo, hi)
	}
}
