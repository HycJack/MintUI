package chart

import (
	"math"
	"sort"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/theme"
)

// This file is the arithmetic a chart type needs and a window does not: the
// parts of drawing that can be settled without one. Every function here is
// pure — same numbers in, same numbers out, no context, no tokens, no clock —
// so a test can pin the answer down exactly rather than watching a rectangle
// and hoping. The parts that need a palette or a plot area live in the chart
// files beside the types that draw them.

// ── stacks ─────────────────────────────────────────────────────────────────

// Stack turns stacked values into percentages: every column comes to 100,
// its rows sharing it in proportion to what they held.
//
// It is how a percent bar chart, a stacked area and a 100% stacked column all
// say the same thing without three copies of the arithmetic. The result has
// the shape it was given: a column that only some rows reach fills the rows
// that reach it, because a missing value is a value of nothing.
//
// A column whose values come to nothing — all of them zero, or one up and one
// down — has no shares to give, and stays at zero rather than dividing by its
// own zero into a chart full of NaN.
func Stack(series [][]float64) [][]float64 {
	if len(series) == 0 {
		return nil
	}
	out := make([][]float64, len(series))
	columns := 0
	for i, row := range series {
		out[i] = make([]float64, len(row))
		columns = max(columns, len(row))
	}
	for j := range columns {
		total := 0.0
		for _, row := range series {
			if j < len(row) {
				total += row[j]
			}
		}
		if total == 0 {
			continue
		}
		for i, row := range series {
			if j < len(row) {
				out[i][j] = row[j] / total * 100
			}
		}
	}
	return out
}

// Bounds is what each column of a set of stacked series fills: where its
// bottom sits, and where its top does. The bottom of the first row is the
// baseline, so a stacked bar or a waterfall reads off two numbers per column
// rather than walking the stack again at draw time.
func Bounds(series [][]float64) (lower, upper [][]float64) {
	if len(series) == 0 {
		return nil, nil
	}
	columns := 0
	for _, row := range series {
		columns = max(columns, len(row))
	}
	lower = make([][]float64, len(series))
	upper = make([][]float64, len(series))
	for i := range series {
		lower[i] = make([]float64, len(series[i]))
		upper[i] = make([]float64, len(series[i]))
	}
	below := make([]float64, columns)
	for j := range columns {
		// The stack is walked column by column rather than row by row: a
		// bar chart reads its rows in order, and so does its height.
		below[j] = 0
		for i, row := range series {
			if j >= len(row) {
				continue
			}
			lower[i][j] = below[j]
			below[j] += row[j]
			upper[i][j] = below[j]
		}
	}
	return lower, upper
}

// Running is the running total of values: what a waterfall's bars stand on and
// what a bar's own length would be if it were drawn on its own.
func Running(values []float64) []float64 {
	if len(values) == 0 {
		return nil
	}
	out := make([]float64, len(values))
	sum := 0.0
	for i, v := range values {
		sum += v
		out[i] = sum
	}
	return out
}

// Cumulative is [Running] as a share of the whole, in percent: the line a
// Pareto chart draws over its bars, and the rule for finding the few
// categories that make up most of a total.
//
// The last value is set to 100 rather than whatever the division left there,
// so "these five of twelve are three quarters of it" is a fact about the
// answer and not a rounding error the caller has to know about.
func Cumulative(values []float64) []float64 {
	out := Running(values)
	total := 0.0
	for _, v := range values {
		total += v
	}
	if total == 0 || len(out) == 0 {
		// A column of nothing has no share of nothing to divide up.
		return out
	}
	for i := range out {
		out[i] = out[i] / total * 100
	}
	out[len(out)-1] = 100
	return out
}

