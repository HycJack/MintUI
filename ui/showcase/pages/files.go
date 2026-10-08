package pages

import (
	"fmt"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/files"
	"github.com/HycJack/MintUI/ui/showcase"
	"github.com/HycJack/MintUI/ui/theme"
)

// Demo state outlives the frame: every demo below hands its component
// a pointer, and a pointer into a frame-local is a click the next
// frame undoes — a tab that will not switch, a dropdown that snaps
// shut, a slider that springs back.
var (
	files_selected = "ui/diffview.go"
	gridSel        = ""
	rename         = "report?.pdf"
	saving         = "report-2026.pdf"
	perm           = files.CanComment
	shareOpen      = true
	invite         = "new@example.invalid"
)

func init() {
	showcase.Register(showcase.Page{
		Package: "files",
		Title:   "ui/files — 文件界面里不随窗口变化的那一半",
		Note:    "浏览、网格、预览、改名、分享、归档、传输、协作；名字、路径、大小、种类全是纯函数先算出来的",
		Width:   1000,
		Height:  4200,
		Want: []string{
			// browse
			"ui", "ui/showcase", "diffview.go", "shared.go", "draft.go",
			"PathBar — 一段一段，每一段都是一下点击",
			// kinds and icons
			"report.pdf", "shot.png", "bundle.zip", "demo.mp4", "core.unknown",
			// grid and preview
			"notes.md", "Cover",
			// rename and share
			"Rename to", "Share report.pdf", "Invite by email or name",
			"Ada Lovelace", "Can comment", "Can edit", "Can view", "Owner",
			// archive and recent
			"release-0.4.tar.gz", "8 entries",
			// progress and storage
			"Copying", "Moving", "Deleting", "Deleting",
			"19 GB / 20 GB (93%)", "12 GB · no limit",
			// collaboration
			"Someone is typing", "Mia Chen", "Ravi Patel", "Lena Ford",
			// pure
			"Join", "KindOf", "IsDir(\"ui/\")", "Ext(\"archive.tar.gz\")",
			"HumanSize", "Percent", "Sanitize", "SanitizeOr(\"\", \"untitled\")",
			"FileIcon", "Viewer",
		},
		Render: func(c *ui.Context) {
			filesPage(c)
		},
	})
}

// filesPage is one folder, looked at five ways: as a tree, as tiles, as a
// preview, as somebody else's window onto the same file, and as the numbers
// that say how big it all is.
//
// Nothing on this page reads a disk. A folder here is a slice of Entry, an
// archive is a slice of ArchiveEntry, a text file is a string, and a transfer
// is a pair of byte counts — which is not only how the library is written, it
// is the only way a page like this can be looked at twice and be the same
// both times.
func filesPage(c *ui.Context) {
	filesBrowseSection(c)
	filesGridSection(c)
	filesShareSection(c)
	filesArchiveSection(c)
	filesProgressSection(c)
	filesCollabSection(c)
	filesPureSection(c)
}

// ── the data ────────────────────────────────────────────────────────────────

// filesEntries is one folder with one folder open inside it: enough for a
// tree to have a level, a twisty and a branch in it.
var filesEntries = []files.Entry{
	{Name: "ui", Path: "ui", Dir: true, Expanded: true, Count: 21,
		Children: []files.Entry{
			{Name: "showcase", Path: "ui/showcase", Dir: true, Count: 12,
				Modified: "14:02"},
			{Name: "core", Path: "ui/core", Dir: true, Count: 6, Modified: "Tue"},
			{Name: "diffview.go", Path: "ui/diffview.go", Size: 14795, Modified: "14:02"},
			{Name: "shared.go", Path: "ui/shared.go", Size: 5508, Modified: "13:40"},
			{Name: "draft.go", Path: "ui/draft.go", Size: 23865, Modified: "Mon"},
		}},
	{Name: "docs", Path: "docs", Dir: true, Count: 4, Modified: "Tue"},
	{Name: "report.pdf", Path: "report.pdf", Size: 284160, Modified: "09:12"},
	{Name: "shot.png", Path: "shot.png", Size: 2411724, Modified: "09:12"},
	{Name: "notes.md", Path: "notes.md", Size: 4210, Modified: "08:55"},
	{Name: "bundle.zip", Path: "bundle.zip", Size: 15728640, Modified: "Fri"},
	{Name: "demo.mp4", Path: "demo.mp4", Size: 94371840, Modified: "Thu"},
}

