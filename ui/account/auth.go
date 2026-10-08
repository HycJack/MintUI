package account

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/theme"
)

// defaultCodeLength is how many digits a one-time code has when the caller
// does not say. It is the same six PinInput takes, and it is written here as
// well so that Complete can be answered without PinInputResult having to grow
// a field nobody else needs.
const defaultCodeLength = 6

// LoginFormOptions configure a LoginForm.
type LoginFormOptions struct {
	// Email and Password are the caller's strings, written as they are
	// typed. The form draws the errors the caller has already computed
	// rather than computing its own, so that the hairline, the message and
	// what a screen reader says are all one value written once.
	Email, Password *string
	// Remember is the caller's bool: whether this device may keep the
	// session. It is a switch rather than a checkbox because the thing being
	// promised is a duration, not a choice to review.
	Remember *bool
	// Revealed is the caller's bool for the eye on the password field. It is
	// the caller's because whether a password is visible is a fact about the
	// screen as a whole, and a toolbar button elsewhere may want the same
	// answer as the field's own.
	Revealed *bool
	// Submit writes true into the caller's bool when the button is pressed.
	// A bit rather than a callback, because the library has no callbacks:
	// the answer to "was Sign in pressed" is a bool the caller reads in its
	// own frame.
	Submit *bool
	// Forgot is the reset path's press. Nil draws no link, which is right
	// only for a screen where a forgotten password cannot happen.
	Forgot func()
	// EmailError and PasswordError are ValidateLogin's two messages. Empty
	// for each is what leaves that field alone.
	EmailError, PasswordError string
	// Title and Subtitle head the panel. Both empty draws no heading, which
	// is what a form inside a sheet that already has a title wants.
	Title, Subtitle string
	// Busy greys the fields out while a request is in flight, and is the
	// caller's because the request is the caller's.
	Busy bool
	// Width bounds the form; zero lets it fill what it is in.
	Width float32
}

// LoginFormResult carries a LoginForm and what was done in it.
type LoginFormResult struct {
	// Element is the whole form.
	Element *ui.Element
	// valid is whether the two fields pass as they stand.
	valid bool
}

// Valid reports whether the two fields pass the package's own rules. The
// submit button is disabled while it is false, so this is here for a caller
// that wants to say why rather than only to refuse.
func (r LoginFormResult) Valid() bool { return r.valid }

// LoginForm is an address, a password and a button: the arrangement a person
// signs in with.
//
// Almost all of it is composition. The two fields are input.TextInput and
// input.PasswordInput, the switch is input.Switch and the button is
// input.Button — all of which already know about the caret, about paste, about
// what the tab order is and about how a field looks in both appearances. What
// is added here is the order they are built in and the rule that says when
// they are wrong, and neither of those belongs to a control.
//
// The fields are built in reading order, so the tab order is the visual one.
// A form whose tab order is not its reading order walks somebody through the
// screen in a sequence they cannot see.
func LoginForm(c *ui.Context, opts LoginFormOptions) LoginFormResult {
	if opts.Email == nil || opts.Password == nil {
		panic("account: LoginForm needs an Email and a Password to point at; " +
			"the form keeps no value of its own")
	}
	if opts.Remember == nil {
		panic("account: LoginForm needs the *bool Remember writes to; a switch " +
			"with nothing behind it is a switch that does nothing")
	}
	if opts.Submit == nil {
		panic("account: LoginForm needs the *bool Submit writes to")
	}

	r := LoginFormResult{valid: opts.EmailError == "" && opts.PasswordError == ""}
	r.Element = input.Form(c, input.FormOptions{
		// Not "Sign in". The form is a group and the button under it is a
		// control, and a screen reader that hears two things called Sign in
		// cannot tell which is the one to press. Naming the group after what
		// it is fixes it.
		Label:       core.Msg(c, "account.loginForm", "Sign in details"),
		Title:       opts.Title,
		Description: opts.Subtitle,
		Width:       opts.Width,
		Divider:     true,
		Fields:      func() { loginFields(c, opts) },
		Actions:     func() { loginActions(c, opts) },
	}).Element
	return r
}

