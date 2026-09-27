package app_test

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	jsonv2 "encoding/json/v2"

	"github.com/Nooshu/govuk-frontend-example-go/internal/app"
	"github.com/Nooshu/govuk-frontend-example-go/internal/config"
	"github.com/Nooshu/govuk-frontend-example-go/internal/render"
)

// now is the clock every test uses, so the start-month list and the age check never depend on
// when the suite runs.
var now = time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC)

// stubComponents stands in for GOV.UK Frontend. These tests are about routing, validation, and
// the response baseline; whether the component HTML matches the official fixtures is proven by
// the fixture suite in internal/govuk, not here.
//
// The stub writes the options it was given into the page, so a test can assert that the right
// answer reached the right component without depending on that component's markup.
func stubComponents() render.Renderer {
	return render.Func(func(name string, params map[string]any) (string, error) {
		// Deterministic keeps map key order stable so ETag / If-None-Match tests stay reliable.
		encoded, err := jsonv2.Marshal(params, jsonv2.Deterministic(true))
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("<div data-component=%q>%s</div>", name, encoded), nil
	})
}

type client struct {
	t       *testing.T
	handler http.Handler
	cookies map[string]string
	token   string
}

func newClient(t *testing.T, options ...func(*app.Options)) *client {
	t.Helper()
	root, err := config.Find(repoRoot(t))
	if err != nil {
		t.Fatalf("config.Find: %v", err)
	}
	cfg, err := config.New(root)
	if err != nil {
		t.Fatalf("config.New: %v", err)
	}
	if _, err := os.Stat(cfg.Stylesheet); err != nil {
		t.Skipf("run `npm run build:styles` before the Go tests: %v", err)
	}
	opts := app.Options{
		Config:       cfg,
		Components:   stubComponents(),
		Now:          func() time.Time { return now },
		DemosEnabled: true,
		Logger:       slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	for _, option := range options {
		option(&opts)
	}
	handler, err := app.New(opts)
	if err != nil {
		t.Fatalf("app.New: %v", err)
	}
	return &client{t: t, handler: handler, cookies: map[string]string{}}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	return filepath.Clean(dir)
}

func (c *client) do(r *http.Request) *httptest.ResponseRecorder {
	c.t.Helper()
	for name, value := range c.cookies {
		r.AddCookie(&http.Cookie{Name: name, Value: value})
	}
	recorder := httptest.NewRecorder()
	c.handler.ServeHTTP(recorder, r)
	for _, cookie := range recorder.Result().Cookies() {
		c.cookies[cookie.Name] = cookie.Value
	}
	return recorder
}

func (c *client) get(path string) *httptest.ResponseRecorder {
	c.t.Helper()
	return c.do(httptest.NewRequest(http.MethodGet, path, nil))
}

func (c *client) post(path string, form url.Values) *httptest.ResponseRecorder {
	c.t.Helper()
	form.Set("csrf", c.csrf())
	request := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return c.do(request)
}

var csrfPattern = regexp.MustCompile(`name="csrf" value="([^"]+)"`)

// csrf reads the token out of a rendered page, which is how a browser would obtain it. The
// token belongs to the session rather than the page, so it is read once and reused.
func (c *client) csrf() string {
	c.t.Helper()
	if c.token != "" {
		return c.token
	}
	page := c.get("/name")
	match := csrfPattern.FindStringSubmatch(page.Body.String())
	if match == nil {
		c.t.Fatalf("no CSRF token on /name (status %d)", page.Code)
	}
	c.token = match[1]
	return c.token
}

func TestTheApplicantCanApplyForALicence(t *testing.T) {
	t.Parallel()
	c := newClient(t)

	start := c.get("/")
	if start.Code != http.StatusOK {
		t.Fatalf("GET / = %d, want 200", start.Code)
	}
	if !strings.Contains(start.Body.String(), "Apply for a rod fishing licence") {
		t.Error("the start page does not name the service")
	}

	if got := c.get("/task-list").Code; got != http.StatusOK {
		t.Fatalf("GET /task-list = %d, want 200", got)
	}

	// Check-your-answers is closed until every required question is answered.
	if location := c.get("/check-answers").Header().Get("Location"); location != "/name" {
		t.Errorf("GET /check-answers redirected to %q, want /name", location)
	}

	for _, step := range journey() {
		reply := c.post(step.path, step.form)
		if reply.Code != http.StatusSeeOther {
			t.Fatalf("POST %s = %d, want 303\n%s", step.path, reply.Code, reply.Body.String())
		}
		if location := reply.Header().Get("Location"); location != step.next {
			t.Fatalf("POST %s redirected to %q, want %q", step.path, location, step.next)
		}
	}

	summary := c.get("/check-answers")
	if summary.Code != http.StatusOK {
		t.Fatalf("GET /check-answers = %d, want 200", summary.Code)
	}
	if got := summary.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("check-answers Cache-Control = %q, want no-store", got)
	}

	submitted := c.post("/check-answers", url.Values{})
	if location := submitted.Header().Get("Location"); location != "/confirmation" {
		t.Fatalf("submitting redirected to %q, want /confirmation", location)
	}

	confirmation := c.get("/confirmation")
	if confirmation.Code != http.StatusOK {
		t.Fatalf("GET /confirmation = %d, want 200", confirmation.Code)
	}
	if !strings.Contains(confirmation.Body.String(), "Application complete") {
		t.Error("the confirmation page does not say the application is complete")
	}

	// Submitting again reaches the same confirmation rather than creating a second application.
	if location := c.post("/check-answers", url.Values{}).Header().Get("Location"); location != "/confirmation" {
		t.Errorf("re-submitting redirected to %q, want /confirmation", location)
	}
	if location := c.get("/check-answers").Header().Get("Location"); location != "/confirmation" {
		t.Errorf("check-answers after submitting redirected to %q, want /confirmation", location)
	}
}

