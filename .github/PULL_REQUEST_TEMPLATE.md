## Summary

<!-- 1–3 bullets: what and why -->

## Audience impact

- [ ] Docs updated for humans (`docs/`, `CONTRIBUTING.md`, and/or `README.md` as needed)
- [ ] Agent index / playbooks updated (`AGENTS.md` links and/or `/docs` playbook) if behaviour changed
- [ ] Dual-audience coverage matches [`docs/documentation-structure.md`](../docs/documentation-structure.md) (onboarding + agent contracts)

## Checklist

- [ ] No frontend UI frameworks introduced for GOV.UK chrome
- [ ] Prefer Nunjucks macros / **backend vs fixture** parity (not Nunjucks-only; no pasted release HTML as source of truth)
- [ ] `npm run verify:docs` passes (or N/A if docs untouched)
- [ ] Fixture HTML not edited to make tests pass
- [ ] If upgrading `govuk-frontend`: reviewed https://github.com/alphagov/govuk-frontend/releases/latest and followed `docs/upgrading-govuk-frontend.md`
- [ ] Coverage remains 100% functions / branches / statements for application code (when present)
- [ ] HTTP responses use `baseline/` (cache kind, CSP, OWASP headers) when the change serves pages or assets
- [ ] Styles come from the Sass pipeline (`styles/`); no `!important` in service CSS; `govuk-overrides.scss` stays last
- [ ] Code follows the recorded language’s latest best practices (`docs/tech-stack.md`)
- [ ] Finished work is split into focused commits with comprehensive messages when landing multiple concerns

## Test plan

- [ ] …
