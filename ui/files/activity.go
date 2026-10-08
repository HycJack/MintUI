package files

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// RecentFile is one file somebody opened lately.
type RecentFile struct {
	// Name is the file's own name.
	Name string
	// Path is where it is.
	Path string
	// Opened is when it was opened, already formatted.
	Opened string
	// Where is which application opened it, or where it was opened from —
	// whichever the caller has, because the two are the same column.
	Where string
	// Kind is what it is; leave it Binary to work it out from the name.
	Kind Kind
}

// RecentFilesOptions configure a RecentFiles.
type RecentFilesOptions struct {
	// Height is the height of the list; required.
	Height float32
	// Max is how many to show at once. Zero shows all of them, which is
	// wrong for a list of two hundred and right for a menu of five.
	Max int
	// Empty draws instead of the list when there is nothing in it.
	Empty func()
}

// RecentFilesResult carries a RecentFiles and what was pressed in it.
type RecentFilesResult struct {
	// Element is the list.
	Element *ui.Element
	// picked is the path pressed this frame, empty for none.
	picked string
}

// Picked returns the path pressed this frame, empty for none.
func (r RecentFilesResult) Picked() string { return r.picked }

// RecentFiles is the files opened lately, most recent first.
//
// It is a list and not a table: there is nothing here to sort by, because the
// order is the whole of what the list is. Every column is something already
// sorted, and a header over a column that cannot be reordered is furniture.
func RecentFiles(c *ui.Context, files []RecentFile, opts RecentFilesOptions) RecentFilesResult {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if opts.Height <= 0 {
		panic("files: RecentFiles needs a Height; a recent list with no height is every file ever opened")
	}
	shown := files
	if opts.Max > 0 && len(shown) > opts.Max {
		shown = shown[:opts.Max]
	}

	var res RecentFilesResult
	list := ui.Scroll(c).Height(opts.Height).FillWidth().Gap(u * 0.5).Role(ui.RoleList)
	list.Children(func() {
		if len(shown) == 0 {
			if opts.Empty != nil {
				opts.Empty()
				return
			}
			ui.Text(c, core.Msg(c, "files.noRecent", core.Def("Nothing opened yet"))).
				TextColor(k.TextMuted).FontSize(core.FontSize(c, theme.RowSize))
			return
		}
		for _, f := range shown {
			f := f
			row := entryRow(c, f.Name, recentDetail(f), kindOrName(f.Kind, f.Name),
				false, 0, func() { res.picked = f.Path })
			row.Label(f.Opened + ": " + f.Name)
		}
	})
	res.Element = list
	return res
}

// recentDetail is the second line: which application opened it and when, with
// the time last because it is the part that makes the list a list.
func recentDetail(f RecentFile) string {
	switch {
	case f.Where != "" && f.Opened != "":
		return f.Where + core.Def(" · ") + f.Opened
	case f.Where != "":
		return f.Where
	}
	return f.Opened
}

// kindOrName is the kind the caller said, or the one the name says.
func kindOrName(kind Kind, name string) Kind {
	if kind != Binary {
		return kind
	}
	return KindOf(name)
}

// Transfer is one file moving.
type Transfer struct {
	// Name is the file's name.
	Name string
	// Path is where it is going to, or coming from.
	Path string
	// Done and Total are the counts of bytes, so a transfer of two files is
	// two entries rather than one with a filename in it.
	Done, Total int64
	// PerSecond is the current rate; zero hides the figure rather than
	// saying zero, because zero bytes a second is not what a transfer that
	// has not started yet means.
	PerSecond int64
	// Failed reports that this one gave up. A failed transfer stays in the
	// list rather than disappearing, because the thing that went wrong is
	// the thing somebody has to know about.
	Failed bool
	// Error is why it failed, when there is a reason.
	Error string
}

// TransferQueueOptions configure a TransferQueue.
type TransferQueueOptions struct {
	// Height is the height of the queue; required.
	Height float32
	// WithRate puts the rate beside the bar. Off by default because a rate
	// that flickers is the one figure in a transfer queue nobody reads.
	WithRate bool
	// Empty draws instead of the queue when there is nothing in it.
	Empty func()
}

