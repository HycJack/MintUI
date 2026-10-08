package showcase

import (
	"github.com/egoist/mygo/ui"
)

// State returns a pointer that outlives the frame, hung on the window's own
// root element under key.
//
// Every gallery demo hands its component a pointer, and a pointer into a
// frame-local is a click the next frame undoes — a tab that will not switch,
// a dropdown that snaps shut, a slider that springs back. The key is global:
// prefix it with the page and the demo.
func State[T any](c *ui.Context, key string, init T) *T {
	return ui.Local(c.Root(), "gallery/"+key, func() T { return init })
}
