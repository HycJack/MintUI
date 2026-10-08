// Package messaging is the conversation half of the interface: the channels
// and people on one side, the messages between them on the other, and the
// mail and call surfaces that share the same vocabulary.
//
// # It is a layer over ui/chat, not a second implementation
//
// Every message in this package is drawn by [chat.MessageBubble]. It is not
// reimplemented, wrapped-from-scratch or approximated — a wrapper is built
// around the call, the bubble decides the ink, the corners, the tail and the
// place in the column, and what this package adds is the decoration a *message
// channel* needs and a transcript does not:
//
//   - the divider above the first unread message
//   - the read receipt under the last of a run of outgoing messages
//   - the quoted block and the channel badge
//
// Reimplementing the bubble here would be the single most expensive mistake
// available to this package: two answers to "how does a bubble look", written
// a year apart, and a window in which the right-hand side of a chat reads
// differently from the right-hand side of the same conversation opened as a
// channel.
//
// # Nothing here reaches the network
//
// Messages, members, mail and call state all arrive as Go values. A message
// that is "arriving" is a message whose text is still growing, which is a
// string the caller owns. There is no socket, no store and no account, so
// every component here can be drawn in a headless test.
//
// # State belongs to the caller
//
// The selected message, the open thread, the reading position, the recipient
// list and the call's mute flag are all pointers the caller holds. Two views
// of one conversation cannot disagree about which turn is chosen, and a
// component that kept its own copy would be a second source of truth about
// something the caller has to persist.
package messaging
