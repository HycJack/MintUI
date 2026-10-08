package messaging

import (
	"strings"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/data"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/overlay"
	"github.com/HycJack/MintUI/ui/theme"
)

// MailFolder is which of the boxes a list is showing.
type MailFolder string

const (
	// FolderInbox is what has arrived.
	FolderInbox MailFolder = "inbox"
	// FolderSent is what was written.
	FolderSent MailFolder = "sent"
	// FolderDrafts is what was not finished, which is a folder rather than a
	// filter because a draft has a different status and the row has to show
	// it.
	FolderDrafts MailFolder = "drafts"
	// FolderArchive is what has been put away.
	FolderArchive MailFolder = "archive"
	// FolderSpam is what somebody else decided.
	FolderSpam MailFolder = "spam"
)

// Message is one mail.
type Message struct {
	// ID identifies it and is required: a mailbox is sorted, filtered and
	// reloaded, and a row the caller cannot recognise afterwards is a row it
	// cannot act on.
	ID string
	// From is who it is from, and To who it is for — a sent message's To is
	// what the reader is scanning for and an inbox one's From is, and the
	// column therefore names whichever of the two the folder puts first.
	From string
	To   string
	// Subject is the line the list shows as the message's own words.
	Subject string
	// Preview is the first lines of the body.
	Preview string
	// Body is the whole thing, which the reader opens. Empty is a message
	// whose body has not been fetched, and the list must not pretend
	// otherwise.
	Body string
	// When is when it arrived, as the caller writes it.
	When string
	// HasAttachments and AttachmentCount are separate because "one" and "four"
	// are different information: the first says there is something else here,
	// the second says how much of the reading it will be.
	HasAttachments  bool
	AttachmentCount int
	// Unread is the caller's flag and is what the list marks the row with.
	Unread bool
	// Important is the caller's flag too. It is not derived from an
	// attachment count or a capital letter in the subject: deciding what is
	// important is the job of whatever rule the application has, and only the
	// application knows the rule.
	Important bool
	// Folder is which box it is in.
	Folder MailFolder
	// SnoozedUntil is when it will come back, and nil for one that will not.
	SnoozedUntil *time.Time
}

// ── the list ───────────────────────────────────────────────────────────────

// mailColumns are the list's columns. The subject takes the leftover width and
// the rest are fixed, because the numbers in a mailbox — how many attachments,
// how long ago — must not be the things that wrap when the window narrows.
var mailColumns = []data.Column{
	{Title: "From", ID: "from", Width: 180},
	{Title: "Subject", ID: "subject", Share: 1},
	{Title: "Attachments", ID: "files", Width: 88, Align: ui.End},
	{Title: "When", ID: "when", Width: 104, Align: ui.End},
}

// MailListOptions configure a MailList.
type MailListOptions struct {
	// Messages are the caller's, in the order they should be shown.
	Messages []Message
	// Folder decides which messages are drawn and which column is the
	// primary one: an inbox scans senders and a sent folder scans recipients,
	// and putting the wrong column first makes a sent folder unreadable.
	Folder MailFolder
	// Selected is the message the keys move from, as an index into the
	// messages that passed the folder filter; -1 for none.
	Selected *int
	// Sort is the column the rows are ordered by, in the caller's state.
	Sort *data.Sort
	// State and Scroll keep the table's place between frames.
	State  *ui.ListState
	Scroll *ui.ScrollState
	// Height is the table's height. It is required, as it is for every list in
	// this library: a table with no height grows to fit every row.
	Height float32
	// Query filters on subject, sender, recipient and body text.
	Query *string
	// UnreadOnly leaves out the read ones, for the unread tab every mailbox
	// has.
	UnreadOnly bool
	// Title heads the panel; empty takes the folder's own name.
	Title string
}

// MailListResult carries a MailList.
type MailListResult struct {
	// Element is the list.
	Element *ui.Element
	// shown is how many rows passed the folder and the query.
	shown int
	// unread is how many of those are unread, which is what the tab's count
	// says and the only thing a reader can act on.
	unread int
}

