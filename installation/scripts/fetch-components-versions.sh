#!/usr/bin/env bash
set -e

CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
VERSIONS_FILE_URL="https://kymainstaller.blob.core.windows.net/kyma-versions/latest.env"
VERSIONS_FILE_PATH="$( cd "$( dirname "${CURRENT_DIR}" )" && pwd )/versions-overrides.env"

POSITIONAL=()
while [[ $# -gt 0 ]]
do
    key="$1"

    case ${key} in
        --url)
          VERSIONS_FILE_URL="$2"
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

if [ ! -f ${VERSIONS_FILE_PATH} ]; then
    echo -e "\nDownloading versions-overrides.env file from ${VERSIONS_FILE_URL}."
    wget -O "${VERSIONS_FILE_PATH}" "${VERSIONS_FILE_URL}" 2>/dev/null || curl -o "${VERSIONS_FILE_PATH}" "${VERSIONS_FILE_URL}"
else
    echo -e "\nFile ${VERSIONS_FILE_PATH} exists, reusing."
fi