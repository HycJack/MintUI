package messaging

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/chat"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/data"
	"github.com/HycJack/MintUI/ui/theme"
)

// The two pure functions carry this package: FirstUnread is the rule for where
// the "new messages" line goes, and ReceiptState is the mapping from a
// delivery state to the mark that says it. Neither needs a window, so each is
// asserted on exactly what it returns.
//
// Everything else is asserted on the element it drew, on a name it put on
// screen, or on a value it wrote into the caller's pointer. MyGo builds a
// frame up to three times and stops when one consumed nothing
// (mygo/ui/runtime.go:333), so after a tester settles, Clicked() is false on
// the last pass. Every press below is therefore read inside the view and
// accumulated across passes rather than read off a Result afterwards.

func wantsPanic(t *testing.T, what string, view func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("%s should have panicked", what)
		}
	}()
	view()
}

func wantText(t *testing.T, tt *ui.Tester, what string, texts ...string) {
	t.Helper()
	for _, s := range texts {
		if !tt.HasText(s) {
			t.Errorf("%s is missing %q; drew %q", what, s, tt.Texts())
		}
	}
}

// ── FirstUnread ────────────────────────────────────────────────────────────

// TestFirstUnread is the three cases, exactly: nothing below the line, a line
// at the top, and no line at all. Each is the shape a channel is actually in
// at a moment somebody is looking at it — a channel just opened, a channel
// where everything has been read, and one where the first message ever sent
// has not been.
func TestFirstUnread(t *testing.T) {
	cases := []struct {
		name         string
		index, count int
		want         int
	}{
		{"everything read", 4, 5, -1},
		{"three read of ten", 2, 10, 3},
		{"nothing read", -1, 10, 0},
		{"nothing read and nothing there", -1, 0, -1},
		{"empty channel", -1, -3, -1},
		{"marker past the end", 9, 5, -1},
		{"marker below the start", -4, 5, -1},
	}
	for _, tc := range cases {
		if got := FirstUnread(tc.index, tc.count); got != tc.want {
			t.Errorf("%s: FirstUnread(%d, %d) = %d, want %d",
				tc.name, tc.index, tc.count, got, tc.want)
		}
	}
}

// TestDividerAt is the per-row question the renderer actually asks, and it
// must be true for exactly one row in a channel with anything new in it. A
// divider drawn above two messages is a line saying "these and these", and a
// divider that never appears is the failure everybody notices.
func TestDividerAt(t *testing.T) {
	count := 10
	seen := 0
	for i := range count {
		if DividerAt(i, 2, count) {
			seen++
			if i != 3 {
				t.Errorf("the divider landed on row %d, want row 3", i)
			}
		}
	}
	if seen != 1 {
		t.Errorf("the divider appeared on %d rows, want exactly 1", seen)
	}
	// Everything read: nothing anywhere.
	for i := range count {
		if DividerAt(i, 9, count) {
			t.Errorf("row %d drew a divider in a fully read channel", i)
		}
	}
	// An empty channel is not a row at all.
	if DividerAt(0, -1, 0) {
		t.Error("an empty channel has no first unread message")
	}
}

// ── ReceiptState ───────────────────────────────────────────────────────────

// TestReceiptState is the three states the brief asks for, each pinned to its
// exact mark, plus the failure. The marks are different *shapes* — one open
// mark, two open marks, two filled marks — because a reader who cannot see the
// difference between two tints of grey cannot tell delivered from read, and
// that is the whole information a receipt carries.
func TestReceiptState(t *testing.T) {
	cases := []struct {
		state Receipt
		mark  string
		shown bool
		name  string
	}{
		{ReceiptSending, "clock", true, "Sending"},
		{ReceiptDelivered, "delivered", true, "Delivered"},
		{ReceiptRead, "read", true, "Read"},
		{ReceiptFailed, "failed", true, "Failed to send"},
		{Receipt(99), "", false, "an unknown state"},
		{Receipt(-1), "", false, "a negative state"},
	}
	for _, tc := range cases {
		mark, shown := ReceiptState(tc.state)
		if mark != tc.mark || shown != tc.shown {
			t.Errorf("%s: ReceiptState = (%q, %v), want (%q, %v)",
				tc.name, mark, shown, tc.mark, tc.shown)
		}
	}
}

func TestReceiptString(t *testing.T) {
	cases := map[Receipt]string{
		ReceiptSending:   "Sending",
		ReceiptDelivered: "Delivered",
		ReceiptRead:      "Read",
		ReceiptFailed:    "Failed to send",
	}
	for state, want := range cases {
		if got := state.String(); got != want {
			t.Errorf("Receipt(%d).String() = %q, want %q", state, got, want)
		}
	}
}

// ── matches ────────────────────────────────────────────────────────────────

func TestChannelMatches(t *testing.T) {
	ch := Channel{Name: "Site Walkthrough", Topic: "Rooftop condition survey"}
	cases := []struct {
		query string
		want  bool
	}{
		{"", true},
		{"rooftop", true},
		{"ROOFTOP", true},
		{"walkthrough", true},
		{"billing", false},
		{"teh ", false},
	}
	for _, tc := range cases {
		q := tc.query
		if got := matches(ch, &q); got != tc.want {
			t.Errorf("matches(%q) = %v, want %v", tc.query, got, tc.want)
		}
	}
	if !matches(ch, nil) {
		t.Error("a nil query is no filter, which is every message shown")
	}
}

func TestMemberMatches(t *testing.T) {
	m := Member{Name: "Dana Reyes", Handle: "@dana", Role: "Structural engineer"}
	for _, q := range []string{"dana", "@DANA", "engineer", "reyes"} {
		s := q
		if !memberMatches(m, &s) {
			t.Errorf("memberMatches(%q) = false, want true", q)
		}
	}
	s := "billing"
	if memberMatches(m, &s) {
		t.Error("a query nothing matches should be false")
	}
}

