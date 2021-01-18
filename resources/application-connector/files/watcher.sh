set -e
set -o pipefail

apk add inotify-tools

export SECRETS_DIR=/etc/secrets

if [[ -f "${SECRETS_DIR}/cert" ]]; then
  export SECRET_FILE=${SECRETS_DIR}/cert
else
  export SECRET_FILE=${SECRETS_DIR}/tls.crt
fi

function createSecret() {
  cat <<EOF | kubectl -n kyma-integration apply -f -
apiVersion: v1
kind: Secret
data:
  "cacert": ""
metadata:
  name: "$1"
  namespace: "$2"
type: Opaque
EOF
}

function patchSecret() {
  TLS_CERT_YAML=$(cat << EOF
---
data:
  tls.crt: "$1"
  tls.key: "$2"
EOF
)
  echo "---> Checking if secret ${MTLS_GATEWAY_NAMESPACE}/${MTLS_GATEWAY_NAME} exists"
  set +e

  msg=$(kubectl get secret "${MTLS_GATEWAY_NAME}" -n "${MTLS_GATEWAY_NAMESPACE}" 2>&1)
  status=$?
  set -e

  if [[ $status -ne 0 ]] && [[ "$msg" == *"not found"* ]]; then
    set +e
    echo "---> Creating secret ${MTLS_GATEWAY_NAMESPACE}/${MTLS_GATEWAY_NAME}"
    msg=$(createSecret "${MTLS_GATEWAY_NAME}" "${MTLS_GATEWAY_NAMESPACE}" 2>&1)
    echo $msg
    set -e
  fi

  echo "---> Patching secret ${MTLS_GATEWAY_NAMESPACE}/${MTLS_GATEWAY_NAME}"
  set +e
  msg=$(kubectl patch secret "${MTLS_GATEWAY_NAME}" --patch "${TLS_CERT_YAML}" -n "${MTLS_GATEWAY_NAMESPACE}" 2>&1)
  status=$?
  set -e

  if [[ $status -ne 0 ]] && [[ ! "$msg" == *"not patched"* ]]; then
    echo "$msg"
    exit $status
fi
}

function syncSecret() {
  echo "---> Get Gateway cert and key"
  if [[ -f "${SECRETS_DIR}/cert" ]]; then
    TLS_CRT=$(cat "${SECRETS_DIR}/cert" | base64 -w 0)
    TLS_KEY=$(cat "${SECRETS_DIR}/key" | base64 -w 0)
  else
    TLS_CRT=$(cat "${SECRETS_DIR}/tls.crt" | base64 -w 0)
    TLS_KEY=$(cat "${SECRETS_DIR}/tls.key" | base64 -w 0)
  fi

  patchSecret "${TLS_CRT}" "${TLS_KEY}"
}

echo "---> Initial sync"
syncSecret

while true; do
  echo "---> Listen for cert changes"
  inotifywait -e DELETE_SELF $SECRET_FILE |
    while read path _ file; do
      echo "---> $path$file modified"
      syncSecret
    done
done
