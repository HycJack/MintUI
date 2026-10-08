package navigation

import (
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/theme"
)

// The tests here assert what a person would get: the words on the screen, the
// boxes they occupy, and the value a component wrote into the caller's own
// variables. A press is asserted on the pointer it moved or the value it
// reported, never on the picture that settled afterwards — a click lands on
// the next frame, and a test that waits for the picture is testing the
// picture.

func findShown(t *testing.T, tt *ui.Tester, name string) ui.Rect {
	t.Helper()
	r, ok := tt.Find(name)
	if !ok {
		t.Fatalf("nothing shows %q; shown: %v", name, tt.Texts())
	}
	return r
}

func wantShown(t *testing.T, tt *ui.Tester, name string) {
	t.Helper()
	if !tt.HasText(name) {
		t.Fatalf("the window does not show %q; shown: %v", name, tt.Texts())
	}
}

func wantNotShown(t *testing.T, tt *ui.Tester, name string) {
	t.Helper()
	if tt.HasText(name) {
		t.Errorf("the window still shows %q; shown: %v", name, tt.Texts())
	}
}

// ── Match ───────────────────────────────────────────────────────────────────

// commands is the set every Match test works on. It is built so that one
// query can be answered four different ways at once, which is the only way to
// hold an order rather than a count:
//
//	"Callbacks"          prefix hit on the title
//	"Open callbacks"     a word hit inside the title
//	"Reassign callbacks" another word hit, later in the list
//	"Team"               a substring hit, and only through a keyword
var commands = []Command{
	{ID: "team", Title: "Team", Keywords: []string{"escalation", "roster"}},
	{ID: "callbacks", Title: "Callbacks"},
	{ID: "open-callback", Title: "Open callbacks"},
	{ID: "customers", Title: "Customers"},
	{ID: "reassign", Title: "Reassign callbacks", Disabled: true},
}

func titlesOf(cmds []Command) []string {
	out := make([]string, 0, len(cmds))
	for _, c := range cmds {
		out = append(out, c.Title)
	}
	return out
}

// The order Match returns is the whole contract, so it is asserted as an
// exact sequence rather than as a length. "Two results came back" would pass
// just as happily with Team first, which is the case that makes a palette
// useless.
func TestMatchOrdersPrefixOverWordOverSubstring(t *testing.T) {
	got := titlesOf(Match("cal", commands))
	want := []string{"Callbacks", "Open callbacks", "Reassign callbacks", "Team"}
	if !slices.Equal(got, want) {
		t.Errorf(`Match("cal") = %v
want %v
"Callbacks" starts with the query and "Team" only contains it, through a
keyword; the two word hits sit between them in the order they were given.`,
			got, want)
	}
}

// The same assertion one row at a time, so a failure says which pair swapped
// rather than only printing two lists that differ.
func TestMatchPutsCallbacksBeforeTeam(t *testing.T) {
	got := titlesOf(Match("cal", commands))
	iCallbacks, iTeam := slices.Index(got, "Callbacks"), slices.Index(got, "Team")
	if iCallbacks < 0 || iTeam < 0 {
		t.Fatalf(`Match("cal") = %v, which does not hold both rows`, got)
	}
	if iCallbacks > iTeam {
		t.Errorf(`Match("cal") = %v: "Callbacks" is at %d and "Team" at %d; a prefix
hit outranks a substring one even when the substring was found first`,
			got, iCallbacks, iTeam)
	}
}

func TestMatchDropsWhatDoesNotMatch(t *testing.T) {
	got := titlesOf(Match("call", commands))
	want := []string{"Callbacks", "Open callbacks", "Reassign callbacks"}
	if !slices.Equal(got, want) {
		t.Errorf(`Match("call") = %v, want %v`, got, want)
	}
}

// Two commands of one grade keep the caller's order: the list is how a caller
// says which of its equally good matches it would rather have.
func TestMatchKeepsTheCallersOrderWithinAGrade(t *testing.T) {
	reversed := []Command{
		commands[4], // Reassign callbacks, the word hit, given first here
		commands[2], // Open callbacks, the same grade
	}
	got := titlesOf(Match("cal", reversed))
	want := []string{"Reassign callbacks", "Open callbacks"}
	if !slices.Equal(got, want) {
		t.Errorf("Match kept %v, want %v — ties must not be reshuffled", got, want)
	}
}

