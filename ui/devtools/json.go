package devtools

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"
)

// ParseJSON turns JSON text into a tree.
//
// The result is the shape encoding/json produces, which is the shape every
// other JSON library in every other language produces: objects are
// map[string]any, arrays are []any, strings are string, numbers are float64,
// and true, false and null are their own values. That is a deliberate choice
// over a hand-written tree type, because the moment a JSON viewer produces
// something only this package can walk, every caller who wanted to compare a
// response against a fixture has to go through this package to do it.
//
// Numbers are float64 rather than json.Number for the same reason. A
// response body's numbers are for looking at; the one place exactness matters
// is an id, and an id large enough to lose a digit at float64 is an id
// nothing in a dashboard is keyed on.
//
// Every JSON value is covered — objects, arrays, strings, numbers, booleans,
// null, escapes and nesting — because a viewer that cannot draw one of them
// cannot draw a real response, and a real response has all of them.
func ParseJSON(src string) (any, error) {
	var out any
	if err := json.Unmarshal([]byte(src), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// MustParseJSON is ParseJSON for a value the caller already knows is good —
// a fixture, a default, a schema written by hand. It panics on a bad parse
// for the same reason every other Must in this library does: the caller has
// said the value is constant, so a failure is a bug here rather than
// something to handle.
func MustParseJSON(src string) any {
	out, err := ParseJSON(src)
	if err != nil {
		panic("devtools: MustParseJSON was given something that is not JSON: " + err.Error())
	}
	return out
}

// JSONKind is what a parsed value is, named so that a tree can say what each
// row is without inspecting it.
type JSONKind int

const (
	// KindObject is {...}.
	KindObject JSONKind = iota
	// KindArray is [...].
	KindArray
	// KindString is a quoted string.
	KindString
	// KindNumber is a number.
	KindNumber
	// KindBool is true or false.
	KindBool
	// KindNull is null.
	KindNull
)

// String names the kind the way JSON writes it, which is what a viewer puts
// in its type column.
func (k JSONKind) String() string {
	switch k {
	case KindArray:
		return "array"
	case KindString:
		return "string"
	case KindNumber:
		return "number"
	case KindBool:
		return "boolean"
	case KindNull:
		return "null"
	}
	return "object"
}

// KindOf is what a parsed value is.
//
// It is the one place the switch from a Go type to a JSON kind lives, so that
// a viewer never has to write a type assertion that could be wrong: a value
// with the wrong type in it would panic in the middle of drawing rather than
// drawing as what it is.
func KindOf(v any) JSONKind {
	switch v.(type) {
	case map[string]any:
		return KindObject
	case []any:
		return KindArray
	case string:
		return KindString
	case float64:
		return KindNumber
	case bool:
		return KindBool
	case nil:
		return KindNull
	default:
		return KindNull
	}
}

// JSONNode is one row of a tree: a key, a value, and where it came from.
//
// Path is the dotted route to the value from the root, "headers.authorization"
// or "items.3.tags". It is built by Flatten rather than by each drawing site,
// because the two things that need it — a copy button that copies one value,
// and a redaction that hides one branch — both need it to be the same string
// for the same value.
type JSONNode struct {
	// Path is the dotted route to this value.
	Path string
	// Key is the last step of the path: the object's key or the array's index.
	Key string
	// Value is the parsed value, which is the whole subtree for an object or
	// an array rather than a summary of it.
	Value any
	// Kind is what Value is, decided once by KindOf.
	Kind JSONKind
	// Depth is how far down the tree this row is, for the indent.
	Depth int
	// Redacted reports that this value's text was replaced rather than drawn.
	// It is carried on the node rather than inferred from the value, because
	// "••••" is itself a string and a viewer cannot tell the two apart.
	Redacted bool
}

// Flatten turns a parsed value into the rows a tree draws, in the order it
// draws them.
//
// Objects come out in key order and arrays in index order, both of which are
// what the JSON itself means: an object's order is not significant, so the
// only stable order is the sorted one, and an array's is. A tree that came
// out in Go's map iteration order would draw the same response differently
// every time it was opened, which makes two responses impossible to compare
// by eye.
//
// The path separator is a dot, and an array index is a plain number, so
// "items.0.sku" is the route to the first item's sku.
func Flatten(v any) []JSONNode {
	return flattenInto(nil, "", v)
}

// flattenInto walks one value and appends its rows.
func flattenInto(out []JSONNode, path string, v any) []JSONNode {
	key := path
	if i := strings.LastIndex(path, "."); i >= 0 {
		key = path[i+1:]
	}
	node := JSONNode{Path: path, Key: key, Value: v, Kind: KindOf(v)}
	out = append(out, node)

	switch t := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			child := k
			if path != "" {
				child = path + "." + k
			}
			out = flattenInto(out, child, t[k])
		}
	case []any:
		for i, item := range t {
			idx := strconv.Itoa(i)
			child := idx
			if path != "" {
				child = path + "." + idx
			}
			out = flattenInto(out, child, item)
		}
	}
	return out
}

