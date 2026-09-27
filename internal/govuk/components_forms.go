package govuk

// Ports for form components (input, radios, date-input, …).

import (
	"strconv"
	"strings"
)

// This file ports the form components. They share a shape: a form group wrapper, then the
// label, hint, and error message, then the control itself. formGroupHeader renders that shared
// prefix and returns the accumulated aria-describedby value.

// formGroupOpen renders the opening `<div class="govuk-form-group">` common to every form
// component.
func formGroupOpen(p *Params) string {
	formGroup := p.Get("formGroup")
	return `<div class="govuk-form-group` +
		flagIf(" govuk-form-group--error", p.Get("errorMessage")) +
		classesIf(get(formGroup, "classes")) + `"` +
		Attributes(get(formGroup, "attributes")) + ">\n"
}

// describedByHint appends a hint's id to an aria-describedby list, the way the macros build it
// up with `{% set describedBy = describedBy + ' ' + hintId if describedBy else hintId %}`.
func describedByAppend(describedBy, id string) string {
	if describedBy != "" {
		return describedBy + " " + id
	}
	return id
}

// hintBlock renders the hint for a form component and returns it with the updated
// aria-describedby value.
func hintBlock(p *Params, id, describedBy string, width int) (string, string) {
	hint := p.Get("hint")
	if !truthy(hint) {
		return "", describedBy
	}
	hintID := id + "-hint"
	describedBy = describedByAppend(describedBy, hintID)
	html := renderHint(NewParams(
		"id", hintID,
		"classes", get(hint, "classes"),
		"attributes", get(hint, "attributes"),
		"html", get(hint, "html"),
		"text", get(hint, "text"),
	))
	return strings.Repeat(" ", width) + indent(trim(html), width, false) + "\n", describedBy
}

// errorBlock renders the error message for a form component and returns it with the updated
// aria-describedby value.
func errorBlock(p *Params, id, describedBy string, width int) (string, string) {
	message := p.Get("errorMessage")
	if !truthy(message) {
		return "", describedBy
	}
	errorID := id + "-error"
	describedBy = describedByAppend(describedBy, errorID)
	html := renderErrorMessage(NewParams(
		"id", errorID,
		"classes", get(message, "classes"),
		"attributes", get(message, "attributes"),
		"html", get(message, "html"),
		"text", get(message, "text"),
		"visuallyHiddenText", get(message, "visuallyHiddenText"),
	))
	return strings.Repeat(" ", width) + indent(trim(html), width, false) + "\n", describedBy
}

// labelBlock renders the label for a form component, indented to width.
func labelBlock(p *Params, id any, width int) string {
	label := p.Get("label")
	html := renderLabel(NewParams(
		"html", get(label, "html"),
		"text", get(label, "text"),
		"classes", get(label, "classes"),
		"isPageHeading", get(label, "isPageHeading"),
		"attributes", get(label, "attributes"),
		"for", id,
	))
	return strings.Repeat(" ", width) + indent(trim(html), width, false) + "\n"
}

// slotContent renders a formGroup.beforeInput / afterInput slot.
func slotContent(slot any, width int, indentFirst bool) string {
	if html := get(slot, "html"); truthy(html) {
		return indent(trim(str(html)), width, indentFirst)
	}
	return out(get(slot, "text"))
}

// componentID is the `params.id if params.id else params.name` fallback shared by the inputs.
func componentID(p *Params) any {
	if id := p.Get("id"); truthy(id) {
		return id
	}
	return p.Get("name")
}

