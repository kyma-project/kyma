#!/bin/bash

set -o errexit

ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

COMPONENT_NAME="$1"
ENABLED="$2"
STATUS="disabled"
FILE_NAME="components.yaml"
FILE_PATH=$ROOT_PATH/../${FILE_NAME}

# Check if the provided value is a valid boolean
if [[ ! ("${ENABLED}" == "true" || "${ENABLED}" == "false") ]]; then
    echo "\"${ENABLED}\" is not a boolean. Please provide a boolean value!"
    exit 1
fi

# Set status
if [ "${ENABLED}" == "true" ]; then
    STATUS="enabled"
fi

# Create the components.yaml file if it does not exist
if [ ! -f "${FILE_PATH}" ]; then
    echo "Generating components.yaml file"
    touch "${FILE_PATH}"
fi

# Remove previous entry in case the provided key exists
if grep -Fq ${COMPONENT_NAME} ${FILE_PATH}; then
    sed -i '' '/'"${COMPONENT_NAME}"'/d' ${FILE_PATH}
fi

# Append the provided key and value to the file
cat >> "${FILE_PATH}" <<EOL
${COMPONENT_NAME}.enabled: ${ENABLED}
EOL

echo "Component ${COMPONENT_NAME} is now ${STATUS}!"