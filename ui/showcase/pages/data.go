package pages

import (
	"strconv"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/data"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/showcase"
	"github.com/HycJack/MintUI/ui/theme"
)

// Demo state outlives the frame: every demo below hands its component
// a pointer, and a pointer into a frame-local is a click the next
// frame undoes — a tab that will not switch, a dropdown that snaps
// shut, a slider that springs back.
var (
	shut = false
	sort = data.Sort{Column: "priority"}
	at   = 0
	img  = 0
)

func init() {
	showcase.Register(showcase.Page{
		Package: "data",
		Title:   "ui/data — 一条记录长什么样",
		Note:    "卡片、统计行、列表、表格、树、树表、时间线、描述列表、手风琴、折叠、轮播、看图",
		Width:   1000,
		Height:  1620,
		Want: []string{
			"Riverside Clinic",
			"CB-2871 · AC repair",
			"Northgate Dental",
			"Harbour Cafe",
			" 27 未处理",
			" 今天已解决 12",
			"Elm Street Gym",
			"昨天已解决 21 条",
			"今天",
			"昨天",
			"本周",
			"Customer",
			"Reference",
			"CB-2871",
			"Callback logged",
			"Assigned",
			"Parts ordered",
			"Sync failed",
			"第 1 张, 1 of 3",
			"现场照片",
			"2026-10-07 09:12",
		},
		Render: func(c *ui.Context) {
			dataPage(c)
		},
	})
}

// dataPage shows a record and the collections records live in. It is grouped
// by how much of the record each one shows: one card, a line of figures, a
// list of rows, a table of columns, and the two that walk a hierarchy.
func dataPage(c *ui.Context) {
	cardSection(c)
	figureSection(c)
	collectionSection(c)
	hierarchySection(c)
	describeSection(c)
}

// cardSection shows the one component a record is drawn on, with each of its
// slots filled and left out.
func cardSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "Card 与 StatLine — 一张卡和一行数字")
	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		// The card takes its width from the row, so three of them in a row
		// share it evenly — which is the only way a gallery can show three
		// cards at once without a board around them.
		ui.Box(c).Grow(1).Children(func() {
			data.Card(c, data.CardOptions{
				Title: "Riverside Clinic", Meta: "CB-2871 · AC repair",
				Header: func() {
					ui.Box(c).Padding(unit(c, 0.25), unit(c, 1.5)).
						Radius(theme.PillRadius).Background(k.WarningBg).Children(func() {
						ui.Text(c, "High").TextColor(k.Warning).
							FontSize(core.FontSize(c, theme.CaptionSize)).Bold()
					})
					ui.Box(c).Grow(1)
					ui.Text(c, "今天 14:30").TextColor(k.TextFaint).
						FontSize(core.FontSize(c, theme.CaptionSize))
				},
				Footer: func() {
					display.Avatar(c, "Andre Thomson")
					ui.Text(c, "Andre Thomson").TextColor(k.Text).
						FontSize(core.FontSize(c, theme.MetaSize)).Grow(1)
					ui.Text(c, "$184.50").TextColor(k.Text).
						FontSize(core.FontSize(c, theme.StatSize)).Bold()
				},
			}, func() {
				ui.Text(c, "客诉机不制冷，上午报的修。").TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.MetaSize))
			})
		})
		ui.Box(c).Grow(1).Children(func() {
			data.Card(c, data.CardOptions{
				Title: "Northgate Dental", Meta: "CB-2869 · Water leak",
				Meta2: "第二次回访", Draggable: "cb-2869",
				FooterRule: true,
				Footer: func() {
					showcase.Stack(c, 1, func() {
						input.Button(c, "接单", input.ButtonOptions{})
						input.Button(c, "延期", input.ButtonOptions{})
					})
				},
			}, nil)
		})
		ui.Box(c).Grow(1).Children(func() {
			data.Card(c, data.CardOptions{
				Title: "Harbour Cafe", Meta: "CB-2864 · Oven",
				Footer: func() {
					ui.Text(c, "已解决").TextColor(k.Success).
						FontSize(core.FontSize(c, theme.CaptionSize)).Bold()
					ui.Box(c).Grow(1)
					ui.Text(c, "$92.00").TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.MetaSize))
				},
			}, nil)
		})
	})

	data.StatLine(c,
		data.Stat{Label: "27 未处理"},
		data.Stat{Label: "4 逾期", Value: "", Muted: true},
		data.Stat{Label: "今天已解决 12"},
	)
}

