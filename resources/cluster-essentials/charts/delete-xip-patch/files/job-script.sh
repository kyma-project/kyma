#!/usr/bin/env bash

RELEASE_NAME="xip-patch"
RELEASE_NAMESPACE="kyma-installer"
MAX_RETRIES=3

set +e

retry=0

#Find the xip-patch status.
while [[ ${retry} -lt ${MAX_RETRIES} ]]; do
    echo "Checking the status of the relase: \"${RELEASE_NAME}\" in namespace: \"${RELEASE_NAMESPACE}\":"
    result=$(helm3 status ${RELEASE_NAME} --namespace ${RELEASE_NAMESPACE} 2>&1)
    err=$?

    # Exit the loop on success
    if [[ ${err} -eq 0 ]]; then
        echo "Release found:"
        echo "${result}"
        break
    fi

    # Handle "release: not found" error.
    # Parsing the error message here, because we can't rely on helm exit codes.
    if [[ "${result}" == "Error: release: not found" ]]; then
        echo "release \"${RELEASE_NAME}\" not found. Skipping..."
        exit  0
    fi

    # Handle all other errors (timeouts/EOF etc.)
    echo "An error occured: ${result}"
    (( retry++ ))

    if [[ ${retry} -lt ${MAX_RETRIES} ]]; then
        echo "Retrying in 3s..."
        sleep 3
    fi
done

echo

if [[ ${retry} -eq ${MAX_RETRIES} ]]; then
    echo "Job failed: Maximum retries exceeded"
    exit 1
fi

#Delete the xip-patch release
while [[ ${retry} -lt ${MAX_RETRIES} ]]; do
    echo "Deleting the relase: \"${RELEASE_NAME}\" in namespace: \"${RELEASE_NAMESPACE}\":"
    result=$(helm3 delete ${RELEASE_NAME} --namespace ${RELEASE_NAMESPACE} 2>&1)
    err=$?

    # Exit the loop on success
    if [[ ${err} -eq 0 ]]; then
        echo "${result}"
        break
    fi

    # Handle errors
    echo "An error occured: ${result}"
    (( retry++ ))

    if [[ ${retry} -lt ${MAX_RETRIES} ]]; then
        echo "Retrying in 3s..."
        sleep 3
    fi
done

echo

if [[ ${retry} -eq ${MAX_RETRIES} ]]; then
    echo "Job failed: Maximum retries exceeded"
    exit 1
fi

# If here, it's a success.
echo "Job finished successfully :)"

