package config_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Nooshu/govuk-frontend-example-go/internal/config"
)

func TestFindWalksUpToTheRepositoryRoot(t *testing.T) {
	t.Parallel()

	root, err := config.Find(workingDir(t))
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	for _, marker := range []string{"package.json", filepath.Join("baseline", "policy.json")} {
		if _, err := os.Stat(filepath.Join(root, marker)); err != nil {
			t.Errorf("root %s is missing %s: %v", root, marker, err)
		}
	}
}

func TestFindRejectsDirectoriesOutsideTheRepository(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte("{}"), 0o600); err != nil {
		t.Fatalf("writing package.json: %v", err)
	}

	// package.json alone is not enough: the shared baseline must be there too.
	if _, err := config.Find(dir); !errors.Is(err, config.ErrRootNotFound) {
		t.Fatalf("Find(%q) error = %v, want ErrRootNotFound", dir, err)
	}
}

func TestFindRejectsRelativeStartDirectories(t *testing.T) {
	t.Parallel()

	if _, err := config.Find("."); err == nil || !strings.Contains(err.Error(), "must be absolute") {
		t.Fatalf(`Find(".") error = %v, want it to mention "must be absolute"`, err)
	}
}

func TestNewResolvesEveryPathFromTheRoot(t *testing.T) {
	t.Parallel()

	root, err := config.Find(workingDir(t))
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	cfg, err := config.New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if cfg.FrontendVersion == "" {
		t.Error("FrontendVersion is empty")
	}
	for name, path := range map[string]string{
		"GovukDist":      cfg.GovukDist,
		"GovukRoot":      cfg.GovukRoot,
		"ComponentsRoot": cfg.ComponentsRoot,
		"FrontendAssets": cfg.FrontendAssets,
		"PolicyFile":     cfg.PolicyFile,
	} {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("%s (%s) does not exist: %v", name, path, err)
		}
	}
	if want := filepath.Join(root, "dist", "stylesheets", "application.css"); cfg.Stylesheet != want {
		t.Errorf("Stylesheet = %s, want %s", cfg.Stylesheet, want)
	}
}

func TestNewReportsAMissingOrBrokenFrontendInstall(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		packageJSON *string
		want        string
	}{
		"not installed": {packageJSON: nil, want: "npm install"},
		"not json":      {packageJSON: new("not json"), want: "reading"},
		"no version":    {packageJSON: new(`{"name":"govuk-frontend"}`), want: "no version"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			if test.packageJSON != nil {
				dir := filepath.Join(root, "node_modules", "govuk-frontend")
				if err := os.MkdirAll(dir, 0o700); err != nil {
					t.Fatalf("creating %s: %v", dir, err)
				}
				file := filepath.Join(dir, "package.json")
				if err := os.WriteFile(file, []byte(*test.packageJSON), 0o600); err != nil {
					t.Fatalf("writing %s: %v", file, err)
				}
			}

			_, err := config.New(root)
			if err == nil {
				t.Fatal("New succeeded, want an error")
			}
			if !strings.Contains(err.Error(), test.want) {
				t.Errorf("New error = %q, want it to mention %q", err, test.want)
			}
		})
	}
}

func TestLoadFindsTheRepositoryFromThePackageDirectory(t *testing.T) {
	t.Parallel()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.FrontendVersion == "" {
		t.Error("FrontendVersion is empty")
	}
}

func TestLoadFromSurfacesWorkingDirectoryFailures(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		workingDir func() (string, error)
		want       string
	}{
		"cannot read the working directory": {
			workingDir: func() (string, error) { return "", errors.New("no such directory") },
			want:       "working directory",
		},
		"outside any repository": {
			workingDir: func() (string, error) { return t.TempDir(), nil },
			want:       "repository root not found",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, err := config.LoadFrom(test.workingDir); err == nil ||
				!strings.Contains(err.Error(), test.want) {
				t.Errorf("LoadFrom error = %v, want it to mention %q", err, test.want)
			}
		})
	}
}

func TestDemosEnabledIsOffInProduction(t *testing.T) {
	t.Parallel()

	tests := map[string]bool{"": true, "development": true, "test": true, "production": false}
	for value, want := range tests {
		if got := config.DemosEnabled(env("NODE_ENV", value)); got != want {
			t.Errorf("DemosEnabled(NODE_ENV=%q) = %t, want %t", value, got, want)
		}
	}
}

func TestResolvePort(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		value   string
		want    int
		wantErr bool
	}{
		"unset":        {value: "", want: config.DefaultPort},
		"zero":         {value: "0", want: 0},
		"in range":     {value: "8080", want: 8080},
		"top of range": {value: "65535", want: 65535},
		"negative":     {value: "-1", wantErr: true},
		"too large":    {value: "65536", wantErr: true},
		"not a number": {value: "eighty", wantErr: true},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := config.ResolvePort(env("PORT", test.value))
			if test.wantErr {
				if err == nil {
					t.Fatalf("ResolvePort(%q) = %d, want an error", test.value, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ResolvePort(%q): %v", test.value, err)
			}
			if got != test.want {
				t.Errorf("ResolvePort(%q) = %d, want %d", test.value, got, test.want)
			}
		})
	}
}

func workingDir(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("working directory: %v", err)
	}
	return dir
}

func env(name, value string) func(string) string {
	return func(key string) string {
		if key == name {
			return value
		}
		return ""
	}
}
