package components

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"
)

// Fixture is one official example from a component's fixtures.json.
//
// HTML is the contract: a renderer given Options must produce exactly these bytes. It is never
// edited to make a test pass.
type Fixture struct {
	Name        string
	Options     map[string]any
	HTML        string
	Hidden      bool
	Description string
}

// ComponentFixtures is the parsed fixtures.json for one component.
type ComponentFixtures struct {
	Component string
	Fixtures  []Fixture
}

// Library reads fixtures from an installed GOV.UK Frontend components directory, caching each
// file after the first read.
type Library struct {
	root  string
	mu    sync.Mutex
	cache map[string]ComponentFixtures
}

// NewLibrary returns a library reading from root, which is dist/govuk/components in the pinned
// package.
func NewLibrary(root string) *Library {
	return &Library{root: root, cache: map[string]ComponentFixtures{}}
}

// Names lists the component directories in the package, in name order.
func (l *Library) Names() ([]string, error) {
	entries, err := os.ReadDir(l.root)
	if err != nil {
		return nil, fmt.Errorf("components: reading %s: %w", l.root, err)
	}
	var names []string
	for _, entry := range entries {
		if entry.IsDir() && IsComponentName(entry.Name()) {
			names = append(names, entry.Name())
		}
	}
	slices.Sort(names)
	return names, nil
}

// Has reports whether the package ships a component with this name.
func (l *Library) Has(name string) bool {
	if !IsComponentName(name) {
		return false
	}
	_, err := os.Stat(filepath.Join(l.root, name, "fixtures.json"))
	return err == nil
}

// Load reads and caches the fixtures for one component.
func (l *Library) Load(name string) (ComponentFixtures, error) {
	if !IsComponentName(name) {
		return ComponentFixtures{}, fmt.Errorf("components: unknown GOV.UK Frontend component: %q", name)
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if cached, ok := l.cache[name]; ok {
		return cached, nil
	}
	raw, err := os.ReadFile(filepath.Join(l.root, name, "fixtures.json"))
	if err != nil {
		return ComponentFixtures{}, fmt.Errorf("components: reading fixtures for %q: %w", name, err)
	}
	parsed, err := ParseFixtures(name, raw)
	if err != nil {
		return ComponentFixtures{}, err
	}
	l.cache[name] = parsed
	return parsed, nil
}

// Fixture finds one fixture by name.
func (l *Library) Fixture(component, fixture string) (Fixture, bool) {
	loaded, err := l.Load(component)
	if err != nil {
		return Fixture{}, false
	}
	for _, candidate := range loaded.Fixtures {
		if candidate.Name == fixture {
			return candidate, true
		}
	}
	return Fixture{}, false
}

// ParseFixtures reads a fixtures.json document, including hidden fixtures.
//
// Hidden fixtures are kept because parity is measured against every fixture in the release, not
// only the ones the Design System website shows.
func ParseFixtures(component string, raw []byte) (ComponentFixtures, error) {
	var document struct {
		Fixtures []struct {
			Name        string         `json:"name"`
			Options     map[string]any `json:"options"`
			HTML        *string        `json:"html"`
			Hidden      bool           `json:"hidden"`
			Description string         `json:"description"`
		} `json:"fixtures"`
	}
	if err := json.Unmarshal(raw, &document); err != nil {
		return ComponentFixtures{}, fmt.Errorf("components: invalid fixtures for %q: %w", component, err)
	}
	if document.Fixtures == nil {
		return ComponentFixtures{}, fmt.Errorf("components: invalid fixtures for %q", component)
	}
	fixtures := make([]Fixture, 0, len(document.Fixtures))
	for index, entry := range document.Fixtures {
		if entry.Name == "" || entry.HTML == nil {
			return ComponentFixtures{}, fmt.Errorf("components: invalid fixture %s#%d", component, index)
		}
		options := entry.Options
		if options == nil {
			options = map[string]any{}
		}
		fixtures = append(fixtures, Fixture{
			Name:        entry.Name,
			Options:     options,
			HTML:        *entry.HTML,
			Hidden:      entry.Hidden,
			Description: entry.Description,
		})
	}
	return ComponentFixtures{Component: component, Fixtures: fixtures}, nil
}

// SelectFixture chooses the fixture to preview.
//
// A named fixture is used when it exists; otherwise the first visible fixture, falling back to
// the first fixture of all when every one is hidden.
func SelectFixture(fixtures []Fixture, requested string) (Fixture, bool) {
	if requested != "" {
		for _, fixture := range fixtures {
			if fixture.Name == requested {
				return fixture, true
			}
		}
		return Fixture{}, false
	}
	for _, fixture := range fixtures {
		if !fixture.Hidden {
			return fixture, true
		}
	}
	if len(fixtures) > 0 {
		return fixtures[0], true
	}
	return Fixture{}, false
}

// ParityBanner returns notification banner options reporting whether rendered HTML matched the
// fixture.
func ParityBanner(matches bool) map[string]any {
	if matches {
		return map[string]any{
			"type":      "success",
			"titleText": "HTML matches the fixture",
			"text":      "The macro output is the same as the official fixture HTML.",
		}
	}
	return map[string]any{
		"titleText": "HTML does not match the fixture",
		"text":      "The macro output is different from the official fixture HTML.",
	}
}
