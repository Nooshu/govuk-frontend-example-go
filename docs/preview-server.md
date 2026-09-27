# Preview server

Local server for human parity checks and pattern demos.

## Status

**Commands TBD** until [tech-stack.md](tech-stack.md) is filled in. Document the idiomatic way to start the preview app for the chosen language there (task runner, CLI, IDE run config — whatever best practice for that stack is).

## Expectations

- Homepage lists components (and patterns) as **links only** — no embedded live demos.
- A preview surface per component renders **only the selected** fixture, with a parity banner vs official `html`.
- A raw-fixture surface returns an HTML **fragment** for automation.
- Preview and fixture surfaces are Development / Testing only.
- Preview responses use the same [`baseline/`](../baseline/) headers as production. On local HTTP, pass `secureTransport: false` so HSTS is not sent.
- Syntax highlighting (if any) loads on Previews only — never on the global layout.
- Optional health / readiness endpoints follow the stack’s normal conventions; missing optional infra should not block Frontend-only preview.

## After code changes

Rebuild or reload as required by the language’s tooling; hard-refresh the browser. Confirm focus states, header/footer, and a failing-form example during visual QA after Frontend upgrades ([upgrading-govuk-frontend.md](upgrading-govuk-frontend.md)).
