package pages

import (
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/chat"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/messaging"
	"github.com/HycJack/MintUI/ui/showcase"
	"github.com/HycJack/MintUI/ui/theme"
)

// Demo state outlives the frame: every demo below hands its component
// a pointer, and a pointer into a frame-local is a click the next
// frame undoes — a tab that will not switch, a dropdown that snaps
// shut, a slider that springs back.
var (
	messaging_selected = 0
	mailSel            = 0
	to                 = []string{"andre@riverside.clinic"}
	cc                 = []string{"dante@riverside.clinic"}
	messaging_body     = "Confirming 09:30. I will bring the controller."
	emptyTo            = []string{}
	emptyBody          = ""
	recipients         = []string{"andre@riverside.clinic"}
	ccOnly             = []string{}
	snoozeOpen         = true
	memberSel          = 0
	memberQuery        = ""
	cardOpen           = true
	set                = "in the workshop until 4"
	setOpen            = false
	messaging_none     = ""
	messagingNoneOpen  = false
	status             = "in the workshop until 4"
	messaging_open     = true
	pickerOpen         = true
	searchOpen         = true
	messaging_search   = "check"
	threadSel          = 1
	pinnedSel          = 1
	noneSel            = -1
)

// The page is one conversation and everything that can be attached to it: the
// channel list beside it, the turns in it, the thread under one of them, the
// mail that arrives on the same account, the people in it, the call that
// interrupts it, and the two panels that float above it.
//
// The order is the order a person's day goes in rather than the order the
// components are declared in: a channel, a turn in it, the mail behind it,
// the people, the call. A gallery read as an index tells a reader what exists;
// a gallery read as a thing tells them what the pieces are for.

func init() {
	showcase.Register(showcase.Page{
		Package: "messaging",
		Title:   "ui/messaging — 消息、邮件、人",
		Note:    "频道、消息、线程、邮件、收件人、稍后处理、成员、表情、通话",
		Width:   1000,
		Height:  4500,
		Want: []string{
			// channels and turns
			"North 分支", "site-visits", "Riverside Clinic", "standup",
			"客诉机不制冷，上午报的修。", "3 replies", "1 reply",
			"我 16:00 到，先看电控箱。", "Andre 把这条标成了置顶。",
			"New messages · 12 messages", "New messages · 1 message", "New messages",
			"Sending 14:02", "Delivered 14:02", "Read 14:02",
			"Failed to send 14:02", "Delivery status unavailable",
			// mail
			"Inbox", "Re: Friday's site visit", "This message has not been downloaded yet.",
			"Reply to Northgate Dental", "New message", "To", "Cc", "Send",
			"To andre@riverside.clinic", "No recipients",
			"andre@riverside.clinic", "1 added",
			// snoozing
			"Snooze until", "Later today", "at 5pm", "Tomorrow", "at 9am",
			"This weekend", "Saturday morning", "Next week", "Monday morning",
			// people
			"Members", "@andre", "@sam", "@dante", "engineer", "dispatcher",
			"Offline", "Away", "Online", "Busy",
			"Set a status", "in the workshop until 4", "in 3 hours",
			"No set a status", "StatusSetter with its editor open",
			// emoji
			"Emoji", "Search emoji", "Reactions", "Faces", "Status",
			"thumbs up", "red heart",
			// calls
			"Mute the microphone", "Unmute the microphone", "Leave the call",
			"Turn the camera off", "Turn the camera on", "Share your screen",
			"Stop sharing", "Sam Okonkwo is muted", "Dante Ruiz raised a hand",
			"Andre Thomson, Sam Okonkwo and 1 others",
			// threads and pins
			"Thread", "last reply 14:04", "我这边备件有，明早到。",
			"我把电控箱的照片发群里了。", "Pinned messages", "Nothing pinned",
			"You keep saying the panel is fine",
			// the pure functions
			"FirstUnread(7, 12)", "FirstUnread(11, 12)", "FirstUnread(-1, 12)",
			"DividerAt(7, 7, 12)", "ReceiptState", "Presence.String()",
			"DefaultGroups()", "DefaultSnoozes(now)", "1 person", "Andre and Sam",
		},
		Render: func(c *ui.Context) {
			messagingPage(c)
		},
	})
}