// TransferQueueResult carries a TransferQueue and what was pressed in it.
type TransferQueueResult struct {
	// Element is the queue.
	Element *ui.Element
	// cancelled is the transfer given up on this frame, -1 for none.
	cancelled int
	// retried is the failed transfer asked to go again this frame, -1 for
	// none.
	retried int
}

// Cancelled returns the index of the transfer given up on this frame, -1 for
// none. The index rather than the path, because a queue can hold two entries
// for one path — a retry and the original — and cancelling "the file" would
// cancel the wrong one.
func (r TransferQueueResult) Cancelled() int { return r.cancelled }

// Retried returns the index of the failed transfer asked to go again this
// frame, -1 for none.
func (r TransferQueueResult) Retried() int { return r.retried }

// TransferQueue is the files going in and out, one row each, with the bar
// and the rate and a way to stop one.
//
// A finished transfer leaves. A failed one stays, marked, with a button to
// try it again: a queue that drops a failure the moment it happens is a queue
// that has lost the only entry anybody needed to read. That is why Failed is
// a field here rather than something the caller filters out.
func TransferQueue(c *ui.Context, transfers []Transfer, opts TransferQueueOptions) TransferQueueResult {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if opts.Height <= 0 {
		panic("files: TransferQueue needs a Height; a queue with no height is a window that fills")
	}

	var res TransferQueueResult
	queue := ui.Scroll(c).Height(opts.Height).FillWidth().Gap(u).Role(ui.RoleList)
	queue.Children(func() {
		if len(transfers) == 0 {
			if opts.Empty != nil {
				opts.Empty()
				return
			}
			ui.Text(c, core.Msg(c, "files.noTransfers", core.Def("No transfers"))).
				TextColor(k.TextMuted).FontSize(core.FontSize(c, theme.RowSize))
			return
		}
		for i, tr := range transfers {
			i, tr := i, tr
			card := ui.Column(c).FillWidth().Gap(u*0.5).
				Padding(u, u*1.25).Radius(theme.SmallRadius).Background(k.Surface)
			card.Children(func() {
				ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u).Children(func() {
					ui.Text(c, tr.Name).Grow(1).Ellipsis(tr.Name).
						FontSize(core.FontSize(c, theme.BodySize)).SingleLine()
					progress(c, tr.Done, tr.Total, tr.Failed)
					ui.Box(c).Width(u * 3.5).Shrink(0)
				})
				if opts.WithRate && tr.PerSecond > 0 {
					ui.Text(c, rateWord(tr.PerSecond)).TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
				}
				if tr.Failed {
					ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u).Children(func() {
						why := tr.Error
						if why == "" {
							why = core.Msg(c, "files.failed", core.Def("Failed"))
						}
						_, ink := core.Danger.Pair(k)
						ui.Text(c, why).TextColor(ink).Grow(1).SingleLine().
							FontSize(core.FontSize(c, theme.MetaSize))
					})
				}
				ui.Row(c).FillWidth().Gap(u).Children(func() {
					if tr.Failed {
						if miniButton(c, core.Msg(c, "files.retry", core.Def("Try again"))).Clicked() {
							res.retried = i
						}
					}
					if !tr.Failed && tr.Done < tr.Total {
						if miniButton(c, core.Msg(c, "files.cancel", core.Def("Cancel"))).Clicked() {
							res.cancelled = i
						}
					}
				})
			})
		}
	})
	res.Element = queue
	return res
}

