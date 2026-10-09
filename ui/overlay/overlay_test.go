package overlay

import (
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/theme"
)

// The tests here assert what a person would see: the words on the panel, where
// the panel sits, and whether a press or a key closed it. Nothing reaches into
// an element's fields, because a test that does only proves the fields.

// page is the window a layer is opened over: a grey surface with a heading and
// a row, so that anything drawn on top of it can be told apart from it.
//
// The page is Surface rather than Background on purpose. A panel is Background
// white in the light palette, so over a white window it is only told apart by
// its hairline — and a test that cannot see the panel cannot test it.
func page(c *ui.Context) {
	k := core.Tokens(c)
	ui.Column(c).Fill().Background(k.Surface).Padding(24).Gap(12).Children(func() {
		ui.Text(c, "Open callbacks").FontSize(theme.DisplaySize).Bold().TextColor(k.Text)
		ui.Text(c, "28 open across 4 columns").FontSize(theme.RowSize).TextColor(k.TextMuted)
	})
}

// find fails the test when the window shows nothing by that name.
func find(t *testing.T, tt *ui.Tester, name string) ui.Rect {
	t.Helper()
	r, ok := tt.Find(name)
	if !ok {
		t.Fatalf("nothing shows %q; texts %v", name, tt.Texts())
	}
	return r
}

// wantText fails when the window does not show s.
func wantText(t *testing.T, tt *ui.Tester, s string) {
	t.Helper()
	if !tt.HasText(s) {
		t.Fatalf("the window does not show %q; texts %v", s, tt.Texts())
	}
}

// wantNoText fails when the window does show s.
func wantNoText(t *testing.T, tt *ui.Tester, s string) {
	t.Helper()
	if tt.HasText(s) {
		t.Errorf("the window still shows %q; texts %v", s, tt.Texts())
	}
}

// pixelAt is one DIP's worth of the rendered frame, which is how a test says
// something about colour without saying anything about a token.
func pixelAt(tt *ui.Tester, x, y int) (r, g, b uint32) {
	img := tt.Image()
	c := img.RGBAAt(min(max(x, 0), img.Bounds().Dx()-1), min(max(y, 0), img.Bounds().Dy()-1))
	return uint32(c.R), uint32(c.G), uint32(c.B)
}

// ── Panel ──────────────────────────────────────────────────────────────────

// Panel is the shell the rest of the package derives from, so the first thing
// to hold is that its title, its subtitle and its body come out in that order
// and inside one panel, rather than as three elements loose on the window.
func TestPanelDrawsItsTitleSubtitleAndBody(t *testing.T) {
	// The fixture is a column sized by its content rather than one filling
	// the window: a panel's own top edge is what "under the header" is
	// measured from, and a panel dropped into the root of a full window has
	// no top edge of its own to measure from.
	panelOf := func(opts PanelOptions) *ui.Tester {
		return ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			k := core.Tokens(c)
			ui.Column(c).Padding(24).Gap(12).Children(func() {
				ui.Text(c, "Open callbacks").FontSize(theme.DisplaySize).Bold().TextColor(k.Text)
				Panel(c, ui.Box(c), opts, func() {
					ui.Text(c, "Customer").Label("field-customer")
				})
			})
		}, 900, 600)
	}

	tall := panelOf(PanelOptions{
		Title:    "Log callback",
		Subtitle: "It lands in New, assigned and priced as you set it here.",
	})
	panel := find(t, tall, "Log callback")
	subtitle := find(t, tall, "It lands in New, assigned and priced as you set it here.")
	body := find(t, tall, "field-customer")

	if subtitle.Y < panel.Y || body.Y < subtitle.Y+subtitle.H {
		t.Errorf("the panel at %v, its subtitle at %v and its body at %v are not stacked in that order",
			panel, subtitle, body)
	}
	for _, r := range []ui.Rect{subtitle, body} {
		if r.X < panel.X || r.X+r.W > panel.X+panel.W {
			t.Errorf("%v is not inside the panel at %v", r, panel)
		}
	}

	// The same panel without a header starts its body higher: a title that
	// takes no room is not a title.
	flat := panelOf(PanelOptions{})
	if bare := find(t, flat, "field-customer"); bare.Y >= body.Y {
		t.Errorf("the body sits at %v without a heading and %v under one; the header takes no room",
			bare.Y, body.Y)
	}
}

// A panel with no title is a panel of content alone — a tip, a bare popover —
// and must not leave an empty header behind it.
func TestPanelWithoutATitleHasNoHeader(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		page(c)
		Panel(c, ui.Box(c), PanelOptions{}, func() {
			ui.Text(c, "Dana Reyes")
		})
	}, 900, 600)

	wantText(t, tt, "Dana Reyes")
	if got := len(tt.Texts()); got != 3 {
		t.Errorf("a panelless panel drew %d texts, want the page's two and its own one: %v",
			got, tt.Texts())
	}
}

