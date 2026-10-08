package chat

import (
	"math"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/feedback"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/theme"
)

// Work in progress: an answer arriving, the model's own thinking, and the two
// buttons that start and stop it.

// StreamingTextOptions configure a StreamingText.
type StreamingTextOptions struct {
	// Streaming says an answer is still arriving, which is what draws the
	// caret. It is the caller's flag because only the caller knows whether
	// the transport is still open.
	Streaming bool
	// Caret is what is drawn at the end while streaming: a block for a model
	// that writes tokens, a bar for one that streams characters. Empty takes
	// a block.
	Caret string
	// Lines is how many lines of the text to show before it scrolls. Zero
	// takes the composer's full height, which is where an answer that is
	// being written is read.
	Lines int
	// Label names the text; empty leaves it unnamed, since the words say what
	// it is.
	Label string
}

// StreamingText is text that is still being written, with a caret at the end.
//
// The caret is the whole reason this is a component rather than a Text: a
// block of words with no mark at the end is indistinguishable from an answer
// that has stopped, and a reader who cannot tell the two will wait for the
// rest of a sentence that is not coming.
func StreamingText(c *ui.Context, text *string, opts StreamingTextOptions) *ui.Element {
	if text == nil {
		panic("chat: StreamingText needs the text to point at; it keeps none of its own")
	}
	u := core.Density(c).Unit()

	caret := opts.Caret
	if caret == "" {
		caret = "▌"
	}

	e := ui.Box(c).FillWidth().Shrink(0)
	if opts.Label != "" {
		e.Label(opts.Label)
	}
	if opts.Lines > 0 {
		e.Height(core.FontSize(c, theme.BodySize) * 1.45 * float32(opts.Lines))
	}
	e.Children(func() {
		ui.Column(c).FillWidth().Grow(1).Gap(u * 0.5).Children(func() {
			MarkdownView(c, MarkdownViewOptions{Source: *text, MaxWidth: ReadingWidth})
			if opts.Streaming {
				caretMark(c, caret)
			}
		})
	})
	return e
}

// caretPeriod is how long a streaming caret takes to blink, in milliseconds.
// It is the loaders' period on purpose: a caret that blinks faster than the
// spinner beside it reads as an alarm rather than as typing.
const caretPeriod float32 = 900

// caretMark is the block at the end of an answer that is still coming.
//
// It is a drawn block rather than a blinking text character, and it blinks
// through its own alpha rather than by being shown and hidden: the caret has
// to be there when it is at its faintest, or a reader looking away for a
// moment comes back to text that has apparently stopped.
//
// Under reduced motion it is simply a steady block. That is the whole answer
// rather than a branch around it, because the caret exists only in motion: a
// reader who asked for no motion gets the resting shape of it — full ink —
// instead of nothing, and a caret drawn as a faded square would read as the
// end of the text.
func caretMark(c *ui.Context, caret string) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	w, h := u*1.25, u*2.5
	still := core.Reduced(c)

	// The block is drawn and the caret character is laid out inside it, so
	// that a screen reader meets the mark as a word while the eye meets it as
	// ink at the size a caret has to be.
	ui.Box(c).Height(h).Shrink(0).Label("still typing").Children(func() {
		ui.Box(c).Width(w).Height(h).Background(k.Accent).
			Draw(func(p *ui.Painter, r ui.Rect) {
				// The alpha is resolved inside the paint rather than captured, so
				// a caret repainting across a palette change takes the window's
				// colour now rather than the one it was built with.
				p.Fill(r, k.Accent.Alpha(caretAlpha(loopPhase(p, caretPeriod, still), still)), 1)
			}).Children(func() {
			// Invisible text of the caret's own shape, so the mark is the same
			// block whether or not the reader can see the character in it.
			ui.Text(c, caret).TextColor(k.Accent.Alpha(0)).FontSize(h)
		})
	})
}

// caretAlpha is how inked a streaming caret is at a point in its blink.
//
// Still, it is 1: the mark is the resting state of the thing and not a frame
// of the animation, and a faint caret reads as text that has ended.
//
// It is a function of two plain values rather than a Painter callback so that
// a test can say what the mark does at any point without painting anything —
// a caret that blinks is the one thing in a chat window a screenshot can
// never check, and this is the value a screenshot would show.
func caretAlpha(phase float32, still bool) float32 {
	if still {
		return 1
	}
	// A raised cosine rather than a hard on/off, so the fade has a moment at
	// each end the eye can follow. A square wave at this rate reads as a
	// flicker rather than as a blink.
	const dim float32 = 0.25
	return float32(float64(dim) + float64(1-dim)*(0.5+0.5*math.Cos(2*math.Pi*float64(phase))))
}

