package data

import (
	"fmt"
	"image"
	"image/color"
	"slices"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// This file covers the row views of the package: the two pure functions they
// are built on, the components themselves, and the three rules that make a
// table a table — its columns keep their width, its head stays, and only the
// rows in view are drawn.

// ── the pure functions ─────────────────────────────────────────────────────
// They are here rather than behind a component because neither a screenshot
// nor a drawn word can show what they get wrong: an unstable sort and a
// Shift-click measured from the top of the table both look perfect on screen.

type record struct {
	name  string
	issue string
	cost  float64
	// seq is never compared, so that a test can tell two rows that sort
	// equal apart by where they came from.
	seq int
}

// order is the names of the rows in the order they came back.
func order(rows []record) []string {
	out := make([]string, len(rows))
	for i, r := range rows {
		out[i] = fmt.Sprintf("%d:%s", r.seq, r.name)
	}
	return out
}

func equal(got, want []string) bool { return slices.Equal(got, want) }

func equalInts(got, want []int) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func comparators() map[string]func(a, b record) int {
	return map[string]func(a, b record) int{
		"name": ByText(func(r record) string { return r.name }),
		"cost": ByNumber(func(r record) float64 { return r.cost }),
	}
}

// unsorted is six callbacks, three of them costing the same.
func unsorted() []record {
	return []record{
		{seq: 0, name: "Maple Street Bakery", issue: "AC repair", cost: 310},
		{seq: 1, name: "Hillside Dental", issue: "Leak", cost: 90},
		{seq: 2, name: "Cedar Clinic", issue: "Boiler", cost: 310},
		{seq: 3, name: "Northgate Vet", issue: "AC repair", cost: 45},
		{seq: 4, name: "Anchor Pharmacy", issue: "Furnace", cost: 310},
		{seq: 5, name: "Birch Lane Gym", issue: "Leak", cost: 120},
	}
}

func TestRowsOrdersByTheColumnAndTheDirection(t *testing.T) {
	rows := unsorted()
	asc := Rows(rows, Sort{Column: "cost"}, comparators())
	if got, want := order(asc), []string{"3:Northgate Vet", "1:Hillside Dental", "5:Birch Lane Gym", "0:Maple Street Bakery", "2:Cedar Clinic", "4:Anchor Pharmacy"}; !equal(got, want) {
		t.Errorf("ascending by cost: %v, want %v", got, want)
	}
	desc := Rows(rows, Sort{Column: "cost", Descending: true}, comparators())
	if got, want := order(desc), []string{"0:Maple Street Bakery", "2:Cedar Clinic", "4:Anchor Pharmacy", "5:Birch Lane Gym", "1:Hillside Dental", "3:Northgate Vet"}; !equal(got, want) {
		t.Errorf("descending by cost: %v, want %v", got, want)
	}
	if got, want := order(Rows(rows, Sort{Column: "name"}, comparators())),
		[]string{"4:Anchor Pharmacy", "5:Birch Lane Gym", "2:Cedar Clinic", "1:Hillside Dental", "0:Maple Street Bakery", "3:Northgate Vet"}; !equal(got, want) {
		t.Errorf("by name: %v, want %v", got, want)
	}
}

// TestRowsKeepsEqualRowsWhereTheyWere is the whole reason Rows is not sort.Slice:
// three callbacks cost the same, and the order they were loaded in is a fact
// about them. A sort that reshuffles them makes a table that was sorted
// twice read differently from one sorted once, for no reason a user can see.
func TestRowsKeepsEqualRowsWhereTheyWere(t *testing.T) {
	rows := unsorted()
	asc := Rows(rows, Sort{Column: "cost"}, comparators())
	want := []string{"3:Northgate Vet", "1:Hillside Dental", "5:Birch Lane Gym", "0:Maple Street Bakery", "2:Cedar Clinic", "4:Anchor Pharmacy"}
	if got := order(asc); !equal(got, want) {
		t.Errorf("ascending: %v, want %v", got, want)
	}
	// The three rows costing 310 are the first, third and fifth of the
	// input: in that order, both ways up.
	desc := Rows(rows, Sort{Column: "cost", Descending: true}, comparators())
	wantDesc := []string{"0:Maple Street Bakery", "2:Cedar Clinic", "4:Anchor Pharmacy", "5:Birch Lane Gym", "1:Hillside Dental", "3:Northgate Vet"}
	if got := order(desc); !equal(got, wantDesc) {
		t.Errorf("descending: %v, want %v", got, wantDesc)
	}
	for _, at := range []struct{ row, seq int }{{3, 0}, {4, 2}, {5, 4}} {
		if asc[at.row].seq != at.seq {
			t.Errorf("row %d of the ascending order came from %d, want %d: %v",
				at.row, asc[at.row].seq, at.seq, order(asc))
		}
	}
	// The same three rows, in the same order among themselves, at the top
	// where the biggest costs are.
	for _, at := range []struct{ row, seq int }{{0, 0}, {1, 2}, {2, 4}} {
		if desc[at.row].seq != at.seq {
			t.Errorf("row %d of the descending order came from %d, want %d: %v",
				at.row, desc[at.row].seq, at.seq, order(desc))
		}
	}
	// Sorting twice is sorting once: the caller's rows come back, and the
	// order they are in again.
	again := Rows(asc, Sort{Column: "cost"}, comparators())
	if got := order(again); !equal(got, want) {
		t.Errorf("sorting the sorted rows: %v, want %v", got, want)
	}
}

func TestRowsNeverTouchesTheCallersSlice(t *testing.T) {
	rows := unsorted()
	before := order(rows)
	_ = Rows(rows, Sort{Column: "cost"}, comparators())
	if got := order(rows); !equal(got, before) {
		t.Errorf("Rows reordered the caller's slice: %v, want %v", got, before)
	}
}

// TestRowsWithNoColumnIsTheCallersOrder: an unsorted table is what the
// caller has, not an empty table and not a sorted one.
func TestRowsWithNoColumnIsTheCallersOrder(t *testing.T) {
	rows := unsorted()
	if got, want := order(Rows(rows, Sort{}, comparators())), order(rows); !equal(got, want) {
		t.Errorf("with no column: %v, want %v", got, want)
	}
}

func TestRowsPanicsForAColumnItCannotCompare(t *testing.T) {
	defer func() {
		msg, ok := recover().(string)
		if !ok || !strings.Contains(msg, "Rows cannot sort by \"due\"") {
			t.Errorf("a column with no comparator: %v", msg)
		}
	}()
	_ = Rows(unsorted(), Sort{Column: "due"}, comparators())
}

func TestToggleSortsAColumnThenReversesIt(t *testing.T) {
	var s Sort
	s = Toggle(s, "cost")
	if s != (Sort{Column: "cost"}) {
		t.Errorf("the first click on a column: %+v", s)
	}
	s = Toggle(s, "cost")
	if s != (Sort{Column: "cost", Descending: true}) {
		t.Errorf("the second click on the same column: %+v", s)
	}
	s = Toggle(s, "name")
	if s != (Sort{Column: "name"}) {
		t.Errorf("a click on another column: %+v", s)
	}
	s = Toggle(s, "cost")
	if s != (Sort{Column: "cost"}) {
		t.Errorf("back to a column sorted the other way: %+v", s)
	}
}

// TestSelectableReachesFromTheRowLastClicked is the rule a Shift-click is
// judged by: click row 0, then Shift-click row 3, and rows 0 to 3 are
// chosen — not the rows 0 to 2 before it, which is what a range measured
// from the top of the table would give on a table scrolled down.
func TestSelectableReachesFromTheRowLastClicked(t *testing.T) {
	var s Selectable
	if s.Anchor() != -1 {
		t.Errorf("a fresh choice has the anchor %d, want -1", s.Anchor())
	}
	s.Click(0, false, false)
	if got, want := s.Rows(), []int{0}; !equalInts(got, want) {
		t.Errorf("after clicking row 0: %v, want %v", got, want)
	}
	if s.Anchor() != 0 {
		t.Errorf("the anchor after clicking row 0: %d", s.Anchor())
	}
	s.Click(3, false, true)
	if got, want := s.Rows(), []int{0, 1, 2, 3}; !equalInts(got, want) {
		t.Errorf("after Shift-clicking row 3: %v, want %v", got, want)
	}
	// The anchor is where it was: a second Shift-click reaches from row 0
	// again and narrows the range, rather than from row 3 widening it.
	s.Click(1, false, true)
	if got, want := s.Rows(), []int{0, 1}; !equalInts(got, want) {
		t.Errorf("after Shift-clicking row 1: %v, want %v", got, want)
	}
	if s.Anchor() != 0 {
		t.Errorf("the anchor moved to %d after a Shift-click, want 0", s.Anchor())
	}
}

func TestSelectableShiftClickBeforeAnyClickChoosesTheRowAlone(t *testing.T) {
	var s Selectable
	s.Click(3, false, true)
	if got, want := s.Rows(), []int{3}; !equalInts(got, want) {
		t.Errorf("a Shift-click with nothing chosen: %v, want %v", got, want)
	}
	if s.Anchor() != 3 {
		t.Errorf("the anchor after that click: %d, want 3", s.Anchor())
	}
}

func TestSelectableAddsAndTakesRows(t *testing.T) {
	var s Selectable
	s.Click(1, false, false)
	s.Click(4, true, false)
	if got, want := s.Rows(), []int{1, 4}; !equalInts(got, want) {
		t.Errorf("after Cmd-clicking row 4: %v, want %v", got, want)
	}
	if s.Anchor() != 4 {
		t.Errorf("the anchor after an added click: %d, want 4", s.Anchor())
	}
	s.Click(1, true, false)
	if got, want := s.Rows(), []int{4}; !equalInts(got, want) {
		t.Errorf("after Cmd-clicking row 1 again: %v, want %v", got, want)
	}
	// A plain click throws the rest of the choice out.
	s.Click(6, false, false)
	if got, want := s.Rows(), []int{6}; !equalInts(got, want) {
		t.Errorf("after a plain click on row 6: %v, want %v", got, want)
	}
	if s.Len() != 1 {
		t.Errorf("the choice holds %d rows, want 1", s.Len())
	}
}

// TestSelectableShiftClickWithCmdKeepsWhatWasChosen: Shift with Cmd adds a
// range to the choice rather than replacing it.
func TestSelectableShiftClickWithCmdKeepsWhatWasChosen(t *testing.T) {
	var s Selectable
	s.Click(0, false, false)
	s.Click(3, true, false)
	if got, want := s.Rows(), []int{0, 3}; !equalInts(got, want) {
		t.Fatalf("after Cmd-clicking row 3: %v, want %v", got, want)
	}
	s.Click(5, true, true)
	if got, want := s.Rows(), []int{0, 3, 4, 5}; !equalInts(got, want) {
		t.Errorf("after Shift-clicking row 5 with Cmd: %v, want %v", got, want)
	}
	// The same Shift-click without Cmd takes the range in place of what was
	// chosen, which is what a Shift-click does on its own.
	s.Click(7, false, true)
	if got, want := s.Rows(), []int{3, 4, 5, 6, 7}; !equalInts(got, want) {
		t.Errorf("after Shift-clicking row 7: %v, want %v", got, want)
	}
}

// TestSelectableInvertChoosesTheComplement is the toolbar's Invert, and it is
// only the complement if it is over the rows there are.
func TestSelectableInvertChoosesTheComplement(t *testing.T) {
	var s Selectable
	s.Click(1, false, false)
	s.Click(3, false, true)
	if got, want := s.Rows(), []int{1, 2, 3}; !equalInts(got, want) {
		t.Fatalf("before inverting: %v, want %v", got, want)
	}
	s.Invert(6)
	if got, want := s.Rows(), []int{0, 4, 5}; !equalInts(got, want) {
		t.Errorf("after inverting over six rows: %v, want %v", got, want)
	}
	s.Invert(6)
	if got, want := s.Rows(), []int{1, 2, 3}; !equalInts(got, want) {
		t.Errorf("inverting again: %v, want %v", got, want)
	}
	// A table of four rows is a different table: the complement is taken
	// over the rows there are, not over the choice alone.
	var short Selectable
	short.All(4)
	short.Invert(4)
	if short.Len() != 0 {
		t.Errorf("inverting a choice of everything: %d rows left", short.Len())
	}
}

func TestSelectableAllAndNone(t *testing.T) {
	var s Selectable
	s.All(5)
	if got, want := s.Rows(), []int{0, 1, 2, 3, 4}; !equalInts(got, want) {
		t.Errorf("after choosing all: %v, want %v", got, want)
	}
	s.None()
	if s.Len() != 0 {
		t.Errorf("after choosing none: %d rows", s.Len())
	}
	if s.Has(0) {
		t.Error("row 0 is still chosen after None")
	}
}

// ── helpers ───────────────────────────────────────────────────────────────

func callback(n int) record {
	return record{
		name:  fmt.Sprintf("Customer %02d", n),
		issue: "AC repair",
		cost:  float64((n*37)%90) + 10,
		seq:   n,
	}
}

func callbacks(n int) []record {
	out := make([]record, n)
	for i := range out {
		out[i] = callback(i)
	}
	return out
}

// boardColumns is the list view the design system describes: two names that
// share the room, three fixed columns, and a figure against the edge.
func boardColumns() []Column {
	return []Column{
		{Title: "Customer", ID: "customer", Share: 2, Sortable: true},
		{Title: "Issue", ID: "issue", Share: 2, Sortable: true},
		{Title: "Priority", ID: "priority", Width: 110, Sortable: true},
		{Title: "Technician", ID: "tech", Width: 130},
		{Title: "Due", ID: "due", Width: 120},
		{Title: "Cost", ID: "cost", Width: 120, Align: ui.End, Sortable: true},
	}
}

// board draws the callbacks list: the table with a chosen set and a row the
// keys move from.
type board struct {
	rows   []record
	cols   []Column
	state  ui.ListState
	scroll ui.ScrollState
	sort   Sort
	lead   int
	choice Selectable
	// built counts the cells built in the frame being drawn, which is how a
	// test sees that a table of thousands draws tens of rows.
	built int
	// sorted is the column a head asked for. A press is a question the view
	// answers in the frame it is asked in, so an app that wants to know
	// which column was clicked notes it there, and so does this.
	sorted string
	// empty makes the table draw this instead of its rows, for the frame
	// with nothing in it.
	empty string
}

func newBoard(n int) *board {
	return &board{rows: callbacks(n), cols: boardColumns(), lead: -1}
}

func (b *board) view(dark bool) func(c *ui.Context) {
	return func(c *ui.Context) {
		mode := core.Light
		if dark {
			mode = core.Dark
		}
		core.Use(c, core.Settings{Mode: mode})
		k, u := core.Tokens(c), core.Density(c).Unit()
		ui.Column(c).Fill().Background(k.Background).Children(func() {
			b.built = 0
			res := DataTable(c, DataTableOptions{
				Columns: b.cols, Rows: len(b.rows),
				Height: 400, State: &b.state, Scroll: &b.scroll,
				Sort: &b.sort, Selected: &b.lead, Choice: &b.choice,
				Key:   func(i int) any { return b.rows[i].name },
				Label: func(i int) string { return b.rows[i].name },
				CellLabel: func(i, col int) string {
					return fmt.Sprintf("%s: %s", b.cols[col].Title, b.rows[i].name)
				},
				Empty: func() {
					if b.empty != "" {
						ui.Text(c, b.empty)
					}
				},
				Cell: func(i, col int) {
					b.built++
					r := b.rows[i]
					switch col {
					case 0:
						ui.Text(c, r.name).SingleLine()
					case 1:
						ui.Text(c, r.issue).SingleLine()
					case 2:
						ui.Text(c, "High")
					case 3:
						ui.Text(c, "Nate Coleman").SingleLine()
					case 4:
						ui.Text(c, "Tue 14").SingleLine()
					case 5:
						ui.Text(c, fmt.Sprintf("$%.0f", r.cost))
					}
				},
			})
			if res.Sorted() != "" {
				b.sorted = res.Sorted()
			}
			ui.Box(c).Height(u)
		})
	}
}

func panics(t *testing.T, what string, fn func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("%s did not panic", what)
			return
		}
		if msg, ok := r.(string); !ok || !strings.HasPrefix(msg, "data: ") {
			t.Errorf("%s panicked with %v, want a data: message", what, r)
		}
	}()
	fn()
}

