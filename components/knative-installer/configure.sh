#!/usr/bin/env bash

KYMA_GW=$(kubectl get gateway -n kyma-system kyma-gateway -o json)

# Prefix port names with "kyma-" to avoid port name duplicates
# See more on: https://istio.io/help/ops/traffic-management/deploy-guidelines/#configuring-multiple-tls-hosts-in-a-gateway

KYMA_GW=$(jq '
    .spec.servers = (
        .spec.servers | map(
            if (.port.name | startswith("http"))
            then .port.name = "kyma_" + .port.name
            else .
            end
        )
    )
' <<<"$KYMA_GW")

kubectl apply -f - <<<"$KYMA_GW"

# Change ingressgateway to knative's
# Somehow I cannot get it to work with the above. istio selector still exists next to knative
kubectl patch gateway -n kyma-system kyma-gateway --type=json -p '[
    {
        "op": "replace",
        "path": "/spec/selector",
        "value": {
            "knative": "ingressgateway"
        }
    }
]'
