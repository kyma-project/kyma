#!/usr/bin/env bash

# This script run tests on local k3d

set -e

KYMA_LOCATION=${KYMA_LOCATION:-$(readlink -f ../..)}
OVERRIDES=""
ARG_TEST_NAME=""

for arg in "$@"; do
  shift
  case "$arg" in
    "--test")
      ARG_TEST_NAME="$1"
    ;;
  esac
done
if [ -n "${ARG_TEST_NAME}" ]; then
    OVERRIDES+="--set testSuite=${ARG_TEST_NAME} "
fi
echo "Running test suite: ${ARG_TEST_NAME:-"default"}"

TEST_IMAGE_REGISTRY="eu.gcr.io/kyma-project"
TEST_IMAGE_NAME="function-controller-test"
TEST_IMAGE_TAG="local"
TEST_IMAGE_FULL_NAME=${TEST_IMAGE_REGISTRY}/${TEST_IMAGE_NAME}:$TEST_IMAGE_TAG

cd "${KYMA_LOCATION}"/tests/function-controller
docker build -t ${TEST_IMAGE_FULL_NAME}  -f Dockerfile .
k3d image import ${TEST_IMAGE_FULL_NAME} --cluster kyma
cd -

OVERRIDES+="--set global.testImages.function_controller_test.name=${TEST_IMAGE_NAME} "
OVERRIDES+="--set global.testImages.function_controller_test.version=${TEST_IMAGE_TAG} "

helm uninstall serverless-test | true
(cd ${KYMA_LOCATION}/resources/serverless &&  helm install serverless-test ./charts/k3s-tests/ -f ./values.yaml ${OVERRIDES})

echo ----------------------------
echo Tests are running on k3d.
echo Now look for the pod starting with "k3s-serverless-test-" and look at its logs.
