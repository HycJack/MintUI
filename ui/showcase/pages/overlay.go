package pages

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/overlay"
	"github.com/HycJack/MintUI/ui/showcase"
	"github.com/HycJack/MintUI/ui/theme"
)

// The layers on this page keep their open-state for the life of the process.
// A frame-local bool is rebuilt true every frame, and an anchored layer
// closes itself on a press outside — a press it consumes — so it would be
// rebuilt under every click: each click anywhere eaten, the layer back again,
// the gallery never taking another one. Anchored controls reopen a layer the
// reader closed on purpose. The three that own the window need it for the
// other reason: a close has to outlive the frame that made it, or a
// frame-local bool would put the layer straight back.
var (
	overlayPopoverOpen = true
	confirmOpen        = true
	roomOpen           = true
	dialogOpen         = true
	alertOpen          = true
	drawerOpen         = true
)

func init() {
	showcase.Register(showcase.Page{
		Package: "overlay",
		Title:   "ui/overlay — 浮在上面的那一层",
		Note:    "对话框、抽屉、气泡、悬停卡、菜单、拆分按钮、提示、面板",
		Width:   1000,
		Height:  1160,
		// Anchored: this page has a selection bar along the bottom and a
		// drawer down the right, both pinned to the window, so the page
		// is the only thing that knows how tall it is.
		Anchored: true,
		Want: []string{
			"Log callback",
			"记一条回访",
			"紧凑面板",
			"一张没有标题的皮",
			"按分支分组",
			"这一块是气泡的内容。",
			"删除回访",
			"删除 CB-2871？",
			"删掉之后不在任何一列里。",
			"Andre Thomson",
			"North 分支 · 本周 14 单",
			"导出 CSV",
			"更多",
			"主要的下拉",
			"禁用的下拉",
			"保存为模板",
			"保存方式",
			"Riverside Clinic · CB-2871",
			"Delete “Northgate Dental”?",
			"这条回访会从三列里消失。",
			"Customer",
			"Issue",
			"Priority",
			"Branch",
			"Save",
			"Cancel",
			"Delete",
		},
		Render: func(c *ui.Context) {
			// The drawer hangs from the right edge of the window, wherever on
			// the page it is opened, so the page keeps a margin that width
			// for it: without one it covers the right of every section above
			// it, and a gallery with a column missing is worse than one with
			// a drawer drawn.
			ui.Box(c).Fill().Padding(0, 0, 0, unit(c, 34)).Children(func() {
				overlayPage(c)
			})
		},
	})
}

// overlayPage draws the library's floating layers. It is arranged by how
// those layers are placed, because that is the thing a reader cannot get
// from a list of names: a panel is in the flow, an anchored layer hangs off
// a control, and a layer with a scrim owns the whole window.
//
// HoverCard is the one layer here with no demo of its own, because a caller
// cannot hold one open and a screenshot has no pointer to hold it with;
// anchorSection says so where its slot would have been.
//
// The three that own the window — the dialog, the alert and the drawer — are
// at the bottom, over a band that stands for the window they would be opened
// over. A modal layer is drawn over the entire window whatever page it is
// built on, so anywhere else on a gallery page it would cover a stranger's
// section; at the bottom, over a mock window, the grey over everything else
// is the scrim doing what a scrim does.
func overlayPage(c *ui.Context) {
	overlayPanelSection(c)
	anchorSection(c)
	overlayMenuSection(c)
	modalSection(c)
}

// overlayPanelSection shows the one surface every other layer in this package wears.
// It is here, in the flow, because it is the only one of them that is.
func overlayPanelSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "Panel — 抽屉、对话框、气泡、提示、确认框都穿这一件")
	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		host := func() *ui.Element { return ui.Box(c) }
		overlay.Panel(c, host(), overlay.PanelOptions{
			Title: "Log callback", Subtitle: "记一条回访", Rule: true,
			MaxWidth: unit(c, 58), TitleSize: theme.SheetSize,
		}, func() {
			ui.Text(c, "抽屉 / 对话框：一块面板 + 自己的动作行").TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.MetaSize))
		})
		overlay.Panel(c, host(), overlay.PanelOptions{
			Title: "紧凑面板", Subtitle: "气泡、提示、确认框穿的是这件",
			Compact: true, MaxWidth: unit(c, 58),
		}, func() {
			ui.Text(c, "ControlRadius，更紧的边距").TextColor(k.Text).
				FontSize(core.FontSize(c, theme.MetaSize))
		})
		overlay.Panel(c, host(), overlay.PanelOptions{
			Compact: true, MaxWidth: unit(c, 58), Label: "一张没有标题的皮",
		}, func() {
			ui.Text(c, "无标题：不占地方，只包住内容").TextColor(k.Text).
				FontSize(core.FontSize(c, theme.MetaSize))
		})
	})
}

