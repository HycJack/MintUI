package media

import (
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/theme"
)

// Frame is one moment of a video, as a decoded still.
//
// It is a bitmap the caller produced, because producing one means decoding a
// frame and this package does not decode. The pair of it — the picture and
// where in the timeline it sits — is what a strip or a filmstrip needs, and
// keeping them together means a caller cannot line a strip of pictures up
// with the wrong moments, which is the mistake a strip of bare images invites.
type Frame struct {
	// At is where in the timeline this frame sits.
	At time.Duration
	// Bitmap is the still. nil draws the frame's own surface, so a strip whose
	// frames have not decoded yet still shows the marks and the times.
	Bitmap *ui.Bitmap
	// Label is what the frame is called out loud; empty takes the time.
	Label string
}

// ── video thumbnails ───────────────────────────────────────────────────────

// VideoThumbnailStripOptions configure a VideoThumbnailStrip.
type VideoThumbnailStripOptions struct {
	// Frames are the stills, in timeline order.
	Frames []Frame
	// Selected is the frame the strip marks as the current one, as an index
	// into Frames. It is the caller's: two views of one video must agree
	// about which frame is current, and neither of them owns the timeline.
	Selected *int
	// Duration is the whole timeline's length, used to say how far along each
	// frame is. Zero puts the frames at even shares, which is what a strip of
	// extracted stills is.
	Duration time.Duration
	// Height is the strip's height; zero takes the frame's width times nine
	// sixteenths, which is the shape of the video it came from.
	Height float32
	// Width is one frame's width; zero shares what the strip is given.
	Width float32
	// Label names the strip, and is required: a row of pictures says nothing
	// about what it is a row of.
	Label string
}

// VideoThumbnailStripResult carries a VideoThumbnailStrip.
type VideoThumbnailStripResult struct {
	// Element is the strip.
	Element *ui.Element
	// picked is the frame that was pressed this frame, or -1.
	picked int
}

// Picked is the frame the reader pressed, as an index into the frames it was
// given, or -1 when none was. It is an index rather than a time because the
// strip does not own the timeline and only the caller can turn a position
// into the seek it means.
func (r VideoThumbnailStripResult) Picked() int { return r.picked }

// VideoThumbnailStrip is the filmstrip under a scrubber: the video's own
// moments, so that a reader can find the part they want before playing it.
//
// A press reports the frame; it does not seek. The strip's job is to say
// which picture was pressed, and a component that also moved the playhead
// would be seeking from a list that may well be showing a different video's
// frames than the rail above it is scrubbing.
func VideoThumbnailStrip(c *ui.Context, opts VideoThumbnailStripOptions) VideoThumbnailStripResult {
	if len(opts.Frames) == 0 {
		panic("media: VideoThumbnailStrip needs at least one Frame; a strip of nothing is " +
			"an empty box")
	}
	if opts.Selected == nil {
		panic("media: VideoThumbnailStrip needs a Selected to mark the current frame with; " +
			"it owns no timeline")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	height := opts.Height
	if height <= 0 {
		width := opts.Width
		if width <= 0 {
			width = u * 14
		}
		height = width * 9.0 / 16.0
	}
	chosen := *opts.Selected

	var res VideoThumbnailStripResult
	res.picked = -1
	res.Element = ui.Row(c).FillWidth().Gap(u * 0.75).Label(opts.Label).
		Role(ui.RoleList).Children(func() {
		for i, f := range opts.Frames {
			frame := f
			name := frame.Label
			if name == "" {
				name = TimeToText(frame.At)
			}
			cell := ui.Box(c).Grow(1).Height(height).Shrink(0).
				Radius(theme.SmallRadius).Clip().
				Label("Frame at " + name).Role(ui.RoleListItem)
			if cell.Clicked() {
				res.picked = i
			}
			// The chosen frame keeps a ring and the others do not, rather than
			// the chosen one being drawn differently: a border that appears and
			// disappears moves every picture beside it by a pixel, which is
			// what a strip of thumbnails must not do.
			if i == chosen {
				cell.BorderWidth(theme.BorderWidth * 2).BorderColor(k.Accent)
			}
			cell.Children(func() {
				display.Image(c, frame.Bitmap, display.ImageOptions{
					Width: 0, Height: height, Cover: true,
					Placeholder: TimeToText(frame.At), Name: name,
				})
			})
		}
	})
	return res
}

// ── image thumbnails ───────────────────────────────────────────────────────

// ImageThumbnailOptions configure an ImageThumbnail.
type ImageThumbnailOptions struct {
	// Src is the picture, the caller's bitmap.
	Src *ui.Bitmap
	// Name is what the thumbnail is called out loud. It is required for the
	// same reason Image's is: a picture has no words of its own.
	Name string
	// Side is the thumbnail's edge; zero takes six density units.
	Side float32
	// Selected marks it with a ring, and is the caller's flag rather than the
	// thumbnail's state.
	Selected bool
	// Ratio is the shape, width over height; zero is square, which is what a
	// gallery grid wants and what a video still does not.
	Ratio float32
	// Badge counts something about the picture — the length of the video it is
	// a frame of, the number of annotations on it.
	Badge string
	// Tone colours the badge; Neutral is the quiet default.
	Tone core.Severity
	// Pressable makes the thumbnail take presses. A thumbnail in a gallery is
	// a link and a thumbnail on an asset row is a picture, and they are the
	// same shape.
	Pressable bool
}

// ImageThumbnailResult carries an ImageThumbnail.
type ImageThumbnailResult struct {
	// Element is the thumbnail.
	Element *ui.Element
	// clicked reports the thumbnail being pressed this frame.
	clicked bool
}

// Clicked reports the thumbnail being pressed.
func (r ImageThumbnailResult) Clicked() bool { return r.clicked }

// ImageThumbnail is one picture in a grid, a strip or a file list.
//
// It is display.Image with three things a thumbnail needs and an image does
// not: a fixed shape, a selection ring, and somewhere to put the badge that
// says how long a video frame is or how many annotations a picture carries.
func ImageThumbnail(c *ui.Context, opts ImageThumbnailOptions) ImageThumbnailResult {
	if opts.Name == "" {
		panic("media: ImageThumbnail needs a Name; a picture has no words of its own to be " +
			"read")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	side := opts.Side
	if side <= 0 {
		side = u * 6
	}
	ratio := opts.Ratio
	if ratio <= 0 {
		ratio = 1
	}

	var res ImageThumbnailResult
	tile := ui.Box(c).Width(side).Height(side / ratio).Shrink(0).
		Radius(theme.SmallRadius).Clip().Label(opts.Name).Role(ui.RoleNone)
	if opts.Pressable {
		tile.Cursor(ui.CursorPointer)
		if tile.Clicked() {
			res.clicked = true
		}
	}
	tile.Children(func() {
		display.Image(c, opts.Src, display.ImageOptions{
			Width: side, Height: side / ratio, Cover: true,
			Placeholder: "—", Name: opts.Name,
		})
		if opts.Badge != "" {
			// The badge hangs off the corner rather than sitting under the
			// picture: a thumbnail's height is what the grid rows are sized
			// from, and a badge that changed that height would make the grid
			// ragged the first time one picture had a duration.
			display.Badge(c, 1, display.BadgeOptions{
				Name: opts.Badge, Tone: opts.Tone,
			}).Attach(ui.AnchorBottomRight, ui.AnchorTopRight)
		}
	})
	if opts.Selected {
		tile.BorderWidth(theme.BorderWidth * 2).BorderColor(k.Accent)
	}
	res.Element = tile
	return res
}