// ── Dialog ─────────────────────────────────────────────────────────────────

// dialogTester opens a dialog from a button in the page, so the test can watch
// the same *bool a caller would.
func dialogTester(t *testing.T, width, height int, nonModal bool) (*ui.Tester, *bool, *int) {
	t.Helper()
	open, edits := new(bool), new(int)
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		page(c)
		// The trigger is named apart from the dialog it opens: two elements
		// answering to one name send a click to whichever was built first,
		// which is not the thing the test means.
		if input.Button(c, "Edit", input.ButtonOptions{Label: "Edit"}).
			Clicked() {
			*open = true
			*edits++
		}
		if *open {
			Dialog(c, open, DialogOptions{
				Title:    "Edit callback",
				Subtitle: "Riverside Clinic",
				Rule:     true,
				NonModal: nonModal,
				Body: func() {
					ui.Text(c, "Customer").Label("field-customer")
				},
				Actions: func() {
					if input.Button(c, "Save", input.ButtonOptions{Primary: true, Label: "Save"}).
						Clicked() {
						*open = false
					}
				},
			})
		}
	}, width, height)
	return tt, open, edits
}

// A closed layer draws nothing at all: not an empty panel, not a scrim. A
// dialog that was built to be closed would still take part in the layout.
func TestDialogClosedDrawsNothing(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		page(c)
		Dialog(c, new(bool), DialogOptions{
			Title: "Edit callback",
			Body:  func() { ui.Text(c, "Customer") },
		})
	}, 900, 600)

	wantNoText(t, tt, "Edit callback")
	if _, ok := tt.Find("Customer"); ok {
		t.Errorf("a closed dialog drew its body; texts %v", tt.Texts())
	}
}

func TestDialogOpenDrawsItsBodyAndActions(t *testing.T) {
	tt, _, _ := dialogTester(t, 900, 600, false)
	if err := tt.Click("Edit"); err != nil {
		t.Fatal(err)
	}

	wantText(t, tt, "Riverside Clinic")
	find(t, tt, "field-customer")
	find(t, tt, "Save")
}

// A modal dialog covers the window: the page behind takes no presses at all,
// and the panel sits in the middle of what is left of the window.
func TestDialogCoversTheWindow(t *testing.T) {
	tt, open, edits := dialogTester(t, 900, 600, false)
	if err := tt.Click("Edit"); err != nil {
		t.Fatal(err)
	}
	if *edits != 1 {
		t.Fatalf("the trigger reported %d presses, want 1", *edits)
	}

	// The panel sits in the middle of what is left of the window.
	panel := find(t, tt, "Edit callback")
	if left, right := panel.X, 900-(panel.X+panel.W); left < right-1 || right < left-1 {
		t.Errorf("the panel is %v from the left and %v from the right; it is not centred", left, right)
	}
	if panel.Y < 0 || panel.Y+panel.H > 600 {
		t.Errorf("the panel at %v does not fit the 900x600 window", panel)
	}

	// A press over the page's own trigger reaches the scrim instead: it
	// closes the dialog, and it does not press what is underneath it.
	if err := tt.Click("Edit"); err != nil {
		t.Fatal(err)
	}
	if *edits != 1 {
		t.Errorf("the page behind the dialog took a press: %d triggers fired, want 1", *edits)
	}
	if *open {
		t.Error("the dialog stayed open for a press that landed on the scrim")
	}
}

func TestDialogEscapeCloses(t *testing.T) {
	tt, open, _ := dialogTester(t, 900, 600, false)
	if err := tt.Click("Edit"); err != nil {
		t.Fatal(err)
	}
	wantText(t, tt, "Riverside Clinic")

	tt.Key(0, ui.KeyEscape)
	if *open {
		t.Error("Escape left the dialog open")
	}
	wantNoText(t, tt, "Riverside Clinic")
}

func TestDialogScrimClickCloses(t *testing.T) {
	tt, open, _ := dialogTester(t, 900, 600, false)
	if err := tt.Click("Edit"); err != nil {
		t.Fatal(err)
	}
	wantText(t, tt, "Riverside Clinic")

	// The corner is scrim in every window wide enough to hold a dialog.
	tt.ClickAt(6, 6)
	if *open {
		t.Error("a press on the scrim left the dialog open")
	}
	wantNoText(t, tt, "Riverside Clinic")
}