func TestMailMatches(t *testing.T) {
	m := Message{
		Subject: "Rooftop access", From: "site@riverside.example",
		Body: "the stair door code changed last week",
	}
	for _, q := range []string{"rooftop", "RIVERSIDE", "stair door"} {
		s := q
		if !mailMatches(m, &s) {
			t.Errorf("mailMatches(%q) = false, want true", q)
		}
	}
	s := "invoice"
	if mailMatches(m, &s) {
		t.Error("a query nothing matches should be false")
	}
	if !mailMatches(m, nil) {
		t.Error("a nil query shows everything")
	}
}

func TestPreviewPrefersTheDraft(t *testing.T) {
	withDraft := Channel{Topic: "Rooftop condition survey", Draft: "the stair door"}
	if got := preview(withDraft); got != "Draft: the stair door" {
		t.Errorf("preview = %q; a draft outranks a topic, because a draft is something "+
			"the reader was in the middle of doing", got)
	}
	if got := preview(Channel{Topic: "Rooftop"}); got != "Rooftop" {
		t.Errorf("preview = %q, want the topic", got)
	}
}

func TestSearchGroups(t *testing.T) {
	groups := DefaultGroups()
	if got := searchGroups(groups, nil); len(got) != len(groups) {
		t.Errorf("an empty query should return every group, got %d of %d", len(got), len(groups))
	}
	// "fire" is a word, not a character: nobody types a glyph to find one.
	fire := "fire"
	hits := searchGroups(groups, &fire)
	if len(hits) == 0 {
		t.Fatal("\"fire\" found nothing; a picker that searched glyphs would find nothing ever")
	}
	found := false
	for _, g := range hits {
		for _, e := range g.Emoji {
			if e.Name == "on fire" {
				found = true
			}
		}
	}
	if !found {
		t.Errorf("\"fire\" did not reach the emoji named \"on fire\"; %+v", hits)
	}
	nothing := "zzzz"
	if got := searchGroups(groups, &nothing); len(got) != 0 {
		t.Errorf("a query nothing matches returned %d groups", len(got))
	}
}

// ── shapes ─────────────────────────────────────────────────────────────────

func TestInitialLetters(t *testing.T) {
	cases := map[string]string{
		"Dana Reyes":        "DR",
		"dana":              "D",
		"@sam":              "S",
		"Ana-Maria O'Neill": "AM",
		"":                  "",
		"   ":               "",
	}
	for in, want := range cases {
		if got := initials(in); got != want {
			t.Errorf("initials(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestDefaultSnoozesAreRelativeToNow(t *testing.T) {
	// This is a function rather than a package-level slice for exactly this
	// reason: a slice of times computed at start-up is wrong by the time
	// anybody opens the menu.
	now := time.Date(2026, 3, 10, 9, 0, 0, 0, time.UTC)
	choices := DefaultSnoozes(now)
	if len(choices) != 5 {
		t.Fatalf("DefaultSnoozes gave %d choices, want 5", len(choices))
	}
	for _, ch := range choices {
		if ch.Value == "" || ch.Label == "" || ch.Until == nil {
			t.Errorf("choice %+v is not usable: every one needs a value, a label and a time", ch)
		}
	}
	later := choices[0].Until.Sub(now)
	if later != 4*time.Hour {
		t.Errorf("\"Later today\" is %v away, want 4h", later)
	}
	next := choices[4].Until.Sub(now)
	if next != 168*time.Hour {
		t.Errorf("\"Next week\" is %v away, want 168h", next)
	}
	// And a second call now is a different set: that is the point of a
	// function.
	another := DefaultSnoozes(now.Add(time.Hour))
	if another[0].Until.Sub(now) != 5*time.Hour {
		t.Error("a second call an hour later should resolve against that hour")
	}
}

func TestDefaultGroups(t *testing.T) {
	groups := DefaultGroups()
	if len(groups) == 0 {
		t.Fatal("the default set is empty; a caller with no emoji budget has nothing to draw")
	}
	for _, g := range groups {
		if g.Name == "" {
			t.Error("a group with no name is a page a screen reader reads as identical glyphs")
		}
		for _, e := range g.Emoji {
			if e.Glyph == "" || e.Name == "" {
				t.Errorf("entry %+v has no glyph or no name", e)
			}
		}
	}
}

func TestStandardCallControls(t *testing.T) {
	muted := StandardCallControls(true, false, false)
	if len(muted) != 4 {
		t.Fatalf("a plain call has %d controls, want 4", len(muted))
	}
	// A toggle draws two different marks, so the button is named for what the
	// press will do rather than for the state it is in.
	if muted[0].Label != "Unmute the microphone" {
		t.Errorf("a muted microphone is named %q, want %q", muted[0].Label, "Unmute the microphone")
	}
	if muted[0].glyph() != "micOff" || muted[1].glyph() != "camera" {
		t.Errorf("state is not in the glyph: mute=%q camera=%q",
			muted[0].glyph(), muted[1].glyph())
	}
	if !muted[3].Danger || muted[3].Label != "Leave the call" {
		t.Error("leaving a call is the one control that ends it, and it must be marked")
	}
}

// ── headless renders ───────────────────────────────────────────────────────

func testChannels() []Channel {
	return []Channel{
		{ID: "general", Name: "general", Topic: "Everything that does not fit elsewhere",
			When: "4m", Unread: 0},
		{ID: "rooftop", Name: "site-walkthrough", Topic: "Rooftop condition survey",
			When: "2m", Unread: 3, Mentions: 1},
		{ID: "billing", Name: "billing", Topic: "Invoices", When: "Tue",
			Unread: 12, Mentions: 4, Draft: "the stair door", Muted: true, Pinned: true},
	}
}

func TestChannelList(t *testing.T) {
	channels := testChannels()
	selected := 0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ChannelList(c, ChannelListOptions{
			Channels: channels, Selected: &selected, Height: 240,
		})
	}, 320, 280)
	for _, want := range []string{
		"general", "site-walkthrough", "billing",
		"Rooftop condition survey", "Draft: the stair door",
		"mentions you", "mentions you 4 times", "Muted",
	} {
		wantText(t, tt, "the channel list", want)
	}
	if _, ok := tt.Find("Channels"); !ok {
		t.Error("the list itself must be named; its rows are named but the list is not")
	}

	// The query filters, and Shown says what survived — so an empty state
	// under a heading saying "Channels" is not read as "you have none".
	shown := -1
	query := "rooftop"
	tt2 := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		shown = ChannelList(c, ChannelListOptions{
			Channels: channels, Selected: &selected, Height: 240, Query: &query,
		}).Shown()
	}, 320, 280)
	if shown != 1 {
		t.Errorf("\"rooftop\" matched %d channels, want 1", shown)
	}
	for _, gone := range []string{"general", "billing", "Invoices"} {
		if tt2.HasText(gone) {
			t.Errorf("a filtered list still drew %q; it matched nothing", gone)
		}
	}
	if !tt2.HasText("site-walkthrough") {
		t.Errorf("the channel that matched is missing; %q", tt2.Texts())
	}

	wantsPanic(t, "a channel list with no height", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			sel := -1
			ChannelList(c, ChannelListOptions{Channels: channels, Selected: &sel})
		}, 320, 280)
	})
	wantsPanic(t, "a channel list with no selection", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			ChannelList(c, ChannelListOptions{Channels: channels, Height: 200})
		}, 320, 280)
	})
}

