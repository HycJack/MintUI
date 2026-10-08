package input

import (
	"strconv"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
)

// QRCode draws a scannable QR code for a piece of text.
//
// It is a real encoder, not a picture of one. Byte mode is laid out the way
// the specification says — a mode indicator, a character count in the width
// that mode allows, the bytes, a terminator, the two pad bytes — corrected
// with Reed-Solomon over the block structure the error-correction level
// implies, placed on the matrix by the zig-zag walk around the function
// patterns, and masked.
//
// Every one of those steps is here because each is a step a scanner relies on,
// and each fails in the same unhelpful way: skip the masking and the pixels
// are wrong; skip the error correction and a code read over a screen may not
// read at all; get the zig-zag order wrong and the bytes land in the wrong
// places, which is the one failure that still produces a plausible-looking
// square. A QR code is the only component in this library whose output is
// read by a machine nobody in the conversation can see, so the arithmetic is
// the whole of it and nothing here is a drawing of something else.
//
// It is one Draw call over a fixed grid, so a code costs what a picture of the
// same size costs, which is the reason it is worth having rather than shipping
// an image.

// QRErrorCorrection is how much of a code is spent on being readable when it
// is partly damaged, at the cost of how much text fits in it.
type QRErrorCorrection int

const (
	// QRLow spends almost nothing on correction and fits the most text. It is
	// for a code shown on a screen the person can look at.
	QRLow QRErrorCorrection = iota
	// QRMedium is the usual answer: about a tenth of a code spent on
	// correction, which survives a screen with glare on it.
	QRMedium
	// QRHigh survives being printed on a label and read at an angle.
	QRHigh
	// QRHighest survives almost anything, and fits a third of what QRLow
	// does. It is for a code that will be on a sticker.
	QRHighest
)

func (e QRErrorCorrection) String() string {
	switch e {
	case QRMedium:
		return "Medium"
	case QRHigh:
		return "High"
	case QRHighest:
		return "Highest"
	}
	return "Low"
}

// qrECCBits is the two bits each level is called by on the wire, in the order
// the levels are declared here: Low, Medium, High, Highest. The order is not
// the order of the values, which is the first thing that goes wrong when this
// is written from memory.
var qrECCBits = [...]uint32{1, 0, 3, 2}

// qrMaskTotal is how many of the eight data masks have been tried, which is
// eight, and it is a constant rather than a literal so that the loop over them
// reads as "all of them" and not as a number somebody chose.
const qrMaskTotal = 8

// QRCodeOptions configure a QRCode.
type QRCodeOptions struct {
	// Label names the code for assistive technology, and is required: a
	// square of black modules is the one thing in this library that says
	// nothing at all about what it is, and a code a reader cannot name is a
	// code a reader cannot be told to look for.
	Label string
	// Size is the code's own width and height in DIPs. Zero is a module at
	// the density's step, which is a code big enough to read on a phone
	// held at a normal distance and small enough to sit in a card.
	Size float32
	// Module is how wide one module is drawn. Zero takes Size divided by the
	// number of modules. It exists because a code is only readable when a
	// module is two or three device pixels across, and a caller with a
	// printer knows the device's pixels and this library does not.
	Module float32
	// Quiet is the quiet zone around the code, in modules. Anything below
	// four is raised to four: the specification asks for four, a scanner's
	// own margin is four, and the finder patterns are the first thing a
	// scanner looks for, so anything touching them is what stops it finding
	// them. A code drawn hard against the edge of a card is a code that will
	// not scan, and it looks perfectly fine.
	Quiet int
	// Foreground and Background draw the modules. Zero alpha takes the
	// window's text colour on its background, which is the one pair that
	// contrasts in both appearances without the caller having to know which
	// appearance the window is in.
	Foreground, Background ui.Color
	// ErrorCorrection says how much of the code is spent on being readable
	// when it is partly damaged.
	ErrorCorrection QRErrorCorrection
}

// QRCodeResult carries a QRCode and what was encoded into it.
type QRCodeResult struct {
	// Element is the code.
	Element *ui.Element
	// version is the symbol version that was used, 1 to 40.
	version int
	// modules is how many modules across the code is, the quiet zone
	// excluded.
	modules int
	// capacity is how many bytes the version and the level could have held,
	// which is what a caller uses to say "that is too long" before a user
	// finds it out from a scanner.
	capacity int
}

// Version is the symbol version that was used, 1 to 40. It is on the result
// rather than asked for, because the version is a consequence of how much text
// there was, and a caller checking a limit needs to know it was met.
func (r QRCodeResult) Version() int { return r.version }

// Modules is how many modules across the code is, the quiet zone excluded.
func (r QRCodeResult) Modules() int { return r.modules }

// Capacity is how many bytes the version and the level could have held. It is
// what to compare a length against before a code is drawn rather than after,
// because a code that is too long has no version at all.
func (r QRCodeResult) Capacity() int { return r.capacity }

