#!/usr/bin/env bash

# Following script provisions and deprovisions GCP Broker in given namespace.
#
# Expected environment variables:
#   - WORKING_NAMESPACE - name of the namespace where GCP Broker should be installed
#
# Input flags:
#   - action - possible values: provision or deprovision

set -o errexit # exit immediately if a command exits with a non-zero status.

readonly CURRENT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
readonly GCP_SECRET_NAME="gcp-broker-data"

function discoveredUnsetVar() {
    local discoveredUnsetVar=false
    for e in WORKING_NAMESPACE; do
        if [[ -z "${!e}" ]] ; then
            echo "ERROR: ${e} is not set"
            discoveredUnsetVar=true
        fi
    done

    if [[ "${discoveredUnsetVar}" = true ]] ; then
        exit 1
    fi
}

function configureGCloud() {
    if kubectl get secrets -n ${WORKING_NAMESPACE} | grep -q "${GCP_SECRET_NAME}"; then
        export GOOGLE_APPLICATION_CREDENTIALS=${CURRENT_DIR}"/sa-key.json"

        projectName=$(kubectl get secret -n ${WORKING_NAMESPACE} ${GCP_SECRET_NAME} -o jsonpath='{ .data.project-name }' | base64 --decode)
        saKey=$(kubectl get secret -n ${WORKING_NAMESPACE} ${GCP_SECRET_NAME} -o jsonpath='{ .data.sa-key }' | base64 --decode)

        local discoveredUnsetVar=false
        for e in projectName saKey; do
            if [[ -z "${!e}" ]] ; then
                echo "ERROR: Value for ${e} in Secret ${GCP_SECRET_NAME} cannot be empty."
                discoveredUnsetVar=true
            fi
        done

        if [[ "${discoveredUnsetVar}" = true ]] ; then
            exit 1
        fi

         echo "${saKey}" > "${GOOGLE_APPLICATION_CREDENTIALS}"

        gcloud config set project "${projectName}"
        gcloud auth activate-service-account --key-file=${GOOGLE_APPLICATION_CREDENTIALS}
    else
        return 1
    fi

    return 0
}

function printToolsVersions() {
    printf "kubectl version\n"
    kubectl version --short

    printf "\ngcloud version\n"
    gcloud version

    printf "\nsc version\n"
    sc version

    printf "\n"
}

# We are creating Jobs in a hook, and we cannot rely upon `helm delete` to remove the resources.
# To destroy such resources, we added "helm.sh/hook-delete-policy" annotation to the hook template file but only for `hook-succeeded`.
# When hooks are failed then we are not removing it automatically, to be able to see its logs. Because of that, we need to remove manually
# those Jobs because Tiller won't do that even on `helm delete` action.
#
# source: https://github.com/helm/helm/blob/master/docs/charts_hooks.md#hook-resources-are-not-managed-with-corresponding-releases
function cleanupAllReleaseJobs() {
    printf "Cleaning up helm release ${RELEASE_NAME} Jobs in namespace ${WORKING_NAMESPACE}\n"
    kubectl delete jobs -n ${WORKING_NAMESPACE} -l release=${RELEASE_NAME}

    deleteErr=$?
    if [[ ${deleteErr} -ne 0 ]]
    then
      printf "FAILED cleaning Jobs.\n"
      return 1
    fi
    printf "Success cleaning Jobs.\n"
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

printToolsVersions
discoveredUnsetVar

case "${OPERATION}" in
  "provision")
        set +o errexit
            configureGCloud
            cfgErr=$?
        set -o errexit

        if [[ ${cfgErr} -ne 0 ]]
        then
            echo "The gcloud tool cannot be configured. The secret ${GCP_SECRET_NAME} was not found in namespace ${WORKING_NAMESPACE}"
            exit 1
        fi

        cmd="sc add-gcp-broker --scope namespace --namespace ${WORKING_NAMESPACE}"
        echo "Executing command: $cmd"
        ${cmd}
        ;;

  "deprovision")
        flags="--skip-deprecated --scope namespace --namespace ${WORKING_NAMESPACE}"

        set +o errexit
            configureGCloud
            cfgErr=$?
        set -o errexit

        if [[ ${cfgErr} -ne 0 ]]
        then
            flags=${flags}" --skip-gcp-integration"
        fi

        cmd="sc remove-gcp-broker ${flags}"
        echo "Executing command: sc remove-gcp-broker ${flags}"
        ${cmd}
        cleanupAllReleaseJobs
        ;;

  "") echo "Flag action need to be specified" && exit 1 ;;
  *) echo "Incorrect value for --action flag. Only provision, deprovision or status-check are supported" && exit 1
esac
