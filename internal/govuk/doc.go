// Package govuk is a native Go port of GOV.UK Frontend component macros.
//
// Each renderer tracks the corresponding template.njk from the pinned
// govuk-frontend package. The primary contract is byte-for-byte equality between
// [Render] output and every official fixtures.json html value for that release.
//
// Options are carried in ordered [Params] values so attribute iteration matches
// Nunjucks object key order. Text escaping follows Nunjucks rules (including
// &#39; and &#92;), not only html.EscapeString.
//
// Do not shell out to Node to render HTML, and do not edit fixture html to make
// tests pass — fix the renderer. See docs/testing-components.md and
// docs/go-conventions.md.
package govuk