// QRCode draws a code for text, which a scanner will read.
//
// Text is encoded whole in byte mode: a URL, a wifi password, a telephone
// number and a sentence in any language all go in the same way and all come
// back out as the same string, which is the property that makes this usable
// for something a person has to type by hand.
//
// There is no ECI header, and that is a decision rather than an omission: a
// code with an ECI header is read as a different character set by a scanner
// that was not expecting one, so putting one in by default would make a plain
// URL harder to read rather than easier. UTF-8 is what byte mode produces
// anyway, and every scanner here assumes it.
func QRCode(c *ui.Context, text string, opts QRCodeOptions) QRCodeResult {
	if opts.Label == "" {
		panic("input: QRCode needs a Label; a square of black modules is the one thing in " +
			"this library that says nothing at all about what it is")
	}
	if opts.ErrorCorrection < QRLow || opts.ErrorCorrection > QRHighest {
		panic("input: QRCode ErrorCorrection is not one of QRLow, QRMedium, QRHigh or QRHighest")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	matrix, version, err := qrEncode([]byte(text), opts.ErrorCorrection)
	if err != nil {
		panic("input: QRCode cannot encode " + strconv.Itoa(len(text)) + " bytes at " +
			opts.ErrorCorrection.String() + " correction: " + err.Error())
	}

	quiet := max(opts.Quiet, 4)
	span := len(matrix) + quiet*2

	module := opts.Module
	if module <= 0 {
		module = u
		if opts.Size > 0 {
			module = opts.Size / float32(span)
		}
	}
	side := float32(span) * module
	fg, bg := opts.Foreground, opts.Background
	if fg.A == 0 {
		fg = k.Text
	}
	if bg.A == 0 {
		bg = k.Background
	}

	// A code is dark on light or light on dark and nothing in between: a
	// tinted code still scans and a low-contrast one does not, and the
	// scanner has no way to tell the person that from a dirty screen.
	//
	// The paint is set inside the box's own Children call because an element
	// has to exist before anything can be told to paint onto it — an element
	// made out here and put in afterwards would be the caller's child rather
	// than the code's.
	code := ui.Box(c).Size(side, side).Shrink(0).Role(ui.RoleImage).
		Label(opts.Label).Tooltip(opts.Label)
	code.Children(func() {
		code.Draw(func(p *ui.Painter, r ui.Rect) {
			p.Fill(r, bg, 0)
			cell := r.W / float32(span)
			// One fill per dark module rather than a bitmap: the grid is
			// fixed, so there is nothing to resample, and a bitmap would
			// have to be built at the size of the box to reach the same
			// place.
			for y := range matrix {
				for x, dark := range matrix[y] {
					if !dark {
						continue
					}
					p.Fill(ui.Rect{
						X: r.X + float32(x+quiet)*cell,
						Y: r.Y + float32(y+quiet)*cell,
						W: cell, H: cell,
					}, fg, 0)
				}
			}
		})
	})

	return QRCodeResult{
		Element:  code,
		version:  version,
		modules:  len(matrix),
		capacity: qrCapacity(version, opts.ErrorCorrection),
	}
}

// ── the encoder ────────────────────────────────────────────────────────────

// qrEncode is the whole of the symbol, in the order the steps have to happen:
// the data code words, the error correction, the function patterns, the code
// words placed around them, the mask and the format information.
//
// It is one function with named steps rather than seven functions because the
// steps share one matrix and one "is this cell spoken for" answer, and
// splitting them means passing that state as four arguments to seven
// functions that are never called apart from each other.
func qrEncode(data []byte, ecc QRErrorCorrection) ([][]bool, int, error) {
	version, ok := qrVersionFor(len(data), ecc)
	if !ok {
		return nil, 0, qrErrTooLong
	}
	size := version*4 + 17

	// The data code words, split into blocks, corrected and interleaved.
	stream := qrCodewords(data, version, ecc)

	// The matrix and the set of cells the patterns own. The format area is
	// marked as spoken for before the code words go down, so that a code word
	// cannot land under a piece of format information.
	m, spoken := qrFunctionPatterns(version, size)

	qrPlaceCodewords(m, spoken, size, stream)

	mask := qrBestMask(m, spoken, size)
	qrApplyMask(m, spoken, size, mask)
	qrWriteFormat(m, size, ecc, mask)
	return m, version, nil
}

// qrErrTooLong is the one error the encoder raises, as a value rather than a
// string so that a caller matching on it does not have to match on prose. The
// component wraps it into a panic, which is how every other failure in this
// package is reported.
var qrErrTooLong = &qrError{"the text is longer than any QR version can carry"}

// qrError is the error type this package's encoder raises.
type qrError struct{ msg string }

func (e *qrError) Error() string { return e.msg }

// qrCodewords is the data bit stream turned into the interleaved code words
// that go on the page: the bytes themselves, cut into blocks, each block given
// its own error correction, and the two sets interleaved.
//
// The blocks are not all the same length, and that is the specification's
// rule rather than a nicety: the data code words are divided by the number of
// blocks, and the remainder goes one byte each to the blocks at the end. A
// code that padded them out instead would fit the same text into the same
// space and be read by a different scanner.
func qrCodewords(data []byte, version int, ecc QRErrorCorrection) []byte {
	bits := qrDataBits(data, version)
	words := qrBytesToWords(bits)
	words = append(words, qrPadWords(len(words), qrDataWords(version, ecc))...)

	numBlocks := qrNumBlocks(version, ecc)
	eccLen := qrECCCodeWords(version, ecc)
	shortLen := len(words) / numBlocks
	// How many blocks are short: all of them when the data divides evenly,
	// and the rest when it does not.
	numShort := numBlocks - len(words)%numBlocks

	blocks := make([][]byte, numBlocks)
	corrs := make([][]byte, numBlocks)
	at := 0
	for i := range numBlocks {
		n := shortLen
		if i >= numShort {
			n++
		}
		// The offset runs rather than being counted from the block's place.
		// The long blocks are one byte longer each, so every block after the
		// first long one starts later than its index suggests — and a code
		// whose blocks are taken from the wrong offsets has the right size
		// and the wrong contents, which is a code that scans to something
		// else entirely rather than one that fails.
		blocks[i] = words[at : at+n]
		at += n
		corrs[i] = qrAddECC(blocks[i], eccLen)
	}

	out := make([]byte, 0, len(words)+numBlocks*eccLen)
	// The data code words first, one from each block in turn. A short block
	// simply stops contributing, which is what lets the long ones' extra byte
	// come last without a hole where the short ones ended.
	for i := range shortLen + 1 {
		for j := range numBlocks {
			if i < len(blocks[j]) {
				out = append(out, blocks[j][i])
			}
		}
	}
	// Then the correction words, interleaved the same way. This is the step
	// that makes a code survive being printed: a scratch across one corner
	// damages a few bytes of every block rather than all of one, and a block
	// a third damaged can still be corrected.
	for i := range eccLen {
		for j := range numBlocks {
			out = append(out, corrs[j][i])
		}
	}
	return out
}

// qrDataBits is the bit stream of byte mode: the mode indicator, the character
// count, the bytes themselves, the terminator, the padding to a whole byte,
// and then the two alternating pad bytes to the end of the data area.
//
// The mode indicator is 0100 and the count is eight bits below version ten and
// sixteen from ten up, which is what lets a symbol hold more than 255
// characters at all. A version is only chosen when the whole of this fits, so
// the room is never negative; asserting it would be a second copy of
// qrVersionFor's answer to the same question.
func qrDataBits(data []byte, version int) []byte {
	countBits := 8
	if version >= 10 {
		countBits = 16
	}
	room := qrDataWords(version, qrDefaultECC) * 8

	out := make([]byte, 0, room)
	// Four bits, not the byte 0x04: the mode indicator is 0100 and a whole
	// byte of it would shift every bit of the count along by three, which
	// produces a code of exactly the right size that decodes to nonsense.
	out = append(out, qrBits(qrModeByte, 4)...)
	out = append(out, qrBits(uint32(len(data)), countBits)...)
	for _, b := range data {
		out = append(out, qrBits(uint32(b), 8)...)
	}
	// The terminator is four zero bits, or as many as fit where there is not
	// room for four. A code with no room for a terminator is still valid; one
	// padded with zeros to make room for one would be a different number of
	// bits and so a different code.
	if term := min(4, room-len(out)); term > 0 {
		out = append(out, make([]byte, term)...)
	}
	// Then up to a whole byte. A code word is eight bits and a partial one
	// cannot be a code word, so the last bits of a full code are padded out
	// with light modules — which is what the remainder bits on the page are.
	for len(out)%8 != 0 {
		out = append(out, 0)
	}
	return out
}

// qrPadWords are the two alternating pad code words that fill the data area
// once the text has run out. They are not arbitrary: a scanner that read them
// as data would show them, so the specification fixes them at 11101100 and
// 00010001, and the alternation continues from wherever the previous pad left
// off so that a code with a whole number of pairs still alternates.
func qrPadWords(have, want int) []byte {
	out := make([]byte, 0, want-have)
	pad := byte(0xec)
	for len(out) < want-have {
		out = append(out, pad)
		pad ^= 0xec ^ 0x11
	}
	return out
}

// qrBytesToWords is a bit stream packed into code words, most significant bit
// first — the order they are written to the page in, so packing and placing
// are the same order and there is no reversal to get wrong between them.
func qrBytesToWords(bits []byte) []byte {
	out := make([]byte, (len(bits)+7)/8)
	for i, b := range bits {
		if b != 0 {
			out[i/8] |= 1 << uint(7-i%8)
		}
	}
	return out
}

// qrAddECC is the Reed-Solomon correction for one block: its data code words
// through the generator polynomial, and the remainder is its correction.
//
// The arithmetic is in GF(256) under the primitive polynomial 0x11d. That is
// what makes the scheme work at all: an eight-bit field where adding and
// multiplying wrap by a different polynomial than a byte would, so a burst of
// errors across a few modules becomes a burst of errors across a few code
// words — a shape the correction can find, where in plain bytes it would be
// noise.
func qrAddECC(block []byte, eccLen int) []byte {
	gen := qrGenerator(eccLen)
	out := make([]byte, eccLen)
	for _, b := range block {
		factor := b ^ out[0]
		copy(out, out[1:])
		out[len(out)-1] = 0
		if factor != 0 {
			for i, g := range gen {
				out[i] ^= gfMul(g, factor)
			}
		}
	}
	return out
}

// qrGenerator is the divisor the correction is computed against: the product
// of (x - r^i) for i in 0..n-1 over GF(256), with its leading coefficient
// dropped.
//
// The drop is not a simplification, it is the shape the synthetic division
// below needs. The polynomial has degree n, so it has n+1 coefficients, and
// the division always assumes the leading one is 1 and keeps the other n. So
// the table here is the coefficient of x^(n-1) first, down to the constant.
//
// It is the same polynomial for every block of the same length, so it is
// built once per length rather than once per block — which is most of the
// work in a large code.
var generatorCache = map[int][]byte{}

// qrGenerator builds the divisor by multiplying in one (x - r^i) at a time,
// each one a shift up a degree plus a scale by r^i, which is all a polynomial
// multiplication is.
func qrGenerator(n int) []byte {
	if g, ok := generatorCache[n]; ok {
		return g
	}
	g := make([]byte, n)
	// The monomial x^0, which is the product of nothing.
	g[n-1] = 1
	root := byte(1)
	for range n {
		for j := range n {
			g[j] = gfMul(g[j], root)
			if j+1 < n {
				g[j] ^= g[j+1]
			}
		}
		// r = 2 is the field's generator, which is what the specification
		// fixes; any other primitive root would give a different and equally
		// valid code that no scanner here reads.
		root = gfMul(root, 2)
	}
	generatorCache[n] = g
	return g
}

// gfExp and gfLog are the exponent and logarithm tables of GF(256) under
// 0x11d: gfLog[x] is the n with r^n = x, and gfExp[n] is r^n. Without them
// every one of the eight multiplications in every code word would be a
// shift-and-reduce loop.
//
// gfExp is doubled rather than wrapped with a modulo, so that the product of
// two logarithms — whose sum runs to 508 — can index it directly. That is the
// whole reason the table is 512 long and it is not an accident of padding.
var (
	gfExp [512]byte
	gfLog [256]byte
)

func init() {
	x := 1
	for i := range 255 {
		gfExp[i] = byte(x)
		gfLog[x] = byte(i)
		x <<= 1
		if x&0x100 != 0 {
			x ^= 0x11d
		}
	}
	for i := 255; i < 512; i++ {
		gfExp[i] = gfExp[i-255]
	}
}

// gfMul is a product of two field elements. A product with a zero is short-
// circuited because the log of zero is undefined and the table is left zero
// there: indexing it would return the field's zero by accident rather than by
// the check, which is a bug waiting for a code that hits it.
func gfMul(a, b byte) byte {
	if a == 0 || b == 0 {
		return 0
	}
	return gfExp[int(gfLog[a])+int(gfLog[b])]
}

// qrPlaceCodewords puts the code words on the matrix in the zig-zag the
// specification lays them out in: two modules wide, right to left, upward
// first, and round the timing column rather than through it.
//
// The direction flips on the parity of the column rather than on a toggle,
// because that is where the pattern starts: the top-right finder is at the
// top, so the first code words run up the right-hand pair of columns from the
// bottom. A toggle gets this right only until it wraps, and a code whose
// first bytes land one column out still looks like a code.
func qrPlaceCodewords(m, spoken [][]bool, size int, words []byte) {
	bitAt := func(i int) bool {
		if i/8 >= len(words) {
			// Past the code words: the remainder is light modules, which is
			// how a scanner finds where the data ends.
			return false
		}
		return words[i/8]&(1<<uint(7-i%8)) != 0
	}
	pos := 0
	for right := size - 1; right >= 1; right -= 2 {
		if right == 6 {
			right = 5 // the timing column belongs to the pattern, not the data
		}
		for vert := range size {
			for j := range 2 {
				x := right - j
				upward := (right+1)&2 == 0
				y := vert
				if upward {
					y = size - 1 - vert
				}
				if spoken[y][x] {
					continue
				}
				m[y][x] = bitAt(pos)
				pos++
			}
		}
	}
}

// qrFunctionPatterns is a blank matrix with everything that is not data drawn
// on it, and the set of cells that are spoken for.
//
// What is drawn: the timing patterns across the whole width, the three finder
// patterns with their separators over the top of them, the alignment patterns,
// and the reservation of both copies of the format information. The last is
// drawn as a reservation rather than as bits because the bits are not known
// until the mask has been chosen — but the cells have to be spoken for before
// the code words go down.
func qrFunctionPatterns(version, size int) (m, spoken [][]bool) {
	m = make([][]bool, size)
	spoken = make([][]bool, size)
	for i := range size {
		m[i] = make([]bool, size)
		spoken[i] = make([]bool, size)
	}
	// The timing patterns first and across the whole run, so that the finder
	// patterns can be drawn over the ends of them: they give the scanner the
	// module pitch, which is how it works out how far it is from the centre
	// and how big a module is.
	for i := range size {
		qrSet(m, spoken, i, 6, i%2 == 0)
		qrSet(m, spoken, 6, i, i%2 == 0)
	}
	qrFinder(m, spoken, 3, 3)
	qrFinder(m, spoken, size-4, 3)
	qrFinder(m, spoken, 3, size-4)

	// The alignment patterns, at every combination of their centres except
	// the three that would land on a finder pattern.
	//
	// The three are not a detail. A finder pattern is seven modules and an
	// alignment pattern is five, and a version 2 symbol's two centres are
	// six and eighteen: every combination of them but the one at (18,18) is
	// inside a finder. Drawing them anyway puts a second, different pattern
	// on top of the first, and the format information strip runs through
	// exactly that band — so a code with them drawn has its own error
	// correction level overwritten by a ring of dark and light modules.
	for _, at := range qrAlignPositions(version) {
		for _, other := range qrAlignPositions(version) {
			switch {
			case at == 6 && other == 6,
				at == 6 && other == size-7,
				at == size-7 && other == 6:
				continue
			}
			qrAlignment(m, spoken, at, other)
		}
	}
	// The format areas, reserved. Both copies: the one round the top-left
	// finder and the one split between the other two. Marking them now is
	// what stops a code word landing under them.
	//
	// The positions are the specification's, and they are written as row and
	// column rather than as column and row on purpose. A version of this that
	// reads them the other way round produces a square of exactly the right
	// size with the right finder patterns, which is the failure mode this
	// whole encoder is written against.
	for i := range 15 {
		switch {
		case i <= 5:
			// Along the bottom of the top-left finder, to the right.
			qrClaim(spoken, i, 8)
		case i == 6:
			// Skipping the timing module in column six, which is why there
			// is a gap here at all.
			qrClaim(spoken, 7, 8)
		case i == 7:
			qrClaim(spoken, 8, 8)
		case i == 8:
			// Round the corner of the separator.
			qrClaim(spoken, 8, 7)
		default:
			// Up the right-hand side of the top-left finder.
			qrClaim(spoken, 8, 14-i)
		}
		switch {
		case i <= 6:
			// Up the right-hand side of the bottom-left finder, from the
			// bottom row, stopping one module short of the dark one.
			qrClaim(spoken, 8, size-1-i)
		case i == 7:
			// The bit that would have gone in the dark module's place goes
			// to the far corner instead, below the separator of the top-right
			// finder. It is the one place in the symbol where the two copies
			// of the format information do not mirror one another, and it is
			// why the two are not two of the same shape.
			qrClaim(spoken, size-8, 8)
		default:
			// Along the bottom of the top-right finder, to the right, with
			// the last bit in the last column.
			qrClaim(spoken, size+i-15, 8)
		}
	}
	// The dark module, reserved and always set. It sits at four times the
	// version plus nine down, in column eight — the one cell in the symbol
	// that carries nothing at all, put there so that this corner can never
	// be all light.
	qrClaim(spoken, 8, size-8)
	return m, spoken
}

// qrFinder draws one finder pattern with its separator, centred at (x, y).
//
// It is a nine-module square of which the outermost ring and the ring two in
// are light: that leaves the seven-module ring with a three-module ring inside
// it and a single dark module at the centre, which is the shape a scanner
// looks for and the only thing on the page that says "this is a QR code".
func qrFinder(m, spoken [][]bool, x, y int) {
	for dy := -4; dy <= 4; dy++ {
		for dx := -4; dx <= 4; dx++ {
			dist := max(absInt(dx), absInt(dy))
			qrSet(m, spoken, x+dx, y+dy, dist != 2 && dist != 4)
		}
	}
}

// qrAlignment draws one alignment pattern centred at (x, y): a five-module
// square, dark ring, light ring, dark centre — the same shape as a finder
// without the separator. They are what lets a code with an unusual number of
// modules be read, and version one has none at all.
func qrAlignment(m, spoken [][]bool, cx, cy int) {
	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			dist := max(absInt(dx), absInt(dy))
			qrSet(m, spoken, cx+dx, cy+dy, dist != 1)
		}
	}
}

