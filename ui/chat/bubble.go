package chat

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// The one style a transcript is drawn in.
//
// A chat window has four kinds of turn — what I said, what the model said,
// what the system said, what a tool said — and the only differences between
// them are worth exactly two: which side of the column the bubble sits on, and
// which corner the tail is under. Everything else a bubble is made of — the
// padding inside it, the curve of its corners, the gap to the next turn, the
// tail's size — is one set of numbers, and it is written down once here.
//
// That is the whole reason [bubble] exists rather than four functions. A
// transcript whose system notices had their own margins and whose tool output
// had its own would read as four interfaces sharing a column: the eye has to
// re-measure the rhythm every time the speaker changes, which is most turns in
// a real session.

// ReadingWidth is how wide a bubble of prose may get before it wraps. It is a
// fixed number for the reason the board's columns are: a line long enough to
// lose its start is a line nobody reads, and the bubble cannot scroll sideways
// to help.
const ReadingWidth float32 = 680

// bubbleMetrics is one transcript's spacing and shape, at one density.
//
// It is a value rather than five functions so that the numbers cannot drift
// apart: a bubble whose padding came from one call and whose radius came from
// another is exactly the drift this file exists to prevent.
type bubbleMetrics struct {
	// PadY and PadX are the padding inside a bubble.
	PadY, PadX float32
	// Radius is the curve on a bubble's corners, before the tail squares one.
	Radius float32
	// Gap is the space between one turn and the next.
	Gap float32
	// Tail is the side of the triangle under the corner that carries it.
	Tail float32
	// MaxWidth is how wide prose may run before it wraps.
	MaxWidth float32
}

// bubbleMetricsFor resolves the metrics at a window's density unit. Every
// number in it is a multiple of u, which is what makes Comfortable denser
// rather than merely larger.
func bubbleMetricsFor(u float32) bubbleMetrics {
	return bubbleMetrics{
		PadY:     u * 2,
		PadX:     u * 3,
		Radius:   theme.CardRadius,
		Gap:      u * 3,
		Tail:     u * 2,
		MaxWidth: ReadingWidth,
	}
}

// Role is who a turn is from. It is the only thing that decides a bubble's
// place in the column and which corner its tail is under.
type Role int

const (
	// RoleUser is what the person typed: the filled bubble, on the right.
	RoleUser Role = iota
	// RoleAssistant is what the model answered: on the left, on the page's
	// own surface, so it is read as the conversation rather than as a quote.
	RoleAssistant
	// RoleSystem is what the application said: dated, joined, centred, and
	// without a tail — it is not addressed to either side.
	RoleSystem
	// RoleTool is what a tool ran and returned: on the left, in a quieter
	// face, and without a tail, because nobody said it.
	RoleTool
)

func (r Role) String() string {
	switch r {
	case RoleUser:
		return "user"
	case RoleAssistant:
		return "assistant"
	case RoleSystem:
		return "system"
	case RoleTool:
		return "tool"
	}
	return "user"
}

// RoleOf reads a role from a string, which is what a stored conversation gives
// back. An unknown name is the user rather than a panic: a message row written
// by an older version of the app is not a bug worth stopping for, and the user
// side is the one that can still be acted on.
func RoleOf(s string) Role {
	switch s {
	case "assistant":
		return RoleAssistant
	case "system":
		return RoleSystem
	case "tool":
		return RoleTool
	}
	return RoleUser
}

// bubbleInk is a bubble's background and text, per role.
//
// The user's bubble is the filled one because it is the one thing on the
// screen that is certainly not the model's to be confused with; the model's is
// the page's own background with a hairline, so a long answer reads as prose on
// the page rather than as a box; system and tool turns share the surface step
// and differ only in tone, which is what makes them read as furniture.
func bubbleInk(role Role, k theme.Tokens) (bg, fg ui.Color) {
	switch role {
	case RoleUser:
		return k.Fill, k.OnFill
	case RoleAssistant:
		return k.Background, k.Text
	case RoleTool:
		return k.Surface, k.Text
	}
	return k.Surface, k.TextMuted
}

// bubbleBorder is the hairline round a bubble. The user's is drawn in its own
// fill so a filled bubble keeps its edge against a filled page; the rest take
// the window's border.
func bubbleBorder(role Role, k theme.Tokens) ui.Color {
	if role == RoleUser {
		return k.Fill
	}
	return k.Border
}

// bubbleSide is where a role's bubble sits in the column. The two centred
// roles are centred because the eye reads a system notice as an interruption
// to the conversation rather than as a turn in it, and a centred thing says so
// with no extra mark.
func bubbleSide(role Role) ui.Align {
	switch role {
	case RoleUser:
		return ui.End
	case RoleSystem:
		return ui.Center
	}
	return ui.Start
}