// A title outranks its own keywords: the word a person typed at the front of
// the thing they are looking for beats a word somebody put in the back of it.
func TestMatchPrefersTheTitleToTheKeyword(t *testing.T) {
	items := []Command{
		{ID: "keyword", Title: "Search", Keywords: []string{"find"}},
		{ID: "title", Title: "Findings"},
	}
	got := titlesOf(Match("find", items))
	want := []string{"Findings", "Search"}
	if !slices.Equal(got, want) {
		t.Errorf("Match = %v, want %v — both are prefix hits, and the title's wins", got, want)
	}
}

// "CB-2871" is found by its number as well as its letters, which is what a
// person reading a callback off a phone does.
func TestMatchFindsAnIdentifierAtItsWordStart(t *testing.T) {
	items := []Command{
		{ID: "other", Title: "Open"},
		{ID: "cb", Title: "CB-2871"},
	}
	got := titlesOf(Match("2871", items))
	if len(got) != 1 || got[0] != "CB-2871" {
		t.Errorf("Match(\"2871\") = %v, want just CB-2871", got)
	}
}

// A disabled command matches and sorts like any other. It is greyed rather
// than hidden, so that the rows below it do not move as you type.
func TestMatchKeepsDisabledCommandsInPlace(t *testing.T) {
	got := titlesOf(Match("cal", commands))
	iOpen, iReassign := slices.Index(got, "Open callbacks"), slices.Index(got, "Reassign callbacks")
	if iReassign < 0 {
		t.Fatalf("Match dropped a disabled command: %v", got)
	}
	if iReassign < iOpen {
		t.Errorf("Match = %v: a disabled command moved ahead of an enabled one of the "+
			"same grade; greying it must not renumber the list", got)
	}
}

func TestMatchOnAnEmptyQueryKeepsEverythingInOrder(t *testing.T) {
	got := titlesOf(Match("   ", commands))
	want := titlesOf(commands)
	if !slices.Equal(got, want) {
		t.Errorf("Match(\"\") = %v, want the list as given: %v", got, want)
	}
}

func TestMatchDoesNotHandBackTheCallersSlice(t *testing.T) {
	got := Match("", commands)
	got[0] = Command{Title: "clobbered"}
	if commands[0].Title == "clobbered" {
		t.Error("Match handed back the caller's own slice; a palette sorting it in " +
			"place would reorder the app's command list for good")
	}
}

// ── Sidebar ─────────────────────────────────────────────────────────────────

func sidebarSections() []SidebarSection {
	open, repeat := 28, 3
	return []SidebarSection{
		{Title: "Views", Items: []SidebarItem{
			{Label: "All open", Icon: display.IconCallbacks, Count: &open, Selected: true},
			{Label: "Repeat failures", Icon: display.IconWarning, Count: &repeat, Tone: core.Danger},
			{Label: "Awaiting parts", Icon: display.IconClock, Disabled: true},
		}},
		{Title: "Team", Items: []SidebarItem{
			{Label: "Rosa Delgado", Icon: display.IconTeam},
			{Label: "Dana Reyes", Icon: display.IconTeam},
		}},
	}
}

func TestSidebarDrawsItsSectionsAndCounts(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Sidebar(c, sidebarSections(), SidebarOptions{Title: "Riverside Clinic"})
	}, 420, 700)
	for _, want := range []string{"Riverside Clinic", "Views", "All open", "28", "Repeat failures", "Team"} {
		wantShown(t, tt, want)
	}
}

func TestSidebarIsNamedAndAsWideAsTheThemeSays(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Sidebar(c, sidebarSections(), SidebarOptions{Title: "Riverside Clinic", Label: "Filters"})
	}, 700, 700)
	bar := findShown(t, tt, "Filters")
	if bar.W != theme.SidebarWidth {
		t.Errorf("the sidebar is %v wide, want the shared %v — a sidebar that "+
			"narrows with the window squeezes the labels it is for",
			bar.W, theme.SidebarWidth)
	}
}