// figureSection shows the two ways of saying "this is empty": a collapsible
// and an accordion of sections.
func figureSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "Collapsible 与 Accordion — 能收起来的东西")
	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		ui.Box(c).Grow(2).Children(func() {
			open := true
			data.Collapsible(c, data.CollapsibleOptions{
				Title: "今天已解决 12 条", Open: &open, Meta: "2026-10-07", Rule: true,
			}, func() {
				ui.Column(c).FillWidth().Gap(unit(c, 0.5)).Children(func() {
					for _, r := range []string{
						"Elm Street Gym · CB-2858", "Pinewood Dental · CB-2851",
					} {
						ui.Text(c, r).TextColor(k.Text).
							FontSize(core.FontSize(c, theme.MetaSize))
					}
				})
			})
			data.Collapsible(c, data.CollapsibleOptions{
				Title: "昨天已解决 21 条", Open: &shut, Meta: "2026-10-06",
			}, func() {
				ui.Text(c, "收起来时里面不画").TextColor(k.TextFaint).
					FontSize(core.FontSize(c, theme.MetaSize))
			})
		})
		ui.Box(c).Grow(1).Children(func() {
			first := showcase.State(c, "data.170.first", true)
			second := showcase.State(c, "data.170.second", false)
			third := showcase.State(c, "data.170.third", false)
			data.Accordion(c, data.AccordionOptions{Single: true},
				data.Section{
					Title: "今天", Open: first, Meta: "12",
					Body: func() {
						ui.Text(c, "今天解决的 12 条").TextColor(k.TextMuted).
							FontSize(core.FontSize(c, theme.MetaSize))
					},
				},
				data.Section{
					Title: "昨天", Open: second, Meta: "21",
					Body: func() {
						ui.Text(c, "昨天解决的 21 条").TextColor(k.TextMuted).
							FontSize(core.FontSize(c, theme.MetaSize))
					},
				},
				data.Section{
					Title: "本周", Open: third, Meta: "96",
					Body: func() {
						ui.Text(c, "本周解决的 96 条").TextColor(k.TextMuted).
							FontSize(core.FontSize(c, theme.MetaSize))
					},
				},
			)
		})
	})
}

