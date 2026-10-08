package input

import (
	"image"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// What the tests here assert is what a person would see and what the app
// would read: the words on screen, the boxes they occupy, the pixels they
// paint, and — above all — the value each control points at.
//
// They do not assert the text drawn after a press. A click lands in the
// frame after it is sent, and MyGo builds a frame up to three times so the
// outcome is on screen by the end of it, which means a Result's Changed and
// Stepped read false again by the time the tester settles. "The field now
// shows 42" is a claim about a frame nobody kept; "the value it points at is
// 42" is a claim about the app's state, which is what a control is for. The
// flags are therefore collected across the frames of a settle, the way a
// caller collects them in a real view.

// ── helpers ────────────────────────────────────────────────────────────────

// fieldTester runs a view in a window of the given size with the library's
// light palette, and returns the tester. Every test starts here, so none of
// them depends on another having set a preference first.
func fieldTester(w, h int, view func(c *ui.Context)) *ui.Tester {
	return ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		// A painted background, so a pixel that is not ink is a known colour
		// rather than whatever the surface happened to be.
		ui.Box(c).Fill().Background(core.Tokens(c).Background).
			Children(func() { view(c) })
	}, w, h)
}

// darkTester is fieldTester for a window following the desktop into dark. The
// mode is core.System, so that the tester's own SetDark is what decides: that
// is how a real window behaves, and it is the only way to catch a control
// that hard-codes one appearance.
func darkTester(w, h int, view func(c *ui.Context)) *ui.Tester {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: core.System})
		ui.Box(c).Fill().Background(core.Tokens(c).Background).
			Children(func() { view(c) })
	}, w, h)
	tt.SetDark(true)
	return tt
}

// boxOf is the box of the element showing or labelled s. An empty ui.Box
// measures zero wide, so anything that has to be measured has to be found by
// a name.
func boxOf(t *testing.T, tt *ui.Tester, s string) ui.Rect {
	t.Helper()
	r, ok := tt.Find(s)
	if !ok {
		t.Fatalf("nothing showing or labelled %q; the frame has %q", s, tt.Texts())
	}
	return r
}

// fieldPixel is the colour of one pixel of a frame, in DIPs. It is named for
// what it reads rather than what it is called after the picture under it,
// because another test file in this package reads pixels too.
func fieldPixel(img *image.RGBA, x, y float32) ui.Color {
	px, py := int(x), int(y)
	if !(image.Point{X: px, Y: py}).In(img.Bounds()) {
		return ui.Color{}
	}
	i := img.PixOffset(px, py)
	return ui.Color{R: img.Pix[i], G: img.Pix[i+1], B: img.Pix[i+2], A: img.Pix[i+3]}
}

// samePixel compares two colours as they are painted: a solid fill comes
// back exactly as it was set, so an exact comparison is the strongest
// assertion available, and anything looser would hide a control drawing the
// wrong token.
func samePixel(got, want ui.Color) bool {
	return got.R == want.R && got.G == want.G && got.B == want.B && got.A == want.A
}

// nearPixel compares two colours within a small per-channel tolerance. It is
// for hairlines: a one-dip outline across a fractional x blends with what it
// stands on, so no single pixel of it is exactly the token — the worst case
// in these palettes is about a twentieth of the way to the background. A
// tolerance this small still tells one token from another — the palette's
// nearest neighbours are dozens of steps apart — without betting on where the
// outline landed.
func nearPixel(got, want ui.Color) bool {
	near := func(a, b uint8) bool {
		d := int(a) - int(b)
		if d < 0 {
			d = -d
		}
		return d <= 12
	}
	return near(got.R, want.R) && near(got.G, want.G) && near(got.B, want.B) && got.A == want.A
}

// wantPixel fails unless the pixel at (x, y) is exactly want.
func wantPixel(t *testing.T, tt *ui.Tester, x, y float32, want ui.Color, what string) {
	t.Helper()
	if got := fieldPixel(tt.Image(), x, y); !samePixel(got, want) {
		t.Errorf("%s: pixel (%v, %v) is %v, want %v", what, x, y, got, want)
	}
}

// fieldEdge is where the field named label has an edge of its own, and the row
// to look at it on: the first pixel across the middle of the field that is not
// the window's background.
//
// It is looked for rather than counted out of a padding, because the well
// around a field has no name of its own and the distance from a control to its
// own outline changes with what a control puts inside itself — a money field
// starts further in than a plain one. It is the rule §15.2 of the design
// system asks for, taken a step further: assert what was drawn, not what an
// element's numbers say.
func fieldEdge(t *testing.T, tt *ui.Tester, label string, bg ui.Color) (x, y float32) {
	t.Helper()
	r := boxOf(t, tt, label)
	y = r.Y + r.H/2
	for x := float32(0); x < r.X; x++ {
		if !samePixel(fieldPixel(tt.Image(), x+0.5, y), bg) {
			return x + 0.5, y
		}
	}
	t.Fatalf("the field %q has no edge of its own at y=%.1f", label, y)
	return 0, y
}

