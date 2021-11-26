set -ex
set -o pipefail

OPERATOR_FILE="/opt/istio/config/operator.yaml"

echo "--> Check overrides"
if [ -f "/opt/istio/config/overrides.yaml" ]; then
    yq merge -x "${OPERATOR_FILE}" /opt/istio/config/overrides.yaml > /opt/istio/combo.yaml

    CM_PRESENT=$(kubectl get cm -n "${NAMESPACE}" "${CONFIGMAP_NAME}" --ignore-not-found)
    if [[ -z "${CM_PRESENT}" ]]; then
        kubectl create cm "${CONFIGMAP_NAME}" -n "${NAMESPACE}" \
            --from-file "${OPERATOR_FILE}" \
            --from-file /opt/istio/config/overrides.yaml \
            --from-file /opt/istio/combo.yaml
    else
        kubectl create cm "${CONFIGMAP_NAME}" -n "${NAMESPACE}" \
            --from-file "${OPERATOR_FILE}" \
            --from-file /opt/istio/config/overrides.yaml \
            --from-file /opt/istio/combo.yaml \
            -o yaml --dry-run | kubectl replace -f -
    fi
    OPERATOR_FILE="/opt/istio/combo.yaml"
fi

echo "--> Install Istio"
istioctl install -f "${OPERATOR_FILE}" -y

echo "Apply custom kyma manifests"
kubectl apply -f /opt/istio/manifests

#This is still needed as mutating webhook disrupts Gardener cluster operations, like being able to hibernate the cluster. See https://github.com/kyma-project/kyma/issues/8868#issuecomment-658764987
echo "Apply Kyma related checks and patches"
kubectl patch MutatingWebhookConfiguration istio-sidecar-injector --type 'json' -p '[{"op":"add","path":"/webhooks/4/namespaceSelector/matchExpressions/-","value":{"key":"gardener.cloud/purpose","operator":"NotIn","values":["kube-system"]}}]'
