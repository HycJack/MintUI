package devtools

import (
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

// ── ParseJSON ──────────────────────────────────────────────────────────────

// A response body that has all seven JSON kinds in it, an escape, a number in
// two forms and two levels of nesting. It is the fixture everything else
// checks against, because a viewer that cannot draw one of these cannot draw a
// real response.
const sampleJSON = `{
  "id": "cb_2871",
  "count": 1204,
  "ratio": 0.125,
  "ok": true,
  "retried": false,
  "error": null,
  "note": "line one\nline two\ttab \"quoted\" é \\ /",
  "tags": ["backend", "flaky"],
  "customer": {
    "name": "Riverside Clinic",
    "seats": 12,
    "nested": {"deep": {"value": "found"}}
  },
  "empty_list": [],
  "empty_object": {}
}`

func TestParseJSONAnObject(t *testing.T) {
	got, err := ParseJSON(sampleJSON)
	if err != nil {
		t.Fatalf("ParseJSON: %v", err)
	}
	obj, ok := got.(map[string]any)
	if !ok {
		t.Fatalf("an object parses to map[string]any, got %T", got)
	}
	if len(obj) != 11 {
		t.Errorf("the object has 11 keys, got %d", len(obj))
	}
	if obj["id"] != "cb_2871" {
		t.Errorf(`id = %v, want "cb_2871"`, obj["id"])
	}
}

func TestParseJSONEveryKind(t *testing.T) {
	got, err := ParseJSON(sampleJSON)
	if err != nil {
		t.Fatal(err)
	}
	obj := got.(map[string]any)

	cases := []struct {
		key  string
		kind JSONKind
	}{
		{"id", KindString},
		{"count", KindNumber},
		{"ok", KindBool},
		{"error", KindNull},
		{"tags", KindArray},
		{"customer", KindObject},
		{"empty_list", KindArray},
		{"empty_object", KindObject},
	}
	for _, c := range cases {
		if got := KindOf(obj[c.key]); got != c.kind {
			t.Errorf("KindOf(%q) = %v, want %v", c.key, got, c.kind)
		}
	}
}

func TestParseJSONNumbers(t *testing.T) {
	got, _ := ParseJSON(`{"i": 42, "neg": -7, "f": 0.125, "exp": 1e3}`)
	obj := got.(map[string]any)
	cases := []struct {
		key  string
		want float64
	}{{"i", 42}, {"neg", -7}, {"f", 0.125}, {"exp", 1000}}
	for _, c := range cases {
		if obj[c.key] != c.want {
			t.Errorf("%s = %v, want %v", c.key, obj[c.key], c.want)
		}
	}
}

func TestParseJSONBooleansAndNull(t *testing.T) {
	got, _ := ParseJSON(`{"t": true, "f": false, "n": null}`)
	obj := got.(map[string]any)
	if obj["t"] != true {
		t.Errorf("t = %v, want true", obj["t"])
	}
	if obj["f"] != false {
		t.Errorf("f = %v, want false", obj["f"])
	}
	if obj["n"] != nil {
		t.Errorf("n = %v, want nil", obj["n"])
	}
	if KindOf(obj["n"]) != KindNull {
		t.Error("null is its own kind")
	}
}

func TestParseJSONEscapes(t *testing.T) {
	got, err := ParseJSON(`"line one\nline two\ttab \"quoted\" é \\ /"`)
	if err != nil {
		t.Fatal(err)
	}
	want := "line one\nline two\ttab \"quoted\" é \\ /"
	if got != want {
		t.Errorf("the escapes did not come back:\n got %q\nwant %q", got, want)
	}
}

func TestParseJSONUnicode(t *testing.T) {
	got, err := ParseJSON(`{"name": "Riverside Clinic — 診所"}`)
	if err != nil {
		t.Fatal(err)
	}
	if obj := got.(map[string]any); obj["name"] != "Riverside Clinic — 診所" {
		t.Errorf("name = %v", obj["name"])
	}
}

func TestParseJSONNesting(t *testing.T) {
	got, err := ParseJSON(`{"a":{"b":{"c":[{"d":"deep"}]}}}`)
	if err != nil {
		t.Fatal(err)
	}
	deep := got.(map[string]any)["a"].(map[string]any)["b"].(map[string]any)["c"].([]any)
	if len(deep) != 1 {
		t.Fatalf("the nested array has %d entries", len(deep))
	}
	if deep[0].(map[string]any)["d"] != "deep" {
		t.Errorf("the value at the bottom is wrong: %v", deep[0])
	}
}

func TestParseJSONOfTheTopLevelKinds(t *testing.T) {
	cases := []struct {
		src  string
		want JSONKind
	}{
		{`[1,2,3]`, KindArray},
		{`"a string"`, KindString},
		{`42`, KindNumber},
		{`true`, KindBool},
		{`null`, KindNull},
		{`{}`, KindObject},
	}
	for _, c := range cases {
		got, err := ParseJSON(c.src)
		if err != nil {
			t.Errorf("ParseJSON(%q): %v", c.src, err)
			continue
		}
		if k := KindOf(got); k != c.want {
			t.Errorf("ParseJSON(%q) is %v, want %v", c.src, k, c.want)
		}
	}
}

func TestParseJSONRefusesWhatIsNotJSON(t *testing.T) {
	cases := []string{
		"",
		"{",
		"{\"a\": }",
		"{a: 1}",
		"[1, 2",
		"nope",
		"{\"a\": 1,}",
		"'single quotes'",
		"{\"a\": 01}",
	}
	for _, src := range cases {
		if v, err := ParseJSON(src); err == nil {
			t.Errorf("ParseJSON(%q) should have failed, got %v", src, v)
		}
	}
}

func TestMustParseJSONPanicsOnRubbish(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("MustParseJSON on rubbish should panic")
		}
	}()
	MustParseJSON("{not json}")
}

