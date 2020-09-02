# Support for old way of managing certificates for Minikube and Prow only
#echo "${GLOBAL_TLS_KEY}" | base64 -d > ${HOME}/key.pem
#echo "${GLOBAL_TLS_CERT}" | base64 -d > ${HOME}/cert.pem

PATCH_YAML=$(cat << EOF
---
data:
  global.ingress.domainName: "${GLOBAL_DOMAIN}"
EOF
)

echo $PATCH_YAML
kubectl patch configmap net-global-overrides --patch "${PATCH_YAML}" -n kyma-installer

#kubectl create secret tls kyma-gateway-certs -n istio-system --key ${HOME}/key.pem --cert ${HOME}/cert.pem

#rm ${HOME}/key.pem
#rm ${HOME}/cert.pem