// ── DataTable ──────────────────────────────────────────────────────────────

func TestDataTableShowsItsRowsAndItsHead(t *testing.T) {
	b := newBoard(60)
	tt := ui.NewTester(b.view(false), 1100, 500)
	for _, want := range []string{
		"Customer", "Issue", "Priority", "Technician", "Due", "Cost",
		"Customer 00", "AC repair", "$10",
	} {
		if !tt.HasText(want) {
			t.Errorf("the table is missing %q", want)
		}
	}
	if tt.HasText("Customer 60") {
		t.Error("the table shows a row it does not have")
	}
	// Every column is named, and every row of it is a cell of its own.
	for _, col := range b.cols {
		if _, ok := tt.Find(col.Title); !ok {
			t.Errorf("no head for the %s column", col.Title)
		}
	}
	if _, ok := tt.Find("Cost: Customer 00"); !ok {
		t.Errorf("a cell is not named, so a screen reader cannot read the row: %q", tt.Texts())
	}
}

// TestDataTableColumnsKeepTheirWidthInANarrowWindow is the rule the board's
// columns follow and the reason a table scrolls sideways: a column that has
// a width keeps it, and a table that does not fit scrolls rather than
// squeezing every column into a ribbon of letters.
func TestDataTableColumnsKeepTheirWidthInANarrowWindow(t *testing.T) {
	b := newBoard(60)
	tt := ui.NewTester(b.view(false), 1100, 500)

	width := func(title string) float32 {
		t.Helper()
		r, ok := tt.Find(title)
		if !ok || r.W == 0 {
			t.Fatalf("no box for %q in the head: %v", title, r)
		}
		return r.W
	}
	wide := map[string]float32{}
	for _, col := range b.cols {
		if col.Width > 0 {
			wide[col.Title] = width(col.Title)
		}
	}
	if len(wide) == 0 {
		t.Fatal("the columns under test have no fixed widths")
	}

	// A window too narrow for the columns.
	tt.SetSize(420, 500)
	tt.Frame()
	// The columns still on screen keep the widths they had; the ones that
	// have been pushed off it are measured after scrolling to them, because
	// a column off the edge of the window has nothing to measure.
	for _, title := range []string{"Priority", "Technician"} {
		if after := width(title); after != wide[title] {
			t.Errorf("in a 420-wide window the %s column is %v wide, was %v",
				title, after, wide[title])
		}
	}
	// A cell is as wide as its column, whatever the window.
	if cell, ok := tt.Find("Priority: Customer 00"); !ok || cell.W != wide["Priority"] {
		t.Errorf("a Priority cell is %v wide, want the column's %v", cell.W, wide["Priority"])
	}
	// The table scrolls sideways, because it is wider than the window: it is
	// the rows that move, not the columns that shrink.
	if b.scroll.MaxX <= 0 {
		t.Fatalf("a table wider than its window does not scroll sideways: MaxX %v", b.scroll.MaxX)
	}
	// A shared column stops at its floor rather than being squeezed thinner:
	// 60 DIPs is about two words of a row, and below that a name is letters.
	if got := width("Customer"); got < 60 {
		t.Errorf("a shared column squeezed to %v", got)
	}

	b.scroll.X = b.scroll.MaxX
	tt.Frame()
	if after := width("Cost"); after != wide["Cost"] {
		t.Errorf("scrolled to its end, the Cost column is %v wide, want %v", after, wide["Cost"])
	}
	if after := width("Due"); after != wide["Due"] {
		t.Errorf("scrolled to its end, the Due column is %v wide, want %v", after, wide["Due"])
	}
}

