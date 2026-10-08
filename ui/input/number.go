package input

import (
	"strconv"
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
)

// The numeric controls: a number with a stepper beside it, and money written
// the way money is written.
//
// Both edit the float the caller owns, and both go through the same code, so
// the two rules that make a number field safe are true of either: what is
// written is clamped to the bounds the caller gave, and an empty field
// writes nothing at all rather than a zero nobody typed.

// numberKey, numberWasKey and numberLastKey are where a numeric field keeps
// the text being typed, what it held last frame, and the value it last wrote.
// The last one is what separates "the app changed the value" from "the user
// typed it", which a field cannot otherwise tell apart.
type (
	numberKey     struct{}
	numberWasKey  struct{}
	numberLastKey struct{}
)

// NumberInputOptions configure a NumberInput.
type NumberInputOptions struct {
	// Label names the field; it is required.
	Label string
	// Placeholder shows while the field is empty.
	Placeholder string
	// Error marks the value as not valid.
	Error string
	// Min and Max bound the value, and they are pointers because zero is a
	// bound people want: a count may not go below a real zero, while a
	// balance may go below it and a whole may not. Nil is no bound at that
	// end.
	Min, Max *float64
	// Step is how far the steppers and the arrow keys move the value, and
	// it says how many decimals the field shows: a step of 1 is a whole
	// number, 0.25 is two places, 0.01 is money's two. One is the default.
	Step float64
	// Prefix sits before the number and Suffix after it — "%", "hours",
	// "min". Neither is editable text, so neither is ever parsed back.
	Prefix, Suffix string
	// Disabled greys the field and takes its steppers out of play.
	Disabled bool
	// ReadOnly keeps the value visible and selectable but not editable.
	ReadOnly bool
	// Width is the field's own width; zero fills the row it is in.
	Width float32
	// NoSteppers leaves off the pair of buttons that move the value, for a
	// field in a column of figures where the arrow keys are enough. They
	// are drawn otherwise, because a stepper is the fastest way to move a
	// number and nobody has to learn it.
	NoSteppers bool
}

// NumberInputResult carries a NumberInput and what its value did.
type NumberInputResult struct {
	// Element is the whole field, steppers included.
	Element                 *ui.Element
	empty, changed, stepped bool
}

// Empty reports that the field holds no number at all — not that it holds
// zero. A field the user has emptied says so, and the value it points at is
// whatever it last held: nothing was written, because there was nothing to
// write. A caller that must tell "not filled in" from "filled in with 0"
// asks this, and never reads the value to find out.
func (r NumberInputResult) Empty() bool { return r.empty }

// Changed reports that the value this field points at moved this frame,
// whether by typing or by a stepper.
func (r NumberInputResult) Changed() bool { return r.changed }

// Stepped reports that the steppers or the arrow keys moved it, which is
// different from somebody typing the number themselves.
func (r NumberInputResult) Stepped() bool { return r.stepped }

// NumberInput is a number the user types or steps, bounded by the caller.
//
// What is typed is written to the caller's pointer as soon as it is a
// number, clamped to Min and Max: a field that waits for a blur to refuse a
// value lets a form submit a number it has already decided is out of range.
// An empty field writes nothing, so a cleared field is "no value" rather
// than a zero the app then saves.
func NumberInput(c *ui.Context, value *float64, opts NumberInputOptions) NumberInputResult {
	if value == nil {
		panic("input: NumberInput needs a value to point at; it keeps no value of its own")
	}
	if opts.Label == "" {
		panic("input: NumberInput needs a Label, or nothing can name the field")
	}
	if opts.Min != nil && opts.Max != nil && *opts.Min > *opts.Max {
		panic("input: NumberInput has a Min above its Max, which bounds nothing")
	}
	n := numericOptions{
		skin: fieldSkin{
			label: opts.Label, err: opts.Error,
			disabled: opts.Disabled, readOnly: opts.ReadOnly, width: opts.Width,
		},
		min: opts.Min, max: opts.Max, step: opts.Step,
		prefix: opts.Prefix, suffix: opts.Suffix,
		steppers: !opts.NoSteppers, placeholder: opts.Placeholder,
		show: plainNumber(stepDecimals(opts.Step)),
		read: readPlainNumber,
	}
	elem, out := numeric(c, value, n)
	return NumberInputResult{Element: elem, empty: out.empty, changed: out.changed, stepped: out.stepped}
}

