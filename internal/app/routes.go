package app

import (
	"crypto/subtle"
	"errors"
	"html/template"
	"net/http"
	"strings"

	"github.com/Nooshu/govuk-frontend-example-go/internal/baseline"
	"github.com/Nooshu/govuk-frontend-example-go/internal/config"
	"github.com/Nooshu/govuk-frontend-example-go/internal/httpx"
	"github.com/Nooshu/govuk-frontend-example-go/internal/pages"
	"github.com/Nooshu/govuk-frontend-example-go/internal/service"
	"github.com/Nooshu/govuk-frontend-example-go/internal/session"
)

// routes builds the route table on the standard library's ServeMux, using the method and
// wildcard patterns added in Go 1.22.
//
// Patterns carry the method, so a GET-only route answers 405 to a POST without a hand-written
// check, and /components/{name} reads as the URL it matches. Everything that is not matched
// falls through to the "/" pattern, which serves the GOV.UK page-not-found page.
func (a *App) routes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		a.writeText(w, r, http.StatusOK, "ok")
	})
	mux.HandleFunc("GET /assets/", a.serveAsset)

	// Start pages. "/{$}" matches only the root, not everything below it.
	mux.Handle("GET /{$}", a.page(func(*session.Session, *http.Request) (outcome, error) {
		return outcome{view: new(startView(pages.LangEN))}, nil
	}))
	mux.Handle("GET /cy", a.page(func(*session.Session, *http.Request) (outcome, error) {
		return outcome{view: new(startView(pages.LangCY))}, nil
	}))
	mux.Handle("GET /new-application", a.page(func(current *session.Session, _ *http.Request) (outcome, error) {
		current.Application = service.NewApplication()
		return outcome{redirect: "/"}, nil
	}))

	// The application itself.
	mux.Handle("GET /task-list", a.page(func(current *session.Session, _ *http.Request) (outcome, error) {
		return outcome{view: new(taskListView(current))}, nil
	}))
	mux.Handle("GET /check-answers", a.page(func(current *session.Session, _ *http.Request) (outcome, error) {
		return a.checkAnswersGet(current), nil
	}))
	mux.Handle("POST /check-answers", a.form(func(current *session.Session, _ *httpx.Body, _ *http.Request) (outcome, error) {
		return a.postCheckAnswers(current), nil
	}))
	mux.Handle("GET /confirmation", a.page(func(current *session.Session, _ *http.Request) (outcome, error) {
		return confirmationGet(current), nil
	}))
	for _, step := range service.Steps() {
		mux.Handle("GET "+step.Path, a.page(a.stepGet(step)))
		mux.Handle("POST "+step.Path, a.form(a.stepPost(step)))
	}

	// Supporting content.
	mux.Handle("GET /fees", a.static(feesView))
	mux.Handle("GET /help", a.static(helpView))
	mux.Handle("GET /guidance", a.static(guidanceView))
	mux.Handle("GET /accessibility", a.static(accessibilityView))
	mux.Handle("GET /about", a.static(aboutView))
	mux.Handle("GET /updates", a.page(func(_ *session.Session, r *http.Request) (outcome, error) {
		return updatesGet(r), nil
	}))

	// Cookies.
	mux.Handle("GET /cookies", a.page(func(current *session.Session, _ *http.Request) (outcome, error) {
		return outcome{view: new(a.cookiesView(current))}, nil
	}))
	mux.Handle("POST /cookies", a.form(func(current *session.Session, body *httpx.Body, _ *http.Request) (outcome, error) {
		return a.postCookies(body, current), nil
	}))
	mux.Handle("POST /cookie-choices", a.form(func(current *session.Session, body *httpx.Body, _ *http.Request) (outcome, error) {
		return postCookieBanner(body, current), nil
	}))

	if a.demosEnabled {
		a.demoRoutes(mux)
	}

	mux.HandleFunc("/", a.fallback)
	return mux
}

