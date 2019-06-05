#!/usr/bin/env bash

set -o errexit

echo "The configure-components.sh script is deprecated and will be removed. Use Kyma CLI instead."

ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

readonly FILE_NAME="components.env"
readonly FILE_PATH=${ROOT_PATH}/../${FILE_NAME}
readonly CM_NAME="kyma-sub-components"
readonly LABEL="installer=overrides"

# Do nothing if the components.env is empty or does not exist at all
if [[ ! -f "${FILE_PATH}" || ! -s "${FILE_PATH}" ]]; then
    exit 0
fi

kubectl create configmap ${CM_NAME} -n kyma-installer --from-env-file="${FILE_PATH}" --dry-run -o yaml | kubectl apply -f -
kubectl label configmap ${CM_NAME} -n kyma-installer ${LABEL}

rm -rf "${FILE_PATH}"
