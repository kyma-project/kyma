set -e

kubectl -n "${ISTIO_GATEWAY_NAMESPACE}" annotate service "${ISTIO_GATEWAY_NAME}" \
	dns.gardener.cloud/class='garden' \
	dns.gardener.cloud/dnsnames='*.'"${GLOBAL_DOMAIN}"'' \
	--overwrite
