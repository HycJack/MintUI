package pages

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/feedback"
	"github.com/HycJack/MintUI/ui/showcase"
	"github.com/HycJack/MintUI/ui/theme"
)

func init() {
	showcase.Register(showcase.Page{
		Package: "feedback",
		Title:   "ui/feedback — 应用说，而不是展示",
		Note:    "告警、结果、空态、Toast、通知中心、活动流、加载、占位、进度、状态、运动",
		Width:   1000,
		Height:  2260,
		Want: []string{
			"同步失败",
			"3 条回访没能送出去。",
			"网络较慢",
			"已切到离线",
			"全部已同步",
			"12 条已解决",
			"什么都没找到",
			"Nothing resolved yet today",
			"Resolve a callback",
			"标题",
			"副行",
			"金额",
			"一行正在来",
			"一段正在来的文字",
			"停止同步",
			"正在拉取 27 条回访…",
			"被盖住的是这一块。",
			"同步进度",
			"几乎完成",
			"卡住了",
			"已连接",
			"重试中",
			"已断开",
			"Andre",
			"Mia",
			"Ravi",
			"Lena",
			"Callback assigned",
			"Parts ordered",
			"Sync failed",
			"通知",
			"有人提到了你",
			"Riverside Clinic",
			"Northgate Dental",
			"Harbour Cafe",
			"Elm Street Gym",
			"已保存 3 条回访",
			"已归档 2 条",
			"撤销",
			"CB-2864 的金额刚被人改过",
		},
		Render: func(c *ui.Context) {
			feedbackPage(c)
		},
	})
}

// feedbackPage draws what the application has to say rather than what it
// shows: something went wrong, there is nothing here yet, it is still
// working, here is how far it got, and here is what just happened.
//
// Two of these live over the window rather than in the flow — the toast and
// the loading overlay — so they are on their own band at the bottom, the way
// BulkActionBar is on the input page and the command palette is on the
// navigation page: a floating layer covers the whole window, and a gallery
// it covers is a gallery nobody reads.
func feedbackPage(c *ui.Context) {
	saySection(c)
	absentSection(c)
	waitSection(c)
	progressSection(c)
	streamSection(c)
	floatSection(c)
}

// saySection shows the two answers a screen gives to a result: a banner that
// sits in the page, and the one that stands on its own.
func saySection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "Alert 与 Result — 出了什么事，或者成了什么事")
	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).Grow(1).Gap(unit(c, 2)).Children(func() {
			ui.Text(c, "Alert — 留在页面里").TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).Bold()
			for _, s := range []struct {
				title, body string
				sev         core.Severity
			}{
				{"同步失败", "3 条回访没能送出去。", core.Danger},
				{"网络较慢", "队列在积压。", core.Warning},
				{"已切到离线", "改动会等网络回来再送。", core.Accent},
				{"全部已同步", "12 条回访已上传。", core.Success},
			} {
				feedback.Alert(c, feedback.AlertOptions{
					Title: s.title, Body: s.body, Severity: s.sev,
					Action: "重试", Dismissable: true,
				})
			}
		})
		ui.Column(c).Grow(1).Gap(unit(c, 2)).Children(func() {
			ui.Text(c, "Result — 一件事的结局").TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).Bold()
			feedback.Result(c, feedback.ResultOptions{
				Title: "12 条已解决", Body: "今天上午解决的那一批。",
				Severity: core.Success, Action: "导出",
			})
			feedback.Result(c, feedback.ResultOptions{
				Title: "什么都没找到", Body: "换一个筛选条件再试一次。",
				Severity: core.Neutral,
			})
		})
	})
}

