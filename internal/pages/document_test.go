package pages

import (
	"fmt"
	"html/template"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/Nooshu/govuk-frontend-example-go/internal/render"
	"github.com/Nooshu/govuk-frontend-example-go/internal/session"
)

func TestRenderDocuments(t *testing.T) {
	renderer := newTestRenderer(t, func(name string, params map[string]any) (string, error) {
		return "<div data-component=\"" + name + "\">" + fmt.Sprint(params) + "</div>", nil
	})

	english := session.New()
	html, err := renderer.Render(View{
		Template:     "start",
		Heading:      "Apply for a rod fishing licence",
		ShowFeedback: true,
		Context:      map[string]any{},
	}, english, "/start?from=home")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"lang=\"en\"", "Skip to main content", "returnPath\" value=\"/start?from=home\"", "<script>snippet</script>", "/assets/application.css"} {
		if !strings.Contains(html, want) {
			t.Errorf("english document missing %q", want)
		}
	}

	welsh := session.New()
	welshHTML, err := renderer.Render(View{
		Template: "fees",
		Heading:  "Ffioedd",
		Lang:     LangCY,
		Breadcrumbs: map[string]any{
			"items": []any{map[string]any{"text": "Ffioedd"}},
		},
		Context: map[string]any{},
	}, welsh, "//evil.example")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(welshHTML, "lang=\"cy\"") || !strings.Contains(welshHTML, "Neidio") {
		t.Fatal("welsh chrome missing")
	}
	if !strings.Contains(welshHTML, "Drwydded y Llywodraeth Agored") {
		t.Fatal("welsh footer licence missing")
	}
	if !strings.Contains(welshHTML, `value="/"`) {
		t.Fatal("off-site return path was not replaced")
	}

	withErrors := session.New()
	withErrors.CookieChoice = session.ChoiceAccept
	errorHTML, err := renderer.Render(View{
		Template:  "name",
		Heading:   "What is your name?",
		HasErrors: true,
		BackLink:  map[string]any{"text": "Back", "href": "/task-list"},
		Personal:  true,
		Context: map[string]any{
			"errorSummary": map[string]any{"titleText": "There is a problem"},
			"returnTo":     "check-answers",
			"firstName":    map[string]any{"name": "first-name"},
			"lastName":     map[string]any{"name": "last-name"},
		},
	}, withErrors, "/name")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(errorHTML, "Error: What is your name?") {
		t.Fatal("error title missing")
	}
	if strings.Contains(errorHTML, "cookie-banner") && strings.Contains(errorHTML, "Cookies on Apply") {
		t.Fatal("cookie banner shown after a choice was saved")
	}

	accept := session.New()
	accept.CookieBanner = session.ChoiceAccept
	if _, err := renderer.Render(View{Template: "fees", Heading: "Fees", Context: map[string]any{}}, accept, "/fees"); err != nil {
		t.Fatal(err)
	}
	reject := session.New()
	reject.CookieBanner = session.ChoiceReject
	if _, err := renderer.Render(View{Template: "fees", Heading: "Fees", Context: map[string]any{}}, reject, "/fees"); err != nil {
		t.Fatal(err)
	}

	address, err := renderer.Render(View{
		Template:     "address",
		Heading:      "Address",
		BackLink:     map[string]any{"href": "/start-month", "text": "Back"},
		ExitThisPage: map[string]any{"redirectUrl": "https://www.bbc.co.uk/weather"},
		Context: map[string]any{
			"fieldset": map[string]any{"legend": map[string]any{"text": "Address"}},
			"line1":    map[string]any{"name": "address-line-1"},
			"line2":    nil,
			"town":     map[string]any{"name": "town"},
			"postcode": map[string]any{"name": "postcode"},
			"enctype":  "multipart/form-data",
		},
	}, session.New(), "/address")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(address, "exit-this-page") || !strings.Contains(address, "multipart/form-data") {
		t.Fatal("address page did not compose its slots")
	}

	demos := newTestRenderer(t, func(string, map[string]any) (string, error) {
		return "<div></div>", nil
	})
	demos.options.DemosEnabled = true
	if _, err := demos.Render(View{Template: "about", Heading: "About", ShowFeedback: true, Context: map[string]any{}}, session.New(), "/about"); err != nil {
		t.Fatal(err)
	}
}

