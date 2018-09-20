#!/usr/bin/env bash

NAMESPACE="istio-system"
CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

set -o errexit

bash ${CURRENT_DIR}/../../installation/scripts/is-ready.sh ${NAMESPACE} istio sidecar-injector
bash ${CURRENT_DIR}/provisioning/provision-bundles.sh ${CURRENT_DIR}/bundles