// absentSection shows the three ways of saying there is nothing here, and
// the two that wait for something to arrive.
func absentSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "Empty、Skeleton、Shimmer、ShimmerText — 「还没有」和「正在来」")
	// Empty grows to fill whatever it is given, so it gets a frame with a
	// height of its own: beside a taller neighbour it would spread its four
	// parts down the whole column and the action button would end up a
	// screen away from the title it belongs to.
	ui.Row(c).FillWidth().Gap(unit(c, 4)).AlignItems(ui.Start).Children(func() {
		ui.Box(c).Width(unit(c, 58)).Height(unit(c, 96)).Children(func() {
			feedback.Empty(c, feedback.EmptyOptions{
				Title: "Nothing resolved yet today", Body: "解决一条回访，它就会出现在这里。",
				Action: "Resolve a callback",
			})
		})
		ui.Column(c).Grow(1).Gap(unit(c, 2)).Children(func() {
			ui.Text(c, "Skeleton 与 Shimmer — 还没到的内容").TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).Bold()
			ui.Box(c).FillWidth().Padding(unit(c, 1.5)).Radius(theme.ControlRadius).
				Background(k.Surface).Children(func() {
				ui.Column(c).FillWidth().Gap(unit(c, 1)).Children(func() {
					for i, w := range []float32{0.9, 0.7, 0.5} {
						feedback.Skeleton(c, feedback.SkeletonOptions{
							Width: unit(c, 40) * w, Height: unit(c, 3),
							Radius: theme.SmallRadius, Label: skeletonName(i),
						})
					}
				})
			})
			ui.Box(c).FillWidth().Padding(unit(c, 1.5)).Radius(theme.ControlRadius).
				Background(k.Surface).Children(func() {
				ui.Column(c).FillWidth().Gap(unit(c, 1)).Children(func() {
					feedback.Shimmer(c, feedback.ShimmerOptions{
						Width: unit(c, 40), Radius: theme.SmallRadius,
						Label: "一行正在来",
					}, func() {
						feedback.Skeleton(c, feedback.SkeletonOptions{
							FillWidth: true, Height: unit(c, 6),
							Radius: theme.SmallRadius, Label: "正在来的一行",
						})
					})
				})
			})
			feedback.ShimmerText(c, feedback.ShimmerTextOptions{
				Lines: 3, Width: unit(c, 46), LineHeight: unit(c, 3), Gap: unit(c, 1),
				Label: "一段正在来的文字",
			})
		})
	})
}

// waitSection shows the five spinners and the button that stops the work.
func waitSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "五个 loader 与 InterruptButton — 正在做，和怎么停下来")
	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		ui.Box(c).Grow(1).Height(unit(c, 22)).Radius(theme.ControlRadius).
			Background(k.Surface).Center().Children(func() {
			ui.Row(c).Gap(unit(c, 5)).AlignItems(ui.Center).Children(func() {
				feedback.BarsLoader(c, feedback.LoaderOptions{Label: "条"})
				feedback.DotsLoader(c, feedback.LoaderOptions{Label: "点"})
				feedback.OrbitLoader(c, feedback.LoaderOptions{Label: "环"})
				feedback.PulseLoader(c, feedback.LoaderOptions{Label: "脉冲"})
				feedback.WaveLoader(c, feedback.LoaderOptions{Label: "波"})
			})
		})
		ui.Box(c).Width(unit(c, 56)).Padding(unit(c, 1.5)).Radius(theme.ControlRadius).
			Background(k.Surface).Center().Children(func() {
			ui.Row(c).Gap(unit(c, 2)).AlignItems(ui.Center).Children(func() {
				feedback.InterruptButton(c, feedback.InterruptButtonOptions{
					Label: "停止同步", Name: "停止同步",
				})
				feedback.InterruptButton(c, feedback.InterruptButtonOptions{
					Label: "正在停…", Busy: true, BusyLabel: "正在停…", Name: "正在停",
				})
			})
		})
	})
	ui.Box(c).FillWidth().MarginY(unit(c, 1)).Children(func() {
		ui.Text(c, "BarsLoader / DotsLoader / OrbitLoader / PulseLoader / WaveLoader").
			TextColor(k.TextFaint).FontSize(core.FontSize(c, theme.CaptionSize))
	})
}

