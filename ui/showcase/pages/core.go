// Package pages holds the gallery's pages: one per ui/ package, each drawing
// everything that package has to show.
//
// The point of the gallery is that a component can be looked at. A test can
// say a card draws its title; only a picture says whether the card is any
// good. So every page here is laid out for a person, not for an assertion:
// grouped, spaced, and tall enough that nothing is squeezed.
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
		Package: "core",
		Title:   "ui/core — 这一帧解析成什么",
		Note:    "每帧一次 core.Use，下面每一块都从它解析出来的调色板、密度、字号里取色取距",
		Width:   1000,
		Height:  1450,
		Want: []string{
			"Background",
			"Surface",
			"SurfaceHover",
			"SurfacePressed",
			"Text",
			"TextMuted",
			"TextFaint",
			"Border",
			"Neutral",
			"Accent",
			"Success",
			"Warning",
			"Danger",
			"Lively",
			"u = 4 · 控件高 30",
			"u = 6 · 控件高 36",
			"DisplaySize 40",
			"Pill 999",
			"已改写 · 来自窗口",
		},
		Render: func(c *ui.Context) {
			corePage(c)
		},
	})
}

// corePage is not a list of the package's functions. core has nothing a
// person presses — it is the frame every other page is resolved against — so
// what this page shows is what core hands down: the palette, the spacing
// step, the type scale, the radii, and the five severity pairs that every
// pill and badge in the library agrees on.
func corePage(c *ui.Context) {
	k := core.Tokens(c)

	showcase.Section(c, "表面 — 整个界面只有两层表面关系")
	ui.Row(c).FillWidth().Gap(unit(c, 3)).Children(func() {
		surfaceCard(c, k.Background, k.Text, "Background", "窗口、图标栏、卡片、抽屉")
		surfaceCard(c, k.Surface, k.Text, "Surface", "侧栏、看板列、图标块、计数药丸")
		surfaceCard(c, k.SurfaceHover, k.Text, "SurfaceHover", "悬停态")
		surfaceCard(c, k.SurfacePressed, k.Text, "SurfacePressed", "按下态")
	})

	showcase.Section(c, "墨色与描边 — 其余全是黑白灰")
	ui.Row(c).FillWidth().Gap(unit(c, 3)).Children(func() {
		inkCard(c, "Text", k.Text, "标题、客户名、金额", true)
		inkCard(c, "TextMuted", k.TextMuted, "工单号、技师名、日期", true)
		inkCard(c, "TextFaint", k.TextFaint, "分隔符等记号", true)
		inkCard(c, "Border", k.Border, "描边、分组线、选中行的药丸", false)
	})

	showcase.Section(c, "状态色 — 成对的底 + 字，业务代码不自己挑颜色")
	severityRamp(c, k)

	showcase.Section(c, "Severity.Pair — 全库唯一决定「哪个底配哪个字」的地方")
	ui.Row(c).FillWidth().Gap(unit(c, 2)).Children(func() {
		for _, s := range []core.Severity{
			core.Neutral, core.Accent, core.Success, core.Warning, core.Danger,
		} {
			bg, fg := s.Pair(k)
			ui.Box(c).Grow(1).Padding(unit(c, 1.5), unit(c, 2)).
				Radius(theme.PillRadius).Background(bg).Center().Children(func() {
				ui.Text(c, s.String()).TextColor(fg).
					FontSize(core.FontSize(c, theme.RowSize)).Bold()
			})
		}
	})

	showcase.Section(c, "深色填充面 Fill / OnFill — 只在控件需要时出现")
	ui.Row(c).FillWidth().Gap(unit(c, 3)).Children(func() {
		ui.Box(c).Grow(1).Height(unit(c, 13)).Radius(theme.ControlRadius).
			Background(k.Fill).Center().Children(func() {
			ui.Text(c, "主按钮 · Toast · 图标栏激活态").TextColor(k.OnFill).
				FontSize(core.FontSize(c, theme.RowSize)).Bold()
		})
		ui.Box(c).Size(unit(c, 13), unit(c, 13)).Radius(unit(c, 6.5)).
			Background(k.Lively).Center().Children(func() {
			ui.Text(c, "Lively").TextColor(k.Text).
				FontSize(core.FontSize(c, theme.CaptionSize)).Bold()
		})
		ui.Box(c).Grow(1).Padding(unit(c, 1.5), unit(c, 2)).
			Radius(theme.ControlRadius).Background(k.Surface).
			Center().Children(func() {
			ui.Text(c, "Accent").TextColor(k.AccentText).
				FontSize(core.FontSize(c, theme.RowSize)).Bold()
		})
	})

	showcase.Section(c, "密度 — 间距网格的一步，Compact = 4，Comfortable = 6")
	ui.Row(c).FillWidth().Gap(unit(c, 4)).Children(func() {
		densitySpecimen(c, theme.Compact, "紧凑 Compact")
		densitySpecimen(c, theme.Comfortable, "舒适 Comfortable")
	})

	showcase.Section(c, "字号刻度 — 半号是有意的，13.5 比 13 立得住")
	typeScale(c, k)

	showcase.Section(c, "圆角族 — 一个族，大值即药丸")
	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.End).Children(func() {
		radiusChip(c, theme.SmallRadius, "Small 8")
		radiusChip(c, theme.ControlRadius, "Control 14")
		radiusChip(c, theme.CardRadius, "Card 18")
		radiusChip(c, theme.PanelRadius, "Panel 24")
		radiusChip(c, theme.PillRadius, "Pill 999")
	})

	showcase.Section(c, "文案覆盖 Msg — 窗口改写库里的字，业务文案不走这里")
	messageDemo(c, k)

	showcase.Section(c, "文字缩放与控件高度 — 窗口说了算，组件只跟着算")
	scaleDemo(c, k)
}

