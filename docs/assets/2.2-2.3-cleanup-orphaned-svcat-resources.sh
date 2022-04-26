#!/usr/bin/env bash

echo "
Deleting application broker"

echo "Framefrogs please contribute"
exit 1

echo "
Deleting rafter"

kubectl -n kyma-system delete crd assetgroups.rafter.kyma-project.io
kubectl -n kyma-system delete crd assets.rafter.kyma-project.io
kubectl -n kyma-system delete crd buckets.rafter.kyma-project.io
kubectl -n kyma-system delete crd clusterassetgroups.rafter.kyma-project.io
kubectl -n kyma-system delete crd clusterassets.rafter.kyma-project.io
kubectl -n kyma-system delete crd clusterbuckets.rafter.kyma-project.io

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

echo "
Deleting Service Catalog"

kubectl delete crd addonsconfigurations.addons.kyma-project.io
kubectl delete crd clusteraddonsconfigurations.addons.kyma-project.io
kubectl delete crd podpresets.settings.svcat.k8s.io
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