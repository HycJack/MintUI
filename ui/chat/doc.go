// Package chat holds the components a conversation is drawn with: the
// bubbles, the composer, the transcript, and everything a model's answer is
// made of — markdown, code, tables, images, citations.
//
// Two conventions run through all of it. The first is the library's: a view
// calls [core.Use] once per frame, every component takes XxxOptions, an
// interactive component returns an XxxResult exposing one method for the one
// thing a person can do to it, no component holds state — what is selected,
// what is expanded and what is typed are all the caller's pointers — and a
// component that cannot be given what it needs panics rather than guessing.
//
// The second is chat's own: a bubble is drawn by one function. Padding, radius,
// the gap between turns, the hairline and the tail all come from bubbleSkin
// and bubble, so the four roles — what I said, what the model said, what the
// system said, what a tool said — differ in where they sit and which corner
// carries the tail, and in nothing else. A transcript where each role picked
// its own margins is a transcript where the eye has to re-learn the spacing at
// every turn.
//
// The two pure functions live here too and are the package's real
// foundations: [Parse] turns markdown into blocks, and [Highlight] turns
// source into tokens. Neither takes a Context, because neither needs one — they
// are answerable without a palette, and a parser that cannot be tested on its
// own is a parser nobody tests.
package chat
