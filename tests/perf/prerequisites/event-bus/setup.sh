#!/usr/bin/env bash

set -e
set -o pipefail

WORKING_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

export NAMESPACE=event-bus-perf-test
export EB_CLUSTER_DOMAIN=$(kubectl get gateways.networking.istio.io kyma-gateway \
                        -n kyma-system -ojsonpath="{.spec.servers[0].hosts[0]}" | sed 's/*//g' )
export SUBSCRIBER_SERVICE_PATH="subscriber-service"
export SUBSCRIBER_STATUS_URL="https://${SUBSCRIBER_SERVICE_PATH}${EB_CLUSTER_DOMAIN}/v1/status"

eventing_specs=(
    namespace.yaml
    event-subscriber.yaml
    event-activation.yaml
    event-subscription.yaml
)

function wait_for_subscriber_to_be_ready() {
    echo -n "Waiting for subscriber to be ready..."
    for i in {1..15}; do #timeout after 30 seconds
        local response=$(curl -i $SUBSCRIBER_STATUS_URL 2>/dev/null)
        if [[ $response == *"200"* ]]; then
            echo -e "\nSubscriber is ready."
            return 0
        fi
        echo -n "."
        sleep 2
    done
    echo -e "\n\n ERROR: Subscriber pod is not ready to listen to events\n"
    exit 1
}

function wait_for_event_activation() {
    echo -e "\nWaiting for Event to be activated...\n"
    for i in {1..5}; do
        local kubectl_result=$(kubectl get eventactivations.applicationconnector.kyma-project.io \
                               perf-test-event-activation -n $NAMESPACE 2>/dev/null)
        if [[ $kubectl_result != "" ]]; then
            echo -e "Event Activated."
            return 0
        fi
        echo -n "."
        sleep 2
    done
    echo -e "\n\n ERROR: There was some problem with event-activation\n"
    exit 1
}

function wait_for_subscription_to_be_ready() {
    echo -e "\nChecking for subscription status...\n"

    for i in {1..5}; do
        local kubectl_result=$(kubectl get subscriptions.eventing.kyma-project.io \
                            hello-with-data-subscription \
                                -n $NAMESPACE -oyaml | grep "type: is-ready" 2>/dev/null)
        if [[ $kubectl_result == *"type: is-ready"* ]]; then
            echo -e "Subscription status is ready."
            return 0
        fi
        echo -n "."
        sleep 2
    done
    echo -e "\n\n ERROR: There was some problem in creating kyma subscription.\n"
    exit 1
}

for specs in "${eventing_specs[@]}"; do
    echo "Deploying spec: $specs"
    envsubst <"${WORKING_DIR}/$specs" | kubectl -n "$NAMESPACE" apply -f -
done

wait_for_subscriber_to_be_ready
wait_for_event_activation
wait_for_subscription_to_be_ready

