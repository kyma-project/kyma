#!/bin/bash

INSTALLER_NS_NAME=kyma-installer
INSTALLER_POD_PATTERN=kyma-installer
REMOTE_DIRECTORY=/kyma
CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
LOCAL_DIRECTORY="$( cd "$( dirname "${BASH_SOURCE[0]}" )/../.." && pwd )"

bash ${CURRENT_DIR}/is-ready.sh ${INSTALLER_NS_NAME} name ${INSTALLER_POD_PATTERN}

POD_NAME=$(kubectl -n ${INSTALLER_NS_NAME} get pods -l name=${INSTALLER_POD_PATTERN} -o jsonpath="{.items[*].metadata.name}")

echo "Copying kyma sources from ${LOCAL_DIRECTORY} into ${POD_NAME}:${REMOTE_DIRECTORY} ..."

kubectl -n ${INSTALLER_NS_NAME} exec ${POD_NAME} -- "/bin/rm" "-rf" "${REMOTE_DIRECTORY}"
kubectl -n ${INSTALLER_NS_NAME} exec ${POD_NAME} -- "/bin/mkdir" "${REMOTE_DIRECTORY}"

kubectl cp "${LOCAL_DIRECTORY}/resources" "${INSTALLER_NS_NAME}/${POD_NAME}:${REMOTE_DIRECTORY}/resources"
kubectl cp "${LOCAL_DIRECTORY}/installation" "${INSTALLER_NS_NAME}/${POD_NAME}:${REMOTE_DIRECTORY}/installation"
kubectl cp "${LOCAL_DIRECTORY}/docs" "${INSTALLER_NS_NAME}/${POD_NAME}:${REMOTE_DIRECTORY}/docs"
