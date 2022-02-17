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

set -eu

cur=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd $cur

bin_path=../../../bin

chaos_test_file="/tmp/chaos-test"

echo "create file"
${bin_path}/chaosd attack file create --file-name ${chaos_test_file} --uid 12345

if [[ ! (-e ${chaos_test_file}) ]]; then
    echo "${chaos_test_file} not exists"
    exit 1
fi

${bin_path}/chaosd recover 12345

if [[ (-e ${chaos_test_file}) ]]; then
    echo "${chaos_test_file} exists after recover"
    exit 1
fi

echo "create directory"
${bin_path}/chaosd attack file create --dir-name ${chaos_test_file} --uid 12345

if [[ ! (-e ${chaos_test_file}) ]]; then
    echo "${chaos_test_file} not exists"
    exit 1
fi

${bin_path}/chaosd recover 12345

if [[ (-e ${chaos_test_file}) ]]; then
    echo "${chaos_test_file} exists after recover"
    exit 1
fi

echo "delete file"
touch ${chaos_test_file}
${bin_path}/chaosd attack file delete --dir-name ${chaos_test_file} --uid 12345

if [[ (-e ${chaos_test_file}) ]]; then
    echo "${chaos_test_file} exists after delete"
    exit 1
fi

${bin_path}/chaosd recover 12345

if [[ ! (-e ${chaos_test_file}) ]]; then
    echo "${chaos_test_file} not exists after recover"
    exit 1
fi


echo "append file"
touch ${chaos_test_file}

${bin_path}/chaosd attack file append --file-name ${chaos_test_file} --data "chaos-mesh" --count 5 --uid 12345

num=`cat ${chaos_test_file} | grep "chaos-mesh" | wc -l`
if [[ ${num} -ne 5 ]]; then
    echo "append file failed"
    exit 1
fi

${bin_path}/chaosd recover 12345
num=`cat ${chaos_test_file} | grep "chaos-mesh" | wc -l`
if [[ ${num} -ne 0 ]]; then
    echo "recover append file failed"
    exit 1
fi

echo "modify file"
${bin_path}/chaosd attack file modify --file-name ${chaos_test_file} --privilege 777 --uid 12345

privilege=`stat -c %a ${chaos_test_file}`
if [[ ${privilege} -ne 777 ]]; then
    echo "modify file failed"
    exit 1
fi

${bin_path}/chaosd recover 12345
privilege=`stat -c %a ${chaos_test_file}`
if [[ ${privilege} -ne 664 ]]; then
    echo "recover modify file failed"
    exit 1
fi

echo "rename file"

${bin_path}/chaosd attack file rename --source-file ${chaos_test_file} --dest-file ${chaos_test_file}-bak --uid 12345
if [[ (-e ${chaos_test_file}) ]]; then
    echo "${chaos_test_file} exists after rename"
    exit 1
fi

if [[ ! (-e ${chaos_test_file}-bak) ]]; then
    echo "${chaos_test_file}-bak not exists after rename"
    exit 1
fi

${bin_path}/chaosd recover 12345

if [[ ! (-e ${chaos_test_file}) ]]; then
    echo "${chaos_test_file} not exists after recover"
    exit 1
fi

if [[ (-e ${chaos_test_file}-bak) ]]; then
    echo "${chaos_test_file}-bak exists after recover"
    exit 1
fi

exit 0
rm ${chaos_test_file}