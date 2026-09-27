package httpx

// Static asset map for stylesheet, Frontend JS, and fonts.

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Nooshu/govuk-frontend-example-go/internal/baseline"
)

// rootFiles are served from dist/govuk rather than dist/govuk/assets.
var rootFiles = map[string]bool{"govuk-frontend.min.js": true}

var contentTypes = map[string]string{
	".css":   "text/css; charset=utf-8",
	".js":    "text/javascript; charset=utf-8",
	".mjs":   "text/javascript; charset=utf-8",
	".woff2": "font/woff2",
	".woff":  "font/woff",
	".svg":   "image/svg+xml",
	".png":   "image/png",
	".ico":   "image/x-icon",
	".json":  "application/json; charset=utf-8",
	".gif":   "image/gif",
	".jpg":   "image/jpeg",
	".jpeg":  "image/jpeg",
}

var (
	hashedFont = regexp.MustCompile(`-[a-f0-9]{8,}-`)
	fontURL    = regexp.MustCompile(`url\(\s*["']?(/assets/fonts/[^"')\s]+\.woff2)["']?\s*\)`)
)

// Asset is a static file that is safe to send.
type Asset struct {
	// Body is the file content when it is already in memory, as fingerprinted files are.
	Body []byte
	// FilePath is where to read the file from when Body is nil.
	FilePath string
	// ContentType is the type to send.
	ContentType string
	// Kind is the baseline cache kind for this URL.
	Kind string
}

// PageAssets are the URLs one page needs in its head and at the end of its body.
type PageAssets struct {
	// StylesheetHref is the compiled application stylesheet from the Sass pipeline.
	StylesheetHref string
	// AppModuleHref is the ES module that calls initAll().
	AppModuleHref string
	// Preloads are the fonts the stylesheet uses, announced in the Link header.
	Preloads []baseline.PreloadLink
}

// Assets serves the compiled stylesheet, GOV.UK Frontend's script, and the files under
// /assets/.
//
// The stylesheet, the Frontend script, and the small module that calls initAll() are read once
// and published at URLs that contain a hash of their bytes. That is what lets them be cached
// immutably for a year: the URL changes as soon as the content does.
type Assets struct {
	govukRoot      string
	frontendAssets string

	stylesheetPath string
	cssHref        string
	scriptHref     string
	appHref        string

	css    []byte
	script []byte
	app    []byte

	preloads []baseline.PreloadLink
}

// NewAssets reads the published files and works out their fingerprinted URLs.
//
// A missing stylesheet is an error rather than a blank page, because it means `npm run
// build:styles` has not run.
func NewAssets(govukRoot, frontendAssets, stylesheetPath string) (*Assets, error) {
	css, err := os.ReadFile(stylesheetPath)
	if err != nil {
		return nil, fmt.Errorf("httpx: missing %s, run `npm run build:styles` first: %w", stylesheetPath, err)
	}
	scriptPath := filepath.Join(govukRoot, "govuk-frontend.min.js")
	script, err := os.ReadFile(scriptPath)
	if err != nil {
		return nil, fmt.Errorf("httpx: reading %s: %w", scriptPath, err)
	}
	cssHref := "/assets/application." + fingerprint(css) + ".css"
	scriptHref := "/assets/govuk-frontend." + fingerprint(script) + ".min.js"
	app := []byte("import { initAll } from '" + scriptHref + "';\n\ninitAll();\n")
	return &Assets{
		govukRoot:      govukRoot,
		frontendAssets: frontendAssets,
		stylesheetPath: stylesheetPath,
		cssHref:        cssHref,
		scriptHref:     scriptHref,
		appHref:        "/assets/app." + fingerprint(app) + ".mjs",
		css:            css,
		script:         script,
		app:            app,
		preloads:       fontPreloads(string(css)),
	}, nil
}

// Page returns the asset URLs and font preloads for a page.
func (a *Assets) Page() PageAssets {
	return PageAssets{
		StylesheetHref: a.cssHref,
		AppModuleHref:  a.appHref,
		Preloads:       append([]baseline.PreloadLink(nil), a.preloads...),
	}
}

// Resolve maps an /assets/… path to a file that is safe to send.
//
// It reports false for anything outside the Frontend package, any extension that is not a known
// asset type, and anything that is not a readable file, so the route can never be used to read
// arbitrary paths from the host.
func (a *Assets) Resolve(urlPath string) (Asset, bool) {
	if !strings.HasPrefix(urlPath, "/assets/") {
		return Asset{}, false
	}
	switch urlPath {
	case a.cssHref:
		return Asset{
			Body:        a.css,
			FilePath:    a.stylesheetPath,
			ContentType: "text/css; charset=utf-8",
			Kind:        baseline.KindFingerprintedAsset,
		}, true
	case a.scriptHref:
		return Asset{
			Body:        a.script,
			FilePath:    filepath.Join(a.govukRoot, "govuk-frontend.min.js"),
			ContentType: "text/javascript; charset=utf-8",
			Kind:        baseline.KindFingerprintedAsset,
		}, true
	case a.appHref:
		return Asset{
			Body:        a.app,
			ContentType: "text/javascript; charset=utf-8",
			Kind:        baseline.KindFingerprintedAsset,
		}, true
	}

	requested := strings.TrimPrefix(urlPath, "/assets/")
	if requested == "" || strings.ContainsRune(requested, 0) {
		return Asset{}, false
	}
	root := a.frontendAssets
	if rootFiles[requested] {
		root = a.govukRoot
	}
	target := filepath.Join(root, filepath.FromSlash(requested))
	if !isInside(root, target) {
		return Asset{}, false
	}
	contentType, known := contentTypes[strings.ToLower(filepath.Ext(target))]
	if !known {
		return Asset{}, false
	}
	info, err := os.Stat(target)
	if err != nil || !info.Mode().IsRegular() {
		return Asset{}, false
	}
	return Asset{FilePath: target, ContentType: contentType, Kind: assetKind(requested)}, true
}

func assetKind(requested string) string {
	if strings.HasPrefix(requested, "fonts/") && hashedFont.MatchString(requested) {
		return baseline.KindFingerprintedAsset
	}
	return baseline.KindStaticAsset
}

// fontPreloads finds the woff2 files the stylesheet references, so the browser can start
// fetching them before it has finished parsing the CSS.
func fontPreloads(css string) []baseline.PreloadLink {
	var links []baseline.PreloadLink
	seen := map[string]bool{}
	for _, match := range fontURL.FindAllStringSubmatch(css, -1) {
		href := match[1]
		if seen[href] {
			continue
		}
		seen[href] = true
		links = append(links, baseline.PreloadLink{Href: href, As: "font", Type: "font/woff2"})
	}
	return links
}

func fingerprint(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])[:10]
}

func isInside(root, target string) bool {
	relative, err := filepath.Rel(filepath.Clean(root), target)
	if err != nil {
		return false
	}
	return relative != "." && !strings.HasPrefix(relative, "..")
}
