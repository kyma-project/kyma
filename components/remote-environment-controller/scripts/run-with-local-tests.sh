#!/usr/bin/env bash

CURRENT_DIR="$( cd "$(dirname "$0")" ; pwd -P )"

eval $(minikube docker-env)

echo ""
echo "------------------------"
echo "Building component image"
echo "------------------------"

docker build $CURRENT_DIR/.. -t remote-environment-controller

kubectl -n kyma-integration patch statefulset remote-environment-controller --patch 'spec:
  template:
    spec:
      containers:
      - name: remote-environment-controller
        image: remote-environment-controller
        imagePullPolicy: Never'

kubectl -n kyma-integration delete po -l control-plane=remote-environment-controller --now --wait=false

$CURRENT_DIR/../../../tests/remote-environment-controller-tests/scripts/run-local-tests.sh
