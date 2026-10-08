package theme_test

import (
	"fmt"
	"math"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/theme"
)

// Example shows what a palette is: a flat set of colours with no mode of its
// own, so a window can mix a dark background with a custom fill without a flag
// to keep in step. Every package in the library measures itself from here.
func Example() {
	k := theme.Light()
	d := theme.Dark()

	fmt.Println("text on surface:", readable(k.Text, k.Surface))
	fmt.Println("same in dark:", readable(d.Text, d.Surface))
	fmt.Println("a border is quieter than its text:",
		contrast(k.Border, k.Surface) < contrast(k.Text, k.Surface))
	fmt.Println("one spacing step, in DIPs:", theme.Compact.Unit(), theme.Comfortable.Unit())
	fmt.Println("a lane keeps its width:", theme.ColumnWidth)

	// Output:
	// text on surface: true
	// same in dark: true
	// a border is quieter than its text: true
	// one spacing step, in DIPs: 4 6
	// a lane keeps its width: 272
}

// readable is the WCAG body-text floor, stated as a question.
func readable(fg, bg ui.Color) bool { return contrast(fg, bg) >= 4.5 }

func lum(c ui.Color) float64 {
	f := func(v uint8) float64 {
		x := float64(v) / 255
		if x <= 0.03928 {
			return x / 12.92
		}
		return math.Pow((x+0.055)/1.055, 2.4)
	}
	return 0.2126*f(c.R) + 0.7152*f(c.G) + 0.0722*f(c.B)
}

func contrast(a, b ui.Color) float64 {
	la, lb := lum(a), lum(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}
