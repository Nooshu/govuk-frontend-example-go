package baseline

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testPolicy(t *testing.T) *Policy {
	t.Helper()
	policy, err := Load(policyPath(t))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return policy
}

func policyPath(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		candidate := filepath.Join(dir, "baseline", "policy.json")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("baseline/policy.json not found")
		}
		dir = parent
	}
}

func TestLoadAndParse(t *testing.T) {
	policy := testPolicy(t)
	if policy.JSEnabledSnippet == "" || policy.JSEnabledScriptHash == "" {
		t.Fatal("policy is missing the js-enabled snippet or its hash")
	}
	if len(policy.Directives()) == 0 {
		t.Fatal("policy has no CSP directives")
	}
	first := policy.Directives()[0].Name
	policy.Directives()[0].Name = "mutated"
	if policy.Directives()[0].Name != first {
		t.Fatal("Directives must return a copy")
	}

	if _, err := Load(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Fatal("Load of a missing file succeeded")
	}

	cases := []struct {
		name string
		raw  string
	}{
		{"not json", "{"},
		{"no snippet", `{"jsEnabledScriptHash":"h","cacheControl":{"document":"no-cache"},"csp":{"directives":{"default-src":["'self'"]}}}`},
		{"no hash", `{"jsEnabledSnippet":"s","cacheControl":{"document":"no-cache"},"csp":{"directives":{"default-src":["'self'"]}}}`},
		{"no cache", `{"jsEnabledSnippet":"s","jsEnabledScriptHash":"h","csp":{"directives":{"default-src":["'self'"]}}}`},
		{"no directives", `{"jsEnabledSnippet":"s","jsEnabledScriptHash":"h","cacheControl":{"document":"no-cache"},"csp":{"directives":{}}}`},
		{"directives not an object", `{"jsEnabledSnippet":"s","jsEnabledScriptHash":"h","cacheControl":{"document":"no-cache"},"csp":{"directives":[]}}`},
		{"directive sources not arrays", `{"jsEnabledSnippet":"s","jsEnabledScriptHash":"h","cacheControl":{"document":"no-cache"},"csp":{"directives":{"default-src":"'self'"}}}`},
		{"truncated directives", `{"jsEnabledSnippet":"s","jsEnabledScriptHash":"h","cacheControl":{"document":"no-cache"},"csp":{"directives":{"default-src":`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Parse([]byte(tc.raw)); err == nil {
				t.Fatal("Parse succeeded")
			}
		})
	}
}

func TestObjectKeysRejectsNonObjects(t *testing.T) {
	if _, err := objectKeys([]byte(`[`)); err == nil {
		t.Fatal("an array was accepted as an object")
	}
	if _, err := objectKeys([]byte(``)); err == nil {
		t.Fatal("empty input was accepted")
	}
	if _, err := objectKeys([]byte(`{"a":`)); err == nil {
		t.Fatal("a truncated object was accepted")
	}
	if _, err := objectKeys([]byte(`{,}`)); err == nil {
		t.Fatal("a broken object key was accepted")
	}
	if _, err := objectKeys([]byte(`{true}`)); err == nil {
		t.Fatal("a non-string object key was accepted")
	}
}

type scriptedObject struct {
	tokens    []json.Token
	mores     []bool
	decodeErr error
	token     int
	more      int
}

func (s *scriptedObject) Token() (json.Token, error) {
	tok := s.tokens[s.token]
	s.token++
	return tok, nil
}

func (s *scriptedObject) More() bool {
	more := s.mores[s.more]
	s.more++
	return more
}

func (s *scriptedObject) Decode(any) error { return s.decodeErr }

func TestObjectKeysFromRejectsNonStringKey(t *testing.T) {
	_, err := objectKeysFrom(&scriptedObject{
		tokens: []json.Token{json.Delim('{'), true},
		mores:  []bool{true},
	})
	if err == nil || !strings.Contains(err.Error(), "expected an object key") {
		t.Fatalf("non-string key: %v", err)
	}
}

func TestDirectivesKeyOrderFailure(t *testing.T) {
	previous := keyOrder
	keyOrder = func([]byte) ([]string, error) { return nil, errors.New("order") }
	t.Cleanup(func() { keyOrder = previous })
	var parsed directives
	if err := parsed.UnmarshalJSON([]byte(`{"default-src":["'none'"]}`)); err == nil {
		t.Fatal("a key-order failure was ignored")
	}
}

