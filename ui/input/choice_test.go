package input

import (
	"image"
	"image/color"
	"math"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// Every control here is headless-rendered and asserted on the value it wrote
// into the caller's pointer, never on what it drew afterwards. The pointer is
// the whole contract: a control that painted the right thing and left the
// wrong value behind is broken, and a test that read the text back could not
// tell the two apart.

// wantsPanic runs view and fails unless it panics, which is how the library
// says a control was asked for something it cannot be given.
func wantsPanic(t *testing.T, what string, view func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Errorf("%s should have panicked", what)
		}
	}()
	view()
}

// needsName is the rule a control drawing only marks has to keep: no words
// anywhere means nothing for a screen reader to read out.
func TestCheckboxInsistsOnAName(t *testing.T) {
	wantsPanic(t, "an unnamed Checkbox", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			state := Unchecked
			Checkbox(c, &state, "", CheckboxOptions{})
		}, 200, 100)
	})
}

// TestCheckboxHasThreeStates walks a box through all three positions, because
// the third is not a variant of the second: a parent that reports itself as
// either on or off has lied about half its children.
func TestCheckboxHasThreeStates(t *testing.T) {
	state := Indeterminate
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Checkbox(c, &state, "Weekly summary", CheckboxOptions{})
	}, 320, 120)

	// A part-chosen box resolves to chosen, as every other control on the
	// desktop does with a half-set parent.
	if err := tt.Click("Weekly summary"); err != nil {
		t.Fatal(err)
	}
	if state != Checked {
		t.Fatalf("a part-chosen box resolved to %v, want Checked", state)
	}
	if err := tt.Click("Weekly summary"); err != nil {
		t.Fatal(err)
	}
	if state != Unchecked {
		t.Fatalf("a chosen box resolved to %v, want Unchecked", state)
	}
}

// TestCheckboxRejectsAnImpossibleState is the one thing a checkbox is strict
// about: the state is an enum, and a value outside it is a bug in the caller
// rather than a value to guess at.
func TestCheckboxRejectsAnImpossibleState(t *testing.T) {
	wantsPanic(t, "a state that is not one of the three", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			state := CheckState(7)
			Checkbox(c, &state, "Weekly summary", CheckboxOptions{})
		}, 200, 100)
	})
}

func TestCheckboxGroupTogglesItsSlice(t *testing.T) {
	sel := []string{"sms"}
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		CheckboxGroup(c, &sel, []Choice{
			{"sms", "Text message"},
			{"email", "Email"},
			{"push", "Push notification"},
		}, CheckboxGroupOptions{Label: "Channels"})
	}, 360, 240)

	if err := tt.Click("Email"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Text message"); err != nil {
		t.Fatal(err)
	}
	// Pressing a chosen box again takes it out, and what is left keeps the
	// order the choices were given in rather than the order they were pressed.
	want := []string{"email"}
	if strings.Join(sel, ",") != strings.Join(want, ",") {
		t.Fatalf("selected = %v, want %v", sel, want)
	}
}

func TestRadioGroupChoosesExactlyOne(t *testing.T) {
	sel := ""
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		RadioGroup(c, &sel, []Choice{
			{"low", "Low"},
			{"high", "High"},
		}, RadioGroupOptions{Label: "Priority"})
	}, 300, 200)
	if err := tt.Click("High"); err != nil {
		t.Fatal(err)
	}
	if sel != "high" {
		t.Fatalf("selected = %q, want %q", sel, "high")
	}
	if err := tt.Click("Low"); err != nil {
		t.Fatal(err)
	}
	if sel != "low" {
		t.Fatalf("selected = %q, want %q", sel, "low")
	}
}

// TestRadioGroupRefusesAValueItDoesNotHave is the mistake a radio group
// cannot show: a selection outside the set draws every radio off and looks
// like a question rather than like a bug.
func TestRadioGroupRefusesAValueItDoesNotHave(t *testing.T) {
	wantsPanic(t, "a selection that is not one of the choices", func() {
		sel := "urgent"
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			RadioGroup(c, &sel, []Choice{{"low", "Low"}, {"high", "High"}}, RadioGroupOptions{})
		}, 300, 200)
	})
}

func TestToggleGroupChoosesASegment(t *testing.T) {
	sel := 0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ToggleGroup(c, &sel, []string{"List", "Grid", "Board"}, ToggleGroupOptions{Label: "View"})
	}, 400, 120)
	if err := tt.Click("Grid"); err != nil {
		t.Fatal(err)
	}
	if sel != 1 {
		t.Fatalf("selected = %d, want 1", sel)
	}
}

func TestToggleGroupRefusesAnIndexPastItsLabels(t *testing.T) {
	wantsPanic(t, "a selection past the last label", func() {
		sel := 7
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			ToggleGroup(c, &sel, []string{"List", "Grid"}, ToggleGroupOptions{})
		}, 300, 100)
	})
}

func TestChoiceChipsAddsAndRemoves(t *testing.T) {
	sel := []string{"overdue"}
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ChoiceChips(c, &sel, []Choice{
			{"overdue", "Overdue"},
			{"open", "Open"},
			{"mine", "Mine"},
		}, ChoiceChipsOptions{Label: "Filters"})
	}, 420, 140)
	if err := tt.Click("Open"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Overdue"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Mine"); err != nil {
		t.Fatal(err)
	}
	if strings.Join(sel, ",") != "open,mine" {
		t.Fatalf("selected = %v, want [open mine]", sel)
	}
}