// CurrencyInputOptions configure a CurrencyInput.
type CurrencyInputOptions struct {
	// Label names the field; it is required.
	Label string
	// Placeholder shows while the field is empty. Empty draws the symbol
	// and a zero, because a money field that reads "0.00" is showing what
	// is in it rather than hiding it behind a hint.
	Placeholder string
	// Error marks the amount as not valid.
	Error string
	// Symbol is what the amount is in. It is a symbol and not a code
	// because this interface has one currency at a time, and "$" is what
	// the amounts in it are written with. Empty prints none.
	Symbol string
	// Min and Max bound the amount, as on a NumberInput; nil is no bound
	// at that end, so an amount may be a credit unless the caller says so.
	Min, Max *float64
	// Step is how far the steppers move the amount, and how many decimals
	// the field shows. A cent is 0.01 and is the default, which is what
	// makes the field print two places without being told. It is a float64
	// and not a float32 because 0.01 in one of those is 0.009999999776, and
	// a field that counts seventeen decimal places of a cent is a field
	// printing nonsense.
	Step float64
	// Disabled and ReadOnly do what they do on a NumberInput.
	Disabled bool
	ReadOnly bool
	// Width is the field's own width; zero fills the row it is in.
	Width float32
}

// CurrencyInputResult carries a CurrencyInput and what its amount did.
type CurrencyInputResult struct {
	// Element is the whole field, steppers included.
	Element                 *ui.Element
	empty, changed, stepped bool
}

// Empty reports that the amount has not been filled in. It is Empty and not
// a zero: an invoice line the user has not priced yet is not an invoice
// line worth nothing.
func (r CurrencyInputResult) Empty() bool { return r.empty }

// Changed reports that the amount moved this frame.
func (r CurrencyInputResult) Changed() bool { return r.changed }

// Stepped reports that a stepper or the arrow keys moved it.
func (r CurrencyInputResult) Stepped() bool { return r.stepped }

// CurrencyInput is an amount, written the way amounts are written: grouped
// thousands, the currency's symbol in front, and the currency's number of
// decimal places whether or not the person typing remembered them.
//
// What the field prints and what the pointer holds are one number: parsing
// drops the symbol, the separators and the grouping, and rounds to the
// field's places, so "$1,234.50" and "1234.5" are the same amount and not
// two strings to reconcile afterwards. An empty field writes nothing.
func CurrencyInput(c *ui.Context, amount *float64, opts CurrencyInputOptions) CurrencyInputResult {
	if amount == nil {
		panic("input: CurrencyInput needs an amount to point at; it keeps no value of its own")
	}
	if opts.Label == "" {
		panic("input: CurrencyInput needs a Label, or nothing can name the field")
	}
	if opts.Min != nil && opts.Max != nil && *opts.Min > *opts.Max {
		panic("input: CurrencyInput has a Min above its Max, which bounds nothing")
	}
	step := opts.Step
	if step <= 0 {
		step = 0.01
	}
	symbol := opts.Symbol
	places := stepDecimals(step)
	n := numericOptions{
		skin: fieldSkin{
			label: opts.Label, err: opts.Error,
			disabled: opts.Disabled, readOnly: opts.ReadOnly, width: opts.Width,
		},
		min: opts.Min, max: opts.Max, step: step,
		prefix: symbol, steppers: true, placeholder: opts.Placeholder,
		show: money(symbol, places),
		read: readMoney(places),
	}
	elem, out := numeric(c, amount, n)
	return CurrencyInputResult{Element: elem, empty: out.empty, changed: out.changed, stepped: out.stepped}
}

// ── the field both numeric controls are ────────────────────────────────────

// numericOptions is what NumberInput and CurrencyInput share. They differ in
// two functions and nothing else: how a value is written, and how what was
// typed is read back.
type numericOptions struct {
	skin        fieldSkin
	min, max    *float64
	step        float64
	prefix      string
	suffix      string
	placeholder string
	steppers    bool
	show        func(float64) string
	read        func(string) (float64, bool)
}

// numericOutcome is what a numeric field did with its value this frame.
type numericOutcome struct {
	empty, changed, stepped bool
}

