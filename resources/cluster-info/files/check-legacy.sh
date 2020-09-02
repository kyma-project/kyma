set -ex

echo "--> Is legacy mode?"
if [[ -s "${CONFIG_MAP_DIR}/global.ingress.tlsCrt" && "${CONFIG_MAP_DIR}/global.ingress.tlsKey" ]]; then
  echo "----> Legacy Cert overrides detected, legacy mode enabled"
else
  echo "----> Not legacy mode, bye"
  exit 0
fi

PATCH_YAML=$(cat << EOF
---
data:
  global.certificates.type: "legacy"
EOF
)

echo "---> Patching cm ${CLUSTER_INFO_NS}/${CLUSTER_INFO_CM}"
set +e
msg=$(kubectl patch cm ${CLUSTER_INFO_CM} --patch "${PATCH_YAML}" -n ${CLUSTER_INFO_NS} 2>&1)
status=$?
set -e

if [[ $status -ne 0 ]] && [[ ! "$msg" == *"not patched"* ]]; then
    echo "$msg"
    exit $status
fi

PATCH_YAML=$(cat << EOF
---
data:
  modules.manager.enabled: "false"
  modules.legacy.enabled: "true"
EOF
)

echo "---> Patching cm ${CLUSTER_INFO_NS}/certificates-overrides"
set +e
msg=$(kubectl patch cm certificates-overrides --patch "${PATCH_YAML}" -n ${CLUSTER_INFO_NS} 2>&1)
status=$?
set -e

if [[ $status -ne 0 ]] && [[ ! "$msg" == *"not patched"* ]]; then
    echo "$msg"
    exit $status
fi
