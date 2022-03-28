#!/usr/bin/env bash
set -e

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