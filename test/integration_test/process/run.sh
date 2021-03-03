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

set -u

cur=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd $cur

bin_path=../../../bin

${bin_path}/dummy > dummy.out &

sleep 1

pid=$(cat dummy.out)

echo ${pid}

# test stop attack
${bin_path}/chaosd attack process stop -p ${pid} > proc.out

sleep 1

uid=$(cat proc.out | grep "Attack process ${pid} successfully" | awk -F: '{print $2}')

stat=$(ps o pid,s | grep ${pid} | awk '{print $2}')

if [[ ${stat} != "T" ]]; then
    echo "target process is not stopped by processed stop attack"
    exit 1
fi

# test recover from stop attack
${bin_path}/chaosd recover ${uid}

sleep 1

stat=$(ps o pid,s | grep ${pid} | awk '{print $2}')

if [[ ${stat} != "R" ]]; then
    echo "target process is not resumed by recovering from process stop attack"
    exit 1
fi

# test kill attack
${bin_path}/chaosd attack process kill -p ${pid} > proc.out

sleep 1

stat=$(ps o pid | grep ${pid})

if [[ -n ${stat} ]]; then
    echo "target process is not killed by processed kill attack"
    exit 1
fi

rm dummy.out
rm proc.out

kill -- -0
