package files

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// The four components in this file are the ones a file interface shows about
// other people rather than about the files. They are all the same shape: a
// mark that says somebody is here, and nothing else.

// Reaction is one emoji somebody left on a file or a comment, and how many of
// them there are.
type Reaction struct {
	// Emoji is the character itself. It is text rather than a glyph from
	// the icon set, because a reaction is something a person picked out of
	// the ones everybody knows, and a bespoke drawing of it would be one
	// more thing to learn.
	Emoji string
	// Count is how many people left it.
	Count int
	// Mine reports that the reader is one of them, which is what gives the
	// chip a filled face instead of an outline. It is the only way to tell
	// at a glance which reactions you are already part of.
	Mine bool
	// Selected reports that the chip is the one the pointer is on, for a
	// tooltip that says who.
	Selected bool
}

// ReactionsOptions configure a Reactions.
type ReactionsOptions struct {
	// Height is the height of the row; zero is the library's own.
	Height float32
	// Max is how many chips to show before the rest become a count. Zero
	// shows them all, which is wrong past about five: a row of nine emoji is
	// a wall.
	Max int
	// Size is the height of one chip; zero gives the library's.
	Size float32
	// Label names the row; it is required, because a row of emoji with no
	// name is nine anonymous marks.
	Label string
}

// ReactionsResult carries a Reactions and what was pressed in it.
type ReactionsResult struct {
	// Element is the row.
	Element *ui.Element
	// reacted is the emoji pressed this frame, empty for none.
	reacted string
	// added reports the add button being pressed this frame, for the caller
	// to put a picker up.
	added bool
}

// Reacted returns the emoji pressed this frame, empty for none. It is the
// emoji rather than a chip's index, because a chip's index is this frame's
// order and the caller's action is about the character.
func (r ReactionsResult) Reacted() string { return r.reacted }

// Added reports the add button being pressed this frame. Which emoji is then
// the caller's business: picking one needs a picker, and a picker needs to
// know what the file is.
func (r ReactionsResult) Added() bool { return r.added }

// Reactions is the emoji on a file, each with its count, and the reader's own
// reactions filled so they can be told apart at a glance.
//
// Past Max the rest become a single count rather than disappearing. A row
// that quietly dropped four reactions is lying about how popular a file is,
// and the count is the honest summary of what was left out.
func Reactions(c *ui.Context, reactions []Reaction, opts ReactionsOptions) ReactionsResult {
	u := core.Density(c).Unit()
	if opts.Label == "" {
		panic("files: Reactions needs a Label; a row of emoji with no name is nine anonymous marks")
	}
	h := opts.Size
	if h <= 0 {
		h = u * 7
	}
	shown := reactions
	if opts.Max > 0 && len(shown) > opts.Max {
		shown = shown[:opts.Max]
	}

	var res ReactionsResult
	row := ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 0.75).
		Label(opts.Label).Role(ui.RoleNone)
	row.Children(func() {
		for _, r := range shown {
			r := r
			chip := reactionChip(c, r, h)
			if chip.Clicked() {
				res.reacted = r.Emoji
			}
		}
		if extra := hiddenEmoji(reactions, opts.Max); extra > 0 {
			moreChip(c, extra, h)
		}
		ui.Box(c).Grow(1)
		add := addButton(c, h)
		if add.Clicked() {
			res.added = true
		}
	})
	res.Element = row
	return res
}

// hiddenEmoji is how many reactions are behind the cap: the whole ones that
// did not fit, and the counts of the ones that were cut in half.
func hiddenEmoji(all []Reaction, max int) int {
	if max <= 0 || len(all) <= max {
		return 0
	}
	n := 0
	for _, r := range all[max:] {
		n += r.Count
	}
	return n
}

