#!/usr/bin/env bash

# Following script provisions and deprovisions GKE cluster.
#
# Expected environment variables:
#   - WORKING_NAMESPACE - name of the namespace where GCP Broker should be installed
#   - CLOUDSDK_CORE_PROJECT - name of GCP project
#   - GOOGLE_APPLICATION_CREDENTIALS - path to the service account credentials file
#
# Input flags:
#   - action - possible values: provision or deprovision

set -o errexit

function discoveredUnsetVar() {
    local discoveredUnsetVar=false
    for e in WORKING_NAMESPACE CLOUDSDK_CORE_PROJECT GOOGLE_APPLICATION_CREDENTIALS; do
        if [ -z "${!e}" ] ; then
            echo "ERROR: ${e} is not set"
            discoveredUnsetVar=true
        fi
    done

    if [ "${discoveredUnsetVar}" = true ] ; then
        exit 1
    fi
}

function configureGCloud() {
    gcloud auth activate-service-account --key-file=${GOOGLE_APPLICATION_CREDENTIALS}
}

OPERATION=""
POSITIONAL=()
while [[ $# -gt 0 ]]
do
    key="$1"

    case ${key} in
        --action)
            OPERATION="$2"
            shift # past argument
            shift # past value
            ;;
        *)    # unknown option
            POSITIONAL+=("$1") # save it in an array for later
            shift # past argument
            ;;
    esac
done
set -- "${POSITIONAL[@]}" # restore positional parameters

discoveredUnsetVar
configureGCloud


case "${OPERATION}" in
  "provision") sc add-gcp-broker --scope namespace --namespace ${WORKING_NAMESPACE} ;;
  "deprovision") sc remove-gcp-broker --skip-deprecated --scope namespace --namespace ${WORKING_NAMESPACE} ;;
  "") echo "Flag action need to be specified" && exit 1 ;;
  *) echo "Incorrect value for --action flag. Only provision or deprovision are supported" && exit 1
esac
