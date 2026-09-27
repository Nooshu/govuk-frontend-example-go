// Package pages composes the GOV.UK page template in Go.
//
// A [Document] describes title, language, before-content, and main content. The
// [Renderer] executes embedded html/template files and calls a [render.Renderer]
// for every govuk-* block so pages never paste component markup.
//
// See docs/page-shell.md and docs/layout-chrome.md.
package pages
