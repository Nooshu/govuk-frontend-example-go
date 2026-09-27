// Package config resolves the paths and settings the example service needs at run time.
//
// It finds the repository root, reads the pinned govuk-frontend version from
// package.json / node_modules, locates the compiled stylesheet, and interprets
// PORT and demo-related environment variables. Failures here should surface at
// process start, not on the first browser request.
package config
