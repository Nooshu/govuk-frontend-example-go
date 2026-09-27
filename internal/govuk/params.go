// Package govuk renders GOV.UK Frontend components to HTML in Go.
//
// GOV.UK Frontend ships its components as Nunjucks macros. This package is a native Go port
// of those macros: no Node process is started and no JavaScript template engine is embedded.
// The parity gate is the official fixtures.json shipped with each release — [Render] must
// return byte-for-byte the same HTML as the fixture for every fixture of every component.
//
// Because the macros were written for a JavaScript template engine, this package reproduces
// the handful of JavaScript and Nunjucks behaviours the macros rely on:
//
//   - object key order is significant, because it decides attribute order, so options are
//     held in an ordered [Params] rather than a Go map;
//   - `undefined` (absent) and `null` are different values, because several macros test for
//     one but not the other;
//   - text is escaped exactly as Nunjucks escapes it, including the backslash;
//   - HTML options are trusted and emitted unescaped, matching `| safe` upstream.
//
// The entry points are [Render], [Names], and [LoadFixtures].
package govuk

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// Params is an ordered set of component options, the Go equivalent of the object literal a
// Nunjucks macro receives.
//
// Order is preserved because GOV.UK Frontend renders the `attributes` option by iterating its
// keys, so a Go map would produce attributes in a different order on every run and break
// fixture parity.
//
// A nil *Params behaves as an empty object: reading any key returns [Undefined]. That mirrors
// Nunjucks, where `params.formGroup.classes` is simply undefined when `formGroup` is absent.
type Params struct {
	keys   []string
	values map[string]any
}

// NewParams builds a parameter object from alternating key and value arguments.
//
// It panics when the arguments are not pairs or a key is not a string, because every call site
// is a literal in this package rather than user input.
func NewParams(pairs ...any) *Params {
	if len(pairs)%2 != 0 {
		panic("govuk: NewParams needs an even number of arguments")
	}
	p := &Params{}
	for i := 0; i < len(pairs); i += 2 {
		key, ok := pairs[i].(string)
		if !ok {
			panic(fmt.Sprintf("govuk: NewParams key %d is not a string", i))
		}
		p.Set(key, pairs[i+1])
	}
	return p
}

// Set stores a value, appending the key the first time it is seen so insertion order is kept.
func (p *Params) Set(key string, value any) {
	if p.values == nil {
		p.values = make(map[string]any)
	}
	if _, seen := p.values[key]; !seen {
		p.keys = append(p.keys, key)
	}
	p.values[key] = value
}

// Get returns the value stored under key, or [Undefined] when the key is absent.
//
// An explicit JSON null is returned as nil, which is a different value from [Undefined].
func (p *Params) Get(key string) any {
	if p == nil {
		return Undefined
	}
	value, ok := p.values[key]
	if !ok {
		return Undefined
	}
	return value
}

// Has reports whether key is present, including when its value is null.
func (p *Params) Has(key string) bool {
	if p == nil {
		return false
	}
	_, ok := p.values[key]
	return ok
}

// Keys returns the keys in insertion order.
func (p *Params) Keys() []string {
	if p == nil {
		return nil
	}
	return p.keys
}

// Len returns the number of keys, which is what Nunjucks' `length` filter reports for an object.
func (p *Params) Len() int {
	if p == nil {
		return 0
	}
	return len(p.keys)
}

// UnmarshalJSON decodes a JSON object while preserving key order.
//
// Numbers are kept as [encoding/json.Number] so they are rendered with their original spelling,
// the way JavaScript would have printed the value it parsed.
func (p *Params) UnmarshalJSON(data []byte) error {
	value, err := parseJSON(data)
	if err != nil {
		return err
	}
	decoded, ok := value.(*Params)
	if !ok {
		return fmt.Errorf("govuk: expected a JSON object, got %T", value)
	}
	*p = *decoded
	return nil
}

// undefinedValue is the type of [Undefined]. It is distinct from nil so that a missing option
// and an option explicitly set to null can be told apart.
type undefinedValue struct{}

// Undefined is the value of an option that was never supplied, matching JavaScript `undefined`.
var Undefined any = undefinedValue{}

// Safe is a string that is already HTML and must be emitted without escaping. It is the Go
// equivalent of a Nunjucks SafeString, produced by `| safe` and by captured `{% set %}` blocks.
type Safe string

// parseJSON decodes JSON into the value model this package uses: *Params for objects, []any for
// arrays, string, bool, [encoding/json.Number], and nil for null.
func parseJSON(data []byte) (any, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	value, err := parseValue(dec)
	if err != nil {
		return nil, err
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return nil, errors.New("govuk: unexpected data after JSON value")
	}
	return value, nil
}

// tokenDecoder is the part of [json.Decoder] parseValue uses. Tests supply a scripted
// decoder for token sequences the standard library never yields, such as a non-string object key.
type tokenDecoder interface {
	Token() (json.Token, error)
	More() bool
}

func parseValue(dec tokenDecoder) (any, error) {
	token, err := dec.Token()
	if err != nil {
		return nil, err
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return token, nil
	}
	switch delim {
	case '{':
		object := &Params{}
		for dec.More() {
			key, err := dec.Token()
			if err != nil {
				return nil, err
			}
			name, ok := key.(string)
			if !ok {
				return nil, fmt.Errorf("govuk: object key is not a string: %v", key)
			}
			value, err := parseValue(dec)
			if err != nil {
				return nil, err
			}
			object.Set(name, value)
		}
		if _, err := dec.Token(); err != nil {
			return nil, err
		}
		return object, nil
	case '[':
		items := []any{}
		for dec.More() {
			value, err := parseValue(dec)
			if err != nil {
				return nil, err
			}
			items = append(items, value)
		}
		if _, err := dec.Token(); err != nil {
			return nil, err
		}
		return items, nil
	default:
		return nil, fmt.Errorf("govuk: unexpected JSON delimiter %v", delim)
	}
}
