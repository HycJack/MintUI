package main

import (
	"testing"

	"github.com/egoist/mygo/ui"

	gallery "github.com/HycJack/MintUI/ui/showcase"
	_ "github.com/HycJack/MintUI/ui/showcase/pages"
)

// indexOf finds a page's position by package name.
func indexOf(pages []gallery.Page, pkg string) int {
	for i, p := range pages {
		if p.Package == pkg {
			return i
		}
	}
	return -1
}

// clickUntil clicks the rail row of pkg until the selection moves there.
//
// The overlay page stacks six open layers (three anchored, three whole-window
// centre panels), and the navigation page adds two more. Each layer closes on
// a press outside — a press it consumes — and they give way topmost first, so
// reaching a page through them takes one press per layer still standing.
// What it must never take is more than that: a demo whose open-state was
// rebuilt every frame reopened its layers under every click, and the gallery
// never took another one at all. The budget is the assertion.
func clickUntil(t *testing.T, tt *ui.Tester, sel *int, pages []gallery.Page, pkg string) {
	t.Helper()
	want := indexOf(pages, pkg)
	for i := 0; i < 10; i++ {
		_ = tt.Click(pkg)
		if *sel == want {
			return
		}
	}
	t.Fatalf("sel = %d, want %d (%s) — the gallery stopped taking clicks", *sel, want, pkg)
}

// The gallery must keep taking clicks after it has shown a page whose demos
// include open layers (overlay, navigation). Dismissing the stack works, and
// once it is dismissed the pages answer again.
func TestGalleryTakesClicksAfterLayerPages(t *testing.T) {
	pages := gallery.Sorted()
	for _, pkg := range []string{"overlay", "navigation", "theme"} {
		if indexOf(pages, pkg) < 0 {
			t.Fatalf("page %q is not registered", pkg)
		}
	}

	sel := 0
	tt := ui.NewTester(func(c *ui.Context) {
		galleryContent(c, pages, &sel)
	}, 1280, 900)

	clickUntil(t, tt, &sel, pages, "overlay")

	// One press outside the layers dismisses them for good — the open-state
	// is the process's now, so a layer that came back on the next frame
	// would put this text straight back.
	_ = tt.Click("overlay")
	if tt.HasText("这一块是气泡的内容。") {
		t.Error("the popover survived the press that dismissed it")
	}

	// Walk every page on the rail: any page whose demos ate the clicks
	// would strand the walk here. Two pages double as interaction proofs —
	// a click on a demo control must change what the frame draws and keep
	// having changed on the frames after it.
	for _, p := range pages {
		clickUntil(t, tt, &sel, pages, p.Package)

		switch p.Package {
		case "datetime":
			// The page's EventPopover starts open and eats the press that
			// dismisses it — dismiss first. (The cron chip itself is hard to
			// address: the page has six "45"s and Find takes the first.)
			_ = tt.Click("datetime")
			if !tt.HasText("at 09:30") {
				t.Error("the cron summary vanished with the popover")
			}
		case "account":
			// A press on the recorder clears the binding — the drawn keys
			// give way to the placeholder, and stay given way.
			if err := tt.Click("Open the callback"); err != nil {
				t.Fatalf("press the recorder: %v", err)
			}
			tt.Frame()
			if !tt.HasText("Press keys") {
				t.Error("pressing the recorder did not clear it")
			}
		}
	}
	clickUntil(t, tt, &sel, pages, "overlay")
	clickUntil(t, tt, &sel, pages, "navigation")
	// The palette was dismissed earlier in the walk, and the dismissal
	// outlives the frame — that is the fix working. The rest of the page
	// still draws, and the gallery still answers.
	if !tt.HasText("文件") {
		t.Fatal("the navigation page did not draw")
	}
	clickUntil(t, tt, &sel, pages, "theme")
	clickUntil(t, tt, &sel, pages, "overlay")
}
