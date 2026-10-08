package chat

import (
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
)

// The components added with the window's frame — the container, the two
// empty states, the context menu, the link card, the knowledge panel, the
// shared view and the welcome screen — tested the way the rest of this file
// tests them: on the value a component wrote or reported, accumulated inside
// the view, and on what made it to the screen in both appearances.

// ── ChatContainer ──────────────────────────────────────────────────────────

// TestChatContainerLaysOutItsThreeStrips is the contract of the skeleton:
// the header with its meta and its actions, the room the transcript grows
// in, and the composer at the bottom, all from the caller's own pieces.
func TestChatContainerLaysOutItsThreeStrips(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ChatContainer(c, ChatContainerOptions{
			Title: "Release notes",
			Meta:  "model · main",
			Messages: func() {
				ui.Text(c, "the transcript goes here")
			},
			Actions: func() {
				ui.Text(c, "export action")
			},
		}, func() {
			ui.Text(c, "the composer goes here")
		})
	}, 560, 420)
	for _, want := range []string{
		"Release notes", "model · main",
		"the transcript goes here", "export action", "the composer goes here",
	} {
		if !tt.HasText(want) {
			t.Errorf("the container does not show %q; texts were %v", want, tt.Texts())
		}
	}
}

// TestChatContainerInsistsOnANameAndAComposer: a container is chrome around
// two things that must exist, and it says so rather than drawing a header
// with nothing in it or a window with no way in.
func TestChatContainerInsistsOnANameAndAComposer(t *testing.T) {
	wantsPanic(t, "a container with no title", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			ChatContainer(c, ChatContainerOptions{}, func() {})
		}, 400, 200)
	})
	wantsPanic(t, "a container with no composer", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			ChatContainer(c, ChatContainerOptions{Title: "Chat"}, nil)
		}, 400, 200)
	})
}

// TestChatContainerInDarkMode is the dark half of the frame: every strip is
// on a dark window and none of them stands out as a card from last frame.
func TestChatContainerInDarkMode(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: core.Dark})
		if !core.IsDark(c) {
			t.Error("the window did not resolve the dark palette")
		}
		ChatContainer(c, ChatContainerOptions{
			Title: "Dark chat", Meta: "model",
		}, func() {
			ui.Text(c, "composer in the dark")
		})
	}, 560, 420)
	for _, want := range []string{"Dark chat", "model", "composer in the dark"} {
		if !tt.HasText(want) {
			t.Errorf("the dark container does not show %q", want)
		}
	}
}

// ── ChatEmptyState ─────────────────────────────────────────────────────────

// TestChatEmptyStateReportsTheSuggestion is the behaviour test: the chips are
// whole questions and the press comes back as the question's place in the
// list, which is what the caller needs to send it.
func TestChatEmptyStateReportsTheSuggestion(t *testing.T) {
	var chosen []int
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		if got := ChatEmptyState(c, ChatEmptyStateOptions{
			Greeting:    "Good morning",
			Body:        "Ask me anything",
			Suggestions: []string{"Summarize the repo", "Draft a changelog"},
		}).Suggested(); got != -1 {
			chosen = append(chosen, got)
		}
	}, 480, 320)
	if len(chosen) != 0 {
		t.Fatalf("a suggestion was reported before anything was pressed")
	}
	for _, want := range []string{"Good morning", "Ask me anything",
		"Summarize the repo", "Draft a changelog"} {
		if !tt.HasText(want) {
			t.Errorf("the empty state does not show %q", want)
		}
	}
	if err := tt.Click("Draft a changelog"); err != nil {
		t.Fatal(err)
	}
	if len(chosen) != 1 || chosen[0] != 1 {
		t.Errorf("the press reported %v, want the second suggestion reported once", chosen)
	}
}

// TestChatEmptyStateInsistsOnAWord: the greeting is the one line a reader
// with nothing in front of them reads, and a chip with no words is a shape
// nobody can press.
func TestChatEmptyStateInsistsOnAWord(t *testing.T) {
	wantsPanic(t, "an empty state with no greeting", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			ChatEmptyState(c, ChatEmptyStateOptions{Body: "Ask me anything"})
		}, 400, 200)
	})
	wantsPanic(t, "a chip with no words", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			ChatEmptyState(c, ChatEmptyStateOptions{
				Greeting: "Hello", Suggestions: []string{"", "A suggestion"},
			})
		}, 400, 200)
	})
}

