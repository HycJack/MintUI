package account

import (
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/input"
)

// Every component below is rendered headlessly and checked for the words it
// is supposed to show. Presses are asserted on variables captured from
// inside the view closure rather than on the text a click produces: MyGo
// builds a frame up to three times and the last pass reports nothing, so the
// drawing after a click is not yet stable. See docs/design-system.md §17.2.

// ── LoginForm ──────────────────────────────────────────────────────────────

func TestLoginFormDrawsItsFields(t *testing.T) {
	var email, password string
	var remember, revealed, submit bool
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		LoginForm(c, LoginFormOptions{
			Email: &email, Password: &password,
			Remember: &remember, Revealed: &revealed, Submit: &submit,
			Title: "Welcome back",
		})
	}, 460, 560)
	for _, want := range []string{"Welcome back", "Email", "Password", "Keep me signed in", "Sign in"} {
		if !tt.HasText(want) {
			t.Errorf("the login form is missing %q; %q", want, tt.Texts())
		}
	}
}

func TestLoginFormWalksItsFieldsInReadingOrder(t *testing.T) {
	// Typed rather than clicked, because a click inside a FormField lands on
	// whichever of its many labelled parts is under the pointer — and what
	// this test is about is the *order*, which only Tab can show. Inside a
	// FormField the label is its own line above the control, so a click on
	// the words of a label does not put the caret in the field under it.
	var email, password string
	var remember, revealed, submit bool
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		LoginForm(c, LoginFormOptions{
			Email: &email, Password: &password,
			Remember: &remember, Revealed: &revealed, Submit: &submit,
		})
	}, 460, 560)

	tt.Key(0, ui.KeyTab)
	if !tt.Focused("Email") {
		t.Fatal("the first stop is the address")
	}
	tt.Type("rosa@riverside.clinic")
	if email != "rosa@riverside.clinic" {
		t.Errorf("the field is a view of the caller's string, got %q", email)
	}

	tt.Key(0, ui.KeyTab)
	if !tt.Focused("Password") {
		t.Fatal("the second stop is the password")
	}
	tt.Type("correct-horse")
	if password != "correct-horse" {
		t.Errorf("the password is the caller's string too, got %q", password)
	}
}

func TestLoginFormReportsTheSubmit(t *testing.T) {
	var email, password string
	var remember, revealed, submit bool
	pressed := false
	// submit is never reset inside the closure. The frame is built up to
	// three times and the click lands on one of the passes, so a reset on
	// each pass would wipe the answer before anything could read it. See
	// docs/design-system.md §17.2.
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		LoginForm(c, LoginFormOptions{
			Email: &email, Password: &password,
			Remember: &remember, Revealed: &revealed, Submit: &submit,
		})
		if submit {
			pressed = true
		}
	}, 460, 560)
	if err := tt.Click("Sign in"); err != nil {
		t.Fatal(err)
	}
	if !pressed {
		t.Error("pressing Sign in should have been reported")
	}
}

func TestLoginFormShowsWhatIsWrongWithEachField(t *testing.T) {
	var email, password string
	var remember, revealed, submit bool
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		LoginForm(c, LoginFormOptions{
			Email: &email, Password: &password,
			Remember: &remember, Revealed: &revealed, Submit: &submit,
			EmailError:    "That does not look like an email address",
			PasswordError: "Passwords are at least 8 characters",
		})
	}, 460, 640)
	for _, want := range []string{
		"That does not look like an email address",
		"Passwords are at least 8 characters",
	} {
		if !tt.HasText(want) {
			t.Errorf("the form is missing the message %q; %q", want, tt.Texts())
		}
	}
}

func TestLoginFormNeedsItsPointers(t *testing.T) {
	defer panics(t, "account: LoginForm needs an Email and a Password", func() {
		var remember, submit bool
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			LoginForm(c, LoginFormOptions{Remember: &remember, Submit: &submit})
		}, 400, 400)
	})
}

// ── ApiKeyManager ──────────────────────────────────────────────────────────

func apiKeys() []ApiKey {
	return []ApiKey{{
		ID: "k1", Name: "Continuous integration", Secret: "cb_live_9f2a8b7c1d4e6f8a",
		Scope: "Read only", LastUsed: "today",
	}, {
		ID: "k2", Name: "Local development", Secret: "abcd",
		LastUsed: "",
	}}
}

func TestApiKeyManagerShowsEveryKey(t *testing.T) {
	var revoked, copied string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ApiKeyManager(c, ApiKeyManagerOptions{
			Keys: apiKeys(), Revoke: &revoked, Copy: &copied,
		})
	}, 720, 420)
	for _, want := range []string{
		"Continuous integration", "Local development",
		"Read only · today", "Full access · never used",
	} {
		if !tt.HasText(want) {
			t.Errorf("the manager is missing %q; %q", want, tt.Texts())
		}
	}
}

func TestApiKeyManagerNeverDrawsTheKey(t *testing.T) {
	var revoked, copied string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ApiKeyManager(c, ApiKeyManagerOptions{
			Keys: apiKeys(), Revoke: &revoked, Copy: &copied,
		})
	}, 720, 420)
	for _, secret := range []string{"cb_live_9f2a8b7c1d4e6f8a", "abcd"} {
		if tt.HasText(secret) {
			t.Errorf("the key %q is on the screen", secret)
		}
	}
	// The mask is: dots for all but the last four.
	if !tt.HasText("••••••••••••••••••••6f8a") {
		t.Errorf("the mask is missing; %q", tt.Texts())
	}
	// A four-character key is all dots, because revealing it would reveal it.
	if !tt.HasText("••••") {
		t.Errorf("a short key must be all dots; %q", tt.Texts())
	}
}