// bubbleTailCorner is which corner carries the tail: bottom-right for the
// side the user is on, bottom-left for the side the model is on. A role with no
// tail gets -1, and bubble draws no triangle for it.
func bubbleTailCorner(role Role) int {
	switch role {
	case RoleUser:
		return 2 // bottom-right
	case RoleAssistant:
		return 3 // bottom-left
	}
	return -1
}

// bubbleCorners returns a bubble's four radii in MyGo's order — top-left,
// top-right, bottom-right, bottom-left — with the tail's corner squared.
//
// Squaring rather than adding a triangle under a rounded bubble is what stops
// the two shapes reading as two things: the bubble's own curve is flattened
// exactly where the tail starts, so the tail looks grown out of it rather than
// stuck on.
func bubbleCorners(role Role, radius float32) (tl, tr, br, bl float32) {
	tl, tr, br, bl = radius, radius, radius, radius
	switch bubbleTailCorner(role) {
	case 2:
		br, bl = 0, radius
	case 3:
		bl, br = 0, radius
	}
	return tl, tr, br, bl
}

// bubbleSkin is what bubble is told about one turn. Everything about how it
// looks is either here or comes from bubbleMetrics; nothing is passed twice
// with two names.
type bubbleSkin struct {
	// Role decides the side, the ink, the corners and the tail.
	Role Role
	// MaxWidth caps the bubble; zero takes the reading width.
	MaxWidth float32
	// Tail draws the triangle. It defaults to on for the two roles that have
	// one and off for the two that do not, so a caller never has to say
	// "no tail" to get no tail.
	Tail *bool
	// Muted draws the body's text in the secondary tone. Only the model's
	// bubble can be muted: the other three are already muted or filled, and
	// muting a filled bubble would be a shade of grey nobody can read.
	Muted bool
	// Monospace sets the body's face. It is set on the box rather than on each
	// child so that a body of forty elements inherits it and none of them has
	// to remember.
	Monospace bool
	// Label names the bubble for assistive technology. Empty leaves it
	// unnamed, which is right for a bubble whose text says it all.
	Label string
}

// bubble is the box every turn in a transcript is drawn in: the padding, the
// corners, the hairline, the tail, the ink, the type and the place in the
// column, all resolved from bubbleMetrics and bubbleSkin and nothing else.
//
// It is deliberately not exported. Four roles, one box — a caller that needed
// "a rounded box, chat-shaped" would be writing a second bubble, and the only
// way to stop that is to leave the one box where the four callers can all see
// it and nowhere else.
func bubble(c *ui.Context, s bubbleSkin, body func()) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	m := bubbleMetricsFor(u)
	bg, fg := bubbleInk(s.Role, k)
	if s.Muted && s.Role == RoleAssistant {
		fg = k.TextMuted
	}
	tl, tr, br, bl := bubbleCorners(s.Role, m.Radius)

	width := s.MaxWidth
	if width <= 0 {
		width = m.MaxWidth
	}

	size := theme.BodySize
	if s.Monospace {
		size = theme.MetaSize
	}
	e := ui.Box(c).Padding(m.PadY, m.PadX).Background(bg).TextColor(fg).
		FontSize(core.FontSize(c, size)).
		Radius(tl, tr, br, bl).
		BorderWidth(theme.BorderWidth).BorderColor(bubbleBorder(s.Role, k)).
		MaxWidth(width)
	if s.Label != "" {
		e.Label(s.Label)
	}

	corner := bubbleTailCorner(s.Role)
	if corner < 0 {
		if s.Tail != nil && *s.Tail {
			panic("chat: a system or tool bubble has no tail to draw; its corner is the " +
				"page's edge, not a speaker's")
		}
	} else if s.Tail == nil || *s.Tail {
		size, side := m.Tail, corner
		e.DrawOver(func(p *ui.Painter, r ui.Rect) {
			drawTail(p, r, size, side, bg)
		})
	}

	if body != nil {
		e.Children(body)
	}
	return e
}

// drawTail paints the triangle under a bubble's tail corner.
//
// It is a path rather than a filled square because a square's corners stick
// out under a curved bubble and the eye reads the outline it does not expect.
// The two points it shares with the bubble are inside the box, so the join is
// covered by the bubble's own background.
func drawTail(p *ui.Painter, r ui.Rect, size float32, corner int, col ui.Color) {
	if size <= 0 || r.W <= 0 {
		return
	}
	var path ui.Path
	switch corner {
	case 2: // bottom-right
		y := r.Y + r.H - 0.5
		path.MoveTo(r.X+r.W-size, y).
			LineTo(r.X+r.W, y).
			LineTo(r.X+r.W, y+size)
	default: // bottom-left
		y := r.Y + r.H - 0.5
		path.MoveTo(r.X+size, y).
			LineTo(r.X, y).
			LineTo(r.X, y+size)
	}
	path.Close()
	p.FillPath(&path, col)
}