// anchorSection shows the three layers that hang off a control, each on its
// own stage: an anchored panel is positioned against the control it belongs
// to, so a page showing three at once has to give each one room rather than
// stack them in a row and let them cover one another.
func anchorSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "Popover、Popconfirm、PopoverRoom、Tooltip — 挂在控件上")
	stage := func(title, note string, build func()) {
		// Each stage is a frame with a fixed height and room at the bottom,
		// because the layer is drawn outside the row of stages and lands
		// under whatever control it is anchored to.
		ui.Text(c, note).TextColor(k.TextFaint).
			FontSize(core.FontSize(c, theme.CaptionSize)).MarginY(unit(c, 0.5))
		ui.Box(c).FillWidth().Height(unit(c, 30)).Radius(theme.ControlRadius).
			Background(k.Background).Border(theme.BorderWidth, k.Border).
			Clip().Children(func() {
			ui.Box(c).FillWidth().Padding(unit(c, 1)).Children(func() {
				ui.Text(c, title).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize))
			})
			build()
		})
	}

	stage("Popover", "挂在按钮下面，空间不够时自己翻到上面去；点外面关掉，再点按钮打开", func() {
		anchor := input.Button(c, "分组方式", input.ButtonOptions{})
		if anchor.Clicked() {
			overlayPopoverOpen = true
			c.Invalidate()
		}
		overlay.Popover(c, anchor, &overlayPopoverOpen, overlay.PopoverOptions{
			// NonModal because a modal popover dims the whole window, and a
			// gallery that is a quarter grey is a gallery nobody reads. The
			// panel is the same one either way.
			Modal:    false,
			Title:    "按分支分组",
			Subtitle: "每条分支一列",
			Body: func() {
				ui.Text(c, "这一块是气泡的内容。").TextColor(core.Tokens(c).TextMuted).
					FontSize(core.FontSize(c, theme.MetaSize))
			},
			MaxWidth: unit(c, 48),
		})
	})
	stage("Popconfirm", "一个问句，两个答案", func() {
		anchor := input.Button(c, "删除回访", input.ButtonOptions{Danger: true})
		if anchor.Clicked() {
			confirmOpen = true
			c.Invalidate()
		}
		overlay.Popconfirm(c, anchor, &confirmOpen, overlay.PopconfirmOptions{
			Title: "删除 CB-2871？", Body: "删掉之后不在任何一列里。",
			Confirm: "删除", Cancel: "留着", Destructive: true,
		})
	})
	stage("PopoverRoom", "一张解释名字的卡；同一块面板，只是宽一点的那面", func() {
		anchor := input.Button(c, "Andre Thomson", input.ButtonOptions{})
		if anchor.Clicked() {
			roomOpen = true
			c.Invalidate()
		}
		// A HoverCard is the card a name would open, and it is deliberately not
		// what is drawn here. ui/overlay/hovercard.go writes false into the
		// caller's *bool on every frame where neither the anchor nor the card
		// is under the pointer, so a caller has no way to pin one open and a
		// headless draw has no pointer at all: leaving the flag true would paint
		// a frame no window is ever in. It also went quietly wrong rather than
		// loudly — `gallery -shots` draws every page once to check it and then
		// again for the PNG, so the check's draw spent the flag and the exported
		// picture came out with no card on it. A PopoverRoom is the same panel
		// on the roomier surface, and it takes the flag without taking it back.
		// It also stands here rather than a second Popover because the stage
		// above is that one.
		overlay.PopoverRoom(c, anchor, &roomOpen, overlay.PopoverRoomOptions{
			Modal:    false,
			Title:    "Andre Thomson",
			Subtitle: "North 分支 · 本周 14 单",
			Body: func() {
				ui.Text(c, "上次在线：今天 09:12").TextColor(core.Tokens(c).TextMuted).
					FontSize(core.FontSize(c, theme.MetaSize))
			},
			MaxWidth: unit(c, 48),
		})
	})
	stage("Tooltip", "没有地方写下的话，写在指针下面", func() {
		// A tip only exists on the frames it is showing, so what this page
		// draws is the control it belongs to.
		anchor := input.Button(c, "导出 CSV", input.ButtonOptions{})
		overlay.Tooltip(c, anchor, "导出当前筛选下的全部回访")
	})
}

