#!/bin/bash

###
# Following script provisions GKE cluster and deprovision.
#
# INPUTS:
# - GCLOUD_SERVICE_KEY_PATH - content of service account credentials json file
# - GCLOUD_PROJECT_NAME - name of GCP project
# - CLUSTER_NAME - name for the new cluster
# - GCLOUD_COMPUTE_ZONE - zone in which the new cluster will be provisioned
#
# OPTIONAL:
# - CLUSTER_VERSION - the k8s version to use for the master and nodes
# - MACHINE_TYPE - the type of machine to use for nodes
# - NUM_NODES - the number of nodes to be created
#
###

set -o errexit

discoverUnsetVar=false

for var in GCLOUD_SERVICE_KEY_PATH GCLOUD_PROJECT_NAME CLUSTER_NAME GCLOUD_COMPUTE_ZONE; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT_DIR=${CURRENT_DIR}/..

echo "
################################################################################
# Provisioning gke cluster
################################################################################
"

bash ${ROOT_DIR}/prow/scripts/provision-gke-cluster.sh

# TODO add other parts

echo "
################################################################################
# Deprovisioning gke cluster
################################################################################
"

bash ${ROOT_DIR}/prow/scripts/deprovision-gke-cluster.sh