// loopPhase is where a mark is in its loop, 0 to 1. Every clock in this
// package goes through it, so two animations on one screen beat together
// rather than against each other.
//
// Still, it is 0 — the top of the cycle — rather than a frozen arbitrary
// moment: a mark caught mid-fade would read as an interface that has stopped
// halfway rather than as a quieter one.
func loopPhase(p *ui.Painter, period float32, still bool) float32 {
	if still || period <= 0 {
		return 0
	}
	ms := float64(p.Now().UnixMilli())
	return float32(math.Mod(ms/float64(period), 1))
}

// ThinkingBlockOptions configure a ThinkingBlock.
type ThinkingBlockOptions struct {
	// Summary is the one line of the model's reasoning that is worth showing —
	// "Checked the connection pool first". It is what the collapsed block
	// shows, so it is required: a collapsed block with nothing in it is a
	// line with no way to open it.
	Summary string
	// Detail is the rest of the reasoning, shown when it is open. Empty draws
	// an empty body, which is the honest state of a model that summarised
	// its own reasoning and nothing else.
	Detail string
	// Open is the caller's fold state. A ThinkingBlock does not hold it: two
	// views of one turn must agree about whether it is open.
	Open bool
	// Busy says the thinking is still going, which is what the mark beside
	// the summary is for.
	Busy bool
	// Duration is how long the thinking took, for a finished turn.
	Duration string
}

// ThinkingBlockResult carries a ThinkingBlock and the press of its summary.
type ThinkingBlockResult struct {
	// Element is the block.
	Element *ui.Element
	toggled bool
}

// Toggled reports the summary being pressed.
func (r ThinkingBlockResult) Toggled() bool { return r.toggled }

// ThinkingBlock is the model's own reasoning, folded away.
//
// It is quiet on purpose. Reasoning is the longest text in a session and the
// least-read, so it starts folded, shows one line, and opens only where
// somebody has asked what the model did before it answered.
func ThinkingBlock(c *ui.Context, opts ThinkingBlockOptions) ThinkingBlockResult {
	if opts.Summary == "" {
		panic("chat: ThinkingBlock needs a Summary; a collapsed block with nothing " +
			"in it is a line with no way to open it")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	var res ThinkingBlockResult
	head := func() {
		row := ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u*1.5).
			Padding(u, u*1.5).Radius(theme.SmallRadius).Label(opts.Summary).
			Cursor(ui.CursorPointer)
		row.Children(func() {
			if opts.Busy {
				feedback.DotsLoader(c, feedback.LoaderOptions{
					Size: u * 3, Color: k.TextMuted, Label: "Thinking",
				})
			} else {
				caretArrow(c, opts.Open)
			}
			ui.Text(c, opts.Summary).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.MetaSize)).Grow(1).SingleLine()
			if opts.Duration != "" {
				ui.Text(c, opts.Duration).TextColor(k.TextFaint).
					FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
			}
		})
		res.toggled = row.Clicked()
	}

	block := ui.Column(c).FillWidth().Children(func() {
		// The head is built in here rather than out there because an element
		// belongs to whatever was being built when it was made: a header made
		// before the block would land as the block's sibling.
		head()
		if opts.Open {
			if opts.Detail != "" {
				ui.Column(c).FillWidth().Padding(u*1.5, u*1.5, u, u*4).Children(func() {
					MarkdownView(c, MarkdownViewOptions{Source: opts.Detail, Tight: true})
				})
			}
		}
	})
	res.Element = block
	return res
}

// caretArrow is the small triangle that says whether a folded block is open.
func caretArrow(c *ui.Context, open bool) {
	k := core.Tokens(c)
	side := core.Density(c).Unit() * 3
	ui.Box(c).Size(side, side).Shrink(0).Role(ui.RoleNone).
		Draw(func(p *ui.Painter, r ui.Rect) {
			var path ui.Path
			cx, cy := r.X+r.W/2, r.Y+r.H/2
			if open {
				path.MoveTo(r.X+r.W*0.25, cy).LineTo(r.X+r.W*0.75, cy).LineTo(cx, r.Y+r.H*0.78)
			} else {
				path.MoveTo(cx, r.Y+r.H*0.25).LineTo(cx, r.Y+r.H*0.75).LineTo(r.X+r.W*0.78, cy)
			}
			path.Close()
			p.FillPath(&path, k.TextFaint)
		})
}

