package account

import (
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/feedback"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/theme"
)

// Quota is one metered thing and how much of it is spent: requests, seats,
// gigabytes. The caller supplies the numbers already counted; this decides
// what colour the bar is and what the sentence under it says.
type Quota struct {
	// ID is the quota's name in the caller's own model, so a row that
	// changes is the row that changed.
	ID string
	// Name is what the row is called.
	Name string
	// Used and Limit are the figures; a Limit of zero or less means the plan
	// does not meter this, which QuotaFraction reads as "nothing spent"
	// rather than as a division by zero.
	Used, Limit int
	// Resets is when the allowance comes back, already formatted. Empty
	// draws no line under the bar rather than an empty one.
	Resets string
}

// UsageQuotaOptions configure a UsageQuota.
type UsageQuotaOptions struct {
	// Quotas are the rows, in the caller's order. Nothing here sorts them:
	// a plan page has an order — the thing somebody is about to run out of
	// first belongs at the top — and that is the caller's to decide.
	Quotas []Quota
	// Title heads the panel; empty draws no heading.
	Title string
	// Upgrade draws a button under the rows. Nil draws none, which is right
	// on the account of somebody already on the plan that has everything.
	Upgrade func()
}

// UsageQuotaResult carries a UsageQuota and what was done in it.
type UsageQuotaResult struct {
	// Element is the whole panel.
	Element *ui.Element
}

// UsageQuota is how much of a plan is gone, as one bar per metered thing.
//
// The bar is a feedback.Progress, so it has the library's own role, its own
// focus behaviour and its own reduced-motion handling. The tone comes from
// QuotaTone rather than from anything decided here, which is what keeps the
// "nothing said until the last fifth" rule in one place instead of spread
// across the plan page and the account menu and the paywall.
//
// A row for something unmetered draws a bar at zero and says so in words. An
// empty row would read as "no usage", which is a different claim from "this
// is not counted".
func UsageQuota(c *ui.Context, opts UsageQuotaOptions) UsageQuotaResult {
	k, u := core.Tokens(c), core.Density(c).Unit()

	var r UsageQuotaResult
	r.Element = ui.Column(c).FillWidth().Gap(u * 4).
		Label(core.Msg(c, "account.usage", "Usage")).Children(func() {
		if opts.Title != "" {
			ui.Text(c, opts.Title).TextColor(k.Text).Bold().
				FontSize(core.FontSize(c, theme.SheetSize))
		}
		for _, q := range opts.Quotas {
			quotaRow(c, q)
		}
		if opts.Upgrade != nil {
			ui.Row(c).FillWidth().Justify(ui.End).Children(func() {
				if input.Button(c, core.Msg(c, "account.upgrade", "Upgrade plan"),
					input.ButtonOptions{Primary: true}).Clicked() {
					opts.Upgrade()
				}
			})
		}
	})
	return r
}

// quotaRow is one metered thing.
func quotaRow(c *ui.Context, q Quota) {
	k, u := core.Tokens(c), core.Density(c).Unit()

	ui.Column(c).FillWidth().Gap(u * 1.5).Label(q.Name).Children(func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 2).Children(func() {
			ui.Text(c, q.Name).TextColor(k.Text).Grow(1).
				FontSize(core.FontSize(c, theme.RowSize)).SingleLine()
			ui.Text(c, quotaFigure(c, q)).TextColor(k.TextMuted).Shrink(0).
				FontSize(core.FontSize(c, theme.MetaSize)).SingleLine()
		})
		feedback.Progress(c, feedback.ProgressOptions{
			Label:    q.Name,
			Value:    QuotaFraction(q.Used, q.Limit),
			Severity: QuotaTone(q.Used, q.Limit),
		})
		if q.Resets != "" {
			ui.Text(c, q.Resets).TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
		}
	})
}

// quotaFigure is the "1,204 of 2,000" line beside a row's name. An unmetered
// thing says so instead of showing a zero against no limit, which would be a
// figure about a quota that does not exist.
func quotaFigure(c *ui.Context, q Quota) string {
	if q.Limit <= 0 {
		return core.Msg(c, "account.unmetered", "Not metered")
	}
	return commas(q.Used) + " / " + commas(q.Limit)
}

// Session is one signed-in device.
type Session struct {
	// ID is what a revoke is addressed by.
	ID string
	// Device is what the row calls it — "MacBook Pro", "iPhone 15". It is
	// the caller's string because only the caller knows the model names its
	// users recognise.
	Device string
	// Where is where it signed in from: a city, an address, an IP. Shown
	// under the device because it is the line that answers "was that me?",
	// and it is the line a person will not recognise the device by.
	Where string
	// Since is when the session began, already formatted.
	Since string
	// Current marks the session the window is being drawn from. It is drawn
	// as a tint and carries no revoke button: revoking the session you are
	// using signs you out mid-frame, which is a different action with a
	// different name.
	Current bool
	// Expired marks a session that has already lapsed. It is greyed out and
	// has no revoke button, because there is nothing left to revoke.
	Expired bool
}

