#!/usr/bin/env bash

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
TMPDIR=`mktemp -d "$CURRENT_DIR/../../temp/XXXXXXXXXX"`
CR_PATH="$TMPDIR/installer-cr-local.yaml"

MINIKUBE_IP=$(minikube ip)
MINIKUBE_CA=$(cat ${HOME}/.minikube/ca.crt | base64 | tr -d '\n')

bash $CURRENT_DIR/create-cr.sh --output ${CR_PATH} --domain "kyma.local" --k8s-apiserver-url "https://${MINIKUBE_IP}:8443" --k8s-apiserver-ca "${MINIKUBE_CA}"

$CURRENT_DIR/installer.sh --local --cr ${CR_PATH}