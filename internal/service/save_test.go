package service_test

import (
	"testing"

	"github.com/Nooshu/govuk-frontend-example-go/internal/service"
)

func TestSavingAValidAnswerMarksItsQuestionComplete(t *testing.T) {
	t.Parallel()
	saved := service.SaveName(service.NewApplication(), " Ada Lovelace ", true)
	if saved.FullName != "Ada Lovelace" {
		t.Errorf("the name was not trimmed: %q", saved.FullName)
	}
	if !saved.IsCompleted(service.StepName) {
		t.Error("a valid name did not complete its question")
	}
}

func TestSavingAnInvalidAnswerKeepsItButReopensTheQuestion(t *testing.T) {
	t.Parallel()
	application := service.SaveName(service.NewApplication(), "Ada Lovelace", true)
	application = service.SaveName(application, "A", false)
	if application.FullName != "A" {
		t.Error("what the applicant typed was discarded")
	}
	if application.IsCompleted(service.StepName) {
		t.Error("an invalid answer left the question marked complete")
	}
}

func TestSavingDoesNotChangeTheCallersApplication(t *testing.T) {
	t.Parallel()
	original := service.NewApplication()
	service.SaveName(original, "Ada Lovelace", true)
	if original.FullName != "" || len(original.Completed) != 0 {
		t.Error("saving changed the application it was given")
	}
}

func TestSaveCountryAndLicenceDropUnknownValues(t *testing.T) {
	t.Parallel()
	application := service.SaveCountry(service.NewApplication(), " Wales ", true)
	if application.Country != "Wales" {
		t.Errorf("the country was not trimmed: %q", application.Country)
	}

	application = service.SaveLicence(application, "forever", false)
	if application.LicenceLength != "" {
		t.Errorf("an unknown licence length was stored as %q", application.LicenceLength)
	}
	application = service.SaveLicence(application, "8-days", true)
	if application.LicenceLength != "8-days" {
		t.Errorf("the licence length was not stored: %q", application.LicenceLength)
	}
}

func TestSaveDateAndEmailKeepWhatWasTyped(t *testing.T) {
	t.Parallel()
	application := service.SaveDate(service.NewApplication(), " 10 ", " 12 ", " 1990 ", true)
	if application.Day != "10" || application.Month != "12" || application.Year != "1990" {
		t.Errorf("the date was not trimmed: %+v", application)
	}
	application = service.SaveEmail(application, " ada@example.com ", true)
	if application.Email != "ada@example.com" {
		t.Errorf("the email was not trimmed: %q", application.Email)
	}
}
