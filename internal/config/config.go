// Package config resolves the paths and settings the example service needs at run time.
//
// Go generates every page; Node is only the delivery mechanism for three inputs. GOV.UK
// Frontend is installed with npm and supplies the script, the assets, and the official
// fixtures; the stylesheet is compiled by the Sass pipeline; and the shared response policy
// lives in baseline/policy.json. All three are found relative to the repository root, so the
// root is resolved once at start-up and passed down rather than recomputed.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// ServiceName is the English service name used in the header, page titles, and phase banner.
const ServiceName = "Apply for a rod fishing licence"

// ServiceNameCy is the Welsh service name used on the Welsh start page.
const ServiceNameCy = "Gwneud cais am drwydded bysgota"

// MaxBodyBytes is the largest request body the service reads.
const MaxBodyBytes = 1_000_000

// DefaultPort is used when PORT is unset or empty.
const DefaultPort = 3000

// ErrRootNotFound is returned when no ancestor directory looks like this repository.
var ErrRootNotFound = errors.New("config: repository root not found")

// Config holds the resolved repository paths and the pinned GOV.UK Frontend version.
type Config struct {
	// Root is the repository root: the directory holding package.json and baseline/.
	Root string
	// GovukDist is the installed govuk-frontend dist directory, the Nunjucks search path.
	GovukDist string
	// GovukRoot is dist/govuk, which holds the page template, components, and the minified script.
	GovukRoot string
	// ComponentsRoot is the component directories that ship fixtures.json.
	ComponentsRoot string
	// FrontendAssets is the fonts, images, and manifest served under /assets.
	FrontendAssets string
	// Stylesheet is the compiled application CSS from `npm run build:styles`.
	// Never serve Frontend's prebuilt govuk-frontend.min.css as the long-term source.
	Stylesheet string
	// PolicyFile is the shared OWASP and cache policy synced from the base template.
	PolicyFile string
	// FrontendVersion is the pinned govuk-frontend release. CSS, JavaScript, and fixtures
	// all come from this package.
	FrontendVersion string
}

// Find walks up from start, which must be an absolute directory, looking for the repository root.
//
// A directory is the root when it holds both package.json and baseline/policy.json, which
// distinguishes it from any parent that merely happens to contain a package.json.
//
// It returns [ErrRootNotFound] when no ancestor matches.
func Find(start string) (string, error) {
	if !filepath.IsAbs(start) {
		return "", fmt.Errorf("config: start directory must be absolute, got %q", start)
	}
	dir := filepath.Clean(start)
	for {
		if isRoot(dir) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("%w above %s", ErrRootNotFound, start)
		}
		dir = parent
	}
}

// New builds a configuration for a known repository root.
//
// It fails when GOV.UK Frontend is not installed, so a missing `npm install` is reported at
// start-up rather than as a confusing render error later.
func New(root string) (*Config, error) {
	govukPackage := filepath.Join(root, "node_modules", "govuk-frontend")
	version, err := frontendVersion(filepath.Join(govukPackage, "package.json"))
	if err != nil {
		return nil, err
	}
	dist := filepath.Join(govukPackage, "dist")
	govukRoot := filepath.Join(dist, "govuk")
	return &Config{
		Root:            root,
		GovukDist:       dist,
		GovukRoot:       govukRoot,
		ComponentsRoot:  filepath.Join(govukRoot, "components"),
		FrontendAssets:  filepath.Join(govukRoot, "assets"),
		Stylesheet:      filepath.Join(root, "dist", "stylesheets", "application.css"),
		PolicyFile:      filepath.Join(root, "baseline", "policy.json"),
		FrontendVersion: version,
	}, nil
}

// Load resolves the configuration from the working directory.
//
// It works from the repository root (`go run ./cmd/server`) and from any package directory
// (`go test ./...`), because [Find] walks upwards.
func Load() (*Config, error) {
	return LoadFrom(os.Getwd)
}

// LoadFrom resolves the configuration from the directory workingDir reports.
//
// Tests use it to cover the failure paths without changing the process working directory.
func LoadFrom(workingDir func() (string, error)) (*Config, error) {
	start, err := workingDir()
	if err != nil {
		return nil, fmt.Errorf("config: working directory: %w", err)
	}
	root, err := Find(start)
	if err != nil {
		return nil, err
	}
	return New(root)
}

// DemosEnabled reports whether the component catalogue and example pages are served.
//
// They are off in production so a real service does not publish the fixture previews.
func DemosEnabled(getenv func(string) string) bool {
	return getenv("NODE_ENV") != "production"
}

// ResolvePort reads the listening port from the environment.
//
// It returns [DefaultPort] when PORT is unset or empty, and an error when PORT is set to
// something that is not a port number, so a typo fails loudly instead of silently binding 3000.
func ResolvePort(getenv func(string) string) (int, error) {
	raw := getenv("PORT")
	if raw == "" {
		return DefaultPort, nil
	}
	port, err := strconv.Atoi(raw)
	if err != nil || port < 0 || port > 65535 {
		return 0, fmt.Errorf("config: invalid PORT: %s", raw)
	}
	return port, nil
}

func isRoot(dir string) bool {
	if _, err := os.Stat(filepath.Join(dir, "package.json")); err != nil {
		return false
	}
	_, err := os.Stat(filepath.Join(dir, "baseline", "policy.json"))
	return err == nil
}

func frontendVersion(packageFile string) (string, error) {
	raw, err := os.ReadFile(packageFile)
	if err != nil {
		return "", fmt.Errorf("config: govuk-frontend is not installed, run `npm install`: %w", err)
	}
	var meta struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(raw, &meta); err != nil {
		return "", fmt.Errorf("config: reading %s: %w", packageFile, err)
	}
	if meta.Version == "" {
		return "", fmt.Errorf("config: %s has no version", packageFile)
	}
	return meta.Version, nil
}
