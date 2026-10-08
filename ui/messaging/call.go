package messaging

import (
	"strconv"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// ── the grid ───────────────────────────────────────────────────────────────

// Participant is one person in a call.
type Participant struct {
	// ID identifies them and is required: a call reorders itself as people
	// join, pin and leave, and a tile the caller cannot recognise afterwards
	// is a tile it cannot act on.
	ID string
	// Name is who it is, and is required: a tile with no name is a rectangle.
	Name string
	// Video is the picture the camera is sending, the caller's bitmap. nil is
	// a camera that is off, which is drawn as initials rather than as a black
	// square — a black square reads as a broken camera rather than as a
	// deliberate choice.
	Video *ui.Bitmap
	// Speaking is the caller's flag. It is not derived from a level here: a
	// component that measured the audio would need the audio, and this package
	// has none.
	Speaking bool
	// Pinned lifts the tile out of the automatic order.
	Pinned bool
	// Muted is the caller's flag for whether their microphone is on, which is
	// what the small mark on the tile says.
	Muted bool
	// Hand is the caller's flag for a raised hand. It is drawn in the accent
	// rather than the warning tone, because a raised hand is somebody asking
	// to be let in rather than reporting a problem.
	Hand bool
	// Sharing is the caller's flag for screen sharing, which is what puts a
	// tile in the grid at all on a call where somebody is sharing.
	Sharing bool
}

// VideoCallGridOptions configure a VideoCallGrid.
type VideoCallGridOptions struct {
	// Participants are the caller's, in the order they should be laid out. The
	// grid does not sort them: the rule for who is pinned and who is next is
	// the application's, and a grid that reordered would move a tile under
	// somebody's eyes every time somebody spoke.
	Participants []Participant
	// Speaker is the ID of the person whose voice is dominant, and is drawn
	// with a ring. Empty when nobody is speaking.
	Speaker string
	// Width and Height are the grid's box; zero lets the layout give it one.
	Width, Height float32
	// Columns caps how wide the grid goes; zero takes three, past which tiles
	// are smaller than the face in them.
	Columns int
	// Ratio is a tile's shape; zero takes 4:3, which is what almost every
	// webcam sends.
	Ratio float32
	// Label names the grid for assistive technology; it is required, because a
	// grid of faces is the hardest thing in the interface to describe without
	// one.
	Label string
}

// VideoCallGrid is the call's tiles: one per person, the speaker ringed.
//
// It is drawn rather than assembled out of images and labels because a call
// redraws on every level change — a dozen times a second while anybody is
// talking — and a grid of laid-out tiles is a dozen elements laid out a dozen
// times a second to produce a picture that is one painter call. The names and
// the counts are laid out on top; the picture is painted.
func VideoCallGrid(c *ui.Context, opts VideoCallGridOptions) *ui.Element {
	if opts.Label == "" {
		panic("messaging: VideoCallGrid needs a Label; a grid of faces is the hardest thing " +
			"in this interface to describe without one")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	ratio := opts.Ratio
	if ratio <= 0 {
		ratio = 4.0 / 3.0
	}
	columns := opts.Columns
	if columns <= 0 {
		columns = 3
	}

	people := opts.Participants
	// The tiles are laid out here rather than in a caller's order, because the
	// one thing a grid has to do by itself is stop being rectangular when the
	// count changes: two people in a three-column grid with two empty cells is
	// a grid with holes in it.
	cols := columns
	if n := len(people); n > 0 && n < cols {
		cols = n
	}
	rows := 1
	if cols > 0 {
		rows = (len(people) + cols - 1) / cols
	}

	e := ui.Box(c).Fill().Label(opts.Label).Role(ui.RoleGroup)
	if opts.Width > 0 {
		e.Width(opts.Width).Shrink(0)
	}
	if opts.Height > 0 {
		e.Height(opts.Height).Shrink(0)
	}

	return e.Children(func() {
		for r := 0; r < rows; r++ {
			ui.Row(c).Fill().Grow(1).Gap(u * 1.5).Children(func() {
				for col := 0; col < cols; col++ {
					i := r*cols + col
					if i >= len(people) {
						ui.Box(c).Grow(1).Shrink(0).Role(ui.RoleNone)
						continue
					}
					p := people[i]
					speaking := p.Speaking || (opts.Speaker != "" && opts.Speaker == p.ID)
					// The tile is an element rather than a painter call per
					// person because it carries three laid-out things — the
					// name, the muted mark and the raised hand — and painting
					// text means measuring it against a font this package does
					// not hold.
					tile := ui.Box(c).Grow(1).Grow(1).Shrink(0).
						Radius(theme.CardRadius).Clip().Background(k.Surface).
						Label(p.Name).Role(ui.RoleNone)
					if speaking {
						tile.BorderWidth(theme.BorderWidth * 2).BorderColor(k.Success)
					} else {
						tile.BorderWidth(theme.BorderWidth).BorderColor(k.Border)
					}
					tile.Children(func() {
						layout.AspectRatio(c, layout.AspectRatioOptions{Ratio: ratio, Cover: true},
							func() {
								if p.Video != nil {
									display.Image(c, p.Video, display.ImageOptions{
										Ratio: ratio, Cover: true, Name: p.Name,
									})
									return
								}
								// No camera: initials, large, in the middle of
								// the tile. An empty tile reads as a camera that
								// has failed, and initials read as a person who
								// turned theirs off.
								ui.Box(c).Fill().Center().Children(func() {
									ui.Text(c, initials(p.Name)).TextColor(k.Text).
										FontSize(core.FontSize(c, theme.TitleSize)).Bold()
								})
							})
						// The marks hang off the tile's own bottom-left rather
						// than being laid out under it, so adding one does not
						// change the tile's height and every tile in the row
						// stays the size the grid laid out.
						tileBadge(c, p)
					})
				}
			})
		}
	})
}

// tileBadge is the two marks a tile can carry: a muted microphone and a raised
// hand. They are drawn in the layout rather than in the picture so that they
// sit over the video at the same place whatever shape the video is.
func tileBadge(c *ui.Context, p Participant) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if !p.Muted && !p.Hand {
		return
	}
	layout.Stack(c, layout.StackOptions{}, func() {
		ui.Box(c).Size(u*2, u*2)
		ui.Box(c).Padding(u*0.5, u*1.25).Radius(theme.PillRadius).
			Background(k.Fill).Shrink(0).Label(tileBadgeName(p)).
			Children(func() {
				if p.Hand {
					ui.Text(c, "✋").TextColor(k.OnFill).
						FontSize(core.FontSize(c, theme.CaptionSize))
				} else {
					ui.Icon(c, glyph("micOff")).Size(u*3, u*3).TextColor(k.OnFill)
				}
			})
	})
}

func tileBadgeName(p Participant) string {
	switch {
	case p.Hand:
		return p.Name + " raised a hand"
	case p.Muted:
		return p.Name + " is muted"
	}
	return p.Name
}

// initials is the letters in a name, for the tile of somebody whose camera is
// off. It takes the first letter of the first two words, which is what an
// avatar does everywhere else in this library and is what the reader already
// recognises.
func initials(name string) string {
	out := make([]rune, 0, 2)
	for _, part := range splitWords(name) {
		if len(out) == 2 {
			break
		}
		for _, r := range part {
			out = append(out, upper(r))
			break
		}
	}
	return string(out)
}

func splitWords(s string) []string {
	var out []string
	cur := ""
	for _, r := range s {
		if r == ' ' || r == '-' || r == '_' || r == '@' || r == '.' {
			if cur != "" {
				out = append(out, cur)
				cur = ""
			}
			continue
		}
		cur += string(r)
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}

func upper(r rune) rune {
	if r >= 'a' && r <= 'z' {
		return r - 'a' + 'A'
	}
	return r
}

// ── the controls ───────────────────────────────────────────────────────────

// CallControl is one of the buttons along the bottom of a call.
type CallControl struct {
	// ID is what the caller recognises the press by, and is required: a call
	// has several controls that look alike and a caller that could not tell
	// them apart could not act on any of them.
	ID string
	// Label is what the button says and what it is called out loud. It is
	// required for the same reason display.Icon's is: a glyph is the one
	// element in this interface with no word of its own.
	Label string
	// On is the caller's flag for a toggle in one of its two states, and is
	// what the button's glyph shows. A button with no glyph of its own is a
	// momentary action.
	On bool
	// Danger takes the button out of the ordinary run: hanging up is the one
	// control in a call that ends it, and it must not be the same colour and
	// the same size as the four beside it.
	Danger bool
}

// CallControlsOptions configure a CallControls.
type CallControlsOptions struct {
	// Controls are the caller's, in the order they should be pressed. They are
	// not assembled here: what a call offers depends entirely on what it is —
	// a screen share, a raise hand, a second call — and a bar with a fixed set
	// would be a bar showing buttons the call cannot act on.
	Controls []CallControl
	// Pressed is the ID of the control pressed this frame, or "". It is
	// reported rather than written, because a press is a request to the
	// application and not a state this component can carry out on its own.
	Pressed string
	// Label names the bar for assistive technology; it is required, because a
	// row of glyphs with nothing over it is the hardest bar in the interface
	// to describe.
	Label string
	// Label names the caller: it is drawn in the middle of the bar, which is
	// where a reader looks first when the window is not the call.
	CallerName string
	// Elapsed is how long the call has been going, as the caller's string.
	Elapsed string
}

// CallControlsResult carries a CallControls.
type CallControlsResult struct {
	// Element is the bar.
	Element *ui.Element
	// pressed is the control pressed this frame, or "".
	pressed string
}

// Pressed is the control the reader pressed, by its ID. Accumulated rather
// than assigned: the view runs up to three times per frame and the settled
// pass has no press, so an assignment would overwrite the answer with the
// empty one — which is the mistake §17.2 of the design system describes.
func (r CallControlsResult) Pressed() string { return r.pressed }

// CallControls is the bar under a call: mute, camera, share, leave, and
// whatever else the caller put in it.
//
// Every button reports and nothing is acted on here. Hanging up ends a call,
// sharing a screen sends the reader's whole desktop to other people, and
// muting is the one control a reader reaches for most often and is most often
// a mistake about; a component that carried any of those out from a glyph
// being pressed would be a component that could not be shown a confirmation
// first.
func CallControls(c *ui.Context, opts CallControlsOptions) CallControlsResult {
	if opts.Label == "" {
		panic("messaging: CallControls needs a Label; a row of glyphs with nothing over it is " +
			"the hardest bar in the interface to describe")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	var res CallControlsResult
	if opts.Pressed != "" {
		res.pressed = opts.Pressed
	}

	res.Element = ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u*1.5).
		Padding(u, u*2).Background(k.Surface).
		BorderWidth(theme.BorderWidth).BorderColor(k.Border).
		Label(opts.Label).Children(func() {
		// The call's own facts sit in the middle because that is where a
		// reader looks when the call is not the window they are in: who it
		// is with, and how long it has been going.
		ui.Column(c).Grow(1).Center().Children(func() {
			who := opts.CallerName
			if who == "" {
				who = "Call"
			}
			ui.Text(c, who).TextColor(k.Text).SingleLine().MaxLines(1).
				FontSize(core.FontSize(c, theme.RowSize))
			if opts.Elapsed != "" {
				ui.Text(c, opts.Elapsed).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize))
			}
		})
		for _, ctl := range opts.Controls {
			if ctl.ID == "" || ctl.Label == "" {
				// A control with neither cannot be pressed usefully or
				// announced, and a bar that skips it silently would look
				// deliberate. A caller that built one by mistake finds out.
				panic("messaging: CallControls has a control with no ID or no Label; a " +
					"button nothing can be told apart from is not one")
			}
			callButton(c, ctl, &res.pressed)
		}
	})
	return res
}