func TestApiKeyManagerReportsARevoke(t *testing.T) {
	var revoked, copied string
	var got []string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ApiKeyManager(c, ApiKeyManagerOptions{
			Keys: apiKeys(), Revoke: &revoked, Copy: &copied,
		})
		// Collected rather than assigned: the caller's string still holds the
		// answer on every pass after the one that pressed the button, so
		// reading it each pass records it once and does not lose it.
		if revoked != "" {
			got = append(got, revoked)
			revoked = ""
		}
	}, 720, 420)
	if err := tt.Click("Revoke"); err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 || got[0] != "k1" {
		t.Errorf("revoking the first key should report k1, got %v", got)
	}
}

func TestApiKeyManagerReportsACopy(t *testing.T) {
	var revoked, copied string
	var got []string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ApiKeyManager(c, ApiKeyManagerOptions{
			Keys: apiKeys(), Revoke: &revoked, Copy: &copied,
		})
		if copied != "" {
			got = append(got, copied)
			copied = ""
		}
	}, 720, 420)
	if err := tt.Click("Copy"); err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 || got[0] != "k1" {
		t.Errorf("copying the first key should report k1, got %v", got)
	}
}

func TestApiKeyManagerWithNoKeysSaysSo(t *testing.T) {
	var revoked, copied string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ApiKeyManager(c, ApiKeyManagerOptions{Revoke: &revoked, Copy: &copied})
	}, 480, 200)
	if !tt.HasText("No API keys yet") {
		t.Errorf("an empty manager must say it is empty; %q", tt.Texts())
	}
}

func TestApiKeyManagerMintsOne(t *testing.T) {
	var revoked, copied string
	var name string
	var added string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := ApiKeyManager(c, ApiKeyManagerOptions{
			Revoke: &revoked, Copy: &copied, Name: &name,
			OnAdd: func(given string) { added = given },
		})
		if r.Added() != "" {
			added = r.Added()
		}
	}, 720, 260)
	if !tt.HasText("Key name") {
		t.Fatalf("the mint row is missing its field; %q", tt.Texts())
	}
	if err := tt.Click("Key name"); err != nil {
		t.Fatal(err)
	}
	tt.Type("CI")
	if name != "CI" {
		t.Fatalf("the name is the caller's string, got %q", name)
	}
	if err := tt.Click("Create key"); err != nil {
		t.Fatal(err)
	}
	if added != "CI" {
		t.Errorf("minting should hand back the name typed, got %q", added)
	}
}

// ── TwoFactorInput ─────────────────────────────────────────────────────────

func TestTwoFactorInputDrawsItsBoxes(t *testing.T) {
	var code string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		TwoFactorInput(c, TwoFactorInputOptions{Value: &code, Length: 6})
	}, 420, 200)
	// Every box is named after the code and its own place in it, so somebody
	// reading one box knows which digit of six they are in.
	for _, want := range []string{"Verification code, digit 1", "Verification code, digit 6"} {
		if _, ok := tt.Find(want); !ok {
			t.Errorf("the code is missing %q; %q", want, tt.Texts())
		}
	}
}

func TestTwoFactorInputTakesTheCodeAndKnowsWhenItIsFull(t *testing.T) {
	var code string
	filled, complete := -1, false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := TwoFactorInput(c, TwoFactorInputOptions{Value: &code, Length: 6})
		filled, complete = r.Filled(), r.Complete()
	}, 420, 200)

	if err := tt.Click("Verification code, digit 1"); err != nil {
		t.Fatal(err)
	}
	// One character at a time, each with a frame to settle in: the field
	// moves the focus on after every digit, and a whole string typed in one
	// go arrives before that move has happened.
	for _, d := range "418902" {
		tt.Type(string(d))
	}
	if code != "418902" {
		t.Errorf("the code is the caller's string, got %q", code)
	}
	if filled != 6 || !complete {
		t.Errorf("six digits in six boxes is complete: filled = %d, complete = %v", filled, complete)
	}
}

func TestTwoFactorInputIsNotCompleteWhenEmpty(t *testing.T) {
	var code string
	complete := true
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		complete = TwoFactorInput(c, TwoFactorInputOptions{Value: &code}).Complete()
	}, 420, 200)
	if complete {
		t.Error("an empty code is not a complete one")
	}
}

// ── ThemeSelector ──────────────────────────────────────────────────────────

func TestThemeSelectorOffersTheThreeAppearances(t *testing.T) {
	mode := "system"
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ThemeSelector(c, ThemeSelectorOptions{Mode: &mode})
	}, 460, 340)
	for _, want := range []string{"Appearance", "Match system", "Light", "Dark"} {
		if !tt.HasText(want) {
			t.Errorf("the selector is missing %q; %q", want, tt.Texts())
		}
	}
}

func TestThemeSelectorWritesTheMode(t *testing.T) {
	mode := "system"
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ThemeSelector(c, ThemeSelectorOptions{Mode: &mode})
	}, 460, 340)
	if err := tt.Click("Dark"); err != nil {
		t.Fatal(err)
	}
	if mode != "dark" {
		t.Errorf("choosing Dark should write dark, got %q", mode)
	}
}

func TestThemeSelectorRefusesAModeItCannotShow(t *testing.T) {
	defer panics(t, "account: ThemeSelector mode neon is not one of its choices", func() {
		mode := "neon"
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			ThemeSelector(c, ThemeSelectorOptions{Mode: &mode})
		}, 400, 300)
	})
}

// ── AccentColorPicker ──────────────────────────────────────────────────────

func TestAccentColorPickerDrawsItsPreview(t *testing.T) {
	accent := ui.Hex("#2563eb")
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		AccentColorPicker(c, AccentColorPickerOptions{
			Value: &accent, Label: "Accent colour",
			Presets: []input.Swatch{
				{Color: ui.Hex("#2563eb"), Name: "Blue"},
				{Color: ui.Hex("#1f7a3d"), Name: "Green"},
			},
		})
	}, 460, 300)
	for _, want := range []string{"Accent colour", "Filled", "Outline", "Tag"} {
		if !tt.HasText(want) {
			t.Errorf("the picker is missing %q; %q", want, tt.Texts())
		}
	}
}

