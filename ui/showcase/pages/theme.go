package pages

import (
	"fmt"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/showcase"
	"github.com/HycJack/MintUI/ui/theme"
)

func init() {
	showcase.Register(showcase.Page{
		Package: "theme",
		Title:   "ui/theme — 看起来是什么样",
		Note:    "浅色与深色两套 Tokens：每个颜色 token 的两个值、圆角、描边、间距、字号、密度、外壳尺寸",
		Width:   1000,
		Height:  3820,
		Want: []string{
			"Background", "Surface", "SurfaceHover", "SurfacePressed", "Border",
			"Text", "TextMuted", "TextFaint", "Fill", "OnFill", "Accent",
			"DangerBg", "Danger", "WarningBg", "Warning",
			"SuccessBg", "Success", "AccentBg", "AccentText", "Lively",
			"IsDark", "Light", "Dark",
			"SmallRadius 8", "ControlRadius 14", "CardRadius 18",
			"PanelRadius 24", "PillRadius 999",
			"BorderWidth 1",
			"SidebarWidth 322", "RailWidth 76", "ColumnWidth 272",
			"Compact", "Comfortable", "u = 4", "u = 6",
		},
		Render: func(c *ui.Context) {
			themePage(c)
		},
	})
}

// themePage is the design system drawn as itself.
//
// It is the one page in the gallery whose subject is a pair of values rather
// than a component, so it is built as two sections that are the same drawing
// run twice: once against theme.Light() and once against theme.Dark(). The
// reader compares the two sections side by side and the answer to "what is
// different between the two appearances" is on the page rather than in a
// sentence.
//
// The two sections are not merely similar, they are identical in code, which
// is the only way the page can be trusted to still be true after somebody
// adds a token to one palette and forgets the other: a token added to both
// lists shows up twice, and a token added to one shows up in one column of
// the other's table as a gap.
func themePage(c *ui.Context) {
	light := theme.Light()
	dark := theme.Dark()

	showcase.Section(c, "浅色 · theme.Light()")
	themePalette(c, light, dark, "Light", "深色", true)

	showcase.Section(c, "深色 · theme.Dark()")
	themePalette(c, dark, light, "Dark", "浅色", false)
}

// themeToken is one colour role, read out of a palette by a function rather
// than by a copied hex string.
//
// The function is the whole reason this page is worth drawing: a table with
// the hexes typed into it is a second copy of the palette that goes stale
// the day a value is changed, and a page that shows a stale colour is worse
// than no page. Reading the field means the number on screen is the number in
// the code, in both directions.
type themeToken struct {
	name string
	use  string
	// read is the token's field. It takes a palette so the same entry can be
	// asked for the light value and the dark value in one row.
	read func(theme.Tokens) ui.Color
}

// themeTokens is every colour in a palette, in the order a reader meets them:
// surfaces, then ink, then the status ramp, then the two that are not part of
// any ramp.
func themeTokens() []themeToken {
	return []themeToken{
		{"Background", "窗口、卡片、抽屉、主区", func(t theme.Tokens) ui.Color { return t.Background }},
		{"Surface", "侧栏、看板列、图标块、计数药丸", func(t theme.Tokens) ui.Color { return t.Surface }},
		{"SurfaceHover", "悬停态", func(t theme.Tokens) ui.Color { return t.SurfaceHover }},
		{"SurfacePressed", "按下态、进度槽", func(t theme.Tokens) ui.Color { return t.SurfacePressed }},
		{"Border", "描边、分组线、选中行的药丸", func(t theme.Tokens) ui.Color { return t.Border }},
		{"Text", "标题、客户名、金额", func(t theme.Tokens) ui.Color { return t.Text }},
		{"TextMuted", "工单号、技师名、日期", func(t theme.Tokens) ui.Color { return t.TextMuted }},
		{"TextFaint", "分隔符等记号，不承载阅读", func(t theme.Tokens) ui.Color { return t.TextFaint }},
		{"Fill", "主按钮、Toast、图标栏激活态", func(t theme.Tokens) ui.Color { return t.Fill }},
		{"OnFill", "上面那些元素上的字", func(t theme.Tokens) ui.Color { return t.OnFill }},
		{"Accent", "系统强调色，只在控件需要时", func(t theme.Tokens) ui.Color { return t.Accent }},
		{"DangerBg", "危险药丸的底", func(t theme.Tokens) ui.Color { return t.DangerBg }},
		{"Danger", "危险药丸的字", func(t theme.Tokens) ui.Color { return t.Danger }},
		{"WarningBg", "告警药丸的底", func(t theme.Tokens) ui.Color { return t.WarningBg }},
		{"Warning", "告警药丸的字", func(t theme.Tokens) ui.Color { return t.Warning }},
		{"SuccessBg", "成功药丸的底", func(t theme.Tokens) ui.Color { return t.SuccessBg }},
		{"Success", "成功药丸的字", func(t theme.Tokens) ui.Color { return t.Success }},
		{"AccentBg", "强调药丸的底", func(t theme.Tokens) ui.Color { return t.AccentBg }},
		{"AccentText", "强调药丸的字", func(t theme.Tokens) ui.Color { return t.AccentText }},
		{"Lively", "唯一不随主题变的亮色：在线点、空态加号", func(t theme.Tokens) ui.Color { return t.Lively }},
	}
}

