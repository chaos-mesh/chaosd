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

set -euo pipefail

function chaosd::version::get_version_vars() {
  if [[ -n ${GIT_COMMIT-} ]] || GIT_COMMIT=$(git rev-parse "HEAD^{commit}" 2>/dev/null); then

    # Use git describe to find the version based on tags.
    if [[ -n ${GIT_VERSION-} ]] || GIT_VERSION=$(git describe --tags --abbrev=14 "${GIT_COMMIT}^{commit}" 2>/dev/null); then
      DASHES_IN_VERSION=$(echo "${GIT_VERSION}" | sed "s/[^-]//g")
      if [[ "${DASHES_IN_VERSION}" == "---" ]] ; then
        # We have distance to subversion (v1.1.0-subversion-1-gCommitHash)
        GIT_VERSION=$(echo "${GIT_VERSION}" | sed "s/-\([0-9]\{1,\}\)-g\([0-9a-f]\{14\}\)$/.\1\+\2/")
      elif [[ "${DASHES_IN_VERSION}" == "--" ]] ; then
        # We have distance to base tag (v1.1.0-1-gCommitHash)
        GIT_VERSION=$(echo "${GIT_VERSION}" | sed "s/-g\([0-9a-f]\{14\}\)$/+\1/")
      fi
    fi
  fi
}

function chaosd::version::ldflag() {
  local key=${1}
  local val=${2}

  echo "-X 'github.com/chaos-mesh/chaosd/pkg/version.${key}=${val}'"
}

# Prints the value that needs to be passed to the -ldflags parameter of go build
function chaosd::version::ldflags() {
  chaosd::version::get_version_vars

  local buildDate=
  [[ -z ${SOURCE_DATE_EPOCH-} ]] || buildDate="--date=@${SOURCE_DATE_EPOCH}"
  local -a ldflags=($(chaosd::version::ldflag "buildDate" "$(date ${buildDate} -u +'%Y-%m-%dT%H:%M:%SZ')"))
  if [[ -n ${GIT_COMMIT-} ]]; then
    ldflags+=($(chaosd::version::ldflag "gitCommit" "${GIT_COMMIT}"))
  fi

  if [[ -n ${GIT_VERSION-} ]]; then
    ldflags+=($(chaosd::version::ldflag "gitVersion" "${GIT_VERSION}"))
  fi

  # The -ldflags parameter takes a single string, so join the output.
  echo "${ldflags[*]-}"
}

# output -ldflags parameters
chaosd::version::ldflags
