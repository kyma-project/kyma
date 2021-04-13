#!/usr/bin/env bash

trap cleanup EXIT SIGTERM

cleanup() {
    if [[ -n "${REMOVE_PODS_FILE}" ]]; then
        echo
        echo "Removing temporary file with pods data: ${REMOVE_PODS_FILE}"
        rm "${REMOVE_PODS_FILE}"
    fi
}

# Retries a command on failure.
# $1 - the max number of attempts
# $2... - the command to run
retry() {
    local -r -i max_attempts="$1"; shift
    local -r cmd="$@"
    local -i attempt_num=1

    until $cmd
    do
        if (( attempt_num == max_attempts ))
        then
            echo "Attempt $attempt_num failed and there are no more attempts left!"
            exit $exitCode
        else
            echo "Attempt $attempt_num failed! Trying again in $attempt_num seconds..."
            sleep $(( attempt_num++ ))
        fi
    done
}


# A function that deletes the object and handles NotFound condition as a success.
tryDelete() {
    local namespace=$1
    local kind=$2
    local name=$3

    local code
    local result

    result=$(kubectl -n "${namespace}" delete "${kind}" "${name}" 2>&1); code=$?

    if [[ ${code} -eq 1 && ${result} == *"NotFound"* ]]
    then
        echo "        Delete operation failed with: \"${result}\". Handling as a success (the ${kind}/${name} is gone)"
        return 0
    fi

    return ${code}
}

# Deletes a pod
deletePod() {
    local namespace=$1
    local podName=$2

    if [[ "${dryRun}" == "false" ]]; then
        echo "    Deleting pod: ${namespace}/${podName}"
        retry "${retriesCount}" tryDelete "${namespace}" pod "${podName}"
        sleep "${sleepAfterPodDeleted}"
    else
        echo "    [dryrun]" kubectl -n "${namespace}" delete pod "${podName}"
    fi
}

# Helper function to handle "replicasets" map values
# Assign value if it does not already exist.
# Append to value using "/" as a separator if it already exists.
appendToMapValue() {
    local namespace=$1
    local parentObjectName=$2
    local podName=$3

    if [[ -z "${replicasets[${namespace}/${parentObjectName}]}" ]]
    then
        replicasets["${namespace}/${parentObjectName}"]="${podName}"
    else
        replicasets["${namespace}/${parentObjectName}"]="${replicasets[${namespace}/${parentObjectName}]}/${podName}"
    fi
}


########################################
# Global variables

declare -A objectsToRollout
declare -A replicasets
declare -A podsToDelete


########################################
# Configuration values check

if [ -z "$ISTIO_PROXY_IMAGE_PREFIX" ]; then
  echo "Error: required ISTIO_PROXY_IMAGE_PREFIX value is missing. Exiting..."
  exit $exitCode
fi

if [ -z "$ISTIO_PROXY_IMAGE_VERSION" ]; then
  echo "Error: required ISTIO_PROXY_IMAGE_VERSION value is missing. Exiting..."
  exit $exitCode
fi

# Retries count in case of an error
retriesCount=${RETRIES_COUNT:-5}
# Dry Run mode only prints commands. True by default.
dryRun="${DRY_RUN:-true}"
# Exit code for entire script. Zero by default means the script will terminate on errors, but it will not fail the process.
exitCode="${EXIT_CODE:-0}"
# Sleep time after pod is deleted
sleepAfterPodDeleted="${SLEEP_AFTER_POD_DELETED}:-0"

########################################
# Processing starts here

# TODO: check if this logic is still applicable and move it elsewhere, see issue: https://github.com/kyma-project/kyma/issues/11078
namespaces=$(retry "${retriesCount}" kubectl get ns -l kyma-project.io/created-by=e2e-upgrade-test-runner -o name | cut -d '/' -f2)

for NS in ${namespaces}; do
    if [[ "${dryRun}" == "false" ]]; then
        retry "${retriesCount}" kubectl delete replicasets -n "${NS}" --all
    else
        echo "[dryrun] kubectl delete rs -n ${NS}"
    fi
done

if [[ -z "${PODS_FILE}" ]]; then
    PODS_FILE=$(mktemp)
    REMOVE_PODS_FILE=${PODS_FILE}

    echo "Getting pods data into file: ${PODS_FILE}"
    allPods=$(retry "${retriesCount}" kubectl get po -A -o json)
    echo "${allPods}" > "${PODS_FILE}"
fi

echo "Processing pods data from file: ${PODS_FILE}"
istioProxyImage="${ISTIO_PROXY_IMAGE_PREFIX}:${ISTIO_PROXY_IMAGE_VERSION}"