// reactionChip is one emoji and its count, filled when the reader is one of
// the people who left it.
func reactionChip(c *ui.Context, r Reaction, h float32) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	bg, fg := k.Surface, k.TextMuted
	if r.Mine {
		bg, fg = k.AccentBg, k.AccentText
	}
	chip := ui.ButtonBase(c).Height(h).Radius(theme.PillRadius).
		Padding(0, u*1.5).Background(bg).TextColor(fg).
		Label(r.Emoji + ", " + itoa(r.Count) + " reactions").Role(ui.RoleNone)
	chip.Children(func() {
		ui.Text(c, r.Emoji).FontSize(core.FontSize(c, theme.BodySize))
		ui.Text(c, itoa(r.Count)).TextColor(fg).
			FontSize(core.FontSize(c, theme.CaptionSize))
	})
	return chip
}

// moreChip is the "+7" that stands for the reactions that did not fit.
func moreChip(c *ui.Context, n int, h float32) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	ui.Box(c).Height(h).Radius(theme.PillRadius).Padding(0, u*1.5).
		Background(k.Surface).Label(itoa(n) + " more reactions").Children(func() {
		ui.Text(c, "+"+itoa(n)).TextColor(k.TextMuted).
			FontSize(core.FontSize(c, theme.CaptionSize))
	})
}

// addButton is the small cross at the end of the row, for the reader to add
// one of their own.
func addButton(c *ui.Context, h float32) *ui.Element {
	k := core.Tokens(c)
	btn := ui.ButtonBase(c).Size(h, h).Shrink(0).Radius(h / 2).
		Background(k.Surface).TextColor(k.TextMuted).
		Label(core.Msg(c, "files.addReaction", core.Def("Add a reaction"))).Role(ui.RoleNone)
	btn.Children(func() {
		ui.Text(c, "+").TextColor(k.TextMuted).
			FontSize(core.FontSize(c, theme.BodySize))
	})
	return btn
}

// LiveIndicatorOptions configure a LiveIndicator.
type LiveIndicatorOptions struct {
	// Label names the mark; required. A pulsing dot with no name is the one
	// mark in this library that a screen reader has nothing at all to say.
	Label string
	// Live says whether anything is actually live. False draws the mark at
	// rest, which is not the same as drawing nothing: the connection is
	// there and it is quiet, and those are different facts.
	Live bool
	// Tone is how loud it is; zero is the library's own lively colour.
	Tone core.Severity
}

// LiveIndicator is the mark that says something is happening right now:
// a dot that pulses, in the one colour in this interface that does not
// change with the theme.
//
// It respects the reduced-motion preference by drawing its resting state
// rather than a still version of the animation. That is not the same thing: a
// still pulsing dot is a dot frozen at its brightest, which reads as a
// brighter dot, and the whole reason a person turns motion off is that a
// bright moving mark is the thing they cannot stop noticing.
func LiveIndicator(c *ui.Context, opts LiveIndicatorOptions) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if opts.Label == "" {
		panic("files: LiveIndicator needs a Label; a dot with no name says nothing out loud")
	}
	col := k.Lively
	if opts.Tone != core.Neutral {
		_, col = opts.Tone.Pair(k)
	}
	dot := u * 2.5

	mark := ui.Box(c).Size(dot*2.6, dot*2.6).Shrink(0).Role(ui.RoleNone).
		Label(opts.Label)
	mark.Children(func() {
		ring := ui.Box(c).Size(dot*2, dot*2).Shrink(0).Radius(dot)
		if opts.Live && !core.Reduced(c) {
			// The halo is a second, larger dot in the lively colour at low
			// alpha. Drawn rather than animated by a transition because a
			// transition is a thing that can be skipped, and this is the
			// one mark in the interface that has to be unmissable.
			ui.Box(c).Size(dot*2, dot*2).Shrink(0).Radius(dot).
				Background(col.Alpha(0.28))
		}
		ring.Children(func() {
			if !opts.Live {
				// Not live: the mark is the ring with nothing in it, which
				// says "the connection is there and nothing is happening".
				ui.Box(c).Size(dot, dot).Shrink(0).Radius(dot/2).
					Background(col.Alpha(0.4)).
					Border(theme.BorderWidth, k.Border)
				return
			}
			ui.Box(c).Size(dot, dot).Shrink(0).Radius(dot / 2).Background(col)
		})
	})
	return mark
}

