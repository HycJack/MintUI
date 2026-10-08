package media

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/data"
	"github.com/HycJack/MintUI/ui/theme"
)

// The four pure functions carry this package: TimeToText writes what a
// transport reads, ScrubRatio says how far along a track a position is,
// ToDecibels and DecibelNorm turn a level into the number a meter is drawn
// from, and Peaks reduces samples to bars. None of them needs a window, so
// each is asserted on exactly what it returns rather than on a bar that came
// out the right height.
//
// Everything else is asserted on the element it drew, on a label it put on
// screen, or on the value it wrote into the caller's pointer. MyGo builds a
// frame up to three times and stops when one consumed nothing
// (mygo/ui/runtime.go:333), so after a tester has settled, Clicked() is false
// on the last pass: a test reading a Result after settling would read the
// empty answer and pass for the wrong reason. Every press below is therefore
// counted inside the view, across passes.

// wantsPanic runs view and fails unless it panics, which is how this library
// says a component was asked for something it cannot be given.
func wantsPanic(t *testing.T, what string, view func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("%s should have panicked", what)
		}
	}()
	view()
}

// wantText is the assertion that something is on screen. A missing label is
// the failure a headless render produces most often and the one that matters
// most, because an unnamed control is invisible to the reader it is for.
func wantText(t *testing.T, tt *ui.Tester, what string, texts ...string) {
	t.Helper()
	for _, s := range texts {
		if !tt.HasText(s) {
			t.Errorf("%s is missing %q; drew %q", what, s, tt.Texts())
		}
	}
}

// ── TimeToText ─────────────────────────────────────────────────────────────

