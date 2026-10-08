package media

import (
	"fmt"
	"strconv"
	"time"
)

// TimeToText writes a position or a length the way a transport reads it.
//
// The format is m:ss below an hour and h:mm:ss at or above it:
//
//	0        → "0:00"
//	59s      → "0:59"
//	60s      → "1:00"
//	61s      → "1:01"
//	3600s    → "1:00:00"
//	3661s    → "1:01:01"
//
// The seconds are always two digits, which is the part that matters: a
// transport whose seconds field is one digit wide makes the bar jump by ten
// columns as it counts, and a reader counting along with it loses their place
// every time. Minutes are not padded, because "1:00:05" is a video an hour
// and five seconds long and "01:00:05" reads as a stopwatch instead.
//
// Negative input reads as zero. A playhead a fraction of a second behind the
// start is a playhead at the start, and "-0:00" on a scrubber would be a
// transport reporting an error the caller has to know to ignore.
func TimeToText(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	// Round to the nearest second rather than truncating, so a playhead at
	// 59.7s does not sit on "0:59" for a whole second after it has already
	// passed 59.5s. A transport that is a second behind is worse than one
	// that rounds early.
	total := int(d.Round(time.Second) / time.Second)
	h, m, sec := total/3600, total/60%60, total%60
	if h > 0 {
		return strconv.Itoa(h) + ":" + pad2(m) + ":" + pad2(sec)
	}
	return strconv.Itoa(m) + ":" + pad2(sec)
}

func pad2(n int) string {
	if n < 10 {
		return "0" + strconv.Itoa(n)
	}
	return strconv.Itoa(n)
}

// ScrubRatio is where a position sits in a length, as a share from 0 to 1.
//
// It is the one number every transport on screen needs — how far along the
// played part is filled, how wide the thumb is, where a marker belongs — and
// it is here rather than at each of those sites because the two ends are the
// interesting ones. A length of zero returns 0 rather than dividing by it, and
// both ends are clamped: a position past the end comes back from a seek that
// landed in a file which turned out to be shorter, and a scrubber that drew
// its fill past its own rail would be reporting a time the video never had.
func ScrubRatio(at, total time.Duration) float32 {
	if total <= 0 || at <= 0 {
		return 0
	}
	r := float32(at) / float32(total)
	if r > 1 {
		return 1
	}
	return r
}

// fmtDuration is a Duration written for a table cell, where there is no room
// for the transport's leading field: "0:59" reads, "00:00:59" does not.
func fmtDuration(d time.Duration) string { return TimeToText(d) }

// labelAt names a position for assistive technology, which cannot hear a bare
// "1:01" and tell a position from a length.
func labelAt(what string, d time.Duration) string {
	return fmt.Sprintf("%s %s", what, TimeToText(d))
}
