package main

import (
	"image"
	"image/color"
	"io"
	"os"
	"path/filepath"
	"strings"
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

// TestShotsToChecksBeforeItRenders pins the order the -shots run works in.
//
// A page is a program, not a picture: run it twice and the second run is not
// the first one again. The overlay page's hover card starts open and closes
// itself on the first frame that has no pointer over it, so its subtitle is
// drawn once and never again. Asking a page what it promised after it has
// already been rendered asks it for something it has given away, which fails
// the run on a tree where nothing is broken.
func TestShotsToChecksBeforeItRenders(t *testing.T) {
	dir := t.TempDir()
	shown := false
	// A page that only shows its promised text on the first draw: the
	// hover card, reduced to the one thing it is actually asserting.
	flip := gallery.Page{
		Package: "flip", Width: 200, Height: 120,
		Want: []string{"promised"},
		Render: func(c *ui.Context) {
			if !shown {
				shown = true
				ui.Text(c, "promised")
			}
		},
	}
	// A nil error is the whole assertion: the check got the first draw, so
	// "promised" was still there when it asked. Were the order reversed the
	// render would have spent it and the run would come back with an error.
	if err := shotsTo(dir, []gallery.Page{flip}, false, 1); err != nil {
		t.Fatalf("shotsTo: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "flip.png")); err != nil {
		t.Errorf("no PNG was written: %v", err)
	}
}

// TestShotsToRefusesToRenderWhatDidNotDraw is the other half of the same rule:
// a page that fails its promise stops the run, so the folder holds a gallery
// rather than a gallery plus a page that came out blank.
//
// It also has to say which page. The check runs before anything is rendered, so
// a failing run writes no PNGs at all and there is no file to match a warning
// against — a run that cannot name its failure is a run nobody can act on.
func TestShotsToRefusesToRenderWhatDidNotDraw(t *testing.T) {
	dir := t.TempDir()
	// Two failures, so a report that named only the first cannot pass. The
	// promised text is the same on both pages, which is the point: a warning
	// naming only the text cannot tell them apart. The names share no
	// substring either, so a run that named only one of them cannot pass by
	// naming the other.
	liar := func(pkg string) gallery.Page {
		return gallery.Page{
			Package: pkg, Width: 200, Height: 120,
			Want:   []string{"promised"},
			Render: func(c *ui.Context) {},
		}
	}
	pages := []gallery.Page{liar("alpha"), liar("beta")}

	var err error
	out := captureStdout(t, func() { err = shotsTo(dir, pages, false, 1) })
	if err == nil {
		t.Fatal("shotsTo wrote pages that promised text they never drew")
	}
	// Every failing package, on the warning and in what the caller prints to
	// stderr: stdout may be redirected, and a caller reading only the error
	// has to be able to act on it too.
	for _, p := range pages {
		if !strings.Contains(err.Error(), p.Package) {
			t.Errorf("the error does not name %q: %v", p.Package, err)
		}
		if !strings.Contains(out, p.Package) {
			t.Errorf("the warning does not name %q: %q", p.Package, out)
		}
	}
	entries, readErr := os.ReadDir(dir)
	if readErr != nil {
		t.Fatalf("read the shots dir: %v", readErr)
	}
	if len(entries) != 0 {
		t.Errorf("a failing run wrote %d PNGs; the folder reads as a gallery", len(entries))
	}
}

// captureStdout runs f with os.Stdout replaced by a pipe and returns what was
// printed to it. The gates report through stdout, so a test that says nothing
// about what they printed is testing half of what they promise.
func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	saved := os.Stdout
	os.Stdout = w
	// Restored twice on purpose: once here, so the pipe is closed before
	// anything reads it, and once on the way out in case f panics and the
	// rest of the run inherits a closed stdout.
	defer func() { os.Stdout = saved }()
	done := make(chan string, 1)
	go func() {
		b, _ := io.ReadAll(r)
		done <- string(b)
	}()
	f()
	os.Stdout = saved
	w.Close()
	out := <-done
	r.Close()
	return out
}

// TestCheckGatesRejectsAPageThatPromisesNothing is the floor -check enforces.
// A page whose Want is empty checks itself perfectly: nothing was promised,
// so nothing can be missing, and the gate went green on a blank page. Three
// is the same floor the pages gate test uses, kept in step deliberately.
//
// Every page here draws exactly what it promises, so the floor is the only
// thing that can fail any of them.
func TestCheckGatesRejectsAPageThatPromisesNothing(t *testing.T) {
	promising := func(want []string) gallery.Page {
		return gallery.Page{
			Width: 200, Height: 120, Want: want,
			Render: func(c *ui.Context) {
				for _, w := range want {
					ui.Text(c, w)
				}
			},
		}
	}

	blank := promising(nil)
	blank.Package = "blank"
	if bad := checkGates([]gallery.Page{blank}); bad != 1 {
		t.Errorf("a page promising nothing: %d pages failed, want 1", bad)
	}

	// One short of the floor fails too, so three is the floor and not merely
	// "empty is not allowed".
	short := promising([]string{"a", "b"})
	short.Package = "short"
	if bad := checkGates([]gallery.Page{short}); bad != 1 {
		t.Errorf("a page promising 2 texts: %d pages failed, want 1", bad)
	}

	atFloor := promising([]string{"one", "two", "three"})
	atFloor.Package = "floor"
	if bad := checkGates([]gallery.Page{atFloor}); bad != 0 {
		t.Errorf("a page at the floor: %d pages failed, want 0", bad)
	}
}

// TestGutterFlush is the Page.Gutter guardrail on a synthetic frame rather than
// on a gallery page. Thirteen of the pages in this tree fail it for real, so
// what is pinned here is the scan that finds them and not the pages it finds.
//
// Each case starts from a blank page and adds exactly one section, because the
// two answers have to be told apart. A scan that looked anywhere but the two
// margins would pass the escaping sections and fail the one that respects the
// gutter; a scan that flagged the whole image would do the reverse.
func TestGutterFlush(t *testing.T) {
	const W, H = 200, 100
	blank := func() *image.RGBA {
		img := image.NewRGBA(image.Rect(0, 0, W, H))
		for y := 0; y < H; y++ {
			for x := 0; x < W; x++ {
				img.SetRGBA(x, y, color.RGBA{255, 255, 255, 255})
			}
		}
		return img
	}
	// A section well clear of both margins: where a page's own content is
	// supposed to end up.
	section := func(img *image.RGBA, x0, x1 int) {
		for y := 40; y < 60; y++ {
			for x := x0; x < x1; x++ {
				img.SetRGBA(x, y, color.RGBA{20, 20, 20, 255})
			}
		}
	}

	t.Run("a section that stops short of the margin", func(t *testing.T) {
		img := blank()
		section(img, gallery.Gutter+12, W-gallery.Gutter-12)
		if GutterFlush(img) {
			t.Error("a section inset past the gutter was reported as flush against the edge")
		}
	})
	// The two edges are checked apart on purpose. They are two bands in one
	// loop, and a second band at the wrong offset or the wrong width passes
	// every page in the gallery as long as the pages are flush on the left.
	t.Run("a section in the left margin", func(t *testing.T) {
		img := blank()
		section(img, 4, gallery.Gutter-4)
		if !GutterFlush(img) {
			t.Error("a section reaching x=4 was not reported as flush against the edge")
		}
	})
	t.Run("a section in the right margin", func(t *testing.T) {
		img := blank()
		section(img, W-gallery.Gutter+4, W-4)
		if !GutterFlush(img) {
			t.Error("a section reaching the last 4 columns was not reported as flush against the edge")
		}
	})
}
