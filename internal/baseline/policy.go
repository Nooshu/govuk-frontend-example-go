// Package baseline applies the shared response policy from baseline/policy.json.
//
// policy.json is synced from the language-agnostic template and is the single source of OWASP
// header values, CSP directives, cache kinds, cookie defaults, and the hash of the js-enabled
// snippet GOV.UK Frontend's page template inlines. The Node modules next to it serve the shared
// test suite; this package is the Go reading of the same file, so the Go server never shells out
// to Node to answer a request.
package baseline

import (
	"bytes"
	"fmt"
	"os"

	"encoding/json/jsontext"
	jsonv2 "encoding/json/v2"
)

// Response kinds. Each one selects a Cache-Control value and, for documents, the full set of
// document security headers.
const (
	// KindDocument is public HTML.
	KindDocument = "document"
	// KindSensitiveDocument is HTML that shows someone's answers, so it is never stored.
	KindSensitiveDocument = "sensitive-document"
	// KindFingerprintedAsset is an asset whose URL changes when its bytes change.
	KindFingerprintedAsset = "fingerprinted-asset"
	// KindStaticAsset is an asset served from a stable URL, so it must be revalidated.
	KindStaticAsset = "static-asset"
	// KindDownload is a public file attachment.
	KindDownload = "download"
	// KindSensitiveDownload is a file attachment that is never stored.
	KindSensitiveDownload = "sensitive-download"
)

// Directive is one Content-Security-Policy directive and its sources.
//
// Directives keep the order they appear in policy.json, because the header is a single ordered
// string and reordering it would make the policy harder to compare with the source of truth.
type Directive struct {
	Name    string
	Sources []string
}

// Compression records the standard and fallback content codings.
type Compression struct {
	// Standard is the coding to use whenever the client advertises it.
	Standard string `json:"standard"`
	// Fallback is used only when the client does not advertise Standard.
	Fallback string `json:"fallback"`
}

// Policy is the parsed baseline/policy.json.
type Policy struct {
	// JSEnabledSnippet is the inline script GOV.UK Frontend's page template runs before anything
	// else. It must be emitted byte for byte or JSEnabledScriptHash stops matching.
	JSEnabledSnippet string `json:"jsEnabledSnippet"`
	// JSEnabledScriptHash is the CSP hash that allows JSEnabledSnippet without 'unsafe-inline'.
	JSEnabledScriptHash string `json:"jsEnabledScriptHash"`

	HSTS struct {
		MaxAge            int  `json:"maxAge"`
		IncludeSubDomains bool `json:"includeSubDomains"`
	} `json:"hsts"`

	Headers struct {
		All      map[string]string `json:"all"`
		Document map[string]string `json:"document"`
	} `json:"headers"`

	// Remove lists headers that must never reach the client, such as Server banners.
	Remove []string `json:"remove"`

	CacheControl map[string]string `json:"cacheControl"`
	ContentTypes map[string]string `json:"contentTypes"`

	CSP struct {
		Directives directives `json:"directives"`
	} `json:"csp"`

	PermissionsPolicy []string `json:"permissionsPolicy"`

	Performance struct {
		Compression Compression `json:"compression"`
	} `json:"performance"`

	Cookie struct {
		SameSite string `json:"sameSite"`
		Secure   bool   `json:"secure"`
		HTTPOnly bool   `json:"httpOnly"`
		Path     string `json:"path"`
	} `json:"cookie"`
}

// Load reads and parses a policy document.
func Load(path string) (*Policy, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("baseline: reading %s: %w", path, err)
	}
	return Parse(raw)
}

// Parse reads a policy document from JSON.
//
// It fails when a field the server depends on is missing, so a truncated or mis-synced
// policy.json is reported at start-up instead of producing responses with holes in them.
func Parse(raw []byte) (*Policy, error) {
	var policy Policy
	if err := jsonv2.Unmarshal(raw, &policy); err != nil {
		return nil, fmt.Errorf("baseline: parsing policy: %w", err)
	}
	switch {
	case policy.JSEnabledSnippet == "":
		return nil, fmt.Errorf("baseline: policy has no jsEnabledSnippet")
	case policy.JSEnabledScriptHash == "":
		return nil, fmt.Errorf("baseline: policy has no jsEnabledScriptHash")
	case len(policy.CacheControl) == 0:
		return nil, fmt.Errorf("baseline: policy has no cacheControl")
	case len(policy.CSP.Directives) == 0:
		return nil, fmt.Errorf("baseline: policy has no csp.directives")
	}
	return &policy, nil
}

// Directives returns the CSP directives in policy order.
func (p *Policy) Directives() []Directive {
	return append([]Directive(nil), p.CSP.Directives...)
}

// directives decodes a JSON object into ordered [Directive] values.
type directives []Directive

// keyOrder is objectKeys. Tests replace it to exercise the failure branch that a successful
// map decode and a second key-order decode cannot otherwise disagree on.
var keyOrder = objectKeys

func (d *directives) UnmarshalJSON(raw []byte) error {
	var sources map[string][]string
	if err := jsonv2.Unmarshal(raw, &sources); err != nil {
		return err
	}
	names, err := keyOrder(raw)
	if err != nil {
		return err
	}
	ordered := make([]Directive, 0, len(names))
	for _, name := range names {
		ordered = append(ordered, Directive{Name: name, Sources: sources[name]})
	}
	*d = ordered
	return nil
}

// objectKeys returns the keys of a JSON object in document order using jsontext.
//
// Non-string member names fail inside [jsontext.Decoder.ReadToken]; they never surface as a
// successful non-string token the way encoding/json's Token API could.
func objectKeys(raw []byte) ([]string, error) {
	dec := jsontext.NewDecoder(bytes.NewReader(raw))
	tok, err := dec.ReadToken()
	if err != nil {
		return nil, err
	}
	if tok.Kind().String() != "{" {
		return nil, fmt.Errorf("baseline: expected a JSON object")
	}
	var keys []string
	for dec.PeekKind().String() != "}" {
		key, err := dec.ReadToken()
		if err != nil {
			return nil, err
		}
		keys = append(keys, key.String())
		if err := dec.SkipValue(); err != nil {
			return nil, err
		}
	}
	_, err = dec.ReadToken() // consume '}'
	return keys, err
}