// progressSection shows the bars and the two small states that sit beside
// them.
func progressSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "Progress、StatusIndicator、Presence — 到了哪一步")
	ui.Column(c).FillWidth().Gap(unit(c, 2)).Children(func() {
		for _, p := range []struct {
			name  string
			value float32
			sev   core.Severity
		}{
			{"同步进度", 0.62, core.Accent},
			{"几乎完成", 0.95, core.Success},
			{"卡住了", 0.18, core.Warning},
		} {
			ui.Column(c).FillWidth().Gap(unit(c, 0.5)).Children(func() {
				ui.Text(c, p.name).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize))
				feedback.Progress(c, feedback.ProgressOptions{
					Value: p.value, Severity: p.sev, Label: p.name, ShowPercent: true,
				})
			})
		}
	})
	ui.Row(c).FillWidth().Gap(unit(c, 2)).Children(func() {
		ui.Box(c).Grow(1).Height(unit(c, 12)).Radius(theme.ControlRadius).
			Background(k.Surface).Center().Children(func() {
			ui.Row(c).Gap(unit(c, 3)).AlignItems(ui.Center).Children(func() {
				feedback.StatusIndicator(c, feedback.StatusIndicatorOptions{
					Label: "已连接", Severity: core.Success,
				})
				feedback.StatusIndicator(c, feedback.StatusIndicatorOptions{
					Label: "重试中", Severity: core.Warning, Busy: true,
				})
				feedback.StatusIndicator(c, feedback.StatusIndicatorOptions{
					Label: "已断开", Severity: core.Danger, Pill: true,
				})
			})
		})
		ui.Box(c).Width(unit(c, 62)).Height(unit(c, 12)).Radius(theme.ControlRadius).
			Background(k.Surface).Center().Children(func() {
			ui.Row(c).Gap(unit(c, 2)).AlignItems(ui.Center).Children(func() {
				for _, p := range []struct {
					name  string
					state feedback.PresenceState
				}{
					{"Andre", feedback.PresenceOnline},
					{"Mia", feedback.PresenceBusy},
					{"Ravi", feedback.PresenceAway},
					{"Lena", feedback.PresenceOffline},
				} {
					feedback.Presence(c, feedback.PresenceOptions{
						Name: p.name, State: p.state, ShowName: true, Size: unit(c, 4),
					})
				}
			})
		})
	})
}

// streamSection shows the two lists of things that happened, plus the two
// animated wrappers a list of records is built through.
func streamSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "ActivityFeed、NotificationCenter、BlinkHighlight — 刚刚发生了什么")
	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).Width(unit(c, 52)).Gap(unit(c, 1)).Children(func() {
			ui.Text(c, "ActivityFeed — 一条一条发生过的事").TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).Bold()
			feedback.ActivityFeed(c, feedback.ActivityFeedOptions{
				Items: []feedback.ActivityItem{
					{Title: "Callback assigned", Detail: "Riverside Clinic → Andre Thomson", Time: "10:40"},
					{Title: "Parts ordered", Detail: "Capacitor, contactor", Time: "13:05",
						Severity: core.Warning},
					{Title: "Sync failed", Detail: "网络原因，稍后重试", Time: "14:30",
						Severity: core.Danger},
				},
			})
		})
		ui.Column(c).Width(unit(c, 62)).Gap(unit(c, 1)).Children(func() {
			ui.Text(c, "NotificationCenter — 通知中心").TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).Bold()
			feedback.NotificationCenter(c, feedback.NotificationCenterOptions{
				Title: "通知", Clearable: true, Rule: true,
				Items: []feedback.Notification{
					{Title: "有人提到了你", Body: "「这单我下午过去」", At: "14:28",
						Severity: core.Accent},
					{Title: "同步失败", Body: "3 条回访没能送出去", At: "14:30",
						Severity: core.Danger},
					{Title: "今天已解决 12 条", Body: "", At: "14:30", Severity: core.Success},
				},
			})
		})
	})

	showcase.Field(c, "LayoutTransition 与 NewStagger — 列表进来和重排的样子")
	ui.Box(c).FillWidth().Padding(unit(c, 1.5)).Radius(theme.ControlRadius).
		Background(k.Surface).Children(func() {
		// The stagger hands out each child's entrance in turn, so a list
		// arrives as a list. The delays are real but the resting frame is
		// the same, which is what the page can show.
		s := feedback.NewStagger(c, feedback.StaggerOptions{})
		ui.Column(c).FillWidth().Gap(unit(c, 1)).Children(func() {
			for i, name := range []string{
				"Riverside Clinic", "Northgate Dental", "Harbour Cafe", "Elm Street Gym",
			} {
				name := name
				s.Wrap(feedback.LayoutTransition(c, feedback.LayoutTransitionOptions{
					Key: name,
				}, func() {
					ui.Box(c).FillWidth().Padding(unit(c, 1), unit(c, 1.5)).
						Radius(theme.SmallRadius).Background(k.Background).Label(name).
						Children(func() {
							ui.Text(c, name).TextColor(k.Text).
								FontSize(core.FontSize(c, theme.MetaSize))
							ui.Text(c, staggerNote(i)).TextColor(k.TextMuted).
								FontSize(core.FontSize(c, theme.CaptionSize))
						})
				}))
			}
		})
	})
	ui.Box(c).FillWidth().Children(func() {
		ui.Text(c, "BlinkHighlight — 刚刚变过的那个地方").TextColor(k.Text).
			FontSize(core.FontSize(c, theme.RowSize)).Bold()
		feedback.BlinkHighlight(c, feedback.BlinkHighlightOptions{
			Level: 0.8, Radius: theme.SmallRadius,
		}, func() {
			ui.Box(c).Width(unit(c, 62)).Padding(unit(c, 1.5)).Radius(theme.SmallRadius).
				Background(k.WarningBg).Children(func() {
				ui.Text(c, "CB-2864 的金额刚被人改过").TextColor(k.Text).
					FontSize(core.FontSize(c, theme.BodySize))
			})
		})
	})
}