func messagingPage(c *ui.Context) {
	messagingChannelSection(c)
	messagingMailSection(c)
	messagingPeopleSection(c)
	messagingCallSection(c)
	messagingEmojiSection(c)
	messagingThreadSection(c)
	messagingPureSection(c)
}

// ── channels and turns ─────────────────────────────────────────────────────

func messagingChannelSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "频道 · ChannelList / ChatMessage / UnreadDivider / ReadReceipt")

	channels := []messaging.Channel{
		{ID: "north", Name: "North 分支", Topic: "客诉与回访", When: "14:02",
			Unread: 4, Mentions: 1, Pinned: true},
		{ID: "site", Name: "site-visits", Topic: "现场照片", When: "13:40",
			Unread: 12, Muted: true},
		{ID: "riverside", Name: "Riverside Clinic", Topic: "压缩机", When: "12:15",
			Draft: "客诉机不制冷，"},
		{ID: "standup", Name: "standup", Topic: "每天 09:30", When: "昨天",
			Unread: 1},
	}
	query := ""

	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).WidthPercent(38).Shrink(0).Gap(unit(c, 1.5)).Children(func() {
			showcase.Field(c, "ChannelList — 未读、@我、草稿、静音、置顶")
			messaging.ChannelList(c, messaging.ChannelListOptions{
				Channels: channels, Selected: &messaging_selected, Query: &query,
				Height: 200,
			})
			ui.Text(c, "Shown = 4 · Query 为空时不过滤").TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize))
		})

		ui.Column(c).WidthPercent(60).Shrink(0).Gap(unit(c, 1.5)).Children(func() {
			showcase.Field(c, "ChatMessage — 分割线、回执、线程链接都由这一件加上")
			ui.Column(c).FillWidth().Gap(unit(c, 1.5)).Children(func() {
				messaging.ChatMessage(c, messaging.ChatMessageOptions{
					Role: chat.RoleAssistant, Author: "Andre Thomson", Time: "14:01",
					Body:    func() { ui.Text(c, "客诉机不制冷，上午报的修。") },
					Replies: 3,
				})
				messaging.ChatMessage(c, messaging.ChatMessageOptions{
					Role: chat.RoleUser, Author: "你", Time: "14:02",
					Quoted: "客诉机不制冷，上午报的修。", QuotedAuthor: "Andre Thomson",
					Selected: true, Receipt: messaging.ReceiptDelivered,
					ReceiptTime: "14:02",
					Body:        func() { ui.Text(c, "我 16:00 到，先看电控箱。") },
					Replies:     1,
				})
				messaging.ChatMessage(c, messaging.ChatMessageOptions{
					Role: chat.RoleSystem, Time: "14:03",
					Body: func() { ui.Text(c, "Andre 把这条标成了置顶。") },
				})
			})
		})
	})

	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).WidthPercent(share(3)).Shrink(0).Gap(unit(c, 1.5)).Children(func() {
			showcase.Field(c, "UnreadDivider — 线在第一条没读的上方")
			messaging.UnreadDivider(c, messaging.UnreadDividerOptions{Count: 12})
			messaging.UnreadDivider(c, messaging.UnreadDividerOptions{
				Count: 1, Since: 40 * time.Minute,
			})
			messaging.UnreadDivider(c, messaging.UnreadDividerOptions{Count: 0})
		})
		ui.Column(c).WidthPercent(share(3)).Shrink(0).Gap(unit(c, 1.5)).Children(func() {
			showcase.Field(c, "QuotedText — 规则挂在气泡同一条边上")
			messaging.QuotedText(c, messaging.QuotedTextOptions{
				Author: "Andre Thomson", Text: "客诉机不制冷，上午报的修。", Lines: 2,
			})
			messaging.QuotedText(c, messaging.QuotedTextOptions{
				Author: "你", Text: "我 16:00 到，先看电控箱。", Side: ui.End,
			})
		})
		ui.Column(c).WidthPercent(share(3)).Shrink(0).Gap(unit(c, 1.5)).Children(func() {
			showcase.Field(c, "ReadReceipt — 形状不同，不是只有颜色不同")
			for _, st := range []messaging.Receipt{
				messaging.ReceiptSending, messaging.ReceiptDelivered,
				messaging.ReceiptRead, messaging.ReceiptFailed,
			} {
				ui.Row(c).Gap(unit(c, 1.5)).AlignItems(ui.Center).Children(func() {
					messaging.ReadReceipt(c, messaging.ReadReceiptOptions{
						State: st, Time: "14:02",
					})
				})
			}
			messaging.ReadReceipt(c, messaging.ReadReceiptOptions{
				State: messaging.ReceiptRead, Time: "14:02", Muted: true,
			})
			messaging.ReadReceipt(c, messaging.ReadReceiptOptions{State: 99})
		})
	})
}

