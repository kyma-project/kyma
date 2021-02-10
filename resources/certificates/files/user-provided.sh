set -e
set -eo pipefail

export HOME="/tmp"

for var in ISSUER_CM_NAME ISSUER_CM_NAMESPACE; do
  if [ -z "${!var}" ] ; then
    echo "ERROR: $var is not set"
    discoverUnsetVar=true
  fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

echo "Checking if running in user-provided mode"

ISSUER_CM="$(kubectl get configmap -n $ISSUER_CM_NAMESPACE $ISSUER_CM_NAME --ignore-not-found -o jsonpath='{.data.issuer}')"
if [ -z "$ISSUER_CM" ]; then
  echo "Issuer ConfigMap $ISSUER_CM_NAME/$ISSUER_CM_NAMESPACE not present. Nothing to do here. Exiting..."
  exit 0
fi

echo "Issuer ConfigMap $ISSUER_CM_NAME/$ISSUER_CM_NAMESPACE found. User-provided mode detected"

for var in DOMAIN ISSUER_NAME KYMA_SECRET_NAME KYMA_SECRET_NAMESPACE APISERVER_PROXY_SECRET_NAME APISERVER_PROXY_SECRET_NAMESPACE IS_SELF_SIGNED; do
  if [ -z "${!var}" ] ; then
    echo "ERROR: $var is not set"
    discoverUnsetVar=true
  fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

echo "Validating Issuer"

if echo "$ISSUER_CM" | grep "ClusterIssuer"; then
  echo "Type is proper"
else
  echo "Provided Issuer is not a ClusterIssuer. Exiting..."
  exit 1
fi

if echo "$ISSUER_CM" | grep "name: $ISSUER_NAME"; then
  echo "Name is proper"
else
  echo "Issuer name should be $ISSUER_NAME. Exiting..."
  exit 1
fi

echo "Creating issuer"

echo "$ISSUER_CM" | kubectl apply -f -

echo "Generating certificates"

cat <<EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: $KYMA_SECRET_NAME
  namespace: $KYMA_SECRET_NAMESPACE
spec:
  duration: 720h
  renewBefore: 10m
  secretName: $KYMA_SECRET_NAME
  issuerRef:
    name: $ISSUER_NAME
    kind: ClusterIssuer
  dnsNames:
  - "*.$DOMAIN"
  commonName: "*.$DOMAIN"
EOF

cat <<EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: $APISERVER_PROXY_SECRET_NAME
  namespace: $APISERVER_PROXY_SECRET_NAMESPACE
spec:
  duration: 720h
  renewBefore: 10m
  secretName: $APISERVER_PROXY_SECRET_NAME
  issuerRef:
    # The issuer created previously
    name: $ISSUER_NAME
    kind: ClusterIssuer
  dnsNames:
  - "apiserver.$DOMAIN"
  commonName: "apiserver.$DOMAIN"
EOF

# Note: The following part can be removed when we get rid of the components that need
# a mounted secret with a certificate to trust it.
if [ "$IS_SELF_SIGNED" == "true" ]; then

  SECONDS=0
  END_TIME=$((SECONDS+180))

  echo "Self-signed certificate requested. Waiting $END_TIME for kyma-gateway-certs secret..."

  while [ ${SECONDS} -lt ${END_TIME} ]; do

    KYMA_CA_CERT=$(kubectl get secret ${KYMA_SECRET_NAME} -n istio-system -o jsonpath="{.data['ca\.crt']}" --ignore-not-found)

    if [ -n "$KYMA_CA_CERT"  ]; then
      echo "Secret is ready. Copying the CA certificate to Secret ingress-tls-cert/kyma-system"

      kubectl create secret generic -n kyma-system ingress-tls-cert \
                --from-literal tls.crt="$(echo "$KYMA_CA_CERT" | base64 --decode)" --dry-run -o yaml \
                | kubectl apply -f -
      exit 0
    fi

    echo "Secret is not ready. Sleeping 10 seconds..."
    sleep 10

  done

  echo "Secret not found! Exiting..."
  exit 1

fi

# NOTE: This part can be removed once we get rid of Values.global.ingress.domainName
kubectl create configmap net-global-overrides \
  --from-literal global.domainName="$DOMAIN" \
  --from-literal global.ingress.domainName="$DOMAIN" \
  -n kyma-installer -o yaml --dry-run | kubectl apply -f -

kubectl label configmap net-global-overrides --overwrite installer=overrides -n kyma-installer
kubectl label configmap net-global-overrides --overwrite kyma-project.io/installation="" -n kyma-installer