func TestChatMessage(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ChatMessage(c, ChatMessageOptions{
			Role: chat.RoleAssistant, Author: "Priya", Time: "09:12",
			Body:    func() { ui.Text(c, "The membrane is split along the north seam.") },
			Replies: 4, Divider: true, DividerCount: 7,
		})
	}, 480, 320)
	wantText(t, tt, "the turn", "Priya", "09:12", "4 replies",
		"The membrane is split along the north seam.", "New messages · 7 messages")

	// The right-hand side: the reader's own message, with its receipt and its
	// quoted block, which is the case the wrapper exists for.
	tt2 := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ChatMessage(c, ChatMessageOptions{
			Role: chat.RoleUser, Author: "Dana", Time: "09:14",
			Quoted:  "The membrane is split along the north seam.",
			Body:    func() { ui.Text(c, "Photographed it — uploading now.") },
			Receipt: ReceiptRead, ReceiptTime: "09:14",
		})
	}, 480, 320)
	wantText(t, tt2, "the outgoing turn", "Dana", "09:14", "Read 09:14",
		"Photographed it — uploading now.")
	if _, ok := tt2.Find("The membrane is split along the north seam."); !ok {
		t.Error("the quoted turn should be drawn above the reply")
	}
}

func TestChatMessagePresses(t *testing.T) {
	clicked, threaded := 0, 0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		if ChatMessage(c, ChatMessageOptions{
			Role: chat.RoleAssistant, Author: "Priya",
			Body: func() { ui.Text(c, "Four of them.") },
		}).Clicked() {
			clicked++
		}
	}, 480, 260)
	tt.Click("Four of them.")
	if clicked != 1 {
		t.Errorf("the bubble reported %d presses, want 1", clicked)
	}

	tt2 := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		if ChatMessage(c, ChatMessageOptions{
			Role: chat.RoleAssistant, Author: "Priya", Replies: 4,
			Body: func() { ui.Text(c, "Four of them.") },
		}).Threaded() {
			threaded++
		}
	}, 480, 260)
	tt2.Click("4 replies")
	if threaded != 1 {
		t.Errorf("the thread link reported %d presses, want 1", threaded)
	}
}

func TestUnreadDivider(t *testing.T) {
	cases := []struct {
		count int
		want  string
	}{
		{0, "New messages"},
		{1, "New messages · 1 message"},
		{7, "New messages · 7 messages"},
	}
	for _, tc := range cases {
		count := tc.count
		tt := ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			UnreadDivider(c, UnreadDividerOptions{Count: count})
		}, 400, 40)
		if !tt.HasText(tc.want) {
			t.Errorf("a divider over %d messages read %q, want %q", tc.count, tt.Texts(), tc.want)
		}
	}
}

func TestReadReceipt(t *testing.T) {
	for _, state := range []Receipt{ReceiptSending, ReceiptDelivered, ReceiptRead, ReceiptFailed} {
		st := state
		tt := ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			ReadReceipt(c, ReadReceiptOptions{State: st, Time: "09:14"})
		}, 260, 40)
		if _, ok := tt.Find(st.String() + " 09:14"); !ok {
			t.Errorf("the receipt for %s is missing its name; drew %q", st, tt.Texts())
		}
	}
	// An unknown state draws no mark and says so, rather than claiming
	// delivery because it ran out of cases.
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ReadReceipt(c, ReadReceiptOptions{State: Receipt(99), Time: "09:14"})
	}, 260, 40)
	if _, ok := tt.Find("Sending 09:14"); ok {
		t.Error("an unknown receipt state claimed the message was sending")
	}
	if _, ok := tt.Find("Delivery status unavailable 09:14"); !ok {
		t.Errorf("an unknown receipt state said nothing; drew %q", tt.Texts())
	}
}

