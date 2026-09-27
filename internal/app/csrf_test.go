package app

import (
	"testing"

	"github.com/Nooshu/govuk-frontend-example-go/internal/httpx"
	"github.com/Nooshu/govuk-frontend-example-go/internal/session"
)

func TestCSRFOK(t *testing.T) {
	current := session.New()
	if !csrfOK(current, &httpx.Body{Fields: map[string][]string{"csrf": {current.CSRF}}}) {
		t.Fatal("matching token was rejected")
	}
	if csrfOK(current, &httpx.Body{Fields: map[string][]string{"csrf": {"wrong"}}}) {
		t.Fatal("wrong token was accepted")
	}
	if csrfOK(current, &httpx.Body{Fields: map[string][]string{}}) {
		t.Fatal("missing token was accepted")
	}
	blank := session.New()
	blank.CSRF = ""
	if csrfOK(blank, &httpx.Body{Fields: map[string][]string{"csrf": {"anything"}}}) {
		t.Fatal("empty session CSRF was accepted")
	}
}
