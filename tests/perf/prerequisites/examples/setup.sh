#!/usr/bin/env bash

WORKING_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

kubectl apply -f ${WORKING_DIR}/example.yaml