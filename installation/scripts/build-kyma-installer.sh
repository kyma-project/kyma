#!/usr/bin/env bash

set -o errexit

echo "The script build-kyma-installer.sh is deprecated and will be removed with Kyma release 1.4, please use Kyma CLI instead"

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT_DIR=${CURRENT_DIR}/../../
IMAGE_NAME="$(${CURRENT_DIR}/extract-kyma-installer-image.sh)"
BUILD_ARG=""

echo "
################################################################################
# Kyma-Installer build
################################################################################
"

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
        --installer-version)
            BUILD_ARG="--build-arg INSTALLER_VERSION=$2"
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

pushd $ROOT_DIR

if [[ "$VM_DRIVER" != "none" ]]; then
    eval $(minikube docker-env --shell bash)
fi

docker build -t ${IMAGE_NAME} ${BUILD_ARG} -f ./tools/kyma-installer/kyma.Dockerfile .

popd 
