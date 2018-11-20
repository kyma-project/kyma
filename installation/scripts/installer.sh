#!/usr/bin/env bash

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
RESOURCES_DIR="${CURRENT_DIR}/../resources"
INSTALLER="${RESOURCES_DIR}/installer.yaml"
INSTALLER_CONFIG=""
FEATURE_GATES_CONFIG="${RESOURCES_DIR}/installer-feature-gates.yaml.tpl"
KNATIVE_CONFIG="${RESOURCES_DIR}/installer-config-knative.yaml.tpl"

POSITIONAL=()
while [[ $# -gt 0 ]]
do
    key="$1"

    case ${key} in
        --local)
            LOCAL=1
            shift
            ;;
        --feature-gates)
            FEATURE_GATES="$2"
            shift
            shift
            ;;
        --cr)
            CR_PATH="$2"
            shift # past argument
            shift # past value
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

kubectl apply -f ${RESOURCES_DIR}/default-sa-rbac-role.yaml

bash ${CURRENT_DIR}/is-ready.sh kube-system k8s-app kube-dns
bash ${CURRENT_DIR}/install-tiller.sh

if [ $LOCAL ]; then
    INSTALLER="${RESOURCES_DIR}/installer-local.yaml"
    INSTALLER_CONFIG="${RESOURCES_DIR}/installer-config-local.yaml.tpl"
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
bash ${CURRENT_DIR}/concat-yamls.sh ${INSTALLER} ${INSTALLER_CONFIG} ${CR_PATH} ${KNATIVE_CONFIG} | kubectl apply -f -

echo -e "\nEnabled feature gates: $FEATURE_GATES"
cat ${FEATURE_GATES_CONFIG} | sed 's/__FEATURE_GATES__/'${FEATURE_GATES}'/g' | kubectl apply -f -

echo -e "\nTriggering installation"
kubectl label installation/kyma-installation action=install