// Descending is a set of values in order of size, largest first, with their
// labels carried along: the order a Pareto chart's bars go in, and the order
// any chart about "which few matter" wants them in.
//
// A sort that is stable, so two categories of the same value stay in the order
// the caller listed them — a chart that reordered equals would make two
// reports of the same data differ in a way nobody can account for.
func Descending(values []float64, labels []string) ([]float64, []string) {
	if len(values) == 0 {
		return nil, nil
	}
	if len(labels) != len(values) {
		panic("chart: Descending needs one label for every value")
	}
	out := make([]float64, len(values))
	copy(out, values)
	names := make([]string, len(labels))
	copy(names, labels)
	order := make([]int, len(values))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool { return values[order[a]] > values[order[b]] })
	sorted := make([]float64, len(values))
	sortedLabels := make([]string, len(labels))
	for i, at := range order {
		sorted[i], sortedLabels[i] = out[at], names[at]
	}
	return sorted, sortedLabels
}

// ── pies ───────────────────────────────────────────────────────────────────

// Sweep is how much of a circle each of values takes, in degrees: the angle a
// pie slice, a rose petal or a chord's arc spans. The angles come to 360, so
// the ring closes.
//
// The last one is set to whatever the others left, which is why the answer
// adds up exactly rather than to 359.99997 — a pie whose last slice does not
// meet its first has a seam down it that no amount of data deserves.
//
// Nothing at all, or values that come to nothing, is no angles: a chart with
// no data says so with [Empty] rather than drawing a full circle of NaN.
func Sweep(values []float64) []float32 {
	if len(values) == 0 {
		return nil
	}
	out := make([]float32, len(values))
	total := 0.0
	for _, v := range values {
		total += v
	}
	if total == 0 {
		return out
	}
	for i, v := range values {
		out[i] = float32(v * 360 / total)
	}
	closes(out, 360)
	return out
}

// Ends is where each of values ends round the circle, in degrees: the
// boundary of every slice, which is what a slice is drawn between. The first
// is where the last one began, and the last is 360 — the same place.
func Ends(values []float64) []float32 {
	if len(values) == 0 {
		return nil
	}
	out := make([]float32, len(values))
	total := 0.0
	for _, v := range values {
		total += v
	}
	if total == 0 {
		return out
	}
	at := 0.0
	for i, v := range values {
		at += v
		out[i] = float32(at * 360 / total)
	}
	out[len(out)-1] = 360
	return out
}

// closes gives the last of a run of angles the share its predecessors left,
// so that a set of angles adds up to a whole turn: one rounded to the last,
// rather than one rounded short by a gap down the seam of a pie.
//
// The rounding is done once, on the exact remainder. Subtracting two float32s
// rounds twice and leaves a seam half an ulp wide, which is enough for two
// slices that should meet not quite meeting.
func closes(angles []float32, whole float32) {
	n := len(angles)
	if n == 0 {
		return
	}
	at := 0.0
	for _, a := range angles[:n-1] {
		at += float64(a)
	}
	angles[n-1] = float32(float64(whole) - at)
}

// ── axes that start at zero ────────────────────────────────────────────────

// FromZero widens d to zero when its values do not already cross it: a bar or
// a histogram that stands on its own smallest value says nothing about it,
// because the length of the bar is read as a length and a length means nothing
// without a zero to count from.
func FromZero(d Domain) Domain {
	switch {
	case !d.Valid():
		return d
	case d.Min > 0:
		return Domain{Min: 0, Max: d.Max}
	case d.Max < 0:
		return Domain{Min: d.Min, Max: 0}
	}
	return d
}

// ── distributions ─────────────────────────────────────────────────────────

// Bin is one bucket of a histogram: the range it covers and how many values
// fell in it.
type Bin struct {
	// Lo and Hi are the bucket's two edges, in the values' own units.
	Lo, Hi float64
	// Count is how many values fell between them.
	Count int
}

// Width is how much room a bin covers.
func (b Bin) Width() float64 { return b.Hi - b.Lo }

// Middle is where a bin's label sits on a categorical axis, so that a label
// names the bucket rather than its edge.
func (b Bin) Middle() float64 { return (b.Lo + b.Hi) / 2 }