// hex is a colour as the string it was written as in theme.go. Format rather
// than a hand-written table for the same reason the entries are functions:
// the page prints what the code holds.
func themeHex(c ui.Color) string {
	return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
}

// themePalette draws one palette: the table of every token with both values,
// then the same values again as the things they are for.
//
// here is the palette the section is about and other is the one it is being
// compared with; both go into every row of the table so the reader never has
// to scroll back to the other section to find the other value.
func themePalette(c *ui.Context, here, other theme.Tokens, which, otherName string, isLight bool) {
	showcase.Field(c, "Tokens.IsDark() · 这一套自己说自己是哪一套："+
		fmt.Sprintf("%v —— 两栏的十六进制值是从 Tokens 里读出来的，不是抄的", here.IsDark()))
	themeTokenTable(c, here, other, which, otherName)

	showcase.Field(c, "表面与墨色 · 按这一套画一遍（左边灰底是 Surface，右边白底是 Background）")
	themeSurfaceSpecimen(c, here)

	showcase.Field(c, "Severity.Pair · 成对的底 + 字，业务代码不自己挑颜色")
	themeSeveritySpecimen(c, here)

	showcase.Field(c, "Fill / OnFill 与 Lively · 深色填充面")
	themeFillSpecimen(c, here)

	themeScalesSection(c, here, isLight)

	showcase.Field(c, "Light() 与 Dark() · 两个常量，字段一一对应")
	themeTokensPair(c, here, other, isLight)
}

// themeTokenTable is the table proper: a swatch, the token's name, the value this
// section is about, the value the other one has, and what it is for.
//
// The row is given an explicit width and the columns are given widths that
// add up to less than it, because a row that fills the page and then grows
// its last column pushes that column off the right edge — and the one thing
// this table cannot afford to lose is the end of the "used for" text.
func themeTokenTable(c *ui.Context, here, other theme.Tokens, which, otherName string) {
	// Every width below is stated rather than grown, and the five of them add
	// up to the content width rather than to less than it: the "used for"
	// column is the one that wraps when it is too narrow, and a wrapped tail
	// of Chinese text on the right edge of a table is the first thing a
	// reader sees.
	const (
		swatch = 4   // u
		name   = 28  // u
		hexCol = 24  // u
		use    = 150 // u
		gap    = 1.5 // u
		// 16 + 112 + 96 + 96 + 600 + 4 * 6 = 944
		row = 236 // u
	)
	ui.Column(c).FillWidth().Gap(unit(c, 1)).Children(func() {
		ui.Row(c).Width(unit(c, row)).Gap(unit(c, gap)).AlignItems(ui.Center).
			Children(func() {
				ui.Box(c).Width(unit(c, swatch))
				ui.Text(c, "Token").TextColor(core.Tokens(c).TextMuted).
					Width(unit(c, name)).FontSize(core.FontSize(c, theme.CaptionSize)).Bold()
				ui.Text(c, which).TextColor(core.Tokens(c).TextMuted).
					Width(unit(c, hexCol)).FontSize(core.FontSize(c, theme.CaptionSize)).Bold()
				ui.Text(c, otherName).TextColor(core.Tokens(c).TextMuted).
					Width(unit(c, hexCol)).FontSize(core.FontSize(c, theme.CaptionSize)).Bold()
				ui.Text(c, "用在哪").TextColor(core.Tokens(c).TextMuted).
					Width(unit(c, use)).FontSize(core.FontSize(c, theme.CaptionSize)).Bold()
			})
		for _, t := range themeTokens() {
			mine, theirs := t.read(here), t.read(other)
			ui.Row(c).Width(unit(c, row)).Gap(unit(c, gap)).AlignItems(ui.Center).
				Children(func() {
					// The swatch carries a border of the page's own ink so that
					// a token whose value is close to the page's background —
					// Dark().Text is #f4f4f5 on a white page — is still a
					// square rather than a hole.
					ui.Box(c).Size(unit(c, swatch), unit(c, swatch)).Shrink(0).
						Radius(theme.SmallRadius).Background(mine).
						Border(theme.BorderWidth, core.Tokens(c).Border).
						Label(t.name + " " + themeHex(mine))
					ui.Text(c, t.name).TextColor(core.Tokens(c).Text).
						Width(unit(c, name)).FontSize(core.FontSize(c, theme.CaptionSize))
					themeSwatchHex(c, mine, hexCol, true)
					themeSwatchHex(c, theirs, hexCol, false)
					ui.Text(c, t.use).TextColor(core.Tokens(c).TextMuted).
						Width(unit(c, use)).FontSize(core.FontSize(c, theme.CaptionSize))
				})
		}
	})
}

