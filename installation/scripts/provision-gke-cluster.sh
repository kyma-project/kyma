#!/bin/bash

###
# Following script provisions GKE cluster.
#
# INPUTS:
# - GCLOUD_SERVICE_KEY - gziped, base64-ed content of service account credentials json file
# - GCLOUD_PROJECT_NAME - name of GCP project
# - CLUSTER_NAME - name for the new cluster
# - GCLOUD_COMPUTE_ZONE - zone in which the new cluster will be provisioned
#
# OPTIONAL:
# - CLUSTER_VERSION - the k8s version to use for the main and nodes
# - MACHINE_TYPE - the type of machine to use for nodes
# - NUM_NODES - the number of nodes to be created
#
# REQUIREMENTS:
# - gcloud
# - gunzip
###

set -o errexit

discoverUnsetVar=false

for var in GCLOUD_SERVICE_KEY GCLOUD_PROJECT_NAME CLUSTER_NAME GCLOUD_COMPUTE_ZONE; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

CLUSTER_VERSION_PARAM=""
MACHINE_TYPE_PARAM=""
NUM_NODES_PARAM=""

if [ ${CLUSTER_VERSION} ]; then CLUSTER_VERSION_PARAM="--cluster-version=${CLUSTER_VERSION}"; fi
if [ ${MACHINE_TYPE} ]; then MACHINE_TYPE_PARAM="--machine-type=${MACHINE_TYPE}"; fi
if [ ${NUM_NODES} ]; then NUM_NODES_PARAM="--num-nodes=${NUM_NODES}"; fi

command -v gcloud
command -v gunzip

KEY_FILE=${HOME}/gcp.json
echo ${GCLOUD_SERVICE_KEY} | base64 --decode | gunzip > ${KEY_FILE}

gcloud auth activate-service-account --key-file=${KEY_FILE}
gcloud config set project ${GCLOUD_PROJECT_NAME}
gcloud config set compute/zone ${GCLOUD_COMPUTE_ZONE}

gcloud container clusters create ${CLUSTER_NAME} ${CLUSTER_VERSION_PARAM} ${MACHINE_TYPE_PARAM} ${NUM_NODES_PARAM}
