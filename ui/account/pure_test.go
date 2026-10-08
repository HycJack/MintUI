package account

import (
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
)

// ── MaskKey ────────────────────────────────────────────────────────────────
// The one function in this package a secret passes through, so its four edges
// are asserted exactly rather than by shape.

func TestMaskKeyOnNothing(t *testing.T) {
	// Empty is empty rather than a row of dots: "no key yet" must not look
	// like "a key I cannot read".
	if got := MaskKey(""); got != "" {
		t.Errorf(`MaskKey("") = %q, want ""`, got)
	}
}

func TestMaskKeyShorterThanTheReveal(t *testing.T) {
	// One, two and three: showing all of a key this short is showing the key.
	for _, k := range []string{"a", "ab", "abc"} {
		want := strings.Repeat("•", len(k))
		if got := MaskKey(k); got != want {
			t.Errorf("MaskKey(%q) = %q, want %q", k, got, want)
		}
	}
}

func TestMaskKeyOfExactlyFourShowsNothing(t *testing.T) {
	// The case the rule exists for: revealing the last four of a four
	// character key reveals the key.
	if got := MaskKey("abcd"); got != "••••" {
		t.Errorf(`MaskKey("abcd") = %q, want "••••"`, got)
	}
	if got := MaskKey("1234"); got != "••••" {
		t.Errorf(`MaskKey("1234") = %q, want "••••"`, got)
	}
}

func TestMaskKeyOfALongKeyShowsItsLastFour(t *testing.T) {
	// Twenty-four characters: twenty dots and the last four.
	if got := MaskKey("cb_live_9f2a8b7c1d4e6f8a"); got != strings.Repeat("•", 20)+"6f8a" {
		t.Errorf("MaskKey of a long key = %q", got)
	}
	// Five is the first length where there is anything to hide.
	if got := MaskKey("abcde"); got != "•bcde" {
		t.Errorf(`MaskKey("abcde") = %q, want "•bcde"`, got)
	}
}

func TestMaskKeyNeverReturnsTheWholeKey(t *testing.T) {
	// The property the three tests above are cases of, checked over a range
	// so that a change to the rule cannot quietly start leaking.
	for n := 1; n <= 40; n++ {
		key := strings.Repeat("k", n)
		got := MaskKey(key)
		if n <= revealLen && strings.Trim(got, "•") != "" {
			t.Errorf("MaskKey of %d characters revealed %q", n, got)
		}
		if len([]rune(got)) != n {
			t.Errorf("MaskKey of %d characters is %d long: %q", n, len([]rune(got)), got)
		}
	}
}

func TestMaskKeyCutsAtACharacterNotAByte(t *testing.T) {
	// A rune-aware mask: five two-byte characters cut after four characters,
	// not in the middle of one.
	// Six characters, so two dots and the last four of them.
	got := MaskKey("ключ-1")
	if want := "••" + "юч-1"; got != want {
		t.Errorf("MaskKey of a non-ASCII key = %q, want %q", got, want)
	}
	if strings.ContainsRune(got, '�') {
		t.Errorf("MaskKey split a character: %q", got)
	}
}

