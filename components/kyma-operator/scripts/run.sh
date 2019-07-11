#!/usr/bin/env bash

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
LOCAL_KYMA="--local"
KYMA_PATH="${CURRENT_DIR}/../../.."
CR_PATH=""

POSITIONAL=()
while [[ $# -gt 0 ]]
do
    key="$1"

    case ${key} in
        --skip-minikube-start)
            SKIP_MINIKUBE_START=true
            shift # past argument
            ;;
        --local)
            LOCAL_KYMA="--local"
            shift
            ;;
        --cr)
            CR_PATH="--cr $2"
            shift # past argument
            shift # past value
            ;;
        *)    # unknown option
            POSITIONAL+=("$1") # save it in an array for later
            shift # past argument
            ;;
    esac
done
set -- "${POSITIONAL[@]}" # restore positional parameters

if [[ ! ${SKIP_MINIKUBE_START} ]]; then
    bash ${KYMA_PATH}/installation/scripts/minikube.sh --domain "kyma.local"
fi

bash ${CURRENT_DIR}/build.sh
bash ${KYMA_PATH}/installation/scripts/build-kyma-installer.sh --installer-version "dev"

if [[ -z ${CR_PATH} ]]; then
    TMPDIR=`mktemp -d "${KYMA_PATH}/temp-XXXXXXXXXX"`
    CR_PATH="${TMPDIR}/installer-cr-local.yaml"
    bash ${KYMA_PATH}/installation/scripts/create-cr.sh --url "" --output "${CR_PATH}" --version 0.0.1
    CR_PATH="--cr $CR_PATH"
fi

bash ${KYMA_PATH}/installation/scripts/installer.sh ${LOCAL_KYMA} ${CR_PATH}

if [ -f "${TMPDIR}/installer-cr-local.yaml" ]; then
    rm -rf $TMPDIR
fi 