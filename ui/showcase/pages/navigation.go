package pages

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/navigation"
	"github.com/HycJack/MintUI/ui/showcase"
	"github.com/HycJack/MintUI/ui/theme"
)

// Demo state outlives the frame: every demo below hands its component
// a pointer, and a pointer into a frame-local is a click the next
// frame undoes — a tab that will not switch, a dropdown that snaps
// shut, a slider that springs back.
var (
	page = 7
)

// The page's open-state lives for the whole process, not for one frame.
// A frame-local bool is rebuilt every frame, and an open layer closes itself
// on a press outside — a press it consumes — so NavigationMenu and the
// command palette would reopen under every click: each click anywhere eaten,
// the layer back again, the gallery never taking another one. The same state
// is what the sidebar groups' own toggles write into.
var (
	navMenuOpen      = true
	paletteOpen      = true
	paletteQuery     = "re"
	paletteHighlight = 0
	navViewsOpen     = true
	navBranchOpen    = false
)

func init() {
	showcase.Register(showcase.Page{
		Package: "navigation",
		Title:   "ui/navigation — 在窗口里移动",
		Note:    "图标栏、侧栏、筛选行、分组、页签、面包屑、步骤条、分页、工具栏、菜单、命令面板",
		Width:   1000,
		Height:  2060,
		Want: []string{
			"总览",
			"回访",
			"告警",
			"设置",
			"Callbacks",
			"我的回访",
			"团队回访",
			"Repeat failures",
			"North（子项）",
			"Views",
			"Branches",
			"看板",
			"列表",
			"日历",
			"North",
			"CB-2871",
			"接单",
			"到场",
			"报价",
			"上一页",
			"下一页",
			"记一条回访",
			"文件",
			"窗口",
			"按分支分组",
			"Resolve a callback",
			"Reassign to a technician",
			"Refresh the board",
			"27 未处理 · 4 逾期 · 今天已解决 12",
		},
		Render: func(c *ui.Context) {
			navigationPage(c)
		},
	})
}

// navigationPage draws the chrome that moves a person around an application.
// It is grouped by how far it moves you: the rail and the sidebar change
// where you are, the tabs change which view of that place you are looking
// at, and the filter rows and the group they live in change what is in it.
func navigationPage(c *ui.Context) {
	placeSection(c)
	listSection(c)
	// The palette is opened before the rest of the chrome is drawn, because
	// it is centred on the window: whatever band it lands on is the band it
	// covers, and a window standing in for a real one is the only thing on
	// this page worth covering.
	paletteSection(c)
	viewSection(c)
	barSection(c)
	menuSection(c)
}

// placeSection shows the two things that say where you are: the rail down
// the edge and the sidebar beside it.
func placeSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "Rail 与 Sidebar — 我在哪")
	ui.Row(c).FillWidth().Gap(unit(c, 5)).AlignItems(ui.Start).Children(func() {
		// Both of these are tall by nature, so each is given a frame with a
		// height rather than being allowed to claim the page.
		ui.Box(c).Height(unit(c, 62)).Radius(theme.ControlRadius).
			Border(theme.BorderWidth, k.Border).Clip().Children(func() {
			navigation.Rail(c, "Andre Thomson", []navigation.RailItem{
				{Name: "总览", Icon: glyphFor(c, display.IconOverview), Selected: true},
				{Name: "回访", Icon: glyphFor(c, display.IconCallbacks), Badge: true},
				{Name: "客户", Icon: glyphFor(c, display.IconCustomers)},
				{Name: "团队", Icon: glyphFor(c, display.IconTeam)},
			}, []navigation.RailItem{
				{Name: "告警", Icon: glyphFor(c, display.IconBell), Badge: true},
				{Name: "设置", Icon: glyphFor(c, display.IconSettings)},
			})
		})

		ui.Box(c).Width(unit(c, 58)).Height(unit(c, 62)).Radius(theme.ControlRadius).
			Border(theme.BorderWidth, k.Border).Clip().Children(func() {
			open := 12
			failures := 3
			navigation.Sidebar(c, []navigation.SidebarSection{
				{Title: "视图", Items: []navigation.SidebarItem{
					{Label: "我的回访", Icon: display.IconProfile, Count: &open, Selected: true},
					{Label: "团队回访", Icon: display.IconTeam, Count: &open},
					{Label: "Repeat failures", Icon: display.IconWarning, Count: &failures, Tone: core.Danger},
				}},
				{Title: "视图下的分支", Items: []navigation.SidebarItem{
					{Label: "North", Icon: display.IconPanel, Count: &open},
					{Label: "Harbour", Icon: display.IconPanel, Count: &open},
					{Label: "Closed", Icon: display.IconPanel, Disabled: true},
				}},
			}, navigation.SidebarOptions{
				Title: "Callbacks", Subtitle: "今天 14:30", Label: "主导航",
				Width: unit(c, 58),
			})
		})
	})
}

