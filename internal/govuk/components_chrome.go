package govuk

import "strings"

// This file ports the page chrome: the GOV.UK logo, the header, the footer, the service and
// language navigation, breadcrumbs, pagination, and the cookie banner.

// logoCrown and logoLogotype are the two halves of the GOV.UK logo, kept with the indentation
// they have in GOV.UK Frontend's logo.njk so the trim-and-indent below matches the macro.
const logoCrown = `    <g>
      <circle cx="20" cy="17.6" r="3.7"/>
      <circle cx="10.2" cy="23.5" r="3.7"/>
      <circle cx="3.7" cy="33.2" r="3.7"/>
      <circle cx="31.7" cy="30.6" r="3.7"/>
      <circle cx="43.3" cy="17.6" r="3.7"/>
      <circle cx="53.2" cy="23.5" r="3.7"/>
      <circle cx="59.7" cy="33.2" r="3.7"/>
      <circle cx="31.7" cy="30.6" r="3.7"/>
      <path d="M33.1,9.8c.2-.1.3-.3.5-.5l4.6,2.4v-6.8l-4.6,1.5c-.1-.2-.3-.3-.5-.5l1.9-5.9h-6.7l1.9,5.9c-.2.1-.3.3-.5.5l-4.6-1.5v6.8l4.6-2.4c.1.2.3.3.5.5l-2.6,8c-.9,2.8,1.2,5.7,4.1,5.7h0c3,0,5.1-2.9,4.1-5.7l-2.6-8ZM37,37.9s-3.4,3.8-4.1,6.1c2.2,0,4.2-.5,6.4-2.8l-.7,8.5c-2-2.8-4.4-4.1-5.7-3.8.1,3.1.5,6.7,5.8,7.2,3.7.3,6.7-1.5,7-3.8.4-2.6-2-4.3-3.7-1.6-1.4-4.5,2.4-6.1,4.9-3.2-1.9-4.5-1.8-7.7,2.4-10.9,3,4,2.6,7.3-1.2,11.1,2.4-1.3,6.2,0,4,4.6-1.2-2.8-3.7-2.2-4.2.2-.3,1.7.7,3.7,3,4.2,1.9.3,4.7-.9,7-5.9-1.3,0-2.4.7-3.9,1.7l2.4-8c.6,2.3,1.4,3.7,2.2,4.5.6-1.6.5-2.8,0-5.3l5,1.8c-2.6,3.6-5.2,8.7-7.3,17.5-7.4-1.1-15.7-1.7-24.5-1.7h0c-8.8,0-17.1.6-24.5,1.7-2.1-8.9-4.7-13.9-7.3-17.5l5-1.8c-.5,2.5-.6,3.7,0,5.3.8-.8,1.6-2.3,2.2-4.5l2.4,8c-1.5-1-2.6-1.7-3.9-1.7,2.3,5,5.2,6.2,7,5.9,2.3-.4,3.3-2.4,3-4.2-.5-2.4-3-3.1-4.2-.2-2.2-4.6,1.6-6,4-4.6-3.7-3.7-4.2-7.1-1.2-11.1,4.2,3.2,4.3,6.4,2.4,10.9,2.5-2.8,6.3-1.3,4.9,3.2-1.8-2.7-4.1-1-3.7,1.6.3,2.3,3.3,4.1,7,3.8,5.4-.5,5.7-4.2,5.8-7.2-1.3-.2-3.7,1-5.7,3.8l-.7-8.5c2.2,2.3,4.2,2.7,6.4,2.8-.7-2.3-4.1-6.1-4.1-6.1h10.6,0Z"/>
    </g>`

