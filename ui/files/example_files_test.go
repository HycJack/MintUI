package files_test

import (
	"fmt"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/files"
)

// The examples in this file are the package's documentation: they run, and
// their output is checked. A component whose signature has drifted stops
// compiling here first.

func ExampleSanitize() {
	for _, in := range []string{
		"/repo/ui/git/diff.go", // a paste of a path gives the name
		"notes\nREADME.md",     // a control character goes
		"what?.md",             // punctuation becomes an underscore
		"NUL.txt",              // a reserved name is escaped
	} {
		fmt.Printf("%q -> %q\n", in, files.Sanitize(in))
	}
	// Output:
	// "/repo/ui/git/diff.go" -> "diff.go"
	// "notes\nREADME.md" -> "notesREADME.md"
	// "what?.md" -> "what_.md"
	// "NUL.txt" -> "_NUL.txt"
}

func ExampleJoin() {
	fmt.Println(files.Join("/Users/ada", "callbacks", "ui", "git", "diff.go"))
	fmt.Println(files.Join("/Users/ada/callbacks/ui/git", "..", "..", "data"))
	fmt.Println(files.Join("/"))
	// Output:
	// /Users/ada/callbacks/ui/git/diff.go
	// /Users/ada/callbacks/data
	// /
}

// HumanSize counts in decimal — powers of a thousand — because the web
// version of this interface counts in decimal too, and a person comparing
// the two is looking at the same file.
func ExampleHumanSize() {
	for _, n := range []int64{0, 999, 1000, 15_000, 1_500_000_000} {
		fmt.Println(files.HumanSize(n))
	}
	// Output:
	// 0 B
	// 999 B
	// 1.0 kB
	// 15 kB
	// 1.5 GB
}

func ExampleKindOf() {
	for _, name := range []string{"main.go", "logo.png", "bundle.zip", "mystery.qqq"} {
		fmt.Printf("%-12s %s\n", name, files.KindOf(name))
	}
	// Output:
	// main.go      Text
	// logo.png     Image
	// bundle.zip   Archive
	// mystery.qqq  Binary
}

func ExampleFileExplorer() {
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		selected := ""
		entries := []files.Entry{
			{Name: "src", Path: "/repo/src", Dir: true, Count: 2, Expanded: true, Children: []files.Entry{
				{Name: "main.go", Path: "/repo/src/main.go", Kind: files.Text, Size: 1_048},
			}},
			{Name: "README.md", Path: "/repo/README.md", Kind: files.Text, Size: 2_400},
		}
		// Nothing is loaded: an entry with children that is not expanded is
		// one row, so a tree behind nine closed folders costs nine rows.
		if v := files.FileExplorer(c, &selected, entries, files.FileExplorerOptions{
			Height: 240, ShowSize: true,
		}); v.Picked() != "" {
			fmt.Println("opened", v.Picked())
		}
	}, 560, 280)
	// Output:
}

func ExamplePathBar() {
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		// Each segment is its own target rather than the whole path as one
		// piece of text: a path of seven segments does not fit in a bar.
		files.PathBar(c, "/Users/ada/callbacks/ui/git", files.PathBarOptions{
			Label: "Location", Root: "callbacks",
		})
	}, 620, 120)
	// Output:
}

// A preview never opens a file. The text and the picture are the caller's,
// because reading one would block the frame and happen again on every
// redraw.
func ExampleFilePreview() {
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		files.FilePreview(c, files.FilePreviewOptions{
			Name: "main.go", Text: "package main\n", Size: 128,
			Width: 560, Height: 280,
		})
	}, 620, 320)
	// Output:
}

func ExampleRenameInline() {
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		name := "report?.pdf"
		// The field shows what was typed; the cleaning happens on submit,
		// so the caret is never somewhere the caller did not put it.
		if v := files.RenameInline(c, &name, files.RenameInlineOptions{
			Label: "New name", Original: "report.pdf",
		}); v.Submitted() {
			fmt.Println(files.SanitizeOr(name, "report.pdf"))
		}
	}, 560, 160)
	// Output:
}

func ExampleStorageUsage() {
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		files.StorageUsage(c, files.StorageUsageOptions{
			Used: 9_600_000_000, Quota: 10_000_000_000, Plan: "Team",
		})
	}, 620, 160)
	// Output:
}