// The fold is the caller's bool. The component reads it, reports the press,
// and never keeps a copy — so the two cannot disagree.
func TestSidebarFoldIsTheCallersBool(t *testing.T) {
	collapsed := false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		// The fold is flipped here, by the app, from what the sidebar
		// reported — the frame after the press. That loop is the whole of
		// the state: nothing in the component remembers it.
		if Sidebar(c, sidebarSections(), SidebarOptions{
			Title: "Riverside Clinic", Label: "Filters", Collapsed: &collapsed,
		}).Collapsed() {
			collapsed = !collapsed
		}
	}, 700, 700)
	bar := findShown(t, tt, "Filters")
	if bar.W != theme.SidebarWidth {
		t.Fatalf("an expanded sidebar is %v wide, want %v", bar.W, theme.SidebarWidth)
	}
	wantShown(t, tt, "Collapse sidebar")
	wantShown(t, tt, "Views")

	if err := tt.Click("Collapse sidebar"); err != nil {
		t.Fatal(err)
	}
	// The pointer moved. That is the assertion: the picture follows from it,
	// and a test that waited for the picture would be testing the picture.
	if !collapsed {
		t.Fatal("pressing the collapse control left the caller's bool false")
	}
}

func TestSidebarFoldedShowsIconsAndKeepsItsNames(t *testing.T) {
	collapsed := true
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Sidebar(c, sidebarSections(), SidebarOptions{
			Title: "Riverside Clinic", Label: "Filters", Collapsed: &collapsed,
		})
	}, 700, 700)
	bar := findShown(t, tt, "Filters")
	if bar.W != theme.RailWidth {
		t.Errorf("a folded sidebar is %v wide, want the rail's %v", bar.W, theme.RailWidth)
	}
	// Folded, there is no room for the words — and none for the headings,
	// which were only ever there to group the words.
	wantNotShown(t, tt, "Views")
	wantShown(t, tt, "Expand sidebar")
	// The rows are still named: a rail of glyphs with no names is a guessing
	// game for anyone not looking at them.
	for _, name := range []string{"All open", "Repeat failures", "Rosa Delgado"} {
		if _, ok := tt.Find(name); !ok {
			t.Errorf("the folded sidebar dropped the name %q", name)
		}
	}
}

func TestSidebarReportsTheCollapseRatherThanDoingIt(t *testing.T) {
	collapsed := false
	reported := false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		if Sidebar(c, sidebarSections(), SidebarOptions{
			Title: "Riverside Clinic", Collapsed: &collapsed,
		}).Collapsed() {
			reported = true
		}
	}, 700, 700)
	if err := tt.Click("Collapse sidebar"); err != nil {
		t.Fatal(err)
	}
	if !reported {
		t.Error("the press was not reported through the result")
	}
	if collapsed {
		t.Error("the sidebar folded itself; the state belongs to the caller")
	}
}

// The index is flat across the sections and counted in the caller's order,
// which is the key it switches views on — not where the row was drawn.
func TestSidebarReportsTheRowPressed(t *testing.T) {
	pressed := -1
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		if i := Sidebar(c, sidebarSections(), SidebarOptions{
			Title: "Riverside Clinic", Collapsed: new(bool),
		}).Pressed(); i >= 0 {
			pressed = i
		}
	}, 700, 700)
	if err := tt.Click("Dana Reyes"); err != nil {
		t.Fatal(err)
	}
	// "Dana Reyes" is the fifth row counted across both sections.
	if pressed != 4 {
		t.Errorf("pressing the fifth row reported %d; the index is flat across sections", pressed)
	}
}

func TestSidebarIgnoresADisabledRow(t *testing.T) {
	pressed := -1
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		if i := Sidebar(c, sidebarSections(), SidebarOptions{
			Title: "Riverside Clinic", Collapsed: new(bool),
		}).Pressed(); i >= 0 {
			pressed = i
		}
	}, 700, 700)
	wantShown(t, tt, "Awaiting parts")
	_ = tt.Click("Awaiting parts")
	if pressed >= 0 {
		t.Errorf("a disabled row reported index %d", pressed)
	}
}

// ── CommandPalette ──────────────────────────────────────────────────────────

