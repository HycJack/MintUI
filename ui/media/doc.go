// Package media is the picture and sound half of the interface: players,
// meters, editors for images and subtitles, and the controls that open a
// device.
//
// # Nothing here touches the machine
//
// No component in this package reads a file, opens a socket or asks an
// operating system for a camera. Every one of them draws what the caller
// hands it, and the things a machine would have to supply — a frame of
// video, a buffer of samples, a list of cameras — arrive as ordinary Go
// values:
//
//	media.VideoPlayer(c, media.VideoPlayerOptions{
//	    Poster:    poster,                     // a *ui.Bitmap the caller loaded
//	    Duration:  312 * time.Second,          // a number, not a probe
//	    At:        &position,                  // the caller's playhead
//	    Thumbs:    thumbs,                     // caller-decoded thumbnails
//	})
//
// This is a deliberate constraint rather than a missing feature. A component
// that can only be exercised by a real camera is a component nobody can test,
// and a media library whose parts all need a device is a media library whose
// bugs are found by users. The wiring — decoding, capturing, playback, device
// enumeration — is the application's job and lives outside this package; what
// is here is the part that can be reasoned about.
//
// # State belongs to the caller
//
// The playhead is a `*time.Duration`, the selected thumbnail an `int`, the
// crop a rectangle, the annotation a list. Nothing is kept between frames by
// a component, so two views of one player cannot disagree about where the
// playhead is, and a component can be drawn twice in one frame.
//
// # The pure functions are the contract
//
// [TimeToText], [ToDecibels], [DecibelNorm] and [Peaks] need no window and no
// state, which is why they are the four functions here that carry the rules a
// caller is most likely to get wrong on its own: what a scrubber reads at
// 1:01:01, what silence maps to in decibels, and how raw samples become bars
// whose heights are all inside [0,1].
package media
