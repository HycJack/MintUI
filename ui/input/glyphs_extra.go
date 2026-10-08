package input

import "github.com/egoist/mygo/ui"

// The marks the controls in this package's later files need. They are drawn
// on the same 24-unit grid, with the same stroke weight and the same
// `currentColor`, as the ones field.go holds, so that a mark beside a field
// and a mark beside a zone are one set of marks.

var (
	glyphFolder = mustGlyph(`<path d="M3.5 7.5A2 2 0 0 1 5.5 5.5h3.2a2 2 0 0 1 1.5.7l1 1.2h6.3a2 2 0 0 1 2 2v7.1a2 2 0 0 1-2 2h-13a2 2 0 0 1-2-2Z"/>`)
	glyphFile   = mustGlyph(`<path d="M13.5 3.5H7a2 2 0 0 0-2 2v13a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V9Z"/><path d="M13.5 3.5V9H19"/>`)
	glyphDrop   = mustGlyph(`<path d="M12 4v10"/><path d="M8 10.5 12 14.5l4-4"/><path d="M4.5 17v1.5a2 2 0 0 0 2 2h11a2 2 0 0 0 2-2V17"/>`)
	glyphUndo   = mustGlyph(`<path d="M4.5 9.5h9a5.5 5.5 0 1 1 0 11H9"/><path d="M8 5.5 4.5 9.5 8 13.5"/>`)
	glyphGrip   = mustGlyph(`<circle cx="9" cy="6" r="1.1" fill="currentColor" stroke="none"/><circle cx="15" cy="6" r="1.1" fill="currentColor" stroke="none"/><circle cx="9" cy="12" r="1.1" fill="currentColor" stroke="none"/><circle cx="15" cy="12" r="1.1" fill="currentColor" stroke="none"/><circle cx="9" cy="18" r="1.1" fill="currentColor" stroke="none"/><circle cx="15" cy="18" r="1.1" fill="currentColor" stroke="none"/>`)
	glyphTrash  = mustGlyph(`<path d="M4.5 7h15"/><path d="M9.5 7V5.5a1.5 1.5 0 0 1 1.5-1.5h2a1.5 1.5 0 0 1 1.5 1.5V7"/><path d="M6.5 7l.8 12a1.5 1.5 0 0 0 1.5 1.4h6.4a1.5 1.5 0 0 0 1.5-1.4l.8-12"/><path d="M10.5 11v6M13.5 11v6"/>`)
)

// glyphOf is the mark a directory row wears rather than a file's.
func glyphOf(dir bool) *ui.SVG {
	if dir {
		return glyphFolder
	}
	return glyphFile
}