// progress is the bar of one transfer, with its own numbers beside it.
//
// The numbers are printed as well as drawn because a bar of 3% and a bar of
// 30% look almost the same at this size and are very different news. The bar
// is the shape; the figure is the fact.
func progress(c *ui.Context, done, total int64, failed bool) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	tone := core.Accent
	if failed {
		tone = core.Danger
	}
	_, ink := tone.Pair(k)
	pct := Percent(done, total)

	ui.Row(c).AlignItems(ui.Center).Gap(u * 0.75).Children(func() {
		ui.Box(c).Width(u * 14).Height(u * 1.5).Shrink(0).Radius(u * 0.75).
			Background(k.SurfacePressed).Role(ui.RoleNone).
			Label(itoa(pct) + " percent").
			Draw(func(p *ui.Painter, r ui.Rect) {
				fill := float32(pct) / 100
				if fill > 0 {
					p.Fill(ui.Rect{X: r.X, Y: r.Y, W: r.W * fill, H: r.H}, ink, r.H/2)
				}
			})
		ui.Text(c, HumanSize(done)).TextColor(k.TextMuted).
			FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
		ui.Text(c, "/").TextColor(k.TextFaint).
			FontSize(core.FontSize(c, theme.CaptionSize))
		ui.Text(c, HumanSize(total)).TextColor(k.TextMuted).
			FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
	})
}

// rateWord is the rate of a transfer, in the same units as the size. It is
// "/s" rather than "a second" because a queue of six rows has no room for
// the long word and everybody who reads it knows what it means.
func rateWord(perSecond int64) string {
	return HumanSize(perSecond) + core.Def("/s")
}

// miniButton is the small control under a transfer: Cancel, Try again.
func miniButton(c *ui.Context, label string) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	btn := ui.ButtonBase(c).Height(u*5).Radius(theme.PillRadius).
		Padding(0, u*2.25).Background(k.Background).TextColor(k.Text).
		Label(label).Tooltip(label)
	btn.Children(func() {
		ui.Text(c, label).FontSize(core.FontSize(c, theme.RowSize))
	})
	return btn
}

// FileOperationOptions configure a FileOperationProgress.
type FileOperationOptions struct {
	// Title is what is being done: "Copying", "Moving", "Deleting". It is
	// required, because a bar with no verb above it is a bar.
	Title string
	// Detail is the second line: the file being worked on now, the folder,
	// anything that is progress rather than the operation.
	Detail string
	// Done and Total are the counts of items — files, not bytes. A copy of
	// four thousand files is four thousand things and not four thousand
	// kilobytes, and the two count very differently.
	Done, Total int
	// Indeterminate draws a bar with no end, for an operation whose size is
	// not known in advance — which is most of them, until somebody has
	// walked the tree.
	Indeterminate bool
	// Tone is how loud the bar is; zero is the accent.
	Tone core.Severity
}

// FileOperationProgress is one long operation and how far along it is: the
// verb, the file it is on, and the bar.
//
// It is a whole surface rather than a row because these operations are the
// ones a person walks away from, and a progress bar the width of a row is a
// progress bar nobody notices is still happening.
//
// The item count is the caller's and is not this component's to fetch: how
// many files a copy has depends on what is on the disk, and asking here
// would mean walking it on every frame.
func FileOperationProgress(c *ui.Context, opts FileOperationOptions) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if opts.Title == "" {
		panic("files: FileOperationProgress needs a Title; a bar with no verb above it is a bar")
	}
	tone := opts.Tone
	if tone == core.Neutral {
		tone = core.Accent
	}
	_, ink := tone.Pair(k)

	pill := func() *ui.Element {
		return ui.Box(c).FillWidth().Height(u * 2).Shrink(0).Radius(u).
			Background(k.SurfacePressed).Role(ui.RoleNone).
			Label(progressLabel(c, opts)).
			Draw(func(p *ui.Painter, r ui.Rect) {
				if opts.Indeterminate {
					// A moving band rather than a bar with a wrong number on
					// it: an indeterminate operation shown at 30% is a lie
					// somebody will quote back at you.
					band := r.W / 4
					for x := r.X - band; x < r.X+r.W; x += band * 2 {
						x0 := max(x, r.X)
						x1 := min(x+band, r.X+r.W)
						if x1 > x0 {
							p.Fill(ui.Rect{X: x0, Y: r.Y, W: x1 - x0, H: r.H}, ink, r.H/2)
						}
					}
					return
				}
				fill := float32(Percent(int64(opts.Done), int64(opts.Total))) / 100
				if fill > 0 {
					p.Fill(ui.Rect{X: r.X, Y: r.Y, W: r.W * fill, H: r.H}, ink, r.H/2)
				}
			})
	}

	card := ui.Column(c).FillWidth().Gap(u * 0.75).Padding(u * 1.5).
		Radius(theme.CardRadius).Background(k.Surface)
	card.Children(func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u).Children(func() {
			ui.Text(c, opts.Title).Grow(1).Ellipsis(opts.Title).
				FontSize(core.FontSize(c, theme.RowSize)).Bold().SingleLine()
			ui.Text(c, progressLabel(c, opts)).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.MetaSize)).SingleLine()
		})
		pill()
		if opts.Detail != "" {
			ui.Text(c, opts.Detail).TextColor(k.TextMuted).Ellipsis(opts.Detail).
				FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
		}
	})
	return card
}

