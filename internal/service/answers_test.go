package service_test

import (
	"strings"
	"testing"

	"github.com/Nooshu/govuk-frontend-example-go/internal/service"
)

func completed() service.Application {
	application := service.NewApplication()
	application = service.SaveName(application, "Ada", "Lovelace", true)
	application = service.SaveDate(application, "10", "12", "1990", true)
	application = service.SaveEmail(application, "ada@example.com", true)
	application = service.SaveContact(application, "telephone", "01632 960 001", true)
	application = service.SaveRegions(application, []string{"wales", "midlands"}, true)
	application = service.SaveLicence(application, "12-month", true)
	application = service.SaveMonth(application, "2026-04", true)
	application = service.SaveAddress(application, service.Address{
		Line1:    "1 Example Street",
		Town:     "Exampleton",
		Postcode: "sw1a1aa",
	}, true)
	application = service.SaveEvidence(application, "licence.pdf", true, true)
	application = service.SaveDetails(application, "Nothing else", true)
	return service.SavePassword(application, true)
}

// rowValue finds a summary row by its key and returns the value shown.
func rowValue(t *testing.T, rows []map[string]any, key string) string {
	t.Helper()
	for _, row := range rows {
		keys, _ := row["key"].(map[string]any)
		if keys["text"] != key {
			continue
		}
		values, _ := row["value"].(map[string]any)
		text, _ := values["text"].(string)
		return text
	}
	t.Fatalf("no summary row for %q", key)
	return ""
}

func TestSummaryRowsShowEveryAnswer(t *testing.T) {
	t.Parallel()
	rows := service.SummaryRows(completed(), now)

	want := map[string]string{
		"Name":                "Ada Lovelace",
		"Date of birth":       "10 December 1990",
		"Email address":       "ada@example.com",
		"Contact preference":  "Telephone",
		"Telephone number":    "01632 960 001",
		"Where you will fish": "Wales, Midlands",
		"Licence length":      "12 months",
		"Start month":         "April 2026",
		"Address":             "1 Example Street, Exampleton, SW1A 1AA",
		"Evidence":            "licence.pdf",
		"Additional details":  "Nothing else",
		"Password":            "Set",
	}
	for key, value := range want {
		if got := rowValue(t, rows, key); got != value {
			t.Errorf("%s = %q, want %q", key, got, value)
		}
	}
}

func TestSummaryRowsNeverShowThePassword(t *testing.T) {
	t.Parallel()
	rows := service.SummaryRows(completed(), now)
	for _, row := range rows {
		values, _ := row["value"].(map[string]any)
		if text, _ := values["text"].(string); strings.Contains(text, "correct horse") {
			t.Fatal("the summary shows a password")
		}
	}
	if got := rowValue(t, service.SummaryRows(service.NewApplication(), now), "Password"); got != "Not provided" {
		t.Errorf("an unset password shows %q, want Not provided", got)
	}
}

func TestSummaryRowsExplainMissingAnswers(t *testing.T) {
	t.Parallel()
	rows := service.SummaryRows(service.NewApplication(), now)
	for _, key := range []string{"Name", "Date of birth", "Address", "Evidence"} {
		if got := rowValue(t, rows, key); got != "Not provided" {
			t.Errorf("an unanswered %s shows %q, want Not provided", key, got)
		}
	}
}

func TestSummaryRowsHideAnUnrealDateOfBirth(t *testing.T) {
	t.Parallel()
	application := service.NewApplication()
	application.Day, application.Month, application.Year = "31", "2", "1990"
	if got := rowValue(t, service.SummaryRows(application, now), "Date of birth"); got != "Not provided" {
		t.Errorf("an impossible date shows %q, want Not provided", got)
	}
	application.Day, application.Month, application.Year = "10", "12", "not a year"
	if got := rowValue(t, service.SummaryRows(application, now), "Date of birth"); got != "Not provided" {
		t.Errorf("a non-numeric year shows %q, want Not provided", got)
	}
}

func TestSummaryRowsReportAnUndecidedApplicant(t *testing.T) {
	t.Parallel()
	application := service.NewApplication()
	application.Regions = []string{service.NotSure}
	if got := rowValue(t, service.SummaryRows(application, now), "Where you will fish"); got != "Not decided yet" {
		t.Errorf("an undecided applicant shows %q", got)
	}
}

func TestSummaryRowsLinkBackToEachQuestion(t *testing.T) {
	t.Parallel()
	for _, row := range service.SummaryRows(completed(), now) {
		actions, _ := row["actions"].(map[string]any)
		items, _ := actions["items"].([]any)
		if len(items) != 1 {
			t.Fatalf("a summary row has %d change links, want 1", len(items))
		}
		item, _ := items[0].(map[string]any)
		href, _ := item["href"].(string)
		if !strings.HasSuffix(href, "?return=check-answers") {
			t.Errorf("the change link %q does not return to the summary", href)
		}
		if item["visuallyHiddenText"] == "" {
			t.Error("a change link has no hidden text, so every link would read as \"Change\"")
		}
	}
}

func TestTaskSectionsTrackProgress(t *testing.T) {
	t.Parallel()
	sections := service.TaskSections(service.NewApplication())
	if len(sections) != 4 {
		t.Fatalf("got %d task sections, want 4", len(sections))
	}
	submit := sections[len(sections)-1].Items[0]
	if _, linked := submit["href"]; linked {
		t.Error("check-your-answers is linked before the questions are answered")
	}
	if !strings.Contains(statusText(t, submit), "Cannot start yet") {
		t.Errorf("the submit task reads %q", statusText(t, submit))
	}
	if got := statusText(t, sections[0].Items[0]); got != "Not started" {
		t.Errorf("an unanswered question reads %q, want Not started", got)
	}

	ready := service.TaskSections(completed())
	submit = ready[len(ready)-1].Items[0]
	if submit["href"] != "/check-answers" {
		t.Error("check-your-answers is not linked once every question is answered")
	}
	if got := statusText(t, submit); got != "Not started" {
		t.Errorf("the submit task reads %q, want Not started", got)
	}
	if got := statusText(t, ready[0].Items[0]); got != "Completed" {
		t.Errorf("an answered question reads %q, want Completed", got)
	}

	submittedApplication := completed()
	submittedApplication.Submitted = true
	done := service.TaskSections(submittedApplication)
	if got := statusText(t, done[len(done)-1].Items[0]); got != "Completed" {
		t.Errorf("a submitted application reads %q, want Completed", got)
	}
}

func statusText(t *testing.T, item map[string]any) string {
	t.Helper()
	status, _ := item["status"].(map[string]any)
	if text, ok := status["text"].(string); ok {
		return text
	}
	tag, _ := status["tag"].(map[string]any)
	text, _ := tag["text"].(string)
	return text
}