// ── mail ───────────────────────────────────────────────────────────────────

func messagingMailSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "邮件 · MailList / MailReader / MailComposer / RecipientInput / SnoozePicker")

	messages := []messaging.Message{
		{ID: "m1", From: "Northgate Dental", Subject: "Re: Friday's site visit",
			Preview: "Confirming 09:30, the same unit is still down…", When: "14:06",
			Unread: true, Important: true, HasAttachments: true,
			AttachmentCount: 2, Folder: messaging.FolderInbox},
		{ID: "m2", From: "Dante Audio", Subject: "Your order has shipped",
			Preview: "Two USB interfaces and the cable…", When: "11:20",
			HasAttachments: true, Folder: messaging.FolderInbox},
		{ID: "m3", From: "Riverside Clinic", Subject: "回访记录", Preview: "李工说电控箱已经换过。",
			When: "昨天", Folder: messaging.FolderInbox},
		{ID: "m4", From: "Payroll", Subject: "Payslip for November", Preview: "…",
			When: "上周", Folder: messaging.FolderInbox},
	}

	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).WidthPercent(56).Shrink(0).Gap(unit(c, 1.5)).Children(func() {
			showcase.Field(c, "MailList — 表头留在滚动之上，列宽不让价格折行")
			messaging.MailList(c, messaging.MailListOptions{
				Messages: messages, Folder: messaging.FolderInbox,
				Selected: &mailSel, Height: 232,
			})
			ui.Text(c, "Shown = 4 · Unread = 1").TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize))
		})

		ui.Column(c).WidthPercent(42).Shrink(0).Gap(unit(c, 1.5)).Children(func() {
			showcase.Field(c, "MailReader — 正文是纯文本，附件是调用方的按钮")
			messaging.MailReader(c, messaging.MailReaderOptions{
				Message: messages[0],
				Body: "Hi,\n\nConfirming 09:30 on Friday. The same unit is still down and " +
					"the site manager has asked for a two-hour window.\n\nThe two photos " +
					"are in the attachments.\n\nThanks,\nNorthgate Dental",
				Snoozed: "Tomorrow, 9am",
				Actions: func() {
					showcase.Stack(c, 1, func() {
						input.Button(c, "Archive", input.ButtonOptions{})
						input.Button(c, "Reply", input.ButtonOptions{Primary: true})
					})
				},
			})
			messaging.MailReader(c, messaging.MailReaderOptions{
				Message: messaging.Message{ID: "m5", From: "Payroll",
					Subject: "Payslip for November", When: "上周"},
			})
		})
	})

	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).WidthPercent(62).Shrink(0).Gap(unit(c, 1.5)).Children(func() {
			showcase.Field(c, "MailComposer — 收件人、抄送、正文、发送，全是调用方的指针")
			messaging.MailComposer(c, messaging.MailComposerOptions{
				To: &to, Cc: &cc, Body: &messaging_body,
				People:    []string{"andre@riverside.clinic", "sam@riverside.clinic", "dante@riverside.clinic"},
				Suggested: "On Friday, 09:30 —\n",
				Sending:   new(false),
				Title:     "Reply to Northgate Dental",
			})
			showcase.Field(c, "MailComposer with no Cc and nothing filled in — 发送是禁用的")
			messaging.MailComposer(c, messaging.MailComposerOptions{
				To: &emptyTo, Body: &emptyBody, Sending: new(false),
			})
		})

		ui.Column(c).WidthPercent(38).Shrink(0).Gap(unit(c, 1.5)).Children(func() {
			showcase.Field(c, "RecipientInput — 不许加的地址连建议都不出现")
			messaging.RecipientInput(c, &recipients, messaging.RecipientInputOptions{
				Label:       "To",
				People:      []string{"andre@riverside.clinic", "sam@riverside.clinic", "outside@example.com"},
				Disallowed:  []string{"outside@example.com"},
				Placeholder: "name or address",
			})
			messaging.RecipientInput(c, &ccOnly, messaging.RecipientInputOptions{
				Label: "Cc", People: []string{"dante@riverside.clinic"}, Muted: true,
			})

			showcase.Field(c, "SnoozePicker — 菜单挂在邮件自己的按钮上")
			now := time.Date(2024, 11, 8, 14, 6, 0, 0, time.UTC)
			anchor := input.Button(c, "Snooze", input.ButtonOptions{Label: "Snooze until"})
			messaging.SnoozePicker(c, messaging.SnoozePickerOptions{
				Anchor: anchor, Open: &snoozeOpen, Now: now,
			})
		})
	})
}

