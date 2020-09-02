# Support for old way of managing certificates for Minikube and Prow only

PATCH_YAML=$(cat << EOF
---
data:
  global.ingress.domainName: "${GLOBAL_DOMAIN}"
EOF
)

echo $PATCH_YAML
kubectl patch configmap net-global-overrides --patch "${PATCH_YAML}" -n kyma-installer
