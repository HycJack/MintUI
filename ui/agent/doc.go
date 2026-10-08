// Package agent draws a run: where it is, what it has done, what it is
// asking for, and what it changed.
//
// It exists because an agent run is not a conversation and not a job. It has
// a hierarchy of steps rather than a sequence of messages, a set of things it
// touched rather than a diff, and moments where it stops and waits for a
// person — and those three shapes are what this package's components are
// named for.
//
// Four rules hold across every component here, and they are the same four the
// rest of the library keeps:
//
//   - Nothing is held. A step's detail, a file's approval, an agent's children
//     and a dialog's answer are all the caller's: passed in as a pointer, and
//     written back the frame after the user acted. A component that owned its
//     own expansion would be a component whose state the caller cannot restore
//     after the view is rebuilt from scratch.
//   - Options in, element or Result out. Every component takes an
//     XxxOptions and returns either the element it drew or an XxxResult
//     carrying it and what was pressed in it.
//   - A thing this package cannot give is a panic, not a silent zero. An
//     unnamed agent, a step list with no steps and no empty state, a dialog
//     with nothing to decide: each of those is a caller who has decided to
//     show nothing while asking to show something.
//   - An element with only a glyph in it carries a name. A status mark with
//     no word beside it is a decoration.
//
// The two things that are arithmetic rather than drawing — which severity a
// status is, and what a three-way merge decided — are plain functions with no
// window in them, so they can be checked line by line without rendering
// anything. That is deliberate: those are the two places in an agent's
// interface where being wrong is not a look but a mistake.
package agent
