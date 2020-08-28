set -e

kubectl -n "${SERVICE_NAMESPACE}" annotate service "${SERVICE_NAME}" \
	cert.gardener.cloud/secretname="${SERVICE_SECRET_NAME}" \
	dns.gardener.cloud/class='garden' \
	dns.gardener.cloud/dnsnames='*.'"${GLOBAL_DOMAIN}"'' \
	--overwrite
