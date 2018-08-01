#!/usr/bin/env bash

CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

echo "- Creating service account for tiller..."
kubectl -n kube-system create sa tiller
kubectl create clusterrolebinding tiller-cluster-admin --clusterrole=cluster-admin --serviceaccount=kube-system:tiller

echo "- Installing Tiller..."
kubectl create -f ${CURRENT_DIR}/../resources/tiller.yaml
bash ${CURRENT_DIR}/is-ready.sh kube-system name tiller
