package pages

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/account"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/showcase"
	"github.com/HycJack/MintUI/ui/theme"
)

// Demo state outlives the frame: every demo below hands its component
// a pointer, and a pointer into a frame-local is a click the next
// frame undoes — a tab that will not switch, a dropdown that snaps
// shut, a slider that springs back.
var (
	accountCode = "4182"
	empty       = ""
	accent      = ui.Hex("#1d4ed8")
	found       = ""
	picker      = ui.Hex("#0f766e")
	quiet       = false
	gone        = ""
	none        = []string{}
	last        = 2
	done        = false
	beyond      = 7
	kind        = "Critical"
)

// The page is one person's account, from the form they sign in with to the
// keyboard they work on, drawn with every component this package has.
//
// It is arranged by what the person is doing rather than by what the files
// are called: who am I, which account am I, what does this app look like,
// what am I allowed, what keys do I press, what does the app do on its first
// run. Every one of those is a screen somebody has looked at, which is the
// only reason a gallery is worth scrolling.
//
// Everything on the page is fixed: the same account, the same three keys, the
// same four releases. A page that drew differently each time it was opened
// could not be compared against the last one, and a screenshot that changed
// between runs is a screenshot nobody reviews.

// monoFamily is the monospaced face. It is used here for the table of pure
// functions and for nothing else: every value in that table is something a
// person compares character by character, and a proportional face makes "1",
// "l" and "I" the same width.
const monoFamily = "monospace"

func init() {
	showcase.Register(AccountPage())
}

// AccountPage is the gallery's page for ui/account.
func AccountPage() showcase.Page {
	return showcase.Page{
		Package: "account",
		Title:   "ui/account — 账号、身份与设置",
		Note:    "登录、档案、切换器、设置、密钥、会话、用量、快捷键、首次运行",
		Width:   1000,
		Height:  4960,
		Want: []string{
			// LoginForm and the one-time code
			"Sign in", "Northgate · the clinic account", "Forgot password?",
			"That does not look like an email address",
			"Passwords are at least 8 characters", "Verification code",
			// ProfileEditor
			"Rosa Vidal", "Head technician", "Display name", "Save changes",
			"A name is not optional",
			// the three switchers
			"Accounts", "Add an account", "rosa@northgate.health", "Rosa Vidal",
			"Workspaces", "6 seats", "South branch", "Locked",
			"Sign out", "Account settings", "Delete this account",
			"Single sign-on",
			// SettingsLayout and SettingsSearch
			"Settings sections", "Notifications", "Appearance",
			"Sign in with a key instead", "More below",
			"Light, dark or whatever the desktop is doing",
			// ThemeSelector and AccentColorPicker
			"Match system", "Light", "Dark", "Accent colour", "Filled", "Outline", "Tag",
			// keys, sessions, usage
			"API keys", "Create key", "Revoke", "Read only · 2 h ago",
			"Full access · never used", "Signed-in devices",
			"This device", "Expired", "Sign out everywhere",
			"Usage this month", "Requests", "Not metered", "Upgrade plan",
			"1,840 / 2,000",
			// shortcuts
			"Keyboard shortcuts", "Go to search", "General", "Windows",
			"Open the callback", "Press keys",
			// first run
			"Welcome to Callbacks", "Skip setup", "Finish", "Back",
			"Choose what reaches you", "Ask anything", "2 of 4", "Got it",
			// the two that speak
			"2.4.0", "Feedback", "What is going wrong?", "Send",
			"The log callback drawer now filters by branch, not by machine.",
			// pure functions
			`MaskKey("cb_live_4f9a2b7c1d8e0a5b")`,
			`FormatChord([]string{"command", "shift", "k"})`,
			`NormalKey("Command")`,
			`NormalKeys([]string{"cmd", "", "esc"})`,
			`ValidateLogin("rosa@northgate.health", "northgate-2026")`,
			`QuotaTone(1840, 2000)`,
			`QuotaTone(1900, 2000)`,
		},
		Render: func(c *ui.Context) {
			accountPage(c)
		},
	}
}

func accountPage(c *ui.Context) {
	identitySection(c)
	switchingSection(c)
	appearanceSection(c)
	accessSection(c)
	shortcutSection(c)
	firstRunSection(c)
	speakingSection(c)
	pureSection(c)
}

