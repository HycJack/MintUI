package chat

import (
	"math"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/feedback"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// The things a turn can carry that are not prose: a recording, a file, an
// error, an artifact — and the turn itself, which is the one component that
// decides which of those a message holds.

// ── audio ──────────────────────────────────────────────────────────────────

// AudioMessageOptions configure an AudioMessage.
type AudioMessageOptions struct {
	// Title names the recording; empty takes the duration.
	Title string
	// Seconds is how long it is. It is required: a player with no length is
	// a scrubber with nothing to scrub, and "0:00" at the end of a bar is the
	// one thing a reader will look at first.
	Seconds float64
	// Playing says it is playing, which moves the button from a triangle to a
	// square and is the only thing the button's shape is saying.
	Playing bool
	// Wave is the levels to draw beside the scrubber, newest last; nil draws
	// a flat track.
	Wave []float32
	// Label names the message; empty takes the title.
	Label string
	// Transcript is what was said, shown under the player when there is one.
	Transcript string
}

// AudioMessageResult carries an AudioMessage and the two presses on it.
type AudioMessageResult struct {
	// Element is the message.
	Element *ui.Element
	toggled bool
	sought  float64
	// soughtSet reports whether a seek happened at all.
	soughtSet bool
}

// Toggled reports the play button being pressed this frame. The player has no
// clock of its own, so it says the press happened and the caller moves the
// transport.
func (r AudioMessageResult) Toggled() bool { return r.toggled }

// Sought is where in the recording the reader dragged to, and false when they
// did not drag. It is in seconds from the start, because that is what the
// player's own state is in.
func (r AudioMessageResult) Sought() (float64, bool) { return r.sought, r.soughtSet }

// AudioMessage is a recording in a transcript: its length, a way to play it,
// and its waveform so the reader can find the part they want before playing
// thirty seconds of it.
func AudioMessage(c *ui.Context, opts AudioMessageOptions) AudioMessageResult {
	if opts.Seconds <= 0 {
		panic("chat: AudioMessage needs a positive Seconds; a recording with no length " +
			"has nothing to scrub")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	title := opts.Title
	if title == "" {
		title = FormatDuration(opts.Seconds)
	}
	name := opts.Label
	if name == "" {
		name = title
	}

	var res AudioMessageResult
	msg := ui.Column(c).FillWidth().Gap(u).Label(name)
	msg.Children(func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u*2).
			Padding(u*1.25, u*2).Radius(theme.ControlRadius).
			Background(k.Surface).Label(name).Children(func() {
			play := ui.ButtonBase(c).Size(u*7, u*7).Radius(theme.PillRadius).
				Background(k.Fill).TextColor(k.OnFill).Label(playLabel(opts.Playing))
			if play.Clicked() {
				res.toggled = true
			}
			play.Children(func() { playMark(c, k.OnFill, opts.Playing) })

			// The waveform is the scrubber. It is the recording's own shape and
			// the only part of it that says where the reader is, so a separate
			// progress bar under it would be a second thing to read and two
			// things to keep in step.
			//
			// The press is read as a position rather than as a drag, because
			// seeking is a place and not a distance: dragging left from 30s to
			// 25s and dragging right from 30s to 25s are the same seek, and a
			// delta cannot tell them apart.
			wrap := ui.Box(c).Grow(1).Cursor(ui.CursorPointer).Children(func() {
				VoiceWaveform(c, VoiceWaveformOptions{
					Levels: opts.Wave, Bars: max(len(opts.Wave), 24),
					Height: u * 5, Color: k.Accent, Label: name,
				})
			})
			if x, _, pressed := wrap.PointerPosition(); pressed {
				width := wrap.Bounds().W
				if width > 0 {
					res.sought = clampF(float64(x)/float64(width), 0, 1) * opts.Seconds
					res.soughtSet = true
				}
			}

			ui.Text(c, FormatDuration(opts.Seconds)).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
		})
		if opts.Transcript != "" {
			ui.Text(c, opts.Transcript).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.MetaSize)).MaxLines(4)
		}
	})
	res.Element = msg
	return res
}

