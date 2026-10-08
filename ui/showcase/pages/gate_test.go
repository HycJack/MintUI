package pages_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

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