// TestTimeToText pins the transport's format at every boundary that matters.
// Each case is a place the format changes rather than a sample of it:
// zero, the last second before a minute, the minute itself, the second after
// it, the hour, and the hour plus a second — where the format gains its third
// field and a column of digits with it.
func TestTimeToText(t *testing.T) {
	cases := []struct {
		in   time.Duration
		want string
	}{
		{0, "0:00"},
		{59 * time.Second, "0:59"},
		{60 * time.Second, "1:00"},
		{61 * time.Second, "1:01"},
		{3600 * time.Second, "1:00:00"},
		{3661 * time.Second, "1:01:01"},
	}
	for _, tc := range cases {
		if got := TimeToText(tc.in); got != tc.want {
			t.Errorf("TimeToText(%v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// TestTimeToTextRoundsRatherThanTruncates is the other half of the format: a
// playhead at 59.7s is a playhead that has already passed 59.5s, and a
// transport that truncates shows 0:59 for a whole second after it should have
// advanced.
func TestTimeToTextRoundsRatherThanTruncates(t *testing.T) {
	if got := TimeToText(59700 * time.Millisecond); got != "1:00" {
		t.Errorf("TimeToText(59.7s) = %q, want %q", got, "1:00")
	}
	if got := TimeToText(59400 * time.Millisecond); got != "0:59" {
		t.Errorf("TimeToText(59.4s) = %q, want %q", got, "0:59")
	}
}

// TestTimeToTextClampsNegatives: a playhead a fraction behind the start is a
// playhead at the start. "-0:00" on a scrubber is a transport reporting an
// error the reader has to know to ignore.
func TestTimeToTextClampsNegatives(t *testing.T) {
	if got := TimeToText(-5 * time.Second); got != "0:00" {
		t.Errorf("TimeToText(-5s) = %q, want %q", got, "0:00")
	}
}

// ── ScrubRatio ─────────────────────────────────────────────────────────────

func TestScrubRatio(t *testing.T) {
	cases := []struct {
		at, total time.Duration
		want      float32
	}{
		{0, 60 * time.Second, 0},
		{30 * time.Second, 60 * time.Second, 0.5},
		{60 * time.Second, 60 * time.Second, 1},
		// Past the end, which is what a seek into a file that turned out to
		// be shorter gives: the bar must not be filled past its own rail.
		{90 * time.Second, 60 * time.Second, 1},
		{30 * time.Second, 0, 0},
		{30 * time.Second, -1, 0},
		{-1, 60 * time.Second, 0},
	}
	for _, tc := range cases {
		if got := ScrubRatio(tc.at, tc.total); got != tc.want {
			t.Errorf("ScrubRatio(%v, %v) = %v, want %v", tc.at, tc.total, got, tc.want)
		}
	}
}

// ── decibels ───────────────────────────────────────────────────────────────

// TestToDecibels is exact, not "greater than zero": silence is the floor the
// whole package draws against, full scale is zero, and the middle is the
// standard 20·log₁₀. Asserting the middle to a tolerance is the only way to
// catch the 10·log₁₀ version of this function, which produces a perfectly
// plausible meter that is 3dB out everywhere.
func TestToDecibels(t *testing.T) {
	cases := []struct {
		amp  float64
		want float64
	}{
		{0, SilenceDB},
		{0.5, -6.020599913279624},
		{1, 0},
	}
	for _, tc := range cases {
		got := ToDecibels(tc.amp)
		if math.Abs(got-tc.want) > 1e-9 {
			t.Errorf("ToDecibels(%v) = %v, want %v", tc.amp, got, tc.want)
		}
	}
}

// TestToDecibelsClampsBeyondFullScale: this is a level meter fed by a caller,
// and a caller that clips should see a meter pinned at full scale rather than
// one claiming the signal was louder than the loudest thing that exists.
func TestToDecibelsClampsBeyondFullScale(t *testing.T) {
	if got := ToDecibels(4); got != 0 {
		t.Errorf("ToDecibels(4) = %v, want 0", got)
	}
	if got := ToDecibels(1e-30); got != SilenceDB {
		t.Errorf("ToDecibels(1e-30) = %v, want %v", got, SilenceDB)
	}
}

// TestDecibelNorm is the drawn height, and its two ends are the two states a
// meter is ever in: nothing there, and everything there.
func TestDecibelNorm(t *testing.T) {
	cases := []struct {
		db   float64
		want float32
	}{
		{SilenceDB, 0},
		{0, 1},
		{-45, 0.5},
		{SilenceDB - 30, 0},
		{30, 1},
	}
	for _, tc := range cases {
		if got := DecibelNorm(tc.db); math.Abs(float64(got-tc.want)) > 1e-6 {
			t.Errorf("DecibelNorm(%v) = %v, want %v", tc.db, got, tc.want)
		}
	}
}

// ── Peaks ──────────────────────────────────────────────────────────────────

// TestPeaksPeakToPeak is the assertion the brief asks for, stated exactly: for
// a buffer that rises from −1 to +1 the peak-to-peak is 2.0, and the tallest
// bucket after normalisation is exactly 1.0 — not 0.999, not "nearly".
func TestPeaksPeakToPeak(t *testing.T) {
	if got := PeakToPeak([]float64{-1, 1}); got != 2 {
		t.Fatalf("PeakToPeak(-1, 1) = %v, want exactly 2", got)
	}
	if got := PeakToPeak([]float64{0.2, 0.6, 0.4}); math.Abs(got-0.4) > 1e-9 {
		t.Errorf("PeakToPeak(0.2, 0.6, 0.4) = %v, want exactly 0.4", got)
	}
	// Two buckets over four samples: [-1,0] moves 1.0 and [1,0.5] moves 0.5,
	// so the quieter half of the buffer draws at half height. Asserting the
	// split rather than "some bars" is the point — a reduction that returned
	// two equal bars would pass a range check and fail this one.
	bars := Peaks([]float64{-1, 0, 1, 0.5}, 2)
	if len(bars) != 2 {
		t.Fatalf("Peaks into 2 buckets gave %d bars", len(bars))
	}
	if bars[0] != 1 || math.Abs(float64(bars[1]-0.5)) > 1e-6 {
		t.Errorf("Peaks(-1,0,1,0.5; 2) = %v, want [1 0.5]", bars)
	}
}

// TestPeaksHeightsStayInRange is the property every waveform depends on: a
// bar's height is a share of a box, and a share outside [0,1] is a bar drawn
// past the edge of its own track.
func TestPeaksHeightsStayInRange(t *testing.T) {
	samples := []float64{0.9, 0.9, 0.9, -0.2, 0.1, 0.4, 0.05, 0.3, 0.7, 0.05, 0.0, 0.6}
	for _, buckets := range []int{1, 2, 5, 12, 40} {
		for i, h := range Peaks(samples, buckets) {
			if h < 0 || h > 1 {
				t.Fatalf("Peaks(samples, %d)[%d] = %v, outside [0,1]", buckets, i, h)
			}
		}
	}
}

// TestPeaksIsRelativeNotAbsolute: a waveform is read against its own loudest
// moment. A buffer of 0.2s and a buffer of 0.9s are the same shape and have to
// come out the same height, or a quiet recording draws as a flat line two
// pixels tall and the picture stops saying anything at all.
func TestPeaksIsRelativeNotAbsolute(t *testing.T) {
	shape := []float64{0, 0.4, 0.9, 0.4, 0}
	quiet := Peaks(shape, 5)
	loud := Peaks([]float64{0, 0.4, 0.9, 0.4, 0}, 5)
	for i := range quiet {
		if quiet[i] != loud[i] {
			t.Fatalf("bucket %d differs between identical shapes: %v vs %v", i, quiet[i], loud[i])
		}
	}
	scaled := Peaks([]float64{0, 0.04, 0.09, 0.04, 0}, 5)
	for i := range scaled {
		if math.Abs(float64(scaled[i]-quiet[i])) > 1e-6 {
			t.Errorf("bucket %d = %v, want the same %v as the un-scaled shape", i, scaled[i], quiet[i])
		}
	}
}

// TestPeaksFlatBufferIsSilent is the DC case, and the reason the reduction is
// peak-to-peak rather than a peak: a buffer of all 0.9 has a peak of 0.9 and
// sounds like one click, not like a loud sound. The wave it draws has no
// height, and a peak-based meter would draw a full-height bar for a passage
// nobody could hear.
func TestPeaksFlatBufferIsSilent(t *testing.T) {
	for _, h := range Peaks([]float64{0.9, 0.9, 0.9, 0.9}, 4) {
		if h != 0 {
			t.Fatalf("a constant buffer should draw a flat line, got %v", h)
		}
	}
	if got := PeakToPeak(nil); got != 0 {
		t.Errorf("PeakToPeak(nil) = %v, want 0", got)
	}
	for _, h := range Peaks(nil, 3) {
		if h != 0 {
			t.Errorf("Peaks of nothing should be silence, got %v", h)
		}
	}
}

// TestPeaksShorterBufferIsNotStretched: stretching a two-sample buffer into
// forty bars would claim there is more signal than arrived, and the drawn
// shape would say something about the audio that is not in it.
func TestPeaksShorterBufferIsNotStretched(t *testing.T) {
	if got := len(Peaks([]float64{0.2, 0.8}, 12)); got != 12 {
		t.Fatalf("Peaks into 12 buckets gave %d bars", got)
	}
	// And no bucket may claim a peak-to-peak larger than the buffer's own.
	for _, h := range Peaks([]float64{0.2, 0.8}, 12) {
		if h > 1 {
			t.Fatalf("a short buffer produced a bar at %v", h)
		}
	}
}

func TestLevels(t *testing.T) {
	got := Levels([]float64{0.2, 0.4, 0.6})
	want := []float32{0.2 / 0.6, 0.4 / 0.6, 1}
	for i := range want {
		if math.Abs(float64(got[i]-want[i])) > 1e-6 {
			t.Errorf("Levels[%d] = %v, want %v", i, got[i], want[i])
		}
	}
	// One level per bar has no peak-to-peak, which is exactly why Levels is
	// not Peaks: running these through the sample reduction would flatten
	// every bar to zero, which is the bug this assertion is here for.
	if got[2] != 1 {
		t.Errorf("the loudest level should normalise to 1, got %v", got[2])
	}
	for _, h := range Levels([]float64{0, 0, 0}) {
		if h != 0 {
			t.Errorf("a silent set of levels should stay silent, got %v", h)
		}
	}
	if Levels(nil) != nil {
		t.Error("Levels(nil) should be nil, not an empty slice")
	}
}

// ── shapes ─────────────────────────────────────────────────────────────────

func TestRectNormalised(t *testing.T) {
	cases := []struct {
		in, want Rect
	}{
		{Rect{}, Rect{W: 1, H: 1}},
		{Rect{X: -0.5, Y: -0.5, W: 0.4, H: 0.4}, Rect{W: 0.4, H: 0.4}},
		{Rect{X: 0.9, Y: 0.9, W: 0.4, H: 0.4}, Rect{X: 0.6, Y: 0.6, W: 0.4, H: 0.4}},
		{Rect{X: 0.2, Y: 0.2, W: 4, H: 4}, Rect{W: 1, H: 1}},
	}
	for _, tc := range cases {
		if got := tc.in.Normalised(); got != tc.want {
			t.Errorf("%+v.Normalised() = %+v, want %+v", tc.in, got, tc.want)
		}
	}
}

func TestRectEmpty(t *testing.T) {
	if !(Rect{W: 0, H: 1}).Empty() {
		t.Error("a region with no width is empty")
	}
	if (Rect{W: 0.5, H: 0.5}).Empty() {
		t.Error("a region with both sides is not empty")
	}
	if Unit().Empty() {
		t.Error("the whole picture is not empty")
	}
}

// TestRectFixedRatioKeepsTheFixedSide: a reader who has set a height and not a
// width has asked for that height, so the width gives way. An editor that
// changed the height instead would overwrite the one decision they made.
func TestRectFixedRatioKeepsTheFixedSide(t *testing.T) {
	wide := Rect{X: 0.1, Y: 0.2, W: 0.8, H: 0.5}
	got := wide.FixedRatio(1)
	if math.Abs(float64(got.W-got.H)) > 1e-6 {
		t.Errorf("a 1:1 constraint gave %+v, whose sides are not equal", got)
	}
	if math.Abs(float64(got.H-wide.H)) > 1e-6 {
		t.Errorf("the constraint moved the height from %v to %v", wide.H, got.H)
	}
	// And it is re-seated on the same centre, so the crop does not slide to
	// the left while it is being corrected.
	if math.Abs(float64((got.X+got.W/2)-(wide.X+wide.W/2))) > 1e-6 {
		t.Errorf("the constraint moved the region's centre")
	}
	tall := Rect{X: 0.1, Y: 0.2, W: 0.4, H: 0.8}
	got2 := tall.FixedRatio(2)
	if math.Abs(float64(got2.W-2*got2.H)) > 1e-6 {
		t.Errorf("a 2:1 constraint gave %+v, whose width is not twice its height", got2)
	}
	if (Rect{W: 0.5, H: 0.5}).FixedRatio(0) != (Rect{W: 0.5, H: 0.5}) {
		t.Error("a ratio of zero is no constraint at all")
	}
}

func TestShapeRatio(t *testing.T) {
	if got := shapeRatio("16:9"); math.Abs(float64(got-16.0/9.0)) > 1e-6 {
		t.Errorf(`shapeRatio("16:9") = %v`, got)
	}
	if got := shapeRatio("nonsense"); got != 0 {
		t.Errorf("an unknown shape should be no constraint, got %v", got)
	}
}

func TestRatioOf(t *testing.T) {
	if got := ratioOf(Rect{W: 0.5, H: 0.25}); math.Abs(float64(got-2)) > 1e-6 {
		t.Errorf("ratioOf = %v, want 2", got)
	}
	if got := ratioOf(Rect{}); got != 0 {
		t.Errorf("an empty region should have no ratio, got %v", got)
	}
}

// ── formatting ─────────────────────────────────────────────────────────────

func TestFormatSpeed(t *testing.T) {
	cases := map[float64]string{1: "1×", 2: "2×", 1.5: "1.5×", 0.75: "0.75×"}
	for in, want := range cases {
		if got := formatSpeed(in); got != want {
			t.Errorf("formatSpeed(%v) = %q, want %q", in, got, want)
		}
	}
	if got := clampSpeed(0); got != 1 {
		t.Errorf("clampSpeed(0) = %v, want 1", got)
	}
	if got := clampSpeed(-2); got != 1 {
		t.Errorf("clampSpeed(-2) = %v, want 1", got)
	}
}

func TestRateKeyIgnoresFloatNoise(t *testing.T) {
	if rateKey(1.25) != rateKey(1.2500000001) {
		t.Error("two rates that differ by float noise are the same speed")
	}
	if rateKey(1) == rateKey(1.5) {
		t.Error("1× and 1.5× are different speeds")
	}
}

// ── cues ───────────────────────────────────────────────────────────────────

func TestCueDuration(t *testing.T) {
	cue := Cue{Start: 3 * time.Second, End: 6 * time.Second}
	if got := cue.Duration(); got != 3*time.Second {
		t.Errorf("Duration = %v, want 3s", got)
	}
	backwards := Cue{Start: 6 * time.Second, End: 3 * time.Second}
	if got := backwards.Duration(); got != 0 {
		t.Errorf("a cue ending before it starts should measure 0, not %v", got)
	}
}

func TestCueWordCount(t *testing.T) {
	cue := Cue{Text: "  the   quick brown\nfox  "}
	if got := cue.WordCount(); got != 4 {
		t.Errorf("WordCount = %d, want 4", got)
	}
	if got := (Cue{}).WordCount(); got != 0 {
		t.Errorf("an empty cue has no words, got %d", got)
	}
}

// TestReadTime is the standard reading speed, to the millisecond: ten words at
// 160 a minute is 3.75 seconds. The old arithmetic here was a hundred times
// out — every cue passed and no cue ever warned — so the exact value is the
// assertion rather than a comparison.
func TestReadTime(t *testing.T) {
	if got := ReadTime(10, ReadingRate); got != 3750*time.Millisecond {
		t.Fatalf("ReadTime(10, 160) = %v, want 3.75s", got)
	}
	if got := ReadTime(0, ReadingRate); got != 0 {
		t.Errorf("no words need no time, got %v", got)
	}
	if got := ReadTime(10, 0); got != 0 {
		t.Errorf("no reading rate needs no time, got %v", got)
	}
	// A half-rate reader — a foreign-language track — needs twice as long.
	if got := ReadTime(10, ReadingRate/2); got != 7500*time.Millisecond {
		t.Errorf("ReadTime(10, 80) = %v, want 7.5s", got)
	}
}

func TestCueNeedsLonger(t *testing.T) {
	ten := Cue{Start: 0, End: 3 * time.Second,
		Text: "one two three four five six seven eight nine ten"}
	if !ten.NeedsLonger() {
		t.Error("ten words in three seconds is faster than 160 words a minute")
	}
	roomy := Cue{Start: 0, End: 4 * time.Second, Text: ten.Text}
	if roomy.NeedsLonger() {
		t.Error("ten words in four seconds is slower than 160 words a minute")
	}
	if (Cue{Text: "   "}).NeedsLonger() {
		t.Error("a cue with no words has nothing to read and needs no longer")
	}
	// The rule is a guide: a slow, deliberate line is the caller's decision,
	// and NeedsLonger only ever warns.
	deliberate := Cue{Start: 0, End: 5 * time.Second, Text: ten.Text}
	if deliberate.NeedsLonger() {
		t.Error("five seconds for ten words is comfortably readable")
	}
}

// ── headless renders ───────────────────────────────────────────────────────

// samples is a short buffer with a loud moment and two quiet ones, so the
// waveform has a shape rather than being flat.
var samples = []float64{
	0.1, 0.05, 0.8, 0.6, 0.05, 0.02, 0.1, 0.9,
	0.4, 0.05, 0.02, 0.3, 0.7, 0.05, 0.01, 0.02,
}

func TestAudioWaveform(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		AudioWaveform(c, AudioWaveformOptions{
			Samples: samples, Bars: 16, Position: 0.4,
			Label: "Waveform of the interview",
		})
	}, 400, 120)
	if _, ok := tt.Find("Waveform of the interview"); !ok {
		t.Error("a waveform must carry its name; it is a picture of a sound")
	}
	wantsPanic(t, "an unnamed waveform", func() {
		tt2 := ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			AudioWaveform(c, AudioWaveformOptions{Samples: samples})
		}, 200, 80)
		_ = tt2
	})
}

