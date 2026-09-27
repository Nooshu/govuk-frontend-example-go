#!/usr/bin/env bash
# Build the example service for Render (and similar hosts).
# Installs the Frontend pin, compiles Sass, then builds the Go binary.
set -euo pipefail

cd "$(dirname "$0")/.."

npm ci
npm run build:styles
go build -o bin/server ./cmd/server
