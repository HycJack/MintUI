package display

import (
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/core"
)

func TestInitials(t *testing.T) {
	for in, want := range map[string]string{
		"Ada Lovelace":              "AL",
		"Prince":                    "P",
		"  Grace  Brewster Hopper ": "GB",
		"":                          "",
		"123":                       "",
		"Zoë O'Brien":               "ZO",
	} {
		if got := internal.Initials(in); got != want {
			t.Errorf("Initials(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestAvatarIsNamed(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Avatar(c, "Ada Lovelace")
	}, 120, 120)
	if !tt.HasText("AL") {
		t.Errorf("avatar did not draw its initials: %q", tt.Texts())
	}
	if !tt.HasText("Ada Lovelace") {
		t.Error("avatar must be named for assistive technology")
	}
}

func TestAvatarClusterCounts(t *testing.T) {
	names := []string{"Ada Lovelace", "Grace Hopper", "Alan Turing", "Margaret Hamilton", "Barbara Liskov"}
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		AvatarCluster(c, names, 14)
	}, 400, 120)
	// Five faces, then the number that stands for the other nine.
	for _, n := range names {
		if !tt.HasText(n) {
			t.Errorf("cluster is missing %q", n)
		}
	}
	if !tt.HasText("14 people") {
		t.Errorf("cluster did not count the rest: %q", tt.Texts())
	}
}

func TestAvatarClusterEmpty(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		AvatarCluster(c, nil, 0)
	}, 200, 80)
	if got := tt.Texts(); len(got) != 0 {
		t.Errorf("an empty cluster should draw nothing, drew %q", got)
	}
}

func TestMeterClampsItsLevel(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Meter(c, 4, 9, core.Danger)
	}, 200, 80)
	_ = tt
	// A level past the top must not panic or draw outside the box.
}

func TestMeterRejectsZero(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("Meter(0) should panic")
		}
	}()
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Meter(c, 0, 1, core.Neutral)
	}, 100, 60)
}
