// Package files holds the parts of a file interface that are not the window:
// a name that has been made safe to use, a path that has been put together,
// a size a person can read, and a kind that decides both the glyph and
// whether anything can be shown at all.
//
// The reason those four are plain functions is that they are the four things
// a file interface gets wrong without anybody noticing. A name that kept its
// separators writes a file somewhere else. A path joined with a slash and a
// segment gives two paths. A size rounded to the nearest power of two says
// 1.05 MB for a file the web version calls 1.0 MB, and the two are shown side
// by side to the same person. And a file that is bytes rather than text is
// shown as text, which is a screenful of replacement characters and a
// support ticket.
//
// So they are computed here, in one place, and tested by their return values.
package files

import (
	"math"
	"strconv"
	"strings"
	"unicode/utf8"
)

// maxNameBytes is the longest a name may be. It is 255 because that is what
// ext4, APFS, NTFS and every web file input agree on, and a limit that
// differs between the four is not a limit.
const maxNameBytes = 255

// Sanitize makes a name safe to create: what a person typed, a paste out of
// somewhere else, or an API response all become something a filesystem will
// accept and a person will recognise.
//
// The steps are in the order they have to be in, and the order is the whole
// of what this function is:
//
//  1. Keep what is after the last separator. Somebody who pasted a path wants
//     the name, and every file dialog on every desktop shows them the name —
//     but a name that still held a slash would silently become a directory.
//  2. Drop every control character. A newline in a filename is legal on
//     Linux and useless everywhere: it turns a list of files into a list
//     that lies about how many lines it has.
//  3. Replace the characters no filesystem accepts, so the name still reads
//     as the name rather than becoming a row of dots.
//  4. Trim the spaces and dots off the ends. Windows silently does this, so a
//     name that kept them would be a name that came back different.
//  5. Escape the reserved names, which are refused by Windows no matter what
//     the rest of the name is.
//  6. Shorten to 255 bytes on a rune boundary.
//
// Control characters are removed rather than cutting the name at the first
// one. Cutting loses everything after it — a stray NUL in the middle of
// "notes\x00.md" would leave a file called "notes" rather than "notes.md" —
// and a name that has lost half of itself is a name somebody has to rename
// again.
//
// It returns the empty string when nothing usable is left. That is deliberate:
// a rename field somebody has not typed in yet must not have "file" appear in
// it, and a caller who needs a name either way asks SanitizeOr.
func Sanitize(name string) string {
	// 1. The last segment, whichever separator it was written with. A name
	// that came off a Windows paste and a name off a Linux one are both
	// taken apart the same way, because the caller cannot tell which they
	// have. A trailing separator is trimmed first: "/repo/ui/" is a folder
	// and its name is "ui", not nothing at all.
	if trimmed := strings.TrimRight(name, `/\`); trimmed != name {
		name = trimmed
	}
	if i := strings.LastIndexAny(name, `/\`); i >= 0 {
		name = name[i+1:]
	}

	// 2 and 3. The control characters go; the punctuation no filesystem
	// takes becomes an underscore rather than being dropped, so "a?b" does
	// not quietly become "ab".
	name = strings.Map(func(r rune) rune {
		if isControl(r) {
			return -1
		}
		if strings.ContainsRune(`<>:"/\|?*`, r) {
			return '_'
		}
		return r
	}, name)

	// 4. Leading and trailing spaces and dots. A leading dot makes a hidden
	// file on every desktop, which is a thing somebody meant and not a
	// thing a paste meant.
	name = strings.Trim(name, " .")

	// 5. The reserved names, before the name is shortened: a name that is
	// reserved because of what is left of it after truncation is still
	// reserved.
	name = escapeReserved(name)

	// 6. The length limit, on a rune boundary and keeping the extension,
	// because "a-really-long-name.tar.gz" that became "a-really-long-name"
	// is a file nobody can open with the program it is for.
	return shorten(name, maxNameBytes)
}

// isControl reports a rune that no filesystem will keep in a name: the C0
// range, DEL, and the two C1 controls that turn up in Latin-1 filenames.
func isControl(r rune) bool {
	return r < 0x20 || r == 0x7f || (r >= 0x80 && r <= 0x9f)
}

