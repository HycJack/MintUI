// Package ui is a component library for the Callbacks interface: the parts of
// the window that are not about callbacks.
//
// # Using it
//
// A view calls [core.Use] once, at the top, then assembles components from
// the domain packages. Every component takes a [github.com/egoist/mygo/ui.Context],
// takes its arguments as an XxxOptions struct, and returns a value whose
// methods report what the user just did:
//
//	core.Use(c, core.Settings{})
//	ui.Row(c).Gap(u).Children(func() {
//		if Button(c, "Save", ButtonOptions{Primary: true}).Clicked() {
//			save()
//		}
//	})
//
// # The packages
//
//   - [callbacks/ui/theme] holds the palettes, the density scale and the
//     measurements. It draws nothing and imports nothing but MyGo.
//   - [callbacks/ui/core] resolves one window's theme and density, and is the
//     only package that talks to the system appearance.
//   - [callbacks/ui/display] draws marks that carry meaning on their own:
//     initials, the priority meter.
//   - [callbacks/ui/input] takes presses and typing.
//   - [callbacks/ui/navigation] is the rail and the filter lists.
//   - [callbacks/ui/layout] is the page frame and the board's lanes.
//   - [callbacks/ui/data] is the card and the stat line.
//   - [callbacks/ui/feedback] is what an application says rather than shows:
//     the empty state and the toast.
//
// # Conventions
//
// Every frame starts with core.Use. Every component begins its work with
// core.Tokens and core.Density. A component that is given something it cannot
// use — an unnamed button, a selection past the last label, a theme nobody set
// — panics rather than guessing, because a silent wrong frame is harder to
// find than a stopped one.
//
// Components own no state. A selection belongs to the caller as a pointer, and
// a press comes back as a method on the result, so a view can be read top to
// bottom without a hidden second copy of what the user chose.
package ui
