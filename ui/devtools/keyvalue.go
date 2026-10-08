package devtools

import (
	"sort"
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/theme"
)

// KeyValue is one pair.
type KeyValue struct {
	// Key is the name, as the protocol writes it. A header name is
	// case-insensitive in the protocol and not in a map, so everything that
	// compares keys folds them — see KeyValue.IsSecret.
	Key string
	// Value is the pair's value, which is a string because headers, query
	// parameters and form fields are all strings even when their contents
	// are not.
	Value string
}

// KeyValues is a set of pairs, in the order they were added.
//
// It is a type rather than a map because the order is the interface: a header
// list that reshuffled itself between frames could not be read, could not be
// compared with another, and would put "authorization" somewhere new each
// time somebody pressed a key.
type KeyValues []KeyValue

// Get is the value of a key, matched without regard to case. HTTP header
// names and query parameter names differ only in case often enough that a
// case-sensitive lookup is a bug waiting for a particular client.
func (kvs KeyValues) Get(key string) (string, bool) {
	for _, kv := range kvs {
		if strings.EqualFold(kv.Key, key) {
			return kv.Value, true
		}
	}
	return "", false
}

// Keys are the names, in order.
func (kvs KeyValues) Keys() []string {
	out := make([]string, 0, len(kvs))
	for _, kv := range kvs {
		out = append(out, kv.Key)
	}
	return out
}

// With returns a copy of the set with key set to value, replacing what was
// there under a different capitalisation and keeping everything else where it
// was.
//
// Replacing in place is what keeps the order stable: a caller adding one
// header to a set of ten should not have that header jump to the end or the
// beginning, and a map would put it wherever the hash said.
func (kvs KeyValues) With(key, value string) KeyValues {
	out := make(KeyValues, len(kvs))
	copy(out, kvs)
	for i := range out {
		if strings.EqualFold(out[i].Key, key) {
			out[i].Value = value
			return out
		}
	}
	return append(out, KeyValue{Key: key, Value: value})
}

// Without is a copy of the set with key removed, again keeping the order of
// what is left.
func (kvs KeyValues) Without(key string) KeyValues {
	out := make(KeyValues, 0, len(kvs))
	for _, kv := range kvs {
		if strings.EqualFold(kv.Key, key) {
			continue
		}
		out = append(out, kv)
	}
	return out
}

// Sorted is a copy in name order, for the places that want a stable order
// rather than the order things were added — a test, and a diff between two
// requests.
func (kvs KeyValues) Sorted() KeyValues {
	out := make(KeyValues, len(kvs))
	copy(out, kvs)
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Key < out[j].Key
	})
	return out
}

// IsSecret reports whether a key is one whose value must not be shown, under
// the rule MatchesSecret describes — the same rule ResponseViewer and Redact
// use, so that a header hidden here and a field of the same name in the body
// are both hidden rather than one of them.
//
// A key this control has been told about directly is always hidden; the words
// only catch the spellings nobody thought to enumerate.
func (kvs KeyValues) IsSecret(key string) bool {
	return MatchesSecret(key, DefaultSecretKeys())
}

// KeyValueInputOptions configure a KeyValueInput.
type KeyValueInputOptions struct {
	// Values is the caller's set, written as rows are added, edited and
	// removed. It is a pointer because a control that copied the set would
	// leave the caller looking at one set and the request carrying another.
	Values *KeyValues
	// KeyLabel and ValueLabel name the two fields; they are required
	// because a pair with no names is two boxes and a person cannot tell
	// which end is which.
	KeyLabel, ValueLabel string
	// NewKey and NewValue are the caller's strings for the pair being typed
	// at the bottom. They are required, and required for the same reason
	// every other field in this library takes one: the pair is typed on one
	// frame and pressed on another, and a string held only inside this
	// component would be empty again by the time somebody presses Add.
	NewKey, NewValue *string
	// Secrets are the names whose values are hidden while they are shown.
	// Nil hides nothing: this control draws what it is given, and whether a
	// value is a secret is the caller's business, not something a table of
	// headers should guess at.
	Secrets []string
	// Placeholder is the empty key field's hint.
	Placeholder string
}

// KeyValueInputResult carries a KeyValueInput and what was done in it.
type KeyValueInputResult struct {
	// Element is the whole control.
	Element *ui.Element
	// removed is the key of the row removed this frame.
	removed string
}

// Removed returns the key of the row that was removed this frame, empty when
// none was. It is a key rather than an index because the caller's set is
// theirs to reorder, and an index would mean a different row by the time
// they read it.
func (r KeyValueInputResult) Removed() string { return r.removed }

