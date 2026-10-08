package pages

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/data"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/media"
	"github.com/HycJack/MintUI/ui/showcase"
	"github.com/HycJack/MintUI/ui/theme"
)

// Demo state outlives the frame: every demo below hands its component
// a pointer, and a pointer into a frame-local is a click the next
// frame undoes — a tab that will not switch, a dropdown that snaps
// shut, a slider that springs back.
var (
	videoAt        = 61 * time.Second
	videoPlaying   = false
	videoSpeedOpen = false
	audioAt        = 9 * time.Second
	audioPlaying   = true
	audioSpeedOpen = false
	barAt          = 12 * time.Second
	barPlaying     = true
	compactAt      = 200 * time.Second
	compactPlaying = false
	plainAt        = 1 * time.Minute
	bufferedAt     = 2 * time.Minute
	offAt          = 0 * time.Second
	speedOpen      = true
	speed          = 1.25
	loose          = 1.0
	cam            = "cam-studio"
	media_none     = ""
	source         = "Riverside Clinic — 客服窗口"
	vAt            = float32(0.55)
	crop           = media.Rect{X: 0.12, Y: 0.18, W: 0.6, H: 0.6}
	square         = media.Rect{X: 0.25, Y: 0.1, W: 0.5, H: 0.8}
	free           = media.Rect{X: 0.2, Y: 0.2, W: 0.5, H: 0.5}
	edit           = mediaCues[1]
	untimed        = media.Cue{Text: "…"}
	media_open     = true

	mediaCues = []media.Cue{
		{Start: 0, End: 2 * time.Second, Text: "Good morning — did you sleep okay?"},
		{Start: 2 * time.Second, End: 3 * time.Second, Text: "Let me check the compressor again."},
		{Start: 3 * time.Second, End: 4 * time.Second, Speaker: "Andre",
			Text: "Yes. It kicks in around four in the afternoon."},
		{Start: 4 * time.Second, End: 7 * time.Second, Text: "I'll bring the panel up on Monday."},
	}
)

// The page is one video and the things that sit around it: two players, the
// waveform, the scrubber, the meters, the device picker, the image editors,
// the subtitle table and the lightbox.
//
// Everything here is handed in. This package's whole contract is that a frame
// of video, a buffer of samples and a list of cameras arrive as ordinary Go
// values, so nothing on this page opens a file, a socket or a device: the
// pictures are made in memory and the sample buffers are arithmetic. That is
// also why the page renders at all — a media gallery that needed a camera to
// be screenshot is a media gallery nobody has ever looked at.

func init() {
	showcase.Register(showcase.Page{
		Package: "media",
		Title:   "ui/media — 画面与声音",
		Note:    "播放器、进度、波形、频谱、电平、设备、图片编辑、字幕",
		Width:   1000,
		Height:  4600,
		Want: []string{
			// players
			"门诊回放 2024-11-08", "5:12", "1:01", "Voice note from Andre",
			"Audio 0:42",
			// sound
			"Waveform of the intake call", "Waveform of the whole clip, played to the end",
			"Waveform with nothing loaded into it",
			"Band levels of the intake call, log axis",
			"Band levels of the intake call, linear axis",
			"Mic level, quiet", "Mic level, hot", "Mic level, clipping", "Mic level, muted",
			// progress
			"Playback position", "Playback position, buffered",
			"Playback position, disabled", "Chapters", "Frame at 1:12",
			// volume and speed
			"Volume control", "Playback speed", "1.25×",
			// devices
			"Cameras", "Microphones", "No screen devices found",
			"Studio camera", "Built-in microphone", "Unavailable", "Default",
			"Waiting for a frame…", "Camera off", "Live",
			"Screen recording", "Stop recording", "Start recording",
			"Recorded area 2560 by 1440, window only",
			// images
			"Riverside intake 01", "Receipt scan", "2 annotations",
			"Site photo, before and after", "Site photo, vertical",
			"Site photo being cropped", "Screenshot 2024-11-08",
			// subtitles
			"Subtitles", "Start", "End", "Words",
			"Let me check the compressor again.", "Cue times", "No subtitles",
			// lightbox
			"Lightbox over a window", "The picture the lightbox enlarges",
			// the pure functions
			"ToDecibels(0.5)", "-6.02", "DecibelNorm(0 dB)", "TimeToText(3661s)",
			"1:01:01", "ScrubRatio(1:01 / 5:12)", "Unit()", "PeakToPeak([0.9 0.9])",
			"Peaks(x, 8)", "ReadTime(12 words, 160/min)", "AnnotationPoint(0.5, 0.25)",
			"0.75×", "1.5×",
		},
		Render: func(c *ui.Context) {
			mediaPage(c)
		},
	})
}

