package agent

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/feedback"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// Artifact is one thing a run produced: a report, a screenshot, a table it
// wrote out.
type Artifact struct {
	// Name is what the thing is called. It is required.
	Name string
	// Kind is what it is — "HTML", "PNG", "CSV". It is drawn as a chip
	// because the kind decides whether the thing can be opened at all, and a
	// reader picks an artifact by kind before they pick it by name.
	Kind string
	// Size is how big it is, already formatted.
	Size string
	// Version is which version this is, counting from one. Zero means it has
	// only ever had one, and then the version is not drawn at all: "v1" on
	// the only version of something is a number about nothing.
	Version int
	// Status is where producing it is.
	Status Status
	// Note is one line about it.
	Note string
}

// ArtifactPanelResult carries an ArtifactPanel and what was pressed in it.
type ArtifactPanelResult struct {
	// Element is the panel.
	Element *ui.Element
	// selected is the artifact pressed this frame, or -1.
	selected int
	opened   string
	answered bool
}

// Selected is the artifact pressed this frame, counted from zero, and -1 for
// none.
func (r ArtifactPanelResult) Selected() int {
	if !r.answered {
		return -1
	}
	return r.selected
}

// Opened is the name of the artifact that was opened this frame, empty for
// none.
func (r ArtifactPanelResult) Opened() string { return r.opened }

// ArtifactPanelOptions configure an ArtifactPanel.
type ArtifactPanelOptions struct {
	// Title heads the panel. It is required: a panel of things a run produced
	// is the one panel whose contents do not say what they are.
	Title string
	// Artifacts are the artifacts, in the order they were made.
	Artifacts []Artifact
	// Selected is the artifact marked as the chosen one, -1 for none.
	Selected *int
	// Empty draws instead of the artifacts when there are none, and is
	// required in that case.
	Empty string
	// ShowVersions numbers the versions each artifact has reached, beside
	// its chip.
	ShowVersions bool
}

// ArtifactPanel is what a run produced: each thing on its own row, with what
// kind it is and how big it got.
//
// The rows are pressable across their surface for the same reason an agent
// card's is: the thing a reader is choosing between two of is the artifact,
// and an "Open" button in the corner of each row is a button nobody presses.
// What opening one means is the caller's; the panel only says which.
func ArtifactPanel(c *ui.Context, opts ArtifactPanelOptions) ArtifactPanelResult {
	u := core.Density(c).Unit()
	if opts.Title == "" {
		panic("agent: ArtifactPanel needs a Title; its contents do not say what they are")
	}
	if len(opts.Artifacts) == 0 && opts.Empty == "" {
		panic("agent: ArtifactPanel with no artifacts needs opts.Empty")
	}
	for i, a := range opts.Artifacts {
		if a.Name == "" {
			panic("agent: ArtifactPanel artifact " + itoa(i) + " has no Name")
		}
	}
	var r ArtifactPanelResult
	r.selected = -1

	panel := panel(c, layout.ContainerOptions{
		Surface: true, Radius: theme.CardRadius, Pad: u * 2.5, Gap: u * 1.5,
	}, func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 2).Children(func() {
			display.Text(c, opts.Title, display.TextOptions{Bold: true, MaxLines: 1})
			ui.Box(c).Grow(1)
			display.Text(c, itoa(len(opts.Artifacts))+" artifacts",
				display.TextOptions{Muted: true, MaxLines: 1})
		})
		if len(opts.Artifacts) == 0 {
			display.Text(c, opts.Empty, display.TextOptions{Muted: true})
			return
		}
		layout.Divider(c, layout.DividerOptions{})
		ui.Column(c).FillWidth().Gap(u * 0.5).Children(func() {
			for i, a := range opts.Artifacts {
				i, a := i, a
				artifactRow(c, a, artifactRowOptions{
					selected: opts.Selected != nil && *opts.Selected == i,
					versions: opts.ShowVersions,
				}, func() {
					r.answered = true
					r.selected = i
				}, func() {
					r.opened = a.Name
				})
			}
		})
	})
	r.Element = panel
	return r
}

// artifactRowOptions is one artifact row's own settings.
type artifactRowOptions struct {
	selected bool
	// versions says whether the panel was asked to show version numbers.
	// Without it "v1" would sit beside the only version of a file, which is a
	// number about nothing.
	versions bool
}

