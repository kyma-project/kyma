#!/usr/bin/env bash

function deleteTestPod {
    echo ""
    echo "------------------------"
    echo "Removing test pod"
    echo "------------------------"

    kubectl -n kyma-integration delete po connector-service-tests --now
}

function waitForTestLogs {
    echo ""
    echo "------------------------"
    echo "Waiting $1 seconds for pod to start..."
    echo "------------------------"
    echo ""

    sleep $1

    kubectl -n kyma-integration logs connector-service-tests -f -c connector-service-tests
}

function buildImage {
    echo ""
    echo "------------------------"
    echo "Building tests image"
    echo "------------------------"

    docker build $CURRENT_DIR/.. -t $1
}