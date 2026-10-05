package app

// Page handlers and view models for the fishing rod licence journey and demos.

import (
	"net/http"

	"github.com/Nooshu/govuk-frontend-example-go/internal/components"
	"github.com/Nooshu/govuk-frontend-example-go/internal/govuk"
	"github.com/Nooshu/govuk-frontend-example-go/internal/pages"
	"github.com/Nooshu/govuk-frontend-example-go/internal/service"
	"github.com/Nooshu/govuk-frontend-example-go/internal/session"
)

// Each view builder decides what one page shows. They hold no HTML: a view names a content
// template and the component options that template needs.

func startView(lang string) pages.View {
	welsh := lang == pages.LangCY
	pick := func(english, cymraeg string) string {
		if welsh {
			return cymraeg
		}
		return english
	}
	return pages.View{
		Template:     "start",
		Status:       http.StatusOK,
		Heading:      pick("Apply for a fishing rod licence", "Gwneud cais am drwydded bysgota"),
		Lang:         lang,
		ShowFeedback: true,
		Context: map[string]any{
			"lede": pick(
				"Use this service to apply for a licence to fish with a rod.",
				"Defnyddiwch y gwasanaeth hwn i wneud cais am drwydded i bysgota gyda gwialen.",
			),
			"timing": pick("Applying takes about 10 minutes.", "Mae’n cymryd tua 10 munud."),
			"startButton": map[string]any{
				"text":          pick("Start now", "Dechrau nawr"),
				"href":          "/licence-length",
				"isStartButton": true,
			},
			"notification": map[string]any{
				"titleText": pick("Important", "Pwysig"),
				"text": pick(
					"The 2026 to 2027 rod licence is now available.",
					"Mae trwydded gwialen 2026 i 2027 ar gael nawr.",
				),
			},
			"warning": map[string]any{
				"text": pick(
					"You must have a valid rod licence before you fish.",
					"Rhaid i chi gael trwydded gwialen ddilys cyn i chi bysgota.",
				),
				"iconFallbackText": pick("Warning", "Rhybudd"),
			},
			"inset": map[string]any{
				"text": pick(
					"You need to be 13 or over. This example does not take payment.",
					"Mae gweddill yr enghraifft hon yn Saesneg.",
				),
			},
			"details": map[string]any{
				"summaryText": pick("What you will need", "Beth fydd ei angen arnoch"),
				"html": pick(
					`<ul class="govuk-list govuk-list--bullet"><li>How long you need the licence</li><li>Your name</li><li>Your date of birth</li><li>The country where you will fish</li><li>Your email address</li></ul>`,
					`<ul class="govuk-list govuk-list--bullet"><li>Pa mor hir mae angen y drwydded</li><li>Eich enw</li><li>Eich dyddiad geni</li><li>Y wlad lle byddwch yn pysgota</li><li>Eich cyfeiriad e-bost</li></ul>`,
				),
			},
		},
	}
}

// stepView shows one question.
//
// The back link goes to check-your-answers when the applicant came from there, so changing one
// answer returns them to their summary rather than walking them through the rest of the journey
// again. Otherwise it follows the linear journey, with the first question linking back to start.
func (a *App) stepView(step service.Step, current *session.Session, r *http.Request, errs []service.FieldError) pages.View {
	returnTo := ""
	if r.URL.Query().Get("return") == "check-answers" {
		returnTo = "check-answers"
	}
	back := "/"
	if returnTo != "" {
		back = "/check-answers"
	} else if previous, ok := service.PreviousStep(step.ID); ok {
		back = previous.Path
	}
	context := a.stepContext(step, current, errs)
	if summary, ok := service.ErrorSummary(errs); ok {
		context["errorSummary"] = summary
	}
	if returnTo != "" {
		context["returnTo"] = returnTo
	}
	return pages.View{
		Template:    string(step.ID),
		Status:      http.StatusOK,
		Heading:     step.Heading,
		HasErrors:   len(errs) > 0,
		Personal:    true,
		BackLink:    map[string]any{"text": "Back", "href": back},
		MainClasses: "govuk-main-wrapper--l",
		Context:     context,
	}
}

func (a *App) stepContext(step service.Step, current *session.Session, errs []service.FieldError) map[string]any {
	application := current.Application
	switch step.ID {
	case service.StepLicenceLength:
		return service.LicenceFields(application, errs)
	case service.StepName:
		return service.NameField(application, errs)
	case service.StepDateOfBirth:
		return service.DateField(application, errs)
	case service.StepWhereYouWillFish:
		return service.CountryFields(application, errs)
	default:
		return service.EmailField(application, errs)
	}
}

