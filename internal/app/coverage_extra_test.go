package app_test

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/Nooshu/govuk-frontend-example-go/internal/app"
	"github.com/Nooshu/govuk-frontend-example-go/internal/components"
	"github.com/Nooshu/govuk-frontend-example-go/internal/config"
	"github.com/Nooshu/govuk-frontend-example-go/internal/render"
)

func TestEveryQuestionPageRenders(t *testing.T) {
	t.Parallel()
	c := newClient(t)
	for _, path := range []string{
		"/name", "/date-of-birth", "/email", "/contact-preference", "/where-you-will-fish",
		"/licence-length", "/start-month", "/address", "/evidence", "/additional-details", "/create-a-password",
	} {
		reply := c.get(path)
		if reply.Code != http.StatusOK {
			t.Errorf("GET %s = %d, want 200", path, reply.Code)
		}
	}
}

func TestCatalogueFixtureAndAssets(t *testing.T) {
	t.Parallel()
	c := newClient(t)

	missing := c.get("/components/not-a-component")
	if missing.Code != http.StatusNotFound {
		t.Errorf("unknown component = %d, want 404", missing.Code)
	}
	unknownFixture := c.get("/components/back-link?fixture=does-not-exist")
	if unknownFixture.Code != http.StatusNotFound {
		t.Errorf("unknown fixture = %d, want 404", unknownFixture.Code)
	}

	preview := c.get("/components/back-link")
	if preview.Code != http.StatusOK || !strings.Contains(preview.Body.String(), "Back link") {
		t.Fatalf("component preview = %d", preview.Code)
	}
	fragment := c.get("/components/back-link/fixture?fixture=default")
	if fragment.Code != http.StatusOK || !strings.Contains(fragment.Body.String(), "govuk-back-link") {
		t.Fatalf("fixture fragment = %d %q", fragment.Code, fragment.Body.String())
	}
	if missingFragment := c.get("/components/back-link/fixture?fixture=does-not-exist"); missingFragment.Code != http.StatusNotFound {
		t.Errorf("missing fragment = %d, want 404", missingFragment.Code)
	}

	page := c.get("/fees")
	href := regexp.MustCompile(`href="(/assets/application\.[^"]+\.css)"`).FindStringSubmatch(page.Body.String())
	if href == nil {
		t.Fatal("page has no fingerprinted stylesheet")
	}
	asset := c.get(href[1])
	if asset.Code != http.StatusOK || !strings.Contains(asset.Header().Get("Content-Type"), "text/css") {
		t.Fatalf("stylesheet = %d %q", asset.Code, asset.Header().Get("Content-Type"))
	}
	module := regexp.MustCompile(`src="(/assets/app\.[^"]+\.mjs)"`).FindStringSubmatch(page.Body.String())
	if module == nil {
		t.Fatal("page has no app module")
	}
	if got := c.get(module[1]); got.Code != http.StatusOK || !strings.Contains(got.Body.String(), "initAll") {
		t.Fatalf("app module = %d", got.Code)
	}
	font := c.get("/assets/fonts/light-94a07e06a1-v2.woff2")
	if font.Code != http.StatusOK || font.Header().Get("Content-Type") != "font/woff2" {
		t.Fatalf("font = %d %q", font.Code, font.Header().Get("Content-Type"))
	}
	if got := c.get("/assets/does-not-exist.png"); got.Code != http.StatusNotFound {
		t.Errorf("missing asset = %d, want 404", got.Code)
	}
	if got := c.get("/assets/fonts/"); got.Code != http.StatusNotFound {
		t.Errorf("font directory = %d, want 404", got.Code)
	}
}

