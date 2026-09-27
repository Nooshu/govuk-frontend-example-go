# Go conventions

Coding conventions for this repository. Follow [Effective Go](https://go.dev/doc/effective_go), the [Code Review Comments](https://go.dev/wiki/CodeReviewComments) wiki, and the pins in [tech-stack.md](tech-stack.md). Agents should treat this as the comment and layout contract.

## Module and layout

- Module path: `github.com/Nooshu/govuk-frontend-example-go`
- `go 1.27.0` + `toolchain` line in `go.mod` (CI uses `go-version-file: go.mod`)
- Commands under `cmd/…`; non-exportable libraries under `internal/…`
- One package per directory; package name matches the directory
- Prefer the standard library; record third-party exceptions in [tech-stack.md](tech-stack.md)

## Formatting and lint

```sh
gofmt -w .
npm run lint:go    # go vet + go fix -diff + go tool staticcheck
```

Do not bypass `gofmt`. Prefer `go fix` modernizers when upgrading the toolchain. `staticcheck` is pinned via the `tool` directive in `go.mod`.

Editor defaults for Go live in [`.vscode/settings.json`](../.vscode/settings.json) (gopls, format on save with `gofmt`). Optional live reload: [`.air.toml`](../.air.toml). Staticcheck defaults: [`staticcheck.conf`](../staticcheck.conf).

## Comments (Go’s standard practice)

Go comments are for **readers of the API and non-obvious intent**, not a line-by-line narration of obvious code.

| What                      | Rule                                                                                                      |
| ------------------------- | --------------------------------------------------------------------------------------------------------- |
| Package                   | Every package has a package comment (prefer `doc.go`, or the primary file) starting with `Package name …` |
| Exported identifiers      | Doc comment immediately above; start with the name (`// Render returns …`)                                |
| Unexported helpers        | Comment when the why is non-obvious; skip restating the code                                              |
| Commands (`package main`) | File comment explaining how to run and what the process owns                                              |
| Complex algorithms        | Explain invariants, ordering requirements, or security properties                                         |
| Fixture / Nunjucks parity | Call out whitespace, escape, or attribute-order constraints that look “odd” in Go                         |

**Do:**

```go
// Render returns the HTML for one GOV.UK Frontend component.
//
// component is the kebab-case directory name used by GOV.UK Frontend.
func Render(component string, params *Params) (string, error)
```

**Don’t:**

```go
// This function renders the component.
// Loop through the params.
// Then return the string.
func Render(...)
```

Security-sensitive code (CSRF compare, cookie flags, CSP hash, body limits) should say **why** the approach is required, with a pointer to [frontend-security.md](frontend-security.md) when useful.

## Errors

- Return `error` values; wrap with `%w` when adding context (`fmt.Errorf("app: …: %w", err)`)
- Fail at start-up when policy, stylesheet, or Frontend assets are missing — not on the first request
- Do not panic except for programmer mistakes in internal literal builders (for example `NewParams` with odd arity)

## Testing

- Table-driven tests with `t.Run`; use `t.Helper` on helpers
- `t.Parallel` only when shared state is safe (avoid mutating package-level hooks in parallel)
- Prefer `httptest` recorders for handlers; `httptest.NewTestServer` + `testing/synctest` for client/timeout cases
- Fixture HTML: ordinal string equality, no normalisation
- Coverage gate: 100% statements on `./internal/...` (`npm run test:go`)

## JSON

- Prefer `encoding/json/v2` for ordinary structs (policy, config meta)
- Ordered component options use `encoding/json/jsontext` via `govuk.Params` — maps alone lose key order
- If marshaling maps into HTML for tests, pass `jsonv2.Deterministic(true)` so ETags stay stable

## Concurrency

- Protect shared maps with `sync.Mutex` (see `session.Store`)
- Prefer `synctest` for deterministic concurrent unit tests
- HTTP handlers must be safe for concurrent use; do not put request state on `App` without synchronisation

## Generated and vendored paths

Do not edit:

- official fixture `html` inside `node_modules/govuk-frontend`
- compiled `dist/stylesheets/` by hand (rebuild with Sass)
- synced `baseline/` policy values to “make tests pass” — fix the Go implementation instead

## Related

- [architecture.md](architecture.md) — package map
- [why-go.md](why-go.md) — why this language line exists
- [testing-components.md](testing-components.md) — parity suite
- [Effective Go](https://go.dev/doc/effective_go)
- [Code Review Comments](https://go.dev/wiki/CodeReviewComments)
