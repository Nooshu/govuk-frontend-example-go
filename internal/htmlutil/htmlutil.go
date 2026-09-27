package htmlutil

// Shared Escape and Join helpers for page text.

import "strings"

var escaper = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	`"`, "&quot;",
	"'", "&#39;",
)

// Escape makes plain text safe to place in HTML.
//
// It matches the escaping GOV.UK Frontend's Nunjucks macros apply, including &#39; for
// apostrophes, so text escaped here is identical to text the macros escape.
func Escape(value string) string {
	return escaper.Replace(value)
}

// PageTitle builds the document title.
//
// The service name is not repeated when it is already the heading, and errors are announced
// first so a screen reader reads "Error:" before anything else.
func PageTitle(heading, serviceName string, hasErrors bool) string {
	prefix := ""
	if hasErrors {
		prefix = "Error: "
	}
	if heading == serviceName {
		return prefix + serviceName + " – GOV.UK"
	}
	return prefix + heading + " – " + serviceName + " – GOV.UK"
}

// SafeLocalPath returns value when it is a same-origin path, and "/" otherwise.
//
// It rejects protocol-relative and absolute URLs so a return path taken from the request cannot
// send the applicant to another site, and rejects newlines so it cannot be used to inject a
// response header.
func SafeLocalPath(value string) string {
	if !strings.HasPrefix(value, "/") ||
		strings.HasPrefix(value, "//") ||
		strings.Contains(value, "://") ||
		strings.Contains(value, `\`) ||
		strings.ContainsAny(value, "\r\n") {
		return "/"
	}
	return value
}
