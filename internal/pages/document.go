// Package pages composes the GOV.UK page template in Go.
//
// Every block of GOV.UK user interface on a page — header, footer, phase banner, inputs, task
// list — is asked for by name from a [render.Renderer], which returns the HTML GOV.UK Frontend
// produces. This package only supplies the structure the page template defines around them:
// the document head, the skip link, the width container, the main landmark, and the footer.
//
// Nothing here spawns Node. The layout mirrors dist/govuk/template.njk so the Go output matches
// what Frontend's own page template would emit, including the js-enabled snippet whose hash is
// pinned in baseline/policy.json.
package pages

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"maps"
	"strings"

	"github.com/Nooshu/govuk-frontend-example-go/internal/htmlutil"
	"github.com/Nooshu/govuk-frontend-example-go/internal/render"
	"github.com/Nooshu/govuk-frontend-example-go/internal/session"
)

//go:embed templates/*.gohtml
var templateFiles embed.FS

// Languages this example serves. The journey itself is English; Welsh covers the start page.
const (
	LangEN = "en"
	LangCY = "cy"
)

// View is everything one page needs beyond the shell.
//
// A page has a back link or breadcrumbs, never both: the two answer the same question and
// showing both leaves users unsure which one goes back.
type View struct {
	// Template names the content template, which matches the page, such as "check-answers".
	Template string
	// Status is the HTTP status to reply with.
	Status int
	// Heading is the page h1, used in the document title.
	Heading string
	// HasErrors prefixes the title with "Error: ".
	HasErrors bool
	// Lang is [LangEN] or [LangCY]. Empty means English.
	Lang string
	// BackLink is back-link component options.
	BackLink map[string]any
	// Breadcrumbs is breadcrumbs component options.
	Breadcrumbs map[string]any
	// MainClasses are extra classes for the main landmark.
	MainClasses string
	// ShowFeedback adds the feedback component above the footer.
	ShowFeedback bool
	// Personal marks a page that shows the applicant's answers, so it is never cached.
	Personal bool
	// ExitThisPage is exit-this-page component options, placed first in the body.
	ExitThisPage map[string]any
	// Context is the data the content template needs.
	Context map[string]any
}

// Assets are the fingerprinted URLs a page puts in its head and at the end of its body.
type Assets struct {
	StylesheetHref string
	AppModuleHref  string
}

// Options configure a [Renderer].
type Options struct {
	// Components renders GOV.UK Frontend components.
	Components render.Renderer
	// Assets returns the current asset URLs.
	Assets func() Assets
	// JSEnabledSnippet is the inline script from baseline/policy.json. It must be emitted byte
	// for byte or the CSP hash in the same file stops matching and the script is blocked.
	JSEnabledSnippet string
	// ServiceName and ServiceNameCy are shown in the service navigation and page titles.
	ServiceName   string
	ServiceNameCy string
	// FrontendVersion is the pinned GOV.UK Frontend release, shown on the about page.
	FrontendVersion string
	// DemosEnabled adds the component catalogue and example pages to the footer.
	DemosEnabled bool
	// TemplateFS, when set, is parsed instead of the embedded page templates. Leave nil to use
	// the templates shipped with this package.
	TemplateFS fs.FS
}

// Renderer turns a [View] into a complete HTML document.
type Renderer struct {
	options   Options
	templates *template.Template
}

// NewRenderer parses the page templates and returns a renderer.
func NewRenderer(options Options) (*Renderer, error) {
	renderer := &Renderer{options: options}
	files := fs.FS(templateFiles)
	if options.TemplateFS != nil {
		files = options.TemplateFS
	}
	parsed, err := template.New("pages").Funcs(template.FuncMap{
		"component": renderer.component,
		"dict":      dict,
		"merge":     merge,
		"concat":    concat,
	}).ParseFS(files, "templates/*.gohtml")
	if err != nil {
		return nil, fmt.Errorf("pages: parsing templates: %w", err)
	}
	renderer.templates = parsed
	return renderer, nil
}

// pageData is what every template sees.
type pageData struct {
	Heading        string
	PageTitle      string
	HTMLLang       string
	SkipLinkText   string
	MainClasses    string
	CSRF           string
	ReturnPath     string
	StylesheetHref string
	AppModuleHref  string
	// JSEnabledSnippet is typed as JavaScript so html/template writes it verbatim.
	JSEnabledSnippet template.JS

	ServiceNavigation map[string]any
	PhaseBanner       map[string]any
	Footer            map[string]any
	CookieBanner      map[string]any
	Feedback          map[string]any
	BackLink          map[string]any
	Breadcrumbs       map[string]any
	ExitThisPage      map[string]any
	ShowFeedback      bool

	DemosEnabled    bool
	FrontendVersion string

	Context map[string]any
	Content template.HTML
}

