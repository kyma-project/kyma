#!/usr/bin/env bash

set -o errexit

usage () {
    echo 'Provide correct input arguments'
    echo 'First argument: "local" or "cluster" - decide which installer you want to generate'
    echo "Second argument: path to versions.env file"
    echo "Third argument: path to installer-config file, cluster or local"
    exit 1
}

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
KYMA_PATH="${CURRENT_DIR}/../.."
INSTALLER_YAML_PATH="${KYMA_PATH}/installation/resources/installer.yaml"

if [[ ! $# -eq 3 ]] ; then
    usage
fi

VERSIONS_ENV_PATH=""
OUTPUT_PATH=""
INSTALLER_CONFIG="$3"

case "$1" in
    "local")   
        OUTPUT_PATH="${KYMA_PATH}/kyma-installer/local-kyma-installer.yaml" ;;
    "cluster") 
        OUTPUT_PATH="${KYMA_PATH}/kyma-installer/cluster-kyma-installer.yaml" ;;    
    *)  usage ;;
esac

case "$2" in
    */versions.env)
        VERSIONS_ENV_PATH="$2" ;;
    *)  usage ;;
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

cat ${INSTALLER_YAML_PATH} >> ${OUTPUT_PATH}

echo "---" >> ${OUTPUT_PATH}

cat ${INSTALLER_CONFIG} >> ${OUTPUT_PATH}

echo "---" >> ${OUTPUT_PATH}

cat ${CR_PATH} >> ${OUTPUT_PATH}

echo "---" >> ${OUTPUT_PATH}

#TODO should be fixed with https://github.com/kyma-project/kyma/issues/959

kubectl create cm versions --from-env-file ${VERSIONS_ENV_PATH} -n kyma-installer --dry-run -o yaml >> ${OUTPUT_PATH}
echo "  labels:" >> ${OUTPUT_PATH}
echo "    installer: overrides" >> ${OUTPUT_PATH}
echo "---" >> ${OUTPUT_PATH}

rm -rf ${TMPDIR}