func TestJSONKindString(t *testing.T) {
	cases := map[JSONKind]string{
		KindObject: "object", KindArray: "array", KindString: "string",
		KindNumber: "number", KindBool: "boolean", KindNull: "null",
	}
	for k, want := range cases {
		if got := k.String(); got != want {
			t.Errorf("%d.String() = %q, want %q", k, got, want)
		}
	}
}

// ── Flatten ────────────────────────────────────────────────────────────────

func TestFlattenIsSortedAndIndexed(t *testing.T) {
	got, _ := ParseJSON(sampleJSON)
	nodes := Flatten(got)

	// The root first, then the object's keys in sorted order — including the
	// whole of the nested object before the key that sorts after it, which is
	// what makes two responses comparable by eye.
	if nodes[0].Path != "" {
		t.Errorf("the root has no path, got %q", nodes[0].Path)
	}
	wantOrder := []string{
		"count", "customer", "customer.name", "customer.nested",
		"customer.nested.deep", "customer.nested.deep.value", "customer.seats",
		"empty_list", "empty_object",
	}
	for i, want := range wantOrder {
		if nodes[i+1].Path != want {
			t.Fatalf("row %d is %q, want %q; the whole order is %v",
				i+1, nodes[i+1].Path, want, paths(nodes))
		}
	}
}

func TestFlattenIndexesArrayElements(t *testing.T) {
	got, _ := ParseJSON(`{"tags": ["a", "b"], "items": [{"sku": "x"}]}`)
	nodes := Flatten(got)
	for _, want := range []string{"tags", "tags.0", "tags.1", "items", "items.0", "items.0.sku"} {
		if !hasPath(nodes, want) {
			t.Errorf("missing %q; the rows are %v", want, paths(nodes))
		}
	}
}

func TestFlattenDoesNotTouchTheCallersOrder(t *testing.T) {
	// Two calls on the same value must give the same order: an object's
	// order is not significant in JSON and the only stable one is sorted.
	a, _ := ParseJSON(sampleJSON)
	b, _ := ParseJSON(sampleJSON)
	if !reflect.DeepEqual(paths(Flatten(a)), paths(Flatten(b))) {
		t.Error("two parses of the same text flattened in different orders")
	}
}

func TestFlattenOfTheEmptyThings(t *testing.T) {
	if got := Flatten(nil); len(got) != 1 || got[0].Kind != KindNull {
		t.Errorf("flattening null gives %v", got)
	}
	if got := Flatten(map[string]any{}); len(got) != 1 {
		t.Errorf("an empty object is just itself, got %d rows", len(got))
	}
	if got := Flatten([]any{}); len(got) != 1 {
		t.Errorf("an empty array is just itself, got %d rows", len(got))
	}
}

func paths(nodes []JSONNode) []string {
	out := make([]string, len(nodes))
	for i, n := range nodes {
		out[i] = n.Path
	}
	return out
}

func hasPath(nodes []JSONNode, want string) bool {
	for _, n := range nodes {
		if n.Path == want {
			return true
		}
	}
	return false
}

// ── previews and numbers ───────────────────────────────────────────────────

func TestJSONPreview(t *testing.T) {
	got, _ := ParseJSON(`{"s":"hello","n":42,"b":true,"z":null,"o":{"k":1},"a":[1,2,3]}`)
	obj := got.(map[string]any)
	cases := []struct {
		key  string
		want string
	}{
		{"s", "hello"}, {"n", "42"}, {"b", "true"}, {"z", "null"},
		{"o", "{1}"}, {"a", "[3]"},
	}
	for _, c := range cases {
		if got := JSONPreview(obj[c.key], 0); got != c.want {
			t.Errorf("JSONPreview(%q) = %q, want %q", c.key, got, c.want)
		}
	}
}

func TestJSONPreviewTruncatesAStringAndNothingElse(t *testing.T) {
	long := "abcdefghij"
	if got := JSONPreview(long, 4); got != "abcd…" {
		t.Errorf("a long string = %q", got)
	}
	// A string inside the limit comes back whole: truncating a value
	// somebody has clicked on is worse than a wide row.
	if got := JSONPreview("abcd", 4); got != "abcd" {
		t.Errorf("a short string = %q", got)
	}
	// Branches are counted, not printed, whatever the limit.
	if got := JSONPreview([]any{1, 2}, 1); got != "[2]" {
		t.Errorf("an array is counted, got %q", got)
	}
}

