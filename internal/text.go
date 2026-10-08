package internal

import (
	"strings"
	"unicode"
)

// Initials takes up to two letters from a name: the first letter of each of
// its first two words, so "Ada Lovelace" reads AL and "Prince" reads P.
func Initials(name string) string {
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

// Plural picks one of two words by count, so a count never reads "1 people".
func Plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

// Commas formats n with thousands separators, the way money and counts read
// in this interface.
func Commas(n int) string {
	s := itoa(n)
	if len(s) <= 3 {
		return s
	}
	var out []byte
	for i := range len(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, s[i])
	}
	return string(out)
}

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
