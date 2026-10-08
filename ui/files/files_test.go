package files

import (
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// ── Sanitize ────────────────────────────────────────────────────────────────
//
// Every case here is something somebody has actually typed into a rename
// box, pasted out of a terminal, or sent through an API. They are asserted
// one at a time with the exact string back, because "roughly right" here is
// a file written somewhere nobody can find.

func TestSanitize(t *testing.T) {
	long := strings.Repeat("a", 300)
	longName := strings.Repeat("b", 300) + ".tar.gz"

	cases := []struct {
		name string
		in   string
		want string
	}{
		{"an ordinary name passes through", "report.pdf", "report.pdf"},
		{"a name with spaces is left alone", "Q3 report (final).pdf", "Q3 report (final).pdf"},
		{"nothing at all", "", ""},
		{"only spaces", "   ", ""},

		// A paste of a path gives the name. Every file dialog on every
		// desktop shows the name, and a name that kept its slash would
		// quietly become a directory.
		{"a forward-slash path", "/repo/ui/git/diff.go", "diff.go"},
		{"a backslash path", `C:\repo\ui\git\diff.go`, "diff.go"},
		{"a mixed path", `C:/repo/ui\git/diff.go`, "diff.go"},
		{"a path with a trailing slash", "/repo/ui/", "ui"},
		{"just separators", "///", ""},

		// A control character is removed rather than cutting the name: a
		// newline in a filename makes a list of files lie about how many
		// lines it has, and cutting at the first one would leave a file
		// called "notes" instead of "notes.md".
		{"a newline is removed", "notes\nREADME.md", "notesREADME.md"},
		{"a NUL is removed", "notes\x00.md", "notes.md"},
		{"a tab is dropped", "a\tb.md", "ab.md"},
		{"a DEL is dropped", "a\x7fb.md", "ab.md"},

		// The punctuation no filesystem takes. Replaced rather than
		// dropped, so "a?b" does not quietly become "ab".
		{"a question mark", "what?.md", "what_.md"},
		{"an asterisk", "x*y.md", "x_y.md"},
		{"a colon", "12:30.log", "12_30.log"},
		{"a less-than and a pipe", "<a|b>.txt", "_a_b_.txt"},
		{"a double quote", `say "hi".txt`, "say _hi_.txt"},

		// Windows silently trims these, so a name that kept them comes
		// back different from the one that was typed.
		{"a leading dot", ".hidden", "hidden"},
		{"a trailing dot", "name.", "name"},
		{"a leading space", " name", "name"},
		{"a trailing space", "name ", "name"},
		{"a name that is only dots", "...", ""},

		// The reserved names. Windows refuses them whatever extension they
		// carry, so "nul.txt" is NUL as far as Windows is concerned.
		{"the console device", "CON", "_CON"},
		{"a reserved name in lower case", "nul", "_nul"},
		{"a reserved name with an extension", "NUL.txt", "_NUL.txt"},
		{"a serial port", "COM1", "_COM1"},
		{"the ninth serial port", "COM9.log", "_COM9.log"},
		{"a parallel port", "LPT1", "_LPT1"},
		{"a name that merely starts reserved", "CONSOLE", "CONSOLE"},
		{"a name that merely contains one", "MYCOM1", "MYCOM1"},
		{"a dotfile that is not reserved", ".gitignore", "gitignore"},

		// The length limit.
		{"a name under the limit", long[:255], long[:255]},
		{"a name over the limit", long, long[:255]},
		// Only the last dot counts, so "archive.tar.gz" keeps ".gz" and
		// loses ".tar". The file still opens with the program it is for;
		// "a-really-long-name.tar" would not.
		{"an over-long name keeps its extension", longName, strings.Repeat("b", 252) + ".gz"},
	}
	for _, tc := range cases {
		if got := Sanitize(tc.in); got != tc.want {
			t.Errorf("%s:\n Sanitize(%q)\n   = %q\nwant %q", tc.name, tc.in, got, tc.want)
		}
	}
}

func TestSanitizeNeverCutsARuneInHalf(t *testing.T) {
	// Truncating by bytes would leave a half-written character at the end,
	// which every filesystem that stores UTF-8 refuses and which renders as
	// a replacement character in the one place somebody is looking.
	long := strings.Repeat("文件", 200)
	got := Sanitize(long)
	if len(got) > maxNameBytes {
		t.Errorf("len = %d, want at most %d", len(got), maxNameBytes)
	}
	if !validUTF8(got) {
		t.Errorf("the result is not valid UTF-8: %q", got)
	}
	if !strings.HasPrefix(long, got) {
		t.Errorf("the result is not a prefix of the input: %q", got)
	}
}

func TestSanitizeOrFallsBackButSanitizeDoesNot(t *testing.T) {
	// A rename field somebody has not typed into must not have "file"
	// appear in it the moment the component is drawn.
	if got := SanitizeOr("", "report.pdf"); got != "report.pdf" {
		t.Errorf("SanitizeOr(\"\", \"report.pdf\") = %q", got)
	}
	if got := SanitizeOr("...", "report.pdf"); got != "report.pdf" {
		t.Errorf("SanitizeOr of nothing usable = %q, want the fallback", got)
	}
	// "???" is usable — it is three characters the filesystem will take —
	// so the fallback does not apply and the name becomes three underscores.
	if got := SanitizeOr("???", "report.pdf"); got != "___" {
		t.Errorf("SanitizeOr(\"???\") = %q, want \"___\"", got)
	}
	if got := SanitizeOr("...", ""); got != "file" {
		t.Errorf("SanitizeOr with no fallback = %q, want \"file\"", got)
	}
	if got := SanitizeOr("...", "CON"); got != "_CON" {
		t.Errorf("a fallback is sanitized too, got %q", got)
	}
	if got := Sanitize(""); got != "" {
		t.Errorf("Sanitize(\"\") = %q, want the empty string", got)
	}
}

// ── Join ────────────────────────────────────────────────────────────────────

func TestJoin(t *testing.T) {
	cases := []struct {
		name  string
		base  string
		parts []string
		want  string
	}{
		{"nothing at all", "", nil, "."},
		{"a base on its own", "/repo/ui", nil, "/repo/ui"},
		{"one segment", "/repo", []string{"ui"}, "/repo/ui"},
		{"several segments", "/repo", []string{"ui", "git"}, "/repo/ui/git"},
		{"an empty segment is skipped", "/repo", []string{"", "ui"}, "/repo/ui"},
		{"only empty segments", "/repo", []string{"", ""}, "/repo"},
		{"a trailing slash on the base", "/repo/", []string{"ui"}, "/repo/ui"},
		{"a trailing slash on a segment", "/repo", []string{"ui/"}, "/repo/ui"},
		{"a leading slash on a segment", "/repo", []string{"/ui"}, "/ui"},
		{"an absolute segment replaces the base", "/repo/ui", []string{"/etc/hosts"}, "/etc/hosts"},
		{"a relative segment is not absolute", "/repo", []string{"ui/git"}, "/repo/ui/git"},
		{"a backslash is normalised", `/repo\ui`, []string{"git"}, "/repo/ui/git"},
		// A drive letter keeps its colon and its slash: "C:/Windows" is one
		// path, and prefixing a "/" to it makes something no Windows tool
		// will open and Linux reads as a file called "C:".
		{"a drive letter is absolute", "/repo", []string{`C:\Windows\System32`}, "C:/Windows/System32"},

		// The dot segments, resolved rather than kept: a path with ".." in
		// it is a path somebody has to think about.
		{"a dot is dropped", "/repo", []string{"."}, "/repo"},
		{"a dot-dot pops one", "/repo/ui/git", []string{".."}, "/repo/ui"},
		{"a dot-dot pops two", "/repo/ui/git", []string{"..", ".."}, "/repo"},
		{"a dot-dot at the root is dropped", "/", []string{".."}, "/"},
		{"a dot-dot past the root is dropped", "/repo", []string{"..", "..", ".."}, "/"},
		{"a dot in the middle is dropped", "/repo", []string{"ui", ".", "git"}, "/repo/ui/git"},
		{"a dot-dot in the middle", "/repo/ui/git", []string{"..", "data"}, "/repo/ui/data"},
		{"everything resolves away", "/", []string{"..", ".."}, "/"},

		// The paths a file interface actually shows.
		{"a home path", "", []string{"/Users/ada/callbacks/ui/git"}, "/Users/ada/callbacks/ui/git"},
		{"a relative path", "ui", []string{"git", "diff.go"}, "/ui/git/diff.go"},
	}
	for _, tc := range cases {
		if got := Join(tc.base, tc.parts...); got != tc.want {
			t.Errorf("%s:\n Join(%q, %q)\n   = %q\nwant %q",
				tc.name, tc.base, tc.parts, got, tc.want)
		}
	}
}

func TestJoinIsIdempotent(t *testing.T) {
	// Joining a path onto itself must not change it, because a path bar
	// builds each crumb's path out of the path it is already in.
	for _, p := range []string{"/repo/ui/git", "/", "/a/b/c/d"} {
		if got := Join(p); got != p {
			t.Errorf("Join(%q) = %q", p, got)
		}
	}
}

// ── HumanSize ───────────────────────────────────────────────────────────────
//
// Decimal, on purpose, so that a number here and a number in the web version
// of this interface are the same number. Each boundary is asserted as an
// exact string.

func TestHumanSize(t *testing.T) {
	cases := []struct {
		in   int64
		want string
	}{
		{0, "0 B"},
		{1, "1 B"},
		{9, "9 B"},
		{999, "999 B"},
		{1000, "1.0 kB"},
		{1023, "1.0 kB"},
		{1024, "1.0 kB"},
		{1500, "1.5 kB"},
		{12_400, "12 kB"},
		{12_600, "13 kB"},
		{15_000, "15 kB"},
		// The unit changes at 999.5, because that is the figure at which
		// "%.0f" prints a thousand — a "1000 kB" line is the one number in
		// this function that is wrong.
		{999_499, "1.0 MB"},
		// The rounding happens before the unit is chosen, so a figure a
		// hair under a thousand does not print as "1000.0 MB".
		{999_999, "1.0 MB"},
		{1_000_000, "1.0 MB"},
		{1_048_576, "1.0 MB"},
		{1_500_000_000, "1.5 GB"},
		{2_400_000_000_000, "2.4 TB"},
		{5_000, "5.0 kB"},
		{-1, "-1 B"},
		{-1_500, "-1.5 kB"},
	}
	for _, tc := range cases {
		if got := HumanSize(tc.in); got != tc.want {
			t.Errorf("HumanSize(%d) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestHumanSizeIsDecimalNotBinary(t *testing.T) {
	// The whole point of choosing SI: the web version of this product shows
	// "1.0 MB" for a file of 1,000,000 bytes and so does this, and a
	// support conversation about "the 1.05 MB file" does not become a
	// lesson in prefixes.
	if got := HumanSize(1_048_576); !strings.Contains(got, "MB") {
		t.Errorf("HumanSize(1 MiB) = %q; the library counts in decimal", got)
	}
	if got := HumanSize(1_048_576); got == "1.0 MiB" {
		t.Error("HumanSize is using binary units")
	}
	// Nothing anywhere says "KiB", "MiB" or "GiB".
	for _, n := range []int64{1, 1024, 1_048_576, 1_073_741_824} {
		if strings.Contains(HumanSize(n), "i") {
			t.Errorf("HumanSize(%d) = %q contains a binary prefix", n, HumanSize(n))
		}
	}
}

func TestHumanSizeNeverPrintsThousandOfAUnit(t *testing.T) {
	for _, n := range []int64{999_999, 999_999_999, 999_999_999_999, 995_000, 999_499} {
		got := HumanSize(n)
		if strings.HasPrefix(got, "1000") {
			t.Errorf("HumanSize(%d) = %q; it should have changed unit", n, got)
		}
	}
}

func TestPercent(t *testing.T) {
	cases := []struct {
		part, whole int64
		want        int
	}{
		{0, 100, 0},
		{50, 100, 50},
		{100, 100, 100},
		{150, 100, 100}, // over the whole: the whole
		{-5, 100, 0},    // a negative share is no share
		{5, 0, 0},       // no whole: nothing to be a fraction of
		{1, 1000, 0},    // under half a point: 0
		{4, 1000, 0},
		{5, 1000, 1}, // exactly half a point rounds up to 1
		{970, 1000, 97},
		{999, 1000, 100}, // 99.9 rounds to a full 100
	}
	for _, tc := range cases {
		if got := Percent(tc.part, tc.whole); got != tc.want {
			t.Errorf("Percent(%d, %d) = %d, want %d", tc.part, tc.whole, got, tc.want)
		}
	}
}

// ── kinds, extensions and icons ─────────────────────────────────────────────

func TestKindOf(t *testing.T) {
	cases := map[string]Kind{
		"main.go":        Text,
		"README.md":      Text,
		"notes.TXT":      Text,
		"config.yaml":    Text,
		"screenshot.png": Image,
		"photo.JPEG":     Image,
		"logo.svg":       Image,
		"release.zip":    Archive,
		"bundle.tar.gz":  Archive,
		"track.mp3":      Media,
		"clip.mov":       Media,
		"app.dmg":        Archive,
		// The conservative direction: an extension nobody has heard of is
		// bytes, not words. A preview that refuses is a sentence; a preview
		// of a binary as text is a screenful of nothing.
		"mystery.qqq": Binary,
		"noextension": Binary,
		"":            Binary,
		// A dotfile's dot is the whole name, not an extension.
		".gitignore": Binary,
	}
	for name, want := range cases {
		if got := KindOf(name); got != want {
			t.Errorf("KindOf(%q) = %v, want %v", name, got, want)
		}
	}
}

func TestKindStringsAndCanPreview(t *testing.T) {
	names := map[Kind]string{Binary: "Binary", Text: "Text", Image: "Image", Archive: "Archive", Media: "Media"}
	for k, want := range names {
		if got := k.String(); got != want {
			t.Errorf("Kind(%d).String() = %q, want %q", int(k), got, want)
		}
	}
	if !Text.CanPreview() || !Image.CanPreview() {
		t.Error("text and images can be previewed")
	}
	for _, k := range []Kind{Binary, Archive, Media} {
		if k.CanPreview() {
			t.Errorf("%v cannot be previewed and should say so", k)
		}
	}
}

func TestExtAndBase(t *testing.T) {
	cases := []struct{ path, ext, base, dir string }{
		{"/repo/ui/git/diff.go", "go", "diff.go", "/repo/ui/git"},
		// Dir normalises, as Join does: a path is compared against strings
		// and opened by something that speaks forward slashes.
		{`C:\repo\main.go`, "go", "main.go", "C:/repo"},
		{"archive.tar.gz", "gz", "archive.tar.gz", "."},
		{"/repo/", "", "repo", "/"},
		{"main.go", "go", "main.go", "."},
		{".gitignore", "", ".gitignore", "."},
		{"/a/b/", "", "b", "/a"},
	}
	for _, tc := range cases {
		if got := Ext(tc.path); got != tc.ext {
			t.Errorf("Ext(%q) = %q, want %q", tc.path, got, tc.ext)
		}
		if got := Base(tc.path); got != tc.base {
			t.Errorf("Base(%q) = %q, want %q", tc.path, got, tc.base)
		}
		if got := Dir(tc.path); got != tc.dir {
			t.Errorf("Dir(%q) = %q, want %q", tc.path, got, tc.dir)
		}
	}
}

func TestIconForHasAFallbackForEveryKind(t *testing.T) {
	// Every kind, including the ones nobody in this repository has a file
	// of, must have a glyph. A kind with no glyph is a file type that draws
	// nothing, which is a bug that looks like a design choice.
	for _, k := range []Kind{Binary, Text, Image, Archive, Media} {
		if _, ok := extIcons[k]; !ok {
			t.Errorf("kind %v has no glyph", k)
		}
	}
	// The fallback for a kind nobody defined is the fallback kind's glyph,
	// so a future kind added without an icon still draws something.
	if extIcons[fallbackKind] == "" {
		t.Error("there is no glyph for the fallback kind")
	}
}

func TestIsDir(t *testing.T) {
	if !IsDir("/repo/ui/") || !IsDir("/") {
		t.Error("a trailing separator makes a directory")
	}
	if IsDir("main.go") {
		t.Error("a file is not a directory")
	}
}

// ── helpers ─────────────────────────────────────────────────────────────────

// validUTF8 is here rather than importing unicode/utf8 in the test, so that
// the test and the package agree about what "cut cleanly" means.
func validUTF8(s string) bool { return strings.ToValidUTF8(s, "") == s }

// ── the components, rendered ────────────────────────────────────────────────
//
// Each component gets at least one headless render. The assertions are on
// what a person reads off the screen and on whether a thing can be found,
// never on a field inside a result.

var demoEntries = []Entry{
	{Name: "src", Path: "/repo/src", Dir: true, Count: 42, Expanded: true, Children: []Entry{
		{Name: "main.go", Path: "/repo/src/main.go", Kind: Text, Size: 1_048, Modified: "2 days ago"},
		{Name: "logo.png", Path: "/repo/src/logo.png", Kind: Image, Size: 1_500_000, Modified: "last week"},
	}},
	{Name: "README.md", Path: "/repo/README.md", Kind: Text, Size: 2_400, Modified: "3 weeks ago"},
	{Name: "callbacks.db", Path: "/repo/callbacks.db", Size: 1_500_000_000, Modified: "an hour ago"},
	{Name: "mystery.qqq", Path: "/repo/mystery.qqq", Size: 12, Modified: "yesterday"},
}

func TestFileExplorerDrawsTheTreeAtItsIndents(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		sel := ""
		FileExplorer(c, &sel, demoEntries, FileExplorerOptions{Height: 320, ShowSize: true})
	}, 620, 360)
	for _, want := range []string{"src", "main.go", "logo.png", "README.md", "callbacks.db", "42 items"} {
		if !tt.HasText(want) {
			t.Errorf("the tree is missing %q; it drew %q", want, tt.Texts())
		}
	}
	// A folder nobody has opened costs one row: its contents are not drawn
	// and are not fetched.
	if tt.HasText("secret.txt") {
		t.Error("a closed folder should not have drawn its contents")
	}
}

func TestFileExplorerNamesEveryGlyph(t *testing.T) {
	// A glyph is the one element with no word of its own, so it has to be
	// given one — and the twisty says whether the folder is open or shut,
	// because a triangle pointing the wrong way is a question.
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		sel := ""
		FileExplorer(c, &sel, demoEntries, FileExplorerOptions{Height: 320})
	}, 620, 360)
	for _, want := range []string{"src folder, open", "main.go file"} {
		if _, ok := tt.Find(want); !ok {
			t.Errorf("nothing is named %q; it drew %q", want, tt.Texts())
		}
	}
}