// overlayMenuSection shows the three controls that open a menu. The menu itself is
// the desktop's, so what this page shows is the trigger each one puts on a
// page and the shape of the run of controls.
func overlayMenuSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "DropdownMenu、ContextMenu、SplitButton — 打开菜单的三种控制")
	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).Width(unit(c, 40)).Gap(unit(c, 1.5)).Children(func() {
			ui.Text(c, "DropdownMenu — 一个带箭头的按钮").TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).Bold()
			showcase.Stack(c, 1, func() {
				overlay.DropdownMenu(c, "更多", overlay.DropdownMenuOptions{
					Build: menuItems(),
				})
				overlay.DropdownMenu(c, "主要的下拉", overlay.DropdownMenuOptions{
					Primary: true, Build: menuItems(),
				})
				overlay.DropdownMenu(c, "禁用的下拉", overlay.DropdownMenuOptions{
					Disabled: true, Build: menuItems(),
				})
			})
		})

		ui.Column(c).Width(unit(c, 42)).Gap(unit(c, 1.5)).Children(func() {
			ui.Text(c, "SplitButton — 左边直接做，右边才选").TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).Bold()
			showcase.Stack(c, 1, func() {
				overlay.SplitButton(c, "保存为模板", overlay.SplitButtonOptions{
					Label: "保存方式", Build: menuItems(),
				})
				overlay.SplitButton(c, "删除", overlay.SplitButtonOptions{
					Label: "删除方式", Danger: true, Build: menuItems(),
				})
			})
		})

		ui.Column(c).Grow(1).Gap(unit(c, 1.5)).Children(func() {
			ui.Text(c, "ContextMenu — 右键那一列").TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).Bold()
			ui.Box(c).FillWidth().Radius(theme.ControlRadius).Background(k.Surface).
				Padding(unit(c, 1)).Children(func() {
				for _, r := range []string{"Riverside Clinic", "Northgate Dental"} {
					row := ui.Box(c).FillWidth().Padding(unit(c, 0.75), unit(c, 1.5)).
						Radius(theme.SmallRadius).Background(k.Background).Label(r).
						Children(func() {
							ui.Text(c, r).TextColor(k.Text).
								FontSize(core.FontSize(c, theme.MetaSize))
						})
					overlay.ContextMenu(c, row, overlay.ContextMenuOptions{Build: menuItems()})
				}
			})
		})
	})
}

