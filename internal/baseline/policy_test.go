package baseline_test

import (
	"net/http"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Nooshu/govuk-frontend-example-go/internal/baseline"
)

func policyPath(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	return filepath.Join(root, "baseline", "policy.json")
}

func TestLoadAndApplyDocumentHeaders(t *testing.T) {
	policy, err := baseline.Load(policyPath(t))
	if err != nil {
		t.Fatal(err)
	}
	if policy.JSEnabledSnippet == "" || policy.JSEnabledScriptHash == "" {
		t.Fatal("expected js-enabled snippet and hash")
	}

	header := http.Header{}
	if err := policy.ApplyHeaders(header, baseline.HeaderOptions{
		Kind:            baseline.KindDocument,
		SecureTransport: true,
		SetsCookie:      true,
	}); err != nil {
		t.Fatal(err)
	}
	if header.Get("Cache-Control") != "private, no-cache" {
		t.Fatalf("cache-control %q", header.Get("Cache-Control"))
	}
	if header.Get("Content-Security-Policy") == "" {
		t.Fatal("expected CSP")
	}
	if header.Get("Strict-Transport-Security") == "" {
		t.Fatal("expected HSTS on HTTPS")
	}
	for _, banned := range []string{"Server", "X-Powered-By"} {
		if header.Get(banned) != "" {
			t.Fatalf("banned header %s still set", banned)
		}
	}
}

func TestApplyHeadersRejectsUnknownKind(t *testing.T) {
	policy, err := baseline.Load(policyPath(t))
	if err != nil {
		t.Fatal(err)
	}
	if err := policy.ApplyHeaders(http.Header{}, baseline.HeaderOptions{Kind: "nope"}); err == nil {
		t.Fatal("expected error")
	}
}

func TestSetCookie(t *testing.T) {
	policy, err := baseline.Load(policyPath(t))
	if err != nil {
		t.Fatal(err)
	}
	secure := true
	maxAge := 3600
	value, err := policy.SetCookie("__Host-session", "abc", baseline.CookieOptions{
		Secure:     &secure,
		HostPrefix: true,
		MaxAge:     &maxAge,
		Path:       "/",
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, part := range []string{"__Host-session=abc", "Path=/", "Secure", "SameSite="} {
		if !strings.Contains(value, part) {
			t.Fatalf("missing %q in %q", part, value)
		}
	}

	if _, err := policy.SetCookie("bad name", "x", baseline.CookieOptions{}); err == nil {
		t.Fatal("expected invalid name error")
	}
}

func TestParseRejectsEmpty(t *testing.T) {
	if _, err := baseline.Parse([]byte(`{}`)); err == nil {
		t.Fatal("expected error")
	}
}

func TestIsDocument(t *testing.T) {
	if !baseline.IsDocument(baseline.KindDocument) || !baseline.IsDocument(baseline.KindSensitiveDocument) {
		t.Fatal("documents")
	}
	if baseline.IsDocument(baseline.KindFingerprintedAsset) {
		t.Fatal("assets are not documents")
	}
}
