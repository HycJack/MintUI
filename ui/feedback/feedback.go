// Package feedback holds what an application says rather than shows: that it
// has nothing, that something went wrong, that a piece of work is halfway
// done, and that there is still more coming.
//
// The components fall into five groups:
//
//	Progress, StatusIndicator, Presence   where a thing stands right now
//	Alert, Result                         what happened, and what to do
//	ActivityFeed, NotificationCenter      what has happened, and is outstanding
//	Skeleton, Shimmer, the loaders        something is on its way
//	BlinkHighlight, LayoutTransition, Stagger, CountUp, Typewriter,
//	LoadingOverlay, UndoToast, InterruptButton    the moving parts
//
// Nothing here holds state. A component draws what it is given this frame and
// reports what was pressed in it, so the caller stays the only thing that
// knows when a number moved, when a line was interrupted, and when a toast
// should go away.
package feedback

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// EmptyOptions configure an Empty.
type EmptyOptions struct {
	// Title says what is missing.
	Title string
	// Body explains what would fill it, in one sentence.
	Body string
	// Action names the button that would fill it; nil draws none.
	Action string
	// Art draws the illustration above the title. nil draws the default:
	// a few tiles and a bright action dot, which says "a thing goes here"
	// without pretending to be a picture of one.
	Art func()
}

// EmptyResult carries an Empty and the press of its action.
type EmptyResult struct {
	// Element is the whole empty state.
	Element *ui.Element
	pressed bool
}

// Pressed reports the action button being pressed.
func (r EmptyResult) Pressed() bool { return r.pressed }

// Empty is what a surface shows when it has nothing to show, with the one
// action that would change that.
//
// It is a card: it sits on the surface rather than in it, so a column holding
// only an Empty reads as a column with a panel in it, not as a broken column.
func Empty(c *ui.Context, opts EmptyOptions) EmptyResult {
	k, u := core.Tokens(c), core.Density(c).Unit()
	var r EmptyResult

	card := ui.Column(c).FillWidth().Grow(1).Radius(theme.CardRadius).
		Background(k.Background).Border(theme.BorderWidth, k.Border).
		Padding(u*5.5, u*4.5, u*5.5, u*4.5).Gap(u * 2.5).Role(ui.RoleGroup)
	if opts.Title != "" {
		card.Label(opts.Title)
	}

	card.Children(func() {
		ui.Box(c).FillWidth().Height(u * 37.5).Radius(theme.ControlRadius).
			Background(k.Surface).Center().Children(func() {
			if opts.Art != nil {
				opts.Art()
			} else {
				defaultArt(c)
			}
		})
		if opts.Title != "" {
			ui.Text(c, opts.Title).TextColor(k.Text).FontSize(theme.LeadSize).Bold()
		}
		if opts.Body != "" {
			ui.Text(c, opts.Body).TextColor(k.TextMuted).FontSize(theme.MetaSize)
		}
		if opts.Action != "" {
			action := ui.Button(c, "").FillWidth().Height(u * 11.5).
				Radius(theme.PillRadius).Background(k.Fill).TextColor(k.OnFill).
				Label(opts.Action).Children(func() {
				ui.Text(c, "+").FontSize(theme.RowSize)
				ui.Text(c, opts.Action)
			})
			r.pressed = action.Clicked()
		}
	})

	r.Element = card
	return r
}

// defaultArt draws stacked tiles falling away from a bright dot: the shape of
// "these would go here", without a picture of any one of them.
func defaultArt(c *ui.Context) {
	u := core.Density(c).Unit()
	dot := u * 10.5
	ui.Box(c).Children(func() {
		ui.Box(c).Size(dot, dot).Radius(dot / 2).Background(k2(c).Lively).Center().
			Children(func() {
				ui.Text(c, "+").TextColor(k2(c).Background).FontSize(22).Bold()
			})
		// The tiles, drawn under the dot rather than beside it, so the pair
		// reads as one mark.
		ui.Box(c).Absolute().Draw(func(p *ui.Painter, r ui.Rect) {
			internal.Tiles(p, ui.Rect{
				X: r.X - dot*1.6, Y: r.Y - u*5, W: dot * 1.5, H: u * 6.5,
			}, []float32{1, 0.85, 0.7}, u*2.25, u*2,
				p.Theme().TextMuted.Alpha(0.22))
		})
	})
}

func k2(c *ui.Context) theme.Tokens { return core.Tokens(c) }

// Toast draws a message across the foot of a window. It reports nothing and
// dismisses itself: the caller sets ToastMessage and lets it expire.
func Toast(c *ui.Context, message string, u float32) *ui.Element {
	k := core.Tokens(c)
	if message == "" {
		return ui.Box(c)
	}
	return ui.Box(c).Padding(u*2.75, u*5, u*2.75, u*5).Radius(theme.PillRadius).
		Background(k.Fill).TextColor(k.OnFill).Label(message).Shadow(
		0, u*1.5, u*4.5, 0, ui.RGBA(0, 0, 0, 0.18)).Children(func() {
		ui.Text(c, message).TextColor(k.OnFill).FontSize(theme.RowSize)
	})
}