// themeSwatchHex is one cell of the two-value pair. The value this section is
// about is in body ink and the other one is faint, so a reader can tell at a
// glance which column is the subject of the section they are reading.
func themeSwatchHex(c *ui.Context, col ui.Color, w float32, mine bool) {
	k := core.Tokens(c)
	ink := k.TextFaint
	if mine {
		ink = k.Text
	}
	ui.Box(c).Width(unit(c, w)).Children(func() {
		ui.Row(c).Gap(unit(c, 1)).AlignItems(ui.Center).Children(func() {
			ui.Box(c).Size(unit(c, 2), unit(c, 2)).Shrink(0).Radius(unit(c, 1)).
				Background(col).Border(theme.BorderWidth, k.Border)
			ui.Text(c, themeHex(col)).TextColor(ink).
				FontSize(core.FontSize(c, theme.CaptionSize))
		})
	})
}

// themeSurfaceSpecimen draws the one relationship the whole palette exists for:
// grey on white, and the three inks that go on each.
func themeSurfaceSpecimen(c *ui.Context, t theme.Tokens) {
	ui.Box(c).Width(unit(c, 236)).Height(unit(c, 40)).Shrink(0).
		Radius(theme.CardRadius).Background(t.Background).
		Border(theme.BorderWidth, t.Border).Clip().Children(func() {
		ui.Row(c).Fill().Children(func() {
			// The sidebar: Surface, and the name is Text on it.
			ui.Box(c).Width(unit(c, 36)).FillHeight().Background(t.Surface).
				Padding(unit(c, 2), unit(c, 2)).Children(func() {
				ui.Column(c).FillWidth().Gap(unit(c, 1.25)).Children(func() {
					ui.Text(c, "Callbacks").TextColor(t.Text).
						FontSize(core.FontSize(c, theme.TitleSize) * 0.6).Bold()
					for _, r := range []struct{ s, tone string }{
						{"Open callbacks", "text"},
						{"Waiting on parts", "muted"},
						{"Shipped today", "faint"},
					} {
						ink := t.Text
						switch r.tone {
						case "muted":
							ink = t.TextMuted
						case "faint":
							ink = t.TextFaint
						}
						ui.Text(c, r.s).TextColor(ink).
							FontSize(core.FontSize(c, theme.RowSize))
					}
					ui.Box(c).FillWidth().Height(theme.BorderWidth).Background(t.Border)
					ui.Box(c).Width(unit(c, 14)).Height(unit(c, 6)).Shrink(0).
						Radius(theme.PillRadius).Background(t.Fill).Center().Children(func() {
						ui.Text(c, "Riverside").TextColor(t.OnFill).
							FontSize(core.FontSize(c, theme.CaptionSize))
					})
				})
			})
			// The main area: Background, with the three inks again.
			ui.Box(c).Grow(1).FillHeight().Padding(unit(c, 2), unit(c, 2)).
				Children(func() {
					ui.Column(c).FillWidth().Gap(unit(c, 1.25)).Children(func() {
						ui.Text(c, "Riverside Clinic").TextColor(t.Text).
							FontSize(core.FontSize(c, theme.BodySize)).Bold()
						ui.Text(c, "CB-2871 · AC repair · 09:14").TextColor(t.TextMuted).
							FontSize(core.FontSize(c, theme.MetaSize))
						ui.Text(c, "分隔符 · 记号 · 不承载阅读").TextColor(t.TextFaint).
							FontSize(core.FontSize(c, theme.CaptionSize))
						ui.Box(c).FillWidth().Height(unit(c, 5)).Radius(theme.SmallRadius).
							Background(t.SurfaceHover)
						ui.Box(c).FillWidth().Height(unit(c, 5)).Radius(theme.SmallRadius).
							Background(t.SurfacePressed)
						ui.Row(c).Gap(unit(c, 1)).AlignItems(ui.Center).Children(func() {
							ui.Box(c).Size(unit(c, 2.5), unit(c, 2.5)).Shrink(0).
								Radius(unit(c, 1.25)).Background(t.Lively)
							ui.Text(c, "Lively 两套一样").TextColor(t.Text).
								FontSize(core.FontSize(c, theme.CaptionSize))
						})
					})
				})
		})
	})
}

