package pages

// The chart page is the whole of ui/chart in one column, in the order a
// person reads a chart rather than the order the files are declared in:
// what a value did over time, what it adds up to, how a set is spread, how
// parts of a whole sit against each other, what flows between what, and then
// the four pieces every one of those is built from.
//
// Every chart below is given real numbers. A gallery of charts drawn over
// nothing would look exactly like a gallery of charts, and the one thing this
// page exists to catch — a chart that draws its data in the wrong place, or
// two thirds of it — is invisible without numbers to compare against.

import (
	"fmt"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/chart"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/showcase"
	"github.com/HycJack/MintUI/ui/theme"
)

func init() {
	showcase.Register(showcase.Page{
		Package: "chart",
		Title:   "ui/chart — 数据画出来是什么样子",
		Note:    "坐标轴、网格、图例、系列、指针，以及 34 种图表和一堆纯函数",
		Width:   1000,
		Height:  7700,
		Want: []string{
			// the series a legend is built from
			"Opened", "Resolved", "Backlog",
			// cartesian
			"opened by day", "resolved by day", "one total, split",
			"ridgeline", "bars", "bars, horizontal",
			// shares, many axes, money
			"calls per category", "share of a whole", "which few matter",
			"six measures, one shape", "a day that repeats",
			"from call to invoice", "area is the value", "angle is the value",
			"tickets resolved per day", "revenue movement",
			"stage by stage", "who pays for what", "who talks to whom",
			"modules and their callers", "the same tree, in rows",
			// the empty state, which is the one a chart must never skip
			"Nothing to plot. no callbacks were resolved this week",
			"Nothing to plot. a bar chart of nothing is still a chart",
			"Nothing to plot. the axis would be a scale over a range no data reaches",
			// the pieces, by the label they answer to
			"Grid", "Annotation", "Range", "Brush", "Crosshair", "Series",
			"Day axis", "Callbacks axis", "a week of callbacks",
			// the pure functions, by call and by answer
			"Nice(0, 97, 5)", "Bins(latency, 3)",
			"Stats / Summary(latency)", "Cumulative([38 27 21 14])",
			"Stack(opened+closed)", "Descending([38 27 21 14])",
			"Squarify(tiles, 400×220)", "Sweep([38 27 21 14])",
			"Depths(5, sankey links)", "HeatColor(40, 0, 60)",
			"DomainOf(latency)", "SeriesDomain(series)",
			"Linear(0, 100, 4)", "Log(1, 1000, 10)",
			"Bounds(stacked)", "Ends([38 27 21 14])",
			"Palette(6)", "Measure(frame)", "Ticks(y)",
		},
		Render: func(c *ui.Context) {
			chartPage(c)
		},
	})
}

// chartPage is the whole package in reading order.
func chartPage(c *ui.Context) {
	chartCartesian(c)
	chartBars(c)
	chartSpark(c)
	chartScatter(c)
	chartDistribution(c)
	chartHeat(c)
	chartManyAxes(c)
	chartShare(c)
	chartMoney(c)
	chartDial(c)
	chartFlow(c)
	chartRelation(c)
	chartPieces(c)
	chartEmpty(c)
	chartFunctions(c)
	chartConventions(c)
}

// ── the data ───────────────────────────────────────────────────────────────
//
// One week of a call centre, in fixed numbers. Nothing here is generated at
// random: a screenshot of this page that changed between two runs could not be
// compared against the last one, and a chart gallery is only useful if the
// reader can say "that bar is wrong" about a specific bar.

var (
	chDays     = []float64{0, 1, 2, 3, 4, 5, 6}
	chWeekdays = []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	chOpened   = []float64{4, 7, 5, 9, 8, 12, 11}
	chClosed   = []float64{2, 4, 3, 6, 5, 7, 6}
	chBacklog  = []float64{18, 24, 27, 31, 29, 26, 22}

	chLatency = []float64{120, 180, 90, 240, 410, 150, 95, 320, 200, 610, 175, 260}

	// Four, not five: these three charts are a third of the page each, and a
	// five-entry legend is wider than a third of the page — a legend that does
	// not fit is a legend that is drawn in half.
	chChannels = []string{"Phone", "Email", "Walk-in", "Portal"}
	chVolume   = []float64{38, 27, 21, 14}

	chCauses = []string{"No power", "Leak", "Noisy", "Blocked drain", "Thermostat", "Wiring"}
	chCounts = []float64{21, 17, 14, 9, 6, 3}
)

// chDayScale is the week's categories as an axis. Handing a chart a scale
// rather than letting it work one out from the x values is the difference
// between "Mon Tue Wed" and "0.857143 2.571429" along the bottom.
func chDayScale() chart.Scale {
	return chart.NewBand(chWeekdays, 0, 1)
}

// chSeries is the week's three series, so the legend, the plot and the
// infrastructure section all read from one place.
func chSeries(c *ui.Context) []chart.Series {
	return []chart.Series{
		chart.SeriesFrom("Opened", chart.FormLine, chDays, chOpened),
		chart.SeriesFrom("Resolved", chart.FormArea, chDays, chClosed),
		chart.SeriesFrom("Backlog", chart.FormLine, chDays, chBacklog),
	}
}

// ── 1. cartesian ───────────────────────────────────────────────────────────

func chartCartesian(c *ui.Context) {
	showcase.Section(c, "折线与面积 · LineChart / AreaChart / AreaMountain")

	showcase.Field(c, "LineChart 与 AreaChart — 同一周，两种读法")
	series := chSeries(c)
	chartTwoUp(c,
		func() {
			chart.LineChart(c, chart.LineOptions{
				ChartOptions: chart.ChartOptions{
					Height: 220, Grid: true, Label: "opened by day",
					X: chart.AxisOptions{Scale: chDayScale(), Label: "Day"},
					Y: chart.AxisOptions{Side: chart.Left, Count: 5, Label: "Callbacks"},
					Legend: chart.LegendOptions{
						Entries: chart.EntriesOf(c, series[:2]),
					},
				},
				Series: series[:2], Dots: true,
			})
		},
		func() {
			chart.AreaChart(c, chart.AreaOptions{
				ChartOptions: chart.ChartOptions{
					Height: 220, Grid: true, Label: "resolved by day",
					X: chart.AxisOptions{Scale: chDayScale(), Label: "Day"},
					Y: chart.AxisOptions{Side: chart.Left, Count: 5, Label: "Callbacks"},
				},
				Series: series[1:2],
			})
		})

	showcase.Field(c, "AreaChart 堆叠 与 AreaMountain — 三条带子叠一层，还是三条山脊压着")
	chartTwoUp(c,
		func() {
			chart.AreaChart(c, chart.AreaOptions{
				ChartOptions: chart.ChartOptions{
					Height: 200, Grid: true, Label: "one total, split",
					X: chart.AxisOptions{Scale: chDayScale(), Label: "Day"},
					// A stack is taller than any of its own series, so its
					// axis is brought rather than worked out: the scale a
					// chart builds is the span of the values, and the top
					// of a stack is their sum.
					Y: chart.AxisOptions{
						Count: 5, Side: chart.Left, Label: "Callbacks",
						Scale: chart.NewLinear(chart.Domain{Min: 0, Max: 20}, 0, 1),
					},
				},
				Series: series[:2], Stacked: true,
			})
		},
		func() {
			chart.AreaMountain(c, chart.AreaMountainOptions{
				ChartOptions: chart.ChartOptions{
					Height: 200, Grid: true, Label: "ridgeline",
					X: chart.AxisOptions{Scale: chDayScale(), Label: "Day"},
					Y: chart.AxisOptions{Count: 5, Side: chart.Left, Label: "Callbacks"},
				},
				Series: series,
			})
		})
}