func qrSet(m, spoken [][]bool, x, y int, dark bool) {
	if x < 0 || y < 0 || y >= len(m) || x >= len(m) {
		return
	}
	m[y][x] = dark
	spoken[y][x] = true
}

func qrClaim(spoken [][]bool, x, y int) {
	if x < 0 || y < 0 || y >= len(spoken) || x >= len(spoken) {
		return
	}
	spoken[y][x] = true
}

// qrAlignPositions are the centres of a version's alignment patterns, along
// the row and the column alike, in the order the specification numbers them.
//
// They are worked out rather than read out of a table of forty rows, because
// the rule behind them is short enough to hold in the head: the centres are
// six from the edge and then evenly spaced back towards the other edge, one
// pattern for every seven versions plus two.
func qrAlignPositions(version int) []int {
	if version < 2 {
		// Version one has no alignment patterns at all. Its code words are
		// short enough that the finders and the timing patterns are the only
		// help a scanner needs.
		return nil
	}
	n := version/7 + 2
	size := version*4 + 17
	step := (version*4 + n*2 + 1) / (n*2 - 2) * 2
	if version == 32 {
		// The one version the rule does not give a whole number for, and the
		// one version where getting it wrong puts a pattern off by two
		// modules — which is enough for a scanner to read the code and
		// decode the wrong thing.
		step = 26
	}
	out := make([]int, 0, n)
	for i, pos := 0, size-7; i < n-1; i, pos = i+1, pos-step {
		out = append(out, pos)
	}
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return append(out, 6)
}