#This query selects all pods that have containers with an istio-proxy image in a version other than expected.
#Istio proxy image is detected by image name prefix, by default: "eu.gcr.io/kyma-project/external/istio/proxyv2"
#Full image address is used to prevent from matching pods that are not connected with Istio but use the same version.
jqQuery='.items | .[] | select(.spec.containers[].image | startswith("'"${ISTIO_PROXY_IMAGE_PREFIX}"'") and (endswith("'"${istioProxyImage}"'") | not))  | "\(.metadata.name)/\(.metadata.namespace)"'

pods=$(jq -rc "${jqQuery}" < "${PODS_FILE}")
podArray=($(echo "${pods}" | tr " " "\n"))

echo
echo "Analyzing Pods - ${#podArray[@]} objects found."

for i in "${podArray[@]}"
do
    namespacedName=($(echo "$i" | tr "/" "\n"))

    podName="${namespacedName[0]}"
    namespace="${namespacedName[1]}"

    podJson=$(retry "${retriesCount}" kubectl get pod "${podName}" -n "${namespace}" -o json)

    #Skip pods in Terminating state
    podPhase=$(jq -r '.status.phase' <<< "${podJson}")
    case "${podPhase}" in
        ("Terminating")
            echo "    Pod ${podName} in terminating state. Skipping..."
            continue
            ;;
        (*)
            ;;
    esac

    parentObjectKind=$(jq -r '.metadata.ownerReferences[0].kind' <<< "${podJson}" | tr '[:upper:]' '[:lower:]')
    parentObjectName=$(jq -r '.metadata.ownerReferences[0].name' <<< "${podJson}")

    case "${parentObjectKind}" in
        ("null")
            ;&
        ("")
            echo "    Pod ${namespace}/${podName} has no parent object (standalone Pod). Skipping..."
            continue
            ;;
        ("replicaset")
            echo "    Pod \"${namespace}/${podName}\" is managed by the ReplicaSet \"${parentObjectName}\". Requires further processing."
            appendToMapValue "${namespace}" "${parentObjectName}" "${podName}"
            ;;
        ("replicationcontroller")
            echo "    Pod \"${namespace}/${podName}\" is managed by the ReplicationController \"${parentObjectName}\". Eligible for delete."
            podsToDelete["${namespace}/${podName}"]=""
            ;;
        (*)
            echo "    Pod \"${namespace}/${podName}\" is managed by \"${parentObjectKind}\" \"${parentObjectName}\". Eligible for rollout. "
            objectsToRollout["${parentObjectKind}/${namespace}/${parentObjectName}"]=""
            ;;
    esac
done


echo "Analyzing ReplicaSets - ${#replicasets[@]} objects found."
if [[ ${#replicasets[@]} -gt 0 ]]; then

   for key in "${!replicasets[@]}"
   do
        attributes=($(echo "${key}" | tr "/" "\n"))
        namespace="${attributes[0]}"
        replicasetName="${attributes[1]}"

        parentDeploymentName=$(retry "${retriesCount}" kubectl -n "${namespace}" get replicaset "${replicasetName}" -o jsonpath='{.metadata.ownerReferences[0].name}')

        case "${parentDeploymentName}" in
            ("null")
                ;&
            ("")
                echo "    ReplicaSet ${namespace}/${replicasetName} has no parent object. It's pods must be deleted"
                podsForReplicaset=$(echo "${replicasets[${key}]}" | tr "/" " ")
                for pod in ${podsForReplicaset}
                do
                  podsToDelete["${namespace}/${pod}"]=""
                done
                ;;
            (*)
                echo "    ReplicaSet ${namespace}/${replicasetName} has a parent deployment: ${parentDeploymentName}. Assigned for rollout"
                objectsToRollout["deployment/${namespace}/${parentDeploymentName}"]=""
                ;;
            esac
    done
fi

echo ""
echo "Processing objects..."
echo ""

echo "Number of pods to delete: ${#podsToDelete[@]}"
if [[ ${#podsToDelete[@]} -gt 0 ]]; then

    for key in "${!podsToDelete[@]}"
    do
        attributes=($(echo "${key}" | tr "/" "\n"))
        namespace="${attributes[0]}"
        podName="${attributes[1]}"

        deletePod "${namespace}" "${podName}"
    done
fi

echo "Number of objects to rollout: ${#objectsToRollout[@]}"
for key in "${!objectsToRollout[@]}"
do

    attributes=($(echo "${key}" | tr "/" "\n"))

    kind="${attributes[0]}"
    namespace="${attributes[1]}"
    name="${attributes[2]}"

    if [[ "${dryRun}" == "false" ]]; then
        retry "${retriesCount}" kubectl rollout restart "${kind}" "${name}" -n "${namespace}"
    else
        echo "    [dryrun] kubectl rollout restart ${kind} ${name} -n ${namespace}"
    fi

done