// ── people ─────────────────────────────────────────────────────────────────

func messagingPeopleSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "人 · MemberList / OnlineStatus / UserProfileCard / StatusSetter / ParticipantSummary")

	members := []messaging.Member{
		{ID: "u1", Name: "Andre Thomson", Handle: "@andre", Role: "engineer",
			Presence: messaging.PresenceAway, Since: "12 minutes ago",
			LocalTime: "16:04", Title: "Field engineer"},
		{ID: "u2", Name: "Sam Okonkwo", Handle: "@sam", Role: "dispatcher",
			Presence: messaging.PresenceOnline, LocalTime: "16:04"},
		{ID: "u3", Name: "Dante Ruiz", Handle: "@dante", Role: "engineer",
			Presence: messaging.PresenceBusy, Since: "since 13:50",
			LocalTime: "09:04"},
		{ID: "u4", Name: "Wen Zhao", Handle: "@wen", Role: "coordinator",
			Presence: messaging.PresenceOffline, LocalTime: "08:04"},
	}

	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).WidthPercent(52).Shrink(0).Gap(unit(c, 1.5)).Children(func() {
			showcase.Field(c, "MemberList — 名字、角色、当地的时间，四样东西对齐")
			messaging.MemberList(c, messaging.MemberListOptions{
				Members: members, Selected: &memberSel, Query: &memberQuery,
				Height: 232, ShowRoles: true, ShowTimes: true,
			})
			ui.Text(c, "Shown = 4").TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize))
		})

		ui.Column(c).WidthPercent(48).Shrink(0).Gap(unit(c, 1.5)).Children(func() {
			showcase.Field(c, "OnlineStatus — 四种在场，离线也占一个位子")
			ui.Row(c).Gap(unit(c, 2)).Children(func() {
				for _, p := range []messaging.Presence{
					messaging.PresenceOffline, messaging.PresenceAway,
					messaging.PresenceOnline, messaging.PresenceBusy,
				} {
					ui.Column(c).AlignItems(ui.Center).Gap(unit(c, 0.5)).Children(func() {
						messaging.OnlineStatus(c, messaging.OnlineStatusOptions{
							Presence: p, Name: "Andre", WithText: true,
						})
					})
				}
			})
			ui.Row(c).Gap(unit(c, 2)).Children(func() {
				for _, p := range []messaging.Presence{
					messaging.PresenceOffline, messaging.PresenceAway,
					messaging.PresenceOnline, messaging.PresenceBusy,
				} {
					messaging.OnlineStatus(c, messaging.OnlineStatusOptions{
						Presence: p, Name: p.String(),
					})
				}
			})
			showcase.Field(c, "UserProfileCard — 指针停在名字上才出来的小卡")
			anchor := input.Button(c, "Andre Thomson", input.ButtonOptions{})
			messaging.UserProfileCard(c, messaging.UserProfileCardOptions{
				Anchor: anchor, Open: &cardOpen, Name: "Andre Thomson",
				Handle: "@andre", Role: "field engineer",
				Presence: messaging.PresenceAway, Since: "12 minutes ago",
				LocalTime: "16:04 in Lisbon", Channels: "#north · #site-visits",
			})
		})
	})

	showcase.Field(c, "StatusSetter — 那一行是控件，编辑器是浮层")
	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).WidthPercent(62).Shrink(0).Gap(unit(c, 1)).Children(func() {
			messaging.StatusSetter(c, messaging.StatusSetterOptions{
				Status: &set, Open: &setOpen, Expires: "in 3 hours",
				Title: "Set a status",
			})
		})
		ui.Column(c).WidthPercent(38).Shrink(0).Gap(unit(c, 1)).Children(func() {
			messaging.StatusSetter(c, messaging.StatusSetterOptions{
				Status: &messaging_none, Open: &messagingNoneOpen,
			})
		})
	})
	// The editor is a dialog, so it owns the window: a scrim over this page
	// would dim every other component on it. It is drawn into a window of its
	// own instead, which is a real render of the real panel.
	messagingStatusEditorShot(c)
}