const logoLogotype = `    <circle class="govuk-logo-dot" cx="226" cy="36" r="7.3"/>
    <path d="M93.94 41.25c.4 1.81 1.2 3.21 2.21 4.62 1 1.4 2.21 2.41 3.61 3.21s3.21 1.2 5.22 1.2 3.61-.4 4.82-1c1.4-.6 2.41-1.4 3.21-2.41.8-1 1.4-2.01 1.61-3.01s.4-2.01.4-3.01v.14h-10.86v-7.02h20.07v24.08h-8.03v-5.56c-.6.8-1.38 1.61-2.19 2.41-.8.8-1.81 1.2-2.81 1.81-1 .4-2.21.8-3.41 1.2s-2.41.4-3.81.4a18.56 18.56 0 0 1-14.65-6.63c-1.6-2.01-3.01-4.41-3.81-7.02s-1.4-5.62-1.4-8.83.4-6.02 1.4-8.83a20.45 20.45 0 0 1 19.46-13.65c3.21 0 4.01.2 5.82.8 1.81.4 3.61 1.2 5.02 2.01 1.61.8 2.81 2.01 4.01 3.21s2.21 2.61 2.81 4.21l-7.63 4.41c-.4-1-1-1.81-1.61-2.61-.6-.8-1.4-1.4-2.21-2.01-.8-.6-1.81-1-2.81-1.4-1-.4-2.21-.4-3.61-.4-2.01 0-3.81.4-5.22 1.2-1.4.8-2.61 1.81-3.61 3.21s-1.61 2.81-2.21 4.62c-.4 1.81-.6 3.71-.6 5.42s.8 5.22.8 5.22Zm57.8-27.9c3.21 0 6.22.6 8.63 1.81 2.41 1.2 4.82 2.81 6.62 4.82S170.2 24.39 171 27s1.4 5.62 1.4 8.83-.4 6.02-1.4 8.83-2.41 5.02-4.01 7.02-4.01 3.61-6.62 4.82-5.42 1.81-8.63 1.81-6.22-.6-8.63-1.81-4.82-2.81-6.42-4.82-3.21-4.41-4.01-7.02-1.4-5.62-1.4-8.83.4-6.02 1.4-8.83 2.41-5.02 4.01-7.02 4.01-3.61 6.42-4.82 5.42-1.81 8.63-1.81Zm0 36.73c1.81 0 3.61-.4 5.02-1s2.61-1.81 3.61-3.01 1.81-2.81 2.21-4.41c.4-1.81.8-3.61.8-5.62 0-2.21-.2-4.21-.8-6.02s-1.2-3.21-2.21-4.62c-1-1.2-2.21-2.21-3.61-3.01s-3.21-1-5.02-1-3.61.4-5.02 1c-1.4.8-2.61 1.81-3.61 3.01s-1.81 2.81-2.21 4.62c-.4 1.81-.8 3.61-.8 5.62 0 2.41.2 4.21.8 6.02.4 1.81 1.2 3.21 2.21 4.41s2.21 2.21 3.61 3.01c1.4.8 3.21 1 5.02 1Zm36.32 7.96-12.24-44.15h9.83l8.43 32.77h.4l8.23-32.77h9.83L200.3 58.04h-12.24Zm74.14-7.96c2.18 0 3.51-.6 3.51-.6 1.2-.6 2.01-1 2.81-1.81s1.4-1.81 1.81-2.81a13 13 0 0 0 .8-4.01V13.9h8.63v28.15c0 2.41-.4 4.62-1.4 6.62-.8 2.01-2.21 3.61-3.61 5.02s-3.41 2.41-5.62 3.21-4.62 1.2-7.02 1.2-5.02-.4-7.02-1.2c-2.21-.8-4.01-1.81-5.62-3.21s-2.81-3.01-3.61-5.02-1.4-4.21-1.4-6.62V13.9h8.63v26.95c0 1.61.2 3.01.8 4.01.4 1.2 1.2 2.21 2.01 2.81.8.8 1.81 1.4 2.81 1.81 0 0 1.34.6 3.51.6Zm34.22-36.18v18.92l15.65-18.92h10.82l-15.03 17.32 16.03 26.83h-10.21l-11.44-20.21-5.62 6.22v13.99h-8.83V13.9"/>`

// renderLogo is GOV.UK Frontend's private govukLogo macro. It is not a component in its own
// right; the header and the footer both embed it.
//
// The output starts with a newline because the upstream macro does, and the footer relies on
// that when it splices the crown in without trimming.
func renderLogo(p *Params) string {
	useLogotype := truthy(def(p.Get("useLogotype"), true))
	width := "32"
	viewBox := "64"
	if useLogotype {
		width = "162"
		viewBox = "324"
	}
	role := "presentation"
	ariaLabel := p.Get("ariaLabelText")
	if truthy(ariaLabel) {
		role = "img"
	}

	var out_ strings.Builder
	out_.WriteString("\n  <svg\n    focusable=\"false\"\n    role=\"" + role + "\"\n" +
		"    xmlns=\"http://www.w3.org/2000/svg\"\n" +
		`    viewBox="0 0 ` + viewBox + " 60\"\n    height=\"30\"\n    width=\"" + width + "\"\n" +
		"    fill=\"currentcolor\"" +
		attributeIf("class", p.Get("classes")) +
		attributeIf("aria-label", ariaLabel) +
		Attributes(p.Get("attributes")) + "\n  >")
	if truthy(ariaLabel) {
		out_.WriteString("<title>" + out(ariaLabel) + "</title>")
	}
	out_.WriteString("    " + indent(trim(logoCrown), 2, false) + "\n")
	if useLogotype {
		out_.WriteString("      " + indent(trim(logoLogotype), 2, false) + "\n")
	}
	out_.WriteString("  </svg>\n")
	return out_.String()
}

