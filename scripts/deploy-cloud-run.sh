#!/usr/bin/env bash
# Build this repo with Cloud Build and deploy the Docker image to Cloud Run.
# Step-by-step: docs/deploying-on-cloud-run.md

set -euo pipefail

PROJECT="${PROJECT:-}"
if [[ -z "$PROJECT" ]]; then
  PROJECT="$(gcloud config get-value project 2>/dev/null || true)"
fi
if [[ -z "$PROJECT" || "$PROJECT" == "(unset)" ]]; then
  echo "Set PROJECT or run: gcloud config set project PROJECT_ID" >&2
  exit 1
fi

REGION="${REGION:-europe-west2}"
SERVICE="${SERVICE:-govuk-frontend-example-go}"
MEMORY="${MEMORY:-512Mi}"
CPU="${CPU:-1}"
CONCURRENCY="${CONCURRENCY:-80}"
MIN_INSTANCES="${MIN_INSTANCES:-0}"
MAX_INSTANCES="${MAX_INSTANCES:-3}"
ENV_VARS="${ENV_VARS:-DEMOS_ENABLED=true}"
PROBE_PATH="${PROBE_PATH:-/health}"

echo "Enabling Cloud Run, Cloud Build, and Artifact Registry on ${PROJECT}"
gcloud services enable \
  run.googleapis.com \
  cloudbuild.googleapis.com \
  artifactregistry.googleapis.com \
  --project "$PROJECT"

project_number="$(gcloud projects describe "$PROJECT" --format='value(projectNumber)')"
runtime_sa="${RUNTIME_SA:-${project_number}-compute@developer.gserviceaccount.com}"

echo "Granting roles/run.builder to ${runtime_sa}"
for attempt in 1 2 3 4 5 6; do
  if gcloud projects add-iam-policy-binding "$PROJECT" \
    --member="serviceAccount:${runtime_sa}" \
    --role="roles/run.builder" \
    --quiet >/dev/null; then
    break
  fi
  if [[ "$attempt" -eq 6 ]]; then
    echo "Could not grant roles/run.builder" >&2
    exit 1
  fi
  sleep "$attempt"
done

echo "Deploying ${SERVICE} in ${REGION}"
gcloud run deploy "$SERVICE" \
  --source . \
  --project "$PROJECT" \
  --region "$REGION" \
  --service-account "$runtime_sa" \
  --allow-unauthenticated \
  --ingress=all \
  --port 8080 \
  --memory "$MEMORY" \
  --cpu "$CPU" \
  --concurrency "$CONCURRENCY" \
  --timeout 60 \
  --min-instances "$MIN_INSTANCES" \
  --max-instances "$MAX_INSTANCES" \
  --cpu-boost \
  --set-env-vars "$ENV_VARS" \
  --startup-probe="initialDelaySeconds=0,timeoutSeconds=4,periodSeconds=5,failureThreshold=36,httpGet.path=${PROBE_PATH},httpGet.port=8080" \
  --liveness-probe="timeoutSeconds=4,periodSeconds=30,failureThreshold=3,httpGet.path=${PROBE_PATH},httpGet.port=8080" \
  --readiness-probe="timeoutSeconds=4,periodSeconds=5,failureThreshold=3,httpGet.path=${PROBE_PATH},httpGet.port=8080"

url="$(gcloud run services describe "$SERVICE" --project "$PROJECT" --region "$REGION" --format='value(status.url)')"
echo "Deployed ${url}"
echo "Check ${url}${PROBE_PATH}"
