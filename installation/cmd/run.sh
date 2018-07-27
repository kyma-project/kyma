#!/usr/bin/env bash

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

DOMAIN="kyma.local"

VM_DRIVER="virtualbox"
if [ `uname -s` = "Darwin" ]; then
    VM_DRIVER="hyperkit"
fi

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
            CR_PATH="$2"
            shift # past argument
            shift # past value
        ;;
        --vm-driver)
            VM_DRIVER="$2"
            shift
            shift
        ;;
        *)    # unknown option
            POSITIONAL+=("$1") # save it in an array for later
            shift # past argument
        ;;
    esac
done
set -- "${POSITIONAL[@]}" # restore positional parameters

if [[ ! ${SKIP_MINIKUBE_START} ]]; then
    bash ${CURRENT_DIR}/../scripts/minikube.sh --domain "${DOMAIN}" --vm-driver "${VM_DRIVER}"
fi

bash $CURRENT_DIR/../scripts/generate-local-config.sh

if [ -z "$CR_PATH" ]; then

    TMPDIR=`mktemp -d "${CURRENT_DIR}/../../temp-XXXXXXXXXX"`
    CR_PATH="${TMPDIR}/installer-cr-local.yaml"

    bash ${CURRENT_DIR}/../scripts/create-cr.sh --output "${CR_PATH}" --domain "${DOMAIN}"
    bash ${CURRENT_DIR}/../scripts/installer.sh --local --cr "${CR_PATH}"

    rm -rf $TMPDIR
else
    bash ${CURRENT_DIR}/../scripts/installer.sh --cr "${CR_PATH}"
fi