// progressLabel is what the figure beside the verb says. An operation whose
// size is not known says so in words rather than in a percentage, because
// "0%" on an operation that has not counted anything yet is a claim.
func progressLabel(c *ui.Context, opts FileOperationOptions) string {
	if opts.Indeterminate || opts.Total <= 0 {
		return core.Msg(c, "files.working", core.Def("Working…"))
	}
	return itoa(opts.Done) + core.Def(" / ") + itoa(opts.Total)
}

// StorageUsageOptions configure a StorageUsage.
type StorageUsageOptions struct {
	// Used and Quota are the counts of bytes. A quota of zero means there is
	// no limit, and the figure says so rather than dividing by it.
	Used, Quota int64
	// Plan is the name of the plan, shown beside the figure.
	Plan string
	// Height is the height of the bar; zero gives the library's own.
	Height float32
}

// StorageUsage is how full a storage is: the bar, the figure, and what it
// will be when it is full.
//
// The figure is printed as well as drawn for the same reason a transfer's is:
// a bar at 97% and one at 87% are two shapes a person has seen before, and
// "97% of 20 GB" is the sentence somebody copies into a ticket.
//
// A quota of zero is "no limit" rather than "full", and is said in words. It
// is the one case where the bar has no end to run out of, and drawing a full
// bar for it would say the disk is out of space when it is not.
func StorageUsage(c *ui.Context, opts StorageUsageOptions) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	h := opts.Height
	if h <= 0 {
		h = u * 2.5
	}

	pct := Percent(opts.Used, opts.Quota)
	tone := storageTone(pct)
	_, ink := tone.Pair(k)

	bar := func() *ui.Element {
		return ui.Box(c).FillWidth().Height(h).Shrink(0).Radius(h / 2).
			Background(k.SurfacePressed).Role(ui.RoleNone).
			Label(storageLabel(c, opts)).
			Draw(func(p *ui.Painter, r ui.Rect) {
				if opts.Quota <= 0 {
					return
				}
				fill := float32(pct) / 100
				if fill > 0 {
					p.Fill(ui.Rect{X: r.X, Y: r.Y, W: r.W * fill, H: r.H}, ink, r.H/2)
				}
			})
	}

	col := ui.Column(c).FillWidth().Gap(u * 0.75).Children(func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u).Children(func() {
			ui.Text(c, storageLabel(c, opts)).Grow(1).SingleLine().
				FontSize(core.FontSize(c, theme.RowSize))
			if opts.Plan != "" {
				ui.Text(c, opts.Plan).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
			}
		})
		bar()
	})
	return col
}

// storageTone is how loud a full disk is. Eighty per cent is when people
// start deleting things, so it is the first step up rather than ninety.
func storageTone(pct int) core.Severity {
	switch {
	case pct >= 100:
		return core.Danger
	case pct >= 90:
		return core.Danger
	case pct >= 80:
		return core.Warning
	}
	return core.Accent
}

// storageLabel is the figure: the used bytes and the quota, or the used
// bytes and the words for "there is no limit".
func storageLabel(c *ui.Context, opts StorageUsageOptions) string {
	if opts.Quota <= 0 {
		return HumanSize(opts.Used) + core.Def(" · ") +
			core.Msg(c, "files.noLimit", core.Def("no limit"))
	}
	return HumanSize(opts.Used) + core.Def(" / ") + HumanSize(opts.Quota) +
		core.Def(" (") + itoa(Percent(opts.Used, opts.Quota)) + core.Def("%)")
}
