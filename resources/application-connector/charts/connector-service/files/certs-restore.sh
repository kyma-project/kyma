set -e
set -o pipefail

restore() {
  namespace="$1"
  secretName="$2"
  backupSecretName="$secretName-backup"

  echo "Restoring $backupSecretName"
  secret=$(kubectl -n "$namespace" get secret "$backupSecretName" --ignore-not-found)
  if [ -n "$secret" ]
  then
    echo "Removing $secretName"
    kubectl delete secret "$secretName" -n="$namespace" --ignore-not-found
    echo "$backupSecretName exists, reverting $secretName to $backupSecretName"
    kubectl get secret "$backupSecretName" -n "$namespace" -oyaml | sed "s/name:.*$/name: $secretName/" | kubectl apply -n "$namespace" -f -
  else
    echo "$backupSecretName does not exist. Aborting..."
  fi
}

restore {{ .value.namespace }} {{ .value.name }}

