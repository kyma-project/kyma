#!/usr/bin/env bash

ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

echo "- Creating service account for tiller..."
kubectl -n kube-system create sa tiller
kubectl create clusterrolebinding tiller-cluster-admin --clusterrole=cluster-admin --serviceaccount=kube-system:tiller

echo "- Installing Tiller..."
kubectl create -f ${ROOT_PATH}/tiller.yaml
bash ${ROOT_PATH}/../../installation/scripts/is-ready.sh kube-system name tiller
