package input

import (
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
)

// The text controls: one line, many lines, hidden, searched, masked, and the
// fixed-length code. They are all MyGo's editor inside this package's box —
// what is worth writing here is what each one does with what is typed: where
// the money's separators go, which digits a mask will take, and when the
// focus moves on by itself.

// TextInputOptions configure a TextInput.
type TextInputOptions struct {
	// Label names the field. It is required: a field a screen reader cannot
	// name, and a test cannot click, is a mystery box.
	Label string
	// Placeholder shows while the value is empty. It is a hint, never a
	// label: it disappears the moment there is a value, which is exactly
	// when somebody needs to read the field's name.
	Placeholder string
	// Error marks the value as not valid and turns the hairline around the
	// well to the danger colour. FormField draws the message itself; this
	// is the same message, for a field used on its own.
	Error string
	// Disabled greys the field and takes it out of the tab order.
	Disabled bool
	// ReadOnly keeps the value selectable and copyable but not editable,
	// for something the app fills in.
	ReadOnly bool
	// Width is the field's own width; zero fills the row it is in.
	Width float32
}

func (o TextInputOptions) skin() fieldSkin {
	return fieldSkin{
		label: o.Label, err: o.Error,
		disabled: o.Disabled, readOnly: o.ReadOnly, width: o.Width,
	}
}

// TextInput is one line of text, editing the string the caller owns. It is
// the ground every other text control here is a variation of, so a form is
// built out of these rather than out of four kinds of box.
//
// What is typed is the caller's value the moment it is typed: the field is a
// view of one value rather than a copy of it, so a form that validates as it
// is filled in needs no commit step.
func TextInput(c *ui.Context, value *string, opts TextInputOptions) *ui.Element {
	if value == nil {
		panic("input: TextInput needs a value to point at; it keeps no value of its own")
	}
	if opts.Label == "" {
		panic("input: TextInput needs a Label, or nothing can name the field")
	}
	u := core.Density(c).Unit()

	return textWell(c, opts.skin(), func(_ *ui.Element) *ui.Element {
		in := bareInput(c, value, opts.Label, opts.Placeholder).
			Grow(1).MinHeight(core.ControlHeight(c) - u*3)
		return dressInput(in, opts.Error, opts.ReadOnly)
	})
}

// dressInput applies the two settings every bare input in this package takes
// the same way, so a read-only field is read-only everywhere and an error
// reaches the control itself for assistive technology as well as for the
// hairline.
func dressInput(in *ui.Element, err string, readOnly bool) *ui.Element {
	if readOnly {
		in.ReadOnly(true)
	}
	if err != "" {
		// On a bare input MyGo's Error adds no words — it has no field to
		// put them in — so this only marks the value as not valid and
		// hands the message to the reader. FormField is what shows it.
		in.Error(err)
	}
	return in
}

// TextAreaOptions configure a TextArea.
type TextAreaOptions struct {
	// Label names the field; it is required, as for a TextInput.
	Label string
	// Placeholder shows while the value is empty.
	Placeholder string
	// Error marks the value as not valid.
	Error string
	// Lines is how many lines the field shows. Four is the usual answer
	// for a note, and anything below two is a TextInput wearing a hat.
	Lines int
	// Disabled and ReadOnly do what they do on a TextInput.
	Disabled bool
	ReadOnly bool
	// Width is the field's own width; zero fills the row it is in.
	Width float32
}

func (o TextAreaOptions) skin() fieldSkin {
	lines := o.Lines
	if lines < 1 {
		lines = 4
	}
	return fieldSkin{
		label: o.Label, err: o.Error, lines: lines,
		disabled: o.Disabled, readOnly: o.ReadOnly, width: o.Width,
	}
}