// TestChoiceChipsRefusesPastItsCap: a cap that quietly did nothing would be
// worse than no cap, so the chip that was refused leaves the slice alone.
func TestChoiceChipsRefusesPastItsCap(t *testing.T) {
	sel := []string{}
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ChoiceChips(c, &sel, []Choice{
			{"a", "Overdue"},
			{"b", "Open"},
			{"c", "Mine"},
		}, ChoiceChipsOptions{Label: "Filters", Max: 2})
	}, 420, 140)
	for _, chip := range []string{"Overdue", "Open", "Mine"} {
		if err := tt.Click(chip); err != nil {
			t.Fatal(err)
		}
	}
	if len(sel) != 2 {
		t.Fatalf("selected = %v, want two chips at the cap of two", sel)
	}
}

func TestSelectChoosesAnOption(t *testing.T) {
	sel := ""
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Select(c, &sel, []Choice{
			{"dana", "Dana Reyes"},
			{"sam", "Sam Okafor"},
		}, SelectOptions{Label: "Assign to", Placeholder: "Nobody yet"})
	}, 360, 300)
	if err := tt.Click("Assign to"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Sam Okafor"); err != nil {
		t.Fatal(err)
	}
	if sel != "sam" {
		t.Fatalf("selected = %q, want %q", sel, "sam")
	}
}

// TestSelectNarrowsTheFrameItDraws is about the panel rather than the value:
// the options are only built while the panel shows, so a control that kept
// them in a field would show them with the panel shut.
func TestSelectShowsItsOptionsOnlyWhileOpen(t *testing.T) {
	sel := ""
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Select(c, &sel, []Choice{{"dana", "Dana Reyes"}}, SelectOptions{Label: "Assign to", Placeholder: "Nobody yet"})
	}, 360, 300)
	if tt.HasText("Dana Reyes") {
		t.Error("a closed select must not show its options")
	}
	if err := tt.Click("Assign to"); err != nil {
		t.Fatal(err)
	}
	if !tt.HasText("Dana Reyes") {
		t.Errorf("the open select should show its options, got %q", tt.Texts())
	}
}

func TestSelectSearchNarrowsAndChooses(t *testing.T) {
	sel, query := "", ""
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		SelectSearch(c, &sel, &query, []Choice{
			{"h", "Hillside"},
			{"r", "Riverside"},
			{"m", "Meadowbrook"},
		}, SelectSearchOptions{Label: "Assign to", Placeholder: "Nobody yet", Search: "Search people"})
	}, 420, 400)
	if err := tt.Click("Assign to"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Search people"); err != nil {
		t.Fatal(err)
	}
	tt.Type("riv")
	if err := tt.Click("Riverside"); err != nil {
		t.Fatal(err)
	}
	// Both pointers, because a form that submits a search submits what was
	// searched for as well as what was found.
	if sel != "r" {
		t.Errorf("selected = %q, want %q", sel, "r")
	}
	if query != "riv" {
		t.Errorf("query = %q, want %q", query, "riv")
	}
}

// TestMultiSelectSelectsAllAndClearsThem is why the control carries those two
// buttons at all: a set of three chosen one press at a time is a set nobody
// empties, and a set of twenty empties never.
func TestMultiSelectSelectsAllAndClearsThem(t *testing.T) {
	sel := []string{"riverside"}
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		MultiSelect(c, &sel, []Choice{
			{"h", "Hillside"},
			{"r", "Riverside"},
			{"m", "Meadowbrook"},
		}, MultiSelectOptions{Label: "Branches", Placeholder: "Every branch"})
	}, 420, 400)
	if err := tt.Click("Branches"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Select all"); err != nil {
		t.Fatal(err)
	}
	// The choices' own order, so a set filled from the top reads back the
	// same way whichever way round it started.
	if strings.Join(sel, ",") != "h,r,m" {
		t.Fatalf("after select all: %v, want [h r m]", sel)
	}
	if err := tt.Click("Clear all"); err != nil {
		t.Fatal(err)
	}
	if len(sel) != 0 {
		t.Fatalf("after clear all: %v, want none", sel)
	}
}

// TestMultiSelectKeepsItsPanelOpen: the whole point of a multi-select is that
// twenty options are chosen with twenty presses, so the panel cannot close
// under the pointer the way a single choice's does.
func TestMultiSelectKeepsItsPanelOpen(t *testing.T) {
	sel := []string{}
	var last MultiSelectResult
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		last = MultiSelect(c, &sel, []Choice{
			{"h", "Hillside"},
			{"m", "Meadowbrook"},
		}, MultiSelectOptions{Label: "Branches", Placeholder: "Every branch"})
	}, 420, 400)
	if err := tt.Click("Branches"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Hillside"); err != nil {
		t.Fatal(err)
	}
	if len(sel) != 1 || sel[0] != "h" {
		t.Fatalf("selected = %v, want [h]", sel)
	}
	if last.Dismissed() {
		t.Error("choosing an option must not close the panel; the next choice is the next press")
	}
}

