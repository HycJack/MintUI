package git

import (
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/data"
	"github.com/HycJack/MintUI/ui/theme"
)

// Stash is one entry of the stash: a commit nobody made on a branch, kept
// under a message somebody wrote at the time.
type Stash struct {
	// Name is the entry's own name, "stash@{0}", which is what a person
	// types into a command and so what a button that runs one needs.
	Name string
	// Message is what the stash was saved with.
	Message string
	// Branch is the branch it was taken from.
	Branch string
	// When is when it was taken, already formatted.
	When string
}

// StashListOptions configure a StashList.
type StashListOptions struct {
	// Height is the height of the list; required.
	Height float32
	// Width is the width of the table; zero measures the window.
	Width float32
	// Selected is the row the keys move from, -1 for none.
	Selected *int
	// Choice is the set of chosen rows when the list chooses several.
	Choice *data.Selectable
	// State is where the list keeps its place between frames.
	State *ui.ListState
	// Empty draws instead of the rows when there are none.
	Empty func()
}

// StashListResult carries a StashList.
type StashListResult struct {
	// Element is the whole table. The chosen stash is the caller's Selected
	// pointer, as it is for every table in the library.
	Element *ui.Element
}

// StashList is what the stash holds: the entries, newest first, with the
// branch each was taken from.
//
// Applying and dropping are the two things a person does here and neither is
// done by the list: applying rewrites the working tree, which is not
// something a painted row may do. The list reports which row was pressed and
// the caller runs the command.
//
// The entry's own name is in the table rather than left implicit, because
// "stash@{2}" is what every error message git prints about a stash says, and
// a list that does not show it makes those messages unmatchable.
func StashList(c *ui.Context, stashes []Stash, opts StashListOptions) StashListResult {
	k := core.Tokens(c)
	if opts.Height <= 0 {
		panic("git: StashList needs a Height; a stash list with no height is every stash there is")
	}

	cols := []data.Column{
		{ID: "name", Title: "Stash", Width: 116},
		{ID: "message", Title: "Message", Share: 3, Sortable: true},
		{ID: "branch", Title: "Branch", Width: 140, Sortable: true},
		{ID: "when", Title: "When", Width: 110},
	}

	tr := data.DataTable(c, data.DataTableOptions{
		Columns: cols, Rows: len(stashes), Height: opts.Height, Width: opts.Width,
		Selected: opts.Selected, Choice: opts.Choice, State: opts.State,
		Key: func(row int) any { return stashes[row].Name },
		Cell: func(row, col int) {
			s := stashes[row]
			switch cols[col].ID {
			case "name":
				hashText(c, s.Name)
			case "message":
				ui.Text(c, s.Message).Grow(1).Ellipsis(s.Message).
					FontSize(core.FontSize(c, theme.BodySize)).SingleLine()
			case "branch":
				ui.Text(c, s.Branch).TextColor(k.TextMuted).SingleLine().
					FontSize(core.FontSize(c, theme.MetaSize))
			case "when":
				ui.Text(c, s.When).TextColor(k.TextMuted).SingleLine().
					FontSize(core.FontSize(c, theme.MetaSize))
			}
		},
		Label: func(row int) string {
			s := stashes[row]
			return s.Name + ": " + s.Message + ", from " + s.Branch
		},
		CellLabel: func(row, col int) string {
			s := stashes[row]
			if cols[col].ID == "message" {
				return s.Message
			}
			return s.Name + ": " + s.Message
		},
		Empty: func() {
			if opts.Empty != nil {
				opts.Empty()
				return
			}
			ui.Text(c, noRows(c, "git.noStashes", core.Def("No stashes"))).
				TextColor(k.TextMuted).FontSize(core.FontSize(c, theme.RowSize))
		},
	})
	return StashListResult{Element: tr.Element}
}

