# Deploying on Render.com

Host the Go example service on [Render](https://render.com) as a public demo (for example for a blog post). This line is a normal Go HTTP process — not a static site — so use a **Web Service**, not Static Sites / Cloudflare Pages.

Authoritative Render docs: [Language support](https://render.com/docs/language-support), [Blueprints](https://render.com/docs/blueprint-spec), [Deploy a Go app](https://render.com/docs/deploy-go-gin).

## What this repo already includes

| File                                                    | Role                                                      |
| ------------------------------------------------------- | --------------------------------------------------------- |
| [`render.yaml`](../render.yaml)                         | Blueprint: free web service, build/start, `/health` check |
| [`scripts/render-build.sh`](../scripts/render-build.sh) | `npm ci` → Sass → `go build -o bin/server`                |
| `cmd/server`                                            | Binds all interfaces by default; honours Render’s `PORT`  |
| `GET /health`                                           | Plain `ok` for Render health checks                       |

Build artefacts kept at runtime: `bin/server`, `node_modules/govuk-frontend`, `dist/stylesheets/application.css`.

## Prerequisites

1. A [GitHub](https://github.com) account with this repo (or a fork) pushed to `main`.
2. A [Render](https://render.com) account (free tier is enough for a demo).
3. Local checks green before you deploy: `npm ci && npm run verify` (optional but recommended).

## Option A — Blueprint (recommended)

Uses the committed [`render.yaml`](../render.yaml).

### Steps

1. **Push this repo to GitHub** (if it is not already), including `render.yaml` and `scripts/render-build.sh`.
2. Open the [Render Dashboard](https://dashboard.render.com/) and sign in (connect GitHub when asked).
3. Click **New +** → **Blueprint**.
4. Select the repository `Nooshu/govuk-frontend-example-go` (or your fork).
5. Confirm Render detects `render.yaml` at the repo root.
6. Review the service:
   - **Name:** `govuk-frontend-example-go`
   - **Runtime:** Go
   - **Plan:** Free (change later if you need always-on)
   - **Region:** Oregon (or pick one closer to you)
   - **Build:** `./scripts/render-build.sh`
   - **Start:** `./bin/server`
   - **Health check:** `/health`
7. Click **Apply** / **Create**.
8. Wait for the first deploy (install npm deps, compile Sass, build Go). Open the **Logs** tab if it fails.
9. When the service is **Live**, open the `.onrender.com` URL from the dashboard.

### After deploy checklist

- [ ] `https://<your-service>.onrender.com/health` returns `ok`
- [ ] Start page loads with GOV.UK styling (`/`)
- [ ] Component catalogue works (`/components`) — Blueprint sets `DEMOS_ENABLED=true`
- [ ] Start page shows **Developer previews** / Preview GOV.UK components
- [ ] A form POST in the licence journey retains the session (cookie)

## Option B — Manual Web Service

Use this if you prefer clicking through the UI without a Blueprint.

1. **New +** → **Web Service**.
2. Connect the GitHub repo and branch `main`.
3. Settings:

   | Field         | Value                       |
   | ------------- | --------------------------- |
   | Language      | Go                          |
   | Region        | Your choice                 |
   | Branch        | `main`                      |
   | Build Command | `./scripts/render-build.sh` |
   | Start Command | `./bin/server`              |
   | Instance type | Free                        |

4. **Advanced** → **Health Check Path:** `/health`
5. **Environment** (optional):

   | Key             | Value    | Notes                                                             |
   | --------------- | -------- | ----------------------------------------------------------------- |
   | `NODE_VERSION`  | `22`     | Matches `.nvmrc` / `package.json` engines                         |
   | `DEMOS_ENABLED` | `true`   | Keeps `/components` and Developer previews on (see below)         |
   | `NODE_ENV`      | _(omit)_ | Render may set `production`; demos still on via `DEMOS_ENABLED`   |
   | `HOST`          | _(omit)_ | Default binds all interfaces; only set if you need a special bind |
   | `PORT`          | _(omit)_ | Render injects this automatically                                 |

6. Create the service and wait for the deploy.

## Environment variables

| Variable        | Required | Default / behaviour                                            |
| --------------- | -------- | -------------------------------------------------------------- |
| `PORT`          | Injected | Render sets this; the server reads it via `config.ResolvePort` |
| `HOST`          | No       | Empty → listen on all interfaces (`:PORT`)                     |
| `DEMOS_ENABLED` | No       | Blueprint sets `true` so catalogue/previews stay on            |
| `NODE_ENV`      | No       | `production` turns demos off unless `DEMOS_ENABLED` overrides  |
| `NODE_VERSION`  | No       | Set to `22` in the Blueprint so npm tooling matches local      |

HTTPS terminates at Render. The app already treats `X-Forwarded-Proto: https` as secure (HSTS / `__Host-session` when appropriate).

## Free plan behaviour

- The service may **spin down** after idle time; the first request after idle can take ~30–60s (cold start). Mention that in a blog post if you link the demo.
- **In-memory sessions** reset when the instance restarts or scales — fine for an example, not for a real multi-instance service.
- Disk is ephemeral; do not rely on writing durable files.

## Local parity with Render

```sh
npm ci
npm run build:styles
go build -o bin/server ./cmd/server
PORT=3000 ./bin/server
# open http://127.0.0.1:3000
```

Or the one-shot script:

```sh
./scripts/render-build.sh && PORT=3000 ./bin/server
```

To mimic production (demos off):

```sh
NODE_ENV=production PORT=3000 ./bin/server
```

## Troubleshooting

| Symptom                                    | Likely fix                                                                                                                    |
| ------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------- |
| Build fails on `npm ci` / engine           | Ensure `NODE_VERSION=22`; lockfile committed; `engine-strict` in `.npmrc`                                                     |
| Build fails on Go version                  | Render uses latest stable Go 1.x; this module needs Go ≥1.27 — redeploy after Render updates, or switch the service to Docker |
| Deploy live but connection refused         | Confirm start command is `./bin/server` and listen addr is `:PORT` (not only `127.0.0.1`)                                     |
| HTML without GOV.UK CSS                    | Build must run `npm run build:styles`; `dist/stylesheets/` must exist at runtime                                              |
| Missing Frontend assets / fixtures         | Build must run `npm ci` so `node_modules/govuk-frontend` is present                                                           |
| `/components` 404 or no Developer previews | Set `DEMOS_ENABLED=true` (Blueprint default), or unset `NODE_ENV` if you are not using the override                           |
| Health check failing                       | Hit `/health` in logs; path must be exactly `/health`                                                                         |
| Session lost between requests              | Expected after free-tier spin-down; or cookie blocked if mixed content                                                        |

## Updating the live demo

Push to `main` (Blueprint auto-deploys by default). Watch the deploy in the Render dashboard. For a Frontend upgrade, follow [upgrading-govuk-frontend.md](upgrading-govuk-frontend.md) locally, push, then confirm the live pin and styles.

## Related

- [example-service.md](example-service.md) — what the demo contains
- [tech-stack.md](tech-stack.md) — Go + Node tooling roles
- [npm-security.md](npm-security.md) — lockfile / `npm ci` policy
- [frontend-security.md](frontend-security.md) — headers and cookies behind a reverse proxy
