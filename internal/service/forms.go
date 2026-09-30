package service

// Build GOV.UK component option maps for each question page.

import (
	"maps"

	"github.com/Nooshu/govuk-frontend-example-go/internal/htmlutil"
)

// Each …Fields function returns the options for the GOV.UK Frontend macros one question page
// needs, with the applicant's answer retained and any error message attached to the right field.

// ErrorSummary returns the error summary options, or false when the page has no errors.
func ErrorSummary(errors []FieldError) (map[string]any, bool) {
	if len(errors) == 0 {
		return nil, false
	}
	list := make([]any, 0, len(errors))
	for _, err := range errors {
		list = append(list, map[string]any{"text": err.Text, "href": err.Href})
	}
	return map[string]any{"titleText": "There is a problem", "errorList": list}, true
}

// NameField returns the text input on the full name question, whose label is the page h1.
func NameField(application Application, errors []FieldError) map[string]any {
	return map[string]any{
		"fullName": textInput("full-name", "What is your full name?", application.FullName, errors, map[string]any{
			"autocomplete": "name",
			"label": map[string]any{
				"text":          "What is your full name?",
				"isPageHeading": true,
				"classes":       "govuk-label--l",
			},
		}),
	}
}

// EmailField returns the text input on the email question, whose label is the page h1.
func EmailField(application Application, errors []FieldError) map[string]any {
	return map[string]any{
		"email": textInput("email", "What is your email address?", application.Email, errors, map[string]any{
			"type":         "email",
			"autocomplete": "email",
			"spellcheck":   false,
			"hint":         map[string]any{"text": "This example stores the address in your browser session only."},
			"label": map[string]any{
				"text":          "What is your email address?",
				"isPageHeading": true,
				"classes":       "govuk-label--l",
			},
		}),
	}
}

// DateField returns the date input on the date of birth question.
func DateField(application Application, errors []FieldError) map[string]any {
	date := map[string]any{
		"id":         "date-of-birth",
		"namePrefix": "date-of-birth",
		"fieldset": map[string]any{
			"legend": map[string]any{
				"text":          "What is your date of birth?",
				"isPageHeading": true,
				"classes":       "govuk-fieldset__legend--l",
			},
		},
		"hint": map[string]any{"text": "For example, 31 3 1980"},
		"items": []any{
			map[string]any{"name": "day", "value": application.Day},
			map[string]any{"name": "month", "value": application.Month},
			map[string]any{"name": "year", "value": application.Year},
		},
	}
	addError(date, errors, "date-of-birth")
	return map[string]any{"dateOfBirth": date}
}

// CountryFields returns the radios on the where-you-will-fish question.
func CountryFields(application Application, errors []FieldError) map[string]any {
	items := make([]any, 0, len(countries))
	for index, option := range countries {
		item := map[string]any{
			"value":   option.Value,
			"text":    option.Text,
			"checked": application.Country == option.Value,
		}
		if index == 0 {
			item["id"] = "country"
		}
		items = append(items, item)
	}
	radios := map[string]any{
		"idPrefix": "country",
		"name":     "country",
		"fieldset": map[string]any{
			"legend": map[string]any{
				"text":          "Where will you fish?",
				"isPageHeading": true,
				"classes":       "govuk-fieldset__legend--l",
			},
		},
		"hint":  map[string]any{"text": "This example is fictional. It does not check a real fishing area."},
		"items": items,
	}
	addError(radios, errors, "country")
	return map[string]any{"radios": radios}
}

// LicenceFields returns the radios on the licence length question (no fees).
func LicenceFields(application Application, errors []FieldError) map[string]any {
	items := make([]any, 0, len(licenceLengths))
	for index, option := range licenceLengths {
		item := map[string]any{
			"value":   option.Value,
			"text":    option.Text,
			"checked": application.LicenceLength == option.Value,
		}
		if index == 0 {
			item["id"] = "licence-length"
		}
		items = append(items, item)
	}
	radios := map[string]any{
		"idPrefix": "licence-length",
		"name":     "licence-length",
		"fieldset": map[string]any{
			"legend": map[string]any{
				"text":          "How long do you need the licence for?",
				"isPageHeading": true,
				"classes":       "govuk-fieldset__legend--l",
			},
		},
		"items": items,
	}
	addError(radios, errors, "licence-length")
	return map[string]any{"radios": radios}
}

