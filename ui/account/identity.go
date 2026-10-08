package account

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/theme"
)

// Profile is the person behind the account, as far as the interface is
// concerned: the name and address the window shows, and nothing about how
// either is verified.
type Profile struct {
	// ID is the account's identifier, for a deep link and for the header.
	ID string
	// Name is the display name, and is what the avatar's initials come from.
	Name string
	// Email is the address the account signs in with.
	Email string
	// Role is the one line about the person — "Technician", "Owner". It is
	// shown where it is shown in the app; this package only carries it.
	Role string
}

// ProfileEditorOptions configure a ProfileEditor.
type ProfileEditorOptions struct {
	// Name and Email are the caller's strings, written as they are typed.
	Name, Email *string
	// Save writes true when the save button is pressed.
	Save *bool
	// Avatar is the account's picture. Nil draws the initials instead, which
	// is what a profile nobody has uploaded a photograph for looks like.
	Avatar *ui.Bitmap
	// Role is shown under the name and is not editable here: a person's role
	// is somebody else's decision, and a field that could edit it would
	// suggest otherwise.
	Role string
	// Errors are the two messages ValidateLogin would give for the address,
	// plus one the caller owns for the name. Empty for each leaves the field
	// alone.
	NameError, EmailError string
	// ReadOnlyEmail is greyed out, for an address that is verified against
	// the mail server and changing it is a separate journey.
	ReadOnlyEmail bool
	// Busy greys the form out while a save is in flight.
	Busy bool
}

// ProfileEditorResult carries a ProfileEditor and what was done in it.
type ProfileEditorResult struct {
	// Element is the whole editor.
	Element *ui.Element
	// saved reports a press of the save button this frame.
	saved bool
}

// Saved reports a press of the save button. The caller reads it on the frame
// after the press — the frame's own last pass reports nothing — so the save
// has to be a flag rather than an action taken here, since this package has
// nothing to save into.
func (r ProfileEditorResult) Saved() bool { return r.saved }

// ProfileEditor is a person's own name and address, with a photograph.
//
// The avatar is the first thing in the panel because it is the one part of a
// profile that is not a text field, and putting it above the fields means the
// panel reads as a person before it reads as a form. It is display.Image when
// there is a picture and display.Avatar's initials when there is not, so a
// profile with no upload looks deliberate rather than broken.
func ProfileEditor(c *ui.Context, opts ProfileEditorOptions) ProfileEditorResult {
	if opts.Name == nil || opts.Email == nil {
		panic("account: ProfileEditor needs a Name and an Email to point at; " +
			"the editor keeps no value of its own")
	}
	if opts.Save == nil {
		panic("account: ProfileEditor needs the *bool Save writes to")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	var r ProfileEditorResult
	r.Element = input.Form(c, input.FormOptions{
		Label: core.Msg(c, "account.profile", "Profile"),
		Fields: func() {
			ui.Row(c).FillWidth().Gap(u*4).AlignItems(ui.Center).Margin(0, 0, 0, u).Children(func() {
				if opts.Avatar != nil {
					display.Image(c, opts.Avatar, display.ImageOptions{
						Name:   *opts.Name,
						Width:  u * 15,
						Height: u * 15,
						Ratio:  1,
						Cover:  true,
						Radius: u * 7.5,
					})
				} else {
					display.Avatar(c, *opts.Name)
				}
				ui.Column(c).Grow(1).Shrink(0).Gap(u * 0.5).Children(func() {
					ui.Text(c, *opts.Name).TextColor(k.Text).Bold().
						FontSize(core.FontSize(c, theme.BodySize)).SingleLine()
					ui.Text(c, opts.Role).TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.MetaSize)).SingleLine()
				})
			})

			nameLabel := core.Msg(c, "account.displayName", "Display name")
			input.FormField(c, input.FormFieldOptions{
				Label: nameLabel,
				Error: opts.NameError,
			}, func(err string) *ui.Element {
				return input.TextInput(c, opts.Name, input.TextInputOptions{
					Label:       nameLabel,
					Placeholder: core.Msg(c, "account.namePlaceholder", "How your name appears"),
					Error:       err,
					Disabled:    opts.Busy,
				})
			})

			emailLabel := core.Msg(c, "account.email", "Email")
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
					ReadOnly:    opts.ReadOnlyEmail,
				})
			})
		},
		Actions: func() {
			ui.Box(c).Grow(1)
			if input.Button(c, core.Msg(c, "account.save", "Save changes"),
				input.ButtonOptions{Primary: true, Disabled: opts.Busy}).Clicked() {
				r.saved = true
				*opts.Save = true
			}
		},
	}).Element
	return r
}

