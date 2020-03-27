#!/usr/bin/env bash

set -euo pipefail
IFS=$'\n\t'

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
kubectl delete functions.serverless.kyma-project.io --all --all-namespaces
kubectl delete triggers.eventing.knative.dev --all --all-namespaces
kubectl delete apirules.gateway.kyma-project.io --all --all-namespaces
sleep 3
kubectl delete -f "${SCRIPT_DIR}/migrator-yamls/job.yaml" || echo "Job not present, continue..."
sleep 2
"${SCRIPT_DIR}/applyKubelessFns.sh"
docker build -t migrator -f migrator/Dockerfile . && kind load docker-image migrator --name migrator && kubectl apply -f "${SCRIPT_DIR}/migrator-yamls/job.yaml"