// themeSeveritySpecimen draws the four ramps as pairs, which is the only way a
// reader can check that the foreground was chosen for that background rather
// than for the window's.
func themeSeveritySpecimen(c *ui.Context, t theme.Tokens) {
	pairs := []struct {
		name   string
		bg, fg ui.Color
	}{
		{"DangerBg / Danger", t.DangerBg, t.Danger},
		{"WarningBg / Warning", t.WarningBg, t.Warning},
		{"SuccessBg / Success", t.SuccessBg, t.Success},
		{"AccentBg / AccentText", t.AccentBg, t.AccentText},
	}
	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 2)).Children(func() {
		for _, p := range pairs {
			ui.Box(c).Grow(1).Padding(unit(c, 1.25), unit(c, 1.5)).
				Radius(theme.ControlRadius).Background(p.bg).Children(func() {
				ui.Column(c).FillWidth().Gap(unit(c, 0.5)).Children(func() {
					ui.Text(c, p.name).TextColor(p.fg).
						FontSize(core.FontSize(c, theme.CaptionSize)).Bold()
					ui.Text(c, themeHex(p.bg)+" 底 / "+themeHex(p.fg)+" 字").
						TextColor(p.fg).FontSize(core.FontSize(c, theme.CaptionSize))
				})
			})
		}
	})
}

// themeFillSpecimen is the saturated-by-ink surface and the two that carry no ramp.
func themeFillSpecimen(c *ui.Context, t theme.Tokens) {
	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 2)).Children(func() {
		ui.Box(c).Width(unit(c, 60)).Height(unit(c, 14)).Shrink(0).
			Radius(theme.ControlRadius).Background(t.Fill).Center().Children(func() {
			ui.Text(c, "主按钮 · Fill / OnFill").TextColor(t.OnFill).
				FontSize(core.FontSize(c, theme.RowSize)).Bold()
		})
		ui.Box(c).Width(unit(c, 48)).Height(unit(c, 14)).Shrink(0).
			Radius(theme.PillRadius).Background(t.Lively).Center().Children(func() {
			ui.Text(c, "Lively").TextColor(t.Text).
				FontSize(core.FontSize(c, theme.RowSize)).Bold()
		})
		ui.Box(c).Width(unit(c, 48)).Height(unit(c, 14)).Shrink(0).
			Radius(theme.ControlRadius).Background(t.AccentBg).Center().Children(func() {
			ui.Text(c, "Accent 文字").TextColor(t.AccentText).
				FontSize(core.FontSize(c, theme.RowSize)).Bold()
		})
		ui.Box(c).Grow(1).Height(unit(c, 14)).Shrink(0).
			Radius(theme.ControlRadius).Background(t.SurfaceHover).Center().Children(func() {
			ui.Text(c, "SurfaceHover").TextColor(t.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize))
		})
	})
}

