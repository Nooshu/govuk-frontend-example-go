package service

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
	emailPattern    = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	phonePattern    = regexp.MustCompile(`^[0-9+() -]{8,20}$`)
	postcodePattern = regexp.MustCompile(`^[A-Z]{1,2}[0-9][A-Z0-9]? [0-9][A-Z]{2}$`)
	evidencePattern = regexp.MustCompile(`(?i)\.(pdf|png|jpe?g)$`)
	oneOrTwoDigits  = regexp.MustCompile(`^[0-9]{1,2}$`)
	fourDigits      = regexp.MustCompile(`^[0-9]{4}$`)
	filenamePattern = regexp.MustCompile(`^[\w. -]+$`)
)

// ValidateName checks the name question. An empty result means the answer can be saved.
func ValidateName(firstName, lastName string) []FieldError {
	var errors []FieldError
	errors = appendNamePart(errors, "first-name", "First name", "Enter your first name", firstName)
	errors = appendNamePart(errors, "last-name", "Last name", "Enter your last name", lastName)
	return errors
}

func appendNamePart(errors []FieldError, field, label, missing, value string) []FieldError {
	trimmed := Clean(value)
	switch {
	case trimmed == "":
		return append(errors, FieldError{Field: field, Href: "#" + field, Text: missing})
	case utf8.RuneCountInString(trimmed) > 100:
		return append(errors, FieldError{
			Field: field,
			Href:  "#" + field,
			Text:  label + " must be 100 characters or fewer",
		})
	default:
		return errors
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
		return fail("Date of birth must include a day, month and year")
	}
	if !oneOrTwoDigits.MatchString(day) || !oneOrTwoDigits.MatchString(month) || !fourDigits.MatchString(year) {
		return fail("Date of birth must be a real date")
	}
	dayNumber, _ := strconv.Atoi(day)
	monthNumber, _ := strconv.Atoi(month)
	yearNumber, _ := strconv.Atoi(year)
	date := time.Date(yearNumber, time.Month(monthNumber), dayNumber, 0, 0, 0, 0, time.UTC)
	if date.Year() != yearNumber || int(date.Month()) != monthNumber || date.Day() != dayNumber {
		return fail("Date of birth must be a real date")
	}
	utc := now.UTC()
	today := time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
	if date.After(today) {
		return fail("Date of birth must be in the past")
	}
	if ageOn(date, today) < 13 {
		return fail("You must be 13 or over to apply for a rod licence")
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

// ValidateContactPreference checks how the applicant wants to be contacted. A telephone number
// is required when they choose telephone.
func ValidateContactPreference(contactBy, telephone string) []FieldError {
	var errors []FieldError
	if contactBy != ContactByEmail && contactBy != ContactByTelephone {
		errors = append(errors, FieldError{
			Field: "contact-by",
			Href:  "#contact-by",
			Text:  "Select how we should contact you",
		})
	}
	trimmed := Clean(telephone)
	switch {
	case contactBy == ContactByTelephone && trimmed == "":
		errors = append(errors, FieldError{Field: "telephone", Href: "#telephone", Text: "Enter a telephone number"})
	case trimmed != "" && !phonePattern.MatchString(trimmed):
		errors = append(errors, FieldError{
			Field: "telephone",
			Href:  "#telephone",
			Text:  "Enter a telephone number, like 01632 960 001",
		})
	}
	return errors
}

// ValidateRegions checks where the applicant will fish. "Not sure" cannot be combined with a
// region, because the two answers contradict each other.
func ValidateRegions(selected []string) []FieldError {
	problem := func(text string) []FieldError {
		return []FieldError{{Field: "regions", Href: "#regions", Text: text}}
	}
	if len(selected) == 0 {
		return problem("Select where you will fish")
	}
	known := make(map[string]bool, len(regions))
	for _, region := range regions {
		known[region.Value] = true
	}
	var chosen []string
	exclusive := false
	for _, region := range selected {
		if region == NotSure {
			exclusive = true
			continue
		}
		chosen = append(chosen, region)
	}
	if exclusive && len(chosen) > 0 {
		return problem("Select where you will fish, or select that you have not decided yet")
	}
	for _, region := range chosen {
		if !known[region] {
			return problem("Select where you will fish")
		}
	}
	return nil
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
		Text:  "Select how long you need a licence for",
	}}
}