func renderGenericHeader(p *Params) string {
	namespace := out(def(p.Get("_namespace"), "govuk-generic"))
	return `<div class="` + namespace + `-header` + classesIf(p.Get("classes")) + `"` +
		Attributes(p.Get("attributes")) + ">\n" +
		`  <div class="` + namespace + `-header__container ` +
		out(defTruthy(p.Get("containerClasses"), "govuk-width-container")) + "\">\n" +
		`    <div class="` + namespace + "-header__logo\">\n" +
		`      <a href="` + out(defTruthy(p.Get("url"), "/")) + `" class="` + namespace +
		"-header__homepage-link\">\n" +
		"        " + content(p, "logoHtml", "logoText") + "\n" +
		"      </a>\n    </div>\n  </div>\n</div>"
}

func renderHeader(p *Params) string {
	logo := renderLogo(NewParams(
		"classes", "govuk-header__logotype",
		"ariaLabelText", "GOV.UK",
	))
	logoContent := "  " + trim(logo) + "\n"
	if productName := p.Get("productName"); truthy(productName) {
		logoContent += `<span class="govuk-header__product-name">` + out(productName) + "</span>"
	}

	return renderGenericHeader(NewParams(
		"_namespace", "govuk",
		"logoHtml", Safe(indent(logoContent, 8, false)),
		"url", defTruthy(p.Get("homepageUrl"), "//gov.uk"),
		"containerClasses", p.Get("containerClasses"),
		"classes", p.Get("classes"),
		"attributes", p.Get("attributes"),
	))
}

// footerLicenceLogo is the Open Government Licence crest. Like the GOV.UK logo it is inlined
// rather than fetched, so the footer needs no extra request.
const footerLicenceLogo = `<svg
            aria-hidden="true"
            focusable="false"
            class="govuk-footer__licence-logo"
            xmlns="http://www.w3.org/2000/svg"
            viewBox="0 0 483.2 195.7"
            height="17"
            width="41"
          >
            <path
              fill="currentColor"
              d="M421.5 142.8V.1l-50.7 32.3v161.1h112.4v-50.7zm-122.3-9.6A47.12 47.12 0 0 1 221 97.8c0-26 21.1-47.1 47.1-47.1 16.7 0 31.4 8.7 39.7 21.8l42.7-27.2A97.63 97.63 0 0 0 268.1 0c-36.5 0-68.3 20.1-85.1 49.7A98 98 0 0 0 97.8 0C43.9 0 0 43.9 0 97.8s43.9 97.8 97.8 97.8c36.5 0 68.3-20.1 85.1-49.7a97.76 97.76 0 0 0 149.6 25.4l19.4 22.2h3v-87.8h-80l24.3 27.5zM97.8 145c-26 0-47.1-21.1-47.1-47.1s21.1-47.1 47.1-47.1 47.2 21 47.2 47S123.8 145 97.8 145"
            />
          </svg>`

const footerCopyrightHref = "https://www.nationalarchives.gov.uk/information-management/re-using-public-sector-information/uk-government-licensing-framework/crown-copyright/"