// surfaceCard shows one palette entry as a surface with something on it,
// which is the only way a swatch says what it is *for*.
func surfaceCard(c *ui.Context, bg, ink ui.Color, name, use string) {
	ui.Box(c).Grow(1).Padding(unit(c, 2)).Radius(theme.ControlRadius).Background(bg).
		Border(theme.BorderWidth, core.Tokens(c).Border).Children(func() {
		ui.Column(c).FillWidth().Gap(unit(c, 0.5)).Children(func() {
			ui.Text(c, name).TextColor(ink).FontSize(core.FontSize(c, theme.RowSize)).Bold()
			ui.Text(c, use).TextColor(core.Tokens(c).TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize))
		})
	})
}

func inkCard(c *ui.Context, name string, ink ui.Color, use string, withBody bool) {
	k := core.Tokens(c)
	ui.Box(c).Grow(1).Padding(unit(c, 2)).Radius(theme.ControlRadius).Background(k.Background).
		Border(theme.BorderWidth, k.Border).Children(func() {
		ui.Column(c).FillWidth().Gap(unit(c, 0.5)).Children(func() {
			if withBody {
				ui.Text(c, "Riverside Clinic").TextColor(ink).
					FontSize(core.FontSize(c, theme.BodySize))
			} else {
				ui.Box(c).FillWidth().Height(unit(c, 6)).Radius(theme.SmallRadius).
					Background(ink)
			}
			ui.Text(c, name).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize)).Bold()
			ui.Text(c, use).TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize))
		})
	})
}

// severityRamp shows both halves of every status pair, so a glance says the
// foreground was picked for that background and not for the window's.
func severityRamp(c *ui.Context, k theme.Tokens) {
	pairs := []struct {
		name   string
		bg, fg ui.Color
	}{
		{"危险", k.DangerBg, k.Danger},
		{"告警", k.WarningBg, k.Warning},
		{"成功", k.SuccessBg, k.Success},
		{"强调", k.AccentBg, k.AccentText},
	}
	ui.Column(c).FillWidth().Gap(unit(c, 1.5)).Children(func() {
		for _, p := range pairs {
			ui.Row(c).FillWidth().Gap(unit(c, 2)).AlignItems(ui.Center).Children(func() {
				ui.Text(c, p.name).TextColor(k.Text).
					FontSize(core.FontSize(c, theme.RowSize)).Width(unit(c, 11)).Shrink(0)
				ui.Box(c).Grow(1).Height(unit(c, 7)).Radius(theme.ControlRadius).
					Background(p.bg).Center().Children(func() {
					ui.Text(c, "底 + 字成对读").TextColor(p.fg).
						FontSize(core.FontSize(c, theme.CaptionSize)).Bold()
				})
			})
		}
	})
}

// densitySpecimen builds the same panel at both densities, so the difference
// is a number the reader can see rather than a sentence about it.
func densitySpecimen(c *ui.Context, d theme.Density, name string) {
	k := core.Tokens(c)
	step := d.Unit()
	h := 30.0
	if d == theme.Comfortable {
		h = 36
	}
	ui.Column(c).Grow(1).Gap(unit(c, 1.5)).Children(func() {
		ui.Text(c, name).TextColor(k.Text).FontSize(core.FontSize(c, theme.RowSize)).Bold()
		ui.Box(c).FillWidth().Padding(step*2, step*2, step*2, step*2).
			Radius(theme.ControlRadius).Background(k.Surface).Children(func() {
			ui.Column(c).FillWidth().Gap(step).Children(func() {
				for range 3 {
					ui.Box(c).FillWidth().Height(step * 2).Radius(theme.SmallRadius).
						Background(k.Border)
				}
			})
		})
		ui.Text(c, fmt.Sprintf("u = %g · 控件高 %g", step, h)).
			TextColor(k.TextMuted).FontSize(core.FontSize(c, theme.CaptionSize))
	})
}