// ── 2. bars ────────────────────────────────────────────────────────────────

func chartBars(c *ui.Context) {
	showcase.Section(c, "柱状 · BarChart / StackedBar / PercentBar")

	showcase.Field(c, "BarChart — 竖着的与横着的；分类名太长时横过来写")
	opened := chSeries(c)[:1]
	chartThree(c, 2,
		func() {
			chart.BarChart(c, chart.BarOptions{
				ChartOptions: chart.ChartOptions{
					Height: 210, Label: "bars",
					X: chart.AxisOptions{Scale: chDayScale(), Label: "Day"},
					Y: chart.AxisOptions{Count: 5, Side: chart.Left, Label: "Opened"},
				},
				Series: opened,
			})
		},
		func() {
			chart.BarChart(c, chart.BarOptions{
				ChartOptions: chart.ChartOptions{
					Height: 210, Grid: true, Label: "bars, horizontal",
					Y: chart.AxisOptions{Count: 5, Side: chart.Left, Scale: chDayScale(), Label: "Day"},
					X: chart.AxisOptions{Count: 5, Label: "Opened"},
				},
				Series: opened, Horizontal: true,
			})
		})

	showcase.Field(c, "StackedBar 与 PercentBar — 总数和占比，同一批数")
	stacked := []chart.Series{
		chart.SeriesFrom("Opened", chart.FormBar, chDays, chOpened),
		chart.SeriesFrom("Resolved", chart.FormBar, chDays, chClosed),
	}
	chartTwoUp(c,
		func() {
			chart.StackedBar(c, chart.StackedBarOptions{
				ChartOptions: chart.ChartOptions{
					Height: 210, Label: "calls per category",
					X: chart.AxisOptions{Scale: chDayScale(), Label: "Day"},
					Y: chart.AxisOptions{Count: 5, Side: chart.Left, Label: "Calls"},
				},
				Labels: chWeekdays, Series: stacked,
			})
		},
		func() {
			chart.PercentBar(c, chart.StackedBarOptions{
				ChartOptions: chart.ChartOptions{
					Height: 210, Label: "share of a whole",
					X: chart.AxisOptions{Scale: chDayScale(), Label: "Day"},
					Y: chart.AxisOptions{Count: 5, Side: chart.Left, Label: "Share of the column"},
				},
				Labels: chWeekdays, Series: stacked,
			})
		})
}

// ── 3. sparklines ──────────────────────────────────────────────────────────

func chartSpark(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "迷你图 · Sparkline / SparkBar — 表格行和卡片角落里的那一条")

	// A fixed-width row rather than a filling one: the whole point of a
	// sparkline is that it is a small mark inside a bigger row, and a row
	// that filled the page would stop looking like one.
	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 3)).AlignItems(ui.Center).
		Children(func() {
			// Six marks across a page that is 236 units wide: 36 each plus
			// five 3-unit gaps is 231, which is the whole arithmetic of why
			// the last bar is still on the page.
			spark := func(name string, opts chart.SparklineOptions) {
				ui.Column(c).Width(unit(c, 36)).Gap(unit(c, 0.5)).Children(func() {
					chart.Sparkline(c, opts)
					ui.Text(c, name).TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.CaptionSize))
				})
			}
			spark("一条线", chart.SparklineOptions{
				Values: chBacklog, Width: unit(c, 36), Height: unit(c, 10)})
			spark("填了底", chart.SparklineOptions{
				Values: chBacklog, Width: unit(c, 36), Height: unit(c, 10), Fill: true})
			spark("自己的颜色", chart.SparklineOptions{
				Values: chOpened, Width: unit(c, 36), Height: unit(c, 10),
				Fill: true, Color: k.Success})
			spark("负数", chart.SparklineOptions{
				Values: []float64{-4, -2, -7, -1, -9, -3, -5}, Width: unit(c, 36),
				Height: unit(c, 10), Fill: true, Color: k.Danger})
			ui.Column(c).Width(unit(c, 36)).Gap(unit(c, 0.5)).Children(func() {
				chart.SparkBar(c, chart.SparkBarOptions{
					Values: chOpened, Width: unit(c, 36), Height: unit(c, 10)})
				ui.Text(c, "SparkBar").TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize))
			})
			ui.Column(c).Width(unit(c, 36)).Gap(unit(c, 0.5)).Children(func() {
				chart.SparkBar(c, chart.SparkBarOptions{
					Values: chBacklog, Width: unit(c, 36), Height: unit(c, 10),
					Color: k.Warning})
				ui.Text(c, "一格一根").TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize))
			})
		})
}

// ── 4. scatter and bubbles ─────────────────────────────────────────────────