// collectionSection shows a list of rows, a table of columns, and the
// carousel and picture viewer that page through a set.
func collectionSection(c *ui.Context) {
	k := core.Tokens(c)
	sel := 1
	showcase.Section(c, "List 与 DataTable — 一行一件事，一行几件事")
	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).Width(unit(c, 46)).Gap(unit(c, 1)).Children(func() {
			ui.Text(c, "List — 只有一列").TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).Bold()
			data.List(c, data.ListOptions{
				Rows: 4, Height: unit(c, 54), RowHeight: unit(c, 11),
				Selected: &sel, Label: func(i int) string { return recordAt(i).Name },
			}, func(row int) {
				r := recordAt(row)
				ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(unit(c, 1.5)).Children(func() {
					display.Avatar(c, r.Tech)
					ui.Text(c, r.Name).TextColor(k.Text).
						FontSize(core.FontSize(c, theme.MetaSize)).Grow(1).SingleLine()
					ui.Text(c, r.Ref).TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.CaptionSize))
				})
			})
			ui.Box(c).FillWidth().MarginY(unit(c, 1)).Children(func() {
				ui.Text(c, "空列表画 Empty，不画一片空白").TextColor(k.TextFaint).
					FontSize(core.FontSize(c, theme.CaptionSize))
				data.List(c, data.ListOptions{
					Rows: 0, Height: unit(c, 12),
					Label: func(int) string { return "空" },
					Empty: func() {
						ui.Text(c, "这一屏还没有回访").TextColor(k.TextMuted).
							FontSize(core.FontSize(c, theme.MetaSize))
					},
				}, func(int) {})
			})
		})

		ui.Column(c).Grow(1).Gap(unit(c, 1)).Children(func() {
			ui.Text(c, "DataTable — 六列，行可排序可选").TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).Bold()
			// The table gets a width of its own so the five columns have a
			// fixed set of widths to share: left to measure whatever is left
			// over, the first column takes what the other four do not use,
			// and a customer name is shown as three letters.
			ui.Box(c).Width(unit(c, 118)).Children(func() {
				cols := []data.Column{
					// The widths add up to the 116 the table is given. A name
					// that wraps makes its row twice as tall as its
					// neighbours and the table stops reading as a table, so
					// Technician gets the room for "Andre Thomson" and
					// Customer gives some back: "Riverside Clinic" still fits.
					{Title: "Customer", ID: "name", Width: unit(c, 42)},
					{Title: "Priority", ID: "priority", Width: unit(c, 17), Sortable: true},
					{Title: "Technician", ID: "tech", Width: unit(c, 28), Sortable: true},
					{Title: "Due", ID: "due", Width: unit(c, 14), Align: ui.End},
					{Title: "Cost", ID: "cost", Width: unit(c, 15), Align: ui.End, Sortable: true},
				}
				rows := data.Rows(records(), sort, map[string]func(a, b record) int{
					"name":     data.ByText(func(r record) string { return r.Name }),
					"priority": data.ByNumber(func(r record) float64 { return float64(r.Rank) }),
					"tech":     data.ByText(func(r record) string { return r.Tech }),
					"cost":     data.ByNumber(func(r record) float64 { return r.Cost }),
				})
				data.DataTable(c, data.DataTableOptions{
					// Width has to be the 118 the box is, not something
					// smaller: the five columns add up to 116, and a table
					// narrower than its own columns squeezes them until
					// "Andre Thomson" wraps. A wrapped name makes its row
					// taller, and a fixed Height that used to fit now
					// clips the last row in half — which is how this page
					// came to show a table cut off at the bottom.
					Columns: cols, Rows: len(rows), Height: unit(c, 74), RowHeight: unit(c, 11),
					Width: unit(c, 118),
					Sort:  &sort, Selected: &sel,
					Key:   func(i int) any { return rows[i].Ref },
					Label: func(i int) string { return rows[i].Name },
					CellLabel: func(i, col int) string {
						return cellText(rows[i], cols[col].ID)
					},
					Cell: func(i, col int) {
						r := rows[i]
						fg, bold := k.Text, false
						switch col {
						case 1:
							sev := priorityOf(r.Rank)
							bg, ink := sev.Pair(k)
							ui.Box(c).Padding(unit(c, 0.25), unit(c, 1.5)).
								Radius(theme.PillRadius).Background(bg).Children(func() {
								ui.Text(c, priorityName(r.Rank)).TextColor(ink).
									FontSize(core.FontSize(c, theme.CaptionSize)).Bold()
							})
							return
						case 4:
							bold = true
							fg = k.Text
						case 3:
							fg = k.TextMuted
						}
						t := ui.Text(c, cellText(r, cols[col].ID)).TextColor(fg).
							FontSize(core.FontSize(c, theme.MetaSize))
						if col == 0 {
							t.Grow(1)
						}
						if bold {
							t.Bold()
						}
					},
				})
			})
			ui.Text(c, "按 priority 升序；点表头会写回 Sort").TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize))
		})
	})
}

