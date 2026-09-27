package httpx_test

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Nooshu/govuk-frontend-example-go/internal/baseline"
	"github.com/Nooshu/govuk-frontend-example-go/internal/httpx"
)

func TestParseCookiesKeepsInvalidEscapes(t *testing.T) {
	if len(httpx.ParseCookies("")) != 0 {
		t.Fatal("empty header was not empty")
	}
	got := httpx.ParseCookies(" rod_session=abc%20def; broken; =novalue; theme=light%zz; extra=1")
	if got["rod_session"] != "abc def" {
		t.Fatalf("session = %q", got["rod_session"])
	}
	if got["theme"] != "light%zz" {
		t.Fatalf("invalid escape was rewritten: %q", got["theme"])
	}
	if _, ok := got["broken"]; ok {
		t.Fatal("a cookie without '=' was kept")
	}
	if got["extra"] != "1" {
		t.Fatalf("extra = %q", got["extra"])
	}
	if _, ok := got[""]; ok {
		t.Fatal("empty name was kept")
	}
}

func TestReadAndParseBody(t *testing.T) {
	t.Run("nil and empty", func(t *testing.T) {
		body, err := httpx.ReadBody(&http.Request{}, 10)
		if err != nil || body.Field("a") != "" || body.Values("a") != nil {
			t.Fatalf("nil body: %#v %v", body, err)
		}
		parsed, err := httpx.ParseBody("application/x-www-form-urlencoded", nil, 10)
		if err != nil || len(parsed.Fields) != 0 {
			t.Fatalf("empty body: %#v %v", parsed, err)
		}
	})

	t.Run("urlencoded", func(t *testing.T) {
		request, err := http.NewRequest(http.MethodPost, "/", strings.NewReader("a=1&a=2&b=x+y"))
		if err != nil {
			t.Fatal(err)
		}
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		body, err := httpx.ReadBody(request, 100)
		if err != nil {
			t.Fatal(err)
		}
		if body.Field("a") != "1" || strings.Join(body.Values("a"), ",") != "1,2" || body.Field("b") != "x y" {
			t.Fatalf("%#v", body.Fields)
		}
	})

	t.Run("errors", func(t *testing.T) {
		cases := []struct {
			contentType string
			raw         string
			max         int64
			status      int
		}{
			{"application/x-www-form-urlencoded", "a=%zz", 100, http.StatusBadRequest},
			{"text/plain", "a=1", 100, http.StatusUnsupportedMediaType},
			{"", "a=1", 100, http.StatusUnsupportedMediaType},
			{"application/x-www-form-urlencoded", "abcdef", 3, http.StatusRequestEntityTooLarge},
			{"multipart/form-data", "not-multipart", 100, http.StatusBadRequest},
		}
		for _, tc := range cases {
			_, err := httpx.ParseBody(tc.contentType, []byte(tc.raw), tc.max)
			var bodyErr *httpx.BodyError
			if !errors.As(err, &bodyErr) || bodyErr.Status != tc.status {
				t.Fatalf("%q %q: %v", tc.contentType, tc.raw, err)
			}
			if bodyErr.Error() == "" {
				t.Fatal("empty error message")
			}
		}
	})

	t.Run("reader failure and size limit", func(t *testing.T) {
		request, err := http.NewRequest(http.MethodPost, "/", errReader{})
		if err != nil {
			t.Fatal(err)
		}
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if _, err := httpx.ReadBody(request, 10); err == nil {
			t.Fatal("a failing body was accepted")
		}
		big, err := http.NewRequest(http.MethodPost, "/", strings.NewReader("abcdef"))
		if err != nil {
			t.Fatal(err)
		}
		big.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if _, err := httpx.ReadBody(big, 3); err == nil {
			t.Fatal("an oversized body was accepted")
		}
	})

	t.Run("multipart", func(t *testing.T) {
		raw := strings.Join([]string{
			"--b",
			`Content-Disposition: form-data; name="name"`,
			"",
			"Ada",
			"--b",
			`Content-Disposition: form-data; filename="ignored.txt"`,
			"",
			"no field name",
			"--b",
			`Content-Disposition: form-data; name="evidence"; filename="note.pdf"`,
			"",
			"pdf-bytes",
			"--b--",
			"",
		}, "\r\n")
		body, err := httpx.ParseBody("multipart/form-data; boundary=b", []byte(raw), 1000)
		if err != nil {
			t.Fatal(err)
		}
		if body.Field("name") != "Ada" {
			t.Fatalf("name = %q", body.Field("name"))
		}
		if body.Upload == nil || body.Upload.Filename != "note.pdf" || body.Upload.FieldName != "evidence" {
			t.Fatalf("upload = %#v", body.Upload)
		}
	})

	t.Run("multipart edges", func(t *testing.T) {
		nameless := strings.Join([]string{
			"--b",
			"Content-Disposition: form-data",
			"",
			"ignored",
			"--b",
			`Content-Disposition: form-data; name="name"`,
			"",
			"Ada",
			"--b--",
			"",
		}, "\r\n")
		body, err := httpx.ParseBody("multipart/form-data; boundary=b", []byte(nameless), 1000)
		if err != nil || body.Field("name") != "Ada" {
			t.Fatalf("nameless part: %#v %v", body, err)
		}

		truncated := "--b\r\nContent-Disposition: form-data; name=\"evidence\"; filename=\"a.pdf\"\r\n\r\nHELLO"
		if _, err := httpx.ParseBody("multipart/form-data; boundary=b", []byte(truncated), 1000); err == nil {
			t.Fatal("truncated upload was accepted")
		}
		truncatedName := "--b\r\nContent-Disposition: form-data; name=\"name\"\r\n\r\nHELLO"
		if _, err := httpx.ParseBody("multipart/form-data; boundary=b", []byte(truncatedName), 1000); err == nil {
			t.Fatal("truncated field was accepted")
		}
		truncatedNameless := "--b\r\nContent-Disposition: form-data\r\n\r\nHELLO"
		if _, err := httpx.ParseBody("multipart/form-data; boundary=b", []byte(truncatedNameless), 1000); err == nil {
			t.Fatal("truncated nameless part was accepted")
		}
		broken := "--b\r\nContent-Disposition: form-data; name=\"a\"\r\n\r\nok\r\n--b\r\nthis is not a header\r\n"
		if _, err := httpx.ParseBody("multipart/form-data; boundary=b", []byte(broken), 1000); err == nil {
			t.Fatal("broken next part was accepted")
		}
	})
}

