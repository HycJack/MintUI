package overlay

import (
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/theme"
)

// The tests for the layers the baseline asked for after the original set:
// the backdrop on its own, the surface placed at a point, the dialog's box
// without its backdrop, the popover's roomier relatives, and the menu that a
// caller opens on its own terms. They assert the same way the rest do: the
// words on screen, where the surface sits, and whether a press reached what
// it was meant to.

// ── Overlay ────────────────────────────────────────────────────────────────

// dimmedAt is the page's corner under whatever the view put over it, as one
// number: the three channels summed, so "darker" has no channel to hide in.
func dimmedAt(mode core.Mode, withOverlay bool) int {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: mode})
		darkPage(c)
		if withOverlay {
			Overlay(c, OverlayOptions{})
		}
	}, 900, 600)
	r, g, b := pixelAt(tt, 6, 6)
	return int(r) + int(g) + int(b)
}

// The wash is the same in both appearances: a page behind the overlay is
// darker than the same page without it, in the light and in the dark. A scrim
// drawn from a fixed colour rather than over the tokens would close the gap
// on the dark palette, where the page is already close to black.
func TestOverlayDimsTheWindowInBothAppearances(t *testing.T) {
	for _, tc := range []struct {
		name string
		mode core.Mode
	}{
		{"light", core.Light},
		{"dark", core.Dark},
	} {
		base := dimmedAt(tc.mode, false)
		dimmed := dimmedAt(tc.mode, true)
		if dimmed >= base {
			t.Errorf("%s: the page at (6,6) is %d behind the overlay and %d bare; the wash did not darken it",
				tc.name, dimmed, base)
		}
	}
}

// A clickable backdrop takes the press that lands on it and reports it: that
// is the press a caller reading the backdrop closes its own panel on.
func TestOverlayClickableReportsThePress(t *testing.T) {
	pressed := 0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		page(c)
		if Overlay(c, OverlayOptions{Clickable: true}).Clicked() {
			pressed++
		}
	}, 900, 600)

	tt.ClickAt(450, 300)
	if pressed != 1 {
		t.Errorf("the backdrop reported %d presses for one press on it, want 1", pressed)
	}
}

// A backdrop that does not take presses is only there to dim: the press goes
// on to the window behind it, which is the reason a page keeps working under
// a layer the caller dismisses by other means.
func TestOverlayWithoutClickableLetsPressesThrough(t *testing.T) {
	trigger := 0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		page(c)
		if input.Button(c, "Edit", input.ButtonOptions{Label: "Edit"}).Clicked() {
			trigger++
		}
		Overlay(c, OverlayOptions{})
	}, 900, 600)

	if err := tt.Click("Edit"); err != nil {
		t.Fatal(err)
	}
	if trigger != 1 {
		t.Errorf("the page behind a non-clickable backdrop took %d presses, want 1", trigger)
	}
}

// ── OverlayAnchored ────────────────────────────────────────────────────────

