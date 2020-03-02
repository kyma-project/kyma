#!/usr/bin/env bash

CURRENT_DIR="$( cd "$(dirname "$0")" ; pwd -P )"

APP_NAME=application-gateway-tests

for var in DOCKER_TAG DOCKER_PUSH_REPOSITORY; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

echo ""
echo "------------------------"
echo "Building tests image"
echo "------------------------"

IMAGE=$DOCKER_PUSH_REPOSITORY/$APP_NAME:$DOCKER_TAG

docker build ${CURRENT_DIR}/.. -t ${IMAGE}
docker push ${IMAGE}

echo ""
echo "------------------------"
echo "Removing Cluster Test Suite"
echo "------------------------"

kubectl delete cts gateway-tests

echo ""
echo "------------------------"
echo "Updating Test definition"
echo "------------------------"

kubectl -n kyma-integration patch testdefinition application-operator-gateway \
--patch "- op: replace
  path: /spec/template/spec/containers/0/image
  value: $IMAGE" --type=json

kubectl -n kyma-integration patch testdefinition application-operator-gateway \
--patch "- op: replace
  path: /spec/template/spec/containers/0/imagePullPolicy
  value: Always" --type=json

echo ""
echo "------------------------"
echo "Starting tests"
echo "------------------------"

cat <<EOF | kubectl apply -f -
apiVersion: testing.kyma-project.io/v1alpha1
kind: ClusterTestSuite
metadata:
  labels:
    controller-tools.k8s.io: "1.0"
  name: gateway-tests
spec:
  maxRetries: 1
  concurrency: 2
  selectors:
    matchNames:
    - name: application-operator-gateway
      namespace: kyma-integration
EOF

echo ""
echo "------------------------"
echo "Waiting for test pod to start..."
echo "------------------------"

sleep 20

kubectl -n kyma-integration logs oct-tp-gateway-tests-application-operator-gateway-0 -c tests -f
