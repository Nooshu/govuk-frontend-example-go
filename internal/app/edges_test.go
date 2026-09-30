package app

import (
	"crypto/tls"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/Nooshu/govuk-frontend-example-go/internal/baseline"
	"github.com/Nooshu/govuk-frontend-example-go/internal/components"
	"github.com/Nooshu/govuk-frontend-example-go/internal/config"
	"github.com/Nooshu/govuk-frontend-example-go/internal/govuk"
	"github.com/Nooshu/govuk-frontend-example-go/internal/httpx"
	"github.com/Nooshu/govuk-frontend-example-go/internal/pages"
	"github.com/Nooshu/govuk-frontend-example-go/internal/render"
	"github.com/Nooshu/govuk-frontend-example-go/internal/session"
)

func TestRequestIsSecure(t *testing.T) {
	https := httptest.NewRequest(http.MethodGet, "https://example.test/fees", nil)
	if !requestIsSecure(https) {
		t.Fatal("https scheme was not secure")
	}
	withTLS := httptest.NewRequest(http.MethodGet, "http://example.test/fees", nil)
	withTLS.TLS = &tls.ConnectionState{}
	if !requestIsSecure(withTLS) {
		t.Fatal("TLS was not secure")
	}
	plain := httptest.NewRequest(http.MethodGet, "http://example.test/fees", nil)
	if requestIsSecure(plain) {
		t.Fatal("plain HTTP was secure")
	}
	forwarded := httptest.NewRequest(http.MethodGet, "http://example.test/fees", nil)
	forwarded.Header.Set("X-Forwarded-Proto", " HTTPS , http")
	if !requestIsSecure(forwarded) {
		t.Fatal("forwarded https was not secure")
	}
	notForwarded := httptest.NewRequest(http.MethodGet, "http://example.test/fees", nil)
	notForwarded.Header.Set("X-Forwarded-Proto", "http")
	if requestIsSecure(notForwarded) {
		t.Fatal("forwarded http was secure")
	}
}

