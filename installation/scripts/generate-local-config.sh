#!/usr/bin/env bash
set -e

ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

# The following variables are optional and must be exported manually in base64 encoded format
# before running local installation (you need them only to enable Azure Broker):
# AZURE_BROKER_TENANT_ID,
# AZURE_BROKER_SUBSCRIPTION_ID,
# AZURE_BROKER_CLIENT_ID,
# AZURE_BROKER_CLIENT_SECRET

if [ -n "${AZURE_BROKER_SUBSCRIPTION_ID}" ]; then

  echo -e "\nAzure-Broker subscription ID found in environment variables. Enabling component..."

  bash ${ROOT_PATH}/manage-component.sh "azure-broker" true

  ##########

  echo -e "\nGenerating the secret for Azure Broker..."

  AZURE_BROKER_TPL_PATH="${ROOT_PATH}/../resources/azure-broker-secret.yaml.tpl"
  AZURE_BROKER_OUTPUT_PATH=$(mktemp)
  cp $AZURE_BROKER_TPL_PATH $AZURE_BROKER_OUTPUT_PATH

  bash ${ROOT_PATH}/replace-placeholder.sh --path ${AZURE_BROKER_OUTPUT_PATH} --placeholder "__AZURE_BROKER_SUBSCRIPTION_ID__" --value "${AZURE_BROKER_SUBSCRIPTION_ID}"
  bash ${ROOT_PATH}/replace-placeholder.sh --path ${AZURE_BROKER_OUTPUT_PATH} --placeholder "__AZURE_BROKER_TENANT_ID__" --value "${AZURE_BROKER_TENANT_ID}"
  bash ${ROOT_PATH}/replace-placeholder.sh --path ${AZURE_BROKER_OUTPUT_PATH} --placeholder "__AZURE_BROKER_CLIENT_ID__" --value "${AZURE_BROKER_CLIENT_ID}"
  bash ${ROOT_PATH}/replace-placeholder.sh --path ${AZURE_BROKER_OUTPUT_PATH} --placeholder "__AZURE_BROKER_CLIENT_SECRET__" --value "${AZURE_BROKER_CLIENT_SECRET}"

  echo -e "\nApplying the secret for Azure Broker..."
  kubectl apply -f "${AZURE_BROKER_OUTPUT_PATH}"

  rm ${AZURE_BROKER_OUTPUT_PATH}
fi

##########

echo -e "\nConfiguring sub-components..."

bash ${ROOT_PATH}/configure-components.sh
