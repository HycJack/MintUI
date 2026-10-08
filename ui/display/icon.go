package display

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// IconName is the name of one of the glyphs the library ships. It is a string
// rather than a *ui.SVG so that a component taking an icon takes a name: the
// name is what a screen reader is given, and an SVG has none.
//
// The set is the vocabulary of a callbacks board — where you go, what it is
// called, and the few marks that mean a thing needs attention. A name outside
// it panics rather than drawing nothing, because a blank square in a sidebar
// is a bug that reads as a design choice.
type IconName string

// The glyphs the library ships.
const (
	IconOverview   IconName = "overview"
	IconCallbacks  IconName = "callbacks"
	IconCustomers  IconName = "customers"
	IconTeam       IconName = "team"
	IconInventory  IconName = "inventory"
	IconJobs       IconName = "jobs"
	IconReports    IconName = "reports"
	IconBell       IconName = "bell"
	IconSettings   IconName = "settings"
	IconProfile    IconName = "profile"
	IconPanel      IconName = "panel"
	IconSliders    IconName = "sliders"
	IconPlus       IconName = "plus"
	IconSearch     IconName = "search"
	IconFilter     IconName = "filter"
	IconCalendar   IconName = "calendar"
	IconClock      IconName = "clock"
	IconCheck      IconName = "check"
	IconDismiss    IconName = "dismiss"
	IconChevronUp  IconName = "chevron-up"
	IconChevron    IconName = "chevron-right"
	IconDownload   IconName = "download"
	IconRefresh    IconName = "refresh"
	IconTrash      IconName = "trash"
	IconEdit       IconName = "edit"
	IconCopy       IconName = "copy"
	IconWarning    IconName = "warning"
	IconStar       IconName = "star"
	IconExternal   IconName = "external"
	IconCollapse   IconName = "collapse"
	IconExpand     IconName = "expand"
	IconPhone      IconName = "phone"
	IconMapPin     IconName = "map-pin"
	IconPaperclip  IconName = "paperclip"
	IconLinkBroken IconName = "link-broken"
)