func TestFileExplorerSaysWhenAFolderIsEmpty(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		sel := ""
		FileExplorer(c, &sel, nil, FileExplorerOptions{Height: 200})
	}, 480, 180)
	if !tt.HasText("This folder is empty") {
		t.Errorf("an empty folder should say so; it drew %q", tt.Texts())
	}
}

func TestFileExplorerNeedsAHeight(t *testing.T) {
	mustPanic(t, "files: FileExplorer needs a Height", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			sel := ""
			FileExplorer(c, &sel, demoEntries, FileExplorerOptions{})
		}, 500, 300)
	})
}

func TestFileExplorerPicksAcrossBuilds(t *testing.T) {
	// The result has to be accumulated across the frame's passes: MyGo
	// builds up to three times so an event's outcome shows in the frame it
	// happened in, and Clicked() is false in the last pass. Reading the
	// result only there would always give -1.
	var picked string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		sel := ""
		if v := FileExplorer(c, &sel, demoEntries, FileExplorerOptions{Height: 320}); v.Picked() != "" {
			picked = v.Picked()
		}
	}, 620, 360)
	if err := tt.Click("README.md"); err != nil {
		t.Fatal(err)
	}
	if picked != "/repo/README.md" {
		t.Errorf("Picked() = %q, want /repo/README.md", picked)
	}
}

