package chat

import (
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// The composer: the box a question is written in, and everything that hangs
// off it — the commands, the mentions, the tools, the parameters, the model,
// and the two things that can be dropped on it.

// PromptComposerOptions configure a PromptComposer.
type PromptComposerOptions struct {
	// Text is the draft. It is the caller's string throughout: the composer
	// never copies it, never truncates it and never clears it, because a
	// composer that clears the draft has thrown away a paragraph nobody can
	// get back.
	Text *string
	// Label names the field; required, as for every field.
	Label string
	// Placeholder shows while the draft is empty.
	Placeholder string
	// Lines is how tall the field is. Four is the usual answer.
	Lines int
	// Busy means an answer is streaming, which is what swaps the send button
	// for the stop one. It is the caller's flag because only the caller knows
	// whether the transport is still open.
	Busy bool
	// Send and Stop name the two buttons; empty takes the library's "Send"
	// and "Stop generating".
	Send, Stop string
	// Chips are the context chips above the field.
	Chips []ContextChip
	// Tools are the tools a person can turn on, and ToolsOn is which of them
	// are. A nil ToolsOn leaves the tool menu out entirely.
	Tools   []string
	ToolsOn *[]string
	// Selected is the row the menu keys move from, which is the caller's
	// because two menus in one window — the commands and the mentions — have
	// two of them.
	Selected *int
	// ShowVoice puts a voice button at the end of the toolbar.
	ShowVoice bool
	// Dragging says files are over the window, which is what DragDropOverlay
	// covers; the composer draws its own border to say the same thing where
	// the drop would land.
	Dragging bool
}

// PromptComposerResult carries a PromptComposer and everything the user did to
// it.
type PromptComposerResult struct {
	// Element is the whole composer.
	Element *ui.Element
	sent    bool
	stopped bool
	voiced  bool
	// chipRemoved is the 1-based index of the context chip that was removed.
	chipRemoved int
	// toolChosen is the tool whose toggle was pressed.
	toolChosen string
}

// Sent reports the send button being pressed this frame. It is false while
// Busy: there is nothing to send while an answer is arriving, and a caller
// that acted on it would be asking for a second request.
func (r PromptComposerResult) Sent() bool { return r.sent }

// Stopped reports the stop button being pressed this frame.
func (r PromptComposerResult) Stopped() bool { return r.stopped }

// Voiced reports the voice button being pressed.
func (r PromptComposerResult) Voiced() bool { return r.voiced }

// ChipRemoved is the 1-based index of the context chip removed this frame, 0
// for none.
func (r PromptComposerResult) ChipRemoved() int { return r.chipRemoved }

// ToolChosen is the tool whose toggle was pressed this frame, empty for none.
func (r PromptComposerResult) ToolChosen() string { return r.toolChosen }

// PromptComposer is the box a question is written in.
//
// It is one component rather than a field and three buttons so that the
// arrangement is stated once: the context chips above, the field in the middle,
// the tools and the send at the end. Every composer in every application then
// has the same shape, and the one thing a chat window must be consistent about
// is where the send button is.
func PromptComposer(c *ui.Context, opts PromptComposerOptions) PromptComposerResult {
	if opts.Text == nil {
		panic("chat: PromptComposer needs the draft to point at; it keeps none of its own")
	}
	if opts.Label == "" {
		panic("chat: PromptComposer needs a Label")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	send := opts.Send
	if send == "" {
		send = core.Msg(c, "chat.send", core.Def("Send"))
	}
	stop := opts.Stop
	if stop == "" {
		stop = core.Msg(c, "chat.stop", core.Def("Stop generating"))
	}
	empty := strings.TrimSpace(*opts.Text) == ""

	var res PromptComposerResult
	composer := ui.Column(c).FillWidth().Gap(u*1.5).Radius(theme.CardRadius).
		Background(k.Surface).Border(theme.BorderWidth, k.Border).Label(opts.Label)
	if opts.Dragging {
		// A second hairline in the accent, rather than a scrim over the whole
		// window: the drop lands here, so here is where the window has to say
		// it will be taken.
		composer = composer.BorderColor(k.Accent).BorderWidth(theme.BorderWidth * 2)
	}

	composer.Children(func() {
		if len(opts.Chips) > 0 {
			chips := ContextChips(c, opts.Chips)
			res.chipRemoved = chips.Removed()
		}

		input.TextArea(c, opts.Text, input.TextAreaOptions{
			Label: opts.Label, Placeholder: opts.Placeholder, Lines: opts.Lines,
		})

		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u).Children(func() {
			if len(opts.Tools) > 0 {
				// The switch has already written the caller's slice by the
				// time Toggled is read, so there is nothing left to do here
				// but say which one it was.
				res.toolChosen = ToolToggleMenu(c, ToolToggleMenuOptions{
					Tools: opts.Tools, On: opts.ToolsOn,
				}).Toggled()
			}
			ui.Box(c).Grow(1)
			if opts.ShowVoice {
				voice := VoiceInputButton(c, VoiceInputButtonOptions{
					Label: core.Msg(c, "chat.voice", core.Def("Voice input")),
				})
				if voice.Toggled() {
					res.voiced = true
				}
			}
			if opts.Busy {
				if StopGeneratingButton(c, StopGeneratingButtonOptions{Label: stop}).Stopped() {
					res.stopped = true
				}
			} else if SendButton(c, SendButtonOptions{
				Label: send, Disabled: empty,
			}).Sent() {
				res.sent = true
			}
		})
	})
	res.Element = composer
	return res
}

// command is one entry of the two pop-up menus the composer can raise: a
// slash command or a mention.
type command struct {
	// Name is what is chosen and what the reader sees.
	Name string
	// Hint is the line under it — what the command does, what the mention
	// refers to.
	Hint string
	// Detail is a third line, for a file's path or a command's arguments.
	Detail string
}

// SlashCommandOptions configure a SlashCommandMenu.
type SlashCommandOptions struct {
	// Commands are the commands on offer, each as "name: what it does".
	// Empty lines are dropped rather than drawn as gaps.
	Commands []string
	// Query is what has been typed after the slash. It is the caller's string
	// because the composer owns it.
	Query string
	// Selected is the row the keys move from, -1 for none.
	Selected *int
	// MaxHeight caps the panel; zero takes six rows.
	MaxHeight float32
	// Label names the menu; empty takes the library's "Commands".
	Label string
}

// SlashCommandResult carries a SlashCommandMenu and the command chosen.
type SlashCommandResult struct {
	// Element is the menu, nil when nothing matched.
	Element *ui.Element
	chosen  string
}

// Chosen is the command pressed this frame, empty when none was. It is the
// name without the slash: what comes back is something a caller can run, not
// the character the reader happened to type before it.
func (r SlashCommandResult) Chosen() string { return r.chosen }

// SlashCommandMenu is the panel of commands a "/" opens in the composer.
//
// It draws nothing when nothing matches, rather than drawing an empty panel:
// a panel that appears with no rows in it says the feature is broken, where a
// panel that does not appear says the typing did not match anything — which is
// the true answer.
func SlashCommandMenu(c *ui.Context, opts SlashCommandOptions) SlashCommandResult {
	list := make([]command, 0, len(opts.Commands))
	for _, line := range opts.Commands {
		if strings.TrimSpace(line) == "" {
			continue
		}
		name, hint, _ := strings.Cut(line, ":")
		list = append(list, command{Name: strings.TrimSpace(name), Hint: strings.TrimSpace(hint)})
	}

	height := opts.MaxHeight
	if height <= 0 {
		height = 6 * commandRowHeight(c)
	}
	name := opts.Label
	if name == "" {
		name = core.Msg(c, "chat.commands", core.Def("Commands"))
	}

	e, chosen := suggestionPanel(c, list, opts.Query, opts.Selected, height, name, "/")
	return SlashCommandResult{Element: e, chosen: chosen}
}

// MentionOptions configure a MentionMenu.
type MentionOptions struct {
	// Items are the things that can be mentioned, each as "name: where it
	// is".
	Items []string
	// Query is what has been typed after the "@".
	Query string
	// Selected is the row the keys move from, -1 for none.
	Selected *int
	// MaxHeight caps the panel; zero takes six rows.
	MaxHeight float32
	// Label names the menu; empty takes the library's "Mentions".
	Label string
}

// MentionResult carries a MentionMenu and the mention chosen.
type MentionResult struct {
	// Element is the menu, nil when nothing matched.
	Element *ui.Element
	chosen  string
}

// Chosen is the mention pressed this frame, empty when none was.
func (r MentionResult) Chosen() string { return r.chosen }

// MentionMenu is the panel an "@" opens in the composer.
//
// It is the same panel as the slash commands and is drawn by the same
// function, because they are the same control: a filtered list of things that
// can be put in the draft, with a row the keys move from. Two copies would
// differ in the width of the left column within a month.
func MentionMenu(c *ui.Context, opts MentionOptions) MentionResult {
	list := make([]command, 0, len(opts.Items))
	for _, line := range opts.Items {
		if strings.TrimSpace(line) == "" {
			continue
		}
		name, hint, _ := strings.Cut(line, ":")
		list = append(list, command{Name: strings.TrimSpace(name), Hint: strings.TrimSpace(hint)})
	}

	height := opts.MaxHeight
	if height <= 0 {
		height = 6 * commandRowHeight(c)
	}
	name := opts.Label
	if name == "" {
		name = core.Msg(c, "chat.mentions", core.Def("Mentions"))
	}

	e, chosen := suggestionPanel(c, list, opts.Query, opts.Selected, height, name, "@")
	return MentionResult{Element: e, chosen: chosen}
}

// commandRowHeight is how tall one row of either menu is, so a menu that
// defaults its height to six rows really is six rows.
func commandRowHeight(c *ui.Context) float32 {
	return core.FontSize(c, theme.RowSize) * 2.6
}

// suggestionPanel draws the one panel both menus use: a filtered list of rows,
// each a button, with the chosen one filled.
//
// It returns nil when nothing matches. The filtered list is the panel's own
// work and not the caller's: the two menus differ in what they search and
// what their trigger character is, and nothing else worth a second function.
func suggestionPanel(c *ui.Context, list []command, query string, selected *int,
	maxHeight float32, name, trigger string,
) (*ui.Element, string) {
	k, u := core.Tokens(c), core.Density(c).Unit()

	q := strings.ToLower(strings.TrimSpace(query))
	matches := list[:0:0]
	for _, cmd := range list {
		if q == "" || strings.Contains(strings.ToLower(cmd.Name), q) {
			matches = append(matches, cmd)
		}
	}
	if len(matches) == 0 {
		return nil, ""
	}

	var chosen string
	panel := func() {
		ui.Column(c).FillWidth().Gap(u * 0.25).Label(name).Children(func() {
			for i, cmd := range matches {
				lead := selected != nil && *selected == i
				row := ui.ButtonBase(c).FillWidth().Justify(ui.Start).AlignItems(ui.Center).
					Gap(u*1.5).Padding(u*0.75, u*1.5).
					Radius(theme.SmallRadius).Label(cmd.Name).
					TextColor(k.Text).Key(menuRowKey(trigger, i))
				// Only the chosen row is filled, for the reason a filter list is:
				// a column of pills reads as a stack of tabs rather than as a list
				// with one place in it.
				if lead {
					row = row.Background(k.SurfaceHover)
				} else if row.Hovered() {
					row = row.Background(k.Surface)
				}
				if row.Clicked() {
					chosen = cmd.Name
				}
				row.Children(func() {
					// The trigger is shown on every row rather than only the first:
					// it is part of what the row is, and a menu of "@" items whose
					// first row has an @ and the rest do not reads as a typo.
					ui.Text(c, trigger).TextColor(k.TextFaint).
						FontSize(core.FontSize(c, theme.RowSize)).Shrink(0)
					ui.Column(c).Grow(1).AlignItems(ui.Start).Children(func() {
						ui.Text(c, cmd.Name).TextColor(k.Text).
							FontSize(core.FontSize(c, theme.RowSize)).SingleLine()
						if cmd.Hint != "" || cmd.Detail != "" {
							ui.Text(c, strings.TrimSpace(cmd.Hint+" "+cmd.Detail)).
								TextColor(k.TextMuted).
								FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
						}
					})
				})
			}
		})
	}

	// The scroll is built here rather than by the caller because a child is
	// whatever was CREATED while Children was running: a scroll made out here
	// and handed in would be the panel's sibling, not its container.
	scrolled := ui.Scroll(c).MaxHeight(maxHeight).FillWidth()
	scrolled.Children(panel)
	return scrolled, chosen
}

// menuRowKey keeps the chosen row and the keyboard position attached to their
// own row across redraws. It is keyed on the trigger because the two menus
// can be open at once — "/" and "@" in the same draft — and two rows with the
// same key under the same parent is a duplicate, which MyGo treats as a bug.
func menuRowKey(trigger string, i int) string { return trigger + ":" + itoa(i) }

// ToolToggleMenuOptions configure a ToolToggleMenu.
type ToolToggleMenuOptions struct {
	// Tools are the tools on offer, in the order they are shown.
	Tools []string
	// On is the caller's set of the tools that are on, and the switches write
	// it directly — the same contract as a checkbox, for the same reason.
	On *[]string
	// Label names the menu; empty takes the library's "Tools".
	Label string
}

// ToolToggleMenuResult carries a ToolToggleMenu and the tool toggled.
type ToolToggleMenuResult struct {
	// Element is the menu.
	Element *ui.Element
	toggled string
}

// Toggled is the tool whose switch was pressed this frame, empty for none.
// The switch has already written the caller's slice by the time this is read,
// so this is a question about which one, not a handover.
func (r ToolToggleMenuResult) Toggled() string { return r.toggled }

// ToolToggleMenu is the panel of switches that turns tools on and off.
//
// It is switches and not checkboxes because a tool is not a value in a form:
// there is no "submit", and a switch that is already in its on position tells
// the reader so without a separate chosen state.
func ToolToggleMenu(c *ui.Context, opts ToolToggleMenuOptions) ToolToggleMenuResult {
	if opts.On == nil {
		panic("chat: ToolToggleMenu needs the set of tools that are on to point at; " +
			"it keeps none of its own")
	}
	if len(opts.Tools) == 0 {
		panic("chat: ToolToggleMenu needs at least one tool")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	name := opts.Label
	if name == "" {
		name = core.Msg(c, "chat.tools", core.Def("Tools"))
	}

	var res ToolToggleMenuResult
	// Each switch gets its own bool, so the row can be rebuilt from the
	// caller's slice every frame without the switch keeping state of its own.
	states := make([]bool, len(opts.Tools))
	for i, tool := range opts.Tools {
		states[i] = contains(*opts.On, tool)
	}

	panel := ui.Column(c).FillWidth().Gap(u).Radius(theme.ControlRadius).
		Background(k.Background).Border(theme.BorderWidth, k.Border).
		Padding(u*1.5, u*2).Label(name)
	panel.Children(func() {
		ui.Text(c, name).TextColor(k.TextMuted).
			FontSize(core.FontSize(c, theme.CaptionSize))
		for i, tool := range opts.Tools {
			row := ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u).
				Padding(u*0.5, 0)
			row.Children(func() {
				sw := input.Switch(c, &states[i], input.SwitchOptions{Label: tool})
				if sw.Clicked() {
					res.toggled = tool
				}
				// The switch flips its own bool the moment it is pressed, so
				// the caller's slice is brought into step right after. It is
				// written by name rather than by position: a tools list is
				// reordered as often as it is edited, and a slice written by
				// index would move a setting from one tool to another.
				if states[i] != contains(*opts.On, tool) {
					*opts.On = toggleIn(*opts.On, tool, states[i])
				}
			})
		}
	})
	res.Element = panel
	return res
}

// toggleIn adds or takes out a name from a set, keeping the order of what is
// left.
func toggleIn(set []string, name string, on bool) []string {
	out := make([]string, 0, len(set)+1)
	found := false
	for _, s := range set {
		if s == name {
			found = true
			if on {
				out = append(out, name)
			}
			continue
		}
		out = append(out, s)
	}
	if on && !found {
		out = append(out, name)
	}
	return out
}

func contains(set []string, name string) bool {
	for _, s := range set {
		if s == name {
			return true
		}
	}
	return false
}

// SuggestionChipsOptions configure SuggestionChips.
type SuggestionChipsOptions struct {
	// Suggestions are the things on offer, each as "label: what it does".
	Suggestions []string
	// Title heads the row; empty draws no heading.
	Title string
	// Max is how many to show. Zero shows all of them; a composer that offers
	// twelve suggestions shows them all and lets them wrap.
	Max int
}

// SuggestionChipsResult carries SuggestionChips and the suggestion chosen.
type SuggestionChipsResult struct {
	// Element is the row.
	Element *ui.Element
	chosen  string
}

// Chosen is the suggestion pressed this frame, empty for none.
func (r SuggestionChipsResult) Chosen() string { return r.chosen }

// SuggestionChips is the row of things a reader can send instead of typing.
//
// It is a wrap rather than a scroll, for the reason context chips are: a
// suggestion the reader cannot see is not a suggestion. It goes above the
// field and not in it, because these are whole questions and a chip inside the
// draft is text the reader has to delete.
func SuggestionChips(c *ui.Context, opts SuggestionChipsOptions) SuggestionChipsResult {
	if len(opts.Suggestions) == 0 {
		panic("chat: SuggestionChips needs at least one suggestion")
	}
	tokens := core.Tokens(c)
	step := core.Density(c).Unit()

	shown := opts.Suggestions
	if opts.Max > 0 && len(shown) > opts.Max {
		shown = shown[:opts.Max]
	}

	var res SuggestionChipsResult
	row := ui.Column(c).FillWidth().Gap(step).Children(func() {
		if opts.Title != "" {
			ui.Text(c, opts.Title).TextColor(tokens.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize))
		}
		ui.Row(c).FillWidth().Gap(step * 0.75).Children(func() {
			for _, line := range shown {
				label, hint, _ := strings.Cut(line, ":")
				label, hint = strings.TrimSpace(label), strings.TrimSpace(hint)
				text := label
				if hint != "" {
					text += " — " + hint
				}
				chip := ui.ButtonBase(c).Radius(theme.PillRadius).
					Padding(step*0.75, step*2).Label(label).
					TextColor(tokens.Text).Children(func() {
					ui.Text(c, text).TextColor(tokens.Text).
						FontSize(core.FontSize(c, theme.RowSize)).SingleLine()
				})
				if chip.Clicked() {
					res.chosen = label
				}
			}
		})
	})
	res.Element = row
	return res
}

