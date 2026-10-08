package media

import (
	"strconv"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/overlay"
	"github.com/HycJack/MintUI/ui/theme"
)

// Speeds are the playback rates a control offers, slowest to fastest.
//
// They are a list rather than a range because a rate is not a number a reader
// picks out of a continuum: "half speed" and "twice" are the two almost
// everybody means, and the rates between them are for the people who already
// know what they want.
var Speeds = []float64{0.5, 0.75, 1, 1.25, 1.5, 2}

// clampSpeed is a rate that means something: zero or negative rates are a
// stopped player, not a fast one, and the control is the only thing that
// knows what "not a rate" should be called.
func clampSpeed(v float64) float64 {
	if v <= 0 {
		return 1
	}
	return v
}

// formatSpeed writes a rate the way a player writes it: "1×", "1.5×". The
// trailing multiplication sign is what keeps it from being read as a
// dimension.
func formatSpeed(v float64) string {
	if v == float64(int(v)) {
		return strconv.Itoa(int(v)) + "×"
	}
	return strconv.FormatFloat(v, 'f', -1, 64) + "×"
}

// rateKey is a rate as the string a ButtonGroup matches on. The rate is
// multiplied by a hundred so that 1.25 and 1.2500000001 are the same choice:
// the rounding a caller's own float carries is not a different speed.
func rateKey(v float64) string { return strconv.FormatFloat(v*100, 'f', 0, 64) }

// knob paints the scrubber's thumb: a filled dot with a ring of the window's
// own colour inside it, so it stays visible against both a played rail and a
// track it has not reached.
func knob(p *ui.Painter, x, y, radius float32, col, ring ui.Color) {
	internal.Dot(p, x, y, radius, ring)
	internal.Dot(p, x, y, radius*0.68, col)
}

// ── the scrubber ───────────────────────────────────────────────────────────

// VideoScrubberOptions configure a VideoScrubber.
type VideoScrubberOptions struct {
	// At is the playhead and Duration the length, both the caller's. The
	// scrubber writes At when it is dragged and reads it every frame, which is
	// what lets one position drive the transport, a chapter list and a
	// thumbnail strip at the same time.
	At       *time.Duration
	Duration time.Duration
	// Buffered is how much has loaded, as a share of Duration. Zero draws no
	// buffered band: a player that hides what has arrived makes waiting for a
	// scrub feel like a broken control.
	Buffered float32
	// Markers sit along the track at a share of Duration — the chapters, the
	// adverts, the cue points.
	Markers []float32
	// Label names the scrubber, and is required: a rail and a thumb carry no
	// words of their own.
	Label string
	// Disabled greys it out, for a recording with nothing on it yet.
	Disabled bool
}

// VideoScrubberResult carries a VideoScrubber and where it was dragged to.
type VideoScrubberResult struct {
	// Element is the scrubber.
	Element *ui.Element
	// sought is where the drag landed, as a share of the track.
	sought    float32
	soughtSet bool
}

// Sought is where the reader dragged to, as a share of the track from 0 to 1,
// and false when they did not drag.
//
// A share rather than a duration because the scrubber does not know the track
// is the whole file: a clip, a chapter and an advert are all scrubbed with the
// same control, and only the caller knows which span a rail covers.
func (r VideoScrubberResult) Sought() (float32, bool) { return r.sought, r.soughtSet }

