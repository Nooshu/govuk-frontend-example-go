// Package app is the HTTP surface of the example service: the route table, the session cookie,
// and the response baseline every reply goes through.
//
// One handler serves every request so that no route can skip the shared behaviour. Each reply
// gets the OWASP headers and cache kind from baseline/policy.json, a session cookie, and
// Brotli-first compression; pages that show the applicant's answers are marked so they are
// never stored.
package app

import (
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Nooshu/govuk-frontend-example-go/internal/baseline"
	"github.com/Nooshu/govuk-frontend-example-go/internal/components"
	"github.com/Nooshu/govuk-frontend-example-go/internal/config"
	"github.com/Nooshu/govuk-frontend-example-go/internal/httpx"
	"github.com/Nooshu/govuk-frontend-example-go/internal/pages"
	"github.com/Nooshu/govuk-frontend-example-go/internal/render"
	"github.com/Nooshu/govuk-frontend-example-go/internal/service"
	"github.com/Nooshu/govuk-frontend-example-go/internal/session"
)

// Session cookie names. Over HTTPS the __Host- prefix locks the cookie to this exact origin and
// forbids a Domain attribute, so a sibling subdomain cannot set or read it. Plain HTTP cannot
// use that prefix, which is why local development falls back to rod_session.
const (
	sessionCookie     = "rod_session"
	hostSessionCookie = "__Host-session"
)

// sessionMaxAge is how long a session cookie lives, in seconds.
const sessionMaxAge = 4 * 60 * 60

// Options configure an [App]. Tests replace the store, the clock, and the component renderer.
type Options struct {
	// Config holds the resolved repository paths and the pinned Frontend version.
	Config *config.Config
	// Components renders GOV.UK Frontend components.
	Components render.Renderer
	// Store keeps sessions. Defaults to an in-memory store.
	Store session.Store
	// Now is the clock used for ages and the start month list. Defaults to time.Now.
	Now func() time.Time
	// DemosEnabled serves the component catalogue and example pages.
	DemosEnabled bool
	// Logger records unexpected failures. Defaults to the standard structured logger.
	Logger *slog.Logger
	// TemplateFS, when set, replaces the embedded page templates. Leave nil in production.
	TemplateFS fs.FS
}

// App serves the example service.
type App struct {
	config       *config.Config
	policy       *baseline.Policy
	assets       *httpx.Assets
	library      *components.Library
	renderer     *pages.Renderer
	components   render.Renderer
	store        session.Store
	now          func() time.Time
	demosEnabled bool
	logger       *slog.Logger
	mux          *http.ServeMux
}

// New builds the application.
//
// It reads the response policy, the compiled stylesheet, and GOV.UK Frontend's script up front,
// so a missing `npm install` or `npm run build:styles` fails at start-up rather than on the
// first request.
func New(options Options) (*App, error) {
	if options.Config == nil {
		return nil, errors.New("app: Config is required")
	}
	if options.Components == nil {
		return nil, errors.New("app: Components renderer is required")
	}
	policy, err := baseline.Load(options.Config.PolicyFile)
	if err != nil {
		return nil, err
	}
	assets, err := httpx.NewAssets(options.Config.GovukRoot, options.Config.FrontendAssets, options.Config.Stylesheet)
	if err != nil {
		return nil, err
	}
	application := &App{
		config:       options.Config,
		policy:       policy,
		assets:       assets,
		library:      components.NewLibrary(options.Config.ComponentsRoot),
		components:   options.Components,
		store:        options.Store,
		now:          options.Now,
		demosEnabled: options.DemosEnabled,
		logger:       options.Logger,
	}
	if application.store == nil {
		application.store = session.NewMemoryStore()
	}
	if application.now == nil {
		application.now = time.Now
	}
	if application.logger == nil {
		application.logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	}
	renderer, err := pages.NewRenderer(pages.Options{
		Components: options.Components,
		Assets: func() pages.Assets {
			page := assets.Page()
			return pages.Assets{StylesheetHref: page.StylesheetHref, AppModuleHref: page.AppModuleHref}
		},
		JSEnabledSnippet: policy.JSEnabledSnippet,
		ServiceName:      config.ServiceName,
		ServiceNameCy:    config.ServiceNameCy,
		FrontendVersion:  options.Config.FrontendVersion,
		DemosEnabled:     options.DemosEnabled,
		TemplateFS:       options.TemplateFS,
	})
	if err != nil {
		return nil, err
	}
	application.renderer = renderer
	application.mux = application.routes()
	return application, nil
}

// outcome is what a route decided to do: show a page, redirect, or send bytes that are not a
// page, such as an asset or a fixture fragment.
type outcome struct {
	view     *pages.View
	redirect string
	raw      *rawResponse
}

type rawResponse struct {
	status      int
	body        []byte
	contentType string
	kind        string
}

// ServeHTTP handles one request.
//
// A trailing slash is answered here rather than by ServeMux's own redirect, so that the reply
// still carries the baseline headers and the session cookie. Everything else goes to the route
// table.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if len(path) > 1 && strings.HasSuffix(path, "/") && !strings.HasPrefix(path, "/assets/") {
		location := strings.TrimSuffix(path, "/")
		if r.URL.RawQuery != "" {
			location += "?" + r.URL.RawQuery
		}
		a.writeRedirect(w, r, location, a.openSession(r))
		return
	}
	a.mux.ServeHTTP(w, r)
}

// openSession finds the caller's session, creating one when there is no usable cookie.
func (a *App) openSession(r *http.Request) *session.Session {
	cookies := httpx.ParseCookies(r.Header.Get("Cookie"))
	id := cookies[hostSessionCookie]
	if id == "" {
		id = cookies[sessionCookie]
	}
	if id != "" {
		if existing, ok := a.store.Get(id); ok {
			return existing
		}
	}
	return a.store.Create()
}

// sessionSetCookie builds the Set-Cookie header for a session.
func (a *App) sessionSetCookie(r *http.Request, current *session.Session) (string, error) {
	secure := requestIsSecure(r)
	name := sessionCookie
	if secure {
		name = hostSessionCookie
	}
	maxAge := sessionMaxAge
	return a.policy.SetCookie(name, current.ID, baseline.CookieOptions{
		Secure:     &secure,
		MaxAge:     &maxAge,
		HostPrefix: secure,
	})
}

// requestIsSecure reports whether the request reached the service over HTTPS, including through
// a proxy that terminated TLS.
func requestIsSecure(r *http.Request) bool {
	if r.TLS != nil || r.URL.Scheme == "https" {
		return true
	}
	forwarded := r.Header.Get("X-Forwarded-Proto")
	if forwarded == "" {
		return false
	}
	first, _, _ := strings.Cut(forwarded, ",")
	return strings.EqualFold(strings.TrimSpace(first), "https")
}

func (a *App) currentPath(r *http.Request) string {
	path := r.URL.Path
	if r.URL.RawQuery != "" {
		path += "?" + r.URL.RawQuery
	}
	return path
}

func (a *App) errorsFor(current *session.Session, path string) []service.FieldError {
	var items []service.FieldError
	if current.Errors != nil && current.Errors.Path == path {
		items = current.Errors.Items
	}
	current.Errors = nil
	return items
}

func (a *App) noticeFor(current *session.Session, path string) string {
	text := ""
	if current.Notice != nil && current.Notice.Path == path {
		text = current.Notice.Text
	}
	current.Notice = nil
	return text
}
