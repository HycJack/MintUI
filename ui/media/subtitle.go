package media

import (
	"strings"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/data"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// Cue is one line of a subtitle: when it appears and what it says.
//
// A cue's times are relative to the file, not to the cue before it. Relative
// times are how subtitles are authored everywhere else and how every existing
// file on disk is written, so a cue list that stored each line's length would
// need a conversion nobody asked for — and would lose the file's own times the
// first time it rewrote one.
type Cue struct {
	// Start and End are the cue's own times in the file.
	Start, End time.Duration
	// Text is the line.
	Text string
	// Speaker is who is talking, for a subtitle that has to attribute lines —
	// a transcript of a call rather than a film.
	Speaker string
}

// Duration is how long the cue is on screen. A cue whose end is before its
// start is zero rather than negative: the editor draws it as a zero-width bar
// for the reader to fix, which is visible, where a negative width is a
// rectangle drawn on the other side of the track.
func (c Cue) Duration() time.Duration {
	if c.End <= c.Start {
		return 0
	}
	return c.End - c.Start
}

// WordCount is how many words the line has, which is what a reader checks
// against the standard and what the editor warns about. Splitting on space
// rather than on the script's own word boundaries is deliberate: the counts
// every subtitle guide in the world uses are word counts, and matching the
// guide is what makes the number mean something.
func (c Cue) WordCount() int {
	return len(strings.Fields(c.Text))
}

// ReadingRate is how many words a subtitle reader gets through in a minute.
// 160 is the figure broadcast guidelines quote and the figure a reader who has
// watched a few thousand subtitles has calibrated to; a viewer reading a
// foreign-language track reads slower, which is what SubtitleEditorOptions'
// ReadRate is for.
const ReadingRate = 160.0

// ReadTime is how long words words take to read at rate words a minute — the
// arithmetic every subtitle timing check in the world is doing, written once
// so the editor and the warning cannot drift apart.
func ReadTime(words int, rate float64) time.Duration {
	if words <= 0 || rate <= 0 {
		return 0
	}
	return time.Duration(float64(words) / rate * float64(time.Minute))
}

// NeedsLonger reports whether the cue's time is too short for its words at the
// standard reading speed, and is what a subtitle editor's warning is built on.
//
// It is a warning rather than a limit. The number is a guide, not a rule, and
// an editor that refused to save a cue because it was a tenth of a second short
// would be stopping the person who has watched the footage and knows the line
// has to go up before the speaker finishes saying it.
func (c Cue) NeedsLonger() bool {
	return c.Duration() < ReadTime(c.WordCount(), ReadingRate)
}

// subtitleColumns are the editor's table's columns. They are a package-level
// value rather than a literal inside the component so that the widths, the
// ids and the titles cannot drift apart between the head and the sort.
var subtitleColumns = []data.Column{
	{Title: "Start", ID: "start", Width: 96, Align: ui.End, Sortable: true},
	{Title: "End", ID: "end", Width: 96, Align: ui.End, Sortable: true},
	{Title: "Line", ID: "text", Share: 1},
	{Title: "Words", ID: "words", Width: 64, Align: ui.End, Sortable: true},
}

// SubtitleEditorOptions configure a SubtitleEditor.
type SubtitleEditorOptions struct {
	// Cues are the caller's lines, in file order.
	Cues []Cue
	// Selected is the cue being edited, as an index; -1 for none.
	Selected *int
	// Sort is the column the rows are ordered by, in the caller's state. The
	// table asks for a new one and the caller reorders Cues with it: the
	// editor holds no order of its own, so what it shows is always what it was
	// handed.
	Sort *data.Sort
	// Height is the table's height, and is required for the reason it is
	// required everywhere else in this library — a table with no height grows
	// to fit its rows, which is every row drawn rather than a table.
	Height float32
	// ReadRate overrides [ReadingTime] for a file written for a slower or a
	// faster audience. Zero takes the standard.
	ReadRate float64
	// State and Scroll keep the table's place and its sideways scroll between
	// frames; either may be nil.
	State  *ui.ListState
	Scroll *ui.ScrollState
}

// SubtitleEditorResult carries a SubtitleEditor.
type SubtitleEditorResult struct {
	// Element is the editor.
	Element *ui.Element
	// sorted is the column whose head was pressed this frame.
	sorted string
}

// Sorted is the column whose head was pressed this frame, empty when none
// was. The caller reorders Cues with it and asks again — the same contract
// data.DataTable has, passed through rather than wrapped, so a table that
// behaves differently here is not a thing that can happen.
func (r SubtitleEditorResult) Sorted() string { return r.sorted }

// SubtitleEditor is the list of a file's subtitles: when each line goes up,
// what it says, and how long it is there for.
//
// It is data.DataTable rather than a hand-built list of rows, for the reason
// a list is a list: the columns keep their widths so a time does not wrap, the
// head stays above the scroll, and only the rows in view are built — which
// matters here because a three-hour file has three thousand lines in it.
func SubtitleEditor(c *ui.Context, opts SubtitleEditorOptions) SubtitleEditorResult {
	if opts.Selected == nil {
		panic("media: SubtitleEditor needs a Selected cue to point at; it owns no list of " +
			"its own")
	}
	if opts.Height <= 0 {
		panic("media: SubtitleEditor needs a Height; a table with no height grows to fit " +
			"every row rather than scrolling")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	rate := opts.ReadRate
	if rate <= 0 {
		rate = ReadingRate
	}

	var res SubtitleEditorResult
	// The table is built inside the container below rather than beside it, for
	// the same reason: made out here it belongs to whatever holds the editor,
	// and the panel comes out with the table above the heading that is meant
	// to be over it. Its answer is read after the panel is built.
	var table data.DataTableResult
	res.Element = layout.Container(c, layout.ContainerOptions{
		Surface: true, Border: true, Radius: theme.CardRadius, Gap: u * 2,
	}, func() {
		ui.Text(c, "Subtitles").TextColor(k.Text).
			FontSize(core.FontSize(c, theme.RowSize)).Bold()
		table = data.DataTable(c, data.DataTableOptions{
			Columns: subtitleColumns,
			Rows:    len(opts.Cues),
			Sort:    opts.Sort,
			Cell: func(row, col int) {
				cue := opts.Cues[row]
				switch subtitleColumns[col].ID {
				case "start":
					ui.Text(c, TimeToText(cue.Start)).TextColor(k.Text).
						FontSize(core.FontSize(c, theme.RowSize))
				case "end":
					// A cue that needs longer than it has is written in the warning
					// tone rather than being flagged with an icon beside it: the
					// reader is scanning a column of times, and a column where one
					// cell is a different colour is found by the same glance as one
					// where a cell is the wrong colour.
					ink := k.Text
					if tooShort(cue, rate) {
						ink = k.Warning
					}
					ui.Text(c, TimeToText(cue.End)).TextColor(ink).
						FontSize(core.FontSize(c, theme.RowSize))
				case "text":
					ui.Text(c, cue.Text).TextColor(k.Text).
						FontSize(core.FontSize(c, theme.RowSize)).MaxLines(1)
				case "words":
					ui.Text(c, itoa(cue.WordCount())).TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.RowSize))
				}
			},
			CellLabel: func(row, col int) string {
				cue := opts.Cues[row]
				return cue.Text
			},
			Key:      func(row int) any { return cueKey(opts.Cues[row]) },
			Label:    func(row int) string { return opts.Cues[row].Text },
			Selected: opts.Selected,
			State:    opts.State,
			Scroll:   opts.Scroll,
			Height:   opts.Height,
			Empty:    func() { SubtitleEmpty(c) },
		})
		table.Element.FillWidth()
	})
	res.sorted = table.Sorted()
	return res
}

