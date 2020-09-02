set -ex

CERT_TYPE=$(cat ${CONFIG_MAP_DIR}/global.certificates.type)

if [[ "$CERT_TYPE" != "detect" ]]; then
  echo "--> $CERT_TYPE requested, nothing to do here"
  exit 0
fi

echo "----> If you can see this, nothing was detected, defaulting to user-provided"

PATCH_YAML=$(cat << EOF
---
data:
  global.certificates.type: "user-provided"
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
  modules.manager.enabled: "true"
  modules.user-provided.enabled: "true"
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
