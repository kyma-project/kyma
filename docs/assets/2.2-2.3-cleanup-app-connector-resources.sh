#!/usr/bin/env bash
echo "Deleting Application Connectivity related resources"
kubectl delete job/application-connector-certs-setup-job -n kyma-integration --ignore-not-found=true
kubectl delete role,rolebinding -l app=application-connector-certs-setup-job -n kyma-integration
kubectl delete role/application-connector-certs-setup-job-ca-cert-role -n istio-system --ignore-not-found=true
kubectl delete rolebinding/application-connector-certs-setup-job-ca-cert-rolebinding -n istio-system --ignore-not-found=true
kubectl delete sa/application-connector-certs-setup-job -n kyma-integration --ignore-not-found=true

echo "Deleting Application Registry"
kubectl delete service,servicemonitor,role,rolebinding,deployment -l app=application-registry -n kyma-integration
kubectl delete clusterrole -l app=application-registry
kubectl delete peerauthentication/application-registry-policy -n kyma-integration --ignore-not-found=true
kubectl delete podsecuritypolicy/application-registry --ignore-not-found=true
kubectl delete cm/application-registry-dashboard -n kyma-system --ignore-not-found=true
kubectl delete svc application-registry-metrics -n kyma-integration --ignore-not-found=true

echo "Deleting Connector Service"
kubectl delete vs,svc,servicemonitor,role,rolebinding,configmap,deployment,authorizationpolicy -l app=connector-service -n kyma-integration
kubectl delete sa/connector-service -n kyma-integration --ignore-not-found=true
kubectl delete peerauthentication/connector-service-policy -n kyma-integration --ignore-not-found=true
kubectl delete podsecuritypolicy/connector-service --ignore-not-found=true
kubectl delete destinationrule/connector-service-external-api-rule -n kyma-integration --ignore-not-found=true
kubectl delete cm/connector-service-dashboard -n kyma-system --ignore-not-found=true
kubectl delete svc connector-service-metrics -n kyma-integration --ignore-not-found=true

echo "Deleting Connection Token Handler"
kubectl delete sa,deployment -l app=connection-token-handler -n kyma-integration
kubectl delete clusterrole,clusterrolebinding -l app=connection-token-handler
kubectl delete podsecuritypolicy/connection-token-handler --ignore-not-found=true
kubectl delete tokenrequest --all -A
kubectl delete crd tokenrequests.applicationconnector.kyma-project.io
