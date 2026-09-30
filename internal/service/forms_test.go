package service_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Nooshu/govuk-frontend-example-go/internal/service"
)

// option walks nested component options, which are plain maps.
func option(t *testing.T, params map[string]any, path ...string) any {
	t.Helper()
	var current any = params
	for _, key := range path {
		if current == nil {
			return nil
		}
		nested, ok := current.(map[string]any)
		if !ok {
			t.Fatalf("%s is not an options map, it is %T", strings.Join(path, "."), current)
		}
		current = nested[key]
	}
	return current
}

func TestErrorSummaryIsOnlyBuiltWhenThereAreErrors(t *testing.T) {
	t.Parallel()
	if _, ok := service.ErrorSummary(nil); ok {
		t.Error("an error summary was built for a page with no errors")
	}
	summary, ok := service.ErrorSummary([]service.FieldError{{Field: "email", Href: "#email", Text: "Enter an email address"}})
	if !ok {
		t.Fatal("no error summary was built")
	}
	if summary["titleText"] != "There is a problem" {
		t.Errorf("the summary title is %v", summary["titleText"])
	}
	list, _ := summary["errorList"].([]any)
	if len(list) != 1 {
		t.Fatalf("the summary lists %d errors, want 1", len(list))
	}
	item, _ := list[0].(map[string]any)
	if item["href"] != "#email" {
		t.Errorf("the summary link is %v, want #email", item["href"])
	}
}

func TestFieldsRetainTheAnswerAndCarryTheError(t *testing.T) {
	t.Parallel()
	application := service.SaveName(service.NewApplication(), "Ada Lovelace", true)
	errs := []service.FieldError{{Field: "full-name", Href: "#full-name", Text: "Enter your full name"}}

	params := service.NameField(application, errs)
	if got := option(t, params, "fullName", "value"); got != "Ada Lovelace" {
		t.Errorf("the name input holds %v, want Ada Lovelace", got)
	}
	if got := option(t, params, "fullName", "errorMessage", "text"); got != "Enter your full name" {
		t.Errorf("the name error is %v", got)
	}
}

func TestEachQuestionBuildsItsComponentOptions(t *testing.T) {
	t.Parallel()
	application := completed()

	if got := option(t, service.EmailField(application, nil), "email", "value"); got != "ada@example.com" {
		t.Errorf("the email input holds %v", got)
	}
	if got := option(t, service.EmailField(application, nil), "email", "hint", "text"); !strings.Contains(fmt.Sprint(got), "browser session") {
		t.Errorf("the email hint is %v", got)
	}
	if got := option(t, service.DateField(application, nil), "dateOfBirth", "namePrefix"); got != "date-of-birth" {
		t.Errorf("the date input prefix is %v", got)
	}
	if got := option(t, service.DateField(application, nil), "dateOfBirth", "hint", "text"); got != "For example, 31 3 1980" {
		t.Errorf("the date hint is %v", got)
	}
}

func TestQuestionsCarryTheirErrorMessages(t *testing.T) {
	t.Parallel()
	application := service.NewApplication()
	tests := map[string]struct {
		field string
		build func([]service.FieldError) map[string]any
		path  []string
	}{
		"date of birth": {"date-of-birth", func(errs []service.FieldError) map[string]any {
			return service.DateField(application, errs)
		}, []string{"dateOfBirth", "errorMessage", "text"}},
		"where you will fish": {"country", func(errs []service.FieldError) map[string]any {
			return service.CountryFields(application, errs)
		}, []string{"radios", "errorMessage", "text"}},
		"licence length": {"licence-length", func(errs []service.FieldError) map[string]any {
			return service.LicenceFields(application, errs)
		}, []string{"radios", "errorMessage", "text"}},
		"cookie choice": {"analytics", func(errs []service.FieldError) map[string]any {
			return service.CookieFields("", errs)
		}, []string{"radios", "errorMessage", "text"}},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			errs := []service.FieldError{{Field: test.field, Href: "#" + test.field, Text: "Something is wrong"}}
			if got := option(t, test.build(errs), test.path...); got != "Something is wrong" {
				t.Errorf("the message is %v, want \"Something is wrong\"", got)
			}
			if got := option(t, test.build(nil), test.path...); got != nil {
				t.Errorf("a page with no errors carries the message %v", got)
			}
		})
	}
}