func TestAccentColorPickerChoosesASwatch(t *testing.T) {
	accent := ui.Hex("#2563eb")
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		AccentColorPicker(c, AccentColorPickerOptions{
			Value: &accent, Label: "Accent colour",
			Presets: []input.Swatch{
				{Color: ui.Hex("#2563eb"), Name: "Blue"},
				{Color: ui.Hex("#1f7a3d"), Name: "Green"},
			},
		})
	}, 460, 300)
	if err := tt.Click("Green"); err != nil {
		t.Fatal(err)
	}
	if accent != (ui.Hex("#1f7a3d")) {
		t.Errorf("choosing Green should write it, got %v", accent)
	}
}

func TestAccentColorPickerInsistsOnALabel(t *testing.T) {
	defer panics(t, "account: AccentColorPicker needs a Label", func() {
		accent := ui.Hex("#2563eb")
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			AccentColorPicker(c, AccentColorPickerOptions{Value: &accent})
		}, 400, 300)
	})
}

// ── ProfileEditor ──────────────────────────────────────────────────────────

func TestProfileEditorShowsThePersonAndTheFields(t *testing.T) {
	var name, email string
	var saved bool
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ProfileEditor(c, ProfileEditorOptions{
			Name: &name, Email: &email, Save: &saved,
			Role: "Technician",
		})
		// The caller's strings are the whole state, so putting a name in
		// them from inside the frame is how the row above the fields shows
		// one — which is the arrangement the editor promises.
		name, email = "Rosa Vidal", "rosa@riverside.clinic"
		saved = false
	}, 520, 460)
	for _, want := range []string{"Technician", "Display name", "Email", "Save changes"} {
		if !tt.HasText(want) {
			t.Errorf("the editor is missing %q; %q", want, tt.Texts())
		}
	}
}

func TestProfileEditorReportsASave(t *testing.T) {
	var name, email string
	var saved bool
	hits := 0 // counted, not set: a press is reported on more than one pass
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := ProfileEditor(c, ProfileEditorOptions{Name: &name, Email: &email, Save: &saved})
		if r.Saved() {
			hits++
		}
	}, 520, 460)
	if err := tt.Click("Save changes"); err != nil {
		t.Fatal(err)
	}
	if hits == 0 {
		t.Error("saving should have been reported")
	}
}

// ── AccountSwitcher ────────────────────────────────────────────────────────

func accounts() []Account {
	return []Account{
		{ID: "a1", Name: "Rosa Vidal", Email: "rosa@riverside.clinic", Kind: "Work"},
		{ID: "a2", Name: "Rosa Vidal", Email: "rosa@home.example", Kind: "Personal"},
	}
}

func TestAccountSwitcherNamesTheAccountItWillSwitchFrom(t *testing.T) {
	selected := "a1"
	open := false
	var switched string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		AccountSwitcher(c, AccountSwitcherOptions{
			Accounts: accounts(), Selected: &selected,
			Open: &open, Switched: &switched,
		})
	}, 420, 200)
	// The name alone would announce as "Rosa Vidal" and say nothing about
	// what pressing it does.
	if _, ok := tt.Find("Rosa Vidal — Switch account"); !ok {
		t.Errorf("the trigger is not named after what it does; %q", tt.Texts())
	}
}

func TestAccountSwitcherChoosesAnAccount(t *testing.T) {
	selected := "a1"
	open := false
	var switched string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		AccountSwitcher(c, AccountSwitcherOptions{
			Accounts: accounts(), Selected: &selected,
			Open: &open, Switched: &switched,
		})
	}, 420, 400)
	if err := tt.Click("Rosa Vidal — Switch account"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Personal"); err != nil {
		t.Fatal(err)
	}
	if switched != "a2" {
		t.Errorf("choosing the second account should write a2, got %q", switched)
	}
	if selected != "a2" {
		t.Errorf("the caller's own choice should follow, got %q", selected)
	}
}

// ── WorkspaceSwitcher ──────────────────────────────────────────────────────

func TestWorkspaceSwitcherRefusesALockedWorkspace(t *testing.T) {
	selected := "w1"
	open := false
	var switched string
	workspaces := []Workspace{
		{ID: "w1", Name: "Riverside", Seats: 12},
		{ID: "w2", Name: "Suspended trial", Seats: 1, Locked: true},
	}
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		WorkspaceSwitcher(c, WorkspaceSwitcherOptions{
			Workspaces: workspaces, Selected: &selected,
			Open: &open, Switched: &switched,
		})
	}, 420, 400)
	if err := tt.Click("Riverside — Switch workspace"); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Riverside", "12 seats", "Locked"} {
		if !tt.HasText(want) {
			t.Errorf("the switcher is missing %q; %q", want, tt.Texts())
		}
	}
	if err := tt.Click("Suspended trial"); err != nil {
		t.Fatal(err)
	}
	// Disabled only greys a row; it does not refuse the press for us. See
	// docs/design-system.md §17.3 — the guard has to be in our own code,
	// and this is the assertion that it is.
	if switched != "" {
		t.Errorf("a locked workspace must not be chosen, got %q", switched)
	}
	if selected != "w1" {
		t.Errorf("the choice must not move, got %q", selected)
	}
}

// ── UsageQuota ─────────────────────────────────────────────────────────────

