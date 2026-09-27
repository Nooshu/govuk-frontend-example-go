package govuk

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseJSONRejectsMalformedInput(t *testing.T) {
	for _, raw := range []string{`{"a":true`, `{`, `[1,`, `]`, `{"a":`} {
		if _, err := parseJSON([]byte(raw)); err == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
	if _, err := parseJSON([]byte(`{"a":1} trailing`)); err == nil {
		t.Fatal("accepted trailing data")
	}
	if _, err := parseJSON([]byte(`true`)); err != nil {
		t.Fatalf("literal true: %v", err)
	}
	if got, err := parseJSON([]byte(`null`)); err != nil || got != nil {
		t.Fatalf("literal null: got %#v err=%v", got, err)
	}
}

func TestHelperEdges(t *testing.T) {
	if got := at([]any{"a", "b"}, 1); got != "b" {
		t.Fatalf("at = %#v", got)
	}
	if at([]any{"a"}, 3) != Undefined || at([]any{"a"}, -1) != Undefined || at("no", 0) != Undefined {
		t.Fatal("out-of-range at was not undefined")
	}

	if str(nil) != "" || str(Undefined) != "" || str(Safe("x")) != "x" || str(true) != "true" {
		t.Fatal("str mismatch")
	}
	if str([]any{"a", true}) != "a,true" || str(NewParams("a", 1)) != "[object Object]" {
		t.Fatal("str composite mismatch")
	}
	if number(json.Number("1.50")) != "1.5" || number(json.Number("1e2")) != "100" || number(json.Number("bad")) != "bad" {
		t.Fatalf("number: %s %s %s", number(json.Number("1.50")), number(json.Number("1e2")), number(json.Number("bad")))
	}
	if out(Safe("<b>")) != "<b>" || !strings.Contains(out("a'b"), "&#39;") {
		t.Fatal("out mismatch")
	}
	if indent("", 2, true) != "" || indent("a\nb", 2, true) != "  a\n  b" || indent("a\nb", 2, false) != "a\n  b" {
		t.Fatal("indent mismatch")
	}

	if length(nil) != 0 || length(false) != 0 || length(true) != 0 || length("ab") != 2 || length(Safe("ab")) != 2 {
		t.Fatal("length mismatch")
	}
	if length([]any{1, 2}) != 2 || length(NewParams("a", 1, "b", 2)) != 2 || length(1) != 0 {
		t.Fatal("length composite mismatch")
	}

	if !truthy(json.Number("2")) || truthy(json.Number("0")) || truthy(json.Number("bad")) {
		t.Fatal("numeric truthiness mismatch")
	}

	if !looseEq(nil, Undefined) || looseEq(nil, "x") || !looseEq(true, true) || looseEq(true, false) {
		t.Fatal("looseEq nil/bool mismatch")
	}
	if !looseEq(true, json.Number("1")) || !looseEq(json.Number("0"), false) || !looseEq(json.Number("1"), "1") {
		t.Fatal("looseEq number coercion mismatch")
	}
	if looseEq(json.Number("1"), "nope") || !looseEq("1", json.Number("1")) || looseEq("  ", json.Number("1")) {
		t.Fatal("looseEq string/number mismatch")
	}
	if !looseEq(json.Number("1.0"), json.Number("1")) || looseEq("a", "b") {
		t.Fatal("looseEq equality mismatch")
	}

	if !strictEq(Undefined, Undefined) || strictEq(Undefined, nil) || !strictEq(nil, nil) || strictEq(nil, "x") {
		t.Fatal("strictEq nil mismatch")
	}
	if !strictEq(json.Number("1"), json.Number("1.0")) || strictEq(json.Number("1"), "1") {
		t.Fatal("strictEq number mismatch")
	}
	if !strictEq(true, true) || strictEq(true, "true") || !strictEq("a", "a") || strictEq("a", "b") {
		t.Fatal("strictEq scalar mismatch")
	}
	params := NewParams("a", 1)
	if !strictEq(params, params) || strictEq(params, NewParams("a", 1)) {
		t.Fatal("strictEq object mismatch")
	}

	if !contains("b", "abc") || !contains("b", Safe("abc")) || contains("z", "abc") {
		t.Fatal("contains string mismatch")
	}
	if !contains(json.Number("1"), []any{json.Number("1")}) || contains("missing", []any{"a"}) {
		t.Fatal("contains array mismatch")
	}
	if !contains("a", NewParams("a", 1)) || contains("a", 1) {
		t.Fatal("contains object mismatch")
	}

	if !sameNumber("0", "  ") || !sameNumber("1", "1.0") || sameNumber("1", "nope") {
		t.Fatal("sameNumber mismatch")
	}
	if boolToNumber(true) != json.Number("1") || boolToNumber(false) != json.Number("0") {
		t.Fatal("boolToNumber mismatch")
	}
}

func TestParamsEdges(t *testing.T) {
	var missing *Params
	if missing.Get("a") != Undefined || missing.Has("a") || missing.Keys() != nil || missing.Len() != 0 {
		t.Fatal("nil params mismatch")
	}

	params := NewParams("b", 1, "a", nil)
	params.Set("b", 2)
	if params.Get("b") != 2 || !params.Has("a") || params.Get("missing") != Undefined || strings.Join(params.Keys(), ",") != "b,a" {
		t.Fatalf("%#v", params.Keys())
	}

	mustPanic(t, "odd", func() { NewParams(oddPairs()...) })
	mustPanic(t, "key", func() { NewParams(1, "value") })

	var decoded Params
	if err := json.Unmarshal([]byte(`{"z":1,"items":[{"text":"A"}],"flag":true,"empty":null}`), &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Has("empty") || decoded.Get("empty") != nil || decoded.Keys()[0] != "z" {
		t.Fatalf("decoded keys = %v", decoded.Keys())
	}
	for _, raw := range []string{`[]`, `1`, `{"a":1} trailing`, `{`, `]`, `{"a":`, `[1,`} {
		var target Params
		if err := json.Unmarshal([]byte(raw), &target); err == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
	if err := new(Params).UnmarshalJSON([]byte(`]`)); err == nil {
		t.Fatal("UnmarshalJSON accepted a closing bracket")
	}
	if _, err := parseJSON([]byte(`{"a":true`)); err == nil {
		t.Fatal("unclosed object accepted")
	}
	if _, err := parseJSON([]byte(`[true`)); err == nil {
		t.Fatal("unclosed array accepted")
	}
	direct := []string{``, `]`, `}`, `{"a":1} 2`, `{`, `{"a":`, `[`, `[1,`, `{,}`, `{true}`, `[}`, `{"a":1,}`}
	for _, raw := range direct {
		if _, err := parseJSON([]byte(raw)); err == nil {
			t.Fatalf("parseJSON accepted %q", raw)
		}
	}
}

func TestAttributesAndSmallRenderers(t *testing.T) {
	if Attributes(Safe(" data-x")) != " data-x" || Attributes(" already") != " already" || Attributes(1) != "" {
		t.Fatal("Attributes passthrough mismatch")
	}
	rendered := Attributes(NewParams(
		"data-safe", Safe("<ok>"),
		"open", NewParams("value", true, "optional", true),
		"skip", NewParams("value", false, "optional", true),
		"data-text", "a'b",
	))
	if !strings.Contains(rendered, " data-safe=\"<ok>\"") || !strings.Contains(rendered, " open") || strings.Contains(rendered, "skip") {
		t.Fatalf("attributes = %q", rendered)
	}
	if i18nAttributes("count", Undefined, "not-an-object") != "" {
		t.Fatal("non-object i18n messages were rendered")
	}
	if i18nAttributes("count", "one", Undefined) == "" || i18nAttributes("count", Undefined, Undefined) != "" {
		t.Fatal("i18n message mismatch")
	}
	if appendedAttributes("nope") != "" || !strings.Contains(appendedAttributes(NewParams("data-x", "1")), ` data-x="1"`) {
		t.Fatal("appended attributes mismatch")
	}
	if slotContent(NewParams("text", "before"), 2, false) == "" {
		t.Fatal("slot text was empty")
	}
	if capitalise("") != "" || capitalise("DAY") != "Day" {
		t.Fatal("capitalise mismatch")
	}
	if !strings.Contains(paginationLinkLabel(NewParams("html", " <b>On</b> "), "Next"), "On") {
		t.Fatal("pagination html label missing")
	}
	link := summaryActionLink(
		NewParams("href", "/change", "text", "Change"),
		NewParams("html", "<b>Name</b>", "text", "Name"),
	)
	if !strings.Contains(link, "Name") {
		t.Fatalf("action link = %s", link)
	}
}

func TestRenderBranchesOutsideFixtures(t *testing.T) {
	label := NewParams("text", "Label")
	cases := []struct {
		name   string
		params *Params
		want   string
	}{
		{"input", NewParams("id", "n", "name", "n", "label", label, "formGroup", NewParams("beforeInput", NewParams("text", "before"))), "before"},
		{"textarea", NewParams("id", "n", "name", "n", "label", label, "formGroup", NewParams("beforeInput", NewParams("html", "<b>before</b>"))), "before"},
		{"select", NewParams("id", "n", "name", "n", "label", label, "formGroup", NewParams("beforeInput", NewParams("text", "before"), "afterInput", NewParams("text", "after"))), "after"},
		{"file-upload", NewParams("id", "n", "name", "n", "label", label, "formGroup", NewParams("beforeInput", NewParams("text", "before"), "afterInput", NewParams("text", "after"))), "after"},
		{"checkboxes", NewParams("name", "r", "fieldset", NewParams("legend", NewParams("text", "Choose")), "formGroup", NewParams("beforeInputs", NewParams("text", "before"), "afterInputs", NewParams("text", "after"))), "after"},
		{"radios", NewParams("name", "r", "fieldset", NewParams("legend", NewParams("text", "Choose")), "formGroup", NewParams("beforeInputs", NewParams("text", "before"), "afterInputs", NewParams("html", "<b>after</b>"))), "after"},
		{"date-input", NewParams("id", "dob", "fieldset", NewParams("legend", NewParams("text", "Date")), "items", []any{NewParams("id", "extra")}, "formGroup", NewParams("beforeInputs", NewParams("text", "before"), "afterInputs", NewParams("text", "after"))), "after"},
		{"character-count", NewParams("id", "d", "name", "d", "label", label, "maxlength", json.Number("10"), "charactersUnderLimitText", "not-an-object", "formGroup", NewParams("afterInput", NewParams("html", "<span>after</span>"), "attributes", NewParams("data-x", "1"))), "after"},
		{"character-count", NewParams("id", "d", "name", "d", "label", label, "maxwords", json.Number("5"), "formGroup", NewParams("afterInput", NewParams("text", "after text"))), "after text"},
		{"password-input", NewParams("id", "p", "name", "p", "label", label, "formGroup", NewParams("afterInput", NewParams("html", "<span>after</span>"))), "after"},
		{"password-input", NewParams("id", "p", "name", "p", "label", label, "formGroup", NewParams("afterInput", NewParams("text", "after text"))), "after text"},
		{"footer", NewParams("navigation", []any{NewParams("title", "Section", "items", []any{NewParams("text", "No href"), NewParams("href", "/x")})}), "Section"},
		{"pagination", NewParams("next", NewParams("href", "/2", "html", "<b>Onward</b>")), "Onward"},
		{"summary-list", NewParams(
			"card", NewParams("title", NewParams("html", "<b>Card</b>")),
			"rows", []any{
				false,
				NewParams(
					"key", NewParams("text", "Name"),
					"value", NewParams("text", "Ada"),
					"actions", NewParams("items", []any{NewParams("href", "/change", "text", "Change")}),
				),
			},
		), "Card"},
	}
	for _, tc := range cases {
		html, err := Render(tc.name, tc.params)
		if err != nil {
			t.Fatalf("Render(%s): %v", tc.name, err)
		}
		if !strings.Contains(html, tc.want) {
			t.Fatalf("Render(%s) missing %q\n%s", tc.name, tc.want, html)
		}
	}

	if !strings.Contains(MustRender("back-link", NewParams("text", "Back", "href", "/")), "govuk-back-link") {
		t.Fatal("MustRender did not render a back link")
	}
	mustPanic(t, "unknown component", func() { MustRender("not-a-component", nil) })
	if _, err := Render("not-a-component", nil); err == nil {
		t.Fatal("unknown component was rendered")
	}
}

func TestFixtureLoadingEdges(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "orphan"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "button"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "button", "fixtures.json"), []byte(`{"component":"button","fixtures":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	names, err := FixtureComponents(dir)
	if err != nil || len(names) != 1 || names[0] != "button" {
		t.Fatalf("names = %v %v", names, err)
	}
	if _, err := FixtureComponents(filepath.Join(dir, "missing")); err == nil {
		t.Fatal("missing directory was listed")
	}
	if _, err := LoadFixtures(dir, "button"); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadFixtures(dir, "orphan"); err == nil {
		t.Fatal("missing fixtures file was loaded")
	}
	if err := os.WriteFile(filepath.Join(dir, "button", "fixtures.json"), []byte(`{"fixtures":`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadFixtures(dir, "button"); err == nil {
		t.Fatal("invalid fixtures were loaded")
	}
}

func mustPanic(t *testing.T, name string, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatalf("%s did not panic", name)
		}
	}()
	fn()
}

// oddPairs returns a one-element slice so NewParams can be called with an odd number of
// arguments without staticcheck SA5012 flagging a literal variadic call site.
func oddPairs() []any { return []any{"only"} }
