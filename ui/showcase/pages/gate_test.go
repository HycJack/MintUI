package pages_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/showcase"

	_ "github.com/HycJack/MintUI/ui/showcase/pages"
)

// TestEveryPackageHasAPage is the inventory check. A ui/ package with no
// gallery page is a package nobody can look at, and the point of the gallery
// is that everything can be looked at.
func TestEveryPackageHasAPage(t *testing.T) {
	entries, err := os.ReadDir("../../")
	if err != nil {
		t.Fatal(err)
	}
	var missing []string
	for _, e := range entries {
		if !e.IsDir() || e.Name() == "showcase" {
			continue
		}
		// A package with no components has nothing to show.
		has, err := packageHasComponents(filepath.Join("../../", e.Name()))
		if err != nil || !has {
			continue
		}
		if _, ok := showcase.Pages[e.Name()]; !ok {
			missing = append(missing, e.Name())
		}
	}
	if len(missing) > 0 {
		t.Errorf("%d packages have no gallery page: %v", len(missing), missing)
	}
}

func packageHasComponents(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".go") &&
			!strings.HasSuffix(e.Name(), "_test.go") {
			return true, nil
		}
	}
	return false, nil
}

// TestPagesDrawWhatTheyPromised stops a page from quietly going blank, which
// is the failure mode a gallery cannot otherwise have: the window opens, the
// page is there, and it is empty.
func TestPagesDrawWhatTheyPromised(t *testing.T) {
	if len(showcase.Pages) == 0 {
		t.Fatal("no pages are registered")
	}
	for _, p := range showcase.Sorted() {
		t.Run(p.Package, func(t *testing.T) {
			drawn := p.Draw()
			if miss := p.Missing(drawn); len(miss) > 0 {
				t.Errorf("%s promised text it did not draw: %v", p.Package, miss)
			}
			if len(p.Want) < 3 {
				t.Errorf("%s promises only %d texts — a page that shows almost "+
					"nothing is not a gallery of that package", p.Package, len(p.Want))
			}
		})
	}
}

// TestPagesDrawInBothAppearances: a gallery is also how a dark-palette bug
// gets seen, and a bug that only shows after dark is the one nobody looks for.
//
// Both appearances are rendered here for the same reason useForShowcase is
// always light: the window the gallery runs in follows the desktop, so a page
// has to be looked at twice before either picture can be called the page. A
// component that only paints under one palette — a name computed from the
// light background, a section guarded by IsDark — is invisible in every
// screenshot and in every other test on this file, because Draw renders light
// and nothing else did.
//
// The count is the engine's own, not paintedTexts': this asks whether a page
// put anything on the screen at all under this palette, and the smallest
// answer to that is the strings the frame reported, labels included. Want is
// checked separately, in light, by TestPagesDrawWhatTheyPromised.
//
// The shape this cannot see is a string that is reported, laid out, and then
// painted in the colour it sits on — an ink taken from the window rather than
// from the box it is drawn inside. paintedTexts tells that apart from the
// frame and is not reachable from here, so it goes unseen here, and Want is
// only read under light. A component that does it passes both gates. It is
// worth looking for by eye in both appearances, which is what the gallery is.
func TestPagesDrawInBothAppearances(t *testing.T) {
	for _, p := range showcase.Sorted() {
		for _, mode := range []core.Mode{core.Light, core.Dark} {
			t.Run(p.Package+"/"+modeName(mode), func(t *testing.T) {
				tt := ui.NewTester(func(c *ui.Context) {
					core.Use(c, core.Settings{Mode: mode})
					p.Paint(c)
				}, p.Width, p.Height)
				drawn := tt.Texts()
				if len(drawn) < 3 {
					t.Errorf("%s in %s drew %d texts: %v",
						p.Package, mode, len(drawn), drawn)
				}
			})
		}
	}
}

// modeName names an appearance the way the subtests above are named, in lower
// case. core.Mode already prints one word for each, and going through it keeps
// a third mode from needing a second list here.
func modeName(m core.Mode) string { return strings.ToLower(m.String()) }

// stubPage is a page whose only content is what the test draws, so a test can
// say what a person would be looking at without a gallery page in the way. It
// is never registered: it is a page under the gate, not one of them.
func stubPage(render func(c *ui.Context), want ...string) showcase.Page {
	return showcase.Page{
		Package: "stub", Title: "stub",
		Width: 400, Height: 300,
		Want: want, Render: render,
	}
}

// TestAnAccessibilityLabelIsNotDrawnText: a Want entry is a promise that a
// person can see the text, while a Label is a promise to a screen reader. An
// empty box carrying one has painted nothing, so it cannot be what the page
// promised, and the check going green on such a page is the failure this
// pins.
func TestAnAccessibilityLabelIsNotDrawnText(t *testing.T) {
	p := stubPage(func(c *ui.Context) {
		ui.Box(c).Label("promised").Size(10, 10)
	}, "promised")
	if miss := p.Missing(p.Draw()); len(miss) != 1 {
		t.Fatalf("a Label on an empty box satisfied Want: %v", miss)
	}
}

// TestOnlyTheUnpromisedGoesMissing is the other half of the same promise: a
// string that really is on the page satisfies its promise, or the check would
// be one that can only fail and would be turned off within the week.
//
// The gap is the room a glyph's ink needs: descenders reach past the box the
// engine reports for a line of text, so a box laid right under one picks up
// its neighbour's pixels and reads as painted.
func TestOnlyTheUnpromisedGoesMissing(t *testing.T) {
	p := stubPage(func(c *ui.Context) {
		ui.Column(c).Gap(core.Density(c).Unit() * 4).Children(func() {
			ui.Text(c, "painted")
			ui.Box(c).Label("only named").Size(10, 10)
		})
	}, "painted", "only named")
	if miss := p.Missing(p.Draw()); len(miss) != 1 || miss[0] != "only named" {
		t.Fatalf("Missing = %v, want only the label-only box", miss)
	}
}

// TestPagesRenderDeterministically draws every page twice with animation off
// and requires the two images to be identical.
//
// Animation is excluded on purpose: a spinner legitimately draws differently a
// millisecond later, and a gate that fails on that teaches nobody anything. What
// must not change between two draws of an untouched page is everything else —
// a page that mutates its own state while drawing leaves the next draw (or the
// dark-mode draw) showing something different, which is how a screenshot came
// to be one thing on the check that ran before it and another thing in the PNG
// that followed.
//
// There is no exemption from this. A page whose demo cannot hold still without a
// pointer is fixed the way overlay and messaging were: it draws the control the
// component hangs off instead. A flag on Page saying "mine needs the clock" would
// be the cheaper answer, and it would also be a way of switching this check off,
// so a page that turns out to need one should earn it here first.
func TestPagesRenderDeterministically(t *testing.T) {
	for _, p := range showcase.Sorted() {
		t.Run(p.Package, func(t *testing.T) {
			draw := func() []uint8 {
				tt := ui.NewTester(func(c *ui.Context) {
					core.Use(c, core.Settings{Mode: core.Light})
					core.WithReducedMotion(c, true)
					p.Paint(c)
				}, p.Width, p.Height)
				return tt.Image().Pix
			}
			first, second := draw(), draw()
			if !bytes.Equal(first, second) {
				t.Errorf("%s drew differently on the second pass (%d of %d bytes differ)",
					p.Package, diffBytes(first, second), len(first))
			}
		})
	}
}

func diffBytes(a, b []uint8) int {
	if len(a) != len(b) {
		return len(a)
	}
	n := 0
	for i := range a {
		if a[i] != b[i] {
			n++
		}
	}
	return n
}
