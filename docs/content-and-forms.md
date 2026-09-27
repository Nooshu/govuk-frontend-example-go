# Content and forms

## Content rules

| Rule          | Practice                                                                  |
| ------------- | ------------------------------------------------------------------------- |
| Sentence case | Headings, labels, buttons — never title case                              |
| Active voice  | “Apply for your licence” not “A licence can be applied for”               |
| Front-loaded  | Most important words first                                                |
| Button text   | Action-specific — “Save and continue”, “Submit application”               |
| Errors        | Clear and actionable — “Enter your email address” not “Email is required” |
| Hints         | Examples/context without duplicating the label                            |

## Form patterns

Every form should include:

1. Clear page title — `h1` for the specific task
2. Fieldset legends for related inputs
3. Label–input pairs — never placeholders as labels
4. Hint text where users need help
5. Server-side validation with linked **Error summary**
6. Single primary action — one green Continue/Submit per page
7. Progress indication on multi-page journeys
8. `novalidate` on forms (use GOV.UK error patterns, not browser bubbles)
9. Retained values after validation failure

## Do’s and don’ts

**Do**

- Use Frontend components via the library API
- Follow the page template and associated style pages when unsure
- Test keyboard-only start to finish
- Prefer two-thirds layout for content-heavy pages
- Keep form pages to one thing per page where possible

**Don’t**

- Add custom styling where GOV.UK patterns exist
- Use colour alone to convey information
- Invent new validation UX — use Error summary + error message
- Hide important actions or information
- Use placeholders instead of labels/hints
- Paste Design System HTML into pages
- Combine breadcrumbs and a back link
- Add decorative elements that do not serve the task
- Use more than one primary button per page
- Replace the GOV.UK header with a black bar or placeholder crown characters