// Viewer is somebody present on a file: who they are, and whether they are
// here right now. It is a type of its own rather than the Person of the
// share panel, because the two answer different questions — one is who you
// may give this to, the other is who is looking at it with you.
type Viewer struct {
	// Name is their name; the avatar is built from it.
	Name string
	// Doing is what they are doing — "Editing", "Viewing" — shown under the
	// name in the caller-supplied form.
	Doing string
	// Here reports that they are on the file at all. Somebody who was here
	// an hour ago is not here now, and an avatar that does not say which it
	// is sends people to a file to find out.
	Here bool
}

// PresenceAvatarsOptions configure a PresenceAvatars.
type PresenceAvatarsOptions struct {
	// Max is how many faces to show before the rest become a count. Zero
	// shows them all, which is right for a small group and wrong for a
	// shared drive with two hundred readers on it.
	Max int
	// Total is how many people there are in all, when it is more than the
	// caller passed. Zero means "however many were given".
	Total int
	// WithNames puts the names under the faces, for a wide enough panel.
	WithNames bool
	// Empty draws instead of the faces when nobody is here.
	Empty func()
}

// PresenceAvatarsResult carries a PresenceAvatars and what was pressed.
type PresenceAvatarsResult struct {
	// Element is the cluster.
	Element *ui.Element
	// picked is the name pressed this frame, empty for none.
	picked string
}

// Picked returns the name pressed this frame, empty for none.
func (r PresenceAvatarsResult) Picked() string { return r.picked }

// PresenceAvatars is who else is on this file: the faces, the ones who are
// here now filled and the rest outlined.
//
// Nobody here draws the caller's own words rather than a generic "nobody".
// "Ada is editing" and "nobody else is here" are different sentences about
// the same state, and only the caller knows which one it is — a comment
// thread says one, a shared document says the other.
//
// The faces overlap, which is the only way a dozen of them fit in the corner
// of a header. The one in front is the most recent, because a reader looking
// at a crowded corner wants to see who just arrived.
func PresenceAvatars(c *ui.Context, people []Viewer, opts PresenceAvatarsOptions) PresenceAvatarsResult {
	k, u := core.Tokens(c), core.Density(c).Unit()

	shown := people
	if opts.Max > 0 && len(shown) > opts.Max {
		shown = shown[:opts.Max]
	}
	total := opts.Total
	if total <= 0 {
		total = len(people)
	}

	var res PresenceAvatarsResult
	row := ui.Row(c).AlignItems(ui.Center).Gap(u * 0.5).
		Label(presenceLabel(people)).Role(ui.RoleNone)
	row.Children(func() {
		if len(shown) == 0 {
			if opts.Empty != nil {
				opts.Empty()
			} else {
				ui.Text(c, core.Msg(c, "files.alone", core.Def("Nobody else is here"))).
					TextColor(k.TextMuted).FontSize(core.FontSize(c, theme.RowSize))
			}
			return
		}
		for i, p := range shown {
			p, i := p, i
			// Each face overlaps the one before it, so a dozen take the width
			// of four and the front one stays whole. The overlap is the
			// margin, and a margin has to be applied before the element is
			// made: MyGo parents a child by where it was CREATED, so a face
			// adjusted after its row began is that row's sibling.
			overlap := float32(0)
			if i > 0 {
				overlap = -u * 3.25
			}
			ui.Row(c).MarginX(overlap).Children(func() {
				if opts.WithNames {
					ui.Column(c).AlignItems(ui.Center).Gap(u * 0.25).Children(func() {
						avatarFace(c, p)
						ui.Text(c, p.Name).TextColor(k.TextMuted).MaxLines(1).
							FontSize(core.FontSize(c, theme.CaptionSize))
					})
					return
				}
				avatarFace(c, p)
			})
		}
		if rest := total - len(shown); rest > 0 {
			moreFace(c, rest)
		}
	})
	res.Element = row
	return res
}

