package baseline

// Apply OWASP and cache headers for a response kind.

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// PreloadLink is one same-origin resource to announce in the Link header.
type PreloadLink struct {
	// Href is a same-origin path, such as /assets/fonts/light-94a07e06a1-v2.woff2.
	Href string
	// As is the destination: "font", "style", "script", "image", or "fetch".
	As string
	// Type is the MIME type, when it is worth stating.
	Type string
}

// HeaderOptions describes one response so the baseline can choose its headers.
type HeaderOptions struct {
	// Kind is one of the Kind… constants and decides caching and document hardening.
	Kind string
	// SecureTransport reports whether the request arrived over HTTPS. HSTS is only sent when it
	// did, so a local HTTP session is not pinned to a scheme the developer cannot serve.
	SecureTransport bool
	// SetsCookie downgrades a public document to private caching.
	SetsCookie bool
	// ContentType overrides the kind's default type.
	ContentType string
	// Preload lists resources to announce in the Link header.
	Preload []PreloadLink
}

var (
	preloadAs   = map[string]bool{"style": true, "script": true, "font": true, "image": true, "fetch": true}
	preloadType = regexp.MustCompile(`^[\w.+-]+/[\w.+-]+$`)
	charsetType = regexp.MustCompile(`(?i)charset=`)
	textType    = regexp.MustCompile(`(?i)^text/`)
	utf8Type    = regexp.MustCompile(`(?i)json|xml|javascript`)
)

// IsDocument reports whether a response kind is an HTML document.
func IsDocument(kind string) bool {
	return kind == KindDocument || kind == KindSensitiveDocument
}

// ApplyHeaders writes the baseline headers for one response onto h.
//
// It removes the headers policy.json bans before setting anything, so a header added by a proxy
// or an earlier handler cannot survive. Documents additionally get the frame, cross-origin,
// permissions, and content security policies; other kinds do not, because those headers only
// mean something for a browsing context.
func (p *Policy) ApplyHeaders(h http.Header, options HeaderOptions) error {
	cacheControl, ok := p.CacheControl[options.Kind]
	if !ok {
		return fmt.Errorf("baseline: unknown response kind: %q", options.Kind)
	}
	document := IsDocument(options.Kind)
	if options.SetsCookie && !document {
		return fmt.Errorf("baseline: Set-Cookie belongs on HTML documents, not on %q", options.Kind)
	}

	for _, name := range p.Remove {
		h.Del(name)
	}

	if contentType, err := p.resolveContentType(options); err != nil {
		return err
	} else if contentType != "" {
		h.Set("Content-Type", contentType)
	}

	if options.SetsCookie && options.Kind == KindDocument {
		cacheControl = "private, no-cache"
	}
	h.Set("Cache-Control", cacheControl)

	for name, value := range p.Headers.All {
		h.Set(name, value)
	}
	if document {
		for name, value := range p.Headers.Document {
			h.Set(name, value)
		}
		h.Set("Permissions-Policy", p.PermissionsPolicyHeader())
		h.Set("Content-Security-Policy", p.ContentSecurityPolicy())
	}
	if options.SecureTransport {
		h.Set("Strict-Transport-Security", p.strictTransportSecurity())
	}
	h.Set("Vary", "Accept-Encoding")
	if len(options.Preload) > 0 {
		link, err := PreloadLinkHeader(options.Preload)
		if err != nil {
			return err
		}
		h.Set("Link", link)
	}
	return nil
}

// ContentSecurityPolicy builds the CSP header value.
//
// The hash of the js-enabled snippet is appended to script-src so GOV.UK Frontend's inline
// snippet runs without opening the policy to 'unsafe-inline'.
func (p *Policy) ContentSecurityPolicy() string {
	parts := make([]string, 0, len(p.CSP.Directives))
	for _, directive := range p.CSP.Directives {
		if len(directive.Sources) == 0 {
			parts = append(parts, directive.Name)
			continue
		}
		sources := append([]string(nil), directive.Sources...)
		if directive.Name == "script-src" {
			sources = append(sources, "'"+p.JSEnabledScriptHash+"'")
		}
		parts = append(parts, directive.Name+" "+strings.Join(sources, " "))
	}
	return strings.Join(parts, "; ")
}

// PermissionsPolicyHeader denies every feature listed in the policy.
func (p *Policy) PermissionsPolicyHeader() string {
	features := make([]string, 0, len(p.PermissionsPolicy))
	for _, name := range p.PermissionsPolicy {
		features = append(features, name+"=()")
	}
	return strings.Join(features, ", ")
}

// PreloadLinkHeader formats preload hints for the Link header.
//
// Hrefs must be same-origin paths: a preload is a promise about this service's own assets, and
// an absolute URL would leak the visit to another origin.
func PreloadLinkHeader(links []PreloadLink) (string, error) {
	if len(links) == 0 {
		return "", fmt.Errorf("baseline: preload links must not be empty")
	}
	formatted := make([]string, 0, len(links))
	for _, link := range links {
		if !isSameOriginPath(link.Href) {
			return "", fmt.Errorf("baseline: preload href must be a same-origin path: %q", link.Href)
		}
		if !preloadAs[link.As] {
			return "", fmt.Errorf("baseline: unsupported preload as: %q", link.As)
		}
		parts := []string{"<" + link.Href + ">", "rel=preload", "as=" + link.As}
		if link.Type != "" {
			if !preloadType.MatchString(link.Type) {
				return "", fmt.Errorf("baseline: invalid preload type: %q", link.Type)
			}
			parts = append(parts, `type="`+link.Type+`"`)
		}
		if link.As == "font" {
			parts = append(parts, "crossorigin")
		}
		formatted = append(formatted, strings.Join(parts, "; "))
	}
	return strings.Join(formatted, ", "), nil
}

// StrongETag returns a strong validator for a response body.
func StrongETag(body []byte) string {
	sum := sha256.Sum256(body)
	return `"` + base64.RawURLEncoding.EncodeToString(sum[:]) + `"`
}

func (p *Policy) resolveContentType(options HeaderOptions) (string, error) {
	chosen := options.ContentType
	if chosen == "" {
		chosen = p.ContentTypes[options.Kind]
	}
	if chosen == "" {
		return "", nil
	}
	if strings.ContainsAny(chosen, "\r\n") {
		return "", fmt.Errorf("baseline: invalid Content-Type: %q", chosen)
	}
	if charsetType.MatchString(chosen) {
		return chosen, nil
	}
	if textType.MatchString(chosen) || utf8Type.MatchString(chosen) {
		return chosen + "; charset=utf-8", nil
	}
	return chosen, nil
}

func (p *Policy) strictTransportSecurity() string {
	value := "max-age=" + strconv.Itoa(p.HSTS.MaxAge)
	if p.HSTS.IncludeSubDomains {
		value += "; includeSubDomains"
	}
	return value
}

func isSameOriginPath(href string) bool {
	return strings.HasPrefix(href, "/") &&
		!strings.HasPrefix(href, "//") &&
		!strings.ContainsAny(href, " \t\"<>")
}
