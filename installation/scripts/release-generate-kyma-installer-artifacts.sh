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
INSTALLER_YAML_PATH="${RESOURCES_DIR}/installer.yaml"
INSTALLER_LOCAL_CONFIG_PATH="${RESOURCES_DIR}/installer-config-local.yaml.tpl"
INSTALLER_CLUSTER_CONFIG_PATH="${RESOURCES_DIR}/installer-config-cluster.yaml.tpl"
INSTALLER_LOCAL_CR_PATH="${RESOURCES_DIR}/installer-cr.yaml.tpl"
INSTALLER_CLUSTER_CR_PATH="${RESOURCES_DIR}/installer-cr-cluster.yaml.tpl"

TMP_LOCAL_CR=$(mktemp)
TMP_CLUSTER_CR=$(mktemp)

bash ${CURRENT_DIR}/create-cr.sh --url "" --output "${TMP_LOCAL_CR}" --version 0.0.1 --crtpl_path "${INSTALLER_LOCAL_CR_PATH}"
bash ${CURRENT_DIR}/create-cr.sh --url "" --output "${TMP_CLUSTER_CR}" --version 0.0.1 --crtpl_path "${INSTALLER_CLUSTER_CR_PATH}"

${CURRENT_DIR}/concat-yamls.sh ${INSTALLER_YAML_PATH} ${INSTALLER_LOCAL_CONFIG_PATH} ${TMP_LOCAL_CR} \
  | sed -E ";s;image: eu.gcr.io\/kyma-project\/develop\/installer:.+;image: eu.gcr.io/kyma-project/${KYMA_INSTALLER_PUSH_DIR}kyma-installer:${KYMA_INSTALLER_VERSION};" \
  > ${ARTIFACTS_DIR}/kyma-config-local.yaml

${CURRENT_DIR}/concat-yamls.sh ${INSTALLER_YAML_PATH} ${INSTALLER_CLUSTER_CONFIG_PATH} ${TMP_CLUSTER_CR} \
  | sed -E ";s;image: eu.gcr.io\/kyma-project\/develop\/installer:.+;image: eu.gcr.io/kyma-project/${KYMA_INSTALLER_PUSH_DIR}kyma-installer:${KYMA_INSTALLER_VERSION};" \
  > ${ARTIFACTS_DIR}/kyma-config-cluster.yaml

rm -rf ${TMP_LOCAL_CR}
rm -rf ${TMP_CLUSTER_CR}
