#!/usr/bin/env bash

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT_DIR="${CURRENT_DIR}/../.."
INSTALLER_YAML_PATH="${ROOT_DIR}/installation/resources/installer-local.yaml"

if [ -f $INSTALLER_YAML_PATH ]; then
    VERSION=$(grep "image: " $INSTALLER_YAML_PATH | grep "kyma-installer" | cut -d":" -f2-) 
    echo $VERSION
else
    echo "${INSTALLER_YAML_PATH} not found"
    exit 1
fi 
