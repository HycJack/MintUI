package account

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/theme"
)

// KeyboardShortcutsListOptions configure a KeyboardShortcutsList.
type KeyboardShortcutsListOptions struct {
	// Shortcuts are the bindings, in the caller's order. The list does not
	// group them: the headings come from each binding's own Group, and a
	// caller that wants a different order has a different order.
	Shortcuts []Shortcut
	// Recorded is the binding that was just re-recorded, by ID. It is the
	// caller's rather than the list's own so that the list can point at it:
	// a chord that has just changed is the one somebody wants to check.
	Recorded string
	// Highlight draws the Recorded binding in the accent tone, for the few
	// frames after a change. It is off by default because a settings screen
	// that flashes on every redraw is worse than one that never does.
	Highlight bool
}

// KeyboardShortcutsListResult carries a KeyboardShortcutsList.
type KeyboardShortcutsListResult struct {
	// Element is the whole list.
	Element *ui.Element
	// conflict is whether the caller's bindings cannot all be true at once.
	conflict bool
}

// Conflict reports that two bindings are the same, or two commands share a
// chord. It is a finding rather than an error: the list still draws both,
// because a settings screen that hid one of them would leave somebody
// wondering where their shortcut went. The runtime gives a key to the
// innermost element that asks for it, so one of the two silently never fires
// — and the only honest way to say so is to say it here.
func (r KeyboardShortcutsListResult) Conflict() bool { return r.conflict }

// KeyboardShortcutsList is every binding on the machine, under the heading its
// group names.
//
// Each chord is drawn with display.Kbd rather than as text. "Cmd+K" read as
// one word is impossible to scan — there is nothing in it saying where one key
// ends and the next begins — and a settings page of them is a page of words
// with no shape at all. Kbd puts each key in its own box, which is the same
// mark the toolbar uses, so a binding looks the same wherever it is shown.
func KeyboardShortcutsList(c *ui.Context, opts KeyboardShortcutsListOptions) KeyboardShortcutsListResult {
	k, u := core.Tokens(c), core.Density(c).Unit()

	var r KeyboardShortcutsListResult
	r.conflict = ChordsConflict(opts.Shortcuts)

	last := ""
	r.Element = ui.Column(c).FillWidth().Gap(u * 1).
		Label(core.Msg(c, "account.shortcuts", "Keyboard shortcuts")).Children(func() {
		if len(opts.Shortcuts) == 0 {
			ui.Text(c, core.Msg(c, "account.noShortcuts", "No shortcuts yet")).
				TextColor(k.TextFaint).FontSize(core.FontSize(c, theme.RowSize)).SingleLine()
			return
		}
		for _, s := range opts.Shortcuts {
			// A heading is written once per run of the same group, so an
			// unsorted list still reads as sections rather than as a list
			// with the heading repeated down it.
			if s.Group != "" && s.Group != last {
				last = s.Group
				ui.Text(c, s.Group).TextColor(k.TextMuted).Bold().
					FontSize(core.FontSize(c, theme.CaptionSize)).
					Margin(u*1.5, u, 0, u*0.5).SingleLine()
			}
			shortcutRow(c, opts, s)
		}
	})
	return r
}

// shortcutRow is one binding: what it does, and what does it.
func shortcutRow(c *ui.Context, opts KeyboardShortcutsListOptions, s Shortcut) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	justRecorded := opts.Highlight && s.ID == opts.Recorded

	ui.Row(c).FillWidth().Gap(u*3).AlignItems(ui.Center).
		Padding(u*0.75, u*2).Radius(theme.ControlRadius).Label(s.Name).
		Role(ui.RoleListItem).Children(func() {
		ui.Text(c, s.Name).TextColor(k.Text).Grow(1).
			FontSize(core.FontSize(c, theme.RowSize)).SingleLine()
		if justRecorded {
			// The accent is the only colour in this list, so "this is the one
			// you just changed" cannot be mistaken for a property of the
			// binding itself.
			//
			// A mark of the row's own height, not a rule across it: a
			// FillWidth here is a hundred per cent of the row, which pushes
			// the name and the keys off the end of it and leaves the list
			// showing one empty line where a binding should be.
			ui.Box(c).Width(theme.BorderWidth*2).Height(u*4).Shrink(0).
				Radius(theme.BorderWidth).Background(k.Accent).Margin(u, 0)
		}
		display.Kbd(c, ChordKeys(s.Keys)...)
	})
}

// ShortcutRecorderOptions configure a ShortcutRecorder.
type ShortcutRecorderOptions struct {
	// Label is what the binding does, which is also the row's name.
	Label string
	// Value is the caller's chord, written as a space-separated list of key
	// names the moment one is pressed — "cmd k". It is the form a settings
	// file holds and the form ChordsConflict compares, so a recording is
	// comparable the instant it exists.
	Value *[]string
	// Mods are the modifiers the recorder watches for. They are an argument
	// rather than something recorded, because MyGo's Shortcut matches one
	// exact combination and cannot ask for "whatever modifier is down": a
	// recorder that could would have to be told which key was pressed first,
	// and the modifier is the part an application has already decided.
	Mods ui.Modifiers
	// Disabled takes the recorder out of the tab order. As everywhere, the
	// press is refused in the caller's own code and not by the greying.
	Disabled bool
}