// filesGrid is a smaller folder, for the tile view: tiles are pictures, and a
// picture wants a folder with room in it.
var filesGrid = []files.Entry{
	{Name: "Cover", Path: "Cover", Kind: files.Image, Size: 482112, Modified: "09:12"},
	{Name: "designs", Path: "designs", Dir: true, Count: 18, Modified: "09:10"},
	{Name: "report.pdf", Path: "report.pdf", Size: 284160, Modified: "09:12"},
	{Name: "notes.md", Path: "notes.md", Size: 4210, Modified: "08:55"},
	{Name: "bundle.zip", Path: "bundle.zip", Size: 15728640, Modified: "Fri"},
	{Name: "demo.mp4", Path: "demo.mp4", Size: 94371840, Modified: "Thu"},
}

var filesArchive = []files.ArchiveEntry{
	{Name: "release-0.4/", Dir: true},
	{Name: "release-0.4/bin/callbacks", Size: 18350080, Compressed: 5242880},
	{Name: "release-0.4/docs/design-system.md", Size: 41210, Compressed: 9216},
	{Name: "release-0.4/ui/", Dir: true},
	{Name: "release-0.4/ui/core/state.go", Size: 6144, Compressed: 2048},
	{Name: "release-0.4/ui/git/diffview.go", Size: 14795, Compressed: 3584},
	{Name: "release-0.4/ui/theme/theme.go", Size: 6935, Compressed: 2048},
	{Name: "release-0.4/notes.txt", Size: 912, Compressed: 720},
}

var filesRecent = []files.RecentFile{
	{Name: "report.pdf", Path: "~/work/report.pdf", Opened: "09:12", Where: "Preview"},
	{Name: "notes.md", Path: "~/work/notes.md", Opened: "08:55", Where: "Editor"},
	{Name: "shot.png", Path: "~/work/shot.png", Opened: "yesterday", Where: "Preview"},
	{Name: "release-0.4.tar.gz", Path: "~/Downloads/release-0.4.tar.gz",
		Opened: "Tuesday", Where: "Archive"},
	{Name: "demo.mp4", Path: "~/Movies/demo.mp4", Opened: "last week", Where: "Player"},
}

var filesPeople = []files.Person{
	{Name: "Ada Lovelace", Email: "ada@example.invalid", Permission: files.IsOwner, You: true},
	{Name: "Mia Chen", Email: "mia@example.invalid", Permission: files.CanEdit},
	{Name: "Ravi Patel", Email: "ravi@example.invalid", Permission: files.CanComment},
	{Name: "Lena Ford", Email: "lena@example.invalid", Permission: files.CanView},
}

var filesTransfers = []files.Transfer{
	{Name: "release-0.4.tar.gz", Path: "~/Downloads", Done: 48234496, Total: 48234496,
		PerSecond: 4194304},
	{Name: "shot.png", Path: "~/work", Done: 1835008, Total: 2411724, PerSecond: 1048576},
	{Name: "bundle.zip", Path: "~/work", Done: 0, Total: 15728640, Failed: true,
		Error: "the source folder moved while it was being read"},
}

var filesViewers = []files.Viewer{
	{Name: "Mia Chen", Doing: "Editing", Here: true},
	{Name: "Ravi Patel", Doing: "Viewing", Here: true},
	{Name: "Lena Ford", Doing: "", Here: false},
	{Name: "Sam Ortiz", Doing: "Reading", Here: false},
}

var filesReactions = []files.Reaction{
	{Emoji: "👍", Count: 3, Mine: true},
	{Emoji: "🎉", Count: 1},
	{Emoji: "👀", Count: 2},
	{Emoji: "🚀", Count: 1},
	{Emoji: "✅", Count: 1},
	{Emoji: "🤔", Count: 1},
}

// filesNoteText is what a preview of a text file shows. It is a constant here
// for the same reason every other dataset on this page is: the gallery must
// not read a disk to draw a picture, and a page that read one would be a
// different page on a different machine.
const filesNoteText = `# design system

Two surfaces, one above the other. A card is Background; a column is
Surface. Get that backwards and the board reads flat.

- spacing is a 4 grid, and u = Density.Unit()
- radii are one family; the large value is a pill
- the outline weight is 1, everywhere
`