// wantHairline fails unless the field named label outlines itself in want.
//
// It is a pixel assertion because the outline is the only thing that says a
// field is wrong: the message under it says what is wrong, and only a colour
// at the edge says that this field is the one at fault.
//
// The outline is one dip wide and text metrics move it onto fractional x
// values, so its ink blends across two physical pixels and neither is exactly
// the token. Either of the two carrying it — within nearPixel's tolerance —
// is the field outlining itself in want; a wrong token is dozens of steps
// away and still fails.
func wantHairline(t *testing.T, tt *ui.Tester, label string, k theme.Tokens, want ui.Color) {
	t.Helper()
	x, y := fieldEdge(t, tt, label, k.Background)
	for dx := float32(0); dx < 2; dx++ {
		if got := fieldPixel(tt.Image(), x+dx, y); nearPixel(got, want) {
			return
		}
	}
	got := fieldPixel(tt.Image(), x, y)
	t.Errorf("the outline of %s: pixel (%v, %v) is %v, want %v", label, x, y, got, want)
}

// wantFill fails unless the inside of the field named label is the surface of
// the palette it is being drawn in. A control that painted its own background
// out of a palette of its own would pass every other test here and fail this
// one in the dark.
func wantFill(t *testing.T, tt *ui.Tester, label string, k theme.Tokens, want ui.Color) {
	t.Helper()
	x, y := fieldEdge(t, tt, label, k.Background)
	wantPixel(t, tt, x+3, y, want, "the inside of "+label)
}

// typingInto clicks the field named s and types text into it, which is what
// a person does: a field nobody pressed does not take a keystroke.
func typingInto(t *testing.T, tt *ui.Tester, s, text string) {
	t.Helper()
	if err := tt.Click(s); err != nil {
		t.Fatal(err)
	}
	tt.Type(text)
}

// typingDigits types a one-time code the way a keyboard sends one: a key at
// a time, so that each digit lands in the box the focus moved to. A paste
// would arrive as a single insert and the field would have to guess.
func typingDigits(t *testing.T, tt *ui.Tester, s, digits string) {
	t.Helper()
	if err := tt.Click(s); err != nil {
		t.Fatal(err)
	}
	for _, d := range digits {
		tt.Type(string(d))
	}
}

// emptying clicks the field, selects everything in it and presses Backspace,
// which is the only way to take the last character out of a field that has
// one.
func emptying(t *testing.T, tt *ui.Tester, s string) {
	t.Helper()
	if err := tt.Click(s); err != nil {
		t.Fatal(err)
	}
	tt.Command("selectAll")
	tt.Key(0, ui.KeyBackspace)
}

// bound is the address of a bound, for the Options that take one: a field's
// ends are pointers so that a real zero can be a bound, which a plain number
// could not say.
func bound(v float64) *float64 { return &v }

func countOccurrences(texts []string, want string) int {
	n := 0
	for _, s := range texts {
		if strings.Contains(s, want) {
			n++
		}
	}
	return n
}

// ── TextInput ──────────────────────────────────────────────────────────────

func TestTextInputTypesIntoTheStringItPointsAt(t *testing.T) {
	name := ""
	tt := fieldTester(440, 140, func(c *ui.Context) {
		TextInput(c, &name, TextInputOptions{
			Label: "Customer", Placeholder: "Who is waiting",
		})
	})
	// What the field holds is what the app reads: the tester's names are
	// what the window says, not what a field has inside it, and the value
	// is the thing a caller would go on to save.
	typingInto(t, tt, "Customer", "Riverside Clinic")
	if name != "Riverside Clinic" {
		t.Errorf("value = %q, want %q", name, "Riverside Clinic")
	}
	if !tt.HasText("Customer") {
		t.Errorf("the field should be named for what it is: %q", tt.Texts())
	}
}

func TestTextInputInsistsOnAName(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("an unnamed field should panic: nothing can name it")
		}
	}()
	fieldTester(320, 120, func(c *ui.Context) {
		TextInput(c, new(string), TextInputOptions{})
	})
}

func TestTextInputRefusesNoValueToPointAt(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("a field with nothing to point at should panic")
		}
	}()
	fieldTester(320, 120, func(c *ui.Context) {
		TextInput(c, nil, TextInputOptions{Label: "Customer"})
	})
}

func TestAFieldWithAnErrorOutlinesItselfInTheDangerColour(t *testing.T) {
	name := "Riverside"
	tt := fieldTester(440, 140, func(c *ui.Context) {
		TextInput(c, &name, TextInputOptions{
			Label: "Customer", Error: "That name is taken",
		})
	})
	wantHairline(t, tt, "Customer", theme.Light(), theme.Light().Danger)
}

func TestAFieldWithNothingWrongKeepsTheOrdinaryOutline(t *testing.T) {
	name := "Riverside"
	tt := fieldTester(440, 140, func(c *ui.Context) {
		TextInput(c, &name, TextInputOptions{Label: "Customer"})
	})
	// And it is the ordinary border, not the danger one it would be with an
	// error: the difference has to be visible, or the two states are the same
	// state with a message under it.
	wantHairline(t, tt, "Customer", theme.Light(), theme.Light().Border)
}

// ── TextArea ───────────────────────────────────────────────────────────────

