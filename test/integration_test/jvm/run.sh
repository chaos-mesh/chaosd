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

set -eu

cur=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd $cur

bin_path=../../../bin

echo "download byteman example"
if [[ ! (-e byteman-example) ]]; then
    git clone https://github.com/WangXiangUSTC/byteman-example.git
fi

echo "download byteman && set environment variable"
byteman_dir="byteman-chaos-mesh-download-v4.0.18-0.9"
if [[ ! (-e ${byteman_dir}.tar.gz) ]]; then
    curl -fsSL -o ${byteman_dir}.tar.gz https://mirrors.chaos-mesh.org/${byteman_dir}.tar.gz
    tar zxvf ${byteman_dir}.tar.gz
fi
export BYTEMAN_HOME=$cur/${byteman_dir}
export PATH=$PATH:${BYTEMAN_HOME}/bin

echo "build && run Java example program helloworld"
cd byteman-example/example.helloworld
javac HelloWorld/Main.java
jar cfme HelloWorld.jar Manifest.txt HelloWorld.Main HelloWorld/Main.class
cd -
java -jar byteman-example/example.helloworld/HelloWorld.jar > helloworld.log &
# make sure it works
sleep 3
cat helloworld.log
# TODO: get the PID more accurately
pid=`pgrep -n java`

echo "run chaosd to inject failure into JVM, and check"

$bin_path/chaosd attack jvm return --class Main --method getnum --port 9288  --value 99999 --pid $pid
sleep 1
check_contains "99999" helloworld.log

$bin_path/chaosd attack jvm exception  --class Main --method sayhello --port 9288 --exception 'java.io.IOException("BOOM")' --pid $pid
sleep 1
check_contains "BOOM" helloworld.log

# TODO: add test for latency, stress and gc

echo "clean"
kill $pid
