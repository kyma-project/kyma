#!/usr/bin/env bash

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
KYMA_PATH="${CURRENT_DIR}/../.."
INSTALLER_YAML_PATH="${KYMA_PATH}/installation/resources/installer.yaml"

if [ -f $INSTALLER_YAML_PATH ]; then
    VERSION=$(grep "image: " $INSTALLER_YAML_PATH | cut -f3 -d":") 
    echo $VERSION
else
    echo "${INSTALLER_YAML_PATH} not found"
    exit 1
fi 