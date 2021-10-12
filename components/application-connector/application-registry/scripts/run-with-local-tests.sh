#!/usr/bin/env bash

CURRENT_DIR="$( cd "$(dirname "$0")" ; pwd -P )"

eval $(minikube docker-env)

echo ""
echo "------------------------"
echo "Building component image"
echo "------------------------"

docker build $CURRENT_DIR/.. -t application-registry

echo ""
echo "------------------------"
echo "Updating deployment"
echo "------------------------"

kubectl -n kyma-integration patch deployment application-registry --patch 'spec:
  template:
    spec:
      containers:
      - name: application-registry
        image: application-registry
        imagePullPolicy: Never'

echo ""
echo "------------------------"
echo "Removing old pods"
echo "------------------------"

kubectl -n kyma-integration delete po -l app=application-registry --now --wait=false

$CURRENT_DIR/../../../tests/application-registry-tests/scripts/run-local-tests.sh
