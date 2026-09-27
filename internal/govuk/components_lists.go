package govuk

import (
	"strconv"
	"strings"
)

// This file ports the components that render a list or a table of items: accordion, error
// summary, notification banner, summary list, table, tabs, and task list.

func renderAccordion(p *Params) string {
	var out_ strings.Builder
	out_.WriteString(`<div class="govuk-accordion` + classesIf(p.Get("classes")) +
		`" data-module="govuk-accordion" id="` + out(p.Get("id")) + `"` +
		i18nAttributes("hide-all-sections", p.Get("hideAllSectionsText"), Undefined) +
		i18nAttributes("hide-section", p.Get("hideSectionText"), Undefined) +
		i18nAttributes("hide-section-aria-label", p.Get("hideSectionAriaLabelText"), Undefined) +
		i18nAttributes("show-all-sections", p.Get("showAllSectionsText"), Undefined) +
		i18nAttributes("show-section", p.Get("showSectionText"), Undefined) +
		i18nAttributes("show-section-aria-label", p.Get("showSectionAriaLabelText"), Undefined))
	if remember := p.Get("rememberExpanded"); !isUndefined(remember) {
		out_.WriteString(` data-remember-expanded="` + escape(str(remember)) + `"`)
	}
	out_.WriteString(Attributes(p.Get("attributes")) + ">\n")

	for index, item := range items(p.Get("items")) {
		if !truthy(item) {
			continue
		}
		out_.WriteString(accordionItem(p, item, index+1))
	}

	out_.WriteString("</div>")
	return out_.String()
}

func accordionItem(p *Params, item any, index int) string {
	level := heading(p.Get("headingLevel"), "2")
	id := out(p.Get("id"))
	position := strconv.Itoa(index)
	itemHeading := get(item, "heading")
	summary := get(item, "summary")
	itemContent := get(item, "content")

	var out_ strings.Builder
	out_.WriteString(`  <div class="govuk-accordion__section` +
		flagIf(" govuk-accordion__section--expanded", get(item, "expanded")) + "\">\n")
	out_.WriteString("    <div class=\"govuk-accordion__section-header\">\n")
	out_.WriteString("      <h" + level + " class=\"govuk-accordion__section-heading\">\n")
	out_.WriteString(`        <span class="govuk-accordion__section-button" id="` + id +
		"-heading-" + position + "\">\n")
	out_.WriteString("          " + contentIndent(itemHeading, "html", "text", 8) + "\n")
	out_.WriteString("        </span>\n      </h" + level + ">\n")
	if truthy(get(summary, "html")) || truthy(get(summary, "text")) {
		out_.WriteString(`      <div class="govuk-accordion__section-summary govuk-body" id="` +
			id + "-summary-" + position + "\">\n")
		out_.WriteString("        " + contentIndent(summary, "html", "text", 8) + "\n")
		out_.WriteString("      </div>\n")
	}
	out_.WriteString("    </div>\n")
	out_.WriteString(`    <div id="` + id + "-content-" + position +
		"\" class=\"govuk-accordion__section-content\">\n")
	switch html, text := get(itemContent, "html"), get(itemContent, "text"); {
	case truthy(html):
		out_.WriteString("      " + indent(trim(str(html)), 6, false) + "\n")
	case truthy(text):
		out_.WriteString("      <p class=\"govuk-body\">\n")
		out_.WriteString("        " + escape(indent(trim(str(text)), 8, false)) + "\n")
		out_.WriteString("      </p>\n")
	}
	out_.WriteString("    </div>\n  </div>\n")
	return out_.String()
}

