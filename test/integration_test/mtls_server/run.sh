#!/usr/bin/env bash

# Copyright 2021 Chaos Mesh Authors.
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

set -u

cur=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd $cur

bin_path=../../../bin

RED='\033[0;31m'
NC='\033[0m' # No Color
HTTPS_PORT=31768
ENDPOINT="https://localhost:$HTTPS_PORT/api/experiments"

function failtest() {
    msg="$1"
    kill $server_pid
    echo -e "${RED}FAIL: $msg$NC"
    exit 1
}

function request() {
    cert_prefix=$1
    result_success=$2
    cmd="curl -Lskw \n%{http_code} $ENDPOINT"
    if [[ -n "$cert_prefix" ]]; then
        cmd="$cmd --cert client/${cert_prefix}_cert.pem --key client/${cert_prefix}_key.pem"
    fi
    response=$($cmd)
    body=$(echo "$response" | head -n1)
    status=$(echo "$response" | tail -n1)
    if [[ "$result_success" == "true" ]] && [[ "$status" != "200" ]]; then
        failtest "expected 200, got $status"
    elif [[ "$result_success" != "true" ]] && [[ "$status" == "200" ]]; then
        failtest "expected not 200, got $status"
    fi
}

echo "Generating certificates"
bash +ex ./gen_certs.sh

echo "Starting Server in mTLS mode"
${bin_path}/chaosd server \
    --https-port $HTTPS_PORT \
    --cert ./server/server_cert.pem \
    --key ./server/server_key.pem \
    --CA ./server/server_cert.pem \
> server.out &

server_pid=$!

sleep 2

if ! kill -0 $server_pid > /dev/null 2>&1; then
    echo -e "${RED}ERROR: Couldn't start server$NC"
    cat server.out
    exit 1
fi

echo -n "Test with valid certificate... "
request 'valid' true
echo "Passed"

echo -n "Test with no certificate... "
request '' false
echo "Passed"

echo -n "Test with invalid certificate... "
request 'invalid' false
echo "Passed"

kill $server_pid
rm -f server.out
rm -rf client server