// hierarchySection shows the two components that walk a hierarchy, which
// differ in whether the items have columns.
func hierarchySection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "Tree 与 TreeTable — 有层级的东西")
	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).Width(unit(c, 34)).Gap(unit(c, 1)).Children(func() {
			ui.Text(c, "Tree — 只有一列").TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).Bold()
			var open data.Open[string]
			open.Open("north")
			sel := 0
			data.Tree(c, data.TreeOptions[string]{
				Roots:     branches,
				Children:  branchChildren,
				Open:      &open,
				Height:    unit(c, 40),
				RowHeight: unit(c, 10),
				Selected:  &sel,
				Label:     func(s string) string { return s },
				Row: func(item string, depth int) {
					ui.Row(c).AlignItems(ui.Center).Children(func() {
						ui.Text(c, item).TextColor(k.Text).
							FontSize(core.FontSize(c, theme.MetaSize))
					})
				},
			})
			ui.Text(c, "箭头把开合写回 Open，组件自己不存").TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize))
		})

		ui.Column(c).Grow(1).Gap(unit(c, 1)).Children(func() {
			ui.Text(c, "TreeTable — 层级 + 列").TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).Bold()
			var open data.Open[string]
			open.Open("north")
			cols := []data.Column{
				{Title: "Branch", ID: "name", Width: unit(c, 40)},
				{Title: "Open", ID: "open", Width: unit(c, 16), Align: ui.End},
			}
			ui.Box(c).Width(unit(c, 58)).Children(func() {
				data.TreeTable(c, data.TreeTableOptions[string]{
					Roots:    branches,
					Children: branchChildren,
					Open:     &open,
					Columns:  cols,
					Height:   unit(c, 40),
					Label:    func(s string) string { return s },
					CellLabel: func(item string, col int) string {
						return cellText(recordAt(0), cols[col].ID)
					},
					Cell: func(item string, col int) {
						if col == 0 {
							ui.Text(c, item).TextColor(k.Text).
								FontSize(core.FontSize(c, theme.MetaSize)).Grow(1)
							return
						}
						ui.Text(c, cellText(recordAt(branchIndex(item)), "open")).
							TextColor(k.TextMuted).FontSize(core.FontSize(c, theme.MetaSize))
					},
				})
			})
		})
	})
}

// describeSection shows the last three: what a set of pictures looks like,
// the labelled pairs a record is described in, and the timeline.
func describeSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "Carousel、ImageViewer、DescriptionList 与 Timeline")
	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).Width(unit(c, 30)).Gap(unit(c, 1)).Children(func() {
			ui.Text(c, "Carousel").TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).Bold()
			data.Carousel(c, data.CarouselOptions{
				At: &at, Height: unit(c, 22), Dots: true,
				Slides: []data.Slide{
					{Label: "第 1 张", Body: func() {
						slideBody(c, "今天", k)
					}},
					{Label: "第 2 张", Body: func() {
						slideBody(c, "明天", k)
					}},
					{Label: "第 3 张", Body: func() {
						slideBody(c, "本周", k)
					}},
				},
			})
		})
		ui.Column(c).Width(unit(c, 30)).Gap(unit(c, 1)).Children(func() {
			ui.Text(c, "ImageViewer").TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).Bold()
			data.ImageViewer(c, data.ImageViewerOptions{
				Images:  pictures(),
				At:      &img,
				Height:  unit(c, 22),
				Caption: "现场照片",
				Strip:   true,
				Label:   func(i int) string { return "现场照片 " + itoa(i+1) },
			})
		})
		ui.Column(c).Grow(1).Gap(unit(c, 1)).Children(func() {
			ui.Text(c, "DescriptionList — 名 / 值").TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).Bold()
			data.DescriptionList(c, data.DescriptionListOptions{
				TermWidth: unit(c, 18), Rules: true,
			},
				data.Term{Term: "Customer", Value: "Riverside Clinic"},
				data.Term{Term: "Reference", Value: "CB-2871"},
				data.Term{Term: "Branch", Value: "North"},
				data.Term{Term: "Opened", Value: "2026-10-07 09:12"},
			)
		})
	})

	showcase.Field(c, "Timeline — 一条记录经历过什么")
	data.Timeline(c, data.TimelineOptions{
		Height: unit(c, 40), Stamps: true, StampWidth: unit(c, 26),
	},
		data.Event{When: "09:12", Title: "Callback logged", Detail: "Riverside Clinic", Marker: markerDot(c, core.Accent)},
		data.Event{When: "10:40", Title: "Assigned", Detail: "Andre Thomson", Marker: markerDot(c, core.Neutral)},
		data.Event{When: "13:05", Title: "Parts ordered", Detail: "Capacitor, contactor", Tone: core.Warning, Marker: markerDot(c, core.Warning)},
		data.Event{When: "14:30", Title: "Sync failed", Detail: "网络原因，稍后重试", Tone: core.Danger, Marker: markerDot(c, core.Danger)},
	)
}

