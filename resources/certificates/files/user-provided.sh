set -e
set -eo pipefail

export HOME="/tmp"

echo "Checking if running in user-provided mode"

if [ -z "$ISSUER" ]; then
  echo "Issuer not provided. Nothing to do here. Exiting..."
  exit 0
fi

echo "Issuer provided. User-provided mode detected"

echo "Checking if running on Gardener"

SHOOT_INFO="$(kubectl -n kube-system get configmap shoot-info --ignore-not-found)"
if [ -n "$SHOOT_INFO" ]; then
  echo "Shoot ConfigMap shoot-info/kube-system present. Ignoring user provided values. Exiting..."
  exit 0
fi

for var in DOMAIN ISSUER_NAME KYMA_SECRET_NAME KYMA_SECRET_NAMESPACE; do
  if [ -z "${!var}" ] ; then
    echo "ERROR: $var is not set"
    discoverUnsetVar=true
  fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

echo "Validating Issuer"

if echo "$ISSUER" | grep "ClusterIssuer"; then
  echo "Type is proper"
else
  echo "Provided Issuer is not a ClusterIssuer. Exiting..."
  exit 1
fi

if echo "$ISSUER" | grep "name: $ISSUER_NAME"; then
  echo "Name is proper"
else
  echo "Issuer name should be $ISSUER_NAME. Exiting..."
  exit 1
fi

echo "Creating issuer"

cat <<EOF | kubectl apply -f -
{{ .Values.global.issuer | printf "%s" | nindent 16 }}
EOF

echo "Generating certificate"

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