func TestTextAreaShowsAsManyLinesAsItWasAsked(t *testing.T) {
	note := ""
	tt := fieldTester(480, 280, func(c *ui.Context) {
		TextArea(c, &note, TextAreaOptions{Label: "What happened", Lines: 3})
	})
	typingInto(t, tt, "What happened", "Customer called about the invoice")
	if note != "Customer called about the invoice" {
		t.Errorf("value = %q, want the note as typed", note)
	}
	// Three lines of text and no more: a text area is as tall as what it
	// shows, so the text inside it scrolls rather than the form growing down
	// the page every time somebody writes a paragraph.
	want := 3 * theme.BodySize * 1.45
	if r := boxOf(t, tt, "What happened"); r.H < want-3 || r.H > want+3 {
		t.Errorf("a three-line area is %.1f high, want about %.1f", r.H, want)
	}
}

func TestTextAreaNeedsALabel(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("an unnamed text area should panic")
		}
	}()
	fieldTester(320, 200, func(c *ui.Context) {
		TextArea(c, new(string), TextAreaOptions{Lines: 2})
	})
}

// ── PasswordInput ──────────────────────────────────────────────────────────

func TestPasswordHidesWhatItHoldsAndNamesItsKey(t *testing.T) {
	secret := "correct-horse"
	shown := false
	tt := fieldTester(460, 140, func(c *ui.Context) {
		PasswordInput(c, &secret, PasswordInputOptions{
			Label: "Callback token", Revealed: &shown,
		})
	})
	if !tt.HasText("Show password") {
		t.Fatalf("the key needs a name of its own: %q", tt.Texts())
	}
	if tt.HasText("correct-horse") {
		t.Error("a hidden field must not show what it holds")
	}
	if secret != "correct-horse" {
		t.Errorf("value = %q, want it untouched by being drawn", secret)
	}
}

func TestTheKeysWritesIntoTheSwitchTheCallerPointedAt(t *testing.T) {
	secret := "correct-horse"
	shown := false
	tt := fieldTester(460, 140, func(c *ui.Context) {
		PasswordInput(c, &secret, PasswordInputOptions{
			Label: "Callback token", Revealed: &shown,
		})
	})
	if err := tt.Click("Show password"); err != nil {
		t.Fatal(err)
	}
	if !shown {
		t.Error("the key writes into the switch the caller pointed at, not into itself")
	}
	if !tt.HasText("Hide password") {
		t.Errorf("a revealed field offers to hide itself again: %q", tt.Texts())
	}
}

func TestAPasswordWithNothingToRevealStillOffersNoKey(t *testing.T) {
	// A field with no switch to write into must not carry a key that does
	// nothing when it is pressed.
	secret := "hunter2"
	tt := fieldTester(460, 140, func(c *ui.Context) {
		PasswordInput(c, &secret, PasswordInputOptions{Label: "Callback token"})
	})
	if tt.HasText("Show password") {
		t.Errorf("there is no switch to write into: %q", tt.Texts())
	}
	if tt.HasText("hunter2") {
		t.Error("the value is not drawn as text")
	}
}

// ── SearchInput ────────────────────────────────────────────────────────────

func TestSearchReportsWhatHappenedToTheQuery(t *testing.T) {
	query := ""
	var changed, cleared bool
	tt := fieldTester(460, 140, func(c *ui.Context) {
		r := SearchInput(c, &query, SearchInputOptions{Label: "Search callbacks"})
		changed, cleared = changed || r.Changed(), cleared || r.Cleared()
	})
	typingInto(t, tt, "Search callbacks", "hillside")
	if query != "hillside" {
		t.Errorf("query = %q, want %q", query, "hillside")
	}
	if !changed {
		t.Error("typing into a search field is a change to the caller")
	}
	if !tt.HasText("Clear search") {
		t.Errorf("a field with a query offers a way to empty it: %q", tt.Texts())
	}
	if err := tt.Click("Clear search"); err != nil {
		t.Fatal(err)
	}
	if query != "" {
		t.Errorf("query = %q after clearing, want empty", query)
	}
	if !cleared || !changed {
		t.Error("clearing is both a clear and a change")
	}
}

func TestSearchSubmitsOnEnter(t *testing.T) {
	query := ""
	var submitted bool
	tt := fieldTester(460, 140, func(c *ui.Context) {
		r := SearchInput(c, &query, SearchInputOptions{Label: "Search callbacks"})
		submitted = submitted || r.Submitted()
	})
	typingInto(t, tt, "Search callbacks", "riverside")
	tt.Key(0, ui.KeyEnter)
	if !submitted {
		t.Error("Enter in a search field submits it")
	}
}

// ── MaskedInput ────────────────────────────────────────────────────────────

func TestMaskedInputPutsTheDigitsWhereTheMaskAsks(t *testing.T) {
	phone := ""
	tt := fieldTester(480, 140, func(c *ui.Context) {
		MaskedInput(c, &phone, MaskedInputOptions{
			Label: "Callback phone", Mask: "(000) 000-0000",
		})
	})
	typingInto(t, tt, "Callback phone", "5551234567")
	if phone != "(555) 123-4567" {
		t.Errorf("value = %q, want %q", phone, "(555) 123-4567")
	}
}