// demoRoutes serve the component catalogue and the example pages.
//
// They are registered only when demos are on, because a live service should not publish fixture
// previews of components it does not use.
func (a *App) demoRoutes(mux *http.ServeMux) {
	mux.Handle("GET /components", a.page(func(*session.Session, *http.Request) (outcome, error) {
		view, err := a.componentsView()
		if err != nil {
			return outcome{}, err
		}
		return outcome{view: &view}, nil
	}))
	mux.Handle("GET /components/{name}", a.page(func(_ *session.Session, r *http.Request) (outcome, error) {
		view, err := a.componentView(r.PathValue("name"), r.URL.Query().Get("fixture"))
		if err != nil {
			return outcome{}, err
		}
		if view == nil {
			return outcome{view: new(notFoundView())}, nil
		}
		return outcome{view: view}, nil
	}))
	mux.Handle("GET /components/{name}/fixture", a.page(func(_ *session.Session, r *http.Request) (outcome, error) {
		return a.fixtureFragment(r.PathValue("name"), r.URL.Query().Get("fixture")), nil
	}))
	mux.Handle("GET /examples", a.static(examplesView))
	mux.Handle("GET /examples/exit-this-page", a.static(exitThisPageView))
	mux.Handle("GET /examples/service-unavailable", a.static(unavailableView))
	mux.Handle("GET /examples/problem-with-the-service", a.static(func() pages.View {
		return problemView(http.StatusOK)
	}))
}

// fallback answers anything the route table did not match.
//
// A method this service never uses gets 405 rather than the page-not-found page, because the
// path may well exist — it is the verb that does not.
func (a *App) fallback(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet, http.MethodHead, http.MethodPost:
		a.writePage(w, r, notFoundView(), a.openSession(r))
	default:
		a.writeText(w, r, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// page adapts a handler that needs the session to an [http.Handler].
//
// Opening the session here rather than in each handler is what guarantees every reply carries a
// session cookie and the baseline response headers.
func (a *App) page(handle func(*session.Session, *http.Request) (outcome, error)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := a.openSession(r)
		result, err := handle(current, r)
		if err != nil {
			a.logger.Error("request failed", "path", r.URL.Path, "error", err)
			a.writePage(w, r, problemView(http.StatusInternalServerError), current)
			return
		}
		a.store.Save(current)
		a.write(w, r, result, current)
	})
}

// static adapts a page that depends on nothing but its own content.
func (a *App) static(build func() pages.View) http.Handler {
	return a.page(func(*session.Session, *http.Request) (outcome, error) {
		return outcome{view: new(build())}, nil
	})
}

// form adapts a handler that reads a posted body.
//
// The body is read and the CSRF token checked before the handler runs, so no POST handler can
// forget either. A token that does not match the session is treated as an expired session
// rather than a silent failure, because that is what it usually is.
// readBody is [httpx.ReadBody]. Tests replace it with a failure that is not a [httpx.BodyError],
// which the reader itself does not return.
var readBody = httpx.ReadBody

func (a *App) form(handle func(*session.Session, *httpx.Body, *http.Request) (outcome, error)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := a.openSession(r)
		body, err := readBody(r, config.MaxBodyBytes)
		if err != nil {
			if bodyError, ok := errors.AsType[*httpx.BodyError](err); ok {
				a.writeText(w, r, bodyError.Status, bodyError.Message)
				return
			}
			a.logger.Error("reading the request body failed", "path", r.URL.Path, "error", err)
			a.writePage(w, r, problemView(http.StatusInternalServerError), current)
			return
		}
		if !csrfOK(current, body) {
			a.store.Save(current)
			a.writePage(w, r, sessionExpiredView(), current)
			return
		}
		result, err := handle(current, body, r)
		if err != nil {
			a.logger.Error("request failed", "path", r.URL.Path, "error", err)
			a.writePage(w, r, problemView(http.StatusInternalServerError), current)
			return
		}
		a.store.Save(current)
		a.write(w, r, result, current)
	})
}

func (a *App) write(w http.ResponseWriter, r *http.Request, result outcome, current *session.Session) {
	switch {
	case result.raw != nil:
		a.writeRaw(w, r, *result.raw)
	case result.redirect != "":
		a.writeRedirect(w, r, result.redirect, current)
	case result.view != nil:
		a.writePage(w, r, *result.view, current)
	default:
		a.writePage(w, r, notFoundView(), current)
	}
}