func TestFileGridPutsFoldersFirst(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		sel := ""
		FileGrid(c, &sel, demoEntries, FileGridOptions{Height: 300, ShowSize: true})
	}, 620, 320)
	texts := tt.Texts()
	if len(texts) == 0 {
		t.Fatal("the grid drew nothing")
	}
	// Folders come before files: opening one is the common action and it
	// should not be below a screenshot called 2024.
	srcAt, readmeAt := indexOf(texts, "src"), indexOf(texts, "README.md")
	if srcAt < 0 || readmeAt < 0 {
		t.Fatalf("the grid is missing entries: %q", texts)
	}
	if srcAt > readmeAt {
		t.Errorf("the folder came after a file: %q", texts)
	}
	if !tt.HasText("1.5 GB") {
		t.Errorf("ShowSize should have printed the size; it drew %q", texts)
	}
}

func TestPathBarShowsTheSegmentsAsTargets(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		PathBar(c, "/Users/ada/callbacks/ui/git", PathBarOptions{
			Label: "Location", Root: "callbacks",
		})
	}, 700, 120)
	for _, want := range []string{"callbacks", "git"} {
		if !tt.HasText(want) {
			t.Errorf("the path bar is missing %q; it drew %q", want, tt.Texts())
		}
	}
	// The path is never one piece of text: each segment is its own target,
	// because a path of seven segments does not fit in a bar.
	if tt.HasText("/Users/ada/callbacks/ui/git") {
		t.Error("the whole path should not be drawn as one word")
	}
}