// qrApplyMask flips the data modules wherever the mask says, which is what
// stops a code having wide flat runs of one colour that a camera would find
// hard to find the edges of.
func qrApplyMask(m, spoken [][]bool, size, mask int) {
	for y := range size {
		for x := range size {
			if spoken[y][x] {
				continue
			}
			if qrMaskAt(mask, y, x) {
				m[y][x] = !m[y][x]
			}
		}
	}
}

// qrMaskAt is whether data mask n applies at (y, x): the eight patterns,
// numbered as the specification numbers them.
func qrMaskAt(mask, y, x int) bool {
	switch mask {
	case 0:
		return (x+y)%2 == 0
	case 1:
		return y%2 == 0
	case 2:
		return x%3 == 0
	case 3:
		return (x+y)%3 == 0
	case 4:
		return (y/2+x/3)%2 == 0
	case 5:
		return (x*y)%2+(x*y)%3 == 0
	case 6:
		return ((x*y)%2+(x*y)%3)%2 == 0
	case 7:
		return ((x*x+y*y)%2+(x*y)%3)%2 == 0
	}
	return false
}

// qrBestMask is the mask whose result has the fewest faults, which is the
// specification's figure of merit, and then the smallest number on a tie.
//
// It is the one step that cannot be skipped: an unmasked or wrongly masked
// code is still a square of the right size with the right finder patterns, so
// it looks right and scans badly. All eight are tried because the cost is
// eight passes over a matrix of a few thousand cells and the alternative is a
// code that fails on the desks it was not tested on.
func qrBestMask(m, spoken [][]bool, size int) int {
	best, bestScore := 0, -1
	for mask := range qrMaskTotal {
		candidate := make([][]bool, size)
		for y := range size {
			candidate[y] = make([]bool, size)
			copy(candidate[y], m[y])
		}
		qrApplyMask(candidate, spoken, size, mask)
		if score := qrPenalty(candidate, size); bestScore < 0 || score < bestScore {
			best, bestScore = mask, score
		}
	}
	return best
}