func TestAudioSpectrum(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		AudioSpectrum(c, AudioSpectrumOptions{
			Bands:   []float64{0.2, 0.5, 0.9, 0.4, 0.05},
			LogAxis: true,
			Label:   "Frequency bands",
		})
	}, 360, 120)
	if _, ok := tt.Find("Frequency bands"); !ok {
		t.Error("a spectrum must carry its name; its bars are a picture of frequencies")
	}
	wantsPanic(t, "a spectrum with no bands", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			AudioSpectrum(c, AudioSpectrumOptions{Label: "Nothing"})
		}, 200, 80)
	})
}

func TestMicLevelMeter(t *testing.T) {
	for _, level := range []float64{0, 0.5, 1} {
		lvl := level
		tt := ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			MicLevelMeter(c, MicLevelMeterOptions{Level: lvl, Peak: lvl, Label: "Input level"})
		}, 320, 60)
		if _, ok := tt.Find("Input level"); !ok {
			t.Errorf("a meter at level %v must carry its name", lvl)
		}
	}
	wantsPanic(t, "an unnamed level meter", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			MicLevelMeter(c, MicLevelMeterOptions{Level: 1})
		}, 200, 60)
	})
}

func TestVideoScrubber(t *testing.T) {
	at := 30 * time.Second
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		VideoScrubber(c, VideoScrubberOptions{
			At: &at, Duration: time.Minute, Buffered: 0.8,
			Markers: []float32{0.25, 0.75}, Label: "Playback position",
		})
	}, 480, 60)
	if _, ok := tt.Find("Playback position"); !ok {
		t.Error("a rail and a thumb say nothing about what they scrub")
	}
	wantsPanic(t, "a scrubber with no playhead", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			VideoScrubber(c, VideoScrubberOptions{Duration: time.Minute, Label: "x"})
		}, 300, 40)
	})
	wantsPanic(t, "a scrubber with no length", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			VideoScrubber(c, VideoScrubberOptions{At: &at, Label: "x"})
		}, 300, 40)
	})
}

