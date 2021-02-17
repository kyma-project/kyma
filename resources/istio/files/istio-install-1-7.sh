set -ex

OPERATOR_FILE="/etc/istio/operator-1-7.yaml"

echo "--> Check overrides"
if [ -f "/etc/istio/overrides.yaml" ]; then
    yq merge -x "${OPERATOR_FILE}" /etc/istio/overrides.yaml > /etc/combo.yaml
    kubectl create cm "${CONFIGMAP_NAME}" -n "${NAMESPACE}" \
        --from-file "${OPERATOR_FILE}" \
        --from-file /etc/istio/overrides.yaml \
        --from-file /etc/combo.yaml
    OPERATOR_FILE="/etc/combo.yaml"
fi

echo "--> Install Istio 1.7"
istioctl install -f "${OPERATOR_FILE}" -y

echo "Apply custom kyma manifests"
kubectl apply -f /etc/manifests

echo "Apply Kyma related checks and patches"
kubectl patch MutatingWebhookConfiguration istio-sidecar-injector --type 'json' -p '[{"op":"add","path":"/webhooks/0/namespaceSelector/matchExpressions/-","value":{"key":"gardener.cloud/purpose","operator":"NotIn","values":["kube-system"]}}]'
