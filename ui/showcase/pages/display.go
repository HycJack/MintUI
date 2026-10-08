package pages

import (
	"image"
	"image/color"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/showcase"
	"github.com/HycJack/MintUI/ui/theme"
)

func init() {
	showcase.Register(showcase.Page{
		Package: "display",
		Title:   "ui/display — 不用文字就读得懂的记号",
		Note:    "头像、药丸、角标、图标、图片、标题、正文、快捷键、链接、标签",
		Width:   1000,
		Height:  1140,
		Want: []string{
			"AT",
			"MC",
			"Andre Thomson",
			"9 people",
			"+3",
			"Low",
			"Medium",
			"High",
			"Critical",
			"3 条未读",
			"27 条待办",
			"99+",
			"128 条通知",
			"overview",
			"callbacks",
			"chevron-up",
			"link-broken",
			"已加载 · 裁剪填满",
			"占位",
			"失败",
			"空槽位",
			"Level 1",
			"Level 3",
			"Riverside Clinic — 正文字号，15.5，粗体是客户名",
			"命令面板",
			"打开工单",
			"Riverside",
		},
		Render: func(c *ui.Context) {
			displayPage(c)
		},
	})
}

// displayPage is every mark the library draws that a person reads without a
// sentence, plus the few text components that are here so the type has one
// home. All of it is read-only: nothing on this page takes a press.
func displayPage(c *ui.Context) {
	avatarSection(c)
	badgeSection(c)
	iconSection(c)
	imageSection(c)
	textSection(c)
	inlineSection(c)
}

// avatarSection shows the three ways a person is summarised: one face, a
// row of faces with the rest as a figure, and a group whose rest can be
// opened.
func avatarSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "Avatar — 名字还没换成照片时的替身")
	ui.Row(c).FillWidth().Gap(unit(c, 3)).Children(func() {
		ui.Column(c).Grow(1).Gap(unit(c, 1)).Children(func() {
			ui.Row(c).Gap(unit(c, 2)).Children(func() {
				display.Avatar(c, "Andre Thomson")
				display.Avatar(c, "Mia Chen")
				display.Avatar(c, "Ravi Patel")
				display.Avatar(c, "")
			})
			ui.Text(c, "空名只画圆，不画缩写").TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize))
		})
		ui.Column(c).Grow(1).Gap(unit(c, 1)).Children(func() {
			display.AvatarCluster(c,
				[]string{"Andre Thomson", "Mia Chen", "Ravi Patel", "Lena Ford", "Sam Ortiz"}, 9)
			ui.Text(c, "AvatarCluster — 静态，前 5 个 + 一个人数").TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize))
		})
		ui.Column(c).Grow(1).Gap(unit(c, 1)).Children(func() {
			display.AvatarGroup(c,
				[]string{"Andre Thomson", "Mia Chen", "Ravi Patel", "Lena Ford", "Sam Ortiz", "Nina Roy", "Omar Aziz"},
				display.AvatarGroupOptions{More: "另外 3 人"})
			ui.Text(c, "AvatarGroup — 「+n」可点，是一扇门").TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize))
		})
	})

	showcase.Field(c, "Meter — 优先级与信号格，level 会被 clamp")
	ui.Row(c).FillWidth().Gap(unit(c, 2)).Children(func() {
		priority := func(name string, sev core.Severity, level int) {
			bg, fg := sev.Pair(k)
			ui.Box(c).Grow(1).Padding(unit(c, 1.25), unit(c, 2)).
				Radius(theme.PillRadius).Background(bg).Center().Children(func() {
				ui.Row(c).AlignItems(ui.Center).Gap(unit(c, 1.5)).Children(func() {
					display.Meter(c, 4, level, sev)
					ui.Text(c, name).TextColor(fg).
						FontSize(core.FontSize(c, theme.CaptionSize)).Bold()
				})
			})
		}
		priority("Low", core.Neutral, 1)
		priority("Medium", core.Neutral, 2)
		priority("High", core.Warning, 3)
		priority("Critical", core.Danger, 4)
	})
	ui.Row(c).Gap(unit(c, 2)).Children(func() {
		display.Meter(c, 4, 0, core.Neutral)
		display.Meter(c, 4, 9, core.Danger)
		ui.Text(c, "0 与超出 n 的 level").TextColor(k.TextFaint).
			FontSize(core.FontSize(c, theme.CaptionSize))
	})
}