// A press inside the panel is the panel's, not the scrim's: clicking on the
// dialog's own body must not close it.
func TestDialogPressInsideThePanelDoesNotClose(t *testing.T) {
	tt, open, _ := dialogTester(t, 900, 600, false)
	if err := tt.Click("Edit"); err != nil {
		t.Fatal(err)
	}
	panel := find(t, tt, "Edit callback")

	tt.ClickAt(panel.X+panel.W/2, panel.Y+panel.H-8)
	if !*open {
		t.Error("a press on the dialog's own surface closed it")
	}
}

// NonModal is the documented way out: the dialog does not swallow the window,
// so Escape no longer closes it and the page behind stays readable.
func TestNonModalDialogIgnoresEscape(t *testing.T) {
	tt, open, _ := dialogTester(t, 900, 600, true)
	if err := tt.Click("Edit"); err != nil {
		t.Fatal(err)
	}
	wantText(t, tt, "Riverside Clinic")

	tt.Key(0, ui.KeyEscape)
	if !*open {
		t.Error("Escape closed a dialog that was told not to be modal")
	}
}

// NonModal says the window behind stays live, so the press on the page is
// the page's. That is the whole of what the option promises, and it is the
// only part of a layer that can be told not to take the window, so it is the
// part worth pressing on.
func TestNonModalDialogLeavesThePageBehindItClickable(t *testing.T) {
	tt, open, edits := dialogTester(t, 900, 600, true)
	if err := tt.Click("Edit"); err != nil {
		t.Fatal(err)
	}
	wantText(t, tt, "Riverside Clinic")

	// The trigger is above the panel and clear of it, so the press is on the
	// page and nothing else: the dialog neither takes it nor closes under it.
	if err := tt.Click("Edit"); err != nil {
		t.Fatal(err)
	}
	if *edits != 2 {
		t.Errorf("the page behind a NonModal dialog took %d presses, want 2", *edits)
	}
	if !*open {
		t.Error("a press on the page behind closed a dialog told not to be modal")
	}
}

func TestDialogNeedsSomethingToShow(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("a dialog with no body and no actions should panic")
		}
	}()
	open := true
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		page(c)
		Dialog(c, &open, DialogOptions{})
	}, 900, 600)
}

// ── AlertDialog ────────────────────────────────────────────────────────────

// alertTester opens an alert from the page.
func alertTester(destructive bool) (*ui.Tester, *bool) {
	open := new(bool)
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		page(c)
		if input.Button(c, "Delete", input.ButtonOptions{Danger: true, Label: "Delete row"}).
			Clicked() {
			*open = true
		}
		if *open {
			AlertDialog(c, open, AlertDialogOptions{
				Title:       "Delete CB-2871?",
				Body:        "The callback and its history go with it.",
				Actions:     []string{"Cancel", "Delete"},
				Destructive: destructive,
			})
		}
	}, 900, 600)
	return tt, open
}

func TestAlertDialogOpenDrawsItsQuestionAndAnswers(t *testing.T) {
	tt, _ := alertTester(true)
	if err := tt.Click("Delete row"); err != nil {
		t.Fatal(err)
	}
	wantText(t, tt, "Delete CB-2871?")
	wantText(t, tt, "The callback and its history go with it.")
	find(t, tt, "Cancel")
	find(t, tt, "Delete")
}

func TestAlertDialogReportsTheAnswer(t *testing.T) {
	tt, _ := alertTester(true)
	if err := tt.Click("Delete row"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Delete"); err != nil {
		t.Fatal(err)
	}
	wantNoText(t, tt, "Delete CB-2871?")
}

func TestAlertDialogEscapeCancels(t *testing.T) {
	tt, open := alertTester(true)
	if err := tt.Click("Delete row"); err != nil {
		t.Fatal(err)
	}
	tt.Key(0, ui.KeyEscape)
	if *open {
		t.Error("Escape left the alert open rather than picking Cancel")
	}
	wantNoText(t, tt, "Delete CB-2871?")
}

// Escape on an alert is an answer rather than only a dismissal, so Chosen has
// to say which button it picked. Watching the alert close is not the same
// thing: closing is what every dialog does with Escape, alert or not, which
// is how an alert that closed and reported nothing at all could pass.
//
// The index is wherever the button labelled Cancel sits, matched by name
// rather than by place, so both rows below are worth holding to: an alert
// whose Cancel is last and one whose Cancel is first.
func TestAlertDialogEscapeChoosesCancel(t *testing.T) {
	for _, tc := range []struct {
		name    string
		actions []string
		want    int
	}{
		{"cancel last", []string{"Keep", "Cancel"}, 1},
		{"cancel first", []string{"Cancel", "Delete"}, 0},
	} {
		open, chosen := true, -1
		tt := ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			page(c)
			res := AlertDialog(c, &open, AlertDialogOptions{
				Title:   "Delete CB-2871?",
				Body:    "The callback and its history go with it.",
				Actions: tc.actions,
			})
			// Read inside the view, as every report in this library is: the
			// settled pass has no key in it and reports nothing.
			if i := res.Chosen(); i >= 0 {
				chosen = i
			}
		}, 900, 600)

		tt.Key(0, ui.KeyEscape)
		if chosen != tc.want {
			t.Errorf("%s: Escape chose %d, want %d, the index of the button labelled Cancel",
				tc.name, chosen, tc.want)
		}
		if open {
			t.Errorf("%s: Escape left the alert open rather than picking Cancel", tc.name)
		}
		wantNoText(t, tt, "Delete CB-2871?")
	}
}