// PromptLibraryOptions configure a PromptLibrary.
type PromptLibraryOptions struct {
	// Title heads the library; empty draws no heading.
	Title string
	// Prompts are the saved prompts, each as "name: the prompt itself".
	Prompts []string
	// Selected is the row the keys move from, -1 for none.
	Selected *int
	// Height is the list's height; it is required for the reason every list
	// in this library requires one.
	Height float32
	// Query filters the names, ignoring case.
	Query string
	// Empty draws instead of the rows when nothing matches.
	Empty func()
}

// PromptLibraryResult carries a PromptLibrary and what the user did with it.
type PromptLibraryResult struct {
	// Element is the library.
	Element *ui.Element
	used    int
	removed int
}

// Used is the 1-based index of the prompt pressed this frame, 0 for none.
func (r PromptLibraryResult) Used() int { return r.used }

// Removed is the 1-based index of the prompt whose remove button was pressed.
func (r PromptLibraryResult) Removed() int { return r.removed }

// PromptLibrary is the list of prompts a person has saved to reuse.
//
// It reports the index and not the text, because the caller already has the
// text: it handed it over in Prompts. Handing the string back would be a
// second copy of the same words that could drift from the list.
func PromptLibrary(c *ui.Context, opts PromptLibraryOptions) PromptLibraryResult {
	if opts.Height <= 0 {
		panic("chat: PromptLibrary needs a Height; a list with no height is every row " +
			"drawn rather than a list")
	}
	if len(opts.Prompts) == 0 {
		panic("chat: PromptLibrary needs at least one prompt")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	type saved struct{ name, text string }
	var all []saved
	for _, line := range opts.Prompts {
		name, body, ok := strings.Cut(line, ":")
		name = strings.TrimSpace(name)
		if !ok || name == "" {
			// A line with no colon is a prompt with no name of its own, and a
			// row with no name cannot be chosen — so the line is dropped
			// rather than shown as a row with a blank on it.
			continue
		}
		all = append(all, saved{name: name, text: strings.TrimSpace(body)})
	}
	if len(all) == 0 {
		panic("chat: PromptLibrary has no prompts with names; each line is " +
			"\"name: the prompt itself\"")
	}

	q := strings.ToLower(strings.TrimSpace(opts.Query))
	var res PromptLibraryResult
	var shown int
	list := func() {
		ui.Column(c).FillWidth().Gap(u * 0.25).Children(func() {
			shown = 0
			for i, p := range all {
				if q != "" && !strings.Contains(strings.ToLower(p.name), q) {
					continue
				}
				shown++
				index := i
				lead := opts.Selected != nil && *opts.Selected == i
				row := ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u).
					Padding(u*0.75, u*1.5).Radius(theme.SmallRadius).
					Label(p.name).Cursor(ui.CursorPointer)
				if lead {
					row = row.Background(k.SurfaceHover)
				} else if row.Hovered() {
					row = row.Background(k.Surface)
				}
				if row.Clicked() {
					res.used = index + 1
				}
				row.Children(func() {
					ui.Column(c).Grow(1).AlignItems(ui.Start).Children(func() {
						ui.Text(c, p.name).TextColor(k.Text).
							FontSize(core.FontSize(c, theme.RowSize)).SingleLine()
						ui.Text(c, oneLine(p.text)).TextColor(k.TextMuted).
							FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
					})
					cross := ui.Box(c).Size(u*2.5, u*2.5).Shrink(0).Role(ui.RoleNone).
						Label("Remove " + p.name).
						Draw(func(p *ui.Painter, r ui.Rect) { markCross(p, r, k.TextFaint) })
					if cross.Clicked() {
						res.removed = index + 1
					}
				})
			}
			if shown == 0 {
				if opts.Empty != nil {
					opts.Empty()
				} else {
					ui.Text(c, core.Msg(c, "chat.prompts.none", core.Def("No saved prompts"))).
						TextColor(k.TextFaint).FontSize(core.FontSize(c, theme.CaptionSize))
				}
			}
		})
	}

	panel := ui.Column(c).FillWidth().Gap(u * 1.5)
	if opts.Title != "" {
		panel.Label(opts.Title)
	}
	panel.Children(func() {
		if opts.Title != "" {
			ui.Text(c, opts.Title).TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize))
		}
		// The scroll is made in here rather than out there: a child is whatever
		// was CREATED while Children was running, so one built outside would
		// land as the panel's sibling and not inside it.
		ui.Scroll(c).Height(opts.Height).FillWidth().Children(list)
	})
	res.Element = panel
	return res
}

