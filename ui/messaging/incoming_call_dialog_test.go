package messaging

import (
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// The incoming call is the one dialog a messaging window owns that is not
// about a message: somebody is ringing, and the window has to say who, that
// it is ringing, and offer the two answers. The tests assert the words on the
// panel and the press each button gives back.

func TestIncomingCallDialogDrawsTheCaller(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		IncomingCallDialog(c, "Dana Reyes", IncomingCallDialogOptions{
			Subtitle: "+1 555 0100",
		})
	}, 320, 360)
	wantText(t, tt, "the incoming call", "Dana Reyes", "+1 555 0100",
		"Ringing…", "Accept", "Decline")
}

// A call with no picture of its own draws the caller's initials rather than
// a square: an empty face reads as a failed call, and initials read as a face
// not yet sent.
func TestIncomingCallDialogWithoutAPictureDrawsInitials(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		IncomingCallDialog(c, "Dana Reyes", IncomingCallDialogOptions{})
	}, 320, 360)
	if _, ok := tt.Find("DR"); !ok {
		t.Errorf("a call with no picture drew no face; %q", tt.Texts())
	}
}

// Each button reports its own answer and only that one: accepting is starting
// a call and declining is ending it, and the component that mixed them up
// would have made the wrong call for the reader.
func TestIncomingCallDialogReportsTheAnswer(t *testing.T) {
	for _, tc := range []struct {
		button  string
		accept  int
		decline int
	}{
		{"Accept", 1, 0},
		{"Decline", 0, 1},
	} {
		accepted, declined := 0, 0
		tt := ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			res := IncomingCallDialog(c, "Dana Reyes", IncomingCallDialogOptions{})
			if res.Accepted() {
				accepted++
			}
			if res.Declined() {
				declined++
			}
		}, 320, 360)
		if err := tt.Click(tc.button); err != nil {
			t.Fatal(err)
		}
		if accepted != tc.accept || declined != tc.decline {
			t.Errorf("pressing %s reported accepted=%d declined=%d, want %d and %d",
				tc.button, accepted, declined, tc.accept, tc.decline)
		}
	}
}

// A ringing panel with no name is a call nobody can be told about, and it is
// the case worth stopping for.
func TestIncomingCallDialogNeedsAName(t *testing.T) {
	for _, what := range []string{"", "   "} {
		wantsPanic(t, "an incoming call about nobody ("+what+")", func() {
			ui.NewTester(func(c *ui.Context) {
				core.Use(c, core.Settings{})
				IncomingCallDialog(c, what, IncomingCallDialogOptions{})
			}, 300, 300)
		})
	}
}

// The dialog is drawn from the window's palette like the rest of a call: in
// the dark appearance the tokens it resolves are the dark ones, and the
// caller's name is still on the panel.
func TestIncomingCallDialogInDark(t *testing.T) {
	var bg ui.Color
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: core.Dark})
		bg = core.Tokens(c).Background
		IncomingCallDialog(c, "dm.caller.name", IncomingCallDialogOptions{
			Subtitle: "dm.caller.sub",
		})
	}, 320, 360)
	if bg != theme.Dark().Background {
		t.Errorf("a dark window resolved the background %v, want %v", bg, theme.Dark().Background)
	}
	wantText(t, tt, "the dark incoming call", "dm.caller.name", "dm.caller.sub",
		"Ringing…", "Accept", "Decline")
}
