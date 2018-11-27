#!/usr/bin/env bash

# Following script checks if GCP Broker is registered in Service Catalog
#
# Expected environment variables:
#   - WORKING_NAMESPACE - name of the namespace where GCP Broker should be installed
#
# Input flags:
#   --sleep-duration-sec
#   --max-retries

OPERATION=""
POSITIONAL=()
SLEEP_DURATION_SEC=3
MAX_RETRIES=40
while [[ $# -gt 0 ]]
do
    key="$1"

    case ${key} in
        --sleep-duration-sec)
            SLEEP_DURATION_SEC="$2"
            shift # past argument
            shift # past value
            ;;
        --max-retries)
            MAX_RETRIES="$2"
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

function printToolsVersions() {
    printf "kubectl version\n"
    kubectl version --short

    printf "\n"
}

function checkIfGCPBrokerIsRegistered() {
    local cnt=0

    echo "Checking if GCP ServiceBroker is registered in Service Catalog..."
    set +o errexit

    while :
    do
      local status=$(kubectl get servicebroker gcp-broker -n ${WORKING_NAMESPACE} -o jsonpath='{ .status.conditions[?(@.type=="Ready")].status }')

      if [[ "${status}" == "True" ]]
        then
          echo "GCP ServiceBroker is registered successfully."
          break
        else
          ((cnt++))
          if (( cnt > $MAX_RETRIES )); then
            echo "Max retries has been reached (retries $MAX_RETRIES). Expected GCP Broker Ready status to be True but got '${status}'."
            exit 1
          fi

          echo "[Retry ${cnt}/${MAX_RETRIES}] GCP ServiceBroker condition Ready is not set to True - retry in ${SLEEP_DURATION_SEC}s"
          sleep ${SLEEP_DURATION_SEC}
        fi
    done

    set -o errexit
}

printToolsVersions
checkIfGCPBrokerIsRegistered
