#!/usr/bin/env bash

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

#By default each command that requires retries will attempt at most 5 times
RETRIES_COUNT=5

#We expect all sidecars to be in this version
expectedIstioProxyVersion="${EXPECTED_ISTIO_PROXY_IMAGE:-eu.gcr.io/kyma-project/external/istio/proxyv2:1.9.1-distroless}"
#We use this image prefix to detect if there's an istio proxy in the Pod. If you have different prefixes (shouldn't be the case), just define several jobs.
istioProxyImageNamePrefix="${COMMON_ISTIO_PROXY_IMAGE_PREFIX:-eu.gcr.io/kyma-project/external/istio/proxyv2}"

dryRun="${DRY_RUN:-false}"

namespaces=$(retry "${RETRIES_COUNT}" kubectl get ns -l kyma-project.io/created-by=e2e-upgrade-test-runner -o name | cut -d '/' -f2)

for NS in ${namespaces}; do
    if [[ "${dryRun}" == "false" ]]; then
        retry "${RETRIES_COUNT}" kubectl delete rs -n "${NS}" --all
    else
        echo "[dryrun] kubectl delete rs -n ${NS}"
    fi
done

declare -A objectsToRestart

#This query selects all pods that have containers with an istio-proxy image in a version other than expected.
#Istio proxy image is detected by image name prefix, by default: "eu.gcr.io/kyma-project/external/istio/proxyv2"
jqQuery='.items | .[] | select(.spec.containers[].image | startswith("'"${istioProxyImageNamePrefix}"'") and (endswith("'"${expectedIstioProxyVersion}"'") | not))  | "\(.metadata.name)/\(.metadata.namespace)"'

pods=$(retry "${RETRIES_COUNT}" kubectl get po -A -o json | jq -rc "${jqQuery}" )

podArray=($(echo $pods | tr " " "\n"))

echo "NUMBER OF PODS MATCHED: ${#podArray[@]}"

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
            ;;
        ("")
            echo "Pod ${podName} in namespace ${namespace} has no parent object. Skipping..."
            continue
            ;;
        ("replicaset")
            parentDeploymentName=$(retry "${RETRIES_COUNT}" kubectl get "${parentObjectKind}" "${parentObjectName}" -n "${namespace}" -o jsonpath='{.metadata.ownerReferences[0].name}')
            echo "deployment ${parentDeploymentName} in namespace ${namespace} eligible for restart"
            objectsToRestart["deployment/${namespace}/${parentDeploymentName}"]=""
            ;;
        (*)
            echo "${parentObjectKind} ${parentObjectName} in namespace ${namespace} eligible for restart"
            objectsToRestart["${parentObjectKind}/${namespace}/${parentObjectName}"]=""
            ;;
    esac
done

echo "NUMBER OF OBJECTS TO RESTART: ${#objectsToRestart[@]}"

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