// Shown is how many messages passed the filters.
func (r MailListResult) Shown() int { return r.shown }

// Unread is how many of the shown messages are unread. It is reported rather
// than left for the caller to count because the tab that carries this number
// sits outside the table and would otherwise be counting a filtered list it
// does not have.
func (r MailListResult) Unread() int { return r.unread }

// MailList is the column of messages in a folder.
//
// It is [data.DataTable] and not a bespoke list, for the same reason the
// member list is: a mailbox is records in columns, the columns have to keep
// their widths so a date does not wrap, and the folder is a filter on a
// shared table rather than a second table.
func MailList(c *ui.Context, opts MailListOptions) MailListResult {
	if opts.Selected == nil {
		panic("messaging: MailList needs a Selected message to point at; it owns no mailbox " +
			"of its own")
	}
	if opts.Height <= 0 {
		panic("messaging: MailList needs a Height; a list with no height grows to fit every " +
			"message rather than scrolling")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	folder := opts.Folder
	if folder == "" {
		folder = FolderInbox
	}
	title := opts.Title
	if title == "" {
		title = folderName(folder)
	}

	shown := make([]int, 0, len(opts.Messages))
	unread := 0
	for i, m := range opts.Messages {
		if m.Folder != folder {
			continue
		}
		if opts.UnreadOnly && !m.Unread {
			continue
		}
		if !mailMatches(m, opts.Query) {
			continue
		}
		if m.Unread {
			unread++
		}
		shown = append(shown, i)
	}

	cols := make([]data.Column, 0, len(mailColumns))
	for _, col := range mailColumns {
		switch col.ID {
		case "from":
			if folder == FolderSent {
				col.Title, col.ID = "To", "to"
			}
			cols = append(cols, col)
		case "files":
			if !anyAttachments(opts.Messages) {
				continue
			}
			cols = append(cols, col)
		default:
			cols = append(cols, col)
		}
	}

	res := MailListResult{shown: len(shown), unread: unread}

	res.Element = layout.Container(c, layout.ContainerOptions{
		Surface: true, Border: true, Radius: theme.CardRadius, Gap: u, Pad: u * 1.5,
	}, func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u).Children(func() {
			ui.Text(c, title).TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).Bold().Grow(1)
			if unread > 0 {
				ui.Text(c, itoa(unread)).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
			}
		})
		table := data.DataTable(c, data.DataTableOptions{
			Columns: cols,
			Rows:    len(shown),
			Sort:    opts.Sort,
			Cell: func(row, col int) {
				m := opts.Messages[shown[row]]
				switch cols[col].ID {
				case "from":
					who := m.From
					if who == "" {
						who = m.To
					}
					ink := k.Text
					if !m.Unread {
						ink = k.TextMuted
					}
					ui.Text(c, who).TextColor(ink).MaxLines(1).
						FontSize(core.FontSize(c, theme.RowSize))
				case "subject":
					ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 0.75).
						Role(ui.RoleNone).Children(func() {
						// An important message is marked by a dot rather than by
						// a colour or by moving the subject into bold: the row is
						// already told apart from a read one by its weight, and a
						// second difference in the same place makes the reader
						// check the dot on every row to know which one is which.
						if m.Important {
							ui.Box(c).Size(u, u).Radius(u / 2).Background(k.Danger).
								Shrink(0).Label("Important").Role(ui.RoleNone)
						}
						if m.HasAttachments {
							ui.Icon(c, glyph("mail")).Size(u*3.25, u*3.25).
								TextColor(k.TextFaint).Label("Has attachments").Shrink(0)
						}
						ink := k.Text
						if !m.Unread {
							ink = k.TextMuted
						}
						ui.Text(c, m.Subject).TextColor(ink).Grow(1).SingleLine().
							FontSize(core.FontSize(c, theme.RowSize))
					})
				case "files":
					if !m.HasAttachments {
						ui.Text(c, "—").TextColor(k.TextFaint).
							FontSize(core.FontSize(c, theme.RowSize))
						return
					}
					if m.AttachmentCount > 1 {
						ui.Text(c, itoa(m.AttachmentCount)).TextColor(k.TextMuted).
							FontSize(core.FontSize(c, theme.RowSize))
						return
					}
					ui.Text(c, "1").TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.RowSize))
				case "when":
					ink := k.TextMuted
					if m.SnoozedUntil != nil {
						ink = k.AccentText
					}
					ui.Text(c, m.When).TextColor(ink).
						FontSize(core.FontSize(c, theme.RowSize))
				}
			},
			CellLabel: func(row, col int) string {
				m := opts.Messages[shown[row]]
				return m.Subject
			},
			Key:      func(row int) any { return opts.Messages[shown[row]].ID },
			Label:    func(row int) string { return opts.Messages[shown[row]].Subject },
			Selected: opts.Selected,
			State:    opts.State,
			Scroll:   opts.Scroll,
			Height:   opts.Height,
			Empty: func() {
				ui.Column(c).FillWidth().Center().Padding(u * 3).Gap(u).
					Label("No messages").Children(func() {
					ui.Text(c, emptyMail(folder, opts.Query)).TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.BodySize))
				})
			},
		})
		// Built in here rather than beside the panel above: an element
		// belongs to whatever container is current where it is made, so
		// one made out there lands above the panel and the list comes out
		// with its own folder heading sitting underneath it.
		table.Element.FillWidth()
	})
	return res
}

