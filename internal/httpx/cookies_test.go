package httpx_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Nooshu/govuk-frontend-example-go/internal/httpx"
)

func TestCookieValue(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Cookie", `rod_session=abc123; other=value; __Host-session=hostid`)

	if got := httpx.CookieValue(request, "rod_session"); got != "abc123" {
		t.Fatalf("session = %q", got)
	}
	if got := httpx.CookieValue(request, "__Host-session"); got != "hostid" {
		t.Fatalf("host session = %q", got)
	}
	if got := httpx.CookieValue(request, "missing"); got != "" {
		t.Fatalf("missing = %q", got)
	}
	if got := httpx.CookieValue(httptest.NewRequest(http.MethodGet, "/", nil), "rod_session"); got != "" {
		t.Fatalf("empty header = %q", got)
	}
}