func (a *App) serveAsset(w http.ResponseWriter, r *http.Request) {
	asset, ok := a.assets.Resolve(r.URL.Path)
	if !ok {
		a.writeText(w, r, http.StatusNotFound, "Not found")
		return
	}
	a.writeAsset(w, r, asset)
}

func (a *App) stepGet(step service.Step) func(*session.Session, *http.Request) (outcome, error) {
	return func(current *session.Session, r *http.Request) (outcome, error) {
		view, err := a.stepView(step, current, r, a.errorsFor(current, step.Path))
		if err != nil {
			return outcome{}, err
		}
		return outcome{view: &view}, nil
	}
}

func (a *App) stepPost(step service.Step) func(*session.Session, *httpx.Body, *http.Request) (outcome, error) {
	return func(current *session.Session, body *httpx.Body, _ *http.Request) (outcome, error) {
		return a.postStep(step, body, current), nil
	}
}

// postStep validates and saves one answer, then redirects.
//
// A failed answer redirects back to its own question rather than rendering in place, so the
// browser's URL always matches the page shown and a refresh never re-posts.
func (a *App) postStep(step service.Step, body *httpx.Body, current *session.Session) outcome {
	errs := a.validateStep(step, body)
	current.Application = applyStep(step, body, current.Application, len(errs) == 0)
	if len(errs) > 0 {
		current.Errors = &session.Errors{Path: step.Path, Items: errs}
		return outcome{redirect: withReturn(step.Path, body)}
	}
	current.Errors = nil
	if body.Field("returnTo") == "check-answers" {
		return outcome{redirect: "/check-answers"}
	}
	if next, ok := service.NextStep(step.ID); ok {
		return outcome{redirect: next.Path}
	}
	return outcome{redirect: "/check-answers"}
}

func (a *App) validateStep(step service.Step, body *httpx.Body) []service.FieldError {
	switch step.ID {
	case service.StepName:
		return service.ValidateName(body.Field("first-name"), body.Field("last-name"))
	case service.StepDateOfBirth:
		return service.ValidateDateOfBirth(
			body.Field("date-of-birth-day"),
			body.Field("date-of-birth-month"),
			body.Field("date-of-birth-year"),
			a.now(),
		)
	case service.StepEmail:
		return service.ValidateEmail(body.Field("email"))
	case service.StepContactPreference:
		return service.ValidateContactPreference(body.Field("contact-by"), body.Field("telephone"))
	case service.StepWhereYouWillFish:
		return service.ValidateRegions(body.Values("regions"))
	case service.StepLicenceLength:
		return service.ValidateLicenceLength(body.Field("licence-length"))
	case service.StepStartMonth:
		return service.ValidateStartMonth(body.Field("start-month"), a.now())
	case service.StepAddress:
		return service.ValidateAddress(body.Field("address-line-1"), body.Field("town"), body.Field("postcode"))
	case service.StepEvidence:
		filename, _ := uploadedName(body)
		return service.ValidateEvidence(filename)
	case service.StepAdditionalDetails:
		return service.ValidateAdditionalDetails(body.Field("additional-details"))
	default:
		return service.ValidatePassword(body.Field("password"), body.Field("password-confirm"))
	}
}

func applyStep(step service.Step, body *httpx.Body, application service.Application, valid bool) service.Application {
	switch step.ID {
	case service.StepName:
		return service.SaveName(application, body.Field("first-name"), body.Field("last-name"), valid)
	case service.StepDateOfBirth:
		return service.SaveDate(
			application,
			body.Field("date-of-birth-day"),
			body.Field("date-of-birth-month"),
			body.Field("date-of-birth-year"),
			valid,
		)
	case service.StepEmail:
		return service.SaveEmail(application, body.Field("email"), valid)
	case service.StepContactPreference:
		return service.SaveContact(application, body.Field("contact-by"), body.Field("telephone"), valid)
	case service.StepWhereYouWillFish:
		return service.SaveRegions(application, body.Values("regions"), valid)
	case service.StepLicenceLength:
		return service.SaveLicence(application, body.Field("licence-length"), valid)
	case service.StepStartMonth:
		return service.SaveMonth(application, body.Field("start-month"), valid)
	case service.StepAddress:
		return service.SaveAddress(application, service.Address{
			Line1:    body.Field("address-line-1"),
			Line2:    body.Field("address-line-2"),
			Town:     body.Field("town"),
			Postcode: body.Field("postcode"),
		}, valid)
	case service.StepEvidence:
		filename, hasFile := uploadedName(body)
		return service.SaveEvidence(application, filename, hasFile, valid)
	case service.StepAdditionalDetails:
		return service.SaveDetails(application, body.Field("additional-details"), valid)
	default:
		return service.SavePassword(application, valid)
	}
}

