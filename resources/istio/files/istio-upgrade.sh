set -ex
set -o pipefail

OPERATOR_FILE="/etc/istio/operator.yaml"

echo "--> Check overrides"
if [ -f "/etc/istio/overrides.yaml" ]; then
    yq merge -x "${OPERATOR_FILE}" /etc/istio/overrides.yaml > /etc/combo.yaml

    CM_PRESENT=$(kubectl get cm -n "${NAMESPACE}" "${CONFIGMAP_NAME}" --ignore-not-found)
    if [[ -z "${CM_PRESENT}" ]]; then
        kubectl create cm "${CONFIGMAP_NAME}" -n "${NAMESPACE}" \
            --from-file "${OPERATOR_FILE}" \
            --from-file /etc/istio/overrides.yaml \
            --from-file /etc/combo.yaml
    else
        kubectl create cm "${CONFIGMAP_NAME}" -n "${NAMESPACE}" \
            --from-file "${OPERATOR_FILE}" \
            --from-file /etc/istio/overrides.yaml \
            --from-file /etc/combo.yaml \
            -o yaml --dry-run | kubectl replace -f -
    fi
    OPERATOR_FILE="/etc/combo.yaml"
fi

echo "--> Install Istio"
istioctl upgrade -f "${OPERATOR_FILE}" -y

echo "Apply custom kyma manifests"
kubectl apply -f /etc/manifests

echo "Apply Kyma related checks and patches"
kubectl patch MutatingWebhookConfiguration istio-sidecar-injector --type 'json' -p '[{"op":"add","path":"/webhooks/0/namespaceSelector/matchExpressions/-","value":{"key":"gardener.cloud/purpose","operator":"NotIn","values":["kube-system"]}}]'

kubeSystemSelector="$(kubectl get mutatingwebhookconfigurations.admissionregistration.k8s.io istio-sidecar-injector -o jsonpath='{.webhooks[0].namespaceSelector.matchExpressions[0]}' | { grep  "kube-system" || test $? == 1; })"

if [ -z "$kubeSystemSelector" ]
then
  echo "patching namespace selector of mutating webhook istio-sidecar-injector with kube-system"
	kubectl patch MutatingWebhookConfiguration istio-sidecar-injector --type 'json' -p '[{"op":"add","path":"/webhooks/0/namespaceSelector/matchExpressions/0/values/-","value": "kube-system"}]'
else
  echo "namespace selector for kube-system in of mutating webhook istio-sidecar-injector already exists, skipping operation"
fi