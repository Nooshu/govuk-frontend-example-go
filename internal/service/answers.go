package service

// Presentation helpers for check-answers and confirmation.

import (
	"slices"
	"strconv"
	"strings"
	"time"
)

// SummaryRows returns the check-your-answers rows, including the change link back to each
// question.
//
// The password row says whether a password was set, never what it was.
func SummaryRows(application Application, now time.Time) []map[string]any {
	return []map[string]any{
		row("Name", joinName(application), "/name", "name"),
		row("Date of birth", formatDateOfBirth(application), "/date-of-birth", "date of birth"),
		row("Email address", application.Email, "/email", "email address"),
		row("Contact preference", LabelFor(contactOptions, application.ContactBy), "/contact-preference", "contact preference"),
		row("Telephone number", application.Telephone, "/contact-preference", "telephone number"),
		row("Where you will fish", formatRegions(application.Regions), "/where-you-will-fish", "where you will fish"),
		row("Licence length", LabelFor(LicenceLengthOptions(), application.LicenceLength), "/licence-length", "licence length"),
		row("Start month", LabelFor(StartMonths(now), application.StartMonth), "/start-month", "start month"),
		row("Address", formatAddress(application), "/address", "address"),
		row("Evidence", application.EvidenceFilename, "/evidence", "evidence"),
		row("Additional details", application.AdditionalDetails, "/additional-details", "additional details"),
		row("Password", passwordState(application), "/create-a-password", "password"),
	}
}

// TaskSection is one headed group of tasks on the task list.
type TaskSection struct {
	Heading  string
	IDPrefix string
	Items    []map[string]any
}

// TaskSections returns the task list for the licence journey.
//
// Statuses are "Completed", "Not started", or "Cannot start yet". Check-your-answers has no link
// until every required question is complete, so the applicant cannot skip ahead.
func TaskSections(application Application) []TaskSection {
	ready := RequiredStepsComplete(application)
	return []TaskSection{
		{
			Heading:  "Personal details",
			IDPrefix: "personal-details",
			Items: []map[string]any{
				task(application, StepName, "Your name"),
				task(application, StepDateOfBirth, "Date of birth"),
				task(application, StepEmail, "Email address"),
				task(application, StepContactPreference, "Contact preference"),
			},
		},
		{
			Heading:  "Your licence",
			IDPrefix: "your-licence",
			Items: []map[string]any{
				task(application, StepWhereYouWillFish, "Where you will fish"),
				task(application, StepLicenceLength, "Licence length"),
				task(application, StepStartMonth, "Start month"),
			},
		},
		{
			Heading:  "More about you",
			IDPrefix: "more-about-you",
			Items: []map[string]any{
				task(application, StepAddress, "Your address"),
				task(application, StepEvidence, "Concession evidence"),
				task(application, StepAdditionalDetails, "Additional details"),
				task(application, StepCreateAPassword, "Password"),
			},
		},
		{
			Heading:  "Apply",
			IDPrefix: "apply",
			Items:    []map[string]any{submitTask(application, ready)},
		},
	}
}

func submitTask(application Application, ready bool) map[string]any {
	item := map[string]any{"title": map[string]any{"text": "Check your answers and submit"}}
	switch {
	case !ready:
		item["status"] = notStartedTag("Cannot start yet")
	case application.Submitted:
		item["href"] = "/check-answers"
		item["status"] = map[string]any{"text": "Completed"}
	default:
		item["href"] = "/check-answers"
		item["status"] = notStartedTag("Not started")
	}
	return item
}

func task(application Application, id StepID, text string) map[string]any {
	status := notStartedTag("Not started")
	if application.IsCompleted(id) {
		status = map[string]any{"text": "Completed"}
	}
	return map[string]any{
		"title":  map[string]any{"text": text},
		"href":   "/" + string(id),
		"status": status,
	}
}

func notStartedTag(text string) map[string]any {
	return map[string]any{"tag": map[string]any{"text": text, "classes": "govuk-tag--grey"}}
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

func passwordState(application Application) string {
	if application.PasswordCreated {
		return "Set"
	}
	return ""
}

func joinName(application Application) string {
	return strings.TrimSpace(application.FirstName + " " + application.LastName)
}

func formatDateOfBirth(application Application) string {
	day, dayErr := strconv.Atoi(application.Day)
	month, monthErr := strconv.Atoi(application.Month)
	year, yearErr := strconv.Atoi(application.Year)
	if dayErr != nil || monthErr != nil || yearErr != nil || day == 0 || month == 0 || year == 0 {
		return ""
	}
	date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	if date.Day() != day || int(date.Month()) != month {
		return ""
	}
	return date.Format("2 January 2006")
}

func formatRegions(selected []string) string {
	if slices.Contains(selected, NotSure) {
		return "Not decided yet"
	}
	labels := make([]string, 0, len(selected))
	for _, region := range selected {
		labels = append(labels, LabelFor(regions, region))
	}
	return strings.Join(labels, ", ")
}

func formatAddress(application Application) string {
	parts := make([]string, 0, 4)
	for _, part := range []string{
		application.AddressLine1,
		application.AddressLine2,
		application.Town,
		application.Postcode,
	} {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return strings.Join(parts, ", ")
}