// Account is one identity a window can be signed in as.
type Account struct {
	// ID is what the switcher's choice is written to.
	ID string
	// Name and Email are what the row shows.
	Name, Email string
	// Kind is the badge on the row — "Work", "Personal" — which is what
	// stops somebody with two accounts of the same name from picking the
	// wrong one.
	Kind string
}

// AccountSwitcherOptions configure an AccountSwitcher.
type AccountSwitcherOptions struct {
	// Accounts are the identities to switch between, and Selected is the
	// caller's choice among them. The switcher holds neither: which account
	// is current is a fact about the session, and a component that kept a
	// copy of it would be a second answer to a question somebody else is
	// already answering.
	Accounts []Account
	// Selected is the chosen account's ID, empty for none.
	Selected *string
	// Open is the panel's open state; the caller closes it on a choice, on
	// Escape and on a press outside, because when to close is a decision
	// about the screen and not about a menu.
	Open *bool
	// Switched writes the chosen account's ID when one is taken. Empty when
	// the panel was dismissed without choosing, which is the frame a caller
	// re-applies its filters on.
	Switched *string
	// AddAccount draws an extra row at the bottom. Nil draws none.
	AddAccount func()
}

// AccountSwitcherResult carries an AccountSwitcher and what was chosen in it.
type AccountSwitcherResult struct {
	// Element is the trigger, with the panel built under it while it shows.
	Element *ui.Element
}

// AccountSwitcher is the identity in the corner: the current account, and the
// others it can be swapped for.
//
// It is a trigger and a panel rather than a select because a Select's trigger
// shows the value and nothing else, and an account row is a name, an address
// and a badge. The trigger therefore builds its own row and hands it to
// layout.Panel, which is the library's one answer to what a floating layer
// looks like — so this panel has the same radius, hairline, shadow and title
// bar as every other one, rather than being the one that is slightly
// different.
//
// Choosing writes the caller's Switched and leaves closing the panel to the
// caller, for the same reason a DatePicker does: a switcher in a menu bar
// wants the panel to close on a choice, and the same switcher in a sidebar
// wants it to stay open.
func AccountSwitcher(c *ui.Context, opts AccountSwitcherOptions) AccountSwitcherResult {
	if opts.Selected == nil {
		panic("account: AccountSwitcher needs the *string Selected writes to")
	}
	if opts.Open == nil {
		panic("account: AccountSwitcher needs the *bool its panel opens in")
	}
	if opts.Switched == nil {
		panic("account: AccountSwitcher needs the *string Switched writes to")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	trigger := ui.Row(c).FillWidth().Gap(u*2).AlignItems(ui.Center).
		Padding(u*1.5, u*2.5).Radius(theme.ControlRadius).Background(k.Surface).
		Cursor(ui.CursorPointer).
		Label(switcherLabel(c, opts.Selected, opts.Accounts)).
		Tooltip(core.Msg(c, "account.switch", "Switch account"))
	if trigger.Clicked() {
		*opts.Open = !*opts.Open
	}
	trigger.Children(func() {
		display.Avatar(c, nameOf(opts.Selected, opts.Accounts))
		ui.Text(c, nameOf(opts.Selected, opts.Accounts)).TextColor(k.Text).
			FontSize(core.FontSize(c, theme.RowSize)).Bold().Grow(1).SingleLine()
		ui.Text(c, core.Msg(c, "account.switchHint", "Switch")).TextColor(k.TextMuted).
			FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0).SingleLine()
	})

	ui.PopoverBase(c, trigger, opts.Open, func(panel *ui.Element) {
		panelFace(c, panel, panelOptions{
			label: core.Msg(c, "account.accounts", "Accounts"),
			body: func() {
				for _, a := range opts.Accounts {
					accountRow(c, opts, a)
				}
				if opts.AddAccount != nil {
					ui.Box(c).FillWidth().Height(theme.BorderWidth).
						MarginY(u).Background(k.Border)
					if input.Button(c, core.Msg(c, "account.addAccount", "Add an account"),
						input.ButtonOptions{Label: core.Msg(c, "account.addAccount", "Add an account")}).Clicked() {
						opts.AddAccount()
					}
				}
			},
		})
	})
	return AccountSwitcherResult{Element: trigger}
}