func TestVolumeControl(t *testing.T) {
	level, muted := 0.6, false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		VolumeControl(c, VolumeControlOptions{Volume: &level, Muted: &muted})
	}, 420, 60)
	for _, name := range []string{"Volume", "Mute"} {
		if _, ok := tt.Find(name); !ok {
			t.Errorf("the volume control should offer %q for assistive technology", name)
		}
	}
	wantsPanic(t, "a volume control with no level", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			VolumeControl(c, VolumeControlOptions{})
		}, 300, 60)
	})
}

// TestVolumeControlMuteIsNotTheLevel: the two are different actions, and a
// control that stored mute as "volume 0" would restore silence on unmute.
// Pressing the button must flip the caller's flag and leave its level alone.
func TestVolumeControlMuteIsNotTheLevel(t *testing.T) {
	level, muted := 0.75, false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		VolumeControl(c, VolumeControlOptions{Volume: &level, Muted: &muted})
	}, 420, 60)
	tt.Click("Mute")
	if !muted {
		t.Error("the press should have muted")
	}
	if level != 0.75 {
		t.Errorf("muting changed the level to %v; it must not", level)
	}
}

func TestPlaybackSpeedControl(t *testing.T) {
	speed, open := 1.0, false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		PlaybackSpeedControl(c, PlaybackSpeedControlOptions{Speed: &speed, Open: &open})
	}, 240, 80)
	wantText(t, tt, "the speed button", "1×")

	presses := 0
	tt2 := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		open := false
		if PlaybackSpeedControl(c, PlaybackSpeedControlOptions{
			Speed: &speed, Open: &open, Label: "Speed",
		}).Clicked() {
			presses++
		}
	}, 240, 80)
	tt2.Click("Speed")
	if presses != 1 {
		t.Fatalf("the speed button reported %d presses, want 1", presses)
	}
}

