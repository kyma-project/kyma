#!/usr/bin/env bash

###
# Following script generates kyma-installer artifacts for a release.
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
INSTALLER_LOCAL_CR_PATH="${RESOURCES_DIR}/installer-cr.yaml.tpl"
INSTALLER_CLUSTER_CR_PATH="${RESOURCES_DIR}/installer-cr-cluster.yaml.tpl"
INSTALLER_RUNTIME_CLUSTER_CR_PATH="${RESOURCES_DIR}/installer-cr-cluster-runtime.yaml.tpl"

function generateLocalArtifact() {
    TMP_LOCAL_CR=$(mktemp)

    ${CURRENT_DIR}/create-cr.sh --url "" --output "${TMP_LOCAL_CR}" --version 0.0.1 --crtpl_path "${INSTALLER_LOCAL_CR_PATH}"

    ${CURRENT_DIR}/concat-yamls.sh ${INSTALLER_YAML_PATH} ${TMP_LOCAL_CR} \
      | sed -E ";s;image: eu.gcr.io\/kyma-project\/develop\/installer:.+;image: eu.gcr.io/kyma-project/${KYMA_INSTALLER_PUSH_DIR}kyma-installer:${KYMA_INSTALLER_VERSION};" \
      > ${ARTIFACTS_DIR}/kyma-installer-local.yaml

    cp ${INSTALLER_LOCAL_CONFIG_PATH} ${ARTIFACTS_DIR}/kyma-config-local.yaml
    cp ${INSTALLER_LOCAL_CR_PATH}  ${ARTIFACTS_DIR}/kyma-installer-cr-local.yaml
    cp ${INSTALLER_YAML_PATH} ${ARTIFACTS_DIR}/kyma-installer.yaml

    rm -rf ${TMP_LOCAL_CR}
}

function generateClusterArtifact() {
    TMP_CLUSTER_CR=$(mktemp)

    ${CURRENT_DIR}/create-cr.sh --url "" --output "${TMP_CLUSTER_CR}" --version 0.0.1 --crtpl_path "${INSTALLER_CLUSTER_CR_PATH}"

    ${CURRENT_DIR}/concat-yamls.sh ${INSTALLER_YAML_PATH} ${TMP_CLUSTER_CR} \
      | sed -E ";s;image: eu.gcr.io\/kyma-project\/develop\/installer:.+;image: eu.gcr.io/kyma-project/${KYMA_INSTALLER_PUSH_DIR}kyma-installer:${KYMA_INSTALLER_VERSION};" \
      > ${ARTIFACTS_DIR}/kyma-installer-cluster.yaml

    cp ${INSTALLER_CLUSTER_CR_PATH}  ${ARTIFACTS_DIR}/kyma-installer-cr-cluster.yaml

    rm -rf ${TMP_CLUSTER_CR}
}

function generateRuntimeClusterArtifact() {
    TMP_RUNTIME_CLUSTER_CR=$(mktemp)

    ${CURRENT_DIR}/create-cr.sh --url "" --output "${TMP_RUNTIME_CLUSTER_CR}" --version 0.0.1 --crtpl_path "${INSTALLER_RUNTIME_CLUSTER_CR_PATH}"

    ${CURRENT_DIR}/concat-yamls.sh ${INSTALLER_YAML_PATH} ${TMP_RUNTIME_CLUSTER_CR} \
      | sed -E ";s;image: eu.gcr.io\/kyma-project\/develop\/installer:.+;image: eu.gcr.io/kyma-project/${KYMA_INSTALLER_PUSH_DIR}kyma-installer:${KYMA_INSTALLER_VERSION};" \
      > ${ARTIFACTS_DIR}/kyma-installer-cluster-runtime.yaml


    cp ${INSTALLER_RUNTIME_CLUSTER_CR_PATH}  ${ARTIFACTS_DIR}/kyma-installer-cr-cluster-runtime.yaml


    rm -rf ${TMP_RUNTIME_CLUSTER_CR}
}

generateLocalArtifact
generateClusterArtifact
generateRuntimeClusterArtifact