// SystemPromptEditorOptions configure a SystemPromptEditor.
type SystemPromptEditorOptions struct {
	// Prompt is the caller's string, edited in place.
	Prompt *string
	// Label names the field; required.
	Label string
	// Title heads the editor.
	Title string
	// Reset and Save name the two buttons; empty takes the library's "Reset"
	// and "Save".
	Reset, Save string
	// Original is what Reset goes back to. Empty takes an empty prompt,
	// which is the only sensible default: a reset with nothing to go back to
	// would be a button that does nothing and says so by clearing the field.
	Original string
	// Lines is how tall the field is. Eight is the usual answer for a system
	// prompt, which is several paragraphs rather than a note.
	Lines int
}

// SystemPromptEditorResult carries a SystemPromptEditor and what was done to it.
type SystemPromptEditorResult struct {
	// Element is the editor.
	Element *ui.Element
	saved   bool
	reset   bool
}

// Saved reports the save button being pressed this frame.
func (r SystemPromptEditorResult) Saved() bool { return r.saved }

// Reset reports the reset button being pressed this frame. The field has
// already been put back by the time this is read, so a caller that wants to
// know whether the reset changed anything compares Prompt to Original itself.
func (r SystemPromptEditorResult) Reset() bool { return r.reset }

// SystemPromptEditor is the field a whole system prompt is written in, with a
// way back to the one it started as.
//
// Reset is here rather than left to the caller because the field is the
// caller's string: a button that put the value back from outside the frame
// would have to write to a string the field might be holding a selection in,
// and the frame is the only place that can.
func SystemPromptEditor(c *ui.Context, opts SystemPromptEditorOptions) SystemPromptEditorResult {
	if opts.Prompt == nil {
		panic("chat: SystemPromptEditor needs the prompt to point at; it keeps none " +
			"of its own")
	}
	if opts.Label == "" {
		panic("chat: SystemPromptEditor needs a Label; an editor with no name is a box")
	}
	u := core.Density(c).Unit()

	reset := opts.Reset
	if reset == "" {
		reset = core.Msg(c, "chat.prompt.reset", core.Def("Reset"))
	}
	save := opts.Save
	if save == "" {
		save = core.Msg(c, "chat.prompt.save", core.Def("Save"))
	}

	var res SystemPromptEditorResult
	panel := ui.Column(c).FillWidth().Gap(u * 1.5)
	if opts.Title != "" {
		panel.Label(opts.Title)
	}
	panel.Children(func() {
		if opts.Title != "" {
			ui.Text(c, opts.Title).TextColor(core.Tokens(c).Text).
				FontSize(core.FontSize(c, theme.RowSize))
		}
		input.TextArea(c, opts.Prompt, input.TextAreaOptions{
			Label: opts.Label, Lines: opts.Lines,
		})
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u).Children(func() {
			if input.Button(c, reset, input.ButtonOptions{}).Clicked() {
				*opts.Prompt = opts.Original
				res.reset = true
			}
			ui.Box(c).Grow(1)
			if input.Button(c, save, input.ButtonOptions{Primary: true}).Clicked() {
				res.saved = true
			}
		})
	})
	res.Element = panel
	return res
}