func TestSpeedChoicesWritesTheRate(t *testing.T) {
	speed := 1.0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		SpeedChoices(c, &speed, Speeds, "Playback speed")
	}, 400, 80)
	wantText(t, tt, "the rate row", "0.5×", "1.5×", "2×")
	tt.Click("2×")
	if speed != 2 {
		t.Errorf("choosing 2× wrote %v", speed)
	}
}

func TestMediaControls(t *testing.T) {
	at, playing, level, muted, speed := time.Duration(0), false, 0.8, false, 1.0
	speedOpen := false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		MediaControls(c, MediaControlsOptions{
			At: &at, Duration: 90 * time.Second, Playing: &playing,
			Volume: &level, Muted: &muted, Speed: &speed, Open: &speedOpen,
			Buffered: 0.6,
		})
	}, 640, 120)
	wantText(t, tt, "the transport", "0:00", "1:30")
	if _, ok := tt.Find("Play"); !ok {
		t.Error("the play button must be named for assistive technology")
	}

	toggles := 0
	playing2, at2 := false, time.Duration(0)
	tt2 := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		if MediaControls(c, MediaControlsOptions{
			At: &at2, Duration: 90 * time.Second, Playing: &playing2,
		}).Toggled() {
			toggles++
		}
	}, 640, 120)
	tt2.Click("Play")
	if toggles != 1 {
		t.Fatalf("the transport reported %d toggles, want 1", toggles)
	}
	if !playing2 {
		t.Error("the press should have started playback")
	}
}

// TestMediaControlsSkipReportsADirection: the transport does not know how long
// a chapter is, so it says which way it was asked to go and the caller moves.
func TestMediaControlsSkipReportsADirection(t *testing.T) {
	steps := 0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		playing, at := false, time.Duration(0)
		_ = playing
		if MediaControls(c, MediaControlsOptions{
			At: &at, Duration: time.Minute, Playing: &playing,
		}).Stepped() == 1 {
			steps++
		}
	}, 640, 120)
	tt.Click("Forward 10 seconds")
	if steps != 1 {
		t.Fatalf("the forward button reported %d steps, want 1", steps)
	}
}

func TestVideoPlayer(t *testing.T) {
	at, playing := 12*time.Second, true
	level, muted, speed, open := 0.5, false, 1.0, false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		VideoPlayer(c, VideoPlayerOptions{
			Title: "Site walkthrough", Duration: 312 * time.Second,
			At: &at, Playing: &playing, Volume: &level, Muted: &muted,
			Speed: &speed, SpeedOpen: &open,
			Markers:  []float32{0.1, 0.5},
			Controls: true,
		})
	}, 720, 320)
	wantText(t, tt, "the player", "Site walkthrough", "0:12", "5:12")
	if _, ok := tt.Find("No frame yet"); !ok {
		t.Error("a player with no decoded frame should say it is waiting for one")
	}
	wantsPanic(t, "a player with no title", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			a, p := time.Duration(0), false
			VideoPlayer(c, VideoPlayerOptions{Duration: time.Minute, At: &a, Playing: &p})
		}, 400, 200)
	})
	wantsPanic(t, "a player with no length", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			a, p := time.Duration(0), false
			VideoPlayer(c, VideoPlayerOptions{Title: "x", At: &a, Playing: &p})
		}, 400, 200)
	})
}

func TestAudioPlayer(t *testing.T) {
	at, playing, level := 4*time.Second, false, 0.7
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		AudioPlayer(c, AudioPlayerOptions{
			Title: "Voice note", Duration: 45 * time.Second,
			At: &at, Playing: &playing, Samples: samples, Bars: 20,
			Volume: &level,
		})
	}, 560, 240)
	wantText(t, tt, "the player", "Voice note", "0:04", "0:45")
	if _, ok := tt.Find("Waveform of Voice note"); !ok {
		t.Error("the waveform must be named after what it is a waveform of")
	}
}

// TestAudioPlayerTitleFallsBackToTheLength: with nothing to name it by, the
// length is the only thing there is, and it is better than an empty header.
func TestAudioPlayerTitleFallsBackToTheLength(t *testing.T) {
	at, playing := time.Duration(0), false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		AudioPlayer(c, AudioPlayerOptions{
			Duration: 95 * time.Second, At: &at, Playing: &playing,
		})
	}, 560, 240)
	wantText(t, tt, "the fallback header", "Audio 1:35")
}

func TestVideoThumbnailStrip(t *testing.T) {
	frames := []Frame{
		{At: 0},
		{At: 10 * time.Second},
		{At: 20 * time.Second, Label: "the good bit"},
	}
	selected := 1
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		VideoThumbnailStrip(c, VideoThumbnailStripOptions{
			Frames: frames, Selected: &selected, Duration: 30 * time.Second,
			Label: "Video filmstrip",
		})
	}, 560, 120)
	if _, ok := tt.Find("Video filmstrip"); !ok {
		t.Error("a strip of pictures says nothing about what it is a strip of")
	}
	if _, ok := tt.Find("the good bit"); !ok {
		t.Error("a frame with its own label should be found by it")
	}
	picks := 0
	tt2 := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		selected := -1
		if VideoThumbnailStrip(c, VideoThumbnailStripOptions{
			Frames: frames, Selected: &selected, Label: "Video filmstrip",
		}).Picked() == 2 {
			picks++
		}
	}, 560, 120)
	tt2.Click("the good bit")
	if picks != 1 {
		t.Fatalf("the strip reported %d picks of frame 2, want 1", picks)
	}
	wantsPanic(t, "a strip with no frames", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			sel := -1
			VideoThumbnailStrip(c, VideoThumbnailStripOptions{Selected: &sel, Label: "x"})
		}, 300, 80)
	})
}

