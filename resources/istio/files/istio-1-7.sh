set -ex

OPERATOR_FILE="/etc/istio/config.yaml"
echo "--> Check overrides"
if [ -f "/etc/istio/overrides.yaml" ]; then
	yq merge -x /etc/istio/config.yaml /etc/istio/overrides.yaml > /etc/combo.yaml
	OPERATOR_FILE="/etc/combo.yaml"
fi

echo "--> Remove current installation of Istio"
istioctl manifest generate -f "${OPERATOR_FILE}" | yq r -d '*' --prettyPrint -j - | jq 'select(.kind != "CustomResourceDefinition") | select(.kind != "Namespace")' | kubectl delete -f -

echo "--> Get Istio 1.7"
export ISTIOCTL_VERSION=1.7.1
curl -L https://github.com/istio/istio/releases/download/${ISTIOCTL_VERSION}/istioctl-${ISTIOCTL_VERSION}-linux-amd64.tar.gz -o istioctl.tar.gz &&\
tar xvzf istioctl.tar.gz
chmod +x istioctl
mv istioctl /usr/local/bin/istioctl-${ISTIOCTL_VERSION}
rm istioctl.tar.gz

echo "--> Install Istio 1.7"
/usr/local/bin/istioctl-${ISTIOCTL_VERSION} install -f /etc/istio/operator-1-7.yaml -y

echo "Apply custom kyma manifests"
kubectl apply -f /etc/manifests

echo "Apply Kyma related checks and patches"
kubectl patch MutatingWebhookConfiguration istio-sidecar-injector --type 'json' -p '[{"op":"add","path":"/webhooks/0/namespaceSelector/matchExpressions/-","value":{"key":"gardener.cloud/purpose","operator":"NotIn","values":["kube-system"]}}]'