// JSONPreview is a value as one line of text, for a row that is not expanded.
//
// It is a function because the same shortening is wanted in three places —
// the tree row, the clipboard copy of a branch, and the empty state — and
// three copies of a rule about ellipses is three ellipses. A string is
// returned whole whatever its length: truncating a value somebody has asked
// to see by clicking it is worse than a wide row, and the value is one line
// away in the editor either way.
func JSONPreview(v any, max int) string {
	if max <= 0 {
		max = 80
	}
	switch t := v.(type) {
	case nil:
		return "null"
	case bool:
		return strconv.FormatBool(t)
	case float64:
		return FormatNumber(t)
	case string:
		if max > 0 && len([]rune(t)) > max {
			r := []rune(t)
			return string(r[:max]) + "…"
		}
		return t
	case []any:
		return "[" + strconv.Itoa(len(t)) + "]"
	case map[string]any:
		return "{" + strconv.Itoa(len(t)) + "}"
	default:
		return KindOf(v).String()
	}
}

// FormatNumber is a JSON number written the way a person would write it.
//
// The point is the integers. encoding/json gives every number as a float64,
// and printing those with %v gives "1.204" as "1204" and "1e+06" as "1e+06"
// — so a request id of 1204 reads as 1204 but a count of 1200000 reads in a
// form nobody writes. Integers are written plainly, with thousands
// separators, and everything else keeps up to six decimal places with the
// trailing zeros taken off.
func FormatNumber(f float64) string {
	if f == float64(int64(f)) && f < 1e15 && f > -1e15 {
		return Commas(int64(f))
	}
	s := strconv.FormatFloat(f, 'f', 6, 64)
	s = strings.TrimRight(s, "0")
	s = strings.TrimSuffix(s, ".")
	return s
}

// Commas is an integer with thousands separators, split into its sign, its
// integer part and — if it has one — its fraction, because a separator may go
// between any two digits of the integer part and between none of the rest.
func Commas(n int64) string {
	neg := n < 0
	if neg {
		n = -n
	}
	digits := strconv.FormatInt(n, 10)
	var out []byte
	for i := 0; i < len(digits); i++ {
		if i > 0 && (len(digits)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, digits[i])
	}
	if neg {
		return "-" + string(out)
	}
	return string(out)
}

// redacted is what a hidden value is drawn as. It is a value rather than a
// boolean on the node because a viewer that forgot to check the flag would
// then draw the real thing, and a flag that is checked in one place out of
// three is three places to get wrong.
const redacted = "••••••••"

// Redact returns a copy of a parsed value with the values under any of the
// named keys replaced.
//
// The keys are matched at every depth and under the rule MatchesSecret
// describes, so "authorization" hides "headers.authorization" and
// "users.2.authorization" alike, and "key" hides "x-api-key". A caller that
// had to write both paths would eventually write one of them wrong, and the
// cost of that is a token on a screenshot.
//
// The walk copies rather than edits in place: the caller's own tree is not
// modified, because the same tree is probably about to be drawn somewhere
// else — in a diff, in a log line — where hiding is not wanted. Objects and
// arrays are rebuilt; everything else is copied by value.
func Redact(v any, keys []string) any {
	if len(keys) == 0 {
		return v
	}
	return redactValue(v, keys)
}

// redactValue is Redact's walk over one value.
func redactValue(v any, keys []string) any {
	switch t := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, item := range t {
			if MatchesSecret(k, keys) {
				out[k] = redacted
				continue
			}
			out[k] = redactValue(item, keys)
		}
		return out
	case []any:
		out := make([]any, len(t))
		for i, item := range t {
			out[i] = redactValue(item, keys)
		}
		return out
	default:
		return v
	}
}

// DefaultSecretKeys is the set of words whose values are hidden unless the
// caller says otherwise.
//
// It is a list of whole words rather than a pattern, because a pattern catches
// things that are not secrets — "token_type" is not a token, "key_length" is
// not a key — and a viewer that hides half the field names of a response is a
// viewer nobody trusts to show the real ones.
//
// It is the same set KeyValues.IsSecret uses, and the two are matched the same
// way by MatchesSecret. A package with two rules for "is this a secret" is a
// package where the request headers are hidden and the equivalent field in
// the body is not, which is the one place that difference matters.
func DefaultSecretKeys() []string {
	return []string{
		"authorization", "auth", "cookie", "set-cookie",
		"password", "passwd", "secret", "token", "signature", "credential",
		"credentials", "key", "apikey", "api_key",
		"private_key", "private-key",
		"access_token", "refresh_token", "auth_token", "auth-token",
		"session_token", "client_secret",
	}
}