func renderFooter(p *Params) string {
	var out_ strings.Builder
	out_.WriteString(`<div class="govuk-footer` + classesIf(p.Get("classes")) + `"` +
		Attributes(p.Get("attributes")) + ">\n")
	out_.WriteString(`  <div class="govuk-width-container` +
		classesIf(p.Get("containerClasses")) + `">`)
	out_.WriteString(renderLogo(NewParams("classes", "govuk-footer__crown", "useLogotype", false)))
	out_.WriteString("\n")

	if navigation := items(p.Get("navigation")); len(navigation) > 0 {
		out_.WriteString("      <div class=\"govuk-footer__navigation\">\n")
		for _, nav := range navigation {
			out_.WriteString(`          <div class="govuk-footer__section govuk-grid-column-` +
				out(defTruthy(get(nav, "width"), "full")) + "\">\n")
			out_.WriteString(`            <h2 class="govuk-footer__heading govuk-heading-m">` +
				out(get(nav, "title")) + "</h2>\n")
			if links := items(get(nav, "items")); len(links) > 0 {
				listClasses := ""
				if columns := get(nav, "columns"); truthy(columns) {
					listClasses = " govuk-footer__list--columns-" + escape(str(columns))
				}
				out_.WriteString(`              <ul class="govuk-footer__list` + listClasses + "\">\n")
				for _, link := range links {
					if !truthy(get(link, "href")) || !truthy(get(link, "text")) {
						continue
					}
					out_.WriteString("                    <li class=\"govuk-footer__list-item\">\n")
					out_.WriteString(`                      <a class="govuk-footer__link" href="` +
						out(get(link, "href")) + `"` + Attributes(get(link, "attributes")) + ">\n")
					out_.WriteString("                        " + out(get(link, "text")) + "\n")
					out_.WriteString("                      </a>\n                    </li>\n")
				}
				out_.WriteString("              </ul>\n")
			}
			out_.WriteString("          </div>\n")
		}
		out_.WriteString("      </div>\n")
		out_.WriteString("      <hr class=\"govuk-footer__section-break\">\n")
	}

	out_.WriteString("    <div class=\"govuk-footer__meta\">\n")
	out_.WriteString("      <div class=\"govuk-footer__meta-item govuk-footer__meta-item--grow\">\n")

	if meta := p.Get("meta"); truthy(meta) {
		out_.WriteString(`        <h2 class="govuk-visually-hidden">` +
			out(defTruthy(get(meta, "visuallyHiddenTitle"), "Support links")) + "</h2>\n")
		if links := items(get(meta, "items")); len(links) > 0 {
			out_.WriteString("        <ul class=\"govuk-footer__inline-list\">\n")
			for _, link := range links {
				out_.WriteString("          <li class=\"govuk-footer__inline-list-item\">\n")
				out_.WriteString(`            <a class="govuk-footer__link" href="` +
					out(get(link, "href")) + `"` + Attributes(get(link, "attributes")) + ">\n")
				out_.WriteString("              " + out(get(link, "text")) + "\n")
				out_.WriteString("            </a>\n          </li>\n")
			}
			out_.WriteString("        </ul>\n")
		}
		if truthy(get(meta, "text")) || truthy(get(meta, "html")) {
			out_.WriteString("        <div class=\"govuk-footer__meta-custom\">\n")
			out_.WriteString("          " + contentIndent(meta, "html", "text", 10) + "\n")
			out_.WriteString("        </div>\n")
		}
	}

	// A caller can drop the licence statement entirely by passing an explicit null, which is
	// why this tests for null rather than for falsiness.
	if licence := p.Get("contentLicence"); licence != nil {
		out_.WriteString("          " + footerLicenceLogo + "\n")
		out_.WriteString("          <span class=\"govuk-footer__licence-description\">\n")
		if truthy(get(licence, "html")) || truthy(get(licence, "text")) {
			out_.WriteString("            " + contentIndent(licence, "html", "text", 12) + "\n")
		} else {
			out_.WriteString("            All content is available under the\n" +
				"            <a\n" +
				"              class=\"govuk-footer__link\"\n" +
				"              href=\"https://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/\"\n" +
				"              rel=\"license\"\n" +
				"            >Open Government Licence v3.0</a>, except where otherwise stated\n")
		}
		out_.WriteString("          </span>\n")
	}

	out_.WriteString("      </div>\n")
	out_.WriteString("      <div class=\"govuk-footer__meta-item\">\n")
	out_.WriteString("        <a\n" +
		"          class=\"govuk-footer__link govuk-footer__copyright-logo\"\n" +
		`          href="` + footerCopyrightHref + "\"\n        >\n")
	if copyright := p.Get("copyright"); truthy(get(copyright, "html")) || truthy(get(copyright, "text")) {
		out_.WriteString("          " + contentIndent(copyright, "html", "text", 10) + "\n")
	} else {
		out_.WriteString("          \u00a9 Crown copyright\n")
	}
	out_.WriteString("        </a>\n      </div>\n")
	out_.WriteString("    </div>\n  </div>\n</div>")
	return out_.String()
}

