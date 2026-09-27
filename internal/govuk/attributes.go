package govuk

// Attribute and i18n helpers matching Nunjucks macros.

import "strings"

// Attributes renders the `attributes` option the way GOV.UK Frontend's private
// `govukAttributes` macro does, returning HTML that already starts with a leading space when
// it is not empty.
//
// A string (or [Safe]) value is passed through untouched, which is how the character count and
// password input components forward a pre-rendered attribute string to the component they wrap.
//
// An object value is rendered key by key, in insertion order:
//
//   - ` name="value"` by default, with the value escaped;
//   - ` name` alone when the value is `{value: true, optional: true}`, giving an HTML boolean
//     attribute;
//   - nothing at all when the value is absent, null, or false and `optional` is true.
func Attributes(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case Safe:
		return string(typed)
	case *Params:
		var out strings.Builder
		for _, name := range typed.Keys() {
			out.WriteString(attribute(name, typed.Get(name)))
		}
		return out.String()
	default:
		return ""
	}
}

// attribute renders a single entry of the `attributes` option.
func attribute(name string, item any) string {
	value := item
	optional := false
	if options, ok := item.(*Params); ok {
		value = options.Get("value")
		flag, isBool := options.Get("optional").(bool)
		optional = isBool && flag
	}

	empty := value == nil || isUndefined(value)
	escaped := ""
	if !empty {
		if safe, ok := value.(Safe); ok {
			escaped = string(safe)
		} else {
			escaped = escape(str(value))
		}
	}

	if optional {
		flag, isBool := value.(bool)
		if isBool && flag {
			return " " + escape(name)
		}
		if empty || (isBool && !flag) {
			return ""
		}
	}
	return " " + escape(name) + `="` + escaped + `"`
}

// i18nAttributes renders translated text into `data-i18n.*` attributes, matching GOV.UK
// Frontend's private `govukI18nAttributes` macro.
//
// messages takes precedence over message and produces one attribute per CLDR plural rule.
func i18nAttributes(key string, message any, messages any) string {
	if truthy(messages) {
		object, ok := messages.(*Params)
		if !ok {
			return ""
		}
		var out strings.Builder
		for _, rule := range object.Keys() {
			out.WriteString(" data-i18n." + key + "." + rule + `="` + escape(str(object.Get(rule))) + `"`)
		}
		return out.String()
	}
	if truthy(message) {
		return " data-i18n." + key + `="` + escape(str(message)) + `"`
	}
	return ""
}

// attributeIf renders ` name="value"` when the option is truthy, the pattern the templates
// spell as `{%- if params.x %} name="{{ params.x }}"{% endif %}`.
func attributeIf(name string, value any) string {
	if !truthy(value) {
		return ""
	}
	return " " + name + `="` + out(value) + `"`
}

// classesIf renders ` some-class` when the option is truthy, for appending optional classes to
// a class attribute.
func classesIf(value any) string {
	if !truthy(value) {
		return ""
	}
	return " " + out(value)
}

// flagIf renders a literal suffix when the option is truthy, used for boolean attributes such
// as ` disabled` and for conditional class names.
func flagIf(suffix string, value any) string {
	if !truthy(value) {
		return ""
	}
	return suffix
}

// content renders the common `x.html | safe if x.html else x.text` choice: HTML is trusted and
// emitted as-is, text is escaped.
func content(params any, htmlKey, textKey string) string {
	if html := get(params, htmlKey); truthy(html) {
		return str(html)
	}
	return out(get(params, textKey))
}

// contentIndent renders `x.html | safe | trim | indent(width) if x.html else x.text`, the
// variant used wherever the HTML is nested inside an indented block.
func contentIndent(params any, htmlKey, textKey string, width int) string {
	if html := get(params, htmlKey); truthy(html) {
		return indent(trim(str(html)), width, false)
	}
	return out(get(params, textKey))
}