// messagingStatusEditorShot draws an messaging_open StatusSetter into a window of its own
// and puts the result on the page as a picture, for the reason the lightbox
// has on the media page.
func messagingStatusEditorShot(c *ui.Context) {
	mode := core.Light
	if core.IsDark(c) {
		mode = core.Dark
	}
	shot := ui.NewTester(func(inner *ui.Context) {
		core.Use(inner, core.Settings{Mode: mode})
		ui.Column(inner).Fill().Padding(core.Density(inner).Unit() * 2).
			Gap(core.Density(inner).Unit() * 2).Children(func() {
			for _, line := range []string{"Riverside Clinic · North 分支", "客诉机不制冷，上午报的修。"} {
				ui.Text(inner, line).TextColor(core.Tokens(inner).TextMuted).
					FontSize(core.FontSize(inner, theme.MetaSize))
			}
			messaging.StatusSetter(inner, messaging.StatusSetterOptions{
				Status: &status, Open: &messaging_open, Title: "Set a status",
				Placeholder: "in the workshop until 4",
			})
		})
	}, 460, 260).Image()

	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		display.Image(c, ui.NewBitmap(shot), display.ImageOptions{
			Width: 460, Height: 260, Radius: theme.SmallRadius,
			Name: "StatusSetter with its editor open",
		})
	})
}

// ── calls ──────────────────────────────────────────────────────────────────

func messagingCallSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "通话 · CallControls / VideoCallGrid / ParticipantSummary")

	participants := []messaging.Participant{
		{ID: "u1", Name: "Andre Thomson", Speaking: true, Video: messagingPicture(320, 240, 1)},
		{ID: "u2", Name: "Sam Okonkwo", Muted: true, Video: messagingPicture(320, 240, 2)},
		{ID: "u3", Name: "Dante Ruiz", Hand: true, Pinned: true},
	}

	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).WidthPercent(57).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "VideoCallGrid — 一个方块一个人，缺席的人也在")
			messaging.VideoCallGrid(c, messaging.VideoCallGridOptions{
				// The height is given rather than left to the grid's own Fill():
				// in a column whose height comes from its content, a child at
				// 100% has nothing to be 100% of, and the grid comes out
				// overlapping the section heading above it.
				Participants: participants, Columns: 2, Label: "North 分支",
				Height: 440,
			})
			ui.Text(c, messaging.ParticipantSummary(participants)).TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).Bold()
		})

		ui.Column(c).WidthPercent(41).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "CallControls — 静音中、摄像头关着、正在共享")
			messaging.CallControls(c, messaging.CallControlsOptions{
				Controls: messaging.StandardCallControls(true, false, true),
				Label:    "Call controls", CallerName: "North 分支", Elapsed: "14:07",
			})
			showcase.Field(c, "CallControls — 什么都在默认状态")
			messaging.CallControls(c, messaging.CallControlsOptions{
				Controls: messaging.StandardCallControls(false, true, false),
				Label:    "Call controls", CallerName: "North 分支", Elapsed: "0:04",
			})
			showcase.Field(c, "CallControls — 通话自己的那套按钮")
			messaging.CallControls(c, messaging.CallControlsOptions{
				Label: "Call controls", CallerName: "Northgate Dental", Elapsed: "2m 14s",
				Controls: []messaging.CallControl{
					{ID: "mute", Label: "Unmute the microphone", On: true},
					{ID: "camera", Label: "Turn the camera off", On: true},
					{ID: "record", Label: "Start recording"},
					{ID: "notes", Label: "Take a note"},
					{ID: "hangup", Label: "Leave the call", Danger: true},
				},
			})
		})
	})
}

