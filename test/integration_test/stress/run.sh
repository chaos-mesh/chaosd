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

set -eu

cur=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd $cur

bin_path=../../../bin

echo "download memStress && set environment variable"
if [[ ! (-e memStress_v0.3-x86_64-linux-gnu.tar.gz) ]]; then
    curl -fsSL -o memStress_v0.3-x86_64-linux-gnu.tar.gz https://github.com/chaos-mesh/memStress/releases/download/v0.3/memStress_v0.3-x86_64-linux-gnu.tar.gz
	tar zxvf memStress_v0.3-x86_64-linux-gnu.tar.gz
fi

# test cpu stress
${bin_path}/chaosd attack stress cpu -l 10 -w 1 > cpu.out

PID=`cat cpu.out | grep "stress-ng" | sed 's/.*Pid=\([0-9]*\).*/\1/g'`

stress_ng_num=`ps aux > test.temp && grep "stress-ng" test.temp | wc -l && rm test.temp`
if [ ${stress_ng_num} -lt 1 ]; then
    echo "stress-ng is not run when executing stress cpu attack"
    exit 1
fi

uid=`cat cpu.out | grep "Attack stress cpu successfully" | awk -F: '{print $2}'`
${bin_path}/chaosd recover ${uid}

echo "wait stress-ng $PID exit after recovering stress cpu attack"
timeout 5s tail --pid=$PID -f /dev/null

ps aux | grep stress-ng

# test mem stress
${bin_path}/chaosd attack stress mem --size 10M > mem.out

PID=`cat mem.out | grep "memStress" | sed 's/.*Pid=\([0-9]*\).*/\1/g'`

memStress_num=`ps aux > test.temp && grep "memStress" test.temp | wc -l && rm test.temp`
if [ ${memStress_num} -lt 1 ]; then
    echo "memStress is not run when executing stress mem attack"
    exit 1
fi

uid=`cat mem.out | grep "Attack stress mem successfully" | awk -F: '{print $2}'`
${bin_path}/chaosd recover ${uid}

echo "wait memStress $PID exit after recovering stress mem attack"
timeout 5s tail --pid=$PID -f /dev/null

ps aux | grep memStress

rm cpu.out
rm mem.out