func renderErrorSummary(p *Params) string {
	var out_ strings.Builder
	out_.WriteString(`<div class="govuk-error-summary` + classesIf(p.Get("classes")) + `"`)
	if autoFocus := p.Get("disableAutoFocus"); !isUndefined(autoFocus) {
		out_.WriteString(` data-disable-auto-focus="` + out(autoFocus) + `"`)
	}
	out_.WriteString(Attributes(p.Get("attributes")) + ` data-module="govuk-error-summary">`)

	// role="alert" sits on a separate child container so that focusing the summary does not
	// race the announcement and swallow part of it.
	out_.WriteString("\n  <div role=\"alert\">\n")
	out_.WriteString("    <h2 class=\"govuk-error-summary__title\">\n")
	out_.WriteString("      " + contentIndent(p, "titleHtml", "titleText", 6) + "\n")
	out_.WriteString("    </h2>\n")
	out_.WriteString("    <div class=\"govuk-error-summary__body\">\n")

	if truthy(p.Get("descriptionHtml")) || truthy(p.Get("descriptionText")) {
		out_.WriteString("      <p>\n        " +
			contentIndent(p, "descriptionHtml", "descriptionText", 8) + "\n      </p>\n")
	}

	if errorList := items(p.Get("errorList")); len(errorList) > 0 {
		out_.WriteString("        <ul class=\"govuk-list govuk-error-summary__list\">\n")
		for _, item := range errorList {
			out_.WriteString("          <li>\n")
			if href := get(item, "href"); truthy(href) {
				out_.WriteString(`            <a href="` + out(href) + `"` +
					Attributes(get(item, "attributes")) + ">" +
					contentIndent(item, "html", "text", 12) + "</a>\n")
			} else {
				out_.WriteString("            " + contentIndent(item, "html", "text", 10) + "\n")
			}
			out_.WriteString("          </li>\n")
		}
		out_.WriteString("        </ul>\n")
	}

	out_.WriteString("    </div>\n  </div>\n</div>")
	return out_.String()
}

func renderNotificationBanner(p *Params) string {
	success := str(p.Get("type")) == "success"
	typeClass := ""
	if success {
		typeClass = " govuk-notification-banner--" + escape(str(p.Get("type")))
	}

	// A success banner is announced immediately with role="alert"; anything else becomes a
	// landmark region so it can be navigated to instead of interrupting.
	role := "region"
	switch {
	case truthy(p.Get("role")):
		role = str(p.Get("role"))
	case success:
		role = "alert"
	}

	var title string
	switch {
	case truthy(p.Get("titleHtml")):
		title = str(p.Get("titleHtml"))
	case truthy(p.Get("titleText")):
		title = out(p.Get("titleText"))
	case success:
		title = "Success"
	default:
		title = "Important"
	}

	titleID := out(defTruthy(p.Get("titleId"), "govuk-notification-banner-title"))
	level := out(defTruthy(p.Get("titleHeadingLevel"), "2"))

	var out_ strings.Builder
	out_.WriteString(`<div class="govuk-notification-banner` + typeClass +
		classesIf(p.Get("classes")) + `" role="` + escape(role) + `" aria-labelledby="` +
		titleID + `" data-module="govuk-notification-banner"`)
	if autoFocus := p.Get("disableAutoFocus"); !isUndefined(autoFocus) {
		out_.WriteString(` data-disable-auto-focus="` + out(autoFocus) + `"`)
	}
	out_.WriteString(Attributes(p.Get("attributes")) + ">\n")
	out_.WriteString("  <div class=\"govuk-notification-banner__header\">\n")
	out_.WriteString("    <h" + level + ` class="govuk-notification-banner__title" id="` +
		titleID + "\">\n")
	out_.WriteString("      " + title + "\n    </h" + level + ">\n  </div>\n")
	out_.WriteString("  <div class=\"govuk-notification-banner__content\">\n")
	switch html, text := p.Get("html"), p.Get("text"); {
	case truthy(html):
		out_.WriteString("    " + indent(trim(str(html)), 4, false) + "\n")
	case truthy(text):
		// Single-line content gets the heading style by default.
		out_.WriteString("    <p class=\"govuk-notification-banner__heading\">\n")
		out_.WriteString("      " + escape(indent(trim(str(text)), 6, false)) + "\n")
		out_.WriteString("    </p>\n")
	}
	out_.WriteString("  </div>\n</div>")
	return out_.String()
}

