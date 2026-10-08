package feedback

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// ActivityItem is one thing that happened.
type ActivityItem struct {
	// Title is what happened, in the present tense: "Callback assigned",
	// "Sync failed". It is required — a feed row with no title is a dot.
	Title string
	// Detail is the line under it: who, which one, why.
	Detail string
	// Time is when, already formatted by the caller: a feed that formats its
	// own times cannot be shown in a window whose locale is not the
	// library's.
	Time string
	// Severity marks the row, and carries through to the row's dot and the
	// action's colour where a Result reports one. Zero is Neutral, which is
	// right for most of what a feed carries.
	Severity core.Severity
}

// ActivityFeedOptions configure an ActivityFeed.
type ActivityFeedOptions struct {
	// Items are the entries, newest first. At least one is required: a feed
	// with nothing in it is not a feed, and an empty one silently rendering
	// nothing hides a bug behind a plausible-looking screen.
	Items []ActivityItem
	// Rule draws a hairline between rows. It is on by default because a
	// feed of two-line rows without them reads as one column of prose.
	Rule bool
	// NoRule turns them off for a feed whose rows are separated by their own
	// avatars or padding.
	NoRule bool
	// Width bounds the feed; zero lets it fill its parent.
	Width float32
}

// ActivityFeed is what happened, most recent first, as a column of rows each
// marked by a dot and a time.
//
// It is the answer to "what has this application been doing while I was
// away", which is why it carries no filter and no paging: a caller that has
// those builds the list, and the feed gives the list a shape.
func ActivityFeed(c *ui.Context, opts ActivityFeedOptions) *ui.Element {
	if len(opts.Items) == 0 {
		panic("feedback: ActivityFeed needs at least one item; an empty feed " +
			"would hide a bug behind a plausible-looking screen")
	}
	u := core.Density(c).Unit()

	feed := ui.Column(c).FillWidth().Role(ui.RoleList)
	if opts.Width > 0 {
		feed.Width(opts.Width)
	}
	rule := !opts.NoRule

	feed.Children(func() {
		for i, item := range opts.Items {
			if i > 0 && rule {
				layout.Divider(c, layout.DividerOptions{})
			}
			activityRow(c, item, u)
		}
	})
	return feed
}

// activityRow draws one entry: a dot in the gutter, the title and detail in
// the middle, and the time at the end.
//
// The gutter is a fixed width so every dot lines up down the column. A dot
// that moved with its title would make the column a ragged edge, and the dots
// are the only thing that lines up at all.
func activityRow(c *ui.Context, item ActivityItem, u float32) {
	k := core.Tokens(c)
	dot := u * 2.25

	ui.Row(c).FillWidth().AlignItems(ui.Start).Gap(u*2.5).
		Padding(u*2.25, 0, u*2.25, 0).Label(item.Title).Children(func() {
		ui.Box(c).Size(dot, dot).Radius(dot/2).Shrink(0).
			Background(severityInk(k, item.Severity)).
			// The dot sits on the first line of the title rather than
			// centred: it belongs to the title, and the detail below it is
			// the same event's explanation.
			Margin(u*0.75, 0, 0, 0)
		ui.Column(c).Grow(1).Gap(u * 0.5).Children(func() {
			ui.Text(c, item.Title).TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).Bold()
			if item.Detail != "" {
				ui.Text(c, item.Detail).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.MetaSize))
			}
		})
		if item.Time != "" {
			// A fixed slot at the end, so a column of times is a column and
			// not a ragged right edge that shifts as titles change length.
			ui.Text(c, item.Time).TextColor(k.TextFaint).Shrink(0).
				FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
		}
	})
}