// playLabel is what the play button is called while it is in one state or the
// other, which is the label the caller reads from the result too.
func playLabel(playing bool) string {
	if playing {
		return core.Def("Pause recording")
	}
	return core.Def("Play recording")
}

// playMark is the triangle or the square in the play button.
func playMark(c *ui.Context, col ui.Color, playing bool) {
	side := core.Density(c).Unit() * 4
	ui.Box(c).Size(side, side).Shrink(0).Role(ui.RoleNone).
		Draw(func(p *ui.Painter, r ui.Rect) {
			if playing {
				// The square is drawn rather than set in the label: it is the one
				// mark on this screen that is not a word, and at the size a button
				// is, a character would be a smudge.
				w := r.W * 0.62
				x := r.X + (r.W-w)/2
				p.Fill(ui.Rect{X: x, Y: r.Y + (r.H-w)/2, W: w, H: w}, col, 1.5)
				return
			}
			var path ui.Path
			path.MoveTo(r.X+r.W*0.32, r.Y+r.H*0.22).
				LineTo(r.X+r.W*0.32, r.Y+r.H*0.78).
				LineTo(r.X+r.W*0.8, r.Y+r.H/2)
			path.Close()
			p.FillPath(&path, col)
		})
}

// FormatDuration writes a length as m:ss, or h:mm:ss when it is long enough to
// need one.
func FormatDuration(seconds float64) string {
	if seconds < 0 {
		seconds = 0
	}
	total := int(seconds + 0.5)
	h, m, sec := total/3600, total/60%60, total%60
	if h > 0 {
		return itoa(h) + ":" + pad2(m) + ":" + pad2(sec)
	}
	return itoa(m) + ":" + pad2(sec)
}

func pad2(n int) string {
	if n < 10 {
		return "0" + itoa(n)
	}
	return itoa(n)
}

func clampF(v, lo, hi float64) float64 {
	switch {
	case v < lo:
		return lo
	case v > hi:
		return hi
	}
	return v
}

// ── file ───────────────────────────────────────────────────────────────────

// FileMessageResult carries a FileMessage and the two presses on it.
type FileMessageResult struct {
	// Element is the message.
	Element *ui.Element
	opened  bool
	// removed is the 1-based index of the chip removed, 0 for none.
	removed int
}

// Opened reports the chip being opened.
func (r FileMessageResult) Opened() bool { return r.opened }

// Removed is the 1-based index of the attachment removed this frame.
func (r FileMessageResult) Removed() int { return r.removed }

// FileMessage is a turn that is a file: a chip with what is known about it
// beside the bubble.
//
// It is a bubble with a chip in it rather than a chip with a bubble around it,
// because the file is what the turn says. Anything else — a title, a note —
// is the caller's body, which goes in the same place it would for prose.
func FileMessage(c *ui.Context, opts FileMessageOptions) FileMessageResult {
	u := core.Density(c).Unit()
	if opts.Attachment.Name == "" {
		panic("chat: FileMessage needs an attachment with a Name")
	}

	var res FileMessageResult
	turn := ui.Row(c).FillWidth().AlignItems(ui.Start).Gap(u * 1.5)
	turn.Children(func() {
		selectionMark(c, opts.Selected, bubbleSide(opts.Role))
		ui.Column(c).Grow(1).AlignItems(bubbleSide(opts.Role)).Gap(u * 0.75).Children(func() {
			if opts.Author != "" || opts.Time != "" {
				MessageHeader(c, MessageHeaderOptions{
					Author: opts.Author, Time: opts.Time, Role: opts.Role,
				})
			}
			bubble(c, bubbleSkin{Role: opts.Role, Label: opts.Attachment.Name}, func() {
				chip := AttachmentChip(c, opts.Attachment)
				if chip.Opened() {
					res.opened = true
				}
				if chip.Removed() {
					res.removed = 1
				}
				if opts.Caption != "" {
					// The caption takes no colour of its own: it is a child of
					// the bubble, so it inherits the ink the bubble chose. A
					// caption drawn in the window's text colour is the bubble's
					// own fill on a filled bubble — the two tokens are the same
					// colour in the dark palette — and a caption the eye cannot
					// find is a caption the transcript never said.
					ui.Text(c, opts.Caption).
						FontSize(core.FontSize(c, theme.MetaSize))
				}
			})
		})
	})
	res.Element = turn
	return res
}

