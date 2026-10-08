package media

import (
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/overlay"
	"github.com/HycJack/MintUI/ui/theme"
)

// closeButton is the way out of a layer this package builds. It is one
// component rather than a call to input.Button at each site, because the label
// is required — a glyph with no name is a control a screen reader cannot
// announce — and because three call sites spelling the same button three ways
// is how a layer ends up with a close icon that does not close.
func closeButton(c *ui.Context, name string) *ui.Element {
	return input.IconButton(c, glyph("closed"), name, input.ButtonOptions{})
}

// ── video ──────────────────────────────────────────────────────────────────

// VideoPlayerOptions configure a VideoPlayer.
type VideoPlayerOptions struct {
	// Picture is the frame on screen. It is a bitmap the caller has already
	// decoded: this package has no decoder, no file and no network, so a
	// player with no picture draws its own surface — which is exactly what a
	// player looks like before anything has loaded into it.
	Picture *ui.Bitmap
	// Poster is the still shown while there is no frame. Two fields rather
	// than one because a player that only knew about the poster would have
	// nowhere to put the frame it exists to show.
	Poster *ui.Bitmap
	// Title heads the player and names it for assistive technology. It is
	// required: a player is a box, and a box says nothing about what it is
	// playing.
	Title string
	// Duration is the length. It is required: "0:00" at the end of a bar is
	// the one thing a reader looks at first, and a transport with no length is
	// a scrubber with nothing to scrub.
	Duration time.Duration
	// At is the playhead and Playing the transport, both the caller's.
	At      *time.Duration
	Playing *bool
	// Volume, Muted, Speed and SpeedOpen are the rest of the transport, all
	// the caller's. Each may be nil for a player that is only ever previewing
	// with the sound off.
	Volume    *float64
	Muted     *bool
	Speed     *float64
	SpeedOpen *bool
	// Buffered is how much has loaded and Markers where the chapters are,
	// both as a share of Duration.
	Buffered float32
	Markers  []float32
	// PictureName is what the frame is called out loud; empty uses Title.
	PictureName string
	// Ratio is the frame's shape, width over height. Zero takes 16:9, which
	// is what almost every video is — and a black box that is not 16:9 is
	// noticed against every frame that is.
	Ratio float32
	// Controls turns the transport off, for a still with nothing to play.
	Controls bool
	// Compact is the smaller transport, for a card rather than a page.
	Compact bool
}

// VideoPlayerResult carries a VideoPlayer.
type VideoPlayerResult struct {
	// Element is the player.
	Element *ui.Element
	// toggled reports the play button being pressed this frame.
	toggled bool
}

// Toggled reports the play button being pressed. The player holds no clock, so
// it says the press happened and the caller decides what playing means —
// which is what lets one component draw a video, a recording and a preview of
// the thing that is being recorded.
func (r VideoPlayerResult) Toggled() bool { return r.toggled }