func folderName(f MailFolder) string {
	switch f {
	case FolderSent:
		return "Sent"
	case FolderDrafts:
		return "Drafts"
	case FolderArchive:
		return "Archive"
	case FolderSpam:
		return "Spam"
	}
	return "Inbox"
}

// mailMatches searches the four things a reader searches a mailbox for. The
// body is included because the search field over a mailbox is the tool people
// use to find the message they half-remember, and they remember what it said
// far more often than what it was called.
func mailMatches(m Message, query *string) bool {
	if query == nil || strings.TrimSpace(*query) == "" {
		return true
	}
	needle := lower(strings.TrimSpace(*query))
	return contains(lower(m.Subject), needle) ||
		contains(lower(m.From), needle) ||
		contains(lower(m.To), needle) ||
		contains(lower(m.Body), needle) ||
		contains(lower(m.Preview), needle)
}

func anyAttachments(msgs []Message) bool {
	for _, m := range msgs {
		if m.HasAttachments {
			return true
		}
	}
	return false
}

func emptyMail(folder MailFolder, query *string) string {
	if query != nil && strings.TrimSpace(*query) != "" {
		return "Nothing here matches that"
	}
	switch folder {
	case FolderDrafts:
		return "No drafts"
	case FolderSpam:
		return "No spam"
	}
	return "Nothing in the " + strings.ToLower(folderName(folder))
}

// ── the reader ─────────────────────────────────────────────────────────────

// MailReaderOptions configure a MailReader.
type MailReaderOptions struct {
	// Message is the mail being read, and Body is its text. Both are the
	// caller's: a body that has not been fetched arrives as an empty string,
	// and the reader says so rather than showing an empty page that reads as
	// an empty mail.
	Message Message
	// Body is the whole message when it has been fetched.
	Body string
	// Actions is the row of buttons over the message — reply, forward,
	// archive. The reader does not draw them: what a mail client can do to a
	// message is the application's decision and this library has no opinion
	// about deleting things.
	Actions func()
	// Snoozed says when the message will come back, and is shown in the
	// header so a reader who cannot remember whether they dealt with it can
	// see whether it is coming.
	Snoozed string
	// Width is the reader's own width; zero fills what it is given.
	Width float32
	// Label names the reader for assistive technology.
	Label string
}