func TestMaskedInputStopsAtItsOwnLength(t *testing.T) {
	card := ""
	tt := fieldTester(540, 140, func(c *ui.Context) {
		MaskedInput(c, &card, MaskedInputOptions{
			Label: "Card number", Mask: "0000 0000 0000 0000",
		})
	})
	// Sixteen digits and then four more, as a paste out of a wallet sends
	// them: the field takes the card's length and no more.
	typingInto(t, tt, "Card number", "41111111111111119999")
	if card != "4111 1111 1111 1111" {
		t.Errorf("value = %q, want one card's worth of digits", card)
	}
}

func TestMaskedInputIgnoresWhateverIsNotADigit(t *testing.T) {
	date := ""
	tt := fieldTester(480, 140, func(c *ui.Context) {
		MaskedInput(c, &date, MaskedInputOptions{
			Label: "Due date", Mask: "00/00/0000",
		})
	})
	// A paste of a whole date, separators and all, comes out in the mask's
	// shape rather than doubled up with the field's own.
	typingInto(t, tt, "Due date", "12/31/2026")
	if date != "12/31/2026" {
		t.Errorf("value = %q, want %q", date, "12/31/2026")
	}
}

func TestMaskedInputNeedsAMask(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("a masked field with no mask should panic")
		}
	}()
	fieldTester(320, 120, func(c *ui.Context) {
		MaskedInput(c, new(string), MaskedInputOptions{Label: "Phone"})
	})
}

// ── PinInput ───────────────────────────────────────────────────────────────

func TestPinInputTakesTheCodeOneBoxAtATime(t *testing.T) {
	code := ""
	var got PinInputResult
	tt := fieldTester(600, 140, func(c *ui.Context) {
		got = PinInput(c, &code, PinInputOptions{Label: "Verification code", Length: 6})
	})
	// Every box is named after the code and its own place in it, so a reader
	// landing in one knows which digit of six it is.
	if !tt.HasText("Verification code, digit 1") || !tt.HasText("Verification code, digit 6") {
		t.Fatalf("each box needs its own name: %q", tt.Texts())
	}
	typingDigits(t, tt, "Verification code, digit 1", "418922")
	if code != "418922" {
		t.Errorf("code = %q, want %q", code, "418922")
	}
	if !got.Complete() || got.Filled() != 6 {
		t.Errorf("filled = %d, complete = %v; want six digits and no more", got.Filled(), got.Complete())
	}
}

func TestPinInputMovesTheFocusOnAsDigitsAreTyped(t *testing.T) {
	code := ""
	tt := fieldTester(600, 140, func(c *ui.Context) {
		PinInput(c, &code, PinInputOptions{Label: "Verification code", Length: 4})
	})
	typingDigits(t, tt, "Verification code, digit 1", "7")
	// The focus followed the keystroke, so the next digit lands in the next
	// box without the pointer being touched again.
	if !tt.Focused("Verification code, digit 2") {
		t.Errorf("the focus should have moved on; the frame has %q", tt.Texts())
	}
	if code != "7" {
		t.Errorf("code = %q, want the one digit typed", code)
	}
}

func TestPinInputTakesTheCodeTheCallerAlreadyHolds(t *testing.T) {
	code := "418922"
	var got PinInputResult
	fieldTester(600, 140, func(c *ui.Context) {
		got = PinInput(c, &code, PinInputOptions{Label: "Verification code", Length: 6})
	})
	if !got.Complete() {
		t.Errorf("a code the caller already holds is complete: filled = %d", got.Filled())
	}
	if code != "418922" {
		t.Errorf("code = %q, want the value left exactly as it was", code)
	}
}

// ── NumberInput ────────────────────────────────────────────────────────────

func TestNumberInputWritesTheNumberItIsGiven(t *testing.T) {
	copies := 1.0
	tt := fieldTester(480, 140, func(c *ui.Context) {
		NumberInput(c, &copies, NumberInputOptions{
			Label: "Copies", Min: bound(1), Max: bound(99), Step: 1,
		})
	})
	emptying(t, tt, "Copies")
	tt.Type("12")
	if copies != 12 {
		t.Errorf("value = %v, want 12", copies)
	}
}

// "NaN" and "Inf" are the two things ParseFloat accepts without complaint,
// and neither of them is a number a form can submit. The clamp cannot save
// either one: every comparison against NaN is false, so it passes both bounds
// untouched, and a field that writes it has written it into the caller's value
// for good.
func TestNumberInputRefusesWhatIsNotANumber(t *testing.T) {
	for _, what := range []string{"NaN", "Inf"} {
		t.Run(what, func(t *testing.T) {
			copies := 5.0
			tt := fieldTester(480, 140, func(c *ui.Context) {
				NumberInput(c, &copies, NumberInputOptions{
					Label: "Copies", Min: bound(1), Max: bound(99), Step: 1,
				})
			})
			emptying(t, tt, "Copies")
			tt.Type(what)
			if copies != 5 {
				t.Errorf("typing %q left the value at %v, want the 5 it had", what, copies)
			}
		})
	}
}

