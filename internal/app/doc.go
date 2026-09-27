// Package app is the HTTP surface of the example service.
//
// One [App] owns the route table, session cookie lifecycle, CSRF checks, and the
// write path that applies the shared baseline (OWASP headers, cache kinds, ETags,
// Brotli-first compression). Handlers compose pages through [pages.Renderer] and
// never emit ad-hoc govuk-* markup.
//
// Tests typically build an App with a stub [render.Renderer]; fixture parity for
// components lives in package govuk, not here.
package app