// ── who am I ─────────────────────────────────────────────────────────────

// identitySection is the two forms an account is written with — signing in and
// editing yourself — and the one control in between them that a form cannot
// express: a code that has to be typed once.
func identitySection(c *ui.Context) {
	showcase.Section(c, "身份 · LoginForm / ProfileEditor / TwoFactorInput")

	email := showcase.State(c, "account.133.email", "rosa@northgate.health")
	password := showcase.State(c, "account.133.password", "northgate-2026")
	remember := showcase.State(c, "account.134.remember", true)
	revealed := showcase.State(c, "account.134.revealed", false)
	submit := showcase.State(c, "account.134.submit", false)
	// Two mistakes on purpose: ValidateLogin answers both at once, so this is
	// the frame a person is told about the address and the password together
	// rather than one of them and then the other.
	emailErr, passErr := account.ValidateLogin("rosa@", "1234")

	name := showcase.State(c, "account.140.name", "Rosa Vidal")
	address := showcase.State(c, "account.140.address", "rosa@northgate.health")
	saved := showcase.State(c, "account.141.saved", false)
	readOnly := showcase.State(c, "account.141.readOnly", true)

	ui.Row(c).FillWidth().Gap(unit(c, 6)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).WidthPercent(38).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "LoginForm — 两条错误同时画出来")
			login := account.LoginForm(c, account.LoginFormOptions{
				Email: email, Password: password, Remember: remember,
				Revealed: revealed, Submit: submit,
				EmailError: emailErr, PasswordError: passErr,
				Title:    "Sign in",
				Subtitle: "Northgate · the clinic account",
				Forgot:   func() {},
			})
			resultLine(c, "LoginForm.Valid()", fmtBool(login.Valid()))

			showcase.Field(c, "TwoFactorInput — 六个格子，填了几个格子是答案")
			two := account.TwoFactorInput(c, account.TwoFactorInputOptions{
				Value: &accountCode, Length: 6,
			})
			resultLine(c, "TwoFactorInput.Filled() · Complete()",
				fmtInt(two.Filled())+" / 6 · "+fmtBool(two.Complete()))
			// The same control, empty and masked: a code somebody is typing
			// in public is the reason the boxes are boxes at all.
			account.TwoFactorInput(c, account.TwoFactorInputOptions{
				Value: &empty, Length: 4, Masked: true,
			})
		})

		ui.Column(c).WidthPercent(60).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "ProfileEditor — 头像在上，字段在下；邮箱只读")
			profile := account.ProfileEditor(c, account.ProfileEditorOptions{
				Name: name, Email: address, Save: saved,
				Role: "Head technician", ReadOnlyEmail: *readOnly,
			})
			resultLine(c, "ProfileEditor.Saved()", fmtBool(profile.Saved()))
			showcase.Field(c, "ProfileEditor — 没有照片时是首字母，不是空框")
			account.ProfileEditor(c, account.ProfileEditorOptions{
				Name: name, Email: address, Save: saved,
				NameError: "A name is not optional",
			})
		})
	})
}

// ── which account am I ───────────────────────────────────────────────────