func chartScatter(c *ui.Context) {
	showcase.Section(c, "散点与气泡 · ScatterChart / BubbleChart")

	showcase.Field(c, "ScatterChart — 每个点一个观察，中间什么也不连")
	scatter := []chart.Series{
		chart.NewSeries("Hourly", chart.FormPoint,
			chart.Point{X: 1, Y: 12}, chart.Point{X: 2, Y: 19}, chart.Point{X: 3, Y: 15},
			chart.Point{X: 4, Y: 28}, chart.Point{X: 5, Y: 24}, chart.Point{X: 6, Y: 34},
			chart.Point{X: 7, Y: 31}, chart.Point{X: 8, Y: 42}, chart.Point{X: 9, Y: 37},
			chart.Point{X: 10, Y: 48}, chart.Point{X: 11, Y: 44}),
		chart.NewSeries("Weekly", chart.FormPoint,
			chart.Point{X: 1, Y: 8}, chart.Point{X: 3, Y: 22}, chart.Point{X: 5, Y: 17},
			chart.Point{X: 7, Y: 35}, chart.Point{X: 9, Y: 30}, chart.Point{X: 11, Y: 52}),
	}
	chartTwoUp(c,
		func() {
			chart.ScatterChart(c, chart.ScatterOptions{
				ChartOptions: chart.ChartOptions{
					Height: 230, Grid: true, Label: "calls vs first response",
					X:      chart.AxisOptions{Count: 5, Label: "Hour"},
					Y:      chart.AxisOptions{Side: chart.Left, Count: 5, Label: "Minutes"},
					Legend: chart.LegendOptions{Entries: chart.EntriesOf(c, scatter)},
				},
				Series: scatter,
			})
		},
		func() {
			chart.BubbleChart(c, chart.BubbleOptions{
				ChartOptions: chart.ChartOptions{
					Height: 230, Grid: true, Label: "one mark, three numbers",
					X: chart.AxisOptions{Count: 5, Label: "Cost"},
					Y: chart.AxisOptions{Count: 5, Side: chart.Left, Label: "Hours"},
				},
				// Named one by one: a legend of seven marks all called
				// "Job" is the same word seven times.
				Bubbles: chBubbles(c),
			})
		})
}

// chBubbles is a week of jobs as three numbers each: what it cost, how long
// it took, and how much of the technician's day it was.
func chBubbles(c *ui.Context) []chart.Bubble {
	cost := []float64{12, 28, 45, 62, 95}
	hours := []float64{1.2, 2.1, 3.4, 2.8, 5.1}
	share := []float64{4, 9, 14, 7, 22}
	// Short names, and only five of them: the legend is as wide as the
	// chart, and a legend of seven full customer names is wider than half
	// a page.
	names := []string{"Maple", "Oak", "Riverside", "Harbour", "Cedar"}
	pal := chart.Palette(c, len(names))
	out := make([]chart.Bubble, len(names))
	for i, name := range names {
		out[i] = chart.Bubble{
			Point: chart.Point{X: cost[i], Y: hours[i]},
			Size:  share[i], Name: name, Color: pal[i],
		}
	}
	return out
}

// ── 5. distribution ────────────────────────────────────────────────────────

func chartDistribution(c *ui.Context) {
	showcase.Section(c, "分布 · Histogram / BoxPlot / ViolinPlot")

	samples := []chart.Sample{
		{Label: "Phone", Values: []float64{80, 95, 110, 130, 150, 165, 190, 210, 260}},
		{Label: "Email", Values: []float64{200, 260, 300, 340, 420, 460, 510, 580, 610}},
		{Label: "Portal", Values: []float64{40, 60, 70, 90, 110, 130, 160, 180, 240}},
	}

	chartThree(c, 2,
		func() {
			chart.Histogram(c, chart.HistogramOptions{
				ChartOptions: chart.ChartOptions{
					Height: 220, Label: "latency spread",
					// Whole minutes: without a format the bucket edges come
					// out as 90.000000–176.666667.
					X: chart.AxisOptions{Count: 5, Format: chart.Fixed(0),
						Label: "Minutes"},
					Y: chart.AxisOptions{Count: 5, Side: chart.Left, Label: "Calls"},
				},
				Values: chLatency, Bins: 4,
			})
		},
		func() {
			chart.BoxPlot(c, chart.BoxPlotOptions{
				ChartOptions: chart.ChartOptions{
					Height: 220, Grid: true, Label: "five numbers per group",
					Y: chart.AxisOptions{Count: 5, Side: chart.Left, Label: "Minutes"},
				},
				Samples: samples,
			})
		})

	chart.ViolinPlot(c, chart.ViolinOptions{
		ChartOptions: chart.ChartOptions{
			Height: 230, Grid: true, Label: "shape, not size",
			Y: chart.AxisOptions{Count: 5, Side: chart.Left, Label: "Minutes"},
		},
		Samples: samples,
	})
}

// ── 6. heat ────────────────────────────────────────────────────────────────

func chartHeat(c *ui.Context) {
	showcase.Section(c, "热力 · HeatmapChart / CalendarHeatmap")

	rows := []string{"Mon", "Tue", "Wed", "Thu", "Fri"}
	cols := []string{"08", "09", "10", "11", "12", "13", "14", "15"}
	values := [][]float64{
		{3, 9, 14, 11, 6, 8, 12, 5},
		{7, 18, 22, 16, 9, 13, 19, 8},
		{5, 12, 16, 13, 7, 10, 15, 6},
		{9, 21, 26, 19, 11, 15, 24, 9},
		{6, 14, 18, 15, 8, 12, 17, 7},
	}
	showcase.Field(c, "HeatmapChart — 写上数字（Write）与只留颜色，两种读法")
	chartTwoUp(c,
		func() {
			chart.HeatmapChart(c, chart.HeatmapOptions{
				ChartOptions: chart.ChartOptions{
					Height: 200, Label: "calls per hour",
					X: chart.AxisOptions{Count: 5, Label: "Hour"},
					Y: chart.AxisOptions{Count: 5, Side: chart.Left, Label: "Day"},
				},
				Rows: rows, Cols: cols, Values: values, Write: true,
			})
		},
		func() {
			chart.HeatmapChart(c, chart.HeatmapOptions{
				ChartOptions: chart.ChartOptions{
					Height: 200, Label: "colour only",
					X: chart.AxisOptions{Count: 5, Label: "Hour"},
					Y: chart.AxisOptions{Count: 5, Side: chart.Left, Label: "Day"},
				},
				Rows: rows, Cols: cols, Values: values, Reverse: true,
			})
		})

	showcase.Field(c, "CalendarHeatmap — 半年 182 天，按它们真正落在的星期排")
	chart.CalendarHeatmap(c, chart.CalendarHeatmapOptions{
		ChartOptions: chart.ChartOptions{
			Height: 150, Label: "half a year of callbacks",
		},
		Days:        chYear(),
		Weekday:     time.Monday,
		MonthLabels: true,
	})
}