func numeric(c *ui.Context, value *float64, o numericOptions) (*ui.Element, numericOutcome) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	step := o.step
	if step <= 0 {
		step = 1
	}
	places := stepDecimals(step)
	var out numericOutcome

	elem := textWell(c, o.skin, func(w *ui.Element) *ui.Element {
		text := ui.Local(w, numberKey{}, func() string { return o.show(*value) })
		wrote := ui.Local(w, numberLastKey{}, func() float64 { return *value })
		wasEmpty := ui.Local(w, numberWasKey{}, func() bool { return false })

		if o.prefix != "" {
			fieldText(c, o.prefix, k.TextMuted)
		}
		in := bareInput(c, text, o.skin.label, o.placeholder).
			Grow(1).MinHeight(core.ControlHeight(c) - u*3)
		if o.min != nil || o.max != nil {
			// Assistive technology needs the bounds and the value to say
			// what the field is at, and they are the one thing the app knows
			// and the field does not.
			lo, hi := 0.0, 0.0
			if o.min != nil {
				lo = *o.min
			}
			if o.max != nil {
				hi = *o.max
			}
			in.Range(lo, hi, *value)
		}

		// write is the only way the field's value moves, and everything
		// that moves it goes through the clamp: a number out of range is
		// pulled back to the bound rather than refused, so what the caller
		// reads is always a number it can use.
		write := func(v float64) {
			v = roundTo(v, places)
			v = clampTo(v, o.min, o.max)
			if v != *value {
				*value = v
				out.changed = true
			}
			*wrote = v
		}
		// stepBy moves the value one step in dir, which is 1 or -1: what it
		// moves by is the step, and a stepper and an arrow key move by the
		// same amount.
		stepBy := func(dir float64) {
			from := *value
			if *wasEmpty {
				// An empty field has no value to move, so the first press
				// sets the bound it moves away from: one more than the
				// least of something is what "one more" means.
				switch {
				case dir > 0 && o.min != nil:
					from = *o.min
				case dir < 0 && o.max != nil:
					from = *o.max
				default:
					from = 0
				}
			}
			write(from + dir*step)
			*text = o.show(*value)
			*wasEmpty = false
			out.stepped = true
			// The editor reads the pointer back at the top of the next
			// frame; ask for that frame, so a stepper pressed while the
			// field has the focus shows its number at once.
			c.Invalidate()
		}

		if in.Changed() {
			// What was typed becomes the value the moment it is a number.
			// What it is not — half a sign, an empty field, letters — is
			// not a value yet, and leaves the pointer alone: a field
			// emptied with Backspace must not write a zero.
			if v, ok := o.read(strings.TrimSpace(*text)); ok {
				write(v)
			}
		}
		out.empty = strings.TrimSpace(*text) == ""
		*wasEmpty = out.empty

		switch {
		case in.Changed():
			// What the person typed stays on screen as they typed it; a
			// field that reformatted under the caret would move it out
			// from under them.
		case *value != *wrote:
			// The app changed the value, so the field follows it.
			*text = o.show(*value)
		case !in.Focused() && !*wasEmpty:
			// The focus left a filled field: show it in full, with the
			// separators and the places the field owes.
			*text = o.show(*value)
		}
		dressInput(in, o.skin.err, o.skin.readOnly)

		if in.Shortcut(0, ui.KeyUp) {
			stepBy(1)
		}
		if in.Shortcut(0, ui.KeyDown) {
			stepBy(-1)
		}
		if o.suffix != "" {
			fieldText(c, o.suffix, k.TextMuted)
		}
		if o.steppers {
			less, more := steppers(c, value, o.min, o.max)
			if less.Clicked() {
				stepBy(-1)
			}
			if more.Clicked() {
				stepBy(1)
			}
		}
		return in
	})
	return elem, out
}