// MessageBubbleOptions configure a MessageBubble.
type MessageBubbleOptions struct {
	// Role decides the side, the ink, the corners and the tail.
	Role Role
	// Author is who is speaking, for the roles that have a name.
	Author string
	// Time is when, for the roles that are dated.
	Time string
	// Quoted is the turn this one is answering, drawn above the bubble.
	Quoted string
	// Selected puts a mark in the gutter beside the turn. It is the caller's
	// flag: two views of one transcript must agree about which turn is chosen.
	Selected bool
	// Monospace draws the body in the monospaced face, for a bubble that is
	// mostly code.
	Monospace bool
	// Muted draws the body in the secondary tone.
	Muted bool
	// MaxWidth caps the bubble; zero takes the reading width.
	MaxWidth float32
	// Tail overrides whether the triangle is drawn. nil keeps the role's own
	// answer, which is what nearly every caller wants.
	Tail *bool
	// Label names the bubble for assistive technology; empty is fine, since
	// the body usually says enough.
	Label string
}

// MessageBubbleResult carries a MessageBubble and the press of it.
type MessageBubbleResult struct {
	// Element is the whole turn: the header, the bubble and its actions.
	Element *ui.Element
	clicked bool
}

// Clicked reports the bubble being pressed, which is how a transcript asks to
// choose a turn. Selection is the caller's flag; the press is the only thing
// the bubble reports, because it is the only thing a person does to one.
func (r MessageBubbleResult) Clicked() bool { return r.clicked }

// MessageBubble is one turn of a conversation: who said it, when, what they
// said, and — only when the caller asks — what can be done with it.
//
// The body is the caller's, because what a turn contains is not the box's
// business: a markdown answer, a list of tool calls, a table, a player's
// waveform. What this component owns is everything around it, so that all of
// those land in the same column with the same rhythm.
func MessageBubble(c *ui.Context, opts MessageBubbleOptions, body func()) MessageBubbleResult {
	u := core.Density(c).Unit()

	label := opts.Label
	if label == "" {
		label = bubbleLabel(opts)
	}

	// The turn is a row with a gutter, not a column: the gutter is what puts
	// the selection mark on the side the bubble is not on, and a bubble with
	// a mark beside it cannot be laid out by AlignItems alone.
	turn := ui.Row(c).FillWidth().AlignItems(ui.Start).Gap(u * 1.5).Label(label)

	var res MessageBubbleResult
	turn.Children(func() {
		selectionMark(c, opts.Selected, bubbleSide(opts.Role))
		ui.Column(c).Grow(1).AlignItems(bubbleSide(opts.Role)).Gap(u * 0.75).Children(func() {
			if opts.Author != "" || opts.Time != "" {
				MessageHeader(c, MessageHeaderOptions{
					Author: opts.Author, Time: opts.Time, Role: opts.Role,
				})
			}
			if opts.Quoted != "" {
				// The quote hangs off the side this reply is on, so it lands
				// against the same edge as the bubble below it rather than
				// always on the left.
				QuoteReply(c, QuoteReplyOptions{
					Text: opts.Quoted, Side: bubbleSide(opts.Role),
				})
			}
			res.clicked = bubble(c, bubbleSkin{
				Role:      opts.Role,
				MaxWidth:  opts.MaxWidth,
				Tail:      opts.Tail,
				Muted:     opts.Muted,
				Monospace: opts.Monospace,
				Label:     label,
			}, body).Clicked()
		})
	})
	res.Element = turn
	return res
}

// selectionMark is the dot in the gutter that says which turn is chosen.
//
// It is a gutter mark rather than a ring round the bubble for a reason that is
// easy to get wrong: the bubble's corners are rounded on three corners and
// square on the one its tail is under, so a ring drawn round it would trace
// that corner twice and the outline would look broken exactly where it is
// meant to look finished. A mark outside the bubble has no corner to trace.
//
// It keeps its place whether or not it is drawn, so choosing a turn moves
// nothing: a transcript that jumped sideways by two dots every time the
// selection moved would be unreadable while being read.
func selectionMark(c *ui.Context, selected bool, side ui.Align) *ui.Element {
	u := core.Density(c).Unit()
	k := core.Tokens(c)
	d := u * 1.5
	mark := ui.Box(c).Width(d).Height(d).Radius(d / 2).Shrink(0).Role(ui.RoleNone)
	if selected {
		mark = mark.Background(k.Accent).Label("selected")
	}
	if side == ui.End {
		// The user's bubble is on the right, so its mark goes on the left,
		// before the bubble — which is why the mark is made first either way.
		mark = mark.AlignItems(ui.End)
	}
	return mark
}

// bubbleLabel is what a bubble is called when the caller gave it nothing: the
// author and the role, so a screen reader says "Assistant" before it reads the
// whole answer rather than saying nothing at all.
func bubbleLabel(opts MessageBubbleOptions) string {
	switch {
	case opts.Author != "":
		return opts.Author + ", " + opts.Role.String()
	case opts.Role == RoleSystem, opts.Role == RoleTool:
		return opts.Role.String() + " message"
	}
	return ""
}