// chYear is a fixed run of daily counts, so the calendar heatmap always draws
// the same wall.
//
// Half a year rather than a whole one because of the page, not the chart: a
// year is fifty-three week columns, and in a page 944 points wide that is
// seventeen points a cell — at which the month names and the week dates
// underneath them are written on top of each other. A gallery is a place
// where a component can be looked at, and a heatmap nobody can read is not.
func chYear() []chart.Day {
	out := make([]chart.Day, 0, 182)
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := range 182 {
		d := start.AddDate(0, 0, i)
		// A week that has a rhythm in it: quiet at the weekend, busy on
		// Mondays. A heatmap of pure noise says nothing about anything.
		n := float64(6 + (i*7+i/3)%17)
		if d.Weekday() == time.Saturday || d.Weekday() == time.Sunday {
			n /= 3
		}
		out = append(out, chart.Day{Date: d.Format(chartDayLayout), Value: n})
	}
	return out
}

const chartDayLayout = "2006-01-02"

// ── 7. many axes ───────────────────────────────────────────────────────────

func chartManyAxes(c *ui.Context) {
	showcase.Section(c, "多轴 · ParallelCoordinates / ParetoChart / RadarChart / PolarChart")

	showcase.Field(c, "ParallelCoordinates — 每一列有自己的刻度，每一行连成一条线")
	parallel := []chart.Series{
		chart.SeriesFrom("Riverside", chart.FormLine,
			[]float64{4, 22, 184, 92, 88},
			[]float64{31, 18, 12, 74, 96}),
		chart.SeriesFrom("Northgate", chart.FormLine,
			[]float64{2, 14, 96, 40, 61},
			[]float64{55, 29, 41, 30, 78}),
		chart.SeriesFrom("Harbour", chart.FormLine,
			[]float64{7, 9, 240, 15, 44},
			[]float64{20, 63, 8, 88, 52}),
	}
	chart.ParallelCoordinates(c, chart.ParallelOptions{
		ChartOptions: chart.ChartOptions{
			Height: 230, Grid: true, Label: "jobs side by side",
			Legend: chart.LegendOptions{Entries: chart.EntriesOf(c, parallel)},
		},
		Axes:   []string{"Jobs", "Minutes", "Cost", "Days open", "Satisfaction"},
		Series: parallel, Rings: 4,
	})

	showcase.Field(c, "ParetoChart — 排序后的柱，加上跑到 100% 的那条线")
	chart.ParetoChart(c, chart.ParetoOptions{
		ChartOptions: chart.ChartOptions{
			Height: 230, Grid: true, Label: "which few matter",
			X: chart.AxisOptions{Count: 5, Label: "Cause"},
			Y: chart.AxisOptions{Count: 5, Side: chart.Left, Label: "Calls"},
		},
		Labels: chCauses, Values: chCounts,
	})

	showcase.Field(c, "RadarChart 与 PolarChart — 同样绕一圈，一个比多条辐条，一个比一条曲线")
	radar := []chart.Series{
		chart.SeriesFrom("This week", chart.FormArea,
			[]float64{0, 1, 2, 3, 4, 5}, []float64{82, 64, 91, 47, 73, 58}),
		chart.SeriesFrom("Last week", chart.FormArea,
			[]float64{0, 1, 2, 3, 4, 5}, []float64{70, 55, 78, 52, 61, 44}),
	}
	polar := []chart.Series{
		chart.SeriesFrom("Morning", chart.FormLine, chDays, chOpened),
		chart.SeriesFrom("Evening", chart.FormLine, chDays, chClosed),
	}
	chartTwoUp(c,
		func() {
			chart.RadarChart(c, chart.RadarOptions{
				ChartOptions: chart.ChartOptions{
					Height: 260, Label: "six measures, one shape",
					Legend: chart.LegendOptions{Entries: chart.EntriesOf(c, radar)},
				},
				Axes:   []string{"Reach", "Speed", "Cost", "Close", "Trust", "Repeat"},
				Series: radar, Rings: 4,
			})
		},
		func() {
			chart.PolarChart(c, chart.PolarOptions{
				ChartOptions: chart.ChartOptions{
					Height: 260, Label: "a day that repeats",
					Legend: chart.LegendOptions{Entries: chart.EntriesOf(c, polar)},
					// The same week as everywhere else, so the numbers round
					// the circle are the same numbers the other charts read.
					X: chart.AxisOptions{Scale: chDayScale(), Label: "Day"},
					Y: chart.AxisOptions{Side: chart.Left, Count: 4, Label: "Callbacks"},
				},
				Series: polar, Rings: 4,
			})
		})
}

// ── 8. share of a whole ────────────────────────────────────────────────────

func chartShare(c *ui.Context) {
	showcase.Section(c, "占比 · FunnelChart / PieChart / DonutChart / NightingaleChart / Treemap / SunburstChart")

	showcase.Field(c, "FunnelChart — 每一级有多宽，就是有多少走到了那一级")
	chart.FunnelChart(c, chart.FunnelOptions{
		ChartOptions: chart.ChartOptions{Height: 270, Label: "from call to invoice"},
		Steps: []chart.FunnelStep{
			{Label: "Calls", Value: 420},
			{Label: "Booked", Value: 318},
			{Label: "Attended", Value: 244},
			{Label: "Invoiced", Value: 197},
			{Label: "Paid", Value: 182},
		},
	})

	showcase.Field(c, "PieChart、DonutChart 与 NightingaleChart — 同一批数三种画法")
	pie := chart.PieOptions{
		ChartOptions: chart.ChartOptions{Height: 300},
		Labels:       chChannels, Values: chVolume,
	}
	chartThree(c, 3,
		func() { chart.PieChart(c, pie) },
		func() {
			chart.DonutChart(c, chart.DonutOptions{
				PieOptions: pie, Hole: 0.62,
			})
		},
		func() {
			chart.NightingaleChart(c, chart.NightingaleOptions{PieOptions: pie})
		})

	showcase.Field(c, "Treemap 与 SunburstChart — 一块一个与一圈一圈，同一棵树")
	tiles := []chart.Tile{
		{Label: "Phone", Value: 38}, {Label: "Email", Value: 24},
		{Label: "Walk-in", Value: 17}, {Label: "Portal", Value: 12},
		{Label: "Chat", Value: 9},
	}
	chartTwoUp(c,
		func() {
			chart.Treemap(c, chart.TreemapOptions{
				ChartOptions: chart.ChartOptions{Height: 230, Label: "area is the value"},
				Tiles:        tiles, Labels: true,
			})
		},
		func() {
			chart.SunburstChart(c, chart.SunburstOptions{
				ChartOptions: chart.ChartOptions{Height: 230, Label: "angle is the value"},
				Root: chart.Node{Label: "Calls", Value: 100, Children: []chart.Node{
					{Label: "Phone", Value: 38},
					{Label: "Email", Value: 24, Children: []chart.Node{
						{Label: "Outbound", Value: 15}, {Label: "Inbound", Value: 9}}},
					{Label: "Walk-in", Value: 17},
					{Label: "Portal", Value: 12},
					{Label: "Chat", Value: 9},
				}},
			})
		})
}

