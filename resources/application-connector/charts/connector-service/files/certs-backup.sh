set -e
set -o pipefail

secretName="connector-service-app-ca"
backupSecretName="$secretName-backup"

echo "Checking existence of $backupSecretName"
secret=$(kubectl -n kyma-integration get secret $backupSecretName --ignore-not-found)
if [ -n "$secret" ]
then
  echo "$backupSecretName exists, removing..."
  kubectl delete secret $backupSecretName --namespace=kyma-integration
fi

echo "Checking existence of $secretName"
secret=$(kubectl -n kyma-integration get secret $secretName --ignore-not-found)
if [ -n "$secret" ]
then
  echo "$secretName exists, backing up to $backupSecretName"
  kubectl get secret $secretName --namespace=kyma-integration -oyaml | sed "s/name:.*$/name: $backupSecretName/" | kubectl apply --namespace=kyma-integration -f -
fi