// A non-modal alert is a question asked beside a page someone is still
// working in, not an interruption, so Escape is not its key — the same key a
// Dialog told not to be modal refuses. Its own buttons are how it is answered.
func TestNonModalAlertDialogIgnoresEscape(t *testing.T) {
	open, chosen := true, -1
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		page(c)
		res := AlertDialog(c, &open, AlertDialogOptions{
			Title: "Delete CB-2871?", Body: "The callback and its history go with it.",
			Actions: []string{"Cancel", "Delete"}, NonModal: true,
		})
		if i := res.Chosen(); i >= 0 {
			chosen = i
		}
	}, 900, 600)

	tt.Key(0, ui.KeyEscape)
	if !open {
		t.Error("Escape closed an alert told not to be modal")
	}
	if chosen != -1 {
		t.Errorf("Escape answered %d on an alert told not to be modal", chosen)
	}
	wantText(t, tt, "Delete CB-2871?")
}

// Asking for Escape is a registration, and the registration is what takes the
// delivery, not the read: an alert that asked while only non-modal would
// swallow the one key the modal layer underneath it is waiting for. The alert
// is built last, so it is the topmost overlay and would be the one to win.
func TestNonModalAlertLeavesEscapeForTheModalLayerUnderIt(t *testing.T) {
	dialog, alert := true, true
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		page(c)
		Dialog(c, &dialog, DialogOptions{
			Title: "Edit callback",
			Body:  func() { ui.Text(c, "Customer").Label("field-customer") },
		})
		AlertDialog(c, &alert, AlertDialogOptions{
			Title: "Delete CB-2871?", Body: "The callback and its history go with it.",
			Actions: []string{"Cancel", "Delete"}, NonModal: true,
		})
	}, 900, 600)

	tt.Key(0, ui.KeyEscape)
	if dialog {
		t.Error("the modal dialog did not get an Escape a non-modal alert had taken")
	}
	if !alert {
		t.Error("the non-modal alert closed on a key it should never have asked for")
	}
	wantNoText(t, tt, "Edit callback")
	wantText(t, tt, "Delete CB-2871?")
}

// An alert nobody can answer is not an alert.
func TestAlertDialogNeedsAnAction(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("an alert with no actions should panic")
		}
	}()
	open := true
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		page(c)
		AlertDialog(c, &open, AlertDialogOptions{Title: "Delete?"})
	}, 900, 600)
}

// ── Drawer ─────────────────────────────────────────────────────────────────

// A drawer is a dialog hung from an edge: full height, against the window, and
// closing the same two ways.
func TestDrawerHangsFromTheRightEdge(t *testing.T) {
	open := new(bool)
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		page(c)
		if input.Button(c, "Log", input.ButtonOptions{Label: "Log"}).
			Clicked() {
			*open = true
		}
		if *open {
			Drawer(c, open, DrawerOptions{
				Side:  ui.End,
				Title: "Log callback",
				Body:  func() { ui.Text(c, "Customer").Label("field-customer") },
			})
		}
	}, 1000, 700)

	wantNoText(t, tt, "Log callback")
	if err := tt.Click("Log"); err != nil {
		t.Fatal(err)
	}
	sheet := find(t, tt, "Log callback")

	if sheet.H < 700-sheet.H {
		t.Errorf("the sheet is %v tall in a 700pt window; it does not reach the top and bottom",
			sheet.H)
	}
	if right := sheet.X + sheet.W; right < 1000-2 {
		t.Errorf("the sheet ends at %v, short of the right edge at 1000", right)
	}
	find(t, tt, "field-customer")

	tt.Key(0, ui.KeyEscape)
	if *open {
		t.Error("Escape left the sheet open")
	}
	wantNoText(t, tt, "Log callback")
}

