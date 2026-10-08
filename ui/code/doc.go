// Package code shows code: a file with line numbers and highlighting, the
// difference between two of them, the bytes of one, and the panels a
// development window wears around all of that.
//
// Everything in this package that shows lines goes through one set of
// primitives in shared.go — the row, the gutter, the current line's band, the
// selection. That is not tidiness. A viewer, a diff, a hex view and a log each
// want their own idea of where a line number goes and what the current line
// looks like, and four answers to that is four views that disagree by one
// pixel on the left and one colour on the row the caret is on.
//
// Syntax highlighting here is a token scanner, not a parser. It says which
// parts of a line are keywords, strings, numbers and comments, and stops. It
// does not build a tree, it does not know what a function is, and it will
// highlight a half-written line without complaining. That is the trade worth
// making for a component that has to highlight a file of three thousand lines
// in a frame: a real parser is a dependency and a cost, and the thing a
// reader actually uses it for — telling a string from an identifier — a
// scanner gets right.
package code
