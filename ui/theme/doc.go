// Package theme is what the interface looks like before anyone draws it: the
// two palettes, the density scale, and the measurements everything else is
// sized from.
//
// A [Tokens] is a flat set of colours with no mode of its own — it is dark if
// its surfaces are dark, and it can mix a dark background with a custom fill
// without a flag to keep in step. [Light] and [Dark] are the library's.
//
// Sizes live here too, and they are all one family. [Unit] is the spacing step
// every gap is a multiple of; the font sizes ascend from [CaptionSize] to
// [TitleSize]; [ColumnWidth] is fixed on purpose, because a Kanban lane that
// narrows with the window wraps its card titles into two lines and the board
// stops being scannable.
//
// Nothing here draws or reads the window. Every contrast in the palettes is
// checked by the package's tests, so a colour that drifts out of range fails
// the build rather than the review.
package theme
