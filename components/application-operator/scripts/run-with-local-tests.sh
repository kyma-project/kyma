#!/usr/bin/env bash

CURRENT_DIR="$( cd "$(dirname "$0")" ; pwd -P )"

eval $(minikube docker-env)

echo ""
echo "------------------------"
echo "Building component image"
echo "------------------------"

docker build $CURRENT_DIR/.. -t application-operator

echo ""
echo "------------------------"
echo "Updating stateful set"
echo "------------------------"

kubectl -n kyma-integration patch statefulset application-operator --patch 'spec:
  template:
    spec:
      containers:
      - name: application-operator
        image: application-operator
        imagePullPolicy: Never'

echo ""
echo "------------------------"
echo "Removing old pods"
echo "------------------------"

kubectl -n kyma-integration delete po -l control-plane=application-operator --now --wait=false

$CURRENT_DIR/../../../tests/application-operator-tests/scripts/run-local-tests.sh