func TestUsageQuotaDrawsEveryRow(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		UsageQuota(c, UsageQuotaOptions{
			Title: "March",
			Quotas: []Quota{
				{ID: "calls", Name: "API calls", Used: 7420, Limit: 10000, Resets: "Resets 1 April"},
				{ID: "seats", Name: "Technicians", Used: 12, Limit: 0},
				{ID: "storage", Name: "Storage", Used: 10000, Limit: 10000},
			},
		})
	}, 520, 460)
	for _, want := range []string{
		"March", "API calls", "7,420 / 10,000", "Resets 1 April",
		"Technicians", "Not metered", "Storage", "10,000 / 10,000",
	} {
		if !tt.HasText(want) {
			t.Errorf("the panel is missing %q; %q", want, tt.Texts())
		}
	}
}

func TestUsageQuotaOffersTheUpgrade(t *testing.T) {
	upgraded := false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		UsageQuota(c, UsageQuotaOptions{
			Quotas:  []Quota{{ID: "calls", Name: "API calls", Used: 9999, Limit: 10000}},
			Upgrade: func() { upgraded = true },
		})
	}, 520, 300)
	if err := tt.Click("Upgrade plan"); err != nil {
		t.Fatal(err)
	}
	if !upgraded {
		t.Error("the upgrade button did not fire")
	}
}

// ── OnboardingWizard ───────────────────────────────────────────────────────

func steps() []WizardStep {
	return []WizardStep{
		{Title: "Connect your board", Body: "Point us at the board you already use."},
		{Title: "Invite the team", Body: "Nobody sees anything until you invite them."},
		{Title: "You're set", Body: "Callbacks start arriving straight away."},
	}
}

func TestOnboardingWizardShowsTheFirstStep(t *testing.T) {
	at := 0
	var finished, skipped bool
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := OnboardingWizard(c, OnboardingWizardOptions{
			Steps: steps(), Step: &at, Finished: &finished, Skipped: &skipped,
			ShowSkip: true,
		})
		if r.AtEnd() {
			t.Error("the first of three steps is not the end")
		}
	}, 520, 460)
	for _, want := range []string{"1", " / 3", "Connect your board", "Next", "Skip setup"} {
		if !tt.HasText(want) {
			t.Errorf("the wizard is missing %q; %q", want, tt.Texts())
		}
	}
	if tt.HasText("You're set") {
		t.Error("the third step's title is drawn on the first step")
	}
}

func TestOnboardingWizardWalksAndFinishes(t *testing.T) {
	at := 0
	var finished, skipped bool
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		OnboardingWizard(c, OnboardingWizardOptions{
			Steps: steps(), Step: &at, Finished: &finished, Skipped: &skipped,
		})
	}, 520, 460)
	if err := tt.Click("Next"); err != nil {
		t.Fatal(err)
	}
	if at != 1 {
		t.Fatalf("Next should move to step 2, got %d", at)
	}
	if err := tt.Click("Next"); err != nil {
		t.Fatal(err)
	}
	if at != 2 {
		t.Fatalf("Next should move to step 3, got %d", at)
	}
	if err := tt.Click("Finish"); err != nil {
		t.Fatal(err)
	}
	if !finished {
		t.Error("the last step should finish the wizard")
	}
	if tt.HasText("Back") != true {
		t.Error("the last step should offer Back")
	}
}

func TestOnboardingWizardGoesBack(t *testing.T) {
	at := 2
	var finished, skipped bool
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		OnboardingWizard(c, OnboardingWizardOptions{
			Steps: steps(), Step: &at, Finished: &finished, Skipped: &skipped,
		})
	}, 520, 460)
	if tt.HasText("Back") == false {
		t.Fatal("the third step should offer Back")
	}
	if err := tt.Click("Back"); err != nil {
		t.Fatal(err)
	}
	if at != 1 {
		t.Errorf("Back should move to step 2, got %d", at)
	}
}

func TestOnboardingWizardSkips(t *testing.T) {
	at := 0
	var finished, skipped bool
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		OnboardingWizard(c, OnboardingWizardOptions{
			Steps: steps(), Step: &at, Finished: &finished, Skipped: &skipped,
			ShowSkip: true,
		})
	}, 520, 460)
	if err := tt.Click("Skip setup"); err != nil {
		t.Fatal(err)
	}
	if !skipped {
		t.Error("skipping should be remembered, or it is shown again next run")
	}
	if finished {
		t.Error("skipping is not finishing")
	}
}

func TestOnboardingWizardClampsAnIndexFromALongerWizard(t *testing.T) {
	at := 9
	var finished, skipped bool
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		OnboardingWizard(c, OnboardingWizardOptions{
			Steps: steps(), Step: &at, Finished: &finished, Skipped: &skipped,
		})
	}, 520, 460)
	if at != 2 {
		t.Errorf("an index past the end should land on the last step, got %d", at)
	}
	if !tt.HasText("You're set") {
		t.Errorf("the last step is not drawn; %q", tt.Texts())
	}
}

// ── Coachmark ──────────────────────────────────────────────────────────────

func TestCoachmarkPointsAtItsAnchor(t *testing.T) {
	open := true
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		anchor := input.Button(c, "Log a callback", input.ButtonOptions{})
		Coachmark(c, CoachmarkOptions{
			Anchor: anchor, Open: &open,
			Title: "Log a callback", Body: "It takes four fields and nobody else sees it.",
		})
	}, 520, 420)
	for _, want := range []string{"Log a callback", "It takes four fields", "Got it"} {
		if !tt.HasText(want) {
			t.Errorf("the mark is missing %q; %q", want, tt.Texts())
		}
	}
}

func TestCoachmarkClosesItself(t *testing.T) {
	open := true
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		anchor := input.Button(c, "Log a callback", input.ButtonOptions{})
		r := Coachmark(c, CoachmarkOptions{
			Anchor: anchor, Open: &open, Title: "Log a callback",
		})
		if r.Done() {
			open = false
		}
	}, 520, 420)
	if err := tt.Click("Got it"); err != nil {
		t.Fatal(err)
	}
	if open {
		t.Error("the mark should have closed itself")
	}
}