func mediaPage(c *ui.Context) {
	mediaPlayersSection(c)
	mediaSoundSection(c)
	mediaScrubSection(c)
	mediaControlSection(c)
	mediaDeviceSection(c)
	mediaImageSection(c)
	mediaSubtitleSection(c)
	mediaLightboxSection(c)
	mediaPureSection(c)
}

// ── players ────────────────────────────────────────────────────────────────

func mediaPlayersSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "播放 · VideoPlayer / AudioPlayer / MediaControls")

	videoVol := showcase.State(c, "media.133.videoVol", 0.8)
	videoMuted := showcase.State(c, "media.133.videoMuted", false)
	videoSpeed := showcase.State(c, "media.133.videoSpeed", 1.0)
	audioVol := showcase.State(c, "media.134.audioVol", 0.6)
	audioMuted := showcase.State(c, "media.134.audioMuted", false)
	audioSpeed := showcase.State(c, "media.134.audioSpeed", 1.0)

	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 1.5)).Children(func() {
			showcase.Field(c, "VideoPlayer — 画面、标题、传输条都是调用方给的")
			media.VideoPlayer(c, media.VideoPlayerOptions{
				Title: "门诊回放 2024-11-08", Duration: 312 * time.Second,
				At: &videoAt, Playing: &videoPlaying,
				Volume: videoVol, Muted: videoMuted,
				Speed: videoSpeed, SpeedOpen: &videoSpeedOpen,
				Picture:  mediaPicture(480, 270, 2),
				Buffered: 0.62,
				Markers:  []float32{0.18, 0.47, 0.81},
				Controls: true,
			})
		})

		ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 1.5)).Children(func() {
			showcase.Field(c, "AudioPlayer — 波形就是拖动条，播放头写回 At")
			media.AudioPlayer(c, media.AudioPlayerOptions{
				Title: "Voice note from Andre", Duration: 42 * time.Second,
				At: &audioAt, Playing: &audioPlaying,
				Samples: mediaSamples(240),
				Volume:  audioVol, Muted: audioMuted,
				Speed: audioSpeed, SpeedOpen: &audioSpeedOpen,
			})
			// A player with nothing decoded into it draws its own surface, and
			// that is worth seeing beside one with something in it: it is the
			// state every player is in for the second after it opens.
			showcase.Field(c, "AudioPlayer with no Title and no samples")
			quietAt := showcase.State(c, "media.164.quietAt", time.Duration(0))
			quietPlaying := showcase.State(c, "media.164.quietPlaying", false)
			media.AudioPlayer(c, media.AudioPlayerOptions{
				Duration: 42 * time.Second, At: quietAt, Playing: quietPlaying,
				Compact: true,
			})
		})
	})

	showcase.Field(c, "MediaControls — 单独的传输条：完整的那一条，和紧凑的那一条")
	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		barVol := showcase.State(c, "media.174.barVol", 0.35)
		barMuted := showcase.State(c, "media.174.barMuted", false)
		barSpeed := showcase.State(c, "media.175.barSpeed", 1.0)
		barOpen := showcase.State(c, "media.175.barOpen", false)
		ui.Column(c).WidthPercent(66).Shrink(0).Gap(unit(c, 1)).Children(func() {
			media.MediaControls(c, media.MediaControlsOptions{
				At: &barAt, Duration: 312 * time.Second, Playing: &barPlaying,
				Volume: barVol, Muted: barMuted,
				Speed: barSpeed, Open: barOpen,
				Buffered: 0.4, Markers: []float32{0.3, 0.66},
			})
			ui.Text(c, "0:12 已播 · 5:12 总长 · 缓冲 40% · 两处章节").TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize))
		})
		ui.Column(c).WidthPercent(32).Shrink(0).Gap(unit(c, 1)).Children(func() {
			media.MediaControls(c, media.MediaControlsOptions{
				At: &compactAt, Duration: 312 * time.Second,
				Playing: &compactPlaying, Compact: true,
			})
			ui.Text(c, "Compact：没有音量轨，也没有倍速").TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize))
		})
	})
}