// ThinkingIndicatorOptions configure a ThinkingIndicator.
type ThinkingIndicatorOptions struct {
	// Label says what is being thought about. It is required: three moving
	// dots beside nothing is a loading screen, and this is a thing that is
	// being worked on.
	Label string
	// Detail is a second line under the label — a tool being called, a file
	// being read.
	Detail string
}

// ThinkingIndicator is the one mark that says the model is working but has
// said nothing yet.
//
// It is deliberately three dots and no words beyond the label: this is the
// state a reader looks at most often and reads least, and anything more on
// screen is something they will wait for rather than read.
func ThinkingIndicator(c *ui.Context, opts ThinkingIndicatorOptions) *ui.Element {
	if opts.Label == "" {
		panic("chat: ThinkingIndicator needs a Label; three dots beside nothing is a " +
			"loading screen, not a status")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	e := ui.Column(c).AlignItems(ui.Start).Gap(u * 0.5).Label(opts.Label)
	e.Children(func() {
		ui.Row(c).AlignItems(ui.Center).Gap(u * 1.5).Children(func() {
			feedback.DotsLoader(c, feedback.LoaderOptions{
				Size: u * 3.5, Color: k.TextMuted, Label: opts.Label,
			})
			ui.Text(c, opts.Label).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.MetaSize))
		})
		if opts.Detail != "" {
			ui.Text(c, opts.Detail).TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize))
		}
	})
	return e
}

// TypingIndicatorOptions configure a TypingIndicator.
type TypingIndicatorOptions struct {
	// Who is typing. It is required, for the same reason ThinkingIndicator
	// needs one: an unattributed dot is a decoration.
	Who string
	// Bubbles is how many bubbles to show, one per line the other party is
	// about to write. Zero shows one.
	Bubbles int
}