func paletteTester(open *bool, query *string, hl *int, items []Command,
	dst *paletteLog) *ui.Tester {

	return ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		k := core.Tokens(c)
		// A Surface page, so the palette's Background panel can be told
		// from it: a panel the test cannot see is a panel it cannot test.
		ui.Column(c).Fill().Background(k.Surface).Padding(24).Gap(12).Children(func() {
			ui.Text(c, "Open callbacks").FontSize(theme.DisplaySize).Bold().TextColor(k.Text)
		})
		r := CommandPalette(c, open, items, CommandPaletteOptions{Query: query, Highlight: hl})
		dst.rows = titlesOf(r.Results())
		dst.total = r.Total()
		dst.highlight, dst.drawn = r.Highlight()
		if id, ok := r.Chosen(); ok {
			dst.chosen, dst.ran = id, true
		}
	}, 900, 600)
}

type paletteLog struct {
	rows      []string
	total     int
	highlight int
	drawn     int
	chosen    string
	ran       bool
}

func TestCommandPaletteDrawsWhatMatchedAndInWhatOrder(t *testing.T) {
	open, hl := true, 0
	query := "cal"
	var log paletteLog
	tt := paletteTester(&open, &query, &hl, commands, &log)
	wantShown(t, tt, "Command palette")
	wantShown(t, tt, "Type a command")

	want := []string{"Callbacks", "Open callbacks", "Reassign callbacks", "Team"}
	if !slices.Equal(log.rows, want) {
		t.Errorf("the palette drew %v, want %v", log.rows, want)
	}
	if log.total != len(want) {
		t.Errorf("Total() = %d, want %d — four commands match and all four fit",
			log.total, len(want))
	}
}

func TestCommandPaletteSaysWhenNothingMatched(t *testing.T) {
	open, hl := true, 0
	query := "zzzz"
	var log paletteLog
	tt := paletteTester(&open, &query, &hl, commands, &log)
	wantShown(t, tt, "No matching commands")
	if len(log.rows) != 0 {
		t.Errorf("the palette drew %v for a query nothing matches", log.rows)
	}
	if log.highlight != 0 || log.drawn != 0 {
		t.Errorf("the highlight is %d of %d rows; an empty list has no highlight",
			log.highlight, log.drawn)
	}
}

func TestCommandPaletteCapsTheHighlightAtTheRowsItDraws(t *testing.T) {
	items := make([]Command, 0, 12)
	for i := range 12 {
		items = append(items, Command{ID: string(rune('a' + i)), Title: "Command " + string(rune('a'+i))})
	}
	open, hl := true, 9
	query := ""
	var log paletteLog
	paletteTester(&open, &query, &hl, items, &log)
	if log.drawn != 8 {
		t.Fatalf("the palette drew %d rows, want the cap of 8", log.drawn)
	}
	if log.highlight != 0 || hl != 0 {
		t.Errorf("the highlight is %d (caller's %d); row 9 is not drawn, so it "+
			"cannot be the one Enter would run", log.highlight, hl)
	}
	if log.total != 12 {
		t.Errorf("Total() = %d, want the 12 that matched before the cap", log.total)
	}
}

// Down moves the highlight. The assertion is on the caller's own *int and on
// the number the result carries, because that is the state a keystroke
// changes — the highlight's colour is a consequence of it.
func TestCommandPaletteDownMovesTheHighlight(t *testing.T) {
	open, hl := true, 0
	query := ""
	var log paletteLog
	tt := paletteTester(&open, &query, &hl, commands, &log)

	tt.Key(0, ui.KeyDown)
	if hl != 1 {
		t.Errorf("after Down the caller's highlight is %d, want 1", hl)
	}
	if log.highlight != 1 {
		t.Errorf("the result reported highlight %d, want 1", log.highlight)
	}

	tt.Key(0, ui.KeyDown)
	if hl != 2 {
		t.Errorf("after a second Down the highlight is %d, want 2", hl)
	}

	tt.Key(0, ui.KeyUp)
	if hl != 1 {
		t.Errorf("after Up the highlight is %d, want 1", hl)
	}
}

// Down from the last row wraps to the first, so a person holding the key
// through nine commands is back where they started rather than stuck.
func TestCommandPaletteHighlightWraps(t *testing.T) {
	open, hl := true, 0
	query := ""
	var log paletteLog
	tt := paletteTester(&open, &query, &hl, commands, &log)
	_ = log

	tt.Key(0, ui.KeyUp)
	if hl != len(commands)-1 {
		t.Errorf("Up from the first row left the highlight at %d, want %d",
			hl, len(commands)-1)
	}
	tt.Key(0, ui.KeyDown)
	if hl != 0 {
		t.Errorf("Down from the last row left the highlight at %d, want 0", hl)
	}
}