// The surface sits beside its point, with the edge the Side names facing it:
// below puts the top edge a step under the point, above the bottom edge a
// step over it, and the two horizontal sides the same along the other axis.
func TestOverlayAnchoredHangsBesideItsPoint(t *testing.T) {
	const x, y = 120.0, 120.0
	for _, tc := range []struct {
		name string
		side OverlaySide
	}{
		{"below", OverlayBelow},
		{"above", OverlayAbove},
		{"start", OverlayStart},
		{"end", OverlayEnd},
	} {
		surface := ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			k := core.Tokens(c)
			ui.Box(c).Fill().Background(k.Surface)
			OverlayAnchored(c, OverlayAnchoredOptions{
				X: x, Y: y, Side: tc.side, Title: "Mark",
				Body: func() { ui.Text(c, "North seam") },
			})
		}, 400, 400)
		r := find(t, surface, "Mark")

		switch tc.side {
		case OverlayBelow:
			if r.Y < y+3 || r.Y > y+5 {
				t.Errorf("below: the surface top is at %v, want a step under the point at %v", r, y)
			}
			if r.X < x-1 || r.X > x+1 {
				t.Errorf("below: the surface starts at %v, want the point's x at %v", r, x)
			}
		case OverlayAbove:
			if bottom := r.Y + r.H; bottom < y-5 || bottom > y-3 {
				t.Errorf("above: the surface bottom is at %v, want a step over the point at %v", bottom, y)
			}
			if r.X < x-1 || r.X > x+1 {
				t.Errorf("above: the surface starts at %v, want the point's x at %v", r, x)
			}
		case OverlayStart:
			if right := r.X + r.W; right < x-5 || right > x-3 {
				t.Errorf("start: the surface right is at %v, want a step left of the point at %v", right, x)
			}
			if r.Y < y-1 || r.Y > y+1 {
				t.Errorf("start: the surface starts at %v, want the point's y at %v", r, y)
			}
		case OverlayEnd:
			if r.X < x+3 || r.X > x+5 {
				t.Errorf("end: the surface left is at %v, want a step right of the point at %v", r, x)
			}
			if r.Y < y-1 || r.Y > y+1 {
				t.Errorf("end: the surface starts at %v, want the point's y at %v", r, y)
			}
		}
		find(t, surface, "North seam")
	}
}

// The caret is the arrow between the surface and its point: the pixel in the
// gap, on the line from the point to the surface, is the surface's own colour
// — and without the caret it is the page's.
func TestOverlayAnchoredDrawsItsCaretTowardsThePoint(t *testing.T) {
	const x, y = 120.0, 120.0
	render := func(caret bool) (r, g, b uint32) {
		tt := ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			k := core.Tokens(c)
			ui.Box(c).Fill().Background(k.Surface)
			OverlayAnchored(c, OverlayAnchoredOptions{
				X: x, Y: y, Side: OverlayBelow, Caret: caret, Title: "Mark",
				Body: func() { ui.Text(c, "North seam") },
			})
		}, 400, 400)
		// The caret sits in the gap above the surface's top edge; a little
		// down from the apex, at the centre of its width, is inside the
		// triangle whatever the renderer rounds.
		return pixelAt(tt, int(x+6), int(y+2))
	}

	cr, cg, cb := render(true)
	if cr < 0xF6 || cg < 0xF6 || cb < 0xF6 {
		t.Errorf("the gap above the surface is #%02X%02X%02X with the caret; it should be the surface's white", cr, cg, cb)
	}
	pr, pg, pb := render(false)
	if pr > 0xF5 || pg > 0xF5 || pb > 0xF5 {
		t.Errorf("the gap above the surface is #%02X%02X%02X without the caret; it should be the page", pr, pg, pb)
	}
}

// A surface that does not know which edge faces its anchor hangs from nothing,
// and a surface with nothing in it is a shadow with nothing under it.
func TestOverlayAnchoredNeedsASideAndABody(t *testing.T) {
	for name, draw := range map[string]func(c *ui.Context){
		// A body on the no-side case, so the side is what fires.
		"no side": func(c *ui.Context) {
			OverlayAnchored(c, OverlayAnchoredOptions{
				X: 120, Y: 120, Side: OverlaySide(9), Body: func() {},
			})
		},
		"no body": func(c *ui.Context) {
			OverlayAnchored(c, OverlayAnchoredOptions{X: 120, Y: 120, Side: OverlayBelow})
		},
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("an anchored surface with %s should panic", name)
				}
			}()
			ui.NewTester(func(c *ui.Context) {
				core.Use(c, core.Settings{})
				page(c)
				draw(c)
			}, 400, 400)
		}()
	}
}

// ── DialogPanel ────────────────────────────────────────────────────────────