// TestDataTableHeadStaysWhileTheRowsScroll is what a head is for: it is the
// one part of the table that does not move, because the columns it names
// are the only thing that says what a row of figures means.
func TestDataTableHeadStaysWhileTheRowsScroll(t *testing.T) {
	b := newBoard(400)
	tt := ui.NewTester(b.view(false), 1100, 500)

	head, ok := tt.Find("Priority")
	if !ok {
		t.Fatal("the head did not lay out")
	}
	first, ok := tt.Find("Customer: Customer 00")
	if !ok {
		t.Fatal("the first row did not lay out")
	}
	tt.Scroll(400, 300, 0, 240)
	tt.Frame()

	afterHead, _ := tt.Find("Priority")
	if afterHead.Y != head.Y {
		t.Errorf("the head moved from y=%v to y=%v as the rows scrolled", head.Y, afterHead.Y)
	}
	// The rows did move: the list is showing a later row at the top.
	lo, hi := b.state.Visible()
	if lo <= 0 || hi-lo < 3 {
		t.Errorf("after scrolling down 240 the rows in view are %d..%d, want the fifth or so on", lo, hi)
	}
	if first.H > 0 && tt.HasText("Customer: Customer 00") {
		top, ok := tt.Find("Customer: Customer 00")
		if ok && top.Y == first.Y {
			t.Error("the rows did not move when the table was scrolled")
		}
	}
}

