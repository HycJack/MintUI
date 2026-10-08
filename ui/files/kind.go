package files

import (
	"strings"

	"github.com/HycJack/MintUI/ui/display"
)

// Kind is what a file is, as far as this package is concerned — which is to
// say, as far as the decision "can this be shown, and with what glyph" goes.
//
// It is four kinds rather than a hundred because a hundred is a table of
// extensions that goes stale the week somebody adds a format, and the two
// things that depend on it — the glyph and the preview — do not actually
// differ between, say, a .tar.gz and a .zip.
type Kind int

const (
	// Binary is anything that is not text and not something with a picture
	// in it. It is the zero value and the default, because a file nobody
	// recognises is far more often binary than not, and showing bytes as
	// text is the worse mistake of the two.
	Binary Kind = iota
	// Text is something that can be shown in an editor.
	Text
	// Image is something with a picture in it.
	Image
	// Archive is a bundle of other files, which has contents of its own to
	// show rather than a preview.
	Archive
	// Media is audio or video: something with a duration and a transport
	// rather than a picture or a page of words.
	Media
)

func (k Kind) String() string {
	switch k {
	case Text:
		return "Text"
	case Image:
		return "Image"
	case Archive:
		return "Archive"
	case Media:
		return "Media"
	}
	return "Binary"
}

// CanPreview reports whether anything can be shown of a file of this kind,
// which is the question FilePreview asks before it decides whether to draw
// a view or an apology.
func (k Kind) CanPreview() bool { return k == Text || k == Image }

// extKinds is the extension-to-kind table. It is a map rather than a switch
// because the names are data: a caller reading it should be able to see what
// it claims without reading a hundred lines of control flow.
//
// The keys are lower-case and without the dot. An extension nobody has heard
// of is not in here, and lands on Binary — which is the safe direction to
// fail in, because a preview of bytes is a screenful of nothing and a
// preview of a picture that was not one is a screenful of nonsense.
var extKinds = map[string]Kind{
	// Text.
	"txt": Text, "md": Text, "mdx": Text, "markdown": Text, "rst": Text,
	"log": Text, "csv": Text, "tsv": Text, "json": Text, "jsonl": Text,
	"yaml": Text, "yml": Text, "toml": Text, "xml": Text, "html": Text,
	"htm": Text, "css": Text, "scss": Text, "less": Text, "ini": Text,
	"conf": Text, "cfg": Text, "env": Text, "properties": Text,
	"js": Text, "jsx": Text, "mjs": Text, "cjs": Text, "ts": Text, "tsx": Text,
	"go": Text, "py": Text, "rb": Text, "rs": Text, "java": Text, "kt": Text,
	"kts": Text, "c": Text, "h": Text, "cc": Text, "cpp": Text, "hpp": Text,
	"m": Text, "mm": Text, "swift": Text, "cs": Text, "php": Text, "pl": Text,
	"sh": Text, "bash": Text, "zsh": Text, "fish": Text, "ps1": Text,
	"sql": Text, "graphql": Text, "gql": Text, "proto": Text, "tf": Text,
	"dockerfile": Text, "gitignore": Text, "makefile": Text, "lock": Text,
	"diff": Text, "patch": Text, "tex": Text, "bib": Text,

	// Image.
	"png": Image, "jpg": Image, "jpeg": Image, "gif": Image, "webp": Image,
	"avif": Image, "bmp": Image, "tif": Image, "tiff": Image, "svg": Image,
	"ico": Image, "heic": Image, "heif": Image, "raw": Image,

	// Archive.
	"zip": Archive, "tar": Archive, "gz": Archive, "tgz": Archive,
	"bz2": Archive, "xz": Archive, "zst": Archive, "7z": Archive,
	"rar": Archive, "jar": Archive, "war": Archive, "apk": Archive,
	"whl": Archive, "dmg": Archive, "iso": Archive,

	// Media.
	"mp3": Media, "wav": Media, "flac": Media, "ogg": Media, "oga": Media,
	"m4a": Media, "aac": Media, "aiff": Media, "opus": Media,
	"mp4": Media, "mov": Media, "avi": Media, "mkv": Media, "webm": Media,
	"m4v": Media, "mpg": Media, "mpeg": Media, "wmv": Media,
}

// extIcons is which of the library's thirty-five glyphs a kind wears. Every
// one of them is the built-in set rather than a drawn one, for the reason
// display.Icon exists: a file type with its own hand-drawn glyph is the
// moment a file manager stops matching the rest of the interface.
//
// The mapping is by shape rather than by subject, because the set is small
// and has no picture of a photograph in it:
//
//   - a grid of four squares is a gallery, which is what a picture is
//   - a cylinder is something stored, which is what an archive is
//   - a document with a tick on it is a file with words in it
//   - a slider panel is a mixing desk, which is what audio is
//   - a bordered panel is a frame, which is what a video is
//   - and the fallback is the cylinder again, because an unknown file is a
//     thing whose contents nobody has looked at
var extIcons = map[Kind]display.IconName{
	Text:    display.IconJobs,
	Image:   display.IconOverview,
	Archive: display.IconInventory,
	Media:   display.IconPanel,
	Binary:  display.IconInventory,
}

// fallbackKind is what an extension nobody has heard of is. Binary, so that
// a preview of bytes is refused rather than attempted.
const fallbackKind = Binary

// Ext is a path's extension, lower-cased and without the dot: "gz" for
// "archive.tar.gz".
//
// The last dot in the name counts, not the first, so a dotfile called
// ".gitignore" has no extension — which is right, because ".gitignore" is
// the whole name. The exception is a name that is nothing but an extension,
// like ".env", which is a name people write and mean.
func Ext(path string) string {
	name := Base(path)
	dot := strings.LastIndexByte(name, '.')
	if dot <= 0 {
		return ""
	}
	return strings.ToLower(name[dot+1:])
}

// Base is a path's last segment: "main.go" for "/repo/ui/main.go", and
// "main.go" for "C:\\repo\\main.go" as well, because a path is compared
// against strings here and opened by something that speaks forward slashes.
func Base(path string) string {
	p := strings.ReplaceAll(path, `\`, "/")
	p = strings.TrimRight(p, "/")
	if i := strings.LastIndexByte(p, '/'); i >= 0 {
		return p[i+1:]
	}
	return p
}

// Dir is a path's parent, in the same normalised form Join produces: "." for
// a path with no directory in it, because that is where a bare name lives.
//
// It is the one place a file tree shows a name on its own — the root's own
// row — and showing "/" there would be a path that means the whole disk.
func Dir(path string) string {
	p := strings.ReplaceAll(path, `\`, "/")
	p = strings.TrimRight(p, "/")
	i := strings.LastIndexByte(p, '/')
	switch {
	case i < 0:
		return "."
	case i == 0:
		return "/"
	}
	return p[:i]
}

// KindOf is what a file is, from its extension.
//
// An extension nobody has heard of is Binary, not Text. That is the
// conservative direction and it is the one that matters: a preview that
// refuses is a sentence, and a preview of a binary file as text is a screen
// full of nothing that somebody has to read to find out it is nothing.
func KindOf(name string) Kind {
	if k, ok := extKinds[Ext(name)]; ok {
		return k
	}
	return fallbackKind
}

// IsDir reports a name being a directory rather than a file. It takes the
// name rather than the path because a caller that knows is looking at a
// directory has one entry and not two lists, and a trailing separator is the
// only thing that distinguishes them — which is also why Join drops the
// empty segment a trailing separator makes.
func IsDir(name string) bool {
	return name == "/" || strings.HasSuffix(name, "/")
}