// steppers are the two buttons at the trailing edge of a numeric field. They
// are not tab stops: the arrow keys inside the field do the same two things,
// and a field that took two stops of its own would make Tab a worse way
// through a form than the arrows are.
func steppers(c *ui.Context, value *float64, min, max *float64) (less, more *ui.Element) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	side := u * 6.5
	button := func(glyph *ui.SVG, name string, bound *float64, off bool) *ui.Element {
		return ui.Box(c).Size(side, side).Shrink(0).Role(ui.RoleButton).
			Label(name).Tooltip(name).Cursor(ui.CursorPointer).
			// A stepper at its bound is not merely quiet: pressing it must
			// do nothing at all, or a number can be walked off the end of
			// its own range one press at a time.
			Disabled(bound != nil && ((off && *value <= *bound) || (!off && *value >= *bound))).
			Children(func() {
				ui.Icon(c, glyph).TextColor(k.TextMuted).Size(u*3.5, u*3.5)
			})
	}
	return button(glyphLess, core.Msg(c, "input.decrease", "Decrease"), min, true),
		button(glyphMore, core.Msg(c, "input.increase", "Increase"), max, false)
}

func clampTo(v float64, min, max *float64) float64 {
	if min != nil && v < *min {
		v = *min
	}
	if max != nil && v > *max {
		v = *max
	}
	return v
}

// roundTo pulls a value back to the places a step says the field shows, the
// way a browser's number inputs do: a step of 0.25 cannot leave a field
// showing three places. It goes through the formatted string rather than
// multiplying by a power of ten, so a large amount does not overflow on its
// way to two decimals.
func roundTo(v float64, places int) float64 {
	if places <= 0 || places > 15 {
		return v
	}
	r, err := strconv.ParseFloat(strconv.FormatFloat(v, 'f', places, 64), 64)
	if err != nil {
		return v
	}
	return r
}

// ── writing and reading numbers ────────────────────────────────────────────

// stepDecimals is how many decimal places a step has, which is how a
// numeric field knows what to print: a step of 0.25 is two places, a step of
// 1 is none.
//
// It stops at eight, because a step with more places than that is a number
// nobody chose: it is what a step becomes after it has been through a float
// that could not hold it, and a field printing its places would be printing
// the accident.
func stepDecimals(step float64) int {
	if step <= 0 {
		return 0
	}
	s := strconv.FormatFloat(step, 'f', -1, 64)
	i := strings.IndexByte(s, '.')
	if i < 0 {
		return 0
	}
	return min(len(s)-i-1, 8)
}

// plainNumber writes a number the way the interface writes numbers: no
// separators, no symbol, and always the places its step says.
func plainNumber(places int) func(float64) string {
	return func(v float64) string {
		return strconv.FormatFloat(roundTo(v, places), 'f', places, 64)
	}
}

func readPlainNumber(s string) (float64, bool) {
	if s == "" {
		return 0, false
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

// money writes an amount with its symbol in front and its thousands grouped,
// which is the only way an amount is written anywhere in this interface.
func money(symbol string, places int) func(float64) string {
	return func(v float64) string {
		s := strconv.FormatFloat(roundTo(v, places), 'f', places, 64)
		if v < 0 {
			// The sign leads, as it does on a bill: "-$5.00", never
			// "$-5.00", which reads as a dash before a currency.
			s = "-" + symbol + groupDigits(s[1:])
		} else {
			s = symbol + groupDigits(s)
		}
		return s
	}
}

// readMoney is the other half of money: it reads back what money writes, and
// what a person types instead — with or without the symbol, with commas or
// without, with a trailing decimal point while they are mid-amount.
func readMoney(places int) func(string) (float64, bool) {
	return func(s string) (float64, bool) {
		s = strings.Map(func(r rune) rune {
			switch {
			case r >= '0' && r <= '9', r == '.', r == '-':
				return r
			}
			return -1
		}, s)
		if s == "" || s == "-" || s == "." {
			return 0, false
		}
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, false
		}
		return roundTo(v, places), true
	}
}

// groupDigits puts a comma every third digit of the whole part and leaves
// the decimals alone, which is what "1,234.50" is and what "1,234.567" is
// not. It lives here rather than in internal because the money there counts
// whole numbers, and an amount is a whole part and a fraction that have to
// stay apart.
func groupDigits(s string) string {
	dot := strings.IndexByte(s, '.')
	whole, frac := s, ""
	if dot >= 0 {
		whole, frac = s[:dot], s[dot:]
	}
	if len(whole) <= 3 {
		return whole + frac
	}
	var out []byte
	// Counted from the front, which puts a comma every third digit from the
	// end — the only end a comma may be counted from.
	for i := range len(whole) {
		if i > 0 && (len(whole)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, whole[i])
	}
	return string(out) + frac
}
