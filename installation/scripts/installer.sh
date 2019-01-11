#!/usr/bin/env bash

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
RESOURCES_DIR="${CURRENT_DIR}/../resources"
INSTALLER="${RESOURCES_DIR}/installer.yaml"
INSTALLER_CONFIG=""
AZURE_BROKER_CONFIG=""

function checkInputParameterValue() {
    if [ -z "${2}" ]; then
        echo "Value parameter for ${1} is empty"
        echo "Make sure parameter value is neither empty nor start with two hyphens"
        exit 1
    fi
    if [ "${2:0:2}" == "--" ]; then
        echo "Invalid parameter value for ${1}:"
        echo "${2}"
        echo "Make sure parameter value is neither empty nor start with two hyphens"
        exit 1
    fi
}

POSITIONAL=()
while [[ $# -gt 0 ]]
do
    
    key="$1"

    case ${key} in
        --local)
            LOCAL=1
            shift
            ;;
        --cr)
            checkInputParameterValue "$1" "$2"
            CR_PATH="$2"
            shift # past argument
            shift # past value
            ;;
        --knative)
            KNATIVE=true
            shift
            ;;
        *) # unknown option
            POSITIONAL+=("$1") # save it in an array for later
            shift # past argument
            ;;
    esac
done
set -- "${POSITIONAL[@]}" # restore positional parameters

echo "
################################################################################
# Kyma Installer setup
################################################################################
"

kubectl apply -f ${RESOURCES_DIR}/default-sa-rbac-role.yaml # to be deleted once the script is used in local scenario only

bash ${CURRENT_DIR}/is-ready.sh kube-system k8s-app kube-dns
bash ${CURRENT_DIR}/install-tiller.sh

if [ $LOCAL ]; then

    INSTALLER="${RESOURCES_DIR}/installer-local.yaml"
    INSTALLER_CONFIG="${RESOURCES_DIR}/installer-config-local.yaml.tpl"
    
    if [ -n "${AZURE_BROKER_SUBSCRIPTION_ID}" ]; then

        echo -e "\nAzure-Broker subscription ID found in environment variables. Enabling component..."
        bash ${CURRENT_DIR}/manage-component.sh "azure-broker" true

        echo -e "\nGenerating the secret for Azure Broker..."
        AZURE_BROKER_CONFIG=$(mktemp)
        bash ${CURRENT_DIR}/configure-azure-broker.sh ${AZURE_BROKER_CONFIG}

    fi

fi

if [ $CR_PATH ]; then

    case $CR_PATH in
    /*) ;;
    *) CR_PATH="$(pwd)/$CR_PATH";;
    esac

    if [ ! -f $CR_PATH ]; then
        echo "CR file not found in path $CR_PATH"
        exit 1
    fi

fi


echo -e "\nApplying installation combo yaml"
COMBO_YAML=$(bash ${CURRENT_DIR}/concat-yamls.sh ${INSTALLER} ${INSTALLER_CONFIG} ${AZURE_BROKER_CONFIG} ${CR_PATH})

rm -rf ${AZURE_BROKER_CONFIG}

if [ $KNATIVE ]; then
    COMBO_YAML=$(sed 's/global\.knative: .*/global.knative: "true"/g' <<<"$COMBO_YAML")
fi

kubectl apply -f - <<<"$COMBO_YAML"

echo -e "\nConfiguring sub-components"
bash ${CURRENT_DIR}/configure-components.sh

echo -e "\nTriggering installation"
kubectl label installation/kyma-installation action=install