// FileMessageOptions configure a FileMessage.
type FileMessageOptions struct {
	// Role is the side the file sits on; it is the user's in every case, so
	// it is here only so a system turn can be given one.
	Role Role
	// Author and Time head the turn.
	Author, Time string
	// Attachment is the file. Required, with a Name.
	Attachment AttachmentChipOptions
	// Caption is what the turn says about the file, under the chip.
	Caption string
	// Selected puts a mark in the gutter.
	Selected bool
}

// ── errors ─────────────────────────────────────────────────────────────────

// ErrorMessageOptions configure an ErrorMessage.
type ErrorMessageOptions struct {
	// Title says what went wrong. It is required: a red box with a paragraph
	// in it is a paragraph that happens to be red.
	Title string
	// Detail is the part a person would paste into a bug report.
	Detail string
	// Retry names a button that asks again; empty draws none, for a failure
	// that asking again will not fix.
	Retry string
	// Dismiss names a button that puts it away; empty draws none.
	Dismiss string
	// Severity tints it. Zero is Neutral, which takes the danger colour: this
	// is a thing that went wrong, and the zero value should be the safe one.
	Severity core.Severity
}

// ErrorMessageResult carries an ErrorMessage and its two presses.
type ErrorMessageResult struct {
	// Element is the message.
	Element *ui.Element
	retried bool
	dismiss bool
}

// Retried reports the retry button being pressed this frame.
func (r ErrorMessageResult) Retried() bool { return r.retried }

// Dismissed reports the dismiss button being pressed this frame.
func (r ErrorMessageResult) Dismissed() bool { return r.dismiss }

// ErrorMessage is a turn that failed, in the column where the answer would
// have been.
//
// It is drawn in the model's place rather than in a toast or a corner
// banner, because that is where the reader's eye already is and where the gap
// in the conversation is. A toast would say the same thing in a place the
// reader has to look away to, and would be gone before they had looked.
func ErrorMessage(c *ui.Context, opts ErrorMessageOptions) ErrorMessageResult {
	if opts.Title == "" {
		panic("chat: ErrorMessage needs a Title; a red box with a paragraph in it is " +
			"a paragraph that happens to be red")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	sev := opts.Severity
	if sev == core.Neutral {
		sev = core.Danger
	}
	bg, fg := sev.Pair(k)

	var res ErrorMessageResult
	e := ui.Column(c).FillWidth().Radius(theme.SmallRadius).
		Background(bg).Border(theme.BorderWidth, fg.Alpha(0.3)).
		Padding(u*2, u*2.5).Gap(u).Label(opts.Title)
	e.Children(func() {
		ui.Row(c).FillWidth().AlignItems(ui.Start).Gap(u * 1.5).Children(func() {
			display.Icon(c, display.IconWarning, display.IconOptions{
				Name: opts.Title, Tone: sev, Size: u * 4.5,
			})
			ui.Column(c).Grow(1).AlignItems(ui.Start).Gap(u * 0.5).Children(func() {
				ui.Text(c, opts.Title).TextColor(fg).
					FontSize(core.FontSize(c, theme.RowSize)).Bold()
				if opts.Detail != "" {
					// The detail is what the reader copies into a report, so it
					// is selectable text and not a drawn line of dashes: a
					// string a screen reader can read is the whole reason it
					// is text.
					ui.Text(c, opts.Detail).TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.CaptionSize)).MaxLines(6)
				}
			})
		})
		if opts.Retry != "" || opts.Dismiss != "" {
			ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u).Children(func() {
				if opts.Retry != "" {
					if input.Button(c, opts.Retry, input.ButtonOptions{}).Clicked() {
						res.retried = true
					}
				}
				if opts.Dismiss != "" {
					if input.Button(c, opts.Dismiss, input.ButtonOptions{}).Clicked() {
						res.dismiss = true
					}
				}
			})
		}
	})
	res.Element = e
	return res
}

// ── empty state ────────────────────────────────────────────────────────────

// EmptyStateResult carries an EmptyState and the press of its action.
type EmptyStateResult struct {
	// Element is the empty state.
	Element *ui.Element
	pressed bool
}

