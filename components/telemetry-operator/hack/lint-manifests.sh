#!/usr/bin/env bash

readonly OPERATOR_DIR=$1
readonly CONTROLLER_GEN=$2
readonly CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
readonly ROOT_PATH="$( cd "${CURRENT_DIR}/../../../" && pwd )"

TMP_DIR=$(mktemp -d)

source "${ROOT_PATH}/hack/utilities.sh" || { echo 'Cannot load CI utilities.'; exit 1; }

cleanup() {
    rm -rf "${TMP_DIR}" || true
}

trap cleanup EXIT SIGINT

echo $ROOT_PATH

$CONTROLLER_GEN rbac:roleName=manager-role crd webhook paths="${OPERATOR_DIR}/..." output:crd:artifacts:config="${TMP_DIR}"

DIFF=$(diff -q $TMP_DIR ${OPERATOR_DIR}/config/crd/bases)
if [ -n "${DIFF}" ]; then
    echo -e "${RED}x manifests linting failed${NC}"
    echo -e "Generated manifests in config/crd/bases are not on par with API types in apis/telemetry/v1alpha1"
    echo -e "Please run 'make manifests-local'"
    exit 1
fi
