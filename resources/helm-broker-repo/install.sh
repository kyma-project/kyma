#!/usr/bin/env bash

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

bash ${CURRENT_DIR}/provisioning/provision-bundles.sh ${CURRENT_DIR}/bundles
