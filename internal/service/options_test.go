package service_test

import (
	"testing"

	"github.com/Nooshu/govuk-frontend-example-go/internal/service"
)

func TestLabelForFallsBackToTheStoredValue(t *testing.T) {
	t.Parallel()
	if got := service.LabelFor(service.Countries(), "Wales"); got != "Wales" {
		t.Errorf("LabelFor(Wales) = %q", got)
	}
	if got := service.LabelFor(service.Countries(), "Atlantis"); got != "Atlantis" {
		t.Errorf("an unknown value was labelled %q, want it shown unchanged", got)
	}
	if got := service.LabelFor(service.LicenceLengths(), "8-days"); got != "8 days" {
		t.Errorf("LabelFor(8-days) = %q", got)
	}
}

func TestTheOptionListsAreCopies(t *testing.T) {
	t.Parallel()
	countries := service.Countries()
	countries[0].Text = "changed"
	if service.Countries()[0].Text == "changed" {
		t.Error("callers can rewrite the country list")
	}
	lengths := service.LicenceLengths()
	lengths[0].Text = "changed"
	if service.LicenceLengths()[0].Text == "changed" {
		t.Error("callers can rewrite the licence lengths")
	}
	fees := service.LicenceFees()
	fees[0].Fee = "£0.00"
	if service.LicenceFees()[0].Fee == "£0.00" {
		t.Error("callers can rewrite the fees")
	}
	steps := service.Steps()
	steps[0].Heading = "changed"
	if service.Steps()[0].Heading == "changed" {
		t.Error("callers can rewrite the journey")
	}
}

func TestEveryLicenceLengthAndFeeIsComplete(t *testing.T) {
	t.Parallel()
	for _, length := range service.LicenceLengths() {
		if length.Text == "" || length.Value == "" {
			t.Errorf("licence length %+v is incomplete", length)
		}
	}
	for _, fee := range service.LicenceFees() {
		if fee.Fee == "" || fee.Text == "" {
			t.Errorf("licence fee %+v is incomplete", fee)
		}
	}
	if len(service.Countries()) != 3 {
		t.Errorf("got %d countries, want 3", len(service.Countries()))
	}
}
