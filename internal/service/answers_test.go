package service_test

import (
	"strings"
	"testing"

	"github.com/Nooshu/govuk-frontend-example-go/internal/service"
)

func completed() service.Application {
	application := service.NewApplication()
	application = service.SaveLicence(application, "12-months", true)
	application = service.SaveName(application, "Ada Lovelace", true)
	application = service.SaveDate(application, "31", "3", "1980", true)
	application = service.SaveCountry(application, "England", true)
	return service.SaveEmail(application, "ada@example.com", true)
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
	rows := service.SummaryRows(completed())

	want := map[string]string{
		"Licence length":      "12 months",
		"Name":                "Ada Lovelace",
		"Date of birth":       "31 3 1980",
		"Where you will fish": "England",
		"Email address":       "ada@example.com",
	}
	if len(rows) != len(want) {
		t.Fatalf("got %d summary rows, want %d", len(rows), len(want))
	}
	for key, value := range want {
		if got := rowValue(t, rows, key); got != value {
			t.Errorf("%s = %q, want %q", key, got, value)
		}
	}
}

func TestSummaryRowsExplainMissingAnswers(t *testing.T) {
	t.Parallel()
	rows := service.SummaryRows(service.NewApplication())
	for _, key := range []string{"Licence length", "Name", "Date of birth", "Where you will fish", "Email address"} {
		if got := rowValue(t, rows, key); got != "Not provided" {
			t.Errorf("an unanswered %s shows %q, want Not provided", key, got)
		}
	}
}

func TestSummaryRowsHideAnIncompleteDateOfBirth(t *testing.T) {
	t.Parallel()
	application := service.NewApplication()
	application.Day, application.Month = "31", "3"
	if got := rowValue(t, service.SummaryRows(application), "Date of birth"); got != "Not provided" {
		t.Errorf("an incomplete date shows %q, want Not provided", got)
	}
}

func TestSummaryRowsLinkBackToEachQuestion(t *testing.T) {
	t.Parallel()
	for _, row := range service.SummaryRows(completed()) {
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