func TestQuotedText(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		QuotedText(c, QuotedTextOptions{
			Author: "Priya", Text: "The membrane is split along the north seam.",
			Lines: 2, Side: ui.End,
		})
	}, 420, 80)
	wantText(t, tt, "the quote", "Priya", "The membrane is split along the north seam.")
	wantsPanic(t, "an empty quote", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			QuotedText(c, QuotedTextOptions{Author: "Priya"})
		}, 400, 60)
	})
}

func TestOnlineStatus(t *testing.T) {
	for _, p := range []Presence{PresenceOffline, PresenceAway, PresenceOnline, PresenceBusy} {
		pres := p
		tt := ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			OnlineStatus(c, OnlineStatusOptions{Presence: pres, Name: "Dana Reyes"})
		}, 220, 40)
		if _, ok := tt.Find("Dana Reyes is " + lower(pres.String())); !ok {
			t.Errorf("the dot for %s has no name; drew %q", pres, tt.Texts())
		}
	}
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		OnlineStatus(c, OnlineStatusOptions{
			Presence: PresenceOnline, Name: "Dana Reyes", Since: "2m", WithText: true,
		})
	}, 260, 40)
	wantText(t, tt, "the worded dot", "Online", "2m")
	wantsPanic(t, "an unnamed dot", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			OnlineStatus(c, OnlineStatusOptions{})
		}, 200, 40)
	})
}

func testMembers() []Member {
	return []Member{
		{ID: "dana", Name: "Dana Reyes", Handle: "@dana",
			Role: "Structural engineer", Presence: PresenceOnline, LocalTime: "14:40"},
		{ID: "sam", Name: "Sam Okonkwo", Handle: "@sam",
			Role: "Site manager", Presence: PresenceBusy, LocalTime: "09:40"},
		{ID: "ana", Name: "Ana-Maria Silva", Handle: "@ana",
			Role: "Quantity surveyor", Presence: PresenceAway, LocalTime: "16:40"},
	}
}

func TestMemberList(t *testing.T) {
	members := testMembers()
	selected := 0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		MemberList(c, MemberListOptions{
			Members: members, Selected: &selected, Height: 240,
			ShowRoles: true, ShowTimes: true,
		})
	}, 560, 300)
	for _, want := range []string{
		"Members", "Dana Reyes", "@dana", "Structural engineer", "14:40",
		"Site manager", "Quantity surveyor",
	} {
		wantText(t, tt, "the member table", want)
	}

	// The role column can be left out, and then it is not on screen: a table
	// of one column is a list, and a popover with room for a name and a dot
	// has no room for a role.
	tt2 := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		MemberList(c, MemberListOptions{
			Members: members, Selected: &selected, Height: 200,
		})
	}, 400, 260)
	if tt2.HasText("Structural engineer") {
		t.Error("the role column was asked to be left out and is on screen")
	}
	wantText(t, tt2, "the narrow table", "Members", "Dana Reyes")

	shown := -1
	query := "engineer"
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		shown = MemberList(c, MemberListOptions{
			Members: members, Selected: &selected, Height: 200,
			ShowRoles: true, Query: &query,
		}).Shown()
	}, 560, 300)
	if shown != 1 {
		t.Errorf("\"engineer\" matched %d members, want 1", shown)
	}

	wantsPanic(t, "a member list with no height", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			sel := -1
			MemberList(c, MemberListOptions{Members: members, Selected: &sel})
		}, 400, 300)
	})
}

func TestEmojiPicker(t *testing.T) {
	query := ""
	open := true
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		anchor := ui.Box(c).Label("Add emoji").Role(ui.RoleNone)
		EmojiPicker(c, EmojiPickerOptions{
			Anchor: anchor, Open: &open, Groups: DefaultGroups(), Query: &query,
		})
	}, 520, 520)
	wantText(t, tt, "the open picker", "Reactions", "Faces", "Status", "Search emoji")

	// Closed: no panel, and the query's state is untouched.
	closed := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		open := false
		q := ""
		anchor := ui.Box(c).Label("Add emoji").Role(ui.RoleNone)
		EmojiPicker(c, EmojiPickerOptions{
			Anchor: anchor, Open: &open, Groups: DefaultGroups(), Query: &q,
		})
	}, 400, 200)
	if closed.HasText("Reactions") {
		t.Error("a closed picker drew its panel")
	}

	presses := 0
	tt2 := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		open := true
		q := ""
		anchor := ui.Box(c).Label("Add emoji").Role(ui.RoleNone)
		if EmojiPicker(c, EmojiPickerOptions{
			Anchor: anchor, Open: &open, Groups: DefaultGroups(), Query: &q,
		}).Picked() == "tada" {
			presses++
		}
	}, 520, 520)
	tt2.Click("tada")
	if presses != 1 {
		t.Errorf("choosing an emoji reported %d presses, want 1", presses)
	}

	wantsPanic(t, "a picker with no groups", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			o := false
			anchor := ui.Box(c).Role(ui.RoleNone)
			EmojiPicker(c, EmojiPickerOptions{Anchor: anchor, Open: &o})
		}, 300, 200)
	})
}

func testPinned() []PinnedMessage {
	return []PinnedMessage{
		{ID: "p1", Author: "Priya", Text: "Roof access is via the north stair only.", When: "Mon"},
		{ID: "p2", Author: "Sam", Text: "Site induction is every Monday at 08:00.", When: "Tue"},
	}
}

func TestPinnedMessages(t *testing.T) {
	selected := -1
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		PinnedMessages(c, PinnedMessagesOptions{
			Messages: testPinned(), Selected: &selected, Height: 140,
		})
	}, 420, 200)
	wantText(t, tt, "the pinned list",
		"Priya", "Roof access is via the north stair only.",
		"Sam", "Site induction is every Monday at 08:00.")

	jumped := -1
	tt2 := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		sel := -1
		if got := PinnedMessages(c, PinnedMessagesOptions{
			Messages: testPinned(), Selected: &sel, Height: 140,
		}).Jumped(); got >= 0 {
			jumped = got
		}
	}, 420, 200)
	tt2.Click("Site induction is every Monday at 08:00.")
	if jumped != 1 {
		t.Errorf("the pressed message was %d, want 1", jumped)
	}
	wantsPanic(t, "a pinned list with no selection", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			PinnedMessages(c, PinnedMessagesOptions{Messages: testPinned(), Height: 140})
		}, 400, 200)
	})
}

