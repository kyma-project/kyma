set -e

DOMAIN=""

if [[ $TYPE == "xip" ]]; then
  # When running on xip get the IP address from apiserver-proxy-ssl service
  IP_ADDRESS=$(kubectl get service -n kyma-system apiserver-proxy-ssl -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
  DOMAIN="${IP_ADDRESS}.xip.io"
else
  DOMAIN="{{ .Values.global.domainName }}"
fi

if [[ -z $DOMAIN ]]; then
  exit 1
fi

echo Domain: "$DOMAIN"
echo $DOMAIN > /injected-config/domain

# CA certificate handling

if [[ $TYPE == "legacy" ]]; then
  echo "{{ .Values.global.tlsCrt }}" | base64 --decode > /injected-config/ca-tls-cert.crt
elif [[ $TYPE == "xip" ]]; then
  kubectl get secret -n "{{ .Values.caSecret.namespace }}" "{{ .Values.caSecret.name }}" -o jsonpath='{.data.tls\.crt}' | base64 --decode > /injected-config/ca-tls-cert.crt
fi