// Pressed reports the action being pressed this frame.
func (r EmptyStateResult) Pressed() bool { return r.pressed }

// EmptyState is what a chat window shows before there is a conversation: what
// it is for, and the one thing to press to start.
//
// It is feedback.Empty's shape rather than its component: the artwork and the
// action there are for a board's empty column, and a chat window's empty state
// is narrower than that and needs a place for suggestions. What is shared is
// the rule — one card, one action, the copy is the caller's — because a chat
// that opened with a wall of suggested questions is not an empty state, it is
// a menu with nothing in it.
func EmptyState(c *ui.Context, opts EmptyStateOptions) EmptyStateResult {
	if opts.Title == "" {
		panic("chat: EmptyState needs a Title; an empty state that says nothing is a gap")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	var res EmptyStateResult
	card := ui.Column(c).FillWidth().AlignItems(ui.Center).Gap(u*2).
		Padding(u*6, u*4).Radius(theme.CardRadius).
		Background(k.Background).Border(theme.BorderWidth, k.Border).
		Label(opts.Title)
	card.Children(func() {
		// A bright dot where feedback.Empty draws its tiles: this is not a
		// thing that failed to load, so it wants a mark that says "starts
		// here" and not a mark that says "nothing".
		ui.Box(c).Size(u*9, u*9).Radius(u * 4.5).Background(k.Lively)
		ui.Text(c, opts.Title).TextColor(k.Text).
			FontSize(core.FontSize(c, theme.LeadSize)).Bold()
		if opts.Body != "" {
			ui.Text(c, opts.Body).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.MetaSize)).TextAlign(ui.Center)
		}
		if opts.Action != "" {
			if input.Button(c, opts.Action, input.ButtonOptions{
				Primary: true, Label: opts.Action,
			}).Clicked() {
				res.pressed = true
			}
		}
		if len(opts.Suggestions) > 0 {
			ui.Box(c).Width(u * 40).Shrink(0)
			SuggestionChips(c, SuggestionChipsOptions{Suggestions: opts.Suggestions})
		}
	})
	res.Element = card
	return res
}

// EmptyStateOptions configure an EmptyState.
type EmptyStateOptions struct {
	// Title says what this window is for. Required.
	Title string
	// Body is the second line.
	Body string
	// Action names the one button that starts a conversation.
	Action string
	// Suggestions are the questions offered below it, each as "label: what
	// it does". They sit under the action rather than above it, because the
	// action is the thing to press and the suggestions are for someone who
	// has read it and decided not to.
	Suggestions []string
}

// ── feedback ───────────────────────────────────────────────────────────────

// FeedbackFormOptions configure a FeedbackForm.
type FeedbackFormOptions struct {
	// Rating is the caller's score, 1 to 5; zero means not yet rated, which
	// draws the whole row unpressed rather than a zero stars.
	Rating int
	// Comment is the caller's words, edited in place.
	Comment *string
	// CommentLabel names the field; required when Comment is given.
	CommentLabel string
	// Title heads the form; empty draws no heading.
	Title string
	// Submit names the button; empty takes the library's "Send feedback".
	Submit string
	// Dismiss names a button that puts the form away; empty draws none.
	Dismiss string
}

// FeedbackFormResult carries a FeedbackForm and its three presses.
type FeedbackFormResult struct {
	// Element is the form.
	Element   *ui.Element
	starred   int
	submitted bool
	dismiss   bool
}

// Starred is the rating pressed this frame, 0 for none. It is a number rather
// than a bool because the five stars are one control and a caller that had to
// count the presses to find out which one would be reconstructing the value
// from the interface.
func (r FeedbackFormResult) Starred() int { return r.starred }

// Submitted reports the submit button being pressed this frame. The comment is
// the caller's string and the rating is the caller's number, so this is a
// question about whether they want to send them and not a handover of either.
func (r FeedbackFormResult) Submitted() bool { return r.submitted }

// Dismissed reports the dismiss button being pressed this frame.
func (r FeedbackFormResult) Dismissed() bool { return r.dismiss }