// callButton is one control: a glyph that changes with its state, in the
// ordinary tone or in the danger one, always named for what it will do.
func callButton(c *ui.Context, ctl CallControl, pressed *string) {
	k := core.Tokens(c)
	size := core.Density(c).Unit() * 6
	fg := k.Text
	if ctl.On {
		fg = k.Accent
	}
	if ctl.Danger {
		fg = k.Danger
	}
	side := size * 2
	btn := ui.Box(c).Size(side, side).Radius(side / 2).Shrink(0).
		Background(k.SurfaceHover).Center().Cursor(ui.CursorPointer).
		Label(ctl.Label).Role(ui.RoleNone)
	if btn.Clicked() {
		*pressed = ctl.ID
	}
	btn.Children(func() {
		ui.Icon(c, glyph(ctl.glyph())).Size(size, size).TextColor(fg)
	})
}

// glyph is the mark for a control in its current state.
//
// A toggle draws two different marks rather than one mark in two colours. A
// microphone that is muted and a microphone that is not are different facts,
// and colour alone makes them the same fact in two colours — which is exactly
// the thing a reader cannot see.
func (c CallControl) glyph() string {
	switch c.ID {
	case "mute":
		if c.On {
			return "micOff"
		}
		return "mic"
	case "camera":
		if c.On {
			return "cameraOff"
		}
		return "camera"
	}
	return "screen"
}

