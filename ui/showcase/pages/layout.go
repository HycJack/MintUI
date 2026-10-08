package pages

import (
	"strconv"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/showcase"
	"github.com/HycJack/MintUI/ui/theme"
)

func init() {
	showcase.Register(showcase.Page{
		Package: "layout",
		Title:   "ui/layout — 一块屏的骨架",
		Note:    "容器、分栏、标题栏、状态栏、App 外壳、页头、看板列 —— 这页画的是它们能搭出什么，不是它们的 API",
		Width:   1000,
		Height:  2100,
		Want: []string{
			"描边",
			"灰底",
			"白底",
			"居中",
			"左栏",
			"右栏",
			"TrafficLights",
			"● 27 未处理",
			"Open callbacks",
			"Riverside Clinic",
			"9 technicians",
			"未处理",
			"进行中",
			"已解决",
			"还有 5 条",
			"这一列是空的",
			"按分支分组",
			"记一条回访",
			"我的回访",
			"Callbacks  /  未处理",
		},
		Render: func(c *ui.Context) {
			layoutPage(c)
		},
	})
}

// layoutPage is the frame every other package is drawn inside, so the page
// does not enumerate the package's functions. It builds the shapes: a
// container, a stack, a grid, a pane, a bar, and then the whole window frame
// with a sidebar, a header, a board of lanes and a status bar — which is
// what "a screen is arranged like this" looks like.
func layoutPage(c *ui.Context) {
	containerSection(c)
	stackAndGridSection(c)
	paneSection(c)
	barsSection(c)
	panelSection(c)
	headerSection(c)
	boardSection(c)
	shellSection(c)
}

// containerSection shows the padded, optionally surfaced box every panel in
// the library is — a plain one, a grey one, a bordered one and a reading
// column with a cap on its width.
func containerSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "Container — 除了组件以外最小的东西")
	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		sample := func(opts layout.ContainerOptions, title, note string) {
			layout.Container(c, opts, func() {
				ui.Column(c).FillWidth().Gap(unit(c, 0.5)).Children(func() {
					ui.Text(c, title).TextColor(k.Text).
						FontSize(core.FontSize(c, theme.RowSize)).Bold()
					ui.Text(c, note).TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.CaptionSize))
				})
			})
		}
		sample(layout.ContainerOptions{
			Width: unit(c, 55), Pad: unit(c, 2), Radius: theme.ControlRadius,
			Border: true,
		}, "描边", "Radius + Border")
		sample(layout.ContainerOptions{
			Width: unit(c, 55), Pad: unit(c, 2), Radius: theme.ControlRadius,
			Surface: true,
		}, "灰底", "Surface，低一层")
		sample(layout.ContainerOptions{
			Width: unit(c, 55), Pad: unit(c, 2), Radius: theme.ControlRadius,
		}, "白底", "什么都不加")
		sample(layout.ContainerOptions{
			Width: unit(c, 55), Height: unit(c, 22), Pad: unit(c, 2),
			Radius: theme.ControlRadius, Border: true, Center: true, Gap: unit(c, 1),
		}, "居中", "Center + 定高")
	})

	showcase.Field(c, "Divider — 一根线；竖线必须是所在容器的孩子，否则一边长高它就落错地方")
	ui.Column(c).FillWidth().Gap(unit(c, 2)).Children(func() {
		ui.Row(c).FillWidth().Height(unit(c, 11)).AlignItems(ui.Center).Gap(unit(c, 2)).
			Children(func() {
				ui.Text(c, "左").TextColor(k.TextMuted).FontSize(core.FontSize(c, theme.CaptionSize))
				layout.Divider(c, layout.DividerOptions{Vertical: true})
				ui.Text(c, "中").TextColor(k.TextMuted).FontSize(core.FontSize(c, theme.CaptionSize))
				layout.Divider(c, layout.DividerOptions{Vertical: true})
				ui.Text(c, "右").TextColor(k.TextMuted).FontSize(core.FontSize(c, theme.CaptionSize))
			})
		layout.Divider(c, layout.DividerOptions{})
		ui.Text(c, "横线：满宽").TextColor(k.TextMuted).FontSize(core.FontSize(c, theme.CaptionSize))
	})
}