func TestTheCountryQuestionOffersEnglandWalesScotland(t *testing.T) {
	t.Parallel()
	application := service.SaveCountry(service.NewApplication(), "Wales", true)
	items, _ := option(t, service.CountryFields(application, nil), "radios", "items").([]any)
	if len(items) != len(service.Countries()) {
		t.Fatalf("the question has %d items, want one per country", len(items))
	}
	second, _ := items[1].(map[string]any)
	if second["checked"] != true || second["value"] != "Wales" {
		t.Errorf("the applicant's choice was not retained: %v", second)
	}
	hint := option(t, service.CountryFields(application, nil), "radios", "hint", "text")
	if !strings.Contains(fmt.Sprint(hint), "fictional") {
		t.Errorf("the country hint is %v", hint)
	}
}

func TestTheLicenceQuestionDoesNotShowFees(t *testing.T) {
	t.Parallel()
	application := service.SaveLicence(service.NewApplication(), "12-months", true)
	items, _ := option(t, service.LicenceFields(application, nil), "radios", "items").([]any)
	if len(items) != len(service.LicenceLengths()) {
		t.Fatalf("the question has %d options", len(items))
	}
	first, _ := items[0].(map[string]any)
	if text, _ := first["text"].(string); strings.Contains(text, "£") {
		t.Errorf("the option reads %q, want no fee", text)
	}
	last, _ := items[len(items)-1].(map[string]any)
	if last["checked"] != true {
		t.Error("the applicant's choice was not retained")
	}
}

func TestTheCookieQuestionReflectsASavedChoice(t *testing.T) {
	t.Parallel()
	tests := map[string]bool{"accept": true, "reject": false}
	for choice, wantYes := range tests {
		items, _ := option(t, service.CookieFields(choice, nil), "radios", "items").([]any)
		yes, _ := items[0].(map[string]any)
		no, _ := items[1].(map[string]any)
		if yes["checked"] != wantYes || no["checked"] == wantYes {
			t.Errorf("with %q saved, yes=%v no=%v", choice, yes["checked"], no["checked"])
		}
	}
	items, _ := option(t, service.CookieFields("", nil), "radios", "items").([]any)
	for _, item := range items {
		if option, _ := item.(map[string]any); option["checked"] == true {
			t.Error("an option is selected before the applicant has chosen")
		}
	}
}

func TestContentPagesBuildTheirComponentOptions(t *testing.T) {
	t.Parallel()
	table := service.FeesTable()
	rows, _ := table["rows"].([]any)
	if len(rows) != len(service.LicenceFees()) {
		t.Errorf("the fees table has %d rows", len(rows))
	}
	if table["firstCellIsHeader"] != true {
		t.Error("the fees table does not mark its first column as headers")
	}

	if items, _ := service.HelpAccordion()["items"].([]any); len(items) != 3 {
		t.Errorf("the help accordion has %d sections", len(items))
	}
	if items, _ := service.GuidanceTabs()["items"].([]any); len(items) != 3 {
		t.Errorf("the guidance page has %d tabs", len(items))
	}
}

func TestTheConfirmationPanelEscapesTheReference(t *testing.T) {
	t.Parallel()
	panel := service.ConfirmationPanel(`FR<script>"'&`)
	html := fmt.Sprint(panel["html"])
	if !strings.Contains(html, "Your example reference number") {
		t.Errorf("the panel does not name the example reference: %q", html)
	}
	if strings.Contains(html, "<script>") {
		t.Errorf("the reference was not escaped: %q", html)
	}
	for _, want := range []string{"&lt;script&gt;", "&quot;", "&#39;", "&amp;"} {
		if !strings.Contains(html, want) {
			t.Errorf("the reference does not contain %q: %q", want, html)
		}
	}
}