// badgeSection shows a count stuck to something, at every tone, plus the dot
// that means "something needs looking at" without saying how much.
func badgeSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "Badge 与 PresenceDot — 角标与状态点")
	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Center).Children(func() {
		// The case Badge is for: a count stuck to a thing. The caller
		// parents it and anchors it, which is what Attach is for.
		ui.Row(c).FillWidth().Wrap().AlignItems(ui.Center).Children(func() {
			for _, b := range []struct {
				count int
				opts  display.BadgeOptions
			}{
				{3, display.BadgeOptions{Tone: core.Danger, Name: "3 条未读"}},
				{27, display.BadgeOptions{Tone: core.Warning, Max: 99, Name: "27 条待办"}},
				{0, display.BadgeOptions{Name: "没有未读"}},
				{128, display.BadgeOptions{Tone: core.Accent, Max: 99, Solid: true, Name: "128 条通知"}},
			} {
				b := b
				// The badge is made inside the box it is stuck to, because
				// an element belongs to whatever container was current
				// where it was created — and an attached child anchors to
				// its own parent, not to a sibling.
				ui.Box(c).Size(unit(c, 13), unit(c, 13)).Radius(unit(c, 6.5)).
					// The right margin is the badge's room: an attached
					// child draws outside its box.
					Margin(0, unit(c, 5), 0, unit(c, 1)).
					Background(k.Surface).Border(theme.BorderWidth, k.Border).
					Center().Children(func() {
					ui.Text(c, "R").TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.MonoSize)).Bold()
					display.Badge(c, b.count, b.opts).
						Attach(ui.AnchorTopRight, ui.AnchorCenter).Right(-unit(c, 4))
				})
			}
		})
	})
	ui.Row(c).FillWidth().Gap(unit(c, 2)).Children(func() {
		dot := func(name string, sev core.Severity) {
			ui.Box(c).Grow(1).Height(unit(c, 10)).Radius(theme.ControlRadius).
				Background(k.Surface).Center().Children(func() {
				display.PresenceDot(c, display.BadgeOptions{Tone: sev, Name: name})
			})
		}
		dot("有待办", core.Warning)
		dot("有故障", core.Danger)
		dot("已同步", core.Success)
		dot("普通标记", core.Neutral)
	})
}