type journeyStep struct {
	path string
	form url.Values
	next string
}

func journey() []journeyStep {
	return []journeyStep{
		{"/name", url.Values{"first-name": {"Ada"}, "last-name": {"Lovelace"}}, "/date-of-birth"},
		{"/date-of-birth", url.Values{
			"date-of-birth-day":   {"10"},
			"date-of-birth-month": {"12"},
			"date-of-birth-year":  {"1990"},
		}, "/email"},
		{"/email", url.Values{"email": {"ada@example.com"}}, "/contact-preference"},
		{"/contact-preference", url.Values{"contact-by": {"email"}}, "/where-you-will-fish"},
		{"/where-you-will-fish", url.Values{"regions": {"north-west", "wales"}}, "/licence-length"},
		{"/licence-length", url.Values{"licence-length": {"12-month"}}, "/start-month"},
		{"/start-month", url.Values{"start-month": {"2026-04"}}, "/address"},
		{"/address", url.Values{
			"address-line-1": {"1 Example Street"},
			"town":           {"Exampleton"},
			"postcode":       {"sw1a 1aa"},
		}, "/evidence"},
		{"/evidence", url.Values{}, "/additional-details"},
		{"/additional-details", url.Values{"additional-details": {"Nothing else"}}, "/create-a-password"},
		{"/create-a-password", url.Values{
			"password":         {"correct horse"},
			"password-confirm": {"correct horse"},
		}, "/check-answers"},
	}
}

func completeJourney(t *testing.T, c *client) {
	t.Helper()
	for _, step := range journey() {
		if reply := c.post(step.path, step.form); reply.Code != http.StatusSeeOther {
			t.Fatalf("POST %s = %d, want 303", step.path, reply.Code)
		}
	}
}

func TestAFailedAnswerReturnsToItsQuestionWithAnErrorSummary(t *testing.T) {
	t.Parallel()
	c := newClient(t)

	reply := c.post("/name", url.Values{"first-name": {""}, "last-name": {""}})
	if reply.Code != http.StatusSeeOther {
		t.Fatalf("POST /name = %d, want 303", reply.Code)
	}
	if location := reply.Header().Get("Location"); location != "/name" {
		t.Fatalf("a failed answer redirected to %q, want /name", location)
	}

	page := c.get("/name")
	body := page.Body.String()
	if !strings.Contains(body, "error-summary") {
		t.Error("the question does not show an error summary after a failed answer")
	}
	if !strings.Contains(body, "<title>Error: ") {
		t.Errorf("the title does not announce the error:\n%s", firstLines(body, 12))
	}

	// Errors are shown once; a refresh shows the clean question again.
	if strings.Contains(c.get("/name").Body.String(), "error-summary") {
		t.Error("the error summary is still shown after it has been read")
	}
}

func TestChangingAnAnswerReturnsToCheckYourAnswers(t *testing.T) {
	t.Parallel()
	c := newClient(t)
	completeJourney(t, c)

	page := c.get("/name?return=check-answers")
	if page.Code != http.StatusOK {
		t.Fatalf("GET /name?return=check-answers = %d, want 200", page.Code)
	}
	if !strings.Contains(page.Body.String(), `name="returnTo" value="check-answers"`) {
		t.Error("the question does not carry the return flag")
	}

	reply := c.post("/name", url.Values{
		"first-name": {"Grace"},
		"last-name":  {"Hopper"},
		"returnTo":   {"check-answers"},
	})
	if location := reply.Header().Get("Location"); location != "/check-answers" {
		t.Errorf("a changed answer redirected to %q, want /check-answers", location)
	}

	if !strings.Contains(c.get("/check-answers").Body.String(), "Grace Hopper") {
		t.Error("the summary does not show the changed name")
	}
}

