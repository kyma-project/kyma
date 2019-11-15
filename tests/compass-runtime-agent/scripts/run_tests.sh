#!/usr/bin/env bash

CURRENT_DIR="$( cd "$(dirname "$0")" ; pwd -P )"

APP_NAME=compass-runtime-agent-tests
NAMESPACE=compass-system

discoverUnsetVar=false

for var in DOCKER_TAG DOCKER_PUSH_REPOSITORY; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

set -e

echo ""
echo "------------------------"
echo "Removing old cluster test suite"
echo "------------------------"

set +e
kubectl -n $NAMESPACE delete cts $APP_NAME
set -e

echo ""
echo "------------------------"
echo "Building tests image"
echo "------------------------"

IMAGE=$DOCKER_PUSH_REPOSITORY/$APP_NAME:$DOCKER_TAG

make build-image
make push-image

echo ""
echo "------------------------"
echo "Updating test definition"
echo "------------------------"

kubectl -n $NAMESPACE patch td compass-runtime-agent-tests --patch "
- op: replace
  path: /spec/template/spec/containers/0/image
  value: $IMAGE
- op: replace
  path: /spec/template/spec/containers/0/imagePullPolicy
  value: Always
" --type=json

echo ""
echo "------------------------"
echo "Creating test suite"
echo "------------------------"

cat <<EOF | kubectl -n $NAMESPACE apply -f -
apiVersion: testing.kyma-project.io/v1alpha1
kind: ClusterTestSuite
metadata:
  labels:
    controller-tools.k8s.io: "1.0"
  name: $APP_NAME
spec:
  maxRetries: 0
  concurrency: 1
  selectors:
    matchNames:
    - name: $APP_NAME
      namespace: $NAMESPACE
EOF

echo ""
echo "------------------------"
echo "Waiting 15 seconds for pod to start..."
echo "------------------------"
echo ""

sleep 15

kubectl -n $NAMESPACE logs -l testing.kyma-project.io/def-name=$APP_NAME -c $APP_NAME -f
