package input

import (
	"strconv"
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// The two ways a file arrives without a person typing its name: chosen from a
// list the caller has, and dragged out of another window.
//
// Neither of them reaches for a system dialog or for the filesystem. A picker
// cannot know what the caller is allowed to see — a directory of the clinic's
// records and a directory of a website's assets are both "files", and only
// the caller knows which it may open — so the list of entries is the caller's
// and the picker is the box, the panel and the writing of the path. Dragging
// is the other way round: the host is the one that hands paths over, and
// DroppedFiles is that.

// FileEntry is one thing in a directory a FilePicker is showing.
type FileEntry struct {
	// Path is the entry's own path, and what the picker writes into the
	// caller's string when it is chosen. It is required: a row a caller
	// cannot open is a row that only takes up space.
	Path string
	// Name is the words shown for it. Empty shows the last element of the
	// path, which is what a directory listing's own name is.
	Name string
	// Dir says the entry is a directory, which the row marks with a folder
	// and a press reports as Up rather than as Chosen — the caller lists the
	// directory's contents, because only it knows how.
	Dir bool
	// Disabled takes the row out of play, for a file the caller cannot open.
	Disabled bool
	// Tip is the row's tooltip, for a row whose name has to be cut short.
	Tip string
}

// FilePickerOptions configure a FilePicker.
type FilePickerOptions struct {
	// Label names the field; it is required, as for every field here. A
	// path is a value and not a name: it changes as the user browses, so a
	// reader cannot be given one to read out.
	Label string
	// Placeholder is what the field says while no path is chosen. Empty
	// takes the library's own "Choose a file", which is a phrase every
	// language here has.
	Placeholder string
	// Error marks the path as not valid.
	Error string
	// Disabled greys the field out.
	Disabled bool
	// Clearable puts a button at the trailing edge that empties the path,
	// because a path is the one value in this library that a person can
	// only otherwise replace by browsing all the way back to its root.
	Clearable bool
	// Width is the field's own width; zero fills the row it is in.
	Width float32
	// Tip is the field's tooltip: a path is longer than any field, and the
	// full one is what a person wants to check before saving.
	Tip string
}

// FilePickerResult carries a FilePicker and what was chosen in it.
type FilePickerResult struct {
	// Element is the field: the path, the browse button, the panel while it
	// is showing.
	Element *ui.Element
	// picked is the path chosen this frame, "" for none.
	picked string
	// intoDir is that a directory was opened this frame, which is the cue to
	// list it rather than to save anything.
	intoDir bool
	// cleared reports the clear button.
	cleared bool
}

// Picked reports the path chosen this frame, "" when nothing was. It is a
// path rather than a bool because the caller may want to act on which file it
// was — "that one is already uploaded" — without comparing paths itself.
func (r FilePickerResult) Picked() string { return r.picked }

// Opened reports that a directory was entered this frame, so the caller can
// list it. It is separate from Picked because a directory is not a file: a
// picker that treated opening one as choosing it would hand a directory to
// whatever was going to save the path.
func (r FilePickerResult) Opened() bool { return r.intoDir }

// Cleared reports that the clear button was pressed this frame.
func (r FilePickerResult) Cleared() bool { return r.cleared }

// FilePicker is a field holding a path, with a panel of the directory the
// caller is showing under it.
//
// The path is the caller's string and the entries are the caller's slice, so
// the picker holds neither. It cannot be otherwise: a picker that listed a
// directory itself would be a second file browser inside an application that
// quite possibly already has one, in a place it quite possibly may not go.
//
// Enter takes a path typed into the field, which is the one way a picker is
// usable without the panel: pasting a path out of a terminal is how most
// people would rather answer.
func FilePicker(c *ui.Context, path *string, entries []FileEntry, opts FilePickerOptions) FilePickerResult {
	if path == nil {
		panic("input: FilePicker needs a path to point at; it keeps no path of its own")
	}
	if opts.Label == "" {
		panic("input: FilePicker needs a Label; a path is the one field whose own text is " +
			"the value, so it has no name of its own to read out")
	}
	for i, e := range entries {
		if e.Path == "" {
			panic("input: FilePicker entry " + strconv.Itoa(i) + " has no Path; a row nothing " +
				"can be opened from is a row that only takes up space")
		}
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	placeholder := opts.Placeholder
	if placeholder == "" {
		placeholder = core.Msg(c, "input.chooseFile", core.Def("Choose a file"))
	}
	browse := core.Msg(c, "input.browse", core.Def("Browse"))
	clearName := core.Msg(c, "input.clearPath", core.Def("Clear path"))
	s := fieldSkin{
		label: opts.Label, err: opts.Error,
		disabled: opts.Disabled, width: opts.Width,
	}

	var r FilePickerResult
	r.Element = textWell(c, s, func(w *ui.Element) *ui.Element {
		// The panel's own open flag lives on the field's box rather than on
		// the caller's state: whether a list is showing is a moment of the
		// interface, not a fact about the file, and a caller holding it would
		// have to remember to close it after every choice.
		open := ui.Local(w, openKey{}, func() bool { return false })

		in := bareInput(c, path, opts.Label, placeholder).
			Grow(1).MinHeight(core.ControlHeight(c) - u*3)
		dressInput(in, opts.Error, false)
		if opts.Tip != "" || *path != "" {
			in.Tooltip(opts.Tip + shortPath(*path))
		}

		button := ui.ButtonBase(c).Shrink(0).Radius(theme.PillRadius).
			Height(core.ControlHeight(c)-u*2).Padding(0, u*2.5, 0, u*2.5).
			Background(k.Surface).TextColor(k.Text).Role(ui.RoleButton).
			Label(browse).Tooltip(browse).Disabled(opts.Disabled)
		if button.Clicked() {
			*open = !*open
		}
		button.Children(func() {
			ui.Icon(c, glyphFolder).TextColor(k.TextMuted).Size(u*4, u*4)
			ui.Text(c, browse).SingleLine().FontSize(core.FontSize(c, theme.RowSize))
		})

		if opts.Clearable && *path != "" {
			clear := ui.Box(c).Size(u*7, u*7).Shrink(0).Role(ui.RoleButton).
				Label(clearName).Tooltip(clearName).Cursor(ui.CursorPointer).
				Children(func() {
					ui.Icon(c, glyphCross).TextColor(k.TextMuted).Size(u*3.5, u*3.5)
				})
			if clear.Clicked() {
				*path = ""
				r.cleared = true
				*open = false
				c.Invalidate()
			}
		}

		ui.PopoverBase(c, button, open, func(panel *ui.Element) {
			panelFace(c, panel)
			if len(entries) == 0 {
				empty(c, core.Msg(c, "input.emptyFolder", core.Def("Nothing here")))
				return
			}
			popupList(c, dropdownMaxHeight).Children(func() {
				for _, e := range entries {
					entry := e
					name := entry.Name
					if name == "" {
						name = baseName(entry.Path)
					}
					tip := entry.Tip
					if tip == "" {
						tip = entry.Path
					}
					row := optionRow(c, name, optionFace{}).
						Tooltip(tip).Disabled(entry.Disabled)
					row.Children(func() {
						// The mark is a folder or a file, and it is beside the
						// name rather than instead of it: an icon-only row is
						// a row a reader cannot name.
						ui.Icon(c, glyphOf(entry.Dir)).TextColor(k.TextMuted).
							Size(u*4, u*4).Shrink(0)
						ui.Text(c, name).SingleLine().
							FontSize(core.FontSize(c, theme.BodySize)).Grow(1)
					})
					if !row.Clicked() {
						continue
					}
					if entry.Dir {
						r.intoDir = true
						*open = false
						continue
					}
					*path = entry.Path
					r.picked = entry.Path
					*open = false
					// The editor reads the pointer back at the top of the
					// next frame, so ask for that frame: the path has to be
					// in the field before the panel closes over it.
					c.Invalidate()
				}
			})
		})
		return in
	})
	return r
}

// FileDropZoneOptions configure a FileDropZone.
type FileDropZoneOptions struct {
	// Label names the zone for assistive technology, and is required: the
	// zone is a rectangle with words in it, and a reader meeting a rectangle
	// with words in it has no idea what it takes.
	Label string
	// Text says what dropping here does. Empty takes the library's own
	// "Drop files here", which is the phrase every file target has.
	Text string
	// Multiple takes a drag of any number of files. Off by default, because
	// a target that takes many and uses the first is a target that silently
	// throws the rest away.
	Multiple bool
	// Height is the zone's own height; zero is enough for its words at this
	// window's density, which is what a drop target wants to be: small
	// enough to drop onto, big enough to point at.
	Height float32
	// Disabled takes the zone out of play, and is what an upload in progress
	// sets: a target that accepts files it is about to reject is worse than
	// one that says nothing.
	Disabled bool
}

// FileDropZoneResult carries a FileDropZone and what was dropped on it.
type FileDropZoneResult struct {
	// Element is the zone.
	Element *ui.Element
	// files are the paths dropped on it since the last frame, nil for none.
	// It is the caller's to read and its own copy afterwards: the paths are
	// handed over once, and a caller that missed the frame has missed the
	// drop — which is what MyGo's own drop contract says, and pretending
	// otherwise here would mean this package keeping a queue of its own.
	files []string
	// over is that files were being dragged over it as this frame was built.
	over bool
}

// Files are the paths dropped on the zone since the last frame.
func (r FileDropZoneResult) Files() []string { return r.files }

// Over reports that files were being dragged over the zone as this frame was
// built, which is the whole of what the zone's highlight is for: the drag is
// the only moment a person can find out whether the target will take it.
func (r FileDropZoneResult) Over() bool { return r.over }

// FileDropZone is a target that takes files dragged into the window from
// somewhere else — a browser's downloads, the desktop, another application.
//
// It holds nothing, and says nothing. The paths are the host's and arrive
// once; the caller's job is to upload them and to know where they went. A
// drop zone that also queued them would be a second copy of an upload that
// the caller is already doing, and two copies of an upload is an upload that
// happens twice.
func FileDropZone(c *ui.Context, opts FileDropZoneOptions) FileDropZoneResult {
	if opts.Label == "" {
		panic("input: FileDropZone needs a Label; the zone is a rectangle with words in it, " +
			"and a reader meeting one has no idea what it takes")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	say := opts.Text
	if say == "" {
		say = core.Msg(c, "input.dropFiles", core.Def("Drop files here"))
	}
	h := opts.Height
	if h <= 0 {
		h = core.ControlHeight(c) * 2.5
	}
	if opts.Disabled {
		say = core.Msg(c, "input.dropDisabled", core.Def("Not accepting files"))
	}

	var r FileDropZoneResult
	// The zone is a closure rather than a value because it reads its own
	// press state: an element has to exist before its own Children call
	// runs, and the answer is wanted inside that call.
	zone := func() *ui.Element {
		e := ui.Box(c).FillWidth().Height(h).Radius(theme.ControlRadius).
			Background(k.Surface).BorderWidth(theme.BorderWidth).BorderColor(k.Border).
			Cursor(ui.CursorPointer).Role(ui.RoleNone).Label(opts.Label)
		return e.Children(func() {
			// Read before drawing the face: the highlight below has to be the
			// answer to this frame's question and not the last one's.
			dropped := e.DroppedFiles()
			r.over = e.FileDragOver()
			if !opts.Disabled && len(dropped) > 0 {
				if !opts.Multiple {
					dropped = dropped[:1]
				}
				r.files = dropped
			}

			switch {
			case r.over:
				e.Background(k.AccentBg).BorderColor(k.Accent)
			case r.files != nil:
				e.Background(k.SuccessBg).BorderColor(k.Success)
			}
			ui.Row(c).FillWidth().AlignItems(ui.Center).Justify(ui.Center).Gap(u).Children(func() {
				// The mark leads the words in both states, so the two reads
				// as one control that is busy rather than as two controls.
				mark, ink := glyphDrop, k.TextMuted
				switch {
				case r.over:
					mark, ink = glyphTick, k.Accent
				case r.files != nil:
					mark, ink = glyphTick, k.Success
				}
				ui.Icon(c, mark).TextColor(ink).Size(u*5, u*5).Shrink(0)
				ui.Text(c, say).SingleLine().TextColor(ink).
					FontSize(core.FontSize(c, theme.BodySize))
			})
		})
	}
	r.Element = zone()
	return r
}

// baseName is the last element of a path. It is one call rather than a use of
// path.Base, because a Windows path has backslashes and path.Base splits on
// slashes only: the tail of "C:\\Users\\sam\\notes.md" would come back as
// the whole string, which is a tooltip showing more than it should.
func baseName(p string) string {
	if i := strings.LastIndexAny(p, "/\\"); i >= 0 {
		return p[i+1:]
	}
	return p
}

// shortPath is the tail of a path, which is what a tooltip wants beside the
// field's own label: a person checking a path wants the end of it, because the
// beginning is the same directory they are already in.
//
// It is filepath.Base for a path with no separators and itself otherwise, so
// a path typed into the field on a system whose separator is not a slash is
// not mangled into its last character.
func shortPath(p string) string {
	if p == "" {
		return ""
	}
	if !strings.ContainsRune(p, '/') && !strings.ContainsRune(p, '\\') {
		return p
	}
	return baseName(p)
}