// The box of a dialog is the panel a dialog wears, drawn apart from its
// backdrop: the title, the subtitle and the body in that order, inside one
// surface.
func TestDialogPanelDrawsItsTitleAndBody(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		k := core.Tokens(c)
		ui.Column(c).Padding(24).Gap(12).Children(func() {
			ui.Text(c, "Open callbacks").FontSize(theme.DisplaySize).Bold().TextColor(k.Text)
			DialogPanel(c, ui.Box(c), DialogPanelOptions{
				Title:    "Log callback",
				Subtitle: "It lands in New, assigned and priced as you set it here.",
				Body: func() {
					ui.Text(c, "Customer").Label("field-customer")
				},
			})
		})
	}, 900, 600)

	panel := find(t, tt, "Log callback")
	subtitle := find(t, tt, "It lands in New, assigned and priced as you set it here.")
	body := find(t, tt, "field-customer")

	if subtitle.Y < panel.Y || body.Y < subtitle.Y+subtitle.H {
		t.Errorf("the panel at %v, its subtitle at %v and its body at %v are not stacked in that order",
			panel, subtitle, body)
	}
	for _, r := range []ui.Rect{subtitle, body} {
		if r.X < panel.X || r.X+r.W > panel.X+panel.W {
			t.Errorf("%v is not inside the panel at %v", r, panel)
		}
	}
}

// A dialog's surface with nothing inside it is a shadow with no dialog under
// it, and it is the case worth stopping for.
func TestDialogPanelNeedsABody(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("a dialog panel with no body should panic")
		}
	}()
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		page(c)
		DialogPanel(c, ui.Box(c), DialogPanelOptions{Title: "Log callback"})
	}, 900, 600)
}

// ── PopoverRoom / PopoverWide ──────────────────────────────────────────────

// popoverSurfaceTester hangs one of the popover's relatives off a row, open
// from the start, so the test measures the surface rather than the opening.
func popoverSurfaceTester(t *testing.T, build func(c *ui.Context, anchor *ui.Element)) *ui.Tester {
	t.Helper()
	return ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		k := core.Tokens(c)
		ui.Column(c).Fill().Background(k.Surface).Padding(24).Gap(16).Children(func() {
			row := ui.Row(c).Grow(0).Radius(theme.ControlRadius).Background(k.Background).
				Padding(8, 12).Label("CB-2871").Children(func() {
				ui.Text(c, "CB-2871").FontSize(theme.BodySize).TextColor(k.Text)
			})
			build(c, row)
		})
	}, 900, 600)
}

// The room surface is the point of PopoverRoom: the same one-line content
// takes more room from the surface's top to its body than under the compact
// surface a Popover wears, because the padding is the room and the title is
// the sheet's size.
func TestPopoverRoomWearsTheRoomSurface(t *testing.T) {
	measure := func(build func(c *ui.Context, anchor *ui.Element)) (top, body float32) {
		tt := popoverSurfaceTester(t, build)
		return find(t, tt, "Riverside Clinic").Y, find(t, tt, "Opened by Dana Reyes").Y
	}
	compactTop, compactBody := measure(func(c *ui.Context, anchor *ui.Element) {
		open := true
		Popover(c, anchor, &open, PopoverOptions{
			Title: "Riverside Clinic",
			Body:  func() { ui.Text(c, "Opened by Dana Reyes") },
		})
	})
	roomTop, roomBody := measure(func(c *ui.Context, anchor *ui.Element) {
		open := true
		PopoverRoom(c, anchor, &open, PopoverRoomOptions{
			Title: "Riverside Clinic",
			Body:  func() { ui.Text(c, "Opened by Dana Reyes") },
		})
	})
	if roomBody-roomTop <= compactBody-compactTop+4 {
		t.Errorf("the room surface takes %v from its top to its body and the compact one %v; "+
			"the room is the point of it", roomBody-roomTop, compactBody-compactTop)
	}
}