// switchingSection is the three controls that answer "who is this", all of
// them a trigger with a panel hanging under it.
//
// Each gets a stage with blank room under the trigger rather than a plain
// row: an anchored panel is drawn in the window's own layer, so it lands on
// top of whatever is below, and three of them side by side would cover each
// other. One stage each, with the room the panel needs, is the only
// arrangement in which all three can be read at once.
func switchingSection(c *ui.Context) {
	showcase.Section(c, "切换 · AccountSwitcher / WorkspaceSwitcher / UserMenu")

	accounts := []account.Account{
		{ID: "acct-1", Name: "Rosa Vidal", Email: "rosa@northgate.health", Kind: "Work"},
		{ID: "acct-2", Name: "Rosa Vidal", Email: "rosa@harbour.cafe", Kind: "Personal"},
		{ID: "acct-3", Name: "R. Vidal", Email: "r.vidal@northgate.health", Kind: "Work"},
	}
	stage(c, "AccountSwitcher", "当前账号是一行名字和地址，不是一个选中的值", func() {
		chosen := showcase.State(c, "account.204.chosen", "acct-1")
		switched := showcase.State(c, "account.204.switched", "")
		open := showcase.State(c, "account.204.open", true)
		account.AccountSwitcher(c, account.AccountSwitcherOptions{
			Accounts: accounts, Selected: chosen, Open: open,
			Switched: switched, AddAccount: func() {},
		})
		resultLine(c, "Switched()", quoteOrEmpty(*switched))
	})

	workspaces := []account.Workspace{
		{ID: "ws-1", Name: "North branch", Seats: 6},
		{ID: "ws-2", Name: "Harbour", Seats: 2},
		{ID: "ws-3", Name: "South branch", Seats: 0, Locked: true},
	}
	stage(c, "WorkspaceSwitcher", "锁住的那一行还在，只是不给点", func() {
		chosen := showcase.State(c, "account.218.chosen", "ws-1")
		switched := showcase.State(c, "account.218.switched", "")
		open := showcase.State(c, "account.218.open", true)
		account.WorkspaceSwitcher(c, account.WorkspaceSwitcherOptions{
			Workspaces: workspaces, Selected: chosen, Open: open,
			Switched: switched,
		})
		resultLine(c, "Switched()", quoteOrEmpty(*switched))
	})

	stage(c, "UserMenu", "触发器只有一个头像，名字在菜单里面", func() {
		open := showcase.State(c, "account.227.open", true)
		chosenID := showcase.State(c, "account.227.chosenID", "")
		account.UserMenu(c, account.UserMenuOptions{
			Profile: account.Profile{
				ID: "acct-1", Name: "Rosa Vidal",
				Email: "rosa@northgate.health", Role: "Head technician",
			},
			Open: open, Chosen: chosenID, SignOut: func() {},
			Items: []account.UserMenuItem{
				{ID: "profile", Label: "Account settings"},
				{ID: "keys", Label: "API keys"},
				{ID: "billing", Label: "Billing"},
				{ID: "delete", Label: "Delete this account", Danger: true},
				{ID: "sso", Label: "Single sign-on", Disabled: true},
			},
		})
		resultLine(c, "Chosen()", quoteOrEmpty(*chosenID))
	})
}

// ── what does this window look like ──────────────────────────────────────

// appearanceSection is a settings window, the search over it, and the two
// controls underneath that decide what the window is made of.
func appearanceSection(c *ui.Context) {
	showcase.Section(c, "设置 · SettingsLayout / SettingsSearch / ThemeSelector / AccentColorPicker")

	sections := []account.Setting{
		{ID: "appearance", Name: "Appearance", Hint: "Light, dark or whatever the desktop is doing",
			Keywords: []string{"theme", "dark", "colour"}},
		{ID: "account", Name: "Notifications", Hint: "Which of these reach you, and where",
			Keywords: []string{"email", "push"}},
		{ID: "security", Name: "Security", Hint: "Sign in with a key instead",
			Keywords: []string{"password", "2fa", "sso"}},
		{ID: "billing", Name: "Billing", Hint: "Plan, seats and invoices"},
	}
	// The search narrows the list to the one section that mentions a key,
	// and the pane shows that same section: a list that filters the left
	// column while the right one shows something else is a settings window
	// with two answers on it at once.
	open := showcase.State(c, "account.266.open", "security")
	chosen := showcase.State(c, "account.266.chosen", "")
	query := showcase.State(c, "account.266.query", "key")
	mode := "system"

	ui.Row(c).FillWidth().Gap(unit(c, 6)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).WidthPercent(58).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "SettingsLayout — 左边是节，右边是打开的那一节")
			// The pane fills both ways, so it is given a frame to fill: a
			// split laid out straight into a page column would take the
			// column's whole height, which is the page's whole height.
			ui.Box(c).FillWidth().Height(unit(c, 36)).Radius(theme.ControlRadius).
				Border(theme.BorderWidth, core.Tokens(c).Border).Children(func() {
				account.SettingsLayout(c, account.SettingsLayoutOptions{
					Sections: sections, Selected: open, Chosen: chosen,
					Query: query, Searchable: true,
					Section: func(id string) { settingsPanel(c, id) },
				})
			})

			showcase.Field(c, "SettingsSearch — 同样的规则，单独一个字段加结果")
			search := account.SettingsSearch(c, account.SettingsSearchOptions{
				Settings: sections, Query: &found, Chosen: chosen, Limit: 2,
				Placeholder: "Search 12 settings",
			})
			resultLine(c, "SettingsSearch.Shown() · Chosen",
				fmtInt(search.Shown())+" · "+(chosenWord(*chosen)))
		})

		ui.Column(c).WidthPercent(40).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "ThemeSelector — 三个词，所以是单选而不是分段")
			account.ThemeSelector(c, account.ThemeSelectorOptions{
				Mode: &mode, Accent: &account.ThemeAccentOptions{
					Value: &accent, Label: "Accent colour",
					Presets: accentSwatches(),
				},
			})
			showcase.Field(c, "AccentColorPicker — 先给八个常用的，再给一个取色盘")
			account.AccentColorPicker(c, account.AccentColorPickerOptions{
				Value: &picker, Label: "Accent colour",
				Presets: accentSwatches(),
			})
		})
	})
}

