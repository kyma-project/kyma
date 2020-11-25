set -e
set -o pipefail

function copyOldCaCertToNewSecret() {
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
  msg=$(kubectl create secret generic "${NEW_SECRET_NAME}" -n "${NEW_SECRET_NAMESPACE}" --from-literal=cacert="$1" 2>&1)
  status=$?
  set -e

  if [[ $status -ne 0 ]]; then
    echo "$msg"
    exit $status
fi
}

echo "---> Starting script"
echo "The old secret is ${OLD_SECRET_NAMESPACE}/${OLD_SECRET_NAME}"
echo "The new secret is ${NEW_SECRET_NAMESPACE}/${NEW_SECRET_NAME}"

echo "---> Checking old secret"
SECRET_OLD_CACERT_BASE64=$(kubectl -n "${OLD_SECRET_NAMESPACE}" get secret "${OLD_SECRET_NAME}" -o jsonpath='{.data.cacert}' --ignore-not-found)

if [ -n "$SECRET_OLD_CACERT_BASE64" ]
then
  echo "---> Checking new secret"
  NEW_SECRET=$(kubectl -n "${NEW_SECRET_NAMESPACE}" get secret "${NEW_SECRET_NAME}" --ignore-not-found)

  if [ -n "$NEW_SECRET" ]
  then
    echo "---> Destination secret already exists"
    copyOldCaCertToNewSecret "${SECRET_OLD_CACERT_BASE64}"
  else
    echo "---> Destination secret does not exist"
    SECRET_OLD_CACERT_DECODED=$(echo "${SECRET_OLD_CACERT_BASE64}" | base64 --decode)
    makeNewSecretWithCaCert "${SECRET_OLD_CACERT_DECODED}"
  fi
else
  echo "---> Could not read cacert from old secret ${OLD_SECRET_NAMESPACE}/${OLD_SECRET_NAME}"
fi

if [ -n "$DELETE_OLD_SECRET_FLAG" ]
then
  echo "---> Delete backup secret ${OLD_SECRET_NAMESPACE}/${OLD_SECRET_NAME}"
  set +e
  msg=$(kubectl delete secret "${OLD_SECRET_NAME}" -n "${OLD_SECRET_NAMESPACE}" 2>&1)
  status=$?
  set -e

  if [[ $status -ne 0 ]]
  then
    echo "$msg"
    exit $status
  fi
fi
