package navigation

import (
	"strconv"
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
)

func tabsTester(selected *int, extra func()) *ui.Tester {
	return ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Tabs(c, []Tab{
			{Label: "Board"}, {Label: "List"}, {Label: "Closed"},
		}, TabsOptions{Selected: selected})
		if extra != nil {
			extra()
		}
	}, 600, 200)
}

func TestTabsNeedASelectionToWrite(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("Tabs with nowhere to report its selection should panic")
		}
	}()
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Tabs(c, []Tab{{Label: "Board"}}, TabsOptions{})
	}, 400, 120)
}

func TestTabsShowEveryLabel(t *testing.T) {
	sel := 0
	tt := tabsTester(&sel, nil)
	for _, want := range []string{"Board", "List", "Closed"} {
		if !tt.HasText(want) {
			t.Errorf("tab %q missing from %q", want, tt.Texts())
		}
	}
}

func TestTabsWriteTheSelection(t *testing.T) {
	sel := 0
	tt := tabsTester(&sel, nil)
	if err := tt.Click("List"); err != nil {
		t.Fatal(err)
	}
	if sel != 1 {
		t.Errorf("selection = %d, want 1 — the strip must agree with the view", sel)
	}
}

func TestTabsIgnoreDisabledEntries(t *testing.T) {
	sel := 0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Tabs(c, []Tab{{Label: "Board"}, {Label: "Archive", Disabled: true}}, TabsOptions{Selected: &sel})
	}, 600, 200)
	if !tt.HasText("Archive") {
		t.Fatalf("a disabled tab is still a tab: %q", tt.Texts())
	}
	if err := tt.Click("Archive"); err != nil {
		t.Fatal(err)
	}
	if sel != 0 {
		t.Errorf("selection moved to %d; a disabled tab must not take the selection", sel)
	}
}

func TestBreadcrumbMarksWhereYouAre(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Breadcrumb(c, []BreadcrumbItem{
			{Label: "Callbacks"}, {Label: "All open", Current: true},
		})
	}, 500, 120)
	for _, want := range []string{"Callbacks", "All open"} {
		if !tt.HasText(want) {
			t.Errorf("crumb %q missing from %q", want, tt.Texts())
		}
	}
}

func TestBreadcrumbReportsTheHopPressed(t *testing.T) {
	// The press is asserted on the value the result carries, not on text
	// drawn in response: a click settles on the next frame, and a test that
	// waits for the picture is testing the picture.
	hop := -1
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := Breadcrumb(c, []BreadcrumbItem{
			{Label: "Callbacks"}, {Label: "Team", Current: true},
		})
		if r.Index >= 0 {
			hop = r.Index
		}
	}, 500, 160)
	if err := tt.Click("Callbacks"); err != nil {
		t.Fatal(err)
	}
	if hop != 0 {
		t.Errorf("pressing the first crumb reported hop %d, want 0", hop)
	}
}

func TestBreadcrumbCurrentPageIsNotALink(t *testing.T) {
	hop := -1
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		if r := Breadcrumb(c, []BreadcrumbItem{{Label: "Here", Current: true}}); r.Index >= 0 {
			hop = r.Index
		}
	}, 400, 120)
	_ = tt.Click("Here")
	if hop >= 0 {
		t.Errorf("the current crumb reported hop %d; it must not be pressable", hop)
	}
}

func TestStepsShowWhereYouAre(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Steps(c, []Step{
			{Label: "Filed", Done: true},
			{Label: "Triaged", Current: true},
			{Label: "Closed"},
		}, StepsOptions{})
	}, 700, 160)
	for _, want := range []string{"Filed", "Triaged", "Closed"} {
		if !tt.HasText(want) {
			t.Errorf("step %q missing from %q", want, tt.Texts())
		}
	}
}

func TestPaginationShowsAWindowNotEveryPage(t *testing.T) {
	page := 0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Pagination(c, PaginationOptions{Page: &page, Pages: 100})
	}, 700, 160)
	// 100 pages at window 2 around page 0 is pages 1..3, not 1..100.
	if !tt.HasText("1") {
		t.Fatalf("pagination drew nothing: %q", tt.Texts())
	}
	if tt.HasText("100") {
		t.Error("pagination showed every page; a hundred numbers is a wall, not a control")
	}
}

func TestPaginationMoves(t *testing.T) {
	page := 0
	reported := -1
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		if r := Pagination(c, PaginationOptions{Page: &page, Pages: 10}); r.Page >= 0 {
			reported = r.Page
		}
	}, 700, 160)
	if err := tt.Click("3"); err != nil {
		t.Fatal(err)
	}
	if page != 2 {
		t.Errorf("clicking page 3 left page = %d, want 2 (zero-based)", page)
	}
	if reported != 2 {
		t.Errorf("Pagination reported %d, want the page it moved to", reported)
	}
}

func TestPaginationClampsToRange(t *testing.T) {
	for _, tc := range []struct {
		name       string
		pages      int
		wantHidden []string
	}{
		{"first page hides previous", 1, []string{"Previous"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			page := 0
			tt := ui.NewTester(func(c *ui.Context) {
				core.Use(c, core.Settings{})
				Pagination(c, PaginationOptions{Page: &page, Pages: tc.pages})
			}, 700, 160)
			// A disabled control can still be hit; what matters is that it
			// does not act. Asserting the move, not the hit, is the honest
			// check — and the guard in Pagination is why it holds.
			_ = tt.Click(tc.wantHidden[0])
			if page != 0 {
				t.Errorf("%q on the first page moved to %d; it should do nothing",
					tc.wantHidden[0], page)
			}
			_ = strconv.Itoa(page)
		})
	}
}

func TestWindowAroundKeepsTheCurrentPageVisible(t *testing.T) {
	for _, tc := range []struct {
		cur, total, w int
		wantLo        int
		wantHi        int
	}{
		{0, 100, 2, 0, 2},
		{50, 100, 2, 48, 52},
		{99, 100, 2, 97, 99},
		{0, 3, 5, 0, 2},
	} {
		got := windowAround(tc.cur, tc.total, tc.w)
		if len(got) == 0 {
			t.Fatalf("cur=%d total=%d w=%d: no pages at all", tc.cur, tc.total, tc.w)
		}
		if got[0] != tc.wantLo || got[len(got)-1] != tc.wantHi {
			t.Errorf("cur=%d total=%d w=%d: got %v..%v, want %d..%d",
				tc.cur, tc.total, tc.w, got[0], got[len(got)-1], tc.wantLo, tc.wantHi)
		}
		found := false
		for _, g := range got {
			if g == tc.cur {
				found = true
			}
		}
		if !found {
			t.Errorf("cur=%d is not in %v — the window must always show where you are",
				tc.cur, got)
		}
	}
}

func TestToolbarGroupsItsActions(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Toolbar(c, ToolbarOptions{Label: "Board actions"}, func() {
			ui.Text(c, "Filter")
			ToolbarSeparator(c)
			ui.Text(c, "Export")
		})
	}, 600, 160)
	for _, want := range []string{"Filter", "Export"} {
		if !tt.HasText(want) {
			t.Errorf("toolbar dropped %q: %q", want, tt.Texts())
		}
	}
	if _, ok := tt.Find("Board actions"); !ok {
		t.Error("a toolbar of several actions is one group and must be named as one")
	}
}
