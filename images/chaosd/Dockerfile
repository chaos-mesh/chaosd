FROM debian:buster-slim

ARG HTTPS_PROXY
ARG HTTP_PROXY

ENV http_proxy $HTTP_PROXY
ENV https_proxy $HTTPS_PROXY

RUN apt-get update && apt-get install -y tzdata iptables ipset stress-ng iproute2 fuse util-linux && rm -rf /var/lib/apt/lists/*

RUN update-alternatives --set iptables /usr/sbin/iptables-legacy

ENV RUST_BACKTRACE 1

COPY --from=pingcap/chaos-binary /bin/chaosd /usr/local/bin/chaosd