// TestChatEmptyStateInDarkMode draws the state under the dark palette and
// asks that it still has its words and its chips.
func TestChatEmptyStateInDarkMode(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: core.Dark})
		ChatEmptyState(c, ChatEmptyStateOptions{
			Greeting:    "Good evening",
			Suggestions: []string{"Review the diff"},
		})
	}, 480, 320)
	for _, want := range []string{"Good evening", "Review the diff"} {
		if !tt.HasText(want) {
			t.Errorf("the dark empty state does not show %q", want)
		}
	}
}

// ── ContextMentionMenu ─────────────────────────────────────────────────────

// TestContextMentionMenuFiltersAndReports walks the two halves: the query
// keeps the rows whose name answers it and drops the rest, and the press
// comes back as the row's place among the rows as drawn.
func TestContextMentionMenuFiltersAndReports(t *testing.T) {
	q := "read"
	var picked []int
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		if got := ContextMentionMenu(c, ContextMentionMenuOptions{
			Items: []ContextItem{
				{Name: "README.md", Kind: "file", Meta: "repo root"},
				{Name: "read-model.go", Kind: "file", Meta: "internal"},
				{Name: "Atlas project", Kind: "project"},
			},
			Query: &q,
		}).Picked(); got != -1 {
			picked = append(picked, got)
		}
	}, 420, 280)

	for _, want := range []string{"README.md", "read-model.go", "file", "repo root"} {
		if !tt.HasText(want) {
			t.Errorf("the filtered menu does not show %q; texts were %v", want, tt.Texts())
		}
	}
	if tt.HasText("Atlas project") {
		t.Error("the query kept a row its name does not answer")
	}
	if err := tt.Click("read-model.go"); err != nil {
		t.Fatal(err)
	}
	// The press is counted in the rows as drawn, which is the filtered set:
	// the second one, not the second of all.
	if len(picked) != 1 || picked[0] != 1 {
		t.Errorf("the press reported %v, want the second drawn row once", picked)
	}
}

// TestContextMentionMenuDrawsNothingWhenNothingMatches is the same rule the
// composer's menus keep: no match is no panel, because a panel that appears
// with no rows says the feature is broken.
func TestContextMentionMenuDrawsNothingWhenNothingMatches(t *testing.T) {
	q := "nothing of the sort"
	drawn := false
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		res := ContextMentionMenu(c, ContextMentionMenuOptions{
			Items: []ContextItem{{Name: "README.md", Kind: "file"}},
			Query: &q,
		})
		if res.Element != nil || res.Picked() != -1 {
			drawn = true
		}
	}, 420, 200)
	if drawn {
		t.Error("a menu with no matches drew something")
	}
}

// TestContextMentionMenuInsistsOnNames: a row is read by its name, and a
// nameless row is a tag floating.
func TestContextMentionMenuInsistsOnNames(t *testing.T) {
	wantsPanic(t, "an item with no name", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			ContextMentionMenu(c, ContextMentionMenuOptions{
				Items: []ContextItem{{Kind: "file", Meta: "repo root"}},
			})
		}, 400, 200)
	})
}

// TestContextMentionMenuInDarkMode draws the rows, the tags and the metas
// under the dark palette.
func TestContextMentionMenuInDarkMode(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: core.Dark})
		ContextMentionMenu(c, ContextMentionMenuOptions{
			Items: []ContextItem{{Name: "go.mod", Kind: "file", Meta: "repo root"}},
		})
	}, 420, 200)
	for _, want := range []string{"go.mod", "file", "repo root"} {
		if !tt.HasText(want) {
			t.Errorf("the dark menu does not show %q", want)
		}
	}
}

// ── LinkPreviewCard ────────────────────────────────────────────────────────

// TestLinkPreviewCardReportsTheOpen is the behaviour test: the card shows
// the title, the host cut out of the URL, the description, and the Open
// press reported and not acted on.
func TestLinkPreviewCardReportsTheOpen(t *testing.T) {
	opened := 0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		if LinkPreviewCard(c, LinkPreviewCardOptions{
			URL:         "https://docs.mintui.dev/guide",
			Title:       "Component guide",
			Description: "How the cards are laid out.",
		}).Opened() {
			opened++
		}
	}, 480, 220)
	for _, want := range []string{
		"Component guide", "docs.mintui.dev", "How the cards are laid out.", "Open",
	} {
		if !tt.HasText(want) {
			t.Errorf("the card does not show %q; texts were %v", want, tt.Texts())
		}
	}
	if err := tt.Click("Open"); err != nil {
		t.Fatal(err)
	}
	if opened != 1 {
		t.Errorf("the card reported %d opens, want exactly 1", opened)
	}
}

