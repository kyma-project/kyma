set -ex

OPERATOR_FILE="/etc/istio/config.yaml"
echo "--> Check overrides"
if [ -f "/etc/istio/overrides.yaml" ]; then
	yq merge -x /etc/istio/config.yaml /etc/istio/overrides.yaml > /etc/combo.yaml
	OPERATOR_FILE="/etc/combo.yaml"
fi

echo "--> Temporary disable ingress-gateway"
kubectl scale deploy -n istio-system istio-ingressgateway --replicas 0

echo "--> Install Istio 1.7"
istioctl upgrade -f /etc/istio/operator-1-7.yaml -y

echo "Apply custom kyma manifests"
kubectl apply -f /etc/manifests

echo "Apply Kyma related checks and patches"
kubectl patch MutatingWebhookConfiguration istio-sidecar-injector --type 'json' -p '[{"op":"add","path":"/webhooks/0/namespaceSelector/matchExpressions/-","value":{"key":"gardener.cloud/purpose","operator":"NotIn","values":["kube-system"]}}]'

echo "--> Enable ingress-gateway"
kubectl scale deploy -n istio-system istio-ingressgateway --replicas 1