#!/usr/bin/env bash
echo "Deleting Application Connectivity related resources"
kubectl delete job/application-connector-certs-setup-job -n kyma-integration
kubectl delete role,rolebinding -l app=application-connector-certs-setup-job -n kyma-integration
kubectl delete role/application-connector-certs-setup-job-ca-cert-role -n istio-system
kubectl delete rolebinding/application-connector-certs-setup-job-ca-cert-rolebinding -n istio-system
kubectl delete sa/application-connector-certs-setup-job -n kyma-integration

echo "Deleting Application Registry"
kubectl delete service,servicemonitor,role,rolebinding,deployment -l app=application-registry -n kyma-integration
kubectl delete clusterrole -l app=application-registry
kubectl delete peerauthentication/application-registry-policy -n kyma-integration
kubectl delete podsecuritypolicy/application-registry
kubectl delete cm/application-registry-dashboard -n kyma-system
kubectl delete svc application-registry-metrics -n kyma-integration

echo "Deleting Connector Service"
kubectl delete vs,svc,servicemonitor,role,rolebinding,configmap,deployment,authorizationpolicy -l app=connector-service -n kyma-integration
kubectl delete sa/connector-service -n kyma-integration
kubectl delete peerauthentication/connector-service-policy -n kyma-integration
kubectl delete podsecuritypolicy/connector-service
kubectl delete destinationrule/connector-service-external-api-rule -n kyma-integration
kubectl delete cm/connector-service-dashboard -n kyma-system
kubectl delete svc connector-service-metrics -n kyma-integration

echo "Deleting Connection Token Handler"
kubectl delete sa,deployment -l app=connection-token-handler -n kyma-integration
kubectl delete clusterrole,clusterrolebinding -l app=connection-token-handler
kubectl delete podsecuritypolicy/connection-token-handler

for NAMESPACE in $(kubectl get tokenrequest -A -o=custom-columns=NS:.metadata.namespace | uniq | tail -n +2)
do
  kubectl delete tokenrequest --all -n "$NAMESPACE"
done

kubectl delete crd tokenrequests.applicationconnector.kyma-project.io
