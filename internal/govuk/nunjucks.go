package govuk

// Nunjucks-parity filters: escape, indent, length, and related helpers.

import (
	"encoding/json"
	"strconv"
	"strings"
)

// escaper mirrors Nunjucks' escape map, backslash included. Using a single [strings.Replacer]
// pass also guarantees that `&` in an already-escaped entity is not escaped twice.
var escaper = strings.NewReplacer(
	"&", "&amp;",
	`"`, "&quot;",
	"'", "&#39;",
	"<", "&lt;",
	">", "&gt;",
	`\`, "&#92;",
)

// get walks a chain of option names, returning [Undefined] as soon as the chain leaves an object.
//
// It stands in for Nunjucks member lookup, where `params.fieldset.legend.text` is undefined
// rather than an error when `fieldset` was never supplied.
func get(value any, names ...string) any {
	for _, name := range names {
		object, ok := value.(*Params)
		if !ok {
			return Undefined
		}
		value = object.Get(name)
	}
	return value
}

// items returns value as a slice, or nil when it is not an array. Nunjucks silently skips a
// `{% for %}` over a missing option, so callers can range over the result unconditionally.
func items(value any) []any {
	list, _ := value.([]any)
	return list
}

// at returns the index'th element of an array option, or [Undefined] when it is out of range.
func at(value any, index int) any {
	list := items(value)
	if index < 0 || index >= len(list) {
		return Undefined
	}
	return list[index]
}

// truthy applies JavaScript truthiness, because Nunjucks compiles `{% if x %}` straight to a
// JavaScript `if`. Notably an empty array and an empty object are both truthy.
func truthy(value any) bool {
	switch typed := value.(type) {
	case nil, undefinedValue:
		return false
	case bool:
		return typed
	case string:
		return typed != ""
	case json.Number:
		number, err := typed.Float64()
		return err == nil && number != 0
	default:
		return true
	}
}

// isUndefined reports whether an option was never supplied, which several macros test with
// `!== undefined` to tell "absent" apart from "set to false".
func isUndefined(value any) bool {
	_, undefined := value.(undefinedValue)
	return undefined
}

// str converts a value to the string JavaScript would produce, before any escaping.
func str(value any) string {
	switch typed := value.(type) {
	case nil, undefinedValue:
		return ""
	case string:
		return typed
	case Safe:
		return string(typed)
	case bool:
		return strconv.FormatBool(typed)
	case json.Number:
		return number(typed)
	case []any:
		parts := make([]string, len(typed))
		for i, item := range typed {
			parts[i] = str(item)
		}
		return strings.Join(parts, ",")
	default:
		return "[object Object]"
	}
}

// number renders a JSON number the way JavaScript prints it, so `5` stays `5` and never
// becomes `5.0` or `5e+00`.
func number(value json.Number) string {
	if integer, err := strconv.ParseInt(value.String(), 10, 64); err == nil {
		return strconv.FormatInt(integer, 10)
	}
	float, err := value.Float64()
	if err != nil {
		return value.String()
	}
	return strconv.FormatFloat(float, 'f', -1, 64)
}

// out renders a value for a `{{ }}` expression in an autoescaping template: [Safe] values pass
// through untouched, everything else is escaped.
func out(value any) string {
	if safe, ok := value.(Safe); ok {
		return string(safe)
	}
	return escape(str(value))
}

// escape is Nunjucks' escape filter for a plain string.
func escape(text string) string {
	return escaper.Replace(text)
}

// trim is Nunjucks' trim filter.
func trim(text string) string {
	return strings.TrimSpace(text)
}

// indent is Nunjucks' indent filter: prefix every line with width spaces, skipping the first
// line unless first is true. An empty string is returned unchanged.
func indent(text string, width int, first bool) string {
	if text == "" {
		return ""
	}
	padding := strings.Repeat(" ", width)
	lines := strings.Split(text, "\n")
	for i := range lines {
		if i == 0 && !first {
			continue
		}
		lines[i] = padding + lines[i]
	}
	return strings.Join(lines, "\n")
}

// def is Nunjucks' `default(fallback)` filter: the fallback applies only when the option is
// undefined, so an explicit null or empty string is kept.
func def(value, fallback any) any {
	if isUndefined(value) {
		return fallback
	}
	return value
}

// defTruthy is Nunjucks' `default(fallback, true)` filter, which also replaces falsy values.
func defTruthy(value, fallback any) any {
	if truthy(value) {
		return value
	}
	return fallback
}

// length is Nunjucks' length filter: element count for arrays, key count for objects, and
// character count for strings.
func length(value any) int {
	switch typed := value.(type) {
	case nil, undefinedValue:
		return 0
	case bool:
		if !typed {
			return 0
		}
		return 0
	case []any:
		return len(typed)
	case *Params:
		return typed.Len()
	case string:
		return len(typed)
	case Safe:
		return len(typed)
	default:
		return 0
	}
}

// looseEq applies JavaScript `==` for the value kinds the macros compare: option values against
// a selected value, and booleans against `false`.
func looseEq(left, right any) bool {
	leftNil := left == nil || isUndefined(left)
	rightNil := right == nil || isUndefined(right)
	if leftNil || rightNil {
		return leftNil && rightNil
	}
	leftBool, leftIsBool := left.(bool)
	rightBool, rightIsBool := right.(bool)
	switch {
	case leftIsBool && rightIsBool:
		return leftBool == rightBool
	case leftIsBool:
		return looseEq(boolToNumber(leftBool), right)
	case rightIsBool:
		return looseEq(left, boolToNumber(rightBool))
	}
	leftNumber, leftIsNumber := left.(json.Number)
	rightNumber, rightIsNumber := right.(json.Number)
	switch {
	case leftIsNumber && rightIsNumber:
		return numeric(leftNumber.String()) == numeric(rightNumber.String())
	case leftIsNumber:
		return sameNumber(leftNumber.String(), str(right))
	case rightIsNumber:
		return sameNumber(str(left), rightNumber.String())
	}
	return str(left) == str(right)
}

// strictEq applies JavaScript `===`, which array membership (`in`) and `includes` both use.
func strictEq(left, right any) bool {
	if isUndefined(left) || isUndefined(right) {
		return isUndefined(left) && isUndefined(right)
	}
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	leftNumber, leftIsNumber := left.(json.Number)
	rightNumber, rightIsNumber := right.(json.Number)
	if leftIsNumber != rightIsNumber {
		return false
	}
	if leftIsNumber {
		return numeric(leftNumber.String()) == numeric(rightNumber.String())
	}
	switch typed := left.(type) {
	case bool:
		other, ok := right.(bool)
		return ok && typed == other
	case string:
		other, ok := right.(string)
		return ok && typed == other
	}
	return left == right
}

// contains applies the Nunjucks `in` operator: substring search for strings, strict element
// search for arrays, and key lookup for objects.
func contains(needle, haystack any) bool {
	switch typed := haystack.(type) {
	case string:
		return strings.Contains(typed, str(needle))
	case Safe:
		return strings.Contains(string(typed), str(needle))
	case []any:
		for _, item := range typed {
			if strictEq(needle, item) {
				return true
			}
		}
		return false
	case *Params:
		return typed.Has(str(needle))
	default:
		return false
	}
}

func boolToNumber(value bool) json.Number {
	if value {
		return json.Number("1")
	}
	return json.Number("0")
}

func numeric(text string) float64 {
	value, _ := strconv.ParseFloat(text, 64)
	return value
}

func sameNumber(left, right string) bool {
	trimmed := strings.TrimSpace(right)
	if trimmed == "" {
		trimmed = "0"
	}
	value, err := strconv.ParseFloat(trimmed, 64)
	if err != nil {
		return false
	}
	return numeric(left) == value
}