func TestCoachmarkNeedsAnAnchor(t *testing.T) {
	defer panics(t, "account: Coachmark needs the Anchor it points at", func() {
		open := true
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			Coachmark(c, CoachmarkOptions{Open: &open, Title: "Hello"})
		}, 400, 300)
	})
}

// ── FeedbackWidget ─────────────────────────────────────────────────────────

func TestFeedbackWidgetDrawsItsParts(t *testing.T) {
	var note, category string
	var sent, dismissed bool
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := FeedbackWidget(c, FeedbackWidgetOptions{
			Note: &note, Sent: &sent, Dismissed: &dismissed, Closeable: true,
			Categories: []string{"Bug", "Idea"}, Category: &category, MaxRunes: 500,
		})
		_, _, _ = r.Sent(), r.Dismissed(), category
	}, 420, 460)
	for _, want := range []string{
		"Feedback", "Bug", "Idea",
		"What is going wrong?", "0 / 500", "Send",
	} {
		if !tt.HasText(want) {
			t.Errorf("the widget is missing %q; %q", want, tt.Texts())
		}
	}
}

func TestFeedbackWidgetSendsANote(t *testing.T) {
	var note, category string
	var sent bool
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		FeedbackWidget(c, FeedbackWidgetOptions{
			Note: &note, Sent: &sent, Categories: []string{"Bug"}, Category: &category,
		})
	}, 420, 460)
	if err := tt.Click("What is going wrong?"); err != nil {
		t.Fatal(err)
	}
	tt.Type("the board scrolled itself")
	if err := tt.Click("Send"); err != nil {
		t.Fatal(err)
	}
	if !sent {
		t.Error("sending should be reported to the caller")
	}
}

func TestFeedbackWidgetCountsRunesNotBytes(t *testing.T) {
	var note string
	var sent bool
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		FeedbackWidget(c, FeedbackWidgetOptions{
			Note: &note, Sent: &sent, MaxRunes: 500,
		})
	}, 420, 460)
	note = "🎯🎯🎯"
	if err := tt.Click("Send"); err != nil {
		t.Fatal(err)
	}
	if !tt.HasText("3 / 500") {
		t.Errorf("three emoji are three characters; %q", tt.Texts())
	}
}

// ── KeyboardShortcutsList ──────────────────────────────────────────────────

func shortcuts() []Shortcut {
	return []Shortcut{
		{ID: "open", Name: "Open the board", Keys: []string{"cmd", "o"}, Group: "General"},
		{ID: "search", Name: "Search", Keys: []string{"cmd", "k"}, Group: "General"},
		{ID: "log", Name: "Log a callback", Keys: []string{"cmd", "shift", "l"}, Group: "Callbacks"},
	}
}

func TestKeyboardShortcutsListGroupsAndShowsChords(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		KeyboardShortcutsList(c, KeyboardShortcutsListOptions{Shortcuts: shortcuts()})
	}, 520, 400)
	for _, want := range []string{
		"Keyboard shortcuts", "General", "Callbacks",
		"Open the board", "Search", "Log a callback",
		// Kbd puts each key in its own box, so the marks are there one by one.
		"⌘", "⇧", "L",
	} {
		if !tt.HasText(want) {
			t.Errorf("the list is missing %q; %q", want, tt.Texts())
		}
	}
}

func TestKeyboardShortcutsListSaysWhenTwoBindingsClash(t *testing.T) {
	clash := false
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := KeyboardShortcutsList(c, KeyboardShortcutsListOptions{Shortcuts: shortcuts()})
		if r.Conflict() {
			clash = true
		}
	}, 520, 400)
	if clash {
		t.Error("three distinct chords are not a conflict")
	}

	clash = false
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := KeyboardShortcutsList(c, KeyboardShortcutsListOptions{Shortcuts: []Shortcut{
			{ID: "a", Name: "One", Keys: []string{"cmd", "k"}},
			{ID: "b", Name: "Two", Keys: []string{"cmd", "K"}},
		}})
		if r.Conflict() {
			clash = true
		}
	}, 520, 300)
	if !clash {
		t.Error("the same chord twice must be reported, whatever case it was written in")
	}
}

func TestKeyboardShortcutsListWithNothingBound(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		KeyboardShortcutsList(c, KeyboardShortcutsListOptions{})
	}, 520, 200)
	if !tt.HasText("No shortcuts yet") {
		t.Errorf("an empty list must say so; %q", tt.Texts())
	}
}

// ── ShortcutRecorder ───────────────────────────────────────────────────────

func TestShortcutRecorderRecordsAChord(t *testing.T) {
	var chord []string
	var recorded string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := ShortcutRecorder(c, ShortcutRecorderOptions{
			Label: "Log a callback", Value: &chord, Mods: ui.Cmd,
		})
		if r.Recorded() != "" {
			recorded = r.Recorded()
		}
	}, 520, 260)
	if !tt.HasText("Press keys") {
		t.Fatalf("an empty recorder should say what to do; %q", tt.Texts())
	}
	if err := tt.Click("Log a callback"); err != nil {
		t.Fatal(err)
	}
	tt.Key(ui.Cmd, ui.KeyL)
	if recorded != "L" {
		t.Fatalf("the recorder should have caught L, got %q", recorded)
	}
	// ui.Cmd is Command on macOS and Control everywhere else, and the
	// recorder records the modifier it was given rather than assuming one.
	want := append(modifierNames(ui.Cmd), "L")
	if len(chord) != len(want) {
		t.Fatalf("the recorded chord = %v, want %v", chord, want)
	}
	for i := range want {
		if chord[i] != want[i] {
			t.Fatalf("the recorded chord = %v, want %v", chord, want)
		}
	}
}