func TestNumberInputClampsWhatItWritesToItsBounds(t *testing.T) {
	copies := 1.0
	tt := fieldTester(480, 140, func(c *ui.Context) {
		NumberInput(c, &copies, NumberInputOptions{
			Label: "Copies", Min: bound(1), Max: bound(99), Step: 1,
		})
	})
	// Past the top: the field pulls the number back to the bound rather than
	// dropping the keystroke, because what the caller reads has to be a
	// number the form can use, and a field that says nothing about the
	// refusal leaves the person to find out later.
	emptying(t, tt, "Copies")
	tt.Type("5000")
	if copies != 99 {
		t.Errorf("value = %v after typing past the top, want the max, 99", copies)
	}

	// And below the bottom is the same rule the other way round.
	emptying(t, tt, "Copies")
	tt.Type("0")
	if copies != 1 {
		t.Errorf("value = %v after typing under the bottom, want the min, 1", copies)
	}
}

func TestAnEmptyNumberFieldIsNotZero(t *testing.T) {
	copies := 5.0
	var got NumberInputResult
	tt := fieldTester(480, 140, func(c *ui.Context) {
		got = NumberInput(c, &copies, NumberInputOptions{Label: "Copies", Step: 1})
	})
	// A field cleared with Backspace must come out as "no number", not as a
	// zero: a form that cannot tell the two saves zeroes into every record
	// somebody opened and left.
	emptying(t, tt, "Copies")

	if copies != 5 {
		t.Errorf("value = %v after clearing, want the last number (5) left alone, not 0", copies)
	}
	if !got.Empty() {
		t.Error("a field with nothing in it should say so")
	}
}

func TestANumberFieldTheAppHasNotTouchedIsNotEmpty(t *testing.T) {
	var got NumberInputResult
	copies := 5.0
	fieldTester(480, 140, func(c *ui.Context) {
		got = NumberInput(c, &copies, NumberInputOptions{Label: "Copies", Step: 1})
	})
	if got.Empty() {
		t.Error("a field showing the value it holds is not empty")
	}
	if copies != 5 {
		t.Errorf("a field nobody typed in must not touch the value: %v", copies)
	}
}

func TestTheSteppersMoveTheNumberAndStopAtItsBounds(t *testing.T) {
	copies := 1.0
	min, max := 1.0, 3.0
	var stepped bool
	tt := fieldTester(560, 140, func(c *ui.Context) {
		r := NumberInput(c, &copies, NumberInputOptions{
			Label: "Copies", Min: &min, Max: &max, Step: 1,
		})
		stepped = stepped || r.Stepped()
	})
	for _, want := range []float64{2, 3, 3} {
		if err := tt.Click("Increase"); err != nil {
			t.Fatal(err)
		}
		if copies != want {
			t.Fatalf("value = %v after a press, want %v", copies, want)
		}
	}
	// The third press found the stepper at its bound and did nothing at all,
	// which is the point: a number must not walk off the end of its range
	// one press at a time.
	if !stepped {
		t.Error("a stepper press is a step, which is not the same as somebody typing")
	}
	if err := tt.Click("Decrease"); err != nil {
		t.Fatal(err)
	}
	if copies != 2 {
		t.Errorf("value = %v after stepping down, want 2", copies)
	}
}

func TestAnEmptyNumberStepsFromItsBound(t *testing.T) {
	copies := 0.0
	min, max := 2.0, 20.0
	tt := fieldTester(560, 140, func(c *ui.Context) {
		NumberInput(c, &copies, NumberInputOptions{
			Label: "Copies", Min: &min, Max: &max, Step: 1,
		})
	})
	// Nobody has typed anything, so there is no number to step from; one
	// more than the least of something is the least of something plus one.
	emptying(t, tt, "Copies")
	if err := tt.Click("Increase"); err != nil {
		t.Fatal(err)
	}
	if copies != 3 {
		t.Errorf("value = %v after one step from empty, want 3", copies)
	}
}

func TestTheArrowKeysStepANumberToo(t *testing.T) {
	copies := 5.0
	tt := fieldTester(480, 140, func(c *ui.Context) {
		NumberInput(c, &copies, NumberInputOptions{Label: "Copies", Step: 1})
	})
	// The steppers take no stop of the tab order, so the arrows are how the
	// field is stepped from the keyboard; they have to work.
	if err := tt.Click("Copies"); err != nil {
		t.Fatal(err)
	}
	tt.Key(0, ui.KeyUp)
	if copies != 6 {
		t.Errorf("value = %v after Up, want 6", copies)
	}
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyDown)
	if copies != 4 {
		t.Errorf("value = %v after two Downs, want 4", copies)
	}
}

func TestNumberInputRefusesBoundsThatBoundNothing(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("a Min above its Max should panic")
		}
	}()
	min, max := 9.0, 1.0
	fieldTester(320, 120, func(c *ui.Context) {
		NumberInput(c, new(float64), NumberInputOptions{
			Label: "Copies", Min: &min, Max: &max,
		})
	})
}

// ── CurrencyInput ──────────────────────────────────────────────────────────

// The formatting is checked here rather than through a drawn field: what a
// field holds is not a string the window says, so the tester cannot read it
// back, and an amount is worth nothing if it is written as "1234.5".
func TestAnAmountIsWrittenTheWayAmountsAreWritten(t *testing.T) {
	show := money("$", 2)
	for _, want := range []struct {
		amount float64
		text   string
	}{
		{0, "$0.00"},
		{12.5, "$12.50"},
		{1234.5, "$1,234.50"},
		{1234567.891, "$1,234,567.89"},
		{0.005, "$0.01"},
		{-1234.5, "-$1,234.50"},
	} {
		if got := show(want.amount); got != want.text {
			t.Errorf("%v is written %q, want %q", want.amount, got, want.text)
		}
	}
	// And a whole-unit field writes no decimals at all, which is what the
	// step says rather than what the caller remembered to pass.
	if got := money("$", stepDecimals(1))(3); got != "$3" {
		t.Errorf("a field of whole units wrote %q, want %q", got, "$3")
	}
}