func TestFormatNumber(t *testing.T) {
	cases := []struct {
		in   float64
		want string
	}{
		{0, "0"},
		{42, "42"},
		{1204, "1,204"},
		{-1204, "-1,204"},
		{1000000, "1,000,000"},
		{0.125, "0.125"},
		{1.5, "1.5"},
		{1.25, "1.25"},
		{0.5, "0.5"},
	}
	for _, c := range cases {
		if got := FormatNumber(c.in); got != c.want {
			t.Errorf("FormatNumber(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestCommas(t *testing.T) {
	for _, c := range []struct {
		in   int64
		want string
	}{
		{0, "0"}, {999, "999"}, {1000, "1,000"},
		{1204, "1,204"}, {1000000, "1,000,000"}, {-1204, "-1,204"},
	} {
		if got := Commas(c.in); got != c.want {
			t.Errorf("Commas(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}

// ── redaction ──────────────────────────────────────────────────────────────

func TestRedactHidesAtEveryDepth(t *testing.T) {
	value, _ := ParseJSON(`{
	  "authorization": "Bearer top",
	  "headers": {"authorization": "Bearer nested", "accept": "json"},
	  "users": [{"api_key": "k1"}, {"api_key": "k2"}]
	}`)
	got := Redact(value, []string{"authorization", "api_key"})

	obj := got.(map[string]any)
	if obj["authorization"] != redacted {
		t.Errorf("the top level was not hidden: %v", obj["authorization"])
	}
	if obj["headers"].(map[string]any)["authorization"] != redacted {
		t.Error("a nested value was not hidden")
	}
	if obj["headers"].(map[string]any)["accept"] != "json" {
		t.Error("a value that was not named was hidden")
	}
	users := obj["users"].([]any)
	if users[0].(map[string]any)["api_key"] != redacted {
		t.Error("a value inside an array was not hidden")
	}
	if users[1].(map[string]any)["api_key"] != redacted {
		t.Error("the second one was not hidden either")
	}
}

func TestRedactDoesNotTouchTheCallersTree(t *testing.T) {
	value, _ := ParseJSON(`{"token": "secret"}`)
	Redact(value, []string{"token"})
	if value.(map[string]any)["token"] != "secret" {
		t.Error("Redact edited the value it was given; the same tree is drawn elsewhere")
	}
}

func TestRedactMatchesWithoutRegardToCase(t *testing.T) {
	value, _ := ParseJSON(`{"Authorization": "Bearer x", "API_KEY": "k"}`)
	got := Redact(value, []string{"authorization", "api_key"}).(map[string]any)
	if got["Authorization"] != redacted || got["API_KEY"] != redacted {
		t.Errorf("a differently-cased name was not hidden: %v", got)
	}
}

func TestRedactWithNoKeysChangesNothing(t *testing.T) {
	value, _ := ParseJSON(`{"token": "secret"}`)
	if got := Redact(value, nil); got.(map[string]any)["token"] != "secret" {
		t.Error("Redact with no keys should change nothing")
	}
}

func TestRedactedPaths(t *testing.T) {
	value, _ := ParseJSON(`{"a": {"token": 1, "b": 2}, "c": [{"password": 3}]}`)
	got := RedactedPaths(value, []string{"token", "password"})
	want := []string{"a.token", "c.0.password"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("RedactedPaths = %v, want %v", got, want)
	}
	if got := RedactedPaths(value, nil); got != nil {
		t.Errorf("no keys is no paths, got %v", got)
	}
}

func TestDefaultSecretKeysCoverTheObviousOnes(t *testing.T) {
	keys := DefaultSecretKeys()
	for _, want := range []string{
		"authorization", "auth", "cookie", "password", "passwd", "secret",
		"token", "signature", "credential", "credentials", "key", "apikey",
		"api_key", "private_key", "access_token", "refresh_token",
		"client_secret", "session_token",
	} {
		if !MatchesSecret(want, keys) {
			t.Errorf("DefaultSecretKeys does not catch %q", want)
		}
	}
}

// TestMatchesSecretIsOneRule covers the rule the whole package uses, because
// a package with two of them is a package where a header is hidden and the
// same field in the body is not.
func TestMatchesSecretIsOneRule(t *testing.T) {
	words := DefaultSecretKeys()

	for _, yes := range []string{
		// Exact: a name somebody thought of.
		"authorization", "Authorization", "password", "api_key", "refresh_token",
		// A hyphen separates, so a prefixed or vendor-spelled one is caught.
		"X-Api-Key", "x-api-key", "proxy-authorization", "private-key",
		"aws-secret-access-key", "x-auth-token",
	} {
		if !MatchesSecret(yes, words) {
			t.Errorf("%q should be secret", yes)
		}
	}

	// Anything a caller adds explicitly is hidden whatever it is — including
	// a name of their own that is in no list anywhere.
	custom := append(append([]string{}, words...), "northgate-signing", "cb_custom_thing")
	for _, name := range []string{"northgate-signing", "x-northgate-signing", "cb_custom_thing"} {
		if !MatchesSecret(name, custom) {
			t.Errorf("a caller's own word should hide %q", name)
		}
	}
	if MatchesSecret("northgate-signing", words) {
		t.Error("a name in no list is not a secret; that is what the list is for")
	}

	for _, no := range []string{
		// Not secrets, and hiding them is how a viewer stops being trusted.
		"token_type", "keyboard", "monkey", "key_length", "authorization_header",
		"accept", "content-type", "user-agent", "request-id",
	} {
		if MatchesSecret(no, words) {
			t.Errorf("%q should not be secret", no)
		}
	}
}

// TestRedactCatchesWhatIsSecretCatches is the property that matters: a value
// is hidden if and only if IsSecret says its name is secret. Two rules in one
// package is how a header gets hidden and the same field in the body does not.
func TestRedactCatchesWhatIsSecretCatches(t *testing.T) {
	body := `{
	  "authorization": "LEAK-a",
	  "X-Api-Key": "LEAK-b",
	  "headers": {"cookie": "LEAK-c", "accept": "application/json"},
	  "token_type": "Bearer",
	  "keyboard": "us",
	  "list": [{"password": "LEAK-d"}]
	}`
	value, err := ParseJSON(body)
	if err != nil {
		t.Fatal(err)
	}
	clean := Redact(value, DefaultSecretKeys())

	for _, path := range []string{"authorization", "X-Api-Key", "headers.cookie", "list.0.password"} {
		if got := at(clean, path); got == redacted {
			continue
		} else if got == nil {
			t.Errorf("%q was removed rather than hidden", path)
		} else {
			t.Errorf("%q was not hidden: %v", path, got)
		}
	}
	for _, path := range []string{"headers.accept", "token_type", "keyboard"} {
		if got := at(clean, path); got == redacted {
			t.Errorf("%q is not a secret and was hidden anyway", path)
		}
	}
}

// at is the value at a dotted path, for a test that is about which paths got
// hidden rather than about the shape of the tree.
func at(v any, path string) any {
	cur := v
	for _, step := range strings.Split(path, ".") {
		switch t := cur.(type) {
		case map[string]any:
			next, ok := t[step]
			if !ok {
				return nil
			}
			cur = next
		case []any:
			i, err := strconv.Atoi(step)
			if err != nil || i >= len(t) {
				return nil
			}
			cur = t[i]
		default:
			return nil
		}
	}
	return cur
}

// ── the schema tree ────────────────────────────────────────────────────────

func TestSchemaRowsAreSortedAndIndexed(t *testing.T) {
	got := SchemaRows(SchemaNode{
		Type:        "object",
		Description: "A callback",
		Properties: map[string]SchemaNode{
			"zebra": {Type: "string"},
			"apple": {Type: "number"},
		},
	})
	want := []string{"", "apple", "zebra"}
	if !reflect.DeepEqual(paths(got), want) {
		t.Errorf("SchemaRows = %v, want %v", paths(got), want)
	}
}

func TestSchemaRowsWalkAnArray(t *testing.T) {
	items := SchemaNode{Type: "string", Description: "one tag"}
	got := SchemaRows(SchemaNode{
		Type:     "object",
		Items:    &items,
		Required: []string{"id"},
	})
	for _, want := range []string{"", "[]"} {
		if !hasPath(got, want) {
			t.Errorf("missing %q; the rows are %v", want, paths(got))
		}
	}
}

func TestSchemaKindOfAnUnknownType(t *testing.T) {
	// A schema that only gives a description says "any", which draws as null
	// rather than as a type nobody asked for.
	if got := schemaKind("whatever"); got != KindNull {
		t.Errorf("an unknown schema type is %v, want null", got)
	}
	for _, c := range []struct {
		name string
		want JSONKind
	}{
		{"object", KindObject}, {"array", KindArray}, {"string", KindString},
		{"integer", KindNumber}, {"number", KindNumber}, {"boolean", KindBool},
	} {
		if got := schemaKind(c.name); got != c.want {
			t.Errorf("schemaKind(%q) = %v, want %v", c.name, got, c.want)
		}
	}
}

// ── uptime ─────────────────────────────────────────────────────────────────

func day(y, d int) time.Time { return time.Date(2026, time.March, d, 12, 0, 0, 0, time.UTC) }

func TestUptimeSlotsIsNinetyDaysLong(t *testing.T) {
	if got := UptimeSlots(nil); len(got) != UptimeSlotsDays {
		t.Errorf("an empty record gives %d slots, want %d", len(got), UptimeSlotsDays)
	}
}

func TestUptimeSlotsOfNothingIsAllUnknown(t *testing.T) {
	got := UptimeSlots(nil)
	for i, s := range got {
		if s != SlotUnknown {
			t.Fatalf("slot %d is %v, want unknown", i, s)
		}
	}
}

func TestUptimeSlotsRunsBackFromTheNewestRecord(t *testing.T) {
	// Thirty good days in March. The bar covers the ninety days ending on
	// the thirtieth, so the first of March is the sixtieth square and the
	// sixty days before it are unknown — not the other way round, which is
	// what anchoring to January would have given.
	var records []Slot
	for d := 1; d <= 30; d++ {
		records = append(records, Slot{At: day(2026, d), Up: true})
	}
	got := UptimeSlots(records)

	// The first square is the 31st of December and the first of March is the
	// sixty-first, so sixty squares are unknown and thirty are up.
	for i := 0; i < 60; i++ {
		if got[i] != SlotUnknown {
			t.Fatalf("square %d is %v, want unknown", i, got[i])
		}
	}
	for d := 1; d <= 30; d++ {
		if got[60+d-1] != SlotUp {
			t.Errorf("1 March to %d March should be up, square %d is %v", d, 60+d, got[60+d-1])
		}
	}
}

func TestUptimeSlotsOnASingleDay(t *testing.T) {
	// One probe is enough to fill the last square and leave the other
	// eighty-nine unknown.
	got := UptimeSlots([]Slot{{At: day(2026, 14), Up: true}})
	if got[UptimeSlotsDays-1] != SlotUp {
		t.Errorf("the newest record should be the last square, got %v", got[UptimeSlotsDays-1])
	}
	if got[0] != SlotUnknown {
		t.Errorf("everything before it is unknown, got %v", got[0])
	}
}

func TestUptimeSlotsDowngradesADayWithOneFailure(t *testing.T) {
	// One failed probe in a day of successes still makes the day down: the
	// bar answers "was there a day I should not have relied on it".
	records := []Slot{
		{At: day(2026, 1), Up: true},
		{At: day(2026, 1), Up: true},
		{At: day(2026, 1), Up: false, Error: "timeout"},
		{At: day(2026, 2), Up: true},
	}
	got := UptimeSlots(records)
	if got[UptimeSlotsDays-2] != SlotDown {
		t.Errorf("1 March is the second-to-last square, got %v", got[UptimeSlotsDays-2])
	}
	if got[UptimeSlotsDays-1] != SlotUp {
		t.Errorf("the second is the last square and is up, got %v", got[UptimeSlotsDays-1])
	}
}

func TestUptimeSlotsTellsPartialFromDown(t *testing.T) {
	// Degraded and broken are a different conversation, and a bar that
	// paints both the same cannot be read.
	records := []Slot{
		{At: day(2026, 1), Up: true, Error: "slow"},
		{At: day(2026, 2), Up: false, Error: "refused"},
	}
	got := UptimeSlots(records)
	if got[UptimeSlotsDays-2] != SlotPartial {
		t.Errorf("an answer with an error is partial, got %v", got[UptimeSlotsDays-2])
	}
	if got[UptimeSlotsDays-1] != SlotDown {
		t.Errorf("a refusal is down, got %v", got[UptimeSlotsDays-1])
	}
}

func TestUptimeSlotsIgnoresDaysOutsideTheWindow(t *testing.T) {
	// The newest record sets the window, and anything more than ninety days
	// before it is dropped rather than wrapped round into a square from the
	// end — which is the failure that would put November's outage on
	// somebody's December.
	records := []Slot{
		{At: day(2026, 14), Up: true},                                  // the anchor
		{At: time.Date(2026, 3, 3, 12, 0, 0, 0, time.UTC), Up: false},  // eleven days back
		{At: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC), Up: false},  // inside
		{At: time.Date(2025, 11, 1, 12, 0, 0, 0, time.UTC), Up: false}, // 133 days: outside
	}
	got := UptimeSlots(records)

	last := UptimeSlotsDays - 1
	if got[last] != SlotUp {
		t.Errorf("the newest record is the last square, got %v", got[last])
	}
	if got[last-11] != SlotDown {
		t.Errorf("eleven days earlier is eleven squares along, got %v", got[last-11])
	}
	if got[last-72] != SlotDown {
		t.Errorf("the first of January is inside the window, got %v", got[last-72])
	}
	if got[0] != SlotUnknown {
		t.Errorf("the first square has no record at all, got %v", got[0])
	}
}

func TestUptimePercentExcludesTheUnknownDays(t *testing.T) {
	// Ninety squares of a service installed on Tuesday must not read as
	// 97%: the days nobody watched are excluded from both sides.
	slots := make([]SlotState, UptimeSlotsDays)
	for i := UptimeSlotsDays - 3; i < UptimeSlotsDays; i++ {
		slots[i] = SlotUp
	}
	if got := UptimePercent(slots); got != 1 {
		t.Errorf("three good days out of three measured is 1, got %v", got)
	}

	slots[UptimeSlotsDays-2] = SlotDown
	if got := UptimePercent(slots); got != 2.0/3.0 {
		t.Errorf("two up out of three measured = %v", got)
	}

	// No record at all is not a perfect record.
	if got := UptimePercent(make([]SlotState, UptimeSlotsDays)); got != 0 {
		t.Errorf("no records is 0, not 1: got %v", got)
	}
	if got := UptimePercent(nil); got != 0 {
		t.Errorf("nothing is 0, got %v", got)
	}
}

func TestUptimeTone(t *testing.T) {
	for _, c := range []struct {
		in   float32
		want string
	}{
		{1, "Success"}, {0.9999, "Success"}, {0.999, "Success"},
		{0.9989, "Neutral"}, {0.99, "Neutral"},
		{0.989, "Warning"}, {0.95, "Warning"},
		{0.949, "Danger"}, {0, "Danger"},
	} {
		if got := UptimeTone(c.in).String(); got != c.want {
			t.Errorf("UptimeTone(%v) = %s, want %s", c.in, got, c.want)
		}
	}
}

// ── the waterfall's arithmetic ─────────────────────────────────────────────

func TestSpanBarSumsToAHundred(t *testing.T) {
	// The claim the whole view rests on: spans at 0, 30, 50 and 90 per cent
	// of a hundred milliseconds are where they say they are.
	spans := []Span{
		{Start: 0, Duration: 30 * time.Millisecond},
		{Start: 30 * time.Millisecond, Duration: 20 * time.Millisecond},
		{Start: 50 * time.Millisecond, Duration: 40 * time.Millisecond},
		{Start: 90 * time.Millisecond, Duration: 10 * time.Millisecond},
	}
	total := TotalOf(spans)
	if total != 100*time.Millisecond {
		t.Fatalf("TotalOf = %v, want 100ms", total)
	}

	wantLeft := []float32{0, 30, 50, 90}
	wantWidth := []float32{30, 20, 40, 10}
	var covered float32
	for i, s := range spans {
		left, width := SpanBar(s, total)
		if !closeTo(left, wantLeft[i]) {
			t.Errorf("span %d starts at %v, want %v", i, left, wantLeft[i])
		}
		if !closeTo(width, wantWidth[i]) {
			t.Errorf("span %d is %v wide, want %v", i, width, wantWidth[i])
		}
		covered += width
	}
	if covered != 100 {
		t.Errorf("the widths of a whole trace add to %v, want 100", covered)
	}
}

func TestSpanBarClampsRatherThanRefuses(t *testing.T) {
	total := 100 * time.Millisecond
	left, width := SpanBar(Span{Start: 90 * time.Millisecond, Duration: 50 * time.Millisecond}, total)
	if left+width > 100 {
		t.Errorf("a span past the end is drawn to the end: %v + %v", left, width)
	}
	if left != 90 {
		t.Errorf("its left edge is untouched, got %v", left)
	}

	// And a span starting before the trace begins does not draw off the left.
	left, width = SpanBar(Span{Start: -20 * time.Millisecond, Duration: 10 * time.Millisecond}, total)
	if left != 0 {
		t.Errorf("a negative start is drawn at the start, got %v", left)
	}
	if width != 10 {
		t.Errorf("its width is untouched, got %v", width)
	}
}

func TestSpanBarOfNothing(t *testing.T) {
	for _, c := range []struct {
		span  Span
		total time.Duration
	}{
		{Span{}, 100 * time.Millisecond},
		{Span{Duration: 10 * time.Millisecond}, 0},
		{Span{Start: -time.Second, Duration: 0}, time.Second},
	} {
		left, width := SpanBar(c.span, c.total)
		if left != 0 || width != 0 {
			t.Errorf("SpanBar(%+v, %v) = %v,%v, want 0,0", c.span, c.total, left, width)
		}
	}
}

func TestTotalOfTakesTheEndOfTheLastSpan(t *testing.T) {
	// Spans nest, so the total is the furthest end and not a sum: a request
	// that waited 100ms and then spent 40ms in a call took 140ms.
	spans := []Span{
		{Start: 0, Duration: 100 * time.Millisecond},
		{Start: 100 * time.Millisecond, Duration: 40 * time.Millisecond},
		{Start: 10 * time.Millisecond, Duration: 5 * time.Millisecond},
	}
	if got := TotalOf(spans); got != 140*time.Millisecond {
		t.Errorf("TotalOf = %v, want 140ms", got)
	}
	if got := TotalOf(nil); got != 0 {
		t.Errorf("no spans is no time, got %v", got)
	}
}

func TestDepthOf(t *testing.T) {
	spans := []Span{
		{Name: "GET /callbacks", Start: 0, Duration: 100 * time.Millisecond},
		{Name: "handler", Start: 2 * time.Millisecond, Duration: 95 * time.Millisecond},
		{Name: "db", Start: 10 * time.Millisecond, Duration: 40 * time.Millisecond},
		{Name: "cache", Start: 60 * time.Millisecond, Duration: 5 * time.Millisecond},
	}
	// The request contains the handler; the handler contains the database
	// call and the cache read, which are siblings rather than one inside the
	// other — so both of them are two deep, not three.
	want := []int{0, 1, 2, 2}
	for i, w := range want {
		if got := DepthOf(spans, i); got != w {
			t.Errorf("depth of %s = %d, want %d", spans[i].Name, got, w)
		}
	}
	if got := DepthOf(spans, -1); got != 0 {
		t.Errorf("an index that is not there is the top, got %d", got)
	}
}

func TestSlowestSpansSortsACopy(t *testing.T) {
	spans := []Span{
		{Name: "a", Duration: 10 * time.Millisecond},
		{Name: "b", Duration: 30 * time.Millisecond},
		{Name: "c", Duration: 20 * time.Millisecond},
	}
	got := SlowestSpans(spans, 2)
	if len(got) != 2 || got[0].Name != "b" || got[1].Name != "c" {
		t.Errorf("SlowestSpans = %v, want b and c", names(got))
	}
	if spans[0].Name != "a" {
		t.Error("the caller's own slice was reordered")
	}
}

// ── log levels and filtering ───────────────────────────────────────────────

func TestLogLevelNames(t *testing.T) {
	cases := map[LogLevel]string{
		Debug: "DEBUG", Info: "INFO", Warn: "WARN", Error: "ERROR", Fatal: "FATAL",
	}
	for l, want := range cases {
		if got := l.String(); got != want {
			t.Errorf("%d = %q, want %q", l, got, want)
		}
	}
}

func TestParseLogLevel(t *testing.T) {
	cases := map[string]LogLevel{
		"debug": Debug, "INFO": Info, "info": Info,
		"WARN": Warn, "warning": Warn,
		"ERROR": Error, "err": Error,
		"FATAL": Fatal, "critical": Fatal, "panic": Fatal,
		// An unknown name is debug, not an error: a viewer that refused to
		// draw a line because somebody wrote it differently is not a viewer
		// anybody keeps open.
		"whatever": Debug, "": Debug,
	}
	for in, want := range cases {
		if got := ParseLogLevel(in); got != want {
			t.Errorf("ParseLogLevel(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestLogLevelSeverity(t *testing.T) {
	for _, c := range []struct {
		in   LogLevel
		want string
	}{{Debug, "Neutral"}, {Info, "Neutral"}, {Warn, "Warning"},
		{Error, "Danger"}, {Fatal, "Danger"}} {
		if got := c.in.Severity().String(); got != c.want {
			t.Errorf("%v is %s, want %s", c.in, got, c.want)
		}
	}
}

func logs() []LogLine {
	at := day(2026, 14)
	return []LogLine{
		{At: at, Level: Debug, Source: "boot", Message: "listening on :8080"},
		{At: at, Level: Info, Source: "api", Message: "GET /callbacks 200"},
		{At: at, Level: Info, Source: "api", Message: "POST /callbacks 201"},
		{At: at, Level: Warn, Source: "db", Message: "slow query 420ms"},
		{At: at, Level: Error, Source: "api", Message: "POST /callbacks 500"},
		{At: at, Level: Fatal, Source: "api", Message: "out of memory"},
	}
}

func TestFilterLogsTakesTheLevelAndLouder(t *testing.T) {
	// Somebody who has asked for errors wants to see them alongside what
	// they came after, not alone.
	got := FilterLogs(logs(), Error, "", 0)
	if len(got) != 2 {
		t.Errorf("Error and louder is two lines, got %d", len(got))
	}
	for _, l := range got {
		if l.Level < Error {
			t.Errorf("a %v line was let through", l.Level)
		}
	}
}

func TestFilterLogsSearchesBothWords(t *testing.T) {
	got := FilterLogs(logs(), Debug, "callbacks", 0)
	if len(got) != 3 {
		t.Errorf("three lines mention callbacks, got %d", len(got))
	}
	// The source counts too: a request id is as often in the logger's name
	// as in its message.
	got = FilterLogs(logs(), Debug, "db", 0)
	if len(got) != 1 {
		t.Errorf("one line is from db, got %d", len(got))
	}
}

func TestFilterLogsIsCaseInsensitive(t *testing.T) {
	if got := FilterLogs(logs(), Debug, "CALLBACKS", 0); len(got) != 3 {
		t.Errorf("the search should not care about case, got %d", len(got))
	}
}

func TestFilterLogsKeepsTheNewestPastTheLimit(t *testing.T) {
	// A stream that stopped at ten thousand lines and showed Tuesday's is
	// worse than useless during an incident.
	got := FilterLogs(logs(), Debug, "", 2)
	if len(got) != 2 {
		t.Fatalf("the limit gave %d lines", len(got))
	}
	if got[0].Message != "POST /callbacks 500" || got[1].Message != "out of memory" {
		t.Errorf("the limit should keep the newest, got %q and %q",
			got[0].Message, got[1].Message)
	}
}

func TestFilterLogsDoesNotReorderTheCallersLines(t *testing.T) {
	in := logs()
	FilterLogs(in, Warn, "", 2)
	if in[0].Message != "listening on :8080" {
		t.Error("the caller's slice was reordered")
	}
}

// ── durations ──────────────────────────────────────────────────────────────

func TestRoundDuration(t *testing.T) {
	cases := []struct {
		in   time.Duration
		want string
	}{
		{0, "0"},
		{-time.Second, "0"},
		{500 * time.Microsecond, "µs"},
		{time.Millisecond, "1ms"},
		{340 * time.Millisecond, "340ms"},
		{time.Second, "1s"},
		{1250 * time.Millisecond, "1.3s"},
		{time.Minute, "1m"},
		{90 * time.Second, "1.5m"},
		{time.Hour, "1h"},
		{90 * time.Minute, "1.5h"},
	}
	for _, c := range cases {
		if got := RoundDuration(c.in); got != c.want {
			t.Errorf("RoundDuration(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestTrimFloat(t *testing.T) {
	for _, c := range []struct {
		in   float64
		want string
	}{{0, "0"}, {1, "1"}, {1.04, "1"}, {1.05, "1.1"}, {1.25, "1.3"}, {42, "42"}} {
		if got := trimFloat(c.in); got != c.want {
			t.Errorf("trimFloat(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}

// ── status codes ───────────────────────────────────────────────────────────

func TestStatusTone(t *testing.T) {
	for _, c := range []struct {
		in   string
		want string
	}{
		{"200 OK", "Success"}, {"200", "Success"}, {"204 No Content", "Success"},
		{"301 Moved", "Accent"}, {"304 Not Modified", "Accent"},
		{"400 Bad Request", "Warning"}, {"404 Not Found", "Warning"},
		{"500 Internal", "Danger"}, {"503 Unavailable", "Danger"},
		// A status line that is only words is neither: it is somebody's
		// placeholder, and colouring it would be colouring an absence.
		{"", "Neutral"}, {"OK", "Neutral"},
	} {
		if got := StatusTone(c.in).String(); got != c.want {
			t.Errorf("StatusTone(%q) = %s, want %s", c.in, got, c.want)
		}
	}
}

func TestStatusCode(t *testing.T) {
	for _, c := range []struct {
		in   string
		want int
	}{{"200 OK", 200}, {"200", 200}, {"  404 Not Found", 404}, {"", 0}, {"OK", 0}, {"999999", 0}} {
		if got := statusCode(c.in); got != c.want {
			t.Errorf("statusCode(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

// ── metrics ────────────────────────────────────────────────────────────────

func TestMetricFraction(t *testing.T) {
	for _, c := range []struct {
		m    Metric
		want float32
	}{
		{Metric{Value: 50, Limit: 100}, 0.5},
		{Metric{Value: 100, Limit: 100}, 1},
		{Metric{Value: 250, Limit: 100}, 1},
		{Metric{Value: 50, Limit: 0}, 0}, // no limit: no edge to be near
		{Metric{Value: -1, Limit: 100}, 0},
	} {
		if got := MetricFraction(c.m); got != c.want {
			t.Errorf("MetricFraction(%+v) = %v, want %v", c.m, got, c.want)
		}
	}
}

func TestMetricTone(t *testing.T) {
	for _, c := range []struct {
		m    Metric
		want string
	}{
		{Metric{Value: 50, Limit: 100}, "Neutral"},
		{Metric{Value: 75, Limit: 100}, "Accent"},
		{Metric{Value: 90, Limit: 100}, "Warning"},
		{Metric{Value: 100, Limit: 100}, "Danger"},
		{Metric{Value: 120, Limit: 100}, "Danger"},
		{Metric{Value: 50, Limit: 0}, "Neutral"},
	} {
		if got := MetricTone(c.m).String(); got != c.want {
			t.Errorf("MetricTone(%+v) = %s, want %s", c.m, got, c.want)
		}
	}
}

func TestMetricFigureReadsTheRightWayRound(t *testing.T) {
	// A cache hit rate at 20 out of 100 is "80 left", not "20 used": the bar
	// already shows how much is spent and the words should finish the
	// thought rather than start a second one.
	high := Metric{Value: 12, Limit: 50, Unit: "ms", HigherIsWorse: true}
	if got := MetricFigure(high); got != "12ms / 50ms" {
		t.Errorf("higher-is-worse = %q", got)
	}
	low := Metric{Value: 20, Limit: 100, Unit: "%"}
	if got := MetricFigure(low); got != "80% left" {
		t.Errorf("lower-is-worse = %q", got)
	}
	none := Metric{Value: 42}
	if got := MetricFigure(none); got != "42" {
		t.Errorf("no limit = %q", got)
	}
}

// ── key/value sets ─────────────────────────────────────────────────────────

func TestKeyValuesGetIsCaseInsensitive(t *testing.T) {
	kvs := KeyValues{{Key: "Content-Type", Value: "application/json"}}
	if got, ok := kvs.Get("content-type"); !ok || got != "application/json" {
		t.Errorf("Get = %q, %v", got, ok)
	}
	if _, ok := kvs.Get("accept"); ok {
		t.Error("a key that is not there was found")
	}
}

func TestKeyValuesWithReplacesInPlace(t *testing.T) {
	// Replacing in place is what keeps the order stable: a header list that
	// reordered itself between frames could not be read.
	in := KeyValues{{Key: "Accept"}, {Key: "Authorization"}, {Key: "Accept"}}
	_ = in
	kvs := KeyValues{{"a", "1"}, {"b", "2"}, {"c", "3"}}
	got := kvs.With("b", "22")
	if got[1].Value != "22" {
		t.Errorf("With should replace in place, got %v", got)
	}
	if len(got) != 3 {
		t.Errorf("With should not add a pair, got %d", len(got))
	}
	if got[0].Key != "a" || got[2].Key != "c" {
		t.Errorf("the order moved: %v", got)
	}
	// A differently-cased key replaces rather than appearing twice.
	got = kvs.With("B", "22")
	if len(got) != 3 || got[1].Value != "22" {
		t.Errorf("a re-cased key should replace: %v", got)
	}
}

func TestKeyValuesWithAddsAtTheEnd(t *testing.T) {
	kvs := KeyValues{{"a", "1"}}
	got := kvs.With("z", "26")
	if len(got) != 2 || got[1].Key != "z" {
		t.Errorf("With should append, got %v", got)
	}
	if kvs[0].Key != "a" || len(kvs) != 1 {
		t.Error("With edited the caller's own slice")
	}
}

func TestKeyValuesWithout(t *testing.T) {
	kvs := KeyValues{{"a", "1"}, {"b", "2"}, {"c", "3"}}
	got := kvs.Without("B")
	if len(got) != 2 || got[0].Key != "a" || got[1].Key != "c" {
		t.Errorf("Without = %v", got)
	}
}

func TestKeyValuesSorted(t *testing.T) {
	in := KeyValues{{"c", "3"}, {"a", "1"}, {"b", "2"}}
	got := in.Sorted()
	if got[0].Key != "a" || got[2].Key != "c" {
		t.Errorf("Sorted = %v", got)
	}
	if in[0].Key != "c" {
		t.Error("Sorted reordered the caller's slice")
	}
}

func TestKeyValueIsSecret(t *testing.T) {
	// Whole words inside a name, so a prefix or a suffix is caught without a
	// pattern per vendor — and "monkey" is not a key.
	for _, yes := range []string{
		"authorization", "Authorization", "X-Api-Key", "api_key",
		"password", "X-Password", "refresh_token", "session-cookie",
		"client_secret", "PRIVATE-KEY",
	} {
		if !(KeyValues{}).IsSecret(yes) {
			t.Errorf("%q should be secret", yes)
		}
	}
	for _, no := range []string{
		"accept", "content-type", "monkey", "key_length", "token_type",
		"keyboard", "passwords", "authorization_header",
	} {
		if (KeyValues{}).IsSecret(no) {
			t.Errorf("%q should not be secret", no)
		}
	}
}

// ── intervals ──────────────────────────────────────────────────────────────

func TestIntervalLabel(t *testing.T) {
	for _, c := range []struct {
		in   int
		want string
	}{{0, "Off"}, {5, "5s"}, {30, "30s"}, {60, "1m"}, {300, "5m"}, {90, "1.5m"}} {
		if got := intervalLabel(c.in); got != c.want {
			t.Errorf("intervalLabel(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestDefaultIntervalsStartWithOff(t *testing.T) {
	got := defaultIntervals()
	if len(got) == 0 || got[0] != 0 {
		t.Errorf("the default list should start with not refreshing, got %v", got)
	}
}

// ── ASCII helpers ──────────────────────────────────────────────────────────

func TestASCIIHelpers(t *testing.T) {
	if got := upperASCII("hello"); got != "HELLO" {
		t.Errorf("upperASCII = %q", got)
	}
	if got := upperASCII("HELLO"); got != "HELLO" {
		t.Errorf("upperASCII should not copy when there is nothing to do: %q", got)
	}
	if got := lowerASCII("HeLLo"); got != "hello" {
		t.Errorf("lowerASCII = %q", got)
	}
	if got := lowerASCII("hello"); got != "hello" {
		t.Errorf("lowerASCII should not copy when there is nothing to do: %q", got)
	}
	if !containsASCII("hello world", "lo w") {
		t.Error("containsASCII missed a substring")
	}
	if containsASCII("hello", "xyz") {
		t.Error("containsASCII found something that is not there")
	}
	if !containsASCII("hello", "") {
		t.Error("everything contains the empty string")
	}
}

// closeTo is "equal to a rounding error's worth", for the percentages a
// span's bar is placed at: they are ratios of two durations and are not
// exactly the round numbers a test writes down.
func closeTo(got, want float32) bool {
	d := got - want
	if d < 0 {
		d = -d
	}
	return d < 0.001
}

func names(spans []Span) []string {
	out := make([]string, len(spans))
	for i, s := range spans {
		out[i] = s.Name
	}
	return out
}
