#!/usr/bin/env bash

readonly OPERATOR_DIR=$1
readonly CONTROLLER_GEN=$2
readonly CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

TMP_DIR=$(mktemp -d)

source "${CURRENT_DIR}/utilities.sh" || { echo 'Cannot load CI utilities.'; exit 1; }

$CONTROLLER_GEN rbac:roleName=manager-role crd webhook paths="${OPERATOR_DIR}/..." output:crd:artifacts:config="${TMP_DIR}"

DIFF=$(diff -q $TMP_DIR ${OPERATOR_DIR}/config/crd/bases)
if [ -n "${DIFF}" ]; then
    echo -e "${RED}x manifests linting failed${NC}"
    echo -e "Generated manifests in config/crd/bases are not on par with API types in apis/telemetry/v1alpha1"
    exit 1
fi