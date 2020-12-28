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

# test cpu stress
./bin/chaosd attack stress cpu -l 10 -w 1 > cpu.out

uid=`cat cpu.out | grep "successfully" | awk -F: '{print $2}'`

num=`ps aux | grep stress-ng`
echo num

./bin/chaosd recover ${uid}

num=`ps aux | grep stress-ng`
echo num

# test mem stress
./bin/chaosd attack stress mem -w 1 > mem.out

uid=`cat cpu.out | grep "successfully" | awk -F: '{print $2}'`

num=`ps aux | grep stress-ng`
echo num

./bin/chaosd recover ${uid}

num=`ps aux | grep stress-ng`
echo num