// loginFields builds the two fields and the switch under them, in reading
// order.
func loginFields(c *ui.Context, opts LoginFormOptions) {
	u := core.Density(c).Unit()
	emailLabel := core.Msg(c, "account.email", "Email")
	passLabel := core.Msg(c, "account.password", "Password")

	input.FormField(c, input.FormFieldOptions{
		Label:    emailLabel,
		Required: true,
		Error:    opts.EmailError,
	}, func(err string) *ui.Element {
		return input.TextInput(c, opts.Email, input.TextInputOptions{
			Label:       emailLabel,
			Placeholder: "you@company.com",
			Error:       err,
			Disabled:    opts.Busy,
		})
	})

	input.FormField(c, input.FormFieldOptions{
		Label:    passLabel,
		Required: true,
		Error:    opts.PasswordError,
	}, func(err string) *ui.Element {
		return input.PasswordInput(c, opts.Password, input.PasswordInputOptions{
			Label:    passLabel,
			Error:    err,
			Revealed: opts.Revealed,
			Disabled: opts.Busy,
		})
	})

	// The switch on the left of its row and the reset link pushed to the far
	// side: one says what will happen, the other is the way to say it should
	// not happen. They share a row because a forgotten password is a
	// different complaint from an unwanted session, and somebody in a hurry
	// needs both without hunting for the second one.
	ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 2).Children(func() {
		input.Switch(c, opts.Remember, input.SwitchOptions{
			Label: core.Msg(c, "account.remember", "Keep me signed in"),
		})
		ui.Box(c).Grow(1)
		if opts.Forgot != nil {
			if textButton(c, core.Msg(c, "account.forgot", "Forgot password?")).Clicked() {
				opts.Forgot()
			}
		}
	})
}

// loginActions builds the submit row. The spacer before the button is what
// pushes it to the far side of the row input.Form already laid out.
func loginActions(c *ui.Context, opts LoginFormOptions) {
	ui.Box(c).Grow(1)
	submit := input.Button(c, core.Msg(c, "account.signIn", "Sign in"), input.ButtonOptions{
		Primary:  true,
		Disabled: opts.Busy,
	})
	if submit.Clicked() {
		*opts.Submit = true
	}
}

// textButton is a pressable piece of secondary copy. It is a box rather than
// ui.Text because a link that only looks like one is not a link: it has to
// take the keyboard and be announced as a control, and it has to be
// underlined so that the way to tell it from a label does not depend on
// seeing the colour.
func textButton(c *ui.Context, label string) *ui.Element {
	k := core.Tokens(c)
	return ui.Box(c).Shrink(0).Label(label).Tooltip(label).
		Cursor(ui.CursorPointer).Children(func() {
		ui.Text(c, label).TextColor(k.Accent).
			FontSize(core.FontSize(c, theme.MetaSize)).Underline()
	})
}

// ApiKey is one credential belonging to an account.
//
// Secret is on the struct because the manager has to be able to show a key
// once at the moment it is made, and because a copy button needs something to
// copy. It is never drawn: everything on screen goes through MaskKey, and
// nothing in this package puts a Secret into a label, a tooltip or a test's
// expected text.
type ApiKey struct {
	// ID is the key's own identifier, which is what a revoke is addressed by.
	ID string
	// Name is what the account calls it — "CI", "Local development".
	Name string
	// Secret is the key itself.
	Secret string
	// Created and LastUsed are when it was made and when it was last sent,
	// already formatted: the timezone a server logs in is not this package's
	// to know. An empty LastUsed means never.
	Created, LastUsed string
	// Scope is what the key may do, which is the one line of a key list that
	// somebody reads before deciding what to revoke.
	Scope string
}

// ApiKeyManagerOptions configure an ApiKeyManager.
type ApiKeyManagerOptions struct {
	// Keys are the account's keys. The manager neither sorts nor filters
	// them: what it shows is what it was handed, in the order it was handed.
	Keys []ApiKey
	// Revoke writes the id of a revoked key into the caller's string, empty
	// when nothing was revoked this frame. A string rather than a bool
	// because "which one" is the whole question a revoke button asks.
	Revoke *string
	// Copy writes the id of the key whose copy button was pressed, empty when
	// none was.
	Copy *string
	// Name is the caller's string for the new key's name. It is required
	// when OnAdd is set, and required rather than optional for the same
	// reason every other field in this package takes one: the name is what
	// the created key will be called, so it has to survive the frame it was
	// typed in, and a name held in this component would have nowhere to go
	// when the caller wanted to check it or show it above the field.
	Name *string
	// OnAdd is the mint row's press, handed the name typed into it. Nil
	// draws no mint row, which is what a manager of keys the caller cannot
	// create wants.
	OnAdd func(name string)
}