// Tag is one tag and what it points at.
type Tag struct {
	// Name is the tag's name, without refs/tags/.
	Name string
	// Subject is the message of the commit it points at.
	Subject string
	// Author is who wrote that commit.
	Author string
	// When is when it was tagged, already formatted.
	When string
	// Annotated reports that the tag carries its own message, which is what
	// makes it worth a tag at all.
	Annotated bool
}

// TagListOptions configure a TagList.
type TagListOptions struct {
	// Height is the height of the list; required.
	Height float32
	// Width is the width of the table; zero measures the window.
	Width float32
	// Selected is the tag the keys move from, by name. A tag is named
	// rather than numbered because a tag list is sorted by date and the
	// number of a tag stops meaning anything the moment it is sorted.
	Selected *string
	// ShowTags is the prefix that marks a name as a tag in the list.
	// Empty drops it, for a list that is nothing but tags.
	ShowPrefix bool
	// State is where the list keeps its place between frames.
	State *ui.ListState
	// Empty draws instead of the rows when there are none.
	Empty func()
}

// TagListResult carries a TagList.
type TagListResult struct {
	// Element is the whole table.
	Element *ui.Element
}

// TagList is the tags, newest first, with the commit each one points at.
//
// An annotated tag and a lightweight one look different because they are:
// the annotated one carries a message and the lightweight one carries
// nothing but a name. Drawing them alike would say that a release somebody
// wrote a paragraph for is the same kind of thing as a bookmark.
func TagList(c *ui.Context, tags []Tag, opts TagListOptions) TagListResult {
	k := core.Tokens(c)
	if opts.Height <= 0 {
		panic("git: TagList needs a Height; a tag list with no height is every tag there is")
	}

	cols := []data.Column{
		{ID: "name", Title: "Tag", Width: 160, Sortable: true},
		{ID: "subject", Title: "Points at", Share: 3, Sortable: true},
		{ID: "when", Title: "When", Width: 110},
	}

	tr := data.DataTable(c, data.DataTableOptions{
		Columns: cols, Rows: len(tags), Height: opts.Height, Width: opts.Width,
		State: opts.State,
		Key:   func(row int) any { return tags[row].Name },
		Cell: func(row, col int) {
			tg := tags[row]
			switch cols[col].ID {
			case "name":
				name := tg.Name
				if opts.ShowPrefix {
					name = "tag: " + name
				}
				ui.Box(c).Label(name).Children(func() {
					refChip(c, name, tagTone(tg.Annotated))
				})
			case "subject":
				ui.Text(c, tg.Subject).Grow(1).Ellipsis(tg.Subject).
					FontSize(core.FontSize(c, theme.BodySize)).SingleLine()
			case "when":
				ui.Text(c, tg.When).TextColor(k.TextMuted).SingleLine().
					FontSize(core.FontSize(c, theme.MetaSize))
			}
		},
		Label: func(row int) string {
			return tags[row].Name + ": " + tags[row].Subject
		},
		CellLabel: func(row, col int) string {
			if cols[col].ID == "subject" {
				return tags[row].Subject
			}
			return tags[row].Name + ": " + tags[row].Subject
		},
		Empty: func() {
			if opts.Empty != nil {
				opts.Empty()
				return
			}
			ui.Text(c, noRows(c, "git.noTags", core.Def("No tags"))).
				TextColor(k.TextMuted).FontSize(core.FontSize(c, theme.RowSize))
		},
	})
	return TagListResult{Element: tr.Element}
}

// tagTone is what a tag's chip wears: a warning tint for an annotated tag,
// because it is a thing somebody wrote rather than a pointer somebody left.
func tagTone(annotated bool) core.Severity {
	if annotated {
		return core.Warning
	}
	return core.Accent
}

