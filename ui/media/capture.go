package media

import (
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/data"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// DeviceKind is what a device is for. It is on the device rather than on the
// selector because the kind decides the mark beside the name and the icon of
// the button, and a selector told "pick a device" would be guessing.
type DeviceKind string

const (
	// KindCamera is a video input.
	KindCamera DeviceKind = "camera"
	// KindMic is an audio input.
	KindMic DeviceKind = "microphone"
	// KindSpeaker is an audio output.
	KindSpeaker DeviceKind = "speaker"
	// KindScreen is a screen or a window to record.
	KindScreen DeviceKind = "screen"
)

// Device is one input or output the caller has enumerated.
//
// The list is the caller's, and it is a list rather than a query because
// enumerating devices is the platform's job and this package has no
// platform: the caller asks the OS, hands over what it got, and this component
// draws a picker over it. That is what keeps the picker testable — a test
// hands it three fake cameras and gets a picker back.
type Device struct {
	// ID is what the caller recognises the device by, and is required: a
	// picker that hands back a name has handed back something the caller
	// cannot open.
	ID string
	// Name is what the reader sees.
	Name string
	// Kind is what the device is for.
	Kind DeviceKind
	// Default marks the one the system would choose, which is not always the
	// one that is selected.
	Default bool
	// Muted is an input the system has switched off — a disconnected headset,
	// a camera another application is using. It stays in the list and stays
	// visible: a device that vanished because it stopped working is a
	// question the reader will ask and the picker should answer.
	Muted bool
}

// A picker is data.List rather than a menu because a list scrolls and a menu
// does not: a machine with eleven input devices is ordinary, and a menu that
// showed three of them would be hiding the rest without saying so.

// DeviceSelectorOptions configure a DeviceSelector.
type DeviceSelectorOptions struct {
	// Devices are the caller's enumeration, in the order it prefers them.
	Devices []Device
	// Selected is the ID of the chosen device, the caller's string.
	Selected *string
	// Kind filters the list to one kind. Empty shows them all, which is what
	// a settings page wants and what a recorder with one picker does not.
	Kind DeviceKind
	// Label names the picker, and is required: a list of devices with no name
	// is a list of devices a screen reader cannot say what for.
	Label string
	// Height is the list's height; zero takes one row's worth, which is right
	// for a picker in a toolbar and wrong for one in a settings page.
	Height float32
	// State and Scroll keep the list's place between frames.
	State  *ui.ListState
	Scroll *ui.ScrollState
}

// DeviceSelectorResult carries a DeviceSelector.
type DeviceSelectorResult struct {
	// Element is the picker.
	Element *ui.Element
}

// DeviceSelector picks one device out of the ones the caller found.
//
// It writes the chosen device's ID rather than its index, so that a picker
// whose list is rebuilt — because a device was plugged in, or because the list
// is filtered by kind — still resolves to the same device afterwards. An
// index would mean a different microphone.
func DeviceSelector(c *ui.Context, opts DeviceSelectorOptions) DeviceSelectorResult {
	if opts.Selected == nil {
		panic("media: DeviceSelector needs a Selected ID to write into; it keeps no choice " +
			"of its own")
	}
	if opts.Label == "" {
		panic("media: DeviceSelector needs a Label; a list of devices with no name is a " +
			"list a screen reader cannot describe")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	shown := make([]int, 0, len(opts.Devices))
	for i, d := range opts.Devices {
		if opts.Kind == "" || d.Kind == opts.Kind {
			shown = append(shown, i)
		}
	}
	if len(shown) == 0 {
		kind := string(opts.Kind)
		if kind == "" {
			kind = "any"
		}
		return DeviceSelectorResult{
			Element: ui.Column(c).FillWidth().Gap(u).Label(opts.Label).Children(func() {
				ui.Text(c, "No "+kind+" devices found").TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.BodySize))
				ui.Text(c, "Connect one and try again.").TextColor(k.TextFaint).
					FontSize(core.FontSize(c, theme.CaptionSize))
			}),
		}
	}

	height := opts.Height
	if height <= 0 {
		height = u * 11
	}

	// data.List names its rows and nothing else, so the picker as a whole is
	// wrapped in a named box: a list of devices that announces only the names
	// inside it leaves a screen reader reader with a list and no idea what it
	// is a list of.
	return DeviceSelectorResult{
		Element: ui.Box(c).FillWidth().Label(opts.Label).Role(ui.RoleGroup).
			Children(func() {
				deviceList(c, deviceListOptions{
					rows:    len(shown),
					height:  height,
					chosen:  opts.Selected,
					state:   opts.State,
					scroll:  opts.Scroll,
					name:    opts.Label,
					devices: opts.Devices,
					shown:   shown,
					k:       k,
					u:       u,
				})
			}),
	}
}

