package render

// Renderer interface and Func adapter for tests.

import "errors"

// Renderer turns a GOV.UK Frontend component name and its macro options into HTML.
//
// name is the kebab-case component directory name, such as "date-input". params matches the
// options object the component's Nunjucks macro accepts, so the output can be compared with the
// official fixtures byte for byte.
type Renderer interface {
	Render(name string, params map[string]any) (string, error)
}

// Func adapts an ordinary function to [Renderer].
type Func func(name string, params map[string]any) (string, error)

// Render calls f.
func (f Func) Render(name string, params map[string]any) (string, error) {
	return f(name, params)
}

// ErrUnavailable reports that no component implementation has been wired in.
var ErrUnavailable = errors.New("render: no GOV.UK Frontend component renderer is configured")

// Unavailable returns a [Renderer] that fails every call with [ErrUnavailable].
//
// It exists so the service fails loudly at the first component instead of quietly serving a page
// with missing markup.
func Unavailable() Renderer {
	return Func(func(string, map[string]any) (string, error) { return "", ErrUnavailable })
}