// cueKey identifies a cue by what it says and when, rather than by its index:
// an index stops identifying a line the moment the rows are sorted, and a row
// whose identity moves with the sort is a row that scrolls to the wrong place.
func cueKey(cue Cue) string {
	return TimeToText(cue.Start) + "|" + cue.Text
}

// tooShort is whether a cue's own time is under what its words need, at a
// given rate.
func tooShort(cue Cue, rate float64) bool {
	return cue.Duration() < ReadTime(cue.WordCount(), rate)
}

// SubtitleEmpty is what a file with no subtitles says. It is its own component
// rather than a string in the table because the table's Empty is drawn in the
// middle of a panel with a row-height grid behind it, and a line of text there
// reads as a broken row rather than as an empty state.
func SubtitleEmpty(c *ui.Context) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	return ui.Column(c).FillWidth().Center().Padding(u * 3).Gap(u).Label("No subtitles").
		Children(func() {
			ui.Text(c, core.Msg(c, "media.subtitles.empty", core.Def("No subtitles in this file"))).
				TextColor(k.TextMuted).FontSize(core.FontSize(c, theme.BodySize))
			ui.Text(c, core.Msg(c, "media.subtitles.empty.hint", core.Def("Import a .srt or write one below."))).
				TextColor(k.TextFaint).FontSize(core.FontSize(c, theme.CaptionSize))
		})
}

// itoa is a small count as a string. It is here so the table's cells do not
// each reach for strconv, and so a word count and a cue index are written the
// same way wherever they appear.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

// CueFields is one cue's times and speaker, as the row of fields under the
// table that edits it.
//
// It is part of the package because the table is read-only by design and a cue
// list with no way to write a cue back into it is a viewer. The two seconds
// fields are the caller's own numbers rather than derived ones, because a cue
// has sub-second times — 3.98s to 6.12s is normal in a subtitle file — and a
// field that showed whole seconds would quietly move every cue by up to half a
// second the first time anybody touched it.
//
// The seconds are written back into the cue after the fields have been built,
// so what the caller reads after a frame is what the fields showed this frame
// and not the frame before.
func CueFields(c *ui.Context, cue *Cue) *ui.Element {
	if cue == nil {
		panic("media: CueFields needs a cue to edit; it keeps no cue of its own")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	start := cue.Start.Seconds()
	end := cue.End.Seconds()
	nonNeg := func() *float64 { z := 0.0; return &z }

	return ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.5).
		Label("Cue times").Children(func() {
		input.NumberInput(c, &start, input.NumberInputOptions{
			Label: "Cue start, in seconds", Step: 0.01,
			Min: nonNeg(), Suffix: "s", NoSteppers: true, Width: u * 18,
		})
		input.NumberInput(c, &end, input.NumberInputOptions{
			Label: "Cue end, in seconds", Step: 0.01,
			Min: nonNeg(), Suffix: "s", NoSteppers: true, Width: u * 18,
		})
		ui.Box(c).Grow(1)
		if cue.Speaker != "" {
			display.Text(c, cue.Speaker, display.TextOptions{Muted: true, MaxLines: 1})
		} else {
			ui.Text(c, core.Msg(c, "media.subtitles.speaker", core.Def("No speaker"))).
				TextColor(k.TextFaint).FontSize(core.FontSize(c, theme.CaptionSize))
		}
		cue.Start = time.Duration(start * float64(time.Second))
		cue.End = time.Duration(end * float64(time.Second))
	})
}