func renderInput(p *Params) string {
	classNames := "govuk-input"
	if classes := p.Get("classes"); truthy(classes) {
		classNames += " " + str(classes)
	}
	if truthy(p.Get("errorMessage")) {
		classNames += " govuk-input--error"
	}

	id := componentID(p)
	describedBy := ""
	if supplied := p.Get("describedBy"); truthy(supplied) {
		describedBy = str(supplied)
	}

	formGroup := p.Get("formGroup")
	prefix, suffix := p.Get("prefix"), p.Get("suffix")
	beforeInput, afterInput := get(formGroup, "beforeInput"), get(formGroup, "afterInput")
	hasPrefix := truthy(prefix) && (truthy(get(prefix, "text")) || truthy(get(prefix, "html")))
	hasSuffix := truthy(suffix) && (truthy(get(suffix, "text")) || truthy(get(suffix, "html")))
	hasBefore := truthy(beforeInput) && (truthy(get(beforeInput, "text")) || truthy(get(beforeInput, "html")))
	hasAfter := truthy(afterInput) && (truthy(get(afterInput, "text")) || truthy(get(afterInput, "html")))

	var out_ strings.Builder
	out_.WriteString(formGroupOpen(p))
	out_.WriteString(labelBlock(p, id, 2))

	hint, describedBy := hintBlock(p, str(id), describedBy, 2)
	out_.WriteString(hint)
	errorMessage, describedBy := errorBlock(p, str(id), describedBy, 2)
	out_.WriteString(errorMessage)

	element := inputElement(p, classNames, id, describedBy)
	if hasPrefix || hasSuffix || hasBefore || hasAfter {
		wrapper := p.Get("inputWrapper")
		out_.WriteString(`  <div class="govuk-input__wrapper` +
			classesIf(get(wrapper, "classes")) + `"` +
			Attributes(get(wrapper, "attributes")) + ">\n")
		if hasBefore {
			out_.WriteString(slotContent(beforeInput, 4, true) + "\n")
		}
		if hasPrefix {
			out_.WriteString(indent(affixItem(prefix, "prefix"), 2, true) + "\n")
		}
		out_.WriteString("    " + element + "\n")
		if hasSuffix {
			out_.WriteString(indent(affixItem(suffix, "suffix"), 2, true) + "\n")
		}
		if hasAfter {
			out_.WriteString(slotContent(afterInput, 4, true) + "\n")
		}
		out_.WriteString("  </div>\n")
	} else {
		out_.WriteString("  " + element + "\n")
	}

	out_.WriteString("</div>")
	return out_.String()
}

// inputElement renders the `<input>` itself. Every attribute goes through the attributes macro
// so that order, escaping, and the optional/boolean rules stay identical to upstream.
func inputElement(p *Params, classNames string, id any, describedBy string) string {
	spellcheck := any(false)
	if flag, ok := p.Get("spellcheck").(bool); ok {
		spellcheck = strconv.FormatBool(flag)
	}
	var ariaDescribedBy any = Undefined
	if describedBy != "" {
		ariaDescribedBy = describedBy
	}

	attributes := NewParams(
		"class", classNames,
		"id", id,
		"name", p.Get("name"),
		"type", defTruthy(p.Get("type"), "text"),
		"spellcheck", NewParams("value", spellcheck, "optional", true),
		"value", NewParams("value", p.Get("value"), "optional", true),
		"disabled", NewParams("value", p.Get("disabled"), "optional", true),
		"aria-describedby", NewParams("value", ariaDescribedBy, "optional", true),
		"autocomplete", NewParams("value", p.Get("autocomplete"), "optional", true),
		"autocapitalize", NewParams("value", p.Get("autocapitalize"), "optional", true),
		"pattern", NewParams("value", p.Get("pattern"), "optional", true),
		"inputmode", NewParams("value", p.Get("inputmode"), "optional", true),
	)
	return "<input" + Attributes(attributes) + Attributes(p.Get("attributes")) + ">"
}

func affixItem(affix any, kind string) string {
	return `  <div class="govuk-input__` + kind + classesIf(get(affix, "classes")) +
		`" aria-hidden="true"` + Attributes(get(affix, "attributes")) + ">" +
		contentIndent(affix, "html", "text", 4) + "</div>"
}

func renderTextarea(p *Params) string {
	id := componentID(p)
	describedBy := ""
	if supplied := p.Get("describedBy"); truthy(supplied) {
		describedBy = str(supplied)
	}
	formGroup := p.Get("formGroup")

	var out_ strings.Builder
	out_.WriteString(formGroupOpen(p))
	out_.WriteString(labelBlock(p, id, 2))

	hint, describedBy := hintBlock(p, str(id), describedBy, 2)
	out_.WriteString(hint)
	errorMessage, describedBy := errorBlock(p, str(id), describedBy, 2)
	out_.WriteString(errorMessage)

	if before := get(formGroup, "beforeInput"); truthy(before) {
		out_.WriteString("  " + slotContent(before, 2, false) + "\n")
	}

	spellcheck := ""
	if flag, ok := p.Get("spellcheck").(bool); ok {
		spellcheck = ` spellcheck="` + strconv.FormatBool(flag) + `"`
	}
	out_.WriteString(`  <textarea class="govuk-textarea` +
		flagIf(" govuk-textarea--error", p.Get("errorMessage")) +
		classesIf(p.Get("classes")) + `" id="` + out(id) + `" name="` + out(p.Get("name")) +
		`" rows="` + out(defTruthy(p.Get("rows"), "5")) + `"` + spellcheck +
		flagIf(" disabled", p.Get("disabled")) +
		attributeIf("aria-describedby", describedBy) +
		attributeIf("autocomplete", p.Get("autocomplete")) +
		Attributes(p.Get("attributes")) + ">" + out(p.Get("value")) + "</textarea>\n")

	if after := get(formGroup, "afterInput"); truthy(after) {
		out_.WriteString("  " + slotContent(after, 2, false) + "\n")
	}

	out_.WriteString("</div>")
	return out_.String()
}