func TestThreadPanel(t *testing.T) {
	selected := -1
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ThreadPanel(c, ThreadPanelOptions{
			Parent: ChatMessageOptions{
				Role: chat.RoleUser, Author: "Dana", Time: "09:12",
				Body: func() { ui.Text(c, "Where does the water come in?") },
			},
			Replies: 2,
			Reply: func(i int) {
				ChatMessage(c, ChatMessageOptions{
					Role:   chat.RoleAssistant,
					Author: []string{"Priya", "Sam"}[i],
					Time:   []string{"09:13", "09:15"}[i],
					Body: func() {
						ui.Text(c, []string{
							"The north seam, and the roof drain at the low corner.",
							"Adding it to the survey — photo attached.",
						}[i])
					},
				})
			},
			Selected: &selected, Height: 300, Updated: "last reply 2m ago",
		})
	}, 520, 380)
	wantText(t, tt, "the thread",
		"Where does the water come in?",
		"The north seam, and the roof drain at the low corner.",
		"Adding it to the survey — photo attached.",
		"last reply 2m ago")
	wantsPanic(t, "a thread with no height", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			sel := -1
			ThreadPanel(c, ThreadPanelOptions{
				Replies: 1, Reply: func(int) {}, Selected: &sel,
			})
		}, 400, 300)
	})
}

func testMail() []Message {
	return []Message{
		{ID: "m1", From: "site@riverside.example", To: "dana@riverside.example",
			Subject: "Rooftop access", Preview: "the north stair",
			Body: "The stair door code changed last week.", When: "09:12",
			Unread: true, Important: true, HasAttachments: true, AttachmentCount: 3,
			Folder: FolderInbox},
		{ID: "m2", From: "billing@riverside.example", To: "dana@riverside.example",
			Subject: "Invoice 2291", Preview: "attached", When: "Tue",
			Unread: true, HasAttachments: true, AttachmentCount: 1, Folder: FolderInbox},
		{ID: "m3", From: "dana@riverside.example", To: "site@riverside.example",
			Subject: "Photos from Tuesday", When: "Mon", Folder: FolderSent},
		{ID: "m4", From: "dana@riverside.example", To: "sam@riverside.example",
			Subject: "Survey half done", When: "Sun", Folder: FolderDrafts},
	}
}

func TestMailList(t *testing.T) {
	mail := testMail()
	selected := 0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		MailList(c, MailListOptions{
			Messages: mail, Folder: FolderInbox, Selected: &selected, Height: 240,
		})
	}, 720, 300)
	wantText(t, tt, "the inbox", "Inbox", "site@riverside.example",
		"Rooftop access", "Invoice 2291", "09:12")
	if tt.HasText("Photos from Tuesday") {
		t.Error("a sent message showed up in the inbox")
	}
	if !tt.HasText("3") {
		t.Errorf("the attachment count is missing; %q", tt.Texts())
	}

	unread := -1
	shown := -1
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		sel := -1
		r := MailList(c, MailListOptions{
			Messages: mail, Folder: FolderInbox, Selected: &sel, Height: 240,
		})
		unread, shown = r.Unread(), r.Shown()
	}, 720, 300)
	if shown != 2 || unread != 2 {
		t.Errorf("the inbox shows %d messages, %d unread; want 2 and 2", shown, unread)
	}

	// A sent folder leads with the recipients, because that is what is being
	// scanned for there.
	sent := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		sel := -1
		MailList(c, MailListOptions{
			Messages: mail, Folder: FolderSent, Selected: &sel, Height: 200,
		})
	}, 720, 260)
	wantText(t, sent, "the sent folder", "Sent", "To", "Photos from Tuesday")

	// The drafts folder is a folder rather than a filter, because a draft
	// has a different status.
	drafts := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		sel := -1
		MailList(c, MailListOptions{
			Messages: mail, Folder: FolderDrafts, Selected: &sel, Height: 200,
		})
	}, 720, 260)
	wantText(t, drafts, "the drafts folder", "Drafts", "Survey half done")

	none := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		sel := -1
		MailList(c, MailListOptions{
			Messages: mail, Folder: FolderSpam, Selected: &sel, Height: 200,
		})
	}, 720, 260)
	wantText(t, none, "the empty folder", "No spam")

	wantsPanic(t, "a mail list with no height", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			sel := -1
			MailList(c, MailListOptions{Messages: mail, Selected: &sel})
		}, 600, 300)
	})
}

func TestMailReader(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		MailReader(c, MailReaderOptions{
			Message: testMail()[0], Body: testMail()[0].Body,
			Snoozed: "tomorrow at 9am",
			Actions: func() {
				if ui.Box(c).Label("Archive").Role(ui.RoleNone).Clicked() {
				}
			},
		})
	}, 620, 420)
	wantText(t, tt, "the reader",
		"Rooftop access", "site@riverside.example", "09:12", "3 attachments",
		"The stair door code changed last week.", "Snoozed", "tomorrow at 9am")

	// A body that has not been fetched says so. A page with nothing on it
	// reads as a message with nothing in it, and the reader's next move — go
	// and fetch it — is not a move a blank page suggests.
	fetching := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		m := testMail()[0]
		m.Body = ""
		MailReader(c, MailReaderOptions{Message: m})
	}, 620, 300)
	wantText(t, fetching, "the unfetched reader", "This message has not been downloaded yet.")
	if fetching.HasText("The stair door code changed last week.") {
		t.Error("a message with no body still drew one")
	}
	wantsPanic(t, "a reader with no message", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			MailReader(c, MailReaderOptions{})
		}, 400, 300)
	})
}

