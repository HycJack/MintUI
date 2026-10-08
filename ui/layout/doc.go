// Package layout is the frame a window is built in: the page header and the
// board's lanes.
//
// [PageHeader] is the top of a page — breadcrumbs, a title, the figures that
// qualify it and a row of controls the page itself owns. [Board] lays lanes
// side by side and scrolls them horizontally, and [Column] is one lane.
//
// Two things about a lane are worth knowing before building one. Its width is
// fixed at [theme.ColumnWidth]: a lane that narrows with the window wraps its
// card titles and the board stops being scannable, so the board scrolls
// instead. And an empty lane draws [ColumnOptions.Empty] outside the scroll
// view, because a button inside a scroll view is hard to press — see the note
// on nested scrolls on [Column].
package layout