// TestMultiSelectReportsBeingDismissed reads the result inside the view, which
// is where a caller reads it: Dismissed is a one-frame pulse, the way Pressed
// and Revealed are everywhere else in this library, and a test that reads it
// after the tester has settled several frames reads a frame too late.
func TestMultiSelectReportsBeingDismissed(t *testing.T) {
	sel := []string{}
	dismissed := 0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := MultiSelect(c, &sel, []Choice{{"h", "Hillside"}},
			MultiSelectOptions{Label: "Branches", Placeholder: "Every branch"})
		if r.Dismissed() {
			dismissed++
		}
	}, 420, 400)
	if err := tt.Click("Branches"); err != nil {
		t.Fatal(err)
	}
	if dismissed != 0 {
		t.Fatalf("dismissed %d times before anything was dismissed", dismissed)
	}
	tt.Key(0, ui.KeyEscape)
	if dismissed != 1 {
		t.Errorf("Escape reported Dismissed %d times, want exactly 1", dismissed)
	}
	if tt.HasText("Hillside") {
		t.Error("Escape should have closed the panel")
	}
}

func TestCascaderChoosesAPath(t *testing.T) {
	path := []string{}
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Cascader(c, &path, branchTree(), CascaderOptions{Label: "Where", Placeholder: "Nowhere yet"})
	}, 460, 320)
	if err := tt.Click("Where"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Amsterdam"); err != nil {
		t.Fatal(err)
	}
	if strings.Join(path, ",") != "ams" {
		t.Fatalf("path = %v, want [ams]", path)
	}
	if err := tt.Click("West"); err != nil {
		t.Fatal(err)
	}
	if strings.Join(path, ",") != "ams,ams-west" {
		t.Fatalf("path = %v, want [ams ams-west]", path)
	}
}

// TestCascaderForgetsWhatIsBelowARepick: the district of a city that is no
// longer the one chosen is not still chosen.
func TestCascaderForgetsWhatIsBelowARepick(t *testing.T) {
	path := []string{"ams", "ams-west"}
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Cascader(c, &path, branchTree(), CascaderOptions{Label: "Where", Placeholder: "Nowhere yet"})
	}, 460, 320)
	if err := tt.Click("Where"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Rotterdam"); err != nil {
		t.Fatal(err)
	}
	if strings.Join(path, ",") != "rtm" {
		t.Fatalf("path = %v, want [rtm]", path)
	}
}

func TestCascaderRefusesAPathItDoesNotHave(t *testing.T) {
	wantsPanic(t, "a path through no branch of the tree", func() {
		path := []string{"ams", "ams-north"}
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			Cascader(c, &path, branchTree(), CascaderOptions{Label: "Where", Placeholder: "Nowhere yet"})
		}, 460, 320)
	})
}

func TestComboboxTakesASuggestion(t *testing.T) {
	sel := ""
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Combobox(c, &sel, []string{"Hillside", "Riverside", "Meadowbrook"},
			ComboboxOptions{Label: "Branch"})
	}, 460, 320)
	if err := tt.Click("Branch"); err != nil {
		t.Fatal(err)
	}
	tt.Type("Riv")
	if err := tt.Click("Riverside"); err != nil {
		t.Fatal(err)
	}
	if sel != "Riverside" {
		t.Fatalf("chosen = %q, want %q", sel, "Riverside")
	}
}

// TestTreeSelectChoosesABranchAndALeaf: a branch is a choice too, which is the
// reason this is a tree select and not a tree. The branch is opened with Right
// on the row that has the focus, the way MyGo's own tree is, because the row
// is the choice and the arrow is the disclosure.
func TestTreeSelectChoosesABranchAndALeaf(t *testing.T) {
	sel := ""
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		TreeSelect(c, &sel, branchTree(), TreeSelectOptions{Label: "Where", Placeholder: "Nowhere yet"})
	}, 460, 360)
	if err := tt.Click("Where"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Amsterdam"); err != nil {
		t.Fatal(err)
	}
	if sel != "ams" {
		t.Fatalf("chosen = %q, want %q", sel, "ams")
	}
	tt.Key(0, ui.KeyRight)
	if err := tt.Click("West"); err != nil {
		t.Fatal(err)
	}
	if sel != "ams-west" {
		t.Fatalf("chosen = %q, want %q", sel, "ams-west")
	}
}

// TestSliderClampsWhatItIsHanded: the control is the only thing that knows the
// bounds, so a value that arrived from a stored record or a query is put in
// range rather than drawn off the end of the rail.
func TestSliderClampsWhatItIsHanded(t *testing.T) {
	high := 400.0
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Slider(c, &high, SliderOptions{Min: 0, Max: 100, Label: "Utilisation", ShowValue: true})
	}, 420, 120)
	if high != 100 {
		t.Errorf("value above Max was left at %v, want it clamped to 100", high)
	}

	low := -40.0
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Slider(c, &low, SliderOptions{Min: 0, Max: 100, Label: "Utilisation"})
	}, 420, 120)
	if low != 0 {
		t.Errorf("value below Min was left at %v, want it clamped to 0", low)
	}
}

// TestSliderSnapsToItsDetents: a value that does not sit on a tick would be
// drawn between two ticks and jump to the nearest one on the first press.
func TestSliderSnapsToItsDetents(t *testing.T) {
	v := 30.0
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Slider(c, &v, SliderOptions{Min: 0, Max: 100, Step: 25, Label: "Utilisation"})
	}, 420, 120)
	if v != 25 {
		t.Errorf("value = %v, want it snapped to 25", v)
	}
}

