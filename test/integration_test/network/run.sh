#!/usr/bin/env bash

# Copyright 2022 Chaos Mesh Authors.
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

# test nic down
nic=$(cat /proc/net/dev | awk '{i++; if(i>2){print $1}}' | sed 's/^[\t]*//g' | sed 's/[:]*$//g' | head -n 1)

ifconfig ${nic}:0 192.168.10.28 up

test_nic=$(ifconfig | grep ${nic}:0 | sed 's/:0:.*/:0/g')
if [[ test_nic == "" ]]; then
    echo "create test nic failed"
    exit 1
fi

${bin_path}/chaosd attack network down -d ${test_nic} --duration 1s
stat=$(ifconfig | grep ${test_nic} | sed 's/:0:.*/:0/g')

if [[ -n ${stat} ]]; then
    echo "fail to disable the test nic"
    ifconfig ${test_nic} down
    exit 1
fi

sleep 1s

stat=$(ifconfig | grep ${test_nic} | sed 's/:0:.*/:0/g')
if [[ ${stat} != ${test_nic} ]]; then
    echo "fail to enable the test nic"
    exit 1
fi

ifconfig ${test_nic} down

exit 0