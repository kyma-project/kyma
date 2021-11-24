#!/usr/bin/env bash

helm uninstall apiserver-proxy -n kyma-system
helm uninstall core -n kyma-system
helm uninstall dex -n kyma-system
helm uninstall iam-kubeconfig-service -n kyma-system
helm uninstall permission-controller -n kyma-system
#helm uninstall uaa-activator
helm uninstall xip-patch -n kyma-installer

kubectl delete ns kyma-installer
kubectl delete installation kyma-installation
kubectl delete crd installations.installer.kyma-project.io