// TypingIndicator is "they are writing" in another person's conversation: a
// name and a dot, with no words and no animation beyond the dot.
//
// The distinction from ThinkingIndicator is who is doing it. This one is a
// person's turn coming, and the reader is waiting rather than working, so it is
// a third of the space and no detail line.
func TypingIndicator(c *ui.Context, opts TypingIndicatorOptions) *ui.Element {
	if opts.Who == "" {
		panic("chat: TypingIndicator needs a Who; an unattributed dot is a decoration")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	rows := opts.Bubbles
	if rows <= 0 {
		rows = 1
	}
	side := u * 2.25

	// The same three widths a message bubble has, in the same ink: a typing
	// indicator that is a different shape from the bubble it is standing in
	// for reads as a different kind of thing, and the reader has to re-learn
	// what is arriving.
	bubbleWidths := []float32{u * 14, u * 18, u * 11}

	e := ui.Column(c).AlignItems(ui.Start).Gap(u * 0.5).Label(opts.Who + " is typing")
	e.Children(func() {
		ui.Text(c, opts.Who).TextColor(k.TextMuted).
			FontSize(core.FontSize(c, theme.CaptionSize))
		for i := range rows {
			w := bubbleWidths[i%len(bubbleWidths)]
			ui.Box(c).Width(w).Height(side).Radius(side / 2).
				Background(k.Surface).Shrink(0)
		}
	})
	return e
}

// SendButtonResult carries a SendButton and its press.
type SendButtonResult struct {
	// Element is the button.
	Element *ui.Element
	sent    bool
}

// Sent reports the button being pressed this frame.
func (r SendButtonResult) Sent() bool { return r.sent }

// SendButton is the one primary action in a chat window.
//
// It is a component rather than a Button because it has the one state that a
// plain button does not: it is disabled when there is nothing to send, and the
// reason it is disabled is that a composer knows and a button does not.
func SendButton(c *ui.Context, opts SendButtonOptions) SendButtonResult {
	if opts.Label == "" {
		panic("chat: SendButton needs a Label; a send button with no name is a shape")
	}
	var res SendButtonResult
	btn := input.Button(c, opts.Label, input.ButtonOptions{
		Primary:  true,
		Disabled: opts.Disabled,
		Icon:     sendGlyph,
		Label:    opts.Label,
	})
	if btn.Clicked() {
		res.sent = true
	}
	res.Element = btn
	return res
}

// SendButtonOptions configure a SendButton.
type SendButtonOptions struct {
	// Label names the action. Required.
	Label string
	// Disabled is the composer's to set: an empty draft cannot be sent, and
	// a caller that sends nothing and says nothing about it will get a button
	// that pretends to work.
	Disabled bool
}

// StopGeneratingButtonResult carries a StopGeneratingButton and its press.
type StopGeneratingButtonResult struct {
	// Element is the button.
	Element *ui.Element
	stopped bool
}

// Stopped reports the button being pressed this frame. It is false while
// Busy, for the reason feedback.InterruptButton gives: the press has already
// been taken, and acting on it again would send the second interrupt.
func (r StopGeneratingButtonResult) Stopped() bool { return r.stopped }

// StopGeneratingButton is the button that stops an answer that is still
// arriving.
//
// It is a separate component rather than a SendButton with a different label
// because it lives in a different place in the window and means the opposite
// thing. A SendButton sits at the end of the draft; this one replaces the
// keyboard's action while an answer streams, because the thing a person wants
// to press while an answer streams is "stop", and the send button is the one
// control within reach.
func StopGeneratingButton(c *ui.Context, opts StopGeneratingButtonOptions) StopGeneratingButtonResult {
	label := opts.Label
	if label == "" {
		label = core.Msg(c, "chat.stop", core.Def("Stop generating"))
	}
	var res StopGeneratingButtonResult
	btn := input.Button(c, label, input.ButtonOptions{
		Disabled: opts.Busy,
		Label:    label,
	})
	if btn.Clicked() {
		res.stopped = true
	}
	res.Element = btn
	return res
}

// StopGeneratingButtonOptions configure a StopGeneratingButton.
type StopGeneratingButtonOptions struct {
	// Label names the action; empty takes the library's "Stop generating".
	Label string
	// Busy says the stop has been asked for and is on its way. The button
	// stops taking presses, which is the one thing a stop button must do the
	// moment it is pressed.
	Busy bool
}

// ScrollToBottomButtonResult carries a ScrollToBottomButton and its press.
type ScrollToBottomButtonResult struct {
	// Element is the button, nil while Hidden is true.
	Element *ui.Element
	pressed bool
	// visible reports whether the button was drawn this frame.
	visible bool
}

// Pressed reports the button being pressed this frame.
func (r ScrollToBottomButtonResult) Pressed() bool { return r.pressed }

// Visible reports whether the button was drawn. It is there because the
// component draws nothing when it is not wanted, and a caller that only asked
// for the press would have no way to know there was nothing on screen to
// press.
func (r ScrollToBottomButtonResult) Visible() bool { return r.visible }

// ScrollToBottomButtonOptions configure a ScrollToBottomButton.
type ScrollToBottomButtonOptions struct {
	// Pending is how many new turns have arrived below the reader. Zero hides
	// the button: a jump-to-bottom control sitting on the transcript with
	// nothing below is a control for going nowhere.
	Pending int
	// Label names the button; empty takes "Jump to latest", with the count
	// appended when there is one.
	Label string
	// Hidden hides it entirely, for a transcript that is never scrolled.
	Hidden bool
}

// ScrollToBottomButton is the pill that takes a reader back to the newest turn
// after they have scrolled up to read something.
//
// It hangs absolutely at the bottom centre rather than sitting in the column,
// because a control in the flow would take a turn's worth of space and push
// the transcript as it appeared and disappeared.
func ScrollToBottomButton(c *ui.Context, opts ScrollToBottomButtonOptions) ScrollToBottomButtonResult {
	var res ScrollToBottomButtonResult
	if opts.Hidden || opts.Pending <= 0 {
		return res
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	label := opts.Label
	if label == "" {
		label = core.Msg(c, "chat.jumpToLatest", core.Def("Jump to latest"))
		if opts.Pending > 1 {
			label += " (" + internal.Commas(opts.Pending) + ")"
		}
	}

	// The button is built inside the holder, not beside it: an element belongs
	// to whatever was being built when it was made, so a button made out here
	// would land as a sibling of the holder and float in the column above it.
	holder := ui.Box(c).Absolute().Bottom(u * 2).LeftPercent(50).Grow(0).Shrink(0).
		AlignItems(ui.Center).Children(func() {
		btn := ui.ButtonBase(c).AlignItems(ui.Center).Gap(u).Radius(theme.PillRadius).
			Padding(u, u*3).Background(k.Fill).TextColor(k.OnFill).
			Label(label).Tooltip(label).Shadow(0, u, u*3, 0, shadowTint(k)).
			Children(func() {
				caretDown(c, k.OnFill)
				ui.Text(c, label).TextColor(k.OnFill).FontSize(core.FontSize(c, theme.RowSize))
			})
		if btn.Clicked() {
			res.pressed = true
		}
	})
	res.Element = holder
	res.visible = true
	return res
}

// shadowTint is how dark a floating pill's shadow is. A dark window already
// sits low against its own page, so the same shadow reads as a heavier box
// there than in the light one.
func shadowTint(k theme.Tokens) ui.Color {
	if k.IsDark() {
		return ui.RGBA(0, 0, 0, 0.4)
	}
	return ui.RGBA(0, 0, 0, 0.18)
}

// caretDown is the arrow on the pill, drawn rather than typed for the reason
// every mark in this library is drawn: a character is a different shape, or a
// different size, or missing.
func caretDown(c *ui.Context, col ui.Color) {
	side := core.Density(c).Unit() * 3
	ui.Box(c).Size(side, side).Shrink(0).Role(ui.RoleNone).
		Draw(func(p *ui.Painter, r ui.Rect) {
			var path ui.Path
			path.MoveTo(r.X+r.W*0.28, r.Y+r.H*0.42).
				LineTo(r.X+r.W*0.5, r.Y+r.H*0.64).
				LineTo(r.X+r.W*0.72, r.Y+r.H*0.42)
			p.StrokePath(&path, 1.5, col)
		})
}

// sendGlyph is the arrow on the SendButton, in the same stroke style as the
// rest of the library's marks.
var sendGlyph = mustGlyph(
	`<path d="M12 4.5v13"/><path d="M6.5 12.5 12 18l5.5-5.5"/>`)

// SearchProgressOptions configure a SearchProgress.
type SearchProgressOptions struct {
	// Query is what is being looked for. It is shown, because a progress bar
	// with no query on it is a bar and the reader has no way to tell whether
	// it is searching for the right thing.
	Query string
	// Found is how many matches there are so far.
	Found int
	// Busy says the search is still running.
	Busy bool
	// Match is the index of the match being read, 1-based; zero for none.
	Match int
}

// SearchProgress is the line under a transcript saying what is being searched
// for and how far along it is.
//
// It is not the reader's own search field: it reports on a search over a
// transcript that is already on screen, and it never takes the focus, because
// the transcript is what the reader is reading.
func SearchProgress(c *ui.Context, opts SearchProgressOptions) *ui.Element {
	if opts.Query == "" {
		panic("chat: SearchProgress needs a Query; a search bar with no query on it " +
			"is a bar")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	e := ui.Column(c).FillWidth().Gap(u).Label(opts.Query)
	e.Children(func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.5).Children(func() {
			display.Icon(c, display.IconSearch, display.IconOptions{
				Name: "Search", Muted: true, Size: u * 3.5,
			})
			ui.Text(c, opts.Query).TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).Grow(1).SingleLine()
			count := internal.Commas(opts.Found)
			if opts.Match > 0 {
				count = internal.Commas(opts.Match) + " of " + count
			}
			ui.Text(c, count).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
		})
		if !opts.Busy {
			// A finished search says so with the word. A bar at a full width
			// and a bar that is not moving look identical in a screenshot,
			// and this is the one place the reader has to be told the answer
			// is the whole answer.
			ui.Text(c, core.Msg(c, "chat.search.done", core.Def("Search finished"))).
				TextColor(k.TextFaint).FontSize(core.FontSize(c, theme.CaptionSize))
		}
		feedback.Progress(c, feedback.ProgressOptions{
			Label:  "Searching the conversation",
			Height: u * 1.5,
			// Negative is the library's own word for work of unknown length,
			// which is what a search over a transcript is: the number of
			// matches is not known until it has finished looking.
			Value: -1,
		})
	})
	return e
}

// mustGlyph parses one of this package's marks. They are compile-time
// constants in effect: a malformed one is a bug in this file, not something a
// caller could cause, so it panics here rather than becoming an empty square.
func mustGlyph(shapes string) *ui.SVG {
	return ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" ` +
		`fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" ` +
		`stroke-linejoin="round">` + shapes + `</svg>`))
}