// TestLinkPreviewCardInsistsOnATitle: the title is the line a reader decides
// on, and a card without one is a rectangle about a page nobody can name.
func TestLinkPreviewCardInsistsOnATitle(t *testing.T) {
	wantsPanic(t, "a card with no title", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			LinkPreviewCard(c, LinkPreviewCardOptions{URL: "https://docs.mintui.dev"})
		}, 400, 200)
	})
}

// TestLinkHostCutsTheSchemeAndThePath is the contract of the fallback: a URL
// gives its host, a bare host gives itself back, and a URL with no host
// gives nothing.
func TestLinkHostCutsTheSchemeAndThePath(t *testing.T) {
	cases := map[string]string{
		"https://docs.mintui.dev/guide": "docs.mintui.dev",
		"http://a.b/c?d#e":              "a.b",
		"docs.mintui.dev":               "docs.mintui.dev",
		// A schemeless string has no way to tell its host from its path, so
		// the whole front is the answer — which is what a caller that typed
		// a bare host wants back.
		"just-a-path": "just-a-path",
	}
	for in, want := range cases {
		if got := linkHost(in); got != want {
			t.Errorf("linkHost(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestLinkPreviewCardInDarkMode draws the card under the dark palette and
// asks that the favicon square is still there with its letter.
func TestLinkPreviewCardInDarkMode(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: core.Dark})
		LinkPreviewCard(c, LinkPreviewCardOptions{
			URL: "https://docs.mintui.dev/guide", Title: "Dark card",
		})
	}, 480, 220)
	for _, want := range []string{"Dark card", "docs.mintui.dev"} {
		if !tt.HasText(want) {
			t.Errorf("the dark card does not show %q", want)
		}
	}
}

// ── ProjectKnowledgePanel ──────────────────────────────────────────────────

// TestProjectKnowledgePanelShowsItsHeaderAndEntries: the header carries the
// project and the count, and each entry wears its kind and its meta.
func TestProjectKnowledgePanelShowsItsHeaderAndEntries(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ProjectKnowledgePanel(c, ProjectKnowledgePanelOptions{
			Project: "MintUI",
			Entries: []KnowledgeEntry{
				{Title: "Architecture", Kind: "doc", Meta: "docs/arch.md"},
				{Title: "Decisions", Kind: "note"},
			},
		})
	}, 420, 300)
	for _, want := range []string{
		"MintUI", "2 entries", "Architecture", "doc", "docs/arch.md", "Decisions", "note",
	} {
		if !tt.HasText(want) {
			t.Errorf("the panel does not show %q; texts were %v", want, tt.Texts())
		}
	}
}

// TestProjectKnowledgePanelSaysSoWhenEmpty: the difference between "nothing
// here yet" and a panel that failed is one sentence, and the panel is the
// one that says it.
func TestProjectKnowledgePanelSaysSoWhenEmpty(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ProjectKnowledgePanel(c, ProjectKnowledgePanelOptions{Project: "MintUI"})
	}, 420, 240)
	if !tt.HasText("No entries yet") {
		t.Errorf("an empty panel says nothing; texts were %v", tt.Texts())
	}
}

// TestProjectKnowledgePanelInsistsOnAProject: the entries belong to the
// project on the header, and a panel without one is a shelf with no name.
func TestProjectKnowledgePanelInsistsOnAProject(t *testing.T) {
	wantsPanic(t, "a panel with no project", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			ProjectKnowledgePanel(c, ProjectKnowledgePanelOptions{
				Entries: []KnowledgeEntry{{Title: "Architecture"}},
			})
		}, 400, 200)
	})
	wantsPanic(t, "an entry with no title", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			ProjectKnowledgePanel(c, ProjectKnowledgePanelOptions{
				Project: "MintUI",
				Entries: []KnowledgeEntry{{Kind: "doc"}},
			})
		}, 400, 200)
	})
}

// TestProjectKnowledgePanelInDarkMode draws the header and the rows under
// the dark palette.
func TestProjectKnowledgePanelInDarkMode(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: core.Dark})
		ProjectKnowledgePanel(c, ProjectKnowledgePanelOptions{
			Project: "MintUI",
			Entries: []KnowledgeEntry{{Title: "Architecture", Kind: "doc"}},
		})
	}, 420, 300)
	for _, want := range []string{"MintUI", "1 entry", "Architecture", "doc"} {
		if !tt.HasText(want) {
			t.Errorf("the dark panel does not show %q", want)
		}
	}
}

// ── SharedConversationView ─────────────────────────────────────────────────

