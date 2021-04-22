set -e
set -eo pipefail

export HOME="/tmp"

echo "Checking if running in legacy mode"

if [ -z "$TLS_KEY" ] || [ -z "$TLS_CRT" ]; then
  echo "TLS key and cert not provided. Nothing to do here. Exiting..."
  exit 0
fi

echo "Legacy mode detected"

echo "Checking if running on Gardener"

SHOOT_INFO="$(kubectl -n kube-system get configmap shoot-info --ignore-not-found)"
if [ -n "$SHOOT_INFO" ]; then
  echo "Shoot ConfigMap shoot-info/kube-system present. Ignoring user provided values. Exiting..."
  exit 0
fi

for var in DOMAIN TLS_KEY TLS_CRT KYMA_SECRET_NAME KYMA_SECRET_NAMESPACE; do
  if [ -z "${!var}" ] ; then
    echo "ERROR: $var is not set"
    discoverUnsetVar=true
  fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

echo "Creating Secret $KYMA_SECRET_NAME/$KYMA_SECRET_NAMESPACE"

kubectl create secret generic -n $KYMA_SECRET_NAMESPACE $KYMA_SECRET_NAME \
    --from-literal tls.crt="$(echo "$TLS_CRT" | base64 --decode)" \
    --from-literal tls.key="$(echo "$TLS_KEY" | base64 --decode)" \
    --dry-run -o yaml | kubectl apply -f -

echo "Checking if running on local.kyma.dev"

if [ "${DOMAIN}" == "local.kyma.dev" ]; then
    echo "Configure Kubernetes DNS to support local.kyma.dev"

    COREDNS_PATCH=$(cat << EOF
data:
  Corefile: |
    .:53 {
        errors
        health
        rewrite name regex (.*)\.local\.kyma\.dev istio-ingressgateway.istio-system.svc.cluster.local
        ready
        kubernetes cluster.local in-addr.arpa ip6.arpa {
          pods insecure
          fallthrough in-addr.arpa ip6.arpa
        }
        hosts /etc/coredns/NodeHosts {
          reload 1s
          fallthrough
        }
        prometheus :9153
        forward . /etc/resolv.conf
        cache 30
        loop
        reload
        loadbalance
    }
EOF
    )
    kubectl patch configmap coredns --patch "${COREDNS_PATCH}" -n kube-system
fi