// ── browse ─────────────────────────────────────────────────────────────────

func filesBrowseSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "浏览 · FileExplorer / PathBar / FileIcon")

	showcase.Field(c, "FileExplorer — 只画传进来的条目；没打开的文件夹只占一行")
	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 4)).AlignItems(ui.Start).
		Children(func() {
			ui.Column(c).WidthPercent(share(2)).Shrink(0).Children(func() {
				ui.Box(c).Width(unit(c, 116)).Radius(theme.ControlRadius).
					Background(k.Surface).Padding(unit(c, 1)).Children(func() {
					files.FileExplorer(c, &files_selected, filesEntries, files.FileExplorerOptions{
						Height: unit(c, 46), ShowSize: true,
					})
				})
				files.FileExplorer(c, nil, nil, files.FileExplorerOptions{Height: unit(c, 8)})
			})
			ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 1.5)).Children(func() {
				showcase.Field(c, "PathBar — 一段一段，每一段都是一下点击")
				files.PathBar(c, "~/work/callbacks/ui/showcase/pages", files.PathBarOptions{
					Label: "PathBar", Root: "callbacks",
				})
				showcase.Field(c, "MaxSegments — 太深时留下头和尾，中间收成一个省略号")
				files.PathBar(c, "~/work/callbacks/ui/showcase/pages", files.PathBarOptions{
					Label: "PathBar", Root: "callbacks", MaxSegments: 4,
				})
				showcase.Field(c, "FileIcon — 五种 kind，五件库自带的字形")
				ui.Row(c).Gap(unit(c, 2)).Wrap().Children(func() {
					for _, name := range []string{
						"notes.md", "shot.png", "bundle.zip", "demo.mp4", "core.unknown",
					} {
						files.FileIcon(c, name, files.FileIconOptions{Size: unit(c, 5)})
						ui.Text(c, name).TextColor(k.TextMuted).
							FontSize(core.FontSize(c, theme.CaptionSize))
					}
				})
				ui.Text(c, "扩展名没人听说过就是 Binary：拒绝预览是一句话，"+
					"把字节当文本画出来是一屏的替换字符加一张工单。").
					TextColor(k.TextFaint).FontSize(core.FontSize(c, theme.CaptionSize))
			})
		})
}

// ── grid and preview ───────────────────────────────────────────────────────

func filesGridSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "网格与预览 · FileGrid / FilePreview")

	showcase.Field(c, "FileGrid — 同一个文件夹的另一种看法；文件夹排在前面")
	ui.Box(c).Width(unit(c, 236)).Radius(theme.ControlRadius).
		Background(k.Surface).Padding(unit(c, 1.5)).Children(func() {
		files.FileGrid(c, &gridSel, filesGrid, files.FileGridOptions{
			Side: unit(c, 20), ShowSize: true,
		})
	})

	showcase.Field(c, "FilePreview — 文字是调用方的字符串，图是调用方建好的元素")
	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 3)).AlignItems(ui.Start).
		Children(func() {
			for _, p := range []files.FilePreviewOptions{
				{Name: "notes.md", Text: filesNoteText, Size: 4210, Modified: "08:55",
					Width: unit(c, 74), Height: unit(c, 34)},
				{Name: "shot.png", Size: 2411724, Modified: "09:12",
					Width: unit(c, 74), Height: unit(c, 34)},
				{Name: "bundle.zip", Size: 15728640, Modified: "Fri",
					Width: unit(c, 74), Height: unit(c, 34)},
			} {
				p := p
				files.FilePreview(c, p)
			}
		})
}

// ── rename and share ───────────────────────────────────────────────────────

func filesShareSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "改名与分享 · RenameInline / ShareDialog / PermissionSelect")

	showcase.Field(c, "RenameInline — 值是原始输入，提交时才清洗；Esc 放回原名")
	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 4)).AlignItems(ui.Start).
		Children(func() {
			ui.Column(c).Width(unit(c, 116)).Shrink(0).Gap(unit(c, 1)).Children(func() {
				ui.Text(c, "改名前").TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize))
				files.RenameInline(c, &rename, files.RenameInlineOptions{
					Label: "Rename to", Kind: files.Image, Original: "report.pdf",
				})
			})
			ui.Column(c).Width(unit(c, 116)).Shrink(0).Gap(unit(c, 1)).Children(func() {
				ui.Text(c, "Saving — 提交之后到写完之间").TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize))
				files.RenameInline(c, &saving, files.RenameInlineOptions{
					Label: "Rename to", Kind: files.Text, Original: "report.pdf",
					Saving: true,
				})
				ui.Text(c, "字段里画的是 somebody 敲的东西，不是清洗后的名字："+
					"光标不会跑到 somebody 没放的地方去。").TextColor(k.TextFaint).
					FontSize(core.FontSize(c, theme.CaptionSize))
			})
		})

	showcase.Field(c, "PermissionSelect — 四个选项都是按钮，不是一个下拉")
	files.PermissionSelect(c, &perm, files.PermissionSelectOptions{
		Label: "Permission",
		Permissions: []files.Permission{
			files.CanView, files.CanComment, files.CanEdit,
		},
	})

	showcase.Field(c, "ShareDialog — 它不是对话框：没有遮罩，后面那个窗口还活着")
	people := append([]files.Person(nil), filesPeople...)
	files.ShareDialog(c, &shareOpen, files.ShareDialogOptions{
		Title: "Share report.pdf", Subtitle: "~/work/report.pdf",
		People: people, Invite: &invite, Width: unit(c, 236),
		Actions: func() {
			ui.Row(c).FillWidth().Gap(unit(c, 1.5)).AlignItems(ui.Center).
				Children(func() {
					ui.Box(c).Grow(1)
					ui.Box(c).Height(unit(c, 6)).Padding(0, unit(c, 2.5)).
						Radius(theme.PillRadius).Background(k.Surface).Center().
						Children(func() {
							ui.Text(c, "Close").TextColor(k.Text).
								FontSize(core.FontSize(c, theme.CaptionSize))
						})
					ui.Box(c).Height(unit(c, 6)).Padding(0, unit(c, 2.5)).
						Radius(theme.PillRadius).Background(k.Fill).Center().
						Children(func() {
							ui.Text(c, "Done").TextColor(k.OnFill).
								FontSize(core.FontSize(c, theme.CaptionSize))
						})
				})
		},
	})
}

// ── archive and recent ─────────────────────────────────────────────────────

func filesArchiveSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "归档与最近 · ArchiveViewer / RecentFiles")

	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 3)).AlignItems(ui.Start).
		Children(func() {
			ui.Column(c).WidthPercent(58).Shrink(0).Children(func() {
				showcase.Field(c, "ArchiveViewer — 不解压也能回答「里面有什么」")
				ui.Box(c).Width(unit(c, 138)).Radius(theme.ControlRadius).
					Background(k.Surface).Padding(unit(c, 1.5)).Children(func() {
					files.ArchiveViewer(c, filesArchive, files.ArchiveViewerOptions{
						Name: "release-0.4.tar.gz", Height: unit(c, 36), WithRatio: true,
					})
				})
			})
			ui.Column(c).WidthPercent(40).Shrink(0).Children(func() {
				showcase.Field(c, "RecentFiles — 顺序本身就是内容，所以没有表头")
				ui.Box(c).Width(unit(c, 94)).Radius(theme.ControlRadius).
					Background(k.Surface).Padding(unit(c, 1)).Children(func() {
					files.RecentFiles(c, filesRecent, files.RecentFilesOptions{
						Height: unit(c, 34), Max: 5,
					})
				})
				files.RecentFiles(c, nil, files.RecentFilesOptions{Height: unit(c, 8)})
			})
		})
}

// ── progress and storage ───────────────────────────────────────────────────

func filesProgressSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "进度与容量 · FileOperationProgress / StorageUsage / TransferQueue")

	showcase.Field(c, "FileOperationProgress — 计的是文件不是字节；不知道总量时不画百分比")
	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 3)).Children(func() {
		for _, op := range []files.FileOperationOptions{
			{Title: "Copying", Detail: "ui/git/diffview.go", Done: 128, Total: 340},
			{Title: "Moving", Detail: "14.2 GB of 20.0 GB", Done: 61, Total: 100,
				Tone: core.Warning},
			{Title: "Deleting", Detail: "counting first — the size is not known yet",
				Indeterminate: true, Tone: core.Danger},
		} {
			op := op
			ui.Box(c).Grow(1).Children(func() {
				files.FileOperationProgress(c, op)
			})
		}
	})

	showcase.Field(c, "StorageUsage — 数字和条一起画：97% 和 87% 是两种形状")
	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 3)).AlignItems(ui.Start).
		Children(func() {
			for _, s := range []files.StorageUsageOptions{
				{Used: 6400000000, Quota: 20000000000, Plan: "Team"},
				{Used: 18600000000, Quota: 20000000000, Plan: "Team"},
				{Used: 12400000000, Quota: 0, Plan: "Unlimited"},
			} {
				s := s
				ui.Box(c).Grow(1).Children(func() {
					files.StorageUsage(c, s)
				})
			}
		})

	showcase.Field(c, "TransferQueue — 完成的离开，失败的留下并留一个「再试一次」")
	ui.Box(c).Width(unit(c, 236)).Radius(theme.ControlRadius).
		Background(k.Surface).Padding(unit(c, 1)).Children(func() {
		files.TransferQueue(c, filesTransfers, files.TransferQueueOptions{
			Height: unit(c, 46), WithRate: true,
		})
	})
	files.TransferQueue(c, nil, files.TransferQueueOptions{Height: unit(c, 8)})
}

// ── collaboration ──────────────────────────────────────────────────────────

func filesCollabSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "协作 · LiveIndicator / PresenceAvatars / Reactions / RemoteCursor")

	showcase.Field(c, "LiveIndicator — 在线点是这一套里唯一不随主题变的亮色")
	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 6)).AlignItems(ui.Center).
		Children(func() {
			for _, l := range []struct {
				label string
				live  bool
				tone  core.Severity
			}{
				{"Someone is typing", true, core.Neutral},
				{"Connected, nothing happening", false, core.Neutral},
				{"The connection dropped", false, core.Danger},
				{"Synced", true, core.Success},
			} {
				l := l
				ui.Row(c).Gap(unit(c, 1.5)).AlignItems(ui.Center).Children(func() {
					files.LiveIndicator(c, files.LiveIndicatorOptions{
						Label: l.label, Live: l.live, Tone: l.tone,
					})
					ui.Text(c, l.label).TextColor(k.Text).
						FontSize(core.FontSize(c, theme.CaptionSize))
				})
			}
		})

	showcase.Field(c, "PresenceAvatars — 在文件上的是实心的，来过但不在的是空心的")
	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 6)).AlignItems(ui.Start).
		Children(func() {
			ui.Box(c).Width(unit(c, 116)).Radius(theme.ControlRadius).
				Background(k.Surface).Padding(unit(c, 1.5)).Children(func() {
				ui.Column(c).FillWidth().Gap(unit(c, 1.5)).Children(func() {
					files.PresenceAvatars(c, filesViewers, files.PresenceAvatarsOptions{
						Max: 3, Total: 5,
					})
					ui.Text(c, "Max 3 / Total 5").TextColor(k.TextFaint).
						FontSize(core.FontSize(c, theme.CaptionSize))
					files.PresenceAvatars(c, filesViewers, files.PresenceAvatarsOptions{
						WithNames: true,
					})
				})
			})
			ui.Column(c).Width(unit(c, 116)).Shrink(0).Gap(unit(c, 1.5)).Children(func() {
				showcase.Field(c, "Reactions — 超过 Max 的收成一个 +n，不消失")
				files.Reactions(c, filesReactions, files.ReactionsOptions{
					Label: "Reactions on report.pdf", Max: 3,
				})
				files.Reactions(c, filesReactions[:2], files.ReactionsOptions{
					Label: "Reactions on report.pdf", Max: 5,
				})
				showcase.Field(c, "RemoteCursor — 颜色来自名字，所以同一个人在哪儿都是同一种")
				// The mark is absolute, so it needs a box with a size of its own
				// to be absolute inside: without one it is placed against the
				// page and lands wherever.
				ui.Box(c).Width(unit(c, 116)).Height(unit(c, 22)).Shrink(0).
					Radius(theme.ControlRadius).Background(k.Surface).
					Border(theme.BorderWidth, k.Border).Clip().Children(func() {
					files.RemoteCursor(c, files.RemoteCursorOptions{
						Name: "Mia Chen", Doing: "Editing",
						X: unit(c, 16), Y: unit(c, 3),
					})
				})
			})
		})
}