// iconShapes is the built-in set, in the same 24×24 stroke style as the rest
// of the interface: one weight, one corner style, so a glyph next to another
// is two marks of the same hand rather than two clip arts.
//
// It is a map and not a switch because the names are data — a caller reading
// Icon("callbacks") should not have to read a hundred lines to find out what
// it draws.
var iconShapes = map[IconName]string{
	IconOverview: `<rect x="3.5" y="3.5" width="6.5" height="6.5" rx="2"/>` +
		`<rect x="14" y="3.5" width="6.5" height="6.5" rx="2"/>` +
		`<rect x="3.5" y="14" width="6.5" height="6.5" rx="2"/>` +
		`<rect x="14" y="14" width="6.5" height="6.5" rx="2"/>`,
	IconCallbacks: `<path d="M6.5 3.5h3l1.5 3-1.8 1.6a11.5 11.5 0 0 0 5.2 5.2l1.6-1.8 3 1.5v3a2 2 0 0 1-2.2 2A15.5 15.5 0 0 1 4.5 5.7 2 2 0 0 1 6.5 3.5Z"/>`,
	IconCustomers: `<circle cx="12" cy="6" r="2.6"/><circle cx="6" cy="17" r="2.6"/>` +
		`<circle cx="18" cy="17" r="2.6"/><path d="M12 8.6v4.2M10 13.4 8 14.9M14 13.4l2 1.5"/>`,
	IconTeam: `<circle cx="9" cy="8.5" r="3"/>` +
		`<path d="M3.5 19.5c.6-3.2 2.8-4.8 5.5-4.8s4.9 1.6 5.5 4.8"/>` +
		`<path d="M16 6.2a3 3 0 0 1 0 5.6M17.5 15.2c2 .7 3.2 2.2 3.6 4.3"/>`,
	IconInventory: `<ellipse cx="12" cy="6" rx="7" ry="2.8"/>` +
		`<path d="M5 6v12c0 1.5 3.1 2.8 7 2.8s7-1.3 7-2.8V6"/>` +
		`<path d="M5 12c0 1.5 3.1 2.8 7 2.8s7-1.3 7-2.8"/>`,
	IconJobs: `<path d="M8 4h9a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2Z"/>` +
		`<path d="m9.5 13 1.8 1.8 3.4-3.8"/><path d="M9 8h5"/>`,
	IconReports: `<rect x="4" y="12" width="3.2" height="8" rx="1.2"/>` +
		`<rect x="10.4" y="7" width="3.2" height="13" rx="1.2"/>` +
		`<rect x="16.8" y="4" width="3.2" height="16" rx="1.2"/>`,
	IconBell: `<path d="M6.5 17.5V11a5.5 5.5 0 0 1 11 0v6.5l1.6 2H4.9l1.6-2Z"/>` +
		`<path d="M10 20a2.2 2.2 0 0 0 4 0"/>`,
	IconSettings: `<circle cx="12" cy="12" r="3"/>` +
		`<path d="M12 3.5v2.2M12 18.3v2.2M20.5 12h-2.2M5.7 12H3.5M18 6l-1.6 1.6M7.6 16.4 6 18M18 18l-1.6-1.6M7.6 7.6 6 6"/>`,
	IconProfile: `<circle cx="12" cy="9" r="3.4"/><path d="M5 19.5c1-3.6 3.6-5.4 7-5.4s6 1.8 7 5.4"/>`,
	IconPanel: `<rect x="3.5" y="4.5" width="17" height="15" rx="3"/>` +
		`<path d="M9.5 4.5v15"/><path d="m16 10-2.5 2 2.5 2"/>`,
	IconSliders: `<path d="M4 6h9M17 6h3M4 12h3M11 12h9M4 18h7M15 18h5"/>` +
		`<circle cx="15" cy="6" r="2"/><circle cx="9" cy="12" r="2"/><circle cx="13" cy="18" r="2"/>`,
	IconPlus:      `<path d="M12 5.5v13M5.5 12h13"/>`,
	IconSearch:    `<circle cx="11" cy="11" r="6"/><path d="m15.5 15.5 4 4"/>`,
	IconFilter:    `<path d="M4 6h16M7 12h10M10 18h4"/>`,
	IconCalendar:  `<rect x="3.5" y="5.5" width="17" height="15" rx="3"/><path d="M3.5 10h17M8 3.5v4M16 3.5v4"/>`,
	IconClock:     `<circle cx="12" cy="12" r="8"/><path d="M12 7.5V12l3 2"/>`,
	IconCheck:     `<path d="m5 12.5 4.5 4.5L19 7"/>`,
	IconDismiss:   `<path d="m6.5 6.5 11 11M17.5 6.5l-11 11"/>`,
	IconChevronUp: `<path d="m6.5 14.5 5.5-5.5 5.5 5.5"/>`,
	IconChevron:   `<path d="m9.5 6.5 5.5 5.5-5.5 5.5"/>`,
	IconDownload:  `<path d="M12 4v11M7.5 10.5 12 15l4.5-4.5"/><path d="M5 19.5h14"/>`,
	IconRefresh:   `<path d="M19 12a7 7 0 1 1-2.2-5.1"/><path d="M19.5 4v4h-4"/>`,
	IconTrash:     `<path d="M5.5 7h13M10 7V5h4v2M7 7l1 12.5h8L17 7"/>`,
	IconEdit:      `<path d="M4.5 19.5h4L19 9a2.1 2.1 0 0 0-3-3L5.5 16.5Z"/>`,
	IconCopy:      `<rect x="8.5" y="8.5" width="11" height="11" rx="2.5"/><path d="M15.5 5.5H6.5a2 2 0 0 0-2 2v9"/>`,
	IconWarning:   `<path d="M12 4.5 21 20H3Z"/><path d="M12 10v4.5M12 17.2v.1"/>`,
	IconStar:      `<path d="m12 4 2.5 5.2 5.5.8-4 3.9 1 5.6-5-2.7-5 2.7 1-5.6-4-3.9 5.5-.8Z"/>`,
	IconExternal: `<path d="M14 4.5h5.5V10"/><path d="m19.5 4.5-8 8"/>` +
		`<path d="M18 14v4.5a1.5 1.5 0 0 1-1.5 1.5H6a1.5 1.5 0 0 1-1.5-1.5V8A1.5 1.5 0 0 1 6 6.5h4"/>`,
	IconCollapse: `<rect x="3.5" y="4.5" width="17" height="15" rx="3"/>` +
		`<path d="M9.5 4.5v15"/><path d="m16 10-2.5 2 2.5 2"/>`,
	IconExpand: `<rect x="3.5" y="4.5" width="17" height="15" rx="3"/>` +
		`<path d="M14.5 4.5v15"/><path d="m11 10-2.5 2 2.5 2"/>`,
	IconPhone:     `<path d="M6 4h3l1.5 4-2 1.4a12 12 0 0 0 6.1 6.1L16 13.5l4 1.5v3a2 2 0 0 1-2.2 2A15.5 15.5 0 0 1 4 6.2 2 2 0 0 1 6 4Z"/>`,
	IconMapPin:    `<path d="M12 21s6.5-6 6.5-10.5a6.5 6.5 0 0 0-13 0C5.5 15 12 21 12 21Z"/><circle cx="12" cy="10.5" r="2.4"/>`,
	IconPaperclip: `<path d="M17 8.5 10 15.5a3 3 0 0 0 4.2 4.2l7-7a4.5 4.5 0 0 0-6.4-6.4l-7 7a6 6 0 0 0 8.5 8.5"/>`,
	IconLinkBroken: `<path d="M9.5 14.5 14.5 9.5"/>` +
		`<path d="M13 7 15 5a3.5 3.5 0 0 1 5 5l-2 2"/>` +
		`<path d="M11 17 9 19a3.5 3.5 0 0 1-5-5l2-2"/>`,
}

