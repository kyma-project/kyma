#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

FILE_NAME=sample-namespace.yaml
NAMESPACE=toad-test-1
BOOTSTRAP_ADMIN=kyma-admin-role
BOOTSTRAP_VIEWER=kyma-reader-role
GLOBAL_COUNTDOWN=20
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
RESOURCE_LIMITS_NAME=kyma-default

EXPECTED_LIMIT_RANGE_MEMORY_DEFAULT_REQUEST=$(echo ${EXPECTED_LIMIT_RANGE_MEMORY_DEFAULT_REQUEST:-32Mi} | ./quantity-to-int)
EXPECTED_LIMIT_RANGE_MEMORY_DEFAULT=$(echo ${EXPECTED_LIMIT_RANGE_MEMORY_DEFAULT:-96Mi} | ./quantity-to-int)
EXPECTED_LIMIT_RANGE_MEMORY_MAX=$(echo ${EXPECTED_LIMIT_RANGE_MEMORY_MAX:-1Gi} | ./quantity-to-int)
EXPECTED_RESOURCE_QUOTA_LIMITS_MEMORY=$(echo ${EXPECTED_RESOURCE_QUOTA_LIMITS_MEMORY:-1Gi} | ./quantity-to-int)
EXPECTED_RESOURCE_QUOTA_REQUESTS_MEMORY=$(echo ${EXPECTED_RESOURCE_QUOTA_REQUESTS_MEMORY:-768Mi} | ./quantity-to-int)

getAdminRole(){
    RESULT=$(kubectl get roles/${BOOTSTRAP_ADMIN} -n ${NAMESPACE} -o jsonpath='{.metadata.name}' || :)
}

getViewerRole(){
    RESULT=$(kubectl get roles/${BOOTSTRAP_VIEWER} -n ${NAMESPACE} -o jsonpath='{.metadata.name}' || :)
}

getIstioInjectionLabel(){
    RESULT=$(kubectl get ns ${NAMESPACE} -o jsonpath='{.metadata.labels.istio-injection}' || :)
}

createNewEnv(){
    kubectl create -f ${DIR}/${FILE_NAME}
}

clearNamespace(){
    kubectl delete -f ${DIR}/${FILE_NAME}
}

testAdminRole(){
    local COUNTDOWN=GLOBAL_COUNTDOWN

    while [[ $COUNTDOWN -gt 0 ]]
    do
        getAdminRole

        if [[ "${RESULT}" != "" ]]
	    then
	        return 0
	    else
            echo "Waiting for admin role..."
            ((COUNTDOWN--))
            sleep 2
	    fi
    done

    echo "Error on getting admin role..."
    return 1
}

testViewerRole(){
    local COUNTDOWN=GLOBAL_COUNTDOWN

    while [[ $COUNTDOWN -gt 0 ]]
    do
        getViewerRole

        if [[ "${RESULT}" != "" ]]
	    then
	        return 0
	    else
            echo "Waiting for viewer role..."
            ((COUNTDOWN--))
            sleep 2
	    fi
    done

    echo "Error on getting viewer role..."
    return 1
}

testIstioInjectionLabel(){
    local COUNTDOWN=GLOBAL_COUNTDOWN

    while [[ $COUNTDOWN -gt 0 ]]
    do
        getIstioInjectionLabel

        if [[ "${RESULT}" != "" ]]
        then
            return 0
        else
            echo "Waiting for istio-injection label..."
            ((COUNTDOWN--))
            sleep 2
        fi
    done

    echo "Error getting istio-injection label"
    return 1
}

