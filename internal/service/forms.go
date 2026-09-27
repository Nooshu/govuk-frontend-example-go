package service

import (
	"maps"
	"time"

	"github.com/Nooshu/govuk-frontend-example-go/internal/htmlutil"
	"github.com/Nooshu/govuk-frontend-example-go/internal/render"
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

// NameFields returns the two text inputs on the name question.
func NameFields(application Application, errors []FieldError) map[string]any {
	return map[string]any{
		"firstName": textInput("first-name", "First name", application.FirstName, errors, map[string]any{
			"autocomplete": "given-name",
			"classes":      "govuk-input--width-20",
			"spellcheck":   false,
		}),
		"lastName": textInput("last-name", "Last name", application.LastName, errors, map[string]any{
			"autocomplete": "family-name",
			"classes":      "govuk-input--width-20",
			"spellcheck":   false,
		}),
	}
}

// EmailField returns the text input on the email question, whose label is the page h1.
func EmailField(application Application, errors []FieldError) map[string]any {
	return map[string]any{
		"email": textInput("email", "Email address", application.Email, errors, map[string]any{
			"type":         "email",
			"autocomplete": "email",
			"spellcheck":   false,
			"classes":      "govuk-input--width-20",
			"hint":         map[string]any{"text": "We will send the decision to this address"},
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
			map[string]any{"name": "day", "autocomplete": "bday-day", "value": application.Day},
			map[string]any{"name": "month", "autocomplete": "bday-month", "value": application.Month},
			map[string]any{"name": "year", "autocomplete": "bday-year", "value": application.Year},
		},
	}
	addError(date, errors, "date-of-birth")
	return map[string]any{"dateOfBirth": date}
}

// ContactFields returns the radios on the contact preference question.
//
// The telephone input is rendered here rather than in the page, because it is revealed by the
// telephone radio and so has to be passed to the radios macro as HTML.
func ContactFields(renderer render.Renderer, application Application, errors []FieldError) (map[string]any, error) {
	telephone := textInput("telephone", "Telephone number", application.Telephone, errors, map[string]any{
		"type":         "tel",
		"autocomplete": "tel",
		"classes":      "govuk-input--width-20",
	})
	conditional, err := renderer.Render("input", telephone)
	if err != nil {
		return nil, err
	}
	items := make([]any, 0, len(contactOptions))
	for _, option := range contactOptions {
		if option.Value == ContactByTelephone {
			items = append(items, map[string]any{
				"value":       option.Value,
				"text":        option.Text,
				"checked":     application.ContactBy == ContactByTelephone,
				"conditional": map[string]any{"html": conditional},
			})
			continue
		}
		items = append(items, map[string]any{
			"value":   option.Value,
			"text":    option.Text,
			"id":      "contact-by",
			"checked": application.ContactBy == option.Value,
		})
	}
	radios := map[string]any{
		"idPrefix": "contact-by",
		"name":     "contact-by",
		"fieldset": map[string]any{
			"legend": map[string]any{
				"text":          "How should we contact you?",
				"isPageHeading": true,
				"classes":       "govuk-fieldset__legend--l",
			},
		},
		"hint":  map[string]any{"text": "We will use this if we need to ask about your application"},
		"items": items,
	}
	addError(radios, errors, "contact-by")
	return map[string]any{"radios": radios}, nil
}

// RegionFields returns the checkboxes on the where-you-will-fish question.
func RegionFields(application Application, errors []FieldError) map[string]any {
	chosen := make(map[string]bool, len(application.Regions))
	for _, region := range application.Regions {
		chosen[region] = true
	}
	items := make([]any, 0, len(regions)+2)
	for index, region := range regions {
		item := map[string]any{
			"value":   region.Value,
			"text":    region.Text,
			"checked": chosen[region.Value],
		}
		if index == 0 {
			item["id"] = "regions"
		}
		items = append(items, item)
	}
	items = append(items, map[string]any{"divider": "or"})
	items = append(items, map[string]any{
		"value":     NotSure,
		"text":      "I have not decided yet",
		"behaviour": "exclusive",
		"checked":   chosen[NotSure],
	})
	checkboxes := map[string]any{
		"idPrefix": "where",
		"name":     "regions",
		"fieldset": map[string]any{
			"legend": map[string]any{
				"text":          "Where will you fish?",
				"isPageHeading": true,
				"classes":       "govuk-fieldset__legend--l",
			},
		},
		"hint":  map[string]any{"text": "Select all that apply"},
		"items": items,
	}
	addError(checkboxes, errors, "regions")
	return map[string]any{"checkboxes": checkboxes}
}

// LicenceFields returns the radios on the licence length question, including the example fee.
func LicenceFields(application Application, errors []FieldError) map[string]any {
	items := make([]any, 0, len(licenceLengths))
	for index, option := range licenceLengths {
		item := map[string]any{
			"value":   option.Value,
			"text":    option.Text + " (" + option.Fee + ")",
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
				"text":          "How long do you need a licence for?",
				"isPageHeading": true,
				"classes":       "govuk-fieldset__legend--l",
			},
		},
		"items": items,
	}
	addError(radios, errors, "licence-length")
	return map[string]any{"radios": radios}
}

// MonthField returns the select on the start month question.
func MonthField(application Application, errors []FieldError, now time.Time) map[string]any {
	months := StartMonths(now)
	items := make([]any, 0, len(months)+1)
	items = append(items, map[string]any{
		"value":    "",
		"text":     "Select a month",
		"selected": application.StartMonth == "",
	})
	for _, month := range months {
		items = append(items, map[string]any{
			"value":    month.Value,
			"text":     month.Text,
			"selected": application.StartMonth == month.Value,
		})
	}
	field := map[string]any{
		"id":   "start-month",
		"name": "start-month",
		"label": map[string]any{
			"text":          "When should the licence start?",
			"isPageHeading": true,
			"classes":       "govuk-label--l",
		},
		"items": items,
	}
	addError(field, errors, "start-month")
	return map[string]any{"select": field}
}

// AddressFields returns the fieldset and inputs on the address question.
func AddressFields(application Application, errors []FieldError) map[string]any {
	return map[string]any{
		"fieldset": map[string]any{
			"legend": map[string]any{
				"text":          "What is your address?",
				"isPageHeading": true,
				"classes":       "govuk-fieldset__legend--l",
			},
		},
		"line1": textInput("address-line-1", "Address line 1", application.AddressLine1, errors, map[string]any{
			"autocomplete": "address-line1",
		}),
		"line2": textInput("address-line-2", "Address line 2 (optional)", application.AddressLine2, errors, map[string]any{
			"autocomplete": "address-line2",
		}),
		"town": textInput("town", "Town or city", application.Town, errors, map[string]any{
			"autocomplete": "address-level2",
			"classes":      "govuk-input--width-20",
		}),
		"postcode": textInput("postcode", "Postcode", application.Postcode, errors, map[string]any{
			"autocomplete": "postal-code",
			"classes":      "govuk-input--width-10",
			"spellcheck":   false,
		}),
		"inset": map[string]any{
			"text": "This example asks you to type your address. It does not look up addresses from a postcode.",
		},
	}
}

// EvidenceField returns the file upload on the optional evidence question.
func EvidenceField(application Application, errors []FieldError) map[string]any {
	upload := map[string]any{
		"id":   "evidence",
		"name": "evidence",
		"label": map[string]any{
			"text":          "Upload evidence of a concession",
			"isPageHeading": true,
			"classes":       "govuk-label--l",
		},
		"hint": map[string]any{
			"text": "PDF, PNG, or JPG. You can skip this question if you do not have a concession.",
		},
	}
	addError(upload, errors, "evidence")
	return map[string]any{"currentFile": application.EvidenceFilename, "upload": upload}
}

// DetailsField returns the character count on the optional extra details question.
func DetailsField(application Application, errors []FieldError) map[string]any {
	details := map[string]any{
		"name":      "additional-details",
		"id":        "additional-details",
		"maxlength": 200,
		"threshold": 75,
		"value":     application.AdditionalDetails,
		"label": map[string]any{
			"text":          "Is there anything else we should know?",
			"isPageHeading": true,
			"classes":       "govuk-label--l",
		},
		"hint": map[string]any{
			"text": "You can skip this question. Do not include payment card numbers or passwords.",
		},
	}
	addError(details, errors, "additional-details")
	return map[string]any{"details": details}
}

// PasswordFields returns the two password inputs. Values are never returned to the page.
func PasswordFields(errors []FieldError) map[string]any {
	password := map[string]any{
		"id":           "password",
		"name":         "password",
		"autocomplete": "new-password",
		"label": map[string]any{
			"text":          "Create a password",
			"isPageHeading": true,
			"classes":       "govuk-label--l",
		},
		"hint": map[string]any{"text": "Must be at least 8 characters. This example does not store your password."},
	}
	addError(password, errors, "password")
	confirm := map[string]any{
		"id":           "password-confirm",
		"name":         "password-confirm",
		"autocomplete": "new-password",
		"label":        map[string]any{"text": "Confirm password"},
	}
	addError(confirm, errors, "password-confirm")
	return map[string]any{"password": password, "confirm": confirm}
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
	rows := make([]any, 0, len(licenceLengths))
	for _, option := range licenceLengths {
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
					"text": "You can apply if you are 13 or over and you will fish with a rod in England or Wales.",
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
					"html": `<h2 class="govuk-heading-l">Before you apply</h2><p class="govuk-body">You need your name, date of birth, email address, and home address.</p>`,
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
		"html":      "Your reference number<br><strong>" + htmlutil.Escape(reference) + "</strong>",
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