// FeedbackForm is the "was that any good" row under an answer.
//
// The rating is five stars drawn as stars rather than a picker: five things to
// press, each labelled with its number, and each the size a star should be at
// a row height. The comment is optional and off by default, because a form
// that asks for prose as well as a star is a form most people close.
func FeedbackForm(c *ui.Context, opts FeedbackFormOptions) FeedbackFormResult {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if opts.Rating < 0 || opts.Rating > 5 {
		panic("chat: FeedbackForm rating is 1 to 5; zero means not yet rated")
	}
	if opts.Comment != nil && opts.CommentLabel == "" {
		panic("chat: FeedbackForm needs a CommentLabel when it takes a comment")
	}

	submit := opts.Submit
	if submit == "" {
		submit = core.Msg(c, "chat.feedback.submit", core.Def("Send feedback"))
	}

	var res FeedbackFormResult
	form := ui.Column(c).FillWidth().Gap(u*1.5).Radius(theme.SmallRadius).
		Background(k.Surface).Border(theme.BorderWidth, k.Border).
		Padding(u*2, u*2.5)
	if opts.Title != "" {
		form.Label(opts.Title)
	}

	form.Children(func() {
		if opts.Title != "" {
			ui.Text(c, opts.Title).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.RowSize))
		}
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u).Children(func() {
			for n := 1; n <= 5; n++ {
				star := n
				// Each star is its own button rather than one control with a
				// number inside it: five stops the keyboard can walk and land
				// on, which a single control with five states does not give.
				btn := ui.ButtonBase(c).Size(u*5, u*5).Shrink(0).
					Radius(theme.SmallRadius).Label(starLabel(n))
				if btn.Clicked() {
					res.starred = star
				}
				btn.Children(func() {
					starMark(c, star <= opts.Rating, k.Warning)
				})
			}
			ui.Box(c).Grow(1)
			if opts.Dismiss != "" {
				if input.Button(c, opts.Dismiss, input.ButtonOptions{}).Clicked() {
					res.dismiss = true
				}
			}
		})
		if opts.Comment != nil {
			input.TextArea(c, opts.Comment, input.TextAreaOptions{
				Label: opts.CommentLabel, Lines: 3,
			})
		}
		if input.Button(c, submit, input.ButtonOptions{Primary: true, Label: submit}).Clicked() {
			res.submitted = true
		}
	})
	res.Element = form
	return res
}

// starLabel is what a star is called out loud: the number, so a screen reader
// says "3" rather than "star" five times.
func starLabel(n int) string { return core.Def("Rate ") + itoa(n) + " of 5" }

// starMark is one filled or empty star.
//
// It is ten points at two radii rather than the ★ character, for the reason
// every mark in this library is drawn: the character is a different size, or a
// different shape, or missing, depending on the font, and a rating that
// changes shape between two machines cannot be compared to itself.
func starMark(c *ui.Context, filled bool, col ui.Color) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	side := u * 4
	const (
		outer float32 = 0.46
		inner float32 = 0.18
	)
	ui.Box(c).Size(side, side).Shrink(0).Role(ui.RoleNone).
		Draw(func(p *ui.Painter, r ui.Rect) {
			cx, cy := r.X+r.W/2, r.Y+r.H/2
			var path ui.Path
			for i := range 10 {
				rad := outer
				if i%2 == 1 {
					rad = inner
				}
				ang := float64(i)/10*2*math.Pi - math.Pi/2
				x := cx + rad*r.W*float32(math.Cos(ang))
				y := cy + rad*r.W*float32(math.Sin(ang))
				if i == 0 {
					path.MoveTo(x, y)
					continue
				}
				path.LineTo(x, y)
			}
			path.Close()
			if filled {
				p.FillPath(&path, col)
				return
			}
			p.StrokePath(&path, 1.3, k.Border)
		})
}

// ── artifact ───────────────────────────────────────────────────────────────

