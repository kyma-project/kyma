#!/usr/bin/env bash

CURRENT_DIR="$( cd "$(dirname "$0")" ; pwd -P )"

APP_NAME=application-gateway-tests
#NAMESPACE=default

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
echo "Updating Image in Operator"
echo "------------------------"

kubectl -n kyma-integration patch statefulset application-operator --patch "
- op: replace
  path: /spec/template/spec/containers/0/args/9
  value: '--applicationGatewayTestsImage='$IMAGE
"