// ── 9. money and time ──────────────────────────────────────────────────────

func chartMoney(c *ui.Context) {
	showcase.Section(c, "价与账 · CandlestickChart / WaterfallChart")

	showcase.Field(c, "CandlestickChart — 涨空心、跌实心，和纸上的画法一样")
	chart.CandlestickChart(c, chart.CandlestickOptions{
		ChartOptions: chart.ChartOptions{
			Height: 230, Grid: true, Label: "tickets resolved per day",
			X: chart.AxisOptions{Scale: chDayScale(), Label: "Day"},
			Y: chart.AxisOptions{Side: chart.Left, Format: chart.Compact(),
				Label: "USD"},
		},
		Candles: []chart.Candle{
			{Label: "Mon", Open: 118, Close: 126, High: 131, Low: 112},
			{Label: "Tue", Open: 126, Close: 121, High: 134, Low: 118},
			{Label: "Wed", Open: 121, Close: 139, High: 142, Low: 119},
			{Label: "Thu", Open: 139, Close: 135, High: 148, Low: 130},
			{Label: "Fri", Open: 135, Close: 152, High: 158, Low: 133},
			{Label: "Sat", Open: 152, Close: 149, High: 156, Low: 144},
			{Label: "Sun", Open: 149, Close: 163, High: 168, Low: 147},
		},
	})

	showcase.Field(c, "WaterfallChart — 一路走到最后的那笔；Total 多加一根从零起的")
	chart.WaterfallChart(c, chart.WaterfallOptions{
		ChartOptions: chart.ChartOptions{
			Height: 220, Grid: true, Label: "revenue movement",
			X: chart.AxisOptions{Count: 5, Label: "Step"},
			Y: chart.AxisOptions{Count: 5, Side: chart.Left, Format: chart.Compact(), Label: "USD"},
		},
		Labels: []string{"Opening", "Calls", "Parts", "Travel", "No-shows", "Closing"},
		Values: []float64{9800, 12400, -3100, -1800, -2400, 9200},
		Total:  true,
	})
}

// ── 10. one number ─────────────────────────────────────────────────────────

func chartDial(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "单一数字 · Gauge / BulletChart")

	// Fixed widths: a dial and a bullet bar are things that size themselves,
	// and a row that filled the page would stretch one of them.
	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 4)).AlignItems(ui.Start).Children(func() {
		chart.Gauge(c, chart.GaugeOptions{
			Value: 72, Min: 0, Max: 100, Width: unit(c, 30), Height: unit(c, 18),
			Label: "72% within budget",
			Zones: []chart.GaugeZone{
				{To: 60, Name: "under"},
				{To: 85, Name: "close"},
				{To: 100, Name: "over"},
			},
		})
		chart.Gauge(c, chart.GaugeOptions{
			Value: 118, Min: 0, Max: 150, Width: unit(c, 30), Height: unit(c, 18),
			Label: "3 of 5 seats taken", Color: k.Warning,
			Zones: []chart.GaugeZone{
				{To: 60, Name: "free", Color: k.Success},
				{To: 120, Name: "half", Color: k.Accent},
				{To: 150, Name: "full", Color: k.Warning},
			},
		})
		ui.Column(c).Grow(1).Shrink(0).Gap(unit(c, 2)).Children(func() {
			chart.BulletChart(c, chart.BulletOptions{
				Value: 72, Min: 0, Max: 100, Target: 80, Label: "Resolution rate",
				Bands: []chart.BulletBand{
					{Min: 0, Max: 50, Name: "poor"},
					{Min: 50, Max: 80, Name: "fine"},
					{Min: 80, Max: 100, Name: "good"},
				},
			})
			chart.BulletChart(c, chart.BulletOptions{
				Value: 46, Min: 0, Max: 100, Target: 35, Label: "Repeat failures",
				Color: k.Danger,
				Bands: []chart.BulletBand{
					{Min: 0, Max: 35, Name: "good"},
					{Min: 35, Max: 70, Name: "watch"},
					{Min: 70, Max: 100, Name: "bad"},
				},
			})
		})
	})
}

// ── 11. flow ───────────────────────────────────────────────────────────────

func chartFlow(c *ui.Context) {
	showcase.Section(c, "流动 · AlluvialChart / SankeyChart / ChordDiagram")

	showcase.Field(c, "AlluvialChart — 每一级都把留下的和传下去的分开说")
	chart.AlluvialChart(c, chart.AlluvialOptions{
		ChartOptions: chart.ChartOptions{Height: 250, Label: "stage by stage"},
		Stages:       []string{"Calls", "Booked", "Attended", "Invoiced", "Paid"},
		Flows: [][]float64{
			{318, 62, 28, 12},
			{244, 48, 26},
			{197, 33, 14},
			{182},
		},
	})

	showcase.Field(c, "SankeyChart 与 ChordDiagram — 同一批数，两种摆法")
	links := []chart.Link{
		{Source: 0, Target: 2, Value: 118}, {Source: 0, Target: 3, Value: 64},
		{Source: 1, Target: 3, Value: 92}, {Source: 1, Target: 2, Value: 41},
		{Source: 2, Target: 3, Value: 26},
	}
	matrix := [][]float64{
		{0, 18, 12, 38},
		{14, 0, 10, 24},
		{9, 11, 0, 17},
		{33, 26, 21, 0},
	}
	chartTwoUp(c,
		func() {
			chart.SankeyChart(c, chart.SankeyOptions{
				ChartOptions: chart.ChartOptions{Height: 250, Label: "who pays for what"},
				Names:        []string{"Phone", "Email", "Walk-in", "Riverside"},
				Links:        links,
			})
		},
		func() {
			chart.ChordDiagram(c, chart.ChordOptions{
				ChartOptions: chart.ChartOptions{Height: 250, Label: "who talks to whom"},
				Names:        []string{"Phone", "Email", "Walk-in", "Portal"},
				Matrix:       matrix,
			})
		})
}

// ── 12. relations ──────────────────────────────────────────────────────────

