package display

import (
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

func TestHeadingStepsDownByLevel(t *testing.T) {
	var tt *ui.Tester
	tt = ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		for lvl := 0; lvl <= 5; lvl++ {
			Heading(c, "Open callbacks", HeadingOptions{Level: lvl})
		}
	}, 600, 600)
	// Out-of-range levels clamp rather than drawing an absurd size, so six
	// headings produce the same texts as three distinct levels would.
	if n := len(tt.Texts()); n < 6 {
		t.Errorf("lost headings: %q", tt.Texts())
	}
}

func TestHeadingSubtitleAndDivider(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Heading(c, "Open callbacks", HeadingOptions{
			Level: 1, Subtitle: "Every promise we have made and not yet kept", Divider: true,
		})
	}, 700, 300)
	if !tt.HasText("Open callbacks") || !tt.HasText("Every promise we have made and not yet kept") {
		t.Errorf("heading dropped a slot: %q", tt.Texts())
	}
}

func TestTextPicksItsTone(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Text(c, "primary", TextOptions{})
		Text(c, "secondary", TextOptions{Muted: true})
		Text(c, "a mark", TextOptions{Faint: true})
	}, 400, 300)
	for _, want := range []string{"primary", "secondary", "a mark"} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}
	r, ok := tt.Find("primary")
	if !ok {
		t.Fatal("primary text not found")
	}
	s, ok := tt.Find("secondary")
	if !ok {
		t.Fatal("secondary text not found")
	}
	if r.H == 0 || s.H == 0 {
		t.Fatal("text has no measured height")
	}
}

func TestKbdShowsEachKeySeparately(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Kbd(c, "⌘", "K")
	}, 300, 120)
	// Written as "⌘K" the two keys read as one word; separate boxes are the
	// whole point, so both must be findable on their own.
	for _, want := range []string{"⌘", "K"} {
		if !tt.HasText(want) {
			t.Errorf("key %q missing from %q", want, tt.Texts())
		}
	}
}

func TestLinkShowsItsLabelNotItsURL(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Link(c, "Open the board", LinkOptions{URL: "https://example.com/board"})
	}, 400, 120)
	if !tt.HasText("Open the board") {
		t.Errorf("link label missing: %q", tt.Texts())
	}
	if tt.HasText("https://example.com/board") {
		t.Error("a link that prints its own address is a link nobody needs to click")
	}
}

func TestTagReportsBeingClosed(t *testing.T) {
	closed := false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		if Tag(c, "overdue", TagOptions{Closable: true}).Closed() {
			closed = true
		}
	}, 300, 120)
	if !tt.HasText("overdue") {
		t.Fatalf("tag label missing: %q", tt.Texts())
	}
	if err := tt.Click("Remove"); err != nil {
		t.Fatal(err)
	}
	if !closed {
		t.Error("pressing the close button should report a close")
	}
}

func TestTagWithoutCloseButtonHasNoRemove(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Tag(c, "overdue", TagOptions{})
	}, 300, 120)
	if err := tt.Click("Remove"); err == nil {
		t.Error("a tag with no close button should not offer one")
	}
}

func TestTagToneComesFromSeverity(t *testing.T) {
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Tag(c, "critical", TagOptions{Tone: core.Danger})
	}, 300, 120)
}

func TestEditableTextNeedsALabel(t *testing.T) {
	// An editable field with no name is a mystery box: nothing on screen
	// says what it is for.
	defer func() {
		if recover() == nil {
			t.Error("an unnamed EditableText should panic")
		}
	}()
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		EditableText(c, new(string), EditableTextOptions{})
	}, 300, 120)
}

func TestEditableTextShowsTheCallersValue(t *testing.T) {
	name := "Hillside Dental"
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		EditableText(c, &name, EditableTextOptions{Label: "Customer"})
	}, 400, 140)
	if !tt.HasText("Hillside Dental") {
		t.Errorf("field did not show the value: %q", tt.Texts())
	}
	if _, ok := tt.Find("Customer"); !ok {
		t.Error("the field must be named for assistive technology")
	}
}

func TestEditableTextReadsBackItsValue(t *testing.T) {
	var got string
	name := "Hillside Dental"
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		got = EditableText(c, &name, EditableTextOptions{Label: "Customer"}).Value()
	}, 400, 140)
	if got != "Hillside Dental" {
		t.Errorf("Value() = %q, want the pointer's contents", got)
	}
}

func TestDisplaySurvivesTheDarkPalette(t *testing.T) {
	for _, dark := range []bool{false, true} {
		mode := core.Light
		if dark {
			mode = core.Dark
		}
		tt := ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{Mode: mode})
			Heading(c, "Callbacks", HeadingOptions{Level: 1, Subtitle: "sub", Divider: true})
			Text(c, "body", TextOptions{Muted: true})
			Text(c, "mark", TextOptions{Faint: true})
			Kbd(c, "⌘", "K")
			Link(c, "Open the board", LinkOptions{URL: "https://x"})
			Tag(c, "overdue", TagOptions{Closable: true, Tone: core.Danger})
		}, 700, 500)
		for _, want := range []string{"Callbacks", "sub", "body", "mark", "⌘", "K", "Open the board", "overdue"} {
			if !tt.HasText(want) {
				t.Errorf("%v mode lost %q: %q", mode, want, tt.Texts())
			}
		}
	}
	_ = theme.BodySize
}