// TestDataTableBuildsOnlyTheRowsInView is the cost of a table of thousands:
// the rows are not drawn, only the ones on show.
func TestDataTableBuildsOnlyTheRowsInView(t *testing.T) {
	const n = 5000
	b := newBoard(n)
	tt := ui.NewTester(b.view(false), 1100, 500)

	lo, hi := b.state.Visible()
	if lo != 0 {
		t.Errorf("a table showing its first rows shows %d..%d", lo, hi)
	}
	// A 400 DIP window of 48 DIP rows: a handful of them, plus the few a
	// list builds beyond each edge for the focus and for the reader.
	if shown := hi - lo + 1; shown > 20 {
		t.Errorf("the table showed %d rows at once, want at most 20", shown)
	}
	if b.built > 80 {
		t.Errorf("a frame built %d cells for a window showing at most 20 rows", b.built)
	}
	if b.built < 6 {
		t.Errorf("a frame built %d cells, which is too few to be the rows in view", b.built)
	}
	// The rows in view are the ones being drawn, and the rows after them are
	// not: scrolling brings the next ones in.
	tt.Scroll(400, 300, 0, 480)
	tt.Frame()
	lo, hi = b.state.Visible()
	if lo <= 0 {
		t.Errorf("after scrolling the rows in view are still %d..%d", lo, hi)
	}
	if b.built > 80 {
		t.Errorf("a frame after scrolling built %d cells", b.built)
	}
}

// TestDataTableAsksToSortAndTheCallerOrdersItsRows: a click on a head writes
// the sort, and the caller reorders with Rows — the table holds no order of
// its own, so the two can never disagree about which column is sorted.
func TestDataTableAsksToSortAndTheCallerOrdersItsRows(t *testing.T) {
	b := newBoard(40)
	tt := ui.NewTester(b.view(false), 1100, 500)

	if err := tt.Click("Cost"); err != nil {
		t.Fatal(err)
	}
	if b.sort != (Sort{Column: "cost"}) {
		t.Errorf("after one click on Cost: %+v", b.sort)
	}
	if b.sorted != "cost" {
		t.Errorf("the frame the click was in reported the column %q", b.sorted)
	}
	if err := tt.Click("Cost"); err != nil {
		t.Fatal(err)
	}
	if b.sort != (Sort{Column: "cost", Descending: true}) {
		t.Errorf("after a second click on Cost: %+v", b.sort)
	}
	// A column that is not sortable asks for nothing.
	before := b.sort
	if err := tt.Click("Technician"); err != nil {
		t.Fatal(err)
	}
	if b.sort != before {
		t.Errorf("a click on a column that does not sort: %+v", b.sort)
	}
	// The order the caller draws with is the order the head says: the rows
	// come back biggest first, and nothing else moves.
	b.rows = Rows(b.rows, b.sort, map[string]func(a, b record) int{
		"cost": ByNumber(func(r record) float64 { return r.cost }),
	})
	if b.sort.Descending {
		for i := 1; i < len(b.rows); i++ {
			if b.rows[i-1].cost < b.rows[i].cost {
				t.Fatalf("the rows are not biggest first: %v", costs(b.rows))
			}
		}
	} else {
		for i := 1; i < len(b.rows); i++ {
			if b.rows[i-1].cost > b.rows[i].cost {
				t.Fatalf("the rows are not smallest first: %v", costs(b.rows))
			}
		}
	}
	if len(b.rows) != 40 {
		t.Errorf("sorting lost rows: %d of 40", len(b.rows))
	}
}

func costs(rows []record) []float64 {
	out := make([]float64, len(rows))
	for i, r := range rows {
		out[i] = r.cost
	}
	return out
}

// TestDataTableChoosesRowsWithTheClickThatWasAskedFor asserts the pointers
// the caller keeps, not the words on screen: what a click chose, and what a
// Shift-click reached for.
func TestDataTableChoosesRowsWithTheClickThatWasAskedFor(t *testing.T) {
	b := newBoard(40)
	tt := ui.NewTester(b.view(false), 1100, 500)

	if err := tt.Click("Customer 00"); err != nil {
		t.Fatal(err)
	}
	if got, want := b.choice.Rows(), []int{0}; !equalInts(got, want) {
		t.Fatalf("after clicking row 0: %v, want %v", got, want)
	}
	if b.lead != 0 {
		t.Errorf("the row the keys move from is %d, want 0", b.lead)
	}

	if err := tt.ClickWith(ui.Shift, "Customer 03"); err != nil {
		t.Fatal(err)
	}
	if got, want := b.choice.Rows(), []int{0, 1, 2, 3}; !equalInts(got, want) {
		t.Errorf("after Shift-clicking row 3: %v, want %v", got, want)
	}

	if err := tt.ClickWith(ui.Cmd, "Customer 03"); err != nil {
		t.Fatal(err)
	}
	if got, want := b.choice.Rows(), []int{0, 1, 2}; !equalInts(got, want) {
		t.Errorf("after Cmd-clicking row 3 again: %v, want %v", got, want)
	}

	// A plain click chooses that row alone.
	if err := tt.Click("Customer 01"); err != nil {
		t.Fatal(err)
	}
	if got, want := b.choice.Rows(), []int{1}; !equalInts(got, want) {
		t.Errorf("after a plain click on row 1: %v, want %v", got, want)
	}
	// The keys move the row they move from, and it is the pointer that says
	// so: a table is navigated from the same place it is drawn from.
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyDown)
	if b.lead != 3 {
		t.Errorf("two presses of Down moved the row to %d, want 3", b.lead)
	}
}