func TestAnAmountIsReadBackHoweverItWasTyped(t *testing.T) {
	read := readMoney(2)
	for _, want := range []struct {
		typed  string
		amount float64
		ok     bool
	}{
		{"$1,234.50", 1234.5, true},
		{"1234.5", 1234.5, true},
		{"1,234.50", 1234.5, true},
		{"-99", -99, true},
		// Half of what was typed is not an amount, and a lone sign is
		// somebody halfway through typing one.
		{"", 0, false},
		{"-", 0, false},
		{".", 0, false},
		{"twelve", 0, false},
	} {
		got, ok := read(want.typed)
		if ok != want.ok || (ok && got != want.amount) {
			t.Errorf("%q read back as %v (%v), want %v (%v)", want.typed, got, ok, want.amount, want.ok)
		}
	}
}

func TestCurrencyReadsBackEveryWayAnAmountIsTyped(t *testing.T) {
	amount := 0.0
	tt := fieldTester(540, 140, func(c *ui.Context) {
		CurrencyInput(c, &amount, CurrencyInputOptions{Label: "Open cost", Symbol: "$"})
	})
	// With the symbol and the groups, as somebody copying a figure off an
	// invoice would type it; without them; as a bare number. All three are
	// the same amount, because a field that made the caller normalise what a
	// person typed has moved the work onto everybody.
	for _, typed := range []string{"$1,234.50", "1234.5", "1234.50"} {
		emptying(t, tt, "Open cost")
		tt.Type(typed)
		if amount != 1234.5 {
			t.Errorf("typing %q gave %v, want 1234.5", typed, amount)
		}
	}
}

func TestCurrencyRoundsToItsOwnPlaces(t *testing.T) {
	amount := 0.0
	tt := fieldTester(540, 140, func(c *ui.Context) {
		CurrencyInput(c, &amount, CurrencyInputOptions{Label: "Open cost", Symbol: "$"})
	})
	// More places than money has: the field keeps the cents it can keep and
	// drops the rest, because nobody can pay 1.239 of anything.
	emptying(t, tt, "Open cost")
	tt.Type("1.239")
	if amount != 1.24 {
		t.Errorf("amount = %v, want 1.24", amount)
	}
}

func TestCurrencyClampsToItsBounds(t *testing.T) {
	amount := 0.0
	min, max := 0.0, 100.0
	tt := fieldTester(540, 140, func(c *ui.Context) {
		CurrencyInput(c, &amount, CurrencyInputOptions{
			Label: "Deposit", Symbol: "$", Min: &min, Max: &max,
		})
	})
	emptying(t, tt, "Deposit")
	tt.Type("250")
	if amount != 100 {
		t.Errorf("amount = %v after typing past the top, want the max, 100", amount)
	}
}

func TestAnEmptyAmountIsNotAnAmountOfNothing(t *testing.T) {
	amount := 12.0
	var got CurrencyInputResult
	tt := fieldTester(540, 140, func(c *ui.Context) {
		got = CurrencyInput(c, &amount, CurrencyInputOptions{Label: "Deposit", Symbol: "$"})
	})
	emptying(t, tt, "Deposit")
	if amount != 12 {
		t.Errorf("amount = %v after clearing, want the last amount left alone, not 0", amount)
	}
	if !got.Empty() {
		t.Error("an amount nobody has typed is not an amount of nothing")
	}
}

func TestCurrencyNeedsALabel(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("an unnamed amount field should panic")
		}
	}()
	fieldTester(320, 120, func(c *ui.Context) {
		CurrencyInput(c, new(float64), CurrencyInputOptions{Symbol: "$"})
	})
}

// ── InputGroup ─────────────────────────────────────────────────────────────

func TestInputGroupPutsItsFurnitureInsideOneField(t *testing.T) {
	host := "api.riverside.test"
	tt := fieldTester(580, 140, func(c *ui.Context) {
		InputGroup(c, &host, InputGroupOptions{
			Label: "Callback host", Prefix: "https://", Suffix: ".com",
			Trailing: func() {
				CopyButton(c, &host, CopyButtonOptions{Label: "Copy host", Plain: true})
			},
		})
	})
	for _, want := range []string{"https://", ".com", "Copy host"} {
		if !tt.HasText(want) {
			t.Errorf("missing %q; the frame has %q", want, tt.Texts())
		}
	}
	// The scheme is not part of the value: what is typed is what the value
	// holds, and the field keeps its own punctuation to itself.
	typingInto(t, tt, "Callback host", "-eu")
	if host != "api.riverside.test-eu" {
		t.Errorf("value = %q, want the host with what was typed", host)
	}
	if strings.Contains(host, "https://") {
		t.Error("the prefix belongs to the field, not to the value")
	}
}