// TextArea is as many lines of text as the note needs, editing the caller's
// string. The box is as tall as its lines and no taller: past that the text
// scrolls inside the field rather than the form growing down the page.
//
// A text area hides nothing on any platform, so a password typed in one is a
// password with witnesses — PasswordInput is the one that hides.
func TextArea(c *ui.Context, value *string, opts TextAreaOptions) *ui.Element {
	if value == nil {
		panic("input: TextArea needs a value to point at; it keeps no value of its own")
	}
	if opts.Label == "" {
		panic("input: TextArea needs a Label, or nothing can name the field")
	}
	s := opts.skin()
	h := float32(s.lines) * lineHeight(c)

	return areaWell(c, s, func(_ *ui.Element) *ui.Element {
		in := ui.TextAreaBase(c, value).Label(opts.Label).FillWidth().Height(h)
		if opts.Placeholder != "" {
			in.Placeholder(opts.Placeholder)
		}
		return dressInput(in, opts.Error, opts.ReadOnly)
	})
}

// PasswordInputOptions configure a PasswordInput.
type PasswordInputOptions struct {
	// Label names the field; it is required.
	Label string
	// Placeholder shows while the value is empty.
	Placeholder string
	// Error marks the value as not valid.
	Error string
	// Revealed is the caller's "show the password" switch. It is a pointer
	// rather than state of this package's own so that the answer belongs
	// to the view that draws the other half of it — a toolbar button that
	// reveals a form's field, say — and so that two fields can never
	// disagree about what is revealed. Nil is always hidden.
	Revealed *bool
	// Disabled and ReadOnly do what they do on a TextInput.
	Disabled bool
	ReadOnly bool
	// Width is the field's own width; zero fills the row it is in.
	Width float32
}