// listSection shows the filter rows a sidebar is made of and the collapsible
// groups they collect into.
func listSection(c *ui.Context) {
	k := core.Tokens(c)
	open := showcase.State(c, "navigation.145.open", 27)
	high := showcase.State(c, "navigation.145.high", 4)
	failures := showcase.State(c, "navigation.145.failures", 3)
	showcase.Section(c, "FilterRow 与 Group — 灰底上的筛选行，只有选中那行有底色")
	ui.Row(c).FillWidth().Gap(unit(c, 5)).AlignItems(ui.Start).Children(func() {
		ui.Box(c).Width(unit(c, 46)).Radius(theme.ControlRadius).Background(k.Surface).
			Padding(unit(c, 1.5)).Children(func() {
			ui.Column(c).FillWidth().Gap(unit(c, 0.5)).Children(func() {
				navigation.FilterRow(c, "未处理", navigation.FilterRowOptions{
					Count: open, Selected: true,
				})
				navigation.FilterRow(c, "逾期", navigation.FilterRowOptions{
					Count: high, Severity: core.Warning,
				})
				navigation.FilterRow(c, "Repeat failures", navigation.FilterRowOptions{
					Count: failures, Severity: core.Danger,
				})
				navigation.FilterRow(c, "已解决", navigation.FilterRowOptions{})
				navigation.FilterRow(c, "North（子项）", navigation.FilterRowOptions{
					Count: open, Indented: true,
				})
			})
		})

		ui.Box(c).Width(unit(c, 52)).Children(func() {
			g := navigation.Group(c, navigation.GroupOptions{
				Title: "Views", Icon: glyphFor(c, display.IconPanel), Open: navViewsOpen,
			}, func() {
				ui.Column(c).FillWidth().Gap(unit(c, 0.5)).Children(func() {
					navigation.FilterRow(c, "我的回访", navigation.FilterRowOptions{
						Count: open, Selected: true,
					})
					navigation.FilterRow(c, "团队回访", navigation.FilterRowOptions{Count: open})
				})
			})
			if g.Toggled() {
				navViewsOpen = !navViewsOpen
			}
			g2 := navigation.Group(c, navigation.GroupOptions{
				Title: "Branches", Icon: glyphFor(c, display.IconPanel), Open: navBranchOpen,
				Indent: true,
			}, func() {
				ui.Column(c).FillWidth().Gap(unit(c, 0.5)).Children(func() {
					navigation.FilterRow(c, "North", navigation.FilterRowOptions{
						Count: open, Indented: true,
					})
				})
			})
			if g2.Toggled() {
				navBranchOpen = !navBranchOpen
			}
			ui.Text(c, "折叠状态在调用方的 Open 里，组件自己不存").TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize))
		})
	})
}