// ArtifactCardOptions configure an ArtifactCard.
type ArtifactCardOptions struct {
	// Title names the artifact; required.
	Title string
	// Kind is what it is — "document", "figure", "spreadsheet". Empty omits
	// the line rather than printing a bare separator.
	Kind string
	// Body is a line about it, drawn as markdown.
	Body string
	// Open names the button that opens it; empty draws none.
	Open string
	// Copy names a button that puts the body on the clipboard; empty draws
	// none, because most artifacts are documents and copying one is not what
	// anybody does with them.
	Copy string
	// Selected is the caller's flag.
	Selected bool
	// Ready says the artifact is finished. It is not busy: an artifact that is
	// still being written is shown by the model as text, not as a card that
	// pretends to be a document and then changes.
	Ready bool
}

// ArtifactCardResult carries an ArtifactCard and its two presses.
type ArtifactCardResult struct {
	// Element is the card.
	Element *ui.Element
	opened  bool
	copied  bool
}

// Opened reports the open button being pressed this frame.
func (r ArtifactCardResult) Opened() bool { return r.opened }

// Copied reports the copy button being pressed this frame.
func (r ArtifactCardResult) Copied() bool { return r.copied }

// ArtifactCard is a thing the model made: a document, a figure, a file —
// sitting in the transcript as a card the reader can take rather than as
// another paragraph describing one.
//
// It is a card rather than a link in the text because an artifact is a whole
// thing with its own name and its own kind, and a link gives it neither. It is
// also the only place in a transcript where a press does not lead to more
// conversation — which is exactly why it is a card: it reads as a different
// kind of object from the text around it.
func ArtifactCard(c *ui.Context, opts ArtifactCardOptions) ArtifactCardResult {
	if opts.Title == "" {
		panic("chat: ArtifactCard needs a Title; an artifact with no name cannot be " +
			"opened or copied")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	var res ArtifactCardResult
	card := ui.Column(c).FillWidth().Radius(theme.CardRadius).
		Background(k.Surface).Border(theme.BorderWidth, k.Border).
		Padding(u*2.5, u*3).Gap(u * 1.5).Label(opts.Title)
	if opts.Selected {
		card = card.BorderColor(k.Accent)
	}

	card.Children(func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.5).Children(func() {
			// The tile is the mark of "a thing" rather than a word for it: an
			// artifact's kind is written by the model and can be anything, and
			// a glyph chosen from the icon set by guessing at the word would
			// be wrong most of the time.
			ui.Box(c).Size(u*6, u*7).Radius(theme.SmallRadius).
				Background(k.Background).Border(theme.BorderWidth, k.Border).
				Shrink(0).Draw(func(p *ui.Painter, r ui.Rect) {
				documentMark(p, r, k.TextMuted)
			})
			ui.Column(c).Grow(1).AlignItems(ui.Start).Children(func() {
				ui.Text(c, opts.Title).TextColor(k.Text).
					FontSize(core.FontSize(c, theme.BodySize)).Bold().SingleLine()
				if opts.Kind != "" {
					ui.Text(c, opts.Kind).TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
				}
			})
			if !opts.Ready {
				feedback.DotsLoader(c, feedback.LoaderOptions{
					Size: u * 3, Color: k.TextMuted, Label: opts.Title,
				})
			}
		})
		if opts.Body != "" {
			MarkdownView(c, MarkdownViewOptions{Source: opts.Body, Tight: true})
		}
		if opts.Open != "" || opts.Copy != "" {
			layout.Divider(c, layout.DividerOptions{})
			ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.5).Children(func() {
				if opts.Open != "" {
					if input.Button(c, opts.Open, input.ButtonOptions{Primary: true}).Clicked() {
						res.opened = true
					}
				}
				if opts.Copy != "" {
					btn := input.Button(c, opts.Copy, input.ButtonOptions{})
					if btn.Clicked() {
						c.WriteClipboard(opts.Body)
						res.copied = true
					}
				}
			})
		}
	})
	res.Element = card
	return res
}

// ── capabilities ───────────────────────────────────────────────────────────

// Capability is one thing this window can do, offered on its empty state.
type Capability struct {
	// Title names it. Required: a card with no name is a shape.
	Title string
	// Body is what it is for.
	Body string
	// Icon names the mark; empty draws none.
	Icon string
}

// CapabilityCardsOptions configure a CapabilityCards.
type CapabilityCardsOptions struct {
	// Title heads the set; empty draws no heading.
	Title string
	// Capabilities are the cards, in the order they are shown.
	Capabilities []Capability
	// Columns is how many across. Zero takes two, which is what three or four
	// cards want beside each other.
	Columns int
	// Max is how many to show; zero shows all of them.
	Max int
}

