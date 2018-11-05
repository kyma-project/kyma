#!/usr/bin/env bash

CURRENT_DIR="$( cd "$(dirname "$0")" ; pwd -P )"

eval $(minikube docker-env)

echo ""
echo "------------------------"
echo "Building component image"
echo "------------------------"

docker build $CURRENT_DIR/.. -t connector-service

echo ""
echo "------------------------"
echo "Updating deployment"
echo "------------------------"

kubectl -n kyma-integration patch deployment connector-service --patch 'spec:
  template:
    spec:
      containers:
      - name: connector-service
        image: connector-service
        imagePullPolicy: Never'

echo ""
echo "------------------------"
echo "Removing old pods"
echo "------------------------"

kubectl -n kyma-integration delete po -l app=connector-service --now --wait=false

$CURRENT_DIR/../../../tests/connector-service-tests/scripts/run-local-tests.sh
