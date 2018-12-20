#!/bin/bash

set -o errexit

ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

readonly FILE_NAME="components.env"
readonly FILE_PATH=${ROOT_PATH}/../${FILE_NAME}

# Do nothing if the components.env is empty or does not exist at all
if [[ ! -f "${FILE_PATH}" || ! -s "${FILE_PATH}" ]]; then
    exit 0
fi

kubectl create configmap "kyma-sub-components" --from-env-file=${FILE_PATH} -n kyma-installer
kubectl label configmap "kyma-sub-components" -n kyma-installer "installer=overrides"

rm -rf "$FILE_PATH"