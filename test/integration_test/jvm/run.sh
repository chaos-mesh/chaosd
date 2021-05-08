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

# download && build && run Java example program
git clone https://github.com/WangXiangUSTC/byteman-example.git
cd byteman-example/example.helloworld
javac HelloWorld/Main.java
jar cfme HelloWorld.jar Manifest.txt HelloWorld.Main HelloWorld/Main.class
cd -
java -jar byteman-example/example.helloworld/HelloWorld.jar > helloworld.log &
# TODO: get the PID more accurately
pid=`pidof java`

# download byteman && set environment variable
curl -fsSL -o chaosd-byteman-download.tar.gz https://mirrors.chaos-mesh.org/jvm/chaosd-byteman-download.tar.gz
tar zxvf chaosd-byteman-download.tar.gz
export BYTEMAN_HOME=$cur/chaosd-byteman-download
export PATH=$PATH:${BYTEMAN_HOME}/bin

$bin_path/chaosd attack jvm install --port 9288 --pid $pid

$bin_path/chaosd attack jvm submit return --class Main --method getnum --port 9288  --value 99999
check_contains "99999" helloworld.log

$bin_path/chaosd attack jvm submit exception  --class Main --method sayhello --port 9288 --exception 'java.io.IOException("BOOM")'
check_contains "BOOM" helloworld.log

# TODO: add test for latency, stress and gc

# clean
kill $pid