func TestShortcutRecorderClearsOnAPress(t *testing.T) {
	chord := []string{"cmd", "K"}
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ShortcutRecorder(c, ShortcutRecorderOptions{
			Label: "Search", Value: &chord, Mods: ui.Cmd,
		})
	}, 520, 260)
	if !tt.HasText("⌘") {
		t.Fatalf("the bound chord should be drawn as its marks; %q", tt.Texts())
	}
	if err := tt.Click("Search"); err != nil {
		t.Fatal(err)
	}
	if len(chord) != 0 {
		t.Errorf("pressing a bound recorder should unbind it, got %v", chord)
	}
}

// ── SettingsLayout ─────────────────────────────────────────────────────────

func settingsFixtureLayout() []Setting {
	return []Setting{
		{ID: "appearance", Name: "Appearance", Hint: "Light, dark or your system"},
		{ID: "keys", Name: "API keys", Hint: "Credentials for the API"},
		{ID: "sync", Name: "Sync", Hint: "Refresh when the app is opened"},
	}
}

func TestSettingsLayoutShowsTheSectionsAndTheOpenOne(t *testing.T) {
	selected := "appearance"
	var chosen, query string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		SettingsLayout(c, SettingsLayoutOptions{
			Sections: settingsFixtureLayout(), Selected: &selected,
			Chosen: &chosen, Query: &query, Searchable: true,
			Section: func(id string) {
				ui.Text(c, "Panel for "+id)
			},
		})
	}, 980, 520)
	for _, want := range []string{
		"Appearance", "API keys", "Sync", "Search settings", "Panel for appearance",
	} {
		if !tt.HasText(want) {
			t.Errorf("the layout is missing %q; %q", want, tt.Texts())
		}
	}
}

func TestSettingsLayoutChoosesASection(t *testing.T) {
	selected := "appearance"
	var chosen, query string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		SettingsLayout(c, SettingsLayoutOptions{
			Sections: settingsFixtureLayout(), Selected: &selected,
			Chosen: &chosen, Query: &query,
			Section: func(id string) { ui.Text(c, "Panel for "+id) },
		})
	}, 980, 520)
	if err := tt.Click("API keys"); err != nil {
		t.Fatal(err)
	}
	if chosen != "keys" {
		t.Errorf("choosing a section should write its id, got %q", chosen)
	}
	if selected != "keys" {
		t.Errorf("the caller's own choice should follow, got %q", selected)
	}
}

func TestSettingsLayoutSaysWhenTheSearchMatchesNothing(t *testing.T) {
	selected := "appearance"
	var chosen, query string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		SettingsLayout(c, SettingsLayoutOptions{
			Sections: settingsFixtureLayout(), Selected: &selected,
			Chosen: &chosen, Query: &query, Searchable: true,
		})
	}, 980, 520)
	if err := tt.Click("Search settings"); err != nil {
		t.Fatal(err)
	}
	tt.Type("zzzz")
	if !tt.HasText("Nothing matches") {
		t.Errorf("a search over nothing must say so; %q", tt.Texts())
	}
}

// ── SettingsSearch ─────────────────────────────────────────────────────────

func TestSettingsSearchRanksAndReportsAChoice(t *testing.T) {
	var query, chosen string
	shown := 0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := SettingsSearch(c, SettingsSearchOptions{
			Settings: settingsFixtureLayout(), Query: &query, Chosen: &chosen,
			Limit: 2,
		})
		shown = r.Shown()
	}, 620, 460)

	if err := tt.Click("Search settings"); err != nil {
		t.Fatal(err)
	}
	for _, r := range "key" {
		tt.Type(string(r))
	}
	if shown != 1 {
		t.Errorf("one setting mentions a key, got %d", shown)
	}
	if err := tt.Click("API keys"); err != nil {
		t.Fatal(err)
	}
	if chosen != "keys" {
		t.Errorf("choosing a row should write its id, got %q", chosen)
	}
}

func TestSettingsSearchCapsTheList(t *testing.T) {
	var query, chosen string
	shown := 0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := SettingsSearch(c, SettingsSearchOptions{
			Settings: settingsFixtureLayout(), Query: &query, Chosen: &chosen, Limit: 1,
		})
		shown = r.Shown()
	}, 620, 460)
	if err := tt.Click("Search settings"); err != nil {
		t.Fatal(err)
	}
	tt.Type("a")
	if shown != 3 {
		t.Errorf("Shown is how many matched, not how many were drawn; got %d", shown)
	}
	if tt.HasText("More below") == false {
		t.Errorf("the cap should be said; %q", tt.Texts())
	}
}

// ── WhatsNewDialog ─────────────────────────────────────────────────────────

func releases() []Release {
	return []Release{
		{Version: "4.2", Date: "14 March", Current: true, What: []string{
			"The board scrolls sideways instead of squeezing its columns.",
			"Shortcuts can be rebound.",
		}},
		{Version: "4.1", Date: "2 March", What: []string{"Sessions can be revoked."}},
	}
}

func TestWhatsNewDialogIsNothingWhileShut(t *testing.T) {
	open := false
	selected := 0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		WhatsNewDialog(c, WhatsNewOptions{
			Open: &open, Releases: releases(), Selected: &selected,
		})
	}, 900, 620)
	if tt.HasText("Shortcuts can be rebound.") {
		t.Errorf("a closed dialog draws nothing; %q", tt.Texts())
	}
}

func TestWhatsNewDialogShowsTheCurrentRelease(t *testing.T) {
	open := true
	selected := -1
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		WhatsNewDialog(c, WhatsNewOptions{
			Open: &open, Releases: releases(), Selected: &selected,
		})
	}, 900, 620)
	for _, want := range []string{
		"4.2", "4.1", "New", "14 March",
		"The board scrolls sideways instead of squeezing its columns.",
		"Shortcuts can be rebound.", "Got it",
	} {
		if !tt.HasText(want) {
			t.Errorf("the dialog is missing %q; %q", want, tt.Texts())
		}
	}
	// It opened on the release the window is running, not on index 0 by
	// accident: this fixture puts the current one first and the old one
	// second, so the assertion that 4.2's sentences are up is the one that
	// says which.
	if selected != 0 {
		t.Errorf("selected = %d, want the current release", selected)
	}
}