func renderSelect(p *Params) string {
	id := componentID(p)
	describedBy := ""
	if supplied := p.Get("describedBy"); truthy(supplied) {
		describedBy = str(supplied)
	}
	formGroup := p.Get("formGroup")

	var out_ strings.Builder
	out_.WriteString(formGroupOpen(p))
	out_.WriteString(labelBlock(p, id, 2))

	hint, describedBy := hintBlock(p, str(id), describedBy, 2)
	out_.WriteString(hint)
	errorMessage, describedBy := errorBlock(p, str(id), describedBy, 2)
	out_.WriteString(errorMessage)

	if before := get(formGroup, "beforeInput"); truthy(before) {
		out_.WriteString("  " + slotContent(before, 2, false) + "\n")
	}

	out_.WriteString(`  <select class="govuk-select` + classesIf(p.Get("classes")) +
		flagIf(" govuk-select--error", p.Get("errorMessage")) + `" id="` + out(id) +
		`" name="` + out(p.Get("name")) + `"` +
		flagIf(" disabled", p.Get("disabled")) +
		attributeIf("aria-describedby", describedBy) +
		Attributes(p.Get("attributes")) + ">\n")

	selected := p.Get("value")
	for _, item := range items(p.Get("items")) {
		if !truthy(item) {
			continue
		}
		value := get(item, "value")
		// An option can be selected by its value or, when it has none, by its text content.
		effective := def(value, get(item, "text"))
		isSelected := truthy(get(item, "selected"))
		if !isSelected && truthy(selected) {
			isSelected = looseEq(effective, selected) && !looseEq(get(item, "selected"), false)
		}

		out_.WriteString("    <option")
		if !isUndefined(value) {
			out_.WriteString(` value="` + out(value) + `"`)
		}
		out_.WriteString(flagIf(" selected", isSelected) +
			flagIf(" disabled", get(item, "disabled")) +
			Attributes(get(item, "attributes")) + ">" + out(get(item, "text")) + "</option>\n")
	}
	out_.WriteString("  </select>\n")

	if after := get(formGroup, "afterInput"); truthy(after) {
		out_.WriteString("  " + slotContent(after, 2, false) + "\n")
	}

	out_.WriteString("</div>")
	return out_.String()
}

func renderFileUpload(p *Params) string {
	id := componentID(p)
	describedBy := ""
	if supplied := p.Get("describedBy"); truthy(supplied) {
		describedBy = str(supplied)
	}
	formGroup := p.Get("formGroup")

	var out_ strings.Builder
	out_.WriteString(formGroupOpen(p))
	out_.WriteString(labelBlock(p, id, 2))

	hint, describedBy := hintBlock(p, str(id), describedBy, 2)
	out_.WriteString(hint)
	errorMessage, describedBy := errorBlock(p, str(id), describedBy, 2)
	out_.WriteString(errorMessage)

	if before := get(formGroup, "beforeInput"); truthy(before) {
		out_.WriteString("  " + slotContent(before, 2, false) + "\n")
	}

	// The enhanced, JavaScript-driven variant wraps the input so the script can replace the
	// browser's own file picker with a GOV.UK button and a drop zone.
	javascript := truthy(p.Get("javascript"))
	if javascript {
		out_.WriteString("  <div\n    class=\"govuk-file-upload-wrapper" +
			classesIf(p.Get("wrapperClasses")) + "\"\n    data-module=\"govuk-file-upload\"" +
			i18nAttributes("choose-files-button", p.Get("chooseFilesButtonText"), Undefined) +
			i18nAttributes("no-file-chosen", p.Get("noFileChosenText"), Undefined) +
			i18nAttributes("multiple-files-chosen", Undefined, p.Get("multipleFilesChosenText")) +
			i18nAttributes("drop-instruction", p.Get("dropInstructionText"), Undefined) +
			i18nAttributes("entered-drop-zone", p.Get("enteredDropZoneText"), Undefined) +
			i18nAttributes("left-drop-zone", p.Get("leftDropZoneText"), Undefined) +
			Attributes(p.Get("wrapperAttributes")) + "\n  >\n")
	}

	out_.WriteString(`  <input class="govuk-file-upload` + classesIf(p.Get("classes")) +
		flagIf(" govuk-file-upload--error", p.Get("errorMessage")) + `" id="` + out(id) +
		`" name="` + out(p.Get("name")) + `" type="file"` +
		flagIf(" disabled", p.Get("disabled")) +
		flagIf(" multiple", p.Get("multiple")) +
		attributeIf("aria-describedby", describedBy) +
		Attributes(p.Get("attributes")) + ">\n")

	if javascript {
		out_.WriteString("  </div>\n")
	}
	if after := get(formGroup, "afterInput"); truthy(after) {
		out_.WriteString("  " + slotContent(after, 2, false) + "\n")
	}

	out_.WriteString("</div>")
	return out_.String()
}

