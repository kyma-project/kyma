set +e

export HOME="/tmp"

KYMA_SECRET="$(kubectl get secret -n $KYMA_SECRET_NAMESPACE $KYMA_SECRET_NAME --ignore-not-found -oyaml | grep $KYMA_SECRET_NAME)"

KYMA_CERT="$(kubectl get certificates.cert-manager.io -n $KYMA_SECRET_NAMESPACE $KYMA_SECRET_NAME --ignore-not-found -oyaml | grep $KYMA_SECRET_NAME)"

KYMA_GARDENER_CERT="$(kubectl get certificates.cert.gardener.cloud -n $KYMA_SECRET_NAMESPACE kyma-tls-cert --ignore-not-found -oyaml | grep kyma-tls-cert)"

LEGACY=false
if [ -n "$KYMA_SECRET" ]; then
  LEGACY=true
fi

USER_PROVIDED=false
if [ -n "$KYMA_CERT" ]; then
  USER_PROVIDED=true
fi

GARDENER=false
if [ -n "$KYMA_GARDENER_CERT" ]; then
  GARDENER=true
fi

if [[ $LEGACY == "true" ]] || [[ $USER_PROVIDED == "true" ]] || [[ $GARDENER == "true" ]]; then
  echo "Custom configuration provided. Nothing to do here. Exiting..."
  exit 0
fi

echo "ERROR: No scenario detected. Exiting..."
exit 1