#!/usr/bin/env bash
######### Knative build & serving ########

echo "Installing Knative build and serving ..."

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

curl -L ${SERVING_URL} \
| sed 's/LoadBalancer/NodePort/' \
| tee knative-serving.yaml \
| kubectl apply -f -


echo "Verifying Knative build and serving installation..."
sleep 2
until kubectl get -f ${SERVING_URL} > /dev/null 2>&1
do
    echo "Knative CRDs not yet synced, re-applying..."
    kubectl apply -f knative-serving.yaml
    sleep 2
done

${DIR}/is-ready.sh knative-build app build-controller
${DIR}/is-ready.sh knative-build app build-webhook
${DIR}/is-ready.sh knative-serving app activator
${DIR}/is-ready.sh knative-serving app autoscaler
${DIR}/is-ready.sh knative-serving app controller
${DIR}/is-ready.sh knative-serving app webhook

# Enable TLS in knative gateway
KNATIVE_GW=$(kubectl get gateway -n knative-serving knative-shared-gateway -o json)
KNATIVE_GW=$(jq '
    .spec.servers = (
        .spec.servers | map (
            if .port.number == 443
            then (
                .tls.mode = "SIMPLE"
                | .tls.privateKey = "/etc/istio/ingressgateway-certs/tls.key"
                | .tls.serverCertificate = "/etc/istio/ingressgateway-certs/tls.crt"
            )
            else .
            end
        )
    ) |
    .spec.selector = {"knative": "ingressgateway"}
' <<<"$KNATIVE_GW")
kubectl replace -f - <<<"$KNATIVE_GW"

echo "Knative build and serving installation verified"

echo "Installing Knative eventing..."

kubectl apply -f ${EVENTING_URL}

echo "Verifying Knative eventing installation..."

${DIR}/is-ready.sh knative-eventing app eventing-controller
${DIR}/is-ready.sh knative-eventing app webhook

echo "Knative eventing installation verified"

