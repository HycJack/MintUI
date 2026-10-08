package feedback_test

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/feedback"
)

// Example shows the two things an application says rather than shows: why
// there is nothing here, and what just happened.
func Example() {
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		u := core.Density(c).Unit()

		ui.Column(c).Padding(u * 6).Children(func() {
			// Empty reports the press of its own action, so the surrounding
			// layout never has to know what absence looks like.
			if feedback.Empty(c, feedback.EmptyOptions{
				Title:  "Nothing resolved yet today",
				Body:   "Callbacks land here once the customer confirms the fix.",
				Action: "Resolve a callback",
			}).Pressed() {
				ui.Text(c, "opening the picker…")
			}

			// Toast withdraws itself when the message is empty.
			feedback.Toast(c, "Saved", u)
		})
	}, 600, 400)

	// Output:
}
