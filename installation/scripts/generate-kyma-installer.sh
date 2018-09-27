#!/usr/bin/env bash

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
KYMA_PATH="${CURRENT_DIR}/../.."
INSTALLER_YAML_PATH="${KYMA_PATH}/installation/resources/installer.yaml"
UI_TEST_SECRET_PATH="${KYMA_PATH}/installation/resources/ui-test-secret.yaml.tpl"

VERSIONS_ENV_PATH=""
INSTALLER_CONFIG=""
OUTPUT_PATH=""

if [[ ! $# -eq 2 ]] ; then
    echo 'Provide correct input arguments'
    exit 1
fi

case "$1" in
    "local")   
        INSTALLER_CONFIG="${KYMA_PATH}/installation/resources/installer-config-local.yaml.tpl"
        OUTPUT_PATH="${KYMA_PATH}/kyma-installer/local-kyma-installer.yaml" ;;
    "cluster") 
        INSTALLER_CONFIG="${KYMA_PATH}/installation/resources/installer-config-cluster.yaml.tpl"
        OUTPUT_PATH="${KYMA_PATH}/kyma-installer/cluster-kyma-installer.yaml" ;;    
    *) 
        echo 'Provide which installer you want to generate by passing as argument "local" or "cluster"'
        exit 1 ;;
esac

case "$2" in
    */versions.env)
        VERSIONS_ENV_PATH="$2" ;;
    *)
        echo "Provide correct versions.env path"
        exit 1 ;;
esac

if [ ! -f ${VERSIONS_ENV_PATH} ]; then
    echo "${VERSIONS_ENV_PATH} not found"
    exit 1
fi

if [ ! -f ${INSTALLER_YAML_PATH} ]; then
    echo "${INSTALLER_YAML_PATH} not found"
    exit 1
fi

TMPDIR=`mktemp -d "${KYMA_PATH}/temp-XXXXXXXXXX"`
CR_PATH="${TMPDIR}/installer-cr-local.yaml"
bash ${KYMA_PATH}/installation/scripts/create-cr.sh --url "" --output "${CR_PATH}" --version 0.0.1

#TODO should be fixed with https://github.com/kyma-project/kyma/issues/959

echo "---" > ${OUTPUT_PATH}

cat ${INSTALLER_YAML_PATH} >> ${OUTPUT_PATH}

echo "---" >> ${OUTPUT_PATH}

cat ${INSTALLER_CONFIG} >> ${OUTPUT_PATH}

echo "---" >> ${OUTPUT_PATH}

cat ${CR_PATH} >> ${OUTPUT_PATH}

echo "---" >> ${OUTPUT_PATH}

if [ "$1" == "cluster" ]; then
    cat ${UI_TEST_SECRET_PATH} >> ${OUTPUT_PATH}
    echo -e "\n---" >> ${OUTPUT_PATH}
fi

kubectl create cm versions --from-env-file ${VERSIONS_ENV_PATH} -n kyma-installer --dry-run -o yaml >> ${OUTPUT_PATH}
echo "  labels:" >> ${OUTPUT_PATH}
echo "    installer: overrides" >> ${OUTPUT_PATH}
echo "---" >> ${OUTPUT_PATH}

rm -rf ${TMPDIR}