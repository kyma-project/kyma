set -e
set -o pipefail

restore() {
  namespace="$1"
  secretName="$2"
  backupSecretName="$secretName-backup"

  secret=$(kubectl -n "$namespace" get secret "$secretName" --ignore-not-found)
  if [ -z "$secret" ]
  then
    echo "No default secret, creating $secretName"
    kubectl create secret generic "$secretName" -n "$namespace"
  fi

  echo "Restoring $backupSecretName"
  secret=$(kubectl -n "$namespace" get secret "$backupSecretName" --ignore-not-found)
  if [ -n "$secret" ]
  then
    echo "$backupSecretName exists, reverting $secretName to $backupSecretName"
    kubectl patch secret -n "$namespace" "$secretName" --type merge -p "$(kubectl get secret "$backupSecretName" -n "$namespace" -ojson | jq "{data}")"

    echo "Removing $backupSecretName"
    kubectl delete secret "$backupSecretName" -n="$namespace" --ignore-not-found
  else
    echo "$backupSecretName does not exist. Aborting..."
  fi
}

restore kyma-integration connector-service-app-ca

