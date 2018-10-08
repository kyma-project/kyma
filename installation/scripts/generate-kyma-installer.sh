#!/usr/bin/env bash

set -o errexit

usage () {
    echo "Provide correct input argument"
    echo "First argument: path to installer-config file, cluster or local"
    exit 1
}

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
KYMA_PATH="${CURRENT_DIR}/../.."
INSTALLER_YAML_PATH="${KYMA_PATH}/installation/resources/installer.yaml"

if [[ ! $# -eq 1 ]] ; then
    usage
fi

INSTALLER_CONFIG_PATH="$1"

if [ ! -f ${INSTALLER_CONFIG_PATH} ]; then
    echo "${INSTALLER_CONFIG_PATH} not found"
    usage
fi

if [ ! -f ${INSTALLER_YAML_PATH} ]; then
    echo "${INSTALLER_YAML_PATH} not found"
    usage
fi

cat ${INSTALLER_YAML_PATH}

echo "---"

cat ${INSTALLER_CONFIG_PATH}

echo "---"

TMP_CR=$(mktemp)
bash ${KYMA_PATH}/installation/scripts/create-cr.sh --url "" --output "${TMP_CR}" --version 0.0.1
cat ${TMP_CR}
rm -rf ${TMP_CR}

echo "---"