// ── sound ──────────────────────────────────────────────────────────────────

func mediaSoundSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "声音 · AudioWaveform / AudioSpectrum / MicLevelMeter / Levels")

	samples := mediaSamples(320)
	bands := mediaBands(24)

	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "AudioWaveform — 播放头后面的条用 SurfaceHover，不用强调色")
			media.AudioWaveform(c, media.AudioWaveformOptions{
				Samples: samples, Bars: 64, Position: 0.42, Height: unit(c, 6),
				Label: "Waveform of the intake call",
			})
			media.AudioWaveform(c, media.AudioWaveformOptions{
				Samples: samples, Bars: 64, Position: 1, Height: unit(c, 4),
				Label: "Waveform of the whole clip, played to the end",
			})
			media.AudioWaveform(c, media.AudioWaveformOptions{
				Height: unit(c, 4), Label: "Waveform with nothing loaded into it",
			})
			ui.Text(c, "上面三条：0.42 / 1.0 / 什么都没有").TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize))
		})

		ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "AudioSpectrum — LogAxis 把频段摊开，高频才看得见")
			media.AudioSpectrum(c, media.AudioSpectrumOptions{
				Bands: bands, Height: unit(c, 6), LogAxis: true,
				Label: "Band levels of the intake call, log axis",
			})
			media.AudioSpectrum(c, media.AudioSpectrumOptions{
				Bands: bands, Height: unit(c, 6),
				Label: "Band levels of the intake call, linear axis",
			})
			showcase.Field(c, "MicLevelMeter — 分段的，最上面两格永远是告警色")
			mic := func(level float64, muted bool, label string) {
				ui.Column(c).FillWidth().Gap(unit(c, 0.5)).Children(func() {
					ui.Text(c, label).TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.CaptionSize))
					// Built here rather than beside it, because an element
					// belongs to whatever column is current where it is made.
					media.MicLevelMeter(c, media.MicLevelMeterOptions{
						Level: level, Peak: math.Max(level, 0.62), Muted: muted,
						Label: label,
					})
				})
			}
			mic(0.18, false, "Mic level, quiet")
			mic(0.72, false, "Mic level, hot")
			mic(0.99, false, "Mic level, clipping")
			mic(0.72, true, "Mic level, muted")
		})
	})
}

// ── progress ───────────────────────────────────────────────────────────────

func mediaScrubSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "进度 · VideoScrubber / VideoThumbnailStrip")

	ui.Column(c).FillWidth().Gap(unit(c, 2)).Children(func() {
		showcase.Field(c, "VideoScrubber — 轨道、缓冲带、章节标记、拇指")
		scrub := func(label, note string, at *time.Duration, buffered float32,
			markers []float32, disabled bool) {
			ui.Row(c).FillWidth().Gap(unit(c, 2)).AlignItems(ui.Center).Children(func() {
				ui.Text(c, note).TextColor(k.TextFaint).
					FontSize(core.FontSize(c, theme.CaptionSize)).
					Width(150).Shrink(0)
				media.VideoScrubber(c, media.VideoScrubberOptions{
					At: at, Duration: 312 * time.Second, Buffered: buffered,
					Markers: markers, Label: label, Disabled: disabled,
				})
			})
		}
		scrub("Playback position", "刚播到 1:01", &plainAt, 0, nil, false)
		scrub("Playback position, buffered", "缓冲 72%，两处章节", &bufferedAt, 0.72,
			[]float32{0.3, 0.66}, false)
		scrub("Playback position, disabled", "还没有录到东西", &offAt, 0, nil, true)

		showcase.Field(c, "VideoThumbnailStrip — 按下只报帧号，不自己挪播放头")
		selected := 2
		frames := make([]media.Frame, 0, 5)
		for i := range 5 {
			at := time.Duration(i) * 72 * time.Second
			frames = append(frames, media.Frame{
				At: at, Bitmap: mediaPicture(160, 90, i),
			})
		}
		media.VideoThumbnailStrip(c, media.VideoThumbnailStripOptions{
			Frames: frames, Selected: &selected,
			Duration: 312 * time.Second, Height: unit(c, 9), Label: "Chapters",
		})
		ui.Text(c, "0:00 · 1:12 · 2:24 · 3:36 · 4:48，播放器时长 5:12").TextColor(k.TextFaint).
			FontSize(core.FontSize(c, theme.CaptionSize))
	})
}

