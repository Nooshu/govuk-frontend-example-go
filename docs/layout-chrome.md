# Layout chrome

Every page must use the shared page template / layout — not a second full HTML document and not hand-written skip link / header / footer markup.

## Required pieces

| Piece              | Component          | Notes                                                                                                                             |
| ------------------ | ------------------ | --------------------------------------------------------------------------------------------------------------------------------- |
| Skip link          | Skip link          | First focusable control; target `#content` (or matching `main` id)                                                                |
| Masthead           | GOV.UK header      | Blue brand header; full logotype SVG (`fill="currentcolor"`, `<title>GOV.UK</title>`); `data-module="govuk-header"`               |
| Service name / nav | Service navigation | **Under** the masthead, inside `<header class="govuk-template__header">` — not inside the blue `govuk-header` block (Frontend 6+) |
| Footer             | GOV.UK footer      | OGL licence text + Crown copyright; `govuk-footer`                                                                                |
| Pattern demos      | Back link          | Shared “Back to patterns” in before-content                                                                                       |

Prefer static/shared chrome models (one skip link, header, footer definition) rather than rebuilding per request.

## Do

- Match markup from the pinned Frontend header/footer/page-template fixtures.
- Keep logotype SVG path data complete.
- Put the service name in service navigation.

## Do not

- Hand-write skip link / header / footer `govuk-*` chrome on pages.
- Use Frontend 5 header classes (`govuk-header__link--homepage`, `govuk-header__service-name` inside the blue bar).
- Add a black rectangle/banner or custom dark header instead of the brand blue masthead.
- Put the service name inside the blue `govuk-header` block.
- Use dots, squares, or placeholder Unicode instead of the crown / logotype SVG.

## Before content

Phase banner, breadcrumbs, **or** back link live in before-content inside the width container — **never breadcrumbs and back link together**.

## Related

- [page-shell.md](page-shell.md)
- [accessibility.md](accessibility.md)
- Header / service navigation / footer component docs on the Design System site