// A drawer from the left and one from the right are the same sheet at two
// edges; a component that could not tell them apart would have to be told
// which side to hang on twice.
func TestDrawerSideIsItsOwnAxis(t *testing.T) {
	for _, tc := range []struct {
		name string
		side ui.Align
		left bool
	}{
		{"start", ui.Start, true},
		{"end", ui.End, false},
	} {
		open := true
		tt := ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			page(c)
			Drawer(c, &open, DrawerOptions{
				Side:  tc.side,
				Title: "Log callback",
			})
		}, 1000, 700)

		sheet := find(t, tt, "Log callback")
		if tc.left && sheet.X > 2 {
			t.Errorf("a drawer from the start edge starts at %v", sheet.X)
		}
		if !tc.left && sheet.X+sheet.W < 1000-2 {
			t.Errorf("a drawer from the end edge ends at %v", sheet.X+sheet.W)
		}
	}
}

// ── Popover ────────────────────────────────────────────────────────────────

// popoverTester puts a row in the page with a popover under it.
func popoverTester(nonModal bool) (*ui.Tester, *bool) {
	open := new(bool)
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		k := core.Tokens(c)
		ui.Column(c).Fill().Background(k.Surface).Padding(24).Gap(16).Children(func() {
			ui.Text(c, "Open callbacks").FontSize(theme.DisplaySize).Bold().TextColor(k.Text)
			row := ui.Row(c).Grow(0).Radius(theme.ControlRadius).Background(k.Background).
				Padding(u4*2, u4*3).Label("CB-2871").Children(func() {
				ui.Text(c, "CB-2871").FontSize(theme.BodySize).TextColor(k.Text)
			})
			if row.Clicked() {
				*open = true
			}
			Popover(c, row, open, PopoverOptions{
				Title:    "Riverside Clinic",
				Subtitle: "AC not cooling",
				Modal:    !nonModal,
				Body:     func() { ui.Text(c, "Opened by Dana Reyes") },
			})
		})
	}, 900, 600)
	return tt, open
}

// The one spacing step of the window's density, for building a fixture without
// a view of its own.
var u4 float32 = 4

func TestPopoverClosedDrawsNothing(t *testing.T) {
	tt, _ := popoverTester(true)
	wantNoText(t, tt, "Riverside Clinic")
	wantNoText(t, tt, "Opened by Dana Reyes")
}

func TestPopoverOpensFromItsAnchor(t *testing.T) {
	tt, open := popoverTester(true)
	if err := tt.Click("CB-2871"); err != nil {
		t.Fatal(err)
	}
	if !*open {
		t.Fatal("the row's press did not open the popover")
	}
	wantText(t, tt, "AC not cooling")
	wantText(t, tt, "Opened by Dana Reyes")

	row := find(t, tt, "CB-2871")
	panel := find(t, tt, "Riverside Clinic")
	if panel.Y < row.Y+row.H {
		t.Errorf("the panel at %v is not under its anchor at %v", panel, row)
	}
}

func TestPopoverEscapeCloses(t *testing.T) {
	tt, open := popoverTester(true)
	if err := tt.Click("CB-2871"); err != nil {
		t.Fatal(err)
	}
	wantText(t, tt, "Opened by Dana Reyes")

	tt.Key(0, ui.KeyEscape)
	if *open {
		t.Error("Escape left the popover open")
	}
	wantNoText(t, tt, "Opened by Dana Reyes")
}

// A modal popover puts a scrim over the window; a non-modal one lets the page
// behind keep working. Both are spelled out in the component, so both are
// worth holding to.
func TestPopoverModality(t *testing.T) {
	for _, tc := range []struct {
		name     string
		nonModal bool
	}{
		{"modal", false},
		{"non-modal", true},
	} {
		tt, open := popoverTester(tc.nonModal)
		if err := tt.Click("CB-2871"); err != nil {
			t.Fatal(err)
		}
		wantText(t, tt, "Opened by Dana Reyes")

		tt.Key(0, ui.KeyEscape)
		if *open {
			t.Errorf("%s: Escape left the popover open", tc.name)
		}
	}
}

// ── Popconfirm ─────────────────────────────────────────────────────────────

func popconfirmTester(destructive bool) (*ui.Tester, *bool, *Choice) {
	open := new(bool)
	choice := new(Choice)
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		k := core.Tokens(c)
		ui.Column(c).Fill().Background(k.Surface).Padding(24).Gap(16).Children(func() {
			ui.Text(c, "Open callbacks").FontSize(theme.DisplaySize).Bold().TextColor(k.Text)
			row := ui.Row(c).Grow(0).Radius(theme.ControlRadius).Background(k.Background).
				Padding(8, 12).Label("CB-2871").Children(func() {
				ui.Text(c, "CB-2871").FontSize(theme.BodySize).TextColor(k.Text)
			})
			if row.Clicked() {
				*open = true
			}
			res := Popconfirm(c, row, open, PopconfirmOptions{
				Title:       "Delete CB-2871?",
				Body:        "The callback and its history go with it.",
				Confirm:     "Delete",
				Cancel:      "Keep",
				Destructive: destructive,
				Modal:       true,
			})
			if res.Chosen() != NoChoice {
				*choice = res.Chosen()
				*open = false
			}
		})
	}, 900, 600)
	return tt, open, choice
}

