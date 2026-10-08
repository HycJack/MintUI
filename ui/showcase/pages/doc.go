// Package pages holds one file per ui/ package, each drawing that package's
// components so a person can look at them.
//
// A page file must:
//
//   - be named <package>.go, matching the ui/ package it covers;
//   - call showcase.Register with a Page whose Package field is <package>;
//   - draw as much of the package as it can, laid out under Section headings;
//   - list in Want every string it is sure it draws.
//
// The gate is showcase_test.go: a page that promised text it did not draw
// fails, and a ui/ package with no page is an unreviewed package.
package pages