// PasswordInput is a TextInput that hides what it holds, with a key to look
// at it by.
//
// The eye takes no stop of the tab order: everything it does, the field
// itself already does — the arrows move the caret and Backspace deletes —
// and a second stop between two fields is a tab that goes nowhere. It is
// still a button to a screen reader and a press to a test.
func PasswordInput(c *ui.Context, value *string, opts PasswordInputOptions) *ui.Element {
	if value == nil {
		panic("input: PasswordInput needs a value to point at; it keeps no value of its own")
	}
	if opts.Label == "" {
		panic("input: PasswordInput needs a Label, or nothing can name the field")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	s := fieldSkin{
		label: opts.Label, err: opts.Error,
		disabled: opts.Disabled, readOnly: opts.ReadOnly, width: opts.Width,
	}
	shown := opts.Revealed != nil && *opts.Revealed

	return textWell(c, s, func(_ *ui.Element) *ui.Element {
		in := bareInput(c, value, opts.Label, opts.Placeholder).
			Grow(1).MinHeight(core.ControlHeight(c) - u*3)
		if !shown {
			in.Password()
		}
		dressInput(in, opts.Error, opts.ReadOnly)

		glyph, name := glyphEye, core.Msg(c, "input.showPassword", "Show password")
		if shown {
			glyph, name = glyphBlind, core.Msg(c, "input.hidePassword", "Hide password")
		}
		if opts.Revealed != nil {
			// A key that has nothing to write to is not drawn: a control
			// that cannot do what it says is worse than no control, because
			// somebody will press it and conclude the password is broken.
			eye := ui.Box(c).Size(u*7, u*7).Shrink(0).Role(ui.RoleButton).
				Label(name).Tooltip(name).Cursor(ui.CursorPointer).
				Children(func() {
					ui.Icon(c, glyph).TextColor(k.TextMuted).Size(u*4, u*4)
				})
			if eye.Clicked() {
				*opts.Revealed = !*opts.Revealed
				// Back to the text afterwards, because the reason to reveal a
				// password is to check what is in it.
				in.Focus()
				c.Invalidate()
			}
		}
		return in
	})
}

// SearchInputOptions configure a SearchInput.
type SearchInputOptions struct {
	// Label names the field; it is required.
	Label string
	// Placeholder shows while the query is empty. Empty takes the library's
	// own "Search", which is a word every language here has.
	Placeholder string
	// Error marks the query as not valid, for a search that is refused.
	Error string
	// Disabled takes the field out of the tab order.
	Disabled bool
	// Width is the field's own width; zero fills the row it is in.
	Width float32
	// AutoFocus gives the field the keyboard in the frame it appears, for a
	// palette or a sheet whose whole purpose is the query.
	AutoFocus bool
}

// SearchInputResult carries a SearchInput and what happened to it.
type SearchInputResult struct {
	// Element is the field: the glass, the text, and the clear button.
	Element                     *ui.Element
	changed, submitted, cleared bool
}

// Changed reports that the query changed this frame, typed or cleared.
func (r SearchInputResult) Changed() bool { return r.changed }

// Submitted reports Enter in the field.
func (r SearchInputResult) Submitted() bool { return r.submitted }

// Cleared reports the clear button, or Escape in a field holding a query.
func (r SearchInputResult) Cleared() bool { return r.cleared }

// SearchInput is a field for a query, with what a search field is: a glass,
// a way to empty it, and Escape to do it from the keyboard.
//
// It is SearchField with the rest of a field's life — a label, an error, a
// disabled state, and an answer to whether the query changed, was submitted
// or was cleared. A list that filters as it is typed needs the first; one
// that fetches on Enter needs the second.
func SearchInput(c *ui.Context, query *string, opts SearchInputOptions) SearchInputResult {
	if query == nil {
		panic("input: SearchInput needs a query to point at; it keeps no value of its own")
	}
	if opts.Label == "" {
		panic("input: SearchInput needs a Label, or nothing can name the field")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	placeholder := opts.Placeholder
	if placeholder == "" {
		placeholder = core.Msg(c, "input.search", "Search")
	}
	clearLabel := core.Msg(c, "input.clearSearch", "Clear search")
	s := fieldSkin{label: opts.Label, err: opts.Error, disabled: opts.Disabled, width: opts.Width}

	var r SearchInputResult
	r.Element = textWell(c, s, func(_ *ui.Element) *ui.Element {
		// The glass is drawn before the text, so a person sees what the
		// field is for before what is in it.
		ui.Icon(c, glyphSearch).TextColor(k.TextMuted).Size(u*4.25, u*4.25).Shrink(0)

		in := bareInput(c, query, opts.Label, placeholder).
			Grow(1).MinHeight(core.ControlHeight(c) - u*3)
		if opts.AutoFocus {
			in.AutoFocus()
		}
		dressInput(in, opts.Error, false)

		switch {
		case in.Changed():
			r.changed = true
		case in.Submitted():
			r.submitted = true
		}
		if strings.TrimSpace(*query) != "" {
			button := ui.Box(c).Size(u*7, u*7).Shrink(0).Role(ui.RoleButton).
				Label(clearLabel).Tooltip(clearLabel).Cursor(ui.CursorPointer).
				Children(func() {
					ui.Icon(c, glyphCross).TextColor(k.TextMuted).Size(u*3.5, u*3.5)
				})
			// Escape empties it from the keyboard, which is how anybody
			// clearing a search expects to do it.
			if button.Clicked() || in.Shortcut(0, ui.KeyEscape) {
				*query = ""
				r.changed, r.cleared = true, true
				in.Focus()
				// The editor still holds what was there until the next
				// frame reads it back from the pointer, so ask for that
				// frame: the clear must not wait on somebody typing.
				c.Invalidate()
			}
		}
		return in
	})
	return r
}

// ── masked ─────────────────────────────────────────────────────────────────

// maskKey is where a masked field keeps the digits it has been given, without
// the punctuation its mask put between them.
type maskKey struct{}

// MaskedInputOptions configure a MaskedInput.
type MaskedInputOptions struct {
	// Label names the field; it is required.
	Label string
	// Mask is the shape the value takes, and it is required: "0" is a
	// digit, "*" is a letter or a digit, and everything else is a
	// separator the field prints itself. "(000) 000-0000" is a phone,
	// "00/00/0000" a date, "0000 0000 0000 0000" a card number.
	Mask string
	// Placeholder shows while the value is empty. Empty takes the mask
	// with its places blanked, which is a better prompt than "value": a
	// shape says what is wanted.
	Placeholder string
	// Error marks the value as not valid.
	Error string
	// Disabled takes the field out of the tab order.
	Disabled bool
	// Width is the field's own width; zero fills the row it is in.
	Width float32
}

// MaskedInput types a value of a known shape — a phone number, a date, a
// card number — into the places of its mask, so the separators are the
// field's business and never something a caller has to normalise after.
//
// The caller's value is the masked text as it is shown, punctuation and all,
// because that is the only form in which it round-trips: read it back and
// the digits are where they belong.
func MaskedInput(c *ui.Context, value *string, opts MaskedInputOptions) *ui.Element {
	if value == nil {
		panic("input: MaskedInput needs a value to point at; it keeps no value of its own")
	}
	if opts.Label == "" {
		panic("input: MaskedInput needs a Label, or nothing can name the field")
	}
	if opts.Mask == "" {
		panic("input: MaskedInput needs a Mask, or it is a TextInput with extra steps")
	}
	u := core.Density(c).Unit()
	mask := opts.Mask
	s := fieldSkin{label: opts.Label, err: opts.Error, disabled: opts.Disabled, width: opts.Width}

	return textWell(c, s, func(w *ui.Element) *ui.Element {
		digits := ui.Local(w, maskKey{}, func() string { return maskDigits(mask, *value) })
		in := bareInput(c, value, opts.Label, blankMask(mask, opts.Placeholder)).
			Grow(1).MinHeight(core.ControlHeight(c) - u*3)
		if in.Changed() {
			// The buffer holds what was typed; what the mask will take out
			// of it is the value. Taking fewer characters than the field
			// already had means something was deleted, and the digits
			// after the caret come along to close the gap — which is what
			// every masked field does, because there is nowhere else for
			// them to go.
			masked := fillMask(mask, maskDigits(mask, *value))
			if masked != *value {
				*value = masked
				*digits = maskDigits(mask, masked)
				// The editor reads the pointer back at the top of the next
				// frame; ask for that frame, so what the mask printed is
				// what is on screen and not one keystroke behind.
				c.Invalidate()
			}
		}
		return dressInput(in, opts.Error, false)
	})
}

// maskDigits returns the characters of s the mask has places for, which is
// the value behind however the caller happened to write it.
func maskDigits(mask, s string) string {
	room := maskPlaces(mask)
	var out strings.Builder
	for _, r := range s {
		if out.Len() >= room {
			break
		}
		switch {
		case r >= '0' && r <= '9', isLetter(r):
			out.WriteRune(upper(r))
		}
	}
	return out.String()
}

// fillMask prints digits into the places of mask and stops at the last one
// filled: a half-typed phone number is "555 123", never "555 123) " with a
// bracket hanging off the end of it.
func fillMask(mask, digits string) string {
	var out strings.Builder
	at := 0
	for _, r := range mask {
		if at >= len(digits) {
			break
		}
		switch r {
		case '0', '*':
			out.WriteByte(digits[at])
			at++
		default:
			out.WriteRune(r)
		}
	}
	return out.String()
}

// blankMask prints a mask as a prompt with its places emptied, which says
// more about the field than the word "value" ever would.
func blankMask(mask, given string) string {
	if given != "" {
		return given
	}
	return strings.Map(func(r rune) rune {
		if r == '0' || r == '*' {
			return '0'
		}
		return r
	}, mask)
}

func maskPlaces(mask string) int {
	n := 0
	for _, r := range mask {
		if r == '0' || r == '*' {
			n++
		}
	}
	return n
}

func isLetter(r rune) bool {
	return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z'
}

func upper(r rune) rune {
	if r >= 'a' && r <= 'z' {
		return r - 'a' + 'A'
	}
	return r
}

// ── one-time code ──────────────────────────────────────────────────────────

// pinKey and pinWasKey are where a one-time code keeps its digits and what
// they were last frame — two stores, because a field cannot tell which digit
// was just typed from which one it already had without the other.
type (
	pinKey    struct{}
	pinWasKey struct{}
)

// PinInputOptions configure a PinInput.
type PinInputOptions struct {
	// Label names the code as a whole; it is required. Each box is named
	// after it — "Code, digit 3" — because a reader lands in one box at a
	// time, and "edit text" six times says nothing at all.
	Label string
	// Length is how many digits the code has; six is the usual answer.
	Length int
	// Error marks the code as not valid, which is the state a code sits in
	// while somebody is typing the wrong one.
	Error string
	// Masked hides the digits, as a code read over somebody's shoulder
	// should be.
	Masked bool
	// Disabled takes the code out of the tab order.
	Disabled bool
	// Box is how wide one box is; zero is a little wider than it is tall,
	// which is what makes a row of them read as one code and not six
	// fields.
	Box float32
	// AutoFocus gives the first box the keyboard in the frame it appears,
	// which is what a sheet asking for a code should do.
	AutoFocus bool
}

// PinInputResult carries a PinInput and how much of the code is in it.
type PinInputResult struct {
	// Element is the row of boxes.
	Element *ui.Element
	filled  int
	length  int
}

// Filled reports how many digits have been typed.
func (r PinInputResult) Filled() int { return r.filled }

// Complete reports that every box holds a digit, which is the one thing a
// caller asking for a code needs to hear.
func (r PinInputResult) Complete() bool { return r.length > 0 && r.filled == r.length }

// PinInput is a one-time code: one box per digit, in a row.
//
// Typing a digit moves the focus on by itself and emptying a box moves it
// back, so the code is typed with one hand on the keyboard and never with
// the pointer. The value is the caller's string, digits and no more — which
// is what a server checks it against.
func PinInput(c *ui.Context, value *string, opts PinInputOptions) PinInputResult {
	if value == nil {
		panic("input: PinInput needs a value to point at; it keeps no value of its own")
	}
	if opts.Label == "" {
		panic("input: PinInput needs a Label, or nothing can name the code")
	}
	u := core.Density(c).Unit()

	n := opts.Length
	if n < 1 {
		n = 6
	}
	box := opts.Box
	if box <= 0 {
		box = core.ControlHeight(c) + u
	}
	seed := maskDigits(strings.Repeat("0", n), *value)
	digit := core.Msg(c, "input.digit", "digit")
	base := fieldSkin{err: opts.Error, disabled: opts.Disabled}

	var r PinInputResult
	r.length = n
	var cells []*ui.Element

	row := ui.Row(c).AlignItems(ui.Center).Gap(u * 1.5).Role(ui.RoleGroup).Label(opts.Label)
	row.Children(func() {
		// dig and prev share the slices the frame's element state keeps,
		// so a digit written here is the digit the next frame reads.
		dig := *ui.Local(row, pinKey{}, func() []string { return splitDigits(seed, n) })
		prev := *ui.Local(row, pinWasKey{}, func() []string { return make([]string, n) })
		// The focus moves on after the row is built: the next box is not in
		// hand yet while this one is being filled.
		advance := -1

		for i := range n {
			at := i
			s := base
			s.label = opts.Label + ", " + digit + " " + itoa(at+1)
			s.width = box

			var cell *ui.Element
			textWell(c, s, func(_ *ui.Element) *ui.Element {
				cell = bareInput(c, &dig[at], s.label, "").
					FillWidth().MinHeight(core.ControlHeight(c) - u*3).
					TextAlign(ui.Center)
				if opts.Masked {
					cell.Password()
				}
				if opts.AutoFocus && at == 0 {
					cell.AutoFocus()
				}
				return cell
			})
			cells = append(cells, cell)

			// Two digits in one box means a second key arrived before the
			// focus moved on; the newest is the one being typed, and the
			// older one is the keystroke the reader did not see.
			if len(dig[at]) > 1 {
				dig[at] = dig[at][len(dig[at])-1:]
			}
			switch {
			case at < n-1 && cell.Changed() && dig[at] != "":
				advance = at + 1
			case at > 0 && at < n-1 && dig[at] == "" && prev[at] != "":
				// Backspace emptied this box, so the one before it is where
				// the caret was before, and where it belongs now.
				cells[at-1].Focus()
			}
		}
		if advance > 0 {
			cells[advance].Focus()
		}
		copy(prev, dig)
		*value = joinDigits(dig)
		r.filled = countFilled(dig)
	})
	r.Element = row
	return r
}

func splitDigits(seed string, n int) []string {
	out := make([]string, n)
	for i := range out {
		if i < len(seed) {
			out[i] = seed[i : i+1]
		}
	}
	return out
}

func joinDigits(digits []string) string {
	var out strings.Builder
	for _, d := range digits {
		out.WriteString(d)
	}
	return out.String()
}

func countFilled(digits []string) int {
	n := 0
	for _, d := range digits {
		if d != "" {
			n++
		}
	}
	return n
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [8]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
