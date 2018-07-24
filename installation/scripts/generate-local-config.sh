#!/usr/bin/env bash
set -e

ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

# The following variables are optional and must be exported manually
# before running local installation (you need them only to enable Azure Broker):
# AZURE_BROKER_TENANT_ID,
# AZURE_BROKER_SUBSCRIPTION_ID,
# AZURE_BROKER_CLIENT_ID,
# AZURE_BROKER_CLIENT_SECRET

MINIKUBE_IP=$(minikube ip)
MINIKUBE_CA=$(cat ${HOME}/.minikube/ca.crt | base64 | tr -d '\n')

UI_TEST_USER=admin@kyma.cx
UI_TEST_PASSWORD=nimda123

##########

echo -e "\nGenerating secret for UI Test"
bash ${ROOT_PATH}/create-generic-secret.sh "${ROOT_PATH}/../resources/ui-test-secret.yaml" \
  "UI_TEST_USER" "${UI_TEST_USER}" \
  "UI_TEST_PASSWORD" "${UI_TEST_PASSWORD}"

echo -e "\nApplying asecret for UI Test"
kubectl create -f "${ROOT_PATH}/../resources/ui-test-secret.yaml"

##########

if [ -n "${AZURE_BROKER_SUBSCRIPTION_ID}" ]; then
  echo -e "\nGenerating secret for Azure Broker"
  bash ${ROOT_PATH}/create-generic-secret.sh "${ROOT_PATH}/../resources/azure-broker-secret.yaml" \
    "AZURE_BROKER_SUBSCRIPTION_ID" "${AZURE_BROKER_SUBSCRIPTION_ID}" \
    "AZURE_BROKER_TENANT_ID" "${AZURE_BROKER_TENANT_ID}" \
    "AZURE_BROKER_CLIENT_ID" "${AZURE_BROKER_CLIENT_ID}" \
    "AZURE_BROKER_CLIENT_SECRET" "${AZURE_BROKER_CLIENT_SECRET}"

  echo -e "\nApplying asecret for Azure Broker"
  kubectl create -f "${ROOT_PATH}/../resources/azure-broker-secret.yaml"
fi

##########

echo -e "\nGenerating secret for Cluster certificate"

TLS_FILE=$ROOT_PATH/../resources/local-tls-certs.yaml
TLS_CERT=$(cat $TLS_FILE | grep 'tls.crt' | sed 's/^.*: //')
TLS_KEY=$(cat $TLS_FILE | grep 'tls.key' | sed 's/^.*: //')

bash ${ROOT_PATH}/create-generic-secret.sh "${ROOT_PATH}/../resources/cluster-certificate-secret.yaml" \
  "TLS_CERT" "${TLS_CERT}" \
  "TLS_KEY" "${TLS_KEY}"

echo -e "\nApplying asecret for Cluster certificate"
kubectl create -f "${ROOT_PATH}/../resources/cluster-certificate-secret.yaml"

##########

echo -e "\nGenerating config map for installation"
OUTPUT=$(mktemp)

bash ${ROOT_PATH}/create-config-map.sh \
--ip-address "" \
--domain "kyma.local" \
--remote-env-ip "" \
--k8s-apiserver-url "${MINIKUBE_IP}" \
--k8s-apiserver-ca "${MINIKUBE_CA}" \
--admin-group "" \
--output ${OUTPUT}

kubectl create -f ${OUTPUT}

rm ${OUTPUT}