// stackAndGridSection shows two children drawn on top of one another, and
// children laid out in tracks.
func stackAndGridSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "Stack 与 Grid — 叠在一起 / 排成轨道")
	ui.Row(c).FillWidth().Gap(unit(c, 4)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).Gap(unit(c, 1.5)).Children(func() {
			ui.Text(c, "Stack：第一个孩子定尺寸，其余的画在上面").TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize))
			ui.Row(c).Gap(unit(c, 6)).Children(func() {
				// A badge over an avatar: the case Stack exists for.
				layout.Stack(c, layout.StackOptions{}, func() {
					ui.Box(c).Size(unit(c, 20), unit(c, 20)).Radius(unit(c, 10)).
						Background(k.Surface).Center().Children(func() {
						ui.Text(c, "AT").TextColor(k.Text).
							FontSize(core.FontSize(c, theme.MonoSize)).Bold()
					})
					display.Badge(c, 3, display.BadgeOptions{
						Tone: core.Danger, Name: "3 条未读",
					}).Attach(ui.AnchorTopRight, ui.AnchorCenter).Right(-unit(c, 2))
				})
				layout.Stack(c, layout.StackOptions{}, func() {
					ui.Box(c).Width(unit(c, 38)).Height(unit(c, 18)).
						Radius(theme.ControlRadius).Background(k.Surface).
						Border(theme.BorderWidth, k.Border)
					layout.Container(c, layout.ContainerOptions{
						Pad: unit(c, 1), Radius: theme.PillRadius, Border: true,
					}, func() {
						ui.Text(c, "盖在上面").TextColor(k.Text).
							FontSize(core.FontSize(c, theme.CaptionSize)).Bold()
					}).Attach(ui.AnchorRight, ui.AnchorCenter)
				})
			})
		})

		ui.Column(c).Grow(1).Gap(unit(c, 1.5)).Children(func() {
			ui.Text(c, "Grid / GridCell：Grid 自己是一列，轨道由你摆").TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize))
			cell := func(text, bg string, track float32) {
				layout.GridCell(c, layout.GridCellOptions{Track: track, Pad: unit(c, 1)}, func() {
					ui.Box(c).FillWidth().Height(unit(c, 8)).Radius(theme.SmallRadius).
						Background(ui.Hex(bg)).Center().Children(func() {
						ui.Text(c, text).TextColor(k.Text).
							FontSize(core.FontSize(c, theme.CaptionSize))
					})
				})
			}
			row := func(parts ...func()) {
				layout.Grid(c, layout.GridOptions{Width: unit(c, 64), Gap: unit(c, 1)}, func() {
					ui.Row(c).FillWidth().Gap(unit(c, 1)).Children(func() {
						for _, p := range parts {
							p()
						}
					})
				})
			}
			row(func() { cell("固定 2 轨", "#ececee", unit(c, 22)) },
				func() { cell("其余平分", "#e8effd", 0) },
				func() { cell("第三轨", "#e8f5e9", 0) })
			row(func() { cell("第二行 · 等宽", "#fdf1e6", 0) },
				func() { cell("", "#fdecec", 0) },
				func() { cell("", "#ececee", 0) })
		})
	})
}

