package account

import "strings"

// Shortcut is one binding: a command and the chord that invokes it.
type Shortcut struct {
	// ID is the command's name, and what a caller writes the recording back
	// to. It is separate from Keys for the reason a Choice's Value is: the
	// chord is what is pressed, the ID is what is acted on, and the two stop
	// agreeing the first time somebody rebinds a key.
	ID string
	// Name is what a person reads.
	Name string
	// Keys is the chord, in press order: "cmd", "shift", "k". Stored as the
	// names rather than as key codes because the binding is written in a
	// file, in a settings screen and in a test, and all three want words.
	Keys []string
	// Group is the heading the binding is listed under.
	Group string
}

// chordMarks and chordWords are the two ways a key can be written.
//
// chordMarks are the ones that have a symbol: they carry their own width and
// shape, and two of them side by side already read as one chord. chordWords
// are the ones that have to be spelled, because there is no symbol for them
// that means the same thing on both platforms — "Ctrl" on a Mac is a
// different key from "⌃".
//
// They are two tables rather than one with a flag because which set a key is
// in is the whole of how it is written, and a lookup that answers both
// questions is a lookup with two answers.
var chordMarks = map[string]string{
	"cmd": "⌘", "alt": "⌥", "shift": "⇧", "tab": "⇥",
	"up": "↑", "down": "↓", "left": "←", "right": "→",
}

var chordWords = map[string]string{
	"ctrl": "Ctrl", "enter": "Return", "esc": "Esc", "space": "Space",
	"delete": "Delete", "home": "Home", "end": "End",
	"pageup": "PageUp", "pagedown": "PageDown",
}

// chordAlias folds the other ways a key is written onto the one this package
// stores. "command", "super" and "meta" are all the same physical key, and a
// bindings file that writes all three for one chord would report a conflict
// against itself.
//
// It is separate from chordNames because that table maps a canonical name to
// its mark, and these map an alias to a canonical name. Two directions, two
// tables.
var chordAlias = map[string]string{
	"command": "cmd",
	"super":   "cmd",
	"meta":    "cmd",
	"control": "ctrl",
	"option":  "alt",
	"return":  "enter",
	"escape":  "esc",
}

// NormalKey is the canonical name of one key of a chord: lower case for
// everything that has a name, upper case for a single character so that "k"
// and "K" cannot both be written and mean two different bindings.
//
// A chord with no name for one of its keys is returned unchanged rather than
// dropped. Silently removing a key turns "cmd+<unheard-of>" into "cmd",
// which is a binding that fires on every press of the modifier.
func NormalKey(key string) string {
	k := strings.ToLower(strings.TrimSpace(key))
	if k == "" {
		return ""
	}
	// A single character is the key itself, whatever case it was written in:
	// "K" and "k" are the same key on every layout this library ships for.
	if r := []rune(k); len(r) == 1 {
		return strings.ToUpper(k)
	}
	if canon, ok := chordAlias[k]; ok {
		return canon
	}
	if isMark(k) {
		return k
	}
	// An F-key is the one family with no entry in chordNames, because the
	// symbol for it is the name.
	if len(k) >= 2 && k[0] == 'f' && isDigits(k[1:]) {
		return "F" + k[1:]
	}
	return k
}

// NormalKeys is a whole chord, each part normalised, with empty parts dropped.
func NormalKeys(keys []string) []string {
	out := make([]string, 0, len(keys))
	for _, k := range keys {
		if n := NormalKey(k); n != "" {
			out = append(out, n)
		}
	}
	return out
}

// FormatChord is a binding as one line of text: "⌘⇧K".
//
// It joins with no separator, and that is the point of the glyphs rather than
// a typographic accident: the symbols are narrow and carry their own
// spacing, so the result reads as one key combination rather than as three
// words with spaces in them. Anything without a symbol keeps its word, so a
// chord on a platform that has no glyph for it still says what it is.
func FormatChord(keys []string) string {
	norm := NormalKeys(keys)
	if len(norm) == 0 {
		return ""
	}
	var b strings.Builder
	prevSymbol := true // nothing written yet reads as "the last was a symbol"
	for _, k := range norm {
		mark, isSymbol := chordMark(k)
		// Two symbols run together: they carry their own width and shape the
		// word "chord". A word on either side of the join gets a plus, so
		// "Ctrl+K" and "⌘+Return" are written the same way rather than one
		// joining and the other not — a chord that changes its punctuation
		// depending on which keys it holds cannot be scanned.
		if b.Len() > 0 && !(isSymbol && prevSymbol) {
			b.WriteString("+")
		}
		b.WriteString(mark)
		prevSymbol = isSymbol
	}
	return b.String()
}

// isMark reports whether a key is one this package has a name for: either a
// symbol or a word it spells. A name it does not have is left alone rather
// than dropped, because dropping it turns "cmd+<unheard-of>" into "cmd", and
// that fires on every press of the modifier.
func isMark(key string) bool {
	lower := strings.ToLower(key)
	if _, ok := chordMarks[lower]; ok {
		return true
	}
	if _, ok := chordWords[lower]; ok {
		return true
	}
	return singleRune(key)
}

// ChordKeys is a binding's keys as display.Kbd should draw them: each one as
// its mark where it has one and as its word where it does not.
//
// It is the counterpart to FormatChord, which joins them into one line. Both
// go through chordMark, so a chord cannot come out as "⌘K" in a table cell and
// "cmd" and "K" in the list beside it.
func ChordKeys(keys []string) []string {
	norm := NormalKeys(keys)
	out := make([]string, 0, len(norm))
	for _, k := range norm {
		mark, _ := chordMark(k)
		out = append(out, mark)
	}
	return out
}

// chordMark is how one key of a chord is drawn: the mark and whether it is a
// mark rather than a word. The order is symbol, then F-key, then word, then
// the name as written — a key this package has no name for is still drawn,
// because the alternative is a chord with a piece missing.
func chordMark(key string) (mark string, isSymbol bool) {
	lower := strings.ToLower(key)
	if glyph, ok := chordMarks[lower]; ok {
		return glyph, true
	}
	if len(lower) >= 2 && lower[0] == 'f' && isDigits(lower[1:]) {
		return "F" + lower[1:], true
	}
	if singleRune(key) {
		return key, true
	}
	if word, ok := chordWords[lower]; ok {
		return word, false
	}
	return key, false
}

// singleRune reports whether a key is one character, which is already a
// glyph's worth of ink whatever it is.
func singleRune(s string) bool { return len([]rune(s)) == 1 }

// ChordsConflict reports the two bindings in a set that cannot both be true —
// the same command, or the same chord twice.
//
// A duplicate chord is the one that matters: the runtime gives a key to the
// innermost element that asks for it and drops the rest, so two commands on
// ⌘K is not a warning, it is one of them silently never firing. A duplicate
// command is the same bug from the other end.
func ChordsConflict(list []Shortcut) bool {
	seenCmd := make(map[string]bool, len(list))
	seenKeys := make(map[string]bool, len(list))
	for _, s := range list {
		if s.ID != "" {
			if seenCmd[s.ID] {
				return true
			}
			seenCmd[s.ID] = true
		}
		chord := FormatChord(s.Keys)
		if chord == "" {
			continue
		}
		if seenKeys[chord] {
			return true
		}
		seenKeys[chord] = true
	}
	return false
}

// isDigits reports whether s is one or more digits and nothing else.
func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
