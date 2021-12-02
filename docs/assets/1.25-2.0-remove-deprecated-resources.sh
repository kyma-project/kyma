#!/usr/bin/env bash

echo "
Deleting orphaned Istio VirtualServices"

kubectl -n kyma-system delete virtualservice logging-loki
kubectl -n kyma-system delete virtualservice logging-log-ui
kubectl -n kyma-system delete virtualservice helm-broker-addons-ui
kubectl -n kyma-system delete virtualservice service-catalog-addons-service-catalog-ui-catalog

echo "
Deleting orphaned Istio DestinationRules"
kubectl -n kyma-system delete destinationrule helm-broker-addons-ui
kubectl -n kyma-system delete destinationrule logging-log-ui
kubectl -n kyma-system delete destinationrule service-catalog-addons-service-catalog-ui

echo "
Deleting orphaned Services"
kubectl -n kyma-system delete service service-catalog-addons-service-catalog-ui
kubectl -n kyma-system delete service logging-log-ui
kubectl -n kyma-system delete service helm-broker-addons-ui

echo "
Deleting orphaned Deployments"
kubectl -n kyma-system delete deployment logging-log-ui
kubectl -n kyma-system delete deployment helm-broker-addons-ui 
kubectl -n kyma-system delete deployment service-catalog-addons-service-catalog-ui

echo "
Deleting orphaned ServiceAccounts"
kubectl -n kyma-system delete serviceaccount logging-log-ui
kubectl -n kyma-system delete serviceaccount helm-broker-addons-ui
kubectl -n kyma-system delete serviceaccount service-catalog-addons-service-catalog-ui
kubectl -n kyma-system delete serviceaccount service-catalog-tests
kubectl -n kyma-system delete serviceaccount api-gateway-tests
kubectl -n kyma-system delete serviceaccount apiserver-proxy-certs-job
kubectl -n kyma-system delete serviceaccount apiserver-proxy-ssl-helper-service-account
kubectl -n kyma-system delete serviceaccount cluster-users-tests
kubectl -n kyma-system delete serviceaccount console-web-tests
kubectl -n kyma-system delete serviceaccount logging-tests
kubectl -n kyma-system delete serviceaccount ory-mechanism-migration
kubectl -n kyma-system delete serviceaccount ory-oathkeeper-keys-helper-service-account
kubectl -n kyma-system delete serviceaccount serverless-tests

echo "
Deleting orphaned Roles"
kubectl -n kyma-system delete role logging-log-ui
kubectl -n kyma-system delete role helm-broker-addons-ui
kubectl -n kyma-system delete role service-catalog-addons-service-catalog-ui
kubectl -n kyma-system delete role apiserver-proxy-certs-job
kubectl -n kyma-system delete role apiserver-proxy-certs-job-gardener-certs-role
kubectl -n kyma-system delete role apiserver-proxy-ssl-helper-role
kubectl -n kyma-system delete role compass-runtime-agent-tests-dex-secrets
kubectl -n kyma-system delete role ory-mechanism-migration
kubectl -n kyma-system delete role ory-oathkeeper-keys-helper-job-role

echo "
Deleting orphaned RoleBindings"
kubectl -n kyma-system delete rolebinding logging-log-ui
kubectl -n kyma-system delete rolebinding helm-broker-addons-ui
kubectl -n kyma-system delete rolebinding service-catalog-addons-service-catalog-ui
kubectl -n kyma-system delete rolebinding apiserver-proxy-certs-job
kubectl -n kyma-system delete rolebinding apiserver-proxy-certs-job-gardener-certs-role
kubectl -n kyma-system delete rolebinding apiserver-proxy-ssl-helper-role-binding
kubectl -n kyma-system delete rolebinding compass-runtime-agent-tests-dex-secrets
kubectl -n kyma-system delete rolebinding ory-mechanism-migration
kubectl -n kyma-system delete rolebinding ory-oathkeeper-keys-helper-job-role-binding

echo "
Deleting orphaned Secrets"
kubectl -n kyma-system delete secret test-developer-user
kubectl -n kyma-system delete secret test-namespace-admin-user
kubectl -n kyma-system delete secret test-no-rights-user
kubectl -n kyma-system delete secret test-read-only-user
kubectl -n kyma-system delete secret admin-user
kubectl -n kyma-system delete secret apiserver-proxy-tls-cert

echo "
Deleting orphaned ConfigMaps"
kubectl -n kyma-system delete configmap addons-ui
kubectl -n kyma-system delete configmap logging-log-ui
kubectl -n kyma-system delete configmap service-catalog-ui
kubectl -n kyma-system delete configmap apiserver-proxy
kubectl -n kyma-system delete configmap cluster-essentials-crd-0
kubectl -n kyma-system delete configmap cluster-essentials-crd-1
kubectl -n kyma-system delete configmap cluster-essentials-crd-2
kubectl -n kyma-system delete configmap cluster-users