// Enter runs the highlighted row, and the answer is the caller's ID rather
// than the title — a caller that switched on the title would break the moment
// somebody reworded it.
func TestCommandPaletteEnterRunsTheHighlightedCommand(t *testing.T) {
	open, hl := true, 0
	// Filtered, so that the second row is not the one the unfiltered list
	// happens to put second: the point is that Enter runs whichever row the
	// arrows lit, not a row anybody hard-coded.
	query := "cal"
	var log paletteLog
	tt := paletteTester(&open, &query, &hl, commands, &log)

	tt.Key(0, ui.KeyDown)
	if hl != 1 {
		t.Fatalf("Down left the highlight at %d, want 1", hl)
	}
	tt.Key(0, ui.KeyEnter)

	if !log.ran {
		t.Fatal("Enter ran nothing")
	}
	if log.chosen != "open-callback" {
		t.Errorf("Enter ran %q, want the second row's ID", log.chosen)
	}
	if open {
		t.Error("the palette stayed open after running a command")
	}
}

// A greyed row is not run by Enter: running the command below it would be
// worse than nothing happening.
func TestCommandPaletteEnterIgnoresADisabledRow(t *testing.T) {
	open, hl := true, 0
	query := "cal"
	var log paletteLog
	tt := paletteTester(&open, &query, &hl, commands, &log)

	// "Reassign callbacks" is the third row of the filtered list.
	hl = 2
	tt.Frame()
	tt.Key(0, ui.KeyEnter)

	if log.ran {
		t.Errorf("Enter ran the disabled command %q", log.chosen)
	}
	if !open {
		t.Error("the palette closed on a command it refused to run")
	}
}

// Escape is the layer's own: the palette is opened on ui/overlay's Dialog, so
// the same key that closes every other layer closes this one.
func TestCommandPaletteEscapeClosesIt(t *testing.T) {
	open, hl := true, 0
	query := ""
	var log paletteLog
	tt := paletteTester(&open, &query, &hl, commands, &log)
	wantShown(t, tt, "Type a command")

	tt.Key(0, ui.KeyEscape)
	if open {
		t.Fatal("Escape left the palette open")
	}
	wantNotShown(t, tt, "Type a command")
}

// A press on a row runs that row, whichever one it is: the pointer is a way
// of choosing too, not only the keyboard.
func TestCommandPaletteAPressRunsItsRow(t *testing.T) {
	open, hl := true, 0
	query := ""
	var log paletteLog
	tt := paletteTester(&open, &query, &hl, commands, &log)

	if err := tt.Click("Team"); err != nil {
		t.Fatal(err)
	}
	if !log.ran || log.chosen != "team" {
		t.Errorf("pressing the Team row ran %q (ran=%v), want \"team\"", log.chosen, log.ran)
	}
}

func TestCommandPaletteDrawsNothingWhileClosed(t *testing.T) {
	open, hl := false, 0
	query := ""
	var log paletteLog
	tt := paletteTester(&open, &query, &hl, commands, &log)
	if log.drawn != 0 {
		t.Errorf("a closed palette drew %d rows", log.drawn)
	}
	wantNotShown(t, tt, "Type a command")
}

func TestCommandPaletteNeedsSomethingToPointAt(t *testing.T) {
	query, hl := "", 0
	for name, build := range map[string]func(c *ui.Context){
		"no open state": func(c *ui.Context) {
			CommandPalette(c, nil, commands, CommandPaletteOptions{Query: &query, Highlight: &hl})
		},
		"no query": func(c *ui.Context) {
			open := true
			CommandPalette(c, &open, commands, CommandPaletteOptions{Highlight: &hl})
		},
		"no highlight": func(c *ui.Context) {
			open := true
			CommandPalette(c, &open, commands, CommandPaletteOptions{Query: &query})
		},
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Errorf("a palette with %s should panic: the component holds no state "+
						"of its own, so it has to be given somewhere to write", name)
				}
			}()
			ui.NewTester(func(c *ui.Context) {
				core.Use(c, core.Settings{})
				build(c)
			}, 600, 400)
		})
	}
}

// ── Menubar ─────────────────────────────────────────────────────────────────

