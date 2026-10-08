package files

import (
	"strings"
	"unicode"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// RenameInlineOptions configure a RenameInline.
type RenameInlineOptions struct {
	// Label names the field; required, as it is for every text field in the
	// library.
	Label string
	// Kind is what the thing being renamed is, so the glyph beside the
	// field is the right one before the name has been typed.
	Kind Kind
	// Original is the name it had before. It is what the field falls back
	// to when the value is emptied, and it is why emptying the field is not
	// destructive: a rename box that cleared on a backspace would rename
	// "report.pdf" to "" the first time somebody reached for the r.
	Original string
	// Height is the height of the field; zero gives the library's own.
	Height float32
	// Saving greys the field and takes it out of the tab order, for the
	// moment between pressing Return and the rename being done.
	Saving bool
}

// RenameInlineResult carries a RenameInline and what was done with it.
type RenameInlineResult struct {
	// Element is the field.
	Element *ui.Element
	// submitted reports the rename being confirmed this frame.
	submitted bool
	// reverted reports the field being put back this frame.
	reverted bool
}

// Submitted reports the rename being confirmed this frame. The name in
// Renamed is the name to use: it has been through Sanitize, so it is a name
// the filesystem will take.
func (r RenameInlineResult) Submitted() bool { return r.submitted }

// Reverted reports the field being put back this frame, which is what the
// Escape key and the cancel button do.
func (r RenameInlineResult) Reverted() bool { return r.reverted }

// RenameInline is a name being typed over, in the row it belongs to.
//
// The value is the caller's the moment it is typed, and it is the raw thing
// somebody typed rather than the cleaned one: a rename field that shows
// "report_.pdf" while somebody is typing "report?.pdf" shows them something
// they did not write, and the caret ends up somewhere they did not put it.
// The cleaning happens on submit, and Renamed says what came out.
//
// Escape puts the original back. That is not a convenience: a half-typed name
// left in a field is a name somebody will come back to and submit by
// accident, and Escape is the only key that means "never mind" anywhere else
// on the desktop either.
func RenameInline(c *ui.Context, name *string, opts RenameInlineOptions) RenameInlineResult {
	if name == nil {
		panic("files: RenameInline needs the name to point at; it keeps no name of its own")
	}
	if opts.Label == "" {
		panic("files: RenameInline needs a Label for the field")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	h := opts.Height
	if h <= 0 {
		h = core.ControlHeight(c)
	}

	var res RenameInlineResult
	well := ui.Box(c).FillWidth().Height(h).Padding(u*0.75, u*1.5).
		Radius(theme.ControlRadius).Background(k.Background).
		BorderWidth(theme.BorderWidth).BorderColor(k.Accent).Role(ui.RoleNone)
	well.Children(func() {
		ui.Row(c).Fill().AlignItems(ui.Center).Gap(u * 1.5).Children(func() {
			FileIcon(c, *name, FileIconOptions{Kind: opts.Kind, Size: u * 3.5})
			// TextInputBase is MyGo's editor — the caret, the selection, the
			// clipboard — with no look of its own; the well above gives it
			// one, and the accent hairline is what says "you are editing
			// this" rather than "this is selected".
			field := ui.TextInputBase(c, name)
			field.Fill().TextColor(k.Text).FontSize(core.FontSize(c, theme.RowSize)).
				Label(opts.Label).Role(ui.RoleTextField).Disabled(opts.Saving)
			if well.Clicked() {
				field.Focus()
			}
			if field.Submitted() {
				// The cleaning happens here rather than as somebody types,
				// so the field shows what was typed.
				cleaned := SanitizeOr(*name, opts.Original)
				*name = cleaned
				res.submitted = true
			}
			if field.Shortcut(0, ui.KeyEscape) {
				*name = opts.Original
				res.reverted = true
			}
		})
	})
	res.Element = well
	return res
}

// Person is somebody a file can be shared with.
type Person struct {
	// Name is their name.
	Name string
	// Email is how they are invited; it is drawn under the name when there
	// is room, because two people at one company share a first name.
	Email string
	// Permission is what they can do with it, as a Permission.
	Permission Permission
	// You reports that this is the reader, who is not in the list to be
	// removed from their own file.
	You bool
}

// Permission is what somebody may do with a shared file.
type Permission int

const (
	// CanView reads the file and nothing else.
	CanView Permission = iota
	// CanComment reads it and leaves notes on it.
	CanComment
	// CanEdit changes the file's contents.
	CanEdit
	// IsOwner owns it, which is what decides who may share it again.
	IsOwner
)

// permissionWord is what a permission is called on screen, and the order
// they are offered in: least first, so the arrow moves towards more.
func permissionWord(c *ui.Context, p Permission) string {
	switch p {
	case CanComment:
		return core.Msg(c, "files.canComment", core.Def("Can comment"))
	case CanEdit:
		return core.Msg(c, "files.canEdit", core.Def("Can edit"))
	case IsOwner:
		return core.Msg(c, "files.isOwner", core.Def("Owner"))
	}
	return core.Msg(c, "files.canView", core.Def("Can view"))
}

// Permissions is every permission, least first, which is what a selector
// offers. It is a function rather than a variable because the words go
// through core.Msg and must be read in the window that asked for them.
func Permissions(c *ui.Context) []Permission {
	return []Permission{CanView, CanComment, CanEdit, IsOwner}
}

// ShareDialogOptions configure a ShareDialog.
type ShareDialogOptions struct {
	// Title is the panel's heading; it is required, and it should name what
	// is being shared rather than say "Share".
	Title string
	// Subtitle is the second line: the file's name, as the path.
	Subtitle string
	// People are the ones it is shared with. The reader is one of them,
	// marked You, and is not removable.
	People []Person
	// Invite is the address box for somebody who is not on the list yet.
	// Nil draws none, which is right for a window that shares with a group
	// rather than with people.
	Invite *string
	// Width is the panel's width; zero lets it fit its content.
	Width float32
	// Actions draws the buttons along the bottom.
	Actions func()
}

// ShareDialogResult carries a ShareDialog and what was done with it.
type ShareDialogResult struct {
	// Element is the panel.
	Element *ui.Element
	// changed is the person whose permission changed this frame, -1 for
	// none.
	changed int
	// permission is what it was changed to.
	permission Permission
	// invited reports the invite button being pressed this frame.
	invited bool
	// removed is the person taken off this frame, -1 for none.
	removed int
}

// Changed returns the index of the person whose permission changed this
// frame, -1 for none.
func (r ShareDialogResult) Changed() int { return r.changed }

// Permission returns what the changed person's permission became.
func (r ShareDialogResult) Permission() Permission { return r.permission }

// Invited reports the invite being sent this frame. The address in the
// caller's string has already been trimmed; whether it is a real address is
// the caller's to know, because only they have the directory.
func (r ShareDialogResult) Invited() bool { return r.invited }

// Removed returns the index of the person taken off this frame, -1 for none.
// The reader is not removable, so the index is never the reader's own.
func (r ShareDialogResult) Removed() int { return r.removed }

// ShareDialog is who a file is shared with and what each of them may do with
// it.
//
// The panel is built here rather than handed to ui/overlay's Dialog, because
// a share dialog is not a dialog: it has no scrim, the window behind it stays
// live, and it is the only layer a person opens and closes by pressing the
// same button again. It is a panel the caller places, which is what
// layout.Panel is for — and which is why the caller owns whether it is
// showing, as it owns everything else here.
//
// Every change is a write into the caller's slice of people. A share list
// that kept a copy would be a copy that is out of date the first time
// somebody is removed by somebody else.
func ShareDialog(c *ui.Context, open *bool, opts ShareDialogOptions) ShareDialogResult {
	if open == nil {
		panic("files: ShareDialog needs the *bool it opens and closes")
	}
	if opts.Title == "" {
		panic("files: ShareDialog needs a Title that says what is being shared")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	if !*open {
		return ShareDialogResult{changed: -1, removed: -1}
	}

	var res ShareDialogResult
	res.changed = -1
	res.removed = -1

	panel := ui.Box(c).Width(opts.Width).Shrink(0)
	panel = layout.Panel(c, panel, layout.PanelOptions{
		Title: opts.Title, Subtitle: opts.Subtitle, Rule: true, Label: opts.Title,
	}, func() {
		scroll := layout.ScrollArea(c, layout.ScrollAreaOptions{
			Vertical: true, Height: 260,
		}, func() {
			body := ui.Column(c).FillWidth().Gap(u * 1.5)
			body.Children(func() {
				if opts.Invite != nil {
					inviteRow(c, opts.Invite, &res)
					layout.Divider(c, layout.DividerOptions{})
				}
				if len(opts.People) == 0 {
					ui.Text(c, core.Msg(c, "files.notShared", core.Def("Shared with nobody yet"))).
						TextColor(k.TextMuted).FontSize(core.FontSize(c, theme.RowSize))
				}
				for i := range opts.People {
					i := i
					personRow(c, i, &opts.People[i], &res)
				}
			})
		})
		scroll.Element.FillWidth()
		if opts.Actions != nil {
			opts.Actions()
		}
	})
	res.Element = panel
	return res
}

// personRow is one person and what they may do with the file.
//
// The permission is a row of buttons rather than a dropdown, for the reason
// every four-way choice in this library is: it is four options, they are all
// meaningful, and a dropdown makes somebody click twice to find out what they
// are about to change. The chosen one wears the accent, so the state of the
// whole list is readable without reading a word of it.
func personRow(c *ui.Context, i int, p *Person, res *ShareDialogResult) {
	k, u := core.Tokens(c), core.Density(c).Unit()

	row := ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u*1.5).Padding(u*0.5, 0)
	row.Children(func() {
		avatarOf(c, p.Name)
		ui.Box(c).Grow(1).Children(func() {
			ui.Text(c, p.Name).FontSize(core.FontSize(c, theme.BodySize)).SingleLine()
			if p.Email != "" {
				ui.Text(c, p.Email).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
			}
		})
		if p.You {
			ui.Text(c, core.Msg(c, "files.you", core.Def("you"))).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
		} else {
			ui.Row(c).Gap(u * 0.5).Children(func() {
				for _, perm := range Permissions(c) {
					perm := perm
					btn := permissionButton(c, p.Name, perm, p.Permission == perm)
					if btn.Clicked() {
						p.Permission = perm
						res.changed = i
						res.permission = perm
					}
				}
			})
			if removeButton(c, p.Name).Clicked() {
				res.removed = i
			}
		}
	})
	row.Label(p.Name + ", " + permissionWord(c, p.Permission))
}

// permissionButton is one of the four, wearing the accent when it is the one
// that is set.
func permissionButton(c *ui.Context, who string, perm Permission, chosen bool) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	label := permissionWord(c, perm)
	bg, fg := k.Background, k.Text
	if chosen {
		bg, fg = k.AccentBg, k.AccentText
	}
	btn := ui.ButtonBase(c).Height(u*6).Radius(theme.ControlRadius).
		Padding(0, u*1.75).Background(bg).TextColor(fg).
		Label(who + ": " + label).Tooltip(label)
	btn.Children(func() {
		ui.Text(c, label).FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
	})
	return btn
}

// removeButton is the little cross beside somebody who is not the reader.
func removeButton(c *ui.Context, who string) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	btn := ui.ButtonBase(c).Size(u*6, u*6).Shrink(0).Radius(u * 3).
		Background(k.Surface).TextColor(k.TextMuted).
		Label(core.Msg(c, "files.remove", core.Def("remove")) + " " + who).Role(ui.RoleNone)
	btn.Children(func() {
		ui.Text(c, "×").TextColor(k.TextMuted).
			FontSize(core.FontSize(c, theme.BodySize))
	})
	return btn
}

