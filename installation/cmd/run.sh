#!/usr/bin/env bash

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SCRIPTS_DIR="${CURRENT_DIR}/../scripts"
DOMAIN="kyma.local"

VM_DRIVER="virtualbox"
if [ `uname -s` = "Darwin" ]; then
    VM_DRIVER="hyperkit"
fi

source $SCRIPTS_DIR/utils.sh

POSITIONAL=()
while [[ $# -gt 0 ]]
do
    
    key="$1"

    case ${key} in
        --skip-minikube-start)
            SKIP_MINIKUBE_START=true
            shift # past argument
        ;;
        --cr)
            checkInputParameterValue "$2"
            CR_PATH="$2"
            shift # past argument
            shift # past value
        ;;
        --vm-driver)
            checkInputParameterValue "$2"
            VM_DRIVER="$2"
            shift
            shift
        ;;
        --knative)
            KNATIVE="--knative"
            shift
        ;;
        --*)
            echo "Unknown flag ${1}"
            exit 1
        ;;
        *)    # unknown option
            POSITIONAL+=("$1") # save it in an array for later
            shift # past argument
        ;;
    esac
done
set -- "${POSITIONAL[@]}" # restore positional parameters

MINIKUBE_EXTRA_ARGS=""

if [ -n "$KNATIVE" ]; then

    MINIKUBE_EXTRA_ARGS="${MINIKUBE_EXTRA_ARGS} --memory 10240 --disk-size 30g"

fi

if [[ ! ${SKIP_MINIKUBE_START} ]]; then
    bash ${SCRIPTS_DIR}/minikube.sh --domain "${DOMAIN}" --vm-driver "${VM_DRIVER}" ${MINIKUBE_EXTRA_ARGS}
fi

bash ${SCRIPTS_DIR}/build-kyma-installer.sh --vm-driver "${VM_DRIVER}"

if [ -z "$CR_PATH" ]; then

    TMPDIR=`mktemp -d "${CURRENT_DIR}/../../temp-XXXXXXXXXX"`
    CR_PATH="${TMPDIR}/installer-cr-local.yaml"
    bash ${SCRIPTS_DIR}/create-cr.sh --output "${CR_PATH}" --domain "${DOMAIN}"

fi

bash ${SCRIPTS_DIR}/installer.sh --local --cr "${CR_PATH}" "${KNATIVE}"
rm -rf $TMPDIR
