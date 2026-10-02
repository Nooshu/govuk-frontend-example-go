package config

// Repository root, Frontend pin, stylesheet path, PORT, demos.

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"

	jsonv2 "encoding/json/v2"
)

// ServiceName is the English service name used in the header, page titles, and phase banner.
const ServiceName = "Apply for a fishing rod licence"

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
// Precedence:
//  1. DEMOS_ENABLED — "true"/"1"/"yes" forces on; "false"/"0"/"no" forces off
//  2. Otherwise off when NODE_ENV=production (hosts such as Render often set that)
//
// The public Render demo sets DEMOS_ENABLED=true so previews stay visible even when
// NODE_ENV is production.
func DemosEnabled(getenv func(string) string) bool {
	switch getenv("DEMOS_ENABLED") {
	case "true", "1", "yes":
		return true
	case "false", "0", "no":
		return false
	}
	return getenv("NODE_ENV") != "production"
}

// ResolvePort reads the listening port from the environment.
//
// It returns [DefaultPort] when PORT is unset or empty, and an error when PORT is set to
// something that is not a port number, so a typo fails loudly instead of silently binding 3000.
// Cloud hosts such as Render inject PORT automatically.
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

// ResolveListen returns the network and address for [net.Listen].
//
// PORT chooses the port ([DefaultPort] when unset). HOST chooses the interface:
//   - unset or empty — 0.0.0.0, the IPv4 wildcard
//   - 127.0.0.1 — local-only binding for a locked-down laptop
//   - an IPv6 literal — that address, on tcp6
//
// The default is IPv4 on purpose. [net.Listen] with network "tcp" and an
// unspecified address (":PORT" or "0.0.0.0:PORT") opens an IPv6 socket, which
// logs as [::]:PORT. Render's port scan only detects an IPv4 listener on
// 0.0.0.0, so that socket makes a healthy process fail deploy with
// "no open ports detected". See https://render.com/docs/web-services#port-binding.
func ResolveListen(getenv func(string) string) (network, address string, err error) {
	port, err := ResolvePort(getenv)
	if err != nil {
		return "", "", err
	}
	host := getenv("HOST")
	if host == "" {
		host = "0.0.0.0"
	}
	return listenNetwork(host), net.JoinHostPort(host, strconv.Itoa(port)), nil
}

// listenNetwork picks tcp4 unless host is an IPv6 literal.
//
// A non-IP host such as "localhost" stays on tcp4 so the socket is an IPv4
// listener. IPv6 literals need tcp6, or Listen rejects the address.
func listenNetwork(host string) string {
	if ip := net.ParseIP(host); ip != nil && ip.To4() == nil {
		return "tcp6"
	}
	return "tcp4"
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
	if err := jsonv2.Unmarshal(raw, &meta); err != nil {
		return "", fmt.Errorf("config: reading %s: %w", packageFile, err)
	}
	if meta.Version == "" {
		return "", fmt.Errorf("config: %s has no version", packageFile)
	}
	return meta.Version, nil
}
