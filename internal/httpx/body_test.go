package httpx_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/Nooshu/govuk-frontend-example-go/internal/httpx"
)

func TestParseCookies(t *testing.T) {
	got := httpx.ParseCookies(`rod_session=abc123; other=a%20b; =skip; lone; empty=`)
	if got["rod_session"] != "abc123" {
		t.Fatalf("session %q", got["rod_session"])
	}
	if got["other"] != "a b" {
		t.Fatalf("decoded %q", got["other"])
	}
	if _, ok := got[""]; ok {
		t.Fatal("empty name should be skipped")
	}
	if got["empty"] != "" {
		t.Fatalf("empty value %q", got["empty"])
	}
	if len(httpx.ParseCookies("")) != 0 {
		t.Fatal("empty header")
	}
}

func TestReadBodyURLEncoded(t *testing.T) {
	request, err := httpRequest("POST", "application/x-www-form-urlencoded", "name=Sam&csrf=token")
	if err != nil {
		t.Fatal(err)
	}
	body, err := httpx.ReadBody(request, 1_000_000)
	if err != nil {
		t.Fatal(err)
	}
	if body.Field("name") != "Sam" || body.Field("missing") != "" {
		t.Fatalf("%#v", body.Fields)
	}
	if got := body.Values("name"); len(got) != 1 || got[0] != "Sam" {
		t.Fatalf("%#v", got)
	}
}

func TestReadBodyTooLarge(t *testing.T) {
	request, err := httpRequest("POST", "application/x-www-form-urlencoded", "abcdef")
	if err != nil {
		t.Fatal(err)
	}
	_, err = httpx.ReadBody(request, 3)
	bodyErr, ok := err.(*httpx.BodyError)
	if !ok || bodyErr.Status != 413 {
		t.Fatalf("got %#v", err)
	}
}

func TestParseBodyUnsupported(t *testing.T) {
	_, err := httpx.ParseBody("text/plain", []byte("x"), 100)
	bodyErr, ok := err.(*httpx.BodyError)
	if !ok || bodyErr.Status != 415 {
		t.Fatalf("got %#v", err)
	}
}

func TestParseBodyEmpty(t *testing.T) {
	body, err := httpx.ParseBody("application/x-www-form-urlencoded", nil, 100)
	if err != nil || len(body.Fields) != 0 {
		t.Fatalf("%#v %v", body, err)
	}
}

func httpRequest(method, contentType, raw string) (*http.Request, error) {
	request, err := http.NewRequest(method, "http://example.test/", strings.NewReader(raw))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", contentType)
	return request, nil
}