func renderBreadcrumbs(p *Params) string {
	classNames := "govuk-breadcrumbs"
	if classes := p.Get("classes"); truthy(classes) {
		classNames += " " + str(classes)
	}
	if truthy(p.Get("collapseOnMobile")) {
		classNames += " govuk-breadcrumbs--collapse-on-mobile"
	}

	var out_ strings.Builder
	out_.WriteString(`<nav class="` + escape(classNames) + `"` + Attributes(p.Get("attributes")) +
		` aria-label="` + out(def(p.Get("labelText"), "Breadcrumb")) + "\">\n")
	out_.WriteString("  <ol class=\"govuk-breadcrumbs__list\">\n")
	for _, item := range items(p.Get("items")) {
		if href := get(item, "href"); truthy(href) {
			out_.WriteString("    <li class=\"govuk-breadcrumbs__list-item\">\n")
			out_.WriteString(`      <a class="govuk-breadcrumbs__link" href="` + out(href) + `"` +
				Attributes(get(item, "attributes")) + ">" + content(item, "html", "text") + "</a>\n")
			out_.WriteString("    </li>\n")
		} else {
			out_.WriteString(`    <li class="govuk-breadcrumbs__list-item" aria-current="page">` +
				content(item, "html", "text") + "</li>\n")
		}
	}
	out_.WriteString("  </ol>\n</nav>")
	return out_.String()
}

func renderLanguageNavigation(p *Params) string {
	var out_ strings.Builder
	out_.WriteString(`<nav class="govuk-language-navigation` + classesIf(p.Get("classes")) + `"` +
		Attributes(p.Get("attributes")) + ` aria-label="` +
		out(def(p.Get("ariaLabel"), "Language")) + "\">\n")
	out_.WriteString("  <ul class=\"govuk-language-navigation__list\">\n")

	for _, item := range items(p.Get("items")) {
		href := get(item, "href")
		out_.WriteString("    <li class=\"govuk-language-navigation__list-item\">\n")
		if truthy(get(item, "current")) || !truthy(href) {
			out_.WriteString(`      <span class="govuk-language-navigation__text` +
				classesIf(get(item, "classes")) + "\"\n        aria-current=\"true\"" +
				attributeIf("lang", get(item, "lang")) +
				attributeIf("dir", get(item, "dir")) +
				Attributes(get(item, "attributes")) + ">" +
				content(item, "html", "text") + "</span>\n")
		} else {
			hrefLang := get(item, "hrefLang")
			if !truthy(hrefLang) {
				hrefLang = get(item, "lang")
			}
			out_.WriteString(`      <a class="govuk-language-navigation__link` +
				classesIf(get(item, "classes")) + `" href="` + out(href) + `" rel="alternate"` +
				attributeIf("lang", get(item, "lang")) +
				attributeIf("hreflang", hrefLang) +
				attributeIf("dir", get(item, "dir")) +
				Attributes(get(item, "attributes")) + ">" +
				content(item, "html", "text"))
			if description := get(item, "languageDescriptionText"); truthy(description) {
				out_.WriteString(`<span class="govuk-visually-hidden"> ` + out(description) + "</span>")
			}
			out_.WriteString("      </a>\n")
		}
		out_.WriteString("    </li>\n")
	}

	out_.WriteString("  </ul>\n</nav>")
	return out_.String()
}