// settingsPanel is the right-hand pane of a settings window: the one section
// that is open, in the shape that section takes.
func settingsPanel(c *ui.Context, id string) {
	k := core.Tokens(c)
	switch id {
	case "appearance":
		mode := "system"
		account.ThemeSelector(c, account.ThemeSelectorOptions{Mode: &mode})
	case "account":
		input.Switch(c, &quiet, input.SwitchOptions{Label: "Quiet hours"})
		input.Switch(c, &quiet, input.SwitchOptions{Label: "Email me a daily digest"})
	case "security":
		ui.Text(c, "Sign in with a key instead").TextColor(k.Text).
			FontSize(core.FontSize(c, theme.RowSize)).Bold()
		ui.Text(c, "Two keys registered. The last one was added on 3 March.").
			TextColor(k.TextMuted).FontSize(core.FontSize(c, theme.MetaSize))
	case "billing":
		ui.Text(c, "North branch · 6 seats").TextColor(k.Text).
			FontSize(core.FontSize(c, theme.RowSize)).Bold()
		ui.Text(c, "Next invoice on 1 April").TextColor(k.TextMuted).
			FontSize(core.FontSize(c, theme.MetaSize))
	default:
		ui.Text(c, "Pick a section").TextColor(k.TextFaint).
			FontSize(core.FontSize(c, theme.RowSize))
	}
}

// ── what am I allowed ───────────────────────────────────────────────────

// accessSection is the three screens a person reads to find out what an
// account may do: the credentials it holds, the places it is signed in, and
// how much of the plan is gone.
func accessSection(c *ui.Context) {
	showcase.Section(c, "权限与用量 · ApiKeyManager / SessionList / UsageQuota")

	revoked := showcase.State(c, "account.345.revoked", "")
	copied := showcase.State(c, "account.345.copied", "")
	keyName := showcase.State(c, "account.345.keyName", "")
	account.ApiKeyManager(c, account.ApiKeyManagerOptions{
		Revoke: revoked, Copy: copied, Name: keyName,
		OnAdd: func(name string) {},
		Keys: []account.ApiKey{
			{ID: "key-ci", Name: "CI", Secret: "cb_live_4f9a2b7c1d8e0a5b",
				Created: "12 Mar", LastUsed: "2 h ago", Scope: "Read only"},
			{ID: "key-dev", Name: "Local development",
				Secret:  "cb_test_8817f2c0d94b6e13",
				Created: "2 Feb", Scope: "Full access"},
			{ID: "key-old", Name: "Zapier", Secret: "cb_live_00c1",
				Created: "8 Nov", LastUsed: "4 months ago"},
		},
	})

	ui.Row(c).FillWidth().Gap(unit(c, 5)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).WidthPercent(52).Shrink(0).Gap(unit(c, 1.5)).Children(func() {
			showcase.Field(c, "SessionList — 当前这台机器不给「退出」按钮")
			account.SessionList(c, account.SessionListOptions{
				Revoke: &gone, SignOutAll: func() {},
				Sessions: []account.Session{
					{ID: "s-1", Device: "MacBook Pro", Where: "Lisbon, PT",
						Since: "today 09:12", Current: true},
					{ID: "s-2", Device: "iPhone 15", Where: "Porto, PT",
						Since: "2 days ago"},
					{ID: "s-3", Device: "Mac mini · office", Where: "Lisbon, PT",
						Since: "3 weeks ago", Expired: true},
				},
			})
		})
		ui.Column(c).WidthPercent(46).Shrink(0).Gap(unit(c, 1.5)).Children(func() {
			showcase.Field(c, "UsageQuota — 不到最后五分之一不说话")
			account.UsageQuota(c, account.UsageQuotaOptions{
				Title: "Usage this month", Upgrade: func() {},
				Quotas: []account.Quota{
					{ID: "q-req", Name: "Requests", Used: 1840, Limit: 2000,
						Resets: "Renews on 1 April"},
					{ID: "q-seat", Name: "Seats", Used: 6, Limit: 8},
					{ID: "q-file", Name: "File storage", Used: 0, Limit: 0},
				},
			})
		})
	})
}