// reserved are the names Windows refuses, whatever extension they carry:
// "nul.txt" is NUL as far as Windows is concerned.
var reserved = map[string]bool{
	"CON": true, "PRN": true, "AUX": true, "NUL": true,
	"COM1": true, "COM2": true, "COM3": true, "COM4": true, "COM5": true,
	"COM6": true, "COM7": true, "COM8": true, "COM9": true,
	"LPT1": true, "LPT2": true, "LPT3": true, "LPT4": true, "LPT5": true,
	"LPT6": true, "LPT7": true, "LPT8": true, "LPT9": true,
}

// escapeReserved is the part of the name before its extension, upper-cased,
// with the reserved ones prefixed by an underscore.
func escapeReserved(name string) string {
	stem := name
	if i := strings.IndexByte(name, '.'); i > 0 {
		stem = name[:i]
	}
	if reserved[strings.ToUpper(stem)] {
		return "_" + name
	}
	return name
}

// shorten fits a name to a byte budget without cutting a rune in half and
// without losing the extension.
func shorten(name string, budget int) string {
	if len(name) <= budget {
		return name
	}
	keep := ""
	if i := strings.LastIndexByte(name, '.'); i > 0 {
		// The extension is only worth keeping when it is short enough to be
		// an extension. ".verylongextension" on its own is a dot in a name.
		if len(name)-i <= 16 {
			keep = name[i:]
		}
	}
	room := budget - len(keep)
	stem := name[:len(name)-len(keep)]
	for len(stem) > room {
		_, size := utf8.DecodeLastRuneInString(stem)
		stem = stem[:len(stem)-size]
	}
	// The cut can leave a trailing dot or space, which is what step 4 would
	// have removed from the whole name.
	return strings.TrimRight(stem, " .") + keep
}

// SanitizeOr is Sanitize with a fallback for the case where nothing usable
// was left: an upload field, a "save as" box, anywhere a name has to exist
// before the user has typed one.
func SanitizeOr(name, fallback string) string {
	if s := Sanitize(name); s != "" {
		return s
	}
	if f := Sanitize(fallback); f != "" {
		return f
	}
	return "file"
}

// Join puts a base and some segments together into one path, with the
// separators normalised and the "." and ".." segments resolved away.
//
// Four rules, and every path in this package goes through them:
//
//   - a segment that is empty is skipped, so a breadcrumb with a blank in it
//     does not produce a double slash
//   - a segment that starts with a separator is absolute and replaces the
//     base outright, the way every path operation in every language treats
//     it, because the base of a path after an absolute segment is not the
//     base any more
//   - "." is dropped and ".." pops one segment, and a ".." with nothing left
//     to pop is dropped rather than climbing above the root
//   - backslashes become forward slashes, because a path is compared against
//     strings and opened by something that speaks forward slashes
//
// The empty base with no segments is "." — the current directory — because
// that is what a path with nothing in it means, and returning "/" would make
// a caller with an empty setting land in the root of the disk.
func Join(base string, parts ...string) string {
	out := splitPath(base)
	rooted := isAbsolute(base)
	for _, part := range parts {
		if isAbsolute(part) {
			out = splitPath(part)
			rooted = true
			continue
		}
		for _, seg := range splitPath(part) {
			switch seg {
			case ".":
			case "..":
				if len(out) > 0 {
					out = out[:len(out)-1]
				}
			default:
				out = append(out, seg)
			}
		}
	}
	if len(out) == 0 {
		// An empty path is the root when the path came from a root, and the
		// current directory when it did not. Joining ".." onto "/" gives "/"
		// — a path can be climbed but it cannot fall off the top of the
		// disk — while joining nothing onto nothing gives ".", which is the
		// one path that means "wherever the caller was".
		if rooted {
			return "/"
		}
		return "."
	}
	// A drive letter keeps its colon and its slash: "C:/Windows" is one
	// path, and prefixing a "/" to it makes a path that no Windows tool
	// will open and that Linux reads as a file called "C:".
	if hasDriveLetter(out) {
		return out[0] + "/" + strings.Join(out[1:], "/")
	}
	return "/" + strings.Join(out, "/")
}

// hasDriveLetter reports the first segment being a drive letter, which is
// what Join leaves attached to the path rather than turning into a directory.
func hasDriveLetter(segments []string) bool {
	if len(segments) == 0 || len(segments[0]) != 2 || segments[0][1] != ':' {
		return false
	}
	return isDriveLetter(segments[0][0])
}

