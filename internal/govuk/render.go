package govuk

import (
	"fmt"
	"sort"
	"strings"
)

// renderers maps a GOV.UK Frontend component name to its Go port of the component's
// template.njk. Every component that ships fixtures has an entry, and the fixture suite in
// render_test.go fails if one is missing.
//
// Each function returns the component's HTML with the leading and trailing whitespace the
// Nunjucks template would have produced; [Render] trims it, exactly as GOV.UK Frontend does
// when it generates fixtures.
var renderers = map[string]func(*Params) string{
	"accordion":           renderAccordion,
	"back-link":           renderBackLink,
	"breadcrumbs":         renderBreadcrumbs,
	"button":              renderButton,
	"character-count":     renderCharacterCount,
	"checkboxes":          renderCheckboxes,
	"cookie-banner":       renderCookieBanner,
	"date-input":          renderDateInput,
	"details":             renderDetails,
	"error-message":       renderErrorMessage,
	"error-summary":       renderErrorSummary,
	"exit-this-page":      renderExitThisPage,
	"feedback":            renderFeedback,
	"fieldset":            renderFieldset,
	"file-upload":         renderFileUpload,
	"footer":              renderFooter,
	"generic-header":      renderGenericHeader,
	"header":              renderHeader,
	"hint":                renderHint,
	"input":               renderInput,
	"inset-text":          renderInsetText,
	"label":               renderLabel,
	"language-navigation": renderLanguageNavigation,
	"notification-banner": renderNotificationBanner,
	"pagination":          renderPagination,
	"panel":               renderPanel,
	"password-input":      renderPasswordInput,
	"phase-banner":        renderPhaseBanner,
	"radios":              renderRadios,
	"select":              renderSelect,
	"service-navigation":  renderServiceNavigation,
	"skip-link":           renderSkipLink,
	"summary-list":        renderSummaryList,
	"table":               renderTable,
	"tabs":                renderTabs,
	"tag":                 renderTag,
	"task-list":           renderTaskList,
	"textarea":            renderTextarea,
	"warning-text":        renderWarningText,
}

// Render returns the HTML for one GOV.UK Frontend component.
//
// component is the kebab-case directory name used by GOV.UK Frontend, such as "back-link" or
// "character-count". params holds the component's options with the same names the Nunjucks
// macros use; see the Design System for each component's options.
//
// The `html` options are trusted and emitted unescaped, matching `| safe` in the upstream
// macros. Only pass HTML you control, and prefer the `text` options, which are escaped.
//
// The returned HTML is trimmed, so it is byte-for-byte equal to the matching entry in the
// release's fixtures.json.
//
// It returns an error only when component is not a GOV.UK Frontend component.
func Render(component string, params *Params) (string, error) {
	render, ok := renderers[component]
	if !ok {
		return "", fmt.Errorf("govuk: %q is not a GOV.UK Frontend component", component)
	}
	return strings.TrimSpace(render(params)), nil
}

// MustRender is [Render] for call sites where the component name is a constant, such as page
// templates. It panics when the component does not exist.
func MustRender(component string, params *Params) string {
	html, err := Render(component, params)
	if err != nil {
		panic(err)
	}
	return html
}

// Components returns the names this package can render, sorted, which is the same set as the
// component directories that ship a fixtures.json.
func Components() []string {
	names := make([]string, 0, len(renderers))
	for name := range renderers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// heading renders a heading level option, falling back to fallback when it is not set.
func heading(level any, fallback string) string {
	if truthy(level) {
		return str(level)
	}
	return fallback
}

// concatIf is the `(" " + option if option)` idiom the macros use to append to a class list.
// Nunjucks compiles an inline `if` with no `else` to an empty string, so a missing option
// contributes nothing rather than the JavaScript "undefined".
func concatIf(prefix string, value any) string {
	if !truthy(value) {
		return ""
	}
	return prefix + str(value)
}