// accountRow is one identity in the switcher's panel, marked as the current
// one when it is.
func accountRow(c *ui.Context, opts AccountSwitcherOptions, a Account) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	current := a.ID == *opts.Selected

	row := ui.Row(c).FillWidth().Gap(u*2).AlignItems(ui.Center).
		Padding(u*1.25, u*2).Radius(theme.ControlRadius).
		Cursor(ui.CursorPointer).Label(a.Name).Role(ui.RoleMenuItem)
	if current {
		// The current account is tinted rather than ticked. A tick is a mark
		// that has to be found among the rows; a tint is the row itself, and
		// it stays legible in both appearances because it comes from the
		// accent surface rather than from a raw colour.
		row.Background(k.SurfaceHover)
	}
	if row.Clicked() {
		*opts.Switched = a.ID
		*opts.Selected = a.ID
	}
	row.Children(func() {
		display.Avatar(c, a.Name)
		ui.Column(c).Grow(1).Shrink(0).Gap(u * 0.25).Children(func() {
			ui.Text(c, a.Name).TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).Bold().SingleLine()
			ui.Text(c, a.Email).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
		})
		if a.Kind != "" {
			display.Tag(c, a.Kind, display.TagOptions{})
		}
	})
}

// nameOf is the chosen account's name, or the first account's when the choice
// names one that is not in the list. It is a value rather than a panic because
// the account being signed in as is frequently not in the list of accounts
// one can switch to — a session restored after somebody was removed.
func nameOf(selected *string, list []Account) string {
	for _, a := range list {
		if a.ID == *selected {
			return a.Name
		}
	}
	if len(list) > 0 {
		return list[0].Name
	}
	return ""
}

// switcherLabel is what the trigger is called for assistive technology: the
// account it will switch from, and what pressing it does. A row showing only
// a name says "Rosa Vidal" and nothing about what Rosa Vidal is for.
func switcherLabel(c *ui.Context, selected *string, list []Account) string {
	name := nameOf(selected, list)
	if name == "" {
		return core.Msg(c, "account.switch", "Switch account")
	}
	return name + " — " + core.Msg(c, "account.switch", "Switch account")
}

// Workspace is one team inside an account.
type Workspace struct {
	// ID is what the switcher's choice is written to.
	ID string
	// Name is what the row shows.
	Name string
	// Seats is how many people are in it, which is the one number that tells
	// two workspaces of the same name apart.
	Seats int
	// Locked marks a workspace that cannot be switched into, which is how a
	// trial or a suspended one shows without being hidden: an account that
	// quietly stops listing a workspace looks like data loss.
	Locked bool
}

// WorkspaceSwitcherOptions configure a WorkspaceSwitcher.
type WorkspaceSwitcherOptions struct {
	// Workspaces are the teams, and Selected is the caller's choice among
	// them, for the same reason AccountSwitcher holds none: the current
	// workspace is a fact about the session.
	Workspaces []Workspace
	// Selected is the chosen workspace's ID, empty for none.
	Selected *string
	// Open is the panel's open state.
	Open *bool
	// Switched writes the chosen workspace's ID when one is taken.
	Switched *string
}

// WorkspaceSwitcherResult carries a WorkspaceSwitcher and its trigger.
type WorkspaceSwitcherResult struct {
	// Element is the trigger, with the panel built under it while it shows.
	Element *ui.Element
}

