#!/usr/bin/env bash
set -e

ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

CONFIG_TPL_PATH="${ROOT_PATH}/../resources/installer-config.yaml.tpl"
CONFIG_OUTPUT_PATH=$(mktemp)

cp $CONFIG_TPL_PATH $CONFIG_OUTPUT_PATH

##########

echo -e "\nGenerating secret for Cluster certificate"

TLS_FILE=${ROOT_PATH}/../resources/local-tls-certs.yaml
TLS_CERT=$(cat ${TLS_FILE} | grep 'tls.crt' | sed 's/^.*: //' | base64 | tr -d '\n')
TLS_KEY=$(cat ${TLS_FILE} | grep 'tls.key' | sed 's/^.*: //' | base64 | tr -d '\n')

bash ${ROOT_PATH}/replace-placeholder.sh --path ${CONFIG_OUTPUT_PATH} --placeholder "__TLS_CERT__" --value "${TLS_CERT}"
bash ${ROOT_PATH}/replace-placeholder.sh --path ${CONFIG_OUTPUT_PATH} --placeholder "__TLS_KEY__" --value "${TLS_KEY}"

##########

echo -e "\nGenerating secret for Remote Environemnts"

bash ${ROOT_PATH}/replace-placeholder.sh --path ${CONFIG_OUTPUT_PATH} --placeholder "__REMOTE_ENV_CA__" --value ""
bash ${ROOT_PATH}/replace-placeholder.sh --path ${CONFIG_OUTPUT_PATH} --placeholder "__REMOTE_ENV_CA_KEY__" --value ""

##########

echo -e "\nGenerating config map for installation"

MINIKUBE_IP=$(minikube ip)
MINIKUBE_CA=$(cat ${HOME}/.minikube/ca.crt | base64 | tr -d '\n')

bash ${ROOT_PATH}/replace-placeholder.sh --path ${CONFIG_OUTPUT_PATH} --placeholder "__DOMAIN__" --value "kyma.local"
bash ${ROOT_PATH}/replace-placeholder.sh --path ${CONFIG_OUTPUT_PATH} --placeholder "__EXTERNAL_IP_ADDRESS__" --value ""
bash ${ROOT_PATH}/replace-placeholder.sh --path ${CONFIG_OUTPUT_PATH} --placeholder "__REMOTE_ENV_IP__" --value ""
bash ${ROOT_PATH}/replace-placeholder.sh --path ${CONFIG_OUTPUT_PATH} --placeholder "__K8S_APISERVER_URL__" --value "${MINIKUBE_IP}"
bash ${ROOT_PATH}/replace-placeholder.sh --path ${CONFIG_OUTPUT_PATH} --placeholder "__K8S_APISERVER_CA__" --value "${MINIKUBE_CA}"
bash ${ROOT_PATH}/replace-placeholder.sh --path ${CONFIG_OUTPUT_PATH} --placeholder "__ADMIN_GROUP__" --value ""
bash ${ROOT_PATH}/replace-placeholder.sh --path ${CONFIG_OUTPUT_PATH} --placeholder "__ENABLE_ETCD_BACKUP_OPERATOR__" --value "false"
bash ${ROOT_PATH}/replace-placeholder.sh --path ${CONFIG_OUTPUT_PATH} --placeholder "__ETCD_BACKUP_ABS_CONTAINER_NAME__" --value ""

##########

echo -e "\nApplying configuration"

kubectl create namespace "kyma-installer"

kubectl apply -f ${CONFIG_OUTPUT_PATH}

rm ${CONFIG_OUTPUT_PATH}

##########

echo -e "\nGenerating secret for UI Test"

UI_TEST_TPL_PATH="${ROOT_PATH}/../resources/ui-test-secret.yaml.tpl"
UI_TEST_OUTPUT_PATH=$(mktemp)

cp $UI_TEST_TPL_PATH $UI_TEST_OUTPUT_PATH

UI_TEST_USER=$(echo -n "admin@kyma.cx" | base64 | tr -d '\n')
UI_TEST_PASSWORD=$(echo -n "nimda123" | base64 | tr -d '\n')

bash ${ROOT_PATH}/replace-placeholder.sh --path ${UI_TEST_OUTPUT_PATH} --placeholder "__UI_TEST_USER__" --value "${UI_TEST_USER}"
bash ${ROOT_PATH}/replace-placeholder.sh --path ${UI_TEST_OUTPUT_PATH} --placeholder "__UI_TEST_PASSWORD__" --value "${UI_TEST_PASSWORD}"

echo -e "\nApplying asecret for UI Test"
kubectl apply -f "${UI_TEST_OUTPUT_PATH}"

rm ${UI_TEST_OUTPUT_PATH}

##########

# The following variables are optional and must be exported manually
# before running local installation (you need them only to enable Azure Broker):
# AZURE_BROKER_TENANT_ID,
# AZURE_BROKER_SUBSCRIPTION_ID,
# AZURE_BROKER_CLIENT_ID,
# AZURE_BROKER_CLIENT_SECRET

if [ -n "${AZURE_BROKER_SUBSCRIPTION_ID}" ]; then
  echo -e "\nGenerating secret for Azure Broker"

  AZURE_BROKER_TPL_PATH="${ROOT_PATH}/../resources/azure-broker-secret.yaml.tpl"
  AZURE_BROKER_OUTPUT_PATH=$(mktemp)
  cp $AZURE_BROKER_TPL_PATH $AZURE_BROKER_OUTPUT_PATH

  bash ${ROOT_PATH}/replace-placeholder.sh --path ${AZURE_BROKER_OUTPUT_PATH} --placeholder "__AZURE_BROKER_SUBSCRIPTION_ID__" --value "${AZURE_BROKER_SUBSCRIPTION_ID}"
  bash ${ROOT_PATH}/replace-placeholder.sh --path ${AZURE_BROKER_OUTPUT_PATH} --placeholder "__AZURE_BROKER_TENANT_ID__" --value "${AZURE_BROKER_TENANT_ID}"
  bash ${ROOT_PATH}/replace-placeholder.sh --path ${AZURE_BROKER_OUTPUT_PATH} --placeholder "__AZURE_BROKER_CLIENT_ID__" --value "${AZURE_BROKER_CLIENT_ID}"
  bash ${ROOT_PATH}/replace-placeholder.sh --path ${AZURE_BROKER_OUTPUT_PATH} --placeholder "__AZURE_BROKER_CLIENT_SECRET__" --value "${AZURE_BROKER_CLIENT_SECRET}"

  echo -e "\nApplying asecret for Azure Broker"
  kubectl apply -f "${AZURE_BROKER_OUTPUT_PATH}"
  
  rm ${AZURE_BROKER_OUTPUT_PATH}
fi
