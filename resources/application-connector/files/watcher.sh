set -e
set -o pipefail

apk add inotify-tools

export SECRETS_DIR=/etc/secrets

if [[ -f "${SECRETS_DIR}/cert" ]]; then
  export SECRET_FILE=${SECRETS_DIR}/cert
else
  export SECRET_FILE=${SECRETS_DIR}/tls.crt
fi

function patchSecret() {
  TLS_CERT_YAML=$(cat << EOF
---
data:
  tls.crt: "$1"
  tls.key: "$2"
EOF
)

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