// viewSection shows the strip that chooses which view of a place you are
// looking at, and the three ways of saying where in it you are.
func viewSection(c *ui.Context) {
	k := core.Tokens(c)
	sel := showcase.State(c, "navigation.204.sel", 0)
	compact := showcase.State(c, "navigation.204.compact", 1)
	showcase.Section(c, "Tabs — 一排互斥的视图")
	ui.Box(c).FillWidth().Radius(theme.ControlRadius).Padding(unit(c, 1.5)).Children(func() {
		ui.Column(c).FillWidth().Gap(unit(c, 2)).Children(func() {
			navigation.Tabs(c, []navigation.Tab{
				{Label: "看板", Icon: glyphFor(c, display.IconOverview)},
				{Label: "列表", Icon: glyphFor(c, display.IconSliders), Badge: "3"},
				{Label: "日历", Icon: glyphFor(c, display.IconCalendar)},
				{Label: "已归档", Disabled: true},
			}, navigation.TabsOptions{Selected: sel, Grow: true})
			navigation.Tabs(c, []navigation.Tab{
				{Label: "紧凑"}, {Label: "正常"}, {Label: "宽松"},
			}, navigation.TabsOptions{Selected: compact, Vertical: true})
		})
	})
	_ = k

	showcase.Field(c, "Breadcrumb — 你在哪，下面还有多深")
	ui.Column(c).FillWidth().Gap(unit(c, 1)).Children(func() {
		navigation.Breadcrumb(c, []navigation.BreadcrumbItem{
			{Label: "Callbacks"}, {Label: "North"}, {Label: "CB-2871", Current: true},
		})
		navigation.Breadcrumb(c, []navigation.BreadcrumbItem{
			{Label: "Callbacks"}, {Label: "North", Current: true},
		})
	})

	showcase.Field(c, "Steps — 事情走到了第几步")
	ui.Column(c).FillWidth().Gap(unit(c, 2)).Children(func() {
		navigation.Steps(c, []navigation.Step{
			{Label: "接单", Done: true},
			{Label: "到场", Current: true},
			{Label: "报价"},
			{Label: "结算", Disabled: true},
		}, navigation.StepsOptions{Numbers: true})
		ui.Box(c).Width(unit(c, 40)).Children(func() {
			navigation.Steps(c, []navigation.Step{
				{Label: "新建", Done: true},
				{Label: "编辑", Current: true},
				{Label: "发布"},
			}, navigation.StepsOptions{Vertical: true, Numbers: true})
		})
	})

	showcase.Field(c, "Pagination — 一页一页地走")
	ui.Column(c).FillWidth().Gap(unit(c, 1)).Children(func() {
		navigation.Pagination(c, navigation.PaginationOptions{
			Page: &page, Pages: 24, Window: 5, Prev: "上一页", Next: "下一页",
		})
		navigation.Pagination(c, navigation.PaginationOptions{
			Page: &page, Pages: 3, Window: 5, Prev: "上一页", Next: "下一页",
		})
	})
}

// barSection shows the bar of buttons a window wears under its title, and
// the menu bar of a desktop application.
func barSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "Toolbar — 窗口里那一排动作")
	ui.Box(c).FillWidth().Radius(theme.ControlRadius).Background(k.Surface).
		Padding(unit(c, 1)).Children(func() {
		navigation.Toolbar(c, navigation.ToolbarOptions{Label: "主工具栏"}, func() {
			ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(unit(c, 1)).
				Padding(unit(c, 0), unit(c, 1)).Children(func() {
				navigation.ToolbarButton(c, "看板", navigation.ToolbarButtonOptions{
					Icon: display.IconOverview, Active: true,
				})
				navigation.ToolbarButton(c, "列表", navigation.ToolbarButtonOptions{
					Icon: display.IconSliders,
				})
				navigation.ToolbarSeparator(c)
				navigation.ToolbarGroup(c, navigation.ToolbarGroupOptions{Label: "危险动作"}, func() {
					ui.Row(c).Gap(unit(c, 1)).Children(func() {
						navigation.ToolbarButton(c, "删除", navigation.ToolbarButtonOptions{
							Icon: display.IconTrash, Danger: true,
						})
						navigation.ToolbarButton(c, "刷新", navigation.ToolbarButtonOptions{
							Icon: display.IconRefresh, Disabled: true,
						})
					})
				})
				ui.Box(c).Grow(1)
				navigation.ToolbarButton(c, "记一条回访", navigation.ToolbarButtonOptions{
					Icon: display.IconPlus, Primary: true,
				})
			})
		})
	})

	showcase.Field(c, "Menubar — 桌面应用自己的那一行")
	ui.Box(c).FillWidth().Radius(theme.ControlRadius).Border(theme.BorderWidth, k.Border).
		Children(func() {
			navigation.Menubar(c, "Callbacks", []navigation.Menu{
				{Label: "文件", Items: []navigation.MenuItem{
					{Label: "新建回访"},
					{Label: "打开…"},
					{Separator: true},
					{Label: "导出", Disabled: true},
				}},
				{Label: "编辑", Items: []navigation.MenuItem{
					{Label: "撤销"}, {Label: "重做", Disabled: true},
				}},
				{Label: "窗口", Items: []navigation.MenuItem{
					{Label: "最小化"}, {Label: "缩合侧栏"},
				}},
			}, navigation.MenubarOptions{Label: "菜单栏"})
		})
}