func TestSliderRefusesARangeWithNoRoomInIt(t *testing.T) {
	wantsPanic(t, "a range whose Max is not above its Min", func() {
		v := 5.0
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			Slider(c, &v, SliderOptions{Min: 10, Max: 10, Label: "Utilisation"})
		}, 300, 120)
	})
}

// TestRangeSliderOrdersWhatItIsHanded: a filter asking for "between 90 and 10"
// is asking for nothing, and drawing it as it was asked would show two thumbs
// that had swapped places.
func TestRangeSliderOrdersWhatItIsHanded(t *testing.T) {
	low, high := 90.0, 10.0
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		RangeSlider(c, &low, &high, RangeSliderOptions{Min: 0, Max: 100, Label: "Age", ShowValue: true})
	}, 420, 160)
	if low != 10 || high != 90 {
		t.Fatalf("range = [%v %v], want [10 90]", low, high)
	}

	under, over := -20.0, 500.0
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		RangeSlider(c, &under, &over, RangeSliderOptions{Min: 0, Max: 100, Label: "Age"})
	}, 420, 160)
	if under != 0 || over != 100 {
		t.Errorf("range = [%v %v], want it clamped to [0 100]", under, over)
	}
}

func TestRangeSliderRefusesOneEndOnly(t *testing.T) {
	wantsPanic(t, "a range with no high end to point at", func() {
		low := 0.0
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			RangeSlider(c, &low, nil, RangeSliderOptions{Min: 0, Max: 100, Label: "Age"})
		}, 420, 160)
	})
}

// TestRatingClearsToNothing: without this the lowest a rating could go is one
// star, and a form has to invent a value for the person who has not rated yet.
func TestRatingClearsToNothing(t *testing.T) {
	v := 3
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Rating(c, &v, RatingOptions{Label: "Quality"})
	}, 360, 120)
	if err := tt.Click("4 of 5"); err != nil {
		t.Fatal(err)
	}
	if v != 4 {
		t.Fatalf("value = %d, want 4", v)
	}
	if err := tt.Click("4 of 5"); err != nil {
		t.Fatal(err)
	}
	if v != 0 {
		t.Fatalf("value = %d, want it cleared to 0", v)
	}
}

func TestRatingKeepsItsValueWhenClearingIsOff(t *testing.T) {
	v := 3
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Rating(c, &v, RatingOptions{Label: "Quality", NoClear: true})
	}, 360, 120)
	if err := tt.Click("3 of 5"); err != nil {
		t.Fatal(err)
	}
	if v != 3 {
		t.Fatalf("value = %d, want it held at 3", v)
	}
}

// TestRatingClampsWhatItIsHanded: a rating out of range is a number from
// somewhere else — an average, a stored value — and the control is the only
// thing that knows how many stars there are.
func TestRatingClampsWhatItIsHanded(t *testing.T) {
	v := 9
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Rating(c, &v, RatingOptions{Label: "Quality"})
	}, 360, 120)
	if v != 5 {
		t.Errorf("value = %d, want it clamped to 5", v)
	}

	under := -2
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Rating(c, &under, RatingOptions{Label: "Quality"})
	}, 360, 120)
	if under != 0 {
		t.Errorf("value = %d, want it clamped to 0", under)
	}
}

func TestRatingInsistsOnAName(t *testing.T) {
	wantsPanic(t, "a rating of unnamed stars", func() {
		v := 0
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			Rating(c, &v, RatingOptions{})
		}, 300, 120)
	})
}

func TestColorPaletteChoosesASwatch(t *testing.T) {
	sel := ui.Hex("#0b0b0f")
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ColorPalette(c, &sel, branchColours(), ColorPaletteOptions{Label: "Colour"})
	}, 420, 200)
	if err := tt.Click("Amber"); err != nil {
		t.Fatal(err)
	}
	if sel != ui.Hex("#f59e0b") {
		t.Fatalf("chosen = %v, want the amber swatch", sel)
	}
}

func TestColorPaletteInsistsOnANameForEverySwatch(t *testing.T) {
	wantsPanic(t, "a swatch with no name", func() {
		sel := ui.Hex("#0b0b0f")
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			ColorPalette(c, &sel, []Swatch{{ui.Hex("#0b0b0f"), ""}}, ColorPaletteOptions{Label: "Colour"})
		}, 300, 160)
	})
}

func TestColorPickerWritesTheColourUnderThePointer(t *testing.T) {
	col := ui.Hex("#ff0000")
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ColorPicker(c, &col, ColorPickerOptions{Label: "Accent"})
	}, 520, 460)
	if err := tt.Click("Accent"); err != nil {
		t.Fatal(err)
	}
	square, ok := tt.Find("Saturation and brightness")
	if !ok {
		t.Fatalf("the panel should hold a saturation and brightness square, got %q", tt.Texts())
	}
	tt.Press(square.X+square.W*0.8, square.Y+square.H*0.3)
	tt.Release(square.X+square.W*0.8, square.Y+square.H*0.3)
	if col == (ui.Color{R: 255, G: 0, B: 0, A: 255}) {
		t.Error("dragging across the square should write a new color")
	}
}