// qrPenalty is the specification's figure of merit: four rules, each costing
// points, added together. A lower score is a better mask.
func qrPenalty(m [][]bool, size int) int {
	score := 0
	// Rule one: a run of five or more like modules in a row or a column.
	for y := range size {
		run := 1
		for x := 1; x < size; x++ {
			if m[y][x] == m[y][x-1] {
				run++
			} else {
				run = 1
			}
			score += qrRunPenalty(run)
		}
	}
	for x := range size {
		run := 1
		for y := 1; y < size; y++ {
			if m[y][x] == m[y-1][x] {
				run++
			} else {
				run = 1
			}
			score += qrRunPenalty(run)
		}
	}
	// Rule two: a block of two by two in one colour, which is what a camera
	// reads as the inside of a finder pattern and so goes looking for more.
	for y := 0; y < size-1; y++ {
		for x := 0; x < size-1; x++ {
			v := m[y][x]
			if v == m[y][x+1] && v == m[y+1][x] && v == m[y+1][x+1] {
				score += 3
			}
		}
	}
	// Rule three: a finder-like run with four light modules beside it, in a
	// row or a column. This is the rule that catches a code whose data has
	// accidentally grown a finder pattern in the middle of itself.
	needle := []bool{true, false, true, true, true, false, true}
	clear := []bool{false, false, false, false}
	for y := range size {
		for x := 0; x+7 <= size; x++ {
			if !qrRowIs(m, y, x, needle) {
				continue
			}
			before := x >= 4 && qrRowIs(m, y, x-4, clear)
			after := x+11 <= size && qrRowIs(m, y, x+7, clear)
			if before || after {
				score += 40
			}
		}
	}
	for x := range size {
		for y := 0; y+7 <= size; y++ {
			if !qrColIs(m, x, y, needle) {
				continue
			}
			before := y >= 4 && qrColIs(m, x, y-4, clear)
			after := y+11 <= size && qrColIs(m, x, y+7, clear)
			if before || after {
				score += 40
			}
		}
	}
	// Rule four: how much of the code is dark, against how much is light. The
	// figure of merit wants it near half, in steps of five percent.
	dark := 0
	for y := range size {
		for x := range size {
			if m[y][x] {
				dark++
			}
		}
	}
	percent := dark * 100 / (size * size)
	score += (absInt(percent-50) / 5) * 10
	return score
}