// ── what keys do I press ─────────────────────────────────────────────────

// shortcutSection is the list of every binding and the one control that
// changes one.
func shortcutSection(c *ui.Context) {
	showcase.Section(c, "快捷键 · KeyboardShortcutsList / ShortcutRecorder")

	bindings := []account.Shortcut{
		{ID: "search", Name: "Go to search", Keys: []string{"cmd", "k"}, Group: "General"},
		{ID: "new", Name: "New callback", Keys: []string{"cmd", "n"}, Group: "General"},
		{ID: "save", Name: "Save the open callback", Keys: []string{"cmd", "s"}},
		{ID: "windows", Name: "Windows", Keys: []string{"ctrl", "alt", "w"}, Group: "View"},
		{ID: "board", Name: "Board view", Keys: []string{"cmd", "1"}, Group: "View"},
		{ID: "list", Name: "List view", Keys: []string{"cmd", "2"}, Group: "View"},
		{ID: "help", Name: "Keyboard shortcuts", Keys: []string{"cmd", "slash"}, Group: "Help"},
	}
	list := account.KeyboardShortcutsList(c, account.KeyboardShortcutsListOptions{
		Shortcuts: bindings, Recorded: "windows", Highlight: true,
	})
	resultLine(c, "KeyboardShortcutsList.Conflict()", fmtBool(list.Conflict()))

	showcase.Field(c, "ShortcutRecorder — 记下按下的那一个键；按钮按字宽，不拉满")
	// Stack rather than a column: a recorder is a button, and a button in a
	// FillWidth column would be a bar the width of the page.
	open := showcase.State(c, "account.recorder", []string{"cmd", "shift", "l"})
	var recorder account.ShortcutRecorderResult
	showcase.Stack(c, 1, func() {
		recorder = account.ShortcutRecorder(c, account.ShortcutRecorderOptions{
			Label: "Open the callback", Value: open, Mods: ui.Super,
		})
		// The same control empty and greyed: a binding with no keys is
		// "unbound", which is a real answer rather than a missing one.
		account.ShortcutRecorder(c, account.ShortcutRecorderOptions{
			Label: "Filter to this branch", Value: &none, Mods: ui.Super | ui.Shift,
		})
		account.ShortcutRecorder(c, account.ShortcutRecorderOptions{
			Label: "Not available in this window", Value: open, Mods: ui.Super,
			Disabled: true,
		})
	})
	resultLine(c, "ShortcutRecorder.Recorded()", quoteOrEmpty(recorder.Recorded()))
}

// ── the first run ───────────────────────────────────────────────────────