// ValidateStartMonth checks the month the licence should start against the months offered at now.
func ValidateStartMonth(value string, now time.Time) []FieldError {
	for _, month := range StartMonths(now) {
		if month.Value == value {
			return nil
		}
	}
	return []FieldError{{
		Field: "start-month",
		Href:  "#start-month",
		Text:  "Select when the licence should start",
	}}
}

// ValidateAddress checks the address question. Address line 2 is optional.
func ValidateAddress(line1, town, postcode string) []FieldError {
	var errors []FieldError
	trimmedLine1 := Clean(line1)
	switch {
	case trimmedLine1 == "":
		errors = append(errors, FieldError{Field: "address-line-1", Href: "#address-line-1", Text: "Enter address line 1"})
	case utf8.RuneCountInString(trimmedLine1) > 100:
		errors = append(errors, FieldError{
			Field: "address-line-1",
			Href:  "#address-line-1",
			Text:  "Address line 1 must be 100 characters or fewer",
		})
	}
	if Clean(town) == "" {
		errors = append(errors, FieldError{Field: "town", Href: "#town", Text: "Enter a town or city"})
	}
	normalised := NormalisePostcode(postcode)
	if normalised == "" || !postcodePattern.MatchString(normalised) {
		errors = append(errors, FieldError{Field: "postcode", Href: "#postcode", Text: "Enter a full UK postcode"})
	}
	return errors
}

// ValidateEvidence checks an optional evidence filename. An empty name is valid because the
// question can be skipped.
func ValidateEvidence(filename string) []FieldError {
	if filename == "" {
		return nil
	}
	if !evidencePattern.MatchString(filename) {
		return []FieldError{{
			Field: "evidence",
			Href:  "#evidence",
			Text:  "The selected file must be a PDF, PNG, or JPG",
		}}
	}
	return nil
}

// ValidateAdditionalDetails checks the optional extra details. Empty is valid.
func ValidateAdditionalDetails(value string) []FieldError {
	if utf8.RuneCountInString(value) > 200 {
		return []FieldError{{
			Field: "additional-details",
			Href:  "#additional-details",
			Text:  "Additional details must be 200 characters or fewer",
		}}
	}
	return nil
}

// ValidatePassword checks the password and its confirmation. Neither value is stored.
func ValidatePassword(password, confirm string) []FieldError {
	if utf8.RuneCountInString(password) < 8 {
		return []FieldError{{Field: "password", Href: "#password", Text: "Password must be at least 8 characters"}}
	}
	if password != confirm {
		return []FieldError{{
			Field: "password-confirm",
			Href:  "#password-confirm",
			Text:  "Enter the same password in both fields",
		}}
	}
	return nil
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

// NormalisePostcode returns an upper case postcode with a single space before the inward code.
//
// It returns an empty string for anything too short to be a postcode, which [ValidateAddress]
// reports as an error.
func NormalisePostcode(value string) string {
	compact := strings.ReplaceAll(strings.ToUpper(Clean(value)), " ", "")
	if len(compact) < 5 {
		return ""
	}
	return compact[:len(compact)-3] + " " + compact[len(compact)-3:]
}

// Clean trims surrounding whitespace from a posted field value.
func Clean(value string) string {
	return strings.TrimSpace(value)
}

// AsContactBy narrows a posted contact method, returning an empty string for anything else.
func AsContactBy(value string) string {
	if value == ContactByEmail || value == ContactByTelephone {
		return value
	}
	return ""
}

// AsLicenceLength narrows a posted licence length, returning an empty string for anything else.
func AsLicenceLength(value string) string {
	if value == LicenceOneDay || value == LicenceEightDay || value == LicenceTwelveMth {
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
