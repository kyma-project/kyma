#!/usr/bin/env bash

TLS_CRT_VAL=$(kubectl get secret -n istio-system kyma-gateway-certs -o jsonpath='{.data.cert}')
echo "TLS_CRT_VAL: ${TLS_CRT_VAL}"

TLS_CRT_PATCH=$(cat << EOF
---
  data:
    tls.crt: "${TLS_CRT_VAL}"
EOF
)
echo "TLS_CRT_PATCH: ${TLS_CRT_PATCH}"

kubectl patch secret -n kyma-system ingress-tls-cert --patch "${TLS_CRT_PATCH}"
