#!/usr/bin/env bash

CURRENT_DIR="$( cd "$(dirname "$0")" ; pwd -P )"

eval $(minikube docker-env)

echo ""
echo "------------------------"
echo "Removing test pod"
echo "------------------------"


kubectl -n kyma-integration delete po connection-token-handler-tests --now

echo ""
echo "------------------------"
echo "Building tests image"
echo "------------------------"

docker build $CURRENT_DIR/.. -t connection-token-handler-tests

echo ""
echo "------------------------"
echo "Creating test pod"
echo "------------------------"

cat <<EOF | kubectl -n kyma-integration apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: connection-token-handler-tests
  annotations:
    sidecar.istio.io/inject: “false”
spec:
  serviceAccountName: connection-token-handler-tests
  containers:
  - name: connection-token-handler-tests
    image: connection-token-handler-tests
    imagePullPolicy: Never
  restartPolicy: Never
EOF

echo ""
echo "------------------------"
echo "Waiting 5 seconds for pod to start..."
echo "------------------------"
echo ""

sleep 5

kubectl -n kyma-integration logs connection-token-handler-tests -f