// iconSection shows the whole built-in set. There are thirty-eight of them
// and they are the vocabulary of a callbacks board, so the page draws all of
// them rather than picking a favourite four.
func iconSection(c *ui.Context) {
	showcase.Section(c, "Icon — 库自带的整套字形，名字不认识就 panic")
	icons := []struct {
		name  display.IconName
		tone  core.Severity
		muted bool
	}{
		{display.IconOverview, core.Neutral, false},
		{display.IconCallbacks, core.Neutral, false},
		{display.IconCustomers, core.Neutral, false},
		{display.IconTeam, core.Neutral, false},
		{display.IconInventory, core.Neutral, false},
		{display.IconJobs, core.Neutral, false},
		{display.IconReports, core.Neutral, false},
		{display.IconBell, core.Neutral, false},
		{display.IconSettings, core.Neutral, false},
		{display.IconProfile, core.Neutral, false},
		{display.IconPanel, core.Neutral, false},
		{display.IconSliders, core.Neutral, false},
		{display.IconPlus, core.Neutral, false},
		{display.IconSearch, core.Neutral, false},
		{display.IconFilter, core.Neutral, false},
		{display.IconCalendar, core.Neutral, false},
		{display.IconClock, core.Neutral, false},
		{display.IconCheck, core.Neutral, false},
		{display.IconDismiss, core.Neutral, false},
		{display.IconChevronUp, core.Neutral, false},
		{display.IconChevron, core.Neutral, false},
		{display.IconDownload, core.Neutral, false},
		{display.IconRefresh, core.Neutral, false},
		{display.IconTrash, core.Neutral, false},
		{display.IconEdit, core.Neutral, false},
		{display.IconCopy, core.Neutral, false},
		{display.IconWarning, core.Warning, false},
		{display.IconStar, core.Neutral, false},
		{display.IconExternal, core.Neutral, false},
		{display.IconCollapse, core.Neutral, false},
		{display.IconExpand, core.Neutral, false},
		{display.IconPhone, core.Neutral, false},
		{display.IconMapPin, core.Neutral, false},
		{display.IconPaperclip, core.Neutral, false},
		{display.IconLinkBroken, core.Danger, false},
		{display.IconDismiss, core.Neutral, true},
		{display.IconCheck, core.Success, false},
	}
	ui.Row(c).FillWidth().Wrap().Gap(unit(c, 2)).Children(func() {
		for _, ic := range icons {
			ic := ic
			ui.Column(c).Width(unit(c, 13)).AlignItems(ui.Center).Gap(unit(c, 0.75)).
				Children(func() {
					display.Icon(c, ic.name, display.IconOptions{
						Name: string(ic.name), Tone: ic.tone, Muted: ic.muted,
					})
					ui.Text(c, string(ic.name)).TextColor(core.Tokens(c).TextFaint).
						FontSize(core.FontSize(c, 9)).SingleLine()
				})
		}
	})
}

// imageSection shows the three states a picture slot can be in — loaded,
// waiting, and failed — because they are drawn differently on purpose.
func imageSection(c *ui.Context) {
	showcase.Section(c, "Image — 两种「没有图」画得不一样")
	ui.Row(c).FillWidth().Gap(unit(c, 3)).Children(func() {
		// A bitmap made here rather than loaded: the gallery must not read
		// a disk or a network to draw a picture, and a flat swatch is
		// enough to show cover, radius and the fixed box.
		photo := sampleBitmap()
		cell := func(src *ui.Bitmap, opts display.ImageOptions) {
			ui.Column(c).AlignItems(ui.Center).Gap(unit(c, 1)).Children(func() {
				display.Image(c, src, opts)
				ui.Text(c, opts.Name).TextColor(core.Tokens(c).TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize))
			})
		}
		cell(photo, display.ImageOptions{
			Width: unit(c, 30), Cover: true, Radius: theme.ControlRadius,
			Name: "已加载 · 裁剪填满",
		})
		cell(photo, display.ImageOptions{
			Width: unit(c, 30), Ratio: 1.6, Name: "已加载 · 按比例",
		})
		cell(photo, display.ImageOptions{
			Width: unit(c, 30), Cover: true, Name: "已加载 · 方角",
		})
		cell(nil, display.ImageOptions{
			Width: unit(c, 30), Ratio: 1.6, Placeholder: "等照片", Name: "占位",
		})
		cell(nil, display.ImageOptions{
			Width: unit(c, 30), Ratio: 1.6, Failed: "加载失败", Name: "失败",
		})
		cell(nil, display.ImageOptions{
			Width: unit(c, 30), Ratio: 1.6, Name: "空槽位",
		})
	})
}

