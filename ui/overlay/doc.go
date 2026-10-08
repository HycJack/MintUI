// Package overlay holds what a window puts on top of itself: dialogs, sheets,
// popovers, tips, menus and the one button that carries a menu beside its own
// action.
//
// # One look, one place
//
// [Panel] is the only thing in the library that decides what a floating layer
// looks like — the radius, the hairline, the shadow, the padding, the title
// bar. Every other component here hands its host element to Panel and adds
// nothing of its own, so a dialog, a popover and a tip read as three sizes of
// the same object rather than three objects.
//
// # Layers and modality
//
// Every component here that draws anything puts it through ui.Overlay, so its
// layer lands above whatever the window built. On top of that there are two
// more things a layer may do, and they are separate on purpose:
//
// A layer is modal when the window behind it is inert — presses outside go
// to the scrim, Tab goes round what is inside, and what is behind is out of
// reach to a screen reader. Modal is the default everywhere in this package.
// Dialog, AlertDialog, Drawer, Popover, Popconfirm and HoverCard each take a
// NonModal option, and that option is the only way out: it is called
// NonModal because turning modality off is a decision the caller has to make
// rather than a default worth guessing at.
//
// NonModal is right for a control attached to something on the page — a
// popover under a filter row, a confirm under a toolbar button — because the
// page behind it is what the user is working in, and a scrim across it says
// the work stopped. A menu is NonModal for the same reason. A dialog or a
// sheet is modal because the thing behind it is exactly what the dialog is
// about to change.
//
// The two menus are a third thing again: [ContextMenu] and [DropdownMenu]
// hand their items to the desktop, which draws them in its own look. Panel
// has nothing to say about them, and asking the desktop to wear the library's
// hairline would be asking it to be something it is not.
//
// # Closing
//
// Nothing here holds its own open state. The *bool is the caller's, and every
// way of closing a layer writes to it: a press on the scrim, Escape, a press
// outside an anchored panel, a button. The component that renders a button
// reports the press back through its result instead of closing anything, so
// what a dialog does on "Delete" is the caller's code and not a hidden one.
package overlay