func renderCharacterCount(p *Params) string {
	maxwords, maxlength := p.Get("maxwords"), p.Get("maxlength")
	hasNoLimit := !truthy(maxwords) && !truthy(maxlength)
	id := componentID(p)

	// Without a limit the count message can only be built in the browser, so the server-side
	// hint is left empty and the text is handed to the script as a translation instead.
	var descriptionNoLimit any = Undefined
	if !hasNoLimit {
		limit := maxlength
		if truthy(maxwords) {
			limit = maxwords
		}
		unit := "characters"
		if truthy(maxwords) {
			unit = "words"
		}
		description := "You can enter up to %{count} " + unit
		if supplied := p.Get("textareaDescriptionText"); truthy(supplied) {
			description = str(supplied)
		}
		descriptionNoLimit = strings.ReplaceAll(description, "%{count}", str(limit))
	}

	countMessage := p.Get("countMessage")
	countMessageHTML := trim(renderHint(NewParams(
		"text", descriptionNoLimit,
		"id", str(id)+"-info",
		"classes", "govuk-character-count__message"+concatIf(" ", get(countMessage, "classes")),
	))) + "\n"

	formGroup := p.Get("formGroup")
	if after := get(formGroup, "afterInput"); truthy(after) {
		if html := get(after, "html"); truthy(html) {
			countMessageHTML += trim(str(html)) + "\n"
		} else {
			countMessageHTML += out(get(after, "text")) + "\n"
		}
	}

	attributesHTML := Attributes(NewParams(
		"data-module", "govuk-character-count",
		"data-maxlength", NewParams("value", maxlength, "optional", true),
		"data-threshold", NewParams("value", p.Get("threshold"), "optional", true),
		"data-maxwords", NewParams("value", maxwords, "optional", true),
	))
	if description := p.Get("textareaDescriptionText"); hasNoLimit && truthy(description) {
		attributesHTML += i18nAttributes("textarea-description", Undefined,
			NewParams("other", description))
	}
	attributesHTML += i18nAttributes("characters-under-limit", Undefined, p.Get("charactersUnderLimitText")) +
		i18nAttributes("characters-at-limit", p.Get("charactersAtLimitText"), Undefined) +
		i18nAttributes("characters-over-limit", Undefined, p.Get("charactersOverLimitText")) +
		i18nAttributes("words-under-limit", Undefined, p.Get("wordsUnderLimitText")) +
		i18nAttributes("words-at-limit", p.Get("wordsAtLimitText"), Undefined) +
		i18nAttributes("words-over-limit", Undefined, p.Get("wordsOverLimitText"))
	attributesHTML += appendedAttributes(get(formGroup, "attributes"))

	label := p.Get("label")
	return trim(renderTextarea(NewParams(
		"id", id,
		"name", p.Get("name"),
		"describedBy", str(id)+"-info",
		"rows", p.Get("rows"),
		"spellcheck", p.Get("spellcheck"),
		"value", p.Get("value"),
		"formGroup", NewParams(
			"classes", "govuk-character-count"+concatIf(" ", get(formGroup, "classes")),
			"attributes", attributesHTML,
			"beforeInput", get(formGroup, "beforeInput"),
			"afterInput", NewParams("html", Safe(countMessageHTML)),
		),
		"classes", "govuk-js-character-count"+concatIf(" ", p.Get("classes")),
		"label", NewParams(
			"html", get(label, "html"),
			"text", get(label, "text"),
			"classes", get(label, "classes"),
			"isPageHeading", get(label, "isPageHeading"),
			"attributes", get(label, "attributes"),
			"for", id,
		),
		"hint", p.Get("hint"),
		"errorMessage", p.Get("errorMessage"),
		"attributes", p.Get("attributes"),
	)))
}

// appendedAttributes renders a caller's formGroup.attributes onto an attribute string that the
// component has already built, which is how character count and password input merge the two.
func appendedAttributes(attributes any) string {
	object, ok := attributes.(*Params)
	if !ok {
		return ""
	}
	var out_ strings.Builder
	for _, name := range object.Keys() {
		out_.WriteString(" " + escape(name) + `="` + escape(str(object.Get(name))) + `"`)
	}
	return out_.String()
}