func TestMenubarDrawsItsTitles(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Menubar(c, "Callbacks", []Menu{
			{Label: "File", Items: []MenuItem{{Label: "New callback"}, {Separator: true}, {Label: "Close"}}},
			{Label: "Edit", Items: []MenuItem{{Label: "Undo"}}},
		}, MenubarOptions{})
	}, 800, 300)
	for _, want := range []string{"Callbacks", "File", "Edit"} {
		wantShown(t, tt, want)
	}
	if _, ok := tt.Find("Menu bar"); !ok {
		t.Error("the bar of menus is not named as one thing")
	}
}

func TestMenubarReportsTheTitlePressed(t *testing.T) {
	opened := -1
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		if i := Menubar(c, "Callbacks", []Menu{
			{Label: "File", Items: []MenuItem{{Label: "New callback"}}},
			{Label: "Edit", Items: []MenuItem{{Label: "Undo"}}},
		}, MenubarOptions{}).Open(); i >= 0 {
			opened = i
		}
	}, 800, 300)
	if err := tt.Click("Edit"); err != nil {
		t.Fatal(err)
	}
	// Index 2: the application menu is the bar's first title, so File is 1.
	if opened != 2 {
		t.Errorf("pressing Edit reported title %d, want 2", opened)
	}
}

func TestMenubarNeedsAMenuToOpen(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("a bar with no menus should panic; it is a rule with nothing behind it")
		}
	}()
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Menubar(c, "", nil, MenubarOptions{})
	}, 800, 300)
}

// ── NavigationMenu ──────────────────────────────────────────────────────────

func TestNavigationMenuOpensAndReportsTheDestination(t *testing.T) {
	open := false
	chosen, answered := "", false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		k := core.Tokens(c)
		ui.Column(c).Fill().Background(k.Surface).Padding(24).Children(func() {
			ui.Text(c, "Open callbacks").FontSize(theme.DisplaySize).Bold().TextColor(k.Text)
		})
		if label, ok := NavigationMenu(c, "Go to", []NavigationMenuItem{
			{Label: "Callbacks", Description: "28 open", Selected: true},
			{Label: "Team", Description: "Six technicians"},
			{Label: "Archive", Disabled: true},
		}, NavigationMenuOptions{Open: &open, Modal: false}).Chosen(); ok {
			chosen, answered = label, true
		}
	}, 800, 600)

	wantShown(t, tt, "Go to")
	if open {
		t.Fatal("the menu was open before it was pressed")
	}
	if err := tt.Click("Go to"); err != nil {
		t.Fatal(err)
	}
	if !open {
		t.Fatal("pressing the trigger did not write into the caller's bool")
	}
	wantShown(t, tt, "Callbacks")
	wantShown(t, tt, "Six technicians")

	if err := tt.Click("Team"); err != nil {
		t.Fatal(err)
	}
	if !answered || chosen != "Team" {
		t.Errorf("pressing the Team row reported %q (answered=%v)", chosen, answered)
	}
}

func TestNavigationMenuEscapeClosesIt(t *testing.T) {
	open := true
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		k := core.Tokens(c)
		ui.Column(c).Fill().Background(k.Surface).Padding(24).Children(func() {
			ui.Text(c, "Open callbacks").FontSize(theme.DisplaySize).Bold().TextColor(k.Text)
		})
		NavigationMenu(c, "Go to", []NavigationMenuItem{
			{Label: "Callbacks"}, {Label: "Team"},
		}, NavigationMenuOptions{Open: &open, Modal: false})
	}, 800, 600)
	wantShown(t, tt, "Team")
	tt.Key(0, ui.KeyEscape)
	if open {
		t.Error("Escape left the navigation menu open")
	}
}

func TestNavigationMenuNeedsAnOpenState(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("a navigation menu with no *bool should panic: it would forget its " +
				"own open state on the next rebuild")
		}
	}()
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		NavigationMenu(c, "Go to", []NavigationMenuItem{{Label: "Callbacks"}}, NavigationMenuOptions{})
	}, 800, 600)
}

// ── toolbar ─────────────────────────────────────────────────────────────────