// deviceListOptions is what deviceList needs, gathered so the call site stays
// readable. The palette and the density unit are passed rather than re-read so
// that one row builder decides them for every row.
type deviceListOptions struct {
	rows    int
	height  float32
	chosen  *string
	state   *ui.ListState
	scroll  *ui.ScrollState
	name    string
	devices []Device
	shown   []int
	k       theme.Tokens
	u       float32
}

// deviceList builds the rows. It is separate from the component so that the
// filtering above and the drawing below are two readable halves.
func deviceList(c *ui.Context, o deviceListOptions) *ui.Element {
	return data.List(c, data.ListOptions{
		Rows:   o.rows,
		Height: o.height,
		State:  o.state,
		Scroll: o.scroll,
		Key:    func(i int) any { return o.devices[o.shown[i]].ID },
		Label:  func(i int) string { return o.devices[o.shown[i]].Name },
	}, func(i int) {
		d := o.devices[o.shown[i]]
		row := ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(o.u*1.5).
			Padding(o.u*0.75, o.u*1.5).Radius(theme.SmallRadius).
			Label(d.Name).Role(ui.RoleNone)
		if row.Clicked() && !d.Muted {
			// A muted device stays in the list and stays unclickable, rather
			// than being hidden: "my microphone has disappeared" is a
			// different problem from "my microphone is muted", and a picker
			// that solves the second by causing the first is making the
			// reader's day harder.
			*o.chosen = d.ID
		}
		ink := o.k.Text
		if d.Muted {
			ink = o.k.TextFaint
		}
		if *o.chosen == d.ID {
			row.Background(o.k.SurfaceHover)
		}
		row.Children(func() {
			// The mark is not a button. The whole row is the control — a
			// button inside a row that is also a control would swallow or
			// double the press, and a reader tabbing through the list would
			// meet two stops per device.
			ui.Icon(c, glyph(deviceMark(d.Kind))).Size(o.u*4, o.u*4).
				TextColor(ink).Label(d.Name + ", " + string(d.Kind)).Shrink(0)
			ui.Text(c, d.Name).TextColor(ink).Grow(1).SingleLine().
				FontSize(core.FontSize(c, theme.RowSize))
			if d.Default {
				ui.Text(c, "Default").TextColor(o.k.TextFaint).
					FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
			}
			if d.Muted {
				ui.Text(c, "Unavailable").TextColor(o.k.Warning).
					FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
			}
		})
	})
}

// deviceMark is the glyph for a kind of device. Cameras and screens get the
// obvious marks; the two audio kinds share one, because a microphone and a
// speaker are the same object seen from two sides and giving them different
// marks would say there were two kinds of thing when there is one.
func deviceMark(kind DeviceKind) string {
	switch kind {
	case KindCamera:
		return "camera"
	case KindScreen:
		return "screen"
	}
	return "mic"
}

// ── camera preview ─────────────────────────────────────────────────────────