// Parameter is one row of a ParameterPanel: a name and the control that sets
// it.
type Parameter struct {
	// Name is what the parameter is called. Required: a control with no name
	// beside it in a panel of named controls is a mystery.
	Name string
	// Hint is the line under the name — what the parameter does, what a high
	// value costs.
	Hint string
	// Control builds the control, and is called every frame so that a control
	// which writes to the caller's number is up to date.
	Control func()
	// Reset puts the parameter back to its default, drawn as a small button
	// beside the name. nil draws none.
	Reset func()
}

// ParameterPanelOptions configure a ParameterPanel.
type ParameterPanelOptions struct {
	// Title heads the panel; empty draws no heading.
	Title string
	// Parameters are the rows, in the order they are shown.
	Parameters []Parameter
	// ResetAll names a button that puts every parameter back; empty draws
	// none.
	ResetAll string
}

// ParameterPanelResult carries a ParameterPanel and its reset-all press.
type ParameterPanelResult struct {
	// Element is the panel.
	Element  *ui.Element
	resetAll bool
}

// ResetAll reports the reset-all button being pressed this frame. The rows'
// own Reset functions are the caller's, so this does not claim to have reset
// anything by itself.
func (r ParameterPanelResult) ResetAll() bool { return r.resetAll }

// ParameterPanel is the column of named controls that set how a model answers.
//
// The controls are the caller's and only the rows are here, because a parameter
// panel's shape is the same whatever the parameters are — a name on the left,
// a control on the right, and the hint under the name — and every kind of
// control already exists in ui/input. Writing four more here would be four
// controls that behave slightly differently from the ones in a settings pane.
func ParameterPanel(c *ui.Context, opts ParameterPanelOptions) ParameterPanelResult {
	if len(opts.Parameters) == 0 {
		panic("chat: ParameterPanel needs at least one parameter")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	var res ParameterPanelResult
	panel := ui.Column(c).FillWidth().Gap(u * 1.5)
	if opts.Title != "" {
		panel.Label(opts.Title)
	}

	panel.Children(func() {
		if opts.Title != "" {
			ui.Row(c).FillWidth().AlignItems(ui.Center).Children(func() {
				ui.Text(c, opts.Title).TextColor(k.Text).
					FontSize(core.FontSize(c, theme.RowSize)).Grow(1)
				if opts.ResetAll != "" {
					if input.Button(c, opts.ResetAll, input.ButtonOptions{}).Clicked() {
						res.resetAll = true
					}
				}
			})
			layout.Divider(c, layout.DividerOptions{})
		}
		for i, p := range opts.Parameters {
			if p.Name == "" {
				panic("chat: ParameterPanel row " + internal.Commas(i+1) + " needs a Name")
			}
			if p.Control == nil {
				panic("chat: ParameterPanel row " + internal.Commas(i+1) + " (" + p.Name +
					") needs a Control")
			}
			ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u*2).
				Padding(u*0.5, 0).Children(func() {
				ui.Column(c).Grow(1).AlignItems(ui.Start).Children(func() {
					ui.Row(c).AlignItems(ui.Center).Gap(u).Children(func() {
						ui.Text(c, p.Name).TextColor(k.Text).
							FontSize(core.FontSize(c, theme.RowSize))
						if p.Reset != nil {
							if ui.ButtonBase(c).Label("Reset " + p.Name).
								TextColor(k.TextMuted).Children(func() {
								ui.Text(c, "Reset").TextColor(k.TextMuted).
									FontSize(core.FontSize(c, theme.CaptionSize))
							}).Clicked() {
								p.Reset()
							}
						}
					})
					if p.Hint != "" {
						ui.Text(c, p.Hint).TextColor(k.TextMuted).
							FontSize(core.FontSize(c, theme.CaptionSize))
					}
				})
				// The control is given the room it asks for and no more, so a
				// number input and a switch sit on the same edge of the row
				// instead of the switch floating in the middle of a slider's
				// space.
				ui.Box(c).Shrink(0).AlignItems(ui.End).Children(p.Control)
			})
		}
	})
	res.Element = panel
	return res
}