// Bins splits values into count buckets of equal width over their span, which
// is the one histogram rule there is: everything below the smallest, and
// everything above the largest, belongs in the ends.
//
// Values that all sit on one number are one bucket rather than a row of empty
// ones — there is one thing the data says, and a histogram of it says that
// once.
func Bins(values []float64, count int) []Bin {
	d, ok := DomainOf(values)
	if !ok {
		return nil
	}
	count = max(count, 1)
	out := make([]Bin, count)
	width := d.Span() / float64(count)
	for i := range out {
		out[i] = Bin{Lo: d.Min + width*float64(i), Hi: d.Min + width*float64(i+1)}
	}
	if width == 0 {
		// Every reading is on the same number: there is one thing to say
		// about them, and a histogram of it says it once.
		return []Bin{{Lo: d.Min, Hi: d.Max, Count: len(values)}}
	}
	for _, v := range values {
		i := int(math.Floor((v - d.Min) / width))
		// The value at the very top belongs in the last bucket: it is not
		// past the data, it is the end of it.
		out[min(max(i, 0), count-1)].Count++
	}
	return out
}

// Stats is the five-number summary of a sample: the box and the whiskers a
// box plot draws, and the numbers its tooltip quotes.
type Stats struct {
	// Min, Q1, Median, Q3 and Max are the five: the box runs from Q1 to Q3
	// with the median in it, and the whiskers reach out to the ends.
	Min, Q1, Median, Q3, Max float64
	// Lower and Upper are where the whiskers actually stop: the furthest
	// values still inside 1.5 interquartile ranges of the box, so that one
	// wild reading stretches a whisker rather than deciding the scale.
	Lower, Upper float64
	// Outliers are the values outside them, drawn as points of their own.
	Outliers []float64
}

// Summary is the five-number summary of values. It reports false for a sample
// with no numbers in it, which is the same answer [DomainOf] gives and for
// the same reason: nothing to summarise is not a summary of zero.
func Summary(values []float64) (Stats, bool) {
	sorted := make([]float64, 0, len(values))
	for _, v := range values {
		if !math.IsNaN(v) && !math.IsInf(v, 0) {
			sorted = append(sorted, v)
		}
	}
	if len(sorted) == 0 {
		return Stats{}, false
	}
	sort.Float64s(sorted)
	s := Stats{
		Min:    sorted[0],
		Q1:     quantile(sorted, 0.25),
		Median: quantile(sorted, 0.5),
		Q3:     quantile(sorted, 0.75),
		Max:    sorted[len(sorted)-1],
	}
	// The fences are 1.5 boxes out from the box. Everything inside them
	// reaches the whisker; everything outside is an outlier.
	lo, hi := s.Q1-1.5*(s.Q3-s.Q1), s.Q3+1.5*(s.Q3-s.Q1)
	s.Lower, s.Upper = math.Inf(1), math.Inf(-1)
	for _, v := range sorted {
		switch {
		case v < lo, v > hi:
			s.Outliers = append(s.Outliers, v)
		default:
			s.Lower, s.Upper = minf64(s.Lower, v), maxf64(s.Upper, v)
		}
	}
	return s, true
}

// quantile is the p-th value of a sorted sample, interpolated the way a
// spreadsheet does it — the 25th percentile of 1…9 is 3, of 1…8 it is 2.75 —
// so that the answer is a number a caller can reproduce by hand.
func quantile(sorted []float64, p float64) float64 {
	n := len(sorted)
	if n == 0 {
		return 0
	}
	if n == 1 {
		return sorted[0]
	}
	at := float64(n-1) * p
	lo := int(math.Floor(at))
	if lo >= n-1 {
		return sorted[n-1]
	}
	return sorted[lo] + (at-float64(lo))*(sorted[lo+1]-sorted[lo])
}

