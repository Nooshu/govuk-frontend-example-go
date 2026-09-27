package components_test

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/Nooshu/govuk-frontend-example-go/internal/components"
	"github.com/Nooshu/govuk-frontend-example-go/internal/config"
)

func TestNames(t *testing.T) {
	cases := []struct {
		name, macro, title string
		ok                 bool
	}{
		{"date-input", "govukDateInput", "Date Input", true},
		{"button", "govukButton", "Button", true},
		{"a", "govukA", "A", true},
		{"", "", "", false},
		{"Date", "", "", false},
		{"date-", "", "", false},
		{"date--input", "", "", false},
	}
	for _, tc := range cases {
		if got := components.IsComponentName(tc.name); got != tc.ok {
			t.Errorf("IsComponentName(%q)=%v", tc.name, got)
		}
		if !tc.ok {
			continue
		}
		if got := components.MacroName(tc.name); got != tc.macro {
			t.Errorf("MacroName(%q)=%q want %q", tc.name, got, tc.macro)
		}
		if got := components.TitleFromKebab(tc.name); got != tc.title {
			t.Errorf("TitleFromKebab(%q)=%q want %q", tc.name, got, tc.title)
		}
	}
	if components.TitleFromKebab("date-") != "Date " {
		t.Fatalf("TitleFromKebab(%q)=%q", "date-", components.TitleFromKebab("date-"))
	}
}

func TestDescribeAndCatalogueCopy(t *testing.T) {
	known := components.Describe("button")
	if known.Name != "button" || known.Title != "Button" || known.DesignSystemURL == "" {
		t.Fatalf("%#v", known)
	}
	unknown := components.Describe("future-widget")
	if unknown.Title != "Future Widget" || !strings.Contains(unknown.DesignSystemURL, "future-widget") {
		t.Fatalf("%#v", unknown)
	}
	names := components.DescribedNames()
	if len(names) < 30 || !slices.Contains(names, "button") {
		t.Fatalf("described names = %d", len(names))
	}
}

func TestParseAndSelectFixtures(t *testing.T) {
	parsed, err := components.ParseFixtures("button", []byte(`{
		"fixtures": [
			{"name": "hidden one", "html": "<p>h</p>", "hidden": true},
			{"name": "default", "html": "<p>d</p>", "description": "The usual button"}
		]
	}`))
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Component != "button" || len(parsed.Fixtures) != 2 || parsed.Fixtures[1].Options == nil {
		t.Fatalf("%#v", parsed)
	}
	if fixture, ok := components.SelectFixture(parsed.Fixtures, ""); !ok || fixture.Name != "default" {
		t.Fatalf("visible fixture = %#v %v", fixture, ok)
	}
	if fixture, ok := components.SelectFixture(parsed.Fixtures, "hidden one"); !ok || !fixture.Hidden {
		t.Fatalf("named fixture = %#v %v", fixture, ok)
	}
	if _, ok := components.SelectFixture(parsed.Fixtures, "missing"); ok {
		t.Fatal("missing fixture was selected")
	}
	onlyHidden := []components.Fixture{{Name: "secret", Hidden: true, HTML: "x"}}
	if fixture, ok := components.SelectFixture(onlyHidden, ""); !ok || fixture.Name != "secret" {
		t.Fatalf("hidden fallback = %#v %v", fixture, ok)
	}
	if _, ok := components.SelectFixture(nil, ""); ok {
		t.Fatal("empty list selected a fixture")
	}

	invalid := []string{
		`not json`,
		`{}`,
		`{"fixtures":[{"html":"<p></p>"}]}`,
		`{"fixtures":[{"name":"x"}]}`,
	}
	for _, raw := range invalid {
		if _, err := components.ParseFixtures("button", []byte(raw)); err == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
}

func TestParityBanner(t *testing.T) {
	match := components.ParityBanner(true)
	if match["type"] != "success" || match["titleText"] == "" {
		t.Fatalf("%#v", match)
	}
	mismatch := components.ParityBanner(false)
	if _, ok := mismatch["type"]; ok || mismatch["titleText"] == "" {
		t.Fatalf("%#v", mismatch)
	}
}

func TestLibrary(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	library := components.NewLibrary(cfg.ComponentsRoot)
	names, err := library.Names()
	if err != nil || len(names) == 0 || !slices.IsSorted(names) {
		t.Fatalf("names: %v %v", names, err)
	}
	if !library.Has("button") || library.Has("not a name") || library.Has("missing-widget") {
		t.Fatal("Has mismatch")
	}
	loaded, err := library.Load("button")
	if err != nil || len(loaded.Fixtures) == 0 {
		t.Fatal(err)
	}
	again, err := library.Load("button")
	if err != nil || again.Fixtures[0].Name != loaded.Fixtures[0].Name {
		t.Fatal(err)
	}
	fixture, ok := library.Fixture("button", loaded.Fixtures[0].Name)
	if !ok || fixture.HTML == "" {
		t.Fatal("fixture lookup failed")
	}
	if _, ok := library.Fixture("button", "this fixture does not exist"); ok {
		t.Fatal("missing fixture was found")
	}
	if _, ok := library.Fixture("../button", "default"); ok {
		t.Fatal("invalid component was found")
	}
	if _, err := library.Load("not a name"); err == nil {
		t.Fatal("invalid name was loaded")
	}
	if _, err := components.NewLibrary(t.TempDir()).Load("button"); err == nil {
		t.Fatal("missing fixtures file was loaded")
	}
	if _, err := components.NewLibrary(filepath.Join(t.TempDir(), "missing")).Names(); err == nil {
		t.Fatal("missing directory was listed")
	}

	catalogue, err := library.Catalogue()
	if err != nil || len(catalogue) != len(names) || catalogue[0].Name != names[0] {
		t.Fatalf("catalogue: %v %v", err, len(catalogue))
	}
	if _, err := components.NewLibrary(filepath.Join(t.TempDir(), "missing")).Catalogue(); err == nil {
		t.Fatal("catalogue of a missing directory succeeded")
	}

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "NotAComponent"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "button"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "button", "fixtures.json"), []byte(`{"fixtures":[{"name":"bad"`), 0o644); err != nil {
		t.Fatal(err)
	}
	local := components.NewLibrary(dir)
	listed, err := local.Names()
	if err != nil || len(listed) != 1 || listed[0] != "button" {
		t.Fatalf("filtered names = %v %v", listed, err)
	}
	if local.Has("button") != true {
		t.Fatal("button fixtures were not detected")
	}
	if _, err := local.Load("button"); err == nil {
		t.Fatal("invalid fixtures.json was accepted")
	}
}
