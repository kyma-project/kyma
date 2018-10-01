#!/bin/bash

set -o errexit

delete_if_exists () {
    EXISTS="$(az storage blob exists -c ${CONTAINER} -n $1 --connection-string ${CONNECTION_STRING} | jq .exists)"
    if [ "${EXISTS}" = "true" ]; then
        echo "BLOB $1 already exists, deleting..."
        az storage blob delete -c ${CONTAINER} -n $1 --connection-string ${CONNECTION_STRING}
    fi
}

KYMA_VERSIONS_SOURCE_FILE="/installation/versions-overrides.env"

CONTAINER="kyma-versions"
STORAGE_ACCOUNT_NAME="kymainstaller"
RESOURCE_GROUP="kyma-installer-rg"

az login --service-principal -u ${AZBR_CLIENT_ID} -p ${AZBR_CLIENT_SECRET} --tenant ${AZBR_TENANT_ID} > /dev/null
az account set --subscription ${AZBR_SUBSCRIPTION_ID} > /dev/null
CONNECTION_STRING=$(az storage account show-connection-string --name ${STORAGE_ACCOUNT_NAME} -g ${RESOURCE_GROUP} | jq .connectionString)

delete_if_exists "${KYMA_VERSIONS_FILE_NAME}"
az storage blob upload -f ${KYMA_VERSIONS_SOURCE_FILE} -c ${CONTAINER} -n ${KYMA_VERSIONS_FILE_NAME} --connection-string ${CONNECTION_STRING}