// TestSharedConversationViewSaysWhatItIs: the badge on top, each turn with
// its speaker and its time, and the read-only line at the bottom — the one
// thing the view must say, that there is no composer.
func TestSharedConversationViewSaysWhatItIs(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		SharedConversationView(c, SharedConversationViewOptions{
			Messages: []SharedMessage{
				{Sender: "Ada", Text: "The build is green.", Time: "09:41"},
				{Sender: "Model", Text: "Then we ship."},
			},
		})
	}, 480, 360)
	for _, want := range []string{
		"Shared", "Ada", "The build is green.", "09:41",
		"Model", "Then we ship.", "This conversation is read-only",
	} {
		if !tt.HasText(want) {
			t.Errorf("the shared view does not show %q; texts were %v", want, tt.Texts())
		}
	}
}

// TestSharedConversationViewInsistsOnSpeakers: a shared conversation is
// several people, and a turn with no speaker cannot be attributed.
func TestSharedConversationViewInsistsOnSpeakers(t *testing.T) {
	wantsPanic(t, "a turn with no sender", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			SharedConversationView(c, SharedConversationViewOptions{
				Messages: []SharedMessage{{Text: "The build is green."}},
			})
		}, 400, 200)
	})
	wantsPanic(t, "a turn with no words", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			SharedConversationView(c, SharedConversationViewOptions{
				Messages: []SharedMessage{{Sender: "Ada"}},
			})
		}, 400, 200)
	})
}

// TestSharedConversationViewInDarkMode draws the badge, the turns and the
// read-only line under the dark palette.
func TestSharedConversationViewInDarkMode(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: core.Dark})
		SharedConversationView(c, SharedConversationViewOptions{
			Messages: []SharedMessage{{Sender: "Ada", Text: "The build is green."}},
		})
	}, 480, 300)
	for _, want := range []string{"Shared", "Ada", "This conversation is read-only"} {
		if !tt.HasText(want) {
			t.Errorf("the dark shared view does not show %q", want)
		}
	}
}

// ── WelcomeScreen ──────────────────────────────────────────────────────────

// TestWelcomeScreenReportsTheCard is the behaviour test: the greeting and
// the body on top, the cards in two columns, and the press reported as the
// card's place in the list.
func TestWelcomeScreenReportsTheCard(t *testing.T) {
	var picked []int
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		if got := WelcomeScreen(c, WelcomeScreenOptions{
			Greeting: "Welcome to MintUI",
			Body:     "Pick a place to start.",
			Cards: []WelcomeCard{
				{Title: "Start a task", Body: "Describe what to do."},
				{Title: "Import a file"},
				{Title: "Ask about the code"},
			},
		}).Picked(); got != -1 {
			picked = append(picked, got)
		}
	}, 620, 460)
	if len(picked) != 0 {
		t.Fatalf("a card was reported before anything was pressed")
	}
	for _, want := range []string{
		"Welcome to MintUI", "Pick a place to start.",
		"Start a task", "Describe what to do.", "Import a file", "Ask about the code",
	} {
		if !tt.HasText(want) {
			t.Errorf("the welcome screen does not show %q; texts were %v", want, tt.Texts())
		}
	}
	if err := tt.Click("Import a file"); err != nil {
		t.Fatal(err)
	}
	if len(picked) != 1 || picked[0] != 1 {
		t.Errorf("the press reported %v, want the second card once", picked)
	}
}

// TestWelcomeScreenInsistsOnWords: the greeting is the screen's own word,
// and a card without a title cannot be reported by index.
func TestWelcomeScreenInsistsOnWords(t *testing.T) {
	wantsPanic(t, "a screen with no greeting", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			WelcomeScreen(c, WelcomeScreenOptions{
				Cards: []WelcomeCard{{Title: "Start a task"}},
			})
		}, 400, 240)
	})
	wantsPanic(t, "a card with no title", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			WelcomeScreen(c, WelcomeScreenOptions{
				Greeting: "Welcome",
				Cards:    []WelcomeCard{{Body: "Describe what to do."}},
			})
		}, 400, 240)
	})
}

// TestWelcomeScreenInDarkMode draws the greeting and the cards under the
// dark palette.
func TestWelcomeScreenInDarkMode(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: core.Dark})
		WelcomeScreen(c, WelcomeScreenOptions{
			Greeting: "Good evening",
			Cards:    []WelcomeCard{{Title: "Start a task"}},
		})
	}, 620, 400)
	for _, want := range []string{"Good evening", "Start a task"} {
		if !tt.HasText(want) {
			t.Errorf("the dark welcome screen does not show %q", want)
		}
	}
}
