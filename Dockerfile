# syntax=docker/dockerfile:1

FROM golang:1.27.2-bookworm AS builder

ARG HTTPS_PROXY
ARG HTTP_PROXY

ENV http_proxy=$HTTP_PROXY
ENV https_proxy=$HTTPS_PROXY

RUN apt-get update && apt-get install -y --no-install-recommends python3 pkg-config libfuse-dev fuse \
    && rm -rf /var/lib/apt/lists/*

FROM builder AS go_build

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

FROM scratch AS binary

WORKDIR /bin
COPY --from=go_build /src/bin/chaosd /bin/chaosd
COPY --from=go_build /src/bin/tools /bin/tools

FROM debian:bookworm-slim AS runtime

ARG HTTPS_PROXY
ARG HTTP_PROXY

ENV http_proxy=$HTTP_PROXY
ENV https_proxy=$HTTPS_PROXY

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates tzdata iptables ipset iproute2 net-tools iperf fuse util-linux \
    && rm -rf /var/lib/apt/lists/*
RUN update-alternatives --set iptables /usr/sbin/iptables-legacy

ENV RUST_BACKTRACE=1
ENV PATH="/usr/local/bin/tools:${PATH}"

COPY --from=binary /bin/chaosd /usr/local/bin/chaosd
COPY --from=binary /bin/tools /usr/local/bin/tools
