package httpx_test

import (
	"bytes"
	"compress/gzip"
	"io"
	"testing"

	"github.com/Nooshu/govuk-frontend-example-go/internal/httpx"
	"github.com/andybalholm/brotli"
)

func TestCompressible(t *testing.T) {
	cases := map[string]bool{
		"text/html; charset=utf-8": true,
		"application/javascript":   true,
		"application/json":         true,
		"image/svg+xml":            true,
		"image/png":                false,
		"font/woff2":               false,
	}
	for input, want := range cases {
		if got := httpx.Compressible(input); got != want {
			t.Fatalf("Compressible(%q)=%v want %v", input, got, want)
		}
	}
}

func TestCompressBrotliPreferred(t *testing.T) {
	body := bytes.Repeat([]byte("hello govuk "), 40)
	out, coding := httpx.Compress(body, "gzip, br", "text/html")
	if coding != "br" {
		t.Fatalf("coding %q", coding)
	}
	reader := brotli.NewReader(bytes.NewReader(out))
	got, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, body) {
		t.Fatal("brotli round-trip mismatch")
	}
}

func TestCompressGzipFallback(t *testing.T) {
	body := bytes.Repeat([]byte("hello govuk "), 40)
	out, coding := httpx.Compress(body, "gzip", "text/html")
	if coding != "gzip" {
		t.Fatalf("coding %q", coding)
	}
	reader, err := gzip.NewReader(bytes.NewReader(out))
	if err != nil {
		t.Fatal(err)
	}
	got, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, body) {
		t.Fatal("gzip round-trip mismatch")
	}
}

func TestCompressReadsQualityAndEmptyTokens(t *testing.T) {
	body := []byte("hello govuk hello govuk hello govuk hello govuk")
	out, coding := httpx.Compress(body, "gzip;q=0.5, , br;q=1.0", "text/html; charset=utf-8")
	if coding != "br" || len(out) == 0 {
		t.Fatalf("coding %q len %d", coding, len(out))
	}
	if _, coding := httpx.Compress(body, "identity", "application/json"); coding != "" {
		t.Fatalf("identity should leave the body plain, got %q", coding)
	}
}

func TestCompressSkipsEmptyAndUncompressible(t *testing.T) {
	if out, coding := httpx.Compress(nil, "br", "text/html"); coding != "" || out != nil {
		t.Fatalf("empty body: %q %#v", coding, out)
	}
	body := []byte("png-bytes")
	if out, coding := httpx.Compress(body, "br", "image/png"); coding != "" || !bytes.Equal(out, body) {
		t.Fatalf("png should stay raw: %q", coding)
	}
	if out, coding := httpx.Compress(body, "", "text/html"); coding != "" || !bytes.Equal(out, body) {
		t.Fatalf("no accept-encoding: %q", coding)
	}
}