// MatchesSecret reports whether a field's name means its value is a secret,
// under the rule the whole package uses.
//
// The rule is: the name is one of the words, or any run of its
// hyphen-separated pieces is.
//
// The first half is what lets a caller hide a field nobody has thought of —
// "northgate-signing" is theirs to name and it is hidden because they said
// so. The runs are what catch the spellings nobody enumerated: "key" inside
// "X-Api-Key", and "northgate-signing" inside "x-northgate-signing". A name
// is split into at most a handful of pieces, so checking every run of them is
// a few dozen comparisons and no cleverness.
//
// A hyphen separates words and an underscore joins them, because that is what
// each does in the two places these names are written: HTTP header names and
// the identifiers in a JSON body. So "X-Api-Key" is caught by "key" and
// "token_type" is not caught by "token" — which is the difference between
// hiding a credential and hiding a field that says what kind of credential
// something is, and a viewer that got it backwards would hide the second on
// every response it ever showed.
func MatchesSecret(key string, words []string) bool {
	lower := lowerASCII(key)
	set := wordSet(words)
	if set[lower] {
		return true
	}
	if !strings.Contains(lower, "-") {
		return false
	}
	pieces := strings.Split(lower, "-")
	// Every contiguous run of pieces, longest first. Longest first because a
	// caller who named "api-key" means it in preference to the bare "key" it
	// contains, and both match either way — but the order makes the answer to
	// "why was this hidden" the most specific one available.
	for start := 0; start < len(pieces); start++ {
		for end := len(pieces); end > start; end-- {
			if set[strings.Join(pieces[start:end], "-")] {
				return true
			}
		}
	}
	return false
}

// wordSet is the lower-cased word list as a set, built once per call site
// rather than once per key: a response has hundreds of keys and the set does
// not change while it is being walked.
func wordSet(words []string) map[string]bool {
	out := make(map[string]bool, len(words))
	for _, w := range words {
		out[lowerASCII(w)] = true
	}
	return out
}

// RedactedPaths are the dotted routes in a tree whose values have been
// hidden. It exists so that a viewer can mark those rows as redacted rather
// than inferring it from the text: "••••••••" is itself a valid JSON string,
// and a row drawn from it has to say which it is.
func RedactedPaths(v any, keys []string) []string {
	if len(keys) == 0 {
		return nil
	}
	var out []string
	collectRedacted("", v, keys, &out)
	return out
}

// collectRedacted walks the tree recording where a value was replaced. It is
// the same walk as Redact's under the same rule; there are two of them
// because one returns a tree and the other a list of paths, and folding one
// into the other would mean re-walking what Redact has already walked.
func collectRedacted(path string, v any, keys []string, out *[]string) {
	switch t := v.(type) {
	case map[string]any:
		names := make([]string, 0, len(t))
		for k := range t {
			names = append(names, k)
		}
		sort.Strings(names)
		for _, k := range names {
			child := k
			if path != "" {
				child = path + "." + k
			}
			if MatchesSecret(k, keys) {
				*out = append(*out, child)
				continue
			}
			collectRedacted(child, t[k], keys, out)
		}
	case []any:
		for i, item := range t {
			child := strconv.Itoa(i)
			if path != "" {
				child = path + "." + child
			}
			collectRedacted(child, item, keys, out)
		}
	}
}

// SchemaNode is one entry of a schema: a type, a description and whatever it
// says about the entries inside it.
type SchemaNode struct {
	// Type is what the value is, as a JSON kind's name: "object", "array",
	// "string". Empty means "any", which is what a schema that only gives a
	// description says.
	Type string
	// Description is the schema's own sentence, which is the only part of a
	// schema a person reads.
	Description string
	// Required lists the keys that must be there.
	Required []string
	// Properties are the object's own entries.
	Properties map[string]SchemaNode
	// Items is the array's entry, when it has one.
	Items *SchemaNode
}

// SchemaRows turns a schema into the rows a tree draws, in the same shape
// Flatten produces for a value — so a schema and a parsed response can be
// drawn by one component rather than two.
//
// It is not called SchemaTree because the component is: this is what a schema
// *means* and SchemaTree is where it goes.
//
// The order is the sorted key order for the same reason Flatten sorts: an
// object has no order of its own, and a schema that came out in a different
// one each time would make two responses impossible to compare. Required
// keys come first within a key's own row rather than being a separate
// section, because whether one key is required is a fact about that key.
func SchemaRows(s SchemaNode) []JSONNode {
	var out []JSONNode
	schemaInto(&out, "", "", s)
	return out
}

// schemaInto walks one schema node and appends its rows.
func schemaInto(out *[]JSONNode, path, key string, s SchemaNode) {
	kind := schemaKind(s.Type)
	*out = append(*out, JSONNode{
		Path: path, Key: key, Kind: kind,
		Value: s.Description,
	})

	names := make([]string, 0, len(s.Properties))
	for name := range s.Properties {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		child := name
		if path != "" {
			child = path + "." + name
		}
		schemaInto(out, child, name, s.Properties[name])
	}
	if s.Items != nil {
		child := "[]"
		if path != "" {
			child = path + ".[]"
		}
		schemaInto(out, child, "[]", *s.Items)
	}
}

// schemaKind is a schema type name as a JSON kind. Anything unrecognised is
// KindNull, which draws as "any" rather than as a type nobody asked for.
func schemaKind(name string) JSONKind {
	switch strings.ToLower(name) {
	case "object":
		return KindObject
	case "array":
		return KindArray
	case "string":
		return KindString
	case "number", "integer":
		return KindNumber
	case "boolean":
		return KindBool
	case "null":
		return KindNull
	}
	return KindNull
}