func TestToolbarGroupIsNamedAndCarriesItsButtons(t *testing.T) {
	pressed := ""
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Toolbar(c, ToolbarOptions{Label: "Board actions"}, func() {
			ToolbarGroup(c, ToolbarGroupOptions{Label: "Editing"}, func() {
				if ToolbarButton(c, "Assign", ToolbarButtonOptions{Icon: display.IconJobs}).Clicked() {
					pressed = "assign"
				}
			})
			ToolbarSeparator(c)
			ToolbarGroup(c, ToolbarGroupOptions{Label: "Exporting"}, func() {
				if ToolbarButton(c, "Export", ToolbarButtonOptions{Icon: display.IconDownload}).Clicked() {
					pressed = "export"
				}
			})
		})
	}, 700, 200)
	for _, want := range []string{"Assign", "Export"} {
		wantShown(t, tt, want)
	}
	for _, want := range []string{"Editing", "Exporting"} {
		if _, ok := tt.Find(want); !ok {
			t.Errorf("the run of buttons around %q is not named as a group", want)
		}
	}
	if err := tt.Click("Export"); err != nil {
		t.Fatal(err)
	}
	if pressed != "export" {
		t.Errorf("pressing Export reported %q, want \"export\"", pressed)
	}
}

func TestToolbarGroupNeedsALabel(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("an unnamed ToolbarGroup should panic: a run of buttons is a list, not a group")
		}
	}()
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ToolbarGroup(c, ToolbarGroupOptions{}, func() {
			ui.Text(c, "Assign")
		})
	}, 400, 160)
}

func TestToolbarButtonCarriesItsIconAndItsWord(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Toolbar(c, ToolbarOptions{Label: "Board actions"}, func() {
			ToolbarButton(c, "Assign", ToolbarButtonOptions{Icon: display.IconJobs})
		})
	}, 500, 200)
	// The glyph is named after the button it sits in, so the button is
	// findable by the word a reader would use for it; a name of the icon's
	// own would be read before the word that explains it.
	findShown(t, tt, "Assign")
}

// An icon-only button is allowed, but then the icon is the button's whole
// content and its Name is the only word it will ever be given.
func TestToolbarButtonIconOnlyNeedsAName(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("an icon-only ToolbarButton with no Label should panic")
		}
	}()
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Toolbar(c, ToolbarOptions{Label: "Board actions"}, func() {
			ToolbarButton(c, "", ToolbarButtonOptions{Icon: display.IconJobs})
		})
	}, 500, 200)
}

func TestToolbarButtonDoesNotMoveWhenDisabled(t *testing.T) {
	pressed := false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Toolbar(c, ToolbarOptions{Label: "Board actions"}, func() {
			if ToolbarButton(c, "Reassign", ToolbarButtonOptions{Disabled: true}).Clicked() {
				pressed = true
			}
		})
	}, 500, 200)
	wantShown(t, tt, "Reassign")
	_ = tt.Click("Reassign")
	if pressed {
		t.Error("a disabled toolbar button fired")
	}
}

// ── the dark palette ────────────────────────────────────────────────────────

func TestNavigationSurvivesTheDarkPalette(t *testing.T) {
	for _, mode := range []core.Mode{core.Light, core.Dark} {
		open, hl := true, 0
		query := "cal"
		var log paletteLog
		tt := ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{Mode: mode})
			k := core.Tokens(c)
			ui.Column(c).Fill().Background(k.Surface).Padding(24).Children(func() {
				ui.Text(c, "Open callbacks").FontSize(theme.DisplaySize).Bold().TextColor(k.Text)
			})
			Sidebar(c, sidebarSections(), SidebarOptions{
				Title: "Riverside Clinic", Label: "Filters", Collapsed: new(bool),
			})
			Menubar(c, "Callbacks", []Menu{{Label: "File"}}, MenubarOptions{})
			log.rows = titlesOf(CommandPalette(c, &open, commands,
				CommandPaletteOptions{Query: &query, Highlight: &hl}).Results())
		}, 900, 700)
		for _, want := range []string{"Riverside Clinic", "Views", "Callbacks", "Open callbacks"} {
			wantShown(t, tt, want)
		}
		// The order is a claim about the interface, not about its colours:
		// it must be the same in both appearances or one of them is lying.
		want := []string{"Callbacks", "Open callbacks", "Reassign callbacks", "Team"}
		if !slices.Equal(log.rows, want) {
			t.Errorf("%v mode matched %v, want %v", mode, log.rows, want)
		}
	}
}