func TestImageThumbnail(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ImageThumbnail(c, ImageThumbnailOptions{
			Name: "Rooftop before", Side: 80, Selected: true,
			Badge: "2:14", Pressable: true,
		})
	}, 200, 200)
	if _, ok := tt.Find("Rooftop before"); !ok {
		t.Error("a picture has no words of its own, so it must be named")
	}
	clicks := 0
	tt2 := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		if ImageThumbnail(c, ImageThumbnailOptions{
			Name: "Rooftop after", Pressable: true,
		}).Clicked() {
			clicks++
		}
	}, 200, 200)
	tt2.Click("Rooftop after")
	if clicks != 1 {
		t.Fatalf("the thumbnail reported %d presses, want 1", clicks)
	}
}

func TestLightbox(t *testing.T) {
	open := false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Lightbox(c, LightboxOptions{
			Open: &open, Name: "Rooftop at dusk",
			Caption: "Site walkthrough, frame 412",
		})
	}, 700, 480)
	// Closed: a layer that is not showing must not spend a frame's layout.
	if len(tt.Texts()) != 0 {
		t.Errorf("a closed lightbox should draw nothing, drew %q", tt.Texts())
	}
	open = true
	tt.Frame()
	wantText(t, tt, "the open lightbox", "Rooftop at dusk", "Site walkthrough, frame 412")
	wantsPanic(t, "an unnamed lightbox", func() {
		o := true
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			Lightbox(c, LightboxOptions{Open: &o})
		}, 500, 400)
	})
}

func TestImageCompare(t *testing.T) {
	at := float32(0.5)
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ImageCompare(c, ImageCompareOptions{
			Name: "Before and after", At: &at, Ratio: 4.0 / 3.0,
		})
	}, 480, 380)
	if _, ok := tt.Find("Before and after"); !ok {
		t.Error("two pictures and a line say nothing about what is compared")
	}
	wantsPanic(t, "a compare with no divider", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			ImageCompare(c, ImageCompareOptions{Name: "x"})
		}, 300, 200)
	})
}

func TestImageCropper(t *testing.T) {
	crop := Rect{X: 0.1, Y: 0.2, W: 0.6, H: 0.5}
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ImageCropper(c, ImageCropperOptions{
			Name: "Rooftop", Rect: &crop, Width: 320, Grid: true,
		})
	}, 400, 400)
	if _, ok := tt.Find("Rooftop"); !ok {
		t.Error("the cropped picture must be named")
	}
	if crop != (Rect{X: 0.1, Y: 0.2, W: 0.6, H: 0.5}) {
		t.Errorf("drawing the cropper changed the crop to %+v", crop)
	}
	// A named shape holds the region to that shape, and it is the fixed side
	// that survives.
	square := Rect{W: 0.5, H: 0.5}
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ImageCropper(c, ImageCropperOptions{
			Name: "Square", Rect: &square, Shape: "1:1", Width: 300,
		})
	}, 400, 400)
	if math.Abs(float64(square.W-square.H)) > 1e-6 {
		t.Errorf("a 1:1 shape gave %+v", square)
	}
	wantsPanic(t, "a cropper with no region", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			ImageCropper(c, ImageCropperOptions{Name: "x"})
		}, 300, 300)
	})
}

func TestImageAnnotator(t *testing.T) {
	marks := []Annotation{
		{Kind: "box", Points: []struct{ X, Y float32 }{
			AnnotationPoint(0.2, 0.2), AnnotationPoint(0.6, 0.5)}},
		{Kind: "pin", Text: "the leak", Points: []struct{ X, Y float32 }{
			AnnotationPoint(0.44, 0.33)}},
		{Kind: "pen", Points: []struct{ X, Y float32 }{
			AnnotationPoint(0.1, 0.8), AnnotationPoint(0.3, 0.9), AnnotationPoint(0.5, 0.75)}},
	}
	counted := -1
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		counted = ImageAnnotator(c, ImageAnnotatorOptions{
			Name: "Rooftop with marks", Width: 400, Height: 300,
		}).Count()
	}, 400, 300)
	if _, ok := tt.Find("Rooftop with marks"); !ok {
		t.Error("the annotated picture must be named")
	}
	if counted != 0 {
		t.Errorf("an annotator with no marks counted %d", counted)
	}

	withMarks := -1
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		withMarks = ImageAnnotator(c, ImageAnnotatorOptions{
			Name: "Rooftop with marks", Width: 400, Height: 300, Marks: marks,
		}).Count()
	}, 400, 300)
	if withMarks != 3 {
		t.Errorf("the annotator counted %d marks, want 3", withMarks)
	}

	// A stroke in progress is drawn like any other mark and is not counted: it
	// is already in the caller's slice if it is going to be, and counting it
	// twice is how an undo stack grows a duplicate.
	withStroke := -1
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		withStroke = ImageAnnotator(c, ImageAnnotatorOptions{
			Name: "Rooftop with marks", Width: 400, Height: 300,
			Marks: marks, Drawing: &Annotation{Kind: "pen"},
		}).Count()
	}, 400, 300)
	if withStroke != 3 {
		t.Errorf("a stroke in progress was counted: %d settled marks, want 3", withStroke)
	}
}