// artifactRow is one artifact in the panel.
func artifactRow(c *ui.Context, a Artifact, opts artifactRowOptions, onPress, onOpen func()) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	row := ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u*1.5).
		Padding(u*0.75, u).Radius(theme.SmallRadius).Children(func() {
		display.Icon(c, display.IconPaperclip, display.IconOptions{
			Name: a.Name, Size: u * 4, Muted: true,
		})
		ui.Column(c).Grow(1).Shrink(0).Gap(0).Children(func() {
			ui.Text(c, a.Name).TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).MaxLines(1)
			if a.Note != "" {
				ui.Text(c, a.Note).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize)).MaxLines(1)
			}
		})
		if a.Version > 0 && opts.versions {
			display.Tag(c, "v"+itoa(a.Version), display.TagOptions{Tone: core.Accent})
		}
		if a.Kind != "" {
			display.Tag(c, a.Kind, display.TagOptions{Tone: core.Neutral})
		}
		if a.Size != "" {
			display.Text(c, a.Size, display.TextOptions{Faint: true, MaxLines: 1})
		}
		AgentStatus(c, AgentStatusOptions{Status: a.Status})
		if display.Icon(c, display.IconExternal, display.IconOptions{
			Name: "Open " + a.Name, Size: u * 4, Muted: true,
		}).Clicked() {
			if onOpen != nil {
				onOpen()
			}
		}
	})
	row.Label(a.Name)
	if opts.selected {
		row.Background(k.SurfacePressed)
	}
	if onPress != nil && row.Clicked() {
		onPress()
	}
}

// ArtifactVersion is one version of one artifact.
type ArtifactVersion struct {
	// Number is the version, counting from one. It is required: a version
	// with no number is a revision with nothing to compare it to.
	Number int
	// Label is what the version is called when the number is not the whole
	// story — "after the review", "the one with the fixture".
	Label string
	// Note is one line about what changed.
	Note string
	// Status is where making it is.
	Status Status
}

// ArtifactVersionSwitcherResult carries an ArtifactVersionSwitcher and what
// was chosen in it.
type ArtifactVersionSwitcherResult struct {
	// Element is the switcher.
	Element  *ui.Element
	chosen   int
	answered bool
}

// Chosen is the version picked this frame, counted from zero, and -1 in a
// frame in which nothing was.
func (r ArtifactVersionSwitcherResult) Chosen() int {
	if !r.answered {
		return -1
	}
	return r.chosen
}

// ArtifactVersionSwitcherOptions configure an ArtifactVersionSwitcher.
type ArtifactVersionSwitcherOptions struct {
	// Name is what the artifact is called, drawn before the versions. It is
	// required: "3" beside nothing says nothing about what is being counted.
	Name string
	// Versions are the versions, oldest first.
	Versions []ArtifactVersion
	// Current is the version on show, counted from zero, and the caller's.
	Current *int
}

// ArtifactVersionSwitcher is one artifact's run of versions, with the one on
// show named.
//
// The chosen version wears the accent's own pair and the others are left
// quiet, rather than the other way round: a switcher is read by finding the
// one that stands out, and a page of seven equally loud versions is a page
// with nothing on it. The *int is the caller's because which version of an
// artifact a window is showing is a fact about the window, not about the
// switcher.
func ArtifactVersionSwitcher(c *ui.Context, opts ArtifactVersionSwitcherOptions) ArtifactVersionSwitcherResult {
	u := core.Density(c).Unit()
	if opts.Name == "" {
		panic("agent: ArtifactVersionSwitcher needs a Name; versions of nothing cannot be chosen " +
			"between")
	}
	if len(opts.Versions) == 0 {
		panic("agent: ArtifactVersionSwitcher needs at least one version")
	}
	for i, v := range opts.Versions {
		if v.Number <= 0 {
			panic("agent: ArtifactVersionSwitcher version " + itoa(i) + " has no Number")
		}
	}
	cur := 0
	if opts.Current != nil && *opts.Current >= 0 && *opts.Current < len(opts.Versions) {
		cur = *opts.Current
	}
	k := core.Tokens(c)
	var r ArtifactVersionSwitcherResult

	e := ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.5).Children(func() {
		ui.Text(c, opts.Name).TextColor(k.TextMuted).
			FontSize(core.FontSize(c, theme.CaptionSize)).MaxLines(1).Shrink(0)
		for i, v := range opts.Versions {
			i, v := i, v
			bg, fg := k.Surface, k.Text
			if i == cur {
				bg, fg = core.Accent.Pair(k)
			}
			label := "v" + itoa(v.Number)
			if v.Label != "" {
				label += " " + v.Label
			}
			pill := ui.Box(c).Padding(u*0.75, u*2).Radius(theme.PillRadius).Shrink(0).
				Background(bg).Border(theme.BorderWidth, k.Border).AlignItems(ui.Center).
				Children(func() {
					ui.Text(c, label).TextColor(fg).
						FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
				})
			pill.Label(opts.Name + " " + label)
			if pill.Clicked() {
				if opts.Current != nil {
					*opts.Current = i
				}
				r.answered = true
				r.chosen = i
			}
		}
	})
	e.Label(opts.Name + " versions")
	r.Element = e
	return r
}