// ModelSelectorOptions configure a ModelSelector.
type ModelSelectorOptions struct {
	// Selected is the caller's model id. It is a string rather than an index
	// because a model id is what a request carries and an index into a list
	// that changes with every release is not.
	Selected *string
	// Models are the ids, each as "id: what it is called". An entry with no
	// colon is dropped: a model whose row is blank cannot be chosen.
	Models []string
	// Label names the control; empty takes the library's "Model".
	Label string
	// Placeholder says what the trigger shows before a model is chosen.
	Placeholder string
	// Disabled greys the control out, for a plan that has one model.
	Disabled bool
}

// ModelSelector is the control that says which model is answering.
//
// It is ui/input's Select rather than a fourth dropdown: the whole keyboard —
// Down to open, arrows to move, Enter to choose, Escape to close — comes from
// that package's SelectBase, and a chat window's model picker that behaved
// slightly differently from every other picker in the app would be found.
func ModelSelector(c *ui.Context, opts ModelSelectorOptions) *ui.Element {
	if opts.Selected == nil {
		panic("chat: ModelSelector needs the selected model to point at")
	}
	choices := make([]input.Choice, 0, len(opts.Models))
	for _, line := range opts.Models {
		id, name, ok := strings.Cut(line, ":")
		id, name = strings.TrimSpace(id), strings.TrimSpace(name)
		if !ok || id == "" {
			continue
		}
		if name == "" {
			name = id
		}
		choices = append(choices, input.Choice{Value: id, Label: name})
	}
	if len(choices) == 0 {
		panic("chat: ModelSelector needs at least one model with an id; each entry " +
			"is \"id: what it is called\"")
	}

	name := opts.Label
	if name == "" {
		name = core.Msg(c, "chat.model", core.Def("Model"))
	}
	placeholder := opts.Placeholder
	if placeholder == "" {
		placeholder = core.Msg(c, "chat.model.choose", core.Def("Choose a model"))
	}
	return input.Select(c, opts.Selected, choices, input.SelectOptions{
		Label: name, Placeholder: placeholder, Disabled: opts.Disabled,
	})
}