func renderServiceNavigation(p *Params) string {
	slots := p.Get("slots")
	menuButtonText := defTruthy(p.Get("menuButtonText"), "Menu")
	navigationID := out(defTruthy(p.Get("navigationId"), "navigation"))

	endSlot := get(slots, "end")
	endSlotHTML := get(endSlot, "html")
	if text, ok := endSlot.(string); ok {
		endSlotHTML = text
	}
	endSlotObject, endSlotIsObject := endSlot.(*Params)
	endSlotInline := endSlotIsObject && looseEq(endSlotObject.Get("align"), "inline")

	commonAttributes := `class="govuk-service-navigation` + classesIf(p.Get("classes")) + "\"\n" +
		"data-module=\"govuk-service-navigation\"" + Attributes(p.Get("attributes")) + "\n"

	var inner strings.Builder
	inner.WriteString(`  <div class="govuk-width-container` +
		flagIf(" govuk-service-navigation__inlining-container", endSlotInline) + "\">\n\n    ")
	if start := get(slots, "start"); truthy(start) {
		inner.WriteString(str(start))
	}
	inner.WriteString("<div class=\"govuk-service-navigation__container\">\n      \n")

	if serviceName := p.Get("serviceName"); truthy(serviceName) {
		inner.WriteString("        <span class=\"govuk-service-navigation__service-name\">\n")
		if serviceURL := p.Get("serviceUrl"); truthy(serviceURL) {
			inner.WriteString(`            <a href="` + out(serviceURL) +
				"\" class=\"govuk-service-navigation__link\">\n")
			inner.WriteString("              " + out(serviceName) + "\n            </a>\n")
		} else {
			inner.WriteString(`            <span class="govuk-service-navigation__text">` +
				out(serviceName) + "</span>\n")
		}
		inner.WriteString("        </span>\n")
	}
	inner.WriteString("\n      \n")

	navigationItems := []any{}
	for _, item := range items(p.Get("navigation")) {
		if truthy(item) {
			navigationItems = append(navigationItems, item)
		}
	}
	collapse := truthy(def(p.Get("collapseNavigationOnMobile"), len(navigationItems) > 1))

	navigationStart := get(slots, "navigationStart")
	navigationEnd := get(slots, "navigationEnd")
	if len(navigationItems) > 0 || truthy(navigationStart) || truthy(navigationEnd) {
		inner.WriteString(`        <nav aria-label="` +
			out(defTruthy(p.Get("navigationLabel"), menuButtonText)) +
			`" class="govuk-service-navigation__wrapper` +
			classesIf(p.Get("navigationClasses")) + "\">\n")
		if collapse {
			menuButtonLabel := p.Get("menuButtonLabel")
			ariaLabel := ""
			if truthy(menuButtonLabel) && !looseEq(menuButtonLabel, menuButtonText) {
				ariaLabel = ` aria-label="` + out(menuButtonLabel) + `"`
			}
			inner.WriteString(`          <button type="button" class="govuk-service-navigation__toggle govuk-js-service-navigation-toggle" aria-controls="` +
				navigationID + `"` + ariaLabel + " hidden aria-hidden=\"true\">\n")
			inner.WriteString("            " + out(menuButtonText) + "\n          </button>\n")
		}
		inner.WriteString("\n          <ul class=\"govuk-service-navigation__list\" id=\"" +
			navigationID + "\" >\n\n            ")
		if truthy(navigationStart) {
			inner.WriteString(str(navigationStart))
		}
		inner.WriteString("\n")

		for _, item := range navigationItems {
			active := truthy(get(item, "active")) || truthy(get(item, "current"))
			// Nunjucks captures the {% set linkInnerContent %} block with its leading blank
			// line and indentation; replicate that so fixtures match byte-for-byte.
			var linkInner string
			if active {
				linkInner = "\n                                    \n                  <strong class=\"govuk-service-navigation__active-fallback\">" +
					content(item, "html", "text") + "</strong>\n"
			} else {
				linkInner = "\n                                    \n" + content(item, "html", "text")
			}

			ariaCurrent := ""
			if active {
				value := "true"
				if truthy(get(item, "current")) {
					value = "page"
				}
				ariaCurrent = ` aria-current="` + value + `"`
			}

			inner.WriteString("              \n")
			inner.WriteString(`              <li class="govuk-service-navigation__item` +
				flagIf(" govuk-service-navigation__item--active", active) + "\">\n")
			switch href := get(item, "href"); {
			case truthy(href):
				inner.WriteString(`                  <a class="govuk-service-navigation__link" href="` +
					out(href) + `"` + ariaCurrent + Attributes(get(item, "attributes")) +
					">" + linkInner + "\n                  </a>\n")
			case truthy(get(item, "html")) || truthy(get(item, "text")):
				inner.WriteString(`                  <span class="govuk-service-navigation__text"` +
					ariaCurrent + ">" + linkInner + "\n                  </span>\n")
			}
			inner.WriteString("              </li>\n\n")
		}

		inner.WriteString("            ")
		if truthy(navigationEnd) {
			inner.WriteString(str(navigationEnd))
		}
		inner.WriteString("</ul>\n        </nav>\n")
	}

	inner.WriteString("    </div>\n\n    ")
	if truthy(endSlotHTML) {
		inner.WriteString(str(endSlotHTML))
	}
	inner.WriteString("</div>\n")

	if truthy(p.Get("serviceName")) || truthy(get(slots, "start")) || truthy(endSlotHTML) {
		return `  <section aria-label="` + out(def(p.Get("ariaLabel"), "Service information")) +
			`" ` + commonAttributes + ">\n    " + inner.String() + "\n  </section>\n"
	}
	return "  <div " + commonAttributes + ">\n    " + inner.String() + "\n  </div>\n"
}

