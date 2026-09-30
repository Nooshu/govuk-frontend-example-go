package htmlutil_test

import (
	"testing"

	"github.com/Nooshu/govuk-frontend-example-go/internal/htmlutil"
)

func TestEscape(t *testing.T) {
	got := htmlutil.Escape(`<>&"'`)
	want := "&lt;&gt;&amp;&quot;&#39;"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestPageTitle(t *testing.T) {
	cases := []struct {
		heading, service string
		errors           bool
		want             string
	}{
		{"Apply for a fishing rod licence", "Apply for a fishing rod licence", false, "Apply for a fishing rod licence – GOV.UK"},
		{"What is your full name?", "Apply for a fishing rod licence", false, "What is your full name? – Apply for a fishing rod licence – GOV.UK"},
		{"What is your full name?", "Apply for a fishing rod licence", true, "Error: What is your full name? – Apply for a fishing rod licence – GOV.UK"},
		{"Apply for a fishing rod licence", "Apply for a fishing rod licence", true, "Error: Apply for a fishing rod licence – GOV.UK"},
	}
	for _, tc := range cases {
		if got := htmlutil.PageTitle(tc.heading, tc.service, tc.errors); got != tc.want {
			t.Fatalf("PageTitle(%q,%q,%v)=%q want %q", tc.heading, tc.service, tc.errors, got, tc.want)
		}
	}
}

func TestSafeLocalPath(t *testing.T) {
	cases := map[string]string{
		"/name":           "/name",
		"/check-answers":  "/check-answers",
		"//evil":          "/",
		"https://x":       "/",
		"/x://y":          "/",
		`/path\trick`:     "/",
		"/path\r\ninject": "/",
		"relative":        "/",
		"":                "/",
	}
	for input, want := range cases {
		if got := htmlutil.SafeLocalPath(input); got != want {
			t.Fatalf("SafeLocalPath(%q)=%q want %q", input, got, want)
		}
	}
}
