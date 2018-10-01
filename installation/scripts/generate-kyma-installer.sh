#!/usr/bin/env bash

set -o errexit

usage () {
    echo 'Provide correct input arguments'
    echo "First argument: path to versions.env file"
    echo "Second argument: path to installer-config file, cluster or local"
    exit 1
}

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
KYMA_PATH="${CURRENT_DIR}/../.."
INSTALLER_YAML_PATH="${KYMA_PATH}/installation/resources/installer.yaml"

if [[ ! $# -eq 2 ]] ; then
    usage
fi

VERSIONS_ENV_PATH="$1"
INSTALLER_CONFIG_PATH="$2"

if [ ! -f ${VERSIONS_ENV_PATH} ]; then
    echo "${VERSIONS_ENV_PATH} not found"
    usage
fi

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

TMPDIR=`mktemp -d "${KYMA_PATH}/temp-XXXXXXXXXX"`
CR_PATH="${TMPDIR}/installer-cr-local.yaml"
bash ${KYMA_PATH}/installation/scripts/create-cr.sh --url "" --output "${CR_PATH}" --version 0.0.1
cat ${CR_PATH}
rm -rf ${TMPDIR}

echo "---"

#TODO should be fixed with https://github.com/kyma-project/kyma/issues/959
kubectl create cm versions --from-env-file "${VERSIONS_ENV_PATH}" -n kyma-installer --dry-run -o yaml
echo "  labels:"
echo "    installer: overrides"
echo "---"