// VideoScrubber is the rail under a video: what has been played, what has
// loaded, where the chapters are, and the thumb you drag.
//
// It moves the playhead by the drag's delta rather than jumping to the
// pointer's absolute position, because MyGo's Dragged reports a delta and
// because a scrubber that jumps to wherever the pointer lands the moment it
// arrives is impossible to land on a frame: the rail moves under the thumb as
// the playhead moves, and the pointer is always behind.
func VideoScrubber(c *ui.Context, opts VideoScrubberOptions) VideoScrubberResult {
	if opts.At == nil {
		panic("media: VideoScrubber needs an At to write the playhead into; it keeps no " +
			"position of its own")
	}
	if opts.Duration <= 0 {
		panic("media: VideoScrubber needs a positive Duration; a scrubber with nothing to " +
			"scrub across is a line")
	}
	if opts.Label == "" {
		panic("media: VideoScrubber needs a Label; a rail and a thumb say nothing about " +
			"what they scrub")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	pos := ScrubRatio(*opts.At, opts.Duration)
	buffered := clamp01(opts.Buffered)
	markers := append([]float32(nil), opts.Markers...)

	var res VideoScrubberResult
	e := ui.Box(c).FillWidth().Height(u * 4).Shrink(0).
		Label(opts.Label).Role(ui.RoleSlider).
		Disabled(opts.Disabled).
		Draw(func(p *ui.Painter, r ui.Rect) {
			if r.W <= 0 {
				return
			}
			mid := r.Y + r.H/2
			thickness := u * 1.25
			p.Fill(ui.Rect{X: r.X, Y: mid - thickness/2, W: r.W, H: thickness},
				k.SurfacePressed, thickness/2)
			if buffered > 0 {
				p.Fill(ui.Rect{X: r.X, Y: mid - thickness/2, W: r.W * buffered, H: thickness},
					k.SurfaceHover, thickness/2)
			}
			p.Fill(ui.Rect{X: r.X, Y: mid - thickness/2, W: r.W * pos, H: thickness},
				k.Accent, thickness/2)
			for _, at := range markers {
				at = clamp01(at)
				x := r.X + r.W*at
				p.Line(x, mid-u*1.5, x, mid+u*1.5, theme.BorderWidth*2, k.TextMuted)
			}
			knob(p, r.X+r.W*pos, mid, u*1.5, k.Accent, k.Background)
		})

	if dx, _, held := e.Dragged(); held && !opts.Disabled {
		box := e.Bounds()
		if box.W > 0 {
			next := clamp01(pos + dx/box.W)
			// Only a real move is a change: Dragged reports a press even when
			// the pointer has not moved, and reporting every frame a scrubber
			// is merely being held would make "did the user scrub" true while
			// they are reaching for it.
			if next != pos {
				res.sought = next
				res.soughtSet = true
				*opts.At = time.Duration(next * float32(opts.Duration))
			}
		}
	}
	res.Element = e
	return res
}

// clamp01 is a share inside [0,1]. Both ends matter here: a drag past the end
// of a rail should pin to the end rather than run the playhead off the file.
func clamp01(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// ── volume ─────────────────────────────────────────────────────────────────

// VolumeControlOptions configure a VolumeControl.
type VolumeControlOptions struct {
	// Volume is the caller's 0 to 1. A drag writes it back; the component
	// keeps no level, so two players in one window cannot disagree.
	Volume *float64
	// Muted is the caller's too, and is separate from Volume because muting is
	// not setting the level to zero. Unmuting has to restore the level, and a
	// control that stored mute as "volume 0" would restore silence.
	//
	// A nil Muted is allowed and means the button sets the level to zero
	// instead — which is all a caller who has no notion of mute can do.
	Muted *bool
	// Label names the control; empty takes the library's "Volume".
	Label string
	// Compact leaves the button alone, for a transport with no room for a
	// rail: one control that mutes, and the level lives in a menu.
	Compact bool
}

// VolumeControlResult carries a VolumeControl.
type VolumeControlResult struct {
	// Element is the control.
	Element *ui.Element
	// changed reports the level having moved under a drag this frame.
	changed bool
}

// Changed reports the level moving.
func (r VolumeControlResult) Changed() bool { return r.changed }

// VolumeControl is the speaker and its level: one button and, unless it is
// compact, a rail beside it.
//
// The button is the mute, not the level. Those are two different actions, and
// conflating them is the bug every player has shipped at some point: dragging
// the speaker to zero should not turn mute on, and a player that cannot
// remember its level across a mute is a player in which the mute button is
// destructive.
func VolumeControl(c *ui.Context, opts VolumeControlOptions) VolumeControlResult {
	if opts.Volume == nil {
		panic("media: VolumeControl needs a Volume to write into; it keeps no level of its own")
	}
	u := core.Density(c).Unit()

	label := opts.Label
	if label == "" {
		label = core.Msg(c, "media.volume", core.Def("Volume"))
	}
	muted := opts.Muted != nil && *opts.Muted
	mark, action := "volume", "Mute"
	switch {
	case muted, *opts.Volume <= 0:
		mark, action = "mute", "Unmute"
	case *opts.Volume < 0.5:
		mark = "volumeDown"
	}

	var res VolumeControlResult
	before := *opts.Volume

	res.Element = ui.Row(c).AlignItems(ui.Center).Gap(u * 1.5).
		Label(label + " control").Children(func() {
		// The button is built in here rather than beside it: an element
		// belongs to whatever container is current where it is made, so one
		// made out here lands in whatever holds this control and the speaker
		// comes out on a line of its own above the rail it belongs beside.
		//
		// It is named for what pressing it does, not for the control it
		// belongs to. "Mute" beside a speaker is the action; "Volume" is the
		// control, and a screen reader that reads both would say the same
		// word twice and never say what the button is for.
		if btn := input.IconButton(c, glyph(mark), action, input.ButtonOptions{}); btn.Clicked() {
			if opts.Muted != nil {
				*opts.Muted = !*opts.Muted
			} else {
				*opts.Volume = 0
			}
		}
		if !opts.Compact {
			input.Slider(c, opts.Volume, input.SliderOptions{
				Min: 0, Max: 1, Label: label, Format: "%d%%",
			}).Width(u * 20).Grow(0)
		}
	})
	// The slider writes the value rather than reporting the press, so a change
	// is observed by comparing against the level the frame began with. Asking
	// the slider instead would count every frame the control was drawn.
	if *opts.Volume != before {
		res.changed = true
	}
	return res
}

// ── speed ──────────────────────────────────────────────────────────────────

// PlaybackSpeedControlOptions configure a PlaybackSpeedControl.
type PlaybackSpeedControlOptions struct {
	// Speed is the caller's rate, and the chosen one is written into it.
	Speed *float64
	// Rates overrides the offered speeds. Empty takes [Speeds], which is the
	// list almost every player offers.
	Rates []float64
	// Open is the *bool the menu opens and closes with, and is required: a
	// menu that closed itself would have to remember whether it was open.
	Open *bool
	// Label names the control for assistive technology.
	Label string
}

// PlaybackSpeedControl is the rate selector: a button showing the current rate
// and a menu of the others.
//
// It is a menu rather than a button that cycles. A button cycling
// 0.5 → 0.75 → 1 → 1.25 … means the reader clicks eleven times to get back to
// 1× and cannot tell what it is set to without already knowing the order,
// which is the case for a list, not for a button.
func PlaybackSpeedControl(c *ui.Context, opts PlaybackSpeedControlOptions) *ui.Element {
	if opts.Speed == nil {
		panic("media: PlaybackSpeedControl needs a Speed to write into; it keeps no rate " +
			"of its own")
	}
	if opts.Open == nil {
		panic("media: PlaybackSpeedControl needs the *bool its menu opens and closes with")
	}
	label := opts.Label
	if label == "" {
		label = core.Msg(c, "media.speed", core.Def("Playback speed"))
	}
	rates := opts.Rates
	if len(rates) == 0 {
		rates = Speeds
	}

	anchor := input.Button(c, formatSpeed(*opts.Speed), input.ButtonOptions{Label: label})
	if anchor.Clicked() {
		*opts.Open = !*opts.Open
	}
	// Built in here rather than handed back, so the anchor and the menu are
	// the same control and a caller cannot anchor the menu to something else
	// by accident: the menu and the rate it shows have to be one thing.
	overlay.Popover(c, anchor, opts.Open, overlay.PopoverOptions{
		Modal: false,
		Title: label,
		Body:  func() { SpeedChoices(c, opts.Speed, rates, label) },
		Label: label,
	})
	return anchor
}

// SpeedChoices is the run of rate buttons a speed menu is made of, writing the
// chosen one into the caller's float.
//
// It is exported so a caller building their own menu gets the same labels and
// the same writing rather than a second set of them.
func SpeedChoices(c *ui.Context, speed *float64, rates []float64, label string) *ui.Element {
	if speed == nil {
		panic("media: SpeedChoices needs a Speed to write into; it keeps no rate of its own")
	}
	if len(rates) == 0 {
		rates = Speeds
	}
	if label == "" {
		label = core.Msg(c, "media.speed", core.Def("Playback speed"))
	}
	buttons := make([]input.GroupButton, 0, len(rates))
	for _, r := range rates {
		v := clampSpeed(r)
		buttons = append(buttons, input.GroupButton{
			Value: rateKey(v), Label: formatSpeed(v),
		})
	}
	// The run is keyed by a local string rather than the float because that is
	// what ButtonGroup matches on; the float is written once the run has
	// settled, so a redraw inside the same frame cannot read a half-chosen
	// value.
	key := rateKey(*speed)
	group := input.ButtonGroup(c, &key, buttons, input.ButtonGroupOptions{
		Label: label, LockChoice: true,
	})
	if key != rateKey(*speed) {
		for _, r := range rates {
			if rateKey(clampSpeed(r)) == key {
				*speed = r
			}
		}
	}
	return group
}

// ── the transport ──────────────────────────────────────────────────────────

// MediaControlsOptions configure a MediaControls.
type MediaControlsOptions struct {
	// At is the playhead and Duration the length, both the caller's.
	At       *time.Duration
	Duration time.Duration
	// Playing is the caller's: the transport has no clock, so it reports the
	// press and the caller moves the transport.
	Playing *bool
	// Volume is the caller's 0 to 1, and Muted its mute; either may be nil for
	// a control that has neither.
	Volume *float64
	Muted  *bool
	// Speed is the caller's rate and Open the *bool its menu uses.
	Speed *float64
	Open  *bool
	// Buffered is how much has loaded, as a share of Duration.
	Buffered float32
	// Markers sit along the scrubber at a share of Duration.
	Markers []float32
	// Labels carried by the scrubber, the volume control and the speed button.
	// Empty takes the library's own words, which a window can reword through
	// core.WithMessages.
	ScrubberLabel string
	VolumeLabel   string
	SpeedLabel    string
	// Compact leaves the volume rail and the speed menu out.
	Compact bool
}

// MediaControlsResult carries a MediaControls and the presses on it.
type MediaControlsResult struct {
	// Element is the transport.
	Element *ui.Element
	// toggled reports the play button being pressed this frame.
	toggled bool
	// stepped is the direction the skip buttons asked for: -1, 1, or 0.
	stepped int
}

// Toggled reports the play button being pressed. The transport has no clock,
// so it says the press happened and the caller decides what playing means —
// which is the only way the same bar can drive a video, an audio file and a
// live stream.
func (r MediaControlsResult) Toggled() bool { return r.toggled }

// Stepped is the direction a skip button was pressed in: -1 back, 1 forward,
// 0 for neither. It is a direction rather than a target time because the
// transport does not know how long a chapter is; the caller does.
func (r MediaControlsResult) Stepped() int { return r.stepped }

// MediaControls is the bar under a player: a scrubber, a play button, a pair
// of skips, a volume control and a rate button.
//
// The play button reports and the caller acts. That is not a limitation of
// this component, it is the reason a player can be a video, a recording and a
// preview of the same thing: each of them has a different idea of what
// "playing" means and none of them is this bar's business.
func MediaControls(c *ui.Context, opts MediaControlsOptions) MediaControlsResult {
	if opts.At == nil {
		panic("media: MediaControls needs an At to write the playhead into; it keeps no " +
			"position of its own")
	}
	if opts.Playing == nil {
		panic("media: MediaControls needs a Playing flag to write; it does not keep a clock " +
			"and cannot tell whether it is playing")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	var res MediaControlsResult

	scrubberLabel := opts.ScrubberLabel
	if scrubberLabel == "" {
		scrubberLabel = core.Msg(c, "media.seek", core.Def("Playback position"))
	}

	res.Element = ui.Column(c).FillWidth().Gap(u * 1.5).Children(func() {
		VideoScrubber(c, VideoScrubberOptions{
			At: opts.At, Duration: opts.Duration,
			Buffered: opts.Buffered, Markers: opts.Markers,
			Label: scrubberLabel,
		})
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.5).Children(func() {
			// The two times sit either side of the buttons rather than in a
			// pair: the playhead is what the reader is looking at and the
			// length is the frame around it, and putting the length last makes
			// the bar read left-to-right the way a tape counter does.
			// The name goes on the wrapper, not on the Text: a Text with a
			// Label is announced by that label instead of by the words it
			// shows, so naming the digits would replace them with "played
			// 0:12" and the number would stop being readable at all.
			ui.Box(c).Label(labelAt("played", *opts.At)).Role(ui.RoleStatus).
				Children(func() {
					ui.Text(c, fmtDuration(*opts.At)).TextColor(k.Text).
						FontSize(core.FontSize(c, theme.CaptionSize))
				})
			ui.Box(c).Grow(1)
			// The three buttons are built in here rather than beside it, for
			// the reason every element in this library is built where it is
			// drawn: one made out here belongs to the column above, and the
			// transport comes out as three buttons stacked down the side of
			// the player with the rest of the bar underneath them.
			if b := input.IconButton(c, glyph("back"), "Back 10 seconds",
				input.ButtonOptions{}); b.Clicked() {
				res.stepped = -1
			}
			if b := input.IconButton(c, transportGlyph(*opts.Playing),
				transportName(*opts.Playing),
				input.ButtonOptions{Primary: true}); b.Clicked() {
				res.toggled = true
				*opts.Playing = !*opts.Playing
			}
			if b := input.IconButton(c, glyph("forward"), "Forward 10 seconds",
				input.ButtonOptions{}); b.Clicked() {
				res.stepped = 1
			}
			ui.Box(c).Grow(1)
			ui.Text(c, TimeToText(opts.Duration)).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize))
			if opts.Volume != nil {
				VolumeControl(c, VolumeControlOptions{
					Volume: opts.Volume, Muted: opts.Muted,
					Label: opts.VolumeLabel, Compact: opts.Compact,
				})
			}
			if opts.Speed != nil && opts.Open != nil {
				PlaybackSpeedControl(c, PlaybackSpeedControlOptions{
					Speed: opts.Speed, Open: opts.Open, Label: opts.SpeedLabel,
				})
			}
		})
	})
	return res
}
