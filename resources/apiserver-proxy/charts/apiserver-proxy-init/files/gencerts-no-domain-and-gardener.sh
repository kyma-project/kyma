#!/usr/bin/env bash
#
# This script runs when Helm override global.domainName is not set and we're in Gardener environment
#
set -e
if [ "$DOMAIN" = "" ]; then
  source /app/utils.sh
  INGRESS_IP=$(getLoadBalancerIP {{ template "name" . }}-ssl {{ .Release.Namespace }})
  DOMAIN="$INGRESS_IP.xip.io"
  kubectl create configmap {{ template "name" . }} --from-literal DOMAIN="$DOMAIN"
fi
if [ "$(cat /etc/apiserver-proxy-tls-cert/tls.key)" = "" ]; then
  # when running on Gardener && there is key mouted we skip, do nothing
  echo "Skipping processing existing cert because environment is Gardener which rotates certs by itself"
fi