// qrRunPenalty is what a run of like modules costs: three for reaching five,
// and one more for each module past it.
func qrRunPenalty(run int) int {
	switch {
	case run < 5:
		return 0
	case run == 5:
		return 3
	}
	return 1
}

func qrRowIs(m [][]bool, y, x int, want []bool) bool {
	if x < 0 || x+len(want) > len(m) {
		return false
	}
	for i, w := range want {
		if m[y][x+i] != w {
			return false
		}
	}
	return true
}

func qrColIs(m [][]bool, x, y int, want []bool) bool {
	if y < 0 || y+len(want) > len(m) {
		return false
	}
	for i, w := range want {
		if m[y+i][x] != w {
			return false
		}
	}
	return true
}

// qrWriteFormat puts the fifteen format bits in both of their places: the
// error-correction level, the mask, and the BCH check bits a scanner reads
// them back through, all masked so that an all-light format area is not read
// as the absence of one.
func qrWriteFormat(m [][]bool, size int, ecc QRErrorCorrection, mask int) {
	format := qrECCBits[ecc]<<3 | uint32(mask)
	rem := format
	for range 10 {
		rem = (rem << 1) ^ ((rem >> 9) * 0x537)
	}
	bits := (format<<10 | rem) ^ 0x5412

	// The first copy, round the top-left finder: six along the bottom of it,
	// then round the corner and up its right-hand side.
	for i := range 6 {
		m[8][i] = qrBitAt(bits, i)
	}
	m[8][7] = qrBitAt(bits, 6)
	m[8][8] = qrBitAt(bits, 7)
	m[7][8] = qrBitAt(bits, 8)
	for i := 9; i < 15; i++ {
		m[14-i][8] = qrBitAt(bits, i)
	}
	// The second copy, split between the other two finders: seven up the
	// right-hand side of the bottom-left one, then seven along the bottom of
	// the top-right one. It is there because one copy is enough for a clean
	// code and not enough for one with a corner torn off — and the seven and
	// the seven are not symmetric, because the bit that would have gone in
	// the dark module's place goes to the far corner instead.
	for i := range 7 {
		m[size-1-i][8] = qrBitAt(bits, i)
	}
	m[8][size-8] = qrBitAt(bits, 7)
	for i := 8; i < 15; i++ {
		m[8][size+i-15] = qrBitAt(bits, i)
	}
	// The dark module: always set, always here, carrying nothing.
	m[size-8][8] = true
}

