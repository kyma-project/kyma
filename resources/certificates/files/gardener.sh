set -e
set -eo pipefail

export HOME="/tmp"

echo "Checking if running in Gardener mode"

SHOOT_INFO="$(kubectl -n kube-system get configmap shoot-info --ignore-not-found)"
if [ -z "$SHOOT_INFO" ]; then
  echo "Shoot ConfigMap shoot-info/kube-system not present. Nothing to do here. Exiting..."
  exit 0
fi

echo "Gardener mode detected"

# TODO: remove this when Gardener detection is added to the CLI/installation library
echo "Getting shoot domain"
DOMAIN="$(kubectl -n kube-system get configmap shoot-info -o jsonpath='{.data.domain}')"

for var in DOMAIN KYMA_SECRET_NAME KYMA_SECRET_NAMESPACE; do
  if [ -z "${!var}" ]; then
    echo "ERROR: $var is not set"
    discoverUnsetVar=true
  fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

echo "Creating Certificate $KYMA_SECRET_NAME/$KYMA_SECRET_NAMESPACE"

cat <<EOF | kubectl apply -f -
---
apiVersion: cert.gardener.cloud/v1alpha1
kind: Certificate
metadata:
  name: kyma-tls-cert
  namespace: $KYMA_SECRET_NAMESPACE
spec:
  commonName: "*.${DOMAIN}"
  secretName: "$KYMA_SECRET_NAME"
EOF

echo "Annotating istio-ingressgateway/istio-system service"

kubectl -n istio-system annotate service istio-ingressgateway dns.gardener.cloud/class='garden' dns.gardener.cloud/dnsnames='*.'"${DOMAIN}"'' --overwrite

# TODO: remove this when global.ingress.domainName is removed
kubectl create configmap net-global-overrides \
  --from-literal global.domainName="$DOMAIN" \
  --from-literal global.ingress.domainName="$DOMAIN" \
  -n kyma-installer -o yaml --dry-run | kubectl apply -f -

kubectl label configmap net-global-overrides --overwrite installer=overrides -n kyma-installer
kubectl label configmap net-global-overrides --overwrite kyma-project.io/installation="" -n kyma-installer