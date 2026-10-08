package data

import (
	"strconv"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// zoomSteps are the magnifications an ImageViewer steps through: as the
// picture fits, then as much again, and so on.
//
// They are numbers and not a slider because the pictures that need a viewer
// are the ones whose size nobody can tell by looking at them: a screenshot
// of a form is legible at any size and a photograph of a receipt is legible
// at none, and the number is what tells the two apart.
var zoomSteps = []float32{1, 1.5, 2, 3}

// ImageViewerOptions configure an ImageViewer.
type ImageViewerOptions struct {
	// Images are the pictures, in order; the first is what it opens on.
	Images []ui.ImageSource
	// At is the picture on show, in the caller's state.
	At *int
	// Zoom is how far the picture is magnified, in the caller's state: 1 is
	// as large as it fits, and a magnified picture is panned with the
	// scrollbars. Nil fits the picture and draws no zoom controls, for the
	// one picture nobody is going to enlarge.
	Zoom *float32
	// Height is the height of the window the picture shows in, and the width
	// is what the parent gives. Height is required: a picture in a window
	// with no height is a picture with no window.
	Height float32
	// Caption is said under the picture.
	Caption string
	// Label names each picture, for the strip and for assistive technology.
	Label func(image int) string
	// Strip draws the pictures along the bottom, for a viewer with more than
	// two: without it there is no way to jump to the fifth.
	Strip bool
	// NoArrows draws no arrows over the picture, for a caller paging the
	// pictures itself.
	NoArrows bool
	// Loop pages from the last picture back to the first.
	Loop bool
	// Scroll is where a magnified picture is panned to, when the caller
	// keeps one: a picture enlarged once should still be where it was when
	// the viewer is shown again.
	Scroll *ui.ScrollState
}

// ImageViewerResult carries an ImageViewer and what the user did with it.
type ImageViewerResult struct {
	// Element is the whole viewer.
	Element *ui.Element
	// zoomed reports a press of a zoom control this frame.
	zoomed bool
	// paged is how far the user paged this frame, and pagedBy says whether
	// they did.
	paged   int
	pagedBy bool
}

// Zoomed reports a press of a zoom control this frame.
func (r ImageViewerResult) Zoomed() bool { return r.zoomed }

// Paged returns how far the user paged this frame, and whether they did:
// -1 for the picture before, 1 for the one after, and a jump into the strip
// as the distance to the picture jumped to.
func (r ImageViewerResult) Paged() (int, bool) { return r.paged, r.pagedBy }

// ImageViewer is one picture at a time, with the controls that go to
// another: the arrows over the picture, the strip underneath, and the zoom
// that magnifies the picture and lets it be panned.
//
// The picture on show and how far it is magnified are the caller's, so the
// viewer opens on the picture the caller last looked at, at the size the
// caller last left it, and a test can put it anywhere without pressing
// anything.
func ImageViewer(c *ui.Context, opts ImageViewerOptions) ImageViewerResult {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if opts.At == nil {
		panic("data: ImageViewer needs the At index to point at")
	}
	if len(opts.Images) == 0 {
		panic("data: ImageViewer needs at least one picture")
	}
	if *opts.At < 0 || *opts.At >= len(opts.Images) {
		panic("data: ImageViewer is showing picture " + strconv.Itoa(*opts.At) +
			", which is not one of its pictures")
	}
	if opts.Height <= 0 {
		panic("data: ImageViewer needs a Height; without one it cannot show a picture")
	}

	var res ImageViewerResult
	at, n := *opts.At, len(opts.Images)
	step := func(d int) {
		to := at + d
		switch {
		case to < 0:
			if !opts.Loop {
				return
			}
			to = n - 1
		case to > n-1:
			if !opts.Loop {
				return
			}
			to = 0
		}
		*opts.At = to
		res.paged, res.pagedBy = d, true
	}

	res.Element = ui.Column(c).FillWidth().Shrink(0).Gap(u).Children(func() {
		// The picture in a window that scrolls both ways, because a picture
		// magnified past the window is panned with the scrollbars — which is
		// what a magnified picture in every other program does, and the only
		// way to reach its far side without a mode of its own.
		window := layout.ScrollArea(c, layout.ScrollAreaOptions{
			Vertical: true, Horizontal: true, Height: opts.Height, State: opts.Scroll,
		}, nil).Element.FillWidth().Background(k.Surface).
			Radius(theme.ControlRadius).Clip()

		// The window's own width is the one it had in the frame before, and
		// there is none on the frame the viewer appears on: that frame shows
		// the picture as it is and asks for the one that measures the window.
		win := window.Bounds().W
		if win <= 0 {
			c.Invalidate()
		}
		zoom := float32(1)
		if opts.Zoom != nil {
			zoom = max(*opts.Zoom, 1)
		}
		window.Children(func() {
			// The picture's box is the window at the zoom asked for, and the
			// picture fits inside it: so at one the picture is as large as
			// the window allows, and past one the box is larger than the
			// window and the window scrolls it. Its width is said outright,
			// because a percentage of a scroll window is the width of what
			// fits in it rather than the width of the window.
			frame := ui.Box(c).Height(opts.Height).Shrink(0)
			if win > 0 {
				frame.Width(win * zoom).Height(opts.Height * zoom)
			}
			frame.Children(func() {
				ui.Image(c, opts.Images[at]).Fit(ui.Contain).Fill()
			})
		})
		if !opts.NoArrows {
			// The arrows are over the picture and outside its scroll, so
			// paging does not move them and panning does not take them away.
			// Beside the picture they would narrow it, and a picture that
			// changed width as it was paged would read as two pictures.
			window.Children(func() {
				over(c, iconBack, core.Msg(c, "data.prevPicture", "Previous picture"),
					false, opts.Height, at == 0 && !opts.Loop, func() { step(-1) })
			})
			window.Children(func() {
				over(c, iconNext, core.Msg(c, "data.nextPicture", "Next picture"),
					true, opts.Height, at == n-1 && !opts.Loop, func() { step(1) })
			})
		}

		if opts.Caption != "" || opts.Zoom != nil {
			ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u*2).
				Padding(0, u, u, u).Children(func() {
				if opts.Caption != "" {
					ui.Text(c, opts.Caption).TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.MetaSize)).Grow(1).SingleLine()
				} else {
					ui.Box(c).Grow(1)
				}
				if opts.Zoom != nil {
					zoomControls(c, opts, &res)
				}
			})
		}
		if opts.Strip && n > 2 {
			ui.Row(c).FillWidth().Gap(u).Children(func() {
				for i := range opts.Images {
					thumb := ui.Box(c).Grow(1).Height(u * 11).Shrink(0).
						Radius(theme.ControlRadius).Role(ui.RoleNone).
						Label(imageName(c, opts, i, n)).Clip()
					if i == at {
						// The picture on show wears the accent, which is the
						// one thing the strip has to say at a glance.
						thumb.Border(bar, k.Accent)
					} else {
						thumb.Border(theme.BorderWidth, k.Border)
					}
					thumb.Children(func() {
						ui.Image(c, opts.Images[i]).Fit(ui.Cover).Fill()
					})
					if thumb.Clicked() {
						*opts.At = i
						res.paged, res.pagedBy = i-at, true
					}
				}
			})
		}
	})
	return res
}

