set -e
set -o pipefail

backup() {
  namespace="$1"
  secretName="$2"
  backupSecretName="$secretName-backup"

  kubectl delete secret "$backupSecretName" -n="$namespace" --ignore-not-found

  echo "Checking existence of $secretName"
  secret=$(kubectl -n "$namespace" get secret "$secretName" --ignore-not-found)
  if [ -n "$secret" ]
  then
    echo "$secretName exists, backing up to $backupSecretName"
    kubectl get secret "$secretName" -n "$namespace" -oyaml | sed "s/name:.*$/name: $backupSecretName/" | kubectl apply -n "$namespace" -f -
    kubectl delete secret "$backupSecretName" -n="$namespace" --ignore-not-found
  fi
}

backup {{ .value.namespace }} {{ .value.name }}

