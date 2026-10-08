package chart

import (
	"math"
	"strconv"
	"strings"
)

// Format renders a value as the label of one tick. step is the distance
// between two ticks, which a format uses to decide how many decimals its
// values need; the ones that do not care ignore it.
//
// The formats here are the ones a chart of ordinary numbers needs: [Auto] for
// most axes, [Compact] where the numbers are large enough to crowd the gutter,
// [Percent] for rates, [Fixed] where the caller knows better.
type Format func(v, step float64) string

// Auto is the format an axis uses when none is given: as many decimals as the
// step itself takes, with the trailing zeroes dropped — so an axis stepped by
// 2.5 reads 0, 2.5, 5, 7.5, 10 rather than 0.0, 2.5, 5.0, 7.5, 10.0, and a
// step of 200 reads 1,200 rather than 1,200.000000.
func Auto() Format {
	return func(v, step float64) string {
		return tidy(Group(v, decimalsFor(step)))
	}
}

// Fixed is a format with a fixed number of decimals, for an axis whose
// values mean the same without them — a rate, a temperature, a price. Nothing
// is trimmed: "3.50" is what a price is written as.
func Fixed(decimals int) Format {
	if decimals < 0 {
		decimals = 0
	}
	return func(v, _ float64) string { return Group(v, decimals) }
}

// Compact is a format that shortens the large numbers — 12,400 reads "12.4k"
// — for an axis whose labels would otherwise run into each other.
func Compact() Format {
	return func(v, step float64) string {
		switch a := math.Abs(v); {
		case a >= 1e9:
			return short(v/1e9) + "B"
		case a >= 1e6:
			return short(v/1e6) + "M"
		case a >= 1e3:
			return short(v/1e3) + "k"
		}
		return tidy(Group(v, decimalsFor(step)))
	}
}

// Percent is a format for values that are fractions: 0.156 reads "15.6%".
// The step is scaled with them, so a percent axis keeps the same decimals as
// the fractions it shows.
func Percent() Format {
	return func(v, step float64) string {
		return tidy(Group(v*100, decimalsFor(step*100))) + "%"
	}
}

// tidy drops the zeroes a value does not need after its point: 5.0 is 5, and
// 1,200.000000 is 1,200. It never touches the digits before the point, which
// are the ones carrying the commas.
func tidy(s string) string {
	intPart, frac, hasFrac := strings.Cut(s, ".")
	if !hasFrac {
		return s
	}
	frac = strings.TrimRight(frac, "0")
	if frac == "" {
		return intPart
	}
	return intPart + "." + frac
}

// Group writes v with exactly decimals decimals and a thousands separator,
// the way the interface writes every other number: 12,400.5 rather than
// 12400.50000. A value that rounds to nothing negative is written as 0.
func Group(v float64, decimals int) string {
	if decimals < 0 {
		decimals = 0
	}
	if math.IsNaN(v) {
		return "—"
	}
	if math.IsInf(v, 0) {
		if v < 0 {
			return "-∞"
		}
		return "∞"
	}
	s := strconv.FormatFloat(v, 'f', decimals, 64)
	sign, body := "", s
	if strings.HasPrefix(body, "-") {
		sign, body = "-", body[1:]
	}
	intPart, frac, hasFrac := strings.Cut(body, ".")
	if grouped := group(intPart); grouped != intPart {
		body = grouped
		if hasFrac {
			body += "." + frac
		}
	}
	if sign == "-" && allZero(body) {
		sign = ""
	}
	return sign + body
}

// group inserts a comma every three digits of a run of digits.
func group(digits string) string {
	if len(digits) <= 3 {
		return digits
	}
	var b strings.Builder
	head := len(digits) % 3
	if head > 0 {
		b.WriteString(digits[:head])
	}
	for i := head; i < len(digits); i += 3 {
		if b.Len() > 0 {
			b.WriteByte(',')
		}
		b.WriteString(digits[i : i+3])
	}
	return b.String()
}

func allZero(s string) bool {
	for _, r := range s {
		if r != '0' && r != '.' && r != ',' {
			return false
		}
	}
	return true
}

// short writes a shortened number with at most one decimal and none at all
// when it is whole: 1.5, 12, not 12.0.
func short(v float64) string {
	s := strconv.FormatFloat(v, 'f', 1, 64)
	return strings.TrimSuffix(s, ".0")
}

// decimalsFor is how many decimals a step takes to write: 2.5 wants one, 0.05
// two, 200 none. Reading it off the step rather than off the value is what
// keeps an axis's labels the same width as each other.
func decimalsFor(step float64) int {
	if step == 0 || math.IsNaN(step) || math.IsInf(step, 0) {
		return 0
	}
	s := strconv.FormatFloat(math.Abs(step), 'f', -1, 64)
	if _, frac, ok := strings.Cut(s, "."); ok {
		return min(len(frac), 6)
	}
	return 0
}
