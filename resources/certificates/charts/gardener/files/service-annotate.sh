set -e

echo "--> Get Domain from ${SHOOT_CM_NAMESPACE}/${SHOOT_CM_NAME}"
GLOBAL_DOMAIN="$(kubectl -n ${SHOOT_CM_NAMESPACE} get cm ${SHOOT_CM_NAME} -o jsonpath='{.data.domain}')"

echo "--> Annotate ${SERVICE_NAMESPACE}/${SERVICE_NAME}"
kubectl -n "${SERVICE_NAMESPACE}" annotate service "${SERVICE_NAME}" \
	cert.gardener.cloud/secretname="${SERVICE_SECRET_NAME}" \
	dns.gardener.cloud/class='garden' \
	dns.gardener.cloud/dnsnames='*.'"${GLOBAL_DOMAIN}"'' \
	--overwrite

echo "--> Update ${CLUSTER_INFO_CM_NAMESPACE}/${CLUSTER_INFO_CM_NAME}"

PATCH_YAML=$(cat << EOF
---
data:
  global.ingress.domainName: ${GLOBAL_DOMAIN}
  global.domainName: ${GLOBAL_DOMAIN}
EOF
)

echo "---> Patching cm ${CLUSTER_INFO_CM_NAMESPACE}/${CLUSTER_INFO_CM_NAME}"
set +e
msg=$(kubectl patch cm ${CLUSTER_INFO_CM_NAME} --patch "${PATCH_YAML}" -n ${CLUSTER_INFO_CM_NAMESPACE} 2>&1)
status=$?
set -e

if [[ $status -ne 0 ]] && [[ ! "$msg" == *"not patched"* ]]; then
    echo "$msg"
    exit $status
fi
