package account

import (
	"strings"

	"github.com/HycJack/MintUI/ui/core"
)

// ValidateLogin checks what a sign-in form can check without asking a server,
// and returns the messages it would show. An empty string means the field is
// fine.
//
// Both answers at once, rather than one at a time, because the form shows
// them together: a person who mistyped an address should be told about the
// address in the same frame the password is judged, not on a second submit.
//
// What it does not check is whether the pair is right. That is a server's
// answer and a form that guesses it would be claiming a fact it does not have.
func ValidateLogin(email, password string) (emailErr, passwordErr string) {
	email = strings.TrimSpace(email)
	switch {
	case email == "":
		emailErr = core.Def("Enter your email address")
	case !strings.Contains(email, "@") || strings.HasPrefix(email, "@") || strings.HasSuffix(email, "@"):
		// Deliberately not a regex for a whole address. A pattern strict
		// enough to be right rejects addresses that work, and the people who
		// hit that are the ones who cannot fix their address; the shape
		// check catches the typos and lets the server decide the rest.
		emailErr = core.Def("That does not look like an email address")
	}

	switch {
	case password == "":
		passwordErr = core.Def("Enter your password")
	case len([]rune(password)) < 8:
		// Eight, not "strong": a rule the reader cannot satisfy is a rule
		// they work around, and the length is the one part of a password
		// the reader can actually control.
		passwordErr = core.Def("Passwords are at least 8 characters")
	}
	return emailErr, passwordErr
}

// QuotaFraction is how much of a quota is used, 0 to 1.
//
// A limit of zero or less means the plan does not meter this, which is not
// the same as having used none of it. It returns 0 rather than dividing,
// because a bar that says a quarter used when nothing is metered is a claim
// about the account that the account did not make.
func QuotaFraction(used, limit int) float32 {
	if limit <= 0 || used <= 0 {
		return 0
	}
	f := float32(used) / float32(limit)
	if f > 1 {
		// Over the limit is drawn as a full bar. The number beside it is
		// where the real figure is read; the bar only has to say the quota
		// is spent.
		return 1
	}
	return f
}

// QuotaTone is the severity a quota is drawn at, which is the one place in
// this package where a decision about colour is made about numbers.
//
// The steps are where somebody would actually do something. Nothing is said
// about the first three quarters of a quota — a bar that turns amber at 50%
// is a bar nobody reads any more — and the last fifth is where a plan runs
// out of work at the worst possible moment.
func QuotaTone(used, limit int) core.Severity {
	if limit <= 0 || used <= 0 {
		return core.Neutral
	}
	switch {
	case used >= limit:
		return core.Danger
	case float32(used)/float32(limit) >= 0.9:
		return core.Warning
	case float32(used)/float32(limit) >= 0.75:
		return core.Accent
	default:
		return core.Neutral
	}
}

// Setting is one row of a settings screen: enough to search for and to
// navigate by, and nothing about how the panel behind it is drawn.
type Setting struct {
	// ID is what the screen is addressed by, and what a deep link carries.
	ID string
	// Name is what a person reads.
	Name string
	// Hint is the sentence under the name, and is searched along with it:
	// the words somebody remembers are usually the explanation, not the
	// title.
	Hint string
	// Keywords are the words that are in neither, which is how "sync" finds
	// "Refresh when the app is opened" without either having to be worded
	// for the search.
	Keywords []string
}

// MatchSettings is the rows a settings search shows for query, best first.
//
// An empty query returns everything in the order it was given, because a
// search that has not been typed has no opinion and reordering the list under
// somebody who has not typed anything looks like the screen reordering itself.
//
// The ranking is three tiers, and it is deliberately crude: a row whose name
// starts with the query, then one whose name contains it, then one that only
// mentions it in its hint or its keywords. The tiers are ordered by how
// surprised somebody would be to see a row where it is, and rows within a tier
// keep the order they were given, so a settings screen does not reshuffle
// itself as somebody narrows a search down.
func MatchSettings(query string, items []Setting) []Setting {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		out := make([]Setting, len(items))
		copy(out, items)
		return out
	}

	var first, second, third []Setting
	for _, it := range items {
		name := strings.ToLower(it.Name)
		switch {
		case strings.HasPrefix(name, q):
			first = append(first, it)
		case strings.Contains(name, q):
			second = append(second, it)
		case mentions(it, q):
			third = append(third, it)
		}
	}
	out := make([]Setting, 0, len(items))
	out = append(out, first...)
	out = append(out, second...)
	out = append(out, third...)
	return out
}

// mentions reports whether the query is anywhere in a row except at the start
// of its name.
func mentions(it Setting, q string) bool {
	if strings.Contains(strings.ToLower(it.Hint), q) {
		return true
	}
	for _, kw := range it.Keywords {
		if strings.Contains(strings.ToLower(kw), q) {
			return true
		}
	}
	return false
}