// over is one of the arrows, laid over the picture at one side of it: a box
// as tall as the window, with the arrow in the middle of it, so that the
// arrow is in the middle of the picture rather than in the middle of the
// top of it.
func over(c *ui.Context, glyph *ui.SVG, name string, right bool, height float32, off bool, step func()) {
	u := core.Density(c).Unit()
	side := ui.Box(c).Absolute().Top(0).Height(height).Justify(ui.Center).Shrink(0)
	if right {
		side.Right(u * 2)
	} else {
		side.Left(u * 2)
	}
	side.Children(func() {
		if input.IconButton(c, glyph, name, input.ButtonOptions{Disabled: off}).Clicked() {
			step()
		}
	})
}

// zoomControls is the pair of controls that magnify the picture and put it
// back, and the number that says how far.
func zoomControls(c *ui.Context, opts ImageViewerOptions, res *ImageViewerResult) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	// Where between the steps the caller's number falls, so that a zoom the
	// caller set itself is stepped from and not snapped away.
	at := 0
	for i, z := range zoomSteps {
		if z <= *opts.Zoom {
			at = i
		}
	}
	if input.IconButton(c, iconMinus, core.Msg(c, "data.zoomOut", "Zoom out"),
		input.ButtonOptions{Disabled: at == 0}).Clicked() {
		*opts.Zoom = zoomSteps[max(at-1, 0)]
		res.zoomed = true
	}
	ui.Box(c).Padding(u*0.5, u*1.5).Radius(theme.PillRadius).
		Background(k.Surface).Label(zoomLabel(*opts.Zoom)).Children(func() {
		ui.Text(c, zoomLabel(*opts.Zoom)).TextColor(k.TextMuted).
			FontSize(core.FontSize(c, theme.CaptionSize))
	})
	if input.IconButton(c, iconPlus, core.Msg(c, "data.zoomIn", "Zoom in"),
		input.ButtonOptions{Disabled: at == len(zoomSteps)-1}).Clicked() {
		*opts.Zoom = zoomSteps[min(at+1, len(zoomSteps)-1)]
		res.zoomed = true
	}
}

// zoomLabel is the number a magnified picture is measured in.
func zoomLabel(z float32) string {
	return strconv.Itoa(int(z*100)) + "%"
}

// imageName is what a picture is called: the caller's own name for it, and
// how far along the viewer it is, so that a picture read out says which one
// it is rather than only what it is of.
func imageName(c *ui.Context, opts ImageViewerOptions, i, n int) string {
	at := strconv.Itoa(i+1) + " of " + strconv.Itoa(n)
	if opts.Label != nil {
		if name := opts.Label(i); name != "" {
			return name + ", " + at
		}
	}
	return core.Msg(c, "data.pictureAt", "Picture ") + at
}
