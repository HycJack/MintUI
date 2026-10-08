// Package devtools holds the components an application's own machinery is
// looked at with: the JSON a response came back as, the log lines streaming
// past, the spans of a trace laid out against each other, and the numbers a
// dashboard is made of.
//
// # Reading is most of it
//
// The interesting parts of this package are functions. ParseJSON turns text
// into a tree, Redact walks that tree hiding the values somebody must not be
// shown, UptimeSlots turns a year of probes into ninety squares, and
// SpanBar turns a set of spans into percentages of a trace. Each is a
// function with a name rather than a screen with a flag, because the tests
// for them are exact and the alternatives are not: a JSON tree that is
// correct but cannot be asserted on is a JSON tree that breaks silently.
//
// # What is drawn is not what is parsed
//
// JsonViewer draws a tree that ParseJSON produced; it does not parse anything
// itself. TraceWaterfall draws bars from SpanBar's numbers; it does no
// arithmetic. The split is the same one this library makes everywhere — a
// function says what a thing means and a component says where it goes — and
// it is why every pure function here can be tested to the decimal place
// without a window.
//
// # Secrets
//
// ResponseViewer parses the body and draws the tree with the sensitive values
// already hidden. Redaction is on by default and off by an explicit request,
// because the failure mode of getting it backwards is a token pasted into a
// screenshot of a bug report.
//
// # Conventions
//
// As everywhere in this library: core.Use is called once by the harness, the
// parameters are in XxxOptions, the interactions come back on an XxxResult as
// a single method, nothing here keeps a value the caller did not hand it, and
// anything this package cannot draw is a panic rather than a guess.
package devtools