// Icon returns the built-in glyph called name.
//
// A glyph is the one element in this library that has no word of its own, so
// Name is required rather than optional: an icon nobody named is a shape a
// screen reader has to describe as "image" and a test has nothing to find.
// Where the glyph sits beside a label, pass the same words as Name — the
// label already says them, and a name that disagrees with what is drawn is
// worse than none.
func Icon(c *ui.Context, name IconName, opts IconOptions) *ui.Element {
	shapes, ok := iconShapes[name]
	if !ok {
		panic("display: unknown icon " + string(name))
	}
	if opts.Name == "" {
		panic("display: Icon needs a Name; a glyph is the one element with no word of its own")
	}
	size := opts.Size
	if size <= 0 {
		size = theme.IconSize
	}
	k := core.Tokens(c)
	// Body ink by default, the secondary tone beside unselected text, and
	// the tone's own foreground for a glyph that means something is wrong —
	// all three read out of the palette, so no component picks a colour.
	col := k.Text
	switch {
	case opts.Muted:
		col = k.TextMuted
	case opts.Tone != core.Neutral:
		_, col = opts.Tone.Pair(k)
	}
	return ui.Icon(c, mustIcon(shapes)).Size(size, size).TextColor(col).Label(opts.Name)
}

// IconOptions configure an Icon.
type IconOptions struct {
	// Size is the glyph's side in DIPs; zero gives the navigation size, so
	// an icon in a button and one in a tab are the same mark unless a caller
	// says otherwise.
	Size float32
	// Tone picks the ink from the severity ramp. Neutral is the quiet
	// default, and the only tone a glyph in a row of rows should wear.
	Tone core.Severity
	// Muted draws in the secondary tone, which is what a glyph beside
	// unselected text should be so the two read as one line.
	Muted bool
	// Name is what the glyph is called out loud. It is required.
	Name string
}

// mustIcon parses one of the built-in glyphs. They are compile-time constants
// in effect: a malformed one is a bug in this file, not something a caller
// could cause, so it panics here rather than becoming an empty square.
func mustIcon(shapes string) *ui.SVG {
	return ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" ` +
		`fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" ` +
		`stroke-linejoin="round">` + shapes + `</svg>`))
}