var feedback = map[string]any{
	"titleText": "Help us improve this service",
	"html":      `<p class="govuk-body">This example does not send feedback. <a class="govuk-link" href="/help">Get help with this example</a>.</p>`,
}

// Render builds the full HTML document for a view.
//
// currentPath is the path and query of the request. It becomes the cookie banner's return path,
// so accepting or rejecting cookies brings the applicant back to the page they were reading.
func (r *Renderer) Render(view View, current *session.Session, currentPath string) (string, error) {
	if view.BackLink != nil && view.Breadcrumbs != nil {
		return "", fmt.Errorf("pages: %q sets both a back link and breadcrumbs", view.Template)
	}
	lang := view.Lang
	if lang == "" {
		lang = LangEN
	}
	serviceName := r.options.ServiceName
	skipLinkText := "Skip to main content"
	if lang == LangCY {
		serviceName = r.options.ServiceNameCy
		skipLinkText = "Neidio i'r prif gynnwys"
	}
	navigation, err := r.serviceNavigation(lang)
	if err != nil {
		return "", err
	}
	assets := r.options.Assets()
	data := &pageData{
		Heading:           view.Heading,
		PageTitle:         htmlutil.PageTitle(view.Heading, serviceName, view.HasErrors),
		HTMLLang:          lang,
		SkipLinkText:      skipLinkText,
		MainClasses:       view.MainClasses,
		CSRF:              current.CSRF,
		ReturnPath:        htmlutil.SafeLocalPath(currentPath),
		StylesheetHref:    assets.StylesheetHref,
		AppModuleHref:     assets.AppModuleHref,
		JSEnabledSnippet:  template.JS(r.options.JSEnabledSnippet),
		ServiceNavigation: navigation,
		PhaseBanner:       phaseBanner(lang),
		Footer:            r.footer(lang),
		CookieBanner:      cookieBanner(current),
		Feedback:          feedback,
		BackLink:          view.BackLink,
		Breadcrumbs:       view.Breadcrumbs,
		ExitThisPage:      view.ExitThisPage,
		ShowFeedback:      view.ShowFeedback,
		DemosEnabled:      r.options.DemosEnabled,
		FrontendVersion:   r.options.FrontendVersion,
		Context:           view.Context,
	}

	var content bytes.Buffer
	if err := r.templates.ExecuteTemplate(&content, view.Template+".gohtml", data); err != nil {
		return "", fmt.Errorf("pages: rendering %q: %w", view.Template, err)
	}
	data.Content = template.HTML(content.String())

	var document bytes.Buffer
	if err := r.templates.ExecuteTemplate(&document, "layout.gohtml", data); err != nil {
		return "", fmt.Errorf("pages: rendering the page shell for %q: %w", view.Template, err)
	}
	return document.String(), nil
}

// component asks the renderer for one component's HTML.
//
// params is typed as any so templates can pass a value straight out of a context map.
func (r *Renderer) component(name string, params any) (template.HTML, error) {
	options, err := asParams(params)
	if err != nil {
		return "", fmt.Errorf("pages: %s: %w", name, err)
	}
	html, err := r.options.Components.Render(name, options)
	if err != nil {
		return "", fmt.Errorf("pages: rendering the %s component: %w", name, err)
	}
	return template.HTML(html), nil
}

func asParams(params any) (map[string]any, error) {
	switch typed := params.(type) {
	case nil:
		return nil, nil
	case map[string]any:
		return typed, nil
	default:
		return nil, fmt.Errorf("component options must be a map, got %T", params)
	}
}

// dict builds a map inside a template, for macros that take a small wrapper object.
func dict(pairs ...any) (map[string]any, error) {
	if len(pairs)%2 != 0 {
		return nil, fmt.Errorf("pages: dict needs an even number of arguments")
	}
	result := make(map[string]any, len(pairs)/2)
	for index := 0; index < len(pairs); index += 2 {
		key, ok := pairs[index].(string)
		if !ok {
			return nil, fmt.Errorf("pages: dict keys must be strings, got %T", pairs[index])
		}
		result[key] = pairs[index+1]
	}
	return result, nil
}

// merge copies component options and adds more, so a template can pass HTML it has just
// rendered — such as the inputs inside a fieldset — without the options map knowing about it.
func merge(base any, pairs ...any) (map[string]any, error) {
	original, err := asParams(base)
	if err != nil {
		return nil, fmt.Errorf("pages: merge: %w", err)
	}
	added, err := dict(pairs...)
	if err != nil {
		return nil, err
	}
	result := make(map[string]any, len(original)+len(added))
	maps.Copy(result, original)
	maps.Copy(result, added)
	return result, nil
}