// TestDataTableInDarkMode looks at the pixels, because "it follows the
// system" is not a thing a text assertion can say: the chosen row is painted
// the dark accent, not the light one.
func TestDataTableInDarkMode(t *testing.T) {
	b := newBoard(20)
	b.choice.Click(1, false, false)
	b.choice.Click(3, false, true)
	tt := ui.NewTester(b.view(true), 1100, 500)

	chosen, ok := tt.Find("Customer: Customer 01")
	if !ok || chosen.W == 0 {
		t.Fatalf("a chosen row did not lay out: %v", chosen)
	}
	plain, ok := tt.Find("Customer: Customer 05")
	if !ok || plain.W == 0 {
		t.Fatalf("a row that is not chosen did not lay out: %v", plain)
	}
	img := tt.Image()
	dark := theme.Dark()
	if !finds(img, inset(chosen, 1, 0.3, 0.6), dark.AccentBg) {
		t.Errorf("a chosen row is not painted %v in the dark palette", dark.AccentBg)
	}
	if finds(img, inset(plain, 1, 0.3, 0.6), dark.AccentBg) {
		t.Errorf("a row that is not chosen is painted as though it were")
	}
	if !finds(img, inset(plain, 1, 0.3, 0.6), dark.Background) {
		t.Errorf("the rows are not painted over %v", dark.Background)
	}
	// The same table in the light palette is the light one.
	light := newBoard(20)
	light.choice.Click(1, false, false)
	lt := ui.NewTester(light.view(false), 1100, 500)
	lrow, _ := lt.Find("Customer: Customer 01")
	if !finds(lt.Image(), inset(lrow, 1, 0.3, 0.6), theme.Light().AccentBg) {
		t.Errorf("a chosen row is not painted %v in the light palette", theme.Light().AccentBg)
	}
}

// TestDataTableWithOneChosenRowTakesTheLibrarysOwn: a table that chooses one
// row paints it the way every other list in the library paints a chosen row,
// rather than with the table's own idea of one — there are two looks for a
// chosen row in an app otherwise, and only one of them is the system's.
func TestDataTableWithOneChosenRowTakesTheLibrarysOwn(t *testing.T) {
	rows := callbacks(20)
	var state ui.ListState
	lead := 3
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ui.Column(c).Fill().Children(func() {
			DataTable(c, DataTableOptions{
				Columns: boardColumns(), Rows: len(rows), Height: 320,
				State: &state, Selected: &lead,
				CellLabel: func(i, col int) string {
					return fmt.Sprintf("%s: %s", boardColumns()[col].Title, rows[i].name)
				},
				Cell: func(i, col int) { ui.Text(c, rows[i].name).SingleLine() },
			})
		})
	}, 1100, 500)

	chosen, ok := tt.Find("Customer: Customer 03")
	if !ok || chosen.W == 0 {
		t.Fatalf("the chosen row did not lay out: %v", chosen)
	}
	if !finds(tt.Image(), inset(chosen, 1, 0.3, 0.6), theme.Light().Accent) {
		t.Errorf("the chosen row is not painted %v", theme.Light().Accent)
	}
	if err := tt.Click("Customer 05"); err != nil {
		t.Fatal(err)
	}
	if lead != 5 {
		t.Errorf("a click chose row %d, want 5", lead)
	}
}

// TestDataTableWithNoRowsDrawsTheEmptyState says it gave: a table with
// nothing in it draws the caller's explanation, and does not wrap it in a
// scroll view, where a control in it would be hard to press.
func TestDataTableWithNoRowsDrawsTheEmptyState(t *testing.T) {
	b := newBoard(0)
	b.empty = "No callbacks match"
	tt := ui.NewTester(b.view(false), 900, 500)
	if !tt.HasText("No callbacks match") {
		t.Errorf("a table with no rows drew %q", tt.Texts())
	}
}

func TestDataTablePanicsForWhatItCannotBeGiven(t *testing.T) {
	panics(t, "a table with no columns", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			DataTable(c, DataTableOptions{Rows: 1, Height: 100,
				Cell: func(int, int) {}})
		}, 400, 300)
	})
	panics(t, "a table with no cells to build", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			DataTable(c, DataTableOptions{
				Columns: boardColumns(), Rows: 1, Height: 100})
		}, 400, 300)
	})
	panics(t, "a table sorted by a column it does not have", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			DataTable(c, DataTableOptions{
				Columns: boardColumns(), Rows: 1, Height: 100,
				Sort: &Sort{Column: "due date"},
				Cell: func(int, int) {},
			})
		}, 400, 300)
	})
	panics(t, "a table with no height", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			DataTable(c, DataTableOptions{
				Columns: boardColumns(), Rows: 1,
				Cell: func(int, int) {},
			})
		}, 400, 300)
	})
}

// ── List ───────────────────────────────────────────────────────────────────

func TestListShowsRowsOfOneHeightAndChoosesThem(t *testing.T) {
	var state ui.ListState
	var lead = -1
	var choice Selectable
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		List(c, ListOptions{
			Rows: 500, Height: 300, State: &state,
			Selected: &lead, Choice: &choice,
			Label: func(i int) string { return fmt.Sprintf("Row %d", i) },
		}, func(i int) {
			ui.Text(c, fmt.Sprintf("Row %d", i)).SingleLine()
		})
	}, 600, 400)

	if !tt.HasText("Row 0") {
		t.Fatalf("the list drew %q", tt.Texts())
	}
	if tt.HasText("Row 499") {
		t.Error("a list of 500 rows drew all of them")
	}
	lo, hi := state.Visible()
	if lo != 0 || hi-lo > 10 {
		t.Errorf("the rows in view are %d..%d", lo, hi)
	}
	first, ok := tt.Find("Row 0")
	if !ok {
		t.Fatal("the first row is not in the frame")
	}
	second, _ := tt.Find("Row 1")
	if second.Y-first.Y <= 0 {
		t.Errorf("rows are not laid out one under the other: %v then %v", first, second)
	}
	if err := tt.Click("Row 2"); err != nil {
		t.Fatal(err)
	}
	if got, want := choice.Rows(), []int{2}; !equalInts(got, want) {
		t.Errorf("after clicking the third row: %v, want %v", got, want)
	}
	if lead != 2 {
		t.Errorf("the row the keys move from is %d, want 2", lead)
	}
	if err := tt.ClickWith(ui.Shift, "Row 5"); err != nil {
		t.Fatal(err)
	}
	if got, want := choice.Rows(), []int{2, 3, 4, 5}; !equalInts(got, want) {
		t.Errorf("after Shift-clicking the sixth: %v, want %v", got, want)
	}
}

// ── Tree ───────────────────────────────────────────────────────────────────

// treeKids is a board's branches and who is under them, three deep: the
// depth is what a tree indents by.
func treeKids(item string) []string {
	switch item {
	case "North":
		return []string{"Ada", "Nate"}
	case "Ada":
		return []string{"Ada, Tuesday"}
	}
	return nil
}

