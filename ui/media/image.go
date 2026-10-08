package media

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// Rect is a region of a picture, in shares of its width and height: 0 to 1 on
// both axes, not pixels.
//
// Shares rather than pixels is the whole reason a crop survives a resize. A
// crop stored in pixels is a crop of one particular rendering: the moment the
// window changes size or the picture is shown at a thumbnail's width, the
// numbers no longer describe the region, and every crop tool that has stored
// pixels has shipped a crop that grew a black band on one side. Shares are
// the only thing that means the same thing at two sizes.
type Rect struct {
	// X and Y are the region's top-left, Y and Width are as shares of the
	// picture's own width and height.
	X, Y, W, H float32
}

// Unit is the whole picture, which is the crop a picture starts as and the one
// every "reset" returns to.
func Unit() Rect { return Rect{W: 1, H: 1} }

// Empty reports whether the region covers nothing, which is a crop that has
// been dragged off its own picture and is not a valid crop to keep.
func (r Rect) Empty() bool { return r.W <= 0 || r.H <= 0 }

// Normalised is the region pulled back inside the picture.
//
// It is called before anything draws or reports, because the only thing that
// knows the picture's bounds is the cropper, and a crop that has been dragged
// past the edge should stop at the edge rather than being kept as a region of
// nothing.
func (r Rect) Normalised() Rect {
	if r.W <= 0 {
		r.W = 1
	}
	if r.H <= 0 {
		r.H = 1
	}
	if r.W > 1 {
		r.W = 1
	}
	if r.H > 1 {
		r.H = 1
	}
	if r.X < 0 {
		r.X = 0
	}
	if r.Y < 0 {
		r.Y = 0
	}
	if r.X > 1-r.W {
		r.X = 1 - r.W
	}
	if r.Y > 1-r.H {
		r.Y = 1 - r.H
	}
	return r
}

// FixedRatio holds a crop to a shape — 16:9 for a video, 4:5 for a profile
// picture — by pulling the region's width back inside its height rather than
// moving either edge.
//
// It moves nothing else because that is what a ratio constraint has to do: a
// reader who has set a height and not a width has asked for that height, and
// an editor that silently changed the height would be overwriting the one
// decision they made. A ratio below zero is no ratio.
func (r Rect) FixedRatio(ratio float32) Rect {
	r = r.Normalised()
	if ratio <= 0 {
		return r
	}
	if w := r.H * ratio; w < r.W {
		// Too wide for the shape: the height is the fixed side, so the width
		// gives way and the region is re-seated on the same centre.
		centre := r.X + r.W/2
		r.W = w
		r.X = centre - w/2
	}
	if h := r.W / ratio; h < r.H {
		centre := r.Y + r.H/2
		r.H = h
		r.Y = centre - h/2
	}
	return r.Normalised()
}

// ratioOf is a region's shape, width over height. Zero for an empty one,
// rather than a division by zero that would be a NaN in a layout.
func ratioOf(r Rect) float32 {
	if r.Empty() {
		return 0
	}
	return r.W / r.H
}

// ── compare ────────────────────────────────────────────────────────────────

// ImageCompareOptions configure an ImageCompare.
type ImageCompareOptions struct {
	// Before and After are the two pictures, the caller's bitmaps.
	Before, After *ui.Bitmap
	// Name is what the comparison is called out loud; it is required, and
	// both halves are named by it so a screen reader reads one picture and
	// then the other rather than one nameless picture.
	Name string
	// At is the divider's position across the picture, 0 to 1. It is the
	// caller's: the divider is also the value a caller saves as "this is where
	// they said the edge was".
	At *float32
	// Ratio is the box's shape; zero takes the whole width and lets the
	// height follow the layout.
	Ratio float32
	// Vertical slides the divider up and down rather than across, which is
	// what a pair of overhead photographs of a building wants.
	Vertical bool
	// Handle draws a grip on the divider. Off by default because a compare
	// embedded in a document is a picture, not a control.
	Handle bool
}

// ImageCompareResult carries an ImageCompare.
type ImageCompareResult struct {
	// Element is the comparison.
	Element *ui.Element
}

