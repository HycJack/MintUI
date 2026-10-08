package git

import (
	"strings"
	"unicode/utf8"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/theme"
)

// Git is the one program in a desktop that writes escape sequences into text
// it hands to somebody else: `git diff --color`, `git log --color`, `git
// branch --color`, every pager, every `git config color.ui auto` in the world.
// That text arrives at a component already coloured by a palette that knows
// nothing about this window's theme.
//
// So there are two ways to take it apart and both are here as plain
// functions. StripANSI throws the sequences away, which is what measuring and
// searching need — a row of text whose width includes fourteen invisible
// characters wraps in the wrong place, and `git grep` for "error" misses a
// red one. ParseANSI keeps what each sequence said and throws the rest away,
// which is what drawing needs.

// esc is the byte that starts every ANSI sequence, and escText is the
// same byte as a string, which is what the one scan below needs.
const (
	esc     = byte(0x1b)
	escText = string(rune(esc))
)

// StripANSI removes every escape sequence from s and leaves the text.
//
// It knows three kinds, because those are the three git's palette actually
// emits and the three anything else emits:
//
//   - CSI: ESC [ … final byte — the colours, the cursor moves, the erases
//   - OSC: ESC ] … BEL or ESC \ — the window title, which a pager writes
//   - two-character escapes: ESC followed by one byte, such as ESC ( B
//
// Anything it does not recognise is dropped up to and including the next
// printable byte rather than left in the string. A half-stripped escape is
// worse than a missing one: it is invisible in the editor and visible in the
// width, which is the failure this function exists to prevent.
//
// Invalid UTF-8 is not repaired here. A rune that was already wrong before it
// reached us is a bug in what produced the text, and pretending otherwise
// here would hide it from the one place that can fix it.
func StripANSI(s string) string {
	if !strings.Contains(s, escText) {
		// The common case, and the one that matters for a diff with no
		// colour in it: no byte to look for, so the string goes back as it
		// came in, with no copy made.
		return s
	}

	var out strings.Builder
	out.Grow(len(s))
	for i := 0; i < len(s); {
		if s[i] != esc {
			out.WriteByte(s[i])
			i++
			continue
		}
		i++
		if i >= len(s) {
			// A lone ESC at the end of the string: there is nothing after it
			// to drop, so nothing is.
			break
		}
		switch s[i] {
		case '[': // CSI
			i++
			for i < len(s) && !isCSIEnd(s[i]) {
				i++
			}
			if i < len(s) {
				i++ // the final byte is part of the sequence
			}
		case ']': // OSC
			i++
			for i < len(s) {
				if s[i] == 0x07 { // BEL ends it
					i++
					break
				}
				if s[i] == esc && i+1 < len(s) && s[i+1] == '\\' {
					i += 2 // ST ends it
					break
				}
				i++
			}
		case 'P', 'X', '^', '_': // DCS, SOS, PM, APC: all end at ST
			i++
			for i < len(s) {
				if s[i] == esc && i+1 < len(s) && s[i+1] == '\\' {
					i += 2
					break
				}
				i++
			}
		default:
			// ESC followed by an intermediate byte (0x20-0x2F) runs to a
			// final byte: ESC ( B designates a character set, ESC # 8 puts
			// up a line, ESC % G selects a font. One byte is not enough,
			// and leaving the final byte in the string is how "ESC ( B"
			// ends up printing a stray B in the middle of a log line.
			if i < len(s) && s[i] >= 0x20 && s[i] <= 0x2f {
				for i < len(s) && s[i] >= 0x20 && s[i] <= 0x2f {
					i++
				}
				if i < len(s) {
					i++ // the final byte
				}
				break
			}
			i++ // ESC with nothing after it but one byte: ESC 7, ESC =
		}
	}
	return out.String()
}

// isCSIEnd reports whether b ends a CSI sequence. The range runs from @ (the
// first parameter byte) to ~ (the last one before the final byte), so the
// intermediate bytes of a colour are all inside it.
func isCSIEnd(b byte) bool { return b >= '@' && b <= '~' }

// AnsiSpan is a run of text that was written under one set of SGR codes.
//
// The codes are kept rather than resolved to a colour, because a colour needs
// the window's palette and this function has none of it: parsing is a
// property of the text, and the same text must read the same way in both
// appearances.
type AnsiSpan struct {
	// Text is the run itself, with no escape sequences in it.
	Text string
	// Codes are the SGR parameters in force over the run: 31 for red, 1 for
	// bold, 0 for neither. Empty means the run is in whatever the terminal
	// default is, which is the window's own body ink.
	Codes []int
}