func TestApplyHeaders(t *testing.T) {
	policy := testPolicy(t)

	t.Run("document", func(t *testing.T) {
		header := make(http.Header)
		header.Set("Server", "secret")
		err := policy.ApplyHeaders(header, HeaderOptions{
			Kind:            KindDocument,
			SecureTransport: true,
			SetsCookie:      true,
			Preload: []PreloadLink{{
				Href: "/assets/fonts/light.woff2",
				As:   "font",
				Type: "font/woff2",
			}},
		})
		if err != nil {
			t.Fatal(err)
		}
		if header.Get("Server") != "" {
			t.Fatal("Server header was not removed")
		}
		if header.Get("Cache-Control") != "private, no-cache" {
			t.Fatalf("cookie document cache = %q", header.Get("Cache-Control"))
		}
		if !strings.Contains(header.Get("Strict-Transport-Security"), "max-age=") {
			t.Fatalf("HSTS = %q", header.Get("Strict-Transport-Security"))
		}
		if !strings.Contains(header.Get("Content-Security-Policy"), policy.JSEnabledScriptHash) {
			t.Fatal("CSP is missing the js-enabled hash")
		}
		if !strings.Contains(header.Get("Link"), "crossorigin") {
			t.Fatalf("Link = %q", header.Get("Link"))
		}
		if header.Get("Vary") != "Accept-Encoding" {
			t.Fatalf("Vary = %q", header.Get("Vary"))
		}
	})

	t.Run("sensitive document omits HSTS on plain HTTP", func(t *testing.T) {
		header := make(http.Header)
		if err := policy.ApplyHeaders(header, HeaderOptions{Kind: KindSensitiveDocument}); err != nil {
			t.Fatal(err)
		}
		if header.Get("Strict-Transport-Security") != "" {
			t.Fatal("HSTS was sent on plain HTTP")
		}
		if header.Get("Cache-Control") == "" || header.Get("Content-Security-Policy") == "" {
			t.Fatal("a sensitive document is missing document headers")
		}
	})

	t.Run("fingerprinted asset", func(t *testing.T) {
		header := make(http.Header)
		if err := policy.ApplyHeaders(header, HeaderOptions{Kind: KindFingerprintedAsset}); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(header.Get("Cache-Control"), "immutable") {
			t.Fatalf("cache = %q", header.Get("Cache-Control"))
		}
		if header.Get("Content-Security-Policy") != "" {
			t.Fatal("an asset received a CSP")
		}
	})

	t.Run("rejections", func(t *testing.T) {
		if err := policy.ApplyHeaders(make(http.Header), HeaderOptions{Kind: "nope"}); err == nil {
			t.Fatal("unknown kind was accepted")
		}
		if err := policy.ApplyHeaders(make(http.Header), HeaderOptions{Kind: KindStaticAsset, SetsCookie: true}); err == nil {
			t.Fatal("Set-Cookie on an asset was accepted")
		}
		if err := policy.ApplyHeaders(make(http.Header), HeaderOptions{
			Kind:    KindDocument,
			Preload: []PreloadLink{{Href: "https://example.com/a.woff2", As: "font"}},
		}); err == nil {
			t.Fatal("an off-site preload was accepted")
		}
	})

	t.Run("content types", func(t *testing.T) {
		cases := []struct {
			in, want string
			ok       bool
		}{
			{"text/plain; charset=utf-8", "text/plain; charset=utf-8", true},
			{"text/plain", "text/plain; charset=utf-8", true},
			{"application/json", "application/json; charset=utf-8", true},
			{"image/png", "image/png", true},
			{"text/plain\r\nX: 1", "", false},
		}
		for _, tc := range cases {
			header := make(http.Header)
			err := policy.ApplyHeaders(header, HeaderOptions{Kind: KindDownload, ContentType: tc.in})
			if tc.ok != (err == nil) {
				t.Fatalf("content type %q: err=%v", tc.in, err)
			}
			if tc.ok && header.Get("Content-Type") != tc.want {
				t.Fatalf("content type %q = %q, want %q", tc.in, header.Get("Content-Type"), tc.want)
			}
		}
	})

	t.Run("empty content type is omitted", func(t *testing.T) {
		bare := *policy
		bare.ContentTypes = map[string]string{}
		header := make(http.Header)
		if err := bare.ApplyHeaders(header, HeaderOptions{Kind: KindStaticAsset}); err != nil {
			t.Fatal(err)
		}
		if header.Get("Content-Type") != "" {
			t.Fatalf("Content-Type = %q", header.Get("Content-Type"))
		}
	})

	t.Run("HSTS without includeSubDomains", func(t *testing.T) {
		bare := *policy
		bare.HSTS.IncludeSubDomains = false
		header := make(http.Header)
		if err := bare.ApplyHeaders(header, HeaderOptions{Kind: KindDocument, SecureTransport: true}); err != nil {
			t.Fatal(err)
		}
		if strings.Contains(header.Get("Strict-Transport-Security"), "includeSubDomains") {
			t.Fatalf("HSTS = %q", header.Get("Strict-Transport-Security"))
		}
	})
}