func renderPasswordInput(p *Params) string {
	id := componentID(p)
	formGroup := p.Get("formGroup")

	attributesHTML := ` data-module="govuk-password-input"` +
		i18nAttributes("show-password", p.Get("showPasswordText"), Undefined) +
		i18nAttributes("hide-password", p.Get("hidePasswordText"), Undefined) +
		i18nAttributes("show-password-aria-label", p.Get("showPasswordAriaLabelText"), Undefined) +
		i18nAttributes("hide-password-aria-label", p.Get("hidePasswordAriaLabelText"), Undefined) +
		i18nAttributes("password-shown-announcement", p.Get("passwordShownAnnouncementText"), Undefined) +
		i18nAttributes("password-hidden-announcement", p.Get("passwordHiddenAnnouncementText"), Undefined) +
		appendedAttributes(get(formGroup, "attributes"))

	// The toggle starts hidden and is revealed by the component's script, so the control is
	// never offered to someone whose browser cannot operate it.
	button := p.Get("button")
	buttonHTML := trim(renderButton(NewParams(
		"type", "button",
		"classes", "govuk-button--secondary govuk-password-input__toggle govuk-js-password-input-toggle"+
			concatIf(" ", get(button, "classes")),
		"text", def(p.Get("showPasswordText"), "Show"),
		"attributes", NewParams(
			"aria-controls", id,
			"aria-label", def(p.Get("showPasswordAriaLabelText"), "Show password"),
			"hidden", NewParams("value", true, "optional", true),
		),
	))) + "\n"
	if after := get(formGroup, "afterInput"); truthy(after) {
		if html := get(after, "html"); truthy(html) {
			buttonHTML += trim(str(html)) + "\n"
		} else {
			buttonHTML += out(get(after, "text")) + "\n"
		}
	}

	return trim(renderInput(NewParams(
		"formGroup", NewParams(
			"classes", "govuk-password-input"+concatIf(" ", get(formGroup, "classes")),
			"attributes", attributesHTML,
			"beforeInput", get(formGroup, "beforeInput"),
			"afterInput", NewParams("html", Safe(buttonHTML)),
		),
		"inputWrapper", NewParams("classes", "govuk-password-input__wrapper"),
		"label", p.Get("label"),
		"hint", p.Get("hint"),
		"classes", "govuk-password-input__input govuk-js-password-input-input"+
			concatIf(" ", p.Get("classes")),
		"errorMessage", p.Get("errorMessage"),
		"id", id,
		"name", p.Get("name"),
		"type", "password",
		"spellcheck", false,
		"autocapitalize", "none",
		"autocomplete", defTruthy(p.Get("autocomplete"), "current-password"),
		"value", p.Get("value"),
		"disabled", p.Get("disabled"),
		"describedBy", p.Get("describedBy"),
		"attributes", p.Get("attributes"),
	)))
}

func renderCheckboxes(p *Params) string {
	idPrefix := p.Get("idPrefix")
	if !truthy(idPrefix) {
		idPrefix = p.Get("name")
	}
	fieldset := p.Get("fieldset")
	describedBy := ""
	if supplied := p.Get("describedBy"); truthy(supplied) {
		describedBy = str(supplied)
	}
	if supplied := get(fieldset, "describedBy"); truthy(supplied) {
		describedBy = str(supplied)
	}
	hasFieldset := truthy(fieldset)

	var inner strings.Builder
	hint, describedBy := hintBlock(p, str(idPrefix), describedBy, 2)
	inner.WriteString(hint)
	errorMessage, describedBy := errorBlock(p, str(idPrefix), describedBy, 2)
	inner.WriteString(errorMessage)

	formGroup := p.Get("formGroup")
	inner.WriteString(`  <div class="govuk-checkboxes` + classesIf(p.Get("classes")) + `"` +
		Attributes(p.Get("attributes")) + " data-module=\"govuk-checkboxes\">\n")
	if before := get(formGroup, "beforeInputs"); truthy(before) {
		inner.WriteString("    " + slotContent(before, 4, false) + "\n")
	}
	for index, item := range items(p.Get("items")) {
		if !truthy(item) {
			continue
		}
		inner.WriteString(checkboxItem(p, item, index+1, str(idPrefix), describedBy, hasFieldset))
	}
	if after := get(formGroup, "afterInputs"); truthy(after) {
		inner.WriteString("    " + slotContent(after, 4, false) + "\n")
	}
	inner.WriteString("  </div>\n")

	return fieldsetWrapper(p, inner.String(), describedBy, Undefined, false)
}