// ── volume and speed ───────────────────────────────────────────────────────

func mediaControlSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "音量与倍速 · VolumeControl / PlaybackSpeedControl / SpeedChoices")

	// The rate menu is an anchored layer: it is drawn outside the row and lands
	// under its button, so the row of stages gives it room the same way a
	// popover's stage does on the overlay page.
	showcase.Field(c, "VolumeControl — 按钮是静音，不是音量：静音后音量还在")
	row := func(level float64, muted bool, note string) {
		vol := showcase.State(c, "media.volume."+note, level)
		mutedFlag := showcase.State(c, "media.muted."+note, muted)
		ui.Column(c).Grow(1).Shrink(0).Gap(unit(c, 0.5)).Children(func() {
			media.VolumeControl(c, media.VolumeControlOptions{
				Volume: vol, Muted: mutedFlag, Label: "Volume",
			})
			ui.Text(c, note).TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize))
		})
	}
	ui.Row(c).FillWidth().Gap(unit(c, 3)).Children(func() {
		row(0, false, "0% · 只剩按钮")
		row(0.35, false, "35% · volumeDown 的记号")
		row(0.9, false, "90% · volume 的记号")
		row(0.9, true, "静音 · 记号换回 mute")
	})

	showcase.Field(c, "PlaybackSpeedControl — 菜单里的六个倍速，1.25× 现在是选中的")
	ui.Box(c).FillWidth().Height(unit(c, 34)).Radius(theme.ControlRadius).
		Background(k.Background).Border(theme.BorderWidth, k.Border).Clip().Children(func() {
		ui.Box(c).FillWidth().Padding(unit(c, 1.5), unit(c, 2)).Children(func() {
			ui.Text(c, "菜单挂在按钮下面，浮出时向上翻").TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize))
		})
		media.PlaybackSpeedControl(c, media.PlaybackSpeedControlOptions{
			Speed: &speed, Open: &speedOpen, Label: "Playback speed",
		})
	})
	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		showcase.Field(c, "SpeedChoices — 单独画一次，因为它自己也导出")
		media.SpeedChoices(c, &loose, media.Speeds, "Playback speed")
	})
}

// ── devices ────────────────────────────────────────────────────────────────

// mediaDevices is the enumeration a caller would get from the operating
// system. It is written out here rather than asked for, because the whole
// point of this package is that nothing in it opens a device: a fake list is
// what makes the picker testable at all.
func mediaDevices() []media.Device {
	return []media.Device{
		{ID: "cam-studio", Name: "Studio camera", Kind: media.KindCamera, Default: true},
		{ID: "cam-face", Name: "FaceTime HD", Kind: media.KindCamera},
		{ID: "cam-capture", Name: "OBS Virtual Camera", Kind: media.KindCamera, Muted: true},
		{ID: "mic-builtin", Name: "Built-in microphone", Kind: media.KindMic, Default: true},
		{ID: "mic-screen", Name: "Screen recording audio", Kind: media.KindMic, Muted: true},
		{ID: "mic-dante", Name: "Dante USB", Kind: media.KindMic},
		{ID: "out-builtin", Name: "Built-in output", Kind: media.KindSpeaker, Default: true},
	}
}

func mediaDeviceSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "设备 · DeviceSelector / CameraPreview / ScreenRecorderControls")

	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).WidthPercent(share(3)).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "DeviceSelector — 写回 ID 而不是序号，列表变了也还是同一台")
			media.DeviceSelector(c, media.DeviceSelectorOptions{
				Devices: mediaDevices(), Selected: &cam,
				Kind: media.KindCamera, Label: "Cameras", Height: unit(c, 36),
			})
			mic := "mic-dante"
			media.DeviceSelector(c, media.DeviceSelectorOptions{
				Devices: mediaDevices(), Selected: &mic,
				Kind: media.KindMic, Label: "Microphones", Height: unit(c, 36),
			})
			media.DeviceSelector(c, media.DeviceSelectorOptions{
				Devices: mediaDevices(), Selected: &media_none,
				Kind: media.KindScreen, Label: "Screens",
			})
		})

		ui.Column(c).WidthPercent(share(3)).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "CameraPreview — 有帧、等待中、关掉，三种说法不同")
			media.CameraPreview(c, media.CameraPreviewOptions{
				Frame: mediaPicture(320, 240, 1), Name: "Studio camera",
				Live: true, Level: 0.55, Badge: "Live",
			})
			media.CameraPreview(c, media.CameraPreviewOptions{
				Frame: mediaPicture(320, 240, 3), Name: "FaceTime HD",
				Live: true, Level: 0.55, Muted: true, Mirror: true, Badge: "you",
			})
			media.CameraPreview(c, media.CameraPreviewOptions{
				Name: "OBS Virtual Camera", Live: false,
			})
			media.CameraPreview(c, media.CameraPreviewOptions{
				Name: "Studio camera, frame not in yet", Live: true,
			})
		})

		ui.Column(c).WidthPercent(share(3)).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "ScreenRecorderControls — 录制的区域要画出来")
			recording := showcase.State(c, "media.402.recording", true)
			recordingAt := showcase.State(c, "media.402.recordingAt", 74*time.Second)
			media.ScreenRecorderControls(c, media.ScreenRecorderControlsOptions{
				Recording: recording, At: recordingAt, Source: &source,
				Width: 2560, Height: 1440, Zoomed: true,
			})
			// A second frame with the same pointers: the component writes both
			// of them, so asking for the settled state means asking again.
			media.ScreenRecorderControls(c, media.ScreenRecorderControlsOptions{
				Recording: recording, At: recordingAt, Source: &source,
				Width: 2560, Height: 1440,
			})
			idle := showcase.State(c, "media.413.idle", false)
			idleAt := showcase.State(c, "media.413.idleAt", time.Duration(0))
			media.ScreenRecorderControls(c, media.ScreenRecorderControlsOptions{
				Recording: idle, At: idleAt, Source: &source,
				Width: 1920, Height: 1080,
			})
			ui.Text(c, "计时在录制时是危险色；按钮的名字说的是按下会做什么").TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize))
		})
	})
}

// ── images ─────────────────────────────────────────────────────────────────

func mediaImageSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "图片 · ImageThumbnail / ImageCompare / ImageCropper / ImageAnnotator")

	showcase.Field(c, "ImageThumbnail — 角标挂在角上，不改格子高度")
	selected := true
	thumbRow := func(name string, src *ui.Bitmap, badge string, tone core.Severity, sel bool) {
		media.ImageThumbnail(c, media.ImageThumbnailOptions{
			Src: src, Name: name, Side: unit(c, 12), Selected: sel,
			Badge: badge, Tone: tone, Pressable: true,
		})
	}
	ui.Row(c).Gap(unit(c, 1.5)).Children(func() {
		thumbRow("Riverside intake 01", mediaPicture(180, 180, 0), "", core.Neutral, false)
		thumbRow("Riverside intake 02", mediaPicture(180, 180, 1), "", core.Neutral, selected)
		thumbRow("Ultrasound still", mediaPicture(180, 180, 2), "", core.Neutral, false)
		thumbRow("Receipt scan", mediaPicture(180, 180, 3), "2 annotations", core.Warning, false)
		thumbRow("Site photo", nil, "", core.Neutral, false)
		media.ImageThumbnail(c, media.ImageThumbnailOptions{
			Name: "Riverside intake 03", Src: mediaPicture(320, 180, 4),
			Side: unit(c, 12), Ratio: 16.0 / 9.0, Badge: "0:21", Tone: core.Accent,
			Selected: true,
		})
	})

	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "ImageCompare — 两张图共用一个框，中间那条线才是比较")
			at := float32(0.42)
			media.ImageCompare(c, media.ImageCompareOptions{
				Before: mediaPicture(480, 270, 1), After: mediaPicture(480, 270, 2),
				Name: "Site photo, before and after", At: &at,
				Ratio: 16.0 / 9.0, Handle: true,
			})
			showcase.Field(c, "ImageCompare Vertical — 分割线上下移动")
			media.ImageCompare(c, media.ImageCompareOptions{
				Before: mediaPicture(480, 270, 2), After: mediaPicture(480, 270, 3),
				Name: "Site photo, vertical", At: &vAt, Vertical: true,
				Ratio: 16.0 / 9.0,
			})
			ui.Text(c, "at = 0.42 / 0.55").TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize))
		})

		ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "ImageCropper — 区域是比例，不是像素；Shape 比 Ratio 优先")
			media.ImageCropper(c, media.ImageCropperOptions{
				Src: mediaPicture(480, 320, 2), Name: "Site photo being cropped",
				Rect: &crop, Shape: "16:9", Width: 240, Grid: true,
			})
			media.ImageCropper(c, media.ImageCropperOptions{
				Src: mediaPicture(480, 320, 2), Name: "Site photo cropped to a square",
				Rect: &square, Shape: "1:1", Width: 240,
			})
			media.ImageCropper(c, media.ImageCropperOptions{
				Src: mediaPicture(480, 320, 1), Name: "Site photo cropped free",
				Rect: &free, Width: 240,
			})
			showcase.Field(c, "ImageAnnotator — 笔画、图钉、方框，点是比例")
			pen := media.Annotation{Kind: "pen", Points: []struct{ X, Y float32 }{
				media.AnnotationPoint(0.12, 0.62), media.AnnotationPoint(0.26, 0.44),
				media.AnnotationPoint(0.4, 0.6), media.AnnotationPoint(0.55, 0.38),
				media.AnnotationPoint(0.7, 0.52),
			}}
			box := media.Annotation{Kind: "box", Points: []struct{ X, Y float32 }{
				media.AnnotationPoint(0.58, 0.2), media.AnnotationPoint(0.86, 0.52),
			}}
			pin := media.Annotation{Kind: "pin", Points: []struct{ X, Y float32 }{
				media.AnnotationPoint(0.3, 0.3),
			}}
			media.ImageAnnotator(c, media.ImageAnnotatorOptions{
				Src: mediaPicture(480, 270, 3), Name: "Screenshot 2024-11-08",
				Width: 240, Height: 140, Tool: "pen",
				Marks: []media.Annotation{pen, box, pin},
			})
		})
	})
}

