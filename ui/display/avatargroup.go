package display

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// DefaultMaxFaces is how many faces AvatarGroup shows before the rest become
// a count. Four leaves room for the "+7" pill beside them inside a card's
// header; five is the point at which the group is wider than the card.
const DefaultMaxFaces = 4

// AvatarGroupOptions configure an AvatarGroup.
type AvatarGroupOptions struct {
	// Max is how many faces draw before the rest become a count; zero gives
	// DefaultMaxFaces.
	Max int
	// Size is a face's diameter in DIPs; zero gives the card-sized one.
	Size float32
	// More names the overflow control for assistive technology. It is
	// required when the group has a rest at all — a "+7" that reads only as
	// "+7" is a puzzle — and unused when it does not.
	More string
	// Tone tints the overflow pill, for a group whose rest is an escalation.
	Tone core.Severity
}

// AvatarGroupResult carries an AvatarGroup and what was pressed on it.
type AvatarGroupResult struct {
	// Element is the whole group.
	Element *ui.Element
	// expanded reports a press of the overflow control this frame.
	expanded bool
	// overflow reports how many faces the pill stands for.
	overflow int
}

// Expanded reports a press of the "+n" control, which is what a caller hangs
// its "everyone on this callback" sheet off. The group does not open it: the
// list behind that sheet is somebody else's data, and the caller's business
// when to ask for it.
func (r AvatarGroupResult) Expanded() bool { return r.expanded }

// Overflow returns how many faces are behind the "+n" pill — 0 when the whole
// group is drawn. A test can say what is missing this way without counting
// circles.
func (r AvatarGroupResult) Overflow() int { return r.overflow }

// AvatarGroup is a set of faces with somewhere to go: draw a few, and let a
// person open the rest.
//
// It is not AvatarCluster. AvatarCluster answers "who is on this callback" —
// a caption beside a header, static, never pressed, with the count spelled out
// in words. AvatarGroup answers "how many people are in here" and is a
// control: its overflow pill is pressable, reports through Expanded, and the
// caller's sheet lists every name. Two components because the questions are
// two: one is a figure under a heading, the other is a door.
//
// The whole group is clickable as well as the pill, because a person aiming
// at four small circles should not have to hit the last three pixels of one.
func AvatarGroup(c *ui.Context, names []string, opts AvatarGroupOptions) AvatarGroupResult {
	k, u := core.Tokens(c), core.Density(c).Unit()
	limit := opts.Max
	if limit <= 0 {
		limit = DefaultMaxFaces
	}
	side := opts.Size
	if side <= 0 {
		side = u * 7
	}
	pillBg, pillFg := opts.Tone.Pair(k)

	var r AvatarGroupResult
	shown := min(len(names), limit)
	rest := len(names) - shown

	label := opts.More
	if rest > 0 && label == "" {
		panic("display: AvatarGroup needs More to name the faces behind the count")
	}

	r.Element = ui.Row(c).AlignItems(ui.Center).Label(groupName(c, names, label)).Children(func() {
		for i := range shown {
			// The 2px ring in the window's colour is what separates one
			// circle from the next where they overlap; without it the group
			// is one lobed shape.
			box := ui.Box(c).Size(side, side).Radius(side/2).Background(k.Surface).
				Border(2, k.Background).Center().Label(names[i])
			if i > 0 {
				box.Margin(0, 0, 0, -side*0.22)
			}
			box.Children(func() {
				ui.Text(c, internal.Initials(names[i])).TextColor(k.Text).
					FontSize(theme.MonoSize).Bold()
			})
		}
		if rest > 0 {
			pill := ui.Box(c).Size(side, side).Radius(side/2).
				Background(pillBg).Border(2, k.Background).
				Center().Label(label)
			if shown > 0 {
				pill.Margin(0, 0, 0, -side*0.22)
			}
			pill.Children(func() {
				ui.Text(c, "+"+internal.Commas(rest)).TextColor(pillFg).
					FontSize(core.FontSize(c, theme.MonoSize)).Bold()
			})
			if pill.Clicked() {
				r.expanded = true
			}
		}
	})

	// Pressed anywhere on the group is a press on the group: the faces are
	// 28 DIPs across and the pill is the last 28 of them, which is a small
	// target for the thing people most want to open.
	if r.Element.Clicked() {
		r.expanded = true
	}
	r.overflow = rest
	return r
}

// groupName is what the group is called out loud: the overflow control's own
// name when there is a rest to open, and otherwise how many people are in it,
// so a group of two that fits is still not an unnamed row of circles.
func groupName(c *ui.Context, names []string, more string) string {
	if more != "" {
		return more
	}
	n := len(names)
	return internal.Commas(n) + " " + internal.Plural(n, "person", "people") + " " +
		core.Msg(c, "avatargroup.on", "on this callback")
}