func TestColorPickerInsistsOnAName(t *testing.T) {
	wantsPanic(t, "an unnamed colour well", func() {
		col := ui.Hex("#ff0000")
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			ColorPicker(c, &col, ColorPickerOptions{})
		}, 300, 160)
	})
}

func TestTagsInputAddsAndRemoves(t *testing.T) {
	tags := []string{}
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		TagsInput(c, &tags, TagsInputOptions{
			Label:       "Tags",
			Placeholder: "Add a tag",
			Suggestions: []string{"escalated", "awaiting-customer"},
		})
	}, 460, 220)
	if err := tt.Click("Tags"); err != nil {
		t.Fatal(err)
	}
	tt.Type("escalated")
	tt.TypeKey(0, ui.KeyEnter, "\n")
	if strings.Join(tags, ",") != "escalated" {
		t.Fatalf("tags = %v, want [escalated]", tags)
	}
	// Enter again with an empty field adds nothing rather than an empty tag.
	tt.TypeKey(0, ui.KeyEnter, "\n")
	if len(tags) != 1 {
		t.Fatalf("tags = %v, want the empty field to add nothing", tags)
	}
	if err := tt.Click("Remove escalated"); err != nil {
		t.Fatal(err)
	}
	if len(tags) != 0 {
		t.Fatalf("tags = %v, want none after removing the chip", tags)
	}
}

func TestTagsInputHoldsToItsCap(t *testing.T) {
	tags := []string{"one", "two", "three"}
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		TagsInput(c, &tags, TagsInputOptions{
			Label:       "Tags",
			Placeholder: "Add a tag",
			Max:         2,
		})
	}, 460, 220)
	if strings.Join(tags, ",") != "one,two" {
		t.Fatalf("tags = %v, want the cap to hold them to [one two]", tags)
	}
}

func TestTagsInputInsistsOnAPlaceholder(t *testing.T) {
	wantsPanic(t, "a tag field that does not say what a tag is", func() {
		tags := []string{}
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			TagsInput(c, &tags, TagsInputOptions{Label: "Tags"})
		}, 460, 220)
	})
}

// TestAChosenSwatchStaysVisibleInDarkMode is the contrast rule this library
// cannot take from a token: the swatch is a color the caller chose, so the
// mark on it has to ask the swatch which ink it takes. In a dark window a
// dark swatch on a dark surface is the case that gets lost, and a mark in the
// window's text colour would sit at the same value as both of them.
//
// The test reads the rendered frame rather than the tokens, because the thing
// worth guarding is what is on the screen, not what was decided about it.
func TestAChosenSwatchStaysVisibleInDarkMode(t *testing.T) {
	assertChosenSwatchIsVisible(t, core.Dark, "#0b0b0f", "Ink")
}

// TestAChosenPaleSwatchStaysVisibleInLightMode is the other end of the same
// rule, and the reason the mark asks its swatch rather than always being
// white: on a near-white swatch a white tick is invisible.
func TestAChosenPaleSwatchStaysVisibleInLightMode(t *testing.T) {
	assertChosenSwatchIsVisible(t, core.Light, "#fbfbfd", "Snow")
}

func assertChosenSwatchIsVisible(t *testing.T, mode core.Mode, hex, name string) {
	t.Helper()
	swatch := ui.Hex(hex)
	sel := swatch
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: mode})
		ColorPalette(c, &sel, []Swatch{
			{swatch, name},
			{ui.Hex("#f59e0b"), "Amber"},
		}, ColorPaletteOptions{Label: "Colour"})
	}, 420, 200)
	r, ok := tt.Find(name)
	if !ok {
		t.Fatalf("no swatch named %q", name)
	}
	frame := tt.Image()

	k := theme.Light()
	if mode == core.Dark {
		k = theme.Dark()
	}
	if got := strongestContrastIn(frame, r, swatch); got < 3 {
		t.Errorf("%s: the mark on the %s swatch reaches only %.2f:1 against the swatch itself, which is under the 3:1 a mark has to clear",
			mode, hex, got)
	}
	// And the ring, which is the ink that reads on the surface rather than on
	// the swatch: without it a swatch of very nearly the surface's own colour
	// has no edge at all.
	if got := strongestContrastIn(frame, r, k.Background); got < 3 {
		t.Errorf("%s: the ring around the %s swatch reaches only %.2f:1 against the surface, which is under the 3:1 an edge has to clear",
			mode, hex, got)
	}
}

// TestAnUnchosenSwatchIsNotMarked is the other half of the rule: the contrast
// a chosen swatch is held to has to come from its mark and not from the
// window, or the assertion above would pass on any palette at all.
func TestAnUnchosenSwatchIsNotMarked(t *testing.T) {
	chosen, plain := ui.Hex("#0b0b0f"), ui.Hex("#15151a")
	sel := chosen
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: core.Dark})
		ColorPalette(c, &sel, []Swatch{{chosen, "Ink"}, {plain, "Plain"}}, ColorPaletteOptions{Label: "Colour"})
	}, 420, 200)
	r, ok := tt.Find("Plain")
	if !ok {
		t.Fatal("no swatch named Plain")
	}
	if got := strongestContrastIn(tt.Image(), r, plain); got > 3 {
		t.Errorf("an unchosen swatch reaches %.2f:1 against itself, so something other than its mark is doing it", got)
	}
}

