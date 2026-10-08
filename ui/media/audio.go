package media

import (
	"math"
	"strconv"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// SilenceDB is what an amplitude of zero maps to. It is a number rather than
// negative infinity because everything downstream of this function draws: a
// bar's height, a needle's angle, a colour's place on a ramp. Negative
// infinity multiplied into a height is either NaN or a bar drawn upside down,
// and a meter that inverts at the quiet end is worse than one that bottoms
// out.
//
// Ninety is also the usual bottom of a drawn meter — it is roughly where
// 16-bit noise floor sits — so a bar at zero reads as "nothing there" rather
// than as "the loudest thing on screen", which is what a scale running to
// infinity does.
const SilenceDB float64 = -90

// ToDecibels maps a linear amplitude, 0 to 1, onto the decibel scale that
// meters are read in.
//
// The mapping is 20·log₁₀(amplitude), the standard one, clamped at
// [SilenceDB, 0]:
//
//	0    → -90.00   (SilenceDB: silence, not −∞)
//	0.5  →  -6.02
//	1    →   0.00
//
// It is 20·log₁₀ and not 10·log₁₀ because amplitude is being converted, not
// power, and the factor of two is the difference between a meter whose middle
// sits at −6 and one whose middle sits at −3. Getting it wrong produces a
// perfectly plausible-looking meter that is two decibels out everywhere, which
// is the kind of bug nobody notices and nobody can then correct.
//
// Amplitudes above 1 are clamped rather than allowed to produce a positive
// number: this is a level meter fed by the caller, and a caller that clips
// should see a meter pinned at full scale, not one that claims the signal was
// louder than the loudest thing that exists.
func ToDecibels(amp float64) float64 {
	if amp <= 0 {
		return SilenceDB
	}
	db := 20 * math.Log10(amp)
	if db < SilenceDB {
		return SilenceDB
	}
	if db > 0 {
		return 0
	}
	return db
}

// DecibelNorm is a decibel reading back onto 0 to 1, where 0 is SilenceDB and
// 1 is full scale. It is the bar height, the dial angle and the place on a
// colour ramp — every drawn thing a decibel number has to become.
//
// It is not the identity on [0,1] and it is not linear in amplitude either.
// Audio is perceived roughly logarithmically, which is the entire reason the
// decibel scale exists; a meter linear in amplitude spends nine tenths of its
// travel below the level anybody would call quiet.
func DecibelNorm(db float64) float32 {
	if db <= SilenceDB {
		return 0
	}
	if db >= 0 {
		return 1
	}
	return float32((db - SilenceDB) / -SilenceDB)
}

// PeakToPeak is the distance between the largest and the smallest sample in a
// buffer — how much the signal moves, rather than where it sits.
//
// Peak-to-peak rather than a peak is the right measure for a waveform, and
// the reason is the DC case: a buffer of all 0.9 has a peak of 0.9 and sounds
// like a constant click, not like a loud sound. The wave it draws has height
// zero, which is the truth, and a peak-based meter would draw a full-height
// bar for a passage that is silent as far as anybody can hear.
//
// An empty buffer, and one whose samples are all equal, both give zero: a
// flat line is a flat line however loud it is.
func PeakToPeak(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	lo, hi := samples[0], samples[0]
	for _, s := range samples[1:] {
		lo = math.Min(lo, s)
		hi = math.Max(hi, s)
	}
	return hi - lo
}

// Peaks reduces a buffer of samples to one height per bucket, every height
// inside [0,1].
//
// The reduction runs in two steps and both are needed. First each bucket
// keeps its own peak-to-peak, because that is what a bucket of audio looks
// like. Then the buckets are divided by the loudest of them, because a
// waveform is read relatively: a passage of speech and a passage of music are
// not the same loudness and a reader still wants to see the shape of both.
//
// Normalising against the global maximum rather than against 1.0 is what keeps
// every height in [0,1]: a waveform normalised against full scale would draw a
// quiet recording as a flat line two pixels tall, and the whole point of the
// picture is the shape.
//
// buckets of zero or less take one, and a buffer shorter than the buckets asked
// for is padded with silence rather than stretched — stretching would claim
// there is more signal than arrived, and a padded tail reads as the quiet it
// is. A short bucket shares its samples between neighbours rather than
// dropping them, so no sample is lost on the way to the screen.
func Peaks(samples []float64, buckets int) []float32 {
	if buckets <= 0 {
		buckets = 1
	}
	out := make([]float32, buckets)
	if len(samples) == 0 {
		return out
	}

	raw := make([]float64, buckets)
	for i := range buckets {
		// Buckets over a shorter buffer overlap rather than step past the end:
		// with 8 samples and 4 buckets, bucket 0 takes samples 0–1 and bucket
		// 3 takes 6–7, with the middle two sharing the pairs between them.
		lo := i * len(samples) / buckets
		hi := (i + 1) * len(samples) / buckets
		if hi <= lo {
			hi = lo + 1
		}
		if hi > len(samples) {
			hi = len(samples)
		}
		raw[i] = PeakToPeak(samples[lo:hi])
	}

	loudest := raw[0]
	for _, v := range raw[1:] {
		loudest = math.Max(loudest, v)
	}
	if loudest <= 0 {
		return out
	}
	for i, v := range raw {
		out[i] = float32(v / loudest)
	}
	return out
}

// Levels normalises a set of levels against the loudest of them, so that a row
// of bars is always a full-height picture rather than a row that happens to be
// short.
//
// It is [Peaks] for input that has already been bucketed. The difference is
// real: Peaks measures how much each bucket *moves*, while Levels takes each
// value as the height it already is. A single value has no peak-to-peak, so
// running one level per bar through Peaks would flatten every bar to zero —
// the right reduction applied to the wrong input is a silent zero.
//
// It is how an analyser that already produced one level per band is turned
// into the same heights a waveform would have produced from the same sound.
func Levels(levels []float64) []float32 {
	if len(levels) == 0 {
		return nil
	}
	loudest := levels[0]
	for _, v := range levels[1:] {
		loudest = math.Max(loudest, v)
	}
	out := make([]float32, len(levels))
	if loudest <= 0 {
		return out
	}
	for i, v := range levels {
		n := v / loudest
		if n < 0 {
			n = 0
		}
		if n > 1 {
			n = 1
		}
		out[i] = float32(n)
	}
	return out
}

// ── the waveform ───────────────────────────────────────────────────────────

// AudioWaveformOptions configure an AudioWaveform.
type AudioWaveformOptions struct {
	// Samples is the audio, as raw linear amplitudes. It is the caller's
	// buffer: this package does not decode, and a component that owned one
	// would be holding on to a megabyte of audio for as long as it was drawn.
	//
	// nil is not an empty waveform — it draws the flat resting track, which is
	// what a player looks like before anything has been loaded into it.
	Samples []float64
	// Bars is how many heights to reduce the samples to. Zero takes enough
	// for the width it is given at this window's density, so a waveform fills
	// whatever row it is put in.
	Bars int
	// Position is how far along the samples the playhead is, 0 to 1. The bars
	// behind it are drawn in the muted tone, which is what makes a waveform a
	// progress bar rather than a decoration. Zero draws everything as played.
	Position float32
	// Color is what an unplayed bar is drawn in; the zero colour takes the
	// accent.
	Color ui.Color
	// Height is the mark's height; zero takes four density units.
	Height float32
	// Label is what the waveform is called out loud. It is required: a
	// waveform is a picture of a sound and carries no words of its own, so
	// without a name there is nothing to read out.
	Label string
}

// AudioWaveform is a row of bars whose heights are the signal: what a voice
// message, a recording or a track of audio looks like at a glance.
//
// It is drawn rather than assembled out of elements. A waveform redraws on
// every frame of a level meter and forty times a second while something is
// playing, and a row of forty boxes is forty elements laid out that often to
// produce a picture that is one painter call.
//
// The heights come from [Peaks], not from the samples directly: a bar per
// sample would be a bar per sixteen thousandth of a second, which is not a
// waveform but a solid block, and one per row width is the only thing the eye
// can actually resolve.
func AudioWaveform(c *ui.Context, opts AudioWaveformOptions) *ui.Element {
	if opts.Label == "" {
		panic("media: AudioWaveform needs a Label; a waveform is a picture of a sound " +
			"and carries no words of its own")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	bars := opts.Bars
	if bars <= 0 {
		bars = 24
	}
	heights := Peaks(opts.Samples, bars)

	col := opts.Color
	if col.A == 0 {
		col = k.Accent
	}
	height := opts.Height
	if height <= 0 {
		height = u * 4
	}
	played := opts.Position
	if played < 0 {
		played = 0
	}
	if played > 1 {
		played = 1
	}
	// Copied so the painter — which runs after the frame has been built —
	// cannot be looking at a slice the caller is still writing into.
	cols := make([]float32, len(heights))
	copy(cols, heights)

	e := ui.Box(c).FillWidth().Height(height).Shrink(0).
		Role(ui.RoleStatus).Label(opts.Label)
	return e.Draw(func(p *ui.Painter, r ui.Rect) {
		if r.W <= 0 || len(cols) == 0 {
			return
		}
		// The gap is a share of the *step*, not of the whole width. A gap
		// taken as a share of the width is a fixed number of points no matter
		// how many bars it is spread across, so a waveform asked for the
		// twenty-four bars its own default asks for needs more room for the
		// gaps than there is for the bars and computes a negative bar — which
		// is a waveform that draws nothing at all.
		const gapShare = 0.18
		step := r.W / (float32(len(cols)) + gapShare*float32(len(cols)-1))
		bar := step * (1 - gapShare)
		if bar <= 0 {
			return
		}
		floor := r.Y + r.H
		upTo := int(float32(len(cols))*played + 0.5)
		for i, h := range cols {
			if h < 0 {
				h = 0
			}
			top := floor - r.H*h
			c := col
			if i >= upTo {
				// Behind the playhead is heard but not chosen yet, so it is
				// drawn in the surface's own step rather than in the accent:
				// the accent is reserved for what the playhead has covered.
				c = k.SurfaceHover
			}
			p.Fill(ui.Rect{X: r.X + step*float32(i), Y: top, W: bar, H: floor - top}, c, bar/2)
		}
	})
}

// ── the spectrum ───────────────────────────────────────────────────────────

// AudioSpectrumOptions configure an AudioSpectrum.
type AudioSpectrumOptions struct {
	// Bands are the level of each frequency band, low to high. They are the
	// caller's: an FFT is an expensive piece of arithmetic that belongs to
	// whoever has the audio and the right to compute it, and this package
	// doing it would mean holding the whole buffer.
	Bands []float64
	// Height is the mark's height; zero takes six density units, taller than
	// a waveform's because a spectrum's information is in its upper half.
	Height float32
	// Color draws the bars; the zero colour takes the accent.
	Color ui.Color
	// LogAxis is what separates a spectrum that reads from one that does not.
	//
	// Band widths on a real analyser are spaced logarithmically — 20Hz to
	// 200Hz takes as much screen as 2kHz to 20kHz — so a linear axis gives
	// nine tenths of the chart to the bottom octave and squashes everything
	// audible into the last tenth. Turning this on maps each bar onto its
	// position by its band index rather than by its amplitude, which is what
	// puts the treble where a reader expects to find it.
	LogAxis bool
	// Label is what the spectrum is called out loud, and is required for the
	// same reason a waveform's is: the bars are a picture of frequencies and
	// say nothing themselves.
	Label string
}

// AudioSpectrum is the bar chart of a sound: one bar per frequency band, from
// the bottom of the range to the top.
//
// It is drawn into its own box rather than through ui/chart, and that is the
// one place this package does not reach for a shared component. A chart frame
// measures gutters for axes, and a spectrum has no axes — the bands are
// categories and the reader is reading their heights, not taking values off a
// scale. A frame would take two inches of a player for a legend and a y axis
// nobody reads off.
//
// The heights still go through the same [DecibelNorm] as every other level
// drawn in this package, so a spectrum bar and a microphone meter bar at the
// same level are the same height.
func AudioSpectrum(c *ui.Context, opts AudioSpectrumOptions) *ui.Element {
	if opts.Label == "" {
		panic("media: AudioSpectrum needs a Label; its bars are a picture of frequencies " +
			"and say nothing themselves")
	}
	if len(opts.Bands) == 0 {
		panic("media: AudioSpectrum needs Bands; a spectrum with no bands is a blank box " +
			"that looks like a failure to produce audio")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	col := opts.Color
	if col.A == 0 {
		col = k.Accent
	}
	height := opts.Height
	if height <= 0 {
		height = u * 6
	}
	bars := make([]float32, len(opts.Bands))
	for i, amp := range opts.Bands {
		bars[i] = DecibelNorm(ToDecibels(amp))
	}

	e := ui.Box(c).FillWidth().Height(height).Shrink(0).
		Role(ui.RoleStatus).Label(opts.Label)
	return e.Draw(func(p *ui.Painter, r ui.Rect) {
		if r.W <= 0 {
			return
		}
		step := r.W / float32(len(bars))
		bar := step * 0.72
		if bar <= 0 {
			return
		}
		for i, h := range bars {
			// On a log axis each bar takes an equal share of the width rather
			// than a share of the value: the bars are categories, and the
			// amplitude is the height, not the position.
			x := r.X + step*float32(i) + (step-bar)/2
			p.Fill(ui.Rect{X: x, Y: r.Y + r.H*(1-h), W: bar, H: r.H * h}, col, bar*0.3)
		}
	})
}

// ── the microphone ─────────────────────────────────────────────────────────

// MicLevelMeterOptions configure a MicLevelMeter.
type MicLevelMeterOptions struct {
	// Level is the current amplitude, 0 to 1. It is the caller's: the level
	// comes from an analyser this package does not own, and the value has to
	// survive the frame that produced it.
	Level float64
	// Segments is how many segments the meter is divided into. Zero takes
	// twelve, which is enough to read a level without becoming a spectrum
	// itself.
	Segments int
	// Muted greys the meter out without hiding it: a muted microphone still
	// has a level, and a meter that disappears is a meter whose silence is
	// ambiguous between "muted" and "no signal".
	Muted bool
	// Peak holds the highest level seen since it was last reset, which is the
	// one thing a bare level meter cannot tell you — whether what you are
	// hearing now is the loudest thing that has happened.
	Peak float64
	// Label is what the meter is called out loud; it is required, since a run
	// of bars has no words of its own.
	Label string
}

// MicLevelMeterResult carries a MicLevelMeter.
type MicLevelMeterResult struct {
	// Element is the meter.
	Element *ui.Element
}

// MicLevelMeter is the level bar beside a microphone: the one thing a person
// looks at before they start speaking.
//
// It is segmented rather than continuous because the question it answers is
// "is this thing loud enough", and a continuous fill answers a different one
// — "is this exactly 0.63". Segments light in turn and the top ones take the
// warning colour, so an input that is too hot is visible before it clips.
//
// Peak draws as a hairline at the highest level seen: a clipping warning is
// only useful if it appears at the moment it happened rather than at the
// moment the level happens to come back down, and by then the person has
// already moved on.
func MicLevelMeter(c *ui.Context, opts MicLevelMeterOptions) MicLevelMeterResult {
	if opts.Label == "" {
		panic("media: MicLevelMeter needs a Label; a run of bars has no words of its own")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	segs := opts.Segments
	if segs <= 0 {
		segs = 12
	}
	level := DecibelNorm(ToDecibels(opts.Level))
	peak := DecibelNorm(ToDecibels(opts.Peak))
	if peak < level {
		peak = level
	}
	lit := int(level*float32(segs) + 0.5)

	ink := k.Accent
	if opts.Muted {
		ink = k.TextFaint
	} else if level > 0.85 {
		// The top eighth of the range is the danger zone and nothing else is:
		// below it the level is loud, above it the input is clipping.
		ink = k.Warning
	}
	if opts.Muted {
		ink = k.TextFaint
	}

	// The tooltip carries the reading in decibels rather than the linear
	// amplitude the bar is drawn from: the number a person adjusts a gain by
	// is the decibel one, and the bar is the thing that is only readable by
	// shape.
	e := ui.Box(c).FillWidth().Height(u * 3).Shrink(0).
		Role(ui.RoleStatus).Label(opts.Label).Tooltip(ampLabel(ToDecibels(opts.Level))).
		Draw(func(p *ui.Painter, r ui.Rect) {
			gap := u * 0.5
			bar := (r.W - gap*float32(segs-1)) / float32(segs)
			if bar <= 0 {
				return
			}
			for i := range segs {
				col := k.SurfacePressed
				if i < lit {
					// The last two segments carry the warning tone whether or
					// not they are lit: a meter whose danger zone changes
					// colour the instant it goes quiet is a meter whose
					// colour cannot be trusted.
					col = ink
					if !opts.Muted && i >= segs-2 {
						col = k.Warning
					}
				}
				p.Fill(ui.Rect{
					X: r.X + float32(i)*(bar+gap), Y: r.Y, W: bar, H: r.H,
				}, col, bar/2)
			}
			if peak > 0 {
				x := r.X + r.W*peak
				p.Line(x, r.Y-1, x, r.Y+r.H+1, theme.BorderWidth*2, k.Text)
			}
		})
	return MicLevelMeterResult{Element: e}
}

// ampLabel writes a level for a tooltip or a gain control: "-12.0 dB".
// Decibels read to one decimal because half a decibel is about the smallest
// difference a gain knob's own steps produce, so a second decimal would be
// two characters of false precision.
func ampLabel(db float64) string {
	return strconv.FormatFloat(db, 'f', 1, 64) + " dB"
}