// fieldsetWrapper closes a checkboxes, radios, or date input component: the captured inner HTML
// either goes straight into the form group or is nested in a fieldset with the legend.
func fieldsetWrapper(p *Params, inner, describedBy string, role any, indentFieldset bool) string {
	fieldset := p.Get("fieldset")
	body := trim(inner)
	if truthy(fieldset) {
		html := renderFieldset(NewParams(
			"describedBy", describedBy,
			"classes", get(fieldset, "classes"),
			"role", role,
			"attributes", get(fieldset, "attributes"),
			"legend", get(fieldset, "legend"),
			"html", Safe(body),
		))
		body = trim(html)
		if indentFieldset {
			body = indent(body, 2, false)
		}
	}
	return formGroupOpen(p) + "  " + body + "\n</div>"
}

func checkboxItem(p *Params, item any, index int, idPrefix, describedBy string, hasFieldset bool) string {
	itemID := str(idPrefix)
	if index > 1 {
		itemID += "-" + strconv.Itoa(index)
	}
	if supplied := get(item, "id"); truthy(supplied) {
		itemID = str(supplied)
	}
	itemName := p.Get("name")
	if supplied := get(item, "name"); truthy(supplied) {
		itemName = supplied
	}
	conditionalID := "conditional-" + itemID

	if divider := get(item, "divider"); truthy(divider) {
		return `    <div class="govuk-checkboxes__divider">` + out(divider) + "</div>\n"
	}

	checked := truthy(get(item, "checked"))
	if !checked && truthy(p.Get("values")) {
		checked = contains(get(item, "value"), p.Get("values")) &&
			!looseEq(get(item, "checked"), false)
	}
	hint := get(item, "hint")
	hasHint := truthy(get(hint, "text")) || truthy(get(hint, "html"))
	itemHintID := ""
	if hasHint {
		itemHintID = itemID + "-item-hint"
	}
	itemDescribedBy := ""
	if !hasFieldset {
		itemDescribedBy = describedBy
	}
	itemDescribedBy = trim(itemDescribedBy + " " + itemHintID)

	conditional := get(item, "conditional")
	label := get(item, "label")

	var out_ strings.Builder
	out_.WriteString("    <div class=\"govuk-checkboxes__item\">\n")
	out_.WriteString(`      <input class="govuk-checkboxes__input" id="` + escape(itemID) +
		`" name="` + out(itemName) + `" type="checkbox" value="` + out(get(item, "value")) + `"` +
		flagIf(" checked", checked) +
		flagIf(" disabled", get(item, "disabled")) +
		attributeIf("data-aria-controls", ifTruthy(get(conditional, "html"), conditionalID)) +
		attributeIf("data-behaviour", get(item, "behaviour")) +
		attributeIf("aria-describedby", itemDescribedBy) +
		Attributes(get(item, "attributes")) + ">\n")
	out_.WriteString("      " + indent(trim(renderLabel(NewParams(
		"html", get(item, "html"),
		"text", get(item, "text"),
		"classes", "govuk-checkboxes__label"+concatIf(" ", get(label, "classes")),
		"attributes", get(label, "attributes"),
		"for", itemID,
	))), 6, false) + "\n")
	if hasHint {
		out_.WriteString("      " + indent(trim(renderHint(NewParams(
			"id", itemHintID,
			"classes", "govuk-checkboxes__hint"+concatIf(" ", get(hint, "classes")),
			"attributes", get(hint, "attributes"),
			"html", get(hint, "html"),
			"text", get(hint, "text"),
		))), 6, false) + "\n")
	}
	out_.WriteString("    </div>\n")
	if html := get(conditional, "html"); truthy(html) {
		out_.WriteString(`    <div class="govuk-checkboxes__conditional` +
			flagIf(" govuk-checkboxes__conditional--hidden", !checked) + `" id="` +
			escape(conditionalID) + "\">\n      " + trim(str(html)) + "\n    </div>\n")
	}
	return out_.String()
}

// ifTruthy returns value when condition is truthy, so an attribute is only rendered when the
// option it depends on is present.
func ifTruthy(condition any, value string) any {
	if truthy(condition) {
		return value
	}
	return Undefined
}