func renderSummaryList(p *Params) string {
	card := p.Get("card")
	cardTitle := get(card, "title")

	// A third column is only added when at least one row has actions, so rows without them
	// line up with the rows that have.
	anyRowHasActions := false
	for _, row := range items(p.Get("rows")) {
		if length(get(row, "actions", "items")) > 0 {
			anyRowHasActions = true
		}
	}

	var list strings.Builder
	list.WriteString(`<dl class="govuk-summary-list` + classesIf(p.Get("classes")) + `"` +
		Attributes(p.Get("attributes")) + ">\n")
	for _, row := range items(p.Get("rows")) {
		if !truthy(row) {
			continue
		}
		key, value, actions := get(row, "key"), get(row, "value"), get(row, "actions")
		list.WriteString(`  <div class="govuk-summary-list__row` +
			flagIf(" govuk-summary-list__row--no-actions",
				anyRowHasActions && !truthy(get(actions, "items"))) +
			classesIf(get(row, "classes")) + "\">\n")
		list.WriteString(`    <dt class="govuk-summary-list__key` +
			classesIf(get(key, "classes")) + "\">\n      " +
			contentIndent(key, "html", "text", 6) + "\n    </dt>\n")
		list.WriteString(`    <dd class="govuk-summary-list__value` +
			classesIf(get(value, "classes")) + "\">\n      " +
			contentIndent(value, "html", "text", 6) + "\n    </dd>\n")

		if entries := items(get(actions, "items")); len(entries) > 0 {
			list.WriteString(`    <dd class="govuk-summary-list__actions` +
				classesIf(get(actions, "classes")) + "\">\n")
			if len(entries) == 1 {
				list.WriteString(indent(trim(summaryActionLink(entries[0], cardTitle)), 6, true) + "\n")
			} else {
				list.WriteString("      <ul class=\"govuk-summary-list__actions-list\">\n")
				for _, action := range entries {
					list.WriteString("        <li class=\"govuk-summary-list__actions-list-item\">\n")
					list.WriteString("          " +
						indent(trim(summaryActionLink(action, cardTitle)), 8, false) + "\n")
					list.WriteString("        </li>\n")
				}
				list.WriteString("      </ul>\n")
			}
			list.WriteString("    </dd>\n")
		}
		list.WriteString("  </div>\n")
	}
	list.WriteString("</dl>")

	if truthy(card) {
		return summaryCard(card, indent(trim(list.String()), 4, false))
	}
	return trim(list.String())
}

// summaryActionLink renders one action link. cardTitle, when present, is appended visually
// hidden so "Change" is unambiguous when several cards are on the page.
func summaryActionLink(action, cardTitle any) string {
	var out_ strings.Builder
	out_.WriteString(`  <a class="govuk-link` + classesIf(get(action, "classes")) +
		`" href="` + out(get(action, "href")) + `"` +
		Attributes(get(action, "attributes")) + ">")
	if html := get(action, "html"); truthy(html) {
		out_.WriteString(indent(str(html), 4, false))
	} else {
		out_.WriteString(out(get(action, "text")))
	}
	if visuallyHidden := get(action, "visuallyHiddenText"); truthy(visuallyHidden) || truthy(cardTitle) {
		out_.WriteString(`<span class="govuk-visually-hidden">`)
		if truthy(visuallyHidden) {
			out_.WriteString(" " + out(visuallyHidden))
		}
		if truthy(cardTitle) {
			title := out(get(cardTitle, "text"))
			if html := get(cardTitle, "html"); truthy(html) {
				title = indent(str(html), 6, false)
			}
			out_.WriteString(" (" + title + ")")
		}
		out_.WriteString("</span>")
	}
	out_.WriteString("</a>\n")
	return out_.String()
}

func summaryCard(card any, body string) string {
	title := get(card, "title")
	level := heading(get(title, "headingLevel"), "2")
	actions := get(card, "actions")

	var out_ strings.Builder
	out_.WriteString(`<div class="govuk-summary-card` + classesIf(get(card, "classes")) + `"` +
		Attributes(get(card, "attributes")) + ">\n")
	out_.WriteString("  <div class=\"govuk-summary-card__title-wrapper\">\n")
	if truthy(title) {
		out_.WriteString("    <h" + level + ` class="govuk-summary-card__title` +
			classesIf(get(title, "classes")) + "\">\n      " +
			contentIndent(title, "html", "text", 6) + "\n    </h" + level + ">\n")
	}
	if entries := items(get(actions, "items")); len(entries) > 0 {
		if len(entries) == 1 {
			out_.WriteString(`    <div class="govuk-summary-card__actions` +
				classesIf(get(actions, "classes")) + "\">\n")
			out_.WriteString("      " +
				indent(trim(summaryActionLink(entries[0], title)), 4, false) + "\n")
			out_.WriteString("    </div>\n")
		} else {
			out_.WriteString(`    <ul class="govuk-summary-card__actions` +
				classesIf(get(actions, "classes")) + "\">\n")
			for _, action := range entries {
				out_.WriteString("      <li class=\"govuk-summary-card__action\">\n")
				out_.WriteString("        " +
					indent(trim(summaryActionLink(action, title)), 8, false) + "\n")
				out_.WriteString("      </li>\n")
			}
			out_.WriteString("    </ul>\n")
		}
	}
	out_.WriteString("  </div>\n\n")
	out_.WriteString("  <div class=\"govuk-summary-card__content\">\n    " + body +
		"\n  </div>\n</div>\n")
	return out_.String()
}

