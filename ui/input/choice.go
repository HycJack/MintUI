package input

import (
	"strconv"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// CheckState is what a check box is showing. It is a type of its own rather
// than a bool because the third state is not a flavour of the second: "part
// of a group is chosen" is what a parent row says, and a parent that reports
// itself as either on or off has lied about half its children.
type CheckState int

const (
	// Unchecked is the box as it comes.
	Unchecked CheckState = iota
	// Checked is the box with a tick in it.
	Checked
	// Indeterminate is the box with a dash in it: some of what it stands for
	// is chosen and some is not.
	Indeterminate
)

func (s CheckState) String() string {
	switch s {
	case Checked:
		return "Checked"
	case Indeterminate:
		return "Indeterminate"
	}
	return "Unchecked"
}

// mark is the box this state is drawn as. Indeterminate shares Checked's
// filled face and differs only in what is drawn inside it, so the two read as
// one control in three positions rather than as two different boxes.
func (s CheckState) mark() mark {
	switch s {
	case Checked:
		return tickOn
	case Indeterminate:
		return dashMark
	}
	return tickMark
}

// CheckboxOptions configure a Checkbox.
type CheckboxOptions struct {
	// Label names the control for assistive technology. A box with a word
	// beside it needs none; a box standing alone, showing only its tick,
	// always does.
	Label string
	// Disabled greys it out: it takes neither clicks nor focus.
	Disabled bool
}

// Checkbox is a box that is chosen or not, with a third state for a parent
// of a partly chosen group.
//
// The state is the caller's pointer, and a press moves it there and then, so
// there is no handler to keep in step with it. Pressing an Indeterminate box
// resolves it to Checked, which is what every other control on the desktop
// does with a half-set parent: the next press is an assertion that the whole
// of it is wanted, and a second press takes it back.
//
//	core.Use(c, core.Settings{})
//	var agree bool
//	Checkbox(c, stateOf(&agree), "Send me the weekly summary", CheckboxOptions{})
func Checkbox(c *ui.Context, state *CheckState, label string, opts CheckboxOptions) *ui.Element {
	if state == nil {
		panic("input: Checkbox needs a state to point at")
	}
	if label == "" && opts.Label == "" {
		panic("input: Checkbox needs a label or options.Label; a tick with no words beside it has nothing to read out")
	}
	if *state < Unchecked || *state > Indeterminate {
		panic("input: Checkbox got a state that is not Unchecked, Checked or Indeterminate")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	name := label
	if name == "" {
		name = opts.Label
	}
	// MyGo's base is two states, and it is the base that knows how a check
	// box behaves — the click, Space, the focus, the assistive technology's
	// checked or not. The third state is translated onto it rather than
	// reimplemented, so the keyboard is the desktop's and not ours.
	on := *state == Checked
	row := ui.CheckboxBase(c, &on).Gap(u*1.5).Radius(theme.SmallRadius).
		Padding(u*0.5, u).FillWidth().Label(name).Tooltip(name).Disabled(opts.Disabled)
	if row.Changed() {
		if on {
			*state = Checked
		} else {
			*state = Unchecked
		}
	}
	row.Children(func() {
		markFace(c, state.mark(), u*4)
		if label != "" {
			fg := k.Text
			if opts.Disabled {
				fg = k.TextFaint
			}
			ui.Text(c, label).SingleLine().TextColor(fg).FontSize(core.FontSize(c, theme.BodySize))
		}
	})
	return row
}

// CheckboxGroupOptions configure a CheckboxGroup.
type CheckboxGroupOptions struct {
	// Label names the group for assistive technology. It is the words of the
	// question, which a list of boxes cannot state by itself: four boxes
	// under nothing at all are four unrelated controls.
	Label string
	// Disabled greys out every box in the group.
	Disabled bool
}

// CheckboxGroup is a column of boxes choosing any number of a set, out of
// the caller's []string. The values live in that slice and nowhere else, so
// what is chosen can be read without asking the control.
//
// A value is added in the order the choices are given and taken out again in
// place, so the slice stays the same shape across a session rather than
// reordering itself as boxes are pressed.
func CheckboxGroup(c *ui.Context, selected *[]string, choices []Choice, opts CheckboxGroupOptions) *ui.Element {
	if selected == nil {
		panic("input: CheckboxGroup needs a selection to point at")
	}
	if len(choices) == 0 {
		panic("input: CheckboxGroup needs at least one choice")
	}
	for i, ch := range choices {
		if ch.Label == "" {
			panic("input: CheckboxGroup choice " + strconv.Itoa(i) + " has no Label; the box would be a box with nothing in front of it")
		}
	}
	u := core.Density(c).Unit()

	group := ui.Column(c).FillWidth().Gap(u * 0.75)
	if opts.Label != "" {
		group.Label(opts.Label)
	}
	group.Children(func() {
		for _, ch := range choices {
			st := Unchecked
			if hasValue(*selected, ch.Value) {
				st = Checked
			}
			row := Checkbox(c, &st, ch.Label, CheckboxOptions{
				Label:    ch.Value,
				Disabled: opts.Disabled,
			})
			if !row.Changed() {
				continue
			}
			if st == Checked {
				// A value already in the slice is not added twice, so a
				// caller that seeded the slice by hand does not end up with
				// the same option twice over.
				if !hasValue(*selected, ch.Value) {
					*selected = append(*selected, ch.Value)
				}
				continue
			}
			*selected = without(*selected, ch.Value)
		}
	})
	return group
}

// RadioGroupOptions configure a RadioGroup.
type RadioGroupOptions struct {
	// Label names the group for assistive technology, and is what it should
	// say: "Priority", not "Critical, High, Normal". A radio group read out
	// as a bare list of three radios does not say what choosing between them
	// is for.
	Label string
	// Direction puts the radios in a row, for a short set that fits on one
	// line; the default is a column, which is the one that still reads when
	// the labels are sentences.
	Direction Row
	// Disabled greys out every radio in the group.
	Disabled bool
}

// Row is which way a group of choices runs.
type Row int

const (
	// Stacked is a column, one choice under another.
	Stacked Row = iota
	// Inline is a row, which suits a set of short words.
	Inline
)

// RadioGroup is a set of radios of which exactly one is chosen, out of the
// caller's string. The string is the value, so a form submits it without
// asking the control to translate an index back into a word.
//
// It is built on ui.Radio and ui.RadioGroup, which is what makes the whole set
// one stop of Tab with the arrows moving within it; core.Use has already
// re-themed MyGo, so the dots come out in this library's accent.
func RadioGroup(c *ui.Context, selected *string, choices []Choice, opts RadioGroupOptions) *ui.Element {
	if selected == nil {
		panic("input: RadioGroup needs a selection to point at")
	}
	if len(choices) == 0 {
		panic("input: RadioGroup needs at least one choice")
	}
	for i, ch := range choices {
		if ch.Label == "" {
			panic("input: RadioGroup choice " + strconv.Itoa(i) + " has no Label; a radio with nothing beside it is a dot with no name")
		}
	}
	if *selected != "" && !hasValue(checkChoices(choices), *selected) {
		// A selection that is not one of the choices is the one mistake a
		// radio group cannot show: it would draw every radio off and look
		// like a question rather than a bug.
		panic("input: RadioGroup selected " + *selected + ", which is not one of its choices")
	}
	u := core.Density(c).Unit()

	group := ui.RadioGroup(c, func() {
		for _, ch := range choices {
			ui.Radio(c, selected, ch.Value, ch.Label).Disabled(opts.Disabled)
		}
	})
	if opts.Direction == Inline {
		group.Row().Gap(u * 2.5)
	} else {
		group.Gap(u * 1.5)
	}
	group.FillWidth()
	if opts.Label != "" {
		group.Label(opts.Label)
	}
	return group
}

// ToggleGroupOptions configure a ToggleGroup.
type ToggleGroupOptions struct {
	// Label names the group for assistive technology, as RadioGroup's does.
	Label string
	// Scrollable lets a set wider than its window scroll sideways rather
	// than squeeze its segments: a view switcher with nine views has to be
	// reachable, and squeezing "Assignments" to fit is not reaching it.
	Scrollable bool
	// Spread pushes the segments to the two ends of the track, for a group
	// that is meant to fill the row it sits in — a tab bar under a title,
	// rather than a control beside a label.
	Spread bool
	// Disabled greys out every segment.
	Disabled bool
}

// ToggleGroup is a row of buttons of which one is chosen, out of the
// caller's index. It is Segmented's stronger sibling: the same control with
// a set that may be wider than the window, and with the segments able to
// spread to the ends of their track.
//
// The index is the caller's, so it is the same *int a Segmented takes, and a
// view that already keeps one can swap to this without moving its state.
func ToggleGroup(c *ui.Context, selected *int, labels []string, opts ToggleGroupOptions) *ui.Element {
	if selected == nil {
		panic("input: ToggleGroup needs a selection to point at")
	}
	if len(labels) == 0 {
		panic("input: ToggleGroup needs at least one label")
	}
	if *selected < 0 || *selected >= len(labels) {
		panic("input: ToggleGroup selection is out of range")
	}
	for i, l := range labels {
		if l == "" {
			panic("input: ToggleGroup label " + strconv.Itoa(i) + " is empty; a segment with no words is a divider")
		}
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	// SegmentedBase is the control underneath: a radio group whose radios
	// are the segments, so the arrows move between them within one stop of
	// Tab. Only the track and the faces are ours.
	parts := ui.SegmentedBase(c, selected, len(labels))
	track := parts.Track.AlignItems(ui.Stretch).Padding(u * 0.5).Gap(u * 0.5).
		Radius(theme.ControlRadius).Background(k.Surface).
		BorderWidth(theme.BorderWidth).BorderColor(k.Border)
	if opts.Spread && !opts.Scrollable {
		// Spread needs a track to fill: a scrolling one is as wide as its
		// segments and has no ends to push towards.
		track.FillWidth().Justify(ui.SpaceBetween)
	}
	track.Children(func() {
		for i, label := range labels {
			s := parts.Segment(i).Disabled(opts.Disabled).
				Padding(u, u*2.5).Radius(theme.ControlRadius - u*0.5).Shrink(0)
			var face ui.Color
			switch {
			case i == *selected:
				// Raised off the track rather than a colour: a chosen
				// segment keeps the library's one ink face in both
				// appearances, which is why MyGo's own does the same.
				face = segmentFace(c)
				s.Shadow(0, 1, 2, 0, ui.RGBA(0, 0, 0, 0.12))
			case s.Hovered():
				face = k.SurfaceHover
			}
			fg := k.Text
			if opts.Disabled {
				face, fg = k.Surface, k.TextFaint
			}
			s.Background(face).TextColor(fg)
			s.Children(func() {
				ui.Text(c, label).SingleLine().FontSize(core.FontSize(c, theme.BodySize))
			})
		}
	})
	if opts.Label != "" {
		track.Label(opts.Label)
	}
	group := track
	if opts.Scrollable {
		// The whole track scrolls rather than its contents, so the track's
		// own edges and its border travel together instead of the border
		// staying put with a sliding row inside it.
		scroller := ui.ScrollHorizontal(c).FillWidth()
		scroller.Children(func() { group = track })
		group = scroller
	}
	return group
}

// segmentFace is the face of a chosen segment: the window's own colour where
// it is light, so the segment lifts off a grey track, and the pressed grey
// where it is dark, where white would be a hole rather than a lift.
func segmentFace(c *ui.Context) ui.Color {
	if core.IsDark(c) {
		return core.Tokens(c).SurfacePressed
	}
	return core.Tokens(c).Background
}

// ChoiceChipsOptions configure a ChoiceChips.
type ChoiceChipsOptions struct {
	// Label names the group for assistive technology, as RadioGroup's does.
	Label string
	// Max caps how many chips may be chosen at once; zero is no cap. A chip
	// pressed at the cap refuses rather than quietly doing nothing: the face
	// does not change, so a person can see it did not take.
	Max int
	// Wrap lets the chips run onto a second line rather than overflow, which
	// is what a filter bar narrower than its own filters needs.
	Wrap bool
	// Disabled greys out every chip.
	Disabled bool
}

// ChoiceChips is a run of tags, each one a choice: chosen chips are filled
// with ink, the rest are surface. It is the filter bar's control and the
// tag-picker's, and it differs from CheckboxGroup in that the chosen state
// is meant to be seen across the page rather than read down a column.
//
// The chips draw from the caller's slice rather than from the press, so a
// chip refused by Max never lights up for a frame as if it had taken.
func ChoiceChips(c *ui.Context, selected *[]string, choices []Choice, opts ChoiceChipsOptions) *ui.Element {
	if selected == nil {
		panic("input: ChoiceChips needs a selection to point at")
	}
	if len(choices) == 0 {
		panic("input: ChoiceChips needs at least one choice")
	}
	if opts.Max < 0 {
		panic("input: ChoiceChips Max cannot be negative")
	}
	for i, ch := range choices {
		if ch.Label == "" {
			panic("input: ChoiceChips choice " + strconv.Itoa(i) + " has no Label; a chip with no words on it is a pill")
		}
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	row := ui.Row(c).Gap(u).AlignItems(ui.Center)
	if opts.Wrap {
		row.Wrap().AlignContent(ui.Start)
	}
	if opts.Label != "" {
		row.Label(opts.Label)
	}
	row.Children(func() {
		for _, ch := range choices {
			chosen := hasValue(*selected, ch.Value)
			on := chosen
			// ToggleBase is the behaviour underneath: it flips the bool it
			// is given on a click or on Space and reports it as a change,
			// which is all a chip is.
			chip := ui.ToggleBase(c, &on).Padding(u*0.75, u*2).
				Radius(theme.PillRadius).Shrink(0)
			if !opts.Wrap {
				chip.NoWrap()
			}
			if chip.Changed() {
				if on {
					if opts.Max > 0 && len(*selected) >= opts.Max {
						// Refused: the face below still reads from the
						// caller's slice, so it does not change.
					} else if !hasValue(*selected, ch.Value) {
						*selected = append(*selected, ch.Value)
					}
				} else {
					*selected = without(*selected, ch.Value)
				}
			}
			bg, fg := k.Surface, k.Text
			switch {
			case chosen:
				bg, fg = k.Fill, k.OnFill
			case chip.Hovered():
				bg = k.SurfaceHover
			}
			if opts.Disabled {
				bg, fg = k.Surface, k.TextFaint
			}
			chip.Background(bg).TextColor(fg).Tooltip(ch.Label)
			chip.Children(func() {
				ui.Text(c, ch.Label).SingleLine().
					FontSize(core.FontSize(c, theme.RowSize))
			})
		}
	})
	return row
}