func renderRadios(p *Params) string {
	idPrefix := p.Get("idPrefix")
	if !truthy(idPrefix) {
		idPrefix = p.Get("name")
	}
	fieldset := p.Get("fieldset")
	describedBy := ""
	if supplied := get(fieldset, "describedBy"); truthy(supplied) {
		describedBy = str(supplied)
	}

	var inner strings.Builder
	hint, describedBy := hintBlock(p, str(idPrefix), describedBy, 2)
	inner.WriteString(hint)
	errorMessage, describedBy := errorBlock(p, str(idPrefix), describedBy, 2)
	inner.WriteString(errorMessage)

	formGroup := p.Get("formGroup")
	inner.WriteString(`  <div class="govuk-radios` + classesIf(p.Get("classes")) + `"` +
		Attributes(p.Get("attributes")) + " data-module=\"govuk-radios\">\n")
	if before := get(formGroup, "beforeInputs"); truthy(before) {
		inner.WriteString("    " + slotContent(before, 4, false) + "\n")
	}
	for index, item := range items(p.Get("items")) {
		if !truthy(item) {
			continue
		}
		inner.WriteString(radioItem(p, item, index+1, str(idPrefix)))
	}
	if after := get(formGroup, "afterInputs"); truthy(after) {
		inner.WriteString("    " + slotContent(after, 4, false) + "\n")
	}
	inner.WriteString("  </div>\n")

	return fieldsetWrapper(p, inner.String(), describedBy, Undefined, false)
}

func radioItem(p *Params, item any, index int, idPrefix string) string {
	itemID := idPrefix
	if index > 1 {
		itemID += "-" + strconv.Itoa(index)
	}
	if supplied := get(item, "id"); truthy(supplied) {
		itemID = str(supplied)
	}
	conditionalID := "conditional-" + itemID

	if divider := get(item, "divider"); truthy(divider) {
		return `    <div class="govuk-radios__divider">` + out(divider) + "</div>\n"
	}

	checked := truthy(get(item, "checked"))
	if !checked && truthy(p.Get("value")) {
		checked = looseEq(get(item, "value"), p.Get("value")) &&
			!looseEq(get(item, "checked"), false)
	}
	hint := get(item, "hint")
	hasHint := truthy(get(hint, "text")) || truthy(get(hint, "html"))
	itemHintID := itemID + "-item-hint"
	conditional := get(item, "conditional")
	label := get(item, "label")

	var out_ strings.Builder
	out_.WriteString("    <div class=\"govuk-radios__item\">\n")
	out_.WriteString(`      <input class="govuk-radios__input" id="` + escape(itemID) +
		`" name="` + out(p.Get("name")) + `" type="radio" value="` +
		out(get(item, "value")) + `"` +
		flagIf(" checked", checked) +
		flagIf(" disabled", get(item, "disabled")) +
		attributeIf("data-aria-controls", ifTruthy(get(conditional, "html"), conditionalID)) +
		attributeIf("aria-describedby", ifTruthy(hasHint, itemHintID)) +
		Attributes(get(item, "attributes")) + ">\n")
	out_.WriteString("      " + indent(trim(renderLabel(NewParams(
		"html", get(item, "html"),
		"text", get(item, "text"),
		"classes", "govuk-radios__label"+concatIf(" ", get(label, "classes")),
		"attributes", get(label, "attributes"),
		"for", itemID,
	))), 6, false) + "\n")
	if hasHint {
		out_.WriteString("      " + indent(trim(renderHint(NewParams(
			"id", itemHintID,
			"classes", "govuk-radios__hint"+concatIf(" ", get(hint, "classes")),
			"attributes", get(hint, "attributes"),
			"html", get(hint, "html"),
			"text", get(hint, "text"),
		))), 6, false) + "\n")
	}
	out_.WriteString("    </div>\n")
	if html := get(conditional, "html"); truthy(html) {
		out_.WriteString(`    <div class="govuk-radios__conditional` +
			flagIf(" govuk-radios__conditional--hidden", !checked) + `" id="` +
			escape(conditionalID) + "\">\n      " + trim(str(html)) + "\n    </div>\n")
	}
	return out_.String()
}