// firstRunSection is the wizard the first run shows and the bubble that
// explains one thing in an app somebody already knows.
func firstRunSection(c *ui.Context) {
	showcase.Section(c, "首次运行 · OnboardingWizard / Coachmark")

	ui.Row(c).FillWidth().Gap(unit(c, 6)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).WidthPercent(52).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "OnboardingWizard — 每一步都能跳过")
			step := showcase.State(c, "account.443.step", 0)
			finished := showcase.State(c, "account.443.finished", false)
			skipped := showcase.State(c, "account.443.skipped", false)
			wizard := account.OnboardingWizard(c, account.OnboardingWizardOptions{
				Steps: wizardSteps(c),
				Step:  step, Finished: finished, Skipped: skipped,
				ShowSkip: true,
			})
			resultLine(c, "OnboardingWizard.AtEnd() · Finished() · Skipped()",
				fmtBool(wizard.AtEnd())+" · "+fmtBool(wizard.Finished())+
					" · "+fmtBool(wizard.Skipped()))

			showcase.Field(c, "OnboardingWizard — 最后一步的按钮写的是 Finish")
			account.OnboardingWizard(c, account.OnboardingWizardOptions{
				Steps: wizardSteps(c), Step: &last, Finished: &done,
			})
		})

		ui.Column(c).WidthPercent(46).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "OnboardingWizard — 同一步，索引越界时夹到最后一步")
			// Deliberately a saved index past the end: a wizard that drew a
			// step that is not there would be a wizard with a hole in it.
			account.OnboardingWizard(c, account.OnboardingWizardOptions{
				Steps: wizardSteps(c), Step: &beyond, Finished: new(bool),
			})
		})
	})

	// A stage of its own rather than the third column of the row above: the
	// bubble is as wide as its body, and in a half-width column it would hang
	// leftwards over the wizard beside it.
	stage(c, "Coachmark", "挂在控件下面，不遮整页", func() {
		open := showcase.State(c, "account.coachmark", true)
		mark := account.Coachmark(c, account.CoachmarkOptions{
			Anchor: input.Button(c, "Open the callback", input.ButtonOptions{}),
			Open:   open, Title: "Ask anything",
			Body: "This button opens the callback you last looked at, " +
				"with the log already filtered to that machine.",
			Step: "2 of 4", ShowSkip: true,
		})
		resultLine(c, "Coachmark.Next() · Done()",
			fmtBool(mark.Next())+" · "+fmtBool(mark.Done()))
	})
}

// wizardSteps is the first run, written out. It is a function because both
// copies of the wizard on this page are the same three steps, and a wizard
// that showed a different list in its last step would be making a claim about
// the sequence that is not true of the first one.
func wizardSteps(c *ui.Context) []account.WizardStep {
	return []account.WizardStep{
		{Title: "Welcome to Callbacks",
			Body: "One window for the callbacks your machines send you: " +
				"what broke, what is waiting on you, and what got done."},
		{Title: "Point it at a machine",
			Body: "Install the agent on the machine that sends the callbacks. " +
				"It will print a key; paste it here.",
			Step: func() {
				input.TextInput(c, ptrStr("mbp-northgate.local"), input.TextInputOptions{
					Label: "Machine", Placeholder: "mbp-northgate.local",
				})
			}},
		{Title: "Choose what reaches you",
			Body: "Every callback is logged. Only the ones you choose send you a message.",
			Step: func() {
				// A row, not a column: three tags stacked down a wizard is
				// a list, and a list is not a choice.
				ui.Row(c).Gap(unit(c, 1.5)).Children(func() {
					for _, name := range []string{"Critical", "Repeat failure", "Daily digest"} {
						display.Tag(c, name, display.TagOptions{Tone: core.Accent})
					}
				})
			}},
	}
}

// ── the two that speak ───────────────────────────────────────────────────

// speakingSection is the two things this package says at the person: what a
// release changed, and a box in the corner asking what is wrong.
func speakingSection(c *ui.Context) {
	showcase.Section(c, "开口 · WhatsNewDialog / FeedbackWidget")

	ui.Row(c).FillWidth().Gap(unit(c, 6)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).WidthPercent(58).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "FeedbackWidget — 角上的一个盒子，分类 + 字数 + 发送")
			note := showcase.State(c, "account.527.note", "")
			sent := showcase.State(c, "account.527.sent", false)
			dismissed := showcase.State(c, "account.527.dismissed", false)
			widget := account.FeedbackWidget(c, account.FeedbackWidgetOptions{
				Note: note, Sent: sent, Dismissed: dismissed,
				Category: &kind, MaxRunes: 500, Closeable: true,
				Categories: []string{"Critical", "Confusing", "Slow", "Typo"},
			})
			resultLine(c, "FeedbackWidget.Sent() · Dismissed()",
				fmtBool(widget.Sent())+" · "+fmtBool(widget.Dismissed()))

			showcase.Field(c, "WhatsNewDialog — 模态，打开时是一层盖住窗口的浮层")
			ui.Text(c, "它不在这页上打开：一个遮罩会把这整页压暗，画廊就没法读了。").
				TextColor(core.Tokens(c).TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize))
		})

		ui.Column(c).WidthPercent(40).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "WhatsNewDialog — 它打开时画的是这些")
			ui.Column(c).FillWidth().Gap(unit(c, 2)).Radius(theme.ControlRadius).
				Background(core.Tokens(c).Surface).Padding(unit(c, 2.5), unit(c, 3)).
				Children(func() {
					ui.Text(c, "2.4.0").TextColor(core.Tokens(c).Text).Bold().
						FontSize(core.FontSize(c, theme.SheetSize))
					ui.Text(c, "18 March").TextColor(core.Tokens(c).TextMuted).
						FontSize(core.FontSize(c, theme.CaptionSize))
					ui.Box(c).FillWidth().Height(theme.BorderWidth).
						Background(core.Tokens(c).Border).MarginY(unit(c, 0.5))
					for _, line := range []string{
						"The log callback drawer now filters by branch, not by machine.",
						"Repeat failures get a pill instead of a red row.",
						"⌘K searches callbacks, machines and people from one field.",
					} {
						releaseLine(c, line)
					}
					ui.Text(c, "这个对话框的字段都在：版本、日期、句子和 Got it。上面那一版是"+
						"没有 NonModal 的那一个，所以画的是内容而不是浮层。").TextColor(core.Tokens(c).TextFaint).
						FontSize(core.FontSize(c, theme.CaptionSize))
				})
		})
	})
}