testLimitRangeCreation(){
    local COUNTDOWN=GLOBAL_COUNTDOWN

    while [[ $COUNTDOWN -gt 0 ]]
    do
        local RESULT=$(kubectl get limitrange ${RESOURCE_LIMITS_NAME} -n ${NAMESPACE} || :)
        local MEMORY_DEFAULT_REQUEST=$(kubectl get limitrange ${RESOURCE_LIMITS_NAME} -n ${NAMESPACE} -o jsonpath='{.spec.limits[0].defaultRequest.memory}' | ./quantity-to-int || :)
        local MEMORY_DEFAULT=$(kubectl get limitrange ${RESOURCE_LIMITS_NAME} -n ${NAMESPACE} -o jsonpath='{.spec.limits[0].default.memory}' | ./quantity-to-int || :)
        local MEMORY_MAX=$(kubectl get limitrange ${RESOURCE_LIMITS_NAME} -n ${NAMESPACE} -o jsonpath='{.spec.limits[0].max.memory}' | ./quantity-to-int || :)

        if [[ "${RESULT}" != "" ]]
        then
            if [[ ${MEMORY_DEFAULT_REQUEST} != ${EXPECTED_LIMIT_RANGE_MEMORY_DEFAULT_REQUEST} ]]
            then
                echo "MEMORY_DEFAULT_REQUEST: ${MEMORY_DEFAULT_REQUEST} not equals ${EXPECTED_LIMIT_RANGE_MEMORY_DEFAULT_REQUEST}"
                return 1
            fi

            if [[ ${MEMORY_DEFAULT} != ${EXPECTED_LIMIT_RANGE_MEMORY_DEFAULT} ]]
            then
                echo "MEMORY_DEFAULT: ${MEMORY_DEFAULT} not equals ${EXPECTED_LIMIT_RANGE_MEMORY_DEFAULT}"
                return 1
            fi

            if [[ ${MEMORY_MAX} != ${EXPECTED_LIMIT_RANGE_MEMORY_MAX} ]]
            then
                echo "MEMORY_MAX: ${MEMORY_MAX} not equals ${EXPECTED_LIMIT_RANGE_MEMORY_MAX}"
                return 1
            fi
            return 0
        else
            echo "Waiting for a limit range..."
            ((COUNTDOWN--))
            sleep 2
        fi
    done

    echo "Error getting limit range"
    return 1
}

testResourceQuotaCreation(){
    local COUNTDOWN=GLOBAL_COUNTDOWN

    while [[ $COUNTDOWN -gt 0 ]]
    do
        local RESULT=$(kubectl get resourcequota ${RESOURCE_LIMITS_NAME} -n ${NAMESPACE} || :)
        local LIMITS_MEMORY=$(kubectl get resourcequota ${RESOURCE_LIMITS_NAME} -n ${NAMESPACE} -o "jsonpath={.spec.hard['limits\.memory']}" | ./quantity-to-int || :)
        local REQUESTS_MEMORY=$(kubectl get resourcequota ${RESOURCE_LIMITS_NAME} -n ${NAMESPACE} -o "jsonpath={.spec.hard['requests\.memory']}" | ./quantity-to-int || :)

        if [[ "${RESULT}" != "" ]]
        then
            if [[ ${LIMITS_MEMORY} != ${EXPECTED_RESOURCE_QUOTA_LIMITS_MEMORY} ]]
            then
                echo "LIMITS_MEMORY: ${LIMITS_MEMORY} not equals ${EXPECTED_RESOURCE_QUOTA_LIMITS_MEMORY}"
                return 1
            fi
            if [[ ${REQUESTS_MEMORY} != ${EXPECTED_RESOURCE_QUOTA_REQUESTS_MEMORY} ]]
            then
                echo "REQUESTS_MEMORY: ${REQUESTS_MEMORY} not equals ${EXPECTED_RESOURCE_QUOTA_REQUESTS_MEMORY}"
                return 1
            fi
            return 0
        else
            echo "Waiting for a limit range..."
            ((COUNTDOWN--))
            sleep 2
        fi
    done

    echo "Error getting resource quota"
    return 1
}

echo "------------------------------------"
echo "Test injecting roles and limits"
echo "------------------------------------"

echo "------------------------------------"
echo "Creating namespace... "
echo "------------------------------------"
createNewEnv

echo "------------------------------------"
echo "Testing admin role... "
echo "------------------------------------"
testAdminRole

echo "------------------------------------"
echo "Testing viewer role... "
echo "------------------------------------"
testViewerRole

echo "------------------------------------"
echo "Testing istio-injection label..."
echo "------------------------------------"
testIstioInjectionLabel

echo "------------------------------------"
echo "Testing limit range creation..."
echo "------------------------------------"
testLimitRangeCreation

echo "------------------------------------"
echo "Testing resource quota creation..."
echo "------------------------------------"
testResourceQuotaCreation

echo "------------------------------------"
echo "Deleting namespace... "
echo "------------------------------------"
clearNamespace

echo "------------------------------------"
echo "All tests passed..."
echo "------------------------------------"
exit 0