// ParseANSI cuts s into the runs of text between its SGR sequences.
//
// Only SGR is read. A cursor move or an erase changes what is on the screen
// rather than what colour a character is, and a component that drew one would
// have to become a terminal emulator to make it mean anything. A sequence
// that is not SGR is consumed and otherwise ignored, so the text after it
// still parses.
//
// The first run always exists, even for an empty string, because a component
// drawing spans needs something to draw before the first escape as much as
// after it.
func ParseANSI(s string) []AnsiSpan {
	var out []AnsiSpan
	var codes []int
	start := 0

	for i := 0; i < len(s); {
		if s[i] != esc {
			i++
			continue
		}
		if i+1 < len(s) && s[i+1] == '[' {
			j := i + 2
			for j < len(s) && !isCSIEnd(s[j]) {
				j++
			}
			if j >= len(s) {
				// A sequence that never ends: the rest of the string is
				// inside it, so there is no rest of the string.
				break
			}
			if i > start {
				// The run is closed either way. A sequence that is not SGR
				// is consumed and then dropped, exactly as StripANSI drops
				// it: a cursor move is not text and leaving it in the span
				// would put the escape bytes into a measurement.
				out = append(out, AnsiSpan{Text: s[start:i], Codes: codes})
			}
			if s[j] == 'm' {
				codes = parseSGR(s[i+2 : j])
			}
			start = j + 1
			i = j + 1
			continue
		}
		i++
	}
	if start < len(s) || len(out) == 0 {
		out = append(out, AnsiSpan{Text: s[start:], Codes: codes})
	}
	return out
}

// parseSGR reads the numbers of one SGR sequence, ignoring the private and
// intermediate bytes some tools put in front of them.
func parseSGR(params string) []int {
	params = strings.TrimLeft(params, " ?>!")
	if params == "" {
		// A bare ESC [ m is SGR 0, which resets rather than does nothing.
		return []int{0}
	}
	var out []int
	for _, field := range strings.Split(params, ";") {
		n, err := atoi(field)
		if err != nil {
			continue
		}
		out = append(out, n)
	}
	return out
}

// AnsiInk is what one SGR code says about a run's appearance in this window's
// palette: the ink to draw it in, and whether it is bold.
//
// The sixteen colours of the ANSI palette have no tokens of their own, so
// they are mapped onto the four the library has. That mapping is a judgement
// rather than a lookup, and it is the same judgement in both appearances:
// bright codes are for "something is wrong" and land on the danger ink, dim
// ones on the muted ink, and the two that carry no feeling — cyan and
// magenta — stay out of the status ramp entirely and take the accent.
//
// A code that says nothing about colour, such as 1 for bold, reports the ink
// as ok=false and leaves the caller with what it already had.
func AnsiInk(code int, k theme.Tokens) (ink ui.Color, bold, ok bool) {
	switch code {
	case 1, 2:
		return k.Text, true, true
	case 22:
		return k.Text, false, true
	case 30, 90:
		return k.Text, false, true
	case 31, 91:
		return k.Danger, false, true
	case 32, 92:
		return k.Success, false, true
	case 33, 93:
		return k.Warning, false, true
	case 34, 94:
		return k.Accent, false, true
	case 35, 95:
		return k.Accent, false, true
	case 36, 96:
		return k.Accent, false, true
	case 37, 97:
		return k.Text, false, true
	case 39:
		return k.Text, false, true
	}
	return ink, false, false
}

// atoi is strconv.Atoi without the error, for the numbers inside an escape
// sequence where a field that is not a number is a field to skip.
func atoi(s string) (int, error) {
	if s == "" {
		return 0, nil
	}
	n := 0
	for i := range len(s) {
		if s[i] < '0' || s[i] > '9' {
			return 0, errNotANumber
		}
		n = n*10 + int(s[i]-'0')
		if n > 1<<20 {
			// Past any real SGR number, and far past anything a terminfo
			// entry uses. Stop rather than overflow on a hostile string.
			return 0, errNotANumber
		}
	}
	return n, nil
}

type numberError struct{}

func (numberError) Error() string { return "not a number" }

var errNotANumber = numberError{}

// ValidUTF8 reports whether s is text rather than bytes, which is how a
// component tells a file it can show from one it cannot.
//
// It is here because the same question is asked twice — once of diff text and
// once of a file to preview — and both answers must agree, or the same file
// previews as text in one window and as binary in another. The test is
// deliberately cheap and deliberately boring: no encoding sniffing, because
// guessing is how a half-decoded page becomes a screenful of replacement
// characters.
func ValidUTF8(s string) bool { return utf8.ValidString(s) }