// PopoverWide asks for the full width rather than the content's: the same
// one-line content spans the window under it, where under a Popover it takes
// what it needs.
func TestPopoverWideFillsTheWidth(t *testing.T) {
	measure := func(build func(c *ui.Context, anchor *ui.Element)) float32 {
		tt := popoverSurfaceTester(t, build)
		return find(t, tt, "Log callback").W
	}
	compact := measure(func(c *ui.Context, anchor *ui.Element) {
		open := true
		Popover(c, anchor, &open, PopoverOptions{Title: "Log callback"})
	})
	wide := measure(func(c *ui.Context, anchor *ui.Element) {
		open := true
		PopoverWide(c, anchor, &open, PopoverWideOptions{Title: "Log callback"})
	})
	if wide < 800 {
		t.Errorf("the wide popover is %v wide in a 900pt window; it was asked to fill it", wide)
	}
	if wide < compact*2 {
		t.Errorf("the wide popover at %v is not much wider than the content popover at %v; the width is the difference",
			wide, compact)
	}
}

// ── PopupMenuOpen ──────────────────────────────────────────────────────────

// menuTester puts a row in the page with its menu open beside it, and hands
// the test the menu's result read the way a caller reads it: inside the view,
// where the pass that consumed the press still has it.
func menuTester(t *testing.T, items []PopupMenuItem, chosen *int) *ui.Tester {
	t.Helper()
	open := true
	return ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		k := core.Tokens(c)
		ui.Column(c).Fill().Background(k.Surface).Padding(24).Gap(16).Children(func() {
			anchor := ui.Row(c).Grow(0).Radius(theme.ControlRadius).Background(k.Background).
				Padding(8, 12).Label("CB-2871").Children(func() {
				ui.Text(c, "CB-2871").FontSize(theme.BodySize).TextColor(k.Text)
			})
			res := PopupMenuOpen(c, PopupMenuOpenOptions{
				Anchor: anchor, Open: &open, Items: items,
			})
			if i := res.Chosen(); i >= 0 {
				*chosen = i
			}
		})
	}, 900, 600)
}

func TestPopupMenuOpenDrawsItsRows(t *testing.T) {
	chosen := -1
	tt := menuTester(t, []PopupMenuItem{
		{Label: "Rename"},
		{Label: "Duplicate", Disabled: true},
		{Label: "Delete", Destructive: true},
	}, &chosen)
	for _, want := range []string{"Rename", "Duplicate", "Delete"} {
		wantText(t, tt, want)
	}
}

// The press lands on one row and only that row reports: the index is what the
// caller acts on, and the disabled row takes no press at all.
func TestPopupMenuOpenReportsTheRowPressed(t *testing.T) {
	chosen := -1
	tt := menuTester(t, []PopupMenuItem{
		{Label: "Rename"},
		{Label: "Duplicate", Disabled: true},
		{Label: "Delete", Destructive: true},
	}, &chosen)
	if err := tt.Click("Delete"); err != nil {
		t.Fatal(err)
	}
	if chosen != 2 {
		t.Errorf("pressing Delete reported row %d, want 2", chosen)
	}

	chosen2 := -1
	tt2 := menuTester(t, []PopupMenuItem{
		{Label: "Rename"},
		{Label: "Duplicate", Disabled: true},
	}, &chosen2)
	if err := tt2.Click("Duplicate"); err != nil {
		t.Fatal(err)
	}
	if chosen2 != -1 {
		t.Errorf("pressing the disabled row reported %d, want no report", chosen2)
	}
}

// An empty menu is a panel with nothing to choose, and it is the case worth
// stopping for.
func TestPopupMenuOpenNeedsItems(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("a popup menu with no items should panic")
		}
	}()
	open := true
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		page(c)
		PopupMenuOpen(c, PopupMenuOpenOptions{
			Anchor: ui.Box(c), Open: &open,
		})
	}, 900, 600)
}

// ── both appearances ───────────────────────────────────────────────────────

