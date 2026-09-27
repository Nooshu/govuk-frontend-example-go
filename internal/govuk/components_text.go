package govuk

import "strings"

// This file ports the components whose template is a single block of text or a short wrapper:
// back link, hint, inset text, skip link, tag, warning text, error message, details, label,
// panel, phase banner, feedback, and fieldset.

func renderBackLink(p *Params) string {
	text := out(defTruthy(p.Get("text"), "Back"))
	if html := p.Get("html"); truthy(html) {
		text = str(html)
	}
	return `<a href="` + out(defTruthy(p.Get("href"), "#")) + `" class="govuk-back-link` +
		classesIf(p.Get("classes")) + `"` + Attributes(p.Get("attributes")) + `>` + text + `</a>`
}

func renderSkipLink(p *Params) string {
	return `<a href="` + out(defTruthy(p.Get("href"), "#content")) + `" class="govuk-skip-link` +
		classesIf(p.Get("classes")) + `"` + Attributes(p.Get("attributes")) +
		` data-module="govuk-skip-link">` + content(p, "html", "text") + `</a>`
}

func renderHint(p *Params) string {
	return `<div` + attributeIf("id", p.Get("id")) + ` class="govuk-hint` +
		classesIf(p.Get("classes")) + `"` + Attributes(p.Get("attributes")) + ">\n  " +
		contentIndent(p, "html", "text", 2) + "\n</div>"
}

func renderInsetText(p *Params) string {
	return `<div` + attributeIf("id", p.Get("id")) + ` class="govuk-inset-text` +
		classesIf(p.Get("classes")) + `"` + Attributes(p.Get("attributes")) + ">\n  " +
		contentIndent(p, "html", "text", 2) + "\n</div>"
}

func renderTag(p *Params) string {
	return `<strong class="govuk-tag` + classesIf(p.Get("classes")) + `"` +
		Attributes(p.Get("attributes")) + ">\n  " + contentIndent(p, "html", "text", 2) +
		"\n</strong>"
}

func renderWarningText(p *Params) string {
	return `<div class="govuk-warning-text` + classesIf(p.Get("classes")) + `"` +
		Attributes(p.Get("attributes")) + ">\n" +
		"  <span class=\"govuk-warning-text__icon\" aria-hidden=\"true\">!</span>\n" +
		"  <strong class=\"govuk-warning-text__text\">\n" +
		`    <span class="govuk-visually-hidden">` +
		out(defTruthy(p.Get("iconFallbackText"), "Warning")) + "</span>\n" +
		"    " + content(p, "html", "text") + "\n" +
		"  </strong>\n</div>"
}

func renderErrorMessage(p *Params) string {
	visuallyHidden := def(p.Get("visuallyHiddenText"), "Error")
	message := contentIndent(p, "html", "text", 2)

	var out_ strings.Builder
	out_.WriteString(`<p` + attributeIf("id", p.Get("id")) + ` class="govuk-error-message` +
		classesIf(p.Get("classes")) + `"` + Attributes(p.Get("attributes")) + ">\n")
	if truthy(visuallyHidden) {
		out_.WriteString(`  <span class="govuk-visually-hidden">` + out(visuallyHidden) +
			`:</span> ` + message + "\n")
	} else {
		out_.WriteString("  " + message + "\n")
	}
	out_.WriteString("</p>")
	return out_.String()
}

func renderDetails(p *Params) string {
	return `<details` + attributeIf("id", p.Get("id")) + ` class="govuk-details` +
		classesIf(p.Get("classes")) + `"` + Attributes(p.Get("attributes")) +
		flagIf(" open", p.Get("open")) + ">\n" +
		"  <summary class=\"govuk-details__summary\">\n" +
		"    <span class=\"govuk-details__summary-text\">\n" +
		"      " + contentIndent(p, "summaryHtml", "summaryText", 6) + "\n" +
		"    </span>\n  </summary>\n" +
		"  <div class=\"govuk-details__text\">\n" +
		"    " + content(p, "html", "text") + "\n" +
		"  </div>\n</details>"
}

// renderLabel returns the empty string when neither text nor HTML is supplied, which is how the
// form components omit a label without a special case at each call site.
func renderLabel(p *Params) string {
	if !truthy(p.Get("html")) && !truthy(p.Get("text")) {
		return ""
	}
	label := `<label class="govuk-label` + classesIf(p.Get("classes")) + `"` +
		Attributes(p.Get("attributes")) + attributeIf("for", p.Get("for")) + ">\n  " +
		contentIndent(p, "html", "text", 2) + "\n</label>\n"

	if truthy(p.Get("isPageHeading")) {
		return "<h1 class=\"govuk-label-wrapper\">\n  " + indent(trim(label), 2, false) + "\n</h1>\n"
	}
	return trim(label) + "\n"
}

