#!/usr/bin/env bash

kubectl -n kyma-system patch deployments.apps telemetry-operator --type json -p='[{"op": "remove", "path": "/spec/template/spec/initContainers/0"}, {"op": "remove", "path": "/spec/template/spec/containers/0/volumeMounts/0"}, {"op": "remove", "path": "/spec/template/spec/volumes/0"}]' || true