// themeScalesSection is everything that is NOT colour: the radii, the outline
// weight, the spacing ladder, the type scale, the two densities and the three
// shell measurements.
//
// It is drawn twice, once per palette, and the point of drawing it twice is
// that the two drawings are the same. Only the colours behind them change —
// which is the answer to the question a reader of a two-appearance interface
// actually has, and it is much easier to see than to read.
//
// The whole of it sits on a band of this palette's own Surface, because the
// ink in it is this palette's own Text: a dark palette's Text is #f4f4f5, and
// #f4f4f5 on the light page is nothing at all. The band is what makes the
// second drawing legible — and it is also the honest statement of the claim,
// since the scales are drawn in body ink on a panel in both sections.
func themeScalesSection(c *ui.Context, t theme.Tokens, isLight bool) {
	same := "两套完全一致"
	if !isLight {
		same = "与浅色那一节逐项相同"
	}

	ui.Box(c).Width(unit(c, 236)).Radius(theme.CardRadius).
		Background(t.Surface).Padding(unit(c, 3), unit(c, 3), unit(c, 3), unit(c, 3)).
		Children(func() {
			ui.Column(c).FillWidth().Children(func() {
				showcase.Field(c, "圆角族 · "+same+"（一个族，大值即药丸）")
				ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 2)).AlignItems(ui.End).
					Children(func() {
						for _, r := range []struct {
							v    float32
							name string
						}{
							{theme.SmallRadius, "SmallRadius 8"},
							{theme.ControlRadius, "ControlRadius 14"},
							{theme.CardRadius, "CardRadius 18"},
							{theme.PanelRadius, "PanelRadius 24"},
							{theme.PillRadius, "PillRadius 999"},
						} {
							themeRadiusChip(c, t, r.v, r.name)
						}
					})

				showcase.Field(c, "描边 · "+same+"：全库一个描边值，所以每条边都一样粗")
				ui.Column(c).Width(unit(c, 236)).Gap(unit(c, 1)).Children(func() {
					ui.Row(c).Gap(unit(c, 2)).AlignItems(ui.Center).Children(func() {
						ui.Box(c).Width(unit(c, 20)).Height(unit(c, 8)).Shrink(0).
							Radius(theme.SmallRadius).Background(t.Background).
							Border(theme.BorderWidth, t.Border)
						ui.Text(c, "BorderWidth 1").TextColor(t.Text).
							FontSize(core.FontSize(c, theme.CaptionSize)).Bold()
						ui.Text(c, "1px "+themeHex(t.Border)+" 描边").TextColor(t.TextMuted).
							FontSize(core.FontSize(c, theme.CaptionSize))
					})
					ui.Text(c, "圆角那一族的五个值、描边的 1，都不随主题变；"+
						"变的是描边本身取哪个色。").TextColor(t.TextFaint).
						FontSize(core.FontSize(c, theme.CaptionSize))
				})

				showcase.Field(c, "间距阶 · "+same+"：u = Density.Unit()，组件里所有的 u*n 都是它")
				themeSpacingLadder(c, t)

				showcase.Field(c, "字号刻度 · "+same+"：半号是有意的，13.5 的次级标签比 13 立得住")
				themeTypeScale(c, t)

				showcase.Field(c, "密度 · theme.Compact 与 theme.Comfortable · "+same+
					"：这一族只有两档，Compact 是零值")
				themeDensitySpecimen(c, t, theme.Compact, "紧凑 Compact")
				themeDensitySpecimen(c, t, theme.Comfortable, "舒适 Comfortable")

				showcase.Field(c, "外壳尺寸 · "+same+"：常量不是比例，窗口窄了就滚动，不压窄")
				ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 2)).Children(func() {
					themeShellMetric(c, t, "SidebarWidth 322", theme.SidebarWidth)
					themeShellMetric(c, t, "RailWidth 76", theme.RailWidth)
					themeShellMetric(c, t, "ColumnWidth 272", theme.ColumnWidth)
				})
			})
		})
}

// themeRadiusChip is one value of the family, drawn at its own curvature so the
// shape is the reading rather than the number under it.
func themeRadiusChip(c *ui.Context, t theme.Tokens, r float32, name string) {
	ui.Column(c).AlignItems(ui.Center).Gap(unit(c, 1)).Children(func() {
		// Background rather than Surface: the chip is already standing on a
		// Surface band, and a Surface chip on a Surface band is a hole.
		ui.Box(c).Width(unit(c, 18)).Height(unit(c, 9)).Radius(r).Background(t.Background).
			Border(theme.BorderWidth, t.Border)
		ui.Text(c, name).TextColor(t.TextMuted).
			FontSize(core.FontSize(c, theme.CaptionSize))
	})
}

