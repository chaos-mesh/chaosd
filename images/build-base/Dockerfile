FROM golang:1.27.2-bookworm

ARG HTTPS_PROXY
ARG HTTP_PROXY

ENV http_proxy=$HTTP_PROXY
ENV https_proxy=$HTTPS_PROXY

RUN apt-get update && apt-get install -y --no-install-recommends python3 pkg-config libfuse-dev fuse \
    && rm -rf /var/lib/apt/lists/*
