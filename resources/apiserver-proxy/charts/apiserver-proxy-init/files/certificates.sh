set -e
set -eo pipefail

export HOME="/tmp"

echo "Checking if running in Gardener mode"

SHOOT_INFO="$(kubectl -n kube-system get configmap shoot-info --ignore-not-found)"
if [ -z "$SHOOT_INFO" ]; then
  echo "Shoot ConfigMap shoot-info/kube-system not present. Nothing to do here. Exiting..."
  exit 0
fi

echo "Annotating apiserver-proxy-ssl/kyma-system service"

kubectl -n kyma-system annotate service apiserver-proxy-ssl dns.gardener.cloud/class='garden' dns.gardener.cloud/dnsnames='apiserver.'"${DOMAIN}"'' --overwrite
