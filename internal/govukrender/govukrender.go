// Package govukrender connects the page layer to the Go port of GOV.UK Frontend's macros.
//
// It is deliberately the only package that knows about both sides. Pages depend on the
// [render.Renderer] interface, so the component port can be tested against the official
// fixtures on its own, and pages can be tested without rendering real components.
package govukrender

import (
	"maps"
	"slices"

	"github.com/Nooshu/govuk-frontend-example-go/internal/govuk"
	"github.com/Nooshu/govuk-frontend-example-go/internal/render"
)

// New returns the [render.Renderer] backed by GOV.UK Frontend's components.
//
// It adapts the plain option maps pages build into the ordered [govuk.Params] the component
// port takes. Keys are converted in sorted order, which is safe because the only option whose
// key order reaches the HTML is `attributes`, and pages in this service set attributes through
// named component options instead.
func New() render.Renderer {
	return render.Func(func(name string, params map[string]any) (string, error) {
		return govuk.Render(name, toParams(params))
	})
}

func toParams(params map[string]any) *govuk.Params {
	if params == nil {
		return nil
	}
	converted := &govuk.Params{}
	for _, key := range slices.Sorted(maps.Keys(params)) {
		converted.Set(key, toValue(params[key]))
	}
	return converted
}

func toValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return toParams(typed)
	case []any:
		items := make([]any, 0, len(typed))
		for _, item := range typed {
			items = append(items, toValue(item))
		}
		return items
	case []map[string]any:
		items := make([]any, 0, len(typed))
		for _, item := range typed {
			items = append(items, toParams(item))
		}
		return items
	default:
		return value
	}
}
