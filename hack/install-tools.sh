#!/usr/bin/env bash

# Copyright 2020 Chaos Mesh Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# See the License for the specific language governing permissions and
# limitations under the License.

set -euo pipefail

target_os=${1:?expected target OS}
target_arch=${2:?expected target architecture}
destination=${3:?expected output directory}
if [[ "$target_os" != linux ]]; then
    echo "Bundled tools require a Linux target" >&2
    exit 1
fi
case "$target_arch" in
    amd64) arch=x86_64 ;;
    arm64) arch=aarch64 ;;
    *) echo "Unsupported architecture: $target_arch" >&2; exit 1 ;;
esac

cache=${CHAOSD_TOOLS_CACHE:-${XDG_CACHE_HOME:-$HOME/.cache}/chaosd}
mkdir -p "$cache" "$destination"
staging=$(mktemp -d "$destination/.download.XXXXXX")
trap 'rm -rf "$staging"' EXIT

download() {
    local url=$1 archive=$2
    if [[ ! -s "$archive" ]]; then
        local temporary
        temporary=$(mktemp "$archive.XXXXXX")
        if ! curl --fail --location --retry 3 --output "$temporary" "$url"; then
            rm -f "$temporary"
            return 1
        fi
        mv "$temporary" "$archive"
    fi
}

stress_version=v0.14.02
mem_version=v0.3.1
tproxy_version=v0.5.6
byteman_version=v4.0.20-0.12

if [[ ! -x "$destination/stress-ng" || $(cat "$destination/.stress-ng-version" 2>/dev/null || true) != "$stress_version-$arch" ]]; then
    archive="$cache/stress-ng-$stress_version-$arch"
    download "https://github.com/chaos-mesh/stress-ng/releases/download/$stress_version/stress-ng-$arch" "$archive"
    install -m 755 "$archive" "$destination/stress-ng"
    echo "$stress_version-$arch" > "$destination/.stress-ng-version"
fi

if [[ ! -x "$destination/memStress" || $(cat "$destination/.memStress-version" 2>/dev/null || true) != "$mem_version-$arch" ]]; then
    archive="$cache/memStress_$mem_version-$arch-linux-gnu.tar.gz"
    download "https://github.com/chaos-mesh/memStress/releases/download/$mem_version/$(basename "$archive")" "$archive"
    tar -xzf "$archive" -C "$staging"
    install -m 755 "$staging/memStress" "$destination/memStress"
    echo "$mem_version-$arch" > "$destination/.memStress-version"
fi

if [[ ! -x "$destination/tproxy" || $(cat "$destination/.tproxy-version" 2>/dev/null || true) != "$tproxy_version-$arch" ]]; then
    archive="$cache/tproxy-$tproxy_version-$arch.tar.gz"
    download "https://github.com/chaos-mesh/chaos-tproxy/releases/download/$tproxy_version/tproxy-$arch.tar.gz" "$archive"
    tar -xzf "$archive" -C "$staging"
    install -m 755 "$staging/tproxy" "$destination/tproxy"
    echo "$tproxy_version-$arch" > "$destination/.tproxy-version"
fi

if [[ ! -f "$destination/byteman/lib/byteman.jar" || $(cat "$destination/.byteman-version" 2>/dev/null || true) != "$byteman_version" ]]; then
    directory="byteman-chaos-mesh-download-$byteman_version"
    archive="$cache/$directory.tar.gz"
    download "https://mirrors.chaos-mesh.org/$directory.tar.gz" "$archive"
    tar -xzf "$archive" -C "$staging"
    rm -rf "$destination/byteman"
    mv "$staging/$directory" "$destination/byteman"
    echo "$byteman_version" > "$destination/.byteman-version"
fi