// modalSection is the last band: the layers that own the whole window. They
// are opened over a mock window so the scrim has something to be a scrim
// over, and the page says so rather than leaving a reader to wonder why
// everything above is grey.
func modalSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "Dialog、AlertDialog、Drawer — 整窗的浮层，盖住上面所有东西")
	ui.Box(c).FillWidth().Height(unit(c, 46)).Radius(theme.ControlRadius).
		Background(k.Surface).Border(theme.BorderWidth, k.Border).Clip().Children(func() {
		ui.Column(c).FillHeight().Padding(unit(c, 2)).Gap(unit(c, 1.5)).Children(func() {
			ui.Text(c, "Callbacks").TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).Bold()
			ui.Box(c).FillWidth().Grow(1).Radius(theme.CardRadius).
				Background(k.Background).Padding(unit(c, 1.5)).Children(func() {
				ui.Column(c).FillWidth().Gap(unit(c, 1)).Children(func() {
					for _, r := range []string{
						"Riverside Clinic · CB-2871",
						"Northgate Dental · CB-2869",
						"Harbour Cafe · CB-2864",
					} {
						ui.Text(c, r).TextColor(k.TextMuted).
							FontSize(core.FontSize(c, theme.MetaSize))
					}
				})
			})
		})
	})

	// Three layers that each want the window to themselves. They are built
	// NonModal, because a modal layer's scrim would lay a sheet of grey over
	// every other section of this gallery and the page would stop being
	// readable; the layers themselves are the same three either way.
	//
	// A dialog and an alert are both centred, so one is drawn over the
	// other — which is a state a real window is in whenever an alert is
	// raised from a dialog, and the only way both can be seen in one frame.
	//
	// Being NonModal is also what leaves each of them with no way out but
	// its own controls: layer() takes neither a press outside nor Escape
	// from a layer that does not dim the page, so the only thing that can
	// dismiss one of these is a control on it. Which is what each component
	// says it does. A Dialog and a Drawer report no result of their own —
	// their buttons report themselves and the caller writes false into the
	// *bool — so both buttons here put the layer away, there being nothing
	// on a gallery page for a Save to have saved. An AlertDialog hands its
	// answer back through Chosen, and either answer ends the question. So:
	// the buttons under the stage are how a reader brings one back, because
	// a close has to be something a control says.
	overlay.Dialog(c, &dialogOpen, overlay.DialogOptions{
		Title: "Log callback", Subtitle: "记一条回访", Width: unit(c, 74), Rule: true,
		NonModal: true,
		Body: func() {
			ui.Column(c).FillWidth().Gap(unit(c, 1)).Children(func() {
				ui.Text(c, "Customer").TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize))
				ui.Text(c, "Riverside Clinic").TextColor(k.Text).
					FontSize(core.FontSize(c, theme.BodySize))
				ui.Text(c, "Issue").TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize))
				ui.Text(c, "客诉机不制冷，上午报的修。").TextColor(k.Text).
					FontSize(core.FontSize(c, theme.BodySize))
			})
		},
		Actions: func() {
			showcase.Stack(c, 1, func() {
				if input.Button(c, "Cancel", input.ButtonOptions{}).Clicked() {
					dialogOpen = false
					c.Invalidate()
				}
				if input.Button(c, "Save", input.ButtonOptions{Primary: true}).Clicked() {
					dialogOpen = false
					c.Invalidate()
				}
			})
		},
	})
	if res := overlay.AlertDialog(c, &alertOpen, overlay.AlertDialogOptions{
		Title: "Delete “Northgate Dental”?", Body: "这条回访会从三列里消失。",
		Actions: []string{"Cancel", "Delete"}, Destructive: true, NonModal: true,
	}); res.Chosen() >= 0 {
		// The answer, not which answer: a gallery that deleted nothing has
		// nothing to do with Cancel and Delete but to stop asking.
		alertOpen = false
		c.Invalidate()
	}
	overlay.Drawer(c, &drawerOpen, overlay.DrawerOptions{
		Side: ui.End, Title: "Log callback", Subtitle: "抽屉挂在窗口的一条边上",
		Width: unit(c, 32), Rule: true, NonModal: true,
		Body: func() {
			ui.Column(c).FillWidth().Gap(unit(c, 1)).Children(func() {
				for _, s := range []string{"Customer", "Issue", "Priority", "Branch"} {
					ui.Text(c, s).TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.CaptionSize))
				}
			})
		},
		Actions: func() {
			if input.Button(c, "Save", input.ButtonOptions{Primary: true}).Clicked() {
				drawerOpen = false
				c.Invalidate()
			}
		},
	})
	ui.Row(c).FillWidth().Gap(unit(c, 2)).Children(func() {
		ui.Text(c, "用层里的按钮关掉它，再从这里打开一次：").TextColor(k.TextFaint).
			FontSize(core.FontSize(c, theme.CaptionSize))
		if input.Button(c, "Dialog", input.ButtonOptions{}).Clicked() {
			dialogOpen = true
			c.Invalidate()
		}
		if input.Button(c, "AlertDialog", input.ButtonOptions{}).Clicked() {
			alertOpen = true
			c.Invalidate()
		}
		if input.Button(c, "Drawer", input.ButtonOptions{}).Clicked() {
			drawerOpen = true
			c.Invalidate()
		}
	})
}

// menuItems is the menu every trigger on this page opens. The menu itself is
// the desktop's, so a gallery cannot draw it — but the items are what a
// reader wants to see named beside the trigger that opens them.
func menuItems() func(m *ui.Menu) {
	return func(m *ui.Menu) {
		m.Item("指派给技术员")
		m.Item("改到别的分支")
		m.Separator()
		m.Item("导出")
		m.Item("删除").Disabled(true)
	}
}