// PullRequest is a request to merge one branch into another, as a card shows
// it.
type PullRequest struct {
	// Number is the request's number, which is how a person refers to it and
	// how a URL names it.
	Number int
	// Title is its one-line title.
	Title string
	// Author is who opened it.
	Author string
	// From and To are the two branches.
	From, To string
	// State is where it is: open, merged, or closed. It is a word rather
	// than an enum because every forge spells this one its own way and a
	// window that has to be reworded for each of them should be reworded in
	// copy rather than in code.
	State string
	// Comments and Reviews are the counts on it.
	Comments, Reviews int
	// ChecksPassed and ChecksTotal are the continuous-integration counts,
	// and ChecksFailed is how many of them are red — which is a separate
	// number because "20 of 21 passed" and "21 of 21 passed" are the same
	// sentence and only one of them is good news.
	ChecksPassed, ChecksTotal, ChecksFailed int
	// Draft says it is not ready to be looked at.
	Draft bool
}

// PullRequestCardOptions configure a PullRequestCard.
type PullRequestCardOptions struct {
	// Height is the least the card is high; zero is the library's own.
	Height float32
	// Actions draws the buttons along the bottom, where a caller that wants
	// "Merge", "Close" and "Convert to draft" puts them.
	Actions func()
	// Label names the card for assistive technology; empty uses the title.
	Label string
}

// PullRequestCardResult carries a PullRequestCard and whether it was pressed.
type PullRequestCardResult struct {
	// Element is the whole card.
	Element *ui.Element
	// pressed reports the card being pressed this frame — the open-it action,
	// which is the whole card when the card is in a list of them.
	pressed bool
}

// Pressed reports the card being pressed this frame.
func (r PullRequestCardResult) Pressed() bool { return r.pressed }

// PullRequestState is the tone a pull request's state is drawn in, which is
// the one thing about a request that has to be readable before its title is.
//
// It is a function rather than a field because "Merged" and "Closed" are
// both finished and are not the same news: one is in the target branch and
// the other is not. Success and neutral say exactly that.
func PullRequestState(state string) core.Severity {
	switch strings.ToLower(strings.TrimSpace(state)) {
	case "merged":
		return core.Success
	case "closed", "draft":
		// A closed request and a draft are both finished for now and both
		// are quiet: neither is waiting on anybody in a way that should
		// light up a list of cards.
		return core.Neutral
	}
	return core.Accent
}

// PullRequestCard is one pull request: the number, the title, the two
// branches, who opened it, and what is waiting on it.
//
// The whole card is the press, because in a list of them the card is the
// only thing that is the size of a thing you click, and a title with a
// separate small "open" beside it is two targets where there is one obvious
// one. The buttons underneath are children of the card and take the press
// before the card does.
func PullRequestCard(c *ui.Context, pr PullRequest, opts PullRequestCardOptions) PullRequestCardResult {
	k, u := core.Tokens(c), core.Density(c).Unit()

	var res PullRequestCardResult
	// Column, not the Row a ButtonBase is: a card stacks its number line, its
	// title, its meta line and its actions down the page, and a Row put
	// those four side by side — which put the action buttons on the same
	// line as the title and squeezed the title into two words a line.
	card := ui.ButtonBase(c).Column().FillWidth().Justify(ui.Start).AlignItems(ui.Stretch).
		Padding(u * 2).Radius(theme.CardRadius).Gap(u).Background(k.Surface).
		Label(prLabel(pr, opts)).Role(ui.RoleNone)
	card.Children(func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.5).Children(func() {
			ui.Text(c, "#"+itoa(pr.Number)).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.MetaSize)).SingleLine()
			refChip(c, prWord(c, pr), PullRequestState(pr.State))
			if pr.Draft {
				refChip(c, core.Msg(c, "git.draft", core.Def("Draft")), core.Neutral)
			}
			ui.Box(c).Grow(1)
			ui.Text(c, pr.Author).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.MetaSize)).SingleLine()
		})
		ui.Text(c, pr.Title).FillWidth().Ellipsis(pr.Title).
			FontSize(core.FontSize(c, theme.BodySize)).Bold()
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.5).Children(func() {
			ui.Text(c, pr.From+" → "+pr.To).TextColor(k.TextMuted).Grow(1).SingleLine().
				FontSize(core.FontSize(c, theme.CaptionSize))
			if pr.Comments > 0 {
				ui.Text(c, core.Msg(c, "git.comments", core.Def("comments"))+": "+itoa(pr.Comments)).
					TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
			}
			if pr.Reviews > 0 {
				ui.Text(c, core.Msg(c, "git.reviews", core.Def("reviews"))+": "+itoa(pr.Reviews)).
					TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
			}
			checks(c, pr)
		})
		if opts.Actions != nil {
			opts.Actions()
		}
	})
	if opts.Height > 0 {
		card.MinHeight(opts.Height)
	}
	if card.Clicked() {
		res.pressed = true
	}
	res.Element = card
	return res
}