// ModeSelectorOptions configure a ModeSelector.
type ModeSelectorOptions struct {
	// Selected is the caller's index into Modes; it is required and must be
	// in range, because a mode control that is out of range is showing a mode
	// the caller did not ask for.
	Selected *int
	// Modes are the mode names, left to right.
	Modes []string
	// Label names the control for assistive technology; empty leaves it
	// unnamed, which is only right when the names say everything.
	Label string
}

// ModeSelector is the segmented control that says what kind of answer is
// wanted — ask, edit, agent — which is why it is a segmented control rather
// than a dropdown: the modes are few, they are all worth seeing at once, and
// switching between two of them is the most common thing a person does here.
func ModeSelector(c *ui.Context, opts ModeSelectorOptions) *ui.Element {
	if opts.Selected == nil {
		panic("chat: ModeSelector needs the selected mode to point at")
	}
	if len(opts.Modes) == 0 {
		panic("chat: ModeSelector needs at least one mode")
	}
	if *opts.Selected < 0 || *opts.Selected >= len(opts.Modes) {
		panic("chat: ModeSelector selection is out of range")
	}
	row := input.Segmented(c, opts.Selected, opts.Modes...)
	if opts.Label != "" {
		row = row.Label(opts.Label)
	}
	return row
}

// VoiceInputButtonResult carries a VoiceInputButton and its press.
type VoiceInputButtonResult struct {
	// Element is the button.
	Element *ui.Element
	toggled bool
}