// MailReader is one message open: its sender, its subject, its date, its
// body, and whatever the caller can do to it.
//
// It draws the body as plain text and does not render it as HTML. That is a
// deliberate limit rather than an oversight: rendering an arbitrary
// attachment-shaped document is an entire product with its own failure modes,
// and a mail reader in a component library is the place where somebody would
// try to do it.
func MailReader(c *ui.Context, opts MailReaderOptions) *ui.Element {
	if opts.Message.Subject == "" && opts.Message.From == "" {
		panic("messaging: MailReader needs a Message; an empty one is a page with nothing to " +
			"be a page of")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	label := opts.Label
	if label == "" {
		label = opts.Message.Subject
		if label == "" {
			label = "Message from " + opts.Message.From
		}
	}

	body := opts.Body
	if body == "" {
		body = opts.Message.Body
	}

	return layout.Container(c, layout.ContainerOptions{
		Surface: true, Border: true, Radius: theme.CardRadius,
		Pad: u * 2.5, Gap: u * 2,
	}, func() {
		display.Heading(c, opts.Message.Subject, display.HeadingOptions{
			Level: 2, Subtitle: mailSubtitle(opts.Message), Divider: true,
		})
		if opts.Snoozed != "" {
			ui.Row(c).AlignItems(ui.Center).Gap(u).Children(func() {
				ui.Text(c, "Snoozed").TextColor(k.AccentText).
					FontSize(core.FontSize(c, theme.CaptionSize))
				ui.Text(c, opts.Snoozed).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize))
			})
		}
		if body == "" {
			// Said plainly, because a page with nothing on it is read as a
			// message with nothing in it — and the reader's next move is to go
			// and fetch it, which is not a move a blank page suggests.
			ui.Text(c, "This message has not been downloaded yet.").
				TextColor(k.TextMuted).FontSize(core.FontSize(c, theme.BodySize))
		} else {
			ui.Text(c, body).TextColor(k.Text).
				FontSize(core.FontSize(c, theme.BodySize)).MaxLines(0)
		}
		if opts.Actions != nil {
			ui.Row(c).FillWidth().Gap(u*1.5).Justify(ui.End).
				Margin(u, u*2, 0, 0).Children(opts.Actions)
		}
	})
}

func mailSubtitle(m Message) string {
	parts := []string{}
	if m.From != "" {
		parts = append(parts, m.From)
	}
	if m.When != "" {
		parts = append(parts, m.When)
	}
	if m.HasAttachments {
		if m.AttachmentCount > 1 {
			parts = append(parts, itoa(m.AttachmentCount)+" attachments")
		} else {
			parts = append(parts, "1 attachment")
		}
	}
	return strings.Join(parts, " · ")
}

// ── the composer ───────────────────────────────────────────────────────────

// MailComposerOptions configure a MailComposer.
type MailComposerOptions struct {
	// To is required and Cc is not: a nil Cc means the window has no
	// carbon-copy field at all, rather than one that is currently empty. That
	// is the honest reading of a composer without a cc button on it, and it
	// keeps the composer from having to invent a slice for a field nobody can
	// see.
	//
	// To and Cc are the caller's slices of addresses, by pointer.
	//
	// Slices rather than one string, because a message with three recipients
	// is three fields of truth and a comma-joined string has to be parsed
	// again by whatever sends it. By pointer rather than by value, because
	// the recipient fields below write into them: a slice header passed by
	// value shares its array but not its length, so an address added in the
	// composer would be in the field and not in the caller's slice — which is
	// exactly the bug a send button with a stale recipient list produces.
	To, Cc *[]string
	// Body is the draft, the caller's string.
	Body *string
	// People are who can be added, for the mention field's suggestions.
	People []string
	// Suggested is the text the current message fills in — a reply's quoted
	// original, a forward's headers. Empty is a new message.
	Suggested string
	// Sending and Sent are the caller's: the button reports and the caller
	// moves it, because what "sent" means is the application's business and a
	// component that flipped the flag itself would be sending mail the caller
	// had decided not to send.
	Sending *bool
	Sent    bool
	// Submitting and Submitted report the send button and Enter.
	Submitting bool
	Submitted  bool
	// Title heads the panel; empty takes the library's "New message".
	Title string
	// Labels for the three fields, so a window can reword them.
	ToLabel, CcLabel, BodyLabel string
}

// MailComposerResult carries a MailComposer.
type MailComposerResult struct {
	// Element is the composer.
	Element *ui.Element
}