// paneSection shows the two ways a screen splits — side by side and stacked
// — with a scroll area beside them, and a box that keeps its shape whatever
// the window does.
func paneSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "SplitPane — 两栏加一根可拖的把手")
	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		pane := func(sideBySide bool, name string) {
			ui.Column(c).Grow(1).Gap(unit(c, 1)).Children(func() {
				ui.Box(c).FillWidth().Height(unit(c, 34)).Radius(theme.ControlRadius).
					Border(theme.BorderWidth, k.Border).Clip().Children(func() {
					layout.SplitPane(c, layout.SplitPaneOptions{
						First: 0.32, SideBySide: sideBySide,
						FirstName: name + " 左栏", SecondName: name + " 右栏",
					}, func() {
						layout.Container(c, layout.ContainerOptions{
							Surface: true, Center: true, Gap: unit(c, 0.5),
						}, func() {
							ui.Text(c, "左栏").TextColor(k.Text).
								FontSize(core.FontSize(c, theme.RowSize)).Bold()
							ui.Text(c, "First 0.32").TextColor(k.TextMuted).
								FontSize(core.FontSize(c, theme.CaptionSize))
						}).Fill()
					}, func() {
						layout.Container(c, layout.ContainerOptions{
							Center: true, Gap: unit(c, 0.5),
						}, func() {
							ui.Text(c, "右栏").TextColor(k.Text).
								FontSize(core.FontSize(c, theme.RowSize)).Bold()
							ui.Text(c, "1 − First").TextColor(k.TextMuted).
								FontSize(core.FontSize(c, theme.CaptionSize))
						}).Fill()
					})
				})
				ui.Text(c, name).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize))
			})
		}
		pane(true, "SideBySide — 并排")
		pane(false, "上下 — 竖着放")

		ui.Column(c).Grow(1).Gap(unit(c, 1)).Children(func() {
			ui.Text(c, "ScrollArea — 给了高度才滚得动").TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize))
			ui.Box(c).Height(unit(c, 28)).Radius(theme.ControlRadius).
				Border(theme.BorderWidth, k.Border).Clip().Children(func() {
				layout.ScrollArea(c, layout.ScrollAreaOptions{
					Vertical: true, Height: unit(c, 28), Pad: unit(c, 1),
				}, func() {
					ui.Column(c).FillWidth().Gap(unit(c, 1)).Children(func() {
						for range 10 {
							ui.Box(c).FillWidth().Height(unit(c, 4)).
								Radius(theme.SmallRadius).Background(k.Border)
						}
					})
				})
			})
			ui.Text(c, "内容比视口高；滚动条悬停时出现（overlay 式）").TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize))
		})
	})

	showcase.Field(c, "AspectRatio — 缩略图的形状不能被窗口改掉")
	ui.Row(c).Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		for _, r := range []float32{1, 1.6, 0.8} {
			ui.Box(c).Width(unit(c, 22)).Children(func() {
				layout.AspectRatio(c, layout.AspectRatioOptions{Ratio: r}, func() {
					ui.Box(c).Fill().Radius(theme.SmallRadius).Background(k.Surface).
						Border(theme.BorderWidth, k.Border).Center().Children(func() {
						ui.Text(c, ratioName(r)).TextColor(k.TextMuted).
							FontSize(core.FontSize(c, theme.CaptionSize))
					})
				})
			})
		}
	})
}

// barsSection shows the two bars a desktop window wears, plus the traffic
// lights they share with the OS.
func barsSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "TitleBar 与 StatusBar — 窗口的上下两条")
	ui.Column(c).FillWidth().Gap(unit(c, 3)).Children(func() {
		layout.TitleBar(c, layout.TitleBarOptions{
			Leading: func() { layout.TrafficLights(c) },
			Title: func() {
				ui.Text(c, "Riverside Clinic").TextColor(k.Text).FontSize(core.FontSize(c, theme.RowSize)).Bold()
			},
			Center: func() {
				ui.Text(c, "第 3 层窗口").TextColor(k.TextFaint).FontSize(core.FontSize(c, theme.CaptionSize))
			},
			Trailing: func() {
				ui.Text(c, "◀︎ ▶︎").TextColor(k.TextMuted).FontSize(core.FontSize(c, theme.CaptionSize))
			},
		})
		layout.Container(c, layout.ContainerOptions{Width: unit(c, 40), Height: unit(c, 12), Border: true, Center: true}, func() {
			layout.TrafficLights(c).Label("TrafficLights")
		})

		layout.StatusBar(c, layout.StatusBarOptions{Border: true, Leading: func() {
			ui.Text(c, "● 27 未处理").TextColor(k.Text).FontSize(core.FontSize(c, theme.CaptionSize))
		}, Trailing: func() {
			ui.Text(c, "已连接 · 14:30").TextColor(k.TextMuted).FontSize(core.FontSize(c, theme.CaptionSize))
		}})
	})
}

