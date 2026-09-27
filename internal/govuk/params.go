package govuk

// Ordered Params and jsontext decoding for macro options.

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"encoding/json/jsontext"
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
// the way JavaScript would have printed the value it parsed. Decoding uses
// [encoding/json/jsontext] (Go 1.25+) for streaming tokens.
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
	dec := jsontext.NewDecoder(bytes.NewReader(data))
	value, err := parseValue(dec)
	if err != nil {
		return nil, err
	}
	if _, err := dec.ReadToken(); err == nil {
		return nil, errors.New("govuk: unexpected data after JSON value")
	} else if !errors.Is(err, io.EOF) {
		return nil, err
	}
	return value, nil
}

func parseValue(dec *jsontext.Decoder) (any, error) {
	tok, err := dec.ReadToken()
	if err != nil {
		return nil, err
	}
	switch tok.Kind() {
	case jsontext.KindBeginObject:
		object := &Params{}
		for dec.PeekKind() != jsontext.KindEndObject {
			keyTok, err := dec.ReadToken()
			if err != nil {
				return nil, err
			}
			// Capture the name before parseValue advances the decoder; Tokens are voided afterward.
			name := keyTok.String()
			value, err := parseValue(dec)
			if err != nil {
				return nil, err
			}
			object.Set(name, value)
		}
		_, err := dec.ReadToken() // consume '}'
		return object, err
	case jsontext.KindBeginArray:
		items := []any{}
		for dec.PeekKind() != jsontext.KindEndArray {
			value, err := parseValue(dec)
			if err != nil {
				return nil, err
			}
			items = append(items, value)
		}
		_, err := dec.ReadToken() // consume ']'
		return items, err
	case jsontext.KindString:
		return tok.String(), nil
	case jsontext.KindNumber:
		// Preserve spelling (e.g. 1.50) for Nunjucks/JavaScript number stringification parity.
		return json.Number(tok.String()), nil
	case jsontext.KindTrue, jsontext.KindFalse:
		return tok.Bool(), nil
	default:
		// KindNull. jsontext.ReadToken errors on any other kind at value position, so a
		// successful token that is not handled above is always null.
		return nil, nil
	}
}
