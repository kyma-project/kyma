#!/usr/bin/env bash

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
KYMA_PATH="${CURRENT_DIR}/../.."
VERSIONS_ENV_PATH="${KYMA_PATH}/installation/versions.env"
INSTALLER_YAML_PATH="${KYMA_PATH}/installation/resources/installer.yaml"

if [ ! -f ${VERSIONS_ENV_PATH} ]; then
    echo "${VERSIONS_ENV_PATH} not found"
    exit 1
fi

if [ ! -f ${INSTALLER_YAML_PATH} ]; then
    echo "${INSTALLER_YAML_PATH} not found"
    exit 1
fi

TMPFILE=$(mktemp ${KYMA_PATH}/installation/scripts/temp.XXXXXX)

echo "---" >> ${TMPFILE}
kubectl create cm versions --from-env-file ${VERSIONS_ENV_PATH} -n kyma-installer --dry-run -o yaml >> ${TMPFILE}
echo "  labels:" >> ${TMPFILE}
echo "    installer: overrides" >> ${TMPFILE}
echo "---" >> ${TMPFILE}

cat ${INSTALLER_YAML_PATH} >> ${TMPFILE}

cp -f ${TMPFILE} ${INSTALLER_YAML_PATH}

rm -f ${TMPFILE}