const paginationArrowPrevious = `  <svg class="govuk-pagination__icon govuk-pagination__icon--prev" xmlns="http://www.w3.org/2000/svg" height="13" width="15" aria-hidden="true" focusable="false" viewBox="0 0 15 13">
    <path d="m6.5938-0.0078125-6.7266 6.7266 6.7441 6.4062 1.377-1.449-4.1856-3.9768h12.896v-2h-12.984l4.2931-4.293-1.414-1.414z"></path>
  </svg>`

const paginationArrowNext = `  <svg class="govuk-pagination__icon govuk-pagination__icon--next" xmlns="http://www.w3.org/2000/svg" height="13" width="15" aria-hidden="true" focusable="false" viewBox="0 0 15 13">
    <path d="m8.107-0.0078125-1.4136 1.414 4.2926 4.293h-12.986v2h12.896l-4.1855 3.9766 1.377 1.4492 6.7441-6.4062-6.7246-6.7266z"></path>
  </svg>`

func renderPagination(p *Params) string {
	previous, next := p.Get("previous"), p.Get("next")
	// Pagination switches to the stacked "block" layout when it has only previous and next
	// links rather than a list of page numbers.
	blockLevel := !truthy(p.Get("items")) && (truthy(next) || truthy(previous))

	var out_ strings.Builder
	out_.WriteString(`<nav class="govuk-pagination` +
		flagIf(" govuk-pagination--block", blockLevel) + classesIf(p.Get("classes")) +
		`" aria-label="` + out(defTruthy(p.Get("landmarkLabel"), "Pagination")) + `"` +
		Attributes(p.Get("attributes")) + ">\n")

	if truthy(previous) && truthy(get(previous, "href")) {
		out_.WriteString(paginationArrowLink(previous, "prev", blockLevel,
			paginationLinkLabel(previous, "Previous")))
	}

	if entries := p.Get("items"); truthy(entries) {
		out_.WriteString("  <ul class=\"govuk-pagination__list\">\n")
		for _, item := range items(entries) {
			if item == nil || length(item) == 0 {
				continue
			}
			out_.WriteString("      " + indent(paginationPageItem(item), 2, false) + "\n")
		}
		out_.WriteString("  </ul>\n")
	}

	if truthy(next) && truthy(get(next, "href")) {
		out_.WriteString(paginationArrowLink(next, "next", blockLevel,
			paginationLinkLabel(next, "Next")))
	}

	out_.WriteString("</nav>")
	return out_.String()
}

// paginationLinkLabel is the body of the `{% call %}` block upstream: the caller's text or
// HTML, or the default label with the visually hidden "page" suffix.
func paginationLinkLabel(link any, fallback string) string {
	html, text := get(link, "html"), get(link, "text")
	switch {
	case truthy(html):
		return trim(indent(trim(str(html)), 8, false))
	case truthy(text):
		return out(text)
	default:
		return fallback + `<span class="govuk-visually-hidden"> page</span>`
	}
}

func paginationArrowLink(link any, kind string, blockLevel bool, label string) string {
	arrow := paginationArrowNext
	if kind == "prev" {
		arrow = paginationArrowPrevious
	}

	var out_ strings.Builder
	out_.WriteString(`  <div class="govuk-pagination__` + kind + "\">\n")
	out_.WriteString(`    <a class="govuk-link govuk-pagination__link" href="` +
		out(get(link, "href")) + `" rel="` + kind + `"` +
		Attributes(get(link, "attributes")) + ">\n")
	if blockLevel || kind == "prev" {
		out_.WriteString(indent(arrow, 4, true) + "\n")
	}
	labelText := get(link, "labelText")
	out_.WriteString(`      <span class="govuk-pagination__link-title` +
		flagIf(" govuk-pagination__link-title--decorated", blockLevel && !truthy(labelText)) +
		"\">\n        " + label + "\n      </span>\n")
	if truthy(labelText) && blockLevel {
		out_.WriteString("      <span class=\"govuk-visually-hidden\">:</span>\n")
		out_.WriteString(`      <span class="govuk-pagination__link-label">` +
			out(labelText) + "</span>\n")
	}
	if !blockLevel && kind == "next" {
		out_.WriteString(indent(arrow, 4, true) + "\n")
	}
	out_.WriteString("    </a>\n  </div>\n")
	return out_.String()
}