// ── the functions ────────────────────────────────────────────────────────

// pureSection is the part of the package with nothing a person presses: the
// rules the components above are made of, each one drawn as the call on the
// left and the value it returned on the right.
func pureSection(c *ui.Context) {
	showcase.Section(c, "纯函数 · 只调不测的那一半")
	ui.Column(c).FillWidth().Gap(unit(c, 1.5)).Children(func() {
		specimen(c, `MaskKey("cb_live_4f9a2b7c1d8e0a5b")`, account.MaskKey("cb_live_4f9a2b7c1d8e0a5b"))
		specimen(c, `MaskKey("abc")`, account.MaskKey("abc"))
		specimen(c, `MaskKey("")`, quoteOrEmpty(account.MaskKey("")))
		specimen(c, `FormatChord([]string{"command", "shift", "k"})`,
			account.FormatChord([]string{"command", "shift", "k"}))
		specimen(c, `FormatChord([]string{"ctrl", "k"})`,
			account.FormatChord([]string{"ctrl", "k"}))
		specimen(c, `FormatChord([]string{"cmd", "shift", "left"})`,
			account.FormatChord([]string{"cmd", "shift", "left"}))
		specimen(c, `NormalKey("Command")`, account.NormalKey("Command"))
		specimen(c, `NormalKey("f5")`, account.NormalKey("f5"))
		specimen(c, `NormalKeys([]string{"cmd", "", "esc"})`,
			joinKeys(account.NormalKeys([]string{"cmd", "", "esc"})))
		specimen(c, `ValidateLogin("rosa@northgate.health", "northgate-2026")`,
			joinPair(validate(account.ValidateLogin("rosa@northgate.health", "northgate-2026"))))
		specimen(c, `ValidateLogin("rosa@", "1234")`,
			joinPair(validate(account.ValidateLogin("rosa@", "1234"))))
		specimen(c, `QuotaTone(1840, 2000)`, account.QuotaTone(1840, 2000).String())
		specimen(c, `QuotaTone(1900, 2000)`, account.QuotaTone(1900, 2000).String())
		specimen(c, `QuotaTone(40, 0)`, account.QuotaTone(40, 0).String())
	})
}

// ── page furniture ──────────────────────────────────────────────────────

// stage is the frame an anchored panel is shown on: a name and a line about
// it on the left, the trigger on the right, and blank room underneath for
// the panel to hang into.
//
// The room is the whole point. A panel is drawn in the window's own overlay
// layer, so it covers whatever is under it whether the stage is tall enough
// or not — a stage is how the next component is kept out of the way, and the
// height is what keeps it out.
func stage(c *ui.Context, name, note string, build func()) {
	k := core.Tokens(c)
	ui.Box(c).FillWidth().Height(unit(c, 64)).Margin(0, 0, 0, unit(c, 2)).
		Children(func() {
			ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
				ui.Column(c).Width(unit(c, 38)).Shrink(0).Gap(unit(c, 0.5)).Children(func() {
					if name != "" {
						ui.Text(c, name).TextColor(k.Text).
							FontSize(core.FontSize(c, theme.RowSize)).Bold().SingleLine()
					}
					ui.Text(c, note).TextColor(k.TextFaint).
						FontSize(core.FontSize(c, theme.CaptionSize))
				})
				ui.Column(c).WidthPercent(60).Shrink(0).Children(build)
			})
		})
}

