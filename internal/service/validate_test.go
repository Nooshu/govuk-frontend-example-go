package service_test

import (
	"strings"
	"testing"
	"time"

	"github.com/Nooshu/govuk-frontend-example-go/internal/service"
)

var now = time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC)

// fields returns the field names that failed, which is what a test usually cares about.
func fields(errors []service.FieldError) []string {
	names := make([]string, 0, len(errors))
	for _, err := range errors {
		names = append(names, err.Field)
	}
	return names
}

func equal(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for index := range got {
		if got[index] != want[index] {
			return false
		}
	}
	return true
}

func TestValidateName(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		first, last string
		want        []string
	}{
		"both given":    {"Ada", "Lovelace", nil},
		"trimmed":       {"  Ada  ", " Lovelace ", nil},
		"both missing":  {"", "   ", []string{"first-name", "last-name"}},
		"first missing": {"", "Lovelace", []string{"first-name"}},
		"last missing":  {"Ada", "", []string{"last-name"}},
		"first too long": {
			strings.Repeat("a", 101), "Lovelace", []string{"first-name"},
		},
		"last too long": {
			"Ada", strings.Repeat("a", 101), []string{"last-name"},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got := fields(service.ValidateName(test.first, test.last)); !equal(got, test.want) {
				t.Errorf("ValidateName(%q, %q) failed on %v, want %v", test.first, test.last, got, test.want)
			}
		})
	}
}

func TestValidateDateOfBirth(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		day, month, year string
		want             string
	}{
		"a real date":      {"10", "12", "1990", ""},
		"missing day":      {"", "12", "1990", "must include a day"},
		"missing month":    {"10", "", "1990", "must include a day"},
		"missing year":     {"10", "12", "", "must include a day"},
		"not numbers":      {"tenth", "12", "1990", "must be a real date"},
		"month too long":   {"10", "123", "1990", "must be a real date"},
		"two-digit year":   {"10", "12", "90", "must be a real date"},
		"31 February":      {"31", "2", "1990", "must be a real date"},
		"month 13":         {"10", "13", "1990", "must be a real date"},
		"in the future":    {"2", "3", "2026", "must be in the past"},
		"exactly 13 today": {"1", "3", "2013", ""},
		"a day under 13":   {"2", "3", "2013", "13 or over"},
		"under 13 by month": {
			"1", "4", "2013", "13 or over",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := service.ValidateDateOfBirth(test.day, test.month, test.year, now)
			if test.want == "" {
				if len(got) != 0 {
					t.Fatalf("want no errors, got %q", got[0].Text)
				}
				return
			}
			if len(got) != 1 || !strings.Contains(got[0].Text, test.want) {
				t.Fatalf("got %v, want a message containing %q", got, test.want)
			}
			if got[0].Href != "#date-of-birth-day" {
				t.Errorf("the summary link is %q, want #date-of-birth-day", got[0].Href)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	t.Parallel()
	for _, valid := range []string{"ada@example.com", " ada@example.com "} {
		if got := service.ValidateEmail(valid); len(got) != 0 {
			t.Errorf("ValidateEmail(%q) failed", valid)
		}
	}
	for _, invalid := range []string{"", "ada", "ada@example", "ada @example.com", "@example.com"} {
		if got := service.ValidateEmail(invalid); len(got) != 1 {
			t.Errorf("ValidateEmail(%q) was accepted", invalid)
		}
	}
}

func TestValidateContactPreference(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		contactBy, telephone string
		want                 []string
	}{
		"email":                 {"email", "", nil},
		"telephone with number": {"telephone", "01632 960 001", nil},
		"nothing chosen":        {"", "", []string{"contact-by"}},
		"unknown choice":        {"pigeon", "", []string{"contact-by"}},
		"telephone without a number": {
			"telephone", "  ", []string{"telephone"},
		},
		"a number that is not one": {"email", "not a number", []string{"telephone"}},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got := fields(service.ValidateContactPreference(test.contactBy, test.telephone)); !equal(got, test.want) {
				t.Errorf("got %v, want %v", got, test.want)
			}
		})
	}
}

func TestValidateRegions(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		regions []string
		want    string
	}{
		"one region":          {[]string{"wales"}, ""},
		"several regions":     {[]string{"wales", "midlands"}, ""},
		"not decided":         {[]string{service.NotSure}, ""},
		"nothing selected":    {nil, "Select where you will fish"},
		"not decided and one": {[]string{service.NotSure, "wales"}, "or select that you have not decided"},
		"unknown region":      {[]string{"atlantis"}, "Select where you will fish"},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := service.ValidateRegions(test.regions)
			if test.want == "" {
				if len(got) != 0 {
					t.Fatalf("want no errors, got %q", got[0].Text)
				}
				return
			}
			if len(got) != 1 || !strings.Contains(got[0].Text, test.want) {
				t.Fatalf("got %v, want a message containing %q", got, test.want)
			}
		})
	}
}

