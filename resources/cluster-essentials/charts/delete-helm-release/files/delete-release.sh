#!/usr/bin/env bash

########################################
# Must come from env:                  #
# RELEASE_NAME                         #
# RELEASE_NAMESPACE                    #
# RETRIES_ON_COMMAND_FAILURE           #
########################################

set +e

if [[ -z "${RELEASE_NAME}" ]]; then
    echo "Missing RELEASE_NAME env variable"
    exit 1
fi


if [[ -z "${RELEASE_NAMESPACE}" ]]; then
    echo "Missing RELEASE_NAMESPACE env variable"
    exit 1
fi


# Setup default value
MAX_RETRIES=3

if [[ -n "${RETRIES_ON_COMMAND_FAILURE}" ]]; then
    if [[ "${RETRIES_ON_COMMAND_FAILURE}" =~ ^[1-9][0-9]*$ ]]; then
        MAX_RETRIES=${RETRIES_ON_COMMAND_FAILURE}
        echo "RETRIES_ON_COMMAND_FAILURE configured to ${RETRIES_ON_COMMAND_FAILURE}"
    else
        echo "RETRIES_ON_COMMAND_FAILURE is not a number (value: \"${RETRIES_ON_COMMAND_FAILURE}\")"
        exit 1
    fi
fi

retry=0

#Find the release status.
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

#Delete the release (uninstall)
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

