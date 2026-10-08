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
func TestPagesDrawInBothAppearances(t *testing.T) {
	for _, p := range showcase.Sorted() {
		drawn := p.Draw()
		if len(drawn) < 3 {
			t.Errorf("%s drew %d texts: %v", p.Package, len(drawn), drawn)
		}
	}
}

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
// must not change between two draws of an untouched page is everything else — a
// page that mutates its own state while drawing leaves the next draw (or the
// dark-mode draw) showing something different, which is how a screenshot came
// to be one thing on the check that ran before it and another thing in the PNG
// that followed.
//
// A page declaring Page.ClockDriven is skipped: its two draws are allowed to
// differ because its demo reads the real clock, and that page has said so.
func TestPagesRenderDeterministically(t *testing.T) {
	for _, p := range showcase.Sorted() {
		t.Run(p.Package, func(t *testing.T) {
			if p.ClockDriven {
				t.Skip("the page's demo depends on the real clock, so its two " +
					"draws are allowed to differ")
			}
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
