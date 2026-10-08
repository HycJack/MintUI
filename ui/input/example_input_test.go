package input_test

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/input"
)

// Example shows the shape every control shares: the value belongs to the
// caller, and the press comes back as a method. There is no OnChange to keep
// in step with the variable that already holds the answer.
func Example() {
	view := 0 // Board
	query := ""

	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		u := core.Density(c).Unit()

		ui.Row(c).Gap(u * 3).Padding(u * 4).Children(func() {
			if input.Button(c, "Log callback", input.ButtonOptions{Primary: true}).Clicked() {
				query = ""
			}
			input.Segmented(c, &view, "Board", "List")
			input.SearchField(c, &query, "Search callbacks")
		})
	}, 900, 160)

	// Output:
}