func chartRelation(c *ui.Context) {
	showcase.Section(c, "关系 · NetworkGraph / DecompositionTree")

	showcase.Field(c, "NetworkGraph — 结点按 group 聚成一段一段，边宽是流量")
	nodes := []chart.GraphNode{
		{Label: "Router", Value: 40, Group: 0},
		{Label: "Engine", Value: 34, Group: 0},
		{Label: "Docs", Value: 18, Group: 0},
		{Label: "Auth", Value: 26, Group: 1},
		{Label: "API", Value: 30, Group: 1},
		{Label: "Billing", Value: 22, Group: 2},
		{Label: "Mail", Value: 14, Group: 2},
		{Label: "SMS", Value: 9, Group: 2},
	}
	graphLinks := []chart.GraphLink{
		{Source: 0, Target: 3, Value: 9}, {Source: 0, Target: 4, Value: 7},
		{Source: 1, Target: 4, Value: 12}, {Source: 1, Target: 3, Value: 6},
		{Source: 3, Target: 5, Value: 8}, {Source: 4, Target: 5, Value: 11},
		{Source: 2, Target: 6, Value: 4}, {Source: 5, Target: 7, Value: 3},
		{Source: 3, Target: 7, Value: 2},
	}
	chartTwoUp(c,
		func() {
			chart.NetworkGraph(c, chart.NetworkOptions{
				ChartOptions: chart.ChartOptions{Height: 270, Label: "modules and their callers"},
				Nodes:        nodes, Links: graphLinks, Curvature: 0.35,
			})
		},
		func() {
			chart.DecompositionTree(c, chart.DecompositionOptions{
				ChartOptions: chart.ChartOptions{Height: 270, Label: "the same tree, in rows"},
				Root: chart.Node{Label: "All", Value: 100, Children: []chart.Node{
					{Label: "Retail", Value: 46, Children: []chart.Node{
						{Label: "Bakery", Value: 24}, {Label: "Gym", Value: 22}}},
					{Label: "Medical", Value: 34, Children: []chart.Node{
						{Label: "Clinic", Value: 20}, {Label: "Dental", Value: 14}}},
					{Label: "Other", Value: 20},
				}},
			})
		})
}

// ── 13. the pieces every chart is made of ───────────────────────────────────

func chartPieces(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "零件 · Frame / Axis / Grid / Legend / Tooltip / Crosshair / Brush / Annotation / RangeHighlight / Palette")

	showcase.Field(c, "Frame — 一张自己拼的图：两根轴、网格、线、参考线、十字线、提示")
	chartFrameDemo(c)

	showcase.Field(c, "Legend — 横排、竖排，以及手动写死的那一份")
	series := chSeries(c)
	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 4)).AlignItems(ui.Start).Children(func() {
		ui.Box(c).Width(unit(c, 116)).Children(func() {
			chart.Legend(c, chart.LegendOptions{
				Entries: chart.EntriesOf(c, series),
			})
		})
		ui.Box(c).Width(unit(c, 30)).Children(func() {
			chart.Legend(c, chart.LegendOptions{
				Entries: chart.EntriesOf(c, series), Vertical: true,
			})
		})
		ui.Column(c).Width(unit(c, 78)).Gap(unit(c, 0.5)).Children(func() {
			ui.Text(c, "Mark 三种记号：方块、圆点、短线").TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize))
			chart.Legend(c, chart.LegendOptions{Entries: []chart.LegendEntry{
				{Name: "square", Color: k.Accent, Mark: chart.MarkSquare},
				{Name: "dot", Color: k.Success, Mark: chart.MarkDot},
				{Name: "line", Color: k.Danger, Mark: chart.MarkLine},
			}})
		})
	})

	showcase.Field(c, "Palette — 从窗口的强调色退出来的一把梯子，明暗两套界面用的是同一把")
	pal := chart.Palette(c, 6)
	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 1)).Children(func() {
		for i, col := range pal {
			col := col
			ui.Column(c).Width(unit(c, 12)).Gap(unit(c, 0.5)).Children(func() {
				ui.Box(c).FillWidth().Height(unit(c, 9)).Radius(theme.SmallRadius).
					Background(col)
				ui.Text(c, fmt.Sprintf("P%d", i+1)).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize))
			})
		}
		ui.Column(c).Width(unit(c, 14)).Gap(unit(c, 0.5)).Children(func() {
			ui.Box(c).FillWidth().Height(unit(c, 9)).Radius(theme.SmallRadius).
				Background(k.Border)
			ui.Text(c, "6 colours").TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize))
		})
	})

	showcase.Field(c, "RangeHighlight 与 Annotation — 「这周」和「目标 8 单」")
	chartRangeAndAnnotation(c)
}

// chartFrameDemo builds one frame out of the parts by hand, which is the only
// way a gallery can show that Grid, Plot, Axis, Crosshair, Brush, Tooltip and
// Annotation are six separate things rather than one component wearing six
// hats.
//
// The pointer is pinned: a headless render has no pointer, so the crosshair
// would have nothing to follow and the page would show a chart with all its
// parts missing. Pinning it is what the Tooltip's own At field is for.
func chartFrameDemo(c *ui.Context) {
	series := chSeries(c)[:2]
	d, _ := chart.SeriesDomain(series)
	xs := chDayScale()
	ys := chart.NewLinear(chart.Nice(d.Min, d.Max, 4), 0, 1)
	xAxis := chart.AxisOptions{Scale: xs, Label: "Day"}
	yAxis := chart.AxisOptions{Side: chart.Left, Scale: ys, Count: 5, Label: "Callbacks"}

	var hover chart.Hover
	var pinnedX, pinnedY string

	chart.Frame(c, chart.FrameOptions{
		Height: 240, Label: "a week of callbacks",
		X: xAxis, Y: yAxis,
		Legend: chart.LegendOptions{Entries: chart.EntriesOf(c, series)},
		Hover:  &hover,
		Border: true, Surface: true,
	}, func(frame chart.FrameResult) {
		// Put the pointer somewhere real before anything that follows it
		// draws. This is a page, not a test: the picture has to show the
		// crosshair and the tooltip, and neither exists without a pointer.
		ui.Box(c).Absolute().Fill().Draw(func(_ *ui.Painter, _ ui.Rect) {
			plot := frame.Plot()
			hover.Over = plot.W > 0 && plot.H > 0
			hover.X = plot.X + plot.W*0.58
			hover.Y = plot.Y + plot.H*0.28
			pinnedX = fmt.Sprintf("%.1f", hover.ValueX)
			pinnedY = fmt.Sprintf("%.0f", hover.ValueY)
		})

		chart.Grid(c, chart.GridOptions{Frame: frame, X: xAxis, Y: yAxis})
		chart.Annotation(c, chart.AnnotationOptions{
			Frame: frame, X: xs, Y: ys,
			Line:  &chart.Line{Value: 8, Text: "target", Dashed: true},
			Label: &chart.Callout{At: chart.Point{X: 5, Y: 12}, Text: "peak"},
		})
		for _, s := range series {
			chart.Plot(c, chart.PlotOptions{Frame: frame, Series: s, X: xs, Y: ys, Dots: true})
		}
		// A brush with no drag in it draws nothing at all, which is the
		// point: it is a hit area over the data, not a picture. The
		// selection it would leave is drawn by RangeHighlight below.
		chart.Brush(c, chart.BrushOptions{Frame: frame, X: xs, Range: &chart.Range{}, Handles: true})
		chart.Crosshair(c, chart.CrosshairOptions{
			Frame: frame, Hover: &hover, X: xs, Y: ys,
			ValueX: pinnedX, ValueY: pinnedY,
		})
		chart.Tooltip(c, chart.TooltipOptions{
			Frame: frame, Hover: &hover, X: xs, Y: ys,
			Title: "Friday",
			Rows: []chart.TooltipRow{
				{Name: "Opened", Color: chart.EntriesOf(c, series)[0].Color, Value: "12"},
				{Name: "Resolved", Color: chart.EntriesOf(c, series)[1].Color, Value: "7"},
			},
		})
		chart.Axis(c, frame, xAxis)
		chart.Axis(c, frame, yAxis)
	})
}

