package git

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// The marks a version-control interface is built from, in one file, because
// half of them are drawn two or three times and the other half are the reason
// two views of the same repository have to look alike.

// hashInk is what a hash is drawn in: the muted tone, at the monospaced size.
//
// A hash is a reference, not a heading. Drawing it in body ink makes it the
// loudest thing on a commit row, which is exactly backwards — the subject is
// what a person reads, and the hash is what they copy.
func hashInk(c *ui.Context) ui.Color { return core.Tokens(c).TextMuted }

// hashText is a hash as a graph and a table print it.
func hashText(c *ui.Context, hash string) *ui.Element {
	return ui.Text(c, hash).TextColor(hashInk(c)).
		FontSize(core.FontSize(c, theme.MonoSize)).SingleLine()
}

// commitDot is the node of a commit graph: a filled circle, hollow for a
// commit that is a tip — one nothing branches off.
//
// Hollow is the difference between a tip and everything else, and it is what
// lets somebody see where work has stopped without reading a single label.
func commitDot(c *ui.Context, filled bool) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	rad := u * 1.5
	return ui.Box(c).Size(rad*2, rad*2).Shrink(0).Role(ui.RoleNone).
		Draw(func(p *ui.Painter, r ui.Rect) {
			cx, cy := r.X+r.W/2, r.Y+r.H/2
			if filled {
				internal.Dot(p, cx, cy, rad, k.Text)
				return
			}
			// The ring is drawn over the window's own background so it reads
			// as a hole in a line rather than as a line drawn twice.
			internal.Dot(p, cx, cy, rad, k.Background)
			internal.Ring(p, cx, cy, rad-u*0.45, u*0.9, k.Text)
		})
}

// refChip is a branch or tag name beside a commit. It is a pill rather than
// plain text because a name among subjects has to be findable by shape, not
// only by reading: a graph with twenty subjects and three refs should show
// the three at a glance.
func refChip(c *ui.Context, name string, tone core.Severity) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	bg, fg := tone.Pair(k)
	return ui.Box(c).Padding(u*0.25, u*1.25).Radius(theme.PillRadius).
		Background(bg).Label(name).Children(func() {
		ui.Text(c, name).TextColor(fg).
			FontSize(core.FontSize(c, theme.CaptionSize)).Bold().SingleLine()
	})
}

// pickRow is a row of a list that takes a press and says which one it was.
//
// It is a ButtonBase rather than a Box because that is the only exported way
// to get a row that takes presses and the keyboard. The cost is a stop of Tab
// per row, which inside a list is the point rather than the problem.
func pickRow(c *ui.Context, name string, chosen bool, build func()) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	row := ui.ButtonBase(c).FillWidth().Justify(ui.Start).AlignItems(ui.Center).
		Gap(u*1.5).Padding(u*0.75, u*1.5).Radius(theme.SmallRadius).
		TextColor(k.Text).Label(name).Role(ui.RoleListItem)
	switch {
	case chosen:
		row.Background(k.SurfaceHover)
	case row.Hovered():
		row.Background(k.SurfaceHover)
	}
	row.Children(build)
	return row
}

// press is a small pill that takes a press — the "commit", "amend", "apply"
// buttons a git panel wears. It is a ButtonBase so the caller reads the press
// off the element, which is how every other button in the library works.
func press(c *ui.Context, label string, primary bool, tone core.Severity) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	bg, fg := k.Surface, k.Text
	switch {
	case primary:
		bg, fg = k.Fill, k.OnFill
	case tone != core.Neutral:
		bg, fg = tone.Pair(k)
	}
	btn := ui.ButtonBase(c).Height(core.ControlHeight(c)).Radius(theme.PillRadius).
		Padding(0, u*3).Background(bg).TextColor(fg).Label(label).Tooltip(label)
	btn.Children(func() {
		ui.Text(c, label).FontSize(core.FontSize(c, theme.RowSize))
	})
	return btn
}

// gutter is the width a two-column number gutter takes: enough for the widest
// line number a diff of any size reaches, so the text beside it does not jump
// sideways as the user scrolls.
func gutter(c *ui.Context, lines []DiffLine) float32 {
	u := core.Density(c).Unit()
	widest := 0
	for _, l := range lines {
		widest = max(widest, digits(max(l.OldNo, l.NewNo)))
	}
	return float32(widest)*u*0.8 + u*2.5
}

// digits is how many characters n prints as, which is how a gutter is sized
// without measuring a font.
func digits(n int) int {
	n = max(n, 0)
	switch {
	case n == 0:
		return 1
	case n < 10:
		return 1
	case n < 100:
		return 2
	case n < 1000:
		return 3
	default:
		return 4
	}
}

// padNo is a line number written into a fixed width, so two numbers in the
// gutter are right-aligned against the same edge.
func padNo(n, width int) string {
	s := itoa(n)
	if n <= 0 {
		return ""
	}
	for len(s) < width {
		s = " " + s
	}
	return s
}

// noRows is what a git list says when there is nothing in it. Every list here
// says it through core.Msg so a window can reword it, as it can every other
// word the library owns.
func noRows(c *ui.Context, key, def string) string { return core.Msg(c, key, def) }
