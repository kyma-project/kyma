#!/usr/bin/env bash
set -e

ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

KYMA_PUBLIC_IP_NAME="KYMA-ingress-public-ip"
KYMA_REMOTE_ENV_IP_NAME="KYMA-nginx-ingress-public-ip"

FULL_DOMAIN=${DOMAIN_PREFIX}.${yBaseDomain}
VAULT_SECRET_NAME_PREFIX=$(echo "$FULL_DOMAIN"| tr '.' '-')
CERT_PATH="/etc/acme/data"
CERT_SAN="*.$FULL_DOMAIN"
PUBLIC_IP=""
#FULL_DOMAIN already set
REMOTE_ENV_IP=""
K8S_APISERVER_URL=""
K8S_APISERVER_CA=""
#SCI_TENANT comes from environment
#SCI_DOMAIN comes from environment
#SCI_CA_PEM comes from environment
#ADMIN_GROUP comes from environment
#ARM_CLIENT_ID comes from environment
#ARM_CLIENT_SECRET comes from environment
#ARM_TENANT_ID comes from environment
#AZURE_SUBSCIPTION_ID comes from environment
#VAULT_NAME comes from environment
#AZURE_BROKER_SUBSCRIPTION_ID comes from environment
#AZURE_BROKER_TENANT_ID comes from environment
#AZURE_BROKER_CLIENT_ID comes from environment
#AZURE_BROKER_CLIENT_SECRET comes from environment
#KYMA_RELEASES_AZURE_BLOBSTORE_KEY comes from environment
#UI_TEST_USER comes from environment
#UI_TEST_PASSWORD comes from environment

function azure_login {
  az login --service-principal -u ${ARM_CLIENT_ID} -p ${ARM_CLIENT_SECRET} --tenant ${ARM_TENANT_ID} > /dev/null
  az account set --subscription ${AZURE_SUBSCIPTION_ID} > /dev/null
}

function fetchPublicIP {
  echo -e "\nFetching public IP"

  local RG_NAME="kyma-cluster-$FULL_DOMAIN"

  PUBLIC_IP=$(az network public-ip show --resource-group ${RG_NAME} --name ${KYMA_PUBLIC_IP_NAME} | jq -r .ipAddress)
  if [[ -z "$PUBLIC_IP" ]]
  then
    echo -e "\nNo public IP with name: ${KYMA_PUBLIC_IP_NAME} found in resource group: $RG_NAME"
    exit 1
  fi

  REMOTE_ENV_IP=$(az network public-ip show --resource-group ${RG_NAME} --name ${KYMA_REMOTE_ENV_IP_NAME} | jq -r .ipAddress)
  if [[ -z "$REMOTE_ENV_IP" ]]
  then
    echo -e "\nNo public IP with name: ${KYMA_REMOTE_ENV_IP_NAME} found in resource group: $RG_NAME"
    exit 1
  fi
}

function parseYaml {
   local prefix=$2
   local s='[[:space:]]*' w='[a-zA-Z0-9_-]*' fs=$(echo @|tr @ '\034')
   sed -ne "s|^\($s\):|\1|" \
        -e "s|^\($s\)\($w\)$s:$s[\"']\(.*\)[\"']$s\$|\1$fs\2$fs\3|p" \
        -e "s|^\($s\)\($w\)$s:$s\(.*\)$s\$|\1$fs\2$fs\3|p"  $1 |
   awk -F$fs '{
      indent = length($1)/2;
      vname[indent] = $2;
      for (i in vname) {if (i > indent) {delete vname[i]}}
      if (length($3) > 0) {
         vn=""; for (i=0; i<indent; i++) {vn=(vn)(vname[i])("_")}
         gsub("-", "_", $2)
         printf("%s%s%s=\"%s\"\n", "'$prefix'",vn, $2, $3);
      }
   }'
}

function getK8sApiServerParams {

  if [ -n "${KUBECONFIG}" ]; then

    K8S_APISERVER_URL=$(cat ${KUBECONFIG} | jq -r '.clusters[0].cluster.server')

    if [ -n "${K8S_APISERVER_URL}" ]; then

      K8S_APISERVER_CA=$(cat ${KUBECONFIG} | jq -r '.clusters[0].cluster["certificate-authority-data"]')
      if [ -z "${K8S_APISERVER_CA}" ]; then
        K8S_APISERVER_CA=$(cat ${KUBECONFIG} | jq -r '.clusters[0].cluster["certificate-authority"]')
      fi

    else

      eval $(parseYaml ${KUBECONFIG} "KUBE_CONFIG_")

      K8S_APISERVER_URL="${KUBE_CONFIG_clusters__server}"

      if [ -n "${KUBE_CONFIG_clusters__certificate_authority_data}" ]; then
        K8S_APISERVER_CA="${KUBE_CONFIG_clusters__certificate_authority_data}"
      elif [ -n "${KUBE_CONFIG_clusters__certificate_authority}" ]; then
        K8S_APISERVER_CA="${KUBE_CONFIG_clusters__certificate_authority}"
      fi
    fi
  fi
}

