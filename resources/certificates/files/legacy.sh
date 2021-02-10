set -e
set -eo pipefail

export HOME="/tmp"

echo "Checking if running in legacy mode"

if [ -z "$TLS_KEY" ] || [ -z "$TLS_CRT" ]; then
  echo "TLS key and cert not provided. Nothing to do here. Exiting..."
  exit 0
fi

echo "Legacy mode detected"

for var in DOMAIN TLS_KEY TLS_CRT KYMA_SECRET_NAME KYMA_SECRET_NAMESPACE APISERVER_PROXY_SECRET_NAME APISERVER_PROXY_SECRET_NAMESPACE; do
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

echo "Creating Secret $APISERVER_PROXY_SECRET_NAME/$APISERVER_PROXY_SECRET_NAMESPACE"

kubectl create secret generic -n $APISERVER_PROXY_SECRET_NAMESPACE $APISERVER_PROXY_SECRET_NAME \
    --from-literal tls.crt="$(echo "$TLS_CRT" | base64 --decode)" \
    --from-literal tls.key="$(echo "$TLS_KEY" | base64 --decode)" \
    --dry-run -o yaml | kubectl apply -f -

echo "Creating Secret ingress-tls-cert/kyma-system"

kubectl create secret generic -n kyma-system ingress-tls-cert \
    --from-literal tls.crt="$(echo "$TLS_CRT" | base64 --decode)" \
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

# NOTE: This part can be removed once we get rid of Values.global.ingress.domainName
kubectl create configmap net-global-overrides \
  --from-literal global.domainName="$DOMAIN" \
  --from-literal global.ingress.domainName="$DOMAIN" \
  -n kyma-installer -o yaml --dry-run | kubectl apply -f -

kubectl label configmap net-global-overrides --overwrite installer=overrides -n kyma-installer
kubectl label configmap net-global-overrides --overwrite kyma-project.io/installation="" -n kyma-installer