// concat joins rendered component HTML into one string, for macros that take an html option.
func concat(parts ...template.HTML) string {
	var joined strings.Builder
	for _, part := range parts {
		joined.WriteString(string(part))
	}
	return joined.String()
}

func (r *Renderer) serviceNavigation(lang string) (map[string]any, error) {
	serviceName, serviceURL := r.options.ServiceName, "/"
	ariaLabel := "Language"
	items := []any{
		map[string]any{"text": "English", "lang": "en", "current": true},
		map[string]any{"text": "Cymraeg", "lang": "cy", "href": "/cy"},
	}
	if lang == LangCY {
		serviceName, serviceURL = r.options.ServiceNameCy, "/cy"
		ariaLabel = "Iaith"
		items = []any{
			map[string]any{"text": "English", "lang": "en", "href": "/"},
			map[string]any{"text": "Cymraeg", "lang": "cy", "current": true},
		}
	}
	languages, err := r.options.Components.Render("language-navigation", map[string]any{
		"ariaLabel": ariaLabel,
		"items":     items,
	})
	if err != nil {
		return nil, fmt.Errorf("pages: rendering the language navigation: %w", err)
	}
	return map[string]any{
		"serviceName": serviceName,
		"serviceUrl":  serviceURL,
		"slots":       map[string]any{"end": languages},
	}, nil
}

func (r *Renderer) footer(lang string) map[string]any {
	items := []any{
		map[string]any{"href": "/help", "text": "Help"},
		map[string]any{"href": "/fees", "text": "Licence fees"},
		map[string]any{"href": "/updates", "text": "Service updates"},
		map[string]any{"href": "/guidance", "text": "Guidance"},
		map[string]any{"href": "/cookies", "text": "Cookies"},
		map[string]any{"href": "/accessibility", "text": "Accessibility"},
		map[string]any{"href": "/about", "text": "About this example"},
	}
	if r.options.DemosEnabled {
		items = append(items,
			map[string]any{"href": "/components", "text": "Component catalogue"},
			map[string]any{"href": "/examples", "text": "Example pages"},
		)
	}
	footer := map[string]any{"meta": map[string]any{"items": items}}
	if lang != LangCY {
		return footer
	}
	footer["contentLicence"] = map[string]any{
		"html": `Mae’r holl gynnwys ar gael dan <a class="govuk-footer__link" href="https://www.nationalarchives.gov.uk/doc/open-government-licence-cymraeg/version/3/" rel="license">Drwydded y Llywodraeth Agored v3.0</a>, ac eithrio lle nodir yn wahanol`,
	}
	footer["copyright"] = map[string]any{"html": "<span>Hawlfraint y Goron</span>"}
	return footer
}

func phaseBanner(lang string) map[string]any {
	if lang == LangCY {
		return map[string]any{
			"tag":  map[string]any{"text": "Enghraifft"},
			"html": `Mae hon yn wasanaeth enghreifftiol – bydd eich <a class="govuk-link" href="/about">adborth</a> yn ein helpu i wella’r gwasanaeth.`,
		}
	}
	return map[string]any{
		"tag":  map[string]any{"text": "Example"},
		"html": `This is an example service – your <a class="govuk-link" href="/about">feedback</a> will help us to improve it.`,
	}
}

// cookieBanner returns the banner to show, or nil when there is nothing to say.
//
// Once a choice is saved the banner disappears; the confirmation that replaces it is announced
// with role="alert" because it appears without the page reloading under it.
func cookieBanner(current *session.Session) map[string]any {
	switch current.CookieBanner {
	case session.ChoiceAccept:
		return confirmationBanner("You have accepted analytics cookies.")
	case session.ChoiceReject:
		return confirmationBanner("You have rejected analytics cookies.")
	}
	if current.CookieChoice != "" {
		return nil
	}
	return map[string]any{
		"messages": []any{map[string]any{
			"headingText": "Cookies on Apply for a rod fishing licence",
			"text":        "We use analytics cookies to understand how you use this example service. This example does not set analytics cookies.",
			"actions": []any{
				map[string]any{"text": "Accept analytics cookies", "type": "submit", "name": "cookies", "value": "accept"},
				map[string]any{"text": "Reject analytics cookies", "type": "submit", "name": "cookies", "value": "reject"},
				map[string]any{"text": "View cookies", "href": "/cookies"},
			},
		}},
	}
}

func confirmationBanner(text string) map[string]any {
	return map[string]any{
		"messages": []any{map[string]any{
			"text": text,
			"role": "alert",
			"actions": []any{
				map[string]any{"text": "Hide cookie message", "type": "submit", "name": "cookies", "value": "hide"},
			},
		}},
	}
}
