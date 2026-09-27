package render_test

import (
	"errors"
	"testing"

	"github.com/Nooshu/govuk-frontend-example-go/internal/render"
)

func TestFuncRender(t *testing.T) {
	r := render.Func(func(name string, params map[string]any) (string, error) {
		if name != "button" {
			t.Fatalf("name %q", name)
		}
		return "ok", nil
	})
	got, err := r.Render("button", map[string]any{"text": "Save"})
	if err != nil || got != "ok" {
		t.Fatalf("got %q %v", got, err)
	}
}

func TestUnavailable(t *testing.T) {
	_, err := render.Unavailable().Render("button", nil)
	if !errors.Is(err, render.ErrUnavailable) {
		t.Fatalf("got %v", err)
	}
}