// ApiKeyManagerResult carries an ApiKeyManager and what was done in it.
type ApiKeyManagerResult struct {
	// Element is the whole manager.
	Element *ui.Element
	// added is the name typed into the mint row.
	added string
}

// Added returns the name typed into the mint row at the moment it was
// pressed. A manager with no mint row has nothing to add and says so.
func (r ApiKeyManagerResult) Added() string { return r.added }

// ApiKeyManager is the list of credentials behind an account, with a way to
// copy one and a way to revoke one.
//
// A key is shown through MaskKey and through nothing else. There is no "show
// the whole key" button here, and there is no way to ask this component for
// the plaintext: a manager that could display a key could also be made to
// display it in a screenshot, in a log line, or in a failing test's expected
// text, and the copy button is the answer that never needs the screen to say
// it out loud.
//
// The revoke is written to the caller rather than performed. This package has
// no backend, and a component that removed a row from its own slice would
// leave the caller believing a key was revoked when nothing had been told to
// anybody.
func ApiKeyManager(c *ui.Context, opts ApiKeyManagerOptions) ApiKeyManagerResult {
	if opts.Revoke == nil {
		panic("account: ApiKeyManager needs the *string Revoke writes to")
	}
	if opts.Copy == nil {
		panic("account: ApiKeyManager needs the *string Copy writes to")
	}
	if opts.OnAdd != nil && opts.Name == nil {
		panic("account: a mint row needs the *string Name writes to; a key " +
			"whose name is only held here has no name once the frame is over")
	}
	u := core.Density(c).Unit()

	var r ApiKeyManagerResult
	mgr := ui.Column(c).FillWidth().Gap(u * 3).Role(ui.RoleList).
		Label(core.Msg(c, "account.apiKeys", "API keys"))
	mgr.Children(func() {
		if opts.OnAdd != nil {
			mintRow(c, opts, &r)
		}
		if len(opts.Keys) == 0 {
			noKeys(c)
			return
		}
		for i := range opts.Keys {
			keyRow(c, opts, &opts.Keys[i])
		}
	})
	r.Element = mgr
	return r
}

// mintRow is the field and button that make a new key.
func mintRow(c *ui.Context, opts ApiKeyManagerOptions, r *ApiKeyManagerResult) {
	u := core.Density(c).Unit()

	ui.Row(c).FillWidth().Gap(u * 2).AlignItems(ui.End).Children(func() {
		ui.Box(c).Grow(1).Shrink(0).Children(func() {
			input.TextInput(c, opts.Name, input.TextInputOptions{
				Label:       core.Msg(c, "account.keyName", "Key name"),
				Placeholder: core.Msg(c, "account.keyNameHint", "What is it for?"),
			})
		})
		mint := input.Button(c, core.Msg(c, "account.createKey", "Create key"),
			input.ButtonOptions{Primary: true, Disabled: *opts.Name == ""})
		if mint.Clicked() && *opts.Name != "" {
			r.added = *opts.Name
			opts.OnAdd(*opts.Name)
		}
	})
}

// keyRow is one credential: its name, its masked secret, what it may do and
// when it was last used, and the two things that can be done to it.
func keyRow(c *ui.Context, opts ApiKeyManagerOptions, key *ApiKey) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	row := ui.Row(c).FillWidth().Gap(u*3).AlignItems(ui.Center).
		Padding(u*2.5, u*3).Radius(theme.ControlRadius).Background(k.Surface).
		Role(ui.RoleListItem).Label(key.Name)
	row.Children(func() {
		ui.Column(c).Grow(1).Shrink(0).Gap(u * 0.5).Children(func() {
			ui.Text(c, key.Name).TextColor(k.Text).
				FontSize(core.FontSize(c, theme.BodySize)).Bold().SingleLine()
			// The mask, never the key. See ApiKeyManager.
			ui.Text(c, MaskKey(key.Secret)).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.MonoSize)).SingleLine()
			ui.Text(c, keyLine(c, key)).TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
		})

		if input.Button(c, core.Msg(c, "account.copyKey", "Copy"),
			input.ButtonOptions{}).Clicked() {
			*opts.Copy = key.ID
		}
		if input.Button(c, core.Msg(c, "account.revokeKey", "Revoke"),
			input.ButtonOptions{Danger: true}).Clicked() {
			*opts.Revoke = key.ID
		}
	})
}