// sampleBitmap is a deterministic picture: three flat bands and a mark, so
// the gallery's own images do not depend on anything on this machine.
func sampleBitmap() *ui.Bitmap {
	const w, h = 96, 60
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			switch {
			case y < h/3:
				img.Set(x, y, color.RGBA{R: 0xd9, G: 0xf2, B: 0x4b, A: 0xff})
			case y < 2*h/3:
				img.Set(x, y, color.RGBA{R: 0x18, G: 0x1a, B: 0x1f, A: 0xff})
			default:
				img.Set(x, y, color.RGBA{R: 0xe8, G: 0xef, B: 0xfd, A: 0xff})
			}
		}
	}
	return ui.NewBitmap(img)
}

// textSection shows the three levels of a titled block and the body copy
// that goes under it.
func textSection(c *ui.Context) {
	showcase.Section(c, "Heading 与 Text — 字号按角色给，不按数字给")
	ui.Row(c).FillWidth().Gap(unit(c, 4)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).Grow(1).Gap(unit(c, 3)).Children(func() {
			display.Heading(c, "Level 1", display.HeadingOptions{
				Level: 1, Subtitle: "页面自己的大标题", Divider: true,
			})
			display.Heading(c, "Level 2", display.HeadingOptions{
				Level: 2, Subtitle: "面板的标题",
			})
		})
		ui.Column(c).Grow(1).Gap(unit(c, 3)).Children(func() {
			display.Heading(c, "Level 3", display.HeadingOptions{
				Level: 3, Subtitle: "区块的标题",
			})
			display.Heading(c, "Level 9", display.HeadingOptions{
				Level: 9, Subtitle: "超过 3 就是 3",
			})
		})
	})
	ui.Column(c).FillWidth().Gap(unit(c, 0.5)).Children(func() {
		display.Text(c, "Riverside Clinic — 正文字号，15.5，粗体是客户名", display.TextOptions{Bold: true})
		display.Text(c, "CB-2871 · AC repair — 副行 13 点，次级墨色", display.TextOptions{Muted: true})
		display.Text(c, "Andre Thomson · 2026-10-07 — 记号，不承载阅读", display.TextOptions{Faint: true})
		display.Text(c, "/var/lib/callbacks/CB-2871.json — 等宽", display.TextOptions{Mono: true})
		display.Text(c, "很长的一行会被截断并加省略号，因为它总是待在卡片标题那种一行里，"+
			"折行会把一整列的信息密度压塌。", display.TextOptions{MaxLines: 1})
	})
}

// inlineSection shows the small marks that sit inside a sentence: a keyboard
// shortcut, a link, and a tag that can be taken off.
func inlineSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "Kbd、Link、Tag — 句子里的小零件")
	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).Grow(1).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "Kbd — 快捷键的每个键各占一个盒子")
			ui.Row(c).Gap(unit(c, 2)).AlignItems(ui.Center).Children(func() {
				display.Kbd(c, "⌘", "K")
				ui.Text(c, "命令面板").TextColor(k.Text).FontSize(core.FontSize(c, theme.BodySize))
			})
			ui.Row(c).Gap(unit(c, 2)).AlignItems(ui.Center).Children(func() {
				display.Kbd(c, "⇧", "⌘", "⏎")
				ui.Text(c, "保存并关闭").TextColor(k.Text).FontSize(core.FontSize(c, theme.BodySize))
			})
		})
		ui.Column(c).Grow(1).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "Link — 永远带下划线，只靠颜色是不行的")
			ui.Column(c).Gap(unit(c, 1)).Children(func() {
				display.Link(c, "打开工单", display.LinkOptions{URL: "https://example.invalid/cb-2871"})
				display.Link(c, "行文里的链接", display.LinkOptions{Muted: true})
			})
		})
		ui.Column(c).Grow(1).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "Tag — 可以拿掉的一个标签")
			tags := &[]string{"AC", "Riverside"}
			ui.Row(c).Gap(unit(c, 1.5)).Wrap().Children(func() {
				for _, t := range *tags {
					display.Tag(c, t, display.TagOptions{Closable: true})
				}
				*tags = (*tags)[:len(*tags)-1]
			})
		})
	})
}