// avatarFace is one person's initial in a ring, filled when they are here and
// outlined when they are not.
func avatarFace(c *ui.Context, p Viewer) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	side := u * 7
	bg, fg := k.Surface, k.TextMuted
	if p.Here {
		bg, fg = k.AccentBg, k.AccentText
	}
	face := ui.ButtonBase(c).Size(side, side).Shrink(0).Radius(side / 2).
		Background(bg).Label(p.Name + ", " + doingWord(p)).
		Tooltip(p.Name + " — " + doingWord(p))
	face.Children(func() {
		ui.Text(c, initials(p.Name)).TextColor(fg).Bold().
			FontSize(core.FontSize(c, theme.CaptionSize))
	})
	return face
}

// moreFace is the count of the people who did not fit.
func moreFace(c *ui.Context, n int) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	side := u * 7
	ui.Box(c).Size(side, side).Shrink(0).Radius(side / 2).MarginX(-u * 3.25).
		Background(k.SurfacePressed).Label(itoa(n) + " more people").Children(func() {
		ui.Text(c, "+"+itoa(n)).TextColor(k.TextMuted).
			FontSize(core.FontSize(c, theme.CaptionSize)).Bold()
	})
}

// doingWord is what a person is doing here, defaulting to "here" rather than
// to an empty string: an avatar with nothing under it reads as broken.
func doingWord(p Viewer) string {
	if p.Doing != "" {
		return p.Doing
	}
	if p.Here {
		return core.Def("here")
	}
	return core.Def("away")
}

// presenceLabel is the cluster read out loud: how many people and who they
// are, because nine initials are nine noises.
func presenceLabel(people []Viewer) string {
	here := 0
	names := make([]string, 0, len(people))
	for _, p := range people {
		if p.Here {
			here++
		}
		names = append(names, p.Name+", "+doingWord(p))
	}
	if here == 0 {
		return itoa(len(people)) + " people, none here: " + joinNames(names)
	}
	return itoa(here) + " of " + itoa(len(people)) + " here: " + joinNames(names)
}

func joinNames(names []string) string {
	switch len(names) {
	case 0:
		return ""
	case 1:
		return names[0]
	case 2:
		return names[0] + " and " + names[1]
	}
	out := ""
	for i, n := range names {
		if i == len(names)-1 {
			out += " and " + n
			break
		}
		if i > 0 {
			out += ", "
		}
		out += n
	}
	return out
}

// RemoteCursorOptions configure a RemoteCursor.
type RemoteCursorOptions struct {
	// Name is whose cursor it is. It is required, and it is in the label
	// rather than in a tooltip: a cursor with a tooltip is a cursor that
	// has to be hovered to be understood.
	Name string
	// Doing is what they are doing, shown under the name.
	Doing string
	// X and Y are where the cursor is, relative to whatever it is drawn
	// over — in DIPs, so the same numbers work at every scale factor.
	X, Y float32
	// Height is how tall the pointer is; zero gives the library's.
	Height float32
}

// RemoteCursorResult carries a RemoteCursor and whether it was followed.
type RemoteCursorResult struct {
	// Element is the mark.
	Element *ui.Element
	// followed reports the mark being pressed this frame — the "jump to
	// them" action, which is the only thing a person does with somebody
	// else's cursor.
	followed bool
}

// Followed reports the mark being pressed this frame.
func (r RemoteCursorResult) Followed() bool { return r.followed }

