set -e
set -o pipefail

echo "Checking if running in user-provided mode"

ISSUER_CM="$(kubectl get configmap -n $ISSUER_CM_NAMESPACE $ISSUER_CM_NAME --ignore-not-found -o jsonpath='{.data.issuer}')"
if [ -z "$ISSUER_CM" ]; then
  echo "Issuer ConfigMap $ISSUER_CM_NAME/$ISSUER_CM_NAMESPACE not present. Nothing to do here. Exiting..."
  exit 0
fi

echo "Issuer ConfigMap $ISSUER_CM_NAME/$ISSUER_CM_NAMESPACE found."

if [ -z "$DOMAIN" ]; then
  echo "User-provided mode requested, but domain not provided. Exiting..."
  exit 1
fi

echo "Validating Issuer"

IS_CLUSTER_ISSUER=$(echo $ISSUER_CM | grep "ClusterIssuer")
if [ -z "$IS_CLUSTER_ISSUER" ]; then
  echo "Provided Issuer is not a ClusterIssuer. Exiting..."
  exit 1
fi

IS_NAME_PROPER=$(echo $ISSUER_CM | grep "name: $ISSUER_NAME")
if [ -z "$IS_NAME_PROPER" ]; then
  echo "Issuer name should be $ISSUER_NAME. Exiting..."
  exit 1
fi

echo "Checking if Issuer type is self-signed"

IS_SELF_SIGNED="" # $(echo $ISSUER_CM | grep "selfSigned: {}")

if [ -n "$IS_SELF_SIGNED" ]; then
  echo "Self-signed certificate requested. Generating required overrides"
  kubectl create configmap net-global-overrides \
        --from-literal global.certificates.selfSigned="true" \
        -n kyma-installer -o yaml --dry-run | kubectl apply -f -
fi

echo "Creating issuer"

echo "$ISSUER_CM"
echo "$ISSUER_CM" | kubectl apply -f -

echo "Generating certificates"

cat <<EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: $KYMA_SECRET_NAME
  namespace: istio-system
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
  namespace: kyma-system
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

# TODO: if self signed wait for the certificate