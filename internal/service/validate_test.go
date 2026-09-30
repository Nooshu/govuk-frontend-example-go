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
		name string
		want []string
	}{
		"given":     {"Ada Lovelace", nil},
		"trimmed":   {"  Ada Lovelace  ", nil},
		"missing":   {"", []string{"full-name"}},
		"too short": {"A", []string{"full-name"}},
		"too long":  {strings.Repeat("a", 101), []string{"full-name"}},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got := fields(service.ValidateName(test.name)); !equal(got, test.want) {
				t.Errorf("ValidateName(%q) failed on %v, want %v", test.name, got, test.want)
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
		"a real date":       {"10", "12", "1990", ""},
		"missing day":       {"", "12", "1990", "Enter your date of birth"},
		"missing month":     {"10", "", "1990", "Enter your date of birth"},
		"missing year":      {"10", "12", "", "Enter your date of birth"},
		"not numbers":       {"tenth", "12", "1990", "Enter a real date of birth"},
		"month too long":    {"10", "123", "1990", "Enter a real date of birth"},
		"two-digit year":    {"10", "12", "90", "Enter a real date of birth"},
		"31 February":       {"31", "2", "1990", "Enter a real date of birth"},
		"month 13":          {"10", "13", "1990", "Enter a real date of birth"},
		"in the future":     {"2", "3", "2026", "must be in the past"},
		"exactly 13 today":  {"1", "3", "2013", ""},
		"a day under 13":    {"2", "3", "2013", "at least 13"},
		"under 13 by month": {"1", "4", "2013", "at least 13"},
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

func TestValidateCountry(t *testing.T) {
	t.Parallel()
	for _, country := range []string{"England", "Wales", "Scotland"} {
		if got := service.ValidateCountry(country); len(got) != 0 {
			t.Errorf("ValidateCountry(%q) failed", country)
		}
	}
	if got := service.ValidateCountry(""); len(got) != 1 {
		t.Error("an empty country was accepted")
	}
	if got := service.ValidateCountry("atlantis"); len(got) != 1 {
		t.Error("an unknown country was accepted")
	}
}

func TestValidateLicenceLength(t *testing.T) {
	t.Parallel()
	if got := service.ValidateLicenceLength("12-months"); len(got) != 0 {
		t.Error("a known licence length was rejected")
	}
	if got := service.ValidateLicenceLength("forever"); len(got) != 1 {
		t.Error("an unknown licence length was accepted")
	}
	if got := service.ValidateLicenceLength("12-month"); len(got) != 1 {
		t.Error("the old 12-month value was accepted")
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

func TestNarrowingPostedValues(t *testing.T) {
	t.Parallel()
	for _, known := range []string{"1-day", "8-days", "12-months"} {
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
