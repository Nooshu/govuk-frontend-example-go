package govuk_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/Nooshu/govuk-frontend-example-go/internal/config"
	"github.com/Nooshu/govuk-frontend-example-go/internal/govuk"
)

// componentsDir resolves the installed GOV.UK Frontend component directory, which holds the
// official fixtures. It fails the test rather than skipping, because a missing `npm install`
// would otherwise look like a green parity suite.
func componentsDir(t *testing.T) string {
	t.Helper()
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("resolving configuration: %v", err)
	}
	if _, err := os.Stat(cfg.ComponentsRoot); err != nil {
		t.Fatalf("GOV.UK Frontend is not installed, run `npm install`: %v", err)
	}
	return cfg.ComponentsRoot
}

// TestRenderMatchesFixtures is the parity gate: for every component GOV.UK Frontend ships and
// every fixture it declares, hidden ones included, the Go renderer must return the fixture's
// HTML byte for byte.
func TestRenderMatchesFixtures(t *testing.T) {
	root := componentsDir(t)
	components, err := govuk.FixtureComponents(root)
	if err != nil {
		t.Fatalf("listing components: %v", err)
	}
	if len(components) == 0 {
		t.Fatal("no components with fixtures found")
	}

	total := 0
	for _, component := range components {
		set, err := govuk.LoadFixtures(root, component)
		if err != nil {
			t.Fatalf("loading fixtures: %v", err)
		}
		total += len(set.Fixtures)

		t.Run(component, func(t *testing.T) {
			for _, fixture := range set.Fixtures {
				t.Run(fixture.Name, func(t *testing.T) {
					got, err := govuk.Render(component, fixture.Options)
					if err != nil {
						t.Fatalf("Render(%q): %v", component, err)
					}
					if got != fixture.HTML {
						t.Errorf("HTML does not match fixture\n%s", difference(fixture.HTML, got))
					}
				})
			}
		})
	}

	t.Logf("checked %d fixtures across %d components", total, len(components))
}

// TestComponentsCoverEveryFixtureComponent guards against a GOV.UK Frontend upgrade adding a
// component that this package does not render yet.
func TestComponentsCoverEveryFixtureComponent(t *testing.T) {
	root := componentsDir(t)
	shipped, err := govuk.FixtureComponents(root)
	if err != nil {
		t.Fatalf("listing components: %v", err)
	}

	supported := make(map[string]bool, len(govuk.Components()))
	for _, name := range govuk.Components() {
		supported[name] = true
	}
	for _, name := range shipped {
		if !supported[name] {
			t.Errorf("component %q ships fixtures but has no Go renderer", name)
		}
	}
}

func TestRenderRejectsUnknownComponent(t *testing.T) {
	if _, err := govuk.Render("not-a-component", nil); err == nil {
		t.Fatal("expected an error for an unknown component")
	}
}

// difference reports the first line that differs, which keeps failures readable when a
// component renders hundreds of lines.
func difference(want, got string) string {
	wantLines := strings.Split(want, "\n")
	gotLines := strings.Split(got, "\n")
	for i := 0; i < len(wantLines) || i < len(gotLines); i++ {
		wantLine, gotLine := "<missing line>", "<missing line>"
		if i < len(wantLines) {
			wantLine = wantLines[i]
		}
		if i < len(gotLines) {
			gotLine = gotLines[i]
		}
		if wantLine != gotLine {
			return fmt.Sprintf("first difference on line %d\nwant: %q\ngot:  %q", i+1, wantLine, gotLine)
		}
	}
	return fmt.Sprintf("line-by-line equal but strings differ\nwant: %q\ngot:  %q", want, got)
}