// MailComposer is the window for writing a message: who it is for, what it
// says, and the button that sends it.
//
// The send is reported and the caller acts. That is the only honest shape for
// it: sending is irreversible, the recipient list has to be checked before it
// happens rather than after, and a component with a Send button that pressed
// itself would be a component with a send button that cannot be talked out of.
func MailComposer(c *ui.Context, opts MailComposerOptions) MailComposerResult {
	if opts.Body == nil {
		panic("messaging: MailComposer needs a Body to write into; it keeps no draft of its own")
	}
	if opts.Sending == nil {
		panic("messaging: MailComposer needs a Sending flag; it does not send anything and " +
			"cannot tell whether it is sending")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	toLabel := opts.ToLabel
	if toLabel == "" {
		toLabel = core.Msg(c, "messaging.mail.to", core.Def("To"))
	}
	bodyLabel := opts.BodyLabel
	if bodyLabel == "" {
		bodyLabel = core.Msg(c, "messaging.mail.body", core.Def("Message"))
	}
	ccLabel := opts.CcLabel
	if ccLabel == "" {
		ccLabel = core.Msg(c, "messaging.mail.cc", core.Def("Cc"))
	}

	if opts.Suggested != "" && *opts.Body == "" {
		// Filled in once, on the frame the suggested text arrives. The `== ""`
		// guard is what stops it coming back every frame and retyping the
		// draft out from under the reader.
		*opts.Body = opts.Suggested
	}

	res := MailComposerResult{}
	res.Element = layout.Container(c, layout.ContainerOptions{
		Surface: true, Border: true, Radius: theme.CardRadius,
		Pad: u * 2.5, Gap: u * 1.5,
	}, func() {
		ui.Text(c, mailTitle(opts.Title)).TextColor(k.Text).
			FontSize(core.FontSize(c, theme.SheetSize)).Bold()
		RecipientInput(c, opts.To, RecipientInputOptions{
			Label: toLabel, People: opts.People,
		})
		if opts.Cc != nil {
			RecipientInput(c, opts.Cc, RecipientInputOptions{
				Label: ccLabel, People: opts.People, Muted: true,
			})
		}
		input.TextArea(c, opts.Body, input.TextAreaOptions{
			Label: bodyLabel, Lines: 8, Placeholder: "Write something",
		})

		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.5).Children(func() {
			ui.Box(c).Grow(1)
			ui.Text(c, recipientSummary(opts.To)).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
			if input.Button(c, sendLabel(*opts.Sending), input.ButtonOptions{
				Primary:  true,
				Disabled: *opts.Sending || len(*opts.To) == 0,
			}).Clicked() {
				opts.Submitting = true
			}
		})
		// The two answers are copied out after the children are built, because
		// a closure that assigned them would have the settled pass — which has
		// no press — overwrite the answer with false.
		opts.Submitted = opts.Submitted || opts.Submitting
	})
	return res
}

func mailTitle(t string) string {
	if t != "" {
		return t
	}
	return core.Def("New message")
}

func sendLabel(sending bool) string {
	if sending {
		return "Sending…"
	}
	return "Send"
}

// recipientSummary is what the button row says about the message: who it is
// going to, and what it is going to do about it.
func recipientSummary(to *[]string) string {
	if to == nil {
		return "No recipients"
	}
	list := *to
	switch len(list) {
	case 0:
		return "No recipients"
	case 1:
		return "To " + list[0]
	case 2:
		return "To " + list[0] + " and " + list[1]
	}
	return "To " + list[0] + " and " + itoa(len(list)-1) + " others"
}

// ── recipients ─────────────────────────────────────────────────────────────

// RecipientInputOptions configure a RecipientInput.
type RecipientInputOptions struct {
	// Label names the field and is required.
	Label string
	// People are the addresses that can be added, offered as the field is
	// typed.
	People []string
	// Placeholder shows while the field is empty; it is required, for the same
	// reason input.TagsInput's is: a field of chips that does not say what a
	// chip is leaves the first one unexplained.
	Placeholder string
	// Muted draws the chips in the secondary tone, for the carbon-copy field,
	// which is a secondary address and should not draw the eye off the To.
	Muted bool
	// Max caps how many addresses there may be; zero is no cap.
	Max int
	// Disallowed are addresses that may not be added, for a field that has to
	// refuse an external recipient.
	Disallowed []string
}