func TestPopconfirmAnswersBothWays(t *testing.T) {
	for _, tc := range []struct {
		button string
		want   Choice
	}{
		{"Delete", Confirmed},
		{"Keep", Cancelled},
	} {
		tt, open, choice := popconfirmTester(true)
		if err := tt.Click("CB-2871"); err != nil {
			t.Fatal(err)
		}
		wantText(t, tt, "Delete CB-2871?")
		if err := tt.Click(tc.button); err != nil {
			t.Fatal(err)
		}
		if *choice != tc.want {
			t.Errorf("pressing %q answered %v, want %v", tc.button, *choice, tc.want)
		}
		if *open {
			t.Errorf("pressing %q left the confirm open", tc.button)
		}
		wantNoText(t, tt, "Delete CB-2871?")
	}
}

func TestPopconfirmEscapeCancels(t *testing.T) {
	tt, open, choice := popconfirmTester(false)
	if err := tt.Click("CB-2871"); err != nil {
		t.Fatal(err)
	}
	tt.Key(0, ui.KeyEscape)
	if *open {
		t.Fatal("Escape left the confirm open")
	}
	if *choice != Cancelled {
		t.Errorf("Escape answered %v, want Cancelled", *choice)
	}
}

// A confirm with no question in it is a button somebody is being asked to
// trust.
func TestPopconfirmNeedsAQuestion(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("a confirm with no title should panic")
		}
	}()
	open, row := true, (*ui.Element)(nil)
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		page(c)
		Popconfirm(c, row, &open, PopconfirmOptions{Confirm: "Yes", Cancel: "No"})
	}, 900, 600)
}

// ── Tooltip ────────────────────────────────────────────────────────────────

func TestTooltipShowsForTheFocusAndGoesWithEscape(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		k := core.Tokens(c)
		ui.Column(c).Fill().Background(k.Surface).Padding(24).Gap(12).Children(func() {
			save := input.Button(c, "Save", input.ButtonOptions{Label: "Save"})
			Tooltip(c, save, "Saves the callback and clears the form")
		})
	}, 900, 600)

	wantNoText(t, tt, "Saves the callback and clears the form")

	// The keyboard focus brings a tip at once — it has already arrived where
	// it is going, so there is nothing to wait for.
	tt.Key(0, ui.KeyTab)
	wantText(t, tt, "Saves the callback and clears the form")

	tt.Key(0, ui.KeyEscape)
	wantNoText(t, tt, "Saves the callback and clears the form")
}

// An empty tip covers the control it hides and says nothing.
func TestTooltipNeedsText(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("a tip with no text should panic")
		}
	}()
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		page(c)
		Tooltip(c, ui.Box(c), "")
	}, 900, 600)
}

// ── ContextMenu ────────────────────────────────────────────────────────────

func TestContextMenuOnSecondaryClick(t *testing.T) {
	var chosen string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		k := core.Tokens(c)
		row := ui.Row(c).Grow(0).Radius(theme.ControlRadius).Background(k.Background).
			Padding(8, 12).Label("CB-2871").Children(func() {
			ui.Text(c, "CB-2871").FontSize(theme.BodySize).TextColor(k.Text)
		})
		ContextMenu(c, row, ContextMenuOptions{Build: func(m *ui.Menu) {
			if m.Item("Rename").Chosen() {
				chosen = "Rename"
			}
			m.Separator()
			if m.Item("Delete").Chosen() {
				chosen = "Delete"
			}
		}})
	}, 900, 600)

	if err := tt.RightClick("CB-2871"); err != nil {
		t.Fatal(err)
	}
	menu := tt.Menu()
	if len(menu) != 3 {
		t.Fatalf("the menu shows %q, want Rename, a separator and Delete", menu)
	}
	if menu[1] != "-" {
		t.Errorf("the separator is %q", menu[1])
	}

	if err := tt.ChooseMenuItem("Delete"); err != nil {
		t.Fatal(err)
	}
	if chosen != "Delete" {
		t.Errorf("choosing Delete reported %q", chosen)
	}
}