// ── the pure half ──────────────────────────────────────────────────────────

// filesPureSection is the part of this package that decides rather than draws.
//
// It is the same shape as the git page's: every one of these is a question
// about a string or a number, answered before a single element is made, and
// every one of them is a thing a file interface gets quietly wrong — a name
// that kept its separator writes a file somewhere else, a path joined badly
// is two paths, a size rounded the wrong way disagrees with the web version.
func filesPureSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "纯函数 — 先算出答案，再画")

	showcase.Field(c, "路径 — Join 收尾、Base 取名、Ext 取扩展名、IsDir 看分隔符")
	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 3)).AlignItems(ui.Start).
		Children(func() {
			ui.Column(c).WidthPercent(share(2)).Shrink(0).Children(func() {
				cols := fileCols(72, 40)
				filePureRow(c, cols, "Join", "a/b")
				filePureRow(c, cols, "Join(\"ui\", \"showcase\", \"\")",
					files.Join("ui", "showcase", ""))
				filePureRow(c, cols, "Join(\"ui/showcase\", \"../..\")",
					files.Join("ui/showcase", "../.."))
				filePureRow(c, cols, "Join(\"ui\", \"/etc/hosts\")",
					files.Join("ui", "/etc/hosts"))
				filePureRow(c, cols, "Join(`C:\\work`, \"ui\")",
					files.Join(`C:\work`, "ui"))
				filePureRow(c, cols, "Join(\"\", \"\")", files.Join("", ""))
				filePureRow(c, cols, "Base(\"C:\\\\work\\\\ui\\\\main.go\")",
					files.Base(`C:\work\ui\main.go`))
				filePureRow(c, cols, "Ext(\"archive.tar.gz\")", files.Ext("archive.tar.gz"))
				filePureRow(c, cols, "Ext(\".gitignore\")", fmt.Sprintf("%q", files.Ext(".gitignore")))
				filePureRow(c, cols, "IsDir(\"ui/\")", filesYes(files.IsDir("ui/")))
				filePureRow(c, cols, "IsDir(\"main.go\")", filesYes(files.IsDir("main.go")))
			})
			ui.Column(c).WidthPercent(share(2)).Shrink(0).Children(func() {
				cols := fileCols(52, 30, 44)
				filePureRow(c, cols, "KindOf", "kind", "画的是哪一件")
				for _, name := range []string{
					"notes.md", "shot.png", "bundle.zip", "demo.mp4", "core.unknown",
				} {
					k := files.KindOf(name)
					filePureRow(c, cols, name, k.String(), filesIconFor(k))
				}
			})
		})

	showcase.Field(c, "名字 — Sanitize 六步，SananitizeOr 兜底")
	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 3)).AlignItems(ui.Start).
		Children(func() {
			ui.Column(c).WidthPercent(share(2)).Shrink(0).Children(func() {
				ui.Text(c, "Sanitize").TextColor(k.Text).
					FontSize(core.FontSize(c, theme.RowSize)).Bold()
				cols := fileCols(84, 40)
				filePureRow(c, cols, "敲进去的", "Sanitize 之后")
				for _, in := range []string{
					"report?.pdf", "  report.pdf  ", "/repo/ui/", "notes\x00.md",
					"NUL.txt", "a/b/c.txt", "",
				} {
					filePureRow(c, cols, fmt.Sprintf("%q", in), fmt.Sprintf("%q", files.Sanitize(in)))
				}
				filePureRow(c, cols, "SanitizeOr(\"\", \"untitled\")",
					files.SanitizeOr("", "untitled"))
				filePureRow(c, cols, "SanitizeOr(\"\", \"\")", files.SanitizeOr("", ""))
			})
			ui.Column(c).WidthPercent(share(2)).Shrink(0).Children(func() {
				cols := fileCols(46, 26, 30, 22)
				filePureRow(c, cols, "HumanSize", "Percent", "Percent", "")
				for _, n := range []int64{
					0, 999, 1000, 1500, 999999, 1000000, 284160, 2411724,
					94371840, 12400000000,
				} {
					filePureRow(c, cols, files.HumanSize(n),
						fmt.Sprintf("%d%%", files.Percent(n, 20000000000)),
						fmt.Sprintf("%d%%", files.Percent(n, 1024)), "")
				}
				filePureRow(c, cols, "HumanSize(-1500)", files.HumanSize(-1500), "", "", "")
				filePureRow(c, cols, "Percent(1, 0)", fmt.Sprintf("%d%%", files.Percent(1, 0)), "", "", "")
			})
		})

	showcase.Field(c, "Kind — 五种，扩展名没人听说过就是 Binary")
	ui.Text(c, "FileIcon").TextColor(k.Text).
		FontSize(core.FontSize(c, theme.RowSize)).Bold()
	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 3)).AlignItems(ui.Start).
		Children(func() {
			for _, kd := range []files.Kind{
				files.Binary, files.Text, files.Image, files.Archive, files.Media,
			} {
				kd := kd
				ui.Column(c).AlignItems(ui.Center).Gap(unit(c, 0.75)).Children(func() {
					files.FileIcon(c, "report.pdf", files.FileIconOptions{Kind: kd})
					ui.Text(c, kd.String()).TextColor(k.Text).
						FontSize(core.FontSize(c, theme.CaptionSize)).Bold()
					ui.Text(c, fmt.Sprintf("CanPreview %v", kd.CanPreview())).
						TextColor(k.TextFaint).FontSize(core.FontSize(c, theme.CaptionSize))
				})
			}
			ui.Text(c, "FileIcon 从不自己画：每个系统都给自己的文件类型画了一个小图，"+
				"一个把那些图和库这一套混在一起的窗口，看起来就是两个应用。").TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize))
		})

	showcase.Field(c, "Viewer 与 Person — 两份名单，回答两个问题")
	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 3)).AlignItems(ui.Start).
		Children(func() {
			ui.Column(c).WidthPercent(share(2)).Shrink(0).Children(func() {
				ui.Text(c, "Viewer").TextColor(k.Text).
					FontSize(core.FontSize(c, theme.RowSize)).Bold()
				cols := fileCols(40, 30, 30)
				filePureRow(c, cols, "Viewer.Name", "Doing", "Here")
				for _, v := range filesViewers {
					filePureRow(c, cols, v.Name, filesOr(v.Doing, "—"), filesYes(v.Here))
				}
			})
			ui.Column(c).WidthPercent(share(2)).Shrink(0).Children(func() {
				cols := fileCols(40, 60, 24)
				filePureRow(c, cols, "Person.Name", "Email", "Permission")
				for _, p := range filesPeople {
					filePureRow(c, cols, p.Name, p.Email, fmt.Sprintf("%d", int(p.Permission)))
				}
			})
		})

	ui.Column(c).FillWidth().Gap(unit(c, 1)).Children(func() {
		for _, line := range []string{
			"组件不读磁盘：文件夹、归档条目、文本、图片、传输都是参数。",
			"名字、路径、大小、种类四个纯函数是这个包存在的理由，测试断言的是返回值。",
			"ShareDialog 不是对话框：没有遮罩，它是一块调用方放的面板。",
			"给不了的东西 panic(\"files: …\")，不是画一个空盒子。",
		} {
			ui.Text(c, line).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize))
		}
	})
}

// fileCols is the column widths of one of the tables above, in spacing units.
func fileCols(w ...float32) []float32 { return w }

// filePureRow is one line of a pure function's output.
func filePureRow(c *ui.Context, cols []float32, cells ...string) {
	k := core.Tokens(c)
	const gap = 1.5
	ui.Row(c).Gap(unit(c, gap)).Children(func() {
		for i, cell := range cells {
			if i >= len(cols) {
				break
			}
			ui.Text(c, cell).TextColor(k.TextMuted).Width(unit(c, cols[i])).
				FontSize(core.FontSize(c, theme.CaptionSize)).MaxLines(1)
		}
	})
}

// filesIconFor is which of the library's glyphs a kind wears, said in the
// glyph's own name: the mapping is by shape, not by file type, because the
// set has no picture of a photograph in it.
func filesIconFor(k files.Kind) string {
	switch k {
	case files.Text:
		return "jobs"
	case files.Image:
		return "overview"
	case files.Archive, files.Binary:
		return "inventory"
	case files.Media:
		return "panel"
	}
	return "inventory"
}

// filesOr keeps a column of values aligned.
func filesOr(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

// filesYes is a bool as a word a table can read.
func filesYes(b bool) string {
	if b {
		return "是"
	}
	return "否"
}