// ── subtitles ──────────────────────────────────────────────────────────────

func mediaSubtitleSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "字幕 · SubtitleEditor / CueFields / SubtitleEmpty")

	selected := 1
	sort := &data.Sort{}

	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).WidthPercent(64).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "SubtitleEditor — 时间不够的行，End 用告警色写出来")
			media.SubtitleEditor(c, media.SubtitleEditorOptions{
				Cues: mediaCues, Selected: &selected, Sort: sort, Height: 232,
			})
			showcase.Field(c, "CueFields — 秒是小数，动了就写回同一条")
			media.CueFields(c, &edit)
			media.CueFields(c, &untimed)
		})

		ui.Column(c).WidthPercent(34).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "SubtitleEmpty — 表格的空态，不是空表格")
			ui.Box(c).FillWidth().Radius(theme.CardRadius).
				Background(k.Surface).Border(theme.BorderWidth, k.Border).Children(func() {
				media.SubtitleEmpty(c)
			})
			ui.Text(c, "Start / End / Line / Words 四列，Start、End、Words 可排序").TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize))
		})
	})
}

// ── the lightbox ───────────────────────────────────────────────────────────

// The lightbox is the one component here that owns the whole window: it is an
// overlay.Dialog with a scrim, and a scrim over this page would dim every
// other component on it — which is the opposite of what a gallery is for.
//
// So it is drawn into a window of its own and put on the page as a picture.
// The picture is a real render of the real component, in the same palette; what
// it does not do is reach over the rest of the page, which is the one thing a
// gallery must not let a component do.
func mediaLightboxSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "放大 · Lightbox — 整窗浮层，画在自己的窗口里")

	mode := core.Light
	if core.IsDark(c) {
		mode = core.Dark
	}
	shot := mediaPicture(480, 300, 3)
	// The window the lightbox is drawn into is deliberately bigger than the
	// panel: a preview the same size as the layer would show nothing but the
	// layer, which is the one part of it a gallery can show anywhere.
	frame := ui.NewTester(func(inner *ui.Context) {
		core.Use(inner, core.Settings{Mode: mode})
		// Stand-in for the window the lightbox is opened over: the thing the
		// scrim is there to dim.
		ui.Box(inner).Fill().Background(core.Tokens(inner).Surface).Children(func() {
			ui.Column(inner).Fill().Padding(core.Density(inner).Unit() * 2).
				Gap(core.Density(inner).Unit()).Children(func() {
				ui.Text(inner, "Site photos").TextColor(core.Tokens(inner).Text).
					FontSize(core.FontSize(inner, theme.RowSize)).Bold()
				for _, name := range []string{"Riverside intake 01", "Riverside intake 02"} {
					ui.Text(inner, name).TextColor(core.Tokens(inner).TextMuted).
						FontSize(core.FontSize(inner, theme.MetaSize))
				}
			})
		})
		media.Lightbox(inner, media.LightboxOptions{
			Open: &media_open, Src: shot, Name: "Riverside intake 02",
			Caption: "2024-11-08 · 前面那间", Width: 260,
		})
	}, 460, 340).Image()

	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		display.Image(c, shot, display.ImageOptions{
			Width: 230, Height: 144, Radius: theme.SmallRadius,
			Name: "The picture the lightbox enlarges",
		})
		display.Image(c, ui.NewBitmap(frame), display.ImageOptions{
			Width: 460, Height: 340, Radius: theme.SmallRadius,
			Name: "Lightbox over a window",
		})
		ui.Column(c).Grow(1).Shrink(0).Gap(unit(c, 1)).Children(func() {
			ui.Text(c, "Open 是调用方的 *bool").TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).Bold()
			ui.Text(c, "Src 是调用方解好的位图").TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize))
			ui.Text(c, "Name 必填：一张图自己不带字").TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize))
			ui.Text(c, "底下是 overlay.Dialog：Esc 和点遮罩都关").TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize))
		})
	})
}