func TestPreloadLinkHeader(t *testing.T) {
	ok, err := PreloadLinkHeader([]PreloadLink{
		{Href: "/assets/app.css", As: "style", Type: "text/css"},
		{Href: "/assets/app.js", As: "script"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(ok, `type="text/css"`) || strings.Contains(ok, "crossorigin") {
		t.Fatalf("preload = %q", ok)
	}

	cases := []PreloadLink{
		{Href: "https://example.com/a.css", As: "style"},
		{Href: "//cdn.example/a.css", As: "style"},
		{Href: "/a.css", As: "video"},
		{Href: "/a.css", As: "style", Type: "not a type"},
		{Href: "/a b.css", As: "style"},
	}
	for _, link := range cases {
		if _, err := PreloadLinkHeader([]PreloadLink{link}); err == nil {
			t.Fatalf("accepted %#v", link)
		}
	}
	if _, err := PreloadLinkHeader(nil); err == nil {
		t.Fatal("empty preload list was accepted")
	}
}

func TestStrongETagAndDocumentKinds(t *testing.T) {
	if StrongETag([]byte("a")) != StrongETag([]byte("a")) {
		t.Fatal("ETag is not stable")
	}
	if !strings.HasPrefix(StrongETag([]byte("a")), `"`) {
		t.Fatal("ETag is not a strong quoted validator")
	}
	if IsDocument(KindDocument) != true || IsDocument(KindSensitiveDocument) != true || IsDocument(KindStaticAsset) != false {
		t.Fatal("IsDocument mismatch")
	}
}

func TestSetCookie(t *testing.T) {
	policy := testPolicy(t)
	secure := false
	httpOnly := false
	age := 60

	ok, err := policy.SetCookie("rod_session", "abc", CookieOptions{MaxAge: &age, Secure: &secure, HTTPOnly: &httpOnly})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"rod_session=abc", "Max-Age=60", "Path=/", "SameSite=Lax"} {
		if !strings.Contains(ok, want) {
			t.Fatalf("%q missing %q", ok, want)
		}
	}
	if strings.Contains(ok, "Secure") || strings.Contains(ok, "HttpOnly") {
		t.Fatalf("overrides were ignored: %q", ok)
	}

	host := true
	prefixed, err := policy.SetCookie("__Host-session", "abc", CookieOptions{Secure: &host, HostPrefix: true, SameSite: "Strict"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(prefixed, "Secure") || !strings.Contains(prefixed, "SameSite=Strict") {
		t.Fatalf("host cookie = %q", prefixed)
	}

	none := true
	if _, err := policy.SetCookie("analytics", "1", CookieOptions{SameSite: "None", Secure: &none, Path: "/"}); err != nil {
		t.Fatal(err)
	}

	bad := []struct {
		name, value string
		options     CookieOptions
	}{
		{"bad name", "a", CookieOptions{}},
		{"ok", "a b", CookieOptions{}},
		{"ok", "a", CookieOptions{SameSite: "Maybe"}},
		{"ok", "a", CookieOptions{SameSite: "None", Secure: &secure}},
		{"ok", "a", CookieOptions{Path: "relative"}},
		{"session", "a", CookieOptions{HostPrefix: true}},
		{"__Host-session", "a", CookieOptions{Secure: &secure}},
		{"__Host-session", "a", CookieOptions{Secure: &host, Path: "/app"}},
		{"ok", "a", CookieOptions{MaxAge: intPtr(-1)}},
	}
	for _, tc := range bad {
		if _, err := policy.SetCookie(tc.name, tc.value, tc.options); err == nil {
			t.Fatalf("accepted name=%q value=%q options=%#v", tc.name, tc.value, tc.options)
		}
	}
}

func TestContentSecurityPolicyEmptySources(t *testing.T) {
	policy := testPolicy(t)
	policy.CSP.Directives = append(policy.CSP.Directives, Directive{Name: "upgrade-insecure-requests"})
	csp := policy.ContentSecurityPolicy()
	if !strings.Contains(csp, "upgrade-insecure-requests;") && !strings.HasSuffix(csp, "upgrade-insecure-requests") {
		t.Fatalf("CSP = %q", csp)
	}
	if policy.PermissionsPolicyHeader() == "" {
		t.Fatal("permissions policy is empty")
	}
	policy.PermissionsPolicy = nil
	if policy.PermissionsPolicyHeader() != "" {
		t.Fatal("empty permissions policy was not empty")
	}
}

func intPtr(v int) *int { return &v }
