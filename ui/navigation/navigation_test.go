package navigation

import (
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
)

func TestFilterRowFillsOnlyTheSelection(t *testing.T) {
	n := 27
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ui.Column(c).Width(300).Padding(12).Children(func() {
			FilterRow(c, "All open", FilterRowOptions{Count: &n, Selected: true})
			FilterRow(c, "High impact", FilterRowOptions{})
		})
	}, 340, 200)
	for _, want := range []string{"All open", "27", "High impact"} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}
}

func TestFilterRowCountsDown(t *testing.T) {
	one := 1
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		FilterRow(c, "Awaiting parts", FilterRowOptions{Count: &one})
	}, 300, 120)
	if !tt.HasText("1") {
		t.Errorf("count missing: %q", tt.Texts())
	}
}

func TestGroupReportsTheToggle(t *testing.T) {
	open := true
	toggled := false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		if Group(c, GroupOptions{Title: "Views", Open: open}, func() {
			ui.Text(c, "Repeat failures")
		}).Toggled() {
			toggled = true
			open = !open
		}
	}, 320, 220)
	if !tt.HasText("Repeat failures") {
		t.Fatal("an open group should show its body")
	}
	if err := tt.Click("Views"); err != nil {
		t.Fatal(err)
	}
	if !toggled {
		t.Error("pressing the heading should report a toggle")
	}
	if open {
		t.Error("the caller's fold state should have flipped")
	}
	if tt.HasText("Repeat failures") {
		t.Error("a folded group should hide its body")
	}
}

func TestGroupHidesBodyWhenClosed(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Group(c, GroupOptions{Title: "Views", Open: false}, func() {
			ui.Text(c, "Repeat failures")
		})
	}, 320, 220)
	if tt.HasText("Repeat failures") {
		t.Error("a closed group should hide its body")
	}
	if !tt.HasText("Views") {
		t.Error("a closed group still shows its heading")
	}
}

func TestRailDrawsIdentityAndTools(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Rail(c, "Rosa Delgado",
			[]RailItem{{Name: "Callbacks", Selected: true}, {Name: "Team"}},
			[]RailItem{{Name: "Alerts", Badge: true}, {Name: "Settings"}})
	}, 120, 700)
	for _, want := range []string{"Rosa Delgado", "RD", "Callbacks", "Team", "Alerts", "Settings"} {
		if !tt.HasText(want) {
			t.Errorf("rail is missing %q; %q", want, tt.Texts())
		}
	}
}