// typeScale draws every size at the size it names, so the page is a
// specimen sheet and not a table of numbers.
func typeScale(c *ui.Context, k theme.Tokens) {
	sizes := []struct {
		name string
		size float32
		use  string
	}{
		{"DisplaySize 40", theme.DisplaySize, "页面大标题"},
		{"TitleSize 24", theme.TitleSize, "面板标题"},
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
	ui.Column(c).FillWidth().Children(func() {
		for _, s := range sizes {
			ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Center).Children(func() {
				ui.Text(c, s.name).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize)).Width(unit(c, 32))
				ui.Text(c, "Riverside Clinic").TextColor(k.Text).
					FontSize(core.FontSize(c, s.size)).Grow(1)
				ui.Text(c, s.use).TextColor(k.TextFaint).
					FontSize(core.FontSize(c, theme.CaptionSize))
			})
		}
	})
}

func radiusChip(c *ui.Context, r float32, name string) {
	k := core.Tokens(c)
	ui.Column(c).AlignItems(ui.Center).Gap(unit(c, 1)).Children(func() {
		ui.Box(c).Width(unit(c, 18)).Height(unit(c, 9)).Radius(r).Background(k.Surface).
			Border(theme.BorderWidth, k.Border)
		ui.Text(c, name).TextColor(k.TextMuted).FontSize(core.FontSize(c, theme.CaptionSize))
	})
}

// messageDemo shows a string the library would draw, and the same string as
// the window's copy has overridden it — the reason Msg exists at all.
func messageDemo(c *ui.Context, k theme.Tokens) {
	// A key nothing else in the library reads, so installing an override
	// here cannot leak into another component drawn on this page.
	c = core.WithMessages(c, map[string]string{"core.demo": "已改写 · 来自窗口"})
	ui.Row(c).FillWidth().Gap(unit(c, 2)).Children(func() {
		demo := func(key, def, note string) {
			ui.Column(c).Grow(1).Padding(unit(c, 1.5), unit(c, 2)).
				Radius(theme.ControlRadius).Background(k.Surface).Gap(unit(c, 0.5)).
				Children(func() {
					ui.Text(c, core.Msg(c, key, def)).TextColor(k.Text).
						FontSize(core.FontSize(c, theme.RowSize)).Bold()
					ui.Text(c, note).TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.CaptionSize))
				})
		}
		demo("core.demo", "默认文案", "窗口装了覆盖")
		demo("core.none", "默认文案", "窗口没装这个键")
	})
}

// scaleDemo shows the two per-frame knobs a window sets and a page does not:
// the desktop's text scale, and the control height it implies.
func scaleDemo(c *ui.Context, k theme.Tokens) {
	ui.Row(c).FillWidth().Gap(unit(c, 3)).Children(func() {
		for _, s := range []float32{1, 1.15, 1.3} {
			ui.Column(c).Grow(1).Padding(unit(c, 1.5), unit(c, 2)).
				Radius(theme.ControlRadius).Background(k.Surface).Gap(unit(c, 1.5)).
				Children(func() {
					ui.Text(c, "文字缩放 ×"+fmt.Sprintf("%g", s)).TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.CaptionSize))
					ui.Text(c, "Riverside").TextColor(k.Text).
						FontSize(core.FontSize(c, theme.BodySize) * s)
					ui.Box(c).FillWidth().Height(core.ControlHeight(c)).
						Radius(theme.PillRadius).Background(k.Fill).Center().Children(func() {
						ui.Text(c, "标准控件").TextColor(k.OnFill).
							FontSize(core.FontSize(c, theme.CaptionSize))
					})
				})
		}
	})
}

// unit is the one line of arithmetic every page on this gallery needs: the
// density's own spacing step times n. It is here rather than in each page
// because the whole library is built from it, and a gallery that spelled it
// out eight times would be eight chances to spell it differently.
func unit(c *ui.Context, n float32) float32 {
	return core.Density(c).Unit() * n
}

// share is the width percentage one of n columns takes in a row whose gaps
// are unit(c, 3) wide.
//
// It exists because percentages that add up to 100 are 1000 points wide
// *before* the gaps, so a two-column row at 50 and 50 is 1012 points wide on
// a 1000-point page: the right-hand column is drawn off the page and every
// component in it is clipped rather than laid out. Leaving each gap its share
// is the whole of the correction.
func share(n int) float32 {
	const (
		page = 1000.0 // every page on this gallery is 1000 points wide
		gap  = 12.0   // unit(c, 3) at the compact density
	)
	free := 100 - float32(n-1)*gap/page*100
	return free / float32(n)
}