func TestCopyButtonCopiesTheValueAndSaysSo(t *testing.T) {
	token := "cb_9f2a"
	tt := fieldTester(560, 140, func(c *ui.Context) {
		CopyButton(c, &token, CopyButtonOptions{Label: "Copy token"})
	})
	if err := tt.Click("Copy token"); err != nil {
		t.Fatal(err)
	}
	if got := tt.Clipboard(); got != "cb_9f2a" {
		t.Errorf("clipboard = %q, want the value the caller points at", got)
	}
	// The button says it worked, so somebody who looked away for the click
	// is not left pressing it again.
	if !tt.HasText("Copied") {
		t.Errorf("the button should show that it worked: %q", tt.Texts())
	}
}

func TestCopyButtonRefusesNothingToCopy(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("a copy button with nothing to copy should panic")
		}
	}()
	fieldTester(320, 120, func(c *ui.Context) {
		CopyButton(c, nil, CopyButtonOptions{Label: "Copy token"})
	})
}

// ── FormField ──────────────────────────────────────────────────────────────

func TestFormFieldPutsItsLabelAboveItsControl(t *testing.T) {
	cost := 120.0
	tt := fieldTester(560, 320, func(c *ui.Context) {
		Form(c, FormOptions{
			Label:  "Log callback",
			Fields: func() { callbackFields(c, &cost, "") },
		})
	})
	label := boxOf(t, tt, "Customer")
	control := boxOf(t, tt, "Customer, name")
	// The label is above its control and not beside it. A long label in a
	// narrow drawer would wrap against a column of labels and misalign every
	// field below it; above the control it wraps against the control's own
	// width, which is the width it has to fit in anyway.
	if label.Y >= control.Y {
		t.Errorf("the label is at y=%.1f and its control at y=%.1f; the label is meant to be above",
			label.Y, control.Y)
	}
}

func TestFormFieldShowsTheErrorOnceAndOutlinesItsControl(t *testing.T) {
	cost := 120.0
	err := "Enter an amount before saving"
	tt := fieldTester(560, 320, func(c *ui.Context) {
		Form(c, FormOptions{
			Label:  "Log callback",
			Fields: func() { callbackFields(c, &cost, err) },
		})
	})
	if !tt.HasText(err) {
		t.Errorf("the field should explain itself: %q", tt.Texts())
	}
	if n := countOccurrences(tt.Texts(), err); n != 1 {
		t.Errorf("the error is on screen %d times, want once", n)
	}
	// One outline, in the danger colour, drawn by the control's own box: the
	// field did not draw a second one around it.
	wantHairline(t, tt, "Open cost, amount", theme.Light(), theme.Light().Danger)

	k := theme.Light()
	plate := boxOf(t, tt, err)
	// The message sits on the danger pair, which is the one combination in
	// the palette that says "this is wrong" without a line of red text to be
	// read carefully to be found. The sample is the plate's own right end:
	// the mark and the words are at the left, and a point between the glyphs
	// reads as antialiased text rather than as the plate.
	wantPixel(t, tt, plate.X+plate.W-3, plate.Y+plate.H/2, k.DangerBg, "the plate under an error")
}

func TestAnUnmarkedFieldKeepsTheOrdinaryOutlineInsideAForm(t *testing.T) {
	cost := 120.0
	tt := fieldTester(560, 320, func(c *ui.Context) {
		Form(c, FormOptions{
			Label:  "Log callback",
			Fields: func() { callbackFields(c, &cost, "") },
		})
	})
	wantHairline(t, tt, "Open cost, amount", theme.Light(), theme.Light().Border)
}

func TestFormFieldMarksItsDescription(t *testing.T) {
	cost := 0.0
	tt := fieldTester(560, 320, func(c *ui.Context) {
		Form(c, FormOptions{
			Label: "Log callback",
			Fields: func() {
				FormField(c, FormFieldOptions{
					Label: "Open cost", Description: "As it appears on the invoice",
				}, func(string) *ui.Element {
					return CurrencyInput(c, &cost, CurrencyInputOptions{
						Label: "Open cost",
					}).Element
				})
			},
		})
	})
	if !tt.HasText("As it appears on the invoice") {
		t.Errorf("the field should say what it wants: %q", tt.Texts())
	}
}

func TestAFormIsItsFieldsAndASubmitRow(t *testing.T) {
	cost := 120.0
	saved := 0
	tt := fieldTester(600, 460, func(c *ui.Context) {
		Form(c, FormOptions{
			Label:   "Log callback",
			Title:   "Log callback",
			Fields:  func() { callbackFields(c, &cost, "") },
			Divider: true,
			Actions: func() {
				if Button(c, "Save", ButtonOptions{Primary: true}).Clicked() {
					saved++
				}
			},
		})
	})
	for _, want := range []string{"Log callback", "Customer", "Open cost", "Save"} {
		if !tt.HasText(want) {
			t.Errorf("missing %q; the frame has %q", want, tt.Texts())
		}
	}
	if err := tt.Click("Save"); err != nil {
		t.Fatal(err)
	}
	if saved != 1 {
		t.Errorf("saves = %d, want 1", saved)
	}
}

func TestFormInsistsOnALabel(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("an unnamed form should panic")
		}
	}()
	fieldTester(420, 200, func(c *ui.Context) {
		Form(c, FormOptions{})
	})
}