func TestWhatsNewDialogPagesToAnOlderRelease(t *testing.T) {
	open := true
	selected := 0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		WhatsNewDialog(c, WhatsNewOptions{
			Open: &open, Releases: releases(), Selected: &selected,
		})
	}, 900, 620)
	if err := tt.Click("4.1"); err != nil {
		t.Fatal(err)
	}
	if selected != 1 {
		t.Errorf("choosing an older release should move the index, got %d", selected)
	}
}

func TestWhatsNewDialogClosesOnGotIt(t *testing.T) {
	open := true
	selected := 0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		WhatsNewDialog(c, WhatsNewOptions{
			Open: &open, Releases: releases(), Selected: &selected,
		})
	}, 900, 620)
	if err := tt.Click("Got it"); err != nil {
		t.Fatal(err)
	}
	if open {
		t.Error("Got it should close the dialog")
	}
}

func TestWhatsNewDialogNeedsARelease(t *testing.T) {
	defer panics(t, "account: WhatsNewDialog needs at least one release", func() {
		open := true
		selected := 0
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			WhatsNewDialog(c, WhatsNewOptions{Open: &open, Selected: &selected})
		}, 900, 620)
	})
}

// ── SessionList ────────────────────────────────────────────────────────────

func sessions() []Session {
	return []Session{
		{ID: "s1", Device: "MacBook Pro", Where: "Riverside Clinic", Since: "today", Current: true},
		{ID: "s2", Device: "iPhone 15", Where: "Austin, TX", Since: "3 days ago"},
		{ID: "s3", Device: "Old iPad", Where: "Austin, TX", Since: "last March", Expired: true},
	}
}

func TestSessionListShowsEveryDevice(t *testing.T) {
	var revoked string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		SessionList(c, SessionListOptions{Sessions: sessions(), Revoke: &revoked})
	}, 620, 420)
	for _, want := range []string{
		"MacBook Pro", "This device",
		"iPhone 15", "Austin, TX · 3 days ago",
		"Old iPad", "Expired",
	} {
		if !tt.HasText(want) {
			t.Errorf("the list is missing %q; %q", want, tt.Texts())
		}
	}
}

func TestSessionListRevokesADeviceButNotThisOne(t *testing.T) {
	var revoked string
	var got []string
	allOut := false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		SessionList(c, SessionListOptions{
			Sessions: sessions(), Revoke: &revoked,
			SignOutAll: func() { allOut = true },
		})
		if revoked != "" {
			got = append(got, revoked)
			revoked = ""
		}
	}, 620, 460)
	if err := tt.Click("Sign out"); err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 || got[0] != "s2" {
		t.Errorf("the only other live device should be revoked, got %v", got)
	}
	if err := tt.Click("Sign out everywhere"); err != nil {
		t.Fatal(err)
	}
	if !allOut {
		t.Error("sign out everywhere did not fire")
	}
}

func TestSessionListWithNoDevices(t *testing.T) {
	var revoked string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		SessionList(c, SessionListOptions{Revoke: &revoked})
	}, 520, 200)
	if !tt.HasText("No other devices") {
		t.Errorf("an empty list must say so; %q", tt.Texts())
	}
}

// ── UserMenu ───────────────────────────────────────────────────────────────

func TestUserMenuOpensFromTheAvatarAndTakesAChoice(t *testing.T) {
	open := false
	var chosen string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		UserMenu(c, UserMenuOptions{
			Profile: Profile{ID: "p1", Name: "Rosa Vidal", Email: "rosa@riverside.clinic"},
			Items: []UserMenuItem{
				{ID: "profile", Label: "Profile"},
				{ID: "keys", Label: "API keys"},
			},
			Open: &open, Chosen: &chosen,
			SignOut: func() {},
		})
	}, 480, 440)
	if err := tt.Click("Rosa Vidal"); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"rosa@riverside.clinic", "API keys", "Sign out"} {
		if !tt.HasText(want) {
			t.Errorf("the menu is missing %q; %q", want, tt.Texts())
		}
	}
	if err := tt.Click("API keys"); err != nil {
		t.Fatal(err)
	}
	if chosen != "keys" {
		t.Errorf("choosing an entry should write its id, got %q", chosen)
	}
}

func TestUserMenuSignOut(t *testing.T) {
	open := false
	var chosen string
	out := false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		UserMenu(c, UserMenuOptions{
			Profile: Profile{Name: "Rosa Vidal"},
			Open:    &open, Chosen: &chosen,
			SignOut: func() { out = true },
		})
	}, 480, 440)
	if err := tt.Click("Rosa Vidal"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Sign out"); err != nil {
		t.Fatal(err)
	}
	if !out {
		t.Error("sign out did not fire")
	}
}

// ── the dark palette ───────────────────────────────────────────────────────
// One pass over the whole package, because a component that reads a raw
// colour anywhere shows up here and nowhere else.