// CapabilityCardsResult carries a CapabilityCards and what was pressed.
type CapabilityCardsResult struct {
	// Element is the row of cards.
	Element *ui.Element
	chosen  string
}

// Chosen is the capability pressed this frame, empty for none. It is the
// title and not an index, because the caller's list is the one that ordered
// them and the two must not be able to disagree.
func (r CapabilityCardsResult) Chosen() string { return r.chosen }

// CapabilityCards is the row of cards that says what a window is for.
//
// It is cards and not a list of links because these are the three or four
// things a person can do here, and a list of three links reads as a menu they
// have already failed to find what they wanted in. The body of each card is
// one sentence, which is the whole of what a capability is worth saying
// before it has been tried.
func CapabilityCards(c *ui.Context, opts CapabilityCardsOptions) CapabilityCardsResult {
	if len(opts.Capabilities) == 0 {
		panic("chat: CapabilityCards needs at least one capability")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	for i, cap := range opts.Capabilities {
		if cap.Title == "" {
			panic("chat: CapabilityCards card " + internal.Commas(i+1) + " needs a Title")
		}
	}
	shown := opts.Capabilities
	if opts.Max > 0 && len(shown) > opts.Max {
		shown = shown[:opts.Max]
	}
	cols := opts.Columns
	if cols <= 0 {
		cols = 2
	}

	var res CapabilityCardsResult
	set := ui.Column(c).FillWidth().Gap(u * 2)
	if opts.Title != "" {
		set.Label(opts.Title)
	}
	set.Children(func() {
		if opts.Title != "" {
			ui.Text(c, opts.Title).TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize))
		}
		for start := 0; start < len(shown); start += cols {
			chunk := shown[start:min(start+cols, len(shown))]
			ui.Row(c).FillWidth().Gap(u * 2).Children(func() {
				for _, capability := range chunk {
					card := ui.ButtonBase(c).Grow(1).AlignItems(ui.Start).
						Padding(u*2, u*2.5).Radius(theme.CardRadius).
						Background(k.Background).Border(theme.BorderWidth, k.Border).
						Label(capability.Title).TextColor(k.Text)
					if card.Clicked() {
						res.chosen = capability.Title
					}
					card.Children(func() {
						ui.Column(c).FillWidth().AlignItems(ui.Start).Gap(u * 0.75).Children(func() {
							if capability.Icon != "" {
								display.Icon(c, display.IconName(capability.Icon),
									display.IconOptions{
										Name: capability.Icon, Muted: true, Size: u * 4.5,
									})
							}
							ui.Text(c, capability.Title).TextColor(k.Text).
								FontSize(core.FontSize(c, theme.RowSize)).Bold()
							if capability.Body != "" {
								ui.Text(c, capability.Body).TextColor(k.TextMuted).
									FontSize(core.FontSize(c, theme.CaptionSize))
							}
						})
					})
				}
			})
		}
	})
	res.Element = set
	return res
}

// ── one turn ───────────────────────────────────────────────────────────────

// MessageKind is what a turn carries. It decides which component draws the
// body of the bubble; everything around the body is the same for all of them.
type MessageKind int

const (
	// KindText is prose, drawn as markdown.
	KindText MessageKind = iota
	// KindCode is a fenced code block.
	KindCode
	// KindThinking is the model's own reasoning, folded away.
	KindThinking
	// KindAudio is a recording.
	KindAudio
	// KindFile is an attachment.
	KindFile
	// KindError is a failure in the place the answer would have been.
	KindError
	// KindActions is an empty turn that exists only for the actions under it,
	// which is what a retry row is.
	KindActions
)