func TestFormFieldNeedsALabel(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("a field with no label is a field nobody can read")
		}
	}()
	fieldTester(420, 200, func(c *ui.Context) {
		FormField(c, FormFieldOptions{}, func(string) *ui.Element { return nil })
	})
}

// callbackFields is the form most of the tests above build: the two fields a
// callback is logged with, which is what this interface edits records with.
// The error goes to the field that is wrong, and the field hands it to the
// control that carries it.
func callbackFields(c *ui.Context, cost *float64, err string) {
	name, issue := "", ""
	FormField(c, FormFieldOptions{Label: "Customer", Required: true},
		func(fieldErr string) *ui.Element {
			return TextInput(c, &name, TextInputOptions{
				Label: "Customer, name", Placeholder: "Who is waiting", Error: fieldErr,
			})
		})
	FormField(c, FormFieldOptions{Label: "Issue"}, func(string) *ui.Element {
		return TextInput(c, &issue, TextInputOptions{Label: "Issue"})
	})
	// The control's own name says which part of the field it is, as well as
	// what the field is: a field's visible label and its control answer to
	// the same words, and two elements sharing one name is how a test clicks
	// the wrong one.
	FormField(c, FormFieldOptions{Label: "Open cost", Error: err},
		func(fieldErr string) *ui.Element {
			return CurrencyInput(c, cost, CurrencyInputOptions{
				Label: "Open cost, amount", Symbol: "$", Error: fieldErr,
			}).Element
		})
}

// ── Banner ─────────────────────────────────────────────────────────────────

func TestBannerSaysWhatItHasToSay(t *testing.T) {
	tt := fieldTester(600, 160, func(c *ui.Context) {
		Banner(c, BannerOptions{
			Text: "This callback has been open for 6 days", Severity: core.Warning,
		})
	})
	if !tt.HasText("This callback has been open for 6 days") {
		t.Errorf("the banner must say its text: %q", tt.Texts())
	}
	// The severity tints its one mark, not the strip it sits on: a banner at
	// the top of a form that coloured its surface would be shouting at
	// somebody who has not done anything wrong yet.
	k := theme.Light()
	r := boxOf(t, tt, "This callback has been open for 6 days")
	wantPixel(t, tt, r.X+r.W/2, r.Y+r.H-2, k.Surface, "the banner's own surface")
}

func TestBannerNeedsSomethingToSay(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("an empty banner should panic: it is a coloured box")
		}
	}()
	fieldTester(420, 160, func(c *ui.Context) {
		Banner(c, BannerOptions{})
	})
}

func TestBannerReportsWhatWasPressedInIt(t *testing.T) {
	var actioned, dismissed bool
	tt := fieldTester(600, 160, func(c *ui.Context) {
		r := Banner(c, BannerOptions{
			Text: "The last save failed", Severity: core.Danger,
			Action: "Retry", Dismissable: true,
		})
		actioned, dismissed = actioned || r.Actioned(), dismissed || r.Dismissed()
	})
	if err := tt.Click("Retry"); err != nil {
		t.Fatal(err)
	}
	if !actioned {
		t.Error("the action button reports its press")
	}
	if err := tt.Click("Dismiss"); err != nil {
		t.Fatal(err)
	}
	if !dismissed {
		t.Error("the close button reports its press")
	}
}

// ── dark ───────────────────────────────────────────────────────────────────

func TestTheTextControlsComeOutRightInTheDark(t *testing.T) {
	name, code := "", ""
	cost := 1234.5
	tt := darkTester(640, 340, func(c *ui.Context) {
		Form(c, FormOptions{
			Label: "Log callback",
			Fields: func() {
				FormField(c, FormFieldOptions{Label: "Customer"},
					func(fieldErr string) *ui.Element {
						return TextInput(c, &name, TextInputOptions{
							Label: "Customer, name", Error: fieldErr,
						})
					})
				FormField(c, FormFieldOptions{Label: "Open cost"},
					func(fieldErr string) *ui.Element {
						return CurrencyInput(c, &cost, CurrencyInputOptions{
							Label: "Open cost, amount", Symbol: "$", Error: fieldErr,
						}).Element
					})
				FormField(c, FormFieldOptions{
					Label: "Code", Error: "That code is wrong",
				}, func(fieldErr string) *ui.Element {
					return MaskedInput(c, &code, MaskedInputOptions{
						Label: "Code, digits", Mask: "000-000", Error: fieldErr,
					})
				})
			},
		})
	})
	// Every control drew, and every one of them took its colours from the
	// dark palette rather than from the light one they were written against.
	for _, want := range []string{"Customer, name", "Open cost", "Code, digits", "That code is wrong"} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in the dark; the frame has %q", want, tt.Texts())
		}
	}
	dark := theme.Dark()
	wantHairline(t, tt, "Open cost, amount", dark, dark.Border)
	wantHairline(t, tt, "Code, digits", dark, dark.Danger)

	// And the well is the dark surface, not the light one it would be if a
	// control had painted its own background out of a palette of its own.
	wantFill(t, tt, "Customer, name", dark, dark.Surface)
	wantFill(t, tt, "Code, digits", dark, dark.Surface)
}