func TestResponsesUseTheBaseline(t *testing.T) {
	t.Parallel()
	c := newClient(t)

	page := c.get("/fees")
	etag := page.Header().Get("ETag")
	if etag == "" {
		t.Fatal("public page has no ETag")
	}
	again := httptest.NewRequest(http.MethodGet, "/fees", nil)
	again.Header.Set("If-None-Match", etag)
	if got := c.do(again); got.Code != http.StatusNotModified {
		t.Errorf("matching ETag = %d, want 304", got.Code)
	}

	head := c.do(httptest.NewRequest(http.MethodHead, "/fees", nil))
	if head.Code != http.StatusOK || head.Body.Len() != 0 {
		t.Errorf("HEAD = %d, body %d bytes", head.Code, head.Body.Len())
	}

	br := httptest.NewRequest(http.MethodGet, "/fees", nil)
	br.Header.Set("Accept-Encoding", "gzip, br")
	if got := c.do(br); got.Header().Get("Content-Encoding") != "br" {
		t.Errorf("encoding = %q, want br", got.Header().Get("Content-Encoding"))
	}
	gzip := httptest.NewRequest(http.MethodGet, "/fees", nil)
	gzip.Header.Set("Accept-Encoding", "gzip")
	if got := c.do(gzip); got.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("encoding = %q, want gzip", got.Header().Get("Content-Encoding"))
	}

	secure := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/fees", nil)
	secure.Header.Set("X-Forwarded-Proto", "https")
	reply := c.do(secure)
	if !strings.Contains(reply.Header().Get("Set-Cookie"), "__Host-session") || reply.Header().Get("Strict-Transport-Security") == "" {
		t.Errorf("https proxy cookie = %q hsts = %q", reply.Header().Get("Set-Cookie"), reply.Header().Get("Strict-Transport-Security"))
	}

	withTLS := httptest.NewRequest(http.MethodGet, "https://127.0.0.1/about", nil)
	withTLS.TLS = &tls.ConnectionState{}
	if got := c.do(withTLS); !strings.Contains(got.Header().Get("Set-Cookie"), "Secure") {
		t.Errorf("tls cookie = %q", got.Header().Get("Set-Cookie"))
	}
}

func TestSessionCookieLookup(t *testing.T) {
	t.Parallel()
	c := newClient(t)
	c.post("/name", url.Values{"first-name": {"Ada"}, "last-name": {"Lovelace"}})
	id := c.cookies["rod_session"]
	if id == "" {
		t.Fatal("no session cookie")
	}

	kept := httptest.NewRequest(http.MethodGet, "/name", nil)
	kept.AddCookie(&http.Cookie{Name: "__Host-session", Value: id})
	if !strings.Contains(c.do(kept).Body.String(), "Ada") {
		t.Fatal("the __Host- cookie did not find the session")
	}

	fresh := httptest.NewRequest(http.MethodGet, "/name", nil)
	fresh.AddCookie(&http.Cookie{Name: "rod_session", Value: "not-a-session"})
	recorder := httptest.NewRecorder()
	c.handler.ServeHTTP(recorder, fresh)
	if strings.Contains(recorder.Body.String(), "Ada") {
		t.Fatal("an unknown session cookie reused another applicant's answers")
	}
}

func TestRejectedPosts(t *testing.T) {
	t.Parallel()
	c := newClient(t)

	if location := c.post("/check-answers", url.Values{}).Header().Get("Location"); location != "/name" {
		t.Errorf("incomplete submit redirected to %q, want /name", location)
	}

	oversized := httptest.NewRequest(http.MethodPost, "/name", strings.NewReader(strings.Repeat("a", config.MaxBodyBytes+1)))
	oversized.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if got := c.do(oversized); got.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("oversized post = %d, want 413", got.Code)
	}

	plain := httptest.NewRequest(http.MethodPost, "/name", strings.NewReader("first-name=Ada"))
	plain.Header.Set("Content-Type", "text/plain")
	if got := c.do(plain); got.Code != http.StatusUnsupportedMediaType {
		t.Errorf("plain post = %d, want 415", got.Code)
	}

	broken := httptest.NewRequest(http.MethodPost, "/name", strings.NewReader("first-name=%zz"))
	broken.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if got := c.do(broken); got.Code != http.StatusBadRequest {
		t.Errorf("malformed post = %d, want 400", got.Code)
	}

	if location := c.upload("/evidence", "other", "note.pdf", "bytes").Header().Get("Location"); location != "/additional-details" {
		t.Errorf("a file on the wrong field redirected to %q", location)
	}

	reject := c.post("/cookie-choices", url.Values{"cookies": {"reject"}, "returnPath": {"/help"}})
	if reject.Header().Get("Location") != "/help" {
		t.Errorf("reject redirected to %q", reject.Header().Get("Location"))
	}
}

