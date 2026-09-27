# Example service

**Apply for a rod fishing licence** is the reference GOV.UK service in this repository. It is an example. It does not take payment, send email, or issue a licence.

Pages are **Go**. Component HTML comes from **native Go renderers** that track GOV.UK Frontend macros and match every official fixture. The pin is **6.5.1**. See [tech-stack.md](tech-stack.md).

## Run it

```sh
npm install
npm start
```

Opens at <http://127.0.0.1:3000>. Set `PORT` to use another port.

`NODE_ENV=production` hides the component catalogue and the extra example pages. The licence journey stays available.

## Start to confirmation

The journey is one service, from the start page through to confirmation.

1. Start at `/` (English) or `/cy` (Welsh start page only). Choose **Start now**.
2. The task list at `/task-list` links to each question.
3. Answer the questions in order: name, date of birth, email, contact preference, where you will fish, licence length, start month, address, evidence (optional), additional details (optional), and password.
4. Check your answers at `/check-answers`. Change links return to a question and then come back.
5. Submit. The confirmation page at `/confirmation` shows a reference. The password is not shown.

Invalid answers stay on the same question, with an error summary and the values you entered. You cannot open confirmation until the required questions are complete.

## Pages

| Path                                 | What it shows                                                                                                                                          |
| ------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `/` and `/cy`                        | Start page. Welsh is the start page and chrome only; the rest of the journey is in English. When demos are on, links to the component preview homepage |
| `/task-list`                         | Task list, then the questions, check your answers, and confirmation                                                                                    |
| `/fees`, `/help`, `/guidance`        | Fees table, help accordion, and guidance tabs                                                                                                          |
| `/updates`, `/cookies`               | Service updates with pagination, and cookie settings                                                                                                   |
| `/accessibility`, `/about`           | Accessibility statement and what this example is                                                                                                       |
| `/components`                        | **Component preview homepage** — every component in this Frontend release as **links only** (no embedded demos)                                        |
| `/components/:name`                  | One component’s own page: Go `Render` HTML for a fixture, with a banner saying whether it matches the official fixture                                 |
| `/components/:name?fixture=`         | A named fixture on that component page                                                                                                                 |
| `/components/:name/fixture?fixture=` | The fixture HTML fragment only. For tests and debugging                                                                                                |
| `/examples/exit-this-page`           | Exit this page. The button leaves this example and opens the BBC weather forecast                                                                      |

Question pages use one `h1`, `novalidate`, an error summary, and field errors. Answers are kept when validation fails. A page uses a back link or breadcrumbs, not both.

## Responses

Pages and assets use the shared [baseline](frontend-security.md). Public HTML that sets the session cookie is `private, no-cache`, with a strong `ETag`. Question, task list, check your answers, confirmation, and cookie settings pages are `no-store`. The compiled Sass stylesheet (`application.css`), Frontend script, and the external `initAll()` module are fingerprinted and cached as immutable. The `js-enabled` snippet is the one line hashed in `baseline/policy.json`.

The server compresses with Brotli when the browser sends `Accept-Encoding: br`. Gzip is only used when the browser does not advertise `br`. Local `npm start` is HTTP, so the session cookie is not `Secure` and responses do not send HSTS.

## Tests

```sh
npm test
```

This runs the shared baseline suite and the Sass pipeline tests at **100%** line, branch, and function coverage, compiles the stylesheet, then runs `go test ./...`. Application packages are held to **100%** function, branch, and statement coverage. `cmd/server` is the process entry and is excluded.

Component tests render **every** official fixture shipped with the pinned `govuk-frontend` release, including hidden fixtures. The comparison is Go `Render` output against the fixture `html` string. The renderer trims only the outer whitespace of its own output so that output can equal the fixture. Tests do not edit fixture HTML and do not normalise it before comparing.

The service tests walk the licence journey, including validation, retained answers, check your answers, confirmation, cookies, Welsh, file upload, and the catalogue.

## Limits

- Sessions are stored in memory and end when the process stops.
- The password is checked and then discarded. It is not stored or shown again.
- A cookie choice is stored. This example does not set analytics cookies.
- An upload stores the file name only, and only for PDF, PNG, or JPG.
- Using this repo does not make a service assessment-ready. See [service-assessment-readiness.md](service-assessment-readiness.md).
