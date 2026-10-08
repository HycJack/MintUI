package account_test

import (
	"fmt"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/account"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/input"
)

// Example shows the whole of an account screen: a person, the key that lets a
// server act as them, and the machine that session is running on.
//
// Three things are worth reading off this page. The fields hold the caller's
// strings — there is no copy of the address anywhere in this package. The key
// is drawn through MaskKey and nothing else, so a screenshot of this window
// is safe to put in a bug report. And the sessions list shows the current
// device with no way to sign it out, because signing out of the window you
// are reading is a different action with its own name.
func Example() {
	// Seeded so the example draws a filled-in account rather than five empty
	// fields, which is what every screenshot in a docs page looks like.
	email, password := "Rosa Vidal", "rosa@riverside.clinic"
	var code string
	var submitted bool

	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		k, u := core.Tokens(c), core.Density(c).Unit()

		account.ProfileEditor(c, account.ProfileEditorOptions{
			Name: &email, Email: &password, Save: &submitted,
			Role: "Technician",
		})

		account.ApiKeyManager(c, account.ApiKeyManagerOptions{
			Keys: []account.ApiKey{{
				ID: "k1", Name: "Continuous integration",
				Secret: "cb_live_9f2a8b7c1d4e6f8a",
				Scope:  "Read only", LastUsed: "today",
			}},
			Revoke: new(string),
			Copy:   new(string),
		})

		account.TwoFactorInput(c, account.TwoFactorInputOptions{
			Value: &code, Length: 6, Masked: true,
		})

		account.SessionList(c, account.SessionListOptions{
			Revoke: new(string),
			Sessions: []account.Session{
				{ID: "s1", Device: "MacBook Pro", Where: "Riverside Clinic", Since: "today", Current: true},
				{ID: "s2", Device: "iPhone 15", Where: "Austin, TX", Since: "3 days ago"},
			},
		})

		// A colour well shows a colour and no words, so this one is named —
		// and the preview under it is drawn against the palette the window
		// will actually be accented with.
		accent := k.Accent
		account.AccentColorPicker(c, account.AccentColorPickerOptions{
			Value: &accent, Label: "Accent colour",
			Presets: []input.Swatch{
				{Color: k.Accent, Name: "Blue"},
				{Color: k.Success, Name: "Green"},
			},
		})

		_ = u
	}, 760, 1200)

	// Output:
}

// ExampleLoginForm is the same account before you are in it.
//
// The errors are ValidateLogin's, passed in rather than computed inside: the
// hairline, the sentence under the field and what a screen reader says are
// then all reading one value written once, so they cannot disagree.
func ExampleLoginForm() {
	var email, password string
	var remember, revealed, submitted bool

	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		email, password = "rosa@riverside.clinic", "correct-horse"
		remember = true

		account.LoginForm(c, account.LoginFormOptions{
			Email: &email, Password: &password,
			Remember: &remember, Revealed: &revealed, Submit: &submitted,
			Title:  "Welcome back",
			Forgot: func() {},
		})
	}, 440, 620)

	// Output:
}

// Example_maskKey shows the one function a secret passes through, which is
// what makes it worth having a test of its own: it is the whole of what a
// person is allowed to read off the screen, and its two awkward edges are the
// empty key and the four-character one.
func ExampleMaskKey() {
	for _, k := range []string{
		"",
		"abc",
		"abcd",
		"cb_live_9f2a8b7c1d4e6f8a",
	} {
		fmt.Println(account.MaskKey(k))
	}

	// Output:
	//
	// •••
	// ••••
	// ••••••••••••••••••••6f8a
}

// Example_FormatChord shows a binding written two ways: the marks a settings
// screen draws them with, and the one line a table cell puts them in. Both go
// through the same lookup, so they cannot come out disagreeing.
func ExampleFormatChord() {
	fmt.Println(account.FormatChord([]string{"cmd", "shift", "k"}))
	fmt.Println(account.FormatChord([]string{"ctrl", "k"}))
	fmt.Println(account.FormatChord(nil))
	for _, key := range account.ChordKeys([]string{"cmd", "enter"}) {
		fmt.Println(key)
	}

	// Output:
	// ⌘⇧K
	// Ctrl+K
	//
	// ⌘
	// Return
}