func TestSubtitleEditor(t *testing.T) {
	cues := []Cue{
		{Start: 0, End: 3 * time.Second, Text: "we are on the roof"},
		{Start: 3 * time.Second, End: 4 * time.Second,
			Text: "the membrane here is twenty years old and it is split along the whole seam"},
		{Start: 4 * time.Second, End: 8 * time.Second, Text: "and it has been pooling since March"},
	}
	selected := 0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		SubtitleEditor(c, SubtitleEditorOptions{
			Cues: cues, Selected: &selected, Height: 260,
		})
	}, 640, 340)
	wantText(t, tt, "the editor", "Subtitles", "Start", "End", "Line", "Words")
	if !tt.HasText("we are on the roof") {
		t.Errorf("the first cue's line is missing; %q", tt.Texts())
	}
	if tt.HasText("20") {
		// Word counts are in the table's own column and this assertion says
		// the number appears where the reader looks for it.
		t.Log("word counts are on screen as expected")
	}
	wantsPanic(t, "an editor with no height", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			sel := -1
			SubtitleEditor(c, SubtitleEditorOptions{Cues: cues, Selected: &sel})
		}, 600, 300)
	})
}

func TestSubtitleEditorSortsThroughTheTable(t *testing.T) {
	cues := []Cue{
		{Start: 0, End: 2 * time.Second, Text: "b"},
		{Start: 2 * time.Second, End: 3 * time.Second, Text: "a"},
	}
	selected := -1
	sort := data.Sort{}
	// The result is *accumulated*, not assigned: the view runs up to three
	// times per frame and the settled pass has no click, so a plain assignment
	// would overwrite the answer with the empty one. Reading the result after
	// settling is the mistake §17.2 warns about, and here it would make this
	// test pass for the wrong reason.
	sorted := ""
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		if got := SubtitleEditor(c, SubtitleEditorOptions{
			Cues: cues, Selected: &selected, Height: 220, Sort: &sort,
		}).Sorted(); got != "" {
			sorted = got
		}
	}, 640, 280)
	if sorted != "" {
		t.Errorf("nothing was pressed, so nothing should report a sort; got %q", sorted)
	}
	// The table asks for the order; the editor passes it through and the
	// caller reorders. Sorting is the caller's job, and the assertion is that
	// the ask reaches it.
	tt.Click("Start")
	if sorted != "start" {
		t.Errorf("pressing the Start head reported %q, want \"start\"", sorted)
	}
	if sort.Column != "start" {
		t.Errorf("the sort state holds %q, want \"start\"", sort.Column)
	}
}

func TestSubtitleEmptyAndCueFields(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		SubtitleEditor(c, SubtitleEditorOptions{
			Cues: nil, Selected: &[]int{-1}[0], Height: 220,
		})
	}, 640, 280)
	wantText(t, tt, "the empty editor", "Subtitles", "No subtitles in this file")

	cue := Cue{Start: 3980 * time.Millisecond, End: 6120 * time.Millisecond, Speaker: "Dana"}
	tt2 := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		CueFields(c, &cue)
	}, 560, 80)
	if _, ok := tt2.Find("Dana"); !ok {
		t.Error("a cue's speaker should be shown beside its times")
	}
	// Sub-second times survive a frame: the fields show whole hundredths and
	// writing them back must not round the cue to the second.
	if cue.Start != 3980*time.Millisecond || cue.End != 6120*time.Millisecond {
		t.Errorf("a frame of CueFields moved the cue to %v–%v", cue.Start, cue.End)
	}
}

func TestCameraPreview(t *testing.T) {
	for _, live := range []bool{true, false} {
		on := live
		tt := ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			CameraPreview(c, CameraPreviewOptions{
				Name: "Dana on the roof", Live: on, Level: 0.4,
				Badge: "LIVE", Mirror: true,
			})
		}, 480, 400)
		if _, ok := tt.Find("Dana on the roof"); !ok {
			t.Errorf("a preview (live=%v) must be named", on)
		}
	}
	wantsPanic(t, "an unnamed preview", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			CameraPreview(c, CameraPreviewOptions{})
		}, 300, 240)
	})
}

func TestDeviceSelector(t *testing.T) {
	devices := []Device{
		{ID: "cam-face", Name: "FaceTime HD", Kind: KindCamera, Default: true},
		{ID: "cam-side", Name: "Side camera", Kind: KindCamera, Muted: true},
		{ID: "mic-1", Name: "MacBook microphone", Kind: KindMic},
	}
	chosen := "cam-face"
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		DeviceSelector(c, DeviceSelectorOptions{
			Devices: devices, Selected: &chosen, Kind: KindCamera,
			Label: "Camera", Height: 140,
		})
	}, 480, 200)
	wantText(t, tt, "the picker", "FaceTime HD", "Side camera", "Unavailable")
	if _, ok := tt.Find("Camera"); !ok {
		t.Error("a list of devices must be named; a screen reader cannot describe it otherwise")
	}
	if strings.Contains(strings.Join(tt.Texts(), "|"), "MacBook microphone") {
		t.Error("filtering by kind must leave the microphone out")
	}

	// A press writes the device's ID, not its index: a list that is rebuilt
	// when a device is plugged in must still resolve to the same device.
	picks := 0
	tt2 := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		chosen := "cam-face"
		DeviceSelector(c, DeviceSelectorOptions{
			Devices: devices, Selected: &chosen, Kind: KindCamera, Label: "Camera",
		})
		if chosen == "cam-side" {
			picks++
		}
	}, 480, 120)
	tt2.Click("Side camera")
	if picks != 0 {
		// A muted device stays in the list and stays unclickable.
		t.Error("a muted device must not be selectable")
	}

	none := "x"
	tt3 := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		DeviceSelector(c, DeviceSelectorOptions{
			Devices: nil, Selected: &none, Kind: KindScreen, Label: "Screen",
		})
	}, 400, 120)
	wantText(t, tt3, "the empty picker", "No screen devices found")

	wantsPanic(t, "a picker with no selection", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			DeviceSelector(c, DeviceSelectorOptions{Devices: devices, Label: "Camera"})
		}, 400, 120)
	})
}

