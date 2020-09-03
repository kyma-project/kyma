set -e

GLOBAL_DOMAIN="$(kubectl -n ${SHOOT_CM_NAMESPACE} get cm ${SHOOT_CM_NAME} -o jsonpath='{.data.domain}')"

kubectl -n "${SERVICE_NAMESPACE}" annotate service "${SERVICE_NAME}" \
	cert.gardener.cloud/secretname="${SERVICE_SECRET_NAME}" \
	dns.gardener.cloud/class='garden' \
	dns.gardener.cloud/dnsnames='*.'"${GLOBAL_DOMAIN}"'' \
	--overwrite
