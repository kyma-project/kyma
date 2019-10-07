#!/usr/bin/env bash

CURRENT_DIR="$( cd "$(dirname "$0")" ; pwd -P )"

source $CURRENT_DIR/test-runner.sh

deleteTestPod

eval $(minikube docker-env)

buildImage connector-service-tests

echo ""
echo "------------------------"
echo "Creating test pod"
echo "------------------------"

MINIKUBE_IP=$(minikube ip)

cat <<EOF | kubectl -n kyma-integration apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: connector-service-tests
  annotations:
    sidecar.istio.io/inject: "false"
spec:
  hostAliases:
  - ip: "$MINIKUBE_IP"
    hostnames:
    - "connector-service.kyma.local"
    - "gateway.kyma.local"
  containers:
  - name: connector-service-tests
    image: connector-service-tests
    imagePullPolicy: Never
    env:
    - name: INTERNAL_API_URL
      value: http://connector-service-internal-api:8080
    - name: EXTERNAL_API_URL
      value: http://connector-service-external-api:8081
    - name: GATEWAY_URL
      value: https://gateway.kyma.local
    - name: SKIP_SSL_VERIFY
      value: "true"
    - name: CENTRAL
      value: "false"
  restartPolicy: Never
EOF

waitForTestLogs 5
