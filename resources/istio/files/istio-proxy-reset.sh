#!/usr/bin/env bash

cleanup() {
    if [[ -n "${REMOVE_PODS_FILE}" ]]; then
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
            return 1
        else
            echo "Attempt $attempt_num failed! Trying again in $attempt_num seconds..."
            sleep $(( attempt_num++ ))
        fi
    done
}

getReplicationControllerSelector() {
    local namespace=$1
    local objectName=$2
    local pattern='go-template={{range $key, $value := .spec.selector}}{{ $key }}{{ "=" }}{{ $value }}{{ "\n" }}{{end}}'
    echo $(kubectl get -n "${namespace}" replicationcontroller "${objectName}" -o "${pattern}")
}

getPodByLabels() {
    local namespace=$1
    local labels=$2
    echo $(kubectl get pods -n "${namespace}" -o jsonpath='{.items[*].metadata.name}' -l "${labels}")
}

deletePodsForReplicationController() {
    local namespace=$1
    local objectName=$2
    local labels=$(retry "${RETRIES_COUNT}" getReplicationControllerSelector "${namespace}" "${objectName}")
    local labelsWithCommas=$(echo -n "${labels}" | sed 's/ /,/g')
    podsToDelete=$(retry "${RETRIES_COUNT}" getPodByLabels "${namespace}" "${labelsWithCommas}")

    echo
    echo "    Restarting $(echo "${podsToDelete}" | wc -w | sed 's/ //g') pods for ReplicationController: ${namespace}/${objectName}"

    for pod in ${podsToDelete};
    do
        if [[ "${dryRun}" == "false" ]]; then
            echo "        Deleting pod: ${namespace}/${pod}"
            sleep 2
            retry "${RETRIES_COUNT}" kubectl -n "${namespace}" delete pod "${pod}"
        else
            echo "        [dryrun]" kubectl -n "${namespace}" delete pod "${pod}"
        fi
    done
}

trap cleanup EXIT SIGINT SIGTERM

#By default each command that requires retries will attempt at most 5 times
RETRIES_COUNT=5 #TODO Configurable by env

#We expect all sidecars to be in this version
expectedIstioProxyVersion="${EXPECTED_ISTIO_PROXY_IMAGE:-eu.gcr.io/kyma-project/external/istio/proxyv2:1.9.1-distroless}"
#We use this image prefix to detect if there's an istio proxy in the Pod. If you have different prefixes (shouldn't be the case), just define several jobs.
istioProxyImageNamePrefix="${COMMON_ISTIO_PROXY_IMAGE_PREFIX:-eu.gcr.io/kyma-project/external/istio/proxyv2}"

dryRun="${DRY_RUN:-true}"

#TODO: Is this is a cleanup? If so, it shouldn't be here
namespaces=$(retry "${RETRIES_COUNT}" kubectl get ns -l kyma-project.io/created-by=e2e-upgrade-test-runner -o name | cut -d '/' -f2)

for NS in ${namespaces}; do
    if [[ "${dryRun}" == "false" ]]; then
        retry "${RETRIES_COUNT}" kubectl delete replicasets -n "${NS}" --all
    else
        echo "[dryrun] kubectl delete rs -n ${NS}"
    fi
done

declare -A objectsToRestart
declare -A replicationControllers


if [[ -z "${PODS_FILE}" ]]; then
    PODS_FILE=$(mktemp)
    REMOVE_PODS_FILE=${PODS_FILE}
fi

allPods=$(retry "${RETRIES_COUNT}" kubectl get po -A -o json)
echo "${allPods}" > ${PODS_FILE}
echo "Processing pods data from: ${PODS_FILE}"

#This query selects all pods that have containers with an istio-proxy image in a version other than expected.
#Istio proxy image is detected by image name prefix, by default: "eu.gcr.io/kyma-project/external/istio/proxyv2"
jqQuery='.items | .[] | select(.spec.containers[].image | startswith("'"${istioProxyImageNamePrefix}"'") and (endswith("'"${expectedIstioProxyVersion}"'") | not))  | "\(.metadata.name)/\(.metadata.namespace)"'

pods=$(jq -rc "${jqQuery}" < ${PODS_FILE})
podArray=($(echo "${pods}" | tr " " "\n"))

echo "Number of pods matched: ${#podArray[@]}"

for i in "${podArray[@]}"
do
    namespacedName=($(echo $i | tr "/" "\n"))

    podName="${namespacedName[0]}"
    namespace="${namespacedName[1]}"

    podJson=$(retry "${RETRIES_COUNT}" kubectl get pod "${podName}" -n "${namespace}" -o json)

    parentObjectKind=$(jq -r '.metadata.ownerReferences[0].kind' <<< "${podJson}" | tr '[:upper:]' '[:lower:]')
    parentObjectName=$(jq -r '.metadata.ownerReferences[0].name' <<< "${podJson}")

    case "${parentObjectKind}" in
        ("null")
            ;&
        ("")
            echo "Pod ${podName} in namespace ${namespace} has no parent object (standalone Pod). Skipping..."
            continue
            ;;
        ("replicaset")
            parentDeploymentName=$(retry "${RETRIES_COUNT}" kubectl get "${parentObjectKind}" "${parentObjectName}" -n "${namespace}" -o jsonpath='{.metadata.ownerReferences[0].name}')
            echo "deployment ${parentDeploymentName} in namespace ${namespace} eligible for restart"
            objectsToRestart["deployment/${namespace}/${parentDeploymentName}"]=""
            ;;
        ("replicationcontroller")
            replicationControllers["${namespace}/${parentObjectName}"]=""
            ;;
        (*)
            echo "${parentObjectKind} ${parentObjectName} in namespace ${namespace} eligible for restart"
            objectsToRestart["${parentObjectKind}/${namespace}/${parentObjectName}"]=""
            ;;
    esac
done

if [[ ${#replicationControllers[@]} -gt 0 ]]; then
    echo "Processing ReplicationControllers"

    for key in "${!replicationControllers[@]}"
    do
        attributes=($(echo "${key}" | tr "/" "\n"))
        namespace="${attributes[0]}"
        name="${attributes[1]}"

        deletePodsForReplicationController "${namespace}" "${name}"
    done
fi

echo "Number of objects to restart: ${#objectsToRestart[@]}"

for key in "${!objectsToRestart[@]}"
do

    attributes=($(echo "${key}" | tr "/" "\n"))

    kind="${attributes[0]}"
    namespace="${attributes[1]}"
    name="${attributes[2]}"

    if [[ "${dryRun}" == "false" ]]; then
        retry "${RETRIES_COUNT}" kubectl rollout restart "${kind}" "${name}" -n "${namespace}"
    else
        echo "[dryrun] kubectl rollout restart ${kind} ${name} -n ${namespace}"
    fi

done
