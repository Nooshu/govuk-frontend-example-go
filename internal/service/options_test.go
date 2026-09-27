package service_test

import (
	"testing"
	"time"

	"github.com/Nooshu/govuk-frontend-example-go/internal/service"
)

func TestStartMonthsRunFromTheCurrentMonth(t *testing.T) {
	t.Parallel()
	months := service.StartMonths(now)
	if len(months) != 12 {
		t.Fatalf("got %d months, want 12", len(months))
	}
	if months[0].Value != "2026-03" || months[0].Text != "March 2026" {
		t.Errorf("the first month is %+v, want March 2026", months[0])
	}
	if last := months[11]; last.Value != "2027-02" || last.Text != "February 2027" {
		t.Errorf("the last month is %+v, want February 2027", last)
	}
}

func TestStartMonthsAreCalculatedInUTC(t *testing.T) {
	t.Parallel()
	// Just before midnight UTC on the last day of a month, a server running west of UTC would
	// otherwise offer the previous month.
	late := time.Date(2026, time.March, 31, 23, 30, 0, 0, time.FixedZone("UTC+2", 2*60*60))
	if got := service.StartMonths(late)[0].Value; got != "2026-03" {
		t.Errorf("the first month is %q, want 2026-03", got)
	}
}

func TestLabelForFallsBackToTheStoredValue(t *testing.T) {
	t.Parallel()
	if got := service.LabelFor(service.Regions(), "wales"); got != "Wales" {
		t.Errorf("LabelFor(wales) = %q", got)
	}
	if got := service.LabelFor(service.Regions(), "atlantis"); got != "atlantis" {
		t.Errorf("an unknown value was labelled %q, want it shown unchanged", got)
	}
	if got := service.LabelFor(service.ContactOptions(), "email"); got != "Email" {
		t.Errorf("LabelFor(email) = %q", got)
	}
	if got := service.LabelFor(service.LicenceLengthOptions(), "8-day"); got != "8 days" {
		t.Errorf("LabelFor(8-day) = %q", got)
	}
}

func TestTheOptionListsAreCopies(t *testing.T) {
	t.Parallel()
	regions := service.Regions()
	regions[0].Text = "changed"
	if service.Regions()[0].Text == "changed" {
		t.Error("callers can rewrite the region list")
	}
	lengths := service.LicenceLengths()
	lengths[0].Fee = "£0.00"
	if service.LicenceLengths()[0].Fee == "£0.00" {
		t.Error("callers can rewrite the fees")
	}
	contacts := service.ContactOptions()
	contacts[0].Text = "changed"
	if service.ContactOptions()[0].Text == "changed" {
		t.Error("callers can rewrite the contact options")
	}
	steps := service.Steps()
	steps[0].Heading = "changed"
	if service.Steps()[0].Heading == "changed" {
		t.Error("callers can rewrite the journey")
	}
}

func TestEveryLicenceLengthHasAFee(t *testing.T) {
	t.Parallel()
	for _, length := range service.LicenceLengths() {
		if length.Fee == "" || length.Text == "" || length.Value == "" {
			t.Errorf("licence length %+v is incomplete", length)
		}
	}
}