// slideBody is what one carousel slide holds.
func slideBody(c *ui.Context, title string, k theme.Tokens) {
	ui.Box(c).Fill().Padding(unit(c, 1.5)).Radius(theme.ControlRadius).
		Background(k.Surface).Center().Children(func() {
		ui.Text(c, title).TextColor(k.Text).FontSize(core.FontSize(c, theme.BodySize)).Bold()
	})
}

// markerDot is a timeline row's own mark; without one the row is a line and
// some text.
func markerDot(c *ui.Context, sev core.Severity) func() {
	return func() {
		_, fg := sev.Pair(core.Tokens(c))
		ui.Box(c).Size(unit(c, 2.25), unit(c, 2.25)).Radius(unit(c, 1.2)).Background(fg)
	}
}

// record is one row of the table, the list and the tree on this page. The
// page's data is a fixed slice built in code: a gallery that read a database
// or a clock would draw a different picture every time it was rendered.
type record struct {
	Name, Ref, Tech, Due, Branch string
	Rank                         int
	Cost                         float64
}

var allRecords = []record{
	{"Riverside Clinic", "CB-2871", "Andre Thomson", "今天", "North", 3, 184.50},
	{"Northgate Dental", "CB-2869", "Mia Chen", "今天", "North", 2, 96.00},
	{"Harbour Cafe", "CB-2864", "Ravi Patel", "明天", "Harbour", 4, 412.75},
	{"Elm Street Gym", "CB-2858", "Lena Ford", "明天", "North", 1, 58.20},
	{"Pinewood Dental", "CB-2851", "Sam Ortiz", "10-09", "Harbour", 3, 231.00},
	{"Westside Vet", "CB-2844", "Mia Chen", "10-09", "Harbour", 2, 74.00},
	{"Lakeside Inn", "CB-2839", "Andre Thomson", "10-10", "North", 1, 39.90},
}

func records() []record { return allRecords }

func recordAt(i int) record { return allRecords[i%len(allRecords)] }

// cellText is one cell of the sample table, by column id.
func cellText(r record, id string) string {
	switch id {
	case "name":
		return r.Name
	case "priority":
		return priorityName(r.Rank)
	case "tech":
		return r.Tech
	case "due":
		return r.Due
	case "cost":
		return money(r.Cost)
	case "open":
		return itoa(recordAt(branchIndex(r.Branch)).Rank) + " 条"
	}
	return ""
}

func money(v float64) string {
	return "$" + strconv.FormatFloat(v, 'f', 2, 64)
}

// priorityOf maps a rank to the severity the design system gives it.
func priorityOf(rank int) core.Severity {
	switch rank {
	case 4:
		return core.Danger
	case 3:
		return core.Warning
	}
	return core.Neutral
}

func priorityName(rank int) string {
	switch rank {
	case 4:
		return "Critical"
	case 3:
		return "High"
	case 2:
		return "Medium"
	}
	return "Low"
}

// branches is the fixed hierarchy the tree and the tree table walk.
var branches = []string{"North", "Harbour", "Closed"}

func branchChildren(item string) []string {
	switch item {
	case "North":
		return []string{"Riverside Clinic", "Northgate Dental", "Elm Street Gym"}
	case "Harbour":
		return []string{"Harbour Cafe", "Pinewood Dental", "Westside Vet"}
	}
	return nil
}

func branchIndex(branch string) int {
	if branch == "Harbour" {
		return 2
	}
	return 0
}

// pictures are the bitmaps the carousel's viewer pages through. They are made
// here rather than loaded, so the gallery depends on nothing on disk.
func pictures() []ui.ImageSource {
	out := make([]ui.ImageSource, 3)
	for i := range out {
		out[i] = sampleBitmap()
	}
	return out
}

func itoa(n int) string { return strconv.Itoa(n) }