func renderPanel(p *Params) string {
	classes := p.Get("classes")
	interruption := truthy(classes) && contains("govuk-panel--interruption", classes)
	level := heading(p.Get("headingLevel"), "1")

	var out_ strings.Builder
	out_.WriteString(`<div class="govuk-panel`)
	if !interruption {
		out_.WriteString(" govuk-panel--confirmation")
	}
	out_.WriteString(classesIf(classes) + `"` + Attributes(p.Get("attributes")) + ">\n")
	out_.WriteString("  <h" + level + ` class="govuk-panel__title">` + "\n    " +
		content(p, "titleHtml", "titleText") + "\n  </h" + level + ">\n")

	if truthy(p.Get("html")) || truthy(p.Get("text")) {
		out_.WriteString("  <div class=\"govuk-panel__body\">\n    " +
			contentIndent(p, "html", "text", 4) + "\n  </div>\n")
	}

	if actions := p.Get("actions"); interruption && truthy(actions) {
		out_.WriteString(`  <div class="govuk-panel__actions` +
			classesIf(get(actions, "classes")) + `"` +
			Attributes(get(actions, "attributes")) + ">")
		if entries := items(get(actions, "items")); len(entries) > 0 {
			out_.WriteString("<div class=\"govuk-button-group\">\n")
			for _, action := range entries {
				out_.WriteString("      " + indent(trim(panelAction(action)), 6, false) + "\n")
			}
			out_.WriteString("    </div>")
		}
		out_.WriteString("</div>\n")
	}

	out_.WriteString("</div>")
	return out_.String()
}

// panelAction renders one interruption panel action: a button, or an inverse link when the
// action has an href and is not explicitly a button.
func panelAction(action any) string {
	href := get(action, "href")
	if !truthy(href) || str(get(action, "type")) == "button" {
		return renderButton(NewParams(
			"text", get(action, "text"),
			"type", defTruthy(get(action, "type"), "button"),
			"classes", "govuk-button--inverse"+concatIf(" ", get(action, "classes")),
			"href", href,
			"attributes", get(action, "attributes"),
		))
	}
	return `<a class="govuk-link govuk-link--inverse` + classesIf(get(action, "classes")) +
		`" href="` + out(href) + `"` + Attributes(get(action, "attributes")) + `>` +
		out(get(action, "text")) + `</a>`
}

func renderPhaseBanner(p *Params) string {
	tag := p.Get("tag")
	tagHTML := renderTag(NewParams(
		"text", get(tag, "text"),
		"html", get(tag, "html"),
		"classes", "govuk-phase-banner__content__tag"+concatIf(" ", get(tag, "classes")),
	))
	return `<div class="govuk-phase-banner govuk-width-container` + classesIf(p.Get("classes")) +
		`"` + Attributes(p.Get("attributes")) + ">\n" +
		"  <p class=\"govuk-phase-banner__content\">\n" +
		"    " + indent(trim(tagHTML), 4, false) + "\n" +
		"    <span class=\"govuk-phase-banner__text\">\n" +
		"      " + contentIndent(p, "html", "text", 6) + "\n" +
		"    </span>\n  </p>\n</div>"
}

func renderFeedback(p *Params) string {
	level := heading(p.Get("headingLevel"), "2")

	var out_ strings.Builder
	out_.WriteString(`<div class="govuk-feedback govuk-width-container` +
		classesIf(p.Get("classes")) + `"` + Attributes(p.Get("attributes")) + ">\n")
	out_.WriteString("  <div class=\"govuk-grid-row\">\n")
	out_.WriteString("    <div class=\"govuk-grid-column-two-thirds\">\n")
	out_.WriteString("      <h" + level + ` class="govuk-feedback__title">` + "\n        " +
		content(p, "titleHtml", "titleText") + "\n      </h" + level + ">\n")

	html, text := p.Get("html"), p.Get("text")
	if truthy(html) || truthy(text) {
		out_.WriteString("        <div class=\"govuk-feedback__body\">\n")
		switch {
		case truthy(html):
			out_.WriteString("            " + indent(trim(str(html)), 4, false) + "\n")
		case truthy(text):
			out_.WriteString("            <p class=\"govuk-body\">\n")
			out_.WriteString("              " + escape(indent(trim(str(text)), 6, false)) + "\n")
			out_.WriteString("            </p>\n")
		}
		out_.WriteString("        </div>\n")
	}

	out_.WriteString("    </div>\n  </div>\n</div>")
	return out_.String()
}

func renderFieldset(p *Params) string {
	var out_ strings.Builder
	out_.WriteString(`<fieldset class="govuk-fieldset` + classesIf(p.Get("classes")) + `"` +
		attributeIf("role", p.Get("role")) +
		attributeIf("aria-describedby", p.Get("describedBy")) +
		Attributes(p.Get("attributes")) + ">\n")

	legend := p.Get("legend")
	if truthy(get(legend, "html")) || truthy(get(legend, "text")) {
		out_.WriteString(`  <legend class="govuk-fieldset__legend` +
			classesIf(get(legend, "classes")) + "\">\n")
		if truthy(get(legend, "isPageHeading")) {
			out_.WriteString("    <h1 class=\"govuk-fieldset__heading\">\n")
			out_.WriteString("      " + contentIndent(legend, "html", "text", 6) + "\n")
			out_.WriteString("    </h1>\n")
		} else {
			out_.WriteString("    " + contentIndent(legend, "html", "text", 4) + "\n")
		}
		out_.WriteString("  </legend>\n")
	}

	if html := p.Get("html"); truthy(html) {
		out_.WriteString("  " + str(html) + "\n")
	}

	out_.WriteString("</fieldset>")
	return out_.String()
}
