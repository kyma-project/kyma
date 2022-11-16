#!/usr/bin/env bash

readonly OPERATOR_DIR=$1
readonly CONTROLLER_GEN=$2

TMP_DIR=$(mktemp -d)

$CONTROLLER_GEN rbac:roleName=manager-role crd webhook paths="${OPERATOR_DIR}/..." output:crd:artifacts:config="${TMP_DIR}"

DIFF=$(diff -q $TMP_DIR ${OPERATOR_DIR}/config/crd/bases)
if [ -n "${DIFF}" ]; then
    echo -e "Some error message..."
    exit 1
fi