func TestTreeOpensAndIndents(t *testing.T) {
	var open Open[string]
	var lead = -1
	// A press is a question the view answers in the frame it is asked in, so
	// the toggled item is noted there, as an app would note it.
	var toggled string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		res := Tree(c, TreeOptions[string]{
			Roots: []string{"North", "South"}, Children: treeKids, Open: &open,
			Height: 240, Selected: &lead,
			Label: func(item string) string { return "Branch " + item },
			Row:   func(item string, depth int) { ui.Text(c, item).SingleLine() },
		})
		if item, ok := res.Toggled(); ok {
			toggled = item
		}
	}, 600, 400)

	if !tt.HasText("North") {
		t.Fatalf("the tree drew %q", tt.Texts())
	}
	if tt.HasText("Ada") {
		t.Error("a tree with nothing open shows its children")
	}
	if n := count(tt.Texts(), "Expand"); n != 1 {
		t.Errorf("%d branches can be opened, want 1: %q", n, tt.Texts())
	}

	if err := tt.Click("Expand"); err != nil {
		t.Fatal(err)
	}
	if toggled != "North" {
		t.Errorf("the tree reported toggling %q, want North", toggled)
	}
	if !open.Has("North") {
		t.Error("the caller's set does not have the item the tree opened")
	}
	for _, want := range []string{"Ada", "Nate"} {
		if !tt.HasText(want) {
			t.Errorf("the children of an open item are not shown: %q", tt.Texts())
		}
	}
	if tt.HasText("Ada, Tuesday") {
		t.Error("the children of a closed item are shown")
	}
	// Each level is in from the one above it, and the leaves of one level are
	// in one column: South and Nate are both leaves, and Nate is one step in
	// because it is a child. A branch is further in again, behind the arrow
	// that opens it.
	south, _ := tt.Find("South")
	nate, _ := tt.Find("Nate")
	north, _ := tt.Find("North")
	ada, _ := tt.Find("Ada")
	if nate.X <= south.X {
		t.Errorf("a child is not in from its parent: %v then %v", south, nate)
	}
	if ada.X <= nate.X {
		t.Errorf("a branch is not in from the items beside it: %v then %v", nate, ada)
	}
	if north.X <= south.X {
		t.Errorf("the items at the top are not in one column: %v then %v", south, north)
	}
	if n := count(tt.Texts(), "Expand"); n != 1 {
		t.Errorf("%d branches are closed and can be opened, want 1: %q", n, tt.Texts())
	}
	if n := count(tt.Texts(), "Collapse"); n != 1 {
		t.Errorf("%d branches are open, want 1: %q", n, tt.Texts())
	}

	// Open the level below, and the marks go on and off with what they open.
	if err := tt.Click("Expand"); err != nil {
		t.Fatal(err)
	}
	if toggled != "Ada" {
		t.Errorf("the tree reported toggling %q, want Ada", toggled)
	}
	if !open.Has("Ada") {
		t.Error("the caller's set does not have the item the tree opened")
	}
	if !tt.HasText("Ada, Tuesday") {
		t.Errorf("the children of an open branch are not shown: %q", tt.Texts())
	}
	tuesday, _ := tt.Find("Ada, Tuesday")
	if tuesday.X <= nate.X {
		t.Errorf("a grandchild is not in from the level above it: %v then %v", nate, tuesday)
	}
	if err := tt.Click("Collapse"); err != nil {
		t.Fatal(err)
	}
	if open.Has("North") {
		t.Error("the caller's set still has the item the tree closed")
	}
	if tt.HasText("Ada") {
		t.Error("the tree kept showing children the caller closed")
	}
}

func count(texts []string, want string) int {
	n := 0
	for _, s := range texts {
		if s == want {
			n++
		}
	}
	return n
}

func TestTreePanicsForWhatItCannotBeGiven(t *testing.T) {
	var open Open[string]
	panics(t, "a tree with nothing to ask about what is below an item", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			Tree(c, TreeOptions[string]{
				Roots: []string{"North"}, Open: &open, Height: 200,
				Row: func(string, int) {},
			})
		}, 400, 300)
	})
	panics(t, "a tree with no set of open items", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			Tree(c, TreeOptions[string]{
				Roots: []string{"North"}, Children: treeKids, Height: 200,
				Row: func(string, int) {},
			})
		}, 400, 300)
	})
}

// ── TreeTable ──────────────────────────────────────────────────────────────

func TestTreeTableShowsBranchesInAColumnAndSorts(t *testing.T) {
	var open Open[string]
	var sort Sort
	var toggled string
	var was bool
	cols := []Column{
		{Title: "Branch", ID: "branch", Share: 2},
		{Title: "Callbacks", ID: "n", Width: 110, Align: ui.End, Sortable: true},
	}
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		res := TreeTable(c, TreeTableOptions[string]{
			Roots: []string{"North", "South"}, Children: treeKids, Open: &open,
			Columns: cols, Height: 220, Sort: &sort,
			CellLabel: func(item string, col int) string {
				return fmt.Sprintf("%s: %s", cols[col].Title, item)
			},
			Cell: func(item string, col int) {
				switch col {
				case 0:
					ui.Text(c, item).SingleLine()
				case 1:
					ui.Text(c, fmt.Sprintf("%d", len(item)*3))
				}
			},
		})
		if item, ok := res.Toggled(); ok {
			toggled, was = item, true
		}
	}, 700, 400)

	if !tt.HasText("Branch") || !tt.HasText("Callbacks") {
		t.Fatalf("the head did not lay out: %q", tt.Texts())
	}
	if !tt.HasText("North") {
		t.Fatalf("the table drew %q", tt.Texts())
	}
	head, ok := tt.Find("Branch")
	if !ok || head.W == 0 {
		t.Fatalf("the head of the first column measured %v", head)
	}
	cell, ok := tt.Find("Branch: North")
	if !ok || cell.W != head.W {
		t.Errorf("a cell of the first column is %v wide, want the column's %v", cell.W, head.W)
	}
	if err := tt.Click("Expand"); err != nil {
		t.Fatal(err)
	}
	if !was || toggled != "North" {
		t.Errorf("the table reported toggling %q (%v), want North", toggled, was)
	}
	if !tt.HasText("Ada") {
		t.Errorf("the children of an open branch are not shown: %q", tt.Texts())
	}
	// The tree goes in the first column: a table whose names are in the last
	// column reads backwards.
	if _, ok := tt.Find("Branch: Ada"); !ok {
		t.Errorf("a row is not named: %q", tt.Texts())
	}
	// The tree indents in the first column: a leaf is in from the items at
	// the top, and a branch is further in again behind its own arrow.
	south, _ := tt.Find("South")
	nate, _ := tt.Find("Nate")
	ada, _ := tt.Find("Ada")
	if nate.X <= south.X {
		t.Errorf("an item of a branch is not in from the items at the top: %v then %v", south, nate)
	}
	if ada.X <= nate.X {
		t.Errorf("a branch is not in from the items beside it: %v then %v", nate, ada)
	}
	if err := tt.Click("Callbacks"); err != nil {
		t.Fatal(err)
	}
	if sort != (Sort{Column: "n"}) {
		t.Errorf("after a click on the head of Callbacks: %+v", sort)
	}
}

