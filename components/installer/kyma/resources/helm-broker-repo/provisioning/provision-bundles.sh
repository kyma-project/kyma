#!/bin/bash

################################################################################
#
# Copies bundles from given source into Helm-Broker repository, converts them to tgz files and generates index.yaml
# $1 - bundles source folder
# Sample: bash provision-bundles.sh bundles
#
################################################################################

NAMESPACE=kyma-system
SOURCE=$1
BUNDLES_DIR_NAME=$(basename ${SOURCE})
PROVISIONING_DIR=$(dirname $0)

set -o errexit

if ! [ -d $1 ]
then
    echo "$1 does not exists"
    exit 1
fi
kubectl apply -f ${PROVISIONING_DIR}/pvc.yaml -n ${NAMESPACE}
kubectl apply -f ${PROVISIONING_DIR}/po.yaml -n ${NAMESPACE}

bash ${PROVISIONING_DIR}/../../../installation/scripts/is-ready.sh ${NAMESPACE} app  bundles-repository-provisioner

kubectl exec bundles-repository-provisioner -n ${NAMESPACE} -- rm -rf /data/repository
kubectl exec bundles-repository-provisioner -n ${NAMESPACE} -- rm -rf /data/bundles

kubectl exec  bundles-repository-provisioner -n ${NAMESPACE} -- mkdir /data/repository
kubectl cp ${SOURCE} bundles-repository-provisioner:/data/ -n ${NAMESPACE}
kubectl exec bundles-repository-provisioner -n ${NAMESPACE} -- targz /data/${BUNDLES_DIR_NAME} /data/repository
kubectl exec bundles-repository-provisioner -n ${NAMESPACE} -- indexbuilder -s /data/${BUNDLES_DIR_NAME} -d /data/repository

kubectl delete pod bundles-repository-provisioner -n ${NAMESPACE}