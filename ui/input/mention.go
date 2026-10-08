package input

import (
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// MentionInput is a text field that offers names when you type @, and puts
// the one you take in as a plain word.
//
// The value is the caller's string and stays plain: a mention is written as
// "@sam" and read back as "@sam", not as a reference to something in a table
// that may be gone by the time the message is read. That is a deliberate
// choice against a component that holds its own mention objects — it would
// mean a message whose text cannot be sent without the widget that wrote it,
// and a draft that has outlived its own editor.

// MentionInputOptions configure a MentionInput.
type MentionInputOptions struct {
	// Label names the field; it is required, as for every field here.
	Label string
	// Placeholder shows while the value is empty. Empty takes the library's
	// own "Write a message", which is what a mention field usually is
	// beside.
	Placeholder string
	// Error marks the value as not valid.
	Error string
	// Lines is how many lines the field shows; zero is four, as TextArea's is.
	Lines int
	// People are who can be mentioned. Every one needs a Name, because the
	// name is what is written into the field: a mention of somebody the
	// field cannot spell is a mention nobody can reply to.
	People []string
	// MaxSuggestions caps what the panel offers. Zero is eight, which is more
	// than fits under a caret and few enough that the panel is not the field.
	MaxSuggestions int
	// Disabled greys the field out.
	Disabled bool
	// Width is the field's own width; zero fills the row it is in.
	Width float32
}

// MentionResult carries a MentionInput and what happened to it.
type MentionResult struct {
	// Element is the field.
	Element *ui.Element
	// changed, submitted and mentioned are this frame's answers.
	changed, submitted bool
	// mentioned is the name taken this frame, "" for none.
	mentioned string
}

// Changed reports that the value changed this frame, typed or inserted.
func (r MentionResult) Changed() bool { return r.changed }

// Submitted reports Enter with the panel shut, which is a message going rather
// than a mention being taken.
func (r MentionResult) Submitted() bool { return r.submitted }

// Mentioned is the name taken from the panel this frame, "" on every other
// frame. It is the name as the field spells it, "@sam", because that is what
// the caller's string now holds and what it would have to match on later.
func (r MentionResult) Mentioned() string { return r.mentioned }

// MentionInput is a field for writing a message that can talk about people.
//
// The panel opens on an @ and closes when the word after it stops matching
// anybody, which is the rule that makes the panel usable: it is a
// continuation of the field rather than a window over it, so there is no way
// to leave it somewhere you cannot get back from.
//
// Taking a suggestion replaces the half-word the user was typing with the
// whole name, so what they typed first is kept and only the tail is
// completed. Replacing the whole token would throw away "@sa" and make the
// field fight the person typing in it.
func MentionInput(c *ui.Context, value *string, opts MentionInputOptions) MentionResult {
	if value == nil {
		panic("input: MentionInput needs a value to point at; it keeps no value of its own")
	}
	if opts.Label == "" {
		panic("input: MentionInput needs a Label, or nothing can name the field")
	}
	for i, p := range opts.People {
		if strings.TrimSpace(p) == "" {
			panic("input: MentionInput person " + itoa(i) + " has no name; a mention of somebody " +
				"the field cannot spell is a mention nobody can reply to")
		}
	}
	u := core.Density(c).Unit()

	lines := opts.Lines
	if lines < 1 {
		lines = 4
	}
	limit := opts.MaxSuggestions
	if limit <= 0 {
		limit = 8
	}
	s := fieldSkin{
		label: opts.Label, err: opts.Error, lines: lines,
		disabled: opts.Disabled, width: opts.Width,
	}

	var r MentionResult
	r.Element = areaWell(c, s, func(w *ui.Element) *ui.Element {
		in := ui.TextAreaBase(c, value).Label(opts.Label).FillWidth().
			Height(float32(lines) * lineHeight(c))
		if opts.Placeholder != "" {
			in.Placeholder(opts.Placeholder)
		}
		dressInput(in, opts.Error, false)

		// The panel's open state and which suggestion is on are the field's
		// own: whether a list is showing is a moment of the interface, not a
		// fact about the message, and a caller holding it would be modelling
		// the widget rather than the message.
		state := *ui.Local(w, mentionKey{}, func() *mentionState { return &mentionState{} })
		if in.Changed() {
			r.changed = true
		}

		// The panel follows the field, so everything about it is decided
		// from the text: the word after the last @ is what is being
		// completed, and the names containing it are what it could become.
		word := mentionWord(*value)
		matches := matchingPeople(opts.People, word)

		dismissed := false
		if state.open && len(matches) > 0 {
			switch {
			case in.Shortcut(0, ui.KeyEscape):
				// Escape closes the panel and not the field. Somebody typing
				// a message who presses Escape wants to stop mentioning
				// people, not to lose what they wrote.
				dismissed = true
				state.highlighted = ""
			case in.Shortcut(0, ui.KeyDown):
				state.highlighted = nextName(matches, state.highlighted, 1)
			case in.Shortcut(0, ui.KeyUp):
				state.highlighted = nextName(matches, state.highlighted, -1)
			case in.Submitted():
				taken := state.highlighted
				if taken == "" {
					taken = matches[0]
				}
				takeMention(c, value, &r, state, taken)
			}
		}
		state.open = word != "" && len(matches) > 0 && !dismissed
		if !state.open {
			state.highlighted = ""
		} else if state.highlighted != "" && !hasString(matches, state.highlighted) {
			// The highlight outlived the name it was on — the word was edited
			// under it — so it goes rather than pointing at nothing.
			state.highlighted = ""
		}
		if in.Submitted() && !state.open {
			r.submitted = true
		}

		if state.open && len(matches) > 0 && !opts.Disabled {
			shown := matches
			if len(shown) > limit {
				shown = shown[:limit]
			}
			names := shown
			k := core.Tokens(c)
			// Under the field rather than over the window: a mention list
			// that flies off to a corner of the screen is a list somebody
			// has to look away from their own words to read.
			ui.Column(c).FillWidth().Padding(u * 0.5).Radius(theme.SmallRadius).
				Background(k.Surface).Children(func() {
				for _, name := range names {
					row := optionRow(c, "@"+name, optionFace{Chosen: name == state.highlighted})
					if row.Clicked() {
						takeMention(c, value, &r, state, name)
					}
				}
			})
		}
		return in
	})
	return r
}

// mentionKey is where a MentionInput's panel state lives, on the field's own
// box, which is the one element that is in the same place every frame.
type mentionKey struct{}

// mentionState is whether the panel is showing and which suggestion the
// arrows are on.
type mentionState struct {
	open        bool
	highlighted string
}

// takeMention writes the chosen name in over the half-word being typed, and
// says so on the result.
//
// Only the tail is replaced: "@sa" plus "sam" gives "@sam ", keeping the
// letters the person typed. Replacing the whole token would throw those away
// and make the field fight the person typing in it.
func takeMention(c *ui.Context, value *string, r *MentionResult, state *mentionState, name string) {
	word := mentionWord(*value)
	if name == "" {
		return
	}
	*value = strings.TrimSuffix(*value, word) + "@" + name + " "
	state.open, state.highlighted = false, ""
	r.mentioned = "@" + name
	r.changed = true
	// The editor reads the pointer back at the top of the next frame, so ask
	// for that one: the mention has to be in the field before the panel
	// closes over it.
	c.Invalidate()
}

// matchingPeople are the names a trailing @-word could become, those starting
// with it first — which is what somebody typing the first letters of a name
// means — and then those merely containing it.
func matchingPeople(people []string, word string) []string {
	if word == "" {
		return nil
	}
	q := strings.ToLower(word)
	var starts, contains []string
	for _, p := range people {
		name := strings.TrimPrefix(strings.TrimSpace(p), "@")
		switch l := strings.ToLower(name); {
		case strings.HasPrefix(l, q):
			starts = append(starts, name)
		case strings.Contains(l, q):
			contains = append(contains, name)
		}
	}
	return append(starts, contains...)
}

// mentionWord is the word after the last @ in a value, which is the part a
// mention is completing. Empty means there is nothing being typed.
func mentionWord(value string) string {
	at := strings.LastIndex(value, "@")
	if at < 0 {
		return ""
	}
	word := value[at+1:]
	// A space ends the word, which is what stops "@sam and @jo" from
	// completing against everybody after the first mention.
	if strings.ContainsAny(word, " \t\n") {
		return ""
	}
	return word
}

// nextName is the name after the one highlighted, wrapping at both ends, which
// is what the arrows do in every list on the desktop.
func nextName(names []string, now string, by int) string {
	if len(names) == 0 {
		return ""
	}
	at := 0
	for i, n := range names {
		if n == now {
			at = i
			break
		}
	}
	at = ((at+by)%len(names) + len(names)) % len(names)
	return names[at]
}

func hasString(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}