// inviteRow is the box for somebody who is not on the list yet, and the
// button that adds them.
func inviteRow(c *ui.Context, address *string, res *ShareDialogResult) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	label := core.Msg(c, "files.invite", core.Def("Invite by email or name"))

	row := ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.5)
	row.Children(func() {
		field := ui.TextInputBase(c, address)
		field.Grow(1).TextColor(k.Text).Label(label).Role(ui.RoleTextField).
			Placeholder(label)
		btn := inviteButton(c, core.Msg(c, "files.send", core.Def("Send")), *address != "")
		if btn.Clicked() && *address != "" {
			// The address is trimmed and normalised here rather than by the
			// caller, because an address with a space at the end is not an
			// address and every invite that has ever failed has failed on
			// one.
			*address = strings.TrimSpace(*address)
			res.invited = true
		}
	})
}

func inviteButton(c *ui.Context, label string, on bool) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	bg, fg := k.Surface, k.Text
	if on {
		bg, fg = k.Fill, k.OnFill
	}
	btn := ui.ButtonBase(c).Height(core.ControlHeight(c)).Radius(theme.PillRadius).
		Padding(0, u*2.5).Background(bg).TextColor(fg).
		Label(label).Disabled(!on).Children(func() {
		ui.Text(c, label).FontSize(core.FontSize(c, theme.RowSize))
	})
	return btn
}

