package govukrender

// render.Renderer adapter over govuk.Render.

import (
	"maps"
	"slices"

	"github.com/Nooshu/govuk-frontend-example-go/internal/govuk"
	"github.com/Nooshu/govuk-frontend-example-go/internal/render"
)

// New returns the [render.Renderer] backed by GOV.UK Frontend's components.
//
// It adapts the plain option maps pages build into the ordered [govuk.Params] the component
// port takes. Keys are converted in sorted order so page HTML is deterministic. That is safe
// for service pages, which set attributes through named component options. Fixture previews
// must not use this path: they call [govuk.LoadFixtures] and [govuk.Render] directly so
// attribute key order and JSON number spelling stay identical to the parity suite.
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
