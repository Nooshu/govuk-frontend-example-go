package service

// Presentation helpers for check-answers and confirmation.

import "strings"

// SummaryRows returns the check-your-answers rows, including the change link back to each
// question.
func SummaryRows(application Application) []map[string]any {
	return []map[string]any{
		row("Licence length", LabelFor(licenceLengths, application.LicenceLength), "/licence-length", "licence length"),
		row("Name", application.FullName, "/name", "name"),
		row("Date of birth", formatDateOfBirth(application), "/date-of-birth", "date of birth"),
		row("Where you will fish", application.Country, "/where-you-will-fish", "where you will fish"),
		row("Email address", application.Email, "/email", "email address"),
	}
}

func row(key, value, href, hidden string) map[string]any {
	shown := value
	if strings.TrimSpace(shown) == "" {
		shown = "Not provided"
	}
	return map[string]any{
		"key":   map[string]any{"text": key},
		"value": map[string]any{"text": shown},
		"actions": map[string]any{
			"items": []any{map[string]any{
				"href":               href + "?return=check-answers",
				"text":               "Change",
				"visuallyHiddenText": hidden,
			}},
		},
	}
}

func formatDateOfBirth(application Application) string {
	day := strings.TrimSpace(application.Day)
	month := strings.TrimSpace(application.Month)
	year := strings.TrimSpace(application.Year)
	if day == "" || month == "" || year == "" {
		return ""
	}
	return day + " " + month + " " + year
}
