#!/usr/bin/env bash

# Copyright 2026 Chaos Mesh Authors.
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

image=${1:?expected runtime image}
docker run --rm -i --network none "$image" sh -eu -s <<'EOF'
chaosd version
chaosd --help
/usr/local/bin/tools/FileTool --help
/usr/local/bin/tools/PortOccupyTool --help
test "$(command -v stress-ng)" = /usr/local/bin/tools/stress-ng
stress-ng --version

case "$(dpkg --print-architecture)" in
    amd64) expected_machine=62 ;;
    arm64) expected_machine=183 ;;
    *) echo "Unsupported image architecture" >&2; exit 1 ;;
esac
for executable in /usr/local/bin/chaosd /usr/local/bin/tools/FileTool \
    /usr/local/bin/tools/PortOccupyTool /usr/local/bin/tools/stress-ng \
    /usr/local/bin/tools/memStress /usr/local/bin/tools/tproxy; do
    test -x "$executable"
    machine=$(od -An -tu2 -j18 -N2 "$executable" | tr -d ' ')
    test "$machine" = "$expected_machine"
    libraries=$(ldd "$executable" 2>&1) || true
    case "$libraries" in
        *"not found"*) printf '%s\n%s\n' "$executable" "$libraries" >&2; exit 1 ;;
    esac
done

test -s /usr/local/bin/tools/byteman/lib/byteman.jar
test -x /usr/local/bin/tools/byteman/bin/bminstall.sh
test -s /etc/ssl/certs/ca-certificates.crt
test -f /usr/share/zoneinfo/UTC
for command in tc ipset iptables ifconfig iperf stress-ng fusermount fallocate; do
    command -v "$command"
done
test "$(readlink /etc/alternatives/iptables)" = /usr/sbin/iptables-legacy
test "$RUST_BACKTRACE" = 1
test ! -e /usr/local/bin/chaosd.dat
EOF
