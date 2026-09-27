# Accessibility

## Baseline

- **WCAG 2.2 AA** for all components and patterns.
- Introduction: [Making your service accessible](https://www.gov.uk/service-manual/helping-people-to-use-your-service/making-your-service-accessible-an-introduction).
- Design System overview: [Accessibility](https://design-system.service.gov.uk/accessibility/).

**Using the Design System or this repo does not automatically make a service accessible.** Still need user research (including disabled users), an accessibility audit, an accessibility statement by public beta, and assisted digital planning.

## Non-negotiables in this repo

| Requirement             | Detail                                                                                                                                                    |
| ----------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Keyboard                | All interactive elements usable keyboard-only; verify skip link and tab order                                                                             |
| Focus                   | 4px yellow focus from Frontend — [focus states](https://design-system.service.gov.uk/get-started/focus-states/); never `outline: none` hacks              |
| Contrast                | 4.5:1 normal text; 3:1 large text (Frontend tokens)                                                                                                       |
| Touch                   | ≥ 40px minimum interactive targets                                                                                                                        |
| Headings                | One `h1`; sensible hierarchy                                                                                                                              |
| Landmarks               | Skip link → header → before-content → `main` → footer                                                                                                     |
| ARIA                    | Only via components / documented pattern usage — do not sprinkle ARIA by hand                                                                             |
| Progressive enhancement | Core journey without JS; Frontend JS via `type="module"` + `initAll()`; [browser support](https://frontend.design-system.service.gov.uk/browser-support/) |
| Labels / legends        | [Labels, legends and headings](https://design-system.service.gov.uk/get-started/labels-legends-headings/)                                                 |
| Errors                  | Error summary linking to fields + field error messages                                                                                                    |
| Plain English           | Reading age ~9 where possible — [content-and-forms.md](content-and-forms.md)                                                                              |

## Page template

Follow [page-shell.md](page-shell.md) and the official [page template](https://design-system.service.gov.uk/styles/page-template/).

## Assessment

Align UI work with [service-assessment-readiness.md](service-assessment-readiness.md) (especially Service Standard points 4, 5, 13, 14).