// qrBitAt is bit i of a fifteen-bit value, counting from the top.
func qrBitAt(bits uint32, i int) bool { return bits&(1<<uint(14-i)) != 0 }

// qrBits is a value written as exactly n bits, most significant first.
func qrBits(v uint32, n int) []byte {
	out := make([]byte, n)
	for i := range n {
		out[i] = byte(v>>uint(n-1-i)) & 1
	}
	return out
}

// ── versions and capacity ──────────────────────────────────────────────────

// qrModeByte is the four-bit mode indicator for byte mode: 0100.
const qrModeByte = 0x4

// qrDefaultECC is the level a capacity question is answered at when the
// caller does not name one. Low is right, and it is right because it fits the
// most: a code that fits at Low fits at every other level, so a caller
// checking a length against it is never told something fits when it does not.
const qrDefaultECC = QRLow

// qrECCCodeWords is how many error-correction code words each block of a
// version carries at one level.
//
// These two tables — this one and qrNumBlocks — are the specification's, and
// they are the only numbers here that are written out rather than worked out.
// Everything else is arithmetic on the module count, and those two are not:
// the correction a level buys grows in jumps as a version grows, and the block
// count is then the smallest that keeps each block's own correction inside the
// limit. Getting either from a formula rather than from the table produces
// codes that are the right size and do not scan.
var qrECCPerBlockTable = [4][41]int{
	// 0  1  2  3  4  5  6  7  8  9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30 31 32 33 34 35 36 37 38 39 40
	{-1, 7, 10, 15, 20, 26, 18, 20, 24, 30, 18, 20, 24, 26, 30, 22, 24, 28, 30, 28, 28, 28, 28, 30, 30, 26, 28, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30},  // Low,  // QRLow (the specification's L)
	{-1, 10, 16, 26, 18, 24, 16, 18, 22, 22, 26, 30, 22, 22, 24, 24, 28, 28, 26, 26, 26, 26, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28}, // Medium,  // QRMedium (M)
	{-1, 13, 22, 18, 26, 18, 24, 18, 22, 20, 24, 28, 26, 24, 20, 30, 24, 28, 28, 26, 30, 28, 30, 30, 30, 30, 28, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30}, // QRHigh is the specification's Q,  // QRHigh (Q)
	{-1, 17, 28, 22, 16, 22, 28, 26, 26, 24, 28, 24, 28, 22, 24, 24, 30, 28, 28, 26, 28, 30, 24, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30}, // QRHighest is the specification's H,  // QRHighest (H)
}