func renderTable(p *Params) string {
	var out_ strings.Builder
	out_.WriteString(`<table class="govuk-table` + classesIf(p.Get("classes")) + `"` +
		Attributes(p.Get("attributes")) + ">\n")

	if caption := p.Get("caption"); truthy(caption) {
		out_.WriteString(`  <caption class="govuk-table__caption` +
			classesIf(p.Get("captionClasses")) + "\">" + out(caption) + "</caption>\n")
	}

	if head := items(p.Get("head")); truthy(p.Get("head")) {
		out_.WriteString("  <thead class=\"govuk-table__head\">\n")
		out_.WriteString("    <tr class=\"govuk-table__row\">\n")
		for _, item := range head {
			out_.WriteString(`      <th scope="col" class="govuk-table__header` +
				formatClass("govuk-table__header--", get(item, "format")) +
				classesIf(get(item, "classes")) + `"` +
				attributeIf("colspan", get(item, "colspan")) +
				attributeIf("rowspan", get(item, "rowspan")) +
				Attributes(get(item, "attributes")) + ">" +
				content(item, "html", "text") + "</th>\n")
		}
		out_.WriteString("    </tr>\n  </thead>\n")
	}

	out_.WriteString("  <tbody class=\"govuk-table__body\">\n")
	for _, row := range items(p.Get("rows")) {
		if !truthy(row) {
			continue
		}
		out_.WriteString("    <tr class=\"govuk-table__row\">\n")
		for index, cell := range items(row) {
			common := attributeIf("colspan", get(cell, "colspan")) +
				attributeIf("rowspan", get(cell, "rowspan")) +
				Attributes(get(cell, "attributes"))
			if index == 0 && truthy(p.Get("firstCellIsHeader")) {
				out_.WriteString(`      <th scope="row" class="govuk-table__header` +
					classesIf(get(cell, "classes")) + `"` + common + ">" +
					content(cell, "html", "text") + "</th>\n")
			} else {
				out_.WriteString(`      <td class="govuk-table__cell` +
					formatClass("govuk-table__cell--", get(cell, "format")) +
					classesIf(get(cell, "classes")) + `"` + common + ">" +
					content(cell, "html", "text") + "</td>\n")
			}
		}
		out_.WriteString("    </tr>\n")
	}
	out_.WriteString("  </tbody>\n</table>")
	return out_.String()
}

// formatClass renders the ` govuk-table__cell--numeric` style modifier from a format option.
func formatClass(prefix string, format any) string {
	if !truthy(format) {
		return ""
	}
	return " " + prefix + out(format)
}

func renderTabs(p *Params) string {
	// Without an id prefix the tab panels fall back to bare "-1", "-2" ids, matching upstream.
	idPrefix := ""
	if prefix := p.Get("idPrefix"); truthy(prefix) {
		idPrefix = str(prefix)
	}

	var out_ strings.Builder
	out_.WriteString(`<div` + attributeIf("id", p.Get("id")) + ` class="govuk-tabs` +
		classesIf(p.Get("classes")) + `"` + Attributes(p.Get("attributes")) +
		" data-module=\"govuk-tabs\">\n")
	out_.WriteString("  <h2 class=\"govuk-tabs__title\">\n    " +
		out(def(p.Get("title"), "Contents")) + "\n  </h2>\n")

	entries := items(p.Get("items"))
	if len(entries) > 0 {
		out_.WriteString("  <ul class=\"govuk-tabs__list\">\n")
		for index, item := range entries {
			if !truthy(item) {
				continue
			}
			out_.WriteString(indent(trim(tabListItem(item, index+1, idPrefix)), 4, true) + "\n")
		}
		out_.WriteString("  </ul>\n")
		for index, item := range entries {
			if !truthy(item) {
				continue
			}
			out_.WriteString(indent(trim(tabPanel(item, index+1, idPrefix)), 2, true) + "\n")
		}
	}

	out_.WriteString("</div>")
	return out_.String()
}

