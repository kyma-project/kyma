#!/usr/bin/env bash

kubectl delete -n kyma-system authorizationpolicies.security.istio.io kiali
kubectl delete -n kyma-system clusterroles.rbac.authorization.k8s.io kiali
kubectl delete -n kyma-system clusterroles.rbac.authorization.k8s.io kiali-viewer
kubectl delete -n kyma-system clusterrolebindings.rbac.authorization.k8s.io kiali
kubectl delete -n kyma-system configmaps kiali
kubectl delete -n kyma-system configmaps kiali-auth-proxy
kubectl delete -n kyma-system deployments.apps kiali
kubectl delete -n kyma-system deployments.apps kiali-auth-proxy
kubectl delete -n kyma-system peerauthentications.security.istio.io kiali
kubectl delete -n kyma-system roles.rbac.authorization.k8s.io kiali-controlplane
kubectl delete -n kyma-system rolebindings.rbac.authorization.k8s.io kiali-controlplane
kubectl delete -n kyma-system secrets kiali-auth-proxy-default
kubectl delete -n kyma-system services kiali
kubectl delete -n kyma-system services kiali-secured
kubectl delete -n kyma-system serviceaccounts kiali
kubectl delete -n kyma-system serviceaccounts kiali-auth-proxy
kubectl delete -n kyma-system servicemonitors.monitoring.coreos.com kiali
kubectl delete -n kyma-system virtualservices.networking.istio.io kiali