// panelSection shows the one floating surface in the library, which every
// dialog, drawer, popover and tip wears.
func panelSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "Panel — 所有浮层共用的那一张皮")
	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		host := func() *ui.Element { return ui.Box(c) }
		layout.Panel(c, host(), layout.PanelOptions{
			Title: "Log callback", Subtitle: "记一条回访", Rule: true, MaxWidth: unit(c, 60),
		}, func() {
			ui.Text(c, "抽屉、对话框、气泡都从这里长出来").TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.MetaSize))
		})
		layout.Panel(c, host(), layout.PanelOptions{
			Title: "紧凑面板", Compact: true, MaxWidth: unit(c, 60),
		}, func() {
			ui.Text(c, "气泡、提示、确认框穿的是这件").TextColor(k.Text).
				FontSize(core.FontSize(c, theme.MetaSize))
		})
		layout.Panel(c, host(), layout.PanelOptions{
			Compact: true, MaxWidth: unit(c, 60),
		}, func() {
			ui.Text(c, "无标题的裸面板").TextColor(k.Text).
				FontSize(core.FontSize(c, theme.MetaSize))
		})
	})
}

// headerSection shows PageHeader with everything it can carry.
func headerSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "PageHeader — 面包屑 / 大标题 / 统计 / 人 / 工具")
	ui.Box(c).FillWidth().Radius(theme.ControlRadius).Background(k.Background).
		Border(theme.BorderWidth, k.Border).Children(func() {
		layout.PageHeader(c, layout.HeaderOptions{
			Crumbs:      []string{"Callbacks", "未处理"},
			Title:       "Open callbacks",
			Meta:        "27 未处理 · 4 逾期 · 今天已解决 12",
			AvatarSlots: []string{"Andre Thomson", "Mia Chen", "Ravi Patel", "Lena Ford", "Sam Ortiz"},
			SlotCount:   9,
			Tools: func() {
				ui.Box(c).Padding(unit(c, 0.75), unit(c, 2.25)).
					Radius(theme.PillRadius).Background(k.Border).Children(func() {
					ui.Text(c, "按分支分组").TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.CaptionSize))
				})
				ui.Box(c).Padding(unit(c, 1.25), unit(c, 2.5)).
					Radius(theme.PillRadius).Background(k.Fill).Children(func() {
					ui.Text(c, "记一条回访").TextColor(k.OnFill).
						FontSize(core.FontSize(c, theme.CaptionSize)).Bold()
				})
			},
		})
	})
}

// boardSection shows the board of lanes: one with cards, one nearly full, and
// one empty — the empty case is the reason Column takes an Empty slot.
func boardSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "Board 与 Column — 列宽是常量，窄了就横向滚")
	ui.Box(c).FillWidth().Height(unit(c, 82)).Radius(theme.ControlRadius).
		Border(theme.BorderWidth, k.Border).Children(func() {
		layout.Board(c, func() {
			lane(c, "未处理", 12, "还有 5 条", func() {
				laneCard(c, "Riverside Clinic", "CB-2871 · AC repair", "Andre Thomson")
				laneCard(c, "Northgate Dental", "CB-2869 · Water leak", "Mia Chen")
				laneCard(c, "Harbour Cafe", "CB-2864 · Oven", "Ravi Patel")
			})
			lane(c, "进行中", 4, "", func() {
				laneCard(c, "Elm Street Gym", "CB-2858 · Boiler", "Lena Ford")
				laneCard(c, "Pinewood Dental", "CB-2851 · Thermostat", "Sam Ortiz")
			})
			lane(c, "已解决", 0, "", func() {
				layout.Container(c, layout.ContainerOptions{
					Center: true, Gap: unit(c, 1), Pad: unit(c, 2),
				}, func() {
					ui.Box(c).Size(unit(c, 11), unit(c, 11)).Radius(unit(c, 4)).
						Background(k.Surface).Center().Children(func() {
						ui.Text(c, "○").TextColor(k.TextFaint).
							FontSize(core.FontSize(c, theme.BodySize))
					})
					ui.Text(c, "这一列是空的").TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.CaptionSize))
				}).Fill()
			})
		})
	})
}