func TestRenderFailures(t *testing.T) {
	renderer := newTestRenderer(t, func(string, map[string]any) (string, error) {
		return "<div></div>", nil
	})
	_, err := renderer.Render(View{
		Template:    "fees",
		BackLink:    map[string]any{"href": "/"},
		Breadcrumbs: map[string]any{"items": []any{}},
	}, session.New(), "/")
	if err == nil || !strings.Contains(err.Error(), "both a back link and breadcrumbs") {
		t.Fatalf("both navigations: %v", err)
	}
	if _, err := renderer.Render(View{Template: "missing", Heading: "Missing"}, session.New(), "/"); err == nil {
		t.Fatal("missing template was rendered")
	}

	failing := newTestRenderer(t, func(name string, _ map[string]any) (string, error) {
		if name == "language-navigation" {
			return "", errStop
		}
		return "<div></div>", nil
	})
	if _, err := failing.Render(View{Template: "fees", Heading: "Fees", Context: map[string]any{}}, session.New(), "/fees"); err == nil {
		t.Fatal("language navigation failure was ignored")
	}

	broken := newTestRenderer(t, func(name string, _ map[string]any) (string, error) {
		if name == "skip-link" {
			return "", errStop
		}
		return "<div></div>", nil
	})
	if _, err := broken.Render(View{Template: "fees", Heading: "Fees", Context: map[string]any{}}, session.New(), "/fees"); err == nil {
		t.Fatal("skip link failure was ignored")
	}

	if _, err := renderer.component("button", "not a map"); err == nil {
		t.Fatal("non-map component options were accepted")
	}
	if html, err := renderer.component("button", nil); err != nil || html == "" {
		t.Fatalf("nil options: %s %v", html, err)
	}
	if _, err := renderer.component("button", map[string]any{"text": "Go"}); err != nil {
		t.Fatal(err)
	}
}

func TestTemplateHelpers(t *testing.T) {
	if _, err := dict("only"); err == nil {
		t.Fatal("odd dict was accepted")
	}
	if _, err := dict(1, "value"); err == nil {
		t.Fatal("non-string dict key was accepted")
	}
	built, err := dict("text", "Continue")
	if err != nil || built["text"] != "Continue" {
		t.Fatal(err)
	}
	if _, err := merge("nope", "text", "x"); err == nil {
		t.Fatal("non-map merge was accepted")
	}
	if _, err := merge(map[string]any{}, "only"); err == nil {
		t.Fatal("odd merge was accepted")
	}
	merged, err := merge(map[string]any{"legend": "Address"}, "html", "<input>")
	if err != nil || merged["html"] != "<input>" || merged["legend"] != "Address" {
		t.Fatalf("%#v %v", merged, err)
	}
	mergedNil, err := merge(nil, "html", "x")
	if err != nil || mergedNil["html"] != "x" {
		t.Fatal(err)
	}
	if concat() != "" || concat(template.HTML("<a>"), template.HTML("<b>")) != "<a><b>" {
		t.Fatal("concat mismatch")
	}
}

func newTestRenderer(t *testing.T, renderFn render.Func) *Renderer {
	t.Helper()
	renderer, err := NewRenderer(Options{
		Components: renderFn,
		Assets: func() Assets {
			return Assets{StylesheetHref: "/assets/application.css", AppModuleHref: "/assets/app.mjs"}
		},
		JSEnabledSnippet: "snippet",
		ServiceName:      "Apply for a rod fishing licence",
		ServiceNameCy:    "Gwneud cais am drwydded bysgota",
		FrontendVersion:  "6.5.1",
	})
	if err != nil {
		t.Fatal(err)
	}
	return renderer
}

func TestNewRendererRejectsBrokenTemplates(t *testing.T) {
	_, err := NewRenderer(Options{
		TemplateFS: fstest.MapFS{
			"templates/broken.gohtml": &fstest.MapFile{Data: []byte("{{")},
		},
	})
	if err == nil {
		t.Fatal("broken templates were accepted")
	}
}

type stopError struct{}

func (stopError) Error() string { return "stop" }

var errStop = stopError{}
