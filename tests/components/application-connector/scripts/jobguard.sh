#!/bin/bash
export GO111MODULE=on

ROOT_PATH=$(dirname "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)")

KYMA_PROJECT_DIR=${KYMA_PROJECT_DIR:-"/home/prow/go/src/github.com/kyma-project"}
JOB_NAME_PATTERN=${JOB_NAME_PATTERN:-"(pre-main-kyma-components-.*)|(pre-main-kyma-tests-.*)|(pre-kyma-components-.*)|(pre-kyma-tests-.*)|(pull-.*-build)"}
TIMEOUT=${JOBGUARD_TIMEOUT:-"15m"}

export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"

if [ -z "$PULL_PULL_SHA" ]; then
  echo "WORKAROUND: skip jobguard execution - not on PR commit"
  exit 0
fi

args=(
  "-github-endpoint=http://ghproxy"
  "-github-endpoint=https://api.github.com"
  "-github-token-path=/etc/github/token"
  "-fail-on-no-contexts=false"
  "-timeout=$TIMEOUT"
  "-org=$REPO_OWNER"
  "-repo=$REPO_NAME"
  "-base-ref=$PULL_PULL_SHA"
  "-expected-contexts-regexp=$JOB_NAME_PATTERN"
)

if [ -x "/prow-tools/jobguard" ]; then
  /prow-tools/jobguard "${args[@]}"
else
  cd "${ROOT_PATH}/cmd/jobguard" || exit 1
  go run main.go "${args[@]}"
fi
