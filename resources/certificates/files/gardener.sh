set -e
set -eo pipefail

export HOME="/tmp"

CERT_CHECK_TIMEOUT=${CERT_CHECK_TIMEOUT:=5}

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

get_cert_status () {
  local cert=$1
  local nspace=$2
  kubectl -n $nspace get certificates.cert.gardener.cloud $cert -o jsonpath='{.status.state}'
}

check_cert_status() {
  local cert=$1
  local nspace=$2
  local cert_check_timeout=$3

  local next_wait_time=0
  local cert_status=$(get_cert_status $cert $nspace)

  while [[ $next_wait_time -lt $cert_check_timeout && $cert_status != "Ready" ]]; do
    sleep $(( next_wait_time++ ))
    cert_status=$(get_cert_status $cert $nspace)
  done
  if [[ $cert_status != "Ready" ]]; then
      echo "certificate $cert in ns $nspace is not Ready"
      exit 1
  fi
}

check_cert_status "kyma-tls-cert" $KYMA_SECRET_NAMESPACE $CERT_CHECK_TIMEOUT

echo "Annotating istio-ingressgateway/istio-system service"

kubectl -n istio-system annotate service istio-ingressgateway dns.gardener.cloud/class='garden' dns.gardener.cloud/dnsnames='*.'"${DOMAIN}"'' --overwrite