func tabPanelID(item any, index int, idPrefix string) string {
	if id := get(item, "id"); truthy(id) {
		return str(id)
	}
	return idPrefix + "-" + strconv.Itoa(index)
}

func tabListItem(item any, index int, idPrefix string) string {
	return `<li class="govuk-tabs__list-item` +
		flagIf(" govuk-tabs__list-item--selected", index == 1) + "\">\n" +
		`  <a class="govuk-tabs__tab" href="#` + escape(tabPanelID(item, index, idPrefix)) + `"` +
		Attributes(get(item, "attributes")) + ">\n    " + out(get(item, "label")) +
		"\n  </a>\n</li>\n"
}

func tabPanel(item any, index int, idPrefix string) string {
	panel := get(item, "panel")
	var out_ strings.Builder
	out_.WriteString(`<div class="govuk-tabs__panel` +
		flagIf(" govuk-tabs__panel--hidden", index > 1) + `" id="` +
		escape(tabPanelID(item, index, idPrefix)) + `"` +
		Attributes(get(panel, "attributes")) + ">\n")
	switch html, text := get(panel, "html"), get(panel, "text"); {
	case truthy(html):
		out_.WriteString("  " + indent(trim(str(html)), 2, false) + "\n")
	case truthy(text):
		out_.WriteString(`  <p class="govuk-body">` + out(text) + "</p>\n")
	}
	out_.WriteString("</div>\n")
	return out_.String()
}

func renderTaskList(p *Params) string {
	idPrefix := "task-list"
	if prefix := p.Get("idPrefix"); truthy(prefix) {
		idPrefix = str(prefix)
	}

	var out_ strings.Builder
	out_.WriteString(`<ul class="govuk-task-list` + classesIf(p.Get("classes")) + `"` +
		Attributes(p.Get("attributes")) + ">\n")
	// Nunjucks keeps loop.index for skipped items and still emits a blank line per falsy item.
	for index, item := range items(p.Get("items")) {
		if truthy(item) {
			out_.WriteString(taskListItem(item, index+1, idPrefix) + "\n")
		} else {
			out_.WriteString("\n")
		}
	}
	out_.WriteString("</ul>")
	return out_.String()
}

func taskListItem(item any, index int, idPrefix string) string {
	position := strconv.Itoa(index)
	hintID := idPrefix + "-" + position + "-hint"
	statusID := idPrefix + "-" + position + "-status"
	title := get(item, "title")
	hint := get(item, "hint")
	status := get(item, "status")

	var out_ strings.Builder
	out_.WriteString(`  <li class="govuk-task-list__item` +
		flagIf(" govuk-task-list__item--with-link", get(item, "href")) +
		classesIf(get(item, "classes")) + "\">\n")
	out_.WriteString("    <div class=\"govuk-task-list__name-and-hint\">\n")

	if href := get(item, "href"); truthy(href) {
		describedBy := statusID
		if truthy(hint) {
			describedBy = hintID + " " + statusID
		}
		out_.WriteString(`      <a class="govuk-link govuk-task-list__link` +
			classesIf(get(title, "classes")) + `" href="` + out(href) +
			`" aria-describedby="` + escape(describedBy) + "\">\n")
		out_.WriteString("        " + contentIndent(title, "html", "text", 8) + "\n")
		out_.WriteString("      </a>\n")
	} else {
		out_.WriteString("      <div" + attributeIf("class", get(title, "classes")) + ">\n")
		out_.WriteString("        " + contentIndent(title, "html", "text", 8) + "\n")
		out_.WriteString("      </div>\n")
	}

	if truthy(hint) {
		out_.WriteString(`      <div id="` + escape(hintID) +
			"\" class=\"govuk-task-list__hint\">\n")
		out_.WriteString("        " + contentIndent(hint, "html", "text", 8) + "\n")
		out_.WriteString("      </div>\n")
	}
	out_.WriteString("    </div>\n")

	out_.WriteString(`    <div class="govuk-task-list__status` +
		classesIf(get(status, "classes")) + `" id="` + escape(statusID) + "\">\n")
	if tag := get(status, "tag"); truthy(tag) {
		tagParams, _ := tag.(*Params)
		out_.WriteString("      " + indent(trim(renderTag(tagParams)), 6, false) + "\n")
	} else {
		out_.WriteString("      " + contentIndent(status, "html", "text", 6) + "\n")
	}
	out_.WriteString("    </div>\n  </li>")
	return out_.String()
}