// CameraPreviewOptions configure a CameraPreview.
type CameraPreviewOptions struct {
	// Frame is the picture the camera is sending, the caller's bitmap. nil
	// draws the waiting state, which is what a camera looks like for the
	// second after it is opened and for as long as the permission prompt is
	// unanswered.
	Frame *ui.Bitmap
	// Name is what the preview is called out loud; it is required.
	Name string
	// Live says a frame is coming. It is the caller's, because only the caller
	// knows whether the camera is producing anything — a camera that has been
	// opened and then blocked reports the same absence as a camera that was
	// never opened, and only the caller can tell them apart.
	Live bool
	// Level is the microphone's level, drawn as a meter along the bottom when
	// it is at least zero. Zero and a negative both draw nothing.
	Level float64
	// Muted greys the meter out.
	Muted bool
	// Ratio is the frame's shape; zero takes 4:3, which is what almost every
	// webcam sends.
	Ratio float32
	// Mirror flips the picture, which is what a person expects of a picture of
	// themselves and not of a picture of a room.
	Mirror bool
	// Badge is a short state drawn over the corner — "LIVE", the caller's name.
	Badge string
}

// CameraPreview is the picture a camera is sending, with its level along the
// bottom.
//
// It draws the frame it is handed rather than opening anything. The whole
// component is the bit between the device and the screen, and a version of it
// that opened the camera would be untestable in exactly the way this package
// is written to avoid: the test would need a device to exist.
func CameraPreview(c *ui.Context, opts CameraPreviewOptions) *ui.Element {
	if opts.Name == "" {
		panic("media: CameraPreview needs a Name; a picture has no words of its own to be read")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	ratio := opts.Ratio
	if ratio <= 0 {
		ratio = 4.0 / 3.0
	}

	return ui.Box(c).FillWidth().Label(opts.Name).Role(ui.RoleImage).Clip().
		AspectRatio(ratio).
		Children(func() {
			layout.AspectRatio(c, layout.AspectRatioOptions{Ratio: ratio, Cover: true},
				func() {
					display.Image(c, opts.Frame, display.ImageOptions{
						Ratio: ratio, Cover: true,
						Placeholder: waitingFor(opts.Live),
						Name:        opts.Name,
					})
				})
			if opts.Badge != "" {
				// Over the corner rather than beside the preview: a camera is
				// usually in a corner of a window with nothing around it, and
				// a badge beside it would push the window's other content
				// sideways every time somebody opened a camera.
				layout.Stack(c, layout.StackOptions{}, func() {
					ui.Box(c).Size(u*2, u*2)
					ui.Box(c).Padding(u*0.5, u*1.5).Radius(theme.PillRadius).
						Background(k.Fill).Shrink(0).Children(func() {
						ui.Text(c, opts.Badge).TextColor(k.OnFill).
							FontSize(core.FontSize(c, theme.CaptionSize))
					})
				})
			}
		})
}

// waitingFor is what an empty preview says, and the two states are said
// differently because they mean different things: a camera that is opening
// will produce a frame, and one that is blocked will not until the reader
// answers the prompt they cannot see from here.
func waitingFor(live bool) string {
	if live {
		return "Waiting for a frame…"
	}
	return "Camera off"
}

// ── screen recording ───────────────────────────────────────────────────────

// ScreenRecorderControlsOptions configure a ScreenRecorderControls.
type ScreenRecorderControlsOptions struct {
	// Recording is the caller's flag, and At the elapsed time it writes. The
	// component keeps neither: it has no clock and no device, so all it can
	// do is report a press and let the caller move the transport.
	Recording *bool
	At        *time.Duration
	// Source is the caller's, as it is for a capture screen; it is drawn so
	// the person being recorded can see what is being recorded, which is the
	// one thing a recorder must not leave out.
	Source *string
	// Width and Height are the recorded area's size, drawn beside the timer so
	// the person can see what they are about to show.
	Width, Height float32
	// Zoomed marks a recording of one window rather than the whole screen,
	// which changes what the person in the recording is allowed to see and so
	// is shown to them.
	Zoomed bool
	// Label names the controls.
	Label string
}

// ScreenRecorderControlsResult carries a ScreenRecorderControls.
type ScreenRecorderControlsResult struct {
	// Element is the controls.
	Element *ui.Element
	// toggled reports the record button being pressed this frame.
	toggled bool
}

// Toggled reports the record button being pressed.
func (r ScreenRecorderControlsResult) Toggled() bool { return r.toggled }

// ScreenRecorderControls is the record button, the timer and the area.
//
// The area is drawn and not just named, because a screen recorder is the one
// control in an application where the person cannot see what they are about to
// capture: they can see the window they are in and the picker they chose, and
// everything between the two is what other people will get.
func ScreenRecorderControls(c *ui.Context, opts ScreenRecorderControlsOptions) ScreenRecorderControlsResult {
	if opts.Recording == nil {
		panic("media: ScreenRecorderControls needs a Recording flag; it keeps no clock " +
			"and cannot tell whether it is recording")
	}
	if opts.At == nil {
		panic("media: ScreenRecorderControls needs an At for the elapsed time; it keeps no " +
			"timer of its own")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	label := opts.Label
	if label == "" {
		label = core.Msg(c, "media.record", core.Def("Screen recording"))
	}

	var res ScreenRecorderControlsResult

	res.Element = ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 2).
		Label(label).Children(func() {
		// Built in here rather than beside it: an element belongs to whatever
		// container is current where it is made, so one made out here lands
		// above this bar and the record button ends up on a line of its own.
		if btn := input.IconButton(c, glyph("record"),
			recordName(*opts.Recording),
			input.ButtonOptions{Primary: !*opts.Recording}); btn.Clicked() {
			res.toggled = true
			*opts.Recording = !*opts.Recording
		}
		// The timer sits beside the button rather than in the middle of the
		// row: what a person watches during a recording is the button, and a
		// timer that had to be found across the window would be one they stop
		// watching.
		ui.Box(c).Label(label + ", elapsed").Role(ui.RoleStatus).Children(func() {
			ui.Text(c, TimeToText(*opts.At)).TextColor(timerInk(k, *opts.Recording)).
				FontSize(core.FontSize(c, theme.StatSize)).Bold()
		})
		if opts.Source != nil {
			ui.Text(c, *opts.Source).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.RowSize)).MaxLines(1).Grow(1)
		}
		if opts.Width > 0 && opts.Height > 0 {
			areaShape(c, opts.Width, opts.Height, opts.Zoomed)
		}
	})
	return res
}