// ── the pure functions ─────────────────────────────────────────────────────

func mediaPureSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "纯函数 · 不需要窗口就能算的那几个")

	samples := mediaSamples(64)
	peaks := media.Peaks(samples, 8)

	line := func(call, result string) {
		ui.Row(c).FillWidth().Gap(unit(c, 2)).AlignItems(ui.Center).Children(func() {
			ui.Text(c, call).TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).
				Width(210).Shrink(0)
			ui.Text(c, result).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.RowSize)).Grow(1).SingleLine()
		})
	}
	peakLine := ""
	for i, p := range peaks {
		if i > 0 {
			peakLine += "  "
		}
		peakLine += fmt.Sprintf("%.2f", p)
	}

	ui.Row(c).FillWidth().Gap(unit(c, 4)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 1)).Children(func() {
			line("ToDecibels(0)", fmt.Sprintf("%.2f  = SilenceDB", media.ToDecibels(0)))
			line("ToDecibels(0.5)", fmt.Sprintf("%.2f", media.ToDecibels(0.5)))
			line("ToDecibels(1)", fmt.Sprintf("%.2f", media.ToDecibels(1)))
			line("ToDecibels(2)", fmt.Sprintf("%.2f  ← 超过 1 夹到 0", media.ToDecibels(2)))
			line("DecibelNorm(-90)", fmt.Sprintf("%.2f", float64(media.DecibelNorm(media.SilenceDB))))
			line("DecibelNorm(-6.02)", fmt.Sprintf("%.2f", float64(media.DecibelNorm(-6.02))))
			line("DecibelNorm(0 dB)", fmt.Sprintf("%.2f", float64(media.DecibelNorm(0))))
			line("PeakToPeak([0.9 0.9])",
				fmt.Sprintf("%.2f  ← 全是直流，高度为零", media.PeakToPeak([]float64{0.9, 0.9})))
			line("PeakToPeak([])", fmt.Sprintf("%.2f", media.PeakToPeak(nil)))
		})
		ui.Column(c).WidthPercent(share(2)).Shrink(0).Gap(unit(c, 1)).Children(func() {
			line("Peaks(x, 8)", peakLine)
			line("Levels([0.2 0.9 0.5])", levelsText(media.Levels([]float64{0.2, 0.9, 0.5})))
			line("TimeToText(0)", media.TimeToText(0))
			line("TimeToText(61s)", media.TimeToText(61*time.Second))
			line("TimeToText(3661s)", media.TimeToText(3661*time.Second))
			line("TimeToText(-4s)", media.TimeToText(-4*time.Second)+"  ← 负数读作 0")
			line("ScrubRatio(1:01 / 5:12)",
				fmt.Sprintf("%.3f", media.ScrubRatio(61*time.Second, 312*time.Second)))
			line("ScrubRatio(9:00 / 5:12)",
				fmt.Sprintf("%.3f  ← 超出的夹到 1", media.ScrubRatio(9*time.Minute, 312*time.Second)))
			line("ReadTime(12 words, 160/min)",
				media.TimeToText(media.ReadTime(12, media.ReadingRate)))
			line("Unit()", fmt.Sprintf("X %.2f Y %.2f W %.2f H %.2f",
				media.Unit().X, media.Unit().Y, media.Unit().W, media.Unit().H))
			line("AnnotationPoint(0.5, 0.25)", fmt.Sprintf("X %g Y %g",
				media.AnnotationPoint(0.5, 0.25).X, media.AnnotationPoint(0.5, 0.25).Y))
			line("Speeds", speedsText())
		})
	})
}