echo "
Deleting orphaned PriorityClass"
kubectl delete priorityclass kyma-installer

echo "
Deleting orphaned ClusterRole"
kubectl delete clusterrole service-catalog-tests
kubectl delete clusterrole serverless-tests
kubectl delete clusterrole monitoring-tests
kubectl delete clusterrole logging-tests
kubectl delete clusterrole kyma-namespace-admin-essentials
kubectl delete clusterrole kyma-namespace-create
kubectl delete clusterrole kyma-mf-view
kubectl delete clusterrole kyma-mf-admin
kubectl delete clusterrole kyma-addons-admin
kubectl delete clusterrole kyma-addons-edit
kubectl delete clusterrole kyma-api-ns-admin
kubectl delete clusterrole kyma-admin
kubectl delete clusterrole api-gateway-tests
kubectl delete clusterrole application-operator-gateway-tests
kubectl delete clusterrole application-operator-tests
kubectl delete clusterrole application-connector-tests
kubectl delete clusterrole dex-admin
kubectl delete clusterrole dex-edit
kubectl delete clusterrole dex-view

echo "
Deleting orphaned ClusterRoleBinding"
kubectl delete clusterrolebinding api-gateway-tests
kubectl delete clusterrolebinding application-operator-gateway-tests
kubectl delete clusterrolebinding application-operator-tests
kubectl delete clusterrolebinding application-connector-tests
kubectl delete clusterrolebinding cluster-users-tests
kubectl delete clusterrolebinding kyma-admin-binding
kubectl delete clusterrolebinding kyma-essentials-binding
kubectl delete clusterrolebinding kyma-installer
kubectl delete clusterrolebinding kyma-namespace-admin-essentials-binding
kubectl delete clusterrolebinding kyma-ns-label
kubectl delete clusterrolebinding kyma-view-binding
kubectl delete clusterrolebinding service-catalog-tests
kubectl delete clusterrolebinding serverless-tests
kubectl delete clusterrolebinding monitoring-tests
kubectl delete clusterrolebinding logging-tests

echo "
Deleting orphaned Kyma modules"
helm -n kyma-system uninstall apiserver-proxy
helm -n kyma-system uninstall console
helm -n kyma-system uninstall core
helm -n kyma-system uninstall dex 
helm -n kyma-system uninstall iam-kubeconfig-service 
helm -n kyma-system uninstall permission-controller 
helm -n kyma-installer uninstall xip-patch
helm -n kyma-system uninstall testing

echo "
Discarding Helm metadata for Kyma modules"
kubectl -n istio-system delete secret -l owner=helm,name=istio
kubectl -n kyma-integration delete secret -l owner=helm,name=application-connector
kubectl -n kyma-system delete secret -l owner=helm,name=api-gateway
kubectl -n kyma-system delete secret -l owner=helm,name=cluster-essentials
kubectl -n kyma-system delete secret -l owner=helm,name=cluster-users
kubectl -n kyma-system delete secret -l owner=helm,name=eventing
kubectl -n kyma-system delete secret -l owner=helm,name=helm-broker
kubectl -n kyma-system delete secret -l owner=helm,name=logging
kubectl -n kyma-system delete secret -l owner=helm,name=monitoring
kubectl -n kyma-system delete secret -l owner=helm,name=ory
kubectl -n kyma-system delete secret -l owner=helm,name=rafter
kubectl -n kyma-system delete secret -l owner=helm,name=serverless
kubectl -n kyma-system delete secret -l owner=helm,name=service-catalog-addons
kubectl -n kyma-system delete secret -l owner=helm,name=service-catalog
kubectl -n kyma-system delete secret -l owner=helm,name=service-manager-proxy

echo "
Deleting orphaned CRDs"
kubectl delete crd backendmodules.ui.kyma-project.io
kubectl delete crd clustermicrofrontends.ui.kyma-project.io
kubectl delete crd installations.installer.kyma-project.io
kubectl delete crd releases.release.kyma-project.io
kubectl delete crd clustertestsuites.testing.kyma-project.io
kubectl delete crd testdefinitions.testing.kyma-project.io
kubectl delete crd authcodes.dex.coreos.com
kubectl delete crd authrequests.dex.coreos.com
kubectl delete crd connectors.dex.coreos.com
kubectl delete crd devicerequests.dex.coreos.com
kubectl delete crd devicetokens.dex.coreos.com
kubectl delete crd oauth2clients.dex.coreos.com
kubectl delete crd offlinesessionses.dex.coreos.com
kubectl delete crd passwords.dex.coreos.com
kubectl delete crd refreshtokens.dex.coreos.com
kubectl delete crd signingkeies.dex.coreos.com

echo "
Deleting orphaned namespace"
kubectl delete namespace kyma-installer