// floatSection is last on the page: a toast and a loading overlay both draw
// themselves over the whole window, so here they are shown over a mock one
// and the rest of the page stays readable.
func floatSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "Toast、UndoToast、LoadingOverlay — 浮在整窗上的那几句")
	ui.Box(c).FillWidth().Height(unit(c, 46)).Radius(theme.ControlRadius).
		Background(k.Surface).Border(theme.BorderWidth, k.Border).Clip().Children(func() {
		ui.Column(c).FillHeight().Padding(unit(c, 2)).Gap(unit(c, 1.5)).Children(func() {
			ui.Text(c, "Callbacks").TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).Bold()
			ui.Box(c).FillWidth().Grow(1).Radius(theme.CardRadius).
				Background(k.Background).Padding(unit(c, 1.5)).Children(func() {
				ui.Text(c, "这一块就是浮层下面那块页面。").TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.MetaSize))
			})
			// A toast is a single line's worth of height, and it sits at the
			// foot of the window it belongs to rather than below it.
			feedback.Toast(c, "已保存 3 条回访", unit(c, 1))
		})
	})

	showcase.Field(c, "LoadingOverlay 与 UndoToast 的样子")
	ui.Box(c).FillWidth().Height(unit(c, 30)).Radius(theme.ControlRadius).
		Background(k.Background).Border(theme.BorderWidth, k.Border).Clip().
		Children(func() {
			feedback.LoadingOverlay(c, feedback.LoadingOverlayOptions{
				Message:     "正在拉取 27 条回访…",
				Loader:      feedback.LoaderOptions{Label: "加载中"},
				Scrim:       0.3,
				Radius:      theme.PanelRadius,
				CancelLabel: "取消",
			}, func() {
				ui.Column(c).FillHeight().Padding(unit(c, 2)).Gap(unit(c, 1)).Children(func() {
					ui.Text(c, "被盖住的是这一块。").TextColor(k.Text).
						FontSize(core.FontSize(c, theme.BodySize))
					ui.Text(c, "遮罩是半透明的，所以下面还认得出来。").TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.MetaSize))
				})
			})
		})
	ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(unit(c, 2)).Children(func() {
		ui.Text(c, "UndoToast：").TextColor(k.TextMuted).
			FontSize(core.FontSize(c, theme.CaptionSize))
		feedback.UndoToast(c, feedback.UndoToastOptions{
			Message: "已归档 2 条", Action: "撤销",
		})
		ui.Text(c, "Toast 与 UndoToast 一样只占内容那一行的高度").TextColor(k.TextFaint).
			FontSize(core.FontSize(c, theme.CaptionSize))
	})
}

// skeletonName names one of the skeleton bars, so a reader and a test can
// both say which is which.
func skeletonName(i int) string {
	return [...]string{"标题", "副行", "金额"}[i]
}

func staggerNote(i int) string {
	return [...]string{
		"第 0 项 · 第一个到的，没有额外延迟",
		"第 1 项 · 多等一步",
		"第 2 项 · 再多等一步",
		"第 3 项 · 最后到的",
	}[i]
}
