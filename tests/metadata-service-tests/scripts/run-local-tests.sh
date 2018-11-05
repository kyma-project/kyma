#!/usr/bin/env bash

CURRENT_DIR="$( cd "$(dirname "$0")" ; pwd -P )"

eval $(minikube docker-env)

echo ""
echo "------------------------"
echo "Removing test pod"
echo "------------------------"


kubectl -n kyma-integration delete po metadata-service-tests --now

echo ""
echo "------------------------"
echo "Building tests image"
echo "------------------------"

docker build $CURRENT_DIR/.. -t metadata-service-tests

echo ""
echo "------------------------"
echo "Creating test pod"
echo "------------------------"

cat <<EOF | kubectl -n kyma-integration apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: metadata-service-tests
spec:
  containers:
  - name: metadata-service-tests
    image: metadata-service-tests
    imagePullPolicy: Never
    env:
    - name: METADATA_URL
      value: http://metadata-service-external-api:8081
    - name: NAMESPACE
      value: kyma-integration
  restartPolicy: Never
EOF

echo ""
echo "------------------------"
echo "Waiting 5 seconds for pod to start..."
echo "------------------------"
echo ""

sleep 5

kubectl -n kyma-integration logs metadata-service-tests -f