// Density is a Gaussian kernel estimate of how much of a sample sits near v:
// what a violin plot's width is, and what a ridgeline's is.
//
// A kernel rather than a histogram because a violin is read as a shape, and a
// shape with steps in it is a histogram wearing a violin costume.
func Density(v float64, samples []float64, bandwidth float64) float64 {
	if len(samples) == 0 || bandwidth <= 0 {
		return 0
	}
	const norm = 1 / 2.5066282746310002 // 1 / √(2π)
	sum := 0.0
	for _, s := range samples {
		u := (v - s) / bandwidth
		sum += math.Exp(-0.5 * u * u)
	}
	return sum * norm / (bandwidth * float64(len(samples)))
}

// Column is the span of one axis of a set of series — the i-th value of each
// of them — and whether any of those values is a number. It is what a
// parallel-coordinates chart scales each of its axes by, since each axis has
// its own units and none of them can share a scale.
func Column(series []Series, i int) (Domain, bool) {
	values := make([]float64, 0, len(series))
	for _, s := range series {
		if i < len(s.Points) {
			values = append(values, s.Points[i].Y)
		}
	}
	return DomainOf(values)
}

// ── heat ───────────────────────────────────────────────────────────────────

// HeatColor is the colour a value takes on a heat scale: where the scale
// starts, faded right back to the surface, and where it ends, the accent.
//
// The ends are the tokens themselves rather than a colour mixed towards them,
// so the palest and fullest cells of a heatmap are exactly the two colours the
// rest of the interface uses — a chart that fades to a grey of its own is a
// chart with a second palette to keep in step.
//
// A value outside the range, a range that spans nothing, or a value that is
// not a number is the near end rather than a colour worked out from a
// division by nothing.
func HeatColor(v, min, max float64, k theme.Tokens) ui.Color {
	switch {
	case math.IsNaN(v) || max <= min, v <= min:
		return k.Surface
	case v >= max:
		return k.Accent
	}
	return k.Surface.Mix(k.Accent, float32((v-min)/(max-min)))
}

// ── shapes ─────────────────────────────────────────────────────────────────

// ArcAngle is where a value falls on an arc of span degrees starting at
// start: the angle of a gauge's needle. A value past either end is put at the
// end it ran past, because a needle off the end of its gauge is a needle that
// has stopped meaning anything.
func ArcAngle(value, lo, hi, start, span float32) float32 {
	if !(hi > lo) {
		return start
	}
	at := (value - lo) / (hi - lo)
	at = min(max(at, 0), 1)
	return start + at*span
}

// Body is the box a candle's open and close stand in: the higher of the two on
// top, the lower underneath, whatever order they opened and closed in.
func Body(open, close float64) (top, bottom float64) {
	if open >= close {
		return open, close
	}
	return close, open
}

// Squarify lays values out over r as rectangles of proportional area, each as
// nearly square as the set allows: the treemap layout that says "how much" by
// area and reads as "how much" because a square is the shape a reader's eye
// measures without being told to.
//
// The order is the caller's — squarify keeps it, because the order is where a
// treemap puts its categories (biggest first, or grouped by parent) and a
// layout that reordered them would silently undo that.
func Squarify(values []float64, r ui.Rect) []ui.Rect {
	out := make([]ui.Rect, len(values))
	total := 0.0
	for _, v := range values {
		if v > 0 && !math.IsNaN(v) {
			total += v
		}
	}
	if total <= 0 || r.W <= 0 || r.H <= 0 {
		return out
	}
	scale := r.W * r.H / float32(total)
	// live is what is left of r, walked down as rows are laid into it.
	live := r
	for i := 0; i < len(values); {
		row := []int{i}
		i++
		// A row keeps growing while the worst shape in it keeps getting
		// closer to a square; the first value that makes it worse is where
		// the row is placed and a new one starts.
		for i < len(values) && worstShape(append(row, i), values, live, scale) <= worstShape(row, values, live, scale) {
			row = append(row, i)
			i++
		}
		live = placeRow(row, values, live, scale, out)
	}
	return out
}

