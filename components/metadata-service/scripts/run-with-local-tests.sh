#!/usr/bin/env bash

CURRENT_DIR="$( cd "$(dirname "$0")" ; pwd -P )"

eval $(minikube docker-env)

echo ""
echo "------------------------"
echo "Building component image"
echo "------------------------"

docker build $CURRENT_DIR/.. -t metadata-service

echo ""
echo "------------------------"
echo "Updating deployment"
echo "------------------------"

kubectl -n kyma-integration patch deployment metadata-service --patch 'spec:
  template:
    spec:
      containers:
      - name: metadata-service
        image: metadata-service
        imagePullPolicy: Never'

echo ""
echo "------------------------"
echo "Removing old pods"
echo "------------------------"

kubectl -n kyma-integration delete po -l app=metadata-service --now --wait=false

$CURRENT_DIR/../../../tests/metadata-service-tests/scripts/run-local-tests.sh