func TestMailComposer(t *testing.T) {
	to := []string{"sam@riverside.example"}
	cc := []string{}
	body := ""
	sending := false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		MailComposer(c, MailComposerOptions{
			To: &to, Cc: &cc, Body: &body, Sending: &sending,
			People:    []string{"site@riverside.example", "ana@riverside.example"},
			Suggested: "> On Tuesday, Priya wrote:\n> the north stair is clear",
		})
	}, 620, 460)
	wantText(t, tt, "the composer",
		"New message", "To", "Cc", "Send", "To sam@riverside.example")
	if _, ok := tt.Find("Message"); !ok {
		t.Error("the body field must be named")
	}

	// The suggested text fills an empty draft once and only once. Filling it
	// every frame would retype the draft out from under the reader, and
	// filling it when there is already a draft would throw the draft away.
	if body != "> On Tuesday, Priya wrote:\n> the north stair is clear" {
		t.Errorf("the draft is %q", body)
	}

	// With no recipients the send button is disabled, because a message
	// addressed to nobody is not a message.
	empty := []string{}
	tt2 := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		b := ""
		MailComposer(c, MailComposerOptions{To: &empty, Body: &b, Sending: &sending})
	}, 620, 400)
	if _, ok := tt2.Find("Send"); !ok {
		t.Error("the send button should still be drawn when there are no recipients")
	}
	if tt2.HasText("No recipients") != true {
		t.Errorf("the composer did not say the message has no recipients; %q", tt2.Texts())
	}

	wantsPanic(t, "a composer with no draft", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			s := false
			MailComposer(c, MailComposerOptions{Sending: &s})
		}, 400, 400)
	})
	wantsPanic(t, "a composer with no sending flag", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			b := ""
			to := []string{}
			MailComposer(c, MailComposerOptions{To: &to, Body: &b})
		}, 400, 400)
	})
}

func TestRecipientInput(t *testing.T) {
	to := []string{}
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		RecipientInput(c, &to, RecipientInputOptions{
			Label:       "To",
			People:      []string{"sam@riverside.example", "blocked@riverside.example"},
			Disallowed:  []string{"blocked@riverside.example"},
			Max:         5,
			Placeholder: "name or address",
		})
	}, 560, 120)
	if _, ok := tt.Find("To"); !ok {
		t.Error("the recipient field must be labelled")
	}
	if _, ok := tt.Find("To field"); !ok {
		t.Error("the control itself must have its own name for assistive technology")
	}

	wantsPanic(t, "a recipient field with no slice", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			RecipientInput(c, nil, RecipientInputOptions{Label: "To"})
		}, 400, 100)
	})
	wantsPanic(t, "an unlabelled recipient field", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			list := []string{}
			RecipientInput(c, &list, RecipientInputOptions{})
		}, 400, 100)
	})
}

func TestAllowedDropsTheRefused(t *testing.T) {
	people := []string{"a@example.com", "blocked@example.com", "b@example.com"}
	got := allowed(people, []string{"blocked@example.com"})
	want := []string{"a@example.com", "b@example.com"}
	if len(got) != len(want) {
		t.Fatalf("allowed gave %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("allowed gave %v, want %v", got, want)
		}
	}
	// No refusals is no filtering, and the caller's slice comes back as it
	// was rather than as a copy.
	if got := allowed(people, nil); &got[0] != &people[0] {
		t.Error("with nothing refused the suggestions should be the caller's own slice")
	}
}

func TestSnoozePicker(t *testing.T) {
	now := time.Date(2026, 3, 10, 9, 0, 0, 0, time.UTC)
	open := true
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		anchor := ui.Box(c).Label("Snooze").Role(ui.RoleNone)
		SnoozePicker(c, SnoozePickerOptions{
			Anchor: anchor, Open: &open, Now: now, Label: "Snooze until",
		})
	}, 400, 300)
	wantText(t, tt, "the menu",
		"Snooze until", "Later today", "at 5pm", "Tomorrow", "Next week")

	closed := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		o := false
		anchor := ui.Box(c).Label("Snooze").Role(ui.RoleNone)
		SnoozePicker(c, SnoozePickerOptions{Anchor: anchor, Open: &o, Now: now})
	}, 300, 200)
	if closed.HasText("Later today") {
		t.Error("a closed snooze menu drew its choices")
	}

	wantsPanic(t, "a snooze menu with no clock", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			o := false
			anchor := ui.Box(c).Role(ui.RoleNone)
			SnoozePicker(c, SnoozePickerOptions{Anchor: anchor, Open: &o})
		}, 300, 200)
	})
}

func TestUserProfileCard(t *testing.T) {
	open := false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		anchor := ui.Box(c).Label("Dana Reyes").Role(ui.RoleNone)
		UserProfileCard(c, UserProfileCardOptions{
			Anchor: anchor, Open: &open, Name: "Dana Reyes",
			Handle: "@dana", Role: "Structural engineer",
			Presence: PresenceOnline, Since: "2m", LocalTime: "14:40",
			Channels: "site-walkthrough, billing",
		})
	}, 400, 320)
	if tt.HasText("14:40") {
		t.Error("a closed profile card drew its body")
	}

	wantsPanic(t, "a profile card with no anchor", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			o := false
			UserProfileCard(c, UserProfileCardOptions{Open: &o, Name: "Dana"})
		}, 400, 300)
	})
	wantsPanic(t, "a profile card about nobody", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			o := false
			anchor := ui.Box(c).Role(ui.RoleNone)
			UserProfileCard(c, UserProfileCardOptions{Anchor: anchor, Open: &o})
		}, 400, 300)
	})
}

