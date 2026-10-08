package data

import (
	"strconv"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// Section is one section of an Accordion: what it is called, whether it is
// open, and what it holds.
type Section struct {
	// Title heads the section, and is also what names it for assistive
	// technology.
	Title string
	// Open is the section's open or closed, in the caller's state — the
	// whole of it. A component keeping its own would be a section that
	// opened and closed at a time of its own choosing.
	Open *bool
	// Meta is a count or a note at the other end of the head.
	Meta string
	// Body is what the section holds while it is open.
	Body func()
}

// AccordionOptions configure an Accordion.
type AccordionOptions struct {
	// Single opens one section at a time: opening one closes the others.
	// It can, because each section's Open is a bool in the caller's state
	// and the accordion is allowed to write it — the caller sees which
	// section it toggled and may do as it likes with the rest.
	Single bool
}

// AccordionResult carries an Accordion and what the user did with it.
type AccordionResult struct {
	// Element is the whole accordion.
	Element *ui.Element
	// toggled is the section whose head was pressed this frame, and was
	// says whether there was one.
	toggled int
	was     bool
}

// Toggled returns the section whose head was pressed this frame, and whether
// one was.
func (r AccordionResult) Toggled() (int, bool) { return r.toggled, r.was }

// Accordion is sections in one bordered box, one above the other, each
// opening and closing on its own: the settings page, the groups of a long
// form, the details under a summary.
//
// It holds no section's state — each section's Open is a bool the caller
// pointed at — so the sections drawn open are the sections the caller has
// open, and a caller that wants one at a time says so in Single.
func Accordion(c *ui.Context, opts AccordionOptions, sections ...Section) AccordionResult {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if len(sections) == 0 {
		panic("data: Accordion needs at least one section")
	}
	for i, s := range sections {
		if s.Open == nil {
			panic("data: Accordion section " + strconv.Itoa(i) + " needs the Open bool to point at")
		}
		if s.Title == "" {
			panic("data: Accordion section " + strconv.Itoa(i) + " needs a title")
		}
	}

	var res AccordionResult
	res.Element = ui.Column(c).FillWidth().Shrink(0).Radius(theme.ControlRadius).
		Border(theme.BorderWidth, k.Border).Clip().
		Children(func() {
			for i, s := range sections {
				part := ui.Column(c).FillWidth().Shrink(0)
				if i > 0 {
					// One rule between the sections and none around them: the
					// box is one thing, and a rule inside a border reads as two.
					part.Children(func() { layout.Divider(c, layout.DividerOptions{}) })
				}
				part.Children(func() {
					parts := ui.CollapsibleBase(c, s.Open)
					head := disclosureHead(c, parts, s.Title, s.Meta, k.Text, u*2.5, u*3)
					if head.Changed() {
						res.toggled, res.was = i, true
						// The others close behind the frame: those above this
						// one have already been built, and the frame after
						// shows what this press did to them.
						if opts.Single && *s.Open {
							for j, other := range sections {
								if j != i {
									*other.Open = false
								}
							}
						}
					}
					if panel := parts.Panel(s.Body); panel != nil {
						// In from the head, so that what is in the panel reads
						// as the head's contents rather than as the next
						// section's.
						panel.Padding(u*2.5, u*3, u*3, u*3).Gap(u * 2)
					}
				})
			}
		})
	return res
}

// CollapsibleOptions configure a Collapsible.
type CollapsibleOptions struct {
	// Title is what the disclosure is called, and what names it for
	// assistive technology.
	Title string
	// Open is whether the content is showing, in the caller's state.
	Open *bool
	// Meta is a count or a note at the other end of the head.
	Meta string
	// Rule draws a hairline under the head, for a block that sits among
	// others and would otherwise run into the next.
	Rule bool
}

// CollapsibleResult carries a Collapsible and what the user did with it.
type CollapsibleResult struct {
	// Element is the whole block.
	Element *ui.Element
	// toggled reports a press of the head this frame.
	toggled bool
}

// Toggled reports a press of the head this frame.
func (r CollapsibleResult) Toggled() bool { return r.toggled }

// Collapsible is one block that folds away: a head that opens and closes
// what is below it.
//
// It is an Accordion of one, without the box: for the details under a
// summary, or an advanced block in a form, where a border around the only
// section would say "here is a group" when there is no group.
func Collapsible(c *ui.Context, opts CollapsibleOptions, body func()) CollapsibleResult {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if opts.Open == nil {
		panic("data: Collapsible needs the Open bool to point at")
	}
	if opts.Title == "" {
		panic("data: Collapsible needs a title to put on its head")
	}
	parts := ui.CollapsibleBase(c, opts.Open)

	var res CollapsibleResult
	res.Element = ui.Column(c).FillWidth().Shrink(0).Children(func() {
		head := disclosureHead(c, parts, opts.Title, opts.Meta, k.Text, u*2, u*1)
		res.toggled = head.Changed()
		if opts.Rule {
			layout.Divider(c, layout.DividerOptions{})
		}
		if panel := parts.Panel(body); panel != nil {
			panel.Padding(0, u, u*2, u).Gap(u * 2)
		}
	})
	return res
}

// disclosureHead builds the head of a collapsible: what it is called,
// whatever else is said of it, and the mark that opens and closes it.
//
// The mark points right while the block is closed and down while it is
// open, which is the same way a tree's arrow turns and the same reason: one
// mark that turns reads as one thing moving, where two marks swapped in
// read as a flicker.
func disclosureHead(c *ui.Context, parts ui.CollapsibleParts, title, meta string, ink ui.Color, padY, padX float32) *ui.Element {
	u := core.Density(c).Unit()
	head := parts.Trigger.AlignItems(ui.Center).Gap(u*2).Padding(padY, padX).
		TextColor(ink)
	if head.Hovered() {
		head.Background(core.Tokens(c).SurfaceHover)
	}
	head.Children(func() {
		ui.Text(c, title).TextColor(ink).
			FontSize(core.FontSize(c, theme.BodySize)).Bold().Grow(1).SingleLine()
		if meta != "" {
			ui.Text(c, meta).TextColor(core.Tokens(c).TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize))
		}
		chevron(c, u*4, 90*parts.Progress(), core.Tokens(c).TextMuted)
	})
	return head
}
