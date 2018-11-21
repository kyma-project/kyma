#!/usr/bin/env bash

CURRENT_DIR="$( cd "$(dirname "$0")" ; pwd -P )"

eval $(minikube docker-env)

echo ""
echo "------------------------"
echo "Removing test pod"
echo "------------------------"


kubectl -n kyma-integration delete po connector-service-tests --now

echo ""
echo "------------------------"
echo "Building tests image"
echo "------------------------"

docker build $CURRENT_DIR/.. -t connector-service-tests

NODE_PORT=$(kubectl -n kyma-system get svc application-connector-nginx-ingress-controller -o 'jsonpath={.spec.ports[?(@.port==443)].nodePort}')

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
      value: https://gateway.kyma.local:$NODE_PORT
    - name: SKIP_SSL_VERIFY
      value: "true"
  restartPolicy: Never
EOF

echo ""
echo "------------------------"
echo "Waiting 5 seconds for pod to start..."
echo "------------------------"
echo ""

sleep 5

kubectl -n kyma-integration logs connector-service-tests -f