// WorkspaceSwitcher is the team the window is looking at, and the others it
// can be moved to.
//
// It is the same shape as AccountSwitcher and for the same reason: the thing
// being switched between is one per window and only ever lives in one corner.
// Two components for the two levels rather than one component with a level
// flag, because the rows are different — an account row has an address and a
// badge, a workspace row has a seat count — and a component that drew both
// would need to know which kind of thing it was drawing on every row.
//
// A locked workspace is shown greyed out rather than hidden, and pressing it
// does nothing. That is the one case where Disabled is not enough on its own:
// MyGo's Disabled only greys the box, so the row is marked disabled and the
// press is refused in the row's own code. See docs/design-system.md §17.3.
func WorkspaceSwitcher(c *ui.Context, opts WorkspaceSwitcherOptions) WorkspaceSwitcherResult {
	if opts.Selected == nil {
		panic("account: WorkspaceSwitcher needs the *string Selected writes to")
	}
	if opts.Open == nil {
		panic("account: WorkspaceSwitcher needs the *bool its panel opens in")
	}
	if opts.Switched == nil {
		panic("account: WorkspaceSwitcher needs the *string Switched writes to")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	current := workspaceOf(opts.Selected, opts.Workspaces)
	trigger := ui.Row(c).FillWidth().Gap(u*2).AlignItems(ui.Center).
		Padding(u*1.5, u*2.5).Radius(theme.ControlRadius).Background(k.Surface).
		Cursor(ui.CursorPointer).
		Label(workspaceLabel(c, current)).
		Tooltip(core.Msg(c, "account.switchWorkspace", "Switch workspace"))
	if trigger.Clicked() {
		*opts.Open = !*opts.Open
	}
	trigger.Children(func() {
		// The mark before the name rather than an avatar: a workspace is a
		// team and has nobody's initials to show, so the mark is the one
		// piece of vocabulary here that says "more than one of you".
		ui.Box(c).Size(u*5, u*5).Shrink(0).Radius(u * 1.5).Background(k.SurfaceHover).
			Children(func() {
				ui.Icon(c, glyphTeam).TextColor(k.TextMuted).Size(u*3, u*3)
			})
		ui.Text(c, current.Name).TextColor(k.Text).
			FontSize(core.FontSize(c, theme.RowSize)).Bold().Grow(1).SingleLine()
		ui.Text(c, core.Msg(c, "account.switchHint", "Switch")).TextColor(k.TextMuted).
			FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0).SingleLine()
	})

	ui.PopoverBase(c, trigger, opts.Open, func(panel *ui.Element) {
		panelFace(c, panel, panelOptions{
			label: core.Msg(c, "account.workspaces", "Workspaces"),
			body: func() {
				for _, w := range opts.Workspaces {
					workspaceRow(c, opts, w)
				}
			},
		})
	})
	return WorkspaceSwitcherResult{Element: trigger}
}

// workspaceRow is one team in the switcher's panel.
func workspaceRow(c *ui.Context, opts WorkspaceSwitcherOptions, w Workspace) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	current := w.ID == *opts.Selected

	row := ui.Row(c).FillWidth().Gap(u*2).AlignItems(ui.Center).
		Padding(u*1.25, u*2).Radius(theme.ControlRadius).
		Cursor(ui.CursorPointer).Label(w.Name).Role(ui.RoleMenuItem).
		Disabled(w.Locked)
	if current {
		row.Background(k.SurfaceHover)
	}
	// The guard as well as the greying: Disabled draws, it does not refuse.
	// See docs/design-system.md §17.3.
	if row.Clicked() && !w.Locked {
		*opts.Switched = w.ID
		*opts.Selected = w.ID
	}
	row.Children(func() {
		ui.Box(c).Grow(1).Shrink(0).Children(func() {
			ui.Text(c, w.Name).TextColor(nameInk(k, w.Locked)).
				FontSize(core.FontSize(c, theme.RowSize)).Bold().SingleLine()
			ui.Text(c, itoa(w.Seats)+" "+core.Msg(c, "account.seats", "seats")).
				TextColor(k.TextFaint).FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
		})
		if w.Locked {
			display.Tag(c, core.Msg(c, "account.locked", "Locked"),
				display.TagOptions{Tone: core.Warning})
		}
	})
}

// nameInk is the text colour for a row that cannot be pressed: the faint tone,
// which is the palette's own way of saying "this is here and it is not
// available" without inventing a second disabled colour.
func nameInk(k theme.Tokens, locked bool) ui.Color {
	if locked {
		return k.TextFaint
	}
	return k.Text
}

// workspaceOf is the chosen workspace, or the first one when the choice names
// something not in the list — which is what a window open against a workspace
// somebody was removed from looks like, and it should still name one.
func workspaceOf(selected *string, list []Workspace) Workspace {
	for _, w := range list {
		if w.ID == *selected {
			return w
		}
	}
	if len(list) > 0 {
		return list[0]
	}
	return Workspace{}
}

// workspaceLabel is what the trigger is called out loud.
func workspaceLabel(c *ui.Context, w Workspace) string {
	verb := core.Msg(c, "account.switchWorkspace", "Switch workspace")
	if w.Name == "" {
		return verb
	}
	return w.Name + " — " + verb
}

