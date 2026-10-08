package messaging

import (
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/theme"
)

// IncomingCallDialogOptions configure an IncomingCallDialog.
type IncomingCallDialogOptions struct {
	// Avatar is the caller's picture, the caller's bitmap. nil draws the
	// initials of the name instead, as a tile of a camera that is off: a
	// black square would read as a call that has failed, and initials read
	// as a face the caller has not sent yet.
	Avatar *ui.Bitmap
	// Subtitle is the line under the name — a number, "video call", a
	// device. Empty draws the name alone.
	Subtitle string
	// AcceptLabel names the button that takes the call; empty takes the
	// library's "Accept".
	AcceptLabel string
	// DeclineLabel names the one that ends it; empty takes the library's
	// "Decline".
	DeclineLabel string
	// Label names the dialog for assistive technology; empty uses the name,
	// which is the whole of what it is.
	Label string
}

// IncomingCallDialogResult carries an IncomingCallDialog and the answer it
// was given.
type IncomingCallDialogResult struct {
	// Element is the dialog.
	Element  *ui.Element
	accepted bool
	declined bool
}

// Accepted reports the accept button being pressed this frame. Read it
// inside the view, as every report in this library is: the view runs up to
// three times per frame and the settled pass has no press.
func (r IncomingCallDialogResult) Accepted() bool { return r.accepted }

// Declined reports the decline button being pressed this frame, on the same
// terms Accepted is read.
func (r IncomingCallDialogResult) Declined() bool { return r.declined }

// IncomingCallDialog is the panel of a call that has arrived: the caller,
// the fact that it is ringing, and the two answers a call can be given.
//
// It is drawn rather than hung off an anchor, because a call that has
// arrived belongs to the window rather than to a control in it: the caller
// draws it while its call state is incoming, and stops drawing it when the
// answer comes, which is the caller's state and not the dialog's. The
// buttons report and do not act, for the reason CallControls spells out —
// taking a call starts audio the application has to set up, and declining
// one is a decision the application makes, not the component's.
//
// name is required and panics when empty: a ringing panel with no name is a
// call nobody can be told about, and a panel a screen reader reads as "the
// dialog" is one the reader cannot answer, because there is nobody to
// answer.
//
//	core.Use(c, core.Settings{})
//	if app.incoming != "" {
//	    res := IncomingCallDialog(c, app.incoming, IncomingCallDialogOptions{
//	        Subtitle: "video call",
//	    })
//	    if res.Accepted() {
//	        app.answer()
//	    }
//	    if res.Declined() {
//	        app.endCall()
//	    }
//	}
func IncomingCallDialog(c *ui.Context, name string, opts IncomingCallDialogOptions) IncomingCallDialogResult {
	if strings.TrimSpace(name) == "" {
		panic("messaging: IncomingCallDialog needs a name; a ringing panel with no name is a " +
			"call nobody can be told about")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	accept := opts.AcceptLabel
	if accept == "" {
		accept = core.Msg(c, "messaging.incoming.accept", core.Def("Accept"))
	}
	decline := opts.DeclineLabel
	if decline == "" {
		decline = core.Msg(c, "messaging.incoming.decline", core.Def("Decline"))
	}
	label := opts.Label
	if label == "" {
		label = name
	}
	ringing := core.Msg(c, "messaging.incoming.ringing", core.Def("Ringing…"))

	var res IncomingCallDialogResult
	res.Element = ui.Box(c).Column().Center().
		Radius(theme.PanelRadius).Padding(u*3, u*4).
		Background(k.Background).
		BorderWidth(theme.BorderWidth).BorderColor(k.Border).
		Shadow(0, u*2.5, u*7.5, 0, ui.RGBA(0, 0, 0, callShadowAlpha(c))).
		Label(label).Role(ui.RoleDialog).Children(func() {
		// The face: the picture when the caller has sent one, the initials
		// when they have not.
		if opts.Avatar != nil {
			display.Image(c, opts.Avatar, display.ImageOptions{
				Width: u * 10, Height: u * 10, Ratio: 1, Cover: true,
				Radius: theme.PillRadius, Name: name,
			})
		} else {
			ui.Box(c).Size(u*10, u*10).Radius(theme.PillRadius).
				Background(k.Surface).Center().Children(func() {
				ui.Text(c, initials(name)).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.TitleSize)).Bold()
			})
		}
		ui.Text(c, name).TextColor(k.Text).
			FontSize(core.FontSize(c, theme.SheetSize)).Bold().MaxLines(1)
		if opts.Subtitle != "" {
			ui.Text(c, opts.Subtitle).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.MetaSize))
		}
		ui.Text(c, ringing).TextColor(k.TextMuted).
			FontSize(core.FontSize(c, theme.RowSize))
		// The two answers, side by side: the refusal first, as the call
		// controls put the leave on the edge, and the answer the one that
		// takes the focus, so Enter accepts without the pointer.
		ui.Row(c).Gap(u*1.5).Justify(ui.Center).Margin(u*2, 0, 0, 0).Children(func() {
			if input.Button(c, decline, input.ButtonOptions{Danger: true, Label: decline}).
				Clicked() {
				res.declined = true
			}
			if input.Button(c, accept, input.ButtonOptions{Primary: true, Label: accept}).
				AutoFocus().Clicked() {
				res.accepted = true
			}
		})
	})
	return res
}

// callShadowAlpha is how dark the dialog's shadow is. A dark window already
// sits low against its own page, so the same shadow reads as a heavier box
// there; a deeper alpha in dark keeps the panel equally lifted in both
// appearances, as the panels in layout do.
func callShadowAlpha(c *ui.Context) float32 {
	if core.IsDark(c) {
		return 0.34
	}
	return 0.18
}