// ── the controls this package draws for a plain call ───────────────────────

// StandardCallControls is the set of buttons a two-way video call needs, in
// the order they are pressed in. It is exported as a convenience for the
// common case and nothing more: a call with a screen share and a second
// participant builds its own, because the set is the application's decision.
func StandardCallControls(muted, camera, sharing bool) []CallControl {
	return []CallControl{
		{ID: "mute", Label: muteLabel(muted), On: muted},
		{ID: "camera", Label: cameraLabel(camera), On: camera},
		{ID: "share", Label: shareLabel(sharing), On: sharing},
		{ID: "hangup", Label: "Leave the call", Danger: true},
	}
}

func muteLabel(muted bool) string {
	if muted {
		return "Unmute the microphone"
	}
	return "Mute the microphone"
}

func cameraLabel(off bool) string {
	if off {
		return "Turn the camera on"
	}
	return "Turn the camera off"
}

func shareLabel(sharing bool) string {
	if sharing {
		return "Stop sharing"
	}
	return "Share your screen"
}

// participantCount is the summary a call's title bar carries. It is here
// because "3 people" and "Dana, Sam and 1 other" are different sentences, and
// a window that writes both is a window whose header changes width as people
// join.
func participantCount(people []Participant) string {
	switch len(people) {
	case 0:
		return "Nobody here"
	case 1:
		return "1 person"
	case 2:
		return people[0].Name + " and " + people[1].Name
	}
	return people[0].Name + ", " + people[1].Name + " and " +
		strconv.Itoa(len(people)-2) + " others"
}

// ParticipantSummary is what a call's own window says about who is on it, and
// it is exported rather than kept private because that sentence is needed
// wherever a call's title is drawn and writing it twice produces two windows
// whose headers disagree about who is in the room.
func ParticipantSummary(people []Participant) string { return participantCount(people) }
