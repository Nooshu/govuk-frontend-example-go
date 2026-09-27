package service

import (
	"fmt"
	"time"
)

// NotSure is the checkbox value for an applicant who has not decided where they will fish.
const NotSure = "not-sure"

// Option is a selectable value and the label shown for it.
type Option struct {
	Value string
	Text  string
}

// LicenceOption is a licence length and the example fee shown alongside it.
type LicenceOption struct {
	Value string
	Text  string
	Fee   string
}

var regions = []Option{
	{Value: "north-west", Text: "North West"},
	{Value: "north-east", Text: "North East"},
	{Value: "midlands", Text: "Midlands"},
	{Value: "south-west", Text: "South West"},
	{Value: "south-east", Text: "South East"},
	{Value: "wales", Text: "Wales"},
}

var licenceLengths = []LicenceOption{
	{Value: LicenceOneDay, Text: "1 day", Fee: "£7.10"},
	{Value: LicenceEightDay, Text: "8 days", Fee: "£14.20"},
	{Value: LicenceTwelveMth, Text: "12 months", Fee: "£36.80"},
}

var contactOptions = []Option{
	{Value: ContactByEmail, Text: "Email"},
	{Value: ContactByTelephone, Text: "Telephone"},
}

// Regions returns the regions the applicant can fish in.
func Regions() []Option { return append([]Option(nil), regions...) }

// LicenceLengths returns the licence lengths and their example fees.
func LicenceLengths() []LicenceOption { return append([]LicenceOption(nil), licenceLengths...) }

// ContactOptions returns the contact methods on the contact-preference question.
func ContactOptions() []Option { return append([]Option(nil), contactOptions...) }

// LicenceLengthOptions returns the licence lengths without their fees, so they can be labelled
// with [LabelFor].
func LicenceLengthOptions() []Option {
	options := make([]Option, 0, len(licenceLengths))
	for _, length := range licenceLengths {
		options = append(options, Option{Value: length.Value, Text: length.Text})
	}
	return options
}

// StartMonths returns the next 12 months the licence can start, counting from the month now
// falls in. Dates are UTC so the list does not shift with the server's time zone.
func StartMonths(now time.Time) []Option {
	utc := now.UTC()
	start := time.Date(utc.Year(), utc.Month(), 1, 0, 0, 0, 0, time.UTC)
	months := make([]Option, 0, 12)
	for index := range 12 {
		month := start.AddDate(0, index, 0)
		months = append(months, Option{
			Value: fmt.Sprintf("%04d-%02d", month.Year(), int(month.Month())),
			Text:  month.Format("January 2006"),
		})
	}
	return months
}

// LabelFor returns the label for a selected value.
//
// An unknown value is returned unchanged, so a stored answer is never silently blanked.
func LabelFor(options []Option, value string) string {
	for _, option := range options {
		if option.Value == value {
			return option.Text
		}
	}
	return value
}
