package service

// Server-side validation rules for the fishing rod licence journey.

import (
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// FieldError is one validation failure, shaped for both the error summary and the field itself.
type FieldError struct {
	// Field is the logical field name, used to find the message for a component.
	Field string
	// Href is the error summary link target, which must move focus to the first failing control.
	Href string
	// Text is the message shown to the applicant.
	Text string
}

var (
	emailPattern   = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	oneOrTwoDigits = regexp.MustCompile(`^[0-9]{1,2}$`)
	fourDigits     = regexp.MustCompile(`^[0-9]{4}$`)
	filenamePattern = regexp.MustCompile(`^[\w. -]+$`)
)

// ValidateName checks the full name question. An empty result means the answer can be saved.
func ValidateName(fullName string) []FieldError {
	trimmed := Clean(fullName)
	switch {
	case utf8.RuneCountInString(trimmed) < 2:
		return []FieldError{{Field: "full-name", Href: "#full-name", Text: "Enter your full name"}}
	case utf8.RuneCountInString(trimmed) > 100:
		return []FieldError{{
			Field: "full-name",
			Href:  "#full-name",
			Text:  "Full name must be 100 characters or fewer",
		}}
	default:
		return nil
	}
}

// ValidateDateOfBirth checks the date of birth. The applicant must be 13 or older on now, which
// is compared in UTC.
func ValidateDateOfBirth(day, month, year string, now time.Time) []FieldError {
	fail := func(text string) []FieldError {
		return []FieldError{{Field: "date-of-birth", Href: "#date-of-birth-day", Text: text}}
	}
	day, month, year = Clean(day), Clean(month), Clean(year)
	if day == "" || month == "" || year == "" {
		return fail("Enter your date of birth")
	}
	if !oneOrTwoDigits.MatchString(day) || !oneOrTwoDigits.MatchString(month) || !fourDigits.MatchString(year) {
		return fail("Enter a real date of birth")
	}
	dayNumber, _ := strconv.Atoi(day)
	monthNumber, _ := strconv.Atoi(month)
	yearNumber, _ := strconv.Atoi(year)
	date := time.Date(yearNumber, time.Month(monthNumber), dayNumber, 0, 0, 0, 0, time.UTC)
	if date.Year() != yearNumber || int(date.Month()) != monthNumber || date.Day() != dayNumber {
		return fail("Enter a real date of birth")
	}
	utc := now.UTC()
	today := time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
	if date.After(today) {
		return fail("Date of birth must be in the past")
	}
	if ageOn(date, today) < 13 {
		return fail("You must be at least 13 to use this example")
	}
	return nil
}

// ValidateEmail checks the email address question.
func ValidateEmail(email string) []FieldError {
	if !emailPattern.MatchString(Clean(email)) {
		return []FieldError{{
			Field: "email",
			Href:  "#email",
			Text:  "Enter an email address in the correct format, like name@example.com",
		}}
	}
	return nil
}

// ValidateCountry checks where the applicant will fish.
func ValidateCountry(country string) []FieldError {
	for _, option := range countries {
		if option.Value == country {
			return nil
		}
	}
	return []FieldError{{Field: "country", Href: "#country", Text: "Select where you will fish"}}
}

// ValidateLicenceLength checks the licence length question.
func ValidateLicenceLength(value string) []FieldError {
	for _, option := range licenceLengths {
		if option.Value == value {
			return nil
		}
	}
	return []FieldError{{
		Field: "licence-length",
		Href:  "#licence-length",
		Text:  "Select how long you need the licence for",
	}}
}

// ValidateCookieChoice checks the cookie settings answer.
func ValidateCookieChoice(value string) []FieldError {
	if value != "yes" && value != "no" {
		return []FieldError{{
			Field: "analytics",
			Href:  "#analytics",
			Text:  "Select yes if you want to accept analytics cookies",
		}}
	}
	return nil
}

// Clean trims surrounding whitespace from a posted field value.
func Clean(value string) string {
	return strings.TrimSpace(value)
}

// AsLicenceLength narrows a posted licence length, returning an empty string for anything else.
func AsLicenceLength(value string) string {
	if value == LicenceOneDay || value == LicenceEightDays || value == LicenceTwelveMths {
		return value
	}
	return ""
}

// SafeFilename returns the base name of an upload when it is safe to show back to the applicant.
//
// It reports false for path traversal, control characters, and names that are too long, so a
// filename from the request is never echoed unchecked.
func SafeFilename(filename string) (string, bool) {
	base := filename
	if index := strings.LastIndexAny(base, `/\`); index != -1 {
		base = base[index+1:]
	}
	if base == "" || base == "." || base == ".." {
		return "", false
	}
	if utf8.RuneCountInString(base) > 120 || !filenamePattern.MatchString(base) {
		return "", false
	}
	return base, true
}

func ageOn(dob, today time.Time) int {
	age := today.Year() - dob.Year()
	monthDelta := int(today.Month()) - int(dob.Month())
	if monthDelta < 0 || (monthDelta == 0 && today.Day() < dob.Day()) {
		age--
	}
	return age
}
