#!/usr/bin/env bash

set -eu -o pipefail

readonly KNATIVE_BROKER_API_GROUP="brokers.eventing.knative.dev"
readonly ISTIO_POLICY_API_GROUP="policies.authentication.istio.io"

# this migration script ensures that Knative broker metrics are still scrapable
# Application-broker creates a Istio PeerAuthentication for newly provisioned ServiceInstances
# for existing ones, the following migration has to be performed:
# - iterate over all namespaces which have knative broker injection enabled => that means there is a broker in the namespace
# - iterate over each broker in the found namespaces (there should only be one, but just in case)
#   - delete the old Istio Policy
#   - create a new Istio PeerAuthentication
function migrate() {
     local namespaces=$(kubectl get ns --no-headers -l knative-eventing-injection | awk '{print $1}')
     echo "found the following namespaces with broker injection enabled: \"${namespaces}\""
     for namespace in ${namespaces}; do
        local brokers=$(kubectl get ${KNATIVE_BROKER_API_GROUP} --no-headers --namespace ${namespace} | awk '{print $1}')
        echo "found the following Knative Brokers in namespace: \"${namespace}\”: \”${brokers}\""
        for broker in ${brokers}; do
            echo "migration of broker ${broker}/${namespace} [started]"
            delete_istio_policy "${broker}" "${namespace}"
            create_istio_peer_authentication "${namespace}"
            echo "migration of broker ${broker}/${namespace} [done]"
        done
     done
}

# creates an Istio PeerAuthentication in the the given namespace
function create_istio_peer_authentication() {
    local namespace="${1}"
    # the peerauthentication uses the namespace name + the broker suffix, see here: https://github.com/kyma-project/kyma/blob/d4be1327717a6737177b64cd730467eb17982213/components/application-broker/internal/broker/provision.go#L447
    local name="${namespace}-broker"
    kubectl apply -f - <<EOF
apiVersion: security.istio.io/v1beta1
kind: PeerAuthentication
metadata:
  annotations:
  labels:
    eventing.knative.dev/broker: default
  name: ${name}
  namespace: ${namespace}
spec:
  portLevelMtls:
    "9090":
      mode: PERMISSIVE
  selector:
    matchLabels:
      eventing.knative.dev/broker: default
EOF
}

# deletes an Istio Policy with in the given namespace
function delete_istio_policy() {
    local namespace="${1}"
    # the peerauthentication uses the namespace name + the broker suffix, see here: https://github.com/kyma-project/kyma/blob/d4be1327717a6737177b64cd730467eb17982213/components/application-broker/internal/broker/provision.go#L447
    local name="${namespace}-broker"
    echo "deleting istio policy ${name}/${namespace} [started]"
    # delete the policy if it exists, otherwise there is nothing todo (--ignore-not-found prevents the command from failing in this case)
    kubectl delete "${ISTIO_POLICY_API_GROUP}" --ignore-not-found --namespace "${namespace}" "${name}"
    echo "deleting istio policy ${name}/${namespace} [done]"
}

echo "migration [started]"
migrate
echo "migration [done]"