// Toggled reports the button being pressed this frame. The caller flips its own
// flag; this says the press happened.
func (r VoiceInputButtonResult) Toggled() bool { return r.toggled }

// VoiceInputButtonOptions configure a VoiceInputButton.
type VoiceInputButtonOptions struct {
	// Label names the button; required, since it is a mark and not a word.
	Label string
	// Listening says the microphone is open, which turns the button into the
	// thing that closes it — the same button, because the same press stops
	// and starts and the reader should not have to find a second one.
	Listening bool
	// Levels are the levels to draw beside the button; nil draws none.
	Levels []float32
}

// VoiceInputButton is the microphone at the end of the composer.
//
// It is one button in two states rather than two buttons, because the press
// that starts recording is the press that stops it and the reader should not
// have to look for a different control to stop what they just started.
func VoiceInputButton(c *ui.Context, opts VoiceInputButtonOptions) VoiceInputButtonResult {
	if opts.Label == "" {
		panic("chat: VoiceInputButton needs a Label; a microphone with no name is a shape")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	var res VoiceInputButtonResult
	text := opts.Label
	if opts.Listening {
		text = core.Msg(c, "chat.voice.stop", core.Def("Stop dictation"))
	}

	col := ui.Column(c).Shrink(0).Gap(u * 0.5).Label(opts.Label)
	col.Children(func() {
		btn := ui.ButtonBase(c).Size(u*9, u*9).Radius(theme.PillRadius).
			Background(k.Surface).Border(theme.BorderWidth, k.Border).
			Label(text).Tooltip(text).TextColor(k.Text)
		if btn.Clicked() {
			res.toggled = true
		}
		btn.Children(func() {
			// The mark is drawn inside the button's Children, because an
			// element belongs to whatever was being built when it was made.
			micMark(c, k.Text, opts.Listening)
		})
		if opts.Listening && len(opts.Levels) > 0 {
			VoiceWaveform(c, VoiceWaveformOptions{
				Levels: opts.Levels, Bars: len(opts.Levels),
				Width: u * 14, Height: u * 2.5, Color: k.Danger,
				Label: text,
			})
		}
	})

	res.Element = col
	return res
}

// micMark is the microphone glyph, drawn rather than typed for the reason
// every mark in this library is drawn.
func micMark(c *ui.Context, col ui.Color, live bool) {
	side := core.Density(c).Unit() * 4.5
	ui.Box(c).Size(side, side).Shrink(0).Role(ui.RoleNone).Draw(func(p *ui.Painter, r ui.Rect) {
		cx := r.X + r.W/2
		var path ui.Path
		path.MoveTo(cx, r.Y+r.H*0.12).
			LineTo(cx, r.Y+r.H*0.5)
		p.StrokePath(&path, 1.5, col)
		var stand ui.Path
		stand.MoveTo(r.X+r.W*0.28, r.Y+r.H*0.42).
			CubeTo(r.X+r.W*0.28, r.Y+r.H*0.86, r.X+r.W*0.72, r.Y+r.H*0.86, r.X+r.W*0.72, r.Y+r.H*0.42)
		p.StrokePath(&stand, 1.5, col)
		if live {
			// A live mark has a ring round it, so the button's state is
			// legible without the tooltip and at the size the button is.
			p.Stroke(r, col.Alpha(0.35), r.W/2, 1.2)
		}
	})
}

// PasteImagePreviewResult carries a PasteImagePreview and the removal.
type PasteImagePreviewResult struct {
	// Element is the strip.
	Element *ui.Element
	removed int
}

// Removed is the 1-based index of the image removed this frame, 0 for none.
func (r PasteImagePreviewResult) Removed() int { return r.removed }

// PasteImagePreview is the strip of images waiting to be sent, shown as soon
// as the first one is pasted.
//
// It appears rather than waits for the file to finish reading, because a paste
// that shows nothing until it has finished reads as a paste that did not
// happen, and a reader who presses it again sends the image twice.
func PasteImagePreview(c *ui.Context, opts PasteImagePreviewOptions) PasteImagePreviewResult {
	if len(opts.Images) == 0 {
		panic("chat: PasteImagePreview needs at least one image; an empty strip is a gap")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	var res PasteImagePreviewResult
	strip := ui.Row(c).FillWidth().Gap(u).Label(core.Msg(c, "chat.pasted", core.Def("Pasted images")))
	strip.Children(func() {
		for i, img := range opts.Images {
			if img.Alt == "" {
				panic("chat: PasteImagePreview image " + internal.Commas(i+1) + " needs Alt")
			}
			index := i + 1
			tile := ui.Box(c).Size(u*15, u*15).Radius(theme.SmallRadius).
				Background(k.Surface).Border(theme.BorderWidth, k.Border).
				Clip().Label(img.Alt)
			tile.Children(func() {
				if img.Draw != nil {
					img.Draw()
					return
				}
				ui.Text(c, img.Alt).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
			})
			// The cross sits beside the tile rather than over it. An element
			// over a picture would have to be absolutely placed inside a
			// clipped box, and the clip that keeps the image inside its tile
			// would keep the cross there too — so the button would be cut off
			// by the very corner it was placed in.
			cross := ui.ButtonBase(c).Size(u*3.5, u*3.5).Radius(u * 1.75).
				Background(k.Fill).TextColor(k.OnFill).
				Label("Remove " + img.Alt).Tooltip("Remove " + img.Alt)
			if cross.Clicked() {
				res.removed = index
			}
			cross.Children(func() {
				ui.Text(c, "×").TextColor(k.OnFill).
					FontSize(core.FontSize(c, theme.RowSize))
			})
		}
	})
	res.Element = strip
	return res
}

// PasteImagePreviewOptions configure a PasteImagePreview.
type PasteImagePreviewOptions struct {
	// Images are the images waiting to be sent.
	Images []ImageItem
}

// DragDropOverlayResult carries a DragDropOverlay and whether anything was
// dropped on it.
type DragDropOverlayResult struct {
	// Element is the overlay, nil when there is nothing being dragged.
	Element *ui.Element
	dropped bool
}

// Dropped reports files being dropped this frame.
func (r DragDropOverlayResult) Dropped() bool { return r.dropped }

// DragDropOverlay is the sheet that covers a chat window while files are over
// it.
//
// It is drawn from Hovered rather than from a flag: the window knows files are
// over it by being hovered on, and a flag the caller has to set on the way in
// and clear on the way out is a flag that will be left set.
func DragDropOverlay(c *ui.Context, opts DragDropOverlayOptions) DragDropOverlayResult {
	if opts.Host == nil {
		panic("chat: DragDropOverlay needs the Host it is drawn over; an overlay that " +
			"cannot tell what it is over would cover the whole window")
	}
	label := opts.Label
	if label == "" {
		label = core.Msg(c, "chat.drop", core.Def("Drop files to attach"))
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	var res DragDropOverlayResult
	sheet := ui.Box(c).Fill().Label(label)
	if opts.Host.Hovered() {
		res.Element = sheet.Children(func() {
			ui.Box(c).Fill().Background(k.Accent.Alpha(0.12)).Radius(theme.CardRadius).
				Border(theme.BorderWidth*2, k.Accent).Center().Children(func() {
				ui.Column(c).AlignItems(ui.Center).Gap(u * 1.5).Children(func() {
					display.Icon(c, display.IconDownload, display.IconOptions{
						Name: label, Size: u * 9,
					})
					ui.Text(c, label).TextColor(k.Text).
						FontSize(core.FontSize(c, theme.RowSize))
				})
			})
		})
	}
	return res
}

// DragDropOverlayOptions configure a DragDropOverlay.
type DragDropOverlayOptions struct {
	// Host is the element the files are being dragged over. It is required:
	// an overlay that cannot tell what it is over would be drawn over the
	// whole window by every window that had one.
	Host *ui.Element
	// Label says what a drop will do.
	Label string
}
