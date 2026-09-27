package govuk

// Ports for button-related Frontend components.

import "strings"

// startIcon is the arrow GOV.UK Frontend appends to a start button. The leading newline is
// part of the upstream macro's output and is load-bearing for fixture parity.
const startIcon = "\n" +
	`  <svg class="govuk-button__start-icon" xmlns="http://www.w3.org/2000/svg" width="17.5" height="19" viewBox="0 0 33 40" aria-hidden="true" focusable="false">` + "\n" +
	`    <path fill="currentColor" d="M0 0h13l20 20-20 20H0l20-20z"/>` + "\n" +
	"  </svg>"

func renderButton(p *Params) string {
	classNames := "govuk-button"
	if classes := p.Get("classes"); truthy(classes) {
		classNames += " " + str(classes)
	}
	startButton := truthy(p.Get("isStartButton"))
	if startButton {
		classNames += " govuk-button--start"
	}

	// commonAttributes is shared by the link and the button element, so both carry the same
	// class list, module hook, and caller-supplied attributes in the same order.
	commonAttributes := ` class="` + escape(classNames) + `" data-module="govuk-button"` +
		Attributes(p.Get("attributes")) + attributeIf("id", p.Get("id"))

	html := p.Get("html")
	text := out(p.Get("text"))
	switch {
	case truthy(html) && startButton:
		// Start buttons lay their content out with flexbox, which drops the whitespace between
		// child elements, so HTML content is wrapped to keep it as one flex item.
		text = "<span>" + trim(str(html)) + "</span>"
	case truthy(html):
		text = trim(str(html))
	}

	var out_ strings.Builder
	if href := p.Get("href"); truthy(href) {
		out_.WriteString(`<a href="` + out(href) + `" role="button" draggable="false"` +
			commonAttributes + ">\n  " + indent(text, 2, false))
	} else {
		out_.WriteString(`<button type="` + out(defTruthy(p.Get("type"), "submit")) + `"` +
			attributeIf("value", p.Get("value")) +
			attributeIf("name", p.Get("name")) +
			flagIf(` disabled aria-disabled="true"`, p.Get("disabled")))
		if preventDoubleClick := p.Get("preventDoubleClick"); !isUndefined(preventDoubleClick) {
			out_.WriteString(` data-prevent-double-click="` + out(preventDoubleClick) + `"`)
		}
		out_.WriteString(commonAttributes + ">\n  " + indent(text, 2, false))
	}
	if startButton {
		out_.WriteString(startIcon)
	}
	if truthy(p.Get("href")) {
		out_.WriteString("\n</a>")
	} else {
		out_.WriteString("\n</button>")
	}
	return out_.String()
}

// exitThisPageDefaultHTML is the button label used when the caller supplies neither text nor
// HTML. The trailing newline matches the captured `{% set %}` block upstream.
const exitThisPageDefaultHTML = `  <span class="govuk-visually-hidden">Emergency</span> Exit this page` + "\n"

func renderExitThisPage(p *Params) string {
	html := p.Get("html")
	if !truthy(html) && !truthy(p.Get("text")) {
		html = Safe(exitThisPageDefaultHTML)
	}

	button := renderButton(NewParams(
		"html", html,
		"text", p.Get("text"),
		"classes", "govuk-button--warning govuk-exit-this-page__button govuk-js-exit-this-page-button",
		"href", defTruthy(p.Get("redirectUrl"), "https://www.bbc.co.uk/weather"),
		"attributes", NewParams("rel", "nofollow noreferrer"),
	))

	return `<div` + attributeIf("id", p.Get("id")) + ` class="govuk-exit-this-page` +
		classesIf(p.Get("classes")) + `" data-module="govuk-exit-this-page"` +
		Attributes(p.Get("attributes")) +
		attributeIf("data-i18n.activated", p.Get("activatedText")) +
		attributeIf("data-i18n.timed-out", p.Get("timedOutText")) +
		attributeIf("data-i18n.press-two-more-times", p.Get("pressTwoMoreTimesText")) +
		attributeIf("data-i18n.press-one-more-time", p.Get("pressOneMoreTimeText")) +
		">\n  " + indent(trim(button), 2, false) + "\n</div>"
}