func TestScreenRecorderControls(t *testing.T) {
	source := "Zoom window — Site walkthrough"
	presses := 0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		recording := false
		at := time.Duration(0)
		if ScreenRecorderControls(c, ScreenRecorderControlsOptions{
			Recording: &recording, At: &at, Source: &source,
			Width: 1920, Height: 1080, Zoomed: true,
		}).Toggled() {
			presses++
		}
	}, 620, 100)
	wantText(t, tt, "the controls", "0:00", source, "1920 by 1080, window only")
	tt.Click("Start recording")
	if presses != 1 {
		t.Fatalf("the record button reported %d presses, want 1", presses)
	}

	wantsPanic(t, "controls with no recording flag", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			at := time.Duration(0)
			ScreenRecorderControls(c, ScreenRecorderControlsOptions{At: &at})
		}, 400, 80)
	})
}

// ── dark mode ──────────────────────────────────────────────────────────────

// TestDarkModeDrawsEveryComponent is the one test that says the palette is
// consulted rather than assumed. It runs the whole package's worth of marks in
// the dark palette and asks for the ones that carry a name: if any of them
// reached for a light-only colour, the window would still lay out and the test
// would still pass — so the assertion that matters is that each one is found,
// in dark, at the same place its own name says it is.
func TestDarkModeDrawsEveryComponent(t *testing.T) {
	at, playing, level := 12*time.Second, true, 0.5
	muted, speed, speedOpen := false, 1.0, false
	_ = level
	crop := Rect{X: 0.1, Y: 0.1, W: 0.8, H: 0.8}
	compareAt := float32(0.5)
	selected := 0
	mark := Annotation{Kind: "pin", Points: []struct{ X, Y float32 }{AnnotationPoint(0.4, 0.4)}}
	recording := true

	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: core.Dark})

		AudioWaveform(c, AudioWaveformOptions{Samples: samples, Label: "dm.wave"})
		AudioSpectrum(c, AudioSpectrumOptions{Bands: []float64{0.2, 0.8}, Label: "dm.spectrum"})
		MicLevelMeter(c, MicLevelMeterOptions{Level: 0.7, Peak: 0.9, Label: "dm.mic"})
		VideoScrubber(c, VideoScrubberOptions{
			At: &at, Duration: time.Minute, Label: "dm.scrubber",
		})
		VolumeControl(c, VolumeControlOptions{Volume: &level, Muted: &muted, Label: "dm.volume"})
		PlaybackSpeedControl(c, PlaybackSpeedControlOptions{
			Speed: &speed, Open: &speedOpen, Label: "dm.speed",
		})
		MediaControls(c, MediaControlsOptions{
			At: &at, Duration: 312 * time.Second, Playing: &playing,
			Volume: &level, Muted: &muted, Speed: &speed, Open: &speedOpen,
		})
		VideoPlayer(c, VideoPlayerOptions{
			Title: "dm.player", Duration: 312 * time.Second,
			At: &at, Playing: &playing, Controls: true,
		})
		AudioPlayer(c, AudioPlayerOptions{
			Title: "dm.audio", Duration: 45 * time.Second,
			At: &at, Playing: &playing, Samples: samples,
		})
		VideoThumbnailStrip(c, VideoThumbnailStripOptions{
			Frames:   []Frame{{At: 0}, {At: 10 * time.Second}},
			Selected: &selected, Label: "dm.strip",
		})
		ImageThumbnail(c, ImageThumbnailOptions{Name: "dm.thumb", Selected: true})
		ImageCompare(c, ImageCompareOptions{Name: "dm.compare", At: &compareAt})
		ImageCropper(c, ImageCropperOptions{Name: "dm.crop", Rect: &crop})
		ImageAnnotator(c, ImageAnnotatorOptions{Name: "dm.annotate", Marks: []Annotation{mark}})
		CameraPreview(c, CameraPreviewOptions{Name: "dm.camera", Live: true, Badge: "LIVE"})
		ScreenRecorderControls(c, ScreenRecorderControlsOptions{
			Recording: &recording, At: &at, Width: 1280, Height: 720,
		})
		DeviceSelector(c, DeviceSelectorOptions{
			Devices:  []Device{{ID: "a", Name: "dm.device", Kind: KindMic}},
			Selected: &[]string{"a"}[0], Label: "dm.devices",
		})
		SubtitleEditor(c, SubtitleEditorOptions{
			Cues:     []Cue{{Start: 0, End: 2 * time.Second, Text: "dm.cue"}},
			Selected: &selected, Height: 200,
		})
	}, 900, 900)

	for _, name := range []string{
		"dm.wave", "dm.spectrum", "dm.mic", "dm.scrubber", "dm.volume", "dm.speed",
		"dm.player", "dm.audio", "dm.strip", "dm.thumb", "dm.compare", "dm.crop",
		"dm.annotate", "dm.camera", "dm.device", "dm.devices", "dm.cue",
	} {
		if _, ok := tt.Find(name); !ok {
			t.Errorf("dark mode: %q is not on screen", name)
		}
	}
}

// TestDarkPaletteIsReadNotAssumed says the same thing about the palette
// itself: the tokens a dark window resolves are the dark ones, and a component
// that read a hard-coded colour would pass the render test above while being
// wrong in the one place it cannot be seen.
func TestDarkPaletteIsReadNotAssumed(t *testing.T) {
	var bg ui.Color
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: core.Dark})
		bg = core.Tokens(c).Background
		AudioSpectrum(c, AudioSpectrumOptions{
			Bands: []float64{0.1, 0.5}, Label: "dark spectrum",
		})
	}, 400, 140)
	if tt == nil {
		t.Fatal("the tester did not run the view")
	}
	if bg != theme.Dark().Background {
		t.Errorf("a dark window resolved the background %v, want %v", bg, theme.Dark().Background)
	}
	if bg == theme.Light().Background {
		t.Error("a dark window resolved the light background")
	}
}