// menuSection shows the navigation menu: a trigger and a list of places.
func menuSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "NavigationMenu — 一个触发器和它要去的地方")
	ui.Box(c).FillWidth().Radius(theme.ControlRadius).Background(k.Surface).
		Padding(unit(c, 2)).Children(func() {
		navigation.NavigationMenu(c, "按分支分组", []navigation.NavigationMenuItem{
			{Label: "不分组", Description: "一列走到底", Selected: true},
			{Label: "按分支", Description: "每条分支一列", Icon: glyphFor(c, display.IconPanel)},
			{Label: "按技师", Description: "谁的手里就归谁", Icon: glyphFor(c, display.IconTeam)},
			{Label: "按优先级", Description: "还没接", Disabled: true},
		}, navigation.NavigationMenuOptions{Open: &navMenuOpen, Label: "分组方式", MaxWidth: unit(c, 60)})
	})
}

// paletteSection is last on the page on purpose, and says so. The command
// palette is a modal dialog of the whole window, so wherever it is opened it
// covers this page and lays its scrim over every section above it. The grey
// over the rest of the gallery is the palette's own scrim, not a fault in the
// page — which is worth a reader seeing once rather than only in the app.
func paletteSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "CommandPalette — 整窗的模态，它浮在下面那个窗口上")
	// The palette is centred on the window, so the band under it is a whole
	// application screen rather than another two components: a scrim over
	// nothing is a grey page, and a scrim over an app is the thing a reader
	// needs to see once.
	ui.Box(c).FillWidth().Height(unit(c, 150)).Radius(theme.ControlRadius).
		Background(k.Surface).Border(theme.BorderWidth, k.Border).Clip().Children(func() {
		ui.Column(c).FillHeight().Padding(unit(c, 2)).Gap(unit(c, 2)).Children(func() {
			ui.Text(c, "Callbacks").TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).Bold()
			ui.Text(c, "27 未处理 · 4 逾期 · 今天已解决 12").TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize))
			ui.Box(c).FillWidth().Grow(1).Radius(theme.CardRadius).
				Background(k.Background).Padding(unit(c, 1.5)).Children(func() {
				ui.Column(c).FillWidth().Gap(unit(c, 1)).Children(func() {
					for _, r := range []string{
						"Riverside Clinic · CB-2871 · AC repair",
						"Northgate Dental · CB-2869 · Water leak",
						"Harbour Cafe · CB-2864 · Oven",
						"Elm Street Gym · CB-2858 · Boiler",
						"Pinewood Dental · CB-2851 · Thermostat",
						"Westside Vet · CB-2844 · Thermostat",
						"Lakeside Inn · CB-2839 · Fridge",
					} {
						ui.Text(c, r).TextColor(k.Text).
							FontSize(core.FontSize(c, theme.MetaSize)).SingleLine()
					}
				})
			})
		})
	})

	// The palette's open-state is the process's own: a frame-local bool here
	// was rebuilt true every frame, and since an open palette consumes the
	// press that closes it, the gallery took no click after this section.
	navigation.CommandPalette(c, &paletteOpen, []navigation.Command{
		{ID: "resolve", Title: "Resolve a callback", Subtitle: "把选中的标成已解决",
			Shortcut: []string{"⌘", "R"}, Keywords: []string{"done", "完成"}},
		{ID: "reassign", Title: "Reassign to a technician", Subtitle: "换一个人接",
			Keywords: []string{"指派", "assign"}},
		{ID: "refresh", Title: "Refresh the board", Subtitle: "重新拉一遍",
			Keywords: []string{"同步"}},
		{ID: "export", Title: "Export the list view", Subtitle: "导出 CSV",
			Disabled: true},
	}, navigation.CommandPaletteOptions{
		Query: &paletteQuery, Highlight: &paletteHighlight, Rows: 4, Width: unit(c, 72),
		Placeholder: "搜索命令", Label: "命令面板",
		// Undimmmed here so one frame shows the palette and the rail and
		// tabs it was opened over. In an app the scrim is right.
		NonModal: true,
	})
}

