#!/usr/bin/env bash

readonly PROW_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source "${PROW_DIR}/../../test-infra/scripts/presubmit.sh"

readonly SUBDIRECTORY="${PROW_DIR}/../$(echo "${JOB_NAME}" | sed 's/prow\///')"

function resolve_tests() {
    make resolve
}

function validate_tests() {
    make validate
}

function build_tests() {
    make build
}

function unit_tests() {
    make unit-tests
}

function integration_tests() {
    make integration-tests
}

function build_image_tests() {
    make build-image
}

function push_image_tests() {
    make push-image
}

main $@