func levelsText(levels []float32) string {
	out := ""
	for i, v := range levels {
		if i > 0 {
			out += "  "
		}
		out += fmt.Sprintf("%.2f", v)
	}
	return out
}

func speedsText() string {
	out := ""
	for i, v := range media.Speeds {
		if i > 0 {
			out += "  "
		}
		out += fmt.Sprintf("%g×", v)
	}
	return out
}

// ── the pictures and the samples ───────────────────────────────────────────

// mediaSamples is a buffer of amplitudes made by arithmetic: the point of the
// package is that a waveform is drawn from numbers the caller measured, so the
// gallery's own "recording" has to be numbers too.
func mediaSamples(n int) []float64 {
	out := make([]float64, n)
	for i := range out {
		t := float64(i) / float64(n)
		// A tremolo decaying across the buffer: loud at the start, quiet at
		// the end, with beats a reader can see as a row of bars.
		out[i] = 0.18 + 0.72*math.Abs(math.Sin(t*26))*math.Exp(-t*1.8)
	}
	return out
}

// mediaBands is an analyser's output: one level per frequency band, low to
// high, with the treble quieter than the middle the way most speech is.
func mediaBands(n int) []float64 {
	out := make([]float64, n)
	for i := range out {
		t := float64(i) / float64(n)
		out[i] = 0.9*math.Exp(-3*math.Pow(t-0.18, 2)) +
			0.55*math.Exp(-6*math.Pow(t-0.52, 2)) +
			0.22*math.Exp(-2*math.Pow(t-0.88, 2))
	}
	return out
}

// mediaPicture is a picture made in memory. No file, no network, no camera:
// which is the whole constraint this package is written under, and the reason
// a page like this one can be drawn at all.
func mediaPicture(w, h, seed int) *ui.Bitmap {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	horizon := float64(h) * (0.52 + 0.06*math.Sin(float64(seed)))
	sun := float64(w) * (0.18 + 0.16*float64(seed%6))
	r := float64(min(w, h)) * 0.09
	for y := range h {
		for x := range w {
			fx, fy := float64(x), float64(y)
			var cr, cg, cb uint8
			switch {
			case fy < horizon:
				// Sky, lighter towards the horizon.
				t := fy / horizon
				cr = uint8(196 - 60*t + 8*float64(seed%3))
				cg = uint8(214 - 46*t + 6*float64(seed%2))
				cb = uint8(232 - 30*t)
				if math.Hypot(fx-sun, fy-horizon*0.42) < r {
					cr, cg, cb = 252, 238, 196
				}
			default:
				// Ground, with a band that moves with the seed so two frames
				// of the "same" video are visibly not the same frame.
				t := (fy - horizon) / (float64(h) - horizon)
				shade := 0.82 + 0.12*math.Sin(float64(seed)*1.7+t*4)
				cr = uint8(96 * shade)
				cg = uint8(112 * shade)
				cb = uint8(84 * shade)
			}
			// Two blocks on the horizon: a skyline, and the thing a before /
			// after pair actually differs in.
			if fy > horizon-float64(h)*0.16 &&
				fy < horizon &&
				fx > float64(w)*(0.3+0.18*float64(seed%4)) &&
				fx < float64(w)*(0.3+0.18*float64(seed%4))+float64(w)*0.22 {
				cr, cg, cb = 58, 62, 74
			}
			img.SetRGBA(x, y, color.RGBA{R: cr, G: cg, B: cb, A: 255})
		}
	}
	return ui.NewBitmap(img)
}
