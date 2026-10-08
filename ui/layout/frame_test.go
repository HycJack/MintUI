package layout

import (
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

func TestContainerSlotsItsChildren(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Container(c, ContainerOptions{Pad: 12, Surface: true, Border: true}, func() {
			ui.Text(c, "Riverside Clinic")
		})
	}, 400, 200)
	if !tt.HasText("Riverside Clinic") {
		t.Errorf("container dropped its children: %q", tt.Texts())
	}
}

func TestDividerIsAHairline(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Divider(c, DividerOptions{})
	}, 300, 100)
	// A divider is a drawn line, not text; assert it laid out at all.
	if len(tt.Texts()) != 0 {
		t.Errorf("a divider should draw no text, drew %q", tt.Texts())
	}
}

func TestScrollAreaNeedsAHeight(t *testing.T) {
	// Without one the viewport grows to fit and never scrolls, which is a
	// silently inert control. It has to stop instead.
	defer func() {
		if recover() == nil {
			t.Error("a scrolling ScrollArea with no Height should panic")
		}
	}()
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ScrollArea(c, ScrollAreaOptions{Vertical: true}, func() {
			ui.Text(c, "tall")
		})
	}, 300, 100)
}

func TestScrollAreaWithoutScrollingIsFine(t *testing.T) {
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ScrollArea(c, ScrollAreaOptions{}, func() {
			ui.Text(c, "just a box")
		})
	}, 300, 100)
}

func TestScrollAreaKeepsItsPlace(t *testing.T) {
	var st ui.ScrollState
	// The result is what a caller reads, so capture it out of the frame.
	var res ScrollResult
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		res = ScrollArea(c, ScrollAreaOptions{Vertical: true, Height: 100, State: &st}, func() {
			for i := 0; i < 40; i++ {
				ui.Text(c, "line")
			}
		})
	}, 300, 100)

	if st.MaxY <= 0 {
		t.Fatal("content taller than the viewport should report a scroll range")
	}
	if _, _, _, maxY := res.Offset(); maxY != st.MaxY {
		t.Errorf("Offset reports maxY %v, state says %v", maxY, st.MaxY)
	}
	if res.AtEnd() {
		t.Error("a fresh area starts at the top, not the end")
	}
	st.Y = st.MaxY
	if !res.AtEnd() {
		t.Error("scrolled to the bottom, AtEnd should say so")
	}
}

func TestStackKeepsTheBaseSize(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Stack(c, StackOptions{}, func() {
			ui.Box(c).Size(40, 40).Label("base").Children(func() {})
			ui.Box(c).Size(12, 12).Label("badge").Children(func() {})
		})
	}, 200, 200)
	base, ok := tt.Find("base")
	if !ok {
		t.Fatal("the stack lost its base")
	}
	if base.W > 48 || base.H > 48 {
		t.Errorf("stack grew to %vx%v to fit the badge; it should keep the base's 40x40",
			base.W, base.H)
	}
}

func TestAspectRatioHoldsItsShape(t *testing.T) {
	for _, ratio := range []float32{1, 16.0 / 9.0, 0.5} {
		tt := ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			AspectRatio(c, AspectRatioOptions{Ratio: ratio}, func() {
				ui.Box(c).Fill().Label("inner").Children(func() {})
			}).Width(320)
		}, 600, 600)
		r, ok := tt.Find("inner")
		if !ok {
			t.Fatalf("ratio %v: child missing", ratio)
		}
		got := r.W / r.H
		if got < ratio*0.9 || got > ratio*1.1 {
			t.Errorf("ratio %v: child is %v, want about %v", ratio, got, ratio)
		}
	}
}

func TestSplitPaneSplitsEvenlyByDefault(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		SplitPane(c, SplitPaneOptions{
			FirstName: "first pane", SecondName: "second pane",
		}, func() {
			ui.Text(c, "first")
		}, func() {
			ui.Text(c, "second").Label("second content")
		})
	}, 800, 400)
	// Measure the content, not the pane: an empty box measures zero, but
	// where the second pane's content *starts* is exactly the first pane's
	// share, which is the thing a split is for.
	a, ok1 := tt.Find("second content")
	if !ok1 {
		t.Fatalf("second pane's content missing: %q", tt.Texts())
	}
	// Stacked by default, so the split shows up on Y, not X.
	if a.Y < 150 || a.Y > 250 {
		t.Errorf("second pane starts at y=%v; an even split of 400 puts it near 200", a.Y)
	}
	if a.X > 40 {
		t.Errorf("second pane is at x=%v; a stacked split stays in the same column", a.X)
	}
}

