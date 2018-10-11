#!/bin/bash

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

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

kubectl apply -f ${CURRENT_DIR}/../resources/default-sa-rbac-role.yaml

bash ${CURRENT_DIR}/install-tiller.sh

if [ $LOCAL -eq 1 ]; then
    kubectl apply -f ${CURRENT_DIR}/../resources/installer-local.yaml
else
    kubectl apply -f ${CURRENT_DIR}/../resources/installer.yaml
fi

${CURRENT_DIR}/is-ready.sh kube-system k8s-app kube-dns

if [ $CR_PATH ]; then

    case $CR_PATH in
    /*) ;;
    *) CR_PATH="$(pwd)/$CR_PATH";;
    esac

    if [ -f $CR_PATH ]; then
        echo "Applying CR for installer from path $CR_PATH"
        kubectl apply -f $CR_PATH
        kubectl label installation/kyma-installation action=install
    else
        echo "CR file not found in path $CR_PATH"
    fi

fi
