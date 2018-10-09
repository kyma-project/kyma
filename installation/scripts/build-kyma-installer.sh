#!/usr/bin/env bash

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
IMAGE_NAME="$(${CURRENT_DIR}/extract-kyma-installer-image.sh)"

POSITIONAL=()
while [[ $# -gt 0 ]]
do
    key="$1"

    case ${key} in
        --vm-driver)
            VM_DRIVER="$2"
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

# pushd kyma

if [[ "$VM_DRIVER" != "none" ]]; then
    eval $(minikube docker-env --shell bash)
fi

docker build -t ${IMAGE_NAME} -f ${CURRENT_DIR}/../../kyma-installer/kyma.Dockerfile .

# popd