func TestDefaultsAndFailurePaths(t *testing.T) {
	application := newEdgeApp(t, func(string, map[string]any) (string, error) {
		return "<div></div>", nil
	})
	application.store = nil
	application.now = nil
	application.logger = nil
	// New already filled the defaults. Exercise them by building a second app.
	fresh, err := New(Options{
		Config:     application.config,
		Components: render.Func(func(string, map[string]any) (string, error) { return "<div></div>", nil }),
	})
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	fresh.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/health", nil))
	if recorder.Body.String() != "ok" {
		t.Fatalf("health = %q", recorder.Body.String())
	}

	t.Run("rendering failure falls back to plain text", func(t *testing.T) {
		failing, err := New(Options{
			Config: application.config,
			Components: render.Func(func(string, map[string]any) (string, error) {
				return "", errors.New("render failed")
			}),
			Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		})
		if err != nil {
			t.Fatal(err)
		}
		recorder := httptest.NewRecorder()
		failing.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/fees", nil))
		if recorder.Code != http.StatusInternalServerError || !strings.Contains(recorder.Body.String(), "problem with the service") {
			t.Fatalf("status %d body %q", recorder.Code, recorder.Body.String())
		}
		if strings.Contains(recorder.Header().Get("Content-Type"), "text/html") {
			t.Fatal("the plain-text fallback was sent as HTML")
		}
	})

	t.Run("a question render failure shows the problem page", func(t *testing.T) {
		failing, err := New(Options{
			Config: application.config,
			Components: render.Func(func(name string, _ map[string]any) (string, error) {
				if name == "input" {
					return "", errors.New("input failed")
				}
				return "<div></div>", nil
			}),
			Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		})
		if err != nil {
			t.Fatal(err)
		}
		recorder := httptest.NewRecorder()
		failing.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/name", nil))
		if recorder.Code != http.StatusInternalServerError || !strings.Contains(recorder.Body.String(), "<html") {
			t.Fatalf("status %d", recorder.Code)
		}
	})

	t.Run("header and cookie failures are logged", func(t *testing.T) {
		broken := newEdgeApp(t, func(string, map[string]any) (string, error) {
			return "<p>ok</p>", nil
		})
		broken.policy.CacheControl = map[string]string{}
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/fees", nil)
		broken.writePage(recorder, request, pages.View{Template: "fees", Heading: "Fees", Status: http.StatusOK}, session.New())
		if recorder.Code != http.StatusInternalServerError {
			t.Fatalf("writePage = %d", recorder.Code)
		}
		redirect := httptest.NewRecorder()
		broken.writeRedirect(redirect, request, "/fees", session.New())
		if redirect.Code != http.StatusSeeOther {
			t.Fatalf("writeRedirect = %d", redirect.Code)
		}
		raw := httptest.NewRecorder()
		broken.writeRaw(raw, request, rawResponse{status: http.StatusOK, body: []byte("x"), contentType: "text/plain", kind: baseline.KindDownload})
		if raw.Code != http.StatusInternalServerError {
			t.Fatalf("writeRaw = %d", raw.Code)
		}
		asset := httptest.NewRecorder()
		broken.writeAsset(asset, request, httpx.Asset{Body: []byte("css"), ContentType: "text/css", Kind: baseline.KindStaticAsset})
		if asset.Code != http.StatusInternalServerError {
			t.Fatalf("writeAsset = %d", asset.Code)
		}
	})

	t.Run("a missing asset file is not found", func(t *testing.T) {
		application := newEdgeApp(t, func(string, map[string]any) (string, error) { return "<div></div>", nil })
		recorder := httptest.NewRecorder()
		application.writeAsset(recorder, httptest.NewRequest(http.MethodGet, "/assets/missing.css", nil), httpx.Asset{
			FilePath:    filepath.Join(t.TempDir(), "missing.css"),
			ContentType: "text/css",
			Kind:        baseline.KindStaticAsset,
		})
		if recorder.Code != http.StatusNotFound {
			t.Fatalf("missing file = %d", recorder.Code)
		}
	})

	t.Run("writing the body can fail", func(t *testing.T) {
		application := newEdgeApp(t, func(string, map[string]any) (string, error) { return "<div></div>", nil })
		request := httptest.NewRequest(http.MethodGet, "/fees", nil)
		request.Header.Set("Accept-Encoding", "br")
		application.writeBody(&failingWriter{}, request, http.StatusOK, []byte("<p>hello hello hello</p>"), "text/html")
		head := httptest.NewRequest(http.MethodHead, "/fees", nil)
		recorder := httptest.NewRecorder()
		application.writeBody(recorder, head, http.StatusOK, []byte("<p>hello</p>"), "text/html")
		if recorder.Body.Len() != 0 {
			t.Fatal("HEAD wrote a body")
		}
	})

	t.Run("an empty outcome is a not-found page", func(t *testing.T) {
		application := newEdgeApp(t, func(string, map[string]any) (string, error) { return "<div></div>", nil })
		recorder := httptest.NewRecorder()
		application.write(recorder, httptest.NewRequest(http.MethodGet, "/missing", nil), outcome{}, session.New())
		if recorder.Code != http.StatusNotFound {
			t.Fatalf("empty outcome = %d", recorder.Code)
		}
	})

	t.Run("a bad content type is rejected", func(t *testing.T) {
		application := newEdgeApp(t, func(string, map[string]any) (string, error) { return "<div></div>", nil })
		recorder := httptest.NewRecorder()
		application.writeRaw(recorder, httptest.NewRequest(http.MethodGet, "/file", nil), rawResponse{
			status:      http.StatusOK,
			body:        []byte("x"),
			contentType: "text/plain\r\nX: 1",
			kind:        baseline.KindDownload,
		})
		if recorder.Code != http.StatusInternalServerError {
			t.Fatalf("bad content type = %d", recorder.Code)
		}
	})

	t.Run("a form handler failure shows the problem page", func(t *testing.T) {
		application := newEdgeApp(t, func(string, map[string]any) (string, error) { return "<div></div>", nil })
		current := application.store.Create()
		body := "csrf=" + current.CSRF
		request := httptest.NewRequest(http.MethodPost, "/custom", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		request.AddCookie(&http.Cookie{Name: "rod_session", Value: current.ID})
		handler := application.form(func(*session.Session, *httpx.Body, *http.Request) (outcome, error) {
			return outcome{}, errors.New("handler failed")
		})
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusInternalServerError {
			t.Fatalf("handler failure = %d", recorder.Code)
		}
	})

	t.Run("a fixture file that cannot be parsed is a problem page", func(t *testing.T) {
		application := newEdgeApp(t, func(string, map[string]any) (string, error) { return "<div></div>", nil })
		dir := t.TempDir()
		if err := os.Mkdir(filepath.Join(dir, "button"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "button", "fixtures.json"), []byte(`{"fixtures":`), 0o644); err != nil {
			t.Fatal(err)
		}
		cfg := *application.config
		cfg.ComponentsRoot = dir
		application.config = &cfg
		application.library = components.NewLibrary(dir)
		application.demosEnabled = true
		application.mux = application.routes()
		recorder := httptest.NewRecorder()
		application.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/components/button", nil))
		if recorder.Code != http.StatusInternalServerError {
			t.Fatalf("bad fixtures = %d", recorder.Code)
		}
	})

	t.Run("a fixture render failure is a problem page", func(t *testing.T) {
		application := newEdgeApp(t, func(string, map[string]any) (string, error) { return "<div></div>", nil })
		previous := renderComponentFixture
		renderComponentFixture = func(string, *govuk.Params) (string, error) {
			return "", errors.New("render failed")
		}
		t.Cleanup(func() { renderComponentFixture = previous })
		application.demosEnabled = true
		application.mux = application.routes()
		recorder := httptest.NewRecorder()
		application.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/components/button", nil))
		if recorder.Code != http.StatusInternalServerError {
			t.Fatalf("render failure = %d", recorder.Code)
		}
	})

	t.Run("a session cookie that cannot be built is logged", func(t *testing.T) {
		application := newEdgeApp(t, func(string, map[string]any) (string, error) { return "<div></div>", nil })
		application.policy.Cookie.SameSite = "Maybe"
		recorder := httptest.NewRecorder()
		application.setSessionCookie(recorder, httptest.NewRequest(http.MethodGet, "/fees", nil), session.New())
		if recorder.Header().Get("Set-Cookie") != "" {
			t.Fatal("an invalid cookie was set")
		}
	})

	t.Run("a body read that is not a BodyError shows the problem page", func(t *testing.T) {
		application := newEdgeApp(t, func(string, map[string]any) (string, error) { return "<div></div>", nil })
		previous := readBody
		readBody = func(*http.Request, int64) (*httpx.Body, error) {
			return nil, errors.New("read failed")
		}
		t.Cleanup(func() { readBody = previous })
		request := httptest.NewRequest(http.MethodPost, "/custom", strings.NewReader("csrf=x"))
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		handler := application.form(func(*session.Session, *httpx.Body, *http.Request) (outcome, error) {
			t.Fatal("handler ran after a failed body read")
			return outcome{}, nil
		})
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusInternalServerError {
			t.Fatalf("unexpected body error = %d", recorder.Code)
		}
	})

	t.Run("broken page templates fail startup", func(t *testing.T) {
		application := newEdgeApp(t, func(string, map[string]any) (string, error) { return "<div></div>", nil })
		_, err := New(Options{
			Config: application.config,
			Components: render.Func(func(string, map[string]any) (string, error) {
				return "<div></div>", nil
			}),
			TemplateFS: fstest.MapFS{
				"templates/broken.gohtml": &fstest.MapFile{Data: []byte("{{")},
			},
		})
		if err == nil {
			t.Fatal("broken page templates were accepted")
		}
	})
}

func newEdgeApp(t *testing.T, renderFn func(string, map[string]any) (string, error)) *App {
	t.Helper()
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(cfg.Stylesheet); err != nil {
		t.Skipf("run npm run build:styles: %v", err)
	}
	application, err := New(Options{
		Config:     cfg,
		Components: render.Func(renderFn),
		Logger:     slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	if err != nil {
		t.Fatal(err)
	}
	return application
}

type failingWriter struct{ header http.Header }

func (f *failingWriter) Header() http.Header {
	if f.header == nil {
		f.header = make(http.Header)
	}
	return f.header
}
func (f *failingWriter) WriteHeader(int) {}
func (f *failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}