// RecipientInput is the field of address chips above a draft.
//
// It is [input.TagsInput] because a recipient field *is* a tags field: chips,
// taken away with their own button, suggestions as the name is typed, and
// Enter or a comma to add. Writing a second one here would be a second set of
// rules about what Enter does in a field of chips.
//
// The one thing it adds is refusal. A recipient field is the one tags field
// where adding the wrong chip is not a mistake the reader can undo and forget —
// it is a message to somebody they did not mean to write to — so an address
// the caller has disallowed is never offered and never added.
func RecipientInput(c *ui.Context, to *[]string, opts RecipientInputOptions) *ui.Element {
	if to == nil {
		panic("messaging: RecipientInput needs a slice of addresses to point at; it keeps " +
			"none of its own")
	}
	if opts.Label == "" {
		panic("messaging: RecipientInput needs a Label; a field of chips has no words of its " +
			"own to be read out by")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	placeholder := opts.Placeholder
	if placeholder == "" {
		placeholder = core.Msg(c, "messaging.mail.recipient", core.Def("name or address"))
	}

	people := allowed(opts.People, opts.Disallowed)

	return ui.Column(c).FillWidth().Gap(u * 0.5).Children(func() {
		ui.Text(c, opts.Label).TextColor(k.TextMuted).
			FontSize(core.FontSize(c, theme.CaptionSize))
		input.TagsInput(c, to, input.TagsInputOptions{
			Label: opts.Label + " field", Placeholder: placeholder,
			Suggestions: people, Max: opts.Max,
		})
		// The chips are painted over in the muted tone when the field is
		// secondary. TagsInput has no tone of its own because a tag is the
		// same thing everywhere else in this library, and the only field
		// where it is not is the one a message is sent from.
		if opts.Muted {
			ui.Text(c, itoa(len(*to))+" added").TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize))
		}
	})
}

// allowed is the suggestions minus the refused ones. It is computed rather than
// left to the caller because the rule is the same wherever the field is: a
// suggestion the field would refuse is worse than no suggestion at all, since
// it invites the press that then does nothing.
func allowed(people, disallowed []string) []string {
	if len(disallowed) == 0 {
		return people
	}
	out := make([]string, 0, len(people))
	for _, p := range people {
		if !has(disallowed, p) {
			out = append(out, p)
		}
	}
	return out
}

// ── snoozing ───────────────────────────────────────────────────────────────

// Snooze is a time a message can be brought back at, and the two things that
// have to be true of every one of them: it is a time the caller can compute,
// and it is described in the reader's own words rather than in the library's.
type Snooze struct {
	// Value is what is written into the caller's choice. Required: a snooze
	// the caller cannot recognise afterwards is one it cannot schedule.
	Value string
	// Label is what the reader sees.
	Label string
	// Hint is the secondary line — "until tomorrow morning". Empty leaves it
	// out.
	Hint string
	// Until is what the library writes the choice back as, when the caller
	// wants the actual time rather than the word. It is set for the fixed
	// choices and nil for a custom one, which is the one the caller has to
	// resolve itself.
	Until *time.Time
}

// DefaultSnoozes is the set every mail client offers. It is a function rather
// than a variable because the times are relative to now, and a package-level
// slice of times would be computed once at start-up and then be wrong by the
// time anybody opened the menu.
func DefaultSnoozes(now time.Time) []Snooze {
	later := func(d time.Duration) *time.Time {
		t := now.Add(d)
		return &t
	}
	hours := func(h int) *time.Time { return later(time.Duration(h) * time.Hour) }
	return []Snooze{
		{Value: "later-today", Label: "Later today", Hint: "at 5pm", Until: hours(4)},
		{Value: "tonight", Label: "Tonight", Hint: "at 9pm", Until: hours(9)},
		{Value: "tomorrow", Label: "Tomorrow", Hint: "at 9am", Until: later(24 * time.Hour)},
		{Value: "weekend", Label: "This weekend", Hint: "Saturday morning", Until: later(72 * time.Hour)},
		{Value: "next-week", Label: "Next week", Hint: "Monday morning", Until: later(168 * time.Hour)},
	}
}

