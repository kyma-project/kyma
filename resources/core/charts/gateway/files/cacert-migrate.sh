set -e
set -o pipefail

function migrateOldCaCertForSecret() {
  TLS_CERT_YAML=$(cat << EOF
---
data:
  cacert: "$1"
EOF
)

  echo "---> Populating cacert key in secret ${NEW_SECRET_NAMESPACE}/${NEW_SECRET_NAME} based on the value from ${OLD_SECRET_NAMESPACE}/${OLD_SECRET_NAME}"
  set +e
  msg=$(kubectl patch secret "${NEW_SECRET_NAME}" --patch "${TLS_CERT_YAML}" -n "${NEW_SECRET_NAMESPACE}" 2>&1)
  status=$?
  set -e

  if [[ $status -ne 0 ]] && [[ ! "$msg" == *"not patched"* ]]; then
    echo "$msg"
    exit $status
fi
}

function makeNewSecretWithCaCert() {
  echo "---> Creating secret ${NEW_SECRET_NAMESPACE}/${NEW_SECRET_NAME} based on cacert value from ${OLD_SECRET_NAMESPACE}/${OLD_SECRET_NAME}"
  set +e
  msg=$(kubectl create secret generic "${NEW_SECRET_NAME}" -n "${NEW_SECRET_NAMESPACE}" --from-literal=.data.cacert="$1" 2>&1)
  status=$?
  set -e

  if [[ $status -ne 0 ]]; then
    echo "$msg"
    exit $status
fi
}

SECRET_OLD_CACERT=$(kubectl -n "${OLD_SECRET_NAMESPACE}" get secret "${OLD_SECRET_NAME}" -o jsonpath='{.data.cacert}' --ignore-not-found)
if [ -n "SECRET_OLD_CACERT" ]; then

  NEW_SECRET=$(kubectl -n "${NEW_SECRET_NAMESPACE}" get secret "${NEW_SECRET_NAME}" --ignore-not-found)
  if [-n "NEW_SECRET" ]; then
    copyOldCaCertToNewSecret "${SECRET_OLD_CACERT}"
  elif
    makeNewSecretWithCaCert "${SECRET_OLD_CACERT}"
  fi
fi
