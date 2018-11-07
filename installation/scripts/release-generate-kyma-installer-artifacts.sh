#!/usr/bin/env bash

###
# Following script generates kyma-installer artifacts for a release.
# It produces two files: kyma-config-local.yaml and kyma-config-cluster.yaml
#
# INPUTS:
# - KYMA_INSTALLER_PUSH_DIR - (optional) directory where kyma-installer docker image is pushed, if specified should ends with a slash (/)
# - KYMA_INSTALLER_VERSION - version (image tag) of kyma-installer
# - ARTIFACTS_DIR - path to directory where artifacts will be stored
#
###

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
RESOURCES_DIR="${CURRENT_DIR}/../resources"

${CURRENT_DIR}/generate-kyma-installer.sh ${RESOURCES_DIR}/installer-config-local.yaml.tpl \
  | sed -E ";s;image: eu.gcr.io\/kyma-project\/develop\/installer:.+;image: eu.gcr.io/kyma-project${KYMA_INSTALLER_PUSH_DIR}/kyma-installer:${KYMA_INSTALLER_VERSION};" \
  > ${ARTIFACTS_DIR}/kyma-config-local.yaml

${CURRENT_DIR}/generate-kyma-installer.sh ${RESOURCES_DIR}/installer-config-cluster.yaml.tpl \
  | sed -E ";s;image: eu.gcr.io\/kyma-project\/develop\/installer:.+;image: eu.gcr.io/kyma-project${KYMA_INSTALLER_PUSH_DIR}/kyma-installer:${KYMA_INSTALLER_VERSION};" \
  > ${ARTIFACTS_DIR}/kyma-config-cluster.yaml
