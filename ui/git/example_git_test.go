package git_test

import (
	"fmt"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/git"
)

// The examples in this file are the package's documentation: they run, and
// their output is checked. A component whose signature has drifted stops
// compiling here first.

func ExampleStatusLetter() {
	for _, s := range []git.FileStatus{
		git.Modified, git.Added, git.Deleted, git.Renamed, git.Untracked,
	} {
		fmt.Printf("%-2s %s\n", git.StatusLetter(s), git.StatusWord(s))
	}
	// Output:
	// M  Modified
	// A  Added
	// D  Deleted
	// R  Renamed
	// ?? Untracked
}

// The status decides the letter, the word and the colour, and the colour
// comes from the severity ramp rather than from a choice made here.
func ExampleStatusSeverity() {
	fmt.Println(git.StatusSeverity(git.Conflicted))
	fmt.Println(git.StatusSeverity(git.Added))
	// Output:
	// Danger
	// Success
}

// A diff is data before it is a picture, and SplitDiff is where the line
// numbers are decided. An added line is in the new file only, so its old
// number is 0 rather than a copy of the new one.
func ExampleSplitDiff() {
	unified := "@@ -1,3 +1,4 @@\n one\n-two\n+two\n+three\n three\n"
	for _, line := range git.SplitDiff(unified) {
		fmt.Printf("%s%3d %3d %q\n", line.Op, line.OldNo, line.NewNo, line.Text)
	}
	// Output:
	//    1   1 "one"
	// -  2   0 "two"
	// +  0   2 "two"
	// +  0   3 "three"
	//    3   4 "three"
}

// StripANSI is what a width is measured with. A row of text whose width
// includes fourteen invisible characters wraps in the wrong place.
func ExampleStripANSI() {
	fmt.Println(git.StripANSI("\x1b[1;31mfatal:\x1b[0m not a repository"))
	// Output:
	// fatal: not a repository
}

// A view is three lines long: core.Use once at the top, then the components.
func ExampleBranchList() {
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		selected := 0
		branches := []git.Branch{
			{Name: "main", Current: true, Subject: "Ship the composer"},
			{Name: "fix/crash", Subject: "Don't panic on an empty body", Ahead: 2},
		}
		if v := git.BranchList(c, &selected, branches, git.BranchListOptions{
			Height: 200,
		}); v.Picked() >= 0 {
			// The result is this frame's press. The selection itself is the
			// caller's pointer and stays where they left it.
			fmt.Println("switched to", branches[v.Picked()].Name)
		}
	}, 480, 200)
	// Output:
}

func ExampleDiffViewer() {
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		git.DiffViewer(c, "@@ -1,3 +1,4 @@\n one\n-two\n+two\n+three\n",
			git.DiffViewerOptions{Height: 220, File: "main.go", WithHunks: true})
	}, 640, 260)
	// Output:
}

// The lanes of a commit graph are computed from the parents, not from the
// order the commits happened to arrive in, which is what makes a merge draw
// where it belongs.
func ExampleCommitsFor() {
	history := []git.Commit{
		{Hash: "c1"},
		{Hash: "c2", Parents: []string{"c1"}},
		{Hash: "s1", Parents: []string{"c1"}},
		{Hash: "top", Parents: []string{"c2", "s1"}},
	}
	for _, row := range git.CommitsFor(history, 0) {
		kind := " "
		if row.Merge {
			kind = "*"
		}
		fmt.Printf("%s col=%d lanes=%v\n", kind, row.Column, row.Lanes)
	}
	// Output:
	//   col=0 lanes=[]
	//   col=0 lanes=[0]
	//   col=1 lanes=[0 1]
	// * col=0 lanes=[0 1]
}

func ExampleCommitInput() {
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		message := "Fix the crash on an empty body"
		var amend bool
		if v := git.CommitInput(c, &message, git.CommitInputOptions{
			Label: "Commit message", Staged: 3, Amend: &amend, Height: 90,
		}); v.Committed() {
			// Committing runs git. This component says it was pressed and
			// nothing more: it has no working directory.
			fmt.Println("committing", message)
		}
	}, 560, 300)
	// Output:
}

func ExampleGitStatusBadge() {
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ui.Column(c).Gap(4).Children(func() {
			git.GitStatusBadge(c, git.Modified, git.GitStatusBadgeOptions{WithWord: true})
			git.GitStatusBadge(c, git.Conflicted, git.GitStatusBadgeOptions{WithWord: true})
		})
	}, 420, 160)
	// Output:
}
