#!/usr/bin/env bash
echo "Deleting deprecated Application Connector resources"

kubectl -n kyma-integration delete deployment application-broker
kubectl -n kyma-integration delete service application-broker
kubectl -n kyma-integration delete configmap app-broker-config-map
kubectl -n kyma-integration delete authorizationpolicy application-connector-application-broker
kubectl delete clusterrolebinding application-broker
kubectl delete clusterrole application-broker
kubectl -n kyma-integration delete serviceaccount application-broker

kubectl -n kyma-integration delete statefulset application-operator
# Make sure that Application Operator is removed before triggering the Service Instances deletion
kubectl -n kyma-integration wait --for delete pod --selector=control-plane=application-operator
kubectl -n kyma-integration delete service application-operator-health
kubectl -n kyma-integration delete service application-operator-service
kubectl -n kyma-integration delete virtualservice application-operator
kubectl -n kyma-integration delete destinationrule application-operator-health-rule
kubectl delete podsecuritypolicy application-operator
kubectl delete clusterrolebinding application-operator
kubectl delete clusterrole application-operator
kubectl -n kyma-integration delete serviceaccount application-operator

kubectl delete serviceinstance --all-namespaces --all

echo "Deleting rafter"

kubectl delete crd assetgroups.rafter.kyma-project.io
kubectl delete crd assets.rafter.kyma-project.io
kubectl delete crd buckets.rafter.kyma-project.io
kubectl delete crd clusterassetgroups.rafter.kyma-project.io
kubectl delete crd clusterassets.rafter.kyma-project.io
kubectl delete crd clusterbuckets.rafter.kyma-project.io

kubectl -n kyma-system delete service rafter-asyncapi-service
kubectl -n kyma-system delete service rafter-controller-manager
kubectl -n kyma-system delete service rafter-front-matter-service
kubectl -n kyma-system delete service rafter-minio
kubectl -n kyma-system delete service rafter-upload-service

kubectl -n kyma-system delete deployment rafter-asyncapi-svc
kubectl -n kyma-system delete deployment rafter-ctrl-mngr
kubectl -n kyma-system delete deployment rafter-front-matter-svc
kubectl -n kyma-system delete deployment rafter-minio
kubectl -n kyma-system delete deployment rafter-upload-svc

echo "Deleting Service Catalog"

kubectl delete crd addonsconfigurations.addons.kyma-project.io
kubectl delete crd clusteraddonsconfigurations.addons.kyma-project.io
kubectl delete crd clusterservicebrokers.servicecatalog.k8s.io
kubectl delete crd clusterserviceclasses.servicecatalog.k8s.io
kubectl delete crd clusterserviceplans.servicecatalog.k8s.io
kubectl delete crd servicebindings.servicecatalog.k8s.io
kubectl delete crd servicebindingusages.servicecatalog.kyma-project.io
kubectl delete crd servicebrokers.servicecatalog.k8s.io
kubectl delete crd serviceclasses.servicecatalog.k8s.io
kubectl delete crd serviceinstances.servicecatalog.k8s.io
kubectl delete crd serviceplans.servicecatalog.k8s.io
kubectl delete crd usagekinds.servicecatalog.kyma-project.io
