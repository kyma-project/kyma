#!/usr/bin/env bash

echo "Deleting secrets from kyma-integration namespace"
kubectl delete --all secrets -n kyma-integration