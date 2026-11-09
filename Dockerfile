# Cloud Run image.
# Stage 1: Node — pin govuk-frontend and compile Sass.
# Stage 2: Go — static server binary.
# Stage 3: Runtime — binary plus Frontend assets and baseline policy.

FROM node:22-bookworm AS styles
WORKDIR /build
COPY package.json package-lock.json ./
RUN npm ci
COPY styles ./styles
COPY scripts ./scripts
RUN npm run build:styles

FROM golang:1.27-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -trimpath -o /out/server ./cmd/server

FROM debian:bookworm-slim
WORKDIR /app
RUN apt-get update \
  && apt-get install -y --no-install-recommends ca-certificates \
  && rm -rf /var/lib/apt/lists/*
COPY --from=build /out/server /app/server
COPY --from=styles /build/node_modules/govuk-frontend ./node_modules/govuk-frontend
COPY --from=styles /build/dist ./dist
COPY package.json ./
COPY baseline ./baseline
ENV DEMOS_ENABLED=true
EXPOSE 8080
CMD ["/app/server"]
