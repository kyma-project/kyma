set -ex

function legacyMode() {
  cat << EOF > overrides-global.yaml
---
data:
  global.certificates.type: "legacy"
EOF

  cat << EOF > overrides-modules.yaml
---
data:
  modules.legacy.enabled: "true"
EOF

  patchCM "${OVERRIDES_NAME}" "$PWD/overrides-global.yaml"
  patchCM "certificates-overrides" "$PWD/overrides-modules.yaml"
}

function gardenerMode() {

  cat << EOF > overrides-global.yaml
---
data:
  global.certificates.type: "gardener"
EOF

  cat << EOF > overrides-modules.yaml
---
data:
  modules.gardener.enabled: "true"
EOF

  patchCM "${OVERRIDES_NAME}" "$PWD/overrides-global.yaml"
  patchCM "certificates-overrides" "$PWD/overrides-modules.yaml"
}

function userProvidedMode() {
  cat << EOF > overrides-global.yaml
---
data:
  global.certificates.type: "user-provided"
EOF

  cat << EOF > overrides-modules.yaml
---
data:
  modules.user-provided.enabled: "true"
EOF

  cat << EOF > certmanager-modules.yaml
---
data:
  modules.manager.enabled: "true"
EOF

  patchCM "${OVERRIDES_NAME}" "$PWD/overrides-global.yaml"
  patchCM "certificates-overrides" "$PWD/overrides-modules.yaml"
  patchCM "cert-manager-overrides" "$PWD/certmanager-modules.yaml"
}

function patchCM() {
  CM_NAME="$1"
  PATCH_YAML=$(cat "$2")

  echo "---> Patching cm ${OVERRIDES_NS}/${CM_NAME}"
  set +e
  msg=$(kubectl patch cm ${CM_NAME} --patch "${PATCH_YAML}" -n ${OVERRIDES_NS} 2>&1)
  status=$?
  set -e

  if [[ $status -ne 0 ]] && [[ ! "$msg" == *"not patched"* ]]; then
      echo "$msg"
      exit $status
  fi
}

if [[ "$CERT_TYPE" != "detect" ]]; then
  echo "--> $CERT_TYPE requested. No need to detect"
  exit 0
fi

echo "--> Is gardener mode?"
API_VERSIONS=$(kubectl api-versions)
if echo $API_VERSIONS | grep -c cert.gardener.cloud ; then
  echo "--> Gardener Certificate CR present, gardener mode enabled"
  gardenerMode
  exit 0
fi

echo "--> Is legacy mode?"
TLS_CRT_EXISTS=$(kubectl get cm -n "${OVERRIDES_NS}" "${OVERRIDES_NAME}" -o jsonpath='{.data.global\.tlsCrt}' --ignore-not-found)
if [[ -n "$TLS_CRT_EXISTS" ]]; then
  echo "----> Legacy Cert overrides detected, legacy mode enabled"
  legacyMode
  exit 0
fi

echo "--> Defaulting to user-provided mode"
userProvidedMode
