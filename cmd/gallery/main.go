// Command gallery is the library's own component gallery.
//
//	go run ./cmd/gallery              # a real window: a sidebar of packages,
//	                                  # a page on the right
//	go run ./cmd/gallery -shots out   # headless PNG of every page, no window
//	go run ./cmd/gallery -check       # every page draws what it promised
//
// The -shots mode is the one that matters for review: it writes a PNG per
// package so the components can be looked at rather than inferred from a
// passing test.
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	gallery "github.com/HycJack/MintUI/ui/showcase"
	_ "github.com/HycJack/MintUI/ui/showcase/pages"
)

func main() {
	shots := flag.String("shots", "", "write a PNG of every page into this directory and exit")
	only := flag.String("only", "", "restrict to these packages, comma separated")
	dark := flag.Bool("dark", false, "with -shots, also render every page in the dark palette")
	scale := flag.Int("scale", 2, "pixel ratio for -shots")
	check := flag.Bool("check", false, "verify every page draws what it promised, then exit")
	fitFlag := flag.Bool("fit", false, "print the height each page needs and exit")
	flag.Parse()

	pages := gallery.Sorted()
	if *only != "" {
		want := map[string]bool{}
		for _, p := range strings.Split(*only, ",") {
			want[strings.TrimSpace(p)] = true
		}
		var kept []gallery.Page
		for _, p := range pages {
			if want[p.Package] {
				kept = append(kept, p)
			}
		}
		pages = kept
	}
	if len(pages) == 0 {
		fmt.Fprintln(os.Stderr, "no pages")
		os.Exit(1)
	}

	switch {
	case *check:
		bad := 0
		for _, p := range pages {
			miss := p.Missing(p.Draw())
			if len(miss) > 0 {
				fmt.Printf("✗ %-12s did not draw %v\n", p.Package, miss)
				bad++
				continue
			}
			if !p.Anchored && BottomFlush(render(p, false)) {
				// Not a failure: the page drew everything it promised. But
				// its last row is not empty, so whatever is at the end of it
				// is cut off, and a PNG of that reads as a page that ended
				// there on purpose.
				fmt.Printf("✓ %-12s %d promised texts, but taller than its own %d — "+
					"the bottom is cut\n", p.Package, len(p.Want), p.Height)
				bad++
				continue
			}
			fmt.Printf("✓ %-12s %d promised texts\n", p.Package, len(p.Want))
		}
		if bad > 0 {
			os.Exit(1)
		}
	case *fitFlag:
		// A page says how tall it is, and that number goes stale the moment
		// anyone adds a section. This prints the height each page actually
		// needs, so the number in the source can be copied straight over
		// instead of guessed at and found wrong later.
		for _, p := range pages {
			fmt.Printf("%-12s %d\n", p.Package, fit(p))
		}
	case *shots != "":
		shotsTo(*shots, pages, *dark, *scale)
	default:
		window(pages)
	}
}

func shotsTo(dir string, pages []gallery.Page, dark bool, scale int) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	bad := 0
	for _, p := range pages {
		for _, d := range []bool{false, true} {
			if d && !dark {
				continue
			}
			name := p.Package
			if d {
				name += "-dark"
			}
			img := render(p, d)
			if scale > 1 {
				img = scaleUp(img, scale)
			}
			f, err := os.Create(filepath.Join(dir, name+".png"))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			if err := png.Encode(f, img); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			f.Close()
			fmt.Printf("✓ %-16s %v\n", name, filepath.Join(dir, name+".png"))
		}
		if miss := p.Missing(p.Draw()); len(miss) > 0 {
			fmt.Printf("  ⚠ did not draw %v\n", miss)
			bad++
		}
	}
	if bad > 0 {
		fmt.Fprintf(os.Stderr, "%d pages did not draw what they promised\n", bad)
		os.Exit(1)
	}
}

func render(p gallery.Page, dark bool) *image.RGBA {
	mode := core.Light
	if dark {
		mode = core.Dark
	}
	return ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: mode})
		p.Paint(c)
	}, p.Width, p.Height).Image()
}

// fit finds the height a page needs, starting from the height it claims.
//
// It searches upwards rather than measuring, for two reasons. A page's own
// declared height is nearly right, so a search from there is a handful of
// renders instead of a dozen. And a page cannot simply be measured from its
// pixels: a drawer hangs off the window's right edge and a selection bar off
// its bottom, so those two sit wherever the page was laid out, and a page
// measured tall and then cut back to its content leaves them floating in the
// middle of nowhere. Laying it out at the right height is the only way they
// land where they belong.
func fit(p gallery.Page) int {
	if p.Anchored {
		return p.Height
	}
	h := p.Height
	if h < 240 {
		h = 240
	}
	for n := 0; n < 400 && h < 20000; n++ {
		if !BottomFlush(renderAt(p, h)) {
			return h
		}
		h += 60
	}
	return h
}