// placeRow puts one row of a squarified layout into the free rect and returns
// what is left of it. A row goes down the left of what is left with its tiles
// stacked, or along the top with them side by side, whichever of the two makes
// the squarer tiles — which is the whole reason the layout is called squarify.
func placeRow(row []int, values []float64, r ui.Rect, scale float32, out []ui.Rect) ui.Rect {
	areas := rowAreas(row, values, scale)
	if len(areas) == 0 || r.W <= 0 || r.H <= 0 {
		return r
	}
	stacked, beside := rowShapes(areas, r)
	if stacked <= beside {
		// A strip down the left, its tiles stacked down it.
		w := sumf(areas) / r.H
		y := r.Y
		for i, area := range areas {
			h := area / maxf(w, 0.001)
			out[row[i]] = ui.Rect{X: r.X, Y: y, W: w, H: h}
			y += h
		}
		return ui.Rect{X: r.X + w, Y: r.Y, W: r.W - w, H: r.H}
	}
	h := sumf(areas) / r.W
	x := r.X
	for i, area := range areas {
		w := area / maxf(h, 0.001)
		out[row[i]] = ui.Rect{X: x, Y: r.Y, W: w, H: h}
		x += w
	}
	return ui.Rect{X: r.X, Y: r.Y + h, W: r.W, H: r.H - h}
}

// worstShape is the least square tile a row of these values would make in r —
// the number the row is grown while it gets smaller.
func worstShape(row []int, values []float64, r ui.Rect, scale float32) float32 {
	areas := rowAreas(row, values, scale)
	if len(areas) == 0 || r.W <= 0 || r.H <= 0 {
		return math.MaxFloat32
	}
	stacked, beside := rowShapes(areas, r)
	return minf(stacked, beside)
}

// rowShapes is the worst tile shape a row of areas could be laid into r at,
// both ways round: stacked down a strip on the left, and side by side along a
// strip on the top. The better of the two is what a row is grown and placed
// by, so the measuring and the placing cannot disagree.
func rowShapes(areas []float32, r ui.Rect) (stacked, beside float32) {
	if r.W <= 0 || r.H <= 0 {
		return math.MaxFloat32, math.MaxFloat32
	}
	total := sumf(areas)
	w, h := total/r.H, total/r.W
	stacked, beside = 0, 0
	for _, area := range areas {
		stacked = maxf(stacked, aspect(w, area/maxf(w, 0.001)))
		beside = maxf(beside, aspect(area/maxf(h, 0.001), h))
	}
	return stacked, beside
}

// rowAreas is what room each of a row's values takes, in DIPs²: its share of
// the area scaled from the values onto the rect. A value of nothing takes no
// room and is given a sliver rather than a division by zero.
func rowAreas(row []int, values []float64, scale float32) []float32 {
	areas := make([]float32, len(row))
	for i, at := range row {
		if v := values[at]; v > 0 && !math.IsNaN(v) {
			areas[i] = float32(v) * scale
		}
	}
	return areas
}

func sumf(vs []float32) float32 {
	total := float32(0)
	for _, v := range vs {
		total += v
	}
	return total
}

// aspect is how far a rectangle of w by h is from being a square: 1 is square,
// 4 is a line four times as long as it is thick.
func aspect(w, h float32) float32 {
	if w <= 0 || h <= 0 {
		return math.MaxFloat32
	}
	return maxf(w/h, h/w)
}

// ── flows ──────────────────────────────────────────────────────────────────

// Link is one edge of a flow: how much of one node's value arrives at
// another. Nodes are named by their place in the caller's own list, so a
// caller can order them the way the story goes and have the layout keep that
// order.
type Link struct {
	// Source and Target are the two nodes' indices.
	Source, Target int
	// Value is how much flows along it, which is what its ribbon's width is.
	Value float64
}

