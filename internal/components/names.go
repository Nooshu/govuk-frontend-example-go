// Package components reads GOV.UK Frontend's own metadata: the component directories in the
// pinned package, the official fixtures each one ships, and the catalogue copy this example adds
// on top.
//
// The fixtures are the contract for component HTML. Nothing here renders markup; it supplies the
// options and the expected HTML that a renderer is measured against.
package components

import (
	"regexp"
	"strings"
	"unicode"
)

var componentName = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// IsComponentName reports whether name is a kebab-case GOV.UK Frontend component directory.
func IsComponentName(name string) bool {
	return componentName.MatchString(name)
}

// MacroName returns the Nunjucks macro name for a component, such as govukDateInput for
// "date-input".
func MacroName(component string) string {
	return "govuk" + strings.Join(capitaliseWords(component), "")
}

// TitleFromKebab turns a kebab-case component name into a title, such as "Date input" — used
// when a release adds a component the catalogue has no copy for yet.
func TitleFromKebab(component string) string {
	return strings.Join(capitaliseWords(component), " ")
}

func capitaliseWords(component string) []string {
	parts := strings.Split(component, "-")
	words := make([]string, 0, len(parts))
	for _, part := range parts {
		runes := []rune(part)
		if len(runes) > 0 {
			runes[0] = unicode.ToUpper(runes[0])
		}
		words = append(words, string(runes))
	}
	return words
}