// ImageCompare is two pictures of the same thing and a divider between them.
//
// The divider is the whole component. Two pictures side by side are a pair of
// pictures, and only a divider makes them one thing a reader can look across:
// the eye has to be able to move from one to the other at the same place in
// the frame, or the comparison is two images in a row and the reader is being
// asked to do the alignment themselves.
func ImageCompare(c *ui.Context, opts ImageCompareOptions) ImageCompareResult {
	if opts.At == nil {
		panic("media: ImageCompare needs an At to write the divider into; it keeps no " +
			"position of its own")
	}
	if opts.Name == "" {
		panic("media: ImageCompare needs a Name; two pictures and a line say nothing about " +
			"what is being compared")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	at := clamp01(*opts.At)

	e := ui.Box(c).FillWidth().Label(opts.Name).Role(ui.RoleImage).
		Clip().
		Draw(func(p *ui.Painter, r ui.Rect) {
			if r.W <= 0 || r.H <= 0 {
				return
			}
			// Both pictures are drawn into the same box and clipped to their
			// own side of the divider, rather than laid out as two columns:
			// two columns of the same photo do not line up, and every pixel of
			// offset between them is a pixel of difference the reader is being
			// shown as if it were in the photograph.
			before, after := r, r
			if opts.Vertical {
				cut := r.Y + r.H*at
				before.H = cut - r.Y
				after.Y = cut
				after.H = r.Y + r.H - cut
			} else {
				cut := r.X + r.W*at
				before.W = cut - r.X
				after.X = cut
				after.W = r.X + r.W - cut
			}
			p.Clip(before, 0, func() {
				if opts.Before != nil {
					p.Image(opts.Before, before, ui.Cover)
				}
			})
			p.Clip(after, 0, func() {
				if opts.After != nil {
					p.Image(opts.After, after, ui.Cover)
				}
			})
			// The divider is drawn last so it sits over both pictures rather
			// than being clipped by either of them: a line at the seam has to
			// be visible on both sides or it reads as an edge of one of them.
			line := theme.BorderWidth * 2
			if opts.Vertical {
				x := r.X + r.W*at
				p.Line(x, r.Y, x, r.Y+r.H, line, k.Text)
				if opts.Handle {
					internal.Dot(p, x, r.Y+r.H/2, u*2, k.Text)
				}
			} else {
				y := r.Y + r.H*at
				p.Line(r.X, y, r.X+r.W, y, line, k.Text)
				if opts.Handle {
					internal.Dot(p, r.X+r.W/2, y, u*2, k.Text)
				}
			}
		})

	if dx, _, held := e.Dragged(); held {
		box := e.Bounds()
		span := box.W
		if opts.Vertical {
			span = box.H
		}
		if span > 0 {
			next := clamp01(at + dx/span)
			if next != at {
				*opts.At = next
			}
		}
	}

	if opts.Ratio > 0 {
		return ImageCompareResult{
			Element: e.AspectRatio(opts.Ratio),
		}
	}
	_ = u
	return ImageCompareResult{Element: e}
}

// ── crop ───────────────────────────────────────────────────────────────────

// ImageCropperOptions configure an ImageCropper.
type ImageCropperOptions struct {
	// Src is the picture being cropped.
	Src *ui.Bitmap
	// Name is what the picture is called out loud; it is required.
	Name string
	// Rect is the region, the caller's. It is read at the top of the frame
	// and written back at the bottom of the same frame, so the region a caller
	// sees is always the one that was drawn.
	Rect *Rect
	// Ratio holds the region to a shape. Zero is free.
	Ratio float32
	// Shape holds the region to one of the named shapes — "1:1", "4:3",
	// "16:9", "4:5" — which wins over Ratio when both are given. It is a
	// string because the shape is what a reader picks out of a row of
	// buttons, and a float that came from one of those buttons would be a
	// float the reader cannot recognise.
	Shape string
	// Width and Height are the editor's box; zero takes the ratio or the
	// layout's share.
	Width, Height float32
	// Grid draws the rule of thirds, which is the reason a cropper exists
	// rather than four number fields.
	Grid bool
}

// ImageCropperResult carries an ImageCropper.
type ImageCropperResult struct {
	// Element is the editor.
	Element *ui.Element
	// ratio reports the region's current shape as the caller drags it, so a
	// caller can show it or refuse it.
	ratio float32
}

// Ratio is the region's shape as width over height, and false when the region
// is empty. It is reported on the result rather than read back off the rect,
// because a caller that wanted the shape has it and a caller that has to
// compute it will eventually compute it from the wrong rectangle.
func (r ImageCropperResult) Ratio() (float32, bool) {
	if r.ratio <= 0 {
		return 0, false
	}
	return r.ratio, true
}

// ImageCropper is a picture with a region dragged over it.
//
// It writes the region as shares and enforces the ratio itself, in that
// order: the ratio is applied after the drag, so a reader who drags past the
// shape gets the shape held rather than a rectangle they then have to undo.
func ImageCropper(c *ui.Context, opts ImageCropperOptions) ImageCropperResult {
	if opts.Rect == nil {
		panic("media: ImageCropper needs a Rect to write the region into; it keeps no crop " +
			"of its own")
	}
	if opts.Name == "" {
		panic("media: ImageCropper needs a Name; a picture has no words of its own to be read")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	ratio := opts.Ratio
	if named := shapeRatio(opts.Shape); named > 0 {
		ratio = named
	}
	region := opts.Rect.FixedRatio(ratio).Normalised()

	var res ImageCropperResult
	e := ui.Box(c).Label(opts.Name).Role(ui.RoleImage).Clip().
		Draw(func(p *ui.Painter, r ui.Rect) {
			if r.W <= 0 || r.H <= 0 {
				return
			}
			if opts.Src != nil {
				p.Image(opts.Src, r, ui.Cover)
			}
			// Everything outside the region is scrimmed rather than covered by
			// a box: a solid cover would hide the parts of the photograph that
			// tell the reader where the crop is, and the scrim still shows
			// them, just quieter.
			region := ui.Rect{
				X: r.X + r.W*region.X, Y: r.Y + r.H*region.Y,
				W: r.W * region.W, H: r.H * region.H,
			}
			scrim := ui.Hex("#000000").Alpha(0.5)
			p.Fill(ui.Rect{X: r.X, Y: r.Y, W: r.W, H: region.Y - r.Y}, scrim, 0)
			p.Fill(ui.Rect{X: r.X, Y: region.Y + region.H, W: r.W, H: r.Y + r.H - (region.Y + region.H)}, scrim, 0)
			p.Fill(ui.Rect{X: r.X, Y: region.Y, W: region.X - r.X, H: region.H}, scrim, 0)
			p.Fill(ui.Rect{X: region.X + region.W, Y: region.Y, W: r.X + r.W - (region.X + region.W), H: region.H}, scrim, 0)
			p.Stroke(region, k.Accent, theme.SmallRadius, theme.BorderWidth*2)

			if opts.Grid {
				for i := 1; i < 3; i++ {
					x := region.X + region.W*float32(i)/3
					y := region.Y + region.H*float32(i)/3
					p.Line(x, region.Y, x, region.Y+region.H, theme.BorderWidth, k.TextMuted.Alpha(0.6))
					p.Line(region.X, y, region.X+region.W, y, theme.BorderWidth, k.TextMuted.Alpha(0.6))
				}
			}
			// The grips sit at the corners because a corner is the only place
			// a drag means "make this edge move" unambiguously; a handle in
			// the middle of an edge has to guess which edge is meant.
			for _, corner := range [][2]float32{
				{region.X, region.Y}, {region.X + region.W, region.Y},
				{region.X, region.Y + region.H}, {region.X + region.W, region.Y + region.H},
			} {
				internal.Dot(p, corner[0], corner[1], u, k.Accent)
			}
		})

	if dx, dy, held := e.Dragged(); held {
		box := e.Bounds()
		if box.W > 0 && box.H > 0 {
			region = region.Normalised()
			region.X += dx / box.W
			region.Y += dy / box.H
			region = region.FixedRatio(ratio).Normalised()
		}
	}
	*opts.Rect = region
	res.ratio = ratioOf(region)

	if opts.Width > 0 {
		e.Width(opts.Width).Shrink(0)
	}
	if opts.Height > 0 {
		e.Height(opts.Height).Shrink(0)
	} else if ratio > 0 && opts.Width > 0 {
		e.Height(opts.Width / ratio).Shrink(0)
	}
	res.Element = e
	return res
}

// ── annotate ───────────────────────────────────────────────────────────────

// Annotation is one mark on a picture: a pen stroke, a pin, a box around
// something.
type Annotation struct {
	// Kind is what the mark is drawn as. Empty is a pen stroke, which is the
	// one every other is made out of.
	Kind string
	// Points are the mark's own points, as shares of the picture, so a mark
	// drawn on a large preview is the same mark on the exported file.
	Points []struct{ X, Y float32 }
	// Color is what it is drawn in; the zero colour takes the danger tone,
	// because an annotation is a person saying "look here" and red is the
	// tone the interface already uses for that.
	Color ui.Color
	// Text sits beside a pin.
	Text string
	// Author is who drew it, for a picture several people are annotating.
	Author string
}

// AnnotationPoint builds one point of an annotation's path. It is a function
// rather than a literal because a bare `[]struct{X, Y float32}{...}` in a
// caller's source is unreadable, and the anonymous type is what the slice
// above has to be.
func AnnotationPoint(x, y float32) struct{ X, Y float32 } {
	return struct{ X, Y float32 }{X: x, Y: y}
}

// ImageAnnotatorOptions configure an ImageAnnotator.
type ImageAnnotatorOptions struct {
	// Src is the picture being marked.
	Src *ui.Bitmap
	// Name is what the picture is called out loud; it is required.
	Name string
	// Marks are the caller's annotations, in the order they were drawn.
	Marks []Annotation
	// Drawing is the mark being drawn right now. It is the caller's, because
	// only the caller can decide when a stroke has ended and has to become a
	// mark rather than a preview of one.
	Drawing *Annotation
	// DrawingOn says the pointer is down and strokes are being taken. A
	// separate flag rather than "Drawing != nil" because a pointer-down with
	// no movement yet is a stroke of one point, which is a real click.
	DrawingOn bool
	// Width and Height are the picture's box.
	Width, Height float32
	// Tool is what a new mark will be: "pen", "pin" or "box".
	Tool string
}

// ImageAnnotatorResult carries an ImageAnnotator.
type ImageAnnotatorResult struct {
	// Element is the marked picture.
	Element *ui.Element
	// counts is how many marks are on it, which is what the caller's own
	// undo stack needs after a frame has added to them.
	counts int
}

// Count is how many marks are on the picture. It is reported so a caller can
// tell a frame that added one from a frame that drew the same marks again,
// which is every frame after the first.
func (r ImageAnnotatorResult) Count() int { return r.counts }

// ImageAnnotator is a picture with marks on it: pen strokes, pins and boxes.
//
// The marks are the caller's and are written in shares, for the same reason a
// crop is: a mark drawn against one rendering of a picture and stored in
// pixels is in the wrong place on every other rendering. The component draws
// what it is given and strokes into the one mark it was handed, and it keeps
// no list of its own.
func ImageAnnotator(c *ui.Context, opts ImageAnnotatorOptions) ImageAnnotatorResult {
	if opts.Name == "" {
		panic("media: ImageAnnotator needs a Name; a picture has no words of its own to be read")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	tool := opts.Tool
	if tool == "" {
		tool = "pen"
	}
	marks := append([]Annotation(nil), opts.Marks...)
	if opts.Drawing != nil {
		// The stroke in progress is drawn like any other mark and is not
		// counted: it is already in the caller's slice if it is going to be,
		// and counting it twice is how an undo stack grows a duplicate.
		marks = append(marks, *opts.Drawing)
	}

	var res ImageAnnotatorResult
	// The settled marks only. The stroke in progress was appended to `marks`
	// above so it could be drawn, but it is already in the caller's own slice
	// if it is going anywhere, and counting it here is how an undo stack grows
	// a duplicate every frame the pointer is down.
	res.counts = len(opts.Marks)
	e := ui.Box(c).FillWidth().Label(opts.Name).Role(ui.RoleImage).Clip().
		Draw(func(p *ui.Painter, r ui.Rect) {
			if r.W <= 0 || r.H <= 0 {
				return
			}
			if opts.Src != nil {
				p.Image(opts.Src, r, ui.Cover)
			}
			for i, mark := range marks {
				live := i == len(marks)-1 && opts.Drawing != nil
				drawMark(p, r, mark, markInk(mark, k), u, tool, live)
			}
		})
	if opts.Width > 0 {
		e.Width(opts.Width).Shrink(0)
	}
	if opts.Height > 0 {
		e.Height(opts.Height).Shrink(0)
	}
	res.Element = e
	return res
}

// markInk is a mark's colour: its own, or the danger tone. Danger rather than
// the accent because a mark on a picture is a person saying "look here", and
// danger is the tone this palette already reserves for exactly that.
func markInk(mark Annotation, k theme.Tokens) ui.Color {
	if mark.Color.A != 0 {
		return mark.Color
	}
	return k.Danger
}

// drawMark paints one annotation in the picture's own coordinates, since every
// point is a share and the box is the picture's rendered size.
func drawMark(p *ui.Painter, r ui.Rect, mark Annotation, col ui.Color, u float32, tool string, live bool) {
	width := theme.BorderWidth * 2
	switch mark.Kind {
	case "box":
		if len(mark.Points) < 2 {
			return
		}
		x0 := r.X + r.W*mark.Points[0].X
		y0 := r.Y + r.H*mark.Points[0].Y
		p.Stroke(ui.Rect{
			X: x0, Y: y0,
			W: r.W * (mark.Points[1].X - mark.Points[0].X),
			H: r.H * (mark.Points[1].Y - mark.Points[0].Y),
		}, col, theme.SmallRadius, width)
		return

	case "pin":
		if len(mark.Points) == 0 {
			return
		}
		x := r.X + r.W*mark.Points[0].X
		y := r.Y + r.H*mark.Points[0].Y
		// The pin is drawn as a ring, not a dot: a filled dot on a
		// photograph disappears into whatever is dark behind it, and the pin
		// is the one mark that has to be found at a glance.
		internal.Ring(p, x, y, u*1.75, width, col)
		return

	default: // pen
		if len(mark.Points) < 2 {
			if len(mark.Points) == 1 {
				internal.Dot(p, r.X+r.W*mark.Points[0].X, r.Y+r.H*mark.Points[0].Y, u*0.75, col)
			}
			return
		}
		var path ui.Path
		for i, pt := range mark.Points {
			x, y := r.X+r.W*pt.X, r.Y+r.H*pt.Y
			if i == 0 {
				path.MoveTo(x, y)
				continue
			}
			// Round joins, because a pen stroke's corners are rounded and a
			// mitered one on a sharp turn makes a spike out of the corner the
			// reader was trying to point at.
			path.LineTo(x, y)
		}
		// A stroke in progress is drawn thin and the finished one at weight:
		// the reader can see which mark they are still making.
		if live {
			width = theme.BorderWidth
		}
		p.StrokePath(&path, width, col)
	}
}

// shapeRatio is the ratio a named shape means, for a cropper that is offered
// the same list of shapes every other crop tool is. "free" and anything
// unknown are zero, which is what [Rect.FixedRatio] treats as no constraint
// at all — so an unrecognised name degrades to a free crop rather than to a
// rectangle of the wrong shape.
func shapeRatio(name string) float32 {
	switch name {
	case "1:1":
		return 1
	case "4:3":
		return 4.0 / 3.0
	case "16:9":
		return 16.0 / 9.0
	case "4:5":
		return 4.0 / 5.0
	}
	return 0
}
