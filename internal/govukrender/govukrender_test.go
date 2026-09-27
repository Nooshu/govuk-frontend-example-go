package govukrender_test

import (
	"strings"
	"testing"

	"github.com/Nooshu/govuk-frontend-example-go/internal/govukrender"
)

func TestNewRendersButton(t *testing.T) {
	html, err := govukrender.New().Render("button", map[string]any{"text": "Continue"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, "govuk-button") || !strings.Contains(html, "Continue") {
		t.Fatalf("unexpected html: %s", html)
	}
}

func TestNewNestedParamsAndSlices(t *testing.T) {
	html, err := govukrender.New().Render("radios", map[string]any{
		"name": "contact",
		"items": []any{
			map[string]any{"value": "email", "text": "Email"},
			map[string]any{"value": "phone", "text": "Phone"},
		},
		"fieldset": map[string]any{
			"legend": map[string]any{"text": "How should we contact you?"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, "govuk-radios") {
		t.Fatalf("unexpected html: %s", html)
	}
}

func TestNewNilParams(t *testing.T) {
	html, err := govukrender.New().Render("button", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, "govuk-button") {
		t.Fatalf("unexpected html: %s", html)
	}
}

func TestNewMapSliceItems(t *testing.T) {
	html, err := govukrender.New().Render("summary-list", map[string]any{
		"rows": []map[string]any{
			{
				"key":   map[string]any{"text": "Name"},
				"value": map[string]any{"text": "Sam"},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, "govuk-summary-list") {
		t.Fatalf("unexpected html: %s", html)
	}
}