// KeyValueInput is a set of pairs as a list of rows with an add field at the
// bottom.
//
// The value of a pair whose name is in Secrets is drawn as dots and cannot be
// read, but it is *not* removed from the caller's set: hiding a header from
// the screen is not hiding it from the request, and a control that dropped
// it would quietly send a request without its authorization.
//
// A row is edited by typing into it. That is the whole interaction, and it
// writes on every keystroke rather than on a commit — the same rule every
// field in this library follows, and the reason a header list never needs a
// Save button.
func KeyValueInput(c *ui.Context, opts KeyValueInputOptions) KeyValueInputResult {
	if opts.Values == nil {
		panic("devtools: KeyValueInput needs the *KeyValues it writes to; it " +
			"keeps no set of its own")
	}
	if opts.KeyLabel == "" || opts.ValueLabel == "" {
		panic("devtools: KeyValueInput needs a KeyLabel and a ValueLabel; two " +
			"boxes with no names do not say which end is which")
	}
	if opts.NewKey == nil || opts.NewValue == nil {
		panic("devtools: KeyValueInput needs the *string NewKey and NewValue to " +
			"point at; a pair typed on one frame and pressed on another cannot " +
			"be held in here")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	hidden := secretSet(opts.Secrets)

	var r KeyValueInputResult
	list := ui.Column(c).FillWidth().Gap(u * 0.5).Role(ui.RoleList).
		Label(opts.KeyLabel + ", " + opts.ValueLabel)
	list.Children(func() {
		// Indexed rather than ranged, and stepped back after a removal.
		//
		// A range over a slice takes its length once, and Without replaces the
		// slice with a shorter one — so the next iteration would index the new
		// slice at an offset belonging to the old one and either read the
		// wrong row or run off the end. That is the whole reason removing a
		// row is not a one-line filter.
		for i := 0; i < len(*opts.Values); i++ {
			// A copy of the pair rather than a pointer into the slice: the
			// slice is replaced under us a few lines below, and a pointer
			// into it would be left addressing the old backing array.
			kv := (*opts.Values)[i]
			removed := false
			ui.Row(c).FillWidth().Gap(u).AlignItems(ui.Center).Children(func() {
				ui.Box(c).Width(kvNameWidth).Shrink(0).Children(func() {
					input.TextInput(c, &kv.Key, input.TextInputOptions{
						Label: opts.KeyLabel,
						Width: kvNameWidth,
					})
				})
				if hidden[strings.ToLower(kv.Key)] {
					// The value is not drawn and not offered: a button that
					// revealed it would be a way to put a secret on a
					// screenshot, and the caller who needs to see it knows
					// it already.
					ui.Text(c, redacted).TextColor(k.Danger).Grow(1).Font(monoFamily).
						FontSize(core.FontSize(c, theme.RowSize)).SingleLine()
				} else {
					ui.Box(c).Grow(1).Shrink(0).Children(func() {
						input.TextInput(c, &kv.Value, input.TextInputOptions{
							Label: opts.ValueLabel,
						})
					})
				}
				remove := input.IconButton(c, glyphClose,
					core.Msg(c, "devtools.remove", "Remove "+kv.Key),
					input.ButtonOptions{})
				if remove.Clicked() {
					*opts.Values = opts.Values.Without(kv.Key)
					r.removed = kv.Key
					removed = true
				}
			})
			if removed {
				// The rows below this one have moved up into its place, so
				// look at that index again rather than skipping past the
				// first one that was not removed.
				i--
			}
		}
		keyRow(c, opts)
	})
	r.Element = list
	return r
}

// kvNameWidth is the name field's width: long enough for a header name and
// short enough that the value beside it gets most of the row.
const kvNameWidth float32 = 170

// keyRow is the add-a-pair field at the bottom of the list.
func keyRow(c *ui.Context, opts KeyValueInputOptions) {
	u := core.Density(c).Unit()
	// The add row's fields are named differently from the rows above them.
	// Two fields in one control with the same name is a screen reader
	// announcing the same thing twice and a test clicking the first one it
	// finds, which is a row's key rather than the empty one at the bottom.
	ui.Row(c).FillWidth().Gap(u).AlignItems(ui.End).Children(func() {
		ui.Box(c).Width(kvNameWidth).Shrink(0).Children(func() {
			input.TextInput(c, opts.NewKey, input.TextInputOptions{
				Label:       core.Msg(c, "devtools.newKey", "New ") + strings.ToLower(opts.KeyLabel),
				Placeholder: opts.Placeholder,
				Width:       kvNameWidth,
			})
		})
		ui.Box(c).Grow(1).Shrink(0).Children(func() {
			input.TextInput(c, opts.NewValue, input.TextInputOptions{
				Label: core.Msg(c, "devtools.newValue", "New ") + strings.ToLower(opts.ValueLabel),
			})
		})
		add := input.Button(c, core.Msg(c, "devtools.add", "Add"),
			input.ButtonOptions{Disabled: *opts.NewKey == ""})
		if add.Clicked() && *opts.NewKey != "" {
			*opts.Values = opts.Values.With(*opts.NewKey, *opts.NewValue)
		}
	})
}