func TestPathBarCollapsesTheMiddleOfALongPath(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		PathBar(c, "/a/b/c/d/e/f/g/h", PathBarOptions{Label: "Location", MaxSegments: 3})
	}, 620, 120)
	if !tt.HasText("…") {
		t.Errorf("a long path should collapse its middle; it drew %q", tt.Texts())
	}
	if !tt.HasText("h") {
		t.Errorf("the end of the path is the part the reader came for; it drew %q", tt.Texts())
	}
	if tt.HasText("e") {
		t.Errorf("the collapsed middle should not be drawn; it drew %q", tt.Texts())
	}
}

func TestPathBarNeedsALabel(t *testing.T) {
	mustPanic(t, "files: PathBar needs a Label", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			PathBar(c, "/a/b", PathBarOptions{})
		}, 400, 120)
	})
}

func TestFilePreviewShowsATextFile(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		FilePreview(c, FilePreviewOptions{
			Name: "main.go", Text: "package main\n\nfunc main() {}\n",
			Size: 128, Modified: "2 days ago", Width: 640, Height: 360,
		})
	}, 700, 420)
	for _, want := range []string{"main.go", "package main", "func main() {}", "128 B", "2 days ago"} {
		if !tt.HasText(want) {
			t.Errorf("the preview is missing %q; it drew %q", want, tt.Texts())
		}
	}
}