// themeSpacingLadder is the grid itself: eight bars of one, two, three … eight
// steps, so a reader can count rather than be told.
func themeSpacingLadder(c *ui.Context, t theme.Tokens) {
	for _, d := range []theme.Density{theme.Compact, theme.Comfortable} {
		u := d.Unit()
		ui.Column(c).FillWidth().Gap(unit(c, 0.5)).Children(func() {
			ui.Row(c).Gap(unit(c, 2)).AlignItems(ui.Center).Children(func() {
				ui.Text(c, d.String()).TextColor(t.Text).
					Width(unit(c, 22)).
					FontSize(core.FontSize(c, theme.CaptionSize)).Bold()
				ui.Text(c, fmt.Sprintf("u = %g", u)).TextColor(t.TextMuted).
					Width(unit(c, 12)).FontSize(core.FontSize(c, theme.CaptionSize))
			})
			ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 2)).AlignItems(ui.End).
				Children(func() {
					for n := 1; n <= 8; n++ {
						ui.Column(c).AlignItems(ui.Center).Gap(unit(c, 0.5)).Children(func() {
							ui.Box(c).Width(unit(c, 4)).Height(u * float32(n)).Shrink(0).
								Radius(unit(c, 0.5)).Background(t.SurfacePressed)
							ui.Text(c, fmt.Sprintf("u×%d", n)).TextColor(t.TextFaint).
								FontSize(core.FontSize(c, 9)).SingleLine()
						})
					}
				})
		})
	}
}

// themeTypeScale draws every size at the size it names, so the section is a
// specimen sheet and not a table of numbers.
func themeTypeScale(c *ui.Context, t theme.Tokens) {
	sizes := []struct {
		name string
		size float32
		use  string
	}{
		{"DisplaySize 40", theme.DisplaySize, "页面大标题"},
		{"TitleSize 24", theme.TitleSize, "侧栏标题"},
		{"SheetSize 21", theme.SheetSize, "抽屉 / 对话框标题"},
		{"LeadSize 17", theme.LeadSize, "空态标题"},
		{"BodySize 15.5", theme.BodySize, "卡片客户名"},
		{"StatSize 14.5", theme.StatSize, "统计与金额"},
		{"RowSize 13.5", theme.RowSize, "列表行、筛选行"},
		{"MetaSize 13", theme.MetaSize, "卡片副行、单元"},
		{"CaptionSize 12", theme.CaptionSize, "药丸、角标"},
		{"MonoSize 11.5", theme.MonoSize, "头像缩写"},
		{"GlyphSize 11", theme.GlyphSize, "信号格"},
		{"IconSize 21", theme.IconSize, "图标尺寸"},
	}
	ui.Column(c).Width(unit(c, 236)).Children(func() {
		for _, s := range sizes {
			ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 2)).AlignItems(ui.Center).
				Children(func() {
					ui.Text(c, s.name).TextColor(t.TextMuted).
						Width(unit(c, 24)).FontSize(core.FontSize(c, theme.CaptionSize))
					ui.Text(c, "Riverside Clinic").TextColor(t.Text).
						Width(unit(c, 120)).FontSize(core.FontSize(c, s.size)).SingleLine()
					ui.Text(c, s.use).TextColor(t.TextFaint).
						Width(unit(c, 76)).FontSize(core.FontSize(c, theme.CaptionSize))
				})
		}
	})
}

// themeDensitySpecimen builds one panel at one density, so the difference between
// the two is a number the reader can see rather than a sentence about it.
func themeDensitySpecimen(c *ui.Context, t theme.Tokens, d theme.Density, name string) {
	step := d.Unit()
	h := 30.0
	if d == theme.Comfortable {
		h = 36
	}
	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 3)).Children(func() {
		ui.Box(c).Width(unit(c, 26)).FillHeight().Background(t.Background).
			Radius(theme.ControlRadius).Padding(step*2, step*2).Children(func() {
			ui.Column(c).FillWidth().Gap(step).Children(func() {
				for range 3 {
					ui.Box(c).FillWidth().Height(step * 2).Shrink(0).
						Radius(theme.SmallRadius).Background(t.Border)
				}
			})
		})
		ui.Column(c).FillHeight().Gap(unit(c, 1)).Children(func() {
			ui.Text(c, name).TextColor(t.Text).
				FontSize(core.FontSize(c, theme.RowSize)).Bold()
			ui.Text(c, fmt.Sprintf("u = %g · 控件高 %g", step, h)).
				TextColor(t.TextMuted).FontSize(core.FontSize(c, theme.CaptionSize))
		})
	})
}

