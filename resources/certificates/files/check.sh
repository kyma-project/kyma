set +e

KYMA_SECRET="$(kubectl get secret -n $KYMA_SECRET_NAMESPACE $KYMA_SECRET_NAME --ignore-not-found)"
APISERVER_PROXY_SECRET="$(kubectl get secret -n $APISERVER_PROXY_SECRET_NAMESPACE $APISERVER_PROXY_SECRET_NAME --ignore-not-found)"

KYMA_CERT="$(kubectl get certificates.cert-manager.io -n $KYMA_SECRET_NAMESPACE $KYMA_SECRET_NAME --ignore-not-found)"
APISERVER_PROXY_CERT="$(kubectl get certificates.cert-manager.io -n $APISERVER_PROXY_SECRET_NAMESPACE $APISERVER_PROXY_SECRET_NAME --ignore-not-found)"

KYMA_GARDENER_CERT="$(kubectl get certificates.cert.gardener.cloud -n $KYMA_SECRET_NAMESPACE kyma-tls-cert --ignore-not-found)"
APISERVER_PROXY_GARDENER_CERT="$(kubectl get certificates.cert.gardener.cloud -n $APISERVER_PROXY_SECRET_NAMESPACE apiserver-proxy-tls-cert --ignore-not-found)"

LEGACY=false
if [ -z "$KYMA_SECRET" ] && [ -z "$APISERVER_PROXY_SECRET" ]; then
  LEGACY=true
fi

echo $LEGACY

USER_PROVIDED=false
if [ -z "$KYMA_CERT" ] && [ -z "$APISERVER_PROXY_CERT" ]; then
  USER_PROVIDED=true
fi

echo $USER_PROVIDED

GARDENER=false
if [ -z "$KYMA_GARDENER_CERT" ] && [ -z "$APISERVER_PROXY_GARDENER_CERT" ]; then
  GARDENER=true
fi

echo $GARDENER

if [ $LEGACY ] || [ $USER_PROVIDED ] || [ $GARDENER ]; then
  exit 0
else
  echo "None of the scenarios were launched"
  exit 1
fi