package app

import (
	"testing"

	"github.com/Nooshu/govuk-frontend-example-go/internal/govuk"
)

func TestSelectGovukFixture(t *testing.T) {
	hidden := govuk.Fixture{Name: "hidden", Hidden: true, HTML: "h"}
	visible := govuk.Fixture{Name: "default", HTML: "d"}

	if got, ok := selectGovukFixture([]govuk.Fixture{hidden, visible}, ""); !ok || got.Name != "default" {
		t.Fatalf("visible default = %#v %v", got, ok)
	}
	if got, ok := selectGovukFixture([]govuk.Fixture{hidden, visible}, "hidden"); !ok || !got.Hidden {
		t.Fatalf("named hidden = %#v %v", got, ok)
	}
	if _, ok := selectGovukFixture([]govuk.Fixture{hidden, visible}, "missing"); ok {
		t.Fatal("missing fixture was selected")
	}
	if got, ok := selectGovukFixture([]govuk.Fixture{hidden}, ""); !ok || got.Name != "hidden" {
		t.Fatalf("hidden fallback = %#v %v", got, ok)
	}
	if _, ok := selectGovukFixture(nil, ""); ok {
		t.Fatal("empty list selected a fixture")
	}
}

func TestGovukFixtureLinks(t *testing.T) {
	links := govukFixtureLinks([]govuk.Fixture{{Name: "a"}, {Name: "b"}}, "b")
	if len(links) != 2 || links[0].Current || !links[1].Current {
		t.Fatalf("%#v", links)
	}
}