func TestFilePreviewSaysWhyItCannotShowABinary(t *testing.T) {
	// "Cannot preview" with nothing after it is a dead end. The refusal
	// says what the file is and how big it is, which is what somebody needs
	// to decide whether to download it.
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		FilePreview(c, FilePreviewOptions{
			Name: "callbacks.db", Size: 1_500_000_000, Modified: "an hour ago",
			Width: 640, Height: 360,
		})
	}, 700, 420)
	if !tt.HasText("No preview for binary files") {
		t.Errorf("a binary should be refused in words; it drew %q", tt.Texts())
	}
	if !tt.HasText("1.5 GB") {
		t.Errorf("the refusal should still say how big the file is; it drew %q", tt.Texts())
	}
}

func TestFilePreviewShowsAnImageTheCallerBuilt(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		FilePreview(c, FilePreviewOptions{
			Name: "logo.png", Kind: Image, Size: 1_500_000, Modified: "last week",
			Image: ui.Text(c, "the picture"),
			Width: 640, Height: 360,
		})
	}, 700, 420)
	if !tt.HasText("the picture") {
		t.Errorf("the caller's picture was not drawn; it drew %q", tt.Texts())
	}
}

func TestFilePreviewRefusesToLookAtADisk(t *testing.T) {
	// The text and the image are the caller's because reading a file is the
	// one thing a drawing function must not do: it would block the frame,
	// happen again on every redraw, and make the component untestable.
	mustPanic(t, "files: FilePreview needs the file's name", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			FilePreview(c, FilePreviewOptions{Width: 400, Height: 300})
		}, 500, 350)
	})
	mustPanic(t, "files: FilePreview needs a Width and a Height", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			FilePreview(c, FilePreviewOptions{Name: "a.txt"})
		}, 500, 350)
	})
}

func TestFilePreviewTruncatesALongTextFile(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		FilePreview(c, FilePreviewOptions{
			Name: "app.log", Text: "one\ntwo\nthree\nfour\nfive\n",
			Width: 520, Height: 320, MaxLines: 2,
		})
	}, 600, 380)
	if tt.HasText("five") {
		t.Errorf("the lines past the cap should not be drawn; it drew %q", tt.Texts())
	}
	if !tt.HasText("3 more lines") {
		t.Errorf("the truncation should say how much was left out; it drew %q", tt.Texts())
	}
}

func TestFilePreviewInDarkMode(t *testing.T) {
	// A dark-mode test that only checks the text drew has proved nothing:
	// the text draws the same in both appearances. What has to be checked is
	// that the window resolved the dark palette and that what it drew is in
	// it, so the tokens are read inside the view.
	var resolved theme.Tokens
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: core.Dark})
		resolved = core.Tokens(c)
		FilePreview(c, FilePreviewOptions{
			Name: "main.go", Text: "package main\n", Size: 128,
			Width: 640, Height: 320,
		})
	}, 700, 380)
	tt.SetDark(true)
	tt.Frame()

	for _, want := range []string{"main.go", "package main"} {
		if !tt.HasText(want) {
			t.Errorf("the dark preview is missing %q; it drew %q", want, tt.Texts())
		}
	}
	if resolved.Background != theme.Dark().Background {
		t.Errorf("background = %s, want the dark palette's %s",
			hex(resolved.Background), hex(theme.Dark().Background))
	}
	if !resolved.IsDark() {
		t.Error("the window did not resolve the dark palette")
	}
	if resolved.Text == theme.Light().Text {
		t.Error("the dark window is drawing in the light palette's text ink")
	}
}

func TestRecentFilesDrawsTheFilesOpened(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		RecentFiles(c, []RecentFile{
			{Name: "main.go", Path: "/repo/main.go", Opened: "2 hours ago", Where: "GoLand"},
			{Name: "logo.png", Path: "/repo/logo.png", Kind: Image, Opened: "yesterday", Where: "Preview"},
		}, RecentFilesOptions{Height: 220})
	}, 620, 260)
	for _, want := range []string{"main.go", "2 hours ago", "GoLand", "logo.png", "Preview"} {
		if !tt.HasText(want) {
			t.Errorf("the recent list is missing %q; it drew %q", want, tt.Texts())
		}
	}
}

func TestRecentFilesCapsAtMax(t *testing.T) {
	files := []RecentFile{
		{Name: "a.go", Path: "/a.go", Opened: "now"},
		{Name: "b.go", Path: "/b.go", Opened: "now"},
		{Name: "c.go", Path: "/c.go", Opened: "now"},
	}
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		RecentFiles(c, files, RecentFilesOptions{Height: 220, Max: 2})
	}, 620, 260)
	if tt.HasText("c.go") {
		t.Errorf("Max should have cut the third file; it drew %q", tt.Texts())
	}
}