// RemoteCursor is somebody else's pointer, where they left it, with their
// name beside it.
//
// The name is in a chip hanging off the pointer rather than in a tooltip
// because a tooltip arrives on a hover and this arrives under a moving
// cursor: somebody typing fast never hovers anything, and a remote cursor
// that says nothing is a cursor you cannot follow.
//
// The colour comes from the name, so the same person is the same colour in
// every window and in every session. It is a hash of the name into the four
// accent steps rather than a caller's choice, because a caller that picked
// a colour per person would pick a different one in a different view and
// the two would not match.
func RemoteCursor(c *ui.Context, opts RemoteCursorOptions) RemoteCursorResult {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if opts.Name == "" {
		panic("files: RemoteCursor needs a Name; a pointer with nobody's on it cannot be followed")
	}
	h := opts.Height
	if h <= 0 {
		h = u * 5
	}
	ink := cursorInk(opts.Name, k)

	var res RemoteCursorResult
	mark := ui.ButtonBase(c).Absolute().Left(opts.X).Top(opts.Y).
		FillWidth().AlignItems(ui.Start).Role(ui.RoleNone).
		Label(opts.Name + ", " + doingWord(Viewer{Name: opts.Name, Doing: opts.Doing, Here: true}))
	mark.Children(func() {
		ui.Box(c).Width(u * 3.5).Shrink(0).Children(func() {
			// The pointer is a triangle drawn rather than an SVG, because
			// the library's set has no cursor in it and a remote cursor is
			// the one mark that must not be confused with any other: it sits
			// over somebody's text and has to read as a pointer.
			ui.Box(c).Width(u * 3.5).Height(h).Shrink(0).
				Draw(func(p *ui.Painter, r ui.Rect) {
					var path ui.Path
					path.MoveTo(r.X, r.Y).
						LineTo(r.X+r.W, r.Y+r.H*0.78).
						LineTo(r.X+r.W*0.5, r.Y+r.H*0.62).
						LineTo(r.X+r.W*0.36, r.Y+r.H)
					p.FillPath(&path, ink)
					internal.Rule(p, r, ink)
				})
		})
		chip := ui.Box(c).Padding(u*0.25, u*1.25).Radius(theme.PillRadius).
			Background(ink).Label(opts.Name)
		chip.Children(func() {
			// The ink on the chip is asked of the colour rather than taken
			// from a token: the same person may be a light accent in one
			// window and a dark one in another, and a label that vanishes
			// on one of them is worse than no colour at all.
			on := inkOn(ink)
			ui.Text(c, opts.Name).TextColor(on).Bold().
				FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
			if opts.Doing != "" {
				ui.Text(c, opts.Doing).TextColor(on.Alpha(0.75)).
					FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
			}
		})
		// The chip hangs off the point of the pointer: up and to the right,
		// which is where a name goes on every tool that draws somebody
		// else's cursor and the only place that does not cover the text the
		// pointer is over.
		chip.Margin(u*-0.5, 0, 0, u*1.25)
	})
	if mark.Clicked() {
		res.followed = true
	}
	res.Element = mark
	return res
}

// cursorInk is one of the four accent steps for a name. The name is hashed
// rather than ordered so that two people whose names are alphabetically
// adjacent do not get the same colour — which is what an index into a
// palette would do, and would make a list of names a list of stripes.
func cursorInk(name string, k theme.Tokens) ui.Color {
	var h uint32 = 2166136261
	for i := range len(name) {
		h ^= uint32(name[i])
		h *= 16777619
	}
	steps := []ui.Color{k.Accent, k.Success, k.Warning, k.TextMuted}
	return steps[h%uint32(len(steps))]
}

// inkOn is the black or white that reads on a colour. It is the plain
// brightness average rather than the WCAG measure: the question here is which
// of two inks is more visible on a swatch a person picked, and the two
// measures agree on that for every colour this library can produce.
func inkOn(bg ui.Color) ui.Color {
	if lum(bg) > 0.5 {
		return ui.RGB(0, 0, 0)
	}
	return ui.RGB(255, 255, 255)
}

// lum is a colour's brightness, from 0 to 1.
func lum(col ui.Color) float32 {
	return (float32(col.R) + float32(col.G) + float32(col.B)) / (3 * 255)
}
