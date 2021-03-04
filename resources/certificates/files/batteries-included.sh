#!/bin/bash
set +e

export HOME="/tmp"

DOMAIN=kyma.example.dev #TODO: remove
ISSUER_NAME=kyma-ca-issuer #TODO: remove
KYMA_SECRET_NAME=kyma-gateway-certs #TODO: remove
KYMA_SECRET_NAMESPACE=istio-system #TODO: remove

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

echo "No scenario detected. Defaulting to batteries-included mode..."

echo "Creating issuer"

cat <<EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: $ISSUER_NAME
spec:
  selfSigned: {}
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
  isCA: true
  issuerRef:
    name: $ISSUER_NAME
    kind: ClusterIssuer
  dnsNames:
  - "*.$DOMAIN"
  commonName: "*.$DOMAIN"
EOF

echo "Checking if running on kyma.example.dev"

if [ "${DOMAIN}" == "kyma.example.dev" ]; then
    echo "Configuring Kubernetes DNS to support kyma.example.dev"

    COREDNS_PATCH=$(cat << EOF
data:
  Corefile: |
    .:53 {
        errors
        health
        rewrite name regex (.*)\.kyma\.example\.dev istio-ingressgateway.istio-system.svc.cluster.local
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
    kubectl patch configmap coredns-coredns --patch "${COREDNS_PATCH}" -n kube-system
fi

# TODO: add patching KubeDNS configmap

# TODO: remove this when global.ingress.domainName is removed
kubectl create configmap net-global-overrides \
  --from-literal global.domainName="$DOMAIN" \
  --from-literal global.ingress.domainName="$DOMAIN" \
  -n kyma-installer -o yaml --dry-run | kubectl apply -f -

kubectl label configmap net-global-overrides --overwrite installer=overrides -n kyma-installer
kubectl label configmap net-global-overrides --overwrite kyma-project.io/installation="" -n kyma-installer

echo "Exiting..."
exit 0