// uploadedName returns the safe base name of the evidence upload.
//
// The file's bytes are never stored; only the name is shown back to the applicant, and only
// after it has been checked.
func uploadedName(body *httpx.Body) (string, bool) {
	if body.Upload == nil || body.Upload.FieldName != "evidence" {
		return "", false
	}
	return service.SafeFilename(body.Upload.Filename)
}

// fixtureFragment serves the official fixture HTML on its own, so a preview can be compared
// with what this service renders.
func (a *App) fixtureFragment(name, fixtureName string) outcome {
	fixture, ok := a.library.Fixture(name, fixtureName)
	if !ok {
		return outcome{view: new(notFoundView())}
	}
	return outcome{raw: &rawResponse{
		status:      http.StatusOK,
		body:        []byte(fixture.HTML),
		contentType: "text/html; charset=utf-8",
		kind:        baseline.KindDocument,
	}}
}

func postCookieBanner(body *httpx.Body, current *session.Session) outcome {
	switch choice := body.Field("cookies"); choice {
	case session.ChoiceAccept, session.ChoiceReject:
		current.CookieChoice = choice
		current.CookieBanner = choice
	case "hide":
		current.CookieBanner = ""
	}
	return outcome{redirect: safeReturn(body.Field("returnPath"))}
}

func (a *App) postCookies(body *httpx.Body, current *session.Session) outcome {
	errs := service.ValidateCookieChoice(body.Field("analytics"))
	if len(errs) > 0 {
		current.Errors = &session.Errors{Path: "/cookies", Items: errs}
		current.Notice = nil
		return outcome{redirect: "/cookies"}
	}
	current.CookieChoice = session.ChoiceReject
	if body.Field("analytics") == "yes" {
		current.CookieChoice = session.ChoiceAccept
	}
	current.CookieBanner = ""
	current.Errors = nil
	current.Notice = &session.Notice{Path: "/cookies", Text: "Your cookie settings were saved"}
	return outcome{redirect: "/cookies"}
}

// postCheckAnswers submits the application.
//
// Submitting twice does not create a second application: an already-submitted session goes
// straight to its confirmation, so a refresh or a back-then-submit cannot duplicate it.
func (a *App) postCheckAnswers(current *session.Session) outcome {
	if current.Application.Submitted {
		return outcome{redirect: "/confirmation"}
	}
	if incomplete, ok := service.FirstIncompleteStep(current.Application); ok {
		return outcome{redirect: incomplete.Path}
	}
	current.Application.Submitted = true
	current.Application.Reference = session.ReferenceFor(current.ID)
	return outcome{redirect: "/confirmation"}
}

func csrfOK(current *session.Session, body *httpx.Body) bool {
	token := body.Field("csrf")
	if token == "" || current.CSRF == "" {
		return false
	}
	// Constant-time compare avoids leaking the token length or contents via response timing.
	return subtle.ConstantTimeCompare([]byte(token), []byte(current.CSRF)) == 1
}

// withReturn keeps the "came from check your answers" flag across the redirect that shows the
// error, so fixing an answer still returns the applicant to their summary.
func withReturn(path string, body *httpx.Body) string {
	if body.Field("returnTo") == "check-answers" {
		return path + "?return=check-answers"
	}
	return path
}

// safeReturn only allows a same-origin path, so the cookie banner cannot be used as an open
// redirect.
func safeReturn(value string) string {
	if !strings.HasPrefix(value, "/") ||
		strings.HasPrefix(value, "//") ||
		strings.Contains(value, "://") ||
		strings.Contains(value, `\`) ||
		strings.ContainsAny(value, "\r\n") {
		return "/"
	}
	return value
}

// fixtureLink is one entry in the fixture list on a component page.
type fixtureLink struct {
	Name    string
	Current bool
}

func htmlOf(value string) template.HTML { return template.HTML(value) }
