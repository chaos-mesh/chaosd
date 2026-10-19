# syntax=docker/dockerfile:1

FROM pingcap/chaos-build-base AS go_build

WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .

ARG SWAGGER
ARG LDFLAGS
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/.cache/chaosd \
    SWAGGER=$SWAGGER IMG_LDFLAGS=$LDFLAGS make binary

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates curl tar \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /bin
COPY --from=go_build /src/bin/chaosd /bin/chaosd
COPY --from=go_build /src/bin/tools /bin/tools
