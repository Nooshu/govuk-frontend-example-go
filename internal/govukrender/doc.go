// Package govukrender connects the page layer to the Go port of GOV.UK Frontend macros.
//
// It implements [render.Renderer] by decoding page-supplied option maps into
// ordered [govuk.Params] and calling [govuk.Render]. Keep this adapter thin so
// fixture parity stays owned by package govuk.
package govukrender