func TestDemosCanBeTurnedOff(t *testing.T) {
	t.Parallel()
	c := newClient(t, func(options *app.Options) { options.DemosEnabled = false })
	if got := c.get("/components"); got.Code != http.StatusNotFound {
		t.Errorf("catalogue with demos off = %d, want 404", got.Code)
	}
	start := c.get("/")
	if start.Code != http.StatusOK {
		t.Errorf("start page = %d, want 200", start.Code)
	}
	if strings.Contains(start.Body.String(), `href="/components"`) {
		t.Error("start page still links to the component catalogue when demos are off")
	}
}

func TestStartPageLinksToComponentCatalogueWhenDemosAreOn(t *testing.T) {
	t.Parallel()
	c := newClient(t)

	start := c.get("/")
	if start.Code != http.StatusOK {
		t.Fatalf("start page = %d, want 200", start.Code)
	}
	body := start.Body.String()
	if !strings.Contains(body, `href="/components"`) || !strings.Contains(body, "Preview GOV.UK components") {
		t.Error("start page does not link to the component preview homepage")
	}

	welsh := c.get("/cy")
	if !strings.Contains(welsh.Body.String(), `href="/components"`) {
		t.Error("Welsh start page does not link to the component preview homepage")
	}

	catalogue := c.get("/components")
	if catalogue.Code != http.StatusOK {
		t.Fatalf("catalogue = %d, want 200", catalogue.Code)
	}
	catalogueBody := catalogue.Body.String()
	if !strings.Contains(catalogueBody, "preview homepage") {
		t.Error("catalogue does not describe itself as the preview homepage")
	}
	if !strings.Contains(catalogueBody, `href="/components/button"`) {
		t.Error("catalogue is missing a per-component page link")
	}
}

func TestNewRejectsABrokenSetup(t *testing.T) {
	if _, err := app.New(app.Options{}); err == nil {
		t.Fatal("missing config was accepted")
	}
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.New(app.Options{Config: cfg}); err == nil {
		t.Fatal("missing renderer was accepted")
	}
	broken := *cfg
	broken.PolicyFile = filepath.Join(t.TempDir(), "missing.json")
	if _, err := app.New(app.Options{Config: &broken, Components: stubComponents()}); err == nil {
		t.Fatal("missing policy was accepted")
	}
	broken = *cfg
	broken.Stylesheet = filepath.Join(t.TempDir(), "missing.css")
	if _, err := app.New(app.Options{Config: &broken, Components: stubComponents()}); err == nil {
		t.Fatal("missing stylesheet was accepted")
	}
}

func TestAMatchingRendererReportsFixtureParity(t *testing.T) {
	t.Parallel()
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	fixture, ok := components.NewLibrary(cfg.ComponentsRoot).Fixture("back-link", "default")
	if !ok {
		t.Fatal("default back-link fixture missing")
	}
	c := newClient(t, func(options *app.Options) {
		options.Components = render.Func(func(name string, params map[string]any) (string, error) {
			if name == "back-link" {
				return fixture.HTML, nil
			}
			encoded, err := json.Marshal(params)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("<div data-component=%q>%s</div>", name, encoded), nil
		})
	})
	page := c.get("/components/back-link?fixture=default")
	if !strings.Contains(page.Body.String(), "HTML matches the fixture") {
		t.Fatalf("preview did not report a match:\n%s", page.Body.String())
	}
}

func TestComponentRenderFailureShowsTheProblemPage(t *testing.T) {
	t.Parallel()
	c := newClient(t, func(options *app.Options) {
		options.Components = render.Func(func(name string, _ map[string]any) (string, error) {
			if name == "accordion" {
				return "", io.EOF
			}
			return "<div></div>", nil
		})
		options.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	})
	reply := c.get("/components/accordion")
	if reply.Code != http.StatusInternalServerError || !strings.Contains(reply.Body.String(), "problem with the service") {
		t.Fatalf("render failure = %d", reply.Code)
	}
}

func TestCatalogueFailureShowsTheProblemPage(t *testing.T) {
	t.Parallel()
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	broken := *cfg
	broken.ComponentsRoot = filepath.Join(t.TempDir(), "missing")
	handler, err := app.New(app.Options{
		Config:       &broken,
		Components:   stubComponents(),
		DemosEnabled: true,
		Logger:       slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(cfg.Stylesheet); err != nil {
		t.Skip(err)
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/components", nil))
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("catalogue = %d, want 500", recorder.Code)
	}
}