// ShortcutRecorderResult carries a ShortcutRecorder and what it recorded.
type ShortcutRecorderResult struct {
	// Element is the button.
	Element *ui.Element
	// recorded is the key name pressed this frame, empty when none was.
	recorded string
}

// Recorded returns the key name the recorder caught this frame, empty when
// none was. The caller writes it into its own chord; this component has no
// binding of its own to write it into.
func (r ShortcutRecorderResult) Recorded() string { return r.recorded }

// ShortcutRecorder is a button that watches for the next key and says what it
// caught.
//
// It cannot be "press any key": MyGo's Shortcut matches one exact modifier
// and key combination, and there is no way to ask it for whichever key is
// next. So the recorder takes the modifier as an argument — almost always
// ui.Cmd, the one a command shortcut is bound to — and asks about each key it
// is willing to record, in order, taking the first that fires. That is why
// Mods is required and why the loop below is over a fixed list: it is the
// only way the underlying API can be asked the question, and pretending
// otherwise would produce a control that silently records nothing.
//
// The loop costs one Shortcut call per candidate per frame. It is bounded and
// small, and it is only paid by a screen that is showing a recorder at all.
func ShortcutRecorder(c *ui.Context, opts ShortcutRecorderOptions) ShortcutRecorderResult {
	if opts.Value == nil {
		panic("account: ShortcutRecorder needs the *[]string Value writes to")
	}
	if opts.Label == "" {
		panic("account: ShortcutRecorder needs a Label; a button showing only " +
			"the chord it will record has nothing saying what it is for")
	}
	k := core.Tokens(c)

	var r ShortcutRecorderResult
	// The visible label is empty on purpose: a button draws its own label
	// text and its children, and both saying "Press keys" — or both saying
	// the chord, once as text and once as key boxes — reads as two of it.
	// The accessible name stays the recorder's own label, which is how this
	// control is found and pressed.
	btn := input.Button(c, "", input.ButtonOptions{
		Label:    opts.Label,
		Disabled: opts.Disabled,
	})
	// Clicking the button clears the binding rather than recording anything:
	// a chord with no keys is "unbound", which is a real answer, and it is
	// the only answer a click can give.
	if btn.Clicked() && !opts.Disabled {
		*opts.Value = nil
	}
	btn.Children(func() {
		if len(*opts.Value) == 0 {
			ui.Text(c, core.Msg(c, "recordShortcut", "Press keys")).
				TextColor(k.TextFaint).FontSize(core.FontSize(c, theme.MetaSize))
		} else {
			display.Kbd(c, ChordKeys(*opts.Value)...)
		}
	})

	for _, cand := range recordable {
		if btn.Shortcut(opts.Mods, cand.key) {
			*opts.Value = NormalKeys(append(modifierNames(opts.Mods), cand.name))
			r.recorded = cand.name
			break
		}
	}
	r.Element = btn
	return r
}

// chordText is what a recorder's button says when it is not drawing the keys
// as boxes, which in this component is never — it is here so that a caller
// putting the same chord in a table cell has one function for it.
func chordText(c *ui.Context, keys []string) string {
	if len(keys) == 0 {
		return core.Msg(c, "recordShortcut", "Press keys")
	}
	return FormatChord(keys)
}

// candidate is one key the recorder will listen for.
type candidate struct {
	key  ui.Key
	name string
}

// recordableKeys is every key the recorder will listen for, in a fixed order.
//
// The order is the order of the row above: modifiers are not in it, because a
// modifier on its own is not a shortcut anybody binds. Everything else is,
// including the function keys, because a shortcuts screen with no F-keys in
// it is a shortcuts screen for people who have never needed one.
//
// It is a package-level slice rather than something built per frame because
// it never changes, and the loop in ShortcutRecorder runs it every frame.
var recordable = func() []candidate {
	var out []candidate
	add := func(k ui.Key, name string) { out = append(out, candidate{key: k, name: name}) }

	add(ui.KeyEnter, "enter")
	add(ui.KeyEscape, "esc")
	add(ui.KeySpace, "space")
	add(ui.KeyTab, "tab")
	add(ui.KeyBackspace, "backspace")
	add(ui.KeyDelete, "delete")
	add(ui.KeyHome, "home")
	add(ui.KeyEnd, "end")
	add(ui.KeyPageUp, "pageup")
	add(ui.KeyPageDown, "pagedown")
	add(ui.KeyLeft, "left")
	add(ui.KeyRight, "right")
	add(ui.KeyUp, "up")
	add(ui.KeyDown, "down")

	for i := 0; i <= 9; i++ {
		add(ui.Key0+ui.Key(i), string(rune('0'+i)))
	}
	for i := 0; i < 26; i++ {
		add(ui.KeyA+ui.Key(i), string(rune('A'+i)))
	}
	for i := 1; i <= 12; i++ {
		add(ui.KeyF1+ui.Key(i-1), "f"+itoa(i))
	}
	return out
}()

// modifierNames is the modifier half of a chord, in the order a person would
// write it: cmd, ctrl, alt, shift — the order the platform's own menus use,
// so a recorded chord reads the same as a printed one.
func modifierNames(mods ui.Modifiers) []string {
	var out []string
	switch {
	case mods&ui.Super != 0:
		out = append(out, "cmd")
	case mods&ui.Ctrl != 0:
		out = append(out, "ctrl")
	}
	if mods&ui.Alt != 0 {
		out = append(out, "alt")
	}
	if mods&ui.Shift != 0 {
		out = append(out, "shift")
	}
	return out
}
