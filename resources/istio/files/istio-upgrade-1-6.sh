set -ex
set -o pipefail

echo "--> Checking current Istio version"

CURRENT_VERSION=$(kubectl -n istio-system get deployment istiod --ignore-not-found -o jsonpath='{.spec.template.spec.containers[0].image}' | awk -F: '{print $2}')
if [[ "${CURRENT_VERSION}" == "${TARGET_VERSION}" ]]; then
   echo "Istio is already in version ${TARGET_VERSION}. Exiting..."
   exit 0
fi

OPERATOR_FILE="/etc/istio/operator-1-6.yaml"

echo "--> Remove deprecated resources"
if kubectl get customresourcedefinitions.apiextensions.k8s.io | grep -c clusterrbacconfigs.rbac.istio.io ; then
    kubectl delete clusterrbacconfigs.rbac.istio.io default --ignore-not-found=true
fi

if kubectl get customresourcedefinitions.apiextensions.k8s.io | grep -c meshpolicies.authentication.istio.io ; then
    kubectl delete meshpolicies.authentication.istio.io -n istio-system default --ignore-not-found=true
fi

echo "--> Delete deprecated CRDs"
kubectl delete customresourcedefinitions.apiextensions.k8s.io clusterrbacconfigs.rbac.istio.io --ignore-not-found
kubectl delete customresourcedefinitions.apiextensions.k8s.io meshpolicies.authentication.istio.io --ignore-not-found
kubectl delete customresourcedefinitions.apiextensions.k8s.io policies.authentication.istio.io --ignore-not-found
kubectl delete customresourcedefinitions.apiextensions.k8s.io rbacconfigs.rbac.istio.io --ignore-not-found
kubectl delete customresourcedefinitions.apiextensions.k8s.io servicerolebindings.rbac.istio.io --ignore-not-found
kubectl delete customresourcedefinitions.apiextensions.k8s.io serviceroles.rbac.istio.io --ignore-not-found

echo "--> Upgrade to Istio 1.6"
istioctl upgrade -f "${OPERATOR_FILE}" -y

echo "Apply custom kyma manifests"
kubectl apply -f /etc/manifests

echo "Apply Kyma related checks and patches"
kubectl patch MutatingWebhookConfiguration istio-sidecar-injector --type 'json' -p '[{"op":"add","path":"/webhooks/0/namespaceSelector/matchExpressions/-","value":{"key":"gardener.cloud/purpose","operator":"NotIn","values":["kube-system"]}}]'