function getCertificateDataFromVault {
  mkdir -p $CERT_PATH/$CERT_SAN
  az keyvault secret download --name "$VAULT_SECRET_NAME_PREFIX-privateKey" --vault-name $VAULT_NAME --file "$CERT_PATH/$CERT_SAN/$CERT_SAN.key" --encoding base64
  az keyvault secret download --name "$VAULT_SECRET_NAME_PREFIX-certificate" --vault-name $VAULT_NAME --file "$CERT_PATH/$CERT_SAN/combined.cer" --encoding base64

  mkdir -p $CERT_PATH/ingress
	az keyvault secret download --name "$VAULT_SECRET_NAME_PREFIX-ingress-ca-certificate" --vault-name $VAULT_NAME --file "$CERT_PATH/ingress/ca.pem" --encoding base64
	az keyvault secret download --name "$VAULT_SECRET_NAME_PREFIX-ingress-ca-key" --vault-name $VAULT_NAME --file "$CERT_PATH/ingress/ca.key" --encoding base64

  setCertVariables "$CERT_PATH/$CERT_SAN/combined.cer" "$CERT_PATH/$CERT_SAN/$CERT_SAN.key" "$CERT_PATH/ingress/ca.pem" "$CERT_PATH/ingress/ca.key"
}

function setCertVariables {
  echo -e "\nEncoding certificates"

  #This must be base64-encoded
  TLS_CERT=$(cat $1 | base64 | tr -d '\n')
  TLS_KEY=$(cat $2 | base64 | tr -d '\n')
  REMOTE_ENV_CA=$(cat $3 | base64 | tr -d '\n')
  REMOTE_ENV_CA_KEY=$(cat $4 | base64 | tr -d '\n')
}

azure_login
fetchPublicIP
getK8sApiServerParams
getCertificateDataFromVault

echo -e "\nGenerating secret for Cluster certificate"
bash ${ROOT_PATH}/create-generic-secret.sh "${ROOT_PATH}/../resources/cluster-certificate-secret.yaml" \
  "TLS_CERT" "${TLS_CERT}" \
  "TLS_KEY" "${TLS_KEY}"

echo -e "\nApplying asecret for Cluster certificate"
kubectl create -f "${ROOT_PATH}/../resources/cluster-certificate-secret.yaml"

##########

echo -e "\nGenerating secret for Remote Env certificate"
bash ${ROOT_PATH}/create-generic-secret.sh "${ROOT_PATH}/../resources/remote-env-certificate-secret.yaml" \
  "REMOTE_ENV_CA" "${REMOTE_ENV_CA}" \
  "REMOTE_ENV_CA_KEY" "${REMOTE_ENV_CA_KEY}"

echo -e "\nApplying asecret for C Remote Env certificate"
kubectl create -f "${ROOT_PATH}/../resources/remote-env-certificate-secret.yaml"

##########

echo -e "\nGenerating secret for UI Test"
bash ${ROOT_PATH}/create-generic-secret.sh "${ROOT_PATH}/../resources/ui-test-secret.yaml" \
  "UI_TEST_USER" "${UI_TEST_USER}" \
  "UI_TEST_PASSWORD" "${UI_TEST_PASSWORD}"

echo -e "\nApplying asecret for UI Test"
kubectl create -f "${ROOT_PATH}/../resources/ui-test-secret.yaml"

##########

echo -e "\nGenerating secret for Azure Broker"
bash ${ROOT_PATH}/create-generic-secret.sh "${ROOT_PATH}/../resources/azure-broker-secret.yaml" \
  "AZURE_BROKER_SUBSCRIPTION_ID" "${AZURE_BROKER_SUBSCRIPTION_ID}" \
  "AZURE_BROKER_TENANT_ID" "${AZURE_BROKER_TENANT_ID}" \
  "AZURE_BROKER_CLIENT_ID" "${AZURE_BROKER_CLIENT_ID}" \
  "AZURE_BROKER_CLIENT_SECRET" "${AZURE_BROKER_CLIENT_SECRET}"

echo -e "\nApplying asecret for Azure Broker"
kubectl create -f "${ROOT_PATH}/../resources/azure-broker-secret.yaml"

##########

echo -e "\nGenerating secret for azure blobstore"
bash ${ROOT_PATH}/create-generic-secret.sh "${ROOT_PATH}/../resources/azure-blobstore-secret.yaml" \
  "KYMA_RELEASES_AZURE_BLOBSTORE_KEY" "${KYMA_RELEASES_AZURE_BLOBSTORE_KEY}"

echo -e "\nApplying azure blobstore secret"
kubectl create -f ${ROOT_PATH}/../resources/azure-blobstore-secret.yaml

##########

echo -e "\nGenerating config map for installation"
bash ${ROOT_PATH}/create-config-map.sh \
--ip-address "${PUBLIC_IP}" \
--domain "${FULL_DOMAIN}" \
--remote-env-ip "${REMOTE_ENV_IP}" \
--k8s-apiserver-url "${K8S_APISERVER_URL}" \
--k8s-apiserver-ca "${K8S_APISERVER_CA}" \
--admin-group "${ADMIN_GROUP}" \
--output ${ROOT_PATH}/../resources/installation-config.yaml

echo -e "\nApplying config map..."
kubectl create -f ${ROOT_PATH}/../resources/installation-config.yaml

