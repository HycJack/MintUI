// Package account holds the components an account screen is made of: signing
// in, the keys and sessions behind an account, the preferences that belong to
// the person rather than to a workspace, and the furniture around them — the
// switcher in the corner, the onboarding wizard, the feedback box.
//
// None of it draws a control of its own. A login form here is a layout and a
// pair of rules; the text field is input.TextInput, the code is
// input.PinInput, the switch is input.Switch. That is deliberate twice over.
// A second text field is a second thing to keep right about focus, about
// paste, about what the caret does at the end of the value, and about what
// happens on a keyboard it was never tried on. And a form that reassembles
// input.TextInput by hand stops looking like the rest of the app the moment
// the rest of the app's field changes.
//
// What the package adds is the part the controls cannot know: which of them a
// login screen wants, in what order, and what counts as wrong.
//
// # The rules a form needs
//
// Validation is here because a form is the only place in the interface where
// the answer to "is this allowed" is not somebody else's backend. The rules
// are functions — ValidateLogin and its neighbours — so that a view can test
// the same rules a screen applies, and so that a rule is written once.
//
// # Secrets
//
// A key is never drawn. MaskKey is the only thing in this package that turns
// a secret into something a person may see, and everything that shows a key
// goes through it. There is no second path to the plaintext: a key that must
// be copied goes through input.CopyButton, which reads the value it is given
// rather than printing it.
//
// # Conventions
//
// As everywhere in this library: core.Use is called once by the harness, the
// parameters are in XxxOptions, the interactions come back on an XxxResult as
// a single method, nothing here keeps a value the caller did not hand it, and
// anything this package cannot draw is a panic rather than a guess.
package account