// Depths is how many links deep each node is: the first node of a flow at 0,
// the one it arrives at at 1, and so on. It is what lays a Sankey out in
// columns, an alluvial in stages and a decomposition tree in rows.
//
// It is the longest path from a source rather than the shortest, so that a
// node waits for everything that feeds it rather than sitting in a column
// before its own inflow has arrived. A cycle has no such order, and a bounded
// number of passes gives it a stable one rather than looping for ever; a flow
// diagram whose links describe a circle is a caller's mistake, and this at
// least draws it.
func Depths(count int, links []Link) []int {
	depth := make([]int, max(count, 0))
	for _, l := range links {
		if l.Source < 0 || l.Source >= len(depth) || l.Target < 0 || l.Target >= len(depth) {
			panic("chart: a flow link points at a node that is not there")
		}
	}
	for range max(count, 0) {
		moved := false
		for _, l := range links {
			if want := depth[l.Source] + 1; depth[l.Target] < want {
				depth[l.Target] = want
				moved = true
			}
		}
		if !moved {
			break
		}
	}
	return depth
}

// ── calendars ──────────────────────────────────────────────────────────────

// Day is one cell of a calendar heatmap: a date and what it was worth.
type Day struct {
	// Date is the day, as "2006-01-02". The chart lays the days out on the
	// calendar they fall on, so the caller keeps the dates rather than the
	// chart guessing what a column of numbers meant.
	Date string
	// Value is what the cell is coloured by.
	Value float64
}

// CalendarGrid is where each date of a calendar heatmap falls: the week it is
// in, counting from the week the first date is in, and which weekday it is.
//
// The weeks start on the weekday the caller asks for, because a calendar
// heatmap that starts on Monday in one place and Sunday in the next is a
// chart whose rows mean different things in two panels.
type CalendarGrid struct {
	// Weekday is the day the first column starts on.
	Weekday time.Weekday
	// Weeks is how many week columns the grid has, from the earliest date
	// drawn to the latest.
	Weeks int
	// From and To are the first and last date the grid covers.
	From, To time.Time

	start time.Time
}

// NewCalendarGrid is the grid days fall on, laid out from the earliest of
// them, and whether any of them was a date at all. A day whose Date is not
// one is left out of the grid rather than guessed at.
func NewCalendarGrid(days []Day, weekday time.Weekday) (CalendarGrid, bool) {
	var from, to time.Time
	for _, d := range days {
		at, err := time.Parse("2006-01-02", d.Date)
		if err != nil {
			continue
		}
		if from.IsZero() || at.Before(from) {
			from = at
		}
		if to.IsZero() || at.After(to) {
			to = at
		}
	}
	if from.IsZero() {
		return CalendarGrid{Weekday: weekday}, false
	}
	// The first column is the week the earliest day is in, back to the day
	// that week starts on, so that every column has all seven rows.
	start := from.AddDate(0, 0, -weekdayOffset(from.Weekday(), weekday))
	return CalendarGrid{
		Weekday: weekday,
		Weeks:   int(to.Sub(start).Hours()/24/7) + 1,
		From:    start,
		To:      to,
		start:   start,
	}, true
}

// Cell is where a date falls on the grid: its week column and its weekday
// row. A date that is not one, or one outside the grid, is not a cell — which
// is what a heatmap means by a day nobody gave it.
func (g CalendarGrid) Cell(date string) (week, day int, ok bool) {
	at, err := time.Parse("2006-01-02", date)
	if err != nil || g.start.IsZero() {
		return 0, 0, false
	}
	day = weekdayOffset(at.Weekday(), g.Weekday)
	week = int(at.Sub(g.start).Hours() / 24 / 7)
	if week < 0 || week >= g.Weeks {
		return 0, 0, false
	}
	return week, day, true
}

// weekdayOffset is how far a weekday is from the one a week starts on,
// counted forwards round the week.
func weekdayOffset(day, start time.Weekday) int {
	return (int(day) - int(start) + 7) % 7
}
