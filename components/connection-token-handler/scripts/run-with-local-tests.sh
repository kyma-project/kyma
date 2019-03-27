#!/usr/bin/env bash

CURRENT_DIR="$( cd "$(dirname "$0")" ; pwd -P )"

eval $(minikube docker-env)

echo ""
echo "------------------------"
echo "Building component image"
echo "------------------------"

docker build $CURRENT_DIR/.. -t connection-token-handler

echo ""
echo "------------------------"
echo "Updating deployment"
echo "------------------------"

kubectl -n kyma-integration patch deployment connection-token-handler --patch 'spec:
  template:
    spec:
      containers:
      - name: connection-token-handler
        image: connection-token-handler
        imagePullPolicy: Never'

echo ""
echo "------------------------"
echo "Removing old pods"
echo "------------------------"

kubectl -n kyma-integration delete po -l app=connection-token-handler --now --wait=false

$CURRENT_DIR/../../../tests/connection-token-handler-tests/scripts/run-local-tests.sh