// strongestContrastIn is the highest contrast ratio any pixel of r has with
// bg, which is how much the thing in r stands out from it.
func strongestContrastIn(frame *image.RGBA, r ui.Rect, bg ui.Color) float32 {
	best := float32(0)
	for y := int(r.Y); y < int(r.Y+r.H); y++ {
		for x := int(r.X); x < int(r.X+r.W); x++ {
			if y < 0 || y >= frame.Bounds().Dy() || x < 0 || x >= frame.Bounds().Dx() {
				continue
			}
			p := frame.RGBAAt(x, y)
			if cr := contrast(ui.RGBA(p.R, p.G, p.B, float32(p.A)/255), bg); cr > best {
				best = cr
			}
		}
	}
	return best
}

// TestEveryChoiceControlRendersInBothAppearances is the smoke test for the
// file: a control that panics, or draws nothing, in one of the two palettes is
// caught here rather than on somebody's desktop.
func TestEveryChoiceControlRendersInBothAppearances(t *testing.T) {
	controls := []struct {
		name string
		view func(c *ui.Context)
	}{
		{"Checkbox", func(c *ui.Context) {
			s := Checked
			Checkbox(c, &s, "Weekly summary", CheckboxOptions{})
		}},
		{"CheckboxGroup", func(c *ui.Context) {
			var sel []string
			CheckboxGroup(c, &sel, []Choice{{"a", "Email"}}, CheckboxGroupOptions{Label: "Channels"})
		}},
		{"RadioGroup", func(c *ui.Context) {
			sel := "a"
			RadioGroup(c, &sel, []Choice{{"a", "Low"}, {"b", "High"}}, RadioGroupOptions{Label: "Priority"})
		}},
		{"ToggleGroup", func(c *ui.Context) {
			sel := 0
			ToggleGroup(c, &sel, []string{"List", "Grid"}, ToggleGroupOptions{Label: "View"})
		}},
		{"ChoiceChips", func(c *ui.Context) {
			var sel []string
			ChoiceChips(c, &sel, []Choice{{"a", "Overdue"}}, ChoiceChipsOptions{Label: "Filters"})
		}},
		{"Select", func(c *ui.Context) {
			sel := "a"
			Select(c, &sel, []Choice{{"a", "Dana Reyes"}}, SelectOptions{Label: "Assign to", Placeholder: "Nobody yet"})
		}},
		{"SelectSearch", func(c *ui.Context) {
			sel, query := "", ""
			SelectSearch(c, &sel, &query, []Choice{{"a", "Dana Reyes"}},
				SelectSearchOptions{Label: "Assign to", Placeholder: "Nobody yet", Search: "Search people"})
		}},
		{"MultiSelect", func(c *ui.Context) {
			var sel []string
			MultiSelect(c, &sel, []Choice{{"a", "Hillside"}},
				MultiSelectOptions{Label: "Branches", Placeholder: "Every branch"})
		}},
		{"Cascader", func(c *ui.Context) {
			var path []string
			Cascader(c, &path, branchTree(), CascaderOptions{Label: "Where", Placeholder: "Nowhere yet"})
		}},
		{"Combobox", func(c *ui.Context) {
			sel := ""
			Combobox(c, &sel, []string{"Hillside"}, ComboboxOptions{Label: "Branch"})
		}},
		{"TreeSelect", func(c *ui.Context) {
			sel := ""
			TreeSelect(c, &sel, branchTree(), TreeSelectOptions{Label: "Where", Placeholder: "Nowhere yet"})
		}},
		{"Switch", func(c *ui.Context) {
			on := true
			Switch(c, &on, SwitchOptions{Label: "Group by branch"})
		}},
		{"Slider", func(c *ui.Context) {
			v := 40.0
			Slider(c, &v, SliderOptions{Min: 0, Max: 100, Step: 25, Label: "Utilisation", ShowValue: true})
		}},
		{"RangeSlider", func(c *ui.Context) {
			low, high := 20.0, 80.0
			RangeSlider(c, &low, &high, RangeSliderOptions{Min: 0, Max: 100, Label: "Age", ShowValue: true})
		}},
		{"Rating", func(c *ui.Context) {
			v := 3
			Rating(c, &v, RatingOptions{Label: "Quality"})
		}},
		{"ColorPicker", func(c *ui.Context) {
			col := ui.Hex("#2563eb")
			ColorPicker(c, &col, ColorPickerOptions{Label: "Accent"})
		}},
		{"ColorPalette", func(c *ui.Context) {
			sel := ui.Hex("#2563eb")
			ColorPalette(c, &sel, branchColours(), ColorPaletteOptions{Label: "Colour"})
		}},
		{"TagsInput", func(c *ui.Context) {
			tags := []string{"escalated"}
			TagsInput(c, &tags, TagsInputOptions{Label: "Tags", Placeholder: "Add a tag"})
		}},
	}
	for _, mode := range []core.Mode{core.Light, core.Dark} {
		for _, ctl := range controls {
			tt := ui.NewTester(func(c *ui.Context) {
				core.Use(c, core.Settings{Mode: mode})
				ctl.view(c)
			}, 460, 520)
			if len(tt.Texts()) == 0 {
				t.Errorf("%s in %s drew nothing at all", ctl.name, mode)
			}
		}
	}
}