func TestFingerprint(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", ""},
		// Too short to show both ends without showing the lot, so it falls
		// back to the mask rather than to a leak.
		{"abcd", "••••"},
		// Twenty-four characters: two ends of four and sixteen elided.
		{"cb_live_9f2a8b7c1d4e6f8a", "cb_l…+16…6f8a"},
		// Exactly eight is both ends of four, which is the whole key, so it
		// falls back to the mask rather than printing it back at you.
		{"abcdefgh", "••••efgh"},
		{"abcdefghi", "abcd…+1…fghi"},
	}
	for _, c := range cases {
		if got := Fingerprint(c.in); got != c.want {
			t.Errorf("Fingerprint(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// ── the rules a form needs ─────────────────────────────────────────────────

func TestValidateLoginSaysNothingAboutAnEmptyForm(t *testing.T) {
	emailErr, passErr := ValidateLogin("", "")
	if emailErr != "Enter your email address" {
		t.Errorf("empty address = %q", emailErr)
	}
	if passErr != "Enter your password" {
		t.Errorf("empty password = %q", passErr)
	}
}

func TestValidateLoginOnAWellFormedPair(t *testing.T) {
	emailErr, passErr := ValidateLogin("  rosa@riverside.clinic ", "correct-horse")
	if emailErr != "" || passErr != "" {
		t.Errorf("a good pair gave %q and %q", emailErr, passErr)
	}
}

func TestValidateLoginOnTheEdges(t *testing.T) {
	cases := []struct {
		email, password, wantEmail, wantPass string
	}{
		{"rosa@clinic.com", "short", "", "Passwords are at least 8 characters"},
		{"rosaclinic.com", "correct-horse", "That does not look like an email address", ""},
		{"@clinic.com", "correct-horse", "That does not look like an email address", ""},
		{"rosa@", "correct-horse", "That does not look like an email address", ""},
		{"   ", "12345678", "Enter your email address", ""},
		{"a@b.co", "1234567", "", "Passwords are at least 8 characters"},
	}
	for _, c := range cases {
		gotEmail, gotPass := ValidateLogin(c.email, c.password)
		if gotEmail != c.wantEmail {
			t.Errorf("ValidateLogin(%q).address = %q, want %q", c.email, gotEmail, c.wantEmail)
		}
		if gotPass != c.wantPass {
			t.Errorf("ValidateLogin(%q).password = %q, want %q", c.password, gotPass, c.wantPass)
		}
	}
}

// ── quota ──────────────────────────────────────────────────────────────────

func TestQuotaFraction(t *testing.T) {
	cases := []struct {
		used, limit int
		want        float32
	}{
		{0, 100, 0},
		{25, 100, 0.25},
		{100, 100, 1},
		{250, 100, 1}, // over the limit draws as full, not as 2.5
		{50, 0, 0},    // unmetered is not "no limit, so all of it used"
		{50, -1, 0},   // nor is a negative limit a division
		{0, 0, 0},
	}
	for _, c := range cases {
		if got := QuotaFraction(c.used, c.limit); got != c.want {
			t.Errorf("QuotaFraction(%d, %d) = %v, want %v", c.used, c.limit, got, c.want)
		}
	}
}

func TestQuotaToneStepsOnlyAtTheEnd(t *testing.T) {
	cases := []struct {
		used, limit int
		want        core.Severity
	}{
		{0, 1000, core.Neutral},
		{500, 1000, core.Neutral},
		{749, 1000, core.Neutral},
		{750, 1000, core.Accent},
		{899, 1000, core.Accent},
		{900, 1000, core.Warning},
		{999, 1000, core.Warning},
		{1000, 1000, core.Danger},
		{1200, 1000, core.Danger},
		{500, 0, core.Neutral},
	}
	for _, c := range cases {
		if got := QuotaTone(c.used, c.limit); got != c.want {
			t.Errorf("QuotaTone(%d, %d) = %v, want %v", c.used, c.limit, got, c.want)
		}
	}
}

// ── the settings search ────────────────────────────────────────────────────

func settingsFixture() []Setting {
	return []Setting{
		{ID: "appearance", Name: "Appearance", Hint: "Light, dark or your system", Keywords: []string{"theme", "colour"}},
		{ID: "keys", Name: "API keys", Hint: "Credentials for the API"},
		{ID: "sync", Name: "Sync", Hint: "Refresh when the app is opened", Keywords: []string{"refresh"}},
		{ID: "account", Name: "Account", Hint: "Your name and email"},
		{ID: "shortcuts", Name: "Shortcuts", Hint: "Every key binding"},
	}
}

func TestMatchSettingsWithNothingTyped(t *testing.T) {
	got := MatchSettings("", settingsFixture())
	if len(got) != 5 || got[0].ID != "appearance" || got[4].ID != "shortcuts" {
		t.Errorf("an empty query must leave the list alone, got %v", ids(got))
	}
}

func TestMatchSettingsRanksByHowSureItIs(t *testing.T) {
	// "key" is at the start of one name and inside another's hint, and is a
	// keyword of a third. The order says how surprised somebody would be to
	// find each row where it is.
	got := MatchSettings("key", settingsFixture())
	if want := []string{"keys", "shortcuts"}; !sameIDs(got, want) {
		t.Errorf("MatchSettings(key) = %v, want %v", ids(got), want)
	}

	// "sy" is the start of one name and two letters into another's hint, so
	// this asserts both the prefix tier and the mention tier in one query.
	got = MatchSettings("sy", settingsFixture())
	if want := []string{"sync", "appearance"}; !sameIDs(got, want) {
		t.Errorf("MatchSettings(sy) = %v, want %v", ids(got), want)
	}
}

func TestMatchSettingsIsCaseAndSpaceInsensitive(t *testing.T) {
	got := MatchSettings("  ACCOUNT  ", settingsFixture())
	if want := []string{"account"}; !sameIDs(got, want) {
		t.Errorf("MatchSettings(ACCOUNT) = %v, want %v", ids(got), want)
	}
}

func TestMatchSettingsMatchingNothing(t *testing.T) {
	got := MatchSettings("zzzz", settingsFixture())
	if len(got) != 0 {
		t.Errorf("MatchSettings for a word that is nowhere = %v", ids(got))
	}
}

func TestMatchSettingsKeepsEqualRowsWhereTheyWere(t *testing.T) {
	// Two rows that match the same tier keep the caller's order, so the list
	// does not reshuffle itself as somebody narrows a search down.
	list := []Setting{
		{ID: "b", Name: "Beta", Hint: "x"},
		{ID: "a", Name: "Alpha", Hint: "y"},
	}
	if got := MatchSettings("", list); got[0].ID != "b" {
		t.Errorf("order changed: %v", ids(got))
	}
}

func TestMatchSettingsDoesNotHandBackTheCallersSlice(t *testing.T) {
	list := settingsFixture()
	_ = MatchSettings("key", list)
	if list[0].ID != "appearance" {
		t.Error("MatchSettings reordered the caller's own slice")
	}
}

func ids(s []Setting) []string {
	out := make([]string, len(s))
	for i := range s {
		out[i] = s[i].ID
	}
	return out
}

func sameIDs(got []Setting, want []string) bool {
	g := ids(got)
	if len(g) != len(want) {
		return false
	}
	for i := range g {
		if g[i] != want[i] {
			return false
		}
	}
	return true
}

// ── chords ─────────────────────────────────────────────────────────────────

func TestNormalKey(t *testing.T) {
	cases := []struct{ in, want string }{
		{"k", "K"}, // one character is upper-cased whatever case it was written in
		{"K", "K"}, //
		{"CMD", "cmd"},
		{"  Shift ", "shift"},
		{"ESC", "esc"},
		{"escape", "esc"},
		{"f12", "F12"},
		{"F5", "F5"},
		{"fx", "fx"}, // not an F-key: left exactly as written
		{"", ""},
	}
	for _, c := range cases {
		if got := NormalKey(c.in); got != c.want {
			t.Errorf("NormalKey(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestNormalKeysDropsEmptyParts(t *testing.T) {
	got := NormalKeys([]string{"cmd", "", " ", "k"})
	if want := []string{"cmd", "K"}; len(got) != len(want) {
		t.Fatalf("NormalKeys = %v, want %v", got, want)
	} else {
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("NormalKeys = %v, want %v", got, want)
			}
		}
	}
}

func TestFormatChord(t *testing.T) {
	cases := []struct {
		in   []string
		want string
	}{
		{nil, ""},
		{[]string{}, ""},
		// Glyphs carry their own width, so they run together: this is what
		// makes a chord read as one key combination.
		{[]string{"cmd", "shift", "k"}, "⌘⇧K"},
		{[]string{"cmd", "k"}, "⌘K"},
		{[]string{"f5"}, "F5"},
		// A word gets a separator; a symbol never does.
		{[]string{"ctrl", "k"}, "Ctrl+K"},
		{[]string{"cmd", "enter"}, "⌘+Return"},
		{[]string{"escape"}, "Esc"},
		{[]string{"cmd", "alt", "delete"}, "⌘⌥+Delete"},
		{[]string{"k"}, "K"},
		{[]string{"space"}, "Space"},
		{[]string{"up"}, "↑"},
	}
	for _, c := range cases {
		if got := FormatChord(c.in); got != c.want {
			t.Errorf("FormatChord(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestChordsConflictOnADuplicateKey(t *testing.T) {
	// Two commands on one chord is not a warning: the runtime gives the key
	// to the innermost element that asks for it, so one silently never fires.
	ok := []Shortcut{
		{ID: "open", Keys: []string{"cmd", "k"}},
		{ID: "save", Keys: []string{"cmd", "s"}},
	}
	if ChordsConflict(ok) {
		t.Error("two different chords were called a conflict")
	}

	clash := append(append([]Shortcut{}, ok...), Shortcut{ID: "search", Keys: []string{"CMD", "K"}})
	if !ChordsConflict(clash) {
		t.Error("the same chord twice must be a conflict, whatever case it was written in")
	}
}

func TestChordsConflictOnADuplicateCommand(t *testing.T) {
	list := []Shortcut{
		{ID: "open", Keys: []string{"cmd", "k"}},
		{ID: "open", Keys: []string{"cmd", "p"}},
	}
	if !ChordsConflict(list) {
		t.Error("one command bound twice must be a conflict")
	}
}

func TestChordsConflictIgnoresAnUnboundRow(t *testing.T) {
	list := []Shortcut{
		{ID: "open", Keys: []string{"cmd", "k"}},
		{ID: "clear", Keys: nil},
		{ID: "reset", Keys: []string{" "}},
	}
	if ChordsConflict(list) {
		t.Error("rows with no chord are not a conflict with each other")
	}
}

// ── the accent's foreground ────────────────────────────────────────────────

func TestContrastInkCrossover(t *testing.T) {
	// A light colour takes near-black ink and a dark one takes white, which
	// is the whole job: the two palette inks with the wrong one behind them
	// are the two ways a preview row becomes unreadable.
	if got := contrastInk(ui.Hex("#ffffff")); got != ui.Hex("#18181b") {
		t.Errorf("white takes near-black ink, got %v", got)
	}
	if got := contrastInk(ui.Hex("#18181b")); got != ui.Hex("#ffffff") {
		t.Errorf("near-black takes white ink, got %v", got)
	}
	// The library's two accents, which between them cover both sides.
	if got := contrastInk(ui.Hex("#d9f24b")); got != ui.Hex("#18181b") {
		t.Errorf("the lively colour takes near-black, got %v", got)
	}
	if got := contrastInk(ui.Hex("#1d4ed8")); got != ui.Hex("#ffffff") {
		t.Errorf("a deep blue takes white, got %v", got)
	}
}

func TestRelativeLuminanceIsTheWCAGOne(t *testing.T) {
	if got := relativeLuminance(ui.Hex("#000000")); got != 0 {
		t.Errorf("black's luminance = %v, want 0", got)
	}
	if got := relativeLuminance(ui.Hex("#ffffff")); got < 0.9999 {
		t.Errorf("white's luminance = %v, want 1", got)
	}
}

// ── the small helpers ──────────────────────────────────────────────────────

func TestItoa(t *testing.T) {
	cases := []struct {
		in   int
		want string
	}{{0, "0"}, {7, "7"}, {10, "10"}, {99, "99"}, {100, "100"}, {1234567, "1234567"}, {-42, "-42"}}
	for _, c := range cases {
		if got := itoa(c.in); got != c.want {
			t.Errorf("itoa(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestCommas(t *testing.T) {
	cases := []struct {
		in   int
		want string
	}{
		{0, "0"}, {7, "7"}, {999, "999"},
		{1000, "1,000"}, {1204, "1,204"},
		{20000, "20,000"}, {1000000, "1,000,000"},
		{-1204, "-1,204"},
	}
	for _, c := range cases {
		if got := commas(c.in); got != c.want {
			t.Errorf("commas(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestTrimmed(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", ""}, {"   ", ""}, {"\n\t ", ""},
		{"a", "a"}, {"  a  ", "a"}, {"a b", "a b"},
	}
	for _, c := range cases {
		if got := trimmed(c.in); got != c.want {
			t.Errorf("trimmed(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestRuneCountCountsCharacters(t *testing.T) {
	if got := runeCount("abc"); got != 3 {
		t.Errorf("runeCount(abc) = %d", got)
	}
	// Three emoji are three characters and nine bytes; a limit of 500 has to
	// mean 500 of the former.
	if got := runeCount("🎯🎯🎯"); got != 3 {
		t.Errorf("runeCount of three emoji = %d, want 3", got)
	}
}

func TestFirstLine(t *testing.T) {
	if got := firstLine("Log a callback\nIt takes four fields"); got != "Log a callback" {
		t.Errorf("firstLine = %q", got)
	}
	if got := firstLine("One line"); got != "One line" {
		t.Errorf("firstLine of one line = %q", got)
	}
}

func TestSince(t *testing.T) {
	now := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		at   time.Time
		want string
	}{
		{now.Add(-30 * time.Second), "just now"},
		{now.Add(-5 * time.Minute), "5 min ago"},
		{now.Add(-3 * time.Hour), "3 h ago"},
		{now.Add(-50 * time.Hour), "2 d ago"},
	}
	for _, c := range cases {
		if got := since(c.at, now); got != c.want {
			t.Errorf("since = %q, want %q", got, c.want)
		}
	}
}
