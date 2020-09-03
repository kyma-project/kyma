set -e

DOMAIN=""
# When running on xip get the IP address from apiserver-proxy-ssl service
if [[ $TYPE == "xip" ]]; then
  IP_ADDRESS=$(kubectl get service -n kyma-system apiserver-proxy-ssl -o jsonpath='{.spec.clusterIP}')
  DOMAIN="${IP_ADDRESS}.xip.io"
elif [[ $TYPE == "gardener" ]]; then
  URL=$(kubectl get service -n kyma-system apiserver-proxy-ssl -o jsonpath='{.metadata.annotations.dns\.gardener\.cloud/dnsnames}')
  # need to remove the apiserver. prefix from the annotation to get a domain
  DOMAIN=${URL#"apiserver."}
elif [[ $TYPE == "legacy" || $TYPE == "user-provided" ]]; then
  DOMAIN="{{ .Values.global.domainName }}"
fi

if [[ -z $DOMAIN ]]; then
  exit 1
fi

echo Domain: "$DOMAIN"
echo $DOMAIN > /injected-config/domain