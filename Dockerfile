# syntax=docker/dockerfile:1.7

FROM golang:1.25-bookworm AS build
WORKDIR /src

COPY go.work go.work.sum ./
COPY bridge ./bridge
COPY core ./core
COPY scale ./scale
COPY bot ./bot
COPY zebra ./zebra

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    cd scale && go build -o /out/scale .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    cd bot && go build -o /out/bot ./cmd/bot

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    cd zebra && go build -o /out/zebra .

FROM debian:bookworm-slim AS runtime
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    tzdata \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
RUN mkdir -p /app/bin /app/bot /tmp/gscale-zebra

COPY --from=build /out/scale /app/bin/scale
COPY --from=build /out/bot /app/bin/bot
COPY --from=build /out/zebra /app/bin/zebra
COPY bot/.env.example /app/bot/.env.example

ENV BRIDGE_STATE_FILE=/tmp/gscale-zebra/bridge_state.json

ENTRYPOINT ["/app/bin/scale"]