// checkAnswersGet shows the summary, but only once every question is answered and before the
// application is submitted.
func (a *App) checkAnswersGet(current *session.Session) outcome {
	if current.Application.Submitted {
		return outcome{redirect: "/confirmation"}
	}
	if incomplete, ok := service.FirstIncompleteStep(current.Application); ok {
		return outcome{redirect: incomplete.Path}
	}
	return outcome{view: &pages.View{
		Template:    "check-answers",
		Status:      http.StatusOK,
		Heading:     "Check your answers",
		BackLink:    map[string]any{"text": "Back", "href": "/email"},
		MainClasses: "govuk-main-wrapper--l",
		Personal:    true,
		Context:     map[string]any{"rows": service.SummaryRows(current.Application)},
	}}
}

func confirmationGet(current *session.Session) outcome {
	if !current.Application.Submitted {
		return outcome{redirect: "/"}
	}
	return outcome{view: &pages.View{
		Template:     "confirmation",
		Status:       http.StatusOK,
		Heading:      "Application complete",
		ShowFeedback: true,
		Personal:     true,
		Context:      map[string]any{"panel": service.ConfirmationPanel(current.Application.Reference)},
	}}
}

// updatesGet paginates the service updates page. An unknown page number redirects to page one
// rather than showing an empty list.
func updatesGet(r *http.Request) outcome {
	requested := r.URL.Query().Get("page")
	if requested != "" && requested != "1" && requested != "2" {
		return outcome{redirect: "/updates"}
	}
	page := 1
	if requested == "2" {
		page = 2
	}
	pagination := map[string]any{
		"items": []any{
			map[string]any{"number": 1, "href": "/updates", "current": page == 1},
			map[string]any{"number": 2, "href": "/updates?page=2", "current": page == 2},
		},
	}
	if page > 1 {
		pagination["previous"] = map[string]any{"href": "/updates"}
	}
	if page < 2 {
		pagination["next"] = map[string]any{"href": "/updates?page=2"}
	}
	body := "Example fees for the 2026 to 2027 season are on the fees page."
	if page == 2 {
		body = "There are no further fee changes planned in this example."
	}
	return outcome{view: &pages.View{
		Template:    "updates",
		Status:      http.StatusOK,
		Heading:     "Service updates",
		Breadcrumbs: crumbs("Service updates"),
		Context:     map[string]any{"body": body, "pagination": pagination},
	}}
}

func feesView() pages.View {
	return pages.View{
		Template:    "fees",
		Status:      http.StatusOK,
		Heading:     "Licence fees",
		Breadcrumbs: crumbs("Licence fees"),
		Context:     map[string]any{"table": service.FeesTable()},
	}
}

func helpView() pages.View {
	return pages.View{
		Template:     "help",
		Status:       http.StatusOK,
		Heading:      "Help",
		ShowFeedback: true,
		Breadcrumbs:  crumbs("Help"),
		Context:      map[string]any{"accordion": service.HelpAccordion()},
	}
}

func guidanceView() pages.View {
	return pages.View{
		Template:    "guidance",
		Status:      http.StatusOK,
		Heading:     "Guidance",
		Breadcrumbs: crumbs("Guidance"),
		Context:     map[string]any{"tabs": service.GuidanceTabs()},
	}
}

func (a *App) cookiesView(current *session.Session) pages.View {
	errs := a.errorsFor(current, "/cookies")
	notice := a.noticeFor(current, "/cookies")
	context := service.CookieFields(current.CookieChoice, errs)
	if summary, ok := service.ErrorSummary(errs); ok {
		context["errorSummary"] = summary
	}
	if notice != "" {
		context["notice"] = map[string]any{"type": "success", "titleText": "Success", "text": notice}
	}
	return pages.View{
		Template:    "cookies",
		Status:      http.StatusOK,
		Heading:     "Cookies",
		HasErrors:   len(errs) > 0,
		Personal:    true,
		Breadcrumbs: crumbs("Cookies"),
		Context:     context,
	}
}

func accessibilityView() pages.View {
	return pages.View{
		Template:     "accessibility",
		Status:       http.StatusOK,
		Heading:      "Accessibility statement",
		ShowFeedback: true,
		Breadcrumbs:  crumbs("Accessibility statement"),
		Context:      map[string]any{},
	}
}

func aboutView() pages.View {
	return pages.View{
		Template:     "about",
		Status:       http.StatusOK,
		Heading:      "About this example",
		ShowFeedback: true,
		Breadcrumbs:  crumbs("About this example"),
		Context:      map[string]any{},
	}
}

func (a *App) componentsView() (pages.View, error) {
	catalogue, err := a.library.Catalogue()
	if err != nil {
		return pages.View{}, err
	}
	return pages.View{
		Template:    "components",
		Status:      http.StatusOK,
		Heading:     "Component catalogue",
		Breadcrumbs: crumbs("Component catalogue"),
		Context:     map[string]any{"components": catalogue},
	}, nil
}