func renderAt(p gallery.Page, h int) *image.RGBA {
	mode := core.Light
	return ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: mode})
		p.Paint(c)
	}, p.Width, h).Image()
}

// BottomFlush reports whether a page has something painted near its own last
// row of pixels, which means the page is taller than it says and whatever is
// at the end of it has been cut in half.
//
// It looks at a few rows rather than one, because one row lands in the gap
// between two sections often enough to matter: a page cut at exactly the
// whitespace above its last heading has an empty bottom row and looks fine.
//
// The height a page declares cannot be measured instead, because some
// components are anchored to the window rather than to the content: a drawer
// hangs off the right edge, a selection bar off the bottom. Those land
// wherever the page is laid out, so a page that was measured and then trimmed
// to its content would leave them floating in the middle of nowhere. The
// declared height is the only place they can land correctly — so the page has
// to keep it up to date, and this is what says when it has not.
func BottomFlush(img *image.RGBA) bool {
	b := img.Bounds()
	bg := img.RGBAAt(b.Min.X, b.Min.Y)
	for _, up := range []int{1, 6, 16, 34, 60} {
		y := b.Max.Y - up
		if y < b.Min.Y {
			continue
		}
		for x := b.Min.X; x < b.Max.X; x++ {
			if !near(img.RGBAAt(x, y), bg) {
				return true
			}
		}
	}
	return false
}

func near(got, want color.RGBA) bool {
	dr := int(got.R) - int(want.R)
	dg := int(got.G) - int(want.G)
	db := int(got.B) - int(want.B)
	da := int(got.A) - int(want.A)
	return dr*dr+dg*dg+db*db+da*da < 900
}

func scaleUp(src *image.RGBA, s int) *image.RGBA {
	w, h := src.Bounds().Dx()*s, src.Bounds().Dy()*s
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dst.Set(x, y, src.At(x/s, y/s))
		}
	}
	return dst
}

// window opens the gallery for real. It is the same pages the screenshots
// come from, so what a person scrolls through is what was rendered.
func window(pages []gallery.Page) {
	sel := 0
	mygo.App.WhenReady(func() {
		mygo.NewWindow(mygo.WindowOptions{
			Title:  "MintUI · 组件画廊",
			Width:  1280,
			Height: 900,
			Content: ui.View(func(c *ui.Context) {
				galleryContent(c, pages, &sel)
			}),
		})
	})
	if err := mygo.App.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// galleryContent is the gallery window's own interface: a rail of packages on
// the left and the chosen page on the right. The rail is the app's own
// sidebar, not a component the gallery is showing — a page list is the one
// thing here that has to work before anything else can.
func galleryContent(c *ui.Context, pages []gallery.Page, sel *int) {
	core.Use(c, core.Settings{})
	k, u := core.Tokens(c), core.Density(c).Unit()

	ui.Row(c).Fill().Children(func() {
		ui.Column(c).Width(240).FillHeight().Padding(u * 3).Gap(u).
			Background(k.Surface).Children(func() {
			ui.Text(c, "组件画廊").TextColor(k.Text).
				FontSize(core.FontSize(c, 20)).FontWeight(700).MarginY(u)
			for i, p := range pages {
				label := p.Package
				fg := k.TextMuted
				if i == *sel {
					fg = k.Text
				}
				row := ui.Box(c).FillWidth().Padding(u*1.5, u*2).
					Radius(14).Label(label).Children(func() {
					ui.Text(c, label).TextColor(fg).
						FontSize(core.FontSize(c, 14))
				})
				if i == *sel {
					row.Background(k.Background)
				}
				if row.Clicked() {
					*sel = i
					c.Invalidate()
				}
			}
		})

		p := pages[*sel]
		ui.Column(c).Grow(1).FillHeight().Padding(u * 5).Gap(u * 3).
			Children(func() {
				ui.Column(c).FillWidth().Gap(u * 0.5).Children(func() {
					ui.Text(c, p.Title).TextColor(k.Text).
						FontSize(core.FontSize(c, 28)).FontWeight(700)
					if p.Note != "" {
						ui.Text(c, p.Note).TextColor(k.TextMuted).
							FontSize(core.FontSize(c, 14))
					}
				})
				ui.Box(c).Grow(1).FillHeight().Children(func() {
					ui.Scroll(c).Children(func() { p.Paint(c) })
				})
			})
	})
}