func TestContextMenuNeedsAnElementAndABuild(t *testing.T) {
	for name, fn := range map[string]func(){
		"no anchor": func() {
			ui.NewTester(func(c *ui.Context) {
				core.Use(c, core.Settings{})
				ContextMenu(c, nil, ContextMenuOptions{Build: func(*ui.Menu) {}})
			}, 900, 600)
		},
		"no build": func() {
			ui.NewTester(func(c *ui.Context) {
				core.Use(c, core.Settings{})
				ContextMenu(c, ui.Box(c), ContextMenuOptions{})
			}, 900, 600)
		},
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s: a context menu with %s should panic", name, name)
				}
			}()
			fn()
		}()
	}
}

// ── DropdownMenu ───────────────────────────────────────────────────────────

func TestDropdownMenuOpensOnItsTrigger(t *testing.T) {
	var chosen string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		page(c)
		DropdownMenu(c, "Group by branch", DropdownMenuOptions{Label: "Group by branch",
			Build: func(m *ui.Menu) {
				if m.Item("Branch").Chosen() {
					chosen = "Branch"
				}
				if m.Item("Technician").Chosen() {
					chosen = "Technician"
				}
			},
		})
	}, 900, 600)

	if err := tt.Click("Group by branch"); err != nil {
		t.Fatal(err)
	}
	if menu := tt.Menu(); len(menu) != 2 {
		t.Fatalf("the menu shows %q, want Branch and Technician", menu)
	}
	if err := tt.ChooseMenuItem("Technician"); err != nil {
		t.Fatal(err)
	}
	if chosen != "Technician" {
		t.Errorf("choosing Technician reported %q", chosen)
	}
}

// ── HoverCard ──────────────────────────────────────────────────────────────

// hoverTester puts a card in the page that opens a hover card on the pointer.
func hoverTester() (*ui.Tester, *bool) {
	open := new(bool)
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		k := core.Tokens(c)
		ui.Column(c).Fill().Background(k.Surface).Padding(24).Gap(16).Children(func() {
			ui.Text(c, "Open callbacks").FontSize(theme.DisplaySize).Bold().TextColor(k.Text)
			card := ui.Row(c).Grow(0).Radius(theme.CardRadius).Background(k.Background).
				Padding(12, 16).Label("Maple Street Bakery").Children(func() {
				ui.Text(c, "Maple Street Bakery").FontSize(theme.BodySize).TextColor(k.Text)
			})
			HoverCard(c, card, open, HoverCardOptions{
				Modal:    false,
				Title:    "Last resolved",
				Subtitle: "Riverside Clinic, 4 March",
			})
		})
	}, 900, 600)
	return tt, open
}

func TestHoverCardFollowsThePointer(t *testing.T) {
	tt, open := hoverTester()
	wantNoText(t, tt, "Riverside Clinic, 4 March")

	card := find(t, tt, "Maple Street Bakery")
	tt.Move(card.X+card.W/2, card.Y+card.H/2)
	// The card is built on the frame after the pointer arrives, as any layer
	// is: a window keeps drawing, and the test has to ask for the frame.
	tt.Frame()
	if !*open {
		t.Fatal("the pointer resting on the card did not open the hover card")
	}
	wantText(t, tt, "Riverside Clinic, 4 March")

	// Off both the card and its hover card, and it goes.
	tt.Move(880, 580)
	tt.Frame()
	if *open {
		t.Error("the hover card stayed after the pointer left it")
	}
	wantNoText(t, tt, "Riverside Clinic, 4 March")
}

// ── SplitButton ────────────────────────────────────────────────────────────

func TestSplitButtonPressesOnlyItsActionHalf(t *testing.T) {
	var chosen string
	var presses int
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		page(c)
		if SplitButton(c, "Log callback", SplitButtonOptions{
			Label:   "More ways to log a callback",
			Primary: true,
			Build: func(m *ui.Menu) {
				if m.Item("Log from a template").Chosen() {
					chosen = "Log from a template"
				}
			},
		}).Pressed() {
			presses++
		}
	}, 900, 600)

	// The action half answers a press and nothing else does.
	if err := tt.Click("Log callback"); err != nil {
		t.Fatal(err)
	}
	if presses != 1 {
		t.Errorf("the action half reported %d presses, want 1", presses)
	}
	if tt.Menu() != nil {
		t.Error("pressing the action half opened the menu")
	}

	// The arrow opens the menu and is not the action.
	if err := tt.Click("More ways to log a callback"); err != nil {
		t.Fatal(err)
	}
	if presses != 1 {
		t.Errorf("opening the menu counted as %d presses on the action", presses)
	}
	if menu := tt.Menu(); len(menu) != 1 {
		t.Fatalf("the menu shows %q, want the one item", menu)
	}
	if err := tt.ChooseMenuItem("Log from a template"); err != nil {
		t.Fatal(err)
	}
	if chosen != "Log from a template" {
		t.Errorf("choosing the item reported %q", chosen)
	}
}

