# Deploying on Google Cloud Run

Host this Go example on [Cloud Run](https://cloud.google.com/run/docs). The binary listens on `0.0.0.0:$PORT` with an IPv4 socket (`tcp4`). Cloud Run’s default port is 8080. TLS stops at the platform, which sets `X-Forwarded-Proto`.

You need a Google Cloud project with billing enabled and the [gcloud CLI](https://cloud.google.com/sdk/docs/install). Cloud Run can scale to zero. Build minutes, Artifact Registry storage, and requests above the free tier are billed.

This service’s name is `govuk-frontend-example-go`, region `europe-west2` (London).

## What this repo includes

| File                                                            | Role                                                         |
| --------------------------------------------------------------- | ------------------------------------------------------------ |
| [`Dockerfile`](../Dockerfile)                                   | Node Sass, then a static Go binary, plus Frontend assets     |
| [`scripts/deploy-cloud-run.sh`](../scripts/deploy-cloud-run.sh) | Enables APIs and deploys from source                         |
| `GET /health`                                                   | Plain `ok` for startup, readiness, and liveness              |

## Deploy

From the repo root, with a project selected:

```sh
gcloud config set project PROJECT_ID
./scripts/deploy-cloud-run.sh
```

| Variable        | Default                     |
| --------------- | --------------------------- |
| `PROJECT`       | gcloud’s active project     |
| `REGION`        | `europe-west2`              |
| `SERVICE`       | `govuk-frontend-example-go` |
| `MEMORY`        | `512Mi`                     |
| `CPU`           | `1`                         |
| `MIN_INSTANCES` | `0`                         |
| `MAX_INSTANCES` | `3`                         |

The script sets `DEMOS_ENABLED=true`, so the component catalogue stays on. `NODE_ENV=production` hides the catalogue unless that flag is set.

### After deploy checklist

- [ ] `https://<service>-<hash>.<region>.run.app/health` returns `ok`
- [ ] Start page loads with GOV.UK styling
- [ ] `/components` works
- [ ] A form POST in the licence journey keeps the answers on the next page

## Run the image locally

```sh
docker build -t govuk-frontend-example-go .
docker run --rm -p 8080:8080 -e PORT=8080 govuk-frontend-example-go
```

Open <http://127.0.0.1:8080>.