func TestAccountSurvivesTheDarkPalette(t *testing.T) {
	// Every sink the page writes to is declared here and read at the end.
	// A dark pass that collects answers and does not look at them cannot
	// catch a control firing on its own, which is the failure a dark palette
	// exists to surface.
	var (
		email, password, note, category, query, chosen string
		added, revoked, copied, switched, recorded     string
		mode, selected                                 = "system", "a1"
		at, release                                    = 0, 0
		remember, revealed, submit, sent, dismissed    bool
		open, finished, skipped, signedOut, upgraded   bool
		chord                                          []string
	)
	accent := ui.Hex("#5b8dff")

	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: core.Dark})
		LoginForm(c, LoginFormOptions{
			Email: &email, Password: &password, Remember: &remember,
			Revealed: &revealed, Submit: &submit,
		})
		ApiKeyManager(c, ApiKeyManagerOptions{
			Keys: apiKeys(), Revoke: &revoked, Copy: &copied,
		})
		TwoFactorInput(c, TwoFactorInputOptions{Value: &password})
		ThemeSelector(c, ThemeSelectorOptions{Mode: &mode})
		AccentColorPicker(c, AccentColorPickerOptions{Value: &accent, Label: "Accent colour"})
		ProfileEditor(c, ProfileEditorOptions{
			Name: &email, Email: &email, Save: &submit, Role: "Technician",
		})
		AccountSwitcher(c, AccountSwitcherOptions{
			Accounts: accounts(), Selected: &selected, Open: &open, Switched: &switched,
		})
		WorkspaceSwitcher(c, WorkspaceSwitcherOptions{
			Workspaces: []Workspace{{ID: "w1", Name: "Riverside", Seats: 12}},
			Selected:   &selected, Open: &open, Switched: &switched,
		})
		UsageQuota(c, UsageQuotaOptions{
			Title: "March",
			Quotas: []Quota{
				{ID: "calls", Name: "API calls", Used: 9999, Limit: 10000, Resets: "Resets 1 April"},
			},
			Upgrade: func() { upgraded = true },
		})
		OnboardingWizard(c, OnboardingWizardOptions{
			Steps: steps(), Step: &at, Finished: &finished, Skipped: &skipped, ShowSkip: true,
		})
		FeedbackWidget(c, FeedbackWidgetOptions{
			Note: &note, Sent: &sent, Dismissed: &dismissed, Closeable: true,
			Categories: []string{"Bug"}, Category: &category, MaxRunes: 500,
		})
		KeyboardShortcutsList(c, KeyboardShortcutsListOptions{Shortcuts: shortcuts()})
		ShortcutRecorder(c, ShortcutRecorderOptions{
			Label: "Search", Value: &chord, Mods: ui.Cmd,
		})
		SettingsLayout(c, SettingsLayoutOptions{
			Sections: settingsFixtureLayout(), Selected: &selected,
			Chosen: &chosen, Query: &query, Searchable: true,
			Section: func(id string) { ui.Text(c, "Panel for "+id) },
		})
		SettingsSearch(c, SettingsSearchOptions{
			Settings: settingsFixtureLayout(), Query: &query, Chosen: &chosen,
		})
		WhatsNewDialog(c, WhatsNewOptions{
			Open: &open, Releases: releases(), Selected: &release,
		})
		SessionList(c, SessionListOptions{Sessions: sessions(), Revoke: &revoked})
		UserMenu(c, UserMenuOptions{
			Profile: Profile{Name: "Rosa Vidal", Email: "rosa@riverside.clinic"},
			Items:   []UserMenuItem{{ID: "keys", Label: "API keys"}},
			Open:    &open, Chosen: &chosen,
			SignOut: func() { signedOut = true },
		})
	}, 1200, 1400)

	for _, want := range []string{
		"Sign in", "API keys", "Appearance", "Accent colour",
		"API calls", "9,999 / 10,000", "Feedback", "Keyboard shortcuts",
		"Appearance", "MacBook Pro", "Settings results",
	} {
		if !tt.HasText(want) {
			t.Errorf("the dark pass is missing %q; %q", want, tt.Texts())
		}
	}
	// The chords come out as their marks here too: the check that the list is
	// not quietly reading the palette's light values.
	if !tt.HasText("⌘") {
		t.Errorf("the chord marks are missing in the dark; %q", tt.Texts())
	}
	// The key must still be a key in the dark: a mask that leaked would be a
	// leak on a screen somebody is looking at in a dim room.
	if tt.HasText("cb_live_9f2a8b7c1d4e6f8a") {
		t.Error("the secret is on the screen in the dark palette")
	}
	// Nineteen components on one frame is nineteen chances to write to a
	// caller's pointer. Nothing here was pressed, so nothing may have moved.
	if signedOut || upgraded || dismissed || sent || finished || skipped ||
		open || submit {
		t.Errorf("something fired without being pressed: signedOut=%v upgraded=%v "+
			"dismissed=%v sent=%v finished=%v skipped=%v open=%v submit=%v",
			signedOut, upgraded, dismissed, sent, finished, skipped, open, submit)
	}
	for _, got := range []struct{ what, v string }{
		{"added", added}, {"revoked", revoked}, {"copied", copied},
		{"switched", switched}, {"chosen", chosen}, {"query", query},
		{"recorded", recorded}, {"note", note}, {"category", category},
	} {
		if got.v != "" {
			t.Errorf("%s was written to without being pressed: %q", got.what, got.v)
		}
	}
	if at != 0 || release != 0 {
		t.Errorf("the wizard or the release moved on its own: step=%d release=%d", at, release)
	}
	if mode != "system" || selected != "a1" {
		t.Errorf("a setting moved on its own: mode=%q selected=%q", mode, selected)
	}
	if len(chord) != 0 {
		t.Errorf("the recorder bound something on its own: %v", chord)
	}
	_ = remember
	_ = revealed
	_ = email
	_ = password
	_ = accent
}

// ── a helper ───────────────────────────────────────────────────────────────

// panics asserts that building the thing panics with a message starting with
// want. The library's rule is that anything a component cannot be given is a
// panic rather than a guess, so these are as much a part of the contract as
// what it draws.
func panics(t *testing.T, want string, fn func()) {
	t.Helper()
	{
		defer func() {
			r := recover()
			if r == nil {
				t.Fatalf("expected a panic saying %q", want)
			}
			got, ok := r.(string)
			if !ok {
				t.Fatalf("panic value is %T, not a string: %v", r, r)
			}
			if !strings.HasPrefix(got, want) {
				t.Errorf("panic = %q, want it to start with %q", got, want)
			}
		}()
		fn()
	}
}