// The two halves are one control: same height, and the action wider than the
// arrow because it is the one carrying the words.
func TestSplitButtonHalvesAreOneControl(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		page(c)
		SplitButton(c, "Log callback", SplitButtonOptions{
			Label:   "More ways to log a callback",
			Primary: true,
			Build:   func(m *ui.Menu) { m.Item("Log from a template") },
		})
	}, 900, 600)

	action := find(t, tt, "Log callback")
	arrow := find(t, tt, "More ways to log a callback")

	if action.H != arrow.H {
		t.Errorf("the action half is %v tall and the arrow %v; they are one control",
			action.H, arrow.H)
	}
	if arrow.W >= action.W {
		t.Errorf("the arrow is %v wide and the action %v; the words need the room",
			arrow.W, action.W)
	}
	if arrow.X < action.X+action.W {
		t.Errorf("the arrow at %v overlaps the action ending at %v", arrow.X, action.X+action.W)
	}
}

func TestSplitButtonNeedsALabelledArrow(t *testing.T) {
	for name, opts := range map[string]SplitButtonOptions{
		"no build": {Label: "More"},
		"no label": {Build: func(*ui.Menu) {}},
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("a split button with %s should panic", name)
				}
			}()
			ui.NewTester(func(c *ui.Context) {
				core.Use(c, core.Settings{})
				page(c)
				SplitButton(c, "Log callback", opts)
			}, 900, 600)
		}()
	}
}

// ── both appearances ───────────────────────────────────────────────────────

// Every layer is drawn from the window's palette, so the same dialog in both
// appearances is two different pictures rather than one picture with the light
// one painted over. The page underneath is Surface in each, which is what
// makes the panel's own colour readable at all in the light palette.
func TestOverlaysDrawInBothAppearances(t *testing.T) {
	light := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: core.Light})
		darkPage(c)
		open := true
		Dialog(c, &open, DialogOptions{
			Title: "Log callback",
			Body:  func() { ui.Text(c, "Customer").Label("field-customer") },
		})
	}, 900, 600)

	dark := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: core.Dark})
		darkPage(c)
		open := true
		Dialog(c, &open, DialogOptions{
			Title: "Log callback",
			Body:  func() { ui.Text(c, "Customer").Label("field-customer") },
		})
	}, 900, 600)

	panel := find(t, light, "Log callback")
	// Four DIPs inside the panel's top edge is inside its padding and clear
	// of both the hairline and the title.
	sx, sy := int(panel.X+panel.W/2), int(panel.Y)+4

	lr, lg, lb := pixelAt(light, sx, sy)
	dr, dg, db := pixelAt(dark, sx, sy)

	// The panel is the window's own Background in each appearance: white on
	// the light page, near-black on the dark one.
	if lr < 0xF0 || lg < 0xF0 || lb < 0xF0 {
		t.Errorf("the light panel is not white: #%02X%02X%02X", lr, lg, lb)
	}
	if dr > 0x30 || dg > 0x30 || db > 0x30 {
		t.Errorf("the dark panel is not near-black: #%02X%02X%02X", dr, dg, db)
	}
	// And the page behind it is still the page — dimmed by the scrim, not
	// repainted. What matters is that it reads darker than the panel sitting
	// on it, in both appearances, and that the light page stays a long way
	// above the dark one. That last pair is the point: a scrim drawn from a
	// fixed colour rather than from the tokens would close the gap on the
	// dark palette, where the page is already close to black.
	px, py := int(6), int(6)
	pr, pg, pb := pixelAt(dark, px, py)
	if pr > 0x60 || pg > 0x60 || pb > 0x60 {
		t.Errorf("the dark page behind the dialog is not dark: #%02X%02X%02X", pr, pg, pb)
	}
	lr, lg, lb = pixelAt(light, px, py)
	if int(lr)-int(pr) < 80 {
		t.Errorf("light page #%02X%02X%02X and dark page #%02X%02X%02X are too close: "+
			"the page behind the dialog must still follow the window's appearance",
			lr, lg, lb, pr, pg, pb)
	}
	// A modal dialog dims what it covers. The page has to be visibly darker
	// than the panel, or the dialog is floating on nothing.
	panelL, _, _ := pixelAt(light, sx, sy)
	if lr > panelL {
		t.Errorf("page #%02X is lighter than the panel #%02X; the scrim is not being drawn",
			lr, panelL)
	}
}

// darkPage is page for the appearance tests, where the window's own mode is
// being set rather than followed.
func darkPage(c *ui.Context) {
	k := core.Tokens(c)
	ui.Box(c).Fill().Background(k.Surface).Children(func() {})
}