// checks is the continuous-integration line: "21 checks, 1 failing". A red
// count is drawn separately from the total because a total that includes a
// failure is the shape people learn to read past.
func checks(c *ui.Context, pr PullRequest) {
	if pr.ChecksTotal <= 0 {
		return
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	ui.Row(c).Gap(u).AlignItems(ui.Center).Children(func() {
		word := itoa(pr.ChecksPassed) + core.Def("/") + itoa(pr.ChecksTotal)
		if pr.ChecksFailed > 0 {
			_, ink := core.Danger.Pair(k)
			ui.Text(c, itoa(pr.ChecksFailed)+core.Msg(c, "git.failing", core.Def(" failing"))).
				TextColor(ink).Bold().
				FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
		}
		ui.Text(c, word).TextColor(k.TextMuted).
			FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
	})
}

// prWord is the request's state as a word, defaulting to open, because every
// card in a list of them has one and an empty chip says less than "open".
func prWord(c *ui.Context, pr PullRequest) string {
	if pr.State != "" {
		return pr.State
	}
	return core.Msg(c, "git.open", core.Def("Open"))
}

// prLabel is the card's name for assistive technology and for a test to
// find: the number, the title, and where it wants to go.
func prLabel(pr PullRequest, opts PullRequestCardOptions) string {
	if opts.Label != "" {
		return opts.Label
	}
	return "#" + itoa(pr.Number) + " " + pr.Title + ", " + pr.From + " into " + pr.To
}

// ReviewNote is one comment on one line of a diff. It is called a note
// rather than a ReviewComment because that is the component that draws it.
type ReviewNote struct {
	// Author is who left it.
	Author string
	// When is when, already formatted.
	When string
	// Body is what they wrote.
	Body string
	// Path and Line say where in the file it hangs. A comment with nowhere
	// to hang is a general comment, which is a different thing and is
	// drawn as one.
	Path string
	// Line is the line number it is on, 0 for a comment about the file.
	Line int
	// Side is "old" or "new", which side of a diff the line is on. Empty
	// is the new side, which is where a comment nearly always is.
	Side string
	// Resolved says somebody has marked the thread done. A resolved thread
	// stays visible rather than disappearing, because a reader who was not
	// there when it was resolved needs to know it was.
	Resolved bool
	// Reactions are the emoji and their counts on the comment.
	Reactions []Reaction
}

// Reaction is one emoji on a comment or a review.
type Reaction struct {
	// Emoji is the character itself: "👍", "🎉", "+1". It is text, not a
	// glyph from the icon set, because a reaction is a thing somebody
	// picked out of the ones everybody knows.
	Emoji string
	// Count is how many of them there are.
	Count int
	// Mine reports that the reader is one of the ones who left it, which
	// is what gives it a filled face rather than an outline.
	Mine bool
}

// ReviewCommentOptions configure a ReviewComment.
type ReviewCommentOptions struct {
	// Width is the width of the comment; zero fills the row it is in.
	Width float32
	// Action draws what the reader can do to the thread — resolve, react,
	// reply — underneath the body.
	Action func()
	// Reactions shows the emoji row. Off when the comment has none, which
	// is most of them, rather than drawing an empty line.
	Reactions bool
}

// ReviewCommentResult carries a ReviewComment and what was done with it.
type ReviewCommentResult struct {
	// Element is the whole comment.
	Element *ui.Element
	// resolved reports the resolve button being pressed this frame.
	resolved bool
	// reacted is the emoji reacted with this frame, empty for none.
	reacted string
}

// Resolved reports the resolve control being pressed this frame.
func (r ReviewCommentResult) Resolved() bool { return r.resolved }

// Reacted returns the emoji the reader reacted with this frame, empty for
// none. An emoji rather than an index, because a thread's reactions are the
// reader's to send and an index would be wrong the moment somebody else
// reacts first.
func (r ReviewCommentResult) Reacted() string { return r.reacted }

// ReviewComment is one person's note on one line of somebody else's diff.
//
// It hangs off the line it is about — the path and the line number in a
// muted line above the body — because a review comment without its line is
// advice about a file rather than about a change, and the whole of reviewing
// is the line.
//
// A resolved thread is dimmed and marked rather than hidden. Somebody reading
// the review tomorrow was not there when it was resolved, and a thread that
// quietly vanished is a question they cannot ask.
func ReviewComment(c *ui.Context, rc ReviewNote, opts ReviewCommentOptions) ReviewCommentResult {
	k, u := core.Tokens(c), core.Density(c).Unit()

	ink := k.Text
	if rc.Resolved {
		// Resolved is not danger: it is a thread that has been dealt with,
		// and drawing it in the alarm colour would make a finished review
		// look like a broken one.
		ink = k.TextMuted
	}

	var res ReviewCommentResult
	card := ui.Column(c).FillWidth().Gap(u).Padding(u * 1.5).
		Radius(theme.SmallRadius).Background(k.Surface)
	if opts.Width > 0 {
		card.Width(opts.Width).Shrink(0)
	}
	card.Children(func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.5).Children(func() {
			ui.Text(c, rc.Author).TextColor(ink).Bold().
				FontSize(core.FontSize(c, theme.RowSize)).SingleLine()
			ui.Text(c, rc.When).TextColor(k.TextFaint).Grow(1).
				FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
			if rc.Resolved {
				refChip(c, core.Msg(c, "git.resolved", core.Def("Resolved")), core.Success)
			}
		})
		if rc.Path != "" {
			where := rc.Path
			if rc.Line > 0 {
				side := rc.Side
				if side == "" {
					side = core.Def("new")
				}
				where += core.Def(":") + itoa(rc.Line) + core.Def(" (") + side + core.Def(")")
			}
			ui.Text(c, where).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
		}
		ui.Text(c, rc.Body).TextColor(ink).FillWidth().
			FontSize(core.FontSize(c, theme.BodySize))

		if opts.Reactions && len(rc.Reactions) > 0 {
			ui.Row(c).FillWidth().Gap(u * 0.75).Children(func() {
				for _, r := range rc.Reactions {
					r := r
					btn := reactionChip(c, r)
					if btn.Clicked() {
						res.reacted = r.Emoji
					}
				}
			})
		}
		if opts.Action != nil {
			opts.Action()
		}
	})
	res.Element = card
	return res
}

// reactionChip is one emoji and its count. The reader's own reaction is
// filled and outlined is everybody else's, which is the only way to see at a
// glance which of them you are already in.
func reactionChip(c *ui.Context, r Reaction) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	bg, fg := k.Background, k.TextMuted
	if r.Mine {
		bg, fg = k.AccentBg, k.AccentText
	}
	chip := ui.ButtonBase(c).Padding(u*0.25, u*1.25).Radius(theme.PillRadius).
		Background(bg).TextColor(fg).
		Label(r.Emoji + ", " + itoa(r.Count) + " reactions").Role(ui.RoleNone)
	chip.Children(func() {
		ui.Text(c, r.Emoji).FontSize(core.FontSize(c, theme.RowSize))
		ui.Text(c, itoa(r.Count)).TextColor(fg).
			FontSize(core.FontSize(c, theme.CaptionSize))
	})
	return chip
}
