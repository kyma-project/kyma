#!/bin/bash

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

POSITIONAL=()
while [[ $# -gt 0 ]]
do
    key="$1"

    case ${key} in
        --local)
            LOCAL=true
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

kubectl apply -f ${CURRENT_DIR}/../../resources/cluster-prerequisites/default-sa-rbac-role.yaml
kubectl apply -f ${CURRENT_DIR}/../../resources/cluster-prerequisites/limit-range.yaml -n "kyma-installer"
kubectl apply -f ${CURRENT_DIR}/../../resources/cluster-prerequisites/resource-quotas-installer.yaml

bash ${CURRENT_DIR}/install-tiller.sh

kubectl apply -f ${CURRENT_DIR}/../resources/installer.yaml -n "kyma-installer"

${CURRENT_DIR}/is-ready.sh kube-system k8s-app kube-dns

if [ $LOCAL ]; then
    bash ${CURRENT_DIR}/copy-resources.sh
fi

if [ $CR_PATH ]; then

    case $CR_PATH in
    /*) ;;
    *) CR_PATH="$(pwd)/$CR_PATH";;
    esac

    if [ -f $CR_PATH ]; then
        echo "Applying CR for installer from path $CR_PATH"
        kubectl apply -f $CR_PATH
    else
        echo "CR file not found in path $CR_PATH"
    fi

fi