// ChatMessageOptions configure a ChatMessage.
type ChatMessageOptions struct {
	// Kind is what the turn carries. It is the only thing that says which
	// component draws the body.
	Kind MessageKind
	// Role, Author, Time and Quoted are the turn's, as for a MessageBubble.
	Role     Role
	Author   string
	Time     string
	Quoted   string
	Selected bool
	// Text is the prose of a KindText turn.
	Text string
	// Code and Lang are a KindCode turn.
	Code, Lang string
	// CodeTitle, LineNumbers and MaxCodeHeight dress the code block.
	CodeTitle     string
	LineNumbers   bool
	MaxCodeHeight float32
	// Streaming says the prose is still arriving, which draws the caret.
	Streaming bool
	// Thinking is the KindThinking turn's own options.
	Thinking ThinkingBlockOptions
	// Audio is the KindAudio turn's own options.
	Audio AudioMessageOptions
	// File is the KindFile turn's own options.
	File FileMessageOptions
	// Err is the KindError turn's own options.
	Err ErrorMessageOptions
	// Actions are the buttons under the turn; empty draws none.
	Actions []MessageAction
	// MaxWidth caps the bubble; zero takes the reading width.
	MaxWidth float32
}

// ChatMessageResult carries a ChatMessage and everything the user did to it.
type ChatMessageResult struct {
	// Element is the whole turn.
	Element *ui.Element
	// action is the label of the action pressed this frame.
	action string
	// toggled reports the KindThinking summary being pressed.
	toggled bool
}

// Action is the label of the action pressed this frame, empty for none. It is
// a label rather than an index so that a caller matching on it cannot be
// wrong about which button a rebuilt row produced.
func (r ChatMessageResult) Action() string { return r.action }

// Toggled reports the KindThinking summary being pressed this frame.
func (r ChatMessageResult) Toggled() bool { return r.toggled }

// ChatMessage is one turn of a conversation, whatever it carries.
//
// It exists because a transcript is a list of these and the decision "which
// component draws this one" has to be made in exactly one place. Written out
// per row it would be a switch in every view that renders a message, and two
// views would then disagree about whether a code turn gets a header — which is
// the kind of difference nobody notices until a screenshot of the same
// conversation in two windows does not match.
func ChatMessage(c *ui.Context, opts ChatMessageOptions) ChatMessageResult {
	var res ChatMessageResult

	// KindFile draws its own turn — it is a chip inside a bubble with its own
	// header and its own selection — so it does not go through the body
	// builder below. KindError is inside a bubble but not as the body's shape,
	// for the same reason it draws its own card.
	if opts.Kind == KindFile {
		res.Element = FileMessage(c, FileMessageOptions{
			Role: opts.Role, Author: opts.Author, Time: opts.Time,
			Attachment: opts.File.Attachment, Caption: opts.File.Caption,
			Selected: opts.Selected,
		}).Element
		return res
	}

	bubbleOpts := MessageBubbleOptions{
		Role: opts.Role, Author: opts.Author, Time: opts.Time,
		Quoted: opts.Quoted, Selected: opts.Selected,
		MaxWidth: opts.MaxWidth,
	}

	body := func() {
		switch opts.Kind {
		case KindCode:
			CodeBlock(c, CodeBlockOptions{
				Code: opts.Code, Lang: opts.Lang, Title: opts.CodeTitle,
				LineNumbers: opts.LineNumbers, MaxHeight: opts.MaxCodeHeight,
			})
		case KindThinking:
			res.toggled = ThinkingBlock(c, opts.Thinking).Toggled()
		case KindAudio:
			AudioMessage(c, opts.Audio)
		case KindError:
			ErrorMessage(c, opts.Err)
		case KindActions:
			// Nothing: the turn exists for the row of buttons below, and an
			// empty space where the answer would be is what says so without a
			// word of copy.
		default:
			if opts.Streaming {
				StreamingText(c, &opts.Text, StreamingTextOptions{Streaming: true})
			} else {
				MarkdownView(c, MarkdownViewOptions{Source: opts.Text})
			}
		}
		// The actions are inside the bubble rather than under it, so that a
		// turn is one element and one press target: a row of buttons floating
		// below a bubble would belong to the column and would land on the next
		// turn's gutter as soon as the transcript scrolled by a line.
		if len(opts.Actions) > 0 {
			layout.Divider(c, layout.DividerOptions{})
			res.action = MessageActions(c, opts.Actions).Pressed()
		}
	}

	res.Element = MessageBubble(c, bubbleOpts, body).Element
	return res
}