// qrBlockTable is how many blocks a version's code is split into at one level,
// the second half of the same pair.
var qrBlockTable = [4][41]int{
	// 0  1  2  3  4  5  6  7  8  9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30 31 32 33 34 35 36 37 38 39 40
	{-1, 1, 1, 1, 1, 1, 2, 2, 2, 2, 4, 4, 4, 4, 4, 6, 6, 6, 6, 7, 8, 8, 9, 9, 10, 12, 12, 12, 13, 14, 15, 16, 17, 18, 19, 19, 20, 21, 22, 24, 25},              // Low,  // QRLow (the specification's L)
	{-1, 1, 1, 1, 2, 2, 4, 4, 4, 5, 5, 5, 8, 9, 9, 10, 10, 11, 13, 14, 16, 17, 17, 18, 20, 21, 23, 25, 26, 28, 29, 31, 33, 35, 37, 38, 40, 43, 45, 47, 49},     // Medium,  // QRMedium (M)
	{-1, 1, 1, 2, 2, 4, 4, 6, 6, 8, 8, 8, 10, 12, 16, 12, 17, 16, 18, 21, 20, 23, 23, 25, 27, 29, 34, 34, 35, 38, 40, 43, 45, 48, 51, 53, 56, 59, 62, 65, 68},  // QRHigh is the specification's Q,  // QRHigh (Q)
	{-1, 1, 1, 2, 4, 4, 4, 5, 6, 8, 8, 11, 11, 16, 16, 18, 16, 19, 21, 25, 25, 25, 34, 30, 32, 35, 37, 40, 42, 45, 48, 51, 54, 57, 60, 63, 66, 70, 74, 77, 81}, // QRHighest is the specification's H,  // QRHighest (H)
}

// qrECCCodeWords is how many correction code words one block carries at a
// version and level, and qrNumBlocks is how many blocks there are.
func qrECCCodeWords(version int, ecc QRErrorCorrection) int {
	return qrECCPerBlockTable[ecc][version]
}

func qrNumBlocks(version int, ecc QRErrorCorrection) int {
	return qrBlockTable[ecc][version]
}

// qrVersionFor is the smallest version that holds n bytes of text at that
// level.
func qrVersionFor(n int, ecc QRErrorCorrection) (int, bool) {
	for v := 1; v <= 40; v++ {
		if n <= qrCapacity(v, ecc) {
			return v, true
		}
	}
	return 0, false
}

// qrCapacity is how many bytes of text a version can carry at that level.
//
// The mode indicator and the character count are not text, and neither is a
// whole code word's worth of them: the four bits of the mode indicator share
// the first code word with the count, so subtracting twelve bits and then
// dividing is what makes the answer seventeen for version one rather than
// eighteen. A capacity a byte optimistic reports a link as fitting that a
// scanner cannot then read.
func qrCapacity(version int, ecc QRErrorCorrection) int {
	if version < 1 || version > 40 {
		return 0
	}
	countBits := 8
	if version >= 10 {
		countBits = 16
	}
	return (qrDataWords(version, ecc)*8 - (4 + countBits)) / 8
}

// qrDataWords is how many code words of a version carry data rather than
// error correction, at one level: all of them less what every block spends on
// correcting itself.
func qrDataWords(version int, ecc QRErrorCorrection) int {
	return qrTotalWords(version) - qrNumBlocks(version, ecc)*qrECCCodeWords(version, ecc)
}

// qrTotalWords is how many code words a version has in all. It is worked out
// from the module count rather than read out of a table of forty numbers,
// because a table is forty chances to type one wrong and no way to notice.
func qrTotalWords(version int) int {
	return qrRawModules(version) / 8
}

// qrRawModules is how many modules of a version are available for anything at
// all: the square, less the function patterns, which take a shape that grows
// with the alignment patterns rather than with the version.
func qrRawModules(version int) int {
	result := (16*version+128)*version + 64
	if version >= 2 {
		align := version/7 + 2
		// Each alignment pattern is five modules square, and the ones in a
		// row overlap the one before it by a module, which is where the 25
		// and the 10 and the 55 come from: the square of the count, the
		// shared edges, and the fixed part of the finders and the timing
		// patterns themselves.
		result -= (25*align-10)*align - 55
		if version >= 7 {
			// The second run of alignment patterns, along the other axis,
			// overlaps the first one and the finder patterns at the corners.
			result -= 36
		}
	}
	return result
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
