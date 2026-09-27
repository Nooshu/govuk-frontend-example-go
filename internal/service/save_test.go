package service_test

import (
	"testing"

	"github.com/Nooshu/govuk-frontend-example-go/internal/service"
)

func TestSavingAValidAnswerMarksItsQuestionComplete(t *testing.T) {
	t.Parallel()
	saved := service.SaveName(service.NewApplication(), " Ada ", " Lovelace ", true)
	if saved.FirstName != "Ada" || saved.LastName != "Lovelace" {
		t.Errorf("the name was not trimmed: %q %q", saved.FirstName, saved.LastName)
	}
	if !saved.IsCompleted(service.StepName) {
		t.Error("a valid name did not complete its question")
	}
}

func TestSavingAnInvalidAnswerKeepsItButReopensTheQuestion(t *testing.T) {
	t.Parallel()
	application := service.SaveName(service.NewApplication(), "Ada", "Lovelace", true)
	application = service.SaveName(application, "Ada", "", false)
	if application.FirstName != "Ada" {
		t.Error("what the applicant typed was discarded")
	}
	if application.IsCompleted(service.StepName) {
		t.Error("an invalid answer left the question marked complete")
	}
}

func TestSavingDoesNotChangeTheCallersApplication(t *testing.T) {
	t.Parallel()
	original := service.NewApplication()
	service.SaveName(original, "Ada", "Lovelace", true)
	if original.FirstName != "" || len(original.Completed) != 0 {
		t.Error("saving changed the application it was given")
	}
}

func TestSaveContactAndRegionsDropUnknownValues(t *testing.T) {
	t.Parallel()
	application := service.SaveContact(service.NewApplication(), "pigeon", " 01632 960 001 ", false)
	if application.ContactBy != "" {
		t.Errorf("an unknown contact method was stored as %q", application.ContactBy)
	}
	if application.Telephone != "01632 960 001" {
		t.Errorf("the telephone was not trimmed: %q", application.Telephone)
	}

	application = service.SaveRegions(application, []string{"wales", "atlantis", service.NotSure}, true)
	if len(application.Regions) != 2 {
		t.Errorf("stored regions = %v, want wales and not-sure", application.Regions)
	}

	application = service.SaveLicence(application, "forever", false)
	if application.LicenceLength != "" {
		t.Errorf("an unknown licence length was stored as %q", application.LicenceLength)
	}
	application = service.SaveMonth(application, "2026-04", true)
	if application.StartMonth != "2026-04" {
		t.Errorf("the start month was not stored: %q", application.StartMonth)
	}
}

func TestSaveAddressOnlyNormalisesAValidPostcode(t *testing.T) {
	t.Parallel()
	valid := service.SaveAddress(service.NewApplication(), service.Address{
		Line1:    " 1 Example Street ",
		Line2:    " Flat 2 ",
		Town:     " Exampleton ",
		Postcode: "sw1a1aa",
	}, true)
	if valid.Postcode != "SW1A 1AA" {
		t.Errorf("a valid postcode was stored as %q", valid.Postcode)
	}
	if valid.AddressLine1 != "1 Example Street" || valid.AddressLine2 != "Flat 2" || valid.Town != "Exampleton" {
		t.Errorf("the address lines were not trimmed: %+v", valid)
	}

	invalid := service.SaveAddress(service.NewApplication(), service.Address{Postcode: " NOPE "}, false)
	if invalid.Postcode != "NOPE" {
		t.Errorf("an invalid postcode was stored as %q, want it shown back as typed", invalid.Postcode)
	}
}

func TestSaveEvidenceKeepsThePreviousFileWhenNothingUsableIsUploaded(t *testing.T) {
	t.Parallel()
	application := service.SaveEvidence(service.NewApplication(), "licence.pdf", true, true)
	if application.EvidenceFilename != "licence.pdf" {
		t.Fatalf("the file name was not stored: %q", application.EvidenceFilename)
	}

	if got := service.SaveEvidence(application, "", false, true); got.EvidenceFilename != "licence.pdf" {
		t.Errorf("submitting without a file erased the name: %q", got.EvidenceFilename)
	}
	if got := service.SaveEvidence(application, "virus.exe", true, false); got.EvidenceFilename != "licence.pdf" {
		t.Errorf("a rejected file replaced the name: %q", got.EvidenceFilename)
	}
	if got := service.SaveEvidence(application, "", false, false); got.IsCompleted(service.StepEvidence) {
		t.Error("an invalid answer left the question complete")
	}
}

func TestSavePasswordRecordsOnlyThatOneWasSet(t *testing.T) {
	t.Parallel()
	saved := service.SavePassword(service.NewApplication(), true)
	if !saved.PasswordCreated || !saved.IsCompleted(service.StepCreateAPassword) {
		t.Error("a valid password was not recorded")
	}
	if rejected := service.SavePassword(saved, false); rejected.PasswordCreated {
		t.Error("a rejected password was recorded as set")
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
	application = service.SaveDetails(application, "Something", true)
	if application.AdditionalDetails != "Something" {
		t.Errorf("the details were not stored: %q", application.AdditionalDetails)
	}
}