// TestClampingLandsOnAStep is the property that matters about clampStep: what
// it hands back is on the step and inside the range, so the first press moves
// the value by one step rather than snapping it somewhere else first.
//
// The cases where the step does not divide the range are the ones worth
// having. A 0..10 rail stepping by 4 offers 0, 4 and 8; a value of 11 clamps
// to 8 and not to 10, because 10 is not a step at all and a thumb resting on
// 10 would move to 8 the moment it was pressed.
func TestClampingLandsOnAStep(t *testing.T) {
	for _, c := range []struct{ from, min, max, step float64 }{
		{-3, 0, 100, 25}, {7, 0, 100, 25}, {37, 0, 100, 25},
		{103, 0, 100, 25}, {11, 0, 10, 4}, {13, 0, 10, 4},
		{2.5, 0, 1, 0.4}, {1.1, 0, 1, 0.3}, {0.9, 0, 0.5, 0.2},
	} {
		got := clampStep(c.from, c.min, c.max, c.step)
		if got < c.min || got > c.max {
			t.Errorf("clampStep(%v, %v, %v, %v) = %v, which is out of range", c.from, c.min, c.max, c.step, got)
			continue
		}
		if off := (got - c.min) - math.Round((got-c.min)/c.step)*c.step; math.Abs(off) > 1e-9 {
			t.Errorf("clampStep(%v, %v, %v, %v) = %v, which is %v off the step", c.from, c.min, c.max, c.step, got, off)
		}
		if again := clampStep(got, c.min, c.max, c.step); again != got {
			t.Errorf("clampStep is not idempotent: %v became %v and then %v", c.from, got, again)
		}
	}
}

// TestSliderRefusesAStepWiderThanItsRange: a rail stepped by more than the
// whole range has exactly one value it can take, so it draws as a slider and
// never moves. That is a segmented control of one segment, and it is cheaper
// to say so here than to let someone find it by dragging.
func TestSliderRefusesAStepWiderThanItsRange(t *testing.T) {
	for _, who := range []string{"Slider", "RangeSlider"} {
		t.Run(who, func(t *testing.T) {
			wantsPanic(t, "a step wider than the range", func() {
				v, lo, hi := 1.0, 0.0, 3.0
				ui.NewTester(func(c *ui.Context) {
					core.Use(c, core.Settings{})
					if who == "Slider" {
						Slider(c, &v, SliderOptions{Min: lo, Max: hi, Step: 10, Label: "Count"})
						return
					}
					RangeSlider(c, &lo, &hi, RangeSliderOptions{Min: 0, Max: 3, Step: 10, Label: "Count"})
				}, 300, 120)
			})
		})
	}
}

// TestASteppedSliderDrawsOneRail is the test for a bug no value assertion
// could ever have caught. ui.Slider attaches what it builds to the parent it
// is given, so a rail chosen by calling ui.Slider and then ui.StepSlider
// leaves the first one in the tree behind the second: two rails are drawn,
// while the label, the value, the clicks and every Find all still come from
// the second and say nothing is wrong.
//
// Counting the painted rows of the WHOLE frame is the only way to see it: the
// stray rail has no label of its own, so measuring the labelled one alone
// looks at only the half that was never wrong.
func TestASteppedSliderDrawsOneRail(t *testing.T) {
	rails := func(step float64) int {
		v := 40.0
		tt := ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			Slider(c, &v, SliderOptions{Min: 0, Max: 100, Step: step, Label: "Utilisation"})
		}, 420, 120)
		if _, ok := tt.Find("Utilisation"); !ok {
			t.Fatalf("no rail at step %v: %q", step, tt.Texts())
		}
		return accentRuns(tt.Image(), theme.Light().Accent)
	}
	if plain, stepped := rails(0), rails(25); stepped != plain {
		t.Errorf("a stepped slider draws %d bands of accent and a plain one %d — the detents changed the rail, they did not double it",
			stepped, plain)
	}
}

// accentRuns is how many separate horizontal bands of near enough the accent
// colour are painted anywhere in the frame.
func accentRuns(frame *image.RGBA, accent ui.Color) int {
	near := func(p colorRGBA) bool {
		d := func(a, b uint8) int {
			if v := int(a) - int(b); v < 0 {
				return -v
			} else {
				return v
			}
		}
		return d(p.R, accent.R) < 12 && d(p.G, accent.G) < 12 && d(p.B, accent.B) < 12
	}
	lit := make([]bool, 0, frame.Bounds().Dy())
	for y := 0; y < frame.Bounds().Dy(); y++ {
		n := 0
		for x := 0; x < frame.Bounds().Dx(); x++ {
			if near(frame.RGBAAt(x, y)) {
				n++
			}
		}
		lit = append(lit, n > 20)
	}
	runs, prev := 0, -2
	for i, on := range lit {
		if on && i != prev+1 {
			runs++
		}
		if on {
			prev = i
		}
	}
	return runs
}

