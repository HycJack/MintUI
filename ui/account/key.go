package account

import "strings"

// revealLen is how much of a key the mask leaves readable.
//
// Four is the number every other product arrived at independently, and the
// reason is a trade rather than a tradition: enough to tell two of somebody's
// own keys apart in a list, short enough that what remains is useless to
// anybody who has seen it. The mask is not a decoration on the key, it is the
// whole of what a person is allowed to read off the screen, so the length is
// a decision about how much secret to show and not a style choice.
const revealLen = 4

// MaskKey is what a secret looks like on screen: dots for all of it but the
// last four characters, which are shown as they are.
//
// It is a function and not a drawing rule because the rule has to be the same
// everywhere a key appears — the row in the manager, the field in the editor,
// the text copied into a bug report — and a rule written three times is three
// rules that will eventually disagree. It takes the key and returns the only
// form of it that may be drawn; it knows nothing about panels, and nothing
// here ever passes its result anywhere but into a label.
//
// The edges are the interesting part:
//
//   - An empty key is empty, not a row of dots. "No key" has to look unlike
//     "a key I cannot read", or a caller that has not loaded its keys yet
//     shows a screen of secrets.
//   - A key of four characters or fewer is all dots. Revealing the last four
//     of a four-character key reveals the key, so the rule has to bend here
//     or the one case where it would leak is the one case where it applies.
//
// It counts runes rather than bytes so that a key containing a non-ASCII
// character is cut at a character boundary; a mask that split a rune in half
// would put a replacement character where a digit should be, which looks like
// a different key rather than like a bug.
func MaskKey(k string) string {
	rs := []rune(k)
	switch {
	case len(rs) == 0:
		return ""
	case len(rs) <= revealLen:
		return strings.Repeat("•", len(rs))
	default:
		return strings.Repeat("•", len(rs)-revealLen) + string(rs[len(rs)-revealLen:])
	}
}

// Fingerprint is the short form of a key that may be shown in full and still
// says nothing: the first four characters and the last four, with the length
// in between as a count rather than as the characters themselves.
//
// It is what a support conversation quotes. "It starts cb_live and is 40
// characters" identifies the key well enough to look it up in a database and
// useless to anybody reading over the desk, which is the same job
// MaskKey does on screen with a narrower net.
func Fingerprint(k string) string {
	rs := []rune(k)
	switch {
	case len(rs) == 0:
		return ""
	case len(rs) <= revealLen*2:
		// Too short to show both ends without showing all of it, so this
		// falls back to the mask rather than to a leak.
		return MaskKey(k)
	default:
		// The count sits between two ellipses rather than beside one of them,
		// so the shape is the same however long the key is and only the
		// number changes as it is compared with another.
		return string(rs[:revealLen]) +
			"…+" + itoa(len(rs)-revealLen*2) + "…" +
			string(rs[len(rs)-revealLen:])
	}
}

// itoa is the one thing strconv is needed for here, and itoa(n) is written
// out rather than imported so that the package's dependency list stays the
// three it needs to draw with.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
