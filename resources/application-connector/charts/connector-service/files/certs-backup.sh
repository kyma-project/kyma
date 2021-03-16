set -e
set -o pipefail

backup() {
  namespace="$1"
  secretName="$2"
  backupSecretName="$secretName-backup"

  echo "Checking existence of $backupSecretName"
  secret=$(kubectl -n "$namespace" get secret "$backupSecretName" --ignore-not-found)
  if [ -n "$secret" ]
  then
    echo "$backupSecretName exists, removing..."
    kubectl delete secret "$backupSecretName" -n="$namespace"
  fi

  echo "Checking existence of $secretName"
  secret=$(kubectl -n "$namespace" get secret "$secretName" --ignore-not-found)
  if [ -n "$secret" ]
  then
    echo "$secretName exists, backing up to $backupSecretName"
    kubectl get secret "$secretName" -n "$namespace" -oyaml | sed "s/name:.*$/name: $backupSecretName/" | kubectl apply -n "$namespace" -f -
  fi
}

backup {{ .value.namespace }} {{ .value.name }}

