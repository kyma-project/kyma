#!/usr/bin/env bash

echo "Copying secrets to kyma-system namespace"
kubectl get secrets -n kyma-integration -o yaml | sed 's/namespace: kyma-integration/namespace: kyma-system/' | kubectl apply -f -

echo "Deleting secrets from kyma-integration namespace"
kubectl delete --all secrets -n kyma-integration