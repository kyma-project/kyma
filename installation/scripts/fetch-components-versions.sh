#!/usr/bin/env bash
set -e

CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
VERSIONS_FILE_URL="https://kymainstaller.blob.core.windows.net/kyma-versions/latest.env"
VERSIONS_FILE_PATH="$( cd "$( dirname "${CURRENT_DIR}" )" && pwd )/versions.env"

if [ ! -f ${VERSIONS_FILE_PATH} ]; then
    echo -e "\nDownloading versions.env file."
    wget -O "${VERSIONS_FILE_PATH}" "${VERSIONS_FILE_URL}" 2>/dev/null || curl -o "${VERSIONS_FILE_PATH}" "${VERSIONS_FILE_URL}"
else
    echo -e "\nFile ${VERSIONS_FILE_PATH} exists, reusing."
fi