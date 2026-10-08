package overlay_test

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/overlay"
	"github.com/HycJack/MintUI/ui/theme"
)

// Example shows the four ways a window puts something on top of itself: a
// dialog over the page, a sheet from the edge, a popover and a confirm under a
// control, and a tip on the control that needs one.
//
// Nothing here holds its own open state. Each *bool belongs to the caller, and
// every way of closing a layer writes to it — which is why the example reads
// as one piece of app state rather than as four widgets with hidden copies of
// the same question.
func Example() {
	var editing, sheet, explaining, asking bool

	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		k, u := core.Tokens(c), core.Density(c).Unit()

		ui.Column(c).Padding(u * 6).Gap(u * 2).Children(func() {
			// A dialog: modal, so the scrim covers the window and Escape
			// writes false into editing itself.
			if input.Button(c, "Edit", input.ButtonOptions{Label: "Edit"}).Clicked() {
				editing = true
			}
			if editing {
				overlay.Dialog(c, &editing, overlay.DialogOptions{
					Title: "Log callback",
					Body: func() {
						ui.Text(c, "Riverside Clinic").TextColor(k.TextMuted)
					},
					Actions: func() {
						if input.Button(c, "Save", input.ButtonOptions{Primary: true}).
							Clicked() {
							editing = false
						}
					},
				})
			}

			// A sheet: the same dialog hung from the right edge.
			if input.Button(c, "Open log", input.ButtonOptions{Label: "Open log"}).
				Clicked() {
				sheet = true
			}
			if sheet {
				overlay.Drawer(c, &sheet, overlay.DrawerOptions{
					Side:  ui.End,
					Title: "Log callback",
					Body:  func() { ui.Text(c, "It lands in New, assigned and priced here.") },
				})
			}

			// A popover and a confirm: both hung off a row, and both non-modal
			// because the page they hang over is the page being worked in.
			row := ui.Row(c).Grow(0).Radius(theme.ControlRadius).Background(k.Surface).
				Padding(u*2, u*3).Label("CB-2871").Children(func() {
				ui.Text(c, "CB-2871").FontSize(theme.BodySize)
			})
			if row.Clicked() {
				explaining = true
			}
			overlay.Popover(c, row, &explaining, overlay.PopoverOptions{
				Modal: false,
				Title: "Opened by",
				Body:  func() { ui.Text(c, "Dana Reyes") },
			})

			overlay.Popconfirm(c, row, &asking, overlay.PopconfirmOptions{
				Modal:       false,
				Title:       "Delete CB-2871?",
				Body:        "The callback and its history go with it.",
				Confirm:     "Delete",
				Cancel:      "Keep",
				Destructive: true,
			})

			// A tip, for the control whose own label cannot say what it does.
			save := input.Button(c, "Save", input.ButtonOptions{Label: "Save"})
			overlay.Tooltip(c, save, "Saves the callback and clears the form")
		})
	}, 640, 480)

	// Output:
}