// keyLine is the third line of a key row: what it may do, and when it was
// last used. Both halves are the caller's own strings — this package does not
// format a date and does not know what a server calls a scope — with a
// fallback for each so that a key with neither still says something.
func keyLine(c *ui.Context, key *ApiKey) string {
	scope := key.Scope
	if scope == "" {
		scope = core.Msg(c, "account.keyFullScope", "Full access")
	}
	used := key.LastUsed
	if used == "" {
		used = core.Msg(c, "account.keyNeverUsed", "never used")
	}
	return scope + " · " + used
}

// noKeys is what a manager with nothing in it says. It is a panel rather than
// nothing at all, because an empty column next to a heading that says "API
// keys" reads as a loading state rather than as an answer.
func noKeys(c *ui.Context) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	ui.Box(c).FillWidth().Padding(u*4, u*3).Radius(theme.ControlRadius).
		Background(k.Surface).Center().Children(func() {
		ui.Text(c, core.Msg(c, "account.noKeys", "No API keys yet")).
			TextColor(k.TextMuted).FontSize(core.FontSize(c, theme.RowSize))
	})
}

// TwoFactorInputOptions configure a TwoFactorInput.
type TwoFactorInputOptions struct {
	// Value is the caller's string, digits and no more — the shape a server
	// checks it against.
	Value *string
	// Length is how many digits the code has; zero is six.
	Length int
	// Error marks the code as not valid, which is where a code spends most
	// of its life: somebody is typing it wrong.
	Error string
	// Masked hides the digits as they go in, for a code typed where somebody
	// can see the screen.
	Masked bool
	// AutoFocus gives the first box the keyboard in the frame it appears,
	// which is what a sheet whose whole purpose is the code wants.
	AutoFocus bool
	// Disabled takes the code out of the tab order.
	Disabled bool
}

// TwoFactorInputResult carries a TwoFactorInput and how full the code is.
type TwoFactorInputResult struct {
	// Element is the row of boxes.
	Element *ui.Element
	// filled and length are this frame's code.
	filled, length int
}

// Filled reports how many digits have been typed, so a caller can offer a
// resend without reading the string itself and counting.
func (r TwoFactorInputResult) Filled() int { return r.filled }

// Complete reports that every box holds a digit, which is the one thing a
// caller waiting for a code needs to hear. It is false for a code of no
// length at all, so a misconfigured length never reads as a filled one.
func (r TwoFactorInputResult) Complete() bool {
	return r.length > 0 && r.filled == r.length
}

// TwoFactorInput is the boxes of a one-time code.
//
// It is input.PinInput with a name for what the code is for. Typing a digit
// moves the focus on by itself and emptying a box moves it back, so the code
// is entered with one hand on the keyboard and never with the pointer — which
// is the whole reason this is six boxes and not one field with a max length.
//
// Complete is reported rather than the caller counting digits: "every box
// holds something" is a question about the control, and the control is the
// only thing that knows how many boxes it drew.
func TwoFactorInput(c *ui.Context, opts TwoFactorInputOptions) TwoFactorInputResult {
	if opts.Value == nil {
		panic("account: TwoFactorInput needs a Value to point at; it keeps no " +
			"value of its own")
	}
	n := opts.Length
	if n < 1 {
		n = defaultCodeLength
	}

	pin := input.PinInput(c, opts.Value, input.PinInputOptions{
		Label:     core.Msg(c, "account.code", "Verification code"),
		Length:    n,
		Error:     opts.Error,
		Masked:    opts.Masked,
		Disabled:  opts.Disabled,
		AutoFocus: opts.AutoFocus,
	})
	return TwoFactorInputResult{
		Element: pin.Element,
		filled:  pin.Filled(),
		length:  n,
	}
}
