// Package display holds read-only components: avatars, pills, meters and the
// figures a screen shows rather than accepts.
package display

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// MaxAvatars is how many faces AvatarCluster draws before switching to a
// count. Five fits beside a header's other content without crowding it.
const MaxAvatars = 5

// Avatar draws a person's initials in a circle, standing in for a photograph
// the caller has not loaded. An empty name draws the circle with no initials,
// which reads as a placeholder rather than as a mistake.
func Avatar(c *ui.Context, name string) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	side := u * 10
	box := ui.Box(c).Size(side, side).Radius(side / 2).Background(k.Surface).
		Center().Label(name)
	if in := internal.Initials(name); in != "" {
		box.Children(func() {
			ui.Text(c, in).TextColor(k.Text).FontSize(theme.MonoSize).Bold()
		})
	}
	return box
}

// AvatarCluster draws up to MaxAvatars overlapping circles followed by a
// count, the way a team is summarised in a header. It does not repeat past
// MaxAvatars: the number carries the rest. A total of zero or fewer draws
// nothing at all.
func AvatarCluster(c *ui.Context, names []string, total int) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if total <= 0 {
		return ui.Box(c)
	}
	side := u * 7.5
	label := internal.Plural(total, "person", "people")
	return ui.Row(c).AlignItems(ui.Center).Label(label).Children(func() {
		shown := min(len(names), MaxAvatars)
		for i := range shown {
			box := ui.Box(c).Size(side, side).Radius(side/2).Background(k.Surface).
				Border(2, k.Background).Center().Label(names[i])
			if i > 0 {
				box.Margin(0, 0, 0, -side*0.22) // the overlap
			}
			box.Children(func() {
				ui.Text(c, internal.Initials(names[i])).
					TextColor(k.Text).FontSize(theme.MonoSize).Bold()
			})
		}
		if total > shown {
			ui.Text(c, fmtInt(total)+" "+label).TextColor(k.TextMuted).FontSize(theme.MetaSize)
		}
	})
}

// Meter draws a run of n bars, filled to level, the way a priority or a signal
// strength reads at a glance. Level is clamped to n, so a caller that ranks
// past the top gets the full bar rather than a broken one.
func Meter(c *ui.Context, n, level int, severity core.Severity) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if n <= 0 {
		panic("display: Meter needs a positive n")
	}
	_, fg := severity.Pair(k)
	// A fixed width: the meter sits inside a pill beside its label, so it
	// must not grow to fill the row.
	return ui.Box(c).Width(u * 7).Height(u * 3).Shrink(0).
		Draw(func(p *ui.Painter, r ui.Rect) {
			internal.Meter(p, r, clamp(level, 0, n), n, fg, k.Border)
		})
}

// fmtInt renders a count without pulling in strconv at every call site.
func fmtInt(n int) string { return internal.Commas(n) }

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