// shellSection puts everything above into one window: sidebar, main, status
// bar. This is the arrangement every other page on this gallery is a piece
// of.
func shellSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "AppShell — 侧栏 + 主区 + 状态栏，侧栏宽是定数")
	var collapsed bool
	ui.Box(c).FillWidth().Height(unit(c, 80)).Radius(theme.ControlRadius).
		Border(theme.BorderWidth, k.Border).Clip().Children(func() {
		layout.AppShell(c, &collapsed, layout.AppShellOptions{
			SidebarWidth: theme.RailWidth + unit(c, 40),
			TitleBar: func() {
				layout.TitleBar(c, layout.TitleBarOptions{
					Title: func() {
						ui.Text(c, "Callbacks").TextColor(k.Text).
							FontSize(core.FontSize(c, theme.RowSize)).Bold()
					},
				})
			},
		}, func() {
			ui.Column(c).FillHeight().Padding(unit(c, 2)).Gap(unit(c, 1)).Children(func() {
				ui.Text(c, "视图").TextColor(k.TextFaint).FontSize(core.FontSize(c, theme.CaptionSize))
				for i, s := range []string{"我的回访", "团队回访", "已解决", "全部"} {
					ui.Box(c).FillWidth().Padding(unit(c, 1), unit(c, 1.5)).
						Radius(theme.SmallRadius).Label(s).Children(func() {
						if i == 0 {
							ui.Text(c, s).TextColor(k.Text).
								FontSize(core.FontSize(c, theme.RowSize)).Bold()
						} else {
							ui.Text(c, s).TextColor(k.TextMuted).
								FontSize(core.FontSize(c, theme.RowSize))
						}
					})
				}
			})
		}, func() {
			ui.Column(c).FillHeight().Padding(unit(c, 2)).Gap(unit(c, 2)).Children(func() {
				layout.PageHeader(c, layout.HeaderOptions{
					Crumbs: []string{"Callbacks", "我的回访"},
					Title:  "Open callbacks",
					Meta:   "27 未处理 · 4 逾期",
				})
				ui.Box(c).FillWidth().Height(unit(c, 26)).Radius(theme.ControlRadius).
					Background(k.Surface).Children(func() {
					ui.Column(c).FillWidth().Padding(unit(c, 1.5)).Gap(unit(c, 0.75)).Children(func() {
						for i, s := range []string{"Riverside Clinic", "Northgate Dental"} {
							ui.Text(c, s).TextColor(k.Text).
								FontSize(core.FontSize(c, theme.BodySize)).SingleLine()
							ui.Text(c, metaLine(i)).TextColor(k.TextMuted).
								FontSize(core.FontSize(c, theme.MetaSize)).SingleLine()
						}
					})
				})
			})
		}, func() {
			layout.StatusBar(c, layout.StatusBarOptions{
				Leading: func() {
					ui.Text(c, "27 未处理").TextColor(k.Text).
						FontSize(core.FontSize(c, theme.CaptionSize))
				},
				Trailing: func() {
					ui.Text(c, "2026-10-07 14:30").TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.CaptionSize))
				},
			})
		})
	})
}

// lane is one Column of the board sample.
func lane(c *ui.Context, title string, count int, more string, body func()) {
	layout.Column(c, layout.ColumnOptions[string]{
		Title: title, Count: &count, Menu: title + " 更多", More: more,
	}, body)
}

// laneCard is the smallest thing a lane holds, drawn with the same three
// lines a real card has.
func laneCard(c *ui.Context, name, meta, tech string) {
	k := core.Tokens(c)
	layout.Container(c, layout.ContainerOptions{
		Pad: unit(c, 1.75), Radius: theme.CardRadius, Border: true, Gap: unit(c, 0.5),
	}, func() {
		ui.Box(c).AlignSelf(ui.Start).Padding(unit(c, 0.25), unit(c, 1.5)).
			Radius(theme.PillRadius).Background(k.Border).Children(func() {
			ui.Text(c, "未处理").TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize))
		})
		ui.Text(c, name).TextColor(k.Text).FontSize(core.FontSize(c, theme.BodySize)).Bold().SingleLine()
		ui.Text(c, meta).TextColor(k.TextMuted).FontSize(core.FontSize(c, theme.MetaSize)).SingleLine()
		ui.Text(c, tech).TextColor(k.TextFaint).FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
	})
}

func metaLine(i int) string {
	return []string{"CB-2871 · AC repair", "CB-2869 · Water leak", "CB-2864 · Oven"}[i]
}

// cellName names one of the two sample grid rows, so a test and a reader can
// both say which row is on screen.
func cellName(i int) string { return "A" + strconv.Itoa(i+1) }

func ratioName(r float32) string {
	switch r {
	case 1.6:
		return "16:10"
	case 0.8:
		return "4:5"
	}
	return "1:1"
}