// UserMenu is one entry in the account menu.
type UserMenuItem struct {
	// ID is what a caller acts on when the entry is pressed.
	ID string
	// Label is what the entry says.
	Label string
	// Danger marks an entry that destroys something, so that it is drawn in
	// the danger colour rather than being left to a warning dialog to
	// announce what it is about.
	Danger bool
	// Disabled greys an entry out. As everywhere, the press is refused in
	// the caller's own code rather than by the grey.
	Disabled bool
}

// UserMenuOptions configure a UserMenu.
type UserMenuOptions struct {
	// Profile is whose menu it is; an empty profile draws an empty circle
	// rather than nothing, so the trigger never changes size.
	Profile Profile
	// Items are the entries, in order, with a hairline above the first of
	// the last group.
	Items []UserMenuItem
	// Open is the panel's open state.
	Open *bool
	// Chosen writes the pressed entry's ID, empty when nothing was pressed.
	Chosen *string
	// SignOut is the last entry's press, asked separately because it is the
	// one a caller always has and the one a test always wants to reach.
	SignOut func()
	// Label names the menu for assistive technology; empty uses the
	// profile's name, which is what a menu about a person is called.
	Label string
}

// UserMenuResult carries a UserMenu and its trigger.
type UserMenuResult struct {
	// Element is the trigger, with the panel built under it while it shows.
	Element *ui.Element
}

// UserMenu is the avatar in the corner and the menu under it.
//
// The trigger is an avatar and nothing else. A button with the person's name
// beside it is a button whose width changes with the length of their name,
// which moves everything to the right of it whenever somebody signs in as
// somebody else. The name is inside the panel instead, where it does not push
// the window around.
func UserMenu(c *ui.Context, opts UserMenuOptions) UserMenuResult {
	if opts.Open == nil {
		panic("account: UserMenu needs the *bool its menu opens in")
	}
	if opts.Chosen == nil {
		panic("account: UserMenu needs the *string Chosen writes to")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	name := opts.Label
	if name == "" {
		name = opts.Profile.Name
	}
	trigger := display.Avatar(c, name)
	trigger.Cursor(ui.CursorPointer)
	trigger.Tooltip(core.Msg(c, "account.menu", "Account menu"))
	if trigger.Clicked() {
		*opts.Open = !*opts.Open
	}

	ui.PopoverBase(c, trigger, opts.Open, func(panel *ui.Element) {
		panelFace(c, panel, panelOptions{
			compact: true,
			label:   name,
			body: func() {
				ui.Column(c).FillWidth().Gap(u*0.25).Margin(0, 0, 0, u).Children(func() {
					ui.Text(c, opts.Profile.Name).TextColor(k.Text).Bold().
						FontSize(core.FontSize(c, theme.RowSize)).SingleLine()
					ui.Text(c, opts.Profile.Email).TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
				})
				ui.Box(c).FillWidth().Height(theme.BorderWidth).MarginY(u * 0.5).
					Background(k.Border)
				for _, item := range opts.Items {
					menuRow(c, opts, item)
				}
				if opts.SignOut != nil {
					ui.Box(c).FillWidth().Height(theme.BorderWidth).MarginY(u * 0.5).
						Background(k.Border)
					leave := input.Button(c, core.Msg(c, "account.signOut", "Sign out"),
						input.ButtonOptions{Danger: true})
					if leave.Clicked() {
						opts.SignOut()
					}
				}
			},
		})
	})
	return UserMenuResult{Element: trigger}
}

// menuRow is one entry of the menu.
func menuRow(c *ui.Context, opts UserMenuOptions, item UserMenuItem) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	row := ui.Row(c).FillWidth().Padding(u*1.25, u*2).Radius(theme.ControlRadius).
		AlignItems(ui.Center).Cursor(ui.CursorPointer).Role(ui.RoleMenuItem).
		Label(item.Label).Disabled(item.Disabled)
	// The guard as well as the greying; see docs/design-system.md §17.3.
	if row.Clicked() && !item.Disabled {
		*opts.Chosen = item.ID
	}
	row.Children(func() {
		ink := k.Text
		if item.Danger {
			ink = k.Danger
		}
		if item.Disabled {
			ink = k.TextFaint
		}
		ui.Text(c, item.Label).TextColor(ink).
			FontSize(core.FontSize(c, theme.RowSize)).SingleLine()
	})
}