func renderDateInput(p *Params) string {
	fieldset := p.Get("fieldset")
	describedBy := ""
	if supplied := get(fieldset, "describedBy"); truthy(supplied) {
		describedBy = str(supplied)
	}
	values := p.Get("values")

	day := def(p.Get("day"), NewParams(
		"name", "day", "value", get(values, "day"), "classes", "govuk-input--width-2"))
	month := def(p.Get("month"), NewParams(
		"name", "month", "value", get(values, "month"), "classes", "govuk-input--width-2"))
	year := def(p.Get("year"), NewParams(
		"name", "year", "value", get(values, "year"), "classes", "govuk-input--width-4"))

	dateItems := items(p.Get("items"))
	if len(dateItems) == 0 {
		dateItems = []any{day, month, year}
	}

	anyItemHasError := false
	for _, item := range dateItems {
		if dateItemHasError(item) {
			anyItemHasError = true
		}
	}

	var inner strings.Builder
	hint, describedBy := hintBlock(p, str(p.Get("id")), describedBy, 2)
	inner.WriteString(hint)
	errorMessage, describedBy := errorBlock(p, str(p.Get("id")), describedBy, 2)
	inner.WriteString(errorMessage)

	formGroup := p.Get("formGroup")
	inner.WriteString(`  <div class="govuk-date-input` + classesIf(p.Get("classes")) + `"` +
		Attributes(p.Get("attributes")) + attributeIf("id", p.Get("id")) + ">\n")
	if before := get(formGroup, "beforeInputs"); truthy(before) {
		inner.WriteString("    " + slotContent(before, 4, false) + "\n")
	}
	for _, item := range dateItems {
		if !truthy(item) {
			continue
		}
		inner.WriteString(indent(trim(dateInputItem(p, item, anyItemHasError, day, month, year)),
			4, true) + "\n")
	}
	if after := get(formGroup, "afterInputs"); truthy(after) {
		inner.WriteString("    " + slotContent(after, 4, false) + "\n")
	}
	inner.WriteString("  </div>\n")

	// The fieldset's role is overridden to "group" because otherwise JAWS does not announce the
	// description for a fieldset made of text inputs.
	return fieldsetWrapper(p, inner.String(), describedBy, "group", true)
}

func dateItemHasError(item any) bool {
	if truthy(get(item, "error")) {
		return true
	}
	classes := get(item, "classes")
	return truthy(classes) && contains("govuk-input--error", classes)
}

func dateInputItem(p *Params, item any, anyItemHasError bool, day, month, year any) string {
	itemName := get(item, "name")
	itemValue := get(item, "value")
	itemWidth := "2"
	itemClasses := ""
	itemHasError := dateItemHasError(item)

	// An item is recognised as the day, month, or year part either by identity with the
	// defaults or by its name, which decides its width and where its value comes from.
	switch name := get(item, "name"); {
	case item == day || (truthy(name) && contains(name, []any{"day", get(day, "name")})):
		itemName = def(name, "day")
		itemValue = def(itemValue, get(day, "value"))
	case item == month || (truthy(name) && contains(name, []any{"month", get(month, "name")})):
		itemName = def(name, "month")
		itemValue = def(itemValue, get(month, "value"))
	case item == year || (truthy(name) && contains(name, []any{"year", get(year, "name")})):
		itemName = def(name, "year")
		itemValue = def(itemValue, get(year, "value"))
		itemWidth = "4"
	}

	classes := get(item, "classes")
	hasErrorClass := truthy(classes) && contains("govuk-input--error", classes)
	if !hasErrorClass && (itemHasError ||
		(!looseEq(get(item, "error"), false) && truthy(p.Get("errorMessage")) && !anyItemHasError)) {
		itemClasses = trim(itemClasses + " govuk-input--error")
	}
	if !truthy(classes) || !contains("govuk-input--width-", classes) {
		itemClasses = trim(itemClasses + " govuk-input--width-" + itemWidth)
	}
	if truthy(classes) {
		itemClasses = trim(itemClasses + " " + str(classes))
	}

	namePrefix := ""
	if prefix := p.Get("namePrefix"); truthy(prefix) {
		namePrefix = str(prefix) + "-"
	}

	label := get(item, "label")
	if !truthy(label) {
		label = capitalise(str(itemName))
	}
	id := get(item, "id")
	if !truthy(id) {
		id = str(p.Get("id")) + "-" + str(itemName)
	}
	value := itemValue
	if isUndefined(value) {
		value = get(p.Get("values"), namePrefix+str(itemName))
	}

	input := renderInput(NewParams(
		"label", NewParams("text", label, "classes", "govuk-date-input__label"),
		"id", id,
		"classes", "govuk-date-input__input"+concatIf(" ", itemClasses),
		"name", namePrefix+str(itemName),
		"value", value,
		"type", "text",
		"inputmode", defTruthy(get(item, "inputmode"), "numeric"),
		"autocomplete", get(item, "autocomplete"),
		"pattern", get(item, "pattern"),
		"attributes", get(item, "attributes"),
	))

	return "<div class=\"govuk-date-input__item\">\n  " + indent(trim(input), 2, false) +
		"\n</div>"
}

// capitalise is Nunjucks' capitalize filter: lower-case the string, then upper-case the first
// character.
func capitalise(text string) string {
	if text == "" {
		return ""
	}
	lowered := strings.ToLower(text)
	return strings.ToUpper(lowered[:1]) + lowered[1:]
}
