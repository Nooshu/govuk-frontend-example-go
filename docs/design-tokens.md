# Design tokens

Use GOV.UK Frontend classes and Sass tokens — **do not invent a parallel CSS system**. Service adjustments go in [`govuk-overrides.scss`](../styles/govuk-overrides.scss) via specificity (never `!important`). See [styles.md](styles.md).

## Colours

| Category   | Token               | Hex       | Notes                    |
| ---------- | ------------------- | --------- | ------------------------ |
| Text       | text                | `#0b0c0c` | Primary body text        |
|            | secondary-text      | `#484949` | Secondary text           |
|            | inverse-text        | `#ffffff` | Text on dark backgrounds |
| Link       | link                | `#1a65a6` | Default                  |
|            | link-hover          | `#0f385c` | Hover                    |
|            | link-visited        | `#54319f` | Visited                  |
|            | link-active         | `#0b0c0c` | Active                   |
| Border     | border              | `#cecece` | Standard                 |
|            | input-border        | `#0b0c0c` | Inputs                   |
| Background | template-background | `#f4f8fb` | `html` background        |
|            | body-background     | `#ffffff` | `body` background        |
| Focus      | focus               | `#ffdd00` | Keyboard focus only      |
|            | focus-text          | `#0b0c0c` | Text on focused elements |
| Error      | error               | `#ca3535` | Errors                   |
| Success    | success             | `#0f7a52` | Success                  |
| Hover      | hover               | `#cecece` | Input hover              |
| Brand      | brand               | `#1d70b8` | Primary brand            |
| Surface    | surface-background  | `#f4f8fb` | Surfaces                 |
|            | surface-text        | `#0b0c0c` | Text on surfaces         |
|            | surface-border      | `#8eb8dc` | Surface borders          |

Official reference: [https://design-system.service.gov.uk/styles/colour/](https://design-system.service.gov.uk/styles/colour/)

## Typography

- **Display / body:** GDS Transport from Frontend assets
- **Code:** system monospace
- Weights: regular (400) and bold (700) only

Type scale (approximate): Display 48px, Headline 36px, Section 24px, Subhead/Body 19px, Small 16px, Caption 14px.

Official reference: [https://design-system.service.gov.uk/styles/typography/](https://design-system.service.gov.uk/styles/typography/)

## Spacing

- Base unit: 4px
- Scale: 5, 10, 15, 20, 25, 30, 40, 50, 60px
- Component padding: small 10×15, medium 15×20, large 20×30
- Section spacing: 30px mobile, 40px tablet, 50px desktop
- Container max width: 960px; horizontal padding 15px mobile / 30px desktop
- Form groups: 20px between groups; 30px before submit

## Border radius

**0px** — sharp corners for consistency and accessibility. No rounded elements except where browsers require them.

## Elevation

Minimal shadows. Cards/panels: 1px borders. Focus: thick yellow borders (4px), not shadows. Prefer Frontend component styles over custom elevation.

## Component appearance notes

Agents should implement appearance **only** via components (button variants, error summary, phase banner, etc.), Frontend override classes, or cascade rules in `govuk-overrides.scss` — not by re-stating hex values with `!important`. Examples:

- Start button: `govuk-button govuk-button--start` + start icon SVG from Frontend template
- Primary / secondary / warning buttons via Frontend classes
- Error summary, inset text (`role="note"` where the component does), phase banner, breadcrumbs, footer OGL text — all via components

See [govuk-components.md](govuk-components.md) and [layout-chrome.md](layout-chrome.md).