func TestValidateLicenceLengthAndStartMonth(t *testing.T) {
	t.Parallel()
	if got := service.ValidateLicenceLength("12-month"); len(got) != 0 {
		t.Error("a known licence length was rejected")
	}
	if got := service.ValidateLicenceLength("forever"); len(got) != 1 {
		t.Error("an unknown licence length was accepted")
	}
	if got := service.ValidateStartMonth("2026-03", now); len(got) != 0 {
		t.Error("the current month was rejected")
	}
	if got := service.ValidateStartMonth("2020-01", now); len(got) != 1 {
		t.Error("a month in the past was accepted")
	}
}

func TestValidateAddress(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		line1, town, postcode string
		want                  []string
	}{
		"complete":         {"1 Example Street", "Exampleton", "SW1A 1AA", nil},
		"lower case entry": {"1 Example Street", "Exampleton", "sw1a1aa", nil},
		"nothing":          {"", "", "", []string{"address-line-1", "town", "postcode"}},
		"line 1 too long": {
			strings.Repeat("a", 101), "Exampleton", "SW1A 1AA", []string{"address-line-1"},
		},
		"postcode too short": {"1 Example Street", "Exampleton", "SW1", []string{"postcode"}},
		"not a postcode":     {"1 Example Street", "Exampleton", "NOTAPOSTCODE", []string{"postcode"}},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got := fields(service.ValidateAddress(test.line1, test.town, test.postcode)); !equal(got, test.want) {
				t.Errorf("got %v, want %v", got, test.want)
			}
		})
	}
}

func TestValidateOptionalQuestions(t *testing.T) {
	t.Parallel()
	for _, accepted := range []string{"", "licence.pdf", "PHOTO.JPEG", "scan.jpg", "scan.png"} {
		if got := service.ValidateEvidence(accepted); len(got) != 0 {
			t.Errorf("ValidateEvidence(%q) failed", accepted)
		}
	}
	if got := service.ValidateEvidence("virus.exe"); len(got) != 1 {
		t.Error("an unsupported file type was accepted")
	}

	if got := service.ValidateAdditionalDetails(""); len(got) != 0 {
		t.Error("empty details were rejected")
	}
	if got := service.ValidateAdditionalDetails(strings.Repeat("é", 200)); len(got) != 0 {
		t.Error("200 characters were rejected; the limit counts characters, not bytes")
	}
	if got := service.ValidateAdditionalDetails(strings.Repeat("a", 201)); len(got) != 1 {
		t.Error("201 characters were accepted")
	}
}

func TestValidatePassword(t *testing.T) {
	t.Parallel()
	if got := service.ValidatePassword("correct horse", "correct horse"); len(got) != 0 {
		t.Error("a matching password was rejected")
	}
	if got := fields(service.ValidatePassword("short", "short")); !equal(got, []string{"password"}) {
		t.Errorf("a short password reported %v", got)
	}
	if got := fields(service.ValidatePassword("correct horse", "battery staple")); !equal(got, []string{"password-confirm"}) {
		t.Errorf("a mismatch reported %v", got)
	}
}

func TestValidateCookieChoice(t *testing.T) {
	t.Parallel()
	for _, valid := range []string{"yes", "no"} {
		if got := service.ValidateCookieChoice(valid); len(got) != 0 {
			t.Errorf("ValidateCookieChoice(%q) failed", valid)
		}
	}
	if got := service.ValidateCookieChoice("maybe"); len(got) != 1 {
		t.Error("an unknown cookie choice was accepted")
	}
}

func TestNormalisePostcode(t *testing.T) {
	t.Parallel()
	tests := map[string]string{
		"sw1a1aa":   "SW1A 1AA",
		" SW1A 1AA": "SW1A 1AA",
		"m11ae":     "M1 1AE",
		"M1":        "",
		"":          "",
	}
	for input, want := range tests {
		if got := service.NormalisePostcode(input); got != want {
			t.Errorf("NormalisePostcode(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestNarrowingPostedValues(t *testing.T) {
	t.Parallel()
	if service.AsContactBy("email") != "email" || service.AsContactBy("telephone") != "telephone" {
		t.Error("a known contact method was dropped")
	}
	if service.AsContactBy("pigeon") != "" {
		t.Error("an unknown contact method was kept")
	}
	for _, known := range []string{"1-day", "8-day", "12-month"} {
		if service.AsLicenceLength(known) != known {
			t.Errorf("licence length %q was dropped", known)
		}
	}
	if service.AsLicenceLength("forever") != "" {
		t.Error("an unknown licence length was kept")
	}
	if service.Clean("  spaced  ") != "spaced" {
		t.Error("Clean did not trim")
	}
}

func TestSafeFilename(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		want string
		ok   bool
	}{
		"licence.pdf":              {"licence.pdf", true},
		`C:\Users\ada\licence.pdf`: {"licence.pdf", true},
		"/tmp/licence.pdf":         {"licence.pdf", true},
		"../../etc/passwd":         {"passwd", true},
		"..":                       {"", false},
		".":                        {"", false},
		"":                         {"", false},
		"licence;rm -rf.pdf":       {"", false},
		"licence\n.pdf":            {"", false},
		strings.Repeat("a", 121):   {"", false},
	}
	for input, want := range tests {
		got, ok := service.SafeFilename(input)
		if ok != want.ok || got != want.want {
			t.Errorf("SafeFilename(%q) = %q, %t; want %q, %t", input, got, ok, want.want, want.ok)
		}
	}
}
