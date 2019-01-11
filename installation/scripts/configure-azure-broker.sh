#!/usr/bin/env bash
set -e

ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

# The following variables are optional and must be exported manually in base64 encoded format
# before running local installation (you need them only to enable Azure Broker):
# AZURE_BROKER_TENANT_ID,
# AZURE_BROKER_SUBSCRIPTION_ID,
# AZURE_BROKER_CLIENT_ID,
# AZURE_BROKER_CLIENT_SECRET.

if [ "$#" -ne 1 ]; then
    echo "configure-azure-broker.sh error: this script accepts one argument - path to the Azure-Broker secret output file"
fi

for var in AZURE_BROKER_TENANT_ID AZURE_BROKER_CLIENT_ID AZURE_BROKER_CLIENT_SECRET; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        exit 1
    fi
done

AZURE_BROKER_TPL_PATH="${ROOT_PATH}/../resources/azure-broker-secret.yaml.tpl"
AZURE_BROKER_OUTPUT_PATH=$1
cp $AZURE_BROKER_TPL_PATH $AZURE_BROKER_OUTPUT_PATH

bash ${ROOT_PATH}/replace-placeholder.sh --path ${AZURE_BROKER_OUTPUT_PATH} --placeholder "__AZURE_BROKER_SUBSCRIPTION_ID__" --value "${AZURE_BROKER_SUBSCRIPTION_ID}"
bash ${ROOT_PATH}/replace-placeholder.sh --path ${AZURE_BROKER_OUTPUT_PATH} --placeholder "__AZURE_BROKER_TENANT_ID__" --value "${AZURE_BROKER_TENANT_ID}"
bash ${ROOT_PATH}/replace-placeholder.sh --path ${AZURE_BROKER_OUTPUT_PATH} --placeholder "__AZURE_BROKER_CLIENT_ID__" --value "${AZURE_BROKER_CLIENT_ID}"
bash ${ROOT_PATH}/replace-placeholder.sh --path ${AZURE_BROKER_OUTPUT_PATH} --placeholder "__AZURE_BROKER_CLIENT_SECRET__" --value "${AZURE_BROKER_CLIENT_SECRET}"
