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

bash ${CURRENT_DIR}/build.sh ${KYMA_PATH}
bash ${KYMA_PATH}/installation/scripts/generate-local-config.sh
bash ${KYMA_PATH}/installation/scripts/installer.sh ${CR_PATH} ${LOCAL_KYMA}
