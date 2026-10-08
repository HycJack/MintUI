package agent

import (
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
)

// The three parts added with parts.go: the roster of agents, the small code
// block, and the on/off row. Same arrangement as the rest of this file's
// tests — what a component draws in both appearances, and what it refuses to
// be given without.

func wantsPanic(t *testing.T, what string, view func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("%s should have panicked", what)
		}
	}()
	view()
}

// ── AgentRows ──────────────────────────────────────────────────────────────

// TestAgentRowsListsEachAgent: every row's name, status and figure make it
// to the screen, one compact line per agent.
func TestAgentRowsListsEachAgent(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		AgentRows(c, []AgentRow{
			{Name: "Ada", Status: "Running", Meta: "3m"},
			{Name: "Grace", Status: "Done", Meta: "12 steps"},
		})
	}, 480, 240)
	for _, want := range []string{
		"Ada", "Running", "3m", "Grace", "Done", "12 steps",
	} {
		if !tt.HasText(want) {
			t.Errorf("the roster does not show %q; texts were %v", want, tt.Texts())
		}
	}
}

// TestAgentRowsInsistOnNames: the roster is read by the names, and an avatar
// and a status with no name are a mark and a word about nobody.
func TestAgentRowsInsistOnNames(t *testing.T) {
	wantsPanic(t, "an empty roster", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			AgentRows(c, nil)
		}, 400, 200)
	})
	wantsPanic(t, "a row with no name", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			AgentRows(c, []AgentRow{{Status: "Running"}})
		}, 400, 200)
	})
}

// TestAgentRowsInDarkMode draws the roster under the dark palette and asks
// for the words back.
func TestAgentRowsInDarkMode(t *testing.T) {
	tt := dark(t, core.Dark, func(c *ui.Context) {
		AgentRows(c, []AgentRow{{Name: "Ada", Status: "Waiting"}})
	})
	for _, want := range []string{"Ada", "Waiting"} {
		if !tt.HasText(want) {
			t.Errorf("the dark roster does not show %q", want)
		}
	}
}

// ── AgentCode ──────────────────────────────────────────────────────────────

// TestAgentCodeShowsTheCodeAndItsLanguage: the snippet is on screen in full,
// the language tag is beside it, and the copy button is named for what it
// copies.
func TestAgentCodeShowsTheCodeAndItsLanguage(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		AgentCode(c, AgentCodeOptions{
			Code: "fmt.Println(\"hi\")", Lang: "go", Copy: true,
		})
	}, 480, 220)
	for _, want := range []string{"fmt.Println(\"hi\")", "go", "Copy code"} {
		if !tt.HasText(want) {
			t.Errorf("the code block does not show %q; texts were %v", want, tt.Texts())
		}
	}
}

// TestAgentCodeNeedsSomethingInIt: an empty block is a plate, and the block
// is for the code in it.
func TestAgentCodeNeedsSomethingInIt(t *testing.T) {
	wantsPanic(t, "a code block with no code", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			AgentCode(c, AgentCodeOptions{Lang: "go"})
		}, 400, 200)
	})
}

// TestAgentCodeInDarkMode draws the block under the dark palette — the one
// appearance a monospaced block most often lives in.
func TestAgentCodeInDarkMode(t *testing.T) {
	tt := dark(t, core.Dark, func(c *ui.Context) {
		AgentCode(c, AgentCodeOptions{Code: "x := 1", Lang: "go"})
	})
	for _, want := range []string{"x := 1", "go"} {
		if !tt.HasText(want) {
			t.Errorf("the dark code block does not show %q", want)
		}
	}
}

// ── AgentToggle ────────────────────────────────────────────────────────────

// TestAgentToggleDrawsTheRow: the name and the description on the left, and
// the switch on the right, answering to the agent's name so a screen reader
// says what it turns off.
func TestAgentToggleDrawsTheRow(t *testing.T) {
	on := true
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		AgentToggle(c, &on, AgentToggleOptions{
			Name: "Reviewer", Description: "Checks the diff before merge",
		})
	}, 480, 160)
	for _, want := range []string{"Reviewer", "Checks the diff before merge"} {
		if !tt.HasText(want) {
			t.Errorf("the toggle row does not show %q; texts were %v", want, tt.Texts())
		}
	}
}

// TestAgentToggleInsistsOnItsFlagAndItsAgent: the switch writes the caller's
// bool, and a row without an agent beside it is about nothing.
func TestAgentToggleInsistsOnItsFlagAndItsAgent(t *testing.T) {
	wantsPanic(t, "a toggle with no flag to point at", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			AgentToggle(c, nil, AgentToggleOptions{Name: "Reviewer"})
		}, 400, 200)
	})
	wantsPanic(t, "a toggle with no agent", func() {
		on := true
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			AgentToggle(c, &on, AgentToggleOptions{Description: "Checks the diff"})
		}, 400, 200)
	})
}

// TestAgentToggleInDarkMode draws the row under the dark palette.
func TestAgentToggleInDarkMode(t *testing.T) {
	on := false
	tt := dark(t, core.Dark, func(c *ui.Context) {
		AgentToggle(c, &on, AgentToggleOptions{Name: "Reviewer"})
	})
	if !tt.HasText("Reviewer") {
		t.Error("the dark toggle row lost its name")
	}
}
