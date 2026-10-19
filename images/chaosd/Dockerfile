FROM pingcap/chaos-binary AS binary
FROM debian:bookworm-slim

ARG HTTPS_PROXY
ARG HTTP_PROXY

ENV http_proxy=$HTTP_PROXY
ENV https_proxy=$HTTPS_PROXY

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates tzdata iptables ipset stress-ng iproute2 net-tools iperf fuse util-linux \
    && rm -rf /var/lib/apt/lists/*
RUN update-alternatives --set iptables /usr/sbin/iptables-legacy

ENV RUST_BACKTRACE=1

COPY --from=binary /bin/chaosd /usr/local/bin/chaosd
COPY --from=binary /bin/tools /usr/local/bin/tools