// TestAnOptionRowAnswersThePointer: a panel whose rows do not light under the
// pointer reads as a picture rather than as something to point at, and
// pointing is how most of a dropdown's rows get chosen.
func TestAnOptionRowAnswersThePointer(t *testing.T) {
	sel := []string{}
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		MultiSelect(c, &sel, []Choice{{"h", "Hillside"}, {"m", "Meadowbrook"}},
			MultiSelectOptions{Label: "Branches", Placeholder: "Every branch"})
	}, 420, 400)
	if err := tt.Click("Branches"); err != nil {
		t.Fatal(err)
	}
	idle, ok := tt.Find("Hillside")
	if !ok {
		t.Fatalf("no row: %q", tt.Texts())
	}
	rest := pixelAt(tt.Image(), idle.X+idle.W/2, idle.Y+idle.H/2)
	tt.Move(idle.X+idle.W/2, idle.Y+idle.H/2)
	tt.Frame()
	row, _ := tt.Find("Hillside")
	over := pixelAt(tt.Image(), row.X+row.W/2, row.Y+row.H/2)
	if over == rest {
		t.Error("an option row looks the same with the pointer on it as with it away")
	}
	if contrast(over, rest) < 1.1 {
		t.Errorf("the hovered row is %v against %v, which is too close to tell", over, rest)
	}
}

func pixelAt(frame *image.RGBA, x, y float32) ui.Color {
	p := frame.RGBAAt(int(x), int(y))
	return ui.RGBA(p.R, p.G, p.B, float32(p.A)/255)
}

// colorRGBA is the pixel type of a frame, named here so the two image
// packages do not have to be imported under their own names.
type colorRGBA = color.RGBA

// TestTheValueSitsBesideTheRail is the layout half of ShowValue, and it
// exists because asserting the value is not enough: a rail and its number in
// two separate rows are still both drawn, and the number is still correct.
// The doc's own advice applies — assert where things are rather than what
// they say, because a position says what the layout did.
func TestTheValueSitsBesideTheRail(t *testing.T) {
	v := 40.0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Slider(c, &v, SliderOptions{Min: 0, Max: 100, Label: "Utilisation", ShowValue: true})
	}, 420, 120)
	rail, ok := tt.Find("Utilisation")
	if !ok {
		t.Fatalf("no rail: %q", tt.Texts())
	}
	number, ok := tt.Find("40")
	if !ok {
		t.Fatalf("no value beside the rail: %q", tt.Texts())
	}
	if number.X <= rail.X+rail.W {
		t.Errorf("the value is at x=%v, at or left of the rail's right edge %v — it is not beside it",
			number.X, rail.X+rail.W)
	}
	if number.Y > rail.Y+rail.H {
		t.Errorf("the value is at y=%v, below the rail's bottom %v — it is under it, not beside it",
			number.Y, rail.Y+rail.H)
	}
}

// TestTheRangeNumbersSitUnderTheRail is the same assertion for the other
// layout, and the other half of why the rail is built where it sits: the two
// numbers belong on their own line below it, across its width.
func TestTheRangeNumbersSitUnderTheRail(t *testing.T) {
	low, high := 20.0, 80.0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		RangeSlider(c, &low, &high, RangeSliderOptions{Min: 0, Max: 100, Label: "Age", ShowValue: true})
	}, 420, 160)
	rail, ok := tt.Find("Age")
	if !ok {
		t.Fatalf("no rail: %q", tt.Texts())
	}
	numbers, ok := tt.Find("20 – 80")
	if !ok {
		t.Fatalf("no numbers under the rail: %q", tt.Texts())
	}
	if numbers.Y < rail.Y+rail.H {
		t.Errorf("the numbers are at y=%v, on the rail's own line (%v..%v) — they belong under it",
			numbers.Y, rail.Y, rail.Y+rail.H)
	}
	if numbers.W < rail.W/2 {
		t.Errorf("the numbers span %v of the rail's %v — they should sit across it", numbers.W, rail.W)
	}
}

// TestTheTriggerFollowsTheDensity is the design system's spacing contract in
// one assertion: a control is sized from core.ControlHeight rather than from
// its own number, so a window that asks for Comfortable really does get
// taller controls rather than only roomier gaps.
func TestTheTriggerFollowsTheDensity(t *testing.T) {
	sel := ""
	height := func(d theme.Density) float32 {
		tt := ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{Density: d})
			Select(c, &sel, []Choice{{"dana", "Dana Reyes"}},
				SelectOptions{Label: "Assign to", Placeholder: "Nobody yet"})
		}, 360, 300)
		r, ok := tt.Find("Assign to")
		if !ok {
			t.Fatalf("no trigger at %s", d)
		}
		return r.H
	}
	compact, comfortable := height(theme.Compact), height(theme.Comfortable)
	if comfortable <= compact {
		t.Errorf("trigger heights: Compact %v, Comfortable %v — a control is sized from "+
			"core.ControlHeight, so Comfortable has to be the taller one", compact, comfortable)
	}
}

// branchTree is the two-level tree the cascader and the tree select choose
// from: a city and its districts, which is the shape both are for.
func branchTree() []Node {
	return []Node{
		{Value: "ams", Label: "Amsterdam", Children: []Node{
			{Value: "ams-west", Label: "West"},
			{Value: "ams-east", Label: "East"},
		}},
		{Value: "rtm", Label: "Rotterdam"},
	}
}

// branchColours is a palette with a near-black and a near-white in it, which
// are the two that a mark written in one fixed ink cannot serve.
func branchColours() []Swatch {
	return []Swatch{
		{ui.Hex("#0b0b0f"), "Ink"},
		{ui.Hex("#fbfbfd"), "Snow"},
		{ui.Hex("#2563eb"), "Blue"},
		{ui.Hex("#f59e0b"), "Amber"},
	}
}