// SessionListOptions configure a SessionList.
type SessionListOptions struct {
	// Sessions are the account's sessions, newest first as the caller has
	// them. Nothing here sorts them: a session list is chronological and the
	// caller's order is the order somebody wrote them down in.
	Sessions []Session
	// Revoke writes the id of the revoked session, empty when none was.
	Revoke *string
	// SignOutAll is pressed for the button under the list, asked separately
	// because it is the one action every caller has and the one a test wants
	// to be able to reach without setting up a row first.
	SignOutAll func()
}

// SessionListResult carries a SessionList and what was done in it.
type SessionListResult struct {
	// Element is the whole list.
	Element *ui.Element
}

// SessionList is where an account is signed in: one row per device, with the
// way to end any session but the current one.
//
// The current session is tinted and has no revoke button. It is not disabled
// — a disabled control that still takes a press is the bug in
// docs/design-system.md §17.3 — it simply has nothing to offer, because the
// action on it would be "sign yourself out of the window you are reading",
// which has its own name and its own confirmation.
//
// An expired session is greyed out and shown, not hidden. A list that quietly
// drops them looks like the sign-in worked on fewer devices than it did, and
// somebody checking whether a stranger has access needs the opposite of that.
func SessionList(c *ui.Context, opts SessionListOptions) SessionListResult {
	if opts.Revoke == nil {
		panic("account: SessionList needs the *string Revoke writes to")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	var r SessionListResult
	r.Element = ui.Column(c).FillWidth().Gap(u * 3).
		Label(core.Msg(c, "account.sessions", "Signed-in devices")).Children(func() {
		if len(opts.Sessions) == 0 {
			ui.Box(c).FillWidth().Padding(u*4, u*3).Radius(theme.ControlRadius).
				Background(k.Surface).Center().Children(func() {
				ui.Text(c, core.Msg(c, "account.noSessions", "No other devices")).
					TextColor(k.TextMuted).FontSize(core.FontSize(c, theme.RowSize))
			})
		}
		for _, s := range opts.Sessions {
			sessionRow(c, opts, s)
		}
		if opts.SignOutAll != nil {
			ui.Row(c).FillWidth().Justify(ui.End).Margin(0, u, 0, 0).Children(func() {
				if input.Button(c, core.Msg(c, "account.signOutAll", "Sign out everywhere"),
					input.ButtonOptions{Danger: true}).Clicked() {
					opts.SignOutAll()
				}
			})
		}
	})
	return r
}

// sessionRow is one device.
func sessionRow(c *ui.Context, opts SessionListOptions, s Session) {
	k, u := core.Tokens(c), core.Density(c).Unit()

	row := ui.Row(c).FillWidth().Gap(u*3).AlignItems(ui.Center).
		Padding(u*2, u*3).Radius(theme.ControlRadius).
		Label(s.Device).Role(ui.RoleListItem)
	switch {
	case s.Current:
		row.Background(k.SurfaceHover)
	case s.Expired:
		row.Opacity(0.55)
	default:
		row.Background(k.Surface)
	}
	row.Children(func() {
		ui.Column(c).Grow(1).Shrink(0).Gap(u * 0.25).Children(func() {
			ui.Text(c, s.Device).TextColor(sessionInk(k, s)).
				FontSize(core.FontSize(c, theme.RowSize)).Bold().SingleLine()
			ui.Text(c, sessionLine(c, s)).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
		})
		// The badge comes before the button rather than after it, because a
		// count of what is going on is context and the button is the action;
		// the eye should reach the action after reading the sentence, not
		// before.
		if s.Current {
			display.Tag(c, core.Msg(c, "account.thisDevice", "This device"),
				display.TagOptions{Tone: core.Success})
		}
		if s.Expired {
			display.Tag(c, core.Msg(c, "account.expired", "Expired"),
				display.TagOptions{})
			return
		}
		if !s.Current && input.Button(c, core.Msg(c, "account.signOutDevice", "Sign out"),
			input.ButtonOptions{Danger: true}).Clicked() {
			*opts.Revoke = s.ID
		}
	})
}

// sessionInk is a row's device name in the right tone: faint for a session
// that has lapsed, ordinary otherwise.
func sessionInk(k theme.Tokens, s Session) ui.Color {
	if s.Expired {
		return k.TextFaint
	}
	return k.Text
}

// sessionLine is where and when, joined into the one line somebody reads to
// decide whether a device is theirs.
func sessionLine(c *ui.Context, s Session) string {
	switch {
	case s.Where != "" && s.Since != "":
		return s.Where + " · " + s.Since
	case s.Where != "":
		return s.Where
	default:
		return s.Since
	}
}

// since is the line under a login record's device: how long ago it happened,
// in the words a person would use rather than in a timestamp.
func since(at, now time.Time) string {
	d := now.Sub(at)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return itoa(int(d/time.Minute)) + " min ago"
	case d < 24*time.Hour:
		return itoa(int(d/time.Hour)) + " h ago"
	default:
		return itoa(int(d/(24*time.Hour))) + " d ago"
	}
}

// commas is an integer with thousands separators, written out here because
// internal.Commas is in the package that draws and this package would rather
// not depend on it for one line.
func commas(n int) string {
	s := itoa(n)
	neg := false
	if len(s) > 0 && s[0] == '-' {
		neg, s = true, s[1:]
	}
	var out []byte
	for i := 0; i < len(s); i++ {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, s[i])
	}
	if neg {
		return "-" + string(out)
	}
	return string(out)
}