// themeShellMetric is one window measurement as a bar at its own width, so "322"
// is a number and a proportion at the same time.
//
// The bars share one scale — 88 points wide is SidebarWidth — rather than
// their real widths, because three real widths is 670 points of a 944-point
// page and the last bar is the one that falls off the edge.
func themeShellMetric(c *ui.Context, t theme.Tokens, name string, v float32) {
	ui.Column(c).Grow(1).AlignItems(ui.Center).Gap(unit(c, 1)).Children(func() {
		const scale = 22 // u
		ui.Box(c).Width(unit(c, scale)).Height(unit(c, 5)).Shrink(0).
			Radius(theme.SmallRadius).Background(t.SurfacePressed).Children(func() {
			ui.Box(c).Width(unit(c, scale) * v / theme.SidebarWidth).
				FillHeight().Radius(theme.SmallRadius).Background(t.Accent)
		})
		ui.Text(c, name).TextColor(t.TextMuted).
			FontSize(core.FontSize(c, theme.CaptionSize))
	})
}

// themeTokensPair is the last thing on the section: the two constants side by side,
// with the IsDark() answer under each, because that is the whole of what
// "which appearance is this" means once the tokens are resolved.
func themeTokensPair(c *ui.Context, here, other theme.Tokens, isLight bool) {
	name := func(t theme.Tokens, isLight bool) string {
		if isLight {
			return "Light"
		}
		return "Dark"
	}
	card := func(t theme.Tokens, isLight bool) {
		ui.Column(c).Grow(1).Gap(unit(c, 1)).Children(func() {
			// A window, at this palette's own values: white (or near-black)
			// with a grey panel on it and one accent mark. It is the same
			// three-block picture in both cards, so the only thing a reader
			// is comparing is the colour.
			ui.Box(c).FillWidth().Height(unit(c, 12)).Shrink(0).
				Radius(theme.SmallRadius).Background(t.Background).
				Border(theme.BorderWidth, t.Border).Padding(unit(c, 1.5)).
				Children(func() {
					// The row is given a height: the four blocks under it are
					// FillHeight, which in a row whose height is its content's
					// resolves to nothing — the row would be as tall as nothing
					// and so would the blocks.
					ui.Row(c).FillWidth().Height(unit(c, 8)).Gap(unit(c, 1.5)).
						Children(func() {
							ui.Box(c).Width(unit(c, 6)).FillHeight().Shrink(0).
								Radius(theme.SmallRadius).Background(t.Surface)
							ui.Box(c).Width(unit(c, 4)).FillHeight().Shrink(0).
								Radius(theme.SmallRadius).Background(t.SurfaceHover)
							ui.Box(c).Width(unit(c, 4)).FillHeight().Shrink(0).
								Radius(theme.SmallRadius).Background(t.SuccessBg)
							ui.Box(c).Width(unit(c, 3)).FillHeight().Shrink(0).
								Radius(unit(c, 1.5)).Background(t.Accent)
						})
				})
			ui.Text(c, name(t, isLight)+"()").TextColor(core.Tokens(c).Text).
				FontSize(core.FontSize(c, theme.RowSize)).Bold()
			ui.Text(c, "IsDark").TextColor(core.Tokens(c).TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize)).Bold()
			ui.Text(c, fmt.Sprintf("%v", t.IsDark())).TextColor(core.Tokens(c).TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize))
			ui.Text(c, "Background "+themeHex(t.Background)+" · Text "+themeHex(t.Text)).
				TextColor(core.Tokens(c).TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
		})
	}
	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 6)).Children(func() {
		if isLight {
			card(here, true)
			card(other, false)
		} else {
			card(other, true)
			card(here, false)
		}
	})
}
