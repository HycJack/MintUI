package display

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// ImageOptions configure an Image.
type ImageOptions struct {
	// Width and Height fix the box. Zero takes it from the bitmap, or from
	// Ratio when that is set, and a box with neither simply sits at the
	// picture's own size.
	Width, Height float32
	// Ratio is width over height, for a slot that must be the same shape
	// whatever loads into it — a job photo, an avatar slot. Zero keeps the
	// picture's own shape.
	Ratio float32
	// Cover crops the picture to fill the box; Contain fits it inside with
	// the rest left as the box's own surface.
	Cover bool
	// Radius rounds the picture's corners; zero leaves them square, which is
	// what an icon tile wants and what a photo in a sheet does not.
	Radius float32
	// Placeholder is drawn while there is nothing to show: a customer's
	// avatar before their photograph has loaded. Empty draws the box alone,
	// which is the right answer for a slot whose emptiness means something.
	Placeholder string
	// Failed is drawn when loading failed, in the danger colour so it is
	// not mistaken for a picture that simply has not arrived yet. Empty
	// falls back to Placeholder.
	Failed string
	// Name is what the image is called out loud. It is required: an image
	// has no text of its own to be read instead.
	Name string
}

// Image is a picture in a box of a known shape, with an honest way of showing
// that there is not one.
//
// The two absences are different and are drawn differently. A placeholder is
// a wait — the customer's photograph is coming, and the box keeps its shape
// meanwhile so nothing around it moves. A failure is a fact, and it wears the
// danger hairline: a slot that looks like a wait for an hour is a support
// ticket nobody filed, because nothing on screen ever said it stopped trying.
//
// src may be nil, which is the placeholder state; a caller that has tried and
// failed says so through Options.Failed rather than by passing something else
// for src.
func Image(c *ui.Context, src *ui.Bitmap, opts ImageOptions) *ui.Element {
	k := core.Tokens(c)
	if opts.Name == "" {
		panic("display: Image needs a Name; a picture has no words of its own to be read")
	}

	// The box comes first and is always the same shape: sizing the frame
	// around whatever arrived is how a list of thumbnails ends up ragged.
	box := ui.Box(c).Radius(opts.Radius).Clip().Background(k.Surface).
		Center().Role(ui.RoleImage).Label(opts.Name)
	if opts.Width > 0 {
		box.Width(opts.Width)
	}
	if opts.Height > 0 {
		box.Height(opts.Height)
	} else if opts.Ratio > 0 && opts.Width > 0 {
		box.Height(opts.Width / opts.Ratio)
	}
	if opts.Ratio > 0 && opts.Height > 0 && opts.Width <= 0 {
		box.Width(opts.Height * opts.Ratio)
	}
	if opts.Ratio > 0 && opts.Width <= 0 && opts.Height <= 0 {
		// A ratio with neither side fixed is the whole of the box's shape: it
		// takes its width from the layout around it and its height from the
		// ratio. Sized from its content instead — which is what a box with no
		// width and no height falls back to — it is a line of placeholder text
		// on a hairline, so a slot that exists to keep a camera preview's
		// shape while nothing has arrived keeps no shape at all.
		box.FillWidth().AspectRatio(opts.Ratio)
	}

	switch {
	case src != nil:
		// The picture is built inside the closure rather than beside it,
		// because an element is parented by whatever container is current
		// where it is created: one made out here would land in the window
		// beside the box instead of inside it, and the box would measure
		// empty. It is not labelled of its own either — the box around it is
		// the image, and a second node with the same name is one more thing
		// for a screen reader to say twice.
		return box.Children(func() {
			img := ui.Image(c, src)
			if opts.Cover {
				img.Fit(ui.Cover)
			}
			if opts.Width > 0 {
				img.FillWidth()
			}
			if opts.Height > 0 {
				img.FillHeight()
			}
		})

	case opts.Failed != "":
		return box.BorderWidth(theme.BorderWidth).BorderColor(k.Danger).
			Children(func() {
				ui.Text(c, opts.Failed).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize))
			})

	case opts.Placeholder != "":
		return box.BorderWidth(theme.BorderWidth).BorderColor(k.Border).
			Children(func() {
				ui.Text(c, opts.Placeholder).TextColor(k.TextFaint).
					FontSize(core.FontSize(c, theme.CaptionSize))
			})

	default:
		// A slot that is only a slot. The tint is the faintest thing in the
		// palette so a row of them reads as the row's own background.
		return box.Background(k.SurfaceHover)
	}
}