type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("read failed") }
func (errReader) Close() error             { return nil }

func TestAssets(t *testing.T) {
	root := t.TempDir()
	govuk := filepath.Join(root, "govuk")
	assets := filepath.Join(govuk, "assets")
	fonts := filepath.Join(assets, "fonts")
	if err := os.MkdirAll(fonts, 0o755); err != nil {
		t.Fatal(err)
	}
	css := []byte(`font: url(/assets/fonts/light-aaaaaaaa-v2.woff2); font: url("/assets/fonts/light-aaaaaaaa-v2.woff2");`)
	stylesheet := filepath.Join(root, "application.css")
	if err := os.WriteFile(stylesheet, css, 0o644); err != nil {
		t.Fatal(err)
	}
	script := []byte("/* frontend */")
	if err := os.WriteFile(filepath.Join(govuk, "govuk-frontend.min.js"), script, 0o644); err != nil {
		t.Fatal(err)
	}
	font := []byte("woff2")
	if err := os.WriteFile(filepath.Join(fonts, "light-aaaaaaaa-v2.woff2"), font, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(assets, "plain.txt"), []byte("no"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(assets, "subdir"), 0o755); err != nil {
		t.Fatal(err)
	}

	if _, err := httpx.NewAssets(govuk, assets, filepath.Join(root, "missing.css")); err == nil {
		t.Fatal("missing stylesheet was accepted")
	}
	if _, err := httpx.NewAssets(filepath.Join(root, "empty"), assets, stylesheet); err == nil {
		t.Fatal("missing script was accepted")
	}

	loaded, err := httpx.NewAssets(govuk, assets, stylesheet)
	if err != nil {
		t.Fatal(err)
	}
	page := loaded.Page()
	if page.StylesheetHref == "" || page.AppModuleHref == "" || len(page.Preloads) != 1 {
		t.Fatalf("page assets = %#v", page)
	}
	if page.Preloads[0].As != "font" || page.Preloads[0].Type != "font/woff2" {
		t.Fatalf("preload = %#v", page.Preloads[0])
	}
	// The returned slice is a copy.
	page.Preloads[0].Href = "changed"
	if loaded.Page().Preloads[0].Href == "changed" {
		t.Fatal("Page shared its preload slice")
	}

	for _, path := range []string{page.StylesheetHref, loaded.Page().AppModuleHref} {
		asset, ok := loaded.Resolve(path)
		if !ok || len(asset.Body) == 0 || asset.Kind != baseline.KindFingerprintedAsset {
			t.Fatalf("resolve %s = %#v %v", path, asset, ok)
		}
	}
	scriptAsset, ok := loaded.Resolve(mustScriptHref(t, loaded))
	if !ok || string(scriptAsset.Body) != string(script) {
		t.Fatalf("script asset = %#v %v", scriptAsset, ok)
	}

	fontAsset, ok := loaded.Resolve("/assets/fonts/light-aaaaaaaa-v2.woff2")
	if !ok || fontAsset.Kind != baseline.KindFingerprintedAsset || fontAsset.ContentType != "font/woff2" || fontAsset.Body != nil {
		t.Fatalf("font = %#v %v", fontAsset, ok)
	}
	rootScript, ok := loaded.Resolve("/assets/govuk-frontend.min.js")
	if !ok || rootScript.Kind != baseline.KindStaticAsset || !strings.HasSuffix(rootScript.FilePath, "govuk-frontend.min.js") {
		t.Fatalf("root script = %#v %v", rootScript, ok)
	}

	rejected := []string{
		"/other/file.css",
		"/assets/",
		"/assets/plain.txt",
		"/assets/subdir",
		"/assets/missing.png",
		"/assets/../govuk-frontend.min.js",
		"/assets/fonts/light-aaaaaaaa-v2.woff2\x00.png",
	}
	for _, path := range rejected {
		if _, ok := loaded.Resolve(path); ok {
			t.Fatalf("resolved %q", path)
		}
	}

	emptyRoot, err := httpx.NewAssets(govuk, "", stylesheet)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := emptyRoot.Resolve("/assets//secret.png"); ok {
		t.Fatal("an empty asset root served an absolute path")
	}
}

func mustScriptHref(t *testing.T, assets *httpx.Assets) string {
	t.Helper()
	// The script URL is not on PageAssets; recover it by resolving the known prefix.
	// NewAssets publishes it at /assets/govuk-frontend.<fingerprint>.min.js.
	// Probe is unnecessary: read it back by scanning Resolve of the href embedded in the module.
	module, ok := assets.Resolve(assets.Page().AppModuleHref)
	if !ok {
		t.Fatal("app module missing")
	}
	const marker = "from '"
	text := string(module.Body)
	start := strings.Index(text, marker)
	if start < 0 {
		t.Fatalf("module = %s", text)
	}
	rest := text[start+len(marker):]
	end := strings.Index(rest, "'")
	return rest[:end]
}

// Keep io imported for readers that tests may wrap. errReader is sufficient today.
var _ io.Reader = errReader{}