// specimen is one row of a page's table of pure functions: the call on the
// left, the value it returned on the right.
//
// It is defined once here and used by the account, project and devtools
// pages, because a table of values drawn slightly differently on each page
// is three tables to compare rather than one.
func specimen(c *ui.Context, call, got string) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	// The call above its own answer rather than beside it: a call with an
	// array in it is 45 characters, and a table of calls and answers on one
	// line is a table with either a truncated call or a squeezed answer.
	ui.Column(c).FillWidth().Gap(u * 0.25).Children(func() {
		ui.Text(c, call).TextColor(k.TextMuted).FillWidth().
			Font(monoFamily).FontSize(core.FontSize(c, theme.CaptionSize))
		ui.Text(c, "→ "+got).TextColor(k.Text).FillWidth().
			Font(monoFamily).FontSize(core.FontSize(c, theme.MetaSize))
	})
}

// resultLine is what a component answered, written as a caption above its
// value. Every result on this page is a function the caller reads on the
// frame after, and a gallery that draws the components without their answers
// leaves the reader to guess which ones have one.
func resultLine(c *ui.Context, label, value string) {
	k := core.Tokens(c)
	ui.Row(c).FillWidth().Gap(unit(c, 1.5)).AlignItems(ui.Center).Children(func() {
		ui.Text(c, label).TextColor(k.TextFaint).
			FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0).SingleLine()
		ui.Text(c, value).TextColor(k.TextMuted).
			FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0).SingleLine()
	})
}

// releaseLine is one sentence of what changed, with a mark in front of it —
// the same shape account's own releaseLine draws, so the two are one line
// rather than two.
func releaseLine(c *ui.Context, line string) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	ui.Row(c).FillWidth().Gap(u * 2).AlignItems(ui.Start).Children(func() {
		ui.Box(c).Width(u).MinHeight(core.FontSize(c, theme.MetaSize)).Shrink(0).
			Margin(u, u*1.6, 0, 0).Radius(u / 2).Background(k.Accent)
		ui.Text(c, line).TextColor(k.Text).Grow(1).
			FontSize(core.FontSize(c, theme.RowSize))
	})
}

// accentSwatches is the eight accents every window offers. They are named
// because a swatch shows a colour and nothing else.
func accentSwatches() []input.Swatch {
	return []input.Swatch{
		{Color: ui.Hex("#1d4ed8"), Name: "Blue"},
		{Color: ui.Hex("#0f766e"), Name: "Teal"},
		{Color: ui.Hex("#15803d"), Name: "Green"},
		{Color: ui.Hex("#b45309"), Name: "Amber"},
		{Color: ui.Hex("#be123c"), Name: "Crimson"},
		{Color: ui.Hex("#7c3aed"), Name: "Violet"},
		{Color: ui.Hex("#0f172a"), Name: "Ink"},
		{Color: ui.Hex("#71717a"), Name: "Graphite"},
	}
}

func fmtBool(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func fmtInt(n int) string { return itoaOf(n) }

func itoaOf(n int) string {
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

func ptrStr(s string) *string { return &s }

// joinKeys is a slice of names as the bracketed list a caller would have
// written it, so a table of pure functions reads the same on both sides of
// the arrow.
func joinKeys(keys []string) string {
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k)
	}
	return "[" + join(parts, " ") + "]"
}

func join(parts []string, sep string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += sep
		}
		out += p
	}
	return out
}

func quoteOrEmpty(s string) string {
	if s == "" {
		return `""`
	}
	return s
}

func chosenWord(s string) string {
	if s == "" {
		return `"" — nothing pressed this frame`
	}
	return s
}

func validate(emailErr, passErr string) []string {
	return []string{orDefault(emailErr, "no error"), orDefault(passErr, "no error")}
}

func orDefault(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

func joinPair(vals []string) string {
	return "email: " + vals[0] + " · password: " + vals[1]
}
