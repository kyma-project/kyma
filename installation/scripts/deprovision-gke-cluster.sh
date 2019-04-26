#!/bin/bash

###
# Following script deprovisions GKE cluster.
#
# INPUTS:
# - GCLOUD_SERVICE_KEY - gziped, base64-ed content of service account credentials json file
# - GCLOUD_PROJECT_NAME - name of GCP project
# - CLUSTER_NAME - name for the new cluster
# - GCLOUD_COMPUTE_ZONE - zone in which the new cluster will be deprovisioned
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

command -v gcloud
command -v gunzip

KEY_FILE=${HOME}/gcp.json
echo ${GCLOUD_SERVICE_KEY} | base64 --decode | gunzip > ${KEY_FILE}

gcloud auth activate-service-account --key-file=${KEY_FILE}
gcloud config set project ${GCLOUD_PROJECT_NAME}
gcloud config set compute/zone ${GCLOUD_COMPUTE_ZONE}

gcloud container clusters delete ${CLUSTER_NAME} --quiet