func TestSplitPaneHonoursItsFraction(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		SplitPane(c, SplitPaneOptions{
			First: 0.25, FirstName: "first pane", SecondName: "second pane",
		}, func() {
			ui.Text(c, "first")
		}, func() {
			ui.Text(c, "second").Label("second content")
		})
	}, 800, 400)
	second, ok := tt.Find("second content")
	if !ok {
		t.Fatalf("second pane's content missing: %q", tt.Texts())
	}
	// 25% of 400 is 100; the gutter is a few DIPs on top of it.
	if second.Y < 90 || second.Y > 130 {
		t.Errorf("second pane starts at y=%v; First=0.25 of 400 puts it near 100", second.Y)
	}
}

func TestSplitPaneStacksWhenNotVertical(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		SplitPane(c, SplitPaneOptions{
			SideBySide: true, First: 0.75,
			FirstName: "left pane", SecondName: "right pane",
		}, func() {
			ui.Text(c, "left").Label("left content")
		}, func() {
			ui.Text(c, "right").Label("right content")
		})
	}, 600, 400)
	right, ok := tt.Find("right content")
	if !ok {
		t.Fatalf("right pane's content missing: %q", tt.Texts())
	}
	if right.X < 380 || right.X > 480 {
		t.Errorf("right pane's content starts at %v; 75%% of 600 is 450", right.X)
	}
}

func TestSplitPaneStacksByDefault(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		SplitPane(c, SplitPaneOptions{First: 0.25}, func() {
			ui.Text(c, "top").Label("top content")
		}, func() {
			ui.Text(c, "bottom").Label("bottom content")
		})
	}, 600, 400)
	bottom, ok := tt.Find("bottom content")
	if !ok {
		t.Fatalf("bottom pane's content missing: %q", tt.Texts())
	}
	// Stacked by default: the second pane moves down, not right.
	if bottom.X > 40 {
		t.Errorf("bottom pane's content is at x=%v; a stacked split is below, not beside", bottom.X)
	}
}

func TestAppShellPutsThingsInTheirPlaces(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		AppShell(c, nil, AppShellOptions{}, func() {
			ui.Text(c, "sidebar")
		}, func() {
			ui.Text(c, "main")
		}, func() {
			StatusBar(c, StatusBarOptions{Leading: func() { ui.Text(c, "ready") }})
		})
	}, 1200, 700)
	side, ok1 := tt.Find("sidebar")
	main, ok2 := tt.Find("main")
	if !ok1 || !ok2 {
		t.Fatalf("shell dropped a region: %q", tt.Texts())
	}
	if side.W < theme.SidebarWidth-2 {
		t.Errorf("sidebar is %v wide, want %v", side.W, theme.SidebarWidth)
	}
	if main.X < side.X+side.W {
		t.Errorf("main starts at %v, inside the sidebar that ends at %v",
			main.X, side.X+side.W)
	}
	if !tt.HasText("ready") {
		t.Error("the status bar was not drawn")
	}
}

func TestAppShellCollapsesToARail(t *testing.T) {
	collapsed := true
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		AppShell(c, &collapsed, AppShellOptions{}, func() {
			ui.Text(c, "sidebar")
		}, func() {
			ui.Text(c, "main")
		}, nil)
	}, 1200, 700)
	side, _ := tt.Find("sidebar")
	if side.W > theme.RailWidth+2 {
		t.Errorf("a collapsed shell has a %v sidebar, want the %v rail",
			side.W, theme.RailWidth)
	}
}

func TestTitleBarPutsTitleLeftAndTrailingRight(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		TitleBar(c, TitleBarOptions{
			Title:    func() { ui.Text(c, "Callbacks") },
			Trailing: func() { ui.Text(c, "Sync") },
		})
	}, 800, 120)
	title, ok1 := tt.Find("Callbacks")
	trail, ok2 := tt.Find("Sync")
	if !ok1 || !ok2 {
		t.Fatalf("title bar dropped a slot: %q", tt.Texts())
	}
	if trail.X < title.X+title.W {
		t.Errorf("trailing at %v sits left of the title ending at %v",
			trail.X, title.X+title.W)
	}
}

func TestTrafficLightsAreThreeDots(t *testing.T) {
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		TrafficLights(c)
	}, 200, 80)
}
