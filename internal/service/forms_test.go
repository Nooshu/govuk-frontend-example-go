package service_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/Nooshu/govuk-frontend-example-go/internal/render"
	"github.com/Nooshu/govuk-frontend-example-go/internal/service"
)

func stub() render.Renderer {
	return render.Func(func(name string, _ map[string]any) (string, error) {
		return "<" + name + ">", nil
	})
}

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
	application := service.SaveName(service.NewApplication(), "Ada", "Lovelace", true)
	errs := []service.FieldError{{Field: "last-name", Href: "#last-name", Text: "Enter your last name"}}

	params := service.NameFields(application, errs)
	if got := option(t, params, "firstName", "value"); got != "Ada" {
		t.Errorf("the first name input holds %v, want Ada", got)
	}
	if got := option(t, params, "firstName", "errorMessage"); got != nil {
		t.Error("a field with no error carries an error message")
	}
	if got := option(t, params, "lastName", "errorMessage", "text"); got != "Enter your last name" {
		t.Errorf("the last name error is %v", got)
	}
}

func TestEachQuestionBuildsItsComponentOptions(t *testing.T) {
	t.Parallel()
	application := completed()

	if got := option(t, service.EmailField(application, nil), "email", "value"); got != "ada@example.com" {
		t.Errorf("the email input holds %v", got)
	}
	if got := option(t, service.DateField(application, nil), "dateOfBirth", "namePrefix"); got != "date-of-birth" {
		t.Errorf("the date input prefix is %v", got)
	}
	if got := option(t, service.MonthField(application, nil, now), "select", "id"); got != "start-month" {
		t.Errorf("the select id is %v", got)
	}
	if got := option(t, service.AddressFields(application, nil), "postcode", "value"); got != "SW1A 1AA" {
		t.Errorf("the postcode input holds %v", got)
	}
	if got := option(t, service.EvidenceField(application, nil), "currentFile"); got != "licence.pdf" {
		t.Errorf("the evidence question shows %v as the current file", got)
	}
	if got := option(t, service.DetailsField(application, nil), "details", "maxlength"); got != 200 {
		t.Errorf("the character count limit is %v, want 200", got)
	}
	if got := option(t, service.PasswordFields(nil), "password", "autocomplete"); got != "new-password" {
		t.Errorf("the password input autocomplete is %v", got)
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
		"where you will fish": {"regions", func(errs []service.FieldError) map[string]any {
			return service.RegionFields(application, errs)
		}, []string{"checkboxes", "errorMessage", "text"}},
		"licence length": {"licence-length", func(errs []service.FieldError) map[string]any {
			return service.LicenceFields(application, errs)
		}, []string{"radios", "errorMessage", "text"}},
		"start month": {"start-month", func(errs []service.FieldError) map[string]any {
			return service.MonthField(application, errs, now)
		}, []string{"select", "errorMessage", "text"}},
		"evidence": {"evidence", func(errs []service.FieldError) map[string]any {
			return service.EvidenceField(application, errs)
		}, []string{"upload", "errorMessage", "text"}},
		"additional details": {"additional-details", func(errs []service.FieldError) map[string]any {
			return service.DetailsField(application, errs)
		}, []string{"details", "errorMessage", "text"}},
		"password": {"password", func(errs []service.FieldError) map[string]any {
			return service.PasswordFields(errs)
		}, []string{"password", "errorMessage", "text"}},
		"password confirmation": {"password-confirm", func(errs []service.FieldError) map[string]any {
			return service.PasswordFields(errs)
		}, []string{"confirm", "errorMessage", "text"}},
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

func TestTheContactQuestionRevealsTheTelephoneInput(t *testing.T) {
	t.Parallel()
	application := service.SaveContact(service.NewApplication(), "telephone", "01632 960 001", true)
	params, err := service.ContactFields(stub(), application, nil)
	if err != nil {
		t.Fatalf("ContactFields: %v", err)
	}
	items, _ := option(t, params, "radios", "items").([]any)
	if len(items) != 2 {
		t.Fatalf("the question has %d options, want 2", len(items))
	}
	telephone, _ := items[1].(map[string]any)
	conditional, _ := telephone["conditional"].(map[string]any)
	if conditional["html"] != "<input>" {
		t.Errorf("the revealed content is %v, want the rendered input", conditional["html"])
	}
	if telephone["checked"] != true {
		t.Error("the applicant's choice was not retained")
	}
}

func TestTheContactQuestionReportsARendererFailure(t *testing.T) {
	t.Parallel()
	broken := render.Func(func(string, map[string]any) (string, error) {
		return "", errors.New("no such component")
	})
	if _, err := service.ContactFields(broken, service.NewApplication(), nil); err == nil {
		t.Error("a renderer failure was swallowed")
	}
}

func TestTheRegionQuestionOffersAnExclusiveOption(t *testing.T) {
	t.Parallel()
	application := service.SaveRegions(service.NewApplication(), []string{service.NotSure}, true)
	items, _ := option(t, service.RegionFields(application, nil), "checkboxes", "items").([]any)
	if len(items) != len(service.Regions())+2 {
		t.Fatalf("the question has %d items, want one per region plus a divider and an opt-out", len(items))
	}
	last, _ := items[len(items)-1].(map[string]any)
	if last["behaviour"] != "exclusive" {
		t.Error("the opt-out is not exclusive, so it could be combined with a region")
	}
	if last["checked"] != true {
		t.Error("the applicant's choice was not retained")
	}
	divider, _ := items[len(items)-2].(map[string]any)
	if divider["divider"] != "or" {
		t.Errorf("the divider reads %v, want or", divider["divider"])
	}
}

func TestTheLicenceQuestionShowsTheFee(t *testing.T) {
	t.Parallel()
	application := service.SaveLicence(service.NewApplication(), "12-month", true)
	items, _ := option(t, service.LicenceFields(application, nil), "radios", "items").([]any)
	if len(items) != len(service.LicenceLengths()) {
		t.Fatalf("the question has %d options", len(items))
	}
	first, _ := items[0].(map[string]any)
	if text, _ := first["text"].(string); !strings.Contains(text, "£") {
		t.Errorf("the option reads %q, want it to include the fee", text)
	}
	last, _ := items[len(items)-1].(map[string]any)
	if last["checked"] != true {
		t.Error("the applicant's choice was not retained")
	}
}

func TestTheStartMonthQuestionOffersTwelveMonthsAndAPrompt(t *testing.T) {
	t.Parallel()
	items, _ := option(t, service.MonthField(service.NewApplication(), nil, now), "select", "items").([]any)
	if len(items) != 13 {
		t.Fatalf("the select has %d options, want 12 months and a prompt", len(items))
	}
	prompt, _ := items[0].(map[string]any)
	if prompt["value"] != "" || prompt["selected"] != true {
		t.Errorf("the prompt is %v", prompt)
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
	if len(rows) != len(service.LicenceLengths()) {
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
	panel := service.ConfirmationPanel(`RL<script>"'&`)
	html := fmt.Sprint(panel["html"])
	if strings.Contains(html, "<script>") {
		t.Errorf("the reference was not escaped: %q", html)
	}
	for _, want := range []string{"&lt;script&gt;", "&quot;", "&#39;", "&amp;"} {
		if !strings.Contains(html, want) {
			t.Errorf("the reference does not contain %q: %q", want, html)
		}
	}
}