// ── emoji ──────────────────────────────────────────────────────────────────

func messagingEmojiSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "表情 · EmojiPicker / DefaultGroups")

	query := ""
	ui.Box(c).FillWidth().Height(unit(c, 62)).Radius(theme.ControlRadius).
		Background(k.Background).Border(theme.BorderWidth, k.Border).Clip().Children(func() {
		ui.Box(c).FillWidth().Padding(unit(c, 1.5), unit(c, 2)).Children(func() {
			ui.Text(c, "面板挂在输入框的按钮下面；查询为空时先出最近用的").TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize))
		})
		anchor := input.Button(c, "🙂", input.ButtonOptions{Label: "Emoji"})
		messaging.EmojiPicker(c, messaging.EmojiPickerOptions{
			Anchor: anchor, Open: &pickerOpen, Groups: messaging.DefaultGroups(),
			Recent: []messaging.Emoji{
				{Glyph: "👍", Name: "thumbs up"}, {Glyph: "🙏", Name: "thank you"},
				{Glyph: "🔥", Name: "on fire"},
			},
			Query: &query, Width: 320,
		})
	})

	showcase.Field(c, "EmojiPicker with a query — Searches 说了这个搜索找到了几组")
	ui.Box(c).FillWidth().Height(unit(c, 52)).Radius(theme.ControlRadius).
		Background(k.Background).Border(theme.BorderWidth, k.Border).Clip().Children(func() {
		ui.Box(c).FillWidth().Padding(unit(c, 1.5), unit(c, 2)).Children(func() {
			ui.Text(c, "query = \"check\"").TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize))
		})
		anchor := input.Button(c, "🔍", input.ButtonOptions{Label: "Emoji"})
		messaging.EmojiPicker(c, messaging.EmojiPickerOptions{
			Anchor: anchor, Open: &searchOpen, Groups: messaging.DefaultGroups(),
			Query: &messaging_search,
		})
	})
}

// ── threads ────────────────────────────────────────────────────────────────

func messagingThreadSection(c *ui.Context) {
	showcase.Section(c, "线程与置顶 · ThreadPanel / PinnedMessages")

	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).WidthPercent(57).Shrink(0).Gap(unit(c, 1.5)).Children(func() {
			showcase.Field(c, "ThreadPanel — 顶上是那条被回复的消息，下面是回复")
			messaging.ThreadPanel(c, messaging.ThreadPanelOptions{
				Height: 260, Selected: &threadSel, Updated: "last reply 14:04",
				Parent: messaging.ChatMessageOptions{
					Role: chat.RoleAssistant, Author: "Andre Thomson", Time: "14:01",
					Body: func() { ui.Text(c, "客诉机不制冷，上午报的修。") },
				},
				Replies: 3,
				Reply: func(i int) {
					role, author, text := chat.RoleAssistant, "Sam Okonkwo",
						"我这边备件有，明早到。"
					if i == 1 {
						role, author, text = chat.RoleUser, "你", "那就明早 09:30。"
					}
					if i == 2 {
						role, author, text = chat.RoleAssistant, "Dante Ruiz",
							"我把电控箱的照片发群里了。"
					}
					messaging.ChatMessage(c, messaging.ChatMessageOptions{
						Role: role, Author: author, Time: "14:03",
						Body: func() {
							ui.Text(c, text)
						},
					})
				},
			})
		})

		ui.Column(c).WidthPercent(41).Shrink(0).Gap(unit(c, 1.5)).Children(func() {
			showcase.Field(c, "PinnedMessages — 置顶的是记录，不是消息")
			messaging.PinnedMessages(c, messaging.PinnedMessagesOptions{
				Height: 120, Selected: &pinnedSel,
				Messages: []messaging.PinnedMessage{
					{ID: "p1", Author: "Sam Okonkwo", Text: "You keep saying the panel is fine",
						When: "yesterday"},
					{ID: "p2", Author: "Andre Thomson", Text: "Site access is 09:30 to 16:00",
						When: "Monday"},
				},
			})
			messaging.PinnedMessages(c, messaging.PinnedMessagesOptions{
				Height: 90, Selected: &noneSel,
			})
		})
	})
}

