# ---------------------------------------------------------------------------
# --- Multi-Stage: builder
# --- This stage will generate the executable / binary called "web".
# --- The last stage "runner" will copy this and use it as entrypoint.

FROM    golang:1.25.5-trixie AS builder
WORKDIR /build
COPY    go.mod .
COPY    go.sum .
RUN     go mod download
COPY    .      .
RUN     go build ./cmd/web


# ---------------------------------------------------------------------------
# --- Multi-Stage: bundler
# --- This stage will bundle the JavaScript (and TypeScript) and css files.
# --- The last stage "runner" will copy these and use them for the web app.

FROM    node:lts-alpine AS bundler
RUN     apk add make
WORKDIR /bundle

COPY ./package.json         .
COPY ./package-lock.json    .
RUN  npm install

COPY ./Makefile    .
COPY ./make        ./make
COPY ./ui          ui

# RUN make bundle


# ---------------------------------------------------------------------------
# --- Multi-Stage: runner

FROM debian:trixie-slim AS runner

# Setup a non-root user
RUN groupadd --system --gid 999 nonroot \
 && useradd --system --gid 999 --uid 999 --create-home nonroot

# install trusted root TLS certificates
RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy the application from the builder
COPY --from=builder --chown=nonroot:nonroot /build/web /app/web

# Copy the JS/TS and CSS bundles from the bundler
COPY --from=bundler --chown=nonroot:nonroot /bundle/ui/static/dist /app/ui/static/dist

RUN ln -s /app/web /usr/local/bin/web

# Use the non-root user to run the webs erver
USER nonroot

ENTRYPOINT ["web"]
