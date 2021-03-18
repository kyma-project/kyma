set -e
set -o pipefail

backup() {
  namespace="$1"
  secretName="$2"
  backupSecretName="$secretName-backup"

  echo "Removing old backup $backupSecretName"
  kubectl delete secret "$backupSecretName" -n="$namespace" --ignore-not-found

  secret=$(kubectl -n "$namespace" get secret "$secretName" --ignore-not-found)
  if [ -n "$secret" ]
  then
    echo "$secretName exists, backing up to $backupSecretName"

    kubectl get secret "$secretName" -n "$namespace" -ojson | jq "{apiVersion, data, kind, type, metadata: {name: \"$backupSecretName\", namespace: .metadata.namespace}}" | kubectl apply -n "$namespace" -f -
  fi
}

backup kyma-integration connector-service-app-ca