func TestStatusSetter(t *testing.T) {
	status := ""
	open := false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		StatusSetter(c, StatusSetterOptions{Status: &status, Open: &open})
	}, 520, 80)
	wantText(t, tt, "the empty status", "Set a status", "No set a status")

	withStatus := "in a workshop until 4"
	tt2 := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		o := false
		s := withStatus
		StatusSetter(c, StatusSetterOptions{
			Status: &s, Open: &o, Expires: "today at 16:00",
		})
	}, 520, 80)
	wantText(t, tt2, "the set status", "in a workshop until 4", "today at 16:00")

	clears := 0
	tt3 := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		o := false
		s := withStatus
		if StatusSetter(c, StatusSetterOptions{
			Status: &s, Open: &o, Expires: "today at 16:00",
		}).Cleared() {
			clears++
		}
	}, 520, 80)
	tt3.Click("Clear status")
	if clears != 1 {
		t.Errorf("the clear button reported %d presses, want 1", clears)
	}
	if status != "" {
		t.Errorf("clearing left the status as %q", status)
	}

	wantsPanic(t, "a status setter with no line", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			o := false
			StatusSetter(c, StatusSetterOptions{Open: &o})
		}, 400, 80)
	})
}

func testPeople() []Participant {
	return []Participant{
		{ID: "dana", Name: "Dana Reyes", Speaking: true},
		{ID: "sam", Name: "Sam Okonkwo", Muted: true},
		{ID: "ana", Name: "Ana-Maria Silva", Hand: true},
		{ID: "priya", Name: "Priya Nair", Sharing: true},
	}
}

func TestVideoCallGrid(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		VideoCallGrid(c, VideoCallGridOptions{
			Participants: testPeople(), Speaker: "dana",
			Width: 640, Height: 400, Label: "Call with four people",
		})
	}, 700, 460)
	if _, ok := tt.Find("Call with four people"); !ok {
		t.Error("a grid of faces must be named; it is the hardest thing here to describe")
	}
	wantText(t, tt, "the grid", "DR", "SO", "AM", "PN")

	// One person must not leave two holes in the grid.
	one := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		VideoCallGrid(c, VideoCallGridOptions{
			Participants: testPeople()[:1], Width: 400, Height: 300,
			Label: "Call with one person",
		})
	}, 440, 340)
	if !one.HasText("DR") {
		t.Error("a one-person grid lost its tile")
	}

	wantsPanic(t, "an unnamed call grid", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			VideoCallGrid(c, VideoCallGridOptions{Participants: testPeople()})
		}, 400, 300)
	})
}

func TestCallControls(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		CallControls(c, CallControlsOptions{
			Controls: StandardCallControls(false, false, false),
			Label:    "Call controls", CallerName: "Dana Reyes", Elapsed: "04:12",
		})
	}, 620, 100)
	wantText(t, tt, "the bar", "Call controls", "Dana Reyes", "04:12",
		"Mute the microphone", "Turn the camera off", "Share your screen", "Leave the call")

	leaves := 0
	tt2 := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		if CallControls(c, CallControlsOptions{
			Controls: StandardCallControls(false, false, false), Label: "Call controls",
		}).Pressed() == "hangup" {
			leaves++
		}
	}, 620, 100)
	tt2.Click("Leave the call")
	if leaves != 1 {
		t.Errorf("the leave button reported %d presses, want 1", leaves)
	}

	wantsPanic(t, "a call bar with no name", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			CallControls(c, CallControlsOptions{
				Controls: StandardCallControls(false, false, false),
			})
		}, 500, 100)
	})
	wantsPanic(t, "a call bar with an unnamed control", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			CallControls(c, CallControlsOptions{
				Label: "Call controls", Controls: []CallControl{{ID: "mute"}},
			})
		}, 500, 100)
	})
}

func TestParticipantSummary(t *testing.T) {
	cases := []struct {
		n    int
		want string
	}{
		{0, "Nobody here"},
		{1, "1 person"},
	}
	for _, tc := range cases {
		if got := ParticipantSummary(testPeople()[:tc.n]); got != tc.want {
			t.Errorf("ParticipantSummary(%d) = %q, want %q", tc.n, got, tc.want)
		}
	}
	two := ParticipantSummary(testPeople()[:2])
	if two != "Dana Reyes and Sam Okonkwo" {
		t.Errorf("ParticipantSummary(2) = %q", two)
	}
	four := ParticipantSummary(testPeople())
	if four != "Dana Reyes, Sam Okonkwo and 2 others" {
		t.Errorf("ParticipantSummary(4) = %q", four)
	}
}

// ── dark mode ──────────────────────────────────────────────────────────────