// ── Timeline ───────────────────────────────────────────────────────────────

func TestTimelineDrawsItsEventsInOrder(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Timeline(c, TimelineOptions{Height: 260, Stamps: true},
			Event{When: "09:02", Title: "Logged by phone", Detail: "Maple Street Bakery"},
			Event{When: "11:30", Title: "Third callback", Tone: core.Warning,
				Detail: "Andre Thomas"},
			Event{When: "16:45", Title: "Resolved", Tone: core.Success},
		)
	}, 600, 400)

	for _, want := range []string{"09:02", "Logged by phone", "Maple Street Bakery", "11:30", "Third callback", "16:45", "Resolved"} {
		if !tt.HasText(want) {
			t.Errorf("the timeline is missing %q; %q", want, tt.Texts())
		}
	}
	first, _ := tt.Find("Logged by phone")
	stamp, _ := tt.Find("09:02")
	if stamp.X >= first.X {
		t.Errorf("the stamps are not in a column of their own: %v then %v", stamp, first)
	}
	second, _ := tt.Find("Third callback")
	if second.Y <= first.Y {
		t.Errorf("the events are not one under the other: %v then %v", first, second)
	}
}

// ── Accordion and Collapsible ───────────────────────────────────────────────

func TestAccordionOpensOneSectionAtATime(t *testing.T) {
	var general, advanced bool
	var toggled int
	var was bool
	tt := ui.NewTester(func(c *ui.Context) {
		// Less motion, so that a section folded away in a frame is folded
		// away by the time the frame after it is drawn.
		core.WithReducedMotion(c, true)
		core.Use(c, core.Settings{})
		res := Accordion(c, AccordionOptions{Single: true},
			Section{Title: "General", Open: &general, Meta: "3",
				Body: func() { ui.Text(c, "Every callback comes back in 48 hours.") }},
			Section{Title: "Advanced", Open: &advanced,
				Body: func() { ui.Text(c, "Nobody has opened this yet.") }},
		)
		if i, ok := res.Toggled(); ok {
			toggled, was = i, true
		}
		_ = res
	}, 600, 400)

	if tt.HasText("Every callback comes back") {
		t.Error("a closed section shows what it holds")
	}
	if err := tt.Click("General"); err != nil {
		t.Fatal(err)
	}
	if !general {
		t.Error("the caller's bool for the section is still false after a click")
	}
	if !was || toggled != 0 {
		t.Errorf("the accordion reported toggling %d (%v), want 0", toggled, was)
	}
	if !tt.HasText("Every callback comes back") {
		t.Errorf("an open section shows nothing: %q", tt.Texts())
	}

	if err := tt.Click("Advanced"); err != nil {
		t.Fatal(err)
	}
	if !advanced {
		t.Error("the second section did not open")
	}
	if general {
		t.Error("a section that was open stayed open with Single set")
	}
	was = false
	if err := tt.Click("General"); err != nil {
		t.Fatal(err)
	}
	if !was || toggled != 0 {
		t.Errorf("the frame the press was in reported section %d (%v), want 0", toggled, was)
	}
	if !tt.HasText("Every callback comes back") {
		t.Errorf("the section that opened shows nothing: %q", tt.Texts())
	}
}

func TestAccordionHoldsSeveralOpenWhenAsked(t *testing.T) {
	var a, b bool
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Accordion(c, AccordionOptions{},
			Section{Title: "General", Open: &a, Body: func() { ui.Text(c, "first") }},
			Section{Title: "Advanced", Open: &b, Body: func() { ui.Text(c, "second") }},
		)
	}, 600, 400)
	if err := tt.Click("General"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Advanced"); err != nil {
		t.Fatal(err)
	}
	if !a || !b {
		t.Errorf("both sections open at once: %v, %v", a, b)
	}
	if !tt.HasText("first") || !tt.HasText("second") {
		t.Errorf("the frame shows %q", tt.Texts())
	}
}

func TestCollapsibleFoldsItsContentAway(t *testing.T) {
	var open bool
	var toggled bool
	tt := ui.NewTester(func(c *ui.Context) {
		core.WithReducedMotion(c, true)
		core.Use(c, core.Settings{})
		res := Collapsible(c, CollapsibleOptions{Title: "Why is this here?", Open: &open, Rule: true},
			func() { ui.Text(c, "Because the board is long.") })
		if res.Toggled() {
			toggled = true
		}
	}, 600, 400)

	if tt.HasText("Because the board is long") {
		t.Error("a closed collapsible shows its content")
	}
	if err := tt.Click("Why is this here?"); err != nil {
		t.Fatal(err)
	}
	if !open {
		t.Error("the caller's bool is still false after a click")
	}
	if !toggled {
		t.Error("the frame the click was in reported no press")
	}
	if !tt.HasText("Because the board is long") {
		t.Errorf("an open collapsible shows nothing: %q", tt.Texts())
	}
	if err := tt.Click("Why is this here?"); err != nil {
		t.Fatal(err)
	}
	if open {
		t.Error("a second click did not close it")
	}
}

// ── Carousel ───────────────────────────────────────────────────────────────

func slides(c *ui.Context, n int) []Slide {
	out := make([]Slide, n)
	for i := range out {
		out[i] = Slide{Label: fmt.Sprintf("Slide %d", i+1), Body: func() {
			ui.Text(c, fmt.Sprintf("Body %d", i+1)).SingleLine()
		}}
	}
	return out
}

// pic is a picture of one flat colour, so that a viewer has something to
// show and a test can tell one picture from another.
func pic(w, h int, tint uint8) *ui.Bitmap {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.SetRGBA(x, y, color.RGBA{R: tint, G: tint / 2, B: 255 - tint, A: 255})
		}
	}
	return ui.NewBitmap(img)
}

