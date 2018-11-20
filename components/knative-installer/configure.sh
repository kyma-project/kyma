#!/usr/bin/env bash

set -e
set -x

KYMA_GW=$(kubectl get gateway -n kyma-system kyma-gateway -o json)

# Prefix port names with "kyma-" to avoid port name duplicates
# Read more: https://istio.io/help/ops/traffic-management/deploy-guidelines/#configuring-multiple-tls-hosts-in-a-gateway
KYMA_GW=$(jq '
    .spec.servers = (
        .spec.servers | map(
            if (.port.name | startswith("http"))
            then .port.name = "kyma_" + .port.name
            else .
            end
        )
    ) |
    .spec.selector = {"knative": "ingressgateway"}
' <<<"$KYMA_GW")

kubectl replace -f - <<<"$KYMA_GW"

if [[ -n "$IS_LOCAL_ENV" ]]; then
    # Disable hostPorts on istio-ingressgateway in local environment
    ISTIO_INGRESSGW=$(kubectl get deployment -n istio-system istio-ingressgateway -o json)
    ISTIO_INGRESSGW=$(jq '
        .spec.template.spec.containers[0].ports = (
            .spec.template.spec.containers[0].ports | map(
                del(.hostPort)
            )
        ) |
        del(.status)
    ' <<<"$ISTIO_INGRESSGW")

    kubectl replace -f - <<<"$ISTIO_INGRESSGW"

    # Enable hostPorts on knative-ingressgateway in local environment
    KNATIVE_INGRESSGW=$(kubectl get deployment -n istio-system knative-ingressgateway -o json)
    KNATIVE_INGRESSGW=$(jq '
        .spec.template.spec.containers[0].ports = (
            .spec.template.spec.containers[0].ports | map(
                if (.containerPort == 80 or .containerPort == 443)
                then .hostPort = .containerPort
                else .
                end
            )
        ) |
        del(.status)
    ' <<<"$KNATIVE_INGRESSGW")

    kubectl replace -f - <<<"$KNATIVE_INGRESSGW"
fi