func TestTransferQueueDrawsABarPerTransfer(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		TransferQueue(c, []Transfer{
			{Name: "logo.png", Done: 500_000, Total: 1_500_000, PerSecond: 250_000},
			{Name: "broken.zip", Done: 0, Total: 900, Failed: true, Error: "the disk is full"},
		}, TransferQueueOptions{Height: 240, WithRate: true})
	}, 680, 300)
	for _, want := range []string{"logo.png", "500 kB", "1.5 MB", "broken.zip", "the disk is full", "Try again"} {
		if !tt.HasText(want) {
			t.Errorf("the queue is missing %q; it drew %q", want, tt.Texts())
		}
	}
	if !tt.HasText("250 kB/s") {
		t.Errorf("WithRate should have printed the rate; it drew %q", tt.Texts())
	}
}

func TestTransferQueueOffersToRetryAFailure(t *testing.T) {
	// A queue that drops a failure the moment it happens has lost the only
	// entry anybody needed to read.
	var retried = -99
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		if v := TransferQueue(c, []Transfer{
			{Name: "broken.zip", Done: 0, Total: 900, Failed: true, Error: "the disk is full"},
			{Name: "fine.bin", Done: 10, Total: 20},
		}, TransferQueueOptions{Height: 220}); v.Retried() >= 0 {
			retried = v.Retried()
		}
	}, 680, 300)
	if err := tt.Click("Try again"); err != nil {
		t.Fatal(err)
	}
	if retried != 0 {
		t.Errorf("Retried() = %d, want 0", retried)
	}
}

func TestRenameInlineShowsWhatWasTypedAndCleansOnSubmit(t *testing.T) {
	// The field shows the raw thing somebody typed: a rename box that
	// shows "report_.pdf" while "report?.pdf" is being written shows them
	// something they did not write, and puts the caret somewhere they did
	// not put it.
	name := "report?.pdf"
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		RenameInline(c, &name, RenameInlineOptions{
			Label: "New name", Original: "report.pdf",
		})
	}, 560, 160)
	if !tt.HasText("report?.pdf") {
		t.Errorf("the field should show what was typed; it drew %q", tt.Texts())
	}
	if Sanitize("report?.pdf") != "report_.pdf" {
		t.Fatal("the test's own assumption about Sanitize is wrong")
	}
}

func TestRenameInlineNeedsANameToPointAt(t *testing.T) {
	mustPanic(t, "files: RenameInline needs the name", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			RenameInline(c, nil, RenameInlineOptions{Label: "New name"})
		}, 500, 150)
	})
	mustPanic(t, "files: RenameInline needs a Label", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			n := "a"
			RenameInline(c, &n, RenameInlineOptions{})
		}, 500, 150)
	})
}

func TestStorageUsagePrintsTheFigures(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		StorageUsage(c, StorageUsageOptions{
			Used: 9_600_000_000, Quota: 10_000_000_000, Plan: "Team",
		})
	}, 620, 160)
	if !tt.HasText("9.6 GB / 10 GB (96%)") {
		t.Errorf("the figures should be printed as well as drawn; it drew %q", tt.Texts())
	}
	if !tt.HasText("Team") {
		t.Errorf("the plan name is missing; it drew %q", tt.Texts())
	}
}

func TestStorageUsageSaysWhenThereIsNoLimit(t *testing.T) {
	// A quota of zero is "no limit", not "full", and drawing a full bar for
	// it would say the disk is out of space when it is not.
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		StorageUsage(c, StorageUsageOptions{Used: 9_600_000_000})
	}, 620, 160)
	if !tt.HasText("no limit") {
		t.Errorf("a quota of zero should be said in words; it drew %q", tt.Texts())
	}
	if tt.HasText("100%") {
		t.Errorf("an unlimited storage is not full; it drew %q", tt.Texts())
	}
}

func TestFileOperationProgressSaysSomethingWhenItCannotCount(t *testing.T) {
	// An indeterminate operation shown at 0% is a claim. Most long
	// operations cannot count until somebody has walked the tree.
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		FileOperationProgress(c, FileOperationOptions{
			Title: "Copying", Detail: "/repo/src", Indeterminate: true,
		})
	}, 620, 200)
	if !tt.HasText("Working…") {
		t.Errorf("an operation that cannot count should say so; it drew %q", tt.Texts())
	}
	if tt.HasText("0%") {
		t.Errorf("an indeterminate bar should not claim a number; it drew %q", tt.Texts())
	}
}

func TestFileOperationProgressCountsTheItems(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		FileOperationProgress(c, FileOperationOptions{
			Title: "Copying 4,000 files", Done: 1234, Total: 4000,
		})
	}, 620, 200)
	if !tt.HasText("1234 / 4000") {
		t.Errorf("the counts should be printed; it drew %q", tt.Texts())
	}
}

func TestFileOperationProgressNeedsAVerb(t *testing.T) {
	mustPanic(t, "files: FileOperationProgress needs a Title", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			FileOperationProgress(c, FileOperationOptions{})
		}, 500, 180)
	})
}

func TestShareDialogDrawsThePeopleAndTheirPermissions(t *testing.T) {
	open := true
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		invite := "grace@example.com"
		ShareDialog(c, &open, ShareDialogOptions{
			Title: "Share callbacks.db", Subtitle: "/repo/callbacks.db",
			Invite: &invite,
			People: []Person{
				{Name: "Ada Lovelace", Email: "ada@example.com", Permission: IsOwner},
				{Name: "Grace Hopper", Email: "grace@example.com", Permission: CanEdit, You: true},
			},
			Width: 640,
		})
	}, 720, 520)
	for _, want := range []string{
		"Share callbacks.db", "/repo/callbacks.db",
		"Ada Lovelace", "Grace Hopper", "Can edit", "Owner", "you",
		"Invite by email or name",
	} {
		if !tt.HasText(want) {
			t.Errorf("the share panel is missing %q; it drew %q", want, tt.Texts())
		}
	}
}

