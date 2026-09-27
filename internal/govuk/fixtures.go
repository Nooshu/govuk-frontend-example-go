package govuk

// Fixture loading helpers used by the parity suite.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// Fixture is one entry from a component's fixtures.json: a set of options and the HTML GOV.UK
// Frontend produces for them.
//
// HTML is the contract. It is never normalised, reformatted, or edited to make a test pass.
type Fixture struct {
	// Name identifies the fixture within its component, for example "default" or "with hint".
	Name string `json:"name"`
	// Options are the macro options, with object key order preserved.
	Options *Params `json:"options"`
	// Hidden marks fixtures that the Design System does not publish as examples. They are
	// still part of the contract and are still checked.
	Hidden bool `json:"hidden"`
	// Description explains what the fixture covers, when upstream supplies one.
	Description string `json:"description"`
	// HTML is the expected output, already trimmed by GOV.UK Frontend's fixture build.
	HTML string `json:"html"`
}

// FixtureSet is a component's whole fixtures.json file.
type FixtureSet struct {
	Component string    `json:"component"`
	Fixtures  []Fixture `json:"fixtures"`
}

// FixtureComponents returns the component names under componentsDir that ship a fixtures.json,
// sorted. componentsDir is the installed package's dist/govuk/components directory.
//
// The list comes from the installed package rather than from a hard-coded slice, so upgrading
// GOV.UK Frontend surfaces a new component as a test failure instead of passing unnoticed.
func FixtureComponents(componentsDir string) ([]string, error) {
	entries, err := os.ReadDir(componentsDir)
	if err != nil {
		return nil, fmt.Errorf("govuk: reading %s: %w", componentsDir, err)
	}
	var names []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join(componentsDir, entry.Name(), "fixtures.json")); err != nil {
			continue
		}
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	return names, nil
}

// LoadFixtures reads one component's fixtures.json from the installed GOV.UK Frontend package.
func LoadFixtures(componentsDir, component string) (*FixtureSet, error) {
	path := filepath.Join(componentsDir, component, "fixtures.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("govuk: reading %s: %w", path, err)
	}
	set := &FixtureSet{}
	if err := json.Unmarshal(raw, set); err != nil {
		return nil, fmt.Errorf("govuk: parsing %s: %w", path, err)
	}
	return set, nil
}
