#!/usr/bin/env bash

CURRENT_DIR="$( cd "$(dirname "$0")" ; pwd -P )"


if [[ -z "${TEST_IMAGE}" ]]; then
  echo "TEST_IMAGE env is not set. It should be set to full path of and image including tag, ex: mydockerhub/connector-service-tests:0.0.1"
  exit 1
fi

if [[ -z "${DOMAIN}" ]]; then
  echo "DOMAIN_NAME env is not set. It should be set to cluster domain name, ex: nightly.cluster.kyma.cx"
  exit 1
fi

echo "Current cluster context: $(kubectl config current-context)"
echo "Image: $TEST_IMAGE"
echo "Domain: $DOMAIN"

source $CURRENT_DIR/test-runner.sh

deleteTestPod

buildImage $TEST_IMAGE

echo ""
echo "------------------------"
echo "Pushing tests image"
echo "------------------------"

docker push $TEST_IMAGE

echo ""
echo "------------------------"
echo "Creating test pod"
echo "------------------------"

cat <<EOF | kubectl -n kyma-integration apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: connector-service-tests
  annotations:
    sidecar.istio.io/inject: "false"
spec:
  containers:
  - name: connector-service-tests
    image: $TEST_IMAGE
    imagePullPolicy: Always
    env:
    - name: INTERNAL_API_URL
      value: http://connector-service-internal-api:8080
    - name: EXTERNAL_API_URL
      value: http://connector-service-external-api:8081
    - name: GATEWAY_URL
      value: https://gateway.$DOMAIN
    - name: SKIP_SSL_VERIFY
      value: "true"
    - name: CENTRAL
      value: "false"
  restartPolicy: Never
EOF

waitForTestLogs 10