func TestShareDialogIsClosedUntilItIsOpened(t *testing.T) {
	open := false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ShareDialog(c, &open, ShareDialogOptions{Title: "Share", People: []Person{{Name: "Ada"}}})
	}, 600, 400)
	if tt.HasText("Share") {
		t.Errorf("a closed panel should have drawn nothing; it drew %q", tt.Texts())
	}
}

func TestShareDialogWritesTheCallersPeople(t *testing.T) {
	// A share list that kept a copy would be out of date the first time
	// somebody is removed by somebody else.
	people := []Person{{Name: "Grace Hopper", Permission: CanView}}
	var changed = -99
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		open := true
		if v := ShareDialog(c, &open, ShareDialogOptions{
			Title: "Share callbacks.db", People: people, Width: 640,
		}); v.Changed() >= 0 {
			changed = v.Changed()
		}
	}, 720, 480)
	if err := tt.Click("Grace Hopper: Can edit"); err != nil {
		t.Fatal(err)
	}
	if changed != 0 {
		t.Fatalf("Changed() = %d, want 0", changed)
	}
	if people[0].Permission != CanEdit {
		t.Errorf("the caller's slice still says %v", people[0].Permission)
	}
}

func TestShareDialogNeedsItsTitle(t *testing.T) {
	mustPanic(t, "files: ShareDialog needs a Title", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			open := true
			ShareDialog(c, &open, ShareDialogOptions{})
		}, 500, 400)
	})
}

func TestPermissionSelectIsSeparateFromTheDialog(t *testing.T) {
	// It is asked for on its own: a file row in a list has a permissions
	// popover and a sidebar has one per person, and neither is a dialog.
	perm := CanView
	var picked *Permission
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		if v := PermissionSelect(c, &perm, PermissionSelectOptions{
			Label: "Access for callbacks.db",
		}); v.Picked() != nil {
			picked = v.Picked()
		}
	}, 640, 160)
	if err := tt.Click("Access for callbacks.db: Can comment"); err != nil {
		t.Fatal(err)
	}
	if picked == nil || *picked != CanComment {
		t.Fatalf("Picked() = %v, want CanComment", picked)
	}
	if perm != CanComment {
		t.Errorf("the caller's permission is %v, want CanComment", perm)
	}
}

func TestPermissionSelectOffersTheFourLeastFirst(t *testing.T) {
	got := Permissions(nil)
	want := []Permission{CanView, CanComment, CanEdit, IsOwner}
	if len(got) != len(want) {
		t.Fatalf("got %d permissions, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("permission %d is %v, want %v (least first, so the arrow moves towards more)", i, got[i], want[i])
		}
	}
}

func TestPermissionSelectNeedsASelectionAndALabel(t *testing.T) {
	mustPanic(t, "files: PermissionSelect needs a permission", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			PermissionSelect(c, nil, PermissionSelectOptions{Label: "Access"})
		}, 500, 150)
	})
	mustPanic(t, "files: PermissionSelect needs a Label", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			perm := CanView
			PermissionSelect(c, &perm, PermissionSelectOptions{})
		}, 500, 150)
	})
}

func TestArchiveViewerDrawsTheEntriesAndTheCompression(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ArchiveViewer(c, []ArchiveEntry{
			{Name: "ui/", Dir: true},
			{Name: "ui/main.go", Size: 1_048_576, Compressed: 262_144, Modified: "2 days ago"},
			{Name: "logo.png", Size: 1_500_000, Compressed: 1_480_000},
		}, ArchiveViewerOptions{Name: "callbacks.zip", Height: 260, WithRatio: true})
	}, 680, 320)
	for _, want := range []string{"callbacks.zip", "3 entries", "main.go", "1.0 MB", "75% saved"} {
		if !tt.HasText(want) {
			t.Errorf("the archive listing is missing %q; it drew %q", want, tt.Texts())
		}
	}
	// A file that did not compress still says its ratio — 1% is the fact
	// somebody is looking for, and it is usually the reason an archive is
	// as big as it is. What is hidden is a ratio for a file with no
	// compressed size at all, because there is nothing to compare.
	if !tt.HasText("1% saved") {
		t.Errorf("the barely-compressed file should say its ratio; it drew %q", tt.Texts())
	}
	if !tt.HasText("1.0 MB · 75% saved") {
		t.Errorf("the well-compressed file should say its ratio; it drew %q", tt.Texts())
	}
}

func TestArchiveViewerNeedsAHeight(t *testing.T) {
	mustPanic(t, "files: ArchiveViewer needs a Height", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			ArchiveViewer(c, nil, ArchiveViewerOptions{})
		}, 500, 300)
	})
}

func TestReactionsFillsTheOnesYouAreIn(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Reactions(c, []Reaction{
			{Emoji: "👍", Count: 4, Mine: true},
			{Emoji: "🎉", Count: 2},
		}, ReactionsOptions{Label: "Reactions on callbacks.db"})
	}, 620, 160)
	for _, want := range []string{"👍, 4 reactions", "🎉, 2 reactions", "Add a reaction"} {
		if _, ok := tt.Find(want); !ok {
			t.Errorf("a reaction chip should be named %q; it drew %q", want, tt.Texts())
		}
	}
}