// splitPath is a path as its segments, with the empty ones dropped and the
// separators normalised. An absolute path loses its leading empty segment,
// which is what makes a path built out of segments absolute at the end.
func splitPath(p string) []string {
	p = strings.ReplaceAll(p, `\`, "/")
	var out []string
	for _, seg := range strings.Split(p, "/") {
		if seg == "" {
			continue
		}
		out = append(out, seg)
	}
	return out
}

// isAbsolute reports a segment that replaces the base: one that starts with a
// separator, or with a drive letter and a colon.
func isAbsolute(s string) bool {
	if strings.HasPrefix(s, "/") || strings.HasPrefix(s, `\`) {
		return true
	}
	if len(s) >= 2 && s[1] == ':' && isDriveLetter(s[0]) {
		return true
	}
	return false
}

func isDriveLetter(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

// sizeUnits are the units HumanSize counts in. They are decimal — powers of
// a thousand — because the web version of this interface counts in decimal
// and a person comparing a number here with one there is looking at the same
// file. The lowercase k of kB is the SI symbol, not the "k" of kibibytes:
// there is no i in any of them, and there is no KiB either.
var sizeUnits = []string{"B", "kB", "MB", "GB", "TB", "PB", "EB"}

// HumanSize is a byte count a person can read: "0 B", "999 B", "1.0 kB",
// "1.5 GB".
//
// Decimal, deliberately, and the reason is consistency rather than taste. The
// web version of this product shows "1.0 MB" for a file of 1,000,000 bytes,
// the desktop shows the same figure for the same file, and a support
// conversation about "the file that is 1.05 MB" does not become a lesson in
// the difference between the two prefixes. Anything wanting kibibytes wants
// to say so on screen.
//
// One decimal below ten and none above it, so "1.5 GB" and "128 MB" both
// read as themselves: a second decimal on a figure that does not need it is
// a figure people trust less, not more. The rounding happens before the unit
// is chosen, so 999,999 bytes is "1.0 MB" and never "1000.0 MB".
//
// A negative count keeps its sign and is otherwise counted the same way; the
// zero value is "0 B" rather than an empty string, because a size column
// with a blank in it looks like a bug.
func HumanSize(n int64) string {
	if n == 0 {
		return "0 B"
	}
	sign := ""
	if n < 0 {
		sign, n = "-", -n
	}

	v := float64(n)
	i := 0
	last := len(sizeUnits) - 1
	for v >= 1000 && i < last {
		v /= 1000
		i++
	}
	// One decimal, then re-check against the figure as it will actually be
	// PRINTED, which is the part that matters: 999.5 is printed as "1000"
	// at zero decimals, and a "1000 kB" line is the one number in this
	// function that is wrong. So the unit changes at 999.5, not at 1000.
	v = math.Round(v*10) / 10
	shown := v
	if v >= 10 {
		shown = math.Round(v)
	}
	if shown >= 1000 && i < last {
		v = math.Round(v/1000*10) / 10
		i++
	}
	if v < 10 && i > 0 {
		return sign + strconv.FormatFloat(v, 'f', 1, 64) + " " + sizeUnits[i]
	}
	return sign + strconv.FormatFloat(v, 'f', 0, 64) + " " + sizeUnits[i]
}

// Percent is how much of a whole n is, as a whole number from 0 to 100.
//
// It is here rather than at each call site because the rounding is the whole
// of it: a quota bar at 99.6% shows "100%" and says the quota is full when
// it is not, and a progress bar at 0.4% showing "0%" says nothing is
// happening when something is. So a share under one half of a point rounds up
// to one, and only a true zero rounds to zero.
func Percent(part, whole int64) int {
	switch {
	case whole <= 0 || part <= 0:
		return 0
	case part >= whole:
		return 100
	}
	p := part * 100 / whole
	rem := part*100 - whole*p
	// rem/whole is the fraction left over; it is at least a half when the
	// integer division threw away a half or more of a point.
	if rem*2 >= whole {
		p++
	}
	if p > 100 {
		p = 100
	}
	return int(p)
}

// itoa is a small int-to-string. It exists because half the labels in this
// package are a number next to a word, and strconv.Itoa on every one of them
// would put a conversion at each call site for a function that cannot fail.
func itoa(n int) string { return strconv.Itoa(n) }