// renderComponentFixture is the Go port used by component previews. Tests replace it to
// exercise the error path without inventing a broken GOV.UK Frontend component.
var renderComponentFixture = govuk.Render

// componentView previews one fixture and says whether the HTML this service renders matches the
// official fixture. It returns nil when the component or the named fixture does not exist.
//
// Fixtures are loaded and rendered through [govuk.LoadFixtures] and [govuk.Render] — the same
// path as the parity suite — so attribute key order and JSON number spelling are preserved.
// Going through map[string]any would re-sort keys and turn numbers into float64, which falsely
// reports "HTML does not match the fixture" on the preview pages.
func (a *App) componentView(name, requested string) (*pages.View, error) {
	if !a.library.Has(name) {
		return nil, nil
	}
	set, err := govuk.LoadFixtures(a.config.ComponentsRoot, name)
	if err != nil {
		return nil, err
	}
	fixture, ok := selectGovukFixture(set.Fixtures, requested)
	if !ok {
		return nil, nil
	}
	rendered, err := renderComponentFixture(name, fixture.Options)
	if err != nil {
		return nil, err
	}
	info := components.Describe(name)
	return &pages.View{
		Template: "component",
		Status:   http.StatusOK,
		Heading:  info.Title,
		BackLink: map[string]any{"text": "Back", "href": "/components"},
		Context: map[string]any{
			"componentName":   name,
			"componentTitle":  info.Title,
			"designSystemUrl": info.DesignSystemURL,
			"description":     fixture.Description,
			"fixtureName":     fixture.Name,
			"rendered":        htmlOf(rendered),
			"parity":          components.ParityBanner(rendered == fixture.HTML),
			"fixtures":        govukFixtureLinks(set.Fixtures, fixture.Name),
		},
	}, nil
}

// selectGovukFixture chooses the fixture to preview from an ordered fixture set.
//
// A named fixture is used when it exists; otherwise the first visible fixture, falling back to
// the first fixture of all when every one is hidden.
func selectGovukFixture(fixtures []govuk.Fixture, requested string) (govuk.Fixture, bool) {
	if requested != "" {
		for _, fixture := range fixtures {
			if fixture.Name == requested {
				return fixture, true
			}
		}
		return govuk.Fixture{}, false
	}
	for _, fixture := range fixtures {
		if !fixture.Hidden {
			return fixture, true
		}
	}
	if len(fixtures) > 0 {
		return fixtures[0], true
	}
	return govuk.Fixture{}, false
}

func govukFixtureLinks(fixtures []govuk.Fixture, selected string) []fixtureLink {
	links := make([]fixtureLink, 0, len(fixtures))
	for _, fixture := range fixtures {
		links = append(links, fixtureLink{Name: fixture.Name, Current: fixture.Name == selected})
	}
	return links
}

func examplesView() pages.View {
	return pages.View{
		Template:    "examples",
		Status:      http.StatusOK,
		Heading:     "Example pages",
		Breadcrumbs: crumbs("Example pages"),
		Context:     map[string]any{},
	}
}

func exitThisPageView() pages.View {
	return pages.View{
		Template:     "exit-this-page",
		Status:       http.StatusOK,
		Heading:      "Exit this page",
		BackLink:     map[string]any{"text": "Back", "href": "/examples"},
		ExitThisPage: map[string]any{"redirectUrl": "https://www.bbc.co.uk/weather"},
		Context:      map[string]any{},
	}
}

func unavailableView() pages.View {
	return pages.View{
		Template: "unavailable",
		Status:   http.StatusOK,
		Heading:  "Sorry, the service is unavailable",
		BackLink: map[string]any{"text": "Back", "href": "/examples"},
		Context:  map[string]any{},
	}
}

func notFoundView() pages.View {
	return pages.View{
		Template: "not-found",
		Status:   http.StatusNotFound,
		Heading:  "Page not found",
		Context:  map[string]any{},
	}
}

func problemView(status int) pages.View {
	return pages.View{
		Template: "problem",
		Status:   status,
		Heading:  "Sorry, there is a problem with the service",
		Context:  map[string]any{},
	}
}

func sessionExpiredView() pages.View {
	return pages.View{
		Template: "session-expired",
		Status:   http.StatusForbidden,
		Heading:  "Sorry, your session has expired",
		Context:  map[string]any{},
	}
}

func crumbs(current string) map[string]any {
	return map[string]any{
		"items": []any{
			map[string]any{"href": "/", "text": "Home"},
			map[string]any{"text": current},
		},
	}
}