// chartRangeAndAnnotation shows the two "about this chart rather than of it"
// pieces over the same week, one with a range washed across it and one with a
// target drawn across it.
func chartRangeAndAnnotation(c *ui.Context) {
	series := chSeries(c)[:2]
	d, _ := chart.SeriesDomain(series)
	xs := chDayScale()
	ys := chart.NewLinear(chart.Nice(d.Min, d.Max, 4), 0, 1)
	xAxis := chart.AxisOptions{Scale: xs, Label: "Day"}
	yAxis := chart.AxisOptions{Side: chart.Left, Scale: ys, Count: 5, Label: "Callbacks"}

	chartTwoUp(c,
		func() {
			chart.Frame(c, chart.FrameOptions{
				Height: 210, Label: "range highlight",
				X: xAxis, Y: yAxis,
			}, func(frame chart.FrameResult) {
				chart.RangeHighlight(c, chart.RangeHighlightOptions{
					Frame: frame, X: xs, From: 1, To: 4,
					Label: "Tue to Thu", Edges: true,
				})
				chart.Annotation(c, chart.AnnotationOptions{
					Frame: frame, X: xs, Y: ys,
					Line: &chart.Line{Value: 8, Text: "target", Dashed: true},
				})
				for _, s := range series {
					chart.Plot(c, chart.PlotOptions{Frame: frame, Series: s, X: xs, Y: ys})
				}
				chart.Axis(c, frame, xAxis)
				chart.Axis(c, frame, yAxis)
			})
		},
		func() {
			chart.Frame(c, chart.FrameOptions{
				Height: 210, Label: "two annotations",
				X: xAxis, Y: yAxis,
			}, func(frame chart.FrameResult) {
				chart.Grid(c, chart.GridOptions{Frame: frame, X: xAxis, Y: yAxis})
				chart.Annotation(c, chart.AnnotationOptions{
					Frame: frame, X: xs, Y: ys,
					Line: &chart.Line{Value: 8, Text: "target", Dashed: true},
				})
				chart.Annotation(c, chart.AnnotationOptions{
					Frame: frame, X: xs, Y: ys,
					Label: &chart.Callout{At: chart.Point{X: 3, Y: 9}, Text: "outage", Dx: 8, Dy: -12},
				})
				for _, s := range series {
					chart.Plot(c, chart.PlotOptions{Frame: frame, Series: s, X: xs, Y: ys, Dots: true})
				}
				chart.Axis(c, frame, xAxis)
				chart.Axis(c, frame, yAxis)
			})
		})
}

// ── 14. nothing to draw ────────────────────────────────────────────────────

func chartEmpty(c *ui.Context) {
	showcase.Section(c, "空数据 · Empty — 有数据画不出来时唯一该画的东西")

	showcase.Field(c, "一行类型空着，和 chart.Empty 直接画在框里")
	chartTwoUp(c,
		func() {
			chart.LineChart(c, chart.LineOptions{
				ChartOptions: chart.ChartOptions{
					Height: 210, Label: "empty line chart",
					EmptyTitle: "Nothing to plot",
					EmptyBody:  "no callbacks were resolved this week",
				},
			})
		},
		func() {
			chart.BarChart(c, chart.BarOptions{
				ChartOptions: chart.ChartOptions{
					Height: 210, Border: true, Surface: true, Label: "empty frame",
					EmptyTitle: "Nothing to plot",
					EmptyBody:  "a bar chart of nothing is still a chart",
				},
			})
		})
	showcase.Field(c, "chart.Empty 直接画在自己的框里")
	chart.Frame(c, chart.FrameOptions{
		Height: 130, Surface: true, Border: true, Label: "empty state",
	}, func(frame chart.FrameResult) {
		chart.Empty(c, chart.EmptyOptions{
			Frame: frame,
			Title: "Nothing to plot",
			Body:  "the axis would be a scale over a range no data reaches",
		})
	})
}

// ── 15. the pure functions ─────────────────────────────────────────────────

