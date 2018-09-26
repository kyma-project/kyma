#!/usr/bin/env bash

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
KYMA_PATH="${CURRENT_DIR}/../.."
VERSIONS_ENV_PATH="${KYMA_PATH}/installation/versions.env"
INSTALLER_YAML_PATH="${KYMA_PATH}/installation/resources/installer.yaml"
LOCAL_KYMA_INSTALLER_YAML_PATH="${KYMA_PATH}/kyma-installer/local-kyma-installer.yaml"

if [ ! -f ${VERSIONS_ENV_PATH} ]; then
    echo "${VERSIONS_ENV_PATH} not found"
    exit 1
fi

if [ ! -f ${INSTALLER_YAML_PATH} ]; then
    echo "${INSTALLER_YAML_PATH} not found"
    exit 1
fi

#TODO should be fixed with https://github.com/kyma-project/kyma/issues/959

echo "---" > ${LOCAL_KYMA_INSTALLER_YAML_PATH}
kubectl create cm versions --from-env-file ${VERSIONS_ENV_PATH} -n kyma-installer --dry-run -o yaml >> ${LOCAL_KYMA_INSTALLER_YAML_PATH}
echo "  labels:" >> ${LOCAL_KYMA_INSTALLER_YAML_PATH}
echo "    installer: overrides" >> ${LOCAL_KYMA_INSTALLER_YAML_PATH}
echo "---" >> ${LOCAL_KYMA_INSTALLER_YAML_PATH}

cat ${INSTALLER_YAML_PATH} >> ${LOCAL_KYMA_INSTALLER_YAML_PATH}