// railGlyphs are the marks this page's rail, tabs and menus draw. The
// library keeps its own icon shapes private and hands them out as drawn
// elements, while Rail, Tabs, Group, Toolbar and NavigationMenu all take a
// raw *ui.SVG — so the gallery carries the few shapes it needs in the same
// 24×24 stroke style display.Icon uses, and a mark on the rail is the same
// hand as a mark in an icon grid on the display page.
var railGlyphs = map[display.IconName]string{
	display.IconOverview: `<rect x="3.5" y="3.5" width="6.5" height="6.5" rx="2"/>` +
		`<rect x="14" y="3.5" width="6.5" height="6.5" rx="2"/>` +
		`<rect x="3.5" y="14" width="6.5" height="6.5" rx="2"/>` +
		`<rect x="14" y="14" width="6.5" height="6.5" rx="2"/>`,
	display.IconCallbacks: `<path d="M6.5 3.5h3l1.5 3-1.8 1.6a11.5 11.5 0 0 0 5.2 5.2l1.6-1.8 3 1.5v3a2 2 0 0 1-2.2 2A15.5 15.5 0 0 1 4.5 5.7 2 2 0 0 1 6.5 3.5Z"/>`,
	display.IconCustomers: `<circle cx="12" cy="6" r="2.6"/><circle cx="6" cy="17" r="2.6"/>` +
		`<circle cx="18" cy="17" r="2.6"/><path d="M12 8.6v4.2M10 13.4 8 14.9M14 13.4l2 1.5"/>`,
	display.IconTeam: `<circle cx="9" cy="8.5" r="3"/>` +
		`<path d="M3.5 19.5c.6-3.2 2.8-4.8 5.5-4.8s4.9 1.6 5.5 4.8"/>` +
		`<path d="M16 6.2a3 3 0 0 1 0 5.6M17.5 15.2c2 .7 3.2 2.2 3.6 4.3"/>`,
	display.IconBell: `<path d="M6.5 17.5V11a5.5 5.5 0 0 1 11 0v6.5l1.6 2H4.9l1.6-2Z"/>` +
		`<path d="M10 20a2.2 2.2 0 0 0 4 0"/>`,
	display.IconSettings: `<circle cx="12" cy="12" r="3"/>` +
		`<path d="M12 3.5v2.2M12 18.3v2.2M20.5 12h-2.2M5.7 12H3.5M18 6l-1.6 1.6M7.6 16.4 6 18M18 18l-1.6-1.6M7.6 7.6 6 6"/>`,
	display.IconPanel: `<rect x="3.5" y="4.5" width="17" height="15" rx="3"/>` +
		`<path d="M9.5 4.5v15"/><path d="m16 10-2.5 2 2.5 2"/>`,
	display.IconSliders: `<path d="M4 6h9M17 6h3M4 12h3M11 12h9M4 18h7M15 18h5"/>` +
		`<circle cx="15" cy="6" r="2"/><circle cx="9" cy="12" r="2"/><circle cx="13" cy="18" r="2"/>`,
	display.IconCalendar: `<rect x="3.5" y="5.5" width="17" height="15" rx="3"/>` +
		`<path d="M3.5 10h17M8 3.5v4M16 3.5v4"/>`,
	display.IconTrash:   `<path d="M5.5 7h13M10 7V5h4v2M7 7l1 12.5h8L17 7"/>`,
	display.IconRefresh: `<path d="M19 12a7 7 0 1 1-2.2-5.1"/><path d="M19.5 4v4h-4"/>`,
	display.IconPlus:    `<path d="M12 5.5v13M5.5 12h13"/>`,
	display.IconWarning: `<path d="M12 4.5 21 20H3Z"/><path d="M12 10v4.5M12 17.2v.1"/>`,
}

// glyphFor is the SVG of one of the marks above, in the same 24×24 stroke
// style the library draws its own icons in. An unknown name gets the check
// mark rather than nothing, so a typo on this page is visible.
func glyphFor(c *ui.Context, name display.IconName) *ui.SVG {
	shapes, ok := railGlyphs[name]
	if !ok {
		shapes = `<path d="m5 12.5 4.5 4.5L19 7"/>`
	}
	return ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" ` +
		`fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" ` +
		`stroke-linejoin="round">` + shapes + `</svg>`))
}