// chartFunctions is the half of the package a person cannot see in a chart:
// the arithmetic underneath, with its answers written out. A caller sizing a
// slider under a chart, or a summary line beside it, uses these and needs to
// be sure they agree with what the chart draws.
func chartFunctions(c *ui.Context) {
	showcase.Section(c, "纯函数 · 画之前算的那些数，每一行右边是它真正的返回值")

	stacked := [][]float64{chOpened, chClosed}
	rows := chart.Bins(chLatency, 3)
	st, _ := chart.Summary(chLatency)
	latencyDomain, _ := chart.DomainOf(chLatency)
	seriesDomain, _ := chart.SeriesDomain(chSeries(c))
	// Measure with real scales in it, because that is the whole point of it:
	// it is the function a frame uses to work out its own gutters, and a
	// caller can size whatever it is drawing around the chart from the answer.
	d2, _ := chart.SeriesDomain(chSeries(c))
	inset := chart.Measure(c, chart.FrameOptions{
		X: chart.AxisOptions{Scale: chDayScale()},
		Y: chart.AxisOptions{
			Side:  chart.Left,
			Scale: chart.NewLinear(chart.Nice(d2.Min, d2.Max, 4), 0, 1), Count: 5,
		},
		Legend: chart.LegendOptions{Entries: chart.EntriesOf(c, chSeries(c))},
	})
	ticks := chart.Ticks(c, chart.NewLinear(chart.Nice(0, 60, 4), 0, 400),
		chart.TickOptions{Count: 4})

	for _, f := range []struct{ name, value string }{
		{"Nice(0, 97, 5)", fmt.Sprintf("%v", chart.Nice(0, 97, 5))},
		{"Bins(latency, 3)", fmt.Sprintf("%d 桶 · %g…%g · %d 个",
			len(rows), rows[0].Lo, rows[len(rows)-1].Hi, rows[0].Count)},
		{"Stats / Summary(latency)", fmt.Sprintf("min %g · q1 %g · med %g · q3 %g · max %g",
			st.Min, st.Q1, st.Median, st.Q3, st.Max)},
		{"Cumulative([38 27 21 14])", fmt.Sprintf("%.0f%%", chart.Cumulative(chVolume)[len(chVolume)-1])},
		{"Stack(opened+closed)", fmt.Sprintf("第 2 列 = %.0f%%", chart.Stack(stacked)[0][1])},
		{"Descending([38 27 21 14])", fmt.Sprintf("最大的是 %s", chartSorted(chVolume, chChannels))},
		{"Squarify(tiles, 400×220)", fmt.Sprintf("%d 块 · 最大 %d×%d",
			len(chart.Squarify(chVolume, ui.Rect{W: 400, H: 220})),
			int(chart.Squarify(chVolume, ui.Rect{W: 400, H: 220})[0].W),
			int(chart.Squarify(chVolume, ui.Rect{W: 400, H: 220})[0].H))},
		{"Sweep([38 27 21 14])", fmt.Sprintf("加起来 %.1f°", sum(chart.Sweep(chVolume)))},
		{"Depths(5, sankey links)", fmt.Sprint(chart.Depths(5, []chart.Link{
			{Source: 0, Target: 3, Value: 1}, {Source: 1, Target: 4, Value: 1},
		}))},
		{"HeatColor(40, 0, 60)", fmt.Sprintf("40 在 0…60 之间是 rgb(%d %d %d)",
			chart.HeatColor(40, 0, 60, core.Tokens(c)).R,
			chart.HeatColor(40, 0, 60, core.Tokens(c)).G,
			chart.HeatColor(40, 0, 60, core.Tokens(c)).B)},
		{"DomainOf(latency)", fmt.Sprintf("%.0f … %.0f", latencyDomain.Min, latencyDomain.Max)},
		{"SeriesDomain(series)", fmt.Sprintf("%.0f … %.0f", seriesDomain.Min, seriesDomain.Max)},
		{"Linear(0, 100, 4)", fmt.Sprint(chart.Linear(chart.Domain{Min: 0, Max: 100}, 4))},
		{"Log(1, 1000, 10)", fmt.Sprint(chart.Log(1, 1000, 10))},
		{"Bounds(stacked)", chartBounds(stacked)},
		{"Ends([38 27 21 14])", fmt.Sprintf("%.0f° … %.0f°",
			chart.Ends(chVolume)[0], chart.Ends(chVolume)[len(chVolume)-1])},
		{"Palette(6)", fmt.Sprintf("一把 %d 色的梯子", len(chart.Palette(c, 6)))},
		{"Measure(frame)", fmt.Sprintf("上 %.0f 左 %.0f 下 %.0f 右 %.0f",
			inset.Top, inset.Left, inset.Bottom, inset.Right)},
		{"Ticks(y)", fmt.Sprintf("%d 条刻度：%v", len(ticks), tickLabels(ticks))},
	} {
		chartFnRow(c, f.name, f.value)
	}
}

// chartFnRow is one line of the pure-function table: the call on the left, its
// answer on the right. The name column has a fixed width so the answers line
// up, and the row has a width of its own rather than filling the page — a
// table whose last column is stretched to the window edge reads as cut off
// rather than as padded.
func chartFnRow(c *ui.Context, name, value string) {
	k := core.Tokens(c)
	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 2)).AlignItems(ui.Center).Children(func() {
		ui.Text(c, name).TextColor(k.Text).
			FontSize(core.FontSize(c, theme.RowSize)).Width(unit(c, 44)).Shrink(0)
		ui.Text(c, value).TextColor(k.TextMuted).
			FontSize(core.FontSize(c, theme.CaptionSize))
	})
}

// chartSorted is the label of the biggest category, which is what a stable
// descending sort puts first.
func chartSorted(values []float64, labels []string) string {
	_, sorted := chart.Descending(values, labels)
	if len(sorted) == 0 {
		return "—"
	}
	return sorted[0]
}

// chartBounds is what one column of a stack fills: its bottom and its top.
func chartBounds(rows [][]float64) string {
	lower, upper := chart.Bounds(rows)
	if len(lower) == 0 || len(lower[0]) < 4 {
		return "—"
	}
	return fmt.Sprintf("列 3 的底 %.0f 顶 %.0f", lower[0][3], upper[0][3])
}

func tickLabels(ticks []chart.Tick) []string {
	out := make([]string, 0, len(ticks))
	for _, t := range ticks {
		out = append(out, t.Label)
	}
	return out
}

func sum(v []float32) float32 {
	var t float32
	for _, x := range v {
		t += x
	}
	return t
}

// ── 16. what the package is for ─────────────────────────────────────────────

func chartConventions(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "约定")
	ui.Column(c).Width(unit(c, 236)).Gap(unit(c, 1)).Children(func() {
		for _, line := range []string{
			"组件不持有状态：指针在调用方的 Hover，框选在调用方的 Range。",
			"数据是调用方的：本包不生成任何一个数，也不读时钟。",
			"颜色是 token：只有 series 从 Palette 拿颜色，明暗两套界面用的是同一把梯子。",
			"刻度尺是从最大的那个数量的出来的（Measure / Widest），不是猜的。",
			"画不了的东西 panic(\"chart: …\")，不是画一个看起来差不多的空盒子。",
			"每个图表的 Label 是读屏唯一能拿到的东西，所以这一页每个都写了。",
		} {
			ui.Text(c, line).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.RowSize))
		}
	})
}

// ── layout helpers ─────────────────────────────────────────────────────────

// chartTwoUp puts two things side by side, each taking half the page's own
// width less the gap between them.
func chartTwoUp(c *ui.Context, first, second func()) {
	chartThree(c, 2, first, second)
}

// chartThree puts n things side by side. A filling row with equal growers in
// it is the one arrangement that is safe: they split what is there, so the
// last one cannot be pushed off the page's edge and clipped.
func chartThree(c *ui.Context, n int, children ...func()) {
	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		for range n {
			fn := children[0]
			children = children[1:]
			ui.Box(c).Grow(1).Shrink(0).Children(fn)
		}
	})
}
