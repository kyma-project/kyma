#!/usr/bin/env bash

echo "Deleting orphaned Kiali resources (renamed)"

kubectl delete clusterroles.rbac.authorization.k8s.io kiali-server
kubectl delete clusterroles.rbac.authorization.k8s.io kiali-server-viewer
kubectl delete clusterrolebindings.rbac.authorization.k8s.io kiali-server
kubectl delete -n kyma-system configmap kiali-server
kubectl delete -n kyma-system deployments.apps kiali-server
kubectl delete -n kyma-system service kiali-server
kubectl delete -n kyma-system serviceaccount kiali-server