// CookieFields returns the radios on the cookie settings page. choice is "accept", "reject", or
// empty when the applicant has not chosen.
func CookieFields(choice string, errors []FieldError) map[string]any {
	selected := ""
	switch choice {
	case "accept":
		selected = "yes"
	case "reject":
		selected = "no"
	}
	radios := map[string]any{
		"idPrefix": "analytics",
		"name":     "analytics",
		"fieldset": map[string]any{
			"legend": map[string]any{
				"text":          "Do you want to accept analytics cookies?",
				"isPageHeading": true,
				"classes":       "govuk-fieldset__legend--l",
			},
		},
		"hint": map[string]any{"text": "This example stores your choice. It does not set analytics cookies."},
		"items": []any{
			map[string]any{"value": "yes", "text": "Yes", "id": "analytics", "checked": selected == "yes"},
			map[string]any{"value": "no", "text": "No", "checked": selected == "no"},
		},
	}
	addError(radios, errors, "analytics")
	return map[string]any{"radios": radios}
}

// FeesTable returns the table of example licence fees.
func FeesTable() map[string]any {
	rows := make([]any, 0, len(licenceFees))
	for _, option := range licenceFees {
		rows = append(rows, []any{
			map[string]any{"text": option.Text},
			map[string]any{"text": option.Fee, "format": "numeric"},
		})
	}
	return map[string]any{
		"caption":           "Rod licence fees",
		"captionClasses":    "govuk-table__caption--m",
		"firstCellIsHeader": true,
		"head":              []any{map[string]any{"text": "Licence"}, map[string]any{"text": "Fee", "format": "numeric"}},
		"rows":              rows,
	}
}

// HelpAccordion returns the accordion on the help page.
func HelpAccordion() map[string]any {
	return map[string]any{
		"id": "help",
		"items": []any{
			map[string]any{
				"heading": map[string]any{"text": "Who can apply"},
				"content": map[string]any{
					"text": "You can apply if you are 13 or over and you will fish with a rod in England, Wales or Scotland.",
				},
			},
			map[string]any{
				"heading": map[string]any{"text": "What a licence covers"},
				"content": map[string]any{
					"html": `<ul class="govuk-list govuk-list--bullet"><li>Rod and line fishing</li><li>Up to 2 rods where the licence allows it</li><li>The dates printed on your licence</li></ul>`,
				},
			},
			map[string]any{
				"heading": map[string]any{"text": "If you need help to apply"},
				"content": map[string]any{
					"text": "You can ask someone to apply for you. This example service does not offer a phone application line.",
				},
			},
		},
	}
}

// GuidanceTabs returns the tabs on the guidance page.
func GuidanceTabs() map[string]any {
	return map[string]any{
		"id": "guidance",
		"items": []any{
			map[string]any{
				"label": "Before you apply",
				"id":    "before-you-apply",
				"panel": map[string]any{
					"html": `<h2 class="govuk-heading-l">Before you apply</h2><p class="govuk-body">You need how long you need the licence, your name, date of birth, the country where you will fish, and your email address.</p>`,
				},
			},
			map[string]any{
				"label": "Fees",
				"id":    "fees",
				"panel": map[string]any{
					"html": `<h2 class="govuk-heading-l">Fees</h2><p class="govuk-body">Fees depend on the length of the licence. <a class="govuk-link" href="/fees">See licence fees</a>.</p>`,
				},
			},
			map[string]any{
				"label": "After you apply",
				"id":    "after-you-apply",
				"panel": map[string]any{
					"html": `<h2 class="govuk-heading-l">After you apply</h2><p class="govuk-body">This example shows a confirmation page with a reference number. It does not send email and it does not take payment.</p>`,
				},
			},
		},
	}
}

// ConfirmationPanel returns the panel that acts as the confirmation page h1.
//
// The reference is escaped because the panel takes HTML, not text.
func ConfirmationPanel(reference string) map[string]any {
	return map[string]any{
		"titleText": "Application complete",
		"html":      "Your example reference number<br><strong>" + htmlutil.Escape(reference) + "</strong>",
	}
}

func textInput(id, label, value string, errors []FieldError, extra map[string]any) map[string]any {
	field := map[string]any{
		"id":    id,
		"name":  id,
		"label": map[string]any{"text": label},
		"value": value,
	}
	maps.Copy(field, extra)
	addError(field, errors, id)
	return field
}

func addError(params map[string]any, errors []FieldError, field string) {
	if message, ok := messageFor(errors, field); ok {
		params["errorMessage"] = map[string]any{"text": message}
	}
}

func messageFor(errors []FieldError, field string) (string, bool) {
	for _, err := range errors {
		if err.Field == field {
			return err.Text, true
		}
	}
	return "", false
}