// VideoPlayer is a picture with a transport under it.
//
// What is missing is the point. There is no decoder, no network, no file: the
// frame on screen is whatever bitmap the caller put in Picture, and the
// transport is a handful of pointers the caller wrote. A player that could
// only be exercised by a real video could not be drawn in a headless test at
// all, and a media library whose components need a device is a media library
// whose bugs are found by users rather than by us.
func VideoPlayer(c *ui.Context, opts VideoPlayerOptions) VideoPlayerResult {
	if opts.Title == "" {
		panic("media: VideoPlayer needs a Title; a player is a box and a box says nothing " +
			"about what it is playing")
	}
	if opts.Duration <= 0 {
		panic("media: VideoPlayer needs a positive Duration; a player with no length " +
			"cannot say where it has got to")
	}
	if opts.At == nil {
		panic("media: VideoPlayer needs an At to write the playhead into; it keeps no " +
			"position of its own")
	}
	if opts.Playing == nil {
		panic("media: VideoPlayer needs a Playing flag; it does not keep a clock and cannot " +
			"tell whether it is playing")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	ratio := opts.Ratio
	if ratio <= 0 {
		ratio = 16.0 / 9.0
	}
	pictureName := opts.PictureName
	if pictureName == "" {
		pictureName = opts.Title
	}
	frame := opts.Picture
	if frame == nil {
		frame = opts.Poster
	}

	var res VideoPlayerResult
	res.Element = layout.Container(c, layout.ContainerOptions{
		Surface: true, Border: true, Radius: theme.CardRadius,
		Gap: u * 1.5,
	}, func() {
		ui.Text(c, opts.Title).TextColor(k.Text).
			FontSize(core.FontSize(c, theme.RowSize)).Bold()
		layout.AspectRatio(c, layout.AspectRatioOptions{Ratio: ratio, Cover: true},
			func() {
				display.Image(c, frame, display.ImageOptions{
					Ratio: ratio, Cover: true,
					Placeholder: "No frame yet",
					Name:        pictureName,
				})
			})
		if opts.Controls {
			transport := MediaControls(c, MediaControlsOptions{
				At: opts.At, Duration: opts.Duration, Playing: opts.Playing,
				Volume: opts.Volume, Muted: opts.Muted,
				Speed: opts.Speed, Open: opts.SpeedOpen,
				Buffered:      opts.Buffered,
				Markers:       opts.Markers,
				ScrubberLabel: "Playback position for " + opts.Title,
				Compact:       opts.Compact,
			})
			if transport.Toggled() {
				res.toggled = true
			}
		}
	})
	return res
}

// ── audio ──────────────────────────────────────────────────────────────────

// AudioPlayerOptions configure an AudioPlayer.
type AudioPlayerOptions struct {
	// Title heads the player; empty takes the length, which is all there is to
	// name it by until a caller gives it a name.
	Title string
	// Duration is the length, and is required for the same reason a video
	// player's is.
	Duration time.Duration
	// At and Playing are the caller's, as everywhere else in a transport.
	At      *time.Duration
	Playing *bool
	// Samples is the audio the waveform is drawn from, as raw linear
	// amplitudes. nil draws the flat resting track, which is what a player
	// looks like before anything has been decoded into it.
	Samples []float64
	// Bars is how many heights the waveform is reduced to; zero lets it fill
	// the width it is given.
	Bars int
	// Volume, Muted, Speed and SpeedOpen are the rest of the transport.
	Volume    *float64
	Muted     *bool
	Speed     *float64
	SpeedOpen *bool
	// Compact is the smaller transport.
	Compact bool
}

// AudioPlayerResult carries an AudioPlayer.
type AudioPlayerResult struct {
	// Element is the player.
	Element *ui.Element
	// toggled reports the play button being pressed this frame.
	toggled bool
	// sought is where the reader dragged the waveform to, as a share.
	sought    float32
	soughtSet bool
}

// Toggled reports the play button being pressed.
func (r AudioPlayerResult) Toggled() bool { return r.toggled }

// Sought is where the waveform was dragged to, as a share of it, and false
// when it was not dragged.
//
// The playhead is written too: the waveform is the seek control here, the way
// the rail is in a video, and a caller that had to add its own drag to make
// seeking work would add a second one and get two playheads.
func (r AudioPlayerResult) Sought() (float32, bool) { return r.sought, r.soughtSet }

// AudioPlayer is a waveform with a transport under it.
//
// It is deliberately not chat.AudioMessage. That component is a recording
// inside a transcript — a bubble's worth of padding, a title, a transcript
// underneath — while an audio player is a full-width control carrying the rest
// of a transport's states. One box shared between them would be wrong for one
// of the two, which is the usual outcome when two things look alike.
func AudioPlayer(c *ui.Context, opts AudioPlayerOptions) AudioPlayerResult {
	if opts.Duration <= 0 {
		panic("media: AudioPlayer needs a positive Duration; a player with no length " +
			"cannot say where it has got to")
	}
	if opts.At == nil {
		panic("media: AudioPlayer needs an At to write the playhead into; it keeps no " +
			"position of its own")
	}
	if opts.Playing == nil {
		panic("media: AudioPlayer needs a Playing flag; it does not keep a clock and cannot " +
			"tell whether it is playing")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	title := opts.Title
	if title == "" {
		title = "Audio " + TimeToText(opts.Duration)
	}

	pos := ScrubRatio(*opts.At, opts.Duration)
	transport := MediaControlsOptions{
		At: opts.At, Duration: opts.Duration, Playing: opts.Playing,
		Volume: opts.Volume, Muted: opts.Muted,
		Speed: opts.Speed, Open: opts.SpeedOpen,
		ScrubberLabel: "Playback position for " + title,
		Compact:       opts.Compact,
	}

	var res AudioPlayerResult
	res.Element = layout.Container(c, layout.ContainerOptions{
		Surface: true, Border: true, Radius: theme.CardRadius,
		Pad: u * 2, Gap: u * 2,
	}, func() {
		ui.Text(c, title).TextColor(k.Text).
			FontSize(core.FontSize(c, theme.RowSize)).Bold()

		// The waveform is built in here rather than beside it: an element
		// belongs to whatever container is current where it is made, and one
		// made outside this closure would land in the window and leave the
		// panel measuring itself around two children instead of three.
		wave := AudioWaveform(c, AudioWaveformOptions{
			Samples: opts.Samples, Bars: opts.Bars, Position: pos,
			Label: "Waveform of " + title,
		})
		// The drag moves the playhead by the pointer's delta rather than
		// jumping to wherever the pointer landed, for the reason the scrubber
		// does it: the waveform's played band follows the playhead, so an
		// absolute jump always leaves the thumb behind the cursor.
		if dx, _, held := wave.Dragged(); held {
			box := wave.Bounds()
			if box.W > 0 {
				next := clamp01(pos + dx/box.W)
				if next != pos {
					res.sought, res.soughtSet = next, true
					*opts.At = time.Duration(next * float32(opts.Duration))
				}
			}
		}

		if MediaControls(c, transport).Toggled() {
			res.toggled = true
		}
	})
	return res
}

// ── the lightbox ───────────────────────────────────────────────────────────

// LightboxOptions configure a Lightbox.
type LightboxOptions struct {
	// Open is the *bool the layer opens and closes with, and is required for
	// the same reason it is on every layer in this library: a layer that
	// closed itself would have to remember whether it was open.
	Open *bool
	// Src is the picture, the caller's bitmap. nil is allowed and draws the
	// waiting state, which is what a picture whose file has not arrived yet
	// should look like.
	Src *ui.Bitmap
	// Name is what the picture is called out loud. It is required: a picture
	// enlarged is still a picture, and still has no words of its own.
	Name string
	// Caption sits under the picture.
	Caption string
	// Width and Height cap the picture; zero lets the dialog size itself to
	// the window, which is what a portrait should do and a panorama should not.
	Width, Height float32
}

// Lightbox is a picture at the size it should be looked at, over the window.
//
// It is overlay.Dialog rather than something of its own, and the whole
// difference is what is inside it: a picture over the window is the same
// relationship a dialog has with it — a panel, a scrim, Escape and a press on
// the backdrop both dismissing — and a second implementation of that
// relationship would be a second set of rules about what closes a layer.
func Lightbox(c *ui.Context, opts LightboxOptions) *ui.Element {
	if opts.Open == nil {
		panic("media: Lightbox needs the *bool it opens and closes with")
	}
	if opts.Name == "" {
		panic("media: Lightbox needs a Name; a picture has no words of its own to be read")
	}
	u := core.Density(c).Unit()
	k := core.Tokens(c)

	return overlay.Dialog(c, opts.Open, overlay.DialogOptions{
		Rule:     true,
		MaxWidth: 960,
		Body: func() {
			display.Image(c, opts.Src, display.ImageOptions{
				Width: opts.Width, Height: opts.Height,
				Cover: true, Placeholder: "No picture yet",
				Name: opts.Name,
			})
			if opts.Caption != "" {
				ui.Text(c, opts.Caption).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize)).TextAlign(ui.Center)
			}
		},
		Actions: func() {
			ui.Box(c).Grow(1)
			closeButton(c, "Close picture")
		},
	}).Margin(u*0, 0, 0, 0)
}