// ── the pure functions ─────────────────────────────────────────────────────

func messagingPureSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "纯函数 · 不需要窗口就能算的那几个")

	line := func(call, result string) {
		ui.Row(c).FillWidth().Gap(unit(c, 2)).AlignItems(ui.Start).Children(func() {
			ui.Text(c, call).TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).Width(210).Shrink(0)
			ui.Text(c, result).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.RowSize)).Grow(1)
		})
	}
	states := ""
	for i, s := range []messaging.Receipt{
		messaging.ReceiptSending, messaging.ReceiptDelivered,
		messaging.ReceiptRead, messaging.ReceiptFailed,
	} {
		mark, ok := messaging.ReceiptState(s)
		if i > 0 {
			states += "   "
		}
		states += s.String() + " → " + mark
		if !ok {
			states += " (nothing)"
		}
	}
	presence := ""
	for i, p := range []messaging.Presence{
		messaging.PresenceOffline, messaging.PresenceAway,
		messaging.PresenceOnline, messaging.PresenceBusy,
	} {
		if i > 0 {
			presence += "   "
		}
		presence += p.String()
	}
	divider := ""
	for i := range 12 {
		if messaging.DividerAt(i, 7, 12) {
			divider = "true at i = " + itoaPage(i)
		}
	}
	snoozes := ""
	for i, s := range messaging.DefaultSnoozes(time.Date(2024, 11, 8, 14, 6, 0, 0, time.UTC)) {
		if i > 0 {
			snoozes += "   "
		}
		snoozes += s.Label
	}

	ui.Row(c).FillWidth().Gap(unit(c, 4)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 1)).Children(func() {
			line("FirstUnread(7, 12)", itoaPage(messaging.FirstUnread(7, 12))+
				"  ← 第 8 条是第一条没读的")
			line("FirstUnread(11, 12)", itoaPage(messaging.FirstUnread(11, 12))+"  ← 全部读过了")
			line("FirstUnread(-1, 12)", itoaPage(messaging.FirstUnread(-1, 12))+"  ← 一条都没读")
			line("FirstUnread(0, 0)", itoaPage(messaging.FirstUnread(0, 0))+"  ← 空列表")
			line("DividerAt(7, 7, 12)", divider)
			line("ReceiptState", states[:len(states)/2])
			line("", states[len(states)/2:])
			line("Receipt(99).String()", messaging.Receipt(99).String()+
				"  ← 认不出的状态只能给一个像样的值")
			line("Presence.String()", presence)
		})
		ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 1)).Children(func() {
			line("DefaultGroups()", emojiNames(messaging.DefaultGroups()))
			line("DefaultSnoozes(now)", snoozes[:len(snoozes)/2])
			line("", snoozes[len(snoozes)/2:])
			line("ParticipantSummary(1)", messaging.ParticipantSummary(
				[]messaging.Participant{{Name: "Andre Thomson"}}))
			line("ParticipantSummary(2)", messaging.ParticipantSummary(
				[]messaging.Participant{{Name: "Andre"}, {Name: "Sam"}}))
		})
	})
}

func emojiNames(groups []messaging.EmojiGroup) string {
	out := ""
	for i, g := range groups {
		if i > 0 {
			out += "   "
		}
		out += g.Name
	}
	return out
}

func itoaPage(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

// messagingPicture is a picture made in memory, so the call grid has frames
// without a camera: see the media page for the same helper's reasoning.
func messagingPicture(w, h, seed int) *ui.Bitmap {
	return mediaPicture(w, h, seed)
}
