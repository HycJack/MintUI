// Package project holds the components a piece of work is planned with: the
// board its cards sit on, the card and the detail panel behind it, the
// pickers that change one of its fields, and the two views that say how the
// work is going — the burndown and the clock.
//
// # The board's shape is not this package's to change
//
// A column is theme.ColumnWidth wide and does not narrow. Below that width a
// card's title wraps, and a column of wrapped titles is a wall of text that
// cannot be scanned across; so the board scrolls sideways instead, through
// layout.Board, and this package's KanbanBoard and SprintBoard are built on
// that rather than on a row of their own. See docs/design-system.md §6.7.
//
// # Dragging is a string
//
// A task's identity is its id, a string, and a drag carries that id between
// columns. ui.Drop[string] is typed on the receiving end, so a lane says what
// it takes and a card that arrives with the wrong type does not compile rather
// than arriving as nil. The move is reported to the caller rather than
// performed here: this package has no store, and a component that reordered a
// caller's slice would leave them believing a card had moved when nothing had
// been saved.
//
// # The chart is a chart
//
// BurndownChart is ui/chart's LineChart with two series over it. It does not
// draw its own axes or its own plot area, because the frame in ui/chart is the
// thing that measures the gutters and thins the labels so they do not
// collide; a second implementation of that is a second answer to a question
// that has to have one.
//
// # Conventions
//
// As everywhere in this library: core.Use is called once by the harness, the
// parameters are in XxxOptions, the interactions come back on an XxxResult as
// a single method, nothing here keeps a value the caller did not hand it, and
// anything this package cannot draw is a panic rather than a guess.
package project