// The new layers are drawn from the window's palette, like the rest: the same
// surface in the dark appearance is near-black, not white with the dark page
// painted over it.
func TestDialogPanelInDark(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: core.Dark})
		DialogPanel(c, ui.Box(c), DialogPanelOptions{
			Title: "dm.dialog.title",
			Body:  func() { ui.Text(c, "dm.dialog.body") },
		})
	}, 480, 320)
	wantText(t, tt, "dm.dialog.title")
	wantText(t, tt, "dm.dialog.body")
	panel := find(t, tt, "dm.dialog.title")
	r, g, b := pixelAt(tt, int(panel.X+3), int(panel.Y+3))
	if r > 0x30 || g > 0x30 || b > 0x30 {
		t.Errorf("the dark panel is #%02X%02X%02X; it should be near-black", r, g, b)
	}
}

func TestOverlayAnchoredInDark(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: core.Dark})
		k := core.Tokens(c)
		ui.Box(c).Fill().Background(k.Surface)
		OverlayAnchored(c, OverlayAnchoredOptions{
			X: 120, Y: 120, Side: OverlayBelow, Caret: true, Title: "dm.anchored",
			Body: func() { ui.Text(c, "dm.anchored.body") },
		})
	}, 400, 400)
	wantText(t, tt, "dm.anchored")
	wantText(t, tt, "dm.anchored.body")
	panel := find(t, tt, "dm.anchored")
	r, g, b := pixelAt(tt, int(panel.X+3), int(panel.Y+3))
	if r > 0x30 || g > 0x30 || b > 0x30 {
		t.Errorf("the dark anchored surface is #%02X%02X%02X; it should be near-black", r, g, b)
	}
}

func TestPopoverRoomInDark(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: core.Dark})
		k := core.Tokens(c)
		ui.Column(c).Fill().Background(k.Surface).Padding(24).Gap(16).Children(func() {
			row := ui.Row(c).Grow(0).Radius(theme.ControlRadius).Background(k.Background).
				Padding(8, 12).Label("CB-2871").Children(func() {
				ui.Text(c, "CB-2871")
			})
			open := true
			PopoverRoom(c, row, &open, PopoverRoomOptions{
				Title: "dm.room.title",
				Body:  func() { ui.Text(c, "dm.room.body") },
			})
		})
	}, 900, 600)
	wantText(t, tt, "dm.room.title")
	wantText(t, tt, "dm.room.body")
}

func TestPopoverWideInDark(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: core.Dark})
		k := core.Tokens(c)
		ui.Column(c).Fill().Background(k.Surface).Padding(24).Gap(16).Children(func() {
			row := ui.Row(c).Grow(0).Radius(theme.ControlRadius).Background(k.Background).
				Padding(8, 12).Label("CB-2871").Children(func() {
				ui.Text(c, "CB-2871")
			})
			open := true
			PopoverWide(c, row, &open, PopoverWideOptions{
				Title: "dm.wide.title",
				Body:  func() { ui.Text(c, "dm.wide.body") },
			})
		})
	}, 900, 600)
	wantText(t, tt, "dm.wide.title")
	wantText(t, tt, "dm.wide.body")
}

func TestPopupMenuOpenInDark(t *testing.T) {
	open := true
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: core.Dark})
		k := core.Tokens(c)
		ui.Column(c).Fill().Background(k.Surface).Padding(24).Gap(16).Children(func() {
			anchor := ui.Row(c).Grow(0).Radius(theme.ControlRadius).Background(k.Background).
				Padding(8, 12).Label("CB-2871").Children(func() {
				ui.Text(c, "CB-2871")
			})
			PopupMenuOpen(c, PopupMenuOpenOptions{
				Anchor: anchor, Open: &open,
				Items: []PopupMenuItem{{Label: "dm.menu.rename"}, {Label: "dm.menu.delete", Destructive: true}},
			})
		})
	}, 900, 600)
	wantText(t, tt, "dm.menu.rename")
	wantText(t, tt, "dm.menu.delete")
}