// PermissionSelectOptions configure a PermissionSelect.
type PermissionSelectOptions struct {
	// Label names the control; required.
	Label string
	// Permissions are the options offered. Empty gives all four, least
	// first; give a shorter list to leave out the ones this file does not
	// have — nobody should be able to make themselves the owner of a shared
	// document.
	Permissions []Permission
}

// PermissionSelectResult carries a PermissionSelect and what was chosen.
type PermissionSelectResult struct {
	// Element is the row of buttons.
	Element *ui.Element
	// picked is the permission chosen this frame, nil when none was.
	picked *Permission
}

// Picked returns the permission chosen this frame, nil when none was. It is
// a pointer to a copy rather than the value so that "nothing was chosen" and
// "CanView was chosen" cannot be confused — CanView is the zero value, and a
// zero value that doubles as "nothing" is a bug waiting for the first caller
// that treats it as an answer.
func (r PermissionSelectResult) Picked() *Permission { return r.picked }

// PermissionSelect is one person's access, as a row of the four choices with
// the current one marked.
//
// It is separate from ShareDialog because it is asked for on its own: a file
// row in a list has a permissions popover, a sidebar has one per person, and
// neither of them is a dialog.
func PermissionSelect(c *ui.Context, selected *Permission, opts PermissionSelectOptions) PermissionSelectResult {
	u := core.Density(c).Unit()
	if selected == nil {
		panic("files: PermissionSelect needs a permission to point at; it keeps none of its own")
	}
	if opts.Label == "" {
		panic("files: PermissionSelect needs a Label; four buttons with no name say nothing")
	}
	perms := opts.Permissions
	if len(perms) == 0 {
		perms = Permissions(c)
	}

	var res PermissionSelectResult
	row := ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u).Label(opts.Label)
	row.Children(func() {
		for _, perm := range perms {
			perm := perm
			btn := permissionButton(c, opts.Label, perm, *selected == perm)
			if btn.Clicked() {
				*selected = perm
				picked := perm
				res.picked = &picked
			}
		}
	})
	res.Element = row
	return res
}

// avatarOf is a person's initials in a circle, which is the same mark the
// rest of the interface uses for a person and not a second one.
func avatarOf(c *ui.Context, name string) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	side := u * 7
	ui.Box(c).Size(side, side).Shrink(0).Radius(side / 2).
		Background(k.SurfacePressed).Label(name).Children(func() {
		ui.Text(c, initials(name)).TextColor(k.Text).
			FontSize(core.FontSize(c, theme.CaptionSize)).Bold()
	})
}

// initials is the first letter of each of a name's first two words, so
// "Ada Lovelace" reads AL and "Prince" reads P.
//
// It is written out rather than taken from ui/internal because the share
// panel needs it for names it was handed, not for records it has, and the
// package's own is a function over the same two words.
func initials(name string) string {
	out := make([]rune, 0, 2)
	for i, word := range strings.Fields(name) {
		if i == 2 {
			break
		}
		for _, r := range word {
			if unicode.IsLetter(r) {
				out = append(out, unicode.ToUpper(r))
				break
			}
		}
	}
	return string(out)
}