func TestAFailedChangeKeepsTheReturnFlag(t *testing.T) {
	t.Parallel()
	c := newClient(t)

	reply := c.post("/name", url.Values{"first-name": {""}, "returnTo": {"check-answers"}})
	if location := reply.Header().Get("Location"); location != "/name?return=check-answers" {
		t.Errorf("a failed change redirected to %q, want /name?return=check-answers", location)
	}
}

func TestAnEvidenceUploadKeepsOnlyItsName(t *testing.T) {
	t.Parallel()
	c := newClient(t)

	reply := c.upload("/evidence", "evidence", "concession.pdf", "not a real pdf")
	if location := reply.Header().Get("Location"); location != "/additional-details" {
		t.Fatalf("uploading redirected to %q, want /additional-details", location)
	}
	if !strings.Contains(c.get("/evidence").Body.String(), "Current file: concession.pdf") {
		t.Error("the question does not show the uploaded file name")
	}

	rejected := c.upload("/evidence", "evidence", "virus.exe", "MZ")
	if location := rejected.Header().Get("Location"); location != "/evidence" {
		t.Errorf("an unsupported file redirected to %q, want /evidence", location)
	}
	if !strings.Contains(c.get("/evidence").Body.String(), "Current file: concession.pdf") {
		t.Error("a rejected upload should not erase the previous file name")
	}
}

