set -ex
set -o pipefail

OPERATOR_FILE="/etc/istio/operator-1-7.yaml"

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

echo "--> Temporary disable ingress-gateway"
kubectl scale deploy -n istio-system istio-ingressgateway --replicas 0

echo "--> Install Istio 1.7"
istioctl upgrade -f "${OPERATOR_FILE}" -y

echo "Apply custom kyma manifests"
kubectl apply -f /etc/manifests

echo "Apply Kyma related checks and patches"
kubectl patch MutatingWebhookConfiguration istio-sidecar-injector --type 'json' -p '[{"op":"add","path":"/webhooks/0/namespaceSelector/matchExpressions/-","value":{"key":"gardener.cloud/purpose","operator":"NotIn","values":["kube-system"]}}]'

echo "--> Enable ingress-gateway"
kubectl scale deploy -n istio-system istio-ingressgateway --replicas 1