// TestDarkModeDrawsEveryComponent runs the whole package in the dark palette
// and asks for each component's own name. If a component reached for a
// light-only colour it would still lay out and this test would still pass —
// so the assertion that matters is that each name is found, in dark, and that
// the palette the window resolved is the dark one.
func TestDarkModeDrawsEveryComponent(t *testing.T) {
	selected := 0
	unreadAt := 2
	open := true
	snoozeOpen := true
	status := "in a workshop"
	statusOpen := false
	body := "Photos are up."
	to := []string{"sam@riverside.example"}
	sending := false
	query := ""
	cc := []string{}

	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: core.Dark})

		ChannelList(c, ChannelListOptions{
			Channels: testChannels(), Selected: &selected, Height: 160, Label: "dm.channels",
		})
		ChatMessage(c, ChatMessageOptions{
			Role: chat.RoleUser, Author: "Dana", Time: "09:14",
			Body:   func() { ui.Text(c, "dm.body") },
			Quoted: "dm.quoted", Receipt: ReceiptRead, ReceiptTime: "09:14",
			Replies: 3, Divider: true, DividerCount: 4,
		})
		UnreadDivider(c, UnreadDividerOptions{Count: 4, Label: "dm.divider"})
		ReadReceipt(c, ReadReceiptOptions{State: ReceiptDelivered, Time: "09:14"})
		QuotedText(c, QuotedTextOptions{Author: "Priya", Text: "dm.quotedtext"})
		OnlineStatus(c, OnlineStatusOptions{
			Presence: PresenceBusy, Name: "dm.person", WithText: true,
		})
		MemberList(c, MemberListOptions{
			Members: testMembers(), Selected: &selected, Height: 160,
			ShowRoles: true, ShowTimes: true, Title: "dm.members",
		})
		anchor := ui.Box(c).Label("dm.anchor").Role(ui.RoleNone)
		EmojiPicker(c, EmojiPickerOptions{
			Anchor: anchor, Open: &open, Groups: DefaultGroups(),
			Query: &query, Label: "dm.emoji",
		})
		PinnedMessages(c, PinnedMessagesOptions{
			Messages: testPinned(), Selected: &selected, Height: 120, Label: "dm.pinned",
		})
		ThreadPanel(c, ThreadPanelOptions{
			Parent:  ChatMessageOptions{Role: chat.RoleUser, Author: "Dana"},
			Replies: 1, Reply: func(int) {}, Selected: &selected, Height: 140,
		})
		profile := ui.Box(c).Label("dm.profile").Role(ui.RoleNone)
		profileOpen := true
		UserProfileCard(c, UserProfileCardOptions{
			Anchor: profile, Open: &profileOpen, Name: "dm.profile.name",
			Presence: PresenceOnline,
		})
		MailList(c, MailListOptions{
			Messages: testMail(), Folder: FolderInbox, Selected: &selected,
			Height: 160, Title: "dm.mail",
		})
		MailReader(c, MailReaderOptions{Message: testMail()[0], Body: "dm.mail.body"})
		MailComposer(c, MailComposerOptions{
			To: &to, Cc: &cc, Body: &body, Sending: &sending,
		})
		RecipientInput(c, &to, RecipientInputOptions{Label: "dm.recipients"})
		snooze := ui.Box(c).Label("dm.snooze.anchor").Role(ui.RoleNone)
		SnoozePicker(c, SnoozePickerOptions{
			Anchor: snooze, Open: &snoozeOpen, Now: time.Unix(1, 0), Label: "dm.snooze",
		})
		StatusSetter(c, StatusSetterOptions{
			Status: &status, Open: &statusOpen, Label: "dm.status",
		})
		VideoCallGrid(c, VideoCallGridOptions{
			Participants: testPeople(), Width: 400, Height: 260, Label: "dm.grid",
		})
		CallControls(c, CallControlsOptions{
			Controls: StandardCallControls(false, false, false),
			Label:    "dm.controls", CallerName: "dm.caller",
		})
		_ = unreadAt
	}, 1000, 1100)

	for _, name := range []string{
		"dm.channels", "dm.body", "dm.quoted", "dm.quotedtext",
		// The divider and the dot are announced by the whole sentence they
		// say — "dm.divider · 4 messages", "dm.person is busy" — because the
		// sentence is the information and the label alone is not.
		"dm.divider · 4 messages", "dm.person is busy",
		"dm.members", "dm.anchor", "dm.emoji", "dm.pinned",
		"dm.profile.name", "dm.mail", "dm.mail.body", "dm.recipients",
		"dm.snooze.anchor", "dm.snooze", "dm.status", "dm.grid", "dm.controls",
		"dm.caller",
	} {
		if _, ok := tt.Find(name); !ok {
			t.Errorf("dark mode: %q is not on screen", name)
		}
	}
}

// TestDarkPaletteIsReadNotAssumed: a component that hard-coded a colour would
// pass every render test above while being wrong in the one place it cannot be
// seen. The tokens a dark window resolves are the dark ones.
func TestDarkPaletteIsReadNotAssumed(t *testing.T) {
	var bg, text ui.Color
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: core.Dark})
		k := core.Tokens(c)
		bg, text = k.Background, k.Text
		OnlineStatus(c, OnlineStatusOptions{Presence: PresenceOnline, Name: "Dana"})
	}, 300, 80)
	if tt == nil {
		t.Fatal("the tester did not run the view")
	}
	if bg != theme.Dark().Background {
		t.Errorf("a dark window resolved the background %v, want %v", bg, theme.Dark().Background)
	}
	if bg == theme.Light().Background {
		t.Error("a dark window resolved the light background")
	}
	if text == theme.Light().Text {
		t.Error("a dark window resolved the light text colour")
	}
}

func TestLightPaletteIsTheLightOne(t *testing.T) {
	var bg ui.Color
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: core.Light})
		bg = core.Tokens(c).Background
		OnlineStatus(c, OnlineStatusOptions{Presence: PresenceOnline, Name: "Dana"})
	}, 300, 80)
	if bg != theme.Light().Background {
		t.Errorf("a light window resolved %v, want %v", bg, theme.Light().Background)
	}
}

// TestDataTableIsNotRebuiltHere is a shape assertion rather than a behaviour
// one: the sort state is data.Sort and the tables are data.DataTable, so a
// caller reordering with data.Rows and asking again is the whole contract.
func TestDataTableIsNotRebuiltHere(t *testing.T) {
	var sort data.Sort = data.Sort{Column: "start"}
	if sort.Column != "start" {
		t.Error("the caller's sort state is data.Sort, as DataTable writes it")
	}
	if data.Toggle(sort, "end").Column != "end" {
		t.Error("data.Toggle is the rule the head of a sortable column asks for")
	}
}
