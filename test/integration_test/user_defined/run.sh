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

# test user defined attack
${bin_path}/chaosd attack user-defined --attack-cmd "touch /tmp/chaos" --recover-cmd "rm /tmp/chaos" --uid 1234

if [ -e /tmp/chaos ]
then
    echo "ok"
else
    echo "file not exist"
    exit 1
fi

${bin_path}/chaosd recover 1234

if [ -e /tmp/chaos ]
then
    echo "file still exist"
    exit 1
fi

exit 0