func TestCarouselPagesWithItsArrowsAndItsDots(t *testing.T) {
	at := 0
	view := func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Carousel(c, CarouselOptions{
			At: &at, Height: 140, Width: 320, Dots: true, Slides: slides(c, 3),
		})
	}
	tt := ui.NewTester(view, 500, 400)

	// A slide is in the window when it is the one the index says, a slide
	// width along: the ones either side of it are off to the sides, which is
	// what a carousel showing one slide at a time means.
	shown := func(n int) bool {
		t.Helper()
		r, ok := tt.Find(fmt.Sprintf("Body %d", n))
		return ok && r.W > 0 && r.X >= float32(320*at) && r.X < float32(320*(at+1))
	}
	if !shown(1) || shown(2) || shown(3) {
		first, _ := tt.Find("Body 1")
		second, _ := tt.Find("Body 2")
		t.Errorf("the carousel is showing the wrong slide: %v then %v", first, second)
	}
	if err := tt.Click("Next slide"); err != nil {
		t.Fatal(err)
	}
	if at != 1 {
		t.Errorf("after Next the index is %d, want 1", at)
	}
	if !shown(2) || shown(1) {
		r1, _ := tt.Find("Body 1")
		r2, _ := tt.Find("Body 2")
		r3, _ := tt.Find("Body 3")
		t.Errorf("after Next the carousel is on the wrong slide: %d: %v %v %v", at, r1, r2, r3)
	}
	if err := tt.Click("Previous slide"); err != nil {
		t.Fatal(err)
	}
	if at != 0 {
		t.Errorf("after Previous the index is %d, want 0", at)
	}
	if err := tt.Click("Slide 3, 3 of 3"); err != nil {
		t.Fatal(err)
	}
	if at != 2 {
		t.Errorf("after clicking the third dot the index is %d, want 2", at)
	}
	// At the end with no loop there is no next: the arrow is pressed rather
	// than taken away, because an arrow that came and went moves the window
	// under the pointer that was reaching for it.
	if err := tt.Click("Next slide"); err != nil {
		t.Fatal(err)
	}
	if at != 2 {
		t.Errorf("past the end: the index is %d, want 2", at)
	}
}

func TestCarouselLoopsWhenAsked(t *testing.T) {
	at := 2
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Carousel(c, CarouselOptions{
			At: &at, Height: 140, Width: 320, Slides: slides(c, 3), Loop: true,
		})
	}, 500, 400)
	if err := tt.Click("Next slide"); err != nil {
		t.Fatal(err)
	}
	if at != 0 {
		t.Errorf("from the last slide, Next went to %d, want 0", at)
	}
	if err := tt.Click("Previous slide"); err != nil {
		t.Fatal(err)
	}
	if at != 2 {
		t.Errorf("from the first slide, Previous went to %d, want 2", at)
	}
}

// ── DescriptionList ────────────────────────────────────────────────────────

func TestDescriptionListLaysItsPairsOutInTwoColumns(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		DescriptionList(c, DescriptionListOptions{Rules: true},
			Term{Term: "Opened", Value: "27 callbacks"},
			Term{Term: "Open cost", Value: "$7,340"},
			Term{Term: "Branch", Value: "Riverside Clinic"},
		)
	}, 600, 400)

	for _, want := range []string{"Opened", "27 callbacks", "Open cost", "$7,340", "Branch", "Riverside Clinic"} {
		if !tt.HasText(want) {
			t.Errorf("the list is missing %q; %q", want, tt.Texts())
		}
	}
	// Every value starts where the last one started, which is the whole of
	// a column of terms: without it the values are ragged and the terms are
	// the only thing that says what a value belongs to.
	first, _ := tt.Find("27 callbacks")
	second, _ := tt.Find("$7,340")
	third, _ := tt.Find("Riverside Clinic")
	if first.X != second.X || second.X != third.X {
		t.Errorf("the values do not start in one column: %v, %v, %v", first.X, second.X, third.X)
	}
	opened, _ := tt.Find("Opened")
	if opened.X >= first.X {
		t.Errorf("a term is not to the left of its value: %v then %v", opened, first)
	}
	below, _ := tt.Find("Open cost")
	if below.Y <= opened.Y {
		t.Errorf("the pairs are not one under the other: %v then %v", opened, below)
	}
}

// ── ImageViewer ────────────────────────────────────────────────────────────

func TestImageViewerZoomsAndPages(t *testing.T) {
	pictures := []ui.ImageSource{pic(90, 60, 210), pic(60, 90, 90), pic(80, 80, 30)}
	at, zoom := 0, float32(1)
	var zoomed, paged bool
	var step int
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		res := ImageViewer(c, ImageViewerOptions{
			Images: pictures, At: &at, Zoom: &zoom, Height: 200, Strip: true,
			Caption: "Three pictures",
			Label:   func(i int) string { return fmt.Sprintf("Pic %d", i+1) },
		})
		if res.Zoomed() {
			zoomed = true
		}
		if n, ok := res.Paged(); ok {
			step, paged = n, true
		}
	}, 700, 500)

	if !tt.HasText("Three pictures") {
		t.Fatalf("the viewer drew %q", tt.Texts())
	}
	if !tt.HasText("Pic 1, 1 of 3") {
		t.Errorf("the pictures are not named: %q", tt.Texts())
	}
	if err := tt.Click("Zoom in"); err != nil {
		t.Fatal(err)
	}
	if zoom != 1.5 {
		t.Errorf("after one click on Zoom in: %v, want 1.5", zoom)
	}
	if !zoomed {
		t.Error("the frame the zoom was in reported no press")
	}
	if !tt.HasText("150%") {
		t.Errorf("the viewer does not say how far it is zoomed: %q", tt.Texts())
	}
	if err := tt.Click("Next picture"); err != nil {
		t.Fatal(err)
	}
	if at != 1 {
		t.Errorf("after Next the picture is %d, want 1", at)
	}
	if !paged || step != 1 {
		t.Errorf("the frame reported paging %d (%v), want 1", step, paged)
	}
	if err := tt.Click("Pic 3, 3 of 3"); err != nil {
		t.Fatal(err)
	}
	if at != 2 {
		t.Errorf("after clicking the third thumbnail: %d", at)
	}
}

func TestImageViewerPanicsForWhatItCannotBeGiven(t *testing.T) {
	at := 0
	panics(t, "a viewer with no pictures", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			ImageViewer(c, ImageViewerOptions{At: &at, Height: 100})
		}, 400, 300)
	})
	panics(t, "a viewer showing a picture it does not have", func() {
		missing := 5
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			ImageViewer(c, ImageViewerOptions{
				Images: []ui.ImageSource{pic(10, 10, 1)}, At: &missing, Height: 100,
			})
		}, 400, 300)
	})
}

// ── pixels ─────────────────────────────────────────────────────────────────
// A colour is what a table does with a chosen row, and there is nothing else
// in the frame to read it from.

func finds(img *image.RGBA, r ui.Rect, want ui.Color) bool {
	for y := int(r.Y); y < int(r.Y+r.H); y++ {
		for x := int(r.X); x < int(r.X+r.W); x++ {
			if pixel(img, x, y) == want {
				return true
			}
		}
	}
	return false
}

func pixel(img *image.RGBA, x, y int) ui.Color {
	b := img.Bounds()
	if x < b.Min.X || y < b.Min.Y || x >= b.Max.X || y >= b.Max.Y {
		return ui.Color{}
	}
	c := img.RGBAAt(x, y)
	return ui.RGB(c.R, c.G, c.B)
}

// inset is the middle of a box, a little in from its left edge, where the
// text of a cell is least likely to be.
func inset(r ui.Rect, fromLeft, from, to float32) ui.Rect {
	return ui.Rect{
		X: r.X + fromLeft,
		Y: r.Y + r.H*from,
		W: r.W / 2,
		H: r.H * (to - from),
	}
}