// Screenshot is one picture a run took of what it was doing.
type Screenshot struct {
	// Name is what the picture is of — "the gallery page", "the failing
	// test". It is required.
	Name string
	// At is when it was taken, already formatted.
	At string
	// Note is one line about it.
	Note string
	// Pixels is the image's own size, formatted by the caller, and is drawn
	// in the caption rather than used to size the frame: a stream of
	// thumbnails at their own sizes is a stream that jumps about underneath
	// the reader as it scrolls.
	Pixels string
}

// ScreenshotStreamOptions configure a ScreenshotStream.
type ScreenshotStreamOptions struct {
	// Shots are the pictures, oldest first.
	Shots []Screenshot
	// Selected is the picture marked as the chosen one, -1 for none.
	Selected *int
	// Columns is how many pictures sit in a row. Zero takes three.
	Columns int
	// Frame is one picture's own width. Zero shares the row evenly between
	// the pictures, which is what a stream of equal pictures wants; a number
	// gives every frame the same width whatever it is showing.
	Frame float32
	// Body draws inside one picture's frame, for a caller holding the actual
	// bitmaps. Nil leaves a skeleton in its place, which is what a picture
	// that has been asked for but has not arrived yet should look like.
	Body func(c *ui.Context, shot Screenshot)
	// Empty draws instead of the pictures when there are none, and is
	// required in that case.
	Empty string
}

// ScreenshotStream is a run's pictures in a grid, oldest first.
//
// It takes a Body rather than bitmaps because a picture is fetched
// asynchronously and a stream is rebuilt from the transcript every time the
// transcript arrives: the caller that holds the bitmaps draws them through
// Body, and the one that has not fetched them yet gets a frame that says so.
// A component that took a *ui.Bitmap could only ever be drawn once every
// picture had arrived, which is the one moment the component is least wanted.
func ScreenshotStream(c *ui.Context, opts ScreenshotStreamOptions) *ui.Element {
	u := core.Density(c).Unit()
	if len(opts.Shots) == 0 && opts.Empty == "" {
		panic("agent: ScreenshotStream with no shots needs opts.Empty")
	}
	if opts.Empty != "" && len(opts.Shots) == 0 {
		return emptyCard(c, opts.Empty)
	}
	cols := opts.Columns
	if cols <= 0 {
		cols = 3
	}
	return ui.Column(c).FillWidth().Gap(u * 1.5).Children(func() {
		ui.Row(c).FillWidth().Wrap().GapX(u * 1.5).GapY(u * 1.5).Children(func() {
			for i, shot := range opts.Shots {
				i, shot := i, shot
				frame := func() {
					shotFrame(c, shot, shotFrameOptions{
						selected: opts.Selected != nil && *opts.Selected == i,
						body:     opts.Body,
					})
				}
				// The ratio is fixed at 16:10 for every picture whatever its
				// own shape, because a stream whose pictures are different
				// heights is a stream whose captions never line up.
				if opts.Frame > 0 {
					ui.Box(c).Width(opts.Frame).Shrink(0).Children(func() {
						layout.AspectRatio(c, layout.AspectRatioOptions{Ratio: 1.6}, frame)
					})
					continue
				}
				layout.AspectRatio(c, layout.AspectRatioOptions{Ratio: 1.6}, frame)
			}
		})
	})
}

// shotFrameOptions is one picture's own settings.
type shotFrameOptions struct {
	selected bool
	// body is the caller's way of drawing the picture itself.
	body func(c *ui.Context, shot Screenshot)
}

// shotFrame is one picture: the frame, whatever goes in it, and the caption.
func shotFrame(c *ui.Context, shot Screenshot, opts shotFrameOptions) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	frame := panel(c, layout.ContainerOptions{
		Radius: theme.CardRadius, Pad: u * 1.25, Gap: u, Surface: true,
	}, func() {
		// The picture's own box, drawn inside the frame rather than as the
		// frame: the caption needs to be beside the picture, not inside it.
		pic := ui.Box(c).FillWidth().Radius(theme.SmallRadius).Shrink(0).
			Background(k.SurfaceHover).Clip().Label(shot.Name)
		if opts.selected {
			pic.Border(theme.BorderWidth*2, k.Accent)
		}
		pic.Children(func() {
			if opts.body != nil {
				opts.body(c, shot)
				return
			}
			feedback.Skeleton(c, feedback.SkeletonOptions{
				FillWidth: true, Height: core.FontSize(c, theme.BodySize) * 3,
				Label: shot.Name + " loading",
			})
		})
		ui.Column(c).FillWidth().Gap(0).Children(func() {
			ui.Text(c, shot.Name).TextColor(k.Text).
				FontSize(core.FontSize(c, theme.CaptionSize)).MaxLines(1)
			caption := join(" · ", shot.At, shot.Pixels, shot.Note)
			if caption != "" {
				ui.Text(c, caption).TextColor(k.TextFaint).
					FontSize(core.FontSize(c, theme.CaptionSize)).MaxLines(1)
			}
		})
	})
	frame.Label(shot.Name)
}