func (c *client) upload(path, field, filename, content string) *httptest.ResponseRecorder {
	c.t.Helper()
	token := c.csrf()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("csrf", token); err != nil {
		c.t.Fatalf("WriteField: %v", err)
	}
	part, err := writer.CreateFormFile(field, filename)
	if err != nil {
		c.t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := part.Write([]byte(content)); err != nil {
		c.t.Fatalf("writing the upload: %v", err)
	}
	if err := writer.Close(); err != nil {
		c.t.Fatalf("closing the multipart writer: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, path, &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return c.do(request)
}

func TestAPostWithoutAMatchingTokenIsTreatedAsAnExpiredSession(t *testing.T) {
	t.Parallel()
	c := newClient(t)

	request := httptest.NewRequest(http.MethodPost, "/name", strings.NewReader("first-name=Ada&csrf=wrong"))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reply := c.do(request)

	if reply.Code != http.StatusForbidden {
		t.Errorf("POST with a bad token = %d, want 403", reply.Code)
	}
	if !strings.Contains(reply.Body.String(), "your session has expired") {
		t.Error("the reply does not explain that the session expired")
	}
}

func TestTheCookieBannerSavesAChoiceAndReturnsToThePage(t *testing.T) {
	t.Parallel()
	c := newClient(t)

	if !strings.Contains(c.get("/fees").Body.String(), "cookie-banner") {
		t.Fatal("the cookie banner is not shown before a choice is made")
	}

	accepted := c.post("/cookie-choices", url.Values{"cookies": {"accept"}, "returnPath": {"/fees"}})
	if location := accepted.Header().Get("Location"); location != "/fees" {
		t.Errorf("accepting cookies redirected to %q, want /fees", location)
	}

	hidden := c.post("/cookie-choices", url.Values{"cookies": {"hide"}, "returnPath": {"/fees"}})
	if hidden.Code != http.StatusSeeOther {
		t.Fatalf("hiding the banner = %d, want 303", hidden.Code)
	}
	if strings.Contains(c.get("/fees").Body.String(), "cookie-banner") {
		t.Error("the banner is still shown after the choice was saved and hidden")
	}
}

func TestTheCookieBannerCannotRedirectOffSite(t *testing.T) {
	t.Parallel()
	c := newClient(t)

	for _, path := range []string{"//example.com", "https://example.com", `/\example.com`, "/ok\r\nX: 1", "not-a-path"} {
		reply := c.post("/cookie-choices", url.Values{"cookies": {"accept"}, "returnPath": {path}})
		if location := reply.Header().Get("Location"); location != "/" {
			t.Errorf("returnPath %q redirected to %q, want /", path, location)
		}
	}
}

func TestTheCookiesPageSavesAndReportsTheChoice(t *testing.T) {
	t.Parallel()
	c := newClient(t)

	invalid := c.post("/cookies", url.Values{"analytics": {""}})
	if location := invalid.Header().Get("Location"); location != "/cookies" {
		t.Fatalf("an empty choice redirected to %q, want /cookies", location)
	}
	if !strings.Contains(c.get("/cookies").Body.String(), "error-summary") {
		t.Error("the cookies page does not show an error summary")
	}

	saved := c.post("/cookies", url.Values{"analytics": {"yes"}})
	if saved.Code != http.StatusSeeOther {
		t.Fatalf("saving cookie settings = %d, want 303", saved.Code)
	}
	if !strings.Contains(c.get("/cookies").Body.String(), "notification-banner") {
		t.Error("the cookies page does not confirm that the settings were saved")
	}
	// The confirmation is shown once.
	if strings.Contains(c.get("/cookies").Body.String(), "notification-banner") {
		t.Error("the confirmation is still shown after it has been read")
	}

	if c.post("/cookies", url.Values{"analytics": {"no"}}).Code != http.StatusSeeOther {
		t.Error("rejecting analytics cookies was not saved")
	}
}

func TestSupportingPagesAreServed(t *testing.T) {
	t.Parallel()
	c := newClient(t)

	for path, want := range map[string]string{
		"/fees":                              "Licence fees",
		"/help":                              "Help",
		"/guidance":                          "Guidance",
		"/updates":                           "Service updates",
		"/updates?page=2":                    "Service updates",
		"/accessibility":                     "Accessibility statement",
		"/about":                             "About this example",
		"/cy":                                "Gwneud cais am drwydded bysgota",
		"/components":                        "Component catalogue",
		"/examples":                          "Example pages",
		"/examples/exit-this-page":           "Exit this page",
		"/examples/service-unavailable":      "the service is unavailable",
		"/examples/problem-with-the-service": "there is a problem with the service",
	} {
		reply := c.get(path)
		if reply.Code != http.StatusOK {
			t.Errorf("GET %s = %d, want 200", path, reply.Code)
			continue
		}
		if !strings.Contains(reply.Body.String(), want) {
			t.Errorf("GET %s does not mention %q", path, want)
		}
	}

	if location := c.get("/updates?page=9").Header().Get("Location"); location != "/updates" {
		t.Errorf("an unknown page number redirected to %q, want /updates", location)
	}
}

func TestAnUnknownPathShowsThePageNotFoundPage(t *testing.T) {
	t.Parallel()
	c := newClient(t)

	reply := c.get("/this-page-does-not-exist")
	if reply.Code != http.StatusNotFound {
		t.Errorf("an unknown path = %d, want 404", reply.Code)
	}
	if !strings.Contains(reply.Body.String(), "Page not found") {
		t.Error("the reply is not the GOV.UK page-not-found page")
	}
}

func TestAnUnsupportedMethodIsRejected(t *testing.T) {
	t.Parallel()
	c := newClient(t)

	reply := c.do(httptest.NewRequest(http.MethodDelete, "/somewhere", nil))
	if reply.Code != http.StatusMethodNotAllowed {
		t.Errorf("DELETE = %d, want 405", reply.Code)
	}
	if reply := c.do(httptest.NewRequest(http.MethodPut, "/name", nil)); reply.Code != http.StatusMethodNotAllowed {
		t.Errorf("PUT on a known path = %d, want 405", reply.Code)
	}
}

func TestATrailingSlashRedirectsToTheCanonicalPath(t *testing.T) {
	t.Parallel()
	c := newClient(t)

	reply := c.get("/fees/")
	if reply.Code != http.StatusSeeOther {
		t.Fatalf("GET /fees/ = %d, want 303", reply.Code)
	}
	if location := reply.Header().Get("Location"); location != "/fees" {
		t.Errorf("GET /fees/ redirected to %q, want /fees", location)
	}
	if location := c.get("/fees/?a=1").Header().Get("Location"); location != "/fees?a=1" {
		t.Errorf("the query was not kept: %q", location)
	}
}

func TestHealthIsServedWithoutASession(t *testing.T) {
	t.Parallel()
	c := newClient(t)

	reply := c.get("/health")
	if reply.Code != http.StatusOK || reply.Body.String() != "ok" {
		t.Errorf("GET /health = %d %q, want 200 \"ok\"", reply.Code, reply.Body.String())
	}
	if reply.Header().Get("Set-Cookie") != "" {
		t.Error("the health check should not start a session")
	}
}

func TestStartingAgainClearsTheAnswers(t *testing.T) {
	t.Parallel()
	c := newClient(t)
	completeJourney(t, c)

	if location := c.get("/new-application").Header().Get("Location"); location != "/" {
		t.Fatal("starting again did not return to the start page")
	}
	if location := c.get("/check-answers").Header().Get("Location"); location != "/name" {
		t.Errorf("after starting again, check-answers redirected to %q, want /name", location)
	}
}

func TestConfirmationIsOnlyReachableAfterSubmitting(t *testing.T) {
	t.Parallel()
	c := newClient(t)

	if location := c.get("/confirmation").Header().Get("Location"); location != "/task-list" {
		t.Errorf("GET /confirmation before submitting redirected to %q, want /task-list", location)
	}
}

func firstLines(text string, count int) string {
	lines := strings.SplitN(text, "\n", count+1)
	if len(lines) > count {
		lines = lines[:count]
	}
	return strings.Join(lines, "\n")
}
