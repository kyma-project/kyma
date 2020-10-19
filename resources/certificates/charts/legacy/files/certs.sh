# Support for old way of managing certificates for Minikube and Prow only

PATCH_YAML=$(cat << EOF
---
data:
  global.ingress.domainName: "${GLOBAL_DOMAIN}"
EOF
)

echo "---> Patching CM ${CM_NAMESPACE}/${CM_NAME} with ingress domain name"
set +e
msg=$(kubectl patch cm ${CM_NAME} --patch "${PATCH_YAML}" -n ${CM_NAMESPACE} 2>&1)
status=$?
set -e

if [[ $status -ne 0 ]] && [[ ! "$msg" == *"not patched"* ]]; then
    echo "$msg"
    exit $status
fi
