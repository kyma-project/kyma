set -e
set -o pipefail

kubectl get secret connector-service-app-ca --namespace=kyma-integration -oyaml | sed 's/name:.*$/name: connector-service-app-ca-backup/' | kubectl apply --namespace=kyma-integration -f -