func paginationPageItem(item any) string {
	var out_ strings.Builder
	out_.WriteString(`<li class="govuk-pagination__item` +
		flagIf(" govuk-pagination__item--current", get(item, "current")) +
		flagIf(" govuk-pagination__item--ellipsis", get(item, "ellipsis")) + "\">\n")
	if truthy(get(item, "ellipsis")) {
		out_.WriteString("    &ctdot;\n")
	} else {
		out_.WriteString(`    <a class="govuk-link govuk-pagination__link" href="` +
			out(get(item, "href")) + `" aria-label="` +
			out(def(get(item, "visuallyHiddenText"), "Page "+str(get(item, "number")))) + `"` +
			flagIf(` aria-current="page"`, get(item, "current")) +
			Attributes(get(item, "attributes")) + ">\n")
		out_.WriteString("      " + out(get(item, "number")) + "\n    </a>\n")
	}
	out_.WriteString("  </li>")
	return out_.String()
}

func renderCookieBanner(p *Params) string {
	var out_ strings.Builder
	out_.WriteString(`<div class="govuk-cookie-banner` + classesIf(p.Get("classes")) +
		`" data-nosnippet role="region" aria-label="` +
		out(defTruthy(p.Get("ariaLabel"), "Cookie banner")) + `"` +
		flagIf(" hidden", p.Get("hidden")) + Attributes(p.Get("attributes")) + ">\n")

	for _, message := range items(p.Get("messages")) {
		out_.WriteString(`  <div class="govuk-cookie-banner__message` +
			classesIf(get(message, "classes")) + ` govuk-width-container"` +
			attributeIf("role", get(message, "role")) +
			Attributes(get(message, "attributes")) +
			flagIf(" hidden", get(message, "hidden")) + ">\n\n")
		out_.WriteString("    <div class=\"govuk-grid-row\">\n")
		out_.WriteString("      <div class=\"govuk-grid-column-two-thirds\">\n")
		if truthy(get(message, "headingHtml")) || truthy(get(message, "headingText")) {
			out_.WriteString("        <h2 class=\"govuk-cookie-banner__heading govuk-heading-m\">\n")
			out_.WriteString("          " +
				contentIndent(message, "headingHtml", "headingText", 10) + "\n")
			out_.WriteString("        </h2>\n")
		}
		out_.WriteString("        <div class=\"govuk-cookie-banner__content\">\n")
		switch html, text := get(message, "html"), get(message, "text"); {
		case truthy(html):
			out_.WriteString("          " + indent(trim(str(html)), 10, false) + "\n")
		case truthy(text):
			out_.WriteString(`          <p class="govuk-body">` + out(text) + "</p>\n")
		}
		out_.WriteString("        </div>\n      </div>\n    </div>\n\n")

		if actions := items(get(message, "actions")); actions != nil {
			out_.WriteString("    <div class=\"govuk-button-group\">\n")
			for _, action := range actions {
				out_.WriteString("      " + indent(trim(cookieBannerAction(action)), 6, false) + "\n")
			}
			out_.WriteString("    </div>\n")
		}

		out_.WriteString("\n  </div>\n")
	}

	out_.WriteString("</div>")
	return out_.String()
}

func cookieBannerAction(action any) string {
	href := get(action, "href")
	if !truthy(href) || str(get(action, "type")) == "button" {
		return renderButton(NewParams(
			"text", get(action, "text"),
			"type", defTruthy(get(action, "type"), "button"),
			"name", get(action, "name"),
			"value", get(action, "value"),
			"classes", get(action, "classes"),
			"href", href,
			"attributes", get(action, "attributes"),
		))
	}
	return `<a class="govuk-link` + classesIf(get(action, "classes")) + `" href="` + out(href) +
		`"` + Attributes(get(action, "attributes")) + ">" + out(get(action, "text")) + "</a>"
}
