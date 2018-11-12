#!/usr/bin/env bash

set -o errexit

readonly ROOT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
readonly FINAL_IMAGE="gcp-broker-provider"
readonly KUBECTL_CLI_VERSION=1.10.0
readonly SC_CLI_VERSION=1.0.0-beta.5

pushd ${ROOT_DIR} > /dev/null

# Exit handler. This function is called anytime an EXIT signal is received.
# This function should never be explicitly called.
function _trap_exit () {
    popd > /dev/null
}
trap _trap_exit EXIT

function buildDockerImage() {
    docker build --no-cache -t ${FINAL_IMAGE} \
        --build-arg KUBECTL_CLI_VERSION=${KUBECTL_CLI_VERSION} \
        --build-arg SC_CLI_VERSION=${SC_CLI_VERSION} .
}

# Only for testing purpose - need to be deleted
function pushImage() {
    docker tag ${FINAL_IMAGE} mszostok/${FINAL_IMAGE}:0.0.4
    docker push mszostok/${FINAL_IMAGE}:0.0.4
}

buildDockerImage
pushImage