func TestReactionsCountsWhatItHid(t *testing.T) {
	// A row that quietly dropped four reactions is lying about how popular
	// a file is, and the count is the honest summary of what was left out.
	all := []Reaction{
		{Emoji: "👍", Count: 4}, {Emoji: "🎉", Count: 2},
		{Emoji: "🚀", Count: 3}, {Emoji: "❤️", Count: 1},
	}
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		Reactions(c, all, ReactionsOptions{Label: "Reactions", Max: 2})
	}, 620, 160)
	if !tt.HasText("+4") {
		t.Errorf("the hidden reactions should be counted; it drew %q", tt.Texts())
	}
	if tt.HasText("🚀") {
		t.Errorf("Max should have cut the third reaction; it drew %q", tt.Texts())
	}
}

func TestReactionsNeedsALabel(t *testing.T) {
	mustPanic(t, "files: Reactions needs a Label", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			Reactions(c, nil, ReactionsOptions{})
		}, 500, 150)
	})
}

func TestLiveIndicatorDrawsTheLivelyColour(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ui.Row(c).AlignItems(ui.Center).Gap(12).Children(func() {
			LiveIndicator(c, LiveIndicatorOptions{Label: "Live", Live: true})
			LiveIndicator(c, LiveIndicatorOptions{Label: "Idle", Live: false})
		})
	}, 300, 120)
	for _, want := range []string{"Live", "Idle"} {
		if _, ok := tt.Find(want); !ok {
			t.Errorf("a dot with no name says nothing out loud; it drew %q", tt.Texts())
		}
	}
}

func TestLiveIndicatorNeedsALabel(t *testing.T) {
	mustPanic(t, "files: LiveIndicator needs a Label", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			LiveIndicator(c, LiveIndicatorOptions{})
		}, 300, 120)
	})
}

func TestPresenceAvatarsShowWhoIsHere(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		PresenceAvatars(c, []Viewer{
			{Name: "Ada Lovelace", Doing: "Editing", Here: true},
			{Name: "Grace Hopper", Here: true},
			{Name: "Alan Turing", Here: false},
		}, PresenceAvatarsOptions{Total: 3})
	}, 420, 140)
	// Nine initials are nine noises, so the cluster is named as a sentence
	// about who is here rather than as a list of letters.
	if _, ok := tt.Find("2 of 3 here: Ada Lovelace, Editing, Grace Hopper, here and Alan Turing, away"); !ok {
		t.Errorf("the cluster should name who is here; it drew %q", tt.Texts())
	}
	if !tt.HasText("AL") {
		t.Errorf("the faces should be drawn; it drew %q", tt.Texts())
	}
}

func TestPresenceAvatarsSaysNobodyWhenNobodyIs(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		PresenceAvatars(c, nil, PresenceAvatarsOptions{})
	}, 480, 140)
	if !tt.HasText("Nobody else is here") {
		t.Errorf("an empty cluster should say so; it drew %q", tt.Texts())
	}
}

func TestRemoteCursorIsFollowable(t *testing.T) {
	var followed bool
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		// The press is reported, not held: a component that followed
		// somebody itself would be scrolling somebody else's view on
		// their own machine, which is not a thing a library may do.
		if v := RemoteCursor(c, RemoteCursorOptions{
			Name: "Ada Lovelace", Doing: "Editing", X: 40, Y: 30,
		}); v.Followed() {
			followed = true
		}
	}, 420, 240)
	// The name is on the chip, not in a tooltip: a tooltip arrives on a
	// hover and this arrives under a moving cursor, and somebody typing
	// fast never hovers anything.
	if _, ok := tt.Find("Ada Lovelace, Editing"); !ok {
		t.Errorf("a remote cursor must name who it is or it cannot be followed; it drew %q", tt.Texts())
	}
	if err := tt.Click("Ada Lovelace, Editing"); err != nil {
		t.Fatal(err)
	}
	if !followed {
		t.Error("pressing the cursor did not report Followed()")
	}
}

func TestRemoteCursorNeedsAName(t *testing.T) {
	// A pointer with nobody's on it cannot be followed, and a follower with
	// no target is the worst thing to hand a caller.
	mustPanic(t, "files: RemoteCursor needs a Name", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			RemoteCursor(c, RemoteCursorOptions{})
		}, 400, 240)
	})
}

func TestFileIconNeedsTheFilesName(t *testing.T) {
	mustPanic(t, "files: FileIcon needs the file's name", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			FileIcon(c, "", FileIconOptions{})
		}, 300, 150)
	})
}

func TestFileIconIsNamedForTheFileItStandsFor(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ui.Row(c).Gap(12).Children(func() {
			FileIcon(c, "main.go", FileIconOptions{})
			FileIcon(c, "mystery.qqq", FileIconOptions{})
		})
	}, 300, 140)
	for _, want := range []string{"main.go file", "mystery.qqq file"} {
		if _, ok := tt.Find(want); !ok {
			t.Errorf("a glyph must be named for the file it stands for; it drew %q", tt.Texts())
		}
	}
}

// ── helpers ─────────────────────────────────────────────────────────────────

func mustPanic(t *testing.T, want string, f func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("expected a panic mentioning %q", want)
			return
		}
		msg, ok := r.(string)
		if !ok {
			t.Errorf("panic value is %T, want a string: %v", r, r)
			return
		}
		if !strings.Contains(msg, want) {
			t.Errorf("panic = %q, want it to mention %q", msg, want)
		}
	}()
	f()
}

func indexOf(list []string, s string) int {
	for i, v := range list {
		if v == s {
			return i
		}
	}
	return -1
}

// hex is a colour as the web version of this interface writes it.
func hex(c ui.Color) string {
	const digits = "0123456789abcdef"
	buf := []byte{'#', 0, 0, 0, 0, 0, 0}
	n := 0
	for _, v := range []uint8{c.R, c.G, c.B} {
		buf[1+n*2] = digits[v>>4]
		buf[2+n*2] = digits[v&0xf]
		n++
	}
	return string(buf)
}