// SnoozePickerOptions configure a SnoozePicker.
type SnoozePickerOptions struct {
	// Anchor is the button the menu hangs off, and is required.
	Anchor *ui.Element
	// Open is the *bool the menu opens and closes with, and is required.
	Open *bool
	// Now is the caller's clock. It is required rather than read from the
	// system because a snooze is a promise about a moment, and a component
	// that computed it from a clock the caller cannot see would be making
	// that promise on the caller's behalf with a time they never chose.
	Now time.Time
	// Choices overrides the offered times. Empty takes [DefaultSnoozes] at Now.
	Choices []Snooze
	// Chosen is the value picked this frame, or "". Reported rather than
	// written, because scheduling the message is the caller's to do.
	Chosen string
	// Label names the menu; empty takes the library's "Snooze until".
	Label string
}

// SnoozePickerResult carries a SnoozePicker.
type SnoozePickerResult struct {
	// Element is the anchor, with the menu hung off it.
	Element *ui.Element
	// until is the time a fixed choice resolves to, and has is whether there
	// was one: a custom snooze has no time until the caller picks one, and
	// reporting a time for it would be inventing it.
	until *time.Time
	has   bool
}

// Until is the moment a fixed choice resolves to, and false when nothing was
// chosen this frame or the choice was a custom one. It is reported separately
// from the choice because the two are different kinds of answer: "next week"
// is a label the caller may want to store, and a timestamp it must schedule
// against.
func (r SnoozePickerResult) Until() (time.Time, bool) {
	if !r.has || r.until == nil {
		return time.Time{}, false
	}
	return *r.until, true
}

// SnoozePicker is the menu of times a message can be brought back at.
//
// It hangs off an anchor rather than being one, because a snooze is always
// pressed on a specific message: the button that says "Snooze" belongs to the
// message being snoozed, and a snooze control standing on its own has to be
// told which message it applies to, which is a field the caller would
// inevitably forget to set.
func SnoozePicker(c *ui.Context, opts SnoozePickerOptions) SnoozePickerResult {
	if opts.Anchor == nil {
		panic("messaging: SnoozePicker needs the Anchor the menu hangs off")
	}
	if opts.Open == nil {
		panic("messaging: SnoozePicker needs the *bool it opens and closes with")
	}
	if opts.Now.IsZero() {
		panic("messaging: SnoozePicker needs Now; a snooze is a promise about a moment and " +
			"the library has no clock to make it with")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	label := opts.Label
	if label == "" {
		label = core.Msg(c, "messaging.snooze", core.Def("Snooze until"))
	}
	choices := opts.Choices
	if len(choices) == 0 {
		choices = DefaultSnoozes(opts.Now)
	}

	var res SnoozePickerResult
	overlay.Popover(c, opts.Anchor, opts.Open, overlay.PopoverOptions{
		Modal: false,
		Title: label,
		Label: label,
		Body: func() {
			for _, ch := range choices {
				if ch.Value == "" {
					continue
				}
				choice := ch
				// A Row, not a Box: the time is the second half of the
				// choice's sentence, not the line under it.
				row := ui.Row(c).FillWidth().Padding(u*0.75, u*1.5).Gap(u * 2).
					Radius(theme.SmallRadius).Label(choice.Label).
					Cursor(ui.CursorPointer).Role(ui.RoleNone)
				if row.Hovered() {
					row.Background(k.SurfaceHover)
				}
				if row.Clicked() {
					res.until, res.has = choice.Until, choice.Until != nil
					*opts.Open = false
				}
				row.Children(func() {
					ui.Text(c, choice.Label).TextColor(k.Text).Grow(1).
						FontSize(core.FontSize(c, theme.RowSize))
					if choice.Hint != "" {
						ui.Text(c, choice.Hint).TextColor(k.TextFaint).
							FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
					}
				})
			}
		},
	})

	if opts.Chosen != "" {
		res.has = false
		res.until = nil
	}
	return res
}