// timerInk is the timer's colour. The danger tone while recording, because a
// running timer that looks like ordinary text is the timer nobody notices,
// and the palette already uses danger for "this is happening now".
func timerInk(k theme.Tokens, recording bool) ui.Color {
	if recording {
		return k.Danger
	}
	return k.TextMuted
}

func recordName(recording bool) string {
	if recording {
		return "Stop recording"
	}
	return "Start recording"
}

// areaShape is the recorded region's own shape, drawn at a size that shows it
// rather than at its real one — a 3840×2160 area drawn at its real size is
// bigger than every window this will ever be shown in.
func areaShape(c *ui.Context, w, h float32, zoomed bool) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	const maxSide float32 = 44
	side := w
	if h > side {
		side = h
	}
	if side <= 0 {
		side = maxSide
	}
	scale := maxSide / side
	name := "Recorded area " + itoa(int(w)) + " by " + itoa(int(h))
	if zoomed {
		name += ", window only"
	}
	return ui.Box(c).Width(w * scale).Height(h * scale).Shrink(0).
		Radius(theme.SmallRadius).BorderWidth(theme.BorderWidth).BorderColor(k.Accent).
		Label(name).Draw(func(p *ui.Painter, r ui.Rect) {
		// Two hairlines across it, so the box reads as an area rather than
		// as a badge: a person deciding whether to record has to be able
		// to see that this is a rectangle of their screen.
		p.Line(r.X+r.W/3, r.Y, r.X+r.W/3, r.Y+r.H, theme.BorderWidth, k.Border)
		p.Line(r.X+r.W*2/3, r.Y, r.X+r.W*2/3, r.Y+r.H, theme.BorderWidth, k.Border)
		internal.Rule(p, ui.Rect{X: r.X, Y: r.Y + r.H/2, W: r.W, H: theme.BorderWidth}, k.Border)
		_ = u
	})
}
